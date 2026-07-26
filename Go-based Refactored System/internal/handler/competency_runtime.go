package handler

import (
	"crypto/subtle"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

type CompetencyRuntimeHandler struct {
	svc *service.CompetencyRuntimeService
	cfg *config.Config
}

func NewCompetencyRuntimeHandler(db *gorm.DB, cfg *config.Config) *CompetencyRuntimeHandler {
	return &CompetencyRuntimeHandler{svc: service.NewCompetencyRuntimeService(db, cfg), cfg: cfg}
}

func (h *CompetencyRuntimeHandler) RuntimeService() *service.CompetencyRuntimeService { return h.svc }

func (h *CompetencyRuntimeHandler) Publish(c *gin.Context) {
	var body struct {
		ExamID string `json:"examId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ExamID == "" {
		response.RestErr(c, "examId 为空")
		return
	}
	var publishedBy *int64
	if value, ok := c.Get("loginUser"); ok {
		if login, ok := value.(*model.LoginUser); ok {
			id := login.UserID
			publishedBy = &id
		}
	}
	summary, err := h.svc.Publish(body.ExamID, publishedBy)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, summary)
}

func (h *CompetencyRuntimeHandler) CreatePaper(c *gin.Context) {
	var body struct {
		ExamID           string `json:"examId"`
		ParticipantID    string `json:"participantId"`
		ParticipantType  string `json:"participantType"`
		ParticipantToken string `json:"participantToken"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	token := body.ParticipantToken
	if token == "" {
		token = c.GetHeader(h.cfg.Jwt.Header)
	}
	claims, err := service.ParseCompetencyToken(h.cfg.Jwt.Secret, token, service.CompetencyTokenPurposeParticipant)
	if err != nil {
		response.RestErr(c, "参与者认证失败")
		return
	}
	if err := claims.ValidateBinding(body.ParticipantID, body.ExamID, ""); err != nil || claims.ParticipantType != body.ParticipantType {
		response.RestErr(c, "参与者与测评不匹配")
		return
	}
	access, err := h.svc.CreateOrRestorePaper(claims)
	if err != nil {
		if errors.Is(err, service.ErrCompetencyNotPublished) {
			response.RestErr(c, "测评尚未发布，请联系管理员先发布并冻结题目")
			return
		}
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, access)
}

func (h *CompetencyRuntimeHandler) paperClaims(c *gin.Context, paperID string) (service.CompetencyTokenClaims, error) {
	token := c.GetHeader(h.cfg.Jwt.Header)
	if token == "" {
		token = c.GetHeader("X-Competency-Token")
	}
	claims, err := service.ParseCompetencyToken(h.cfg.Jwt.Secret, token, service.CompetencyTokenPurposePaper)
	if err != nil {
		return claims, err
	}
	if claims.PaperID != paperID {
		return claims, errors.New("试卷令牌不匹配")
	}
	return claims, nil
}

func (h *CompetencyRuntimeHandler) PaperDetail(c *gin.Context) {
	var body struct {
		PaperID string `json:"paperId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	claims, err := h.paperClaims(c, body.PaperID)
	if err != nil {
		response.RestErr(c, "试卷认证失败")
		return
	}
	view, err := h.svc.PaperDetail(claims)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, view)
}

func (h *CompetencyRuntimeHandler) FillAnswer(c *gin.Context) {
	var body struct {
		PaperID         string `json:"paperId"`
		PaperQuestionID string `json:"paperQuestionId"`
		RawValue        int    `json:"rawValue"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.PaperID == "" || body.PaperQuestionID == "" {
		response.RestErr(c, "参数错误")
		return
	}
	claims, err := h.paperClaims(c, body.PaperID)
	if err != nil {
		response.RestErr(c, "试卷认证失败")
		return
	}
	count, err := h.svc.FillAnswer(claims, body.PaperQuestionID, body.RawValue)
	if errors.Is(err, service.ErrCompetencyPaperExpired) {
		summary, submitErr := h.svc.Submit(claims, service.CompetencySubmitTimeout)
		if submitErr != nil {
			response.RestErr(c, submitErr.Error())
			return
		}
		response.Rest(c, gin.H{"expired": true, "submittedAt": summary.SubmittedAt})
		return
	}
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, gin.H{"answeredCount": count})
}

func (h *CompetencyRuntimeHandler) Submit(c *gin.Context) {
	var body struct {
		PaperID    string `json:"paperId"`
		SubmitType string `json:"submitType"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.PaperID == "" {
		response.RestErr(c, "参数错误")
		return
	}
	if body.SubmitType != "" && body.SubmitType != service.CompetencySubmitManual {
		response.RestErr(c, "参与者只能手工提交，超时由系统判定")
		return
	}
	claims, err := h.paperClaims(c, body.PaperID)
	if err != nil {
		response.RestErr(c, "试卷认证失败")
		return
	}
	summary, err := h.svc.Submit(claims, service.CompetencySubmitManual)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, summary)
}

func (h *CompetencyRuntimeHandler) ResultsPaging(c *gin.Context) {
	var body struct {
		ExamID        string `json:"examId"`
		Current       int    `json:"current"`
		Size          int    `json:"size"`
		SortBy        string `json:"sortBy"`
		SortDirection string `json:"sortDirection"`
		DimensionID   string `json:"dimensionId"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Current < 1 {
		body.Current = 1
	}
	if body.Size < 1 {
		body.Size = 10
	}
	if body.Size > 500 {
		body.Size = 500
	}
	request := service.CompetencyResultPageRequest{
		ExamID: body.ExamID, Current: body.Current, Size: body.Size,
		SortBy: body.SortBy, SortDirection: body.SortDirection, DimensionID: body.DimensionID,
	}
	rows, total, err := h.svc.ResultPaging(request)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, gin.H{"records": rows, "total": total, "current": request.Current, "size": request.Size})
}
func (h *CompetencyRuntimeHandler) ResultDetail(c *gin.Context) {
	var body struct {
		PaperID string `json:"paperId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	data, err := h.svc.ResultDetail(body.PaperID)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, data)
}
func (h *CompetencyRuntimeHandler) AdminReportData(c *gin.Context) {
	paperID := c.Query("paperId")
	if paperID == "" {
		paperID = c.PostForm("paperId")
	}
	if paperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	data, err := h.svc.FormalReportData(paperID)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, data)
}
func (h *CompetencyRuntimeHandler) InternalReportData(c *gin.Context) {
	token := c.GetHeader("X-Internal-Token")
	if token == "" {
		token = c.Query("token")
	}
	if h.cfg.PdfGen.InternalToken == "" || subtle.ConstantTimeCompare([]byte(token), []byte(h.cfg.PdfGen.InternalToken)) != 1 {
		response.AjaxUnauthorized(c, "内部报告认证失败")
		return
	}
	h.AdminReportData(c)
}

func createParticipantToken(cfg *config.Config, participantType, participantID, examID string) (string, error) {
	ttl := time.Duration(cfg.Jwt.ExpireMinutes) * time.Minute
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return service.CreateCompetencyToken(cfg.Jwt.Secret, service.CompetencyTokenClaims{Purpose: service.CompetencyTokenPurposeParticipant, ParticipantType: participantType, ParticipantID: participantID, ExamID: examID}, ttl)
}
func parseInt64(value string) *int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}
