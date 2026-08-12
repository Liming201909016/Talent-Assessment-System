package model

import (
	"time"

	"github.com/shopspring/decimal"
)

func init() {
	decimal.MarshalJSONWithoutQuotes = true
}

const (
	CompetencyQuestionTypeDimension = "dimension"
	CompetencyQuestionTypeValidity  = "validity"
)

// CompetencyDimension 胜任力维度主数据 el_competency_dimension。
type CompetencyDimension struct {
	ID                 string     `gorm:"column:id;primaryKey"          json:"id"`
	Code               string     `gorm:"column:code"                   json:"code"`
	Name               string     `gorm:"column:name"                   json:"name"`
	VIRDLevel          string     `gorm:"column:vird_level"             json:"virdLevel"`
	ApplicableCategory string     `gorm:"column:applicable_category"    json:"applicableCategory"`
	CoreMeaning        string     `gorm:"column:core_meaning"           json:"coreMeaning"`
	DisplayOrder       int        `gorm:"column:display_order"          json:"displayOrder"`
	Status             int8       `gorm:"column:status"                 json:"status"`
	CreateTime         *time.Time `gorm:"column:create_time"            json:"createTime"`
	UpdateTime         *time.Time `gorm:"column:update_time"            json:"updateTime"`
}

func (CompetencyDimension) TableName() string { return "el_competency_dimension" }

// ExamCompetencyDimension 保存胜任力测评草稿选择；发布时补齐快照字段并冻结。
type ExamCompetencyDimension struct {
	ID                 string     `gorm:"column:id;primaryKey"       json:"id"`
	ExamID             string     `gorm:"column:exam_id"             json:"examId"`
	GroupID            *string    `gorm:"column:group_id"            json:"groupId"`
	DimensionID        string     `gorm:"column:dimension_id"        json:"dimensionId"`
	DimensionCode      string     `gorm:"column:dimension_code"      json:"dimensionCode"`
	DimensionName      string     `gorm:"column:dimension_name"      json:"dimensionName"`
	VIRDLevel          string     `gorm:"column:vird_level"          json:"virdLevel"`
	ApplicableCategory string     `gorm:"column:applicable_category" json:"applicableCategory"`
	CoreMeaning        string     `gorm:"column:core_meaning"        json:"coreMeaning"`
	DisplayOrder       int        `gorm:"column:display_order"       json:"displayOrder"`
	QuestionCount      int        `gorm:"column:question_count"      json:"questionCount"`
	CreateTime         *time.Time `gorm:"column:create_time"         json:"createTime"`
	SnapshotTime       *time.Time `gorm:"column:snapshot_time"       json:"snapshotTime"`
}

func (ExamCompetencyDimension) TableName() string { return "el_exam_competency_dimension" }

// ExamCompetencyQuestion is the immutable source used by every participant paper.
type ExamCompetencyQuestion struct {
	ID                     string     `gorm:"column:id;primaryKey"              json:"id"`
	ExamID                 string     `gorm:"column:exam_id"                    json:"examId"`
	ExamDimensionID        string     `gorm:"column:exam_dimension_id"          json:"examDimensionId"`
	SourceQuID             string     `gorm:"column:source_qu_id"               json:"sourceQuId"`
	QuestionCode           string     `gorm:"column:question_code"              json:"questionCode"`
	CompetencyQuestionType *string    `gorm:"column:competency_question_type"    json:"competencyQuestionType"`
	DimensionItemNo        int        `gorm:"column:dimension_item_no"          json:"dimensionItemNo"`
	QuestionContent        string     `gorm:"column:question_content"           json:"questionContent"`
	ObservationPoint       string     `gorm:"column:observation_point"          json:"observationPoint"`
	ScoringDirection       string     `gorm:"column:scoring_direction"          json:"scoringDirection"`
	OptionsSnapshot        string     `gorm:"column:options_snapshot"           json:"optionsSnapshot"`
	SourceUpdateTime       *time.Time `gorm:"column:source_update_time"         json:"sourceUpdateTime"`
	SnapshotOrder          int        `gorm:"column:snapshot_order"             json:"snapshotOrder"`
	CreateTime             *time.Time `gorm:"column:create_time"                json:"createTime"`
}

func (ExamCompetencyQuestion) TableName() string { return "el_exam_competency_question" }

// ExamCompetencyGroup freezes one first-level competency group and its selected dimensions for an exam.
type ExamCompetencyGroup struct {
	ID             string     `gorm:"column:id;primaryKey"       json:"id"`
	ExamID         string     `gorm:"column:exam_id"             json:"examId"`
	GroupCode      string     `gorm:"column:group_code"          json:"groupCode"`
	GroupName      string     `gorm:"column:group_name"          json:"groupName"`
	DisplayOrder   int        `gorm:"column:display_order"       json:"displayOrder"`
	DimensionCount int        `gorm:"column:dimension_count"     json:"dimensionCount"`
	QuestionCount  int        `gorm:"column:question_count"      json:"questionCount"`
	CreateTime     *time.Time `gorm:"column:create_time"         json:"createTime"`
	SnapshotTime   *time.Time `gorm:"column:snapshot_time"       json:"snapshotTime"`
}

func (ExamCompetencyGroup) TableName() string { return "el_exam_competency_group" }

type CompetencyDimensionResult struct {
	ID                    string           `gorm:"column:id;primaryKey"         json:"id"`
	PaperID               string           `gorm:"column:paper_id"              json:"paperId"`
	ExamDimensionID       string           `gorm:"column:exam_dimension_id"     json:"examDimensionId"`
	DimensionID           string           `gorm:"column:dimension_id"          json:"dimensionId"`
	DimensionCode         string           `gorm:"column:dimension_code"        json:"dimensionCode"`
	DimensionName         string           `gorm:"column:dimension_name"        json:"dimensionName"`
	DisplayOrder          int              `gorm:"column:display_order"         json:"displayOrder"`
	TotalQuestionCount    int              `gorm:"column:total_question_count"  json:"totalQuestionCount"`
	AnsweredQuestionCount int              `gorm:"column:answered_question_count" json:"answeredQuestionCount"`
	ScoreSum              int              `gorm:"column:score_sum"             json:"scoreSum"`
	DimensionScore        *decimal.Decimal `gorm:"column:dimension_score;type:decimal(18,6)" json:"dimensionScore"`
	LevelCode             *string          `gorm:"column:level_code"            json:"levelCode"`
	IsComplete            int8             `gorm:"column:is_complete"           json:"isComplete"`
	CreateTime            *time.Time       `gorm:"column:create_time"           json:"createTime"`
}

func (CompetencyDimensionResult) TableName() string { return "el_competency_dimension_result" }

type CompetencyResult struct {
	PaperID                        string           `gorm:"column:paper_id;primaryKey"       json:"paperId"`
	ExamID                         string           `gorm:"column:exam_id"                  json:"examId"`
	TotalQuestionCount             int              `gorm:"column:total_question_count"      json:"totalQuestionCount"`
	AnsweredQuestionCount          int              `gorm:"column:answered_question_count"   json:"answeredQuestionCount"`
	DimensionQuestionCount         int              `gorm:"column:dimension_question_count"  json:"dimensionQuestionCount"`
	AnsweredDimensionQuestionCount int              `gorm:"column:answered_dimension_question_count" json:"answeredDimensionQuestionCount"`
	EffectiveDimensionCount        int              `gorm:"column:effective_dimension_count" json:"effectiveDimensionCount"`
	OverallScore                   *decimal.Decimal `gorm:"column:overall_score;type:decimal(18,6)" json:"overallScore"`
	EvaluationAverage              *decimal.Decimal `gorm:"column:evaluation_average;type:decimal(18,6)" json:"evaluationAverage"`
	ParticipantType                string           `gorm:"column:participant_type" json:"participantType"`
	ParticipantID                  string           `gorm:"column:participant_id" json:"participantId"`
	ParticipantName                string           `gorm:"column:participant_name" json:"participantName"`
	ParticipantTelephone           string           `gorm:"column:participant_telephone" json:"participantTelephone"`
	ParticipantAge                 *int             `gorm:"column:participant_age" json:"participantAge"`
	ParticipantGender              string           `gorm:"column:participant_gender" json:"participantGender"`
	ParticipantAffiliation         string           `gorm:"column:participant_affiliation" json:"participantAffiliation"`
	ParticipantPost                string           `gorm:"column:participant_post" json:"participantPost"`
	ParticipantDegree              string           `gorm:"column:participant_degree" json:"participantDegree"`
	ParticipantMajor               string           `gorm:"column:participant_major" json:"participantMajor"`
	EvaluationLevel                *string          `gorm:"column:evaluation_level"          json:"evaluationLevel"`
	ReportAudience                 string           `gorm:"column:report_audience"           json:"reportAudience"`
	IsComplete                     int8             `gorm:"column:is_complete"               json:"isComplete"`
	SubmitType                     string           `gorm:"column:submit_type"               json:"submitType"`
	ProductVersion                 string           `gorm:"column:product_version"           json:"productVersion"`
	ScoringVersion                 string           `gorm:"column:scoring_version"           json:"scoringVersion"`
	ContentVersion                 string           `gorm:"column:content_version"           json:"contentVersion"`
	ReportTemplateVersion          string           `gorm:"column:report_template_version"   json:"reportTemplateVersion"`
	SubmittedAt                    *time.Time       `gorm:"column:submitted_at"              json:"submittedAt"`
	CreateTime                     *time.Time       `gorm:"column:create_time"               json:"createTime"`
	UpdateTime                     *time.Time       `gorm:"column:update_time"               json:"updateTime"`
}

func (CompetencyResult) TableName() string { return "el_competency_result" }

// CompetencyGroupResult stores a first-level aggregate without defining how its score or level is calculated.
type CompetencyGroupResult struct {
	ID                      string           `gorm:"column:id;primaryKey"              json:"id"`
	PaperID                 string           `gorm:"column:paper_id"                   json:"paperId"`
	ExamGroupID             string           `gorm:"column:exam_group_id"              json:"examGroupId"`
	GroupCode               string           `gorm:"column:group_code"                 json:"groupCode"`
	GroupName               string           `gorm:"column:group_name"                 json:"groupName"`
	DisplayOrder            int              `gorm:"column:display_order"              json:"displayOrder"`
	TotalDimensionCount     int              `gorm:"column:total_dimension_count"       json:"totalDimensionCount"`
	EffectiveDimensionCount int              `gorm:"column:effective_dimension_count"   json:"effectiveDimensionCount"`
	TotalQuestionCount      int              `gorm:"column:total_question_count"        json:"totalQuestionCount"`
	AnsweredQuestionCount   int              `gorm:"column:answered_question_count"     json:"answeredQuestionCount"`
	GroupScore              *decimal.Decimal `gorm:"column:group_score;type:decimal(18,6)" json:"groupScore"`
	LevelCode               *string          `gorm:"column:level_code"                  json:"levelCode"`
	IsComplete              int8             `gorm:"column:is_complete"                 json:"isComplete"`
	ScoringVersion          string           `gorm:"column:scoring_version"             json:"scoringVersion"`
	CreateTime              *time.Time       `gorm:"column:create_time"                 json:"createTime"`
}

func (CompetencyGroupResult) TableName() string { return "el_competency_group_result" }

// CompetencyValidityResult stores validity output while unresolved score/status rules remain nullable.
type CompetencyValidityResult struct {
	PaperID               string           `gorm:"column:paper_id;primaryKey"       json:"paperId"`
	TotalQuestionCount    int              `gorm:"column:total_question_count"      json:"totalQuestionCount"`
	AnsweredQuestionCount int              `gorm:"column:answered_question_count"   json:"answeredQuestionCount"`
	ValidityScore         *decimal.Decimal `gorm:"column:validity_score;type:decimal(18,6)" json:"validityScore"`
	ValidityStatus        *string          `gorm:"column:validity_status"           json:"validityStatus"`
	IsComplete            int8             `gorm:"column:is_complete"               json:"isComplete"`
	ScoringVersion        string           `gorm:"column:scoring_version"           json:"scoringVersion"`
	CreateTime            *time.Time       `gorm:"column:create_time"               json:"createTime"`
	UpdateTime            *time.Time       `gorm:"column:update_time"               json:"updateTime"`
}

func (CompetencyValidityResult) TableName() string { return "el_competency_validity_result" }

type CompetencyReportText struct {
	ID             string     `gorm:"column:id;primaryKey" json:"id"`
	ContentVersion string     `gorm:"column:content_version" json:"contentVersion"`
	Audience       string     `gorm:"column:audience" json:"audience"`
	ContentType    string     `gorm:"column:content_type" json:"contentType"`
	DimensionID    string     `gorm:"column:dimension_id" json:"dimensionId"`
	LevelCode      string     `gorm:"column:level_code" json:"levelCode"`
	Content        string     `gorm:"column:content" json:"content"`
	Disclaimer     string     `gorm:"column:disclaimer" json:"disclaimer"`
	IsTemporary    int8       `gorm:"column:is_temporary" json:"isTemporary"`
	Status         int8       `gorm:"column:status" json:"status"`
	CreateTime     *time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime     *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (CompetencyReportText) TableName() string { return "el_competency_report_text" }

// CompetencyReportContentPackage records the independent content and psychometric approvals
// required before a phase-1 report content version can be used in a formal environment.
type CompetencyReportContentPackage struct {
	ID                     string     `gorm:"column:id;primaryKey" json:"id"`
	ProductVersion         string     `gorm:"column:product_version" json:"productVersion"`
	ScoringVersion         string     `gorm:"column:scoring_version" json:"scoringVersion"`
	ContentVersion         string     `gorm:"column:content_version" json:"contentVersion"`
	TemplateVersion        string     `gorm:"column:template_version" json:"templateVersion"`
	Audience               string     `gorm:"column:audience" json:"audience"`
	ApprovalStatus         string     `gorm:"column:approval_status" json:"approvalStatus"`
	ContentApprovedBy      string     `gorm:"column:content_approved_by" json:"contentApprovedBy"`
	ContentApprovedAt      *time.Time `gorm:"column:content_approved_at" json:"contentApprovedAt"`
	PsychometricApprovedBy string     `gorm:"column:psychometric_approved_by" json:"psychometricApprovedBy"`
	PsychometricApprovedAt *time.Time `gorm:"column:psychometric_approved_at" json:"psychometricApprovedAt"`
	QuestionSourceSHA256   string     `gorm:"column:question_source_sha256" json:"questionSourceSha256"`
	ContentSourceSHA256    string     `gorm:"column:content_source_sha256" json:"contentSourceSha256"`
	EffectiveEnvironment   string     `gorm:"column:effective_environment" json:"effectiveEnvironment"`
	Disclaimer             string     `gorm:"column:disclaimer" json:"disclaimer"`
	CreateTime             *time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime             *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (CompetencyReportContentPackage) TableName() string {
	return "el_competency_report_content_package"
}

type CompetencyReport struct {
	ID              string     `gorm:"column:id;primaryKey" json:"id"`
	PaperID         string     `gorm:"column:paper_id" json:"paperId"`
	ExamID          string     `gorm:"column:exam_id" json:"examId"`
	Audience        string     `gorm:"column:audience" json:"audience"`
	ContentVersion  string     `gorm:"column:content_version" json:"contentVersion"`
	TemplateVersion string     `gorm:"column:template_version" json:"templateVersion"`
	TextSnapshot    string     `gorm:"column:text_snapshot" json:"-"`
	ScoreSnapshot   string     `gorm:"column:score_snapshot" json:"-"`
	PDFPath         string     `gorm:"column:pdf_path" json:"pdfPath"`
	PDFSHA256       string     `gorm:"column:pdf_sha256" json:"pdfSha256"`
	PDFSize         int64      `gorm:"column:pdf_size" json:"pdfSize"`
	Status          string     `gorm:"column:status" json:"status"`
	ErrorMessage    string     `gorm:"column:error_message" json:"errorMessage"`
	GeneratedBy     *int64     `gorm:"column:generated_by" json:"generatedBy"`
	GeneratedAt     *time.Time `gorm:"column:generated_at" json:"generatedAt"`
	CreateTime      *time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime      *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (CompetencyReport) TableName() string { return "el_competency_report" }

type CompetencyReportAudit struct {
	ID           string     `gorm:"column:id;primaryKey" json:"id"`
	ReportID     string     `gorm:"column:report_id" json:"reportId"`
	PaperID      string     `gorm:"column:paper_id" json:"paperId"`
	Action       string     `gorm:"column:action" json:"action"`
	OperatorID   *int64     `gorm:"column:operator_id" json:"operatorId"`
	Status       int8       `gorm:"column:status" json:"status"`
	ErrorMessage string     `gorm:"column:error_message" json:"errorMessage"`
	ClientIP     string     `gorm:"column:client_ip" json:"clientIp"`
	CreateTime   *time.Time `gorm:"column:create_time" json:"createTime"`
}

func (CompetencyReportAudit) TableName() string { return "el_competency_report_audit" }
