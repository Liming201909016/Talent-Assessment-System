package service

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

type Phase1V2FormalReportData struct {
	Report                 Phase1V2ReportData
	PaperID                string
	ExamID                 string
	ParticipantType        string
	ParticipantID          string
	ParticipantName        string
	ParticipantAge         *int
	ParticipantGender      string
	ParticipantTelephone   string
	ParticipantAffiliation string
	ParticipantPost        string
	RequiredFields         string
	SubmittedAt            *time.Time
	UserTime               int
	ExactScores            Phase1V2ScoreResult
	DimensionNormScores    []*big.Rat
}

func (s *CompetencyRuntimeService) FindPhase1V2FormalReportData(paperID string) (*Phase1V2FormalReportData, bool, error) {
	var run model.CompetencyResultRun
	err := s.db.Where("paper_id = ? AND scoring_version = ?", paperID, CompetencyPhase1ScoringVersionV2).Take(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if err := validatePhase1V2FormalRunHeader(run, paperID); err != nil {
		return nil, true, err
	}

	var overall model.CompetencyResultRunOverall
	if err := s.db.Where("result_run_id = ?", run.ID).Take(&overall).Error; err != nil {
		return nil, true, err
	}
	modules := make([]model.CompetencyResultRunModule, 0, 3)
	if err := s.db.Where("result_run_id = ?", run.ID).Order("display_order ASC").Find(&modules).Error; err != nil {
		return nil, true, err
	}
	dimensions := make([]model.CompetencyResultRunDimension, 0, 10)
	if err := s.db.Where("result_run_id = ?", run.ID).Order("display_order ASC").Find(&dimensions).Error; err != nil {
		return nil, true, err
	}
	var validity model.CompetencyResultRunValidity
	if err := s.db.Where("result_run_id = ?", run.ID).Take(&validity).Error; err != nil {
		return nil, true, err
	}
	scores, moduleScores, validityStatus, err := phase1V2ReportInputsFromRunRows(overall, modules, dimensions, validity)
	if err != nil {
		return nil, true, err
	}

	var contentPackage model.CompetencyReportContentPackage
	versions := Phase1V2VersionSet()
	if err := s.db.Where("product_version = ? AND scoring_version = ? AND content_version = ? AND template_version = ? AND audience = ?", versions.ProductVersion, versions.ScoringVersion, versions.ContentVersion, versions.ReportTemplateVersion, run.ReportAudience).Take(&contentPackage).Error; err != nil {
		return nil, true, ErrPhase1ReportContentNotApproved
	}
	if err := ValidatePhase1ReportContentApprovalForEnvironment(contentPackage, CompetencyReportEffectiveEnvironment()); err != nil {
		return nil, true, err
	}
	texts := make([]model.CompetencyReportText, 0)
	if err := s.db.Where("content_version = ? AND audience = ? AND status = 0", run.ContentVersion, run.ReportAudience).Find(&texts).Error; err != nil {
		return nil, true, err
	}
	report, err := BuildPhase1V2ReportData(run.ID, scores, moduleScores, validityStatus, texts)
	if err != nil {
		return nil, true, err
	}
	if report.Disclaimer != contentPackage.Disclaimer {
		return nil, true, errors.New("phase-1 v2 report disclaimer does not match approved content")
	}
	var paperMeta struct {
		ExamID         string `gorm:"column:exam_id"`
		UserID         string `gorm:"column:user_id"`
		RequiredFields string `gorm:"column:required_fields"`
	}
	if err := s.db.Table("el_paper p").
		Select("p.exam_id, p.user_id, COALESCE(e.required_fields, '') AS required_fields").
		Joins("INNER JOIN el_exam e ON e.id = p.exam_id").
		Where("p.id = ?", paperID).Take(&paperMeta).Error; err != nil {
		return nil, true, err
	}
	if err := validatePhase1V2RunPaperIdentity(run, paperMeta.ExamID, paperMeta.UserID); err != nil {
		return nil, true, err
	}
	return &Phase1V2FormalReportData{
		Report: report, PaperID: run.PaperID, ExamID: run.ExamID,
		ParticipantType: run.ParticipantType, ParticipantID: run.ParticipantID, ParticipantName: run.ParticipantName,
		ParticipantAge: run.ParticipantAge, ParticipantGender: run.ParticipantGender, ParticipantTelephone: run.ParticipantTelephone,
		ParticipantAffiliation: run.ParticipantAffiliation, ParticipantPost: run.ParticipantPost,
		RequiredFields: paperMeta.RequiredFields,
		SubmittedAt:    overall.SubmittedAt, UserTime: overall.UserTime, ExactScores: scores,
		DimensionNormScores: Phase1V2DimensionNormScores(),
	}, true, nil
}

func validatePhase1V2FormalRunHeader(run model.CompetencyResultRun, paperID string) error {
	if strings.TrimSpace(run.ID) == "" || run.PaperID != paperID || run.Status != phase1V2ResultRunStatusCompleted ||
		(CompetencyVersionSet{ProductVersion: run.ProductVersion, ScoringVersion: run.ScoringVersion, ContentVersion: run.ContentVersion, ReportTemplateVersion: run.ReportTemplateVersion}) != Phase1V2VersionSet() ||
		run.ReportAudience != CompetencyReportAudienceFrontlineEmployee || run.CompletedAt == nil || validatePhase1V2ResultRunSource(run.Source) != nil ||
		(run.ParticipantType != CompetencyParticipantCandidate && run.ParticipantType != CompetencyParticipantTester) ||
		strings.TrimSpace(run.ParticipantID) == "" || strings.TrimSpace(run.ParticipantName) == "" || strings.TrimSpace(run.ExamID) == "" {
		return errors.New("phase-1 v2 report result run is invalid")
	}
	return nil
}

func validatePhase1V2RunPaperIdentity(run model.CompetencyResultRun, paperExamID, paperUserID string) error {
	if strings.TrimSpace(paperExamID) == "" || strings.TrimSpace(paperUserID) == "" ||
		run.ExamID != paperExamID || run.ParticipantID != paperUserID {
		return errors.New("phase-1 v2 report result run identity does not match paper")
	}
	return nil
}

func phase1V2ReportInputsFromRunRows(overall model.CompetencyResultRunOverall, moduleRows []model.CompetencyResultRunModule, dimensionRows []model.CompetencyResultRunDimension, validity model.CompetencyResultRunValidity) (Phase1V2ScoreResult, []Phase1V2ModuleScore, string, error) {
	if overall.IsComplete != 1 || overall.TotalQuestionCount != 90 || overall.AnsweredQuestionCount != 90 ||
		overall.DimensionQuestionCount != 80 || overall.AnsweredDimensionQuestionCount != 80 ||
		overall.EffectiveDimensionCount != 10 || overall.OverallScore == nil || overall.LevelCode == nil ||
		overall.NormScore == nil || overall.NormComparisonCode == nil || overall.UserTime < 1 {
		return Phase1V2ScoreResult{}, nil, "", errors.New("phase-1 v2 persisted overall result is incomplete")
	}
	definitions := Phase1V2Dimensions()
	normScores := Phase1V2DimensionNormScores()
	if len(dimensionRows) != len(definitions) {
		return Phase1V2ScoreResult{}, nil, "", errors.New("phase-1 v2 persisted dimensions are incomplete")
	}
	sort.Slice(dimensionRows, func(i, j int) bool { return dimensionRows[i].DisplayOrder < dimensionRows[j].DisplayOrder })
	scores := Phase1V2ScoreResult{
		Dimensions: make([]Phase1V2DimensionScore, 0, len(definitions)), TotalQuestionCount: 80,
		AnsweredQuestionCount: 80, EffectiveDimensionCount: 10, IsComplete: true,
	}
	for index, definition := range definitions {
		row := dimensionRows[index]
		if row.ResultRunID != overall.ResultRunID || row.DimensionID != definition.ID || row.DimensionCode != definition.DisplayCode ||
			row.DimensionName != definition.Name || row.DisplayOrder != definition.DisplayOrder || row.TotalQuestionCount != 8 ||
			row.AnsweredQuestionCount != 8 || row.IsComplete != 1 || row.DimensionScore == nil || row.LevelCode == nil || row.NormScore == nil {
			return Phase1V2ScoreResult{}, nil, "", fmt.Errorf("phase-1 v2 persisted dimension is invalid: %s", definition.ID)
		}
		exact := new(big.Rat).Sub(new(big.Rat).Mul(big.NewRat(25, 1), big.NewRat(int64(row.ScoreSum), 8)), big.NewRat(25, 1))
		level, err := Phase1V2LevelForScore(exact)
		if err != nil || !row.DimensionScore.Equal(decimal.NewFromBigRat(exact, 6)) || *row.LevelCode != level ||
			!row.NormScore.Equal(decimal.NewFromBigRat(normScores[index], 6)) {
			return Phase1V2ScoreResult{}, nil, "", fmt.Errorf("phase-1 v2 persisted dimension score drift: %s", definition.ID)
		}
		scores.Dimensions = append(scores.Dimensions, Phase1V2DimensionScore{
			DimensionID: definition.ID, StableKey: definition.StableKey, DisplayCode: definition.DisplayCode,
			DimensionName: definition.Name, ModuleCode: definition.ModuleCode, DisplayOrder: definition.DisplayOrder,
			TotalQuestionCount: 8, AnsweredQuestionCount: 8, ScoreSum: row.ScoreSum, Score: exact, Level: level, IsComplete: true,
		})
	}
	total := new(big.Rat)
	for _, dimension := range scores.Dimensions {
		total.Add(total, dimension.Score)
	}
	scores.OverallScore = new(big.Rat).Quo(total, big.NewRat(10, 1))
	scores.OverallLevel, _ = Phase1V2LevelForScore(scores.OverallScore)
	overallComparison, err := Phase1V2OverallNormComparison(scores.OverallScore)
	if err != nil || !overall.OverallScore.Equal(decimal.NewFromBigRat(scores.OverallScore, 6)) ||
		*overall.LevelCode != scores.OverallLevel || !overall.NormScore.Equal(decimal.NewFromBigRat(overallComparison.NormScore, 6)) ||
		*overall.NormComparisonCode != overallComparison.ComparisonCode {
		return Phase1V2ScoreResult{}, nil, "", errors.New("phase-1 v2 persisted overall score drift")
	}

	modules, err := CalculatePhase1V2ModuleResults(scores.Dimensions)
	if err != nil || len(moduleRows) != len(modules) {
		return Phase1V2ScoreResult{}, nil, "", errors.New("phase-1 v2 persisted modules are incomplete")
	}
	sort.Slice(moduleRows, func(i, j int) bool { return moduleRows[i].DisplayOrder < moduleRows[j].DisplayOrder })
	for index, expected := range modules {
		row := moduleRows[index]
		if row.ResultRunID != overall.ResultRunID || row.ModuleID != expected.ModuleCode || row.ModuleCode != expected.ModuleCode || row.ModuleName != expected.ModuleName ||
			row.DisplayOrder != expected.DisplayOrder || row.TotalDimensionCount != expected.TotalDimensionCount ||
			row.EffectiveDimensionCount != expected.EffectiveDimensionCount || row.IsComplete != 1 || row.ModuleScore == nil ||
			row.LevelCode == nil || row.NormScore == nil || row.NormComparisonCode == nil ||
			!row.ModuleScore.Equal(decimal.NewFromBigRat(expected.Score, 6)) || *row.LevelCode != expected.Level ||
			!row.NormScore.Equal(decimal.NewFromBigRat(expected.NormScore, 6)) || *row.NormComparisonCode != expected.ComparisonCode {
			return Phase1V2ScoreResult{}, nil, "", fmt.Errorf("phase-1 v2 persisted module score drift: %s", expected.ModuleCode)
		}
	}
	if overall.SubmittedAt == nil || (overall.SubmitType != CompetencySubmitManual && overall.SubmitType != CompetencySubmitTimeout) {
		return Phase1V2ScoreResult{}, nil, "", errors.New("phase-1 v2 persisted submission metadata is invalid")
	}
	if validity.ResultRunID != overall.ResultRunID || validity.TotalQuestionCount != 10 || validity.AnsweredQuestionCount != 10 ||
		validity.IsComplete != 1 || validity.ValidityScore == nil || validity.ValidityStatus == nil ||
		(*validity.ValidityStatus != CompetencyPhase1ValidityGood && *validity.ValidityStatus != CompetencyPhase1ValidityQuestionable) {
		return Phase1V2ScoreResult{}, nil, "", errors.New("phase-1 v2 persisted validity result is invalid")
	}
	expectedValidityStatus := CompetencyPhase1ValidityGood
	if validity.ValidityScore.GreaterThan(decimal.NewFromInt(35)) {
		expectedValidityStatus = CompetencyPhase1ValidityQuestionable
	}
	if validity.ValidityScore.LessThan(decimal.NewFromInt(10)) || validity.ValidityScore.GreaterThan(decimal.NewFromInt(50)) || *validity.ValidityStatus != expectedValidityStatus {
		return Phase1V2ScoreResult{}, nil, "", errors.New("phase-1 v2 persisted validity score drift")
	}
	return scores, modules, *validity.ValidityStatus, nil
}
