package handler

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/big"
	"path"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/talent-assessment/refactored/internal/service"
)

const managementTraitsTestTemplateSHA = "a986ba0f3c5985486cf5f2a781346beb4140d2f2e79952a25c3beb57a273eb46"
const managementTraitsRuntimeTestTemplateSHA = "05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c"
const mngWordZeroRingGray = "E7E6E6"
const mngWordNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
const mngChartNS = "http://schemas.openxmlformats.org/drawingml/2006/chart"
const mngDrawingNS = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"

// Offsets permit exact value-only splices. No XML reserialization, no style/run
// reconstruction, no 00401 renderer, no chart layout/color normalization.
type mngWordNode struct {
	name                                 xml.Name
	attrs                                []xml.Attr
	start, contentStart, contentEnd, end int
	children                             []*mngWordNode
}
type mngWordEdit struct {
	start, end int
	value      string
}

func mngWordTree(raw []byte) (*mngWordNode, error) {
	d := xml.NewDecoder(bytes.NewReader(raw))
	root := &mngWordNode{}
	stack := []*mngWordNode{root}
	for {
		start := int(d.InputOffset())
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			n := &mngWordNode{name: t.Name, attrs: t.Attr, start: start, contentStart: int(d.InputOffset())}
			p := stack[len(stack)-1]
			p.children = append(p.children, n)
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) < 2 {
				return nil, errors.New("unbalanced XML")
			}
			n := stack[len(stack)-1]
			n.contentEnd, n.end = start, int(d.InputOffset())
			stack = stack[:len(stack)-1]
		case xml.Directive:
			return nil, errors.New("XML directives forbidden")
		}
	}
	if len(stack) != 1 {
		return nil, errors.New("incomplete XML")
	}
	return root, nil
}
func (n *mngWordNode) attr(local string) string {
	for _, a := range n.attrs {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}
func (n *mngWordNode) all(space, local string) []*mngWordNode {
	out := make([]*mngWordNode, 0)
	var walk func(*mngWordNode)
	walk = func(p *mngWordNode) {
		if p.name.Space == space && p.name.Local == local {
			out = append(out, p)
		}
		for _, c := range p.children {
			walk(c)
		}
	}
	walk(n)
	return out
}
func mngWordValueEdit(n *mngWordNode, value string) (mngWordEdit, error) {
	if n == nil || len(n.children) > 0 || n.contentEnd < n.contentStart || !utf8.ValidString(value) {
		return mngWordEdit{}, errors.New("invalid text slot")
	}
	var b bytes.Buffer
	if err := xml.EscapeText(&b, []byte(value)); err != nil {
		return mngWordEdit{}, err
	}
	return mngWordEdit{n.contentStart, n.contentEnd, b.String()}, nil
}
func mngWordApply(raw []byte, edits []mngWordEdit) ([]byte, error) {
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	limit := len(raw)
	out := append([]byte{}, raw...)
	for _, e := range edits {
		if e.start < 0 || e.end < e.start || e.end > limit {
			return nil, errors.New("overlapping value slots")
		}
		out = append(append(append([]byte{}, out[:e.start]...), []byte(e.value)...), out[e.end:]...)
		limit = e.start
	}
	return out, nil
}

func renderManagementTraitsTestWord(template []byte, dto service.ManagementTraitsTestReportData) ([]byte, error) {
	return renderManagementTraitsTestWordWithSHA(template, managementTraitsRuntimeTestTemplateSHA, dto)
}

func renderManagementTraitsTestWordWithSHA(template []byte, expectedSHA string, dto service.ManagementTraitsTestReportData) ([]byte, error) {
	stage := "input"
	fail := func() ([]byte, error) {
		return nil, fmt.Errorf("管理特质测试报告模板或数据校验失败 (%s)", stage)
	}
	hash := sha256.Sum256(template)
	if len(expectedSHA) != sha256.Size*2 || strings.ToLower(expectedSHA) != expectedSHA {
		return fail()
	}
	if decoded, err := hex.DecodeString(expectedSHA); err != nil || len(decoded) != sha256.Size {
		return fail()
	}
	if len(template) > 20<<20 || hex.EncodeToString(hash[:]) != expectedSHA || !dto.TestOnly || dto.Schema != "mng-test-report-data-v1" || dto.ContentSourceSHA != "b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c" || dto.RunID == "" || dto.PaperID == "" || dto.ExamID == "" || len(dto.Dimensions) != 13 || len(dto.Modules) != 4 || len(dto.Highest) != 3 || len(dto.Lowest) != 3 || len(dto.Overall.Advice) != 3 {
		return fail()
	}
	fields := map[string]string{"participant.name": dto.Participant.Name, "participant.gender": dto.Participant.Gender, "participant.telephone": dto.Participant.Telephone, "participant.affiliation": dto.Participant.Affiliation, "participant.post": dto.Participant.Post, "result.submittedAt": dto.SubmittedAt, "overall.score": dto.Overall.Score, "overall.level": dto.Overall.Level, "overall.diagnosis": dto.Overall.Diagnosis}
	fields["report.testTitle"], fields["report.testLabel"] = ManagementTraitsTestReportTitle, ManagementTraitsTestReportLabel
	for i, s := range dto.Overall.Advice {
		if strings.TrimSpace(s) == "" {
			return fail()
		}
		fields[fmt.Sprintf("overall.advice.%d", i+1)] = s
	}
	charts := make(map[string][][]string, 6)
	zeroRings := make(map[string]bool, 5)
	complement := func(key, score, display string) (string, bool) {
		r, ok := new(big.Rat).SetString(score)
		if !ok || r.Sign() < 0 || r.Cmp(big.NewRat(100, 1)) > 0 || r.FloatString(2) != display {
			return "", false
		}
		if key != "" && r.Sign() == 0 {
			zeroRings[key] = true
		}
		return new(big.Rat).Sub(big.NewRat(100, 1), r).FloatString(12), true
	}
	c, ok := complement("chart.overall", dto.Overall.ChartScore, dto.Overall.Score)
	if !ok {
		return fail()
	}
	charts["chart.overall"] = [][]string{{dto.Overall.ChartScore, c}}
	defs := service.ManagementTraitsDimensions()
	values, norms := make([]string, 0, 13), make([]string, 0, 13)
	for i, d := range dto.Dimensions {
		if d.Key != defs[i].Key || d.Name != defs[i].Name || strings.TrimSpace(d.Diagnosis) == "" || strings.TrimSpace(d.Advice) == "" {
			return fail()
		}
		if _, ok := complement("", d.ChartScore, d.Score); !ok {
			return fail()
		}
		if d.ChartNorm != defs[i].Norm.FloatString(12) || d.Norm != defs[i].Norm.FloatString(2) {
			return fail()
		}
		prefix := "dimension." + d.Key + "."
		fields[prefix+"score"], fields[prefix+"norm"], fields[prefix+"level"], fields[prefix+"diagnosis"], fields[prefix+"advice"] = d.Score, d.Norm, d.Level, d.Diagnosis, d.Advice
		values = append(values, d.ChartScore)
		norms = append(norms, d.ChartNorm)
	}
	charts["chart.dimension.comparison"] = [][]string{values, norms}
	seenModules := map[string]bool{}
	for _, m := range dto.Modules {
		if seenModules[m.Key] || !(m.Key == "self" || m.Key == "interpersonal" || m.Key == "task" || m.Key == "development") {
			return fail()
		}
		seenModules[m.Key] = true
		c, ok := complement("chart.module."+m.Key, m.ChartScore, m.Score)
		if !ok {
			return fail()
		}
		fields["module."+m.Key+".score"] = m.Score
		charts["chart.module."+m.Key] = [][]string{{m.ChartScore, c}}
	}
	for _, g := range []struct {
		name  string
		items []service.ManagementTraitsTestReportSelection
	}{{"high", dto.Highest}, {"low", dto.Lowest}} {
		for i, s := range g.items {
			if s.Name == "" || strings.TrimSpace(s.Text) == "" {
				return fail()
			}
			fields[fmt.Sprintf("overview.%s.%d.name", g.name, i+1)] = s.Name
			fields[fmt.Sprintf("overview.%s.%d.text", g.name, i+1)] = s.Text
		}
	}
	stage = "package"
	z, err := zip.NewReader(bytes.NewReader(template), int64(len(template)))
	if err != nil {
		return fail()
	}
	parts := make(map[string][]byte, len(z.File))
	total := int64(0)
	for _, f := range z.File {
		if _, exists := parts[f.Name]; exists || path.Clean(f.Name) != strings.TrimSuffix(f.Name, "/") || strings.HasPrefix(f.Name, "/") || strings.Contains(f.Name, "\\") {
			return fail()
		}
		r, e := f.Open()
		if e != nil {
			return fail()
		}
		b, e := io.ReadAll(io.LimitReader(r, (64<<20)-total+1))
		closeErr := r.Close()
		total += int64(len(b))
		if e != nil || closeErr != nil || total > 64<<20 {
			return fail()
		}
		parts[f.Name] = b
	}
	doc, err := mngWordTree(parts["word/document.xml"])
	if err != nil {
		return fail()
	}
	stage = "controls"
	edits := make([]mngWordEdit, 0, 100)
	seen := map[string]int{}
	requiredFields := map[string]bool{}
	for _, field := range managementTraitsTemplateSemanticFields() {
		if field.Required {
			requiredFields[field.Key] = true
		}
	}
	for _, sdt := range doc.all(mngWordNS, "sdt") {
		tags := sdt.all(mngWordNS, "tag")
		text := sdt.all(mngWordNS, "t")
		if len(tags) != 1 || len(text) == 0 {
			return fail()
		}
		key := tags[0].attr("val")
		value, ok := fields[key]
		if !ok {
			return fail()
		}
		seen[key]++
		for i, n := range text {
			v := ""
			if i == 0 {
				v = value
			}
			e, err := mngWordValueEdit(n, v)
			if err != nil {
				return fail()
			}
			edits = append(edits, e)
		}
	}
	for key := range requiredFields {
		if seen[key] == 0 {
			return fail()
		}
	}
	stage = "relationships"
	rels, err := mngWordTree(parts["word/_rels/document.xml.rels"])
	if err != nil {
		return fail()
	}
	targets := map[string]string{}
	for _, r := range rels.all("http://schemas.openxmlformats.org/package/2006/relationships", "Relationship") {
		if r.attr("TargetMode") == "External" {
			return fail()
		}
		targets[r.attr("Id")] = path.Clean(path.Join("word", r.attr("Target")))
	}
	resolved := map[string]bool{}
	numeric := map[string]bool{}
	stage = "charts"
	for _, anchor := range doc.all(mngDrawingNS, "anchor") {
		props := anchor.all(mngDrawingNS, "docPr")
		if len(props) != 1 {
			continue
		}
		key := props[0].attr("title")
		if strings.HasSuffix(key, ".numeric-label") {
			base := strings.TrimSuffix(key, ".numeric-label")
			texts := anchor.all(mngWordNS, "t")
			field := "overall.score"
			if strings.HasPrefix(base, "chart.module.") {
				field = strings.TrimPrefix(base, "chart.") + ".score"
			}
			if numeric[base] || (len(texts) != 3 && len(texts) != 4) || charts[base] == nil {
				return fail()
			}
			if len(texts) == 4 && strings.TrimSpace(string(parts["word/document.xml"][texts[3].contentStart:texts[3].contentEnd])) != "" {
				return fail()
			}
			e, err := mngWordValueEdit(texts[1], fields[field])
			if err != nil {
				return fail()
			}
			edits = append(edits, e)
			numeric[base] = true
			continue
		}
		chart := anchor.all(mngChartNS, "chart")
		if len(chart) == 0 {
			continue
		}
		if len(chart) != 1 || resolved[key] || charts[key] == nil {
			return fail()
		}
		resolved[key] = true
		part := targets[chart[0].attr("id")]
		if !strings.HasPrefix(part, "word/charts/") {
			return fail()
		}
		tree, err := mngWordTree(parts[part])
		if err != nil {
			return fail()
		}
		series := tree.all(mngChartNS, "ser")
		if len(series) != len(charts[key]) {
			return fail()
		}
		chartEdits := make([]mngWordEdit, 0, 26)
		for i, s := range series {
			val := s.all(mngChartNS, "val")
			if len(val) != 1 {
				return fail()
			}
			literals := val[0].all(mngChartNS, "numLit")
			if len(literals) != 1 {
				return fail()
			}
			points := literals[0].all(mngChartNS, "pt")
			if len(points) != len(charts[key][i]) {
				return fail()
			}
			for j, p := range points {
				v := p.all(mngChartNS, "v")
				if p.attr("idx") != strconv.Itoa(j) || len(v) != 1 {
					return fail()
				}
				e, err := mngWordValueEdit(v[0], charts[key][i][j])
				if err != nil {
					return fail()
				}
				chartEdits = append(chartEdits, e)
			}
		}
		// MT-ZERO-RING-01: sole finite rendering style exception. Exact zero
		// exposes the remainder in template light gray; line noFill, values,
		// labels and all nonzero/comparison styles remain byte-for-byte intact.
		if zeroRings[key] {
			fills := 0
			for _, point := range series[0].all(mngChartNS, "dPt") {
				idx := point.all(mngChartNS, "idx")
				if len(idx) != 1 || idx[0].attr("val") != "1" {
					continue
				}
				for _, sp := range point.children {
					if sp.name.Space != mngChartNS || sp.name.Local != "spPr" {
						continue
					}
					for _, fill := range sp.children {
						if fill.name.Space == "http://schemas.openxmlformats.org/drawingml/2006/main" && fill.name.Local == "noFill" {
							chartEdits = append(chartEdits, mngWordEdit{fill.start, fill.end, `<a:solidFill><a:srgbClr val="` + mngWordZeroRingGray + `"/></a:solidFill>`})
							fills++
						}
					}
				}
			}
			if fills != 1 {
				return fail()
			}
		}
		parts[part], err = mngWordApply(parts[part], chartEdits)
		if err != nil {
			return fail()
		}
	}
	stage = "chart-cardinality"
	if len(resolved) != 6 || len(numeric) != 5 {
		return fail()
	}
	parts["word/document.xml"], err = mngWordApply(parts["word/document.xml"], edits)
	if err != nil {
		return fail()
	}
	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	for _, f := range z.File {
		entry, e := writer.CreateHeader(&zip.FileHeader{Name: f.Name, Method: zip.Deflate})
		if e != nil {
			writer.Close()
			return fail()
		}
		if _, e = entry.Write(parts[f.Name]); e != nil {
			writer.Close()
			return fail()
		}
	}
	if writer.Close() != nil {
		return fail()
	}
	return out.Bytes(), nil
}
