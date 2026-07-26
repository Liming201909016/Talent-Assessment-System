package handler

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestLoadEnabledQuestionCounts_PropagatesDatabaseFailure(t *testing.T) {
	sqlDB, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:1)/element?timeout=1ms")
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := loadEnabledQuestionCounts(db, []string{"d1"}); err == nil {
		t.Fatal("closed database must propagate enabled-question count query failure")
	}
}

func TestValidateEnabledQuestionCounts(t *testing.T) {
	dimensions := []model.CompetencyDimension{
		{ID: "d1", Code: "D01", Name: "沟通表达"},
		{ID: "d2", Code: "D02", Name: "人际交往"},
	}
	tests := []struct {
		name   string
		counts map[string]int
		want   string
	}{
		{"all dimensions have questions", map[string]int{"d1": 8, "d2": 6}, ""},
		{"first dimension empty", map[string]int{"d2": 6}, "D01 沟通表达没有启用题目"},
		{"second dimension empty", map[string]int{"d1": 8, "d2": 0}, "D02 人际交往没有启用题目"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnabledQuestionCounts(dimensions, tt.counts)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.want {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSameStringSet(t *testing.T) {
	tests := []struct {
		name        string
		left, right []string
		want        bool
	}{
		{"same order", []string{"d1", "d2"}, []string{"d1", "d2"}, true},
		{"different order", []string{"d1", "d2"}, []string{"d2", "d1"}, true},
		{"different values", []string{"d1", "d2"}, []string{"d1", "d3"}, false},
		{"different length", []string{"d1"}, []string{"d1", "d2"}, false},
		{"duplicate mismatch", []string{"d1", "d1"}, []string{"d1", "d2"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameStringSet(tt.left, tt.right); got != tt.want {
				t.Fatalf("sameStringSet(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestExamSave_CompetencyConfigurationGuards(t *testing.T) {
	src := readSourceFile(t, "exam.go")
	body := extractFunctionBody(t, src, "func (h *ExamHandler) Save(")

	for _, required := range []string{
		"AssessmentType",
		"`json:\"assessmentType\"`",
		"CompetencyReportAudience",
		"`json:\"competencyReportAudience\"`",
		"DimensionIDs",
		"`json:\"dimensionIds\"`",
		"ValidateAssessmentMode",
		"ValidateCompetencyReportAudience",
		"ValidateCompetencyDimensionIDs",
		"胜任力测评必须配置大于0的答题时长",
		`Where("id IN ? AND status = ?", body.DimensionIDs, 0)`,
		"所选测评维度不存在或已停用",
		"loadEnabledQuestionCounts",
		"validateEnabledQuestionCounts",
		"QuestionCount:      enabledQuestionCounts[dimension.ID]",
		"已发布胜任力测评不能修改报告版本",
		"已发布胜任力测评不能修改测评维度",
		"CreateInBatches(associations, 100)",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("Exam.Save competency guard missing %q", required)
		}
	}
}

func TestCompetencyDimensionList_ReturnsQuestionCountsWithoutNPlusOne(t *testing.T) {
	src := readSourceFile(t, "competency_dimension.go")
	body := extractFunctionBody(t, src, "func (h *CompetencyDimensionHandler) List(")
	for _, required := range []string{
		"loadEnabledQuestionCounts",
		"QuestionCount",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("dimension list missing %q", required)
		}
	}
	if strings.Contains(body, "for _, dimension := range rows {\n\t\th.db") {
		t.Error("dimension list must not query question counts inside a loop")
	}
}

func TestExamDetail_ReturnsCompetencyConfiguration(t *testing.T) {
	src := readSourceFile(t, "exam.go")
	body := extractFunctionBody(t, src, "func (h *ExamHandler) Detail(")
	for _, required := range []string{
		`"assessmentType"`,
		`"competencyReportAudience"`,
		`"publishStatus"`,
		`"dimensionIds"`,
		`"competencyDimensions"`,
		"ExamCompetencyDimension",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("Exam.Detail competency response missing %q", required)
		}
	}
}

// TestFeedbackUF001_CompetencyQuestionBank00401IsReachable
// 对应：docs/user-feedback-log.md UF-001
func TestFeedbackUF001_CompetencyQuestionBank00401IsReachable(t *testing.T) {
	repoSource := readSourceFile(t, "qu_repo.go")
	paging := extractFunctionBody(t, repoSource, "func (h *RepoHandler) Paging(")
	for _, required := range []string{"competencyVirtualRepo", "countCompetencyQuestions"} {
		if !strings.Contains(paging, required) {
			t.Errorf("repository paging missing 00401 entry %q", required)
		}
	}
	for _, required := range []string{`Code: "00401"`, `Title: "胜任力测验题库"`, `Virtual: true`} {
		if !strings.Contains(repoSource, required) {
			t.Errorf("virtual repository definition missing %q", required)
		}
	}

	dimensionSource := readSourceFile(t, "competency_dimension.go")
	for _, required := range []string{"func (h *CompetencyDimensionHandler) QuestionPaging(", "dimension_id IS NOT NULL", "CompetencyQuestionPageRow"} {
		if !strings.Contains(dimensionSource, required) {
			t.Errorf("competency question paging missing %q", required)
		}
	}
	router := readSourceFile(t, "../router/router.go")
	if !strings.Contains(router, `POST("/paging", competencyDimensionH.QuestionPaging)`) {
		t.Error("competency question paging route missing")
	}
	frontendRouter := readSourceFile(t, "../../ruoyi-ui/src/router/index.js")
	frontendRepo := readSourceFile(t, "../../ruoyi-ui/src/views/qu/repo/index.vue")
	for _, required := range []string{"CompetencyQuestionList", "competency/questions"} {
		if !strings.Contains(frontendRouter, required) {
			t.Errorf("frontend competency question route missing %q", required)
		}
	}
	for _, required := range []string{"isCompetencyRepo", "handleQuestionManagement", "CompetencyQuestionList"} {
		if !strings.Contains(frontendRepo, required) {
			t.Errorf("repository page missing 00401 navigation %q", required)
		}
	}
}

func TestCompetencyQuestionUpdate_UsesEditableFieldWhitelist(t *testing.T) {
	source := readSourceFile(t, "competency_dimension.go")
	for _, required := range []string{
		"type CompetencyQuestionUpdateRequest struct",
		"func normalizeCompetencyQuestionStatus(",
		"func validateCompetencyQuestionUpdate(",
		"func (h *CompetencyDimensionHandler) QuestionUpdate(",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("competency question update missing %q", required)
		}
	}
	body := extractFunctionBody(t, source, "func (h *CompetencyDimensionHandler) QuestionUpdate(")
	for _, required := range []string{
		`Where("id = ? AND dimension_id IS NOT NULL", request.ID)`,
		`"content":`, `"observation_point":`, `"scoring_direction":`,
		`"question_status":`, `"remark":`, `"update_time":`,
		"Updates(updates)", "RowsAffected",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("QuestionUpdate missing %q", required)
		}
	}
	for _, forbidden := range []string{"Save(", `"question_code":`, `"dimension_id":`, `"dimension_item_no":`, `"create_time":`} {
		if strings.Contains(body, forbidden) {
			t.Errorf("QuestionUpdate must not write immutable field %q", forbidden)
		}
	}
	router := readSourceFile(t, "../router/router.go")
	if !strings.Contains(router, `POST("/update", competencyDimensionH.QuestionUpdate)`) {
		t.Error("competency question update route missing")
	}
}

func TestValidateCompetencyQuestionUpdate(t *testing.T) {
	tests := []struct {
		name    string
		request CompetencyQuestionUpdateRequest
		wantErr bool
	}{
		{"valid forward enabled", CompetencyQuestionUpdateRequest{ID: "q1", Content: "题目", ObservationPoint: "考察点", ScoringDirection: "forward", QuestionStatus: 0}, false},
		{"valid reverse disabled", CompetencyQuestionUpdateRequest{ID: "q1", Content: "题目", ObservationPoint: "考察点", ScoringDirection: "reverse", QuestionStatus: 1}, false},
		{"missing id", CompetencyQuestionUpdateRequest{Content: "题目", ObservationPoint: "考察点", ScoringDirection: "forward", QuestionStatus: 0}, true},
		{"blank content", CompetencyQuestionUpdateRequest{ID: "q1", ObservationPoint: "考察点", ScoringDirection: "forward", QuestionStatus: 0}, true},
		{"blank observation", CompetencyQuestionUpdateRequest{ID: "q1", Content: "题目", ScoringDirection: "forward", QuestionStatus: 0}, true},
		{"invalid direction", CompetencyQuestionUpdateRequest{ID: "q1", Content: "题目", ObservationPoint: "考察点", ScoringDirection: "sideways", QuestionStatus: 0}, true},
		{"invalid status", CompetencyQuestionUpdateRequest{ID: "q1", Content: "题目", ObservationPoint: "考察点", ScoringDirection: "forward", QuestionStatus: 2}, true},
		{"empty string status", CompetencyQuestionUpdateRequest{ID: "q1", Content: "题目", ObservationPoint: "考察点", ScoringDirection: "forward", QuestionStatus: ""}, true},
		{"fractional status", CompetencyQuestionUpdateRequest{ID: "q1", Content: "题目", ObservationPoint: "考察点", ScoringDirection: "forward", QuestionStatus: 0.5}, true},
		{"nil status", CompetencyQuestionUpdateRequest{ID: "q1", Content: "题目", ObservationPoint: "考察点", ScoringDirection: "forward", QuestionStatus: nil}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCompetencyQuestionUpdate(&test.request)
			if (err != nil) != test.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, test.wantErr)
			}
		})
	}
}

func TestValidateCompetencyDimensionUpdate(t *testing.T) {
	tests := []struct {
		name    string
		request CompetencyDimensionUpdateRequest
		wantErr bool
	}{
		{"valid enabled", CompetencyDimensionUpdateRequest{ID: "d1", Name: "沟通表达", VIRDLevel: "Versatility 胜任力", ApplicableCategory: "基层通用", CoreMeaning: "清晰传递信息", DisplayOrder: 1, Status: 0}, false},
		{"valid disabled string values", CompetencyDimensionUpdateRequest{ID: "d1", Name: "沟通表达", VIRDLevel: "Versatility 胜任力", ApplicableCategory: "基层通用", CoreMeaning: "清晰传递信息", DisplayOrder: "1", Status: "1"}, false},
		{"missing id", CompetencyDimensionUpdateRequest{Name: "沟通表达", VIRDLevel: "V", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 1, Status: 0}, true},
		{"blank name", CompetencyDimensionUpdateRequest{ID: "d1", VIRDLevel: "V", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 1, Status: 0}, true},
		{"blank vird", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 1, Status: 0}, true},
		{"blank category", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "V", CoreMeaning: "含义", DisplayOrder: 1, Status: 0}, true},
		{"blank meaning", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "V", ApplicableCategory: "基层通用", DisplayOrder: 1, Status: 0}, true},
		{"invalid vird", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "Other", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 1, Status: 0}, true},
		{"invalid category", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "Versatility 胜任力", ApplicableCategory: "其他", CoreMeaning: "含义", DisplayOrder: 1, Status: 0}, true},
		{"zero order", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "V", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 0, Status: 0}, true},
		{"order above 48", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "V", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 49, Status: 0}, true},
		{"fractional order", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "V", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 1.5, Status: 0}, true},
		{"empty status", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "V", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 1, Status: ""}, true},
		{"invalid status", CompetencyDimensionUpdateRequest{ID: "d1", Name: "名称", VIRDLevel: "V", ApplicableCategory: "基层通用", CoreMeaning: "含义", DisplayOrder: 1, Status: 2}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCompetencyDimensionUpdate(&test.request)
			if (err != nil) != test.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, test.wantErr)
			}
		})
	}
}

func TestCompetencyDimensionUpdate_UsesEditableFieldWhitelist(t *testing.T) {
	source := readSourceFile(t, "competency_dimension.go")
	body := extractFunctionBody(t, source, "func (h *CompetencyDimensionHandler) Update(")
	for _, required := range []string{
		`Where("id = ?", request.ID)`, `"name":`, `"vird_level":`, `"applicable_category":`,
		`"core_meaning":`, `"display_order":`, `"status":`, `"update_time":`,
		"Take(&existing)", "Updates(updates)", "胜任力维度不存在",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("dimension update missing %q", required)
		}
	}
	for _, forbidden := range []string{"Save(", `"code":`, `"id":`, `"create_time":`} {
		if strings.Contains(body, forbidden) {
			t.Errorf("dimension update must not write identity/history field %q", forbidden)
		}
	}
	router := readSourceFile(t, "../router/router.go")
	if !strings.Contains(router, `POST("/update", competencyDimensionH.Update)`) {
		t.Error("competency dimension update route missing")
	}
}

func TestCompetencyDimensionUpdate_AtomicallySwapsOccupiedDisplayOrder(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "competency_dimension.go"), "func (h *CompetencyDimensionHandler) Update(")
	for _, required := range []string{
		"Transaction(", "existing.DisplayOrder != displayOrder", "orderOwner",
		`Update("display_order", -existing.DisplayOrder)`, `Update("display_order", existing.DisplayOrder)`,
	} {
		if !strings.Contains(body, required) {
			t.Errorf("dimension order swap missing %q", required)
		}
	}
}

func TestCompetencyDimensionSeedRerunPreservesAdministratorMaintenance(t *testing.T) {
	sql := readSourceFile(t, "../../../scripts/sql/competency_002_dimensions.sql")
	for _, forbidden := range []string{
		"`name`=VALUES(`name`)", "`vird_level`=VALUES(`vird_level`)",
		"`applicable_category`=VALUES(`applicable_category`)", "`core_meaning`=VALUES(`core_meaning`)",
		"`display_order`=VALUES(`display_order`)", "`status`=VALUES(`status`)",
	} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("dimension seed rerun overwrites administrator maintenance: %s", forbidden)
		}
	}
}

// TestFeedbackUF002_UnpublishedCompetencyCannotEnterStartFlow
// 对应：docs/user-feedback-log.md UF-002
func TestFeedbackUF002_UnpublishedCompetencyCannotEnterStartFlow(t *testing.T) {
	online := extractFunctionBody(t, readSourceFile(t, "exam.go"), "func (h *ExamHandler) OnlinePaging(")
	for _, required := range []string{"AssessmentTypeCompetency", "publish_status = 1"} {
		if !strings.Contains(online, required) {
			t.Errorf("online exam list must hide unpublished competency exams: missing %q", required)
		}
	}
	create := extractFunctionBody(t, readSourceFile(t, "competency_runtime.go"), "func (h *CompetencyRuntimeHandler) CreatePaper(")
	if !strings.Contains(create, "测评尚未发布，请联系管理员先发布并冻结题目") {
		t.Error("CreatePaper must return actionable Chinese guidance for an unpublished competency exam")
	}
}
