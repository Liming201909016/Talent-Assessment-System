package service

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/model"
)

// TestBugFB178_Phase1V2ReportDTOFreezesVersionsScoresAndExactTexts
// Corresponds to docs/regression-tests.md FB-178.
func TestBugFB178_Phase1V2ReportDTOFreezesVersionsScoresAndExactTexts(t *testing.T) {
	scores, modules := phase1V2SelectorFixture(t, []int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	rows := phase1V2ReportTextRows(scores, modules, CompetencyPhase1ValidityGood)
	dto, err := BuildPhase1V2ReportData("run-v2-1", scores, modules, CompetencyPhase1ValidityGood, rows)
	if err != nil {
		t.Fatal(err)
	}
	if dto.ResultRunID != "run-v2-1" || dto.Versions != Phase1V2VersionSet() || dto.Audience != CompetencyReportAudienceFrontlineEmployee {
		t.Fatalf("identity=%+v", dto)
	}
	if dto.Overall.Score != "68.13" || dto.Overall.NormScore != "57.75" || dto.Overall.LevelCode != CompetencyPhase1V2LevelQualified || dto.Overall.ComparisonCode != CompetencyPhase1V2ComparisonAboveNorm {
		t.Fatalf("overall=%+v", dto.Overall)
	}
	if got := moduleCodesFromDTO(dto.Modules); !reflect.DeepEqual(got, []string{CompetencyPhase1V2ModuleSelf, CompetencyPhase1V2ModuleTask, CompetencyPhase1V2ModuleInterpersonal}) {
		t.Fatalf("module order=%v", got)
	}
	if dto.Modules[0].Score != "75.00" || dto.Modules[1].Score != "67.50" || dto.Modules[2].Score != "59.38" {
		t.Fatalf("module displays=%+v", dto.Modules)
	}
	if len(dto.Dimensions) != 10 || dto.Dimensions[7].Score != "78.13" || dto.Dimensions[7].PerformanceText != "performance:truth_pragmatism:good" {
		t.Fatalf("dimensions=%+v", dto.Dimensions)
	}
	if got := reportItemKeys(dto.Strengths); !reflect.DeepEqual(got, []string{"digital_application", "truth_pragmatism", "self_discipline"}) {
		t.Fatalf("strengths=%v", got)
	}
	if dto.Strengths[0].RuleText != "performance:digital_application:good" || dto.Developments[0].RuleText != "development:achievement_orientation:qualified" {
		t.Fatalf("category texts=%+v/%+v", dto.Strengths, dto.Developments)
	}
	if got := reportItemKeys(dto.AdviceDimensions); !reflect.DeepEqual(got, []string{"achievement_orientation", "communication"}) || dto.Overall.AdviceText != "advice:qualified" {
		t.Fatalf("advice=%v/%q", got, dto.Overall.AdviceText)
	}
	if dto.Validity.Status != CompetencyPhase1ValidityGood || dto.Validity.DisplayText != "validity:good" || dto.Disclaimer != "v2-disclaimer" {
		t.Fatalf("validity/disclaimer=%+v/%q", dto.Validity, dto.Disclaimer)
	}
	encoded, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"545/8", "625/8", "validityScore", "threshold"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Errorf("participant DTO leaks %q: %s", forbidden, encoded)
		}
	}
}

func TestBugFB178_Phase1V2ReportTextNeverFallsBackOrDuplicates(t *testing.T) {
	scores, modules := phase1V2SelectorFixture(t, []int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	valid := phase1V2ReportTextRows(scores, modules, CompetencyPhase1ValidityGood)
	for name, test := range map[string]struct {
		mutate func([]model.CompetencyReportText) []model.CompetencyReportText
		want   string
	}{
		"wrong version": {func(rows []model.CompetencyReportText) []model.CompetencyReportText {
			rows[0].ContentVersion = CompetencyPhase1ContentVersion
			return rows
		}, "contentVersion=competency-phase1-content-v2"},
		"wrong audience": {func(rows []model.CompetencyReportText) []model.CompetencyReportText {
			rows[0].Audience = CompetencyReportAudienceLeader
			return rows
		}, "audience=frontline_employee"},
		"retired": {func(rows []model.CompetencyReportText) []model.CompetencyReportText { rows[0].Status = 1; return rows }, "contentType=overall"},
		"temporary": {func(rows []model.CompetencyReportText) []model.CompetencyReportText {
			rows[0].IsTemporary = 1
			return rows
		}, "temporary"},
		"blank content": {func(rows []model.CompetencyReportText) []model.CompetencyReportText {
			rows[0].Content = " "
			return rows
		}, "blank"},
		"duplicate": {func(rows []model.CompetencyReportText) []model.CompetencyReportText { return append(rows, rows[0]) }, "duplicate"},
		"disclaimer mismatch": {func(rows []model.CompetencyReportText) []model.CompetencyReportText {
			rows[1].Disclaimer = "other"
			return rows
		}, "disclaimer"},
	} {
		t.Run(name, func(t *testing.T) {
			rows := append([]model.CompetencyReportText(nil), valid...)
			_, err := BuildPhase1V2ReportData("run-v2-1", scores, modules, CompetencyPhase1ValidityGood, test.mutate(rows))
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(test.want)) {
				t.Fatalf("error=%v want fragment=%q", err, test.want)
			}
		})
	}
}

func TestBugFB222_Phase1V2ReportRequiresDevelopmentTextsForSelectedItems(t *testing.T) {
	scores, modules := phase1V2SelectorFixture(t, []int{31, 31, 31, 31, 31, 31, 31, 31, 31, 31})
	rows := phase1V2ReportTextRows(scores, modules, CompetencyPhase1ValidityQuestionable)
	filtered := make([]model.CompetencyReportText, 0, len(rows))
	for _, row := range rows {
		if row.ContentType != CompetencyReportContentV2Development {
			filtered = append(filtered, row)
		}
	}
	_, err := BuildPhase1V2ReportData("run-v2-2", scores, modules, CompetencyPhase1ValidityQuestionable, filtered)
	if err == nil || !strings.Contains(err.Error(), "contentType=development") {
		t.Fatalf("error=%v", err)
	}
}

// TestBugFB198_Phase1V2SelectedItemsUseCorrespondingLevelTexts
// Corresponds to docs/regression-tests.md FB-198.
// Score-only selection can choose any level; development items use the workbook's corresponding-level short assessment.
func TestBugFB198_Phase1V2SelectedItemsUseCorrespondingLevelTexts(t *testing.T) {
	scores, modules := phase1V2SelectorFixture(t, []int{24, 23, 22, 21, 20, 19, 18, 17, 16, 15})
	rows := phase1V2ReportTextRows(scores, modules, CompetencyPhase1ValidityGood)
	dto, err := BuildPhase1V2ReportData("run-v2-score-ranking", scores, modules, CompetencyPhase1ValidityGood, rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(dto.Strengths) != 3 || dto.Strengths[0].RuleText != "performance:logical_reasoning:qualified" {
		t.Fatalf("strengths=%+v", dto.Strengths)
	}
	if len(dto.Developments) != 2 || dto.Developments[0].RuleText != "development:dedication:weak" {
		t.Fatalf("developments=%+v", dto.Developments)
	}
}

// TestBugFB222_Phase1V2DevelopmentUsesWorkbookShortAssessment
// Corresponds to docs/regression-tests.md FB-222.
func TestBugFB222_Phase1V2DevelopmentUsesWorkbookShortAssessment(t *testing.T) {
	scores, modules := phase1V2SelectorFixture(t, []int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	rows := phase1V2ReportTextRows(scores, modules, CompetencyPhase1ValidityGood)
	dto, err := BuildPhase1V2ReportData("run-v2-development-text", scores, modules, CompetencyPhase1ValidityGood, rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(dto.Developments) != 2 || dto.Developments[0].RuleText != "development:achievement_orientation:qualified" {
		t.Fatalf("developments=%+v", dto.Developments)
	}
}

func phase1V2ReportTextRows(scores Phase1V2ScoreResult, modules []Phase1V2ModuleScore, validity string) []model.CompetencyReportText {
	rows := make([]model.CompetencyReportText, 0, 40)
	add := func(contentType, identity, condition, content string) {
		rows = append(rows, model.CompetencyReportText{
			ContentVersion: CompetencyPhase1ContentVersionV2, Audience: CompetencyReportAudienceFrontlineEmployee,
			ContentType: contentType, DimensionID: identity, LevelCode: condition, Content: content,
			Disclaimer: "v2-disclaimer", Status: 0,
		})
	}
	add(CompetencyReportContentOverall, "", scores.OverallLevel, "overall:"+scores.OverallLevel)
	add(CompetencyReportContentV2OverallAdvice, "", scores.OverallLevel, "advice:"+scores.OverallLevel)
	add(CompetencyReportContentValidity, "", validity, "validity:"+validity)
	for _, module := range modules {
		add(CompetencyReportContentV2ModuleComparison, module.ModuleCode, module.ComparisonCode, "module:"+module.ModuleCode+":"+module.ComparisonCode)
	}
	for _, dimension := range scores.Dimensions {
		add(CompetencyReportContentDimension, dimension.DimensionID, dimension.Level, "performance:"+dimension.StableKey+":"+dimension.Level)
		if dimension.Level == CompetencyPhase1V2LevelGood || dimension.Level == CompetencyPhase1V2LevelExcellent {
			add(CompetencyReportContentV2Strength, dimension.DimensionID, dimension.Level, "strength:"+dimension.StableKey+":"+dimension.Level)
		} else {
			add(CompetencyReportContentV2Development, dimension.DimensionID, dimension.Level, "development:"+dimension.StableKey+":"+dimension.Level)
		}
	}
	return rows
}

func moduleCodesFromDTO(modules []Phase1V2ReportModule) []string {
	codes := make([]string, 0, len(modules))
	for _, module := range modules {
		codes = append(codes, module.ModuleCode)
	}
	return codes
}

func reportItemKeys(items []Phase1V2ReportSelectedDimension) []string {
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.StableKey)
	}
	return keys
}
