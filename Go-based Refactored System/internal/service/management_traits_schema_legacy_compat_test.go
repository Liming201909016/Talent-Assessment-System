package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

// MT-SCHEMA-LEGACY-002: docs/regression-tests.md. Only the verified opaque
// legacy edges differ; all eleven sidecars and fifteen FKs remain intact.
func managementSchemaLegacy002Fixture(d managementTraitsSchemaDefinition, narrow, mb3 bool) managementTraitsSchemaMetadata {
	m := managementTraitsSchemaFixture(d)
	for i := range m.Columns {
		c := &m.Columns[i]
		if c.Charset != "" {
			c.Collation = "utf8mb4_0900_ai_ci"
		}
		if narrow && c.Table == "el_paper_qu_answer" && (c.Name == "qu_id" || c.Name == "answer_id") {
			c.ColumnType = "varchar(32)"
		}
		if mb3 && c.Table == "el_repo" && c.Name == "id" {
			c.Charset, c.Collation = "utf8mb3", "utf8mb3_general_ci"
		}
	}
	return m
}

func managementSchemaLegacy002Service(t *testing.T, narrow, mb3 bool) (*ManagementTraitsRuntimeService, *gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	db, mock := managementRuntimeDB(t)
	managementSchemaExpectMetadata(mock, d, managementSchemaLegacy002Fixture(d, narrow, mb3))
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	if err := s.CheckRuntimeSchema(context.Background()); err != nil {
		t.Fatal("exact legacy metadata rejected", err)
	}
	return s, db, mock
}

func TestBugMTSchemaLegacy002Metadata(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil || len(d.Tables) != 11 || len(d.ForeignKeys) != 15 {
		t.Fatal("sidecar contract changed", err)
	}
	for _, flags := range [][2]bool{{true, false}, {false, true}, {true, true}} {
		t.Run(fmt.Sprint(flags), func(t *testing.T) {
			managementSchemaCheckTwice(t, d, managementSchemaLegacy002Fixture(d, flags[0], flags[1]), nil)
		})
	}
}

func TestBugMTSchemaLegacy002Drift(t *testing.T) {
	d, _ := managementTraitsSchemaContract()
	for _, key := range []string{"el_qu.id", "el_qu_answer.id", "el_paper_qu_answer.qu_id", "el_paper_qu_answer.answer_id", "el_repo.id", "el_exam_repo.repo_id", "el_qu_repo.repo_id"} {
		for _, change := range []string{"nullable", "char", "text", "capacity", "charset", "collation"} {
			t.Run(key+"/"+change, func(t *testing.T) {
				m := managementSchemaLegacy002Fixture(d, true, true)
				for i := range m.Columns {
					c := &m.Columns[i]
					if c.Table+"."+c.Name != key {
						continue
					}
					switch change {
					case "nullable":
						c.Nullable = "YES"
					case "char":
						c.ColumnType = "char(64)"
					case "text":
						c.ColumnType = "text"
					case "capacity":
						c.ColumnType = "varchar(31)"
						if key == "el_qu.id" || key == "el_qu_answer.id" {
							c.ColumnType = "varchar(63)"
						}
					case "charset":
						c.Charset = "latin1"
					case "collation":
						c.Collation = "utf8mb4_general_ci"
					}
				}
				managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
	for _, child := range []string{"qu_id", "answer_id"} {
		for _, length := range []int{1, 31, 33, 35, 63, 65} {
			t.Run(fmt.Sprintf("%s/%d", child, length), func(t *testing.T) {
				m := managementSchemaLegacy002Fixture(d, true, true)
				for i := range m.Columns {
					if m.Columns[i].Table == "el_paper_qu_answer" && m.Columns[i].Name == child {
						m.Columns[i].ColumnType = fmt.Sprintf("varchar(%d)", length)
					}
				}
				managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
	for _, collation := range []string{"utf8mb4_bin", "utf8mb4_general_ci"} {
		t.Run("unverified_base/"+collation, func(t *testing.T) {
			m := managementSchemaLegacy002Fixture(d, true, true)
			for i := range m.Columns {
				if m.Columns[i].Charset == "utf8mb4" {
					m.Columns[i].Collation = collation
				}
			}
			managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

func TestBugMTSchemaLegacy002NewContractsStrict(t *testing.T) {
	d, _ := managementTraitsSchemaContract()
	fixture := managementSchemaLegacy002Fixture(d, true, true)
	for at, column := range fixture.Columns {
		if !strings.HasPrefix(column.Table, "el_mng_") {
			continue
		}
		for _, change := range []string{"type", "nullable", "collation"} {
			if change == "collation" && column.Charset == "" {
				continue
			}
			t.Run(column.Table+"."+column.Name+"/"+change, func(t *testing.T) {
				m := managementSchemaLegacy002Fixture(d, true, true)
				switch change {
				case "type":
					m.Columns[at].ColumnType = "varchar(32)"
				case "nullable":
					m.Columns[at].Nullable = "YES"
					if column.Nullable == "YES" {
						m.Columns[at].Nullable = "NO"
					}
				case "collation":
					m.Columns[at].Collation = "utf8mb4_general_ci"
				}
				managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
	for at := range fixture.ForeignKeys {
		t.Run(fmt.Sprint("fk/", at), func(t *testing.T) {
			m := managementSchemaLegacy002Fixture(d, true, true)
			m.ForeignKeys[at].DeleteRule = "CASCADE"
			managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
		})
	}
	for at := range fixture.Indexes {
		t.Run(fmt.Sprint("index/", at), func(t *testing.T) {
			m := managementSchemaLegacy002Fixture(d, true, true)
			m.Indexes[at].Prefix++
			managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

func TestBugMTSchemaLegacy002SourceBudgets(t *testing.T) {
	for _, field := range []string{"question", "option"} {
		for _, id := range []string{strings.Repeat("x", 31), strings.Repeat("x", 32), strings.Repeat("x", 33), "nonascii-界", "emoji-😀", "", string([]byte{0xff})} {
			t.Run(fmt.Sprintf("%s/%x", field, id), func(t *testing.T) {
				s, _, mock := managementSchemaLegacy002Service(t, true, true)
				rows := managementRuntimeSourceRows()
				if field == "question" {
					for i := 0; i < 5; i++ {
						rows[i].QuestionID = id
					}
				} else {
					rows[0].OptionID = id
				}
				values := make([]any, len(rows))
				for i := range rows {
					values[i] = rows[i]
				}
				mock.ExpectQuery("SELECT qr.id AS relation_id").WithArgs("2019583026766479361").WillReturnRows(managementRuntimeModelRows(t, values...))
				loaded, err := loadManagementTraitsRuntimeSource(context.Background(), s.db, "2019583026766479361")
				valid := id == strings.Repeat("x", 31) || id == strings.Repeat("x", 32)
				if valid {
					if err != nil || !reflect.DeepEqual(loaded, rows) {
						t.Fatal("valid source changed or rejected", err)
					}
					for _, code := range []string{"00201", "00202"} {
						m1, p1, e1 := BuildManagementTraitsRuntimeCurrentSource(code, rows)
						m2, p2, e2 := BuildManagementTraitsRuntimeCurrentSource(code, loaded)
						if e1 != nil || e2 != nil || !reflect.DeepEqual(m1, m2) || !reflect.DeepEqual(p1, p2) || len(m2.Questions) != 140 {
							t.Fatal("source/hash/frozen 140 changed", e1, e2)
						}
					}
				} else if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
					t.Fatal("unsupported source reached freeze", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestBugMTSchemaLegacy002RepoEncoding(t *testing.T) {
	s, original, mock := managementSchemaLegacy002Service(t, false, true)
	good := strings.Repeat("r", 64)
	mock.ExpectQuery("SELECT qr.id AS relation_id").WithArgs(good).WillReturnRows(managementRuntimeModelRows(t, managementRuntimeSourceRows()[0]))
	for _, id := range []string{strings.Repeat("r", 65), "repo-界", "repo-😀"} {
		if _, err := loadManagementTraitsRuntimeSource(context.Background(), s.db, id); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatal("unsupported repo encoding reached SQL", err)
		}
	}
	if _, err := loadManagementTraitsRuntimeSource(context.Background(), s.db, good); err != nil {
		t.Fatal("64 ASCII characters treated as mb3 byte capacity", err)
	}
	// Keep the original registry/legacy API untouched, including Unicode IDs.
	mock.ExpectQuery("SELECT").WillReturnRows(managementRuntimeModelRows(t, model.Repo{ID: "repo-界"}))
	var repo model.Repo
	if err := original.Where("id = ?", "repo-界").Take(&repo).Error; err != nil {
		t.Fatal("legacy shared DB changed", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugMTSchemaLegacy002WriterPreflight(t *testing.T) {
	for _, field := range []string{"question", "option"} {
		for _, id := range []string{strings.Repeat("x", 33), "nonascii-界"} {
			t.Run(fmt.Sprintf("%s/%x", field, id), func(t *testing.T) {
				s, _, mock := managementSchemaLegacy002Service(t, true, true)
				r := managementTraitsRuntimePaperRecords{Paper: model.Paper{ID: "paper"}, LegacyQuestions: []model.PaperQu{{ID: "pq", PaperID: "paper", QuID: "question"}}, Buckets: []model.PaperQuAnswer{{ID: "bucket", PaperID: "paper", QuID: "question", AnswerID: "option"}}}
				if field == "question" {
					r.LegacyQuestions[0].QuID = id
				} else {
					r.Buckets[0].AnswerID = id
				}
				mock.ExpectBegin()
				mock.ExpectRollback()
				err := s.db.Transaction(func(tx *gorm.DB) error { return persistManagementTraitsRuntimePaper(tx, r) })
				if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
					t.Fatal("writer failed open", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal("writer attempted SQL before all source IDs fit", err)
				}
			})
		}
	}
}

func TestBugMTSchemaLegacy002ReadAndSelectedBudgets(t *testing.T) {
	for _, path := range []string{"loaded_question", "loaded_option", "selected", "mapping", "options", "checked"} {
		for _, id := range []string{strings.Repeat("x", 31), strings.Repeat("x", 32), strings.Repeat("x", 33), "nonascii-界", "emoji-😀"} {
			t.Run(fmt.Sprintf("%s/%x", path, id), func(t *testing.T) {
				s, _, mock := managementSchemaLegacy002Service(t, true, true)
				valid := id == strings.Repeat("x", 31) || id == strings.Repeat("x", 32)
				var value any
				switch path {
				case "loaded_question":
					value = &model.ManagementTraitsPaperQuestionSnapshot{SourceQuestionID: id}
				case "loaded_option":
					value = &model.PaperQuAnswer{AnswerID: id}
				case "selected":
					value = &model.ManagementTraitsPaperQuestionSnapshot{SelectedOptionID: &id}
				case "mapping":
					f, _, _, _, _ := managementRuntimeLoadFixture(t, false)
					var mapping managementTraitsCanonicalMapping
					if managementTraitsDecodeStrict([]byte(f.Paper.MappingSnapshot), 1<<20, &mapping) != nil {
						t.Fatal("mapping fixture")
					}
					mapping.Questions[0].SourceQuestionID = id
					encoded, _ := managementTraitsCanonicalBytes(mapping)
					value = &model.ManagementTraitsExamProfile{MappingSnapshot: string(encoded.JSON)}
				case "options":
					encoded, _ := managementTraitsCanonicalBytes([]ManagementTraitsMappedOption{{SourceOptionID: id, Raw: 1, Content: "text", DisplayOrder: 1}})
					value = &model.ManagementTraitsPaperQuestionSnapshot{OptionsSnapshot: string(encoded.JSON)}
				}
				var err error
				if path == "checked" {
					if valid {
						mock.ExpectBegin()
						mock.ExpectExec("UPDATE `el_paper_qu_answer`").WithArgs(id, "paper", "question").WillReturnResult(sqlmock.NewResult(0, 5))
						mock.ExpectCommit()
					}
					err = s.db.Model(&model.PaperQuAnswer{}).Where("paper_id = ? AND qu_id = ?", "paper", "question").Update("checked", gorm.Expr("CASE WHEN answer_id = ? THEN 1 ELSE 0 END", id)).Error
				} else {
					mock.ExpectQuery("SELECT").WillReturnRows(managementRuntimeModelRows(t, reflect.ValueOf(value).Elem().Interface()))
					value = reflect.New(reflect.ValueOf(value).Elem().Type()).Interface()
					err = s.db.Find(value).Error
				}
				if valid && err != nil || !valid && !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
					t.Fatal("actual read/selected budget mismatch", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestBugMTSchemaLegacy002UUIDParentsStillReject(t *testing.T) {
	d, _ := managementTraitsSchemaContract()
	for _, parent := range []string{"el_paper", "el_paper_qu"} {
		for length := 1; length < 36; length++ {
			t.Run(fmt.Sprintf("%s/%d", parent, length), func(t *testing.T) {
				m := managementSchemaLegacy002Fixture(d, true, true)
				for i := range m.Columns {
					if m.Columns[i].Table == parent && m.Columns[i].Name == "id" {
						m.Columns[i].ColumnType = fmt.Sprintf("varchar(%d)", length)
					}
				}
				managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
}

func TestBugMTSchemaLegacy002PublicRunLoad(t *testing.T) {
	s, _, mock := managementSchemaLegacy002Service(t, true, true)
	f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
	mock.ExpectCommit()
	got, err := s.LoadValidatedRun(context.Background(), records.Run.ID)
	if err != nil || got.RunID != records.Run.ID || got.Result.OverallScore.RatString() != "50" {
		t.Fatal("public frozen run loader rejected compatible legacy schema", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugMTSchemaLegacy002PaperWriterPositive(t *testing.T) {
	s, _, mock := managementSchemaLegacy002Service(t, true, true)
	manifest, mapping, err := BuildManagementTraitsRuntimeCurrentSource("00201", managementRuntimeSourceRows())
	if err != nil {
		t.Fatal(err)
	}
	mc, _ := CanonicalManagementTraitsManifest(manifest)
	pc, _ := CanonicalManagementTraitsMapping(manifest, mapping)
	now := time.Date(2026, 10, 3, 8, 0, 0, 0, time.UTC)
	frozen := now.Add(-time.Minute)
	fc, _ := managementTraitsCanonicalBytes(managementTraitsRuntimeFieldContract{Schema: "mng-candidate-fields-v1", RequiredFields: []string{"name"}, TimePolicy: "candidate-only-personal-25-minutes-v1", Source: "current-source-not-client-leader-full-not-historical", RepoCode: "00201", CapturedAt: frozen, ManifestSHA: mc.SHA256, MappingSHA: pc.SHA256})
	profile := model.ManagementTraitsExamProfile{ExamID: "exam", BundleID: "bundle", MappingSnapshot: string(pc.JSON), MappingSHA: pc.SHA256, FieldContract: string(fc.JSON), TotalTimeMinutes: 25, FrozenAt: &frozen, CreatedAt: frozen}
	bundle := model.ManagementTraitsDefinitionBundle{ID: "bundle", ProductVersion: manifest.Versions.Product, QuestionVersion: manifest.Versions.Question, ScoringVersion: manifest.Versions.Scoring, NormVersion: manifest.Versions.Norm, Questionnaire: manifest.Questionnaire, ScoringManifest: string(mc.JSON), ScoringManifestSHA: mc.SHA256, Status: "candidate-current-source"}
	order := make([]string, len(mapping.Questions))
	for i, q := range mapping.Questions {
		order[i] = q.SourceQuestionID
	}
	r, err := buildManagementTraitsRuntimePaper(profile, bundle, model.Exam{ID: "exam", State: 0, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25}, managementTraitsRuntimeOwner{Kind: "candidate", ID: "owner", ExamID: "exam", Name: "Person"}, "paper", now, order, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `el_paper`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("SAVEPOINT").WillReturnResult(sqlmock.NewResult(0, 0))
	for i := 0; i < 2; i++ {
		mock.ExpectExec("INSERT INTO `el_paper_qu`").WillReturnResult(sqlmock.NewResult(0, 100))
	}
	mock.ExpectExec("SAVEPOINT").WillReturnResult(sqlmock.NewResult(0, 0))
	for i := 0; i < 7; i++ {
		mock.ExpectExec("INSERT INTO `el_paper_qu_answer`").WillReturnResult(sqlmock.NewResult(0, 100))
	}
	mock.ExpectExec("INSERT INTO `el_mng_paper_snapshot`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("SAVEPOINT").WillReturnResult(sqlmock.NewResult(0, 0))
	for i := 0; i < 2; i++ {
		mock.ExpectExec("INSERT INTO `el_mng_paper_question_snapshot`").WillReturnResult(sqlmock.NewResult(0, 100))
	}
	mock.ExpectCommit()
	if err := s.db.Transaction(func(tx *gorm.DB) error { return persistManagementTraitsRuntimePaper(tx, r) }); err != nil {
		t.Fatal("140/700 writer rejected exact legacy schema", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
