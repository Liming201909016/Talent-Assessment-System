package service

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/talent-assessment/refactored/internal/model"
)

const (
	CompetencyTemporaryContentVersion = "temp-v1"
	CompetencyTemporaryDisclaimer     = "临时测试文案，仅用于系统功能验证，不可作为人才决策依据。"
	CompetencyReportContentOverall    = "overall"
	CompetencyReportContentDimension  = "dimension"
)

type CompetencyReportTextSnapshot struct {
	ContentVersion string            `json:"contentVersion"`
	Audience       string            `json:"audience"`
	Disclaimer     string            `json:"disclaimer"`
	IsTemporary    bool              `json:"isTemporary"`
	OverallText    string            `json:"overallText"`
	DimensionTexts map[string]string `json:"dimensionTexts"`
}

func BuildCompetencyReportTextSnapshot(audience, overallLevel string, dimensionLevels map[string]string, rows []model.CompetencyReportText) (string, error) {
	snapshot := CompetencyReportTextSnapshot{
		ContentVersion: CompetencyTemporaryContentVersion,
		Audience:       audience, Disclaimer: CompetencyTemporaryDisclaimer, IsTemporary: true,
		DimensionTexts: make(map[string]string, len(dimensionLevels)),
	}
	for _, row := range rows {
		if row.ContentVersion != "" && row.ContentVersion != CompetencyTemporaryContentVersion {
			continue
		}
		if row.Audience != audience || row.Status != 0 {
			continue
		}
		if row.Disclaimer != "" {
			snapshot.Disclaimer = row.Disclaimer
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
			CompetencyTemporaryContentVersion, audience, overallLevel,
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
				CompetencyTemporaryContentVersion, audience, id, dimensionLevels[id],
			)
		}
	}
	data, err := json.Marshal(snapshot)
	return string(data), err
}
