package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type managementTraitsRuntimeFieldWire struct {
	Schema         string   `json:"schema"`
	RequiredFields []string `json:"requiredFields"`
	TimePolicy     string   `json:"timePolicy"`
	Source         string   `json:"source"`
	RepoCode       string   `json:"repoCode"`
	CapturedAt     string   `json:"capturedAt"`
	ManifestSHA    string   `json:"manifestSha"`
	MappingSHA     string   `json:"mappingSha"`
}

func validateManagementTraitsRuntimeProfile(profile model.ManagementTraitsExamProfile, bundle model.ManagementTraitsDefinitionBundle, budget int) error {
	if budget <= 0 || profile.FrozenAt == nil || profile.FrozenAt.IsZero() || !profile.CreatedAt.Equal(*profile.FrozenAt) || profile.TotalTimeMinutes != 25 || profile.BundleID != bundle.ID || !managementTraitsRuntimeSourceTrusted(bundle.Status) {
		return ErrManagementTraitsRuntimeInvalid
	}
	m, err := DecodeManagementTraitsManifest([]byte(bundle.ScoringManifest), budget)
	if err != nil || validateManagementTraitsRuntimeVersions(m.Questionnaire, m.Versions) != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	mc, err := CanonicalManagementTraitsManifest(m)
	if err != nil || mc.SHA256 != bundle.ScoringManifestSHA || bundle.Questionnaire != m.Questionnaire || bundle.ProductVersion != m.Versions.Product || bundle.QuestionVersion != m.Versions.Question || bundle.ScoringVersion != m.Versions.Scoring || bundle.NormVersion != m.Versions.Norm {
		return ErrManagementTraitsRuntimeInvalid
	}
	mapping, err := DecodeManagementTraitsMapping([]byte(profile.MappingSnapshot), m, budget)
	if err != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	pc, err := CanonicalManagementTraitsMapping(m, mapping)
	if err != nil || pc.SHA256 != profile.MappingSHA {
		return ErrManagementTraitsRuntimeInvalid
	}
	var fields managementTraitsRuntimeFieldWire
	if managementTraitsDecodeStrict([]byte(profile.FieldContract), budget, &fields) != nil || fields.Schema != "mng-candidate-fields-v1" || fields.TimePolicy != "candidate-only-personal-25-minutes-v1" || fields.Source != "current-source-not-client-leader-full-not-historical" || fields.CapturedAt != profile.FrozenAt.Format(time.RFC3339Nano) || fields.ManifestSHA != mc.SHA256 || fields.MappingSHA != pc.SHA256 {
		return ErrManagementTraitsRuntimeInvalid
	}
	q, v, err := managementTraitsRuntimeVersions(fields.RepoCode)
	age, flag, text := 1, 0, "present"
	if err != nil || q != m.Questionnaire || v != m.Versions || !managementTraitsRuntimeFields(fields.RequiredFields, managementTraitsRuntimeIdentity{Name: text, Gender: text, Telephone: text, Affiliation: text, Post: text, Age: &age, StuFlag: &flag, Degree: &text, Major: &text}) {
		return ErrManagementTraitsRuntimeInvalid
	}
	return nil
}

// Not exposed to handlers while the complete old-write scope gate is missing.
// Even valid source hashes cannot make this private transaction publicly usable.
func (s *ManagementTraitsRuntimeService) freezeProfileTransaction(ctx context.Context, examID string, clock func() time.Time) (model.ManagementTraitsExamProfile, error) {
	if s.db == nil || !managementTraitsOpaqueID(examID) || clock == nil || s.maxJSONBytes <= 0 {
		return model.ManagementTraitsExamProfile{}, ErrManagementTraitsRuntimeInvalid
	}
	var profile model.ManagementTraitsExamProfile
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exam model.Exam
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", examID).Take(&exam).Error != nil || exam.AssessmentType != "legacy" || exam.ScoringMode != "legacy" {
			return ErrManagementTraitsRuntimeInvalid
		}
		var existing []model.ManagementTraitsExamProfile
		if tx.Where("exam_id = ?", examID).Limit(2).Find(&existing).Error != nil || len(existing) > 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		if len(existing) == 1 {
			var bundle model.ManagementTraitsDefinitionBundle
			if tx.Where("id = ?", existing[0].BundleID).Take(&bundle).Error != nil || validateManagementTraitsRuntimeProfile(existing[0], bundle, s.maxJSONBytes) != nil {
				return ErrManagementTraitsRuntimeInvalid
			}
			profile = existing[0]
			return nil
		}
		draft, err := ReadManagementTraitsDraft(ctx, tx, examID)
		if err != nil {
			return err
		}
		if draft != nil && draft.Lifecycle != "draft" {
			return ErrManagementTraitsRuntimeInvalid
		}
		var paperCount int64
		if tx.Model(&model.Paper{}).Where("exam_id = ?", examID).Count(&paperCount).Error != nil || paperCount != 0 || exam.TotalTime != 25 {
			return ErrManagementTraitsRuntimeInvalid
		}
		var links []model.ExamRepo
		if tx.Where("exam_id = ?", examID).Limit(2).Find(&links).Error != nil || len(links) != 1 || links[0].RadioCount != 140 || links[0].MultiCount != 0 || links[0].JudgeCount != 0 || links[0].SaqCount != 0 {
			return ErrManagementTraitsRuntimeInvalid
		}
		var repo model.Repo
		if tx.Where("id = ?", links[0].RepoID).Take(&repo).Error != nil || repo.RadioCount != 140 || repo.MultiCount != 0 || repo.JudgeCount != 0 {
			return ErrManagementTraitsRuntimeInvalid
		}
		if _, _, err := managementTraitsRuntimeVersions(repo.Code); err != nil {
			return err
		}
		// A new profile requires the persistent creation marker. Existing frozen
		// 002 profiles returned above retain their original canonical source.
		if draft == nil {
			return ErrManagementTraitsDraftPolicy
		}
		if draft != nil && (draft.RepoID != repo.ID || draft.RepoCode != repo.Code) {
			return ErrManagementTraitsRuntimeInvalid
		}
		var global []model.Repo
		if tx.Where("code = ?", repo.Code).Limit(2).Find(&global).Error != nil || len(global) != 1 || global[0].ID != repo.ID {
			return ErrManagementTraitsRuntimeInvalid
		}
		rows, err := loadManagementTraitsRuntimeSource(ctx, tx, repo.ID)
		if err != nil {
			return err
		}
		manifest, mapping, err := BuildManagementTraitsRuntimeCurrentSource(repo.Code, rows)
		if err != nil {
			return err
		}
		mc, err := CanonicalManagementTraitsManifest(manifest)
		if err != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		pc, err := CanonicalManagementTraitsMapping(manifest, mapping)
		if err != nil || len(mc.JSON)+len(pc.JSON) > s.maxJSONBytes {
			return ErrManagementTraitsRuntimeInvalid
		}
		now := clock().Truncate(time.Second)
		fields := managementTraitsRuntimeFieldContract{Schema: "mng-candidate-fields-v1", RequiredFields: strings.Split(exam.RequiredFields, ","), TimePolicy: "candidate-only-personal-25-minutes-v1", Source: "current-source-not-client-leader-full-not-historical", RepoCode: repo.Code, CapturedAt: now, ManifestSHA: mc.SHA256, MappingSHA: pc.SHA256}
		fc, err := managementTraitsCanonicalBytes(fields)
		if err != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var bundles []model.ManagementTraitsDefinitionBundle
		v := manifest.Versions
		if tx.Where("product_version = ? AND question_version = ? AND scoring_version = ? AND norm_version = ?", v.Product, v.Question, v.Scoring, v.Norm).Limit(2).Find(&bundles).Error != nil || len(bundles) > 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		bundle := model.ManagementTraitsDefinitionBundle{ID: uuid.NewString(), ProductVersion: v.Product, QuestionVersion: v.Question, ScoringVersion: v.Scoring, NormVersion: v.Norm, Questionnaire: manifest.Questionnaire, ScoringManifest: string(mc.JSON), ScoringManifestSHA: mc.SHA256, Status: "candidate-current-source", CreatedAt: now}
		if len(bundles) == 1 {
			bundle = bundles[0]
			if bundle.ScoringManifestSHA != mc.SHA256 || bundle.ScoringManifest != string(mc.JSON) || bundle.Status != "candidate-current-source" || bundle.Questionnaire != manifest.Questionnaire {
				return ErrManagementTraitsRuntimeInvalid
			}
		}
		profile = model.ManagementTraitsExamProfile{ExamID: examID, BundleID: bundle.ID, MappingSnapshot: string(pc.JSON), MappingSHA: pc.SHA256, FieldContract: string(fc.JSON), TotalTimeMinutes: 25, FrozenAt: &now, CreatedAt: now}
		if validateManagementTraitsRuntimeProfile(profile, bundle, s.maxJSONBytes) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		if len(bundles) == 0 && tx.Create(&bundle).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		if tx.Create(&profile).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		if draft != nil {
			updatedAt := clock()
			if updatedAt.Before(draft.UpdatedAt) {
				return ErrManagementTraitsRuntimeInvalid
			}
			r := tx.Model(&model.ManagementTraitsExamDraft{}).Where("exam_id = ? AND lifecycle = ?", examID, "draft").Updates(map[string]any{"lifecycle": "frozen", "frozen_at": now, "updated_at": updatedAt})
			if r.Error != nil || r.RowsAffected != 1 {
				return ErrManagementTraitsRuntimeInvalid
			}
		}
		return nil
	})
	if err != nil {
		return model.ManagementTraitsExamProfile{}, ErrManagementTraitsRuntimeInvalid
	}
	return profile, nil
}
