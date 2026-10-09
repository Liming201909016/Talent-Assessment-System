package service

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func readV2ResultRunSource(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestBugFB184_V2ResultRunRecordsUseExactScores(t *testing.T) {
	dimensionInputs := phase1V2InputsFromV1Sums([]int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	validityInputs := make([]Phase1ValidityInput, 0, 10)
	for index := 1; index <= 10; index++ {
		validityInputs = append(validityInputs, Phase1ValidityInput{QuestionCode: "V" + string(rune('A'+index-1)), Order: index, QuestionType: CompetencyQuestionTypeValidity, Direction: CompetencyDirectionForward, Answered: true, RawValue: 3})
	}
	now := time.Date(2026, 9, 18, 15, 0, 0, 0, time.Local)
	legacy := model.CompetencyResult{
		PaperID: "paper-1", ExamID: "exam-1", ParticipantType: CompetencyParticipantCandidate,
		ParticipantID: "participant-1", ParticipantName: "张三", ParticipantTelephone: "13800000000",
		ReportAudience: CompetencyReportAudienceFrontlineEmployee, IsComplete: 1, SubmitType: CompetencySubmitManual,
		ProductVersion: CompetencyPhase1ProductVersion, ScoringVersion: CompetencyPhase1ScoringVersion,
		ContentVersion: CompetencyPhase1ContentVersion, ReportTemplateVersion: CompetencyPhase1ReportTemplateVersion,
		SubmittedAt: &now,
	}
	records, err := buildPhase1V2ResultRunRecords("run-v2-1", legacy, dimensionInputs, validityInputs, 20, "historical_recompute", nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if records.Run.ID != "run-v2-1" || records.Run.PaperID != "paper-1" || records.Run.Source != "historical_recompute" || records.Run.Status != "completed" ||
		records.Run.ProductVersion != CompetencyPhase1ProductVersionV2 || records.Run.ScoringVersion != CompetencyPhase1ScoringVersionV2 {
		t.Fatalf("run=%+v", records.Run)
	}
	if records.Overall.OverallScore == nil || records.Overall.OverallScore.String() != "68.125" || len(records.Modules) != 3 || len(records.Dimensions) != 10 ||
		records.Modules[1].ModuleScore.String() != "59.375" || records.Dimensions[7].DimensionScore.String() != "78.125" {
		t.Fatalf("overall/modules/dimensions=%+v/%+v/%+v", records.Overall, records.Modules, records.Dimensions)
	}
	if records.Validity.ValidityScore == nil || records.Validity.ValidityScore.String() != "30" || records.Validity.ValidityStatus == nil || *records.Validity.ValidityStatus != CompetencyPhase1ValidityGood {
		t.Fatalf("validity=%+v", records.Validity)
	}
}

func TestBugFB184_ResultRunWriteIsAtomicIdempotentAndSubmitWired(t *testing.T) {
	source := readV2ResultRunSource(t, "competency_v2_result_run.go")
	for _, required := range []string{
		"func (s *CompetencyRuntimeService) RecomputePhase1V2ResultRun(", "Transaction(", `clause.Locking{Strength: "UPDATE"}`,
		`paper_id = ? AND scoring_version = ?`, "CreateInBatches", "CompetencyResultRunOverall", "CompetencyResultRunModule",
		"CompetencyResultRunDimension", "CompetencyResultRunValidity", "phase1V2ResultRunSchemaState",
		"CalculateCompetencyQuestionScore", "validity.ValidityScore.Equal",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("v2 result-run persistence missing %q", required)
		}
	}
	schemaSource := readV2ResultRunSource(t, "competency_v2_schema.go")
	for _, required := range []string{"information_schema.COLUMNS", "information_schema.STATISTICS", "information_schema.REFERENTIAL_CONSTRAINTS", "validatePhase1V2SchemaSignature"} {
		if !strings.Contains(schemaSource, required) {
			t.Errorf("v2 result-run schema validation missing %q", required)
		}
	}
	for _, forbidden := range []string{"Save(&", "DELETE FROM el_competency_result_run", "Updates(map"} {
		if strings.Contains(source, forbidden) {
			t.Errorf("v2 immutable result-run writer contains %q", forbidden)
		}
	}
	submit := readV2ResultRunSource(t, "competency_runtime.go")
	if !strings.Contains(submit, "persistPhase1V2ResultRun") || !strings.Contains(submit, "phase1V2ResultRunSchemaState") {
		t.Fatal("new phase-1 submission is not wired to the atomic v2 result-run writer")
	}
}

// TestBugFB185A_V2ResultRunRejectsNonFrontlineAudience
// 对应：docs/regression-tests.md #FB-185A
// 复现：已完成v1一期结果的冻结报告受众不是基层员工。
// 期望：在创建不可变v2 run前失败，不能产生报告读取器永远拒绝的completed run。
func TestBugFB185A_V2ResultRunRejectsNonFrontlineAudience(t *testing.T) {
	dimensionInputs := phase1V2InputsFromV1Sums([]int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	validityInputs := make([]Phase1ValidityInput, 0, 10)
	for index := 1; index <= 10; index++ {
		validityInputs = append(validityInputs, Phase1ValidityInput{
			QuestionCode: "V" + string(rune('A'+index-1)), Order: index,
			QuestionType: CompetencyQuestionTypeValidity, Direction: CompetencyDirectionForward,
			Answered: true, RawValue: 3,
		})
	}
	now := time.Date(2026, 9, 19, 9, 0, 0, 0, time.Local)
	legacy := model.CompetencyResult{
		PaperID: "paper-1", ExamID: "exam-1", ParticipantType: CompetencyParticipantCandidate,
		ParticipantID: "participant-1", ParticipantName: "张三", ReportAudience: CompetencyReportAudienceLeader,
		IsComplete: 1, SubmitType: CompetencySubmitManual, ProductVersion: CompetencyPhase1ProductVersion,
		ScoringVersion: CompetencyPhase1ScoringVersion, ContentVersion: CompetencyPhase1ContentVersion,
		ReportTemplateVersion: CompetencyPhase1ReportTemplateVersion, SubmittedAt: &now,
	}
	if _, err := buildPhase1V2ResultRunRecords("run-v2-1", legacy, dimensionInputs, validityInputs, 20, phase1V2ResultRunSourceHistoricalRecompute, nil, now); err == nil {
		t.Fatal("non-frontline v1 result must not create a phase-1 v2 run")
	}
}

// TestBugFB185A_FrozenAnswerSnapshotsMustBelongToLockedExam
// 对应：docs/regression-tests.md #FB-185A
// 复现：paper_qu异常指向另一测评的冻结题目或维度快照。
// 期望：查询同时约束question和dimension的exam_id，零匹配后按90题完整性失败。
func TestBugFB185A_FrozenAnswerSnapshotsMustBelongToLockedExam(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT .* FROM el_paper_qu pq .*q.exam_id = \\?.*d.exam_id = \\?.*pq.paper_id = \\?.*ORDER BY pq.sort ASC").
		WithArgs("exam-1", "exam-1", "paper-1").
		WillReturnRows(sqlmock.NewRows([]string{"sort", "answered", "raw_answer", "final_score", "dimension_id", "dimension_code", "dimension_name", "display_order", "question_code", "question_type", "dimension_item_no", "direction"}))

	if _, _, err := loadPhase1V2AnswerInputs(db, "paper-1", "exam-1"); err == nil {
		t.Fatal("cross-exam or missing frozen snapshots must fail")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// TestBugFB185C_ChildInsertFailureRollsBackWholeResultRun
// 对应：docs/regression-tests.md #FB-185C
// 复现：run和overall已写入，module批量插入失败。
// 期望：真实writer所在事务执行ROLLBACK且绝不COMMIT，不留下部分结果。
func TestBugFB185C_ChildInsertFailureRollsBackWholeResultRun(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	legacy, dimensionInputs, validityInputs, now := phase1V2ResultRunFixture()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `el_competency_result_run` WHERE paper_id = \\? AND scoring_version = \\?.*FOR UPDATE").
		WithArgs(legacy.PaperID, CompetencyPhase1ScoringVersionV2, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO `el_competency_result_run`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `el_competency_result_run_overall`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `el_competency_result_run_module`").WillReturnError(errors.New("injected module insert failure"))
	mock.ExpectRollback()

	err = db.Transaction(func(tx *gorm.DB) error {
		_, _, writeErr := persistPhase1V2ResultRun(tx, legacy, dimensionInputs, validityInputs, 20, phase1V2ResultRunSourceHistoricalRecompute, nil, now)
		return writeErr
	})
	if err == nil || !strings.Contains(err.Error(), "injected module insert failure") {
		t.Fatalf("transaction error=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugFB185C_ValidExistingRunIsReusedWithoutWrites(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	legacy, dimensionInputs, validityInputs, now := phase1V2ResultRunFixture()
	expected, err := buildPhase1V2ResultRunRecords("run-v2-existing", legacy, dimensionInputs, validityInputs, 20, phase1V2ResultRunSourceSubmission, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `el_competency_result_run` WHERE paper_id = \\? AND scoring_version = \\?.*FOR UPDATE").
		WithArgs(legacy.PaperID, CompetencyPhase1ScoringVersionV2, 1).
		WillReturnRows(resultRunSQLMockRow(expected.Run))
	mock.ExpectQuery("SELECT .* FROM `el_competency_result_run_overall` WHERE result_run_id = \\?.*LIMIT \\?").
		WithArgs(expected.Run.ID, 1).WillReturnRows(resultRunOverallSQLMockRow(expected.Overall))
	mock.ExpectQuery("SELECT .* FROM `el_competency_result_run_module` WHERE result_run_id = \\?.*ORDER BY display_order ASC").
		WithArgs(expected.Run.ID).WillReturnRows(resultRunModuleSQLMockRows(expected.Modules))
	mock.ExpectQuery("SELECT .* FROM `el_competency_result_run_dimension` WHERE result_run_id = \\?.*ORDER BY display_order ASC").
		WithArgs(expected.Run.ID).WillReturnRows(resultRunDimensionSQLMockRows(expected.Dimensions))
	mock.ExpectQuery("SELECT .* FROM `el_competency_result_run_validity` WHERE result_run_id = \\?.*LIMIT \\?").
		WithArgs(expected.Run.ID, 1).WillReturnRows(resultRunValiditySQLMockRow(expected.Validity))
	mock.ExpectCommit()

	var reused bool
	err = db.Transaction(func(tx *gorm.DB) error {
		_, reused, err = persistPhase1V2ResultRun(tx, legacy, dimensionInputs, validityInputs, 20, phase1V2ResultRunSourceHistoricalRecompute, nil, now.Add(time.Minute))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reused {
		t.Fatal("valid existing run was not reused")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func resultRunSQLMockRow(row model.CompetencyResultRun) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "paper_id", "exam_id", "product_version", "scoring_version", "content_version", "report_template_version", "report_audience", "participant_type", "participant_id", "participant_name", "participant_telephone", "participant_age", "participant_gender", "participant_affiliation", "participant_post", "participant_degree", "participant_major", "source", "status", "error_message", "completed_at"}).
		AddRow(row.ID, row.PaperID, row.ExamID, row.ProductVersion, row.ScoringVersion, row.ContentVersion, row.ReportTemplateVersion, row.ReportAudience, row.ParticipantType, row.ParticipantID, row.ParticipantName, row.ParticipantTelephone, row.ParticipantAge, row.ParticipantGender, row.ParticipantAffiliation, row.ParticipantPost, row.ParticipantDegree, row.ParticipantMajor, row.Source, row.Status, row.ErrorMessage, row.CompletedAt)
}

func resultRunOverallSQLMockRow(row model.CompetencyResultRunOverall) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"result_run_id", "total_question_count", "answered_question_count", "dimension_question_count", "answered_dimension_question_count", "effective_dimension_count", "overall_score", "level_code", "norm_score", "norm_comparison_code", "is_complete", "submit_type", "submitted_at", "user_time"}).
		AddRow(row.ResultRunID, row.TotalQuestionCount, row.AnsweredQuestionCount, row.DimensionQuestionCount, row.AnsweredDimensionQuestionCount, row.EffectiveDimensionCount, row.OverallScore, row.LevelCode, row.NormScore, row.NormComparisonCode, row.IsComplete, row.SubmitType, row.SubmittedAt, row.UserTime)
}

func resultRunModuleSQLMockRows(rows []model.CompetencyResultRunModule) *sqlmock.Rows {
	result := sqlmock.NewRows([]string{"id", "result_run_id", "module_id", "module_code", "module_name", "display_order", "total_dimension_count", "effective_dimension_count", "module_score", "level_code", "norm_score", "norm_comparison_code", "is_complete"})
	for _, row := range rows {
		result.AddRow(row.ID, row.ResultRunID, row.ModuleID, row.ModuleCode, row.ModuleName, row.DisplayOrder, row.TotalDimensionCount, row.EffectiveDimensionCount, row.ModuleScore, row.LevelCode, row.NormScore, row.NormComparisonCode, row.IsComplete)
	}
	return result
}

func resultRunDimensionSQLMockRows(rows []model.CompetencyResultRunDimension) *sqlmock.Rows {
	result := sqlmock.NewRows([]string{"id", "result_run_id", "dimension_id", "dimension_code", "dimension_name", "display_order", "total_question_count", "answered_question_count", "score_sum", "dimension_score", "level_code", "norm_score", "is_complete"})
	for _, row := range rows {
		result.AddRow(row.ID, row.ResultRunID, row.DimensionID, row.DimensionCode, row.DimensionName, row.DisplayOrder, row.TotalQuestionCount, row.AnsweredQuestionCount, row.ScoreSum, row.DimensionScore, row.LevelCode, row.NormScore, row.IsComplete)
	}
	return result
}

func resultRunValiditySQLMockRow(row model.CompetencyResultRunValidity) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"result_run_id", "total_question_count", "answered_question_count", "validity_score", "validity_status", "is_complete"}).
		AddRow(row.ResultRunID, row.TotalQuestionCount, row.AnsweredQuestionCount, row.ValidityScore, row.ValidityStatus, row.IsComplete)
}

func phase1V2ResultRunFixture() (model.CompetencyResult, []CompetencyScoreInput, []Phase1ValidityInput, time.Time) {
	now := time.Date(2026, 9, 19, 11, 0, 0, 0, time.Local)
	legacy := model.CompetencyResult{
		PaperID: "paper-1", ExamID: "exam-1", ParticipantType: CompetencyParticipantCandidate,
		ParticipantID: "participant-1", ParticipantName: "张三", ReportAudience: CompetencyReportAudienceFrontlineEmployee,
		IsComplete: 1, SubmitType: CompetencySubmitManual, ProductVersion: CompetencyPhase1ProductVersion,
		ScoringVersion: CompetencyPhase1ScoringVersion, ContentVersion: CompetencyPhase1ContentVersion,
		ReportTemplateVersion: CompetencyPhase1ReportTemplateVersion, SubmittedAt: &now,
	}
	dimensionInputs := phase1V2InputsFromV1Sums([]int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	validityInputs := make([]Phase1ValidityInput, 0, 10)
	for index := 1; index <= 10; index++ {
		validityInputs = append(validityInputs, Phase1ValidityInput{
			QuestionCode: "V" + string(rune('A'+index-1)), Order: index,
			QuestionType: CompetencyQuestionTypeValidity, Direction: CompetencyDirectionForward,
			Answered: true, RawValue: 3,
		})
	}
	return legacy, dimensionInputs, validityInputs, now
}

// TestBugFB185D_ResultRunSchemaRequiresExactUniqueAndForeignKeySignatures
// 对应：docs/regression-tests.md #FB-185D
func TestBugFB185D_ResultRunSchemaRequiresExactUniqueAndForeignKeySignatures(t *testing.T) {
	signature := phase1V2ExpectedSchemaSignature()
	if err := validatePhase1V2SchemaSignature(signature); err != nil {
		t.Fatalf("expected schema rejected: %v", err)
	}
	missingOrder := signature
	missingOrder.UniqueIndexes = append([]phase1V2SchemaIndex(nil), signature.UniqueIndexes...)
	missingOrder.UniqueIndexes = missingOrder.UniqueIndexes[:len(missingOrder.UniqueIndexes)-1]
	if err := validatePhase1V2SchemaSignature(missingOrder); err == nil {
		t.Fatal("schema missing an order uniqueness signature was accepted")
	}
	wrongForeignKey := signature
	wrongForeignKey.ForeignKeys = append([]phase1V2SchemaForeignKey(nil), signature.ForeignKeys...)
	wrongForeignKey.ForeignKeys[0].ReferencedTable = "wrong_table"
	if err := validatePhase1V2SchemaSignature(wrongForeignKey); err == nil {
		t.Fatal("schema with a wrong foreign-key target was accepted")
	}
	wrongColumn := signature
	wrongColumn.Columns = append([]phase1V2SchemaColumn(nil), signature.Columns...)
	wrongColumn.Columns[0].DataType = "text"
	if err := validatePhase1V2SchemaSignature(wrongColumn); err == nil {
		t.Fatal("schema with a wrong column definition was accepted")
	}
	extraColumn := signature
	extraColumn.Columns = append(append([]phase1V2SchemaColumn(nil), signature.Columns...), phase1V2SchemaColumn{Table: phase1V2ResultRunTables[0], Name: "unexpected"})
	if err := validatePhase1V2SchemaSignature(extraColumn); err == nil {
		t.Fatal("schema with an extra scoped column was accepted")
	}
	for _, column := range signature.Columns {
		if column.Name == "display_order" && column.DefaultValue != "<NULL>" {
			t.Fatalf("migration-shaped display_order default=%q", column.DefaultValue)
		}
		if column.Table == "el_competency_result_run_overall" && column.Name == "user_time" && (column.DataType != "int" || column.DefaultValue != "0") {
			t.Fatalf("migration-shaped user_time=%+v", column)
		}
	}
	wrongOptionalIndex := signature
	wrongOptionalIndex.UniqueIndexes = append(append([]phase1V2SchemaIndex(nil), signature.UniqueIndexes...), phase1V2SchemaIndex{Table: "el_competency_result_run", Name: "uk_result_run_id_paper", Columns: "paper_id,id"})
	if err := validatePhase1V2SchemaSignature(wrongOptionalIndex); err == nil {
		t.Fatal("malformed optional 013 run index was accepted")
	}
}

// TestBugFB185D_ResultRunSchemaPreflightIsCachedOutsideSubmitTransactions
// 对应：docs/regression-tests.md #FB-185D
func TestBugFB185D_ResultRunSchemaPreflightIsCachedOutsideSubmitTransactions(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM information_schema.TABLES").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	service := NewCompetencyRuntimeService(db, nil)
	for attempt := 0; attempt < 2; attempt++ {
		ready, err := service.phase1V2ResultRunSchemaState()
		if err != nil || ready {
			t.Fatalf("attempt %d ready=%v err=%v", attempt, ready, err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	submitSource := readV2ResultRunSource(t, "competency_runtime.go")
	submitAt := strings.Index(submitSource, "func (s *CompetencyRuntimeService) Submit(")
	if submitAt < 0 {
		t.Fatal("Submit function is missing")
	}
	submitSource = submitSource[submitAt:]
	transactionAt := strings.Index(submitSource, "s.db.Transaction")
	preflightAt := strings.Index(submitSource, "s.phase1V2ResultRunSchemaState()")
	if preflightAt < 0 || transactionAt < 0 || preflightAt > transactionAt {
		t.Fatal("result-run schema preflight is not completed before the submit transaction")
	}
}

// TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun
// 对应：docs/regression-tests.md #FB-185I
// 真实MySQL门禁：设置FB185_MYSQL_DSN后创建并清理独立临时数据库。
func TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FB185_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("FB185_MYSQL_DSN is not configured")
	}
	adminConfig, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	adminConfig.DBName = ""
	adminConfig.MultiStatements = true
	databaseName := strings.TrimSpace(os.Getenv("FB185_MYSQL_DATABASE"))
	if databaseName == "" {
		adminDB, err := sql.Open("mysql", adminConfig.FormatDSN())
		if err != nil {
			t.Fatal(err)
		}
		defer adminDB.Close()
		databaseName = fmt.Sprintf("fb185_%d", time.Now().UnixNano())
		if _, err := adminDB.Exec("CREATE DATABASE `" + databaseName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"); err != nil {
			t.Fatal(err)
		}
		defer adminDB.Exec("DROP DATABASE IF EXISTS `" + databaseName + "`")
	} else if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(databaseName) {
		t.Fatal("FB185_MYSQL_DATABASE contains unsafe characters")
	}
	testConfig := adminConfig
	testConfig.DBName = databaseName
	testSQL, err := sql.Open("mysql", testConfig.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer testSQL.Close()
	testSQL.SetMaxOpenConns(2)
	for _, statement := range []string{
		"CREATE TABLE el_paper (id varchar(64) NOT NULL PRIMARY KEY) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci",
		"CREATE TABLE el_exam (id varchar(64) NOT NULL PRIMARY KEY) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci",
		"CREATE TABLE el_candidate (id varchar(64) NOT NULL PRIMARY KEY) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci",
	} {
		if _, err := testSQL.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	migrationPath := filepath.Clean(filepath.Join("..", "..", "..", "scripts", "sql", "competency_011_result_runs.sql"))
	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testSQL.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}
	legacy, dimensionInputs, validityInputs, now := phase1V2ResultRunFixture()
	if _, err := testSQL.Exec("INSERT INTO el_exam(id) VALUES (?)", legacy.ExamID); err != nil {
		t.Fatal(err)
	}
	if _, err := testSQL.Exec("INSERT INTO el_paper(id) VALUES (?)", legacy.PaperID); err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: testSQL, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		reused bool
		err    error
	}
	start := make(chan struct{})
	outcomes := make(chan outcome, 2)
	for worker := 0; worker < 2; worker++ {
		go func(worker int) {
			<-start
			source := phase1V2ResultRunSourceSubmission
			if worker == 1 {
				source = phase1V2ResultRunSourceHistoricalRecompute
			}
			var reused bool
			err := db.Transaction(func(tx *gorm.DB) error {
				var lockedID string
				if err := tx.Raw("SELECT id FROM el_paper WHERE id = ? FOR UPDATE", legacy.PaperID).Scan(&lockedID).Error; err != nil {
					return err
				}
				_, reused, err = persistPhase1V2ResultRun(tx, legacy, dimensionInputs, validityInputs, 20, source, nil, now)
				return err
			})
			outcomes <- outcome{reused: reused, err: err}
		}(worker)
	}
	close(start)
	created, reused := 0, 0
	for worker := 0; worker < 2; worker++ {
		result := <-outcomes
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.reused {
			reused++
		} else {
			created++
		}
	}
	if created != 1 || reused != 1 {
		t.Fatalf("created=%d reused=%d", created, reused)
	}
	for table, want := range map[string]int64{
		"el_competency_result_run": 1, "el_competency_result_run_overall": 1,
		"el_competency_result_run_module": 3, "el_competency_result_run_dimension": 10,
		"el_competency_result_run_validity": 1,
	} {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Fatalf("%s count=%d want=%d", table, count, want)
		}
	}
}

// TestBugFB186_SchemaPreflightAvoidsMySQLReservedAliases
// 对应：docs/regression-tests.md #FB-186
// staging实证：`AS precision`在MySQL 8.0.46触发1064，导致全部重算失败。
func TestBugFB186_SchemaPreflightAvoidsMySQLReservedAliases(t *testing.T) {
	source := readV2ResultRunSource(t, "competency_v2_schema.go")
	for _, forbidden := range []string{" AS precision", " AS scale"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("schema preflight contains reserved alias %q", forbidden)
		}
	}
	for _, required := range []string{"numeric_precision_value", "numeric_scale_value"} {
		if !strings.Contains(source, required) {
			t.Fatalf("schema preflight missing safe alias %q", required)
		}
	}
}

// TestBugFB187_SchemaPreflightMapsInformationSchemaColumnsExplicitly
// 对应：docs/regression-tests.md #FB-187
func TestBugFB187_SchemaPreflightMapsInformationSchemaColumnsExplicitly(t *testing.T) {
	source := readV2ResultRunSource(t, "competency_v2_schema.go")
	for _, mapping := range []string{
		"Table        string `gorm:\"column:table_name\"`",
		"Name         string `gorm:\"column:name\"`",
		"DataType     string `gorm:\"column:data_type\"`",
		"Nullable     string `gorm:\"column:nullable\"`",
		"DefaultValue string `gorm:\"column:default_value\"`",
		"CharLength   int64  `gorm:\"column:char_length\"`",
		"CharacterSet string `gorm:\"column:character_set\"`",
		"Collation    string `gorm:\"column:collation\"`",
	} {
		if !strings.Contains(source, mapping) {
			t.Fatalf("schema preflight missing explicit projection mapping %q", mapping)
		}
	}
	for _, mapping := range []string{
		"Table   string `gorm:\"column:table_name\"`",
		"Columns string `gorm:\"column:columns\"`",
		"ReferencedTable string `gorm:\"column:referenced_table\"`",
		"Referenced      string `gorm:\"column:referenced\"`",
		"UpdateRule      string `gorm:\"column:update_rule\"`",
		"DeleteRule      string `gorm:\"column:delete_rule\"`",
	} {
		if !strings.Contains(source, mapping) {
			t.Fatalf("schema preflight missing explicit index/fk mapping %q", mapping)
		}
	}
}
