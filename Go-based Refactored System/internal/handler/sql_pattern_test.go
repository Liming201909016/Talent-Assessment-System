package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================
// 回归测试 — FB-018 / FB-020 SQL 模式守护
// 对应：docs/regression-tests.md FB-018, FB-020
//
// 这些测试通过扫描源代码确保关键 SQL 模式不会被意外回退。
// 不需要数据库，纯静态检查，CI 安全。
// ============================================================

// readSourceFile 读取 handler 目录下指定源文件
func readSourceFile(t *testing.T, name string) string {
	t.Helper()
	// _test.go 与被测文件同目录
	wd, _ := os.Getwd()
	path := filepath.Join(wd, name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", name, err)
	}
	return string(b)
}

// TestBugFB018_MbtiFillAnswer_NoOnDuplicateKeyUpdate
// FB-018: mbti.FillAnswer 不能依赖 ON DUPLICATE KEY UPDATE（DB 可能没唯一索引）
// FB-032: 不能用 Transaction（多个并发 INSERT 会引发 Deadlock）
// 必须用 SELECT → UPDATE 或 INSERT 模式 + ID 用 nextID()
func TestBugFB018_MbtiFillAnswer_NoOnDuplicateKeyUpdate(t *testing.T) {
	src := readSourceFile(t, "mbti.go")

	// 找到 FillAnswer 函数体
	startIdx := strings.Index(src, "func (h *MbtiHandler) FillAnswer(")
	if startIdx < 0 {
		t.Fatal("FillAnswer function not found in mbti.go")
	}
	rest := src[startIdx:]
	endIdx := strings.Index(rest[1:], "\nfunc ")
	if endIdx < 0 {
		endIdx = len(rest)
	}
	body := rest[:endIdx]

	// 守护 1：禁止 ON DUPLICATE KEY UPDATE（FB-018）
	if strings.Contains(body, "ON DUPLICATE KEY UPDATE") {
		t.Error("FB-018 regression: FillAnswer must NOT use ON DUPLICATE KEY UPDATE")
	}

	// 守护 2 (FB-032)：禁止 Transaction（并发 Deadlock）
	if strings.Contains(body, "Transaction(") {
		t.Error("FB-032 regression: FillAnswer must NOT use Transaction (causes MySQL Deadlock)")
	}

	// 守护 3：必须有 INSERT
	if !strings.Contains(body, "INSERT INTO el_mbti_answer") {
		t.Error("FB-018 regression: FillAnswer must have INSERT branch")
	}

	// 守护 4：必须有 UPDATE
	if !strings.Contains(body, "UPDATE el_mbti_answer") {
		t.Error("FB-018 regression: FillAnswer must have UPDATE branch")
	}

	// 守护 5 (FB-032)：ID 必须用 nextID()
	if !strings.Contains(body, "nextID()") {
		t.Error("FB-032 regression: ID must use nextID() (atomic counter), not time.Now().UnixNano()")
	}
}

// TestBugFB020_TesterList_OrderMatchesSelect
// FB-020: tester.List 的 ORDER BY 必须与 SELECT 表达式一致
// 否则会出现"返回字段已 COALESCE 但排序未 COALESCE"的错乱
func TestBugFB020_TesterList_OrderMatchesSelect(t *testing.T) {
	src := readSourceFile(t, "tester.go")

	// 找 List 函数
	startIdx := strings.Index(src, "func (h *TesterHandler) List(")
	if startIdx < 0 {
		t.Fatal("List function not found in tester.go")
	}
	rest := src[startIdx:]
	endIdx := strings.Index(rest[1:], "\nfunc ")
	if endIdx < 0 {
		endIdx = len(rest)
	}
	body := rest[:endIdx]

	// 如果 SELECT 用了 COALESCE create_time，ORDER BY 也必须 COALESCE
	hasCoalesceInSelect := strings.Contains(body, "COALESCE(pa.create_time, t.create_time) AS create_time")
	if hasCoalesceInSelect {
		hasCoalesceInOrder := strings.Contains(body, `Order("COALESCE(pa.create_time, t.create_time)`)
		if !hasCoalesceInOrder {
			t.Error("FB-020 regression: SELECT uses COALESCE(pa.create_time, t.create_time) " +
				"but ORDER BY does not — sort field must match return field")
		}
	}
}

// extractFunctionBody 提取指定函数的代码体（从函数声明到下一个 func）
func extractFunctionBody(t *testing.T, src, fnSig string) string {
	t.Helper()
	startIdx := strings.Index(src, fnSig)
	if startIdx < 0 {
		t.Fatalf("function %q not found", fnSig)
	}
	rest := src[startIdx:]
	endIdx := strings.Index(rest[1:], "\nfunc ")
	if endIdx < 0 {
		return rest
	}
	return rest[:endIdx]
}

// TestBugFB021_ExamDelete_ChecksRelations
// FB-021: exam.Delete 必须检查 tester/candidate/paper 关联避免孤儿数据
func TestBugFB021_ExamDelete_ChecksRelations(t *testing.T) {
	src := readSourceFile(t, "exam.go")
	body := extractFunctionBody(t, src, "func (h *ExamHandler) Delete(")

	// 必须查 3 个关联表
	musts := []string{
		`Table("el_tester")`,
		`Table("el_candidate")`,
		`Table("el_paper")`,
	}
	for _, m := range musts {
		if !strings.Contains(body, m) {
			t.Errorf("FB-021 regression: Delete must check %s for orphan prevention", m)
		}
	}
	// 必须有"无法删除"或类似的拒绝逻辑
	if !strings.Contains(body, "无法删除") {
		t.Error("FB-021 regression: Delete must reject when relations exist (missing rejection message)")
	}
}

// TestBugFB022_PaperFillAnswer_ChecksState
// FB-022: paper.FillAnswer 必须校验 paper.state，已交卷不允许改答案
func TestBugFB022_PaperFillAnswer_ChecksState(t *testing.T) {
	src := readSourceFile(t, "paper.go")
	body := extractFunctionBody(t, src, "func (h *PaperHandler) FillAnswer(")

	// 必须查 paper state
	if !strings.Contains(body, "paperStateING") {
		t.Error("FB-022 regression: FillAnswer must check paperStateING")
	}
	// 必须有拒绝逻辑（已交卷或状态不允许）
	if !strings.Contains(body, "已交卷") && !strings.Contains(body, "状态不允许") {
		t.Error("FB-022 regression: FillAnswer must reject when paper not in ING state")
	}
}

// TestBugFB023_PaperFillAnswer_ChecksQuOwnership
// FB-023: paper.FillAnswer 必须校验 quId 属于该 paperID
func TestBugFB023_PaperFillAnswer_ChecksQuOwnership(t *testing.T) {
	src := readSourceFile(t, "paper.go")
	body := extractFunctionBody(t, src, "func (h *PaperHandler) FillAnswer(")

	// 必须查 el_paper_qu 验证 (paper_id, qu_id) 关系
	if !strings.Contains(body, "el_paper_qu") {
		t.Error("FB-023 regression: FillAnswer must verify quId belongs to paperId via el_paper_qu")
	}
	if !strings.Contains(body, "paper_id = ? AND qu_id = ?") {
		t.Error("FB-023 regression: FillAnswer must use composite WHERE for ownership check")
	}
	if !strings.Contains(body, "题目不属于此试卷") {
		t.Error("FB-023 regression: FillAnswer must reject when quId not in paper")
	}
}

// TestBugFB024_QuRepoRemove_ChecksExamReference
// FB-024: qu_repo.Remove 必须检查是否被 exam 引用
func TestBugFB024_QuRepoRemove_ChecksExamReference(t *testing.T) {
	src := readSourceFile(t, "qu_repo.go")
	body := extractFunctionBody(t, src, "func (h *RepoHandler) Remove(")

	if !strings.Contains(body, "el_exam_repo") {
		t.Error("FB-024 regression: Remove must check el_exam_repo for exam usage")
	}
	if !strings.Contains(body, "被") || !strings.Contains(body, "考试引用") {
		t.Error("FB-024 regression: Remove must reject with 'used by N exams' message")
	}
}

// TestBugFB025_QuRepoRemove_CleansAssociation
// FB-025: qu_repo.Remove 必须清理 el_qu_repo 关联表
func TestBugFB025_QuRepoRemove_CleansAssociation(t *testing.T) {
	src := readSourceFile(t, "qu_repo.go")
	body := extractFunctionBody(t, src, "func (h *RepoHandler) Remove(")

	// 必须用 Transaction 保证原子性
	if !strings.Contains(body, "Transaction(") {
		t.Error("FB-025 regression: Remove must use Transaction for atomic delete")
	}
	// 必须删除 el_qu_repo 关联
	if !strings.Contains(body, `Table("el_qu_repo")`) {
		t.Error("FB-025 regression: Remove must clean el_qu_repo association table")
	}
}
