package handler

import (
	"fmt"
	"sort"

	"github.com/talent-assessment/refactored/internal/service"
)

type templateSemanticField struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Repeatable  bool   `json:"repeatable"`
}

func phase1V2TemplateSemanticFields() []templateSemanticField {
	fields := []templateSemanticField{
		{Key: "participant.name", Name: "姓名", Description: "受测者姓名"},
		{Key: "participant.age", Name: "年龄", Description: "受测者年龄"},
		{Key: "participant.gender", Name: "性别", Description: "受测者性别"},
		{Key: "participant.telephone", Name: "手机号", Description: "受测者联系电话"},
		{Key: "participant.affiliation", Name: "单位", Description: "受测者所属单位"},
		{Key: "participant.post", Name: "岗位", Description: "受测者岗位"},
		{Key: "result.submittedAt", Name: "提交时间", Description: "答卷提交时间"},
		{Key: "result.userTime", Name: "作答时长", Description: "作答耗时；当前报告默认隐藏"},
		{Key: "validity.status", Name: "效度状态", Description: "本次作答是否有效"},
		{Key: "validity.text", Name: "效度说明", Description: "效度评价说明；可选"},
		{Key: "report.disclaimer", Name: "免责声明", Description: "经批准的报告免责声明；可选"},
		{Key: "overall.score", Name: "总体得分", Description: "百分制总体得分"},
		{Key: "overall.level", Name: "总体等级", Description: "总体得分对应等级"},
		{Key: "overall.normComparison", Name: "总体常模比较", Description: "总体得分与常模的比较结论"},
		{Key: "rule.moduleSummary", Name: "模块总结", Description: "三个能力模块的综合总结"},
		{Key: "rule.overallAdvice", Name: "总体建议", Description: "总体发展建议"},
	}
	for index := 1; index <= 3; index++ {
		fields = append(fields, templateSemanticField{Key: fmt.Sprintf("rule.strength.%d", index), Name: fmt.Sprintf("优势项%d", index), Description: "优势维度名称及对应等级评价"})
	}
	for index := 1; index <= 2; index++ {
		fields = append(fields, templateSemanticField{Key: fmt.Sprintf("rule.development.%d", index), Name: fmt.Sprintf("待发展项%d", index), Description: "待发展维度名称及对应等级短评"})
	}
	modules := []struct{ key, name string }{
		{"task_management", "任务管理"},
		{"interpersonal_management", "人际管理"},
		{"self_management", "自我管理"},
	}
	for _, module := range modules {
		fields = append(fields,
			templateSemanticField{Key: "module." + module.key + ".score", Name: module.name + "得分", Description: module.name + "模块百分制得分"},
			templateSemanticField{Key: "module." + module.key + ".level", Name: module.name + "等级", Description: module.name + "模块等级"},
			templateSemanticField{Key: "module." + module.key + ".normComparison", Name: module.name + "常模比较", Description: module.name + "模块与常模的比较结论"},
		)
	}
	dimensions := []struct{ key, name string }{
		{"logical_reasoning", "逻辑思维"}, {"plan_execution", "计划执行"},
		{"digital_application", "数字应用"}, {"achievement_orientation", "成就导向"},
		{"continuous_learning", "持续学习"}, {"communication", "沟通表达"},
		{"cooperation", "合作意识"}, {"truth_pragmatism", "求真务实"},
		{"self_discipline", "自律性"}, {"dedication", "敬业奉献"},
	}
	for _, dimension := range dimensions {
		fields = append(fields,
			templateSemanticField{Key: "dimension." + dimension.key + ".score", Name: dimension.name + "得分", Description: dimension.name + "维度百分制得分"},
			templateSemanticField{Key: "dimension." + dimension.key + ".level", Name: dimension.name + "等级", Description: dimension.name + "维度等级"},
			templateSemanticField{Key: "dimension." + dimension.key + ".performance", Name: dimension.name + "表现评价", Description: dimension.name + "对应等级的完整表现评价"},
		)
	}
	for index := range fields {
		fields[index].Repeatable = true
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Key < fields[j].Key })
	return fields
}

func managementTraitsTemplateSemanticFields() []templateSemanticField {
	fields := []templateSemanticField{
		{Key: "participant.name", Name: "姓名", Description: "受测者姓名", Required: true},
		{Key: "participant.gender", Name: "性别", Description: "受测者性别", Required: true},
		{Key: "participant.telephone", Name: "手机号", Description: "受测者联系电话", Required: true},
		{Key: "participant.affiliation", Name: "单位", Description: "受测者所属单位", Required: true},
		{Key: "participant.post", Name: "岗位", Description: "受测者岗位", Required: true},
		{Key: "result.submittedAt", Name: "提交时间", Description: "答卷提交时间", Required: true},
		{Key: "report.testTitle", Name: "报告标题", Description: "管理特质测评报告标题", Required: true},
		{Key: "report.testLabel", Name: "测试标识", Description: "正式或测试环境的报告标识", Required: true},
		{Key: "overall.score", Name: "总体得分", Description: "总体百分制得分；模板也可使用固定数字标签"},
		{Key: "overall.level", Name: "总体等级", Description: "总体得分对应等级", Required: true},
		{Key: "overall.diagnosis", Name: "总体诊断", Description: "总体表现诊断", Required: true},
	}
	for index := 1; index <= 3; index++ {
		fields = append(fields, templateSemanticField{Key: fmt.Sprintf("overall.advice.%d", index), Name: fmt.Sprintf("总体建议%d", index), Description: "总体发展建议", Required: true})
	}
	for _, module := range []struct{ key, name string }{{"self", "自我管理"}, {"interpersonal", "人际管理"}, {"task", "任务管理"}, {"development", "发展潜力"}} {
		fields = append(fields, templateSemanticField{Key: "module." + module.key + ".score", Name: module.name + "得分", Description: module.name + "模块百分制得分；模板也可使用固定数字标签"})
	}
	for _, dimension := range service.ManagementTraitsDimensions() {
		prefix := "dimension." + dimension.Key + "."
		fields = append(fields,
			templateSemanticField{Key: prefix + "score", Name: dimension.Name + "得分", Description: dimension.Name + "维度百分制得分", Required: true},
			templateSemanticField{Key: prefix + "norm", Name: dimension.Name + "常模", Description: dimension.Name + "维度常模分", Required: true},
			templateSemanticField{Key: prefix + "level", Name: dimension.Name + "等级", Description: dimension.Name + "维度等级", Required: true},
			templateSemanticField{Key: prefix + "diagnosis", Name: dimension.Name + "诊断", Description: dimension.Name + "维度表现诊断", Required: true},
			templateSemanticField{Key: prefix + "advice", Name: dimension.Name + "建议", Description: dimension.Name + "维度发展建议", Required: true},
		)
	}
	for _, group := range []struct{ key, name string }{{"high", "优势"}, {"low", "待发展"}} {
		for index := 1; index <= 3; index++ {
			fields = append(fields,
				templateSemanticField{Key: fmt.Sprintf("overview.%s.%d.name", group.key, index), Name: fmt.Sprintf("%s项%d名称", group.name, index), Description: group.name + "维度名称", Required: true},
				templateSemanticField{Key: fmt.Sprintf("overview.%s.%d.text", group.key, index), Name: fmt.Sprintf("%s项%d说明", group.name, index), Description: group.name + "维度摘要说明", Required: true},
			)
		}
	}
	for index := range fields {
		fields[index].Repeatable = true
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Key < fields[j].Key })
	return fields
}
