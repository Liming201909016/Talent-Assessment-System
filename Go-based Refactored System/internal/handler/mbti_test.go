package handler

import (
	"testing"
)

// ============================================================
// 回归测试 — FB-006 / FB-007 / FB-008 (mbti.calcMbtiScores 业务规则)
// 对应：docs/regression-tests.md FB-006/007/008
// 说明：以下测试以 RED→GREEN 方式驱动 mbti 评分函数的健壮性提升
// ============================================================

// TestBugFB006_AllZeroScoreRejectsType
// FB-006: 全 0 分仍生成 INFP 报告（默认同分选 I/N/F/P）会误导客户
// 期望：纯函数 aggregateMbtiScores 输出结果中应有 totalAnswered 字段
//
//	调用方据此拒绝生成报告
func TestBugFB006_AllZeroScoreReportEmpty(t *testing.T) {
	// GIVEN: 没有任何答题记录
	rows := []mbtiAnswerRow{}

	// WHEN: 计算分数
	scores, mbtiType, totalAnswered := aggregateMbtiScores(rows)

	// THEN: 总答题数 = 0，应被业务规则拒绝
	if totalAnswered != 0 {
		t.Errorf("totalAnswered: want 0, got %d", totalAnswered)
	}
	// 8 个维度都应为 0
	for k, v := range scores {
		if v != 0 {
			t.Errorf("dimension %s: want 0, got %d", k, v)
		}
	}
	// type 仍会按默认规则生成（INFP），但调用方应根据 totalAnswered 拒绝使用
	if mbtiType == "" {
		t.Errorf("mbtiType should not be empty even on zero scores")
	}
}

// TestBugFB006_IsValidMbtiSubmission
// FB-006: IsValidMbtiSubmission 应根据答题数判断
func TestBugFB006_IsValidMbtiSubmission(t *testing.T) {
	tests := []struct {
		name     string
		answered int
		want     bool
	}{
		{"零答题", 0, false},
		{"答 1 题", 1, false},
		{"答 23 题（< 阈值）", 23, false},
		{"答 24 题（= 阈值，半数）", 24, true},
		{"答 47 题", 47, true},
		{"答 48 题（满）", 48, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidMbtiSubmission(tt.answered)
			if got != tt.want {
				t.Errorf("answered=%d: want %v, got %v", tt.answered, tt.want, got)
			}
		})
	}
}

// TestBugFB008_NonVPrefixContentSilentlyIgnored
// FB-008: 非 V1-V48 格式题号被静默忽略
// 期望：aggregateMbtiScores 返回 invalidCount，调用方可记日志告警
func TestBugFB008_InvalidContentReported(t *testing.T) {
	rows := []mbtiAnswerRow{
		{Content: "V1", ScoreA: 5, ScoreB: 0},  // 有效
		{Content: "V2", ScoreA: 3, ScoreB: 2},  // 有效
		{Content: "BAD", ScoreA: 1, ScoreB: 4}, // 无效
		{Content: "V99", ScoreA: 2, ScoreB: 3}, // 超范围
		{Content: "V0", ScoreA: 4, ScoreB: 1},  // 超范围
	}

	scores, _, totalAnswered := aggregateMbtiScores(rows)
	invalid := CountInvalidMbtiAnswers(rows)

	if totalAnswered != 2 {
		t.Errorf("totalAnswered: want 2 (only V1+V2), got %d", totalAnswered)
	}
	if invalid != 3 {
		t.Errorf("invalid count: want 3 (BAD/V99/V0), got %d", invalid)
	}
	if scores["E"] != 5 || scores["I"] != 0 {
		t.Errorf("V1 (E-I): want E=5 I=0, got E=%d I=%d", scores["E"], scores["I"])
	}
	if scores["S"] != 3 || scores["N"] != 2 {
		t.Errorf("V2 (S-N): want S=3 N=2, got S=%d N=%d", scores["S"], scores["N"])
	}
}

// TestAggregateMbtiScores_AllDimensions
// 完整覆盖 4 个维度的题号映射
func TestAggregateMbtiScores_AllDimensions(t *testing.T) {
	// 制造每个维度首尾 + 中间各 2 题的答案，验证累加
	rows := []mbtiAnswerRow{
		// E-I (mod==1): 1, 5, 45
		{Content: "V1", ScoreA: 5, ScoreB: 0},
		{Content: "V5", ScoreA: 4, ScoreB: 1},
		{Content: "V45", ScoreA: 3, ScoreB: 2},
		// S-N (mod==2): 2, 46
		{Content: "V2", ScoreA: 5, ScoreB: 0},
		{Content: "V46", ScoreA: 5, ScoreB: 0},
		// T-F (mod==3): 3, 47
		{Content: "V3", ScoreA: 0, ScoreB: 5},
		{Content: "V47", ScoreA: 1, ScoreB: 4},
		// J-P (mod==0): 4, 48
		{Content: "V4", ScoreA: 5, ScoreB: 0},
		{Content: "V48", ScoreA: 4, ScoreB: 1},
	}

	scores, mbtiType, total := aggregateMbtiScores(rows)

	if total != 9 {
		t.Errorf("total: want 9, got %d", total)
	}
	// E = 5+4+3 = 12, I = 0+1+2 = 3 → E
	if scores["E"] != 12 || scores["I"] != 3 {
		t.Errorf("E/I: want 12/3, got %d/%d", scores["E"], scores["I"])
	}
	// S = 5+5 = 10, N = 0 → S
	if scores["S"] != 10 || scores["N"] != 0 {
		t.Errorf("S/N: want 10/0, got %d/%d", scores["S"], scores["N"])
	}
	// T = 0+1 = 1, F = 5+4 = 9 → F
	if scores["T"] != 1 || scores["F"] != 9 {
		t.Errorf("T/F: want 1/9, got %d/%d", scores["T"], scores["F"])
	}
	// J = 5+4 = 9, P = 0+1 = 1 → J
	if scores["J"] != 9 || scores["P"] != 1 {
		t.Errorf("J/P: want 9/1, got %d/%d", scores["J"], scores["P"])
	}
	// 类型：E S F J = "ESFJ"
	if mbtiType != "ESFJ" {
		t.Errorf("type: want ESFJ, got %s", mbtiType)
	}
}

// TestAggregateMbtiScores_TieBreaking
// 验证同分时的默认选择：I, N, F, P
func TestAggregateMbtiScores_TieBreaking(t *testing.T) {
	rows := []mbtiAnswerRow{}
	_, mbtiType, _ := aggregateMbtiScores(rows)
	// 全零 → 4 个维度都是 0=0 → I, N, F, P
	if mbtiType != "INFP" {
		t.Errorf("all-zero tie: want INFP, got %s", mbtiType)
	}
}
