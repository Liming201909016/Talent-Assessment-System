package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
)

type managementTraitsResultValidationFixture struct {
	Bundle     model.ManagementTraitsDefinitionBundle
	Paper      model.ManagementTraitsPaperSnapshot
	Questions  []model.ManagementTraitsPaperQuestionSnapshot
	Run        model.ManagementTraitsResultRun
	Dimensions []model.ManagementTraitsResultDimension
	Modules    []model.ManagementTraitsResultModule
}

func managementTraitsResultValidationDecimal(r *big.Rat) *decimal.Decimal {
	d := decimal.NewFromBigRat(r, 6)
	return &d
}

func managementTraitsResultValidationText(s string) *string { return &s }

// Independent oracle: the pre-existing explicit fixture, not Calculate or the
// validator under test, supplies V directions, sums, names, norms and modules.
func managementTraitsResultValidationFixtureFor(t *testing.T, audience, participant string, raw, final int) (managementTraitsResultValidationFixture, ManagementTraitsResult) {
	t.Helper()
	b, p, qs, _ := managementTraitsDecodeStoredFixture(t, audience, participant, raw)
	if final != 0 {
		for _, d := range managementTraitsFixture() {
			for _, n := range d.numbers {
				r := final
				if managementTraitsFixtureIsReverse(d, n) {
					r = 6 - r
				}
				*qs[n-1].RawAnswer, *qs[n-1].FinalScore = r, final
				*qs[n-1].SelectedOptionID = fmt.Sprintf("source-o-%d-%d", n, r)
			}
		}
	}
	validated, err := ValidateManagementTraitsStoredInput(b, p, qs, managementTraitsDecodeBudget(b, p, qs))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(validated.Canonical.JSON)
	if validated.Canonical.SHA256 != hex.EncodeToString(h[:]) {
		t.Fatal("baseline input SHA")
	}
	f := managementTraitsResultValidationFixture{Bundle: b, Paper: p, Questions: qs,
		Run: model.ManagementTraitsResultRun{ID: "opaque-run", PaperID: p.PaperID, ExamID: p.ExamID,
			ParticipantType: p.ParticipantType, ParticipantID: p.ParticipantID, Questionnaire: audience,
			ProductVersion: b.ProductVersion, QuestionVersion: b.QuestionVersion, ScoringVersion: b.ScoringVersion, NormVersion: b.NormVersion,
			ScoringManifestSHA: b.ScoringManifestSHA, InputSHA: hex.EncodeToString(h[:]), Status: "completed", TotalQuestionCount: 140, AnsweredQuestionCount: 140},
		Dimensions: make([]model.ManagementTraitsResultDimension, 0, 13), Modules: make([]model.ManagementTraitsResultModule, 0, 4)}
	want := ManagementTraitsResult{Questionnaire: audience, TotalQuestionCount: 140, AnsweredQuestionCount: 140, IsComplete: true,
		Dimensions: make([]ManagementTraitsDimensionScore, 0, 13), Modules: make([]ManagementTraitsModuleScore, 0, 4),
		Highest: make([]string, 0, 3), Lowest: make([]string, 0, 3), OverallScore: new(big.Rat), OverallNorm: big.NewRat(705, 13)}
	for i, d := range managementTraitsFixture() {
		count, sum := len(d.numbers), 3*len(d.numbers)
		score := big.NewRat(50, 1)
		if raw == 4 {
			sum, score = d.allFourSum, new(big.Rat).Set(d.allFourScore)
		}
		if final != 0 {
			sum, score = count*final, big.NewRat(int64(25*(final-1)), 1)
		}
		level := "qualified"
		if score.Cmp(big.NewRat(70, 1)) >= 0 {
			level = "good"
		}
		if final == 1 {
			level = "insufficient"
		}
		if final == 5 {
			level = "excellent"
		}
		want.Dimensions = append(want.Dimensions, ManagementTraitsDimensionScore{Key: d.key, Name: d.name, Module: d.module, Order: i + 1,
			QuestionCount: count, AnsweredCount: count, ScoreSum: sum, Score: score, Norm: new(big.Rat).Set(d.norm), Level: level})
		f.Dimensions = append(f.Dimensions, model.ManagementTraitsResultDimension{ID: fmt.Sprintf("dimension-%d", i), RunID: f.Run.ID,
			DimensionKey: d.key, DimensionName: d.name, ModuleKey: d.module, DisplayOrder: i + 1, QuestionCount: count, AnsweredCount: count,
			ScoreSum: sum, Score: managementTraitsResultValidationDecimal(score), Norm: managementTraitsResultValidationDecimal(d.norm), Level: managementTraitsResultValidationText(level)})
		want.OverallScore.Add(want.OverallScore, score)
	}
	for i, key := range []string{"self", "interpersonal", "task", "development"} {
		count, sum := 0, new(big.Rat)
		for _, d := range want.Dimensions {
			if d.Module == key {
				count++
				sum.Add(sum, d.Score)
			}
		}
		score := sum.Quo(sum, big.NewRat(int64(count), 1))
		want.Modules = append(want.Modules, ManagementTraitsModuleScore{Key: key, DimensionCount: count, Score: score})
		f.Modules = append(f.Modules, model.ManagementTraitsResultModule{ID: fmt.Sprintf("module-%d", i), RunID: f.Run.ID,
			ModuleKey: key, DisplayOrder: i + 1, DimensionCount: count, Score: managementTraitsResultValidationDecimal(score)})
	}
	want.OverallScore.Quo(want.OverallScore, big.NewRat(13, 1))
	want.OverallLevel = "qualified"
	if final == 1 {
		want.OverallLevel = "insufficient"
	}
	if final == 5 {
		want.OverallLevel = "excellent"
	}
	want.Highest = []string{"self_confidence", "emotional_stability", "self_discipline"}
	want.Lowest = []string{"self_confidence", "emotional_stability", "self_discipline"}
	if raw == 4 && final == 0 {
		want.Highest = []string{"planning", "decisiveness", "self_confidence"}
		want.Lowest = []string{"sociality", "self_discipline", "responsibility"}
	}
	f.Run.OverallScore, f.Run.OverallNorm = managementTraitsResultValidationDecimal(want.OverallScore), managementTraitsResultValidationDecimal(want.OverallNorm)
	f.Run.OverallLevel = managementTraitsResultValidationText(want.OverallLevel)
	return f, want
}

func managementTraitsResultValidationCall(f managementTraitsResultValidationFixture, budget int) (ManagementTraitsResult, error) {
	return ValidateManagementTraitsStoredResult(f.Bundle, f.Paper, f.Questions, f.Run, f.Dimensions, f.Modules, budget)
}

func managementTraitsResultValidationBudget(f managementTraitsResultValidationFixture) int {
	return managementTraitsDecodeBudget(f.Bundle, f.Paper, f.Questions)
}

func managementTraitsResultValidationClone(t *testing.T, f managementTraitsResultValidationFixture) managementTraitsResultValidationFixture {
	t.Helper()
	cloned := managementTraitsContractClone(t, f)
	cloned.Run.ID = f.Run.ID // Preserve invalid UTF-8 ID bytes in rejection snapshots.
	copyDecimal := func(d *decimal.Decimal) *decimal.Decimal {
		if d == nil {
			return nil
		}
		v := d.Copy()
		return &v
	}
	cloned.Run.OverallScore, cloned.Run.OverallNorm = copyDecimal(f.Run.OverallScore), copyDecimal(f.Run.OverallNorm)
	for i, d := range f.Dimensions {
		cloned.Dimensions[i].Score, cloned.Dimensions[i].Norm = copyDecimal(d.Score), copyDecimal(d.Norm)
	}
	for i, m := range f.Modules {
		cloned.Modules[i].Score = copyDecimal(m.Score)
	}
	return cloned
}

// Literal snapshots include independent decimal coefficients/exponents as well
// as all pointer-backed input values, so an in-place mutation cannot hide.
func managementTraitsResultValidationSnapshot(t *testing.T, f managementTraitsResultValidationFixture) []byte {
	t.Helper()
	values := []interface{}{f}
	for _, d := range f.Dimensions {
		values = append(values, d.Score, d.Norm, d.Level)
	}
	for _, m := range f.Modules {
		values = append(values, m.Score)
	}
	for _, d := range []*decimal.Decimal{f.Run.OverallScore, f.Run.OverallNorm} {
		if d != nil {
			values = append(values, d.Coefficient().String(), d.Exponent())
		}
	}
	for _, d := range f.Dimensions {
		for _, v := range []*decimal.Decimal{d.Score, d.Norm} {
			if v != nil {
				values = append(values, v.Coefficient().String(), v.Exponent())
			}
		}
	}
	for _, m := range f.Modules {
		if m.Score != nil {
			values = append(values, m.Score.Coefficient().String(), m.Score.Exponent())
		}
	}
	return managementTraitsDecodeJSON(t, values)
}

func managementTraitsResultValidationReject(t *testing.T, f managementTraitsResultValidationFixture, budget int) {
	t.Helper()
	before := managementTraitsResultValidationSnapshot(t, f)
	cloned := managementTraitsResultValidationClone(t, f)
	got, err := managementTraitsResultValidationCall(f, budget)
	if err == nil || !reflect.DeepEqual(got, ManagementTraitsResult{}) {
		t.Fatal("accepted corruption or returned partial result")
	}
	if err.Error() != "invalid management traits stored result" {
		t.Fatalf("nonuniform/non-sensitive error: %q", err.Error())
	}
	if !bytes.Equal(before, managementTraitsResultValidationSnapshot(t, f)) || !reflect.DeepEqual(cloned, f) {
		t.Fatal("rejection repaired/mutated input")
	}
}

func TestManagementTraitsResultValidationValid(t *testing.T) {
	for _, audience := range []string{"staff", "leader"} {
		for _, participant := range []string{"candidate", "tester"} {
			for _, raw := range []int{3, 4} {
				for _, profile := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/raw%d/profile%v", audience, participant, raw, profile), func(t *testing.T) {
						f, want := managementTraitsResultValidationFixtureFor(t, audience, participant, raw, 0)
						if profile {
							f.Paper.ProfileExamID = managementTraitsResultValidationText(f.Paper.ExamID)
						}
						before := managementTraitsResultValidationSnapshot(t, f)
						got, err := managementTraitsResultValidationCall(f, managementTraitsResultValidationBudget(f))
						if err != nil || !reflect.DeepEqual(got, want) {
							t.Fatal("independent exact oracle mismatch", err)
						}
						if !bytes.Equal(before, managementTraitsResultValidationSnapshot(t, f)) {
							t.Fatal("valid input mutated")
						}
						if raw == 4 {
							managementTraitsAssertRat(t, "learning", got.Dimensions[11].Score, big.NewRat(625, 11))
							if f.Dimensions[11].Score.String() != "56.818182" {
								t.Fatal("six-place rounding")
							}
						}
					})
				}
			}
		}
	}
	for _, final := range []int{1, 5} {
		t.Run(fmt.Sprintf("direction_adjusted_final%d", final), func(t *testing.T) {
			f, want := managementTraitsResultValidationFixtureFor(t, "staff", "tester", 3, final)
			base, _, qs, _ := managementTraitsDecodeStoredFixture(t, "staff", "tester", 3)
			if base.ScoringManifestSHA != f.Bundle.ScoringManifestSHA {
				t.Fatal("raw changed manifest SHA")
			}
			for i := range qs {
				if qs[i].ScoringSnapshotSHA != f.Questions[i].ScoringSnapshotSHA {
					t.Fatal("raw changed frozen question SHA")
				}
			}
			got, err := managementTraitsResultValidationCall(f, managementTraitsResultValidationBudget(f))
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("extreme reverse scoring", err)
			}
		})
	}
}

func TestManagementTraitsResultValidationReorderedDetached(t *testing.T) {
	f, want := managementTraitsResultValidationFixtureFor(t, "leader", "tester", 4, 0)
	f.Modules[0].ID = f.Dimensions[0].ID // IDs are unique within each table, not across tables.
	for _, d := range []*decimal.Decimal{f.Run.OverallScore, f.Run.OverallNorm, f.Dimensions[0].Score, f.Dimensions[0].Norm, f.Modules[0].Score} {
		*d = decimal.New(d.Coefficient().Mul(d.Coefficient(), big.NewInt(1000)).Int64(), d.Exponent()-3)
	}
	for i := 0; i < len(f.Dimensions)/2; i++ {
		j := len(f.Dimensions) - 1 - i
		f.Dimensions[i], f.Dimensions[j] = f.Dimensions[j], f.Dimensions[i]
	}
	for i := 0; i < 2; i++ {
		f.Modules[i], f.Modules[3-i] = f.Modules[3-i], f.Modules[i]
	}
	for i := 0; i < 70; i++ {
		f.Questions[i], f.Questions[139-i] = f.Questions[139-i], f.Questions[i]
	}
	before := managementTraitsResultValidationSnapshot(t, f)
	cloned := managementTraitsResultValidationClone(t, f)
	got, err := managementTraitsResultValidationCall(f, managementTraitsResultValidationBudget(f))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("reorder/alternate decimal representation", err)
	}
	got.OverallScore.SetInt64(1)
	got.OverallNorm.SetInt64(2)
	for i := range got.Dimensions {
		got.Dimensions[i].Score.SetInt64(3)
		got.Dimensions[i].Norm.SetInt64(4)
		got.Dimensions[i].Name = "polluted"
	}
	for i := range got.Modules {
		got.Modules[i].Score.SetInt64(5)
		got.Modules[i].Key = "polluted"
	}
	got.Highest[0], got.Lowest[0] = "polluted", "polluted"
	fresh, err := managementTraitsResultValidationCall(f, managementTraitsResultValidationBudget(f))
	if err != nil || !reflect.DeepEqual(fresh, want) || !bytes.Equal(before, managementTraitsResultValidationSnapshot(t, f)) || !reflect.DeepEqual(cloned, f) {
		t.Fatal("returned mutable values alias caches/future calls", err)
	}
	if !reflect.DeepEqual(ManagementTraitsDimensions()[0].Norm, big.NewRat(115, 2)) {
		t.Fatal("catalog polluted")
	}
}

func managementTraitsResultValidationSet(target interface{}, field string, value interface{}) {
	v := reflect.ValueOf(target).Elem().FieldByName(field)
	if value == nil {
		v.Set(reflect.Zero(v.Type()))
		return
	}
	v.Set(reflect.ValueOf(value))
}

func TestManagementTraitsResultValidationRunRejects(t *testing.T) {
	badDecimal := decimal.RequireFromString("-1")
	cases := map[string][]interface{}{
		"ID":      {"", " run ", "secret\x00run", strings.Repeat("x", 65), string([]byte{255})},
		"Status":  {"", "incomplete", "unknown", "Completed", "completed "},
		"PaperID": {"", "foreign-paper"}, "ExamID": {"", "foreign-exam"},
		"ParticipantType": {"", "tester", "user"}, "ParticipantID": {"", "foreign-participant"},
		"Questionnaire":  {"", "leader"},
		"ProductVersion": {"", "other"}, "QuestionVersion": {"", "other"}, "ScoringVersion": {"", "other"}, "NormVersion": {"", "other"},
		"ScoringManifestSHA": {"", strings.Repeat("a", 64)}, "InputSHA": {"", strings.Repeat("a", 64)},
		"TotalQuestionCount": {0, 139, 141}, "AnsweredQuestionCount": {0, 139, 141},
		"OverallScore": {nil, &badDecimal, managementTraitsResultValidationDecimal(big.NewRat(101, 1))},
		"OverallNorm":  {nil, &badDecimal, managementTraitsResultValidationDecimal(big.NewRat(5423, 100))},
		"OverallLevel": {nil, managementTraitsResultValidationText(""), managementTraitsResultValidationText("good")},
	}
	for field, values := range cases {
		for i, value := range values {
			t.Run(fmt.Sprintf("%s/%d", field, i), func(t *testing.T) {
				f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 3, 0)
				managementTraitsResultValidationSet(&f.Run, field, value)
				managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
			})
		}
	}
	for _, field := range []string{"InputSHA", "ScoringManifestSHA", "OverallScore", "OverallNorm"} {
		t.Run(field+"/case_or_extra_precision", func(t *testing.T) {
			f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
			v := reflect.ValueOf(&f.Run).Elem().FieldByName(field)
			if v.Kind() == reflect.String {
				v.SetString(strings.ToUpper(v.String()))
			} else {
				d := v.Interface().(*decimal.Decimal).Add(decimal.RequireFromString("0.0000001"))
				v.Set(reflect.ValueOf(&d))
			}
			managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
		})
	}
}

func TestManagementTraitsResultValidationDimensionRejects(t *testing.T) {
	negative := decimal.RequireFromString("-1")
	cases := map[string][]interface{}{
		"ID": {"", " secret ", "secret\nID", strings.Repeat("x", 65)}, "RunID": {"", "foreign-run"},
		"DimensionKey": {"", "unknown"}, "DisplayOrder": {0, 14}, "DimensionName": {"", "wrong"}, "ModuleKey": {"", "wrong"},
		"QuestionCount": {0, 139}, "AnsweredCount": {0, 139}, "ScoreSum": {0, 700},
		"Score": {nil, &negative, managementTraitsResultValidationDecimal(big.NewRat(101, 1))},
		"Norm":  {nil, &negative, managementTraitsResultValidationDecimal(big.NewRat(101, 1))},
		"Level": {nil, managementTraitsResultValidationText(""), managementTraitsResultValidationText("excellent")},
	}
	for index := 0; index < 13; index++ {
		for field, values := range cases {
			for i, value := range values {
				t.Run(fmt.Sprintf("row%d/%s/%d", index, field, i), func(t *testing.T) {
					f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
					managementTraitsResultValidationSet(&f.Dimensions[index], field, value)
					managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
				})
			}
		}
		for _, field := range []string{"Score", "Norm"} {
			t.Run(fmt.Sprintf("row%d/%s/extra_precision", index, field), func(t *testing.T) {
				f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
				v := reflect.ValueOf(&f.Dimensions[index]).Elem().FieldByName(field)
				d := v.Interface().(*decimal.Decimal).Add(decimal.RequireFromString("0.0000001"))
				v.Set(reflect.ValueOf(&d))
				managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
			})
		}
	}
	for _, kind := range []string{"missing", "extra", "duplicate_id", "duplicate_key", "duplicate_order", "swapped_names", "wrong_integer_source_with_matching_cache"} {
		t.Run(kind, func(t *testing.T) {
			f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
			switch kind {
			case "missing":
				f.Dimensions = f.Dimensions[:12]
			case "extra":
				f.Dimensions = append(f.Dimensions, f.Dimensions[0])
			case "duplicate_id":
				f.Dimensions[1].ID = f.Dimensions[0].ID
			case "duplicate_key":
				f.Dimensions[1].DimensionKey = f.Dimensions[0].DimensionKey
			case "duplicate_order":
				f.Dimensions[1].DisplayOrder = f.Dimensions[0].DisplayOrder
			case "swapped_names":
				f.Dimensions[0].DimensionName, f.Dimensions[1].DimensionName = f.Dimensions[1].DimensionName, f.Dimensions[0].DimensionName
			case "wrong_integer_source_with_matching_cache":
				d := &f.Dimensions[0]
				d.ScoreSum++
				d.Score = managementTraitsResultValidationDecimal(big.NewRat(int64(25*(d.ScoreSum-d.QuestionCount)), int64(d.QuestionCount)))
			}
			managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
		})
	}
}

func TestManagementTraitsResultValidationModuleRejects(t *testing.T) {
	negative := decimal.RequireFromString("-1")
	cases := map[string][]interface{}{
		"ID": {"", " bad ", "secret\x00module", strings.Repeat("x", 65)}, "RunID": {"", "foreign-run"},
		"ModuleKey": {"", "unknown"}, "DisplayOrder": {0, 5}, "DimensionCount": {0, 13},
		"Score": {nil, &negative, managementTraitsResultValidationDecimal(big.NewRat(101, 1))},
	}
	for index := 0; index < 4; index++ {
		for field, values := range cases {
			for i, value := range values {
				t.Run(fmt.Sprintf("row%d/%s/%d", index, field, i), func(t *testing.T) {
					f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
					managementTraitsResultValidationSet(&f.Modules[index], field, value)
					managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
				})
			}
		}
		t.Run(fmt.Sprintf("row%d/extra_precision", index), func(t *testing.T) {
			f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
			d := f.Modules[index].Score.Add(decimal.RequireFromString("0.0000001"))
			f.Modules[index].Score = &d
			managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
		})
	}
	for _, kind := range []string{"missing", "extra", "duplicate_id", "duplicate_key", "duplicate_order", "swapped_keys", "weighted_by_questions", "rounded_member_aggregate"} {
		t.Run(kind, func(t *testing.T) {
			f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
			switch kind {
			case "missing":
				f.Modules = f.Modules[:3]
			case "extra":
				f.Modules = append(f.Modules, f.Modules[0])
			case "duplicate_id":
				f.Modules[1].ID = f.Modules[0].ID
			case "duplicate_key":
				f.Modules[1].ModuleKey = f.Modules[0].ModuleKey
			case "duplicate_order":
				f.Modules[1].DisplayOrder = f.Modules[0].DisplayOrder
			case "swapped_keys":
				f.Modules[0].ModuleKey, f.Modules[1].ModuleKey = f.Modules[1].ModuleKey, f.Modules[0].ModuleKey
			case "weighted_by_questions":
				f.Modules[0].Score = managementTraitsResultValidationDecimal(big.NewRat(25*(44+34+32-32), 32))
			case "rounded_member_aggregate":
				s := decimal.Zero
				for _, d := range f.Dimensions {
					if d.ModuleKey == "task" {
						s = s.Add(*d.Score)
					}
				}
				s = s.DivRound(decimal.NewFromInt(3), 6)
				f.Modules[2].Score = &s
			}
			managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
		})
	}
}

func TestManagementTraitsResultValidationUpstreamRejects(t *testing.T) {
	for _, kind := range []string{"manifest_json", "mapping_json", "options_json", "manifest_hash", "mapping_hash", "question_hash", "source_option", "raw_provenance", "forward_final", "reverse_final", "zero_budget", "negative_budget", "small_budget"} {
		t.Run(kind, func(t *testing.T) {
			f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
			budget := managementTraitsResultValidationBudget(f)
			switch kind {
			case "manifest_json":
				f.Bundle.ScoringManifest = `{"secret":"private-json"}`
			case "mapping_json":
				f.Paper.MappingSnapshot += " {}"
				budget += 3
			case "options_json":
				f.Questions[0].OptionsSnapshot = "null"
			case "manifest_hash":
				f.Bundle.ScoringManifestSHA = strings.Repeat("b", 64)
				f.Paper.ScoringManifestSHA = f.Bundle.ScoringManifestSHA
				f.Run.ScoringManifestSHA = f.Bundle.ScoringManifestSHA
			case "mapping_hash":
				f.Paper.MappingSHA = strings.Repeat("b", 64)
			case "question_hash":
				f.Questions[0].ScoringSnapshotSHA = strings.Repeat("b", 64)
			case "source_option":
				*f.Questions[0].SelectedOptionID = *f.Questions[1].SelectedOptionID
			case "raw_provenance":
				*f.Questions[0].RawAnswer = 3
				*f.Questions[0].FinalScore = 3
			case "forward_final", "reverse_final":
				for i := range f.Questions {
					if f.Questions[i].Reverse == (kind == "reverse_final") {
						*f.Questions[i].FinalScore = 6 - *f.Questions[i].FinalScore
						break
					}
				}
			case "zero_budget":
				budget = 0
			case "negative_budget":
				budget = -1
			case "small_budget":
				budget--
			}
			managementTraitsResultValidationReject(t, f, budget)
		})
	}
}

func TestManagementTraitsResultValidationIncomplete(t *testing.T) {
	for _, all := range []bool{false, true} {
		t.Run(fmt.Sprintf("all_unanswered_%v", all), func(t *testing.T) {
			f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 4, 0)
			for i := range f.Questions {
				if all || i == 119 {
					f.Questions[i].SelectedOptionID, f.Questions[i].RawAnswer, f.Questions[i].FinalScore = nil, nil, nil
				}
			}
			v, err := ValidateManagementTraitsStoredInput(f.Bundle, f.Paper, f.Questions, managementTraitsResultValidationBudget(f))
			if err != nil {
				t.Fatal("valid incomplete input", err)
			}
			h := sha256.Sum256(v.Canonical.JSON)
			f.Run.InputSHA = hex.EncodeToString(h[:])
			managementTraitsResultValidationReject(t, f, managementTraitsResultValidationBudget(f))
		})
	}
}

func TestManagementTraitsResultValidationDeferredLifecycle(t *testing.T) {
	f, want := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 3, 0)
	f.Bundle.Status = "not-approved"
	f.Run.Source = "arbitrary-secret-source"
	f.Paper.Source = "arbitrary-history"
	f.Paper.ParticipantSnapshot, f.Paper.FieldContract, f.Paper.EvidenceSnapshot = "not json", "not json", "not json"
	f.Paper.EvidenceSHA, f.Paper.IdentitySource = "unverified", "unverified"
	f.Run.CreatedAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	seconds := -999
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	f.Run.UserTimeSeconds, f.Run.SubmittedAt = &seconds, &future
	f.Paper.StartedAt, f.Paper.LimitTime = future, &f.Run.CreatedAt
	for i := range f.Questions {
		f.Questions[i].SubmittedAt = &future
		f.Questions[i].CreatedAt = future
	}
	for i := range f.Dimensions {
		f.Dimensions[i].CreatedAt = future
	}
	for i := range f.Modules {
		f.Modules[i].CreatedAt = future
	}
	before := managementTraitsResultValidationSnapshot(t, f)
	got, err := managementTraitsResultValidationCall(f, managementTraitsResultValidationBudget(f))
	if err != nil || !reflect.DeepEqual(got, want) || !bytes.Equal(before, managementTraitsResultValidationSnapshot(t, f)) {
		t.Fatal("deferred lifecycle was treated as authorization/submission", err)
	}
	b, _ := json.Marshal(got)
	for _, s := range []string{"approved", "authorized", "submittedAt", "source", "participantId"} {
		if bytes.Contains(b, []byte(s)) {
			t.Fatal("result grants deferred authority")
		}
	}
}
