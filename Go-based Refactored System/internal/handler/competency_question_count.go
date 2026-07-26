package handler

import (
	"fmt"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

type enabledQuestionCountRow struct {
	DimensionID   string `gorm:"column:dimension_id"`
	QuestionCount int    `gorm:"column:question_count"`
}

// loadEnabledQuestionCounts uses one grouped query for all requested dimensions.
// An empty dimensionIDs slice means all dimensions.
func loadEnabledQuestionCounts(db *gorm.DB, dimensionIDs []string) (map[string]int, error) {
	counts := make(map[string]int)
	query := db.Table("el_qu").
		Select("dimension_id, COUNT(*) AS question_count").
		Where("dimension_id IS NOT NULL AND question_status = ?", 0)
	if len(dimensionIDs) > 0 {
		query = query.Where("dimension_id IN ?", dimensionIDs)
	}
	var rows []enabledQuestionCountRow
	if err := query.Group("dimension_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.DimensionID] = row.QuestionCount
	}
	return counts, nil
}

func validateEnabledQuestionCounts(dimensions []model.CompetencyDimension, counts map[string]int) error {
	for _, dimension := range dimensions {
		if counts[dimension.ID] <= 0 {
			return fmt.Errorf("%s %s没有启用题目", dimension.Code, dimension.Name)
		}
	}
	return nil
}
