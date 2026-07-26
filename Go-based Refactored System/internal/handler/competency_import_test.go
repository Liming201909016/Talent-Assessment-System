package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/xuri/excelize/v2"
)

func TestCompetencyImportTemplate_ReturnsXLSX(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCompetencyImportHandler(nil)
	router := gin.New()
	router.GET("/template", h.ImportTemplate)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/template", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "spreadsheetml.sheet") {
		t.Fatalf("content type = %q", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Header().Get("Content-Disposition"), "filename*=UTF-8''") {
		t.Fatalf("content disposition = %q", w.Header().Get("Content-Disposition"))
	}
	if w.Body.Len() == 0 {
		t.Fatal("template is empty")
	}
	workbook, err := excelize.OpenReader(bytes.NewReader(w.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer workbook.Close()
	sheet := workbook.GetSheetList()[0]
	rows, err := workbook.GetRows(sheet)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || len(rows[0]) != len(service.CompetencyImportHeaders) {
		t.Fatalf("template rows/columns = %d/%d", len(rows), len(rows[0]))
	}
	for index, header := range service.CompetencyImportHeaders {
		if rows[0][index] != header {
			t.Errorf("header %d = %q, want %q", index+1, rows[0][index], header)
		}
	}
	merged, err := workbook.GetMergeCells(sheet)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged) != 0 {
		t.Fatalf("template must not contain merged cells: %v", merged)
	}
}

func TestCompetencyImportHandlers_PreviewIsReadOnlyAndImportIsAtomic(t *testing.T) {
	src := readSourceFile(t, "competency_import.go")
	preview := extractFunctionBody(t, src, "func (h *CompetencyImportHandler) ImportPreview(")
	formal := extractFunctionBody(t, src, "func (h *CompetencyImportHandler) Import(")

	for _, forbidden := range []string{"Create(", "CreateInBatches(", "Save(", "Updates(", "Delete("} {
		if strings.Contains(preview, forbidden) {
			t.Errorf("preview must not write database; found %q", forbidden)
		}
	}
	for _, required := range []string{"expectedHash", "subtle.ConstantTimeCompare", "Transaction(", "CreateInBatches", "len(validation.Errors) > 0"} {
		if !strings.Contains(formal, required) {
			t.Errorf("formal import missing %q", required)
		}
	}
}

func TestCompetencyImportPreview_RejectsMissingAndWrongFileType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCompetencyImportHandler(nil)
	router := gin.New()
	router.POST("/preview", h.ImportPreview)

	missing := httptest.NewRecorder()
	router.ServeHTTP(missing, httptest.NewRequest(http.MethodPost, "/preview", nil))
	if !strings.Contains(missing.Body.String(), "请上传导入文件") {
		t.Fatalf("missing-file response = %s", missing.Body.String())
	}

	body, contentType := multipartRequest(t, "questions.csv", []byte("not,xlsx"), nil)
	wrongType := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/preview", body)
	req.Header.Set("Content-Type", contentType)
	router.ServeHTTP(wrongType, req)
	if !strings.Contains(wrongType.Body.String(), "只支持xlsx格式文件") {
		t.Fatalf("wrong-file response = %s", wrongType.Body.String())
	}
}

func TestCompetencyImport_RejectsFileChangedSincePreview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCompetencyImportHandler(nil)
	router := gin.New()
	router.POST("/import", h.Import)

	file := excelize.NewFile()
	buffer, err := file.WriteToBuffer()
	file.Close()
	if err != nil {
		t.Fatal(err)
	}
	body, contentType := multipartRequest(t, "questions.xlsx", buffer.Bytes(), map[string]string{
		"expectedHash": strings.Repeat("0", 64),
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/import", body)
	req.Header.Set("Content-Type", contentType)
	router.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "导入文件与预览文件不一致") {
		t.Fatalf("hash-mismatch response = %s", w.Body.String())
	}
}

func multipartRequest(t *testing.T, fileName string, data []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

func TestCompetencyImportRoutes_AreAdminOnly(t *testing.T) {
	src := readSourceFile(t, "../router/router.go")
	for _, required := range []string{
		`GET("/import-template", competencyImportH.ImportTemplate)`,
		`POST("/import-preview", competencyImportH.ImportPreview)`,
		`POST("/import", competencyImportH.Import)`,
	} {
		if !strings.Contains(src, required) {
			t.Errorf("router missing %q", required)
		}
	}
}
