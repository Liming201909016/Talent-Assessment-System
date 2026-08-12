package handler

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
)

const phase1WordTemplateFileName = "competency-phase1-report.docx"

type phase1WordTemplateContract struct {
	ContentControls int `json:"contentControls"`
	Charts          int `json:"charts"`
	VisibleTokens   int `json:"visibleTokens"`
}

type phase1WordTemplateInfo struct {
	Exists          bool   `json:"exists"`
	FileName        string `json:"fileName"`
	Size            int64  `json:"size"`
	ModTime         string `json:"modTime"`
	SHA256          string `json:"sha256"`
	Valid           bool   `json:"valid"`
	ValidationError string `json:"validationError,omitempty"`
	phase1WordTemplateContract
}

func (h *CompetencyReportHandler) templateIdentity(c *gin.Context) (*model.LoginUser, bool) {
	value, ok := c.Get("loginUser")
	if !ok {
		response.AjaxUnauthorized(c, "")
		return nil, false
	}
	login, ok := value.(*model.LoginUser)
	if !ok || !canPublishCompetencyExam(login) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "无权管理胜任力报告模板"})
		return nil, false
	}
	return login, true
}

func (h *CompetencyReportHandler) phase1WordTemplatePath() (string, error) {
	if h == nil || h.examH == nil || h.examH.cfg == nil {
		return "", errors.New("一期Word报告模板未配置")
	}
	path := strings.TrimSpace(h.examH.cfg.Phase1WordReport.TemplatePath)
	if path == "" || !strings.EqualFold(filepath.Ext(path), ".docx") {
		return "", errors.New("一期Word报告模板路径无效")
	}
	return filepath.Clean(path), nil
}

func (h *CompetencyReportHandler) Phase1TemplateInfo(c *gin.Context) {
	if _, ok := h.templateIdentity(c); !ok {
		return
	}
	path, err := h.phase1WordTemplatePath()
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	h.templateMu.RLock()
	defer h.templateMu.RUnlock()
	info, err := readPhase1WordTemplateInfo(path)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, info)
}

func (h *CompetencyReportHandler) DownloadPhase1Template(c *gin.Context) {
	if _, ok := h.templateIdentity(c); !ok {
		return
	}
	path, err := h.phase1WordTemplatePath()
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	h.templateMu.RLock()
	defer h.templateMu.RUnlock()
	file, err := os.Open(path)
	if err != nil {
		response.RestErr(c, "一期Word报告模板不存在")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		response.RestErr(c, "读取一期Word报告模板失败")
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", "attachment; filename="+phase1WordTemplateFileName+"; filename*=UTF-8''"+encodeRFC5987FileName(phase1WordTemplateFileName))
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if _, err := io.Copy(c.Writer, file); err != nil {
		return
	}
}

func (h *CompetencyReportHandler) UploadPhase1Template(c *gin.Context) {
	if _, ok := h.templateIdentity(c); !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxPhase1WordTemplateBytes+(1<<20))
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.RestErr(c, "请选择一期Word报告模板")
		return
	}
	if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".docx") {
		response.RestErr(c, "只支持.docx格式")
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxPhase1WordTemplateBytes {
		response.RestErr(c, "模板文件大小必须在20MB以内")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.RestErr(c, "读取上传模板失败")
		return
	}
	data, readErr := io.ReadAll(io.LimitReader(file, maxPhase1WordTemplateBytes+1))
	file.Close()
	if readErr != nil || len(data) == 0 || len(data) > maxPhase1WordTemplateBytes {
		response.RestErr(c, "读取上传模板失败")
		return
	}
	contract, err := validatePhase1WordTemplateUpload(data)
	if err != nil {
		response.RestErr(c, "模板校验失败："+err.Error())
		return
	}
	path, err := h.phase1WordTemplatePath()
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	h.templateMu.Lock()
	backup, err := installPhase1WordTemplate(path, data, time.Now())
	h.templateMu.Unlock()
	if err != nil {
		response.RestErr(c, "保存模板失败")
		return
	}
	info, err := readPhase1WordTemplateInfo(path)
	if err != nil {
		response.RestErr(c, "读取已保存模板失败")
		return
	}
	info.phase1WordTemplateContract = contract
	response.Rest(c, gin.H{"template": info, "backupFile": filepath.Base(backup)})
}

func readPhase1WordTemplateInfo(path string) (phase1WordTemplateInfo, error) {
	info := phase1WordTemplateInfo{FileName: phase1WordTemplateFileName}
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return info, nil
	}
	if err != nil || stat.IsDir() {
		return info, errors.New("读取一期Word报告模板失败")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return info, errors.New("读取一期Word报告模板失败")
	}
	digest := sha256.Sum256(data)
	info.Exists = true
	info.Size = stat.Size()
	info.ModTime = stat.ModTime().Format("2006-01-02 15:04:05")
	info.SHA256 = hex.EncodeToString(digest[:])
	contract, validationErr := validatePhase1WordTemplateUpload(data)
	if validationErr != nil {
		info.ValidationError = validationErr.Error()
		return info, nil
	}
	info.Valid = true
	info.phase1WordTemplateContract = contract
	return info, nil
}

func validatePhase1WordTemplateUpload(data []byte) (phase1WordTemplateContract, error) {
	contract := phase1WordTemplateContract{}
	if len(data) == 0 || len(data) > maxPhase1WordTemplateBytes {
		return contract, errors.New("模板文件大小无效")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return contract, errors.New("DOCX文件结构无效")
	}
	parts := make(map[string][]byte, 13)
	for _, file := range reader.File {
		if file.Name != "word/document.xml" && !strings.HasPrefix(file.Name, "word/charts/chart") {
			continue
		}
		rc, openErr := file.Open()
		if openErr != nil {
			return contract, errors.New("读取DOCX内容失败")
		}
		body, readErr := io.ReadAll(io.LimitReader(rc, maxPhase1WordTemplateBytes+1))
		rc.Close()
		if readErr != nil || len(body) > maxPhase1WordTemplateBytes {
			return contract, errors.New("读取DOCX内容失败")
		}
		parts[file.Name] = body
	}
	document, ok := parts["word/document.xml"]
	if !ok {
		return contract, errors.New("模板缺少Word正文")
	}
	tags := wordContentControlTagPattern.FindAllSubmatch(document, -1)
	seen := make(map[string]int, len(tags))
	for _, match := range tags {
		seen[string(match[1])]++
	}
	for tag, count := range seen {
		if count > 1 {
			return contract, fmt.Errorf("内容控件Tag重复：%s", tag)
		}
	}
	required := requiredPhase1WordTemplateTags()
	missing := make([]string, 0)
	for _, tag := range required {
		if seen[tag] != 1 {
			missing = append(missing, tag)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return contract, fmt.Errorf("缺少内容控件Tag：%s", strings.Join(missing, ", "))
	}
	if len(seen) != len(required) {
		return contract, errors.New("模板包含未支持的内容控件Tag")
	}
	visibleTokens := wordTemplateTokenPattern.FindAll(document, -1)
	if len(visibleTokens) > 0 {
		return contract, fmt.Errorf("模板存在可见占位符：%s", visibleTokens[0])
	}
	for chartIndex := 1; chartIndex <= 12; chartIndex++ {
		name := fmt.Sprintf("word/charts/chart%d.xml", chartIndex)
		chart, exists := parts[name]
		if !exists {
			return contract, fmt.Errorf("模板缺少图表：chart%d", chartIndex)
		}
		valueCount := 2
		if chartIndex == 2 {
			valueCount = 10
		}
		if _, err := replaceWordChartValues(chart, make([]float64, valueCount)); err != nil {
			return contract, fmt.Errorf("图表数据点无效：chart%d", chartIndex)
		}
	}
	contract.ContentControls = len(seen)
	contract.Charts = 12
	contract.VisibleTokens = len(visibleTokens)
	return contract, nil
}

func requiredPhase1WordTemplateTags() []string {
	tags := []string{
		"report.date", "participant.name", "participant.age", "participant.gender", "participant.telephone",
		"participant.affiliation", "participant.post", "result.submittedAt", "result.userTime", "overall.level",
		"overall.diagnosis", "validity.notice", "report.disclaimer", "group.general_ability.score",
		"group.general_ability.level", "group.general_ability.description", "group.psychological_quality.score",
		"group.psychological_quality.level", "group.psychological_quality.description",
	}
	for _, dimensionID := range service.NormalizePhase1CompetencyConfiguration().DimensionIDs {
		for _, field := range []string{"score", "level", "diagnosis"} {
			tags = append(tags, "dimension."+dimensionID+"."+field)
		}
	}
	return tags
}

func installPhase1WordTemplate(target string, data []byte, now time.Time) (string, error) {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return "", err
	}
	backup := filepath.Join(backupDir, phase1WordTemplateFileName+"."+now.Format("20060102_150405_000")+".bak")
	var current []byte
	if existing, err := os.ReadFile(target); err == nil {
		current = append([]byte(nil), existing...)
		if err := os.WriteFile(backup, current, 0o600); err != nil {
			return "", err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	} else {
		backup = ""
	}
	temporary, err := os.CreateTemp(dir, ".phase1-word-template-*.docx")
	if err != nil {
		return "", err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return "", err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		if runtime.GOOS == "windows" && len(current) > 0 {
			_ = os.WriteFile(target, current, 0o644)
		}
		return "", err
	}
	return backup, nil
}
