package service

import (
	"errors"
	"math/big"
)

const (
	CompetencyPhase1V2ModuleTask          = "task_management"
	CompetencyPhase1V2ModuleInterpersonal = "interpersonal_management"
	CompetencyPhase1V2ModuleSelf          = "self_management"

	CompetencyPhase1V2ComparisonStandout  = "standout"
	CompetencyPhase1V2ComparisonSuperior  = "superior"
	CompetencyPhase1V2ComparisonAboveNorm = "above_norm"
	CompetencyPhase1V2ComparisonAtNorm    = "at_norm"
	CompetencyPhase1V2ComparisonBelowNorm = "below_norm"
)

type Phase1V2ModuleScore struct {
	ModuleCode              string
	ModuleName              string
	DisplayOrder            int
	ChildDimensionIDs       []string
	TotalDimensionCount     int
	EffectiveDimensionCount int
	TotalQuestionCount      int
	AnsweredQuestionCount   int
	Score                   *big.Rat
	Level                   string
	NormScore               *big.Rat
	ComparisonCode          string
	IsComplete              bool
}

type Phase1V2NormComparison struct {
	NormScore      *big.Rat
	ComparisonCode string
}

type phase1V2ModuleDefinition struct {
	code, name        string
	displayOrder      int
	childDimensionIDs []string
	normScore         *big.Rat
	atThreshold       *big.Rat
	aboveThreshold    *big.Rat
	standoutThreshold *big.Rat
}

func phase1V2ModuleDefinitions() []phase1V2ModuleDefinition {
	return []phase1V2ModuleDefinition{
		{
			code: CompetencyPhase1V2ModuleTask, name: "任务管理类", displayOrder: 1,
			childDimensionIDs: []string{"competency-logical-reasoning", "competency-plan-execution", "competency-digital-application", "competency-achievement-orientation", "competency-continuous-learning"},
			normScore:         big.NewRat(58, 1), atThreshold: big.NewRat(56, 1), aboveThreshold: big.NewRat(60, 1), standoutThreshold: big.NewRat(75, 1),
		},
		{
			code: CompetencyPhase1V2ModuleInterpersonal, name: "人际管理类", displayOrder: 2,
			childDimensionIDs: []string{"competency-communication", "competency-cooperation"},
			normScore:         big.NewRat(215, 4), atThreshold: big.NewRat(50, 1), aboveThreshold: big.NewRat(56, 1), standoutThreshold: big.NewRat(70, 1),
		},
		{
			code: CompetencyPhase1V2ModuleSelf, name: "自我管理类", displayOrder: 3,
			childDimensionIDs: []string{"competency-truth-pragmatism", "competency-self-discipline", "competency-dedication"},
			normScore:         big.NewRat(60, 1), atThreshold: big.NewRat(58, 1), aboveThreshold: big.NewRat(63, 1), standoutThreshold: big.NewRat(75, 1),
		},
	}
}

func CalculatePhase1V2ModuleResults(dimensions []Phase1V2DimensionScore) ([]Phase1V2ModuleScore, error) {
	definitions := Phase1V2Dimensions()
	if len(dimensions) != len(definitions) {
		return nil, errors.New("phase-1 v2 module aggregation requires exactly 10 dimensions")
	}

	byID := make(map[string]Phase1V2DimensionScore, len(dimensions))
	for _, dimension := range dimensions {
		if _, exists := byID[dimension.DimensionID]; exists {
			return nil, errors.New("duplicate phase-1 v2 dimension result")
		}
		byID[dimension.DimensionID] = dimension
	}
	for _, definition := range definitions {
		dimension, exists := byID[definition.ID]
		if !exists {
			return nil, errors.New("missing phase-1 v2 dimension result")
		}
		if dimension.StableKey != definition.StableKey || dimension.DisplayCode != definition.DisplayCode ||
			dimension.DimensionName != definition.Name || dimension.ModuleCode != definition.ModuleCode ||
			dimension.DisplayOrder != definition.DisplayOrder || dimension.TotalQuestionCount != 8 ||
			dimension.AnsweredQuestionCount < 0 || dimension.AnsweredQuestionCount > 8 ||
			dimension.ScoreSum < dimension.AnsweredQuestionCount || dimension.ScoreSum > dimension.AnsweredQuestionCount*5 {
			return nil, errors.New("invalid phase-1 v2 dimension result metadata")
		}
		if dimension.IsComplete {
			if dimension.AnsweredQuestionCount != 8 || dimension.Score == nil {
				return nil, errors.New("complete phase-1 v2 dimension requires a score")
			}
			expectedScore := new(big.Rat).Sub(
				new(big.Rat).Mul(big.NewRat(25, 1), big.NewRat(int64(dimension.ScoreSum), 8)),
				big.NewRat(25, 1),
			)
			level, err := Phase1V2LevelForScore(dimension.Score)
			if err != nil || dimension.Score.Cmp(expectedScore) != 0 || dimension.Level != level {
				return nil, errors.New("phase-1 v2 dimension score and level mismatch")
			}
		} else if dimension.AnsweredQuestionCount == 8 || dimension.Score != nil || dimension.Level != "" {
			return nil, errors.New("incomplete phase-1 v2 dimension cannot have a formal score")
		}
	}

	moduleDefinitions := phase1V2ModuleDefinitions()
	modules := make([]Phase1V2ModuleScore, 0, len(moduleDefinitions))
	for _, definition := range moduleDefinitions {
		module := Phase1V2ModuleScore{
			ModuleCode:        definition.code,
			ModuleName:        definition.name,
			DisplayOrder:      definition.displayOrder,
			ChildDimensionIDs: append([]string(nil), definition.childDimensionIDs...),
			IsComplete:        true,
		}
		total := new(big.Rat)
		for _, dimensionID := range definition.childDimensionIDs {
			dimension := byID[dimensionID]
			module.TotalDimensionCount++
			module.TotalQuestionCount += dimension.TotalQuestionCount
			module.AnsweredQuestionCount += dimension.AnsweredQuestionCount
			if !dimension.IsComplete {
				module.IsComplete = false
				continue
			}
			module.EffectiveDimensionCount++
			total.Add(total, dimension.Score)
		}
		if module.IsComplete {
			module.Score = new(big.Rat).Quo(total, big.NewRat(int64(module.TotalDimensionCount), 1))
			level, err := Phase1V2LevelForScore(module.Score)
			if err != nil {
				return nil, err
			}
			comparison, err := Phase1V2ModuleNormComparison(module.ModuleCode, module.Score)
			if err != nil {
				return nil, err
			}
			module.Level = level
			module.NormScore = comparison.NormScore
			module.ComparisonCode = comparison.ComparisonCode
		}
		modules = append(modules, module)
	}
	return modules, nil
}

func Phase1V2ModuleNormComparison(moduleCode string, score *big.Rat) (Phase1V2NormComparison, error) {
	if err := validatePhase1V2NormScore(score); err != nil {
		return Phase1V2NormComparison{}, err
	}
	for _, definition := range phase1V2ModuleDefinitions() {
		if definition.code != moduleCode {
			continue
		}
		comparison := Phase1V2NormComparison{NormScore: new(big.Rat).Set(definition.normScore)}
		switch {
		case score.Cmp(definition.standoutThreshold) >= 0:
			comparison.ComparisonCode = CompetencyPhase1V2ComparisonStandout
		case score.Cmp(definition.aboveThreshold) >= 0:
			comparison.ComparisonCode = CompetencyPhase1V2ComparisonAboveNorm
		case score.Cmp(definition.atThreshold) >= 0:
			comparison.ComparisonCode = CompetencyPhase1V2ComparisonAtNorm
		default:
			comparison.ComparisonCode = CompetencyPhase1V2ComparisonBelowNorm
		}
		return comparison, nil
	}
	return Phase1V2NormComparison{}, errors.New("unknown phase-1 v2 module")
}

func Phase1V2OverallNormComparison(score *big.Rat) (Phase1V2NormComparison, error) {
	if err := validatePhase1V2NormScore(score); err != nil {
		return Phase1V2NormComparison{}, err
	}
	comparison := Phase1V2NormComparison{NormScore: big.NewRat(231, 4)}
	switch {
	case score.Cmp(big.NewRat(70, 1)) >= 0:
		comparison.ComparisonCode = CompetencyPhase1V2ComparisonSuperior
	case score.Cmp(big.NewRat(63, 1)) >= 0:
		comparison.ComparisonCode = CompetencyPhase1V2ComparisonAboveNorm
	case score.Cmp(big.NewRat(55, 1)) >= 0:
		comparison.ComparisonCode = CompetencyPhase1V2ComparisonAtNorm
	default:
		comparison.ComparisonCode = CompetencyPhase1V2ComparisonBelowNorm
	}
	return comparison, nil
}

func validatePhase1V2NormScore(score *big.Rat) error {
	if score == nil || score.Cmp(big.NewRat(0, 1)) < 0 || score.Cmp(big.NewRat(100, 1)) > 0 {
		return errors.New("phase-1 v2 norm comparison score must be between 0 and 100")
	}
	return nil
}
