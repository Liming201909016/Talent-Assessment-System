package service

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm/schema"
)

func managementRuntimeModelRows(t *testing.T, values ...any) *sqlmock.Rows {
	t.Helper()
	s, err := schema.Parse(values[0], &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	rows := sqlmock.NewRows(s.DBNames)
	for _, value := range values {
		v := reflect.ValueOf(value)
		data := make([]driver.Value, len(s.DBNames))
		for i, key := range s.DBNames {
			x, _ := s.FieldsByDBName[key].ValueOf(context.Background(), v)
			if reflect.ValueOf(x).Kind() == reflect.Ptr && reflect.ValueOf(x).IsNil() {
				data[i] = nil
				continue
			}
			converted, e := driver.DefaultParameterConverter.ConvertValue(x)
			if e != nil {
				t.Fatal(e)
			}
			data[i] = converted
		}
		rows.AddRow(data...)
	}
	return rows
}

func managementRuntimeLoadFixture(t *testing.T, incomplete bool) (managementTraitsResultValidationFixture, managementTraitsRuntimeRecords, model.Paper, model.ManagementTraitsExamProfile, managementTraitsRuntimeOwner) {
	t.Helper()
	f, start := managementRuntimeResultFixture(t, incomplete)
	code, captured := "00201", start.Add(-time.Minute)
	contract := managementTraitsRuntimeFieldContract{Schema: "mng-candidate-fields-v1", RequiredFields: []string{"name", "telephone"}, TimePolicy: "candidate-only-personal-25-minutes-v1", Source: "current-source-not-client-leader-full-not-historical", RepoCode: code, CapturedAt: captured, ManifestSHA: f.Bundle.ScoringManifestSHA, MappingSHA: f.Paper.MappingSHA}
	c, err := managementTraitsCanonicalBytes(contract)
	if err != nil {
		t.Fatal(err)
	}
	f.Paper.FieldContract = string(c.JSON)
	profileID := f.Paper.ExamID
	f.Paper.ProfileExamID, f.Paper.Source, f.Paper.IdentitySource = &profileID, "new_creation", "submitted_snapshot"
	f.Paper.SourceCapturedAt = captured
	owner := managementTraitsRuntimeOwner{Kind: "candidate", ID: f.Paper.ParticipantID, ExamID: f.Paper.ExamID, PaperID: f.Paper.PaperID, Name: "test-person", Telephone: "13800000000"}
	identity := managementTraitsRuntimeIdentity{Name: owner.Name, Gender: "", Telephone: owner.Telephone, Affiliation: "", Post: ""}
	ic, _ := managementTraitsCanonicalBytes(identity)
	f.Paper.ParticipantSnapshot = string(ic.JSON)
	evidence := managementTraitsRuntimeEvidence{Schema: "mng-current-source-evidence-v1", Source: contract.Source, RepoCode: code, CapturedAt: captured, ManifestSHA: contract.ManifestSHA, MappingSHA: contract.MappingSHA}
	ec, _ := managementTraitsCanonicalBytes(evidence)
	f.Paper.EvidenceSnapshot, f.Paper.EvidenceSHA = string(ec.JSON), ec.SHA256
	at := start.Add(5 * time.Minute)
	if incomplete {
		at = *f.Paper.LimitTime
	}
	records, err := buildManagementTraitsRuntimeRecords(f.Bundle, f.Paper, f.Questions, "opaque-run", at, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	for i := range f.Questions {
		f.Questions[i].SubmittedAt = &at
		f.Questions[i].CreatedAt = start
	}
	owner.EndTime = &at
	paper := model.Paper{ID: f.Paper.PaperID, ExamID: f.Paper.ExamID, UserID: "101", State: 2, TotalTime: 25, UserTime: *records.Run.UserTimeSeconds, CreateTime: &start, LimitTime: f.Paper.LimitTime}
	profile := model.ManagementTraitsExamProfile{ExamID: f.Paper.ExamID, BundleID: f.Bundle.ID, MappingSnapshot: f.Paper.MappingSnapshot, MappingSHA: f.Paper.MappingSHA, FieldContract: f.Paper.FieldContract, TotalTimeMinutes: 25, FrozenAt: &captured, CreatedAt: captured}
	return f, records, paper, profile, owner
}

func managementRuntimeExpectLoad(t *testing.T, mock sqlmock.Sqlmock, f managementTraitsResultValidationFixture, records managementTraitsRuntimeRecords, paper model.Paper, profile model.ManagementTraitsExamProfile, owner managementTraitsRuntimeOwner) {
	t.Helper()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_mng_result_run.*WHERE id = ").WillReturnRows(managementRuntimeModelRows(t, records.Run))
	mock.ExpectQuery("SELECT .*el_paper.*WHERE id = .*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, paper))
	mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
	mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WillReturnRows(managementRuntimeModelRows(t, owner))
	questions := make([]any, len(f.Questions))
	for i := range f.Questions {
		questions[i] = f.Questions[i]
	}
	mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(managementRuntimeModelRows(t, questions...))
	pq := make([]any, len(f.Questions))
	for i, q := range f.Questions {
		p := model.PaperQu{ID: q.PaperQuestionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, Sort: q.DisplayOrder, QuType: 1}
		if q.RawAnswer != nil {
			p.Answered, p.ActualScore = 1, *q.RawAnswer
		}
		pq[i] = p
	}
	mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(managementRuntimeModelRows(t, pq...))
	ds := make([]any, len(records.Dimensions))
	for i := range records.Dimensions {
		ds[i] = records.Dimensions[i]
	}
	mock.ExpectQuery("SELECT .*el_mng_result_dimension").WillReturnRows(managementRuntimeModelRows(t, ds...))
	ms := make([]any, len(records.Modules))
	for i := range records.Modules {
		ms[i] = records.Modules[i]
	}
	mock.ExpectQuery("SELECT .*el_mng_result_module").WillReturnRows(managementRuntimeModelRows(t, ms...))
	mock.ExpectQuery("SELECT .*el_mng_runtime_receipt").WillReturnRows(managementRuntimeModelRows(t, records.Receipt))
	managementRuntimeExpectBuckets(t, mock, f.Questions)
}

func managementRuntimeExpectBuckets(t *testing.T, mock sqlmock.Sqlmock, questions []model.ManagementTraitsPaperQuestionSnapshot) {
	t.Helper()
	values := make([]any, 0, 700)
	for _, q := range questions {
		var options []ManagementTraitsMappedOption
		if managementTraitsDecodeStrict([]byte(q.OptionsSnapshot), 1024*1024, &options) != nil {
			t.Fatal("invalid fixture")
		}
		for _, o := range options {
			checked := int8(0)
			if q.SelectedOptionID != nil && *q.SelectedOptionID == o.SourceOptionID {
				checked = 1
			}
			values = append(values, model.PaperQuAnswer{ID: "bucket-" + o.SourceOptionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, AnswerID: o.SourceOptionID, Checked: checked, Score: o.Raw, Sort: o.DisplayOrder})
		}
	}
	mock.ExpectQuery("SELECT .*el_paper_qu_answer").WillReturnRows(managementRuntimeModelRows(t, values...))
}

func TestManagementTraitsRuntimeLoadValidatedRunZeroWrites(t *testing.T) {
	f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
	mock.ExpectCommit()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
	got, err := s.LoadValidatedRun(context.Background(), records.Run.ID)
	if err != nil || got.RunID != records.Run.ID || got.Result.OverallScore.RatString() != "50" {
		t.Fatal("validated read rejected", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsRuntimeLoadRejectsDBFailure(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnError(errors.New("private DSN credentials"))
	mock.ExpectRollback()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1024*1024)
	if _, err := s.LoadValidatedRun(context.Background(), "run"); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("DB error leaked", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsRuntimeIncompleteRetryValidation(t *testing.T) {
	f, records, _, _, _ := managementRuntimeLoadFixture(t, true)
	if _, err := validateManagementTraitsRuntimeResult(f.Bundle, f.Paper, f.Questions, records, 1024*1024); err != nil {
		t.Fatal(err)
	}
	records.Dimensions[0].Score = managementTraitsResultValidationDecimal(ManagementTraitsDimensions()[0].Norm)
	if _, err := validateManagementTraitsRuntimeResult(f.Bundle, f.Paper, f.Questions, records, 1024*1024); err == nil {
		t.Fatal("incomplete official value accepted")
	}
}
