package handler

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	phase1V2ChartSeriesPattern = regexp.MustCompile(`(?s)<c:ser>.*?</c:ser>`)
	phase1V2ChartValuePattern  = regexp.MustCompile(`(?s)<c:val>.*?</c:val>`)
	phase1V2WordRunPattern     = regexp.MustCompile(`(?s)<w:r(?:\s[^>]*)?>.*?</w:r>`)
	phase1V2WordRunPrPattern   = regexp.MustCompile(`(?s)<w:rPr>.*?</w:rPr>`)
	phase1V2WordBoldPattern    = regexp.MustCompile(`<w:b(?:Cs)?\b[^>]*/>`)
)

func renderPhase1V2WordTemplate(template []byte, fields map[string]string, charts map[string][][]float64, requiredFields string) ([]byte, error) {
	if len(template) == 0 || len(template) > maxPhase1WordTemplateBytes {
		return nil, errors.New("v2 Word报告模板大小无效")
	}
	reader, err := zip.NewReader(bytes.NewReader(template), int64(len(template)))
	if err != nil {
		return nil, errors.New("打开v2 Word报告模板失败")
	}
	parts := make(map[string][]byte, len(reader.File))
	for _, file := range reader.File {
		rc, openErr := file.Open()
		if openErr != nil {
			return nil, errors.New("读取v2 Word报告模板失败")
		}
		body, readErr := io.ReadAll(io.LimitReader(rc, maxPhase1WordTemplateBytes+1))
		rc.Close()
		if readErr != nil || len(body) > maxPhase1WordTemplateBytes {
			return nil, errors.New("v2 Word报告模板内容无效")
		}
		parts[file.Name] = body
	}
	for name, body := range parts {
		if strings.HasSuffix(name, ".rels") && bytes.Contains(body, []byte(`TargetMode="External"`)) {
			return nil, fmt.Errorf("v2 Word报告模板包含外部关系：%s", name)
		}
	}
	if err := validatePhase1V2WordFields(parts, fields); err != nil {
		return nil, err
	}
	chartParts, err := resolvePhase1V2BusinessChartParts(parts)
	if err != nil {
		return nil, err
	}
	if err := validatePhase1V2WordCharts(chartParts, charts); err != nil {
		return nil, err
	}
	if strings.TrimSpace(requiredFields) != "" {
		filtered, filterErr := filterPhase1V2WordProfileRows(parts["word/document.xml"], requiredFields)
		if filterErr != nil {
			return nil, fmt.Errorf("过滤v2 Word报告个人信息失败：%w", filterErr)
		}
		parts["word/document.xml"] = filtered
	}
	for name, body := range parts {
		if name == "word/document.xml" || (strings.HasPrefix(name, "word/header") && strings.HasSuffix(name, ".xml")) {
			replaced, replaceErr := replacePhase1V2ContentControlValues(body, fields)
			if replaceErr != nil {
				return nil, fmt.Errorf("更新v2 Word报告内容控件失败：%s：%w", name, replaceErr)
			}
			parts[name] = replaced
		}
	}
	for key, part := range chartParts {
		replaced, replaceErr := replacePhase1V2ChartSeriesValues(parts[part], charts[key])
		if replaceErr != nil {
			return nil, fmt.Errorf("更新v2 Word报告图表失败：%s：%w", key, replaceErr)
		}
		if key == "chart.dimension.comparison" {
			replaced, replaceErr = replacePhase1V2ComparisonBarColors(replaced, charts[key][0])
			if replaceErr != nil {
				return nil, fmt.Errorf("更新v2 Word报告柱形颜色失败：%w", replaceErr)
			}
		}
		parts[part] = replaced
	}

	output := new(bytes.Buffer)
	writer := zip.NewWriter(output)
	for _, file := range reader.File {
		body := parts[file.Name]
		method := uint16(zip.Deflate)
		if strings.HasSuffix(file.Name, "/") || len(body) == 0 {
			method = zip.Store
		}
		entry, createErr := writer.CreateHeader(&zip.FileHeader{Name: file.Name, Method: method})
		if createErr != nil {
			_ = writer.Close()
			return nil, errors.New("创建v2 Word报告失败")
		}
		if _, writeErr := entry.Write(body); writeErr != nil {
			_ = writer.Close()
			return nil, errors.New("写入v2 Word报告失败")
		}
	}
	if err := writer.Close(); err != nil {
		return nil, errors.New("完成v2 Word报告失败")
	}
	return output.Bytes(), nil
}

func filterPhase1V2WordProfileRows(document []byte, requiredFields string) ([]byte, error) {
	configured := make(map[string]bool, 6)
	for _, field := range strings.Split(requiredFields, ",") {
		configured[strings.TrimSpace(field)] = true
	}
	tagToField := map[string]string{
		"participant.name": "name", "participant.age": "age", "participant.gender": "gender",
		"participant.telephone": "telephone", "participant.affiliation": "affiliation", "participant.post": "post",
	}
	content := string(document)
	profileAt := strings.Index(content, `w:val="participant.name"`)
	if profileAt < 0 {
		return nil, errors.New("v2 Word报告模板缺少个人信息表")
	}
	tableStart := strings.LastIndex(content[:profileAt], "<w:tbl>")
	tableEnd := strings.Index(content[profileAt:], "</w:tbl>")
	if tableStart < 0 || tableEnd < 0 {
		return nil, errors.New("v2 Word报告个人信息表结构无效")
	}
	tableEnd += profileAt + len("</w:tbl>")
	table := content[tableStart:tableEnd]
	rowPattern := regexp.MustCompile(`(?s)<w:tr\b.*?</w:tr>`)
	rows := rowPattern.FindAllStringIndex(table, -1)
	for index := len(rows) - 1; index >= 0; index-- {
		bounds := rows[index]
		row := table[bounds[0]:bounds[1]]
		field := ""
		for tag, mappedField := range tagToField {
			if strings.Contains(row, `w:val="`+tag+`"`) {
				field = mappedField
				break
			}
		}
		if field != "" && !configured[field] {
			table = table[:bounds[0]] + table[bounds[1]:]
		}
	}
	return []byte(content[:tableStart] + table + content[tableEnd:]), nil
}

func validatePhase1V2WordFields(parts map[string][]byte, fields map[string]string) error {
	expected := phase1V2WordFieldKeys()
	if len(fields) != len(expected) {
		return fmt.Errorf("v2 Word报告字段数量=%d，要求=%d", len(fields), len(expected))
	}
	for _, key := range expected {
		if _, exists := fields[key]; !exists {
			return fmt.Errorf("v2 Word报告字段缺失：%s", key)
		}
	}
	known := make(map[string]bool, len(expected))
	for _, key := range expected {
		known[key] = true
	}
	seen := make(map[string]bool, len(expected))
	for name, body := range parts {
		if name != "word/document.xml" && !(strings.HasPrefix(name, "word/header") && strings.HasSuffix(name, ".xml")) {
			continue
		}
		for _, match := range wordContentControlTagPattern.FindAllSubmatch(body, -1) {
			key := string(match[1])
			if !known[key] {
				return fmt.Errorf("v2 Word报告模板包含未知字段：%s", key)
			}
			seen[key] = true
		}
	}
	for _, key := range expected {
		if !seen[key] {
			return fmt.Errorf("v2 Word报告模板缺少字段：%s", key)
		}
	}
	return nil
}

func replacePhase1V2ContentControlValues(part []byte, fields map[string]string) ([]byte, error) {
	content := string(part)
	matches := wordContentControlPattern.FindAllStringIndex(content, -1)
	for index := len(matches) - 1; index >= 0; index-- {
		bounds := matches[index]
		control := content[bounds[0]:bounds[1]]
		tagMatch := wordContentControlTagPattern.FindStringSubmatch(control)
		if len(tagMatch) != 2 {
			continue
		}
		value, exists := fields[tagMatch[1]]
		if !exists {
			return nil, fmt.Errorf("字段值缺失：%s", tagMatch[1])
		}
		if (strings.HasPrefix(tagMatch[1], "rule.strength.") || strings.HasPrefix(tagMatch[1], "rule.development.")) && strings.Contains(value, "：") {
			replaced, err := replacePhase1V2SelectedItemControl(control, value)
			if err != nil {
				return nil, fmt.Errorf("内容控件无效：%s", tagMatch[1])
			}
			content = content[:bounds[0]] + replaced + content[bounds[1]:]
			continue
		}
		var escaped bytes.Buffer
		if err := xml.EscapeText(&escaped, []byte(value)); err != nil {
			return nil, fmt.Errorf("字段值转义失败：%s", tagMatch[1])
		}
		replaced, err := replaceWordContentControlText(control, escaped.String())
		if err != nil {
			return nil, fmt.Errorf("内容控件无效：%s", tagMatch[1])
		}
		content = content[:bounds[0]] + replaced + content[bounds[1]:]
	}
	return []byte(content), nil
}

func replacePhase1V2SelectedItemControl(control, value string) (string, error) {
	separator := strings.Index(value, "：")
	if separator <= 0 || separator+len("：") >= len(value) {
		return "", errors.New("selected item value is invalid")
	}
	contentStart := strings.Index(control, "<w:sdtContent>")
	contentEnd := strings.LastIndex(control, "</w:sdtContent>")
	if contentStart < 0 || contentEnd <= contentStart {
		return "", errors.New("selected item control content is invalid")
	}
	contentBodyStart := contentStart + len("<w:sdtContent>")
	body := control[contentBodyStart:contentEnd]
	runs := phase1V2WordRunPattern.FindAllString(body, -1)
	if len(runs) == 0 {
		return "", errors.New("selected item control has no run")
	}
	labelProperties := ""
	bodyProperties := ""
	for _, run := range runs {
		properties := phase1V2WordRunPrPattern.FindString(run)
		if labelProperties == "" && phase1V2WordBoldPattern.MatchString(properties) {
			labelProperties = properties
		}
		if bodyProperties == "" && !phase1V2WordBoldPattern.MatchString(properties) {
			bodyProperties = properties
		}
	}
	if labelProperties == "" {
		labelProperties = phase1V2WordRunPrPattern.FindString(runs[0])
	}
	if bodyProperties == "" {
		bodyProperties = phase1V2WordBoldPattern.ReplaceAllString(labelProperties, "")
	}
	label := value[:separator+len("：")]
	description := value[separator+len("："):]
	escape := func(text string) (string, error) {
		var escaped bytes.Buffer
		if err := xml.EscapeText(&escaped, []byte(text)); err != nil {
			return "", err
		}
		return escaped.String(), nil
	}
	escapedLabel, err := escape(label)
	if err != nil {
		return "", err
	}
	escapedDescription, err := escape(description)
	if err != nil {
		return "", err
	}
	replacement := "<w:r>" + labelProperties + "<w:t>" + escapedLabel + "</w:t></w:r>" +
		"<w:r>" + bodyProperties + "<w:t>" + escapedDescription + "</w:t></w:r>"
	return control[:contentBodyStart] + replacement + control[contentEnd:], nil
}

func resolvePhase1V2BusinessChartParts(parts map[string][]byte) (map[string]string, error) {
	document := parts["word/document.xml"]
	relationships := parts["word/_rels/document.xml.rels"]
	if len(document) == 0 || len(relationships) == 0 {
		return nil, errors.New("v2 Word报告模板缺少正文或关系")
	}
	var relationDocument struct {
		Relationships []struct {
			ID         string `xml:"Id,attr"`
			Target     string `xml:"Target,attr"`
			TargetMode string `xml:"TargetMode,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(relationships, &relationDocument); err != nil {
		return nil, errors.New("v2 Word报告模板正文关系无效")
	}
	relationTargets := make(map[string]string)
	for _, relationship := range relationDocument.Relationships {
		if relationship.TargetMode == "External" {
			return nil, errors.New("v2 Word报告模板不得包含外部关系")
		}
		target := strings.ReplaceAll(relationship.Target, "\\", "/")
		if strings.HasPrefix(target, "charts/") {
			relationTargets[relationship.ID] = "word/" + target
		}
	}
	expected := phase1V2WordChartKeys()
	known := make(map[string]bool, len(expected))
	for _, key := range expected {
		known[key] = true
	}
	resolved := make(map[string]string, len(expected))
	for _, drawing := range wordDrawingPattern.FindAll(document, -1) {
		titleMatch := wordDrawingTitlePattern.FindSubmatch(drawing)
		chartMatch := wordDrawingChartIDPattern.FindSubmatch(drawing)
		if len(titleMatch) != 2 || len(chartMatch) != 2 || !strings.HasPrefix(string(titleMatch[1]), "chart.") {
			continue
		}
		key := string(titleMatch[1])
		if !known[key] {
			return nil, fmt.Errorf("v2 Word报告模板包含未知图表键：%s", key)
		}
		if resolved[key] != "" {
			return nil, fmt.Errorf("v2 Word报告模板图表键重复：%s", key)
		}
		part := relationTargets[string(chartMatch[1])]
		if part == "" || len(parts[part]) == 0 {
			return nil, fmt.Errorf("v2 Word报告模板图表关系无效：%s", key)
		}
		resolved[key] = part
	}
	for _, key := range expected {
		if resolved[key] == "" {
			return nil, fmt.Errorf("v2 Word报告模板缺少图表键：%s", key)
		}
	}
	return resolved, nil
}

func validatePhase1V2WordCharts(parts map[string]string, charts map[string][][]float64) error {
	if len(charts) != len(parts) {
		return fmt.Errorf("v2 Word报告图表数量=%d，要求=%d", len(charts), len(parts))
	}
	for key := range charts {
		if parts[key] == "" {
			return fmt.Errorf("v2 Word报告包含未知图表数据：%s", key)
		}
	}
	return nil
}

func replacePhase1V2ChartSeriesValues(chart []byte, values [][]float64) ([]byte, error) {
	if bytes.Contains(chart, []byte("<c:numRef")) || bytes.Contains(chart, []byte("<c:strRef")) ||
		bytes.Contains(chart, []byte("<c:externalData")) || bytes.Contains(chart, []byte("<c:f>")) {
		return nil, errors.New("图表包含引用、公式或外部数据")
	}
	content := string(chart)
	series := phase1V2ChartSeriesPattern.FindAllStringIndex(content, -1)
	if len(series) != len(values) {
		return nil, fmt.Errorf("图表系列数量=%d，要求=%d", len(series), len(values))
	}
	for index := len(series) - 1; index >= 0; index-- {
		bounds := series[index]
		seriesContent := content[bounds[0]:bounds[1]]
		valueBounds := phase1V2ChartValuePattern.FindStringIndex(seriesContent)
		if valueBounds == nil {
			return nil, errors.New("图表系列缺少数值区域")
		}
		valueContent := seriesContent[valueBounds[0]:valueBounds[1]]
		replaced, err := replaceWordChartValues([]byte(valueContent), values[index])
		if err != nil {
			return nil, err
		}
		seriesContent = seriesContent[:valueBounds[0]] + string(replaced) + seriesContent[valueBounds[1]:]
		content = content[:bounds[0]] + seriesContent + content[bounds[1]:]
	}
	return []byte(content), nil
}

func replacePhase1V2ComparisonBarColors(chart []byte, scores []float64) ([]byte, error) {
	if len(scores) != 10 {
		return nil, fmt.Errorf("十维柱形分值数量=%d，要求=10", len(scores))
	}
	content := string(chart)
	series := phase1V2ChartSeriesPattern.FindAllStringIndex(content, -1)
	if len(series) != 2 {
		return nil, fmt.Errorf("十维组合图系列数量=%d，要求=2", len(series))
	}
	bounds := series[0]
	barSeries := content[bounds[0]:bounds[1]]
	pointPattern := regexp.MustCompile(`(?s)<c:dPt>.*?</c:dPt>`)
	barSeries = pointPattern.ReplaceAllString(barSeries, "")
	insertAt := strings.Index(barSeries, "<c:dLbls>")
	if insertAt < 0 {
		insertAt = strings.Index(barSeries, "<c:cat>")
	}
	if insertAt < 0 {
		insertAt = strings.Index(barSeries, "<c:val>")
	}
	if insertAt < 0 {
		return nil, errors.New("十维柱形系列缺少标签或数据区域")
	}
	var points strings.Builder
	for index, score := range scores {
		if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 100 {
			return nil, fmt.Errorf("十维柱形分值无效：index=%d", index)
		}
		points.WriteString(`<c:dPt><c:idx val="`)
		points.WriteString(strconv.Itoa(index))
		points.WriteString(`"/><c:invertIfNegative val="0"/><c:bubble3D val="0"/><c:spPr><a:solidFill><a:srgbClr val="`)
		points.WriteString(phase1V2ScoreBandColor(score))
		points.WriteString(`"/></a:solidFill><a:ln><a:noFill/></a:ln><a:effectLst/></c:spPr></c:dPt>`)
	}
	barSeries = barSeries[:insertAt] + points.String() + barSeries[insertAt:]
	return []byte(content[:bounds[0]] + barSeries + content[bounds[1]:]), nil
}

func phase1V2ScoreBandColor(score float64) string {
	switch {
	case score >= 90:
		return "00A651"
	case score >= 70:
		return "38B86A"
	case score >= 30:
		return "A8D889"
	case score >= 10:
		return "F2A45F"
	default:
		return "E88937"
	}
}

func phase1V2WordFieldKeys() []string {
	keys := []string{
		"participant.name", "participant.age", "participant.gender", "participant.telephone", "participant.affiliation", "participant.post",
		"result.submittedAt", "result.userTime", "validity.status", "validity.text", "report.disclaimer", "overall.score", "overall.level", "overall.normComparison",
		"rule.moduleSummary", "rule.overallAdvice", "rule.strength.1", "rule.strength.2", "rule.strength.3", "rule.development.1", "rule.development.2",
	}
	for _, module := range []string{"task_management", "interpersonal_management", "self_management"} {
		for _, suffix := range []string{"score", "level", "normComparison"} {
			keys = append(keys, "module."+module+"."+suffix)
		}
	}
	for _, dimension := range []string{"logical_reasoning", "plan_execution", "digital_application", "achievement_orientation", "continuous_learning", "communication", "cooperation", "truth_pragmatism", "self_discipline", "dedication"} {
		for _, suffix := range []string{"score", "level", "performance"} {
			keys = append(keys, "dimension."+dimension+"."+suffix)
		}
	}
	sort.Strings(keys)
	return keys
}

func phase1V2WordChartKeys() []string {
	keys := []string{"chart.overall.score", "chart.dimension.comparison"}
	for _, dimension := range []string{"logical_reasoning", "plan_execution", "digital_application", "achievement_orientation", "continuous_learning", "communication", "cooperation", "truth_pragmatism", "self_discipline", "dedication"} {
		keys = append(keys, "chart.dimension."+dimension)
	}
	return keys
}
