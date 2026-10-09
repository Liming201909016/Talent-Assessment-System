package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestManagementTraitsSetExamState(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		from, to   int
		wantOK     bool
	}{
		{"enable 00501", "00501", 1, 0, true},
		{"disable 00502", "00502", 0, 1, true},
		{"idempotent disabled", "00501", 1, 1, true},
		{"reject legacy 002", "00201", 1, 0, false},
		{"reject terminal state", "00501", 3, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
			if _, versions, err := managementTraitsRuntimeVersions(tc.code); err == nil {
				fixture.Bundle.ProductVersion = versions.Product
			}
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WithArgs("exam", 1).WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: "exam", State: tc.from, AssessmentType: "legacy", ScoringMode: "legacy"}))
			if tc.from == 0 || tc.from == 1 {
				mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WithArgs("exam", 1).WillReturnRows(managementRuntimeModelRows(t, profile))
				mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WithArgs(profile.BundleID, 1).WillReturnRows(managementRuntimeModelRows(t, fixture.Bundle))
				mock.ExpectQuery("SELECT r.code.*el_exam_repo").WithArgs("exam", 2).WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow(tc.code))
			}
			if tc.wantOK && tc.from != tc.to {
				mock.ExpectExec("UPDATE `el_exam` SET `state`=\\? WHERE id = \\? AND state = \\?").WithArgs(tc.to, "exam", tc.from).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			} else if tc.wantOK {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			got, err := s.SetExamState(context.Background(), "exam", tc.to)
			if tc.wantOK {
				if err != nil || got != tc.to {
					t.Fatalf("state=%d err=%v", got, err)
				}
			} else if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Fatalf("unsafe transition err=%v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsSetExamStateRejectsInputBeforeSQL(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	for _, tc := range []struct {
		id    string
		state int
	}{{"", 0}, {"bad/id", 0}, {"exam", -1}, {"exam", 2}} {
		if _, err := s.SetExamState(context.Background(), tc.id, tc.state); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatalf("unsafe input accepted: %+v err=%v", tc, err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
