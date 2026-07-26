package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
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
	for _, required := range []string{"SortBy", "SortDirection", "DimensionID", "CompetencyResultPageRequest"} {
		if !strings.Contains(body, required) {
			t.Errorf("ResultsPaging missing %q", required)
		}
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
