package handler

import (
	"strings"
	"testing"
)

// TestBugFB065_CandidateCreateUsesConcurrentSafeID
// 对应：docs/regression-tests.md FB-065
// 复现：100份试卷容量链以10并发创建candidate时，UnixMilli在同一毫秒生成重复主键并触发MySQL 1062。
// 期望：新candidate使用全局原子ID生成器，不直接使用time.Now().UnixMilli()。
func TestBugFB065_CandidateCreateUsesConcurrentSafeID(t *testing.T) {
	body := extractFunctionBody(t, readSourceFile(t, "candidate.go"), "func (h *CandidateHandler) Save(")
	if !strings.Contains(body, "nextID()") {
		t.Error("Candidate.Save must use the concurrent-safe global ID generator")
	}
	if strings.Contains(body, "time.Now().UnixMilli()") {
		t.Error("Candidate.Save still uses millisecond timestamps that collide under concurrency")
	}
}
