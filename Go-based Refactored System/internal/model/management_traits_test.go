package model

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/schema"
)

// S2A is a local model contract only: no database, DDL, writer, HTTP,
// activation or rendering. Structs cannot enforce 140/13/4 cardinality,
// source/identity-source allowlists, JSON document validity, hashes or
// immutable scoring data versus captured submission fields. Those remain
// later runtime validation responsibilities; indexes here are GORM metadata,
// not evidence of installed database constraints or foreign keys.
type managementTraitsModelContract struct {
	value   interface{}
	table   string
	primary string
	fields  string
	indexes map[string][]string
}

// Each token is GoName:column:json:type[:databaseType]. Types are string,
// int, bool, time.Time and their specified nullable pointer counterparts.
func managementTraitsModelContracts() []managementTraitsModelContract {
	return []managementTraitsModelContract{
		{ManagementTraitsDefinitionBundle{}, "el_mng_definition_bundle", "ID", `
ID:id:id:s:varchar(64)
ProductVersion:product_version:productVersion:s:varchar(64)
QuestionVersion:question_version:questionVersion:s:varchar(64)
ScoringVersion:scoring_version:scoringVersion:s:varchar(64)
NormVersion:norm_version:normVersion:s:varchar(64)
Questionnaire:questionnaire:questionnaire:s
ScoringManifest:scoring_manifest:scoringManifest:s:longtext
ScoringManifestSHA:scoring_manifest_sha:scoringManifestSha:s:char(64)
Status:status:status:s
CreatedAt:created_at:createdAt:t`, map[string][]string{
			"uk_mng_bundle_versions": {"product_version", "question_version", "scoring_version", "norm_version"},
		}},
		{ManagementTraitsExamProfile{}, "el_mng_exam_profile", "ExamID", `
ExamID:exam_id:examId:s:varchar(64)
BundleID:bundle_id:bundleId:s:varchar(64)
MappingSnapshot:mapping_snapshot:mappingSnapshot:s:longtext
FieldContract:field_contract:fieldContract:s:longtext
MappingSHA:mapping_sha:mappingSha:s:char(64)
TotalTimeMinutes:total_time_minutes:totalTimeMinutes:i
FrozenAt:frozen_at:frozenAt:pt
CreatedAt:created_at:createdAt:t`, map[string][]string{}},
		{ManagementTraitsPaperSnapshot{}, "el_mng_paper_snapshot", "PaperID", `
PaperID:paper_id:paperId:s:varchar(64)
ExamID:exam_id:examId:s:varchar(64)
ProfileExamID:profile_exam_id:profileExamId:ps:varchar(64)
BundleID:bundle_id:bundleId:s:varchar(64)
Source:source:source:s
EvidenceSnapshot:evidence_snapshot:evidenceSnapshot:s:longtext
EvidenceSHA:evidence_sha:evidenceSha:s:char(64)
MappingSnapshot:mapping_snapshot:mappingSnapshot:s:longtext
MappingSHA:mapping_sha:mappingSha:s:char(64)
ScoringManifestSHA:scoring_manifest_sha:scoringManifestSha:s:char(64)
ParticipantType:participant_type:participantType:s
ParticipantID:participant_id:participantId:s:varchar(64)
ParticipantSnapshot:participant_snapshot:participantSnapshot:s:longtext
FieldContract:field_contract:fieldContract:s:longtext
IdentitySource:identity_source:identitySource:s
SourceCapturedAt:source_captured_at:sourceCapturedAt:t
StartedAt:started_at:startedAt:t
LimitTime:limit_time:limitTime:pt
CreatedAt:created_at:createdAt:t`, map[string][]string{
			"uk_mng_snapshot_paper_exam": {"paper_id", "exam_id"},
		}},
		{ManagementTraitsPaperQuestionSnapshot{}, "el_mng_paper_question_snapshot", "ID", `
ID:id:id:s:varchar(64)
PaperID:paper_id:paperId:s:varchar(64)
PaperQuestionID:paper_question_id:paperQuestionId:s:varchar(64)
SourceQuestionID:source_question_id:sourceQuestionId:s:varchar(64)
Number:v_number:number:i
DisplayOrder:display_order:displayOrder:i
DimensionKey:dimension_key:dimensionKey:s
Reverse:reverse:reverse:b
Content:content:content:s:longtext
OptionsSnapshot:options_snapshot:optionsSnapshot:s:longtext
ScoringSnapshotSHA:scoring_snapshot_sha:scoringSnapshotSha:s:char(64)
SelectedOptionID:selected_option_id:selectedOptionId:ps:varchar(64)
RawAnswer:raw_answer:rawAnswer:pi
FinalScore:final_score:finalScore:pi
SubmittedAt:submitted_at:submittedAt:pt
CreatedAt:created_at:createdAt:t`, map[string][]string{
			"uk_mng_question_v":     {"paper_id", "v_number"},
			"uk_mng_question_pq":    {"paper_id", "paper_question_id"},
			"uk_mng_question_order": {"paper_id", "display_order"},
		}},
		{ManagementTraitsResultRun{}, "el_mng_result_run", "ID", `
ID:id:id:s:varchar(64)
PaperID:paper_id:paperId:s:varchar(64)
ExamID:exam_id:examId:s:varchar(64)
ProductVersion:product_version:productVersion:s:varchar(64)
QuestionVersion:question_version:questionVersion:s:varchar(64)
ScoringVersion:scoring_version:scoringVersion:s:varchar(64)
NormVersion:norm_version:normVersion:s:varchar(64)
Questionnaire:questionnaire:questionnaire:s
ParticipantType:participant_type:participantType:s
ParticipantID:participant_id:participantId:s:varchar(64)
ScoringManifestSHA:scoring_manifest_sha:scoringManifestSha:s:char(64)
InputSHA:input_sha:inputSha:s:char(64)
Status:status:status:s
Source:source:source:s
TotalQuestionCount:total_question_count:totalQuestionCount:i
AnsweredQuestionCount:answered_question_count:answeredQuestionCount:i
OverallScore:overall_score:overallScore:pd:decimal(18,6)
OverallNorm:overall_norm:overallNorm:pd:decimal(18,6)
OverallLevel:overall_level:overallLevel:ps
UserTimeSeconds:user_time_seconds:userTimeSeconds:pi
SubmittedAt:submitted_at:submittedAt:pt
CreatedAt:created_at:createdAt:t`, map[string][]string{
			"uk_mng_run_versions": {"paper_id", "scoring_version", "norm_version"},
			"uk_mng_run_identity": {"id", "paper_id", "exam_id"},
		}},
		{ManagementTraitsResultDimension{}, "el_mng_result_dimension", "ID", `
ID:id:id:s:varchar(64)
RunID:run_id:runId:s:varchar(64)
DimensionKey:dimension_key:dimensionKey:s
DisplayOrder:display_order:displayOrder:i
DimensionName:dimension_name:dimensionName:s
ModuleKey:module_key:moduleKey:s
QuestionCount:question_count:questionCount:i
AnsweredCount:answered_count:answeredCount:i
ScoreSum:score_sum:scoreSum:i
Score:score:score:pd:decimal(18,6)
Norm:norm:norm:pd:decimal(18,6)
Level:level:level:ps
CreatedAt:created_at:createdAt:t`, map[string][]string{
			"uk_mng_dimension_key":   {"run_id", "dimension_key"},
			"uk_mng_dimension_order": {"run_id", "display_order"},
		}},
		{ManagementTraitsResultModule{}, "el_mng_result_module", "ID", `
ID:id:id:s:varchar(64)
RunID:run_id:runId:s:varchar(64)
ModuleKey:module_key:moduleKey:s
DisplayOrder:display_order:displayOrder:i
DimensionCount:dimension_count:dimensionCount:i
Score:score:score:pd:decimal(18,6)
CreatedAt:created_at:createdAt:t`, map[string][]string{
			"uk_mng_module_key":   {"run_id", "module_key"},
			"uk_mng_module_order": {"run_id", "display_order"},
		}},
	}
}

func managementTraitsExpectedType(t *testing.T, code string) reflect.Type {
	t.Helper()
	types := map[string]reflect.Type{
		"s": reflect.TypeOf(""), "i": reflect.TypeOf(int(0)), "b": reflect.TypeOf(false),
		"t": reflect.TypeOf(time.Time{}), "ps": reflect.TypeOf((*string)(nil)),
		"pi": reflect.TypeOf((*int)(nil)), "pt": reflect.TypeOf((*time.Time)(nil)),
		"pd": reflect.TypeOf((*decimal.Decimal)(nil)),
	}
	typ, ok := types[code]
	if !ok {
		t.Fatalf("unknown contract type %q", code)
	}
	return typ
}

func managementTraitsParseSchema(t *testing.T, value interface{}) *schema.Schema {
	t.Helper()
	s, err := schema.Parse(value, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse model schema: %v", err)
	}
	return s
}

func TestManagementTraitsTableNamesAndPrimaryKeys(t *testing.T) {
	contracts := managementTraitsModelContracts()
	if len(contracts) != 7 {
		t.Fatalf("model contract count=%d want 7", len(contracts))
	}
	for _, c := range contracts {
		t.Run(reflect.TypeOf(c.value).Name(), func(t *testing.T) {
			namer, ok := c.value.(interface{ TableName() string })
			if !ok || namer.TableName() != c.table {
				t.Fatalf("TableName must return %q (implemented=%v)", c.table, ok)
			}
			s := managementTraitsParseSchema(t, c.value)
			if s.Table != c.table || len(s.PrimaryFields) != 1 || s.PrimaryFields[0].Name != c.primary {
				t.Fatalf("table=%q primary=%v want table=%q sole primary=%s", s.Table, s.PrimaryFieldDBNames, c.table, c.primary)
			}
			if !s.PrimaryFields[0].PrimaryKey || s.PrimaryFields[0].AutoIncrement {
				t.Error("string primary key must not auto-increment")
			}
		})
	}
}

func TestManagementTraitsExactFieldsTagsTypesAndNullableSchema(t *testing.T) {
	for _, c := range managementTraitsModelContracts() {
		t.Run(reflect.TypeOf(c.value).Name(), func(t *testing.T) {
			typ := reflect.TypeOf(c.value)
			s := managementTraitsParseSchema(t, c.value)
			fields := strings.Fields(c.fields)
			if typ.Kind() != reflect.Struct || typ.NumField() != len(fields) || len(s.Fields) != len(fields) {
				t.Fatalf("field counts reflect=%d GORM=%d want=%d; no extra or embedded fields", typ.NumField(), len(s.Fields), len(fields))
			}
			for _, spec := range fields {
				parts := strings.Split(spec, ":")
				name, column, jsonName, code := parts[0], parts[1], parts[2], parts[3]
				field, ok := typ.FieldByName(name)
				if !ok {
					t.Errorf("missing field %s", name)
					continue
				}
				if field.Anonymous || field.PkgPath != "" || field.Type != managementTraitsExpectedType(t, code) {
					t.Errorf("%s must be an exported non-embedded %v, got %v", name, managementTraitsExpectedType(t, code), field.Type)
				}
				if field.Tag.Get("json") != jsonName {
					t.Errorf("%s JSON tag=%q want exact %q without omitempty", name, field.Tag.Get("json"), jsonName)
				}
				gf := s.FieldsByName[name]
				if gf == nil {
					t.Errorf("missing GORM field %s", name)
					continue
				}
				if gf.TagSettings["COLUMN"] != column || gf.DBName != column || gf.IgnoreMigration || !gf.Creatable || !gf.Updatable || !gf.Readable {
					t.Errorf("%s must explicitly map readable/writable column %q", name, column)
				}
				if gf.PrimaryKey != (name == c.primary) {
					t.Errorf("%s primaryKey=%v want=%v", name, gf.PrimaryKey, name == c.primary)
				}
				if strings.HasPrefix(code, "p") && (gf.NotNull || gf.FieldType.Kind() != reflect.Pointer) {
					t.Errorf("%s must retain nullable pointer metadata; NotNull=%v type=%v", name, gf.NotNull, gf.FieldType)
				}
				wantDataType := map[string]string{"s": "string", "ps": "string", "i": "int", "pi": "int", "b": "bool", "t": "time", "pt": "time"}[code]
				if len(parts) == 5 {
					wantDataType = parts[4]
					if strings.ToLower(gf.TagSettings["TYPE"]) != wantDataType {
						t.Errorf("%s explicit database type=%q want=%q", name, gf.TagSettings["TYPE"], wantDataType)
					}
				}
				// DataType is GORM's declared DB type; no dialector or DB is opened.
				if strings.ToLower(string(gf.DataType)) != wantDataType {
					t.Errorf("%s GORM DB DataType=%q want=%q", name, gf.DataType, wantDataType)
				}
			}
			if len(s.Relationships.Relations) != 0 {
				t.Error("plain models must not include GORM relationships; this does not claim database FK enforcement")
			}
		})
	}
}

func TestManagementTraitsOrderedUniqueIndexes(t *testing.T) {
	indexCount := 0
	for _, c := range managementTraitsModelContracts() {
		indexCount += len(c.indexes)
		t.Run(reflect.TypeOf(c.value).Name(), func(t *testing.T) {
			s := managementTraitsParseSchema(t, c.value)
			indexes := s.ParseIndexes()
			if len(indexes) != len(c.indexes) {
				t.Errorf("parsed index count=%d want=%d", len(indexes), len(c.indexes))
			}
			for name, columns := range c.indexes {
				idx, ok := indexes[name]
				if !ok || idx.Name != name || idx.Class != "UNIQUE" {
					t.Errorf("%s must parse as named UNIQUE index; exists=%v class=%q", name, ok, idx.Class)
					continue
				}
				got := make([]string, 0, len(idx.Fields))
				for i, option := range idx.Fields {
					got = append(got, option.DBName)
					if option.Expression != "" || option.Length != 0 || option.Sort != "" || option.Collate != "" {
						t.Errorf("%s must index full plain columns", name)
					}
					priorityFound := false
					for _, tag := range strings.Split(option.Tag.Get("gorm"), ";") {
						keyValue := strings.SplitN(tag, ":", 2)
						if len(keyValue) != 2 || !strings.EqualFold(strings.TrimSpace(keyValue[0]), "uniqueIndex") {
							continue
						}
						indexParts := strings.SplitN(keyValue[1], ",", 2)
						if indexParts[0] == name && len(indexParts) == 2 {
							settings := schema.ParseTagSetting(indexParts[1], ",")
							priorityFound = settings["PRIORITY"] == strconv.Itoa(i+1)
						}
					}
					if !priorityFound {
						t.Errorf("%s column %s must declare explicit priority %d", name, option.DBName, i+1)
					}
				}
				if !reflect.DeepEqual(got, columns) || idx.Where != "" || idx.Type != "" || idx.Option != "" {
					t.Errorf("%s parsed columns=%v want=%v; must be unfiltered full unique index", name, got, columns)
				}
			}
		})
	}
	if indexCount != 11 {
		t.Errorf("specified composite unique indexes=%d want 11", indexCount)
	}
}

func managementTraitsJSON(t *testing.T, value interface{}) map[string]json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal model: %v", err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatalf("decode model JSON: %v", err)
	}
	return object
}

func TestManagementTraitsJSONCompleteKeysAndNilValues(t *testing.T) {
	for _, c := range managementTraitsModelContracts() {
		t.Run(reflect.TypeOf(c.value).Name(), func(t *testing.T) {
			object := managementTraitsJSON(t, c.value)
			fields := strings.Fields(c.fields)
			if len(object) != len(fields) {
				t.Errorf("JSON key count=%d want=%d", len(object), len(fields))
			}
			for _, spec := range fields {
				parts := strings.Split(spec, ":")
				value, ok := object[parts[2]]
				if !ok {
					t.Errorf("missing JSON key %s", parts[2])
				} else if strings.HasPrefix(parts[3], "p") && string(value) != "null" {
					t.Errorf("nil %s JSON=%s want null", parts[0], value)
				}
			}
		})
	}
}

func TestManagementTraitsJSONExactDecimalNumbersAndRoundTrip(t *testing.T) {
	for _, c := range managementTraitsModelContracts() {
		for _, spec := range strings.Fields(c.fields) {
			parts := strings.Split(spec, ":")
			if parts[3] != "pd" {
				continue
			}
			t.Run(reflect.TypeOf(c.value).Name()+"/"+parts[0], func(t *testing.T) {
				for _, text := range []string{"68.125000", "0.000000", "99.999999"} {
					score, err := decimal.NewFromString(text)
					if err != nil {
						t.Fatal(err)
					}
					value := reflect.New(reflect.TypeOf(c.value))
					value.Elem().FieldByName(parts[0]).Set(reflect.ValueOf(&score))
					object := managementTraitsJSON(t, value.Interface())
					raw := object[parts[2]]
					if len(raw) == 0 || raw[0] == '"' || bytes.Equal(raw, []byte("null")) {
						t.Fatalf("%s JSON must be a number, got %s", parts[0], raw)
					}
					got, err := decimal.NewFromString(string(raw))
					if err != nil || !got.Equal(score) {
						t.Fatalf("JSON=%s want exact %s error=%v", raw, text, err)
					}
					if text == "68.125000" && string(raw) != "68.125" && string(raw) != "68.125000" {
						t.Errorf("decimal sample JSON=%s want 68.125 or 68.125000", raw)
					}
					encoded, err := json.Marshal(value.Interface())
					if err != nil {
						t.Fatal(err)
					}
					decoded := reflect.New(reflect.TypeOf(c.value))
					if err := json.Unmarshal(encoded, decoded.Interface()); err != nil {
						t.Fatal(err)
					}
					result := decoded.Elem().FieldByName(parts[0]).Interface().(*decimal.Decimal)
					if result == nil || !result.Equal(score) {
						t.Errorf("JSON model round trip lost exact decimal %s", text)
					}
				}
			})
		}
	}
}

func TestManagementTraitsNullableZeroAndHistoricalIdentity(t *testing.T) {
	zero := 0
	run := ManagementTraitsResultRun{UserTimeSeconds: &zero}
	if got := string(managementTraitsJSON(t, run)["userTimeSeconds"]); got != "0" {
		t.Errorf("zero duration JSON=%s want 0, not null", got)
	}
	question := ManagementTraitsPaperQuestionSnapshot{RawAnswer: &zero, FinalScore: &zero}
	for _, key := range []string{"rawAnswer", "finalScore"} {
		if got := string(managementTraitsJSON(t, question)[key]); got != "0" {
			t.Errorf("captured zero %s JSON=%s want 0", key, got)
		}
	}
	for _, source := range []string{"submitted_snapshot", "captured_at_recompute"} {
		snapshot := ManagementTraitsPaperSnapshot{ProfileExamID: nil, IdentitySource: source}
		object := managementTraitsJSON(t, snapshot)
		if string(object["profileExamId"]) != "null" {
			t.Error("historical paper must preserve profileExamId:null")
		}
		encoded, err := json.Marshal(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		var restored ManagementTraitsPaperSnapshot
		if err := json.Unmarshal(encoded, &restored); err != nil {
			t.Fatal(err)
		}
		if restored.ProfileExamID != nil || restored.IdentitySource != source {
			t.Errorf("historical identity changed on JSON round trip: %+v", restored)
		}
	}
}

func TestManagementTraitsDecimalDriverValueScanWithoutDB(t *testing.T) {
	for _, c := range managementTraitsModelContracts() {
		for _, spec := range strings.Fields(c.fields) {
			parts := strings.Split(spec, ":")
			if parts[3] != "pd" {
				continue
			}
			t.Run(reflect.TypeOf(c.value).Name()+"/"+parts[0], func(t *testing.T) {
				for _, text := range []string{"68.125000", "0.000000", "99.999999"} {
					original, err := decimal.NewFromString(text)
					if err != nil {
						t.Fatal(err)
					}
					modelValue := reflect.New(reflect.TypeOf(c.value))
					modelValue.Elem().FieldByName(parts[0]).Set(reflect.ValueOf(&original))
					valuer, ok := modelValue.Elem().FieldByName(parts[0]).Interface().(driver.Valuer)
					if !ok {
						t.Fatal("decimal model field must implement driver.Valuer")
					}
					stored, err := valuer.Value()
					if err != nil {
						t.Fatal(err)
					}
					storedText, ok := stored.(string)
					if !ok {
						t.Fatalf("driver value type=%T want exact string, never float", stored)
					}
					exact, err := decimal.NewFromString(storedText)
					if err != nil || !exact.Equal(original) {
						t.Fatalf("driver value=%q lost %s error=%v", storedText, text, err)
					}
					for _, input := range []interface{}{stored, []byte(storedText)} {
						restored := reflect.New(managementTraitsExpectedType(t, "pd").Elem())
						scanner, ok := restored.Interface().(sql.Scanner)
						if !ok {
							t.Fatal("decimal model field must implement sql.Scanner")
						}
						if err := scanner.Scan(input); err != nil {
							t.Fatal(err)
						}
						if !restored.Interface().(*decimal.Decimal).Equal(original) {
							t.Errorf("Scan(%T) lost exact %s", input, text)
						}
					}
				}
			})
		}
	}
}

func TestManagementTraitsNoRenderingAxesRelationsOrFloatFields(t *testing.T) {
	for _, c := range managementTraitsModelContracts() {
		typ := reflect.TypeOf(c.value)
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			underlying := field.Type
			if underlying.Kind() == reflect.Pointer {
				underlying = underlying.Elem()
			}
			if field.Anonymous || underlying.Kind() == reflect.Float32 || underlying.Kind() == reflect.Float64 || underlying.Kind() == reflect.Slice || underlying.Kind() == reflect.Map {
				t.Errorf("%s.%s must not embed, use float or add relationship collections", typ.Name(), field.Name)
			}
			lower := strings.ToLower(field.Name)
			if typ.Name() == "ManagementTraitsDefinitionBundle" || typ.Name() == "ManagementTraitsResultRun" {
				for _, axis := range []string{"content", "template", "report", "pdf", "render"} {
					if strings.Contains(lower, axis) {
						t.Errorf("%s.%s mixes a rendering axis into scoring identity", typ.Name(), field.Name)
					}
				}
			}
			if typ.Name() == "ManagementTraitsPaperSnapshot" && lower == "userid" {
				t.Error("paper snapshot must use participant identity, not legacy UserID")
			}
			if typ.Name() == "ManagementTraitsResultModule" {
				for _, unresolved := range []string{"level", "compare", "comparison", "norm"} {
					if strings.Contains(lower, unresolved) {
						t.Errorf("module must not predefine unresolved %s field", field.Name)
					}
				}
			}
		}
	}
}

func TestManagementTraitsPlainModelsHaveOnlyLiteralTableNameMethods(t *testing.T) {
	wants := make(map[string]string)
	for _, c := range managementTraitsModelContracts() {
		wants[reflect.TypeOf(c.value).Name()] = c.table
		typ := reflect.PointerTo(reflect.TypeOf(c.value))
		if typ.NumMethod() != 1 || typ.Method(0).Name != "TableName" {
			t.Errorf("%s must have only TableName, no hooks or business methods", typ)
		}
	}
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]int)
	for _, entry := range files {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ownsModel := false
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typ := spec.(*ast.TypeSpec)
				if _, ok := wants[typ.Name.Name]; ok {
					ownsModel = true
					seen[typ.Name.Name]++
					if _, ok := typ.Type.(*ast.StructType); !ok || typ.Assign.IsValid() {
						t.Errorf("%s must be its own struct, not an alias", typ.Name.Name)
					}
				}
			}
		}
		if !ownsModel {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv == nil {
				t.Errorf("model-owning file %s must not contain init or business function %s", entry.Name(), fn.Name.Name)
				continue
			}
			receiver := fn.Recv.List[0].Type
			if pointer, ok := receiver.(*ast.StarExpr); ok {
				receiver = pointer.X
			}
			id, ok := receiver.(*ast.Ident)
			if !ok {
				continue
			}
			table, targeted := wants[id.Name]
			if !targeted {
				continue
			}
			if fn.Name.Name != "TableName" || fn.Body == nil || len(fn.Body.List) != 1 {
				t.Errorf("%s must only have a literal TableName return, no hooks/business logic", id.Name)
				continue
			}
			ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
			if !ok || len(ret.Results) != 1 {
				t.Errorf("%s.TableName must return one literal", id.Name)
				continue
			}
			literal, ok := ret.Results[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				t.Errorf("%s.TableName must return string literal", id.Name)
				continue
			}
			got, err := strconv.Unquote(literal.Value)
			if err != nil || got != table {
				t.Errorf("%s.TableName literal=%q want=%q error=%v", id.Name, got, table, err)
			}
		}
	}
	for name := range wants {
		if seen[name] != 1 {
			t.Errorf("%s concrete struct declarations=%d want 1", name, seen[name])
		}
	}
}
