package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
)

// ============================================================
// 回归测试 — FB-033 移动端预览页 401 异常
// 对应：docs/regression-tests.md FB-033
//
// 用户反馈：移动端打开 MBTI 预览页报"系统接口401异常"
// 根因：preview.vue 调用 /exam/api/tester/idNumber/{idNumber}
//       未在 anonymousPrefixes 列表中，考生无 token → 401
// ============================================================

func TestBugFB033_TesterIdNumberIsAnonymous(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"FB-033: tester/idNumber/xxx 必须匿名（preview 页要查考生信息）",
			"/exam/api/tester/idNumber/411111199001011234", true},
		{"FB-033: tester/idNumber 带 query 参数",
			"/exam/api/tester/idNumber/test?examId=x", true},

		// 现有规则不能被破坏
		{"login 必须匿名", "/login", true},
		{"captchaImage 必须匿名", "/captchaImage", true},
		{"exam/exam/detail 必须匿名（preview 用）", "/exam/api/exam/exam/detail", true},
		{"mbti/paper-detail 必须匿名", "/exam/api/mbti/paper-detail", true},
		{"tester/login 必须匿名", "/exam/api/tester/login", true},

		// 反例：管理类接口必须需要 token
		{"tester/list 不能匿名", "/exam/api/tester/list", false},
		{"qu/save 不能匿名", "/exam/api/qu/qu/save", false},
		{"system/user/list 不能匿名", "/system/user/list", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAnonymous(tt.path)
			if got != tt.want {
				t.Errorf("IsAnonymous(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCompetencyParticipantRoutesUseExactMethodAndPath(t *testing.T) {
	participantPaths := []string{
		"/exam/api/competency/participant/create-paper",
		"/exam/api/competency/participant/paper-detail",
		"/exam/api/competency/participant/fill-answer",
		"/exam/api/competency/participant/submit",
	}
	for _, path := range participantPaths {
		t.Run("POST "+path, func(t *testing.T) {
			if !IsAnonymousMethod("POST", path) {
				t.Fatalf("POST %s must pass JWT middleware for participant flow", path)
			}
		})
		for _, method := range []string{"GET", "PUT", "DELETE", "PATCH"} {
			t.Run(method+" "+path, func(t *testing.T) {
				if IsAnonymousMethod(method, path) {
					t.Fatalf("%s %s must not be anonymous", method, path)
				}
			})
		}
		t.Run("POST suffix "+path, func(t *testing.T) {
			if IsAnonymousMethod("POST", path+"/admin") {
				t.Fatalf("POST %s/admin must not match by prefix", path)
			}
		})
	}
}

func TestCompetencyManagementRoutesRequireAdminJWT(t *testing.T) {
	paths := []string{
		"/exam/api/competency/dimensions/paging",
		"/exam/api/competency/questions/paging",
		"/exam/api/competency/exams/publish",
		"/exam/api/competency/results/paging",
		"/exam/api/competency/results/detail",
		"/exam/api/competency/admin/report-data",
	}
	for _, path := range paths {
		if IsAnonymous(path) || IsAnonymousMethod("POST", path) {
			t.Errorf("management route must require admin JWT: %s", path)
		}
	}
}

// TestBugFB101_DictManagementUsesJWTWhileReadEndpointsStayAnonymous
// 对应：docs/regression-tests.md #FB-101
// 复现：`/system/dict/` 宽泛匿名前缀让字典管理请求跳过 JWT，handler 因无登录上下文固定返回 403。
// 期望：仅公开字典读取端点匿名，type/data 管理端点必须进入 JWT 中间件。
func TestBugFB101_DictManagementUsesJWTWhileReadEndpointsStayAnonymous(t *testing.T) {
	for _, path := range []string{
		"/system/dict/type",
		"/system/dict/type/list",
		"/system/dict/data",
		"/system/dict/data/list",
	} {
		if IsAnonymous(path) || IsAnonymousMethod(http.MethodPost, path) {
			t.Errorf("dict management route must require JWT: %s", path)
		}
	}
	if !IsAnonymousMethod(http.MethodGet, "/system/dict/data/type/sys_user_sex") {
		t.Error("public dictionary lookup must remain anonymous")
	}
	if !IsAnonymousMethod(http.MethodPost, "/system/dict/data/batch") {
		t.Error("public dictionary batch lookup must remain anonymous")
	}
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/system/dict/data/type/sys_user_sex"},
		{http.MethodGet, "/system/dict/data/type/"},
		{http.MethodGet, "/system/dict/data/batch"},
		{http.MethodPost, "/system/dict/data/batch/extra"},
	} {
		if IsAnonymousMethod(tc.method, tc.path) {
			t.Errorf("unexpected anonymous dictionary route: %s %s", tc.method, tc.path)
		}
	}
}

func TestCompetencyInternalReportRouteIsExact(t *testing.T) {
	path := "/exam/api/competency/internal/report-data"
	if !IsAnonymousMethod("GET", path) {
		t.Fatal("internal report route must pass JWT middleware for Chromedp and validate its internal token in the handler")
	}
	for _, tc := range []struct {
		method string
		path   string
	}{
		{"POST", path},
		{"GET", path + "/extra"},
		{"GET", "/exam/api/competency/internal"},
	} {
		if IsAnonymousMethod(tc.method, tc.path) {
			t.Errorf("unexpected anonymous internal report route: %s %s", tc.method, tc.path)
		}
	}
}

func TestCompetencyAnonymousRouting_HTTPMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{Jwt: config.JwtCfg{Header: "Authorization"}}
	auth := service.NewAuthService(cfg, nil, nil)
	router := gin.New()
	router.Use(JWT(cfg, auth))
	router.POST("/exam/api/competency/participant/create-paper", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.POST("/exam/api/competency/participant/create-paper/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	allowed := httptest.NewRecorder()
	allowedReq := httptest.NewRequest(http.MethodPost, "/exam/api/competency/participant/create-paper", nil)
	router.ServeHTTP(allowed, allowedReq)
	if allowed.Code != http.StatusNoContent {
		t.Fatalf("exact participant route status = %d, want %d", allowed.Code, http.StatusNoContent)
	}

	blocked := httptest.NewRecorder()
	blockedReq := httptest.NewRequest(http.MethodPost, "/exam/api/competency/participant/create-paper/admin", nil)
	router.ServeHTTP(blocked, blockedReq)
	if blocked.Code != http.StatusUnauthorized {
		t.Fatalf("suffix route status = %d, want %d", blocked.Code, http.StatusUnauthorized)
	}
}
