package handler

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

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
