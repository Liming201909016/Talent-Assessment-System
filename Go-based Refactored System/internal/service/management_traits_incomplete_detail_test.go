package service

import (
	"context"
	"testing"
)

func TestManagementTraitsIncompleteAdminDetailWithoutReport(t *testing.T) {
	f, records, paper, profile, owner := managementRuntimeLoadFixture(t, true)
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
	mock.ExpectCommit()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	detail, err := s.ResultDetail(context.Background(), records.Run.ID)
	if err != nil || detail.Result.IsComplete || detail.Result.AnsweredQuestionCount != 139 || detail.Result.OverallScore != nil {
		t.Fatal("incomplete audit detail rejected or official score leaked", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
