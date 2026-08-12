package service

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/talent-assessment/refactored/internal/model"
)

// TestBugFB045_TimeoutRequiresExpiredPaper
// 对应：docs/regression-tests.md FB-045
func TestBugFB045_TimeoutRequiresExpiredPaper(t *testing.T) {
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.Local)
	future := now.Add(time.Minute)
	past := now.Add(-time.Minute)
	if err := validateCompetencyTimeout(&future, now); !errors.Is(err, ErrCompetencyTimeoutNotReached) {
		t.Fatalf("future limit error = %v, want ErrCompetencyTimeoutNotReached", err)
	}
	if err := validateCompetencyTimeout(nil, now); !errors.Is(err, ErrCompetencyDurationRequired) {
		t.Fatalf("nil limit error = %v, want ErrCompetencyDurationRequired", err)
	}
	if err := validateCompetencyTimeout(&past, now); err != nil {
		t.Fatalf("expired paper should allow timeout: %v", err)
	}
}

// TestBugFB046_CompetencyTimingBoundaries
// 对应：docs/regression-tests.md FB-046
func TestBugFB046_CompetencyTimingBoundaries(t *testing.T) {
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.Local)
	beforeEnd := now.Add(time.Minute)
	afterEnd := now.Add(-time.Minute)
	tests := []struct {
		name string
		exam model.Exam
		want error
	}{
		{"positive duration before end", model.Exam{TotalTime: 30, EndTime: &beforeEnd}, nil},
		{"zero duration", model.Exam{TotalTime: 0, EndTime: &beforeEnd}, ErrCompetencyDurationRequired},
		{"negative duration", model.Exam{TotalTime: -1, EndTime: &beforeEnd}, ErrCompetencyDurationRequired},
		{"after exam end", model.Exam{TotalTime: 30, EndTime: &afterEnd}, ErrCompetencyExamEnded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCompetencyStart(&tt.exam, now)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

// TestBugFB047_IncompleteResultCannotGenerateFormalReport
// 对应：docs/regression-tests.md FB-047
func TestBugFB047_IncompleteResultCannotGenerateFormalReport(t *testing.T) {
	if err := validateCompetencyFormalReport(model.CompetencyResult{IsComplete: 0}); !errors.Is(err, ErrCompetencyIncompleteReport) {
		t.Fatalf("incomplete result error = %v, want ErrCompetencyIncompleteReport", err)
	}
	if err := validateCompetencyFormalReport(model.CompetencyResult{IsComplete: 1}); err != nil {
		t.Fatalf("complete result should allow report: %v", err)
	}
}

// TestBugFB049_PublishRefreshesDimensionSnapshot
// 对应：docs/regression-tests.md FB-049
func TestBugFB049_PublishRefreshesDimensionSnapshot(t *testing.T) {
	master := model.CompetencyDimension{
		ID: "d1", Code: "D01", Name: "新名称", VIRDLevel: "新层级",
		ApplicableCategory: "新类别", CoreMeaning: "新含义", DisplayOrder: 9,
	}
	updates := competencyDimensionSnapshotUpdates(master, "group-1", 8, time.Now())
	for key, want := range map[string]any{
		"dimension_code": "D01", "dimension_name": "新名称", "vird_level": "新层级",
		"applicable_category": "新类别", "core_meaning": "新含义", "display_order": 9,
		"question_count": 8,
		"group_id":       "group-1",
	} {
		if updates[key] != want {
			t.Errorf("updates[%s] = %v, want %v", key, updates[key], want)
		}
	}
	if updates["snapshot_time"] == nil {
		t.Fatal("snapshot_time must be frozen at publish")
	}
}

func TestValidateCompetencyResultSort(t *testing.T) {
	tests := []struct {
		name, sortBy, direction, dimensionID string
		wantOrder, wantJoin                  string
		wantErr                              bool
	}{
		{"default", "", "", "", "r.submitted_at DESC, r.paper_id DESC", "", false},
		{"submitted ascending", "submittedAt", "asc", "", "r.submitted_at ASC, r.paper_id ASC", "", false},
		{"overall descending", "overallScore", "desc", "", "r.overall_score DESC, r.paper_id DESC", "", false},
		{"dimension ascending", "dimensionScore", "asc", "d1", "dr.dimension_score IS NULL ASC, dr.dimension_score ASC, r.paper_id ASC", "d1", false},
		{"dimension missing id", "dimensionScore", "desc", "", "", "", true},
		{"unknown field", "drop table", "asc", "", "", "", true},
		{"unknown direction", "overallScore", "sideways", "", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateCompetencyResultSort(tt.sortBy, tt.direction, tt.dimensionID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got.OrderClause != tt.wantOrder || got.DimensionID != tt.wantJoin {
				t.Fatalf("sort = %+v, want order=%q dimension=%q", got, tt.wantOrder, tt.wantJoin)
			}
		})
	}
}

func TestCompetencyRankingDefaults_ExcludeIncompleteAndQuestionable(t *testing.T) {
	for _, tt := range []struct {
		name, sortBy, completion, validity string
		wantCompletion, wantValidity       string
	}{
		{"management chronology stays all", "submittedAt", "", "", "", ""},
		{"overall ranking defaults eligible", "overallScore", "", "", "complete", "good"},
		{"dimension ranking defaults eligible", "dimensionScore", "", "", "complete", "good"},
		{"explicit all remains all", "overallScore", "all", "all", "all", "all"},
		{"questionable defaults complete", "overallScore", "", "questionable", "complete", "questionable"},
		{"incomplete defaults good", "overallScore", "incomplete", "", "incomplete", "good"},
		{"explicit questionable remains visible", "overallScore", "complete", "questionable", "complete", "questionable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			gotCompletion, gotValidity := competencyRankingDefaults(tt.sortBy, tt.completion, tt.validity)
			if gotCompletion != tt.wantCompletion || gotValidity != tt.wantValidity {
				t.Fatalf("defaults=(%q,%q), want=(%q,%q)", gotCompletion, gotValidity, tt.wantCompletion, tt.wantValidity)
			}
		})
	}
}

func TestValidatePhase1ValidityQuestionOrder(t *testing.T) {
	questionType := model.CompetencyQuestionTypeValidity
	questions := make([]model.Qu, 10)
	for index := range questions {
		itemNo := index + 1
		questions[index] = model.Qu{CompetencyQuestionType: &questionType, DimensionItemNo: &itemNo}
	}
	if err := validatePhase1ValidityQuestionOrder(questions); err != nil {
		t.Fatalf("valid order rejected: %v", err)
	}
	duplicate := 1
	questions[9].DimensionItemNo = &duplicate
	if err := validatePhase1ValidityQuestionOrder(questions); err == nil {
		t.Fatal("duplicate/missing validity order accepted")
	}
}

func TestCompetencyResultPaging_RequiresExamContextBeforeQuery(t *testing.T) {
	service := &CompetencyRuntimeService{}
	rows, total, err := service.ResultPaging(CompetencyResultPageRequest{SortBy: "submittedAt"})
	if err == nil || err.Error() != "examId 为空" {
		t.Fatalf("error = %v, want examId 为空", err)
	}
	if rows != nil || total != 0 {
		t.Fatalf("rows=%v total=%d, want nil/0", rows, total)
	}
}

func TestCompetencyResultFilters_NormalizeAndValidate(t *testing.T) {
	tests := []struct {
		name, inputName, telephone, completion, validity string
		wantName, wantTelephone                          string
		wantComplete                                     *int
		wantValidity                                     string
		wantErr                                          bool
	}{
		{"empty", "", "", "", "", "", "", nil, "", false},
		{"trimmed complete and good", "  张三 ", " 139 ", "complete", " good ", "张三", "139", intPointer(1), "good", false},
		{"incomplete result and questionable validity", "", "", "incomplete", "questionable", "", "", intPointer(0), "questionable", false},
		{"explicit incomplete validity", "", "", "all", "incomplete", "", "", nil, "incomplete", false},
		{"all", "", "", "all", "all", "", "", nil, "", false},
		{"invalid completion", "", "", "finished", "", "", "", nil, "", true},
		{"invalid validity", "", "", "", "trusted", "", "", nil, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeCompetencyResultFilters(tt.inputName, tt.telephone, tt.completion, tt.validity)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, tt.wantErr)
			}
			if got.Name != tt.wantName || got.Telephone != tt.wantTelephone {
				t.Fatalf("filters=%+v", got)
			}
			if (got.IsComplete == nil) != (tt.wantComplete == nil) || got.IsComplete != nil && *got.IsComplete != *tt.wantComplete {
				t.Fatalf("isComplete=%v want=%v", got.IsComplete, tt.wantComplete)
			}
			if got.ValidityStatus != tt.wantValidity {
				t.Fatalf("validityStatus=%q want=%q", got.ValidityStatus, tt.wantValidity)
			}
		})
	}
}

func intPointer(value int) *int { return &value }

func TestCompetencyOptionsJSON_ForwardAndReverse(t *testing.T) {
	for _, tt := range []struct {
		direction string
		want      []int
	}{
		{CompetencyDirectionForward, []int{1, 2, 3, 4, 5}},
		{CompetencyDirectionReverse, []int{5, 4, 3, 2, 1}},
	} {
		raw, err := competencyOptionsJSON(tt.direction)
		if err != nil {
			t.Fatal(err)
		}
		var options []struct {
			RawValue   int `json:"rawValue"`
			FinalScore int `json:"finalScore"`
		}
		if err := json.Unmarshal([]byte(raw), &options); err != nil {
			t.Fatal(err)
		}
		if len(options) != 5 {
			t.Fatalf("options=%d", len(options))
		}
		for i, option := range options {
			if option.RawValue != i+1 || option.FinalScore != tt.want[i] {
				t.Fatalf("direction=%s option=%+v", tt.direction, option)
			}
		}
	}
}

func TestParticipantTable(t *testing.T) {
	if participantTable(CompetencyParticipantTester) != "el_tester" {
		t.Fatal("tester table")
	}
	if participantTable(CompetencyParticipantCandidate) != "el_candidate" {
		t.Fatal("candidate table")
	}
	if participantTable("admin") != "" {
		t.Fatal("invalid participant table")
	}
}

func TestCompetencyResultPaging_ProjectsStartAndDurationForManagement(t *testing.T) {
	source, err := os.ReadFile("competency_runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		`StartedAt            *time.Time`,
		`json:"startedAt"`,
		`p.create_time AS started_at`,
		`p.user_time`,
	} {
		if !strings.Contains(text, required) {
			t.Errorf("result paging management projection missing %q", required)
		}
	}
}

func TestCompetencyFormalReportData_ProjectsTemplateMetadata(t *testing.T) {
	source, err := os.ReadFile("competency_runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		`"examTitle"`,
		`"requiredFields"`,
		`"startedAt"`,
		`"userTime"`,
		`"generatedAt"`,
		`"dimensionCoreMeanings"`,
		`core_meaning`,
	} {
		if !strings.Contains(text, required) {
			t.Errorf("formal report metadata missing %q", required)
		}
	}
}

// TestBugFB105_ApprovedPhase1ReportBuildsFormalData
// 对应：docs/regression-tests.md #FB-105
// 复现：一期内容包完成双批准后，FormalReportData仍固定返回“内容尚未导入”。
// 期望：读取精确版本文案并调用一期快照与强类型DTO构建器，不保留固定占位错误。
func TestBugFB105_ApprovedPhase1ReportBuildsFormalData(t *testing.T) {
	source, err := os.ReadFile("competency_runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	if strings.Contains(text, `return nil, errors.New("一期正式报告内容尚未导入")`) {
		t.Fatal("approved phase-1 report path still ends in a permanent placeholder error")
	}
	for _, required := range []string{
		`BuildPhase1ReportTextSnapshot(`,
		`BuildPhase1FormalReportData(`,
		`CompetencyReportContentGroup`,
		`CompetencyReportContentValidity`,
	} {
		if !strings.Contains(text, required) {
			t.Errorf("approved phase-1 report path missing %q", required)
		}
	}
}

func TestPhase1FormalReportData_IncludesCustomerTemplateMetadata(t *testing.T) {
	source, err := os.ReadFile("competency_runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	start := strings.Index(text, "func (s *CompetencyRuntimeService) phase1FormalReportData")
	end := strings.Index(text[start:], "type CompetencyResultPageRequest")
	if start < 0 || end < 0 {
		t.Fatal("phase-1 formal report function not found")
	}
	phase1 := text[start : start+end]
	for _, required := range []string{`"meta"`, `"examTitle"`, `"requiredFields"`, `"startedAt"`, `"userTime"`, `"generatedAt"`} {
		if !strings.Contains(phase1, required) {
			t.Errorf("phase-1 formal report metadata missing %q", required)
		}
	}
}

func TestCompetencyPhase1Runtime_PublishesAndPersistsCompleteResultSet(t *testing.T) {
	source, err := os.ReadFile("competency_runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		`CompetencyQuestionType: question.CompetencyQuestionType`,
		`CreateInBatches(groupSnapshots`,
		`competencyDimensionSnapshotUpdates(master, groupID`,
		`"group_id":`,
		`model.CompetencyQuestionTypeValidity`,
		`CalculatePhase1CompetencyResult(dimensionInputs)`,
		`CalculatePhase1GroupResults(calculated.Dimensions)`,
		`CalculatePhase1ValidityResult(validityInputs)`,
		`CreateInBatches(groupResults`,
		`Create(&validityResult)`,
		`DimensionQuestionCount: calculated.TotalQuestionCount`,
		`AnsweredDimensionQuestionCount: calculated.AnsweredQuestionCount`,
		`TotalQuestionCount: len(rows)`,
	} {
		if !strings.Contains(text, required) {
			t.Errorf("phase-1 runtime integration missing %q", required)
		}
	}
}

func TestCompetencyPhase1Options_UseConfirmedLabels(t *testing.T) {
	data, err := competencyOptionsJSON(CompetencyDirectionForward)
	if err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"完全不符合", "比较不符合", "不确定", "比较符合", "完全符合"} {
		if !strings.Contains(data, label) {
			t.Errorf("phase-1 option snapshot missing %q: %s", label, data)
		}
	}
	for _, retired := range []string{"非常不符合", "不太符合", "一般", "非常符合"} {
		if strings.Contains(data, retired) {
			t.Errorf("phase-1 option snapshot retains old label %q", retired)
		}
	}
}
