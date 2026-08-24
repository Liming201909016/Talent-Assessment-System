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
	wordParagraphPattern         = regexp.MustCompile(`(?s)<w:p(?:\s[^>]*)?>.*?</w:p>`)
	numericChartBlockPattern     = regexp.MustCompile(`(?s)<c:(?:numCache|numLit)>.*?</c:(?:numCache|numLit)>`)
	chartValuePattern            = regexp.MustCompile(`<c:v>[^<]*</c:v>`)
	wordDrawingPattern           = regexp.MustCompile(`(?s)<wp:(?:anchor|inline)\b.*?</wp:(?:anchor|inline)>`)
	wordDrawingTitlePattern      = regexp.MustCompile(`<wp:docPr\b[^>]*\btitle="([a-zA-Z0-9_.-]+)"[^>]*/>`)
	wordDrawingChartIDPattern    = regexp.MustCompile(`<c:chart\b[^>]*\br:id="([a-zA-Z0-9_.-]+)"[^>]*/>`)
	phase1VisibleChartLabel      = regexp.MustCompile(`(?s)<c:dLbl><c:idx val="0"/>.*?</c:dLbl>`)
	phase1ChartLabelX            = regexp.MustCompile(`<c:x val="([^"]+)"/>`)
	phase1ChartLabelY            = regexp.MustCompile(`<c:y val="([^"]+)"/>`)
	phase1ChartLabelTextProperty = regexp.MustCompile(`<a:(?:rPr|defRPr|endParaRPr)\b[^>]*>`)
	phase1ChartLabelBoldProperty = regexp.MustCompile(`\sb="[^"]*"`)
	phase1ChartLabelBodyProperty = regexp.MustCompile(`<a:bodyPr\b[^>]*>`)
	phase1ChartLabelAnchor       = regexp.MustCompile(`\sanchor="[^"]*"`)
	phase1ChartLabelParagraph    = regexp.MustCompile(`<a:pPr\b[^>]*>`)
	phase1ChartLabelAlignment    = regexp.MustCompile(`\salgn="[^"]*"`)
	phase1ChartLabelShowLegend   = regexp.MustCompile(`<c:showLegendKey\b`)
	phase1ChartAutoTitleDeleted  = regexp.MustCompile(`<c:autoTitleDeleted val="1"/>`)
	phase1ChartExistingTitle     = regexp.MustCompile(`(?s)<c:title>.*?</c:title>`)
	phase1ChartShowValue         = regexp.MustCompile(`<c:showVal val="[^"]*"/>`)
	phase1ChartLabelsBlock       = regexp.MustCompile(`(?s)<c:dLbls>.*?</c:dLbls>`)
	phase1RadarValueAxis         = regexp.MustCompile(`(?s)<c:valAx>.*?</c:valAx>`)
	phase1RadarScaling           = regexp.MustCompile(`(?s)<c:scaling>.*?</c:scaling>`)
	phase1RadarMinimum           = regexp.MustCompile(`<c:min val="[^"]*"/>`)
	phase1RadarMaximum           = regexp.MustCompile(`<c:max val="[^"]*"/>`)
	phase1RadarMajorUnit         = regexp.MustCompile(`<c:majorUnit val="[^"]*"/>`)
	phase1RadarMajorGridlines    = regexp.MustCompile(`(?s)<c:majorGridlines(?:>.*?</c:majorGridlines>|\s*/>)`)
)

var phase1DoughnutTitlePositions = map[int][2]string{
	3:  {"0.4086842105263158", "0.3936363636363636"},
	4:  {"0.4073684210526316", "0.3890909090909091"},
	5:  {"0.4073684210526316", "0.3890909090909091"},
	6:  {"0.41", "0.3913636363636364"},
	7:  {"0.4073684210526316", "0.3858080808080808"},
	8:  {"0.4165789473684211", "0.4277272727272727"},
	9:  {"0.4218421052631579", "0.4027272727272727"},
	10: {"0.4218421052631579", "0.4027272727272727"},
	11: {"0.4257894736842106", "0.405"},
	12: {"0.4152631578947368", "0.405"},
}

var phase1LibreOfficeChartLabelOffsets = map[int][2]decimal.Decimal{
	3:  {decimal.RequireFromString("9.26818182"), decimal.RequireFromString("-22.01682692")},
	4:  {decimal.RequireFromString("32.6"), decimal.RequireFromString("-46.53378378")},
	5:  {decimal.RequireFromString("-2.5"), decimal.RequireFromString("-17.32692308")},
	6:  {decimal.RequireFromString("15"), decimal.RequireFromString("-18.5")},
	7:  {decimal.RequireFromString("3"), decimal.RequireFromString("-25.88679245")},
	8:  {decimal.RequireFromString("6"), decimal.RequireFromString("-21.84166667")},
	9:  {decimal.RequireFromString("35.19047619"), decimal.RequireFromString("-67.97095436")},
	10: {decimal.RequireFromString("27.34"), decimal.RequireFromString("-51")},
	11: {decimal.RequireFromString("-1.92307692"), decimal.RequireFromString("-39.65112994")},
	12: {decimal.RequireFromString("5.32894737"), decimal.RequireFromString("-17.51630435")},
}

type phase1DocumentConverter interface {
	Convert(ctx context.Context, fileName string, docx []byte) ([]byte, error)
}

type phase1WordReportRenderer struct {
	templatePath     string
	converter        phase1DocumentConverter
	timeout          time.Duration
	fallbackChromium bool
	calibrateLabels  bool
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
		RequiredFields        string            `json:"requiredFields"`
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
	_, calibrateLabels := converter.(*libreofficepdf.Client)
	return &phase1WordReportRenderer{templatePath: cfg.Phase1WordReport.TemplatePath, converter: converter, timeout: timeout, fallbackChromium: cfg.Phase1WordReport.FallbackChromium, calibrateLabels: calibrateLabels}
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
	if r.calibrateLabels {
		docx, err = calibratePhase1LibreOfficeChartLabels(docx)
		if err != nil {
			return nil, err
		}
	}
	convertCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	fileName := fmt.Sprintf("competency-phase1-%d.docx", time.Now().UnixNano())
	return r.converter.Convert(convertCtx, fileName, docx)
}

func calibratePhase1LibreOfficeChartLabels(docx []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		return nil, errors.New("打开一期Word报告失败")
	}
	output := new(bytes.Buffer)
	writer := zip.NewWriter(output)
	changed := make(map[int]bool, len(phase1LibreOfficeChartLabelOffsets))
	for _, file := range reader.File {
		rc, openErr := file.Open()
		if openErr != nil {
			return nil, errors.New("读取一期Word报告失败")
		}
		body, readErr := io.ReadAll(io.LimitReader(rc, maxPhase1WordTemplateBytes+1))
		rc.Close()
		if readErr != nil || len(body) > maxPhase1WordTemplateBytes {
			return nil, errors.New("一期Word报告内容无效")
		}
		var chartIndex int
		if _, scanErr := fmt.Sscanf(file.Name, "word/charts/chart%d.xml", &chartIndex); scanErr == nil {
			if offsets, ok := phase1LibreOfficeChartLabelOffsets[chartIndex]; ok {
				body, err = calibratePhase1LibreOfficeChartLabel(body, offsets)
				if err != nil {
					return nil, fmt.Errorf("一期Word报告图表%d标签校准失败", chartIndex)
				}
				changed[chartIndex] = true
			}
		}
		header := file.FileHeader
		part, createErr := writer.CreateHeader(&header)
		if createErr != nil {
			return nil, errors.New("写入一期Word报告失败")
		}
		if _, writeErr := part.Write(body); writeErr != nil {
			return nil, errors.New("写入一期Word报告失败")
		}
	}
	if len(changed) != len(phase1LibreOfficeChartLabelOffsets) {
		return nil, errors.New("一期Word报告环形图标签不完整")
	}
	if err := writer.Close(); err != nil {
		return nil, errors.New("完成一期Word报告失败")
	}
	return output.Bytes(), nil
}

func calibratePhase1LibreOfficeChartLabel(chart []byte, offsets [2]decimal.Decimal) ([]byte, error) {
	if phase1ChartExistingTitle.Match(chart) {
		return chart, nil
	}
	label := phase1VisibleChartLabel.Find(chart)
	if label == nil || len(phase1ChartLabelX.FindAllSubmatch(label, -1)) != 1 || len(phase1ChartLabelY.FindAllSubmatch(label, -1)) != 1 {
		return nil, errors.New("环形图可见标签坐标无效")
	}
	xMatch := phase1ChartLabelX.FindSubmatch(label)
	yMatch := phase1ChartLabelY.FindSubmatch(label)
	x, err := decimal.NewFromString(string(xMatch[1]))
	if err != nil {
		return nil, errors.New("环形图横坐标无效")
	}
	y, err := decimal.NewFromString(string(yMatch[1]))
	if err != nil {
		return nil, errors.New("环形图纵坐标无效")
	}
	x = x.Sub(offsets[0].Div(decimal.NewFromInt(380)))
	y = y.Sub(offsets[1].Div(decimal.NewFromInt(220)))
	calibrated := phase1ChartLabelX.ReplaceAll(label, []byte(`<c:x val="`+x.String()+`"/>`))
	calibrated = phase1ChartLabelY.ReplaceAll(calibrated, []byte(`<c:y val="`+y.String()+`"/>`))
	return bytes.Replace(chart, label, calibrated, 1), nil
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
	if payload.Meta.GeneratedAt.IsZero() {
		payload.Meta.GeneratedAt = time.Now()
	}
	groupsByCode := make(map[string]service.Phase1ReportGroup, len(payload.Groups))
	for _, group := range payload.Groups {
		if group.GroupScore == nil || strings.TrimSpace(group.GroupCode) == "" {
			return nil, nil, errors.New("一期Word报告一级维度分数不完整")
		}
		groupsByCode[group.GroupCode] = group
	}
	for _, code := range []string{"general_ability", "psychological_quality"} {
		if _, ok := groupsByCode[code]; !ok {
			return nil, nil, errors.New("一期Word报告一级维度身份不完整")
		}
	}
	dimensionsByID := make(map[string]service.Phase1ReportDimension, len(payload.Dimensions))
	for _, dimension := range payload.Dimensions {
		if dimension.DimensionScore == nil || strings.TrimSpace(dimension.DimensionID) == "" {
			return nil, nil, errors.New("一期Word报告二级维度分数不完整")
		}
		dimensionsByID[dimension.DimensionID] = dimension
	}
	for _, dimensionID := range service.NormalizePhase1CompetencyConfiguration().DimensionIDs {
		if _, ok := dimensionsByID[dimensionID]; !ok {
			return nil, nil, errors.New("一期Word报告二级维度身份不完整")
		}
	}
	tokens := make(map[string]string, len(phase1TemplateFieldRegistry()))
	for _, field := range phase1TemplateFieldRegistry() {
		tokens["{{"+field.Key+"}}"] = field.Resolve(payload)
	}
	tokens["{{__profile.requiredFields}}"] = strings.TrimSpace(payload.Meta.RequiredFields)
	tokens["{{__validity.status}}"] = strings.TrimSpace(payload.Validity.Status)
	charts := make(map[string][]float64, 12)
	groupScores := make([]float64, 0, 2)
	for _, code := range []string{"general_ability", "psychological_quality"} {
		group := groupsByCode[code]
		score := group.GroupScore.InexactFloat64()
		groupScores = append(groupScores, score)
	}
	charts["chart.group.overview"] = groupScores
	dimensionScores := make([]float64, 0, 10)
	for _, dimensionID := range service.NormalizePhase1CompetencyConfiguration().DimensionIDs {
		dimension := dimensionsByID[dimensionID]
		score := dimension.DimensionScore.InexactFloat64()
		dimensionScores = append(dimensionScores, score)
		charts["chart.dimension."+dimensionID] = []float64{score, decimal.NewFromInt(5).Sub(*dimension.DimensionScore).InexactFloat64()}
	}
	charts["chart.dimension.radar"] = dimensionScores
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
	parts := make(map[string][]byte, len(reader.File))
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
		parts[file.Name] = body
	}
	businessChartParts, err := resolvePhase1BusinessChartParts(parts)
	if err != nil {
		return nil, err
	}
	chartValuesByPart := make(map[string][]float64, len(charts))
	doughnutChartParts := make(map[string]int, 10)
	radarChartPart := ""
	if len(businessChartParts) == 0 {
		for index, chart := range phase1TemplateChartRegistry() {
			chartValuesByPart[chart.LegacyPart] = charts[chart.Key]
			if chart.Key == "chart.dimension.radar" {
				radarChartPart = chart.LegacyPart
			}
			if index >= 2 {
				doughnutChartParts[chart.LegacyPart] = index + 1
			}
		}
	} else {
		for index, chart := range phase1TemplateChartRegistry() {
			part := businessChartParts[chart.Key]
			chartValuesByPart[part] = charts[chart.Key]
			if chart.Key == "chart.dimension.radar" {
				radarChartPart = part
			}
			if index >= 2 {
				doughnutChartParts[part] = index + 1
			}
		}
	}
	output := new(bytes.Buffer)
	writer := zip.NewWriter(output)
	for _, file := range reader.File {
		body := parts[file.Name]
		if file.Name == "word/document.xml" {
			requiredFields, hasProfileConfig := tokens["{{__profile.requiredFields}}"]
			if hasProfileConfig {
				body, err = filterPhase1WordProfileTable(body, requiredFields)
				if err != nil {
					return nil, err
				}
			}
			validityStatus, hasValidityStatus := tokens["{{__validity.status}}"]
			if hasValidityStatus {
				body, err = filterPhase1WordValidityNotice(body, validityStatus)
				if err != nil {
					return nil, err
				}
			}
			visibleTokens := make(map[string]string, len(tokens))
			for token, value := range tokens {
				if token != "{{__profile.requiredFields}}" && token != "{{__validity.status}}" {
					visibleTokens[token] = value
				}
			}
			if hasValidityStatus && validityStatus == service.CompetencyPhase1ValidityGood {
				delete(visibleTokens, "{{validity.notice}}")
			}
			if hasProfileConfig && strings.TrimSpace(requiredFields) != "" {
				configured := make(map[string]bool, 6)
				for _, field := range strings.Split(requiredFields, ",") {
					configured[strings.TrimSpace(field)] = true
				}
				for field, token := range map[string]string{
					"name": "{{participant.name}}", "age": "{{participant.age}}", "gender": "{{participant.gender}}",
					"telephone": "{{participant.telephone}}", "affiliation": "{{participant.affiliation}}", "post": "{{participant.post}}",
				} {
					if !configured[field] {
						delete(visibleTokens, token)
					}
				}
			}
			body, err = replaceWordTemplateTokens(body, visibleTokens)
			if err != nil {
				return nil, err
			}
		} else if file.Name == phase1ChartWorkbookPath {
			body, err = replacePhase1EmbeddedChartWorkbook(body, charts)
			if err != nil {
				return nil, err
			}
		} else if values, ok := chartValuesByPart[file.Name]; ok {
			body, err = replaceWordChartValues(body, values)
			if err != nil {
				return nil, fmt.Errorf("更新一期Word报告图表失败：%s", file.Name)
			}
			if file.Name == radarChartPart {
				body, err = normalizePhase1RadarGrid(body)
				if err != nil {
					return nil, fmt.Errorf("更新一期Word报告雷达图网格失败：%s", file.Name)
				}
			}
			if chartIndex, ok := doughnutChartParts[file.Name]; ok {
				body, err = normalizePhase1DoughnutChartLabel(body, chartIndex, values[0])
				if err != nil {
					return nil, fmt.Errorf("更新一期Word报告环形图格式失败：%s", file.Name)
				}
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

func normalizePhase1RadarGrid(chart []byte) ([]byte, error) {
	axes := phase1RadarValueAxis.FindAll(chart, -1)
	if len(axes) == 0 {
		return nil, errors.New("雷达图值轴不存在")
	}
	normalized := phase1RadarValueAxis.ReplaceAllFunc(chart, func(axis []byte) []byte {
		scaling := phase1RadarScaling.Find(axis)
		if scaling == nil || !phase1RadarMajorGridlines.Match(axis) {
			return axis
		}
		scaling = phase1RadarMinimum.ReplaceAll(scaling, nil)
		scaling = phase1RadarMaximum.ReplaceAll(scaling, nil)
		scaling = bytes.Replace(scaling, []byte(`</c:scaling>`), []byte(`<c:max val="5"/><c:min val="0"/></c:scaling>`), 1)
		axis = phase1RadarScaling.ReplaceAll(axis, scaling)
		if phase1RadarMajorUnit.Match(axis) {
			axis = phase1RadarMajorUnit.ReplaceAll(axis, []byte(`<c:majorUnit val="1"/>`))
		} else {
			axis = bytes.Replace(axis, []byte(`</c:valAx>`), []byte(`<c:majorUnit val="1"/></c:valAx>`), 1)
		}
		return axis
	})
	if bytes.Count(normalized, []byte(`<c:min val="0"/>`)) < len(axes) ||
		bytes.Count(normalized, []byte(`<c:max val="5"/>`)) < len(axes) ||
		bytes.Count(normalized, []byte(`<c:majorUnit val="1"/>`)) < len(axes) {
		return nil, errors.New("雷达图值轴网格结构无效")
	}
	return normalized, nil
}

func normalizePhase1DoughnutChartLabel(chart []byte, chartIndex int, score float64) ([]byte, error) {
	label := phase1VisibleChartLabel.Find(chart)
	if label == nil {
		return nil, errors.New("环形图可见标签不存在")
	}
	position, ok := phase1DoughnutTitlePositions[chartIndex]
	if !ok || phase1ChartExistingTitle.Match(chart) || len(phase1ChartAutoTitleDeleted.FindAll(chart, -1)) != 1 {
		return nil, errors.New("环形图居中标题结构无效")
	}
	properties := phase1ChartLabelTextProperty.FindAllIndex(label, -1)
	bodyProperties := phase1ChartLabelBodyProperty.FindAllIndex(label, -1)
	paragraphProperties := phase1ChartLabelParagraph.FindAllIndex(label, -1)
	if len(properties) == 0 && len(bodyProperties) == 0 && len(paragraphProperties) == 0 {
		if !phase1ChartLabelShowLegend.Match(label) {
			return nil, errors.New("环形图可见标签字体格式不存在")
		}
		textProperties := []byte(`<c:txPr><a:bodyPr anchor="ctr"/><a:lstStyle/><a:p><a:pPr algn="ctr"><a:defRPr b="0"/></a:pPr><a:endParaRPr b="0"/></a:p></c:txPr>`)
		label = phase1ChartLabelShowLegend.ReplaceAll(label, append(textProperties, []byte(`<c:showLegendKey`)...))
	} else {
		if len(properties) == 0 || len(bodyProperties) == 0 || len(paragraphProperties) == 0 {
			return nil, errors.New("环形图可见标签字体格式不存在")
		}
		label = phase1ChartLabelTextProperty.ReplaceAllFunc(label, func(property []byte) []byte {
			return setPhase1ChartLabelAttribute(property, phase1ChartLabelBoldProperty, `b="0"`)
		})
		label = phase1ChartLabelBodyProperty.ReplaceAllFunc(label, func(property []byte) []byte {
			return setPhase1ChartLabelAttribute(property, phase1ChartLabelAnchor, `anchor="ctr"`)
		})
		label = phase1ChartLabelParagraph.ReplaceAllFunc(label, func(property []byte) []byte {
			return setPhase1ChartLabelAttribute(property, phase1ChartLabelAlignment, `algn="ctr"`)
		})
	}
	chart = bytes.Replace(chart, phase1VisibleChartLabel.Find(chart), nil, 1)
	chart = phase1ChartShowValue.ReplaceAll(chart, []byte(`<c:showVal val="0"/>`))
	chart = phase1ChartLabelsBlock.ReplaceAll(chart, nil)
	title := fmt.Sprintf(`<c:title><c:tx><c:rich><a:bodyPr anchor="ctr"/><a:lstStyle/><a:p><a:pPr algn="ctr"><a:defRPr sz="1100" b="0"/></a:pPr><a:r><a:rPr sz="1100" b="0"/><a:t>%.2f</a:t></a:r><a:endParaRPr sz="1100" b="0"/></a:p></c:rich></c:tx><c:layout><c:manualLayout><c:layoutTarget val="outer"/><c:xMode val="edge"/><c:yMode val="edge"/><c:wMode val="factor"/><c:hMode val="factor"/><c:x val="%s"/><c:y val="%s"/><c:w val="0.18"/><c:h val="0.14"/></c:manualLayout></c:layout><c:overlay val="1"/><c:spPr><a:noFill/><a:ln><a:noFill/></a:ln></c:spPr></c:title>`, score, position[0], position[1])
	return phase1ChartAutoTitleDeleted.ReplaceAll(chart, []byte(`<c:autoTitleDeleted val="0"/>`+title)), nil
}

func setPhase1ChartLabelAttribute(property []byte, attribute *regexp.Regexp, replacement string) []byte {
	if attribute.Match(property) {
		return attribute.ReplaceAll(property, []byte(" "+replacement))
	}
	position := len(property) - 1
	if position > 0 && property[position-1] == '/' {
		position--
	}
	result := make([]byte, 0, len(property)+len(replacement)+1)
	result = append(result, property[:position]...)
	result = append(result, ' ')
	result = append(result, replacement...)
	result = append(result, property[position:]...)
	return result
}

func filterPhase1WordValidityNotice(document []byte, status string) ([]byte, error) {
	status = strings.TrimSpace(status)
	if status != service.CompetencyPhase1ValidityGood && status != service.CompetencyPhase1ValidityQuestionable {
		return nil, errors.New("一期Word报告效度状态无效")
	}
	content := string(document)
	matches := wordParagraphPattern.FindAllStringIndex(content, -1)
	paragraphs := make([][]int, 0, 1)
	for _, bounds := range matches {
		if strings.Contains(content[bounds[0]:bounds[1]], `w:val="validity.notice"`) {
			paragraphs = append(paragraphs, bounds)
		}
	}
	if len(paragraphs) != 1 {
		return nil, errors.New("一期Word报告效度提示段结构无效")
	}
	if status == service.CompetencyPhase1ValidityQuestionable {
		return document, nil
	}
	bounds := paragraphs[0]
	return []byte(content[:bounds[0]] + content[bounds[1]:]), nil
}

func replacePhase1EmbeddedChartWorkbook(workbook []byte, charts map[string][]float64) ([]byte, error) {
	book, err := excelize.OpenReader(bytes.NewReader(workbook))
	if err != nil {
		return nil, errors.New("打开一期Word内嵌图表数据失败")
	}
	defer book.Close()
	groupScores := charts["chart.group.overview"]
	dimensionScores := charts["chart.dimension.radar"]
	if len(groupScores) != 2 || len(dimensionScores) != 10 {
		return nil, errors.New("一期Word内嵌图表数据不完整")
	}
	chartDataSheet := "ChartData"
	legacyWorkbook := false
	if index, err := book.GetSheetIndex(chartDataSheet); err != nil || index < 0 {
		legacyWorkbook = true
	}
	for index, value := range groupScores {
		sheet, cell := chartDataSheet, fmt.Sprintf("C%d", index+2)
		if legacyWorkbook {
			sheet, cell = "Sheet1", fmt.Sprintf("B%d", index+2)
		}
		if err := book.SetCellValue(sheet, cell, value); err != nil {
			return nil, errors.New("更新一期Word内嵌一级维度数据失败")
		}
	}
	for index, value := range dimensionScores {
		sheet, cell := chartDataSheet, fmt.Sprintf("C%d", index+4)
		if legacyWorkbook {
			sheet, cell = "Sheet2", fmt.Sprintf("B%d", index+4)
		}
		if err := book.SetCellValue(sheet, cell, value); err != nil {
			return nil, errors.New("更新一期Word内嵌雷达图数据失败")
		}
	}
	for index, dimensionID := range service.NormalizePhase1CompetencyConfiguration().DimensionIDs {
		values := charts["chart.dimension."+dimensionID]
		if len(values) != 2 {
			return nil, fmt.Errorf("一期Word内嵌环形图数据不完整：%s", dimensionID)
		}
		row := index + 4
		sheet, scoreCell, remainderCell := chartDataSheet, fmt.Sprintf("C%d", row), fmt.Sprintf("D%d", row)
		if legacyWorkbook {
			row = index + 33
			sheet, scoreCell, remainderCell = "Sheet2", fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row)
		}
		if err := book.SetCellValue(sheet, scoreCell, values[0]); err != nil {
			return nil, errors.New("更新一期Word内嵌环形图数据失败")
		}
		if err := book.SetCellValue(sheet, remainderCell, values[1]); err != nil {
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
		definition, registered := phase1FieldDefinition(tag)
		if len(controlMatches) > 1 && (!registered || !definition.Repeatable) {
			return nil, fmt.Errorf("一期Word报告模板存在重复内容控件：%s", tag)
		}
		if len(controlMatches) > 0 {
			for index := len(controlMatches) - 1; index >= 0; index-- {
				bounds := controlMatches[index]
				control, err := replaceWordContentControlText(content[bounds[0]:bounds[1]], escaped.String())
				if err != nil {
					return nil, fmt.Errorf("一期Word报告内容控件无效：%s", tag)
				}
				content = content[:bounds[0]] + control + content[bounds[1]:]
			}
			continue
		}
		if !strings.Contains(content, token) {
			if registered && !definition.Required {
				continue
			}
			return nil, fmt.Errorf("一期Word报告模板缺少必需字段：%s", tag)
		}
		content = strings.ReplaceAll(content, token, escaped.String())
	}
	if unresolved := wordTemplateTokenPattern.FindString(content); unresolved != "" {
		return nil, fmt.Errorf("一期Word报告模板存在未映射占位符：%s", unresolved)
	}
	return []byte(content), nil
}

func filterPhase1WordProfileTable(document []byte, requiredFields string) ([]byte, error) {
	configured := make(map[string]bool, 6)
	for _, field := range strings.Split(requiredFields, ",") {
		configured[strings.TrimSpace(field)] = true
	}
	if strings.TrimSpace(requiredFields) == "" {
		for _, field := range []string{"name", "age", "gender", "telephone", "affiliation", "post"} {
			configured[field] = true
		}
	}
	tagToField := map[string]string{
		"participant.name": "name", "participant.age": "age", "participant.gender": "gender",
		"participant.telephone": "telephone", "participant.affiliation": "affiliation", "participant.post": "post",
	}
	content := string(document)
	profileAt := strings.Index(content, `w:val="participant.name"`)
	if profileAt < 0 {
		return nil, errors.New("一期Word报告模板缺少个人信息表")
	}
	tableStart := strings.LastIndex(content[:profileAt], "<w:tbl>")
	tableEnd := strings.Index(content[profileAt:], "</w:tbl>")
	if tableStart < 0 || tableEnd < 0 {
		return nil, errors.New("一期Word报告个人信息表结构无效")
	}
	tableEnd += profileAt + len("</w:tbl>")
	table := content[tableStart:tableEnd]
	rowPattern := regexp.MustCompile(`(?s)<w:tr\b.*?</w:tr>`)
	cellPattern := regexp.MustCompile(`(?s)<w:tc\b.*?</w:tc>`)
	rows := rowPattern.FindAllStringIndex(table, -1)
	for index := len(rows) - 1; index >= 0; index-- {
		bounds := rows[index]
		row := table[bounds[0]:bounds[1]]
		participantRow := false
		keptCells := make([]string, 0, 2)
		for _, cellBounds := range cellPattern.FindAllStringIndex(row, -1) {
			cell := row[cellBounds[0]:cellBounds[1]]
			keep := true
			participantCell := false
			for tag, field := range tagToField {
				if strings.Contains(cell, `w:val="`+tag+`"`) {
					participantRow = true
					participantCell = true
					keep = configured[field]
					break
				}
			}
			if participantCell && keep {
				keptCells = append(keptCells, cell)
			}
		}
		if !participantRow {
			continue
		}
		if len(keptCells) == 0 {
			table = table[:bounds[0]] + table[bounds[1]:]
			continue
		}
		if len(keptCells) == 1 && !strings.Contains(keptCells[0], "<w:gridSpan") {
			keptCells[0] = strings.Replace(keptCells[0], "</w:tcPr>", `<w:gridSpan w:val="2"/></w:tcPr>`, 1)
		}
		rowStartEnd := strings.Index(row, ">") + 1
		row = row[:rowStartEnd] + strings.Join(keptCells, "") + "</w:tr>"
		table = table[:bounds[0]] + row + table[bounds[1]:]
	}
	return []byte(content[:tableStart] + table + content[tableEnd:]), nil
}

func resolvePhase1BusinessChartParts(parts map[string][]byte) (map[string]string, error) {
	document := parts["word/document.xml"]
	relationships := parts["word/_rels/document.xml.rels"]
	if len(document) == 0 || len(relationships) == 0 {
		return nil, nil
	}
	var relationshipDocument struct {
		Relationships []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(relationships, &relationshipDocument); err != nil {
		return nil, errors.New("一期Word报告模板正文关系无效")
	}
	relationTargets := make(map[string]string)
	for _, relationship := range relationshipDocument.Relationships {
		target := strings.ReplaceAll(relationship.Target, "\\", "/")
		if strings.HasPrefix(target, "charts/") {
			relationTargets[relationship.ID] = "word/" + target
		}
	}
	known := make(map[string]bool)
	for _, chart := range phase1TemplateChartRegistry() {
		known[chart.Key] = true
	}
	resolved := make(map[string]string)
	for _, drawing := range wordDrawingPattern.FindAll(document, -1) {
		titleMatch := wordDrawingTitlePattern.FindSubmatch(drawing)
		chartMatch := wordDrawingChartIDPattern.FindSubmatch(drawing)
		if len(titleMatch) != 2 || len(chartMatch) != 2 || !strings.HasPrefix(string(titleMatch[1]), "chart.") {
			continue
		}
		key := string(titleMatch[1])
		if !known[key] {
			return nil, fmt.Errorf("一期Word报告模板包含未知图表业务键：%s", key)
		}
		if resolved[key] != "" {
			return nil, fmt.Errorf("一期Word报告模板存在重复图表业务键：%s", key)
		}
		part := relationTargets[string(chartMatch[1])]
		if part == "" || len(parts[part]) == 0 {
			return nil, fmt.Errorf("一期Word报告模板图表关系无效：%s", key)
		}
		resolved[key] = part
	}
	if len(resolved) == 0 {
		return nil, nil
	}
	for _, chart := range phase1TemplateChartRegistry() {
		if resolved[chart.Key] == "" {
			return nil, fmt.Errorf("一期Word报告模板缺少图表业务键：%s", chart.Key)
		}
	}
	return resolved, nil
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
