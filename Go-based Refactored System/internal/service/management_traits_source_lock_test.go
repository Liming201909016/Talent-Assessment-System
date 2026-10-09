package service

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

// Regression MT-SOURCE-LOCK: ordered driver expectations execute the real
// transaction and require rollback, not just a source-code SQL substring.
func TestBugManagementTraitsSourceLockRevokedWriterRollback(t *testing.T) {
	for _, operation := range []string{"fill", "submit"} {
		t.Run(operation, func(t *testing.T) {
			f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
			paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
			for i := range f.Questions {
				f.Questions[i].SubmittedAt = nil
			}
			db, mock := managementRuntimeDB(t)
			mock.ExpectBegin()
			managementRuntimeExpectPaperBody(t, mock, f, paper, profile, owner)
			revoked := f.Bundle
			revoked.Status = "review-revoked"
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, revoked))
			mock.ExpectRollback()
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID}
			clockCalled := false
			clock := func() time.Time { clockCalled = true; return paper.CreateTime.Add(time.Minute) }
			var err error
			if operation == "fill" {
				_, err = s.fillAnswerTransaction(context.Background(), claims, f.Questions[0].PaperQuestionID, *f.Questions[0].SelectedOptionID, clock)
			} else {
				_, err = s.submitTransaction(context.Background(), claims, "manual", clock)
			}
			if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) || clockCalled {
				t.Fatalf("revocation not checked before writes/clock: %v, clock=%v", err, clockCalled)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBugManagementTraitsSourceLockRevokedNewCreationRollback(t *testing.T) {
	f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	owner.PaperID, owner.EndTime = "", nil
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: paper.ExamID, State: 0, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25}))
	mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, owner))
	mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
	revoked := f.Bundle
	revoked.Status = "review-revoked"
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, revoked))
	mock.ExpectRollback()
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	c := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, ExpiresAt: time.Now().Add(time.Minute).Unix()}
	if _, err := s.CreatePaper(context.Background(), c); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("revoked new paper accepted", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBugManagementTraitsSourceLockReportFinishRevokeCleanup(t *testing.T) {
	f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	wb, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "260929管理特质测评-优化", "260928测评内容+数据图.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	content, err := LoadManagementTraitsTestContent(wb)
	if err != nil {
		t.Fatal(err)
	}
	db, mock := managementRuntimeDB(t)
	managementRuntimeExpectSchema(t, mock)
	managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
	// Initial and final transactions both acquire a current source lease
	// after loading immutable facts and before persisting any report metadata.
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectCommit()
	mock.ExpectExec("MNG_RENDER_SOURCE_REVOKED_OUTSIDE_TRANSACTION").WillReturnResult(sqlmock.NewResult(0, 0))
	// loadRuntimePaper is deliberately the same immutable fixture; only the
	// current locking read observes revocation instead of its MVCC snapshot.
	managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
	revoked := f.Bundle
	revoked.Status = "review-revoked"
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, revoked))
	mock.ExpectRollback()
	root := t.TempDir()
	old := []byte("historical PDF sentinel")
	oldPath := filepath.Join(root, "historical.pdf")
	if err := os.WriteFile(oldPath, old, 0600); err != nil {
		t.Fatal(err)
	}
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	rendered := false
	_, err = s.GenerateTestReport(context.Background(), records.Run.ID, 1, root, content, func(context.Context, ManagementTraitsTestReportData) ([]byte, error) {
		rendered = true
		if e := db.Exec("MNG_RENDER_SOURCE_REVOKED_OUTSIDE_TRANSACTION").Error; e != nil {
			t.Fatal(e)
		}
		return append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("test-only"), 200)...), nil
	})
	if !rendered || !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("report finish accepted revoked source", err)
	}
	got, err := os.ReadFile(oldPath)
	if err != nil || !bytes.Equal(got, old) {
		t.Fatal("old report changed")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatal("new revoked PDF orphan", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsSourceLockExistingContractAndDrift(t *testing.T) {
	for _, status := range []string{"candidate-current-source", "retired", "draft", "reviewed", "review-revoked", "unknown", ""} {
		for _, operation := range []string{"fill", "submit"} {
			t.Run(operation+"/"+status, func(t *testing.T) {
				f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
				if operation == "fill" {
					paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
					for i := range f.Questions {
						f.Questions[i].SubmittedAt = nil
					}
				}
				before := managementTraitsContractClone(t, f.Questions)
				db, mock := managementRuntimeDB(t)
				mock.ExpectBegin()
				managementRuntimeExpectPaperBody(t, mock, f, paper, profile, owner)
				current := f.Bundle
				current.Status = status
				mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, current))
				allowed := status == "candidate-current-source" || status == "retired"
				if allowed {
					if operation == "submit" {
						mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(managementRuntimeModelRows(t, records.Run))
						ds, ms := make([]any, 13), make([]any, 4)
						for i := range ds {
							ds[i] = records.Dimensions[i]
						}
						for i := range ms {
							ms[i] = records.Modules[i]
						}
						mock.ExpectQuery("SELECT .*el_mng_result_dimension").WillReturnRows(managementRuntimeModelRows(t, ds...))
						mock.ExpectQuery("SELECT .*el_mng_result_module").WillReturnRows(managementRuntimeModelRows(t, ms...))
						mock.ExpectQuery("SELECT .*el_mng_runtime_receipt").WillReturnRows(managementRuntimeModelRows(t, records.Receipt))
					}
					managementRuntimeExpectBuckets(t, mock, f.Questions)
					mock.ExpectCommit()
				} else {
					mock.ExpectRollback()
				}
				s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
				c := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: paper.CreateTime.Add(2 * time.Hour).Unix()}
				var err error
				if operation == "fill" {
					var answered int
					answered, err = s.fillAnswerTransaction(context.Background(), c, f.Questions[0].PaperQuestionID, *f.Questions[0].SelectedOptionID, func() time.Time { return paper.CreateTime.Add(time.Minute) })
					if allowed && answered != 140 {
						t.Fatal("idempotent answer changed count", answered)
					}
				} else {
					var got ManagementTraitsRuntimeSubmission
					got, err = s.submitTransaction(context.Background(), c, "manual", func() time.Time { return paper.CreateTime.Add(time.Hour) })
					if allowed && (!got.Reused || got.RunID != records.Run.ID || !got.SubmittedAt.Equal(*records.Run.SubmittedAt)) {
						t.Fatal("submission retry changed frozen facts")
					}
				}
				if allowed && err != nil || !allowed && !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
					t.Fatal("status contract changed", err)
				}
				if !reflect.DeepEqual(f.Questions, before) {
					t.Fatal("frozen input mutated")
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsSourceLockCurrentReadErrorsRollback(t *testing.T) {
	for _, scenario := range []string{"query-error", "row-error", "scan-error", "missing", "wrong-id", "product", "question", "scoring", "norm", "questionnaire", "manifest", "sha", "created-at"} {
		t.Run(scenario, func(t *testing.T) {
			f, _, _, _, _ := managementRuntimeLoadFixture(t, false)
			db, mock := managementRuntimeDB(t)
			mock.ExpectBegin()
			query := mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID)
			current := f.Bundle
			switch scenario {
			case "query-error":
				query.WillReturnError(errors.New("private lock failure"))
			case "row-error":
				query.WillReturnRows(managementRuntimeModelRows(t, current).RowError(0, errors.New("private row failure")))
			case "scan-error":
				query.WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(current.ID, "not-a-time"))
			case "missing":
				query.WillReturnRows(sqlmock.NewRows([]string{"id"}))
			default:
				switch scenario {
				case "wrong-id":
					current.ID += "-changed"
				case "product":
					current.ProductVersion += "-changed"
				case "question":
					current.QuestionVersion += "-changed"
				case "scoring":
					current.ScoringVersion += "-changed"
				case "norm":
					current.NormVersion += "-changed"
				case "questionnaire":
					current.Questionnaire = "leader"
				case "manifest":
					current.ScoringManifest += " "
				case "sha":
					current.ScoringManifestSHA = "changed"
				case "created-at":
					current.CreatedAt = current.CreatedAt.Add(time.Second)
				}
				query.WillReturnRows(managementRuntimeModelRows(t, current))
			}
			mock.ExpectRollback()
			err := db.Transaction(func(tx *gorm.DB) error {
				_, err := revalidateManagementTraitsWriteBundle(context.Background(), tx, f.Bundle)
				return err
			})
			if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Fatal("current-read drift/error accepted", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsSourceLockNewCreationStatusContract(t *testing.T) {
	for _, status := range []string{"candidate-current-source", "retired", "draft", "reviewed", "review-revoked"} {
		t.Run(status, func(t *testing.T) {
			f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
			owner.PaperID, owner.EndTime = "", nil
			current := f.Bundle
			current.Status = status
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: paper.ExamID, State: 0, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25}))
			mock.ExpectQuery("SELECT .*el_candidate.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, owner))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(managementRuntimeModelRows(t, profile))
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, current))
			if status == "candidate-current-source" {
				// Reaching the real insert proves current source is admitted;
				// inject failure to keep the fixture independent of generated UUIDs.
				mock.ExpectExec("INSERT INTO `el_paper`").WillReturnError(errors.New("private paper failure"))
			}
			mock.ExpectRollback()
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			c := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, ExpiresAt: time.Now().Add(time.Minute).Unix()}
			if _, err := s.CreatePaper(context.Background(), c); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Fatal("status/insert failure accepted", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsSourceLockReportAdmissionRejectsBeforeRendering(t *testing.T) {
	for _, scenario := range []string{"revoked", "lock-error", "manifest-drift"} {
		t.Run(scenario, func(t *testing.T) {
			f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
			wb, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "260929管理特质测评-优化", "260928测评内容+数据图.xlsx"))
			if err != nil {
				t.Fatal(err)
			}
			content, err := LoadManagementTraitsTestContent(wb)
			if err != nil {
				t.Fatal(err)
			}
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			managementRuntimeExpectLoad(t, mock, f, records, paper, profile, owner)
			query := mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID)
			current := f.Bundle
			if scenario == "lock-error" {
				query.WillReturnError(errors.New("private lock failure"))
			} else {
				if scenario == "revoked" {
					current.Status = "review-revoked"
				} else {
					current.ScoringManifest += " "
				}
				query.WillReturnRows(managementRuntimeModelRows(t, current))
			}
			mock.ExpectRollback()
			root := t.TempDir()
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			rendered := false
			_, err = s.GenerateTestReport(context.Background(), records.Run.ID, 1, root, content, func(context.Context, ManagementTraitsTestReportData) ([]byte, error) {
				rendered = true
				return nil, nil
			})
			if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) || rendered {
				t.Fatal("untrusted report started rendering", err)
			}
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 0 {
				t.Fatal("rejected admission wrote a file")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
