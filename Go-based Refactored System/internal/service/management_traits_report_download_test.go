package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestManagementTraitsReportIDDownloadChecksFactsAndFile(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(map[bool]string{false: "download", true: "corrupt-file-rejected"}[corrupt], func(t *testing.T) {
			f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
			wb, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "260929管理特质测评-优化", "260928测评内容+数据图.xlsx"))
			if err != nil {
				t.Fatal(err)
			}
			content, err := LoadManagementTraitsTestContent(wb)
			if err != nil {
				t.Fatal(err)
			}
			dto, err := BuildManagementTraitsTestReport(f.Bundle, f.Paper, f.Questions, records.Run, records.Dimensions, records.Modules, records.Receipt, content, 1<<20)
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(dto)
			dh := sha256.Sum256(raw)
			pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("test-only"), 200)...)
			ph := sha256.Sum256(pdf)
			root := t.TempDir()
			id := "33333333-3333-4333-8333-333333333333"
			key, err := writeManagementTraitsTestPDF(root, id, pdf, nil)
			if err != nil {
				t.Fatal(err)
			}
			r := model.ManagementTraitsReportRevision{ID: id, RunID: records.Run.ID, PaperID: paper.ID, ExamID: paper.ExamID, Revision: 1, Mode: "test", TestTitle: ManagementTraitsTestPurposeTitle, TestLabel: ManagementTraitsTestPurposeLabel, ContentVersion: "mng-customer-text-test-v1", ContentSHA: content.SourceSHA(), TemplateVersion: ManagementTraitsTestTemplateVersion, TemplateSHA: managementTraitsReportTemplateSHA, DataSnapshot: string(raw), DataSHA: hex.EncodeToString(dh[:]), FileKey: key, FileSHA: hex.EncodeToString(ph[:]), FileBytes: int64(len(pdf)), CreatedBy: 1, CreatedAt: time.Now()}
			if corrupt {
				pdf[len(pdf)-1] ^= 1
				if err := os.WriteFile(filepath.Join(root, key), pdf, 0600); err != nil {
					t.Fatal(err)
				}
			}
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_mng_report_revision").WillReturnRows(managementRuntimeModelRows(t, r))
			mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(managementRuntimeModelRows(t, records.Run))
			mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, paper))
			mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
			mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WillReturnRows(managementRuntimeModelRows(t, owner))
			qs, legacy, ds, ms := make([]any, 140), make([]any, 140), make([]any, 13), make([]any, 4)
			for i, q := range f.Questions {
				qs[i] = q
				legacy[i] = model.PaperQu{ID: q.PaperQuestionID, PaperID: q.PaperID, QuID: q.SourceQuestionID, Sort: q.DisplayOrder, QuType: 1, Answered: 1, ActualScore: *q.RawAnswer}
			}
			for i := range ds {
				ds[i] = records.Dimensions[i]
			}
			for i := range ms {
				ms[i] = records.Modules[i]
			}
			mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WillReturnRows(managementRuntimeModelRows(t, qs...))
			mock.ExpectQuery("SELECT .*el_paper_qu").WillReturnRows(managementRuntimeModelRows(t, legacy...))
			mock.ExpectQuery("SELECT .*el_mng_result_dimension").WillReturnRows(managementRuntimeModelRows(t, ds...))
			mock.ExpectQuery("SELECT .*el_mng_result_module").WillReturnRows(managementRuntimeModelRows(t, ms...))
			mock.ExpectQuery("SELECT .*el_mng_runtime_receipt").WillReturnRows(managementRuntimeModelRows(t, records.Receipt))
			managementRuntimeExpectBuckets(t, mock, f.Questions)
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
			if corrupt {
				mock.ExpectRollback()
			} else {
				mock.ExpectExec("INSERT INTO `el_mng_report_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			got, data, err := s.ReadTestReportPDF(context.Background(), id, 1, root, content)
			if corrupt && (err == nil || data != nil) {
				t.Fatal("corrupt PDF exposed")
			}
			if !corrupt && (err != nil || got.ID != id || !bytes.Equal(data, pdf)) {
				t.Fatal("download rejected", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
