package service

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Offline audit contract, independent of el_mng models/new_creation. An auditor
// must attest the ORIGINAL mapping, text, options, owner and completion times;
// signing a current-source reconstruction does not make it historical evidence.
type managementTraitsReissueWire struct {
	Schema            string                          `json:"schema"`
	Policy            string                          `json:"policy"`
	Synthetic         bool                            `json:"synthetic"`
	RepoCode          string                          `json:"repoCode"`
	EvidenceKind      string                          `json:"evidenceKind"`
	EvidenceReference string                          `json:"evidenceReference"`
	Manifest          ManagementTraitsManifest        `json:"manifest"`
	Mapping           ManagementTraitsMapping         `json:"mapping"`
	Input             ManagementTraitsSnapshotInput   `json:"input"`
	Owner             managementTraitsReissueOwner    `json:"owner"`
	Buckets           []managementTraitsReissueBucket `json:"buckets"`
}

type managementTraitsReissueOwner struct {
	PaperID         string                          `json:"paperId"`
	ExamID          string                          `json:"examId"`
	ParticipantType string                          `json:"participantType"`
	ParticipantID   string                          `json:"participantId"`
	Identity        managementTraitsRuntimeIdentity `json:"identity"`
	RequiredFields  []string                        `json:"requiredFields"`
	IdentitySource  string                          `json:"identitySource"`
	StartedAt       string                          `json:"startedAt"`
	SubmittedAt     string                          `json:"submittedAt"`
}

type managementTraitsReissueOption struct {
	SourceOptionID string `json:"sourceOptionId"`
	Raw            int    `json:"raw"`
	IsRight        bool   `json:"isRight"`
	Checked        bool   `json:"checked"`
}

type managementTraitsReissueBucket struct {
	PaperQuestionID  string                          `json:"paperQuestionId"`
	SourceQuestionID string                          `json:"sourceQuestionId"`
	Options          []managementTraitsReissueOption `json:"options"`
}

type managementTraitsReissueScore struct {
	Key   string `json:"key"`
	Sum   int    `json:"sum"`
	Count int    `json:"count"`
	Exact string `json:"exact"`
	Norm  string `json:"norm"`
	Level string `json:"level"`
}

type managementTraitsReissueAnswer struct {
	Number  int  `json:"number"`
	Raw     int  `json:"raw"`
	Reverse bool `json:"reverse"`
	Final   int  `json:"final"`
}

type managementTraitsReissueModule struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
	Exact string `json:"exact"`
}

type managementTraitsReissueSnapshot struct {
	Schema            string                          `json:"schema"`
	Policy            string                          `json:"policy"`
	Synthetic         bool                            `json:"synthetic"`
	SourceSHA         string                          `json:"sourceSha"`
	SignatureSHA      string                          `json:"signatureSha"`
	AuditorKeySHA     string                          `json:"auditorKeySha"`
	EvidenceKind      string                          `json:"evidenceKind"`
	EvidenceReference string                          `json:"evidenceReference"`
	ManifestSHA       string                          `json:"manifestSha"`
	MappingSHA        string                          `json:"mappingSha"`
	StartedAt         string                          `json:"startedAt"`
	SubmittedAt       string                          `json:"submittedAt"`
	UserTimeSeconds   int                             `json:"userTimeSeconds"`
	ReverseCount      int                             `json:"reverseCount"`
	OverallExact      string                          `json:"overallExact"`
	OverallNormExact  string                          `json:"overallNormExact"`
	Scores            []managementTraitsReissueScore  `json:"scores"`
	Modules           []managementTraitsReissueModule `json:"modules"`
	Answers           []managementTraitsReissueAnswer `json:"answers"`
	Report            ManagementTraitsTestReportData  `json:"report"`
}

// Opaque detached bytes; no exported mutable state or DB identity is invented.
type ManagementTraitsAuditedReissue struct{ source, snapshot []byte }

func (r ManagementTraitsAuditedReissue) SourceBytes() []byte  { return append([]byte{}, r.source...) }
func (r ManagementTraitsAuditedReissue) SnapshotJSON() []byte { return append([]byte{}, r.snapshot...) }
func (r ManagementTraitsAuditedReissue) ReportData() (ManagementTraitsTestReportData, error) {
	var s managementTraitsReissueSnapshot
	if len(r.snapshot) == 0 || json.Unmarshal(r.snapshot, &s) != nil {
		return ManagementTraitsTestReportData{}, errors.New("management traits audited reissue rejected")
	}
	return s.Report, nil
}

func managementTraitsReissueSHA(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

func managementTraitsReissueTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05"} {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil && !t.IsZero() {
			return t, nil
		}
	}
	return time.Time{}, errors.New("historical time missing or invalid")
}

// trustedAuditKey is an application-owned offline trust root, NEVER a key taken
// from a request/evidence payload. No signer, HTTP route, activation, persistence
// or formal permission is supplied here. Signature verifies attestation bytes,
// not archival truth: the auditor must reject insufficient original sources.
func BuildManagementTraitsAuditedReissue(raw, signature []byte, trustedAuditKey ed25519.PublicKey, content ManagementTraitsTestContent, budget int) (ManagementTraitsAuditedReissue, error) {
	fail := func() (ManagementTraitsAuditedReissue, error) {
		return ManagementTraitsAuditedReissue{}, errors.New("management traits audited reissue rejected")
	}
	if budget <= 0 || budget > 20<<20 || len(raw) > budget || len(trustedAuditKey) != ed25519.PublicKeySize || len(signature) != ed25519.SignatureSize || content.RuleCount() != 205 {
		return fail()
	}
	// Detach BEFORE signature verification/decoding; caller buffers cannot be
	// retained in the immutable report. Caller must not mutate concurrently.
	raw = append([]byte{}, raw...)
	signature = append([]byte{}, signature...)
	trustedAuditKey = append(ed25519.PublicKey{}, trustedAuditKey...)
	if !ed25519.Verify(trustedAuditKey, append([]byte("mng-audited-reissue-v1\x00"), raw...), signature) {
		return fail()
	}
	var w managementTraitsReissueWire
	if managementTraitsDecodeStrict(raw, budget, &w) != nil || w.Schema != "mng-audited-reissue-source-v1" || w.Policy != "legacy_verified_snapshot" || w.EvidenceKind != "historical_export" || strings.TrimSpace(w.EvidenceReference) == "" || len(w.EvidenceReference) > 255 {
		return fail()
	}
	questionnaire := ""
	switch w.RepoCode {
	case "00201":
		questionnaire = "staff"
	case "00202":
		questionnaire = "leader"
	default:
		return fail()
	}
	v := w.Manifest.Versions
	if w.Manifest.Questionnaire != questionnaire || v.Product != "mng-traits-v2" || v.Scoring != "mng-percent-scoring-v1" || v.Norm != "mng-norm-20260928-v1" {
		return fail()
	}
	input, err := CanonicalManagementTraitsInput(w.Manifest, w.Mapping, w.Input)
	if err != nil {
		return fail()
	}
	o := w.Owner
	if o.PaperID != w.Input.PaperID || o.ExamID != w.Input.ExamID || o.ParticipantType != w.Input.ParticipantType || o.ParticipantID != w.Input.ParticipantID || o.IdentitySource != "historical_completion_snapshot" || !managementTraitsIdentityLegal(managementTraitsRuntimeFieldWire{RequiredFields: o.RequiredFields}, o.Identity) {
		return fail()
	}
	started, e1 := managementTraitsReissueTime(o.StartedAt)
	submitted, e2 := managementTraitsReissueTime(o.SubmittedAt)
	if e1 != nil || e2 != nil || submitted.Before(started) || submitted.Sub(started) > time.Duration(2147483647)*time.Second {
		return fail()
	}
	if !managementTraitsReissueBuckets(w) {
		return fail()
	}
	exact, err := CalculateManagementTraits(questionnaire, input.Answers)
	if err != nil || !exact.IsComplete {
		return fail()
	}
	identity := o.Identity
	dto, err := managementTraitsReportFromExact(exact, content, ManagementTraitsTestReportData{Schema: "mng-test-report-data-v1", TestOnly: true, RunID: "audited-" + managementTraitsReissueSHA(raw)[:32], PaperID: w.Input.PaperID, ExamID: w.Input.ExamID, Questionnaire: questionnaire, Versions: v, InputSHA: input.Canonical.SHA256, ContentSourceSHA: content.SourceSHA(), Participant: ManagementTraitsTestReportParticipant{Name: identity.Name, Gender: identity.Gender, Telephone: identity.Telephone, Affiliation: identity.Affiliation, Post: identity.Post, Age: identity.Age, Degree: identity.Degree, Major: identity.Major, StuFlag: identity.StuFlag}, SubmittedAt: submitted.In(started.Location()).Format("2006-01-02 15:04")})
	if err != nil {
		return fail()
	}
	s := managementTraitsReissueSnapshot{Schema: "mng-audited-reissue-snapshot-v1", Policy: "audited_reissue", Synthetic: w.Synthetic, SourceSHA: managementTraitsReissueSHA(raw), SignatureSHA: managementTraitsReissueSHA(signature), AuditorKeySHA: managementTraitsReissueSHA(trustedAuditKey), EvidenceKind: w.EvidenceKind, EvidenceReference: w.EvidenceReference, ManifestSHA: w.Input.ManifestSHA, MappingSHA: w.Input.MappingSHA, StartedAt: started.Format(time.RFC3339Nano), SubmittedAt: submitted.Format(time.RFC3339Nano), UserTimeSeconds: int(submitted.Sub(started) / time.Second), OverallExact: exact.OverallScore.RatString(), OverallNormExact: exact.OverallNorm.RatString(), Scores: make([]managementTraitsReissueScore, 0, 13), Answers: make([]managementTraitsReissueAnswer, 0, 140), Report: dto}
	for _, d := range exact.Dimensions {
		s.Scores = append(s.Scores, managementTraitsReissueScore{d.Key, d.ScoreSum, d.QuestionCount, d.Score.RatString(), d.Norm.RatString(), d.Level})
	}
	s.Modules = make([]managementTraitsReissueModule, 0, 4)
	for _, m := range exact.Modules {
		s.Modules = append(s.Modules, managementTraitsReissueModule{m.Key, m.DimensionCount, m.Score.RatString()})
	}
	ordered := managementTraitsOrderedManifest(w.Manifest)
	for _, a := range input.Answers {
		final := a.Raw
		reverse := ordered[a.Number].Reverse
		if reverse {
			final = 6 - final
			s.ReverseCount++
		}
		s.Answers = append(s.Answers, managementTraitsReissueAnswer{a.Number, a.Raw, reverse, final})
	}
	encoded, err := json.Marshal(s)
	if err != nil || len(encoded) > budget {
		return fail()
	}
	return ManagementTraitsAuditedReissue{raw, encoded}, nil
}

func managementTraitsReissueBuckets(w managementTraitsReissueWire) bool {
	if len(w.Buckets) != 140 {
		return false
	}
	answers := make(map[string]ManagementTraitsSnapshotAnswer, 140)
	mapped := make(map[string]ManagementTraitsMappedQuestion, 140)
	for _, a := range w.Input.Answers {
		answers[a.PaperQuestionID] = a
	}
	for _, q := range w.Mapping.Questions {
		mapped[q.SourceQuestionID] = q
	}
	manifest := managementTraitsOrderedManifest(w.Manifest)
	seen := make(map[string]bool, 140)
	for _, b := range w.Buckets {
		a, ok := answers[b.PaperQuestionID]
		q := mapped[b.SourceQuestionID]
		if !ok || seen[b.PaperQuestionID] || b.SourceQuestionID != a.SourceQuestionID || len(b.Options) != 5 || !a.Answered {
			return false
		}
		seen[b.PaperQuestionID] = true
		opts := make(map[string]int, 5)
		for _, o := range q.Options {
			opts[o.SourceOptionID] = o.Raw
		}
		used := make(map[string]bool, 5)
		right, checked := 0, 0
		for _, o := range b.Options {
			if used[o.SourceOptionID] || opts[o.SourceOptionID] != o.Raw || o.Raw < 1 || o.Raw > 5 {
				return false
			}
			used[o.SourceOptionID] = true
			if o.IsRight {
				right++
				want := 5
				if manifest[a.Number].Reverse {
					want = 1
				}
				if o.Raw != want {
					return false
				}
			}
			if o.Checked {
				checked++
				if o.SourceOptionID != a.SelectedOptionID || o.Raw != a.Raw {
					return false
				}
			}
		}
		if right != 1 || checked != 1 {
			return false
		}
	}
	return true
}
