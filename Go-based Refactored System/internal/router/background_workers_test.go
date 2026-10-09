package router

import "testing"

func TestBackgroundWorkersStartByDefault(t *testing.T) {
	calls := 0
	started := startBackgroundWorker(defaultSetupOptions().BackgroundWorkersEnabled, func() { calls++ })
	if !started || calls != 1 {
		t.Fatalf("default background worker startup mismatch: started=%t calls=%d", started, calls)
	}
}

func TestBackgroundWorkersCanBeDisabledWithoutStarting(t *testing.T) {
	calls := 0
	started := startBackgroundWorker(false, func() { calls++ })
	if started || calls != 0 {
		t.Fatalf("disabled background worker started: started=%t calls=%d", started, calls)
	}
}
