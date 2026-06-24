package handler

import (
	"strings"
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

// TestBugFB042_ReplaceDocumentFieldsStripsW14EffectsFromBody
// 对应：docs/regression-tests.md FB-042
// 复现：完整版模板正文静态段落仍保留 w14:textFill / w14:props3d，导致 PDF 渲染方框字
// 期望：生成的 document.xml 不应再包含这些高风险 w14 特效
func TestBugFB042_ReplaceDocumentFieldsStripsW14EffectsFromBody(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document>
  <w:body>
    <w:p>
      <w:r><w:t>姓名：</w:t></w:r>
      <w:r><w:t>____</w:t></w:r>
    </w:p>
    <w:p>
      <w:r>
        <w:rPr>
          <w:rFonts w:ascii="微软雅黑" w:hAnsi="微软雅黑" w:eastAsia="微软雅黑"/>
          <w14:textFill><w14:solidFill><w14:srgbClr val="75BD42"/></w14:solidFill></w14:textFill>
          <w14:props3d w14:extrusionH="57150" w14:contourW="0" w14:prstMaterial="softEdge"/>
        </w:rPr>
        <w:t>适配类型 1：ENFP</w:t>
      </w:r>
    </w:p>
    <w:p>
      <w:r><w:t>2026年XX月XX日</w:t></w:r>
    </w:p>
  </w:body>
</w:document>`

	got := (&MbtiReportHandler{}).replaceDocumentFields([]byte(content), map[string]string{
		"姓名：": "Liming",
	}, "2026年6月24日")

	if !strings.Contains(string(got), "Liming") {
		t.Fatalf("expected field replacement to remain intact")
	}
	if strings.Contains(string(got), "w14:textFill") {
		t.Fatalf("expected w14:textFill to be stripped from full report body")
	}
	if strings.Contains(string(got), "w14:props3d") {
		t.Fatalf("expected w14:props3d to be stripped from full report body")
	}
	if !strings.Contains(string(got), "适配类型 1：ENFP") {
		t.Fatalf("expected static body text to remain present")
	}
}

// TestBugFB043_ReplaceDocumentFieldsNormalizesRiskyFonts
// 对应：docs/regression-tests.md FB-043
// 复现：完整版静态段落使用汉仪字体族时，PDF 渲染可能出现方框字
// 期望：生成 document.xml 时将高风险字体替换为稳定字体 Noto Sans CJK SC
func TestBugFB043_ReplaceDocumentFieldsNormalizesRiskyFonts(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document>
	<w:body>
		<w:p>
			<w:r>
				<w:rPr>
					<w:rFonts w:ascii="汉仪雅酷黑 75W" w:hAnsi="汉仪雅酷黑 75W" w:eastAsia="汉仪雅酷黑 75W"/>
				</w:rPr>
				<w:t>职场角色：</w:t>
			</w:r>
		</w:p>
		<w:p>
			<w:r><w:t>姓名：</w:t></w:r>
			<w:r><w:t>____</w:t></w:r>
		</w:p>
		<w:p>
			<w:r><w:t>2026年XX月XX日</w:t></w:r>
		</w:p>
	</w:body>
</w:document>`

	got := (&MbtiReportHandler{}).replaceDocumentFields([]byte(content), map[string]string{
		"姓名：": "Liming",
	}, "2026年6月24日")
	out := string(got)

	if strings.Contains(out, "汉仪雅酷黑") {
		t.Fatalf("expected risky font family to be normalized")
	}
	if !strings.Contains(out, "Noto Sans CJK SC") {
		t.Fatalf("expected fallback stable font family to be present")
	}
}

// TestBugFB044_ReplaceDocumentFieldsStabilizesEastAsiaHintOnlyFonts
// 对应：docs/regression-tests.md FB-044
// 复现：ESTP 模板中 "功利型"、"凭借" 等正文 run 只有 <w:rFonts w:hint="eastAsia"/>，
// LibreOffice/Linux 会选用不稳定 fallback/subset，导致 PDF 渲染方框字
// 期望：生成 document.xml 时为 hint-only 东亚字体 run 补齐 Noto Sans CJK SC
func TestBugFB044_ReplaceDocumentFieldsStabilizesEastAsiaHintOnlyFonts(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document>
	<w:body>
		<w:p>
			<w:r>
				<w:rPr>
					<w:rFonts w:hint="eastAsia"/>
					<w:sz w:val="24"/>
				</w:rPr>
				<w:t>功利型</w:t>
			</w:r>
		</w:p>
		<w:p>
			<w:r><w:t>姓名：</w:t></w:r>
			<w:r><w:t>____</w:t></w:r>
		</w:p>
		<w:p>
			<w:r><w:t>2026年XX月XX日</w:t></w:r>
		</w:p>
	</w:body>
</w:document>`

	got := (&MbtiReportHandler{}).replaceDocumentFields([]byte(content), map[string]string{
		"姓名：": "Liming",
	}, "2026年6月24日")
	out := string(got)

	if strings.Contains(out, `<w:rFonts w:hint="eastAsia"/>`) {
		t.Fatalf("expected hint-only eastAsia font tag to be stabilized")
	}
	if !strings.Contains(out, `w:eastAsia="Noto Sans CJK SC"`) {
		t.Fatalf("expected eastAsia stable fallback font to be present")
	}
	if !strings.Contains(out, "功利型") {
		t.Fatalf("expected static ESTP role text to remain present")
	}
}
