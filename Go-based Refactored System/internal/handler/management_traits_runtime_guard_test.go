package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
)

func legacyGuardTables(mock sqlmock.Sqlmock, present bool) {
	rows := sqlmock.NewRows([]string{"table_name"})
	if present {
		for _, name := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_paper_question_snapshot", "el_mng_result_run", "el_mng_result_dimension", "el_mng_result_module"} {
			rows.AddRow(name)
		}
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(rows)
}

func TestManagementTraitsLegacyGuardRouteMatrix(t *testing.T) {
	for _, route := range managementTraitsLegacyRoutes {
		t.Run(route.method+route.path, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			legacyGuardTables(mock, true)
			path := strings.ReplaceAll(route.path, ":ids", "protected-person")
			path = strings.ReplaceAll(path, ":idNumber", "protected-identifier")
			path = strings.ReplaceAll(path, ":id", "protected-person")
			body := `{"id":"protected-person","examId":"new-exam","paperId":"new-paper","ids":["old","new"],"idNumber":"protected-identifier","telephone":"13800000000","file":"old.pdf","quList":[{"id":"question"}]}`
			if route.identity {
				body = `{"examId":"new-exam","idNumber":"person","name":"Person","telephone":"13800000000"}`
			}
			request := httptest.NewRequest(route.method, path, strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			if !route.identity {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = request
				target, err := managementTraitsLegacyTargets(c, route)
				if err != nil {
					t.Fatal(err)
				}
				if target.AllLegacy {
					mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile LIMIT 1").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
				} else {
					if len(target.PaperQuestionIDs) > 0 {
						mock.ExpectQuery("SELECT paper_id FROM el_paper_qu WHERE").WillReturnRows(sqlmock.NewRows([]string{"paper_id"}))
					}
					if len(target.PaperIDs)+len(target.ExamIDs) > 0 {
						mock.ExpectQuery("SELECT id, exam_id FROM el_paper WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}))
					}
					for _, item := range []struct {
						table string
						n     int
					}{{"el_candidate", len(target.CandidateIDs) + len(target.PaperIDs) + len(target.ExamIDs) + len(target.CandidateTelephones) + len(target.PDFPaths)}, {"el_tester", len(target.TesterIDs) + len(target.PaperIDs) + len(target.ExamIDs) + len(target.TesterIdentifiers) + len(target.PDFPaths)}} {
						if item.n > 0 {
							mock.ExpectQuery("SELECT id, exam_id, paper_id FROM " + item.table + " WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
						}
					}
					if len(target.ExamIDs) > 0 {
						mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
					} else {
						mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_paper_snapshot WHERE").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
					}
				}
			}
			called := false
			r := gin.New()
			r.Use(ManagementTraitsLegacyScopeGuard(db))
			r.Handle(route.method, route.path, func(c *gin.Context) { called = true; c.Status(204) })
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			if route.identity {
				if !called || w.Code != 204 {
					t.Fatal("identity safe branch not delegated", w.Code, w.Body.String())
				}
			} else if called || w.Code != 403 {
				t.Fatal("protected old endpoint reached handler", w.Code, w.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyGuardAbsentAndFailure(t *testing.T) {
	for _, failure := range []bool{false, true} {
		db, mock := identityHandlerDB(t)
		if failure {
			mock.ExpectQuery("information_schema.tables").WillReturnError(errors.New("PRIVATE_DB_SECRET"))
		} else {
			legacyGuardTables(mock, false)
		}
		called := false
		r := gin.New()
		r.Use(ManagementTraitsLegacyScopeGuard(db))
		r.POST("/exam/api/paper/paper/save", func(c *gin.Context) { called = true; c.Status(204) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/exam/api/paper/paper/save", strings.NewReader(`{"id":"legacy","examId":"legacy-exam"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if called == failure || strings.Contains(w.Body.String(), "PRIVATE") {
			t.Fatal(called, w.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagementTraitsLegacyGuardBodyRestored(t *testing.T) {
	cases := []struct{ content, body string }{{"application/json", `{"ID":"old","examId":"new","paperId":"paper","quList":[{"id":"foreign-q"}]}`}, {"application/x-www-form-urlencoded", "id=old&examId=new&paperId=paper"}}
	var b bytes.Buffer
	m := multipart.NewWriter(&b)
	_ = m.WriteField("paperId", "paper")
	_ = m.WriteField("examId", "new")
	f, _ := m.CreateFormFile("file", "test.pdf")
	_, _ = f.Write([]byte("%PDF-test"))
	_ = m.Close()
	cases = append(cases, struct{ content, body string }{m.FormDataContentType(), b.String()})
	for _, tc := range cases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/exam/api/paper/paper/review-paper?id=query", strings.NewReader(tc.body))
		c.Request.Header.Set("Content-Type", tc.content)
		target, err := managementTraitsLegacyTargets(c, managementTraitsLegacyRoute{method: "POST", path: "/exam/api/paper/paper/review-paper", kind: "paper"})
		if err != nil || len(target.ExamIDs) != 1 || len(target.PaperIDs) < 2 {
			t.Fatal(target, err)
		}
		raw, _ := io.ReadAll(c.Request.Body)
		if string(raw) != tc.body {
			t.Fatal("body not restored")
		}
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
	}
}

func TestManagementTraitsLegacyGuardRejectsAmbiguousJSON(t *testing.T) {
	for _, body := range []string{`{"id":"old","ID":"new"}`, `{"examId":"old","exam\u0049d":"new"}`, `{"id":12}`, `{"ids":["old",null]}`, `{"id":"old"} {"id":"new"}`} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/exam/api/paper/paper/save", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		if _, err := managementTraitsLegacyTargets(c, managementTraitsLegacyRoute{kind: "paper"}); err == nil {
			t.Fatal("ambiguous ID accepted", body)
		}
	}
}

func TestManagementTraitsLegacyGuardHoldsFreezeThroughResponse(t *testing.T) {
	db, mock := identityHandlerDB(t)
	legacyGuardTables(mock, false)
	entered, finish, returned := make(chan struct{}), make(chan struct{}), make(chan struct{})
	r := gin.New()
	r.Use(ManagementTraitsLegacyScopeGuard(db))
	r.POST("/exam/api/paper/paper/save", func(c *gin.Context) { close(entered); <-finish; c.Status(204) })
	go func() {
		req := httptest.NewRequest("POST", "/exam/api/paper/paper/save", strings.NewReader(`{"id":"legacy"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
		close(returned)
	}()
	<-entered
	frozen := make(chan func(), 1)
	go func() { frozen <- service.LockManagementTraitsRuntimeFreeze() }()
	select {
	case release := <-frozen:
		release()
		t.Fatal("freeze passed in-flight handler")
	case <-time.After(20 * time.Millisecond):
	}
	close(finish)
	<-returned
	select {
	case release := <-frozen:
		release()
	case <-time.After(time.Second):
		t.Fatal("freeze not released")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyGuardOldCompressionCannotEscape(t *testing.T) {
	for _, file := range []string{"candidate.go", "tester_score.go", "exam_report_gen.go"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "go compressPDF(saved)") {
			t.Error("HTTP compression escapes freeze gate", file)
		}
	}
}

func TestManagementTraitsLegacyGuardUnrelatedRoutesDoNotProbe(t *testing.T) {
	db, mock := identityHandlerDB(t)
	r := gin.New()
	r.Use(ManagementTraitsLegacyScopeGuard(db))
	r.POST("/exam/api/qu/qu/save", func(c *gin.Context) { c.Status(204) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/exam/api/qu/qu/save", nil))
	if w.Code != 204 {
		t.Fatal(w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyGuardNullableLegacyDTO(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("PUT", "/exam/api/tester", strings.NewReader(`{"id":"tester","examId":"old-exam","idNumber":null,"telephone":"13800000000","paperId":null}`))
	c.Request.Header.Set("Content-Type", "application/json")
	target, err := managementTraitsLegacyTargets(c, managementTraitsLegacyRoute{kind: "tester"})
	if err != nil || len(target.TesterIDs) != 1 || len(target.PaperIDs) != 0 {
		t.Fatal("valid legacy nulls rejected", target, err)
	}
}

func TestManagementTraitsLegacyGuardReadChildDoesNotDeadlockQueuedFreeze(t *testing.T) {
	parent := service.LockManagementTraitsLegacyMutation()
	frozen := make(chan func(), 1)
	go func() { frozen <- service.LockManagementTraitsRuntimeFreeze() }()
	db, mock := identityHandlerDB(t)
	legacyGuardTables(mock, false)
	r := gin.New()
	r.Use(ManagementTraitsLegacyScopeGuard(db))
	r.POST("/exam/api/paper/paper/stand-score", func(c *gin.Context) { c.Status(204) })
	done := make(chan int, 1)
	go func() {
		req := httptest.NewRequest("POST", "/exam/api/paper/paper/stand-score", strings.NewReader(`{"id":"old"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		done <- w.Code
	}()
	select {
	case code := <-done:
		if code != 204 {
			t.Error(code)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("renderer read waits for freeze, freeze waits for parent")
	}
	parent.Release()
	release := <-frozen
	release()
	select {
	case <-done:
	case <-time.After(time.Second):
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyGuardCollectionActualFilters(t *testing.T) {
	for _, tc := range []struct {
		method, path, body string
		broad              bool
	}{
		{"GET", "/exam/api/tester", "", true},
		{"GET", "/exam/api/tester/list", "", true},
		{"GET", "/exam/api/tester/tester-list", "", true},
		{"GET", "/exam/api/tester/list?examId=legacy", "", false},
		{"POST", "/exam/api/candidate/tester-list?examId=ignored", `{}`, true},
		{"POST", "/exam/api/candidate/tester-list", `{"examId":"legacy"}`, false},
		{"POST", "/exam/api/paper/paper/paging", `{"params":{"examId":"legacy"}}`, false},
		{"POST", "/exam/api/paper/paper/paging?examId=ignored", `{"examId":"ignored","params":{}}`, true},
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		c.Request.Header.Set("Content-Type", "application/json")
		route, ok := managementTraitsLegacyMatch(tc.method, c.Request.URL.Path)
		if !ok {
			t.Error("collection missing", tc.path)
			continue
		}
		target, err := managementTraitsLegacyTargets(c, route)
		if err != nil || target.AllLegacy != tc.broad {
			t.Error(tc.path, target, err)
		}
	}
}

func TestManagementTraitsLegacyGuardMBTIBackgroundDrains(t *testing.T) {
	db, mock := identityHandlerDB(t)
	legacyGuardTables(mock, false)
	started, finish, returned := make(chan struct{}), make(chan struct{}), make(chan struct{})
	r := gin.New()
	r.Use(ManagementTraitsLegacyScopeGuard(db))
	r.POST("/exam/api/mbti/submit", func(c *gin.Context) {
		managementTraitsLegacyBackground(c, func() { close(started); <-finish })
		c.Status(204)
	})
	go func() {
		req := httptest.NewRequest("POST", "/exam/api/mbti/submit", strings.NewReader(`{"paperId":"legacy"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
		close(returned)
	}()
	<-started
	<-returned
	frozen := make(chan func(), 1)
	go func() { frozen <- service.LockManagementTraitsRuntimeFreeze() }()
	select {
	case release := <-frozen:
		release()
		t.Fatal("freeze skipped in-flight background report")
	case <-time.After(20 * time.Millisecond):
	}
	close(finish)
	select {
	case release := <-frozen:
		release()
	case <-time.After(time.Second):
		t.Fatal("background lease leaked")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyGuardRequiredInventory(t *testing.T) {
	// Fixed requirements independent of the implementation's route registry.
	required := []managementTraitsLegacyRoute{
		{"POST", "/exam/api/exam/exam/save", "exam", false},
		{"POST", "/exam/api/exam/exam/state", "exam", false},
		{"POST", "/exam/api/exam/exam/delete", "exam", false},
		{"POST", "/exam/api/exam/exam/pdf-team", "exam", false},
		{"POST", "/exam/api/exam/exam/generate-report", "paper", false},
		{"POST", "/exam/api/paper/paper/save", "paper", false},
		{"POST", "/exam/api/paper/paper/delete", "paper", false},
		{"POST", "/exam/api/paper/paper/create-paper", "paper", false},
		{"POST", "/exam/api/paper/paper/fill-answer", "paper", false},
		{"POST", "/exam/api/paper/paper/hand-exam", "paper", false},
		{"POST", "/exam/api/paper/paper/review-paper", "paper", false},
		{"POST", "/exam/api/candidate/save", "candidate", true},
		{"POST", "/exam/api/candidate/update", "candidate", false},
		{"PUT", "/exam/api/candidate", "candidate", false},
		{"DELETE", "/exam/api/candidate/:ids", "candidate", false},
		{"DELETE", "/exam/api/candidate/logistic/:ids", "candidate", false},
		{"DELETE", "/exam/api/candidate/logicDeletePdfByIds/:ids", "both", false},
		{"PUT", "/exam/api/candidate/logicDeletePdfByIds/:ids", "both", false},
		{"POST", "/exam/api/candidate/end-time", "candidate", false},
		{"POST", "/exam/api/candidate/pdf-persistence", "candidate", false},
		{"POST", "/exam/api/tester/login", "tester", true},
		{"POST", "/exam/api/tester", "tester", false},
		{"PUT", "/exam/api/tester", "tester", false},
		{"DELETE", "/exam/api/tester/:ids", "tester", false},
		{"DELETE", "/exam/api/tester/logistic/:ids", "tester", false},
		{"POST", "/exam/api/tester/end-time", "tester", false},
		{"POST", "/exam/api/tester/pdf-persistence", "tester", false},
		{"POST", "/exam/api/tester/importData", "tester", false},
		{"POST", "/exam/api/mbti/paper-detail", "paper", false},
		{"POST", "/exam/api/mbti/fill-answer", "paper", false},
		{"POST", "/exam/api/mbti/submit", "paper", false},
		{"POST", "/exam/api/mbti/score", "paper", false},
		{"POST", "/exam/api/mbti/generate-report", "paper", false},
		{"POST", "/exam/api/mbti/download-report", "paper", false},
		{"POST", "/exam/api/mbti/batch-download-simple", "paper", false},
		{"POST", "/exam/api/candidate/tester-list", "collection", false},
		{"GET", "/exam/api/tester", "collection", false},
		{"GET", "/exam/api/tester/list", "collection", false},
		{"GET", "/exam/api/tester/tester-list", "collection", false},
		{"POST", "/exam/api/paper/paper/paging", "collection", false},
	}
	for _, want := range required {
		t.Run(want.method+want.path, func(t *testing.T) {
			path := strings.ReplaceAll(want.path, ":ids", "old,new")
			got, ok := managementTraitsLegacyMatch(want.method, path)
			if !ok || got != want {
				t.Fatalf("required route missing or incorrectly classified: got=%+v want=%+v", got, want)
			}
		})
	}
	t.Logf("fixed required routes=%d; MBTI=7; participant collections=4; paper paging=1", len(required))
}

func TestManagementTraitsLegacyGuardActualHandlersProtected(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body, scope string
	}{
		{"Paper.Save", "POST", "/exam/api/paper/paper/save", `{"id":"new-paper","title":"Changed"}`, "paper"},
		{"Paper.DeleteOldFirst", "POST", "/exam/api/paper/paper/delete", `{"ids":["old-paper","new-paper"]}`, "paper-batch"},
		{"Paper.DeleteNewFirst", "POST", "/exam/api/paper/paper/delete", `{"ids":["new-paper","old-paper"]}`, "paper-batch"},
		{"Paper.FillAnswer", "POST", "/exam/api/paper/paper/fill-answer", `{"paperId":"new-paper","quId":"source-qu","answers":["answer"],"answer":"A"}`, "paper"},
		{"Paper.HandExam", "POST", "/exam/api/paper/paper/hand-exam", `{"id":"new-paper"}`, "paper"},
		{"Paper.ReviewForeignQuestion", "POST", "/exam/api/paper/paper/review-paper", `{"id":"old-paper","quList":[{"id":"foreign-question","actualScore":9}]}`, "review"},
		{"Candidate.Update", "PUT", "/exam/api/candidate", `{"id":"new-candidate","name":"Changed","paperId":null,"telephone":null}`, "candidate"},
		{"Candidate.UpdateAlias", "POST", "/exam/api/candidate/update", `{"id":"new-candidate","name":"Changed"}`, "candidate"},
		{"Candidate.DeletePDF", "DELETE", "/exam/api/candidate/logicDeletePdfByIds/old-person,new-person", ``, "both"},
		{"Candidate.PutDeletePDF", "PUT", "/exam/api/candidate/logicDeletePdfByIds/new-person,old-person", ``, "both"},
		{"Candidate.EndTime", "POST", "/exam/api/candidate/end-time", `{"paperId":"new-paper"}`, "paper"},
		{"Candidate.PdfPersistence", "POST", "/exam/api/candidate/pdf-persistence", ``, "paper"},
		{"Tester.Create", "POST", "/exam/api/tester", `{"examId":"new-exam","name":"Person","telephone":"13800000000"}`, "exam"},
		{"Tester.Update", "PUT", "/exam/api/tester", `{"id":"new-tester","name":"Changed","examId":null,"idNumber":null,"paperId":null}`, "tester"},
		{"Tester.EndTime", "POST", "/exam/api/tester/end-time", `{"idNumber":"new-tester"}`, "identifier"},
		{"Tester.PdfPersistence", "POST", "/exam/api/tester/pdf-persistence", ``, "identifier"},
		{"Exam.Save", "POST", "/exam/api/exam/exam/save", `{"id":"new-exam","title":"Changed","isOpen":2,"joinType":1,"repoList":[]}`, "exam"},
		{"Exam.StateOldFirst", "POST", "/exam/api/exam/exam/state", `{"ids":["old-exam","new-exam"],"state":1}`, "exam-batch"},
		{"Exam.StateNewFirst", "POST", "/exam/api/exam/exam/state", `{"ids":["new-exam","old-exam"],"state":1}`, "exam-batch"},
		{"Exam.Delete", "POST", "/exam/api/exam/exam/delete", `{"ids":["old-exam","new-exam"]}`, "exam-batch"},
		{"Exam.GenerateReport", "POST", "/exam/api/exam/exam/generate-report", `{"paperId":"new-paper"}`, "paper"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			dir := t.TempDir()
			sentinel := filepath.Join(dir, "existing.pdf")
			const pdf = "%PDF-protected-sentinel"
			if err := os.WriteFile(sentinel, []byte(pdf), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg := &config.Config{}
			cfg.Upload.Path = dir
			paper := NewPaperHandler(db)
			candidate := NewCandidateHandler(db, cfg)
			tester := NewTesterHandler(db, cfg)
			// The public constructor performs migrations; this isolated fixture must not.
			exam := &ExamHandler{db: db, cfg: cfg}
			handlers := map[string]gin.HandlerFunc{
				"Paper.Save":                  paper.Save,
				"Paper.DeleteOldFirst":        paper.Delete,
				"Paper.DeleteNewFirst":        paper.Delete,
				"Paper.FillAnswer":            paper.FillAnswer,
				"Paper.HandExam":              paper.HandExam,
				"Paper.ReviewForeignQuestion": paper.ReviewPaper,
				"Candidate.Update":            candidate.Update,
				"Candidate.UpdateAlias":       candidate.Update,
				"Candidate.DeletePDF":         candidate.LogicDeletePdfByIds,
				"Candidate.PutDeletePDF":      candidate.LogicDeletePdfByIds,
				"Candidate.EndTime":           candidate.EndTimeCandidate,
				"Candidate.PdfPersistence":    candidate.PdfPersistence,
				"Tester.Create":               tester.Create,
				"Tester.Update":               tester.Update,
				"Tester.EndTime":              tester.EndTime,
				"Tester.PdfPersistence":       tester.PdfPersistence,
				"Exam.Save":                   exam.Save,
				"Exam.StateOldFirst":          exam.State,
				"Exam.StateNewFirst":          exam.State,
				"Exam.Delete":                 exam.Delete,
				"Exam.GenerateReport":         exam.GenerateReport,
			}
			mutations := 0
			audit := func(*gorm.DB) { mutations++ }
			for _, err := range []error{
				db.Callback().Create().Before("gorm:create").Register("guard_test_no_create", audit),
				db.Callback().Update().Before("gorm:update").Register("guard_test_no_update", audit),
				db.Callback().Delete().Before("gorm:delete").Register("guard_test_no_delete", audit),
				db.Callback().Raw().Before("gorm:raw").Register("guard_test_no_exec", audit),
			} {
				if err != nil {
					t.Fatal(err)
				}
			}
			legacyGuardTables(mock, true)
			if tc.scope == "review" {
				mock.ExpectQuery("SELECT paper_id FROM el_paper_qu WHERE").WithArgs("foreign-question").WillReturnRows(sqlmock.NewRows([]string{"paper_id"}).AddRow("new-paper"))
			}
			switch tc.scope {
			case "paper", "paper-batch", "review", "exam", "exam-batch":
				q := mock.ExpectQuery("SELECT id, exam_id FROM el_paper WHERE")
				switch tc.scope {
				case "paper":
					q.WithArgs("new-paper")
				case "paper-batch", "review":
					q.WithArgs("new-paper", "old-paper")
				case "exam":
					q.WithArgs("new-exam")
				case "exam-batch":
					q.WithArgs("new-exam", "old-exam")
				}
				q.WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}))
			}
			if tc.scope != "tester" && tc.scope != "identifier" {
				q := mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE")
				switch tc.scope {
				case "candidate":
					q.WithArgs("new-candidate")
				case "both":
					q.WithArgs("new-person", "old-person")
				}
				q.WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
			}
			if tc.scope != "candidate" {
				q := mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_tester WHERE")
				switch tc.scope {
				case "tester":
					q.WithArgs("new-tester")
				case "identifier":
					q.WithArgs("new-tester", "new-tester", "new-tester")
				case "both":
					q.WithArgs("new-person", "old-person")
				}
				q.WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
			}
			switch tc.scope {
			case "exam", "exam-batch":
				q := mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE")
				if tc.scope == "exam" {
					q.WithArgs("new-exam")
				} else {
					q.WithArgs("new-exam", "old-exam")
				}
				q.WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			default:
				q := mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_paper_snapshot WHERE")
				switch tc.scope {
				case "paper":
					q.WithArgs("new-paper")
				case "paper-batch", "review":
					q.WithArgs("new-paper", "old-paper")
				case "candidate":
					q.WithArgs("candidate", "new-candidate")
				case "tester", "identifier":
					q.WithArgs("tester", "new-tester")
				case "both":
					q.WithArgs("candidate", "new-person", "candidate", "old-person", "tester", "new-person", "tester", "old-person")
				}
				q.WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			}
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			if strings.HasSuffix(tc.path, "/pdf-persistence") {
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				key, value := "paperId", "new-paper"
				if tc.scope == "identifier" {
					key, value = "idNumber", "new-tester"
				}
				if err := writer.WriteField(key, value); err != nil {
					t.Fatal(err)
				}
				file, err := writer.CreateFormFile("file", "replacement.pdf")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := file.Write([]byte("%PDF-replacement")); err != nil {
					t.Fatal(err)
				}
				if err := writer.Close(); err != nil {
					t.Fatal(err)
				}
				request = httptest.NewRequest(tc.method, tc.path, &body)
				request.Header.Set("Content-Type", writer.FormDataContentType())
			}
			pattern := tc.path
			if tc.scope == "both" {
				pattern = "/exam/api/candidate/logicDeletePdfByIds/:ids"
			}
			entered := false
			r := gin.New()
			r.Use(ManagementTraitsLegacyScopeGuard(db))
			r.Handle(tc.method, pattern, func(c *gin.Context) { entered = true; c.Next() }, handlers[tc.name])
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			var result struct {
				Code    int    `json:"code"`
				Message string `json:"msg"`
				Success bool   `json:"success"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil || w.Code != http.StatusForbidden || result.Code != 1 || result.Success || !strings.Contains(result.Message, "专用入口") || entered || mutations != 0 {
				t.Fatalf("actual handler escaped: HTTP=%d body=%s entered=%v DML=%d", w.Code, w.Body.String(), entered, mutations)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
			remaining, err := os.ReadFile(sentinel)
			if err != nil || string(remaining) != pdf {
				t.Fatalf("sentinel changed/deleted: %q %v", remaining, err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("unexpected uploaded files: %v %v", entries, err)
			}
			t.Logf("actual=%s HTTP=%d code=%d handlerEntered=%v DML=%d sentinel=unchanged", tc.name, w.Code, result.Code, entered, mutations)
		})
	}
}

func TestManagementTraitsLegacyGuardActualDTONullsAbsentSchema(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		t.Run(kind, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			legacyGuardTables(mock, false)
			body := `{"id":"legacy-person","examId":null,"idNumber":null,"paperId":null,"telephone":null,"age":null,"stuFlag":null}`
			path := "/exam/api/" + kind
			called := false
			r := gin.New()
			r.Use(ManagementTraitsLegacyScopeGuard(db))
			r.PUT(path, func(c *gin.Context) {
				called = true
				raw, err := io.ReadAll(c.Request.Body)
				if err != nil || string(raw) != body {
					t.Fatal("legacy body not restored", string(raw), err)
				}
				c.Request.Body = io.NopCloser(bytes.NewReader(raw))
				if kind == "tester" {
					var dto testerReq
					if err := c.ShouldBindJSON(&dto); err != nil || dto.ID != "legacy-person" || dto.ExamID != "" || dto.IDNumber != "" || dto.PaperID != nil || dto.Telephone != nil || dto.Age != nil || dto.StuFlag != nil {
						t.Fatal("actual tester DTO null compatibility", dto, err)
					}
				} else {
					var dto Candidate
					if err := c.ShouldBindJSON(&dto); err != nil || dto.ID != "legacy-person" || dto.ExamID != "" || dto.PaperID != nil || dto.Telephone != nil || dto.Age != nil || dto.StuFlag != nil {
						t.Fatal("actual candidate DTO null compatibility", dto, err)
					}
				}
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest("PUT", path, strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			if !called || w.Code != http.StatusNoContent {
				t.Fatal("legacy nullable DTO blocked", w.Code, w.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyGuardRootBatchNullCompatibility(t *testing.T) {
	for _, body := range []string{`[]`, `[null]`, `[null,"old-paper","new-paper"]`, `["new-paper",null,"old-paper"]`} {
		t.Run(body, func(t *testing.T) {
			var legacyIDs []string
			if err := json.Unmarshal([]byte(body), &legacyIDs); err != nil {
				t.Fatal(err)
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/exam/api/paper/paper/delete", strings.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			target, err := managementTraitsLegacyTargets(c, managementTraitsLegacyRoute{method: "POST", path: "/exam/api/paper/paper/delete", kind: "paper"})
			if err != nil {
				t.Fatal(err)
			}
			want := make([]string, 0)
			for _, id := range legacyIDs {
				if id != "" {
					want = append(want, id)
				}
			}
			if !slices.Equal(target.PaperIDs, want) {
				t.Fatal("null element discarded a nonempty batch ID", target.PaperIDs, want)
			}
			var restored []string
			if err := c.ShouldBindJSON(&restored); err != nil || !reflect.DeepEqual(restored, legacyIDs) {
				t.Fatal("legacy root array binding changed", restored, legacyIDs, err)
			}
			if len(want) > 0 {
				db, mock := identityHandlerDB(t)
				legacyGuardTables(mock, true)
				mock.ExpectQuery("SELECT id, exam_id FROM el_paper WHERE").WithArgs("new-paper", "old-paper").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}))
				mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE").WithArgs("new-paper", "old-paper").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
				mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_tester WHERE").WithArgs("new-paper", "old-paper").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
				mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_paper_snapshot WHERE").WithArgs("new-paper", "old-paper").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
				called := false
				r := gin.New()
				r.Use(ManagementTraitsLegacyScopeGuard(db))
				r.POST("/exam/api/paper/paper/delete", func(c *gin.Context) { called = true; c.Status(http.StatusNoContent) })
				request := httptest.NewRequest("POST", "/exam/api/paper/paper/delete", strings.NewReader(body))
				request.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, request)
				if called || w.Code != http.StatusForbidden {
					t.Fatal("null bypassed whole-batch protection", w.Code, w.Body.String())
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
				t.Logf("root batch=%s HTTP=%d protected ID retained; legacy null binds to empty string", body, w.Code)
			}
		})
	}
}
