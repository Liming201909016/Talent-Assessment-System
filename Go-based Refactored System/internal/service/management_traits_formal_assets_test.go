package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagementTraitsFormalTemplateRealAssets(t *testing.T) {
	candidate := filepath.Join("..", "..", "..", "docs", "generated", "management-traits-word-candidate-20261001", "management-traits-002-lo-compatible-template.docx")
	b, err := os.ReadFile(candidate)
	if err != nil {
		t.Fatal(err)
	}
	before := formalSHA(b)
	_, err = validateManagementTraitsFormalTemplate(b)
	if err == nil {
		t.Fatal("candidate with external LINK field accepted")
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
