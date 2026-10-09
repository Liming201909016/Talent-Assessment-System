package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// These local declarations do not install database constraints or enforce
// immutability. JSON snapshots and hashes are unvalidated data; Source, Status
// and identity-source strings have no runtime allowlist enforcement here.

type ManagementTraitsDefinitionBundle struct {
	ID                 string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	ProductVersion     string    `gorm:"column:product_version;type:varchar(64);uniqueIndex:uk_mng_bundle_versions,priority:1" json:"productVersion"`
	QuestionVersion    string    `gorm:"column:question_version;type:varchar(64);uniqueIndex:uk_mng_bundle_versions,priority:2" json:"questionVersion"`
	ScoringVersion     string    `gorm:"column:scoring_version;type:varchar(64);uniqueIndex:uk_mng_bundle_versions,priority:3" json:"scoringVersion"`
	NormVersion        string    `gorm:"column:norm_version;type:varchar(64);uniqueIndex:uk_mng_bundle_versions,priority:4" json:"normVersion"`
	Questionnaire      string    `gorm:"column:questionnaire" json:"questionnaire"`
	ScoringManifest    string    `gorm:"column:scoring_manifest;type:longtext" json:"scoringManifest"`
	ScoringManifestSHA string    `gorm:"column:scoring_manifest_sha;type:char(64)" json:"scoringManifestSha"`
	Status             string    `gorm:"column:status" json:"status"`
	CreatedAt          time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsDefinitionBundle) TableName() string { return "el_mng_definition_bundle" }

type ManagementTraitsExamProfile struct {
	ExamID           string     `gorm:"column:exam_id;type:varchar(64);primaryKey" json:"examId"`
	BundleID         string     `gorm:"column:bundle_id;type:varchar(64)" json:"bundleId"`
	MappingSnapshot  string     `gorm:"column:mapping_snapshot;type:longtext" json:"mappingSnapshot"`
	FieldContract    string     `gorm:"column:field_contract;type:longtext" json:"fieldContract"`
	MappingSHA       string     `gorm:"column:mapping_sha;type:char(64)" json:"mappingSha"`
	TotalTimeMinutes int        `gorm:"column:total_time_minutes" json:"totalTimeMinutes"`
	FrozenAt         *time.Time `gorm:"column:frozen_at" json:"frozenAt"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsExamProfile) TableName() string { return "el_mng_exam_profile" }

type ManagementTraitsPaperSnapshot struct {
	PaperID             string     `gorm:"column:paper_id;type:varchar(64);primaryKey;uniqueIndex:uk_mng_snapshot_paper_exam,priority:1" json:"paperId"`
	ExamID              string     `gorm:"column:exam_id;type:varchar(64);uniqueIndex:uk_mng_snapshot_paper_exam,priority:2" json:"examId"`
	ProfileExamID       *string    `gorm:"column:profile_exam_id;type:varchar(64)" json:"profileExamId"`
	BundleID            string     `gorm:"column:bundle_id;type:varchar(64)" json:"bundleId"`
	Source              string     `gorm:"column:source" json:"source"`
	EvidenceSnapshot    string     `gorm:"column:evidence_snapshot;type:longtext" json:"evidenceSnapshot"`
	EvidenceSHA         string     `gorm:"column:evidence_sha;type:char(64)" json:"evidenceSha"`
	MappingSnapshot     string     `gorm:"column:mapping_snapshot;type:longtext" json:"mappingSnapshot"`
	MappingSHA          string     `gorm:"column:mapping_sha;type:char(64)" json:"mappingSha"`
	ScoringManifestSHA  string     `gorm:"column:scoring_manifest_sha;type:char(64)" json:"scoringManifestSha"`
	ParticipantType     string     `gorm:"column:participant_type" json:"participantType"`
	ParticipantID       string     `gorm:"column:participant_id;type:varchar(64)" json:"participantId"`
	ParticipantSnapshot string     `gorm:"column:participant_snapshot;type:longtext" json:"participantSnapshot"`
	FieldContract       string     `gorm:"column:field_contract;type:longtext" json:"fieldContract"`
	IdentitySource      string     `gorm:"column:identity_source" json:"identitySource"`
	SourceCapturedAt    time.Time  `gorm:"column:source_captured_at" json:"sourceCapturedAt"`
	StartedAt           time.Time  `gorm:"column:started_at" json:"startedAt"`
	LimitTime           *time.Time `gorm:"column:limit_time" json:"limitTime"`
	CreatedAt           time.Time  `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsPaperSnapshot) TableName() string { return "el_mng_paper_snapshot" }

type ManagementTraitsPaperQuestionSnapshot struct {
	ID                 string     `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	PaperID            string     `gorm:"column:paper_id;type:varchar(64);uniqueIndex:uk_mng_question_v,priority:1;uniqueIndex:uk_mng_question_pq,priority:1;uniqueIndex:uk_mng_question_order,priority:1" json:"paperId"`
	PaperQuestionID    string     `gorm:"column:paper_question_id;type:varchar(64);uniqueIndex:uk_mng_question_pq,priority:2" json:"paperQuestionId"`
	SourceQuestionID   string     `gorm:"column:source_question_id;type:varchar(64)" json:"sourceQuestionId"`
	Number             int        `gorm:"column:v_number;uniqueIndex:uk_mng_question_v,priority:2" json:"number"`
	DisplayOrder       int        `gorm:"column:display_order;uniqueIndex:uk_mng_question_order,priority:2" json:"displayOrder"`
	DimensionKey       string     `gorm:"column:dimension_key" json:"dimensionKey"`
	Reverse            bool       `gorm:"column:reverse" json:"reverse"`
	Content            string     `gorm:"column:content;type:longtext" json:"content"`
	OptionsSnapshot    string     `gorm:"column:options_snapshot;type:longtext" json:"optionsSnapshot"`
	ScoringSnapshotSHA string     `gorm:"column:scoring_snapshot_sha;type:char(64)" json:"scoringSnapshotSha"`
	SelectedOptionID   *string    `gorm:"column:selected_option_id;type:varchar(64)" json:"selectedOptionId"`
	RawAnswer          *int       `gorm:"column:raw_answer" json:"rawAnswer"`
	FinalScore         *int       `gorm:"column:final_score" json:"finalScore"`
	SubmittedAt        *time.Time `gorm:"column:submitted_at" json:"submittedAt"`
	CreatedAt          time.Time  `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsPaperQuestionSnapshot) TableName() string {
	return "el_mng_paper_question_snapshot"
}

type ManagementTraitsResultRun struct {
	ID                    string           `gorm:"column:id;type:varchar(64);primaryKey;uniqueIndex:uk_mng_run_identity,priority:1" json:"id"`
	PaperID               string           `gorm:"column:paper_id;type:varchar(64);uniqueIndex:uk_mng_run_versions,priority:1;uniqueIndex:uk_mng_run_identity,priority:2" json:"paperId"`
	ExamID                string           `gorm:"column:exam_id;type:varchar(64);uniqueIndex:uk_mng_run_identity,priority:3" json:"examId"`
	ProductVersion        string           `gorm:"column:product_version;type:varchar(64)" json:"productVersion"`
	QuestionVersion       string           `gorm:"column:question_version;type:varchar(64)" json:"questionVersion"`
	ScoringVersion        string           `gorm:"column:scoring_version;type:varchar(64);uniqueIndex:uk_mng_run_versions,priority:2" json:"scoringVersion"`
	NormVersion           string           `gorm:"column:norm_version;type:varchar(64);uniqueIndex:uk_mng_run_versions,priority:3" json:"normVersion"`
	Questionnaire         string           `gorm:"column:questionnaire" json:"questionnaire"`
	ParticipantType       string           `gorm:"column:participant_type" json:"participantType"`
	ParticipantID         string           `gorm:"column:participant_id;type:varchar(64)" json:"participantId"`
	ScoringManifestSHA    string           `gorm:"column:scoring_manifest_sha;type:char(64)" json:"scoringManifestSha"`
	InputSHA              string           `gorm:"column:input_sha;type:char(64)" json:"inputSha"`
	Status                string           `gorm:"column:status" json:"status"`
	Source                string           `gorm:"column:source" json:"source"`
	TotalQuestionCount    int              `gorm:"column:total_question_count" json:"totalQuestionCount"`
	AnsweredQuestionCount int              `gorm:"column:answered_question_count" json:"answeredQuestionCount"`
	OverallScore          *decimal.Decimal `gorm:"column:overall_score;type:decimal(18,6)" json:"overallScore"`
	OverallNorm           *decimal.Decimal `gorm:"column:overall_norm;type:decimal(18,6)" json:"overallNorm"`
	OverallLevel          *string          `gorm:"column:overall_level" json:"overallLevel"`
	UserTimeSeconds       *int             `gorm:"column:user_time_seconds" json:"userTimeSeconds"`
	SubmittedAt           *time.Time       `gorm:"column:submitted_at" json:"submittedAt"`
	CreatedAt             time.Time        `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsResultRun) TableName() string { return "el_mng_result_run" }

type ManagementTraitsResultDimension struct {
	ID            string           `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	RunID         string           `gorm:"column:run_id;type:varchar(64);uniqueIndex:uk_mng_dimension_key,priority:1;uniqueIndex:uk_mng_dimension_order,priority:1" json:"runId"`
	DimensionKey  string           `gorm:"column:dimension_key;uniqueIndex:uk_mng_dimension_key,priority:2" json:"dimensionKey"`
	DisplayOrder  int              `gorm:"column:display_order;uniqueIndex:uk_mng_dimension_order,priority:2" json:"displayOrder"`
	DimensionName string           `gorm:"column:dimension_name" json:"dimensionName"`
	ModuleKey     string           `gorm:"column:module_key" json:"moduleKey"`
	QuestionCount int              `gorm:"column:question_count" json:"questionCount"`
	AnsweredCount int              `gorm:"column:answered_count" json:"answeredCount"`
	ScoreSum      int              `gorm:"column:score_sum" json:"scoreSum"`
	Score         *decimal.Decimal `gorm:"column:score;type:decimal(18,6)" json:"score"`
	Norm          *decimal.Decimal `gorm:"column:norm;type:decimal(18,6)" json:"norm"`
	Level         *string          `gorm:"column:level" json:"level"`
	CreatedAt     time.Time        `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsResultDimension) TableName() string { return "el_mng_result_dimension" }

type ManagementTraitsResultModule struct {
	ID             string           `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	RunID          string           `gorm:"column:run_id;type:varchar(64);uniqueIndex:uk_mng_module_key,priority:1;uniqueIndex:uk_mng_module_order,priority:1" json:"runId"`
	ModuleKey      string           `gorm:"column:module_key;uniqueIndex:uk_mng_module_key,priority:2" json:"moduleKey"`
	DisplayOrder   int              `gorm:"column:display_order;uniqueIndex:uk_mng_module_order,priority:2" json:"displayOrder"`
	DimensionCount int              `gorm:"column:dimension_count" json:"dimensionCount"`
	Score          *decimal.Decimal `gorm:"column:score;type:decimal(18,6)" json:"score"`
	CreatedAt      time.Time        `gorm:"column:created_at" json:"createdAt"`
}

func (ManagementTraitsResultModule) TableName() string { return "el_mng_result_module" }
