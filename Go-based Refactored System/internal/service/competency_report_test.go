package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestTemporaryCompetencyReportTextSelection_RequiresExactCompleteSet(t *testing.T) {
	rows := []model.CompetencyReportText{
		{ContentVersion: CompetencyTemporaryContentVersion, ContentType: CompetencyReportContentOverall, Audience: CompetencyReportAudienceLeader, LevelCode: CompetencyLevelHigh, Content: "临时领导总体评价"},
		{ContentVersion: CompetencyTemporaryContentVersion, ContentType: CompetencyReportContentDimension, Audience: CompetencyReportAudienceLeader, DimensionID: "d1", LevelCode: CompetencyLevelGood, Content: "临时维度建议"},
		{ContentVersion: "other-v1", ContentType: CompetencyReportContentOverall, Audience: CompetencyReportAudienceLeader, LevelCode: CompetencyLevelHigh, Content: "错误版本总体评价"},
	}
	got, err := BuildCompetencyReportTextSnapshot(CompetencyTemporaryContentVersion, CompetencyReportAudienceLeader, CompetencyLevelHigh, map[string]string{"d1": CompetencyLevelGood}, rows)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "临时领导总体评价") || !strings.Contains(got, "临时维度建议") {
		t.Fatalf("snapshot missing exact text: %s", got)
	}
	if strings.Contains(got, "错误版本总体评价") {
		t.Fatalf("snapshot crossed content versions: %s", got)
	}

	if _, err := BuildCompetencyReportTextSnapshot(CompetencyTemporaryContentVersion, CompetencyReportAudienceFrontlineEmployee, CompetencyLevelHigh, map[string]string{"d1": CompetencyLevelGood}, rows); err == nil {
		t.Fatal("must not fall back across report audiences")
	}
	if _, err := BuildCompetencyReportTextSnapshot(CompetencyTemporaryContentVersion, CompetencyReportAudienceLeader, CompetencyLevelHigh, map[string]string{"d1": CompetencyLevelLow}, rows); err == nil {
		t.Fatal("missing dimension level text must fail")
	}
}

func TestPhase1ReportContentApproval_RequiresDualApprovalAndSources(t *testing.T) {
	now := time.Now()
	valid := model.CompetencyReportContentPackage{
		ContentVersion: CompetencyPhase1ContentVersion, ProductVersion: CompetencyPhase1ProductVersion,
		ScoringVersion: CompetencyPhase1ScoringVersion, TemplateVersion: CompetencyPhase1ReportTemplateVersion,
		Audience: CompetencyReportAudienceFrontlineEmployee, ApprovalStatus: CompetencyReportApprovalApproved,
		ContentApprovedBy: "content-owner", ContentApprovedAt: &now,
		PsychometricApprovedBy: "psychometric-owner", PsychometricApprovedAt: &now,
		QuestionSourceSHA256: strings.Repeat("a", 64), ContentSourceSHA256: strings.Repeat("b", 64),
		EffectiveEnvironment: "production", Disclaimer: "正式报告须结合多方证据谨慎使用。",
	}
	if err := ValidatePhase1ReportContentApproval(valid); err != nil {
		t.Fatalf("valid approval rejected: %v", err)
	}
	for name, mutate := range map[string]func(*model.CompetencyReportContentPackage){
		"wrong product version":      func(row *model.CompetencyReportContentPackage) { row.ProductVersion = "other-v1" },
		"wrong scoring version":      func(row *model.CompetencyReportContentPackage) { row.ScoringVersion = "other-v1" },
		"wrong content version":      func(row *model.CompetencyReportContentPackage) { row.ContentVersion = "other-v1" },
		"wrong template version":     func(row *model.CompetencyReportContentPackage) { row.TemplateVersion = "other-v1" },
		"wrong audience":             func(row *model.CompetencyReportContentPackage) { row.Audience = CompetencyReportAudienceLeader },
		"draft":                      func(row *model.CompetencyReportContentPackage) { row.ApprovalStatus = CompetencyReportApprovalDraft },
		"retired":                    func(row *model.CompetencyReportContentPackage) { row.ApprovalStatus = CompetencyReportApprovalRetired },
		"content approver":           func(row *model.CompetencyReportContentPackage) { row.ContentApprovedBy = "" },
		"content approval time":      func(row *model.CompetencyReportContentPackage) { row.ContentApprovedAt = nil },
		"psychometric approver":      func(row *model.CompetencyReportContentPackage) { row.PsychometricApprovedBy = "" },
		"psychometric approval time": func(row *model.CompetencyReportContentPackage) { row.PsychometricApprovedAt = nil },
		"question hash":              func(row *model.CompetencyReportContentPackage) { row.QuestionSourceSHA256 = "invalid" },
		"content hash":               func(row *model.CompetencyReportContentPackage) { row.ContentSourceSHA256 = "" },
		"environment":                func(row *model.CompetencyReportContentPackage) { row.EffectiveEnvironment = " " },
		"disclaimer":                 func(row *model.CompetencyReportContentPackage) { row.Disclaimer = " " },
	} {
		t.Run(name, func(t *testing.T) {
			row := valid
			mutate(&row)
			if err := ValidatePhase1ReportContentApproval(row); err == nil {
				t.Fatal("incomplete approval accepted")
			}
		})
	}
}

func TestBuildPhase1ReportTextSnapshot_RequiresExactFormalSet(t *testing.T) {
	dimensionLevels := make(map[string]string, 10)
	rows := []model.CompetencyReportText{
		{ContentVersion: CompetencyPhase1ContentVersion, Audience: CompetencyReportAudienceFrontlineEmployee, ContentType: CompetencyReportContentOverall, LevelCode: CompetencyPhase1OverallWeak, Content: "总体正式文案", Disclaimer: "正式免责声明"},
		{ContentVersion: CompetencyPhase1ContentVersion, Audience: CompetencyReportAudienceFrontlineEmployee, ContentType: CompetencyReportContentGroup, DimensionID: CompetencyPhase1GroupGeneralAbility, LevelCode: CompetencyReportGroupDescriptionLevel, Content: "通用能力说明", Disclaimer: "正式免责声明"},
		{ContentVersion: CompetencyPhase1ContentVersion, Audience: CompetencyReportAudienceFrontlineEmployee, ContentType: CompetencyReportContentGroup, DimensionID: CompetencyPhase1GroupPsychologicalQuality, LevelCode: CompetencyReportGroupDescriptionLevel, Content: "心理素养说明", Disclaimer: "正式免责声明"},
		{ContentVersion: CompetencyPhase1ContentVersion, Audience: CompetencyReportAudienceFrontlineEmployee, ContentType: CompetencyReportContentValidity, LevelCode: CompetencyPhase1ValidityGood, Content: "效度良好提示", Disclaimer: "正式免责声明"},
	}
	for index, dimensionID := range NormalizePhase1CompetencyConfiguration().DimensionIDs {
		dimensionLevels[dimensionID] = CompetencyPhase1LevelL3
		rows = append(rows, model.CompetencyReportText{ContentVersion: CompetencyPhase1ContentVersion, Audience: CompetencyReportAudienceFrontlineEmployee, ContentType: CompetencyReportContentDimension, DimensionID: dimensionID, LevelCode: CompetencyPhase1LevelL3, Content: "维度正式文案" + string(rune('A'+index)), Disclaimer: "正式免责声明"})
	}
	snapshot, err := BuildPhase1ReportTextSnapshot(CompetencyPhase1ContentVersion, CompetencyReportAudienceFrontlineEmployee, CompetencyPhase1OverallWeak, CompetencyPhase1ValidityGood, dimensionLevels, rows)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"总体正式文案", "通用能力说明", "心理素养说明", "效度良好提示", "正式免责声明"} {
		if !strings.Contains(snapshot, required) {
			t.Errorf("phase-1 snapshot missing %q", required)
		}
	}
	rows[len(rows)-1].IsTemporary = 1
	if _, err := BuildPhase1ReportTextSnapshot(CompetencyPhase1ContentVersion, CompetencyReportAudienceFrontlineEmployee, CompetencyPhase1OverallWeak, CompetencyPhase1ValidityGood, dimensionLevels, rows); err == nil {
		t.Fatal("temporary phase-1 text accepted")
	}
	rows[len(rows)-1].IsTemporary = 0

	for name, mutate := range map[string]func([]model.CompetencyReportText){
		"overall whitespace": func(items []model.CompetencyReportText) { items[0].Content = " " },
		"general group missing": func(items []model.CompetencyReportText) {
			items[1].Status = 1
		},
		"psychological group whitespace": func(items []model.CompetencyReportText) { items[2].Content = " " },
		"validity whitespace":            func(items []model.CompetencyReportText) { items[3].Content = " " },
		"dimension missing": func(items []model.CompetencyReportText) {
			items[len(items)-1].ContentVersion = "other-v1"
		},
		"disclaimer whitespace": func(items []model.CompetencyReportText) {
			for index := range items {
				items[index].Disclaimer = " "
			}
		},
		"disclaimer mismatch": func(items []model.CompetencyReportText) { items[0].Disclaimer = "冲突免责声明" },
	} {
		t.Run(name, func(t *testing.T) {
			items := append([]model.CompetencyReportText(nil), rows...)
			mutate(items)
			if _, err := BuildPhase1ReportTextSnapshot(CompetencyPhase1ContentVersion, CompetencyReportAudienceFrontlineEmployee, CompetencyPhase1OverallWeak, CompetencyPhase1ValidityGood, dimensionLevels, items); err == nil {
				t.Fatal("incomplete or inconsistent formal text accepted")
			}
		})
	}
}

func TestPhase1ReportFramework_UsesFixedTenPages(t *testing.T) {
	framework := NewPhase1ReportFramework()
	if framework.SchemaVersion != "competency-phase1-report-data-v1" || framework.ReportKind != "frontline_phase1" || framework.Renderable {
		t.Fatalf("framework=%+v", framework)
	}
	if len(framework.Pages) != 10 {
		t.Fatalf("pages=%d want=10", len(framework.Pages))
	}
	wantKinds := []string{"cover", "reading_guide", "person_overall_validity", "group_overview", "dimension_radar", "dimension_details", "dimension_details", "dimension_details", "dimension_details", "dimension_details"}
	for index, page := range framework.Pages {
		if page.Number != index+1 || page.Kind != wantKinds[index] {
			t.Fatalf("page[%d]=%+v", index, page)
		}
	}
}

func TestBuildPhase1FormalReportData_ExposesTwoGroupsTenDimensionsWithoutValidityScore(t *testing.T) {
	now := time.Now()
	overall := decimal.NewFromInt(36)
	level := CompetencyPhase1OverallQualified
	result := model.CompetencyResult{PaperID: "paper-1", ParticipantName: "测试人员", OverallScore: &overall, EvaluationLevel: &level, ReportAudience: CompetencyReportAudienceFrontlineEmployee, IsComplete: 1, ProductVersion: CompetencyPhase1ProductVersion, ScoringVersion: CompetencyPhase1ScoringVersion, ContentVersion: CompetencyPhase1ContentVersion, ReportTemplateVersion: CompetencyPhase1ReportTemplateVersion, SubmittedAt: &now}
	groups := make([]model.CompetencyGroupResult, 0, 2)
	for index, group := range Phase1CompetencyGroups() {
		score := decimal.NewFromInt(3)
		groupLevel := CompetencyPhase1LevelL3
		groups = append(groups, model.CompetencyGroupResult{GroupCode: group.Code, GroupName: group.Name, DisplayOrder: index + 1, GroupScore: &score, LevelCode: &groupLevel, IsComplete: 1})
	}
	dimensions := make([]model.CompetencyDimensionResult, 0, 10)
	for index, id := range NormalizePhase1CompetencyConfiguration().DimensionIDs {
		score := decimal.NewFromInt(3)
		dimensionLevel := CompetencyPhase1LevelL3
		dimensions = append(dimensions, model.CompetencyDimensionResult{DimensionID: id, DimensionCode: id, DimensionName: id, DisplayOrder: index + 1, DimensionScore: &score, LevelCode: &dimensionLevel, IsComplete: 1})
	}
	validityStatus := CompetencyPhase1ValidityGood
	validityScore := decimal.NewFromInt(10)
	validity := model.CompetencyValidityResult{ValidityStatus: &validityStatus, ValidityScore: &validityScore, IsComplete: 1}
	text := Phase1ReportTextSnapshot{ContentVersion: CompetencyPhase1ContentVersion, Audience: CompetencyReportAudienceFrontlineEmployee, Disclaimer: "正式免责声明", OverallText: "总体", GroupTexts: map[string]string{}, DimensionTexts: map[string]string{}, ValidityText: "作答有效"}
	data, err := BuildPhase1FormalReportData(result, groups, dimensions, validity, text)
	if err != nil {
		t.Fatal(err)
	}
	if !data.Renderable || len(data.Pages) != 10 || len(data.Groups) != 2 || len(data.Dimensions) != 10 {
		t.Fatalf("data=%+v", data)
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"validityScore", "evaluationAverage", "threshold"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Errorf("participant DTO leaks %q: %s", forbidden, encoded)
		}
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
			_, err := BuildCompetencyReportTextSnapshot(CompetencyTemporaryContentVersion, test.audience, test.overallLevel, test.dimensionLevels, test.rows)
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
