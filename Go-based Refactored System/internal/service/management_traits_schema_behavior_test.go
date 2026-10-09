package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

// MT-SCHEMA: exercise the public gate with real GORM scans of SQL metadata,
// not a source-string assertion or a pre-populated schema cache.
func managementSchemaExpectMetadata(mock sqlmock.Sqlmock, d managementTraitsSchemaDefinition, m managementTraitsSchemaMetadata) {
	tables := sqlmock.NewRows([]string{"table_name", "engine"})
	for _, table := range d.Tables {
		tables.AddRow(table.Name, "InnoDB")
	}
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(tables)
	columns := sqlmock.NewRows([]string{"table_name", "column_name", "column_type", "is_nullable", "character_set_name", "collation_name"})
	for _, c := range m.Columns {
		columns.AddRow(c.Table, c.Name, c.ColumnType, c.Nullable, c.Charset, c.Collation)
	}
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type").WillReturnRows(columns)
	indexes := sqlmock.NewRows([]string{"table_name", "index_name", "non_unique", "seq_in_index", "column_name", "prefix_length"})
	for _, i := range m.Indexes {
		indexes.AddRow(i.Table, i.Name, i.NonUnique, i.Position, i.Column, i.Prefix)
	}
	mock.ExpectQuery("SELECT table_name AS table_name, index_name AS index_name, non_unique AS non_unique").WillReturnRows(indexes)
	fks := sqlmock.NewRows([]string{"table_name", "constraint_name", "column_name", "ordinal_position", "referenced_table_name", "referenced_column_name", "update_rule", "delete_rule"})
	for _, f := range m.ForeignKeys {
		fks.AddRow(f.Table, f.Name, f.Column, f.Position, f.ReferencedTable, f.ReferencedColumn, f.UpdateRule, f.DeleteRule)
	}
	mock.ExpectQuery("SELECT k.table_name AS table_name, k.constraint_name AS constraint_name").WillReturnRows(fks)
}

func managementSchemaCheckTwice(t *testing.T, d managementTraitsSchemaDefinition, m managementTraitsSchemaMetadata, want error) {
	t.Helper()
	db, mock := managementRuntimeDB(t)
	managementSchemaExpectMetadata(mock, d, m)
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	for n := 0; n < 2; n++ {
		if n == 1 && want != nil {
			// Leave valid metadata pending: a retry must not merely turn an
			// unexpected SQL query into the same apparent cached error.
			managementSchemaExpectMetadata(mock, d, managementTraitsSchemaFixture(d))
		}
		if got := s.CheckRuntimeSchema(context.Background()); !errors.Is(got, want) {
			t.Fatalf("call %d: got %v, want %v", n+1, got, want)
		}
	}
	if want != nil {
		fresh := NewManagementTraitsRuntimeService(db, "test", 1<<20)
		if got := fresh.CheckRuntimeSchema(context.Background()); got != nil {
			t.Fatal("new instance did not refresh valid metadata", got)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugMTSchemaMetadataColumnsFailClosed(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	fixture := managementTraitsSchemaFixture(d)
	for at, column := range fixture.Columns {
		t.Run(column.Table+"."+column.Name, func(t *testing.T) {
			changes := []struct {
				name   string
				mutate func(*managementTraitsSchemaMetadata)
			}{
				{"missing", func(m *managementTraitsSchemaMetadata) { m.Columns = append(m.Columns[:at], m.Columns[at+1:]...) }},
				{"duplicate", func(m *managementTraitsSchemaMetadata) { m.Columns = append(m.Columns, m.Columns[at]) }},
				{"type", func(m *managementTraitsSchemaMetadata) { m.Columns[at].ColumnType = "bigint(garbage)" }},
				{"width_or_precision", func(m *managementTraitsSchemaMetadata) {
					typ := column.ColumnType
					switch {
					case strings.HasPrefix(typ, "varchar"):
						typ = "varchar(65)"
					case strings.HasPrefix(typ, "char"):
						typ = "char(63)"
					case typ == "bigint":
						typ = "int(20)"
					case typ == "tinyint(1)":
						typ = "tinyint(2)"
					case typ == "datetime(6)":
						typ = "datetime(3)"
					case typ == "decimal(18,6)":
						typ = "decimal(18,5)"
					case typ == "longtext":
						typ = "text"
					default:
						t.Fatalf("uncovered column type %s", typ)
					}
					m.Columns[at].ColumnType = typ
				}},
				{"nullable", func(m *managementTraitsSchemaMetadata) {
					m.Columns[at].Nullable = "YES"
					if column.Nullable == "YES" {
						m.Columns[at].Nullable = "NO"
					}
				}},
			}
			if column.Charset != "" {
				changes = append(changes, struct {
					name   string
					mutate func(*managementTraitsSchemaMetadata)
				}{"charset", func(m *managementTraitsSchemaMetadata) { m.Columns[at].Charset = "utf8" }})
				changes = append(changes, struct {
					name   string
					mutate func(*managementTraitsSchemaMetadata)
				}{"collation", func(m *managementTraitsSchemaMetadata) { m.Columns[at].Collation = "utf8mb4_general_ci" }})
				changes = append(changes, struct {
					name   string
					mutate func(*managementTraitsSchemaMetadata)
				}{"empty_collation", func(m *managementTraitsSchemaMetadata) { m.Columns[at].Collation = "" }})
			}
			for _, change := range changes {
				t.Run(change.name, func(t *testing.T) {
					m := managementTraitsSchemaFixture(d)
					change.mutate(&m)
					managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
				})
			}
		})
	}
}

func TestBugMTSchemaParentMetadataSignature(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, parent := range []string{"el_exam", "el_paper", "el_paper_qu"} {
		for _, typ := range []string{"char(64)", "varchar(65)", "varchar(0)", "varchar(garbage)", "bigint", "varchar(64) unsigned"} {
			t.Run(parent+"/"+typ, func(t *testing.T) {
				m := managementTraitsSchemaFixture(d)
				for n := range m.Columns {
					if m.Columns[n].Table == parent {
						m.Columns[n].ColumnType = typ
					}
				}
				managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
}

func TestBugMTSchemaBigintMetadataNormalization(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, typ := range []string{"bigint(garbage)", "bigint()", "bigint(-1)", "bigint(+20)", "bigint(20x)", "bigint(0)", "bigint(256)", "bigint(20) unsigned", "bigint(20) zerofill", "bigint(20)(20)"} {
		t.Run(typ, func(t *testing.T) {
			m := managementTraitsSchemaFixture(d)
			for n := range m.Columns {
				if m.Columns[n].ColumnType == "bigint" {
					m.Columns[n].ColumnType = typ
					break
				}
			}
			managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

func TestManagementTraitsSchemaMetadataCompatibleSignatures(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Tables) != 11 || len(d.ForeignKeys) != 15 {
		t.Fatal("unexpected sidecar contract size")
	}
	for _, collation := range []string{"utf8mb4_bin", "utf8mb4_general_ci", "utf8mb4_0900_ai_ci"} {
		for _, width := range []string{"bigint", "bigint(20)", "BIGINT(1)", "bigint(255)"} {
			for _, length := range []int{1, 32, 64} {
				t.Run(fmt.Sprintf("%s/%s/parent%d", collation, width, length), func(t *testing.T) {
					m := managementTraitsSchemaFixture(d)
					for n := range m.Columns {
						c := &m.Columns[n]
						if c.Charset != "" {
							c.Collation = collation
						}
						if c.ColumnType == "bigint" {
							c.ColumnType = width
						}
						if !strings.HasPrefix(c.Table, "el_mng_") {
							c.ColumnType = fmt.Sprintf("varchar(%d)", length)
						}
					}
					var want error
					if length < 36 {
						want = ErrManagementTraitsRuntimeInvalid
					}
					managementSchemaCheckTwice(t, d, m, want)
				})
			}
		}
	}
}

func TestManagementTraitsSchemaMetadataIndexSignatures(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for at, index := range managementTraitsSchemaFixture(d).Indexes {
		for _, kind := range []string{"missing", "duplicate", "name", "table", "unique", "position", "column", "prefix"} {
			t.Run(fmt.Sprintf("%s/%s/%d/%s", index.Table, index.Name, index.Position, kind), func(t *testing.T) {
				m := managementTraitsSchemaFixture(d)
				switch kind {
				case "missing":
					m.Indexes = append(m.Indexes[:at], m.Indexes[at+1:]...)
				case "duplicate":
					m.Indexes = append(m.Indexes, m.Indexes[at])
				case "name":
					m.Indexes[at].Name += "_wrong"
				case "table":
					m.Indexes[at].Table += "_wrong"
				case "unique":
					m.Indexes[at].NonUnique = 1 - index.NonUnique
				case "position":
					m.Indexes[at].Position++
				case "column":
					m.Indexes[at].Column += "_wrong"
				case "prefix":
					m.Indexes[at].Prefix++
				}
				managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
}

func TestManagementTraitsSchemaMetadataFKSignatures(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for at, fk := range managementTraitsSchemaFixture(d).ForeignKeys {
		for _, kind := range []string{"missing", "duplicate", "name", "table", "column", "position", "ref_table", "ref_column", "update", "delete"} {
			t.Run(fmt.Sprintf("%s/%d/%s", fk.Name, fk.Position, kind), func(t *testing.T) {
				m := managementTraitsSchemaFixture(d)
				switch kind {
				case "missing":
					m.ForeignKeys = append(m.ForeignKeys[:at], m.ForeignKeys[at+1:]...)
				case "duplicate":
					m.ForeignKeys = append(m.ForeignKeys, m.ForeignKeys[at])
				case "name":
					m.ForeignKeys[at].Name += "_wrong"
				case "table":
					m.ForeignKeys[at].Table += "_wrong"
				case "column":
					m.ForeignKeys[at].Column += "_wrong"
				case "position":
					m.ForeignKeys[at].Position++
				case "ref_table":
					m.ForeignKeys[at].ReferencedTable += "_wrong"
				case "ref_column":
					m.ForeignKeys[at].ReferencedColumn += "_wrong"
				case "update":
					m.ForeignKeys[at].UpdateRule = "CASCADE"
				case "delete":
					m.ForeignKeys[at].DeleteRule = "SET NULL"
				}
				managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
}

func TestManagementTraitsSchemaMetadataTablesFailClosed(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for at, table := range d.Tables {
		for _, kind := range []string{"missing", "duplicate", "engine"} {
			t.Run(table.Name+"/"+kind, func(t *testing.T) {
				db, mock := managementRuntimeDB(t)
				rows := sqlmock.NewRows([]string{"table_name", "engine"})
				for n, entry := range d.Tables {
					if n == at && kind == "missing" {
						continue
					}
					engine := "InnoDB"
					if n == at && kind == "engine" {
						engine = "MyISAM"
					}
					rows.AddRow(entry.Name, engine)
				}
				if kind == "duplicate" {
					rows.AddRow(table.Name, "InnoDB")
				}
				mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(rows)
				s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
				if got := s.CheckRuntimeSchema(context.Background()); !errors.Is(got, ErrManagementTraitsRuntimeInvalid) {
					t.Fatal("invalid table enabled", got)
				}
				managementSchemaExpectMetadata(mock, d, managementTraitsSchemaFixture(d))
				if got := s.CheckRuntimeSchema(context.Background()); !errors.Is(got, ErrManagementTraitsRuntimeInvalid) {
					t.Fatal("invalid table cache retried", got)
				}
				if got := NewManagementTraitsRuntimeService(db, "test", 1<<20).CheckRuntimeSchema(context.Background()); got != nil {
					t.Fatal("table cache did not refresh on new instance", got)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsSchemaMetadataErrorsCached(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	queries := []string{"SELECT table_name AS table_name, engine AS engine", "SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type", "SELECT table_name AS table_name, index_name AS index_name, non_unique AS non_unique", "SELECT k.table_name AS table_name, k.constraint_name AS constraint_name"}
	for stage := range queries {
		kinds := []string{"query", "row"}
		if stage >= 2 {
			kinds = append(kinds, "scan")
		}
		if stage >= 1 {
			kinds = append(kinds, "empty")
		}
		for _, kind := range kinds {
			t.Run(fmt.Sprintf("stage%d/%s", stage, kind), func(t *testing.T) {
				db, mock := managementRuntimeDB(t)
				if stage > 0 {
					mock.ExpectQuery(queries[0]).WillReturnRows(managementRuntimeModelRows(t, managementTraitsSchemaTableRows(d)...))
				}
				m := managementTraitsSchemaFixture(d)
				if stage > 1 {
					values := make([]any, len(m.Columns))
					for n := range values {
						values[n] = m.Columns[n]
					}
					mock.ExpectQuery(queries[1]).WillReturnRows(managementRuntimeModelRows(t, values...))
				}
				if stage > 2 {
					values := make([]any, len(m.Indexes))
					for n := range values {
						values[n] = m.Indexes[n]
					}
					mock.ExpectQuery(queries[2]).WillReturnRows(managementRuntimeModelRows(t, values...))
				}
				expect := mock.ExpectQuery(queries[stage])
				switch kind {
				case "query":
					expect.WillReturnError(errors.New("private-schema-secret"))
				case "scan":
					field := "non_unique"
					if stage == 3 {
						field = "ordinal_position"
					}
					expect.WillReturnRows(sqlmock.NewRows([]string{field}).AddRow("not-an-integer"))
				case "row":
					expect.WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("el_mng_partial").RowError(0, errors.New("private-schema-secret")))
				case "empty":
					expect.WillReturnRows(sqlmock.NewRows([]string{"table_name"}))
				}
				if kind == "empty" && stage == 1 {
					values := make([]any, len(m.Indexes))
					for n := range values {
						values[n] = m.Indexes[n]
					}
					mock.ExpectQuery(queries[2]).WillReturnRows(managementRuntimeModelRows(t, values...))
				}
				if kind == "empty" && stage < 3 {
					values := make([]any, len(m.ForeignKeys))
					for n := range values {
						values[n] = m.ForeignKeys[n]
					}
					mock.ExpectQuery(queries[3]).WillReturnRows(managementRuntimeModelRows(t, values...))
				}
				s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
				for n := 0; n < 2; n++ {
					if n == 1 {
						managementSchemaExpectMetadata(mock, d, managementTraitsSchemaFixture(d))
					}
					got := s.CheckRuntimeSchema(context.Background())
					if !errors.Is(got, ErrManagementTraitsRuntimeInvalid) || strings.Contains(got.Error(), "private-schema-secret") {
						t.Fatal("metadata failure leaked or enabled", got)
					}
				}
				if got := NewManagementTraitsRuntimeService(db, "test", 1<<20).CheckRuntimeSchema(context.Background()); got != nil {
					t.Fatal("failed metadata did not refresh on new instance", got)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsSchemaCacheContextAndInstanceIsolation(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	db, mock := managementRuntimeDB(t)
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, invalid := range []context.Context{nil, ctx} {
		if got := s.CheckRuntimeSchema(invalid); !errors.Is(got, ErrManagementTraitsRuntimeInvalid) {
			t.Fatal(got)
		}
	}
	managementSchemaExpectMetadata(mock, d, managementTraitsSchemaFixture(d))
	var wg sync.WaitGroup
	results := make(chan error, 16)
	for n := 0; n < cap(results); n++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- s.CheckRuntimeSchema(context.Background()) }()
	}
	wg.Wait()
	close(results)
	for got := range results {
		if got != nil {
			t.Fatal(got)
		}
	}
	if got := s.CheckRuntimeSchema(ctx); !errors.Is(got, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("cached success ignored canceled context", got)
	}
	// A new instance must re-read, not borrow the preceding instance's success.
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
	fresh := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	if got := fresh.CheckRuntimeSchema(context.Background()); !errors.Is(got, ErrManagementTraitsRuntimeClosed) {
		t.Fatal("absence failed open", got)
	}
	managementSchemaExpectMetadata(mock, d, managementTraitsSchemaFixture(d))
	if got := fresh.CheckRuntimeSchema(context.Background()); !errors.Is(got, ErrManagementTraitsRuntimeClosed) {
		t.Fatal("absence cache retried", got)
	}
	if got := NewManagementTraitsRuntimeService(db, "test", 1<<20).CheckRuntimeSchema(context.Background()); got != nil {
		t.Fatal("absence did not refresh on new instance", got)
	}
	if got := s.CheckRuntimeSchema(context.Background()); got != nil {
		t.Fatal("instance cache changed", got)
	}
	var absent *ManagementTraitsRuntimeService
	if got := absent.CheckRuntimeSchema(context.Background()); !errors.Is(got, ErrManagementTraitsRuntimeClosed) {
		t.Fatal(got)
	}
	if got := (&ManagementTraitsRuntimeService{}).CheckRuntimeSchema(context.Background()); !errors.Is(got, ErrManagementTraitsRuntimeClosed) {
		t.Fatal(got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsSchemaMetadataOrderingAndMixedParentLengths(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	m := managementTraitsSchemaFixture(d)
	for n := range m.Columns {
		if m.Columns[n].Table == "el_exam" && m.Columns[n].Name == "id" {
			m.Columns[n].ColumnType = "varchar(32)"
		}
		if m.Columns[n].Table == "el_paper_qu" && m.Columns[n].Name == "id" {
			m.Columns[n].ColumnType = "varchar(36)"
		}
	}
	for a, b := 0, len(m.Columns)-1; a < b; a, b = a+1, b-1 {
		m.Columns[a], m.Columns[b] = m.Columns[b], m.Columns[a]
	}
	for a, b := 0, len(m.Indexes)-1; a < b; a, b = a+1, b-1 {
		m.Indexes[a], m.Indexes[b] = m.Indexes[b], m.Indexes[a]
	}
	for a, b := 0, len(m.ForeignKeys)-1; a < b; a, b = a+1, b-1 {
		m.ForeignKeys[a], m.ForeignKeys[b] = m.ForeignKeys[b], m.ForeignKeys[a]
	}
	managementSchemaCheckTwice(t, d, m, nil)
}

func TestBugMTSchemaUniformInvalidCollationFailsClosed(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, collation := range []string{"", "bogus", "utf8_general_ci", "utf8mb4_", "utf8mb4_bin;DROP", "UTF8MB4_BIN"} {
		t.Run(collation, func(t *testing.T) {
			m := managementTraitsSchemaFixture(d)
			for n := range m.Columns {
				if m.Columns[n].Charset != "" {
					m.Columns[n].Collation = collation
				}
			}
			managementSchemaCheckTwice(t, d, m, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

func managementSchemaReviewService(t *testing.T) (*ManagementTraitsRuntimeService, *gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	m := managementTraitsSchemaFixture(d)
	for n := range m.Columns {
		c := &m.Columns[n]
		if (c.Table == "el_exam" || c.Table == "el_tester") && c.Name == "id" {
			c.ColumnType = "varchar(4)"
		}
		if (c.Table == "el_repo" || c.Table == "el_qu" || c.Table == "el_qu_answer" || c.Table == "el_qu_repo") && c.Name == "id" {
			c.ColumnType = "varchar(8)"
		}
	}
	db, mock := managementRuntimeDB(t)
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	managementSchemaExpectMetadata(mock, d, m)
	if err := s.CheckRuntimeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s, db, mock
}

func TestBugMTSchemaReviewImmutablePrivateRegistry(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	original := s.db
	if original == db {
		t.Error("constructor shares original callback registry")
	}
	d, _ := managementTraitsSchemaContract()
	managementSchemaExpectMetadata(mock, d, managementTraitsSchemaFixture(d))
	if err := s.CheckRuntimeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	if original != s.db {
		t.Error("schema once replaced service DB")
	}
	if s.db.ConnPool != db.ConnPool {
		t.Error("constructor changed connection pool")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugMTSchemaReviewIdentityCountNoPanic(t *testing.T) {
	s, _, mock := managementSchemaReviewService(t)
	mock.ExpectQuery("SELECT count").WithArgs("good").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Row.Scan panicked: %v", r)
			}
		}()
		if _, err := managementTraitsIdentityHasProfile(context.Background(), s.db, "abcde"); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Errorf("over-capacity count accepted: %v", err)
		}
	}()
	if ok, err := managementTraitsIdentityHasProfile(context.Background(), s.db, "good"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugMTSchemaReviewJoinParametersFailClosed(t *testing.T) {
	for _, join := range []string{"JOIN el_exam AS e ON e.id = ?", "JOIN `el_exam` AS `e` ON `e`.`id` = ?", "JOIN el_exam e ON e.id IN (?)", "JOIN el_exam e ON unresolved.id = ?"} {
		t.Run(join, func(t *testing.T) {
			s, _, mock := managementSchemaReviewService(t)
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("good"))
			var rows []model.ManagementTraitsResultRun
			if err := s.db.Table("el_mng_result_run r").Joins(join, "abcde").Find(&rows).Error; !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Errorf("JOIN parameter escaped: %v", err)
			}
			var e model.Exam
			if err := s.db.Where("id = ?", "good").Take(&e).Error; err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugMTSchemaReviewLoadedIDsAndJSONBudgets(t *testing.T) {
	for _, path := range []string{"exam", "owner", "profile_mapping", "snapshot_mapping", "question_id", "question_options", "bucket"} {
		for _, id := range []string{"123456789", "界界界", string([]byte{0xff})} {
			t.Run(path+"/"+fmt.Sprintf("%x", id), func(t *testing.T) {
				s, _, mock := managementSchemaReviewService(t)
				var value any
				switch path {
				case "exam":
					value = &model.Exam{ID: id}
				case "owner":
					value = &managementTraitsRuntimeOwner{Kind: "tester", ID: id, ExamID: "good"}
				case "profile_mapping", "snapshot_mapping":
					f, _, _, p, _ := managementRuntimeLoadFixture(t, false)
					var m managementTraitsCanonicalMapping
					if managementTraitsDecodeStrict([]byte(p.MappingSnapshot), 1<<20, &m) != nil {
						t.Fatal("mapping fixture")
					}
					m.Questions[0].SourceQuestionID = id
					encoded, _ := managementTraitsCanonicalBytes(m)
					if path == "profile_mapping" {
						p.ExamID = "good"
						p.MappingSnapshot = string(encoded.JSON)
						value = &p
					} else {
						f.Paper.ExamID = "good"
						f.Paper.ProfileExamID = nil
						f.Paper.MappingSnapshot = string(encoded.JSON)
						value = &f.Paper
					}
				case "question_id":
					value = &model.ManagementTraitsPaperQuestionSnapshot{SourceQuestionID: id}
				case "question_options":
					encoded, _ := managementTraitsCanonicalBytes([]ManagementTraitsMappedOption{{SourceOptionID: id, Raw: 1, Content: "x", DisplayOrder: 1}})
					text := string(encoded.JSON)
					if id == string([]byte{0xff}) {
						text = strings.Replace(text, "\\ufffd", id, 1)
					}
					value = &model.ManagementTraitsPaperQuestionSnapshot{OptionsSnapshot: text}
				case "bucket":
					value = &model.PaperQuAnswer{AnswerID: id}
				}
				mock.ExpectQuery("SELECT").WillReturnRows(managementRuntimeModelRows(t, reflect.ValueOf(value).Elem().Interface()))
				value = reflect.New(reflect.ValueOf(value).Elem().Type()).Interface()
				q := s.db
				if path == "owner" {
					q = q.Table("el_tester")
				}
				if err := q.Find(value).Error; !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
					t.Errorf("loaded %s ID escaped: %v", path, err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestBugMTSchemaReviewRawSourceLoadedCapacity(t *testing.T) {
	for _, field := range []string{"relation_id", "question_id", "option_id"} {
		t.Run(field, func(t *testing.T) {
			s, _, mock := managementSchemaReviewService(t)
			row := managementRuntimeSourceRows()[0]
			switch field {
			case "relation_id":
				row.RelationID = "123456789"
			case "question_id":
				row.QuestionID = "界界界"
			case "option_id":
				row.OptionID = "123456789"
			}
			mock.ExpectQuery("SELECT qr.id AS relation_id").WithArgs("repo").WillReturnRows(managementRuntimeModelRows(t, row))
			if _, err := loadManagementTraitsRuntimeSource(context.Background(), s.db, "repo"); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Errorf("Raw.Scan source escaped: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugMTSchemaReviewPublicLoadRequiresSchema(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	if _, err := s.LoadValidatedRun(context.Background(), "run"); !errors.Is(err, ErrManagementTraitsRuntimeClosed) {
		t.Errorf("public loader bypassed schema: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
