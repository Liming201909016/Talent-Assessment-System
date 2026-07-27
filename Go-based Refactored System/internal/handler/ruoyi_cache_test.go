package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/redisx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func cacheTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	return db, mock
}

func cacheTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return server, client
}

func decodeAjaxData(t *testing.T, recorder *httptest.ResponseRecorder) any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid response: %v body=%s", err, recorder.Body.String())
	}
	if payload["code"] != float64(200) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
	return payload["data"]
}

// TestBugFB096_ConfigCacheReadThroughAndFailureFallback
// 对应：docs/regression-tests.md #FB-096
func TestBugFB096_ConfigCacheReadThroughAndFailureFallback(t *testing.T) {
	if got := configCacheKey("sys.account.registerUser"); got != redisx.SysConfigKey+"sys.account.registerUser" {
		t.Fatalf("config cache key=%q", got)
	}

	t.Run("cache hit avoids database", func(t *testing.T) {
		_, client := cacheTestRedis(t)
		old := redisx.Client
		redisx.Client = client
		t.Cleanup(func() { redisx.Client = old })
		if err := client.Set(context.Background(), configCacheKey("hit"), "cached", time.Hour).Err(); err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/system/config/configKey/hit", nil)
		c.Params = gin.Params{{Key: "configKey", Value: "hit"}}
		(&RuoYiSystemHandler{}).ConfigByKey(c)
		if got := decodeAjaxData(t, recorder); got != "cached" {
			t.Fatalf("data=%v", got)
		}
	})

	t.Run("miss reads database and caches one hour", func(t *testing.T) {
		server, client := cacheTestRedis(t)
		old := redisx.Client
		redisx.Client = client
		t.Cleanup(func() { redisx.Client = old })
		db, mock := cacheTestDB(t)
		mock.ExpectQuery("SELECT \\* FROM `sys_config` WHERE config_key = \\? ORDER BY `sys_config`.`config_id` LIMIT \\?").
			WithArgs("miss", 1).
			WillReturnRows(sqlmock.NewRows([]string{"config_id", "config_key", "config_value"}).AddRow(1, "miss", "database"))
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/system/config/configKey/miss", nil)
		c.Params = gin.Params{{Key: "configKey", Value: "miss"}}
		(&RuoYiSystemHandler{db: db}).ConfigByKey(c)
		if got := decodeAjaxData(t, recorder); got != "database" {
			t.Fatalf("data=%v", got)
		}
		if got, err := client.Get(context.Background(), configCacheKey("miss")).Result(); err != nil || got != "database" {
			t.Fatalf("cached=%q err=%v", got, err)
		}
		if ttl := server.TTL(configCacheKey("miss")); ttl != time.Hour {
			t.Fatalf("ttl=%s", ttl)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("redis failure falls back to database", func(t *testing.T) {
		old := redisx.Client
		redisx.Client = redis.NewClient(&redis.Options{Addr: "127.0.0.1:0", DialTimeout: time.Millisecond, ReadTimeout: time.Millisecond, WriteTimeout: time.Millisecond, MaxRetries: 0})
		t.Cleanup(func() { _ = redisx.Client.Close(); redisx.Client = old })
		db, mock := cacheTestDB(t)
		mock.ExpectQuery("SELECT \\* FROM `sys_config` WHERE config_key = \\? ORDER BY `sys_config`.`config_id` LIMIT \\?").
			WithArgs("fallback", 1).
			WillReturnRows(sqlmock.NewRows([]string{"config_id", "config_key", "config_value"}).AddRow(2, "fallback", "database"))
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/system/config/configKey/fallback", nil)
		c.Params = gin.Params{{Key: "configKey", Value: "fallback"}}
		(&RuoYiSystemHandler{db: db}).ConfigByKey(c)
		if got := decodeAjaxData(t, recorder); got != "database" {
			t.Fatalf("data=%v", got)
		}
	})

	t.Run("not found keeps empty string but other database errors are controlled", func(t *testing.T) {
		_, client := cacheTestRedis(t)
		old := redisx.Client
		redisx.Client = client
		t.Cleanup(func() { redisx.Client = old })
		for _, tc := range []struct {
			name string
			err  error
			code float64
			data any
		}{
			{name: "not found", err: gorm.ErrRecordNotFound, code: 200, data: ""},
			{name: "database failure", err: errors.New("db unavailable"), code: 500},
		} {
			t.Run(tc.name, func(t *testing.T) {
				db, mock := cacheTestDB(t)
				mock.ExpectQuery("SELECT \\* FROM `sys_config` WHERE config_key = \\? ORDER BY `sys_config`.`config_id` LIMIT \\?").
					WithArgs(tc.name, 1).WillReturnError(tc.err)
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodGet, "/system/config/configKey/test", nil)
				c.Params = gin.Params{{Key: "configKey", Value: tc.name}}
				(&RuoYiSystemHandler{db: db}).ConfigByKey(c)
				var payload map[string]any
				if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				if payload["code"] != tc.code || (tc.code == 200 && payload["data"] != tc.data) {
					t.Fatalf("response=%s", recorder.Body.String())
				}
			})
		}
	})
}

// TestBugFB097_CacheInvalidationRefreshAndPermissions
// 对应：docs/regression-tests.md #FB-097
func TestBugFB097_CacheInvalidationRefreshAndPermissions(t *testing.T) {
	if dictCacheKey("sys_user_sex") != redisx.SysDictKey+"sys_user_sex" {
		t.Fatal("dict cache key does not use SysDictKey")
	}

	t.Run("exact invalidation and prefix scan preserve unrelated keys", func(t *testing.T) {
		server, client := cacheTestRedis(t)
		ctx := context.Background()
		keys := map[string]string{
			configCacheKey("old"): "1", configCacheKey("new"): "2",
			dictCacheKey("old"): "3", dictCacheKey("new"): "4",
			redisx.LoginTokenKey + "token": "5", redisx.CaptchaKey + "uuid": "6",
		}
		for key, value := range keys {
			if err := client.Set(ctx, key, value, 0).Err(); err != nil {
				t.Fatal(err)
			}
		}
		if err := deleteCacheKeys(ctx, client, configCacheKey("old"), configCacheKey("new")); err != nil {
			t.Fatal(err)
		}
		if server.Exists(configCacheKey("old")) || server.Exists(configCacheKey("new")) || !server.Exists(dictCacheKey("old")) {
			t.Fatal("exact config invalidation deleted the wrong keys")
		}
		if err := scanDeletePrefix(ctx, client, redisx.SysDictKey, 1); err != nil {
			t.Fatal(err)
		}
		if server.Exists(dictCacheKey("old")) || server.Exists(dictCacheKey("new")) {
			t.Fatal("dict prefix refresh left matching keys")
		}
		if !server.Exists(redisx.LoginTokenKey+"token") || !server.Exists(redisx.CaptchaKey+"uuid") {
			t.Fatal("cache refresh deleted login or captcha keys")
		}
		if err := scanDeletePrefix(ctx, client, redisx.SysDictKey, 1); err != nil {
			t.Fatalf("empty prefix refresh failed: %v", err)
		}
	})

	t.Run("refresh redis failure is controlled", func(t *testing.T) {
		old := redisx.Client
		redisx.Client = redis.NewClient(&redis.Options{Addr: "127.0.0.1:0", DialTimeout: time.Millisecond, ReadTimeout: time.Millisecond, WriteTimeout: time.Millisecond, MaxRetries: 0})
		t.Cleanup(func() { _ = redisx.Client.Close(); redisx.Client = old })
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodDelete, "/system/config/refreshCache", nil)
		c.Set("loginUser", &model.LoginUser{UserID: 9, Permissions: []string{"system:config:edit"}})
		(&RuoYiSystemHandler{}).ConfigRefreshCache(c)
		var payload map[string]any
		_ = json.Unmarshal(recorder.Body.Bytes(), &payload)
		if payload["code"] != float64(500) {
			t.Fatalf("response=%s", recorder.Body.String())
		}
	})

	t.Run("empty dictionary miss returns array", func(t *testing.T) {
		_, client := cacheTestRedis(t)
		old := redisx.Client
		redisx.Client = client
		t.Cleanup(func() { redisx.Client = old })
		db, mock := cacheTestDB(t)
		mock.ExpectQuery("SELECT \\* FROM `sys_dict_data` WHERE dict_type = \\? AND status = '0' ORDER BY dict_sort").
			WithArgs("empty").WillReturnRows(sqlmock.NewRows([]string{"dict_code", "dict_type"}))
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/system/dict/data/type/empty", nil)
		c.Params = gin.Params{{Key: "dictType", Value: "empty"}}
		(&DictHandler{db: db}).DataByType(c)
		if !strings.Contains(recorder.Body.String(), `"data":[]`) {
			t.Fatalf("response=%s", recorder.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	for _, tc := range []struct {
		name string
		user *model.LoginUser
		perm string
		want bool
	}{
		{"config exact", &model.LoginUser{UserID: 9, Permissions: []string{"system:config:edit"}}, "system:config:edit", true},
		{"dict exact", &model.LoginUser{UserID: 9, Permissions: []string{"system:dict:edit"}}, "system:dict:edit", true},
		{"wildcard", &model.LoginUser{UserID: 9, Permissions: []string{"*:*:*"}}, "system:dict:edit", true},
		{"admin", &model.LoginUser{UserID: 9, Roles: []string{"admin"}}, "system:config:edit", true},
		{"unrelated", &model.LoginUser{UserID: 9, Permissions: []string{"system:user:edit"}}, "system:config:edit", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasSystemPermission(tc.user, tc.perm); got != tc.want {
				t.Fatalf("permission=%v want=%v", got, tc.want)
			}
		})
	}

	t.Run("edit and refresh handlers reject unrelated permission before dependencies", func(t *testing.T) {
		h := &RuoYiSystemHandler{}
		for _, tc := range []struct {
			name   string
			method string
			path   string
			body   string
			call   gin.HandlerFunc
		}{
			{"config edit", http.MethodPut, "/system/config", `{}`, h.ConfigEdit},
			{"dict edit", http.MethodPut, "/system/dict/type", `{}`, h.DictTypeEdit},
			{"config refresh", http.MethodDelete, "/system/config/refreshCache", ``, h.ConfigRefreshCache},
			{"dict refresh", http.MethodDelete, "/system/dict/type/refreshCache", ``, h.DictRefreshCache},
		} {
			t.Run(tc.name, func(t *testing.T) {
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Set("loginUser", &model.LoginUser{UserID: 9, Permissions: []string{"system:user:edit"}})
				tc.call(c)
				if recorder.Code != http.StatusForbidden {
					t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
				}
			})
		}
	})
}

func TestBugFB097_CacheRoutesAndMutationSourcesHaveNoStubOrKeys(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	handlerDir := filepath.Dir(current)
	routerRaw, err := os.ReadFile(filepath.Join(handlerDir, "..", "router", "router.go"))
	if err != nil {
		t.Fatal(err)
	}
	routerSource := string(routerRaw)
	for _, marker := range []string{
		`sys.DELETE("/config/refreshCache", ruoyiSysH.ConfigRefreshCache)`,
		`sys.DELETE("/dict/type/refreshCache", ruoyiSysH.DictRefreshCache)`,
	} {
		if !strings.Contains(routerSource, marker) {
			t.Errorf("real cache route missing: %s", marker)
		}
	}
	for _, marker := range []string{
		`sys.DELETE("/config/refreshCache", handler.AjaxStub`,
		`sys.DELETE("/dict/type/refreshCache", handler.AjaxStub`,
	} {
		if strings.Contains(routerSource, marker) {
			t.Errorf("cache route still uses stub: %s", marker)
		}
	}

	cacheRaw, err := os.ReadFile(filepath.Join(handlerDir, "ruoyi_cache.go"))
	if err != nil {
		t.Fatalf("cache implementation missing: %v", err)
	}
	cacheSource := string(cacheRaw)
	if !strings.Contains(cacheSource, ".Scan(") || strings.Contains(cacheSource, ".Keys(") {
		t.Fatal("cache refresh must use SCAN and must not use KEYS")
	}
	crudRaw, err := os.ReadFile(filepath.Join(handlerDir, "ruoyi_crud.go"))
	if err != nil {
		t.Fatal(err)
	}
	crudSource := string(crudRaw)
	for _, marker := range []string{
		"func (h *RuoYiSystemHandler) ConfigAdd", "func (h *RuoYiSystemHandler) ConfigEdit", "func (h *RuoYiSystemHandler) ConfigDelete",
		"func (h *RuoYiSystemHandler) DictTypeAdd", "func (h *RuoYiSystemHandler) DictTypeEdit", "func (h *RuoYiSystemHandler) DictTypeDelete",
		"func (h *RuoYiSystemHandler) DictDataAdd", "func (h *RuoYiSystemHandler) DictDataEdit", "func (h *RuoYiSystemHandler) DictDataDelete",
	} {
		start := strings.Index(crudSource, marker)
		if start < 0 {
			t.Fatalf("handler missing: %s", marker)
		}
		rest := crudSource[start+len(marker):]
		end := strings.Index(rest, "\nfunc ")
		if end >= 0 {
			rest = rest[:end]
		}
		if !strings.Contains(rest, "invalidate") {
			t.Errorf("mutation lacks exact cache invalidation: %s", marker)
		}
		if !strings.Contains(rest, "Error") && !strings.Contains(rest, "err") {
			t.Errorf("mutation lacks database error handling: %s", marker)
		}
	}
}
