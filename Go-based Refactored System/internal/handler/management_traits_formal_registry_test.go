package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestManagementTraitsFormalHTTPBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, method, path, body string
		user                     *model.LoginUser
		status                   int
	}{
		{"anonymous", "GET", "/versions", "", nil, 401},
		{"ordinary_exam_permission", "GET", "/versions", "", &model.LoginUser{UserID: 2, Permissions: []string{"exam:list"}}, 403},
		{"zero_id_wildcard", "GET", "/versions", "", &model.LoginUser{Permissions: []string{"*:*:*"}}, 403},
		{"admin_closed", "GET", "/versions", "", &model.LoginUser{UserID: 1}, 503},
		{"wildcard_closed", "GET", "/versions", "", &model.LoginUser{UserID: 7, Permissions: []string{"*:*:*"}}, 503},
		{"arbitrary_path", "POST", "/versions/register", `{"versionCode":"v1","examId":"exam","assetKey":"v1","path":"C:/private"}`, &model.LoginUser{UserID: 1}, 400},
		{"forged_actor", "POST", "/versions/approve", `{"versionId":"v1","expectedIdentitySha":"a","expectedEpoch":"1","kind":"content","actorId":"1"}`, &model.LoginUser{UserID: 1}, 400},
		{"forged_time", "POST", "/versions/approve", `{"versionId":"v1","expectedIdentitySha":"a","expectedEpoch":"1","kind":"content","approvedAt":"2026-10-01"}`, &model.LoginUser{UserID: 1}, 400},
		{"duplicate", "POST", "/versions/register", `{"versionCode":"v1","examId":"exam","assetKey":"v1","assetKey":"v2"}`, &model.LoginUser{UserID: 1}, 400},
		{"null", "POST", "/versions/register", `{"versionCode":"v1","examId":null,"assetKey":"v1"}`, &model.LoginUser{UserID: 1}, 400},
		{"empty", "POST", "/versions/register", `{"versionCode":"v1","examId":"","assetKey":"v1"}`, &model.LoginUser{UserID: 1}, 400},
		{"invalid_kind", "POST", "/versions/approve", `{"versionId":"v1","expectedIdentitySha":"a","expectedEpoch":"1","kind":"both"}`, &model.LoginUser{UserID: 1}, 400},
		{"invalid_epoch", "POST", "/versions/activate", `{"versionId":"v1","expectedIdentitySha":"a","expectedEpoch":"01"}`, &model.LoginUser{UserID: 1}, 400},
		{"download_not_installed", "GET", "/reports/download", "", &model.LoginUser{UserID: 1}, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.user != nil {
					c.Set("loginUser", tc.user)
				}
			})
			NewManagementTraitsFormalRegistryHandler(nil).RegisterRoutes(r.Group(""))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if w.Code != tc.status {
				t.Fatalf("status %d, want %d", w.Code, tc.status)
			}
			if strings.Contains(w.Body.String(), "C:/private") {
				t.Fatal("path disclosed")
			}
		})
	}
}
