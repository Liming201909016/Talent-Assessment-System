package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm/logger"
)

// UF053: real signed credential and existing Gin Submit route; DB is sqlmock.
// This verifies the HTTP-to-service attribution, not main MySQL contention.
func TestManagementTraitsRaceObservationHTTPBoundary(t *testing.T) {
	const paper = "f5e21ed0-4125-40a6-900a-a634881983ee"
	hash := sha256.Sum256([]byte("mng-race-paper-v1\x00" + paper))
	for _, scenario := range []string{"accepted-entry", "disabled", "missing-token", "foreign-paper", "client-timeout", "unknown-caller-field"} {
		t.Run(scenario, func(t *testing.T) {
			t.Setenv("MNG_RACE_OBSERVE_ENV", "local")
			t.Setenv("MNG_RACE_OBSERVE_PAPER_SHA256", hex.EncodeToString(hash[:]))
			if scenario == "disabled" {
				t.Setenv("MNG_RACE_OBSERVE_ENV", "")
			}
			var output bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
			t.Cleanup(func() { slog.SetDefault(previous) })
			db, mock := identityHandlerDB(t)
			db.Config.Logger = logger.Discard
			cfg := &config.Config{}
			cfg.Jwt.Secret = "UF053_TEST_ONLY_SECRET"
			svc := service.NewManagementTraitsRuntimeService(db, cfg.Jwt.Secret, 1<<20)
			h := NewManagementTraitsRuntimeHandler(db, cfg, svc)
			claims := service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimePaperPurpose, PaperID: paper, ExamID: "synthetic-exam", ParticipantID: "synthetic-owner", ParticipantType: "candidate", ExpiresAt: time.Now().Add(time.Hour).Unix()}
			token, err := service.CreateManagementTraitsRuntimeToken(cfg.Jwt.Secret, claims, time.Now())
			if err != nil {
				t.Fatal("synthetic token unavailable")
			}
			body := `{"paperId":"` + paper + `","submitType":"manual"}`
			want := 409
			switch scenario {
			case "accepted-entry", "disabled":
				identityHandlerRuntimeSchema(t, mock)
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnError(errors.New("PRIVATE_SQL_TOKEN_DSN_SENTINEL"))
				mock.ExpectRollback()
			case "missing-token":
				token, want = "", 401
			case "foreign-paper":
				body = `{"paperId":"foreign","submitType":"manual"}`
				want = 401
			case "client-timeout":
				body = `{"paperId":"` + paper + `","submitType":"timeout"}`
				want = 400
			case "unknown-caller-field":
				body = `{"paperId":"` + paper + `","submitType":"manual","caller":"expiry_worker"}`
				want = 400
			}
			gin.SetMode(gin.TestMode)
			r := gin.New()
			h.RegisterRoutes(r.Group("/exam/api/management-traits"))
			req := httptest.NewRequest("POST", "/exam/api/management-traits/participant/submit", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Management-Traits-Token", token)
			req.Header.Set("X-Caller", "expiry_worker")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != want || mock.ExpectationsWereMet() != nil {
				t.Fatal("HTTP/SQL contract changed")
			}
			if scenario != "accepted-entry" {
				if output.Len() != 0 {
					t.Fatal("rejected/disabled request observed")
				}
				return
			}
			lines := strings.Split(strings.TrimSpace(output.String()), "\n")
			if len(lines) != 5 {
				t.Fatal("HTTP stages missing")
			}
			for _, line := range lines {
				var event map[string]any
				if json.Unmarshal([]byte(line), &event) != nil || len(event) != 12 || event["caller"] != "http_participant" || event["outcome"] != "failed" {
					t.Fatal("unsafe HTTP attribution")
				}
			}
			for _, secret := range []string{paper, token, cfg.Jwt.Secret, claims.ParticipantID, claims.ExamID, "PRIVATE_SQL_TOKEN_DSN_SENTINEL"} {
				if strings.Contains(output.String(), secret) || strings.Contains(w.Body.String(), secret) {
					t.Fatal("private HTTP data emitted")
				}
			}
		})
	}
}
