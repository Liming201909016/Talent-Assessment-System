package service

import (
	"math/big"
	"reflect"
	"testing"
)

// TestBugFB177_Phase1V2SelectorsUseExactScoresAndStableTies
// Corresponds to docs/regression-tests.md FB-177.
func TestBugFB177_Phase1V2SelectorsUseExactScoresAndStableTies(t *testing.T) {
	scores, modules := phase1V2SelectorFixture(t, []int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	selection, err := SelectPhase1V2ReportOverview(scores.Dimensions, modules)
	if err != nil {
		t.Fatal(err)
	}
	if got := moduleCodes(selection.Modules); !reflect.DeepEqual(got, []string{CompetencyPhase1V2ModuleSelf, CompetencyPhase1V2ModuleTask, CompetencyPhase1V2ModuleInterpersonal}) {
		t.Fatalf("module order=%v", got)
	}
	if got := dimensionKeys(selection.Strengths); !reflect.DeepEqual(got, []string{"digital_application", "truth_pragmatism", "self_discipline"}) {
		t.Fatalf("strengths=%v", got)
	}
	if got := dimensionKeys(selection.Developments); !reflect.DeepEqual(got, []string{"achievement_orientation", "communication"}) {
		t.Fatalf("developments=%v", got)
	}

	tiedModules := append([]Phase1V2ModuleScore(nil), modules...)
	for index := range tiedModules {
		tiedModules[index].Score = big.NewRat(60, 1)
		tiedModules[index].Level = CompetencyPhase1V2LevelQualified
		comparison, comparisonErr := Phase1V2ModuleNormComparison(tiedModules[index].ModuleCode, tiedModules[index].Score)
		if comparisonErr != nil {
			t.Fatal(comparisonErr)
		}
		tiedModules[index].NormScore = comparison.NormScore
		tiedModules[index].ComparisonCode = comparison.ComparisonCode
	}
	for left, right := 0, len(tiedModules)-1; left < right; left, right = left+1, right-1 {
		tiedModules[left], tiedModules[right] = tiedModules[right], tiedModules[left]
	}
	sorted, err := SortPhase1V2ModulesForOverview(tiedModules)
	if err != nil {
		t.Fatal(err)
	}
	if got := moduleCodes(sorted); !reflect.DeepEqual(got, []string{CompetencyPhase1V2ModuleTask, CompetencyPhase1V2ModuleInterpersonal, CompetencyPhase1V2ModuleSelf}) {
		t.Fatalf("tied module order=%v", got)
	}
}

func TestBugFB177_Phase1V2SelectorsAlwaysPopulateScoreRankings(t *testing.T) {
	allGood, modules := phase1V2SelectorFixture(t, []int{31, 31, 31, 31, 31, 31, 31, 31, 31, 31})
	selection, err := SelectPhase1V2ReportOverview(allGood.Dimensions, modules)
	if err != nil {
		t.Fatal(err)
	}
	if len(selection.Strengths) != 3 || len(selection.Developments) != 2 {
		t.Fatalf("all-good selection=%+v", selection)
	}

	noGood, modules := phase1V2SelectorFixture(t, []int{24, 24, 24, 24, 24, 24, 24, 24, 24, 24})
	selection, err = SelectPhase1V2ReportOverview(noGood.Dimensions, modules)
	if err != nil {
		t.Fatal(err)
	}
	if len(selection.Strengths) != 3 || len(selection.Developments) != 2 {
		t.Fatalf("no-good selection=%+v", selection)
	}

	oneGood, modules := phase1V2SelectorFixture(t, []int{31, 28, 28, 28, 28, 28, 28, 28, 28, 28})
	selection, err = SelectPhase1V2ReportOverview(oneGood.Dimensions, modules)
	if err != nil || len(selection.Strengths) != 3 || len(selection.Developments) != 2 || selection.Strengths[0].StableKey != "logical_reasoning" {
		t.Fatalf("one-good selection=%+v error=%v", selection, err)
	}
}

// TestBugFB198_Phase1V2OverviewAlwaysUsesHighestThreeAndLowestTwoScores
// Corresponds to docs/regression-tests.md FB-198.
// Reproduction: when all dimensions are below good, the old category filter emits no strengths.
// Expected: strengths are always the three highest scores and developments are always the two lowest scores.
func TestBugFB198_Phase1V2OverviewAlwaysUsesHighestThreeAndLowestTwoScores(t *testing.T) {
	scores, modules := phase1V2SelectorFixture(t, []int{24, 23, 22, 21, 20, 19, 18, 17, 16, 15})
	selection, err := SelectPhase1V2ReportOverview(scores.Dimensions, modules)
	if err != nil {
		t.Fatal(err)
	}
	if got := dimensionKeys(selection.Strengths); !reflect.DeepEqual(got, []string{"logical_reasoning", "plan_execution", "digital_application"}) {
		t.Fatalf("highest-three strengths=%v", got)
	}
	if got := dimensionKeys(selection.Developments); !reflect.DeepEqual(got, []string{"dedication", "self_discipline"}) {
		t.Fatalf("lowest-two developments=%v", got)
	}
}

func TestBugFB177_Phase1V2AdviceUsesSeparateTwoAndThreeRules(t *testing.T) {
	scores, _ := phase1V2SelectorFixture(t, []int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	for _, test := range []struct {
		level string
		want  []string
	}{
		{CompetencyPhase1V2LevelExcellent, []string{"achievement_orientation", "communication"}},
		{CompetencyPhase1V2LevelGood, []string{"achievement_orientation", "communication"}},
		{CompetencyPhase1V2LevelQualified, []string{"achievement_orientation", "communication"}},
		{CompetencyPhase1V2LevelWeak, []string{"achievement_orientation", "communication", "plan_execution"}},
		{CompetencyPhase1V2LevelInsufficient, []string{"achievement_orientation", "communication", "plan_execution"}},
	} {
		got, err := SelectPhase1V2OverallAdviceDimensions(scores.Dimensions, test.level)
		if err != nil || !reflect.DeepEqual(dimensionKeys(got), test.want) {
			t.Fatalf("level=%q dimensions=%v error=%v want=%v", test.level, dimensionKeys(got), err, test.want)
		}
	}
}

func TestBugFB177_Phase1V2SelectorsRejectIncompleteOrMalformedInput(t *testing.T) {
	scores, modules := phase1V2SelectorFixture(t, []int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	badDimensions := append([]Phase1V2DimensionScore(nil), scores.Dimensions...)
	badDimensions[0].Level = CompetencyPhase1V2LevelExcellent
	if _, err := SelectPhase1V2ReportOverview(badDimensions, modules); err == nil {
		t.Fatal("dimension level mismatch must be rejected")
	}
	badDimensions = append([]Phase1V2DimensionScore(nil), scores.Dimensions...)
	badDimensions[0].IsComplete = false
	badDimensions[0].Score = nil
	badDimensions[0].Level = ""
	if _, err := SelectPhase1V2OverallAdviceDimensions(badDimensions, scores.OverallLevel); err == nil {
		t.Fatal("incomplete dimension must be rejected")
	}
	badModules := append([]Phase1V2ModuleScore(nil), modules...)
	badModules[0].ComparisonCode = "unknown"
	if _, err := SortPhase1V2ModulesForOverview(badModules); err == nil {
		t.Fatal("malformed module must be rejected")
	}
	badModules = append([]Phase1V2ModuleScore(nil), modules...)
	badModules[0].Score = big.NewRat(60, 1)
	badModules[0].Level = CompetencyPhase1V2LevelQualified
	comparison, err := Phase1V2ModuleNormComparison(badModules[0].ModuleCode, badModules[0].Score)
	if err != nil {
		t.Fatal(err)
	}
	badModules[0].NormScore = comparison.NormScore
	badModules[0].ComparisonCode = comparison.ComparisonCode
	if _, err := SelectPhase1V2ReportOverview(scores.Dimensions, badModules); err == nil {
		t.Fatal("module scores inconsistent with dimensions must be rejected")
	}
	if _, err := SelectPhase1V2OverallAdviceDimensions(scores.Dimensions, "unknown"); err == nil {
		t.Fatal("unknown overall level must be rejected")
	}
}

func phase1V2SelectorFixture(t *testing.T, sums []int) (Phase1V2ScoreResult, []Phase1V2ModuleScore) {
	t.Helper()
	scores, err := CalculatePhase1V2CompetencyResultFromV1(phase1V2InputsFromV1Sums(sums))
	if err != nil {
		t.Fatal(err)
	}
	modules, err := CalculatePhase1V2ModuleResults(scores.Dimensions)
	if err != nil {
		t.Fatal(err)
	}
	return scores, modules
}

func dimensionKeys(dimensions []Phase1V2DimensionScore) []string {
	keys := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		keys = append(keys, dimension.StableKey)
	}
	return keys
}

func moduleCodes(modules []Phase1V2ModuleScore) []string {
	codes := make([]string, 0, len(modules))
	for _, module := range modules {
		codes = append(codes, module.ModuleCode)
	}
	return codes
}
