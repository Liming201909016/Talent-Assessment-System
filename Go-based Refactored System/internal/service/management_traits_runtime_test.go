package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	jwt "github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func managementRuntimeDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	g, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	return g, mock
}

func TestManagementTraitsRuntimeTokenStrict(t *testing.T) {
	now := time.Date(2026, 10, 2, 8, 0, 0, 0, time.UTC)
	c := ManagementTraitsRuntimeClaims{Purpose: "management_traits_participant", ParticipantType: "candidate", ParticipantID: "opaque-participant", ExamID: "opaque-exam", ExpiresAt: now.Add(time.Hour).Unix()}
	token, err := CreateManagementTraitsRuntimeToken("test-only-secret", c, now)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseManagementTraitsRuntimeToken("test-only-secret", token, c.Purpose, now)
	if err != nil || !reflect.DeepEqual(c, parsed) {
		t.Fatal("valid token rejected", err)
	}
	if parsed.ValidateBinding("candidate", c.ParticipantID, c.ExamID, "other-paper") == nil {
		t.Fatal("cross paper accepted")
	}
	for _, exp := range []any{nil, "1790000000", float64(c.ExpiresAt) + 0.5, json.Number("1e1000"), now.Unix(), now.Unix() - 1} {
		t.Run(fmt.Sprint(exp), func(t *testing.T) {
			m := jwt.MapClaims{"purpose": c.Purpose, "participant_type": c.ParticipantType, "participant_id": c.ParticipantID, "exam_id": c.ExamID, "paper_id": "", "exp": exp}
			signed, e := jwt.NewWithClaims(jwt.SigningMethodHS512, m).SignedString([]byte("test-only-secret"))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = ParseManagementTraitsRuntimeToken("test-only-secret", signed, c.Purpose, now); e == nil {
				t.Fatal("invalid exp accepted")
			}
		})
	}
	for _, method := range []jwt.SigningMethod{jwt.SigningMethodHS256, jwt.SigningMethodHS384} {
		signed, _ := jwt.NewWithClaims(method, jwt.MapClaims{"purpose": c.Purpose, "participant_type": "candidate", "participant_id": c.ParticipantID, "exam_id": c.ExamID, "paper_id": "", "exp": c.ExpiresAt}).SignedString([]byte("test-only-secret"))
		if _, e := ParseManagementTraitsRuntimeToken("test-only-secret", signed, c.Purpose, now); e == nil {
			t.Fatal("wrong alg accepted")
		}
	}
	for _, purpose := range []string{"competency_participant", "competency_paper", "system", "management_traits_paper"} {
		if _, e := ParseManagementTraitsRuntimeToken("test-only-secret", token, purpose, now); e == nil {
			t.Fatal("wrong purpose accepted")
		}
	}
	for _, id := range []string{"", " padded", "bad\nID", string(make([]byte, 65))} {
		bad := c
		bad.ParticipantID = id
		if _, e := CreateManagementTraitsRuntimeToken("test-only-secret", bad, now); e == nil {
			t.Fatal("invalid id accepted")
		}
	}
}

func managementRuntimeSourceRows() []ManagementTraitsRuntimeSourceRow {
	rows := make([]ManagementTraitsRuntimeSourceRow, 0, 700)
	texts := []string{"", "不符合", "不太符合", "一般", "比较符合", "很符合"}
	for _, d := range ManagementTraitsDimensions() {
		for _, item := range d.Items {
			for raw := 1; raw <= 5; raw++ {
				right := 0
				if (!item.Reverse && raw == 5) || (item.Reverse && raw == 1) {
					right = 1
				}
				rows = append(rows, ManagementTraitsRuntimeSourceRow{RelationID: fmt.Sprint("r", item.Number), QuestionID: fmt.Sprint("q", item.Number), Sort: item.Number, RelationType: 1, QuestionType: 1, Code: fmt.Sprint("V", item.Number), Title: fmt.Sprint("current-source-", item.Number), OptionID: fmt.Sprintf("o%d-%d", item.Number, raw), Raw: raw, IsRight: right, OptionContent: texts[raw]})
			}
		}
	}
	return rows
}

func TestManagementTraitsRuntimeCurrentSource(t *testing.T) {
	rows := managementRuntimeSourceRows()
	before := append([]ManagementTraitsRuntimeSourceRow(nil), rows...)
	for _, code := range []string{"00201", "00202"} {
		m, mapping, err := BuildManagementTraitsRuntimeCurrentSource(code, rows)
		if err != nil || len(m.Questions) != 140 || len(mapping.Questions) != 140 {
			t.Fatal(err)
		}
		if m.Questions[66].Content != "current-source-67" || m.Questions[95].Content != "current-source-96" {
			t.Fatal("source text replaced")
		}
		if code == "00202" && m.Versions.Question != "mng-leader-db-current-v1" {
			t.Fatal("leader version misrepresented")
		}
		if !reflect.DeepEqual(rows, before) {
			t.Fatal("source mutated")
		}
	}
	mutations := []func([]ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow{
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow { return r[:699] },
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow { return append(r, r[0]) },
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow {
			r[0].QuestionID = ""
			return r
		},
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow { r[0].Raw = 0; return r },
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow {
			r[0].Sort = 141
			return r
		},
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow {
			r[0].IsRight = 1 - r[0].IsRight
			return r
		},
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow {
			r[0].OptionContent = "changed"
			return r
		},
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow {
			r[0].Title = "changed"
			return r
		},
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow {
			r[0].Code = "V2"
			return r
		},
		func(r []ManagementTraitsRuntimeSourceRow) []ManagementTraitsRuntimeSourceRow {
			r[0].RelationType = 2
			return r
		},
	}
	for i, mutate := range mutations {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			_, _, err := BuildManagementTraitsRuntimeCurrentSource("00201", mutate(append([]ManagementTraitsRuntimeSourceRow(nil), rows...)))
			if err == nil {
				t.Fatal("invalid source accepted")
			}
		})
	}
	for _, code := range []string{"00101", "00301", "00401", "002", "00202 "} {
		if _, _, err := BuildManagementTraitsRuntimeCurrentSource(code, rows); err == nil {
			t.Fatal("wrong repo accepted")
		}
	}
}

func TestManagementTraitsRuntimeSourceQueryFailure(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectQuery("SELECT .* FROM el_qu_repo qr LEFT JOIN el_qu q.*LEFT JOIN el_qu_answer qa").WithArgs("opaque-repo").WillReturnError(errors.New("sensitive database failure"))
	_, err := loadManagementTraitsRuntimeSource(context.Background(), db, "opaque-repo")
	if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("database error exposed", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsRuntimeClosedGate(t *testing.T) {
	s := NewManagementTraitsRuntimeService(nil, "test-only-secret", 1024*1024)
	if s.CandidateRuntimeEnabled() {
		t.Fatal("runtime opened before old write protection")
	}
	if _, err := s.FreezeProfile(context.Background(), "exam"); !errors.Is(err, ErrManagementTraitsRuntimeClosed) {
		t.Fatal("freeze did not fail closed", err)
	}
}

func managementRuntimeResultFixture(t *testing.T, unanswered bool) (managementTraitsResultValidationFixture, time.Time) {
	t.Helper()
	f, _ := managementTraitsResultValidationFixtureFor(t, "staff", "candidate", 3, 0)
	m, err := DecodeManagementTraitsManifest([]byte(f.Bundle.ScoringManifest), 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	_, m.Versions, _ = managementTraitsRuntimeVersions("00201")
	mc, err := CanonicalManagementTraitsManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := DecodeManagementTraitsMapping([]byte(f.Paper.MappingSnapshot), ManagementTraitsManifest{Questionnaire: "staff", Versions: ManagementTraitsScoringVersions{Product: f.Bundle.ProductVersion, Question: f.Bundle.QuestionVersion, Scoring: f.Bundle.ScoringVersion, Norm: f.Bundle.NormVersion}, Questions: m.Questions}, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	mapping.ManifestSHA = mc.SHA256
	pc, err := CanonicalManagementTraitsMapping(m, mapping)
	if err != nil {
		t.Fatal(err)
	}
	f.Bundle.ProductVersion, f.Bundle.QuestionVersion, f.Bundle.ScoringVersion, f.Bundle.NormVersion = m.Versions.Product, m.Versions.Question, m.Versions.Scoring, m.Versions.Norm
	f.Bundle.ScoringManifest, f.Bundle.ScoringManifestSHA, f.Bundle.Status = string(mc.JSON), mc.SHA256, "candidate-current-source"
	f.Paper.MappingSnapshot, f.Paper.MappingSHA, f.Paper.ScoringManifestSHA = string(pc.JSON), pc.SHA256, mc.SHA256
	for i := range f.Questions {
		q := &f.Questions[i]
		frozen, err := managementTraitsCanonicalBytes(managementTraitsFrozenQuestion{"mng-frozen-question-v1", mc.SHA256, pc.SHA256, mapping.Questions[q.Number-1], q.DimensionKey, q.Reverse, q.PaperQuestionID, q.DisplayOrder})
		if err != nil {
			t.Fatal(err)
		}
		q.ScoringSnapshotSHA = frozen.SHA256
		q.SubmittedAt = nil
	}
	if unanswered {
		f.Questions[0].SelectedOptionID, f.Questions[0].RawAnswer, f.Questions[0].FinalScore = nil, nil, nil
	}
	start := time.Date(2026, 10, 2, 8, 0, 0, 0, time.UTC)
	f.Paper.StartedAt, f.Paper.CreatedAt, f.Paper.SourceCapturedAt = start, start, start.Add(-time.Minute)
	limit := start.Add(25 * time.Minute)
	f.Paper.LimitTime = &limit
	return f, start
}

func TestManagementTraitsRuntimeResultRecords(t *testing.T) {
	for _, incomplete := range []bool{false, true} {
		f, start := managementRuntimeResultFixture(t, incomplete)
		at := start.Add(5 * time.Minute)
		if incomplete {
			at = start.Add(25 * time.Minute)
		}
		records, err := buildManagementTraitsRuntimeRecords(f.Bundle, f.Paper, f.Questions, "opaque-run", at, 1024*1024)
		if err != nil {
			t.Fatal(err)
		}
		if len(records.Dimensions) != 13 || len(records.Modules) != 4 || records.Run.Source != "submission" || *records.Run.UserTimeSeconds != int(at.Sub(start).Seconds()) {
			t.Fatal("invalid record facts")
		}
		if records.Receipt.SubmitType != map[bool]string{false: "manual", true: "timeout"}[incomplete] {
			t.Fatal("client timeout used")
		}
		if incomplete {
			if records.Run.Status != "incomplete" || records.Run.OverallScore != nil || records.Run.OverallNorm != nil || records.Run.OverallLevel != nil || records.Run.AnsweredQuestionCount != 139 {
				t.Fatal("incomplete official values")
			}
			for _, d := range records.Dimensions {
				if d.Score != nil || d.Norm != nil || d.Level != nil {
					t.Fatal("incomplete dimension value")
				}
			}
			for _, m := range records.Modules {
				if m.Score != nil {
					t.Fatal("incomplete module value")
				}
			}
		} else if _, err := ValidateManagementTraitsStoredResult(f.Bundle, f.Paper, f.Questions, records.Run, records.Dimensions, records.Modules, 1024*1024); err != nil {
			t.Fatal("not S2D compatible", err)
		}
	}
	f, start := managementRuntimeResultFixture(t, true)
	if _, err := buildManagementTraitsRuntimeRecords(f.Bundle, f.Paper, f.Questions, "run", start.Add(time.Minute), 1024*1024); !errors.Is(err, ErrManagementTraitsRuntimeMissing) {
		t.Fatal("early missing answers accepted", err)
	}
}

func TestManagementTraitsRuntimeResultWriterRollback(t *testing.T) {
	f, start := managementRuntimeResultFixture(t, false)
	records, err := buildManagementTraitsRuntimeRecords(f.Bundle, f.Paper, f.Questions, "run", start.Add(time.Minute), 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	db, mock := managementRuntimeDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `el_mng_result_run`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `el_mng_result_dimension`").WillReturnResult(sqlmock.NewResult(1, 13))
	mock.ExpectExec("INSERT INTO `el_mng_result_module`").WillReturnError(errors.New("private DB error"))
	mock.ExpectRollback()
	err = db.Transaction(func(tx *gorm.DB) error { return persistManagementTraitsRuntimeRecords(tx, records) })
	if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("write error not controlled", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsRuntimeOwnerAmbiguity(t *testing.T) {
	for _, rows := range [][]managementTraitsRuntimeOwner{
		{},
		{{Kind: "candidate", ID: "owner", ExamID: "exam", PaperID: "paper", DelFlag: 0}, {Kind: "tester", ID: "other", ExamID: "exam", PaperID: "paper", DelFlag: 0, Status: "0"}},
		{{Kind: "candidate", ID: "owner", ExamID: "other", PaperID: "paper", DelFlag: 0}},
		{{Kind: "candidate", ID: "owner", ExamID: "exam", PaperID: "other", DelFlag: 0}},
	} {
		if validateManagementTraitsRuntimeOwner(rows, "candidate", "owner", "exam", "paper") == nil {
			t.Fatal("ambiguous ownership accepted")
		}
	}
	if validateManagementTraitsRuntimeOwner([]managementTraitsRuntimeOwner{{Kind: "candidate", ID: "owner", ExamID: "exam", PaperID: "paper", DelFlag: 0}}, "candidate", "owner", "exam", "paper") != nil {
		t.Fatal("real owner rejected")
	}
}
