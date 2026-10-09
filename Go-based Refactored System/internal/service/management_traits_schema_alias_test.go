package service

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MT-SCHEMA-ALIASES: docs/regression-tests.md. Model MySQL metadata labels
// from the executed projection, not from GORM tags or the desired result.
// Only TABLE_NAME/ENGINE have remote evidence; other labels are regression
// simulations of the same driver contract, not additional remote evidence.
type managementSchemaAliasQuery struct {
	sql     string
	labels  []string
	want    []string
	scanned bool
}

var managementSchemaAliasProjection = regexp.MustCompile(`(?i)(?:^|, )(COALESCE\([^)]*\)|[a-z_.]+)(?: AS ([a-z_]+))?`)
var managementSchemaAliasClause = regexp.MustCompile(`(?i) AS [a-z_]+`)

func managementSchemaAliasDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, map[string]*managementSchemaAliasQuery) {
	t.Helper()
	queries := make(map[string]*managementSchemaAliasQuery)
	matcher := sqlmock.QueryMatcherFunc(func(expected, actual string) error {
		q := queries[expected]
		if q == nil {
			return sqlmock.QueryMatcherEqual.Match(expected, actual)
		}
		// The complete old SQL is an independent oracle for filters, JOINs,
		// bounds and ordering. Only projection aliases may change.
		strip := func(s string) string { return managementSchemaAliasClause.ReplaceAllString(s, "") }
		if strip(q.sql) != strip(actual) {
			return fmt.Errorf("metadata SQL changed beyond aliases: %s", actual)
		}
		end := strings.Index(actual, " FROM ")
		if end < 0 {
			return fmt.Errorf("metadata projection missing FROM")
		}
		fields := managementSchemaAliasProjection.FindAllStringSubmatch(strings.TrimPrefix(actual[:end], "SELECT "), -1)
		if len(fields) != len(q.labels) {
			return fmt.Errorf("metadata projection size: %d != %d", len(fields), len(q.labels))
		}
		for i, field := range fields {
			label := field[2]
			if label == "" {
				parts := strings.Split(field[1], ".")
				label = strings.ToUpper(parts[len(parts)-1])
			}
			// NewRows keeps this label slice: the simulated driver reports
			// the actual alias, or uppercase original identifier without AS.
			q.labels[i] = label
		}
		q.scanned = true
		return nil
	})
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	g, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	return g, mock, queries
}

func managementSchemaAliasExpect(mock sqlmock.Sqlmock, queries map[string]*managementSchemaAliasQuery, key, sql string, labels []string, values [][]driver.Value) {
	q := &managementSchemaAliasQuery{sql: sql, labels: append([]string(nil), labels...), want: append([]string(nil), labels...)}
	queries[key] = q
	rows := sqlmock.NewRows(q.labels)
	for _, v := range values {
		rows.AddRow(v...)
	}
	mock.ExpectQuery(key).WillReturnRows(rows).RowsWillBeClosed()
}

func managementSchemaAliasExpectFull(mock sqlmock.Sqlmock, queries map[string]*managementSchemaAliasQuery, d managementTraitsSchemaDefinition, m managementTraitsSchemaMetadata) {
	tables := make([][]driver.Value, 0, len(d.Tables))
	for _, table := range d.Tables {
		tables = append(tables, []driver.Value{table.Name, "InnoDB"})
	}
	managementSchemaAliasExpect(mock, queries, "tables", "SELECT table_name, engine FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name, 7) = 'el_mng_' ORDER BY table_name", []string{"table_name", "engine"}, tables)
	columns := make([][]driver.Value, 0, len(m.Columns))
	for _, c := range m.Columns {
		columns = append(columns, []driver.Value{c.Table, c.Name, c.ColumnType, c.Nullable, c.Charset, c.Collation})
	}
	managementSchemaAliasExpect(mock, queries, "columns", "SELECT table_name, column_name, column_type, is_nullable, COALESCE(character_set_name,'') AS character_set_name, COALESCE(collation_name,'') AS collation_name FROM information_schema.columns WHERE table_schema = DATABASE() AND (LEFT(table_name, 7) = 'el_mng_' OR (table_name IN ('el_exam','el_exam_repo','el_paper','el_paper_qu','el_paper_qu_answer','el_candidate','el_tester','el_repo','el_qu_repo','el_qu','el_qu_answer') AND column_name IN ('id','exam_id','paper_id','user_id','repo_id','qu_id','answer_id'))) ORDER BY table_name, ordinal_position", []string{"table_name", "column_name", "column_type", "is_nullable", "character_set_name", "collation_name"}, columns)
	indexes := make([][]driver.Value, 0, len(m.Indexes))
	for _, i := range m.Indexes {
		indexes = append(indexes, []driver.Value{i.Table, i.Name, int64(i.NonUnique), int64(i.Position), i.Column, int64(i.Prefix)})
	}
	managementSchemaAliasExpect(mock, queries, "indexes", "SELECT table_name, index_name, non_unique, seq_in_index, column_name, COALESCE(sub_part,0) AS prefix_length FROM information_schema.statistics WHERE table_schema = DATABASE() AND LEFT(table_name, 7) = 'el_mng_' ORDER BY table_name,index_name,seq_in_index", []string{"table_name", "index_name", "non_unique", "seq_in_index", "column_name", "prefix_length"}, indexes)
	fks := make([][]driver.Value, 0, len(m.ForeignKeys))
	for _, f := range m.ForeignKeys {
		fks = append(fks, []driver.Value{f.Table, f.Name, f.Column, int64(f.Position), f.ReferencedTable, f.ReferencedColumn, f.UpdateRule, f.DeleteRule})
	}
	managementSchemaAliasExpect(mock, queries, "fks", "SELECT k.table_name, k.constraint_name, k.column_name, k.ordinal_position, k.referenced_table_name, k.referenced_column_name, r.update_rule, r.delete_rule FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.constraint_schema=DATABASE() AND LEFT(k.table_name, 7)='el_mng_' ORDER BY k.table_name,k.constraint_name,k.ordinal_position", []string{"table_name", "constraint_name", "column_name", "ordinal_position", "referenced_table_name", "referenced_column_name", "update_rule", "delete_rule"}, fks)
}

func TestBugMTSchemaAliasDriverControl(t *testing.T) {
	for _, aliased := range []bool{false, true} {
		t.Run(fmt.Sprint(aliased), func(t *testing.T) {
			db, mock, queries := managementSchemaAliasDB(t)
			projection := "table_name, engine"
			if aliased {
				projection = "table_name AS table_name, engine AS engine"
			}
			sql := "SELECT " + projection + " FROM information_schema.tables"
			managementSchemaAliasExpect(mock, queries, "control", sql, []string{"table_name", "engine"}, [][]driver.Value{{"el_mng_definition_bundle", "InnoDB"}})
			var rows []managementTraitsSchemaTableRow
			if err := db.Raw(sql).Scan(&rows).Error; err != nil || len(rows) != 1 {
				t.Fatal(rows, err)
			}
			want := managementTraitsSchemaTableRow{}
			labels := []string{"TABLE_NAME", "ENGINE"}
			if aliased {
				want = managementTraitsSchemaTableRow{Name: "el_mng_definition_bundle", Engine: "InnoDB"}
				labels = []string{"table_name", "engine"}
			}
			if rows[0] != want || !reflect.DeepEqual(queries["control"].labels, labels) {
				t.Fatal("driver label simulation did not reproduce real GORM scan", rows, queries["control"].labels)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugMTSchemaAliasesCompleteGate(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil || len(d.Tables) != 11 || len(d.ForeignKeys) != 15 {
		t.Fatal("contract changed", err)
	}
	for _, kind := range []string{"valid", "actual_legacy", "column_type", "column_nullable", "column_collation", "index_prefix", "fk_action", "fk_order"} {
		t.Run(kind, func(t *testing.T) {
			m := managementTraitsSchemaFixture(d)
			switch kind {
			case "actual_legacy":
				m = managementSchemaLegacy002Fixture(d, true, true)
			case "column_type":
				m.Columns[0].ColumnType = "varchar(63)"
			case "column_nullable":
				m.Columns[0].Nullable = "YES"
			case "column_collation":
				m.Columns[0].Collation = "utf8mb4_general_ci"
			case "index_prefix":
				m.Indexes[0].Prefix++
			case "fk_action":
				m.ForeignKeys[0].DeleteRule = "CASCADE"
			case "fk_order":
				m.ForeignKeys[0].Position++
			}
			db, mock, queries := managementSchemaAliasDB(t)
			managementSchemaAliasExpectFull(mock, queries, d, m)
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			var want error
			if kind != "valid" && kind != "actual_legacy" {
				want = ErrManagementTraitsRuntimeInvalid
			}
			for n := 0; n < 2; n++ {
				if got := s.CheckRuntimeSchema(context.Background()); !errors.Is(got, want) {
					t.Errorf("call %d: got %v want %v", n+1, got, want)
				}
			}
			for key, q := range queries {
				if !q.scanned || !reflect.DeepEqual(q.labels, q.want) {
					t.Errorf("%s: exact executed aliases missing: got %v want %v scanned=%v", key, q.labels, q.want, q.scanned)
				}
			}
			if want == nil && s.schemaCapacity.Load() == nil {
				t.Error("complete metadata did not publish capacity")
			}
			if want != nil && s.schemaCapacity.Load() != nil {
				t.Error("invalid metadata published capacity")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestBugMTSchemaAliasesLegacyMetadata(t *testing.T) {
	for _, kind := range []string{"identity_capture", "guard_capture"} {
		t.Run(kind, func(t *testing.T) {
			db, mock, queries := managementSchemaAliasDB(t)
			if kind == "identity_capture" {
				for n := 0; n < 7; n++ {
					mock.ExpectQuery("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?").WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))
				}
				managementSchemaAliasExpect(mock, queries, "capture", "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name, 7) = 'el_mng_' AND (table_name LIKE '%capture%' OR table_name = 'el_mng_exam_draft') ORDER BY table_name", []string{"table_name"}, [][]driver.Value{{"el_mng_capture_marker"}})
				got, err := probeManagementTraitsIdentitySchema(context.Background(), db)
				if err != nil || !got.present || !got.run || !reflect.DeepEqual(got.captures, []string{"el_mng_capture_marker"}) {
					t.Errorf("identity capture metadata lost: %+v %v", got, err)
				}
			} else {
				managementSchemaAliasExpect(mock, queries, "tables", "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name,7) = 'el_mng_' ORDER BY table_name", []string{"table_name"}, [][]driver.Value{{"el_mng_capture_marker"}})
				managementSchemaAliasExpect(mock, queries, "columns", "SELECT table_name, column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND ((table_name = ?))", []string{"table_name", "column_name"}, [][]driver.Value{{"el_mng_capture_marker", "paper_id"}})
				mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_capture_marker LIMIT 1").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
				got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{AllLegacy: true})
				if err != nil || !got {
					t.Errorf("guard failed to recognize protected capture: %v %v", got, err)
				}
			}
			for key, q := range queries {
				if !q.scanned || !reflect.DeepEqual(q.labels, q.want) {
					t.Errorf("%s: exact capture aliases missing: got %v want %v scanned=%v", key, q.labels, q.want, q.scanned)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}
