package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
)

func (h *ManagementTraitsRuntimeHandler) registerTestReportRoutes(group *gin.RouterGroup) {
	group.POST("/reports/generate-test", h.GenerateTestReport)
	group.GET("/reports/view", h.ViewTestReport)
	group.GET("/reports/download", h.DownloadTestReport)
	group.GET("/reports/template", h.ManagementTraitsTemplateInfo)
	group.GET("/reports/template/download", h.DownloadManagementTraitsTemplate)
	group.POST("/reports/template/upload", h.UploadManagementTraitsTemplate)
}

func (h *ManagementTraitsRuntimeHandler) testReportInputs(loadTemplate bool) (string, service.ManagementTraitsTestContent, []byte, error) {
	root, content, template, _, err := h.testReportInputsWithTemplateSHA(loadTemplate)
	return root, content, template, err
}

func (h *ManagementTraitsRuntimeHandler) testReportInputsWithTemplateSHA(loadTemplate bool) (string, service.ManagementTraitsTestContent, []byte, string, error) {
	if !config.ManagementTraitsTestRuntimeEnabled() {
		return "", service.ManagementTraitsTestContent{}, nil, "", service.ErrManagementTraitsRuntimeClosed
	}
	// Mandatory private server directory, never a browser-supplied path. There
	// is no activation/formal option; lack of configuration fails closed.
	root := os.Getenv("MNG_TEST_REPORT_DIR")
	if !filepath.IsAbs(root) {
		return "", service.ManagementTraitsTestContent{}, nil, "", service.ErrManagementTraitsRuntimeClosed
	}
	contentPath := os.Getenv("MNG_TEST_CONTENT_PATH")
	if contentPath == "" {
		contentPath = filepath.Join("configs", "export-templates", "management-traits-002-test-content-v1.xlsx")
	}
	templatePath := os.Getenv("MNG_TEST_TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = filepath.Join("configs", "export-templates", "management-traits-002-test-only-v2.docx")
	}
	raw, err := os.ReadFile(contentPath)
	if err != nil {
		return "", service.ManagementTraitsTestContent{}, nil, "", service.ErrManagementTraitsRuntimeClosed
	}
	content, err := service.LoadManagementTraitsTestContent(raw)
	if err != nil {
		return "", content, nil, "", service.ErrManagementTraitsRuntimeInvalid
	}
	if !loadTemplate {
		return root, content, nil, "", nil
	}
	h.templateMu.RLock()
	template, err := os.ReadFile(templatePath)
	h.templateMu.RUnlock()
	if err != nil {
		return "", content, nil, "", service.ErrManagementTraitsRuntimeClosed
	}
	digest := sha256.Sum256(template)
	return root, content, template, hex.EncodeToString(digest[:]), nil
}

func (h *ManagementTraitsRuntimeHandler) GenerateTestReport(c *gin.Context) {
	if !managementTraitsRuntimeAdmin(c, "management-traits:report:generate-test") {
		return
	}
	body, ok := managementTraitsRuntimeBody(c, "runId")
	if !ok {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：须选择明确run")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	root, content, template, templateSHA, err := h.testReportInputsWithTemplateSHA(true)
	if err != nil {
		managementTraitsRuntimeRespond(c, nil, err)
		return
	}
	actor := c.MustGet("loginUser").(*model.LoginUser).UserID
	lo := ""
	if h.cfg != nil {
		lo = h.cfg.Phase1WordReport.LibreOfficePath
	}
	render := func(ctx context.Context, dto service.ManagementTraitsTestReportData) ([]byte, error) {
		docx, err := renderManagementTraitsTestWordWithSHA(template, templateSHA, dto)
		if err != nil {
			return nil, err
		}
		return libreofficepdf.NewClient(lo).Convert(ctx, "mng-test-report.docx", docx)
	}
	data, err := h.svc.GenerateTestReportWithTemplateSHA(ctx, body["runId"], actor, root, content, templateSHA, render)
	managementTraitsRuntimeRespond(c, data, err)
}

func (h *ManagementTraitsRuntimeHandler) ViewTestReport(c *gin.Context) {
	h.serveTestReport(c, "inline")
}
func (h *ManagementTraitsRuntimeHandler) DownloadTestReport(c *gin.Context) {
	h.serveTestReport(c, "attachment")
}
func (h *ManagementTraitsRuntimeHandler) serveTestReport(c *gin.Context, disposition string) {
	if !managementTraitsRuntimeAdmin(c, "management-traits:report:read") {
		return
	}
	id := c.Query("reportId")
	if id == "" || len(id) > 64 || len(c.Request.URL.Query()) != 1 || len(c.QueryArray("reportId")) != 1 {
		managementTraitsRuntimeHTTPError(c, 400, "须选择明确reportId，不接受文件路径")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	root, content, _, err := h.testReportInputs(false)
	if err != nil {
		managementTraitsRuntimeRespond(c, nil, err)
		return
	}
	actor := c.MustGet("loginUser").(*model.LoginUser).UserID
	allowedTemplateSHAs, err := h.managementTraitsTemplateSHAs()
	if err != nil {
		managementTraitsRuntimeRespond(c, nil, service.ErrManagementTraitsRuntimeInvalid)
		return
	}
	report, pdf, err := h.svc.ReadTestReportPDFWithTemplateSHAs(ctx, id, actor, root, content, allowedTemplateSHAs)
	if err != nil {
		managementTraitsRuntimeRespond(c, nil, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", disposition+`; filename="management-traits-TEST.pdf"; filename*=UTF-8''`+url.PathEscape(report.TestTitle+"-"+report.ID+".pdf"))
	c.Header("Content-Length", strconv.Itoa(len(pdf)))
	c.Data(http.StatusOK, "application/pdf", pdf)
}
