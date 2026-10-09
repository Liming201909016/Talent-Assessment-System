package handler

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/service"
)

func managementTestWordTemplate(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-only-v2.docx"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func managementTestWordDTO() service.ManagementTraitsTestReportData {
	d := service.ManagementTraitsTestReportData{Schema: "mng-test-report-data-v1", TestOnly: true, RunID: "run", PaperID: "paper", ExamID: "exam", ContentSourceSHA: "b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c", Participant: service.ManagementTraitsTestReportParticipant{Name: "DTO人员", Post: "冻结岗位"}, SubmittedAt: "2026-10-03 08:00", Overall: service.ManagementTraitsTestReportOverall{Score: "50.00", Level: "合格", Diagnosis: "总体现有文案", Advice: []string{"建议第一段", "建议第二段", "建议第三段"}, ChartScore: "50.000000000000"}}
	for _, dim := range service.ManagementTraitsDimensions() {
		d.Dimensions = append(d.Dimensions, service.ManagementTraitsTestReportDimension{Key: dim.Key, Name: dim.Name, Score: "50.00", Norm: dim.Norm.FloatString(2), Level: "合格", Diagnosis: "原始现有评价", Advice: "原始现有建议", ChartScore: "50.000000000000", ChartNorm: dim.Norm.FloatString(12)})
	}
	for _, key := range []string{"self", "interpersonal", "task", "development"} {
		d.Modules = append(d.Modules, service.ManagementTraitsTestReportModule{Key: key, Score: "50.00", ChartScore: "50.000000000000"})
	}
	for _, dim := range d.Dimensions[:3] {
		x := service.ManagementTraitsTestReportSelection{Key: dim.Key, Name: dim.Name, Text: "原始摘要"}
		d.Highest = append(d.Highest, x)
		d.Lowest = append(d.Lowest, x)
	}
	return d
}

func managementWordParts(t *testing.T, raw []byte) map[string][]byte {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	parts := map[string][]byte{}
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		parts[f.Name] = b
	}
	return parts
}

func TestManagementTraitsTestWordIndependentValueOnly(t *testing.T) {
	raw := managementTestWordTemplate(t)
	dto := managementTestWordDTO()
	dto.Overall.Score, dto.Overall.ChartScore = "28.85", "28.850000000000"
	for i, score := range []string{"25", "75", "100", "50"} {
		dto.Modules[i].Score = score + ".00"
		dto.Modules[i].ChartScore = score + ".000000000000"
	}
	out, err := renderManagementTraitsTestWord(raw, dto)
	if err != nil {
		t.Fatal(err)
	}
	before, after := managementWordParts(t, raw), managementWordParts(t, out)
	if len(before) != len(after) {
		t.Fatal("OPC parts added/deleted")
	}
	for name, b := range before {
		if name != "word/document.xml" && !strings.HasPrefix(name, "word/charts/chart") {
			if !bytes.Equal(b, after[name]) {
				t.Fatal("non-value part changed", name)
			}
		}
	}
	// MT-WORD-RUNTIME-01: actual semantic keys/OPC bindings, not chart types.
	expected := map[string][][]string{
		"word/charts/chart1.xml": {{"28.850000000000", "71.150000000000"}},
		"word/charts/chart2.xml": {{"25.000000000000", "75.000000000000"}},
		"word/charts/chart3.xml": {{"100.000000000000", "0.000000000000"}},
		"word/charts/chart4.xml": {{"50.000000000000", "50.000000000000"}},
		"word/charts/chart5.xml": {{"75.000000000000", "25.000000000000"}},
	}
	values, norms := make([]string, 0, 13), make([]string, 0, 13)
	for _, d := range dto.Dimensions {
		values = append(values, d.ChartScore)
		norms = append(norms, d.ChartNorm)
	}
	expected["word/charts/chart6.xml"] = [][]string{values, norms}
	for part, want := range expected {
		oldTree, err := mngWordTree(before[part])
		if err != nil {
			t.Fatal(err)
		}
		newTree, err := mngWordTree(after[part])
		if err != nil {
			t.Fatal(err)
		}
		oldSeries, newSeries := oldTree.all(mngChartNS, "ser"), newTree.all(mngChartNS, "ser")
		if len(oldSeries) != len(want) || len(newSeries) != len(want) {
			t.Fatal("series cardinality", part)
		}
		oldEdits, newEdits := make([]mngWordEdit, 0), make([]mngWordEdit, 0)
		for i, series := range newSeries {
			oldPoints := oldSeries[i].all(mngChartNS, "val")[0].all(mngChartNS, "pt")
			newPoints := series.all(mngChartNS, "val")[0].all(mngChartNS, "pt")
			if len(oldPoints) != len(want[i]) || len(newPoints) != len(want[i]) {
				t.Fatal("point cardinality", part)
			}
			for j, p := range newPoints {
				n := p.all(mngChartNS, "v")[0]
				if string(after[part][n.contentStart:n.contentEnd]) != want[i][j] {
					t.Fatal("wrong business chart value", part, i, j)
				}
				e, err := mngWordValueEdit(n, "")
				if err != nil {
					t.Fatal(err)
				}
				newEdits = append(newEdits, e)
				e, err = mngWordValueEdit(oldPoints[j].all(mngChartNS, "v")[0], "")
				if err != nil {
					t.Fatal(err)
				}
				oldEdits = append(oldEdits, e)
			}
		}
		oldXML, err := mngWordApply(before[part], oldEdits)
		if err != nil {
			t.Fatal(err)
		}
		newXML, err := mngWordApply(after[part], newEdits)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(oldXML, newXML) {
			t.Fatal("chart non-value XML changed", part)
		}
	}
	project := func(raw []byte) []byte {
		tree, err := mngWordTree(raw)
		if err != nil {
			t.Fatal(err)
		}
		edits := make([]mngWordEdit, 0)
		for _, n := range tree.all(mngWordNS, "t") {
			e, err := mngWordValueEdit(n, "")
			if err != nil {
				t.Fatal(err)
			}
			edits = append(edits, e)
		}
		b, err := mngWordApply(raw, edits)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	if !bytes.Equal(project(before["word/document.xml"]), project(after["word/document.xml"])) {
		t.Fatal("document non-text XML/style changed")
	}
	doc := string(after["word/document.xml"])
	for _, value := range []string{"测试报告", "DTO人员", "冻结岗位", "建议第一段", "建议第二段", "建议第三段", "原始现有评价", "原始现有建议"} {
		if !strings.Contains(doc, value) {
			t.Fatal("missing bound value", value)
		}
	}
	tree, err := mngWordTree(after["word/document.xml"])
	if err != nil {
		t.Fatal(err)
	}
	labels := map[string]string{"chart.overall.numeric-label": "28.85", "chart.module.self.numeric-label": "25.00", "chart.module.task.numeric-label": "100.00", "chart.module.development.numeric-label": "50.00", "chart.module.interpersonal.numeric-label": "75.00"}
	for _, a := range tree.all(mngDrawingNS, "anchor") {
		props := a.all(mngDrawingNS, "docPr")
		if len(props) == 1 {
			key := props[0].attr("title")
			if value, ok := labels[key]; ok {
				n := a.all(mngWordNS, "t")[1]
				if string(after["word/document.xml"][n.contentStart:n.contentEnd]) != value {
					t.Fatal("wrong numeric slot", key)
				}
				delete(labels, key)
			}
		}
	}
	if len(labels) != 0 {
		t.Fatal("missing numeric label bindings")
	}
	if strings.Contains(doc, "59.45") || strings.Contains(doc, "64.44") {
		t.Fatal("numeric sample label survived")
	}
	if !bytes.Contains(after["word/footer1.xml"], []byte("NUMPAGES")) && bytes.Contains(before["word/footer1.xml"], []byte("NUMPAGES")) {
		t.Fatal("source footer policy modified")
	}
	if output := os.Getenv("MNG_TEST_WORD_OUTPUT"); output != "" {
		if err := os.WriteFile(output, out, 0600); err != nil {
			t.Fatal(err)
		}
		all50, err := renderManagementTraitsTestWord(raw, managementTestWordDTO())
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(filepath.Dir(output), "management-traits-runtime-all50.docx"), all50, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, bad := range []service.ManagementTraitsTestReportData{{}, func() service.ManagementTraitsTestReportData { x := dto; x.TestOnly = false; return x }(), func() service.ManagementTraitsTestReportData { x := dto; x.Dimensions = x.Dimensions[:12]; return x }(), func() service.ManagementTraitsTestReportData {
		x := dto
		x.Overall.Advice = x.Overall.Advice[:2]
		return x
	}()} {
		if _, err := renderManagementTraitsTestWord(raw, bad); err == nil {
			t.Fatal("broken/formal DTO accepted")
		}
	}
	raw[len(raw)/2] ^= 1
	if _, err := renderManagementTraitsTestWord(raw, dto); err == nil {
		t.Fatal("broken/unapproved template accepted")
	}
}
