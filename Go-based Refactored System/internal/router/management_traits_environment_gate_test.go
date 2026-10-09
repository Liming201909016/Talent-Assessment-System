package router

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/service"
)

var managementTraitsTestRuntimeRoutes = []string{
	"/exam/api/management-traits/participant/create-paper",
	"/exam/api/management-traits/participant/paper-detail",
	"/exam/api/management-traits/participant/fill-answer",
	"/exam/api/management-traits/participant/submit",
	"/exam/api/management-traits/admin/candidate/resume",
	"/exam/api/management-traits/admin/exam/state",
	"/exam/api/management-traits/profile/freeze",
	"/exam/api/management-traits/profile/detail",
	"/exam/api/management-traits/results/list",
	"/exam/api/management-traits/results/detail",
	"/exam/api/management-traits/reports/generate-test",
	"/exam/api/management-traits/reports/view",
	"/exam/api/management-traits/reports/download",
	"/exam/api/management-traits/reports/template",
	"/exam/api/management-traits/reports/template/download",
	"/exam/api/management-traits/report-reissues/qualification",
	"/exam/api/management-traits/report-reissues",
	"/exam/api/management-traits/report-reissues/generate",
	"/exam/api/management-traits/report-reissues/view",
	"/exam/api/management-traits/report-reissues/download",
	"/exam/api/management-traits/formal/versions",
}

func setupManagementTraitsEnvironmentRouter(t *testing.T, environment string) (*httptest.ResponseRecorder, map[string]bool) {
	t.Helper()
	t.Setenv("REPORT_EFFECTIVE_ENV", environment)
	t.Setenv("MNG_TEST_REPORT_ENV", environment)
	enabled := environment == "local" || environment == "staging" || environment == "production"
	db, mock, log := managementTraitsDIDB(t)
	cfg := managementTraitsDIConfig()
	if enabled {
		mock.ExpectQuery(managementTraitsDIPrecheck).WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"})).RowsWillBeClosed()
	}
	managementTraitsDIConstructorQueries(mock)
	var services []*service.ManagementTraitsRuntimeService
	if enabled {
		services = append(services, service.NewManagementTraitsRuntimeService(db, cfg.Jwt.Secret, 1<<20))
	}
	engine := managementTraitsDISetup(t, cfg, db, mock, log, services...)
	if !enabled && atomic.LoadInt64(&log.prechecks) != 0 {
		t.Fatalf("disabled router executed %d management-traits schema prechecks", atomic.LoadInt64(&log.prechecks))
	}
	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	request := httptest.NewRequest(http.MethodPost, "/exam/api/management-traits/participant/create-paper?environment=local", nil)
	request.Header.Set("X-Environment", "local")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response, routes
}

// TestProductionRouterRegistersManagementTraitsTestRuntime
// 对应：docs/regression-tests.md #FB-216
func TestProductionRouterRegistersManagementTraitsTestRuntime(t *testing.T) {
	_, routes := setupManagementTraitsEnvironmentRouter(t, "production")
	for _, path := range managementTraitsTestRuntimeRoutes {
		if !routes[http.MethodPost+" "+path] && !routes[http.MethodGet+" "+path] {
			t.Errorf("production TEST runtime omitted %s", path)
		}
	}
}

func TestMalformedEnvironmentOmitsManagementTraitsTestRuntime(t *testing.T) {
	response, routes := setupManagementTraitsEnvironmentRouter(t, "prod")
	for _, path := range managementTraitsTestRuntimeRoutes {
		if routes[http.MethodPost+" "+path] || routes[http.MethodGet+" "+path] {
			t.Errorf("disabled TEST runtime registered %s", path)
		}
	}
	for _, legacy := range []string{
		"POST /exam/api/paper/paper/create-paper",
		"POST /exam/api/paper/paper/paper-detail",
		"POST /exam/api/paper/paper/fill-answer",
		"POST /exam/api/paper/paper/hand-exam",
		"POST /exam/api/exam/exam/generate-report",
	} {
		if !routes[legacy] {
			t.Errorf("legacy route removed: %s", legacy)
		}
	}
	if response.Code != http.StatusNotFound {
		t.Fatalf("disabled route spoof response=%d body=%s, want 404", response.Code, response.Body.String())
	}
}

func TestBugFB216_LocalAndStagingRouterRegistersManagementTraitsTestRuntime(t *testing.T) {
	for _, environment := range []string{"local", "staging"} {
		t.Run(environment, func(t *testing.T) {
			_, routes := setupManagementTraitsEnvironmentRouter(t, environment)
			for _, path := range managementTraitsTestRuntimeRoutes {
				if !routes[http.MethodPost+" "+path] && !routes[http.MethodGet+" "+path] {
					t.Errorf("enabled TEST runtime omitted %s", path)
				}
			}
		})
	}
}
