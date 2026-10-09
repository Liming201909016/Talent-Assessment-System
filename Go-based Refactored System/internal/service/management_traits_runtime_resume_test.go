package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/talent-assessment/refactored/internal/model"
)

func managementRuntimeExpectSchema(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	contract, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	metadata := managementTraitsSchemaFixture(contract)
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(managementRuntimeModelRows(t, managementTraitsSchemaTableRows(contract)...))
	cols := make([]any, len(metadata.Columns))
	for i := range cols {
		cols[i] = metadata.Columns[i]
	}
	idx := make([]any, len(metadata.Indexes))
	for i := range idx {
		idx[i] = metadata.Indexes[i]
	}
	fks := make([]any, len(metadata.ForeignKeys))
	for i := range fks {
		fks[i] = metadata.ForeignKeys[i]
	}
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type").WillReturnRows(managementRuntimeModelRows(t, cols...))
	mock.ExpectQuery("SELECT table_name AS table_name, index_name AS index_name, non_unique AS non_unique").WillReturnRows(managementRuntimeModelRows(t, idx...))
	mock.ExpectQuery("SELECT k.table_name AS table_name, k.constraint_name AS constraint_name").WillReturnRows(managementRuntimeModelRows(t, fks...))
}

func TestManagementTraitsPublicCreateResumePreservesDeadlineAfterExamEnd(t *testing.T) {
	f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	// Existing paper can be read without reapplying the exam window or retired
	// new-admission policy. The original timestamp remains immutable.
	paper.State = 1
	paper.UserTime = 0
	owner.EndTime = nil
	f.Bundle.Status = "retired"
	for i := range f.Questions {
		f.Questions[i].SubmittedAt = nil
	}
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectBegin()
	end := time.Now().Add(-time.Hour)
	mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: paper.ExamID, State: 3, EndTime: &end}))
	mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, owner))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, paper))
	mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
	mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WillReturnRows(managementRuntimeModelRows(t, owner))
	questions := make([]any, len(f.Questions))
	legacy := make([]any, len(f.Questions))
	for i, q := range f.Questions {
		questions[i] = q
		legacy[i] = model.PaperQu{ID: q.PaperQuestionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, QuType: 1, Sort: q.DisplayOrder, Answered: 1, ActualScore: *q.RawAnswer}
	}
	mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(managementRuntimeModelRows(t, questions...))
	mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(managementRuntimeModelRows(t, legacy...))
	managementRuntimeExpectBuckets(t, mock, f.Questions)
	mock.ExpectCommit()
	s := NewManagementTraitsRuntimeService(db, "test-only-secret", 1<<20)
	c := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: owner.ExamID, ExpiresAt: time.Now().Add(time.Minute).Unix()}
	detail, err := s.CreatePaper(context.Background(), c)
	if err != nil || detail.PaperID != paper.ID || !detail.Deadline.Equal(*f.Paper.LimitTime) || !detail.ReminderAt.Equal(f.Paper.StartedAt.Add(20*time.Minute)) || len(detail.Questions) != 140 || detail.Answered != 140 {
		t.Fatal("resume changed frozen facts", err)
	}
	parsed, err := ParseManagementTraitsRuntimeToken("test-only-secret", detail.PaperToken, ManagementTraitsRuntimePaperPurpose, time.Now())
	if err != nil || parsed.ValidateBinding(owner.Kind, owner.ID, paper.ExamID, paper.ID) != nil {
		t.Fatal("wrong paper purpose/binding", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
