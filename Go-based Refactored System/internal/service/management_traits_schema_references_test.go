package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Independent inventory of columns actually read or written by the runtime.
func managementSchemaLegacyReferences() []managementTraitsSchemaColumnRow {
	var rows []managementTraitsSchemaColumnRow
	for table, names := range map[string][]string{
		"el_exam_repo": {"exam_id", "repo_id"},
		"el_repo":      {"id"}, "el_qu_repo": {"id", "repo_id", "qu_id"},
		"el_qu": {"id"}, "el_qu_answer": {"id", "qu_id"},
		"el_candidate": {"id", "exam_id", "paper_id"},
		"el_tester":    {"id", "exam_id", "paper_id"},
		"el_paper":     {"exam_id", "user_id"}, "el_paper_qu": {"paper_id", "qu_id"},
		"el_paper_qu_answer": {"id", "paper_id", "qu_id", "answer_id"},
	} {
		for _, name := range names {
			nullable := "NO"
			if (table == "el_candidate" && name == "paper_id") || (table == "el_tester" && name != "id") {
				nullable = "YES"
			}
			rows = append(rows, managementTraitsSchemaColumnRow{Table: table, Name: name, ColumnType: "varchar(64)", Nullable: nullable, Charset: "utf8mb4", Collation: "utf8mb4_bin"})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Table+"."+rows[i].Name < rows[j].Table+"."+rows[j].Name })
	return rows
}

func TestBugMTSchemaActualReferenceSignatures(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range managementSchemaLegacyReferences() {
		for _, kind := range []string{"missing", "duplicate", "type", "nullable", "charset", "collation", "capacity"} {
			t.Run(c.Table+"."+c.Name+"/"+kind, func(t *testing.T) {
				m := managementTraitsSchemaFixture(d)
				for n := range m.Columns {
					row := &m.Columns[n]
					if row.Table != c.Table || row.Name != c.Name {
						continue
					}
					switch kind {
					case "missing":
						m.Columns = append(m.Columns[:n], m.Columns[n+1:]...)
					case "duplicate":
						m.Columns = append(m.Columns, *row)
					case "type":
						row.ColumnType = "char(64)"
					case "nullable":
						if row.Nullable == "NO" {
							row.Nullable = "YES"
						} else {
							row.Nullable = "NO"
						}
					case "charset":
						row.Charset = "utf8"
					case "collation":
						row.Collation = "utf8mb4_general_ci"
					case "capacity":
						row.ColumnType = "varchar(65)"
					}
					break
				}
				managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
}

func TestBugMTSchemaActualReferenceCapacity(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"el_paper.exam_id", "el_paper.user_id", "el_exam_repo.exam_id", "el_exam_repo.repo_id", "el_qu_repo.repo_id", "el_qu_repo.qu_id", "el_qu_answer.qu_id", "el_candidate.exam_id", "el_candidate.paper_id", "el_tester.exam_id", "el_tester.paper_id", "el_paper_qu.paper_id", "el_paper_qu.qu_id", "el_paper_qu_answer.paper_id", "el_paper_qu_answer.qu_id", "el_paper_qu_answer.answer_id", "el_candidate.id", "el_paper_qu_answer.id"} {
		t.Run(key, func(t *testing.T) {
			m := managementTraitsSchemaFixture(d)
			for n := range m.Columns {
				if m.Columns[n].Table+"."+m.Columns[n].Name == key {
					m.Columns[n].ColumnType = "varchar(35)"
				}
			}
			managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

func TestBugMTSchemaActualExamQueryAndProfileWrite(t *testing.T) {
	d, _ := managementTraitsSchemaContract()
	m := managementTraitsSchemaFixture(d)
	for n := range m.Columns {
		if m.Columns[n].Table == "el_exam" && m.Columns[n].Name == "id" {
			m.Columns[n].ColumnType = "varchar(4)"
		}
	}
	db, mock := managementRuntimeDB(t)
	managementSchemaExpectMetadata(mock, d, m)
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	if err := s.CheckRuntimeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	// A pending sentinel proves rejection precedes SQL, not sqlmock's error.
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("good"))
	for _, id := range []string{"abcde", "界界", strings.Repeat("x", 65)} {
		var exam model.Exam
		if err := s.db.Where("id = ?", id).Take(&exam).Error; !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatalf("query ID accepted: %q: %v", id, err)
		}
		p := model.ManagementTraitsExamProfile{ExamID: id, BundleID: "bundle"}
		if err := s.db.Session(&gorm.Session{SkipDefaultTransaction: true}).Create(&p).Error; !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatalf("profile ID accepted: %q: %v", id, err)
		}
	}
	var exam model.Exam
	if err := s.db.Where("id = ?", "good").Take(&exam).Error; err != nil {
		t.Fatal(err)
	}
	// Original DB must not inherit this service's callbacks.
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("abcde"))
	if err := db.Where("id = ?", "abcde").Take(&exam).Error; err != nil {
		t.Fatal("shared DB modified", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugMTSchemaActualSourceAndParticipantBudgets(t *testing.T) {
	for _, path := range []string{"source", "option", "participant", "raw", "alias", "column", "map"} {
		t.Run(path, func(t *testing.T) {
			d, _ := managementTraitsSchemaContract()
			m := managementTraitsSchemaFixture(d)
			for n := range m.Columns {
				c := &m.Columns[n]
				if (c.Table == "el_qu" || c.Table == "el_qu_answer" || c.Table == "el_repo") && c.Name == "id" {
					c.ColumnType = "varchar(8)"
				}
				if (c.Table == "el_tester" || c.Table == "el_exam") && c.Name == "id" {
					c.ColumnType = "varchar(4)"
				}
			}
			db, mock := managementRuntimeDB(t)
			managementSchemaExpectMetadata(mock, d, m)
			s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
			if err := s.CheckRuntimeSchema(context.Background()); err != nil {
				t.Fatal(err)
			}
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("ok"))
			ids := []string{"123456789", "界界界", string([]byte{0xff})}
			if path == "participant" || path == "alias" || path == "column" {
				ids = []string{"abcde", "界界"}
			}
			for _, id := range ids {
				guarded := s.db.Session(&gorm.Session{SkipDefaultTransaction: true})
				var err error
				switch path {
				case "source":
					q := model.PaperQu{ID: strings.Repeat("p", 36), PaperID: strings.Repeat("s", 36), QuID: id}
					err = guarded.Create(&q).Error
				case "option":
					err = guarded.Model(&model.ManagementTraitsPaperQuestionSnapshot{}).Where("paper_question_id = ?", strings.Repeat("p", 36)).Update("selected_option_id", id).Error
				case "participant":
					r := model.ManagementTraitsRuntimeReceipt{RunID: "run", PaperID: "paper", ExamID: "exam", ParticipantType: "tester", ParticipantID: id}
					err = guarded.Create(&r).Error
				case "raw":
					var rows []ManagementTraitsRuntimeSourceRow
					err = guarded.Raw(managementTraitsRuntimeSourceSQL, id).Scan(&rows).Error
				case "alias":
					var rows []model.ManagementTraitsResultRun
					err = guarded.Table("el_mng_result_run r").Where("r.exam_id = ?", id).Find(&rows).Error
				case "column":
					var rows []model.ManagementTraitsResultRun
					err = guarded.Table("el_mng_result_run r").Where(clause.Eq{Column: clause.Column{Table: "r", Name: "exam_id"}, Value: id}).Find(&rows).Error
				case "map":
					err = guarded.Table("el_candidate").Where("id = ?", "owner").Updates(map[string]any{"exam_id": id}).Error
				}
				if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
					t.Fatalf("%s escaped cached byte budget: %q: %v", path, id, err)
				}
			}
			var repo model.Repo
			if err := s.db.Where("id = ?", "ok").Take(&repo).Error; err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsSchemaOpaqueSourceNotUUID(t *testing.T) {
	d, _ := managementTraitsSchemaContract()
	m := managementTraitsSchemaFixture(d)
	for n := range m.Columns {
		c := &m.Columns[n]
		if (c.Table == "el_qu" || c.Table == "el_qu_answer" || c.Table == "el_repo" || c.Table == "el_qu_repo") && c.Name == "id" {
			c.ColumnType = "varchar(8)"
		}
	}
	db, mock := managementRuntimeDB(t)
	managementSchemaExpectMetadata(mock, d, m)
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	if err := s.CheckRuntimeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	q := model.PaperQu{ID: strings.Repeat("p", 36), PaperID: strings.Repeat("s", 36), QuID: "界界ab"}
	mock.ExpectExec("INSERT INTO `el_paper_qu`").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.db.Session(&gorm.Session{SkipDefaultTransaction: true}).Create(&q).Error; err != nil {
		t.Fatal("8-byte opaque ID rejected", err)
	}
	if err := s.CheckRuntimeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
