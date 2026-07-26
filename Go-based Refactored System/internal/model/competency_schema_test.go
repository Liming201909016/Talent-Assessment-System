package model

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestExam_CompetencyDispatchFields(t *testing.T) {
	type fieldExpectation struct {
		name    string
		gormTag string
		jsonTag string
	}
	expected := []fieldExpectation{
		{"AssessmentType", "column:assessment_type", "assessmentType"},
		{"ScoringMode", "column:scoring_mode", "scoringMode"},
		{"CompetencyReportAudience", "column:competency_report_audience", "competencyReportAudience"},
		{"PublishStatus", "column:publish_status", "publishStatus"},
		{"PublishedAt", "column:published_at", "publishedAt"},
		{"PublishedBy", "column:published_by", "publishedBy"},
	}

	typ := reflect.TypeOf(Exam{})
	for _, want := range expected {
		field, ok := typ.FieldByName(want.name)
		if !ok {
			t.Errorf("Exam missing field %s", want.name)
			continue
		}
		if field.Tag.Get("gorm") != want.gormTag {
			t.Errorf("Exam.%s gorm tag = %q, want %q", want.name, field.Tag.Get("gorm"), want.gormTag)
		}
		if field.Tag.Get("json") != want.jsonTag {
			t.Errorf("Exam.%s json tag = %q, want %q", want.name, field.Tag.Get("json"), want.jsonTag)
		}
	}
}

func TestCompetencyExamDispatchMigration_IsMySQL57Idempotent(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	migrationPath := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_001_schema.sql"))
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration failed: %v", err)
	}
	sql := string(data)

	for _, column := range []string{"assessment_type", "scoring_mode", "competency_report_audience", "publish_status", "published_at", "published_by"} {
		if !strings.Contains(sql, "COLUMN_NAME = '"+column+"'") {
			t.Errorf("migration does not check information_schema for column %s", column)
		}
		if !strings.Contains(sql, "ADD COLUMN `"+column+"`") {
			t.Errorf("migration does not add column %s", column)
		}
	}
	for _, required := range []string{
		"TABLE_SCHEMA = DATABASE()",
		"PREPARE stmt FROM @sql",
		"DEALLOCATE PREPARE stmt",
		"INDEX_NAME = 'idx_exam_assessment_publish'",
		"CREATE INDEX `idx_exam_assessment_publish` ON `el_exam` (`assessment_type`, `publish_status`)",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("migration missing required MySQL 5.7 idempotency fragment %q", required)
		}
	}
	for _, forbidden := range []string{"ADD COLUMN IF NOT EXISTS", "CREATE INDEX IF NOT EXISTS", "AutoMigrate", "DROP TABLE", "DROP COLUMN"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("migration contains forbidden fragment %q", forbidden)
		}
	}
}

func TestCompetencyDimensionMigration_DefinesMasterAndExamSelection(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	migrationPath := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_002_dimensions.sql"))
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration failed: %v", err)
	}
	sql := string(data)

	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS `el_competency_dimension`",
		"UNIQUE KEY `uk_competency_dimension_code` (`code`)",
		"UNIQUE KEY `uk_competency_dimension_name` (`name`)",
		"UNIQUE KEY `uk_competency_dimension_order` (`display_order`)",
		"KEY `idx_competency_dimension_status_order` (`status`,`display_order`)",
		"CREATE TABLE IF NOT EXISTS `el_exam_competency_dimension`",
		"UNIQUE KEY `uk_exam_competency_dimension` (`exam_id`,`dimension_id`)",
		"KEY `idx_exam_competency_dimension_order` (`exam_id`,`display_order`)",
		"SELECT CHARACTER_SET_NAME, COLLATION_NAME",
		"INTO @exam_id_charset, @exam_id_collation",
		"ALTER TABLE `el_exam_competency_dimension` MODIFY COLUMN `exam_id` varchar(64) CHARACTER SET ",
		"ON DUPLICATE KEY UPDATE",
		"'competency-d42','D42','权力动机'",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("dimension migration missing required fragment %q", required)
		}
	}
	if count := strings.Count(sql, "('competency-d"); count != 48 {
		t.Errorf("dimension migration seed row count = %d, want 48", count)
	}
	for _, forbidden := range []string{"DROP TABLE", "DELETE FROM", "TRUNCATE TABLE"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("dimension migration contains destructive fragment %q", forbidden)
		}
	}
}

func TestCompetencyDimensionModels_TableNames(t *testing.T) {
	if got := (CompetencyDimension{}).TableName(); got != "el_competency_dimension" {
		t.Fatalf("CompetencyDimension.TableName() = %q", got)
	}
	if got := (ExamCompetencyDimension{}).TableName(); got != "el_exam_competency_dimension" {
		t.Fatalf("ExamCompetencyDimension.TableName() = %q", got)
	}
}

func TestCompetencyQuestionMigration_DefinesMetadataAndIndexes(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	migrationPath := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_003_questions.sql"))
	data, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration failed: %v", err)
	}
	sql := string(data)

	for _, column := range []string{"question_code", "dimension_id", "dimension_item_no", "observation_point", "scoring_direction", "question_status"} {
		if !strings.Contains(sql, "COLUMN_NAME = '"+column+"'") {
			t.Errorf("question migration does not check column %s", column)
		}
		if !strings.Contains(sql, "ADD COLUMN `"+column+"`") {
			t.Errorf("question migration does not add column %s", column)
		}
	}
	for _, required := range []string{
		"INDEX_NAME = 'uk_qu_question_code'",
		"CREATE UNIQUE INDEX `uk_qu_question_code` ON `el_qu` (`question_code`)",
		"INDEX_NAME = 'uk_qu_dimension_item'",
		"CREATE UNIQUE INDEX `uk_qu_dimension_item` ON `el_qu` (`dimension_id`, `dimension_item_no`)",
		"INDEX_NAME = 'idx_qu_dimension_status'",
		"CREATE INDEX `idx_qu_dimension_status` ON `el_qu` (`dimension_id`, `question_status`)",
		"INTO @dimension_id_charset, @dimension_id_collation",
		"MODIFY COLUMN `dimension_id` varchar(64) CHARACTER SET ",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("question migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"DROP TABLE", "DROP COLUMN", "DELETE FROM", "TRUNCATE TABLE", "AutoMigrate"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("question migration contains destructive fragment %q", forbidden)
		}
	}
}

func TestQu_CompetencyMetadataFields(t *testing.T) {
	expected := []struct {
		name, gormTag, jsonTag string
	}{
		{"QuestionCode", "column:question_code", "questionCode"},
		{"DimensionID", "column:dimension_id", "dimensionId"},
		{"DimensionItemNo", "column:dimension_item_no", "dimensionItemNo"},
		{"ObservationPoint", "column:observation_point", "observationPoint"},
		{"ScoringDirection", "column:scoring_direction", "scoringDirection"},
		{"QuestionStatus", "column:question_status", "questionStatus"},
	}
	typ := reflect.TypeOf(Qu{})
	for _, want := range expected {
		field, ok := typ.FieldByName(want.name)
		if !ok {
			t.Errorf("Qu missing field %s", want.name)
			continue
		}
		if field.Tag.Get("gorm") != want.gormTag || field.Tag.Get("json") != want.jsonTag {
			t.Errorf("Qu.%s tags gorm=%q json=%q", want.name, field.Tag.Get("gorm"), field.Tag.Get("json"))
		}
	}
}

func TestCompetencyRuntimeMigration_DefinesSnapshotsAnswersAndResults(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_004_runtime.sql"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration failed: %v", err)
	}
	sql := string(data)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS `el_exam_competency_question`",
		"uk_exam_competency_question_source",
		"uk_exam_competency_question_code",
		"idx_exam_competency_question_dimension",
		"ADD COLUMN `exam_question_id`",
		"ADD COLUMN `raw_answer`",
		"ADD COLUMN `final_score`",
		"uk_paper_exam_question",
		"idx_paper_question_answered",
		"idx_paper_state_limit_time",
		"CREATE TABLE IF NOT EXISTS `el_competency_dimension_result`",
		"uk_competency_dimension_result",
		"idx_competency_dimension_score",
		"CREATE TABLE IF NOT EXISTS `el_competency_result`",
		"idx_competency_result_exam_score",
		"SELECT CHARACTER_SET_NAME, COLLATION_NAME",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("runtime migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"DROP TABLE", "DROP COLUMN", "TRUNCATE TABLE", "AutoMigrate"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("runtime migration contains destructive fragment %q", forbidden)
		}
	}
}

func TestCompetencyRuntimeModels_TableNames(t *testing.T) {
	tests := []struct {
		got, want string
	}{
		{(ExamCompetencyQuestion{}).TableName(), "el_exam_competency_question"},
		{(CompetencyDimensionResult{}).TableName(), "el_competency_dimension_result"},
		{(CompetencyResult{}).TableName(), "el_competency_result"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("TableName()=%q want=%q", tt.got, tt.want)
		}
	}
}

func TestPaperQu_CompetencyAnswerFields(t *testing.T) {
	typ := reflect.TypeOf(PaperQu{})
	for _, name := range []string{"ExamQuestionID", "RawAnswer", "FinalScore"} {
		if _, ok := typ.FieldByName(name); !ok {
			t.Errorf("PaperQu missing %s", name)
		}
	}
}

func TestCompetencyHardeningMigration_DefinesSnapshotsIndexesAndConstraints(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_005_hardening.sql"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, required := range []string{
		"participant_type", "participant_id", "participant_name", "participant_telephone",
		"uk_exam_competency_question_order", "idx_competency_result_exam_submitted",
		"fk_ecdim_exam", "fk_ecdim_dimension", "fk_ecq_exam", "fk_ecq_exam_dimension", "fk_ecq_source_qu",
		"fk_pq_exam_question", "fk_cdr_paper", "fk_cdr_exam_dimension", "fk_cdr_dimension", "fk_cr_paper", "fk_cr_exam",
		"information_schema.REFERENTIAL_CONSTRAINTS", "information_schema.STATISTICS", "information_schema.COLUMNS",
		"@participant_cs", "@participant_co", "MODIFY `participant_id`",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("hardening migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"DROP TABLE", "DROP COLUMN", "TRUNCATE TABLE", "AutoMigrate"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("hardening migration contains destructive fragment %q", forbidden)
		}
	}
}

func TestCompetencyResult_UsesExactDecimalAndIdentitySnapshots(t *testing.T) {
	resultType := reflect.TypeOf(CompetencyResult{})
	for _, name := range []string{"ParticipantType", "ParticipantID", "ParticipantName", "ParticipantTelephone", "ParticipantAge", "ParticipantGender", "ParticipantAffiliation", "ParticipantPost", "ParticipantDegree", "ParticipantMajor"} {
		if _, ok := resultType.FieldByName(name); !ok {
			t.Errorf("CompetencyResult missing %s", name)
		}
	}
	decimalType := reflect.TypeOf(decimal.Decimal{})
	if field, _ := resultType.FieldByName("OverallScore"); field.Type != decimalType {
		t.Errorf("OverallScore type=%v want decimal.Decimal", field.Type)
	}
	if field, _ := resultType.FieldByName("EvaluationAverage"); field.Type != reflect.PointerTo(decimalType) {
		t.Errorf("EvaluationAverage type=%v want *decimal.Decimal", field.Type)
	}
	dimensionType := reflect.TypeOf(CompetencyDimensionResult{})
	if field, _ := dimensionType.FieldByName("DimensionScore"); field.Type != reflect.PointerTo(decimalType) {
		t.Errorf("DimensionScore type=%v want *decimal.Decimal", field.Type)
	}
}

func TestCompetencyReportMigration_DefinesVersionedTextInstanceAndAudit(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_006_reports.sql"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, required := range []string{
		"el_competency_report_text", "el_competency_report", "el_competency_report_audit",
		"content_version", "text_snapshot", "score_snapshot", "pdf_sha256", "error_message",
		"uk_competency_report_text_match", "uk_competency_report_paper_version",
		"fk_competency_report_paper", "fk_competency_report_audit_report",
		"temp-v1", "临时测试文案", "information_schema.REFERENTIAL_CONSTRAINTS",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("report migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"DROP TABLE", "DROP COLUMN", "TRUNCATE TABLE", "AutoMigrate"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("report migration contains destructive fragment %q", forbidden)
		}
	}
}

func TestCompetencyReportModels_TableNames(t *testing.T) {
	for got, want := range map[string]string{
		(CompetencyReportText{}).TableName():  "el_competency_report_text",
		(CompetencyReport{}).TableName():      "el_competency_report",
		(CompetencyReportAudit{}).TableName(): "el_competency_report_audit",
	} {
		if got != want {
			t.Errorf("table name=%q want=%q", got, want)
		}
	}
}
