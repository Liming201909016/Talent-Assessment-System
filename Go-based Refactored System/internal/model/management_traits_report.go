package model

import "time"

// Test revisions are independent of legacy participant pdf_path and of scoring
// runs. No hooks, auto migrations, implicit associations or formal approval.
type ManagementTraitsReportRevision struct {
	ID              string    `gorm:"column:id;type:varchar(64);primaryKey;uniqueIndex:uk_mng_report_identity,priority:1" json:"id"`
	RunID           string    `gorm:"column:run_id;type:varchar(64);uniqueIndex:uk_mng_report_revision,priority:1" json:"runId"`
	PaperID         string    `gorm:"column:paper_id;type:varchar(64);uniqueIndex:uk_mng_report_identity,priority:2" json:"paperId"`
	ExamID          string    `gorm:"column:exam_id;type:varchar(64)" json:"examId"`
	Revision        int       `gorm:"column:revision;uniqueIndex:uk_mng_report_revision,priority:2" json:"revision"`
	Mode            string    `gorm:"column:mode;type:varchar(16)" json:"mode"`
	TestTitle       string    `gorm:"column:test_title;type:varchar(255)" json:"testTitle"`
	TestLabel       string    `gorm:"column:test_label;type:varchar(255)" json:"testLabel"`
	ContentVersion  string    `gorm:"column:content_version;type:varchar(64)" json:"contentVersion"`
	ContentSHA      string    `gorm:"column:content_sha;type:char(64)" json:"contentSha"`
	TemplateVersion string    `gorm:"column:template_version;type:varchar(64)" json:"templateVersion"`
	TemplateSHA     string    `gorm:"column:template_sha;type:char(64)" json:"templateSha"`
	DataSnapshot    string    `gorm:"column:data_snapshot;type:longtext" json:"-"`
	DataSHA         string    `gorm:"column:data_sha;type:char(64)" json:"dataSha"`
	FileKey         string    `gorm:"column:file_key;type:varchar(255)" json:"-"`
	FileSHA         string    `gorm:"column:file_sha;type:char(64)" json:"fileSha"`
	FileBytes       int64     `gorm:"column:file_bytes" json:"fileBytes"`
	CreatedBy       int64     `gorm:"column:created_by" json:"createdBy"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsReportRevision) TableName() string { return "el_mng_report_revision" }

type ManagementTraitsReportCurrent struct {
	PaperID   string    `gorm:"column:paper_id;type:varchar(64);primaryKey" json:"paperId"`
	ReportID  string    `gorm:"column:report_id;type:varchar(64);uniqueIndex:uk_mng_current_report" json:"reportId"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (ManagementTraitsReportCurrent) TableName() string { return "el_mng_report_current" }

type ManagementTraitsReportAudit struct {
	ID        string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	ReportID  string    `gorm:"column:report_id;type:varchar(64)" json:"reportId"`
	ActorID   int64     `gorm:"column:actor_id" json:"actorId"`
	Action    string    `gorm:"column:action;type:varchar(16)" json:"action"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsReportAudit) TableName() string { return "el_mng_report_audit" }
