package service

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/talent-assessment/refactored/internal/model"
)

const (
	CompetencyReportContentV2ModuleComparison = "module_comparison"
	CompetencyReportContentV2OverallAdvice    = "overall_advice"
	CompetencyReportContentV2Strength         = "strength"
	CompetencyReportContentV2Development      = "development"
)

type Phase1V2ReportOverall struct {
	Score          string `json:"score"`
	LevelCode      string `json:"levelCode"`
	NormScore      string `json:"normScore"`
	ComparisonCode string `json:"comparisonCode"`
	AssessmentText string `json:"assessmentText"`
	AdviceText     string `json:"adviceText"`
}

type Phase1V2ReportModule struct {
	ModuleCode     string `json:"moduleCode"`
	ModuleName     string `json:"moduleName"`
	DisplayOrder   int    `json:"displayOrder"`
	Score          string `json:"score"`
	LevelCode      string `json:"levelCode"`
	NormScore      string `json:"normScore"`
	ComparisonCode string `json:"comparisonCode"`
	ComparisonText string `json:"comparisonText"`
}

type Phase1V2ReportDimension struct {
	DimensionID     string `json:"dimensionId"`
	StableKey       string `json:"stableKey"`
	DisplayCode     string `json:"displayCode"`
	DimensionName   string `json:"dimensionName"`
	ModuleCode      string `json:"moduleCode"`
	DisplayOrder    int    `json:"displayOrder"`
	Score           string `json:"score"`
	LevelCode       string `json:"levelCode"`
	PerformanceText string `json:"performanceText"`
}

type Phase1V2ReportSelectedDimension struct {
	DimensionID   string `json:"dimensionId"`
	StableKey     string `json:"stableKey"`
	DisplayCode   string `json:"displayCode"`
	DimensionName string `json:"dimensionName"`
	DisplayOrder  int    `json:"displayOrder"`
	Score         string `json:"score"`
	LevelCode     string `json:"levelCode"`
	RuleText      string `json:"ruleText"`
}

type Phase1V2ReportValidity struct {
	Status      string `json:"status"`
	DisplayText string `json:"displayText"`
}

type Phase1V2ReportData struct {
	SchemaVersion    string                            `json:"schemaVersion"`
	ResultRunID      string                            `json:"resultRunId"`
	Versions         CompetencyVersionSet              `json:"versions"`
	Audience         string                            `json:"audience"`
	Overall          Phase1V2ReportOverall             `json:"overall"`
	Modules          []Phase1V2ReportModule            `json:"modules"`
	Dimensions       []Phase1V2ReportDimension         `json:"dimensions"`
	Strengths        []Phase1V2ReportSelectedDimension `json:"strengths"`
	Developments     []Phase1V2ReportSelectedDimension `json:"developments"`
	AdviceDimensions []Phase1V2ReportSelectedDimension `json:"adviceDimensions"`
	Validity         Phase1V2ReportValidity            `json:"validity"`
	Disclaimer       string                            `json:"disclaimer"`
}

type phase1V2ReportTextKey struct {
	contentType string
	identity    string
	condition   string
}

func BuildPhase1V2ReportData(resultRunID string, scores Phase1V2ScoreResult, modules []Phase1V2ModuleScore, validityStatus string, rows []model.CompetencyReportText) (Phase1V2ReportData, error) {
	if strings.TrimSpace(resultRunID) == "" {
		return Phase1V2ReportData{}, errors.New("phase-1 v2 report result run id is required")
	}
	if err := validatePhase1V2ReportScores(scores); err != nil {
		return Phase1V2ReportData{}, err
	}
	if validityStatus != CompetencyPhase1ValidityGood && validityStatus != CompetencyPhase1ValidityQuestionable {
		return Phase1V2ReportData{}, errors.New("phase-1 v2 report validity status is invalid")
	}
	overview, err := SelectPhase1V2ReportOverview(scores.Dimensions, modules)
	if err != nil {
		return Phase1V2ReportData{}, err
	}
	adviceDimensions, err := SelectPhase1V2OverallAdviceDimensions(scores.Dimensions, scores.OverallLevel)
	if err != nil {
		return Phase1V2ReportData{}, err
	}
	overallComparison, err := Phase1V2OverallNormComparison(scores.OverallScore)
	if err != nil {
		return Phase1V2ReportData{}, err
	}
	texts, disclaimer, err := indexPhase1V2ReportTexts(rows)
	if err != nil {
		return Phase1V2ReportData{}, err
	}
	lookup := func(contentType, identity, condition string) (string, error) {
		key := phase1V2ReportTextKey{contentType: contentType, identity: identity, condition: condition}
		content, exists := texts[key]
		if !exists {
			return "", fmt.Errorf("phase-1 v2 report text missing: contentVersion=%s, audience=%s, contentType=%s, identity=%s, condition=%s", CompetencyPhase1ContentVersionV2, CompetencyReportAudienceFrontlineEmployee, contentType, identity, condition)
		}
		return content, nil
	}

	overallText, err := lookup(CompetencyReportContentOverall, "", scores.OverallLevel)
	if err != nil {
		return Phase1V2ReportData{}, err
	}
	adviceText, err := lookup(CompetencyReportContentV2OverallAdvice, "", scores.OverallLevel)
	if err != nil {
		return Phase1V2ReportData{}, err
	}
	validityText, err := lookup(CompetencyReportContentValidity, "", validityStatus)
	if err != nil {
		return Phase1V2ReportData{}, err
	}

	moduleViews := make([]Phase1V2ReportModule, 0, len(overview.Modules))
	for _, module := range overview.Modules {
		text, lookupErr := lookup(CompetencyReportContentV2ModuleComparison, module.ModuleCode, module.ComparisonCode)
		if lookupErr != nil {
			return Phase1V2ReportData{}, lookupErr
		}
		moduleViews = append(moduleViews, Phase1V2ReportModule{
			ModuleCode: module.ModuleCode, ModuleName: module.ModuleName, DisplayOrder: module.DisplayOrder,
			Score: formatPhase1V2ReportScore(module.Score), LevelCode: module.Level,
			NormScore: formatPhase1V2ReportScore(module.NormScore), ComparisonCode: module.ComparisonCode, ComparisonText: text,
		})
	}

	dimensionViews := make([]Phase1V2ReportDimension, 0, len(scores.Dimensions))
	for _, dimension := range scores.Dimensions {
		text, lookupErr := lookup(CompetencyReportContentDimension, dimension.DimensionID, dimension.Level)
		if lookupErr != nil {
			return Phase1V2ReportData{}, lookupErr
		}
		dimensionViews = append(dimensionViews, Phase1V2ReportDimension{
			DimensionID: dimension.DimensionID, StableKey: dimension.StableKey, DisplayCode: dimension.DisplayCode,
			DimensionName: dimension.DimensionName, ModuleCode: dimension.ModuleCode, DisplayOrder: dimension.DisplayOrder,
			Score: formatPhase1V2ReportScore(dimension.Score), LevelCode: dimension.Level, PerformanceText: text,
		})
	}
	strengthViews, err := buildPhase1V2SelectedDimensionViews(overview.Strengths, CompetencyReportContentDimension, lookup)
	if err != nil {
		return Phase1V2ReportData{}, err
	}
	developmentViews, err := buildPhase1V2SelectedDimensionViews(overview.Developments, CompetencyReportContentV2Development, lookup)
	if err != nil {
		return Phase1V2ReportData{}, err
	}
	adviceViews := make([]Phase1V2ReportSelectedDimension, 0, len(adviceDimensions))
	for _, dimension := range adviceDimensions {
		adviceViews = append(adviceViews, phase1V2SelectedDimensionView(dimension, ""))
	}

	return Phase1V2ReportData{
		SchemaVersion: "competency-phase1-report-data-v2", ResultRunID: resultRunID,
		Versions: Phase1V2VersionSet(), Audience: CompetencyReportAudienceFrontlineEmployee,
		Overall: Phase1V2ReportOverall{
			Score: formatPhase1V2ReportScore(scores.OverallScore), LevelCode: scores.OverallLevel,
			NormScore: formatPhase1V2ReportScore(overallComparison.NormScore), ComparisonCode: overallComparison.ComparisonCode,
			AssessmentText: overallText, AdviceText: adviceText,
		},
		Modules: moduleViews, Dimensions: dimensionViews, Strengths: strengthViews, Developments: developmentViews,
		AdviceDimensions: adviceViews, Validity: Phase1V2ReportValidity{Status: validityStatus, DisplayText: validityText}, Disclaimer: disclaimer,
	}, nil
}

func validatePhase1V2ReportScores(scores Phase1V2ScoreResult) error {
	if !scores.IsComplete || scores.TotalQuestionCount != 80 || scores.AnsweredQuestionCount != 80 ||
		scores.EffectiveDimensionCount != 10 || scores.OverallScore == nil {
		return errors.New("phase-1 v2 report scores are incomplete")
	}
	if err := validatePhase1V2SelectorDimensions(scores.Dimensions); err != nil {
		return err
	}
	total := new(big.Rat)
	for _, dimension := range scores.Dimensions {
		total.Add(total, dimension.Score)
	}
	expectedOverall := new(big.Rat).Quo(total, big.NewRat(10, 1))
	level, err := Phase1V2LevelForScore(scores.OverallScore)
	if err != nil || scores.OverallScore.Cmp(expectedOverall) != 0 || scores.OverallLevel != level {
		return errors.New("phase-1 v2 report overall score and level mismatch")
	}
	return nil
}

func indexPhase1V2ReportTexts(rows []model.CompetencyReportText) (map[phase1V2ReportTextKey]string, string, error) {
	texts := make(map[phase1V2ReportTextKey]string)
	disclaimer := ""
	for _, row := range rows {
		if row.ContentVersion != CompetencyPhase1ContentVersionV2 || row.Audience != CompetencyReportAudienceFrontlineEmployee || row.Status != 0 {
			continue
		}
		if row.IsTemporary != 0 {
			return nil, "", errors.New("phase-1 v2 report cannot use temporary text")
		}
		content := strings.TrimSpace(row.Content)
		if content == "" {
			return nil, "", errors.New("phase-1 v2 report text is blank")
		}
		rowDisclaimer := strings.TrimSpace(row.Disclaimer)
		if rowDisclaimer == "" {
			return nil, "", errors.New("phase-1 v2 report disclaimer is blank")
		}
		if disclaimer != "" && disclaimer != rowDisclaimer {
			return nil, "", errors.New("phase-1 v2 report disclaimer mismatch")
		}
		disclaimer = rowDisclaimer
		key := phase1V2ReportTextKey{contentType: row.ContentType, identity: row.DimensionID, condition: row.LevelCode}
		if _, exists := texts[key]; exists {
			return nil, "", fmt.Errorf("phase-1 v2 report text duplicate: contentType=%s, identity=%s, condition=%s", row.ContentType, row.DimensionID, row.LevelCode)
		}
		texts[key] = content
	}
	if disclaimer == "" {
		return nil, "", errors.New("phase-1 v2 report disclaimer is blank")
	}
	return texts, disclaimer, nil
}

func buildPhase1V2SelectedDimensionViews(dimensions []Phase1V2DimensionScore, contentType string, lookup func(string, string, string) (string, error)) ([]Phase1V2ReportSelectedDimension, error) {
	views := make([]Phase1V2ReportSelectedDimension, 0, len(dimensions))
	for _, dimension := range dimensions {
		text, err := lookup(contentType, dimension.DimensionID, dimension.Level)
		if err != nil {
			return nil, err
		}
		views = append(views, phase1V2SelectedDimensionView(dimension, text))
	}
	return views, nil
}

func phase1V2SelectedDimensionView(dimension Phase1V2DimensionScore, ruleText string) Phase1V2ReportSelectedDimension {
	return Phase1V2ReportSelectedDimension{
		DimensionID: dimension.DimensionID, StableKey: dimension.StableKey, DisplayCode: dimension.DisplayCode,
		DimensionName: dimension.DimensionName, DisplayOrder: dimension.DisplayOrder,
		Score: formatPhase1V2ReportScore(dimension.Score), LevelCode: dimension.Level, RuleText: ruleText,
	}
}

func formatPhase1V2ReportScore(score *big.Rat) string {
	if score == nil {
		return ""
	}
	return score.FloatString(2)
}
