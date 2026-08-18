package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestPhase1WordTemplateManagementRequiresAdministrator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	templatePath := "../../configs/export-templates/competency-phase1-report.docx"
	handler := &CompetencyReportHandler{examH: &ExamHandler{cfg: &config.Config{Phase1WordReport: config.Phase1WordReportCfg{TemplatePath: templatePath}}}}
	for _, test := range []struct {
		name       string
		login      *model.LoginUser
		wantStatus int
	}{
		{name: "administrator", login: &model.LoginUser{UserID: 1}, wantStatus: http.StatusOK},
		{name: "global permission", login: &model.LoginUser{UserID: 99, Permissions: []string{"*:*:*"}}, wantStatus: http.StatusOK},
		{name: "exam list is insufficient", login: &model.LoginUser{UserID: 99, Permissions: []string{"exam:list"}}, wantStatus: http.StatusForbidden},
		{name: "unrelated permission", login: &model.LoginUser{UserID: 99, Permissions: []string{"system:user:list"}}, wantStatus: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("loginUser", test.login)
				c.Next()
			})
			router.GET("/template", handler.Phase1TemplateInfo)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/template", nil)
			router.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestPhase1WordTemplateManagementRoutes(t *testing.T) {
	source := readSourceFile(t, "../router/router.go")
	for _, required := range []string{
		`GET("/template", competencyReportH.Phase1TemplateInfo)`,
		`GET("/template/download", competencyReportH.DownloadPhase1Template)`,
		`POST("/template/upload", competencyReportH.UploadPhase1Template)`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("template management route missing: %s", required)
		}
	}
}

func TestBugFB122_TemplateDownloadDisablesBrowserCaching(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &CompetencyReportHandler{examH: &ExamHandler{cfg: &config.Config{Phase1WordReport: config.Phase1WordReportCfg{
		TemplatePath: "../../configs/export-templates/competency-phase1-report.docx",
	}}}}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("loginUser", &model.LoginUser{UserID: 1})
		c.Next()
	})
	router.GET("/template/download", handler.DownloadPhase1Template)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/template/download", nil)
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store, no-cache, must-revalidate" {
		t.Fatalf("Cache-Control=%q", got)
	}
	if got := response.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("Pragma=%q", got)
	}
}
