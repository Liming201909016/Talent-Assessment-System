package handler

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/middleware"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	jwtpkg "github.com/talent-assessment/refactored/pkg/jwt"
	"github.com/talent-assessment/refactored/pkg/redisx"
	"gorm.io/gorm"
)

const resumeHandlerBase = "/exam/api/management-traits"
const resumeHandlerBody = `{"examId":"exam","participantId":"existing","paperId":"paper"}`

// No parallel subtests: AuthService uses the process-wide Redis client.
func resumeHandlerRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *config.Config, func(int64, []string) string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.Jwt.Secret = "resume-handler-test-only-secret"
	cfg.Jwt.Header, cfg.Jwt.Prefix, cfg.Jwt.LoginUserKey = "Authorization", "Bearer ", "login_user_key"
	cfg.Jwt.ExpireMinutes = 60
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	previous := redisx.Client
	redisx.Client = client
	t.Cleanup(func() { redisx.Client = previous; _ = client.Close() })
	sequence := 0
	login := func(id int64, permissions []string) string {
		sequence++
		key := fmt.Sprintf("resume-test-session-%d", sequence)
		user := model.LoginUser{UserID: id, Token: key, Permissions: permissions, ExpireTime: time.Now().Add(time.Hour).UnixMilli()}
		raw, err := json.Marshal(user)
		if err != nil {
			t.Fatal("cannot encode synthetic login metadata")
		}
		if err := server.Set(redisx.LoginTokenKey+key, string(raw)); err != nil {
			t.Fatal("cannot store synthetic login metadata")
		}
		server.SetTTL(redisx.LoginTokenKey+key, time.Hour)
		token, err := jwtpkg.Create(cfg.Jwt.Secret, map[string]any{cfg.Jwt.LoginUserKey: key})
		if err != nil {
			t.Fatal("cannot sign synthetic administrator credential")
		}
		return token
	}
	r := gin.New()
	r.Use(middleware.JWT(cfg, service.NewAuthService(cfg, nil, nil)))
	NewManagementTraitsRuntimeHandler(db, cfg).RegisterRoutes(r.Group(resumeHandlerBase))
	return r, cfg, login
}

func resumeHandlerRequest(r *gin.Engine, path, body, admin, participant string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", resumeHandlerBase+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if admin != "" {
		req.Header.Set("Authorization", "Bearer "+admin)
	}
	if participant != "" {
		req.Header.Set("X-Management-Traits-Token", participant)
	}
	r.ServeHTTP(w, req)
	return w
}

func resumeHandlerStatus(t *testing.T, w *httptest.ResponseRecorder, want int, sensitive ...string) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("HTTP status=%d want=%d", w.Code, want)
	}
	for _, value := range sensitive {
		if value != "" && strings.Contains(w.Body.String(), value) {
			t.Fatal("response disclosed private metadata or credential")
		}
	}
	var envelope struct {
		Code    int             `json:"code"`
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &envelope) != nil {
		t.Fatal("response is not a JSON envelope")
	}
	if want == 200 {
		if envelope.Code != 0 || !envelope.Success {
			t.Fatalf("unsuccessful envelope code=%d", envelope.Code)
		}
	} else if envelope.Code != want || envelope.Success || (len(envelope.Data) != 0 && string(envelope.Data) != "null") {
		t.Fatalf("unsafe failure envelope code=%d success=%v", envelope.Code, envelope.Success)
	}
}

func resumeHandlerSQLDone(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	// Never print sqlmock's error: it can contain query parameters or fixture JSON.
	if mock.ExpectationsWereMet() != nil {
		t.Fatal("read-only SQL expectations were not met")
	}
}

// Same DDL-derived metadata as the identity fixture, plus the real legacy
// reference columns now required by CheckRuntimeSchema. No private cache bypass.
func resumeHandlerRuntimeSchema(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "sql", "management_traits_001_runtime.sql"))
	if err != nil {
		t.Fatal("cannot read runtime DDL fixture")
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
		t.Fatal("runtime DDL must contain eleven tables")
	}
	fkCount := 0
	for _, table := range matches {
		name, body := table[1], table[2]
		tables.AddRow(name, "InnoDB")
		for _, c := range columnRE.FindAllStringSubmatch(body, -1) {
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
					t.Fatal("invalid DDL index metadata")
				}
				prefix := 0
				if c[2] != "" {
					prefix, err = strconv.Atoi(c[2])
					if err != nil {
						t.Fatal("invalid DDL index prefix")
					}
				}
				indexes.AddRow(name, key, nonUnique, i+1, c[1], prefix)
			}
		}
		for _, fk := range fkRE.FindAllStringSubmatch(body, -1) {
			fkCount++
			local, remote := strings.Split(fk[2], ","), strings.Split(fk[4], ",")
			if len(local) != len(remote) {
				t.Fatal("invalid DDL foreign-key metadata")
			}
			for i, c := range local {
				fks.AddRow(name, fk[1], c, i+1, fk[3], remote[i], fk[6], fk[5])
			}
		}
	}
	if fkCount != 15 {
		t.Fatal("runtime DDL must contain fifteen foreign keys")
	}
	for table, names := range map[string][]string{
		"el_exam": {"id"}, "el_exam_repo": {"exam_id", "repo_id"},
		"el_paper": {"id", "exam_id", "user_id"}, "el_paper_qu": {"id", "paper_id", "qu_id"},
		"el_paper_qu_answer": {"id", "paper_id", "qu_id", "answer_id"},
		"el_candidate":       {"id", "exam_id", "paper_id"}, "el_tester": {"id", "exam_id", "paper_id"},
		"el_repo": {"id"}, "el_qu_repo": {"id", "repo_id", "qu_id"}, "el_qu": {"id"}, "el_qu_answer": {"id", "qu_id"},
	} {
		for _, name := range names {
			nullable := "NO"
			if (table == "el_candidate" && name == "paper_id") || (table == "el_tester" && (name == "paper_id" || name == "exam_id")) {
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

type resumeHandlerFixture struct {
	bundle             model.ManagementTraitsDefinitionBundle
	profile            model.ManagementTraitsExamProfile
	paper              model.Paper
	snapshot           model.ManagementTraitsPaperSnapshot
	questions          []any
	legacy             []any
	buckets            []any
	ownerKind, ownerID string
	delFlag            int
	endTime            *time.Time
}

// Independent fixture; identityHandlerFrozenResume deliberately stays untouched.
// Both HTTP requests read the same 140-question/700-option frozen snapshot.
func resumeHandlerFixtureNew(t *testing.T, remaining time.Duration) resumeHandlerFixture {
	t.Helper()
	bundle, profile, manifest, mapping := identityHandlerSource(t)
	deadline := time.Now().Add(remaining).Truncate(time.Second)
	start := deadline.Add(-25 * time.Minute)
	evidence := struct {
		Schema      string    `json:"schema"`
		Source      string    `json:"source"`
		RepoCode    string    `json:"repoCode"`
		CapturedAt  time.Time `json:"capturedAt"`
		ManifestSHA string    `json:"manifestSha"`
		MappingSHA  string    `json:"mappingSha"`
	}{"mng-current-source-evidence-v1", "current-source-not-client-leader-full-not-historical", "00201", *profile.FrozenAt, bundle.ScoringManifestSHA, profile.MappingSHA}
	raw, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal("cannot encode synthetic evidence")
	}
	v := resumeHandlerFixture{bundle: bundle, profile: profile, ownerKind: "candidate", ownerID: "existing"}
	v.paper = model.Paper{ID: "paper", ExamID: "exam", UserID: "existing", State: 1, TotalTime: 25, CreateTime: &start, LimitTime: &deadline}
	v.snapshot = model.ManagementTraitsPaperSnapshot{PaperID: "paper", ExamID: "exam", ProfileExamID: &v.profile.ExamID, BundleID: bundle.ID, Source: "new_creation", EvidenceSnapshot: string(raw), EvidenceSHA: fmt.Sprintf("%x", sha256.Sum256(raw)), MappingSnapshot: profile.MappingSnapshot, MappingSHA: profile.MappingSHA, ScoringManifestSHA: bundle.ScoringManifestSHA, ParticipantType: "candidate", ParticipantID: "existing", ParticipantSnapshot: `{"name":"Person","gender":"","telephone":"13800000000","affiliation":"","post":""}`, FieldContract: profile.FieldContract, IdentitySource: "submitted_snapshot", SourceCapturedAt: *profile.FrozenAt, StartedAt: start, LimitTime: &deadline, CreatedAt: start}
	v.questions, v.legacy, v.buckets = make([]any, 0, 140), make([]any, 0, 140), make([]any, 0, 700)
	validated := make([]model.ManagementTraitsPaperQuestionSnapshot, 0, 140)
	for i, q := range mapping.Questions {
		mq := manifest.Questions[q.Number-1]
		pqID := fmt.Sprintf("pq%d", q.Number)
		options, err := json.Marshal(q.Options)
		if err != nil {
			t.Fatal("cannot encode synthetic options")
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
			t.Fatal("cannot encode synthetic question")
		}
		question := model.ManagementTraitsPaperQuestionSnapshot{ID: fmt.Sprintf("snapshot%d", q.Number), PaperID: "paper", PaperQuestionID: pqID, SourceQuestionID: q.SourceQuestionID, Number: q.Number, DisplayOrder: i + 1, DimensionKey: mq.DimensionKey, Reverse: mq.Reverse, Content: q.Content, OptionsSnapshot: string(options), ScoringSnapshotSHA: fmt.Sprintf("%x", sha256.Sum256(frozenJSON)), CreatedAt: start}
		validated = append(validated, question)
		v.questions = append(v.questions, question)
		v.legacy = append(v.legacy, model.PaperQu{ID: pqID, PaperID: "paper", QuID: q.SourceQuestionID, QuType: 1, Sort: i + 1})
		for _, o := range q.Options {
			v.buckets = append(v.buckets, model.PaperQuAnswer{ID: "bucket-" + o.SourceOptionID, PaperID: "paper", QuID: q.SourceQuestionID, AnswerID: o.SourceOptionID, Sort: o.DisplayOrder, Score: o.Raw})
		}
	}
	if _, err := service.ValidateManagementTraitsStoredInput(bundle, v.snapshot, validated, 1<<20); err != nil {
		t.Fatal("synthetic frozen input is invalid")
	}
	return v
}

func resumeHandlerOwnerRows(v resumeHandlerFixture) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"kind", "id", "exam_id", "paper_id", "name", "telephone", "del_flag", "status", "end_time"}).AddRow(v.ownerKind, v.ownerID, "exam", "paper", "mutable-name-not-frozen", "13900000000", v.delFlag, "0", v.endTime)
}

// stop=5: owner rejected; stop=7: loader metadata rejected; stop=0: full loader.
func resumeHandlerExpectLoad(t *testing.T, mock sqlmock.Sqlmock, v resumeHandlerFixture, stop int) {
	t.Helper()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WithArgs("paper", 1).WillReturnRows(identityHandlerModelRows(t, v.paper))
	mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WithArgs("paper", 1).WillReturnRows(identityHandlerModelRows(t, v.snapshot))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WithArgs("bundle", 1).WillReturnRows(identityHandlerModelRows(t, v.bundle))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WithArgs("exam", 1).WillReturnRows(identityHandlerModelRows(t, v.profile))
	mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WithArgs("paper", "paper").WillReturnRows(resumeHandlerOwnerRows(v))
	if stop == 5 {
		return
	}
	mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WithArgs("paper", 141).WillReturnRows(identityHandlerModelRows(t, v.questions...))
	mock.ExpectQuery("SELECT .*el_paper_qu").WithArgs("paper", 141).WillReturnRows(identityHandlerModelRows(t, v.legacy...))
}

func resumeHandlerExpectIssue(t *testing.T, mock sqlmock.Sqlmock, v resumeHandlerFixture, mode string) {
	t.Helper()
	stop := 0
	if mode == "deleted" {
		stop = 5
	}
	resumeHandlerExpectLoad(t, mock, v, stop)
	switch mode {
	case "deleted", "revoked", "expired", "completed", "ended_owner", "wrong_exam", "wrong_participant", "tester":
		mock.ExpectRollback()
		return
	}
	locked := v.bundle
	if mode == "revoked_during_lock" {
		locked.Status = "review-revoked"
	}
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*FOR UPDATE").WithArgs("bundle", 1).WillReturnRows(identityHandlerModelRows(t, locked))
	if mode == "revoked_during_lock" {
		mock.ExpectRollback()
		return
	}
	profile := v.profile
	if mode == "unfrozen" {
		profile.FrozenAt = nil
	}
	mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WithArgs("exam", 1).WillReturnRows(identityHandlerModelRows(t, profile))
	if mode == "unfrozen" {
		mock.ExpectRollback()
		return
	}
	owners := resumeHandlerOwnerRows(v)
	if mode == "deleted_during_lock" {
		deleted := v
		deleted.delFlag = 2
		owners = resumeHandlerOwnerRows(deleted)
	}
	mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WithArgs("paper", 3).WillReturnRows(owners)
	testers := sqlmock.NewRows([]string{"kind", "id", "exam_id", "paper_id", "del_flag", "status"})
	if mode == "competing_tester" {
		testers.AddRow("tester", "other", "exam", "paper", 2, "1")
	}
	mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WithArgs("paper", 3).WillReturnRows(testers)
	if mode == "deleted_during_lock" || mode == "competing_tester" {
		mock.ExpectRollback()
		return
	}
	end := time.Now().Add(-time.Hour)
	exam := model.Exam{ID: "exam", IsOpen: 1, State: 3, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, EndTime: &end}
	if mode == "closed_exam" {
		exam.IsOpen = 2
	}
	mock.ExpectQuery("SELECT .*el_exam").WithArgs("exam", 1).WillReturnRows(identityHandlerModelRows(t, exam))
	if mode == "closed_exam" {
		mock.ExpectRollback()
		return
	}
	buckets := v.buckets
	if mode == "missing_bucket" {
		buckets = buckets[:699]
	}
	mock.ExpectQuery("SELECT .*el_paper_qu_answer").WithArgs("paper", 701).WillReturnRows(identityHandlerModelRows(t, buckets...))
	if mode == "missing_bucket" {
		mock.ExpectRollback()
	} else {
		mock.ExpectCommit()
	}
}

func resumeHandlerDecodeIssue(t *testing.T, w *httptest.ResponseRecorder, cfg *config.Config, deadline time.Time, before time.Time) service.ManagementTraitsIdentityResponse {
	t.Helper()
	resumeHandlerStatus(t, w, 200, cfg.Jwt.Secret, "mutable-name-not-frozen", "13900000000")
	var envelope struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &envelope) != nil || len(envelope.Data) != 5 {
		t.Fatal("identity response must have exactly five fields")
	}
	for _, field := range []string{"id", "examId", "paperId", "name", "participantToken"} {
		var value string
		if json.Unmarshal(envelope.Data[field], &value) != nil || value == "" {
			t.Fatal("missing identity field", field)
		}
	}
	raw, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatal("cannot decode identity field metadata")
	}
	var out service.ManagementTraitsIdentityResponse
	if json.Unmarshal(raw, &out) != nil || out.ID != "existing" || out.ExamID != "exam" || out.PaperID != "paper" || out.Name != "Person" {
		t.Fatal("frozen identity binding changed")
	}
	claims, err := service.ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, out.ParticipantToken, service.ManagementTraitsRuntimePaperPurpose, time.Now())
	if err != nil || claims.ValidateBinding("candidate", "existing", "exam", "paper") != nil || claims.ExpiresAt > before.Add(5*time.Minute+time.Second).Unix() || claims.ExpiresAt > deadline.Unix() {
		t.Fatal("invalid short paper credential metadata")
	}
	if _, err := service.ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, out.ParticipantToken, service.ManagementTraitsRuntimeParticipantPurpose, time.Now()); err == nil {
		t.Fatal("resume credential can create a new paper")
	}
	// The real runtime parser accepts HS512 only, not merely any HMAC method.
	return out
}

func TestManagementTraitsHandlerResumeJWTAndAdminGate(t *testing.T) {
	for _, mode := range []string{"anonymous", "invalid_jwt", "query_admin", "participant_header", "participant_bearer", "missing_session", "exam_list", "exam_export", "resume_permission"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			r, cfg, login := resumeHandlerRouter(t, db)
			admin, participant, path, want := "", "", "/admin/candidate/resume", 401
			switch mode {
			case "invalid_jwt":
				admin = "invalid-synthetic-jwt"
			case "query_admin":
				path += "?token=" + login(1, nil)
			case "participant_header", "participant_bearer":
				token, err := service.CreateManagementTraitsRuntimeToken(cfg.Jwt.Secret, service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimePaperPurpose, ParticipantType: "candidate", ParticipantID: "existing", ExamID: "exam", PaperID: "paper", ExpiresAt: time.Now().Add(time.Minute).Unix()}, time.Now())
				if err != nil {
					t.Fatal("cannot sign synthetic participant credential")
				}
				if mode == "participant_header" {
					participant = token
				} else {
					admin = token
				}
			case "missing_session":
				var err error
				admin, err = jwtpkg.Create(cfg.Jwt.Secret, map[string]any{cfg.Jwt.LoginUserKey: "not-stored"})
				if err != nil {
					t.Fatal("cannot sign synthetic missing-session credential")
				}
			case "exam_list":
				admin, want = login(9, []string{"exam:list"}), 403
			case "exam_export":
				admin, want = login(9, []string{"exam:export"}), 403
			case "resume_permission":
				admin, want = login(9, []string{"management-traits:candidate:resume"}), 403
			}
			w := resumeHandlerRequest(r, path, resumeHandlerBody, admin, participant)
			resumeHandlerStatus(t, w, want, cfg.Jwt.Secret, admin, participant)
			resumeHandlerSQLDone(t, mock)
		})
	}
}

func TestManagementTraitsHandlerResumeStrictBodyNoQueries(t *testing.T) {
	bodies := []string{``, `{}`, `[]`, `null`, `{"examId":"exam","participantId":"existing"}`, `{"examId":"exam","participantId":"existing","paperId":"paper","extra":"x"}`, resumeHandlerBody + ` {}`, `{"examId":"exam","participantId":"existing","paperId":"paper","paperId":"other"}`, `{"examId":"exam","participantId":"existing","paperId":"paper","p\u0061perId":"other"}`, `{"ExamId":"exam","participantId":"existing","paperId":"paper"}`, `{"examId":"exam","participantId":"existing","paperId":"paper"`}
	for _, field := range []string{"examId", "participantId", "paperId"} {
		for _, value := range []string{`""`, `" "`, `null`, `0`, `false`, `[]`, `{}`, `"bad/id"`, `"bad?token"`, `"bad\n"`, `"` + strings.Repeat("x", 65) + `"`} {
			body := map[string]json.RawMessage{"examId": json.RawMessage(`"exam"`), "participantId": json.RawMessage(`"existing"`), "paperId": json.RawMessage(`"paper"`)}
			body[field] = json.RawMessage(value)
			raw, err := json.Marshal(body)
			if err != nil {
				t.Fatal("cannot encode strict-body test metadata")
			}
			bodies = append(bodies, string(raw))
		}
	}
	bodies = append(bodies, `{"examId":"`+strings.Repeat("x", 16<<10)+`","participantId":"existing","paperId":"paper"}`)
	for i, body := range bodies {
		t.Run(fmt.Sprintf("case_%02d", i), func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			r, cfg, login := resumeHandlerRouter(t, db)
			admin := login(1, nil)
			w := resumeHandlerRequest(r, "/admin/candidate/resume", body, admin, "")
			resumeHandlerStatus(t, w, 400, cfg.Jwt.Secret, admin)
			resumeHandlerSQLDone(t, mock)
		})
	}
}

func TestManagementTraitsHandlerResumeIssueThenDetailZeroWrites(t *testing.T) {
	for _, mode := range []string{"admin", "wildcard", "retired", "near_deadline"} {
		t.Run(mode, func(t *testing.T) {
			remaining := 20 * time.Minute
			if mode == "near_deadline" {
				remaining = 90 * time.Second
			}
			v := resumeHandlerFixtureNew(t, remaining)
			if mode == "retired" {
				v.bundle.Status = "retired"
			}
			db, mock := identityHandlerDB(t)
			r, cfg, login := resumeHandlerRouter(t, db)
			admin := login(1, nil)
			if mode == "wildcard" {
				admin = login(9, []string{"*:*:*"})
			}
			resumeHandlerRuntimeSchema(t, mock)
			resumeHandlerExpectIssue(t, mock, v, "")
			before := time.Now()
			w := resumeHandlerRequest(r, "/admin/candidate/resume", resumeHandlerBody, admin, "")
			issued := resumeHandlerDecodeIssue(t, w, cfg, *v.snapshot.LimitTime, before)
			resumeHandlerSQLDone(t, mock)
			initial, err := service.ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, issued.ParticipantToken, service.ManagementTraitsRuntimePaperPurpose, time.Now())
			if err != nil {
				t.Fatal("issued credential rejected by runtime parser")
			}
			// Schema is cached on the same handler's service, not queued again.
			for repeat := 0; repeat < 2; repeat++ {
				resumeHandlerExpectLoad(t, mock, v, 0)
				mock.ExpectQuery("SELECT .*el_paper_qu_answer").WithArgs("paper", 701).WillReturnRows(identityHandlerModelRows(t, v.buckets...))
				mock.ExpectCommit()
				w = resumeHandlerRequest(r, "/participant/paper-detail", `{"paperId":"paper"}`, "", issued.ParticipantToken)
				resumeHandlerStatus(t, w, 200, cfg.Jwt.Secret, "mutable-name-not-frozen")
				var detail struct {
					Data service.ManagementTraitsRuntimePaperDetail `json:"data"`
				}
				if json.Unmarshal(w.Body.Bytes(), &detail) != nil || detail.Data.PaperID != "paper" || detail.Data.ExamID != "exam" || !detail.Data.Deadline.Equal(*v.snapshot.LimitTime) || !detail.Data.StartedAt.Equal(v.snapshot.StartedAt) || detail.Data.State != 1 || detail.Data.Answered != 0 || len(detail.Data.Questions) != 140 {
					t.Fatal("paper snapshot or deadline changed during resume")
				}
				for _, q := range detail.Data.Questions {
					if len(q.Options) != 5 || q.SelectedOptionID != nil {
						t.Fatal("frozen options or answer state changed")
					}
				}
				refreshed, err := service.ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, detail.Data.PaperToken, service.ManagementTraitsRuntimePaperPurpose, time.Now())
				if err != nil || refreshed.ValidateBinding("candidate", "existing", "exam", "paper") != nil || refreshed.ExpiresAt > initial.ExpiresAt {
					t.Fatal("HTTP detail extended administrator resume credential")
				}
				issued.ParticipantToken = detail.Data.PaperToken
				resumeHandlerSQLDone(t, mock)
			}
		})
	}
}

func TestManagementTraitsHandlerResumeRejectedFrozenScope(t *testing.T) {
	for _, mode := range []string{"wrong_exam", "wrong_participant", "tester", "deleted", "completed", "ended_owner", "revoked", "revoked_during_lock", "unfrozen", "deleted_during_lock", "competing_tester", "closed_exam", "missing_bucket", "expired"} {
		t.Run(mode, func(t *testing.T) {
			v := resumeHandlerFixtureNew(t, 20*time.Minute)
			body := resumeHandlerBody
			switch mode {
			case "wrong_exam":
				body = `{"examId":"foreign","participantId":"existing","paperId":"paper"}`
			case "wrong_participant":
				body = `{"examId":"exam","participantId":"foreign","paperId":"paper"}`
			case "tester":
				v.ownerKind, v.snapshot.ParticipantType = "tester", "tester"
			case "deleted":
				v.delFlag = 2
			case "completed":
				v.paper.State = 2
			case "ended_owner":
				now := time.Now()
				v.endTime = &now
			case "revoked":
				v.bundle.Status = "review-revoked"
			case "expired":
				v = resumeHandlerFixtureNew(t, -time.Minute)
			}
			db, mock := identityHandlerDB(t)
			r, cfg, login := resumeHandlerRouter(t, db)
			admin := login(1, nil)
			resumeHandlerRuntimeSchema(t, mock)
			resumeHandlerExpectIssue(t, mock, v, mode)
			w := resumeHandlerRequest(r, "/admin/candidate/resume", body, admin, "")
			resumeHandlerStatus(t, w, 409, cfg.Jwt.Secret, admin, "participantToken", "scoringManifest", "mappingSnapshot")
			if mode == "expired" && !strings.Contains(w.Body.String(), "到期") {
				t.Fatal("expiry did not produce controlled expiry message")
			}
			resumeHandlerSQLDone(t, mock)
		})
	}
}

func TestManagementTraitsHandlerResumePrivateReadErrors(t *testing.T) {
	for _, mode := range []string{"schema", "schema_absent", "paper", "foreign_paper"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			r, cfg, login := resumeHandlerRouter(t, db)
			admin := login(1, nil)
			body, want := resumeHandlerBody, 409
			if mode == "schema" {
				mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnError(errors.New("PRIVATE_SYNTHETIC_READ_SENTINEL"))
			} else if mode == "schema_absent" {
				mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
				want = 503
			} else {
				resumeHandlerRuntimeSchema(t, mock)
				mock.ExpectBegin()
				paper := "paper"
				if mode == "foreign_paper" {
					paper = "foreign-paper"
					body = `{"examId":"exam","participantId":"existing","paperId":"foreign-paper"}`
				}
				query := mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WithArgs(paper, 1)
				if mode == "foreign_paper" {
					query.WillReturnRows(sqlmock.NewRows([]string{"id"}))
				} else {
					query.WillReturnError(errors.New("PRIVATE_SYNTHETIC_READ_SENTINEL"))
				}
				mock.ExpectRollback()
			}
			w := resumeHandlerRequest(r, "/admin/candidate/resume", body, admin, "")
			resumeHandlerStatus(t, w, want, cfg.Jwt.Secret, admin, "PRIVATE_SYNTHETIC_READ_SENTINEL", "SELECT", "participantToken")
			resumeHandlerSQLDone(t, mock)
		})
	}
}

func TestManagementTraitsHandlerResumeCredentialMisuse(t *testing.T) {
	for _, mode := range []string{"participant_purpose", "tampered", "expired", "query_only", "cross_paper", "cross_candidate", "cross_exam", "tester_binding", "create_paper"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			r, cfg, _ := resumeHandlerRouter(t, db)
			purpose, kind, id, exam, paper := service.ManagementTraitsRuntimePaperPurpose, "candidate", "existing", "exam", "paper"
			path, body, want := "/participant/paper-detail", `{"paperId":"paper"}`, 401
			expires := time.Now().Add(time.Minute).Unix()
			switch mode {
			case "participant_purpose":
				purpose, paper = service.ManagementTraitsRuntimeParticipantPurpose, ""
			case "expired":
				expires = time.Now().Add(-time.Minute).Unix()
			case "cross_paper":
				body = `{"paperId":"other-paper"}`
			case "cross_candidate":
				id, want = "other-candidate", 409
			case "cross_exam":
				exam, want = "other-exam", 409
			case "tester_binding":
				kind, want = "tester", 409
			case "create_paper":
				path, body = "/participant/create-paper", `{"examId":"exam"}`
			}
			token, err := jwtpkg.Create(cfg.Jwt.Secret, map[string]any{"purpose": purpose, "participant_type": kind, "participant_id": id, "exam_id": exam, "paper_id": paper, "exp": expires})
			if err != nil {
				t.Fatal("cannot sign synthetic misuse credential")
			}
			if mode == "tampered" {
				parts := strings.Split(token, ".")
				parts[2] = strings.Repeat("A", len(parts[2]))
				token = strings.Join(parts, ".")
			}
			header := token
			if mode == "query_only" {
				path += "?token=" + token
				header = ""
			}
			if want == 409 {
				v := resumeHandlerFixtureNew(t, 20*time.Minute)
				resumeHandlerRuntimeSchema(t, mock)
				resumeHandlerExpectLoad(t, mock, v, 0)
				mock.ExpectRollback()
			}
			w := resumeHandlerRequest(r, path, body, "", header)
			resumeHandlerStatus(t, w, want, cfg.Jwt.Secret, token, "paperToken", "participantToken")
			resumeHandlerSQLDone(t, mock)
		})
	}
}

func TestManagementTraitsHandlerResumeReadFixture(t *testing.T) {
	// Validate the independent SQL fixture before the new HTTP route exists.
	for _, mode := range []string{"candidate-current-source", "retired"} {
		t.Run(mode, func(t *testing.T) {
			v := resumeHandlerFixtureNew(t, 20*time.Minute)
			v.bundle.Status = mode
			db, mock := identityHandlerDB(t)
			resumeHandlerRuntimeSchema(t, mock)
			resumeHandlerExpectIssue(t, mock, v, "")
			svc := service.NewManagementTraitsRuntimeService(db, "fixture-test-only", 1<<20)
			issued, err := svc.IssueCandidateResume(context.Background(), "exam", "existing", "paper")
			if err != nil {
				t.Fatalf("issuance fixture rejected: controlled error type=%T", err)
			}
			resumeHandlerSQLDone(t, mock)
			claims, err := service.ParseManagementTraitsRuntimeToken("fixture-test-only", issued.ParticipantToken, service.ManagementTraitsRuntimePaperPurpose, time.Now())
			if err != nil {
				t.Fatal("fixture credential rejected")
			}
			resumeHandlerExpectLoad(t, mock, v, 0)
			mock.ExpectQuery("SELECT .*el_paper_qu_answer").WithArgs("paper", 701).WillReturnRows(identityHandlerModelRows(t, v.buckets...))
			mock.ExpectCommit()
			detail, err := svc.PaperDetail(context.Background(), claims)
			if err != nil || detail.PaperID != "paper" || len(detail.Questions) != 140 || !detail.Deadline.Equal(*v.snapshot.LimitTime) {
				t.Fatal("consumption fixture rejected or snapshot changed")
			}
			resumeHandlerSQLDone(t, mock)
		})
	}
}
