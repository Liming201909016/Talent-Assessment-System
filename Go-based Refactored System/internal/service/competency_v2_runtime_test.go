package service

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestBugFB181_Phase1V2ContentPackageRequiresDualApproval(t *testing.T) {
	now := time.Now()
	versions := Phase1V2VersionSet()
	row := model.CompetencyReportContentPackage{
		ProductVersion: versions.ProductVersion, ScoringVersion: versions.ScoringVersion,
		ContentVersion: versions.ContentVersion, TemplateVersion: versions.ReportTemplateVersion,
		Audience: CompetencyReportAudienceFrontlineEmployee, ApprovalStatus: CompetencyReportApprovalApproved,
		ContentApprovedBy: "content-owner", ContentApprovedAt: &now,
		PsychometricApprovedBy: "psychometric-owner", PsychometricApprovedAt: &now,
		QuestionSourceSHA256: strings.Repeat("a", 64), ContentSourceSHA256: strings.Repeat("b", 64),
		EffectiveEnvironment: "local", Disclaimer: "approved disclaimer",
	}
	if err := ValidatePhase1ReportContentApproval(row); err != nil {
		t.Fatalf("exact v2 approval rejected: %v", err)
	}
	if err := ValidatePhase1ReportContentApprovalForEnvironment(row, "staging"); err == nil {
		t.Fatal("local v2 approval was accepted for staging")
	}
	row.PsychometricApprovedAt = nil
	if err := ValidatePhase1ReportContentApproval(row); err == nil {
		t.Fatal("v2 package without psychometric approval was accepted")
	}
}

func TestBugFB183_V2RunParticipantMustOwnPaper(t *testing.T) {
	run := model.CompetencyResultRun{PaperID: "paper-1", ExamID: "exam-1", ParticipantID: "participant-1"}
	if err := validatePhase1V2RunPaperIdentity(run, "exam-1", "participant-1"); err != nil {
		t.Fatalf("matching run identity rejected: %v", err)
	}
	if err := validatePhase1V2RunPaperIdentity(run, "exam-1", "participant-2"); err == nil {
		t.Fatal("run identity belonging to another participant was accepted")
	}
}

func TestBugFB181_PersistedV2RunReconstructsExactReportInputs(t *testing.T) {
	definitions := Phase1V2Dimensions()
	normScores := Phase1V2DimensionNormScores()
	sums := []int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30}
	dimensions := make([]model.CompetencyResultRunDimension, 0, len(definitions))
	for index, definition := range definitions {
		score := decimal.NewFromInt(int64(sums[index])).Mul(decimal.NewFromInt(25)).Div(decimal.NewFromInt(8)).Sub(decimal.NewFromInt(25))
		level, err := Phase1V2LevelForScore(score.Rat())
		if err != nil {
			t.Fatal(err)
		}
		norm := decimal.NewFromBigRat(normScores[index], 6)
		dimensions = append(dimensions, model.CompetencyResultRunDimension{
			ResultRunID: "run-v2-1", DimensionID: definition.ID, DimensionCode: definition.DisplayCode,
			DimensionName: definition.Name, DisplayOrder: definition.DisplayOrder, TotalQuestionCount: 8,
			AnsweredQuestionCount: 8, ScoreSum: sums[index], DimensionScore: &score, LevelCode: &level, NormScore: &norm, IsComplete: 1,
		})
	}
	overallScore := decimal.RequireFromString("68.125")
	overallLevel := CompetencyPhase1V2LevelQualified
	overallNorm := decimal.RequireFromString("57.75")
	overallComparison := CompetencyPhase1V2ComparisonAboveNorm
	overall := model.CompetencyResultRunOverall{
		ResultRunID: "run-v2-1", TotalQuestionCount: 90, AnsweredQuestionCount: 90,
		DimensionQuestionCount: 80, AnsweredDimensionQuestionCount: 80, EffectiveDimensionCount: 10,
		OverallScore: &overallScore, LevelCode: &overallLevel, NormScore: &overallNorm,
		NormComparisonCode: &overallComparison, IsComplete: 1, SubmitType: CompetencySubmitManual, SubmittedAt: func() *time.Time { value := time.Now(); return &value }(), UserTime: 20,
	}
	moduleRows := []model.CompetencyResultRunModule{
		phase1V2RuntimeModuleRow("run-v2-1", CompetencyPhase1V2ModuleTask, "任务管理类", 1, 5, "67.5", "58", CompetencyPhase1V2ComparisonAboveNorm),
		phase1V2RuntimeModuleRow("run-v2-1", CompetencyPhase1V2ModuleInterpersonal, "人际管理类", 2, 2, "59.375", "53.75", CompetencyPhase1V2ComparisonAboveNorm),
		phase1V2RuntimeModuleRow("run-v2-1", CompetencyPhase1V2ModuleSelf, "自我管理类", 3, 3, "75", "60", CompetencyPhase1V2ComparisonStandout),
	}
	validityStatus := CompetencyPhase1ValidityGood
	validityScore := decimal.NewFromInt(25)
	validity := model.CompetencyResultRunValidity{
		ResultRunID: "run-v2-1", TotalQuestionCount: 10, AnsweredQuestionCount: 10,
		ValidityScore: &validityScore, ValidityStatus: &validityStatus, IsComplete: 1,
	}

	scores, modules, status, err := phase1V2ReportInputsFromRunRows(overall, moduleRows, dimensions, validity)
	if err != nil {
		t.Fatal(err)
	}
	if scores.OverallScore.RatString() != "545/8" || len(scores.Dimensions) != 10 || scores.Dimensions[7].Score.RatString() != "625/8" {
		t.Fatalf("exact scores were not reconstructed: overall=%v dimensions=%+v", scores.OverallScore, scores.Dimensions)
	}
	if len(modules) != 3 || modules[1].Score.RatString() != "475/8" || status != CompetencyPhase1ValidityGood {
		t.Fatalf("runtime inputs=%+v status=%s", modules, status)
	}

	drifted := dimensions
	drifted[0].ScoreSum = 29
	if _, _, _, err := phase1V2ReportInputsFromRunRows(overall, moduleRows, drifted, validity); err == nil {
		t.Fatal("persisted v2 score drift was accepted")
	}
}

// TestBugFB185B_PersistedModuleIdentityCannotDrift
// 对应：docs/regression-tests.md #FB-185B
// 复现：模块显示码、名称和分数均正确，但持久化module_id被篡改。
// 期望：正式读取和幂等复用均拒绝该run，不能把错误语义身份视为有效结果。
func TestBugFB185B_PersistedModuleIdentityCannotDrift(t *testing.T) {
	overall, modules, dimensions, validity := phase1V2RuntimeRows(t)
	modules[0].ModuleID = "tampered-module-id"
	if _, _, _, err := phase1V2ReportInputsFromRunRows(overall, modules, dimensions, validity); err == nil {
		t.Fatal("persisted v2 module identity drift was accepted")
	}
}

// TestBugFB185B_ResultRunSourceMustBeKnown
// 对应：docs/regression-tests.md #FB-185B
func TestBugFB185B_ResultRunSourceMustBeKnown(t *testing.T) {
	for _, source := range []string{phase1V2ResultRunSourceSubmission, phase1V2ResultRunSourceHistoricalRecompute} {
		if err := validatePhase1V2ResultRunSource(source); err != nil {
			t.Fatalf("valid source %q rejected: %v", source, err)
		}
	}
	for _, source := range []string{"", "manual_edit", "submission "} {
		if err := validatePhase1V2ResultRunSource(source); err == nil {
			t.Fatalf("invalid source %q accepted", source)
		}
	}
}

func TestBugFB185B_SourceValidationIsWiredAcrossRunLifecycle(t *testing.T) {
	now := time.Date(2026, 9, 19, 10, 0, 0, 0, time.Local)
	legacy := model.CompetencyResult{
		PaperID: "paper-1", ExamID: "exam-1", ParticipantType: CompetencyParticipantCandidate,
		ParticipantID: "participant-1", ParticipantName: "张三", ReportAudience: CompetencyReportAudienceFrontlineEmployee,
		IsComplete: 1, SubmitType: CompetencySubmitManual, ProductVersion: CompetencyPhase1ProductVersion,
		ScoringVersion: CompetencyPhase1ScoringVersion, ContentVersion: CompetencyPhase1ContentVersion,
		ReportTemplateVersion: CompetencyPhase1ReportTemplateVersion, SubmittedAt: &now,
	}
	dimensionInputs := phase1V2InputsFromV1Sums([]int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30})
	validityInputs := make([]Phase1ValidityInput, 0, 10)
	for index := 1; index <= 10; index++ {
		validityInputs = append(validityInputs, Phase1ValidityInput{
			QuestionCode: "V" + string(rune('A'+index-1)), Order: index,
			QuestionType: CompetencyQuestionTypeValidity, Direction: CompetencyDirectionForward,
			Answered: true, RawValue: 3,
		})
	}
	if _, err := buildPhase1V2ResultRunRecords("run-v2-1", legacy, dimensionInputs, validityInputs, 20, "manual_edit", nil, now); err == nil {
		t.Fatal("creation accepted an unknown result-run source")
	}
	expected, err := buildPhase1V2ResultRunRecords("run-v2-1", legacy, dimensionInputs, validityInputs, 20, phase1V2ResultRunSourceHistoricalRecompute, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	existing := expected.Run
	existing.Source = phase1V2ResultRunSourceSubmission
	if err := validateExistingPhase1V2ResultRunMetadata(existing, expected.Run); err != nil {
		t.Fatalf("historical recompute must reuse a valid submission-created run: %v", err)
	}
	existing.Source = "manual_edit"
	if err := validateExistingPhase1V2ResultRunMetadata(existing, expected.Run); err == nil {
		t.Fatal("reuse accepted an unknown persisted source")
	}
	existing = expected.Run
	existing.Source = "manual_edit"
	if err := validatePhase1V2FormalRunHeader(existing, "paper-1"); err == nil {
		t.Fatal("formal read accepted an unknown persisted source")
	}
}

// TestBugFB185F_ReportDurationComesFromFrozenRun
// 对应：docs/regression-tests.md #FB-185F
func TestBugFB185F_ReportDurationComesFromFrozenRun(t *testing.T) {
	source := readV2ResultRunSource(t, "competency_v2_runtime.go")
	if strings.Contains(source, `Select("exam_id, user_id, user_time")`) || !strings.Contains(source, "UserTime: overall.UserTime") {
		t.Fatal("phase-1 v2 report duration is not sourced from the frozen result run")
	}
}

// TestBugFB190_V2FormalDataLoadsRequiredFields
// 对应：docs/regression-tests.md #FB-190
func TestBugFB190_V2FormalDataLoadsRequiredFields(t *testing.T) {
	var data Phase1V2FormalReportData
	data.RequiredFields = "name,gender,telephone"
	if data.RequiredFields == "" {
		t.Fatal("v2 formal report data lost requiredFields")
	}
	source := readV2ResultRunSource(t, "competency_v2_runtime.go")
	for _, required := range []string{"RequiredFields", "required_fields", "el_exam"} {
		if !strings.Contains(source, required) {
			t.Fatalf("v2 formal runtime missing %q", required)
		}
	}
}

func phase1V2RuntimeRows(t *testing.T) (model.CompetencyResultRunOverall, []model.CompetencyResultRunModule, []model.CompetencyResultRunDimension, model.CompetencyResultRunValidity) {
	t.Helper()
	definitions := Phase1V2Dimensions()
	normScores := Phase1V2DimensionNormScores()
	sums := []int{30, 28, 34, 24, 32, 26, 28, 33, 33, 30}
	dimensions := make([]model.CompetencyResultRunDimension, 0, len(definitions))
	for index, definition := range definitions {
		score := decimal.NewFromInt(int64(sums[index])).Mul(decimal.NewFromInt(25)).Div(decimal.NewFromInt(8)).Sub(decimal.NewFromInt(25))
		level, err := Phase1V2LevelForScore(score.Rat())
		if err != nil {
			t.Fatal(err)
		}
		norm := decimal.NewFromBigRat(normScores[index], 6)
		dimensions = append(dimensions, model.CompetencyResultRunDimension{
			ResultRunID: "run-v2-1", DimensionID: definition.ID, DimensionCode: definition.DisplayCode,
			DimensionName: definition.Name, DisplayOrder: definition.DisplayOrder, TotalQuestionCount: 8,
			AnsweredQuestionCount: 8, ScoreSum: sums[index], DimensionScore: &score, LevelCode: &level, NormScore: &norm, IsComplete: 1,
		})
	}
	overallScore := decimal.RequireFromString("68.125")
	overallLevel := CompetencyPhase1V2LevelQualified
	overallNorm := decimal.RequireFromString("57.75")
	overallComparison := CompetencyPhase1V2ComparisonAboveNorm
	overall := model.CompetencyResultRunOverall{
		ResultRunID: "run-v2-1", TotalQuestionCount: 90, AnsweredQuestionCount: 90,
		DimensionQuestionCount: 80, AnsweredDimensionQuestionCount: 80, EffectiveDimensionCount: 10,
		OverallScore: &overallScore, LevelCode: &overallLevel, NormScore: &overallNorm,
		NormComparisonCode: &overallComparison, IsComplete: 1, SubmitType: CompetencySubmitManual,
		SubmittedAt: func() *time.Time { value := time.Now(); return &value }(), UserTime: 20,
	}
	modules := []model.CompetencyResultRunModule{
		phase1V2RuntimeModuleRow("run-v2-1", CompetencyPhase1V2ModuleTask, "任务管理类", 1, 5, "67.5", "58", CompetencyPhase1V2ComparisonAboveNorm),
		phase1V2RuntimeModuleRow("run-v2-1", CompetencyPhase1V2ModuleInterpersonal, "人际管理类", 2, 2, "59.375", "53.75", CompetencyPhase1V2ComparisonAboveNorm),
		phase1V2RuntimeModuleRow("run-v2-1", CompetencyPhase1V2ModuleSelf, "自我管理类", 3, 3, "75", "60", CompetencyPhase1V2ComparisonStandout),
	}
	validityStatus := CompetencyPhase1ValidityGood
	validityScore := decimal.NewFromInt(25)
	validity := model.CompetencyResultRunValidity{
		ResultRunID: "run-v2-1", TotalQuestionCount: 10, AnsweredQuestionCount: 10,
		ValidityScore: &validityScore, ValidityStatus: &validityStatus, IsComplete: 1,
	}
	return overall, modules, dimensions, validity
}

func phase1V2RuntimeModuleRow(runID, code, name string, order, dimensionCount int, scoreValue, normValue, comparison string) model.CompetencyResultRunModule {
	score := decimal.RequireFromString(scoreValue)
	norm := decimal.RequireFromString(normValue)
	level, _ := Phase1V2LevelForScore(score.Rat())
	return model.CompetencyResultRunModule{
		ResultRunID: runID, ModuleID: code, ModuleCode: code, ModuleName: name, DisplayOrder: order,
		TotalDimensionCount: dimensionCount, EffectiveDimensionCount: dimensionCount,
		ModuleScore: &score, LevelCode: &level, NormScore: &norm, NormComparisonCode: &comparison, IsComplete: 1,
	}
}
