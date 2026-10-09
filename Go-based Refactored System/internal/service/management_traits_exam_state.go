package service

import (
	"context"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SetExamState is the only state transition for frozen 005 TEST assessments.
// State 1 is visible/disabled and state 0 permits participant admission.
func (s *ManagementTraitsRuntimeService) SetExamState(ctx context.Context, examID string, state int) (int, error) {
	if s == nil || s.db == nil {
		return 0, ErrManagementTraitsRuntimeClosed
	}
	if !managementTraitsOpaqueID(examID) || (state != 0 && state != 1) {
		return 0, ErrManagementTraitsRuntimeInvalid
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return 0, err
	}
	release := LockManagementTraitsRuntimeFreeze()
	defer release()
	result := 0
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exam model.Exam
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id, state, assessment_type, scoring_mode").Where("id = ?", examID).Take(&exam).Error != nil ||
			exam.AssessmentType != "legacy" || exam.ScoringMode != "legacy" || (exam.State != 0 && exam.State != 1) {
			return ErrManagementTraitsRuntimeInvalid
		}
		var profile model.ManagementTraitsExamProfile
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("exam_id = ?", examID).Take(&profile).Error != nil || profile.FrozenAt == nil || profile.FrozenAt.IsZero() {
			return ErrManagementTraitsRuntimeInvalid
		}
		var bundle model.ManagementTraitsDefinitionBundle
		if tx.Where("id = ?", profile.BundleID).Take(&bundle).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var codes []string
		if tx.Table("el_exam_repo er").Select("r.code").Joins("JOIN el_repo r ON r.id = er.repo_id").Where("er.exam_id = ?", examID).Limit(2).Scan(&codes).Error != nil || len(codes) != 1 || ClassifyManagementTraitsProduct(codes[0]) != ManagementTraitsNew005 {
			return ErrManagementTraitsRuntimeInvalid
		}
		_, versions, err := managementTraitsRuntimeVersions(codes[0])
		if err != nil || bundle.ProductVersion != versions.Product {
			return ErrManagementTraitsRuntimeInvalid
		}
		if exam.State != state {
			update := tx.Model(&model.Exam{}).Where("id = ? AND state = ?", examID, exam.State).Update("state", state)
			if update.Error != nil || update.RowsAffected != 1 {
				return ErrManagementTraitsRuntimeInvalid
			}
		}
		result = state
		return nil
	})
	if err != nil {
		return 0, err
	}
	return result, nil
}
