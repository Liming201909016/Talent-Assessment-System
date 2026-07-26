package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

const (
	competencyReportStatusGenerating = "generating"
	competencyReportStatusCompleted  = "completed"
	competencyReportStatusFailed     = "failed"
)

type CompetencyReportHandler struct {
	db    *gorm.DB
	examH *ExamHandler
}

func NewCompetencyReportHandler(db *gorm.DB, examH *ExamHandler) *CompetencyReportHandler {
	return &CompetencyReportHandler{db: db, examH: examH}
}

func (h *CompetencyReportHandler) reportIdentity(c *gin.Context) (*model.LoginUser, bool) {
	value, ok := c.Get("loginUser")
	if !ok {
		response.AjaxUnauthorized(c, "")
		return nil, false
	}
	login, ok := value.(*model.LoginUser)
	if !ok || login == nil {
		response.AjaxUnauthorized(c, "")
		return nil, false
	}
	if !h.examH.canGenerateReport(login) {
		response.Ajax(c, 403, "无权生成或下载胜任力报告", nil)
		return nil, false
	}
	return login, true
}

func (h *CompetencyReportHandler) Generate(c *gin.Context) {
	login, ok := h.reportIdentity(c)
	if !ok {
		return
	}
	if h.examH.pdfPool == nil {
		response.RestErr(c, "后端报告生成未启用（pdfgen.enabled=false）")
		return
	}
	var body struct {
		PaperID string `json:"paperId"`
		Force   bool   `json:"force"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.PaperID) == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	body.PaperID = strings.TrimSpace(body.PaperID)

	runtime := service.NewCompetencyRuntimeService(h.db, h.examH.cfg)
	data, err := runtime.FormalReportData(body.PaperID)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	result, ok := data["result"].(model.CompetencyResult)
	if !ok {
		response.RestErr(c, "报告结果数据异常")
		return
	}

	var existing model.CompetencyReport
	err = h.db.Where("paper_id = ? AND content_version = ?", body.PaperID, service.CompetencyTemporaryContentVersion).Take(&existing).Error
	if err == nil && !body.Force && existing.Status == competencyReportStatusCompleted && existing.PDFPath != "" {
		if info, statErr := os.Stat(existing.PDFPath); statErr == nil && info.Size() > 0 {
			if err := h.writeReportAudit(c, existing.ID, body.PaperID, "reuse", &login.UserID, 1, ""); err != nil {
				response.RestErr(c, "记录报告审计失败")
				return
			}
			response.Rest(c, existing)
			return
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.RestErr(c, "查询报告实例失败")
		return
	}

	now := time.Now()
	report := existing
	if report.ID == "" {
		report.ID = uuid.NewString()
		report.PaperID = body.PaperID
		report.ExamID = result.ExamID
		report.ContentVersion = service.CompetencyTemporaryContentVersion
		report.CreateTime = &now
	}
	report.Audience = result.ReportAudience
	report.Status = competencyReportStatusGenerating
	report.ErrorMessage = ""
	report.GeneratedBy = &login.UserID
	report.UpdateTime = &now
	textSnapshot, _ := json.Marshal(data["reportText"])
	scoreSnapshot, _ := json.Marshal(data)
	report.TextSnapshot = string(textSnapshot)
	report.ScoreSnapshot = string(scoreSnapshot)
	if existing.ID == "" {
		if err := h.db.Create(&report).Error; err != nil {
			response.RestErr(c, "创建报告实例失败")
			return
		}
	} else if err := h.db.Model(&model.CompetencyReport{}).Where("id = ?", report.ID).Updates(map[string]any{
		"audience": report.Audience, "text_snapshot": report.TextSnapshot, "score_snapshot": report.ScoreSnapshot,
		"status": report.Status, "error_message": "", "generated_by": report.GeneratedBy, "update_time": &now,
	}).Error; err != nil {
		response.RestErr(c, "更新报告实例失败")
		return
	}

	pdfBytes, err := h.renderCompetencyReport(body.PaperID)
	if err != nil {
		h.markReportFailed(report.ID, err)
		_ = h.writeReportAudit(c, report.ID, body.PaperID, "generate", &login.UserID, 0, err.Error())
		response.RestErr(c, "生成胜任力报告失败: "+err.Error())
		return
	}
	saved, digest, err := h.saveCompetencyReportPDF(result.ParticipantName, body.PaperID, pdfBytes)
	if err != nil {
		h.markReportFailed(report.ID, err)
		_ = h.writeReportAudit(c, report.ID, body.PaperID, "generate", &login.UserID, 0, err.Error())
		response.RestErr(c, "保存胜任力报告失败")
		return
	}
	generatedAt := time.Now()
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.CompetencyReport{}).Where("id = ?", report.ID).Updates(map[string]any{
			"pdf_path": saved, "pdf_sha256": digest, "pdf_size": len(pdfBytes), "status": competencyReportStatusCompleted,
			"error_message": "", "generated_at": &generatedAt, "update_time": &generatedAt,
		}).Error; err != nil {
			return err
		}
		table := "el_candidate"
		if result.ParticipantType == service.CompetencyParticipantTester {
			table = "el_tester"
		}
		return tx.Table(table).Where("id = ? AND paper_id = ?", result.ParticipantID, body.PaperID).
			Updates(map[string]any{"pdf_path": saved, "pdf_flag": 1, "update_time": &generatedAt}).Error
	}); err != nil {
		_ = os.Remove(saved)
		h.markReportFailed(report.ID, err)
		response.RestErr(c, "保存报告元数据失败")
		return
	}
	if existing.PDFPath != "" && existing.PDFPath != saved {
		h.removeReportFile(existing.PDFPath)
	}
	action := "generate"
	if body.Force {
		action = "regenerate"
	}
	if err := h.writeReportAudit(c, report.ID, body.PaperID, action, &login.UserID, 1, ""); err != nil {
		response.RestErr(c, "记录报告审计失败")
		return
	}
	if err := h.db.Where("id = ?", report.ID).Take(&report).Error; err != nil {
		response.RestErr(c, "读取报告实例失败")
		return
	}
	response.Rest(c, report)
}

func (h *CompetencyReportHandler) renderCompetencyReport(paperID string) ([]byte, error) {
	base := strings.TrimRight(h.examH.cfg.PdfGen.ReportBaseURL, "/")
	if base == "" {
		base = "http://127.0.0.1"
	}
	reportURL := fmt.Sprintf("%s/#/exam/competency/report/%s?_internal=%s", base, url.PathEscape(paperID), url.QueryEscape(h.examH.cfg.PdfGen.InternalToken))
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(h.examH.cfg.PdfGen.PageTimeoutMs)*time.Millisecond)
		data, incomplete, err := h.examH.pdfPool.GeneratePDF(ctx, reportURL, "window.__reportReady === true", "competency", "胜任力测评报告")
		cancel()
		if err == nil && !incomplete && len(data) >= 1024 {
			return data, nil
		}
		if err == nil {
			err = errors.New("报告页面数据未完整加载")
		}
		lastErr = err
	}
	return nil, lastErr
}

func (h *CompetencyReportHandler) saveCompetencyReportPDF(name, paperID string, data []byte) (string, string, error) {
	base := h.examH.cfg.Upload.Path
	if base == "" {
		base = "./tmp"
	}
	dir := filepath.Join(base, "competency", time.Now().Format("20060102"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	safeName := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(name)
	fileName := fmt.Sprintf("%s_胜任力测试报告_%s_%s.pdf", safeName, paperID, time.Now().Format("20060102150405000"))
	path := filepath.Join(dir, fileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(data)
	return path, hex.EncodeToString(sum[:]), nil
}

func (h *CompetencyReportHandler) Download(c *gin.Context) {
	login, ok := h.reportIdentity(c)
	if !ok {
		return
	}
	paperID := strings.TrimSpace(c.Query("paperId"))
	if paperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	var report model.CompetencyReport
	if err := h.db.Where("paper_id = ? AND content_version = ? AND status = ?", paperID, service.CompetencyTemporaryContentVersion, competencyReportStatusCompleted).
		Take(&report).Error; err != nil {
		response.RestErr(c, "报告尚未生成")
		return
	}
	path, err := h.validReportPath(report.PDFPath)
	if err != nil {
		response.RestErr(c, "报告文件路径无效")
		return
	}
	file, err := os.Open(path)
	if err != nil {
		response.RestErr(c, "报告文件不存在")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		response.RestErr(c, "读取报告文件失败")
		return
	}
	fileName := "胜任力测试报告-" + paperID + ".pdf"
	if err := h.writeReportAudit(c, report.ID, paperID, "download", &login.UserID, 1, ""); err != nil {
		response.RestErr(c, "记录报告下载审计失败")
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=competency-report.pdf; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if _, err := io.Copy(c.Writer, file); err != nil {
		return
	}
}

func (h *CompetencyReportHandler) validReportPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" || !strings.EqualFold(filepath.Ext(path), ".pdf") {
		return "", errors.New("invalid report path")
	}
	base := h.examH.cfg.Upload.Path
	if base == "" {
		base = "./tmp"
	}
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	pathAbs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(baseAbs, pathAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("report path escapes upload root")
	}
	return pathAbs, nil
}

func (h *CompetencyReportHandler) removeReportFile(path string) {
	if valid, err := h.validReportPath(path); err == nil {
		_ = os.Remove(valid)
	}
}

func (h *CompetencyReportHandler) markReportFailed(reportID string, reportErr error) {
	now := time.Now()
	message := reportErr.Error()
	if len(message) > 500 {
		message = message[:500]
	}
	_ = h.db.Model(&model.CompetencyReport{}).Where("id = ?", reportID).
		Updates(map[string]any{"status": competencyReportStatusFailed, "error_message": message, "update_time": &now}).Error
}

func (h *CompetencyReportHandler) writeReportAudit(c *gin.Context, reportID, paperID, action string, operatorID *int64, status int8, errorMessage string) error {
	now := time.Now()
	if len(errorMessage) > 500 {
		errorMessage = errorMessage[:500]
	}
	return h.db.Create(&model.CompetencyReportAudit{
		ID: uuid.NewString(), ReportID: reportID, PaperID: paperID, Action: action,
		OperatorID: operatorID, Status: status, ErrorMessage: errorMessage, ClientIP: c.ClientIP(), CreateTime: &now,
	}).Error
}
