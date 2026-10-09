package handler

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
)

var errDraftFixture = errors.New("synthetic draft insert failure")

func draftHandlerSchema(mock sqlmock.Sqlmock) {
	columns := sqlmock.NewRows([]string{"table_name", "column_name", "column_type", "is_nullable", "character_set_name", "collation_name"})
	for _, name := range []string{"exam_id", "repo_id", "repo_code", "lifecycle", "created_at", "updated_at", "frozen_at"} {
		typ, nullable, charset, collation := "varchar(64)", "NO", "utf8mb4", "utf8mb4_bin"
		if name == "repo_code" {
			typ = "varchar(5)"
		}
		if name == "lifecycle" {
			typ = "varchar(6)"
		}
		if name == "created_at" || name == "updated_at" || name == "frozen_at" {
			typ, charset, collation = "datetime(6)", "", ""
		}
		if name == "frozen_at" {
			nullable = "YES"
		}
		if name == "exam_id" {
			collation = "utf8mb4_0900_ai_ci"
		}
		columns.AddRow("el_mng_exam_draft", name, typ, nullable, charset, collation)
	}
	columns.AddRow("el_exam", "id", "varchar(64)", "NO", "utf8mb4", "utf8mb4_0900_ai_ci")
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name.*information_schema.columns").WillReturnRows(columns)
	mock.ExpectQuery("SELECT .*information_schema.statistics").WillReturnRows(sqlmock.NewRows([]string{"table_name", "index_name", "non_unique", "seq_in_index", "column_name", "prefix_length"}).AddRow("el_mng_exam_draft", "PRIMARY", 0, 1, "exam_id", 0))
	mock.ExpectQuery("SELECT .*information_schema.key_column_usage").WillReturnRows(sqlmock.NewRows([]string{"table_name", "constraint_name", "column_name", "ordinal_position", "referenced_table_name", "referenced_column_name", "update_rule", "delete_rule"}).AddRow("el_mng_exam_draft", "fk_mng_draft_exam", "exam_id", 1, "el_exam", "id", "RESTRICT", "RESTRICT"))
}

// MT-NEW-DRAFT: docs/regression-tests.md. Actual Save, not a source oracle.
func TestBugManagementTraitsNewDraftSave(t *testing.T) {
	for _, mode := range []string{"missing_flag", "true", "false", "missing_schema", "mixed", "insert_failure"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			mock.ExpectBegin()
			rows := identityHandlerModelRows(t, model.Repo{ID: "repo", Code: "00501", RadioCount: 140})
			mock.ExpectQuery("SELECT .*el_repo.*").WillReturnRows(rows)
			if mode == "false" || mode == "mixed" {
				mock.ExpectRollback()
			} else {
				present := sqlmock.NewRows([]string{"table_name", "engine"})
				if mode != "missing_schema" {
					present.AddRow("el_mng_exam_draft", "InnoDB")
				}
				mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(present)
				if mode == "missing_schema" {
					mock.ExpectRollback()
				} else {
					draftHandlerSchema(mock)
					mock.ExpectExec("INSERT INTO `el_exam`").WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec("INSERT INTO `el_exam_repo`").WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec("DELETE FROM `el_exam_competency_dimension`").WillReturnResult(sqlmock.NewResult(0, 0))
					insert := mock.ExpectExec("INSERT INTO `el_mng_exam_draft`")
					if mode == "insert_failure" {
						insert.WillReturnError(errDraftFixture)
						mock.ExpectRollback()
					} else {
						insert.WillReturnResult(sqlmock.NewResult(1, 1))
						mock.ExpectCommit()
					}
				}
			}
			flag := ""
			if mode == "true" {
				flag = `,"managementTraitsTestOnly":true`
			}
			if mode == "false" {
				flag = `,"managementTraitsTestOnly":false`
			}
			repos := `[{"repoId":"repo","repoCode":"00101","radioCount":140,"radioScore":5}]`
			if mode == "mixed" {
				repos = `[{"repoId":"repo","radioCount":140},{"repoId":"other","radioCount":1}]`
			}
			h := &ExamHandler{db: db, cfg: &config.Config{}}
			w := identityHandlerPost(func(c *gin.Context) { c.Set("loginUser", &model.LoginUser{UserID: 1}); h.Save(c) }, `{"title":"Synthetic","assessmentType":"legacy","scoringMode":"legacy","joinType":1,"isOpen":1,"totalTime":25,"requiredFields":"name,telephone","repoList":`+repos+flag+`}`)
			var result struct {
				Code int            `json:"code"`
				Data map[string]any `json:"data"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil {
				t.Fatal("invalid response")
			}
			success := mode == "missing_flag" || mode == "true"
			if success && (result.Code != 0 || result.Data["managementTraitsLifecycle"] != "draft" || result.Data["isManagementTraits"] != true || result.Data["managementTraitsProfileFrozen"] != false) {
				t.Fatal("new005 was not persisted/returned as explicit draft")
			}
			if !success && result.Code == 0 {
				t.Fatal("unsafe new005 accepted")
			}
			if mode == "false" && w.Code != 400 {
				t.Fatal("explicit false must be HTTP400")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// MT-005-ISOLATION supersedes mandatory-new002, not existing frozen contracts.
func TestManagementTraits005Ordinary002Save(t *testing.T) {
	for _, code := range []string{"00201", "00202"} {
		for _, flag := range []string{"", `,"managementTraitsTestOnly":false`, `,"managementTraitsTestOnly":true`} {
			t.Run(code+flag, func(t *testing.T) {
				db, mock := identityHandlerDB(t)
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT .*el_repo").WillReturnRows(identityHandlerModelRows(t, model.Repo{ID: "ordinary", Code: code, RadioCount: 140}))
				if flag == `,"managementTraitsTestOnly":true` {
					mock.ExpectRollback()
				} else {
					mock.ExpectExec("INSERT INTO `el_exam`").WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec("INSERT INTO `el_exam_repo`").WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec("DELETE FROM `el_exam_competency_dimension`").WillReturnResult(sqlmock.NewResult(0, 0))
					mock.ExpectCommit()
				}
				w := identityHandlerPost((&ExamHandler{db: db, cfg: &config.Config{}}).Save, `{"title":"Synthetic legacy","joinType":1,"isOpen":1,"totalTime":35,"requiredFields":"name,idNumber,depart","repoList":[{"repoId":"ordinary","repoCode":"00501","radioCount":140,"radioScore":5}]`+flag+`}`)
				var result struct {
					Code int            `json:"code"`
					Msg  string         `json:"msg"`
					Data map[string]any `json:"data"`
				}
				if json.Unmarshal(w.Body.Bytes(), &result) != nil {
					t.Fatal("invalid envelope")
				}
				if flag == `,"managementTraitsTestOnly":true` {
					if w.Code != 400 || result.Code == 0 {
						t.Fatal("new002 TEST opt-in accepted")
					}
				} else if result.Code != 0 || result.Data["totalTime"] != float64(35) {
					t.Fatal("ordinary002 no longer legacy")
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
