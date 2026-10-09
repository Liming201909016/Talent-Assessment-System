package service

import (
	"errors"
	"math/big"
	"sort"
)

type Phase1V2ReportOverviewSelection struct {
	Modules      []Phase1V2ModuleScore
	Strengths    []Phase1V2DimensionScore
	Developments []Phase1V2DimensionScore
}

func SelectPhase1V2ReportOverview(dimensions []Phase1V2DimensionScore, modules []Phase1V2ModuleScore) (Phase1V2ReportOverviewSelection, error) {
	if err := validatePhase1V2SelectorDimensions(dimensions); err != nil {
		return Phase1V2ReportOverviewSelection{}, err
	}
	sortedModules, err := SortPhase1V2ModulesForOverview(modules)
	if err != nil {
		return Phase1V2ReportOverviewSelection{}, err
	}
	expectedModules, err := CalculatePhase1V2ModuleResults(dimensions)
	if err != nil {
		return Phase1V2ReportOverviewSelection{}, err
	}
	actualByCode := make(map[string]Phase1V2ModuleScore, len(modules))
	for _, module := range modules {
		actualByCode[module.ModuleCode] = module
	}
	for _, expected := range expectedModules {
		if !equalPhase1V2ModuleScore(actualByCode[expected.ModuleCode], expected) {
			return Phase1V2ReportOverviewSelection{}, errors.New("phase-1 v2 modules do not match dimensions")
		}
	}

	strengths := append([]Phase1V2DimensionScore(nil), dimensions...)
	developments := append([]Phase1V2DimensionScore(nil), dimensions...)
	sort.Slice(strengths, func(i, j int) bool {
		comparison := strengths[i].Score.Cmp(strengths[j].Score)
		if comparison != 0 {
			return comparison > 0
		}
		return strengths[i].DisplayOrder < strengths[j].DisplayOrder
	})
	sort.Slice(developments, func(i, j int) bool {
		comparison := developments[i].Score.Cmp(developments[j].Score)
		if comparison != 0 {
			return comparison < 0
		}
		return developments[i].DisplayOrder < developments[j].DisplayOrder
	})
	strengths = strengths[:3]
	developments = developments[:2]
	return Phase1V2ReportOverviewSelection{
		Modules:      sortedModules,
		Strengths:    strengths,
		Developments: developments,
	}, nil
}

func SortPhase1V2ModulesForOverview(modules []Phase1V2ModuleScore) ([]Phase1V2ModuleScore, error) {
	definitions := phase1V2ModuleDefinitions()
	if len(modules) != len(definitions) {
		return nil, errors.New("phase-1 v2 overview requires exactly 3 modules")
	}
	byCode := make(map[string]Phase1V2ModuleScore, len(modules))
	for _, module := range modules {
		if _, exists := byCode[module.ModuleCode]; exists {
			return nil, errors.New("duplicate phase-1 v2 module")
		}
		byCode[module.ModuleCode] = module
	}
	for _, definition := range definitions {
		module, exists := byCode[definition.code]
		if !exists || module.ModuleName != definition.name || module.DisplayOrder != definition.displayOrder ||
			!equalStringSlices(module.ChildDimensionIDs, definition.childDimensionIDs) ||
			module.TotalDimensionCount != len(definition.childDimensionIDs) || module.EffectiveDimensionCount != len(definition.childDimensionIDs) ||
			module.TotalQuestionCount != len(definition.childDimensionIDs)*8 || module.AnsweredQuestionCount != len(definition.childDimensionIDs)*8 ||
			!module.IsComplete || module.Score == nil || module.NormScore == nil {
			return nil, errors.New("invalid phase-1 v2 module result")
		}
		level, err := Phase1V2LevelForScore(module.Score)
		if err != nil || module.Level != level {
			return nil, errors.New("phase-1 v2 module score and level mismatch")
		}
		comparison, err := Phase1V2ModuleNormComparison(module.ModuleCode, module.Score)
		if err != nil || module.NormScore.Cmp(comparison.NormScore) != 0 || module.ComparisonCode != comparison.ComparisonCode {
			return nil, errors.New("phase-1 v2 module norm comparison mismatch")
		}
	}

	sorted := append([]Phase1V2ModuleScore(nil), modules...)
	sort.Slice(sorted, func(i, j int) bool {
		comparison := sorted[i].Score.Cmp(sorted[j].Score)
		if comparison != 0 {
			return comparison > 0
		}
		return sorted[i].DisplayOrder < sorted[j].DisplayOrder
	})
	return sorted, nil
}

func SelectPhase1V2OverallAdviceDimensions(dimensions []Phase1V2DimensionScore, overallLevel string) ([]Phase1V2DimensionScore, error) {
	if err := validatePhase1V2SelectorDimensions(dimensions); err != nil {
		return nil, err
	}
	count := 0
	switch overallLevel {
	case CompetencyPhase1V2LevelExcellent, CompetencyPhase1V2LevelGood, CompetencyPhase1V2LevelQualified:
		count = 2
	case CompetencyPhase1V2LevelWeak, CompetencyPhase1V2LevelInsufficient:
		count = 3
	default:
		return nil, errors.New("unknown phase-1 v2 overall level")
	}

	sorted := append([]Phase1V2DimensionScore(nil), dimensions...)
	sort.Slice(sorted, func(i, j int) bool {
		comparison := sorted[i].Score.Cmp(sorted[j].Score)
		if comparison != 0 {
			return comparison < 0
		}
		return sorted[i].DisplayOrder < sorted[j].DisplayOrder
	})
	return sorted[:count], nil
}

func validatePhase1V2SelectorDimensions(dimensions []Phase1V2DimensionScore) error {
	definitions := Phase1V2Dimensions()
	if len(dimensions) != len(definitions) {
		return errors.New("phase-1 v2 selector requires exactly 10 dimensions")
	}
	byID := make(map[string]Phase1V2DimensionScore, len(dimensions))
	for _, dimension := range dimensions {
		if _, exists := byID[dimension.DimensionID]; exists {
			return errors.New("duplicate phase-1 v2 selector dimension")
		}
		byID[dimension.DimensionID] = dimension
	}
	for _, definition := range definitions {
		dimension, exists := byID[definition.ID]
		if !exists || dimension.StableKey != definition.StableKey || dimension.DisplayCode != definition.DisplayCode ||
			dimension.DimensionName != definition.Name || dimension.ModuleCode != definition.ModuleCode || dimension.DisplayOrder != definition.DisplayOrder ||
			dimension.TotalQuestionCount != 8 || dimension.AnsweredQuestionCount != 8 || dimension.ScoreSum < 8 || dimension.ScoreSum > 40 ||
			!dimension.IsComplete || dimension.Score == nil {
			return errors.New("invalid phase-1 v2 selector dimension")
		}
		expectedScore := new(big.Rat).Sub(
			new(big.Rat).Mul(big.NewRat(25, 1), big.NewRat(int64(dimension.ScoreSum), 8)),
			big.NewRat(25, 1),
		)
		level, err := Phase1V2LevelForScore(dimension.Score)
		if err != nil || dimension.Score.Cmp(expectedScore) != 0 || dimension.Level != level {
			return errors.New("phase-1 v2 selector dimension score and level mismatch")
		}
	}
	return nil
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalPhase1V2ModuleScore(left, right Phase1V2ModuleScore) bool {
	return left.ModuleCode == right.ModuleCode && left.ModuleName == right.ModuleName && left.DisplayOrder == right.DisplayOrder &&
		equalStringSlices(left.ChildDimensionIDs, right.ChildDimensionIDs) &&
		left.TotalDimensionCount == right.TotalDimensionCount && left.EffectiveDimensionCount == right.EffectiveDimensionCount &&
		left.TotalQuestionCount == right.TotalQuestionCount && left.AnsweredQuestionCount == right.AnsweredQuestionCount &&
		equalRat(left.Score, right.Score) && left.Level == right.Level && equalRat(left.NormScore, right.NormScore) &&
		left.ComparisonCode == right.ComparisonCode && left.IsComplete == right.IsComplete
}

func equalRat(left, right *big.Rat) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Cmp(right) == 0
}
