package service

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func managementTraitsFormalTables() map[string]bool {
	return map[string]bool{"el_mng_formal_version": true, "el_mng_formal_approval": true, "el_mng_formal_audit": true}
}

func managementTraitsFormalSchemaContract() (managementTraitsSchemaDefinition, error) {
	d := managementTraitsSchemaDefinition{}
	for _, value := range []any{model.ManagementTraitsFormalVersion{}, model.ManagementTraitsFormalApproval{}, model.ManagementTraitsFormalAudit{}} {
		s, err := schema.Parse(value, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			return d, ErrManagementTraitsFormalInvalid
		}
		t := managementTraitsSchemaTable{Name: s.Table}
		primary := managementTraitsSchemaIndex{Name: "PRIMARY"}
		for _, f := range s.Fields {
			t.Columns = append(t.Columns, managementTraitsSchemaColumn{Name: f.DBName, Type: strings.ToLower(f.TagSettings["TYPE"]), Nullable: f.FieldType.Kind() == reflect.Ptr})
			if f.PrimaryKey {
				primary.Columns = append(primary.Columns, f.DBName)
			}
		}
		t.Indexes = append(t.Indexes, primary)
		for _, idx := range s.ParseIndexes() {
			i := managementTraitsSchemaIndex{Name: idx.Name}
			if idx.Class != "UNIQUE" {
				continue
			}
			for _, f := range idx.Fields {
				i.Columns = append(i.Columns, f.DBName)
			}
			t.Indexes = append(t.Indexes, i)
		}
		if s.Table == "el_mng_formal_audit" {
			t.Indexes = append(t.Indexes, managementTraitsSchemaIndex{Name: "idx_mng_formal_audit", Columns: []string{"version_id", "created_at"}, NonUnique: 1})
		}
		d.Tables = append(d.Tables, t)
	}
	d.ForeignKeys = []managementTraitsSchemaFK{{Table: "el_mng_formal_approval", Name: "fk_mng_formal_approval", RefTable: "el_mng_formal_version", Columns: []string{"version_id"}, RefColumns: []string{"id"}}, {Table: "el_mng_formal_audit", Name: "fk_mng_formal_audit", RefTable: "el_mng_formal_version", Columns: []string{"version_id"}, RefColumns: []string{"id"}}}
	return d, nil
}

func validateManagementTraitsFormalSchema(d managementTraitsSchemaDefinition, m managementTraitsSchemaMetadata) error {
	columns := make(map[string]managementTraitsSchemaColumnRow)
	for _, r := range m.Columns {
		key := r.Table + "." + r.Name
		if _, ok := columns[key]; ok {
			return ErrManagementTraitsFormalInvalid
		}
		columns[key] = r
	}
	count := 0
	for _, t := range d.Tables {
		for _, c := range t.Columns {
			count++
			r, ok := columns[t.Name+"."+c.Name]
			nullable := "NO"
			if c.Nullable {
				nullable = "YES"
			}
			if !ok || managementTraitsSchemaType(r.ColumnType) != c.Type || r.Nullable != nullable {
				return ErrManagementTraitsFormalInvalid
			}
			if strings.Contains(c.Type, "char") || c.Type == "longtext" {
				if r.Charset != "utf8mb4" || r.Collation != "utf8mb4_bin" {
					return ErrManagementTraitsFormalInvalid
				}
			}
		}
	}
	if count != len(columns) {
		return ErrManagementTraitsFormalInvalid
	}
	indexes := make(map[string][]managementTraitsSchemaIndexRow)
	for _, r := range m.Indexes {
		indexes[r.Table+"."+r.Name] = append(indexes[r.Table+"."+r.Name], r)
	}
	expected := 0
	for _, t := range d.Tables {
		for _, i := range t.Indexes {
			expected++
			rows := indexes[t.Name+"."+i.Name]
			if len(rows) != len(i.Columns) {
				return ErrManagementTraitsFormalInvalid
			}
			sort.Slice(rows, func(a, b int) bool { return rows[a].Position < rows[b].Position })
			for n, r := range rows {
				if r.Column != i.Columns[n] || r.Position != n+1 || r.NonUnique != i.NonUnique || r.Prefix != 0 {
					return ErrManagementTraitsFormalInvalid
				}
			}
		}
	}
	if expected != len(indexes) {
		return ErrManagementTraitsFormalInvalid
	}
	if len(m.ForeignKeys) != len(d.ForeignKeys) {
		return ErrManagementTraitsFormalInvalid
	}
	for _, f := range d.ForeignKeys {
		found := false
		for _, r := range m.ForeignKeys {
			if r.Table == f.Table && r.Name == f.Name && r.Column == "version_id" && r.Position == 1 && r.ReferencedTable == f.RefTable && r.ReferencedColumn == "id" && r.UpdateRule == "RESTRICT" && r.DeleteRule == "RESTRICT" {
				found = true
			}
		}
		if !found {
			return ErrManagementTraitsFormalInvalid
		}
	}
	return nil
}

// No negative cache or startup requirement: optional tables absent close only
// this capability. Every operation rechecks the installed independent contract.
func checkManagementTraitsFormalSchema(ctx context.Context, db *gorm.DB) error {
	if db == nil || ctx == nil {
		return ErrManagementTraitsFormalClosed
	}
	d, err := managementTraitsFormalSchemaContract()
	if err != nil {
		return err
	}
	names := []string{"el_mng_formal_version", "el_mng_formal_approval", "el_mng_formal_audit"}
	var tables []managementTraitsSchemaTableRow
	db = db.WithContext(ctx)
	if db.Raw("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name IN ?", names).Scan(&tables).Error != nil {
		return ErrManagementTraitsFormalInvalid
	}
	if len(tables) == 0 {
		return ErrManagementTraitsFormalClosed
	}
	installed := make(map[string]bool)
	for _, t := range tables {
		if installed[t.Name] || !managementTraitsFormalTables()[t.Name] || t.Engine != "InnoDB" {
			return ErrManagementTraitsFormalInvalid
		}
		installed[t.Name] = true
	}
	if len(installed) != 3 {
		return ErrManagementTraitsFormalInvalid
	}
	var m managementTraitsSchemaMetadata
	if db.Raw("SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type, is_nullable AS is_nullable, COALESCE(character_set_name,'') AS character_set_name, COALESCE(collation_name,'') AS collation_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name IN ? ORDER BY table_name, ordinal_position", names).Scan(&m.Columns).Error != nil {
		return ErrManagementTraitsFormalInvalid
	}
	if db.Raw("SELECT table_name AS table_name, index_name AS index_name, non_unique AS non_unique, seq_in_index AS seq_in_index, column_name AS column_name, COALESCE(sub_part,0) AS prefix_length FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name IN ? ORDER BY table_name,index_name,seq_in_index", names).Scan(&m.Indexes).Error != nil {
		return ErrManagementTraitsFormalInvalid
	}
	if db.Raw("SELECT k.table_name AS table_name, k.constraint_name AS constraint_name, k.column_name AS column_name, k.ordinal_position AS ordinal_position, k.referenced_table_name AS referenced_table_name, k.referenced_column_name AS referenced_column_name, r.update_rule AS update_rule, r.delete_rule AS delete_rule FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.constraint_schema=DATABASE() AND k.table_name IN ? ORDER BY k.table_name,k.constraint_name,k.ordinal_position", names).Scan(&m.ForeignKeys).Error != nil {
		return ErrManagementTraitsFormalInvalid
	}
	if validateManagementTraitsFormalSchema(d, m) != nil {
		return ErrManagementTraitsFormalInvalid
	}
	rows := make([]struct {
		Invalid int `gorm:"column:invalid"`
	}, 0)
	const integrity = "SELECT 1 AS invalid FROM el_mng_formal_approval a LEFT JOIN el_mng_formal_version v ON v.id=a.version_id WHERE v.id IS NULL OR a.identity_sha<>v.identity_sha OR a.actor_id<=0 OR a.kind NOT IN ('content','psychometrics') UNION ALL SELECT 1 AS invalid FROM el_mng_formal_audit a LEFT JOIN el_mng_formal_version v ON v.id=a.version_id WHERE v.id IS NULL OR a.identity_sha<>v.identity_sha OR a.actor_id<=0 OR a.epoch<1 OR a.epoch>v.epoch UNION ALL SELECT 1 AS invalid FROM el_mng_formal_version v WHERE v.environment<>'local' OR v.repo_code NOT IN ('00201','00202','00501','00502') OR v.state NOT IN ('draft','active','revoked') OR v.epoch<1 OR v.created_by<=0 LIMIT 1"
	if db.Raw(integrity).Scan(&rows).Error != nil || len(rows) != 0 {
		return ErrManagementTraitsFormalInvalid
	}
	return nil
}
