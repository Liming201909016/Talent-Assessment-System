package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestBugFB045_ParticipantCannotClaimTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CompetencyRuntimeHandler{cfg: &config.Config{}}
	router := gin.New()
	router.POST("/submit", h.Submit)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(`{"paperId":"p1","submitType":"timeout"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "超时由系统判定") {
		t.Fatalf("response = %s", w.Body.String())
	}
}

func TestCompetencyInternalReportData_RequiresConfiguredMatchingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name, configured, supplied string
		want                       int
	}{
		{"empty configuration rejects empty", "", "", http.StatusUnauthorized},
		{"wrong token", "expected", "wrong", http.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &CompetencyRuntimeHandler{cfg: &config.Config{PdfGen: config.PdfGenCfg{InternalToken: tt.configured}}}
			r := gin.New()
			r.GET("/report", h.InternalReportData)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/report", nil)
			if tt.supplied != "" {
				req.Header.Set("X-Internal-Token", tt.supplied)
			}
			r.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

// TestBugFB079_InternalReportRejectsTokenQueryParameter
// 对应：docs/regression-tests.md #FB-079
// 复现：内部报告令牌通过 query 传输并进入 nginx access log。
// 期望：query token 即使正确也返回 401，仅接受 X-Internal-Token 请求头。
func TestBugFB079_InternalReportRejectsTokenQueryParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CompetencyRuntimeHandler{cfg: &config.Config{PdfGen: config.PdfGenCfg{InternalToken: "expected"}}}
	router := gin.New()
	router.GET("/report", h.InternalReportData)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/report?paperId=paper-1&token=expected", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

// TestBugFB078_CompetencyResultEndpointsRequireExamPermission
// 对应：docs/regression-tests.md #FB-078
// 复现：任意已登录但无 exam:list/exam:export 权限的后台账号可读取结果和报告数据。
// 期望：分页、详情和管理员报告数据均在调用 service 前返回 HTTP 403。
func TestBugFB078_CompetencyResultEndpointsRequireExamPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CompetencyRuntimeHandler{cfg: &config.Config{}}
	for _, tt := range []struct {
		name, method, path, body string
		handler                  gin.HandlerFunc
	}{
		{"paging", http.MethodPost, "/results", `{"examId":"exam-1"}`, h.ResultsPaging},
		{"detail", http.MethodPost, "/detail", `{"paperId":"paper-1"}`, h.ResultDetail},
		{"admin report data", http.MethodGet, "/report?paperId=paper-1", "", h.AdminReportData},
	} {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("loginUser", &model.LoginUser{UserID: 99, Permissions: []string{"system:user:list"}})
				c.Next()
			})
			router.Handle(tt.method, strings.Split(tt.path, "?")[0], tt.handler)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestCompetencyResultAccess_UsesExistingExamPermissionContract(t *testing.T) {
	for _, tt := range []struct {
		name  string
		login *model.LoginUser
		want  bool
	}{
		{"missing login", nil, false},
		{"administrator", &model.LoginUser{UserID: 1}, true},
		{"global permission", &model.LoginUser{UserID: 99, Permissions: []string{"*:*:*"}}, true},
		{"exam list", &model.LoginUser{UserID: 99, Permissions: []string{"exam:list"}}, true},
		{"exam export", &model.LoginUser{UserID: 99, Permissions: []string{"exam:export"}}, true},
		{"unrelated permission", &model.LoginUser{UserID: 99, Permissions: []string{"system:user:list"}}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := canAccessExamResults(tt.login); got != tt.want {
				t.Fatalf("canAccessExamResults()=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompetencyRuntimeRoutes(t *testing.T) {
	src := readSourceFile(t, "../router/router.go")
	for _, required := range []string{
		`POST("/publish", competencyRuntimeH.Publish)`,
		`POST("/create-paper", competencyRuntimeH.CreatePaper)`,
		`POST("/paper-detail", competencyRuntimeH.PaperDetail)`,
		`POST("/fill-answer", competencyRuntimeH.FillAnswer)`,
		`POST("/submit", competencyRuntimeH.Submit)`,
		`POST("/paging", competencyRuntimeH.ResultsPaging)`,
		`POST("/detail", competencyRuntimeH.ResultDetail)`,
		`GET("/competency/internal/report-data", competencyRuntimeH.InternalReportData)`,
	} {
		if !strings.Contains(src, required) {
			t.Errorf("router missing %q", required)
		}
	}
}

func TestCompetencyResultsPaging_AcceptsSafeSortContract(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "competency_runtime.go"), "func (h *CompetencyRuntimeHandler) ResultsPaging(")
	for _, required := range []string{"SortBy", "SortDirection", "DimensionID", "Name", "Telephone", "Completion", "CompetencyResultPageRequest"} {
		if !strings.Contains(body, required) {
			t.Errorf("ResultsPaging missing %q", required)
		}
	}
}

// TestBugFB083_ResultsPagingRejectsMalformedJSON
// 对应：docs/regression-tests.md #FB-083
// 复现：分页接口忽略 ShouldBindJSON 错误并使用部分零值继续进入 service。
// 期望：畸形 JSON 在数据库查询前返回受控参数错误。
func TestBugFB083_ResultsPagingRejectsMalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CompetencyRuntimeHandler{cfg: &config.Config{}}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("loginUser", &model.LoginUser{UserID: 1})
		c.Next()
	})
	router.POST("/results", h.ResultsPaging)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/results", bytes.NewBufferString(`{"examId":`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "参数格式错误") {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestCompetencyResultManagementFrontend_IsReachable(t *testing.T) {
	router := readSourceFile(t, "../../ruoyi-ui/src/router/index.js")
	index := readSourceFile(t, "../../ruoyi-ui/src/views/exam/exam/index.vue")
	for _, required := range []string{"CompetencyResults", "competency-results"} {
		if !strings.Contains(router, required) {
			t.Errorf("frontend router missing %q", required)
		}
	}
	for _, required := range []string{"competencyResults", "handleCompetencyResults", "assessmentType === 'competency'"} {
		if !strings.Contains(index, required) {
			t.Errorf("exam list missing competency result entry %q", required)
		}
	}
}

func TestCompetencyParticipantResponsesDoNotExposeScores(t *testing.T) {
	src := readSourceFile(t, "competency_runtime.go")
	for _, sig := range []string{"func (h *CompetencyRuntimeHandler) CreatePaper(", "func (h *CompetencyRuntimeHandler) PaperDetail(", "func (h *CompetencyRuntimeHandler) FillAnswer(", "func (h *CompetencyRuntimeHandler) Submit("} {
		body := extractFunctionBody(t, src, sig)
		for _, forbidden := range []string{"OverallScore", "DimensionScore", "FinalScore", "ScoringDirection"} {
			if strings.Contains(body, forbidden) {
				t.Errorf("%s exposes %s", sig, forbidden)
			}
		}
	}
}

func TestCompetencyFormalReportRoutesAndPdfImplementation(t *testing.T) {
	router := readSourceFile(t, "../router/router.go")
	for _, required := range []string{
		`POST("/generate", competencyReportH.Generate)`,
		`GET("/download", competencyReportH.Download)`,
	} {
		if !strings.Contains(router, required) {
			t.Errorf("router missing %q", required)
		}
	}
	source := readSourceFile(t, "competency_report.go")
	for _, required := range []string{
		"FormalReportData", "GeneratePDF", "pdf_sha256", "Content-Disposition",
		"filepath.Rel", "writeReportAudit",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("competency report implementation missing %q", required)
		}
	}
	serviceSource := readSourceFile(t, "../service/competency_runtime.go") + readSourceFile(t, "../service/competency_report.go")
	for _, required := range []string{"BuildCompetencyReportTextSnapshot", "CompetencyTemporaryDisclaimer", "CompetencyReportText"} {
		if !strings.Contains(serviceSource, required) {
			t.Errorf("competency report service missing %q", required)
		}
	}
}
