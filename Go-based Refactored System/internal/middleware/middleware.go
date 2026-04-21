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
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
	})
}

// Anonymous paths (no JWT required) — 对齐 Java SecurityConfig
var anonymousPrefixes = []string{
	"/login", "/register", "/captchaImage", "/logout",
	"/my/", "/exam/",
	"/system/dict/",
	"/profile/", "/common/download",
	"/swagger-ui", "/swagger-resources", "/v3/api-docs", "/v2/api-docs",
	"/druid/", "/doc.html", "/ws/",
	"/health",
}

func IsAnonymous(path string) bool {
	for _, p := range anonymousPrefixes {
		if len(path) >= len(p) && path[:len(p)] == p {
			return true
		}
	}
	return path == "/login" || path == "/logout" || path == "/register" || path == "/captchaImage" || path == "/health"
}

// JWT 中间件
func JWT(cfg *config.Config, auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAnonymous(c.Request.URL.Path) {
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
