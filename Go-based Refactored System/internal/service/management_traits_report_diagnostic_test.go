package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const managementDiagnosticSecret = "INJECTED_SECRET_SQL_DSN_TOKEN_PII_PATH"

func managementDiagnosticCapture(t *testing.T) *bytes.Buffer {
	t.Helper()
	var out bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&out, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &out
}

// Regression MT-REPORT-DIAG-01: real GenerateTestReport failure branches must
// identify stage/class without changing sentinels, transactions or file cleanup.
// Before diagnostics these sqlmock/renderer failures return no stage log.
func TestBugMTReportDiag_GenerateFailureStages(t *testing.T) {
	for _, name := range []string{"schema", "load", "load-commit", "render", "render-deadline", "render-converter", "write", "reload", "revision", "persist-revision", "persist-current", "persist-audit", "persist-commit", "success"} {
		t.Run(name, func(t *testing.T) {
			out := managementDiagnosticCapture(t)
			f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
			wb, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "260929管理特质测评-优化", "260928测评内容+数据图.xlsx"))
			if err != nil {
				t.Fatal("fixture unavailable")
			}
			content, err := LoadManagementTraitsTestContent(wb)
			if err != nil {
				t.Fatal("fixture content invalid")
			}
			db, mock := managementRuntimeDB(t)
			db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			injected := fmt.Errorf("%s: %w", managementDiagnosticSecret, &mysql.MySQLError{Number: 1205, Message: managementDiagnosticSecret})
			stage, class := name, "mysql_1205"
			if name == "schema" {
				mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
				class = "runtime_closed"
			} else {
				managementRuntimeExpectSchema(t, mock)
				if name == "load" {
					mock.ExpectBegin()
					mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnError(injected)
					mock.ExpectRollback()
				} else {
					managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
					mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
					if name == "load-commit" {
						mock.ExpectCommit().WillReturnError(injected)
						stage = "load"
					} else {
						mock.ExpectCommit()
					}
					if strings.HasPrefix(name, "persist-") || name == "revision" || name == "success" {
						managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
						mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
						if name == "revision" {
							mock.ExpectQuery("SELECT COALESCE\\(MAX\\(revision\\),0\\)").WillReturnError(injected)
							mock.ExpectRollback()
						} else {
							mock.ExpectQuery("SELECT COALESCE\\(MAX\\(revision\\),0\\)").WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(0))
							for _, part := range []string{"revision", "current", "audit"} {
								e := mock.ExpectExec("INSERT INTO `el_mng_report_" + part + "`")
								if name == "persist-"+part {
									e.WillReturnError(injected)
									break
								}
								e.WillReturnResult(sqlmock.NewResult(1, 1))
							}
							if name == "persist-commit" {
								mock.ExpectCommit().WillReturnError(injected)
							} else if name != "success" {
								mock.ExpectRollback()
							} else {
								mock.ExpectCommit()
							}
							stage = "persist"
						}
					} else if name == "reload" {
						mock.ExpectBegin()
						mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnError(injected)
						mock.ExpectRollback()
					}
				}
			}
			root := t.TempDir()
			legacy := filepath.Join(root, "historical.pdf")
			if os.WriteFile(legacy, []byte("historical"), 0600) != nil {
				t.Fatal("fixture write failed")
			}
			if name == "write" {
				root = legacy
				class = "os_path"
			}
			calls := 0
			got, err := s.GenerateTestReport(context.Background(), records.Run.ID, 1, root, content, func(ctx context.Context, dto ManagementTraitsTestReportData) ([]byte, error) {
				calls++
				switch name {
				case "render":
					stage, class = "render", "unknown"
					return nil, fmt.Errorf("%s: %w", managementDiagnosticSecret, errors.New(managementDiagnosticSecret))
				case "render-deadline":
					stage, class = "render", "context_deadline"
					return nil, fmt.Errorf("%s: %w", managementDiagnosticSecret, context.DeadlineExceeded)
				case "render-converter":
					stage, class = "render", "lo_input_name"
					return libreofficepdf.NewClient("").Convert(ctx, "../"+managementDiagnosticSecret+".docx", []byte("docx"))
				}
				return append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 2048)...), nil
			})
			if name == "success" {
				if err != nil || got.Revision != 1 || out.Len() != 0 {
					t.Fatal("success behavior/log changed")
				}
			} else {
				wantErr := ErrManagementTraitsRuntimeInvalid
				if name == "schema" {
					wantErr = ErrManagementTraitsRuntimeClosed
				}
				if err != wantErr {
					t.Fatal("returned sentinel changed")
				}
				lines := strings.Split(strings.TrimSpace(out.String()), "\n")
				if len(lines) != 1 {
					t.Fatal("expected one failure diagnostic")
				}
				var entry map[string]any
				if json.Unmarshal([]byte(lines[0]), &entry) != nil || entry["msg"] != "management_traits_report_failure" || entry["stage"] != stage || entry["class"] != class {
					t.Fatalf("missing fixed diagnostic for %s", name)
				}
				if len(entry) != 5 || strings.Contains(out.String(), managementDiagnosticSecret) || strings.Contains(out.String(), records.Run.ID) || strings.Contains(out.String(), owner.Name) || strings.Contains(out.String(), root) {
					t.Fatal("diagnostic disclosed private data")
				}
			}
			if (name == "schema" || name == "load" || name == "load-commit") && calls != 0 {
				t.Fatal("renderer reached before load success")
			}
			old, readErr := os.ReadFile(legacy)
			if readErr != nil || string(old) != "historical" {
				t.Fatal("old file changed")
			}
			files, readErr := os.ReadDir(filepath.Dir(legacy))
			wantFiles := 1
			if name == "success" {
				wantFiles = 2
			}
			if readErr != nil || len(files) != wantFiles {
				t.Fatal("new file cleanup changed")
			}
			if mock.ExpectationsWereMet() != nil {
				t.Fatal("transaction/query contract changed")
			}
		})
	}
}
