package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type managementTraitsDIQueryLog struct {
	logger.Interface
	prechecks int64
	attempts  int64
	errors    int64
}

func (l *managementTraitsDIQueryLog) Trace(_ context.Context, _ time.Time, sql func() (string, int64), err error) {
	atomic.AddInt64(&l.attempts, 1)
	if err != nil {
		atomic.AddInt64(&l.errors, 1)
	}
	query, _ := sql()
	if strings.HasPrefix(query, "SELECT table_name AS table_name, engine AS engine FROM information_schema.tables") {
		atomic.AddInt64(&l.prechecks, 1)
	}
}

func managementTraitsDIDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *managementTraitsDIQueryLog) {
	t.Helper()
	db, mock := identityHandlerDB(t)
	log := &managementTraitsDIQueryLog{Interface: logger.Default.LogMode(logger.Silent)}
	db.Logger = log
	return db, mock, log
}

func managementTraitsDIConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Jwt.Secret = "handler-di-test-only"
	return cfg
}

func managementTraitsDIPost(t *testing.T, r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func managementTraitsDICandidateScope(mock sqlmock.Sqlmock) {
	for i := 0; i < 2; i++ {
		identityHandlerSchema(mock, 1)
		mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
}

func managementTraitsDICandidateCreate(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	mock.ExpectBegin()
	identityHandlerAdmission(t, mock, 1)
	mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "telephone", "del_flag"}))
	mock.ExpectExec("INSERT INTO `el_candidate`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

func managementTraitsDITesterScope(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT .*el_tester").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "password", "telephone", "del_flag", "status"}).AddRow("tester", "exam", "", "Person", "test-password", "13800000000", 0, "0"))
	identityHandlerSchema(mock, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .*el_tester.*id =").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("tester", "exam", ""))
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	identityHandlerFullParticipantScope(mock, "tester", "tester", false)
}

func managementTraitsDITesterLogin(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	mock.ExpectBegin()
	identityHandlerAdmission(t, mock, 2)
	mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "password", "telephone", "del_flag", "status"}).AddRow("tester", "exam", "", "Person", "test-password", "13800000000", 0, "0"))
	mock.ExpectCommit()
}

func managementTraitsDIIdentityResult(t *testing.T, w *httptest.ResponseRecorder, cfg *config.Config, kind string) {
	t.Helper()
	var result struct {
		Code int                                      `json:"code"`
		Data service.ManagementTraitsIdentityResponse `json:"data"`
	}
	want := 0
	if kind == "tester" {
		want = 200
	}
	if json.Unmarshal(w.Body.Bytes(), &result) != nil || w.Code != 200 || result.Code != want {
		t.Fatalf("%s identity did not succeed: %s", kind, w.Body.String())
	}
	claims, err := service.ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, result.Data.ParticipantToken, service.ManagementTraitsRuntimeParticipantPurpose, time.Now())
	if err != nil || claims.ValidateBinding(kind, result.Data.ID, "exam", "") != nil {
		t.Fatal("issued identity token has incorrect binding", err)
	}
}

func TestBugManagementTraitsDIDefaultIdentityAcrossRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, kind := range []string{"candidate", "tester"} {
		t.Run(kind, func(t *testing.T) {
			db, mock, log := managementTraitsDIDatabase(t)
			cfg := managementTraitsDIConfig()
			r := gin.New()
			body := `{"examId":"exam","name":"Person","telephone":"13800000000"}`
			if kind == "candidate" {
				r.POST("/identity", NewCandidateHandler(db, cfg).Save)
			} else {
				r.POST("/identity", NewTesterHandler(db, cfg).LoginForm)
				body = `{"idNumber":"tester","password":"test-password","examId":"exam"}`
			}
			if atomic.LoadInt64(&log.prechecks) != 0 {
				t.Fatal("default constructor executed SQL")
			}
			for i := 0; i < 2; i++ {
				if kind == "candidate" {
					managementTraitsDICandidateScope(mock)
				} else {
					managementTraitsDITesterScope(mock)
				}
				if i == 0 {
					identityHandlerRuntimeSchema(t, mock)
				}
				if kind == "candidate" {
					managementTraitsDICandidateCreate(t, mock)
				} else {
					managementTraitsDITesterLogin(t, mock)
				}
				managementTraitsDIIdentityResult(t, managementTraitsDIPost(t, r, "/identity", body), cfg, kind)
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			}
			if count := atomic.LoadInt64(&log.prechecks); count != 1 {
				t.Fatalf("two real identity requests: precheck query count=%d, want 1", count)
			}
			t.Log("two real Gin identity requests: schema precheck querycount=1")
		})
	}
}

func managementTraitsDIProfileQueries(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	bundle, profile, _, _ := identityHandlerSource(t)
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WithArgs("exam", 1).WillReturnRows(identityHandlerModelRows(t, profile))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WithArgs("bundle", 1).WillReturnRows(identityHandlerModelRows(t, bundle))
}

func managementTraitsDIProfileGet(t *testing.T, r *gin.Engine) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/profile?examId=exam", nil))
	return w
}

func TestBugManagementTraitsDIDefaultRuntimeAcrossRequests(t *testing.T) {
	db, mock, log := managementTraitsDIDatabase(t)
	cfg := managementTraitsDIConfig()
	h := NewManagementTraitsRuntimeHandler(db, cfg)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("loginUser", &model.LoginUser{UserID: 1}); c.Next() })
	r.GET("/profile", h.ProfileDetail)
	if atomic.LoadInt64(&log.prechecks) != 0 {
		t.Fatal("default constructor executed SQL")
	}
	for i := 0; i < 2; i++ {
		if i == 0 {
			identityHandlerRuntimeSchema(t, mock)
		}
		managementTraitsDIProfileQueries(t, mock)
		w := managementTraitsDIProfileGet(t, r)
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
	if count := atomic.LoadInt64(&log.prechecks); count != 1 {
		t.Fatalf("two runtime requests: precheck query count=%d, want 1", count)
	}
	t.Log("two real Gin runtime requests: schema precheck querycount=1")
}

// Reflect only to call the backwards-compatible optional constructor during
// RED; every assertion below observes real HTTP/SQL behavior, not source text.
func managementTraitsDIInjected(t *testing.T, constructor any, db *gorm.DB, cfg *config.Config, svc *service.ManagementTraitsRuntimeService) any {
	t.Helper()
	f := reflect.ValueOf(constructor)
	if !f.Type().IsVariadic() {
		t.Fatal("constructor cannot inject the shared runtime instance")
	}
	return f.Call([]reflect.Value{reflect.ValueOf(db), reflect.ValueOf(cfg), reflect.ValueOf(svc)})[0].Interface()
}

func TestBugManagementTraitsDIThreeEntrypointsShareService(t *testing.T) {
	db, mock, log := managementTraitsDIDatabase(t)
	cfg := managementTraitsDIConfig()
	svc := service.NewManagementTraitsRuntimeService(db, cfg.Jwt.Secret, 1<<20)
	identityHandlerRuntimeSchema(t, mock)
	if err := svc.CheckRuntimeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	candidate := managementTraitsDIInjected(t, NewCandidateHandler, db, cfg, svc).(*CandidateHandler)
	tester := managementTraitsDIInjected(t, NewTesterHandler, db, cfg, svc).(*TesterHandler)
	runtime := managementTraitsDIInjected(t, NewManagementTraitsRuntimeHandler, db, cfg, svc).(*ManagementTraitsRuntimeHandler)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("loginUser", &model.LoginUser{UserID: 1}); c.Next() })
	r.POST("/candidate", candidate.Save)
	r.POST("/tester", tester.LoginForm)
	r.GET("/profile", runtime.ProfileDetail)
	for i := 0; i < 2; i++ {
		managementTraitsDICandidateScope(mock)
		managementTraitsDICandidateCreate(t, mock)
		managementTraitsDIIdentityResult(t, managementTraitsDIPost(t, r, "/candidate", `{"examId":"exam","name":"Person","telephone":"13800000000"}`), cfg, "candidate")
		managementTraitsDITesterScope(mock)
		managementTraitsDITesterLogin(t, mock)
		managementTraitsDIIdentityResult(t, managementTraitsDIPost(t, r, "/tester", `{"idNumber":"tester","password":"test-password","examId":"exam"}`), cfg, "tester")
		managementTraitsDIProfileQueries(t, mock)
		if w := managementTraitsDIProfileGet(t, r); w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
	if count := atomic.LoadInt64(&log.prechecks); count != 1 {
		t.Fatalf("six shared HTTP requests: precheck query count=%d, want 1", count)
	}
	t.Log("six real Gin requests through candidate/tester/runtime: shared schema precheck querycount=1")
	if db.Callback().Query().Get("mng:schema_ids") != nil || db.Callback().Row().Get("mng:schema_ids") != nil || db.Callback().Create().Get("mng:schema_ids") != nil || db.Callback().Update().Get("mng:schema_ids") != nil {
		t.Fatal("original DB callbacks were polluted")
	}
	longID := strings.Repeat("x", 65)
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WithArgs(longID, 1).WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
	var rows []struct{ ExamID string }
	if err := db.Table("el_mng_exam_profile").Where("exam_id = ?", longID).Limit(1).Find(&rows).Error; err != nil {
		t.Fatal("original DB was guarded", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ProfileDetail(context.Background(), longID); !errors.Is(err, service.ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("shared service accepted invalid ID", err)
	}
}

func TestBugManagementTraitsDIFailedSchemaStickyAndFresh(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		for _, failure := range []bool{false, true} {
			t.Run(kind+map[bool]string{false: "/absent", true: "/error"}[failure], func(t *testing.T) {
				db, mock, log := managementTraitsDIDatabase(t)
				cfg := managementTraitsDIConfig()
				for generation := int64(1); generation <= 2; generation++ {
					r := gin.New()
					body := `{"examId":"exam","name":"Person","telephone":"13800000000"}`
					if kind == "candidate" {
						r.POST("/identity", NewCandidateHandler(db, cfg).Save)
					} else {
						r.POST("/identity", NewTesterHandler(db, cfg).LoginForm)
						body = `{"idNumber":"tester","password":"test-password","examId":"exam"}`
					}
					for i := 0; i < 2; i++ {
						if kind == "candidate" {
							managementTraitsDICandidateScope(mock)
						} else {
							managementTraitsDITesterScope(mock)
						}
						if i == 0 {
							if generation == 2 {
								identityHandlerRuntimeSchema(t, mock)
							} else {
								q := mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables")
								if failure {
									q.WillReturnError(errors.New("private schema error"))
								} else {
									q.WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
								}
							}
						}
						if generation == 2 {
							if kind == "candidate" {
								managementTraitsDICandidateCreate(t, mock)
							} else {
								managementTraitsDITesterLogin(t, mock)
							}
						}
						w := managementTraitsDIPost(t, r, "/identity", body)
						if generation == 2 {
							managementTraitsDIIdentityResult(t, w, cfg, kind)
						} else if !strings.Contains(w.Body.String(), "管理特质身份校验失败") || strings.Contains(w.Body.String(), "private schema error") {
							t.Fatal(w.Body.String())
						}
					}
					if count := atomic.LoadInt64(&log.prechecks); count != generation {
						t.Fatalf("failed handler generation %d: query count=%d", generation, count)
					}
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}

func TestBugManagementTraitsDIConfigDoesNotReviveExistingHandler(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		t.Run(kind, func(t *testing.T) {
			db, mock, log := managementTraitsDIDatabase(t)
			cfg := &config.Config{}
			r := gin.New()
			body := `{"examId":"exam","name":"Person","telephone":"13800000000"}`
			if kind == "candidate" {
				r.POST("/identity", NewCandidateHandler(db, cfg).Save)
			} else {
				r.POST("/identity", NewTesterHandler(db, cfg).LoginForm)
				body = `{"idNumber":"tester","password":"test-password","examId":"exam"}`
			}
			for i := 0; i < 2; i++ {
				if kind == "candidate" {
					managementTraitsDICandidateScope(mock)
				} else {
					managementTraitsDITesterScope(mock)
				}
				w := managementTraitsDIPost(t, r, "/identity", body)
				if !strings.Contains(w.Body.String(), "管理特质身份校验失败") {
					t.Fatal(w.Body.String())
				}
				cfg.Jwt.Secret = "new-secret-must-not-revive"
			}
			if count := atomic.LoadInt64(&log.prechecks); count != 0 {
				t.Fatalf("mutated config revived runtime: precheck count=%d", count)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugManagementTraitsDIRuntimeInvalidConfigAndSecret(t *testing.T) {
	for _, mode := range []string{"nil_config", "empty_secret", "wrong_secret"} {
		t.Run(mode, func(t *testing.T) {
			db, mock, log := managementTraitsDIDatabase(t)
			var cfg *config.Config
			if mode != "nil_config" {
				cfg = &config.Config{}
			}
			if mode == "wrong_secret" {
				cfg.Jwt.Secret = "wrong-secret"
			}
			h := NewManagementTraitsRuntimeHandler(db, cfg)
			r := gin.New()
			r.POST("/paper", h.PaperDetail)
			now := time.Now()
			token, err := service.CreateManagementTraitsRuntimeToken("correct-secret", service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimePaperPurpose, ParticipantType: "candidate", ParticipantID: "person", ExamID: "exam", PaperID: "paper", ExpiresAt: now.Add(time.Hour).Unix()}, now)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("POST", "/paper", strings.NewReader(`{"paperId":"paper"}`))
				req.Header.Set("X-Management-Traits-Token", token)
				r.ServeHTTP(w, req)
				if w.Code != 401 {
					t.Fatal("invalid initial configuration accepted or revived", w.Body.String())
				}
				if cfg != nil {
					cfg.Jwt.Secret = "correct-secret"
				}
			}
			if atomic.LoadInt64(&log.prechecks) != 0 {
				t.Fatal("invalid token reached schema DB")
			}
			freshCfg := &config.Config{}
			freshCfg.Jwt.Secret = "correct-secret"
			fresh := NewManagementTraitsRuntimeHandler(db, freshCfg)
			mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
			rFresh := gin.New()
			rFresh.POST("/paper", fresh.PaperDetail)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/paper", strings.NewReader(`{"paperId":"paper"}`))
			req.Header.Set("X-Management-Traits-Token", token)
			rFresh.ServeHTTP(w, req)
			if w.Code != 503 || atomic.LoadInt64(&log.prechecks) != 1 {
				t.Fatal("fresh assembly did not authenticate independently", w.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugManagementTraitsDIRejectInjectedDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, mode := range []string{"wrong_db", "wrong_secret", "wrong_budget", "explicit_nil"} {
		for _, warmed := range []bool{false, true} {
			for _, kind := range []string{"candidate", "tester", "runtime"} {
				name := mode + "/cold/" + kind
				if warmed {
					name = mode + "/complete_schema/" + kind
				}
				t.Run(name, func(t *testing.T) {
					db, mock, log := managementTraitsDIDatabase(t)
					wrongDB, wrongMock, wrongLog := managementTraitsDIDatabase(t)
					cfg := managementTraitsDIConfig()
					serviceDB, secret, budget := db, cfg.Jwt.Secret, 1<<20
					switch mode {
					case "wrong_db":
						serviceDB = wrongDB
					case "wrong_secret":
						secret = "mismatched-service-secret"
					case "wrong_budget":
						budget--
					}
					svc := service.NewManagementTraitsRuntimeService(serviceDB, secret, budget)
					if mode == "explicit_nil" {
						svc = nil
					}
					if warmed && svc != nil {
						serviceMock := mock
						if mode == "wrong_db" {
							serviceMock = wrongMock
						}
						identityHandlerRuntimeSchema(t, serviceMock)
						if err := svc.CheckRuntimeSchema(context.Background()); err != nil {
							t.Fatal("complete schema warmup", err)
						}
					}
					beforeDB, beforeWrong, beforePrecheck := atomic.LoadInt64(&log.attempts), atomic.LoadInt64(&wrongLog.attempts), atomic.LoadInt64(&log.prechecks)
					r := gin.New()
					switch kind {
					case "candidate":
						r.POST("/identity", NewCandidateHandler(db, cfg, svc).Save)
					case "tester":
						r.POST("/identity", NewTesterHandler(db, cfg, svc).LoginForm)
					case "runtime":
						h := NewManagementTraitsRuntimeHandler(db, cfg, svc)
						r.Use(func(c *gin.Context) { c.Set("loginUser", &model.LoginUser{UserID: 1}); c.Next() })
						r.GET("/profile", h.ProfileDetail)
						r.POST("/paper", h.PaperDetail)
					}
					if atomic.LoadInt64(&log.attempts) != beforeDB || atomic.LoadInt64(&wrongLog.attempts) != beforeWrong {
						t.Fatal("constructor executed SQL")
					}
					for i := 0; i < 2; i++ {
						if kind == "runtime" {
							if w := managementTraitsDIProfileGet(t, r); w.Code != 503 {
								t.Errorf("mismatched runtime was not closed: %s", w.Body.String())
							}
							now := time.Now()
							token, err := service.CreateManagementTraitsRuntimeToken(cfg.Jwt.Secret, service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimePaperPurpose, ParticipantType: "candidate", ParticipantID: "person", ExamID: "exam", PaperID: "paper", ExpiresAt: now.Add(time.Hour).Unix()}, now)
							if err != nil {
								t.Fatal(err)
							}
							w := httptest.NewRecorder()
							req := httptest.NewRequest("POST", "/paper", strings.NewReader(`{"paperId":"paper"}`))
							req.Header.Set("X-Management-Traits-Token", token)
							r.ServeHTTP(w, req)
							if w.Code != 503 {
								t.Errorf("valid cfg token reached mismatched service: %s", w.Body.String())
							}
						} else {
							body := `{"examId":"exam","name":"Person","telephone":"13800000000"}`
							if kind == "candidate" {
								identityHandlerSchema(mock, 1)
								mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
							} else {
								// Identifier/password lookup remains on h.db; no legacy rebind may follow.
								mock.ExpectQuery("SELECT .*el_tester").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "password", "telephone", "del_flag", "status"}).AddRow("tester", "exam", "", "Person", "test-password", "13800000000", 0, "0"))
								body = `{"idNumber":"tester","password":"test-password","examId":"exam"}`
							}
							w := managementTraitsDIPost(t, r, "/identity", body)
							want := 1
							if kind == "tester" {
								want = 500
							}
							var result struct {
								Code int                                      `json:"code"`
								Data service.ManagementTraitsIdentityResponse `json:"data"`
							}
							if json.Unmarshal(w.Body.Bytes(), &result) != nil || w.Code != 200 || result.Code != want || !strings.Contains(w.Body.String(), "管理特质身份校验失败") || result.Data.ParticipantToken != "" {
								t.Errorf("wrong injection escaped into identity/legacy write: %s", w.Body.String())
							}
						}
						if err := mock.ExpectationsWereMet(); err != nil {
							t.Fatal(err)
						}
					}
					if got := atomic.LoadInt64(&wrongLog.attempts); got != beforeWrong {
						t.Errorf("wrong DB SQL attempts=%d want=%d (no operations after warmup)", got, beforeWrong)
					}
					if got := atomic.LoadInt64(&log.errors); got != 0 {
						t.Errorf("unexpected h.db SQL (including legacy writes) masked by rejection=%d", got)
					}
					if atomic.LoadInt64(&log.prechecks) != beforePrecheck {
						t.Error("wrong injection rebuilt a real DB service")
					}
					if kind == "runtime" && atomic.LoadInt64(&log.attempts) != beforeDB {
						t.Error("closed runtime touched h.db")
					}
					if err := wrongMock.ExpectationsWereMet(); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestBugManagementTraitsDINilConfigMatchedEmptySecret(t *testing.T) {
	db, mock, log := managementTraitsDIDatabase(t)
	svc := service.NewManagementTraitsRuntimeService(db, "", 1<<20)
	candidate := NewCandidateHandler(db.Session(&gorm.Session{NewDB: true}), nil, svc)
	runtime := NewManagementTraitsRuntimeHandler(db, nil, svc)
	if candidate.managementTraitsRuntime != svc || runtime.svc != svc {
		t.Fatal("same-pool empty-secret injection was rejected")
	}
	r := gin.New()
	r.POST("/identity", candidate.Save)
	r.POST("/paper", runtime.PaperDetail)
	managementTraitsDICandidateScope(mock)
	w := managementTraitsDIPost(t, r, "/identity", `{"examId":"exam","name":"Person","telephone":"13800000000"}`)
	if !strings.Contains(w.Body.String(), "管理特质身份校验失败") {
		t.Fatal("empty secret allowed new identity", w.Body.String())
	}
	if w := managementTraitsDIPost(t, r, "/paper", `{"paperId":"paper"}`); w.Code != 401 {
		t.Fatal("empty secret accepted token", w.Body.String())
	}
	if atomic.LoadInt64(&log.prechecks) != 0 || atomic.LoadInt64(&log.errors) != 0 {
		t.Fatal("empty-secret identity reached schema/writes")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
