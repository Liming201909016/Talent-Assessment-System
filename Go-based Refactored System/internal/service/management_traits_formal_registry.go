package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrManagementTraitsFormalClosed   = errors.New("formal registry unavailable")
	ErrManagementTraitsFormalInvalid  = errors.New("formal registry rejected")
	ErrManagementTraitsFormalConflict = errors.New("formal registry version conflict")
	formalKey                         = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,47}$`)
)

type ManagementTraitsFormalRegistry struct {
	db                     *gorm.DB
	environment, assetRoot string
}

// Independent local-only capability; this neither enables PDF generation nor
// changes TEST configuration. Paths are server configuration, never API input.
func NewManagementTraitsFormalRegistry(db *gorm.DB) *ManagementTraitsFormalRegistry {
	return &ManagementTraitsFormalRegistry{db: db, environment: os.Getenv("MNG_FORMAL_REGISTRY_ENV"), assetRoot: os.Getenv("MNG_FORMAL_ASSET_DIR")}
}

type ManagementTraitsFormalReadiness struct {
	Version                           model.ManagementTraitsFormalVersion    `json:"version"`
	Approvals                         []model.ManagementTraitsFormalApproval `json:"approvals"`
	CanApprove                        bool                                   `json:"canApprove"`
	CanActivate                       bool                                   `json:"canActivate"`
	VersionReady                      bool                                   `json:"versionReady"`
	CanGenerate                       bool                                   `json:"canGenerate"`
	BlockedReason                     string                                 `json:"blockedReason"`
	GenerationBlockedReason           string                                 `json:"generationBlockedReason"`
	RevokedHistoricalAdminReadAllowed bool                                   `json:"revokedHistoricalAdminReadAllowed"`
}

type ManagementTraitsFormalAssetPreview struct {
	RepoCode             string   `json:"repoCode"`
	AssetKey             string   `json:"assetKey"`
	ContentPresent       bool     `json:"contentPresent"`
	TemplatePresent      bool     `json:"templatePresent"`
	ContentRuleCount     int      `json:"contentRuleCount"`
	WorkbookSHA          string   `json:"workbookSha"`
	NormalizedContentSHA string   `json:"normalizedContentSha"`
	TemplateSHA          string   `json:"templateSha"`
	BindingSHA           string   `json:"bindingSha"`
	TagCount             int      `json:"tagCount"`
	ChartCount           int      `json:"chartCount"`
	NumericLabelCount    int      `json:"numericLabelCount"`
	Tags                 []string `json:"tags"`
	Charts               []string `json:"charts"`
	NumericLabels        []string `json:"numericLabels"`
	ReadyForRegistration bool     `json:"readyForRegistration"`
	ReadyForApproval     bool     `json:"readyForApproval"`
	BlockedReasons       []string `json:"blockedReasons"`
}

func newManagementTraitsFormalAssetPreview(repoCode, assetKey string) ManagementTraitsFormalAssetPreview {
	return ManagementTraitsFormalAssetPreview{RepoCode: repoCode, AssetKey: assetKey, Tags: make([]string, 0), Charts: make([]string, 0), NumericLabels: make([]string, 0), BlockedReasons: make([]string, 0)}
}

// PreviewAssets validates controlled candidate files without reading or writing
// registry tables. A preview is evidence for review only; Register and Change
// always re-read the files and enforce the full source/version gates.
func (s *ManagementTraitsFormalRegistry) PreviewAssets(ctx context.Context, repoCode, assetKey string) (ManagementTraitsFormalAssetPreview, error) {
	out := newManagementTraitsFormalAssetPreview(repoCode, assetKey)
	if repoCode != "00501" && repoCode != "00502" || !formalKey.MatchString(assetKey) {
		return out, ErrManagementTraitsFormalInvalid
	}
	if s == nil || ctx == nil || ctx.Err() != nil || s.environment != "local" || !filepath.IsAbs(s.assetRoot) {
		return out, ErrManagementTraitsFormalClosed
	}
	content, err := formalReadAsset(s.assetRoot, assetKey, "content.xlsx")
	if err != nil {
		reason := formalAssetReason(err, "content_read_failed")
		if os.IsNotExist(err) {
			reason = "content_missing"
		}
		out.BlockedReasons = append(out.BlockedReasons, reason)
		return out, nil
	}
	out.ContentPresent = true
	out.NormalizedContentSHA, out.WorkbookSHA, out.ContentRuleCount, err = managementTraitsFormalContentIdentity(content)
	if err != nil {
		out.BlockedReasons = append(out.BlockedReasons, formalAssetReason(err, "content_contract_invalid"))
		return out, nil
	}
	out.ReadyForRegistration = true
	template, err := formalReadAsset(s.assetRoot, assetKey, "template.docx")
	if err != nil {
		reason := formalAssetReason(err, "template_read_failed")
		if os.IsNotExist(err) {
			reason = "template_missing"
		}
		out.BlockedReasons = append(out.BlockedReasons, reason)
		return out, nil
	}
	out.TemplatePresent = true
	binding, err := validateManagementTraitsFormalTemplateBinding(template)
	if err != nil {
		out.BlockedReasons = append(out.BlockedReasons, formalAssetReason(err, "template_contract_invalid"))
		return out, nil
	}
	out.TemplateSHA = formalSHA(template)
	out.BindingSHA = binding.SHA
	out.Tags = append(out.Tags, binding.Tags...)
	out.Charts = append(out.Charts, binding.Charts...)
	out.NumericLabels = append(out.NumericLabels, binding.NumericLabels...)
	out.TagCount, out.ChartCount, out.NumericLabelCount = len(out.Tags), len(out.Charts), len(out.NumericLabels)
	out.ReadyForApproval = true
	return out, nil
}

func formalReadiness(v model.ManagementTraitsFormalVersion, approvals []model.ManagementTraitsFormalApproval, assets bool) ManagementTraitsFormalReadiness {
	r := ManagementTraitsFormalReadiness{Version: v, Approvals: append(make([]model.ManagementTraitsFormalApproval, 0, len(approvals)), approvals...), GenerationBlockedReason: "formal_report_pipeline_not_installed", RevokedHistoricalAdminReadAllowed: v.State == "revoked"}
	if v.Environment != "local" {
		r.BlockedReason = "environment_closed"
		return r
	}
	if v.State == "revoked" {
		r.BlockedReason = "version_revoked"
		return r
	}
	if v.State != "draft" && v.State != "active" || v.Epoch < 1 || v.IdentitySHA != formalVersionIdentity(v) {
		r.BlockedReason = "version_invalid"
		return r
	}
	if !assets {
		r.BlockedReason = "formal_assets_not_ready"
		return r
	}
	seen := make(map[string]bool, 2)
	for _, a := range approvals {
		if a.VersionID != v.ID || a.IdentitySHA != v.IdentitySHA || a.ActorID <= 0 || a.CreatedAt.IsZero() || a.CreatedAt.Before(v.CreatedAt) || (a.Kind != "content" && a.Kind != "psychometrics") || seen[a.Kind] {
			r.BlockedReason = "approval_invalid"
			return r
		}
		seen[a.Kind] = true
	}
	r.CanApprove = v.State == "draft"
	if !seen["content"] || !seen["psychometrics"] {
		r.BlockedReason = "dual_approval_required"
		return r
	}
	r.CanActivate = v.State == "draft"
	if v.State != "active" {
		r.BlockedReason = "activation_required"
		return r
	}
	r.VersionReady = true
	return r
}

func formalSHA(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func formalJSON(value any) string { b, _ := json.Marshal(value); return string(b) }
func formalVersionIdentity(v model.ManagementTraitsFormalVersion) string {
	return formalSHA([]byte(formalJSON(struct{ Domain, Code, Env, Repo, Bundle, Source, Content, Workbook, Template, Binding, Asset string }{"mng-formal-version-v1", v.VersionCode, v.Environment, v.RepoCode, v.BundleID, v.SourceSnapshot, v.ContentSHA, v.WorkbookSHA, v.TemplateSHA, v.BindingSHA, v.AssetKey})))
}

type formalSourceWire struct {
	Versions      ManagementTraitsScoringVersions `json:"versions"`
	Questionnaire string                          `json:"questionnaire"`
	Manifest      string                          `json:"manifest"`
	ManifestSHA   string                          `json:"manifestSha"`
	Mapping       string                          `json:"mapping"`
	MappingSHA    string                          `json:"mappingSha"`
}

func formalSourceSnapshot(b model.ManagementTraitsDefinitionBundle, p model.ManagementTraitsExamProfile) string {
	return formalJSON(formalSourceWire{ManagementTraitsScoringVersions{Product: b.ProductVersion, Question: b.QuestionVersion, Scoring: b.ScoringVersion, Norm: b.NormVersion}, b.Questionnaire, b.ScoringManifest, b.ScoringManifestSHA, p.MappingSnapshot, p.MappingSHA})
}

func (s *ManagementTraitsFormalRegistry) gate(ctx context.Context) error {
	if s == nil || s.db == nil || s.environment != "local" || ctx == nil || ctx.Err() != nil {
		return ErrManagementTraitsFormalClosed
	}
	return checkManagementTraitsFormalSchema(ctx, s.db)
}

func (s *ManagementTraitsFormalRegistry) assetsMatch(v model.ManagementTraitsFormalVersion) bool {
	a, err := loadManagementTraitsFormalAssets(s.assetRoot, v.AssetKey)
	return err == nil && a.templateSHA != "" && a.contentSHA == v.ContentSHA && a.workbookSHA == v.WorkbookSHA && a.templateSHA == v.TemplateSHA && a.bindingSHA == v.BindingSHA
}

func (s *ManagementTraitsFormalRegistry) sourceMatch(tx *gorm.DB, v model.ManagementTraitsFormalVersion) bool {
	var wire formalSourceWire
	if managementTraitsDecodeStrict([]byte(v.SourceSnapshot), 1<<20, &wire) != nil {
		return false
	}
	var b model.ManagementTraitsDefinitionBundle
	if tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", v.BundleID).Take(&b).Error != nil || !managementTraitsRuntimeSourceTrusted(b.Status) {
		return false
	}
	if b.ProductVersion != wire.Versions.Product || b.QuestionVersion != wire.Versions.Question || b.ScoringVersion != wire.Versions.Scoring || b.NormVersion != wire.Versions.Norm || b.Questionnaire != wire.Questionnaire || b.ScoringManifest != wire.Manifest || b.ScoringManifestSHA != wire.ManifestSHA {
		return false
	}
	m, err := DecodeManagementTraitsManifest([]byte(wire.Manifest), 1<<20)
	if err != nil || validateManagementTraitsRuntimeVersions(m.Questionnaire, m.Versions) != nil {
		return false
	}
	mc, err := CanonicalManagementTraitsManifest(m)
	if err != nil || mc.SHA256 != wire.ManifestSHA {
		return false
	}
	p, err := DecodeManagementTraitsMapping([]byte(wire.Mapping), m, 1<<20)
	if err != nil {
		return false
	}
	pc, err := CanonicalManagementTraitsMapping(m, p)
	q, versions, err2 := managementTraitsRuntimeVersions(v.RepoCode)
	return err == nil && err2 == nil && pc.SHA256 == wire.MappingSHA && q == wire.Questionnaire && versions == wire.Versions
}

func (s *ManagementTraitsFormalRegistry) Register(ctx context.Context, code, examID, assetKey string, actor int64) (ManagementTraitsFormalReadiness, error) {
	if !formalKey.MatchString(code) || !formalKey.MatchString(assetKey) || !managementTraitsOpaqueID(examID) || actor <= 0 {
		return ManagementTraitsFormalReadiness{}, ErrManagementTraitsFormalInvalid
	}
	if err := s.gate(ctx); err != nil {
		return ManagementTraitsFormalReadiness{}, err
	}
	assets, err := loadManagementTraitsFormalAssets(s.assetRoot, assetKey)
	if err != nil {
		return ManagementTraitsFormalReadiness{}, ErrManagementTraitsFormalInvalid
	}
	var out ManagementTraitsFormalReadiness
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var profile model.ManagementTraitsExamProfile
		var b model.ManagementTraitsDefinitionBundle
		if tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("exam_id = ?", examID).Take(&profile).Error != nil || tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", profile.BundleID).Take(&b).Error != nil || validateManagementTraitsRuntimeProfile(profile, b, 1<<20) != nil {
			return ErrManagementTraitsFormalInvalid
		}
		var fields managementTraitsRuntimeFieldWire
		if managementTraitsDecodeStrict([]byte(profile.FieldContract), 1<<20, &fields) != nil {
			return ErrManagementTraitsFormalInvalid
		}
		v := model.ManagementTraitsFormalVersion{ID: uuid.NewString(), VersionCode: code, Environment: s.environment, RepoCode: fields.RepoCode, BundleID: b.ID, SourceSnapshot: formalSourceSnapshot(b, profile), ContentSHA: assets.contentSHA, WorkbookSHA: assets.workbookSHA, TemplateSHA: assets.templateSHA, BindingSHA: assets.bindingSHA, AssetKey: assetKey, State: "draft", Epoch: 1, CreatedBy: actor, CreatedAt: time.Now().Truncate(time.Microsecond)}
		v.IdentitySHA = formalVersionIdentity(v)
		if tx.Create(&v).Error != nil || formalAudit(tx, v, "register", actor, "") != nil {
			return ErrManagementTraitsFormalInvalid
		}
		out = formalReadiness(v, nil, assets.templateSHA != "")
		return nil
	})
	if err != nil {
		return ManagementTraitsFormalReadiness{}, ErrManagementTraitsFormalInvalid
	}
	return out, nil
}

func formalAudit(tx *gorm.DB, v model.ManagementTraitsFormalVersion, action string, actor int64, reason string) error {
	return tx.Create(&model.ManagementTraitsFormalAudit{ID: uuid.NewString(), VersionID: v.ID, IdentitySHA: v.IdentitySHA, Action: action, ActorID: actor, Epoch: v.Epoch, Reason: reason, CreatedAt: time.Now().Truncate(time.Microsecond)}).Error
}

func (s *ManagementTraitsFormalRegistry) Change(ctx context.Context, id, sha string, epoch int64, action, kind, reason string, actor int64) (ManagementTraitsFormalReadiness, error) {
	parsed, idErr := uuid.Parse(id)
	decoded, hashErr := hex.DecodeString(sha)
	if idErr != nil || parsed.String() != id || parsed.Version() != 4 || hashErr != nil || len(decoded) != 32 || sha != strings.ToLower(sha) || epoch < 1 || actor <= 0 || !utf8.ValidString(reason) || len([]rune(reason)) > 512 || (action != "approve" && action != "activate" && action != "revoke") || (action == "approve" && kind != "content" && kind != "psychometrics") || (action == "revoke" && strings.TrimSpace(reason) == "") {
		return ManagementTraitsFormalReadiness{}, ErrManagementTraitsFormalInvalid
	}
	if err := s.gate(ctx); err != nil {
		return ManagementTraitsFormalReadiness{}, err
	}
	return s.change(ctx, id, sha, epoch, action, kind, reason, actor, nil)
}

func (s *ManagementTraitsFormalRegistry) change(ctx context.Context, id, sha string, epoch int64, action, kind, reason string, actor int64, testAssets func(model.ManagementTraitsFormalVersion) bool) (ManagementTraitsFormalReadiness, error) {
	var out ManagementTraitsFormalReadiness
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var v model.ManagementTraitsFormalVersion
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&v).Error != nil {
			return ErrManagementTraitsFormalInvalid
		}
		if v.Environment != s.environment || v.IdentitySHA != formalVersionIdentity(v) || sha != v.IdentitySHA || v.Epoch != epoch {
			return ErrManagementTraitsFormalConflict
		}
		if v.State == "revoked" || v.State != "draft" && v.State != "active" {
			return ErrManagementTraitsFormalConflict
		}
		a := make([]model.ManagementTraitsFormalApproval, 0, 2)
		if tx.Where("version_id = ?", id).Order("kind").Limit(3).Find(&a).Error != nil {
			return ErrManagementTraitsFormalInvalid
		}
		assets := false
		if action != "revoke" {
			if testAssets != nil {
				assets = testAssets(v)
			} else {
				assets = s.sourceMatch(tx, v) && s.assetsMatch(v)
			}
		}
		r := formalReadiness(v, a, assets)
		switch action {
		case "approve":
			if !r.CanApprove {
				return ErrManagementTraitsFormalInvalid
			}
			for _, existing := range a {
				if existing.Kind == kind {
					out = r
					return nil
				}
			}
			approval := model.ManagementTraitsFormalApproval{ID: uuid.NewString(), VersionID: id, Kind: kind, IdentitySHA: sha, ActorID: actor, CreatedAt: time.Now().Truncate(time.Microsecond)}
			if tx.Create(&approval).Error != nil {
				return ErrManagementTraitsFormalInvalid
			}
			a = append(a, approval)
		case "activate":
			if r.VersionReady {
				out = r
				return nil
			}
			if !r.CanActivate {
				return ErrManagementTraitsFormalInvalid
			}
			v.State = "active"
		case "revoke":
			v.State = "revoked"
		default:
			return ErrManagementTraitsFormalInvalid
		}
		if action != "approve" {
			if v.Epoch == math.MaxInt64 {
				return ErrManagementTraitsFormalInvalid
			}
			v.Epoch++
			u := tx.Model(&model.ManagementTraitsFormalVersion{}).Where("id = ? AND epoch = ?", id, epoch).Updates(map[string]any{"state": v.State, "epoch": v.Epoch})
			if u.Error != nil || u.RowsAffected != 1 {
				return ErrManagementTraitsFormalInvalid
			}
		}
		if formalAudit(tx, v, action+map[bool]string{true: "_" + kind, false: ""}[action == "approve"], actor, reason) != nil {
			return ErrManagementTraitsFormalInvalid
		}
		out = formalReadiness(v, a, assets)
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrManagementTraitsFormalConflict) {
			return ManagementTraitsFormalReadiness{}, err
		}
		return ManagementTraitsFormalReadiness{}, ErrManagementTraitsFormalInvalid
	}
	return out, nil
}

func (s *ManagementTraitsFormalRegistry) List(ctx context.Context) ([]ManagementTraitsFormalReadiness, error) {
	out := make([]ManagementTraitsFormalReadiness, 0)
	if err := s.gate(ctx); err != nil {
		return out, err
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		versions := make([]model.ManagementTraitsFormalVersion, 0)
		if tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("environment = ?", s.environment).Order("created_at DESC, id").Limit(100).Find(&versions).Error != nil {
			return ErrManagementTraitsFormalInvalid
		}
		if len(versions) == 0 {
			return nil
		}
		ids := make([]string, 0, len(versions))
		for _, v := range versions {
			ids = append(ids, v.ID)
		}
		approvals := make([]model.ManagementTraitsFormalApproval, 0)
		if tx.Where("version_id IN ?", ids).Order("version_id, kind").Find(&approvals).Error != nil {
			return ErrManagementTraitsFormalInvalid
		}
		byID := make(map[string][]model.ManagementTraitsFormalApproval)
		for _, a := range approvals {
			byID[a.VersionID] = append(byID[a.VersionID], a)
		}
		for _, v := range versions {
			out = append(out, formalReadiness(v, byID[v.ID], s.assetsMatch(v)))
		}
		return nil
	})
	if err != nil {
		return make([]ManagementTraitsFormalReadiness, 0), ErrManagementTraitsFormalInvalid
	}
	return out, nil
}
