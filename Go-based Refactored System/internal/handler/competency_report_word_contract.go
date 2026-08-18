package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/service"
)

const (
	phase1WordTemplateSchemaV1 = "competency-phase1-template-schema-v1"
	phase1WordTemplateSchemaV2 = "competency-phase1-template-schema-v2"
)

type phase1TemplateFieldDefinition struct {
	Key         string
	Label       string
	ValueType   string
	Required    bool
	Repeatable  bool
	Example     string
	Description string
	Resolve     func(phase1WordPayload) string
}

type phase1TemplateChartDefinition struct {
	Key        string
	Label      string
	LegacyPart string
	ValueCount int
}

var phase1TemplateDimensionNames = map[string]string{
	"competency-a1-01": "逻辑思维", "competency-a1-02": "数字应用", "competency-a1-03": "计划执行",
	"competency-a1-04": "持续学习", "competency-a1-05": "沟通表达", "competency-b1-01": "敬业奉献",
	"competency-b1-02": "求真务实", "competency-b1-03": "自律性", "competency-b1-04": "成就导向",
	"competency-b1-05": "合作意识",
}

func phase1TemplateFieldRegistry() []phase1TemplateFieldDefinition {
	fields := []phase1TemplateFieldDefinition{
		textField("report.date", "报告日期", true, true, "2026年8月13日", "报告生成日期", func(payload phase1WordPayload) string { return formatChineseDate(phase1GeneratedAt(payload)) }),
		textField("report.generatedAt", "报告生成时间", false, true, "2026年8月13日", "报告生成日期，可选字段", func(payload phase1WordPayload) string { return formatChineseDate(phase1GeneratedAt(payload)) }),
		textField("participant.name", "姓名", true, true, "张三", "提交时冻结的人员姓名", func(payload phase1WordPayload) string { return payload.Result.ParticipantName }),
		textField("participant.age", "年龄", true, true, "30", "提交时冻结的人员年龄", func(payload phase1WordPayload) string { return formatOptionalInt(payload.Result.ParticipantAge) }),
		textField("participant.gender", "性别", true, true, "男", "提交时冻结的人员性别", func(payload phase1WordPayload) string { return phase1GenderLabel(payload.Result.ParticipantGender) }),
		textField("participant.telephone", "手机号", true, true, "13800000000", "提交时冻结的人员手机号", func(payload phase1WordPayload) string { return payload.Result.ParticipantTelephone }),
		textField("participant.affiliation", "单位", true, true, "示例单位", "提交时冻结的人员单位", func(payload phase1WordPayload) string { return payload.Result.ParticipantAffiliation }),
		textField("participant.post", "岗位", true, true, "示例岗位", "提交时冻结的人员岗位", func(payload phase1WordPayload) string { return payload.Result.ParticipantPost }),
		textField("participant.degree", "学历", false, true, "本科", "提交时冻结的人员学历，可选字段", func(payload phase1WordPayload) string { return payload.Result.ParticipantDegree }),
		textField("participant.major", "专业", false, true, "计算机科学", "提交时冻结的人员专业，可选字段", func(payload phase1WordPayload) string { return payload.Result.ParticipantMajor }),
		textField("result.submittedAt", "提交日期", true, true, "2026年8月13日", "答卷提交日期", func(payload phase1WordPayload) string { return formatChineseDate(phase1SubmittedAt(payload)) }),
		textField("result.userTime", "作答时长", true, true, "20", "报告数据中的作答分钟数", func(payload phase1WordPayload) string { return formatWordNumber(payload.Meta.UserTime) }),
		decimalField("overall.score", "总体得分", false, true, "35.00", "十个二级维度得分之和，满分50", func(payload phase1WordPayload) string { return payload.Result.OverallScore.StringFixed(2) }),
		decimalField("overall.maxScore", "总体满分", false, true, "50", "一期总体固定满分", func(phase1WordPayload) string { return "50" }),
		decimalField("overall.percentage", "总体百分比", false, true, "70.00", "总体得分占50分满分的百分比", func(payload phase1WordPayload) string {
			return payload.Result.OverallScore.Mul(decimal.NewFromInt(2)).StringFixed(2)
		}),
		textField("overall.level", "总体等级", true, true, "合格胜任", "一期总体五档等级", func(payload phase1WordPayload) string { return phase1OverallLevelLabel(payload.Result.OverallLevel) }),
		textField("overall.diagnosis", "总体诊断", true, false, "此处显示总体诊断。", "正式内容快照中的总体诊断", func(payload phase1WordPayload) string { return payload.ReportText.OverallText }),
		textField("validity.notice", "效度说明", true, true, "本次作答效度良好。", "正式内容快照中的效度提示", func(payload phase1WordPayload) string { return payload.ReportText.ValidityText }),
		textField("report.disclaimer", "免责声明", true, false, "此处显示正式免责声明。", "经批准的正式免责声明", func(payload phase1WordPayload) string { return payload.ReportText.Disclaimer }),
	}
	for _, group := range []struct{ code, label string }{{"general_ability", "通用能力"}, {"psychological_quality", "心理素养"}} {
		code, label := group.code, group.label
		fields = append(fields,
			decimalField("group."+code+".score", label+"得分", true, true, "3.50", label+"下五个二级维度的平均分", func(payload phase1WordPayload) string {
				return phase1GroupByCode(payload, code).GroupScore.StringFixed(2)
			}),
			textField("group."+code+".level", label+"等级", true, true, "较高分", label+"五档等级", func(payload phase1WordPayload) string {
				return phase1GroupLevelLabel(phase1GroupByCode(payload, code).LevelCode)
			}),
			textField("group."+code+".description", label+"说明", true, false, "此处显示一级维度说明。", "正式内容快照中的一级维度说明", func(payload phase1WordPayload) string { return payload.ReportText.GroupTexts[code] }),
		)
	}
	configuration := service.NormalizePhase1CompetencyConfiguration()
	for _, dimensionID := range configuration.DimensionIDs {
		id := dimensionID
		label := phase1TemplateDimensionNames[id]
		fields = append(fields,
			decimalField("dimension."+id+".score", label+"得分", true, true, "3.50", "该二级维度8题平均分，满分5", func(payload phase1WordPayload) string {
				return phase1DimensionByID(payload, id).DimensionScore.StringFixed(2)
			}),
			decimalField("dimension."+id+".remainingScore", label+"距满分差值", false, true, "1.50", "5减去该二级维度得分，用于环形图", func(payload phase1WordPayload) string {
				return decimal.NewFromInt(5).Sub(*phase1DimensionByID(payload, id).DimensionScore).StringFixed(2)
			}),
			decimalField("dimension."+id+".percentage", label+"百分比", false, true, "70.00", "该二级维度得分占5分满分的百分比", func(payload phase1WordPayload) string {
				return phase1DimensionByID(payload, id).DimensionScore.Mul(decimal.NewFromInt(20)).StringFixed(2)
			}),
			textField("dimension."+id+".level", label+"等级", true, true, "合格", "该二级维度五档等级", func(payload phase1WordPayload) string {
				return phase1DimensionLevelLabel(phase1DimensionByID(payload, id).LevelCode)
			}),
			textField("dimension."+id+".diagnosis", label+"诊断", true, false, "此处显示诊断与发展建议。", "正式内容快照中的维度诊断与建议", func(payload phase1WordPayload) string { return payload.ReportText.DimensionTexts[id] }),
		)
	}
	return fields
}

func phase1TemplateChartRegistry() []phase1TemplateChartDefinition {
	charts := []phase1TemplateChartDefinition{
		{Key: "chart.group.overview", Label: "一级维度概览图", LegacyPart: "word/charts/chart1.xml", ValueCount: 2},
		{Key: "chart.dimension.radar", Label: "十个二级维度雷达图", LegacyPart: "word/charts/chart2.xml", ValueCount: 10},
	}
	for index, dimensionID := range service.NormalizePhase1CompetencyConfiguration().DimensionIDs {
		charts = append(charts, phase1TemplateChartDefinition{
			Key: "chart.dimension." + dimensionID, Label: phase1TemplateDimensionNames[dimensionID] + "环形图",
			LegacyPart: "word/charts/chart" + strconv.Itoa(index+3) + ".xml", ValueCount: 2,
		})
	}
	return charts
}

func textField(key, label string, required, repeatable bool, example, description string, resolve func(phase1WordPayload) string) phase1TemplateFieldDefinition {
	return phase1TemplateFieldDefinition{Key: key, Label: label, ValueType: "text", Required: required, Repeatable: repeatable, Example: example, Description: description, Resolve: resolve}
}

func decimalField(key, label string, required, repeatable bool, example, description string, resolve func(phase1WordPayload) string) phase1TemplateFieldDefinition {
	return phase1TemplateFieldDefinition{Key: key, Label: label, ValueType: "decimal", Required: required, Repeatable: repeatable, Example: example, Description: description, Resolve: resolve}
}

func phase1GeneratedAt(payload phase1WordPayload) time.Time {
	if payload.Meta.GeneratedAt.IsZero() {
		return time.Now()
	}
	return payload.Meta.GeneratedAt
}

func phase1SubmittedAt(payload phase1WordPayload) time.Time {
	if payload.Result.SubmittedAt != nil {
		return *payload.Result.SubmittedAt
	}
	return phase1GeneratedAt(payload)
}

func phase1GroupByCode(payload phase1WordPayload, code string) service.Phase1ReportGroup {
	for _, group := range payload.Groups {
		if group.GroupCode == code {
			return group
		}
	}
	return service.Phase1ReportGroup{}
}

func phase1DimensionByID(payload phase1WordPayload, id string) service.Phase1ReportDimension {
	for _, dimension := range payload.Dimensions {
		if dimension.DimensionID == id {
			return dimension
		}
	}
	return service.Phase1ReportDimension{}
}

func phase1FieldDefinition(key string) (phase1TemplateFieldDefinition, bool) {
	key = strings.TrimPrefix(strings.TrimSuffix(key, "}}"), "{{")
	for _, field := range phase1TemplateFieldRegistry() {
		if field.Key == key {
			return field, true
		}
	}
	return phase1TemplateFieldDefinition{}, false
}
