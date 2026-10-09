package handler

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
)

// Offline vertical slice only: no database/profile/run creation, no old PDF path.
func TestManagementTraitsAuditedReissueRealPDF(t *testing.T) {
	dir := os.Getenv("MNG_REISSUE_ARTIFACT_DIR")
	if dir == "" {
		t.Skip("explicit offline synthetic artifact directory required")
	}
	if !filepath.IsAbs(dir) {
		t.Fatal("absolute output required")
	}
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	raw, sig, key := read("source.json"), read("audit-signature.bin"), read("synthetic-auditor-public.bin")
	var source struct {
		Synthetic bool `json:"synthetic"`
	}
	if json.Unmarshal(raw, &source) != nil || !source.Synthetic {
		t.Fatal("only synthetic local fixture allowed")
	}
	workbook, err := os.ReadFile(filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-content-v1.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	content, err := service.LoadManagementTraitsTestContent(workbook)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := service.BuildManagementTraitsAuditedReissue(raw, sig, ed25519.PublicKey(key), content, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(snapshot.SnapshotJSON(), read("snapshot.json")) {
		t.Fatal("snapshot changed between stages")
	}
	dto, err := snapshot.ReportData()
	if err != nil {
		t.Fatal(err)
	}
	if dto.Overall.Score != "50.00" || dto.Participant.Name != "SYNTHETIC REISSUE SAMPLE" {
		t.Fatal("unexpected fixture")
	}
	template := managementTestWordTemplate(t)
	docx, err := renderManagementTraitsTestWord(template, dto)
	if err != nil {
		t.Fatal(err)
	}
	before, after := managementWordParts(t, template), managementWordParts(t, docx)
	if len(before) != len(after) {
		t.Fatal("OPC cardinality changed")
	}
	unchanged := 0
	for name, b := range before {
		if name != "word/document.xml" && !strings.HasPrefix(name, "word/charts/chart") {
			if !bytes.Equal(b, after[name]) {
				t.Fatal("protected OPC changed", name)
			}
			unchanged++
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	pdf, err := libreofficepdf.NewClient("").Convert(ctx, "audited-reissue.docx", docx)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) || len(pdf) < 1024 {
		t.Fatal("invalid PDF")
	}
	sha := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	receipt, err := json.Marshal(map[string]interface{}{"synthetic": true, "sourceSHA": sha(raw), "snapshotSHA": sha(snapshot.SnapshotJSON()), "templateSHA": sha(template), "contentSHA": sha(workbook), "docxSHA": sha(docx), "pdfSHA": sha(pdf), "pdfBytes": len(pdf), "opcParts": len(before), "protectedUnchangedParts": unchanged, "oldPDFPathWrites": 0, "databaseCalls": 0, "remoteOperations": 0})
	if err != nil {
		t.Fatal(err)
	}
	for name, b := range map[string][]byte{"report.docx": docx, "report.pdf": pdf, "render-receipt.json": receipt} {
		f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		_, e := f.Write(b)
		closeErr := f.Close()
		if e != nil || closeErr != nil {
			t.Fatal("artifact write failed")
		}
	}
	t.Logf("synthetic audited reissue PDF %d bytes SHA %s; protected OPC %d/%d", len(pdf), sha(pdf), unchanged, len(before))
}
