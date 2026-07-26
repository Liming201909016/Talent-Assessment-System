package handler

import (
	"errors"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

var (
	errCompetencyQuestionDedicatedAPI = errors.New("胜任力题目请使用胜任力专用接口")
	errPaperReferencedQuestion        = errors.New("question is referenced by paper")
)

func hasCompetencyQuestionMetadata(question model.Qu) bool {
	return question.QuestionCode != nil || question.DimensionID != nil ||
		question.DimensionItemNo != nil || question.ObservationPoint != nil ||
		question.ScoringDirection != nil || question.QuestionStatus != 0
}

func rejectCompetencyQuestionIDs(db *gorm.DB, questionIDs []string) error {
	if len(questionIDs) == 0 {
		return nil
	}
	var count int64
	if err := db.Model(&model.Qu{}).
		Where("id IN ? AND dimension_id IS NOT NULL", questionIDs).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errCompetencyQuestionDedicatedAPI
	}
	return nil
}

func rejectPaperReferencedQuestionIDs(db *gorm.DB, questionIDs []string) error {
	if len(questionIDs) == 0 {
		return nil
	}
	var count int64
	if err := db.Table("el_paper_qu").Where("qu_id IN ?", questionIDs).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errPaperReferencedQuestion
	}
	return nil
}
