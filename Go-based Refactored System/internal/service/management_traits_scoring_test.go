package service

import (
	"math/big"
	"reflect"
	"testing"
)

// Independent fixture: explicit V terms and reverse terms from standScore2;
// identities, order, modules and exact norms from design sections 3–5.
type managementTraitsFixtureDimension struct {
	key, name, module string
	norm              *big.Rat
	numbers, reverse  []int
	allFourSum        int
	allFourScore      *big.Rat
}

func managementTraitsFixture() []managementTraitsFixtureDimension {
	return []managementTraitsFixtureDimension{
		{"self_confidence", "自信心", "self", big.NewRat(115, 2), []int{6, 19, 33, 47, 60, 66, 78, 92, 101, 107, 120, 133}, []int{120, 133}, 44, big.NewRat(200, 3)},
		{"emotional_stability", "情绪稳定性", "self", big.NewRat(55, 1), []int{10, 24, 38, 52, 68, 82, 96, 111, 124, 137}, []int{111, 124, 137}, 34, big.NewRat(60, 1)},
		{"self_discipline", "自律性", "self", big.NewRat(225, 4), []int{11, 25, 39, 53, 69, 84, 97, 112, 125, 138}, []int{25, 69, 97, 112}, 32, big.NewRat(55, 1)},
		{"sociality", "社会性", "interpersonal", big.NewRat(205, 4), []int{1, 14, 28, 42, 56, 72, 87, 102, 115, 128}, []int{28, 42, 56, 72, 87, 102, 115, 128}, 24, big.NewRat(35, 1)},
		{"leadership", "领导性", "interpersonal", big.NewRat(105, 2), []int{3, 16, 31, 44, 58, 74, 89, 104, 117, 130}, []int{117, 130}, 36, big.NewRat(65, 1)},
		{"interpersonal_sensitivity", "人际敏感性", "interpersonal", big.NewRat(50, 1), []int{5, 18, 32, 46, 50, 61, 65, 76, 83, 91, 106, 119, 132}, []int{106, 119, 132}, 46, big.NewRat(825, 13)},
		{"cooperation", "合作性", "interpersonal", big.NewRat(115, 2), []int{13, 27, 41, 55, 71, 86, 99, 114, 127, 140}, []int{114, 140}, 36, big.NewRat(65, 1)},
		{"planning", "计划性", "task", big.NewRat(215, 4), []int{4, 17, 21, 30, 45, 59, 75, 81, 90, 105, 118, 131}, []int{131}, 46, big.NewRat(425, 6)},
		{"responsibility", "责任心", "task", big.NewRat(235, 4), []int{7, 20, 34, 48, 63, 77, 94, 108, 121, 134}, []int{94, 108, 121, 134}, 32, big.NewRat(55, 1)},
		{"decisiveness", "决断性", "task", big.NewRat(215, 4), []int{12, 26, 35, 40, 54, 62, 70, 85, 98, 113, 126, 139}, []int{12}, 46, big.NewRat(425, 6)},
		{"proactiveness", "进取性", "development", big.NewRat(55, 1), []int{2, 15, 29, 43, 57, 73, 88, 103, 116, 129}, []int{57, 116}, 36, big.NewRat(65, 1)},
		{"learning", "学习力", "development", big.NewRat(215, 4), []int{8, 22, 36, 49, 64, 79, 93, 100, 109, 122, 135}, []int{8, 64, 100, 135}, 36, big.NewRat(625, 11)},
		{"innovation", "创新性", "development", big.NewRat(50, 1), []int{9, 23, 37, 51, 67, 80, 95, 110, 123, 136}, []int{9, 51, 80, 136}, 32, big.NewRat(55, 1)},
	}
}

func managementTraitsFixtureIsReverse(d managementTraitsFixtureDimension, number int) bool {
	for _, n := range d.reverse {
		if n == number {
			return true
		}
	}
	return false
}

func managementTraitsFixtureAnswers(raw int) []ManagementTraitsAnswer {
	inputs := make([]ManagementTraitsAnswer, 140)
	for i := range inputs {
		inputs[i] = ManagementTraitsAnswer{Number: i + 1, Answered: true, Raw: raw}
	}
	return inputs
}

// Build raw inputs from a desired final value without reading the production catalog.
func managementTraitsFixtureFinalAnswers(final int) []ManagementTraitsAnswer {
	inputs := managementTraitsFixtureAnswers(final)
	for _, d := range managementTraitsFixture() {
		for _, n := range d.reverse {
			inputs[n-1].Raw = 6 - final
		}
	}
	return inputs
}

func managementTraitsAssertRat(t *testing.T, label string, got, want *big.Rat) {
	t.Helper()
	if got == nil || want == nil || got.Cmp(want) != 0 {
		t.Fatalf("%s = %v, want exact %v", label, got, want)
	}
}

func managementTraitsAssertAudit(t *testing.T, questionnaire string, inputs []ManagementTraitsAnswer, got ManagementTraitsResult) {
	t.Helper()
	fixture := managementTraitsFixture()
	if got.Questionnaire != questionnaire || got.TotalQuestionCount != 140 || len(got.Dimensions) != 13 || len(got.Modules) != 4 {
		t.Fatalf("unexpected result identity/counts: %+v", got)
	}
	byNumber := make(map[int]ManagementTraitsAnswer, len(inputs))
	answered := 0
	for _, input := range inputs {
		byNumber[input.Number] = input
		if input.Answered {
			answered++
		}
	}
	if got.AnsweredQuestionCount != answered || got.IsComplete != (answered == 140) {
		t.Fatalf("completion = %d/%d, complete %v", got.AnsweredQuestionCount, got.TotalQuestionCount, got.IsComplete)
	}
	for i, d := range fixture {
		score := got.Dimensions[i]
		if score.Key != d.key || score.Name != d.name || score.Module != d.module || score.Order != i+1 || score.QuestionCount != len(d.numbers) {
			t.Fatalf("dimension %d identity/count = %+v", i+1, score)
		}
		sum, count := 0, 0
		for _, n := range d.numbers {
			input := byNumber[n]
			if !input.Answered {
				continue
			}
			count++
			value := input.Raw
			if managementTraitsFixtureIsReverse(d, n) {
				value = 6 - value
			}
			sum += value
		}
		if score.AnsweredCount != count || score.ScoreSum != sum {
			t.Fatalf("%s audit count/sum = %d/%d, want %d/%d", d.key, score.AnsweredCount, score.ScoreSum, count, sum)
		}
	}
	for i, module := range []struct {
		key   string
		count int
	}{{"self", 3}, {"interpersonal", 4}, {"task", 3}, {"development", 3}} {
		if got.Modules[i].Key != module.key || got.Modules[i].DimensionCount != module.count {
			t.Fatalf("module %d identity/count = %+v", i+1, got.Modules[i])
		}
	}
}

func TestManagementTraitsDimensions_IndependentCatalog(t *testing.T) {
	fixture, got := managementTraitsFixture(), ManagementTraitsDimensions()
	if len(got) != 13 {
		t.Fatalf("dimensions = %d, want 13", len(got))
	}
	seen := make(map[int]bool, 140)
	forward, reverse := 0, 0
	normSum := new(big.Rat)
	for i, d := range fixture {
		actual := got[i]
		if actual.Key != d.key || actual.Name != d.name || actual.Module != d.module || actual.Order != i+1 || len(actual.Items) != len(d.numbers) {
			t.Fatalf("catalog dimension %d = %+v", i+1, actual)
		}
		managementTraitsAssertRat(t, d.key+" norm", actual.Norm, d.norm)
		normSum.Add(normSum, actual.Norm)
		for j, n := range d.numbers {
			item := actual.Items[j]
			if item.Number != n || item.Reverse != managementTraitsFixtureIsReverse(d, n) {
				t.Fatalf("%s item %d = %+v, want V%d with fixture direction", d.key, j, item, n)
			}
			if item.Number < 1 || item.Number > 140 || seen[item.Number] {
				t.Fatalf("duplicate/out-of-range V%d", item.Number)
			}
			seen[item.Number] = true
			if item.Reverse {
				reverse++
			} else {
				forward++
			}
		}
	}
	if len(seen) != 140 || forward != 100 || reverse != 40 {
		t.Fatalf("unique/forward/reverse = %d/%d/%d, want 140/100/40", len(seen), forward, reverse)
	}
	managementTraitsAssertRat(t, "norm sum", normSum, big.NewRat(705, 1))
}

func TestManagementTraitsDimensions_FreshDeepCopies(t *testing.T) {
	first, second := ManagementTraitsDimensions(), ManagementTraitsDimensions()
	if len(first) != 13 || len(second) != 13 {
		t.Fatal("catalog must contain 13 dimensions")
	}
	for i := range first {
		if first[i].Norm == nil || second[i].Norm == nil || len(first[i].Items) == 0 || len(second[i].Items) == 0 {
			t.Fatal("norm and items must be present")
		}
		if &first[i] == &second[i] || first[i].Norm == second[i].Norm || &first[i].Items[0] == &second[i].Items[0] {
			t.Fatalf("dimension %d shares mutable storage", i+1)
		}
		first[i].Key, first[i].Name, first[i].Module, first[i].Order = "mutated", "mutated", "mutated", -1
		first[i].Norm.SetInt64(-1)
		first[i].Items[0].Number = -1
		first[i].Items[0].Reverse = !first[i].Items[0].Reverse
	}
	if !reflect.DeepEqual(second, ManagementTraitsDimensions()) {
		t.Fatal("mutating one catalog polluted another or a future catalog")
	}
}

func TestManagementTraitsCalculate_AllThreeAndQuestionnaireIdentity(t *testing.T) {
	if ManagementTraitsQuestionnaireStaff != "staff" || ManagementTraitsQuestionnaireLeader != "leader" {
		t.Fatal("questionnaire constants must be staff/leader, not repository codes")
	}
	inputs := managementTraitsFixtureAnswers(3)
	before := append([]ManagementTraitsAnswer(nil), inputs...)
	staff, err := CalculateManagementTraits(ManagementTraitsQuestionnaireStaff, inputs)
	if err != nil {
		t.Fatal(err)
	}
	leader, err := CalculateManagementTraits(ManagementTraitsQuestionnaireLeader, inputs)
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range []ManagementTraitsResult{staff, leader} {
		managementTraitsAssertAudit(t, got.Questionnaire, inputs, got)
		managementTraitsAssertRat(t, "overall", got.OverallScore, big.NewRat(50, 1))
		managementTraitsAssertRat(t, "overall norm", got.OverallNorm, big.NewRat(705, 13))
		if got.OverallLevel != "qualified" || got.OverallNorm.FloatString(2) != "54.23" {
			t.Fatalf("overall level/norm = %s/%v", got.OverallLevel, got.OverallNorm)
		}
		for i, d := range got.Dimensions {
			managementTraitsAssertRat(t, d.Key, d.Score, big.NewRat(50, 1))
			managementTraitsAssertRat(t, d.Key+" norm", d.Norm, managementTraitsFixture()[i].norm)
			if d.Level != "qualified" {
				t.Fatalf("%s level = %s", d.Key, d.Level)
			}
		}
		for _, m := range got.Modules {
			managementTraitsAssertRat(t, m.Key, m.Score, big.NewRat(50, 1))
		}
		want := []string{"self_confidence", "emotional_stability", "self_discipline"}
		if !reflect.DeepEqual(got.Highest, want) || !reflect.DeepEqual(got.Lowest, want) {
			t.Fatalf("all-tied highest/lowest = %v/%v, want overlap %v", got.Highest, got.Lowest, want)
		}
	}
	if staff.Questionnaire != "staff" || leader.Questionnaire != "leader" {
		t.Fatal("questionnaire identity was inferred or lost")
	}
	leader.Questionnaire = staff.Questionnaire
	if !reflect.DeepEqual(staff, leader) || !reflect.DeepEqual(inputs, before) {
		t.Fatal("questionnaires changed calculation or calculation mutated inputs")
	}
	for left, right := 0, len(inputs)-1; left < right; left, right = left+1, right-1 {
		inputs[left], inputs[right] = inputs[right], inputs[left]
	}
	reordered, err := CalculateManagementTraits("staff", inputs)
	if err != nil || !reflect.DeepEqual(reordered, staff) {
		t.Fatalf("reversed input changed result: %v", err)
	}
}

func TestManagementTraitsCalculate_AllFourExactAggregation(t *testing.T) {
	inputs := managementTraitsFixtureAnswers(4)
	got, err := CalculateManagementTraits("staff", inputs)
	if err != nil {
		t.Fatal(err)
	}
	managementTraitsAssertAudit(t, "staff", inputs, got)
	sum := new(big.Rat)
	totalSum := 0
	for i, d := range managementTraitsFixture() {
		managementTraitsAssertRat(t, d.key, got.Dimensions[i].Score, d.allFourScore)
		managementTraitsAssertRat(t, d.key+" norm", got.Dimensions[i].Norm, d.norm)
		if got.Dimensions[i].ScoreSum != d.allFourSum {
			t.Fatalf("%s sum = %d, want %d", d.key, got.Dimensions[i].ScoreSum, d.allFourSum)
		}
		wantLevel := "qualified"
		if d.key == "planning" || d.key == "decisiveness" {
			wantLevel = "good"
		}
		if got.Dimensions[i].Level != wantLevel {
			t.Fatalf("%s level = %s, want %s", d.key, got.Dimensions[i].Level, wantLevel)
		}
		sum.Add(sum, d.allFourScore)
		totalSum += d.allFourSum
	}
	managementTraitsAssertRat(t, "13-dimension equal overall", got.OverallScore, new(big.Rat).Quo(sum, big.NewRat(13, 1)))
	managementTraitsAssertRat(t, "overall norm", got.OverallNorm, big.NewRat(705, 13))
	if got.OverallLevel != "qualified" {
		t.Fatalf("overall level = %s", got.OverallLevel)
	}
	moduleScores := []*big.Rat{big.NewRat(545, 9), big.NewRat(1485, 26), big.NewRat(590, 9), big.NewRat(1945, 33)}
	moduleSum := new(big.Rat)
	for i, want := range moduleScores {
		managementTraitsAssertRat(t, got.Modules[i].Key+" equal member weights", got.Modules[i].Score, want)
		moduleSum.Add(moduleSum, want)
	}
	if got.OverallScore.Cmp(new(big.Rat).Quo(moduleSum, big.NewRat(4, 1))) == 0 || got.OverallScore.Cmp(big.NewRat(int64(25*(totalSum-140)), 140)) == 0 {
		t.Fatal("overall must not be four-module-equal or question-weighted")
	}
	if got.Modules[0].Score.Cmp(big.NewRat(25*(44+34+32-32), 32)) == 0 || got.Modules[1].Score.Cmp(big.NewRat(25*(24+36+46+36-43), 43)) == 0 {
		t.Fatal("modules must not be question-weighted")
	}
	if !reflect.DeepEqual(got.Highest, []string{"planning", "decisiveness", "self_confidence"}) || !reflect.DeepEqual(got.Lowest, []string{"sociality", "self_discipline", "responsibility"}) {
		t.Fatalf("exact ranking/tie order = %v/%v", got.Highest, got.Lowest)
	}
	for left, right := 0, len(inputs)-1; left < right; left, right = left+1, right-1 {
		inputs[left], inputs[right] = inputs[right], inputs[left]
	}
	reordered, err := CalculateManagementTraits("staff", inputs)
	if err != nil || !reflect.DeepEqual(got, reordered) {
		t.Fatalf("nonuniform scores changed with reversed input: %v", err)
	}
}

func TestManagementTraitsCalculate_DirectionAdjustedExtremes(t *testing.T) {
	for _, questionnaire := range []string{"staff", "leader"} {
		for _, tc := range []struct {
			name  string
			final int
			level string
		}{{"minimum", 1, "insufficient"}, {"maximum", 5, "excellent"}} {
			t.Run(questionnaire+"/"+tc.name, func(t *testing.T) {
				inputs := managementTraitsFixtureFinalAnswers(tc.final)
				got, err := CalculateManagementTraits(questionnaire, inputs)
				if err != nil {
					t.Fatal(err)
				}
				managementTraitsAssertAudit(t, questionnaire, inputs, got)
				want := big.NewRat(int64(25*(tc.final-1)), 1)
				managementTraitsAssertRat(t, "overall", got.OverallScore, want)
				managementTraitsAssertRat(t, "overall norm", got.OverallNorm, big.NewRat(705, 13))
				if got.OverallLevel != tc.level {
					t.Fatalf("overall level = %s, want %s", got.OverallLevel, tc.level)
				}
				for i, d := range got.Dimensions {
					managementTraitsAssertRat(t, d.Key, d.Score, want)
					managementTraitsAssertRat(t, d.Key+" norm", d.Norm, managementTraitsFixture()[i].norm)
					if d.Level != tc.level {
						t.Fatalf("%s level = %s", d.Key, d.Level)
					}
				}
				for _, m := range got.Modules {
					managementTraitsAssertRat(t, m.Key, m.Score, want)
				}
				wantRank := []string{"self_confidence", "emotional_stability", "self_discipline"}
				if !reflect.DeepEqual(got.Highest, wantRank) || !reflect.DeepEqual(got.Lowest, wantRank) {
					t.Fatalf("grade-filtered extreme rankings: %v/%v", got.Highest, got.Lowest)
				}
			})
		}
	}
}

func TestManagementTraitsCalculate_ExactRankingWithoutGradeFilters(t *testing.T) {
	for _, final := range []int{1, 5} {
		inputs := managementTraitsFixtureFinalAnswers(final)
		// One final point of separation in 11/12/13-item dimensions.
		for _, n := range []int{6, 5, 4, 26, 22} {
			if final == 1 {
				inputs[n-1].Raw++
			} else {
				inputs[n-1].Raw--
			}
		}
		got, err := CalculateManagementTraits("staff", inputs)
		if err != nil {
			t.Fatal(err)
		}
		wantSeparated := []string{"learning", "self_confidence", "planning"}
		wantTied := []string{"emotional_stability", "self_discipline", "sociality"}
		wantHigh, wantLow := wantSeparated, wantTied
		wantLevel := "insufficient"
		if final == 5 {
			wantHigh, wantLow = wantTied, wantSeparated
			wantLevel = "excellent"
		}
		if !reflect.DeepEqual(got.Highest, wantHigh) || !reflect.DeepEqual(got.Lowest, wantLow) {
			t.Fatalf("final=%d exact ranking = %v/%v, want %v/%v", final, got.Highest, got.Lowest, wantHigh, wantLow)
		}
		for _, d := range got.Dimensions {
			if d.Level != wantLevel {
				t.Fatalf("final=%d %s level = %s", final, d.Key, d.Level)
			}
		}
	}
}

func TestManagementTraitsCalculate_IncompleteSuppressesAllFormalResults(t *testing.T) {
	for _, questionnaire := range []string{"staff", "leader"} {
		for _, tc := range []struct {
			name    string
			missing int
		}{{"forward_unanswered", 6}, {"reverse_unanswered", 120}, {"no_answers", 0}} {
			t.Run(questionnaire+"/"+tc.name, func(t *testing.T) {
				inputs := managementTraitsFixtureAnswers(4)
				for i := range inputs {
					if tc.missing == 0 || inputs[i].Number == tc.missing {
						inputs[i].Answered, inputs[i].Raw = false, 0
					}
				}
				got, err := CalculateManagementTraits(questionnaire, inputs)
				if err != nil {
					t.Fatal(err)
				}
				managementTraitsAssertAudit(t, questionnaire, inputs, got)
				if got.IsComplete || got.OverallScore != nil || got.OverallNorm != nil || got.OverallLevel != "" || got.Highest == nil || got.Lowest == nil || len(got.Highest) != 0 || len(got.Lowest) != 0 {
					t.Fatalf("incomplete result leaked formal overall/rankings: %+v", got)
				}
				for _, d := range got.Dimensions {
					if d.Score != nil || d.Norm != nil || d.Level != "" {
						t.Fatalf("incomplete paper leaked %s formal score/norm/level: %+v", d.Key, d)
					}
				}
				for _, m := range got.Modules {
					if m.Score != nil {
						t.Fatalf("incomplete paper leaked module %s score", m.Key)
					}
				}
			})
		}
	}
}

func TestManagementTraitsCalculate_InvalidInputBoundary(t *testing.T) {
	cases := []struct {
		name          string
		questionnaire string
		mutate        func([]ManagementTraitsAnswer) []ManagementTraitsAnswer
	}{
		{"nil", "staff", func(_ []ManagementTraitsAnswer) []ManagementTraitsAnswer { return nil }},
		{"139_rows", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer { return a[:139] }},
		{"141_rows", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer {
			return append(a, ManagementTraitsAnswer{Number: 141, Answered: true, Raw: 3})
		}},
		{"duplicate", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer { a[139].Number = 1; return a }},
		{"number_zero", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer { a[0].Number = 0; return a }},
		{"number_141", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer { a[139].Number = 141; return a }},
		{"answered_raw_zero", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer { a[0].Raw = 0; return a }},
		{"answered_raw_six", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer { a[119].Raw = 6; return a }},
		{"answered_raw_negative", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer { a[0].Raw = -1; return a }},
		{"unanswered_raw_one", "staff", func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer {
			a[119].Answered = false
			a[119].Raw = 1
			return a
		}},
	}
	for _, questionnaire := range []string{"", "STAFF", " staff", "leader ", "00201", "00202", "frontline_employee", "unknown"} {
		cases = append(cases, struct {
			name          string
			questionnaire string
			mutate        func([]ManagementTraitsAnswer) []ManagementTraitsAnswer
		}{"questionnaire_" + questionnaire, questionnaire, func(a []ManagementTraitsAnswer) []ManagementTraitsAnswer { return a }})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := CalculateManagementTraits(tc.questionnaire, tc.mutate(managementTraitsFixtureAnswers(3))); err == nil {
				t.Fatal("invalid boundary input was accepted")
			}
		})
	}
}

func TestManagementTraitsLevelForScore_ExactBoundariesAndFinalFormatting(t *testing.T) {
	for _, tc := range []struct {
		name, value, want string
	}{
		{"zero", "0", "insufficient"},
		{"below_ten", "9.999", "insufficient"},
		{"ten", "10", "weak"},
		{"below_thirty", "29.999", "weak"},
		{"thirty", "30", "qualified"},
		{"below_seventy", "69.999", "qualified"},
		{"seventy", "70", "good"},
		{"below_ninety", "89.999", "good"},
		{"ninety", "90", "excellent"},
		{"hundred", "100", "excellent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			score, ok := new(big.Rat).SetString(tc.value)
			if !ok {
				t.Fatal("invalid independent rational fixture")
			}
			before := new(big.Rat).Set(score)
			got, err := ManagementTraitsLevelForScore(score)
			if err != nil || got != tc.want || score.Cmp(before) != 0 {
				t.Fatalf("level(%s) = %s/%v; want %s without mutating input", tc.value, got, err, tc.want)
			}
			if tc.value == "69.999" && score.FloatString(2) != "70.00" {
				t.Fatalf("final formatting = %s, want 70.00 while level stays qualified", score.FloatString(2))
			}
		})
	}
	for _, score := range []*big.Rat{nil, big.NewRat(-1, 1000), big.NewRat(100001, 1000)} {
		if _, err := ManagementTraitsLevelForScore(score); err == nil {
			t.Fatalf("invalid score %v was accepted", score)
		}
	}
	if got := big.NewRat(545, 8).FloatString(2); got != "68.13" {
		t.Fatalf("final HALF_UP 68.125 = %s, want 68.13", got)
	}
}
