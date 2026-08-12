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
		{"CompetencyProductVersion", "column:competency_product_version", "competencyProductVersion"},
		{"CompetencyScoringVersion", "column:competency_scoring_version", "competencyScoringVersion"},
		{"CompetencyContentVersion", "column:competency_content_version", "competencyContentVersion"},
		{"CompetencyReportTemplateVersion", "column:competency_report_template_version", "competencyReportTemplateVersion"},
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
		"'competency-a1-01','A1-01','逻辑思维','通用能力','基层员工','逻辑分析严谨，推理判断有据',1",
		"'competency-b1-05','B1-05','合作意识','心理素养','基层员工','主动协作，乐于分享，促成共赢',10",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("dimension migration missing required fragment %q", required)
		}
	}
	if count := strings.Count(sql, "('competency-a1-") + strings.Count(sql, "('competency-b1-"); count != 10 {
		t.Errorf("dimension migration phase-1 seed row count = %d, want 10", count)
	}
	if strings.Contains(sql, "('competency-d") {
		t.Error("fresh dimension migration must not seed retired D identities")
	}
	for _, forbidden := range []string{"DROP TABLE", "DELETE FROM", "TRUNCATE TABLE"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("dimension migration contains destructive fragment %q", forbidden)
		}
	}
}

func TestCompetencyPhase1IdentityResetMigration_IsGuardedAndIdempotent(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_009_phase1_identity_reset.sql"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read phase-1 identity reset migration failed: %v", err)
	}
	sql := string(data)

	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS `el_competency_migration`",
		"competency-009-phase1-identity-reset",
		"@competency_009_target_environment",
		"@competency_009_writes_quiesced",
		"GET_LOCK(",
		"RELEASE_LOCK(",
		"COMPETENCY_009_REQUIRES_EXPLICIT_STAGING_AUTHORIZATION",
		"COMPETENCY_009_REMAINING_DEPENDENCIES",
		"FROM `el_paper_qu_answer` pqa",
		"FROM `el_user_book` ub",
		"START TRANSACTION",
		"INSERT IGNORE INTO `el_competency_migration`",
		"SET @apply_reset=ROW_COUNT()",
		"DELETE qa FROM `el_qu_answer` qa",
		"DELETE qr FROM `el_qu_repo` qr",
		"DELETE q FROM `el_qu` q",
		"q.`id` IS NOT NULL",
		"q.`dimension_id` IS NOT NULL OR q.`competency_question_type` IS NOT NULL",
		"DELETE rt FROM `el_competency_report_text` rt",
		"rt.`id` IS NOT NULL",
		"DELETE d FROM `el_competency_dimension` d",
		"d.`id` IS NOT NULL",
		"WHERE @apply_reset=1",
		"COMMIT",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("phase-1 identity reset migration missing %q", required)
		}
	}
	for _, identity := range []string{
		"'competency-a1-01','A1-01','逻辑思维','通用能力','基层员工','逻辑分析严谨，推理判断有据',1",
		"'competency-a1-02','A1-02','数字应用','通用能力','基层员工','善用数字化工具与AI技术，具备数据思维',2",
		"'competency-a1-03','A1-03','计划执行','通用能力','基层员工','高效推进计划并达成预期结果',3",
		"'competency-a1-04','A1-04','持续学习','通用能力','基层员工','主动学习，多渠道获取知识并学以致用',4",
		"'competency-a1-05','A1-05','沟通表达','通用能力','基层员工','清晰传递信息，重视倾听与反馈',5",
		"'competency-b1-01','B1-01','敬业奉献','心理素养','基层员工','视工作为使命，全心投入，甘于奉献',6",
		"'competency-b1-02','B1-02','求真务实','心理素养','基层员工','追求真理，尊重事实，注重实效',7",
		"'competency-b1-03','B1-03','自律性','心理素养','基层员工','自我约束，规划在先，言行一致',8",
		"'competency-b1-04','B1-04','成就导向','心理素养','基层员工','追求工作成功，不断挑战更高目标',9",
		"'competency-b1-05','B1-05','合作意识','心理素养','基层员工','主动协作，乐于分享，促成共赢',10",
	} {
		if !strings.Contains(sql, identity) {
			t.Errorf("phase-1 identity reset migration missing identity %q", identity)
		}
	}
	for _, forbidden := range []string{
		"DELETE FROM `el_exam`", "DELETE e FROM `el_exam`", "DELETE FROM `el_paper`",
		"DELETE FROM `el_candidate`", "DELETE FROM `el_tester`", "DROP TABLE", "TRUNCATE TABLE",
		"DELIMITER",
		"'competency-d", "'D01'", "'D48'", "34个维度", "40个维度",
	} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("phase-1 identity reset migration contains unsafe or unresolved fragment %q", forbidden)
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
	if field, _ := resultType.FieldByName("OverallScore"); field.Type != reflect.PointerTo(decimalType) {
		t.Errorf("OverallScore type=%v want *decimal.Decimal", field.Type)
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

func TestCompetencyPhase1ReportApprovalMigration_DefinesSafePackageGate(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_010_phase1_report_framework.sql"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS `el_competency_report_content_package`",
		"content_approved_by", "content_approved_at", "psychometric_approved_by", "psychometric_approved_at",
		"question_source_sha256", "content_source_sha256", "effective_environment", "approval_status", "disclaimer",
		"uk_competency_report_content_package", "idx_competency_report_content_package_status",
		"information_schema.REFERENTIAL_CONSTRAINTS", "PREPARE stmt FROM @sql", "DEALLOCATE PREPARE stmt",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("phase-1 report migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"DROP TABLE", "TRUNCATE TABLE", "AutoMigrate", "INSERT INTO `el_competency_report_content_package`"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("phase-1 report migration contains %q", forbidden)
		}
	}
}

func TestCompetencyVersionMigration_DefinesFrozenVersionFields(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_007_versions.sql"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, required := range []string{
		"competency_product_version", "competency_scoring_version", "competency_content_version", "competency_report_template_version",
		"product_version", "scoring_version", "content_version", "report_template_version", "template_version",
		"competency-generic-v1", "competency-v1", "temp-v1", "competency-report-v1",
		"information_schema.COLUMNS", "TABLE_SCHEMA=DATABASE()", "PREPARE stmt FROM @sql", "DEALLOCATE PREPARE stmt",
		"INDEX_NAME='uk_competency_report_paper_version'",
		"GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX)",
		"DROP INDEX `uk_competency_report_paper_version`",
		"CREATE UNIQUE INDEX `uk_competency_report_paper_version` ON `el_competency_report` (`paper_id`,`content_version`,`template_version`)",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("version migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"DROP TABLE", "DROP COLUMN", "TRUNCATE TABLE", "AutoMigrate", "ADD COLUMN IF NOT EXISTS"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("version migration contains forbidden fragment %q", forbidden)
		}
	}
}

func TestCompetencyVersionModels_FreezeResultAndReportVersions(t *testing.T) {
	resultType := reflect.TypeOf(CompetencyResult{})
	for _, name := range []string{"ProductVersion", "ScoringVersion", "ContentVersion", "ReportTemplateVersion"} {
		if _, ok := resultType.FieldByName(name); !ok {
			t.Errorf("CompetencyResult missing %s", name)
		}
	}
	reportType := reflect.TypeOf(CompetencyReport{})
	if _, ok := reportType.FieldByName("TemplateVersion"); !ok {
		t.Error("CompetencyReport missing TemplateVersion")
	}
}

func TestCompetencyPhase1StructureMigration_DefinesQuestionTypeValidityAndGroups(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "sql", "competency_008_phase1_structures.sql"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, required := range []string{
		"COLUMN_NAME='competency_question_type'",
		"ADD COLUMN `competency_question_type` varchar(32) DEFAULT NULL",
		"CREATE TABLE IF NOT EXISTS `el_exam_competency_group`",
		"ADD COLUMN `group_id` varchar(64) DEFAULT NULL",
		"CREATE TABLE IF NOT EXISTS `el_competency_group_result`",
		"CREATE TABLE IF NOT EXISTS `el_competency_validity_result`",
		"ADD COLUMN `dimension_question_count` int NOT NULL DEFAULT 0",
		"ADD COLUMN `answered_dimension_question_count` int NOT NULL DEFAULT 0",
		"UPDATE `el_competency_result`",
		"`competency_question_type`='dimension'",
		"WHERE `id` IS NOT NULL",
		"WHERE `paper_id` IS NOT NULL",
		"`product_version`='competency-generic-v1'",
		"`scoring_version`='competency-v1'",
		"uk_qu_dimension_item_v2_tmp",
		"RENAME INDEX `uk_qu_dimension_item_v2_tmp` TO `uk_qu_dimension_item`",
		"uk_exam_competency_group_code",
		"idx_exam_competency_dimension_group",
		"uk_competency_group_result",
		"fk_ecg_exam", "fk_ecdim_group", "fk_cgr_paper", "fk_cgr_exam_group", "fk_cvr_paper",
		"information_schema.COLUMNS", "information_schema.STATISTICS", "information_schema.REFERENTIAL_CONSTRAINTS",
		"TABLE_SCHEMA=DATABASE()", "PREPARE stmt FROM @sql", "DEALLOCATE PREPARE stmt",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("phase-1 structure migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"DROP TABLE", "DROP COLUMN", "TRUNCATE TABLE", "AutoMigrate", "ADD COLUMN IF NOT EXISTS", "CHECK ("} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("phase-1 structure migration contains forbidden fragment %q", forbidden)
		}
	}
}

func TestCompetencyPhase1Models_PreserveNullableUnresolvedValues(t *testing.T) {
	stringPointerType := reflect.TypeOf((*string)(nil))
	for _, target := range []struct {
		name      string
		typeOf    reflect.Type
		fieldName string
		gormTag   string
		jsonTag   string
	}{
		{"source question type", reflect.TypeOf(Qu{}), "CompetencyQuestionType", "column:competency_question_type", "competencyQuestionType"},
		{"snapshot question type", reflect.TypeOf(ExamCompetencyQuestion{}), "CompetencyQuestionType", "column:competency_question_type", "competencyQuestionType"},
	} {
		field, ok := target.typeOf.FieldByName(target.fieldName)
		if !ok {
			t.Errorf("%s missing field %s", target.name, target.fieldName)
			continue
		}
		if field.Type != stringPointerType {
			t.Errorf("%s type=%v want *string", target.name, field.Type)
		}
		if field.Tag.Get("gorm") != target.gormTag || field.Tag.Get("json") != target.jsonTag {
			t.Errorf("%s tags gorm=%q json=%q", target.name, field.Tag.Get("gorm"), field.Tag.Get("json"))
		}
	}

	resultType := reflect.TypeOf(CompetencyResult{})
	for _, name := range []string{"DimensionQuestionCount", "AnsweredDimensionQuestionCount"} {
		if _, ok := resultType.FieldByName(name); !ok {
			t.Errorf("CompetencyResult missing %s", name)
		}
	}
	groupIDField, ok := reflect.TypeOf(ExamCompetencyDimension{}).FieldByName("GroupID")
	if !ok || groupIDField.Type != stringPointerType {
		t.Error("ExamCompetencyDimension.GroupID must be nullable")
	}
	groupResultType := reflect.TypeOf(CompetencyGroupResult{})
	for _, name := range []string{"GroupScore", "LevelCode"} {
		field, ok := groupResultType.FieldByName(name)
		if !ok || field.Type.Kind() != reflect.Pointer {
			t.Errorf("CompetencyGroupResult.%s must be nullable", name)
		}
	}
	validityType := reflect.TypeOf(CompetencyValidityResult{})
	for _, name := range []string{"ValidityScore", "ValidityStatus"} {
		field, ok := validityType.FieldByName(name)
		if !ok || field.Type.Kind() != reflect.Pointer {
			t.Errorf("CompetencyValidityResult.%s must be nullable", name)
		}
	}
}

func TestCompetencyPhase1Models_TableNames(t *testing.T) {
	for got, want := range map[string]string{
		(ExamCompetencyGroup{}).TableName():      "el_exam_competency_group",
		(CompetencyGroupResult{}).TableName():    "el_competency_group_result",
		(CompetencyValidityResult{}).TableName(): "el_competency_validity_result",
	} {
		if got != want {
			t.Errorf("TableName()=%q want=%q", got, want)
		}
	}
}
