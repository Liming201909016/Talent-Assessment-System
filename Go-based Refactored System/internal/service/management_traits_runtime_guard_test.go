package service

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var managementLegacyCoreNames = []string{
	"el_mng_definition_bundle", "el_mng_exam_profile", "el_mng_paper_snapshot",
	"el_mng_paper_question_snapshot", "el_mng_result_run", "el_mng_result_dimension", "el_mng_result_module",
}

func managementLegacyMetadata(mock sqlmock.Sqlmock, names ...string) {
	rows := sqlmock.NewRows([]string{"table_name"})
	for _, name := range names {
		rows.AddRow(name)
	}
	mock.ExpectQuery(`SELECT table_name AS table_name FROM information_schema.tables WHERE table_schema = DATABASE\(\) AND LEFT\(table_name,7\) = 'el_mng_' ORDER BY table_name`).WillReturnRows(rows).RowsWillBeClosed()
}

func managementLegacyEmpty(mock sqlmock.Sqlmock, table string) {
	mock.ExpectQuery("SELECT .* FROM " + table + " WHERE ").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"})).RowsWillBeClosed()
}

func managementLegacyNoProtection(mock sqlmock.Sqlmock) {
	for _, table := range []string{"el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM " + table + " WHERE ").WillReturnRows(sqlmock.NewRows([]string{"protected"})).RowsWillBeClosed()
	}
}

func managementLegacyAllExists(mock sqlmock.Sqlmock, table string, found bool, queryError bool) {
	q := mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT 1 AS protected FROM "+table+" LIMIT 1") + "$").WithoutArgs()
	if queryError {
		q.WillReturnError(errors.New("private database failure"))
		return
	}
	rows := sqlmock.NewRows([]string{"protected"})
	if found {
		rows.AddRow(1)
	}
	q.WillReturnRows(rows).RowsWillBeClosed()
}

func TestManagementTraitsLegacyScopeAllLegacyCore(t *testing.T) {
	tables := []string{"el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot"}
	for _, protectedTable := range append(append([]string{}, tables...), "none") {
		t.Run(protectedTable, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, managementLegacyCoreNames...)
			for _, table := range tables {
				managementLegacyAllExists(mock, table, table == protectedTable, false)
				if table == protectedTable {
					break
				}
			}
			// Explicit targets must not trigger legacy closure queries in all-collection mode.
			req := ManagementTraitsLegacyScopeRequest{AllLegacy: true, ExamIDs: []string{"unrelated"}, PaperQuestionIDs: []string{"unrelated-question"}}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, req)
			if err != nil || got != (protectedTable != "none") {
				t.Fatalf("protected=%v error=%v", got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeAllLegacyMarkers(t *testing.T) {
	for _, column := range []string{"paper_id", "exam_id", "profile_exam_id", "candidate_id", "tester_id", "participant_id", "paper_question_id", "pdf_path", "run_id", "result_run_id"} {
		for _, found := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%v", column, found), func(t *testing.T) {
				db, mock := managementRuntimeDB(t)
				names := []string{"el_mng_capture_marker"}
				isRun := column == "run_id" || column == "result_run_id"
				if isRun {
					names = append(names, managementLegacyCoreNames...)
				}
				managementLegacyMetadata(mock, names...)
				mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WithArgs("el_mng_capture_marker").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_capture_marker", column)).RowsWillBeClosed()
				if isRun {
					for _, table := range []string{"el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot"} {
						managementLegacyAllExists(mock, table, false, false)
					}
				}
				managementLegacyAllExists(mock, "el_mng_capture_marker", found, false)
				got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{AllLegacy: true})
				if got != found || err != nil {
					t.Fatal(got, err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsLegacyScopeAllLegacyMetadata(t *testing.T) {
	for _, tc := range []struct {
		name    string
		names   []string
		column  string
		invalid bool
	}{
		{"absent", nil, "", false},
		{"partial", managementLegacyCoreNames[:1], "", true},
		{"missing-seventh", managementLegacyCoreNames[:6], "", true},
		{"unknown-extra", append(append([]string{}, managementLegacyCoreNames...), "el_mng_unknown"), "unsupported", true},
		{"run-marker-without-core", []string{"el_mng_unknown"}, "run_id", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, tc.names...)
			if tc.column != "" {
				mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WithArgs("el_mng_unknown").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_unknown", tc.column)).RowsWillBeClosed()
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{AllLegacy: true})
			if got || (tc.invalid && err != ErrManagementTraitsRuntimeInvalid) || (!tc.invalid && err != nil) {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeAllLegacyQueryErrors(t *testing.T) {
	for _, failingTable := range []string{"el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_capture_marker"} {
		t.Run(failingTable, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, append(append([]string{}, managementLegacyCoreNames...), "el_mng_capture_marker")...)
			mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WithArgs("el_mng_capture_marker").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_capture_marker", "paper_id")).RowsWillBeClosed()
			for _, table := range []string{"el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot", "el_mng_capture_marker"} {
				managementLegacyAllExists(mock, table, false, table == failingTable)
				if table == failingTable {
					break
				}
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{AllLegacy: true})
			if got || err != ErrManagementTraitsRuntimeInvalid {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeHistoricalSiblingPrecision(t *testing.T) {
	for _, tc := range []struct {
		name    string
		profile bool
		broad   bool
		want    bool
	}{
		{"untouched-paper-A", false, false, false},
		{"profile-protects-paper-A", true, false, true},
		{"explicit-exam-E-protects-any-history", false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, managementLegacyCoreNames...)
			req := ManagementTraitsLegacyScopeRequest{PaperIDs: []string{"A"}}
			paperWhere, ownerWhere := "(id = ?)", "(paper_id = ?)"
			args := []driver.Value{"A"}
			if tc.broad {
				req = ManagementTraitsLegacyScopeRequest{ExamIDs: []string{"E"}}
				paperWhere, ownerWhere, args = "(exam_id = ?)", "(exam_id = ?)", []driver.Value{"E"}
			}
			papers := sqlmock.NewRows([]string{"id", "exam_id"}).AddRow("A", "E")
			if tc.broad {
				papers.AddRow("B", "E")
			}
			mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id FROM el_paper WHERE "+paperWhere) + "$").WithArgs(args...).WillReturnRows(papers).RowsWillBeClosed()
			for _, table := range []string{"el_candidate", "el_tester"} {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id, paper_id FROM "+table+" WHERE "+ownerWhere) + "$").WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"})).RowsWillBeClosed()
			}
			if tc.broad {
				for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
					managementLegacyEmpty(mock, table)
				}
			}
			profileRows := sqlmock.NewRows([]string{"protected"})
			if tc.profile {
				profileRows.AddRow(1)
			}
			mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 AS protected FROM el_mng_exam_profile WHERE (exam_id = ?) LIMIT 1") + "$").WithArgs("E").WillReturnRows(profileRows).RowsWillBeClosed()
			if !tc.profile {
				where, runtimeArgs := "(paper_id = ?)", []driver.Value{"A"}
				if tc.broad {
					where, runtimeArgs = "(paper_id = ?) OR (paper_id = ?) OR (exam_id = ?)", []driver.Value{"A", "B", "E"}
				}
				// The sole historical snapshot belongs to B/E, never A.
				snapshotRows := sqlmock.NewRows([]string{"protected"})
				if tc.broad {
					snapshotRows.AddRow(1)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 AS protected FROM el_mng_paper_snapshot WHERE "+where+" LIMIT 1") + "$").WithArgs(runtimeArgs...).WillReturnRows(snapshotRows).RowsWillBeClosed()
				if !tc.broad {
					for _, table := range []string{"el_mng_result_run", "el_mng_paper_question_snapshot"} {
						mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 AS protected FROM "+table+" WHERE (paper_id = ?) LIMIT 1") + "$").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"protected"})).RowsWillBeClosed()
					}
				}
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, req)
			if err != nil || got != tc.want {
				t.Errorf("protected=%v error=%v; want protected=%v", got, err, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeRealOwnerRebinding(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		for _, history := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
			t.Run(kind+"/"+history, func(t *testing.T) {
				db, mock := managementRuntimeDB(t)
				managementLegacyMetadata(mock, managementLegacyCoreNames...)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id FROM el_paper WHERE (id = ?)") + "$").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}).AddRow("A", "E-old"))
				// The actual owner link points to B in another exam. No exam sibling lookup is allowed.
				for _, table := range []string{"el_candidate", "el_tester"} {
					rows := sqlmock.NewRows([]string{"id", "exam_id", "paper_id"})
					if table == "el_"+kind {
						rows.AddRow("owner", "E-owner", "B")
					}
					mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id, paper_id FROM "+table+" WHERE (paper_id = ?)") + "$").WithArgs("A").WillReturnRows(rows)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id FROM el_paper WHERE (id = ?)") + "$").WithArgs("B").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}).AddRow("B", "E-paper"))
				for _, table := range []string{"el_candidate", "el_tester"} {
					where, args := "(paper_id = ?)", []driver.Value{"B"}
					if table == "el_"+kind {
						where, args = "(id = ?) OR (paper_id = ?)", []driver.Value{"owner", "B"}
					}
					mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id, paper_id FROM "+table+" WHERE "+where) + "$").WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 AS protected FROM el_mng_exam_profile WHERE (exam_id = ?) OR (exam_id = ?) OR (exam_id = ?) LIMIT 1")+"$").WithArgs("E-old", "E-owner", "E-paper").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
				where := "(paper_id = ?) OR (paper_id = ?) OR (participant_type = ? AND participant_id = ?)"
				for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
					rows := sqlmock.NewRows([]string{"protected"})
					if table == history {
						// History may name a severed paper X: the direct participant still protects the owner.
						rows.AddRow(1)
					}
					mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 AS protected FROM "+table+" WHERE "+where+" LIMIT 1")+"$").WithArgs("A", "B", kind, "owner").WillReturnRows(rows).RowsWillBeClosed()
					if table == history {
						break
					}
				}
				got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{PaperIDs: []string{"A"}})
				if !got || err != nil {
					t.Fatal(got, err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsLegacyScopeInferredExamMarkerPrecision(t *testing.T) {
	for _, column := range []string{"exam_id", "profile_exam_id", "run_id", "result_run_id"} {
		t.Run(column, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			names := []string{"el_mng_capture_marker"}
			isRun := column == "run_id" || column == "result_run_id"
			if isRun {
				names = append(names, managementLegacyCoreNames...)
			}
			managementLegacyMetadata(mock, names...)
			mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WithArgs("el_mng_capture_marker").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_capture_marker", column))
			mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id FROM el_paper WHERE (id = ?)") + "$").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}).AddRow("A", "E"))
			for _, table := range []string{"el_candidate", "el_tester"} {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id, paper_id FROM "+table+" WHERE (paper_id = ?)") + "$").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
			}
			if isRun {
				mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE").WithArgs("E").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
				for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot"} {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 AS protected FROM "+table+" WHERE (paper_id = ?) LIMIT 1") + "$").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 AS protected FROM el_mng_capture_marker WHERE ("+column+" IN (SELECT id FROM el_mng_result_run WHERE (paper_id = ?))) LIMIT 1") + "$").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
			} else if column == "profile_exam_id" {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 AS protected FROM el_mng_capture_marker WHERE (profile_exam_id = ?) LIMIT 1") + "$").WithArgs("E").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{PaperIDs: []string{"A"}})
			if err != nil || got != (column == "profile_exam_id") {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeMetadata(t *testing.T) {
	for _, tc := range []struct {
		name       string
		tables     []string
		queryError bool
		invalid    bool
	}{
		{"absent", nil, false, false},
		{"partial", managementLegacyCoreNames[:1], false, true},
		{"missing-seventh", managementLegacyCoreNames[:6], false, true},
		{"metadata-error", nil, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			if tc.queryError {
				mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnError(errors.New("secret-schema"))
			} else {
				managementLegacyMetadata(mock, tc.tables...)
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{ExamIDs: []string{"exam"}})
			if got || (tc.invalid && err != ErrManagementTraitsRuntimeInvalid) || (!tc.invalid && err != nil) {
				t.Fatalf("got %v %v", got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
	if _, err := CheckManagementTraitsLegacyScope(context.Background(), nil, ManagementTraitsLegacyScopeRequest{}); err != ErrManagementTraitsRuntimeInvalid {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyScopeProfileAndNoCache(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	managementLegacyMetadata(mock)
	if got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{ExamIDs: []string{"exam"}}); got || err != nil {
		t.Fatal(got, err)
	}
	managementLegacyMetadata(mock, managementLegacyCoreNames...)
	for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
		managementLegacyEmpty(mock, table)
	}
	mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1)).RowsWillBeClosed()
	got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{ExamIDs: []string{"exam", "exam", ""}})
	if !got || err != nil {
		t.Fatal(got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyScopeStoredRebindAndQuestion(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	managementLegacyMetadata(mock, managementLegacyCoreNames...)
	mock.ExpectQuery("SELECT paper_id FROM el_paper_qu WHERE").WithArgs("arbitrary-review-id").WillReturnRows(sqlmock.NewRows([]string{"paper_id"}).AddRow("stored-paper")).RowsWillBeClosed()
	mock.ExpectQuery("SELECT id, exam_id FROM el_paper WHERE").WithArgs("stored-paper", "destination").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id"}).AddRow("stored-paper", "stored-exam")).RowsWillBeClosed()
	mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("owner", "stored-exam", "stored-paper")).RowsWillBeClosed()
	managementLegacyEmpty(mock, "el_tester")
	mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE").WithArgs("destination", "stored-exam").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
	got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{ExamIDs: []string{"destination"}, CandidateIDs: []string{"owner"}, PaperQuestionIDs: []string{"arbitrary-review-id"}})
	if !got || err != nil {
		t.Fatal(got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyScopeSeveredParticipant(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
			t.Run(kind+"/"+table, func(t *testing.T) {
				db, mock := managementRuntimeDB(t)
				managementLegacyMetadata(mock, managementLegacyCoreNames...)
				managementLegacyEmpty(mock, "el_"+kind)
				if table == "el_mng_result_run" {
					mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_paper_snapshot WHERE").WithArgs(kind, "severed").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
				}
				mock.ExpectQuery("SELECT 1 AS protected FROM "+table+" WHERE").WithArgs(kind, "severed").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
				req := ManagementTraitsLegacyScopeRequest{}
				if kind == "candidate" {
					req.CandidateIDs = []string{"severed"}
				} else {
					req.TesterIDs = []string{"severed"}
				}
				got, err := CheckManagementTraitsLegacyScope(context.Background(), db, req)
				if !got || err != nil {
					t.Fatal(got, err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsLegacyScopeIdentifiersScoped(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		for _, exam := range []string{"", "lookup-exam"} {
			t.Run(kind+"/"+exam, func(t *testing.T) {
				db, mock := managementRuntimeDB(t)
				managementLegacyMetadata(mock, managementLegacyCoreNames...)
				q := mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_" + kind + " WHERE")
				if kind == "candidate" {
					if exam == "" {
						q.WithArgs("phone")
					} else {
						q.WithArgs("phone", exam)
					}
				} else {
					if exam == "" {
						q.WithArgs("identifier", "identifier")
					} else {
						q.WithArgs("identifier", exam, "identifier", exam)
					}
				}
				q.WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"})).RowsWillBeClosed()
				req := ManagementTraitsLegacyScopeRequest{IdentifierExamID: exam}
				if kind == "candidate" {
					req.CandidateTelephones = []string{"phone"}
				} else {
					req.TesterIdentifiers = []string{"identifier"}
				}
				got, err := CheckManagementTraitsLegacyScope(context.Background(), db, req)
				if got || err != nil {
					t.Fatal(got, err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsLegacyScopeMarkerCapabilities(t *testing.T) {
	for _, column := range []string{"paper_id", "exam_id", "candidate_id", "tester_id", "participant_id", "pdf_path", "unsupported"} {
		t.Run(column, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, "el_mng_capture_marker")
			cols := sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_capture_marker", column)
			if column == "participant_id" {
				cols.AddRow("el_mng_capture_marker", "participant_type")
			}
			mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WithArgs("el_mng_capture_marker").WillReturnRows(cols).RowsWillBeClosed()
			req := ManagementTraitsLegacyScopeRequest{}
			switch column {
			case "paper_id":
				req.PaperIDs = []string{"paper"}
				managementLegacyEmpty(mock, "el_paper")
				managementLegacyEmpty(mock, "el_candidate")
				managementLegacyEmpty(mock, "el_tester")
			case "exam_id":
				req.ExamIDs = []string{"exam"}
				for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
					managementLegacyEmpty(mock, table)
				}
			case "candidate_id", "participant_id":
				req.CandidateIDs = []string{"owner"}
				managementLegacyEmpty(mock, "el_candidate")
			case "tester_id":
				req.TesterIDs = []string{"owner"}
				managementLegacyEmpty(mock, "el_tester")
			case "pdf_path":
				req.PDFPaths = []string{`reports\exact.pdf`}
				for _, table := range []string{"el_candidate", "el_tester"} {
					mock.ExpectQuery("SELECT id, exam_id, paper_id FROM " + table + " WHERE.*REPLACE").WithArgs("reports/exact.pdf").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
				}
			}
			if column != "unsupported" {
				mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_capture_marker WHERE").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1)).RowsWillBeClosed()
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, req)
			if column == "unsupported" {
				if got || err != ErrManagementTraitsRuntimeInvalid {
					t.Fatal(got, err)
				}
			} else if !got || err != nil {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeErrorsSanitized(t *testing.T) {
	for _, stage := range []string{"columns", "owner", "marker"} {
		t.Run(stage, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, "el_mng_protection")
			q := mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE")
			if stage == "columns" {
				q.WillReturnError(errors.New("private password"))
			} else {
				q.WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_protection", "candidate_id"))
				q2 := mock.ExpectQuery("SELECT id, exam_id, paper_id FROM el_candidate WHERE")
				if stage == "owner" {
					q2.WillReturnError(errors.New("private password"))
				} else {
					q2.WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}))
					mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_protection WHERE").WillReturnError(errors.New("private password"))
				}
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{CandidateIDs: []string{"owner"}})
			if got || err != ErrManagementTraitsRuntimeInvalid || strings.Contains(fmt.Sprint(err), "private") {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func managementLegacyAwait(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("gate deadlock")
	}
}

func TestManagementTraitsLegacyLeaseFreezeHandoff(t *testing.T) {
	lease := LockManagementTraitsLegacyMutation()
	defer lease.Release()
	freeze := make(chan struct{})
	thaw := make(chan struct{})
	done := make(chan struct{})
	go func() {
		release := LockManagementTraitsRuntimeFreeze()
		close(freeze)
		<-thaw
		release()
		release()
		close(done)
	}()
	// Observe the queued writer under the same mutex rather than relying on sleep.
	deadline := time.Now().Add(3 * time.Second)
	for {
		managementTraitsLegacyGate.mu.Lock()
		waiting := managementTraitsLegacyGate.writersWaiting > 0
		managementTraitsLegacyGate.mu.Unlock()
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("writer never queued")
		}
		runtime.Gosched()
	}
	forked := make(chan func(), 1)
	go func() { forked <- lease.Fork() }()
	var releaseFork func()
	select {
	case releaseFork = <-forked:
	case <-time.After(3 * time.Second):
		t.Fatal("queued writer deadlocked Fork")
	}
	if releaseFork == nil {
		t.Fatal("live lease refused fork")
	}
	lease.Release()
	lease.Release()
	if lease.Fork() != nil {
		t.Fatal("released lease forked")
	}
	select {
	case <-freeze:
		t.Fatal("freeze overtook fork")
	default:
	}
	releaseFork()
	releaseFork()
	managementLegacyAwait(t, freeze)
	reader := make(chan struct{})
	readerDone := make(chan struct{})
	go func() { l := LockManagementTraitsLegacyMutation(); close(reader); l.Release(); close(readerDone) }()
	select {
	case <-reader:
		t.Fatal("reader entered exclusive freeze")
	default:
	}
	close(thaw)
	managementLegacyAwait(t, done)
	managementLegacyAwait(t, readerDone)
	var nilLease *ManagementTraitsLegacyLease
	if nilLease.Fork() != nil {
		t.Fatal("nil lease forked")
	}
	nilLease.Release()
}

func TestManagementTraitsLegacyScopePDFExactOwnership(t *testing.T) {
	for _, kind := range []string{"candidate", "tester"} {
		t.Run(kind, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, managementLegacyCoreNames...)
			for _, table := range []string{"el_candidate", "el_tester"} {
				rows := sqlmock.NewRows([]string{"id", "exam_id", "paper_id"})
				if table == "el_"+kind {
					rows.AddRow("owner", "exam", "paper")
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, exam_id, paper_id FROM " + table + " WHERE (BINARY REPLACE(pdf_path, CHAR(92), '/') = ?)")).WithArgs("reports/Exact.pdf").WillReturnRows(rows)
			}
			for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
				managementLegacyEmpty(mock, table)
			}
			mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{PDFPaths: []string{`reports\Exact.pdf`, "reports/Exact.pdf"}})
			if !got || err != nil {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeExistingIDNotIdentifierScoped(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	managementLegacyMetadata(mock, managementLegacyCoreNames...)
	query := "SELECT id, exam_id, paper_id FROM el_candidate WHERE (id = ?) OR (telephone = ? AND exam_id = ?)"
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("existing", "phone", "different-exam").WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow("existing", "old-exam", "old-paper"))
	for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
		managementLegacyEmpty(mock, table)
	}
	mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_exam_profile WHERE").WithArgs("old-exam").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
	got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{CandidateIDs: []string{"existing"}, CandidateTelephones: []string{"phone"}, IdentifierExamID: "different-exam"})
	if !got || err != nil {
		t.Fatal(got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyScopePaperQuestionOrphanProtection(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	managementLegacyMetadata(mock, managementLegacyCoreNames...)
	for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
		managementLegacyEmpty(mock, table)
	}
	for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run"} {
		mock.ExpectQuery("SELECT 1 AS protected FROM " + table + " WHERE").WithArgs("paper").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
	}
	mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_paper_question_snapshot WHERE").WithArgs("paper").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
	got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{PaperIDs: []string{"paper"}})
	if !got || err != nil {
		t.Fatal(got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyScopeAdditionalCapabilities(t *testing.T) {
	for _, column := range []string{"profile_exam_id", "paper_question_id", "run_id", "result_run_id", "empty", "untyped-participant"} {
		t.Run(column, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			names := []string{"el_mng_report"}
			if column == "run_id" || column == "result_run_id" {
				names = append(names, managementLegacyCoreNames...)
			}
			managementLegacyMetadata(mock, names...)
			cols := sqlmock.NewRows([]string{"table_name", "column_name"})
			if column != "empty" {
				c := column
				if c == "untyped-participant" {
					c = "participant_id"
				}
				cols.AddRow("el_mng_report", c)
			}
			mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WithArgs("el_mng_report").WillReturnRows(cols)
			req := ManagementTraitsLegacyScopeRequest{}
			switch column {
			case "profile_exam_id":
				req.ExamIDs = []string{"exam"}
				for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
					managementLegacyEmpty(mock, table)
				}
			case "paper_question_id":
				req.PaperQuestionIDs = []string{"pq"}
				mock.ExpectQuery("SELECT paper_id FROM el_paper_qu WHERE").WithArgs("pq").WillReturnRows(sqlmock.NewRows([]string{"paper_id"}))
			case "untyped-participant":
				req.TesterIDs = []string{"owner"}
				managementLegacyEmpty(mock, "el_tester")
			case "run_id", "result_run_id":
				req.PaperIDs = []string{"paper"}
				for _, table := range []string{"el_paper", "el_candidate", "el_tester"} {
					managementLegacyEmpty(mock, table)
				}
				for _, table := range []string{"el_mng_paper_snapshot", "el_mng_result_run", "el_mng_paper_question_snapshot"} {
					mock.ExpectQuery("SELECT 1 AS protected FROM " + table + " WHERE").WithArgs("paper").WillReturnRows(sqlmock.NewRows([]string{"protected"}))
				}
			}
			if column != "empty" {
				mock.ExpectQuery("SELECT 1 AS protected FROM el_mng_report WHERE").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1))
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, req)
			if column == "empty" {
				if got || err != ErrManagementTraitsRuntimeInvalid {
					t.Fatal(got, err)
				}
			} else if !got || err != nil {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeBulkNoProtection(t *testing.T) {
	for _, count := range []int{1, 1000, 1001} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			managementLegacyMetadata(mock, managementLegacyCoreNames...)
			ids := make([]string, count)
			for i := range ids {
				ids[i] = fmt.Sprintf("e-%04d", i)
			}
			for _, table := range []string{"el_paper", "el_candidate", "el_tester", "el_mng_exam_profile", "el_mng_paper_snapshot", "el_mng_result_run"} {
				for start := 0; start < count; start += 1000 {
					end := start + 1000
					if end > count {
						end = count
					}
					args := make([]driver.Value, 0, end-start)
					for _, id := range ids[start:end] {
						args = append(args, id)
					}
					mock.ExpectQuery("SELECT .* FROM " + table + " WHERE").WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"protected"})).RowsWillBeClosed()
				}
			}
			got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{ExamIDs: ids})
			if got || err != nil {
				t.Fatal(got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsLegacyScopeMetadataSafety(t *testing.T) {
	for _, names := range [][]string{{"elXmng_report"}, {"el_mng_report;DROP"}, {"el_mng_report", "el_mng_report"}} {
		db, mock := managementRuntimeDB(t)
		managementLegacyMetadata(mock, names...)
		if got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{}); got || err != ErrManagementTraitsRuntimeInvalid {
			t.Fatal(got, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
	for missing := range managementLegacyCoreNames {
		db, mock := managementRuntimeDB(t)
		names := append([]string{}, managementLegacyCoreNames[:missing]...)
		names = append(names, managementLegacyCoreNames[missing+1:]...)
		managementLegacyMetadata(mock, names...)
		if got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{}); got || err != ErrManagementTraitsRuntimeInvalid {
			t.Fatal(got, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
	db, mock := managementRuntimeDB(t)
	managementLegacyMetadata(mock, "el_mng_report")
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name"}).AddRow("el_mng_report", "run_id"))
	if got, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{}); got || err != ErrManagementTraitsRuntimeInvalid {
		t.Fatal(got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsLegacyLeaseWriterPreference(t *testing.T) {
	lease := LockManagementTraitsLegacyMutation()
	writerEntered, releaseWriter, writerDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		release := LockManagementTraitsRuntimeFreeze()
		close(writerEntered)
		<-releaseWriter
		release()
		close(writerDone)
	}()
	deadline := time.Now().Add(3 * time.Second)
	for {
		managementTraitsLegacyGate.mu.Lock()
		waiting := managementTraitsLegacyGate.writersWaiting > 0
		managementTraitsLegacyGate.mu.Unlock()
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("writer not queued")
		}
		runtime.Gosched()
	}
	readerStarted, readerEntered := make(chan struct{}), make(chan struct{})
	go func() {
		close(readerStarted)
		l := LockManagementTraitsLegacyMutation()
		close(readerEntered)
		l.Release()
	}()
	managementLegacyAwait(t, readerStarted)
	lease.Release()
	managementLegacyAwait(t, writerEntered)
	select {
	case <-readerEntered:
		t.Fatal("fresh reader overtook queued writer")
	default:
	}
	close(releaseWriter)
	managementLegacyAwait(t, writerDone)
	managementLegacyAwait(t, readerEntered)
}

func TestManagementTraitsLegacyLeaseConcurrentForkRelease(t *testing.T) {
	for i := 0; i < 100; i++ {
		lease := LockManagementTraitsLegacyMutation()
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			if release := lease.Fork(); release != nil {
				release()
				release()
			}
		}()
		go func() { defer wg.Done(); lease.Release(); lease.Release() }()
		wg.Wait()
		managementTraitsLegacyGate.mu.Lock()
		readers := managementTraitsLegacyGate.readers
		managementTraitsLegacyGate.mu.Unlock()
		if readers != 0 || lease.Fork() != nil {
			t.Fatal("leaked reader or revived lease", readers)
		}
	}
}
