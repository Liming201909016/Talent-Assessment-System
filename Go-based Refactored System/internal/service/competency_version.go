package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/talent-assessment/refactored/internal/model"
)

const (
	CompetencyProductV1        = "competency-generic-v1"
	CompetencyScoringV1        = "competency-v1"
	CompetencyReportTemplateV1 = "competency-report-v1"

	CompetencyPhase1ProductVersion        = "competency-frontline-phase1-v1"
	CompetencyPhase1ScoringVersion        = "competency-phase1-scoring-v1"
	CompetencyPhase1ContentVersion        = "competency-phase1-content-v1"
	CompetencyPhase1ReportTemplateVersion = "competency-phase1-report-v1"
)

var competencyPhase1DimensionIDs = []string{
	"competency-a1-01", "competency-a1-02", "competency-a1-03", "competency-a1-04", "competency-a1-05",
	"competency-b1-01", "competency-b1-02", "competency-b1-03", "competency-b1-04", "competency-b1-05",
}

var competencyVersionIdentifier = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,31}$`)

type CompetencyVersionSet struct {
	ProductVersion        string `json:"productVersion"`
	ScoringVersion        string `json:"scoringVersion"`
	ContentVersion        string `json:"contentVersion"`
	ReportTemplateVersion string `json:"reportTemplateVersion"`
}

type Phase1CompetencyConfiguration struct {
	ReportAudience         string
	DimensionIDs           []string
	Versions               CompetencyVersionSet
	DefaultDurationMinutes int
	DimensionQuestionCount int
	ValidityQuestionCount  int
}

type Phase1QuestionTypeCounts struct {
	Dimension int
	Validity  int
	Other     int
}

type Phase1CompetencyGroupDefinition struct {
	Code              string
	Name              string
	DisplayOrder      int
	ChildDimensionIDs []string
}

func NormalizePhase1CompetencyConfiguration() Phase1CompetencyConfiguration {
	return Phase1CompetencyConfiguration{
		ReportAudience: CompetencyReportAudienceFrontlineEmployee,
		DimensionIDs:   append([]string(nil), competencyPhase1DimensionIDs...),
		Versions: CompetencyVersionSet{
			ProductVersion:        CompetencyPhase1ProductVersion,
			ScoringVersion:        CompetencyPhase1ScoringVersion,
			ContentVersion:        CompetencyPhase1ContentVersion,
			ReportTemplateVersion: CompetencyPhase1ReportTemplateVersion,
		},
		DefaultDurationMinutes: 20,
		DimensionQuestionCount: 80,
		ValidityQuestionCount:  10,
	}
}

func Phase1CompetencyGroups() []Phase1CompetencyGroupDefinition {
	return []Phase1CompetencyGroupDefinition{
		{Code: CompetencyPhase1GroupGeneralAbility, Name: "通用能力", DisplayOrder: 1, ChildDimensionIDs: append([]string(nil), competencyPhase1DimensionIDs[:5]...)},
		{Code: CompetencyPhase1GroupPsychologicalQuality, Name: "心理素养", DisplayOrder: 2, ChildDimensionIDs: append([]string(nil), competencyPhase1DimensionIDs[5:]...)},
	}
}

func ValidatePhase1CompetencyConfiguration(audience string, dimensionIDs []string, versions CompetencyVersionSet) error {
	expected := NormalizePhase1CompetencyConfiguration()
	if audience != expected.ReportAudience || versions != expected.Versions || len(dimensionIDs) != len(expected.DimensionIDs) {
		return errors.New("invalid phase-1 competency configuration")
	}
	for index, id := range dimensionIDs {
		if id != expected.DimensionIDs[index] {
			return errors.New("invalid phase-1 competency configuration")
		}
	}
	return nil
}

func ValidatePhase1QuestionInventory(dimensionIDs []string, counts map[string]Phase1QuestionTypeCounts) error {
	for _, dimensionID := range dimensionIDs {
		count := counts[dimensionID]
		if count.Dimension != 8 || count.Validity != 1 || count.Other != 0 {
			return fmt.Errorf("维度%s题本数量不符合一期固定配置", dimensionID)
		}
	}
	return nil
}

func IsPhase1CompetencyVersionSet(versions CompetencyVersionSet) bool {
	return versions == NormalizePhase1CompetencyConfiguration().Versions
}

func NormalizeCompetencyVersionSet(input CompetencyVersionSet) (CompetencyVersionSet, error) {
	versions := CompetencyVersionSet{
		ProductVersion:        strings.TrimSpace(input.ProductVersion),
		ScoringVersion:        strings.TrimSpace(input.ScoringVersion),
		ContentVersion:        strings.TrimSpace(input.ContentVersion),
		ReportTemplateVersion: strings.TrimSpace(input.ReportTemplateVersion),
	}
	if versions.ProductVersion == "" {
		versions.ProductVersion = CompetencyProductV1
	}
	if versions.ScoringVersion == "" {
		versions.ScoringVersion = CompetencyScoringV1
	}
	if versions.ContentVersion == "" {
		versions.ContentVersion = CompetencyTemporaryContentVersion
	}
	if versions.ReportTemplateVersion == "" {
		versions.ReportTemplateVersion = CompetencyReportTemplateV1
	}
	if err := ValidateFrozenCompetencyVersionSet(versions); err != nil {
		return CompetencyVersionSet{}, err
	}
	return versions, nil
}

func ValidateFrozenCompetencyVersionSet(versions CompetencyVersionSet) error {
	for _, version := range []string{
		versions.ProductVersion,
		versions.ScoringVersion,
		versions.ContentVersion,
		versions.ReportTemplateVersion,
	} {
		if !competencyVersionIdentifier.MatchString(version) {
			return errors.New("胜任力版本标识只能包含小写字母、数字、点、下划线和连字符，且最长32字符")
		}
	}
	if err := ValidateExecutableCompetencyVersions(versions); err != nil {
		return err
	}
	return nil
}

func ValidateExecutableCompetencyVersions(versions CompetencyVersionSet) error {
	if versions.ProductVersion != CompetencyProductV1 {
		return errors.New("当前程序不支持该胜任力产品版本")
	}
	if versions.ScoringVersion != CompetencyScoringV1 {
		return errors.New("当前程序不支持该胜任力评分版本")
	}
	if versions.ReportTemplateVersion != CompetencyReportTemplateV1 {
		return errors.New("当前程序不支持该胜任力报告模板版本")
	}
	return nil
}

func CompetencyVersionSetFromExam(exam model.Exam) CompetencyVersionSet {
	return CompetencyVersionSet{
		ProductVersion:        exam.CompetencyProductVersion,
		ScoringVersion:        exam.CompetencyScoringVersion,
		ContentVersion:        exam.CompetencyContentVersion,
		ReportTemplateVersion: exam.CompetencyReportTemplateVersion,
	}
}

func CompetencyVersionSetFromResult(result model.CompetencyResult) CompetencyVersionSet {
	return CompetencyVersionSet{
		ProductVersion:        result.ProductVersion,
		ScoringVersion:        result.ScoringVersion,
		ContentVersion:        result.ContentVersion,
		ReportTemplateVersion: result.ReportTemplateVersion,
	}
}

func ApplyCompetencyVersions(exam *model.Exam, versions CompetencyVersionSet) {
	exam.CompetencyProductVersion = versions.ProductVersion
	exam.CompetencyScoringVersion = versions.ScoringVersion
	exam.CompetencyContentVersion = versions.ContentVersion
	exam.CompetencyReportTemplateVersion = versions.ReportTemplateVersion
}

func ClearCompetencyVersions(exam *model.Exam) {
	exam.CompetencyProductVersion = ""
	exam.CompetencyScoringVersion = ""
	exam.CompetencyContentVersion = ""
	exam.CompetencyReportTemplateVersion = ""
}
