package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const sharedBaselineMarker = "MNG005BASE_20261008_c51130cb775ba019"
const managementTraits00502V67Title = "当我接手具有挑战性的工作时，我通常能鼓励大家创新，并提出创新的解决方案。"
const managementTraits00502V96Title = "当下属反对我的某个决定或者工作安排时，我会保持冷静和理性来应对。"

var sharedBaselineCases = []struct {
	Code, SourceCode, RepoID, Prefix, ExamID, Title, Name, Gender, Phone string
}{
	{"00501", "00201", "m5b501-c51130cb775b", "b501-", "9051103000000000501", "[TEST-保留] MNG005BASE_20261008_c51130cb775ba019 基层员工", "TEST00501-c51130cb", "0", "18851130001"},
	{"00502", "00202", "m5b502-c51130cb775b", "b502-", "9051103000000000502", "[TEST-保留] MNG005BASE_20261008_c51130cb775ba019 干部", "TEST00502-c51130cb", "1", "18851130002"},
}

type sharedBaselineFieldContract struct {
	Schema         string    `json:"schema"`
	RequiredFields []string  `json:"requiredFields"`
	TimePolicy     string    `json:"timePolicy"`
	Source         string    `json:"source"`
	RepoCode       string    `json:"repoCode"`
	CapturedAt     time.Time `json:"capturedAt"`
	ManifestSHA    string    `json:"manifestSha"`
	MappingSHA     string    `json:"mappingSha"`
}

type sharedBaselineRecord struct {
	Code               string            `json:"code"`
	ExamIDHash         string            `json:"examIdHash"`
	RunIDHash          string            `json:"runIdHash"`
	ReportIDHash       string            `json:"reportIdHash"`
	TitleHash          string            `json:"titleHash"`
	IdentityFieldsSHA  map[string]string `json:"identityFieldsSHA"`
	IdentityCommitment string            `json:"identityCommitment"`
	PDF                string            `json:"pdf"`
	PDFBytes           int64             `json:"pdfBytes"`
	PDFSHA256          string            `json:"pdfSHA256"`
	DataSHA256         string            `json:"dataSHA256"`
	Score              string            `json:"score"`
	Questions          int               `json:"questions"`
	Dimensions         int               `json:"dimensions"`
	Modules            int               `json:"modules"`
	Receipts           int               `json:"receipts"`
	Reports            int               `json:"reports"`
	Currents           int               `json:"currents"`
	Audits             int               `json:"audits"`
}

type sharedBaselineReceipt struct {
	Schema               string                 `json:"schema"`
	MarkerSHA256         string                 `json:"markerSHA256"`
	CreatedAt            string                 `json:"createdAt"`
	Execution            string                 `json:"execution"`
	SourceProvenance     map[string]string      `json:"sourceProvenance"`
	Records              []sharedBaselineRecord `json:"records"`
	ExistingReportCounts []int64                `json:"existingReportCounts"`
	FinalReportCounts    []int64                `json:"finalReportCounts"`
}

func sharedSHA(value []byte) string {
	h := sha256.Sum256(value)
	return hex.EncodeToString(h[:])
}

func sharedHashParts(parts ...string) string {
	return sharedSHA([]byte(strings.Join(parts, "\x00")))
}

func sharedBaselineIdentityCommitment(record sharedBaselineRecord) string {
	return sharedHashParts(
		"mng005-shared-baseline-identity-v1",
		record.Code,
		record.ExamIDHash,
		record.RunIDHash,
		record.ReportIDHash,
		record.TitleHash,
		record.PDFSHA256,
		record.DataSHA256,
		record.IdentityFieldsSHA["name"],
		record.IdentityFieldsSHA["gender"],
		record.IdentityFieldsSHA["phone"],
	)
}

func sharedCanonical(value any) ([]byte, string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, "", err
	}
	return raw, sharedSHA(raw), nil
}

func sharedSourceRows(ctx context.Context, db *gorm.DB, repoID string) ([]service.ManagementTraitsRuntimeSourceRow, error) {
	const query = `SELECT qr.id AS relation_id, q.id AS question_id, qr.sort AS sort,
qr.qu_type AS relation_type, q.qu_type AS question_type, q.content AS code, q.title AS title,
qa.id AS option_id, qa.score AS raw, qa.is_right AS is_right, qa.content AS option_content
FROM el_qu_repo qr LEFT JOIN el_qu q ON q.id = qr.qu_id
LEFT JOIN el_qu_answer qa ON qa.qu_id = q.id
WHERE qr.repo_id = ? ORDER BY qr.sort, qr.id, qa.score, qa.id LIMIT 701`
	rows := make([]service.ManagementTraitsRuntimeSourceRow, 0, 700)
	if err := db.WithContext(ctx).Raw(query, repoID).Scan(&rows).Error; err != nil || len(rows) != 700 {
		return nil, errors.New("source rows")
	}
	return rows, nil
}

func sharedWritePrivateJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(raw)
	syncErr := f.Sync()
	closeErr := f.Close()
	return errors.Join(writeErr, syncErr, closeErr)
}

func sharedCloneRepos(ctx context.Context, db *gorm.DB) error {
	repoTitles := map[string]string{
		"00501": "管理特质测验基层员工新版",
		"00502": "管理特质测验干部新版",
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range sharedBaselineCases {
			var occupied int64
			if err := tx.Raw("SELECT COUNT(*) FROM el_repo WHERE code = ? OR id = ?", item.Code, item.RepoID).Scan(&occupied).Error; err != nil || occupied != 0 {
				return errors.New("product occupied")
			}
			var source model.Repo
			if err := tx.Where("code = ?", item.SourceCode).Take(&source).Error; err != nil || source.RadioCount != 140 || source.MultiCount != 0 || source.JudgeCount != 0 {
				return errors.New("source repo")
			}
			now := time.Now()
			repo := model.Repo{ID: item.RepoID, Code: item.Code, Title: repoTitles[item.Code], RadioCount: 140, Remark: sharedBaselineMarker + ":cloned-from:" + item.SourceCode, CreateTime: &now, UpdateTime: &now}
			if err := tx.Create(&repo).Error; err != nil {
				return err
			}
			if err := tx.Exec(`INSERT INTO el_qu (id,qu_type,level,image,content,create_time,update_time,remark,analysis,title,question_code,dimension_id,dimension_item_no,observation_point,scoring_direction,competency_question_type,question_status)
SELECT CONCAT(?,q.id),q.qu_type,q.level,q.image,q.content,?, ?, CONCAT(?,':',COALESCE(q.remark,'')),q.analysis,
CASE WHEN ? = '00502' AND q.content = 'V67' THEN ?
	WHEN ? = '00502' AND q.content = 'V96' THEN ?
	ELSE q.title END,
q.question_code,q.dimension_id,q.dimension_item_no,q.observation_point,q.scoring_direction,q.competency_question_type,q.question_status
FROM el_qu q JOIN el_qu_repo qr ON qr.qu_id=q.id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code=?`, item.Prefix, now, now, sharedBaselineMarker, item.Code, managementTraits00502V67Title, item.Code, managementTraits00502V96Title, item.SourceCode).Error; err != nil {
				return err
			}
			if err := tx.Exec(`INSERT INTO el_qu_answer (id,qu_id,is_right,image,content,analysis,score)
SELECT CONCAT(?,a.id),CONCAT(?,a.qu_id),a.is_right,a.image,a.content,a.analysis,a.score
FROM el_qu_answer a JOIN el_qu_repo qr ON qr.qu_id=a.qu_id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code=?`, item.Prefix, item.Prefix, item.SourceCode).Error; err != nil {
				return err
			}
			if err := tx.Exec(`INSERT INTO el_qu_repo (id,qu_id,repo_id,qu_type,sort)
SELECT CONCAT(?,qr.id),CONCAT(?,qr.qu_id),?,qr.qu_type,qr.sort
FROM el_qu_repo qr JOIN el_repo r ON r.id=qr.repo_id WHERE r.code=?`, item.Prefix, item.Prefix, item.RepoID, item.SourceCode).Error; err != nil {
				return err
			}
			var repos, relations, questions, answers int64
			if err := tx.Raw(`SELECT COUNT(DISTINCT r.id),COUNT(DISTINCT qr.id),COUNT(DISTINCT q.id),COUNT(DISTINCT a.id)
FROM el_repo r LEFT JOIN el_qu_repo qr ON qr.repo_id=r.id LEFT JOIN el_qu q ON q.id=qr.qu_id LEFT JOIN el_qu_answer a ON a.qu_id=q.id WHERE r.id=?`, item.RepoID).Row().Scan(&repos, &relations, &questions, &answers); err != nil || repos != 1 || relations != 140 || questions != 140 || answers != 700 {
				return errors.New("clone counts")
			}
		}
		return nil
	})
}

func sharedCreateFrozenExam(ctx context.Context, db *gorm.DB, item struct {
	Code, SourceCode, RepoID, Prefix, ExamID, Title, Name, Gender, Phone string
}, sourceSHA string) error {
	rows, err := sharedSourceRows(ctx, db, item.RepoID)
	if err != nil {
		return err
	}
	manifest, mapping, err := service.BuildManagementTraitsRuntimeCurrentSource(item.Code, rows)
	if err != nil {
		return err
	}
	manifestCanonical, err := service.CanonicalManagementTraitsManifest(manifest)
	if err != nil {
		return err
	}
	mapping.ManifestSHA = manifestCanonical.SHA256
	mappingCanonical, err := service.CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil {
		return err
	}
	now := time.Now().Truncate(time.Second)
	field := sharedBaselineFieldContract{Schema: "mng-candidate-fields-v1", RequiredFields: []string{"name", "gender", "telephone"}, TimePolicy: "candidate-only-personal-25-minutes-v1", Source: "current-source-not-client-leader-full-not-historical", RepoCode: item.Code, CapturedAt: now, ManifestSHA: manifestCanonical.SHA256, MappingSHA: mappingCanonical.SHA256}
	fieldRaw, fieldSHA, err := sharedCanonical(field)
	if err != nil {
		return err
	}
	bundle := model.ManagementTraitsDefinitionBundle{ID: uuid.NewString(), ProductVersion: manifest.Versions.Product, QuestionVersion: manifest.Versions.Question, ScoringVersion: manifest.Versions.Scoring, NormVersion: manifest.Versions.Norm, Questionnaire: manifest.Questionnaire, ScoringManifest: string(manifestCanonical.JSON), ScoringManifestSHA: manifestCanonical.SHA256, Status: "candidate-current-source", CreatedAt: now}
	profile := model.ManagementTraitsExamProfile{ExamID: item.ExamID, BundleID: bundle.ID, MappingSnapshot: string(mappingCanonical.JSON), FieldContract: string(fieldRaw), MappingSHA: mappingCanonical.SHA256, TotalTimeMinutes: 25, FrozenAt: &now, CreatedAt: now}
	exam := model.Exam{ID: item.ExamID, Title: item.Title, Content: sharedBaselineMarker + ":TEST-only:source=" + item.SourceCode + ":sourceSHA=" + sourceSHA, OpenType: 1, JoinType: 1, IsOpen: 1, AnswerType: 1, Level: 1, State: 0, TimeLimit: 1, ShowPdf: 0, CreateTime: &now, UpdateTime: &now, TotalTime: 25, AssessmentType: "legacy", ScoringMode: "legacy", RequiredFields: "name,gender,telephone"}
	link := model.ExamRepo{ID: uuid.NewString(), ExamID: item.ExamID, RepoID: item.RepoID, RadioCount: 140}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&exam).Error; err != nil {
			return err
		}
		if err := tx.Create(&link).Error; err != nil {
			return err
		}
		if err := tx.Create(&bundle).Error; err != nil {
			return err
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}
		var stored model.ManagementTraitsExamProfile
		if err := tx.Where("exam_id = ?", item.ExamID).Take(&stored).Error; err != nil || stored.MappingSHA != mappingCanonical.SHA256 || sharedSHA([]byte(stored.FieldContract)) != fieldSHA {
			return errors.New("frozen profile")
		}
		return nil
	})
}

func sharedCleanupOwned(ctx context.Context, db *gorm.DB, reportRoot, artifactRoot string) error {
	if db == nil {
		return nil
	}
	examIDs := []string{sharedBaselineCases[0].ExamID, sharedBaselineCases[1].ExamID}
	var keys []string
	_ = db.WithContext(ctx).Raw("SELECT file_key FROM el_mng_report_revision WHERE exam_id IN ?", examIDs).Scan(&keys).Error
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"DELETE a FROM el_mng_report_audit a JOIN el_mng_report_revision r ON r.id=a.report_id WHERE r.exam_id IN ?",
			"DELETE c FROM el_mng_report_current c JOIN el_mng_report_revision r ON r.id=c.report_id WHERE r.exam_id IN ?",
			"DELETE FROM el_mng_report_revision WHERE exam_id IN ?",
			"DELETE FROM el_mng_runtime_receipt WHERE exam_id IN ?",
			"DELETE d FROM el_mng_result_dimension d JOIN el_mng_result_run r ON r.id=d.run_id WHERE r.exam_id IN ?",
			"DELETE m FROM el_mng_result_module m JOIN el_mng_result_run r ON r.id=m.run_id WHERE r.exam_id IN ?",
			"DELETE FROM el_mng_result_run WHERE exam_id IN ?",
			"DELETE q FROM el_mng_paper_question_snapshot q JOIN el_mng_paper_snapshot p ON p.paper_id=q.paper_id WHERE p.exam_id IN ?",
			"DELETE a FROM el_paper_qu_answer a JOIN el_paper p ON p.id=a.paper_id WHERE p.exam_id IN ?",
			"DELETE q FROM el_paper_qu q JOIN el_paper p ON p.id=q.paper_id WHERE p.exam_id IN ?",
			"DELETE FROM el_mng_paper_snapshot WHERE exam_id IN ?",
			"DELETE FROM el_candidate WHERE exam_id IN ?",
			"DELETE FROM el_paper WHERE exam_id IN ?",
			"DELETE FROM el_mng_exam_profile WHERE exam_id IN ?",
			"DELETE FROM el_exam_repo WHERE exam_id IN ?",
			"DELETE FROM el_exam WHERE id IN ?",
		}
		for _, statement := range statements {
			if err := tx.Exec(statement, examIDs).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec("DELETE FROM el_mng_definition_bundle WHERE product_version IN (?,?) AND id NOT IN (SELECT bundle_id FROM el_mng_exam_profile) AND id NOT IN (SELECT bundle_id FROM el_mng_paper_snapshot)", "mng-00501-v1", "mng-00502-v1").Error; err != nil {
			return err
		}
		for _, item := range sharedBaselineCases {
			if err := tx.Exec("DELETE a FROM el_qu_answer a JOIN el_qu_repo qr ON qr.qu_id=a.qu_id WHERE qr.repo_id=?", item.RepoID).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM el_qu_repo WHERE repo_id=?", item.RepoID).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE q FROM el_qu q LEFT JOIN el_qu_repo qr ON qr.qu_id=q.id WHERE qr.id IS NULL AND q.id IN (SELECT source_question_id FROM el_mng_paper_question_snapshot WHERE exam_id=?)", item.ExamID).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE q FROM el_qu q WHERE q.id NOT IN (SELECT qu_id FROM el_qu_repo) AND LEFT(q.id,?)=?", len(item.Prefix), item.Prefix).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM el_repo WHERE id=? AND code=? AND remark LIKE ?", item.RepoID, item.Code, sharedBaselineMarker+"%").Error; err != nil {
				return err
			}
		}
		return nil
	})
	for _, key := range keys {
		if regexp.MustCompile(`^[a-f0-9-]{36}\.pdf$`).MatchString(key) {
			_ = os.Remove(filepath.Join(reportRoot, key))
		}
	}
	_ = os.RemoveAll(artifactRoot)
	return err
}

func TestManagementTraitsSharedStagingProfileRefresh(t *testing.T) {
	if os.Getenv("MNG005_SHARED_PROFILE_REFRESH") != "1" {
		t.Skip("shared staging profile refresh is explicit opt-in")
	}
	hostname, err := os.Hostname()
	if err != nil || hostname != "vm-ubuntu-go-dev" || os.Getenv("REPORT_EFFECTIVE_ENV") != "staging" {
		t.Fatal("staging profile refresh gate")
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	item := sharedBaselineCases[1]
	var database string
	var active int64
	if db.WithContext(ctx).Raw("SELECT DATABASE()").Scan(&database).Error != nil || database != "element" || db.WithContext(ctx).Raw("SELECT COUNT(*) FROM el_paper WHERE exam_id=? AND state=1", item.ExamID).Scan(&active).Error != nil || active != 0 {
		t.Fatal("database or active-paper gate")
	}
	rows, err := sharedSourceRows(ctx, db, item.RepoID)
	if err != nil {
		t.Fatal(err)
	}
	manifest, mapping, err := service.BuildManagementTraitsRuntimeCurrentSource(item.Code, rows)
	if err != nil {
		t.Fatal(err)
	}
	mc, err := service.CanonicalManagementTraitsManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	mapping.ManifestSHA = mc.SHA256
	pc, err := service.CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil || !bytes.Contains(pc.JSON, []byte(managementTraits00502V67Title)) || !bytes.Contains(pc.JSON, []byte(managementTraits00502V96Title)) {
		t.Fatal("approved question text missing from canonical mapping")
	}
	var beforeEvidence string
	evidenceSQL := "SELECT CONCAT((SELECT GROUP_CONCAT(CONCAT(id,':',input_sha) ORDER BY id SEPARATOR '|') FROM el_mng_result_run WHERE exam_id=?),'#',(SELECT GROUP_CONCAT(CONCAT(id,':',data_sha,':',file_sha) ORDER BY id SEPARATOR '|') FROM el_mng_report_revision WHERE exam_id=?),'#',(SELECT COUNT(*) FROM el_mng_report_audit a JOIN el_mng_report_revision r ON r.id=a.report_id WHERE r.exam_id=?))"
	if db.WithContext(ctx).Raw(evidenceSQL, item.ExamID, item.ExamID, item.ExamID).Scan(&beforeEvidence).Error != nil || beforeEvidence == "" {
		t.Fatal("historical evidence gate")
	}
	var newBundleID string
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var profile model.ManagementTraitsExamProfile
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("exam_id = ?", item.ExamID).Take(&profile).Error != nil || profile.FrozenAt == nil || profile.FrozenAt.IsZero() || bytes.Contains([]byte(profile.MappingSnapshot), []byte(managementTraits00502V67Title)) {
			return errors.New("profile ownership or stale-state gate")
		}
		field := sharedBaselineFieldContract{Schema: "mng-candidate-fields-v1", RequiredFields: []string{"name", "gender", "telephone"}, TimePolicy: "candidate-only-personal-25-minutes-v1", Source: "current-source-not-client-leader-full-not-historical", RepoCode: item.Code, CapturedAt: *profile.FrozenAt, ManifestSHA: mc.SHA256, MappingSHA: pc.SHA256}
		fieldRaw, _, e := sharedCanonical(field)
		if e != nil {
			return e
		}
		bundle := model.ManagementTraitsDefinitionBundle{ID: uuid.NewString(), ProductVersion: manifest.Versions.Product, QuestionVersion: manifest.Versions.Question, ScoringVersion: manifest.Versions.Scoring, NormVersion: manifest.Versions.Norm, Questionnaire: manifest.Questionnaire, ScoringManifest: string(mc.JSON), ScoringManifestSHA: mc.SHA256, Status: "candidate-current-source", CreatedAt: *profile.FrozenAt}
		if tx.Create(&bundle).Error != nil {
			return errors.New("new bundle insert")
		}
		update := tx.Model(&model.ManagementTraitsExamProfile{}).Where("exam_id = ? AND bundle_id = ? AND mapping_sha = ?", item.ExamID, profile.BundleID, profile.MappingSHA).Updates(map[string]any{"bundle_id": bundle.ID, "mapping_snapshot": string(pc.JSON), "mapping_sha": pc.SHA256, "field_contract": string(fieldRaw)})
		if update.Error != nil || update.RowsAffected != 1 {
			return errors.New("profile compare-and-swap")
		}
		newBundleID = bundle.ID
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.NewManagementTraitsRuntimeService(db, cfg.Jwt.Secret, 1<<20).ProfileDetail(ctx, item.ExamID); err != nil {
		t.Fatal("refreshed profile validation", err)
	}
	var afterEvidence string
	if db.WithContext(ctx).Raw(evidenceSQL, item.ExamID, item.ExamID, item.ExamID).Scan(&afterEvidence).Error != nil || afterEvidence != beforeEvidence {
		t.Fatal("historical run/report evidence changed")
	}
	fmt.Printf("MNG005_PROFILE_REFRESH_PASS code=%s bundleHash=%s mappingSHA=%s\n", item.Code, sharedSHA([]byte(newBundleID)), pc.SHA256)
}

func TestManagementTraitsSharedStagingBaselineReceiptUpgrade(t *testing.T) {
	if os.Getenv("MNG005_SHARED_RECEIPT_UPGRADE") != "1" {
		t.Skip("shared staging baseline receipt upgrade is explicit opt-in")
	}
	hostname, err := os.Hostname()
	if err != nil || hostname != "vm-ubuntu-go-dev" || os.Getenv("REPORT_EFFECTIVE_ENV") != "staging" {
		t.Fatal("staging receipt upgrade gate")
	}
	artifactRoot := os.Getenv("MNG005_SHARED_ARTIFACT_ROOT")
	if artifactRoot != "/tmp/mng005-baseline-c51130cb775ba019" {
		t.Fatal("owned path gate")
	}
	info, err := os.Lstat(artifactRoot)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0700 {
		t.Fatal("owned directory gate")
	}
	raw, err := os.ReadFile(filepath.Join(artifactRoot, "safe-receipt.json"))
	if err != nil || sharedSHA(raw) != os.Getenv("MNG005_SHARED_RECEIPT_SHA256") {
		t.Fatal("safe receipt gate")
	}
	var receipt sharedBaselineReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil || receipt.Schema != "mng005-shared-baseline-core-v1" || len(receipt.Records) != 2 {
		t.Fatal("safe receipt schema gate")
	}
	hex64 := regexp.MustCompile(`^[a-f0-9]{64}$`)
	for index := range receipt.Records {
		record := &receipt.Records[index]
		values := []string{record.ExamIDHash, record.RunIDHash, record.ReportIDHash, record.TitleHash, record.PDFSHA256, record.DataSHA256, record.IdentityFieldsSHA["name"], record.IdentityFieldsSHA["gender"], record.IdentityFieldsSHA["phone"]}
		for _, value := range values {
			if !hex64.MatchString(value) {
				t.Fatal("safe receipt hash gate")
			}
		}
		record.IdentityCommitment = sharedBaselineIdentityCommitment(*record)
	}
	receipt.Schema = "mng005-shared-baseline-core-v2"
	if err := sharedWritePrivateJSON(filepath.Join(artifactRoot, "safe-receipt-v2.json"), receipt); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(receipt)
	fmt.Println(string(encoded))
}

func TestManagementTraitsSharedStagingBaseline(t *testing.T) {
	if os.Getenv("MNG005_SHARED_BASELINE") != "1" {
		t.Skip("shared staging baseline is explicit opt-in")
	}
	hostname, err := os.Hostname()
	if err != nil || hostname != "vm-ubuntu-go-dev" {
		t.Fatal("staging hostname gate")
	}
	if os.Getenv("REPORT_EFFECTIVE_ENV") != "staging" {
		t.Fatal("staging environment gate")
	}
	reportRoot := os.Getenv("MNG005_SHARED_REPORT_ROOT")
	artifactRoot := os.Getenv("MNG005_SHARED_ARTIFACT_ROOT")
	if reportRoot != "/opt/talent-assessment/private/management-traits-test-reports/baseline-c51130cb775ba019" || artifactRoot != "/tmp/mng005-baseline-c51130cb775ba019" {
		t.Fatal("owned path gate")
	}
	for _, root := range []string{reportRoot, artifactRoot} {
		if info, err := os.Lstat(root); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0700 {
			t.Fatal("owned directory gate")
		}
	}
	cfg := config.Load()
	if cfg.Mysql.DSN == "" {
		t.Fatal("database configuration")
	}
	db, err := gorm.Open(mysql.Open(cfg.Mysql.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("database connection")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	var database string
	if err := db.WithContext(ctx).Raw("SELECT DATABASE()").Scan(&database).Error; err != nil || database != "element" {
		t.Fatal("shared database gate")
	}
	var active, occupied, reissueTables, draftTables int64
	if err := db.WithContext(ctx).Raw("SELECT COUNT(*) FROM el_paper WHERE state=1").Scan(&active).Error; err != nil || active != 0 {
		t.Fatal("active assessment gate")
	}
	if err := db.WithContext(ctx).Raw("SELECT COUNT(*) FROM el_repo WHERE code IN (?,?) OR id IN (?,?)", "00501", "00502", sharedBaselineCases[0].RepoID, sharedBaselineCases[1].RepoID).Scan(&occupied).Error; err != nil || occupied != 0 {
		t.Fatal("005 ownership gate")
	}
	if err := db.WithContext(ctx).Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name IN ('el_mng_report_reissue','el_mng_reissue_audit')").Scan(&reissueTables).Error; err != nil || reissueTables != 0 {
		t.Fatal("reissue schema forbidden")
	}
	if err := db.WithContext(ctx).Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND (table_name='el_mng_exam_draft' OR table_name LIKE 'el_mng_formal%')").Scan(&draftTables).Error; err != nil || draftTables != 0 {
		t.Fatal("draft/formal schema forbidden")
	}
	var beforeRevision, beforeCurrent, beforeAudit int64
	if err := db.WithContext(ctx).Raw("SELECT (SELECT COUNT(*) FROM el_mng_report_revision),(SELECT COUNT(*) FROM el_mng_report_current),(SELECT COUNT(*) FROM el_mng_report_audit)").Row().Scan(&beforeRevision, &beforeCurrent, &beforeAudit); err != nil {
		t.Fatal(err)
	}
	sourceProvenance := make(map[string]string, 2)
	for _, item := range sharedBaselineCases {
		var repo model.Repo
		if err := db.WithContext(ctx).Where("code = ?", item.SourceCode).Take(&repo).Error; err != nil {
			t.Fatal("source repo")
		}
		rows, err := sharedSourceRows(ctx, db, repo.ID)
		if err != nil {
			t.Fatal(err)
		}
		_, sourceSHA, err := sharedCanonical(rows)
		if err != nil {
			t.Fatal(err)
		}
		sourceProvenance[item.Code] = sourceSHA
	}
	ownership := map[string]any{"schema": "mng005-shared-baseline-ownership-v1", "marker": sharedBaselineMarker, "status": "pending", "examIds": []string{sharedBaselineCases[0].ExamID, sharedBaselineCases[1].ExamID}, "repoIds": []string{sharedBaselineCases[0].RepoID, sharedBaselineCases[1].RepoID}, "sourceProvenance": sourceProvenance, "createdAt": time.Now().UTC().Format(time.RFC3339Nano)}
	if err := sharedWritePrivateJSON(filepath.Join(reportRoot, "ownership.private.json"), ownership); err != nil {
		t.Fatal(err)
	}
	keep := false
	defer func() {
		if !keep {
			_ = sharedCleanupOwned(context.Background(), db, reportRoot, artifactRoot)
		}
	}()
	if err := sharedCloneRepos(ctx, db); err != nil {
		t.Fatal(err)
	}
	for _, item := range sharedBaselineCases {
		if err := sharedCreateFrozenExam(ctx, db, item, sourceProvenance[item.Code]); err != nil {
			t.Fatal(err)
		}
	}
	serviceRuntime := service.NewManagementTraitsRuntimeService(db, cfg.Jwt.Secret, 1<<20)
	if err := serviceRuntime.CheckRuntimeSchema(ctx); err != nil {
		t.Fatal("runtime schema")
	}
	workbookPath := filepath.Join("configs", "export-templates", "management-traits-002-test-content-v1.xlsx")
	templatePath := filepath.Join("configs", "export-templates", "management-traits-002-test-only-v2.docx")
	workbook, err := os.ReadFile(workbookPath)
	if err != nil || sharedSHA(workbook) != "b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c" {
		t.Fatal("workbook asset")
	}
	content, err := service.LoadManagementTraitsTestContent(workbook)
	if err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile(templatePath)
	if err != nil || sharedSHA(template) != managementTraitsRuntimeTestTemplateSHA {
		t.Fatal("template asset")
	}
	var actor int64
	if err := db.WithContext(ctx).Raw("SELECT user_id FROM sys_user WHERE user_id=1 AND status='0' AND del_flag='0'").Scan(&actor).Error; err != nil || actor != 1 {
		t.Fatal("actor gate")
	}
	receipt := sharedBaselineReceipt{Schema: "mng005-shared-baseline-core-v1", MarkerSHA256: sharedSHA([]byte(sharedBaselineMarker)), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Execution: "one-shot-test-binary-no-listener-no-worker", SourceProvenance: sourceProvenance, Records: make([]sharedBaselineRecord, 0, 2), ExistingReportCounts: []int64{beforeRevision, beforeCurrent, beforeAudit}}
	for _, item := range sharedBaselineCases {
		identity, handled, err := serviceRuntime.TryRegisterCandidateIdentity(ctx, service.ManagementTraitsCandidateIdentityRequest{ExamID: item.ExamID, Name: item.Name, Gender: item.Gender, Telephone: item.Phone}, "")
		if err != nil || !handled || identity.ID == "" {
			t.Fatal("candidate registration")
		}
		claims := service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: identity.ID, ExamID: item.ExamID, ExpiresAt: time.Now().Add(30 * time.Minute).Unix()}
		detail, err := serviceRuntime.CreatePaper(ctx, claims)
		if err != nil || len(detail.Questions) != 140 {
			t.Fatal("paper creation")
		}
		paperClaims := claims
		paperClaims.Purpose = service.ManagementTraitsRuntimePaperPurpose
		paperClaims.PaperID = detail.PaperID
		for index, question := range detail.Questions {
			optionID := ""
			for _, option := range question.Options {
				if option.DisplayOrder == 3 && strings.TrimSpace(option.Content) == "一般" {
					optionID = option.ID
				}
			}
			if optionID == "" {
				t.Fatal("raw3 option")
			}
			answered, err := serviceRuntime.FillAnswer(ctx, paperClaims, question.ID, optionID)
			if err != nil || answered != index+1 {
				t.Fatal("answer persistence")
			}
		}
		submitted, err := serviceRuntime.SubmitParticipant(ctx, paperClaims, "manual")
		if err != nil || !submitted.Complete || submitted.Status != "completed" || submitted.Answered != 140 {
			t.Fatal("submission")
		}
		reportCtx, reportCancel := context.WithTimeout(ctx, 90*time.Second)
		report, err := serviceRuntime.GenerateTestReport(reportCtx, submitted.RunID, actor, reportRoot, content, func(renderCtx context.Context, dto service.ManagementTraitsTestReportData) ([]byte, error) {
			docx, err := renderManagementTraitsTestWord(template, dto)
			if err != nil {
				return nil, err
			}
			return libreofficepdf.NewClient("/usr/bin/libreoffice").Convert(renderCtx, "mng005-shared-TEST.docx", docx)
		})
		reportCancel()
		if err != nil {
			t.Fatal("report generation")
		}
		filePath := filepath.Join(reportRoot, report.FileKey)
		info, err := os.Lstat(filePath)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0600 || info.Size() != report.FileBytes {
			t.Fatal("report file mode")
		}
		pdf, err := os.ReadFile(filePath)
		if err != nil || !bytes.HasPrefix(pdf, []byte("%PDF-")) || sharedSHA(pdf) != report.FileSHA {
			t.Fatal("report file proof")
		}
		artifactName := item.Code + "-customer-template.pdf"
		if err := os.WriteFile(filepath.Join(artifactRoot, artifactName), pdf, 0600); err != nil {
			t.Fatal(err)
		}
		var questions, dims, modules, receipts, reports, currents, audits int
		var score, storedDataSHA, actualDataSHA string
		query := `SELECT
(SELECT COUNT(*) FROM el_mng_paper_question_snapshot q JOIN el_mng_paper_snapshot p ON p.paper_id=q.paper_id WHERE p.exam_id=? AND q.raw_answer=3),
(SELECT COUNT(*) FROM el_mng_result_dimension d JOIN el_mng_result_run r ON r.id=d.run_id WHERE r.exam_id=?),
(SELECT COUNT(*) FROM el_mng_result_module m JOIN el_mng_result_run r ON r.id=m.run_id WHERE r.exam_id=?),
(SELECT COUNT(*) FROM el_mng_runtime_receipt WHERE exam_id=?),
(SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id=?),
(SELECT COUNT(*) FROM el_mng_report_current c JOIN el_mng_report_revision r ON r.id=c.report_id WHERE r.exam_id=?),
(SELECT COUNT(*) FROM el_mng_report_audit a JOIN el_mng_report_revision r ON r.id=a.report_id WHERE r.exam_id=? AND a.action='generate'),
(SELECT CAST(overall_score AS CHAR) FROM el_mng_result_run WHERE exam_id=?),
(SELECT data_sha FROM el_mng_report_revision WHERE exam_id=?),
(SELECT SHA2(CAST(data_snapshot AS BINARY),256) FROM el_mng_report_revision WHERE exam_id=?)`
		if err := db.WithContext(ctx).Raw(query, item.ExamID, item.ExamID, item.ExamID, item.ExamID, item.ExamID, item.ExamID, item.ExamID, item.ExamID, item.ExamID, item.ExamID).Row().Scan(&questions, &dims, &modules, &receipts, &reports, &currents, &audits, &score, &storedDataSHA, &actualDataSHA); err != nil || questions != 140 || dims != 13 || modules != 4 || receipts != 1 || reports != 1 || currents != 1 || audits != 1 || score != "50.000000" || storedDataSHA != actualDataSHA || storedDataSHA != report.DataSHA {
			t.Fatal("persistent chain proof")
		}
		record := sharedBaselineRecord{Code: item.Code, ExamIDHash: sharedSHA([]byte(item.ExamID)), RunIDHash: sharedSHA([]byte(submitted.RunID)), ReportIDHash: sharedSHA([]byte(report.ID)), TitleHash: sharedSHA([]byte(item.Title)), IdentityFieldsSHA: map[string]string{"name": sharedHashParts("mng005-identity-field-v1", item.Code, "name", item.Name), "gender": sharedHashParts("mng005-identity-field-v1", item.Code, "gender", item.Gender), "phone": sharedHashParts("mng005-identity-field-v1", item.Code, "phone", item.Phone)}, PDF: artifactName, PDFBytes: report.FileBytes, PDFSHA256: report.FileSHA, DataSHA256: report.DataSHA, Score: score, Questions: questions, Dimensions: dims, Modules: modules, Receipts: receipts, Reports: reports, Currents: currents, Audits: audits}
		record.IdentityCommitment = sharedBaselineIdentityCommitment(record)
		receipt.Records = append(receipt.Records, record)
	}
	sort.Slice(receipt.Records, func(i, j int) bool { return receipt.Records[i].Code < receipt.Records[j].Code })
	for _, item := range sharedBaselineCases {
		var source model.Repo
		if err := db.WithContext(ctx).Where("code = ?", item.SourceCode).Take(&source).Error; err != nil {
			t.Fatal(err)
		}
		rows, err := sharedSourceRows(ctx, db, source.ID)
		if err != nil {
			t.Fatal(err)
		}
		_, afterSHA, err := sharedCanonical(rows)
		if err != nil || afterSHA != sourceProvenance[item.Code] {
			t.Fatal("002 source drift")
		}
	}
	var finalRevision, finalCurrent, finalAudit int64
	if err := db.WithContext(ctx).Raw("SELECT (SELECT COUNT(*) FROM el_mng_report_revision),(SELECT COUNT(*) FROM el_mng_report_current),(SELECT COUNT(*) FROM el_mng_report_audit)").Row().Scan(&finalRevision, &finalCurrent, &finalAudit); err != nil || finalRevision != beforeRevision+2 || finalCurrent != beforeCurrent+2 || finalAudit != beforeAudit+2 {
		t.Fatal("shared report delta")
	}
	receipt.FinalReportCounts = []int64{finalRevision, finalCurrent, finalAudit}
	if err := sharedWritePrivateJSON(filepath.Join(artifactRoot, "safe-receipt.json"), receipt); err != nil {
		t.Fatal(err)
	}
	ownership["status"] = "retained"
	ownership["completedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	ownership["recordHashes"] = receipt.Records
	ownershipPath := filepath.Join(reportRoot, "ownership-complete.private.json")
	if err := sharedWritePrivateJSON(ownershipPath, ownership); err != nil {
		t.Fatal(err)
	}
	keep = true
	encoded, _ := json.Marshal(receipt)
	fmt.Println(string(encoded))
}
