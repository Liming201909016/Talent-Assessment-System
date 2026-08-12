package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
)

const (
	CompetencyTemporaryContentVersion     = "temp-v1"
	CompetencyTemporaryDisclaimer         = "临时测试文案，仅用于系统功能验证，不可作为人才决策依据。"
	CompetencyReportContentOverall        = "overall"
	CompetencyReportContentDimension      = "dimension"
	CompetencyReportContentGroup          = "group"
	CompetencyReportContentValidity       = "validity"
	CompetencyReportGroupDescriptionLevel = "description"
	CompetencyReportApprovalDraft         = "draft"
	CompetencyReportApprovalApproved      = "approved"
	CompetencyReportApprovalRetired       = "retired"
)

var ErrPhase1ReportContentNotApproved = errors.New("一期正式报告内容尚未完成双重批准")

type CompetencyReportTextSnapshot struct {
	ContentVersion string            `json:"contentVersion"`
	Audience       string            `json:"audience"`
	Disclaimer     string            `json:"disclaimer"`
	IsTemporary    bool              `json:"isTemporary"`
	OverallText    string            `json:"overallText"`
	DimensionTexts map[string]string `json:"dimensionTexts"`
}

type Phase1ReportTextSnapshot struct {
	ContentVersion string            `json:"contentVersion"`
	Audience       string            `json:"audience"`
	Disclaimer     string            `json:"disclaimer"`
	OverallText    string            `json:"overallText"`
	GroupTexts     map[string]string `json:"groupTexts"`
	DimensionTexts map[string]string `json:"dimensionTexts"`
	ValidityText   string            `json:"validityText"`
}

type Phase1ReportPage struct {
	Number       int      `json:"number"`
	Kind         string   `json:"kind"`
	DimensionIDs []string `json:"dimensionIds,omitempty"`
}

type Phase1ReportFramework struct {
	SchemaVersion             string             `json:"schemaVersion"`
	ReportKind                string             `json:"reportKind"`
	Renderable                bool               `json:"renderable"`
	BlockedReason             string             `json:"blockedReason"`
	OverallMaxScore           int                `json:"overallMaxScore"`
	GroupMaxScore             int                `json:"groupMaxScore"`
	DimensionMaxScore         int                `json:"dimensionMaxScore"`
	ShowRawScoreToParticipant bool               `json:"showRawScoreToParticipant"`
	Pages                     []Phase1ReportPage `json:"pages"`
}

type Phase1ReportResult struct {
	PaperID                string           `json:"paperId"`
	ParticipantName        string           `json:"participantName"`
	ParticipantAge         *int             `json:"participantAge"`
	ParticipantGender      string           `json:"participantGender"`
	ParticipantAffiliation string           `json:"participantAffiliation"`
	ParticipantPost        string           `json:"participantPost"`
	ParticipantDegree      string           `json:"participantDegree"`
	ParticipantMajor       string           `json:"participantMajor"`
	OverallScore           *decimal.Decimal `json:"overallScore"`
	OverallLevel           string           `json:"overallLevel"`
	SubmittedAt            *time.Time       `json:"submittedAt"`
}

type Phase1ReportGroup struct {
	GroupCode  string           `json:"groupCode"`
	GroupName  string           `json:"groupName"`
	GroupScore *decimal.Decimal `json:"groupScore"`
	LevelCode  string           `json:"levelCode"`
}

type Phase1ReportDimension struct {
	DimensionID    string           `json:"dimensionId"`
	DimensionCode  string           `json:"dimensionCode"`
	DimensionName  string           `json:"dimensionName"`
	DimensionScore *decimal.Decimal `json:"dimensionScore"`
	LevelCode      string           `json:"levelCode"`
}

type Phase1ReportValidity struct {
	Status string `json:"status"`
	Notice string `json:"notice"`
}

type Phase1FormalReportData struct {
	Phase1ReportFramework
	Result     Phase1ReportResult       `json:"result"`
	Groups     []Phase1ReportGroup      `json:"groups"`
	Dimensions []Phase1ReportDimension  `json:"dimensions"`
	Validity   Phase1ReportValidity     `json:"validity"`
	ReportText Phase1ReportTextSnapshot `json:"reportText"`
}

func NewPhase1ReportFramework() Phase1ReportFramework {
	dimensionIDs := NormalizePhase1CompetencyConfiguration().DimensionIDs
	pages := []Phase1ReportPage{
		{Number: 1, Kind: "cover"},
		{Number: 2, Kind: "reading_guide"},
		{Number: 3, Kind: "person_overall_validity"},
		{Number: 4, Kind: "group_overview"},
		{Number: 5, Kind: "dimension_radar"},
	}
	for index := 0; index < len(dimensionIDs); index += 2 {
		pages = append(pages, Phase1ReportPage{Number: len(pages) + 1, Kind: "dimension_details", DimensionIDs: append([]string(nil), dimensionIDs[index:index+2]...)})
	}
	return Phase1ReportFramework{
		SchemaVersion: "competency-phase1-report-data-v1", ReportKind: "frontline_phase1",
		Renderable: false, BlockedReason: ErrPhase1ReportContentNotApproved.Error(),
		OverallMaxScore: 50, GroupMaxScore: 5, DimensionMaxScore: 5,
		ShowRawScoreToParticipant: false, Pages: pages,
	}
}

func BuildPhase1FormalReportData(result model.CompetencyResult, groups []model.CompetencyGroupResult, dimensions []model.CompetencyDimensionResult, validity model.CompetencyValidityResult, text Phase1ReportTextSnapshot) (Phase1FormalReportData, error) {
	framework := NewPhase1ReportFramework()
	if !IsPhase1CompetencyVersionSet(CompetencyVersionSetFromResult(result)) || result.IsComplete != 1 || result.OverallScore == nil || result.EvaluationLevel == nil || validity.IsComplete != 1 || validity.ValidityStatus == nil {
		return Phase1FormalReportData{}, errors.New("一期正式报告结果不完整")
	}
	if *validity.ValidityStatus != CompetencyPhase1ValidityGood && *validity.ValidityStatus != CompetencyPhase1ValidityQuestionable {
		return Phase1FormalReportData{}, errors.New("一期正式报告效度状态无效")
	}
	if len(groups) != 2 || len(dimensions) != 10 {
		return Phase1FormalReportData{}, errors.New("一期正式报告必须包含2个一级维度和10个二级维度")
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].DisplayOrder < groups[j].DisplayOrder })
	sort.Slice(dimensions, func(i, j int) bool { return dimensions[i].DisplayOrder < dimensions[j].DisplayOrder })
	groupViews := make([]Phase1ReportGroup, 0, 2)
	for _, group := range groups {
		if group.IsComplete != 1 || group.GroupScore == nil || group.LevelCode == nil {
			return Phase1FormalReportData{}, errors.New("一期正式报告一级维度结果不完整")
		}
		groupViews = append(groupViews, Phase1ReportGroup{GroupCode: group.GroupCode, GroupName: group.GroupName, GroupScore: group.GroupScore, LevelCode: *group.LevelCode})
	}
	dimensionViews := make([]Phase1ReportDimension, 0, 10)
	profile := NormalizePhase1CompetencyConfiguration()
	for index, dimension := range dimensions {
		if dimension.DimensionID != profile.DimensionIDs[index] || dimension.IsComplete != 1 || dimension.DimensionScore == nil || dimension.LevelCode == nil {
			return Phase1FormalReportData{}, errors.New("一期正式报告二级维度结果不完整或顺序错误")
		}
		dimensionViews = append(dimensionViews, Phase1ReportDimension{DimensionID: dimension.DimensionID, DimensionCode: dimension.DimensionCode, DimensionName: dimension.DimensionName, DimensionScore: dimension.DimensionScore, LevelCode: *dimension.LevelCode})
	}
	framework.Renderable = true
	framework.BlockedReason = ""
	return Phase1FormalReportData{
		Phase1ReportFramework: framework,
		Result:                Phase1ReportResult{PaperID: result.PaperID, ParticipantName: result.ParticipantName, ParticipantAge: result.ParticipantAge, ParticipantGender: result.ParticipantGender, ParticipantAffiliation: result.ParticipantAffiliation, ParticipantPost: result.ParticipantPost, ParticipantDegree: result.ParticipantDegree, ParticipantMajor: result.ParticipantMajor, OverallScore: result.OverallScore, OverallLevel: *result.EvaluationLevel, SubmittedAt: result.SubmittedAt},
		Groups:                groupViews, Dimensions: dimensionViews,
		Validity: Phase1ReportValidity{Status: *validity.ValidityStatus, Notice: text.ValidityText}, ReportText: text,
	}, nil
}

func ValidatePhase1ReportContentApproval(row model.CompetencyReportContentPackage) error {
	versions := CompetencyVersionSet{ProductVersion: row.ProductVersion, ScoringVersion: row.ScoringVersion, ContentVersion: row.ContentVersion, ReportTemplateVersion: row.TemplateVersion}
	if !IsPhase1CompetencyVersionSet(versions) || row.Audience != CompetencyReportAudienceFrontlineEmployee || row.ApprovalStatus != CompetencyReportApprovalApproved {
		return ErrPhase1ReportContentNotApproved
	}
	if strings.TrimSpace(row.ContentApprovedBy) == "" || row.ContentApprovedAt == nil || strings.TrimSpace(row.PsychometricApprovedBy) == "" || row.PsychometricApprovedAt == nil {
		return ErrPhase1ReportContentNotApproved
	}
	if !isSHA256(row.QuestionSourceSHA256) || !isSHA256(row.ContentSourceSHA256) || strings.TrimSpace(row.EffectiveEnvironment) == "" || strings.TrimSpace(row.Disclaimer) == "" {
		return ErrPhase1ReportContentNotApproved
	}
	return nil
}

func isSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}

func BuildPhase1ReportTextSnapshot(contentVersion, audience, overallLevel, validityStatus string, dimensionLevels map[string]string, rows []model.CompetencyReportText) (string, error) {
	snapshot := Phase1ReportTextSnapshot{ContentVersion: contentVersion, Audience: audience, GroupTexts: make(map[string]string, 2), DimensionTexts: make(map[string]string, len(dimensionLevels))}
	for _, row := range rows {
		if row.ContentVersion != contentVersion || row.Audience != audience || row.Status != 0 {
			continue
		}
		if row.IsTemporary != 0 {
			return "", errors.New("一期正式报告不得使用临时文案")
		}
		disclaimer := strings.TrimSpace(row.Disclaimer)
		if disclaimer != "" {
			if snapshot.Disclaimer != "" && snapshot.Disclaimer != disclaimer {
				return "", errors.New("一期正式报告免责声明不一致")
			}
			snapshot.Disclaimer = disclaimer
		}
		switch row.ContentType {
		case CompetencyReportContentOverall:
			if row.DimensionID == "" && row.LevelCode == overallLevel {
				snapshot.OverallText = row.Content
			}
		case CompetencyReportContentGroup:
			if row.LevelCode == CompetencyReportGroupDescriptionLevel {
				snapshot.GroupTexts[row.DimensionID] = row.Content
			}
		case CompetencyReportContentDimension:
			if level, ok := dimensionLevels[row.DimensionID]; ok && row.LevelCode == level {
				snapshot.DimensionTexts[row.DimensionID] = row.Content
			}
		case CompetencyReportContentValidity:
			if row.DimensionID == "" && row.LevelCode == validityStatus {
				snapshot.ValidityText = row.Content
			}
		}
	}
	if strings.TrimSpace(snapshot.Disclaimer) == "" || strings.TrimSpace(snapshot.OverallText) == "" || strings.TrimSpace(snapshot.ValidityText) == "" {
		return "", errors.New("一期正式报告总体、效度或免责声明文案不完整")
	}
	for _, groupID := range []string{CompetencyPhase1GroupGeneralAbility, CompetencyPhase1GroupPsychologicalQuality} {
		if strings.TrimSpace(snapshot.GroupTexts[groupID]) == "" {
			return "", fmt.Errorf("一期正式报告一级维度文案缺失: %s", groupID)
		}
	}
	for _, dimensionID := range NormalizePhase1CompetencyConfiguration().DimensionIDs {
		if dimensionLevels[dimensionID] == "" || strings.TrimSpace(snapshot.DimensionTexts[dimensionID]) == "" {
			return "", fmt.Errorf("一期正式报告二级维度文案缺失: %s", dimensionID)
		}
	}
	data, err := json.Marshal(snapshot)
	return string(data), err
}

func BuildCompetencyReportTextSnapshot(contentVersion, audience, overallLevel string, dimensionLevels map[string]string, rows []model.CompetencyReportText) (string, error) {
	snapshot := CompetencyReportTextSnapshot{
		ContentVersion: contentVersion,
		Audience:       audience,
		DimensionTexts: make(map[string]string, len(dimensionLevels)),
	}
	if contentVersion == CompetencyTemporaryContentVersion {
		snapshot.Disclaimer = CompetencyTemporaryDisclaimer
		snapshot.IsTemporary = true
	}
	for _, row := range rows {
		if row.ContentVersion != contentVersion {
			continue
		}
		if row.Audience != audience || row.Status != 0 {
			continue
		}
		if row.Disclaimer != "" {
			snapshot.Disclaimer = row.Disclaimer
		}
		if row.IsTemporary == 1 {
			snapshot.IsTemporary = true
		}
		switch row.ContentType {
		case CompetencyReportContentOverall:
			if row.DimensionID == "" && row.LevelCode == overallLevel {
				snapshot.OverallText = row.Content
			}
		case CompetencyReportContentDimension:
			if level, ok := dimensionLevels[row.DimensionID]; ok && row.LevelCode == level {
				snapshot.DimensionTexts[row.DimensionID] = row.Content
			}
		}
	}
	if snapshot.OverallText == "" {
		return "", fmt.Errorf(
			"报告文案缺失: contentVersion=%s, audience=%s, dimension=overall, level=%s",
			contentVersion, audience, overallLevel,
		)
	}
	ids := make([]string, 0, len(dimensionLevels))
	for id := range dimensionLevels {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if snapshot.DimensionTexts[id] == "" {
			return "", fmt.Errorf(
				"报告文案缺失: contentVersion=%s, audience=%s, dimension=%s, level=%s",
				contentVersion, audience, id, dimensionLevels[id],
			)
		}
	}
	data, err := json.Marshal(snapshot)
	return string(data), err
}
