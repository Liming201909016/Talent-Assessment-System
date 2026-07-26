package service

import (
	"os"
	"strings"
	"testing"
)

func TestCompetencyExpiryWorker_BatchResolvesOwnersWithoutNPlusOne(t *testing.T) {
	source := readWorkerSource(t)
	body := extractWorkerFunction(t, source, "func (w *CompetencyExpiryWorker) RunOnce(")
	for _, required := range []string{
		"LEFT JOIN el_tester", "LEFT JOIN el_candidate", "participant_type",
		"p.limit_time ASC, p.id ASC", "Limit(w.batch)", "w.svc.Submit",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("RunOnce batch owner resolution missing %q", required)
		}
	}
	loopAt := strings.Index(body, "for _, row := range rows")
	if loopAt < 0 {
		t.Fatal("RunOnce paper loop missing")
	}
	for _, forbidden := range []string{"Table(\"el_tester\")", "Table(\"el_candidate\")", ".Count(&count)"} {
		if strings.Contains(body[loopAt:], forbidden) {
			t.Errorf("RunOnce performs N+1 owner query inside loop: %q", forbidden)
		}
	}
}

func TestCompetencyExpiryWorker_StartScansImmediately(t *testing.T) {
	body := extractWorkerFunction(t, readWorkerSource(t), "func (w *CompetencyExpiryWorker) Start(")
	runAt := strings.Index(body, "w.RunOnce(ctx)")
	tickerSelectAt := strings.Index(body, "case <-ticker.C")
	if runAt < 0 || tickerSelectAt < 0 || runAt > tickerSelectAt {
		t.Fatalf("Start must run one scan before waiting for ticker: run=%d ticker=%d", runAt, tickerSelectAt)
	}
	if !strings.Contains(body, "ctx.Err()") {
		t.Error("Start must avoid initial scan after context cancellation")
	}
}

func TestCompetencyExpiryWorker_IsolatesPerPaperFailures(t *testing.T) {
	body := extractWorkerFunction(t, readWorkerSource(t), "func (w *CompetencyExpiryWorker) RunOnce(")
	for _, required := range []string{
		`if row.ParticipantType == ""`, `slog.Error("competency expired paper owner missing"`,
		`continue`, `slog.Error("competency timeout submit failed"`,
	} {
		if !strings.Contains(body, required) {
			t.Errorf("RunOnce failure isolation missing %q", required)
		}
	}
	loopAt := strings.Index(body, "for _, row := range rows")
	if loopAt < 0 || strings.Contains(body[loopAt:], "return err") {
		t.Error("a per-paper owner/submit failure must not abort the remaining batch")
	}
}

func TestShuffleCompetencyQuestionIDs_OneHundredFullCapacityOrders(t *testing.T) {
	const questionCount = 384
	original := make([]string, questionCount)
	for index := range original {
		original[index] = strings.Repeat("q", 1) + string(rune(0x1000+index))
	}
	orders := make(map[string]struct{})
	for run := 0; run < 100; run++ {
		ids := append([]string(nil), original...)
		if err := ShuffleCompetencyQuestionIDs(ids, nil); err != nil {
			t.Fatal(err)
		}
		seen := make(map[string]struct{}, questionCount)
		for _, id := range ids {
			seen[id] = struct{}{}
		}
		if len(ids) != questionCount || len(seen) != questionCount {
			t.Fatalf("run %d lost/duplicated IDs: length=%d unique=%d", run, len(ids), len(seen))
		}
		orders[strings.Join(ids, "|")] = struct{}{}
	}
	if len(orders) < 2 {
		t.Fatalf("100 secure shuffles produced only %d distinct order", len(orders))
	}
}

func readWorkerSource(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("competency_worker.go")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func extractWorkerFunction(t *testing.T, source, signature string) string {
	t.Helper()
	start := strings.Index(source, signature)
	if start < 0 {
		t.Fatalf("signature not found: %s", signature)
	}
	open := strings.Index(source[start:], "{")
	if open < 0 {
		t.Fatalf("opening brace not found: %s", signature)
	}
	open += start
	depth := 0
	for index := open; index < len(source); index++ {
		switch source[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[start : index+1]
			}
		}
	}
	t.Fatalf("function end not found: %s", signature)
	return ""
}
