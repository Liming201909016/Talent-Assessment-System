package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestManagementTraitsExamStateHTTPBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body string
		user       *model.LoginUser
		want       int
	}{
		{"anonymous", `{"examId":"exam","state":"0"}`, nil, 401},
		{"ordinary", `{"examId":"exam","state":"0"}`, &model.LoginUser{UserID: 8, Permissions: []string{"exam:edit"}}, 403},
		{"missing state", `{"examId":"exam"}`, &model.LoginUser{UserID: 1}, 400},
		{"numeric state", `{"examId":"exam","state":0}`, &model.LoginUser{UserID: 1}, 400},
		{"invalid state", `{"examId":"exam","state":"2"}`, &model.LoginUser{UserID: 1}, 400},
		{"unknown field", `{"examId":"exam","state":"0","enabled":true}`, &model.LoginUser{UserID: 1}, 400},
		{"duplicate state", `{"examId":"exam","state":"0","state":"1"}`, &model.LoginUser{UserID: 1}, 400},
		{"valid enable reaches closed service", `{"examId":"exam","state":"0"}`, &model.LoginUser{UserID: 1}, 503},
		{"valid disable reaches closed service", `{"examId":"exam","state":"1"}`, &model.LoginUser{UserID: 1}, 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.user != nil {
					c.Set("loginUser", tc.user)
				}
				c.Next()
			})
			NewManagementTraitsRuntimeHandler(nil, &config.Config{}).RegisterRoutes(r.Group("/mng"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("POST", "/mng/admin/exam/state", strings.NewReader(tc.body)))
			if w.Code != tc.want {
				t.Fatalf("HTTP=%d want=%d body=%s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}
