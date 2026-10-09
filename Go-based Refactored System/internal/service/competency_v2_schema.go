package service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type phase1V2SchemaColumn struct {
	Table        string `gorm:"column:table_name"`
	Name         string `gorm:"column:name"`
	DataType     string `gorm:"column:data_type"`
	Nullable     string `gorm:"column:nullable"`
	DefaultValue string `gorm:"column:default_value"`
	CharLength   int64  `gorm:"column:char_length"`
	Precision    int64  `gorm:"column:numeric_precision_value"`
	Scale        int64  `gorm:"column:numeric_scale_value"`
	CharacterSet string `gorm:"column:character_set"`
	Collation    string `gorm:"column:collation"`
}

type phase1V2SchemaIndex struct {
	Table   string `gorm:"column:table_name"`
	Name    string `gorm:"column:name"`
	Columns string `gorm:"column:columns"`
}

type phase1V2SchemaForeignKey struct {
	Table           string `gorm:"column:table_name"`
	Name            string `gorm:"column:name"`
	Columns         string `gorm:"column:columns"`
	ReferencedTable string `gorm:"column:referenced_table"`
	Referenced      string `gorm:"column:referenced"`
	UpdateRule      string `gorm:"column:update_rule"`
	DeleteRule      string `gorm:"column:delete_rule"`
}

type phase1V2SchemaSignature struct {
	Tables        []string
	Columns       []phase1V2SchemaColumn
	UniqueIndexes []phase1V2SchemaIndex
	ForeignKeys   []phase1V2SchemaForeignKey
}

var phase1V2ResultRunTables = []string{
	"el_competency_result_run",
	"el_competency_result_run_overall",
	"el_competency_result_run_module",
	"el_competency_result_run_dimension",
	"el_competency_result_run_validity",
}

func phase1V2ExpectedSchemaSignature() phase1V2SchemaSignature {
	columns := map[string][]string{
		"el_competency_result_run": {
			"id", "paper_id", "exam_id", "product_version", "scoring_version", "content_version", "report_template_version",
			"report_audience", "participant_type", "participant_id", "participant_name", "participant_telephone", "participant_age",
			"participant_gender", "participant_affiliation", "participant_post", "participant_degree", "participant_major", "source",
			"status", "error_message", "created_by", "completed_at", "create_time", "update_time",
		},
		"el_competency_result_run_overall": {
			"result_run_id", "total_question_count", "answered_question_count", "dimension_question_count",
			"answered_dimension_question_count", "effective_dimension_count", "overall_score", "level_code", "norm_score",
			"norm_comparison_code", "is_complete", "submit_type", "submitted_at", "user_time", "create_time", "update_time",
		},
		"el_competency_result_run_module": {
			"id", "result_run_id", "module_id", "module_code", "module_name", "display_order", "total_dimension_count",
			"effective_dimension_count", "module_score", "level_code", "norm_score", "norm_comparison_code", "is_complete", "create_time",
		},
		"el_competency_result_run_dimension": {
			"id", "result_run_id", "dimension_id", "dimension_code", "dimension_name", "display_order", "total_question_count",
			"answered_question_count", "score_sum", "dimension_score", "level_code", "norm_score", "is_complete", "create_time",
		},
		"el_competency_result_run_validity": {
			"result_run_id", "total_question_count", "answered_question_count", "validity_score", "validity_status", "is_complete", "create_time", "update_time",
		},
	}
	expected := phase1V2SchemaSignature{Tables: append([]string(nil), phase1V2ResultRunTables...)}
	for _, table := range phase1V2ResultRunTables {
		for _, column := range columns[table] {
			expected.Columns = append(expected.Columns, phase1V2ExpectedColumn(table, column))
		}
	}
	expected.UniqueIndexes = []phase1V2SchemaIndex{
		{Table: "el_competency_result_run", Name: "PRIMARY", Columns: "id"},
		{Table: "el_competency_result_run", Name: "uk_competency_result_run_version", Columns: "paper_id,scoring_version"},
		{Table: "el_competency_result_run_overall", Name: "PRIMARY", Columns: "result_run_id"},
		{Table: "el_competency_result_run_module", Name: "PRIMARY", Columns: "id"},
		{Table: "el_competency_result_run_module", Name: "uk_result_run_module", Columns: "result_run_id,module_code"},
		{Table: "el_competency_result_run_module", Name: "uk_result_run_module_order", Columns: "result_run_id,display_order"},
		{Table: "el_competency_result_run_dimension", Name: "PRIMARY", Columns: "id"},
		{Table: "el_competency_result_run_dimension", Name: "uk_result_run_dimension", Columns: "result_run_id,dimension_id"},
		{Table: "el_competency_result_run_dimension", Name: "uk_result_run_dimension_order", Columns: "result_run_id,display_order"},
		{Table: "el_competency_result_run_validity", Name: "PRIMARY", Columns: "result_run_id"},
	}
	expected.ForeignKeys = []phase1V2SchemaForeignKey{
		{Table: "el_competency_result_run", Name: "fk_result_run_paper", Columns: "paper_id", ReferencedTable: "el_paper", Referenced: "id", UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"},
		{Table: "el_competency_result_run", Name: "fk_result_run_exam", Columns: "exam_id", ReferencedTable: "el_exam", Referenced: "id", UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"},
		{Table: "el_competency_result_run_overall", Name: "fk_result_run_overall", Columns: "result_run_id", ReferencedTable: "el_competency_result_run", Referenced: "id", UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"},
		{Table: "el_competency_result_run_module", Name: "fk_result_run_module", Columns: "result_run_id", ReferencedTable: "el_competency_result_run", Referenced: "id", UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"},
		{Table: "el_competency_result_run_dimension", Name: "fk_result_run_dimension", Columns: "result_run_id", ReferencedTable: "el_competency_result_run", Referenced: "id", UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"},
		{Table: "el_competency_result_run_validity", Name: "fk_result_run_validity", Columns: "result_run_id", ReferencedTable: "el_competency_result_run", Referenced: "id", UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"},
	}
	return expected
}

func phase1V2ExpectedColumn(table, name string) phase1V2SchemaColumn {
	column := phase1V2SchemaColumn{Table: table, Name: name, Nullable: "NO", DefaultValue: "<NULL>", CharLength: -1, Precision: -1, Scale: -1}
	if name == "user_time" {
		column.DataType, column.Precision, column.Scale, column.DefaultValue = "int", 10, 0, "0"
	} else if strings.HasSuffix(name, "_time") || name == "completed_at" || name == "submitted_at" {
		column.DataType = "datetime"
	} else if name == "participant_age" {
		column.DataType, column.Nullable, column.Precision, column.Scale = "int", "YES", 10, 0
	} else if name == "created_by" {
		column.DataType, column.Nullable, column.Precision, column.Scale = "bigint", "YES", 19, 0
	} else if name == "display_order" || strings.HasSuffix(name, "_count") || name == "score_sum" {
		column.DataType, column.Precision, column.Scale, column.DefaultValue = "int", 10, 0, "0"
		if name == "display_order" {
			column.DefaultValue = "<NULL>"
		}
	} else if name == "is_complete" {
		column.DataType, column.Precision, column.Scale, column.DefaultValue = "tinyint", 3, 0, "0"
	} else if strings.HasSuffix(name, "_score") {
		column.DataType, column.Nullable, column.Precision, column.Scale = "decimal", "YES", 18, 6
	} else {
		column.DataType, column.CharLength, column.CharacterSet, column.Collation = "varchar", 64, "utf8mb4", "utf8mb4_general_ci"
		switch name {
		case "product_version", "scoring_version", "content_version", "report_template_version", "report_audience", "module_code", "dimension_code", "level_code", "norm_comparison_code", "validity_status":
			column.CharLength = 32
		case "participant_type", "participant_gender", "submit_type":
			column.CharLength = 16
		case "participant_name", "participant_post", "participant_degree", "participant_major", "module_name", "dimension_name":
			column.CharLength = 100
		case "participant_telephone":
			column.CharLength = 32
		case "participant_affiliation":
			column.CharLength = 255
		case "source":
			column.CharLength = 32
		case "status":
			column.CharLength = 20
		case "error_message":
			column.CharLength = 500
		}
		if name == "participant_age" || name == "created_by" {
			column.Nullable = "YES"
		}
		if name == "level_code" || name == "norm_comparison_code" || name == "validity_status" {
			column.Nullable = "YES"
		}
		if name == "participant_name" || name == "participant_telephone" || name == "participant_gender" || name == "participant_affiliation" || name == "participant_post" || name == "participant_degree" || name == "participant_major" || name == "error_message" || name == "submit_type" {
			column.DefaultValue = ""
		}
	}
	if name == "completed_at" || name == "submitted_at" {
		column.Nullable = "YES"
	}
	return column
}

func validatePhase1V2SchemaSignature(actual phase1V2SchemaSignature) error {
	expected := phase1V2ExpectedSchemaSignature()
	if len(actual.Tables) != len(expected.Tables) || len(actual.Columns) != len(expected.Columns) || len(actual.ForeignKeys) != len(expected.ForeignKeys) {
		return errors.New("phase-1 v2 result-run schema contains missing or extra objects")
	}
	tables := make(map[string]struct{}, len(actual.Tables))
	for _, table := range actual.Tables {
		tables[table] = struct{}{}
	}
	for _, table := range expected.Tables {
		if _, ok := tables[table]; !ok {
			return fmt.Errorf("phase-1 v2 result-run table is missing: %s", table)
		}
	}
	columns := make(map[string]phase1V2SchemaColumn, len(actual.Columns))
	for _, column := range actual.Columns {
		columns[column.Table+"."+column.Name] = column
	}
	for _, column := range expected.Columns {
		actualColumn, ok := columns[column.Table+"."+column.Name]
		if !ok {
			return fmt.Errorf("phase-1 v2 result-run column is missing: %s.%s", column.Table, column.Name)
		}
		if strings.ToLower(actualColumn.DataType) != column.DataType || strings.ToUpper(actualColumn.Nullable) != column.Nullable ||
			actualColumn.DefaultValue != column.DefaultValue || actualColumn.CharLength != column.CharLength ||
			actualColumn.Precision != column.Precision || actualColumn.Scale != column.Scale ||
			(column.DataType == "varchar" && (strings.ToLower(actualColumn.CharacterSet) != column.CharacterSet || strings.TrimSpace(actualColumn.Collation) == "")) {
			return fmt.Errorf("phase-1 v2 result-run column definition mismatch: %s.%s", column.Table, column.Name)
		}
	}
	indexes := make(map[string]phase1V2SchemaIndex, len(actual.UniqueIndexes))
	for _, index := range actual.UniqueIndexes {
		indexes[index.Table+"."+index.Name] = index
	}
	for _, index := range expected.UniqueIndexes {
		actualIndex, ok := indexes[index.Table+"."+index.Name]
		if !ok || actualIndex.Columns != index.Columns {
			return fmt.Errorf("phase-1 v2 result-run unique index mismatch: %s.%s", index.Table, index.Name)
		}
	}
	expectedIndexKeys := make(map[string]struct{}, len(expected.UniqueIndexes))
	for _, index := range expected.UniqueIndexes {
		expectedIndexKeys[index.Table+"."+index.Name] = struct{}{}
	}
	for _, index := range actual.UniqueIndexes {
		key := index.Table + "." + index.Name
		if _, required := expectedIndexKeys[key]; required {
			continue
		}
		if !(index.Table == "el_competency_result_run" && index.Name == "uk_result_run_id_paper" && index.Columns == "id,paper_id") {
			return fmt.Errorf("phase-1 v2 result-run unexpected unique index: %s", key)
		}
	}
	foreignKeys := make(map[string]phase1V2SchemaForeignKey, len(actual.ForeignKeys))
	for _, foreignKey := range actual.ForeignKeys {
		foreignKeys[foreignKey.Table+"."+foreignKey.Name] = foreignKey
	}
	for _, foreignKey := range expected.ForeignKeys {
		actualForeignKey, ok := foreignKeys[foreignKey.Table+"."+foreignKey.Name]
		if !ok || actualForeignKey.Columns != foreignKey.Columns || actualForeignKey.ReferencedTable != foreignKey.ReferencedTable ||
			actualForeignKey.Referenced != foreignKey.Referenced || strings.ToUpper(actualForeignKey.UpdateRule) != foreignKey.UpdateRule ||
			strings.ToUpper(actualForeignKey.DeleteRule) != foreignKey.DeleteRule {
			return fmt.Errorf("phase-1 v2 result-run foreign key mismatch: %s.%s", foreignKey.Table, foreignKey.Name)
		}
	}
	return nil
}

func loadPhase1V2SchemaState(db *gorm.DB) (bool, error) {
	if db == nil {
		return false, errors.New("phase-1 v2 result-run database is unavailable")
	}
	var tableCount int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ?", phase1V2ResultRunTables).Scan(&tableCount).Error; err != nil {
		return false, err
	}
	if tableCount == 0 {
		return false, nil
	}
	if tableCount != int64(len(phase1V2ResultRunTables)) {
		return false, errors.New("phase-1 v2 result-run migration is incomplete")
	}
	signature := phase1V2SchemaSignature{Tables: append([]string(nil), phase1V2ResultRunTables...)}
	if err := db.Raw("SELECT TABLE_NAME AS table_name,COLUMN_NAME AS name,DATA_TYPE AS data_type,IS_NULLABLE AS nullable,COALESCE(CAST(COLUMN_DEFAULT AS CHAR),'<NULL>') AS default_value,COALESCE(CHARACTER_MAXIMUM_LENGTH,-1) AS char_length,COALESCE(NUMERIC_PRECISION,-1) AS numeric_precision_value,COALESCE(NUMERIC_SCALE,-1) AS numeric_scale_value,COALESCE(CHARACTER_SET_NAME,'') AS character_set,COALESCE(COLLATION_NAME,'') AS collation FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ?", phase1V2ResultRunTables).Scan(&signature.Columns).Error; err != nil {
		return false, err
	}
	if err := db.Raw("SELECT TABLE_NAME AS table_name,INDEX_NAME AS name,GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX SEPARATOR ',') AS columns FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ? AND NON_UNIQUE=0 GROUP BY TABLE_NAME,INDEX_NAME", phase1V2ResultRunTables).Scan(&signature.UniqueIndexes).Error; err != nil {
		return false, err
	}
	if err := db.Raw("SELECT rc.TABLE_NAME AS table_name,rc.CONSTRAINT_NAME AS name,GROUP_CONCAT(kcu.COLUMN_NAME ORDER BY kcu.ORDINAL_POSITION SEPARATOR ',') AS columns,kcu.REFERENCED_TABLE_NAME AS referenced_table,GROUP_CONCAT(kcu.REFERENCED_COLUMN_NAME ORDER BY kcu.ORDINAL_POSITION SEPARATOR ',') AS referenced,rc.UPDATE_RULE AS update_rule,rc.DELETE_RULE AS delete_rule FROM information_schema.REFERENTIAL_CONSTRAINTS rc INNER JOIN information_schema.KEY_COLUMN_USAGE kcu ON kcu.CONSTRAINT_SCHEMA=rc.CONSTRAINT_SCHEMA AND kcu.TABLE_NAME=rc.TABLE_NAME AND kcu.CONSTRAINT_NAME=rc.CONSTRAINT_NAME WHERE rc.CONSTRAINT_SCHEMA=DATABASE() AND rc.TABLE_NAME IN ? GROUP BY rc.TABLE_NAME,rc.CONSTRAINT_NAME,kcu.REFERENCED_TABLE_NAME,rc.UPDATE_RULE,rc.DELETE_RULE", phase1V2ResultRunTables).Scan(&signature.ForeignKeys).Error; err != nil {
		return false, err
	}
	if err := validatePhase1V2SchemaSignature(signature); err != nil {
		return false, err
	}
	return true, nil
}
