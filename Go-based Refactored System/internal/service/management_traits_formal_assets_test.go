package service

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func formalPreviewFixture(t *testing.T, withTemplate bool) (string, string) {
	t.Helper()
	root := t.TempDir()
	key := "preview"
	dir := filepath.Join(root, key)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	workbook, err := os.ReadFile(filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-content-v1.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "content.xlsx"), workbook, 0600); err != nil {
		t.Fatal(err)
	}
	if withTemplate {
		if err = os.WriteFile(filepath.Join(dir, "template.docx"), formalSyntheticTemplate(t), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root, key
}

func TestManagementTraitsFormalAssetPreview(t *testing.T) {
	root, key := formalPreviewFixture(t, true)
	s := &ManagementTraitsFormalRegistry{environment: "local", assetRoot: root}
	got, err := s.PreviewAssets(t.Context(), "00501", key)
	if err != nil {
		t.Fatal(err)
	}
	if !got.ReadyForRegistration || !got.ReadyForApproval || got.ContentRuleCount != 205 || got.TagCount != 88 || got.ChartCount != 6 || got.NumericLabelCount != 5 {
		t.Fatalf("unexpected preview: %#v", got)
	}
	if len(got.BlockedReasons) != 0 || got.BlockedReasons == nil || got.Tags == nil || got.Charts == nil || got.NumericLabels == nil {
		t.Fatal("preview collections must be non-nil")
	}
	for _, value := range []string{got.WorkbookSHA, got.NormalizedContentSHA, got.TemplateSHA, got.BindingSHA} {
		if len(value) != 64 {
			t.Fatal("preview digest missing")
		}
	}
}

func TestManagementTraitsFormalAssetPreviewDraftAndScope(t *testing.T) {
	root, key := formalPreviewFixture(t, false)
	s := &ManagementTraitsFormalRegistry{environment: "local", assetRoot: root}
	got, err := s.PreviewAssets(t.Context(), "00502", key)
	if err != nil || !got.ReadyForRegistration || got.ReadyForApproval || len(got.BlockedReasons) != 1 || got.BlockedReasons[0] != "template_missing" {
		t.Fatalf("unexpected draft preview: %#v %v", got, err)
	}
	if got.Tags == nil || got.Charts == nil || got.NumericLabels == nil {
		t.Fatal("draft collections must be non-nil")
	}
	for _, repo := range []string{"00201", "00202", "00401", "00301", "", "00503"} {
		if _, err = s.PreviewAssets(t.Context(), repo, key); err == nil {
			t.Fatalf("repo %q accepted", repo)
		}
	}
}

func TestManagementTraitsFormalAssetPreviewReasonsAndBudgets(t *testing.T) {
	root, key := formalPreviewFixture(t, false)
	s := &ManagementTraitsFormalRegistry{environment: "local", assetRoot: root}
	templatePath := filepath.Join(root, key, "template.docx")
	if err := os.WriteFile(templatePath, []byte("bad zip"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := s.PreviewAssets(t.Context(), "00501", key)
	if err != nil || len(got.BlockedReasons) != 1 || got.BlockedReasons[0] != "template_zip_invalid" {
		t.Fatalf("bad zip reason: %#v %v", got, err)
	}

	var many bytes.Buffer
	zw := zip.NewWriter(&many)
	for i := 0; i < 257; i++ {
		w, createErr := zw.Create("part-" + strconv.Itoa(i) + ".xml")
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, createErr = w.Write([]byte("<x/>")); createErr != nil {
			t.Fatal(createErr)
		}
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(templatePath, many.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = s.PreviewAssets(t.Context(), "00501", key)
	if err != nil || len(got.BlockedReasons) != 1 || got.BlockedReasons[0] != "template_parts_limit" {
		t.Fatalf("part limit reason: %#v %v", got, err)
	}

	tooLarge := filepath.Join(root, key, "content.xlsx")
	f, createErr := os.OpenFile(tooLarge, os.O_WRONLY|os.O_TRUNC, 0600)
	if createErr != nil {
		t.Fatal(createErr)
	}
	if createErr = f.Truncate((20 << 20) + 1); createErr != nil {
		f.Close()
		t.Fatal(createErr)
	}
	if createErr = f.Close(); createErr != nil {
		t.Fatal(createErr)
	}
	got, err = s.PreviewAssets(t.Context(), "00501", key)
	if err != nil || len(got.BlockedReasons) != 1 || got.BlockedReasons[0] != "content_file_too_large" {
		t.Fatalf("content limit reason: %#v %v", got, err)
	}
}

func TestFormalReadAssetRejectsIdentityChangeAfterOpen(t *testing.T) {
	root, key := formalPreviewFixture(t, false)
	_, err := formalReadAssetChecked(root, key, "content.xlsx", func(file string) error {
		return os.Rename(file, file+".moved")
	})
	if err == nil {
		t.Fatal("asset identity change accepted")
	}
}

func TestManagementTraitsFormalTemplateRealAssets(t *testing.T) {
	candidate := filepath.Join("..", "..", "..", "docs", "260929管理特质测评-优化", "260928管理潜质测评报告模板修改稿V2.8.docx")
	b, err := os.ReadFile(candidate)
	if err != nil {
		t.Fatal(err)
	}
	before := formalSHA(b)
	_, err = validateManagementTraitsFormalTemplate(b)
	if err == nil {
		t.Fatal("customer template with external relationships accepted")
	}
	current, err := os.ReadFile(candidate)
	if err != nil || formalSHA(current) != before {
		t.Fatal("candidate modified")
	}
	testRaw, err := os.ReadFile(filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-only-v2.docx"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = validateManagementTraitsFormalTemplate(testRaw); err == nil {
		t.Fatal("TEST template promoted")
	}
	if _, err = validateManagementTraitsFormalTemplate([]byte("not a DOCX")); err == nil {
		t.Fatal("bad ZIP accepted")
	}
}

func TestManagementTraitsFormalControlledDraftAssets(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "draft")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-content-v1.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "content.xlsx"), b, 0600); err != nil {
		t.Fatal(err)
	}
	a, err := loadManagementTraitsFormalAssets(root, "draft")
	if err != nil || a.templateSHA != "" || a.workbookSHA != managementTraitsTestWorkbookSHA || len(a.contentSHA) != 64 || a.contentSHA == a.workbookSHA {
		t.Fatal("draft or separate hashes rejected", err)
	}
	for _, key := range []string{"../draft", "C:/private", "", "draft/child", strings.Repeat("a", 49)} {
		if _, err = loadManagementTraitsFormalAssets(root, key); err == nil {
			t.Fatal("unsafe asset identifier")
		}
	}
}

func TestManagementTraitsFormalSourceIdentityExcludesProfilePerson(t *testing.T) {
	f, _, _, p, _ := managementRuntimeLoadFixture(t, false)
	original := formalSourceSnapshot(f.Bundle, p)
	p.ExamID = "another-exam"
	p.FieldContract = "another-person-field-set"
	p.CreatedAt = p.CreatedAt.AddDate(0, 0, 1)
	if formalSourceSnapshot(f.Bundle, p) != original {
		t.Fatal("approval bound to an individual profile")
	}
	p.MappingSHA = strings.Repeat("f", 64)
	if formalSourceSnapshot(f.Bundle, p) == original {
		t.Fatal("mapping change not bound")
	}
}
