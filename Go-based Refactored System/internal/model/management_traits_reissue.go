package model

import "time"

// Independent archival output; no legacy pdf_path or TEST current association.
type ManagementTraitsReportReissue struct {
	ID string `gorm:"column:id;type:varchar(64);primaryKey;uniqueIndex:uk_mng_reissue_identity,priority:1" json:"id"`
	RunID string `gorm:"column:run_id;type:varchar(64);uniqueIndex:uk_mng_reissue_input,priority:1" json:"runId"`
	PaperID string `gorm:"column:paper_id;type:varchar(64);uniqueIndex:uk_mng_reissue_identity,priority:2;index:idx_mng_reissue_paper,priority:1" json:"paperId"`
	ExamID string `gorm:"column:exam_id;type:varchar(64);uniqueIndex:uk_mng_reissue_identity,priority:3" json:"examId"`
	Kind string `gorm:"column:kind;type:varchar(32)" json:"kind"`
	Status string `gorm:"column:status;type:varchar(16)" json:"status"`
	DataSnapshot string `gorm:"column:data_snapshot;type:longtext" json:"-"`
	DataSHA string `gorm:"column:data_sha;type:char(64);uniqueIndex:uk_mng_reissue_input,priority:2" json:"-"`
	TemplateSHA string `gorm:"column:template_sha;type:char(64);uniqueIndex:uk_mng_reissue_input,priority:3" json:"templateSha"`
	ContentSHA string `gorm:"column:content_sha;type:char(64);uniqueIndex:uk_mng_reissue_input,priority:4" json:"contentSha"`
	FileKey string `gorm:"column:file_key;type:varchar(255)" json:"-"`
	FileSHA string `gorm:"column:file_sha;type:char(64)" json:"fileSha"`
	FileBytes int64 `gorm:"column:file_bytes;type:bigint" json:"fileBytes"`
	CreatedBy int64 `gorm:"column:created_by;type:bigint" json:"createdBy"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime(6);index:idx_mng_reissue_paper,priority:2" json:"createdAt"`
}

func (ManagementTraitsReportReissue) TableName() string { return "el_mng_report_reissue" }

type ManagementTraitsReissueAudit struct {
	ID string `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	ReportID string `gorm:"column:report_id;type:varchar(64);index:idx_mng_reissue_audit,priority:1" json:"reportId"`
	PaperID string `gorm:"column:paper_id;type:varchar(64)" json:"paperId"`
	ExamID string `gorm:"column:exam_id;type:varchar(64)" json:"examId"`
	ActorID int64 `gorm:"column:actor_id;type:bigint" json:"actorId"`
	Action string `gorm:"column:action;type:varchar(16)" json:"action"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime(6);index:idx_mng_reissue_audit,priority:2" json:"createdAt"`
}

func (ManagementTraitsReissueAudit) TableName() string { return "el_mng_reissue_audit" }