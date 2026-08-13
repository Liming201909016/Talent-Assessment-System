package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/graphpdf"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
	"github.com/xuri/excelize/v2"
)

// TestBugFB116_Phase1WordTemplateMapsFrozenReportData
// 对应：docs/regression-tests.md #FB-116
// 复现：一期报告只支持Vue/Chromium，客户调整Word后无法直接影响生成结果。
// 期望：冻结DTO确定性映射为Word占位符与12个原生图表数据。
func TestBugFB116_Phase1WordTemplateMapsFrozenReportData(t *testing.T) {
	data := phase1WordTestData()
	tokens, charts, err := buildPhase1WordTemplateData(data)
	if err != nil {
		t.Fatalf("build phase-1 Word data: %v", err)
	}
	for token, expected := range map[string]string{
		"{{participant.name}}":                     "测试人员",
		"{{participant.telephone}}":                "13800000000",
		"{{overall.level}}":                        "合格胜任",
		"{{overall.diagnosis}}":                    "总体诊断",
		"{{validity.notice}}":                      "效度良好",
		"{{dimension.competency-a1-01.score}}":     "3.75",
		"{{dimension.competency-a1-01.level}}":     "较优秀",
		"{{dimension.competency-a1-01.diagnosis}}": "逻辑思维诊断",
	} {
		if tokens[token] != expected {
			t.Fatalf("token %s=%q, want %q", token, tokens[token], expected)
		}
	}
	if got := charts["word/charts/chart1.xml"]; len(got) != 2 || got[0] != 3.75 || got[1] != 3.5 {
		t.Fatalf("group chart values=%v", got)
	}
	if got := charts["word/charts/chart2.xml"]; len(got) != 10 || got[0] != 3.75 || got[9] != 3.5 {
		t.Fatalf("radar values=%v", got)
	}
	if got := charts["word/charts/chart3.xml"]; len(got) != 2 || got[0] != 3.75 || got[1] != 1.25 {
		t.Fatalf("first doughnut values=%v", got)
	}
}

func TestPhase1WordTemplateReplacesTokensAndChartCaches(t *testing.T) {
	template := makeWordTemplateFixture(t)
	tokens := map[string]string{"{{participant.name}}": "张三 & 李四"}
	charts := map[string][]float64{"word/charts/chart1.xml": {3.75, 1.25}}
	result, err := renderPhase1WordTemplate(template, tokens, charts)
	if err != nil {
		t.Fatalf("render Word template: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(result), int64(len(result)))
	if err != nil {
		t.Fatalf("open rendered docx: %v", err)
	}
	parts := map[string]string{}
	for _, file := range reader.File {
		rc, openErr := file.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		body, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		parts[file.Name] = string(body)
	}
	if got := parts["word/document.xml"]; !bytes.Contains([]byte(got), []byte("张三 &amp; 李四")) || bytes.Contains([]byte(got), []byte("{{participant.name}}")) {
		t.Fatalf("document replacement failed: %s", got)
	}
	chart := parts["word/charts/chart1.xml"]
	if !bytes.Contains([]byte(chart), []byte("<c:v>3.75</c:v>")) || !bytes.Contains([]byte(chart), []byte("<c:v>1.25</c:v>")) {
		t.Fatalf("chart replacement failed: %s", chart)
	}
}

func TestPhase1WordTemplateRejectsMissingRequiredToken(t *testing.T) {
	template := makeWordTemplateFixture(t)
	if _, err := renderPhase1WordTemplate(template, map[string]string{"{{participant.name}}": "张三", "{{participant.age}}": "30"}, map[string][]float64{"word/charts/chart1.xml": {3.75, 1.25}}); err == nil {
		t.Fatal("template missing a required placeholder was accepted")
	}
}

func TestPhase1CustomerWordTemplateHasCompleteStableContract(t *testing.T) {
	templatePath := "../../configs/export-templates/competency-phase1-report.docx"
	template, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read customer Word template: %v", err)
	}
	productionData := phase1WordTestData()
	reportText := productionData["reportText"].(service.Phase1ReportTextSnapshot)
	longDiagnosis := strings.Repeat("面对复杂工作情境时能够依据事实分析问题并推进任务，同时建议通过持续复盘和实践练习巩固优势、改善不足。", 3)
	for dimensionID := range reportText.DimensionTexts {
		reportText.DimensionTexts[dimensionID] = longDiagnosis
	}
	reportText.OverallText = strings.Repeat("整体工作表现符合岗位要求，能够完成常规任务并保持稳定交付，建议结合实际工作表现持续提升。", 3)
	reportText.ValidityText = "本次测评作答效度良好，结果具有较好的参考价值。测评结果仍应结合实际工作表现、行为观察、访谈及其他评价信息进行综合解读。"
	reportText.Disclaimer = strings.Repeat("本测评结果基于受测者自陈反应，不应作为人才决策的唯一依据，应结合面试、绩效表现和行为观察进行综合判断。", 2)
	reportText.GroupTexts["general_ability"] = "通用能力由五个子维度构成，反映受测者作为职业人的通用能力综合情况。"
	reportText.GroupTexts["psychological_quality"] = "心理素养由五个子维度构成，反映受测者心理状态和工作动机的综合情况。"
	productionData["reportText"] = reportText
	tokens, charts, err := buildPhase1WordTemplateData(productionData)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderPhase1WordTemplate(template, tokens, charts)
	if err != nil {
		t.Fatalf("render customer Word template: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(rendered), int64(len(rendered)))
	if err != nil {
		t.Fatalf("rendered Word file is invalid: %v", err)
	}
	for _, file := range reader.File {
		if file.Name != "word/document.xml" {
			continue
		}
		rc, openErr := file.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		body, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		if unresolved := wordTemplateTokenPattern.Find(body); unresolved != nil {
			t.Fatalf("unresolved Word placeholder: %s", unresolved)
		}
		return
	}
	t.Fatal("rendered Word document.xml missing")
}

// TestBugFB117_CustomerTemplateUsesHiddenContentControls
// 对应：docs/regression-tests.md #FB-117
// 复现：客户直接打开运行模板时，长{{...}}字段换行并挤压下划线、图形和固定图片。
// 期望：页面只显示正常示例值，49个稳定字段键保存在Word内容控件Tag元数据中。
func TestBugFB117_CustomerTemplateUsesHiddenContentControls(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	document := readWordPart(t, template, "word/document.xml")
	if unresolved := wordTemplateTokenPattern.Find(document); unresolved != nil {
		t.Fatalf("customer-visible placeholder remains: %s", unresolved)
	}
	tags := wordContentControlTagPattern.FindAllSubmatch(document, -1)
	if len(tags) != 49 {
		t.Fatalf("Word content-control tags=%d, want 49", len(tags))
	}
	seen := make(map[string]bool, len(tags))
	for _, match := range tags {
		tag := string(match[1])
		if seen[tag] {
			t.Fatalf("duplicate Word content-control tag: %s", tag)
		}
		seen[tag] = true
	}
	for _, required := range []string{"participant.name", "overall.level", "validity.notice", "dimension.competency-b1-05.diagnosis"} {
		if !seen[required] {
			t.Fatalf("required Word content-control tag missing: %s", required)
		}
	}
}

func TestPhase1WordTemplateUploadValidation(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	contract, err := validatePhase1WordTemplateUpload(template)
	if err != nil {
		t.Fatalf("valid template rejected: %v", err)
	}
	if contract.ContentControls != 49 || contract.Charts != 12 || contract.VisibleTokens != 0 {
		t.Fatalf("contract=%+v", contract)
	}

	document := string(readWordPart(t, template, "word/document.xml"))
	document = strings.Replace(document, `w:val="dimension.competency-a1-04.diagnosis"`, `w:val="dimension.competency-a1-03.diagnosis"`, 1)
	broken := replaceWordFixturePart(t, template, "word/document.xml", []byte(document))
	if _, err := validatePhase1WordTemplateUpload(broken); err == nil || !strings.Contains(err.Error(), "重复") {
		t.Fatalf("duplicate content-control tag error=%v", err)
	}
}

func TestPhase1EmbeddedWorkbookTemplateContract(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-embedded.docx")
	if err != nil {
		t.Fatalf("read embedded workbook template: %v", err)
	}
	if contract, err := validatePhase1WordTemplateUpload(template); err != nil || contract.ContentControls != 49 || contract.Charts != 12 {
		t.Fatalf("embedded template upload contract=%+v error=%v", contract, err)
	}
	workbook := readWordPart(t, template, phase1ChartWorkbookPath)
	book, err := excelize.OpenReader(bytes.NewReader(workbook))
	if err != nil {
		t.Fatalf("open embedded workbook: %v", err)
	}
	defer book.Close()
	if value, _ := book.GetCellValue("Sheet1", "A2"); value != "通用能力" {
		t.Fatalf("Sheet1 A2=%q", value)
	}
	if value, _ := book.GetCellValue("Sheet2", "A13"); value != "合作意识" {
		t.Fatalf("Sheet2 A13=%q", value)
	}
	for chartIndex := 1; chartIndex <= 12; chartIndex++ {
		rels := string(readWordPart(t, template, fmt.Sprintf("word/charts/_rels/chart%d.xml.rels", chartIndex)))
		if strings.Contains(rels, "TargetMode=\"External\"") || !strings.Contains(rels, `relationships/package`) || !strings.Contains(rels, `../embeddings/competency-phase1-chart-data.xlsx`) {
			t.Fatalf("chart%d is not linked to embedded workbook: %s", chartIndex, rels)
		}
		chart := string(readWordPart(t, template, fmt.Sprintf("word/charts/chart%d.xml", chartIndex)))
		if strings.Contains(chart, "260805数据图表.xlsx") || strings.Contains(chart, "[数据图表.xlsx]") || !strings.Contains(chart, "[competency-phase1-chart-data.xlsx]") {
			t.Fatalf("chart%d formulas do not use the embedded workbook", chartIndex)
		}
	}
}

func TestPhase1RenderingUpdatesChartCacheAndEmbeddedWorkbook(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-embedded.docx")
	if err != nil {
		t.Fatal(err)
	}
	tokens, charts, err := buildPhase1WordTemplateData(phase1WordTestData())
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderPhase1WordTemplate(template, tokens, charts)
	if err != nil {
		t.Fatal(err)
	}
	workbook := readWordPart(t, rendered, phase1ChartWorkbookPath)
	book, err := excelize.OpenReader(bytes.NewReader(workbook))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	for cell, expected := range map[string]string{"B2": "3.75", "B3": "3.5"} {
		if value, _ := book.GetCellValue("Sheet1", cell); value != expected {
			t.Fatalf("Sheet1 %s=%q, want %q", cell, value, expected)
		}
	}
	for cell, expected := range map[string]string{"B4": "3.75", "B13": "3.5", "B33": "3.75", "C33": "1.25"} {
		if value, _ := book.GetCellValue("Sheet2", cell); value != expected {
			t.Fatalf("Sheet2 %s=%q, want %q", cell, value, expected)
		}
	}
	chart := string(readWordPart(t, rendered, "word/charts/chart3.xml"))
	if !strings.Contains(chart, "<c:v>3.75</c:v>") || !strings.Contains(chart, "<c:v>1.25</c:v>") {
		t.Fatalf("chart3 cache not updated: %s", chart)
	}
}

func TestInstallPhase1WordTemplateBacksUpAndAtomicallyReplaces(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "competency-phase1-report.docx")
	if err := os.WriteFile(target, []byte("old-template"), 0o600); err != nil {
		t.Fatal(err)
	}
	backup, err := installPhase1WordTemplate(target, []byte("new-template"), time.Date(2026, 8, 12, 18, 30, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new-template" {
		t.Fatalf("target=%q", got)
	}
	if got, _ := os.ReadFile(backup); string(got) != "old-template" {
		t.Fatalf("backup=%q", got)
	}
}

func TestPhase1WordRendererUsesConfiguredTemplateAndConverter(t *testing.T) {
	converter := &capturingPhase1Converter{pdf: append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 2048)...)}
	renderer := &phase1WordReportRenderer{templatePath: "../../configs/export-templates/competency-phase1-report.docx", converter: converter, timeout: time.Second}
	pdf, err := renderer.Render(context.Background(), "paper-1", phase1WordTestData())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !bytes.Equal(pdf, converter.pdf) || len(converter.docx) == 0 || converter.fileName == "" {
		t.Fatal("Word renderer did not pass the populated DOCX through the converter")
	}
	reader, err := zip.NewReader(bytes.NewReader(converter.docx), int64(len(converter.docx)))
	if err != nil {
		t.Fatalf("captured DOCX invalid: %v", err)
	}
	if len(reader.File) == 0 {
		t.Fatal("captured DOCX empty")
	}
}

func TestNewPhase1WordReportRendererSelectsConfiguredConverter(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.Phase1WordReportCfg
		wantNil    bool
		wantClient any
	}{
		{name: "disabled", cfg: config.Phase1WordReportCfg{}, wantNil: true},
		{name: "libreoffice default", cfg: config.Phase1WordReportCfg{Enabled: true, TimeoutSeconds: 30}, wantClient: (*libreofficepdf.Client)(nil)},
		{name: "graph", cfg: config.Phase1WordReportCfg{Enabled: true, Converter: "graph", TimeoutSeconds: 30}, wantClient: (*graphpdf.Client)(nil)},
		{name: "unknown", cfg: config.Phase1WordReportCfg{Enabled: true, Converter: "unknown", TimeoutSeconds: 30}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := newPhase1WordReportRenderer(&config.Config{Phase1WordReport: test.cfg})
			if test.wantNil {
				if renderer != nil {
					t.Fatal("disabled renderer was created")
				}
				return
			}
			if renderer == nil {
				t.Fatal("enabled renderer is nil")
			}
			switch test.wantClient.(type) {
			case *libreofficepdf.Client:
				if _, ok := renderer.converter.(*libreofficepdf.Client); !ok {
					t.Fatalf("converter type=%T, want LibreOffice", renderer.converter)
				}
			case *graphpdf.Client:
				if _, ok := renderer.converter.(*graphpdf.Client); !ok {
					t.Fatalf("converter type=%T, want Graph", renderer.converter)
				}
			default:
				if renderer.converter != nil {
					t.Fatalf("unknown converter type was accepted: %T", renderer.converter)
				}
			}
		})
	}
}

func TestPhase1CustomerWordTemplateLibreOfficeProducesExpectedPages(t *testing.T) {
	executable := os.Getenv("LIBREOFFICE_INTEGRATION_PATH")
	if executable == "" {
		t.Skip("LIBREOFFICE_INTEGRATION_PATH is not configured")
	}
	templatePath := os.Getenv("LIBREOFFICE_INTEGRATION_TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "../../configs/export-templates/competency-phase1-report.docx"
	}
	template, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatal(err)
	}
	tokens, charts, err := buildPhase1WordTemplateData(phase1WordTestData())
	if err != nil {
		t.Fatal(err)
	}
	docx, err := renderPhase1WordTemplate(template, tokens, charts)
	if err != nil {
		t.Fatal(err)
	}
	artifactDir := os.Getenv("LIBREOFFICE_INTEGRATION_ARTIFACT_DIR")
	if artifactDir != "" {
		if err := os.MkdirAll(artifactDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(artifactDir, "phase1-report.docx"), docx, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	convertCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	pdf, err := libreofficepdf.NewClient(executable).Convert(convertCtx, "phase1-report.docx", docx)
	if err != nil {
		t.Fatalf("LibreOffice conversion: %v", err)
	}
	if artifactDir != "" {
		if err := os.WriteFile(filepath.Join(artifactDir, "phase1-report.pdf"), pdf, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	pageCount, err := pdfapi.PageCount(bytes.NewReader(pdf), pdfmodel.NewDefaultConfiguration())
	if err != nil {
		t.Fatalf("read converted PDF: %v", err)
	}
	expectedPages := 11
	if configured := os.Getenv("LIBREOFFICE_INTEGRATION_EXPECTED_PAGES"); configured != "" {
		parsed, parseErr := strconv.Atoi(configured)
		if parseErr != nil || parsed <= 0 {
			t.Fatalf("invalid LIBREOFFICE_INTEGRATION_EXPECTED_PAGES=%q", configured)
		}
		expectedPages = parsed
	}
	if pageCount != expectedPages {
		t.Fatalf("LibreOffice PDF pages=%d, want %d", pageCount, expectedPages)
	}
}

type capturingPhase1Converter struct {
	fileName string
	docx     []byte
	pdf      []byte
}

func (c *capturingPhase1Converter) Convert(_ context.Context, fileName string, docx []byte) ([]byte, error) {
	c.fileName = fileName
	c.docx = append([]byte(nil), docx...)
	return append([]byte(nil), c.pdf...), nil
}

func phase1WordTestData() map[string]any {
	groups := []service.Phase1ReportGroup{
		{GroupCode: "general_ability", GroupName: "通用能力", GroupScore: decimalPtr("3.75"), LevelCode: "L4"},
		{GroupCode: "psychological_quality", GroupName: "心理素养", GroupScore: decimalPtr("3.5"), LevelCode: "L4"},
	}
	dimensionIDs := []string{"competency-a1-01", "competency-a1-02", "competency-a1-03", "competency-a1-04", "competency-a1-05", "competency-b1-01", "competency-b1-02", "competency-b1-03", "competency-b1-04", "competency-b1-05"}
	dimensions := make([]service.Phase1ReportDimension, 0, len(dimensionIDs))
	texts := make(map[string]string, len(dimensionIDs))
	for index, id := range dimensionIDs {
		score := "3.50"
		if index == 0 {
			score = "3.75"
		}
		dimensions = append(dimensions, service.Phase1ReportDimension{DimensionID: id, DimensionCode: fmt.Sprintf("D%02d", index+1), DimensionName: id, DimensionScore: decimalPtr(score), LevelCode: "L4"})
		texts[id] = id + "诊断"
	}
	texts["competency-a1-01"] = "逻辑思维诊断"
	return map[string]any{
		"reportKind": "frontline_phase1",
		"result": map[string]any{
			"participantName": "测试人员", "participantAge": 30, "participantGender": "0", "participantTelephone": "13800000000",
			"participantAffiliation": "测试单位", "participantPost": "测试岗位", "overallScore": 35, "overallLevel": "qualified", "submittedAt": "2026-08-12T10:00:00+08:00",
		},
		"groups":     groups,
		"dimensions": dimensions,
		"validity":   map[string]any{"status": "good", "notice": "效度良好"},
		"reportText": service.Phase1ReportTextSnapshot{Disclaimer: "正式免责声明", OverallText: "总体诊断", GroupTexts: map[string]string{"general_ability": "通用能力说明", "psychological_quality": "心理素养说明"}, DimensionTexts: texts, ValidityText: "效度良好"},
		"meta":       map[string]any{"generatedAt": "2026-08-12T11:00:00+08:00", "userTime": 20, "dimensionCoreMeanings": map[string]string{"competency-a1-01": "逻辑定义"}},
	}
}

func decimalPtr(value string) *decimal.Decimal {
	parsed := decimal.RequireFromString(value)
	return &parsed
}

func makeWordTemplateFixture(t *testing.T) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)
	for name, content := range map[string]string{
		"word/document.xml":      `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>{{participant.name}}</w:t></w:r></w:p></w:body></w:document>`,
		"word/charts/chart1.xml": `<c:chartSpace xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart"><c:numCache><c:pt idx="0"><c:v>1</c:v></c:pt><c:pt idx="1"><c:v>4</c:v></c:pt></c:numCache></c:chartSpace>`,
	} {
		entry, err := writer.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len())); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func readWordPart(t *testing.T, docx []byte, partName string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range reader.File {
		if file.Name != partName {
			continue
		}
		rc, openErr := file.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		body, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		return body
	}
	t.Fatalf("Word part missing: %s", partName)
	return nil
}

func replaceWordFixturePart(t *testing.T, docx []byte, partName string, replacement []byte) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatal(err)
	}
	output := new(bytes.Buffer)
	writer := zip.NewWriter(output)
	for _, file := range reader.File {
		rc, openErr := file.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		body, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		if file.Name == partName {
			body = replacement
		}
		entry, createErr := writer.CreateHeader(&zip.FileHeader{Name: file.Name, Method: zip.Deflate})
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := entry.Write(body); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

var _ = xml.EscapeText
