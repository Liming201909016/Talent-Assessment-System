package service

import (
	"errors"
	"math/big"
)

const (
	CompetencyPhase1V2LevelExcellent    = "excellent"
	CompetencyPhase1V2LevelGood         = "good"
	CompetencyPhase1V2LevelQualified    = "qualified"
	CompetencyPhase1V2LevelWeak         = "weak"
	CompetencyPhase1V2LevelInsufficient = "insufficient"
)

type Phase1V2DimensionScore struct {
	DimensionID           string
	StableKey             string
	DisplayCode           string
	DimensionName         string
	ModuleCode            string
	DisplayOrder          int
	TotalQuestionCount    int
	AnsweredQuestionCount int
	ScoreSum              int
	Score                 *big.Rat
	Level                 string
	IsComplete            bool
}

type Phase1V2ScoreResult struct {
	Dimensions              []Phase1V2DimensionScore
	TotalQuestionCount      int
	AnsweredQuestionCount   int
	EffectiveDimensionCount int
	OverallScore            *big.Rat
	OverallLevel            string
	IsComplete              bool
}

// CalculatePhase1V2CompetencyResultFromV1 maps frozen v1 answer identities to
// v2 semantic dimensions and calculates exact percentage scores. It does not
// perform module aggregation, norm comparison, persistence, or display rounding.
func CalculatePhase1V2CompetencyResultFromV1(inputs []CompetencyScoreInput) (Phase1V2ScoreResult, error) {
	const questionCount = 80
	if len(inputs) != questionCount {
		return Phase1V2ScoreResult{}, errors.New("phase-1 v2 scoring requires exactly 80 dimension questions")
	}

	definitions := Phase1V2Dimensions()
	byID := make(map[string]*Phase1V2DimensionScore, len(definitions))
	for _, definition := range definitions {
		byID[definition.ID] = &Phase1V2DimensionScore{
			DimensionID:   definition.ID,
			StableKey:     definition.StableKey,
			DisplayCode:   definition.DisplayCode,
			DimensionName: definition.Name,
			ModuleCode:    definition.ModuleCode,
			DisplayOrder:  definition.DisplayOrder,
		}
	}
	legacyOrders := make(map[string]int, len(competencyPhase1DimensionIDs))
	for index, id := range competencyPhase1DimensionIDs {
		legacyOrders[id] = index + 1
	}

	result := Phase1V2ScoreResult{
		Dimensions:         make([]Phase1V2DimensionScore, 0, len(definitions)),
		TotalQuestionCount: len(inputs),
		IsComplete:         true,
	}
	for _, input := range inputs {
		if input.QuestionType != CompetencyQuestionTypeDimension {
			return Phase1V2ScoreResult{}, errors.New("phase-1 v2 scoring cannot include validity questions")
		}
		if input.DisplayOrder != legacyOrders[input.DimensionID] {
			return Phase1V2ScoreResult{}, errors.New("phase-1 v1 source dimension order mismatch")
		}
		definition, err := MapPhase1V1DimensionToV2(input.DimensionID)
		if err != nil {
			return Phase1V2ScoreResult{}, err
		}
		dimension := byID[definition.ID]
		dimension.TotalQuestionCount++
		if !input.Answered {
			result.IsComplete = false
			continue
		}
		if input.FinalScore < 1 || input.FinalScore > 5 {
			return Phase1V2ScoreResult{}, errors.New("competency final score must be between 1 and 5")
		}
		dimension.AnsweredQuestionCount++
		dimension.ScoreSum += input.FinalScore
		result.AnsweredQuestionCount++
	}

	for _, definition := range definitions {
		dimension := byID[definition.ID]
		if dimension.TotalQuestionCount != 8 {
			return Phase1V2ScoreResult{}, errors.New("each phase-1 v2 dimension requires exactly 8 questions")
		}
		dimension.IsComplete = dimension.AnsweredQuestionCount == 8
		if !dimension.IsComplete {
			result.IsComplete = false
		} else {
			dimension.Score = new(big.Rat).Sub(
				new(big.Rat).Mul(big.NewRat(25, 1), big.NewRat(int64(dimension.ScoreSum), 8)),
				big.NewRat(25, 1),
			)
			level, err := Phase1V2LevelForScore(dimension.Score)
			if err != nil {
				return Phase1V2ScoreResult{}, err
			}
			dimension.Level = level
			result.EffectiveDimensionCount++
		}
		result.Dimensions = append(result.Dimensions, *dimension)
	}
	if !result.IsComplete {
		return result, nil
	}

	total := new(big.Rat)
	for index := range result.Dimensions {
		total.Add(total, result.Dimensions[index].Score)
	}
	result.OverallScore = new(big.Rat).Quo(total, big.NewRat(int64(len(result.Dimensions)), 1))
	level, err := Phase1V2LevelForScore(result.OverallScore)
	if err != nil {
		return Phase1V2ScoreResult{}, err
	}
	result.OverallLevel = level
	return result, nil
}

func Phase1V2LevelForScore(score *big.Rat) (string, error) {
	if score == nil || score.Cmp(big.NewRat(0, 1)) < 0 || score.Cmp(big.NewRat(100, 1)) > 0 {
		return "", errors.New("phase-1 v2 score must be between 0 and 100")
	}
	switch {
	case score.Cmp(big.NewRat(90, 1)) >= 0:
		return CompetencyPhase1V2LevelExcellent, nil
	case score.Cmp(big.NewRat(70, 1)) >= 0:
		return CompetencyPhase1V2LevelGood, nil
	case score.Cmp(big.NewRat(30, 1)) >= 0:
		return CompetencyPhase1V2LevelQualified, nil
	case score.Cmp(big.NewRat(10, 1)) >= 0:
		return CompetencyPhase1V2LevelWeak, nil
	default:
		return CompetencyPhase1V2LevelInsufficient, nil
	}
}
