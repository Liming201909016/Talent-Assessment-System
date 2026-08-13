package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/graphpdf"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
	"github.com/xuri/excelize/v2"
)

const (
	maxPhase1WordTemplateBytes = 20 << 20
	phase1ChartWorkbookPath    = "word/embeddings/competency-phase1-chart-data.xlsx"
)

var (
	wordTemplateTokenPattern     = regexp.MustCompile(`\{\{[a-zA-Z0-9_.-]+\}\}`)
	wordContentControlPattern    = regexp.MustCompile(`(?s)<w:sdt>.*?</w:sdt>`)
	wordContentControlTagPattern = regexp.MustCompile(`<w:tag\s+w:val="([a-zA-Z0-9_.-]+)"\s*/>`)
	wordTextPattern              = regexp.MustCompile(`(?s)(<w:t(?:\s[^>]*)?>)(.*?)(</w:t>)`)
	numericChartBlockPattern     = regexp.MustCompile(`(?s)<c:(?:numCache|numLit)>.*?</c:(?:numCache|numLit)>`)
	chartValuePattern            = regexp.MustCompile(`<c:v>[^<]*</c:v>`)
)

type phase1DocumentConverter interface {
	Convert(ctx context.Context, fileName string, docx []byte) ([]byte, error)
}

type phase1WordReportRenderer struct {
	templatePath     string
	converter        phase1DocumentConverter
	timeout          time.Duration
	fallbackChromium bool
}

type phase1WordPayload struct {
	Result     service.Phase1ReportResult       `json:"result"`
	Groups     []service.Phase1ReportGroup      `json:"groups"`
	Dimensions []service.Phase1ReportDimension  `json:"dimensions"`
	Validity   service.Phase1ReportValidity     `json:"validity"`
	ReportText service.Phase1ReportTextSnapshot `json:"reportText"`
	Meta       struct {
		GeneratedAt           time.Time         `json:"generatedAt"`
		UserTime              any               `json:"userTime"`
		DimensionCoreMeanings map[string]string `json:"dimensionCoreMeanings"`
	} `json:"meta"`
}

func newPhase1WordReportRenderer(cfg *config.Config) *phase1WordReportRenderer {
	if cfg == nil || !cfg.Phase1WordReport.Enabled {
		return nil
	}
	var converter phase1DocumentConverter
	switch strings.ToLower(strings.TrimSpace(cfg.Phase1WordReport.Converter)) {
	case "", "libreoffice":
		converter = libreofficepdf.NewClient(cfg.Phase1WordReport.LibreOfficePath)
	case "graph":
		graphCfg := graphpdf.Config{
			TenantID: cfg.Phase1WordReport.GraphTenantID, ClientID: cfg.Phase1WordReport.GraphClientID,
			ClientSecret: cfg.Phase1WordReport.GraphClientSecret, DriveID: cfg.Phase1WordReport.GraphDriveID,
			Folder: cfg.Phase1WordReport.GraphFolder, TimeoutSeconds: cfg.Phase1WordReport.GraphTimeoutSeconds,
		}
		converter = graphpdf.NewClient(graphCfg, nil)
	}
	timeout := time.Duration(cfg.Phase1WordReport.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	return &phase1WordReportRenderer{templatePath: cfg.Phase1WordReport.TemplatePath, converter: converter, timeout: timeout, fallbackChromium: cfg.Phase1WordReport.FallbackChromium}
}

func (r *phase1WordReportRenderer) Render(ctx context.Context, paperID string, data map[string]any) ([]byte, error) {
	if r == nil || r.converter == nil || strings.TrimSpace(r.templatePath) == "" {
		return nil, errors.New("一期Word报告渲染器未配置")
	}
	info, err := os.Stat(r.templatePath)
	if err != nil || info.IsDir() || info.Size() <= 0 || info.Size() > maxPhase1WordTemplateBytes {
		return nil, errors.New("一期Word报告模板不可用")
	}
	template, err := os.ReadFile(r.templatePath)
	if err != nil {
		return nil, errors.New("读取一期Word报告模板失败")
	}
	tokens, charts, err := buildPhase1WordTemplateData(data)
	if err != nil {
		return nil, err
	}
	docx, err := renderPhase1WordTemplate(template, tokens, charts)
	if err != nil {
		return nil, err
	}
	convertCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	fileName := fmt.Sprintf("competency-phase1-%d.docx", time.Now().UnixNano())
	return r.converter.Convert(convertCtx, fileName, docx)
}

func buildPhase1WordTemplateData(data map[string]any) (map[string]string, map[string][]float64, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, nil, errors.New("序列化一期Word报告数据失败")
	}
	var payload phase1WordPayload
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, nil, errors.New("解析一期Word报告数据失败")
	}
	if len(payload.Groups) != 2 || len(payload.Dimensions) != 10 || payload.Result.OverallScore == nil {
		return nil, nil, errors.New("一期Word报告数据不完整")
	}
	if payload.ReportText.OverallText == "" || payload.ReportText.ValidityText == "" || payload.ReportText.Disclaimer == "" {
		return nil, nil, errors.New("一期Word报告正式文案不完整")
	}
	generatedAt := payload.Meta.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}
	submittedAt := generatedAt
	if payload.Result.SubmittedAt != nil {
		submittedAt = *payload.Result.SubmittedAt
	}
	tokens := map[string]string{
		"{{report.date}}":             formatChineseDate(generatedAt),
		"{{participant.name}}":        payload.Result.ParticipantName,
		"{{participant.age}}":         formatOptionalInt(payload.Result.ParticipantAge),
		"{{participant.gender}}":      phase1GenderLabel(payload.Result.ParticipantGender),
		"{{participant.telephone}}":   payload.Result.ParticipantTelephone,
		"{{participant.affiliation}}": payload.Result.ParticipantAffiliation,
		"{{participant.post}}":        payload.Result.ParticipantPost,
		"{{result.submittedAt}}":      formatChineseDate(submittedAt),
		"{{result.userTime}}":         formatWordNumber(payload.Meta.UserTime),
		"{{overall.level}}":           phase1OverallLevelLabel(payload.Result.OverallLevel),
		"{{overall.diagnosis}}":       payload.ReportText.OverallText,
		"{{validity.notice}}":         payload.ReportText.ValidityText,
		"{{report.disclaimer}}":       payload.ReportText.Disclaimer,
	}
	charts := make(map[string][]float64, 12)
	groupScores := make([]float64, 0, 2)
	for _, group := range payload.Groups {
		if group.GroupScore == nil {
			return nil, nil, errors.New("一期Word报告一级维度分数不完整")
		}
		score := group.GroupScore.InexactFloat64()
		groupScores = append(groupScores, score)
		prefix := "{{group." + group.GroupCode
		tokens[prefix+".score}}"] = group.GroupScore.StringFixed(2)
		tokens[prefix+".level}}"] = phase1GroupLevelLabel(group.LevelCode)
		tokens[prefix+".description}}"] = payload.ReportText.GroupTexts[group.GroupCode]
	}
	charts["word/charts/chart1.xml"] = groupScores
	dimensionScores := make([]float64, 0, 10)
	for index, dimension := range payload.Dimensions {
		if dimension.DimensionScore == nil || strings.TrimSpace(dimension.DimensionID) == "" {
			return nil, nil, errors.New("一期Word报告二级维度分数不完整")
		}
		score := dimension.DimensionScore.InexactFloat64()
		dimensionScores = append(dimensionScores, score)
		prefix := "{{dimension." + dimension.DimensionID
		tokens[prefix+".score}}"] = dimension.DimensionScore.StringFixed(2)
		tokens[prefix+".level}}"] = phase1DimensionLevelLabel(dimension.LevelCode)
		tokens[prefix+".diagnosis}}"] = payload.ReportText.DimensionTexts[dimension.DimensionID]
		charts[fmt.Sprintf("word/charts/chart%d.xml", index+3)] = []float64{score, decimal.NewFromInt(5).Sub(*dimension.DimensionScore).InexactFloat64()}
	}
	charts["word/charts/chart2.xml"] = dimensionScores
	for token, value := range tokens {
		if strings.TrimSpace(value) == "" && strings.Contains(token, ".diagnosis}}") {
			return nil, nil, fmt.Errorf("一期Word报告文案缺失：%s", token)
		}
	}
	return tokens, charts, nil
}

func renderPhase1WordTemplate(template []byte, tokens map[string]string, charts map[string][]float64) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(template), int64(len(template)))
	if err != nil {
		return nil, errors.New("打开一期Word报告模板失败")
	}
	output := new(bytes.Buffer)
	writer := zip.NewWriter(output)
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, errors.New("读取一期Word报告模板失败")
		}
		body, readErr := io.ReadAll(io.LimitReader(rc, maxPhase1WordTemplateBytes+1))
		rc.Close()
		if readErr != nil || len(body) > maxPhase1WordTemplateBytes {
			return nil, errors.New("一期Word报告模板内容无效")
		}
		if file.Name == "word/document.xml" {
			body, err = replaceWordTemplateTokens(body, tokens)
			if err != nil {
				return nil, err
			}
		} else if file.Name == phase1ChartWorkbookPath {
			body, err = replacePhase1EmbeddedChartWorkbook(body, charts)
			if err != nil {
				return nil, err
			}
		} else if values, ok := charts[file.Name]; ok {
			body, err = replaceWordChartValues(body, values)
			if err != nil {
				return nil, fmt.Errorf("更新一期Word报告图表失败：%s", file.Name)
			}
		}
		method := uint16(zip.Deflate)
		if strings.HasSuffix(file.Name, "/") || len(body) == 0 {
			method = zip.Store
		}
		entry, err := writer.CreateHeader(&zip.FileHeader{Name: file.Name, Method: method})
		if err != nil {
			return nil, errors.New("创建一期Word报告文件失败")
		}
		if _, err := entry.Write(body); err != nil {
			return nil, errors.New("写入一期Word报告文件失败")
		}
	}
	if err := writer.Close(); err != nil {
		return nil, errors.New("完成一期Word报告文件失败")
	}
	return output.Bytes(), nil
}

func replacePhase1EmbeddedChartWorkbook(workbook []byte, charts map[string][]float64) ([]byte, error) {
	book, err := excelize.OpenReader(bytes.NewReader(workbook))
	if err != nil {
		return nil, errors.New("打开一期Word内嵌图表数据失败")
	}
	defer book.Close()
	groupScores := charts["word/charts/chart1.xml"]
	dimensionScores := charts["word/charts/chart2.xml"]
	if len(groupScores) != 2 || len(dimensionScores) != 10 {
		return nil, errors.New("一期Word内嵌图表数据不完整")
	}
	for index, value := range groupScores {
		if err := book.SetCellValue("Sheet1", fmt.Sprintf("B%d", index+2), value); err != nil {
			return nil, errors.New("更新一期Word内嵌一级维度数据失败")
		}
	}
	for index, value := range dimensionScores {
		if err := book.SetCellValue("Sheet2", fmt.Sprintf("B%d", index+4), value); err != nil {
			return nil, errors.New("更新一期Word内嵌雷达图数据失败")
		}
	}
	for index := 0; index < 10; index++ {
		values := charts[fmt.Sprintf("word/charts/chart%d.xml", index+3)]
		if len(values) != 2 {
			return nil, fmt.Errorf("一期Word内嵌环形图数据不完整：chart%d", index+3)
		}
		row := index + 33
		if err := book.SetCellValue("Sheet2", fmt.Sprintf("B%d", row), values[0]); err != nil {
			return nil, errors.New("更新一期Word内嵌环形图数据失败")
		}
		if err := book.SetCellValue("Sheet2", fmt.Sprintf("C%d", row), values[1]); err != nil {
			return nil, errors.New("更新一期Word内嵌环形图数据失败")
		}
	}
	buffer, err := book.WriteToBuffer()
	if err != nil {
		return nil, errors.New("保存一期Word内嵌图表数据失败")
	}
	return buffer.Bytes(), nil
}

func replaceWordTemplateTokens(document []byte, tokens map[string]string) ([]byte, error) {
	content := string(document)
	for token, value := range tokens {
		var escaped bytes.Buffer
		if err := xml.EscapeText(&escaped, []byte(value)); err != nil {
			return nil, errors.New("转义一期Word报告数据失败")
		}
		tag := strings.TrimSuffix(strings.TrimPrefix(token, "{{"), "}}")
		matches := wordContentControlPattern.FindAllStringIndex(content, -1)
		controlMatches := make([][]int, 0, 1)
		for _, bounds := range matches {
			control := content[bounds[0]:bounds[1]]
			tagMatch := wordContentControlTagPattern.FindStringSubmatch(control)
			if len(tagMatch) == 2 && tagMatch[1] == tag {
				controlMatches = append(controlMatches, bounds)
			}
		}
		if len(controlMatches) > 1 {
			return nil, fmt.Errorf("一期Word报告模板存在重复内容控件：%s", tag)
		}
		if len(controlMatches) == 1 {
			bounds := controlMatches[0]
			control, err := replaceWordContentControlText(content[bounds[0]:bounds[1]], escaped.String())
			if err != nil {
				return nil, fmt.Errorf("一期Word报告内容控件无效：%s", tag)
			}
			content = content[:bounds[0]] + control + content[bounds[1]:]
			continue
		}
		if !strings.Contains(content, token) {
			return nil, fmt.Errorf("一期Word报告模板缺少必需字段：%s", tag)
		}
		content = strings.ReplaceAll(content, token, escaped.String())
	}
	if unresolved := wordTemplateTokenPattern.FindString(content); unresolved != "" {
		return nil, fmt.Errorf("一期Word报告模板存在未映射占位符：%s", unresolved)
	}
	return []byte(content), nil
}

func replaceWordContentControlText(control, escapedValue string) (string, error) {
	contentStart := strings.Index(control, "<w:sdtContent>")
	contentEnd := strings.LastIndex(control, "</w:sdtContent>")
	if contentStart < 0 || contentEnd <= contentStart {
		return "", errors.New("content control body missing")
	}
	contentStart += len("<w:sdtContent>")
	body := control[contentStart:contentEnd]
	matches := wordTextPattern.FindAllStringSubmatchIndex(body, -1)
	if len(matches) == 0 {
		return "", errors.New("content control text missing")
	}
	var rebuilt strings.Builder
	last := 0
	for index, match := range matches {
		rebuilt.WriteString(body[last:match[4]])
		if index == 0 {
			rebuilt.WriteString(escapedValue)
		}
		last = match[5]
	}
	rebuilt.WriteString(body[last:])
	return control[:contentStart] + rebuilt.String() + control[contentEnd:], nil
}

func replaceWordChartValues(chart []byte, values []float64) ([]byte, error) {
	content := string(chart)
	blocks := numericChartBlockPattern.FindAllStringIndex(content, -1)
	if len(blocks) == 0 {
		return nil, errors.New("图表数值缓存不存在")
	}
	for _, bounds := range blocks {
		block := content[bounds[0]:bounds[1]]
		matches := chartValuePattern.FindAllStringIndex(block, -1)
		if len(matches) != len(values) {
			continue
		}
		var rebuilt strings.Builder
		last := 0
		for index, match := range matches {
			rebuilt.WriteString(block[last:match[0]])
			rebuilt.WriteString("<c:v>")
			rebuilt.WriteString(strconv.FormatFloat(values[index], 'f', -1, 64))
			rebuilt.WriteString("</c:v>")
			last = match[1]
		}
		rebuilt.WriteString(block[last:])
		return []byte(content[:bounds[0]] + rebuilt.String() + content[bounds[1]:]), nil
	}
	return nil, errors.New("图表数值数量与模板不一致")
}

func formatChineseDate(value time.Time) string {
	return fmt.Sprintf("%d年%d月%d日", value.Year(), int(value.Month()), value.Day())
}

func formatOptionalInt(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func formatWordNumber(value any) string {
	switch typed := value.(type) {
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case json.Number:
		return typed.String()
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func phase1GenderLabel(value string) string {
	switch strings.TrimSpace(value) {
	case "0":
		return "男"
	case "1":
		return "女"
	default:
		return strings.TrimSpace(value)
	}
}

func phase1OverallLevelLabel(value string) string {
	return map[string]string{"excellent": "优秀胜任", "good": "良好胜任", "qualified": "合格胜任", "weak": "薄弱胜任", "unqualified": "尚未胜任"}[value]
}

func phase1GroupLevelLabel(value string) string {
	return map[string]string{"L1": "低分", "L2": "较低分", "L3": "中等分", "L4": "较高分", "L5": "高分"}[value]
}

func phase1DimensionLevelLabel(value string) string {
	return map[string]string{"L1": "差", "L2": "较差", "L3": "合格", "L4": "较优秀", "L5": "优秀"}[value]
}
