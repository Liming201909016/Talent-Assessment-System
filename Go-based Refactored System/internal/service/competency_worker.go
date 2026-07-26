package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/talent-assessment/refactored/internal/config"
	"gorm.io/gorm"
)

type CompetencyExpiryWorker struct {
	svc      *CompetencyRuntimeService
	db       *gorm.DB
	interval time.Duration
	batch    int
}

func NewCompetencyExpiryWorker(db *gorm.DB, cfg *config.Config, svc *CompetencyRuntimeService) *CompetencyExpiryWorker {
	return &CompetencyExpiryWorker{svc: svc, db: db, interval: time.Duration(cfg.Competency.ExpiryScanSeconds) * time.Second, batch: cfg.Competency.ExpiryBatchSize}
}

func (w *CompetencyExpiryWorker) Start(ctx context.Context) {
	go func() {
		if ctx.Err() != nil {
			return
		}
		if err := w.RunOnce(ctx); err != nil {
			slog.Error("competency initial expiry scan failed", "error", err)
		}
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.RunOnce(ctx); err != nil {
					slog.Error("competency expiry scan failed", "error", err)
				}
			}
		}
	}()
}

func (w *CompetencyExpiryWorker) RunOnce(ctx context.Context) error {
	type expiredPaper struct {
		PaperID         string
		ExamID          string
		ParticipantID   string
		ParticipantType string
	}
	rows := make([]expiredPaper, 0)
	if err := w.db.WithContext(ctx).Table("el_paper p").
		Select(`p.id AS paper_id, p.exam_id, p.user_id AS participant_id,
			CASE WHEN t.id IS NOT NULL THEN 'tester'
				 WHEN c.id IS NOT NULL THEN 'candidate' ELSE '' END AS participant_type`).
		Joins("INNER JOIN el_exam e ON e.id = p.exam_id AND e.assessment_type = ? AND e.scoring_mode = ?", AssessmentTypeCompetency, ScoringModeCompetencyAverage).
		Joins("LEFT JOIN el_tester t ON t.id = p.user_id AND t.paper_id = p.id AND t.exam_id = p.exam_id").
		Joins("LEFT JOIN el_candidate c ON c.id = p.user_id AND c.paper_id = p.id AND c.exam_id = p.exam_id").
		Where("p.state = 0 AND p.limit_time IS NOT NULL AND p.limit_time <= ?", time.Now()).
		Order("p.limit_time ASC, p.id ASC").Limit(w.batch).Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if row.ParticipantType == "" {
			slog.Error("competency expired paper owner missing", "paperId", row.PaperID)
			continue
		}
		_, err := w.svc.Submit(CompetencyTokenClaims{Purpose: CompetencyTokenPurposePaper, ParticipantType: row.ParticipantType, ParticipantID: row.ParticipantID, ExamID: row.ExamID, PaperID: row.PaperID}, CompetencySubmitTimeout)
		if err != nil {
			slog.Error("competency timeout submit failed", "paperId", row.PaperID, "error", err)
		}
	}
	return nil
}
