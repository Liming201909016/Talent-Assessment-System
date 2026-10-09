package model

import "time"

// Independent submission facts; no automatic migration, relationships or hooks.
// The polymorphic participant binding is checked by the runtime loader.
type ManagementTraitsRuntimeReceipt struct {
	RunID           string    `gorm:"column:run_id;type:varchar(64);primaryKey" json:"runId"`
	PaperID         string    `gorm:"column:paper_id;type:varchar(64);uniqueIndex:uk_mng_receipt_paper" json:"paperId"`
	ExamID          string    `gorm:"column:exam_id;type:varchar(64)" json:"examId"`
	ParticipantType string    `gorm:"column:participant_type;type:varchar(16)" json:"participantType"`
	ParticipantID   string    `gorm:"column:participant_id;type:varchar(64)" json:"participantId"`
	SubmitType      string    `gorm:"column:submit_type;type:varchar(16)" json:"submitType"`
	StartedAt       time.Time `gorm:"column:started_at" json:"startedAt"`
	LimitTime       time.Time `gorm:"column:limit_time" json:"limitTime"`
	SubmittedAt     time.Time `gorm:"column:submitted_at" json:"submittedAt"`
	UserTimeSeconds int       `gorm:"column:user_time_seconds" json:"userTimeSeconds"`
}

func (ManagementTraitsRuntimeReceipt) TableName() string { return "el_mng_runtime_receipt" }
