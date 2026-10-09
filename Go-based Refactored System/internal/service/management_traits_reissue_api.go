package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

const managementTraitsFrozenReissueKind = "verified_frozen_result"

// Extend only this operation's guard budget AFTER its independent schema gate.
// Do not mutate the runtime singleton's mandatory eleven-table capabilities.
func (s *ManagementTraitsRuntimeService) reissueDB(ctx context.Context) *gorm.DB {
	columns := make(map[string]managementTraitsSchemaIDBudget)
	if prior := managementTraitsSchemaCapabilities(s.db); prior != nil {
		for k, v := range prior.Columns {
			columns[k] = v
		}
	}
	for _, key := range []string{"el_mng_report_reissue.id", "el_mng_report_reissue.run_id", "el_mng_report_reissue.paper_id", "el_mng_report_reissue.exam_id", "el_mng_reissue_audit.id", "el_mng_reissue_audit.report_id", "el_mng_reissue_audit.paper_id", "el_mng_reissue_audit.exam_id"} {
		columns[key] = managementTraitsSchemaIDBudget{Bytes: 64}
	}
	var local atomic.Pointer[managementTraitsRuntimeCapacity]
	local.Store(&managementTraitsRuntimeCapacity{Columns: columns, JSONBytes: s.maxJSONBytes})
	return s.db.Session(&gorm.Session{}).WithContext(ctx).Set("mng:schema_capacity", &local).Session(&gorm.Session{})
}

// The original JSON strings are preserved as values, never masked/rehashed as
// original identities. This private archive is not an HTTP response or log.
type managementTraitsFrozenReissue struct {
	Schema    string                                        `json:"schema"`
	Kind      string                                        `json:"kind"`
	Report    ManagementTraitsTestReportData                `json:"report"`
	Paper     model.ManagementTraitsPaperSnapshot           `json:"paper"`
	Bundle    model.ManagementTraitsDefinitionBundle        `json:"bundle"`
	Questions []model.ManagementTraitsPaperQuestionSnapshot `json:"questions"`
}

// Trust comes from the existing transactional server loader, not a signature
// supplied by a client. Historical unproven sources never reach this adapter.
func (s *ManagementTraitsRuntimeService) loadFrozenReissue(ctx context.Context, tx *gorm.DB, runID string, content ManagementTraitsTestContent) (managementTraitsFrozenReissue, []byte, error) {
	dto, err := s.loadTestReportData(ctx, tx, runID, content, nil)
	if err != nil {
		return managementTraitsFrozenReissue{}, nil, ErrManagementTraitsReissueInvalid
	}
	f := managementTraitsFrozenReissue{Schema: "mng-frozen-reissue-archive-v1", Kind: managementTraitsFrozenReissueKind, Report: dto, Questions: make([]model.ManagementTraitsPaperQuestionSnapshot, 0, 140)}
	if tx.Where("paper_id = ?", dto.PaperID).Take(&f.Paper).Error != nil || tx.Where("id = ?", f.Paper.BundleID).Take(&f.Bundle).Error != nil || tx.Where("paper_id = ?", dto.PaperID).Order("v_number").Limit(141).Find(&f.Questions).Error != nil || len(f.Questions) != 140 {
		return f, nil, ErrManagementTraitsReissueInvalid
	}
	raw, err := json.Marshal(f)
	if err != nil || len(raw) > s.maxJSONBytes {
		return f, nil, ErrManagementTraitsReissueInvalid
	}
	return f, raw, nil
}

type ManagementTraitsReissueQualification struct {
	Eligible bool   `json:"eligible"`
	Kind     string `json:"kind"`
	Purpose  string `json:"purpose"`
}

func (s *ManagementTraitsRuntimeService) QualifyReportReissue(ctx context.Context, runID string, content ManagementTraitsTestContent) (ManagementTraitsReissueQualification, error) {
	if s == nil || !managementTraitsOpaqueID(runID) || content.RuleCount() != 205 {
		return ManagementTraitsReissueQualification{}, ErrManagementTraitsReissueInvalid
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return ManagementTraitsReissueQualification{}, err
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { _, _, err := s.loadFrozenReissue(ctx, tx, runID, content); return err })
	if err != nil {
		return ManagementTraitsReissueQualification{}, ErrManagementTraitsReissueInvalid
	}
	return ManagementTraitsReissueQualification{true, managementTraitsFrozenReissueKind, ManagementTraitsTestPurposeLabel}, nil
}

func (s *ManagementTraitsRuntimeService) ListReportReissues(ctx context.Context, paperID string) ([]model.ManagementTraitsReportReissue, error) {
	rows := make([]model.ManagementTraitsReportReissue, 0)
	if !managementTraitsOpaqueID(paperID) {
		return rows, ErrManagementTraitsReissueInvalid
	}
	if err := s.checkReissueSchema(ctx); err != nil {
		return rows, err
	}
	// Do not retrieve identity/raw-answer archives for a metadata list.
	err := s.reissueDB(ctx).Select("id,run_id,paper_id,exam_id,kind,status,template_sha,content_sha,file_sha,file_bytes,created_by,created_at").Where("paper_id = ?", paperID).Order("created_at DESC,id DESC").Limit(100).Find(&rows).Error
	if err != nil {
		return make([]model.ManagementTraitsReportReissue, 0), ErrManagementTraitsReissueInvalid
	}
	for _, r := range rows {
		if r.PaperID != paperID || r.Kind != managementTraitsFrozenReissueKind || r.Status != "completed" || r.CreatedBy <= 0 || r.CreatedAt.IsZero() {
			return make([]model.ManagementTraitsReportReissue, 0), ErrManagementTraitsReissueInvalid
		}
	}
	return rows, nil
}

func managementTraitsFindReissue(tx *gorm.DB, runID, dataSHA, templateSHA, contentSHA string) (*model.ManagementTraitsReportReissue, error) {
	rows := make([]model.ManagementTraitsReportReissue, 0, 1)
	if tx.Where("run_id = ? AND data_sha = ? AND template_sha = ? AND content_sha = ?", runID, dataSHA, templateSHA, contentSHA).Limit(2).Find(&rows).Error != nil || len(rows) > 1 {
		return nil, ErrManagementTraitsReissueInvalid
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func managementTraitsReissueAudit(tx *gorm.DB, r model.ManagementTraitsReportReissue, actor int64, action string) error {
	a := model.ManagementTraitsReissueAudit{ID: uuid.NewString(), ReportID: r.ID, PaperID: r.PaperID, ExamID: r.ExamID, ActorID: actor, Action: action, CreatedAt: time.Now()}
	if tx.Create(&a).Error != nil {
		return ErrManagementTraitsReissueInvalid
	}
	return nil
}

// Only the report INSERT's exact input constraint is recoverable. Neither
// unrelated 1062 errors nor translated/string errors prove an input winner.
func managementTraitsReissueInputDuplicate(err error) bool {
	var e *driver.MySQLError
	return errors.As(err, &e) && e.Number == 1062 && (strings.HasSuffix(e.Message, " for key 'el_mng_report_reissue.uk_mng_reissue_input'") || strings.HasSuffix(e.Message, " for key 'uk_mng_reissue_input'"))
}

func (s *ManagementTraitsRuntimeService) validateReissueWinner(tx *gorm.DB, r model.ManagementTraitsReportReissue, frozen managementTraitsFrozenReissue, raw []byte, templateSHA, root string) error {
	if !managementTraitsOpaqueID(r.ID) || r.RunID != frozen.Report.RunID || r.PaperID != frozen.Report.PaperID || r.ExamID != frozen.Report.ExamID || r.DataSHA != managementTraitsReissueSHA(raw) || !bytes.Equal([]byte(r.DataSnapshot), raw) || r.TemplateSHA != templateSHA || r.ContentSHA != frozen.Report.ContentSourceSHA {
		return ErrManagementTraitsReissueInvalid
	}
	if _, err := s.readReissueFile(r, root, map[string]bool{templateSHA: true}); err != nil {
		return err
	}
	audits := make([]model.ManagementTraitsReissueAudit, 0, 1)
	if tx.Where("report_id = ? AND paper_id = ? AND exam_id = ? AND action = ?", r.ID, r.PaperID, r.ExamID, "generate").Limit(2).Find(&audits).Error != nil || len(audits) != 1 {
		return ErrManagementTraitsReissueInvalid
	}
	a := audits[0]
	if !managementTraitsOpaqueID(a.ID) || a.ReportID != r.ID || a.PaperID != r.PaperID || a.ExamID != r.ExamID || a.Action != "generate" || a.ActorID != r.CreatedBy || a.CreatedAt.IsZero() {
		return ErrManagementTraitsReissueInvalid
	}
	return nil
}

func (s *ManagementTraitsRuntimeService) GenerateReportReissue(ctx context.Context, runID string, actor int64, root string, content ManagementTraitsTestContent, render func(context.Context, ManagementTraitsTestReportData) ([]byte, error)) (model.ManagementTraitsReportReissue, bool, error) {
	return s.GenerateReportReissueWithTemplateSHA(ctx, runID, actor, root, content, managementTraitsReportTemplateSHA, render)
}

func (s *ManagementTraitsRuntimeService) GenerateReportReissueWithTemplateSHA(ctx context.Context, runID string, actor int64, root string, content ManagementTraitsTestContent, templateSHA string, render func(context.Context, ManagementTraitsTestReportData) ([]byte, error)) (model.ManagementTraitsReportReissue, bool, error) {
	if s == nil || actor <= 0 || !managementTraitsOpaqueID(runID) || render == nil || !filepath.IsAbs(root) || content.RuleCount() != 205 || !validManagementTraitsReportTemplateSHA(templateSHA) {
		return model.ManagementTraitsReportReissue{}, false, ErrManagementTraitsReissueInvalid
	}
	if err := s.checkReissueSchema(ctx); err != nil {
		return model.ManagementTraitsReportReissue{}, false, err
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return model.ManagementTraitsReportReissue{}, false, err
	}
	var frozen managementTraitsFrozenReissue
	var raw []byte
	var existing *model.ManagementTraitsReportReissue
	err := s.reissueDB(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		frozen, raw, err = s.loadFrozenReissue(ctx, tx, runID, content)
		if err != nil {
			return err
		}
		existing, err = managementTraitsFindReissue(tx, runID, managementTraitsReissueSHA(raw), templateSHA, content.SourceSHA())
		if err != nil {
			return err
		}
		if existing != nil {
			if err = s.validateReissueWinner(tx, *existing, frozen, raw, templateSHA, root); err != nil {
				return err
			}
			return managementTraitsReissueAudit(tx, *existing, actor, "reuse")
		}
		return nil
	})
	if err != nil {
		return model.ManagementTraitsReportReissue{}, false, ErrManagementTraitsReissueInvalid
	}
	if existing != nil {
		return *existing, true, nil
	}
	// No source locks or transaction are held across the bounded renderer/LO.
	pdf, err := render(ctx, frozen.Report)
	if err != nil || ctx.Err() != nil {
		return model.ManagementTraitsReportReissue{}, false, ErrManagementTraitsReissueInvalid
	}
	id := uuid.NewString()
	key, err := writeManagementTraitsTestPDF(root, id, pdf, nil)
	if err != nil {
		return model.ManagementTraitsReportReissue{}, false, ErrManagementTraitsReissueInvalid
	}
	committed := false
	defer func() {
		if !committed {
			if p, e := managementTraitsReportFile(root, key); e == nil {
				_ = os.Remove(p)
			}
		}
	}()
	r := model.ManagementTraitsReportReissue{ID: id, RunID: runID, PaperID: frozen.Report.PaperID, ExamID: frozen.Report.ExamID, Kind: managementTraitsFrozenReissueKind, Status: "completed", DataSnapshot: string(raw), DataSHA: managementTraitsReissueSHA(raw), TemplateSHA: templateSHA, ContentSHA: content.SourceSHA(), FileKey: key, FileSHA: managementTraitsReissueSHA(pdf), FileBytes: int64(len(pdf)), CreatedBy: actor, CreatedAt: time.Now()}
	reused := false
	duplicateInput := false
	err = s.reissueDB(ctx).Transaction(func(tx *gorm.DB) error {
		_, fresh, err := s.loadFrozenReissue(ctx, tx, runID, content)
		if err != nil || !bytes.Equal(fresh, raw) {
			return ErrManagementTraitsReissueInvalid
		}
		existing, err = managementTraitsFindReissue(tx, runID, r.DataSHA, templateSHA, r.ContentSHA)
		if err != nil {
			return err
		}
		if existing != nil {
			if err = s.validateReissueWinner(tx, *existing, frozen, raw, templateSHA, root); err != nil {
				return err
			}
			r = *existing
			reused = true
			return managementTraitsReissueAudit(tx, r, actor, "reuse")
		}
		if err := tx.Create(&r).Error; err != nil {
			duplicateInput = managementTraitsReissueInputDuplicate(err)
			return err
		}
		return managementTraitsReissueAudit(tx, r, actor, "generate")
	})
	// A run read before waiting for paper UPDATE establishes a repeatable-read
	// view. The serialized loser can still miss the committed report. End that
	// transaction completely before a single fresh, reuse-only transaction.
	if err != nil && duplicateInput && ctx.Err() == nil {
		err = s.reissueDB(ctx).Transaction(func(tx *gorm.DB) error {
			_, fresh, e := s.loadFrozenReissue(ctx, tx, runID, content)
			if e != nil || !bytes.Equal(fresh, raw) {
				return ErrManagementTraitsReissueInvalid
			}
			winner, e := managementTraitsFindReissue(tx, runID, managementTraitsReissueSHA(raw), templateSHA, content.SourceSHA())
			if e != nil || winner == nil {
				return ErrManagementTraitsReissueInvalid
			}
			if e = s.validateReissueWinner(tx, *winner, frozen, raw, templateSHA, root); e != nil {
				return e
			}
			if e = managementTraitsReissueAudit(tx, *winner, actor, "reuse"); e != nil {
				return e
			}
			r = *winner
			reused = true
			return nil
		})
	}
	if err != nil {
		return model.ManagementTraitsReportReissue{}, false, ErrManagementTraitsReissueInvalid
	}
	committed = !reused
	return r, reused, nil
}

func (s *ManagementTraitsRuntimeService) readReissueFile(r model.ManagementTraitsReportReissue, root string, allowedTemplateSHAs ...map[string]bool) ([]byte, error) {
	fail := ErrManagementTraitsReissueInvalid
	allowed := r.TemplateSHA == managementTraitsReportTemplateSHA
	if len(allowedTemplateSHAs) == 1 {
		allowed = allowedTemplateSHAs[0][r.TemplateSHA]
	}
	if r.FileKey != r.ID+".pdf" || r.Kind != managementTraitsFrozenReissueKind || r.Status != "completed" || r.CreatedBy <= 0 || r.CreatedAt.IsZero() || !validManagementTraitsReportTemplateSHA(r.TemplateSHA) || !allowed || r.FileBytes < 1024 || r.FileBytes > 50<<20 || len(r.DataSnapshot) == 0 || len(r.DataSnapshot) > s.maxJSONBytes || managementTraitsReissueSHA([]byte(r.DataSnapshot)) != r.DataSHA {
		return nil, fail
	}
	var frozen managementTraitsFrozenReissue
	if json.Unmarshal([]byte(r.DataSnapshot), &frozen) != nil || frozen.Schema != "mng-frozen-reissue-archive-v1" || frozen.Kind != r.Kind || !frozen.Report.TestOnly || frozen.Report.RunID != r.RunID || frozen.Report.PaperID != r.PaperID || frozen.Report.ExamID != r.ExamID || frozen.Report.ContentSourceSHA != r.ContentSHA || frozen.Paper.PaperID != r.PaperID || frozen.Paper.ExamID != r.ExamID || len(frozen.Questions) != 140 {
		return nil, fail
	}
	path, err := managementTraitsReportFile(root, r.FileKey)
	if err != nil {
		return nil, fail
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fail
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() != r.FileBytes {
		return nil, fail
	}
	pdf, err := io.ReadAll(io.LimitReader(f, (50<<20)+1))
	if err != nil || int64(len(pdf)) != r.FileBytes || !bytes.HasPrefix(pdf, []byte("%PDF-")) || managementTraitsReissueSHA(pdf) != r.FileSHA {
		return nil, fail
	}
	return pdf, nil
}

// Historical rendered output is read from its immutable archive. Revocation
// blocks new generation, not administrator reading of a previously saved PDF.
func (s *ManagementTraitsRuntimeService) ReadReportReissue(ctx context.Context, id string, actor int64, root string, action string) (model.ManagementTraitsReportReissue, []byte, error) {
	return s.ReadReportReissueWithTemplateSHAs(ctx, id, actor, root, action, map[string]bool{managementTraitsReportTemplateSHA: true})
}

func (s *ManagementTraitsRuntimeService) ReadReportReissueWithTemplateSHAs(ctx context.Context, id string, actor int64, root string, action string, allowedTemplateSHAs map[string]bool) (model.ManagementTraitsReportReissue, []byte, error) {
	if actor <= 0 || !managementTraitsOpaqueID(id) || (action != "view" && action != "download") {
		return model.ManagementTraitsReportReissue{}, nil, ErrManagementTraitsReissueInvalid
	}
	if err := s.checkReissueSchema(ctx); err != nil {
		return model.ManagementTraitsReportReissue{}, nil, err
	}
	var r model.ManagementTraitsReportReissue
	var pdf []byte
	err := s.reissueDB(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Where("id = ?", id).Take(&r).Error != nil || r.ID != id {
			return ErrManagementTraitsReissueInvalid
		}
		var err error
		pdf, err = s.readReissueFile(r, root, allowedTemplateSHAs)
		if err != nil {
			return err
		}
		return managementTraitsReissueAudit(tx, r, actor, action)
	})
	if err != nil {
		return model.ManagementTraitsReportReissue{}, nil, ErrManagementTraitsReissueInvalid
	}
	return r, pdf, nil
}
