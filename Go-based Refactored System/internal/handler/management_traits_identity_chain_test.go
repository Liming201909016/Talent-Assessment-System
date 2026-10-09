package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
)

// MT-IDENTITY-002: exercise the real CandidateHandler.Save with Gin/sqlmock.
func TestBugManagementTraitsHTTPConfiguredTypedEmptyStrings(t *testing.T) {
	db, mock := identityHandlerDB(t)
	identityHandlerSchema(mock, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// Decoding must accept Java-compatible empty typed values and reach the
	// service prerequisite, not reject them as unknown JSON keys.
	identityHandlerSchema(mock, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnError(errors.New("schema unavailable"))
	cfg := &config.Config{}
	cfg.Jwt.Secret = "test-only-secret"
	w := identityHandlerPost(NewCandidateHandler(db, cfg).Save, `{"examId":"exam","age":"","stuFlag":"","degree":"","major":""}`)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("typed empty strings never reached schema prerequisite: %v; response=%s", err, w.Body.String())
	}
}

func TestBugManagementTraitsIdentityConfiguredBodyContract(t *testing.T) {
	for _, raw := range []string{`{"examId":"exam"}`, `{"examId":"exam","age":"","stuFlag":"","degree":"","major":""}`, `{"examId":"exam","age":31,"stuFlag":0,"degree":"degree","major":"major"}`, `{"examId":"exam","age":"31","stuFlag":"0"}`} {
		got, ok := managementTraitsIdentityCandidateBody([]byte(raw))
		if !ok || got.ExamID != "exam" {
			t.Fatalf("valid configured body rejected: %s", raw)
		}
		if strings.Contains(raw, `"age":""`) && (got.Age != nil || got.StuFlag != nil || got.Degree != nil || got.Major != nil) {
			t.Fatal("empty typed values not normalized")
		}
	}
	for _, extra := range []string{`"age":null`, `"age":false`, `"age":31.5`, `"age":3e1`, `"age":"31.5"`, `"age":9223372036854775808`, `"stuFlag":[]`, `"degree":null`, `"major":1`, `"age":31,"age":32`, `"age":31,"\u0061ge":32`, `"Age":31`, `"paperId":"paper"`} {
		if _, ok := managementTraitsIdentityCandidateBody([]byte(`{"examId":"exam",` + extra + `}`)); ok {
			t.Fatal("invalid typed body accepted", extra)
		}
	}
}

func identityHandlerConfiguredAdmission(t *testing.T, mock sqlmock.Sqlmock, required []string) {
	t.Helper()
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
	fields, err := json.Marshal(map[string]any{"schema": "mng-candidate-fields-v1", "requiredFields": required, "timePolicy": "candidate-only-personal-25-minutes-v1", "source": "current-source-not-client-leader-full-not-historical", "repoCode": "00201", "capturedAt": frozen.Format(time.RFC3339Nano), "manifestSha": mc.SHA256, "mappingSha": pc.SHA256})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "is_open", "state", "assessment_type", "scoring_mode"}).AddRow("exam", 1, 0, "legacy", "legacy"))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"exam_id", "bundle_id", "mapping_snapshot", "mapping_sha", "field_contract", "total_time_minutes", "frozen_at", "created_at"}).AddRow("exam", "bundle", string(pc.JSON), pc.SHA256, string(fields), 25, frozen, frozen))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(sqlmock.NewRows([]string{"id", "product_version", "question_version", "scoring_version", "norm_version", "questionnaire", "scoring_manifest", "scoring_manifest_sha", "status"}).AddRow("bundle", manifest.Versions.Product, manifest.Versions.Question, manifest.Versions.Scoring, manifest.Versions.Norm, manifest.Questionnaire, string(mc.JSON), mc.SHA256, "candidate-current-source"))
}

func TestBugManagementTraitsHTTPCandidateConfiguredCreate(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		fields     []string
		success    bool
	}{
		{"name_only", `{"examId":"exam","name":"Person","age":"","stuFlag":"","degree":"","major":""}`, []string{"name"}, true},
		{"extended_only", `{"examId":"exam","age":31,"degree":"degree","major":"major","stuFlag":0}`, []string{"age", "degree", "major", "stuFlag"}, true},
		{"required_age_empty", `{"examId":"exam","age":""}`, []string{"age"}, false},
		{"unconfigured_nonempty", `{"examId":"exam","name":"Person","degree":"unexpected"}`, []string{"name"}, false},
		{"configured_name_mobile", `{"examId":"exam","name":"Synthetic","telephone":"13800000000"}`, []string{"name", "telephone"}, true},
		{"configured_gender", `{"examId":"exam","name":"Synthetic","telephone":"13800000000","gender":"0"}`, []string{"name", "telephone", "gender"}, true},
		{"unconfigured_gender", `{"examId":"exam","name":"Synthetic","telephone":"13800000000","gender":"0"}`, []string{"name", "telephone"}, false},
		{"configured_gender_empty", `{"examId":"exam","name":"Synthetic","telephone":"13800000000","gender":""}`, []string{"name", "telephone", "gender"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := identityHandlerDB(t)
			cfg := &config.Config{}
			cfg.Jwt.Secret = "test-only-secret"
			for i := 0; i < 2; i++ {
				identityHandlerSchema(mock, 1)
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			}
			identityHandlerRuntimeSchema(t, mock)
			mock.ExpectBegin()
			identityHandlerConfiguredAdmission(t, mock, tc.fields)
			mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			if tc.success {
				mock.ExpectExec("INSERT INTO `el_candidate`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			w := identityHandlerPost(NewCandidateHandler(db, cfg).Save, tc.body)
			var result struct {
				Code int            `json:"code"`
				Data map[string]any `json:"data"`
			}
			if json.Unmarshal(w.Body.Bytes(), &result) != nil || (result.Code == 0) != tc.success {
				t.Fatal(w.Body.String())
			}
			if tc.success && len(result.Data) != 5 {
				t.Fatal("identity response signature changed", w.Body.String())
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestBugManagementTraitsHTTPInvalidBodyNeverResumesAfterScopeChange(t *testing.T) {
	db, mock := identityHandlerDB(t)
	cfg := &config.Config{}
	cfg.Jwt.Secret = "test-only-secret"
	identityHandlerSchema(mock, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	identityHandlerSchema(mock, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// Newly protected invalid input must stop before schema admission or any transaction.
	token, err := service.CreateManagementTraitsRuntimeToken(cfg.Jwt.Secret, service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: "existing", ExamID: "exam", ExpiresAt: time.Now().Add(time.Hour).Unix()}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.POST("/identity", NewCandidateHandler(db, cfg).Save)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/identity", strings.NewReader(`{"examId":"exam","name":"Person","telephone":"13800000000","paperId":"injected"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Management-Traits-Token", token)
	r.ServeHTTP(w, req)
	var result struct {
		Code int                                      `json:"code"`
		Data service.ManagementTraitsIdentityResponse `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Code == 0 || result.Data.ParticipantToken != "" {
		t.Fatal("strict-invalid body resumed after scope changed", w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
