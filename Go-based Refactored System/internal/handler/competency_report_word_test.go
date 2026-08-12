package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/service"
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
	tokens, charts, err := buildPhase1WordTemplateData(phase1WordTestData())
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

var _ = xml.EscapeText
