package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

// MT-IDENTITY-002: parent owns the regression and business-branch ledgers.
func TestBugManagementTraitsConfiguredCandidateWithoutNamePhone(t *testing.T) {
	f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
	var fields managementTraitsRuntimeFieldWire
	_ = json.Unmarshal([]byte(profile.FieldContract), &fields)
	fields.RequiredFields = []string{"age", "degree", "major", "stuFlag"}
	b, _ := json.Marshal(fields)
	profile.FieldContract = string(b)
	db, mock := managementRuntimeDB(t)
	identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectBegin()
	identityExpectAdmission(t, mock, model.Exam{ID: profile.ExamID, IsOpen: 1, AssessmentType: "legacy", ScoringMode: "legacy"}, profile, f.Bundle)
	mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO `el_candidate`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	age, flag, degree, major := 31, 0, "degree", "major"
	got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ExamID: profile.ExamID, Age: &age, StuFlag: &flag, Degree: &degree, Major: &major}, "")
	if err != nil || !handled || got.ID == "" {
		t.Fatalf("configured-only create rejected: handled=%v err=%v", handled, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugManagementTraitsTesterEndedRetiredFullFrozenResume(t *testing.T) {
	for _, status := range []string{"retired", "review-revoked"} {
		t.Run(status, func(t *testing.T) {
			f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
			owner.Kind = "tester"
			owner.Status = "0"
			owner.EndTime = nil
			paper.State = 1
			f.Paper.ParticipantType = "tester"
			f.Bundle.Status = status
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT .*el_tester.*id =").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow(owner.ID, owner.ExamID, paper.ID))
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			identityExpectFullParticipantScope(mock, "tester", owner.ID, "", false)
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			end := time.Now().Add(-time.Hour)
			mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: owner.ExamID, IsOpen: 2, State: 3, EndTime: &end, AssessmentType: "legacy", ScoringMode: "legacy"}))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, profile))
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
			zero, enabled, password, phone := 0, "0", "test-password", owner.Telephone
			te := model.Tester{ID: owner.ID, ExamID: &owner.ExamID, PaperID: &paper.ID, Name: owner.Name, Telephone: &phone, DelFlag: &zero, Status: &enabled, Password: password}
			mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, te))
			mock.ExpectCommit()
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, paper))
			mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
			mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WillReturnRows(managementRuntimeModelRows(t, owner))
			qs, legacy := make([]any, 140), make([]any, 140)
			for i, q := range f.Questions {
				qs[i] = q
				legacy[i] = model.PaperQu{ID: q.PaperQuestionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, QuType: 1, Sort: q.DisplayOrder, Answered: 1, ActualScore: *q.RawAnswer}
			}
			mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(managementRuntimeModelRows(t, qs...))
			mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(managementRuntimeModelRows(t, legacy...))
			if status == "retired" {
				managementRuntimeExpectBuckets(t, mock, f.Questions)
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).TryTesterIdentity(context.Background(), owner.ExamID, owner.ID, password)
			if !handled || (err == nil) != (status == "retired") || (err == nil && got.PaperID != paper.ID) {
				t.Fatalf("resume status=%s handled=%v err=%v", status, handled, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugManagementTraitsExtendedFrozenIdentityContract(t *testing.T) {
	age, flag, degree, major := 31, 0, "degree", "major"
	identity := managementTraitsRuntimeIdentity{Age: &age, StuFlag: &flag, Degree: &degree, Major: &major}
	fields := managementTraitsRuntimeFieldWire{RequiredFields: []string{"age", "degree", "major", "stuFlag"}}
	if !managementTraitsIdentityLegal(fields, identity) {
		t.Fatal("configured extended fields rejected")
	}
	base := `{"name":"","gender":"","telephone":"","affiliation":"","post":""}`
	for _, raw := range []string{base, strings.TrimSuffix(base, "}") + `,"age":31,"degree":"degree","major":"major","stuFlag":0}`} {
		var got managementTraitsRuntimeIdentity
		if err := managementTraitsDecodeStrict([]byte(raw), 1<<20, &got); err != nil {
			t.Fatal("compatible frozen decode", err)
		}
	}
	for _, suffix := range []string{`,"age":null}`, `,"age":31,"age":32}`, `,"age":31,"\u0061ge":32}`, `,"Age":31}`, `,"age":31.0}`, `,"age":3.1e1}`, `,"age":"31"}`, `,"age":9223372036854775808}`, `,"degree":null}`, `,"stuFlag":false}`, `,"unknown":0}`} {
		var got managementTraitsRuntimeIdentity
		if managementTraitsDecodeStrict([]byte(strings.TrimSuffix(base, "}")+suffix), 1<<20, &got) == nil {
			t.Fatal("invalid frozen identity accepted", suffix)
		}
	}
	for _, n := range []int{-1, 0, 2147483648} {
		bad := identity
		bad.Age = &n
		if managementTraitsIdentityLegal(fields, bad) {
			t.Fatal("invalid age accepted", n)
		}
	}
	for _, n := range []int{-1, 2} {
		bad := identity
		bad.StuFlag = &n
		if managementTraitsIdentityLegal(fields, bad) {
			t.Fatal("invalid flag accepted", n)
		}
	}
	for _, bad := range []managementTraitsRuntimeIdentity{{Age: &age, Degree: &degree, Major: &major}, {StuFlag: &flag, Degree: &degree, Major: &major}, {Age: &age, StuFlag: &flag, Degree: new(string), Major: &major}, {Age: &age, StuFlag: &flag, Degree: &degree, Major: new(string)}} {
		if managementTraitsIdentityLegal(fields, bad) {
			t.Fatal("missing/empty configured field accepted")
		}
	}
	for _, keys := range [][]string{{}, {"age", "age"}, {"unknown"}} {
		if managementTraitsIdentityLegal(managementTraitsRuntimeFieldWire{RequiredFields: keys}, identity) {
			t.Fatal("invalid field contract accepted", keys)
		}
	}
}

func TestBugManagementTraitsFrozenIdentityIgnoresCurrentProfile(t *testing.T) {
	f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	profile.BundleID = "today-bundle"
	profile.FieldContract = "changed draft"
	profile.MappingSnapshot = "today mapping"
	owner.Name = "changed current name"
	owner.Telephone = "changed current telephone"
	if managementTraitsRuntimeMetadata(f.Bundle, f.Paper, profile, paper, owner, 1<<20) != nil {
		t.Fatal("frozen paper trusted current mutable profile/identity")
	}
	owner.Kind = "tester"
	owner.Status = "1"
	if validateManagementTraitsRuntimeOwner([]managementTraitsRuntimeOwner{owner}, "tester", owner.ID, owner.ExamID, owner.PaperID) == nil {
		t.Fatal("disabled audit owner admitted")
	}
}

func TestManagementTraitsConfiguredNewFreezeProjectsDormantFields(t *testing.T) {
	f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	age, flag, text := 31, 0, "legacy dormant"
	owner.PaperID = ""
	owner.EndTime = nil
	owner.Kind = "tester"
	owner.Status = "0"
	owner.Age = &age
	owner.StuFlag = &flag
	owner.Degree = &text
	owner.Major = &text
	owner.Gender = text
	owner.Post = text
	owner.Affiliation = text
	var fields managementTraitsRuntimeFieldWire
	if json.Unmarshal([]byte(profile.FieldContract), &fields) != nil {
		t.Fatal("fixture")
	}
	for _, configured := range [][]string{{"name"}, {"telephone"}, {"age"}, {"degree"}, {"major"}, {"stuFlag"}, {"name", "gender", "telephone", "affiliation", "post", "age", "degree", "major", "stuFlag"}} {
		fields.RequiredFields = configured
		raw, _ := json.Marshal(fields)
		profile.FieldContract = string(raw)
		m, err := DecodeManagementTraitsManifest([]byte(f.Bundle.ScoringManifest), 1<<20)
		if err != nil {
			t.Fatal(err)
		}
		mapping, err := DecodeManagementTraitsMapping([]byte(profile.MappingSnapshot), m, 1<<20)
		if err != nil {
			t.Fatal(err)
		}
		order := make([]string, 140)
		for i, q := range mapping.Questions {
			order[i] = q.SourceQuestionID
		}
		r, err := buildManagementTraitsRuntimePaper(profile, f.Bundle, model.Exam{ID: profile.ExamID, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25}, owner, paper.ID, *paper.CreateTime, order, 1<<20)
		if err != nil {
			t.Fatal("configured freeze rejected", configured, err)
		}
		var frozen managementTraitsRuntimeIdentity
		if managementTraitsDecodeStrict([]byte(r.Snapshot.ParticipantSnapshot), 1<<20, &frozen) != nil || !managementTraitsIdentityLegal(fields, frozen) {
			t.Fatal("invalid projected freeze", configured)
		}
		expected, _ := json.Marshal(managementTraitsProjectIdentity(fields, managementTraitsOwnerIdentity(owner)))
		if r.Snapshot.ParticipantSnapshot != string(expected) {
			t.Fatal("dormant fields leaked", r.Snapshot.ParticipantSnapshot)
		}
	}
}

func TestManagementTraitsCandidateAuthenticatedFrozenResume(t *testing.T) {
	for _, tc := range []string{"retired_changed_profile", "corrupt_questions", "ambiguous_owner", "disabled_tester_owner", "corrupt_buckets"} {
		t.Run(tc, func(t *testing.T) {
			f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
			owner.EndTime = nil
			paper.State = 1
			f.Bundle.Status = "retired"
			currentProfile := profile
			currentProfile.BundleID = "today-bundle"
			currentProfile.FieldContract = "today draft"
			currentBundle := f.Bundle
			currentBundle.ID = currentProfile.BundleID
			currentBundle.Status = "review-revoked"
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT .*el_candidate.*id =").WillReturnRows(managementRuntimeModelRows(t, owner))
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			identityExpectFullParticipantScope(mock, "candidate", owner.ID, "", false)
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			end := time.Now().Add(-time.Hour)
			identityExpectAdmission(t, mock, model.Exam{ID: owner.ExamID, IsOpen: 1, State: 3, EndTime: &end, AssessmentType: "legacy", ScoringMode: "legacy"}, currentProfile, currentBundle)
			mutableOwner := owner
			mutableOwner.Name = "changed current name"
			mutableOwner.Telephone = "changed current phone"
			// A valid token selects its owner only; incoming telephone can belong to a sibling.
			mock.ExpectQuery("SELECT .*el_candidate.*WHERE id = \\? LIMIT \\? FOR UPDATE").WithArgs(owner.ID, 2).WillReturnRows(managementRuntimeModelRows(t, mutableOwner))
			mock.ExpectCommit()
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, paper))
			mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, currentProfile))
			owners := managementRuntimeModelRows(t, mutableOwner)
			if tc == "ambiguous_owner" {
				other := owner
				other.Kind = "tester"
				other.Status = "0"
				owners = managementRuntimeModelRows(t, mutableOwner, other)
			}
			if tc == "disabled_tester_owner" {
				bad := owner
				bad.Kind = "tester"
				bad.Status = "1"
				owners = managementRuntimeModelRows(t, bad)
			}
			mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WillReturnRows(owners)
			if tc != "ambiguous_owner" && tc != "disabled_tester_owner" {
				if tc == "corrupt_questions" {
					f.Questions[0].Content = "tampered"
				}
				qs, legacy := make([]any, 140), make([]any, 140)
				for i, q := range f.Questions {
					qs[i] = q
					legacy[i] = model.PaperQu{ID: q.PaperQuestionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, QuType: 1, Sort: q.DisplayOrder, Answered: 1, ActualScore: *q.RawAnswer}
				}
				mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(managementRuntimeModelRows(t, qs...))
				mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(managementRuntimeModelRows(t, legacy...))
				if tc == "corrupt_buckets" {
					mock.ExpectQuery("SELECT .*el_paper_qu_answer").WillReturnRows(sqlmock.NewRows([]string{"id"}))
				} else if tc != "corrupt_questions" {
					managementRuntimeExpectBuckets(t, mock, f.Questions)
				}
			}
			if tc == "retired_changed_profile" {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			expiry := time.Now().Add(time.Hour).Unix()
			token, err := CreateManagementTraitsRuntimeToken("test-only", ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: owner.ID, ExamID: owner.ExamID, ExpiresAt: expiry}, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).TryRegisterCandidateIdentity(context.Background(), ManagementTraitsCandidateIdentityRequest{ID: owner.ID, ExamID: owner.ExamID, Name: "ignored changed incoming", Telephone: "ignored incoming phone"}, token)
			if !handled || (err == nil) != (tc == "retired_changed_profile") {
				t.Fatal("resume contract", tc, handled, err)
			}
			if err == nil {
				c, e := ParseManagementTraitsRuntimeToken("test-only", got.ParticipantToken, ManagementTraitsRuntimeParticipantPurpose, time.Now())
				if e != nil || c.ExpiresAt != expiry || got.Name != owner.Name || got.PaperID != paper.ID {
					t.Fatal("frozen identity/binding/expiry changed", got, e)
				}
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestBugManagementTraitsPaperDetailExpiryDuringLock(t *testing.T) {
	f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillDelayFor(1100 * time.Millisecond).WillReturnRows(managementRuntimeModelRows(t, paper))
	mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
	mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WillReturnRows(managementRuntimeModelRows(t, owner))
	qs, legacy := make([]any, 140), make([]any, 140)
	for i, q := range f.Questions {
		qs[i] = q
		legacy[i] = model.PaperQu{ID: q.PaperQuestionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, QuType: 1, Sort: q.DisplayOrder, Answered: 1, ActualScore: *q.RawAnswer}
	}
	mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(managementRuntimeModelRows(t, qs...))
	mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(managementRuntimeModelRows(t, legacy...))
	managementRuntimeExpectBuckets(t, mock, f.Questions)
	mock.ExpectRollback()
	got, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).PaperDetail(context.Background(), ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: owner.ExamID, PaperID: paper.ID, ExpiresAt: time.Now().Unix() + 1})
	if err == nil || got.PaperToken != "" {
		t.Fatal("expired credential renewed during paper lock")
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestManagementTraitsConfiguredTesterNewIdentity(t *testing.T) {
	for _, status := range []string{"candidate-current-source", "retired", "review-revoked"} {
		t.Run(status, func(t *testing.T) {
			f, _, _, profile, owner := managementRuntimeLoadFixture(t, false)
			f.Bundle.Status = status
			var fields managementTraitsRuntimeFieldWire
			_ = json.Unmarshal([]byte(profile.FieldContract), &fields)
			fields.RequiredFields = []string{"age", "degree", "major", "stuFlag"}
			b, _ := json.Marshal(fields)
			profile.FieldContract = string(b)
			db, mock := managementRuntimeDB(t)
			identityExpectSchema(mock, 1, 1, 1, 1, 1, 1, 1)
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT .*el_tester.*id =").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}).AddRow(owner.ID, owner.ExamID))
			mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("SELECT count.*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			identityExpectFullParticipantScope(mock, "tester", owner.ID, "", false)
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			identityExpectAdmission(t, mock, model.Exam{ID: owner.ExamID, IsOpen: 2, AssessmentType: "legacy", ScoringMode: "legacy"}, profile, f.Bundle)
			zero, enabled, age, degree, major := 0, "0", 31, "degree", "major"
			te := model.Tester{ID: owner.ID, ExamID: &owner.ExamID, Name: "dormant unconfigured name", Gender: &degree, Affiliation: &major, Age: &age, Degree: &degree, Major: &major, StuFlag: &zero, Status: &enabled, DelFlag: &zero, Password: "test-password"}
			mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, te))
			if status == "candidate-current-source" {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			got, handled, err := NewManagementTraitsRuntimeService(db, "test-only", 1<<20).TryTesterIdentity(context.Background(), owner.ExamID, owner.ID, "test-password")
			if !handled || (err == nil) != (status == "candidate-current-source") || (err == nil && got.ID != owner.ID) {
				t.Fatal("configured new tester contract", handled, err)
			}
			if err == nil && got.Name != "" {
				t.Fatal("unconfigured dormant tester name leaked", got.Name)
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestManagementTraitsConfiguredReportUsesFrozenExtendedIdentity(t *testing.T) {
	f, r, _, _, _ := managementRuntimeLoadFixture(t, false)
	age, flag, degree, major := 31, 0, "degree", "major"
	var fields managementTraitsRuntimeFieldWire
	_ = json.Unmarshal([]byte(f.Paper.FieldContract), &fields)
	fields.RequiredFields = []string{"age", "degree", "major", "stuFlag"}
	contract, _ := json.Marshal(fields)
	f.Paper.FieldContract = string(contract)
	frozen, _ := json.Marshal(managementTraitsRuntimeIdentity{Age: &age, Degree: &degree, Major: &major, StuFlag: &flag})
	f.Paper.ParticipantSnapshot = string(frozen)
	content, err := LoadManagementTraitsTestContent(managementTestWorkbook(t))
	if err != nil {
		t.Fatal(err)
	}
	got, err := BuildManagementTraitsTestReport(f.Bundle, f.Paper, f.Questions, r.Run, r.Dimensions, r.Modules, r.Receipt, content, 1<<20)
	if err != nil || got.Participant.Age == nil || *got.Participant.Age != age || got.Participant.StuFlag == nil || *got.Participant.StuFlag != 0 || got.Participant.Name != "" || got.Participant.Telephone != "" || got.Participant.Degree == nil || *got.Participant.Degree != degree || got.Participant.Major == nil || *got.Participant.Major != major {
		t.Fatal("report not using configured frozen identity", got.Participant, err)
	}
}
