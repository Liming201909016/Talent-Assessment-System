package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
)

// MT-NEW-DRAFT: actual edit transaction, canonical source/version and create time.
func TestBugManagementTraitsDraftEditNeverLegacy(t *testing.T) {
	for _, code := range []string{"00202", "00101", "00501", "00502"} {
		t.Run(code, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			created := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_exam.*").WillReturnRows(identityHandlerModelRows(t, model.Exam{ID: "exam", AssessmentType: "legacy", ScoringMode: "legacy", CreateTime: &created}))
			mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			draftGateMetadata(mock)
			mock.ExpectQuery("SELECT .*el_mng_exam_draft").WillReturnRows(identityHandlerModelRows(t, draftGateRow()))
			mock.ExpectQuery("SELECT .*el_repo").WillReturnRows(identityHandlerModelRows(t, model.Repo{ID: "new-repo", Code: code, RadioCount: 140}))
			if code != "00202" {
				mock.ExpectRollback()
			} else {
				mock.ExpectExec("UPDATE `el_exam`").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("DELETE FROM `el_exam_repo`").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("DELETE FROM `el_exam_depart`").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("INSERT INTO `el_exam_repo`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("DELETE FROM `el_exam_competency_dimension`").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("UPDATE `el_mng_exam_draft` SET `repo_code`=\\?,`repo_id`=\\?,`updated_at`=\\? WHERE exam_id = \\? AND lifecycle = \\?").WithArgs("00202", "new-repo", sqlmock.AnyArg(), "exam", "draft").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}
			h := &ExamHandler{db: db, cfg: &config.Config{}}
			w := identityHandlerPost(func(c *gin.Context) { c.Set("loginUser", &model.LoginUser{UserID: 1}); h.Save(c) }, `{"id":"exam","title":"Synthetic","joinType":1,"isOpen":1,"totalTime":25,"requiredFields":"name,telephone","repoList":[{"repoId":"new-repo","repoCode":"fake","radioCount":140,"radioScore":5}]}`)
			var result struct {
				Code int `json:"code"`
				Data struct {
					CreateTime time.Time `json:"createTime"`
					Lifecycle  string    `json:"managementTraitsLifecycle"`
				} `json:"data"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil {
				t.Fatal("bad envelope")
			}
			if code == "00202" {
				if result.Code != 0 || result.Data.Lifecycle != "draft" || !result.Data.CreateTime.Equal(created) {
					t.Fatal("draft edit/time contract failed")
				}
			} else if w.Code != 400 || result.Code == 0 {
				t.Fatal("draft switched to legacy")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
