package service

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MT-GUARD-AUDIT: regression ledger is maintained by the main owner.
// Read the unchanged migration, not invented paper_id columns on audit.
func managementAuditDDLFixture(t *testing.T) (managementTraitsSchemaDefinition, managementTraitsSchemaMetadata, []string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "sql", "management_traits_001_runtime.sql"))
	if err != nil {
		t.Fatal(err)
	}
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	m := managementTraitsSchemaFixture(d)
	blocks := regexp.MustCompile(`(?s)CREATE TABLE IF NOT EXISTS (el_mng_[a-z_]+) \((.*?)\) ENGINE=InnoDB`).FindAllStringSubmatch(strings.ReplaceAll(string(raw), "\r\n", "\n"), -1)
	if len(blocks) != 11 || len(d.ForeignKeys) != 15 {
		t.Fatal("expected unchanged 11-table/15-FK migration")
	}
	columns := regexp.MustCompile("(?m)^`([a-z_]+)` ([a-z]+(?:\\([0-9,]+\\))?) (NOT NULL|NULL),$")
	names := make([]string, 0, len(blocks))
	ddlColumns := make([]managementTraitsSchemaColumnRow, 0)
	var auditColumns []string
	for _, block := range blocks {
		names = append(names, block[1])
		for _, c := range columns.FindAllStringSubmatch(block[2], -1) {
			r := managementTraitsSchemaColumnRow{Table: block[1], Name: c[1], ColumnType: c[2], Nullable: "NO"}
			if c[3] == "NULL" {
				r.Nullable = "YES"
			}
			if strings.HasPrefix(c[2], "varchar") || strings.HasPrefix(c[2], "char") || c[2] == "longtext" {
				r.Charset, r.Collation = "utf8mb4", "utf8mb4_bin"
			}
			ddlColumns = append(ddlColumns, r)
			if block[1] == "el_mng_report_audit" {
				auditColumns = append(auditColumns, c[1])
			}
		}
		for _, f := range d.ForeignKeys {
			if f.Table == block[1] && !strings.Contains(block[2], "CONSTRAINT "+f.Name+" FOREIGN KEY("+strings.Join(f.Columns, ",")+") REFERENCES "+f.RefTable+"("+strings.Join(f.RefColumns, ",")+") ON DELETE RESTRICT ON UPDATE RESTRICT") {
				t.Fatal("fixture FK differs from migration", f.Name)
			}
		}
	}
	if !reflect.DeepEqual(auditColumns, []string{"id", "report_id", "actor_id", "action", "created_at"}) {
		t.Fatal("audit must have exactly the real five columns", auditColumns)
	}
	for _, c := range m.Columns {
		if !strings.HasPrefix(c.Table, "el_mng_") {
			ddlColumns = append(ddlColumns, c)
		}
	}
	m.Columns = ddlColumns
	if err := validateManagementTraitsSchema(d, m); err != nil {
		t.Fatal("DDL-derived fixture rejected", err)
	}
	sort.Strings(names)
	return d, m, names
}

func managementAuditExtraColumns(mock sqlmock.Sqlmock, m managementTraitsSchemaMetadata, capture bool) {
	rows := sqlmock.NewRows([]string{"table_name", "column_name"})
	for _, c := range m.Columns {
		if c.Table == "el_mng_report_audit" || c.Table == "el_mng_report_current" || c.Table == "el_mng_report_revision" || c.Table == "el_mng_runtime_receipt" {
			rows.AddRow(c.Table, c.Name)
		}
	}
	args := []driver.Value{"el_mng_report_audit", "el_mng_report_current", "el_mng_report_revision", "el_mng_runtime_receipt"}
	if capture {
		rows.AddRow("el_mng_owned_capture", "participant_type").AddRow("el_mng_owned_capture", "participant_id").AddRow("el_mng_owned_capture", "paper_id")
		args = append([]driver.Value{"el_mng_owned_capture"}, args...)
	}
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WithArgs(args...).WillReturnRows(rows).RowsWillBeClosed()
}

// Independent SQL oracle: audit.report_id -> revision.id -> run.id, with the
// actual composite report/run identity. Orphans cannot disappear in INNER JOINs.
const managementAuditOrphanSQL = "SELECT 1 AS protected FROM el_mng_report_audit a LEFT JOIN el_mng_report_revision r ON r.id = a.report_id LEFT JOIN el_mng_result_run u ON u.id = r.run_id AND u.paper_id = r.paper_id AND u.exam_id = r.exam_id WHERE r.id IS NULL OR u.id IS NULL LIMIT 1"

func managementAuditOrphan(mock sqlmock.Sqlmock, found bool) {
	rows := sqlmock.NewRows([]string{"protected"})
	if found {
		rows.AddRow(1)
	}
	mock.ExpectQuery("^" + regexp.QuoteMeta(managementAuditOrphanSQL) + "$").WithoutArgs().WillReturnRows(rows).RowsWillBeClosed()
}

func managementAuditScope(mock sqlmock.Sqlmock, where string, args []driver.Value, found bool) {
	query := "SELECT 1 AS protected FROM el_mng_report_audit WHERE (report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE " + where + "))) LIMIT 1"
	rows := sqlmock.NewRows([]string{"protected"})
	if found {
		rows.AddRow(1)
	}
	mock.ExpectQuery("^" + regexp.QuoteMeta(query) + "$").WithArgs(args...).WillReturnRows(rows).RowsWillBeClosed()
}

func managementAuditSetup(t *testing.T, capture bool) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	d, m, names := managementAuditDDLFixture(t)
	db, mock := managementRuntimeDB(t)
	if capture {
		names = append(names, "el_mng_owned_capture")
	}
	managementLegacyMetadata(mock, names...)
	managementAuditExtraColumns(mock, m, capture)
	managementSchemaExpectMetadata(mock, d, m)
	managementAuditOrphan(mock, false)
	return db, mock
}

func managementAuditCheck(t *testing.T, db *gorm.DB, mock sqlmock.Sqlmock, req ManagementTraitsLegacyScopeRequest, want bool, wantErr error) {
	t.Helper()
	got, err := CheckManagementTraitsLegacyScope(context.Background(), db, req)
	if got != want || !errors.Is(err, wantErr) {
		t.Errorf("protected=%v error=%v; want %v/%v", got, err, want, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestBugManagementTraitsGuardAuditCompleteEmptyAndCapture(t *testing.T) {
	for _, mode := range []string{"empty", "all-empty", "capture"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := managementAuditSetup(t, mode == "capture")
			req := ManagementTraitsLegacyScopeRequest{}
			if mode == "all-empty" {
				req.AllLegacy = true
				for _, table := range []string{"el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_report_audit", "el_mng_report_current", "el_mng_report_revision", "el_mng_runtime_receipt"} {
					managementLegacyAllExists(mock, table, false, false)
				}
			}
			if mode == "capture" {
				req.CandidateIDs = []string{"C"}
				managementLegacyEmpty(mock, "el_candidate")
				managementAuditScope(mock, "(participant_type = ? AND participant_id = ?)", []driver.Value{"candidate", "C"}, false)
				for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
					mock.ExpectQuery("SELECT 1 AS protected FROM "+table+" WHERE").WithArgs("candidate", "C").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
				}
				mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_owned_capture WHERE").WithArgs("candidate", "C").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			}
			managementAuditCheck(t, db, mock, req, mode == "capture", nil)
		})
	}
}

func TestBugManagementTraitsGuardAuditClosure(t *testing.T) {
	for _, kind := range []string{"paper", "exam", "candidate", "tester", "question", "pdf", "malicious-paper"} {
		t.Run(kind, func(t *testing.T) {
			db, mock := managementAuditSetup(t, false)
			req := ManagementTraitsLegacyScopeRequest{}
			where, args := "(paper_id = ?)", []driver.Value{"P"}
			switch kind {
			case "candidate", "tester":
				if kind == "candidate" {
					req.CandidateIDs = []string{"C"}
				} else {
					req.TesterIDs = []string{"C"}
				}
				managementLegacyEmpty(mock, "el_"+kind)
				where, args = "(participant_type = ? AND participant_id = ?)", []driver.Value{kind, "C"}
			case "exam":
				req.ExamIDs = []string{"E"}
				for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
					managementLegacyEmpty(mock, table)
				}
				where, args = "(exam_id = ?)", []driver.Value{"E"}
			case "pdf":
				req.PDFPaths = []string{`reports\exact.pdf`}
				mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE").WithArgs("reports/exact.pdf").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("C", "E", "P"))
				mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_tester WHERE").WithArgs("reports/exact.pdf").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
				for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
					managementLegacyEmpty(mock, table)
				}
				where, args = "(paper_id = ?) OR (participant_type = ? AND participant_id = ?)", []driver.Value{"P", "candidate", "C"}
			default:
				paper := "P"
				if kind == "malicious-paper" {
					paper = "P' OR 1=1; --"
				}
				if kind == "question" {
					req.PaperQuestionIDs = []string{"Q"}
					mock.ExpectQuery("SELECT paper_id FROM el_paper_qu WHERE").WithArgs("Q").WillReturnRows(sqlmock.NewRows([]string{"paper_id"}).AddRow(paper))
				} else {
					req.PaperIDs = []string{paper}
				}
				for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
					mock.ExpectQuery("SELECT .* FROM " + table + " WHERE").WithArgs(paper).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
				}
				args = []driver.Value{paper}
			}
			managementAuditScope(mock, where, args, true)
			managementAuditCheck(t, db, mock, req, true, nil)
		})
	}
}

func TestBugManagementTraitsGuardAuditSiblingPrecision(t *testing.T) {
	db, mock := managementAuditSetup(t, false)
	mock.ExpectQuery("SELECT id, exam_id FROM el_paper WHERE").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}).AddRow("A", "E"))
	for _, table := range []string{"el_candidate", "el_tester"} {
		mock.ExpectQuery("SELECT id, exam_id, paper_id FROM " + table + " WHERE").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
	}
	// Only sibling B/E has report/run history; inferred E is not a target exam.
	managementAuditScope(mock, "(paper_id = ?)", []driver.Value{"A"}, false)
	mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE").WithArgs("E").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_report_current"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM " + table + " WHERE").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	}
	for _, table := range []string{"el_mng_report_revision", "el_mng_runtime_receipt"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM " + table + " WHERE.*run_id IN").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
		mock.ExpectQuery("SELECT 1 AS protected FROM " + table + " WHERE").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	}
	managementAuditCheck(t, db, mock, ManagementTraitsLegacyScopeRequest{PaperIDs: []string{"A"}}, false, nil)
}

func TestBugManagementTraitsGuardAuditPublicIdentityCapture(t *testing.T) {
	d, m, names := managementAuditDDLFixture(t)
	db, mock := managementRuntimeDB(t)
	for _, table := range []string{"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_result_dimension", "el_mng_result_module"} {
		mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WithArgs(table).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("el_mng_owned_capture"))
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("E").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT .*el_candidate.*id =").WithArgs("C", 2).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_owned_capture", "el_mng_result_run"} {
		count := 0
		if table == "el_mng_owned_capture" {
			count = 1
		}
		mock.ExpectQuery("SELECT count.*"+table).WithArgs("candidate", "C").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	}
	managementLegacyMetadata(mock, append(names, "el_mng_owned_capture")...)
	managementAuditExtraColumns(mock, m, true)
	managementSchemaExpectMetadata(mock, d, m)
	managementAuditOrphan(mock, false)
	managementLegacyEmpty(mock, "el_candidate")
	managementAuditScope(mock, "(participant_type = ? AND participant_id = ?)", []driver.Value{"candidate", "C"}, false)
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM "+table+" WHERE").WithArgs("candidate", "C").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	}
	mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_owned_capture WHERE").WithArgs("candidate", "C").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
	got, err := ManagementTraitsIdentityScope(context.Background(), db, "E", "candidate", "C")
	if !got || err != nil {
		t.Errorf("actual public identity capture blocked: %v/%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestBugManagementTraitsGuardAuditOrphansAndErrors(t *testing.T) {
	for _, kind := range []string{"revision-missing", "run-missing", "identity-mismatch", "query", "row", "scan"} {
		for _, scope := range []string{"empty", "all", "mismatch"} {
			t.Run(kind+"/"+scope, func(t *testing.T) {
				d, m, names := managementAuditDDLFixture(t)
				db, mock := managementRuntimeDB(t)
				managementLegacyMetadata(mock, names...)
				managementAuditExtraColumns(mock, m, false)
				managementSchemaExpectMetadata(mock, d, m)
				q := mock.ExpectQuery("^" + regexp.QuoteMeta(managementAuditOrphanSQL) + "$").WithoutArgs()
				switch kind {
				case "query":
					q.WillReturnError(errors.New("private-dsn"))
				case "row":
					q.WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1).RowError(0, errors.New("private-dsn")))
				case "scan":
					q.WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow("not-an-int"))
				default:
					q.WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
				}
				req := ManagementTraitsLegacyScopeRequest{AllLegacy: scope == "all"}
				if scope == "mismatch" {
					req.PaperIDs = []string{"unrelated"}
				}
				managementAuditCheck(t, db, mock, req, false, ErrManagementTraitsRuntimeInvalid)
			})
		}
	}
}

func TestBugManagementTraitsGuardAuditFullGateDrift(t *testing.T) {
	for _, kind := range []string{"type", "nullable", "collation", "index", "fk-target", "fk-action", "fk-order"} {
		t.Run(kind, func(t *testing.T) {
			d, m, names := managementAuditDDLFixture(t)
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, names...)
			managementAuditExtraColumns(mock, m, false)
			for i := range m.Columns {
				if m.Columns[i].Table == "el_mng_report_audit" && m.Columns[i].Name == "report_id" {
					switch kind {
					case "type":
						m.Columns[i].ColumnType = "varchar(63)"
					case "nullable":
						m.Columns[i].Nullable = "YES"
					case "collation":
						m.Columns[i].Collation = "utf8mb4_general_ci"
					}
				}
			}
			for i := range m.Indexes {
				if kind == "index" && m.Indexes[i].Name == "idx_mng_audit_report" {
					m.Indexes[i].Prefix = 1
				}
			}
			for i := range m.ForeignKeys {
				if m.ForeignKeys[i].Name == "fk_mng_audit_report" {
					switch kind {
					case "fk-target":
						m.ForeignKeys[i].ReferencedTable = "el_paper"
					case "fk-action":
						m.ForeignKeys[i].DeleteRule = "CASCADE"
					case "fk-order":
						m.ForeignKeys[i].Position = 2
					}
				}
			}
			managementSchemaExpectMetadata(mock, d, m)
			managementAuditCheck(t, db, mock, ManagementTraitsLegacyScopeRequest{}, false, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

func TestBugManagementTraitsGuardAuditPartialAndUnknown(t *testing.T) {
	for _, kind := range []string{"audit-only", "missing-revision", "missing-current", "missing-receipt", "unknown-report-id", "missing-column", "extra-column", "duplicate-column", "injected-column", "unexpected-table"} {
		t.Run(kind, func(t *testing.T) {
			_, m, names := managementAuditDDLFixture(t)
			db, mock := managementRuntimeDB(t)
			if kind == "audit-only" {
				names = []string{"el_mng_report_audit"}
			}
			missing := map[string]string{"missing-revision": "el_mng_report_revision", "missing-current": "el_mng_report_current", "missing-receipt": "el_mng_runtime_receipt"}[kind]
			filtered := make([]string, 0)
			for _, n := range names {
				if n != missing {
					filtered = append(filtered, n)
				}
			}
			if kind == "unknown-report-id" {
				managementLegacyMetadata(mock, "el_mng_arbitrary_marker")
				mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name").WithArgs("el_mng_arbitrary_marker").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_arbitrary_marker", "report_id"))
			} else {
				managementLegacyMetadata(mock, filtered...)
				rows := sqlmock.NewRows([]string{"table_name", "column_name"})
				args := make([]driver.Value, 0)
				for _, name := range filtered {
					if name == "el_mng_report_audit" || name == "el_mng_report_revision" || name == "el_mng_report_current" || name == "el_mng_runtime_receipt" {
						args = append(args, name)
						for _, c := range m.Columns {
							if c.Table == name && !(kind == "missing-column" && name == "el_mng_report_audit" && c.Name == "actor_id") {
								rows.AddRow(name, c.Name)
							}
						}
					}
				}
				switch kind {
				case "extra-column":
					rows.AddRow("el_mng_report_audit", "paper_id")
				case "duplicate-column":
					rows.AddRow("el_mng_report_audit", "report_id")
				case "injected-column":
					rows.AddRow("el_mng_report_audit", "report_id;DROP")
				case "unexpected-table":
					rows.AddRow("el_mng_not_requested", "paper_id")
				}
				mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name").WithArgs(args...).WillReturnRows(rows)
			}
			managementAuditCheck(t, db, mock, ManagementTraitsLegacyScopeRequest{AllLegacy: true}, false, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

func TestBugManagementTraitsGuardAuditBoundedBatches(t *testing.T) {
	db, mock := managementAuditSetup(t, false)
	ids := make([]string, 1001)
	for i := range ids {
		ids[i] = fmt.Sprintf("C%04d", i)
	}
	for start := 0; start < len(ids); start += 1000 {
		end := start + 1000
		if end > len(ids) {
			end = len(ids)
		}
		args := make([]driver.Value, 0, end-start)
		for _, id := range ids[start:end] {
			args = append(args, id)
		}
		mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE").WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
	}
	// Each typed owner uses two bindings: 500 + 500 + 1, never >1000 values.
	for start := 0; start < len(ids); start += 500 {
		end := start + 500
		if end > len(ids) {
			end = len(ids)
		}
		parts, args := make([]string, 0), make([]driver.Value, 0)
		for _, id := range ids[start:end] {
			parts = append(parts, "(participant_type = ? AND participant_id = ?)")
			args = append(args, "candidate", id)
		}
		managementAuditScope(mock, strings.Join(parts, " OR "), args, end == len(ids))
	}
	managementAuditCheck(t, db, mock, ManagementTraitsLegacyScopeRequest{CandidateIDs: ids}, true, nil)
}

func TestBugManagementTraitsGuardAuditFullGateQueryErrors(t *testing.T) {
	queries := []string{"SELECT table_name AS table_name, engine AS engine", "SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type", "SELECT table_name AS table_name, index_name AS index_name", "SELECT k.table_name AS table_name, k.constraint_name AS constraint_name"}
	for stage := range queries {
		t.Run(fmt.Sprint(stage), func(t *testing.T) {
			d, m, names := managementAuditDDLFixture(t)
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, names...)
			managementAuditExtraColumns(mock, m, false)
			if stage > 0 {
				mock.ExpectQuery(queries[0]).WillReturnRows(managementRuntimeModelRows(t, managementTraitsSchemaTableRows(d)...))
			}
			if stage > 1 {
				values := make([]any, len(m.Columns))
				for i := range values {
					values[i] = m.Columns[i]
				}
				mock.ExpectQuery(queries[1]).WillReturnRows(managementRuntimeModelRows(t, values...))
			}
			if stage > 2 {
				values := make([]any, len(m.Indexes))
				for i := range values {
					values[i] = m.Indexes[i]
				}
				mock.ExpectQuery(queries[2]).WillReturnRows(managementRuntimeModelRows(t, values...))
			}
			mock.ExpectQuery(queries[stage]).WillReturnError(errors.New("private metadata failure"))
			managementAuditCheck(t, db, mock, ManagementTraitsLegacyScopeRequest{}, false, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

func TestBugManagementTraitsGuardAuditScopedQueryErrors(t *testing.T) {
	for _, failure := range []string{"query", "row", "scan"} {
		t.Run(failure, func(t *testing.T) {
			db, mock := managementAuditSetup(t, false)
			managementLegacyEmpty(mock, "el_candidate")
			q := mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_report_audit WHERE").WithArgs("candidate", "C")
			switch failure {
			case "query":
				q.WillReturnError(errors.New("private audit failure"))
			case "row":
				q.WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1).RowError(0, errors.New("private audit failure")))
			case "scan":
				q.WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow("invalid"))
			}
			managementAuditCheck(t, db, mock, ManagementTraitsLegacyScopeRequest{CandidateIDs: []string{"C"}}, false, ErrManagementTraitsRuntimeInvalid)
		})
	}
}

// Explicit opt-in only. The next owner restores legacy tables and applies the
// unchanged 001 DDL externally. No CREATE/DROP/ALTER, SSH, or main-schema writes.
func TestManagementTraitsGuardAuditMySQLExternal(t *testing.T) {
	dsn, owned := os.Getenv("MNG_GUARD_AUDIT_MYSQL_DSN"), os.Getenv("MNG_GUARD_AUDIT_MYSQL_SCHEMA")
	if dsn == "" {
		t.Skip("dedicated MNG_GUARD_AUDIT_MYSQL_DSN absent; no real DB validation")
	}
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil || !regexp.MustCompile(`^mng_guard_audit_test_[0-9a-f]{16}$`).MatchString(owned) || cfg.DBName != owned || cfg.MultiStatements || cfg.InterpolateParams {
		t.Fatal("dedicated owned schema/DSN assertion failed")
	}
	cfg.ParseTime, cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = true, 10*time.Second, 15*time.Second, 15*time.Second
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: cfg.FormatDSN(), SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal("dedicated database open failed")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal("dedicated pool unavailable")
	}
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	db = db.WithContext(ctx)
	var actual string
	if db.Raw("SELECT DATABASE()").Scan(&actual).Error != nil || actual != owned {
		t.Fatal("actual schema ownership mismatch")
	}
	var crossSchemaFK int64
	if db.Raw("SELECT COUNT(*) FROM information_schema.key_column_usage WHERE table_schema=DATABASE() AND referenced_table_schema IS NOT NULL AND referenced_table_schema<>DATABASE()").Scan(&crossSchemaFK).Error != nil || crossSchemaFK != 0 {
		t.Fatal("owned restore contains cross-schema foreign keys")
	}
	d, _, _ := managementAuditDDLFixture(t)
	if (&ManagementTraitsRuntimeService{db: db}).checkRuntimeSchema(ctx) != nil {
		t.Fatal("externally restored complete schema rejected")
	}
	for _, table := range d.Tables {
		var n int64
		if db.Table(table.Name).Count(&n).Error != nil || n != 0 {
			t.Fatal("owned sidecar tables must initially be empty")
		}
	}
	for _, req := range []ManagementTraitsLegacyScopeRequest{{}, {AllLegacy: true}} {
		if protected, err := CheckManagementTraitsLegacyScope(ctx, db, req); protected || err != nil {
			t.Fatal("complete empty schema blocked")
		}
	}
	var owner struct{ ID, ExamID, PaperID string }
	if db.Raw("SELECT c.id, c.exam_id, c.paper_id FROM el_candidate c JOIN el_paper p ON p.id=c.paper_id AND p.exam_id=c.exam_id JOIN el_exam e ON e.id=c.exam_id ORDER BY c.id LIMIT 1").Scan(&owner).Error != nil || owner.ID == "" || owner.PaperID == "" {
		t.Fatal("restored legacy candidate/paper/exam fixture required")
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal("owned rollback transaction unavailable")
	}
	rolledBack := false
	defer func() {
		if !rolledBack && tx.Rollback().Error != nil {
			t.Error("owned transaction rollback failed")
		}
	}()
	now, id, hash := time.Now(), uuid.NewString(), strings.Repeat("a", 64)
	bundle := model.ManagementTraitsDefinitionBundle{ID: id, ProductVersion: id, QuestionVersion: id, ScoringVersion: id, NormVersion: id, ScoringManifest: "{}", ScoringManifestSHA: hash, CreatedAt: now}
	snapshot := model.ManagementTraitsPaperSnapshot{PaperID: owner.PaperID, ExamID: owner.ExamID, BundleID: id, ParticipantType: "candidate", ParticipantID: owner.ID, EvidenceSnapshot: "{}", MappingSnapshot: "{}", ParticipantSnapshot: "{}", FieldContract: "{}", EvidenceSHA: hash, MappingSHA: hash, ScoringManifestSHA: hash, SourceCapturedAt: now, StartedAt: now, CreatedAt: now}
	run := model.ManagementTraitsResultRun{ID: uuid.NewString(), PaperID: owner.PaperID, ExamID: owner.ExamID, ParticipantType: "candidate", ParticipantID: owner.ID, ScoringVersion: id, NormVersion: id, ScoringManifestSHA: hash, InputSHA: hash, CreatedAt: now}
	revision := model.ManagementTraitsReportRevision{ID: uuid.NewString(), RunID: run.ID, PaperID: owner.PaperID, ExamID: owner.ExamID, Revision: 1, Mode: "test", ContentSHA: hash, TemplateSHA: hash, DataSHA: hash, FileSHA: hash, DataSnapshot: "{}", CreatedAt: now}
	audit := model.ManagementTraitsReportAudit{ID: uuid.NewString(), ReportID: revision.ID, ActorID: 1, Action: "generate", CreatedAt: now}
	for _, row := range []any{&bundle, &snapshot, &run, &revision, &audit} {
		if tx.Create(row).Error != nil {
			t.Fatal("owned audit fixture insert failed")
		}
	}
	for _, req := range []ManagementTraitsLegacyScopeRequest{{PaperIDs: []string{owner.PaperID}}, {ExamIDs: []string{owner.ExamID}}, {CandidateIDs: []string{owner.ID}}, {AllLegacy: true}} {
		if protected, err := CheckManagementTraitsLegacyScope(ctx, tx, req); !protected || err != nil {
			t.Fatal("real audit closure not protected")
		}
	}
	if protected, err := ManagementTraitsIdentityScope(ctx, tx, owner.ExamID, "candidate", owner.ID); !protected || err != nil {
		t.Fatal("real public identity audit scope blocked")
	}
	if tx.Rollback().Error != nil {
		t.Fatal("owned transaction rollback failed")
	}
	rolledBack = true
	for _, table := range d.Tables {
		var n int64
		if db.Table(table.Name).Count(&n).Error != nil || n != 0 {
			t.Fatal("owned rollback left sidecar rows")
		}
	}
}
