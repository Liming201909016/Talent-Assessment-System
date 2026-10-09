package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// LEFT JOIN nulls are scanned to zero values and rejected, not lost via INNER JOIN.
type ManagementTraitsRuntimeSourceRow struct {
	RelationID    string `gorm:"column:relation_id"`
	QuestionID    string `gorm:"column:question_id"`
	Sort          int    `gorm:"column:sort"`
	RelationType  int    `gorm:"column:relation_type"`
	QuestionType  int    `gorm:"column:question_type"`
	Code          string `gorm:"column:code"`
	Title         string `gorm:"column:title"`
	OptionID      string `gorm:"column:option_id"`
	Raw           int    `gorm:"column:raw"`
	IsRight       int    `gorm:"column:is_right"`
	OptionContent string `gorm:"column:option_content"`
}

const managementTraitsRuntimeSourceSQL = `SELECT qr.id AS relation_id, q.id AS question_id, qr.sort AS sort,
qr.qu_type AS relation_type, q.qu_type AS question_type, q.content AS code, q.title AS title,
qa.id AS option_id, qa.score AS raw, qa.is_right AS is_right, qa.content AS option_content
FROM el_qu_repo qr LEFT JOIN el_qu q ON q.id = qr.qu_id
LEFT JOIN el_qu_answer qa ON qa.qu_id = q.id
WHERE qr.repo_id = ? ORDER BY qr.sort, qr.id, qa.score, qa.id LIMIT 701`

func loadManagementTraitsRuntimeSource(ctx context.Context, tx *gorm.DB, repoID string) ([]ManagementTraitsRuntimeSourceRow, error) {
	rows := make([]ManagementTraitsRuntimeSourceRow, 0, 700)
	if !managementTraitsOpaqueID(repoID) || tx.WithContext(ctx).Raw(managementTraitsRuntimeSourceSQL, repoID).Scan(&rows).Error != nil {
		return nil, ErrManagementTraitsRuntimeInvalid
	}
	if caps := managementTraitsSchemaCapabilities(tx); caps != nil {
		for _, row := range rows {
			if !managementTraitsSchemaIDFits(caps.Columns, "el_qu_repo.id", row.RelationID) || !managementTraitsSchemaIDFits(caps.Columns, "el_qu.id", row.QuestionID) || !managementTraitsSchemaIDFits(caps.Columns, "el_qu_answer.id", row.OptionID) {
				return nil, ErrManagementTraitsRuntimeInvalid
			}
		}
	}
	return rows, nil
}

// The entire current database text is the authority, including leader V67/V96.
// This does not claim equivalence with the customer's revised leader workbook.
func BuildManagementTraitsRuntimeCurrentSource(code string, rows []ManagementTraitsRuntimeSourceRow) (ManagementTraitsManifest, ManagementTraitsMapping, error) {
	questionnaire, versions, err := managementTraitsRuntimeVersions(code)
	if err != nil || len(rows) != 700 {
		return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
	}
	manifest := ManagementTraitsManifest{Questionnaire: questionnaire, Versions: versions, Questions: make([]ManagementTraitsManifestQuestion, 140)}
	mapping := ManagementTraitsMapping{Questionnaire: questionnaire, Questions: make([]ManagementTraitsMappedQuestion, 140)}
	var dimensions [141]string
	var reverse [141]bool
	for _, d := range ManagementTraitsDimensions() {
		for _, item := range d.Items {
			dimensions[item.Number], reverse[item.Number] = d.Key, item.Reverse
		}
	}
	var relations [141]string
	var options [141][6]bool
	var rightCount [141]int
	relationIDs, questionIDs, optionIDs := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, r := range rows {
		v := r.Sort
		if v < 1 || v > 140 || !managementTraitsOpaqueID(r.RelationID) || !managementTraitsOpaqueID(r.QuestionID) || !managementTraitsOpaqueID(r.OptionID) || r.RelationType != 1 || r.QuestionType != 1 || r.Code != fmt.Sprint("V", v) || r.Raw < 1 || r.Raw > 5 || optionIDs[r.OptionID] || options[v][r.Raw] || (r.IsRight != 0 && r.IsRight != 1) {
			return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
		}
		if relations[v] == "" {
			if relationIDs[r.RelationID] || questionIDs[r.QuestionID] {
				return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
			}
			relationIDs[r.RelationID], questionIDs[r.QuestionID], relations[v] = true, true, r.RelationID
			manifest.Questions[v-1] = ManagementTraitsManifestQuestion{Number: v, DimensionKey: dimensions[v], Reverse: reverse[v], Content: r.Title, Options: make([]ManagementTraitsManifestOption, 0, 5)}
			mapping.Questions[v-1] = ManagementTraitsMappedQuestion{Number: v, SourceQuestionID: r.QuestionID, Content: r.Title, Options: make([]ManagementTraitsMappedOption, 0, 5)}
		}
		mq, sq := &manifest.Questions[v-1], &mapping.Questions[v-1]
		if relations[v] != r.RelationID || sq.SourceQuestionID != r.QuestionID || mq.Content != r.Title {
			return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
		}
		if r.IsRight == 1 {
			rightCount[v]++
			if (!reverse[v] && r.Raw != 5) || (reverse[v] && r.Raw != 1) {
				return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
			}
		}
		optionIDs[r.OptionID], options[v][r.Raw] = true, true
		mq.Options = append(mq.Options, ManagementTraitsManifestOption{Raw: r.Raw, Content: r.OptionContent, DisplayOrder: r.Raw})
		sq.Options = append(sq.Options, ManagementTraitsMappedOption{SourceOptionID: r.OptionID, Raw: r.Raw, Content: r.OptionContent, DisplayOrder: r.Raw})
	}
	for v := 1; v <= 140; v++ {
		if rightCount[v] != 1 {
			return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
		}
	}
	mc, err := CanonicalManagementTraitsManifest(manifest)
	if err != nil {
		return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
	}
	// Decode canonical bytes so arbitrary query row order cannot change public output.
	manifest, err = DecodeManagementTraitsManifest(mc.JSON, len(mc.JSON))
	if err != nil {
		return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
	}
	mapping.ManifestSHA = mc.SHA256
	pc, err := CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil {
		return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
	}
	mapping, err = DecodeManagementTraitsMapping(pc.JSON, manifest, len(pc.JSON))
	if err != nil {
		return ManagementTraitsManifest{}, ManagementTraitsMapping{}, ErrManagementTraitsRuntimeInvalid
	}
	return manifest, mapping, nil
}
