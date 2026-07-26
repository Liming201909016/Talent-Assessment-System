package handler

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================
// FB-019 回归测试：qu.Paging 必须支持 params.repoId（单数）
// 对应：docs/regression-tests.md FB-019
// ============================================================

// TestBugFB019_QuPagingReq_AcceptsParamsRepoId
// 前端发送 {params: {repoId: "xxx"}} 时必须能正确解析
// 历史 bug：之前只支持 params.repoIds（复数数组）和顶层 repoId，遗漏 params.repoId（嵌套单数）
func TestBugFB019_QuPagingReq_AcceptsParamsRepoId(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		wantTopRepoID  string
		wantParRepoID  string
		wantParRepoIds []string
	}{
		{
			name:          "顶层 repoId（旧格式）",
			json:          `{"current":1,"size":10,"repoId":"REPO-A"}`,
			wantTopRepoID: "REPO-A",
		},
		{
			name:          "params.repoId（前端实际发送格式 — FB-019）",
			json:          `{"current":1,"size":10,"params":{"repoId":"REPO-B"}}`,
			wantParRepoID: "REPO-B",
		},
		{
			name:           "params.repoIds 数组格式",
			json:           `{"current":1,"size":10,"params":{"repoIds":["REPO-C"]}}`,
			wantParRepoIds: []string{"REPO-C"},
		},
		{
			name:           "三种共存：优先顶层，其次嵌套单数，最后数组",
			json:           `{"current":1,"size":10,"repoId":"TOP","params":{"repoId":"PAR","repoIds":["ARR"]}}`,
			wantTopRepoID:  "TOP",
			wantParRepoID:  "PAR",
			wantParRepoIds: []string{"ARR"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req quPagingReq
			if err := json.Unmarshal([]byte(tt.json), &req); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if req.RepoID != tt.wantTopRepoID {
				t.Errorf("top RepoID: want %q, got %q", tt.wantTopRepoID, req.RepoID)
			}
			if req.Params.RepoID != tt.wantParRepoID {
				t.Errorf("Params.RepoID: want %q, got %q", tt.wantParRepoID, req.Params.RepoID)
			}
			if len(req.Params.RepoIds) != len(tt.wantParRepoIds) {
				t.Errorf("Params.RepoIds length: want %d, got %d", len(tt.wantParRepoIds), len(req.Params.RepoIds))
			}
		})
	}
}

// TestBugFB019_RepoIDResolution
// 验证业务层选择 repoID 的优先级：顶层 > 嵌套单数 > 嵌套数组首项
// 模拟 handler 中的真实逻辑
func TestBugFB019_RepoIDResolution(t *testing.T) {
	resolve := func(req *quPagingReq) string {
		repoID := req.RepoID
		if repoID == "" {
			repoID = req.Params.RepoID
		}
		if repoID == "" && len(req.Params.RepoIds) > 0 {
			repoID = req.Params.RepoIds[0]
		}
		return repoID
	}

	tests := []struct {
		name     string
		req      quPagingReq
		expected string
	}{
		{"全空", quPagingReq{}, ""},
		{"仅顶层", quPagingReq{RepoID: "TOP"}, "TOP"},
		{"仅嵌套单数", func() quPagingReq { var r quPagingReq; r.Params.RepoID = "P"; return r }(), "P"},
		{"仅嵌套数组", func() quPagingReq { var r quPagingReq; r.Params.RepoIds = []string{"A1", "A2"}; return r }(), "A1"},
		{"顶层 + 嵌套单数（顶层优先）", func() quPagingReq { r := quPagingReq{RepoID: "TOP"}; r.Params.RepoID = "P"; return r }(), "TOP"},
		{"嵌套单数 + 嵌套数组（单数优先）", func() quPagingReq {
			var r quPagingReq
			r.Params.RepoID = "P"
			r.Params.RepoIds = []string{"A"}
			return r
		}(), "P"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolve(&tt.req); got != tt.expected {
				t.Errorf("want %q, got %q", tt.expected, got)
			}
		})
	}
}

// TestBugFB051_LegacyQuestionAPIsExcludeCompetencyQuestions
// 对应：docs/regression-tests.md FB-051
// 复现：全局传统题目页混入 dimension_id 非空且无 answerList 的胜任力题，页面崩溃；传统写接口还能编辑或删除胜任力源题。
// 期望：传统读接口仅返回 dimension_id IS NULL；所有传统写入口在写入前拒绝胜任力题目 ID/元数据。
func TestBugFB051_LegacyQuestionAPIsExcludeCompetencyQuestions(t *testing.T) {
	quSource := readSourceFile(t, "qu.go")
	checks := []struct {
		signature string
		required  []string
	}{
		{"func (h *QuHandler) Paging(", []string{"q.dimension_id IS NULL"}},
		{"func (h *QuHandler) List(", []string{"dimension_id IS NULL"}},
		{"func (h *QuHandler) Detail(", []string{"dimension_id IS NULL"}},
		{"func (h *QuHandler) Save(", []string{"hasCompetencyQuestionMetadata", "rejectCompetencyQuestionIDs"}},
		{"func (h *QuHandler) Delete(", []string{"rejectCompetencyQuestionIDs"}},
	}
	for _, check := range checks {
		body := extractFunctionBody(t, quSource, check.signature)
		for _, required := range check.required {
			if !strings.Contains(body, required) {
				t.Errorf("%s missing competency isolation %q", check.signature, required)
			}
		}
	}

	repoBatch := extractFunctionBody(t, readSourceFile(t, "qu_repo.go"), "func (h *RepoHandler) BatchAction(")
	if !strings.Contains(repoBatch, "rejectCompetencyQuestionIDs") {
		t.Error("RepoHandler.BatchAction must reject competency source questions before writes")
	}
	exportBody := extractFunctionBody(t, readSourceFile(t, "qu_excel.go"), "func (h *QuHandler) Export(")
	if !strings.Contains(exportBody, "q.dimension_id IS NULL") {
		t.Error("legacy question export must exclude competency source questions")
	}
}

// TestBugFB051_QuestionListDoesNotIndexMissingAnswers
// 对应：docs/regression-tests.md FB-051
func TestBugFB051_QuestionListDoesNotIndexMissingAnswers(t *testing.T) {
	source := readSourceFile(t, "../../ruoyi-ui/src/views/qu/qu/index.vue")
	for _, forbidden := range []string{"answerList[0].content", "answerList[1].content"} {
		if strings.Contains(source, forbidden) {
			t.Errorf("question list still directly indexes missing option: %s", forbidden)
		}
	}
	if !strings.Contains(source, "formatAnswerList") {
		t.Error("question list must render answerList through a nil/empty-safe formatter")
	}
}

// TestBugFB054_QuestionPagingBatchLoadsAnswersAndRepositories
// 对应：docs/regression-tests.md FB-054
// 复现：传统题目分页每行分别查询答案和题库，20行产生约42次数据库查询。
// 期望：主列表后按当前页题目ID各执行一次批量答案/题库查询，组装循环中不访问数据库。
func TestBugFB054_QuestionPagingBatchLoadsAnswersAndRepositories(t *testing.T) {
	source := readSourceFile(t, "qu.go")
	body := extractFunctionBody(t, source, "func (h *QuHandler) Paging(")
	if !strings.Contains(body, "loadQuestionPageRelations") {
		t.Fatal("Paging must batch-load answer and repository relations")
	}
	loopAt := strings.Index(body, "for i, qu := range rows")
	if loopAt < 0 {
		t.Fatal("Paging result assembly loop not found")
	}
	if strings.Contains(body[loopAt:], "h.db") {
		t.Fatal("Paging must not query the database inside the per-question assembly loop")
	}

	helper := extractFunctionBody(t, source, "func loadQuestionPageRelations(")
	for _, required := range []string{
		`Where("qu_id IN ?", questionIDs)`,
		`Where("qr.qu_id IN ?", questionIDs)`,
		`Order("qu_id ASC, id ASC")`,
		`Order("qr.qu_id ASC, qr.sort ASC, qr.id ASC")`,
	} {
		if !strings.Contains(helper, required) {
			t.Errorf("batch relation loader missing %q", required)
		}
	}
}

// TestBugFB055_RepoBatchActionIsAtomicAndChecksErrors
// 对应：docs/regression-tests.md FB-055
// 复现：批量加入/移除在事务外逐题读写，忽略查询、删除、插入、重排和统计刷新错误，失败时留下部分关联和错误统计。
// 期望：整批操作使用一个事务，批量读取题目/旧关联，批量重建关联，并传播每个数据库错误。
func TestBugFB055_RepoBatchActionIsAtomicAndChecksErrors(t *testing.T) {
	source := readSourceFile(t, "qu_repo.go")
	body := extractFunctionBody(t, source, "func (h *RepoHandler) BatchAction(")
	for _, required := range []string{
		"Transaction(",
		"normalizeBatchIDs",
		`Where("id IN ? AND dimension_id IS NULL", b.QuIDs)`,
		`Where("qu_id IN ?", b.QuIDs)`,
		"CreateInBatches",
		"affectedRepoIDs",
		"refreshRepoStat(tx, repoID)",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("BatchAction atomic workflow missing %q", required)
		}
	}
	transactionAt := strings.Index(body, "Transaction(")
	if transactionAt < 0 {
		t.Fatal("BatchAction transaction not found")
	}
	transactionBody := body[transactionAt:]
	for _, forbidden := range []string{"h.db.", "_ = refreshRepoStat", "_ = tx.", "tx.Save("} {
		if strings.Contains(transactionBody, forbidden) {
			t.Errorf("BatchAction transaction contains unchecked/out-of-transaction operation %q", forbidden)
		}
	}
}

// TestBugFB056_QuestionImportIsAtomicAndChecksErrors
// 对应：docs/regression-tests.md FB-056
// 复现：sheet题库在事务外创建，额外题库关联错误被忽略，统计刷新也在提交后忽略错误。
// 期望：题库、题目、答案、全部关联和统计刷新处于同一事务，任一失败整体回滚。
func TestBugFB056_QuestionImportIsAtomicAndChecksErrors(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "qu_excel.go"), "func (h *QuHandler) ImportExcel(")
	for _, required := range []string{
		"Transaction(",
		`tx.Where("title = ?", title)`,
		"tx.Create(&repo)",
		"validateImportRepositoryIDs",
		"refreshRepoStat(tx, repoID)",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("ImportExcel atomic workflow missing %q", required)
		}
	}
	transactionAt := strings.Index(body, "Transaction(")
	if transactionAt < 0 {
		t.Fatal("ImportExcel transaction not found")
	}
	for _, forbidden := range []string{"h.db.Create(&repo)", "_ = tx.Create", "_ = refreshRepoStat"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("ImportExcel contains unchecked/out-of-transaction operation %q", forbidden)
		}
	}
}

// TestBugFB057_QuestionExportBatchLoadsRelations
// 对应：docs/regression-tests.md FB-057
// 复现：导出循环内每题分别查询题库ID和答案，当前855道传统题会产生1700+次关系查询。
// 期望：导出主查询后复用批量关系加载器，写工作簿循环中不访问数据库。
func TestBugFB057_QuestionExportBatchLoadsRelations(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "qu_excel.go"), "func (h *QuHandler) Export(")
	if !strings.Contains(body, "loadQuestionPageRelations") {
		t.Fatal("Export must batch-load question relations")
	}
	loopAt := strings.Index(body, "for i, qu := range qus")
	if loopAt < 0 {
		t.Fatal("Export workbook loop not found")
	}
	if strings.Contains(body[loopAt:], "h.db") {
		t.Fatal("Export must not query the database inside the per-question workbook loop")
	}
	for _, required := range []string{"answersByQuestion[qu.ID]", "repositoriesByQuestion[qu.ID]"} {
		if !strings.Contains(body, required) {
			t.Errorf("Export relation assembly missing %q", required)
		}
	}
}

// TestBugFB058_QuestionDeleteRejectsPaperReferencesAndChecksErrors
// 对应：docs/regression-tests.md FB-058
// 复现：传统题目被历史试卷引用时仍可删除，且答案/题库关联删除错误被忽略，导致历史详情断链或残留孤儿。
// 期望：事务前拒绝任何 el_paper_qu 引用；事务内检查所有读取/删除/统计错误并先删子表后删题目。
func TestBugFB058_QuestionDeleteRejectsPaperReferencesAndChecksErrors(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "qu.go"), "func (h *QuHandler) Delete(")
	for _, required := range []string{
		`Table("el_paper_qu")`,
		`Where("qu_id IN ?", b.IDs)`,
		"无法删除：题目已被试卷引用",
		`Delete(&model.QuAnswer{})`,
		`Delete(&model.QuRepo{})`,
		`Delete(&model.Qu{})`,
	} {
		if !strings.Contains(body, required) {
			t.Errorf("question delete guard missing %q", required)
		}
	}
	answerDelete := strings.Index(body, `Delete(&model.QuAnswer{})`)
	repoDelete := strings.Index(body, `Delete(&model.QuRepo{})`)
	questionDelete := strings.Index(body, `Delete(&model.Qu{})`)
	if answerDelete < 0 || repoDelete < 0 || questionDelete < 0 || answerDelete > questionDelete || repoDelete > questionDelete {
		t.Error("question delete must remove child rows before the question")
	}
	for _, forbidden := range []string{
		`tx.Model(&model.QuRepo{}).Where("qu_id IN ?", b.IDs).
			Distinct("repo_id").Pluck("repo_id", &oldRepoIDs)\n`,
		`tx.Where("qu_id IN ?", b.IDs).Delete(&model.QuAnswer{})\n`,
		`tx.Where("qu_id IN ?", b.IDs).Delete(&model.QuRepo{})\n`,
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("question delete still ignores database error near %q", forbidden)
		}
	}
}

// TestBugFB059_QuestionSaveRejectsPaperReferences
// 对应：docs/regression-tests.md FB-059
// 复现：已进入历史试卷的传统题仍可编辑，Save会替换题目、答案和题库关联，改变历史详情/报告事实。
// 期望：既有题目在写事务前检查 el_paper_qu，存在引用即拒绝；新增题目不受影响。
func TestBugFB059_QuestionSaveRejectsPaperReferences(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "qu.go"), "func (h *QuHandler) Save(")
	for _, required := range []string{
		"rejectPaperReferencedQuestionIDs",
		"题目已被试卷引用，不能修改",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("question save immutability guard missing %q", required)
		}
	}
	guardAt := strings.Index(body, "rejectPaperReferencedQuestionIDs")
	transactionAt := strings.Index(body, "Transaction(")
	if guardAt < 0 || transactionAt < 0 || guardAt > transactionAt {
		t.Error("paper reference guard must run before the question write transaction")
	}
}

// TestBugFB060_RepositoryPagingCapsSizeAndChecksErrors
// 对应：docs/regression-tests.md FB-060
// 复现：题库分页直接接受任意size，且COUNT/列表错误被忽略，可触发超大响应或返回伪成功空数据。
// 期望：复用统一分页上限，并对COUNT和列表查询错误返回受控失败。
func TestBugFB060_RepositoryPagingCapsSizeAndChecksErrors(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "qu_repo.go"), "func (h *RepoHandler) Paging(")
	for _, required := range []string{"capPageSize(req.Size)", "q.Count(&physicalTotal).Error", "Find(&physicalRows).Error", "查询题库总数失败", "查询题库列表失败"} {
		if !strings.Contains(body, required) {
			t.Errorf("repository paging guard missing %q", required)
		}
	}
}

// TestBugFB061_QuestionSaveValidatesRepositoriesAndChecksErrors
// 对应：docs/regression-tests.md FB-061
// 复现：Save接受重复/不存在题库ID，且编辑时忽略原题/旧关联查询与子表删除错误，可产生孤儿或部分替换。
// 期望：题库ID先规范化，在写事务内全部校验存在；编辑链每个查询/删除错误均中止回滚。
func TestBugFB061_QuestionSaveValidatesRepositoriesAndChecksErrors(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "qu.go"), "func (h *QuHandler) Save(")
	for _, required := range []string{
		"body.RepoIDs = normalizeBatchIDs(body.RepoIDs)",
		"validateImportRepositoryIDs(tx, body.RepoIDs)",
		`tx.Select("create_time").Where("id = ?", qu.ID).Take(&orig).Error`,
		`Pluck("repo_id", &oldRepoIDs).Error`,
		`Delete(&model.QuAnswer{}).Error`,
		`Delete(&model.QuRepo{}).Error`,
	} {
		if !strings.Contains(body, required) {
			t.Errorf("question save consistency guard missing %q", required)
		}
	}
	for _, forbidden := range []string{
		`if err := tx.Select("create_time").Where("id = ?", qu.ID).Take(&orig).Error; err == nil`,
		`tx.Where("qu_id = ?", qu.ID).Delete(&model.QuAnswer{})\n`,
		`tx.Where("qu_id = ?", qu.ID).Delete(&model.QuRepo{})\n`,
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("question save still ignores database failure near %q", forbidden)
		}
	}
}

// TestBugFB062_RepositorySaveProtectsComputedFields
// 对应：docs/regression-tests.md FB-062
// 复现：题库保存直接绑定并Save整行，客户端可覆盖create_time和radio/multi/judge统计，且空标题可入库。
// 期望：标题后端必填；新增由服务端初始化；更新仅允许code/title/remark/update_time且不存在时失败。
func TestBugFB062_RepositorySaveProtectsComputedFields(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "qu_repo.go"), "func (h *RepoHandler) Save(")
	for _, required := range []string{
		`strings.TrimSpace(r.Title)`,
		"题库名称不能为空",
		`Updates(map[string]any{`,
		`"code":`,
		`"title":`,
		`"remark":`,
		`"update_time":`,
		"RowsAffected",
		"题库不存在",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("repository save protection missing %q", required)
		}
	}
	if strings.Contains(body, "h.db.Save(&r)") {
		t.Error("repository update must not Save client-bound computed fields")
	}
}

// TestBugFB063_QuestionAndRepositoryReadsCheckErrors
// 对应：docs/regression-tests.md FB-063
// 复现：题库list和题目list/detail忽略答案、关联、题库编码等查询错误，数据库故障时返回伪成功或部分对象。
// 期望：每个读取错误都返回受控失败，列表仍初始化为空数组。
func TestBugFB063_QuestionAndRepositoryReadsCheckErrors(t *testing.T) {
	quSource := readSourceFile(t, "qu.go")
	listBody := extractFunctionBody(t, quSource, "func (h *QuHandler) List(")
	for _, required := range []string{"Find(&rows).Error", "查询题目列表失败"} {
		if !strings.Contains(listBody, required) {
			t.Errorf("question list error handling missing %q", required)
		}
	}
	detailBody := extractFunctionBody(t, quSource, "func (h *QuHandler) Detail(")
	for _, required := range []string{"Find(&answers).Error", `Pluck("repo_id", &repoIDs).Error`, `Pluck("code", &repoCode).Error`, "查询题目答案失败", "查询题目关联失败", "查询题库编码失败"} {
		if !strings.Contains(detailBody, required) {
			t.Errorf("question detail error handling missing %q", required)
		}
	}
	repoList := extractFunctionBody(t, readSourceFile(t, "qu_repo.go"), "func (h *RepoHandler) List(")
	for _, required := range []string{"Find(&rows).Error", "查询题库列表失败"} {
		if !strings.Contains(repoList, required) {
			t.Errorf("repository list error handling missing %q", required)
		}
	}
}
