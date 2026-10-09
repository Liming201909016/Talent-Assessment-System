package service

import (
	"context"
	"errors"
	"time"
)

func runManagementTraitsExpiry(ctx context.Context, interval time.Duration, scan func()) {
	if ctx == nil || interval <= 0 || scan == nil || ctx.Err() != nil {
		return
	}
	scan()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() == nil {
				scan()
			}
		}
	}
}

// RunExpiry is independent of the 00401 worker and legacy expired papers.
// Startup and every restart immediately scan the original frozen deadline;
// the worker never updates timestamps or automatically creates reports.
func (s *ManagementTraitsRuntimeService) RunExpiry(ctx context.Context, interval time.Duration, batch int, onError func(error)) {
	// Owned by this sequential worker, not shared with HTTP calls or other workers.
	var cursor managementTraitsExpiryCursor
	runManagementTraitsExpiry(ctx, interval, func() {
		err := s.scanExpiry(ctx, batch, &cursor)
		if err != nil && !errors.Is(err, ErrManagementTraitsRuntimeClosed) && onError != nil {
			onError(err)
		}
	})
}

const managementTraitsExpiryBatchBudget = 10

type managementTraitsExpiryCursor struct {
	Deadline time.Time
	PaperID  string
}

func (s *ManagementTraitsRuntimeService) ScanExpiry(ctx context.Context, batch int) error {
	var cursor managementTraitsExpiryCursor
	return s.scanExpiry(ctx, batch, &cursor)
}

func (s *ManagementTraitsRuntimeService) scanExpiry(ctx context.Context, batch int, cursor *managementTraitsExpiryCursor) error {
	if ctx == nil || ctx.Err() != nil || batch <= 0 || batch > 1000 {
		return ErrManagementTraitsRuntimeInvalid
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return err
	}
	cutoff, failed := time.Now(), false
	for page := 0; page < managementTraitsExpiryBatchBudget; page++ {
		if ctx.Err() != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		rows := make([]struct {
			PaperID  string    `gorm:"column:paper_id"`
			ExamID   string    `gorm:"column:exam_id"`
			Kind     string    `gorm:"column:participant_type"`
			OwnerID  string    `gorm:"column:participant_id"`
			Deadline time.Time `gorm:"column:limit_time"`
		}, 0, batch)
		query := "SELECT s.paper_id, s.exam_id, s.participant_type, s.participant_id, s.limit_time FROM el_mng_paper_snapshot s INNER JOIN el_paper p ON p.id=s.paper_id AND p.exam_id=s.exam_id WHERE s.source = ? AND p.state = 1 AND s.limit_time <= ?"
		args := []any{"new_creation", cutoff}
		if cursor.PaperID != "" {
			query += " AND (s.limit_time > ? OR (s.limit_time = ? AND s.paper_id > ?))"
			args = append(args, cursor.Deadline, cursor.Deadline, cursor.PaperID)
		}
		query += " ORDER BY s.limit_time,s.paper_id LIMIT ?"
		args = append(args, batch)
		if s.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		for _, r := range rows {
			if ctx.Err() != nil {
				return ErrManagementTraitsRuntimeInvalid
			}
			c := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: r.Kind, ParticipantID: r.OwnerID, ExamID: r.ExamID, PaperID: r.PaperID, ExpiresAt: time.Now().Add(time.Minute).Unix()}
			if _, err := s.Submit(context.WithValue(ctx, managementRaceWorkerKey{}, true), c, "manual"); err != nil {
				failed = true
			}
			// Advance even on failure or removal from state=1, using scanned keys.
			*cursor = managementTraitsExpiryCursor{Deadline: r.Deadline, PaperID: r.PaperID}
		}
		if len(rows) < batch {
			// Retry failed/newly eligible rows on the next scan, not in this sweep.
			*cursor = managementTraitsExpiryCursor{}
			break
		}
	}
	if failed {
		return ErrManagementTraitsRuntimeInvalid
	}
	return nil
}
