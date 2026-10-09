package service

import (
	"context"

	"github.com/talent-assessment/refactored/internal/model"
)

// The list is persisted run metadata, not an approval or report-ready verdict.
// Detail revalidates the complete frozen input and immutable 13/4 result set.
func (s *ManagementTraitsRuntimeService) ListRuns(ctx context.Context, examID string) ([]model.ManagementTraitsResultRun, error) {
	rows := make([]model.ManagementTraitsResultRun, 0)
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return rows, err
	}
	if !managementTraitsOpaqueID(examID) {
		return rows, ErrManagementTraitsRuntimeInvalid
	}
	if s.db.WithContext(ctx).Table("el_mng_result_run r").Select("r.*").Joins("INNER JOIN el_mng_paper_snapshot s ON s.paper_id=r.paper_id AND s.exam_id=r.exam_id").Joins("INNER JOIN el_mng_definition_bundle b ON b.id=s.bundle_id").Where("r.exam_id = ? AND s.source = ? AND b.status IN ?", examID, "new_creation", []string{"candidate-current-source", "retired"}).Order("r.submitted_at DESC,r.id").Limit(201).Find(&rows).Error != nil || len(rows) > 200 {
		return make([]model.ManagementTraitsResultRun, 0), ErrManagementTraitsRuntimeInvalid
	}
	for _, r := range rows {
		if r.ExamID != examID {
			return make([]model.ManagementTraitsResultRun, 0), ErrManagementTraitsRuntimeInvalid
		}
	}
	return rows, nil
}

func (s *ManagementTraitsRuntimeService) ResultDetail(ctx context.Context, runID string) (ManagementTraitsValidatedRuntimeRun, error) {
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return ManagementTraitsValidatedRuntimeRun{}, err
	}
	return s.LoadValidatedRun(ctx, runID)
}
