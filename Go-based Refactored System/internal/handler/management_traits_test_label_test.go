package handler

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
)

// MT-LABEL-02: purpose must be visible in actual PDF text, not metadata or
// an unsupported DrawingML Choice. Corresponds to docs/regression-tests.md.
func TestBugManagementTraitsPDFVisibleTestLabel(t *testing.T) {
	if os.Getenv("MNG_REAL_LO_TEST") != "1" {
		t.Skip("explicit MNG_REAL_LO_TEST=1 required for real LibreOffice evidence")
	}
	dto := managementTestWordDTO()
	if input := os.Getenv("MNG_TEST_REPORT_DATA_INPUT"); input != "" {
		raw, err := os.ReadFile(input)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &dto); err != nil {
			t.Fatal(err)
		}
	}
	docx, err := renderManagementTraitsTestWord(managementTestWordTemplate(t), dto)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	pdf, err := libreofficepdf.NewClient("").Convert(ctx, "mng-label-test.docx", docx)
	if err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(t.TempDir(), "report.pdf")
	if err := os.WriteFile(f, pdf, 0600); err != nil {
		t.Fatal(err)
	}
	raw, err := exec.CommandContext(ctx, "pdftotext", "-layout", "-enc", "UTF-8", f, "-").Output()
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, required := range []string{"TEST", "测试报告", "不可作为人才决策依据"} {
		if !strings.Contains(text, required) {
			t.Fatalf("PDF lacks visible purpose marker %q", required)
		}
	}
	if strings.Contains(text, "report.testLabel") || strings.Contains(text, "\ufffd") {
		t.Fatal("unbound label or invalid text")
	}
	if output := os.Getenv("MNG_REAL_LO_PDF_OUTPUT"); output != "" {
		file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			t.Fatal(err)
		}
		_, err = file.Write(pdf)
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			t.Fatal("cannot write isolated TEST PDF")
		}
	}
	if os.Getenv("MNG_TEST_REPORT_DATA_INPUT") != "" {
		// Layout extraction interleaves the left table-cell label between
		// wrapped lines of the right cell. Verify complete literal paragraphs
		// in PDF reading order as well; do not delete labels or source words.
		readingOrder, err := exec.CommandContext(ctx, "pdftotext", "-raw", "-enc", "UTF-8", f, "-").Output()
		if err != nil {
			t.Fatal(err)
		}
		normal := strings.Join(strings.Fields(string(readingOrder)), "")
		for _, d := range dto.Dimensions {
			for _, v := range []string{d.Name, d.Diagnosis, d.Advice} {
				if !strings.Contains(normal, strings.Join(strings.Fields(v), "")) {
					t.Fatalf("missing full customer text for %s", d.Key)
				}
			}
		}
		for _, v := range append([]string{dto.Overall.Diagnosis}, dto.Overall.Advice...) {
			if !strings.Contains(normal, strings.Join(strings.Fields(v), "")) {
				t.Fatal("missing full overall text")
			}
		}
		if dto.ContentSourceSHA != "b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c" {
			t.Fatal("content source mismatch")
		}
	}
	t.Logf("real PDF bytes=%d purpose markers visible", len(pdf))
}
