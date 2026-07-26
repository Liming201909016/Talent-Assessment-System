package service

import (
	"errors"
	"math/big"
	"sort"
)

const (
	CompetencyDirectionForward = "forward"
	CompetencyDirectionReverse = "reverse"

	CompetencyLevelLow     = "low"
	CompetencyLevelAverage = "average"
	CompetencyLevelGood    = "good"
	CompetencyLevelHigh    = "high"
)

// CompetencyScoreInput is one snapshotted paper question prepared for scoring.
// FinalScore is ignored when Answered is false.
type CompetencyScoreInput struct {
	DimensionID   string
	DimensionCode string
	DimensionName string
	DisplayOrder  int
	Answered      bool
	FinalScore    int
}

// CompetencyDimensionScore contains the exact score for one selected dimension.
// Score is nil when the dimension has no answered questions.
type CompetencyDimensionScore struct {
	DimensionID           string
	DimensionCode         string
	DimensionName         string
	DisplayOrder          int
	TotalQuestionCount    int
	AnsweredQuestionCount int
	ScoreSum              int
	Score                 *big.Rat
	Level                 string
	IsComplete            bool
}

// CompetencyScoreResult contains exact rational values. Presentation and
// persistence layers may format them only after all calculations finish.
type CompetencyScoreResult struct {
	Dimensions              []CompetencyDimensionScore
	TotalQuestionCount      int
	AnsweredQuestionCount   int
	EffectiveDimensionCount int
	OverallScore            *big.Rat
	EvaluationAverage       *big.Rat
	EvaluationLevel         string
	IsComplete              bool
}

// CalculateCompetencyQuestionScore applies the shared five-point scale.
func CalculateCompetencyQuestionScore(rawScore int, direction string) (int, error) {
	if rawScore < 1 || rawScore > 5 {
		return 0, errors.New("competency raw score must be between 1 and 5")
	}
	switch direction {
	case CompetencyDirectionForward:
		return rawScore, nil
	case CompetencyDirectionReverse:
		return 6 - rawScore, nil
	default:
		return 0, errors.New("invalid competency scoring direction")
	}
}

// CalculateCompetencyResult groups by snapshotted dimension identity, averages
// answered questions only, and sums exact dimension averages without rounding.
func CalculateCompetencyResult(inputs []CompetencyScoreInput) (CompetencyScoreResult, error) {
	result := CompetencyScoreResult{
		Dimensions:   make([]CompetencyDimensionScore, 0),
		OverallScore: new(big.Rat),
		IsComplete:   true,
	}
	byDimension := make(map[string]*CompetencyDimensionScore)

	for _, input := range inputs {
		if input.DimensionID == "" {
			return CompetencyScoreResult{}, errors.New("competency dimension id is required")
		}
		dimension, exists := byDimension[input.DimensionID]
		if !exists {
			dimension = &CompetencyDimensionScore{
				DimensionID:   input.DimensionID,
				DimensionCode: input.DimensionCode,
				DimensionName: input.DimensionName,
				DisplayOrder:  input.DisplayOrder,
			}
			byDimension[input.DimensionID] = dimension
		}
		dimension.TotalQuestionCount++
		result.TotalQuestionCount++
		if !input.Answered {
			continue
		}
		if input.FinalScore < 1 || input.FinalScore > 5 {
			return CompetencyScoreResult{}, errors.New("competency final score must be between 1 and 5")
		}
		dimension.AnsweredQuestionCount++
		dimension.ScoreSum += input.FinalScore
		result.AnsweredQuestionCount++
	}

	for _, dimension := range byDimension {
		dimension.IsComplete = dimension.AnsweredQuestionCount == dimension.TotalQuestionCount
		if !dimension.IsComplete {
			result.IsComplete = false
		}
		if dimension.AnsweredQuestionCount > 0 {
			dimension.Score = big.NewRat(int64(dimension.ScoreSum), int64(dimension.AnsweredQuestionCount))
			level, err := CompetencyLevelForScore(dimension.Score)
			if err != nil {
				return CompetencyScoreResult{}, err
			}
			dimension.Level = level
			result.OverallScore.Add(result.OverallScore, dimension.Score)
			result.EffectiveDimensionCount++
		}
		result.Dimensions = append(result.Dimensions, *dimension)
	}

	sort.Slice(result.Dimensions, func(i, j int) bool {
		if result.Dimensions[i].DisplayOrder != result.Dimensions[j].DisplayOrder {
			return result.Dimensions[i].DisplayOrder < result.Dimensions[j].DisplayOrder
		}
		if result.Dimensions[i].DimensionCode != result.Dimensions[j].DimensionCode {
			return result.Dimensions[i].DimensionCode < result.Dimensions[j].DimensionCode
		}
		return result.Dimensions[i].DimensionID < result.Dimensions[j].DimensionID
	})

	if result.EffectiveDimensionCount > 0 {
		result.EvaluationAverage = new(big.Rat).Quo(
			result.OverallScore,
			big.NewRat(int64(result.EffectiveDimensionCount), 1),
		)
		level, err := CompetencyLevelForScore(result.EvaluationAverage)
		if err != nil {
			return CompetencyScoreResult{}, err
		}
		result.EvaluationLevel = level
	}
	return result, nil
}

// CompetencyLevelForScore applies the confirmed equal-width 1.00-5.00 bands.
func CompetencyLevelForScore(score *big.Rat) (string, error) {
	if score == nil || score.Cmp(big.NewRat(1, 1)) < 0 || score.Cmp(big.NewRat(5, 1)) > 0 {
		return "", errors.New("competency score must be between 1 and 5")
	}
	switch {
	case score.Cmp(big.NewRat(2, 1)) < 0:
		return CompetencyLevelLow, nil
	case score.Cmp(big.NewRat(3, 1)) < 0:
		return CompetencyLevelAverage, nil
	case score.Cmp(big.NewRat(4, 1)) < 0:
		return CompetencyLevelGood, nil
	default:
		return CompetencyLevelHigh, nil
	}
}

// FormatCompetencyScore is the only presentation rounding helper for this flow.
func FormatCompetencyScore(score *big.Rat) string {
	if score == nil {
		return ""
	}
	return score.FloatString(2)
}
