package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/talent-assessment/refactored/internal/service"
)

func buildPhase1V2WordTemplateData(data service.Phase1V2FormalReportData) (map[string]string, map[string][][]float64, error) {
	if data.Report.SchemaVersion != "competency-phase1-report-data-v2" || data.Report.ResultRunID == "" ||
		data.Report.Versions != service.Phase1V2VersionSet() || data.Report.Audience != service.CompetencyReportAudienceFrontlineEmployee ||
		len(data.Report.Modules) != 3 || len(data.Report.Dimensions) != 10 || len(data.ExactScores.Dimensions) != 10 ||
		len(data.DimensionNormScores) != 10 || data.ExactScores.OverallScore == nil {
		return nil, nil, errors.New("v2 Word报告数据不完整")
	}
	fields := map[string]string{
		"participant.name": data.ParticipantName, "participant.age": formatOptionalInt(data.ParticipantAge),
		"participant.gender": phase1GenderLabel(data.ParticipantGender), "participant.telephone": data.ParticipantTelephone,
		"participant.affiliation": data.ParticipantAffiliation, "participant.post": data.ParticipantPost,
		"result.submittedAt": formatPhase1V2Date(data.SubmittedAt), "result.userTime": strconv.Itoa(data.UserTime),
		"validity.status": phase1V2ValidityLabel(data.Report.Validity.Status),
		"validity.text":   data.Report.Validity.DisplayText, "report.disclaimer": data.Report.Disclaimer,
		"overall.score": data.Report.Overall.Score, "overall.level": phase1V2LevelLabel(data.Report.Overall.LevelCode),
		"overall.normComparison": phase1V2OverallComparisonLabel(data.Report.Overall.ComparisonCode),
		"rule.moduleSummary":     phase1V2ModuleSummary(data.Report.Modules), "rule.overallAdvice": phase1V2OverallRuleText(data.Report),
		"rule.strength.1": "", "rule.strength.2": "", "rule.strength.3": "",
		"rule.development.1": "", "rule.development.2": "",
	}
	if len(data.Report.Strengths) == 0 {
		fields["rule.strength.1"] = "暂无明显优势项"
	}
	if len(data.Report.Developments) == 0 {
		fields["rule.development.1"] = "暂无明显待发展项"
	}
	for index, item := range data.Report.Strengths {
		if index >= 3 {
			return nil, nil, errors.New("v2 Word报告优势项超过模板上限")
		}
		fields[fmt.Sprintf("rule.strength.%d", index+1)] = phase1V2SelectedItemText(item)
	}
	for index, item := range data.Report.Developments {
		if index >= 2 {
			return nil, nil, errors.New("v2 Word报告待发展项超过模板上限")
		}
		fields[fmt.Sprintf("rule.development.%d", index+1)] = phase1V2SelectedItemText(item)
	}
	moduleByCode := make(map[string]service.Phase1V2ReportModule, len(data.Report.Modules))
	for _, module := range data.Report.Modules {
		moduleByCode[module.ModuleCode] = module
	}
	for _, code := range []string{service.CompetencyPhase1V2ModuleTask, service.CompetencyPhase1V2ModuleInterpersonal, service.CompetencyPhase1V2ModuleSelf} {
		module, exists := moduleByCode[code]
		if !exists {
			return nil, nil, errors.New("v2 Word报告模块数据不完整")
		}
		fields["module."+code+".score"] = module.Score
		fields["module."+code+".level"] = phase1V2LevelLabel(module.LevelCode)
		fields["module."+code+".normComparison"] = phase1V2ModuleComparisonLabel(module.ComparisonCode)
	}

	scores := make([]float64, 0, 10)
	norms := make([]float64, 0, 10)
	charts := make(map[string][][]float64, 12)
	for index, definition := range service.Phase1V2Dimensions() {
		dimension := data.Report.Dimensions[index]
		exact := data.ExactScores.Dimensions[index]
		if dimension.DimensionID != definition.ID || dimension.StableKey != definition.StableKey || exact.DimensionID != definition.ID || exact.Score == nil {
			return nil, nil, errors.New("v2 Word报告维度顺序或精确分值不一致")
		}
		fields["dimension."+definition.StableKey+".score"] = dimension.Score
		fields["dimension."+definition.StableKey+".level"] = phase1V2LevelLabel(dimension.LevelCode)
		fields["dimension."+definition.StableKey+".performance"] = dimension.PerformanceText
		value, _ := exact.Score.Float64()
		norm, _ := data.DimensionNormScores[index].Float64()
		scores = append(scores, value)
		norms = append(norms, norm)
		charts["chart.dimension."+definition.StableKey] = [][]float64{{value, 100 - value}}
	}
	overall, _ := data.ExactScores.OverallScore.Float64()
	charts["chart.overall.score"] = [][]float64{{overall, 100 - overall}}
	charts["chart.dimension.comparison"] = [][]float64{scores, norms}
	if len(fields) != len(phase1V2WordFieldKeys()) || len(charts) != len(phase1V2WordChartKeys()) {
		return nil, nil, errors.New("v2 Word报告字段或图表数量不完整")
	}
	return fields, charts, nil
}

func phase1V2SelectedItemText(item service.Phase1V2ReportSelectedDimension) string {
	name := strings.TrimSpace(item.DimensionName)
	if name == "" {
		return item.RuleText
	}
	return name + "：" + item.RuleText
}

func (r *phase1WordReportRenderer) RenderPhase1V2(ctx context.Context, data service.Phase1V2FormalReportData) ([]byte, error) {
	if r == nil || r.converter == nil || strings.TrimSpace(r.v2TemplatePath) == "" {
		return nil, errors.New("v2 Word报告渲染器未配置")
	}
	info, err := os.Stat(r.v2TemplatePath)
	if err != nil || info.IsDir() || info.Size() <= 0 || info.Size() > maxPhase1WordTemplateBytes {
		return nil, errors.New("v2 Word报告模板不可用")
	}
	template, err := os.ReadFile(r.v2TemplatePath)
	if err != nil {
		return nil, errors.New("读取v2 Word报告模板失败")
	}
	fields, charts, err := buildPhase1V2WordTemplateData(data)
	if err != nil {
		return nil, err
	}
	docx, err := renderPhase1V2WordTemplate(template, fields, charts, data.RequiredFields)
	if err != nil {
		return nil, err
	}
	convertCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	return r.converter.Convert(convertCtx, fmt.Sprintf("competency-phase1-v2-%d.docx", time.Now().UnixNano()), docx)
}

func formatPhase1V2Date(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatChineseDate(*value)
}

func phase1V2LevelLabel(code string) string {
	return map[string]string{
		service.CompetencyPhase1V2LevelExcellent: "优秀", service.CompetencyPhase1V2LevelGood: "良好",
		service.CompetencyPhase1V2LevelQualified: "合格", service.CompetencyPhase1V2LevelWeak: "薄弱",
		service.CompetencyPhase1V2LevelInsufficient: "不足",
	}[code]
}

func phase1V2ValidityLabel(status string) string {
	return map[string]string{service.CompetencyPhase1ValidityGood: "有效", service.CompetencyPhase1ValidityQuestionable: "存疑"}[status]
}

func phase1V2ModuleComparisonLabel(code string) string {
	return map[string]string{
		service.CompetencyPhase1V2ComparisonStandout: "优势突出", service.CompetencyPhase1V2ComparisonAboveNorm: "略高于常模分",
		service.CompetencyPhase1V2ComparisonAtNorm: "与常模分持平", service.CompetencyPhase1V2ComparisonBelowNorm: "低于常模分",
	}[code]
}

func phase1V2OverallComparisonLabel(code string) string {
	return map[string]string{
		service.CompetencyPhase1V2ComparisonSuperior: "优势突出", service.CompetencyPhase1V2ComparisonAboveNorm: "优于常模分",
		service.CompetencyPhase1V2ComparisonAtNorm: "与常模分持平", service.CompetencyPhase1V2ComparisonBelowNorm: "低于常模分",
	}[code]
}

func phase1V2ModuleSummary(modules []service.Phase1V2ReportModule) string {
	parts := make([]string, 0, len(modules))
	for _, module := range modules {
		parts = append(parts, module.ModuleName+module.ComparisonText)
	}
	return strings.Join(parts, "，") + "。"
}

func phase1V2AdviceText(report service.Phase1V2ReportData) string {
	text := report.Overall.AdviceText
	names := make([]string, 0, len(report.AdviceDimensions))
	for _, dimension := range report.AdviceDimensions {
		names = append(names, dimension.DimensionName)
	}
	replacement := "（" + strings.Join(names, "、") + "）"
	text = strings.ReplaceAll(text, "（受测者的分数最低的两个维度）", replacement)
	text = strings.ReplaceAll(text, "（受测者的分数最低的三个维度）", replacement)
	return text
}

func phase1V2OverallRuleText(report service.Phase1V2ReportData) string {
	assessment := strings.TrimSpace(report.Overall.AssessmentText)
	advice := strings.TrimSpace(phase1V2AdviceText(report))
	if assessment == "" {
		return advice
	}
	if advice == "" {
		return assessment
	}
	return assessment + "\n" + advice
}
