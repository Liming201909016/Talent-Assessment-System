package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	jwtpkg "github.com/talent-assessment/refactored/pkg/jwt"
	"github.com/talent-assessment/refactored/pkg/redisx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func retiredModulesTestRoot(t *testing.T) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(current), "..", "..", ".."))
}

func readRetirementFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// TestBugFB098_UnsupportedRoutesAreRetired
// 对应：docs/regression-tests.md #FB-098
// 复现：无真实能力的 monitor/tool/user-repo/wrong-book 路由仍返回成功占位响应。
// 期望：生产路由中不存在这些端点，保留的审计只读列表和服务监控仍存在。
func TestBugFB098_UnsupportedRoutesAreRetired(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{}
	cfg.Jwt.Header = "Authorization"
	cfg.Jwt.Prefix = "Bearer "
	cfg.Jwt.Secret = "retirement-test-secret"
	cfg.Jwt.LoginUserKey = "login_user_key"
	cfg.Competency.ExpiryScanSeconds = 3600
	cfg.Competency.ExpiryBatchSize = 1
	engine, shutdown := Setup(cfg, db)
	shutdown()

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"GET /monitor/online/list", "DELETE /monitor/online/:tokenId",
		"GET /monitor/job/list", "GET /monitor/job/:jobId", "POST /monitor/job", "PUT /monitor/job",
		"DELETE /monitor/job/:jobIds", "PUT /monitor/job/changeStatus", "PUT /monitor/job/run",
		"GET /monitor/jobLog/list", "DELETE /monitor/jobLog/:jobLogIds", "DELETE /monitor/jobLog/clean",
		"GET /monitor/cache",
		"DELETE /monitor/operlog/:operIds", "DELETE /monitor/operlog/clean",
		"DELETE /monitor/logininfor/:infoIds", "DELETE /monitor/logininfor/clean",
		"GET /tool/gen/list", "GET /tool/gen/:tableId", "GET /tool/gen/preview/:tableId", "GET /tool/gen/db/list",
		"POST /tool/gen/importTable", "PUT /tool/gen", "DELETE /tool/gen/:tableIds",
		"GET /tool/gen/genCode/:tableName", "GET /tool/gen/synchDb/:tableName",
	} {
		if routes[route] {
			t.Errorf("retired route is still registered: %s", route)
		}
	}
	for _, prefix := range []string{"/exam/api/user/repo/", "/exam/api/user/wrong-book/"} {
		for route := range routes {
			if strings.Contains(route, prefix) {
				t.Errorf("dead generic route is still registered: %s", route)
			}
		}
	}
	for _, route := range []string{"GET /monitor/server", "GET /monitor/operlog/list", "GET /monitor/logininfor/list"} {
		if !routes[route] {
			t.Errorf("real read-only route was removed: %s", route)
		}
	}

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	oldRedis := redisx.Client
	redisx.Client = redisClient
	t.Cleanup(func() { redisx.Client = oldRedis })
	loginUser := &model.LoginUser{UserID: 1, Token: "retirement-test-token", Permissions: []string{"*:*:*"}, Roles: []string{"admin"}}
	rawLoginUser, err := json.Marshal(loginUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := redisClient.Set(context.Background(), redisx.LoginTokenKey+loginUser.Token, rawLoginUser, 0).Err(); err != nil {
		t.Fatal(err)
	}
	token, err := jwtpkg.Create(cfg.Jwt.Secret, map[string]any{cfg.Jwt.LoginUserKey: loginUser.Token})
	if err != nil {
		t.Fatal(err)
	}
	for _, request := range []struct{ method, path string }{
		{http.MethodGet, "/monitor/online/list"},
		{http.MethodGet, "/monitor/job/list"},
		{http.MethodGet, "/monitor/jobLog/list"},
		{http.MethodGet, "/monitor/cache"},
		{http.MethodDelete, "/monitor/operlog/1"},
		{http.MethodDelete, "/monitor/logininfor/1"},
		{http.MethodGet, "/tool/gen/list"},
		{http.MethodPost, "/exam/api/user/repo/paging"},
		{http.MethodPost, "/exam/api/user/wrong-book/list"},
	} {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(request.method, request.path, nil)
		req.Header.Set(cfg.Jwt.Header, cfg.Jwt.Prefix+token)
		engine.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusNotFound {
			t.Errorf("retired endpoint status=%d, want 404: %s %s", recorder.Code, request.method, request.path)
		}
	}
}

func TestBugFB098_StubImplementationIsRemovedAfterAllCallersDisappear(t *testing.T) {
	root := retiredModulesTestRoot(t)
	routerSource := readRetirementFile(t, filepath.Join(root, "Go-based Refactored System", "internal", "router", "router.go"))
	for _, marker := range []string{"handler.Stub(", "handler.AjaxStub(", "handler.TableStub(", "stubGroup("} {
		if strings.Contains(routerSource, marker) {
			t.Errorf("stub registration remains: %s", marker)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "Go-based Refactored System", "internal", "handler", "stub.go")); !os.IsNotExist(err) {
		t.Error("internal/handler/stub.go must be removed after all usages disappear")
	}
}

// TestBugFB099_UnsupportedMenusAreDisabledByPrimaryKeyMigration
// 对应：docs/regression-tests.md #FB-099
func TestBugFB099_UnsupportedMenusAreDisabledByPrimaryKeyMigration(t *testing.T) {
	root := retiredModulesTestRoot(t)
	path := filepath.Join(root, "scripts", "sql", "system_002_retire_unsupported_modules.sql")
	source := readRetirementFile(t, path)
	lower := strings.ToLower(source)
	for _, forbidden := range []string{"delete ", "drop ", "truncate ", "parent_id in", "where parent_id"} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("migration contains forbidden operation/predicate: %s", forbidden)
		}
	}
	if !strings.Contains(lower, "begin;") || !strings.Contains(lower, "commit;") {
		t.Error("migration must be transactional")
	}
	if !strings.Contains(lower, "where menu_id in") || !strings.Contains(lower, "status = '1'") || !strings.Contains(lower, "visible = '1'") {
		t.Error("migration must disable and hide menus using the menu_id primary key")
	}
	for _, id := range []string{"109", "110", "113", "115", "1041", "1042", "1044", "1045", "1046", "1047", "1048", "1049", "1050", "1051", "1052", "1053", "1054", "1055", "1056", "1057", "1058", "1059", "1060"} {
		if !strings.Contains(source, id) {
			t.Errorf("verified target menu id is missing: %s", id)
		}
	}
}
