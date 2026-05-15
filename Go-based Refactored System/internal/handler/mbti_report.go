package handler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// MbtiReportHandler 生成 MBTI 报告 (docx) — 支持完整版和简版
type MbtiReportHandler struct {
	db          *gorm.DB
	templateDir string // 完整版 16 个 docx 模板所在目录
	simpleDir   string // 简版 16 个 docx 模板所在目录
	outputDir   string // 生成的报告存放目录
}

func NewMbtiReportHandler(db *gorm.DB, templateDir, simpleDir, outputDir string) *MbtiReportHandler {
	return &MbtiReportHandler{db: db, templateDir: templateDir, simpleDir: simpleDir, outputDir: outputDir}
}

// mbtiTesterRow 报告生成用考生信息
type mbtiTesterRow struct {
	Name        string  `gorm:"column:name"`
	Age         *int    `gorm:"column:age"`
	Gender      *string `gorm:"column:gender"`
	Telephone   *string `gorm:"column:telephone"`
	Affiliation *string `gorm:"column:affiliation"`
	Post        *string `gorm:"column:post"`
	ExamID      string  `gorm:"column:exam_id"`
}

// POST /exam/api/mbti/generate-report {paperId, type: "full"|"simple", force?: bool}
// 根据 MBTI 类型选择 docx 模板 → 替换首页字段 + 图表数据 → 存文件 → 返回下载路径
// type=simple 时若已有 simple 文件，默认直接返回（cached:true）；force=true 则删旧重建。
func (h *MbtiReportHandler) GenerateReport(c *gin.Context) {
	var b struct {
		PaperID    string `json:"paperId"`
		ReportType string `json:"type"`  // "full"(默认) 或 "simple"
		Force      bool   `json:"force"` // 强制重新生成，跳过缓存
	}
	_ = c.ShouldBindJSON(&b)
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}

	// 1. 查 MBTI 类型和分值
	mbtiH := &MbtiHandler{db: h.db}
	scores, mbtiType := mbtiH.calcMbtiScores(b.PaperID)
	if mbtiType == "" || len(mbtiType) != 4 {
		response.RestErr(c, "MBTI 类型未计算，请先完成答题")
		return
	}

	// 2. 查考生信息（先查 el_tester，再查 el_candidate）
	var tester mbtiTesterRow
	if err := h.db.Table("el_tester").Where("paper_id = ?", b.PaperID).Take(&tester).Error; err != nil {
		if err2 := h.db.Table("el_candidate").Where("paper_id = ?", b.PaperID).Take(&tester).Error; err2 != nil {
			response.RestErr(c, "考生信息不存在")
			return
		}
	}

	// 查 paper 用时
	var userTime *int
	h.db.Table("el_paper").Where("id = ?", b.PaperID).Pluck("user_time", &userTime)

	// 3. 查 exam 的 requiredFields
	var requiredFields string
	h.db.Table("el_exam").Where("id = ?", tester.ExamID).Pluck("required_fields", &requiredFields)

	// 4. 判断报告类型
	isSimple := b.ReportType == "simple"

	// 4.1 简版幂等：若 PDF 同目录已有 _simple 文件，直接返回（避免 LibreOffice 重复转换）
	// force=true 时跳过缓存：删除旧 simple 文件后继续重生成。
	if isSimple {
		var existingFull string
		h.db.Table("el_candidate").Where("paper_id = ?", b.PaperID).Pluck("pdf_path", &existingFull)
		if existingFull == "" {
			h.db.Table("el_tester").Where("paper_id = ?", b.PaperID).Pluck("pdf_path", &existingFull)
		}
		if existingFull != "" {
			if existingSimple := findSimpleReport(existingFull); existingSimple != "" {
				if _, statErr := os.Stat(existingSimple); statErr == nil {
					if b.Force {
						// 强制重生成：删除旧的同前缀 simple 文件
						dir := filepath.Dir(existingFull)
						base := filepath.Base(existingFull)
						parts := strings.SplitN(base, "_", 3)
						prefix := ""
						if len(parts) >= 2 {
							prefix = parts[0] + "_" + parts[1] + "_"
						}
						if prefix != "" {
							entries, _ := os.ReadDir(dir)
							for _, e := range entries {
								name := e.Name()
								if strings.HasPrefix(name, prefix) && strings.Contains(name, "_simple") {
									_ = os.Remove(filepath.Join(dir, name))
								}
							}
						}
					} else {
						response.Rest(c, gin.H{
							"path":       existingSimple,
							"type":       mbtiType,
							"fileName":   filepath.Base(existingSimple),
							"reportType": "simple",
							"cached":     true,
						})
						return
					}
				}
			}
		}
	}

	// 5. 选择模板文件
	var templateFile string
	if isSimple {
		templateFile = h.findTemplateInDir(h.simpleDir, mbtiType)
	} else {
		templateFile = h.findTemplate(mbtiType)
	}
	if templateFile == "" {
		label := "完整版"
		if isSimple {
			label = "简版"
		}
		response.RestErr(c, fmt.Sprintf("找不到 %s 类型的%s报告模板", mbtiType, label))
		return
	}

	// 6. 构造替换数据
	gender := ""
	if tester.Gender != nil {
		if *tester.Gender == "0" {
			gender = "男"
		} else if *tester.Gender == "1" {
			gender = "女"
		}
	}

	var fields map[string]string
	if isSimple {
		// 简版只有姓名/年龄/性别
		fields = map[string]string{
			"姓名：": tester.Name,
			"年龄：": intToStr(tester.Age),
			"性别：": gender,
		}
	} else {
		// 完整版全部字段
		timeStr := ""
		if userTime != nil && *userTime > 0 {
			timeStr = strconv.Itoa(*userTime) + "分钟"
		}
		fields = map[string]string{
			"姓名：":   tester.Name,
			"年龄：":   intToStr(tester.Age),
			"性别：":   gender,
			"单位：":   ptrStr(tester.Affiliation),
			"岗位：":   ptrStr(tester.Post),
			"联系方式：": ptrStr(tester.Telephone),
			"测评时长：": timeStr,
		}

		// 按 requiredFields 过滤（空 = 全显示）
		rfMap := map[string]string{
			"name": "姓名：", "age": "年龄：", "gender": "性别：",
			"affiliation": "单位：", "post": "岗位：", "telephone": "联系方式：",
		}
		if requiredFields != "" {
			allowed := map[string]bool{"姓名：": true}
			for _, f := range strings.Split(requiredFields, ",") {
				if label, ok := rfMap[f]; ok {
					allowed[label] = true
				}
			}
			for k := range fields {
				if !allowed[k] {
					fields[k] = ""
				}
			}
		}
	}

	// 报告日期
	now := time.Now()
	dateStr := fmt.Sprintf("%d年%d月%d日", now.Year(), now.Month(), now.Day())

	// 7. 执行 docx 模板替换
	outputBuf, err := h.processTemplate(templateFile, fields, dateStr, scores, isSimple)
	if err != nil {
		response.RestErr(c, "生成报告失败: "+err.Error())
		return
	}

	// 8. 保存 docx 文件
	// 简版与完整版放同目录，方便 findSimpleReport 按前缀查找
	day := now.Format("20060102")
	outDir := filepath.Join(h.outputDir, day)
	if isSimple {
		var existingFull string
		h.db.Table("el_candidate").Where("paper_id = ?", b.PaperID).Pluck("pdf_path", &existingFull)
		if existingFull == "" {
			h.db.Table("el_tester").Where("paper_id = ?", b.PaperID).Pluck("pdf_path", &existingFull)
		}
		if existingFull != "" {
			fullDir := filepath.Dir(existingFull)
			// 路径校验：必须在 outputDir 内
			absOut, _ := filepath.Abs(h.outputDir)
			absFull, _ := filepath.Abs(fullDir)
			if strings.HasPrefix(absFull, absOut) {
				outDir = fullDir
			}
		}
	}
	_ = os.MkdirAll(outDir, 0o755)
	suffix := ""
	if isSimple {
		suffix = "_simple"
	}
	baseName := fmt.Sprintf("%s_%s_%s%s", tester.Name, mbtiType, now.Format("20060102150405"), suffix)
	docxPath := filepath.Join(outDir, baseName+".docx")
	if err := os.WriteFile(docxPath, outputBuf.Bytes(), 0o644); err != nil {
		response.RestErr(c, "保存报告失败: "+err.Error())
		return
	}

	// 8. docx → PDF（通过 LibreOffice 转换，全局锁顺序化）
	finalPath := docxPath // 默认返回 docx
	if pdfRet, convErr := convertDocxToPdf(docxPath, outDir); convErr == nil {
		finalPath = pdfRet
		_ = os.Remove(docxPath) // 删除临时 docx
	}

	// 9. 更新 pdfPath + pdfFlag（简版不覆盖 pdf_path，完整版才更新）
	if !isSimple {
		updates := map[string]interface{}{"pdf_path": finalPath, "pdf_flag": 1, "update_time": &now}
		h.db.Table("el_tester").Where("paper_id = ?", b.PaperID).Updates(updates)
		h.db.Table("el_candidate").Where("paper_id = ?", b.PaperID).Updates(updates)
	}

	fileName := filepath.Base(finalPath)
	response.Rest(c, gin.H{"path": finalPath, "type": mbtiType, "fileName": fileName, "reportType": b.ReportType})
}

// GenerateReportByPaperID 异步生成报告（提交答卷后自动调用，无 gin.Context）
func (h *MbtiReportHandler) GenerateReportByPaperID(paperID string) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("report-async: panic", "paperId", paperID, "recover", r)
		}
	}()

	// 检查是否已生成
	var existingPath string
	h.db.Table("el_candidate").Where("paper_id = ? AND pdf_flag = 1 AND pdf_path IS NOT NULL AND pdf_path != ''", paperID).Pluck("pdf_path", &existingPath)
	if existingPath != "" {
		return // 已有报告，跳过
	}
	h.db.Table("el_tester").Where("paper_id = ? AND pdf_flag = 1 AND pdf_path IS NOT NULL AND pdf_path != ''", paperID).Pluck("pdf_path", &existingPath)
	if existingPath != "" {
		return
	}

	// 复用 GenerateReport 的核心逻辑
	mbtiH := &MbtiHandler{db: h.db}
	scores, mbtiType := mbtiH.calcMbtiScores(paperID)
	if mbtiType == "" || len(mbtiType) != 4 {
		slog.Warn("report-async: MBTI type empty", "paperId", paperID)
		return
	}

	var tester mbtiTesterRow
	if err := h.db.Table("el_tester").Where("paper_id = ?", paperID).Take(&tester).Error; err != nil {
		if err2 := h.db.Table("el_candidate").Where("paper_id = ?", paperID).Take(&tester).Error; err2 != nil {
			slog.Warn("report-async: tester not found", "paperId", paperID)
			return
		}
	}

	var userTime *int
	h.db.Table("el_paper").Where("id = ?", paperID).Pluck("user_time", &userTime)

	var requiredFields string
	h.db.Table("el_exam").Where("id = ?", tester.ExamID).Pluck("required_fields", &requiredFields)

	templateFile := h.findTemplate(mbtiType)
	if templateFile == "" {
		slog.Warn("report-async: template not found", "paperId", paperID, "type", mbtiType)
		return
	}

	gender := ""
	if tester.Gender != nil {
		if *tester.Gender == "0" {
			gender = "男"
		} else if *tester.Gender == "1" {
			gender = "女"
		}
	}
	timeStr := ""
	if userTime != nil && *userTime > 0 {
		timeStr = strconv.Itoa(*userTime) + "分钟"
	}
	fields := map[string]string{
		"姓名：": tester.Name, "年龄：": intToStr(tester.Age), "性别：": gender,
		"单位：": ptrStr(tester.Affiliation), "岗位：": ptrStr(tester.Post),
		"联系方式：": ptrStr(tester.Telephone), "测评时长：": timeStr,
	}
	now := time.Now()
	dateStr := fmt.Sprintf("%d年%d月%d日", now.Year(), now.Month(), now.Day())

	rfMap := map[string]string{
		"name": "姓名：", "age": "年龄：", "gender": "性别：",
		"affiliation": "单位：", "post": "岗位：", "telephone": "联系方式：",
	}
	if requiredFields != "" {
		allowed := map[string]bool{"姓名：": true}
		for _, f := range strings.Split(requiredFields, ",") {
			if label, ok := rfMap[f]; ok {
				allowed[label] = true
			}
		}
		for k := range fields {
			if !allowed[k] {
				fields[k] = ""
			}
		}
	}

	outputBuf, err := h.processTemplate(templateFile, fields, dateStr, scores, false) // async always full version
	if err != nil {
		slog.Error("report-async: processTemplate failed", "paperId", paperID, "error", err)
		return
	}

	day := now.Format("20060102")
	outDir := filepath.Join(h.outputDir, day)
	_ = os.MkdirAll(outDir, 0o755)
	baseName := fmt.Sprintf("%s_%s_%s", tester.Name, mbtiType, now.Format("20060102150405"))
	docxPath := filepath.Join(outDir, baseName+".docx")
	if err := os.WriteFile(docxPath, outputBuf.Bytes(), 0o644); err != nil {
		slog.Error("report-async: write docx failed", "paperId", paperID, "error", err)
		return
	}

	finalPath := docxPath
	if pdfRet, convErr := convertDocxToPdf(docxPath, outDir); convErr == nil {
		finalPath = pdfRet
		_ = os.Remove(docxPath)
	} else {
		slog.Error("report-async: libreoffice failed", "paperId", paperID, "error", convErr)
	}

	updates := map[string]interface{}{"pdf_path": finalPath, "pdf_flag": 1, "update_time": &now}
	h.db.Table("el_tester").Where("paper_id = ?", paperID).Updates(updates)
	h.db.Table("el_candidate").Where("paper_id = ?", paperID).Updates(updates)
	slog.Info("report-async: generated", "paperId", paperID, "file", filepath.Base(finalPath))

	// 同时生成简版报告
	h.generateSimpleAsync(paperID, tester, mbtiType, scores, dateStr, now)
}

// generateSimpleAsync 异步生成简版报告（完整版生成后自动调用）
func (h *MbtiReportHandler) generateSimpleAsync(paperID string, tester mbtiTesterRow, mbtiType string, scores map[string]int, dateStr string, now time.Time) {
	simpleTemplate := h.findTemplateInDir(h.simpleDir, mbtiType)
	if simpleTemplate == "" {
		slog.Warn("report-async-simple: template not found", "paperId", paperID, "type", mbtiType)
		return
	}

	gender := ""
	if tester.Gender != nil {
		if *tester.Gender == "0" {
			gender = "男"
		} else if *tester.Gender == "1" {
			gender = "女"
		}
	}
	simpleFields := map[string]string{
		"姓名：": tester.Name,
		"年龄：": intToStr(tester.Age),
		"性别：": gender,
	}

	outputBuf, err := h.processTemplate(simpleTemplate, simpleFields, dateStr, scores, true)
	if err != nil {
		slog.Error("report-async-simple: processTemplate failed", "paperId", paperID, "error", err)
		return
	}

	day := now.Format("20060102")
	outDir := filepath.Join(h.outputDir, day)
	baseName := fmt.Sprintf("%s_%s_%s_simple", tester.Name, mbtiType, now.Format("20060102150405"))
	docxPath := filepath.Join(outDir, baseName+".docx")
	if err := os.WriteFile(docxPath, outputBuf.Bytes(), 0o644); err != nil {
		slog.Error("report-async-simple: write failed", "paperId", paperID, "error", err)
		return
	}

	// 转 PDF（全局锁，避免并发冲突）
	if pdfRet, convErr := convertDocxToPdf(docxPath, outDir); convErr == nil {
		_ = os.Remove(docxPath)
		_ = pdfRet
	} else {
		slog.Error("report-async-simple: libreoffice failed", "paperId", paperID, "error", convErr)
	}
	slog.Info("report-async-simple: generated", "paperId", paperID, "type", mbtiType)
}

// POST /exam/api/mbti/download-report {paperId, type: "full"|"simple"}
// 下载已生成的报告文件
func (h *MbtiReportHandler) DownloadReport(c *gin.Context) {
	var b struct {
		PaperID    string `json:"paperId" form:"paperId"`
		ReportType string `json:"type" form:"type"`
	}
	_ = c.ShouldBind(&b)
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}

	var pdfPath string
	if err := h.db.Table("el_tester").Where("paper_id = ?", b.PaperID).Pluck("pdf_path", &pdfPath).Error; err != nil || pdfPath == "" {
		h.db.Table("el_candidate").Where("paper_id = ?", b.PaperID).Pluck("pdf_path", &pdfPath)
	}
	if pdfPath == "" {
		response.RestErr(c, "报告未生成")
		return
	}

	// 简版：在同目录搜索含 _simple 的文件（因为完整版和简版时间戳可能不同）
	if b.ReportType == "simple" {
		dir := filepath.Dir(pdfPath)
		// 从完整版文件名提取姓名和类型前缀（格式: 姓名_TYPE_时间戳.ext）
		baseName := filepath.Base(pdfPath)
		// 找到同目录下最新的 _simple 文件
		entries, err := os.ReadDir(dir)
		if err != nil {
			response.RestErr(c, "报告目录不存在")
			return
		}
		// 提取前缀：姓名_TYPE_
		parts := strings.SplitN(baseName, "_", 3)
		prefix := ""
		if len(parts) >= 2 {
			prefix = parts[0] + "_" + parts[1] + "_"
		}
		var simplePath string
		for _, e := range entries {
			name := e.Name()
			if strings.Contains(name, "_simple") && (prefix == "" || strings.HasPrefix(name, prefix)) {
				simplePath = filepath.Join(dir, name)
			}
		}
		if simplePath == "" {
			response.RestErr(c, "简版报告未生成，请先生成简版报告")
			return
		}
		pdfPath = simplePath
	}

	if _, err := os.Stat(pdfPath); err != nil {
		response.RestErr(c, "报告文件不存在")
		return
	}

	fname := filepath.Base(pdfPath)
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fname)+
		"; filename*=UTF-8''"+url.QueryEscape(fname))
	// 根据扩展名设置 Content-Type
	ext := strings.ToLower(filepath.Ext(fname))
	switch ext {
	case ".pdf":
		c.Header("Content-Type", "application/pdf")
	case ".docx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	default:
		c.Header("Content-Type", "application/octet-stream")
	}
	c.File(pdfPath)
}

// findTemplate 查找对应类型的完整版 docx 模板
func (h *MbtiReportHandler) findTemplate(mbtiType string) string {
	return h.findTemplateInDir(h.templateDir, mbtiType)
}

// findTemplateInDir 在指定目录查找包含 mbtiType 的 docx 文件
func (h *MbtiReportHandler) findTemplateInDir(dir, mbtiType string) string {
	if dir == "" {
		return ""
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), mbtiType) && strings.HasSuffix(e.Name(), ".docx") && !strings.HasPrefix(e.Name(), "~") {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

// processTemplate 读取 docx 模板，替换首页字段和图表数据
func (h *MbtiReportHandler) processTemplate(templatePath string, fields map[string]string, dateStr string, scores map[string]int, isSimple bool) (*bytes.Buffer, error) {
	// 读取原始 docx (ZIP)
	r, err := zip.OpenReader(templatePath)
	if err != nil {
		return nil, fmt.Errorf("打开模板失败: %w", err)
	}
	defer r.Close()

	// 创建输出 ZIP buffer
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		switch f.Name {
		case "word/document.xml":
			if isSimple {
				data = h.replaceDocumentFieldsSimple(data, fields, dateStr)
			} else {
				data = h.replaceDocumentFields(data, fields, dateStr)
			}
		case "word/charts/chart1.xml":
			if !isSimple {
				data = h.replaceChartValues(data, scores)
			}
		}

		// 写入新 ZIP
		// 目录条目(以/结尾)用 Store，文件条目用 Deflate
		method := zip.Deflate
		if strings.HasSuffix(f.Name, "/") || len(data) == 0 {
			method = zip.Store
		}
		header := &zip.FileHeader{
			Name:   f.Name,
			Method: method,
		}
		writer, err := w.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(data); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

// replaceDocumentFieldsSimple 简版模板的字段替换
// 简版模板结构：标签后跟一个带下划线样式的空格占位 <w:t>
// 策略：替换值 + 去掉下划线样式
//
// 兼容标签被 Word 拆分到多个 <w:r> run 的情况（例如"姓名"和"："分两段）
func (h *MbtiReportHandler) replaceDocumentFieldsSimple(data []byte, fields map[string]string, dateStr string) []byte {
	content := string(data)

	for label, value := range fields {
		// 先尝试整体匹配 "label</w:t>"
		idx := strings.Index(content, label+"</w:t>")
		searchStart := -1
		if idx >= 0 {
			searchStart = idx + len(label) + len("</w:t>")
		} else {
			// 拆分匹配：label 可能被多个 <w:r>...<w:t> 拆开（如"姓名" + "："分两段）
			// 策略：找 label 第一个字符的位置，然后用正则跳过中间标签校验剩余字符
			labelRunes := []rune(label)
			if len(labelRunes) < 2 {
				continue
			}
			// 找 label 首字符所在 <w:t>
			firstChar := string(labelRunes[0])
			pos := 0
			for {
				p := strings.Index(content[pos:], firstChar)
				if p < 0 {
					break
				}
				absPos := pos + p
				// 检查是否在 <w:t>...</w:t> 中（粗略：往前 50 字符内有 <w:t）
				lookback := absPos - 50
				if lookback < 0 {
					lookback = 0
				}
				if !strings.Contains(content[lookback:absPos], "<w:t") {
					pos = absPos + 1
					continue
				}
				// 从 absPos 开始，提取后续 200 字符内的纯文本（剥离 XML 标签），看是否能拼出完整 label
				tail := content[absPos:]
				if len(tail) > 1000 {
					tail = tail[:1000]
				}
				stripped := stripXmlTags(tail)
				if strings.HasPrefix(stripped, label) {
					// 找到了 label 末尾在原 XML 中的位置
					endIdx := findLabelEndInXml(tail, label)
					if endIdx > 0 {
						searchStart = absPos + endIdx
						break
					}
				}
				pos = absPos + 1
			}
		}

		if searchStart < 0 {
			continue
		}

		// 在 searchStart 之后找下一个 <w:t...>内容</w:t> 作为占位符
		afterLabel := content[searchStart:]
		reT := regexp.MustCompile(`(<w:t[^>]*>)([^<]*)(</w:t>)`)
		matchT := reT.FindStringSubmatchIndex(afterLabel)
		if matchT == nil {
			continue
		}

		fillValue := value
		if fillValue == "" {
			fillValue = "  "
		}
		// 只替换文本内容，不修改模板格式（保留下划线等样式）
		replaced := afterLabel[:matchT[4]] + fillValue + afterLabel[matchT[5]:]
		content = content[:searchStart] + replaced
	}

	// 替换日期占位符
	// 模板原文是 "2025年XX月XX日"，但被 Word 拆到多个 <w:t> run，简单正则无法匹配。
	// 用 replaceDocxDate 跨标签拼接文本匹配后整段替换。
	content = replaceDocxDate(content, dateStr)

	return []byte(content)
}

// stripXmlTags 简单剥离 XML 标签，只保留文本内容
func stripXmlTags(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(s, "")
}

// replaceDocxDate 替换 docx 中的日期占位符（兼容跨 <w:t> run 拆分）。
//
// 模板原文形如 "2025年XX月XX日"，但 Word 常把它拆到多个 <w:t> run（年/XX/月/XX/日 各一个）。
// 策略：
//  1. 用正则找出所有 <w:t>...</w:t> 块；
//  2. 寻找第一个包含 "20\d\d" 的块作为起点；
//  3. 把该块及其后 8 个 <w:t> 块的文本拼接，剥离 XML 后用 \d{4}年X+月X+日 匹配；
//  4. 命中后：把第一个 <w:t> 内容替换为 dateStr，其余命中范围内的 <w:t> 内容清空。
//
// 这种做法保留了首个 run 的字体/下划线样式，与原占位符位置/格式一致。
func replaceDocxDate(content, dateStr string) string {
	wt := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
	matches := wt.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content
	}
	dateRe := regexp.MustCompile(`\d{4}年X+月X+日`)

	// 输出 buffer：按原文复制，命中范围按规则改写
	var b strings.Builder
	last := 0
	i := 0
	for i < len(matches) {
		m := matches[i]
		txt := content[m[2]:m[3]]
		if !regexp.MustCompile(`20\d{2}`).MatchString(txt) {
			i++
			continue
		}
		// 拼接 i..i+8 的文本
		joinEnd := i + 9
		if joinEnd > len(matches) {
			joinEnd = len(matches)
		}
		var sb strings.Builder
		for j := i; j < joinEnd; j++ {
			sb.WriteString(content[matches[j][2]:matches[j][3]])
		}
		joined := sb.String()
		loc := dateRe.FindStringIndex(joined)
		if loc == nil {
			i++
			continue
		}
		// 找出命中跨越的 run 范围 [i, k]
		// loc[0] = 0（必须从 i 开始）；找出 k 使前缀长度刚好 ≥ loc[1]
		acc := 0
		k := i
		for j := i; j < joinEnd; j++ {
			acc += matches[j][3] - matches[j][2]
			if acc >= loc[1] {
				k = j
				break
			}
			k = j
		}
		// 写出 last..matches[i][2]
		b.WriteString(content[last:matches[i][2]])
		b.WriteString(dateStr)
		// 跳到 matches[i][3]
		cursor := matches[i][3]
		// 对 i+1..k 的每个 <w:t>，复制其前缀（XML 标签等）+ 空文本 + </w:t>
		for j := i + 1; j <= k; j++ {
			// 复制 cursor..matches[j][2]（中间的 XML 结构）
			b.WriteString(content[cursor:matches[j][2]])
			// 文本清空
			cursor = matches[j][3]
		}
		// 推进
		last = cursor
		i = k + 1
	}
	if last < len(content) {
		b.WriteString(content[last:])
	}
	return b.String()
}

// findLabelEndInXml 在 XML 片段中找到 label 文本结束后的 XML 偏移
// 例如 label="姓名："，XML="姓名</w:t>...<w:t>：</w:t>..."，返回 "</w:t>" 之后的位置
func findLabelEndInXml(xmlStr, label string) int {
	labelPos := 0
	labelRunes := []rune(label)
	i := 0
	for i < len(xmlStr) && labelPos < len(labelRunes) {
		// 跳过 XML 标签
		if xmlStr[i] == '<' {
			closeIdx := strings.Index(xmlStr[i:], ">")
			if closeIdx < 0 {
				return -1
			}
			i += closeIdx + 1
			continue
		}
		// 比较一个 rune
		r, sz := utf8.DecodeRuneInString(xmlStr[i:])
		if r == labelRunes[labelPos] {
			labelPos++
			i += sz
		} else {
			return -1
		}
	}
	if labelPos < len(labelRunes) {
		return -1
	}
	// 找到 label 末尾，跳过紧随其后的 </w:t>
	if strings.HasPrefix(xmlStr[i:], "</w:t>") {
		i += len("</w:t>")
	}
	return i
}

// replaceDocumentFields 在 document.xml 中替换首页字段值（完整版）
// 模板 XML 结构：每个字段在独立的 <w:p> 段落中
// 策略：有值 → 替换占位符并去掉下划线；无值 → 删除整个段落（隐藏该字段）
func (h *MbtiReportHandler) replaceDocumentFields(data []byte, fields map[string]string, dateStr string) []byte {
	content := string(data)

	for label, value := range fields {
		idx := strings.Index(content, label+"</w:t>")
		if idx < 0 {
			continue
		}

		if value == "" {
			// 无值：删除包含该标签的整个 <w:p>...</w:p> 段落
			pStart := strings.LastIndex(content[:idx], "<w:p ")
			pStart2 := strings.LastIndex(content[:idx], "<w:p>")
			if pStart2 > pStart {
				pStart = pStart2
			}
			if pStart < 0 {
				continue
			}
			pEnd := strings.Index(content[idx:], "</w:p>")
			if pEnd < 0 {
				continue
			}
			pEnd = idx + pEnd + len("</w:p>")
			content = content[:pStart] + content[pEnd:]
			continue
		}

		// 有值：替换占位符内容为值，去掉下划线样式
		afterLabel := content[idx+len(label)+len("</w:t>"):]
		reT := regexp.MustCompile(`(<w:t[^>]*>)([^<]*)(</w:t>)`)
		matchT := reT.FindStringSubmatch(afterLabel)
		if matchT != nil {
			newT := matchT[1] + value + matchT[3]
			rebuilt := strings.Replace(afterLabel, matchT[0], newT, 1)
			rebuilt = strings.Replace(rebuilt, `<w:u w:val="single"/>`, "", 1)
			content = content[:idx+len(label)+len("</w:t>")] + rebuilt
		}
	}

	// 替换报告日期（兼容跨 <w:t> run 拆分，与简版一致）
	content = replaceDocxDate(content, dateStr)

	return []byte(content)
}

// replaceChartValues 替换 chart1.xml 中的 8 个柱状图数值
// 图表值顺序: E, I, S, N, T, F, J, P
func (h *MbtiReportHandler) replaceChartValues(data []byte, scores map[string]int) []byte {
	content := string(data)

	// 找到 <c:val> 段中的 <c:v> 标签，按顺序替换
	order := []string{"E", "I", "S", "N", "T", "F", "J", "P"}
	values := make([]string, 8)
	for i, k := range order {
		values[i] = strconv.Itoa(scores[k])
	}

	// 提取 <c:val>...</c:val> 段
	re := regexp.MustCompile(`(<c:val>)(.*?)(</c:val>)`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		// 替换其中的 <c:v> 值
		idx := 0
		reV := regexp.MustCompile(`<c:v>[^<]+</c:v>`)
		result := reV.ReplaceAllStringFunc(match, func(vMatch string) string {
			if idx < len(values) {
				newV := fmt.Sprintf("<c:v>%s</c:v>", values[idx])
				idx++
				return newV
			}
			return vMatch
		})
		return result
	})

	// 同时更新 ptCount
	rePt := regexp.MustCompile(`<c:ptCount val="\d+"`)
	content = rePt.ReplaceAllString(content, `<c:ptCount val="8"`)

	return []byte(content)
}

// 辅助函数
func intToStr(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ScoresJSON 用于序列化
func scoresToJSON(scores map[string]int) string {
	b, _ := json.Marshal(scores)
	return string(b)
}

// ServeReportFile 直接返回 docx 文件流（兼容 pdfDownload 下载方式）
func (h *MbtiReportHandler) ServeReportFile(c *gin.Context) {
	file := c.PostForm("file")
	if file == "" {
		var b struct {
			File string `json:"file"`
		}
		_ = c.ShouldBindJSON(&b)
		file = b.File
	}
	if file == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	// 安全检查：只允许访问 outputDir 下的文件
	absFile, _ := filepath.Abs(file)
	absDir, _ := filepath.Abs(h.outputDir)
	if !strings.HasPrefix(absFile, absDir) {
		c.Status(http.StatusForbidden)
		return
	}
	if _, err := os.Stat(file); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.File(file)
}

// ===================== 模板管理 API =====================

var mbtiTypes = []string{
	"ENFJ", "ENFP", "ENTJ", "ENTP",
	"ESFJ", "ESFP", "ESTJ", "ESTP",
	"INFJ", "INFP", "INTJ", "INTP",
	"ISFJ", "ISFP", "ISTJ", "ISTP",
}

// GET /exam/api/mbti/templates — 列出所有模板
func (h *MbtiReportHandler) ListTemplates(c *gin.Context) {
	type tmplInfo struct {
		Type           string `json:"type"`
		FileName       string `json:"fileName"`
		Size           int64  `json:"size"`
		ModTime        string `json:"modTime"`
		Exists         bool   `json:"exists"`
		SimpleFileName string `json:"simpleFileName"`
		SimpleSize     int64  `json:"simpleSize"`
		SimpleModTime  string `json:"simpleModTime"`
		SimpleExists   bool   `json:"simpleExists"`
	}
	var list []tmplInfo
	for _, t := range mbtiTypes {
		name := "MBTI-" + t + ".docx"
		fp := filepath.Join(h.templateDir, name)
		info := tmplInfo{Type: t, FileName: name}
		if fi, err := os.Stat(fp); err == nil {
			info.Exists = true
			info.Size = fi.Size()
			info.ModTime = fi.ModTime().Format("2006-01-02 15:04:05")
		}
		// 简版
		simpleName := "MBTI-SIMPLE-" + t + ".docx"
		simpleFp := filepath.Join(h.simpleDir, simpleName)
		info.SimpleFileName = simpleName
		if fi, err := os.Stat(simpleFp); err == nil {
			info.SimpleExists = true
			info.SimpleSize = fi.Size()
			info.SimpleModTime = fi.ModTime().Format("2006-01-02 15:04:05")
		}
		list = append(list, info)
	}
	response.Rest(c, list)
}

// GET /exam/api/mbti/templates/download/:type?variant=simple — 下载模板
func (h *MbtiReportHandler) DownloadTemplate(c *gin.Context) {
	mbtiType := strings.ToUpper(c.Param("type"))
	valid := false
	for _, t := range mbtiTypes {
		if t == mbtiType {
			valid = true
			break
		}
	}
	if !valid {
		response.RestErr(c, "无效的 MBTI 类型")
		return
	}
	isSimple := c.Query("variant") == "simple"
	var fp, fname string
	if isSimple {
		fname = "MBTI-SIMPLE-" + mbtiType + ".docx"
		fp = filepath.Join(h.simpleDir, fname)
	} else {
		fname = "MBTI-" + mbtiType + ".docx"
		fp = filepath.Join(h.templateDir, fname)
	}
	if _, err := os.Stat(fp); err != nil {
		response.RestErr(c, "模板文件不存在")
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+fname)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.File(fp)
}

// POST /exam/api/mbti/templates/upload — 上传模板（multipart: type + file + variant）
func (h *MbtiReportHandler) UploadTemplate(c *gin.Context) {
	mbtiType := strings.ToUpper(c.PostForm("type"))
	valid := false
	for _, t := range mbtiTypes {
		if t == mbtiType {
			valid = true
			break
		}
	}
	if !valid {
		response.RestErr(c, "无效的 MBTI 类型")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.RestErr(c, "请选择文件")
		return
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".docx") {
		response.RestErr(c, "只支持 .docx 格式")
		return
	}
	if file.Size > 10*1024*1024 {
		response.RestErr(c, "文件大小不能超过 10MB")
		return
	}

	isSimple := c.PostForm("variant") == "simple"
	var dst string
	if isSimple {
		dst = filepath.Join(h.simpleDir, "MBTI-SIMPLE-"+mbtiType+".docx")
	} else {
		dst = filepath.Join(h.templateDir, "MBTI-"+mbtiType+".docx")
	}
	// 备份旧文件
	if _, err := os.Stat(dst); err == nil {
		backup := dst + ".bak"
		_ = os.Rename(dst, backup)
	}
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.RestErr(c, "保存文件失败: "+err.Error())
		return
	}
	slog.Info("template: uploaded", "type", mbtiType, "bytes", file.Size)
	response.Rest(c, gin.H{"type": mbtiType, "size": file.Size})
}

// loMutex ȫ������LibreOffice ��֧�ֲ���ͬʱ�򿪣����� user profile���������ᱨ javaldx ʧ�ܡ�
// ���� docx��pdf ת���Ŷ�ִ�С�
var loMutex sync.Mutex

// convertDocxToPdf �� LibreOffice �� docx ת pdf�����������ļ�·����
// ת��ʧ��ʱ����ԭ docx ·�������ף���
//
// �� -env:UserInstallation ��ÿ��ת������ profile Ŀ¼��������ʷ lock �ļ��������ţ�
// ͬʱ��ȫ�� mutex ��֤��������LO �Բ�֧����ȫ������ʹ profile ��ͬ����
func convertDocxToPdf(docxPath, outDir string) (string, error) {
loMutex.Lock()
defer loMutex.Unlock()

loCmd := "libreoffice"
if runtime.GOOS == "windows" {
loCmd = `C:\Program Files\LibreOffice\program\soffice.exe`
}
absDocx, _ := filepath.Abs(docxPath)
absOutDir, _ := filepath.Abs(outDir)
// ���� profile��ÿ�����Ŀ¼��ת���겻ɾ���� fontconfig cache ���ü�����������
profile := filepath.Join(os.TempDir(), "lo-profile-"+filepath.Base(docxPath))
envArg := "-env:UserInstallation=file://" + profile
cmd := exec.Command(loCmd, envArg, "--headless", "--convert-to", "pdf", "--outdir", absOutDir, absDocx)
out, err := cmd.CombinedOutput()
if err != nil {
slog.Error("libreoffice: convert failed", "docx", docxPath, "error", err, "output", string(out))
return docxPath, err
}
base := filepath.Base(docxPath)
pdfName := base[:len(base)-len(filepath.Ext(base))] + ".pdf"
pdfPath := filepath.Join(outDir, pdfName)
if _, statErr := os.Stat(pdfPath); statErr != nil {
slog.Info("libreoffice: PDF not found after convert", "path", pdfPath)
return docxPath, statErr
}
return pdfPath, nil
}
