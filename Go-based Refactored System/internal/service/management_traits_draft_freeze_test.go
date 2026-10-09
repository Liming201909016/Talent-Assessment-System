package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

func draftFreezeSchema(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}).AddRow("el_mng_exam_draft", "InnoDB"))
	columns := sqlmock.NewRows([]string{"table_name", "column_name", "column_type", "is_nullable", "character_set_name", "collation_name"})
	for name, typ := range map[string]string{"exam_id": "varchar(64)", "repo_id": "varchar(64)", "repo_code": "varchar(5)", "lifecycle": "varchar(6)", "created_at": "datetime(6)", "updated_at": "datetime(6)", "frozen_at": "datetime(6)"} {
		nullable, cs, coll := "NO", "utf8mb4", "utf8mb4_bin"
		if name == "frozen_at" {
			nullable = "YES"
		}
		if typ == "datetime(6)" {
			cs, coll = "", ""
		}
		columns.AddRow("el_mng_exam_draft", name, typ, nullable, cs, coll)
	}
	columns.AddRow("el_exam", "id", "varchar(64)", "NO", "utf8mb4", "utf8mb4_bin")
	mock.ExpectQuery("SELECT .*information_schema.columns").WillReturnRows(columns)
	mock.ExpectQuery("SELECT .*information_schema.statistics").WillReturnRows(sqlmock.NewRows([]string{"table_name", "index_name", "non_unique", "seq_in_index", "column_name", "prefix_length"}).AddRow("el_mng_exam_draft", "PRIMARY", 0, 1, "exam_id", 0))
	mock.ExpectQuery("SELECT .*information_schema.key_column_usage").WillReturnRows(sqlmock.NewRows([]string{"table_name", "constraint_name", "column_name", "ordinal_position", "referenced_table_name", "referenced_column_name", "update_rule", "delete_rule"}).AddRow("el_mng_exam_draft", "fk_mng_draft_exam", "exam_id", 1, "el_exam", "id", "RESTRICT", "RESTRICT"))
}

// MT-NEW-DRAFT: public FreezeProfile uses real source builder and transaction.
func TestBugManagementTraitsDraftFreezeAtomic(t *testing.T) {
	for _, mode := range []string{"commit", "marker_update_failure", "source_mismatch", "00501_commit", "00502_commit"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			if err := s.CheckRuntimeSchema(context.Background()); err != nil {
				t.Fatal(err)
			}
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: "exam", AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, RequiredFields: "name,telephone"}))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
			draftFreezeSchema(mock)
			now := time.Now().Add(-time.Minute)
			actualCode, repoID := "00202", "repo"
			if strings.HasPrefix(mode, "005") {
				actualCode, repoID = mode[:5], "synthetic-"+mode[:5]+"-repo"
			}
			code := actualCode
			if mode == "source_mismatch" {
				code = "00201"
			}
			mock.ExpectQuery("SELECT .*el_mng_exam_draft.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.ManagementTraitsExamDraft{ExamID: "exam", RepoID: repoID, RepoCode: code, Lifecycle: "draft", CreatedAt: now, UpdatedAt: now}))
			mock.ExpectQuery("SELECT count.*el_paper").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("SELECT .*el_exam_repo").WillReturnRows(managementRuntimeModelRows(t, model.ExamRepo{ID: "link", ExamID: "exam", RepoID: repoID, RadioCount: 140}))
			repo := model.Repo{ID: repoID, Code: actualCode, RadioCount: 140}
			mock.ExpectQuery("SELECT .*el_repo").WillReturnRows(managementRuntimeModelRows(t, repo))
			if mode == "source_mismatch" {
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery("SELECT .*el_repo").WillReturnRows(managementRuntimeModelRows(t, repo))
				source := managementRuntimeSourceRows()
				if strings.HasPrefix(actualCode, "005") {
					for i := range source {
						source[i].RelationID = actualCode + "-" + source[i].RelationID
						source[i].QuestionID = actualCode + "-" + source[i].QuestionID
						source[i].OptionID = actualCode + "-" + source[i].OptionID
					}
				}
				values := make([]any, len(source))
				for i := range source {
					values[i] = source[i]
				}
				mock.ExpectQuery("SELECT .* FROM el_qu_repo qr LEFT JOIN el_qu q.*LEFT JOIN el_qu_answer qa").WithArgs(repoID).WillReturnRows(managementRuntimeModelRows(t, values...))
				mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(sqlmock.NewRows([]string{"id"}))
				mock.ExpectExec("INSERT INTO `el_mng_definition_bundle`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO `el_mng_exam_profile`").WillReturnResult(sqlmock.NewResult(1, 1))
				update := mock.ExpectExec("UPDATE `el_mng_exam_draft` SET .* WHERE exam_id = \\? AND lifecycle = \\?")
				if mode == "marker_update_failure" {
					update.WillReturnError(errors.New("synthetic marker failure"))
					mock.ExpectRollback()
				} else {
					update.WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()
				}
			}
			p, err := s.FreezeProfile(context.Background(), "exam")
			if strings.HasSuffix(mode, "commit") {
				if err != nil || p.ExamID != "exam" || p.FrozenAt == nil {
					t.Fatal("draft freeze not committed", err)
				}
			} else if err == nil || p.ExamID != "" {
				t.Fatal("failed draft freeze returned success")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
