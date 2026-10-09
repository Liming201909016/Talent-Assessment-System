package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

type ManagementTraitsRuntimeSubmission struct {
	RunID       string    `json:"runId"`
	Status      string    `json:"status"`
	Complete    bool      `json:"complete"`
	Answered    int       `json:"answered"`
	SubmittedAt time.Time `json:"submittedAt"`
	Reused      bool      `json:"reused"`
}

func (s *ManagementTraitsRuntimeService) FillAnswer(ctx context.Context, c ManagementTraitsRuntimeClaims, pqID, optionID string) (int, error) {
	if s == nil || s.db == nil {
		return 0, ErrManagementTraitsRuntimeClosed
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return 0, err
	}
	return s.fillAnswerTransaction(ctx, c, pqID, optionID, time.Now)
}

// Submit is the trusted internal submission boundary used by the expiry scan.
// Claims supply owner binding only; participant credential expiry is not an
// internal authority. HTTP callers must use SubmitParticipant instead.
func (s *ManagementTraitsRuntimeService) Submit(ctx context.Context, c ManagementTraitsRuntimeClaims, submitType string) (result ManagementTraitsRuntimeSubmission, err error) {
	caller := managementRaceInternal
	if ctx != nil && ctx.Value(managementRaceWorkerKey{}) == true {
		caller = managementRaceWorker
	}
	ctx, observation := s.startRaceObservation(ctx, c.PaperID, caller)
	defer func() { observation.emit(result, err) }()
	if s == nil || s.db == nil {
		return ManagementTraitsRuntimeSubmission{}, ErrManagementTraitsRuntimeClosed
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return ManagementTraitsRuntimeSubmission{}, err
	}
	return s.submitRuntimeTransaction(ctx, c, submitType, time.Now, nil)
}

func (s *ManagementTraitsRuntimeService) SubmitParticipant(ctx context.Context, c ManagementTraitsRuntimeClaims, submitType string) (result ManagementTraitsRuntimeSubmission, err error) {
	ctx, observation := s.startRaceObservation(ctx, c.PaperID, managementRaceParticipant)
	defer func() { observation.emit(result, err) }()
	if s == nil || s.db == nil {
		return ManagementTraitsRuntimeSubmission{}, ErrManagementTraitsRuntimeClosed
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return ManagementTraitsRuntimeSubmission{}, err
	}
	return s.submitTransaction(ctx, c, submitType, time.Now)
}

// Unexported until all legacy writers share a freeze-safe scope gate. The clock
// is sampled after the same paper lock used by submit, never before waiting.
func (s *ManagementTraitsRuntimeService) fillAnswerTransaction(ctx context.Context, claims ManagementTraitsRuntimeClaims, pqID, optionID string, clock func() time.Time) (int, error) {
	if s.db == nil || !claims.valid() || claims.Purpose != ManagementTraitsRuntimePaperPurpose || !managementTraitsOpaqueID(pqID) || !managementTraitsOpaqueID(optionID) || clock == nil {
		return 0, ErrManagementTraitsRuntimeInvalid
	}
	count := 0
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		paper, snapshot, _, _, owner, questions, err := s.loadRuntimePaperForWrite(ctx, tx, claims.PaperID)
		if err != nil || claims.ValidateBinding(owner.Kind, owner.ID, paper.ExamID, paper.ID) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		now := clock()
		if err := managementTraitsRuntimeCredentialTime(claims, now); err != nil {
			return err
		}
		if paper.State != 1 || owner.EndTime != nil || now.Before(snapshot.StartedAt) {
			return ErrManagementTraitsRuntimeInvalid
		}
		if !now.Before(*snapshot.LimitTime) {
			return ErrManagementTraitsRuntimeExpired
		}
		var target *model.ManagementTraitsPaperQuestionSnapshot
		for i := range questions {
			if questions[i].SubmittedAt != nil {
				return ErrManagementTraitsRuntimeInvalid
			}
			if questions[i].RawAnswer != nil {
				count++
			}
			if questions[i].PaperQuestionID == pqID {
				target = &questions[i]
			}
		}
		if target == nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var options []ManagementTraitsMappedOption
		if managementTraitsDecodeStrict([]byte(target.OptionsSnapshot), s.maxJSONBytes, &options) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var selected ManagementTraitsMappedOption
		for _, o := range options {
			if o.SourceOptionID == optionID {
				selected = o
			}
		}
		if selected.Raw < 1 || selected.Raw > 5 {
			return ErrManagementTraitsRuntimeInvalid
		}
		if validateManagementTraitsRuntimeBuckets(tx, paper.ID, questions, s.maxJSONBytes) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		if target.SelectedOptionID != nil && *target.SelectedOptionID == optionID {
			return nil
		}
		final := selected.Raw
		if target.Reverse {
			final = 6 - final
		}
		q := tx.Model(&model.ManagementTraitsPaperQuestionSnapshot{}).Where("id = ? AND paper_id = ?", target.ID, paper.ID).
			Updates(map[string]any{"selected_option_id": optionID, "raw_answer": selected.Raw, "final_score": final})
		if q.Error != nil || q.RowsAffected != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		// Five existing frozen buckets, strictly one checked. No source option read.
		if tx.Model(&model.PaperQuAnswer{}).Where("paper_id = ? AND qu_id = ?", paper.ID, target.SourceQuestionID).
			Update("checked", gorm.Expr("CASE WHEN answer_id = ? THEN 1 ELSE 0 END", optionID)).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		p := tx.Model(&model.PaperQu{}).Where("id = ? AND paper_id = ?", pqID, paper.ID).
			Updates(map[string]any{"answered": 1, "actual_score": selected.Raw, "answer": selected.Content})
		if p.Error != nil || p.RowsAffected != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		if target.RawAnswer == nil {
			count++
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrManagementTraitsRuntimeToken) {
			return 0, ErrManagementTraitsRuntimeToken
		}
		if errors.Is(err, ErrManagementTraitsRuntimeExpired) {
			return 0, ErrManagementTraitsRuntimeExpired
		}
		return 0, ErrManagementTraitsRuntimeInvalid
	}
	return count, nil
}

func validateManagementTraitsRuntimeBuckets(tx *gorm.DB, paperID string, questions []model.ManagementTraitsPaperQuestionSnapshot, budget int) error {
	var buckets []model.PaperQuAnswer
	if tx.Where("paper_id = ?", paperID).Limit(701).Find(&buckets).Error != nil || len(buckets) != 700 {
		return ErrManagementTraitsRuntimeInvalid
	}
	expected := make(map[string]model.PaperQuAnswer, 700)
	for _, q := range questions {
		var options []ManagementTraitsMappedOption
		if managementTraitsDecodeStrict([]byte(q.OptionsSnapshot), budget, &options) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		for _, o := range options {
			checked := int8(0)
			if q.SelectedOptionID != nil && *q.SelectedOptionID == o.SourceOptionID {
				checked = 1
			}
			expected[o.SourceOptionID] = model.PaperQuAnswer{PaperID: paperID, QuID: q.SourceQuestionID, AnswerID: o.SourceOptionID, Checked: checked, Score: o.Raw, Sort: o.DisplayOrder}
		}
	}
	ids, seen := map[string]bool{}, map[string]bool{}
	for _, b := range buckets {
		e, ok := expected[b.AnswerID]
		if !ok || !managementTraitsOpaqueID(b.ID) || ids[b.ID] || seen[b.AnswerID] || b.PaperID != e.PaperID || b.QuID != e.QuID || b.Checked != e.Checked || b.Score != e.Score || b.Sort != e.Sort {
			return ErrManagementTraitsRuntimeInvalid
		}
		ids[b.ID], seen[b.AnswerID] = true, true
	}
	return nil
}

func (s *ManagementTraitsRuntimeService) submitTransaction(ctx context.Context, claims ManagementTraitsRuntimeClaims, submitType string, clock func() time.Time) (ManagementTraitsRuntimeSubmission, error) {
	return s.submitRuntimeTransaction(ctx, claims, submitType, clock, func(now time.Time) error {
		return managementTraitsRuntimeCredentialTime(claims, now)
	})
}

// A nil credential-time check is reserved for the explicit trusted internal
// Submit boundary, never for the participant transaction or HTTP handler.
func (s *ManagementTraitsRuntimeService) submitRuntimeTransaction(ctx context.Context, claims ManagementTraitsRuntimeClaims, submitType string, clock func() time.Time, validateTime func(time.Time) error) (ManagementTraitsRuntimeSubmission, error) {
	managementRaceMark(ctx, "settlement_entered")
	if submitType != "manual" || s.db == nil || !claims.valid() || claims.Purpose != ManagementTraitsRuntimePaperPurpose || clock == nil {
		return ManagementTraitsRuntimeSubmission{}, ErrManagementTraitsRuntimeInvalid
	}
	var summary ManagementTraitsRuntimeSubmission
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		managementRaceMark(ctx, "transaction_entered")
		paper, snapshot, bundle, _, owner, questions, err := s.loadRuntimePaperForWrite(ctx, tx, claims.PaperID)
		if err != nil || claims.ValidateBinding(owner.Kind, owner.ID, paper.ExamID, paper.ID) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		now := clock()
		if validateTime != nil {
			if err := validateTime(now); err != nil {
				return err
			}
		}
		now = now.Truncate(time.Second)
		var runs []model.ManagementTraitsResultRun
		if tx.Where("paper_id = ?", paper.ID).Limit(2).Find(&runs).Error != nil || len(runs) > 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		if len(runs) == 1 {
			r := runs[0]
			var dims []model.ManagementTraitsResultDimension
			var modules []model.ManagementTraitsResultModule
			var receipts []model.ManagementTraitsRuntimeReceipt
			if tx.Where("run_id = ?", r.ID).Limit(14).Find(&dims).Error != nil || tx.Where("run_id = ?", r.ID).Limit(5).Find(&modules).Error != nil || tx.Where("run_id = ?", r.ID).Limit(2).Find(&receipts).Error != nil || len(receipts) != 1 || paper.State != 2 || r.SubmittedAt == nil || r.UserTimeSeconds == nil || paper.UserTime != *r.UserTimeSeconds || owner.EndTime == nil || !owner.EndTime.Equal(*r.SubmittedAt) {
				return ErrManagementTraitsRuntimeInvalid
			}
			if _, err := validateManagementTraitsRuntimeResult(bundle, snapshot, questions, managementTraitsRuntimeRecords{Run: r, Dimensions: dims, Modules: modules, Receipt: receipts[0]}, s.maxJSONBytes); err != nil {
				return err
			}
			if validateManagementTraitsRuntimeBuckets(tx, paper.ID, questions, s.maxJSONBytes) != nil {
				return ErrManagementTraitsRuntimeInvalid
			}
			summary = ManagementTraitsRuntimeSubmission{RunID: r.ID, Status: r.Status, Complete: r.Status == "completed", Answered: r.AnsweredQuestionCount, SubmittedAt: *r.SubmittedAt, Reused: true}
			managementRaceMark(ctx, "body_returned")
			return nil
		}
		if paper.State != 1 || owner.EndTime != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		for _, q := range questions {
			if q.SubmittedAt != nil {
				return ErrManagementTraitsRuntimeInvalid
			}
		}
		r, err := buildManagementTraitsRuntimeRecords(bundle, snapshot, questions, uuid.NewString(), now, s.maxJSONBytes)
		if err != nil {
			return err
		}
		if validateManagementTraitsRuntimeBuckets(tx, paper.ID, questions, s.maxJSONBytes) != nil || persistManagementTraitsRuntimeRecords(tx, r) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		q := tx.Model(&model.ManagementTraitsPaperQuestionSnapshot{}).Where("paper_id = ? AND submitted_at IS NULL", paper.ID).Update("submitted_at", now)
		if q.Error != nil || q.RowsAffected != 140 {
			return ErrManagementTraitsRuntimeInvalid
		}
		p := tx.Model(&model.Paper{}).Where("id = ? AND state = ?", paper.ID, 1).
			Updates(map[string]any{"state": 2, "user_time": *r.Run.UserTimeSeconds, "update_time": now})
		if p.Error != nil || p.RowsAffected != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		table := "el_candidate"
		if owner.Kind == "tester" {
			table = "el_tester"
		}
		person := tx.Table(table).Where("id = ? AND paper_id = ? AND exam_id = ? AND end_time IS NULL", owner.ID, paper.ID, paper.ExamID).
			Updates(map[string]any{"end_time": now, "update_time": now})
		if person.Error != nil || person.RowsAffected != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		summary = ManagementTraitsRuntimeSubmission{RunID: r.Run.ID, Status: r.Run.Status, Complete: r.Run.Status == "completed", Answered: r.Run.AnsweredQuestionCount, SubmittedAt: now}
		managementRaceMark(ctx, "body_returned")
		return nil
	})
	managementRaceMark(ctx, "transaction_returned")
	if err != nil {
		if errors.Is(err, ErrManagementTraitsRuntimeToken) {
			return ManagementTraitsRuntimeSubmission{}, ErrManagementTraitsRuntimeToken
		}
		if errors.Is(err, ErrManagementTraitsRuntimeMissing) {
			return ManagementTraitsRuntimeSubmission{}, ErrManagementTraitsRuntimeMissing
		}
		return ManagementTraitsRuntimeSubmission{}, ErrManagementTraitsRuntimeInvalid
	}
	return summary, nil
}

// Used only for a sanitized public DTO, never as a source of scoring metadata.
func managementTraitsRuntimePublicOptions(raw string, budget int) ([]ManagementTraitsRuntimeOption, error) {
	var options []ManagementTraitsMappedOption
	if managementTraitsDecodeStrict([]byte(raw), budget, &options) != nil {
		return nil, ErrManagementTraitsRuntimeInvalid
	}
	result := make([]ManagementTraitsRuntimeOption, 0, len(options))
	for _, o := range options {
		result = append(result, ManagementTraitsRuntimeOption{ID: o.SourceOptionID, Content: o.Content, DisplayOrder: o.DisplayOrder})
	}
	return result, nil
}

type ManagementTraitsRuntimeOption struct {
	ID           string `json:"id"`
	Content      string `json:"content"`
	DisplayOrder int    `json:"displayOrder"`
}
