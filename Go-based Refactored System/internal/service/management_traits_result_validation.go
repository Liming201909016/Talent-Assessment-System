package service

import (
	"errors"
	"math/big"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
)

var errManagementTraitsStoredResult = errors.New("invalid management traits stored result")

// ValidateManagementTraitsStoredResult checks only caller-supplied snapshot/cache
// consistency and returns fresh, exact S1 results. The literal completed status
// does not prove actual submission, DB ownership, historical evidence, approved
// versions or authorization. Source, CreatedAt, SubmittedAt and UserTimeSeconds
// are deliberately deferred lifecycle metadata. No runtime is activated, no
// records are loaded, repaired or saved, and there is no legacy fallback.
func ValidateManagementTraitsStoredResult(bundle model.ManagementTraitsDefinitionBundle, paper model.ManagementTraitsPaperSnapshot, questions []model.ManagementTraitsPaperQuestionSnapshot, run model.ManagementTraitsResultRun, dimensions []model.ManagementTraitsResultDimension, modules []model.ManagementTraitsResultModule, maxJSONBytes int) (ManagementTraitsResult, error) {
	if run.Status != "completed" || len(dimensions) != 13 || len(modules) != 4 {
		return ManagementTraitsResult{}, errManagementTraitsStoredResult
	}
	validated, err := ValidateManagementTraitsStoredInput(bundle, paper, questions, maxJSONBytes)
	if err != nil {
		return ManagementTraitsResult{}, errManagementTraitsStoredResult
	}
	result, err := CalculateManagementTraits(bundle.Questionnaire, validated.Answers)
	if err != nil || !result.IsComplete || result.TotalQuestionCount != 140 || result.AnsweredQuestionCount != 140 {
		return ManagementTraitsResult{}, errManagementTraitsStoredResult
	}
	if !managementTraitsOpaqueID(run.ID) || run.PaperID != paper.PaperID || run.ExamID != paper.ExamID || run.ParticipantType != paper.ParticipantType || run.ParticipantID != paper.ParticipantID ||
		run.ProductVersion != bundle.ProductVersion || run.QuestionVersion != bundle.QuestionVersion || run.ScoringVersion != bundle.ScoringVersion || run.NormVersion != bundle.NormVersion || run.Questionnaire != bundle.Questionnaire ||
		run.ScoringManifestSHA != bundle.ScoringManifestSHA || run.ScoringManifestSHA != paper.ScoringManifestSHA || run.InputSHA != validated.Canonical.SHA256 || run.TotalQuestionCount != 140 || run.AnsweredQuestionCount != 140 {
		return ManagementTraitsResult{}, errManagementTraitsStoredResult
	}
	// Compare numeric values to six-place caches derived from exact answer-based
	// rationals. Equivalent decimal exponents are allowed; extra numeric precision
	// is not rounded away. Neither cache values nor stored sums feed calculation.
	cacheMatches := func(stored *decimal.Decimal, exact *big.Rat) bool {
		return stored != nil && exact != nil && stored.Equal(*decimalFromRat(exact))
	}
	if !cacheMatches(run.OverallScore, result.OverallScore) || !cacheMatches(run.OverallNorm, result.OverallNorm) || run.OverallLevel == nil || *run.OverallLevel != result.OverallLevel {
		return ManagementTraitsResult{}, errManagementTraitsStoredResult
	}
	expectedDimensions := make(map[string]ManagementTraitsDimensionScore, 13)
	for _, dimension := range result.Dimensions {
		expectedDimensions[dimension.Key] = dimension
	}
	dimensionIDs, dimensionKeys := make(map[string]bool, 13), make(map[string]bool, 13)
	for _, stored := range dimensions {
		expected, exists := expectedDimensions[stored.DimensionKey]
		if !exists || dimensionKeys[stored.DimensionKey] || !managementTraitsOpaqueID(stored.ID) || dimensionIDs[stored.ID] || stored.RunID != run.ID ||
			stored.DisplayOrder != expected.Order || stored.DimensionName != expected.Name || stored.ModuleKey != expected.Module ||
			stored.QuestionCount != expected.QuestionCount || stored.AnsweredCount != expected.AnsweredCount || stored.ScoreSum != expected.ScoreSum ||
			!cacheMatches(stored.Score, expected.Score) || !cacheMatches(stored.Norm, expected.Norm) || stored.Level == nil || *stored.Level != expected.Level {
			return ManagementTraitsResult{}, errManagementTraitsStoredResult
		}
		dimensionIDs[stored.ID], dimensionKeys[stored.DimensionKey] = true, true
	}
	expectedModules, moduleOrders := make(map[string]ManagementTraitsModuleScore, 4), make(map[string]int, 4)
	for i, module := range result.Modules {
		expectedModules[module.Key], moduleOrders[module.Key] = module, i+1
	}
	moduleIDs, moduleKeys := make(map[string]bool, 4), make(map[string]bool, 4)
	for _, stored := range modules {
		expected, exists := expectedModules[stored.ModuleKey]
		if !exists || moduleKeys[stored.ModuleKey] || !managementTraitsOpaqueID(stored.ID) || moduleIDs[stored.ID] || stored.RunID != run.ID ||
			stored.DisplayOrder != moduleOrders[stored.ModuleKey] || stored.DimensionCount != expected.DimensionCount || !cacheMatches(stored.Score, expected.Score) {
			return ManagementTraitsResult{}, errManagementTraitsStoredResult
		}
		moduleIDs[stored.ID], moduleKeys[stored.ModuleKey] = true, true
	}
	return result, nil
}
