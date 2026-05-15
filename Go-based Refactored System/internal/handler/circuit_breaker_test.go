package handler

import (
	"testing"
	"time"
)

// newCBHandler 构造一个最小 ExamHandler 用于熔断测试（不需要 db/cfg）
func newCBHandler() *ExamHandler {
	return &ExamHandler{}
}

// TestCircuitBreaker_NotOpenInitially 初始状态熔断未打开
func TestCircuitBreaker_NotOpenInitially(t *testing.T) {
	h := newCBHandler()
	if h.circuitOpen() {
		t.Error("circuit should be closed initially")
	}
}

// TestCircuitBreaker_OpensAfter3Failures 连续 3 次失败后打开
func TestCircuitBreaker_OpensAfter3Failures(t *testing.T) {
	h := newCBHandler()
	h.recordFailure()
	if h.circuitOpen() {
		t.Error("circuit should still be closed after 1 failure")
	}
	h.recordFailure()
	if h.circuitOpen() {
		t.Error("circuit should still be closed after 2 failures")
	}
	h.recordFailure()
	if !h.circuitOpen() {
		t.Error("circuit SHOULD be open after 3 failures")
	}
}

// TestCircuitBreaker_SuccessResets 成功重置失败计数
func TestCircuitBreaker_SuccessResets(t *testing.T) {
	h := newCBHandler()
	h.recordFailure()
	h.recordFailure()
	h.recordSuccess()
	// Now 2 more failures should NOT open (counter was reset)
	h.recordFailure()
	h.recordFailure()
	if h.circuitOpen() {
		t.Error("circuit should NOT be open: success reset counter, only 2 fails since")
	}
	h.recordFailure()
	if !h.circuitOpen() {
		t.Error("circuit SHOULD be open: 3 fails after reset")
	}
}

// TestCircuitBreaker_CooldownExpires 冷却窗口过期后熔断关闭
func TestCircuitBreaker_CooldownExpires(t *testing.T) {
	h := newCBHandler()
	h.recordFailure()
	h.recordFailure()
	h.recordFailure()
	if !h.circuitOpen() {
		t.Fatal("setup: circuit must open")
	}
	// Manually expire cooldown
	h.cbMu.Lock()
	h.cbOpenUntil = time.Now().Add(-1 * time.Second)
	h.cbMu.Unlock()
	if h.circuitOpen() {
		t.Error("circuit should be closed after cooldown expires")
	}
}

// TestCircuitBreaker_OpenWindowDuration 熔断窗口约 60s
func TestCircuitBreaker_OpenWindowDuration(t *testing.T) {
	h := newCBHandler()
	before := time.Now()
	h.recordFailure()
	h.recordFailure()
	h.recordFailure()
	h.cbMu.Lock()
	until := h.cbOpenUntil
	h.cbMu.Unlock()
	gap := until.Sub(before)
	if gap < 50*time.Second || gap > 70*time.Second {
		t.Errorf("expected ~60s cooldown, got %v", gap)
	}
}

// TestCircuitBreaker_ConcurrentSafe 并发调用不死锁
func TestCircuitBreaker_ConcurrentSafe(t *testing.T) {
	h := newCBHandler()
	done := make(chan bool, 100)
	for i := 0; i < 50; i++ {
		go func() { h.recordFailure(); done <- true }()
		go func() { h.recordSuccess(); done <- true }()
	}
	for i := 0; i < 100; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("deadlock or hang in concurrent recordFailure/recordSuccess")
		}
	}
	// circuitOpen() should not panic regardless of state
	_ = h.circuitOpen()
}
