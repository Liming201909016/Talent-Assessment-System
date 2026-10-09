package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func sqlUTF8Hex(value string) string {
	return "CONVERT(0x" + hex.EncodeToString([]byte(value)) + " USING utf8mb4)"
}

func TestManagementTraitsSharedStagingProfileV2Export(t *testing.T) {
	if os.Getenv("MNG005_SHARED_PROFILE_EXPORT") != "1" {
		t.Skip("shared staging profile export is explicit opt-in")
	}
	hostname, err := os.Hostname()
	output := os.Getenv("MNG005_SHARED_PROFILE_EXPORT_PATH")
	if err != nil || hostname != "vm-ubuntu-go-dev" || os.Getenv("REPORT_EFFECTIVE_ENV") != "staging" || output != "/tmp/mng005-profile-v2-export.sql" {
		t.Fatal("staging read-only profile export gate")
	}
	cfg := config.Load()
	db, err := gorm.Open(mysql.Open(cfg.Mysql.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("database connection")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	item := sharedBaselineCases[1]
	var database string
	var profile model.ManagementTraitsExamProfile
	if db.WithContext(ctx).Raw("SELECT DATABASE()").Scan(&database).Error != nil || database != "element" || db.WithContext(ctx).Where("exam_id = ?", item.ExamID).Take(&profile).Error != nil || profile.FrozenAt == nil {
		t.Fatal("source profile gate")
	}
	rows, err := sharedSourceRows(ctx, db, item.RepoID)
	if err != nil {
		t.Fatal(err)
	}
	manifest, mapping, err := service.BuildManagementTraitsRuntimeCurrentSource(item.Code, rows)
	if err != nil || manifest.Versions.Question != "mng-00502-db-current-v2" {
		t.Fatal("current source version")
	}
	mc, err := service.CanonicalManagementTraitsManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	mapping.ManifestSHA = mc.SHA256
	pc, err := service.CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil || !strings.Contains(string(pc.JSON), managementTraits00502V67Title) || !strings.Contains(string(pc.JSON), managementTraits00502V96Title) {
		t.Fatal("approved wording missing from mapping")
	}
	field := sharedBaselineFieldContract{Schema: "mng-candidate-fields-v1", RequiredFields: []string{"name", "gender", "telephone"}, TimePolicy: "candidate-only-personal-25-minutes-v1", Source: "current-source-not-client-leader-full-not-historical", RepoCode: item.Code, CapturedAt: *profile.FrozenAt, ManifestSHA: mc.SHA256, MappingSHA: pc.SHA256}
	fieldRaw, _, err := sharedCanonical(field)
	if err != nil {
		t.Fatal(err)
	}
	bundleID := "m5b502-profile-v2-c51130cb775b"
	created := profile.FrozenAt.Format("2006-01-02 15:04:05.000000")
	sql := "-- Derived read-only from staging corrected 00502 source; preserves historical v1 paper snapshots.\n" +
		"SET NAMES utf8mb4;\nSTART TRANSACTION;\n" +
		fmt.Sprintf("INSERT INTO `el_mng_definition_bundle` (`id`,`product_version`,`question_version`,`scoring_version`,`norm_version`,`questionnaire`,`scoring_manifest`,`scoring_manifest_sha`,`status`,`created_at`) VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,'%s');\n", sqlUTF8Hex(bundleID), sqlUTF8Hex(manifest.Versions.Product), sqlUTF8Hex(manifest.Versions.Question), sqlUTF8Hex(manifest.Versions.Scoring), sqlUTF8Hex(manifest.Versions.Norm), sqlUTF8Hex(manifest.Questionnaire), sqlUTF8Hex(string(mc.JSON)), sqlUTF8Hex(mc.SHA256), sqlUTF8Hex("candidate-current-source"), created) +
		fmt.Sprintf("UPDATE `el_mng_exam_profile` SET `bundle_id`=%s,`mapping_snapshot`=%s,`field_contract`=%s,`mapping_sha`=%s WHERE `exam_id`=%s AND `bundle_id`=%s AND `mapping_sha`=%s;\n", sqlUTF8Hex(bundleID), sqlUTF8Hex(string(pc.JSON)), sqlUTF8Hex(string(fieldRaw)), sqlUTF8Hex(pc.SHA256), sqlUTF8Hex(item.ExamID), sqlUTF8Hex(profile.BundleID), sqlUTF8Hex(profile.MappingSHA)) +
		"SET @mng005_v2_ok=(ROW_COUNT()=1 AND (SELECT COUNT(*) FROM el_mng_exam_profile p JOIN el_mng_definition_bundle b ON b.id=p.bundle_id WHERE p.exam_id='9051103000000000502' AND b.question_version='mng-00502-db-current-v2')=1);\n" +
		"SET @mng005_sql=IF(@mng005_v2_ok,'SELECT ''MNG005_00502_PROFILE_V2_OK''','SELECT * FROM information_schema.MNG005_00502_PROFILE_V2_MISMATCH');\n" +
		"PREPARE mng005_stmt FROM @mng005_sql; EXECUTE mng005_stmt; DEALLOCATE PREPARE mng005_stmt;\nCOMMIT;\n"
	file, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := file.WriteString(sql)
	syncErr := file.Sync()
	closeErr := file.Close()
	if writeErr != nil || syncErr != nil || closeErr != nil {
		t.Fatal("profile export write")
	}
	fmt.Printf("MNG005_PROFILE_V2_EXPORT_PASS sha=%s bytes=%d\n", sharedSHA([]byte(sql)), len(sql))
}
