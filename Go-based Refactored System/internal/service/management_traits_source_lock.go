package service

import (
	"context"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

// LOCK IN SHARE MODE is a current read on MySQL 5.7 and 8, not an MVCC
// snapshot. All papers can share this source lease; an external status UPDATE
// needs an exclusive row lock and waits until this transaction ends. Never
// upgrade this lease to UPDATE or acquire an existing paper after this lock.
func loadManagementTraitsWriteBundle(ctx context.Context, tx *gorm.DB, id string) (model.ManagementTraitsDefinitionBundle, error) {
	var bundle model.ManagementTraitsDefinitionBundle
	managementRaceMark(ctx, "bundle_lock_wait")
	q := tx.WithContext(ctx).Raw("SELECT * FROM el_mng_definition_bundle WHERE id = ? LOCK IN SHARE MODE", id).Scan(&bundle)
	if q.Error == nil && q.RowsAffected == 1 && bundle.ID == id {
		managementRaceMark(ctx, "bundle_lock_acquired")
	}
	if q.Error != nil || q.RowsAffected != 1 || bundle.ID != id || !managementTraitsRuntimeSourceTrusted(bundle.Status) {
		return model.ManagementTraitsDefinitionBundle{}, ErrManagementTraitsRuntimeInvalid
	}
	return bundle, nil
}

// The plain loader has already validated the frozen contract. Re-read its
// source under a transaction-owned lease and reject content/version drift as
// well as revoked status. Retirement alone must not invalidate frozen papers.
func revalidateManagementTraitsWriteBundle(ctx context.Context, tx *gorm.DB, expected model.ManagementTraitsDefinitionBundle) (model.ManagementTraitsDefinitionBundle, error) {
	current, err := loadManagementTraitsWriteBundle(ctx, tx, expected.ID)
	if err != nil || current.ProductVersion != expected.ProductVersion || current.QuestionVersion != expected.QuestionVersion || current.ScoringVersion != expected.ScoringVersion || current.NormVersion != expected.NormVersion || current.Questionnaire != expected.Questionnaire || current.ScoringManifest != expected.ScoringManifest || current.ScoringManifestSHA != expected.ScoringManifestSHA || !current.CreatedAt.Equal(expected.CreatedAt) {
		return model.ManagementTraitsDefinitionBundle{}, ErrManagementTraitsRuntimeInvalid
	}
	return current, nil
}

// Existing-paper writes always acquire paper UPDATE before bundle SHARE.
// The unchanged read loader remains usable by identity and result consumers.
func (s *ManagementTraitsRuntimeService) loadRuntimePaperForWrite(ctx context.Context, tx *gorm.DB, paperID string) (model.Paper, model.ManagementTraitsPaperSnapshot, model.ManagementTraitsDefinitionBundle, model.ManagementTraitsExamProfile, managementTraitsRuntimeOwner, []model.ManagementTraitsPaperQuestionSnapshot, error) {
	paper, snapshot, bundle, profile, owner, questions, err := s.loadRuntimePaper(ctx, tx, paperID)
	if err == nil {
		bundle, err = revalidateManagementTraitsWriteBundle(ctx, tx, bundle)
	}
	return paper, snapshot, bundle, profile, owner, questions, err
}
