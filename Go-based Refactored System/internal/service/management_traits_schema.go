package service

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type managementTraitsSchemaColumn struct {
	Name, Type string
	Nullable   bool
}
type managementTraitsSchemaIndex struct {
	Name      string
	Columns   []string
	NonUnique int
	Prefixes  []int
}
type managementTraitsSchemaTable struct {
	Name    string
	Columns []managementTraitsSchemaColumn
	Indexes []managementTraitsSchemaIndex
}
type managementTraitsSchemaFK struct {
	Table, Name, RefTable string
	Columns, RefColumns   []string
}
type managementTraitsSchemaDefinition struct {
	Tables      []managementTraitsSchemaTable
	ForeignKeys []managementTraitsSchemaFK
}

// Every information_schema projection has explicit column tags. MySQL 5.7
// integer display widths are normalised, not confused with data precision.
type managementTraitsSchemaTableRow struct {
	Name   string `gorm:"column:table_name"`
	Engine string `gorm:"column:engine"`
}
type managementTraitsSchemaColumnRow struct {
	Table      string `gorm:"column:table_name"`
	Name       string `gorm:"column:column_name"`
	ColumnType string `gorm:"column:column_type"`
	Nullable   string `gorm:"column:is_nullable"`
	Charset    string `gorm:"column:character_set_name"`
	Collation  string `gorm:"column:collation_name"`
}
type managementTraitsSchemaIndexRow struct {
	Table     string `gorm:"column:table_name"`
	Name      string `gorm:"column:index_name"`
	NonUnique int    `gorm:"column:non_unique"`
	Position  int    `gorm:"column:seq_in_index"`
	Column    string `gorm:"column:column_name"`
	Prefix    int    `gorm:"column:prefix_length"`
}
type managementTraitsSchemaFKRow struct {
	Table            string `gorm:"column:table_name"`
	Name             string `gorm:"column:constraint_name"`
	Column           string `gorm:"column:column_name"`
	Position         int    `gorm:"column:ordinal_position"`
	ReferencedTable  string `gorm:"column:referenced_table_name"`
	ReferencedColumn string `gorm:"column:referenced_column_name"`
	UpdateRule       string `gorm:"column:update_rule"`
	DeleteRule       string `gorm:"column:delete_rule"`
}
type managementTraitsSchemaMetadata struct {
	Columns     []managementTraitsSchemaColumnRow
	Indexes     []managementTraitsSchemaIndexRow
	ForeignKeys []managementTraitsSchemaFKRow
}

func managementTraitsSchemaContract() (managementTraitsSchemaDefinition, error) {
	var d managementTraitsSchemaDefinition
	models := []any{model.ManagementTraitsDefinitionBundle{}, model.ManagementTraitsExamProfile{}, model.ManagementTraitsPaperSnapshot{}, model.ManagementTraitsPaperQuestionSnapshot{}, model.ManagementTraitsResultRun{}, model.ManagementTraitsResultDimension{}, model.ManagementTraitsResultModule{}, model.ManagementTraitsRuntimeReceipt{}, model.ManagementTraitsReportRevision{}, model.ManagementTraitsReportCurrent{}, model.ManagementTraitsReportAudit{}}
	for _, value := range models {
		s, err := schema.Parse(value, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			return d, err
		}
		t := managementTraitsSchemaTable{Name: s.Table}
		primary := managementTraitsSchemaIndex{Name: "PRIMARY"}
		for _, f := range s.Fields {
			typ := f.FieldType
			nullable := typ.Kind() == reflect.Ptr
			if nullable {
				typ = typ.Elem()
			}
			dbType := strings.ToLower(f.TagSettings["TYPE"])
			if dbType == "" {
				switch {
				case typ == reflect.TypeOf(time.Time{}):
					dbType = "datetime(6)"
				case typ == reflect.TypeOf(decimal.Decimal{}):
					return d, fmt.Errorf("decimal type missing")
				case typ.Kind() == reflect.String:
					dbType = "varchar(255)"
				case typ.Kind() == reflect.Bool:
					dbType = "tinyint(1)"
				case typ.Kind() == reflect.Int || typ.Kind() == reflect.Int64:
					dbType = "bigint"
				default:
					return d, fmt.Errorf("unsupported schema field")
				}
			}
			t.Columns = append(t.Columns, managementTraitsSchemaColumn{f.DBName, dbType, nullable})
			if f.PrimaryKey {
				primary.Columns = append(primary.Columns, f.DBName)
			}
		}
		t.Indexes = append(t.Indexes, primary)
		for _, idx := range s.ParseIndexes() {
			if idx.Class != "UNIQUE" {
				continue
			}
			i := managementTraitsSchemaIndex{Name: idx.Name}
			for _, f := range idx.Fields {
				i.Columns = append(i.Columns, f.DBName)
			}
			t.Indexes = append(t.Indexes, i)
		}
		switch t.Name {
		case "el_mng_exam_profile":
			t.Indexes = append(t.Indexes, managementTraitsSchemaIndex{Name: "idx_mng_profile_bundle", Columns: []string{"bundle_id"}, NonUnique: 1})
		case "el_mng_paper_snapshot":
			t.Indexes = append(t.Indexes, managementTraitsSchemaIndex{Name: "idx_mng_snapshot_owner", Columns: []string{"participant_id", "participant_type"}, NonUnique: 1, Prefixes: []int{0, 16}}, managementTraitsSchemaIndex{Name: "idx_mng_snapshot_expiry", Columns: []string{"limit_time", "paper_id"}, NonUnique: 1})
		case "el_mng_result_run":
			t.Indexes = append(t.Indexes, managementTraitsSchemaIndex{Name: "idx_mng_run_exam", Columns: []string{"exam_id", "submitted_at", "id"}, NonUnique: 1})
		case "el_mng_report_audit":
			t.Indexes = append(t.Indexes, managementTraitsSchemaIndex{Name: "idx_mng_audit_report", Columns: []string{"report_id", "created_at"}, NonUnique: 1})
		}
		sort.Slice(t.Indexes, func(i, j int) bool { return t.Indexes[i].Name < t.Indexes[j].Name })
		d.Tables = append(d.Tables, t)
	}
	add := func(table, name, refs string, cols, refcols []string) {
		d.ForeignKeys = append(d.ForeignKeys, managementTraitsSchemaFK{table, name, refs, cols, refcols})
	}
	add("el_mng_exam_profile", "fk_mng_profile_exam", "el_exam", []string{"exam_id"}, []string{"id"})
	add("el_mng_exam_profile", "fk_mng_profile_bundle", "el_mng_definition_bundle", []string{"bundle_id"}, []string{"id"})
	add("el_mng_paper_snapshot", "fk_mng_snapshot_paper", "el_paper", []string{"paper_id"}, []string{"id"})
	add("el_mng_paper_snapshot", "fk_mng_snapshot_exam", "el_exam", []string{"exam_id"}, []string{"id"})
	add("el_mng_paper_snapshot", "fk_mng_snapshot_profile", "el_mng_exam_profile", []string{"profile_exam_id"}, []string{"exam_id"})
	add("el_mng_paper_snapshot", "fk_mng_snapshot_bundle", "el_mng_definition_bundle", []string{"bundle_id"}, []string{"id"})
	add("el_mng_paper_question_snapshot", "fk_mng_question_paper", "el_mng_paper_snapshot", []string{"paper_id"}, []string{"paper_id"})
	add("el_mng_paper_question_snapshot", "fk_mng_question_pq", "el_paper_qu", []string{"paper_question_id"}, []string{"id"})
	add("el_mng_result_run", "fk_mng_run_snapshot", "el_mng_paper_snapshot", []string{"paper_id", "exam_id"}, []string{"paper_id", "exam_id"})
	add("el_mng_result_dimension", "fk_mng_dimension_run", "el_mng_result_run", []string{"run_id"}, []string{"id"})
	add("el_mng_result_module", "fk_mng_module_run", "el_mng_result_run", []string{"run_id"}, []string{"id"})
	add("el_mng_runtime_receipt", "fk_mng_receipt_run", "el_mng_result_run", []string{"run_id", "paper_id", "exam_id"}, []string{"id", "paper_id", "exam_id"})
	add("el_mng_report_revision", "fk_mng_report_run", "el_mng_result_run", []string{"run_id", "paper_id", "exam_id"}, []string{"id", "paper_id", "exam_id"})
	add("el_mng_report_current", "fk_mng_current_report", "el_mng_report_revision", []string{"report_id", "paper_id"}, []string{"id", "paper_id"})
	add("el_mng_report_audit", "fk_mng_audit_report", "el_mng_report_revision", []string{"report_id"}, []string{"id"})
	return d, nil
}

var managementTraitsSchemaBigintWidth = regexp.MustCompile(`^bigint\(([1-9][0-9]{0,2})\)$`)
var managementTraitsSchemaParentType = regexp.MustCompile(`^varchar\(([1-9][0-9]?)\)$`)
var managementTraitsSchemaCollation = regexp.MustCompile(`^utf8mb4_[a-z0-9_]+$`)

// Reference edges describe actual runtime copies/JOINs, not index names.
// Candidate/tester are polymorphic parents; both must fit the shared sinks.
func managementTraitsSchemaReferences() map[string][]string {
	return map[string][]string{
		"el_exam.id":                  {"el_exam_repo.exam_id", "el_paper.exam_id", "el_candidate.exam_id", "el_tester.exam_id", "el_mng_exam_profile.exam_id", "el_mng_paper_snapshot.exam_id", "el_mng_result_run.exam_id", "el_mng_runtime_receipt.exam_id", "el_mng_report_revision.exam_id"},
		"el_paper.id":                 {"el_candidate.paper_id", "el_tester.paper_id", "el_paper_qu.paper_id", "el_paper_qu_answer.paper_id", "el_mng_paper_snapshot.paper_id", "el_mng_paper_question_snapshot.paper_id", "el_mng_result_run.paper_id", "el_mng_runtime_receipt.paper_id", "el_mng_report_revision.paper_id", "el_mng_report_current.paper_id"},
		"el_paper_qu.id":              {"el_mng_paper_question_snapshot.paper_question_id"},
		"el_mng_exam_profile.exam_id": {"el_mng_paper_snapshot.profile_exam_id"},
		"el_candidate.id":             {"el_paper.user_id", "el_mng_paper_snapshot.participant_id", "el_mng_result_run.participant_id", "el_mng_runtime_receipt.participant_id"},
		"el_tester.id":                {"el_paper.user_id", "el_mng_paper_snapshot.participant_id", "el_mng_result_run.participant_id", "el_mng_runtime_receipt.participant_id"},
		"el_repo.id":                  {"el_exam_repo.repo_id", "el_qu_repo.repo_id"},
		"el_qu.id":                    {"el_qu_repo.qu_id", "el_qu_answer.qu_id", "el_paper_qu.qu_id", "el_paper_qu_answer.qu_id", "el_mng_paper_question_snapshot.source_question_id"},
		"el_qu_answer.id":             {"el_paper_qu_answer.answer_id", "el_mng_paper_question_snapshot.selected_option_id"},
		"el_qu_repo.id":               {}, "el_paper_qu_answer.id": {},
	}
}

type managementTraitsSchemaIDBudget struct {
	Bytes     int
	ASCIIOnly bool
}

// Only these observed legacy edges may use a narrower sink or mixed encoding.
// They are not sidecar FKs; every actual ID must fit all sinks as ASCII.
func managementTraitsSchemaLegacyNarrowEdge(parent, child string, columns map[string]managementTraitsSchemaColumnRow) bool {
	if !((parent == "el_qu.id" && child == "el_paper_qu_answer.qu_id") || (parent == "el_qu_answer.id" && child == "el_paper_qu_answer.answer_id")) {
		return false
	}
	p, c := columns[parent], columns[child]
	return p.ColumnType == "varchar(64)" && c.ColumnType == "varchar(32)" && p.Nullable == "NO" && c.Nullable == "NO" && p.Charset == "utf8mb4" && c.Charset == p.Charset && p.Collation == "utf8mb4_0900_ai_ci" && c.Collation == p.Collation
}

func managementTraitsSchemaLegacyRepo(columns map[string]managementTraitsSchemaColumnRow, base managementTraitsSchemaColumnRow) bool {
	p := columns["el_repo.id"]
	if base.Charset != "utf8mb4" || base.Collation != "utf8mb4_0900_ai_ci" || p.ColumnType != "varchar(64)" || p.Nullable != "NO" || p.Charset != "utf8mb3" || p.Collation != "utf8mb3_general_ci" {
		return false
	}
	for _, key := range []string{"el_exam_repo.repo_id", "el_qu_repo.repo_id"} {
		c := columns[key]
		if c.ColumnType != "varchar(64)" || c.Nullable != "NO" || c.Charset != base.Charset || c.Collation != base.Collation {
			return false
		}
	}
	return true
}

func managementTraitsSchemaReferenceCapacities(columns map[string]managementTraitsSchemaColumnRow, base managementTraitsSchemaColumnRow) (map[string]managementTraitsSchemaIDBudget, error) {
	capacity := make(map[string]managementTraitsSchemaIDBudget)
	legacyRepo := managementTraitsSchemaLegacyRepo(columns, base)
	refs := managementTraitsSchemaReferences()
	for parent, children := range refs {
		for _, key := range append([]string{parent}, children...) {
			if _, exists := capacity[key]; exists {
				continue
			}
			row, exists := columns[key]
			match := managementTraitsSchemaParentType.FindStringSubmatch(strings.ToLower(row.ColumnType))
			if !exists || match == nil || ((row.Charset != base.Charset || row.Collation != base.Collation) && !(key == "el_repo.id" && legacyRepo)) {
				return nil, ErrManagementTraitsRuntimeInvalid
			}
			length, err := strconv.Atoi(match[1])
			if err != nil || length > 64 {
				return nil, ErrManagementTraitsRuntimeInvalid
			}
			nullable := "NO"
			if key == "el_candidate.paper_id" || key == "el_tester.paper_id" || key == "el_tester.exam_id" || key == "el_mng_paper_snapshot.profile_exam_id" || key == "el_mng_paper_question_snapshot.selected_option_id" {
				nullable = "YES"
			}
			if row.Nullable != nullable {
				return nil, ErrManagementTraitsRuntimeInvalid
			}
			if (key == "el_candidate.id" || key == "el_paper.id" || key == "el_paper_qu.id" || key == "el_paper_qu_answer.id") && length < 36 {
				return nil, ErrManagementTraitsRuntimeInvalid
			}
			capacity[key] = managementTraitsSchemaIDBudget{Bytes: length, ASCIIOnly: key == "el_repo.id" && legacyRepo}
		}
	}
	for parent, children := range refs {
		for _, child := range children {
			if capacity[child].Bytes < capacity[parent].Bytes {
				if !managementTraitsSchemaLegacyNarrowEdge(parent, child, columns) {
					return nil, ErrManagementTraitsRuntimeInvalid
				}
				capacity[parent] = managementTraitsSchemaIDBudget{Bytes: capacity[child].Bytes, ASCIIOnly: true}
			}
		}
	}
	return capacity, nil
}

func managementTraitsSchemaType(raw string) string {
	raw = strings.ToLower(raw)
	if match := managementTraitsSchemaBigintWidth.FindStringSubmatch(raw); match != nil {
		width, err := strconv.Atoi(match[1])
		if err == nil && width <= 255 {
			return "bigint"
		}
	}
	return raw
}

func validateManagementTraitsSchema(d managementTraitsSchemaDefinition, m managementTraitsSchemaMetadata) error {
	columns := make(map[string]managementTraitsSchemaColumnRow, len(m.Columns))
	for _, c := range m.Columns {
		k := c.Table + "." + c.Name
		if _, ok := columns[k]; ok {
			return ErrManagementTraitsRuntimeInvalid
		}
		columns[k] = c
	}
	base, ok := columns["el_exam.id"]
	if !ok || base.Charset != "utf8mb4" || !managementTraitsSchemaCollation.MatchString(base.Collation) {
		return ErrManagementTraitsRuntimeInvalid
	}
	// Exam IDs are external; paper and paper-question IDs are generated UUIDs.
	// Different varchar lengths are allowed, but generated parents must fit 36.
	for _, name := range []string{"el_exam", "el_paper", "el_paper_qu"} {
		row, ok := columns[name+".id"]
		match := managementTraitsSchemaParentType.FindStringSubmatch(strings.ToLower(row.ColumnType))
		if !ok || match == nil || row.Nullable != "NO" || row.Charset != base.Charset || row.Collation != base.Collation {
			return ErrManagementTraitsRuntimeInvalid
		}
		length, err := strconv.Atoi(match[1])
		if err != nil || length > 64 || (name != "el_exam" && length < 36) {
			return ErrManagementTraitsRuntimeInvalid
		}
	}
	for _, t := range d.Tables {
		for _, c := range t.Columns {
			row, ok := columns[t.Name+"."+c.Name]
			nullable := "NO"
			if c.Nullable {
				nullable = "YES"
			}
			if !ok || managementTraitsSchemaType(row.ColumnType) != c.Type || row.Nullable != nullable {
				return ErrManagementTraitsRuntimeInvalid
			}
			if strings.HasPrefix(c.Type, "varchar") || strings.HasPrefix(c.Type, "char") || c.Type == "longtext" {
				if row.Charset != base.Charset || row.Collation != base.Collation {
					return ErrManagementTraitsRuntimeInvalid
				}
			}
		}
	}
	if _, err := managementTraitsSchemaReferenceCapacities(columns, base); err != nil {
		return err
	}
	indexes := make(map[string][]managementTraitsSchemaIndexRow)
	for _, i := range m.Indexes {
		indexes[i.Table+"."+i.Name] = append(indexes[i.Table+"."+i.Name], i)
	}
	for _, t := range d.Tables {
		for _, i := range t.Indexes {
			rows := indexes[t.Name+"."+i.Name]
			if len(rows) != len(i.Columns) {
				return ErrManagementTraitsRuntimeInvalid
			}
			sort.Slice(rows, func(a, b int) bool { return rows[a].Position < rows[b].Position })
			for n, r := range rows {
				prefix := 0
				if len(i.Prefixes) > n {
					prefix = i.Prefixes[n]
				}
				if r.NonUnique != i.NonUnique || r.Position != n+1 || r.Column != i.Columns[n] || r.Prefix != prefix {
					return ErrManagementTraitsRuntimeInvalid
				}
			}
		}
	}
	fks := make(map[string][]managementTraitsSchemaFKRow)
	for _, f := range m.ForeignKeys {
		fks[f.Table+"."+f.Name] = append(fks[f.Table+"."+f.Name], f)
	}
	for _, f := range d.ForeignKeys {
		rows := fks[f.Table+"."+f.Name]
		if len(rows) != len(f.Columns) {
			return ErrManagementTraitsRuntimeInvalid
		}
		sort.Slice(rows, func(a, b int) bool { return rows[a].Position < rows[b].Position })
		for n, r := range rows {
			if r.Position != n+1 || r.Column != f.Columns[n] || r.ReferencedTable != f.RefTable || r.ReferencedColumn != f.RefColumns[n] || r.UpdateRule != "RESTRICT" || r.DeleteRule != "RESTRICT" {
				return ErrManagementTraitsRuntimeInvalid
			}
			a, aok := columns[f.Table+"."+r.Column]
			b, bok := columns[f.RefTable+"."+r.ReferencedColumn]
			if !aok || !bok || a.Charset != b.Charset || a.Collation != b.Collation {
				return ErrManagementTraitsRuntimeInvalid
			}
			am, bm := managementTraitsSchemaParentType.FindStringSubmatch(a.ColumnType), managementTraitsSchemaParentType.FindStringSubmatch(b.ColumnType)
			if am == nil || bm == nil {
				return ErrManagementTraitsRuntimeInvalid
			}
			ac, _ := strconv.Atoi(am[1])
			bc, _ := strconv.Atoi(bm[1])
			if ac < bc {
				return ErrManagementTraitsRuntimeInvalid
			}
		}
	}
	return nil
}

// Cache belongs to one runtime service instance. Metadata errors and absence
// fail closed and remain cached; restart after a migration. Never query inside
// paper/profile transactions, never infer approval from table existence.
func (s *ManagementTraitsRuntimeService) CheckRuntimeSchema(ctx context.Context) error {
	if s == nil {
		return ErrManagementTraitsRuntimeClosed
	}
	if ctx == nil || ctx.Err() != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	s.schemaOnce.Do(func() {
		if s.db == nil {
			if s.schemaErr == nil {
				s.schemaErr = ErrManagementTraitsRuntimeClosed
			}
			return
		}
		s.schemaErr = s.checkRuntimeSchema(ctx)
	})
	return s.schemaErr
}

func (s *ManagementTraitsRuntimeService) checkRuntimeSchema(ctx context.Context) error {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	var tables []managementTraitsSchemaTableRow
	if s.db.WithContext(ctx).Raw("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name, 7) = 'el_mng_' ORDER BY table_name").Scan(&tables).Error != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	if len(tables) == 0 {
		return ErrManagementTraitsRuntimeClosed
	}
	installed := make(map[string]string)
	for _, t := range tables {
		if _, exists := installed[t.Name]; exists {
			return ErrManagementTraitsRuntimeInvalid
		}
		installed[t.Name] = t.Engine
	}
	for _, t := range d.Tables {
		if installed[t.Name] != "InnoDB" {
			return ErrManagementTraitsRuntimeInvalid
		}
	}
	var m managementTraitsSchemaMetadata
	if s.db.WithContext(ctx).Raw("SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type, is_nullable AS is_nullable, COALESCE(character_set_name,'') AS character_set_name, COALESCE(collation_name,'') AS collation_name FROM information_schema.columns WHERE table_schema = DATABASE() AND (LEFT(table_name, 7) = 'el_mng_' OR (table_name IN ('el_exam','el_exam_repo','el_paper','el_paper_qu','el_paper_qu_answer','el_candidate','el_tester','el_repo','el_qu_repo','el_qu','el_qu_answer') AND column_name IN ('id','exam_id','paper_id','user_id','repo_id','qu_id','answer_id'))) ORDER BY table_name, ordinal_position").Scan(&m.Columns).Error != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	if s.db.WithContext(ctx).Raw("SELECT table_name AS table_name, index_name AS index_name, non_unique AS non_unique, seq_in_index AS seq_in_index, column_name AS column_name, COALESCE(sub_part,0) AS prefix_length FROM information_schema.statistics WHERE table_schema = DATABASE() AND LEFT(table_name, 7) = 'el_mng_' ORDER BY table_name,index_name,seq_in_index").Scan(&m.Indexes).Error != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	if s.db.WithContext(ctx).Raw("SELECT k.table_name AS table_name, k.constraint_name AS constraint_name, k.column_name AS column_name, k.ordinal_position AS ordinal_position, k.referenced_table_name AS referenced_table_name, k.referenced_column_name AS referenced_column_name, r.update_rule AS update_rule, r.delete_rule AS delete_rule FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.constraint_schema=DATABASE() AND LEFT(k.table_name, 7)='el_mng_' ORDER BY k.table_name,k.constraint_name,k.ordinal_position").Scan(&m.ForeignKeys).Error != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	if err := validateManagementTraitsSchema(d, m); err != nil {
		return err
	}
	columns := make(map[string]managementTraitsSchemaColumnRow, len(m.Columns))
	for _, c := range m.Columns {
		columns[c.Table+"."+c.Name] = c
	}
	capacity, err := managementTraitsSchemaReferenceCapacities(columns, columns["el_exam.id"])
	if err != nil {
		return err
	}
	// Include every sidecar ID, including run/report IDs covered by exact FKs.
	for _, t := range d.Tables {
		for _, c := range t.Columns {
			if c.Name == "id" || strings.HasSuffix(c.Name, "_id") {
				match := managementTraitsSchemaParentType.FindStringSubmatch(c.Type)
				if match != nil {
					length, _ := strconv.Atoi(match[1])
					capacity[t.Name+"."+c.Name] = managementTraitsSchemaIDBudget{Bytes: length}
				}
			}
		}
	}
	s.schemaCapacity.Store(&managementTraitsRuntimeCapacity{Columns: managementTraitsSchemaByteBudgets(capacity), JSONBytes: s.maxJSONBytes})
	return nil
}

// Open a separate GORM callback registry over the SAME pool. Session alone
// shares callbacks and would silently guard unrelated legacy handlers.
func managementTraitsSchemaGuardDB(db *gorm.DB, capacity map[string]managementTraitsSchemaIDBudget) (*gorm.DB, error) {
	dialect, ok := db.Dialector.(*mysql.Dialector)
	if !ok {
		return nil, ErrManagementTraitsRuntimeInvalid
	}
	dc := *dialect.Config
	dc.Conn = db.ConnPool
	dc.SkipInitializeWithVersion = true
	guarded, err := gorm.Open(mysql.New(dc), &gorm.Config{NamingStrategy: db.NamingStrategy, Logger: db.Logger, NowFunc: db.NowFunc, SkipDefaultTransaction: db.SkipDefaultTransaction, DisableAutomaticPing: true})
	if err != nil {
		return nil, err
	}
	fixed := managementTraitsSchemaByteBudgets(capacity)
	active := func(tx *gorm.DB) map[string]managementTraitsSchemaIDBudget {
		if caps := managementTraitsSchemaCapabilities(tx); caps != nil {
			return caps.Columns
		}
		return fixed
	}
	check := func(tx *gorm.DB) {
		if c := active(tx); len(c) > 0 {
			managementTraitsSchemaCheckStatement(tx, c)
		}
	}
	query := func(tx *gorm.DB) {
		if len(active(tx)) == 0 || tx.Error != nil {
			return
		}
		callbacks.BuildQuerySQL(tx)
		check(tx)
	}
	loaded := func(tx *gorm.DB) {
		if c := active(tx); len(c) > 0 && tx.Error == nil {
			budget := int(^uint(0) >> 1)
			if caps := managementTraitsSchemaCapabilities(tx); caps != nil {
				budget = caps.JSONBytes
			}
			managementTraitsSchemaCheckRecords(tx, c, budget)
		}
	}
	for _, register := range []func() error{
		func() error { return guarded.Callback().Query().Before("gorm:query").Register("mng:schema_ids", query) },
		func() error { return guarded.Callback().Row().Before("gorm:row").Register("mng:schema_ids", query) },
		func() error {
			return guarded.Callback().Query().After("gorm:query").Register("mng:schema_loaded_ids", loaded)
		},
		func() error {
			return guarded.Callback().Create().Before("gorm:begin_transaction").Register("mng:schema_ids", check)
		},
		func() error {
			return guarded.Callback().Update().Before("gorm:begin_transaction").Register("mng:schema_ids", check)
		},
		func() error {
			return guarded.Callback().Create().Before("gorm:begin_transaction").Register("mng:schema_json_ids", loaded)
		},
		func() error {
			return guarded.Callback().Update().Before("gorm:begin_transaction").Register("mng:schema_json_ids", loaded)
		},
	} {
		if err := register(); err != nil {
			return nil, err
		}
	}
	return guarded, nil
}

type managementTraitsRuntimeCapacity struct {
	Columns   map[string]managementTraitsSchemaIDBudget
	JSONBytes int
}

func managementTraitsSchemaCapabilities(db *gorm.DB) *managementTraitsRuntimeCapacity {
	if db == nil {
		return nil
	}
	if value, ok := db.Get("mng:schema_capacity"); ok {
		if pointer, ok := value.(*atomic.Pointer[managementTraitsRuntimeCapacity]); ok {
			return pointer.Load()
		}
	}
	return nil
}

func managementTraitsSchemaByteBudgets(capacity map[string]managementTraitsSchemaIDBudget) map[string]managementTraitsSchemaIDBudget {
	result := make(map[string]managementTraitsSchemaIDBudget, len(capacity))
	for key, n := range capacity {
		result[key] = n
	}
	for pass := 0; pass < 3; pass++ {
		for parent, children := range managementTraitsSchemaReferences() {
			if parent == "el_candidate.id" || parent == "el_tester.id" {
				continue
			}
			for _, child := range children {
				p, c := result[parent], result[child]
				if p.Bytes > 0 && c.Bytes > 0 {
					if c.Bytes > p.Bytes {
						c.Bytes = p.Bytes
					}
					c.ASCIIOnly = c.ASCIIOnly || p.ASCIIOnly
					result[child] = c
				}
			}
		}
	}
	// Optional draft identifiers inherit only their verified real parent budgets.
	// This does not add the optional table to the mandatory runtime schema.
	if b := result["el_exam.id"]; b.Bytes > 0 {
		result["el_mng_exam_draft.exam_id"] = b
	}
	if b := result["el_repo.id"]; b.Bytes > 0 {
		result["el_mng_exam_draft.repo_id"] = b
	}
	return result
}

func managementTraitsSchemaIDFits(capacity map[string]managementTraitsSchemaIDBudget, key, id string) bool {
	b := capacity[key]
	if b.Bytes <= 0 || !managementTraitsOpaqueID(id) || len(id) > b.Bytes {
		return false
	}
	if b.ASCIIOnly {
		for i := 0; i < len(id); i++ {
			if id[i] >= 0x80 {
				return false
			}
		}
	}
	return true
}

var managementTraitsSchemaSQLParameter = regexp.MustCompile(`(?i)([a-z_][a-z0-9_.]*?)\s*(?:=|IN)\s*\(?\s*\?`)
var managementTraitsSchemaSQLTable = regexp.MustCompile(`(?i)(?:FROM|JOIN)\s+(el_[a-z_]+)(?:\s+(?:AS\s+)?([a-z_][a-z0-9_]*))?`)
var managementTraitsSchemaSQLToken = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*|[()?]`)

type managementTraitsSchemaSQLScope struct {
	table   string
	aliases map[string]string
	parent  *managementTraitsSchemaSQLScope
}

// Preserve byte offsets while excluding literals/comments from placeholder
// counting. Each SELECT owns its FROM/JOIN aliases; parentheses alone do not
// create a new query scope, and correlated qualified references may use parents.
func managementTraitsSchemaSQLBindings(text, table string, aliases map[string]string) (string, map[int]int, map[int]*managementTraitsSchemaSQLScope) {
	masked := []byte(text)
	for i := 0; i < len(masked); {
		start, quote := i, masked[i]
		switch {
		case quote == '\'' || quote == '"':
			i++
			for i < len(masked) {
				if masked[i] == '\\' {
					i += 2
					continue
				}
				if masked[i] == quote {
					i++
					if i < len(masked) && masked[i] == quote {
						i++
						continue
					}
					break
				}
				i++
			}
		case quote == '/' && i+1 < len(masked) && masked[i+1] == '*':
			i += 2
			for i < len(masked) && !(masked[i-1] == '*' && masked[i] == '/') {
				i++
			}
			if i < len(masked) {
				i++
			}
		case quote == '#' || (quote == '-' && i+2 < len(masked) && masked[i+1] == '-' && masked[i+2] <= ' '):
			for i < len(masked) && masked[i] != '\n' {
				i++
			}
		default:
			i++
			continue
		}
		for j := start; j < i && j < len(masked); j++ {
			masked[j] = ' '
		}
	}
	text = string(masked)
	root := &managementTraitsSchemaSQLScope{table: table, aliases: aliases}
	stack := []*managementTraitsSchemaSQLScope{root}
	bindings, scopes := make(map[int]int), make(map[int]*managementTraitsSchemaSQLScope)
	for _, at := range managementTraitsSchemaSQLToken.FindAllStringIndex(text, -1) {
		current := stack[len(stack)-1]
		switch strings.ToUpper(text[at[0]:at[1]]) {
		case "(":
			stack = append(stack, current)
		case ")":
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		case "SELECT":
			if len(stack) > 1 {
				stack[len(stack)-1] = &managementTraitsSchemaSQLScope{aliases: make(map[string]string), parent: current}
			}
		case "FROM", "JOIN":
			m := managementTraitsSchemaSQLTable.FindStringSubmatchIndex(text[at[0]:])
			if m == nil || m[0] != 0 {
				continue
			}
			name := text[at[0]+m[2] : at[0]+m[3]]
			if current.table == "" {
				current.table = name
			}
			current.aliases[name] = name
			if m[4] >= 0 {
				current.aliases[text[at[0]+m[4]:at[0]+m[5]]] = name
			}
		case "?":
			bindings[at[0]] = len(bindings)
			scopes[at[0]] = current
		}
	}
	return text, bindings, scopes
}

func managementTraitsSchemaCheckStatement(tx *gorm.DB, capacity map[string]managementTraitsSchemaIDBudget) {
	if tx.Error != nil {
		return
	}
	stmt := tx.Statement
	table := strings.Fields(stmt.Table)
	if len(table) == 0 {
		table = []string{""}
	}
	if stmt.TableExpr != nil {
		parts := strings.Fields(strings.ReplaceAll(stmt.TableExpr.SQL, "`", ""))
		if len(parts) > 0 && strings.HasPrefix(parts[0], "el_") {
			table = parts
		}
	}
	aliases := map[string]string{table[0]: table[0]}
	if len(table) > 1 {
		aliases[table[len(table)-1]] = table[0]
	}
	sql := strings.ReplaceAll(stmt.SQL.String(), "`", "")
	participantKind := ""
	var sqlScope *managementTraitsSchemaSQLScope
	resolveColumn := func(key string) (string, *managementTraitsSchemaSQLScope) {
		key = strings.ReplaceAll(key, "`", "")
		parts := strings.Split(key, ".")
		if sqlScope != nil {
			if len(parts) == 1 {
				return sqlScope.table + "." + key, sqlScope
			}
			for scope := sqlScope; scope != nil; scope = scope.parent {
				if actual, ok := scope.aliases[parts[0]]; ok {
					return actual + "." + parts[len(parts)-1], scope
				}
			}
			return key, sqlScope
		}
		if len(parts) == 1 {
			key = table[0] + "." + key
		} else if actual, ok := aliases[parts[0]]; ok {
			key = actual + "." + parts[len(parts)-1]
		}
		return key, nil
	}
	resolve := func(key string) string {
		key, _ = resolveColumn(key)
		return key
	}
	limit := func(key string) int {
		key = resolve(key)
		parts := strings.Split(key, ".")
		n := capacity[key].Bytes
		if n == 0 && (parts[len(parts)-1] == "id" || strings.HasSuffix(parts[len(parts)-1], "_id")) {
			tx.AddError(ErrManagementTraitsRuntimeInvalid)
		}
		// Sidecar exam/profile columns are 64, but a real exam may be narrower.
		if strings.HasSuffix(key, ".exam_id") || strings.HasSuffix(key, ".profile_exam_id") {
			if p := capacity["el_exam.id"].Bytes; n > 0 && p < n {
				n = p
			}
		}
		if strings.HasSuffix(key, ".participant_id") && (participantKind == "candidate" || participantKind == "tester") {
			if p := capacity["el_"+participantKind+".id"].Bytes; n > 0 && p < n {
				n = p
			}
		}
		return n
	}
	var checkValue func(string, any)
	checkValue = func(key string, v any) {
		n := limit(key)
		if n == 0 || v == nil {
			return
		}
		r := reflect.ValueOf(v)
		for r.IsValid() && (r.Kind() == reflect.Ptr || r.Kind() == reflect.Interface) {
			if r.IsNil() {
				return
			}
			r = r.Elem()
		}
		if !r.IsValid() {
			return
		}
		if r.Kind() == reflect.Slice || r.Kind() == reflect.Array {
			for i := 0; i < r.Len(); i++ {
				checkValue(key, r.Index(i).Interface())
			}
			return
		}
		if r.Kind() != reflect.String {
			tx.AddError(ErrManagementTraitsRuntimeInvalid)
			return
		}
		id := r.String()
		if id != "" && (!managementTraitsSchemaIDFits(capacity, resolve(key), id) || len(id) > n) {
			tx.AddError(ErrManagementTraitsRuntimeInvalid)
		}
	}
	checkSQL := func(text string, vars []any) {
		text, bindings, scopes := managementTraitsSchemaSQLBindings(text, table[0], aliases)
		kinds := make(map[*managementTraitsSchemaSQLScope]map[string]string)
		oldKind, oldScope := participantKind, sqlScope
		defer func() { participantKind, sqlScope = oldKind, oldScope }()
		parameters := managementTraitsSchemaSQLParameter.FindAllStringSubmatchIndex(text, -1)
		// Collect type bindings first: SQL predicate order does not change the
		// owning participant column's parent capacity, including correlations.
		for _, at := range parameters {
			i, ok := bindings[at[1]-1]
			if !ok || i >= len(vars) {
				tx.AddError(ErrManagementTraitsRuntimeInvalid)
				continue
			}
			sqlScope = scopes[at[1]-1]
			key, owner := resolveColumn(text[at[2]:at[3]])
			parts := strings.Split(key, ".")
			if parts[len(parts)-1] == "participant_type" {
				if kinds[owner] == nil {
					kinds[owner] = make(map[string]string)
				}
				kind, _ := vars[i].(string)
				if kind != "candidate" && kind != "tester" {
					kind = "invalid"
				}
				if kinds[owner][parts[0]] == "" {
					kinds[owner][parts[0]] = kind
				}
			}
		}
		for _, at := range parameters {
			i, ok := bindings[at[1]-1]
			if !ok || i >= len(vars) {
				continue // The first pass already rejected the missing binding.
			}
			sqlScope = scopes[at[1]-1]
			rawKey := text[at[2]:at[3]]
			key, owner := resolveColumn(rawKey)
			parts := strings.Split(key, ".")
			if parts[len(parts)-1] == "participant_type" {
				kind, _ := vars[i].(string)
				if kind != "candidate" && kind != "tester" {
					kind = "invalid"
				}
				kinds[owner][parts[0]] = kind
			}
			participantKind = kinds[owner][parts[0]]
			if parts[len(parts)-1] == "participant_id" && participantKind == "invalid" {
				tx.AddError(ErrManagementTraitsRuntimeInvalid)
			}
			// Pass the original column spelling, never resolve a canonical name
			// again as an alias belonging to a different table.
			checkValue(rawKey, vars[i])
			end := at[1]
			for end < len(text) && (text[end] == ',' || text[end] == ' ' || text[end] == '\t' || text[end] == '\n' || text[end] == '\r' || text[end] == '?') {
				if text[end] == '?' {
					i++
					if i >= len(vars) {
						tx.AddError(ErrManagementTraitsRuntimeInvalid)
						break
					}
					checkValue(rawKey, vars[i])
				}
				end++
			}
		}
	}
	checkSQL(sql, stmt.Vars)
	// Dest is a real model/anonymous insert or Updates map; no SQL text rewrite.
	var checkRecord func(reflect.Value)
	checkRecord = func(r reflect.Value) {
		for r.IsValid() && (r.Kind() == reflect.Ptr || r.Kind() == reflect.Interface) {
			if r.IsNil() {
				return
			}
			r = r.Elem()
		}
		if !r.IsValid() {
			return
		}
		switch r.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < r.Len(); i++ {
				checkRecord(r.Index(i))
			}
		case reflect.Map:
			old := participantKind
			defer func() { participantKind = old }()
			if r.Type().Key().Kind() == reflect.String {
				v := r.MapIndex(reflect.ValueOf("participant_type").Convert(r.Type().Key()))
				if v.IsValid() {
					if kind, ok := v.Interface().(string); ok {
						participantKind = kind
					}
				}
			}
			for _, k := range r.MapKeys() {
				if expression, ok := r.MapIndex(k).Interface().(clause.Expr); ok {
					checkSQL(expression.SQL, expression.Vars)
				}
				if k.Kind() == reflect.String && capacity[table[0]+"."+k.String()].Bytes > 0 {
					checkValue(k.String(), r.MapIndex(k).Interface())
				}
			}
		case reflect.Struct:
			if stmt.Schema == nil || r.Type() != stmt.Schema.ModelType {
				return
			}
			old := participantKind
			defer func() { participantKind = old }()
			if f := stmt.Schema.LookUpField("participant_type"); f != nil {
				v, _ := f.ValueOf(stmt.Context, r)
				if kind, ok := v.(string); ok {
					participantKind = kind
				}
			}
			for _, f := range stmt.Schema.Fields {
				if capacity[table[0]+"."+f.DBName].Bytes > 0 {
					value, _ := f.ValueOf(stmt.Context, r)
					checkValue(f.DBName, value)
				}
			}
		}
	}
	// Query Dest is initially zero-valued; only writes need field validation.
	if stmt.SQL.Len() == 0 && (stmt.Dest != nil) {
		checkRecord(reflect.ValueOf(stmt.Dest))
	}
}

// Query callbacks validate actual model/anonymous projections after Scan. Raw
// Scan has no post-scan callback; source and UNION-owner loaders check explicitly.
func managementTraitsSchemaCheckRecords(tx *gorm.DB, capacity map[string]managementTraitsSchemaIDBudget, budget int) {
	if tx.Error != nil || tx.Statement.Dest == nil {
		return
	}
	table := strings.Fields(strings.ReplaceAll(tx.Statement.Table, "`", ""))
	if len(table) == 0 {
		return
	}
	used := 0
	var visit func(reflect.Value)
	visit = func(r reflect.Value) {
		for r.IsValid() && (r.Kind() == reflect.Ptr || r.Kind() == reflect.Interface) {
			if r.IsNil() {
				return
			}
			r = r.Elem()
		}
		if !r.IsValid() {
			return
		}
		if r.Kind() == reflect.Slice || r.Kind() == reflect.Array {
			for i := 0; i < r.Len(); i++ {
				visit(r.Index(i))
			}
			return
		}
		if r.Kind() != reflect.Struct {
			return
		}
		parsed, err := schema.Parse(r.Interface(), &sync.Map{}, tx.NamingStrategy)
		if err != nil {
			tx.AddError(ErrManagementTraitsRuntimeInvalid)
			return
		}
		kind := ""
		for _, name := range []string{"participant_type", "kind"} {
			if f := parsed.LookUpField(name); f != nil {
				value, _ := f.ValueOf(tx.Statement.Context, r)
				if v, ok := value.(string); ok && v != "" {
					kind = v
				}
			}
		}
		for _, f := range parsed.Fields {
			value, _ := f.ValueOf(tx.Statement.Context, r)
			key := table[0] + "." + f.DBName
			if n := capacity[key].Bytes; n > 0 {
				v := reflect.ValueOf(value)
				for v.IsValid() && (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) {
					if v.IsNil() {
						v = reflect.Value{}
						break
					}
					v = v.Elem()
				}
				if !v.IsValid() {
					continue
				}
				if v.Kind() != reflect.String {
					tx.AddError(ErrManagementTraitsRuntimeInvalid)
					continue
				}
				id := v.String()
				if id == "" {
					continue
				}
				if f.DBName == "participant_id" && (kind == "candidate" || kind == "tester") {
					if parent := capacity["el_"+kind+".id"].Bytes; parent < n {
						n = parent
					}
				}
				if !managementTraitsSchemaIDFits(capacity, key, id) || len(id) > n {
					tx.AddError(ErrManagementTraitsRuntimeInvalid)
				}
			}
			text, ok := value.(string)
			if !ok || text == "" {
				continue
			}
			if f.DBName == "mapping_snapshot" || f.DBName == "options_snapshot" || f.DBName == "scoring_manifest" {
				if budget <= 0 || len(text) > budget-used {
					tx.AddError(ErrManagementTraitsRuntimeInvalid)
					continue
				}
				used += len(text)
			}
			checkOption := func(o ManagementTraitsMappedOption) {
				if !managementTraitsSchemaIDFits(capacity, "el_qu_answer.id", o.SourceOptionID) {
					tx.AddError(ErrManagementTraitsRuntimeInvalid)
				}
			}
			switch f.DBName {
			case "mapping_snapshot":
				var mapping managementTraitsCanonicalMapping
				if managementTraitsDecodeStrict([]byte(text), budget, &mapping) != nil {
					tx.AddError(ErrManagementTraitsRuntimeInvalid)
					continue
				}
				for _, q := range mapping.Questions {
					if !managementTraitsSchemaIDFits(capacity, "el_qu.id", q.SourceQuestionID) {
						tx.AddError(ErrManagementTraitsRuntimeInvalid)
					}
					for _, o := range q.Options {
						checkOption(o)
					}
				}
			case "options_snapshot":
				var options []ManagementTraitsMappedOption
				if managementTraitsDecodeStrict([]byte(text), budget, &options) != nil {
					tx.AddError(ErrManagementTraitsRuntimeInvalid)
					continue
				}
				for _, o := range options {
					checkOption(o)
				}
			}
		}
	}
	visit(reflect.ValueOf(tx.Statement.Dest))
}
