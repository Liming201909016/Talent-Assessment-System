package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrManagementTraitsDraftClosed = errors.New("管理特质新版草稿结构未安装或无效，请联系管理员完成迁移")
var ErrManagementTraitsDraftPolicy = errors.New("新版请使用00501/00502单一140题草稿、25分钟和受支持身份字段；已有002新版配置保留兼容，不可跨产品转换")

// Optional capability, never a startup/negative-cache requirement.
func CheckManagementTraitsDraftSchema(ctx context.Context, db *gorm.DB) error {
	if ctx == nil || db == nil {
		return ErrManagementTraitsDraftClosed
	}
	db = db.WithContext(ctx)
	var tables []managementTraitsSchemaTableRow
	if db.Raw("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'el_mng_exam_draft'").Scan(&tables).Error != nil || len(tables) != 1 || tables[0].Name != "el_mng_exam_draft" || tables[0].Engine != "InnoDB" {
		return ErrManagementTraitsDraftClosed
	}
	var cols []managementTraitsSchemaColumnRow
	if db.Raw("SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type, is_nullable AS is_nullable, COALESCE(character_set_name,'') AS character_set_name, COALESCE(collation_name,'') AS collation_name FROM information_schema.columns WHERE table_schema = DATABASE() AND (table_name = 'el_mng_exam_draft' OR (table_name = 'el_exam' AND column_name = 'id')) ORDER BY table_name, ordinal_position").Scan(&cols).Error != nil || len(cols) != 8 {
		return ErrManagementTraitsDraftClosed
	}
	byName := make(map[string]managementTraitsSchemaColumnRow, 8)
	for _, c := range cols {
		key := c.Table + "." + c.Name
		if _, exists := byName[key]; exists {
			return ErrManagementTraitsDraftClosed
		}
		byName[key] = c
	}
	parent, ok := byName["el_exam.id"]
	if !ok || parent.ColumnType != "varchar(64)" || parent.Nullable != "NO" || parent.Charset != "utf8mb4" || !strings.HasPrefix(parent.Collation, "utf8mb4_") {
		return ErrManagementTraitsDraftClosed
	}
	for name, typ := range map[string]string{"exam_id": "varchar(64)", "repo_id": "varchar(64)", "repo_code": "varchar(5)", "lifecycle": "varchar(6)", "created_at": "datetime(6)", "updated_at": "datetime(6)", "frozen_at": "datetime(6)"} {
		c, exists := byName["el_mng_exam_draft."+name]
		nullable := "NO"
		if name == "frozen_at" {
			nullable = "YES"
		}
		if !exists || c.ColumnType != typ || c.Nullable != nullable {
			return ErrManagementTraitsDraftClosed
		}
		if strings.HasPrefix(typ, "varchar") {
			coll := "utf8mb4_bin"
			if name == "exam_id" {
				coll = parent.Collation
			}
			if c.Charset != "utf8mb4" || c.Collation != coll {
				return ErrManagementTraitsDraftClosed
			}
		} else if c.Charset != "" || c.Collation != "" {
			return ErrManagementTraitsDraftClosed
		}
	}
	var indexes []managementTraitsSchemaIndexRow
	if db.Raw("SELECT table_name AS table_name, index_name AS index_name, non_unique AS non_unique, seq_in_index AS seq_in_index, column_name AS column_name, COALESCE(sub_part,0) AS prefix_length FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'el_mng_exam_draft'").Scan(&indexes).Error != nil || len(indexes) != 1 || indexes[0].Name != "PRIMARY" || indexes[0].Column != "exam_id" || indexes[0].Position != 1 || indexes[0].NonUnique != 0 || indexes[0].Prefix != 0 {
		return ErrManagementTraitsDraftClosed
	}
	var fks []managementTraitsSchemaFKRow
	if db.Raw("SELECT k.table_name AS table_name, k.constraint_name AS constraint_name, k.column_name AS column_name, k.ordinal_position AS ordinal_position, CASE WHEN k.referenced_table_schema=DATABASE() THEN k.referenced_table_name ELSE '' END AS referenced_table_name, k.referenced_column_name AS referenced_column_name, r.update_rule AS update_rule, r.delete_rule AS delete_rule FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.constraint_schema=DATABASE() AND k.table_name = 'el_mng_exam_draft' AND k.referenced_table_name IS NOT NULL").Scan(&fks).Error != nil || len(fks) != 1 {
		return ErrManagementTraitsDraftClosed
	}
	f := fks[0]
	if f.Name != "fk_mng_draft_exam" || f.Column != "exam_id" || f.Position != 1 || f.ReferencedTable != "el_exam" || f.ReferencedColumn != "id" || f.UpdateRule != "RESTRICT" || f.DeleteRule != "RESTRICT" {
		return ErrManagementTraitsDraftClosed
	}
	return nil
}

func ReadManagementTraitsDraft(ctx context.Context, db *gorm.DB, examID string) (*model.ManagementTraitsExamDraft, error) {
	var count int64
	if db.WithContext(ctx).Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'el_mng_exam_draft'").Scan(&count).Error != nil || count < 0 || count > 1 {
		return nil, ErrManagementTraitsDraftClosed
	}
	if count == 0 {
		return nil, nil
	}
	if CheckManagementTraitsDraftSchema(ctx, db) != nil {
		return nil, ErrManagementTraitsDraftClosed
	}
	return managementTraitsInstalledDraft(ctx, db, examID)
}

func managementTraitsInstalledDraft(ctx context.Context, db *gorm.DB, examID string) (*model.ManagementTraitsExamDraft, error) {
	rows, err := managementTraitsInstalledDrafts(ctx, db, []string{examID})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	d := rows[0]
	return &d, nil
}

func managementTraitsInstalledDrafts(ctx context.Context, db *gorm.DB, ids []string) ([]model.ManagementTraitsExamDraft, error) {
	rows := make([]model.ManagementTraitsExamDraft, 0)
	seen := make(map[string]bool)
	for len(ids) > 0 {
		n := len(ids)
		if n > 1000 {
			n = 1000
		}
		batch := ids[:n]
		ids = ids[n:]
		allowed := make(map[string]bool, n)
		for _, id := range batch {
			allowed[id] = true
		}
		part := make([]model.ManagementTraitsExamDraft, 0)
		if db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("exam_id IN ?", batch).Limit(n+1).Find(&part).Error != nil || len(part) > n {
			return nil, ErrManagementTraitsRuntimeInvalid
		}
		for _, d := range part {
			if !allowed[d.ExamID] || seen[d.ExamID] || !managementTraitsOpaqueID(d.RepoID) || !IsManagementTraitsProduct(d.RepoCode) || d.CreatedAt.IsZero() || d.UpdatedAt.Before(d.CreatedAt) || (d.Lifecycle != "draft" && d.Lifecycle != "frozen") || (d.Lifecycle == "draft" && d.FrozenAt != nil) || (d.Lifecycle == "frozen" && (d.FrozenAt == nil || d.FrozenAt.IsZero())) {
				return nil, ErrManagementTraitsRuntimeInvalid
			}
			seen[d.ExamID] = true
			rows = append(rows, d)
		}
	}
	return rows, nil
}

// Called by the actual Save transaction before its first DML. The DB repo code
// is authoritative; request code/boolean cannot manufacture a legacy product.
func PrepareManagementTraitsDraft(ctx context.Context, tx *gorm.DB, exam model.Exam, repos []model.ExamRepo, isNew bool, requested *bool) (*model.ManagementTraitsExamDraft, error) {
	var prior *model.ManagementTraitsExamDraft
	var err error
	if !isNew {
		prior, err = ReadManagementTraitsDraft(ctx, tx, exam.ID)
		if err != nil {
			return nil, err
		}
		if prior != nil && prior.Lifecycle != "draft" {
			return nil, ErrManagementTraitsDraftPolicy
		}
	}
	ids := make([]string, 0, len(repos))
	for _, r := range repos {
		ids = append(ids, r.RepoID)
	}
	actual := make([]model.Repo, 0)
	if len(ids) > 0 && tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id IN ?", ids).Find(&actual).Error != nil {
		return nil, ErrManagementTraitsDraftPolicy
	}
	new005 := false
	for _, r := range actual {
		new005 = new005 || ClassifyManagementTraitsProduct(r.Code) == ManagementTraitsNew005
	}
	if prior == nil && !new005 {
		if requested != nil && *requested {
			return nil, ErrManagementTraitsDraftPolicy
		}
		return nil, nil
	}
	if !isNew && prior == nil {
		return nil, ErrManagementTraitsDraftPolicy
	}
	if requested != nil && !*requested {
		return nil, ErrManagementTraitsDraftPolicy
	}
	if len(actual) != 1 || len(repos) != 1 || !IsManagementTraitsProduct(actual[0].Code) || (prior != nil && ClassifyManagementTraitsProduct(prior.RepoCode) != ClassifyManagementTraitsProduct(actual[0].Code)) || exam.AssessmentType != "legacy" || exam.ScoringMode != "legacy" || exam.JoinType != 1 || exam.TotalTime != 25 || exam.ShowPdf != 0 || (exam.IsOpen != 1 && exam.IsOpen != 2) || actual[0].RadioCount != 140 || actual[0].MultiCount != 0 || actual[0].JudgeCount != 0 || repos[0].RadioCount != 140 || repos[0].MultiCount != 0 || repos[0].JudgeCount != 0 || repos[0].SaqCount != 0 {
		return nil, ErrManagementTraitsDraftPolicy
	}
	fields := strings.Split(exam.RequiredFields, ",")
	age, flag, text := 1, 0, "present"
	if !managementTraitsRuntimeFields(fields, managementTraitsRuntimeIdentity{Name: text, Gender: text, Telephone: text, Affiliation: text, Post: text, Age: &age, Degree: &text, Major: &text, StuFlag: &flag}) {
		return nil, ErrManagementTraitsDraftPolicy
	}
	if isNew && CheckManagementTraitsDraftSchema(ctx, tx) != nil {
		return nil, ErrManagementTraitsDraftClosed
	}
	now := time.Now()
	d := &model.ManagementTraitsExamDraft{ExamID: exam.ID, RepoID: actual[0].ID, RepoCode: actual[0].Code, Lifecycle: "draft", CreatedAt: now, UpdatedAt: now}
	if prior != nil {
		d.CreatedAt = prior.CreatedAt
	}
	return d, nil
}

func PersistManagementTraitsDraft(tx *gorm.DB, d *model.ManagementTraitsExamDraft, isNew bool) error {
	if d == nil {
		return nil
	}
	if isNew {
		if tx.Create(d).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		return nil
	}
	r := tx.Model(&model.ManagementTraitsExamDraft{}).Where("exam_id = ? AND lifecycle = ?", d.ExamID, "draft").Updates(map[string]any{"repo_id": d.RepoID, "repo_code": d.RepoCode, "updated_at": d.UpdatedAt})
	if r.Error != nil || r.RowsAffected != 1 {
		return ErrManagementTraitsRuntimeInvalid
	}
	return nil
}

// Metadata shares the identity probe, so old installations require no extra
// table read and do not turn schema absence into an inferred new draft.
func ManagementTraitsExamLifecycle(ctx context.Context, db *gorm.DB, examID string) (string, error) {
	schema, err := probeManagementTraitsIdentitySchema(ctx, db)
	if err != nil {
		return "", err
	}
	var d *model.ManagementTraitsExamDraft
	if schema.draft {
		if CheckManagementTraitsDraftSchema(ctx, db) != nil {
			return "", ErrManagementTraitsDraftClosed
		}
		d, err = managementTraitsInstalledDraft(ctx, db, examID)
		if err != nil {
			return "", err
		}
	}
	frozen := false
	if schema.present {
		frozen, err = managementTraitsIdentityHasProfile(ctx, db, examID)
		if err != nil {
			return "", err
		}
	}
	if d != nil {
		if (d.Lifecycle == "frozen") != frozen {
			return "", ErrManagementTraitsRuntimeInvalid
		}
		return d.Lifecycle, nil
	}
	if frozen {
		return "frozen", nil
	}
	return "legacy", nil
}
