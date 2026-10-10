package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
)

func managementTraitsTemplateRouter(login *model.LoginUser, h *ManagementTraitsRuntimeHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if login != nil {
			c.Set("loginUser", login)
		}
		c.Next()
	})
	router.GET("/template", h.ManagementTraitsTemplateInfo)
	router.GET("/template/download", h.DownloadManagementTraitsTemplate)
	router.POST("/template/upload", h.UploadManagementTraitsTemplate)
	return router
}

func TestManagementTraitsSharedTemplateManagementAuthorization(t *testing.T) {
	t.Setenv("MNG_TEST_TEMPLATE_PATH", filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-only-v2.docx"))
	for _, tc := range []struct {
		name string
		user *model.LoginUser
		want int
	}{
		{name: "anonymous", want: http.StatusUnauthorized},
		{name: "ordinary", user: &model.LoginUser{UserID: 9, Permissions: []string{"exam:list"}}, want: http.StatusForbidden},
		{name: "administrator", user: &model.LoginUser{UserID: 1}, want: http.StatusOK},
		{name: "global", user: &model.LoginUser{UserID: 9, Permissions: []string{"*:*:*"}}, want: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			managementTraitsTemplateRouter(tc.user, &ManagementTraitsRuntimeHandler{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/template", nil))
			if response.Code != tc.want {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestManagementTraitsSharedTemplateMetadataAndDownload(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-only-v2.docx")
	t.Setenv("MNG_TEST_TEMPLATE_PATH", path)
	router := managementTraitsTemplateRouter(&model.LoginUser{UserID: 1}, &ManagementTraitsRuntimeHandler{})

	metadata := httptest.NewRecorder()
	router.ServeHTTP(metadata, httptest.NewRequest(http.MethodGet, "/template", nil))
	var body struct {
		Code int                          `json:"code"`
		Data managementTraitsTemplateInfo `json:"data"`
	}
	if metadata.Code != http.StatusOK || json.Unmarshal(metadata.Body.Bytes(), &body) != nil || body.Code != 0 {
		t.Fatalf("metadata status=%d body=%s", metadata.Code, metadata.Body.String())
	}
	if !body.Data.Exists || !body.Data.Valid || body.Data.FileName != managementTraitsSharedTemplateFileName || body.Data.ContentControls != 90 || body.Data.BusinessCharts != 6 || body.Data.NumericLabels != 5 || body.Data.ExternalLinks != 0 || strings.Join(body.Data.ProductCodes, ",") != "00501,00502" || len(body.Data.SHA256) != 64 {
		t.Fatalf("metadata=%+v", body.Data)
	}
	if len(body.Data.SemanticFields) != 95 {
		t.Fatalf("semantic fields=%d, want 95", len(body.Data.SemanticFields))
	}

	download := httptest.NewRecorder()
	router.ServeHTTP(download, httptest.NewRequest(http.MethodGet, "/template/download", nil))
	if download.Code != http.StatusOK || !strings.Contains(download.Header().Get("Content-Type"), "application/vnd.openxmlformats-officedocument.wordprocessingml.document") || !strings.Contains(download.Header().Get("Content-Disposition"), "filename*=UTF-8''") || download.Header().Get("Cache-Control") != "no-store, no-cache, must-revalidate" || !bytes.HasPrefix(download.Body.Bytes(), []byte("PK")) {
		t.Fatalf("download status=%d headers=%v bytes=%d", download.Code, download.Header(), download.Body.Len())
	}
}

func TestBugFB226_ManagementTraitsTemplateAllowsRepeatedSemanticField(t *testing.T) {
	template, err := os.ReadFile(filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-only-v2.docx"))
	if err != nil {
		t.Fatal(err)
	}
	document := string(readWordPart(t, template, "word/document.xml"))
	control := regexp.MustCompile(`(?s)<w:sdt>.*?<w:tag w:val="participant\.name".*?</w:sdt>`).FindString(document)
	if control == "" {
		t.Fatal("participant.name control missing")
	}
	overallScoreControl := strings.Replace(control, `w:val="participant.name"`, `w:val="overall.score"`, 1)
	repeated := replaceWordFixturePart(t, template, "word/document.xml", []byte(strings.Replace(document, control, control+control+overallScoreControl, 1)))
	digest := sha256.Sum256(repeated)
	rendered, err := renderManagementTraitsTestWordWithSHA(repeated, hex.EncodeToString(digest[:]), managementTraitsTemplateProbeData())
	if err != nil {
		t.Fatalf("repeated semantic field rejected: %v", err)
	}
	if count := strings.Count(string(readWordPart(t, rendered, "word/document.xml")), "模板校验"); count < 2 {
		t.Fatalf("repeated participant.name filled %d times", count)
	}
	if !strings.Contains(string(readWordPart(t, rendered, "word/document.xml")), ">50.00</w:t>") {
		t.Fatal("new optional overall.score control was not filled")
	}
	info, err := validateManagementTraitsTemplate(repeated)
	if err != nil || info.ContentControls != 92 {
		t.Fatalf("expanded template controls=%d err=%v", info.ContentControls, err)
	}
}

func TestManagementTraitsSharedTemplateUploadIsValidatedAndAtomic(t *testing.T) {
	original, err := os.ReadFile(filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-only-v2.docx"))
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), managementTraitsSharedTemplateFileName)
	if err := os.WriteFile(target, original, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MNG_TEST_TEMPLATE_PATH", target)
	handler := &ManagementTraitsRuntimeHandler{}
	router := managementTraitsTemplateRouter(&model.LoginUser{UserID: 1}, handler)

	invalid := append([]byte(nil), original...)
	invalid[0] = 'X'
	response := uploadManagementTraitsTemplateRequest(t, router, invalid, "invalid.docx")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "模板校验失败") {
		t.Fatalf("invalid upload status=%d body=%s", response.Code, response.Body.String())
	}
	current, _ := os.ReadFile(target)
	if !bytes.Equal(current, original) {
		t.Fatal("invalid upload changed the active template")
	}
	external := replaceWordFixturePart(t, original, "word/charts/_rels/chart1.xml.rels", []byte(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rIdExternal" Target="file:///C:/private.xlsx" TargetMode="External"/></Relationships>`))
	if _, err := validateManagementTraitsTemplate(external); err == nil || !strings.Contains(err.Error(), "外部关系") {
		t.Fatalf("external relationship accepted: %v", err)
	}

	compatible := replaceWordFixturePart(t, original, "docProps/core.xml", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"><cp:title>00501/00502 shared template</cp:title></cp:coreProperties>`))
	response = uploadManagementTraitsTemplateRequest(t, router, compatible, "shared-005.docx")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"code":0`) {
		t.Fatalf("valid upload status=%d body=%s", response.Code, response.Body.String())
	}
	current, _ = os.ReadFile(target)
	if !bytes.Equal(current, compatible) {
		t.Fatal("compatible upload was not installed")
	}
	backups, err := os.ReadDir(filepath.Join(filepath.Dir(target), "backups"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("backups=%d err=%v", len(backups), err)
	}
	allowed, err := handler.managementTraitsTemplateSHAs()
	if err != nil || len(allowed) != 2 {
		t.Fatalf("allowed template SHAs=%d err=%v", len(allowed), err)
	}
	for _, template := range [][]byte{original, compatible} {
		digest := sha256.Sum256(template)
		if !allowed[hex.EncodeToString(digest[:])] {
			t.Fatal("current or historical template SHA is not retained")
		}
	}
}

func uploadManagementTraitsTemplateRequest(t *testing.T, router http.Handler, data []byte, name string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/template/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
