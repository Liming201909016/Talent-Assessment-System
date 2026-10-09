package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestPhase1WordTemplateCandidateUploadContract(t *testing.T) {
	path := os.Getenv("PHASE1_TEMPLATE_CANDIDATE_PATH")
	if path == "" {
		t.Skip("PHASE1_TEMPLATE_CANDIDATE_PATH is not configured")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := validatePhase1WordTemplateUpload(data)
	if err != nil {
		t.Fatal(err)
	}
	if contract.ContentControls == 0 || contract.UsedFields == 0 || contract.ExternalLinks != 0 || contract.VisibleTokens != 0 || contract.Charts != 12 {
		t.Fatalf("candidate contract=%+v", contract)
	}
	tokens, charts, err := buildPhase1WordTemplateData(phase1WordTestData())
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderPhase1WordTemplate(data, tokens, charts)
	if err != nil {
		t.Fatal(err)
	}
	chart := string(readWordPart(t, rendered, "word/charts/chart1.xml"))
	for _, required := range []string{
		`<c:pt idx="0"><c:v>心理素养</c:v></c:pt>`,
		`<c:pt idx="1"><c:v>通用能力</c:v></c:pt>`,
		`<c:pt idx="0"><c:v>3.5</c:v></c:pt>`,
		`<c:pt idx="1"><c:v>3.75</c:v></c:pt>`,
	} {
		if !strings.Contains(chart, required) {
			t.Fatalf("candidate rendered group pie missing %s", required)
		}
	}
}

// TestBugFB195_V2TemplateUsesDedicatedUploadContract
// Corresponds to docs/regression-tests.md FB-195.
func TestBugFB195_V2TemplateUsesDedicatedUploadContract(t *testing.T) {
	data, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-v2.docx")
	if err != nil {
		t.Fatal(err)
	}
	contract, err := validatePhase1V2WordTemplateUpload(data)
	if err != nil {
		t.Fatal(err)
	}
	if contract.SchemaVersion != "competency-phase1-report-template-v2" || contract.RegisteredFields != 60 || contract.UsedFields != 60 || contract.BusinessCharts != 12 || contract.ExternalLinks != 0 {
		t.Fatalf("v2 contract=%+v", contract)
	}
	if _, err := validatePhase1WordTemplateUpload(data); err == nil || !strings.Contains(err.Error(), "未支持的内容控件Tag") {
		t.Fatalf("fixture no longer proves the legacy-validator mismatch: %v", err)
	}
}

// TestBugFB170_Phase1GroupPieLabelsStayOutsideChart
// 对应：docs/regression-tests.md #FB-170
// 复现：当前模板把一级饼图系列标签设为inEnd，生成PDF的两个分值压在色块上。
// 期望：候选模板使用outEnd，并为两个点保留经LibreOffice验证的图外布局，运行时不改回图内。
func TestBugFB170_Phase1GroupPieLabelsStayOutsideChart(t *testing.T) {
	path := os.Getenv("PHASE1_TEMPLATE_CANDIDATE_PATH")
	if path == "" {
		t.Skip("PHASE1_TEMPLATE_CANDIDATE_PATH is not configured")
	}
	template, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	chart := readWordPart(t, template, "word/charts/chart1.xml")
	if bytes.Contains(chart, []byte(`<c:dLblPos val="inEnd"/>`)) {
		t.Fatal("group pie still places value labels inside the colored surface")
	}
	if !bytes.Contains(chart, []byte(`<c:dLblPos val="outEnd"/>`)) {
		t.Fatal("group pie outside label position is missing")
	}
	pointLayouts := regexp.MustCompile(`(?s)<c:dLbl><c:idx val="[01]"/>.*?<c:layout>.*?</c:layout>`).FindAll(chart, -1)
	if len(pointLayouts) != 2 {
		t.Fatalf("group pie point manual layouts=%d, want 2", len(pointLayouts))
	}
	for _, label := range regexp.MustCompile(`(?s)<c:dLbl><c:idx val="[01]"/>.*?</c:dLbl>`).FindAll(chart, -1) {
		if bytes.Contains(label, []byte(`<a:schemeClr val="bg1"/>`)) || !bytes.Contains(label, []byte(`<a:schemeClr val="tx1"/>`)) {
			t.Fatalf("group pie outside label does not use visible foreground text: %s", label)
		}
	}
	tokens, charts, err := buildPhase1WordTemplateData(phase1WordTestData())
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderPhase1WordTemplate(template, tokens, charts)
	if err != nil {
		t.Fatal(err)
	}
	renderedChart := readWordPart(t, rendered, "word/charts/chart1.xml")
	if bytes.Contains(renderedChart, []byte(`<c:dLblPos val="inEnd"/>`)) || !bytes.Contains(renderedChart, []byte(`<c:dLblPos val="outEnd"/>`)) {
		t.Fatal("runtime did not preserve the template outside label position")
	}
}

// TestBugFB164_Phase1FooterUsesCurrentPageOnly
// 对应：docs/regression-tests.md #FB-164
// 复现：客户模板页脚使用 PAGE / NUMPAGES，LibreOffice生成11页PDF却显示1～10 / 12。
// 期望：上传门禁拒绝NUMPAGES；运行时兜底删除分隔符和NUMPAGES，仅保留PAGE。
func TestBugFB164_Phase1FooterUsesCurrentPageOnly(t *testing.T) {
	footer := []byte(`<w:ftr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:p><w:r><w:fldChar w:fldCharType="begin"/></w:r><w:r><w:instrText> PAGE </w:instrText></w:r><w:r><w:fldChar w:fldCharType="separate"/></w:r><w:r><w:t>1</w:t></w:r><w:r><w:fldChar w:fldCharType="end"/></w:r><w:r><w:t> / </w:t></w:r><w:r><w:fldChar w:fldCharType="begin"/></w:r><w:r><w:instrText> NUMPAGES </w:instrText></w:r><w:r><w:fldChar w:fldCharType="separate"/></w:r><w:r><w:t>12</w:t></w:r><w:r><w:fldChar w:fldCharType="end"/></w:r></w:p></w:ftr>`)
	parts := map[string][]byte{"word/footer1.xml": footer}
	if err := validatePhase1PageNumberFields(parts); err == nil || !strings.Contains(err.Error(), "总页数") {
		t.Fatalf("NUMPAGES footer accepted: %v", err)
	}
	normalized, err := normalizePhase1PageNumberFooter(footer)
	if err != nil {
		t.Fatal(err)
	}
	content := string(normalized)
	if !strings.Contains(content, " PAGE ") || strings.Contains(content, "NUMPAGES") || strings.Contains(content, " / ") || strings.Contains(content, ">12<") {
		t.Fatalf("footer was not normalized to PAGE only: %s", content)
	}
}

// TestBugFB162_V1TemplateRejectsExternalChartRelationships
// 对应：docs/regression-tests.md #FB-162
func TestBugFB162_V1TemplateRejectsExternalChartRelationships(t *testing.T) {
	parts := map[string][]byte{
		"word/charts/_rels/chart1.xml.rels": []byte(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Target="file:///E:/author/data.xlsx" TargetMode="External"/></Relationships>`),
	}
	err := validatePhase1ChartRelationships(parts)
	if err == nil || !strings.Contains(err.Error(), "外部Excel链接") {
		t.Fatalf("external V1 chart relationship accepted: %v", err)
	}
	parts["word/charts/_rels/chart1.xml.rels"] = bytes.ReplaceAll(parts["word/charts/_rels/chart1.xml.rels"], []byte(` Target="file:///E:/author/data.xlsx" TargetMode="External"`), []byte(` Target=""`))
	if err := validatePhase1ChartRelationships(parts); err != nil {
		t.Fatalf("link-free chart relationship rejected: %v", err)
	}
	cleaned := removePhase1ChartExternalRelationships([]byte(`<Relationships><Relationship Id="rId1" Target="file:///E:/author/data.xlsx" TargetMode="External"/><Relationship Id="rId2" Target="style.xml"/></Relationships>`))
	if bytes.Contains(cleaned, []byte("External")) || !bytes.Contains(cleaned, []byte("style.xml")) {
		t.Fatalf("external relationship cleanup damaged retained relations: %s", cleaned)
	}
	chart := removePhase1ChartExternalData([]byte(`<c:chartSpace><c:externalData r:id="rId1"><c:autoUpdate val="0"/></c:externalData><c:plotArea/></c:chartSpace>`))
	if bytes.Contains(chart, []byte("externalData")) || !bytes.Contains(chart, []byte("plotArea")) {
		t.Fatalf("externalData cleanup damaged chart: %s", chart)
	}
}

// TestBugFB165_WordResavedTemplateStripsExternalChartLinksBeforeValidation
// 对应：docs/regression-tests.md #FB-165
// 复现：Microsoft Word随意编辑并保存模板后，会重新写入图表外部Excel关系，上传被整体拒绝。
// 期望：服务端先清除外部关系和externalData，再执行完整契约校验并保存零外链模板。
func TestBugFB165_WordResavedTemplateStripsExternalChartLinksBeforeValidation(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	unchanged, removed, err := sanitizePhase1WordTemplateUpload(template)
	if err != nil || removed != 0 || !bytes.Equal(unchanged, template) {
		t.Fatalf("zero-link template changed during sanitization: removed=%d err=%v", removed, err)
	}
	externalRelations := []byte(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rIdExternal" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/package" Target="file:///C:/Users/customer/chart-data.xlsx" TargetMode="External"/></Relationships>`)
	withExternalLink := replaceWordFixturePart(t, template, "word/charts/_rels/chart1.xml.rels", externalRelations)
	parts := readPhase1WordPartsForTest(t, withExternalLink)
	chart := parts["word/charts/chart1.xml"]
	chart = bytes.Replace(chart, []byte(`</c:chartSpace>`), []byte(`<c:externalData r:id="rIdExternal"><c:autoUpdate val="0"/></c:externalData></c:chartSpace>`), 1)
	withExternalLink = replaceWordFixturePart(t, withExternalLink, "word/charts/chart1.xml", chart)
	if _, err := validatePhase1WordTemplateUpload(withExternalLink); err == nil || !strings.Contains(err.Error(), "外部Excel链接") {
		t.Fatalf("external-link fixture was not rejected before sanitization: %v", err)
	}

	sanitized, removed, err := sanitizePhase1WordTemplateUpload(withExternalLink)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 2 {
		t.Fatalf("removed external artifacts=%d, want 2", removed)
	}
	if bytes.Contains(sanitized, []byte(`TargetMode="External"`)) || bytes.Contains(sanitized, []byte(`<c:externalData`)) {
		t.Fatal("sanitized upload still contains an external chart link")
	}
	if _, err := validatePhase1WordTemplateUpload(sanitized); err != nil {
		t.Fatalf("sanitized Word-resaved template rejected: %v", err)
	}
}

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
		`GET("/template-v2", competencyReportH.Phase1V2TemplateInfo)`,
		`GET("/template-v2/download", competencyReportH.DownloadPhase1V2Template)`,
		`POST("/template-v2/upload", competencyReportH.UploadPhase1V2Template)`,
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
