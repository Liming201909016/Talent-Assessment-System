package service

import (
	"fmt"
	"math/big"
)

const (
	CompetencyPhase1ProductVersionV2        = "competency-frontline-phase1-v2"
	CompetencyPhase1ScoringVersionV2        = "competency-phase1-scoring-v2"
	CompetencyPhase1ContentVersionV2        = "competency-phase1-content-v2"
	CompetencyPhase1ReportTemplateVersionV2 = "competency-phase1-report-v2"
)

type Phase1V2DimensionDefinition struct {
	ID           string
	StableKey    string
	DisplayCode  string
	Name         string
	ModuleCode   string
	DisplayOrder int
}

var phase1V2Dimensions = []Phase1V2DimensionDefinition{
	{ID: "competency-logical-reasoning", StableKey: "logical_reasoning", DisplayCode: "A1-01", Name: "逻辑思维", ModuleCode: "task_management", DisplayOrder: 1},
	{ID: "competency-plan-execution", StableKey: "plan_execution", DisplayCode: "A1-02", Name: "计划执行", ModuleCode: "task_management", DisplayOrder: 2},
	{ID: "competency-digital-application", StableKey: "digital_application", DisplayCode: "A1-03", Name: "数字应用", ModuleCode: "task_management", DisplayOrder: 3},
	{ID: "competency-achievement-orientation", StableKey: "achievement_orientation", DisplayCode: "A1-04", Name: "成就导向", ModuleCode: "task_management", DisplayOrder: 4},
	{ID: "competency-continuous-learning", StableKey: "continuous_learning", DisplayCode: "A1-05", Name: "持续学习", ModuleCode: "task_management", DisplayOrder: 5},
	{ID: "competency-communication", StableKey: "communication", DisplayCode: "B1-01", Name: "沟通表达", ModuleCode: "interpersonal_management", DisplayOrder: 6},
	{ID: "competency-cooperation", StableKey: "cooperation", DisplayCode: "B1-02", Name: "合作意识", ModuleCode: "interpersonal_management", DisplayOrder: 7},
	{ID: "competency-truth-pragmatism", StableKey: "truth_pragmatism", DisplayCode: "C1-01", Name: "求真务实", ModuleCode: "self_management", DisplayOrder: 8},
	{ID: "competency-self-discipline", StableKey: "self_discipline", DisplayCode: "C1-02", Name: "自律性", ModuleCode: "self_management", DisplayOrder: 9},
	{ID: "competency-dedication", StableKey: "dedication", DisplayCode: "C1-03", Name: "敬业奉献", ModuleCode: "self_management", DisplayOrder: 10},
}

var phase1V2DimensionNormScores = []*big.Rat{
	big.NewRat(55, 1), big.NewRat(65, 1), big.NewRat(60, 1), big.NewRat(55, 1), big.NewRat(55, 1),
	big.NewRat(105, 2), big.NewRat(55, 1), big.NewRat(115, 2), big.NewRat(125, 2), big.NewRat(60, 1),
}

var phase1V1ToV2DimensionIndex = map[string]int{
	"competency-a1-01": 0,
	"competency-a1-03": 1,
	"competency-a1-02": 2,
	"competency-b1-04": 3,
	"competency-a1-04": 4,
	"competency-a1-05": 5,
	"competency-b1-05": 6,
	"competency-b1-02": 7,
	"competency-b1-03": 8,
	"competency-b1-01": 9,
}

func Phase1V2VersionSet() CompetencyVersionSet {
	return CompetencyVersionSet{
		ProductVersion:        CompetencyPhase1ProductVersionV2,
		ScoringVersion:        CompetencyPhase1ScoringVersionV2,
		ContentVersion:        CompetencyPhase1ContentVersionV2,
		ReportTemplateVersion: CompetencyPhase1ReportTemplateVersionV2,
	}
}

func IsPhase1V2VersionSet(versions CompetencyVersionSet) bool {
	return versions == Phase1V2VersionSet()
}

func Phase1V2Dimensions() []Phase1V2DimensionDefinition {
	return append([]Phase1V2DimensionDefinition(nil), phase1V2Dimensions...)
}

func Phase1V2DimensionNormScores() []*big.Rat {
	scores := make([]*big.Rat, 0, len(phase1V2DimensionNormScores))
	for _, score := range phase1V2DimensionNormScores {
		scores = append(scores, new(big.Rat).Set(score))
	}
	return scores
}

func MapPhase1V1DimensionToV2(sourceDimensionID string) (Phase1V2DimensionDefinition, error) {
	index, ok := phase1V1ToV2DimensionIndex[sourceDimensionID]
	if !ok {
		return Phase1V2DimensionDefinition{}, fmt.Errorf("unsupported phase-1 v1 dimension identity %q", sourceDimensionID)
	}
	return phase1V2Dimensions[index], nil
}
