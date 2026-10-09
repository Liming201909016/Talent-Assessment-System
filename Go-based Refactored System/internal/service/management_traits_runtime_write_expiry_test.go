package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Regression MT-WRITE-CREDENTIAL-EXPIRY: execute the real transaction while
// the mock driver blocks a locking SELECT. Advance the server clock only after
// that SELECT is entered; expired credentials must rollback without any DML.
func TestBugManagementTraitsWriteCredentialExpiresDuringLock(t *testing.T) {
	for _, operation := range []string{"fill", "submit", "submit-retry"} {
		for _, kind := range []string{"candidate", "tester"} {
			for _, lock := range []string{"paper", "source"} {
				for _, offset := range []time.Duration{0, time.Nanosecond, time.Second} {
					t.Run(fmt.Sprintf("%s/%s/%s/%s", operation, kind, lock, offset), func(t *testing.T) {
						f, _, paper, profile, owner := managementRuntimeLoadFixture(t, false)
						owner.Kind, f.Paper.ParticipantType = kind, kind
						if kind == "tester" {
							owner.Status = "0"
						}
						if operation != "submit-retry" {
							paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
							for i := range f.Questions {
								f.Questions[i].SubmittedAt = nil
							}
						}
						before, err := json.Marshal(f)
						if err != nil {
							t.Fatal(err)
						}
						expires := paper.CreateTime.Add(5 * time.Minute)
						now := expires.Add(-time.Second)
						claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: expires.Unix()}
						token, err := CreateManagementTraitsRuntimeToken("test-only", claims, *paper.CreateTime)
						if err != nil {
							t.Fatal(err)
						}
						claims, err = ParseManagementTraitsRuntimeToken("test-only", token, claims.Purpose, now)
						if err != nil {
							t.Fatal("credential must pass admission before lock", err)
						}
						entered, release := make(chan struct{}), make(chan struct{})
						var clockCalls, dml int32
						blocked := false
						matcher := sqlmock.QueryMatcherFunc(func(expected, actual string) error {
							if err := sqlmock.QueryMatcherRegexp.Match(expected, actual); err != nil {
								return err
							}
							isLock := lock == "paper" && strings.Contains(actual, "FOR UPDATE") || lock == "source" && strings.Contains(actual, "LOCK IN SHARE MODE")
							if isLock && !blocked {
								blocked = true
								close(entered)
								<-release
							}
							return nil
						})
						sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { sqlDB.Close() })
						db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
						if err != nil {
							t.Fatal(err)
						}
						s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
						observe := func(*gorm.DB) { atomic.AddInt32(&dml, 1) }
						if err := s.db.Callback().Create().Before("gorm:create").Register("expiry_dml", observe); err != nil {
							t.Fatal(err)
						}
						if err := s.db.Callback().Update().Before("gorm:update").Register("expiry_dml", observe); err != nil {
							t.Fatal(err)
						}
						if err := s.db.Callback().Delete().Before("gorm:delete").Register("expiry_dml", observe); err != nil {
							t.Fatal(err)
						}
						managementRuntimeExpectPaper(t, mock, f, paper, profile, owner)
						mock.ExpectRollback()
						done := make(chan error, 1)
						go func() {
							clock := func() time.Time { atomic.AddInt32(&clockCalls, 1); return now }
							var err error
							if operation == "fill" {
								_, err = s.fillAnswerTransaction(context.Background(), claims, f.Questions[0].PaperQuestionID, *f.Questions[0].SelectedOptionID, clock)
							} else {
								_, err = s.submitTransaction(context.Background(), claims, "manual", clock)
							}
							done <- err
						}()
						select {
						case <-entered:
						case err := <-done:
							close(release)
							t.Fatalf("transaction never waited on lock: %v", err)
						case <-time.After(5 * time.Second):
							close(release)
							t.Fatal("lock was not entered")
						}
						if calls := atomic.LoadInt32(&clockCalls); calls != 0 {
							t.Errorf("clock sampled before lock release: %d", calls)
						}
						now = expires.Add(offset)
						close(release)
						select {
						case err := <-done:
							if !errors.Is(err, ErrManagementTraitsRuntimeToken) {
								t.Errorf("expired admitted credential must retain authentication error, got %v", err)
							}
						case <-time.After(5 * time.Second):
							t.Fatal("transaction did not finish")
						}
						if atomic.LoadInt32(&clockCalls) != 1 || atomic.LoadInt32(&dml) != 0 {
							t.Errorf("clock=%d DML=%d", clockCalls, dml)
						}
						after, err := json.Marshal(f)
						if err != nil || !bytes.Equal(before, after) {
							t.Error("frozen facts or deadline changed")
						}
						if err := mock.ExpectationsWereMet(); err != nil {
							t.Fatal(err)
						}
					})
				}
			}
		}
	}
}

func TestManagementTraitsWriteCredentialTimeBoundaries(t *testing.T) {
	expires := time.Unix(1790928300, 0)
	for _, item := range []struct {
		name string
		exp  int64
		now  time.Time
		want error
	}{
		{"missing", 0, expires.Add(-time.Second), ErrManagementTraitsRuntimeToken},
		{"negative", -1, expires.Add(-time.Second), ErrManagementTraitsRuntimeToken},
		{"before", expires.Unix(), expires.Add(-time.Nanosecond), nil},
		{"equal", expires.Unix(), expires, ErrManagementTraitsRuntimeToken},
		{"after", expires.Unix(), expires.Add(time.Nanosecond), ErrManagementTraitsRuntimeToken},
	} {
		t.Run(item.name, func(t *testing.T) {
			if err := managementTraitsRuntimeCredentialTime(ManagementTraitsRuntimeClaims{ExpiresAt: item.exp}, item.now); !errors.Is(err, item.want) {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsWriteTrustedInternalSubmitWithoutCredential(t *testing.T) {
	for _, exp := range []int64{0, 1} {
		t.Run(fmt.Sprint(exp), func(t *testing.T) {
			f, records, paper, profile, owner := managementRuntimeLoadFixture(t, true)
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			managementRuntimeExpectPaper(t, mock, f, paper, profile, owner)
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
			managementRuntimeExpectBuckets(t, mock, f.Questions)
			mock.ExpectCommit()
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			// No token is signed: internal authority is the caller boundary, not exp.
			claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: exp}
			got, err := s.Submit(context.Background(), claims, "manual")
			if err != nil || !got.Reused || got.RunID != records.Run.ID || !got.SubmittedAt.Equal(*records.Run.SubmittedAt) {
				t.Fatal("internal retry requires participant credential or changes facts", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
