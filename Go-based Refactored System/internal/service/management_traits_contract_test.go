package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func managementTraitsContractFixture(t *testing.T, questionnaire string) (ManagementTraitsManifest, ManagementTraitsMapping, ManagementTraitsSnapshotInput) {
	t.Helper()
	m := ManagementTraitsManifest{Questionnaire: questionnaire, Versions: ManagementTraitsScoringVersions{Product: "test-product", Question: "test-question", Scoring: "test-scoring", Norm: "test-norm"}, Questions: make([]ManagementTraitsManifestQuestion, 140)}
	texts := []string{"不符合", "不太符合", "一般", "比较符合", "很符合"}
	for _, d := range managementTraitsFixture() {
		for _, n := range d.numbers {
			q := ManagementTraitsManifestQuestion{Number: n, DimensionKey: d.key, Reverse: managementTraitsFixtureIsReverse(d, n), Content: fmt.Sprintf(" synthetic V%d <&> \n", n), Options: make([]ManagementTraitsManifestOption, 5)}
			if questionnaire == "leader" && (n == 67 || n == 96) {
				q.Content += "leader synthetic"
			}
			for i := range q.Options {
				q.Options[i] = ManagementTraitsManifestOption{Raw: i + 1, Content: texts[i], DisplayOrder: 5 - i}
			}
			m.Questions[n-1] = q
		}
	}
	c, err := CanonicalManagementTraitsManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	mp := ManagementTraitsMapping{Questionnaire: questionnaire, ManifestSHA: c.SHA256, Questions: make([]ManagementTraitsMappedQuestion, 140)}
	in := ManagementTraitsSnapshotInput{PaperID: "opaque-paper", ExamID: "opaque-exam", ParticipantType: "candidate", ParticipantID: "opaque-participant", Questionnaire: questionnaire, ManifestSHA: c.SHA256, Answers: make([]ManagementTraitsSnapshotAnswer, 140)}
	for i, q := range m.Questions {
		r := ManagementTraitsMappedQuestion{Number: q.Number, SourceQuestionID: fmt.Sprintf("source-q-%d", q.Number), Content: q.Content, Options: make([]ManagementTraitsMappedOption, 5)}
		for j, o := range q.Options {
			r.Options[j] = ManagementTraitsMappedOption{SourceOptionID: fmt.Sprintf("source-o-%d-%d", q.Number, o.Raw), Raw: o.Raw, Content: o.Content, DisplayOrder: o.DisplayOrder}
		}
		mp.Questions[i] = r
		in.Answers[i] = ManagementTraitsSnapshotAnswer{Number: q.Number, PaperQuestionID: fmt.Sprintf("paper-q-%d", q.Number), SourceQuestionID: r.SourceQuestionID, DisplayOrder: 141 - q.Number, Answered: true, SelectedOptionID: r.Options[2].SourceOptionID, Raw: 3}
	}
	c, err = CanonicalManagementTraitsMapping(m, mp)
	if err != nil {
		t.Fatal(err)
	}
	in.MappingSHA = c.SHA256
	return m, mp, in
}

func managementTraitsContractClone[T any](t *testing.T, v T) T {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out T
	if err = json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func managementTraitsContractCheck(t *testing.T, c ManagementTraitsCanonical, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(c.JSON)
	if c.SHA256 != hex.EncodeToString(h[:]) || len(c.SHA256) != 64 || !json.Valid(c.JSON) || bytes.Contains(c.JSON, []byte("\n")) {
		t.Fatal("invalid canonical bytes/hash")
	}
	var decoded interface{}
	if err := json.Unmarshal(c.JSON, &decoded); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(c.JSON, []byte(`"schema":`)) {
		t.Fatal("missing domain schema")
	}
}

func TestManagementTraitsContractCanonicalDeterminism(t *testing.T) {
	for _, questionnaire := range []string{"staff", "leader"} {
		t.Run(questionnaire, func(t *testing.T) {
			m, mp, in := managementTraitsContractFixture(t, questionnaire)
			beforeM, beforeMP, beforeIn := managementTraitsContractClone(t, m), managementTraitsContractClone(t, mp), managementTraitsContractClone(t, in)
			mc, err := CanonicalManagementTraitsManifest(m)
			managementTraitsContractCheck(t, mc, err)
			pc, err := CanonicalManagementTraitsMapping(m, mp)
			managementTraitsContractCheck(t, pc, err)
			ic, err := CanonicalManagementTraitsInput(m, mp, in)
			managementTraitsContractCheck(t, ic.Canonical, err)
			if len(ic.Answers) != 140 {
				t.Fatal("missing S1 answers")
			}
			for i, a := range ic.Answers {
				if a != (ManagementTraitsAnswer{Number: i + 1, Answered: true, Raw: 3}) {
					t.Fatal("S1 input differs")
				}
			}
			if !reflect.DeepEqual(m, beforeM) || !reflect.DeepEqual(mp, beforeMP) || !reflect.DeepEqual(in, beforeIn) {
				t.Fatal("boundary mutated input")
			}
			for i := range m.Questions {
				for j := 0; j < 2; j++ {
					m.Questions[i].Options[j], m.Questions[i].Options[4-j] = m.Questions[i].Options[4-j], m.Questions[i].Options[j]
					mp.Questions[i].Options[j], mp.Questions[i].Options[4-j] = mp.Questions[i].Options[4-j], mp.Questions[i].Options[j]
				}
			}
			for i := 0; i < 70; i++ {
				m.Questions[i], m.Questions[139-i] = m.Questions[139-i], m.Questions[i]
				mp.Questions[i], mp.Questions[139-i] = mp.Questions[139-i], mp.Questions[i]
				in.Answers[i], in.Answers[139-i] = in.Answers[139-i], in.Answers[i]
			}
			beforeM, beforeMP, beforeIn = managementTraitsContractClone(t, m), managementTraitsContractClone(t, mp), managementTraitsContractClone(t, in)
			for repeat := 0; repeat < 3; repeat++ {
				a, err := CanonicalManagementTraitsManifest(m)
				managementTraitsContractCheck(t, a, err)
				b, err := CanonicalManagementTraitsMapping(m, mp)
				managementTraitsContractCheck(t, b, err)
				c, err := CanonicalManagementTraitsInput(m, mp, in)
				managementTraitsContractCheck(t, c.Canonical, err)
				if !reflect.DeepEqual(a, mc) || !reflect.DeepEqual(b, pc) || !reflect.DeepEqual(c, ic) {
					t.Fatal("reordering/repeat changed canonical")
				}
				a.JSON[0], b.JSON[0], c.Canonical.JSON[0] = 'x', 'x', 'x'
				c.Answers[0].Raw = 5
			}
			if !reflect.DeepEqual(m, beforeM) || !reflect.DeepEqual(mp, beforeMP) || !reflect.DeepEqual(in, beforeIn) {
				t.Fatal("sorting mutated input")
			}
			var payload struct {
				Rows []struct {
					Number      int
					QuestionSHA string `json:"questionSha"`
				}
			}
			if err := json.Unmarshal(ic.Canonical.JSON, &payload); err != nil {
				t.Fatal(err)
			}
			if len(payload.Rows) != 140 {
				t.Fatal("missing frozen rows")
			}
			for i, r := range payload.Rows {
				if r.Number != i+1 || len(r.QuestionSHA) != 64 {
					t.Fatal("invalid per-question SHA/order")
				}
				if _, err := hex.DecodeString(r.QuestionSHA); err != nil || r.QuestionSHA != strings.ToLower(r.QuestionSHA) {
					t.Fatal("invalid question SHA")
				}
			}
		})
	}
}

func TestManagementTraitsContractManifestPolicy(t *testing.T) {
	m, _, _ := managementTraitsContractFixture(t, "staff")
	c, err := CanonicalManagementTraitsManifest(m)
	managementTraitsContractCheck(t, c, err)
	var p struct {
		Schema     string
		Versions   ManagementTraitsScoringVersions
		Dimensions []struct {
			Key, Module, Norm string
			Order             int
			Items             []struct {
				Number  int
				Reverse bool
			}
		}
		Policy struct {
			Dimension, Modules, Overall, OverallNorm, NormSum, Reverse, Incomplete, Precision, Tie string
			GradeBounds                                                                            []int `json:"gradeBounds"`
			Levels                                                                                 []string
			Highest, Lowest                                                                        int
			Overlap                                                                                bool
		}
	}
	if err := json.Unmarshal(c.JSON, &p); err != nil {
		t.Fatal(err)
	}
	if p.Schema != "mng-scoring-manifest-v1" || p.Versions != m.Versions || len(p.Dimensions) != 13 {
		t.Fatal("manifest domain/versions/dimensions")
	}
	for i, d := range managementTraitsFixture() {
		a := p.Dimensions[i]
		if a.Key != d.key || a.Module != d.module || a.Order != i+1 || a.Norm != d.norm.RatString() || len(a.Items) != len(d.numbers) {
			t.Fatal("normative dimension mismatch")
		}
		for j, n := range d.numbers {
			if a.Items[j].Number != n || a.Items[j].Reverse != managementTraitsFixtureIsReverse(d, n) {
				t.Fatal("normative V direction mismatch")
			}
		}
	}
	if p.Policy.Dimension != "25*(sum-count)/count" || p.Policy.Modules != "equal-member-dimensions" || p.Policy.Overall != "equal-13-dimensions" || p.Policy.OverallNorm != "705/13" || p.Policy.NormSum != "705" || p.Policy.Reverse != "6-raw-once" || p.Policy.Incomplete != "no-formal-scores" || p.Policy.Precision != "exact-rational;display-half-up-2" || p.Policy.Tie != "dimension-order-ascending;unrounded;no-grade-filter" || p.Policy.Highest != 3 || p.Policy.Lowest != 3 || !p.Policy.Overlap || !reflect.DeepEqual(p.Policy.GradeBounds, []int{90, 70, 30, 10}) || !reflect.DeepEqual(p.Policy.Levels, []string{"excellent", "good", "qualified", "weak", "insufficient"}) {
		t.Fatal("normative scoring policy mismatch")
	}
	if !bytes.Contains(c.JSON, []byte(`\u003c\u0026\u003e`)) || !bytes.Contains(c.JSON, []byte(` synthetic V1`)) {
		t.Fatal("default JSON escaping/original whitespace not preserved")
	}
	for _, v := range []interface{}{ManagementTraitsScoringVersions{}, ManagementTraitsManifest{}, ManagementTraitsMapping{}, ManagementTraitsSnapshotInput{}, ManagementTraitsSnapshotAnswer{}, ManagementTraitsValidatedInput{}} {
		typ := reflect.TypeOf(v)
		for i := 0; i < typ.NumField(); i++ {
			name := strings.ToLower(typ.Field(i).Name)
			for _, excluded := range []string{"contentversion", "template", "approved", "time", "name", "address", "metadata", "final", "score"} {
				if strings.Contains(name, excluded) {
					t.Fatalf("rendering/result field %s", name)
				}
			}
		}
	}
}

func TestManagementTraitsContractHashSensitivity(t *testing.T) {
	m, mp, in := managementTraitsContractFixture(t, "staff")
	baseM, _ := CanonicalManagementTraitsManifest(m)
	baseMP, _ := CanonicalManagementTraitsMapping(m, mp)
	baseIn, _ := CanonicalManagementTraitsInput(m, mp, in)
	for _, change := range []func(*ManagementTraitsManifest){
		func(m *ManagementTraitsManifest) { m.Questions[66].Content += "changed" },
		func(m *ManagementTraitsManifest) { m.Questions[0].Content += " " },
		func(m *ManagementTraitsManifest) { m.Questionnaire = "leader" },
		func(m *ManagementTraitsManifest) { m.Versions.Product = "future.1_unknown" },
		func(m *ManagementTraitsManifest) { m.Versions.Question = "future-q" },
		func(m *ManagementTraitsManifest) { m.Versions.Scoring = "future-s" },
		func(m *ManagementTraitsManifest) { m.Versions.Norm = "future-n" },
		func(m *ManagementTraitsManifest) {
			m.Questions[0].Options[0].DisplayOrder, m.Questions[0].Options[1].DisplayOrder = m.Questions[0].Options[1].DisplayOrder, m.Questions[0].Options[0].DisplayOrder
		},
	} {
		x := managementTraitsContractClone(t, m)
		change(&x)
		c, err := CanonicalManagementTraitsManifest(x)
		managementTraitsContractCheck(t, c, err)
		if c.SHA256 == baseM.SHA256 {
			t.Fatal("manifest change not hashed")
		}
	}
	leader, _, _ := managementTraitsContractFixture(t, "leader")
	leader.Questionnaire = "staff"
	c, _ := CanonicalManagementTraitsManifest(leader)
	if c.SHA256 == baseM.SHA256 {
		t.Fatal("synthetic two-text difference lost")
	}
	for _, change := range []func(*ManagementTraitsMapping){func(x *ManagementTraitsMapping) { x.Questions[0].SourceQuestionID = "another-q" }, func(x *ManagementTraitsMapping) { x.Questions[0].Options[0].SourceOptionID = "another-option" }} {
		x := managementTraitsContractClone(t, mp)
		change(&x)
		c, err := CanonicalManagementTraitsMapping(m, x)
		managementTraitsContractCheck(t, c, err)
		if c.SHA256 == baseMP.SHA256 {
			t.Fatal("source identity not hashed")
		}
	}
	for _, change := range []func(*ManagementTraitsSnapshotInput){
		func(x *ManagementTraitsSnapshotInput) { x.PaperID += "x" }, func(x *ManagementTraitsSnapshotInput) { x.ExamID += "x" }, func(x *ManagementTraitsSnapshotInput) { x.ParticipantID += "x" }, func(x *ManagementTraitsSnapshotInput) { x.ParticipantType = "tester" },
		func(x *ManagementTraitsSnapshotInput) { x.Answers[0].PaperQuestionID += "x" },
		func(x *ManagementTraitsSnapshotInput) {
			x.Answers[0].DisplayOrder, x.Answers[1].DisplayOrder = x.Answers[1].DisplayOrder, x.Answers[0].DisplayOrder
		},
		func(x *ManagementTraitsSnapshotInput) {
			x.Answers[0].Raw = 4
			x.Answers[0].SelectedOptionID = mp.Questions[0].Options[3].SourceOptionID
		},
		func(x *ManagementTraitsSnapshotInput) {
			x.Answers[0].Answered = false
			x.Answers[0].Raw = 0
			x.Answers[0].SelectedOptionID = ""
		},
	} {
		x := managementTraitsContractClone(t, in)
		change(&x)
		c, err := CanonicalManagementTraitsInput(m, mp, x)
		managementTraitsContractCheck(t, c.Canonical, err)
		if c.Canonical.SHA256 == baseIn.Canonical.SHA256 {
			t.Fatal("input change not hashed")
		}
	}
}

func TestManagementTraitsContractManifestRejects(t *testing.T) {
	m, _, _ := managementTraitsContractFixture(t, "staff")
	cases := []struct {
		name   string
		change func(*ManagementTraitsManifest)
	}{
		{"nil", func(x *ManagementTraitsManifest) { x.Questions = nil }}, {"139", func(x *ManagementTraitsManifest) { x.Questions = x.Questions[:139] }}, {"141", func(x *ManagementTraitsManifest) { x.Questions = append(x.Questions, x.Questions[0]) }},
		{"duplicate", func(x *ManagementTraitsManifest) { x.Questions[1].Number = 1 }}, {"zero", func(x *ManagementTraitsManifest) { x.Questions[0].Number = 0 }}, {"out", func(x *ManagementTraitsManifest) { x.Questions[0].Number = 141 }},
		{"dimension", func(x *ManagementTraitsManifest) { x.Questions[0].DimensionKey = "self_confidence" }}, {"reverse", func(x *ManagementTraitsManifest) { x.Questions[0].Reverse = !x.Questions[0].Reverse }},
		{"blank", func(x *ManagementTraitsManifest) { x.Questions[0].Content = " \n\t" }}, {"utf8", func(x *ManagementTraitsManifest) { x.Questions[0].Content = string([]byte{255}) }},
		{"options_nil", func(x *ManagementTraitsManifest) { x.Questions[0].Options = nil }}, {"options_four", func(x *ManagementTraitsManifest) { x.Questions[0].Options = x.Questions[0].Options[:4] }}, {"options_six", func(x *ManagementTraitsManifest) {
			x.Questions[0].Options = append(x.Questions[0].Options, x.Questions[0].Options[0])
		}},
		{"raw_dup", func(x *ManagementTraitsManifest) { x.Questions[0].Options[1].Raw = 1 }}, {"raw_zero", func(x *ManagementTraitsManifest) { x.Questions[0].Options[0].Raw = 0 }}, {"raw_six", func(x *ManagementTraitsManifest) { x.Questions[0].Options[0].Raw = 6 }},
		{"order_dup", func(x *ManagementTraitsManifest) { x.Questions[0].Options[1].DisplayOrder = 5 }}, {"order_zero", func(x *ManagementTraitsManifest) { x.Questions[0].Options[0].DisplayOrder = 0 }}, {"order_six", func(x *ManagementTraitsManifest) { x.Questions[0].Options[0].DisplayOrder = 6 }},
		{"semantic_swap", func(x *ManagementTraitsManifest) {
			x.Questions[0].Options[0].Content, x.Questions[0].Options[4].Content = x.Questions[0].Options[4].Content, x.Questions[0].Options[0].Content
		}},
		{"option_blank", func(x *ManagementTraitsManifest) { x.Questions[0].Options[0].Content = "" }}, {"option_utf8", func(x *ManagementTraitsManifest) { x.Questions[0].Options[0].Content = string([]byte{255}) }},
	}
	for _, q := range []string{"", "STAFF", " staff", "00201", "unknown"} {
		q := q
		cases = append(cases, struct {
			name   string
			change func(*ManagementTraitsManifest)
		}{"questionnaire_" + q, func(x *ManagementTraitsManifest) { x.Questionnaire = q }})
	}
	for axis := 0; axis < 4; axis++ {
		for j, v := range []string{"", " ", " test", "test ", "TEST", "a/b", "a\x00b", "é", ".start", strings.Repeat("a", 65)} {
			axis, v := axis, v
			cases = append(cases, struct {
				name   string
				change func(*ManagementTraitsManifest)
			}{fmt.Sprintf("version_%d_%d", axis, j), func(x *ManagementTraitsManifest) {
				fields := []*string{&x.Versions.Product, &x.Versions.Question, &x.Versions.Scoring, &x.Versions.Norm}
				*fields[axis] = v
			}})
		}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			x := managementTraitsContractClone(t, m)
			tc.change(&x)
			c, err := CanonicalManagementTraitsManifest(x)
			if err == nil || !reflect.DeepEqual(c, ManagementTraitsCanonical{}) {
				t.Fatal("invalid manifest accepted/partial output")
			}
			if strings.Contains(err.Error(), "synthetic V1") {
				t.Fatal("error echoed input")
			}
		})
	}
}

func TestManagementTraitsContractMappingRejects(t *testing.T) {
	m, mp, _ := managementTraitsContractFixture(t, "staff")
	cases := []struct {
		name   string
		change func(*ManagementTraitsMapping)
	}{
		{"questionnaire", func(x *ManagementTraitsMapping) { x.Questionnaire = "leader" }}, {"hash_empty", func(x *ManagementTraitsMapping) { x.ManifestSHA = "" }}, {"hash_arbitrary", func(x *ManagementTraitsMapping) { x.ManifestSHA = strings.Repeat("a", 64) }}, {"hash_case", func(x *ManagementTraitsMapping) { x.ManifestSHA = strings.ToUpper(x.ManifestSHA) }},
		{"nil", func(x *ManagementTraitsMapping) { x.Questions = nil }}, {"139", func(x *ManagementTraitsMapping) { x.Questions = x.Questions[:139] }}, {"141", func(x *ManagementTraitsMapping) { x.Questions = append(x.Questions, x.Questions[0]) }}, {"dup_v", func(x *ManagementTraitsMapping) { x.Questions[1].Number = 1 }}, {"zero", func(x *ManagementTraitsMapping) { x.Questions[0].Number = 0 }}, {"out", func(x *ManagementTraitsMapping) { x.Questions[0].Number = 141 }},
		{"dup_qid", func(x *ManagementTraitsMapping) { x.Questions[1].SourceQuestionID = x.Questions[0].SourceQuestionID }}, {"content", func(x *ManagementTraitsMapping) { x.Questions[0].Content += " " }},
		{"opts_missing", func(x *ManagementTraitsMapping) { x.Questions[0].Options = x.Questions[0].Options[:4] }}, {"opts_extra", func(x *ManagementTraitsMapping) {
			x.Questions[0].Options = append(x.Questions[0].Options, x.Questions[0].Options[0])
		}},
		{"raw", func(x *ManagementTraitsMapping) { x.Questions[0].Options[0].Raw = 6 }}, {"raw_dup", func(x *ManagementTraitsMapping) { x.Questions[0].Options[0].Raw = 2 }}, {"text", func(x *ManagementTraitsMapping) { x.Questions[0].Options[0].Content += " " }}, {"order", func(x *ManagementTraitsMapping) { x.Questions[0].Options[0].DisplayOrder = 1 }},
		{"dup_oid", func(x *ManagementTraitsMapping) {
			x.Questions[0].Options[1].SourceOptionID = x.Questions[0].Options[0].SourceOptionID
		}}, {"cross_v_oid", func(x *ManagementTraitsMapping) {
			x.Questions[1].Options[0].SourceOptionID = x.Questions[0].Options[0].SourceOptionID
		}},
	}
	for i, id := range []string{"", " ", " q", "q ", "q\x00x", "q\nx", string([]byte{255}), strings.Repeat("a", 65)} {
		id := id
		for _, option := range []bool{false, true} {
			option := option
			cases = append(cases, struct {
				name   string
				change func(*ManagementTraitsMapping)
			}{fmt.Sprintf("id_%d_%v", i, option), func(x *ManagementTraitsMapping) {
				if option {
					x.Questions[0].Options[0].SourceOptionID = id
				} else {
					x.Questions[0].SourceQuestionID = id
				}
			}})
		}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			x := managementTraitsContractClone(t, mp)
			tc.change(&x)
			c, err := CanonicalManagementTraitsMapping(m, x)
			if err == nil || !reflect.DeepEqual(c, ManagementTraitsCanonical{}) {
				t.Fatal("invalid mapping accepted/partial")
			}
			if strings.Contains(err.Error(), "source-") || strings.Contains(err.Error(), "synthetic") {
				t.Fatal("error echoed source")
			}
		})
	}
	m.Questions[0].Reverse = !m.Questions[0].Reverse
	c, err := CanonicalManagementTraitsMapping(m, mp)
	if err == nil || !reflect.DeepEqual(c, ManagementTraitsCanonical{}) {
		t.Fatal("manifest boundary not reused")
	}
}

func TestManagementTraitsContractInputRejects(t *testing.T) {
	m, mp, in := managementTraitsContractFixture(t, "staff")
	cases := []struct {
		name   string
		change func(*ManagementTraitsSnapshotInput)
	}{
		{"participant", func(x *ManagementTraitsSnapshotInput) { x.ParticipantType = "user" }}, {"questionnaire", func(x *ManagementTraitsSnapshotInput) { x.Questionnaire = "leader" }}, {"manifest", func(x *ManagementTraitsSnapshotInput) { x.ManifestSHA = strings.Repeat("a", 64) }}, {"mapping", func(x *ManagementTraitsSnapshotInput) { x.MappingSHA = strings.Repeat("b", 64) }}, {"manifest_empty", func(x *ManagementTraitsSnapshotInput) { x.ManifestSHA = "" }}, {"mapping_case", func(x *ManagementTraitsSnapshotInput) { x.MappingSHA = strings.ToUpper(x.MappingSHA) }},
		{"nil", func(x *ManagementTraitsSnapshotInput) { x.Answers = nil }}, {"139", func(x *ManagementTraitsSnapshotInput) { x.Answers = x.Answers[:139] }}, {"141", func(x *ManagementTraitsSnapshotInput) { x.Answers = append(x.Answers, x.Answers[0]) }}, {"dup_v", func(x *ManagementTraitsSnapshotInput) { x.Answers[1].Number = 1 }}, {"zero", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].Number = 0 }}, {"out", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].Number = 141 }},
		{"dup_pqid", func(x *ManagementTraitsSnapshotInput) { x.Answers[1].PaperQuestionID = x.Answers[0].PaperQuestionID }}, {"dup_order", func(x *ManagementTraitsSnapshotInput) { x.Answers[1].DisplayOrder = x.Answers[0].DisplayOrder }}, {"order_zero", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].DisplayOrder = 0 }}, {"order_out", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].DisplayOrder = 141 }},
		{"cross_source", func(x *ManagementTraitsSnapshotInput) {
			x.Answers[0].SourceQuestionID = mp.Questions[1].SourceQuestionID
		}}, {"cross_option", func(x *ManagementTraitsSnapshotInput) {
			x.Answers[0].SelectedOptionID = mp.Questions[1].Options[2].SourceOptionID
		}}, {"empty_selected", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].SelectedOptionID = "" }}, {"unknown_selected", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].SelectedOptionID = "opaque-unknown" }}, {"raw_mismatch", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].Raw = 4 }}, {"raw_zero", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].Raw = 0 }}, {"raw_six", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].Raw = 6 }},
		{"unanswered_raw", func(x *ManagementTraitsSnapshotInput) {
			x.Answers[0].Answered = false
			x.Answers[0].SelectedOptionID = ""
			x.Answers[0].Raw = 1
		}}, {"unanswered_selected", func(x *ManagementTraitsSnapshotInput) { x.Answers[0].Answered = false; x.Answers[0].Raw = 0 }},
	}
	for i, id := range []string{"", " ", " q", "q ", "q\x00x", "q\tx", string([]byte{255}), strings.Repeat("a", 65)} {
		for axis := 0; axis < 4; axis++ {
			axis, id := axis, id
			cases = append(cases, struct {
				name   string
				change func(*ManagementTraitsSnapshotInput)
			}{fmt.Sprintf("id_%d_%d", i, axis), func(x *ManagementTraitsSnapshotInput) {
				ids := []*string{&x.PaperID, &x.ExamID, &x.ParticipantID, &x.Answers[0].PaperQuestionID}
				*ids[axis] = id
			}})
		}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			x := managementTraitsContractClone(t, in)
			tc.change(&x)
			c, err := CanonicalManagementTraitsInput(m, mp, x)
			if err == nil || !reflect.DeepEqual(c, ManagementTraitsValidatedInput{}) {
				t.Fatal("invalid input accepted/partial")
			}
			if strings.Contains(err.Error(), "opaque-") || strings.Contains(err.Error(), "source-") {
				t.Fatal("error echoed identity")
			}
		})
	}
	mp.Questions[0].Options[0].Content = "wrong"
	c, err := CanonicalManagementTraitsInput(m, mp, in)
	if err == nil || !reflect.DeepEqual(c, ManagementTraitsValidatedInput{}) {
		t.Fatal("mapping boundary not reused")
	}
}

func TestManagementTraitsContractIncompleteAndTieDelegation(t *testing.T) {
	for _, final := range []int{1, 5} {
		m, mp, in := managementTraitsContractFixture(t, "staff")
		for _, a := range managementTraitsFixtureFinalAnswers(final) {
			in.Answers[a.Number-1].Raw = a.Raw
			in.Answers[a.Number-1].SelectedOptionID = mp.Questions[a.Number-1].Options[a.Raw-1].SourceOptionID
		}
		v, err := CanonicalManagementTraitsInput(m, mp, in)
		if err != nil {
			t.Fatal(err)
		}
		s, err := CalculateManagementTraits("staff", v.Answers)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"self_confidence", "emotional_stability", "self_discipline"}
		if !reflect.DeepEqual(s.Highest, want) || !reflect.DeepEqual(s.Lowest, want) {
			t.Fatal("extreme ties changed")
		}
		in.Answers[119].Answered = false
		in.Answers[119].Raw = 0
		in.Answers[119].SelectedOptionID = ""
		v, err = CanonicalManagementTraitsInput(m, mp, in)
		managementTraitsContractCheck(t, v.Canonical, err)
		if v.Answers == nil || len(v.Answers) != 140 || v.Answers[119] != (ManagementTraitsAnswer{Number: 120}) {
			t.Fatal("incomplete answer lost")
		}
		s, err = CalculateManagementTraits("staff", v.Answers)
		if err != nil || s.IsComplete || s.OverallScore != nil || s.OverallNorm != nil {
			t.Fatal("incomplete formal score")
		}
		for _, d := range s.Dimensions {
			if d.Score != nil || d.Norm != nil || d.Level != "" {
				t.Fatal("incomplete dimension formal score")
			}
		}
	}
}

func TestManagementTraitsContractCanonicalRoundtripAndFrozenHash(t *testing.T) {
	m, mp, in := managementTraitsContractFixture(t, "staff")
	mc, err := CanonicalManagementTraitsManifest(m)
	managementTraitsContractCheck(t, mc, err)
	pc, err := CanonicalManagementTraitsMapping(m, mp)
	managementTraitsContractCheck(t, pc, err)
	ic, err := CanonicalManagementTraitsInput(m, mp, in)
	managementTraitsContractCheck(t, ic.Canonical, err)
	for _, tc := range []struct {
		c       ManagementTraitsCanonical
		payload interface{}
	}{
		{mc, &managementTraitsCanonicalManifest{}}, {pc, &managementTraitsCanonicalMapping{}}, {ic.Canonical, &managementTraitsCanonicalInput{}},
	} {
		if err := json.Unmarshal(tc.c.JSON, tc.payload); err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(tc.payload)
		if err != nil || !bytes.Equal(b, tc.c.JSON) {
			t.Fatal("canonical typed roundtrip differs")
		}
	}
	var payload struct {
		Rows []struct {
			Number      int
			QuestionSHA string `json:"questionSha"`
		}
	}
	if err := json.Unmarshal(ic.Canonical.JSON, &payload); err != nil {
		t.Fatal(err)
	}
	for i, a := range in.Answers {
		// Independent shape and standard SHA, not the production hashing helper.
		frozen := struct {
			Schema          string                         `json:"schema"`
			ManifestSHA     string                         `json:"manifestSha"`
			MappingSHA      string                         `json:"mappingSha"`
			Question        ManagementTraitsMappedQuestion `json:"question"`
			DimensionKey    string                         `json:"dimensionKey"`
			Reverse         bool                           `json:"reverse"`
			PaperQuestionID string                         `json:"paperQuestionId"`
			DisplayOrder    int                            `json:"displayOrder"`
		}{"mng-frozen-question-v1", mc.SHA256, pc.SHA256, mp.Questions[i], m.Questions[i].DimensionKey, m.Questions[i].Reverse, a.PaperQuestionID, a.DisplayOrder}
		b, err := json.Marshal(frozen)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if payload.Rows[i].QuestionSHA != hex.EncodeToString(h[:]) {
			t.Fatal("frozen question hash omitted facts")
		}
	}
	var policy struct {
		Policy struct {
			ModuleOrder []string `json:"moduleOrder"`
		}
	}
	if err := json.Unmarshal(mc.JSON, &policy); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(policy.Policy.ModuleOrder, []string{"self", "interpersonal", "task", "development"}) {
		t.Fatal("S1 calculation order missing")
	}
	if !bytes.HasPrefix(mc.JSON, []byte(`{"schema":"mng-scoring-manifest-v1","questionnaire":"staff","versions":{"product":`)) || !bytes.HasPrefix(pc.JSON, []byte(`{"schema":"mng-source-mapping-v1",`)) || !bytes.HasPrefix(ic.Canonical.JSON, []byte(`{"schema":"mng-score-input-v1",`)) {
		t.Fatal("canonical domain/key order changed")
	}
}

func TestManagementTraitsContractLinkedHashChanges(t *testing.T) {
	for kind := 0; kind < 5; kind++ {
		m, mp, in := managementTraitsContractFixture(t, "staff")
		base, err := CanonicalManagementTraitsInput(m, mp, in)
		if err != nil {
			t.Fatal(err)
		}
		switch kind {
		case 0:
			m.Versions.Scoring = "another-test-policy"
		case 1:
			m.Questions[0].Content += " revised"
			mp.Questions[0].Content = m.Questions[0].Content
		case 2:
			mp.Questions[0].SourceQuestionID = "opaque-new-source"
			in.Answers[0].SourceQuestionID = mp.Questions[0].SourceQuestionID
		case 3:
			mp.Questions[0].Options[2].SourceOptionID = "opaque-new-option"
			in.Answers[0].SelectedOptionID = mp.Questions[0].Options[2].SourceOptionID
		case 4:
			m.Questions[0].Options[0].DisplayOrder, m.Questions[0].Options[1].DisplayOrder = m.Questions[0].Options[1].DisplayOrder, m.Questions[0].Options[0].DisplayOrder
			mp.Questions[0].Options[0].DisplayOrder, mp.Questions[0].Options[1].DisplayOrder = mp.Questions[0].Options[1].DisplayOrder, mp.Questions[0].Options[0].DisplayOrder
		}
		mc, err := CanonicalManagementTraitsManifest(m)
		if err != nil {
			t.Fatal(err)
		}
		mp.ManifestSHA = mc.SHA256
		in.ManifestSHA = mc.SHA256
		pc, err := CanonicalManagementTraitsMapping(m, mp)
		if err != nil {
			t.Fatal(err)
		}
		in.MappingSHA = pc.SHA256
		got, err := CanonicalManagementTraitsInput(m, mp, in)
		managementTraitsContractCheck(t, got.Canonical, err)
		if got.Canonical.SHA256 == base.Canonical.SHA256 {
			t.Fatal("linked change not reflected in input hash")
		}
		var a, b struct {
			Rows []struct {
				QuestionSHA string `json:"questionSha"`
			}
		}
		if err := json.Unmarshal(base.Canonical.JSON, &a); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(got.Canonical.JSON, &b); err != nil {
			t.Fatal(err)
		}
		if a.Rows[0].QuestionSHA == b.Rows[0].QuestionSHA {
			t.Fatal("linked change not reflected in frozen question hash")
		}
	}
}

func TestManagementTraitsContractOpaqueIDsAndSyntaxLimits(t *testing.T) {
	m, mp, in := managementTraitsContractFixture(t, "staff")
	m.Versions = ManagementTraitsScoringVersions{Product: strings.Repeat("a", 64), Question: "0._-future", Scoring: "unapproved-syntax-only", Norm: "unknown-norm"}
	mc, err := CanonicalManagementTraitsManifest(m)
	managementTraitsContractCheck(t, mc, err)
	mp.ManifestSHA, in.ManifestSHA = mc.SHA256, mc.SHA256
	mp.Questions[0].SourceQuestionID = strings.Repeat("界", 21) + "x" // 64 UTF-8 bytes, opaque nonnumeric identity.
	mp.Questions[0].Options[0].SourceOptionID = strings.Repeat("o", 64)
	in.Answers[0].SourceQuestionID = mp.Questions[0].SourceQuestionID
	in.PaperID, in.ExamID, in.ParticipantID, in.Answers[0].PaperQuestionID = "paper 内部", "非数字", "101", strings.Repeat("p", 64)
	pc, err := CanonicalManagementTraitsMapping(m, mp)
	managementTraitsContractCheck(t, pc, err)
	in.MappingSHA = pc.SHA256
	v, err := CanonicalManagementTraitsInput(m, mp, in)
	managementTraitsContractCheck(t, v.Canonical, err)
	for _, id := range []string{"\u0085", "q\u007fx", "\u00a0q", "q\u00a0", strings.Repeat("界", 22)} {
		x := managementTraitsContractClone(t, mp)
		x.Questions[0].SourceQuestionID = id
		c, err := CanonicalManagementTraitsMapping(m, x)
		if err == nil || !reflect.DeepEqual(c, ManagementTraitsCanonical{}) {
			t.Fatal("invalid opaque ID accepted")
		}
	}
	for _, selected := range []string{"\x00secret-token", string([]byte{255}), " option ", strings.Repeat("x", 65)} {
		x := managementTraitsContractClone(t, in)
		x.Answers[0].SelectedOptionID = selected
		v, err := CanonicalManagementTraitsInput(m, mp, x)
		if err == nil || !reflect.DeepEqual(v, ManagementTraitsValidatedInput{}) || strings.Contains(err.Error(), "secret-token") {
			t.Fatal("malformed selection accepted/leaked")
		}
	}
}
