package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/service"
)

// MT-ZERO-RING-01: docs/regression-tests.md. Exact zero must have a visible
// remainder, without changing values, labels, outlines, or any other OPC style.
func TestBugManagementTraitsZeroRing(t *testing.T) {
	raw := managementTestWordTemplate(t)
	hash := sha256.Sum256(raw)
	if hex.EncodeToString(hash[:]) != "05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c" {
		t.Fatal("original TEST template changed")
	}
	before := managementWordParts(t, raw)
	cases := []struct {
		name   string
		scores []string // overall, self, interpersonal, task, development
	}{
		{"all-zero", []string{"0", "0", "0", "0", "0"}},
		{"near-zero", []string{".004", ".004", ".004", ".004", ".004"}},
		{"score-25", []string{"25", "25", "25", "25", "25"}},
		{"score-50", []string{"50", "50", "50", "50", "50"}},
		{"score-75", []string{"75", "75", "75", "75", "75"}},
		{"score-100", []string{"100", "100", "100", "100", "100"}},
		{"mixed", []string{"25", "0", ".004", "100", "0"}},
		{"signed-zero", []string{"-0.000", "0.00", "+0", "-0", "0.000000000000"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dto := managementTestWordDTO()
			displays := make([]string, len(tc.scores))
			for i, score := range tc.scores {
				r, ok := new(big.Rat).SetString(score)
				if !ok {
					t.Fatal("invalid fixture score")
				}
				displays[i] = r.FloatString(2)
			}
			dto.Overall.ChartScore, dto.Overall.Score = tc.scores[0], displays[0]
			for i := range dto.Modules {
				dto.Modules[i].ChartScore, dto.Modules[i].Score = tc.scores[i+1], displays[i+1]
			}
			for i := range dto.Dimensions {
				j := 0
				if tc.name == "mixed" {
					j = i % len(tc.scores)
				}
				dto.Dimensions[i].ChartScore, dto.Dimensions[i].Score = tc.scores[j], displays[j]
			}
			out, err := renderManagementTraitsTestWord(raw, dto)
			if err != nil {
				t.Fatal(err)
			}
			after := managementWordParts(t, out)
			if len(before) != len(after) {
				t.Fatal("OPC part cardinality changed")
			}
			// This map independently pins all five physical rings to semantic keys.
			rings := map[string]int{"chart.overall": 0, "chart.module.self": 1, "chart.module.interpersonal": 2, "chart.module.task": 3, "chart.module.development": 4}
			physical := map[string]string{"chart.overall": "word/charts/chart1.xml", "chart.module.self": "word/charts/chart2.xml", "chart.module.interpersonal": "word/charts/chart5.xml", "chart.module.task": "word/charts/chart3.xml", "chart.module.development": "word/charts/chart4.xml"}
			doc, err := mngWordTree(after["word/document.xml"])
			if err != nil {
				t.Fatal(err)
			}
			rels, err := mngWordTree(before["word/_rels/document.xml.rels"])
			if err != nil {
				t.Fatal(err)
			}
			targets := map[string]string{}
			for _, n := range rels.children[0].children {
				targets[n.attr("Id")] = "word/" + n.attr("Target")
			}
			labels, bindings := 0, 0
			for _, anchor := range doc.all(mngDrawingNS, "anchor") {
				props := anchor.all(mngDrawingNS, "docPr")
				if len(props) != 1 {
					continue
				}
				key := props[0].attr("title")
				if i, ok := rings[strings.TrimSuffix(key, ".numeric-label")]; ok && strings.HasSuffix(key, ".numeric-label") {
					n := anchor.all(mngWordNS, "t")[1]
					if string(after["word/document.xml"][n.contentStart:n.contentEnd]) != displays[i] {
						t.Fatal("numeric label changed", key)
					}
					labels++
				}
				if part, ok := physical[key]; ok {
					chart := anchor.all(mngChartNS, "chart")
					if len(chart) != 1 || targets[chart[0].attr("id")] != part {
						t.Fatal("semantic/physical ring binding changed", key)
					}
					bindings++
				}
			}
			if labels != 5 || bindings != 5 {
				t.Fatal("missing five semantic labels/rings", labels, bindings)
			}
			expected := map[string][]byte{}
			for key, i := range rings {
				part := physical[key]
				r, _ := new(big.Rat).SetString(tc.scores[i])
				tree, err := mngWordTree(after[part])
				if err != nil {
					t.Fatal(err)
				}
				series := tree.all(mngChartNS, "ser")
				if len(series) != 1 {
					t.Fatal("ring series cardinality", key)
				}
				points := series[0].all(mngChartNS, "val")[0].all(mngChartNS, "pt")
				want := []string{tc.scores[i], new(big.Rat).Sub(big.NewRat(100, 1), r).FloatString(12)}
				if len(points) != 2 {
					t.Fatal("ring point cardinality", key)
				}
				for j, p := range points {
					n := p.all(mngChartNS, "v")[0]
					if string(after[part][n.contentStart:n.contentEnd]) != want[j] {
						t.Fatal("actual score/complement changed", key, j)
					}
				}
				expected[part] = before[part]
				if r.Sign() == 0 {
					oldTree, err := mngWordTree(before[part])
					if err != nil {
						t.Fatal(err)
					}
					// Independent expected byte splice: only the direct fill, NOT ln/noFill.
					var fill *mngWordNode
					for _, pt := range oldTree.all(mngChartNS, "dPt") {
						idx := pt.all(mngChartNS, "idx")
						if len(idx) != 1 || idx[0].attr("val") != "1" {
							continue
						}
						for _, sp := range pt.children {
							if sp.name.Space == mngChartNS && sp.name.Local == "spPr" {
								for _, n := range sp.children {
									if n.name.Space == "http://schemas.openxmlformats.org/drawingml/2006/main" && n.name.Local == "noFill" {
										if fill != nil {
											t.Fatal("duplicate direct remainder fill")
										}
										fill = n
									}
								}
							}
						}
					}
					if fill == nil {
						t.Fatal("missing original direct remainder noFill")
					}
					b := before[part]
					expected[part] = append(append(append([]byte{}, b[:fill.start]...), []byte(`<a:solidFill><a:srgbClr val="E7E6E6"/></a:solidFill>`)...), b[fill.end:]...)
				}
			}
			for part, original := range before {
				actual, exists := after[part]
				if !exists {
					t.Fatal("OPC part removed", part)
				}
				if part == "word/document.xml" {
					if !bytes.Equal(mngZeroRingScrub(t, original, mngWordNS, "t"), mngZeroRingScrub(t, actual, mngWordNS, "t")) {
						t.Fatal("document fonts/coordinates/non-text XML changed")
					}
				} else if strings.HasPrefix(part, "word/charts/chart") && strings.HasSuffix(part, ".xml") {
					want := original
					if b, ok := expected[part]; ok {
						want = b
					}
					if !bytes.Equal(mngZeroRingScrub(t, want, mngChartNS, "v"), mngZeroRingScrub(t, actual, mngChartNS, "v")) {
						t.Errorf("MT-ZERO-RING-01: transparent zero-ring or unauthorized style change: %s", part)
					}
				} else if !bytes.Equal(original, actual) {
					t.Error("unrelated OPC bytes changed", part)
				}
			}
			comparison, err := mngWordTree(after["word/charts/chart6.xml"])
			if err != nil {
				t.Fatal(err)
			}
			for i, series := range comparison.all(mngChartNS, "ser") {
				points := series.all(mngChartNS, "val")[0].all(mngChartNS, "pt")
				if len(points) != 13 {
					t.Fatal("comparison dimension cardinality")
				}
				for j, p := range points {
					want := dto.Dimensions[j].ChartScore
					if i == 1 {
						want = dto.Dimensions[j].ChartNorm
					}
					n := p.all(mngChartNS, "v")[0]
					if string(after["word/charts/chart6.xml"][n.contentStart:n.contentEnd]) != want {
						t.Fatal("comparison score/norm changed", i, j)
					}
				}
			}
			if dir := os.Getenv("MNG_ZERO_RING_OUTPUT_DIR"); dir != "" && !t.Failed() {
				if !filepath.IsAbs(dir) {
					t.Fatal("MNG_ZERO_RING_OUTPUT_DIR must be absolute")
				}
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				f, err := os.OpenFile(filepath.Join(dir, tc.name+".docx"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, writeErr := f.Write(out)
				closeErr := f.Close()
				if writeErr != nil || closeErr != nil {
					t.Fatal(writeErr, closeErr)
				}
				t.Log("rendered DOCX:", f.Name())
			}
		})
	}
	// Optional local evidence replay; the default regression is self-contained.
	if source := os.Getenv("MNG_ZERO_RING_FROZEN_SQL_FILE"); source != "" {
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		var rows []struct {
			Report struct {
				DataRaw string `json:"dataRaw"`
				DataSHA string `json:"dataSha"`
			} `json:"report"`
		}
		if err := json.Unmarshal(data, &rows); err != nil || len(rows) != 4 {
			t.Fatal("invalid four-report local evidence", err)
		}
		dir := os.Getenv("MNG_ZERO_RING_OUTPUT_DIR")
		if !filepath.IsAbs(dir) {
			t.Fatal("frozen replay requires absolute output directory")
		}
		for i, row := range rows {
			t.Run("frozen-"+[]string{"00201-candidate", "00201-tester", "00202-candidate", "00202-tester"}[i], func(t *testing.T) {
				hash := sha256.Sum256([]byte(row.Report.DataRaw))
				if hex.EncodeToString(hash[:]) != row.Report.DataSHA {
					t.Fatal("frozen DTO evidence checksum mismatch")
				}
				var dto service.ManagementTraitsTestReportData
				if err := json.Unmarshal([]byte(row.Report.DataRaw), &dto); err != nil {
					t.Fatal(err)
				}
				out, err := renderManagementTraitsTestWord(raw, dto)
				if err != nil {
					t.Fatal(err)
				}
				file := filepath.Join(dir, "frozen-"+[]string{"00201-candidate", "00201-tester", "00202-candidate", "00202-tester"}[i]+".docx")
				f, err := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, writeErr := f.Write(out)
				closeErr := f.Close()
				if writeErr != nil || closeErr != nil {
					t.Fatal(writeErr, closeErr)
				}
				t.Log("unchanged frozen DTO rendered locally:", file)
			})
		}
	}
}

func mngZeroRingScrub(t *testing.T, raw []byte, space, local string) []byte {
	t.Helper()
	tree, err := mngWordTree(raw)
	if err != nil {
		t.Fatal(err)
	}
	edits := make([]mngWordEdit, 0)
	for _, n := range tree.all(space, local) {
		e, err := mngWordValueEdit(n, "")
		if err != nil {
			t.Fatal(err)
		}
		edits = append(edits, e)
	}
	out, err := mngWordApply(raw, edits)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
