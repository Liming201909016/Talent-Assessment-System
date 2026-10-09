package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/talent-assessment/refactored/internal/model"
)

func managementTraitsDecodeJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func managementTraitsDecodeObject(t *testing.T, b []byte) map[string]interface{} {
	t.Helper()
	var v map[string]interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func managementTraitsDecodeError(t *testing.T, got interface{}, zero interface{}, err error) {
	t.Helper()
	if err == nil || !reflect.DeepEqual(got, zero) {
		t.Fatal("invalid input accepted or partial result returned")
	}
	// All public failures are uniform, including errors delegated to S2B.
	if err.Error() != "invalid management traits stored input" {
		t.Fatalf("nonuniform error: %q", err.Error())
	}
}

func managementTraitsDecodeStoredFixture(t *testing.T, audience, participant string, raw int) (model.ManagementTraitsDefinitionBundle, model.ManagementTraitsPaperSnapshot, []model.ManagementTraitsPaperQuestionSnapshot, ManagementTraitsValidatedInput) {
	t.Helper()
	m, mp, in := managementTraitsContractFixture(t, audience)
	in.ParticipantType = participant
	for i := range in.Answers {
		in.Answers[i].Raw = raw
		in.Answers[i].SelectedOptionID = mp.Questions[i].Options[raw-1].SourceOptionID
	}
	mc, err := CanonicalManagementTraitsManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	pc, err := CanonicalManagementTraitsMapping(m, mp)
	if err != nil {
		t.Fatal(err)
	}
	ic, err := CanonicalManagementTraitsInput(m, mp, in)
	if err != nil {
		t.Fatal(err)
	}
	// Expected row hashes come from the pre-existing S2B contract, not S2C.
	var payload struct {
		Rows []struct {
			Number      int    `json:"number"`
			QuestionSHA string `json:"questionSha"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(ic.Canonical.JSON, &payload); err != nil {
		t.Fatal(err)
	}
	b := model.ManagementTraitsDefinitionBundle{ID: "bundle-fixture", ProductVersion: m.Versions.Product, QuestionVersion: m.Versions.Question, ScoringVersion: m.Versions.Scoring, NormVersion: m.Versions.Norm, Questionnaire: audience, ScoringManifest: string(mc.JSON), ScoringManifestSHA: mc.SHA256}
	p := model.ManagementTraitsPaperSnapshot{PaperID: in.PaperID, ExamID: in.ExamID, BundleID: b.ID, MappingSnapshot: string(pc.JSON), MappingSHA: pc.SHA256, ScoringManifestSHA: mc.SHA256, ParticipantType: participant, ParticipantID: in.ParticipantID}
	rows := make([]model.ManagementTraitsPaperQuestionSnapshot, 140)
	for i, a := range in.Answers {
		selected, answer, final := a.SelectedOptionID, a.Raw, a.Raw
		if m.Questions[i].Reverse {
			final = 6 - answer
		}
		rows[i] = model.ManagementTraitsPaperQuestionSnapshot{ID: fmt.Sprintf("snapshot-%d", i+1), PaperID: p.PaperID, PaperQuestionID: a.PaperQuestionID, SourceQuestionID: a.SourceQuestionID, Number: a.Number, DisplayOrder: a.DisplayOrder, DimensionKey: m.Questions[i].DimensionKey, Reverse: m.Questions[i].Reverse, Content: m.Questions[i].Content, OptionsSnapshot: string(managementTraitsDecodeJSON(t, mp.Questions[i].Options)), ScoringSnapshotSHA: payload.Rows[i].QuestionSHA, SelectedOptionID: &selected, RawAnswer: &answer, FinalScore: &final}
	}
	return b, p, rows, ic
}

func managementTraitsDecodeBudget(b model.ManagementTraitsDefinitionBundle, p model.ManagementTraitsPaperSnapshot, rows []model.ManagementTraitsPaperQuestionSnapshot) int {
	n := len(b.ScoringManifest) + len(p.MappingSnapshot)
	for _, r := range rows {
		n += len(r.OptionsSnapshot)
	}
	return n
}

func TestManagementTraitsDecodeRoundtrip(t *testing.T) {
	for _, audience := range []string{"staff", "leader"} {
		t.Run(audience, func(t *testing.T) {
			m, mp, _ := managementTraitsContractFixture(t, audience)
			mc, _ := CanonicalManagementTraitsManifest(m)
			pc, _ := CanonicalManagementTraitsMapping(m, mp)
			for _, reordered := range []bool{false, true} {
				mo, po := managementTraitsDecodeObject(t, mc.JSON), managementTraitsDecodeObject(t, pc.JSON)
				if reordered {
					for _, obj := range []map[string]interface{}{mo, po} {
						qs := obj["questions"].([]interface{})
						for i := 0; i < len(qs)/2; i++ {
							qs[i], qs[len(qs)-1-i] = qs[len(qs)-1-i], qs[i]
						}
						for _, q := range qs {
							opts := q.(map[string]interface{})["options"].([]interface{})
							opts[0], opts[4] = opts[4], opts[0]
						}
					}
				}
				mb, err := json.MarshalIndent(mo, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				pb, err := json.MarshalIndent(po, "", "\t")
				if err != nil {
					t.Fatal(err)
				}
				before := append([]byte(nil), mb...)
				decoded, err := DecodeManagementTraitsManifest(mb, len(mb))
				if err != nil || !reflect.DeepEqual(decoded, m) || !bytes.Equal(before, mb) {
					t.Fatal("manifest roundtrip or input mutation", err)
				}
				mapped, err := DecodeManagementTraitsMapping(pb, decoded, len(pb))
				if err != nil || !reflect.DeepEqual(mapped, mp) {
					t.Fatal("mapping roundtrip", err)
				}
				decoded.Questions[0].Content = "mutation"
				mapped.Questions[0].Options[0].Content = "mutation"
				fresh, err := DecodeManagementTraitsManifest(mb, len(mb))
				if err != nil || !reflect.DeepEqual(fresh, m) {
					t.Fatal("output polluted future decode")
				}
			}
			// Raw public-input JSON is not a stored canonical payload.
			got, err := DecodeManagementTraitsManifest(managementTraitsDecodeJSON(t, m), len(mc.JSON))
			managementTraitsDecodeError(t, got, ManagementTraitsManifest{}, err)
			mapped, err := DecodeManagementTraitsMapping(managementTraitsDecodeJSON(t, mp), m, len(pc.JSON))
			managementTraitsDecodeError(t, mapped, ManagementTraitsMapping{}, err)
		})
	}
}

func TestManagementTraitsDecodeStrictTokens(t *testing.T) {
	m, mp, _ := managementTraitsContractFixture(t, "staff")
	mc, _ := CanonicalManagementTraitsManifest(m)
	pc, _ := CanonicalManagementTraitsMapping(m, mp)
	for _, kind := range []string{"manifest", "mapping"} {
		t.Run(kind, func(t *testing.T) {
			base := mc.JSON
			if kind == "mapping" {
				base = pc.JSON
			}
			replace := func(a, b string) []byte { return bytes.Replace(base, []byte(a), []byte(b), 1) }
			cases := map[string][]byte{
				"duplicate":           append([]byte(`{"questionnaire":"staff",`), base[1:]...),
				"escaped_duplicate":   append([]byte(`{"qu\u0065stionnaire":"staff",`), base[1:]...),
				"unknown":             append([]byte(`{"secret-token":"private-value",`), base[1:]...),
				"casefold":            replace(`"questionnaire":`, `"Questionnaire":`),
				"alias":               replace(`"number":1`, `"Number":1`),
				"int_float":           replace(`"number":1,`, `"number":1.0,`),
				"int_exponent":        replace(`"number":1,`, `"number":1e0,`),
				"int_string":          replace(`"number":1,`, `"number":"1",`),
				"int_empty":           replace(`"number":1,`, `"number":"",`),
				"overflow":            replace(`"number":1,`, `"number":999999999999999999999999999,`),
				"utf8":                append([]byte{255}, base...),
				"bom":                 append([]byte{239, 187, 191}, base...),
				"trailing_object":     append(append([]byte(nil), base...), []byte(` {}`)...),
				"trailing_garbage":    append(append([]byte(nil), base...), 'x'),
				"prefix":              append([]byte(`x`), base...),
				"syntax":              base[:len(base)-1],
				"root_array":          []byte(`[]`),
				"root_number":         []byte(`1.0`),
				"root_string":         []byte(`"secret-token"`),
				"root_null":           []byte(`null`),
				"depth":               []byte(strings.Repeat("[", 33) + "0" + strings.Repeat("]", 33)),
				"high_surrogate":      replace(`"content":`, `"extra":"\uD800","content":`),
				"low_surrogate":       replace(`"content":`, `"extra":"\uDC00","content":`),
				"unpaired_in_content": replace(`"content":"`, `"content":"\uD800`),
				"nested_duplicate":    replace(`"number":1,`, `"number":1,"num\u0062er":1,`),
				"null_content":        replace(`"content":"不符合"`, `"content":null`),
			}
			cases["low_in_content"] = replace(`"content":"`, `"content":"\uDC00`)
			cases["wrong_pair"] = replace(`"content":"`, `"content":"\uD800\u0041`)
			cases["surrogate_key"] = replace(`"questionnaire":`, `"\uD800":`)
			cases["utf8_in_content"] = bytes.Replace(base, []byte(`"content":"`), []byte{'"', 'c', 'o', 'n', 't', 'e', 'n', 't', '"', ':', '"', 255}, 1)
			if kind == "manifest" {
				cases["bool_string"] = replace(`"reverse":false`, `"reverse":"false"`)
				cases["bool_empty"] = replace(`"reverse":false`, `"reverse":""`)
				cases["bool_null"] = replace(`"reverse":false`, `"reverse":null`)
				cases["missing_false"] = replace(`"reverse":false,`, "")
				cases["reverse_alias"] = replace(`"reverse":false`, `"Reverse":false`)
			}
			for name, data := range cases {
				t.Run(name, func(t *testing.T) {
					if bytes.Equal(data, base) {
						t.Fatal("test transformation missed target")
					}
					if kind == "manifest" {
						got, err := DecodeManagementTraitsManifest(data, len(data)+1)
						managementTraitsDecodeError(t, got, ManagementTraitsManifest{}, err)
					} else {
						got, err := DecodeManagementTraitsMapping(data, m, len(data)+1)
						managementTraitsDecodeError(t, got, ManagementTraitsMapping{}, err)
					}
				})
			}
			for _, limit := range []int{0, -1, len(base) - 1} {
				if kind == "manifest" {
					got, err := DecodeManagementTraitsManifest(base, limit)
					managementTraitsDecodeError(t, got, ManagementTraitsManifest{}, err)
				} else {
					got, err := DecodeManagementTraitsMapping(base, m, limit)
					managementTraitsDecodeError(t, got, ManagementTraitsMapping{}, err)
				}
			}
		})
	}
}

// Traverse one instance of every stored schema field; false and zero are still required.
func managementTraitsDecodeFieldPaths(v interface{}, prefix []string) [][]string {
	paths := make([][]string, 0)
	switch x := v.(type) {
	case map[string]interface{}:
		for key, value := range x {
			p := append(append([]string(nil), prefix...), key)
			paths = append(paths, p)
			paths = append(paths, managementTraitsDecodeFieldPaths(value, p)...)
		}
	case []interface{}:
		if len(x) > 0 {
			paths = append(paths, managementTraitsDecodeFieldPaths(x[0], append(append([]string(nil), prefix...), "0"))...)
		}
	}
	return paths
}

func TestManagementTraitsDecodeEveryFieldRequired(t *testing.T) {
	m, mp, _ := managementTraitsContractFixture(t, "staff")
	mc, _ := CanonicalManagementTraitsManifest(m)
	pc, _ := CanonicalManagementTraitsMapping(m, mp)
	for _, base := range []ManagementTraitsCanonical{mc, pc} {
		obj := managementTraitsDecodeObject(t, base.JSON)
		for _, path := range managementTraitsDecodeFieldPaths(obj, nil) {
			for _, mode := range []string{"missing", "null", "case", "unknown"} {
				t.Run(obj["schema"].(string)+"/"+strings.Join(path, "/")+"/"+mode, func(t *testing.T) {
					var current interface{} = managementTraitsDecodeObject(t, base.JSON)
					root := current
					for _, step := range path[:len(path)-1] {
						if step == "0" {
							current = current.([]interface{})[0]
						} else {
							current = current.(map[string]interface{})[step]
						}
					}
					parent := current.(map[string]interface{})
					key := path[len(path)-1]
					value := parent[key]
					delete(parent, key)
					switch mode {
					case "null":
						parent[key] = nil
					case "case":
						parent[strings.ToUpper(key[:1])+key[1:]] = value
					case "unknown":
						parent[key] = value
						parent["extra"] = "secret-token"
					}
					data := managementTraitsDecodeJSON(t, root)
					if base.SHA256 == mc.SHA256 {
						got, err := DecodeManagementTraitsManifest(data, len(data))
						managementTraitsDecodeError(t, got, ManagementTraitsManifest{}, err)
					} else {
						got, err := DecodeManagementTraitsMapping(data, m, len(data))
						managementTraitsDecodeError(t, got, ManagementTraitsMapping{}, err)
					}
				})
			}
		}
	}
}

func TestManagementTraitsDecodeNormativeAndMappingTamper(t *testing.T) {
	m, mp, _ := managementTraitsContractFixture(t, "staff")
	mc, _ := CanonicalManagementTraitsManifest(m)
	for _, change := range []func(map[string]interface{}){
		func(o map[string]interface{}) { o["schema"] = "another-domain" },
		func(o map[string]interface{}) { o["questionnaire"] = "STAFF" },
		func(o map[string]interface{}) { o["versions"].(map[string]interface{})["scoring"] = " invalid" },
		func(o map[string]interface{}) {
			o["dimensions"].([]interface{})[0].(map[string]interface{})["norm"] = "1"
		},
		func(o map[string]interface{}) { ds := o["dimensions"].([]interface{}); ds[0], ds[1] = ds[1], ds[0] },
		func(o map[string]interface{}) {
			o["dimensions"].([]interface{})[0].(map[string]interface{})["items"].([]interface{})[0].(map[string]interface{})["reverse"] = true
		},
		func(o map[string]interface{}) { o["policy"].(map[string]interface{})["overallNorm"] = "54.23" },
		func(o map[string]interface{}) { o["policy"].(map[string]interface{})["highest"] = 4 },
		func(o map[string]interface{}) { o["policy"].(map[string]interface{})["overlap"] = false },
		func(o map[string]interface{}) {
			o["policy"].(map[string]interface{})["gradeBounds"].([]interface{})[0] = 89
		},
	} {
		o := managementTraitsDecodeObject(t, mc.JSON)
		change(o)
		data := managementTraitsDecodeJSON(t, o)
		got, err := DecodeManagementTraitsManifest(data, len(data))
		managementTraitsDecodeError(t, got, ManagementTraitsManifest{}, err)
	}
	pc, _ := CanonicalManagementTraitsMapping(m, mp)
	for _, change := range []func(map[string]interface{}){
		func(o map[string]interface{}) { o["schema"] = "another-domain" },
		func(o map[string]interface{}) { o["manifestSha"] = strings.Repeat("a", 64) },
		func(o map[string]interface{}) { o["questionnaire"] = "leader" },
		func(o map[string]interface{}) { qs := o["questions"].([]interface{}); qs[1] = qs[0] },
		func(o map[string]interface{}) {
			qs := o["questions"].([]interface{})
			qs[1].(map[string]interface{})["sourceQuestionId"] = qs[0].(map[string]interface{})["sourceQuestionId"]
		},
		func(o map[string]interface{}) {
			o["questions"].([]interface{})[0].(map[string]interface{})["content"] = "changed"
		},
	} {
		o := managementTraitsDecodeObject(t, pc.JSON)
		change(o)
		data := managementTraitsDecodeJSON(t, o)
		got, err := DecodeManagementTraitsMapping(data, m, len(data))
		managementTraitsDecodeError(t, got, ManagementTraitsMapping{}, err)
	}
	badManifest := managementTraitsContractClone(t, m)
	badManifest.Questions[0].Reverse = !badManifest.Questions[0].Reverse
	got, err := DecodeManagementTraitsMapping(pc.JSON, badManifest, len(pc.JSON))
	managementTraitsDecodeError(t, got, ManagementTraitsMapping{}, err)
}

func TestManagementTraitsDecodeUnicode(t *testing.T) {
	for _, text := range []string{"中文原文 � 😀 ", "literal \\uD800", "escaped quote \" and slash \\"} {
		m, mp, _ := managementTraitsContractFixture(t, "staff")
		m.Questions[0].Content, mp.Questions[0].Content = text, text
		mc, _ := CanonicalManagementTraitsManifest(m)
		mp.ManifestSHA = mc.SHA256
		pc, _ := CanonicalManagementTraitsMapping(m, mp)
		mb := bytes.ReplaceAll(mc.JSON, []byte("😀"), []byte(`\uD83D\uDE00`))
		mb = bytes.ReplaceAll(mb, []byte("�"), []byte(`\uFFFD`))
		decoded, err := DecodeManagementTraitsManifest(mb, len(mb))
		if err != nil || decoded.Questions[0].Content != text {
			t.Fatal("valid Unicode rejected or normalized", err)
		}
		pb := bytes.ReplaceAll(pc.JSON, []byte("😀"), []byte(`\uD83D\uDE00`))
		pb = bytes.ReplaceAll(pb, []byte("�"), []byte(`\uFFFD`))
		mapped, err := DecodeManagementTraitsMapping(pb, decoded, len(pb))
		if err != nil || mapped.Questions[0].Content != text {
			t.Fatal("Unicode mapping", err)
		}
	}
}

func TestManagementTraitsDecodeStoredValidAndDetached(t *testing.T) {
	for _, audience := range []string{"staff", "leader"} {
		for _, participant := range []string{"candidate", "tester"} {
			for _, raw := range []int{3, 4} {
				t.Run(fmt.Sprintf("%s/%s/%d", audience, participant, raw), func(t *testing.T) {
					b, p, rows, want := managementTraitsDecodeStoredFixture(t, audience, participant, raw)
					for _, profile := range []bool{false, true} {
						if profile {
							exam := p.ExamID
							p.ProfileExamID = &exam
						}
						before := managementTraitsDecodeJSON(t, []interface{}{b, p, rows})
						budget := managementTraitsDecodeBudget(b, p, rows)
						got, err := ValidateManagementTraitsStoredInput(b, p, rows, budget)
						if err != nil || !reflect.DeepEqual(got, want) || !bytes.Equal(before, managementTraitsDecodeJSON(t, []interface{}{b, p, rows})) {
							t.Fatal("stored roundtrip or mutation", err)
						}
						got.Canonical.JSON[0], got.Answers[0].Raw = 'x', 5
						fresh, err := ValidateManagementTraitsStoredInput(b, p, rows, budget)
						if err != nil || !reflect.DeepEqual(fresh, want) {
							t.Fatal("output pollution")
						}
						for _, limit := range []int{0, -1, budget - 1} {
							bad, err := ValidateManagementTraitsStoredInput(b, p, rows, limit)
							managementTraitsDecodeError(t, bad, ManagementTraitsValidatedInput{}, err)
						}
					}
					for i := range rows {
						var options []ManagementTraitsMappedOption
						if err := json.Unmarshal([]byte(rows[i].OptionsSnapshot), &options); err != nil {
							t.Fatal(err)
						}
						options[0], options[4] = options[4], options[0]
						rows[i].OptionsSnapshot = string(managementTraitsDecodeJSON(t, options))
					}
					rows[0], rows[139] = rows[139], rows[0]
					got, err := ValidateManagementTraitsStoredInput(b, p, rows, managementTraitsDecodeBudget(b, p, rows))
					if err != nil || !reflect.DeepEqual(got, want) {
						t.Fatal("row/option reorder changed result", err)
					}
				})
			}
		}
	}
}

func TestManagementTraitsDecodeStoredRejects(t *testing.T) {
	type change func(*model.ManagementTraitsDefinitionBundle, *model.ManagementTraitsPaperSnapshot, *[]model.ManagementTraitsPaperQuestionSnapshot)
	cases := map[string]change{
		"bundle_id": func(b *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			b.ID = ""
			p.BundleID = ""
		},
		"bundle_questionnaire": func(b *model.ManagementTraitsDefinitionBundle, _ *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			b.Questionnaire = "leader"
		},
		"bundle_sha": func(b *model.ManagementTraitsDefinitionBundle, _ *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			b.ScoringManifestSHA = strings.ToUpper(b.ScoringManifestSHA)
		},
		"bundle_sha_arbitrary": func(b *model.ManagementTraitsDefinitionBundle, _ *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			b.ScoringManifestSHA = strings.Repeat("a", 64)
		},
		"bundle_policy_changed_rehash": func(b *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			o := managementTraitsDecodeObject(t, []byte(b.ScoringManifest))
			o["policy"].(map[string]interface{})["highest"] = 4
			c, err := managementTraitsCanonicalBytes(o)
			if err != nil {
				t.Fatal(err)
			}
			b.ScoringManifest = string(c.JSON)
			b.ScoringManifestSHA = c.SHA256
			p.ScoringManifestSHA = c.SHA256
		},
		"paper_bundle": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			p.BundleID = "other"
		},
		"profile_exam": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			x := "other"
			p.ProfileExamID = &x
		},
		"profile_opaque": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			x := " exam "
			p.ExamID = x
			p.ProfileExamID = &x
		},
		"paper_manifest_sha": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			p.ScoringManifestSHA = ""
		},
		"paper_mapping_sha": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			p.MappingSHA = strings.ToUpper(p.MappingSHA)
		},
		"paper_mapping_json": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			p.MappingSnapshot += " {}"
		},
		"paper_type": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			p.ParticipantType = "user"
		},
		"paper_id": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			p.PaperID = ""
		},
		"exam_id": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			p.ExamID = ""
		},
		"participant_id": func(_ *model.ManagementTraitsDefinitionBundle, p *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			p.ParticipantID = ""
		},
	}
	for axis := 0; axis < 4; axis++ {
		cases[fmt.Sprintf("bundle_version_%d", axis)] = func(b *model.ManagementTraitsDefinitionBundle, _ *model.ManagementTraitsPaperSnapshot, _ *[]model.ManagementTraitsPaperQuestionSnapshot) {
			fields := []*string{&b.ProductVersion, &b.QuestionVersion, &b.ScoringVersion, &b.NormVersion}
			*fields[axis] = "another-version"
		}
	}
	rowCases := map[string]func(*[]model.ManagementTraitsPaperQuestionSnapshot){
		"missing":       func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { *r = (*r)[:139] },
		"extra":         func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { *r = append(*r, (*r)[0]) },
		"id_empty":      func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].ID = "" },
		"id_duplicate":  func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[1].ID = (*r)[0].ID },
		"id_control":    func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].ID = "secret\x00token" },
		"foreign_paper": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].PaperID = "foreign" },
		"v_zero":        func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].Number = 0 },
		"v_out":         func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].Number = 141 },
		"v_duplicate":   func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[1].Number = (*r)[0].Number },
		"pqid_empty":    func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].PaperQuestionID = "" },
		"pqid_duplicate": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			(*r)[1].PaperQuestionID = (*r)[0].PaperQuestionID
		},
		"order_zero":      func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].DisplayOrder = 0 },
		"order_out":       func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].DisplayOrder = 141 },
		"order_duplicate": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[1].DisplayOrder = (*r)[0].DisplayOrder },
		"cross_source": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			(*r)[0].SourceQuestionID = (*r)[1].SourceQuestionID
		},
		"dimension": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].DimensionKey = "wrong" },
		"reverse":   func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].Reverse = !(*r)[0].Reverse },
		"content_whitespace": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			(*r)[0].Content = strings.TrimSpace((*r)[0].Content)
		},
		"selected_partial": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].SelectedOptionID = nil },
		"raw_partial":      func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].RawAnswer = nil },
		"final_partial":    func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].FinalScore = nil },
		"raw_zero":         func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { *(*r)[0].RawAnswer = 0 },
		"raw_six":          func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { *(*r)[0].RawAnswer = 6 },
		"raw_mismatch": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			*(*r)[0].RawAnswer = 4
			*(*r)[0].FinalScore = 4
		},
		"selection_cross_v": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			*(*r)[0].SelectedOptionID = *(*r)[1].SelectedOptionID
		},
		"selection_blank": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { *(*r)[0].SelectedOptionID = "" },
		"final":           func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { *(*r)[0].FinalScore = 4 },
		"sha":             func(r *[]model.ManagementTraitsPaperQuestionSnapshot) { (*r)[0].ScoringSnapshotSHA = "" },
		"sha_case": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			(*r)[0].ScoringSnapshotSHA = strings.ToUpper((*r)[0].ScoringSnapshotSHA)
		},
		"sha_cross_v": func(r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			(*r)[0].ScoringSnapshotSHA = (*r)[1].ScoringSnapshotSHA
		},
	}
	for name, f := range rowCases {
		cases["row_"+name] = func(_ *model.ManagementTraitsDefinitionBundle, _ *model.ManagementTraitsPaperSnapshot, r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			f(r)
		}
	}
	for _, jsonCase := range []string{"null", "[]", "{}", `[null]`, `[[]]`, `[1.0]`, "bad", "missing", "duplicate", "unknown", "case", "null_field", "changed", "duplicate_raw", "escaped_dup", "float", "exponent", "overflow", "bom", "utf8", "surrogate", "trailing"} {
		cases["options_"+jsonCase] = func(_ *model.ManagementTraitsDefinitionBundle, _ *model.ManagementTraitsPaperSnapshot, r *[]model.ManagementTraitsPaperQuestionSnapshot) {
			s := (*r)[0].OptionsSnapshot
			switch jsonCase {
			case "missing":
				s = strings.Replace(s, `"sourceOptionId":"source-o-1-1",`, "", 1)
			case "duplicate":
				s = strings.Replace(s, `"raw":1`, `"raw":1,"raw":1`, 1)
			case "unknown":
				s = strings.Replace(s, `"raw":1`, `"raw":1,"extra":"secret-token"`, 1)
			case "case":
				s = strings.Replace(s, `"raw":1`, `"Raw":1`, 1)
			case "null_field":
				s = strings.Replace(s, `"sourceOptionId":"source-o-1-1"`, `"sourceOptionId":null`, 1)
			case "changed":
				s = strings.Replace(s, "source-o-1-1", "foreign-option", 1)
			case "duplicate_raw":
				s = strings.Replace(s, `"raw":2`, `"raw":1`, 1)
			case "escaped_dup":
				s = strings.Replace(s, `"raw":1`, `"raw":1,"r\u0061w":1`, 1)
			case "float":
				s = strings.Replace(s, `"raw":1`, `"raw":1.0`, 1)
			case "exponent":
				s = strings.Replace(s, `"raw":1`, `"raw":1e0`, 1)
			case "overflow":
				s = strings.Replace(s, `"raw":1`, `"raw":999999999999999999999999`, 1)
			case "bom":
				s = "\uFEFF" + s
			case "utf8":
				s = string([]byte{255}) + s
			case "surrogate":
				s = strings.Replace(s, "source-o-1-1", `\uDC00`, 1)
			case "trailing":
				s += " []"
			default:
				s = jsonCase
			}
			if s == (*r)[0].OptionsSnapshot {
				t.Fatal("options mutation missed target")
			}
			(*r)[0].OptionsSnapshot = s
		}
	}
	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			b, p, rows, _ := managementTraitsDecodeStoredFixture(t, "staff", "candidate", 3)
			f(&b, &p, &rows)
			got, err := ValidateManagementTraitsStoredInput(b, p, rows, managementTraitsDecodeBudget(b, p, rows)+1)
			managementTraitsDecodeError(t, got, ManagementTraitsValidatedInput{}, err)
		})
	}
	// A wrong-direction final score must be checked on both forward and reverse rows.
	for _, reverse := range []bool{false, true} {
		b, p, rows, _ := managementTraitsDecodeStoredFixture(t, "staff", "candidate", 4)
		for i := range rows {
			if rows[i].Reverse == reverse {
				*rows[i].FinalScore = 6 - *rows[i].FinalScore
				break
			}
		}
		got, err := ValidateManagementTraitsStoredInput(b, p, rows, managementTraitsDecodeBudget(b, p, rows))
		managementTraitsDecodeError(t, got, ManagementTraitsValidatedInput{}, err)
	}
}

func TestManagementTraitsDecodeMixedRawAndLargeBudget(t *testing.T) {
	b, p, rows, _ := managementTraitsDecodeStoredFixture(t, "staff", "tester", 3)
	m, mp, in := managementTraitsContractFixture(t, "staff")
	in.ParticipantType = "tester"
	for i := 0; i < len(rows); i += 2 {
		*rows[i].RawAnswer = 4
		*rows[i].FinalScore = 4
		if rows[i].Reverse {
			*rows[i].FinalScore = 2
		}
		*rows[i].SelectedOptionID = mp.Questions[i].Options[3].SourceOptionID
		in.Answers[i].Raw, in.Answers[i].SelectedOptionID = 4, *rows[i].SelectedOptionID
	}
	want, err := CanonicalManagementTraitsInput(m, mp, in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ValidateManagementTraitsStoredInput(b, p, rows, int(^uint(0)>>1))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("mixed raw answers or maximum budget changed contract", err)
	}
}

func TestManagementTraitsDecodeIncompleteAndDeferredMetadata(t *testing.T) {
	b, p, rows, _ := managementTraitsDecodeStoredFixture(t, "staff", "candidate", 4)
	rows[119].SelectedOptionID, rows[119].RawAnswer, rows[119].FinalScore = nil, nil, nil
	// These deliberately invalid/nonapproved metadata values are outside S2C.
	b.Status = "unapproved"
	p.Source, p.IdentitySource, p.EvidenceSnapshot, p.EvidenceSHA = "unverified", "unverified", "not json", "not hash"
	p.ParticipantSnapshot, p.FieldContract = "not json", "not json"
	p.StartedAt = time.Time{}
	now := time.Now()
	rows[119].SubmittedAt = &now
	got, err := ValidateManagementTraitsStoredInput(b, p, rows, managementTraitsDecodeBudget(b, p, rows))
	if err != nil || got.Answers[119] != (ManagementTraitsAnswer{Number: 120}) {
		t.Fatal("incomplete/deferred metadata rejected or answer inferred", err)
	}
	s, err := CalculateManagementTraits("staff", got.Answers)
	if err != nil || s.IsComplete || s.OverallScore != nil || s.OverallNorm != nil {
		t.Fatal("incomplete delegated S1 generated formal scores")
	}
	for _, d := range s.Dimensions {
		if d.Score != nil || d.Norm != nil || d.Level != "" {
			t.Fatal("incomplete dimension generated formal scores")
		}
	}
	for _, v := range []interface{}{got, got.Canonical, got.Answers[0]} {
		typ := reflect.TypeOf(v)
		for i := 0; i < typ.NumField(); i++ {
			name := strings.ToLower(typ.Field(i).Name)
			for _, forbidden := range []string{"approved", "authorized", "submitted", "complete", "score", "status", "evidence"} {
				if strings.Contains(name, forbidden) {
					t.Fatal("adapter output claims deferred facts")
				}
			}
		}
	}
}
