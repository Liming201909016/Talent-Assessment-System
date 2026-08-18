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
	"regexp"
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
	if got := charts["chart.group.overview"]; len(got) != 2 || got[0] != 3.75 || got[1] != 3.5 {
		t.Fatalf("group chart values=%v", got)
	}
	if got := charts["chart.dimension.radar"]; len(got) != 10 || got[0] != 3.75 || got[9] != 3.5 {
		t.Fatalf("radar values=%v", got)
	}
	if got := charts["chart.dimension.competency-a1-01"]; len(got) != 2 || got[0] != 3.75 || got[1] != 1.25 {
		t.Fatalf("first doughnut values=%v", got)
	}
}

func TestPhase1WordTemplateReplacesTokensAndChartCaches(t *testing.T) {
	template := makeWordTemplateFixture(t)
	tokens := map[string]string{"{{participant.name}}": "张三 & 李四"}
	charts := map[string][]float64{"chart.group.overview": {3.75, 1.25}}
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

// TestBugFB140_LibreOfficeChartLabelsUseRenderedPixelCalibration
// 对应：docs/regression-tests.md #FB-140
// 复现：Word中环形图数字居中，但LibreOffice 24.2生成PDF后十图出现不同方向的偏移。
// 期望：仅LibreOffice转换前按实测绘图区比例校准chart3-chart12标签坐标。
func TestBugFB140_LibreOfficeChartLabelsUseRenderedPixelCalibration(t *testing.T) {
	buffer := new(bytes.Buffer)
	writer := zip.NewWriter(buffer)
	for index := 3; index <= 12; index++ {
		part, err := writer.Create(fmt.Sprintf("word/charts/chart%d.xml", index))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(part, `<c:chart><c:dLbl><c:idx val="0"/><c:layout><c:manualLayout><c:x val="-0.1"/><c:y val="-0.2"/></c:manualLayout></c:layout></c:dLbl></c:chart>`); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	calibrated, err := calibratePhase1LibreOfficeChartLabels(buffer.Bytes())
	if err != nil {
		t.Fatalf("calibrate LibreOffice labels: %v", err)
	}
	chart3 := string(readWordPart(t, calibrated, "word/charts/chart3.xml"))
	if !strings.Contains(chart3, `<c:x val="-0.122816985645933"/>`) ||
		!strings.Contains(chart3, `<c:y val="-0.0980113636363636"/>`) {
		t.Fatalf("chart3 calibration mismatch: %s", chart3)
	}
	chart11 := string(readWordPart(t, calibrated, "word/charts/chart11.xml"))
	if !strings.Contains(chart11, `<c:x val="-0.0934210526315789"/>`) ||
		!strings.Contains(chart11, `<c:y val="-0.0162878787878788"/>`) {
		t.Fatalf("chart11 calibration mismatch: %s", chart11)
	}
}

func TestPhase1WordTemplateRejectsMissingRequiredToken(t *testing.T) {
	template := makeWordTemplateFixture(t)
	if _, err := renderPhase1WordTemplate(template, map[string]string{"{{participant.name}}": "张三", "{{participant.age}}": "30"}, map[string][]float64{"chart.group.overview": {3.75, 1.25}}); err == nil {
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
// 期望：页面只显示正常示例值，51个内容控件使用49个稳定字段键；仅已注册可重复字段允许重复。
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
	if len(tags) != 51 {
		t.Fatalf("Word content-control tags=%d, want 51", len(tags))
	}
	seen := make(map[string]int, len(tags))
	for _, match := range tags {
		tag := string(match[1])
		seen[tag]++
		definition, registered := phase1FieldDefinition(tag)
		if !registered || (seen[tag] > 1 && !definition.Repeatable) {
			t.Fatalf("duplicate Word content-control tag: %s", tag)
		}
	}
	for _, required := range []string{"participant.name", "overall.level", "validity.notice", "dimension.competency-b1-05.diagnosis"} {
		if seen[required] == 0 {
			t.Fatalf("required Word content-control tag missing: %s", required)
		}
	}
	for _, repeated := range []string{"group.general_ability.score", "group.psychological_quality.score"} {
		if seen[repeated] != 2 {
			t.Fatalf("repeatable Word content-control tag %s count=%d, want 2", repeated, seen[repeated])
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
	if contract.ContentControls != 51 || contract.UsedFields != 49 || contract.Charts != 12 || contract.VisibleTokens != 0 {
		t.Fatalf("contract=%+v", contract)
	}

	document := string(readWordPart(t, template, "word/document.xml"))
	document = strings.Replace(document, `<w:tag w:val="dimension.competency-a1-04.diagnosis"/>`, `<w:tag w:val="dimension.competency-a1-03.diagnosis"/>`, 1)
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
	if value, _ := book.GetCellValue("FieldDictionary", "B1"); value != phase1WordTemplateSchemaV2 {
		t.Fatalf("FieldDictionary B1=%q", value)
	}
	if value, _ := book.GetCellValue("ChartData", "A13"); value != "competency-b1-05" {
		t.Fatalf("ChartData A13=%q", value)
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

// TestBugFB122_V2TemplateUsesWordCompatibleOPCContentTypes
// 对应：docs/regression-tests.md #FB-122
// 复现：V2生成器把xlsx Default追加到所有Override之后，LibreOffice容忍但Microsoft Word拒绝整个DOCX。
// 期望：内嵌xlsx使用显式Override，且Content Types中所有Default均位于第一个Override之前。
func TestBugFB122_V2TemplateUsesWordCompatibleOPCContentTypes(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-embedded.docx")
	if err != nil {
		t.Fatal(err)
	}
	contentTypes := string(readWordPart(t, template, "[Content_Types].xml"))
	firstOverride := strings.Index(contentTypes, "<Override")
	lastDefault := strings.LastIndex(contentTypes, "<Default")
	if firstOverride < 0 || lastDefault < 0 || lastDefault > firstOverride {
		t.Fatalf("OPC content type order is invalid: lastDefault=%d firstOverride=%d", lastDefault, firstOverride)
	}
	workbookOverride := `<Override PartName="/word/embeddings/competency-phase1-chart-data.xlsx" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"/>`
	if !strings.Contains(contentTypes, workbookOverride) {
		t.Fatal("embedded workbook content type override is missing")
	}
	if strings.Contains(contentTypes, `<Default Extension="xlsx"`) {
		t.Fatal("embedded workbook must not add a late global xlsx default")
	}
	brokenContentTypes := strings.Replace(contentTypes, workbookOverride, `<Default Extension="xlsx" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"/>`, 1)
	broken := replaceWordFixturePart(t, template, "[Content_Types].xml", []byte(brokenContentTypes))
	if _, err := validatePhase1WordTemplateUpload(broken); err == nil || !strings.Contains(err.Error(), "Content Types") {
		t.Fatalf("invalid OPC content type order error=%v", err)
	}
}

func TestPhase1TemplateV2RegistryIsTransparentAndExtensible(t *testing.T) {
	fields := phase1TemplateFieldRegistry()
	seen := make(map[string]bool, len(fields))
	for _, field := range fields {
		if field.Key == "" || field.Label == "" || field.ValueType == "" || field.Description == "" || field.Resolve == nil {
			t.Fatalf("incomplete field definition: %+v", field)
		}
		if seen[field.Key] {
			t.Fatalf("duplicate field key: %s", field.Key)
		}
		seen[field.Key] = true
	}
	for _, key := range []string{
		"participant.name", "participant.degree", "participant.major", "overall.score", "overall.percentage",
		"dimension.competency-a1-01.score", "dimension.competency-a1-01.remainingScore",
	} {
		if !seen[key] {
			t.Fatalf("registered field missing: %s", key)
		}
	}
	if len(fields) <= len(requiredPhase1WordTemplateTags()) {
		t.Fatalf("registered fields=%d, want optional extension fields beyond required contract", len(fields))
	}

	charts := phase1TemplateChartRegistry()
	if len(charts) != 12 {
		t.Fatalf("business charts=%d, want 12", len(charts))
	}
	if charts[0].Key != "chart.group.overview" || charts[1].Key != "chart.dimension.radar" || charts[2].Key != "chart.dimension.competency-a1-01" {
		t.Fatalf("unexpected business chart registry: %+v", charts[:3])
	}
}

func TestPhase1TemplateV2EmbeddedDictionaryAndBusinessCharts(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-embedded.docx")
	if err != nil {
		t.Fatal(err)
	}
	contract, err := validatePhase1WordTemplateUpload(template)
	if err != nil {
		t.Fatalf("V2 template rejected: %v", err)
	}
	if contract.SchemaVersion != phase1WordTemplateSchemaV2 || contract.BusinessCharts != 12 || contract.EmbeddedWorkbooks != 1 || contract.ExternalLinks != 0 {
		t.Fatalf("V2 contract=%+v", contract)
	}
	if contract.RegisteredFields != len(phase1TemplateFieldRegistry()) || contract.UsedFields != 49 {
		t.Fatalf("V2 field contract=%+v", contract)
	}

	workbook := readWordPart(t, template, phase1ChartWorkbookPath)
	book, err := excelize.OpenReader(bytes.NewReader(workbook))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	if value, _ := book.GetCellValue("FieldDictionary", "B1"); value != phase1WordTemplateSchemaV2 {
		t.Fatalf("FieldDictionary schema=%q", value)
	}
	dictionaryKeys := make(map[string]bool)
	rows, err := book.GetRows("FieldDictionary")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if len(row) > 0 {
			dictionaryKeys[row[0]] = true
		}
	}
	for _, field := range phase1TemplateFieldRegistry() {
		if !dictionaryKeys[field.Key] {
			t.Fatalf("FieldDictionary missing %s", field.Key)
		}
	}
	for cell, expected := range map[string]string{"A2": "group.general_ability", "A3": "group.psychological_quality", "A4": "competency-a1-01", "A13": "competency-b1-05"} {
		if value, _ := book.GetCellValue("ChartData", cell); value != expected {
			t.Fatalf("ChartData %s=%q, want %q", cell, value, expected)
		}
	}

	parts := readPhase1WordPartsForTest(t, template)
	resolved, err := resolvePhase1BusinessChartParts(parts)
	if err != nil {
		t.Fatal(err)
	}
	for _, chart := range phase1TemplateChartRegistry() {
		if resolved[chart.Key] == "" {
			t.Fatalf("business chart is not resolved: %s", chart.Key)
		}
	}
}

func TestPhase1TemplateV1PhysicalChartMappingRemainsCompatible(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	contract, err := validatePhase1WordTemplateUpload(template)
	if err != nil {
		t.Fatal(err)
	}
	if contract.SchemaVersion != phase1WordTemplateSchemaV1 || contract.BusinessCharts != 0 {
		t.Fatalf("V1 contract=%+v", contract)
	}
	tokens, charts, err := buildPhase1WordTemplateData(phase1WordTestData())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := renderPhase1WordTemplate(template, tokens, charts); err != nil {
		t.Fatalf("V1 template is no longer compatible: %v", err)
	}
}

func TestPhase1TemplateV2AcceptsOptionalAndRepeatableRegisteredFields(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-embedded.docx")
	if err != nil {
		t.Fatal(err)
	}
	document := string(readWordPart(t, template, "word/document.xml"))
	extraControls := `<w:p><w:sdt><w:sdtPr><w:alias w:val="overall.score"/><w:tag w:val="overall.score"/><w:text/></w:sdtPr><w:sdtContent><w:r><w:t>35.00</w:t></w:r></w:sdtContent></w:sdt></w:p>` +
		`<w:p><w:sdt><w:sdtPr><w:alias w:val="participant.name"/><w:tag w:val="participant.name"/><w:text/></w:sdtPr><w:sdtContent><w:r><w:t>张三</w:t></w:r></w:sdtContent></w:sdt></w:p>`
	document = strings.Replace(document, "</w:body>", extraControls+"</w:body>", 1)
	extended := replaceWordFixturePart(t, template, "word/document.xml", []byte(document))
	contract, err := validatePhase1WordTemplateUpload(extended)
	if err != nil {
		t.Fatalf("registered optional/repeatable fields rejected: %v", err)
	}
	if contract.ContentControls != 51 || contract.UsedFields != 50 || contract.RegisteredFields != 75 {
		t.Fatalf("extended V2 contract=%+v", contract)
	}
	tokens, charts, err := buildPhase1WordTemplateData(phase1WordTestData())
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderPhase1WordTemplate(extended, tokens, charts)
	if err != nil {
		t.Fatal(err)
	}
	renderedDocument := string(readWordPart(t, rendered, "word/document.xml"))
	if strings.Count(renderedDocument, ">测试人员</w:t>") < 2 || !strings.Contains(renderedDocument, ">35.00</w:t>") {
		t.Fatal("optional or repeated field was not populated")
	}
}

func TestPhase1TemplateV2RejectsInvalidFieldsChartsAndWorkbookLinks(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-embedded.docx")
	if err != nil {
		t.Fatal(err)
	}
	document := string(readWordPart(t, template, "word/document.xml"))
	tests := []struct {
		name     string
		part     string
		replace  func(string) string
		wantText string
	}{
		{name: "unknown field", part: "word/document.xml", replace: func(value string) string {
			return strings.Replace(value, `<w:tag w:val="participant.name"/>`, `<w:tag w:val="participant.unknown"/>`, 1)
		}, wantText: "未支持的内容控件Tag"},
		{name: "missing required field", part: "word/document.xml", replace: func(value string) string {
			return strings.Replace(value, `<w:tag w:val="participant.name"/>`, `<w:tag w:val="participant.degree"/>`, 1)
		}, wantText: "缺少内容控件Tag"},
		{name: "unknown chart", part: "word/document.xml", replace: func(value string) string {
			return strings.Replace(value, `title="chart.group.overview"`, `title="chart.group.unknown"`, 1)
		}, wantText: "未知图表业务键"},
		{name: "duplicate chart", part: "word/document.xml", replace: func(value string) string {
			return strings.Replace(value, `title="chart.dimension.radar"`, `title="chart.group.overview"`, 1)
		}, wantText: "重复图表业务键"},
		{name: "missing chart", part: "word/document.xml", replace: func(value string) string {
			return strings.Replace(value, `title="chart.dimension.radar"`, `title="decorative.radar"`, 1)
		}, wantText: "缺少图表业务键"},
		{name: "unexpected chart range", part: "word/charts/chart1.xml", replace: func(value string) string {
			return strings.Replace(value, "ChartData!$C$2:$C$3", "ChartData!$C$20:$C$21", 1)
		}, wantText: "图表数据区域无效"},
		{name: "external workbook", part: "word/charts/_rels/chart1.xml.rels", replace: func(value string) string {
			return strings.Replace(value, `Target="../embeddings/competency-phase1-chart-data.xlsx"`, `Target="file:///C:/chart-data.xlsx" TargetMode="External"`, 1)
		}, wantText: "图表内嵌工作簿关系无效"},
	}
	if !strings.Contains(document, `title="chart.group.overview"`) {
		t.Fatal("V2 business chart title fixture missing")
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := string(readWordPart(t, template, test.part))
			broken := replaceWordFixturePart(t, template, test.part, []byte(test.replace(body)))
			if _, err := validatePhase1WordTemplateUpload(broken); err == nil || !strings.Contains(err.Error(), test.wantText) {
				t.Fatalf("error=%v, want %q", err, test.wantText)
			}
		})
	}
}

func TestResolvePhase1BusinessChartPartsDoesNotDependOnPhysicalChartNumber(t *testing.T) {
	var document strings.Builder
	var relationships strings.Builder
	parts := map[string][]byte{}
	for index, chart := range phase1TemplateChartRegistry() {
		relationID := fmt.Sprintf("rId%d", index+1)
		part := fmt.Sprintf("word/charts/customer-chart-%02d.xml", 12-index)
		document.WriteString(fmt.Sprintf(`<wp:inline><wp:docPr id="%d" name="chart" title="%s"/><c:chart r:id="%s"/></wp:inline>`, index+1, chart.Key, relationID))
		relationships.WriteString(fmt.Sprintf(`<Relationship Id="%s" Target="charts/customer-chart-%02d.xml"/>`, relationID, 12-index))
		parts[part] = []byte("chart")
	}
	parts["word/document.xml"] = []byte(document.String())
	parts["word/_rels/document.xml.rels"] = []byte("<Relationships>" + relationships.String() + "</Relationships>")
	resolved, err := resolvePhase1BusinessChartParts(parts)
	if err != nil {
		t.Fatal(err)
	}
	for index, chart := range phase1TemplateChartRegistry() {
		expected := fmt.Sprintf("word/charts/customer-chart-%02d.xml", 12-index)
		if resolved[chart.Key] != expected {
			t.Fatalf("%s=%q, want %q", chart.Key, resolved[chart.Key], expected)
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
	for cell, expected := range map[string]string{"C2": "3.75", "C3": "3.5", "C4": "3.75", "D4": "1.25", "C13": "3.5", "D13": "1.5"} {
		if value, _ := book.GetCellValue("ChartData", cell); value != expected {
			t.Fatalf("ChartData %s=%q, want %q", cell, value, expected)
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
	renderer := &phase1WordReportRenderer{templatePath: "../../configs/export-templates/competency-phase1-report.docx", converter: converter, timeout: time.Second, calibrateLabels: true}
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
	docx, err = calibratePhase1LibreOfficeChartLabels(docx)
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
	expectedPages := 10
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

// TestBugFB127_Phase1DimensionChartsAreInlineInsideScoreCells
// 对应：docs/regression-tests.md #FB-127
// 复现：逻辑思维的3.50环形图浮到图示说明上方，维度结果表格左侧得分单元格为空。
// 期望：chart3-chart12均为各自表格单元格内的wp:inline对象，不使用页面级浮动锚点。
func TestBugFB127_Phase1DimensionChartsAreInlineInsideScoreCells(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	document := string(readWordPart(t, template, "word/document.xml"))
	var relationships struct {
		Items []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(readWordPart(t, template, "word/_rels/document.xml.rels"), &relationships); err != nil {
		t.Fatal(err)
	}
	IDs := make(map[string]string, 10)
	for _, relationship := range relationships.Items {
		IDs[strings.ReplaceAll(relationship.Target, "\\", "/")] = relationship.ID
	}
	for chartIndex := 3; chartIndex <= 12; chartIndex++ {
		target := fmt.Sprintf("charts/chart%d.xml", chartIndex)
		relationID := IDs[target]
		if relationID == "" {
			t.Fatalf("%s relationship missing", target)
		}
		marker := `r:id="` + relationID + `"`
		chartAt := strings.Index(document, marker)
		if chartAt < 0 {
			t.Fatalf("%s drawing missing", target)
		}
		inlineAt := strings.LastIndex(document[:chartAt], "<wp:inline")
		anchorAt := strings.LastIndex(document[:chartAt], "<wp:anchor")
		if inlineAt < 0 || inlineAt < anchorAt {
			t.Fatalf("%s is not an inline drawing inside its score cell", target)
		}
		cellAt := strings.LastIndex(document[:inlineAt], "<w:tc")
		cellEndAt := strings.LastIndex(document[:inlineAt], "</w:tc>")
		if cellAt < 0 || cellAt < cellEndAt {
			t.Fatalf("%s is outside a table cell", target)
		}
	}
}

// TestBugFB128_Phase1GroupSummaryScoresUseDynamicControls
// 对应：docs/regression-tests.md #FB-128
// 复现：一级汇总行固定显示3.75/3.70，但同页分析区和数据库实际为3.50/3.60。
// 期望：两个一级得分Tag各出现两次，汇总行和分析区由同一冻结数据同步填充。
func TestBugFB128_Phase1GroupSummaryScoresUseDynamicControls(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	document := string(readWordPart(t, template, "word/document.xml"))
	for _, tag := range []string{"group.general_ability.score", "group.psychological_quality.score"} {
		if got := strings.Count(document, `<w:tag w:val="`+tag+`"/>`); got != 2 {
			t.Fatalf("%s controls=%d, want 2", tag, got)
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
	renderedDocument := string(readWordPart(t, rendered, "word/document.xml"))
	for _, value := range []string{"3.75", "3.50"} {
		if got := strings.Count(renderedDocument, ">"+value+"</w:t>"); got < 2 {
			t.Fatalf("rendered group score %s count=%d, want at least 2", value, got)
		}
	}
}

// TestCustomerChoice_Phase1GroupOverviewRetainsThreeDimensionalPie
// 对应：docs/regression-tests.md #FB-129
// 客户选择：保留客户模板的3D一级总览饼图，但标签必须显示两位小数且不得显示百分比。
func TestCustomerChoice_Phase1GroupOverviewRetainsThreeDimensionalPie(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	chart := string(readWordPart(t, template, "word/charts/chart1.xml"))
	for _, required := range []string{
		`<c:pie3DChart>`,
		`<c:numFmt formatCode="0.00" sourceLinked="0"/>`, `<c:showPercent val="0"/>`,
	} {
		if !strings.Contains(chart, required) {
			t.Fatalf("customer first-level overview chart missing %s", required)
		}
	}
}

// TestBugFB130_Phase1ProfileTableHonorsRequiredFields
// 对应：docs/regression-tests.md #FB-130
// 复现：测评只配置姓名、手机号，PDF仍显示年龄/性别/单位/岗位标签并留空。
// 期望：运行时删除未配置字段的行/单元格，只保留姓名、手机号及固定时间信息。
func TestBugFB130_Phase1ProfileTableHonorsRequiredFields(t *testing.T) {
	data := phase1WordTestData()
	meta := data["meta"].(map[string]any)
	meta["requiredFields"] = "name,telephone"
	tokens, charts, err := buildPhase1WordTemplateData(data)
	if err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderPhase1WordTemplate(template, tokens, charts)
	if err != nil {
		t.Fatal(err)
	}
	document := string(readWordPart(t, rendered, "word/document.xml"))
	for _, required := range []string{"姓名：", "手机号：", "时间：", "时长：", ">测试人员</w:t>", ">13800000000</w:t>"} {
		if !strings.Contains(document, required) {
			t.Fatalf("required profile content missing: %s", required)
		}
	}
	for _, absent := range []string{"年龄：", "性别：", "单位：", "岗位：", `w:val="participant.age"`, `w:val="participant.gender"`, `w:val="participant.affiliation"`, `w:val="participant.post"`} {
		if strings.Contains(document, absent) {
			t.Fatalf("unconfigured profile content remains: %s", absent)
		}
	}
}

// TestBugFB131_Phase1TemplateUsesWordCompatibleOOXML
// 对应：docs/regression-tests.md #FB-131
// 复现：LibreOffice可转换模板，但Microsoft Word因严格OOXML错误拒绝打开。
// 期望：清除LO导出的非法start枚举、pPr属性顺序、重复VML ID及非法charset属性。
func TestBugFB131_Phase1TemplateUsesWordCompatibleOOXML(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report.docx")
	if err != nil {
		t.Fatal(err)
	}
	document := string(readWordPart(t, template, "word/document.xml"))
	numbering := string(readWordPart(t, template, "word/numbering.xml"))
	fontTable := string(readWordPart(t, template, "word/fontTable.xml"))
	for part, value := range map[string]string{"document": document, "numbering": numbering} {
		if strings.Contains(value, `w:val="start"`) {
			t.Fatalf("%s contains Word-invalid start enumeration", part)
		}
	}
	if strings.Contains(fontTable, "w:characterSet=") {
		t.Fatal("font table contains Word-invalid characterSet attribute")
	}
	ids := make(map[string]bool)
	for _, match := range regexp.MustCompile(`<v:[^ >]+[^>]*\bid="([^"]+)"`).FindAllStringSubmatch(document, -1) {
		if ids[match[1]] {
			t.Fatalf("duplicate VML id: %s", match[1])
		}
		ids[match[1]] = true
	}
	for _, properties := range regexp.MustCompile(`(?s)<w:pPr\b.*?</w:pPr>`).FindAllString(document, -1) {
		shading := strings.Index(properties, "<w:shd")
		if shading < 0 {
			continue
		}
		for _, later := range []string{"<w:spacing", "<w:ind", "<w:jc", "<w:rPr"} {
			if position := strings.Index(properties, later); position >= 0 && shading > position {
				t.Fatalf("paragraph shading has Word-invalid order after %s", later)
			}
		}
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
		"meta":       map[string]any{"generatedAt": "2026-08-12T11:00:00+08:00", "userTime": 20, "requiredFields": "name,telephone", "dimensionCoreMeanings": map[string]string{"competency-a1-01": "逻辑定义"}},
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

func readPhase1WordPartsForTest(t *testing.T, docx []byte) map[string][]byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatal(err)
	}
	parts := make(map[string][]byte, len(reader.File))
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
		parts[file.Name] = body
	}
	return parts
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
