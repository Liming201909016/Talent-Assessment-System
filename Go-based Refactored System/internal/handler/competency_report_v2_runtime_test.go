package handler

import (
	"bytes"
	"context"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/talent-assessment/refactored/internal/service"
)

func TestBugFB181_V2DTOBuildsExactWordPayload(t *testing.T) {
	dto := service.Phase1V2ReportData{
		SchemaVersion: "competency-phase1-report-data-v2", ResultRunID: "run-v2-1",
		Versions: service.Phase1V2VersionSet(), Audience: service.CompetencyReportAudienceFrontlineEmployee,
		Overall: service.Phase1V2ReportOverall{Score: "68.13", LevelCode: service.CompetencyPhase1V2LevelQualified, NormScore: "57.75", ComparisonCode: service.CompetencyPhase1V2ComparisonAboveNorm, AssessmentText: "总体评价", AdviceText: "总体建议"},
		Modules: []service.Phase1V2ReportModule{
			{ModuleCode: service.CompetencyPhase1V2ModuleSelf, ModuleName: "自我管理类", Score: "75.00", LevelCode: service.CompetencyPhase1V2LevelGood, ComparisonCode: service.CompetencyPhase1V2ComparisonStandout, ComparisonText: "优势突出"},
			{ModuleCode: service.CompetencyPhase1V2ModuleTask, ModuleName: "任务管理类", Score: "67.50", LevelCode: service.CompetencyPhase1V2LevelQualified, ComparisonCode: service.CompetencyPhase1V2ComparisonAboveNorm, ComparisonText: "略高于常模分"},
			{ModuleCode: service.CompetencyPhase1V2ModuleInterpersonal, ModuleName: "人际管理类", Score: "59.38", LevelCode: service.CompetencyPhase1V2LevelQualified, ComparisonCode: service.CompetencyPhase1V2ComparisonAboveNorm, ComparisonText: "略高于常模分"},
		},
		Validity:   service.Phase1V2ReportValidity{Status: service.CompetencyPhase1ValidityGood, DisplayText: "有效"},
		Disclaimer: "正式免责声明",
	}
	definitions := service.Phase1V2Dimensions()
	scores := []string{"68.75", "62.50", "81.25", "50.00", "75.00", "56.25", "62.50", "78.13", "78.13", "68.75"}
	exact := []float64{68.75, 62.5, 81.25, 50, 75, 56.25, 62.5, 78.125, 78.125, 68.75}
	for index, definition := range definitions {
		dto.Dimensions = append(dto.Dimensions, service.Phase1V2ReportDimension{DimensionID: definition.ID, StableKey: definition.StableKey, DisplayCode: definition.DisplayCode, DimensionName: definition.Name, ModuleCode: definition.ModuleCode, DisplayOrder: definition.DisplayOrder, Score: scores[index], LevelCode: service.CompetencyPhase1V2LevelQualified, PerformanceText: "表现-" + definition.StableKey})
	}
	dto.Strengths = []service.Phase1V2ReportSelectedDimension{{RuleText: "优势一"}, {RuleText: "优势二"}}
	dto.Developments = []service.Phase1V2ReportSelectedDimension{{RuleText: "待发展一"}}

	submittedAt := time.Date(2026, 9, 18, 12, 30, 0, 0, time.Local)
	formal := service.Phase1V2FormalReportData{
		Report: dto, PaperID: "paper-v2-1", ParticipantName: "张三", ParticipantAge: func() *int { value := 30; return &value }(),
		ParticipantGender: "0", ParticipantTelephone: "13800000000", ParticipantAffiliation: "示例单位", ParticipantPost: "示例岗位",
		SubmittedAt: &submittedAt, UserTime: 20,
		ExactScores:         service.Phase1V2ScoreResult{OverallScore: big.NewRat(545, 8), IsComplete: true},
		DimensionNormScores: service.Phase1V2DimensionNormScores(),
	}
	for index, definition := range definitions {
		formal.ExactScores.Dimensions = append(formal.ExactScores.Dimensions, service.Phase1V2DimensionScore{DimensionID: definition.ID, Score: new(big.Rat).SetFloat64(exact[index])})
	}
	fields, charts, err := buildPhase1V2WordTemplateData(formal)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 60 || len(charts) != 12 {
		t.Fatalf("payload fields/charts=%d/%d", len(fields), len(charts))
	}
	for key, want := range map[string]string{
		"participant.name": "张三", "participant.gender": "男", "result.userTime": "20",
		"overall.score": "68.13", "overall.level": "合格", "overall.normComparison": "优于常模分",
		"rule.moduleSummary": "自我管理类优势突出，任务管理类略高于常模分，人际管理类略高于常模分。",
		"rule.overallAdvice": "总体评价\n总体建议",
		"rule.strength.1":    "优势一", "rule.strength.3": "", "rule.development.2": "",
	} {
		if fields[key] != want {
			t.Errorf("field %s=%q want=%q", key, fields[key], want)
		}
	}
	if got := charts["chart.overall.score"][0]; len(got) != 2 || got[0] != 68.125 || got[1] != 31.875 {
		t.Fatalf("overall chart=%v", got)
	}
	if got := charts["chart.dimension.comparison"][0]; len(got) != 10 || got[7] != exact[7] {
		t.Fatalf("dimension chart scores=%v", got)
	}
	converter := &capturingPhase1Converter{pdf: append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("v2"), 1024)...)}
	renderer := &phase1WordReportRenderer{
		v2TemplatePath: "../../configs/export-templates/competency-phase1-report-v2.docx",
		converter:      converter, timeout: time.Second, calibrateLabels: true,
	}
	pdf, err := renderer.RenderPhase1V2(context.Background(), formal)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pdf, converter.pdf) || !strings.Contains(converter.fileName, "phase1-v2") || len(converter.docx) == 0 {
		t.Fatal("v2 renderer did not pass the value-only DOCX through the configured converter")
	}
	renderedDocument := string(readWordPart(t, converter.docx, "word/document.xml"))
	if !strings.Contains(renderedDocument, "68.13") || !strings.Contains(renderedDocument, "总体评价") || !strings.Contains(renderedDocument, "总体建议") {
		t.Fatal("v2 converted DOCX does not contain adapted report values")
	}
}

// TestBugFB190_V2WordReportHonorsProfileEmptyStateAndApprovedText
// 对应：docs/regression-tests.md #FB-190
func TestBugFB190_V2WordReportHonorsProfileEmptyStateAndApprovedText(t *testing.T) {
	formal := phase1V2WordFormalFixture(t)
	formal.RequiredFields = "name,gender,telephone"
	formal.UserTime = 1
	formal.Report.Strengths = make([]service.Phase1V2ReportSelectedDimension, 0)
	formal.Report.Validity.DisplayText = "本次测评作答效度良好，结果具有较好的参考价值。"
	formal.Report.Disclaimer = "批准免责声明"

	fields, _, err := buildPhase1V2WordTemplateData(formal)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"result.userTime": "1", "rule.strength.1": "暂无明显优势项",
		"validity.text": formal.Report.Validity.DisplayText, "report.disclaimer": formal.Report.Disclaimer,
	} {
		if fields[key] != want {
			t.Errorf("field %s=%q want=%q", key, fields[key], want)
		}
	}

	converter := &capturingPhase1Converter{pdf: []byte("%PDF-1.7\nformat")}
	renderer := &phase1WordReportRenderer{
		v2TemplatePath: "../../configs/export-templates/competency-phase1-report-v2.docx",
		converter:      converter, timeout: time.Second,
	}
	if _, err := renderer.RenderPhase1V2(context.Background(), formal); err != nil {
		t.Fatal(err)
	}
	if output := os.Getenv("PHASE1_V2_FB190_DOCX"); output != "" {
		if err := os.WriteFile(output, converter.docx, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	document := string(readWordPart(t, converter.docx, "word/document.xml"))
	for _, absent := range []string{"年龄：", "单位：", "岗位："} {
		if strings.Contains(document, absent) {
			t.Errorf("unconfigured profile label remains: %s", absent)
		}
	}
	for _, absent := range []string{"时长：", ">1<", "分钟"} {
		if strings.Contains(document, absent) {
			t.Errorf("rendered v2 document exposes hidden duration content %q", absent)
		}
	}
	for _, required := range []string{"暂无明显优势项", "批准免责声明"} {
		if !strings.Contains(document, required) {
			t.Errorf("rendered v2 document missing %q", required)
		}
	}
}

// TestBugFB193_V2OverviewIncludesDynamicDimensionNames
// Corresponds to docs/regression-tests.md FB-193.
func TestBugFB193_V2OverviewIncludesDynamicDimensionNames(t *testing.T) {
	formal := phase1V2WordFormalFixture(t)
	formal.Report.Strengths = []service.Phase1V2ReportSelectedDimension{{DimensionName: "数字应用", RuleText: "优势描述"}}
	formal.Report.Developments = []service.Phase1V2ReportSelectedDimension{{DimensionName: "沟通表达", RuleText: "待发展描述"}}
	fields, _, err := buildPhase1V2WordTemplateData(formal)
	if err != nil {
		t.Fatal(err)
	}
	if fields["rule.strength.1"] != "数字应用：优势描述" {
		t.Fatalf("strength=%q", fields["rule.strength.1"])
	}
	if fields["rule.development.1"] != "沟通表达：待发展描述" {
		t.Fatalf("development=%q", fields["rule.development.1"])
	}
}

func phase1V2WordFormalFixture(t *testing.T) service.Phase1V2FormalReportData {
	t.Helper()
	definitions := service.Phase1V2Dimensions()
	dto := service.Phase1V2ReportData{
		SchemaVersion: "competency-phase1-report-data-v2", ResultRunID: "run-v2-format",
		Versions: service.Phase1V2VersionSet(), Audience: service.CompetencyReportAudienceFrontlineEmployee,
		Overall: service.Phase1V2ReportOverall{Score: "37.81", LevelCode: service.CompetencyPhase1V2LevelQualified, ComparisonCode: service.CompetencyPhase1V2ComparisonBelowNorm,
			AssessmentText: strings.Repeat("工作表现符合岗位基本要求，能够独立完成日常工作中的常规任务。", 3),
			AdviceText:     "建议在人才决策与培养中，以其低分项（受测者的分数最低的两个维度）为重点关注方向，并结合面试、业绩等其他信息综合判断。"},
		Modules: []service.Phase1V2ReportModule{
			{ModuleCode: service.CompetencyPhase1V2ModuleTask, ModuleName: "任务管理类", Score: "34.38", LevelCode: service.CompetencyPhase1V2LevelQualified, ComparisonCode: service.CompetencyPhase1V2ComparisonBelowNorm, ComparisonText: "低于常模分"},
			{ModuleCode: service.CompetencyPhase1V2ModuleInterpersonal, ModuleName: "人际管理类", Score: "39.06", LevelCode: service.CompetencyPhase1V2LevelQualified, ComparisonCode: service.CompetencyPhase1V2ComparisonBelowNorm, ComparisonText: "低于常模分"},
			{ModuleCode: service.CompetencyPhase1V2ModuleSelf, ModuleName: "自我管理类", Score: "42.71", LevelCode: service.CompetencyPhase1V2LevelQualified, ComparisonCode: service.CompetencyPhase1V2ComparisonBelowNorm, ComparisonText: "低于常模分"},
		},
		Developments: []service.Phase1V2ReportSelectedDimension{{DimensionName: "成就导向", RuleText: strings.Repeat("目标意识需要提升，建议通过小目标积累正向体验。", 2)}, {DimensionName: "敬业奉献", RuleText: strings.Repeat("投入度需要提升，建议从按时完成本职任务开始。", 2)}},
		Strengths:    make([]service.Phase1V2ReportSelectedDimension, 0),
		Validity:     service.Phase1V2ReportValidity{Status: service.CompetencyPhase1ValidityGood},
	}
	formal := service.Phase1V2FormalReportData{
		Report: dto, PaperID: "paper-v2-format", ParticipantName: "张三", ParticipantGender: "0", ParticipantTelephone: "13800000000",
		SubmittedAt: func() *time.Time { value := time.Date(2026, 9, 19, 12, 0, 0, 0, time.Local); return &value }(), UserTime: 1,
		ExactScores:         service.Phase1V2ScoreResult{OverallScore: big.NewRat(605, 16), IsComplete: true},
		DimensionNormScores: service.Phase1V2DimensionNormScores(),
	}
	for _, definition := range definitions {
		dto.Dimensions = append(dto.Dimensions, service.Phase1V2ReportDimension{DimensionID: definition.ID, StableKey: definition.StableKey, DisplayCode: definition.DisplayCode, DimensionName: definition.Name, ModuleCode: definition.ModuleCode, DisplayOrder: definition.DisplayOrder, Score: "37.50", LevelCode: service.CompetencyPhase1V2LevelQualified, PerformanceText: strings.Repeat("面对常规问题能够完成任务，建议结合实际工作持续练习并及时复盘改进。", 4)})
		formal.ExactScores.Dimensions = append(formal.ExactScores.Dimensions, service.Phase1V2DimensionScore{DimensionID: definition.ID, Score: big.NewRat(75, 2)})
	}
	formal.Report = dto
	return formal
}

func TestBugFB181_ReportRuntimeUsesRunBindingAndCurrentPointer(t *testing.T) {
	source := readSourceFile(t, "competency_report.go")
	generate := extractFunctionBody(t, source, "func (h *CompetencyReportHandler) Generate(c *gin.Context) {")
	for _, fragment := range []string{"Phase1V2FormalReportData", "result_run_id", "CompetencyReportCurrent", "RenderPhase1V2", "preserveCompletedV2", "upsertCompetencyReportCurrent(tx", "competencyReportBindingSchemaReady", `Omit("ResultRunID")`} {
		if !strings.Contains(generate+source, fragment) {
			t.Errorf("v2 report runtime missing %q", fragment)
		}
	}
	render := extractFunctionBody(t, source, "func (h *CompetencyReportHandler) renderCompetencyReport(")
	if !strings.Contains(render, `if v2Data != nil`) || !strings.Contains(render, `return nil, errors.New("v2 Word报告渲染器未启用")`) {
		t.Fatal("v2 report may fall through to the legacy Chromium renderer")
	}
	if !strings.Contains(source, "func (h *CompetencyReportHandler) Download(c *gin.Context)") || !strings.Contains(source, "CompetencyReportCurrent") {
		t.Fatal("single report download does not resolve the current report pointer")
	}
}

func TestBugFB183_DownloadIsSerializedWithRegeneration(t *testing.T) {
	download := extractFunctionBody(t, readSourceFile(t, "competency_report.go"), "func (h *CompetencyReportHandler) Download(c *gin.Context) {")
	if !strings.Contains(download, "unlock := h.lockReportGeneration(paperID)") || !strings.Contains(download, "defer unlock()") {
		t.Fatal("single download can race with replacement and deletion of the same paper's PDF")
	}
}
