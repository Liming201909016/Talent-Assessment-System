package handler

import (
	"errors"
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
)

func TestRequireLegacyExam_Modes(t *testing.T) {
	tests := []struct {
		name string
		exam model.Exam
		want error
	}{
		{
			name: "legacy continues",
			exam: model.Exam{AssessmentType: service.AssessmentTypeLegacy, ScoringMode: service.ScoringModeLegacy},
		},
		{
			name: "competency rejected",
			exam: model.Exam{AssessmentType: service.AssessmentTypeCompetency, ScoringMode: service.ScoringModeCompetencyAverage},
			want: ErrCompetencyDedicatedAPI,
		},
		{
			name: "invalid pair rejected",
			exam: model.Exam{AssessmentType: service.AssessmentTypeLegacy, ScoringMode: service.ScoringModeCompetencyAverage},
			want: ErrInvalidAssessmentMode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireLegacyExam(&tt.exam)
			if !errors.Is(err, tt.want) {
				t.Fatalf("requireLegacyExam() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestLegacyPaperGuard_RejectsCompetencyBeforeLegacyWork(t *testing.T) {
	tests := []struct {
		file       string
		signature  string
		legacyWork string
	}{
		{"paper.go", "func (h *PaperHandler) CreatePaper(", "h.createPaperTx(&exam)"},
		{"paper.go", "func (h *PaperHandler) PaperDetail(", `Where("exam_id = ?", paper.ExamID)`},
		{"paper.go", "func (h *PaperHandler) QuDetail(", "var pq model.PaperQu"},
		{"paper.go", "func (h *PaperHandler) FillAnswer(", "// 未作答 → 直接返回"},
		{"paper.go", "func (h *PaperHandler) HandExam(", `Table("el_paper_qu")`},
		{"paper.go", "func (h *PaperHandler) PaperResult(", "type quRow struct"},
		{"paper.go", "func (h *PaperHandler) PaperQuDetail(", `Where("exam_id = ?", paper.ExamID)`},
		{"paper.go", "func (h *PaperHandler) PaperStandScore(", "var qs []struct"},
		{"tester_score.go", "func (h *TesterHandler) StandScore(", "queryPaperQuContent(h.db, b.PaperID)"},
		{"candidate.go", "func (h *CandidateHandler) StandScoreCandidate(", "queryPaperQuContent(h.db, b.PaperID)"},
	}

	for _, tt := range tests {
		t.Run(tt.file+tt.signature, func(t *testing.T) {
			body := extractFunctionBody(t, readSourceFile(t, tt.file), tt.signature)
			guardAt := strings.Index(body, "requireLegacyPaper")
			if tt.signature == "func (h *PaperHandler) CreatePaper(" {
				guardAt = strings.Index(body, "requireLegacyExam")
			}
			workAt := strings.Index(body, tt.legacyWork)
			if guardAt < 0 {
				t.Fatalf("%s must call the legacy assessment guard", tt.signature)
			}
			if workAt < 0 {
				t.Fatalf("legacy work marker %q not found", tt.legacyWork)
			}
			if guardAt > workAt {
				t.Fatalf("legacy assessment guard must run before %q", tt.legacyWork)
			}
		})
	}
}

func TestLegacyPaperGuard_HasStrictDispatchAndControlledErrors(t *testing.T) {
	src := readSourceFile(t, "legacy_paper_guard.go")
	for _, required := range []string{
		"ValidateAssessmentMode",
		"AssessmentTypeCompetency",
		"ScoringModeCompetencyAverage",
		"ErrLegacyPaperNotFound",
		"ErrCompetencyDedicatedAPI",
		"ErrInvalidAssessmentMode",
		`Joins("INNER JOIN el_exam e ON e.id = p.exam_id")`,
		`Where("p.id = ?", paperID)`,
	} {
		if !strings.Contains(src, required) {
			t.Errorf("legacy paper guard missing %q", required)
		}
	}
}

func TestLegacyPaperGuard_WriteEndpointsGuardBeforeTransaction(t *testing.T) {
	src := readSourceFile(t, "paper.go")
	for _, signature := range []string{
		"func (h *PaperHandler) FillAnswer(",
		"func (h *PaperHandler) HandExam(",
	} {
		body := extractFunctionBody(t, src, signature)
		guardAt := strings.Index(body, "requireLegacyPaper")
		transactionAt := strings.Index(body, "Transaction(")
		if guardAt < 0 || transactionAt < 0 || guardAt > transactionAt {
			t.Errorf("%s must reject competency paper before opening a write transaction", signature)
		}
	}
}
