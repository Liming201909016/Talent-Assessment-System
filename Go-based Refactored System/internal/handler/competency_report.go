package handler

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	competencyReportStatusGenerating = "generating"
	competencyReportStatusCompleted  = "completed"
	competencyReportStatusFailed     = "failed"
	competencyReportLockStripes      = 64
)

type CompetencyReportHandler struct {
	db           *gorm.DB
	examH        *ExamHandler
	wordRenderer *phase1WordReportRenderer
	templateMu   sync.RWMutex
	reportLocks  [competencyReportLockStripes]sync.Mutex
}

func NewCompetencyReportHandler(db *gorm.DB, examH *ExamHandler) *CompetencyReportHandler {
	handler := &CompetencyReportHandler{db: db, examH: examH}
	if examH != nil {
		handler.wordRenderer = newPhase1WordReportRenderer(examH.cfg)
	}
	return handler
}

func reportGenerationLockIndex(paperID string) int {
	const offset32 uint32 = 2166136261
	const prime32 uint32 = 16777619
	hash := offset32
	for index := 0; index < len(paperID); index++ {
		hash ^= uint32(paperID[index])
		hash *= prime32
	}
	return int(hash % competencyReportLockStripes)
}

func (h *CompetencyReportHandler) lockReportGeneration(paperID string) func() {
	lock := &h.reportLocks[reportGenerationLockIndex(paperID)]
	lock.Lock()
	return lock.Unlock
}

func competencyReportVersions(data map[string]any) (string, string, error) {
	contentVersion, contentOK := data["contentVersion"].(string)
	templateVersion, templateOK := data["reportTemplateVersion"].(string)
	contentVersion = strings.TrimSpace(contentVersion)
	templateVersion = strings.TrimSpace(templateVersion)
	if !contentOK || !templateOK || contentVersion == "" || templateVersion == "" {
		return "", "", errors.New("报告内容或模板版本未冻结")
	}
	return contentVersion, templateVersion, nil
}

func phase1V2ReportSnapshot(data *service.Phase1V2FormalReportData) map[string]any {
	if data == nil {
		return nil
	}
	payload, _ := json.Marshal(data.Report)
	output := make(map[string]any)
	_ = json.Unmarshal(payload, &output)
	output["reportKind"] = "frontline_phase1_v2"
	output["productVersion"] = data.Report.Versions.ProductVersion
	output["scoringVersion"] = data.Report.Versions.ScoringVersion
	output["contentVersion"] = data.Report.Versions.ContentVersion
	output["reportTemplateVersion"] = data.Report.Versions.ReportTemplateVersion
	output["reportText"] = data.Report
	return output
}

func upsertCompetencyReportCurrent(tx *gorm.DB, report model.CompetencyReport, now time.Time) error {
	current := model.CompetencyReportCurrent{
		PaperID: report.PaperID, Audience: report.Audience, ReportID: report.ID,
		CreateTime: &now, UpdateTime: &now,
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "paper_id"}, {Name: "audience"}},
		DoUpdates: clause.Assignments(map[string]any{"report_id": report.ID, "update_time": &now}),
	}).Create(&current).Error
}

func competencyReportBindingSchemaReady(db *gorm.DB) bool {
	return db != nil && db.Migrator().HasTable(&model.CompetencyResultRun{}) &&
		db.Migrator().HasTable(&model.CompetencyReportCurrent{}) &&
		db.Migrator().HasColumn(&model.CompetencyReport{}, "ResultRunID")
}

func competencyReportQuery(db *gorm.DB, bindingSchemaReady bool) *gorm.DB {
	if bindingSchemaReady {
		return db
	}
	return db.Omit("result_run_id")
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
	if h.wordRenderer == nil && h.examH.pdfPool == nil {
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
	unlock := h.lockReportGeneration(body.PaperID)
	defer unlock()

	runtime := service.NewCompetencyRuntimeService(h.db, h.examH.cfg)
	bindingSchemaReady := competencyReportBindingSchemaReady(h.db)
	var v2Data *service.Phase1V2FormalReportData
	hasV2Run := false
	var err error
	if bindingSchemaReady {
		v2Data, hasV2Run, err = runtime.FindPhase1V2FormalReportData(body.PaperID)
		if err != nil {
			response.RestErr(c, competencyRuntimeErrorMessage(errors.Join(service.ErrPhase1V2ResultRunUnavailable, err)))
			return
		}
	}
	var data map[string]any
	var result model.CompetencyResult
	if hasV2Run {
		data = phase1V2ReportSnapshot(v2Data)
		result = model.CompetencyResult{
			PaperID: body.PaperID, ExamID: v2Data.ExamID, ParticipantType: v2Data.ParticipantType,
			ParticipantID: v2Data.ParticipantID, ParticipantName: v2Data.ParticipantName,
			ReportAudience: v2Data.Report.Audience,
		}
	} else {
		data, err = runtime.FormalReportData(body.PaperID)
		if err != nil {
			response.RestErr(c, err.Error())
			return
		}
		if err := h.db.Where("paper_id = ?", body.PaperID).Take(&result).Error; err != nil {
			response.RestErr(c, "读取报告结果失败")
			return
		}
	}
	contentVersion, templateVersion, err := competencyReportVersions(data)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}

	var existing model.CompetencyReport
	existingQuery := competencyReportQuery(h.db, bindingSchemaReady).Where("paper_id = ? AND content_version = ? AND template_version = ?", body.PaperID, contentVersion, templateVersion)
	if hasV2Run {
		existingQuery = existingQuery.Where("result_run_id = ?", v2Data.Report.ResultRunID)
	}
	err = existingQuery.Take(&existing).Error
	if err == nil && !body.Force && existing.Status == competencyReportStatusCompleted && existing.PDFPath != "" {
		if info, statErr := os.Stat(existing.PDFPath); statErr == nil && info.Size() > 0 {
			if err := h.db.Transaction(func(tx *gorm.DB) error {
				if hasV2Run {
					if err := upsertCompetencyReportCurrent(tx, existing, time.Now()); err != nil {
						return err
					}
				}
				return h.writeReportAuditWithDB(tx, c, existing.ID, body.PaperID, "reuse", &login.UserID, 1, "")
			}); err != nil {
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
		report.ContentVersion = contentVersion
		report.TemplateVersion = templateVersion
		report.CreateTime = &now
	}
	if hasV2Run {
		runID := v2Data.Report.ResultRunID
		report.ResultRunID = &runID
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
	preserveCompletedV2 := hasV2Run && existing.ID != "" && existing.Status == competencyReportStatusCompleted && existing.PDFPath != ""
	if existing.ID == "" {
		createQuery := h.db
		if !bindingSchemaReady {
			createQuery = createQuery.Omit("ResultRunID")
		}
		if err := createQuery.Create(&report).Error; err != nil {
			response.RestErr(c, "创建报告实例失败")
			return
		}
	} else if !preserveCompletedV2 {
		updates := map[string]any{
			"audience": report.Audience, "text_snapshot": report.TextSnapshot, "score_snapshot": report.ScoreSnapshot,
			"status": report.Status, "error_message": "", "generated_by": report.GeneratedBy, "update_time": &now,
		}
		if bindingSchemaReady {
			updates["result_run_id"] = report.ResultRunID
		}
		if err := h.db.Model(&model.CompetencyReport{}).Where("id = ?", report.ID).Updates(updates).Error; err != nil {
			response.RestErr(c, "更新报告实例失败")
			return
		}
	}

	pdfBytes, err := h.renderCompetencyReport(c.Request.Context(), body.PaperID, data, v2Data)
	if err != nil {
		if !preserveCompletedV2 {
			h.markReportFailed(report.ID, err)
		}
		_ = h.writeReportAudit(c, report.ID, body.PaperID, "generate", &login.UserID, 0, err.Error())
		response.RestErr(c, "生成胜任力报告失败: "+err.Error())
		return
	}
	saved, digest, err := h.saveCompetencyReportPDF(result.ParticipantName, body.PaperID, pdfBytes)
	if err != nil {
		if !preserveCompletedV2 {
			h.markReportFailed(report.ID, err)
		}
		_ = h.writeReportAudit(c, report.ID, body.PaperID, "generate", &login.UserID, 0, err.Error())
		response.RestErr(c, "保存胜任力报告失败")
		return
	}
	generatedAt := time.Now()
	action := "generate"
	if body.Force {
		action = "regenerate"
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"pdf_path": saved, "pdf_sha256": digest, "pdf_size": len(pdfBytes), "status": competencyReportStatusCompleted,
			"audience": report.Audience, "text_snapshot": report.TextSnapshot,
			"score_snapshot": report.ScoreSnapshot, "error_message": "", "generated_by": report.GeneratedBy,
			"generated_at": &generatedAt, "update_time": &generatedAt,
		}
		if bindingSchemaReady {
			updates["result_run_id"] = report.ResultRunID
		}
		if err := tx.Model(&model.CompetencyReport{}).Where("id = ?", report.ID).Updates(updates).Error; err != nil {
			return err
		}
		if hasV2Run {
			if err := upsertCompetencyReportCurrent(tx, report, generatedAt); err != nil {
				return err
			}
		} else {
			table := "el_candidate"
			if result.ParticipantType == service.CompetencyParticipantTester {
				table = "el_tester"
			}
			if err := tx.Table(table).Where("id = ? AND paper_id = ?", result.ParticipantID, body.PaperID).
				Updates(map[string]any{"pdf_path": saved, "pdf_flag": 1, "update_time": &generatedAt}).Error; err != nil {
				return err
			}
		}
		return h.writeReportAuditWithDB(tx, c, report.ID, body.PaperID, action, &login.UserID, 1, "")
	}); err != nil {
		_ = os.Remove(saved)
		if !preserveCompletedV2 {
			h.markReportFailed(report.ID, err)
		}
		response.RestErr(c, "保存报告元数据失败")
		return
	}
	if existing.PDFPath != "" && existing.PDFPath != saved {
		h.removeReportFile(existing.PDFPath)
	}
	if err := competencyReportQuery(h.db, bindingSchemaReady).Where("id = ?", report.ID).Take(&report).Error; err != nil {
		response.RestErr(c, "读取报告实例失败")
		return
	}
	response.Rest(c, report)
}

func (h *CompetencyReportHandler) renderCompetencyReport(ctx context.Context, paperID string, data map[string]any, v2Data *service.Phase1V2FormalReportData) ([]byte, error) {
	var wordErr error
	if v2Data != nil {
		if h.wordRenderer == nil {
			return nil, errors.New("v2 Word报告渲染器未启用")
		}
		h.templateMu.RLock()
		pdf, err := h.wordRenderer.RenderPhase1V2(ctx, *v2Data)
		h.templateMu.RUnlock()
		if err == nil {
			return pdf, nil
		}
		return nil, err
	} else if reportKind, _ := data["reportKind"].(string); reportKind == "frontline_phase1" && h.wordRenderer != nil {
		h.templateMu.RLock()
		pdf, err := h.wordRenderer.Render(ctx, paperID, data)
		h.templateMu.RUnlock()
		if err == nil {
			return pdf, nil
		} else {
			wordErr = err
			if !h.wordRenderer.fallbackChromium {
				return nil, wordErr
			}
		}
	}
	if h.examH == nil || h.examH.pdfPool == nil {
		if wordErr != nil {
			return nil, wordErr
		}
		return nil, errors.New("Chromium报告渲染器未启用")
	}
	base := strings.TrimRight(h.examH.cfg.PdfGen.ReportBaseURL, "/")
	if base == "" {
		base = "http://127.0.0.1"
	}
	reportURL := fmt.Sprintf("%s/#/exam/competency/report/%s?_internal=%s", base, url.PathEscape(paperID), url.QueryEscape(h.examH.cfg.PdfGen.InternalToken))
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		renderCtx, cancel := context.WithTimeout(ctx, time.Duration(h.examH.cfg.PdfGen.PageTimeoutMs)*time.Millisecond)
		data, incomplete, err := h.examH.pdfPool.GeneratePDF(renderCtx, reportURL, "window.__reportReady === true", "competency", "胜任力测评报告")
		cancel()
		if err == nil && !incomplete && len(data) >= 1024 {
			return data, nil
		}
		if err == nil {
			err = errors.New("报告页面数据未完整加载")
		}
		lastErr = err
	}
	if wordErr != nil && lastErr != nil {
		return nil, fmt.Errorf("Word模板渲染失败：%v；Chromium回退失败：%v", wordErr, lastErr)
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
	unlock := h.lockReportGeneration(paperID)
	defer unlock()
	var report model.CompetencyReport
	var current model.CompetencyReportCurrent
	bindingSchemaReady := competencyReportBindingSchemaReady(h.db)
	currentErr := gorm.ErrRecordNotFound
	if bindingSchemaReady {
		currentErr = h.db.Where("paper_id = ? AND audience = ?", paperID, service.CompetencyReportAudienceFrontlineEmployee).Take(&current).Error
	}
	if currentErr == nil {
		if err := h.db.Where("id = ? AND paper_id = ? AND audience = ? AND status = ?", current.ReportID, paperID, current.Audience, competencyReportStatusCompleted).Take(&report).Error; err != nil {
			response.RestErr(c, "当前报告尚未生成")
			return
		}
		if report.ResultRunID == nil || strings.TrimSpace(*report.ResultRunID) == "" {
			response.RestErr(c, "当前报告未绑定评分运行")
			return
		}
		var run model.CompetencyResultRun
		if err := h.db.Where("id = ? AND paper_id = ?", *report.ResultRunID, paperID).Take(&run).Error; err != nil {
			response.RestErr(c, "当前报告评分运行不存在")
			return
		}
		versions := service.CompetencyVersionSet{ProductVersion: run.ProductVersion, ScoringVersion: run.ScoringVersion, ContentVersion: run.ContentVersion, ReportTemplateVersion: run.ReportTemplateVersion}
		if run.Status != "completed" || !service.IsPhase1V2VersionSet(versions) || report.ContentVersion != versions.ContentVersion || report.TemplateVersion != versions.ReportTemplateVersion || report.Audience != run.ReportAudience {
			response.RestErr(c, "当前报告版本绑定无效")
			return
		}
		var contentPackage model.CompetencyReportContentPackage
		if err := h.db.Where("product_version = ? AND scoring_version = ? AND content_version = ? AND template_version = ? AND audience = ?", versions.ProductVersion, versions.ScoringVersion, versions.ContentVersion, versions.ReportTemplateVersion, run.ReportAudience).Take(&contentPackage).Error; err != nil {
			response.RestErr(c, service.ErrPhase1ReportContentNotApproved.Error())
			return
		}
		if err := service.ValidatePhase1ReportContentApprovalForEnvironment(contentPackage, service.CompetencyReportEffectiveEnvironment()); err != nil {
			response.RestErr(c, err.Error())
			return
		}
	} else if !errors.Is(currentErr, gorm.ErrRecordNotFound) {
		response.RestErr(c, "查询当前报告失败")
		return
	} else {
		var result model.CompetencyResult
		if err := h.db.Select("product_version, scoring_version, content_version, report_template_version").Where("paper_id = ?", paperID).Take(&result).Error; err != nil {
			response.RestErr(c, "报告结果不存在")
			return
		}
		versions := service.CompetencyVersionSetFromResult(result)
		if service.IsPhase1CompetencyVersionSet(versions) {
			var contentPackage model.CompetencyReportContentPackage
			if err := h.db.Where("product_version = ? AND scoring_version = ? AND content_version = ? AND template_version = ? AND audience = ?", versions.ProductVersion, versions.ScoringVersion, versions.ContentVersion, versions.ReportTemplateVersion, service.CompetencyReportAudienceFrontlineEmployee).Take(&contentPackage).Error; err != nil {
				response.RestErr(c, service.ErrPhase1ReportContentNotApproved.Error())
				return
			}
			if err := service.ValidatePhase1ReportContentApproval(contentPackage); err != nil {
				response.RestErr(c, err.Error())
				return
			}
		} else if err := service.ValidateFrozenCompetencyVersionSet(versions); err != nil {
			response.RestErr(c, err.Error())
			return
		}
		if err := competencyReportQuery(h.db, bindingSchemaReady).Where("paper_id = ? AND content_version = ? AND template_version = ? AND status = ?", paperID, versions.ContentVersion, versions.ReportTemplateVersion, competencyReportStatusCompleted).
			Take(&report).Error; err != nil {
			response.RestErr(c, "报告尚未生成")
			return
		}
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
	c.Header("Content-Disposition", "attachment; filename=competency-report.pdf; filename*=UTF-8''"+encodeRFC5987FileName(fileName))
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if _, err := io.Copy(c.Writer, file); err != nil {
		return
	}
}

type competencyReportArchiveEntry struct {
	PaperID         string
	ReportID        string
	ParticipantName string
	Path            string
}

func sanitizeCompetencyArchiveNamePart(value, fallback string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		clean = fallback
	}
	return strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_").Replace(clean)
}

func competencyReportArchiveFileName(entry competencyReportArchiveEntry) string {
	safeName := sanitizeCompetencyArchiveNamePart(entry.ParticipantName, "受测者")
	safePaperID := sanitizeCompetencyArchiveNamePart(entry.PaperID, "unknown-paper")
	return fmt.Sprintf("%s-%s-胜任力测试报告.pdf", safeName, safePaperID)
}

func writeCompetencyReportArchive(writer io.Writer, entries []competencyReportArchiveEntry) error {
	archive := zip.NewWriter(writer)
	for _, entry := range entries {
		file, err := os.Open(entry.Path)
		if err != nil {
			_ = archive.Close()
			return err
		}
		header := &zip.FileHeader{Name: competencyReportArchiveFileName(entry), Method: zip.Deflate}
		part, err := archive.CreateHeader(header)
		if err == nil {
			_, err = io.Copy(part, file)
		}
		closeErr := file.Close()
		if err != nil {
			_ = archive.Close()
			return err
		}
		if closeErr != nil {
			_ = archive.Close()
			return closeErr
		}
	}
	return archive.Close()
}

func normalizeCompetencyReportPaperIDs(values []string) ([]string, error) {
	paperIDs := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		paperID := strings.TrimSpace(value)
		if paperID == "" {
			return nil, errors.New("paperId 为空")
		}
		if _, exists := seen[paperID]; exists {
			continue
		}
		seen[paperID] = struct{}{}
		paperIDs = append(paperIDs, paperID)
	}
	if len(paperIDs) == 0 {
		return nil, errors.New("paperIds 为空")
	}
	if len(paperIDs) > 100 {
		return nil, errors.New("一次最多下载100份报告")
	}
	return paperIDs, nil
}

func (h *CompetencyReportHandler) loadCompetencyReportArchiveEntries(paperIDs []string) ([]competencyReportArchiveEntry, error) {
	var results []model.CompetencyResult
	if err := h.db.Select("paper_id, participant_name, product_version, scoring_version, content_version, report_template_version").
		Where("paper_id IN ?", paperIDs).Find(&results).Error; err != nil {
		return nil, errors.New("读取报告结果失败")
	}
	resultByPaper := make(map[string]model.CompetencyResult, len(results))
	for _, result := range results {
		resultByPaper[result.PaperID] = result
	}

	bindingSchemaReady := competencyReportBindingSchemaReady(h.db)
	var currents []model.CompetencyReportCurrent
	if bindingSchemaReady {
		if err := h.db.Where("paper_id IN ? AND audience = ?", paperIDs, service.CompetencyReportAudienceFrontlineEmployee).Find(&currents).Error; err != nil {
			return nil, errors.New("查询当前报告失败")
		}
	}
	currentByPaper := make(map[string]model.CompetencyReportCurrent, len(currents))
	for _, current := range currents {
		currentByPaper[current.PaperID] = current
	}

	var reports []model.CompetencyReport
	if err := competencyReportQuery(h.db, bindingSchemaReady).Where("paper_id IN ? AND status = ?", paperIDs, competencyReportStatusCompleted).Find(&reports).Error; err != nil {
		return nil, errors.New("查询报告实例失败")
	}
	reportByKey := make(map[string]model.CompetencyReport, len(reports))
	reportByID := make(map[string]model.CompetencyReport, len(reports))
	runIDs := make([]string, 0, len(reports))
	for _, report := range reports {
		key := report.PaperID + "\x00" + report.ContentVersion + "\x00" + report.TemplateVersion
		reportByKey[key] = report
		reportByID[report.ID] = report
		if report.ResultRunID != nil && strings.TrimSpace(*report.ResultRunID) != "" {
			runIDs = append(runIDs, strings.TrimSpace(*report.ResultRunID))
		}
	}

	var runs []model.CompetencyResultRun
	if len(runIDs) > 0 {
		if err := h.db.Where("id IN ?", runIDs).Find(&runs).Error; err != nil {
			return nil, errors.New("查询报告评分运行失败")
		}
	}
	runByID := make(map[string]model.CompetencyResultRun, len(runs))
	for _, run := range runs {
		runByID[run.ID] = run
	}

	approvalCache := make(map[string]error)
	entries := make([]competencyReportArchiveEntry, 0, len(paperIDs))
	for _, paperID := range paperIDs {
		result, exists := resultByPaper[paperID]
		if !exists {
			return nil, errors.New("报告结果不存在: " + paperID)
		}
		participantName := result.ParticipantName
		var report model.CompetencyReport
		if current, hasCurrent := currentByPaper[paperID]; hasCurrent {
			var reportExists bool
			report, reportExists = reportByID[current.ReportID]
			if !reportExists || report.PaperID != paperID || report.Audience != current.Audience || report.Status != competencyReportStatusCompleted {
				return nil, errors.New("当前报告尚未生成: " + paperID)
			}
			if report.ResultRunID == nil || strings.TrimSpace(*report.ResultRunID) == "" {
				return nil, errors.New("当前报告未绑定评分运行: " + paperID)
			}
			run, runExists := runByID[strings.TrimSpace(*report.ResultRunID)]
			versions := service.CompetencyVersionSet{ProductVersion: run.ProductVersion, ScoringVersion: run.ScoringVersion, ContentVersion: run.ContentVersion, ReportTemplateVersion: run.ReportTemplateVersion}
			if !runExists || run.PaperID != paperID || run.Status != "completed" || !service.IsPhase1V2VersionSet(versions) || report.ContentVersion != versions.ContentVersion || report.TemplateVersion != versions.ReportTemplateVersion || report.Audience != run.ReportAudience {
				return nil, errors.New("当前报告版本绑定无效: " + paperID)
			}
			approvalKey := versions.ProductVersion + "\x00" + versions.ScoringVersion + "\x00" + versions.ContentVersion + "\x00" + versions.ReportTemplateVersion + "\x00" + run.ReportAudience
			approvalErr, checked := approvalCache[approvalKey]
			if !checked {
				var contentPackage model.CompetencyReportContentPackage
				approvalErr = h.db.Where("product_version = ? AND scoring_version = ? AND content_version = ? AND template_version = ? AND audience = ?", versions.ProductVersion, versions.ScoringVersion, versions.ContentVersion, versions.ReportTemplateVersion, run.ReportAudience).Take(&contentPackage).Error
				if approvalErr == nil {
					approvalErr = service.ValidatePhase1ReportContentApprovalForEnvironment(contentPackage, service.CompetencyReportEffectiveEnvironment())
				} else {
					approvalErr = service.ErrPhase1ReportContentNotApproved
				}
				approvalCache[approvalKey] = approvalErr
			}
			if approvalErr != nil {
				return nil, approvalErr
			}
			participantName = run.ParticipantName
		} else {
			versions := service.CompetencyVersionSetFromResult(result)
			if service.IsPhase1CompetencyVersionSet(versions) {
				approvalKey := versions.ProductVersion + "\x00" + versions.ScoringVersion + "\x00" + versions.ContentVersion + "\x00" + versions.ReportTemplateVersion
				approvalErr, checked := approvalCache[approvalKey]
				if !checked {
					var contentPackage model.CompetencyReportContentPackage
					approvalErr = h.db.Where("product_version = ? AND scoring_version = ? AND content_version = ? AND template_version = ? AND audience = ?", versions.ProductVersion, versions.ScoringVersion, versions.ContentVersion, versions.ReportTemplateVersion, service.CompetencyReportAudienceFrontlineEmployee).Take(&contentPackage).Error
					if approvalErr == nil {
						approvalErr = service.ValidatePhase1ReportContentApproval(contentPackage)
					} else {
						approvalErr = service.ErrPhase1ReportContentNotApproved
					}
					approvalCache[approvalKey] = approvalErr
				}
				if approvalErr != nil {
					return nil, approvalErr
				}
			} else if err := service.ValidateFrozenCompetencyVersionSet(versions); err != nil {
				return nil, err
			}

			key := paperID + "\x00" + versions.ContentVersion + "\x00" + versions.ReportTemplateVersion
			report, exists = reportByKey[key]
			if !exists {
				return nil, errors.New("报告尚未生成: " + paperID)
			}
		}
		path, err := h.validReportPath(report.PDFPath)
		if err != nil {
			return nil, errors.New("报告文件路径无效: " + paperID)
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return nil, errors.New("报告文件不存在: " + paperID)
		}
		entries = append(entries, competencyReportArchiveEntry{PaperID: paperID, ReportID: report.ID, ParticipantName: participantName, Path: path})
	}
	return entries, nil
}

func createCompetencyReportArchiveFile(entries []competencyReportArchiveEntry) (*os.File, os.FileInfo, error) {
	archiveFile, err := os.CreateTemp("", "competency-reports-*.zip")
	if err != nil {
		return nil, nil, err
	}
	archivePath := archiveFile.Name()
	if err := writeCompetencyReportArchive(archiveFile, entries); err != nil {
		archiveFile.Close()
		os.Remove(archivePath)
		return nil, nil, err
	}
	if err := archiveFile.Close(); err != nil {
		os.Remove(archivePath)
		return nil, nil, err
	}
	archiveFile, err = os.Open(archivePath)
	if err != nil {
		os.Remove(archivePath)
		return nil, nil, err
	}
	archiveInfo, err := archiveFile.Stat()
	if err != nil || archiveInfo.Size() == 0 {
		archiveFile.Close()
		os.Remove(archivePath)
		if err == nil {
			err = errors.New("empty report archive")
		}
		return nil, nil, err
	}
	return archiveFile, archiveInfo, nil
}

func (h *CompetencyReportHandler) writeBatchDownloadAudits(c *gin.Context, entries []competencyReportArchiveEntry, operatorID *int64) error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			if err := h.writeReportAuditWithDB(tx, c, entry.ReportID, entry.PaperID, "download", operatorID, 1, ""); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *CompetencyReportHandler) BatchDownload(c *gin.Context) {
	login, ok := h.reportIdentity(c)
	if !ok {
		return
	}
	var body struct {
		PaperIDs []string `json:"paperIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	paperIDs, err := normalizeCompetencyReportPaperIDs(body.PaperIDs)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	entries, err := h.loadCompetencyReportArchiveEntries(paperIDs)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	archiveFile, archiveInfo, err := createCompetencyReportArchiveFile(entries)
	if err != nil {
		response.RestErr(c, "创建报告压缩包失败")
		return
	}
	archivePath := archiveFile.Name()
	defer archiveFile.Close()
	defer os.Remove(archivePath)
	if err := h.writeBatchDownloadAudits(c, entries, &login.UserID); err != nil {
		response.RestErr(c, "记录报告下载审计失败")
		return
	}

	archiveName := "胜任力报告-" + time.Now().Format("20060102-150405") + ".zip"
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=competency-reports.zip; filename*=UTF-8''"+encodeRFC5987FileName(archiveName))
	c.Header("Content-Length", strconv.FormatInt(archiveInfo.Size(), 10))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, archiveFile)
}

func encodeRFC5987FileName(fileName string) string {
	return strings.ReplaceAll(url.QueryEscape(fileName), "+", "%20")
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
	return h.writeReportAuditWithDB(h.db, c, reportID, paperID, action, operatorID, status, errorMessage)
}

func (h *CompetencyReportHandler) writeReportAuditWithDB(db *gorm.DB, c *gin.Context, reportID, paperID, action string, operatorID *int64, status int8, errorMessage string) error {
	now := time.Now()
	if len(errorMessage) > 500 {
		errorMessage = errorMessage[:500]
	}
	return db.Create(&model.CompetencyReportAudit{
		ID: uuid.NewString(), ReportID: reportID, PaperID: paperID, Action: action,
		OperatorID: operatorID, Status: status, ErrorMessage: errorMessage, ClientIP: c.ClientIP(), CreateTime: &now,
	}).Error
}
