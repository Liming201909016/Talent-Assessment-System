package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Regression MT-HTTP-01: the guard must be installed in the real Setup,
// not just in an isolated test engine. No real database or credentials.
func TestBugManagementTraitsRealRouterRejectsProtectedLegacy(t *testing.T) {
	t.Setenv("REPORT_EFFECTIVE_ENV", "local")
	t.Setenv("MNG_TEST_REPORT_ENV", "local")
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
	for _, column := range []string{"mbti_type", "mbti_scores"} {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM information_schema.COLUMNS").WithArgs(column).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	cfg := &config.Config{}
	cfg.Jwt.Header, cfg.Jwt.Prefix, cfg.Jwt.Secret = "Authorization", "Bearer ", "local-router-test-only"
	cfg.Competency.ExpiryScanSeconds, cfg.Competency.ExpiryBatchSize = 3600, 1
	engine, shutdown := Setup(cfg, db)
	shutdown()

	for _, path := range []string{"/exam/api/paper/paper/save", "/exam/api/candidate/update", "/exam/api/tester"} {
		t.Run(path, func(t *testing.T) {
			tables := sqlmock.NewRows([]string{"table_name"})
			for _, table := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_paper_question_snapshot", "el_mng_result_run", "el_mng_result_dimension", "el_mng_result_module"} {
				tables.AddRow(table)
			}
			mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(tables)
			mock.ExpectQuery("SELECT id, exam_id FROM el_paper WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}))
			mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
			mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_tester WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
			mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			method := http.MethodPost
			if path == "/exam/api/tester" {
				method = http.MethodPut
			}
			req := httptest.NewRequest(method, path, strings.NewReader(`{"id":"new-person","examId":"new-exam","paperId":"new-paper"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("protected legacy request escaped real router: HTTP=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
	for _, path := range []string{"/exam/api/management-traits/profile/detail", "/exam/api/management-traits/results/detail"} {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("administrator endpoint not authenticated: %s HTTP=%d", path, w.Code)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
