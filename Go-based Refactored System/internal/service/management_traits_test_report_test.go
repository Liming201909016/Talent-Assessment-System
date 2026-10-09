package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func managementTestWorkbook(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "260929管理特质测评-优化", "260928测评内容+数据图.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestManagementTraitsTestContentExactSource(t *testing.T) {
	raw := managementTestWorkbook(t)
	content, err := LoadManagementTraitsTestContent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if content.RuleCount() != 205 || content.SourceSHA() != "b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c" {
		t.Fatal("source/count contract lost")
	}
	raw[len(raw)/2] ^= 1
	if _, err := LoadManagementTraitsTestContent(raw); err == nil {
		t.Fatal("broken source accepted")
	}
}

func TestManagementTraitsTestReportValidatedSnapshot(t *testing.T) {
	content, err := LoadManagementTraitsTestContent(managementTestWorkbook(t))
	if err != nil {
		t.Fatal(err)
	}
	f, records, _, _, _ := managementRuntimeLoadFixture(t, false)
	got, err := BuildManagementTraitsTestReport(f.Bundle, f.Paper, f.Questions, records.Run, records.Dimensions, records.Modules, records.Receipt, content, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if !got.TestOnly || got.Schema != "mng-test-report-data-v1" || got.RunID != records.Run.ID || got.Overall.Score != "50.00" || len(got.Dimensions) != 13 || len(got.Modules) != 4 || len(got.Highest) != 3 || len(got.Lowest) != 3 || len(got.Overall.Advice) != 3 || got.Participant.Name != "test-person" || got.SubmittedAt != records.Run.SubmittedAt.Format("2006-01-02 15:04") {
		t.Fatal("report facts not frozen", got)
	}
	for _, d := range got.Dimensions {
		if d.Level != "合格" || strings.TrimSpace(d.Diagnosis) == "" || strings.TrimSpace(d.Advice) == "" {
			t.Fatal("exact grade rule missing", d)
		}
	}
	for _, mutate := range []func(){
		func() { records.Run.InputSHA = strings.Repeat("0", 64) },
		func() { records.Dimensions[0].ScoreSum++ },
		func() { f.Paper.ParticipantSnapshot = `{"name":"wrong"}` },
	} {
		f, records, _, _, _ = managementRuntimeLoadFixture(t, false)
		mutate()
		if _, err := BuildManagementTraitsTestReport(f.Bundle, f.Paper, f.Questions, records.Run, records.Dimensions, records.Modules, records.Receipt, content, 1<<20); err == nil {
			t.Fatal("broken run/snapshot accepted")
		}
	}
	f, records, _, _, _ = managementRuntimeLoadFixture(t, true)
	if _, err := BuildManagementTraitsTestReport(f.Bundle, f.Paper, f.Questions, records.Run, records.Dimensions, records.Modules, records.Receipt, content, 1<<20); err == nil {
		t.Fatal("incomplete rendered as successful report")
	}
}
