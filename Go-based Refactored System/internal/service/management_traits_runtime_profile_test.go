package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

// Public freeze must drain the legacy mutation lease before starting its transaction.
func TestBugManagementTraitsPublicFreezeWaitsForLegacyLease(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		name := "commit_after_release"
		if cancelled {
			name = "cancel_while_waiting"
		}
		t.Run(name, func(t *testing.T) {
			f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
			if err := s.CheckRuntimeSchema(context.Background()); err != nil {
				t.Fatal(err)
			}
			if !cancelled {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: profile.ExamID, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, RequiredFields: "name,telephone"}))
				mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
				mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
				mock.ExpectCommit()
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			lease := LockManagementTraitsLegacyMutation()
			defer lease.Release()
			type result struct {
				profile model.ManagementTraitsExamProfile
				err     error
			}
			done := make(chan result, 1)
			started := make(chan struct{})
			go func() {
				close(started)
				got, err := s.FreezeProfile(ctx, profile.ExamID)
				done <- result{got, err}
			}()
			<-started
			var got result
			early := false
			select {
			case got = <-done:
				early = true
			case <-time.After(100 * time.Millisecond):
			}
			if cancelled {
				cancel()
			}
			lease.Release()
			if !early {
				select {
				case got = <-done:
				case <-time.After(2 * time.Second):
					t.Fatal("public freeze did not finish after lease release")
				}
			}
			if early {
				t.Fatalf("public freeze returned while legacy lease was held: err=%v", got.err)
			}
			if cancelled {
				if !errors.Is(got.err, ErrManagementTraitsRuntimeInvalid) || got.profile.ExamID != "" {
					t.Fatal("cancelled freeze was not rejected", got.err)
				}
			} else if got.err != nil || got.profile.MappingSHA != profile.MappingSHA || !got.profile.FrozenAt.Equal(*profile.FrozenAt) {
				t.Fatal("public freeze did not preserve frozen profile", got.err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsRuntimeFreezeExistingReadOnly(t *testing.T) {
	f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
	db, mock := managementRuntimeDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: profile.ExamID, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, RequiredFields: "name,telephone"}))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectCommit()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
	got, err := s.freezeProfileTransaction(context.Background(), profile.ExamID, func() time.Time { return time.Now() })
	if err != nil || got.MappingSHA != profile.MappingSHA || !got.FrozenAt.Equal(*profile.FrozenAt) {
		t.Fatal("freeze retry rebound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsRuntimeFreezeRejectsExistingPaper(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: "exam", AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, RequiredFields: "name,telephone"}))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
	mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT count.*el_paper").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
	if _, err := s.freezeProfileTransaction(context.Background(), "exam", time.Now); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("legacy paper scope frozen", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsRuntimeFreezeAtomicCurrentSource(t *testing.T) {
	for _, failure := range []bool{false, true} {
		db, mock := managementRuntimeDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: "exam", AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, RequiredFields: "name,telephone"}))
		mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
		draftFreezeSchema(mock)
		created := time.Now().Add(-time.Minute)
		mock.ExpectQuery("SELECT .*el_mng_exam_draft.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.ManagementTraitsExamDraft{ExamID: "exam", RepoID: "repo", RepoCode: "00202", Lifecycle: "draft", CreatedAt: created, UpdatedAt: created}))
		mock.ExpectQuery("SELECT count.*el_paper").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery("SELECT .*el_exam_repo").WillReturnRows(managementRuntimeModelRows(t, model.ExamRepo{ID: "link", ExamID: "exam", RepoID: "repo", RadioCount: 140}))
		repo := model.Repo{ID: "repo", Code: "00202", RadioCount: 140}
		mock.ExpectQuery("SELECT .*el_repo").WillReturnRows(managementRuntimeModelRows(t, repo))
		mock.ExpectQuery("SELECT .*el_repo").WillReturnRows(managementRuntimeModelRows(t, repo))
		source := managementRuntimeSourceRows()
		values := make([]any, len(source))
		for i := range source {
			values[i] = source[i]
		}
		mock.ExpectQuery("SELECT .* FROM el_qu_repo qr LEFT JOIN el_qu q.*LEFT JOIN el_qu_answer qa").WillReturnRows(managementRuntimeModelRows(t, values...))
		mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectExec("INSERT INTO `el_mng_definition_bundle`").WillReturnResult(sqlmock.NewResult(1, 1))
		if failure {
			mock.ExpectExec("INSERT INTO `el_mng_exam_profile`").WillReturnError(errors.New("private profile failure"))
			mock.ExpectRollback()
		} else {
			mock.ExpectExec("INSERT INTO `el_mng_exam_profile`").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("UPDATE `el_mng_exam_draft` SET .* WHERE exam_id = \\? AND lifecycle = \\?").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
		}
		s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
		profile, err := s.freezeProfileTransaction(context.Background(), "exam", time.Now)
		if failure && !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatal("profile error not rolled back", err)
		}
		if !failure && (err != nil || profile.TotalTimeMinutes != 25 || profile.MappingSHA == "") {
			t.Fatal("freeze rejected", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}
