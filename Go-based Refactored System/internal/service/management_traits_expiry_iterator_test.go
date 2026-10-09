package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm/logger"
)

// MT-EXPIRY-P1: execute the real scanner and Submit transactions; failed rows
// remain in progress, while later frozen new_creation papers must be visited.
const managementExpiryFirstQuery = `SELECT s.paper_id, s.exam_id, s.participant_type, s.participant_id(?:, s.limit_time)? FROM el_mng_paper_snapshot s INNER JOIN el_paper p ON p.id=s.paper_id AND p.exam_id=s.exam_id WHERE s.source = \? AND p.state = 1 AND s.limit_time <= \? ORDER BY s.limit_time,s.paper_id LIMIT \?`
const managementExpiryAfterQuery = `SELECT s.paper_id, s.exam_id, s.participant_type, s.participant_id, s.limit_time FROM el_mng_paper_snapshot s INNER JOIN el_paper p ON p.id=s.paper_id AND p.exam_id=s.exam_id WHERE s.source = \? AND p.state = 1 AND s.limit_time <= \? AND \(s.limit_time > \? OR \(s.limit_time = \? AND s.paper_id > \?\)\) ORDER BY s.limit_time,s.paper_id LIMIT \?`

type managementExpiryStrictLogger struct {
	logger.Interface
	unexpected []string
}

func (l *managementExpiryStrictLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil && (strings.Contains(err.Error(), "was not expected") || strings.Contains(err.Error(), "could not match") || strings.Contains(err.Error(), "arguments do not match") || strings.Contains(err.Error(), "all expectations were already fulfilled")) {
		sql, _ := fc()
		l.unexpected = append(l.unexpected, sql+": "+err.Error())
	}
}

func managementExpiryRows(ids []string, exam, kind, owner string, deadline time.Time) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"paper_id", "exam_id", "participant_type", "participant_id", "limit_time"})
	for _, id := range ids {
		rows.AddRow(id, exam, kind, owner, deadline)
	}
	return rows
}

func managementExpiryExpectAfter(mock sqlmock.Sqlmock, id string, deadline time.Time, batch int) *sqlmock.ExpectedQuery {
	return mock.ExpectQuery(managementExpiryAfterQuery).WithArgs("new_creation", sqlmock.AnyArg(), deadline, deadline, id, batch)
}

func managementExpiryExpectCorrupt(mock sqlmock.Sqlmock, id string) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WithArgs(id, 1).WillReturnError(errors.New("corrupt paper fixture"))
	mock.ExpectRollback()
}

func managementExpiryExpectSubmit(t *testing.T, mock sqlmock.Sqlmock, id string, revoked bool) time.Time {
	t.Helper()
	f, _, paper, profile, owner := managementRuntimeLoadFixture(t, true)
	// All cross-table paper bindings are real and frozen; only the selected ID
	// differs between fixtures. No definition/mapping/question content changes.
	f.Paper.PaperID, paper.ID, owner.PaperID = id, id, id
	paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
	for i := range f.Questions {
		f.Questions[i].PaperID, f.Questions[i].SubmittedAt = id, nil
	}
	if revoked {
		mock.ExpectBegin()
		managementRuntimeExpectPaperBody(t, mock, f, paper, profile, owner)
		current := f.Bundle
		current.Status = "review-revoked"
		mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, current))
		mock.ExpectRollback()
	} else {
		managementRuntimeExpectPaper(t, mock, f, paper, profile, owner)
		mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		managementRuntimeExpectBuckets(t, mock, f.Questions)
		mock.ExpectExec("INSERT INTO `el_mng_result_run`").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO `el_mng_result_dimension`").WillReturnResult(sqlmock.NewResult(1, 13))
		mock.ExpectExec("INSERT INTO `el_mng_result_module`").WillReturnResult(sqlmock.NewResult(1, 4))
		mock.ExpectExec("INSERT INTO `el_mng_runtime_receipt`").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE `el_mng_paper_question_snapshot` SET `submitted_at`=\\? WHERE paper_id = \\? AND submitted_at IS NULL").WithArgs(sqlmock.AnyArg(), id).WillReturnResult(sqlmock.NewResult(0, 140))
		mock.ExpectExec("UPDATE `el_paper`").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("UPDATE `el_candidate`").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
	}
	return *paper.LimitTime
}

func managementExpiryVerify(t *testing.T, mock sqlmock.Sqlmock, strict *managementExpiryStrictLogger) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if len(strict.unexpected) != 0 {
		t.Errorf("unexpected SQL (including revoked/legacy DML): %v", strict.unexpected)
	}
}

func TestBugManagementTraitsExpiryFailedHundredContinueAndRetry(t *testing.T) {
	f, _, _, _, _ := managementRuntimeLoadFixture(t, true)
	deadline := *f.Paper.LimitTime
	db, mock := managementRuntimeDB(t)
	strict := &managementExpiryStrictLogger{Interface: logger.Discard}
	db.Config.Logger = strict
	managementRuntimeExpectSchema(t, mock)
	ids := make([]string, 100)
	for i := range ids {
		ids[i] = fmt.Sprintf("failed-%03d", i)
	}
	for pass := 0; pass < 2; pass++ {
		mock.ExpectQuery(managementExpiryFirstQuery).WithArgs("new_creation", sqlmock.AnyArg(), 100).WillReturnRows(managementExpiryRows(ids, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline))
		for i, id := range ids {
			if i%2 == 0 {
				managementExpiryExpectSubmit(t, mock, id, true)
			} else {
				managementExpiryExpectCorrupt(mock, id)
			}
		}
		if pass == 0 {
			managementExpiryExpectAfter(mock, ids[99], deadline, 100).WillReturnRows(managementExpiryRows([]string{"healthy-101"}, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline))
			managementExpiryExpectSubmit(t, mock, "healthy-101", false)
		} else {
			// The successful row is now absent; the failed prefix is retried.
			managementExpiryExpectAfter(mock, ids[99], deadline, 100).WillReturnRows(managementExpiryRows(nil, "", "", "", deadline))
		}
	}
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	for pass := 0; pass < 2; pass++ {
		if err := s.ScanExpiry(context.Background(), 100); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Errorf("pass %d lost failed-row error: %v", pass, err)
		}
	}
	managementExpiryVerify(t, mock, strict)
}

func TestBugManagementTraitsExpiryRemovedSuccessStillAdvances(t *testing.T) {
	f, _, _, _, _ := managementRuntimeLoadFixture(t, true)
	deadline := *f.Paper.LimitTime
	db, mock := managementRuntimeDB(t)
	strict := &managementExpiryStrictLogger{Interface: logger.Discard}
	db.Config.Logger = strict
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectQuery(managementExpiryFirstQuery).WithArgs("new_creation", sqlmock.AnyArg(), 1).WillReturnRows(managementExpiryRows([]string{"healthy-a"}, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline))
	managementExpiryExpectSubmit(t, mock, "healthy-a", false)
	managementExpiryExpectAfter(mock, "healthy-a", deadline, 1).WillReturnRows(managementExpiryRows([]string{"healthy-b"}, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline.Add(time.Second)))
	managementExpiryExpectSubmit(t, mock, "healthy-b", false)
	managementExpiryExpectAfter(mock, "healthy-b", deadline.Add(time.Second), 1).WillReturnRows(managementExpiryRows(nil, "", "", "", deadline))
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	if err := s.ScanExpiry(context.Background(), 1); err != nil {
		t.Error(err)
	}
	managementExpiryVerify(t, mock, strict)
}

func TestBugManagementTraitsExpiryWorkerBudgetRestartWrap(t *testing.T) {
	f, _, _, _, _ := managementRuntimeLoadFixture(t, true)
	deadline := *f.Paper.LimitTime
	db, mock := managementRuntimeDB(t)
	strict := &managementExpiryStrictLogger{Interface: logger.Discard}
	db.Config.Logger = strict
	managementRuntimeExpectSchema(t, mock)
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	// A fresh RunExpiry starts at the beginning. Ten full failed batches must
	// not reset its cursor at the budget boundary; the next tick reaches #11.
	for restart := 0; restart < 2; restart++ {
		for i := 0; i < 10; i++ {
			id := fmt.Sprintf("failed-%03d", i)
			var query *sqlmock.ExpectedQuery
			if i == 0 {
				query = mock.ExpectQuery(managementExpiryFirstQuery).WithArgs("new_creation", sqlmock.AnyArg(), 1)
			} else {
				query = managementExpiryExpectAfter(mock, fmt.Sprintf("failed-%03d", i-1), deadline, 1)
			}
			query.WillReturnRows(managementExpiryRows([]string{id}, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline))
			managementExpiryExpectCorrupt(mock, id)
		}
		managementExpiryExpectAfter(mock, "failed-009", deadline, 1).WillReturnRows(managementExpiryRows([]string{"healthy-011"}, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline))
		managementExpiryExpectSubmit(t, mock, "healthy-011", false)
		managementExpiryExpectAfter(mock, "healthy-011", deadline, 1).WillReturnRows(managementExpiryRows(nil, "", "", "", deadline))
		// End-of-sweep resets only for the next scan, which retries the prefix.
		mock.ExpectQuery(managementExpiryFirstQuery).WithArgs("new_creation", sqlmock.AnyArg(), 1).WillReturnRows(managementExpiryRows([]string{"failed-000"}, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline))
		managementExpiryExpectCorrupt(mock, "failed-000")
		managementExpiryExpectAfter(mock, "failed-000", deadline, 1).WillReturnRows(managementExpiryRows(nil, "", "", "", deadline))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		calls := 0
		s.RunExpiry(ctx, time.Millisecond, 1, func(error) {
			calls++
			if calls == 2 {
				cancel()
			}
		})
		cancel()
		if calls != 2 {
			t.Errorf("restart %d callbacks=%d", restart, calls)
		}
	}
	managementExpiryVerify(t, mock, strict)
}

func TestBugManagementTraitsExpiryQueryFailuresCheckedAndRetry(t *testing.T) {
	for _, scenario := range []string{"query", "row", "scan", "after-query"} {
		t.Run(scenario, func(t *testing.T) {
			f, _, _, _, _ := managementRuntimeLoadFixture(t, true)
			deadline := *f.Paper.LimitTime
			db, mock := managementRuntimeDB(t)
			strict := &managementExpiryStrictLogger{Interface: logger.Discard}
			db.Config.Logger = strict
			managementRuntimeExpectSchema(t, mock)
			q := mock.ExpectQuery(managementExpiryFirstQuery).WithArgs("new_creation", sqlmock.AnyArg(), 1)
			switch scenario {
			case "query":
				q.WillReturnError(errors.New("private scan failure"))
			case "row":
				q.WillReturnRows(managementExpiryRows([]string{"failed-000"}, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline).RowError(0, errors.New("private row failure")))
			case "scan":
				q.WillReturnRows(sqlmock.NewRows([]string{"paper_id", "limit_time"}).AddRow("failed-000", "not-a-time"))
			case "after-query":
				q.WillReturnRows(managementExpiryRows([]string{"failed-000"}, f.Paper.ExamID, f.Paper.ParticipantType, f.Paper.ParticipantID, deadline))
				managementExpiryExpectCorrupt(mock, "failed-000")
				managementExpiryExpectAfter(mock, "failed-000", deadline, 1).WillReturnError(errors.New("private second page failure"))
			}
			mock.ExpectQuery(managementExpiryFirstQuery).WithArgs("new_creation", sqlmock.AnyArg(), 1).WillReturnRows(managementExpiryRows(nil, "", "", "", deadline))
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			if err := s.ScanExpiry(context.Background(), 1); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Errorf("unchecked %s error: %v", scenario, err)
			}
			if err := s.ScanExpiry(context.Background(), 1); err != nil {
				t.Error("fresh sweep did not recover", err)
			}
			managementExpiryVerify(t, mock, strict)
		})
	}
}

func TestBugManagementTraitsExpiryCanceledWorkerNoSQL(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	strict := &managementExpiryStrictLogger{Interface: logger.Discard}
	db.Config.Logger = strict
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.RunExpiry(ctx, time.Millisecond, 100, func(error) { t.Error("canceled worker reported a scan") })
	if err := s.ScanExpiry(ctx, 100); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Error("canceled scan accepted", err)
	}
	managementExpiryVerify(t, mock, strict)
}
