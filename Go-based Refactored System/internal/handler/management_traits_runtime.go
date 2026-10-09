package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
)

type ManagementTraitsRuntimeHandler struct {
	svc        *service.ManagementTraitsRuntimeService
	secret     string
	cfg        *config.Config
	templateMu sync.RWMutex
}

func NewManagementTraitsRuntimeHandler(db *gorm.DB, cfg *config.Config, runtimeServices ...*service.ManagementTraitsRuntimeService) *ManagementTraitsRuntimeHandler {
	secret := ""
	if cfg != nil {
		secret = cfg.Jwt.Secret
	}
	return &ManagementTraitsRuntimeHandler{svc: managementTraitsIdentityService(db, cfg, runtimeServices...), secret: secret, cfg: cfg}
}

func (h *ManagementTraitsRuntimeHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/participant/create-paper", h.CreatePaper)
	group.POST("/participant/paper-detail", h.PaperDetail)
	group.POST("/participant/fill-answer", h.FillAnswer)
	group.POST("/participant/submit", h.Submit)
	group.POST("/admin/candidate/resume", h.IssueCandidateResume)
	group.POST("/admin/exam/state", h.SetExamState)
	group.POST("/profile/freeze", h.FreezeProfile)
	group.GET("/profile/detail", h.ProfileDetail)
	group.POST("/results/list", h.ResultsList)
	group.GET("/results/detail", h.ResultDetail)
	h.registerTestReportRoutes(group)
	h.registerReissueRoutes(group)
	// No history-recompute, legacy upload, formal or activation routes.
}

func (h *ManagementTraitsRuntimeHandler) SetExamState(c *gin.Context) {
	if !managementTraitsRuntimeAdmin(c, "management-traits:exam:state") {
		return
	}
	body, ok := managementTraitsRuntimeBody(c, "examId", "state")
	if !ok || (body["state"] != "0" && body["state"] != "1") || len(c.Request.URL.Query()) != 0 {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：005 TEST状态仅允许进行中或禁用")
		return
	}
	state, err := strconv.Atoi(body["state"])
	if err != nil {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：005 TEST状态仅允许进行中或禁用")
		return
	}
	updated, err := h.svc.SetExamState(c.Request.Context(), body["examId"], state)
	managementTraitsRuntimeRespond(c, gin.H{"examId": body["examId"], "state": updated}, err)
}

// Flat string-only participant requests: reject arrays, duplicates, aliases,
// unknown fields, nulls and trailing JSON before authentication or DB work.
func managementTraitsRuntimeBody(c *gin.Context, names ...string) (map[string]string, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	d := json.NewDecoder(c.Request.Body)
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return nil, false
	}
	allowed := make(map[string]bool, len(names))
	for _, n := range names {
		allowed[n] = true
	}
	values := make(map[string]string, len(names))
	for d.More() {
		t, err := d.Token()
		key, ok := t.(string)
		if err != nil || !ok || !allowed[key] {
			return nil, false
		}
		if _, exists := values[key]; exists {
			return nil, false
		}
		value, err := d.Token()
		s, ok := value.(string)
		if err != nil || !ok || s == "" || len(s) > 64 || strings.TrimSpace(s) != s {
			return nil, false
		}
		values[key] = s
	}
	end, err := d.Token()
	if err != nil || end != json.Delim('}') || len(values) != len(names) {
		return nil, false
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, false
	}
	return values, true
}

func managementTraitsRuntimeHTTPError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"code": status, "msg": message, "success": false, "data": nil})
}

func (h *ManagementTraitsRuntimeHandler) closed(c *gin.Context) {
	// This gate is not an activation option. Old writers, complete schema
	// signatures and authenticated identity issuance have not been integrated.
	managementTraitsRuntimeHTTPError(c, http.StatusServiceUnavailable, "管理特质候选运行入口尚未启用：旧写保护和结构门禁未完成")
}

func (h *ManagementTraitsRuntimeHandler) participant(c *gin.Context, purpose, id string) bool {
	raw := c.GetHeader("X-Management-Traits-Token")
	if raw == "" {
		raw = c.GetHeader("Authorization")
	}
	claims, err := service.ParseManagementTraitsRuntimeToken(h.secret, raw, purpose, time.Now())
	if err != nil || (purpose == service.ManagementTraitsRuntimeParticipantPurpose && claims.ExamID != id) || (purpose == service.ManagementTraitsRuntimePaperPurpose && claims.PaperID != id) {
		managementTraitsRuntimeHTTPError(c, http.StatusUnauthorized, "管理特质身份验证失败")
		return false
	}
	c.Set("managementTraitsClaims", claims)
	return true
}

func (h *ManagementTraitsRuntimeHandler) CreatePaper(c *gin.Context) {
	body, ok := managementTraitsRuntimeBody(c, "examId")
	if !ok {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误")
		return
	}
	if h.participant(c, service.ManagementTraitsRuntimeParticipantPurpose, body["examId"]) {
		claims := c.MustGet("managementTraitsClaims").(service.ManagementTraitsRuntimeClaims)
		data, err := h.svc.CreatePaper(c.Request.Context(), claims)
		managementTraitsRuntimeRespond(c, data, err)
	}
}

func (h *ManagementTraitsRuntimeHandler) PaperDetail(c *gin.Context) {
	body, ok := managementTraitsRuntimeBody(c, "paperId")
	if !ok {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误")
		return
	}
	if h.participant(c, service.ManagementTraitsRuntimePaperPurpose, body["paperId"]) {
		claims := c.MustGet("managementTraitsClaims").(service.ManagementTraitsRuntimeClaims)
		data, err := h.svc.PaperDetail(c.Request.Context(), claims)
		data, err = h.capPaperCredential(data, claims, err)
		c.Header("Cache-Control", "no-store")
		managementTraitsRuntimeRespond(c, data, err)
	}
}

func (h *ManagementTraitsRuntimeHandler) FillAnswer(c *gin.Context) {
	body, ok := managementTraitsRuntimeBody(c, "paperId", "paperQuestionId", "optionId")
	if !ok {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：仅允许一个题目和一个选项")
		return
	}
	if h.participant(c, service.ManagementTraitsRuntimePaperPurpose, body["paperId"]) {
		claims := c.MustGet("managementTraitsClaims").(service.ManagementTraitsRuntimeClaims)
		count, err := h.svc.FillAnswer(c.Request.Context(), claims, body["paperQuestionId"], body["optionId"])
		managementTraitsRuntimeRespond(c, gin.H{"answered": count}, err)
	}
}

func (h *ManagementTraitsRuntimeHandler) Submit(c *gin.Context) {
	body, ok := managementTraitsRuntimeBody(c, "paperId", "submitType")
	if !ok || body["submitType"] != "manual" {
		managementTraitsRuntimeHTTPError(c, 400, "仅允许manual提交，超时由服务器判定")
		return
	}
	if h.participant(c, service.ManagementTraitsRuntimePaperPurpose, body["paperId"]) {
		claims := c.MustGet("managementTraitsClaims").(service.ManagementTraitsRuntimeClaims)
		data, err := h.svc.SubmitParticipant(c.Request.Context(), claims, body["submitType"])
		managementTraitsRuntimeRespond(c, data, err)
	}
}

func managementTraitsRuntimeAdmin(c *gin.Context, permission string) bool {
	value, exists := c.Get("loginUser")
	if !exists {
		managementTraitsRuntimeHTTPError(c, 401, "请先登录")
		return false
	}
	user, ok := value.(*model.LoginUser)
	if ok && user != nil {
		if user.UserID == 1 {
			return true
		}
		wildcard := false
		for _, p := range user.Permissions {
			wildcard = wildcard || p == "*:*:*"
		}
		// No verified non-administrator exam-scope mechanism exists here.
		if wildcard {
			return true
		}
	}
	managementTraitsRuntimeHTTPError(c, 403, "仅管理员可操作管理特质候选测试")
	return false
}

func (h *ManagementTraitsRuntimeHandler) FreezeProfile(c *gin.Context) {
	if managementTraitsRuntimeAdmin(c, "management-traits:profile:freeze") {
		body, ok := managementTraitsRuntimeBody(c, "examId")
		if !ok {
			managementTraitsRuntimeHTTPError(c, 400, "参数格式错误")
			return
		}
		data, err := h.svc.FreezeProfile(c.Request.Context(), body["examId"])
		managementTraitsRuntimeRespond(c, data, err)
	}
}

func (h *ManagementTraitsRuntimeHandler) ProfileDetail(c *gin.Context) {
	if managementTraitsRuntimeAdmin(c, "management-traits:profile:read") {
		id := c.Query("examId")
		if id == "" || len(id) > 64 || len(c.Request.URL.Query()) != 1 {
			managementTraitsRuntimeHTTPError(c, 400, "参数格式错误")
			return
		}
		data, err := h.svc.ProfileDetail(c.Request.Context(), id)
		managementTraitsRuntimeRespond(c, data, err)
	}
}

func (h *ManagementTraitsRuntimeHandler) ResultsList(c *gin.Context) {
	if managementTraitsRuntimeAdmin(c, "management-traits:result:read") {
		body, ok := managementTraitsRuntimeBody(c, "examId")
		if !ok {
			managementTraitsRuntimeHTTPError(c, 400, "参数格式错误")
			return
		}
		data, err := h.svc.ListRuns(c.Request.Context(), body["examId"])
		managementTraitsRuntimeRespond(c, data, err)
	}
}

func (h *ManagementTraitsRuntimeHandler) ResultDetail(c *gin.Context) {
	if managementTraitsRuntimeAdmin(c, "management-traits:result:read") {
		id := c.Query("runId")
		if id == "" || len(id) > 64 || len(c.Request.URL.Query()) != 1 || len(c.QueryArray("runId")) != 1 {
			managementTraitsRuntimeHTTPError(c, 400, "须选择明确runId")
			return
		}
		data, err := h.svc.ResultDetail(c.Request.Context(), id)
		managementTraitsRuntimeRespond(c, data, err)
	}
}

func managementTraitsRuntimeRespond(c *gin.Context, data any, err error) {
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "", "success": true, "data": data})
		return
	}
	switch {
	case errors.Is(err, service.ErrManagementTraitsRuntimeClosed):
		managementTraitsRuntimeHTTPError(c, 503, "管理特质TEST结构尚未完整安装，请联系管理员")
	case errors.Is(err, service.ErrManagementTraitsRuntimeToken):
		managementTraitsRuntimeHTTPError(c, 401, "管理特质身份验证失败")
	case errors.Is(err, service.ErrManagementTraitsRuntimeMissing):
		managementTraitsRuntimeHTTPError(c, 409, "请完成全部140道题目后交卷")
	case errors.Is(err, service.ErrManagementTraitsRuntimeExpired):
		managementTraitsRuntimeHTTPError(c, 409, "答题已到期，请提交试卷")
	default:
		managementTraitsRuntimeHTTPError(c, 409, "管理特质数据校验失败，请联系管理员")
	}
}
