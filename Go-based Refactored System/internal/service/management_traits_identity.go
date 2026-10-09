package service

import "math/big"

// These are logical identities only, not repository mappings or activated versions.
const (
	ManagementTraitsQuestionnaireStaff  = "staff"
	ManagementTraitsQuestionnaireLeader = "leader"
)

type ManagementTraitsItemDefinition struct {
	Number  int
	Reverse bool
}

type ManagementTraitsDimensionDefinition struct {
	Key    string
	Name   string
	Module string
	Order  int
	Norm   *big.Rat
	Items  []ManagementTraitsItemDefinition
}

type managementTraitsDimensionSpec struct {
	key, name, module string
	normNumerator     int64
	normDenominator   int64
	terms             []int
}

// Negative V numbers identify reverse-scored items. Order is the approved tie order.
var managementTraitsDimensionSpecs = [...]managementTraitsDimensionSpec{
	{"self_confidence", "自信心", "self", 115, 2, []int{6, 19, 33, 47, 60, 66, 78, 92, 101, 107, -120, -133}},
	{"emotional_stability", "情绪稳定性", "self", 55, 1, []int{10, 24, 38, 52, 68, 82, 96, -111, -124, -137}},
	{"self_discipline", "自律性", "self", 225, 4, []int{11, -25, 39, 53, -69, 84, -97, -112, 125, 138}},
	{"sociality", "社会性", "interpersonal", 205, 4, []int{1, 14, -28, -42, -56, -72, -87, -102, -115, -128}},
	{"leadership", "领导性", "interpersonal", 105, 2, []int{3, 16, 31, 44, 58, 74, 89, 104, -117, -130}},
	{"interpersonal_sensitivity", "人际敏感性", "interpersonal", 50, 1, []int{5, 18, 32, 46, 50, 61, 65, 76, 83, 91, -106, -119, -132}},
	{"cooperation", "合作性", "interpersonal", 115, 2, []int{13, 27, 41, 55, 71, 86, 99, -114, 127, -140}},
	{"planning", "计划性", "task", 215, 4, []int{4, 17, 21, 30, 45, 59, 75, 81, 90, 105, 118, -131}},
	{"responsibility", "责任心", "task", 235, 4, []int{7, 20, 34, 48, 63, 77, -94, -108, -121, -134}},
	{"decisiveness", "决断性", "task", 215, 4, []int{-12, 26, 35, 40, 54, 62, 70, 85, 98, 113, 126, 139}},
	{"proactiveness", "进取性", "development", 55, 1, []int{2, 15, 29, 43, -57, 73, 88, 103, -116, 129}},
	{"learning", "学习力", "development", 215, 4, []int{-8, 22, 36, 49, -64, 79, 93, -100, 109, 122, -135}},
	{"innovation", "创新性", "development", 50, 1, []int{-9, 23, 37, -51, 67, -80, 95, 110, 123, -136}},
}

// ManagementTraitsDimensions returns detached definitions, including all mutable values.
func ManagementTraitsDimensions() []ManagementTraitsDimensionDefinition {
	definitions := make([]ManagementTraitsDimensionDefinition, 0, len(managementTraitsDimensionSpecs))
	for index, spec := range managementTraitsDimensionSpecs {
		items := make([]ManagementTraitsItemDefinition, 0, len(spec.terms))
		for _, term := range spec.terms {
			reverse := term < 0
			if reverse {
				term = -term
			}
			items = append(items, ManagementTraitsItemDefinition{Number: term, Reverse: reverse})
		}
		definitions = append(definitions, ManagementTraitsDimensionDefinition{
			Key: spec.key, Name: spec.name, Module: spec.module, Order: index + 1,
			Norm: big.NewRat(spec.normNumerator, spec.normDenominator), Items: items,
		})
	}
	return definitions
}
