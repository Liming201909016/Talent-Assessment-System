package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Explicit real staging opt-in. The caller owns restoration, DDL and deletion.
// Only the exact disposable schema is accessible; no credentials are logged.
type managementStagingObserver struct {
	*sql.DB
	columns [][]string
	rows    []int64
}

func (o *managementStagingObserver) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	r, err := o.DB.QueryContext(ctx, query, args...)
	if err == nil && (strings.Contains(query, "engine AS engine") || strings.Contains(query, "column_type AS column_type") || strings.Contains(query, "prefix_length") || strings.Contains(query, "r.update_rule AS update_rule")) {
		c, e := r.Columns()
		if e != nil {
			r.Close()
			return nil, e
		}
		o.columns = append(o.columns, c)
	}
	return r, err
}

type managementStagingLogger struct {
	logger.Interface
	observer *managementStagingObserver
}

func (l managementStagingLogger) Trace(_ context.Context, _ time.Time, fc func() (string, int64), err error) {
	q, n := fc()
	if strings.Contains(q, "engine AS engine") || strings.Contains(q, "column_type AS column_type") || strings.Contains(q, "prefix_length") || strings.Contains(q, "r.update_rule AS update_rule") {
		if err != nil {
			n = -1
		}
		l.observer.rows = append(l.observer.rows, n)
	}
}

func TestManagementTraitsGuardAuditMySQLStagingExternal(t *testing.T) {
	dsn, owned := os.Getenv("MNG_GUARD_AUDIT_MYSQL_DSN"), os.Getenv("MNG_GUARD_AUDIT_MYSQL_SCHEMA")
	if dsn == "" {
		t.Skip("dedicated staging owned database absent")
	}
	dc, err := driver.ParseDSN(dsn)
	if err != nil || !regexp.MustCompile(`^mng_guard_audit_test_[0-9a-f]{16}$`).MatchString(owned) || dc.DBName != owned || dc.MultiStatements || dc.InterpolateParams || os.Getenv("REPORT_EFFECTIVE_ENV") != "staging" {
		t.Fatal("staging ownership assertion failed")
	}
	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal("owned pool open failed")
	}
	defer pool.Close()
	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)
	o := &managementStagingObserver{DB: pool}
	l := managementStagingLogger{Interface: logger.Default.LogMode(logger.Silent), observer: o}
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: o, SkipInitializeWithVersion: true}), &gorm.Config{Logger: l, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal("owned GORM open failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 110*time.Second)
	defer cancel()
	db = db.WithContext(ctx)
	var actual string
	var cross int64
	if db.Raw("SELECT DATABASE()").Scan(&actual).Error != nil || actual != owned || db.Raw("SELECT COUNT(*) FROM information_schema.key_column_usage WHERE table_schema=DATABASE() AND referenced_table_schema IS NOT NULL AND referenced_table_schema<>DATABASE()").Scan(&cross).Error != nil || cross != 0 {
		t.Fatal("actual database/cross-schema ownership failed")
	}
	cfg := config.Load()
	if cfg.Jwt.Secret == "" {
		t.Fatal("actual configuration has no signing secret")
	}
	s := NewManagementTraitsRuntimeService(db, cfg.Jwt.Secret, 8<<20)
	if s.CheckRuntimeSchema(ctx) != nil || s.CheckRuntimeSchema(ctx) != nil || len(o.rows) != 4 || len(o.columns) != 4 {
		t.Fatal("fresh full production schema/cache failed")
	}
	for i, c := range o.columns {
		t.Logf("ACTUAL_METADATA_%d columns=%s rows=%d", i+1, strings.Join(c, ","), o.rows[i])
		if o.rows[i] != []int64{11, 159, 67, 21}[i] {
			t.Fatal("actual production metadata shape changed")
		}
	}
	if os.Getenv("MNG_STAGING_SCHEMA_ONLY") == "1" {
		return
	}
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal("contract unavailable")
	}
	empty := func() {
		t.Helper()
		for _, table := range d.Tables {
			var n int64
			if db.Table(table.Name).Count(&n).Error != nil || n != 0 {
				t.Fatal("owned sidecar residual rows")
			}
		}
	}
	empty()
	for _, code := range []string{"00201", "00202"} {
		var repo model.Repo
		if db.Where("code = ?", code).Take(&repo).Error != nil {
			t.Fatal("restored source missing")
		}
		id := uuid.NewString()
		exam := model.Exam{ID: id, Title: "owned-staging-freeze", AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, RequiredFields: "name", IsOpen: 1}
		link := model.ExamRepo{ID: uuid.NewString(), ExamID: id, RepoID: repo.ID, RadioCount: 140}
		if db.Create(&exam).Error != nil || db.Create(&link).Error != nil {
			t.Fatal("owned freeze setup failed")
		}
		profile, e := s.FreezeProfile(ctx, id)
		if e != nil {
			t.Fatal("actual public freeze rejected")
		}
		var bundle model.ManagementTraitsDefinitionBundle
		if db.Where("id = ?", profile.BundleID).Take(&bundle).Error != nil {
			t.Fatal("frozen bundle missing")
		}
		manifest, e := DecodeManagementTraitsManifest([]byte(bundle.ScoringManifest), 8<<20)
		if e != nil {
			t.Fatal("actual frozen manifest rejected")
		}
		mapping, e := DecodeManagementTraitsMapping([]byte(profile.MappingSnapshot), manifest, 8<<20)
		options := 0
		for _, q := range mapping.Questions {
			options += len(q.Options)
		}
		if e != nil || len(manifest.Questions) != 140 || len(mapping.Questions) != 140 || options != 700 {
			t.Fatal("actual frozen source cardinality failed")
		}
		t.Logf("ACTUAL_FREEZE code=%s questions=140 options=700 manifest=%s mapping=%s", code, bundle.ScoringManifestSHA, profile.MappingSHA)
		for _, row := range []struct{ table, key, value string }{{"el_mng_exam_profile", "exam_id", id}, {"el_mng_definition_bundle", "id", bundle.ID}, {"el_exam_repo", "id", link.ID}, {"el_exam", "id", id}} {
			if db.Table(row.table).Where(row.key+" = ?", row.value).Delete(map[string]any{}).Error != nil {
				t.Fatal("owned freeze cleanup failed")
			}
		}
	}
	empty()
	now := time.Now()
	claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: "owned-capture-candidate", ExamID: "owned-capture-exam", ExpiresAt: now.Add(time.Minute).Unix()}
	token, e := CreateManagementTraitsRuntimeToken(cfg.Jwt.Secret, claims, now)
	parsed, pErr := ParseManagementTraitsRuntimeToken(cfg.Jwt.Secret, token, claims.Purpose, now)
	if e != nil || pErr != nil || parsed != claims {
		t.Fatal("actual configuration token roundtrip failed")
	}
	t.Log("ACTUAL_CONFIG_TOKEN_ROUNDTRIP_PASS")
	if db.Exec("CREATE TABLE el_mng_owned_capture (participant_type varchar(255) NOT NULL, participant_id varchar(64) NOT NULL, paper_id varchar(64) NOT NULL) ENGINE=InnoDB").Error != nil {
		t.Fatal("exact owned capture creation failed")
	}
	defer func() {
		if db.Exec("DROP TABLE el_mng_owned_capture").Error != nil {
			t.Error("owned capture cleanup failed")
		}
	}()
	if db.Exec("INSERT INTO el_mng_owned_capture (participant_type,participant_id,paper_id) VALUES (?,?,?)", "candidate", claims.ParticipantID, "owned-capture-paper").Error != nil {
		t.Fatal("owned capture insert failed")
	}
	capture := func(label string) {
		t.Helper()
		if protected, e := ManagementTraitsIdentityScope(ctx, db, claims.ExamID, "candidate", claims.ParticipantID); e != nil || !protected {
			t.Fatal("actual public capture identity failed", label)
		}
		for _, req := range []ManagementTraitsLegacyScopeRequest{{CandidateIDs: []string{claims.ParticipantID}}, {PaperIDs: []string{"owned-capture-paper"}}, {AllLegacy: true}} {
			if protected, e := CheckManagementTraitsLegacyScope(ctx, db, req); e != nil || !protected {
				t.Fatal("actual capture legacy scope failed", label)
			}
		}
		t.Log("ACTUAL_CAPTURE_PUBLIC_IDENTITY_ALLLEGACY_PASS", label)
	}
	capture("11-core")
	// Move only four empty owned extras outside the prefix, then restore exactly.
	const hide = "RENAME TABLE el_mng_report_audit TO owned_hold_audit, el_mng_report_current TO owned_hold_current, el_mng_report_revision TO owned_hold_revision, el_mng_runtime_receipt TO owned_hold_receipt"
	const restore = "RENAME TABLE owned_hold_audit TO el_mng_report_audit, owned_hold_current TO el_mng_report_current, owned_hold_revision TO el_mng_report_revision, owned_hold_receipt TO el_mng_runtime_receipt"
	if db.Exec(hide).Error != nil {
		t.Fatal("owned seven-core setup failed")
	}
	restored := false
	defer func() {
		if !restored && db.Exec(restore).Error != nil {
			t.Error("owned full schema restoration failed")
		}
	}()
	capture("7-core")
	if db.Exec(restore).Error != nil {
		t.Fatal("owned full schema restore failed")
	}
	restored = true
	if db.Exec("DELETE FROM el_mng_owned_capture WHERE participant_id = ?", claims.ParticipantID).Error != nil {
		t.Fatal("owned capture row cleanup failed")
	}
	if !t.Run("canonical-audit-rollback", TestManagementTraitsGuardAuditMySQLExternal) {
		t.Fatal("original real external audit test failed")
	}
	// Corrupt rows only inside this disposable connection/transaction. No FK
	// metadata is changed, and checks are restored before returning the pool.
	for _, kind := range []string{"missing-revision", "missing-run", "identity-drift"} {
		if !t.Run(kind, func(t *testing.T) {
			tx := db.Begin()
			if tx.Error != nil {
				t.Fatal("owned negative transaction failed")
			}
			defer tx.Rollback()
			if tx.Exec("SET SESSION FOREIGN_KEY_CHECKS=0").Error != nil {
				t.Fatal("owned negative fixture isolation failed")
			}
			defer func() {
				if tx.Exec("SET SESSION FOREIGN_KEY_CHECKS=1").Error != nil {
					t.Error("owned FK-check reset failed")
				}
			}()
			id, runID, hash := uuid.NewString(), uuid.NewString(), strings.Repeat("a", 64)
			if kind == "identity-drift" {
				r := model.ManagementTraitsResultRun{ID: runID, PaperID: "drift-paper", ExamID: "drift-exam", CreatedAt: now}
				if tx.Create(&r).Error != nil {
					t.Fatal("owned drift run failed")
				}
			}
			if kind != "missing-revision" {
				r := model.ManagementTraitsReportRevision{ID: id, RunID: runID, PaperID: "other-paper", ExamID: "other-exam", Revision: 1, Mode: "test", DataSnapshot: "{}", ContentSHA: hash, TemplateSHA: hash, DataSHA: hash, FileSHA: hash, CreatedAt: now}
				if tx.Create(&r).Error != nil {
					t.Fatal("owned orphan revision failed")
				}
			}
			a := model.ManagementTraitsReportAudit{ID: uuid.NewString(), ReportID: id, ActorID: 1, Action: "generate", CreatedAt: now}
			if tx.Create(&a).Error != nil {
				t.Fatal("owned orphan audit failed")
			}
			for _, req := range []ManagementTraitsLegacyScopeRequest{{}, {AllLegacy: true}, {PaperIDs: []string{"unrelated"}}} {
				if protected, e := CheckManagementTraitsLegacyScope(ctx, tx, req); protected || !errors.Is(e, ErrManagementTraitsRuntimeInvalid) {
					t.Fatal("actual orphan/drift was not fail closed")
				}
			}
		}) {
			t.Fatal("real negative closure failed")
		}
		empty()
	}
	if db.Exec("CREATE TABLE el_mng_unknown_owned (report_id varchar(64) NOT NULL) ENGINE=InnoDB").Error != nil {
		t.Fatal("owned unknown fixture failed")
	}
	protected, e := CheckManagementTraitsLegacyScope(ctx, db, ManagementTraitsLegacyScopeRequest{AllLegacy: true})
	if db.Exec("DROP TABLE el_mng_unknown_owned").Error != nil {
		t.Fatal("owned unknown cleanup failed")
	}
	if protected || !errors.Is(e, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("actual unknown marker not fail closed")
	}
	if db.Exec("RENAME TABLE el_mng_runtime_receipt TO owned_hold_receipt").Error != nil {
		t.Fatal("owned partial fixture failed")
	}
	protected, e = CheckManagementTraitsLegacyScope(ctx, db, ManagementTraitsLegacyScopeRequest{AllLegacy: true})
	if db.Exec("RENAME TABLE owned_hold_receipt TO el_mng_runtime_receipt").Error != nil {
		t.Fatal("owned partial restore failed")
	}
	if protected || !errors.Is(e, ErrManagementTraitsRuntimeInvalid) {
		t.Fatal("actual partial install not fail closed")
	}
	empty()
	if NewManagementTraitsRuntimeService(db, cfg.Jwt.Secret, 8<<20).CheckRuntimeSchema(ctx) != nil {
		t.Fatal("post-negative fresh full schema failed")
	}
	fmt.Println("STAGING_ACTUAL_GUARD_CLOSURE_PASS SIDECAR_ROWS=0")
}
