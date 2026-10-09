package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

func managementRuntimeExpectPaper(t *testing.T, mock sqlmock.Sqlmock, f managementTraitsResultValidationFixture, paper model.Paper, profile model.ManagementTraitsExamProfile, owner managementTraitsRuntimeOwner) {
	t.Helper()
	mock.ExpectBegin()
	managementRuntimeExpectPaperBody(t, mock, f, paper, profile, owner)
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
}

func managementRuntimeExpectPaperBody(t *testing.T, mock sqlmock.Sqlmock, f managementTraitsResultValidationFixture, paper model.Paper, profile model.ManagementTraitsExamProfile, owner managementTraitsRuntimeOwner) {
	t.Helper()
	mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, paper))
	mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
	mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WillReturnRows(managementRuntimeModelRows(t, owner))
	qs, pq := make([]any, 140), make([]any, 140)
	for i, q := range f.Questions {
		qs[i] = q
		p := model.PaperQu{ID: q.PaperQuestionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, Sort: q.DisplayOrder, QuType: 1}
		if q.RawAnswer != nil {
			p.Answered, p.ActualScore = 1, *q.RawAnswer
		}
		pq[i] = p
	}
	mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(managementRuntimeModelRows(t, qs...))
	mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(managementRuntimeModelRows(t, pq...))
}

func TestManagementTraitsRuntimeFillExpiryNoWrites(t *testing.T) {
	f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
	for i := range f.Questions {
		f.Questions[i].SubmittedAt = nil
	}
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectPaper(t, mock, f, paper, profile, owner)
	mock.ExpectRollback()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
	claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: paper.CreateTime.Add(2 * time.Hour).Unix()}
	clockCalled := false
	_, err := s.fillAnswerTransaction(context.Background(), claims, f.Questions[0].PaperQuestionID, *f.Questions[0].SelectedOptionID, func() time.Time { clockCalled = true; return *paper.LimitTime })
	if !clockCalled || !errors.Is(err, ErrManagementTraitsRuntimeExpired) {
		t.Fatal("lock-time expiry not enforced", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsRuntimeFillCrossQuestionNoWrites(t *testing.T) {
	f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
	for i := range f.Questions {
		f.Questions[i].SubmittedAt = nil
	}
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectPaper(t, mock, f, paper, profile, owner)
	mock.ExpectRollback()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
	claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: paper.CreateTime.Add(2 * time.Hour).Unix()}
	_, err := s.fillAnswerTransaction(context.Background(), claims, f.Questions[0].PaperQuestionID, *f.Questions[1].SelectedOptionID, func() time.Time { return paper.CreateTime.Add(time.Minute) })
	if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("foreign option accepted", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsRuntimeSubmitRetryZeroWrites(t *testing.T) {
	for _, incomplete := range []bool{false, true} {
		f, records, paper, profile, owner := managementRuntimeLoadFixture(t, incomplete)
		db, mock := managementRuntimeDB(t)
		managementRuntimeExpectPaper(t, mock, f, paper, profile, owner)
		mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(managementRuntimeModelRows(t, records.Run))
		ds, ms := make([]any, 13), make([]any, 4)
		for i := range records.Dimensions {
			ds[i] = records.Dimensions[i]
		}
		for i := range records.Modules {
			ms[i] = records.Modules[i]
		}
		mock.ExpectQuery("SELECT .*el_mng_result_dimension").WillReturnRows(managementRuntimeModelRows(t, ds...))
		mock.ExpectQuery("SELECT .*el_mng_result_module").WillReturnRows(managementRuntimeModelRows(t, ms...))
		mock.ExpectQuery("SELECT .*el_mng_runtime_receipt").WillReturnRows(managementRuntimeModelRows(t, records.Receipt))
		managementRuntimeExpectBuckets(t, mock, f.Questions)
		mock.ExpectCommit()
		s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
		claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: paper.CreateTime.Add(2 * time.Hour).Unix()}
		got, err := s.submitTransaction(context.Background(), claims, "manual", func() time.Time { return paper.CreateTime.Add(time.Hour) })
		if err != nil || !got.Reused || !got.SubmittedAt.Equal(*records.Run.SubmittedAt) {
			t.Fatal("retry changed facts", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagementTraitsRuntimeSubmitRejectsClientTimeoutBeforeDB(t *testing.T) {
	s := NewManagementTraitsRuntimeService(nil, "test-only", 1024*1024)
	if _, err := s.submitTransaction(context.Background(), ManagementTraitsRuntimeClaims{}, "timeout", time.Now); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("client timeout accepted", err)
	}
}

func TestManagementTraitsRuntimeFillFrozenRawAndRollback(t *testing.T) {
	for _, fail := range []bool{false, true} {
		f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
		paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
		for i := range f.Questions {
			f.Questions[i].SubmittedAt = nil
		}
		before := managementTraitsContractClone(t, f.Questions)
		db, mock := managementRuntimeDB(t)
		managementRuntimeExpectPaper(t, mock, f, paper, profile, owner)
		managementRuntimeExpectBuckets(t, mock, f.Questions)
		mock.ExpectExec("UPDATE `el_mng_paper_question_snapshot`").WillReturnResult(sqlmock.NewResult(0, 1))
		if fail {
			mock.ExpectExec("UPDATE `el_paper_qu_answer`").WillReturnError(errors.New("private bucket failure"))
			mock.ExpectRollback()
		} else {
			mock.ExpectExec("UPDATE `el_paper_qu_answer`").WillReturnResult(sqlmock.NewResult(0, 2))
			mock.ExpectExec("UPDATE `el_paper_qu`").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
		}
		s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
		claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: paper.CreateTime.Add(2 * time.Hour).Unix()}
		count, err := s.fillAnswerTransaction(context.Background(), claims, f.Questions[0].PaperQuestionID, fmt.Sprintf("source-o-%d-4", f.Questions[0].Number), func() time.Time { return paper.CreateTime.Add(time.Minute) })
		if fail && !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatal("write did not rollback", err)
		}
		if !fail && (err != nil || count != 140) {
			t.Fatal("valid single choice rejected", err)
		}
		if !reflect.DeepEqual(before, f.Questions) {
			t.Fatal("source or snapshot metadata mutated")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagementTraitsRuntimeSubmitNewAtomicAndMissing(t *testing.T) {
	for _, scenario := range []string{"completed", "missing", "incomplete", "participant-failure"} {
		incomplete := scenario == "missing" || scenario == "incomplete"
		f, _, paper, profile, owner := managementRuntimeLoadFixture(t, incomplete)
		paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
		for i := range f.Questions {
			f.Questions[i].SubmittedAt = nil
		}
		db, mock := managementRuntimeDB(t)
		managementRuntimeExpectPaper(t, mock, f, paper, profile, owner)
		mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		if scenario == "missing" {
			mock.ExpectRollback()
		} else {
			managementRuntimeExpectBuckets(t, mock, f.Questions)
			mock.ExpectExec("INSERT INTO `el_mng_result_run`").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("INSERT INTO `el_mng_result_dimension`").WillReturnResult(sqlmock.NewResult(1, 13))
			mock.ExpectExec("INSERT INTO `el_mng_result_module`").WillReturnResult(sqlmock.NewResult(1, 4))
			mock.ExpectExec("INSERT INTO `el_mng_runtime_receipt`").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("UPDATE `el_mng_paper_question_snapshot`").WillReturnResult(sqlmock.NewResult(0, 140))
			mock.ExpectExec("UPDATE `el_paper`").WillReturnResult(sqlmock.NewResult(0, 1))
			if scenario == "participant-failure" {
				mock.ExpectExec("UPDATE `el_candidate`").WillReturnError(errors.New("private participant failure"))
				mock.ExpectRollback()
			} else {
				mock.ExpectExec("UPDATE `el_candidate`").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}
		}
		s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
		claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: paper.CreateTime.Add(2 * time.Hour).Unix()}
		now := paper.CreateTime.Add(time.Minute)
		if scenario == "incomplete" {
			now = *paper.LimitTime
		}
		got, err := s.submitTransaction(context.Background(), claims, "manual", func() time.Time { return now })
		if scenario == "missing" && !errors.Is(err, ErrManagementTraitsRuntimeMissing) {
			t.Fatal("missing answers wrote results", err)
		}
		if scenario == "participant-failure" && !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatal("final write did not rollback", err)
		}
		if scenario == "completed" && (err != nil || !got.Complete || got.Answered != 140) {
			t.Fatal("complete submit rejected", err)
		}
		if scenario == "incomplete" && (err != nil || got.Complete || got.Answered != 139 || got.Status != "incomplete") {
			t.Fatal("incomplete facts rejected", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}
