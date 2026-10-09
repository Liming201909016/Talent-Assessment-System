package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
)

func identityExpectSchema(mock sqlmock.Sqlmock, counts ...int) {
	for i, count := range counts {
		table := []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_result_dimension", "el_mng_result_module"}[i]
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM information_schema.tables WHERE table_schema = DATABASE\\(\\) AND table_name = ").WithArgs(table).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name"}))
}

func TestManagementTraitsIdentityScopeSchema(t *testing.T) {
	for _, tc := range []struct {
		name   string
		counts []int
		bad    bool
	}{
		{"absent", []int{0, 0, 0, 0, 0, 0, 0}, false},
		{"partial", []int{1, 0, 0, 0, 0, 0, 0}, true},
		{"run_without_core", []int{0, 0, 0, 1, 0, 0, 0}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, tc.counts...)
			protected, err := ManagementTraitsIdentityScope(context.Background(), db, "legacy", "candidate", "")
			if protected || (err != nil) != tc.bad {
				t.Fatalf("scope=%v err=%v", protected, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("query_error", func(t *testing.T) {
		db, mock := managementRuntimeDB(t)
		mock.ExpectQuery("information_schema.tables").WillReturnError(errors.New("private schema error"))
		if _, err := ManagementTraitsIdentityScope(context.Background(), db, "exam", "tester", "id"); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatal(err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestManagementTraitsIdentityScopeCrossDestination(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("legacy").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT .*el_candidate.*id =").WithArgs("protected-person", 2).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("protected-person", "new-exam", "paper"))
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("new-exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	identityExpectFullParticipantScope(mock, "candidate", "protected-person", "", false)
	protected, err := ManagementTraitsIdentityScope(context.Background(), db, "legacy", "candidate", "protected-person")
	if err != nil || !protected {
		t.Fatal("cross destination escaped", protected, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func identityExpectAdmission(t *testing.T, mock sqlmock.Sqlmock, exam model.Exam, profile model.ManagementTraitsExamProfile, bundle model.ManagementTraitsDefinitionBundle) {
	t.Helper()
	mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, exam))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, profile))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, bundle))
}

func TestManagementTraitsIdentityCandidateCreate(t *testing.T) {
	f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
	db, mock := managementRuntimeDB(t)
	identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectBegin()
	identityExpectAdmission(t, mock, model.Exam{ID: profile.ExamID, IsOpen: 1, State: 0, AssessmentType: "legacy", ScoringMode: "legacy"}, profile, f.Bundle)
	mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO `el_candidate` \\(`id`,`exam_id`,`name`,`gender`,`telephone`,`affiliation`,`post`,`age`,`degree`,`major`,`stu_flag`,`del_flag`,`create_time`,`update_time`\\)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
	got, handled, err := s.TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ExamID: profile.ExamID, Name: "Person", Telephone: "13800000000"}, "")
	if err != nil || !handled || got.ExamID != profile.ExamID || got.PaperID != "" || got.Name != "Person" {
		t.Fatal(got, handled, err)
	}
	if _, err := uuid.Parse(got.ID); err != nil {
		t.Fatal("not server UUID", err)
	}
	claims, err := ParseManagementTraitsRuntimeToken("test-only", got.ParticipantToken, ManagementTraitsRuntimeParticipantPurpose, time.Now())
	if err != nil || claims.ValidateBinding("candidate", got.ID, got.ExamID, "") != nil {
		t.Fatal("wrong token", err)
	}
	b, _ := json.Marshal(got)
	var fields map[string]any
	if json.Unmarshal(b, &fields) != nil || len(fields) != 5 {
		t.Fatal("response leaked fields", string(b))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityCandidateExisting(t *testing.T) {
	for _, authorized := range []bool{false, true} {
		t.Run(map[bool]string{false: "anonymous_reject", true: "token_restore_zero_writes"}[authorized], func(t *testing.T) {
			f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			identityExpectAdmission(t, mock, model.Exam{ID: profile.ExamID, IsOpen: 1, AssessmentType: "legacy", ScoringMode: "legacy"}, profile, f.Bundle)
			mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "telephone", "del_flag"}).AddRow("existing", profile.ExamID, "", "Person", "13800000000", 0))
			token := ""
			if authorized {
				token, _ = CreateManagementTraitsRuntimeToken("test-only", ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: "existing", ExamID: profile.ExamID, ExpiresAt: time.Now().Add(time.Hour).Unix()}, time.Now())
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ExamID: profile.ExamID, Name: "Person", Telephone: "13800000000"}, token)
			if !handled || (err == nil) != authorized || (authorized && got.ID != "existing") {
				t.Fatal(got, handled, err)
			}
			if authorized {
				original, err := ParseManagementTraitsRuntimeToken("test-only", token, ManagementTraitsRuntimeParticipantPurpose, time.Now())
				if err != nil {
					t.Fatal(err)
				}
				restored, err := ParseManagementTraitsRuntimeToken("test-only", got.ParticipantToken, ManagementTraitsRuntimeParticipantPurpose, time.Now())
				if err != nil || original.ExpiresAt != restored.ExpiresAt {
					t.Fatal("restore refreshed token expiry", err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityCandidateUnsupported(t *testing.T) {
	f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
	for _, req := range []ManagementTraitsCandidateIdentityRequest{{Age: new(int)}, {Degree: new(string)}, {Major: new(string)}, {StuFlag: new(int)}} {
		db, mock := managementRuntimeDB(t)
		req.ExamID, req.Name, req.Telephone = profile.ExamID, "Person", "13800000000"
		identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
		mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		managementRuntimeExpectSchema(t, mock)
		mock.ExpectBegin()
		identityExpectAdmission(t, mock, model.Exam{ID: profile.ExamID, IsOpen: 1, AssessmentType: "legacy", ScoringMode: "legacy"}, profile, f.Bundle)
		mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectRollback()
		_, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024).TryRegisterCandidateIdentity(context.Background(), req, "")
		if !handled || err == nil {
			t.Fatal("unsupported identity accepted")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagementTraitsIdentityAdmissionRejects(t *testing.T) {
	for _, tc := range []string{"disabled", "ready", "overdue", "closed", "future", "expired", "profile_missing", "bundle_failure", "invalid_bundle", "missing_phone_contract", "unconfigured_gender"} {
		t.Run(tc, func(t *testing.T) {
			f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
			exam := model.Exam{ID: profile.ExamID, IsOpen: 1, AssessmentType: "legacy", ScoringMode: "legacy"}
			req := ManagementTraitsCandidateIdentityRequest{ExamID: profile.ExamID, Name: "Person", Telephone: "13800000000"}
			now := time.Now()
			switch tc {
			case "disabled":
				exam.State = 1
			case "ready":
				exam.State = 2
			case "overdue":
				exam.State = 3
			case "closed":
				exam.IsOpen = 2
			case "future":
				at := now.Add(time.Hour)
				exam.StartTime = &at
			case "expired":
				at := now.Add(-time.Hour)
				exam.EndTime = &at
			case "invalid_bundle":
				f.Bundle.Status = "retired"
			case "missing_phone_contract":
				var wire managementTraitsRuntimeFieldWire
				if json.Unmarshal([]byte(profile.FieldContract), &wire) != nil {
					t.Fatal("fixture")
				}
				wire.RequiredFields = []string{"name"}
				b, _ := json.Marshal(wire)
				profile.FieldContract = string(b)
			case "unconfigured_gender":
				req.Gender = "x"
			}
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, exam))
			if tc == "profile_missing" {
				mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
			} else if tc == "bundle_failure" {
				mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, profile))
				mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnError(errors.New("bundle error"))
			} else if tc != "closed" {
				mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, profile))
				mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
			}
			if tc != "closed" && tc != "profile_missing" && tc != "bundle_failure" {
				mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}
			mock.ExpectRollback()
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024).TryRegisterCandidateIdentity(context.Background(), req, "")
			if !handled || !errors.Is(err, ErrManagementTraitsRuntimeInvalid) || got.ID != "" || got.ParticipantToken != "" {
				t.Fatal(got, handled, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityTesterZeroWrites(t *testing.T) {
	for _, tc := range []string{"success", "password", "cross_exam", "deleted", "query_error", "status_disabled", "status_nil", "status_empty", "status_unknown", "status_padded", "status_double_zero", "status_negative"} {
		t.Run(tc, func(t *testing.T) {
			f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT .*el_tester.*id =").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}).AddRow("tester", profile.ExamID))
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			identityExpectFullParticipantScope(mock, "tester", "tester", "", false)
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			identityExpectAdmission(t, mock, model.Exam{ID: profile.ExamID, IsOpen: 2, State: 0, AssessmentType: "legacy", ScoringMode: "legacy"}, profile, f.Bundle)
			zero, phone := 0, "13800000000"
			status := "0"
			te := model.Tester{ID: "tester", ExamID: &profile.ExamID, Name: "Person", Telephone: &phone, DelFlag: &zero, Status: &status, Password: "test-only-password"}
			switch tc {
			case "password":
				te.Password = "different"
			case "cross_exam":
				other := "other"
				te.ExamID = &other
			case "deleted":
				one := 1
				te.DelFlag = &one
			case "status_disabled":
				status = "1"
			case "status_nil":
				te.Status = nil
			case "status_empty":
				status = ""
			case "status_unknown":
				status = "2"
			case "status_padded":
				status = "0 "
			case "status_double_zero":
				status = "00"
			case "status_negative":
				status = "-1"
			}
			if tc == "query_error" {
				mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WillReturnError(errors.New("tester failure"))
			} else {
				mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, te))
			}
			if tc == "success" {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024).TryTesterIdentity(context.Background(), profile.ExamID, "tester", "test-only-password")
			if !handled || (err == nil) != (tc == "success") {
				t.Fatal(got, handled, err)
			}
			if tc == "success" {
				claims, err := ParseManagementTraitsRuntimeToken("test-only", got.ParticipantToken, ManagementTraitsRuntimeParticipantPurpose, time.Now())
				if err != nil || claims.ValidateBinding("tester", "tester", profile.ExamID, "") != nil {
					t.Fatal(err)
				}
			} else if got.ID != "" || got.ParticipantToken != "" {
				t.Fatal("failed issuance leaked response")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityScopeOptionalProtection(t *testing.T) {
	for _, tc := range []string{"snapshot", "run", "capture", "capture_failure", "missing_owner"} {
		t.Run(tc, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			for i := 0; i < 7; i++ {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			}
			tables := sqlmock.NewRows([]string{"table_name"})
			if tc == "capture" || tc == "capture_failure" {
				tables.AddRow("el_mng_legacy_pdf_capture")
			}
			mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(tables)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			owners := sqlmock.NewRows([]string{"id", "exam_id", "paper_id"})
			if tc != "missing_owner" {
				owners.AddRow("person", "old-exam", "paper")
			}
			mock.ExpectQuery("SELECT .*el_candidate.*id =").WillReturnRows(owners)
			if tc != "missing_owner" {
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			}
			snapshotCount := 0
			if tc == "snapshot" || tc == "missing_owner" {
				snapshotCount = 1
			}
			mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(snapshotCount))
			if tc == "capture" {
				mock.ExpectQuery("SELECT count.*el_mng_legacy_pdf_capture").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			}
			if tc == "capture_failure" {
				mock.ExpectQuery("SELECT count.*el_mng_legacy_pdf_capture").WillReturnError(errors.New("missing column"))
			} else {
				count := 0
				if tc == "run" {
					count = 1
				}
				mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
				capture := ""
				if tc == "capture" {
					capture = "el_mng_legacy_pdf_capture"
				}
				identityExpectFullParticipantScope(mock, "candidate", "person", capture, false)
			}
			got, err := ManagementTraitsIdentityScope(context.Background(), db, "legacy", "candidate", "person")
			if tc == "capture_failure" {
				if err == nil {
					t.Fatal("failed probe allowed legacy")
				}
			} else if err != nil || !got {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityCandidatePhoneCrossDestination(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT .*el_candidate.*telephone =").WithArgs("legacy", "13800000000", 1).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("protected", "legacy", "paper"))
	identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT .*el_candidate.*id =").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("protected", "legacy", "paper"))
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	identityExpectFullParticipantScope(mock, "candidate", "protected", "", false)
	_, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ExamID: "legacy", Name: "Person", Telephone: "13800000000"}, "")
	if !handled || err == nil {
		t.Fatal("protected phone fell through to legacy", handled, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityCandidateRestoreFailures(t *testing.T) {
	for _, tc := range []string{"paper_purpose", "other_id", "tester_kind", "other_exam", "changed_name", "deleted", "null_deleted", "ambiguous", "lookup_error", "insert_error", "commit_error", "expired_after_lock"} {
		t.Run(tc, func(t *testing.T) {
			f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			identityExpectAdmission(t, mock, model.Exam{ID: profile.ExamID, IsOpen: 1, AssessmentType: "legacy", ScoringMode: "legacy"}, profile, f.Bundle)
			owner := managementTraitsRuntimeOwner{ID: "existing", ExamID: profile.ExamID, Name: "Person", Telephone: "13800000000"}
			claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: owner.ID, ExamID: profile.ExamID, ExpiresAt: time.Now().Add(time.Hour).Unix()}
			switch tc {
			case "paper_purpose":
				claims.Purpose, claims.PaperID = ManagementTraitsRuntimePaperPurpose, "paper"
			case "other_id":
				claims.ParticipantID = "other"
			case "tester_kind":
				claims.ParticipantType = "tester"
			case "other_exam":
				claims.ExamID = "other"
			case "changed_name":
				owner.Name = "Other"
			case "deleted":
				owner.DelFlag = 1
			case "null_deleted":
				owner.DelFlag = -1
			case "expired_after_lock":
				claims.ExpiresAt = time.Now().Unix() + 1
			}
			token, err := CreateManagementTraitsRuntimeToken("test-only", claims, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			expect := mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE")
			switch tc {
			case "ambiguous":
				expect.WillReturnRows(managementRuntimeModelRows(t, owner, owner))
			case "lookup_error":
				expect.WillReturnError(errors.New("lookup error"))
			case "insert_error":
				expect.WillReturnRows(sqlmock.NewRows([]string{"id"}))
				token = ""
				mock.ExpectExec("INSERT INTO `el_candidate`").WillReturnError(errors.New("insert error"))
			case "expired_after_lock":
				expect.WillDelayFor(1100 * time.Millisecond).WillReturnRows(managementRuntimeModelRows(t, owner))
			default:
				expect.WillReturnRows(managementRuntimeModelRows(t, owner))
			}
			if tc == "commit_error" {
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			} else {
				mock.ExpectRollback()
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ExamID: profile.ExamID, Name: "Person", Telephone: "13800000000"}, token)
			if !handled || err == nil || got.ID != "" || got.ParticipantToken != "" {
				t.Fatal("invalid restore escaped", got, handled, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityScopeEmptyMetadataFailsClosed(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"count"}))
	if _, err := ManagementTraitsIdentityScope(context.Background(), db, "exam", "candidate", ""); err == nil {
		t.Fatal("missing count row allowed legacy")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityCandidateExplicitID(t *testing.T) {
	for _, tc := range []struct {
		id     string
		exists bool
	}{{"supplied-id", false}, {"supplied-id", true}, {"0", false}, {"0", true}} {
		exists := tc.exists
		t.Run(tc.id+"_"+map[bool]string{false: "supplied_new_id_rejected", true: "exact_id_restore"}[exists], func(t *testing.T) {
			f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			owner := managementTraitsRuntimeOwner{ID: tc.id, ExamID: profile.ExamID, Name: "Person", Telephone: "13800000000"}
			rows := sqlmock.NewRows([]string{"id"})
			if exists {
				rows = managementRuntimeModelRows(t, owner)
			}
			mock.ExpectQuery("SELECT .*el_candidate.*id =").WithArgs(owner.ID, 2).WillReturnRows(rows)
			if exists {
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			}
			mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WithArgs("candidate", owner.ID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("SELECT count.*el_mng_result_run").WithArgs("candidate", owner.ID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			identityExpectFullParticipantScope(mock, "candidate", owner.ID, "", false)
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			identityExpectAdmission(t, mock, model.Exam{ID: profile.ExamID, IsOpen: 1, AssessmentType: "legacy", ScoringMode: "legacy"}, profile, f.Bundle)
			rows = sqlmock.NewRows([]string{"id"})
			if exists {
				rows = managementRuntimeModelRows(t, owner)
			}
			if exists {
				mock.ExpectQuery("SELECT .*el_candidate.*WHERE id = \\? LIMIT \\? FOR UPDATE").WithArgs(owner.ID, 2).WillReturnRows(rows)
			} else {
				mock.ExpectQuery("SELECT .*el_candidate.*WHERE \\(exam_id = \\? AND telephone = \\?\\) OR id = \\? LIMIT \\? FOR UPDATE").WithArgs(profile.ExamID, owner.Telephone, owner.ID, 2).WillReturnRows(rows)
			}
			token := ""
			if exists {
				token, _ = CreateManagementTraitsRuntimeToken("test-only", ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: owner.ID, ExamID: profile.ExamID, ExpiresAt: time.Now().Add(time.Hour).Unix()}, time.Now())
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ID: owner.ID, ExamID: profile.ExamID, Name: owner.Name, Telephone: owner.Telephone}, token)
			if !handled || (err == nil) != exists || (exists && got.ID != owner.ID) {
				t.Fatal(got, handled, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentitySchemaThreeCoreFailsClosed(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	identityExpectSchema(mock, 1, 1, 1, 0, 0, 0, 0)
	_, err := probeManagementTraitsIdentitySchema(context.Background(), db)
	if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("three-table partial schema admitted identity", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityLegacyFallback(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	identityExpectSchema(mock, 0, 0, 0, 0, 0, 0, 0)
	got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ExamID: "legacy", Telephone: "13800000000"}, "")
	if handled || err != nil || got.ID != "" {
		t.Fatal("fully absent schema did not allow legacy", got, handled, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentityLegacyPhoneIsExamScoped(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("legacy").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// Match the exact row the old Save writes; unrelated exams sharing a
	// telephone must not alter legacy registration semantics.
	mock.ExpectQuery("SELECT .*el_candidate.*exam_id = \\? AND telephone = \\?.*ORDER BY id LIMIT \\?").WithArgs("legacy", "13800000000", 1).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
	_, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ExamID: "legacy", Telephone: "13800000000"}, "")
	if handled || err != nil {
		t.Fatal("legacy phone fallback changed", handled, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsIdentitySchemaSevenTableMatrix(t *testing.T) {
	for mask := 0; mask < 1<<7; mask++ {
		t.Run(fmt.Sprintf("tables_%07b", mask), func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			counts := make([]int, 7)
			for i := range counts {
				counts[i] = (mask >> i) & 1
			}
			identityExpectSchema(mock, counts...)
			schema, err := probeManagementTraitsIdentitySchema(context.Background(), db)
			if mask == 0 || mask == (1<<7)-1 {
				if err != nil || schema.present != (mask != 0) || schema.run != (mask != 0) {
					t.Fatal("absent/full schema rejected", schema, err)
				}
			} else if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) || schema.present || schema.run {
				t.Fatal("partial schema accepted", schema, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func identityExpectFullParticipantScope(mock sqlmock.Sqlmock, kind, id, marker string, queryFailure bool) {
	tables := sqlmock.NewRows([]string{"table_name"})
	for _, name := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_paper_question_snapshot", "el_mng_result_run", "el_mng_result_dimension", "el_mng_result_module"} {
		tables.AddRow(name)
	}
	if marker != "" {
		tables.AddRow(marker)
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(tables)
	if marker != "" {
		mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns").WithArgs(marker).WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow(marker, "participant_type").AddRow(marker, "participant_id"))
	}
	// A severed legacy paper link must not erase independent participant evidence.
	mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_" + kind + " WHERE").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow(id, nil, nil))
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM "+table+" WHERE").WithArgs(kind, id).WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	}
	if marker != "" {
		q := mock.ExpectQuery("SELECT 1 AS protected FROM "+marker+" WHERE").WithArgs(kind, id)
		if queryFailure {
			q.WillReturnError(errors.New("PRIVATE_MARKER_DATABASE_SECRET"))
		} else {
			q.WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
		}
	}
}

func TestManagementTraitsIdentityMarkerOnlyProtected(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		for _, mode := range []string{"scope", "direct", "marker_query_error"} {
			t.Run(kind+"_"+mode, func(t *testing.T) {
				db, mock := managementRuntimeDB(t)
				identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("legacy").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT .*el_"+kind+".*id =").WithArgs("protected", 2).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("protected", "old", ""))
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("old").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WithArgs(kind, "protected").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT count.*el_mng_result_run").WithArgs(kind, "protected").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				identityExpectFullParticipantScope(mock, kind, "protected", "el_mng_legacy_protection", mode == "marker_query_error")
				if mode == "scope" || mode == "marker_query_error" {
					protected, err := ManagementTraitsIdentityScope(context.Background(), db, "legacy", kind, "protected")
					if mode == "marker_query_error" {
						if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) || err.Error() == "PRIVATE_MARKER_DATABASE_SECRET" {
							t.Fatal("marker failure was not sanitized", protected, err)
						}
					} else if err != nil || !protected {
						t.Fatal("marker-only participant escaped", protected, err)
					}
				} else {
					managementRuntimeExpectSchema(t, mock)
					mock.ExpectBegin()
					mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "is_open", "state", "assessment_type", "scoring_mode"}).AddRow("legacy", map[string]int{"candidate": 1, "tester": 2}[kind], 0, "legacy", "legacy"))
					mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
					mock.ExpectRollback()
					s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
					var handled bool
					var err error
					if kind == "candidate" {
						_, handled, err = s.TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ID: "protected", ExamID: "legacy", Name: "Person", Telephone: "13800000000"}, "")
					} else {
						_, handled, err = s.TryTesterIdentity(context.Background(), "legacy", "protected", "test-only-password")
					}
					if !handled || !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
						t.Fatal("marker-only direct call allowed legacy DML", handled, err)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsIdentityMarkerPhoneAndTokenNeverFallBack(t *testing.T) {
	for _, mode := range []string{"phone_lookup", "same_paper_token"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("legacy").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			req := ManagementTraitsCandidateIdentityRequest{ExamID: "legacy", Name: "Person", Telephone: "13800000000"}
			if mode == "phone_lookup" {
				mock.ExpectQuery("SELECT .*el_candidate.*telephone =").WithArgs("legacy", req.Telephone, 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("protected"))
				identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
				mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("legacy").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			} else {
				req.ID = "protected"
			}
			mock.ExpectQuery("SELECT .*el_candidate.*id =").WithArgs("protected", 2).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("protected", "legacy", ""))
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("legacy").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WithArgs("candidate", "protected").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("SELECT count.*el_mng_result_run").WithArgs("candidate", "protected").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			identityExpectFullParticipantScope(mock, "candidate", "protected", "el_mng_legacy_protection", false)
			token := ""
			if mode == "same_paper_token" {
				var err error
				token, err = CreateManagementTraitsRuntimeToken("test-only", ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: "protected", ExamID: "legacy", ExpiresAt: time.Now().Add(time.Hour).Unix()}, time.Now())
				if err != nil {
					t.Fatal(err)
				}
				managementRuntimeExpectSchema(t, mock)
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "is_open", "state", "assessment_type", "scoring_mode"}).AddRow("legacy", 1, 0, "legacy", "legacy"))
				mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
				mock.ExpectRollback()
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).TryRegisterCandidateIdentity(context.Background(), req, token)
			if !handled || !errors.Is(err, ErrManagementTraitsRuntimeInvalid) || got.ID != "" || got.ParticipantToken != "" {
				t.Fatal("protected lookup/restore escaped", got, handled, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityAbsentCoreExistingParticipant(t *testing.T) {
	for _, marker := range []bool{false, true} {
		t.Run(fmt.Sprint("marker_", marker), func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 0, 0, 0, 0, 0, 0, 0)
			tables := sqlmock.NewRows([]string{"table_name"})
			if marker {
				tables.AddRow("el_mng_legacy_protection")
			}
			mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(tables)
			if marker {
				mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns").WithArgs("el_mng_legacy_protection").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_legacy_protection", "candidate_id"))
				mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE").WithArgs("protected").WillReturnRows(sqlmock.NewRows([]string{"id"}))
				mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_legacy_protection WHERE").WithArgs("protected").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ID: "protected", ExamID: "legacy"}, "")
			if handled != marker || (err != nil) != marker || got.ID != "" {
				t.Fatal("absent-core compatibility/protection incorrect", got, handled, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsIdentityHistoricalSiblingPrecision(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("legacy").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT .*el_candidate.*id =").WithArgs("person", 2).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("person", "old-exam", ""))
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("old-exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WithArgs("candidate", "person").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT count.*el_mng_result_run").WithArgs("candidate", "person").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	tables := sqlmock.NewRows([]string{"table_name"})
	for _, name := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_paper_question_snapshot", "el_mng_result_run", "el_mng_result_dimension", "el_mng_result_module", "el_mng_legacy_protection"} {
		tables.AddRow(name)
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(tables)
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns").WithArgs("el_mng_legacy_protection").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_legacy_protection", "exam_id").AddRow("el_mng_legacy_protection", "participant_type").AddRow("el_mng_legacy_protection", "participant_id"))
	mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE \\(id = \\?\\)$").WithArgs("person").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("person", "old-exam", nil))
	mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE \\(exam_id = \\?\\) LIMIT 1").WithArgs("old-exam").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run", "el_mng_legacy_protection"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM "+table+" WHERE \\(participant_type = \\? AND participant_id = \\?\\) LIMIT 1").WithArgs("candidate", "person").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	}
	protected, err := ManagementTraitsIdentityScope(context.Background(), db, "legacy", "candidate", "person")
	if err != nil || protected {
		t.Fatal("unrelated historical sibling widened identity scope", protected, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
