package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm/logger"
)

const managementRaceTestPaper = "f5e21ed0-4125-40a6-900a-a634881983ee"

func managementRaceTestTarget(paper string) string {
	h := sha256.Sum256([]byte("mng-race-paper-v1\x00" + paper))
	return hex.EncodeToString(h[:])
}

func managementRaceTestEnable(t *testing.T) {
	t.Helper()
	t.Setenv("MNG_RACE_OBSERVE_ENV", "local")
	t.Setenv("MNG_RACE_OBSERVE_PAPER_SHA256", managementRaceTestTarget(managementRaceTestPaper))
}

func managementRaceTestEntries(t *testing.T, text string) []map[string]any {
	t.Helper()
	rows := make([]map[string]any, 0)
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		if line == "" {
			continue
		}
		var row map[string]any
		if json.Unmarshal([]byte(line), &row) != nil || row["msg"] != "management_traits_settlement_observation" || len(row) != 12 {
			t.Fatal("non-allowlisted observation")
		}
		rows = append(rows, row)
	}
	return rows
}

// UF053: observation capability is not a main HTTP/Worker race verdict.
func TestManagementTraitsRaceObservationFailClosed(t *testing.T) {
	for _, x := range []struct{ env, target, paper string }{
		{"", "", managementRaceTestPaper},
		{"production", managementRaceTestTarget(managementRaceTestPaper), managementRaceTestPaper},
		{"staging ", managementRaceTestTarget(managementRaceTestPaper), managementRaceTestPaper},
		{"local", "*", managementRaceTestPaper},
		{"local", strings.ToUpper(managementRaceTestTarget(managementRaceTestPaper)), managementRaceTestPaper},
		{"local", managementRaceTestTarget(managementRaceTestPaper), "4c3a2fe9-0063-4166-bb72-e2587892c637"},
		{"local", managementRaceTestTarget("123"), "123"},
		{"local", managementRaceTestTarget(strings.ToUpper(managementRaceTestPaper)), strings.ToUpper(managementRaceTestPaper)},
	} {
		t.Run(x.env+"/"+x.paper, func(t *testing.T) {
			t.Setenv("MNG_RACE_OBSERVE_ENV", x.env)
			t.Setenv("MNG_RACE_OBSERVE_PAPER_SHA256", x.target)
			out := managementDiagnosticCapture(t)
			s := NewManagementTraitsRuntimeService(nil, "test-only", 1<<20)
			ctx, obs := s.startRaceObservation(context.Background(), x.paper, managementRaceParticipant)
			if obs != nil || ctx != context.Background() || out.Len() != 0 {
				t.Fatal("invalid scope enabled observation")
			}
		})
	}
}

func TestManagementTraitsRaceObservationContextIsolation(t *testing.T) {
	managementRaceTestEnable(t)
	s := NewManagementTraitsRuntimeService(nil, "test-only", 1<<20)
	type key struct{}
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), key{}, "PRIVATE_CONTEXT_SENTINEL"))
	ctx, first := s.startRaceObservation(parent, managementRaceTestPaper, managementRaceParticipant)
	if first == nil || ctx.Value(key{}) != "PRIVATE_CONTEXT_SENTINEL" {
		t.Fatal("context not preserved")
	}
	_, second := s.startRaceObservation(ctx, managementRaceTestPaper, managementRaceWorker)
	if second == nil || first == second || first.attempt == second.attempt || first.resource != second.resource || first.session != second.session {
		t.Fatal("attempts not isolated")
	}
	cancel()
	if ctx.Err() != context.Canceled {
		t.Fatal("cancel lost")
	}
	seen := make(chan string, 32)
	var wg sync.WaitGroup
	for i := 0; i < cap(seen); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, obs := s.startRaceObservation(parent, managementRaceTestPaper, managementRaceWorker)
			if obs != nil {
				seen <- obs.attempt
			}
		}()
	}
	wg.Wait()
	close(seen)
	unique := map[string]bool{}
	for id := range seen {
		if unique[id] {
			t.Fatal("duplicate attempt")
		}
		unique[id] = true
	}
	if len(unique) != 32 {
		t.Fatal("observation dropped")
	}
}

func TestManagementTraitsRaceObservationRealSubmitStages(t *testing.T) {
	for _, scenario := range []string{"created", "reused", "missing", "write-failure", "commit-failure", "begin-failure", "paper-failure", "bundle-failure", "schema-failure", "internal", "worker"} {
		t.Run(scenario, func(t *testing.T) {
			managementRaceTestEnable(t)
			out := &managementRacePostTransactionWriter{}
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(out, nil)))
			t.Cleanup(func() { slog.SetDefault(previous) })
			incomplete := scenario == "missing" || scenario == "worker"
			f, _, paper, profile, owner := managementRuntimeLoadFixture(t, incomplete)
			f.Paper.PaperID, paper.ID, owner.PaperID = managementRaceTestPaper, managementRaceTestPaper, managementRaceTestPaper
			paper.State, paper.UserTime, owner.EndTime = 1, 0, nil
			for i := range f.Questions {
				f.Questions[i].PaperID, f.Questions[i].SubmittedAt = managementRaceTestPaper, nil
			}
			if scenario == "missing" {
				start := time.Now().Add(-time.Minute).Truncate(time.Second)
				limit := start.Add(25 * time.Minute)
				f.Paper.StartedAt, f.Paper.LimitTime, paper.CreateTime, paper.LimitTime = start, &limit, &start, &limit
				f.Paper.CreatedAt = start
			}
			db, mock := managementRuntimeDB(t)
			out.complete = func() bool { return mock.ExpectationsWereMet() == nil }
			db.Config.Logger = logger.Discard
			if scenario == "schema-failure" {
				mock.ExpectQuery("SELECT table_name AS table_name").WillReturnError(errors.New(managementDiagnosticSecret))
			} else {
				managementRuntimeExpectSchema(t, mock)
				if scenario == "worker" {
					mock.ExpectQuery(managementExpiryFirstQuery).WillReturnRows(managementExpiryRows([]string{paper.ID}, paper.ExamID, owner.Kind, owner.ID, *paper.LimitTime))
				}
				if scenario == "begin-failure" {
					mock.ExpectBegin().WillReturnError(errors.New(managementDiagnosticSecret))
				} else if scenario == "paper-failure" {
					mock.ExpectBegin()
					mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnError(errors.New(managementDiagnosticSecret))
					mock.ExpectRollback()
				} else {
					var reused managementTraitsRuntimeRecords
					if scenario == "reused" {
						at := *f.Paper.LimitTime
						var fixtureErr error
						reused, fixtureErr = buildManagementTraitsRuntimeRecords(f.Bundle, f.Paper, f.Questions, "opaque-run", at, 1<<20)
						if fixtureErr != nil {
							t.Fatal("reused fixture invalid")
						}
						paper.State, paper.UserTime, owner.EndTime = 2, *reused.Run.UserTimeSeconds, &at
						for i := range f.Questions {
							f.Questions[i].SubmittedAt = &at
						}
					}
					mock.ExpectBegin()
					managementRuntimeExpectPaperBody(t, mock, f, paper, profile, owner)
					q := mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE")
					if scenario == "bundle-failure" {
						q.WillReturnError(errors.New(managementDiagnosticSecret))
						mock.ExpectRollback()
					} else {
						q.WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
						if scenario == "reused" {
							mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(managementRuntimeModelRows(t, reused.Run))
							dims, modules := make([]any, 13), make([]any, 4)
							for i := range reused.Dimensions {
								dims[i] = reused.Dimensions[i]
							}
							for i := range reused.Modules {
								modules[i] = reused.Modules[i]
							}
							mock.ExpectQuery("SELECT .*el_mng_result_dimension").WillReturnRows(managementRuntimeModelRows(t, dims...))
							mock.ExpectQuery("SELECT .*el_mng_result_module").WillReturnRows(managementRuntimeModelRows(t, modules...))
							mock.ExpectQuery("SELECT .*el_mng_runtime_receipt").WillReturnRows(managementRuntimeModelRows(t, reused.Receipt))
							managementRuntimeExpectBuckets(t, mock, f.Questions)
							mock.ExpectCommit()
						} else {
							mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"id"}))
							if scenario == "missing" {
								mock.ExpectRollback()
							} else {
								managementRuntimeExpectBuckets(t, mock, f.Questions)
								r := mock.ExpectExec("INSERT INTO `el_mng_result_run`")
								if scenario == "write-failure" {
									r.WillReturnError(errors.New(managementDiagnosticSecret))
									mock.ExpectRollback()
								} else {
									r.WillReturnResult(sqlmock.NewResult(1, 1))
									mock.ExpectExec("INSERT INTO `el_mng_result_dimension`").WillReturnResult(sqlmock.NewResult(1, 13))
									mock.ExpectExec("INSERT INTO `el_mng_result_module`").WillReturnResult(sqlmock.NewResult(1, 4))
									mock.ExpectExec("INSERT INTO `el_mng_runtime_receipt`").WillReturnResult(sqlmock.NewResult(1, 1))
									mock.ExpectExec("UPDATE `el_mng_paper_question_snapshot`").WillReturnResult(sqlmock.NewResult(0, 140))
									mock.ExpectExec("UPDATE `el_paper`").WillReturnResult(sqlmock.NewResult(0, 1))
									mock.ExpectExec("UPDATE `el_candidate`").WillReturnResult(sqlmock.NewResult(0, 1))
									if scenario == "commit-failure" {
										mock.ExpectCommit().WillReturnError(errors.New(managementDiagnosticSecret))
									} else {
										mock.ExpectCommit()
									}
								}
							}
						}
					}
				}
			}
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: paper.ID, ExpiresAt: time.Now().Add(time.Hour).Unix()}
			var err error
			var result ManagementTraitsRuntimeSubmission
			wantCaller := "http_participant"
			if scenario == "worker" {
				wantCaller = "expiry_worker"
				err = s.ScanExpiry(context.Background(), 100)
			} else if scenario == "internal" {
				wantCaller = "trusted_internal"
				result, err = s.Submit(context.Background(), claims, "manual")
			} else {
				// A client-shaped context value cannot choose the fixed caller.
				ctx := context.WithValue(context.Background(), managementRaceWorkerKey{}, true)
				result, err = s.SubmitParticipant(ctx, claims, "manual")
			}
			wantSuccess := scenario == "created" || scenario == "reused" || scenario == "internal" || scenario == "worker"
			if (err == nil) != wantSuccess || (wantSuccess && scenario != "worker" && result.Answered != 140) {
				t.Fatal("business result changed")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal("SQL contract changed")
			}
			if out.early {
				t.Fatal("logging occurred before transaction returned")
			}
			rows := managementRaceTestEntries(t, out.String())
			if len(rows) == 0 || len(rows) > 10 || rows[0]["stage"] != "entry" {
				t.Fatal("missing/bounded entry evidence")
			}
			stages := map[string]bool{}
			var last float64 = -1
			for _, row := range rows {
				if row["caller"] != wantCaller || row["attempt"] != rows[0]["attempt"] || row["resource"] != rows[0]["resource"] || row["process_ns"].(float64) < last {
					t.Fatal("attribution or order changed")
				}
				last = row["process_ns"].(float64)
				stages[row["stage"].(string)] = true
			}
			wantOutcome := "committed_created"
			if scenario == "reused" {
				wantOutcome = "committed_reused"
				if !result.Reused {
					t.Fatal("reuse result changed")
				}
			}
			if wantSuccess && (!stages["paper_lock_acquired"] || !stages["bundle_lock_acquired"] || !stages["body_returned"] || rows[len(rows)-1]["stage"] != "transaction_returned" || rows[len(rows)-1]["outcome"] != wantOutcome) {
				t.Fatal("successful settlement evidence missing")
			}
			if !wantSuccess && rows[len(rows)-1]["outcome"] != "failed" {
				t.Fatal("failure claimed commit")
			}
			if scenario == "paper-failure" && stages["paper_lock_acquired"] || scenario == "bundle-failure" && stages["bundle_lock_acquired"] || scenario == "begin-failure" && stages["transaction_entered"] {
				t.Fatal("failed lock/BEGIN claimed acquired")
			}
			for _, secret := range []string{managementDiagnosticSecret, managementRaceTestPaper, owner.ID, owner.Name, owner.Telephone, paper.ExamID, "test-only"} {
				if strings.Contains(out.String(), secret) {
					t.Fatal("private data emitted")
				}
			}
		})
	}
}

type managementRacePostTransactionWriter struct {
	bytes.Buffer
	complete func() bool
	early    bool
}

func (w *managementRacePostTransactionWriter) Write(p []byte) (int, error) {
	if !w.complete() {
		w.early = true
	}
	return w.Buffer.Write(p)
}

func TestManagementTraitsRaceObservationBoundedAndImmutable(t *testing.T) {
	managementRaceTestEnable(t)
	out := managementDiagnosticCapture(t)
	s := NewManagementTraitsRuntimeService(nil, "test-only", 1<<20)
	t.Setenv("MNG_RACE_OBSERVE_ENV", "production")
	t.Setenv("MNG_RACE_OBSERVE_PAPER_SHA256", "")
	ctx, obs := s.startRaceObservation(context.Background(), managementRaceTestPaper, managementRaceParticipant)
	if obs == nil {
		t.Fatal("config not construction-time immutable")
	}
	obs.mark(managementDiagnosticSecret)
	if obs.count != 1 {
		t.Fatal("unknown stage accepted")
	}
	for i := 0; i < 100; i++ {
		obs.mark("settlement_entered")
	}
	if obs.count != 10 || out.Len() != 0 {
		t.Fatal("unbounded or eager output")
	}
	_, absent := s.startRaceObservation(ctx, "4c3a2fe9-0063-4166-bb72-e2587892c637", managementRaceWorker)
	if absent != nil {
		t.Fatal("scope inherited")
	}
	obs.emit(ManagementTraitsRuntimeSubmission{}, errors.New(managementDiagnosticSecret))
	if len(managementRaceTestEntries(t, out.String())) != 10 || strings.Contains(out.String(), managementDiagnosticSecret) {
		t.Fatal("unsafe output")
	}
	if cfg := newManagementRaceObservationConfig("local", managementRaceTestTarget(managementRaceTestPaper), strings.NewReader("")); cfg.target != "" {
		t.Fatal("entropy failure enabled observation")
	}
	if cfg := newManagementRaceObservationConfig("staging", managementRaceTestTarget(managementRaceTestPaper), strings.NewReader(strings.Repeat("a", 32))); cfg.target == "" {
		t.Fatal("staging scope rejected")
	}
	_, invalidCaller := s.startRaceObservation(context.Background(), managementRaceTestPaper, managementRaceCaller(255))
	if invalidCaller != nil {
		t.Fatal("unknown caller accepted")
	}
	var disabled *managementRaceObservation
	disabled.mark("entry")
	disabled.emit(ManagementTraitsRuntimeSubmission{}, nil)
}

func TestManagementTraitsRaceObservationIncompleteAttemptNeverClaimsCommit(t *testing.T) {
	managementRaceTestEnable(t)
	out := managementDiagnosticCapture(t)
	s := NewManagementTraitsRuntimeService(nil, "test-only", 1<<20)
	_, obs := s.startRaceObservation(context.Background(), managementRaceTestPaper, managementRaceParticipant)
	obs.mark("transaction_entered")
	// An unwinding panic has no normal Transaction return, even with nil err.
	obs.emit(ManagementTraitsRuntimeSubmission{}, nil)
	for _, row := range managementRaceTestEntries(t, out.String()) {
		if row["outcome"] != "failed" {
			t.Fatal("incomplete attempt claimed commit")
		}
	}
}

func TestManagementTraitsRaceObservationConcurrentRealEntries(t *testing.T) {
	managementRaceTestEnable(t)
	out := managementDiagnosticCapture(t)
	db, mock := managementRuntimeDB(t)
	db.Config.Logger = logger.Discard
	managementRuntimeExpectSchema(t, mock)
	s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
	if s.CheckRuntimeSchema(context.Background()) != nil {
		t.Fatal("schema fixture invalid")
	}
	mock.MatchExpectationsInOrder(false)
	f, _, paper, _, owner := managementRuntimeLoadFixture(t, true)
	mock.ExpectQuery(managementExpiryFirstQuery).WillReturnRows(managementExpiryRows([]string{managementRaceTestPaper}, paper.ExamID, owner.Kind, owner.ID, *f.Paper.LimitTime))
	for i := 0; i < 2; i++ {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WillReturnError(errors.New(managementDiagnosticSecret))
		mock.ExpectRollback()
	}
	claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: paper.ExamID, PaperID: managementRaceTestPaper, ExpiresAt: time.Now().Add(time.Hour).Unix()}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := s.SubmitParticipant(context.Background(), claims, "manual")
		if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Error("participant result changed")
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		if !errors.Is(s.ScanExpiry(context.Background(), 100), ErrManagementTraitsRuntimeInvalid) {
			t.Error("worker result changed")
		}
	}()
	close(start)
	wg.Wait()
	if mock.ExpectationsWereMet() != nil {
		t.Fatal("concurrent SQL expectations unmet")
	}
	rows := managementRaceTestEntries(t, out.String())
	callers, attempts := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		callers[row["caller"].(string)] = true
		attempts[row["attempt"].(string)] = true
		if row["session"] != rows[0]["session"] || row["resource"] != rows[0]["resource"] || row["outcome"] != "failed" {
			t.Fatal("concurrent attribution mixed")
		}
	}
	if len(rows) != 10 || len(attempts) != 2 || len(callers) != 2 || !callers["http_participant"] || !callers["expiry_worker"] {
		t.Fatal("concurrent caller evidence missing")
	}
}
