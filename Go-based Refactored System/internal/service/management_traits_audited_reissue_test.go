package service

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func managementReissueFixture(t *testing.T, code string, final int) managementTraitsReissueWire {
	t.Helper()
	questionnaire := "staff"
	if code == "00202" {
		questionnaire = "leader"
	}
	m, mp, in := managementTraitsContractFixture(t, questionnaire)
	m.Versions = ManagementTraitsScoringVersions{"mng-traits-v2", "synthetic-historical-" + questionnaire + "-v1", "mng-percent-scoring-v1", "mng-norm-20260928-v1"}
	mc, err := CanonicalManagementTraitsManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	mp.ManifestSHA, in.ManifestSHA = mc.SHA256, mc.SHA256
	pc, err := CanonicalManagementTraitsMapping(m, mp)
	if err != nil {
		t.Fatal(err)
	}
	in.MappingSHA = pc.SHA256
	w := managementTraitsReissueWire{Schema: "mng-audited-reissue-source-v1", Policy: "legacy_verified_snapshot", Synthetic: true, RepoCode: code, EvidenceKind: "historical_export", EvidenceReference: "synthetic-original-export-not-a-real-user", Manifest: m, Mapping: mp, Input: in,
		Owner: managementTraitsReissueOwner{PaperID: in.PaperID, ExamID: in.ExamID, ParticipantType: in.ParticipantType, ParticipantID: in.ParticipantID, Identity: managementTraitsRuntimeIdentity{Name: "SYNTHETIC REISSUE SAMPLE"}, RequiredFields: []string{"name"}, IdentitySource: "historical_completion_snapshot", StartedAt: "2026-09-01T10:00:00+08:00", SubmittedAt: "2026-09-01T10:12:00+08:00"}, Buckets: make([]managementTraitsReissueBucket, 140)}
	for i, q := range m.Questions {
		raw := final
		if q.Reverse {
			raw = 6 - final
		}
		w.Input.Answers[i].Raw = raw
		w.Input.Answers[i].SelectedOptionID = mp.Questions[i].Options[raw-1].SourceOptionID
		bucket := managementTraitsReissueBucket{PaperQuestionID: in.Answers[i].PaperQuestionID, SourceQuestionID: mp.Questions[i].SourceQuestionID, Options: make([]managementTraitsReissueOption, 5)}
		for j, o := range mp.Questions[i].Options {
			right := o.Raw == 5
			if q.Reverse {
				right = o.Raw == 1
			}
			bucket.Options[j] = managementTraitsReissueOption{o.SourceOptionID, o.Raw, right, o.Raw == raw}
		}
		w.Buckets[i] = bucket
	}
	return w
}

func managementReissueSign(t *testing.T, w managementTraitsReissueWire, key ed25519.PrivateKey) ([]byte, []byte) {
	t.Helper()
	raw, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	return raw, ed25519.Sign(key, append([]byte("mng-audited-reissue-v1\x00"), raw...))
}

func TestManagementTraitsAuditedReissueExactAndDetached(t *testing.T) {
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	content, err := LoadManagementTraitsTestContent(managementTestWorkbook(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"00201", "00202"} {
		for _, final := range []int{1, 3, 5} {
			t.Run(code+"-"+string(rune('0'+final)), func(t *testing.T) {
				w := managementReissueFixture(t, code, final)
				raw, sig := managementReissueSign(t, w, key)
				sourceBefore := append([]byte{}, raw...)
				got, err := BuildManagementTraitsAuditedReissue(raw, sig, pub, content, 1<<20)
				if err != nil {
					t.Fatal(err)
				}
				dto, err := got.ReportData()
				if err != nil || dto.Overall.Score != map[int]string{1: "0.00", 3: "50.00", 5: "100.00"}[final] || !dto.TestOnly || len(dto.Dimensions) != 13 || len(dto.Modules) != 4 {
					t.Fatal("score/DTO", err)
				}
				var snap managementTraitsReissueSnapshot
				if json.Unmarshal(got.SnapshotJSON(), &snap) != nil || snap.Policy != "audited_reissue" || snap.OverallNormExact != "705/13" || snap.ReverseCount != 40 || snap.UserTimeSeconds != 720 || snap.SourceSHA != managementTraitsReissueSHA(sourceBefore) || len(snap.Scores) != 13 || len(snap.Modules) != 4 || len(snap.Answers) != 140 {
					t.Fatal("provenance/score evidence")
				}
				for _, d := range snap.Scores {
					if d.Sum != d.Count*final {
						t.Fatal("reverse not applied exactly once")
					}
				}
				for _, m := range snap.Modules {
					if m.Exact != map[int]string{1: "0", 3: "50", 5: "100"}[final] {
						t.Fatal("module exact fraction lost")
					}
				}
				if dto.Dimensions[11].Norm != "53.75" || dto.Dimensions[12].Norm != "50.00" || dto.SubmittedAt != "2026-09-01 10:12" {
					t.Fatal("norm/date")
				}
				jsonBytes := got.SnapshotJSON()
				jsonBytes[0] = 'x'
				raw[0] = 'x'
				sig[0] ^= 1
				dto.Dimensions[0].Advice = "mutated"
				again, err := got.ReportData()
				if err != nil || again.Dimensions[0].Advice == "mutated" || !bytes.Equal(got.SourceBytes(), sourceBefore) || !json.Valid(got.SnapshotJSON()) {
					t.Fatal("mutable snapshot escaped")
				}
			})
		}
	}
}

func TestManagementTraitsAuditedReissueRejectsInsufficientEvidence(t *testing.T) {
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	content, err := LoadManagementTraitsTestContent(managementTestWorkbook(t))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		edit func(*managementTraitsReissueWire)
	}{
		{"current-not-historical", func(w *managementTraitsReissueWire) { w.Policy = "new_creation" }},
		{"no-evidence", func(w *managementTraitsReissueWire) { w.EvidenceReference = "" }},
		{"wrong-code", func(w *managementTraitsReissueWire) { w.RepoCode = "00401" }},
		{"cross-owner", func(w *managementTraitsReissueWire) { w.Owner.PaperID = "foreign" }},
		{"live-identity", func(w *managementTraitsReissueWire) { w.Owner.IdentitySource = "current_person" }},
		{"identity-missing", func(w *managementTraitsReissueWire) { w.Owner.Identity.Name = "" }},
		{"time-missing", func(w *managementTraitsReissueWire) { w.Owner.StartedAt = "" }},
		{"time-backwards", func(w *managementTraitsReissueWire) { w.Owner.SubmittedAt = "2026-08-01 10:00:00" }},
		{"duplicate-v", func(w *managementTraitsReissueWire) { w.Input.Answers[1] = w.Input.Answers[0] }},
		{"unknown-dimension", func(w *managementTraitsReissueWire) { w.Manifest.Questions[0].DimensionKey = "unknown" }},
		{"reverse-drift", func(w *managementTraitsReissueWire) {
			w.Manifest.Questions[0].Reverse = !w.Manifest.Questions[0].Reverse
		}},
		{"raw-zero", func(w *managementTraitsReissueWire) { w.Input.Answers[0].Raw = 0 }},
		{"raw-six", func(w *managementTraitsReissueWire) { w.Input.Answers[0].Raw = 6 }},
		{"unanswered", func(w *managementTraitsReissueWire) {
			w.Input.Answers[0].Answered = false
			w.Input.Answers[0].Raw = 0
			w.Input.Answers[0].SelectedOptionID = ""
		}},
		{"139-answers", func(w *managementTraitsReissueWire) { w.Input.Answers = w.Input.Answers[:139] }},
		{"mapping-missing", func(w *managementTraitsReissueWire) { w.Mapping.Questions = nil }},
		{"chosen-score-drift", func(w *managementTraitsReissueWire) { w.Buckets[0].Options[2].Raw = 4 }},
		{"multi-selected", func(w *managementTraitsReissueWire) { w.Buckets[0].Options[0].Checked = true }},
		{"no-selected", func(w *managementTraitsReissueWire) { w.Buckets[0].Options[2].Checked = false }},
		{"direction-bucket-drift", func(w *managementTraitsReissueWire) {
			w.Buckets[0].Options[4].IsRight = false
			w.Buckets[0].Options[0].IsRight = true
		}},
		{"cross-bucket", func(w *managementTraitsReissueWire) { w.Buckets[0].SourceQuestionID = "foreign" }},
		{"unsupported-scoring", func(w *managementTraitsReissueWire) { w.Manifest.Versions.Scoring = "old-scoring" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := managementReissueFixture(t, "00201", 3)
			c.edit(&w)
			raw, sig := managementReissueSign(t, w, key)
			if _, err := BuildManagementTraitsAuditedReissue(raw, sig, pub, content, 1<<20); err == nil {
				t.Fatal("insufficient signed evidence accepted")
			}
		})
	}
	w := managementReissueFixture(t, "00201", 3)
	raw, sig := managementReissueSign(t, w, key)
	for _, c := range []struct {
		name          string
		raw, sig, pub []byte
		budget        int
	}{
		{"unsigned", raw, nil, pub, 1 << 20}, {"no-audit-root", raw, sig, nil, 1 << 20}, {"budget", raw, sig, pub, len(raw) - 1}, {"zero-budget", raw, sig, pub, 0},
		{"foreign-audit-root", raw, sig, make([]byte, ed25519.PublicKeySize), 1 << 20},
		{"tampered", append(append([]byte{}, raw...), byte(' ')), sig, pub, 1 << 20},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := BuildManagementTraitsAuditedReissue(c.raw, c.sig, c.pub, content, c.budget); err == nil {
				t.Fatal("invalid envelope accepted")
			}
		})
	}
	for _, replacement := range []string{`"synthetic":""`, `"synthetic":null`, `"synthetic":1`, `"synthetic":true,"synthetic":true`, `"Synthetic":true`} {
		bad := []byte(strings.Replace(string(raw), `"synthetic":true`, replacement, 1))
		signature := ed25519.Sign(key, append([]byte("mng-audited-reissue-v1\x00"), bad...))
		if _, err := BuildManagementTraitsAuditedReissue(bad, signature, pub, content, 1<<20); err == nil {
			t.Fatal("strict decode accepted", replacement)
		}
	}
	w.Owner.StartedAt = "2026-09-01 10:00:00"
	w.Owner.SubmittedAt = "2026-09-01 10:12:00"
	raw, sig = managementReissueSign(t, w, key)
	if _, err := BuildManagementTraitsAuditedReissue(raw, sig, pub, content, 1<<20); err != nil {
		t.Fatal("local time format rejected", err)
	}
}

func TestManagementTraitsAuditedReissueKeepsNewCreationGuard(t *testing.T) {
	content, err := LoadManagementTraitsTestContent(managementTestWorkbook(t))
	if err != nil {
		t.Fatal(err)
	}
	f, r, _, _, _ := managementRuntimeLoadFixture(t, false)
	f.Paper.Source = "legacy_verified_snapshot"
	if _, err := BuildManagementTraitsTestReport(f.Bundle, f.Paper, f.Questions, r.Run, r.Dimensions, r.Modules, r.Receipt, content, 1<<20); err == nil {
		t.Fatal("old new_creation guard was widened")
	}
	if _, err := (ManagementTraitsAuditedReissue{}).ReportData(); err == nil {
		t.Fatal("zero value could produce report")
	}
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	w := managementReissueFixture(t, "00202", 3)
	w.Owner.ParticipantType = "tester"
	w.Input.ParticipantType = "tester"
	raw, sig := managementReissueSign(t, w, key)
	if _, err := BuildManagementTraitsAuditedReissue(raw, sig, pub, content, 1<<20); err != nil {
		t.Fatal("audited tester source rejected", err)
	}
}

func TestManagementTraitsAuditedReissueSyntheticArtifact(t *testing.T) {
	w := managementReissueFixture(t, "00201", 3)
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	raw, sig := managementReissueSign(t, w, key)
	content, err := LoadManagementTraitsTestContent(managementTestWorkbook(t))
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := BuildManagementTraitsAuditedReissue(raw, sig, pub, content, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if dir := os.Getenv("MNG_REISSUE_ARTIFACT_DIR"); dir != "" {
		if !filepath.IsAbs(dir) {
			t.Fatal("absolute artifact directory required")
		}
		for name, b := range map[string][]byte{"source.json": raw, "audit-signature.bin": sig, "synthetic-auditor-public.bin": pub, "snapshot.json": snapshot.SnapshotJSON()} {
			f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			_, e := f.Write(b)
			closeErr := f.Close()
			if e != nil || closeErr != nil {
				t.Fatal("artifact write failed")
			}
		}
	}
}
