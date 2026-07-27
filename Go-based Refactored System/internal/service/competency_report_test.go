package service

import (
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/model"
)

func TestTemporaryCompetencyReportTextSelection_RequiresExactCompleteSet(t *testing.T) {
	rows := []model.CompetencyReportText{
		{ContentType: CompetencyReportContentOverall, Audience: CompetencyReportAudienceLeader, LevelCode: CompetencyLevelHigh, Content: "临时领导总体评价"},
		{ContentType: CompetencyReportContentDimension, Audience: CompetencyReportAudienceLeader, DimensionID: "d1", LevelCode: CompetencyLevelGood, Content: "临时维度建议"},
	}
	got, err := BuildCompetencyReportTextSnapshot(CompetencyReportAudienceLeader, CompetencyLevelHigh, map[string]string{"d1": CompetencyLevelGood}, rows)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "临时领导总体评价") || !strings.Contains(got, "临时维度建议") {
		t.Fatalf("snapshot missing exact text: %s", got)
	}

	if _, err := BuildCompetencyReportTextSnapshot(CompetencyReportAudienceFrontlineEmployee, CompetencyLevelHigh, map[string]string{"d1": CompetencyLevelGood}, rows); err == nil {
		t.Fatal("must not fall back across report audiences")
	}
	if _, err := BuildCompetencyReportTextSnapshot(CompetencyReportAudienceLeader, CompetencyLevelHigh, map[string]string{"d1": CompetencyLevelLow}, rows); err == nil {
		t.Fatal("missing dimension level text must fail")
	}
}

func TestTemporaryCompetencyReportVersion_IsExplicitlyNonProduction(t *testing.T) {
	if !strings.Contains(CompetencyTemporaryContentVersion, "temp") {
		t.Fatalf("temporary version must be visibly non-production: %s", CompetencyTemporaryContentVersion)
	}
	if !strings.Contains(CompetencyTemporaryDisclaimer, "不可作为人才决策依据") {
		t.Fatalf("temporary disclaimer is not explicit: %s", CompetencyTemporaryDisclaimer)
	}
}

// TestBugFB074_MissingReportTextIdentifiesExactLookupKey
// 对应：docs/regression-tests.md #FB-074
// 复现：总体或维度文案精确匹配失败时，旧错误只说明文案类型。
// 期望：错误明确包含 contentVersion、audience、dimension 和 level。
func TestBugFB074_MissingReportTextIdentifiesExactLookupKey(t *testing.T) {
	tests := []struct {
		name            string
		audience        string
		overallLevel    string
		dimensionLevels map[string]string
		rows            []model.CompetencyReportText
		want            []string
	}{
		{
			name:     "missing overall text",
			audience: CompetencyReportAudienceFrontlineEmployee, overallLevel: CompetencyLevelHigh,
			dimensionLevels: map[string]string{}, rows: nil,
			want: []string{"contentVersion=temp-v1", "audience=frontline_employee", "dimension=overall", "level=high"},
		},
		{
			name:     "missing dimension text",
			audience: CompetencyReportAudienceLeader, overallLevel: CompetencyLevelHigh,
			dimensionLevels: map[string]string{"competency-d01": CompetencyLevelGood},
			rows: []model.CompetencyReportText{{
				ContentVersion: CompetencyTemporaryContentVersion, ContentType: CompetencyReportContentOverall,
				Audience: CompetencyReportAudienceLeader, LevelCode: CompetencyLevelHigh, Content: "总体评价",
			}},
			want: []string{"contentVersion=temp-v1", "audience=leader", "dimension=competency-d01", "level=good"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := BuildCompetencyReportTextSnapshot(test.audience, test.overallLevel, test.dimensionLevels, test.rows)
			if err == nil {
				t.Fatal("missing report text must fail")
			}
			for _, value := range test.want {
				if !strings.Contains(err.Error(), value) {
					t.Errorf("error %q missing %q", err.Error(), value)
				}
			}
		})
	}
}
