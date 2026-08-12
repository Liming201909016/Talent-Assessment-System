package service

import (
	"strings"
	"testing"
)

func TestValidateCompetencyImportRows_ValidAndNoQuestionCountRule(t *testing.T) {
	rows := [][]string{
		CompetencyImportHeaders,
		{"1", "逻辑思维", "维度题", "A1-01-Q01", "1", "面对复杂事项，我会先拆解再处理。", "分解能力", "正向", "启用", "测试题"},
	}
	dimensions := []CompetencyImportDimension{{ID: "competency-a1-01", Order: 1, Name: "逻辑思维", Status: 0}}

	got := ValidateCompetencyImportRows(rows, dimensions, nil, nil)
	if len(got.Errors) != 0 || len(got.ValidRows) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
	row := got.ValidRows[0]
	if row.DimensionID != "competency-a1-01" || row.Direction != CompetencyDirectionForward || row.Status != 0 {
		t.Fatalf("unexpected normalized row: %+v", row)
	}
}

func TestValidateCompetencyImportRows_Phase1MixedDimensionAndValidityQuestions(t *testing.T) {
	rows := [][]string{
		{"维度序号", "维度名称", "题目类型", "题目编号", "维度内题号", "题目内容", "考察点", "计分方向", "启用状态", "备注"},
		{"1", "逻辑思维", "维度题", "A1-01-Q01", "1", "维度题。", "逻辑", "反向", "启用", "一期正式题本"},
		{"1", "逻辑思维", "效度题", "P1-VAL-Q01", "1", "效度题。", "极端美化", "正向", "启用", "一期正式题本"},
	}
	dimensions := []CompetencyImportDimension{{ID: "competency-a1-01", Order: 1, Name: "逻辑思维", Status: 0}}

	got := ValidateCompetencyImportRows(rows, dimensions, nil, nil)
	if len(got.Errors) != 0 || len(got.ValidRows) != 2 {
		t.Fatalf("mixed phase-1 rows = %+v", got)
	}
	if got.ValidRows[0].QuestionType != "dimension" || got.ValidRows[1].QuestionType != "validity" {
		t.Fatalf("question types = %q/%q", got.ValidRows[0].QuestionType, got.ValidRows[1].QuestionType)
	}
	if got.ValidRows[0].Direction != "reverse" || got.ValidRows[1].Direction != "forward" {
		t.Fatalf("directions = %q/%q", got.ValidRows[0].Direction, got.ValidRows[1].Direction)
	}
}

func TestValidateCompetencyImportRows_RejectsUnknownTypeAndReverseValidity(t *testing.T) {
	rows := [][]string{
		{"维度序号", "维度名称", "题目类型", "题目编号", "维度内题号", "题目内容", "考察点", "计分方向", "启用状态", "备注"},
		{"1", "逻辑思维", "未知", "UNKNOWN-Q01", "1", "未知题。", "未知", "正向", "启用", ""},
		{"1", "逻辑思维", "效度题", "P1-VAL-Q01", "1", "效度题。", "极端美化", "反向", "启用", ""},
	}
	dimensions := []CompetencyImportDimension{{ID: "competency-a1-01", Order: 1, Name: "逻辑思维", Status: 0}}

	got := ValidateCompetencyImportRows(rows, dimensions, nil, nil)
	if len(got.Errors) != 2 || len(got.ValidRows) != 0 {
		t.Fatalf("invalid type/direction result = %+v", got)
	}
	all := strings.Join(append(got.Errors[0].Messages, got.Errors[1].Messages...), "|")
	for _, want := range []string{"题目类型只能是维度题或效度题", "效度题计分方向必须是正向"} {
		if !strings.Contains(all, want) {
			t.Errorf("errors missing %q: %s", want, all)
		}
	}
}

func TestValidateCompetencyImportRows_UsesExistingDimensionOrderWithoutGlobalCountAssumption(t *testing.T) {
	rows := [][]string{
		CompetencyImportHeaders,
		{"49", "后续确认维度", "维度题", "FUTURE-Q01", "1", "后续阶段题目。", "后续考察点", "正向", "启用", ""},
	}
	dimensions := []CompetencyImportDimension{{ID: "future-dimension", Order: 49, Name: "后续确认维度", Status: 0}}

	got := ValidateCompetencyImportRows(rows, dimensions, nil, nil)
	if len(got.Errors) != 0 || len(got.ValidRows) != 1 {
		t.Fatalf("existing positive dimension order must not be rejected by a retired 48-dimension limit: %+v", got)
	}
}

func TestValidateCompetencyImportRows_ReportsAllRequiredErrors(t *testing.T) {
	rows := [][]string{
		CompetencyImportHeaders,
		{"0", "错误名称", "", "", "0", "", "", "侧向", "未知", ""},
	}

	got := ValidateCompetencyImportRows(rows, nil, nil, nil)
	if len(got.ValidRows) != 0 || len(got.Errors) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
	message := strings.Join(got.Errors[0].Messages, "|")
	for _, want := range []string{"维度序号", "维度不存在", "题目类型", "题目编号", "维度内题号", "题目内容", "考察点", "计分方向", "启用状态"} {
		if !strings.Contains(message, want) {
			t.Errorf("error message missing %q: %s", want, message)
		}
	}
}

func TestValidateCompetencyImportRows_DetectsMasterAndUniquenessConflicts(t *testing.T) {
	rows := [][]string{
		CompetencyImportHeaders,
		{"1", "错误名称", "维度题", "D01-Q01", "1", "题目一", "考察点一", "正向", "启用", ""},
		{"1", "沟通表达", "维度题", "D01-Q01", "1", "题目二", "考察点二", "反向", "停用", ""},
		{"2", "已停用维度", "维度题", "D02-Q01", "1", "题目三", "考察点三", "正向", "启用", ""},
	}
	dimensions := []CompetencyImportDimension{
		{ID: "competency-d01", Order: 1, Name: "沟通表达", Status: 0},
		{ID: "competency-d02", Order: 2, Name: "已停用维度", Status: 1},
	}
	existingCodes := map[string]struct{}{"D01-Q01": {}}
	existingItems := map[string]struct{}{"competency-d01:dimension:1": {}}

	got := ValidateCompetencyImportRows(rows, dimensions, existingCodes, existingItems)
	if len(got.Errors) != 3 || len(got.ValidRows) != 0 {
		t.Fatalf("unexpected result: %+v", got)
	}
	messages := make([]string, 0, len(got.Errors))
	for _, rowErr := range got.Errors {
		messages = append(messages, strings.Join(rowErr.Messages, "|"))
	}
	all := strings.Join(messages, "|")
	for _, want := range []string{"维度名称与主数据不一致", "题目编号已存在", "维度内题号已存在", "维度已停用", "题目编号在文件内重复", "维度内题号在文件内重复"} {
		if !strings.Contains(all, want) {
			t.Errorf("errors missing %q: %s", want, all)
		}
	}
}

func TestValidateCompetencyImportRows_RejectsInvalidHeader(t *testing.T) {
	got := ValidateCompetencyImportRows([][]string{{"错误表头"}}, nil, nil, nil)
	if len(got.Errors) != 1 || !strings.Contains(strings.Join(got.Errors[0].Messages, "|"), "表头") {
		t.Fatalf("unexpected result: %+v", got)
	}
}

// TestBugFB064_TemplateInstructionRowIsNotImported
// 复现：系统模板第2行是填写说明，预览却将其当作题目并返回错误，导致模板无法正式导入。
func TestBugFB064_TemplateInstructionRowIsNotImported(t *testing.T) {
	rows := [][]string{
		CompetencyImportHeaders,
		{"填写当前有效维度的正整数顺序", "必须与维度主数据名称完全一致", "只能填写维度题或效度题", "全局唯一，例如A1-01-Q01", "同一维度和题目类型内正整数且唯一", "题干，不能为空", "考察点，不能为空", "只能填写正向或反向", "只能填写启用或停用", "可空"},
		{"1", "逻辑思维", "维度题", "UI-TEST-1", "999997", "临时验证题", "导入闭环", "正向", "启用", ""},
	}
	dimensions := []CompetencyImportDimension{{ID: "competency-a1-01", Order: 1, Name: "逻辑思维", Status: 0}}

	got := ValidateCompetencyImportRows(rows, dimensions, nil, nil)
	if len(got.Errors) != 0 || len(got.ValidRows) != 1 || got.ValidRows[0].RowNumber != 3 {
		t.Fatalf("template instruction row must be skipped: %+v", got)
	}
}
