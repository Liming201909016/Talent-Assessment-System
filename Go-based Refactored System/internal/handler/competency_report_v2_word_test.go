package handler

import (
	"archive/zip"
	"bytes"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestBugUF056_Phase1V2CoverHidesDuration
// 对应：docs/regression-tests.md #UF-056
// 复现：production正式PDF首页仍显示“时长：1分钟”。
// 期望：v2模板和渲染结果保留内部result.userTime合同，但整项在Word/PDF中不可见。
func TestBugUF056_Phase1V2CoverHidesDuration(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-v2.docx")
	if err != nil {
		t.Fatal(err)
	}
	document := string(readWordPart(t, template, "word/document.xml"))
	if strings.Contains(document, ">时长：</w:t>") || strings.Contains(document, ">分钟</w:t>") {
		t.Fatal("v2 template still exposes the cover duration label or unit")
	}
	control := regexp.MustCompile(`(?s)<w:sdt>.*?<w:tag w:val="result\.userTime".*?</w:sdt>`).FindString(document)
	if control == "" || !strings.Contains(control, "<w:vanish") {
		t.Fatal("v2 template duration contract is absent or visible")
	}
	fields := phase1V2WordTestFields(t, template)
	fields["result.userTime"] = "1"
	rendered, err := renderPhase1V2WordTemplate(template, fields, phase1V2WordTestCharts(), "")
	if err != nil {
		t.Fatal(err)
	}
	renderedDocument := string(readWordPart(t, rendered, "word/document.xml"))
	if strings.Contains(renderedDocument, ">时长：</w:t>") || strings.Contains(renderedDocument, ">分钟</w:t>") {
		t.Fatal("rendered v2 report exposes the cover duration label or unit")
	}
	renderedControl := regexp.MustCompile(`(?s)<w:sdt>.*?<w:tag w:val="result\.userTime".*?</w:sdt>`).FindString(renderedDocument)
	if renderedControl == "" || !strings.Contains(renderedControl, "<w:vanish") {
		t.Fatal("rendered v2 duration contract became visible")
	}
	if strings.Contains(renderedControl, ">1</w:t>") {
		t.Fatal("rendered v2 duration control still contains the numeric duration")
	}
}

// TestBugFB179_Phase1V2WordRendererIsValueOnly
// Corresponds to docs/regression-tests.md FB-179.
func TestBugFB179_Phase1V2WordRendererIsValueOnly(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-v2.docx")
	if err != nil {
		t.Fatal(err)
	}
	original := append([]byte(nil), template...)
	fields := phase1V2WordTestFields(t, template)
	fields["participant.name"] = "李四 & <测试>"
	fields["validity.status"] = "有效"
	fields["overall.score"] = "68.13"
	fields["rule.strength.3"] = ""
	charts := phase1V2WordTestCharts()

	rendered, err := renderPhase1V2WordTemplate(template, fields, charts, "")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(template, original) {
		t.Fatal("renderer mutated input template")
	}
	if output := os.Getenv("PHASE1_V2_RENDER_OUTPUT"); output != "" {
		if err := os.WriteFile(output, rendered, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	document := string(readWordPart(t, rendered, "word/document.xml"))
	headers := ""
	for _, name := range wordPartNames(t, rendered, `^word/header\d+\.xml$`) {
		headers += string(readWordPart(t, rendered, name))
	}
	for _, required := range []string{"李四 &amp; &lt;测试&gt;", "有效", "68.13"} {
		if !strings.Contains(document+headers, required) {
			t.Fatalf("rendered body/header missing %q", required)
		}
	}
	if strings.Contains(document, "strength:3") {
		t.Fatal("empty predefined slot retained its sample value")
	}

	for _, name := range wordPartNames(t, template, `^(word/document\.xml|word/header\d+\.xml)$`) {
		before := maskPhase1V2ControlValues(t, readWordPart(t, template, name))
		after := maskPhase1V2ControlValues(t, readWordPart(t, rendered, name))
		if !bytes.Equal(before, after) {
			t.Fatalf("non-value Word XML changed: %s", name)
		}
	}
	for _, name := range wordPartNames(t, template, `^word/charts/chart\d+\.xml$`) {
		beforeRaw := readWordPart(t, template, name)
		afterRaw := readWordPart(t, rendered, name)
		if bytes.Equal(beforeRaw, phase1V2ChartByKey(t, template, "chart.dimension.comparison")) {
			beforeRaw = phase1V2AnyChartPointPattern.ReplaceAll(beforeRaw, []byte(`POINT_COLORS`))
			afterRaw = phase1V2AnyChartPointPattern.ReplaceAll(afterRaw, []byte(`POINT_COLORS`))
		}
		before := phase1V2AnyChartValuePattern.ReplaceAll(beforeRaw, []byte(`<c:v>VALUE</c:v>`))
		after := phase1V2AnyChartValuePattern.ReplaceAll(afterRaw, []byte(`<c:v>VALUE</c:v>`))
		if !bytes.Equal(before, after) {
			t.Fatalf("chart style/structure changed: %s", name)
		}
	}
	for _, name := range wordPartNames(t, template, `.*`) {
		if name == "word/document.xml" || strings.HasPrefix(name, "word/header") || strings.HasPrefix(name, "word/footer") || strings.HasPrefix(name, "word/charts/chart") {
			continue
		}
		if !bytes.Equal(readWordPart(t, template, name), readWordPart(t, rendered, name)) {
			t.Fatalf("unrelated template part changed: %s", name)
		}
	}

	overallChart := string(phase1V2ChartByKey(t, rendered, "chart.overall.score"))
	for _, value := range []string{"68.125", "31.875"} {
		if !strings.Contains(overallChart, ">"+value+"<") {
			t.Fatalf("overall chart missing %s", value)
		}
	}
	comparisonChart := string(phase1V2ChartByKey(t, rendered, "chart.dimension.comparison"))
	if !strings.Contains(comparisonChart, ">78.125<") || !strings.Contains(comparisonChart, ">57.5<") {
		t.Fatalf("comparison chart values missing")
	}
}

// TestBugFB193_V2OverviewPreservesLabelAndBodyStyles
// Corresponds to docs/regression-tests.md FB-193.
func TestBugFB193_V2OverviewPreservesLabelAndBodyStyles(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-v2.docx")
	if err != nil {
		t.Fatal(err)
	}
	fields := phase1V2WordTestFields(t, template)
	styledValues := map[string]string{
		"rule.strength.1":    "计划执行：能根据目标制定清晰计划，合理安排时间和资源，并按步骤推进任务落实。",
		"rule.strength.2":    "求真务实：可以做到以事实和效果为依据，核实关键数据和信息来源，确保工作建立在客观事实之上。",
		"rule.strength.3":    "自律性：主动约束自身行为，按计划推进工作，面对常规任务时能保持专注。",
		"rule.development.1": "成就导向：愿意尽力达到既定标准和要求，但面对挑战性工作时常感到焦虑、把握不大，不过仍能努力完成任务。",
		"rule.development.2": "沟通表达：多数情况下能把想法表达清楚，让对方理解自己的意图，倾听时也能抓住主要内容，但表述偶尔不够精准。",
	}
	for key, value := range styledValues {
		fields[key] = value
	}
	rendered, err := renderPhase1V2WordTemplate(template, fields, phase1V2WordTestCharts(), "")
	if err != nil {
		t.Fatal(err)
	}
	if output := os.Getenv("PHASE1_V2_FB193_DOCX"); output != "" {
		if err := os.WriteFile(output, rendered, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	document := string(readWordPart(t, rendered, "word/document.xml"))
	for _, test := range []struct {
		tag, label, body string
	}{
		{"rule.strength.1", "计划执行：", strings.TrimPrefix(styledValues["rule.strength.1"], "计划执行：")},
		{"rule.strength.2", "求真务实：", strings.TrimPrefix(styledValues["rule.strength.2"], "求真务实：")},
		{"rule.strength.3", "自律性：", strings.TrimPrefix(styledValues["rule.strength.3"], "自律性：")},
		{"rule.development.1", "成就导向：", strings.TrimPrefix(styledValues["rule.development.1"], "成就导向：")},
		{"rule.development.2", "沟通表达：", strings.TrimPrefix(styledValues["rule.development.2"], "沟通表达：")},
	} {
		if strings.Contains(document, `w:val="`+test.tag+`"`) {
			t.Fatalf("rendered selected item retained its content-control wrapper: %s", test.tag)
		}
		if !strings.Contains(document, ">"+test.label+"<") || !strings.Contains(document, ">"+test.body+"<") {
			t.Fatalf("styled item missing values: %s", test.tag)
		}
		labelRun := ""
		bodyRun := ""
		for _, run := range phase1V2WordRunPattern.FindAllString(document, -1) {
			if strings.Contains(run, ">"+test.label+"<") {
				labelRun = run
			}
			if strings.Contains(run, ">"+test.body+"<") {
				bodyRun = run
			}
		}
		if labelRun == "" || !phase1V2WordBoldPattern.MatchString(labelRun) || bodyRun == "" || phase1V2WordBoldPattern.MatchString(bodyRun) {
			t.Fatalf("label/body style mismatch: %s", test.tag)
		}
		if !strings.Contains(labelRun, `<w:b w:val="true"/>`) || !strings.Contains(labelRun, `<w:bCs w:val="true"/>`) {
			t.Fatalf("label does not use explicit LibreOffice-compatible bold flags: %s", test.tag)
		}
	}
}

// TestBugFB196_V2ComparisonBarsUseFiveScoreBandColors
// Corresponds to docs/regression-tests.md FB-196.
func TestBugFB196_V2ComparisonBarsUseFiveScoreBandColors(t *testing.T) {
	values := []float64{90, 89.99, 70, 69.99, 30, 29.99, 10, 9.99, 0, 100}
	want := []string{"00A651", "38B86A", "38B86A", "A8D889", "A8D889", "F2A45F", "F2A45F", "E88937", "E88937", "00A651"}
	chart := []byte(`<c:chartSpace><c:barChart><c:ser><c:idx val="0"/><c:order val="0"/><c:spPr><a:solidFill><a:srgbClr val="ACD78D"/></a:solidFill></c:spPr><c:dPt><c:idx val="2"/><c:spPr><a:solidFill><a:srgbClr val="00B050"/></a:solidFill></c:spPr></c:dPt><c:val><c:numLit><c:ptCount val="10"/></c:numLit></c:val></c:ser></c:barChart><c:lineChart><c:ser><c:idx val="1"/><c:order val="1"/><c:spPr><a:ln><a:solidFill><a:srgbClr val="E67E32"/></a:solidFill></a:ln></c:spPr><c:val><c:numLit><c:ptCount val="10"/></c:numLit></c:val></c:ser></c:lineChart></c:chartSpace>`)
	colored, err := replacePhase1V2ComparisonBarColors(chart, values)
	if err != nil {
		t.Fatal(err)
	}
	content := string(colored)
	for index, color := range want {
		pattern := regexp.MustCompile(`(?s)<c:dPt><c:idx val="` + strconv.Itoa(index) + `"/>.*?<a:srgbClr val="` + color + `"/>.*?</c:dPt>`)
		if !pattern.MatchString(content) {
			t.Errorf("point %d missing score-band color %s", index, color)
		}
	}
	if strings.Count(content, "<c:dPt>") != 10 {
		t.Fatalf("colored points=%d, want 10", strings.Count(content, "<c:dPt>"))
	}

	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-v2.docx")
	if err != nil {
		t.Fatal(err)
	}
	charts := phase1V2WordTestCharts()
	rendered, err := renderPhase1V2WordTemplate(template, phase1V2WordTestFields(t, template), charts, "")
	if err != nil {
		t.Fatal(err)
	}
	beforeChart := phase1V2ChartByKey(t, template, "chart.dimension.comparison")
	afterChart := phase1V2ChartByKey(t, rendered, "chart.dimension.comparison")
	beforeSeries := phase1V2ChartSeriesPattern.FindAll(beforeChart, -1)
	afterSeries := phase1V2ChartSeriesPattern.FindAll(afterChart, -1)
	if len(beforeSeries) != 2 || len(afterSeries) != 2 {
		t.Fatalf("comparison series before/after=%d/%d", len(beforeSeries), len(afterSeries))
	}
	beforeLine := phase1V2AnyChartValuePattern.ReplaceAll(beforeSeries[1], []byte(`<c:v>VALUE</c:v>`))
	afterLine := phase1V2AnyChartValuePattern.ReplaceAll(afterSeries[1], []byte(`<c:v>VALUE</c:v>`))
	if !bytes.Equal(beforeLine, afterLine) {
		t.Fatal("dynamic bar coloring changed the template-controlled normal line")
	}
}

func TestBugFB179_Phase1V2WordRendererRejectsIncompleteContract(t *testing.T) {
	template, err := os.ReadFile("../../configs/export-templates/competency-phase1-report-v2.docx")
	if err != nil {
		t.Fatal(err)
	}
	fields := phase1V2WordTestFields(t, template)
	charts := phase1V2WordTestCharts()
	delete(fields, "overall.score")
	if _, err := renderPhase1V2WordTemplate(template, fields, charts, ""); err == nil {
		t.Fatal("missing field accepted")
	}
	fields = phase1V2WordTestFields(t, template)
	fields["unknown"] = "x"
	if _, err := renderPhase1V2WordTemplate(template, fields, charts, ""); err == nil {
		t.Fatal("unknown field accepted")
	}
	fields = phase1V2WordTestFields(t, template)
	delete(charts, "chart.overall.score")
	if _, err := renderPhase1V2WordTemplate(template, fields, charts, ""); err == nil {
		t.Fatal("missing chart accepted")
	}
	charts = phase1V2WordTestCharts()
	externalRelationship := []byte(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Target="file:///C:/customer.xlsx" TargetMode="External"/></Relationships>`)
	unsafeTemplate := replaceWordFixturePart(t, template, "word/charts/_rels/chart1.xml.rels", externalRelationship)
	if _, err := renderPhase1V2WordTemplate(unsafeTemplate, fields, charts, ""); err == nil {
		t.Fatal("external chart relationship accepted")
	}
}

var phase1V2AnyChartValuePattern = regexp.MustCompile(`<c:v>[^<]*</c:v>`)
var phase1V2AnyChartPointPattern = regexp.MustCompile(`(?s)(<c:dPt>.*?</c:dPt>)+`)

func phase1V2WordTestFields(t *testing.T, template []byte) map[string]string {
	t.Helper()
	fields := make(map[string]string)
	for _, name := range wordPartNames(t, template, `^(word/document\.xml|word/header\d+\.xml)$`) {
		for _, match := range wordContentControlTagPattern.FindAllSubmatch(readWordPart(t, template, name), -1) {
			fields[string(match[1])] = "value:" + string(match[1])
		}
	}
	return fields
}

func phase1V2WordTestCharts() map[string][][]float64 {
	scores := []float64{68.75, 62.5, 81.25, 50, 75, 56.25, 62.5, 78.125, 78.125, 68.75}
	norms := []float64{55, 65, 60, 55, 55, 52.5, 55, 57.5, 62.5, 60}
	charts := map[string][][]float64{
		"chart.overall.score":        {{68.125, 31.875}},
		"chart.dimension.comparison": {scores, norms},
	}
	keys := []string{"logical_reasoning", "plan_execution", "digital_application", "achievement_orientation", "continuous_learning", "communication", "cooperation", "truth_pragmatism", "self_discipline", "dedication"}
	for index, key := range keys {
		charts["chart.dimension."+key] = [][]float64{{scores[index], 100 - scores[index]}}
	}
	return charts
}

func wordPartNames(t *testing.T, docx []byte, pattern string) []string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatal(err)
	}
	matcher := regexp.MustCompile(pattern)
	names := make([]string, 0)
	for _, file := range reader.File {
		if matcher.MatchString(file.Name) {
			names = append(names, file.Name)
		}
	}
	sort.Strings(names)
	return names
}

func maskPhase1V2ControlValues(t *testing.T, part []byte) []byte {
	t.Helper()
	content := string(part)
	matches := wordContentControlPattern.FindAllStringIndex(content, -1)
	for index := len(matches) - 1; index >= 0; index-- {
		bounds := matches[index]
		masked, err := replaceWordContentControlText(content[bounds[0]:bounds[1]], "VALUE")
		if err != nil {
			t.Fatal(err)
		}
		content = content[:bounds[0]] + masked + content[bounds[1]:]
	}
	return []byte(unwrapPhase1V2ContentControls(content))
}

func phase1V2ChartByKey(t *testing.T, docx []byte, key string) []byte {
	t.Helper()
	parts := make(map[string][]byte)
	for _, name := range wordPartNames(t, docx, `.*`) {
		parts[name] = readWordPart(t, docx, name)
	}
	resolved, err := resolvePhase1V2BusinessChartParts(parts)
	if err != nil {
		t.Fatal(err)
	}
	return parts[resolved[key]]
}
