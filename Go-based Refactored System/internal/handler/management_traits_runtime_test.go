package handler

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/middleware"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
)

func TestManagementTraitsRuntimeExactJWTExemption(t *testing.T) {
	cfg := &config.Config{}
	cfg.Jwt.Header = "Authorization"
	for _, suffix := range []string{"create-paper", "paper-detail", "fill-answer", "submit"} {
		path := "/exam/api/management-traits/participant/" + suffix
		for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
			for _, tail := range []string{"", "/extra"} {
				gin.SetMode(gin.TestMode)
				r := gin.New()
				r.Use(middleware.JWT(cfg, service.NewAuthService(cfg, nil, nil)))
				r.Handle(method, path+tail, func(c *gin.Context) { c.Status(204) })
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(method, path+tail, nil))
				want := 401
				if method == "POST" && tail == "" {
					want = 204
				}
				if w.Code != want {
					t.Fatalf("%s %s: %d want %d", method, path+tail, w.Code, want)
				}
			}
		}
	}
	if middleware.IsAnonymousMethod("POST", "/exam/api/management-traits/profile/freeze") {
		t.Fatal("admin anonymous")
	}
}

func TestManagementTraitsRuntimeParticipantHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.Jwt.Secret = "test-only-secret"
	h := NewManagementTraitsRuntimeHandler(nil, cfg)
	r := gin.New()
	h.RegisterRoutes(r.Group("/exam/api/management-traits"))
	claims := service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimePaperPurpose, ParticipantType: "candidate", ParticipantID: "owner", ExamID: "exam", PaperID: "paper", ExpiresAt: time.Now().Add(time.Hour).Unix()}
	token, err := service.CreateManagementTraitsRuntimeToken(cfg.Jwt.Secret, claims, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		body string
		want int
	}{
		{`{"paperId":"paper","paperQuestionId":"pq","optionId":"option"}`, 503},
		{`{"paperId":"foreign","paperQuestionId":"pq","optionId":"option"}`, 401},
		{`{"paperId":"paper","paperQuestionId":"pq","optionId":["option"]}`, 400},
		{`{"paperId":"paper","paperQuestionId":"pq","optionId":"option","answers":["option"]}`, 400},
		{`{"paperId":"paper","paperQuestionId":"pq","optionId":"one","optionId":"two"}`, 400},
		{`{"paperId":"paper","paperQuestionId":"pq","optionId":"option"} {}`, 400},
		{`{"paperId":"paper","paperQuestionId":"pq","optionId":""}`, 400},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/exam/api/management-traits/participant/fill-answer", bytes.NewBufferString(item.body))
		req.Header.Set("X-Management-Traits-Token", token)
		r.ServeHTTP(w, req)
		if w.Code != item.want || strings.Contains(w.Body.String(), token) || strings.Contains(w.Body.String(), cfg.Jwt.Secret) {
			t.Fatalf("status %d want %d", w.Code, item.want)
		}
	}
}

func TestManagementTraitsRuntimeAdminGate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, user := range []*model.LoginUser{nil, {UserID: 9, Permissions: []string{"exam:list"}}, {UserID: 9, Permissions: []string{"management-traits:result:read"}}, {UserID: 1}} {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			if user != nil {
				c.Set("loginUser", user)
			}
			c.Next()
		})
		NewManagementTraitsRuntimeHandler(nil, &config.Config{}).RegisterRoutes(r.Group("/exam/api/management-traits"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/exam/api/management-traits/results/detail?runId=run", nil))
		want := 403
		if user == nil {
			want = 401
		}
		if user != nil && user.UserID == 1 {
			want = 503
		}
		if w.Code != want {
			t.Fatalf("admin gate=%d want=%d", w.Code, want)
		}
	}
}
