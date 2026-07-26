package service

import (
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

func assertRat(t *testing.T, got *big.Rat, numerator, denominator int64) {
	t.Helper()
	want := big.NewRat(numerator, denominator)
	if got == nil || got.Cmp(want) != 0 {
		t.Fatalf("score = %v, want %s", got, want.RatString())
	}
}
