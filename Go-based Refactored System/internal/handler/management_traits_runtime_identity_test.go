package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func identityHandlerDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	return gdb, mock
}

func identityHandlerSchema(mock sqlmock.Sqlmock, present int) {
	for _, table := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_result_dimension", "el_mng_result_module"} {
		mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WithArgs(table).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(present))
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name"}))
}

// Metadata comes from the checked-in DDL, not the service's private contract
// or a pre-populated cache. The ordered expectations must precede BEGIN.
func identityHandlerRuntimeSchema(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "sql", "management_traits_001_runtime.sql"))
	if err != nil {
		t.Fatal(err)
	}
	tables := sqlmock.NewRows([]string{"table_name", "engine"})
	columns := sqlmock.NewRows([]string{"table_name", "column_name", "column_type", "is_nullable", "character_set_name", "collation_name"})
	indexes := sqlmock.NewRows([]string{"table_name", "index_name", "non_unique", "seq_in_index", "column_name", "prefix_length"})
	fks := sqlmock.NewRows([]string{"table_name", "constraint_name", "column_name", "ordinal_position", "referenced_table_name", "referenced_column_name", "update_rule", "delete_rule"})
	tableRE := regexp.MustCompile(`(?s)CREATE TABLE IF NOT EXISTS (el_mng_[a-z_]+) \((.*?)\) ENGINE=InnoDB`)
	columnRE := regexp.MustCompile("(?m)^`([a-z_]+)` ([a-z]+(?:\\([0-9,]+\\))?) (NOT NULL|NULL)")
	indexRE := regexp.MustCompile(`(PRIMARY KEY|UNIQUE KEY [a-z_]+|KEY [a-z_]+)\(([a-z_]+(?:\([0-9]+\))?(?:,[a-z_]+(?:\([0-9]+\))?)*)\)`)
	indexColumnRE := regexp.MustCompile(`^([a-z_]+)(?:\(([0-9]+)\))?$`)
	fkRE := regexp.MustCompile(`CONSTRAINT ([a-z_]+) FOREIGN KEY\(([a-z_,]+)\) REFERENCES ([a-z_]+)\(([a-z_,]+)\) ON DELETE (RESTRICT) ON UPDATE (RESTRICT)`)
	matches := tableRE.FindAllStringSubmatch(string(raw), -1)
	if len(matches) != 11 {
		t.Fatalf("DDL table count=%d, want 11", len(matches))
	}
	fkCount := 0
	for _, table := range matches {
		name, body := table[1], table[2]
		tables.AddRow(name, "InnoDB")
		cs := columnRE.FindAllStringSubmatch(body, -1)
		if len(cs) == 0 {
			t.Fatal("DDL columns missing", name)
		}
		for _, c := range cs {
			nullable, charset, collation := "NO", "", ""
			if c[3] == "NULL" {
				nullable = "YES"
			}
			if strings.HasPrefix(c[2], "varchar") || strings.HasPrefix(c[2], "char") || c[2] == "longtext" {
				charset, collation = "utf8mb4", "utf8mb4_bin"
			}
			columns.AddRow(name, c[1], c[2], nullable, charset, collation)
		}
		for _, index := range indexRE.FindAllStringSubmatch(body, -1) {
			key, nonUnique := "PRIMARY", 0
			if index[1] != "PRIMARY KEY" {
				parts := strings.Fields(index[1])
				key = parts[len(parts)-1]
				if parts[0] == "KEY" {
					nonUnique = 1
				}
			}
			for i, rawColumn := range strings.Split(index[2], ",") {
				c := indexColumnRE.FindStringSubmatch(rawColumn)
				if c == nil {
					t.Fatal("invalid DDL index column", rawColumn)
				}
				prefix := 0
				if c[2] != "" {
					prefix, err = strconv.Atoi(c[2])
					if err != nil {
						t.Fatal(err)
					}
				}
				indexes.AddRow(name, key, nonUnique, i+1, c[1], prefix)
			}
		}
		for _, fk := range fkRE.FindAllStringSubmatch(body, -1) {
			fkCount++
			local, remote := strings.Split(fk[2], ","), strings.Split(fk[4], ",")
			if len(local) != len(remote) {
				t.Fatal("invalid DDL FK arity", fk[1])
			}
			for i, c := range local {
				fks.AddRow(name, fk[1], c, i+1, fk[3], remote[i], fk[6], fk[5])
			}
		}
	}
	if fkCount != 15 {
		t.Fatalf("DDL FK count=%d, want 15", fkCount)
	}
	for _, parent := range []string{"el_exam", "el_paper", "el_paper_qu"} {
		columns.AddRow(parent, "id", "varchar(64)", "NO", "utf8mb4", "utf8mb4_bin")
	}
	for table, names := range map[string][]string{
		"el_exam_repo": {"exam_id", "repo_id"},
		"el_repo":      {"id"}, "el_qu_repo": {"id", "repo_id", "qu_id"},
		"el_qu": {"id"}, "el_qu_answer": {"id", "qu_id"},
		"el_candidate": {"id", "exam_id", "paper_id"},
		"el_tester":    {"id", "exam_id", "paper_id"},
		"el_paper":     {"exam_id", "user_id"}, "el_paper_qu": {"paper_id", "qu_id"},
		"el_paper_qu_answer": {"id", "paper_id", "qu_id", "answer_id"},
	} {
		for _, name := range names {
			nullable := "NO"
			if (table == "el_candidate" && name == "paper_id") || (table == "el_tester" && name != "id") {
				nullable = "YES"
			}
			columns.AddRow(table, name, "varchar(64)", nullable, "utf8mb4", "utf8mb4_bin")
		}
	}
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(tables)
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type").WillReturnRows(columns)
	mock.ExpectQuery("SELECT table_name AS table_name, index_name AS index_name, non_unique AS non_unique").WillReturnRows(indexes)
	mock.ExpectQuery("SELECT k.table_name AS table_name, k.constraint_name AS constraint_name").WillReturnRows(fks)
}

func identityHandlerModelRows(t *testing.T, values ...any) *sqlmock.Rows {
	t.Helper()
	s, err := schema.Parse(values[0], &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	rows := sqlmock.NewRows(s.DBNames)
	for _, value := range values {
		data := make([]driver.Value, len(s.DBNames))
		for i, name := range s.DBNames {
			x, _ := s.FieldsByDBName[name].ValueOf(context.Background(), reflect.ValueOf(value))
			data[i], err = driver.DefaultParameterConverter.ConvertValue(x)
			if err != nil {
				t.Fatal(err)
			}
		}
		rows.AddRow(data...)
	}
	return rows
}

func TestManagementTraitsIdentityHandlerRuntimeSchemaFixture(t *testing.T) {
	db, mock := identityHandlerDB(t)
	identityHandlerRuntimeSchema(t, mock)
	s := service.NewManagementTraitsRuntimeService(db, "test-only-secret", 1<<20)
	for i := 0; i < 2; i++ {
		if err := s.CheckRuntimeSchema(context.Background()); err != nil {
			t.Fatal("DDL metadata rejected", err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func identityHandlerPost(h gin.HandlerFunc, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/identity", h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/identity", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestManagementTraitsIdentityHTTPStrictCandidate(t *testing.T) {
	for _, extra := range []string{`,"paperId":"foreign"`, `,"endTime":null`, `,"pdfPath":"private"`, `,"password":"secret"`, `,"token":"secret"`, `,"idNumber":"legacy"`, `,"Name":"alias"`, `,"name":"duplicate"`, `,"na\u006de":"duplicate"`, `,"gender":null`} {
		t.Run(extra, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			identityHandlerSchema(mock, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("new-exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			w := identityHandlerPost(NewCandidateHandler(db, &config.Config{}).Save, `{"examId":"new-exam","name":"Person","telephone":"13800000000"`+extra+`}`)
			if !strings.Contains(w.Body.String(), "参数") {
				t.Fatalf("unsafe input: %s", w.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityHTTPProbeFailure(t *testing.T) {
	db, mock := identityHandlerDB(t)
	mock.ExpectQuery("information_schema.tables").WillReturnError(errors.New("PRIVATE_SCHEMA_SECRET"))
	w := identityHandlerPost(NewCandidateHandler(db, &config.Config{}).Save, `{"examId":"exam","name":"Person","telephone":"13800000000"}`)
	if strings.Contains(w.Body.String(), "PRIVATE") || !strings.Contains(w.Body.String(), "管理特质") {
		t.Fatal(w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityHTTPCandidateInvalidRawID(t *testing.T) {
	for _, id := range []string{`0`, `false`, `[]`, `{}`, `null`} {
		t.Run(id, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			if id == "null" {
				identityHandlerSchema(mock, 1)
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			}
			w := identityHandlerPost(NewCandidateHandler(db, &config.Config{}).Save, `{"examId":"exam","name":"Person","telephone":"13800000000","id":`+id+`}`)
			var result struct {
				Code int `json:"code"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code == 0 {
				t.Fatal("invalid ID allowed candidate creation", w.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityHTTPCandidateMissingExamNeverWrites(t *testing.T) {
	for _, body := range []string{
		`{"id":"protected-person","name":"Person","telephone":"13800000000"}`,
		`{"examId":"","id":"protected-person","name":"Person","telephone":"13800000000"}`,
		`{"examId":null,"id":"protected-person","name":"Person","telephone":"13800000000"}`,
	} {
		db, mock := identityHandlerDB(t)
		w := identityHandlerPost(NewCandidateHandler(db, &config.Config{}).Save, body)
		var result struct {
			Code int `json:"code"`
		}
		if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code == 0 || !strings.Contains(w.Body.String(), "examId") {
			t.Fatal("missing exam allowed legacy write", w.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagementTraitsIdentityHTTPTrailingJSONNeverFallsBack(t *testing.T) {
	db, mock := identityHandlerDB(t)
	w := identityHandlerPost(NewCandidateHandler(db, &config.Config{}).Save, `{"examId":"exam","name":"Person","telephone":"13800000000"} {}`)
	if !strings.Contains(w.Body.String(), "参数错误") {
		t.Fatal("invalid raw JSON reached legacy", w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityHTTPBodyPreservedForLegacy(t *testing.T) {
	db, mock := identityHandlerDB(t)
	identityHandlerSchema(mock, 0)
	identityHandlerSchema(mock, 0)
	h := NewCandidateHandler(db, &config.Config{})
	body := `{"examId":"legacy","name":"Person","telephone":"13800000000","age":"","unknown":"legacy-compatible"}`
	r := gin.New()
	r.POST("/identity", func(c *gin.Context) {
		if h.TryManagementTraitsCandidateSave(c) {
			t.Fatal("legacy handled")
		}
		var fields map[string]any
		if c.ShouldBindJSON(&fields) != nil || fields["unknown"] != "legacy-compatible" || fields["age"] != "" {
			t.Fatal("body consumed")
		}
		c.Status(204)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/identity", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Fatal(w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityHTTPLoginProbeBeforeWrite(t *testing.T) {
	db, mock := identityHandlerDB(t)
	mock.ExpectQuery("SELECT .*el_tester").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "name", "password", "telephone", "del_flag"}).AddRow("tester", "exam", "Person", "test-password", "13800000000", 0))
	mock.ExpectQuery("information_schema.tables").WillReturnError(errors.New("PRIVATE_DATABASE_PASSWORD"))
	w := identityHandlerPost(NewTesterHandler(db, &config.Config{}).LoginForm, `{"idNumber":"tester","password":"test-password","examId":"exam"}`)
	if strings.Contains(w.Body.String(), "PRIVATE") || strings.Contains(w.Body.String(), "test-password") || !strings.Contains(w.Body.String(), "管理特质") {
		t.Fatal(w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityHTTPTesterWrongPasswordBeforeProbe(t *testing.T) {
	db, mock := identityHandlerDB(t)
	mock.ExpectQuery("SELECT .*el_tester").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "password"}).AddRow("tester", "exam", "correct"))
	w := identityHandlerPost(NewTesterHandler(db, &config.Config{}).LoginForm, `{"idNumber":"tester","password":"wrong","examId":"exam"}`)
	if !strings.Contains(w.Body.String(), "密码错误") {
		t.Fatal(w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityHTTPResponseWhitelist(t *testing.T) {
	// Response type is shared by both real HTTP branches, not a legacy model.
	w := identityHandlerPost(func(c *gin.Context) {
		response.AjaxOK(c, service.ManagementTraitsIdentityResponse{ID: "id", ExamID: "exam", PaperID: "paper", Name: "Person", ParticipantToken: "test-token"})
	}, `{}`)
	var body struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &body) != nil || body.Code != 200 || len(body.Data) != 5 {
		t.Fatal(w.Body.String())
	}
	for _, field := range []string{"password", "token", "pdfPath", "paperSnapshot", "endTime"} {
		if _, ok := body.Data[field]; ok {
			t.Fatal("leak", field)
		}
	}
}

func TestManagementTraitsIdentityHTTPTesterBoundExam(t *testing.T) {
	// Explicit mismatch is rejected by the real service scope/admission path,
	// never by rebinding the legacy Tester model.
	db, mock := identityHandlerDB(t)
	mock.ExpectQuery("information_schema.tables").WillReturnError(errors.New("scope unavailable"))
	exam := "existing"
	r := gin.New()
	r.POST("/identity", func(c *gin.Context) {
		if !NewTesterHandler(db, &config.Config{}).TryManagementTraitsTesterLogin(c, "destination", &model.Tester{ID: "tester", ExamID: &exam}, "test-password") {
			t.Fatal("unsafe fallback")
		}
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/identity", nil))
	if !strings.Contains(w.Body.String(), "管理特质") {
		t.Fatal(w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func identityHandlerSource(t *testing.T) (model.ManagementTraitsDefinitionBundle, model.ManagementTraitsExamProfile, service.ManagementTraitsManifest, service.ManagementTraitsMapping) {
	t.Helper()
	rows := make([]service.ManagementTraitsRuntimeSourceRow, 0, 700)
	labels := []string{"", "不符合", "不太符合", "一般", "比较符合", "很符合"}
	for _, dimension := range service.ManagementTraitsDimensions() {
		for _, item := range dimension.Items {
			for raw := 1; raw <= 5; raw++ {
				right := 0
				if (!item.Reverse && raw == 5) || (item.Reverse && raw == 1) {
					right = 1
				}
				rows = append(rows, service.ManagementTraitsRuntimeSourceRow{RelationID: fmt.Sprint("r", item.Number), QuestionID: fmt.Sprint("q", item.Number), Sort: item.Number, RelationType: 1, QuestionType: 1, Code: fmt.Sprint("V", item.Number), Title: fmt.Sprint("synthetic-", item.Number), OptionID: fmt.Sprintf("o%d-%d", item.Number, raw), Raw: raw, IsRight: right, OptionContent: labels[raw]})
			}
		}
	}
	manifest, mapping, err := service.BuildManagementTraitsRuntimeCurrentSource("00201", rows)
	if err != nil {
		t.Fatal(err)
	}
	mc, err := service.CanonicalManagementTraitsManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	pc, err := service.CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil {
		t.Fatal(err)
	}
	frozen := time.Now().Add(-time.Hour).Truncate(time.Second)
	fields, err := json.Marshal(map[string]any{"schema": "mng-candidate-fields-v1", "requiredFields": []string{"name", "telephone"}, "timePolicy": "candidate-only-personal-25-minutes-v1", "source": "current-source-not-client-leader-full-not-historical", "repoCode": "00201", "capturedAt": frozen.Format(time.RFC3339Nano), "manifestSha": mc.SHA256, "mappingSha": pc.SHA256})
	if err != nil {
		t.Fatal(err)
	}
	bundle := model.ManagementTraitsDefinitionBundle{ID: "bundle", ProductVersion: manifest.Versions.Product, QuestionVersion: manifest.Versions.Question, ScoringVersion: manifest.Versions.Scoring, NormVersion: manifest.Versions.Norm, Questionnaire: manifest.Questionnaire, ScoringManifest: string(mc.JSON), ScoringManifestSHA: mc.SHA256, Status: "candidate-current-source", CreatedAt: frozen}
	profile := model.ManagementTraitsExamProfile{ExamID: "exam", BundleID: bundle.ID, MappingSnapshot: string(pc.JSON), MappingSHA: pc.SHA256, FieldContract: string(fields), TotalTimeMinutes: 25, FrozenAt: &frozen, CreatedAt: frozen}
	return bundle, profile, manifest, mapping
}

func identityHandlerAdmission(t *testing.T, mock sqlmock.Sqlmock, open int) {
	t.Helper()
	bundle, profile, _, _ := identityHandlerSource(t)
	mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "is_open", "state", "assessment_type", "scoring_mode"}).AddRow("exam", open, 0, "legacy", "legacy"))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(identityHandlerModelRows(t, profile))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(identityHandlerModelRows(t, bundle))
}

// Enqueue the complete second, read-only transaction after owner COMMIT.
// Fixed bindings: exam/paper, candidate existing or tester tester, Person.
func identityHandlerFrozenResume(t *testing.T, mock sqlmock.Sqlmock, kind string) {
	t.Helper()
	if kind != "candidate" && kind != "tester" {
		t.Fatal("invalid fixture participant kind", kind)
	}
	bundle, profile, manifest, mapping := identityHandlerSource(t)
	id := "existing"
	if kind == "tester" {
		id = "tester"
	}
	start := time.Now().Add(-5 * time.Minute).Truncate(time.Second)
	deadline := start.Add(25 * time.Minute)
	evidence := struct {
		Schema      string    `json:"schema"`
		Source      string    `json:"source"`
		RepoCode    string    `json:"repoCode"`
		CapturedAt  time.Time `json:"capturedAt"`
		ManifestSHA string    `json:"manifestSha"`
		MappingSHA  string    `json:"mappingSha"`
	}{"mng-current-source-evidence-v1", "current-source-not-client-leader-full-not-historical", "00201", *profile.FrozenAt, bundle.ScoringManifestSHA, profile.MappingSHA}
	evidenceJSON, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	evidenceSHA := fmt.Sprintf("%x", sha256.Sum256(evidenceJSON))
	snapshot := model.ManagementTraitsPaperSnapshot{PaperID: "paper", ExamID: "exam", ProfileExamID: &profile.ExamID, BundleID: bundle.ID, Source: "new_creation", EvidenceSnapshot: string(evidenceJSON), EvidenceSHA: evidenceSHA, MappingSnapshot: profile.MappingSnapshot, MappingSHA: profile.MappingSHA, ScoringManifestSHA: bundle.ScoringManifestSHA, ParticipantType: kind, ParticipantID: id, ParticipantSnapshot: `{"name":"Person","gender":"","telephone":"13800000000","affiliation":"","post":""}`, FieldContract: profile.FieldContract, IdentitySource: "submitted_snapshot", SourceCapturedAt: *profile.FrozenAt, StartedAt: start, LimitTime: &deadline, CreatedAt: start}
	questions, legacy, buckets := make([]any, 0, 140), make([]any, 0, 140), make([]any, 0, 700)
	for i, q := range mapping.Questions {
		mq := manifest.Questions[q.Number-1]
		pqID := fmt.Sprintf("pq%d", q.Number)
		options, err := json.Marshal(q.Options)
		if err != nil {
			t.Fatal(err)
		}
		frozen := struct {
			Schema          string                                 `json:"schema"`
			ManifestSHA     string                                 `json:"manifestSha"`
			MappingSHA      string                                 `json:"mappingSha"`
			Question        service.ManagementTraitsMappedQuestion `json:"question"`
			DimensionKey    string                                 `json:"dimensionKey"`
			Reverse         bool                                   `json:"reverse"`
			PaperQuestionID string                                 `json:"paperQuestionId"`
			DisplayOrder    int                                    `json:"displayOrder"`
		}{"mng-frozen-question-v1", bundle.ScoringManifestSHA, profile.MappingSHA, q, mq.DimensionKey, mq.Reverse, pqID, i + 1}
		frozenJSON, err := json.Marshal(frozen)
		if err != nil {
			t.Fatal(err)
		}
		questions = append(questions, model.ManagementTraitsPaperQuestionSnapshot{ID: fmt.Sprintf("snapshot%d", q.Number), PaperID: "paper", PaperQuestionID: pqID, SourceQuestionID: q.SourceQuestionID, Number: q.Number, DisplayOrder: i + 1, DimensionKey: mq.DimensionKey, Reverse: mq.Reverse, Content: q.Content, OptionsSnapshot: string(options), ScoringSnapshotSHA: fmt.Sprintf("%x", sha256.Sum256(frozenJSON)), CreatedAt: start})
		legacy = append(legacy, model.PaperQu{ID: pqID, PaperID: "paper", QuID: q.SourceQuestionID, QuType: 1, Sort: i + 1})
		for _, o := range q.Options {
			buckets = append(buckets, model.PaperQuAnswer{ID: "bucket-" + o.SourceOptionID, PaperID: "paper", QuID: q.SourceQuestionID, AnswerID: o.SourceOptionID, Sort: o.DisplayOrder, Score: o.Raw})
		}
	}
	qs := make([]model.ManagementTraitsPaperQuestionSnapshot, len(questions))
	for i, q := range questions {
		qs[i] = q.(model.ManagementTraitsPaperQuestionSnapshot)
	}
	if _, err := service.ValidateManagementTraitsStoredInput(bundle, snapshot, qs, 1<<20); err != nil {
		t.Fatal("invalid frozen fixture", err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnRows(identityHandlerModelRows(t, model.Paper{ID: "paper", ExamID: "exam", UserID: id, State: 1, TotalTime: 25, CreateTime: &start, LimitTime: &deadline}))
	mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(identityHandlerModelRows(t, snapshot))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(identityHandlerModelRows(t, bundle))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(identityHandlerModelRows(t, profile))
	mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WillReturnRows(sqlmock.NewRows([]string{"kind", "id", "exam_id", "paper_id", "name", "telephone", "del_flag", "status"}).AddRow(kind, id, "exam", "paper", "Person", "13800000000", 0, "0"))
	mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(identityHandlerModelRows(t, questions...))
	mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(identityHandlerModelRows(t, legacy...))
	mock.ExpectQuery("SELECT .*el_paper_qu_answer").WillReturnRows(identityHandlerModelRows(t, buckets...))
	mock.ExpectCommit()
}

func TestManagementTraitsIdentityHandlerFrozenResumeFixture(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		t.Run(kind, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			identityHandlerRuntimeSchema(t, mock)
			identityHandlerFrozenResume(t, mock, kind)
			id := "existing"
			if kind == "tester" {
				id = "tester"
			}
			claims := service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimePaperPurpose, ParticipantType: kind, ParticipantID: id, ExamID: "exam", PaperID: "paper", ExpiresAt: time.Now().Add(time.Hour).Unix()}
			got, err := service.NewManagementTraitsRuntimeService(db, "test-only-secret", 1<<20).PaperDetail(context.Background(), claims)
			if err != nil || got.PaperID != "paper" || got.ExamID != "exam" || got.Answered != 0 || len(got.Questions) != 140 {
				t.Fatal("full read-only frozen fixture rejected", err)
			}
			for _, q := range got.Questions {
				if len(q.Options) != 5 || q.SelectedOptionID != nil {
					t.Fatal("incomplete frozen question/options", q.ID)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityHTTPCandidateCreateRestore(t *testing.T) {
	for _, mode := range []string{"create", "anonymous_phone", "authenticated_restore"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			cfg := &config.Config{}
			cfg.Jwt.Secret = "test-only-secret"
			identityHandlerSchema(mock, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			identityHandlerSchema(mock, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			identityHandlerRuntimeSchema(t, mock)
			mock.ExpectBegin()
			identityHandlerAdmission(t, mock, 1)
			owners := sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "telephone", "del_flag"})
			if mode != "create" {
				owners.AddRow("existing", "exam", "paper", "Person", "13800000000", 0)
			}
			mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(owners)
			if mode == "create" {
				mock.ExpectExec("INSERT INTO `el_candidate` \\(`id`,`exam_id`,`name`,`gender`,`telephone`,`affiliation`,`post`,`age`,`degree`,`major`,`stu_flag`,`del_flag`,`create_time`,`update_time`\\)").WillReturnResult(sqlmock.NewResult(1, 1))
			}
			if mode == "anonymous_phone" {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}
			if mode == "authenticated_restore" {
				identityHandlerFrozenResume(t, mock, "candidate")
			}
			token := ""
			if mode == "authenticated_restore" {
				var err error
				token, err = service.CreateManagementTraitsRuntimeToken(cfg.Jwt.Secret, service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: "existing", ExamID: "exam", ExpiresAt: time.Now().Add(time.Hour).Unix()}, time.Now())
				if err != nil {
					t.Fatal(err)
				}
			}
			r := gin.New()
			r.POST("/identity", NewCandidateHandler(db, cfg).Save)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/identity", strings.NewReader(`{"examId":"exam","name":"Person","telephone":"13800000000"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Management-Traits-Token", token)
			r.ServeHTTP(w, req)
			var result struct {
				Code int                                      `json:"code"`
				Data service.ManagementTraitsIdentityResponse `json:"data"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil {
				t.Fatal(w.Body.String())
			}
			if mode == "anonymous_phone" {
				if result.Code == 0 || result.Data.ID != "" {
					t.Fatal("anonymous overwrite", w.Body.String())
				}
			} else {
				claims, err := service.ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, result.Data.ParticipantToken, service.ManagementTraitsRuntimeParticipantPurpose, time.Now())
				if result.Code != 0 || err != nil || claims.ValidateBinding("candidate", result.Data.ID, "exam", "") != nil {
					t.Fatal("invalid candidate issuance", result.Code, err)
				}
				if mode == "authenticated_restore" && (result.Data.ID != "existing" || result.Data.PaperID != "paper") {
					t.Fatal("binding changed")
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityHTTPTesterLoginZeroWrites(t *testing.T) {
	db, mock := identityHandlerDB(t)
	cfg := &config.Config{}
	cfg.Jwt.Secret = "test-only-secret"
	testerRows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "password", "telephone", "del_flag", "status"}).AddRow("tester", "exam", "paper", "Person", "test-password", "13800000000", 0, "0")
	}
	mock.ExpectQuery("SELECT .*el_tester").WillReturnRows(testerRows())
	identityHandlerSchema(mock, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .*el_tester.*id =").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("tester", "exam", "paper"))
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	identityHandlerFullParticipantScope(mock, "tester", "tester", false)
	identityHandlerRuntimeSchema(t, mock)
	mock.ExpectBegin()
	identityHandlerAdmission(t, mock, 2)
	mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WillReturnRows(testerRows())
	mock.ExpectCommit()
	identityHandlerFrozenResume(t, mock, "tester")
	w := identityHandlerPost(NewTesterHandler(db, cfg).LoginForm, `{"idNumber":"tester","password":"test-password","examId":"exam"}`)
	var result struct {
		Code int                                      `json:"code"`
		Data service.ManagementTraitsIdentityResponse `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code != 200 || result.Data.PaperID != "paper" {
		t.Fatal("login result rejected", result.Code)
	}
	claims, err := service.ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, result.Data.ParticipantToken, service.ManagementTraitsRuntimeParticipantPurpose, time.Now())
	if err != nil || claims.ValidateBinding("tester", "tester", "exam", "") != nil {
		t.Fatal("wrong token", err)
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal(w.Body.Bytes(), &raw) != nil {
		t.Fatal("invalid JSON")
	}
	var fields map[string]any
	if json.Unmarshal(raw["data"], &fields) != nil || len(fields) != 5 {
		t.Fatal("response leaked fields")
	}
	for _, forbidden := range []string{"password", "token", "paperSnapshot", "pdfPath"} {
		if _, found := fields[forbidden]; found {
			t.Fatal("leaked", forbidden)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func identityHandlerEmptyExamScope(mock sqlmock.Sqlmock, mode string, binding any) {
	if mode == "metadata_error" {
		mock.ExpectQuery("information_schema.tables").WillReturnError(errors.New("PRIVATE_SCOPE_SECRET"))
		return
	}
	tables := sqlmock.NewRows([]string{"table_name"})
	if mode != "absent" {
		for _, name := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_paper_question_snapshot", "el_mng_result_run", "el_mng_result_dimension", "el_mng_result_module"} {
			tables.AddRow(name)
		}
	}
	if mode == "capture" {
		tables.AddRow("el_mng_legacy_pdf_capture")
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(tables)
	if mode == "absent" {
		return
	}
	if mode == "capture" {
		mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns").WithArgs("el_mng_legacy_pdf_capture").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_legacy_pdf_capture", "participant_type").AddRow("el_mng_legacy_pdf_capture", "participant_id"))
	}
	owner := mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_tester WHERE").WithArgs("tester")
	if mode == "owner_error" {
		owner.WillReturnError(errors.New("PRIVATE_OWNER_SECRET"))
		return
	}
	owner.WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("tester", binding, nil))
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
		query := mock.ExpectQuery("SELECT 1 AS protected FROM "+table+" WHERE").WithArgs("tester", "tester")
		if mode == "evidence_error" {
			query.WillReturnError(errors.New("PRIVATE_EVIDENCE_SECRET"))
			return
		}
		rows := sqlmock.NewRows([]string{"protected"})
		if (mode == "snapshot" && table == "el_mng_paper_snapshot") || (mode == "run" && table == "el_mng_result_run") {
			query.WillReturnRows(rows.AddRow(1))
			return
		}
		query.WillReturnRows(rows)
	}
	if mode == "capture" {
		mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_legacy_pdf_capture WHERE").WithArgs("tester", "tester").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
	}
}

func TestManagementTraitsIdentityHTTPTesterEmptyExamProtectedOrphan(t *testing.T) {
	for _, binding := range []any{nil, ""} {
		for _, mode := range []string{"snapshot", "run", "capture", "metadata_error", "owner_error", "evidence_error"} {
			t.Run(fmt.Sprintf("binding_%v_%s", binding, mode), func(t *testing.T) {
				for _, request := range []struct {
					name, target, extra string
				}{
					{"omitted", "/identity", ""},
					{"query_only", "/identity?examId=query-exam", ""},
					{"empty_json", "/identity?examId=query-exam", `,"examId":""`},
					{"null_json", "/identity?examId=query-exam", `,"examId":null`},
				} {
					t.Run(request.name, func(t *testing.T) {
						db, mock := identityHandlerDB(t)
						mock.ExpectQuery("SELECT .*el_tester").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "name", "password", "del_flag", "status"}).AddRow("tester", binding, "Person", "test-password", 0, "0"))
						identityHandlerEmptyExamScope(mock, mode, binding)
						r := gin.New()
						r.POST("/identity", NewTesterHandler(db, &config.Config{}).LoginForm)
						w := httptest.NewRecorder()
						req := httptest.NewRequest("POST", request.target, strings.NewReader(`{"idNumber":"tester","password":"test-password"`+request.extra+`}`))
						req.Header.Set("Content-Type", "application/json")
						r.ServeHTTP(w, req)
						var result struct {
							Code int             `json:"code"`
							Data json.RawMessage `json:"data"`
						}
						if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code == 200 || !strings.Contains(w.Body.String(), "管理特质") || strings.Contains(w.Body.String(), "PRIVATE") || strings.Contains(w.Body.String(), "test-password") || strings.Contains(w.Body.String(), `"password"`) || (len(result.Data) != 0 && string(result.Data) != "null") {
							t.Fatal("protected orphan reached legacy login or leaked data", w.Body.String())
						}
						if err := mock.ExpectationsWereMet(); err != nil {
							t.Fatal(err)
						}
					})
				}
			})
		}
	}
}

func TestManagementTraitsIdentityHTTPTesterEmptyExamDirectHelper(t *testing.T) {
	for _, mode := range []string{"snapshot", "metadata_error", "absent", "unprotected"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			identityHandlerEmptyExamScope(mock, mode, nil)
			w := identityHandlerPost(func(c *gin.Context) {
				handled := NewTesterHandler(db, &config.Config{}).TryManagementTraitsTesterLogin(c, "", &model.Tester{ID: "tester"}, "test-password")
				wantHandled := mode != "absent" && mode != "unprotected"
				if handled != wantHandled {
					t.Errorf("handled=%v want=%v", handled, wantHandled)
				}
				if !handled {
					c.Status(204)
				}
			}, `{}`)
			if strings.Contains(w.Body.String(), "PRIVATE") || strings.Contains(w.Body.String(), "test-password") {
				t.Fatal("helper leaked sensitive data")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type identityCountingBody struct {
	io.Reader
	read int
}

func (b *identityCountingBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	b.read += n
	return n, err
}

func TestManagementTraitsIdentityHTTPCandidateBodyReadBounded(t *testing.T) {
	for _, size := range []int{16 << 10, (16 << 10) + 1, 1 << 20} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			body := `{}` + strings.Repeat(" ", size-2)
			reader := &identityCountingBody{Reader: strings.NewReader(body)}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/identity", reader)
			handled := NewCandidateHandler(db, &config.Config{}).TryManagementTraitsCandidateSave(c)
			if handled != (size > 16<<10) || reader.read > (16<<10)+1 {
				t.Fatalf("handled=%v read=%d size=%d", handled, reader.read, size)
			}
			if !handled {
				raw, err := io.ReadAll(c.Request.Body)
				if err != nil || string(raw) != body {
					t.Fatal("boundary legacy body changed", err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func identityHandlerFullParticipantScope(mock sqlmock.Sqlmock, kind, id string, marker bool) {
	tables := sqlmock.NewRows([]string{"table_name"})
	for _, name := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_paper_question_snapshot", "el_mng_result_run", "el_mng_result_dimension", "el_mng_result_module"} {
		tables.AddRow(name)
	}
	if marker {
		tables.AddRow("el_mng_legacy_protection")
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(tables)
	if marker {
		mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns").WithArgs("el_mng_legacy_protection").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_legacy_protection", "participant_type").AddRow("el_mng_legacy_protection", "participant_id"))
	}
	mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_" + kind + " WHERE").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow(id, nil, nil))
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM "+table+" WHERE").WithArgs(kind, id).WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	}
	if marker {
		mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_legacy_protection WHERE").WithArgs(kind, id).WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
	}
}

func TestManagementTraitsIdentityHTTPMarkerOnlyNeverWrites(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		t.Run(kind, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			cfg := &config.Config{}
			cfg.Jwt.Secret = "test-only-secret"
			if kind == "tester" {
				mock.ExpectQuery("SELECT .*el_tester").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "name", "password", "del_flag", "status"}).AddRow("protected", "old", "Person", "test-password", 0, "0"))
			}
			repeat := 1
			if kind == "candidate" {
				repeat = 2
			}
			for i := 0; i < repeat; i++ {
				identityHandlerSchema(mock, 1)
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("legacy").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT .*el_"+kind+".*id =").WithArgs("protected", 2).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("protected", "old", ""))
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("old").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WithArgs(kind, "protected").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT count.*el_mng_result_run").WithArgs(kind, "protected").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				identityHandlerFullParticipantScope(mock, kind, "protected", true)
			}
			identityHandlerRuntimeSchema(t, mock)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "is_open", "state", "assessment_type", "scoring_mode"}).AddRow("legacy", map[string]int{"candidate": 1, "tester": 2}[kind], 0, "legacy", "legacy"))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
			mock.ExpectRollback()
			var w *httptest.ResponseRecorder
			if kind == "candidate" {
				w = identityHandlerPost(NewCandidateHandler(db, cfg).Save, `{"id":"protected","examId":"legacy","name":"Person","telephone":"13800000000"}`)
			} else {
				w = identityHandlerPost(NewTesterHandler(db, cfg).LoginForm, `{"idNumber":"protected","password":"test-password","examId":"legacy"}`)
			}
			var result struct {
				Code int             `json:"code"`
				Data json.RawMessage `json:"data"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code == 0 || result.Code == 200 || !strings.Contains(w.Body.String(), "管理特质") || strings.Contains(w.Body.String(), "test-password") || (len(result.Data) > 0 && string(result.Data) != "null") {
				t.Fatal("marker-only identity reached legacy or leaked response", w.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
