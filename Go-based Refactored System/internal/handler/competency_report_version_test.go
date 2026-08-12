package handler

import (
	"strings"
	"testing"
)

func TestCompetencyReportVersions_RequiresFrozenValues(t *testing.T) {
	content, template, err := competencyReportVersions(map[string]any{
		"contentVersion":        "content-frozen",
		"reportTemplateVersion": "template-frozen",
	})
	if err != nil {
		t.Fatal(err)
	}
	if content != "content-frozen" || template != "template-frozen" {
		t.Fatalf("content=%q template=%q", content, template)
	}
	if _, _, err := competencyReportVersions(map[string]any{"contentVersion": "content-only"}); err == nil {
		t.Fatal("missing report template version must fail")
	}
}

func TestCompetencyReportHandler_UsesFrozenContentAndTemplateVersions(t *testing.T) {
	source := readSourceFile(t, "competency_report.go")
	for _, required := range []string{
		"competencyReportVersions(data)",
		`paper_id = ? AND content_version = ? AND template_version = ?`,
		"report.ContentVersion = contentVersion",
		"report.TemplateVersion = templateVersion",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("competency report version lookup missing %q", required)
		}
	}
	if strings.Contains(source, "service.CompetencyTemporaryContentVersion") {
		t.Error("report handler must not select instances using a process-wide content constant")
	}
}
