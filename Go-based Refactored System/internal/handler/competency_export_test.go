package handler

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/xuri/excelize/v2"
)

func TestCompetencyExportWorkbook_ContainsPersistedDynamicResults(t *testing.T) {
	started := time.Date(2026, 7, 25, 9, 0, 0, 0, time.Local)
	submitted := started.Add(15 * time.Minute)
	d01Score := decimal.NewFromFloat(4.5)
	data := competencyExportData{
		Dimensions: []model.ExamCompetencyDimension{
			{ID: "ed1", DimensionID: "d1", DimensionCode: "D01", DimensionName: "沟通表达", DisplayOrder: 1, QuestionCount: 1},
			{ID: "ed2", DimensionID: "d2", DimensionCode: "D02", DimensionName: "人际交往", DisplayOrder: 2, QuestionCount: 1},
		},
		Persons: []competencyExportPerson{{
			PaperID: "p1", ParticipantID: "c1", ParticipantType: "candidate", Name: "测试人员", Telephone: "13812345678",
			StartedAt: &started, SubmittedAt: &submitted, UserTime: 15, TotalQuestionCount: 2, AnsweredQuestionCount: 1,
			OverallScore: d01Score, EvaluationAverage: &d01Score, EvaluationLevel: "high", IsComplete: 0, SubmitType: "timeout", ReportAudience: "leader",
		}},
		DimensionResults: []model.CompetencyDimensionResult{{PaperID: "p1", DimensionID: "d1", DimensionScore: &d01Score}},
		Answers: []competencyExportAnswer{{
			PaperID: "p1", ParticipantID: "c1", Name: "测试人员", Sort: 2, QuestionCode: "D01-Q01",
			QuestionContent: "我能清晰表达。", DimensionCode: "D01", DimensionName: "沟通表达", ObservationPoint: "表达",
			ScoringDirection: "reverse", OptionsSnapshot: `[{"rawValue":2,"label":"不太符合","finalScore":4}]`, RawAnswer: int8Pointer(2), FinalScore: int8Pointer(4), Answered: 1,
		}},
		Questions: []model.ExamCompetencyQuestion{{
			QuestionCode: "D01-Q01", QuestionContent: "我能清晰表达。", ExamDimensionID: "ed1", DimensionItemNo: 1,
			ObservationPoint: "表达", ScoringDirection: "reverse", OptionsSnapshot: `[{"rawValue":2,"label":"不太符合","finalScore":4}]`, SnapshotOrder: 1,
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
	assertWorkbookCell(t, file, "结果汇总", "R1", "D01 沟通表达")
	assertWorkbookCell(t, file, "结果汇总", "R2", "4.500000")
	assertWorkbookCell(t, file, "结果汇总", "S2", "")
	assertWorkbookCell(t, file, "逐题明细", "L2", "2")
	assertWorkbookCell(t, file, "逐题明细", "M2", "不太符合")
	assertWorkbookCell(t, file, "逐题明细", "N2", "4")
	assertWorkbookCell(t, file, "题目字典", "A2", "1")
	assertWorkbookCell(t, file, "题目字典", "B2", "D01-Q01")
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

func int8Pointer(value int8) *int8 { return &value }

func assertWorkbookCell(t *testing.T, file *excelize.File, sheet, cell, want string) {
	t.Helper()
	got, err := file.GetCellValue(sheet, cell)
	if err != nil || got != want {
		t.Fatalf("%s!%s=%q err=%v want=%q", sheet, cell, got, err, want)
	}
}
