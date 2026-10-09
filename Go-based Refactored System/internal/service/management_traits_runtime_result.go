package service

import (
	"math/big"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

type managementTraitsRuntimeRecords struct {
	Run        model.ManagementTraitsResultRun
	Dimensions []model.ManagementTraitsResultDimension
	Modules    []model.ManagementTraitsResultModule
	Receipt    model.ManagementTraitsRuntimeReceipt
}

func managementTraitsRuntimeDecimal(r *big.Rat) *decimal.Decimal {
	return decimalFromRat(r)
}

func buildManagementTraitsRuntimeRecords(bundle model.ManagementTraitsDefinitionBundle, paper model.ManagementTraitsPaperSnapshot, questions []model.ManagementTraitsPaperQuestionSnapshot, runID string, now time.Time, budget int) (managementTraitsRuntimeRecords, error) {
	v := ManagementTraitsScoringVersions{Product: bundle.ProductVersion, Question: bundle.QuestionVersion, Scoring: bundle.ScoringVersion, Norm: bundle.NormVersion}
	if !managementTraitsOpaqueID(runID) || validateManagementTraitsRuntimeVersions(bundle.Questionnaire, v) != nil || !managementTraitsRuntimeSourceTrusted(bundle.Status) || paper.StartedAt.IsZero() || paper.LimitTime == nil || !paper.LimitTime.Equal(paper.StartedAt.Add(25*time.Minute)) || now.Before(paper.StartedAt) {
		return managementTraitsRuntimeRecords{}, ErrManagementTraitsRuntimeInvalid
	}
	input, err := ValidateManagementTraitsStoredInput(bundle, paper, questions, budget)
	if err != nil {
		return managementTraitsRuntimeRecords{}, ErrManagementTraitsRuntimeInvalid
	}
	result, err := CalculateManagementTraits(bundle.Questionnaire, input.Answers)
	if err != nil {
		return managementTraitsRuntimeRecords{}, ErrManagementTraitsRuntimeInvalid
	}
	expired := !now.Before(*paper.LimitTime)
	if !expired && !result.IsComplete {
		return managementTraitsRuntimeRecords{}, ErrManagementTraitsRuntimeMissing
	}
	// Only the server's post-lock timestamp is supplied by a transactional caller.
	seconds64 := int64(now.Sub(paper.StartedAt) / time.Second)
	if seconds64 < 0 || seconds64 > 2147483647 {
		return managementTraitsRuntimeRecords{}, ErrManagementTraitsRuntimeInvalid
	}
	seconds := int(seconds64)
	status, submitType := "completed", "manual"
	if !result.IsComplete {
		status = "incomplete"
	}
	if expired {
		submitType = "timeout"
	}
	var level *string
	if result.IsComplete {
		x := result.OverallLevel
		level = &x
	}
	r := managementTraitsRuntimeRecords{
		Run: model.ManagementTraitsResultRun{ID: runID, PaperID: paper.PaperID, ExamID: paper.ExamID,
			ProductVersion: v.Product, QuestionVersion: v.Question, ScoringVersion: v.Scoring, NormVersion: v.Norm,
			Questionnaire: bundle.Questionnaire, ParticipantType: paper.ParticipantType, ParticipantID: paper.ParticipantID,
			ScoringManifestSHA: bundle.ScoringManifestSHA, InputSHA: input.Canonical.SHA256, Status: status, Source: "submission",
			TotalQuestionCount: 140, AnsweredQuestionCount: result.AnsweredQuestionCount,
			OverallScore: managementTraitsRuntimeDecimal(result.OverallScore), OverallNorm: managementTraitsRuntimeDecimal(result.OverallNorm), OverallLevel: level,
			UserTimeSeconds: &seconds, SubmittedAt: &now, CreatedAt: now},
		Dimensions: make([]model.ManagementTraitsResultDimension, 0, 13), Modules: make([]model.ManagementTraitsResultModule, 0, 4),
		Receipt: model.ManagementTraitsRuntimeReceipt{RunID: runID, PaperID: paper.PaperID, ExamID: paper.ExamID,
			ParticipantType: paper.ParticipantType, ParticipantID: paper.ParticipantID, SubmitType: submitType,
			StartedAt: paper.StartedAt, LimitTime: *paper.LimitTime, SubmittedAt: now, UserTimeSeconds: seconds},
	}
	for _, d := range result.Dimensions {
		var l *string
		if result.IsComplete {
			x := d.Level
			l = &x
		}
		r.Dimensions = append(r.Dimensions, model.ManagementTraitsResultDimension{ID: runID + "-d-" + d.Key, RunID: runID,
			DimensionKey: d.Key, DisplayOrder: d.Order, DimensionName: d.Name, ModuleKey: d.Module,
			QuestionCount: d.QuestionCount, AnsweredCount: d.AnsweredCount, ScoreSum: d.ScoreSum,
			Score: managementTraitsRuntimeDecimal(d.Score), Norm: managementTraitsRuntimeDecimal(d.Norm), Level: l, CreatedAt: now})
	}
	for i, m := range result.Modules {
		r.Modules = append(r.Modules, model.ManagementTraitsResultModule{ID: runID + "-m-" + m.Key, RunID: runID,
			ModuleKey: m.Key, DisplayOrder: i + 1, DimensionCount: m.DimensionCount, Score: managementTraitsRuntimeDecimal(m.Score), CreatedAt: now})
	}
	// Derived IDs must also fit the actual independent model contract.
	for _, d := range r.Dimensions {
		if !managementTraitsOpaqueID(d.ID) {
			return managementTraitsRuntimeRecords{}, ErrManagementTraitsRuntimeInvalid
		}
	}
	for _, m := range r.Modules {
		if !managementTraitsOpaqueID(m.ID) {
			return managementTraitsRuntimeRecords{}, ErrManagementTraitsRuntimeInvalid
		}
	}
	return r, nil
}

// Caller owns a transaction and holds the paper lock. Never upsert or repair.
func persistManagementTraitsRuntimeRecords(tx *gorm.DB, r managementTraitsRuntimeRecords) error {
	if tx.Create(&r.Run).Error != nil || tx.CreateInBatches(r.Dimensions, 100).Error != nil || tx.CreateInBatches(r.Modules, 100).Error != nil || tx.Create(&r.Receipt).Error != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	return nil
}

func validateManagementTraitsRuntimeResult(bundle model.ManagementTraitsDefinitionBundle, snapshot model.ManagementTraitsPaperSnapshot, questions []model.ManagementTraitsPaperQuestionSnapshot, records managementTraitsRuntimeRecords, budget int) (ManagementTraitsResult, error) {
	r, receipt := records.Run, records.Receipt
	if r.SubmittedAt == nil || r.UserTimeSeconds == nil || r.Source != "submission" || !r.CreatedAt.Equal(*r.SubmittedAt) || snapshot.LimitTime == nil || receipt.RunID != r.ID || receipt.PaperID != snapshot.PaperID || receipt.ExamID != snapshot.ExamID || receipt.ParticipantType != snapshot.ParticipantType || receipt.ParticipantID != snapshot.ParticipantID || !receipt.StartedAt.Equal(snapshot.StartedAt) || !receipt.LimitTime.Equal(*snapshot.LimitTime) || !receipt.SubmittedAt.Equal(*r.SubmittedAt) || receipt.UserTimeSeconds != *r.UserTimeSeconds {
		return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
	}
	for _, q := range questions {
		if q.SubmittedAt == nil || !q.SubmittedAt.Equal(*r.SubmittedAt) || !q.CreatedAt.Equal(snapshot.StartedAt) {
			return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
		}
	}
	for _, d := range records.Dimensions {
		if !d.CreatedAt.Equal(*r.SubmittedAt) {
			return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
		}
	}
	for _, m := range records.Modules {
		if !m.CreatedAt.Equal(*r.SubmittedAt) {
			return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
		}
	}
	expected, err := buildManagementTraitsRuntimeRecords(bundle, snapshot, questions, r.ID, *r.SubmittedAt, budget)
	if err != nil || receipt.SubmitType != expected.Receipt.SubmitType || r.Status != expected.Run.Status {
		return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
	}
	if r.Status == "completed" {
		result, err := ValidateManagementTraitsStoredResult(bundle, snapshot, questions, r, records.Dimensions, records.Modules, budget)
		if err != nil {
			return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
		}
		return result, nil
	}
	// Incomplete results deliberately do not use completed-only S2D. Rebuild
	// counts and NULLs from frozen unanswered/answered rows, never write repairs.
	// Equal instants can have different driver timezone representations.
	expected.Run.CreatedAt, expected.Run.SubmittedAt = r.CreatedAt, r.SubmittedAt
	if r.Status != "incomplete" || !reflect.DeepEqual(r, expected.Run) || len(records.Dimensions) != 13 || len(records.Modules) != 4 {
		return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
	}
	seen := make(map[string]bool, 17)
	for _, d := range records.Dimensions {
		if !managementTraitsOpaqueID(d.ID) || seen["d:"+d.ID] {
			return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
		}
		seen["d:"+d.ID] = true
		found := false
		for _, e := range expected.Dimensions {
			if d.DimensionKey == e.DimensionKey {
				e.ID = d.ID
				e.CreatedAt = d.CreatedAt
				found = reflect.DeepEqual(d, e)
			}
		}
		if !found || seen["dk:"+d.DimensionKey] {
			return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
		}
		seen["dk:"+d.DimensionKey] = true
	}
	for _, m := range records.Modules {
		if !managementTraitsOpaqueID(m.ID) || seen["m:"+m.ID] {
			return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
		}
		seen["m:"+m.ID] = true
		found := false
		for _, e := range expected.Modules {
			if m.ModuleKey == e.ModuleKey {
				e.ID = m.ID
				e.CreatedAt = m.CreatedAt
				found = reflect.DeepEqual(m, e)
			}
		}
		if !found || seen["mk:"+m.ModuleKey] {
			return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
		}
		seen["mk:"+m.ModuleKey] = true
	}
	input, err := ValidateManagementTraitsStoredInput(bundle, snapshot, questions, budget)
	if err != nil {
		return ManagementTraitsResult{}, ErrManagementTraitsRuntimeInvalid
	}
	return CalculateManagementTraits(bundle.Questionnaire, input.Answers)
}
