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
