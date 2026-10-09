package service

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Opt-in only: use a caller-provisioned EMPTY disposable localhost schema named
// mng_source_lock_test_*. Never installs the application schema, touches a
// running deployment, or connects when the dedicated DSN is not supplied.
func TestManagementTraitsSourceLockMySQLExternalUpdate(t *testing.T) {
	dsn := os.Getenv("MNG_SOURCE_LOCK_MYSQL_DSN")
	if dsn == "" {
		t.Skip("dedicated disposable local MySQL source-lock schema not configured")
	}
	cfg, err := drivermysql.ParseDSN(dsn)
	if err != nil || cfg.Net != "tcp" || (cfg.Addr != "127.0.0.1:3306" && !strings.HasPrefix(cfg.Addr, "127.0.0.1:")) || !strings.HasPrefix(cfg.DBName, "mng_source_lock_test_") || !cfg.ParseTime {
		t.Fatal("requires disposable localhost schema and parseTime=true; DSN withheld")
	}
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("local source-lock test connection failed; details withheld")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal("cannot open test pool")
	}
	defer pool.Close()
	pool.SetMaxOpenConns(3)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db = db.WithContext(ctx)
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&count).Error; err != nil || count != 0 {
		t.Fatal("test schema must be empty")
	}
	// CREATE without IF NOT EXISTS: an existing application bundle table can
	// never be adopted or cleaned by this test. Only this newly created table is
	// dropped; caller-provisioned schema and all unrelated objects stay intact.
	if err := db.Exec(`CREATE TABLE el_mng_definition_bundle (
		id varchar(64) PRIMARY KEY,
		product_version varchar(64) NOT NULL, question_version varchar(64) NOT NULL,
		scoring_version varchar(64) NOT NULL, norm_version varchar(64) NOT NULL,
		questionnaire varchar(255) NOT NULL, scoring_manifest longtext NOT NULL,
		scoring_manifest_sha char(64) NOT NULL, status varchar(255) NOT NULL,
		created_at datetime NOT NULL
	) ENGINE=InnoDB`).Error; err != nil {
		t.Fatal("cannot create disposable lock fixture")
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := db.WithContext(cleanup).Exec("DROP TABLE el_mng_definition_bundle").Error; err != nil {
			t.Error("disposable fixture cleanup failed")
		}
	}()
	bundle := model.ManagementTraitsDefinitionBundle{ID: "source-lock-fixture", Status: "candidate-current-source", CreatedAt: time.Now().Truncate(time.Second)}
	if db.Create(&bundle).Error != nil {
		t.Fatal("cannot insert disposable source")
	}
	for _, finish := range []string{"commit", "rollback"} {
		t.Run(finish, func(t *testing.T) {
			if db.Model(&bundle).Where("id = ?", bundle.ID).Update("status", "candidate-current-source").Error != nil {
				t.Fatal("cannot reset disposable source")
			}
			writer := db.Begin()
			if writer.Error != nil {
				t.Fatal("cannot begin writer")
			}
			defer writer.Rollback()
			if _, err := loadManagementTraitsWriteBundle(ctx, writer, bundle.ID); err != nil {
				t.Fatal("writer shared read rejected")
			}
			// A second transaction must acquire a compatible SHARE lock, not an
			// exclusive per-bundle mutex that serializes unrelated participants.
			reader := db.Begin()
			if reader.Error != nil {
				t.Fatal("cannot begin compatible reader")
			}
			defer reader.Rollback()
			if reader.Exec("SET SESSION innodb_lock_wait_timeout = 1").Error != nil {
				t.Fatal("cannot bound reader wait")
			}
			if _, err := loadManagementTraitsWriteBundle(ctx, reader, bundle.ID); err != nil {
				t.Fatal("shared leases incompatible")
			}
			if reader.Rollback().Error != nil {
				t.Fatal("reader rollback failed")
			}
			// Real InnoDB, separate connection: ordinary external UPDATE (no
			// application gate or special revocation API) must hit lock timeout.
			err := db.Transaction(func(revoker *gorm.DB) error {
				if err := revoker.Exec("SET SESSION innodb_lock_wait_timeout = 1").Error; err != nil {
					return err
				}
				return revoker.Exec("UPDATE el_mng_definition_bundle SET status = ? WHERE id = ?", "review-revoked", bundle.ID).Error
			})
			var lockErr *drivermysql.MySQLError
			if !errors.As(err, &lockErr) || lockErr.Number != 1205 {
				t.Fatal("external UPDATE was not blocked by writer shared lease")
			}
			if finish == "commit" {
				err = writer.Commit().Error
			} else {
				err = writer.Rollback().Error
			}
			if err != nil {
				t.Fatal("writer finish failed")
			}
			if db.Exec("UPDATE el_mng_definition_bundle SET status = ? WHERE id = ?", "review-revoked", bundle.ID).Error != nil {
				t.Fatal("external UPDATE still blocked after writer finish")
			}
			err = db.Transaction(func(tx *gorm.DB) error { _, err := loadManagementTraitsWriteBundle(ctx, tx, bundle.ID); return err })
			if !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
				t.Fatal("revoker-first current read accepted source")
			}
		})
	}
	t.Run("revocation-after-MVCC-snapshot", func(t *testing.T) {
		if db.Model(&bundle).Where("id = ?", bundle.ID).Update("status", "candidate-current-source").Error != nil {
			t.Fatal("cannot reset disposable source")
		}
		stale := db.Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if stale.Error != nil {
			t.Fatal("cannot begin repeatable-read transaction")
		}
		defer stale.Rollback()
		var observed string
		if stale.Raw("SELECT status FROM el_mng_definition_bundle WHERE id = ?", bundle.ID).Scan(&observed).Error != nil || observed != "candidate-current-source" {
			t.Fatal("cannot establish MVCC snapshot")
		}
		if db.Exec("UPDATE el_mng_definition_bundle SET status = ? WHERE id = ?", "review-revoked", bundle.ID).Error != nil {
			t.Fatal("external revocation failed")
		}
		if stale.Raw("SELECT status FROM el_mng_definition_bundle WHERE id = ?", bundle.ID).Scan(&observed).Error != nil || observed != "candidate-current-source" {
			t.Fatal("repeatable-read snapshot not retained")
		}
		if _, err := loadManagementTraitsWriteBundle(ctx, stale, bundle.ID); !errors.Is(err, ErrManagementTraitsRuntimeInvalid) {
			t.Fatal("locking read used stale source approval")
		}
		if stale.Rollback().Error != nil {
			t.Fatal("revoked reader rollback failed")
		}
	})
}
