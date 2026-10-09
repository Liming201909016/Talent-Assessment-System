package service

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm/schema"
)

var ErrManagementTraitsReissueClosed = errors.New("报告服务未安装或结构无效")
var ErrManagementTraitsReissueInvalid = errors.New("报告来源或归档校验失败")

func managementTraitsReissueContract() (managementTraitsSchemaDefinition, error) {
	d := managementTraitsSchemaDefinition{}
	for _, v := range []any{model.ManagementTraitsReportReissue{}, model.ManagementTraitsReissueAudit{}} {
		s, err := schema.Parse(v, &sync.Map{}, schema.NamingStrategy{}); if err != nil { return d, err }
		t := managementTraitsSchemaTable{Name:s.Table, Indexes:[]managementTraitsSchemaIndex{{Name:"PRIMARY", Columns:[]string{"id"}}}}
		for _, f := range s.Fields { t.Columns = append(t.Columns, managementTraitsSchemaColumn{Name:f.DBName, Type:strings.ToLower(f.TagSettings["TYPE"])}) }
		for _, idx := range s.ParseIndexes() {
			i := managementTraitsSchemaIndex{Name:idx.Name}; if idx.Class != "UNIQUE" { i.NonUnique = 1 }
			for _, f := range idx.Fields { i.Columns = append(i.Columns, f.DBName) }; t.Indexes = append(t.Indexes, i)
		}
		// MySQL's supporting FK indexes are explicit, never inferred by name.
		if t.Name == "el_mng_report_reissue" { t.Indexes = append(t.Indexes, managementTraitsSchemaIndex{Name:"idx_mng_reissue_run", Columns:[]string{"run_id","paper_id","exam_id"}, NonUnique:1}) } else { t.Indexes = append(t.Indexes, managementTraitsSchemaIndex{Name:"idx_mng_reissue_audit_identity", Columns:[]string{"report_id","paper_id","exam_id"}, NonUnique:1}) }
		d.Tables = append(d.Tables,t)
	}
	d.ForeignKeys = []managementTraitsSchemaFK{
		{Table:"el_mng_report_reissue",Name:"fk_mng_reissue_run",RefTable:"el_mng_result_run",Columns:[]string{"run_id","paper_id","exam_id"},RefColumns:[]string{"id","paper_id","exam_id"}},
		{Table:"el_mng_reissue_audit",Name:"fk_mng_reissue_audit",RefTable:"el_mng_report_reissue",Columns:[]string{"report_id","paper_id","exam_id"},RefColumns:[]string{"id","paper_id","exam_id"}},
	}
	return d,nil
}

// Optional capability: no AutoMigrate, no startup gate and no cached absence.
func (s *ManagementTraitsRuntimeService) checkReissueSchema(ctx context.Context) error {
	if s == nil || s.db == nil || ctx == nil { return ErrManagementTraitsReissueClosed }
	names := []string{"el_mng_report_reissue","el_mng_reissue_audit"}
	db := s.db.WithContext(ctx)
	var tables []managementTraitsSchemaTableRow
	if db.Raw("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name IN ?",names).Scan(&tables).Error != nil || len(tables)!=2 { return ErrManagementTraitsReissueClosed }
	seen := map[string]bool{}; for _,t:= range tables { if (t.Name!=names[0]&&t.Name!=names[1])||seen[t.Name]||t.Engine!="InnoDB" { return ErrManagementTraitsReissueClosed }; seen[t.Name]=true }
	var m managementTraitsSchemaMetadata
	if db.Raw("SELECT table_name AS table_name, column_name AS column_name, column_type AS column_type, is_nullable AS is_nullable, COALESCE(character_set_name,'') AS character_set_name, COALESCE(collation_name,'') AS collation_name FROM information_schema.columns WHERE table_schema=DATABASE() AND (table_name IN ? OR (table_name='el_mng_result_run' AND column_name IN ('id','paper_id','exam_id'))) ORDER BY table_name,ordinal_position",names).Scan(&m.Columns).Error != nil { return ErrManagementTraitsReissueClosed }
	if db.Raw("SELECT table_name AS table_name,index_name AS index_name,non_unique AS non_unique,seq_in_index AS seq_in_index,column_name AS column_name,COALESCE(sub_part,0) AS prefix_length FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name IN ? ORDER BY table_name,index_name,seq_in_index",names).Scan(&m.Indexes).Error != nil { return ErrManagementTraitsReissueClosed }
	if db.Raw("SELECT k.table_name AS table_name,k.constraint_name AS constraint_name,k.column_name AS column_name,k.ordinal_position AS ordinal_position,CASE WHEN k.referenced_table_schema=DATABASE() THEN k.referenced_table_name ELSE '' END AS referenced_table_name,k.referenced_column_name AS referenced_column_name,r.update_rule AS update_rule,r.delete_rule AS delete_rule FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.constraint_schema=DATABASE() AND k.table_name IN ? ORDER BY k.table_name,k.constraint_name,k.ordinal_position",names).Scan(&m.ForeignKeys).Error != nil { return ErrManagementTraitsReissueClosed }
	d,err:=managementTraitsReissueContract(); if err!=nil { return ErrManagementTraitsReissueClosed }
	return validateManagementTraitsReissueSchema(d,m)
}

func validateManagementTraitsReissueSchema(d managementTraitsSchemaDefinition,m managementTraitsSchemaMetadata) error {
	fail:=ErrManagementTraitsReissueClosed
	cols:=map[string]managementTraitsSchemaColumnRow{}; for _,c:=range m.Columns { key:=c.Table+"."+c.Name; if _,ok:=cols[key];ok {return fail}; cols[key]=c }
	for _,key:=range []string{"id","paper_id","exam_id"} { p,ok:=cols["el_mng_result_run."+key]; if !ok||p.ColumnType!="varchar(64)"||p.Nullable!="NO"||p.Charset!="utf8mb4"||!strings.HasPrefix(p.Collation,"utf8mb4_"){return fail} }
	count:=3
	for _,t:=range d.Tables { for _,c:=range t.Columns {
		count++; r,ok:=cols[t.Name+"."+c.Name]; if !ok||managementTraitsSchemaType(r.ColumnType)!=c.Type||r.Nullable!="NO" {return fail}
		if strings.Contains(c.Type,"char")||c.Type=="longtext" {
			coll:="utf8mb4_bin"
			parent:=""; if c.Name=="paper_id"||c.Name=="exam_id"{parent=c.Name}; if c.Name=="run_id"{parent="id"}
			if parent!=""{coll=cols["el_mng_result_run."+parent].Collation}
			if r.Charset!="utf8mb4"||r.Collation!=coll{return fail}
		} else if r.Charset!=""||r.Collation!="" {return fail}
	} }
	if len(cols)!=count {return fail}
	indexes:=map[string][]managementTraitsSchemaIndexRow{}; for _,r:=range m.Indexes{indexes[r.Table+"."+r.Name]=append(indexes[r.Table+"."+r.Name],r)}
	n:=0; for _,t:=range d.Tables{for _,i:=range t.Indexes{n++;rows:=indexes[t.Name+"."+i.Name];if len(rows)!=len(i.Columns){return fail}; used:=map[int]bool{};for _,r:=range rows{if r.Position<1||r.Position>len(i.Columns)||used[r.Position]||r.Column!=i.Columns[r.Position-1]||r.NonUnique!=i.NonUnique||r.Prefix!=0{return fail};used[r.Position]=true}}}; if len(indexes)!=n{return fail}
	if len(m.ForeignKeys)!=6{return fail}; used:=map[string]bool{}
	for _,f:=range d.ForeignKeys{for i,c:=range f.Columns{found:=false;for _,r:=range m.ForeignKeys{if r.Table==f.Table&&r.Name==f.Name&&r.Position==i+1{key:=r.Table+"."+r.Name+"."+r.Column;if used[key]||r.Column!=c||r.ReferencedTable!=f.RefTable||r.ReferencedColumn!=f.RefColumns[i]||r.UpdateRule!="RESTRICT"||r.DeleteRule!="RESTRICT"{return fail};used[key]=true;found=true}};if !found{return fail}}}
	return nil
}