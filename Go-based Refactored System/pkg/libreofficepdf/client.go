package libreofficepdf

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	maxDOCXBytes = 20 << 20
	maxPDFBytes  = 50 << 20
)

var conversionSlot = make(chan struct{}, 1)

type commandRunner interface {
	Run(ctx context.Context, command string, args ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, command string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, command, args...).CombinedOutput()
}

type Client struct {
	executable string
	runner     commandRunner
}

// ConversionError retains the cause for server-side classification, while
// Error returns only the existing fixed public message, never command output.
type ConversionError struct {
	message string
	class   string
	cause   error
}

func (e *ConversionError) Error() string           { return e.message }
func (e *ConversionError) Unwrap() error           { return e.cause }
func (e *ConversionError) DiagnosticClass() string { return e.class }

func conversionError(message, class string, cause error) error {
	return &ConversionError{message: message, class: class, cause: cause}
}

func NewClient(executable string) *Client {
	return newClient(executable, execRunner{})
}

func newClient(executable string, runner commandRunner) *Client {
	executable = strings.TrimSpace(executable)
	if executable == "" {
		if runtime.GOOS == "windows" {
			executable = `C:\Program Files\LibreOffice\program\soffice.com`
		} else {
			executable = "libreoffice"
		}
	}
	if runtime.GOOS == "windows" && strings.EqualFold(filepath.Ext(executable), ".exe") {
		consoleExecutable := strings.TrimSuffix(executable, filepath.Ext(executable)) + ".com"
		if info, err := os.Stat(consoleExecutable); err == nil && !info.IsDir() {
			executable = consoleExecutable
		}
	}
	return &Client{executable: executable, runner: runner}
}

func (c *Client) Convert(ctx context.Context, fileName string, docx []byte) ([]byte, error) {
	if c == nil || c.runner == nil || strings.TrimSpace(c.executable) == "" {
		return nil, conversionError("LibreOffice报告转换未配置", "lo_config", nil)
	}
	if len(docx) == 0 || len(docx) > maxDOCXBytes {
		return nil, conversionError("Word报告文件大小无效", "lo_input_size", nil)
	}
	fileName = strings.TrimSpace(fileName)
	if fileName == "" || filepath.Base(fileName) != fileName || strings.ContainsAny(fileName, `/\`) || !strings.EqualFold(filepath.Ext(fileName), ".docx") {
		return nil, conversionError("Word报告文件名无效", "lo_input_name", nil)
	}
	select {
	case conversionSlot <- struct{}{}:
		defer func() { <-conversionSlot }()
	case <-ctx.Done():
		return nil, conversionError("LibreOffice转换排队超时", "lo_queue", ctx.Err())
	}
	workspace, err := os.MkdirTemp("", "phase1-word-pdf-")
	if err != nil {
		return nil, conversionError("创建Word报告转换目录失败", "lo_workspace", err)
	}
	defer os.RemoveAll(workspace)
	if err := os.Chmod(workspace, 0o700); err != nil {
		return nil, conversionError("保护Word报告转换目录失败", "lo_workspace_permission", err)
	}
	docxPath := filepath.Join(workspace, fileName)
	if err := os.WriteFile(docxPath, docx, 0o600); err != nil {
		return nil, conversionError("写入Word报告临时文件失败", "lo_input_write", err)
	}
	profilePath := filepath.Join(workspace, "profile")
	profileURL := localFileURL(profilePath)
	_, err = c.runner.Run(ctx, c.executable,
		"-env:UserInstallation="+profileURL,
		"--headless",
		"--convert-to", "pdf",
		"--outdir", workspace,
		docxPath,
	)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, conversionError("LibreOffice转换PDF超时", "lo_command", errors.Join(ctx.Err(), err))
		}
		return nil, conversionError("LibreOffice转换PDF失败", "lo_command", errors.Join(ctx.Err(), err))
	}
	pdfName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".pdf"
	pdfFile, err := os.Open(filepath.Join(workspace, pdfName))
	if err != nil {
		return nil, conversionError("LibreOffice转换结果不存在", "lo_output_open", err)
	}
	defer pdfFile.Close()
	pdf, err := io.ReadAll(io.LimitReader(pdfFile, maxPDFBytes+1))
	if err != nil || len(pdf) > maxPDFBytes {
		return nil, conversionError("读取LibreOffice PDF失败", "lo_output_read", err)
	}
	if len(pdf) < 1024 || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		return nil, conversionError("LibreOffice返回的文件不是有效PDF", "lo_output_invalid", nil)
	}
	return pdf, nil
}

func localFileURL(filePath string) string {
	urlPath := filepath.ToSlash(filePath)
	if runtime.GOOS == "windows" && !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}
	return (&url.URL{Scheme: "file", Path: urlPath}).String()
}
