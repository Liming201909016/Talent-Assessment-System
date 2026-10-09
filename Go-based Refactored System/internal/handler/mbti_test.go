package handler

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestBugFB218_MBtiLibreOfficeConversionHasDeadline
// 对应：docs/regression-tests.md #FB-218
// 复现：staging MBTI 完整版报告请求超过客户端180秒，LibreOffice子进程无截止时间。
// 期望：转换命令使用有界context，超时后返回context deadline exceeded。
func TestBugFB218_MBtiLibreOfficeConversionHasDeadline(t *testing.T) {
	source, err := os.ReadFile("mbti_report.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{"context.WithTimeout", "libreofficepdf.NewClient", ".Convert(ctx"} {
		if !strings.Contains(text, required) {
			t.Fatalf("MBTI report conversion missing bounded shared client marker %q", required)
		}
	}
	if strings.Contains(text, "exec.Command(loCmd") || strings.Contains(text, "exec.CommandContext(ctx, loCmd") {
		t.Fatal("MBTI report conversion still starts LibreOffice directly")
	}
}

// TestBugFB218_MBtiAsyncConversionFailureDoesNotPublishDocx
// 对应：docs/regression-tests.md #FB-218
// 复现：异步报告转换失败后仍把DOCX路径写入pdf_path并设置pdf_flag=1。
// 期望：失败返回空发布路径、清理临时DOCX，调用方不得写成功状态。
func TestBugFB218_MBtiAsyncConversionFailureDoesNotPublishDocx(t *testing.T) {
	dir := t.TempDir()
	docx := filepath.Join(dir, "report.docx")
	if err := os.WriteFile(docx, []byte("docx"), 0o600); err != nil {
		t.Fatal(err)
	}
	published, err := finalizeMbtiPDFConversion(docx, dir, func(string, string) (string, error) {
		return docx, errors.New("conversion failed")
	})
	if err == nil || published != "" {
		t.Fatalf("published=%q err=%v, want empty path and conversion error", published, err)
	}
	if _, statErr := os.Stat(docx); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("temporary DOCX remains after conversion failure: %v", statErr)
	}
}

func TestBugFB218_MBtiAsyncFailurePerformsZeroPublicationWrites(t *testing.T) {
	db, mock := identityHandlerDB(t)
	paperID := "fb218-paper"
	mock.ExpectQuery("SELECT `pdf_path` FROM `el_candidate`").WillReturnRows(sqlmock.NewRows([]string{"pdf_path"}))
	mock.ExpectQuery("SELECT `pdf_path` FROM `el_tester`").WillReturnRows(sqlmock.NewRows([]string{"pdf_path"}))
	mock.ExpectQuery("SELECT q.content, ma.score_a, ma.score_b").WillReturnRows(sqlmock.NewRows([]string{"content", "score_a", "score_b"}).AddRow("V1", 3, 2))
	mock.ExpectQuery("SELECT .* FROM `el_tester`").WillReturnRows(sqlmock.NewRows([]string{"name", "age", "gender", "telephone", "affiliation", "post", "exam_id"}).AddRow("Synthetic", 30, "0", "18800000000", "", "", "exam"))
	mock.ExpectQuery("SELECT `user_time` FROM `el_paper`").WillReturnRows(sqlmock.NewRows([]string{"user_time"}).AddRow(1))
	mock.ExpectQuery("SELECT `required_fields` FROM `el_exam`").WillReturnRows(sqlmock.NewRows([]string{"required_fields"}).AddRow("name,gender,telephone"))
	h := &MbtiReportHandler{
		db: db, templateDir: filepath.Join("..", "..", "deploy", "mbti-templates"),
		simpleDir: filepath.Join("..", "..", "deploy", "mbti-templates-simple"), outputDir: t.TempDir(),
		convertPDF: func(string, string) (string, error) { return "", errors.New("synthetic conversion failure") },
	}
	h.GenerateReportByPaperID(paperID)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("conversion failure executed an unexpected publication write", err)
	}
	entries, err := os.ReadDir(h.outputDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("unexpected output root state: entries=%d err=%v", len(entries), err)
	}
	dayEntries, err := os.ReadDir(filepath.Join(h.outputDir, entries[0].Name()))
	if err != nil || len(dayEntries) != 0 {
		t.Fatalf("temporary DOCX/PDF remained: entries=%d err=%v", len(dayEntries), err)
	}
}

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

func TestBugFB044_ProcessTemplateNormalizesStyleAndFontTable(t *testing.T) {
	styleXML := `<w:styles><w:style><w:rPr><w:rFonts w:ascii="微软雅黑" w:hAnsi="微软雅黑" w:eastAsia="宋体" w:cs="Times New Roman"/></w:rPr></w:style></w:styles>`
	fontTableXML := `<w:fonts><w:font w:name="微软雅黑"/><w:font w:name="宋体"/><w:font w:name="汉仪雅酷黑 75W"/></w:fonts>`

	styleOut := string(normalizeReportFontXml([]byte(styleXML)))
	fontTableOut := string(normalizeReportFontXml([]byte(fontTableXML)))

	for _, out := range []string{styleOut, fontTableOut} {
		if strings.Contains(out, "微软雅黑") || strings.Contains(out, "宋体") || strings.Contains(out, "汉仪雅酷黑") {
			t.Fatalf("expected style/font table fonts to be normalized, got %s", out)
		}
		if !strings.Contains(out, "Noto Sans CJK SC") {
			t.Fatalf("expected Noto Sans CJK SC in normalized XML, got %s", out)
		}
	}
}
