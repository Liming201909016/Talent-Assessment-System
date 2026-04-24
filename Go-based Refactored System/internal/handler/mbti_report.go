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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// MbtiReportHandler 生成 MBTI 报告 (docx) — 支持完整版和简版
type MbtiReportHandler struct {
	db              *gorm.DB
	templateDir     string // 完整版 16 个 docx 模板所在目录
	simpleDir       string // 简版 16 个 docx 模板所在目录
	outputDir       string // 生成的报告存放目录
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

// POST /exam/api/mbti/generate-report {paperId, type: "full"|"simple"}
// 根据 MBTI 类型选择 docx 模板 → 替换首页字段 + 图表数据 → 存文件 → 返回下载路径
func (h *MbtiReportHandler) GenerateReport(c *gin.Context) {
	var b struct {
		PaperID    string `json:"paperId"`
		ReportType string `json:"type"` // "full"(默认) 或 "simple"
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
	day := now.Format("20060102")
	outDir := filepath.Join(h.outputDir, day)
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

	// 8. docx → PDF（通过 LibreOffice 转换）
	absDocx, _ := filepath.Abs(docxPath)
	absOutDir, _ := filepath.Abs(outDir)
	pdfPath := filepath.Join(outDir, baseName+".pdf")
	finalPath := docxPath // 默认返回 docx
	loCmd := "libreoffice"
	if runtime.GOOS == "windows" {
		loCmd = `C:\Program Files\LibreOffice\program\soffice.exe`
	}
	cmd := exec.Command(loCmd, "--headless", "--convert-to", "pdf", "--outdir", absOutDir, absDocx)
	cmdOut, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		slog.Error("libreoffice: convert failed", "error", cmdErr, "output", string(cmdOut))
	} else {
		// 转换成功，检查 PDF 是否存在
		if _, err := os.Stat(pdfPath); err == nil {
			finalPath = pdfPath
			_ = os.Remove(docxPath) // 删除临时 docx
		} else {
			slog.Info("libreoffice: PDF not found after convert", "path", pdfPath)
		}
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

	absDocx, _ := filepath.Abs(docxPath)
	absOutDir, _ := filepath.Abs(outDir)
	pdfPath := filepath.Join(outDir, baseName+".pdf")
	finalPath := docxPath
	loCmd := "libreoffice"
	if runtime.GOOS == "windows" {
		loCmd = `C:\Program Files\LibreOffice\program\soffice.exe`
	}
	cmd := exec.Command(loCmd, "--headless", "--convert-to", "pdf", "--outdir", absOutDir, absDocx)
	if cmdOut, cmdErr := cmd.CombinedOutput(); cmdErr != nil {
		slog.Error("report-async: libreoffice failed", "paperId", paperID, "error", cmdErr, "output", string(cmdOut))
	} else if _, err := os.Stat(pdfPath); err == nil {
		finalPath = pdfPath
		_ = os.Remove(docxPath)
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

	// 转 PDF
	absDocx, _ := filepath.Abs(docxPath)
	absOutDir, _ := filepath.Abs(outDir)
	loCmd := "libreoffice"
	if runtime.GOOS == "windows" {
		loCmd = `C:\Program Files\LibreOffice\program\soffice.exe`
	}
	cmd := exec.Command(loCmd, "--headless", "--convert-to", "pdf", "--outdir", absOutDir, absDocx)
	if cmdOut, cmdErr := cmd.CombinedOutput(); cmdErr != nil {
		slog.Error("report-async-simple: libreoffice failed", "paperId", paperID, "error", cmdErr, "output", string(cmdOut))
	} else {
		pdfPath := filepath.Join(outDir, baseName+".pdf")
		if _, err := os.Stat(pdfPath); err == nil {
			_ = os.Remove(docxPath)
		}
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
			data = h.replaceDocumentFields(data, fields, dateStr)
		case "word/charts/chart1.xml":
			if !isSimple { // 简版无图表，跳过替换
				data = h.replaceChartValues(data, scores)
			}
		}

		// 写入新 ZIP — 用 Deflate 压缩，清除 Data Descriptor 标志位
		header := &zip.FileHeader{
			Name:   f.Name,
			Method: zip.Deflate,
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

// replaceDocumentFields 在 document.xml 中替换首页字段值
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

	// 替换报告日期
	re := regexp.MustCompile(`2025年X月X日`)
	content = re.ReplaceAllString(content, dateStr)

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
