package handler

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNormalizeCompetencyReportPaperIDs(t *testing.T) {
	ids, err := normalizeCompetencyReportPaperIDs([]string{" paper-1 ", "paper-1", "paper-2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "paper-1" || ids[1] != "paper-2" {
		t.Fatalf("normalized ids=%v", ids)
	}
	if _, err := normalizeCompetencyReportPaperIDs(nil); err == nil {
		t.Fatal("empty paper ids were accepted")
	}
	if _, err := normalizeCompetencyReportPaperIDs([]string{"paper-1", " "}); err == nil {
		t.Fatal("blank paper id was accepted")
	}
	tooMany := make([]string, 101)
	for index := range tooMany {
		tooMany[index] = "paper-" + strings.Repeat("x", index+1)
	}
	if _, err := normalizeCompetencyReportPaperIDs(tooMany); err == nil {
		t.Fatal("more than 100 paper ids were accepted")
	}
}

// TestBugFB157_BatchDownloadBuildsOneZipWithEverySelectedReport
// 对应：docs/regression-tests.md #FB-157
// 复现：前端逐份保存PDF，浏览器产生多个独立下载。
// 期望：后端一次生成ZIP，所选PDF全部存在且同名人员的文件名仍唯一。
func TestBugFB157_BatchDownloadBuildsOneZipWithEverySelectedReport(t *testing.T) {
	tempDir := t.TempDir()
	firstPath := filepath.Join(tempDir, "first.pdf")
	secondPath := filepath.Join(tempDir, "second.pdf")
	if err := os.WriteFile(firstPath, []byte("%PDF-first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, []byte("%PDF-second"), 0o600); err != nil {
		t.Fatal(err)
	}

	entries := []competencyReportArchiveEntry{
		{PaperID: "paper/1", ParticipantName: "../同名/人员", Path: firstPath},
		{PaperID: "paper-2", ParticipantName: "同名人员", Path: secondPath},
	}
	var output bytes.Buffer
	if err := writeCompetencyReportArchive(&output, entries); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	if len(reader.File) != 2 {
		t.Fatalf("archive entries=%d, want 2", len(reader.File))
	}
	if reader.File[0].Name == reader.File[1].Name {
		t.Fatalf("duplicate archive filename: %s", reader.File[0].Name)
	}
	for _, file := range reader.File {
		if strings.Contains(file.Name, "/") || strings.Contains(file.Name, "\\") {
			t.Fatalf("archive entry contains a path separator: %s", file.Name)
		}
	}
	for index, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			t.Fatal(err)
		}
		want := []byte("%PDF-first")
		if index == 1 {
			want = []byte("%PDF-second")
		}
		if !bytes.Equal(content, want) {
			t.Fatalf("entry %d content=%q, want %q", index, content, want)
		}
	}
}

// TestBugFB228_BatchDownloadUsesCurrentV2ReportBinding
// 对应：docs/regression-tests.md #FB-228
// 复现：v1历史结果已重算为v2报告，单份下载使用current指针成功，批量下载仍按v1版本元组查找并提示报告尚未生成。
// 期望：批量下载优先使用current→completed report→completed v2 run绑定；仅无current时回退legacy版本元组。
func TestBugFB228_BatchDownloadUsesCurrentV2ReportBinding(t *testing.T) {
	source := readSourceFile(t, "competency_report.go")
	loader := extractFunctionBody(t, source, "func (h *CompetencyReportHandler) loadCompetencyReportArchiveEntries(paperIDs []string) ([]competencyReportArchiveEntry, error) {")
	for _, required := range []string{
		"var currents []model.CompetencyReportCurrent",
		"paper_id IN ? AND audience = ?",
		"currentByPaper",
		"reportByID",
		"runByID",
		"ValidatePhase1ReportContentApprovalForEnvironment",
	} {
		if !strings.Contains(loader, required) {
			t.Fatalf("batch download current-v2 binding missing %q", required)
		}
	}
}

// TestBugFB080_SamePaperGenerationIsSerialized
// 对应：docs/regression-tests.md #FB-080
// 复现：同一 paperId 的并发请求可同时进入实例查询、Chromium 渲染和文件替换。
// 期望：同一 paperId 的生成临界区最多只有一个执行者。
func TestBugFB080_SamePaperGenerationIsSerialized(t *testing.T) {
	h := &CompetencyReportHandler{}
	start := make(chan struct{})
	var active int32
	var maximum int32
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			unlock := h.lockReportGeneration("paper-1")
			defer unlock()
			current := atomic.AddInt32(&active, 1)
			for {
				seen := atomic.LoadInt32(&maximum)
				if current <= seen || atomic.CompareAndSwapInt32(&maximum, seen, current) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&active, -1)
		}()
	}
	close(start)
	wg.Wait()

	if maximum != 1 {
		t.Fatalf("same-paper maximum concurrent generators=%d, want 1", maximum)
	}
}

// TestBugFB084_ReportCompletionAndAuditAreAtomic
// 对应：docs/regression-tests.md #FB-084
// 复现：PDF路径和completed状态先提交，成功审计随后单独写；审计失败会返回失败但保留成功产物。
// 期望：completed元数据、人员PDF状态和成功审计在同一数据库事务中提交。
func TestBugFB084_ReportCompletionAndAuditAreAtomic(t *testing.T) {
	source := readSourceFile(t, "competency_report.go")
	transaction := extractFunctionBody(t, source, "if err := h.db.Transaction(func(tx *gorm.DB) error {")
	if !strings.Contains(transaction, "writeReportAuditWithDB(tx") {
		t.Fatal("report completion transaction does not include the success audit insert")
	}
	if !strings.Contains(source, "func (h *CompetencyReportHandler) writeReportAuditWithDB(db *gorm.DB") {
		t.Fatal("report audit helper cannot participate in the caller transaction")
	}
}

// TestBugFB085_ReportFilenameUsesRFC5987PercentEncoding
// 对应：docs/regression-tests.md #FB-085
func TestBugFB085_ReportFilenameUsesRFC5987PercentEncoding(t *testing.T) {
	encoded := encodeRFC5987FileName("胜任力 测试+报告.pdf")
	if strings.Contains(encoded, "+") {
		t.Fatalf("RFC5987 filename contains plus-space encoding: %s", encoded)
	}
	if !strings.Contains(encoded, "%20") || !strings.Contains(encoded, "%2B") {
		t.Fatalf("RFC5987 filename is not percent encoded: %s", encoded)
	}
}

// TestBugFB106_Phase1ApprovedReportDownloadSkipsGenericVersionValidator
// 对应：docs/regression-tests.md #FB-106
// 复现：一期报告已生成，但下载在专属批准门禁后继续调用只支持generic的版本校验器。
// 期望：一期走专属批准校验；仅generic报告调用ValidateFrozenCompetencyVersionSet。
func TestBugFB106_Phase1ApprovedReportDownloadSkipsGenericVersionValidator(t *testing.T) {
	source := readSourceFile(t, "competency_report.go")
	download := extractFunctionBody(t, source, "func (h *CompetencyReportHandler) Download(c *gin.Context) {")
	if !strings.Contains(download, "if service.IsPhase1CompetencyVersionSet(versions) {") {
		t.Fatal("phase-1 download approval branch missing")
	}
	if !strings.Contains(download, "} else if err := service.ValidateFrozenCompetencyVersionSet(versions); err != nil {") {
		t.Fatal("generic version validator still runs after the phase-1 approval branch")
	}
}

func TestReportGenerationLockIndex_IsStableAndBounded(t *testing.T) {
	for _, paperID := range []string{"paper-1", "paper-2", "fc743d2b-0b72-49b2-803f-f285d62730ed"} {
		first := reportGenerationLockIndex(paperID)
		second := reportGenerationLockIndex(paperID)
		if first != second {
			t.Fatalf("unstable lock index for %q: %d != %d", paperID, first, second)
		}
		if first < 0 || first >= competencyReportLockStripes {
			t.Fatalf("lock index out of range for %q: %d", paperID, first)
		}
	}
}
