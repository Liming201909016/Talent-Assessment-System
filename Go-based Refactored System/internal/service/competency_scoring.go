package service

import (
	"errors"
	"math/big"
	"sort"
)

const (
	CompetencyDirectionForward      = "forward"
	CompetencyDirectionReverse      = "reverse"
	CompetencyQuestionTypeDimension = "dimension"
	CompetencyQuestionTypeValidity  = "validity"

	CompetencyLevelLow     = "low"
	CompetencyLevelAverage = "average"
	CompetencyLevelGood    = "good"
	CompetencyLevelHigh    = "high"

	CompetencyPhase1LevelL1 = "L1"
	CompetencyPhase1LevelL2 = "L2"
	CompetencyPhase1LevelL3 = "L3"
	CompetencyPhase1LevelL4 = "L4"
	CompetencyPhase1LevelL5 = "L5"

	CompetencyPhase1OverallExcellent    = "excellent"
	CompetencyPhase1OverallGood         = "good"
	CompetencyPhase1OverallQualified    = "qualified"
	CompetencyPhase1OverallWeak         = "weak"
	CompetencyPhase1OverallNotQualified = "not_qualified"

	CompetencyPhase1GroupGeneralAbility       = "general_ability"
	CompetencyPhase1GroupPsychologicalQuality = "psychological_quality"

	CompetencyPhase1ValidityGood         = "good"
	CompetencyPhase1ValidityQuestionable = "questionable"
	CompetencyPhase1ValidityIncomplete   = "incomplete"
)

// CompetencyScoreInput is one snapshotted paper question prepared for scoring.
// FinalScore is ignored when Answered is false.
type CompetencyScoreInput struct {
	DimensionID   string
	DimensionCode string
	DimensionName string
	DisplayOrder  int
	QuestionType  string
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

type CompetencyGroupScore struct {
	GroupCode               string
	GroupName               string
	DisplayOrder            int
	ChildDimensionIDs       []string
	TotalDimensionCount     int
	EffectiveDimensionCount int
	TotalQuestionCount      int
	AnsweredQuestionCount   int
	Score                   *big.Rat
	Level                   string
	IsComplete              bool
}

type Phase1ValidityInput struct {
	QuestionCode string
	Order        int
	QuestionType string
	Direction    string
	Answered     bool
	RawValue     int
}

type Phase1ValidityResult struct {
	TotalQuestionCount    int
	AnsweredQuestionCount int
	Score                 *int
	Status                string
	IsComplete            bool
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

// CalculatePhase1CompetencyResult applies the fixed phase-1 ten-dimension
// scoring contract. Only dimension questions enter this function; validity is
// calculated independently. Incomplete timeout results retain counts but do
// not expose formal dimension or overall scores.
func CalculatePhase1CompetencyResult(inputs []CompetencyScoreInput) (CompetencyScoreResult, error) {
	profile := NormalizePhase1CompetencyConfiguration()
	if len(inputs) != profile.DimensionQuestionCount {
		return CompetencyScoreResult{}, errors.New("phase-1 scoring requires exactly 80 dimension questions")
	}
	byID := make(map[string]*CompetencyDimensionScore, len(profile.DimensionIDs))
	for index, id := range profile.DimensionIDs {
		byID[id] = &CompetencyDimensionScore{DimensionID: id, DisplayOrder: index + 1}
	}
	result := CompetencyScoreResult{
		Dimensions:              make([]CompetencyDimensionScore, 0, len(profile.DimensionIDs)),
		TotalQuestionCount:      len(inputs),
		OverallScore:            nil,
		EvaluationAverage:       nil,
		EffectiveDimensionCount: 0,
		IsComplete:              true,
	}
	for _, input := range inputs {
		if input.QuestionType != CompetencyQuestionTypeDimension {
			return CompetencyScoreResult{}, errors.New("phase-1 dimension scoring cannot include validity questions")
		}
		dimension, exists := byID[input.DimensionID]
		if !exists {
			return CompetencyScoreResult{}, errors.New("unknown phase-1 competency dimension")
		}
		if input.DisplayOrder != dimension.DisplayOrder {
			return CompetencyScoreResult{}, errors.New("phase-1 competency dimension order mismatch")
		}
		if dimension.DimensionCode == "" {
			dimension.DimensionCode = input.DimensionCode
			dimension.DimensionName = input.DimensionName
		} else if dimension.DimensionCode != input.DimensionCode || dimension.DimensionName != input.DimensionName {
			return CompetencyScoreResult{}, errors.New("phase-1 competency dimension metadata mismatch")
		}
		dimension.TotalQuestionCount++
		if !input.Answered {
			result.IsComplete = false
			continue
		}
		if input.FinalScore < 1 || input.FinalScore > 5 {
			return CompetencyScoreResult{}, errors.New("competency final score must be between 1 and 5")
		}
		dimension.AnsweredQuestionCount++
		dimension.ScoreSum += input.FinalScore
		result.AnsweredQuestionCount++
	}

	for _, id := range profile.DimensionIDs {
		dimension := byID[id]
		if dimension.TotalQuestionCount != 8 {
			return CompetencyScoreResult{}, errors.New("each phase-1 dimension requires exactly 8 questions")
		}
		dimension.IsComplete = dimension.AnsweredQuestionCount == 8
		if !dimension.IsComplete {
			result.IsComplete = false
		} else {
			dimension.Score = big.NewRat(int64(dimension.ScoreSum), 8)
			level, err := Phase1DimensionLevelForScore(dimension.Score)
			if err != nil {
				return CompetencyScoreResult{}, err
			}
			dimension.Level = level
			result.EffectiveDimensionCount++
		}
		result.Dimensions = append(result.Dimensions, *dimension)
	}
	if !result.IsComplete {
		return result, nil
	}
	result.OverallScore = new(big.Rat)
	for index := range result.Dimensions {
		result.OverallScore.Add(result.OverallScore, result.Dimensions[index].Score)
	}
	level, err := Phase1OverallLevelForScore(result.OverallScore)
	if err != nil {
		return CompetencyScoreResult{}, err
	}
	result.EvaluationLevel = level
	return result, nil
}

// CalculatePhase1GroupResults aggregates the fixed five general-ability and
// five psychological-quality dimensions. A group emits a formal score only
// when all five child dimensions are complete.
func CalculatePhase1GroupResults(dimensions []CompetencyDimensionScore) ([]CompetencyGroupScore, error) {
	profile := NormalizePhase1CompetencyConfiguration()
	if len(dimensions) != len(profile.DimensionIDs) {
		return nil, errors.New("phase-1 group aggregation requires exactly 10 dimensions")
	}
	byID := make(map[string]CompetencyDimensionScore, len(dimensions))
	for _, dimension := range dimensions {
		if _, exists := byID[dimension.DimensionID]; exists {
			return nil, errors.New("duplicate phase-1 dimension result")
		}
		byID[dimension.DimensionID] = dimension
	}
	for index, id := range profile.DimensionIDs {
		dimension, exists := byID[id]
		if !exists {
			return nil, errors.New("missing phase-1 dimension result")
		}
		if dimension.DisplayOrder != index+1 || dimension.TotalQuestionCount != 8 || dimension.AnsweredQuestionCount < 0 || dimension.AnsweredQuestionCount > 8 {
			return nil, errors.New("invalid phase-1 dimension result metadata")
		}
		if dimension.IsComplete {
			if dimension.AnsweredQuestionCount != 8 || dimension.Score == nil {
				return nil, errors.New("complete phase-1 dimension requires a score")
			}
			level, err := Phase1DimensionLevelForScore(dimension.Score)
			if err != nil || dimension.Level != level {
				return nil, errors.New("phase-1 dimension score and level mismatch")
			}
		} else if dimension.AnsweredQuestionCount == 8 || dimension.Score != nil || dimension.Level != "" {
			return nil, errors.New("incomplete phase-1 dimension cannot have a formal score")
		}
	}

	definitions := []struct {
		code, name string
		order      int
		ids        []string
	}{
		{CompetencyPhase1GroupGeneralAbility, "通用能力", 1, profile.DimensionIDs[:5]},
		{CompetencyPhase1GroupPsychologicalQuality, "心理素养", 2, profile.DimensionIDs[5:]},
	}
	groups := make([]CompetencyGroupScore, 0, len(definitions))
	for _, definition := range definitions {
		group := CompetencyGroupScore{
			GroupCode:           definition.code,
			GroupName:           definition.name,
			DisplayOrder:        definition.order,
			ChildDimensionIDs:   append([]string(nil), definition.ids...),
			TotalDimensionCount: len(definition.ids),
			IsComplete:          true,
		}
		sum := new(big.Rat)
		for _, id := range definition.ids {
			dimension := byID[id]
			group.TotalQuestionCount += dimension.TotalQuestionCount
			group.AnsweredQuestionCount += dimension.AnsweredQuestionCount
			if !dimension.IsComplete {
				group.IsComplete = false
				continue
			}
			group.EffectiveDimensionCount++
			sum.Add(sum, dimension.Score)
		}
		if group.IsComplete {
			group.Score = new(big.Rat).Quo(sum, big.NewRat(int64(group.TotalDimensionCount), 1))
			level, err := Phase1DimensionLevelForScore(group.Score)
			if err != nil {
				return nil, err
			}
			group.Level = level
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// CalculatePhase1ValidityResult sums the ten original raw values. Validity
// questions are always forward; they never enter dimension/group/overall
// scores. Incomplete timeout results have no formal validity score.
func CalculatePhase1ValidityResult(inputs []Phase1ValidityInput) (Phase1ValidityResult, error) {
	profile := NormalizePhase1CompetencyConfiguration()
	if len(inputs) != profile.ValidityQuestionCount {
		return Phase1ValidityResult{}, errors.New("phase-1 validity requires exactly 10 questions")
	}
	result := Phase1ValidityResult{
		TotalQuestionCount: len(inputs),
		Status:             CompetencyPhase1ValidityIncomplete,
	}
	seenCodes := make(map[string]struct{}, len(inputs))
	score := 0
	for index, input := range inputs {
		if input.QuestionCode == "" {
			return Phase1ValidityResult{}, errors.New("phase-1 validity question code is required")
		}
		if _, exists := seenCodes[input.QuestionCode]; exists {
			return Phase1ValidityResult{}, errors.New("duplicate phase-1 validity question")
		}
		seenCodes[input.QuestionCode] = struct{}{}
		if input.Order != index+1 {
			return Phase1ValidityResult{}, errors.New("phase-1 validity order mismatch")
		}
		if input.QuestionType != CompetencyQuestionTypeValidity || input.Direction != CompetencyDirectionForward {
			return Phase1ValidityResult{}, errors.New("invalid phase-1 validity question metadata")
		}
		if !input.Answered {
			continue
		}
		if input.RawValue < 1 || input.RawValue > 5 {
			return Phase1ValidityResult{}, errors.New("phase-1 validity raw value must be between 1 and 5")
		}
		result.AnsweredQuestionCount++
		score += input.RawValue
	}
	if result.AnsweredQuestionCount != result.TotalQuestionCount {
		return result, nil
	}
	result.IsComplete = true
	result.Score = &score
	if score <= 35 {
		result.Status = CompetencyPhase1ValidityGood
	} else {
		result.Status = CompetencyPhase1ValidityQuestionable
	}
	return result, nil
}

func Phase1DimensionLevelForScore(score *big.Rat) (string, error) {
	if score == nil || score.Cmp(big.NewRat(1, 1)) < 0 || score.Cmp(big.NewRat(5, 1)) > 0 {
		return "", errors.New("phase-1 dimension score must be between 1 and 5")
	}
	switch {
	case score.Cmp(big.NewRat(17, 10)) <= 0:
		return CompetencyPhase1LevelL1, nil
	case score.Cmp(big.NewRat(27, 10)) <= 0:
		return CompetencyPhase1LevelL2, nil
	case score.Cmp(big.NewRat(35, 10)) <= 0:
		return CompetencyPhase1LevelL3, nil
	case score.Cmp(big.NewRat(43, 10)) <= 0:
		return CompetencyPhase1LevelL4, nil
	default:
		return CompetencyPhase1LevelL5, nil
	}
}

func Phase1OverallLevelForScore(score *big.Rat) (string, error) {
	if score == nil || score.Cmp(big.NewRat(10, 1)) < 0 || score.Cmp(big.NewRat(50, 1)) > 0 {
		return "", errors.New("phase-1 overall score must be between 10 and 50")
	}
	switch {
	case score.Cmp(big.NewRat(45, 1)) >= 0:
		return CompetencyPhase1OverallExcellent, nil
	case score.Cmp(big.NewRat(40, 1)) >= 0:
		return CompetencyPhase1OverallGood, nil
	case score.Cmp(big.NewRat(65, 2)) >= 0:
		return CompetencyPhase1OverallQualified, nil
	case score.Cmp(big.NewRat(25, 1)) >= 0:
		return CompetencyPhase1OverallWeak, nil
	default:
		return CompetencyPhase1OverallNotQualified, nil
	}
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
