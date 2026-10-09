package model

import "time"

// Independent new-creation marker. No relationship hooks or AutoMigrate.
type ManagementTraitsExamDraft struct {
	ExamID    string     `gorm:"column:exam_id;type:varchar(64);primaryKey" json:"examId"`
	RepoID    string     `gorm:"column:repo_id;type:varchar(64)" json:"-"`
	RepoCode  string     `gorm:"column:repo_code;type:varchar(5)" json:"repoCode"`
	Lifecycle string     `gorm:"column:lifecycle;type:varchar(6)" json:"lifecycle"`
	CreatedAt time.Time  `gorm:"column:created_at;type:datetime(6)" json:"-"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:datetime(6)" json:"-"`
	FrozenAt  *time.Time `gorm:"column:frozen_at;type:datetime(6)" json:"-"`
}

func (ManagementTraitsExamDraft) TableName() string { return "el_mng_exam_draft" }
