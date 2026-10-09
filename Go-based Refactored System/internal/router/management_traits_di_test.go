package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const managementTraitsDIPrecheck = "SELECT table_name AS table_name, engine AS engine FROM information_schema.tables"

// Count actual GORM SQL attempts, including unexpected queries whose error is
// translated into the same public rejection as a cached schema failure.
type managementTraitsDILogger struct {
	logger.Interface
	prechecks int64
	attempts  int64
	scans     chan struct{}
}

func (l *managementTraitsDILogger) Trace(_ context.Context, _ time.Time, sql func() (string, int64), _ error) {
	atomic.AddInt64(&l.attempts, 1)
	query, _ := sql()
	if strings.HasPrefix(query, managementTraitsDIPrecheck) {
		atomic.AddInt64(&l.prechecks, 1)
	}
	if strings.HasPrefix(query, "SELECT p.id AS paper_id") {
		l.scans <- struct{}{}
	}
}

func managementTraitsDIDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *managementTraitsDILogger) {
	t.Helper()
	pool, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	log := &managementTraitsDILogger{Interface: logger.Default.LogMode(logger.Silent), scans: make(chan struct{}, 1)}
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: pool, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, Logger: log})
	if err != nil {
		t.Fatal(err)
	}
	return db, mock, log
}

func managementTraitsDIConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Jwt.Header, cfg.Jwt.Prefix, cfg.Jwt.Secret = "Authorization", "Bearer ", "router-di-test-only"
	cfg.Competency.ExpiryScanSeconds, cfg.Competency.ExpiryBatchSize = 3600, 1
	return cfg
}

func managementTraitsDIConstructorQueries(mock sqlmock.Sqlmock) {
	for _, column := range []string{"mbti_type", "mbti_scores"} {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM information_schema.COLUMNS").WithArgs(column).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
}

func managementTraitsDISetup(t *testing.T, cfg *config.Config, db *gorm.DB, mock sqlmock.Sqlmock, log *managementTraitsDILogger, svc ...*service.ManagementTraitsRuntimeService) *gin.Engine {
	t.Helper()
	if os.Getenv("REPORT_EFFECTIVE_ENV") == "" && os.Getenv("MNG_TEST_REPORT_ENV") == "" {
		t.Setenv("REPORT_EFFECTIVE_ENV", "local")
		t.Setenv("MNG_TEST_REPORT_ENV", "local")
	}
	mock.MatchExpectationsInOrder(false)
	mock.ExpectQuery("SELECT p.id AS paper_id").WillReturnRows(sqlmock.NewRows([]string{"paper_id", "exam_id", "participant_id", "participant_type"}))
	engine, shutdown := Setup(cfg, db, svc...)
	select {
	case <-log.scans:
	case <-time.After(5 * time.Second):
		shutdown()
		t.Fatal("competency worker initial query did not finish")
	}
	shutdown()
	return engine
}

func managementTraitsDIRequests(t *testing.T, engine *gin.Engine, secret string, status int) {
	t.Helper()
	now := time.Now()
	token, err := service.CreateManagementTraitsRuntimeToken(secret, service.ManagementTraitsRuntimeClaims{
		Purpose: service.ManagementTraitsRuntimePaperPurpose, ParticipantType: "candidate", ParticipantID: "candidate-di", ExamID: "exam-di", PaperID: "paper-di", ExpiresAt: now.Add(time.Hour).Unix(),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		for _, request := range []struct{ path, body string }{
			{"paper-detail", `{"paperId":"paper-di"}`},
			{"fill-answer", `{"paperId":"paper-di","paperQuestionId":"question-di","optionId":"option-di"}`},
			{"submit", `{"paperId":"paper-di","submitType":"manual"}`},
		} {
			req := httptest.NewRequest(http.MethodPost, "/exam/api/management-traits/participant/"+request.path, strings.NewReader(request.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Management-Traits-Token", token)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			if w.Code != status || strings.Contains(w.Body.String(), "private metadata failure") {
				t.Fatalf("%s: status=%d want=%d body=%s", request.path, w.Code, status, w.Body.String())
			}
		}
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK || w.Body.String() != `{"status":"ok"}` {
		t.Fatalf("health did not survive rejected schema: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestBugManagementTraitsDIStartupCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"absent", nil, http.StatusServiceUnavailable},
		{"failed", errors.New("private metadata failure"), http.StatusConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, log := managementTraitsDIDB(t)
			cfg := managementTraitsDIConfig()
			for generation := int64(1); generation <= 2; generation++ {
				q := mock.ExpectQuery(managementTraitsDIPrecheck)
				if tc.err != nil {
					q.WillReturnError(tc.err)
				} else {
					q.WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"})).RowsWillBeClosed()
				}
				managementTraitsDIConstructorQueries(mock)
				engine := managementTraitsDISetup(t, cfg, db, mock, log)
				if count := atomic.LoadInt64(&log.prechecks); count != generation {
					t.Fatalf("Setup generation %d precheck SQL count=%d; want %d before requests", generation, count, generation)
				}
				beforeRequests := atomic.LoadInt64(&log.attempts)
				managementTraitsDIRequests(t, engine, cfg.Jwt.Secret, tc.status)
				managementTraitsDIRequests(t, engine, "wrong-router-secret", http.StatusUnauthorized)
				if got := atomic.LoadInt64(&log.attempts); got != beforeRequests {
					t.Fatalf("cached startup failure executed request-time SQL: %d -> %d", beforeRequests, got)
				}
				if count := atomic.LoadInt64(&log.prechecks); count != generation {
					t.Fatalf("requests retried sticky schema cache: SQL count=%d want=%d", count, generation)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestBugManagementTraitsDIInjectedCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, warmed := range []bool{false, true} {
		name := "cold"
		if warmed {
			name = "warmed"
		}
		t.Run(name, func(t *testing.T) {
			routerDB, routerMock, routerLog := managementTraitsDIDB(t)
			serviceDB := routerDB.Session(&gorm.Session{NewDB: true})
			serviceMock, serviceLog := routerMock, routerLog
			cfg := managementTraitsDIConfig()
			svc := service.NewManagementTraitsRuntimeService(serviceDB, cfg.Jwt.Secret, 1<<20)
			serviceMock.ExpectQuery(managementTraitsDIPrecheck).WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"})).RowsWillBeClosed()
			if warmed {
				if err := svc.CheckRuntimeSchema(context.Background()); !errors.Is(err, service.ErrManagementTraitsRuntimeClosed) {
					t.Fatalf("warm schema: %v", err)
				}
			}
			managementTraitsDIConstructorQueries(routerMock)
			engine := managementTraitsDISetup(t, cfg, routerDB, routerMock, routerLog, svc)
			if atomic.LoadInt64(&serviceLog.prechecks) != 1 {
				t.Fatal("Setup did not consume the injected service's actual DB/cache")
			}
			managementTraitsDIRequests(t, engine, cfg.Jwt.Secret, http.StatusServiceUnavailable)
			if err := svc.CheckRuntimeSchema(context.Background()); !errors.Is(err, service.ErrManagementTraitsRuntimeClosed) {
				t.Fatalf("shared cache changed: %v", err)
			}
			if atomic.LoadInt64(&serviceLog.prechecks) != 1 {
				t.Fatal("HTTP handlers created another service or retried the injected cache")
			}
			for _, mock := range []sqlmock.Sqlmock{routerMock, serviceMock} {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

type managementTraitsDIStrictLogger struct {
	*managementTraitsDILogger
	mu     sync.Mutex
	errors []error
}

func (l *managementTraitsDIStrictLogger) Trace(ctx context.Context, begin time.Time, sql func() (string, int64), err error) {
	l.managementTraitsDILogger.Trace(ctx, begin, sql, err)
	if err != nil {
		query, _ := sql()
		l.mu.Lock()
		l.errors = append(l.errors, fmt.Errorf("%s: %w", query, err))
		l.mu.Unlock()
	}
}

func managementTraitsDICoreScope(mock sqlmock.Sqlmock) {
	tables := sqlmock.NewRows([]string{"table_name"})
	for _, name := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_paper_question_snapshot", "el_mng_result_run", "el_mng_result_dimension", "el_mng_result_module"} {
		tables.AddRow(name)
	}
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT table_name AS table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name,7) = 'el_mng_' ORDER BY table_name") + "$").WillReturnRows(tables).RowsWillBeClosed()
}

func managementTraitsDIIdentityScope(mock sqlmock.Sqlmock, tester bool) {
	for _, table := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_result_dimension", "el_mng_result_module"} {
		mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WithArgs(table).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1)).RowsWillBeClosed()
	}
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT table_name AS table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name, 7) = 'el_mng_' AND (table_name LIKE '%capture%' OR table_name = 'el_mng_exam_draft') ORDER BY table_name") + "$").WillReturnRows(sqlmock.NewRows([]string{"table_name"})).RowsWillBeClosed()
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1)).RowsWillBeClosed()
	if !tester {
		return
	}
	mock.ExpectQuery("SELECT .*el_tester.*id =").WithArgs("tester", 2).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("tester", "exam", "")).RowsWillBeClosed()
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1)).RowsWillBeClosed()
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
		mock.ExpectQuery("SELECT count.*"+table).WithArgs("tester", "tester").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0)).RowsWillBeClosed()
	}
	managementTraitsDICoreScope(mock)
	mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_tester WHERE").WithArgs("tester").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("tester", nil, nil)).RowsWillBeClosed()
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM "+table+" WHERE").WithArgs("tester", "tester").WillReturnRows(sqlmock.NewRows([]string{"protected"})).RowsWillBeClosed()
	}
}

func TestBugManagementTraitsDIRealRouterIdentityInjectedCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routerDB, routerMock, routerLog := managementTraitsDIDB(t)
	serviceDB := routerDB.Session(&gorm.Session{NewDB: true})
	serviceMock, serviceLog := routerMock, routerLog
	routerStrict := &managementTraitsDIStrictLogger{managementTraitsDILogger: routerLog}
	serviceStrict := &managementTraitsDIStrictLogger{managementTraitsDILogger: serviceLog}
	routerDB.Logger, serviceDB.Logger = routerStrict, serviceStrict
	cfg := managementTraitsDIConfig()
	svc := service.NewManagementTraitsRuntimeService(serviceDB, cfg.Jwt.Secret, 1<<20)
	serviceMock.ExpectQuery(managementTraitsDIPrecheck).WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"})).RowsWillBeClosed()
	if err := svc.CheckRuntimeSchema(context.Background()); !errors.Is(err, service.ErrManagementTraitsRuntimeClosed) {
		t.Fatalf("warm missing-schema cache: %v", err)
	}
	managementTraitsDIConstructorQueries(routerMock)
	engine := managementTraitsDISetup(t, cfg, routerDB, routerMock, routerLog, svc)
	routerMock.MatchExpectationsInOrder(true)
	for i := 0; i < 2; i++ {
		for _, request := range []struct {
			path   string
			body   string
			code   int
			tester bool
		}{
			{"/exam/api/candidate/save", `{"examId":"exam","name":"Person","telephone":"13800000000"}`, 1, false},
			{"/exam/api/tester/login", `{"idNumber":"tester","password":"test-password","examId":"exam"}`, 500, true},
		} {
			// The real router's identity exemption still discovers protected tables.
			managementTraitsDICoreScope(routerMock)
			if request.tester {
				routerMock.ExpectQuery("SELECT .*el_tester").WithArgs("tester", "tester", "tester", "exam").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "password", "telephone", "del_flag", "status"}).AddRow("tester", "exam", "", "Person", "test-password", "13800000000", 0, "0")).RowsWillBeClosed()
			} else {
				managementTraitsDIIdentityScope(routerMock, false)
			}
			managementTraitsDIIdentityScope(serviceMock, request.tester)
			req := httptest.NewRequest(http.MethodPost, request.path, strings.NewReader(request.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			var result struct {
				Code    int             `json:"code"`
				Msg     string          `json:"msg"`
				Data    json.RawMessage `json:"data"`
				Success bool            `json:"success"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatalf("%s invalid response: %v body=%s", request.path, err, w.Body.String())
			}
			if w.Code != http.StatusOK || result.Code != request.code || result.Msg != "管理特质身份校验失败，请重新确认测评和身份" || len(result.Data) != 0 || result.Success {
				t.Fatalf("%s anonymous identity did not fail closed: status=%d body=%s", request.path, w.Code, w.Body.String())
			}
			for _, mock := range []sqlmock.Sqlmock{routerMock, serviceMock} {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			}
		}
		managementTraitsDIRequests(t, engine, cfg.Jwt.Secret, http.StatusServiceUnavailable)
		if count := atomic.LoadInt64(&routerLog.prechecks); count != 1 {
			t.Fatalf("iteration %d: same-pool full-schema probes=%d; want 1", i, count)
		}
		for _, log := range []*managementTraitsDIStrictLogger{routerStrict, serviceStrict} {
			log.mu.Lock()
			errs := append([]error(nil), log.errors...)
			log.mu.Unlock()
			if len(errs) != 0 {
				t.Fatalf("SQL errors must not be masked by generic identity/runtime failures: %v", errs)
			}
		}
		for _, mock := range []sqlmock.Sqlmock{routerMock, serviceMock} {
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestBugManagementTraitsDISetupRejectInjectedDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, mode := range []string{"wrong_db", "wrong_secret", "wrong_budget", "explicit_nil"} {
		for _, warmed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/warmed_%t", mode, warmed), func(t *testing.T) {
				db, mock, log := managementTraitsDIDB(t)
				wrongDB, wrongMock, wrongLog := managementTraitsDIDB(t)
				strict := &managementTraitsDIStrictLogger{managementTraitsDILogger: log}
				db.Logger = strict
				cfg := managementTraitsDIConfig()
				serviceDB, serviceMock, serviceLog := db, mock, log
				secret, budget := cfg.Jwt.Secret, 1<<20
				switch mode {
				case "wrong_db":
					serviceDB, serviceMock, serviceLog = wrongDB, wrongMock, wrongLog
				case "wrong_secret":
					secret = "mismatched-service-secret"
				case "wrong_budget":
					budget--
				}
				svc := service.NewManagementTraitsRuntimeService(serviceDB, secret, budget)
				if mode == "explicit_nil" {
					svc = nil
				} else if warmed {
					serviceMock.ExpectQuery(managementTraitsDIPrecheck).WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"})).RowsWillBeClosed()
					if err := svc.CheckRuntimeSchema(context.Background()); !errors.Is(err, service.ErrManagementTraitsRuntimeClosed) {
						t.Fatal(err)
					}
				}
				beforeWrong, beforePrecheck := atomic.LoadInt64(&wrongLog.attempts), atomic.LoadInt64(&serviceLog.prechecks)
				managementTraitsDIConstructorQueries(mock)
				engine := managementTraitsDISetup(t, cfg, db, mock, log, svc)
				managementTraitsDIRequests(t, engine, cfg.Jwt.Secret, http.StatusServiceUnavailable)
				managementTraitsDIRequests(t, engine, "untrusted-secret", http.StatusUnauthorized)
				mock.MatchExpectationsInOrder(true)
				managementTraitsDICoreScope(mock)
				managementTraitsDIIdentityScope(mock, false)
				req := httptest.NewRequest(http.MethodPost, "/exam/api/candidate/save", strings.NewReader(`{"examId":"exam","name":"Person","telephone":"13800000000"}`))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				engine.ServeHTTP(w, req)
				var result struct {
					Code    int             `json:"code"`
					Data    json.RawMessage `json:"data"`
					Success bool            `json:"success"`
				}
				if json.Unmarshal(w.Body.Bytes(), &result) != nil || w.Code != 200 || result.Code != 1 || result.Success || len(result.Data) != 0 || !strings.Contains(w.Body.String(), "管理特质身份校验失败") {
					t.Errorf("router identity escaped fail-closed assembly: %s", w.Body.String())
				}
				if got := atomic.LoadInt64(&wrongLog.attempts); got != beforeWrong {
					t.Errorf("Setup/HTTP touched wrong DB: attempts %d -> %d", beforeWrong, got)
				}
				if got := atomic.LoadInt64(&serviceLog.prechecks); got != beforePrecheck {
					t.Errorf("wrong injection prechecked/rebuilt actual DB: %d -> %d", beforePrecheck, got)
				}
				strict.mu.Lock()
				errs := append([]error(nil), strict.errors...)
				strict.mu.Unlock()
				if len(errs) != 0 {
					t.Errorf("SQL errors masked by public rejection: %v", errs)
				}
				for _, m := range []sqlmock.Sqlmock{mock, wrongMock} {
					if err := m.ExpectationsWereMet(); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}
