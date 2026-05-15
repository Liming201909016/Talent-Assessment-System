package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
)

// POST /exam/api/exam/exam/generate-report
// Body: {paperId}
// 同步：服务器用 chromedp 渲染前端报告页 → PDF → 落盘 → 更新 pdf_path/pdf_flag
//
// 仅 admin 或 exam:list/exam:export 可用。
func (h *ExamHandler) GenerateReport(c *gin.Context) {
	luVal, _ := c.Get("loginUser")
	lu, _ := luVal.(*model.LoginUser)
	if lu == nil {
		response.AjaxUnauthorized(c, "")
		return
	}
	if !h.canGenerateReport(lu) {
		response.Ajax(c, 403, "无权生成报告", nil)
		return
	}
	if h.pdfPool == nil {
		response.RestErr(c, "后端报告生成未启用（pdfgen.enabled=false）")
		return
	}

	var b struct {
		PaperID string `json:"paperId" form:"paperId"`
	}
	_ = c.ShouldBind(&b)
	if b.PaperID == "" {
		b.PaperID = c.Query("paperId")
	}
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}

	// 单份生成 + 失败 1 次重试（每次独立 ctx，避免共享 timeout 已用完）
	pdfPath, err := h.generateWithRetry(b.PaperID, 2)
	if err != nil {
		slog.Error("generate-report failed", "paperId", b.PaperID, "error", err)
		response.RestErr(c, "生成报告失败: "+err.Error())
		return
	}
	response.Rest(c, gin.H{"paperId": b.PaperID, "pdfPath": pdfPath, "ok": true})
}

// generateWithRetry 调用 generateOneReport，失败时重试 (maxAttempts-1) 次。
// 每次用独立的 ctx，避免共享 timeout 已被首次尝试耗尽。
// 重试间 100ms 短退避（chromedp tab 已新建，无需长 backoff）。
// 仅重试可重试错误（chromedp/timeout/网络类），业务错误（试卷不存在/无权/repoCode 不支持）不重试。
// R4: 熔断 - 若处于冷却窗口（连续 3 次失败后 60s 内）则只尝试 1 次，避免系统性故障放大延迟。
func (h *ExamHandler) generateWithRetry(paperID string, maxAttempts int) (string, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	// 检查熔断
	if h.circuitOpen() {
		slog.Warn("generate-report circuit open, skip retry", "paperId", paperID)
		maxAttempts = 1
	}
	var lastErr error
	var lastPath string
	var lastIncomplete bool
	for i := 1; i <= maxAttempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(),
			time.Duration(h.cfg.PdfGen.PageTimeoutMs)*time.Millisecond)
		path, incomplete, err := h.generateOneReport(ctx, paperID)
		cancel()
		if err == nil && !incomplete {
			if i > 1 {
				slog.Info("generate-report succeeded after retry", "paperId", paperID, "attempt", i)
			}
			h.recordSuccess()
			return path, nil
		}
		if err == nil {
			// 生成成功但前端数据不完整 → 保存作为 fallback，还有 attempt 则重试
			lastPath = path
			lastIncomplete = true
			slog.Warn("generate-report incomplete data", "paperId", paperID, "attempt", i, "max", maxAttempts, "path", path)
			if i < maxAttempts {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			// 最后一次仍 incomplete→ 接受（总比报错强）
			h.recordSuccess() // incomplete 也算"产出"，不计为失败
			return lastPath, nil
		}
		lastErr = err
		retryable := isRetryableReportErr(err)
		slog.Warn("generate-report attempt failed", "paperId", paperID, "attempt", i, "max", maxAttempts, "retryable", retryable, "error", err)
		if !retryable {
			// 业务错误：直接返回，不计入熔断（不是基础设施故障）
			return "", err
		}
		if i < maxAttempts {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if lastIncomplete && lastPath != "" {
		h.recordSuccess()
		return lastPath, nil
	}
	h.recordFailure()
	return "", lastErr
}

// circuitOpen 判断熔断窗口是否打开（已连续失败到阈值且未过冷却时间）。
func (h *ExamHandler) circuitOpen() bool {
	h.cbMu.Lock()
	defer h.cbMu.Unlock()
	return time.Now().Before(h.cbOpenUntil)
}

// recordSuccess 重置失败计数和熔断窗口。
func (h *ExamHandler) recordSuccess() {
	h.cbMu.Lock()
	defer h.cbMu.Unlock()
	if h.cbConsecutiveFail > 0 {
		slog.Info("[exam] circuit reset", "previousFails", h.cbConsecutiveFail)
	}
	h.cbConsecutiveFail = 0
	h.cbOpenUntil = time.Time{}
}

// recordFailure 累加失败次数，超阈值则打开熔断窗口 60s。
func (h *ExamHandler) recordFailure() {
	h.cbMu.Lock()
	defer h.cbMu.Unlock()
	h.cbConsecutiveFail++
	if h.cbConsecutiveFail >= 3 {
		h.cbOpenUntil = time.Now().Add(60 * time.Second)
		slog.Warn("[exam] circuit opened", "fails", h.cbConsecutiveFail, "cooldownSec", 60)
	}
}

// isRetryableReportErr 判断是否值得重试。
// 可重试：chromedp 渲染失败、context timeout、网络/IO 错误
// 不可重试：业务校验错误（试卷不存在、无权、repoCode 不支持、PDF 太小）
func isRetryableReportErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// 业务错误关键字（不可重试）
	for _, kw := range []string{"试卷不存在", "不支持的 repoCode", "MBTI", "PDF 太小", "paperId 为空"} {
		if strings.Contains(msg, kw) {
			return false
		}
	}
	// 默认可重试（chromedp render / context deadline / wait ready timeout / pdfcpu / 文件 IO 等）
	return true
}

func (h *ExamHandler) canGenerateReport(lu *model.LoginUser) bool {
	if lu.UserID == 1 {
		return true
	}
	for _, p := range lu.Permissions {
		if p == "*:*:*" || p == "exam:list" || p == "exam:export" {
			return true
		}
	}
	return false
}

// generateOneReport 用 chromedp 渲染一份报告 → 落盘 → 更新 DB → 返回新 pdfPath。
// incomplete=true 表示前端数据未完全加载（window.__reportIncomplete），调用方可选重试。
func (h *ExamHandler) generateOneReport(ctx context.Context, paperID string) (string, bool, error) {
	// 1. 查询 paper / exam / repo / tester(or candidate)
	type papInfo struct {
		PaperID    string `gorm:"column:paper_id"`
		ExamID     string `gorm:"column:exam_id"`
		Title      string `gorm:"column:title"`
		RepoCode   string `gorm:"column:repo_code"`
		StuFlag    int    `gorm:"column:stu_flag"`
		Name       string `gorm:"column:name"`
		PersonID   string `gorm:"column:person_id"`
		PersonType string `gorm:"column:person_type"` // tester / candidate
		OldPdfPath string `gorm:"column:old_pdf_path"`
	}
	var info papInfo

	// 优先 candidate
	row := h.db.Table("el_paper p").
		Joins("INNER JOIN el_exam e ON e.id = p.exam_id").
		Joins("LEFT JOIN el_exam_repo er ON er.exam_id = p.exam_id").
		Joins("LEFT JOIN el_repo r ON r.id = er.repo_id").
		Joins("INNER JOIN el_candidate c ON c.paper_id = p.id").
		Where("p.id = ?", paperID).
		Limit(1).
		Select(`p.id AS paper_id, p.exam_id, e.title, COALESCE(r.code,'') AS repo_code, COALESCE(e.stu_flag, 0) AS stu_flag,
			c.name, c.id AS person_id, 'candidate' AS person_type,
			COALESCE(c.pdf_path,'') AS old_pdf_path`).
		Take(&info)
	if row.Error != nil || info.Name == "" {
		// 兜底 tester
		_ = h.db.Table("el_paper p").
			Joins("INNER JOIN el_exam e ON e.id = p.exam_id").
			Joins("LEFT JOIN el_exam_repo er ON er.exam_id = p.exam_id").
			Joins("LEFT JOIN el_repo r ON r.id = er.repo_id").
			Joins("INNER JOIN el_tester t ON t.paper_id = p.id").
			Where("p.id = ?", paperID).
			Limit(1).
			Select(`p.id AS paper_id, p.exam_id, e.title, COALESCE(r.code,'') AS repo_code, COALESCE(e.stu_flag, 0) AS stu_flag,
				t.name, t.id AS person_id, 'tester' AS person_type,
				COALESCE(t.pdf_path,'') AS old_pdf_path`).
			Take(&info).Error
	}
	if info.PaperID == "" {
		return "", false, errors.New("试卷不存在或未关联考生")
	}

	// 2. 路由选择：001→result, 002→result2，003 不走此流程
	var routePath, headerTitle string
	switch {
	case strings.HasPrefix(info.RepoCode, "001"):
		routePath = "/my/exam/result"
		headerTitle = "职业心理测评报告"
	case strings.HasPrefix(info.RepoCode, "002"):
		routePath = "/my/exam/result2"
		headerTitle = "管理特质测验报告"
	case strings.HasPrefix(info.RepoCode, "003"):
		return "", false, errors.New("MBTI 报告由现有 docx→PDF 流程生成，不走 chromedp")
	default:
		return "", false, fmt.Errorf("不支持的 repoCode: %s", info.RepoCode)
	}

	// 3. URL：用 hash 路由，带 internalToken（防止外部用户故意构造 ?_internal 跳过 createPdf）
	base := h.cfg.PdfGen.ReportBaseURL
	if base == "" {
		base = "http://127.0.0.1"
	}
	tok := h.cfg.PdfGen.InternalToken
	// 路由本身只接受 :id/:testerId 两个 params，examId/stuFlag/repoCode 通过 query 传给前端
	// 修复 FB-035：之前缺这些 query 导致学生版/管理特质类报告模板渲染缺失评估段落
	reportURL := fmt.Sprintf("%s/#%s/%s/%s?_internal=%s&examId=%s&stuFlag=%d&repoCode=%s",
		base, routePath, paperID, info.PersonID, tok,
		url.QueryEscape(info.ExamID), info.StuFlag, url.QueryEscape(info.RepoCode))

	// 4. chromedp 渲染
	// FB-039: 传 reportType 让 chromedp HF 选对应 PNG 图标
	reportType := ""
	if strings.HasPrefix(info.RepoCode, "001") {
		reportType = "001"
	} else if strings.HasPrefix(info.RepoCode, "002") {
		reportType = "002"
	}
	pdfBytes, incomplete, err := h.pdfPool.GeneratePDF(ctx, reportURL, "window.__reportReady === true", reportType, headerTitle)
	if err != nil {
		return "", false, fmt.Errorf("chromedp render: %w", err)
	}
	if len(pdfBytes) < 1024 {
		return "", false, fmt.Errorf("生成的 PDF 太小 (%d 字节)，疑似失败", len(pdfBytes))
	}

	// 5. 落盘（与 PdfPersistence 一致路径规则）
	baseDir := h.cfg.Upload.Path
	if baseDir == "" {
		baseDir = "./tmp"
	}
	day := time.Now().Format("20060102")
	pdfDir := filepath.Join(baseDir, day)
	if err := os.MkdirAll(pdfDir, 0o755); err != nil {
		return "", false, err
	}
	ts := time.Now().Format("20060102150405000")
	// 文件名清理（去除 / \ 等）
	safeName := strings.ReplaceAll(strings.ReplaceAll(info.Name, "/", "_"), "\\", "_")
	safeTitle := strings.ReplaceAll(strings.ReplaceAll(info.Title, "/", "_"), "\\", "_")
	fname := fmt.Sprintf("%s_%s_%s.pdf", safeName, safeTitle, ts)
	saved := filepath.Join(pdfDir, fname)
	if err := os.WriteFile(saved, pdfBytes, 0o644); err != nil {
		return "", false, err
	}

	// 6. 删旧文件（在 allowedDir 内）
	if info.OldPdfPath != "" {
		oldClean := filepath.Clean(info.OldPdfPath)
		allowedBase := filepath.Clean(baseDir)
		if strings.HasPrefix(oldClean, allowedBase) {
			_ = os.Remove(oldClean)
		}
	}

	// 7. 更新 DB
	now := time.Now()
	pdfPartial := 0
	if incomplete {
		pdfPartial = 1
	}
	updates := map[string]any{
		"pdf_path":    saved,
		"pdf_flag":    1,
		"pdf_partial": pdfPartial,
		"update_time": &now,
	}
	tbl := "el_candidate"
	if info.PersonType == "tester" {
		tbl = "el_tester"
	}
	if err := h.db.Table(tbl).Where("id = ?", info.PersonID).Updates(updates).Error; err != nil {
		return "", false, err
	}

	// 8. 异步压缩 PDF（ghostscript），不阻塞响应
	// gs ebook 压缩约 3-4s，能把 1.4MB → 680KB（-50%）
	// 异步：generate-report 立即返回；用户立刻下载可能拿到大文件，但 4s 后就是压缩版
	go compressPDF(saved)

	slog.Info("[exam] report generated",
		"paperId", paperID, "name", info.Name, "size", len(pdfBytes), "path", saved, "incomplete", incomplete)
	return saved, incomplete, nil
}
