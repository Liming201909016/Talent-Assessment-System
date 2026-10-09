package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
)

func TestManagementTraitsReportExactAdminRoutesAndNoPathInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, user := range []*model.LoginUser{nil, {UserID: 8, Permissions: []string{"exam:list"}}, {UserID: 1}} {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			if user != nil {
				c.Set("loginUser", user)
			}
			c.Next()
		})
		NewManagementTraitsRuntimeHandler(nil, &config.Config{}).RegisterRoutes(r.Group("/mng"))
		for _, x := range []struct {
			method, path, body string
			admin              int
		}{
			{"POST", "/mng/reports/generate-test", `{"runId":"run","formal":true}`, 400},
			{"POST", "/mng/reports/generate-test", `{"runId":"run","path":"/secret"}`, 400},
			{"GET", "/mng/reports/download?path=/secret", "", 400},
			{"GET", "/mng/reports/view?reportId=one&reportId=two", "", 400},
		} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(x.method, x.path, bytes.NewBufferString(x.body)))
			want := x.admin
			if user == nil {
				want = 401
			} else if user.UserID != 1 {
				want = 403
			}
			if w.Code != want {
				t.Fatalf("%s status=%d want=%d", x.path, w.Code, want)
			}
		}
	}
}

// MT-REPORT-DIAG-01: real local malformed template remains rejected by the
// renderer; fixed runtime sentinel still produces the original 409 envelope.
func TestBugMTReportDiag_RenderRejectAndHTTPContract(t *testing.T) {
	const secret = "INJECTED_RENDER_SECRET_TEMPLATE_TOKEN"
	if pdf, err := renderManagementTraitsTestWord([]byte(secret), service.ManagementTraitsTestReportData{}); err == nil || len(pdf) != 0 {
		t.Fatal("bad local Word fixture accepted")
	}
	gin.SetMode(gin.TestMode)
	for _, x := range []struct {
		name   string
		err    error
		status int
	}{
		{"invalid", service.ErrManagementTraitsRuntimeInvalid, 409},
		{"closed", service.ErrManagementTraitsRuntimeClosed, 503},
		{"wrapped-secret", fmt.Errorf("%s: %w", secret, service.ErrManagementTraitsRuntimeInvalid), 409},
		{"success", nil, 200},
	} {
		t.Run(x.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			managementTraitsRuntimeRespond(c, nil, x.err)
			var body map[string]any
			if w.Code != x.status || json.Unmarshal(w.Body.Bytes(), &body) != nil || strings.Contains(w.Body.String(), secret) || strings.Contains(w.Body.String(), "stage") || strings.Contains(w.Body.String(), "class") {
				t.Fatal("HTTP diagnostic disclosure or status changed")
			}
			if x.status == 409 && body["msg"] != "管理特质数据校验失败，请联系管理员" {
				t.Fatal("409 message changed")
			}
			if x.err == nil && (body["code"] != float64(0) || body["success"] != true) {
				t.Fatal("success envelope changed")
			}
		})
	}
}
