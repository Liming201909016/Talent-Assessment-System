package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestManagementTraitsGenerateTestReportFullTransactions(t *testing.T) {
	for _, failure := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback-cleanup"}[failure], func(t *testing.T) {
			f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
			workbook, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "260929管理特质测评-优化", "260928测评内容+数据图.xlsx"))
			if err != nil {
				t.Fatal(err)
			}
			content, err := LoadManagementTraitsTestContent(workbook)
			if err != nil {
				t.Fatal(err)
			}
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
			mock.ExpectCommit()
			// Ordered sentinel proves the first DB transaction was committed before
			// the external renderer was called (not merely source-code placement).
			mock.ExpectExec("MNG_RENDER_OUTSIDE_TRANSACTION").WillReturnResult(sqlmock.NewResult(0, 0))
			managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
			mock.ExpectQuery("SELECT COALESCE\\(MAX\\(revision\\),0\\)").WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(2))
			mock.ExpectExec("INSERT INTO `el_mng_report_revision`").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("INSERT INTO `el_mng_report_current`").WillReturnResult(sqlmock.NewResult(1, 1))
			if failure {
				mock.ExpectExec("INSERT INTO `el_mng_report_audit`").WillReturnError(errors.New("private audit failure"))
				mock.ExpectRollback()
			} else {
				mock.ExpectExec("INSERT INTO `el_mng_report_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}
			root := t.TempDir()
			legacy := filepath.Join(root, "historical.pdf")
			old := []byte("historical PDF sentinel")
			if err := os.WriteFile(legacy, old, 0600); err != nil {
				t.Fatal(err)
			}
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("test-only-pdf"), 128)...)
			templateSHA := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
			r, err := s.GenerateTestReportWithTemplateSHA(ctx, records.Run.ID, 1, root, content, templateSHA, func(renderCtx context.Context, dto ManagementTraitsTestReportData) ([]byte, error) {
				if renderCtx != ctx || dto.Overall.Score != "50.00" || dto.Participant.Name != owner.Name || len(dto.Dimensions) != 13 || len(dto.Modules) != 4 || !dto.TestOnly || dto.ContentSourceSHA != content.SourceSHA() {
					t.Fatal("wrong DTO/context")
				}
				if output := os.Getenv("MNG_TEST_REPORT_DATA_OUTPUT"); output != "" && !failure {
					raw, err := json.Marshal(dto)
					if err != nil {
						t.Fatal(err)
					}
					file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
					if err != nil {
						t.Fatal(err)
					}
					_, err = file.Write(raw)
					closeErr := file.Close()
					if err != nil || closeErr != nil {
						t.Fatal("cannot write isolated TEST DTO")
					}
				}
				if err := db.Exec("MNG_RENDER_OUTSIDE_TRANSACTION").Error; err != nil {
					t.Fatal(err)
				}
				return pdf, nil
			})
			if failure && err == nil {
				t.Fatal("failed audit accepted")
			}
			if !failure && (err != nil || r.Revision != 3 || r.Mode != "test" || r.TestLabel != ManagementTraitsTestPurposeLabel || r.TemplateSHA != templateSHA) {
				t.Fatal("pipeline failed", err)
			}
			got, err := os.ReadFile(legacy)
			if err != nil || !bytes.Equal(got, old) {
				t.Fatal("historical PDF overwritten")
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			want := 2
			if failure {
				want = 1
			}
			if len(entries) != want {
				t.Fatal("rollback orphan or missing new report", len(entries))
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
