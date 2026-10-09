package handler

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
)

// MT-IDENTITY-002: the real LoginForm must not disclose unconfigured legacy fields.
func TestManagementTraitsHTTPConfiguredTesterLogin(t *testing.T) {
	db, mock := identityHandlerDB(t)
	cfg := &config.Config{}
	cfg.Jwt.Secret = "test-only-secret"
	rows := make([]service.ManagementTraitsRuntimeSourceRow, 0, 700)
	labels := []string{"", "不符合", "不太符合", "一般", "比较符合", "很符合"}
	for _, d := range service.ManagementTraitsDimensions() {
		for _, item := range d.Items {
			for raw := 1; raw <= 5; raw++ {
				right := 0
				if (!item.Reverse && raw == 5) || (item.Reverse && raw == 1) {
					right = 1
				}
				rows = append(rows, service.ManagementTraitsRuntimeSourceRow{RelationID: fmt.Sprint("r", item.Number), QuestionID: fmt.Sprint("q", item.Number), Sort: item.Number, RelationType: 1, QuestionType: 1, Code: fmt.Sprint("V", item.Number), Title: fmt.Sprint("synthetic-", item.Number), OptionID: fmt.Sprintf("o%d-%d", item.Number, raw), Raw: raw, IsRight: right, OptionContent: labels[raw]})
			}
		}
	}
	manifest, mapping, err := service.BuildManagementTraitsRuntimeCurrentSource("00201", rows)
	if err != nil {
		t.Fatal(err)
	}
	mc, err := service.CanonicalManagementTraitsManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	pc, err := service.CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil {
		t.Fatal(err)
	}
	frozen := time.Now().Add(-time.Hour).Truncate(time.Second)
	fields, err := json.Marshal(map[string]any{"schema": "mng-candidate-fields-v1", "requiredFields": []string{"age", "degree", "major", "stuFlag"}, "timePolicy": "candidate-only-personal-25-minutes-v1", "source": "current-source-not-client-leader-full-not-historical", "repoCode": "00201", "capturedAt": frozen.Format(time.RFC3339Nano), "manifestSha": mc.SHA256, "mappingSha": pc.SHA256})
	if err != nil {
		t.Fatal(err)
	}
	testerRows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "password", "telephone", "gender", "age", "degree", "major", "stu_flag", "del_flag", "status"}).AddRow("tester", "exam", "", "dormant-name", "test-password", "dormant-phone", "dormant-gender", 31, "degree", "major", 0, 0, "0")
	}
	mock.ExpectQuery("SELECT .*el_tester").WillReturnRows(testerRows())
	identityHandlerSchema(mock, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .*el_tester.*id =").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("tester", "exam", ""))
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	identityHandlerFullParticipantScope(mock, "tester", "tester", false)
	identityHandlerRuntimeSchema(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "is_open", "state", "assessment_type", "scoring_mode"}).AddRow("exam", 2, 0, "legacy", "legacy"))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"exam_id", "bundle_id", "mapping_snapshot", "mapping_sha", "field_contract", "total_time_minutes", "frozen_at", "created_at"}).AddRow("exam", "bundle", string(pc.JSON), pc.SHA256, string(fields), 25, frozen, frozen))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(sqlmock.NewRows([]string{"id", "product_version", "question_version", "scoring_version", "norm_version", "questionnaire", "scoring_manifest", "scoring_manifest_sha", "status"}).AddRow("bundle", manifest.Versions.Product, manifest.Versions.Question, manifest.Versions.Scoring, manifest.Versions.Norm, manifest.Questionnaire, string(mc.JSON), mc.SHA256, "candidate-current-source"))
	mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WillReturnRows(testerRows())
	mock.ExpectCommit()
	w := identityHandlerPost(NewTesterHandler(db, cfg).LoginForm, `{"idNumber":"tester","password":"test-password","examId":"exam"}`)
	var result struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code != 200 || len(result.Data) != 5 || result.Data["name"] != "" || result.Data["paperId"] != "" {
		t.Fatal("configured tester response changed or leaked dormant fields")
	}
	claims, err := service.ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, result.Data["participantToken"].(string), service.ManagementTraitsRuntimeParticipantPurpose, time.Now())
	if err != nil || claims.ValidateBinding("tester", "tester", "exam", "") != nil {
		t.Fatal("incorrect participant token", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
