package service

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// MT-REAL-TESTER-SCOPE: reproduce the deployed nested audit query on the
// warmed private callback registry, not the unguarded legacy database.
func TestBugManagementTraitsSQLNestedCapacity(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		args        []any
		valid       bool
	}{
		{"actual_tester", "SELECT 1 AS protected FROM el_mng_report_audit WHERE report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE participant_type = ? AND participant_id = ?))", []any{"tester", "1791035468793800801"}, true},
		{"candidate", "SELECT 1 AS protected FROM el_mng_report_audit WHERE report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE participant_type = ? AND participant_id = ?))", []any{"candidate", strings.Repeat("c", 36)}, true},
		{"simple_in", "SELECT 1 AS protected FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE paper_id IN (?,?) AND exam_id = ?)", []any{"paper", "other", "exam"}, true},
		{"compound", "SELECT 1 AS protected FROM el_mng_report_audit WHERE report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE (paper_id = ?) OR (exam_id = ?) OR (participant_type = ? AND participant_id = ?)))", []any{"paper", "exam", "tester", "1791035468793800801"}, true},
		{"shadowed_alias", "SELECT 1 AS protected FROM el_mng_report_audit r WHERE r.id = ? AND r.report_id IN (SELECT r.id FROM el_mng_report_revision r WHERE r.run_id IN (SELECT r.id FROM el_mng_result_run r WHERE r.participant_type = ? AND r.participant_id = ?))", []any{"audit", "tester", "1791035468793800801"}, true},
		{"literal_question", "SELECT '?' AS note FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE paper_id = ?)", []any{"paper"}, true},
		{"comment_question", "SELECT 1 AS protected FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE /* ? FROM el_tester */ paper_id = ?)", []any{"paper"}, true},
		{"typed_tester_boundary", "SELECT 1 AS protected FROM el_mng_report_audit WHERE report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE participant_type = ? AND participant_id = ?))", []any{"tester", strings.Repeat("t", 64)}, true},
		{"typed_tester_overflow", "SELECT 1 AS protected FROM el_mng_report_audit WHERE report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE participant_type = ? AND participant_id = ?))", []any{"tester", strings.Repeat("t", 65)}, false},
		{"unknown_id", "SELECT 1 AS protected FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE absent_id = ?)", []any{"paper"}, false},
		{"unknown_alias", "SELECT 1 AS protected FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE missing.paper_id = ?)", []any{"paper"}, false},
		{"numeric_id", "SELECT 1 AS protected FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE paper_id = ?)", []any{123}, false},
		{"invalid_utf8", "SELECT 1 AS protected FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE paper_id = ?)", []any{string([]byte{0xff})}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, m, _ := managementAuditDDLFixture(t)
			db, mock := managementRuntimeDB(t)
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			managementSchemaExpectMetadata(mock, d, m)
			if err := s.CheckRuntimeSchema(context.Background()); err != nil {
				t.Fatal("complete schema warmup failed")
			}
			values := make([]driver.Value, len(tc.args))
			for i, v := range tc.args {
				values[i] = v
			}
			if tc.valid {
				mock.ExpectQuery("^" + regexp.QuoteMeta(tc.query) + "$").WithArgs(values...).WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1)).RowsWillBeClosed()
			}
			var rows []struct{ Protected int }
			err := s.db.Raw(tc.query, tc.args...).Scan(&rows).Error
			if tc.valid {
				if err != nil || len(rows) != 1 {
					t.Errorf("legal nested SQL rejected before driver: %v", err)
				}
			} else if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Errorf("invalid ID did not fail closed before driver: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestBugManagementTraitsSQLTypedNestedBudgets(t *testing.T) {
	for _, tc := range []struct {
		name, kind, id string
		query          string
		valid          bool
	}{
		{"tester_4", "tester", "good", "", true},
		{"tester_5", "tester", "abcde", "", false},
		{"tester_utf8_boundary", "tester", "界x", "", true},
		{"tester_utf8_overflow", "tester", "界界", "", false},
		{"candidate_36", "candidate", strings.Repeat("c", 36), "", true},
		{"candidate_37", "candidate", strings.Repeat("c", 37), "", false},
		{"unknown_kind", "other", "good", "", false},
		// MT-REAL-TESTER-SCOPE: canonical table names are not aliases a second time.
		{"canonical_alias_boundary", "tester", "good", "SELECT 1 AS protected FROM el_mng_result_run r JOIN el_mng_report_revision el_mng_result_run ON el_mng_result_run.run_id = r.id WHERE r.participant_type = ? AND r.participant_id = ?", true},
		{"canonical_alias_overflow", "tester", "abcde", "SELECT 1 AS protected FROM el_mng_result_run r JOIN el_mng_report_revision el_mng_result_run ON el_mng_result_run.run_id = r.id WHERE r.participant_type = ? AND r.participant_id = ?", false},
		{"correlated_boundary", "tester", "good", "SELECT 1 AS protected FROM el_mng_result_run r WHERE r.participant_type = ? AND EXISTS (SELECT 1 FROM el_mng_report_audit a WHERE r.participant_id = ?)", true},
		{"correlated_overflow", "tester", "abcde", "SELECT 1 AS protected FROM el_mng_result_run r WHERE r.participant_type = ? AND EXISTS (SELECT 1 FROM el_mng_report_audit a WHERE r.participant_id = ?)", false},
		{"correlated_invalid_kind", "other", "good", "SELECT 1 AS protected FROM el_mng_result_run r WHERE r.participant_type = ? AND EXISTS (SELECT 1 FROM el_mng_report_audit a WHERE r.participant_id = ?)", false},
		{"type_after_id_boundary", "tester", "good", "SELECT 1 AS protected FROM el_mng_result_run r WHERE r.participant_id = ? AND r.participant_type = ?", true},
		{"type_after_id_overflow", "tester", "abcde", "SELECT 1 AS protected FROM el_mng_result_run r WHERE r.participant_id = ? AND r.participant_type = ?", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, m, _ := managementAuditDDLFixture(t)
			for i := range m.Columns {
				c := &m.Columns[i]
				if c.Table == "el_tester" && c.Name == "id" {
					c.ColumnType = "varchar(4)"
				}
				if c.Table == "el_candidate" && c.Name == "id" {
					c.ColumnType = "varchar(36)"
				}
			}
			db, mock := managementRuntimeDB(t)
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			managementSchemaExpectMetadata(mock, d, m)
			if s.CheckRuntimeSchema(context.Background()) != nil {
				t.Fatal("typed schema warmup failed")
			}
			query := "SELECT 1 AS protected FROM el_mng_report_audit WHERE report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE participant_type = ? AND participant_id = ?))"
			if tc.query != "" {
				query = tc.query
			}
			args := []any{tc.kind, tc.id}
			if strings.HasPrefix(tc.name, "type_after_id_") {
				args = []any{tc.id, tc.kind}
			}
			// Invalid cases also have a successful driver response available: RED
			// must expose a missed capacity check, not an unexpected mock query.
			mock.ExpectQuery("^"+regexp.QuoteMeta(query)+"$").WithArgs(args[0], args[1]).WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			var rows []struct{ Protected int }
			err := s.db.Raw(query, args...).Scan(&rows).Error
			if (tc.valid && (err != nil || len(rows) != 1)) || (!tc.valid && !errors.Is(err, ErrManagementTraitsRuntimeInvalid)) {
				t.Errorf("typed parent budget mismatch: valid=%v err=%v", tc.valid, err)
			}
			if err := mock.ExpectationsWereMet(); tc.valid && err != nil {
				t.Error(err)
			} else if !tc.valid && err == nil {
				t.Error("invalid typed ID reached the successful driver response")
			}
		})
	}
}

// MT-REAL-TESTER-SCOPE: the production guard emits type/id atoms in OR batches.
// Different parent budgets must stay local to each atom, not the whole query.
func TestBugManagementTraitsSQLMixedOwnerBudgets(t *testing.T) {
	for _, tc := range []struct {
		name, candidate, tester string
		valid                   bool
	}{
		{"both_boundaries", strings.Repeat("c", 36), "good", true},
		{"tester_overflow", strings.Repeat("c", 36), "abcde", false},
		{"candidate_overflow", strings.Repeat("c", 37), "good", false},
	} {
		for _, reverse := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/reverse_%v", tc.name, reverse), func(t *testing.T) {
				d, m, _ := managementAuditDDLFixture(t)
				for i := range m.Columns {
					c := &m.Columns[i]
					if c.Table == "el_tester" && c.Name == "id" {
						c.ColumnType = "varchar(4)"
					}
					if c.Table == "el_candidate" && c.Name == "id" {
						c.ColumnType = "varchar(36)"
					}
				}
				db, mock := managementRuntimeDB(t)
				s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
				managementSchemaExpectMetadata(mock, d, m)
				if s.CheckRuntimeSchema(context.Background()) != nil {
					t.Fatal("complete mixed-owner schema rejected")
				}
				query := "SELECT 1 AS protected FROM el_mng_report_audit WHERE report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE (participant_type = ? AND participant_id = ?) OR (participant_type = ? AND participant_id = ?)))"
				args := []any{"candidate", tc.candidate, "tester", tc.tester}
				if reverse {
					args = []any{"tester", tc.tester, "candidate", tc.candidate}
				}
				mock.ExpectQuery("^"+regexp.QuoteMeta(query)+"$").WithArgs(args[0], args[1], args[2], args[3]).WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
				var rows []struct{ Protected int }
				err := s.db.Raw(query, args...).Scan(&rows).Error
				if tc.valid {
					if err != nil || len(rows) != 1 || mock.ExpectationsWereMet() != nil {
						t.Fatal("legal mixed-owner atoms rejected", err)
					}
				} else if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) || mock.ExpectationsWereMet() == nil {
					t.Fatal("mixed-owner overflow reached driver", err)
				}
			})
		}
	}
}

func TestBugManagementTraitsSQLGuardedAuditBindingBatches(t *testing.T) {
	for _, count := range []int{500, 501} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			d, m, names := managementAuditDDLFixture(t)
			db, mock := managementRuntimeDB(t)
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			managementSchemaExpectMetadata(mock, d, m)
			if s.CheckRuntimeSchema(context.Background()) != nil {
				t.Fatal("schema warmup failed")
			}
			managementLegacyMetadata(mock, names...)
			managementAuditExtraColumns(mock, m, false)
			managementSchemaExpectMetadata(mock, d, m)
			managementAuditOrphan(mock, false)
			ids, ownerArgs := make([]string, count), make([]driver.Value, count)
			for i := range ids {
				ids[i] = fmt.Sprintf("17910354687938%05d", i)
				ownerArgs[i] = ids[i]
			}
			mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_tester WHERE").WithArgs(ownerArgs...).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
			for start := 0; start < count; start += 500 {
				end := start + 500
				if end > count {
					end = count
				}
				parts, args := make([]string, 0), make([]driver.Value, 0)
				for _, id := range ids[start:end] {
					parts = append(parts, "(participant_type = ? AND participant_id = ?)")
					args = append(args, "tester", id)
				}
				managementAuditScope(mock, strings.Join(parts, " OR "), args, end == count)
			}
			managementAuditCheck(t, s.db, mock, ManagementTraitsLegacyScopeRequest{TesterIDs: ids}, true, nil)
		})
	}
}
