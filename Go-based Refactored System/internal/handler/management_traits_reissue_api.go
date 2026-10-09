package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
)

func (h *ManagementTraitsRuntimeHandler) registerReissueRoutes(g *gin.RouterGroup) {
	g.GET("/report-reissues/qualification", h.QualifyReportReissue)
	g.GET("/report-reissues", h.ListReportReissues)
	g.POST("/report-reissues/generate", h.GenerateReportReissue)
	g.GET("/report-reissues/view", h.ViewReportReissue)
	g.GET("/report-reissues/download", h.DownloadReportReissue)
}

func managementTraitsReissueAdmin(c *gin.Context, permission string) bool {
	if value, ok := c.Get("loginUser"); ok {
		if u, ok := value.(*model.LoginUser); ok && u != nil && u.UserID <= 0 {
			managementTraitsRuntimeHTTPError(c, 403, "仅有效管理员可操作报告")
			return false
		}
	}
	return managementTraitsRuntimeAdmin(c, permission)
}

func managementTraitsReissueRespond(c *gin.Context, data any, err error) {
	if err == nil {
		c.JSON(200, gin.H{"code": 0, "msg": "", "success": true, "data": data})
		return
	}
	status, code, message := 409, "source_or_archive_invalid", "报告来源或归档校验失败"
	if errors.Is(err, service.ErrManagementTraitsReissueClosed) || errors.Is(err, service.ErrManagementTraitsRuntimeClosed) {
		status, code, message = 503, "report_service_not_installed", "报告服务未安装"
	}
	c.AbortWithStatusJSON(status, gin.H{"code": status, "errorCode": code, "msg": message, "success": false, "data": nil})
}

func managementTraitsReissueQuery(c *gin.Context, key string) (string, bool) {
	id := c.Query(key)
	q := c.Request.URL.Query()
	if id == "" || len(id) > 64 || len(q) != 1 || len(q[key]) != 1 {
		managementTraitsRuntimeHTTPError(c, 400, "须选择明确ID，不接受文件路径")
		return "", false
	}
	return id, true
}

func (h *ManagementTraitsRuntimeHandler) QualifyReportReissue(c *gin.Context) {
	if !managementTraitsReissueAdmin(c, "management-traits:report:read") {
		return
	}
	id, ok := managementTraitsReissueQuery(c, "runId")
	if !ok {
		return
	}
	_, content, _, err := h.testReportInputs(false)
	if err != nil {
		managementTraitsReissueRespond(c, nil, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	d, err := h.svc.QualifyReportReissue(ctx, id, content)
	managementTraitsReissueRespond(c, d, err)
}

func (h *ManagementTraitsRuntimeHandler) ListReportReissues(c *gin.Context) {
	if !managementTraitsReissueAdmin(c, "management-traits:report:read") {
		return
	}
	id, ok := managementTraitsReissueQuery(c, "paperId")
	if !ok {
		return
	}
	if _, _, _, err := h.testReportInputs(false); err != nil {
		managementTraitsReissueRespond(c, nil, err)
		return
	}
	d, err := h.svc.ListReportReissues(c.Request.Context(), id)
	managementTraitsReissueRespond(c, d, err)
}

func (h *ManagementTraitsRuntimeHandler) GenerateReportReissue(c *gin.Context) {
	if !managementTraitsReissueAdmin(c, "management-traits:report:generate-test") {
		return
	}
	body, ok := managementTraitsRuntimeBody(c, "runId")
	if !ok {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：仅接受runId")
		return
	}
	root, content, template, templateSHA, err := h.testReportInputsWithTemplateSHA(true)
	if err != nil {
		managementTraitsReissueRespond(c, nil, err)
		return
	}
	// Separate private subtree, never the old TEST current/file namespace.
	root = filepath.Join(root, "reissues")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	lo := ""
	if h.cfg != nil {
		lo = h.cfg.Phase1WordReport.LibreOfficePath
	}
	render := func(ctx context.Context, d service.ManagementTraitsTestReportData) ([]byte, error) {
		docx, e := renderManagementTraitsTestWordWithSHA(template, templateSHA, d)
		if e != nil {
			return nil, e
		}
		return libreofficepdf.NewClient(lo).Convert(ctx, "mng-reissue-TEST.docx", docx)
	}
	actor := c.MustGet("loginUser").(*model.LoginUser).UserID
	r, reused, err := h.svc.GenerateReportReissueWithTemplateSHA(ctx, body["runId"], actor, root, content, templateSHA, render)
	managementTraitsReissueRespond(c, gin.H{"report": r, "reused": reused, "purpose": service.ManagementTraitsTestPurposeLabel}, err)
}

func (h *ManagementTraitsRuntimeHandler) ViewReportReissue(c *gin.Context) {
	h.serveReportReissue(c, "view", "inline")
}
func (h *ManagementTraitsRuntimeHandler) DownloadReportReissue(c *gin.Context) {
	h.serveReportReissue(c, "download", "attachment")
}
func (h *ManagementTraitsRuntimeHandler) serveReportReissue(c *gin.Context, action, disposition string) {
	if !managementTraitsReissueAdmin(c, "management-traits:report:read") {
		return
	}
	id, ok := managementTraitsReissueQuery(c, "reportId")
	if !ok {
		return
	}
	root, _, _, err := h.testReportInputs(false)
	if err != nil {
		managementTraitsReissueRespond(c, nil, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	actor := c.MustGet("loginUser").(*model.LoginUser).UserID
	allowedTemplateSHAs, allowErr := h.managementTraitsTemplateSHAs()
	if allowErr != nil {
		managementTraitsReissueRespond(c, nil, service.ErrManagementTraitsReissueInvalid)
		return
	}
	r, pdf, err := h.svc.ReadReportReissueWithTemplateSHAs(ctx, id, actor, filepath.Join(root, "reissues"), action, allowedTemplateSHAs)
	if err != nil {
		managementTraitsReissueRespond(c, nil, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", disposition+`; filename="management-traits-TEST.pdf"; filename*=UTF-8''`+url.PathEscape(service.ManagementTraitsTestPurposeTitle+"-"+r.ID+".pdf"))
	c.Header("Content-Length", strconv.Itoa(len(pdf)))
	c.Data(http.StatusOK, "application/pdf", pdf)
}
