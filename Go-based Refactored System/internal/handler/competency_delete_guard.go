package handler

import (
	"errors"

	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
)

var errDirectCompetencyDelete = errors.New("胜任力业务数据必须整链删除，请删除所属胜任力测评")

// rejectDirectCompetencyDelete prevents generic paper/participant endpoints
// from bypassing the confirmed competency full-chain transaction.
func rejectDirectCompetencyDelete(db *gorm.DB, entity string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	var table string
	switch entity {
	case "paper":
		table = "el_paper"
	case "candidate":
		table = "el_candidate"
	case "tester":
		table = "el_tester"
	default:
		return errors.New("unsupported competency delete entity")
	}
	var count int64
	if err := db.Table(table+" x").
		Joins("INNER JOIN el_exam e ON e.id = x.exam_id").
		Where("x.id IN ? AND e.assessment_type = ?", ids, service.AssessmentTypeCompetency).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errDirectCompetencyDelete
	}
	return nil
}
