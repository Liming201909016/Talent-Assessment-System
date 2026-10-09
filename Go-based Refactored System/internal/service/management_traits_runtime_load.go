package service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type managementTraitsRuntimeFieldContract struct {
	Schema         string    `json:"schema"`
	RequiredFields []string  `json:"requiredFields"`
	TimePolicy     string    `json:"timePolicy"`
	Source         string    `json:"source"`
	RepoCode       string    `json:"repoCode"`
	CapturedAt     time.Time `json:"capturedAt"`
	ManifestSHA    string    `json:"manifestSha"`
	MappingSHA     string    `json:"mappingSha"`
}

type managementTraitsRuntimeEvidence struct {
	Schema      string    `json:"schema"`
	Source      string    `json:"source"`
	RepoCode    string    `json:"repoCode"`
	CapturedAt  time.Time `json:"capturedAt"`
	ManifestSHA string    `json:"manifestSha"`
	MappingSHA  string    `json:"mappingSha"`
}

type managementTraitsRuntimeIdentity struct {
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Telephone   string  `json:"telephone"`
	Affiliation string  `json:"affiliation"`
	Post        string  `json:"post"`
	Age         *int    `json:"age,omitempty"`
	Degree      *string `json:"degree,omitempty"`
	Major       *string `json:"major,omitempty"`
	StuFlag     *int    `json:"stuFlag,omitempty"`
}

// Scoring facts only; not a formal-report permission or a PDF DTO.
type ManagementTraitsValidatedRuntimeRun struct {
	RunID           string
	PaperID         string
	ExamID          string
	ParticipantType string
	ParticipantID   string
	Questionnaire   string
	Versions        ManagementTraitsScoringVersions
	SubmittedAt     time.Time
	UserTimeSeconds int
	Result          ManagementTraitsResult
}

func managementTraitsRuntimeFields(fields []string, identity managementTraitsRuntimeIdentity) bool {
	if len(fields) == 0 || len(fields) > 9 {
		return false
	}
	values := map[string]string{"name": identity.Name, "gender": identity.Gender, "telephone": identity.Telephone, "affiliation": identity.Affiliation, "post": identity.Post}
	if identity.Age != nil && *identity.Age > 0 && int64(*identity.Age) <= 2147483647 {
		values["age"] = "present"
	}
	if identity.StuFlag != nil && (*identity.StuFlag == 0 || *identity.StuFlag == 1) {
		values["stuFlag"] = "present"
	}
	if identity.Degree != nil {
		values["degree"] = *identity.Degree
	}
	if identity.Major != nil {
		values["major"] = *identity.Major
	}
	seen := make(map[string]bool, 5)
	for _, f := range fields {
		value, exists := values[f]
		if !exists || seen[f] || strings.TrimSpace(value) == "" {
			return false
		}
		seen[f] = true
	}
	return true
}

// Use ordinary JSON for the time.Time leaf, but first require exact keys and
// reject duplicate, unknown, null and case-folded names with the S2C reader.
// A time.Time leaf is represented as a string in the strict wire structs below.
func managementTraitsRuntimeMetadata(bundle model.ManagementTraitsDefinitionBundle, snapshot model.ManagementTraitsPaperSnapshot, profile model.ManagementTraitsExamProfile, paper model.Paper, owner managementTraitsRuntimeOwner, budget int) error {
	// Today's draft governs only new papers; reconstruct the frozen contract.
	captured := snapshot.SourceCapturedAt
	profile = model.ManagementTraitsExamProfile{ExamID: snapshot.ExamID, BundleID: snapshot.BundleID, MappingSHA: snapshot.MappingSHA, MappingSnapshot: snapshot.MappingSnapshot, FieldContract: snapshot.FieldContract, TotalTimeMinutes: 25, FrozenAt: &captured, CreatedAt: captured}
	v := ManagementTraitsScoringVersions{Product: bundle.ProductVersion, Question: bundle.QuestionVersion, Scoring: bundle.ScoringVersion, Norm: bundle.NormVersion}
	if budget <= 0 || snapshot.Source != "new_creation" || snapshot.IdentitySource != "submitted_snapshot" || snapshot.ProfileExamID == nil || *snapshot.ProfileExamID != snapshot.ExamID || paper.ID != snapshot.PaperID || paper.ExamID != snapshot.ExamID || profile.ExamID != snapshot.ExamID || profile.BundleID != bundle.ID || snapshot.BundleID != bundle.ID || !managementTraitsRuntimeSourceTrusted(bundle.Status) || validateManagementTraitsRuntimeVersions(bundle.Questionnaire, v) != nil || profile.TotalTimeMinutes != 25 || paper.TotalTime != 25 || profile.FrozenAt == nil || profile.FrozenAt.IsZero() || !profile.FrozenAt.Equal(snapshot.SourceCapturedAt) || !profile.CreatedAt.Equal(*profile.FrozenAt) || profile.FrozenAt.After(snapshot.StartedAt) || paper.CreateTime == nil || !paper.CreateTime.Equal(snapshot.StartedAt) || !snapshot.CreatedAt.Equal(snapshot.StartedAt) || paper.LimitTime == nil || snapshot.LimitTime == nil || !paper.LimitTime.Equal(*snapshot.LimitTime) || !snapshot.LimitTime.Equal(snapshot.StartedAt.Add(25*time.Minute)) || profile.MappingSHA != snapshot.MappingSHA || profile.MappingSnapshot != snapshot.MappingSnapshot || profile.FieldContract != snapshot.FieldContract {
		return ErrManagementTraitsRuntimeInvalid
	}
	var fields struct {
		Schema         string   `json:"schema"`
		RequiredFields []string `json:"requiredFields"`
		TimePolicy     string   `json:"timePolicy"`
		Source         string   `json:"source"`
		RepoCode       string   `json:"repoCode"`
		CapturedAt     string   `json:"capturedAt"`
		ManifestSHA    string   `json:"manifestSha"`
		MappingSHA     string   `json:"mappingSha"`
	}
	var evidence struct {
		Schema      string `json:"schema"`
		Source      string `json:"source"`
		RepoCode    string `json:"repoCode"`
		CapturedAt  string `json:"capturedAt"`
		ManifestSHA string `json:"manifestSha"`
		MappingSHA  string `json:"mappingSha"`
	}
	var identity managementTraitsRuntimeIdentity
	if managementTraitsDecodeStrict([]byte(snapshot.FieldContract), budget, &fields) != nil || managementTraitsDecodeStrict([]byte(snapshot.EvidenceSnapshot), budget, &evidence) != nil || managementTraitsDecodeStrict([]byte(snapshot.ParticipantSnapshot), budget, &identity) != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	q, expectedVersions, err := managementTraitsRuntimeVersions(fields.RepoCode)
	if err != nil || q != bundle.Questionnaire || v != expectedVersions || fields.Schema != "mng-candidate-fields-v1" || fields.TimePolicy != "candidate-only-personal-25-minutes-v1" || fields.Source != "current-source-not-client-leader-full-not-historical" || fields.CapturedAt != snapshot.SourceCapturedAt.Format(time.RFC3339Nano) || fields.ManifestSHA != bundle.ScoringManifestSHA || fields.MappingSHA != snapshot.MappingSHA || !managementTraitsRuntimeFields(fields.RequiredFields, identity) {
		return ErrManagementTraitsRuntimeInvalid
	}
	expected := managementTraitsRuntimeEvidence{Schema: "mng-current-source-evidence-v1", Source: fields.Source, RepoCode: fields.RepoCode, CapturedAt: snapshot.SourceCapturedAt, ManifestSHA: bundle.ScoringManifestSHA, MappingSHA: snapshot.MappingSHA}
	c, err := managementTraitsCanonicalBytes(expected)
	if err != nil || evidence.Schema != expected.Schema || evidence.Source != expected.Source || evidence.RepoCode != expected.RepoCode || evidence.CapturedAt != fields.CapturedAt || evidence.ManifestSHA != expected.ManifestSHA || evidence.MappingSHA != expected.MappingSHA || c.SHA256 != snapshot.EvidenceSHA || string(c.JSON) != snapshot.EvidenceSnapshot {
		return ErrManagementTraitsRuntimeInvalid
	}
	if !managementTraitsIdentityLegal(managementTraitsRuntimeFieldWire{RequiredFields: fields.RequiredFields}, identity) {
		return ErrManagementTraitsRuntimeInvalid
	}
	return nil
}

// No INSERT/UPDATE/DELETE, no legacy fallback, no current-pointer selection and
// no formal-report API. The paper lock serializes with future save/submit writers.
func (s *ManagementTraitsRuntimeService) LoadValidatedRun(ctx context.Context, runID string) (ManagementTraitsValidatedRuntimeRun, error) {
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return ManagementTraitsValidatedRuntimeRun{}, err
	}
	if s.db == nil || s.maxJSONBytes <= 0 || !managementTraitsOpaqueID(runID) {
		return ManagementTraitsValidatedRuntimeRun{}, ErrManagementTraitsRuntimeInvalid
	}
	var result ManagementTraitsValidatedRuntimeRun
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var runs []model.ManagementTraitsResultRun
		if tx.Where("id = ?", runID).Limit(2).Find(&runs).Error != nil || len(runs) != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		run := runs[0]
		paper, snapshot, bundle, profile, owner, questions, err := s.loadRuntimePaper(ctx, tx, run.PaperID)
		if err != nil {
			return err
		}
		var dims []model.ManagementTraitsResultDimension
		var modules []model.ManagementTraitsResultModule
		var receipts []model.ManagementTraitsRuntimeReceipt
		if tx.Where("run_id = ?", runID).Limit(14).Find(&dims).Error != nil || tx.Where("run_id = ?", runID).Limit(5).Find(&modules).Error != nil || tx.Where("run_id = ?", runID).Limit(2).Find(&receipts).Error != nil || len(receipts) != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		if (run.Status != "completed" && run.Status != "incomplete") || paper.State != 2 || run.SubmittedAt == nil || run.UserTimeSeconds == nil || paper.UserTime != *run.UserTimeSeconds || owner.EndTime == nil || !owner.EndTime.Equal(*run.SubmittedAt) || managementTraitsRuntimeMetadata(bundle, snapshot, profile, paper, owner, s.maxJSONBytes) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		exact, err := validateManagementTraitsRuntimeResult(bundle, snapshot, questions, managementTraitsRuntimeRecords{Run: run, Dimensions: dims, Modules: modules, Receipt: receipts[0]}, s.maxJSONBytes)
		if err != nil {
			return err
		}
		if validateManagementTraitsRuntimeBuckets(tx, paper.ID, questions, s.maxJSONBytes) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		result = ManagementTraitsValidatedRuntimeRun{RunID: run.ID, PaperID: run.PaperID, ExamID: run.ExamID,
			ParticipantType: run.ParticipantType, ParticipantID: run.ParticipantID, Questionnaire: run.Questionnaire,
			Versions: ManagementTraitsScoringVersions{Product: run.ProductVersion, Question: run.QuestionVersion, Scoring: run.ScoringVersion, Norm: run.NormVersion}, SubmittedAt: *run.SubmittedAt, UserTimeSeconds: *run.UserTimeSeconds, Result: exact}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return ManagementTraitsValidatedRuntimeRun{}, ErrManagementTraitsRuntimeInvalid
	}
	return result, nil
}

func (s *ManagementTraitsRuntimeService) loadRuntimePaper(ctx context.Context, tx *gorm.DB, paperID string) (model.Paper, model.ManagementTraitsPaperSnapshot, model.ManagementTraitsDefinitionBundle, model.ManagementTraitsExamProfile, managementTraitsRuntimeOwner, []model.ManagementTraitsPaperQuestionSnapshot, error) {
	var paper model.Paper
	var snapshot model.ManagementTraitsPaperSnapshot
	var bundle model.ManagementTraitsDefinitionBundle
	var profile model.ManagementTraitsExamProfile
	fail := func() (model.Paper, model.ManagementTraitsPaperSnapshot, model.ManagementTraitsDefinitionBundle, model.ManagementTraitsExamProfile, managementTraitsRuntimeOwner, []model.ManagementTraitsPaperQuestionSnapshot, error) {
		return model.Paper{}, model.ManagementTraitsPaperSnapshot{}, model.ManagementTraitsDefinitionBundle{}, model.ManagementTraitsExamProfile{}, managementTraitsRuntimeOwner{}, nil, ErrManagementTraitsRuntimeInvalid
	}
	if !managementTraitsOpaqueID(paperID) {
		return fail()
	}
	managementRaceMark(ctx, "paper_lock_wait")
	if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", paperID).Take(&paper).Error != nil {
		return fail()
	}
	managementRaceMark(ctx, "paper_lock_acquired")
	if tx.Where("paper_id = ?", paperID).Take(&snapshot).Error != nil || tx.Where("id = ?", snapshot.BundleID).Take(&bundle).Error != nil || tx.Where("exam_id = ?", snapshot.ExamID).Take(&profile).Error != nil {
		return fail()
	}
	owners, err := loadManagementTraitsRuntimeOwners(ctx, tx, paperID)
	if err != nil || validateManagementTraitsRuntimeOwner(owners, snapshot.ParticipantType, snapshot.ParticipantID, paper.ExamID, paperID) != nil {
		return fail()
	}
	if caps := managementTraitsSchemaCapabilities(tx); caps != nil {
		for _, owner := range owners {
			if !managementTraitsSchemaIDFits(caps.Columns, "el_"+owner.Kind+".id", owner.ID) || !managementTraitsSchemaIDFits(caps.Columns, "el_exam.id", owner.ExamID) || !managementTraitsSchemaIDFits(caps.Columns, "el_paper.id", owner.PaperID) {
				return fail()
			}
		}
	}
	questions := make([]model.ManagementTraitsPaperQuestionSnapshot, 0, 140)
	legacy := make([]model.PaperQu, 0, 140)
	if tx.Where("paper_id = ?", paperID).Limit(141).Find(&questions).Error != nil || tx.Where("paper_id = ?", paperID).Limit(141).Find(&legacy).Error != nil || len(questions) != 140 || len(legacy) != 140 {
		return fail()
	}
	pq := make(map[string]model.PaperQu, 140)
	for _, q := range legacy {
		if _, duplicate := pq[q.ID]; duplicate || q.PaperID != paperID || q.QuType != 1 {
			return fail()
		}
		pq[q.ID] = q
	}
	for _, q := range questions {
		p, exists := pq[q.PaperQuestionID]
		if !exists || p.QuID != q.SourceQuestionID || p.Sort != q.DisplayOrder || (q.RawAnswer == nil && (p.Answered != 0 || p.ActualScore != 0)) || (q.RawAnswer != nil && (p.Answered != 1 || p.ActualScore != *q.RawAnswer)) {
			return fail()
		}
	}
	if managementTraitsRuntimeMetadata(bundle, snapshot, profile, paper, owners[0], s.maxJSONBytes) != nil || func() bool {
		_, err := ValidateManagementTraitsStoredInput(bundle, snapshot, questions, s.maxJSONBytes)
		return err != nil
	}() {
		return fail()
	}
	return paper, snapshot, bundle, profile, owners[0], questions, nil
}
