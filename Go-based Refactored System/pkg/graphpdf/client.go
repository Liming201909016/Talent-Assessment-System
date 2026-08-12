package graphpdf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	defaultTokenBaseURL = "https://login.microsoftonline.com"
	defaultGraphBaseURL = "https://graph.microsoft.com/v1.0"
	maxDOCXBytes        = 20 << 20
	maxPDFBytes         = 50 << 20
)

type Config struct {
	TenantID       string
	ClientID       string
	ClientSecret   string
	DriveID        string
	Folder         string
	TimeoutSeconds int
}

type Client struct {
	cfg          Config
	httpClient   *http.Client
	tokenBaseURL string
	graphBaseURL string
	tokenMu      sync.Mutex
	accessToken  string
	tokenExpires time.Time
	retryDelay   time.Duration
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type driveItem struct {
	ID string `json:"id"`
}

func NewClient(cfg Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 90 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{cfg: cfg, httpClient: httpClient, tokenBaseURL: defaultTokenBaseURL, graphBaseURL: defaultGraphBaseURL, retryDelay: 500 * time.Millisecond}
}

func (c *Client) Convert(ctx context.Context, fileName string, docx []byte) ([]byte, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	if len(docx) == 0 || len(docx) > maxDOCXBytes {
		return nil, errors.New("Word报告文件大小无效")
	}
	fileName = path.Base(strings.TrimSpace(fileName))
	if fileName == "." || fileName == "" || !strings.EqualFold(path.Ext(fileName), ".docx") {
		return nil, errors.New("Word报告文件名无效")
	}
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	itemID, err := c.upload(ctx, token, fileName, docx)
	if err != nil {
		return nil, err
	}
	defer c.delete(context.WithoutCancel(ctx), token, itemID)
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		pdf, retry, err := c.downloadPDF(ctx, token, itemID)
		if err == nil {
			return pdf, nil
		}
		lastErr = err
		if !retry || attempt == 3 {
			break
		}
		timer := time.NewTimer(c.retryDelay * time.Duration(1<<attempt))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.New("Microsoft Graph转换PDF超时")
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func (c *Client) validate() error {
	if strings.TrimSpace(c.cfg.TenantID) == "" || strings.TrimSpace(c.cfg.ClientID) == "" || strings.TrimSpace(c.cfg.ClientSecret) == "" || strings.TrimSpace(c.cfg.DriveID) == "" {
		return errors.New("Microsoft Graph报告转换配置不完整")
	}
	return nil
}

func (c *Client) token(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.accessToken != "" && time.Now().Add(time.Minute).Before(c.tokenExpires) {
		return c.accessToken, nil
	}
	form := url.Values{
		"client_id":     {c.cfg.ClientID},
		"client_secret": {c.cfg.ClientSecret},
		"scope":         {"https://graph.microsoft.com/.default"},
		"grant_type":    {"client_credentials"},
	}
	endpoint := strings.TrimRight(c.tokenBaseURL, "/") + "/" + url.PathEscape(c.cfg.TenantID) + "/oauth2/v2.0/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errors.New("创建Microsoft Graph认证请求失败")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.New("连接Microsoft Graph认证服务失败")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Microsoft Graph认证失败（HTTP %d）", resp.StatusCode)
	}
	var payload tokenResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil || strings.TrimSpace(payload.AccessToken) == "" {
		return "", errors.New("Microsoft Graph认证响应无效")
	}
	if payload.ExpiresIn <= 0 {
		payload.ExpiresIn = 300
	}
	c.accessToken = payload.AccessToken
	c.tokenExpires = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	return c.accessToken, nil
}

func (c *Client) upload(ctx context.Context, token, fileName string, docx []byte) (string, error) {
	folder := strings.Trim(strings.TrimSpace(c.cfg.Folder), "/")
	segments := make([]string, 0, 2)
	if folder != "" {
		for _, segment := range strings.Split(folder, "/") {
			if segment = strings.TrimSpace(segment); segment != "" && segment != "." && segment != ".." {
				segments = append(segments, url.PathEscape(segment))
			}
		}
	}
	segments = append(segments, url.PathEscape(fileName))
	remotePath := strings.Join(segments, "/")
	endpoint := fmt.Sprintf("%s/drives/%s/root:/%s:/content", strings.TrimRight(c.graphBaseURL, "/"), url.PathEscape(c.cfg.DriveID), remotePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(docx))
	if err != nil {
		return "", errors.New("创建Word报告上传请求失败")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.New("上传Word报告到Microsoft Graph失败")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("上传Word报告失败（HTTP %d）", resp.StatusCode)
	}
	var item driveItem
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&item); err != nil || strings.TrimSpace(item.ID) == "" {
		return "", errors.New("Word报告上传响应无效")
	}
	return item.ID, nil
}

func (c *Client) downloadPDF(ctx context.Context, token, itemID string) ([]byte, bool, error) {
	endpoint := fmt.Sprintf("%s/drives/%s/items/%s/content?format=pdf", strings.TrimRight(c.graphBaseURL, "/"), url.PathEscape(c.cfg.DriveID), url.PathEscape(itemID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, false, errors.New("创建Microsoft Graph PDF请求失败")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, errors.New("Microsoft Graph转换PDF失败")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		retry := resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusLocked || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		return nil, retry, fmt.Errorf("Microsoft Graph转换PDF失败（HTTP %d）", resp.StatusCode)
	}
	pdf, err := io.ReadAll(io.LimitReader(resp.Body, maxPDFBytes+1))
	if err != nil || len(pdf) > maxPDFBytes {
		return nil, false, errors.New("读取Microsoft Graph PDF失败")
	}
	if len(pdf) < 1024 || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		return nil, false, errors.New("Microsoft Graph返回的文件不是有效PDF")
	}
	return pdf, false, nil
}

func (c *Client) delete(ctx context.Context, token, itemID string) {
	cleanupCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	endpoint := fmt.Sprintf("%s/drives/%s/items/%s", strings.TrimRight(c.graphBaseURL, "/"), url.PathEscape(c.cfg.DriveID), url.PathEscape(itemID))
	req, err := http.NewRequestWithContext(cleanupCtx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.httpClient.Do(req)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
}
