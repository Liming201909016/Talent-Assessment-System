package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MT-REAL-TESTER-SCOPE: main owns the regression ledger and parser fix.
// Only the observed tester ID is reused; all credentials and row data are synthetic.
const managementTraitsNestedTesterID = "1791035468793800801"
const managementTraitsNestedPassword = "synthetic-nested-scope-only"
const managementTraitsNestedOrphanSQL = "SELECT 1 AS protected FROM el_mng_report_audit a LEFT JOIN el_mng_report_revision r ON r.id = a.report_id LEFT JOIN el_mng_result_run u ON u.id = r.run_id AND u.paper_id = r.paper_id AND u.exam_id = r.exam_id WHERE r.id IS NULL OR u.id IS NULL LIMIT 1"
const managementTraitsNestedAuditSQL = "SELECT 1 AS protected FROM el_mng_report_audit WHERE (report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE (participant_type = ? AND participant_id = ?)))) LIMIT 1"

// Driver hits are recorded by sqlmock's matcher, not by a pre-driver callback.
// Trace records rejection separately and never prints interpolated SQL or values.
type managementTraitsNestedSQLLog struct {
	logger.Interface
	driverQueries, scopeQueries, orphanQueries, auditQueries int
	auditRejected, otherErrors, writeAttempts                int
}

func (l *managementTraitsNestedSQLLog) Match(expected, actual string) error {
	if sqlmock.QueryMatcherRegexp.Match(expected, actual) != nil {
		return errors.New("nested scope fixture SQL shape mismatch")
	}
	l.driverQueries++
	if strings.Contains(actual, "information_schema") {
		l.scopeQueries++
	}
	if actual == managementTraitsNestedOrphanSQL {
		l.orphanQueries++
	}
	if actual == managementTraitsNestedAuditSQL {
		l.auditQueries++
	}
	return nil
}

func (l *managementTraitsNestedSQLLog) Trace(_ context.Context, _ time.Time, sql func() (string, int64), err error) {
	query, _ := sql()
	upper := strings.ToUpper(strings.TrimSpace(query))
	for _, verb := range []string{"INSERT ", "UPDATE ", "DELETE ", "REPLACE ", "CREATE ", "ALTER ", "DROP "} {
		if strings.HasPrefix(upper, verb) {
			l.writeAttempts++
		}
	}
	if err == nil {
		return
	}
	if errors.Is(err, service.ErrManagementTraitsRuntimeInvalid) && strings.HasPrefix(query, "SELECT 1 AS protected FROM el_mng_report_audit WHERE (report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE ") {
		l.auditRejected++
		return
	}
	l.otherErrors++
}

func managementTraitsNestedSetup(t *testing.T) (*TesterHandler, sqlmock.Sqlmock, *managementTraitsNestedSQLLog) {
	t.Helper()
	l := &managementTraitsNestedSQLLog{Interface: logger.Default.LogMode(logger.Silent)}
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(l))
	if err != nil {
		t.Fatal("cannot initialize isolated mock driver")
	}
	t.Cleanup(func() { _ = db.Close() })
	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{Logger: l})
	if err != nil {
		t.Fatal("cannot initialize isolated GORM database")
	}
	cfg := managementTraitsDIConfig()
	s := service.NewManagementTraitsRuntimeService(gdb, cfg.Jwt.Secret, 1<<20)
	identityHandlerRuntimeSchema(t, mock)
	for i := 0; i < 2; i++ {
		if s.CheckRuntimeSchema(context.Background()) != nil {
			t.Fatal("complete 11-table DDL-derived warm schema rejected")
		}
	}
	if mock.ExpectationsWereMet() != nil || l.driverQueries != 4 || l.otherErrors != 0 {
		t.Fatal("warm schema did not consume exactly four metadata queries")
	}
	h := NewTesterHandler(gdb, cfg, s)
	if h.managementTraitsRuntime != s {
		t.Fatal("real tester handler did not retain injected warmed service")
	}
	l.driverQueries, l.scopeQueries = 0, 0
	return h, mock, l
}

func managementTraitsNestedTesterRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "exam_id", "paper_id", "name", "password", "telephone", "del_flag", "status"}).AddRow(managementTraitsNestedTesterID, "exam", "", "Person", managementTraitsNestedPassword, "13800000000", 0, "0")
}

func managementTraitsNestedLookup(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("^"+regexp.QuoteMeta("SELECT * FROM `el_tester` WHERE ((telephone = ? OR id_number = ? OR id = ?) AND (del_flag IS NULL OR del_flag = '0')) AND exam_id = ?")+"$").WithArgs(managementTraitsNestedTesterID, managementTraitsNestedTesterID, managementTraitsNestedTesterID, "exam").WillReturnRows(managementTraitsNestedTesterRows()).RowsWillBeClosed()
}

// The existing full-participant helper intentionally exposes only seven core
// tables. This independent scope fixture exposes all eleven actual DDL tables
// and every column of the four extra tables; no audit identity columns invented.
func managementTraitsNestedFullScope(t *testing.T, mock sqlmock.Sqlmock, orphan bool) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "sql", "management_traits_001_runtime.sql"))
	if err != nil {
		t.Fatal("cannot read unchanged sidecar DDL")
	}
	blocks := regexp.MustCompile(`(?s)CREATE TABLE IF NOT EXISTS (el_mng_[a-z_]+) \((.*?)\) ENGINE=InnoDB`).FindAllStringSubmatch(strings.ReplaceAll(string(raw), "\r\n", "\n"), -1)
	if len(blocks) != 11 {
		t.Fatal("scope fixture requires exactly eleven sidecar tables")
	}
	names := make([]string, 0, 11)
	extra := map[string]bool{"el_mng_report_audit": true, "el_mng_report_current": true, "el_mng_report_revision": true, "el_mng_runtime_receipt": true}
	columns := sqlmock.NewRows([]string{"table_name", "column_name"})
	columnRE := regexp.MustCompile("(?m)^`([a-z_]+)` ")
	for _, block := range blocks {
		names = append(names, block[1])
		if extra[block[1]] {
			for _, c := range columnRE.FindAllStringSubmatch(block[2], -1) {
				columns.AddRow(block[1], c[1])
			}
		}
	}
	sort.Strings(names)
	tables := sqlmock.NewRows([]string{"table_name"})
	for _, name := range names {
		tables.AddRow(name)
	}
	metadata := "SELECT table_name AS table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name,7) = 'el_mng_' ORDER BY table_name"
	mock.ExpectQuery("^" + regexp.QuoteMeta(metadata) + "$").WithoutArgs().WillReturnRows(tables).RowsWillBeClosed()
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE").WithArgs("el_mng_report_audit", "el_mng_report_current", "el_mng_report_revision", "el_mng_runtime_receipt").WillReturnRows(columns).RowsWillBeClosed()
	identityHandlerRuntimeSchema(t, mock)
	rows := sqlmock.NewRows([]string{"protected"})
	if orphan {
		rows.AddRow(1)
	}
	mock.ExpectQuery("^" + regexp.QuoteMeta(managementTraitsNestedOrphanSQL) + "$").WithoutArgs().WillReturnRows(rows).RowsWillBeClosed()
	if orphan {
		return
	}
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT id, exam_id, paper_id FROM el_tester WHERE (id = ?)") + "$").WithArgs(managementTraitsNestedTesterID).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow(managementTraitsNestedTesterID, "exam", "")).RowsWillBeClosed()
	mock.ExpectQuery("^"+regexp.QuoteMeta(managementTraitsNestedAuditSQL)+"$").WithArgs("tester", managementTraitsNestedTesterID).WillReturnRows(sqlmock.NewRows([]string{"protected"})).RowsWillBeClosed()
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT 1 AS protected FROM el_mng_exam_profile WHERE (exam_id = ?) LIMIT 1") + "$").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"protected"}).AddRow(1)).RowsWillBeClosed()
}

func managementTraitsNestedIdentityScope(t *testing.T, mock sqlmock.Sqlmock, orphan bool) {
	t.Helper()
	identityHandlerSchema(mock, 1)
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1)).RowsWillBeClosed()
	mock.ExpectQuery("SELECT .*el_tester.*id =").WithArgs(managementTraitsNestedTesterID, 2).WillReturnRows(sqlmock.NewRows([]string{"id", "exam_id", "paper_id"}).AddRow(managementTraitsNestedTesterID, "exam", "")).RowsWillBeClosed()
	mock.ExpectQuery("SELECT count.*el_mng_exam_profile").WithArgs("exam").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1)).RowsWillBeClosed()
	mock.ExpectQuery("SELECT count.*el_mng_paper_snapshot").WithArgs("tester", managementTraitsNestedTesterID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0)).RowsWillBeClosed()
	mock.ExpectQuery("SELECT count.*el_mng_result_run").WithArgs("tester", managementTraitsNestedTesterID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0)).RowsWillBeClosed()
	managementTraitsNestedFullScope(t, mock, orphan)
}

func managementTraitsNestedLogin(t *testing.T, h *TesterHandler, password string) (int, int, map[string]json.RawMessage) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/exam/api/tester/login", h.LoginForm)
	body, err := json.Marshal(map[string]string{"examId": "exam", "idNumber": managementTraitsNestedTesterID, "password": password})
	if err != nil {
		t.Fatal("cannot construct synthetic login request")
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/exam/api/tester/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if strings.Contains(w.Body.String(), managementTraitsNestedPassword) || strings.Contains(w.Body.String(), password) || strings.Contains(w.Body.String(), h.cfg.Jwt.Secret) || strings.Contains(w.Body.String(), `"password"`) || strings.Contains(w.Body.String(), `"token"`) {
		t.Fatal("login response exposed a credential or legacy token field")
	}
	var result struct {
		Code int                        `json:"code"`
		Data map[string]json.RawMessage `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &result) != nil {
		t.Fatal("login response is not valid JSON")
	}
	return w.Code, result.Code, result.Data
}

func TestBugManagementTraitsTesterNestedScopeLogin(t *testing.T) {
	h, mock, l := managementTraitsNestedSetup(t)
	managementTraitsNestedLookup(mock)
	managementTraitsNestedIdentityScope(t, mock, false)
	mock.ExpectBegin()
	identityHandlerAdmission(t, mock, 2)
	mock.ExpectQuery("SELECT .*el_tester.*FOR UPDATE").WithArgs(managementTraitsNestedTesterID, 2).WillReturnRows(managementTraitsNestedTesterRows()).RowsWillBeClosed()
	mock.ExpectCommit()
	httpCode, code, data := managementTraitsNestedLogin(t, h, managementTraitsNestedPassword)
	if l.otherErrors != 0 || l.writeAttempts != 0 || l.orphanQueries != 1 {
		t.Fatalf("fixture or non-scope failure: otherErrors=%d writes=%d orphanDriverHits=%d", l.otherErrors, l.writeAttempts, l.orphanQueries)
	}
	if l.auditQueries != 1 {
		// Intentionally pending audit/transaction expectations are consequences,
		// not the RED cause: the actual callback returned the domain sentinel.
		t.Fatalf("NESTED_SCOPE_CAPACITY_RED: HTTP=%d businessCode=%d auditDriverHits=%d preDriverDomainRejects=%d unexpectedDriverErrors=%d legacyWrites=%d; legal 19-byte tester ID rejected before nested audit SQL reached driver", httpCode, code, l.auditQueries, l.auditRejected, l.otherErrors, l.writeAttempts)
	}
	if mock.ExpectationsWereMet() != nil {
		t.Fatal("driver expectations incomplete after nested scope admission")
	}
	if httpCode != 200 || code != 200 || len(data) != 5 || l.auditRejected != 0 {
		t.Fatalf("legal login did not return five-key identity: HTTP=%d businessCode=%d keys=%d", httpCode, code, len(data))
	}
	for _, key := range []string{"id", "examId", "paperId", "name", "participantToken"} {
		if _, ok := data[key]; !ok {
			t.Fatal("five-key identity missing required field", key)
		}
	}
	var view service.ManagementTraitsIdentityResponse
	encoded, _ := json.Marshal(data)
	if json.Unmarshal(encoded, &view) != nil || view.ID != managementTraitsNestedTesterID || view.ExamID != "exam" || view.PaperID != "" || view.Name != "Person" {
		t.Fatal("identity response has incorrect frozen bindings")
	}
	claims, err := service.ParseManagementTraitsRuntimeToken(h.cfg.Jwt.Secret, view.ParticipantToken, service.ManagementTraitsRuntimeParticipantPurpose, time.Now())
	if err != nil || claims.ValidateBinding("tester", managementTraitsNestedTesterID, "exam", "") != nil {
		t.Fatal("participant credential failed binding verification")
	}
	t.Log("LOGIN_PASS: auditDriverHits=1 orphanDriverHits=1 identityKeys=5 legacyWrites=0")
}

func TestBugManagementTraitsTesterNestedScopeWrongPassword(t *testing.T) {
	h, mock, l := managementTraitsNestedSetup(t)
	managementTraitsNestedLookup(mock)
	httpCode, code, data := managementTraitsNestedLogin(t, h, "synthetic-wrong-only")
	if httpCode != 200 || code == 200 || len(data) != 0 || l.driverQueries != 1 || l.scopeQueries != 0 || l.auditQueries != 0 || l.orphanQueries != 0 || l.writeAttempts != 0 || l.otherErrors != 0 || mock.ExpectationsWereMet() != nil {
		t.Fatal("wrong password must stop at identifier lookup with zero scope and writes")
	}
	t.Log("WRONG_PASSWORD_PASS: lookupDriverHits=1 scopeDriverHits=0 writes=0")
}

func TestBugManagementTraitsTesterNestedScopeOrphanFailClosed(t *testing.T) {
	h, mock, l := managementTraitsNestedSetup(t)
	managementTraitsNestedLookup(mock)
	managementTraitsNestedIdentityScope(t, mock, true)
	httpCode, code, data := managementTraitsNestedLogin(t, h, managementTraitsNestedPassword)
	if httpCode != 200 || code == 200 || len(data) != 0 || l.orphanQueries != 1 || l.auditQueries != 0 || l.writeAttempts != 0 || l.otherErrors != 0 || mock.ExpectationsWereMet() != nil {
		t.Fatalf("composite orphan must fail closed: HTTP=%d businessCode=%d orphanDriverHits=%d nestedDriverHits=%d writes=%d otherErrors=%d", httpCode, code, l.orphanQueries, l.auditQueries, l.writeAttempts, l.otherErrors)
	}
	t.Log("ORPHAN_PASS: exactRevisionRunCompositeSQL=1 returnedRows=1 nestedDriverHits=0 writes=0")
}
