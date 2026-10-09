package service

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type managementTraitsRuntimeOwner struct {
	Kind        string     `gorm:"column:kind"`
	ID          string     `gorm:"column:id"`
	ExamID      string     `gorm:"column:exam_id"`
	PaperID     string     `gorm:"column:paper_id"`
	DelFlag     int        `gorm:"column:del_flag"`
	Status      string     `gorm:"column:status"`
	Name        string     `gorm:"column:name"`
	Gender      string     `gorm:"column:gender"`
	Telephone   string     `gorm:"column:telephone"`
	Affiliation string     `gorm:"column:affiliation"`
	Post        string     `gorm:"column:post"`
	Age         *int       `gorm:"column:age"`
	Degree      *string    `gorm:"column:degree"`
	Major       *string    `gorm:"column:major"`
	StuFlag     *int       `gorm:"column:stu_flag"`
	EndTime     *time.Time `gorm:"column:end_time"`
}

// No candidate-first fallback, no paper.user_id=101 inference. Even an invalid
// competing row is evidence of ambiguous same-paper binding and is rejected.
func validateManagementTraitsRuntimeOwner(rows []managementTraitsRuntimeOwner, kind, id, examID, paperID string) error {
	if len(rows) != 1 {
		return ErrManagementTraitsRuntimeInvalid
	}
	r := rows[0]
	if r.Kind != kind || r.ID != id || r.ExamID != examID || r.PaperID != paperID || r.DelFlag != 0 || (kind != "candidate" && kind != "tester") || (kind == "tester" && r.Status != "0") {
		return ErrManagementTraitsRuntimeInvalid
	}
	return nil
}

// Tester status is not guessed here: legacy Login does not yet establish an
// executable status policy for the new flow. New tester issuance remains closed.
func loadManagementTraitsRuntimeOwners(ctx context.Context, tx *gorm.DB, paperID string) ([]managementTraitsRuntimeOwner, error) {
	rows := make([]managementTraitsRuntimeOwner, 0, 2)
	const query = `SELECT 'candidate' AS kind, id, exam_id, paper_id, COALESCE(del_flag,-1) AS del_flag, '' AS status,
name, COALESCE(gender,'') AS gender, COALESCE(telephone,'') AS telephone,
COALESCE(affiliation,'') AS affiliation, COALESCE(post,'') AS post, age, degree, major, stu_flag, end_time
FROM el_candidate WHERE paper_id = ?
UNION ALL SELECT 'tester' AS kind, id, exam_id, paper_id, COALESCE(del_flag,-1) AS del_flag,
COALESCE(status,'') AS status, name, COALESCE(gender,'') AS gender,
COALESCE(telephone,'') AS telephone, COALESCE(affiliation,'') AS affiliation,
COALESCE(post,'') AS post, age, degree, major, stu_flag, end_time FROM el_tester WHERE paper_id = ? LIMIT 3`
	if tx.WithContext(ctx).Raw(query, paperID, paperID).Scan(&rows).Error != nil {
		return nil, ErrManagementTraitsRuntimeInvalid
	}
	return rows, nil
}
