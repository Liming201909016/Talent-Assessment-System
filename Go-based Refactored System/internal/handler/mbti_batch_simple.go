package handler

import (
	"archive/zip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/response"
)

// POST /exam/api/mbti/batch-download-simple
// Body: ["paperId1", "paperId2", ...]
// 批量下载 MBTI 简版报告（zip）。
//
// 简版查找规则与 DownloadReport(type='simple') 一致：
// 在 PDF 同目录搜索含 _simple 的最新文件（前缀按 姓名_TYPE_ 匹配）。
//
// 跳过未生成简版的考生（不阻塞整批），日志记录跳过数量。
func (h *MbtiReportHandler) BatchDownloadSimple(c *gin.Context) {
	var paperIDs []string
	if err := c.ShouldBindJSON(&paperIDs); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if len(paperIDs) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}

	// 查 pdfPath：先 candidate 后 tester（与单条下载一致）
	type row struct {
		PdfPath string `gorm:"column:pdf_path"`
		Name    string `gorm:"column:name"`
	}
	var rows []row
	h.db.Table("el_candidate").
		Select("pdf_path, name").
		Where("paper_id IN ? AND pdf_path IS NOT NULL AND pdf_path != ''", paperIDs).
		Scan(&rows)
	if len(rows) == 0 {
		// 兜底 tester
		h.db.Table("el_tester").
			Select("pdf_path, name").
			Where("paper_id IN ? AND pdf_path IS NOT NULL AND pdf_path != ''", paperIDs).
			Scan(&rows)
	}
	if len(rows) == 0 {
		response.RestErr(c, "没有可下载的报告")
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=mbti-simple-batch.zip")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	added, skipped := 0, 0
	used := make(map[string]int) // 同名计数
	for _, r := range rows {
		simplePath := findSimpleReport(r.PdfPath)
		if simplePath == "" {
			skipped++
			continue
		}
		f, err := os.Open(simplePath)
		if err != nil {
			skipped++
			continue
		}
		// zip entry 名：姓名_简版报告.pdf；同名时追加 (n)
		ext := filepath.Ext(simplePath)
		base := r.Name + "_简版报告"
		entryName := base + ext
		if used[entryName] > 0 {
			entryName = fmt.Sprintf("%s(%d)%s", base, used[entryName]+1, ext)
		}
		used[base+ext]++
		w, err := zw.Create(entryName)
		if err != nil {
			f.Close()
			skipped++
			continue
		}
		if _, err := io.Copy(w, f); err != nil {
			slog.Warn("batch-download-simple: copy failed", "name", r.Name, "error", err)
		}
		f.Close()
		added++
	}
	slog.Info("batch-download-simple", "added", added, "skipped", skipped, "total", len(rows))
}

// findSimpleReport 在给定 PDF 路径所在目录中找含 _simple 的同名前缀文件。
// 规则与 mbti_report.DownloadReport(type='simple') 一致。
func findSimpleReport(fullPdfPath string) string {
	dir := filepath.Dir(fullPdfPath)
	baseName := filepath.Base(fullPdfPath)
	parts := strings.SplitN(baseName, "_", 3)
	prefix := ""
	if len(parts) >= 2 {
		prefix = parts[0] + "_" + parts[1] + "_"
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var simplePath string
	for _, e := range entries {
		name := e.Name()
		if strings.Contains(name, "_simple") && (prefix == "" || strings.HasPrefix(name, prefix)) {
			simplePath = filepath.Join(dir, name)
		}
	}
	return simplePath
}
