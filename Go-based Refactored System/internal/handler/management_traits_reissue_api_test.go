package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
	"gorm.io/gorm/schema"
)

func TestManagementTraitsReissueAPIAuthenticationAndStrictInput(t *testing.T) {
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
		for _, body := range []string{`{"runId":""}`, `{"runId":null}`, `{"runId":"r","runId":"r"}`, `{"runId":"r","path":"/secret"}`, `{"runId":"r","approved":true}`, `{"paperId":"p"}`} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("POST", "/mng/report-reissues/generate", bytes.NewBufferString(body)))
			want := 400
			if user == nil {
				want = 401
			} else if user.UserID != 1 {
				want = 403
			}
			if w.Code != want {
				t.Fatalf("status %d want %d", w.Code, want)
			}
		}
	}
}

func reissueHandlerSchema(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}).AddRow("el_mng_report_reissue", "InnoDB").AddRow("el_mng_reissue_audit", "InnoDB"))
	cols := sqlmock.NewRows([]string{"table_name", "column_name", "column_type", "is_nullable", "character_set_name", "collation_name"})
	for _, n := range []string{"id", "paper_id", "exam_id"} {
		cols.AddRow("el_mng_result_run", n, "varchar(64)", "NO", "utf8mb4", "utf8mb4_0900_ai_ci")
	}
	idx := sqlmock.NewRows([]string{"table_name", "index_name", "non_unique", "seq_in_index", "column_name", "prefix_length"})
	for _, v := range []any{model.ManagementTraitsReportReissue{}, model.ManagementTraitsReissueAudit{}} {
		s, e := schema.Parse(v, &sync.Map{}, schema.NamingStrategy{})
		if e != nil {
			t.Fatal(e)
		}
		for _, f := range s.Fields {
			typ := strings.ToLower(f.TagSettings["TYPE"])
			cs, coll := "", ""
			if strings.Contains(typ, "char") || typ == "longtext" {
				cs, coll = "utf8mb4", "utf8mb4_bin"
				if f.DBName == "run_id" || f.DBName == "paper_id" || f.DBName == "exam_id" {
					coll = "utf8mb4_0900_ai_ci"
				}
			}
			cols.AddRow(s.Table, f.DBName, typ, "NO", cs, coll)
		}
		idx.AddRow(s.Table, "PRIMARY", 0, 1, "id", 0)
		for _, i := range s.ParseIndexes() {
			non := 1
			if i.Class == "UNIQUE" {
				non = 0
			}
			for n, f := range i.Fields {
				idx.AddRow(s.Table, i.Name, non, n+1, f.DBName, 0)
			}
		}
	}
	for n, c := range []string{"run_id", "paper_id", "exam_id"} {
		idx.AddRow("el_mng_report_reissue", "idx_mng_reissue_run", 1, n+1, c, 0)
	}
	for n, c := range []string{"report_id", "paper_id", "exam_id"} {
		idx.AddRow("el_mng_reissue_audit", "idx_mng_reissue_audit_identity", 1, n+1, c, 0)
	}
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name").WillReturnRows(cols)
	mock.ExpectQuery("SELECT table_name AS table_name,index_name AS index_name").WillReturnRows(idx)
	fk := sqlmock.NewRows([]string{"table_name", "constraint_name", "column_name", "ordinal_position", "referenced_table_name", "referenced_column_name", "update_rule", "delete_rule"})
	for n, c := range []string{"run_id", "paper_id", "exam_id"} {
		fk.AddRow("el_mng_report_reissue", "fk_mng_reissue_run", c, n+1, "el_mng_result_run", []string{"id", "paper_id", "exam_id"}[n], "RESTRICT", "RESTRICT")
	}
	for n, c := range []string{"report_id", "paper_id", "exam_id"} {
		fk.AddRow("el_mng_reissue_audit", "fk_mng_reissue_audit", c, n+1, "el_mng_report_reissue", []string{"id", "paper_id", "exam_id"}[n], "RESTRICT", "RESTRICT")
	}
	mock.ExpectQuery("SELECT k.table_name AS table_name,k.constraint_name AS constraint_name").WillReturnRows(fk)
}

func TestManagementTraitsReissueAPIAllRoutesPermissionsAndMissingSchema(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("REPORT_EFFECTIVE_ENV", "local")
	t.Setenv("MNG_TEST_REPORT_ENV", "local")
	t.Setenv("MNG_TEST_REPORT_DIR", t.TempDir())
	t.Setenv("MNG_TEST_CONTENT_PATH", filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-content-v1.xlsx"))
	t.Setenv("MNG_TEST_TEMPLATE_PATH", filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-only-v2.docx"))
	for _, u := range []*model.LoginUser{nil, {UserID: 8, Permissions: []string{"exam:list"}}, {UserID: 0, Permissions: []string{"*:*:*"}}, {UserID: 1}, {UserID: 8, Permissions: []string{"*:*:*"}}} {
		for _, route := range []string{"/report-reissues?paperId=paper", "/report-reissues/qualification?runId=run", "/report-reissues/view?reportId=report", "/report-reissues/download?reportId=report"} {
			db, mock := identityHandlerDB(t)
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if u != nil {
					c.Set("loginUser", u)
				}
				c.Next()
			})
			NewManagementTraitsRuntimeHandler(db, &config.Config{}).RegisterRoutes(r.Group("/mng"))
			want := 403
			if u == nil {
				want = 401
			} else if u.UserID > 0 && (u.UserID == 1 || len(u.Permissions) == 1 && u.Permissions[0] == "*:*:*") {
				want = 503
				if strings.Contains(route, "qualification") { // Mandatory runtime schema is absent, not a fake eligible result.
					identityHandlerRuntimeSchema(t, mock)
					mock.ExpectBegin()
					mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"id"}))
					mock.ExpectRollback()
					want = 409
				} else {
					mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
				}
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", "/mng"+route, nil))
			if w.Code != want {
				t.Fatalf("%s status %d want %d", route, w.Code, want)
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, e := range []error{service.ErrManagementTraitsReissueClosed, fmtReissueWrapped(), service.ErrManagementTraitsReissueInvalid} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		managementTraitsReissueRespond(c, nil, e)
		if strings.Contains(w.Body.String(), "SECRET") || strings.Contains(w.Body.String(), "/private") {
			t.Fatal("raw error leaked")
		}
	}
}

func fmtReissueWrapped() error {
	return errors.Join(errors.New("SECRET /private"), service.ErrManagementTraitsReissueInvalid)
}

func TestManagementTraitsReissueAPIArchivedHTTPViewDownload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	private := filepath.Join(root, "reissues")
	if os.Mkdir(private, 0700) != nil {
		t.Fatal("fixture")
	}
	t.Setenv("REPORT_EFFECTIVE_ENV", "local")
	t.Setenv("MNG_TEST_REPORT_ENV", "local")
	t.Setenv("MNG_TEST_REPORT_DIR", root)
	t.Setenv("MNG_TEST_CONTENT_PATH", filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-content-v1.xlsx"))
	t.Setenv("MNG_TEST_TEMPLATE_PATH", filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-only-v2.docx"))
	sha := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	dto := managementTestWordDTO()
	dto.Participant.Name = "SYNTHETIC ARCHIVE IDENTITY"
	archive, _ := json.Marshal(map[string]any{"schema": "mng-frozen-reissue-archive-v1", "kind": "verified_frozen_result", "report": dto, "paper": model.ManagementTraitsPaperSnapshot{PaperID: dto.PaperID, ExamID: dto.ExamID}, "bundle": model.ManagementTraitsDefinitionBundle{}, "questions": make([]model.ManagementTraitsPaperQuestionSnapshot, 140)})
	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("archive"), 200)...)
	id := "22222222-2222-4222-8222-222222222222"
	if os.WriteFile(filepath.Join(private, id+".pdf"), pdf, 0600) != nil {
		t.Fatal("fixture")
	}
	report := model.ManagementTraitsReportReissue{ID: id, RunID: dto.RunID, PaperID: dto.PaperID, ExamID: dto.ExamID, Kind: "verified_frozen_result", Status: "completed", DataSnapshot: string(archive), DataSHA: sha(archive), TemplateSHA: "05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c", ContentSHA: dto.ContentSourceSHA, FileKey: id + ".pdf", FileSHA: sha(pdf), FileBytes: int64(len(pdf)), CreatedBy: 1, CreatedAt: time.Now()}
	for _, action := range []string{"view", "download"} {
		db, mock := identityHandlerDB(t)
		reissueHandlerSchema(t, mock)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .*el_mng_report_reissue.*WHERE id = ").WithArgs(id, 1).WillReturnRows(identityHandlerModelRows(t, report))
		mock.ExpectExec("INSERT INTO `el_mng_reissue_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set("loginUser", &model.LoginUser{UserID: 1}); c.Next() })
		NewManagementTraitsRuntimeHandler(db, &config.Config{}).RegisterRoutes(r.Group("/mng"))
		server := httptest.NewServer(r)
		resp, e := http.Get(server.URL + "/mng/report-reissues/" + action + "?reportId=" + id)
		if e != nil {
			server.Close()
			t.Fatal(e)
		}
		b, e := io.ReadAll(io.LimitReader(resp.Body, 50<<20))
		resp.Body.Close()
		server.Close()
		if e != nil || !bytes.Equal(b, pdf) || resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "application/pdf" || resp.Header.Get("Cache-Control") != "no-store" || !strings.Contains(resp.Header.Get("Content-Disposition"), "filename*=UTF-8''") {
			t.Fatalf("HTTP archive status=%d bytes=%d contentType=%s error=%v mock=%v", resp.StatusCode, len(b), resp.Header.Get("Content-Type"), e, mock.ExpectationsWereMet())
		}
		if e := mock.ExpectationsWereMet(); e != nil {
			t.Fatal(e)
		}
	}
}

func TestManagementTraitsReissueAPIMaskedRealAnswersLibreOffice(t *testing.T) {
	file := os.Getenv("MNG_REISSUE_MASKED_DTO_INPUT")
	output := os.Getenv("MNG_REISSUE_MASKED_PDF_OUTPUT")
	if file == "" || output == "" {
		t.Skip("explicit synthetic-identity real-answer local PDF opt-in required")
	}
	if !filepath.IsAbs(file) || !filepath.IsAbs(output) {
		t.Fatal("absolute local paths required")
	}
	b, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var dto service.ManagementTraitsTestReportData
	if json.Unmarshal(b, &dto) != nil || dto.Participant.Name != "SYNTHETIC IDENTITY - REAL ANSWERS" || dto.Overall.Score != "58.64" || !dto.TestOnly {
		t.Fatal("wrong safe sample")
	}
	template := managementTestWordTemplate(t)
	docx, e := renderManagementTraitsTestWord(template, dto)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	pdf, e := libreofficepdf.NewClient("").Convert(ctx, "masked-real-answers.docx", docx)
	if e != nil || !bytes.HasPrefix(pdf, []byte("%PDF-")) || len(pdf) < 1024 {
		t.Fatal("actual LO PDF", e)
	}
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.Write(pdf)
	ce := f.Close()
	if e != nil || ce != nil {
		t.Fatal("PDF output")
	}
	h := sha256.Sum256(pdf)
	t.Logf("real answers / synthetic identity TEST PDF bytes=%d SHA=%x", len(pdf), h)
}
