package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestManagementTraitsExpiryStartupImmediateAndCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	runManagementTraitsExpiry(ctx, time.Hour, func() { calls++; cancel() })
	if calls != 1 {
		t.Fatal("startup scan waited for ticker", calls)
	}
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	calls = 0
	runManagementTraitsExpiry(ctx, time.Hour, func() { calls++ })
	if calls != 0 {
		t.Fatal("canceled worker scanned")
	}
}

func TestManagementTraitsExpiryOnlyNewSnapshots(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectQuery("SELECT s.paper_id, s.exam_id, s.participant_type, s.participant_id, s.limit_time FROM el_mng_paper_snapshot s INNER JOIN el_paper p ON p.id=s.paper_id AND p.exam_id=s.exam_id WHERE s.source = \\? AND p.state = 1 AND s.limit_time <= \\? ORDER BY s.limit_time,s.paper_id LIMIT \\?").WithArgs("new_creation", sqlmock.AnyArg(), 100).WillReturnRows(sqlmock.NewRows([]string{"paper_id", "exam_id", "participant_type", "participant_id", "limit_time"}))
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	if err := s.ScanExpiry(context.Background(), 100); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
