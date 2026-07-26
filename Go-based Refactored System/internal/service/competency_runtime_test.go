package service

import (
	"encoding/json"
	"errors"
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
	updates := competencyDimensionSnapshotUpdates(master, 8, time.Now())
	for key, want := range map[string]any{
		"dimension_code": "D01", "dimension_name": "新名称", "vird_level": "新层级",
		"applicable_category": "新类别", "core_meaning": "新含义", "display_order": 9,
		"question_count": 8,
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
