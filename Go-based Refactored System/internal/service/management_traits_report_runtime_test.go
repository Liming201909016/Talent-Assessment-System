package service

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

func TestManagementTraitsReportFileImmutableAndConfined(t *testing.T) {
	root := t.TempDir()
	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 2048)...)
	key, err := writeManagementTraitsTestPDF(root, "22222222-2222-4222-8222-222222222222", pdf, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writeManagementTraitsTestPDF(root, "22222222-2222-4222-8222-222222222222", pdf, nil); err == nil {
		t.Fatal("immutable file overwritten")
	}
	for _, bad := range []string{"../escape.pdf", "/absolute.pdf", "one/../file.pdf", "old.pdf"} {
		if _, err := managementTraitsReportFile(root, bad); err == nil {
			t.Fatal("unsafe key accepted", bad)
		}
	}
	got, err := os.ReadFile(filepath.Join(root, key))
	if err != nil || !bytes.Equal(got, pdf) {
		t.Fatal("old file changed", err)
	}
	if _, err := writeManagementTraitsTestPDF(root, "22222222-2222-4222-8222-222222222222", []byte("not PDF"), nil); err == nil {
		t.Fatal("non-PDF accepted")
	}
}

func TestManagementTraitsReportMetadataAuditFailureRollsBack(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `el_mng_report_revision`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `el_mng_report_current`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `el_mng_report_audit`").WillReturnError(errors.New("private audit failure"))
	mock.ExpectRollback()
	err := db.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		return persistManagementTraitsTestReport(tx, model.ManagementTraitsReportRevision{ID: "report", PaperID: "paper", RunID: "run", Mode: "test"}, model.ManagementTraitsReportAudit{ID: "audit", ReportID: "report"}, nil)
	})
	if err == nil {
		t.Fatal("audit failure committed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
