package handler

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
)

// MT-CANDIDATE-FIELDS: docs/regression-tests.md. Real Detail and Save; no real DB/PII.
func TestBugManagementTraitsCandidatePublicDetail(t *testing.T) {
	for _, kind := range []string{"frozen", "draft", "legacy", "absent_schema", "metadata_failure", "bad_profile", "non002", "repo_lookup_failure"} {
		t.Run(kind, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			mock.ExpectQuery("SELECT .*el_exam.*").WillReturnRows(identityHandlerModelRows(t, model.Exam{ID: "exam", RequiredFields: "name,gender,telephone", AssessmentType: "legacy", ScoringMode: "legacy"}))
			mock.ExpectQuery("SELECT .*el_exam_repo.*").WillReturnRows(identityHandlerModelRows(t, model.ExamRepo{ExamID: "exam", RepoID: "repo"}))
			code := "00201"
			if kind == "non002" {
				code = "00101"
			}
			mock.ExpectQuery("SELECT .*el_repo.*").WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow("repo", code))
			mock.ExpectQuery("SELECT .*el_exam_depart.*").WillReturnRows(sqlmock.NewRows([]string{"depart_id"}))
			if kind == "repo_lookup_failure" {
				mock.ExpectQuery("SELECT .*el_exam_repo AS er.*").WillReturnError(errors.New("synthetic lookup failure"))
			} else {
				mock.ExpectQuery("SELECT .*el_exam_repo AS er.*").WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow(code))
			}
			if kind == "metadata_failure" {
				mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WillReturnError(errors.New("synthetic failure"))
			} else if kind == "draft" {
				draftGateIdentityProbe(mock)
				mock.ExpectQuery("SELECT .*el_mng_exam_draft").WillReturnRows(identityHandlerModelRows(t, draftGateRow()))
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			} else if kind != "non002" && kind != "repo_lookup_failure" {
				present := 1
				if kind == "absent_schema" {
					present = 0
				}
				identityHandlerSchema(mock, present)
				if present == 1 {
					count := 1
					if kind == "legacy" {
						count = 0
					}
					mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
					if count == 1 {
						identityHandlerRuntimeSchema(t, mock)
						bundle, profile, _, _ := identityHandlerSource(t)
						if kind == "bad_profile" {
							profile.FieldContract = "{}"
						}
						mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(identityHandlerModelRows(t, profile))
						mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(identityHandlerModelRows(t, bundle))
					}
				}
			}
			w := identityHandlerPost((&ExamHandler{db: db, cfg: &config.Config{}}).Detail, `{"id":"exam"}`)
			var result struct {
				Code int            `json:"code"`
				Data map[string]any `json:"data"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil {
				t.Fatal("invalid envelope")
			}
			if kind == "metadata_failure" || kind == "bad_profile" || kind == "repo_lookup_failure" {
				if result.Code == 0 || result.Data != nil {
					t.Fatal("failed metadata/profile fell back to legacy")
				}
			} else {
				if result.Code != 0 {
					t.Fatal("public detail rejected")
				}
				if kind == "frozen" {
					if result.Data["isManagementTraits"] != true || result.Data["managementTraitsLifecycle"] != "frozen" || result.Data["managementTraitsProfileFrozen"] != true || result.Data["requiredFields"] != "name,telephone" {
						t.Fatal("public detail did not project frozen fields")
					}
				} else if kind == "draft" {
					if result.Data["isManagementTraits"] != true || result.Data["managementTraitsLifecycle"] != "draft" || result.Data["managementTraitsProfileFrozen"] != false {
						t.Fatal("public draft misclassified as legacy")
					}
				} else {
					value, exists := result.Data["managementTraitsProfileFrozen"]
					frozen, isBool := value.(bool)
					if !exists || !isBool || frozen || result.Data["requiredFields"] != "name,gender,telephone" {
						t.Fatal("legacy detail must include an explicit boolean false frozen marker and unchanged fields")
					}
					if result.Data["isManagementTraits"] != false || result.Data["managementTraitsLifecycle"] != "legacy" {
						t.Fatal("legacy lifecycle changed")
					}
				}
				for _, key := range []string{"fieldContract", "mappingSnapshot", "bundleId", "participantToken", "password", "telephone", "name"} {
					if _, ok := result.Data[key]; ok {
						t.Fatal("public detail exposed private metadata")
					}
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugManagementTraitsCandidateStrictRejectedKeys(t *testing.T) {
	for _, extra := range []string{`"idNumber":null`, `"depart":""`, `"grade":""`, `"sex":"0"`, `"mobile":"13800000000"`} {
		t.Run(strings.Split(extra, ":")[0], func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			identityHandlerSchema(mock, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			w := identityHandlerPost(NewCandidateHandler(db, &config.Config{}).Save, `{"examId":"exam","name":"Synthetic","telephone":"13800000000",`+extra+`}`)
			var result struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code == 0 || result.Msg != "参数错误：管理特质仅允许配置的身份字段" {
				t.Fatal("strict body not rejected")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
