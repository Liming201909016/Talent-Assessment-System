package service

import (
	"errors"
	"math/big"
	"sort"
)

type ManagementTraitsAnswer struct {
	Number   int
	Answered bool
	Raw      int
}

type ManagementTraitsDimensionScore struct {
	Key           string
	Name          string
	Module        string
	Order         int
	QuestionCount int
	AnsweredCount int
	ScoreSum      int
	Score         *big.Rat
	Norm          *big.Rat
	Level         string
}

type ManagementTraitsModuleScore struct {
	Key            string
	DimensionCount int
	Score          *big.Rat
}

type ManagementTraitsResult struct {
	Questionnaire         string
	TotalQuestionCount    int
	AnsweredQuestionCount int
	IsComplete            bool
	Dimensions            []ManagementTraitsDimensionScore
	Modules               []ManagementTraitsModuleScore
	OverallScore          *big.Rat
	OverallNorm           *big.Rat
	OverallLevel          string
	Highest               []string
	Lowest                []string
}

// CalculateManagementTraits is a pure input boundary. It does not identify source
// question text, authorize a participant, persist a result, or activate a version.
func CalculateManagementTraits(questionnaire string, inputs []ManagementTraitsAnswer) (ManagementTraitsResult, error) {
	const questionCount = 140
	if questionnaire != ManagementTraitsQuestionnaireStaff && questionnaire != ManagementTraitsQuestionnaireLeader {
		return ManagementTraitsResult{}, errors.New("unsupported management traits questionnaire")
	}
	if len(inputs) != questionCount {
		return ManagementTraitsResult{}, errors.New("management traits requires exactly 140 question rows")
	}
	var answers [questionCount + 1]ManagementTraitsAnswer
	var seen [questionCount + 1]bool
	answered := 0
	for _, input := range inputs {
		if input.Number < 1 || input.Number > questionCount || seen[input.Number] {
			return ManagementTraitsResult{}, errors.New("management traits requires unique V1 through V140")
		}
		if input.Answered {
			if input.Raw < 1 || input.Raw > 5 {
				return ManagementTraitsResult{}, errors.New("management traits raw answer must be between 1 and 5")
			}
			answered++
		} else if input.Raw != 0 {
			return ManagementTraitsResult{}, errors.New("unanswered management traits item must have raw zero")
		}
		seen[input.Number] = true
		answers[input.Number] = input
	}

	definitions := ManagementTraitsDimensions()
	result := ManagementTraitsResult{
		Questionnaire: questionnaire, TotalQuestionCount: questionCount,
		AnsweredQuestionCount: answered, IsComplete: answered == questionCount,
		Dimensions: make([]ManagementTraitsDimensionScore, 0, len(definitions)),
		Modules:    make([]ManagementTraitsModuleScore, 0, 4),
		Highest:    make([]string, 0, 3), Lowest: make([]string, 0, 3),
	}
	for _, definition := range definitions {
		dimension := ManagementTraitsDimensionScore{
			Key: definition.Key, Name: definition.Name, Module: definition.Module,
			Order: definition.Order, QuestionCount: len(definition.Items),
		}
		for _, item := range definition.Items {
			input := answers[item.Number]
			if !input.Answered {
				continue
			}
			final := input.Raw
			if item.Reverse {
				final = 6 - final
			}
			dimension.AnsweredCount++
			dimension.ScoreSum += final
		}
		if result.IsComplete {
			dimension.Score = big.NewRat(int64(25*(dimension.ScoreSum-dimension.QuestionCount)), int64(dimension.QuestionCount))
			dimension.Norm = definition.Norm
			dimension.Level, _ = ManagementTraitsLevelForScore(dimension.Score)
		}
		result.Dimensions = append(result.Dimensions, dimension)
	}
	// Calculation order only; report module ordering is a separate presentation contract.
	for _, key := range []string{"self", "interpersonal", "task", "development"} {
		module := ManagementTraitsModuleScore{Key: key}
		total := new(big.Rat)
		for _, dimension := range result.Dimensions {
			if dimension.Module != key {
				continue
			}
			module.DimensionCount++
			if result.IsComplete {
				total.Add(total, dimension.Score)
			}
		}
		if result.IsComplete {
			module.Score = total.Quo(total, big.NewRat(int64(module.DimensionCount), 1))
		}
		result.Modules = append(result.Modules, module)
	}
	if !result.IsComplete {
		return result, nil
	}

	total, norm := new(big.Rat), new(big.Rat)
	for _, dimension := range result.Dimensions {
		total.Add(total, dimension.Score)
		norm.Add(norm, dimension.Norm)
	}
	count := big.NewRat(int64(len(result.Dimensions)), 1)
	result.OverallScore = total.Quo(total, count)
	result.OverallNorm = norm.Quo(norm, count)
	result.OverallLevel, _ = ManagementTraitsLevelForScore(result.OverallScore)
	for _, descending := range []bool{true, false} {
		indices := make([]int, len(result.Dimensions))
		for i := range indices {
			indices[i] = i
		}
		sort.Slice(indices, func(i, j int) bool {
			left, right := result.Dimensions[indices[i]], result.Dimensions[indices[j]]
			comparison := left.Score.Cmp(right.Score)
			if comparison == 0 {
				return left.Order < right.Order
			}
			if descending {
				return comparison > 0
			}
			return comparison < 0
		})
		for _, index := range indices[:3] {
			if descending {
				result.Highest = append(result.Highest, result.Dimensions[index].Key)
			} else {
				result.Lowest = append(result.Lowest, result.Dimensions[index].Key)
			}
		}
	}
	return result, nil
}

// ManagementTraitsLevelForScore compares exact percentages, never display values.
func ManagementTraitsLevelForScore(score *big.Rat) (string, error) {
	if score == nil || score.Cmp(big.NewRat(0, 1)) < 0 || score.Cmp(big.NewRat(100, 1)) > 0 {
		return "", errors.New("management traits score must be between 0 and 100")
	}
	switch {
	case score.Cmp(big.NewRat(90, 1)) >= 0:
		return "excellent", nil
	case score.Cmp(big.NewRat(70, 1)) >= 0:
		return "good", nil
	case score.Cmp(big.NewRat(30, 1)) >= 0:
		return "qualified", nil
	case score.Cmp(big.NewRat(10, 1)) >= 0:
		return "weak", nil
	default:
		return "insufficient", nil
	}
}
