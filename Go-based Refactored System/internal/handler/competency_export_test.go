package handler

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/xuri/excelize/v2"
)

func TestCompetencyExportWorkbook_ContainsPersistedDynamicResults(t *testing.T) {
	started := time.Date(2026, 7, 25, 9, 0, 0, 0, time.Local)
	submitted := started.Add(15 * time.Minute)
	d01Score := decimal.NewFromFloat(4.5)
	data := competencyExportData{
		Groups: []model.ExamCompetencyGroup{
			{ID: "eg1", GroupCode: "general_ability", GroupName: "通用能力", DisplayOrder: 1},
			{ID: "eg2", GroupCode: "psychological_quality", GroupName: "心理素养", DisplayOrder: 2},
		},
		Dimensions: []model.ExamCompetencyDimension{
			{ID: "ed1", DimensionID: "d1", DimensionCode: "D01", DimensionName: "沟通表达", DisplayOrder: 1, QuestionCount: 1},
			{ID: "ed2", DimensionID: "d2", DimensionCode: "D02", DimensionName: "人际交往", DisplayOrder: 2, QuestionCount: 1},
		},
		Persons: []competencyExportPerson{{
			PaperID: "p1", ParticipantID: "c1", ParticipantType: "candidate", Name: "测试人员", Telephone: "13812345678",
			StartedAt: &started, SubmittedAt: &submitted, UserTime: 15, TotalQuestionCount: 2, AnsweredQuestionCount: 1,
			OverallScore: &d01Score, EvaluationAverage: &d01Score, EvaluationLevel: "high", IsComplete: 0, SubmitType: "timeout", ReportAudience: "leader",
			ValidityScore: &d01Score, ValidityStatus: "questionable",
		}},
		GroupResults: []model.CompetencyGroupResult{
			{PaperID: "p1", ExamGroupID: "eg1", GroupScore: &d01Score},
			{PaperID: "p1", ExamGroupID: "eg2"},
		},
		DimensionResults: []model.CompetencyDimensionResult{{PaperID: "p1", DimensionID: "d1", DimensionScore: &d01Score}},
		Answers: []competencyExportAnswer{{
			PaperID: "p1", ParticipantID: "c1", Name: "测试人员", Sort: 2, QuestionCode: "D01-Q01",
			QuestionContent: "我能清晰表达。", DimensionCode: "D01", DimensionName: "沟通表达", ObservationPoint: "表达",
			QuestionType: "dimension", ScoringDirection: "reverse", OptionsSnapshot: `[{"rawValue":2,"label":"不太符合","finalScore":4}]`, RawAnswer: int8Pointer(2), FinalScore: int8Pointer(4), Answered: 1,
		}},
		Questions: []model.ExamCompetencyQuestion{{
			QuestionCode: "D01-Q01", QuestionContent: "我能清晰表达。", ExamDimensionID: "ed1", DimensionItemNo: 1,
			CompetencyQuestionType: stringPointer("dimension"), ObservationPoint: "表达", ScoringDirection: "reverse", OptionsSnapshot: `[{"rawValue":2,"label":"不太符合","finalScore":4}]`, SnapshotOrder: 1,
		}},
	}

	file, err := buildCompetencyExportWorkbook(model.Exam{ID: "e1", Title: "胜任力测试"}, data, false)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if got := file.GetSheetList(); len(got) != 3 || got[0] != "结果汇总" || got[1] != "逐题明细" || got[2] != "题目字典" {
		t.Fatalf("sheets=%v", got)
	}
	assertWorkbookCell(t, file, "结果汇总", "A2", "c1")
	assertWorkbookCell(t, file, "结果汇总", "D2", "138****5678")
	assertWorkbookCell(t, file, "结果汇总", "J2", "50.00%")
	assertWorkbookCell(t, file, "结果汇总", "R1", "通用能力")
	assertWorkbookCell(t, file, "结果汇总", "R2", "4.500000")
	assertWorkbookCell(t, file, "结果汇总", "S1", "心理素养")
	assertWorkbookCell(t, file, "结果汇总", "T1", "效度原始分")
	assertWorkbookCell(t, file, "结果汇总", "T2", "4.500000")
	assertWorkbookCell(t, file, "结果汇总", "U2", "questionable")
	assertWorkbookCell(t, file, "结果汇总", "V2", "4.500000")
	assertWorkbookCell(t, file, "结果汇总", "W2", "")
	assertWorkbookCell(t, file, "逐题明细", "G2", "dimension")
	assertWorkbookCell(t, file, "逐题明细", "M2", "2")
	assertWorkbookCell(t, file, "逐题明细", "N2", "不太符合")
	assertWorkbookCell(t, file, "逐题明细", "O2", "4")
	assertWorkbookCell(t, file, "题目字典", "A2", "1")
	assertWorkbookCell(t, file, "题目字典", "B2", "D01-Q01")
	assertWorkbookCell(t, file, "题目字典", "C2", "dimension")
}

func TestCompetencyExportWorkbook_EmptyResultsStillHasThreeHeaders(t *testing.T) {
	file, err := buildCompetencyExportWorkbook(model.Exam{ID: "e1", Title: "空测评"}, competencyExportData{}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	for _, sheet := range []string{"结果汇总", "逐题明细", "题目字典"} {
		value, err := file.GetCellValue(sheet, "A1")
		if err != nil || value == "" {
			t.Fatalf("sheet %s missing header: value=%q err=%v", sheet, value, err)
		}
	}
}

// TestBugFB197_Phase1ExportUsesCompletedV2ResultRuns
// Corresponds to docs/regression-tests.md FB-197.
func TestBugFB197_Phase1ExportUsesCompletedV2ResultRuns(t *testing.T) {
	overallScore := decimal.RequireFromString("68.125000")
	normScore := decimal.RequireFromString("57.750000")
	moduleScore := decimal.RequireFromString("67.500000")
	moduleNorm := decimal.RequireFromString("58.000000")
	dimensionScore := decimal.RequireFromString("68.750000")
	dimensionNorm := decimal.RequireFromString("55.000000")
	validityScore := decimal.RequireFromString("22.000000")
	level := "qualified"
	above := "above_norm"
	validity := "good"
	data := phase1V2ExportData{
		Runs: []model.CompetencyResultRun{{
			ID: "run-v2", PaperID: "paper-v2", ParticipantID: "candidate-v2", ParticipantType: "candidate",
			ParticipantName: "新版人员", ParticipantTelephone: "13812345678", ReportAudience: "frontline_employee",
			ProductVersion: "competency-frontline-phase1-v2", ScoringVersion: "competency-phase1-scoring-v2",
			ContentVersion: "competency-phase1-content-v2", ReportTemplateVersion: "competency-phase1-report-v2", Status: "completed",
		}},
		Overalls: []model.CompetencyResultRunOverall{{
			ResultRunID: "run-v2", TotalQuestionCount: 90, AnsweredQuestionCount: 90, EffectiveDimensionCount: 10,
			OverallScore: &overallScore, LevelCode: &level, NormScore: &normScore, NormComparisonCode: &above, IsComplete: 1, SubmitType: "manual", UserTime: 18,
		}},
		Validities: []model.CompetencyResultRunValidity{{ResultRunID: "run-v2", ValidityScore: &validityScore, ValidityStatus: &validity, IsComplete: 1}},
		Answers: []competencyExportAnswer{{
			PaperID: "paper-v2", ParticipantID: "candidate-v2", Name: "新版人员", Telephone: "13812345678", Sort: 1,
			QuestionCode: "A1-01-Q01", QuestionType: "dimension", QuestionContent: "我会分析问题。",
			DimensionID: "competency-a1-01", DimensionCode: "A1-01", DimensionName: "逻辑思维", ObservationPoint: "分析",
			ScoringDirection: "forward", OptionsSnapshot: `[{"rawValue":4,"label":"比较符合","finalScore":4}]`, RawAnswer: int8Pointer(4), FinalScore: int8Pointer(4), Answered: 1,
		}},
		Questions: []model.ExamCompetencyQuestion{{
			ExamDimensionID: "snapshot-d1", QuestionCode: "A1-01-Q01", CompetencyQuestionType: stringPointer("dimension"),
			QuestionContent: "我会分析问题。", ObservationPoint: "分析", ScoringDirection: "forward", DimensionItemNo: 1, SnapshotOrder: 1,
		}},
		SourceDimensions: []model.ExamCompetencyDimension{{ID: "snapshot-d1", DimensionID: "competency-a1-01"}},
	}
	for index, module := range []struct{ code, name string }{{"task_management", "任务管理类"}, {"interpersonal_management", "人际管理类"}, {"self_management", "自我管理类"}} {
		data.Modules = append(data.Modules, model.CompetencyResultRunModule{ResultRunID: "run-v2", ModuleCode: module.code, ModuleName: module.name, DisplayOrder: index + 1, ModuleScore: &moduleScore, LevelCode: &level, NormScore: &moduleNorm, NormComparisonCode: &above, IsComplete: 1})
	}
	for _, definition := range service.Phase1V2Dimensions() {
		data.Dimensions = append(data.Dimensions, model.CompetencyResultRunDimension{ResultRunID: "run-v2", DimensionID: definition.ID, DimensionCode: definition.DisplayCode, DimensionName: definition.Name, DisplayOrder: definition.DisplayOrder, DimensionScore: &dimensionScore, LevelCode: &level, NormScore: &dimensionNorm, IsComplete: 1})
	}
	file, err := buildPhase1V2CompetencyExportWorkbook(model.Exam{ID: "exam-v2", Title: "新版胜任力"}, data, false)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if got := file.GetSheetList(); len(got) != 3 || got[0] != "结果汇总" || got[1] != "逐题明细" || got[2] != "题目字典" {
		t.Fatalf("sheets=%v", got)
	}
	for cell, want := range map[string]string{
		"A2": "candidate-v2", "D2": "138****5678", "F2": "run-v2", "M2": "competency-frontline-phase1-v2",
		"Q1": "报告对象", "Q2": "基层员工", "R1": "总体得分", "R2": "68.13", "S2": "合格", "U2": "优于常模分",
		"V1": "任务管理类-得分", "V2": "67.50", "Y2": "略高于常模分",
		"AH1": "效度原始分", "AH2": "22", "AI2": "有效",
		"AJ1": "A1-01 逻辑思维-得分", "AJ2": "68.75", "AM2": "高于常模",
	} {
		assertWorkbookCell(t, file, "结果汇总", cell, want)
	}
	for cell, want := range map[string]string{"E2": "run-v2", "J2": "A1-01", "L2": "A1-01", "M2": "逻辑思维", "N2": "任务管理类", "Q2": "4", "R2": "比较符合"} {
		assertWorkbookCell(t, file, "逐题明细", cell, want)
	}
	for cell, want := range map[string]string{"A2": "1", "B2": "A1-01-Q01", "D2": "competency-logical-reasoning", "E2": "A1-01", "F2": "逻辑思维", "G2": "任务管理类"} {
		assertWorkbookCell(t, file, "题目字典", cell, want)
	}
}

func TestBugFB197_Phase1ExportDoesNotFallBackToV1Results(t *testing.T) {
	data := phase1V2ExportData{Runs: make([]model.CompetencyResultRun, 0)}
	file, err := buildPhase1V2CompetencyExportWorkbook(model.Exam{ID: "exam-v2", Title: "无v2结果"}, data, true)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := file.GetRows("结果汇总")
	if err != nil || len(rows) != 1 {
		t.Fatalf("summary rows=%d err=%v, want headers only", len(rows), err)
	}
}

func TestBugFB197_Phase1ExportDispatchesToV2OnlyBuilder(t *testing.T) {
	source := readSourceFile(t, "competency_export.go")
	body := extractFunctionBody(t, source, "func (h *ExamHandler) exportCompetencyWorkbook(")
	for _, required := range []string{"isPhase1V2ExportExam(exam)", "loadPhase1V2CompetencyExportData", "buildPhase1V2CompetencyExportWorkbook"} {
		if !strings.Contains(body, required) {
			t.Errorf("phase-1 export dispatch missing %q", required)
		}
	}
}

func TestCompetencyExportEndpoints_DispatchWithoutChangingLegacyFlow(t *testing.T) {
	source := readSourceFile(t, "exam_pdf.go")
	for _, signature := range []string{"func (h *ExamHandler) ExportRawAnswers(", "func (h *ExamHandler) ExportRawData("} {
		body := extractFunctionBody(t, source, signature)
		for _, required := range []string{"AssessmentTypeCompetency", "exportCompetencyWorkbook"} {
			if !strings.Contains(body, required) {
				t.Errorf("%s missing competency export dispatch %q", signature, required)
			}
		}
	}
	builder := readSourceFile(t, "competency_export.go")
	for _, required := range []string{
		"buildCompetencyExportWorkbook", "el_competency_result", "CompetencyDimensionResult",
		"el_exam_competency_question", "pq.raw_answer", "pq.final_score", "OptionsSnapshot",
	} {
		if !strings.Contains(builder, required) {
			t.Errorf("competency export implementation missing %q", required)
		}
	}
}

func int8Pointer(value int8) *int8       { return &value }
func stringPointer(value string) *string { return &value }

func assertWorkbookCell(t *testing.T, file *excelize.File, sheet, cell, want string) {
	t.Helper()
	got, err := file.GetCellValue(sheet, cell)
	if err != nil || got != want {
		t.Fatalf("%s!%s=%q err=%v want=%q", sheet, cell, got, err, want)
	}
}
