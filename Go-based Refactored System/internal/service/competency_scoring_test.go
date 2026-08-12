package service

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

func TestCalculateCompetencyQuestionScore(t *testing.T) {
	tests := []struct {
		name      string
		raw       int
		direction string
		want      int
		wantErr   bool
	}{
		{"forward 1", 1, CompetencyDirectionForward, 1, false},
		{"forward 2", 2, CompetencyDirectionForward, 2, false},
		{"forward 3", 3, CompetencyDirectionForward, 3, false},
		{"forward 4", 4, CompetencyDirectionForward, 4, false},
		{"forward 5", 5, CompetencyDirectionForward, 5, false},
		{"reverse 1", 1, CompetencyDirectionReverse, 5, false},
		{"reverse 2", 2, CompetencyDirectionReverse, 4, false},
		{"reverse 3", 3, CompetencyDirectionReverse, 3, false},
		{"reverse 4", 4, CompetencyDirectionReverse, 2, false},
		{"reverse 5", 5, CompetencyDirectionReverse, 1, false},
		{"raw zero", 0, CompetencyDirectionForward, 0, true},
		{"raw six", 6, CompetencyDirectionForward, 0, true},
		{"empty direction", 3, "", 0, true},
		{"unknown direction", 3, "sideways", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateCompetencyQuestionScore(tt.raw, tt.direction)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("score = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateCompetencyResult_CompleteAndIncomplete(t *testing.T) {
	inputs := []CompetencyScoreInput{
		{DimensionID: "d1", DimensionCode: "D01", DimensionName: "沟通表达", DisplayOrder: 1, Answered: true, FinalScore: 5},
		{DimensionID: "d1", DimensionCode: "D01", DimensionName: "沟通表达", DisplayOrder: 1, Answered: true, FinalScore: 4},
		{DimensionID: "d1", DimensionCode: "D01", DimensionName: "沟通表达", DisplayOrder: 1, Answered: true, FinalScore: 3},
		{DimensionID: "d2", DimensionCode: "D02", DimensionName: "人际交往", DisplayOrder: 2, Answered: true, FinalScore: 2},
		{DimensionID: "d2", DimensionCode: "D02", DimensionName: "人际交往", DisplayOrder: 2, Answered: false},
		{DimensionID: "d3", DimensionCode: "D03", DimensionName: "数字应用", DisplayOrder: 3, Answered: false},
	}

	got, err := CalculateCompetencyResult(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Dimensions) != 3 {
		t.Fatalf("dimension count = %d, want 3", len(got.Dimensions))
	}

	d1 := got.Dimensions[0]
	if d1.TotalQuestionCount != 3 || d1.AnsweredQuestionCount != 3 || d1.ScoreSum != 12 || !d1.IsComplete {
		t.Fatalf("unexpected complete dimension: %+v", d1)
	}
	assertRat(t, d1.Score, 4, 1)
	if d1.Level != CompetencyLevelHigh {
		t.Fatalf("d1 level = %q", d1.Level)
	}

	d2 := got.Dimensions[1]
	if d2.TotalQuestionCount != 2 || d2.AnsweredQuestionCount != 1 || d2.ScoreSum != 2 || d2.IsComplete {
		t.Fatalf("unexpected incomplete dimension: %+v", d2)
	}
	assertRat(t, d2.Score, 2, 1)

	d3 := got.Dimensions[2]
	if d3.Score != nil || d3.Level != "" || d3.IsComplete {
		t.Fatalf("zero-answer dimension must have no score: %+v", d3)
	}

	if got.TotalQuestionCount != 6 || got.AnsweredQuestionCount != 4 || got.EffectiveDimensionCount != 2 || got.IsComplete {
		t.Fatalf("unexpected overall counts: %+v", got)
	}
	assertRat(t, got.OverallScore, 6, 1)
	assertRat(t, got.EvaluationAverage, 3, 1)
	if got.EvaluationLevel != CompetencyLevelGood {
		t.Fatalf("evaluation level = %q", got.EvaluationLevel)
	}
}

func TestCalculateCompetencyResult_NoAnsweredQuestions(t *testing.T) {
	got, err := CalculateCompetencyResult([]CompetencyScoreInput{
		{DimensionID: "d1", DimensionCode: "D01", DisplayOrder: 1, Answered: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRat(t, got.OverallScore, 0, 1)
	if got.EvaluationAverage != nil || got.EvaluationLevel != "" || got.EffectiveDimensionCount != 0 {
		t.Fatalf("unexpected no-answer result: %+v", got)
	}
}

func TestCalculateCompetencyResult_OrderIndependent(t *testing.T) {
	left := []CompetencyScoreInput{
		{DimensionID: "d2", DimensionCode: "D02", DisplayOrder: 2, Answered: true, FinalScore: 2},
		{DimensionID: "d1", DimensionCode: "D01", DisplayOrder: 1, Answered: true, FinalScore: 5},
		{DimensionID: "d1", DimensionCode: "D01", DisplayOrder: 1, Answered: true, FinalScore: 3},
	}
	right := []CompetencyScoreInput{left[2], left[0], left[1]}

	gotLeft, err := CalculateCompetencyResult(left)
	if err != nil {
		t.Fatal(err)
	}
	gotRight, err := CalculateCompetencyResult(right)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotLeft, gotRight) {
		t.Fatalf("results differ by display order:\nleft=%+v\nright=%+v", gotLeft, gotRight)
	}
}

func TestCompetencyLevelForScore_Boundaries(t *testing.T) {
	tests := []struct {
		numerator, denominator int64
		want                   string
		wantErr                bool
	}{
		{1, 1, CompetencyLevelLow, false},
		{199, 100, CompetencyLevelLow, false},
		{2, 1, CompetencyLevelAverage, false},
		{3, 1, CompetencyLevelGood, false},
		{4, 1, CompetencyLevelHigh, false},
		{5, 1, CompetencyLevelHigh, false},
		{99, 100, "", true},
		{501, 100, "", true},
	}
	for _, tt := range tests {
		score := big.NewRat(tt.numerator, tt.denominator)
		got, err := CompetencyLevelForScore(score)
		if (err != nil) != tt.wantErr {
			t.Fatalf("score %s error=%v wantErr=%v", score.RatString(), err, tt.wantErr)
		}
		if got != tt.want {
			t.Fatalf("score %s level=%q want=%q", score.RatString(), got, tt.want)
		}
	}
}

func TestFormatCompetencyScore(t *testing.T) {
	if got := FormatCompetencyScore(big.NewRat(10, 3)); got != "3.33" {
		t.Fatalf("formatted score = %q", got)
	}
	if got := FormatCompetencyScore(nil); got != "" {
		t.Fatalf("nil formatted score = %q", got)
	}
}

func TestPhase1DimensionLevelForScore_Boundaries(t *testing.T) {
	tests := []struct {
		numerator, denominator int64
		want                   string
		wantErr                bool
	}{
		{1, 1, CompetencyPhase1LevelL1, false},
		{17, 10, CompetencyPhase1LevelL1, false},
		{171, 100, CompetencyPhase1LevelL2, false},
		{27, 10, CompetencyPhase1LevelL2, false},
		{271, 100, CompetencyPhase1LevelL3, false},
		{35, 10, CompetencyPhase1LevelL3, false},
		{351, 100, CompetencyPhase1LevelL4, false},
		{43, 10, CompetencyPhase1LevelL4, false},
		{431, 100, CompetencyPhase1LevelL5, false},
		{5, 1, CompetencyPhase1LevelL5, false},
		{99, 100, "", true},
		{501, 100, "", true},
	}
	for _, test := range tests {
		score := big.NewRat(test.numerator, test.denominator)
		got, err := Phase1DimensionLevelForScore(score)
		if (err != nil) != test.wantErr || got != test.want {
			t.Fatalf("score=%s level=%q error=%v want=%q wantErr=%v", score, got, err, test.want, test.wantErr)
		}
	}
}

func TestPhase1OverallLevelForScore_Boundaries(t *testing.T) {
	tests := []struct {
		numerator, denominator int64
		want                   string
		wantErr                bool
	}{
		{10, 1, CompetencyPhase1OverallNotQualified, false},
		{2499, 100, CompetencyPhase1OverallNotQualified, false},
		{25, 1, CompetencyPhase1OverallWeak, false},
		{3249, 100, CompetencyPhase1OverallWeak, false},
		{65, 2, CompetencyPhase1OverallQualified, false},
		{3999, 100, CompetencyPhase1OverallQualified, false},
		{40, 1, CompetencyPhase1OverallGood, false},
		{4499, 100, CompetencyPhase1OverallGood, false},
		{45, 1, CompetencyPhase1OverallExcellent, false},
		{50, 1, CompetencyPhase1OverallExcellent, false},
		{999, 100, "", true},
		{5001, 100, "", true},
	}
	for _, test := range tests {
		score := big.NewRat(test.numerator, test.denominator)
		got, err := Phase1OverallLevelForScore(score)
		if (err != nil) != test.wantErr || got != test.want {
			t.Fatalf("score=%s level=%q error=%v want=%q wantErr=%v", score, got, err, test.want, test.wantErr)
		}
	}
}

func TestCalculatePhase1CompetencyResult_ExactTenDimensionScore(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	inputs := make([]CompetencyScoreInput, 0, 80)
	for dimensionIndex, id := range profile.DimensionIDs {
		for item := 1; item <= 8; item++ {
			inputs = append(inputs, CompetencyScoreInput{
				DimensionID: id, DimensionCode: id, DimensionName: id, DisplayOrder: dimensionIndex + 1,
				QuestionType: CompetencyQuestionTypeDimension, Answered: true, FinalScore: dimensionIndex%5 + 1,
			})
		}
	}
	result, err := CalculatePhase1CompetencyResult(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsComplete || result.TotalQuestionCount != 80 || result.AnsweredQuestionCount != 80 || len(result.Dimensions) != 10 {
		t.Fatalf("unexpected phase-1 result=%+v", result)
	}
	assertRat(t, result.OverallScore, 30, 1)
	if result.EvaluationAverage != nil || result.EvaluationLevel != CompetencyPhase1OverallWeak {
		t.Fatalf("overall average/level=%v/%q", result.EvaluationAverage, result.EvaluationLevel)
	}
	for index, dimension := range result.Dimensions {
		if dimension.DimensionID != profile.DimensionIDs[index] || dimension.TotalQuestionCount != 8 || dimension.AnsweredQuestionCount != 8 {
			t.Fatalf("dimension[%d]=%+v", index, dimension)
		}
	}
}

func TestCalculatePhase1CompetencyResult_OrderIndependent(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	inputs := make([]CompetencyScoreInput, 0, 80)
	for index := len(profile.DimensionIDs) - 1; index >= 0; index-- {
		for item := 1; item <= 8; item++ {
			inputs = append(inputs, CompetencyScoreInput{
				DimensionID: profile.DimensionIDs[index], DisplayOrder: index + 1,
				QuestionType: CompetencyQuestionTypeDimension, Answered: true, FinalScore: 5,
			})
		}
	}
	result, err := CalculatePhase1CompetencyResult(inputs)
	if err != nil {
		t.Fatal(err)
	}
	assertRat(t, result.OverallScore, 50, 1)
	if result.EvaluationLevel != CompetencyPhase1OverallExcellent {
		t.Fatalf("level=%q", result.EvaluationLevel)
	}
	for index, dimension := range result.Dimensions {
		if dimension.DimensionID != profile.DimensionIDs[index] {
			t.Fatalf("dimension order=%v", result.Dimensions)
		}
	}
}

func TestCalculatePhase1CompetencyResult_IncompleteHasNoFormalScores(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	valid := make([]CompetencyScoreInput, 0, 80)
	for index, id := range profile.DimensionIDs {
		for item := 1; item <= 8; item++ {
			valid = append(valid, CompetencyScoreInput{DimensionID: id, DisplayOrder: index + 1, QuestionType: CompetencyQuestionTypeDimension, Answered: true, FinalScore: 3})
		}
	}
	incomplete := append([]CompetencyScoreInput(nil), valid...)
	incomplete[0].Answered = false
	result, err := CalculatePhase1CompetencyResult(incomplete)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsComplete || result.OverallScore != nil || result.EvaluationLevel != "" || result.Dimensions[0].Score != nil || result.Dimensions[0].Level != "" {
		t.Fatalf("incomplete phase-1 result exposes formal score=%+v", result)
	}
}

func TestCalculatePhase1CompetencyResult_RejectsMalformedOrMixedInput(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	valid := make([]CompetencyScoreInput, 0, 80)
	for index, id := range profile.DimensionIDs {
		for item := 1; item <= 8; item++ {
			valid = append(valid, CompetencyScoreInput{DimensionID: id, DisplayOrder: index + 1, QuestionType: CompetencyQuestionTypeDimension, Answered: true, FinalScore: 3})
		}
	}
	tests := []struct {
		name   string
		mutate func([]CompetencyScoreInput) []CompetencyScoreInput
	}{
		{"missing row", func(rows []CompetencyScoreInput) []CompetencyScoreInput { return rows[:79] }},
		{"validity row", func(rows []CompetencyScoreInput) []CompetencyScoreInput {
			rows[0].QuestionType = CompetencyQuestionTypeValidity
			return rows
		}},
		{"unknown dimension", func(rows []CompetencyScoreInput) []CompetencyScoreInput { rows[0].DimensionID = "other"; return rows }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := append([]CompetencyScoreInput(nil), valid...)
			if _, err := CalculatePhase1CompetencyResult(test.mutate(rows)); err == nil {
				t.Fatal("invalid phase-1 scoring input must be rejected")
			}
		})
	}
}

func TestCalculatePhase1GroupResults_ExactScoresAndOrder(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	dimensions := make([]CompetencyDimensionScore, 0, 10)
	for index, id := range profile.DimensionIDs {
		score := big.NewRat(int64(index%5+1), 1)
		level, err := Phase1DimensionLevelForScore(score)
		if err != nil {
			t.Fatal(err)
		}
		dimensions = append(dimensions, CompetencyDimensionScore{
			DimensionID: id, DisplayOrder: index + 1, TotalQuestionCount: 8, AnsweredQuestionCount: 8,
			Score: score, Level: level, IsComplete: true,
		})
	}
	for left, right := 0, len(dimensions)-1; left < right; left, right = left+1, right-1 {
		dimensions[left], dimensions[right] = dimensions[right], dimensions[left]
	}

	groups, err := CalculatePhase1GroupResults(dimensions)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("groups=%+v", groups)
	}
	for index, expected := range []struct {
		code, name string
		order      int
	}{
		{CompetencyPhase1GroupGeneralAbility, "通用能力", 1},
		{CompetencyPhase1GroupPsychologicalQuality, "心理素养", 2},
	} {
		group := groups[index]
		if group.GroupCode != expected.code || group.GroupName != expected.name || group.DisplayOrder != expected.order ||
			group.TotalDimensionCount != 5 || group.EffectiveDimensionCount != 5 || group.TotalQuestionCount != 40 ||
			group.AnsweredQuestionCount != 40 || !group.IsComplete {
			t.Fatalf("group[%d]=%+v", index, group)
		}
		assertRat(t, group.Score, 3, 1)
		if group.Level != CompetencyPhase1LevelL3 {
			t.Fatalf("group level=%q", group.Level)
		}
	}
}

func TestCalculatePhase1GroupResults_IncompleteGroupHasNoFormalScore(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	dimensions := make([]CompetencyDimensionScore, 0, 10)
	for index, id := range profile.DimensionIDs {
		dimensions = append(dimensions, CompetencyDimensionScore{
			DimensionID: id, DisplayOrder: index + 1, TotalQuestionCount: 8, AnsweredQuestionCount: 8,
			Score: big.NewRat(4, 1), Level: CompetencyPhase1LevelL4, IsComplete: true,
		})
	}
	dimensions[0].AnsweredQuestionCount = 7
	dimensions[0].Score = nil
	dimensions[0].Level = ""
	dimensions[0].IsComplete = false

	groups, err := CalculatePhase1GroupResults(dimensions)
	if err != nil {
		t.Fatal(err)
	}
	general := groups[0]
	if general.IsComplete || general.Score != nil || general.Level != "" || general.EffectiveDimensionCount != 4 ||
		general.TotalQuestionCount != 40 || general.AnsweredQuestionCount != 39 {
		t.Fatalf("incomplete general group=%+v", general)
	}
	psychological := groups[1]
	if !psychological.IsComplete || psychological.EffectiveDimensionCount != 5 {
		t.Fatalf("complete psychological group=%+v", psychological)
	}
	assertRat(t, psychological.Score, 4, 1)
}

func TestCalculatePhase1GroupResults_UsesExactDimensionBoundaries(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	for _, test := range []struct {
		score *big.Rat
		level string
	}{
		{big.NewRat(17, 10), CompetencyPhase1LevelL1},
		{big.NewRat(27, 10), CompetencyPhase1LevelL2},
		{big.NewRat(35, 10), CompetencyPhase1LevelL3},
		{big.NewRat(43, 10), CompetencyPhase1LevelL4},
	} {
		dimensions := make([]CompetencyDimensionScore, 0, 10)
		for index, id := range profile.DimensionIDs {
			dimensions = append(dimensions, CompetencyDimensionScore{
				DimensionID: id, DisplayOrder: index + 1, TotalQuestionCount: 8, AnsweredQuestionCount: 8,
				Score: new(big.Rat).Set(test.score), Level: test.level, IsComplete: true,
			})
		}
		groups, err := CalculatePhase1GroupResults(dimensions)
		if err != nil {
			t.Fatal(err)
		}
		if groups[0].Level != test.level || groups[1].Level != test.level || groups[0].Score.Cmp(test.score) != 0 {
			t.Fatalf("score=%s groups=%+v", test.score, groups)
		}
	}
}

func TestCalculatePhase1GroupResults_RejectsMalformedInput(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	valid := make([]CompetencyDimensionScore, 0, 10)
	for index, id := range profile.DimensionIDs {
		valid = append(valid, CompetencyDimensionScore{
			DimensionID: id, DisplayOrder: index + 1, TotalQuestionCount: 8, AnsweredQuestionCount: 8,
			Score: big.NewRat(3, 1), Level: CompetencyPhase1LevelL3, IsComplete: true,
		})
	}
	for name, mutate := range map[string]func([]CompetencyDimensionScore) []CompetencyDimensionScore{
		"missing":   func(rows []CompetencyDimensionScore) []CompetencyDimensionScore { return rows[:9] },
		"duplicate": func(rows []CompetencyDimensionScore) []CompetencyDimensionScore { rows[9] = rows[0]; return rows },
		"unknown": func(rows []CompetencyDimensionScore) []CompetencyDimensionScore {
			rows[0].DimensionID = "other"
			return rows
		},
		"order": func(rows []CompetencyDimensionScore) []CompetencyDimensionScore {
			rows[0].DisplayOrder = 2
			return rows
		},
		"complete nil score": func(rows []CompetencyDimensionScore) []CompetencyDimensionScore { rows[0].Score = nil; return rows },
	} {
		t.Run(name, func(t *testing.T) {
			rows := append([]CompetencyDimensionScore(nil), valid...)
			if _, err := CalculatePhase1GroupResults(mutate(rows)); err == nil {
				t.Fatal("malformed phase-1 group input must be rejected")
			}
		})
	}
}

func TestCalculatePhase1ValidityResult_Boundary35And36(t *testing.T) {
	for _, test := range []struct {
		name   string
		values []int
		score  int
		status string
	}{
		{"score 35 is good", []int{3, 3, 3, 3, 3, 4, 4, 4, 4, 4}, 35, CompetencyPhase1ValidityGood},
		{"score 36 is questionable", []int{3, 3, 3, 3, 4, 4, 4, 4, 4, 4}, 36, CompetencyPhase1ValidityQuestionable},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := CalculatePhase1ValidityResult(phase1ValidityInputs(test.values, -1))
			if err != nil {
				t.Fatal(err)
			}
			if !result.IsComplete || result.TotalQuestionCount != 10 || result.AnsweredQuestionCount != 10 || result.Score == nil || *result.Score != test.score || result.Status != test.status {
				t.Fatalf("validity=%+v", result)
			}
		})
	}
}

func TestCalculatePhase1ValidityResult_IncompleteHasNoScore(t *testing.T) {
	inputs := phase1ValidityInputs([]int{5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, 4)
	inputs[4].RawValue = 0
	result, err := CalculatePhase1ValidityResult(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsComplete || result.TotalQuestionCount != 10 || result.AnsweredQuestionCount != 9 || result.Score != nil || result.Status != CompetencyPhase1ValidityIncomplete {
		t.Fatalf("incomplete validity=%+v", result)
	}
}

func TestCalculatePhase1ValidityResult_Extremes(t *testing.T) {
	for value, want := range map[int]string{1: CompetencyPhase1ValidityGood, 5: CompetencyPhase1ValidityQuestionable} {
		values := make([]int, 10)
		for index := range values {
			values[index] = value
		}
		result, err := CalculatePhase1ValidityResult(phase1ValidityInputs(values, -1))
		if err != nil {
			t.Fatal(err)
		}
		if result.Score == nil || *result.Score != value*10 || result.Status != want {
			t.Fatalf("value=%d result=%+v", value, result)
		}
	}
}

func TestCalculatePhase1ValidityResult_RejectsMalformedInput(t *testing.T) {
	valid := phase1ValidityInputs([]int{3, 3, 3, 3, 3, 3, 3, 3, 3, 3}, -1)
	for name, mutate := range map[string]func([]Phase1ValidityInput) []Phase1ValidityInput{
		"count": func(rows []Phase1ValidityInput) []Phase1ValidityInput { return rows[:9] },
		"type": func(rows []Phase1ValidityInput) []Phase1ValidityInput {
			rows[0].QuestionType = CompetencyQuestionTypeDimension
			return rows
		},
		"raw low":  func(rows []Phase1ValidityInput) []Phase1ValidityInput { rows[0].RawValue = 0; return rows },
		"raw high": func(rows []Phase1ValidityInput) []Phase1ValidityInput { rows[0].RawValue = 6; return rows },
		"direction": func(rows []Phase1ValidityInput) []Phase1ValidityInput {
			rows[0].Direction = CompetencyDirectionReverse
			return rows
		},
		"identity": func(rows []Phase1ValidityInput) []Phase1ValidityInput { rows[0].QuestionCode = ""; return rows },
		"duplicate": func(rows []Phase1ValidityInput) []Phase1ValidityInput {
			rows[1].QuestionCode = rows[0].QuestionCode
			return rows
		},
		"order": func(rows []Phase1ValidityInput) []Phase1ValidityInput { rows[0].Order = 2; return rows },
	} {
		t.Run(name, func(t *testing.T) {
			rows := append([]Phase1ValidityInput(nil), valid...)
			if _, err := CalculatePhase1ValidityResult(mutate(rows)); err == nil {
				t.Fatal("malformed phase-1 validity input must be rejected")
			}
		})
	}
}

func phase1ValidityInputs(values []int, unansweredIndex int) []Phase1ValidityInput {
	rows := make([]Phase1ValidityInput, len(values))
	for index, value := range values {
		rows[index] = Phase1ValidityInput{
			QuestionCode: fmt.Sprintf("P1-VAL-Q%02d", index+1), Order: index + 1,
			QuestionType: CompetencyQuestionTypeValidity, Direction: CompetencyDirectionForward,
			Answered: index != unansweredIndex, RawValue: value,
		}
	}
	return rows
}

func assertRat(t *testing.T, got *big.Rat, numerator, denominator int64) {
	t.Helper()
	want := big.NewRat(numerator, denominator)
	if got == nil || got.Cmp(want) != 0 {
		t.Fatalf("score = %v, want %s", got, want.RatString())
	}
}
