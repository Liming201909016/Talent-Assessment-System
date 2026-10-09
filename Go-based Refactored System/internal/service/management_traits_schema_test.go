package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestManagementTraitsSchemaAbsentAndFailureCached(t *testing.T) {
	for _, failure := range []bool{false, true} {
		db, mock := managementRuntimeDB(t)
		q := mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables")
		if failure {
			q.WillReturnError(errors.New("private mysql error"))
		} else {
			q.WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
		}
		s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
		for i := 0; i < 2; i++ {
			err := s.CheckRuntimeSchema(context.Background())
			if failure && !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Fatal(err)
			}
			if !failure && !errors.Is(err, ErrManagementTraitsRuntimeClosed) {
				t.Fatal("absent schema enabled", err)
			}
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagementTraitsSchemaExactSignature(t *testing.T) {
	contract, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	metadata := managementTraitsSchemaFixture(contract)
	if err := validateManagementTraitsSchema(contract, metadata); err != nil {
		t.Fatal("valid signature rejected", err)
	}
	for _, change := range []struct {
		name   string
		mutate func(*managementTraitsSchemaMetadata)
	}{
		{"missing column", func(m *managementTraitsSchemaMetadata) { m.Columns = m.Columns[1:] }},
		{"wrong nullable", func(m *managementTraitsSchemaMetadata) { m.Columns[0].Nullable = "YES" }},
		{"wrong type", func(m *managementTraitsSchemaMetadata) { m.Columns[0].ColumnType = "varchar(63)" }},
		{"wrong collation", func(m *managementTraitsSchemaMetadata) { m.Columns[0].Collation = "utf8mb4_general_ci" }},
		{"missing unique", func(m *managementTraitsSchemaMetadata) { m.Indexes = m.Indexes[1:] }},
		{"index nonunique", func(m *managementTraitsSchemaMetadata) { m.Indexes[0].NonUnique = 1 }},
		{"missing fk", func(m *managementTraitsSchemaMetadata) { m.ForeignKeys = m.ForeignKeys[1:] }},
		{"cascade", func(m *managementTraitsSchemaMetadata) { m.ForeignKeys[0].DeleteRule = "CASCADE" }},
		{"wrong referenced column", func(m *managementTraitsSchemaMetadata) { m.ForeignKeys[0].ReferencedColumn = "wrong" }},
		{"wrong fk order", func(m *managementTraitsSchemaMetadata) { m.ForeignKeys[0].Position = 2 }},
	} {
		t.Run(change.name, func(t *testing.T) {
			m := managementTraitsSchemaFixture(contract)
			change.mutate(&m)
			if validateManagementTraitsSchema(contract, m) == nil {
				t.Fatal("drift accepted")
			}
		})
	}
	// A complete signature is read once before any transaction, with all SQL
	// aliases explicitly mapped rather than GORM implicit name inference.
	db, mock := managementRuntimeDB(t)
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(managementRuntimeModelRows(t, managementTraitsSchemaTableRows(contract)...))
	cols := make([]any, len(metadata.Columns))
	for i := range cols {
		cols[i] = metadata.Columns[i]
	}
	idx := make([]any, len(metadata.Indexes))
	for i := range idx {
		idx[i] = metadata.Indexes[i]
	}
	fks := make([]any, len(metadata.ForeignKeys))
	for i := range fks {
		fks[i] = metadata.ForeignKeys[i]
	}
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type").WillReturnRows(managementRuntimeModelRows(t, cols...))
	mock.ExpectQuery("SELECT table_name AS table_name, index_name AS index_name, non_unique AS non_unique").WillReturnRows(managementRuntimeModelRows(t, idx...))
	mock.ExpectQuery("SELECT k.table_name AS table_name, k.constraint_name AS constraint_name").WillReturnRows(managementRuntimeModelRows(t, fks...))
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	for i := 0; i < 2; i++ {
		if err := s.CheckRuntimeSchema(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsDDLModelContract(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "sql", "management_traits_001_runtime.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(raw)
	contract, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range contract.Tables {
		if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS "+table.Name) {
			t.Fatal("missing table", table.Name)
		}
		for _, col := range table.Columns {
			if !strings.Contains(sql, "`"+col.Name+"` "+col.Type) {
				t.Fatal("missing column signature", table.Name, col.Name)
			}
		}
	}
	for _, fk := range contract.ForeignKeys {
		if !strings.Contains(sql, "CONSTRAINT "+fk.Name+" FOREIGN KEY") {
			t.Fatal("missing FK", fk.Name)
		}
	}
	for _, required := range []string{"column_type='varchar(64)'", "is_nullable='NO'", "@mng_parents=3"} {
		if !strings.Contains(sql, required) {
			t.Fatal("parent preflight is not exact", required)
		}
	}
	for _, bad := range []string{"DROP TABLE", "TRUNCATE", "UPDATE el_", "DELETE FROM", "AutoMigrate", "DELIMITER", "utf8mb4_0900"} {
		if strings.Contains(sql, bad) {
			t.Fatal("unsafe/incompatible migration", bad)
		}
	}
	if strings.Contains(sql, "character_maximum_length<=64") {
		t.Fatal("parent preflight accepts a non-identical varchar length")
	}
	if len(raw) > 2 && raw[0] == 0xef && raw[1] == 0xbb && raw[2] == 0xbf {
		t.Fatal("BOM")
	}
}

func managementTraitsSchemaTableRows(d managementTraitsSchemaDefinition) []any {
	rows := make([]any, 0, len(d.Tables))
	for _, t := range d.Tables {
		rows = append(rows, managementTraitsSchemaTableRow{Name: t.Name, Engine: "InnoDB"})
	}
	return rows
}

func managementTraitsSchemaFixture(d managementTraitsSchemaDefinition) managementTraitsSchemaMetadata {
	var m managementTraitsSchemaMetadata
	for _, t := range d.Tables {
		for _, c := range t.Columns {
			n := "NO"
			if c.Nullable {
				n = "YES"
			}
			r := managementTraitsSchemaColumnRow{Table: t.Name, Name: c.Name, ColumnType: c.Type, Nullable: n}
			if strings.HasPrefix(c.Type, "varchar") || strings.HasPrefix(c.Type, "char") || c.Type == "longtext" {
				r.Charset, r.Collation = "utf8mb4", "utf8mb4_bin"
			}
			m.Columns = append(m.Columns, r)
		}
		for _, i := range t.Indexes {
			for n, c := range i.Columns {
				prefix := 0
				if len(i.Prefixes) > n {
					prefix = i.Prefixes[n]
				}
				m.Indexes = append(m.Indexes, managementTraitsSchemaIndexRow{Table: t.Name, Name: i.Name, Position: n + 1, Column: c, NonUnique: i.NonUnique, Prefix: prefix})
			}
		}
	}
	for _, name := range []string{"el_exam", "el_paper", "el_paper_qu"} {
		m.Columns = append(m.Columns, managementTraitsSchemaColumnRow{Table: name, Name: "id", ColumnType: "varchar(64)", Nullable: "NO", Charset: "utf8mb4", Collation: "utf8mb4_bin"})
	}
	m.Columns = append(m.Columns, managementSchemaLegacyReferences()...)
	for _, f := range d.ForeignKeys {
		for n, c := range f.Columns {
			m.ForeignKeys = append(m.ForeignKeys, managementTraitsSchemaFKRow{Table: f.Table, Name: f.Name, Column: c, Position: n + 1, ReferencedTable: f.RefTable, ReferencedColumn: f.RefColumns[n], UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"})
		}
	}
	return m
}

func TestManagementTraitsSchemaExpiryAndReadIndexesRequired(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"idx_mng_profile_bundle", "idx_mng_snapshot_owner", "idx_mng_snapshot_expiry", "idx_mng_run_exam", "idx_mng_audit_report"} {
		found := false
		for _, table := range d.Tables {
			for _, index := range table.Indexes {
				if index.Name == name {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("runtime index absent from schema signature: %s", name)
		}
		m := managementTraitsSchemaFixture(d)
		rows := make([]managementTraitsSchemaIndexRow, 0)
		for _, r := range m.Indexes {
			if r.Name != name {
				rows = append(rows, r)
			}
		}
		m.Indexes = rows
		if validateManagementTraitsSchema(d, m) == nil {
			t.Fatal("missing read/expiry index accepted", name)
		}
	}
}
