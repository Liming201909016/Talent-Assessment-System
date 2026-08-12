package handler

import (
	"fmt"

	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
)

type enabledQuestionCountRow struct {
	DimensionID   string `gorm:"column:dimension_id"`
	QuestionCount int    `gorm:"column:question_count"`
}

type phase1QuestionInventoryRow struct {
	DimensionID   string `gorm:"column:dimension_id"`
	QuestionType  string `gorm:"column:question_type"`
	QuestionCount int    `gorm:"column:question_count"`
}

func loadPhase1QuestionInventory(db *gorm.DB, dimensionIDs []string) (map[string]service.Phase1QuestionTypeCounts, error) {
	counts := make(map[string]service.Phase1QuestionTypeCounts, len(dimensionIDs))
	for _, dimensionID := range dimensionIDs {
		counts[dimensionID] = service.Phase1QuestionTypeCounts{}
	}
	rows := make([]phase1QuestionInventoryRow, 0)
	if err := db.Table("el_qu").
		Select("dimension_id, COALESCE(competency_question_type, '') AS question_type, COUNT(*) AS question_count").
		Where("dimension_id IN ? AND question_status = ?", dimensionIDs, 0).
		Group("dimension_id, competency_question_type").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		count := counts[row.DimensionID]
		switch row.QuestionType {
		case "dimension":
			count.Dimension += row.QuestionCount
		case "validity":
			count.Validity += row.QuestionCount
		default:
			count.Other += row.QuestionCount
		}
		counts[row.DimensionID] = count
	}
	return counts, nil
}

// loadEnabledQuestionCounts uses one grouped query for all requested dimensions.
// An empty dimensionIDs slice means all dimensions.
func loadEnabledQuestionCounts(db *gorm.DB, dimensionIDs []string) (map[string]int, error) {
	counts := make(map[string]int)
	query := db.Table("el_qu").
		Select("dimension_id, COUNT(*) AS question_count").
		Where("dimension_id IS NOT NULL AND question_status = ? AND competency_question_type = ?", 0, "dimension")
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
