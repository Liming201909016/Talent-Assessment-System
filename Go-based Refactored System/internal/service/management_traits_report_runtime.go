package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ManagementTraitsTestPurposeTitle = "管理特质 TEST 测试报告"
const ManagementTraitsTestPurposeLabel = "仅供系统测试，不可作为人才决策依据"
const ManagementTraitsTestTemplateVersion = "mng-test-lo-template-v2"
const managementTraitsReportTemplateSHA = "05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c"

func validManagementTraitsReportTemplateSHA(value string) bool {
	if len(value) != sha256.Size*2 || value != strings.ToLower(value) {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func managementTraitsReportFile(root, key string) (string, error) {
	return managementTraitsReportFileDiagnostic(root, key, nil)
}

func managementTraitsReportFileDiagnostic(root, key string, diagnostic *managementTraitsReportDiagnostic) (string, error) {
	if !filepath.IsAbs(root) || filepath.Base(key) != key || strings.ContainsAny(key, `/\`) || !strings.HasSuffix(key, ".pdf") {
		return "", ErrManagementTraitsRuntimeInvalid
	}
	id := strings.TrimSuffix(key, ".pdf")
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrManagementTraitsRuntimeInvalid
	}
	f := filepath.Join(root, key)
	rel, err := filepath.Rel(root, f)
	if err != nil || rel != key {
		return "", ErrManagementTraitsRuntimeInvalid
	}
	// Refuse symlink/reparse-style directories and existing symlink files.
	for dir := filepath.Clean(root); ; dir = filepath.Dir(dir) {
		info, err := os.Lstat(dir)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", ErrManagementTraitsRuntimeInvalid
		}
		if err != nil && !os.IsNotExist(err) {
			diagnostic.failure(err)
			return "", ErrManagementTraitsRuntimeInvalid
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	if info, err := os.Lstat(f); err == nil && (!info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return "", ErrManagementTraitsRuntimeInvalid
	}
	return f, nil
}

func writeManagementTraitsTestPDF(root, id string, pdf []byte, diagnostic *managementTraitsReportDiagnostic) (string, error) {
	if len(pdf) < 1024 || len(pdf) > 50<<20 || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		diagnostic.failure(ErrManagementTraitsRuntimeInvalid)
		return "", ErrManagementTraitsRuntimeInvalid
	}
	key := id + ".pdf"
	path, err := managementTraitsReportFileDiagnostic(root, key, diagnostic)
	if err != nil {
		diagnostic.failure(err)
		return "", err
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		diagnostic.failure(err)
		return "", ErrManagementTraitsRuntimeInvalid
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		diagnostic.failure(err)
		return "", ErrManagementTraitsRuntimeInvalid
	}
	_, writeErr := f.Write(pdf)
	syncErr := f.Sync()
	closeErr := f.Close()
	if writeErr != nil || syncErr != nil || closeErr != nil {
		diagnostic.failure(errors.Join(writeErr, syncErr, closeErr))
		os.Remove(path)
		return "", ErrManagementTraitsRuntimeInvalid
	}
	return key, nil
}

func persistManagementTraitsTestReport(tx *gorm.DB, r model.ManagementTraitsReportRevision, a model.ManagementTraitsReportAudit, diagnostic *managementTraitsReportDiagnostic) error {
	if err := tx.Create(&r).Error; err != nil {
		diagnostic.failure(err)
		return ErrManagementTraitsRuntimeInvalid
	}
	current := model.ManagementTraitsReportCurrent{PaperID: r.PaperID, ReportID: r.ID, UpdatedAt: r.CreatedAt}
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "paper_id"}}, DoUpdates: clause.AssignmentColumns([]string{"report_id", "updated_at"})}).Create(&current).Error; err != nil {
		diagnostic.failure(err)
		return ErrManagementTraitsRuntimeInvalid
	}
	if err := tx.Create(&a).Error; err != nil {
		diagnostic.failure(err)
		return ErrManagementTraitsRuntimeInvalid
	}
	return nil
}

func (s *ManagementTraitsRuntimeService) loadTestReportData(ctx context.Context, tx *gorm.DB, runID string, content ManagementTraitsTestContent, diagnostic *managementTraitsReportDiagnostic) (ManagementTraitsTestReportData, error) {
	var run model.ManagementTraitsResultRun
	if !managementTraitsOpaqueID(runID) {
		diagnostic.failure(ErrManagementTraitsRuntimeInvalid)
		return ManagementTraitsTestReportData{}, ErrManagementTraitsRuntimeInvalid
	}
	if err := tx.WithContext(ctx).Where("id = ?", runID).Take(&run).Error; err != nil {
		diagnostic.failure(err)
		return ManagementTraitsTestReportData{}, ErrManagementTraitsRuntimeInvalid
	}
	if run.ID != runID {
		diagnostic.failure(ErrManagementTraitsRuntimeInvalid)
		return ManagementTraitsTestReportData{}, ErrManagementTraitsRuntimeInvalid
	}
	paper, snapshot, bundle, _, owner, questions, err := s.loadRuntimePaper(ctx, tx, run.PaperID)
	if err != nil || paper.State != 2 || run.Status != "completed" || run.SubmittedAt == nil || owner.EndTime == nil || !owner.EndTime.Equal(*run.SubmittedAt) || run.UserTimeSeconds == nil || paper.UserTime != *run.UserTimeSeconds {
		return ManagementTraitsTestReportData{}, ErrManagementTraitsRuntimeInvalid
	}
	var dims []model.ManagementTraitsResultDimension
	var modules []model.ManagementTraitsResultModule
	var receipts []model.ManagementTraitsRuntimeReceipt
	if err := tx.Where("run_id = ?", run.ID).Limit(14).Find(&dims).Error; err != nil {
		diagnostic.failure(err)
		return ManagementTraitsTestReportData{}, ErrManagementTraitsRuntimeInvalid
	}
	if err := tx.Where("run_id = ?", run.ID).Limit(5).Find(&modules).Error; err != nil {
		diagnostic.failure(err)
		return ManagementTraitsTestReportData{}, ErrManagementTraitsRuntimeInvalid
	}
	if err := tx.Where("run_id = ?", run.ID).Limit(2).Find(&receipts).Error; err != nil {
		diagnostic.failure(err)
		return ManagementTraitsTestReportData{}, ErrManagementTraitsRuntimeInvalid
	}
	if len(receipts) != 1 || validateManagementTraitsRuntimeBuckets(tx, paper.ID, questions, s.maxJSONBytes) != nil {
		return ManagementTraitsTestReportData{}, ErrManagementTraitsRuntimeInvalid
	}
	bundle, err = revalidateManagementTraitsWriteBundle(ctx, tx, bundle)
	if err != nil {
		return ManagementTraitsTestReportData{}, err
	}
	return BuildManagementTraitsTestReport(bundle, snapshot, questions, run, dims, modules, receipts[0], content, s.maxJSONBytes)
}

// One caller-owned deadline covers DB reads, rendering queue, LO, file write
// and commit. Rendering is outside the DB transaction. A second locked read
// compares the immutable DTO, including source trust, before creating a new
// revision. No legacy PDF, run, content or customer template is overwritten.
func (s *ManagementTraitsRuntimeService) GenerateTestReport(ctx context.Context, runID string, actor int64, root string, content ManagementTraitsTestContent, render func(context.Context, ManagementTraitsTestReportData) ([]byte, error)) (model.ManagementTraitsReportRevision, error) {
	return s.GenerateTestReportWithTemplateSHA(ctx, runID, actor, root, content, managementTraitsReportTemplateSHA, render)
}

func (s *ManagementTraitsRuntimeService) GenerateTestReportWithTemplateSHA(ctx context.Context, runID string, actor int64, root string, content ManagementTraitsTestContent, templateSHA string, render func(context.Context, ManagementTraitsTestReportData) ([]byte, error)) (model.ManagementTraitsReportRevision, error) {
	diagnostic := &managementTraitsReportDiagnostic{stage: "input"}
	defer diagnostic.emit()
	if actor <= 0 || render == nil || content.RuleCount() != 205 || !validManagementTraitsReportTemplateSHA(templateSHA) {
		diagnostic.failure(ErrManagementTraitsRuntimeInvalid)
		return model.ManagementTraitsReportRevision{}, ErrManagementTraitsRuntimeInvalid
	}
	diagnostic.stage = "schema"
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		diagnostic.failure(err)
		return model.ManagementTraitsReportRevision{}, err
	}
	var dto ManagementTraitsTestReportData
	diagnostic.stage = "load"
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		dto, err = s.loadTestReportData(ctx, tx, runID, content, diagnostic)
		return err
	})
	if err != nil {
		diagnostic.failure(err)
		return model.ManagementTraitsReportRevision{}, ErrManagementTraitsRuntimeInvalid
	}
	diagnostic.stage = "marshal"
	raw, err := json.Marshal(dto)
	if err != nil || len(raw) > s.maxJSONBytes {
		diagnostic.failure(err)
		return model.ManagementTraitsReportRevision{}, ErrManagementTraitsRuntimeInvalid
	}
	diagnostic.stage = "render"
	pdf, err := render(ctx, dto)
	if err != nil || ctx.Err() != nil {
		diagnostic.failure(errors.Join(err, ctx.Err()))
		return model.ManagementTraitsReportRevision{}, ErrManagementTraitsRuntimeInvalid
	}
	id := uuid.NewString()
	diagnostic.stage = "write"
	key, err := writeManagementTraitsTestPDF(root, id, pdf, diagnostic)
	if err != nil {
		return model.ManagementTraitsReportRevision{}, err
	}
	committed := false
	defer func() {
		if !committed {
			path, _ := managementTraitsReportFile(root, key)
			os.Remove(path)
		}
	}()
	dataSHA, fileSHA := sha256.Sum256(raw), sha256.Sum256(pdf)
	r := model.ManagementTraitsReportRevision{ID: id, RunID: runID, PaperID: dto.PaperID, ExamID: dto.ExamID, Mode: "test", TestTitle: ManagementTraitsTestPurposeTitle, TestLabel: ManagementTraitsTestPurposeLabel, ContentVersion: "mng-customer-text-test-v1", ContentSHA: content.SourceSHA(), TemplateVersion: ManagementTraitsTestTemplateVersion, TemplateSHA: templateSHA, DataSnapshot: string(raw), DataSHA: hex.EncodeToString(dataSHA[:]), FileKey: key, FileSHA: hex.EncodeToString(fileSHA[:]), FileBytes: int64(len(pdf)), CreatedBy: actor, CreatedAt: time.Now()}
	diagnostic.stage = "reload"
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		fresh, err := s.loadTestReportData(ctx, tx, runID, content, diagnostic)
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(fresh)
		if err != nil || !bytes.Equal(encoded, raw) {
			diagnostic.failure(err)
			return ErrManagementTraitsRuntimeInvalid
		}
		var seq struct {
			Value int `gorm:"column:revision"`
		}
		diagnostic.stage = "revision"
		if err := tx.Model(&model.ManagementTraitsReportRevision{}).Select("COALESCE(MAX(revision),0) AS revision").Where("run_id = ?", runID).Scan(&seq).Error; err != nil {
			diagnostic.failure(err)
			return ErrManagementTraitsRuntimeInvalid
		}
		if seq.Value < 0 || seq.Value >= 2147483646 {
			diagnostic.failure(ErrManagementTraitsRuntimeInvalid)
			return ErrManagementTraitsRuntimeInvalid
		}
		r.Revision = seq.Value + 1
		diagnostic.stage = "persist"
		return persistManagementTraitsTestReport(tx, r, model.ManagementTraitsReportAudit{ID: uuid.NewString(), ReportID: id, ActorID: actor, Action: "generate", CreatedAt: r.CreatedAt}, diagnostic)
	})
	if err != nil {
		diagnostic.failure(err)
		return model.ManagementTraitsReportRevision{}, ErrManagementTraitsRuntimeInvalid
	}
	committed = true
	return r, nil
}

func (s *ManagementTraitsRuntimeService) ReadTestReportPDF(ctx context.Context, reportID string, actor int64, root string, content ManagementTraitsTestContent) (model.ManagementTraitsReportRevision, []byte, error) {
	return s.ReadTestReportPDFWithTemplateSHAs(ctx, reportID, actor, root, content, map[string]bool{managementTraitsReportTemplateSHA: true})
}

func (s *ManagementTraitsRuntimeService) ReadTestReportPDFWithTemplateSHAs(ctx context.Context, reportID string, actor int64, root string, content ManagementTraitsTestContent, allowedTemplateSHAs map[string]bool) (model.ManagementTraitsReportRevision, []byte, error) {
	if actor <= 0 || !managementTraitsOpaqueID(reportID) {
		return model.ManagementTraitsReportRevision{}, nil, ErrManagementTraitsRuntimeInvalid
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return model.ManagementTraitsReportRevision{}, nil, err
	}
	var r model.ManagementTraitsReportRevision
	var pdf []byte
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Where("id = ?", reportID).Take(&r).Error != nil || r.ID != reportID || r.FileKey != r.ID+".pdf" || r.CreatedBy <= 0 || r.CreatedAt.IsZero() || r.Mode != "test" || r.Revision <= 0 || r.TestTitle != ManagementTraitsTestPurposeTitle || r.TestLabel != ManagementTraitsTestPurposeLabel || !validManagementTraitsReportTemplateSHA(r.TemplateSHA) || !allowedTemplateSHAs[r.TemplateSHA] || r.TemplateVersion != ManagementTraitsTestTemplateVersion || r.ContentVersion != "mng-customer-text-test-v1" || r.ContentSHA != content.SourceSHA() {
			return ErrManagementTraitsRuntimeInvalid
		}
		dto, err := s.loadTestReportData(ctx, tx, r.RunID, content, nil)
		if err != nil {
			return err
		}
		raw, err := json.Marshal(dto)
		if err != nil || string(raw) != r.DataSnapshot || dto.PaperID != r.PaperID || dto.ExamID != r.ExamID {
			return ErrManagementTraitsRuntimeInvalid
		}
		h := sha256.Sum256(raw)
		if hex.EncodeToString(h[:]) != r.DataSHA {
			return ErrManagementTraitsRuntimeInvalid
		}
		path, err := managementTraitsReportFile(root, r.FileKey)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() || info.Size() != r.FileBytes || info.Size() < 1024 || info.Size() > 50<<20 {
			return ErrManagementTraitsRuntimeInvalid
		}
		pdf, err = io.ReadAll(io.LimitReader(file, (50<<20)+1))
		if err != nil || int64(len(pdf)) != r.FileBytes || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
			return ErrManagementTraitsRuntimeInvalid
		}
		ph := sha256.Sum256(pdf)
		if hex.EncodeToString(ph[:]) != r.FileSHA {
			return ErrManagementTraitsRuntimeInvalid
		}
		if tx.Create(&model.ManagementTraitsReportAudit{ID: uuid.NewString(), ReportID: r.ID, ActorID: actor, Action: "download", CreatedAt: time.Now()}).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		return nil
	})
	if err != nil {
		return model.ManagementTraitsReportRevision{}, nil, errors.New("管理特质测试报告不可读取")
	}
	return r, pdf, nil
}
