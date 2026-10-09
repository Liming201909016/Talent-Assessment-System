package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

type managementTraitsResumeFixture struct {
	f       managementTraitsResultValidationFixture
	paper   model.Paper
	profile model.ManagementTraitsExamProfile
	owner   managementTraitsRuntimeOwner
}

func managementTraitsResumeFixtureNew(t *testing.T) managementTraitsResumeFixture {
	t.Helper()
	f, _, p, profile, owner := managementRuntimeLoadFixture(t, false)
	start := time.Now().Add(-5 * time.Minute).Truncate(time.Second)
	deadline := start.Add(25 * time.Minute)
	f.Paper.StartedAt, f.Paper.CreatedAt, f.Paper.LimitTime = start, start, &deadline
	p.CreateTime, p.LimitTime, p.State, p.UserTime = &start, &deadline, 1, 0
	owner.EndTime, owner.Name = nil, "changed mutable name"
	for i := range f.Questions {
		f.Questions[i].RawAnswer, f.Questions[i].FinalScore, f.Questions[i].SelectedOptionID, f.Questions[i].SubmittedAt = nil, nil, nil, nil
		f.Questions[i].CreatedAt = start
	}
	return managementTraitsResumeFixture{f, p, profile, owner}
}

func managementTraitsResumeExpectLoad(t *testing.T, mock sqlmock.Sqlmock, v managementTraitsResumeFixture, stop string) {
	t.Helper()
	mock.ExpectBegin()
	queries := []struct {
		sql  string
		rows *sqlmock.Rows
	}{
		{"SELECT .*el_paper.*FOR UPDATE", managementRuntimeModelRows(t, v.paper)},
		{"SELECT .*el_mng_paper_snapshot", managementRuntimeModelRows(t, v.f.Paper)},
		{"SELECT .*el_mng_definition_bundle", managementRuntimeModelRows(t, v.f.Bundle)},
		{"SELECT .*el_mng_exam_profile", managementRuntimeModelRows(t, v.profile)},
		{"SELECT 'candidate'.*UNION ALL SELECT 'tester'", managementRuntimeModelRows(t, v.owner)},
	}
	for _, q := range queries {
		mock.ExpectQuery(q.sql).WillReturnRows(q.rows)
		if stop == q.sql {
			return
		}
	}
	qs, pq := make([]any, 140), make([]any, 140)
	for i, q := range v.f.Questions {
		qs[i] = q
		pq[i] = model.PaperQu{ID: q.PaperQuestionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, QuType: 1, Sort: q.DisplayOrder}
	}
	mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(managementRuntimeModelRows(t, qs...))
	mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(managementRuntimeModelRows(t, pq...))
}

func managementTraitsResumeExpectLocked(t *testing.T, mock sqlmock.Sqlmock, v managementTraitsResumeFixture, lockedBundle model.ManagementTraitsDefinitionBundle) {
	t.Helper()
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, lockedBundle))
	if !managementTraitsRuntimeSourceTrusted(lockedBundle.Status) {
		return
	}
	mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, v.profile))
}

func managementTraitsResumeExpectOwnerAndExam(t *testing.T, mock sqlmock.Sqlmock, v managementTraitsResumeFixture) {
	t.Helper()
	mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, v.owner))
	mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	end := time.Now().Add(-time.Hour)
	mock.ExpectQuery("SELECT .*el_exam").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: v.paper.ExamID, IsOpen: 1, State: 3, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, EndTime: &end}))
	managementRuntimeExpectBuckets(t, mock, v.f.Questions)
}

func TestManagementTraitsAdminResumeIssuedFrozenShortZeroWrites(t *testing.T) {
	for _, status := range []string{"candidate-current-source", "retired"} {
		t.Run(status, func(t *testing.T) {
			v := managementTraitsResumeFixtureNew(t)
			v.f.Bundle.Status = status
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			managementTraitsResumeExpectLoad(t, mock, v, "")
			managementTraitsResumeExpectLocked(t, mock, v, v.f.Bundle)
			managementTraitsResumeExpectOwnerAndExam(t, mock, v)
			mock.ExpectCommit()
			s := NewManagementTraitsRuntimeService(db, "resume-test-only", 1<<20)
			before := time.Now()
			got, err := s.IssueCandidateResume(context.Background(), v.paper.ExamID, v.owner.ID, v.paper.ID)
			if err != nil || got.ID != v.owner.ID || got.ExamID != v.paper.ExamID || got.PaperID != v.paper.ID || got.Name != "test-person" {
				t.Fatal("frozen resume rejected or mutable identity leaked", err)
			}
			claims, err := ParseManagementTraitsRuntimeToken("resume-test-only", got.ParticipantToken, ManagementTraitsRuntimePaperPurpose, time.Now())
			if err != nil || claims.ValidateBinding("candidate", got.ID, got.ExamID, got.PaperID) != nil || claims.ExpiresAt > before.Add(5*time.Minute+time.Second).Unix() || claims.ExpiresAt > v.f.Paper.LimitTime.Unix() {
				t.Fatal("invalid short credential", err)
			}
			if _, err := ParseManagementTraitsRuntimeToken("resume-test-only", got.ParticipantToken, ManagementTraitsRuntimeParticipantPurpose, time.Now()); err == nil {
				t.Fatal("paper credential admitted for creating paper")
			}
			if _, err := ParseManagementTraitsRuntimeToken("resume-test-only", got.ParticipantToken, ManagementTraitsRuntimePaperPurpose, time.Unix(claims.ExpiresAt, 0)); err == nil {
				t.Fatal("expired credential accepted")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsAdminResumeInvalidIDsNoQueries(t *testing.T) {
	for _, id := range []string{"", " ", "bad/id", "bad?token", "bad\n", string(make([]byte, 65))} {
		for i := 0; i < 3; i++ {
			db, mock := managementRuntimeDB(t)
			args := []string{"exam", "candidate", "paper"}
			args[i] = id
			_, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).IssueCandidateResume(context.Background(), args[0], args[1], args[2])
			if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Fatal("invalid ID accepted", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestManagementTraitsAdminResumeRejectsScopeAndFrozenFacts(t *testing.T) {
	for _, mode := range []string{"wrong_exam", "wrong_participant", "wrong_paper", "tester", "deleted", "completed", "ended_owner", "unfrozen", "foreign_profile", "historical", "revoked", "revoked_during_lock", "expired"} {
		t.Run(mode, func(t *testing.T) {
			v := managementTraitsResumeFixtureNew(t)
			exam, participant, paper := v.paper.ExamID, v.owner.ID, v.paper.ID
			stop := ""
			switch mode {
			case "wrong_exam":
				exam = "foreign-exam"
			case "wrong_participant":
				participant = "foreign-candidate"
			case "wrong_paper":
				paper = "foreign-paper"
				stop = "SELECT 'candidate'.*UNION ALL SELECT 'tester'"
			case "tester":
				v.owner.Kind, v.f.Paper.ParticipantType, v.owner.Status = "tester", "tester", "0"
			case "deleted":
				v.owner.DelFlag = 2
				stop = "SELECT 'candidate'.*UNION ALL SELECT 'tester'"
			case "completed":
				v.paper.State = 2
			case "ended_owner":
				end := time.Now()
				v.owner.EndTime = &end
			case "unfrozen":
				v.profile.FrozenAt = nil
			case "foreign_profile":
				v.f.Paper.ProfileExamID = &participant
			case "historical":
				v.f.Paper.Source = "historical_recompute"
			case "revoked":
				v.f.Bundle.Status = "review-revoked"
			case "expired":
				start := time.Now().Add(-25 * time.Minute).Truncate(time.Second)
				end := start.Add(25 * time.Minute)
				v.paper.CreateTime = &start
				v.paper.LimitTime = &end
				v.f.Paper.StartedAt = start
				v.f.Paper.CreatedAt = start
				v.f.Paper.LimitTime = &end
			}
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			managementTraitsResumeExpectLoad(t, mock, v, stop)
			if mode == "unfrozen" || mode == "revoked_during_lock" {
				locked := v.f.Bundle
				if mode == "revoked_during_lock" {
					locked.Status = "review-revoked"
				}
				managementTraitsResumeExpectLocked(t, mock, v, locked)
			}
			mock.ExpectRollback()
			got, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).IssueCandidateResume(context.Background(), exam, participant, paper)
			if err == nil || got.ParticipantToken != "" {
				t.Fatal("unsafe credential issued")
			}
			if mode == "expired" && !errors.Is(err, ErrManagementTraitsRuntimeExpired) {
				t.Fatal("lost controlled expiry", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsAdminResumeReadFailureStable(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnError(errors.New("PRIVATE_DSN_PASSWORD"))
	mock.ExpectRollback()
	got, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).IssueCandidateResume(context.Background(), "exam", "candidate", "paper")
	if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) || got.ParticipantToken != "" {
		t.Fatal("read failure exposed", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
