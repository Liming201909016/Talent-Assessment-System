package handler

import (
	"strings"
	"testing"
)

// TestBugFB048_CompetencyDeleteUsesFullChainTransaction
// 对应：docs/regression-tests.md FB-048
func TestBugFB048_CompetencyDeleteUsesFullChainTransaction(t *testing.T) {
	src := readSourceFile(t, "exam.go")
	deleteBody := extractFunctionBody(t, src, "func (h *ExamHandler) Delete(")
	if !strings.Contains(deleteBody, "deleteCompetencyExamChain") {
		t.Fatal("Exam.Delete must dispatch competency exams to full-chain deletion")
	}
	chain := extractFunctionBody(t, src, "func deleteCompetencyExamChain(")
	requiredInOrder := []string{
		"DELETE FROM el_competency_report_audit",
		"DELETE FROM el_competency_report WHERE",
		"DELETE FROM el_competency_group_result",
		"DELETE FROM el_competency_validity_result",
		"DELETE FROM el_competency_dimension_result",
		"DELETE FROM el_competency_result",
		"DELETE FROM el_paper_qu_answer",
		"DELETE FROM el_paper_qu WHERE",
		"DELETE FROM el_mbti_answer",
		"DELETE FROM el_user_exam",
		"DELETE FROM el_candidate",
		"DELETE FROM el_tester",
		`Delete(&model.Paper{})`,
		"DELETE FROM el_exam_competency_question",
		`Delete(&model.ExamCompetencyDimension{})`,
		`Delete(&model.ExamCompetencyGroup{})`,
		`Delete(&model.ExamRepo{})`,
		`Delete(&model.ExamDepart{})`,
		`Delete(&model.Exam{})`,
	}
	last := -1
	for _, fragment := range requiredInOrder {
		position := strings.Index(chain, fragment)
		if position < 0 {
			t.Errorf("competency delete chain missing %q", fragment)
			continue
		}
		if position <= last {
			t.Errorf("competency delete fragment %q is out of dependency order", fragment)
		}
		last = position
	}
	if strings.Contains(chain, "Transaction(") {
		t.Error("deleteCompetencyExamChain must use the transaction passed by Exam.Delete, not open a nested transaction")
	}
}

func TestBugFB048_LegacyDeleteStillRejectsRelations(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "exam.go"), "func (h *ExamHandler) Delete(")
	for _, required := range []string{"legacyIDs", "testerCount", "candidateCount", "paperCount", "无法删除"} {
		if !strings.Contains(body, required) {
			t.Errorf("legacy delete relation guard missing %q", required)
		}
	}
}

// TestBugFB050_DirectCompetencyDeletesAreRejected
// 对应：docs/regression-tests.md FB-050
func TestBugFB050_DirectCompetencyDeletesAreRejected(t *testing.T) {
	tests := []struct {
		file, signature string
	}{
		{"paper.go", "func (h *PaperHandler) Delete("},
		{"candidate.go", "func (h *CandidateHandler) Remove("},
		{"candidate.go", "func (h *CandidateHandler) Logistic("},
		{"tester.go", "func (h *TesterHandler) Remove("},
		{"tester.go", "func (h *TesterHandler) Logistic("},
	}
	for _, tt := range tests {
		t.Run(tt.file+tt.signature, func(t *testing.T) {
			body := extractFunctionBody(t, readSourceFile(t, tt.file), tt.signature)
			if !strings.Contains(body, "rejectDirectCompetencyDelete") {
				t.Errorf("%s must reject direct competency deletion", tt.signature)
			}
		})
	}
	helper := readSourceFile(t, "competency_delete_guard.go")
	for _, required := range []string{"assessment_type", "competency", "请删除所属胜任力测评"} {
		if !strings.Contains(helper, required) {
			t.Errorf("direct delete guard missing %q", required)
		}
	}
}
