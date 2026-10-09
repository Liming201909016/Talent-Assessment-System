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
		"DELETE FROM el_competency_report_current",
		"DELETE FROM el_competency_report WHERE",
		"DELETE FROM el_competency_result_run_validity",
		"DELETE FROM el_competency_result_run_module",
		"DELETE FROM el_competency_result_run_dimension",
		"DELETE FROM el_competency_result_run_overall",
		"DELETE FROM el_competency_result_run WHERE",
		"DELETE FROM el_competency_group_result",
		"DELETE FROM el_competency_validity_result",
		"DELETE FROM el_competency_dimension_result",
		"DELETE FROM el_competency_result WHERE paper_id",
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

// TestBugFB182_CompetencyDeleteSkipsUnappliedVersionTables
// 对应：docs/regression-tests.md FB-182
// 复现：011/013尚未执行时，整链删除无条件DELETE不存在的current/result_run表并导致事务失败。
// 期望：新增版本表存在时按依赖顺序删除；不存在时继续执行完整v1删除链。
func TestBugFB182_CompetencyDeleteSkipsUnappliedVersionTables(t *testing.T) {
	chain := extractFunctionBody(t, readSourceFile(t, "exam.go"), "func deleteCompetencyExamChain(")
	for _, table := range []string{
		"el_competency_report_current",
		"el_competency_result_run_validity",
		"el_competency_result_run_module",
		"el_competency_result_run_dimension",
		"el_competency_result_run_overall",
		"el_competency_result_run",
	} {
		if !strings.Contains(chain, `HasTable("`+table+`")`) {
			t.Errorf("optional version table %s is deleted without an existence guard", table)
		}
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
