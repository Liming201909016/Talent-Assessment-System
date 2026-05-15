package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
)

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  false,
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Authorization", "Content-Disposition"},
		AllowCredentials: true,
	})
}

// SecurityHeaders 安全响应头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// Anonymous paths (no JWT required)
// 精确列出测评者/考生公开访问的路径，管理后台接口必须认证
var anonymousPrefixes = []string{
	// 系统公开
	"/login", "/register", "/captchaImage", "/logout",
	"/profile/", "/common/download",
	"/system/dict/",
	"/health",

	// 在线测评页面（前端路由）
	"/my/",

	// 测评者登录
	"/exam/api/tester/login",
	// FB-033: 测评者按身份证号查询（preview 页加载考生信息用）
	"/exam/api/tester/idNumber/",

	// 考生（开放模式）
	"/exam/api/candidate/save",
	"/exam/api/candidate/info",
	"/exam/api/candidate/update",
	"/exam/api/candidate/tester-info",
	"/exam/api/candidate/stand-score",
	"/exam/api/candidate/end-time",
	"/exam/api/candidate/pdf-persistence",
	"/exam/api/candidate/pdf-upload",

	// 答卷
	"/exam/api/paper/paper/",

	// MBTI 答题
	"/exam/api/mbti/paper-detail",
	"/exam/api/mbti/fill-answer",
	"/exam/api/mbti/submit",
	"/exam/api/mbti/score",
	"/exam/api/mbti/download-report",

	// 考试列表（测评者查看）
	"/exam/api/exam/exam/online-paging",
	"/exam/api/exam/exam/detail",

	// 测评者相关（答题/报告）
	"/exam/api/tester/stand-score",
	"/exam/api/tester/end-time",
	"/exam/api/tester/pdf-persistence",

	// Swagger / 其他
	"/swagger-ui", "/swagger-resources", "/v3/api-docs", "/v2/api-docs",
	"/druid/", "/doc.html", "/ws/",
}

func IsAnonymous(path string) bool {
	for _, p := range anonymousPrefixes {
		if len(path) >= len(p) && path[:len(p)] == p {
			return true
		}
	}
	return path == "/login" || path == "/logout" || path == "/register" || path == "/captchaImage" || path == "/health"
}

// IsAnonymousMethod 按 method+path 匹配的匿名端点（对齐 Java /exam/** 整段匿名）
// 仅放行考生答题流程必经的 PUT，避免管理后台接口暴露
func IsAnonymousMethod(method, path string) bool {
	// preview.vue 点击"开始测评"会 PUT /exam/api/tester 写回 paperId
	// Java SecurityConfig 中 /exam/** 整段匿名，Go 在此对齐
	if method == "PUT" && path == "/exam/api/tester" {
		return true
	}
	return false
}

// JWT 中间件
func JWT(cfg *config.Config, auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAnonymous(c.Request.URL.Path) || IsAnonymousMethod(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}
		header := c.GetHeader(cfg.Jwt.Header)
		lu, err := auth.ParseAuth(c.Request.Context(), header)
		if err != nil {
			response.AjaxUnauthorized(c, "")
			return
		}
		c.Set("loginUser", lu)
		c.Set("userId", lu.UserID)
		c.Next()
	}
}
