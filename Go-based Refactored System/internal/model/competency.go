package model

import (
	"time"

	"github.com/shopspring/decimal"
)

func init() {
	decimal.MarshalJSONWithoutQuotes = true
}

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
	ID               string     `gorm:"column:id;primaryKey"      json:"id"`
	ExamID           string     `gorm:"column:exam_id"            json:"examId"`
	ExamDimensionID  string     `gorm:"column:exam_dimension_id"  json:"examDimensionId"`
	SourceQuID       string     `gorm:"column:source_qu_id"       json:"sourceQuId"`
	QuestionCode     string     `gorm:"column:question_code"      json:"questionCode"`
	DimensionItemNo  int        `gorm:"column:dimension_item_no"  json:"dimensionItemNo"`
	QuestionContent  string     `gorm:"column:question_content"   json:"questionContent"`
	ObservationPoint string     `gorm:"column:observation_point"  json:"observationPoint"`
	ScoringDirection string     `gorm:"column:scoring_direction"  json:"scoringDirection"`
	OptionsSnapshot  string     `gorm:"column:options_snapshot"   json:"optionsSnapshot"`
	SourceUpdateTime *time.Time `gorm:"column:source_update_time" json:"sourceUpdateTime"`
	SnapshotOrder    int        `gorm:"column:snapshot_order"     json:"snapshotOrder"`
	CreateTime       *time.Time `gorm:"column:create_time"        json:"createTime"`
}

func (ExamCompetencyQuestion) TableName() string { return "el_exam_competency_question" }

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
	PaperID                 string           `gorm:"column:paper_id;primaryKey"       json:"paperId"`
	ExamID                  string           `gorm:"column:exam_id"                  json:"examId"`
	TotalQuestionCount      int              `gorm:"column:total_question_count"      json:"totalQuestionCount"`
	AnsweredQuestionCount   int              `gorm:"column:answered_question_count"   json:"answeredQuestionCount"`
	EffectiveDimensionCount int              `gorm:"column:effective_dimension_count" json:"effectiveDimensionCount"`
	OverallScore            decimal.Decimal  `gorm:"column:overall_score;type:decimal(18,6)" json:"overallScore"`
	EvaluationAverage       *decimal.Decimal `gorm:"column:evaluation_average;type:decimal(18,6)" json:"evaluationAverage"`
	ParticipantType         string           `gorm:"column:participant_type" json:"participantType"`
	ParticipantID           string           `gorm:"column:participant_id" json:"participantId"`
	ParticipantName         string           `gorm:"column:participant_name" json:"participantName"`
	ParticipantTelephone    string           `gorm:"column:participant_telephone" json:"participantTelephone"`
	ParticipantAge          *int             `gorm:"column:participant_age" json:"participantAge"`
	ParticipantGender       string           `gorm:"column:participant_gender" json:"participantGender"`
	ParticipantAffiliation  string           `gorm:"column:participant_affiliation" json:"participantAffiliation"`
	ParticipantPost         string           `gorm:"column:participant_post" json:"participantPost"`
	ParticipantDegree       string           `gorm:"column:participant_degree" json:"participantDegree"`
	ParticipantMajor        string           `gorm:"column:participant_major" json:"participantMajor"`
	EvaluationLevel         *string          `gorm:"column:evaluation_level"          json:"evaluationLevel"`
	ReportAudience          string           `gorm:"column:report_audience"           json:"reportAudience"`
	IsComplete              int8             `gorm:"column:is_complete"               json:"isComplete"`
	SubmitType              string           `gorm:"column:submit_type"               json:"submitType"`
	ScoringVersion          string           `gorm:"column:scoring_version"           json:"scoringVersion"`
	SubmittedAt             *time.Time       `gorm:"column:submitted_at"              json:"submittedAt"`
	CreateTime              *time.Time       `gorm:"column:create_time"               json:"createTime"`
	UpdateTime              *time.Time       `gorm:"column:update_time"               json:"updateTime"`
}

func (CompetencyResult) TableName() string { return "el_competency_result" }

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

type CompetencyReport struct {
	ID             string     `gorm:"column:id;primaryKey" json:"id"`
	PaperID        string     `gorm:"column:paper_id" json:"paperId"`
	ExamID         string     `gorm:"column:exam_id" json:"examId"`
	Audience       string     `gorm:"column:audience" json:"audience"`
	ContentVersion string     `gorm:"column:content_version" json:"contentVersion"`
	TextSnapshot   string     `gorm:"column:text_snapshot" json:"-"`
	ScoreSnapshot  string     `gorm:"column:score_snapshot" json:"-"`
	PDFPath        string     `gorm:"column:pdf_path" json:"pdfPath"`
	PDFSHA256      string     `gorm:"column:pdf_sha256" json:"pdfSha256"`
	PDFSize        int64      `gorm:"column:pdf_size" json:"pdfSize"`
	Status         string     `gorm:"column:status" json:"status"`
	ErrorMessage   string     `gorm:"column:error_message" json:"errorMessage"`
	GeneratedBy    *int64     `gorm:"column:generated_by" json:"generatedBy"`
	GeneratedAt    *time.Time `gorm:"column:generated_at" json:"generatedAt"`
	CreateTime     *time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime     *time.Time `gorm:"column:update_time" json:"updateTime"`
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
