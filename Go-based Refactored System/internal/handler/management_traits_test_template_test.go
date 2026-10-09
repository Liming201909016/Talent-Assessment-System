package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestManagementTraitsTestTemplateDeterministicAndSourceProtected(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "generated", "management-traits-word-candidate-20261001", "management-traits-002-lo-compatible-template.docx")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	a, err := BuildManagementTraitsTestTemplate(source)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BuildManagementTraitsTestTemplate(source)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(a)
	if !bytes.Equal(a, b) || hex.EncodeToString(h[:]) != managementTraitsRuntimeTestTemplateSHA || !bytes.Equal(a, managementTestWordTemplate(t)) {
		t.Fatal("non-deterministic or unpinned test template")
	}
	before, after := managementWordParts(t, source), managementWordParts(t, a)
	if len(before) != len(after) {
		t.Fatal("OPC parts added or deleted")
	}
	for name, data := range before {
		if name != "word/document.xml" && !bytes.Equal(data, after[name]) {
			t.Fatal("source part changed", name)
		}
	}
	tree, err := mngWordTree(after["word/document.xml"])
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.all(mngWordNS, "sdt")) != 90 {
		t.Fatal("test label contract")
	}
	// Remove ONLY the new authorised purpose drawing run. The entire source
	// document, including fixed cover text/rPr/paragraph spacing, is identical.
	var purpose *mngWordNode
	for _, p := range tree.all(mngWordNS, "body")[0].children {
		if p.name.Local == "p" {
			for _, r := range p.children {
				props := r.all(mngDrawingNS, "docPr")
				if len(props) == 1 && props[0].attr("title") == "mng-test-purpose" {
					purpose = r
				}
			}
		}
	}
	if purpose == nil {
		t.Fatal("missing floating marker")
	}
	body := after["word/document.xml"]
	if !bytes.Equal(append(append([]byte{}, body[:purpose.start]...), body[purpose.end:]...), before["word/document.xml"]) {
		t.Fatal("unauthorised source XML mutation")
	}
	source[0] ^= 1
	if _, err := BuildManagementTraitsTestTemplate(source); err == nil {
		t.Fatal("untrusted source accepted")
	}
}
