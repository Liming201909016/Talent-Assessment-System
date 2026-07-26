package handler

import (
	"errors"

	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
)

var (
	ErrLegacyPaperNotFound    = errors.New("legacy paper not found")
	ErrCompetencyDedicatedAPI = errors.New("competency assessment requires dedicated API")
	ErrInvalidAssessmentMode  = errors.New("invalid assessment mode")
)

// requireLegacyExam rejects competency and invalid type/scoring combinations
// before legacy paper creation can write any rows.
func requireLegacyExam(exam *model.Exam) error {
	isCompetency, err := service.ValidateAssessmentMode(exam.AssessmentType, exam.ScoringMode)
	if err != nil {
		return ErrInvalidAssessmentMode
	}
	if isCompetency || exam.AssessmentType == service.AssessmentTypeCompetency || exam.ScoringMode == service.ScoringModeCompetencyAverage {
		return ErrCompetencyDedicatedAPI
	}
	return nil
}

// requireLegacyPaper resolves the paper's owning exam in one indexed JOIN. A
// lookup failure never falls back to legacy behavior.
func requireLegacyPaper(db *gorm.DB, paperID string) error {
	var mode struct {
		AssessmentType string `gorm:"column:assessment_type"`
		ScoringMode    string `gorm:"column:scoring_mode"`
	}
	err := db.Table("el_paper p").
		Select("e.assessment_type, e.scoring_mode").
		Joins("INNER JOIN el_exam e ON e.id = p.exam_id").
		Where("p.id = ?", paperID).
		Take(&mode).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrLegacyPaperNotFound
	}
	if err != nil {
		return ErrInvalidAssessmentMode
	}
	return requireLegacyExam(&model.Exam{
		AssessmentType: mode.AssessmentType,
		ScoringMode:    mode.ScoringMode,
	})
}

func legacyPaperGuardMessage(err error, notFoundMessage string) string {
	switch {
	case errors.Is(err, ErrLegacyPaperNotFound):
		return notFoundMessage
	case errors.Is(err, ErrCompetencyDedicatedAPI):
		return "胜任力测评请使用胜任力专用接口"
	default:
		return "测评类型配置错误"
	}
}
