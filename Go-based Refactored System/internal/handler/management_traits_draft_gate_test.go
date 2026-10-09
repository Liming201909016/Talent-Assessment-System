package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
)

func draftGateMetadata(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}).AddRow("el_mng_exam_draft", "InnoDB"))
	draftHandlerSchema(mock)
}

func draftGateRow() model.ManagementTraitsExamDraft {
	now := time.Date(2026, 10, 7, 1, 0, 0, 0, time.UTC)
	return model.ManagementTraitsExamDraft{ExamID: "exam", RepoID: "repo", RepoCode: "00201", Lifecycle: "draft", CreatedAt: now, UpdatedAt: now}
}

func draftGateIdentityProbe(mock sqlmock.Sqlmock) {
	for _, table := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_result_dimension", "el_mng_result_module"} {
		mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WithArgs(table).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("el_mng_exam_draft"))
	draftGateMetadata(mock)
}

// MT-NEW-DRAFT: public service/actual Gin identity, zero DML expectations.
func TestBugManagementTraitsDraftIdentityNeverLegacy(t *testing.T) {
	for _, action := range []string{"lifecycle", "candidate", "tester"} {
		t.Run(action, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			draftGateIdentityProbe(mock)
			mock.ExpectQuery("SELECT .*el_mng_exam_draft").WillReturnRows(identityHandlerModelRows(t, draftGateRow()))
			if action == "lifecycle" {
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				state, err := service.ManagementTraitsExamLifecycle(context.Background(), db, "exam")
				if err != nil || state != "draft" {
					t.Fatal("draft metadata was inferred as legacy")
				}
			} else if action == "tester" {
				s := service.NewManagementTraitsRuntimeService(db, "synthetic-secret", 1<<20)
				identityHandlerRuntimeSchema(t, mock)
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(identityHandlerModelRows(t, model.Exam{ID: "exam", IsOpen: 2, AssessmentType: "legacy", ScoringMode: "legacy"}))
				mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
				mock.ExpectRollback()
				_, handled, err := s.TryTesterIdentity(context.Background(), "exam", "tester", "synthetic")
				if !handled || err == nil {
					t.Fatal("draft tester fell through")
				}
			} else {
				draftGateIdentityProbe(mock)
				mock.ExpectQuery("SELECT .*el_mng_exam_draft").WillReturnRows(identityHandlerModelRows(t, draftGateRow()))
				w := identityHandlerPost(NewCandidateHandler(db, &config.Config{}).Save, `{"examId":"exam","name":"Synthetic","telephone":"13800000000"}`)
				var result struct {
					Code int `json:"code"`
				}
				if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code == 0 {
					t.Fatal("draft candidate accepted")
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugManagementTraitsDraftLegacyRoutePreparation(t *testing.T) {
	for _, action := range []string{"prepare_admin", "prepare_list_admin", "prepare_low_privilege", "start", "global_list"} {
		t.Run(action, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("el_mng_exam_draft"))
			draftGateMetadata(mock)
			method, path, body := "POST", "/exam/api/tester", `{"examId":"exam","name":"Synthetic"}`
			if action == "prepare_list_admin" {
				method, path, body = "GET", "/exam/api/tester/list", ""
			}
			if action == "start" {
				path = "/exam/api/paper/paper/create-paper"
			}
			if action == "global_list" {
				method, path, body = "GET", "/exam/api/tester/list", ""
				mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_draft LIMIT 1").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			} else {
				mock.ExpectQuery("SELECT id, exam_id FROM el_paper WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}))
				mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
				mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_tester WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
				mock.ExpectQuery("SELECT .*el_mng_exam_draft").WillReturnRows(identityHandlerModelRows(t, draftGateRow()))
			}
			called := false
			r := gin.New()
			r.Use(ManagementTraitsLegacyScopeGuard(db))
			r.Use(func(c *gin.Context) {
				id := int64(1)
				if action == "prepare_low_privilege" {
					id = 2
				}
				c.Set("loginUser", &model.LoginUser{UserID: id})
				c.Next()
			})
			r.Use(ManagementTraitsDraftPreparationGuard())
			r.Handle(method, path, func(c *gin.Context) { called = true; c.Status(204) })
			requestPath := path
			if action == "prepare_list_admin" {
				requestPath += "?examId=exam"
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, requestPath, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if action == "prepare_admin" || action == "prepare_list_admin" {
				if !called || w.Code != 204 {
					t.Fatal("legitimate draft preparation blocked")
				}
			} else if called || w.Code != 403 {
				t.Fatal("draft legacy bypass")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
