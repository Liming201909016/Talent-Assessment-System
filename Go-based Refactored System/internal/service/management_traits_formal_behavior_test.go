package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

// A synthetic structural fixture, NOT a customer template or an approval seed.
func formalSyntheticTemplate(t *testing.T) []byte {
	t.Helper()
	const w = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	const c = "http://schemas.openxmlformats.org/drawingml/2006/chart"
	const wp = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	const rel = "http://schemas.openxmlformats.org/package/2006/relationships"
	keys := make([]string, 0)
	for key := range formalRequiredTags() {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	doc := `<w:document xmlns:w="` + w + `" xmlns:c="` + c + `" xmlns:wp="` + wp + `" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:body>`
	for _, key := range keys {
		doc += `<w:sdt><w:sdtPr><w:tag w:val="` + key + `"/></w:sdtPr><w:sdtContent><w:p><w:r><w:t>synthetic</w:t></w:r></w:p></w:sdtContent></w:sdt>`
	}
	parts := map[string]string{"[Content_Types].xml": `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`, "_rels/.rels": `<Relationships xmlns="` + rel + `"><Relationship Id="r1" Type="document" Target="word/document.xml"/></Relationships>`}
	rels := `<Relationships xmlns="` + rel + `">`
	chartKeys := []string{"chart.overall", "chart.module.self", "chart.module.interpersonal", "chart.module.task", "chart.module.development", "chart.dimension.comparison"}
	for i, key := range chartKeys {
		id := fmt.Sprintf("r%d", i+1)
		doc += `<wp:anchor><wp:docPr title="` + key + `"/><c:chart r:id="` + id + `"/></wp:anchor>`
		if i < 5 {
			doc += `<wp:anchor><wp:docPr title="` + key + `.numeric-label"/><w:t>score</w:t><w:t>50</w:t><w:t>points</w:t></wp:anchor>`
		}
		rels += `<Relationship Id="` + id + `" Type="chart" Target="charts/chart` + fmt.Sprint(i+1) + `.xml"/>`
		series, points := 1, 2
		if i == 5 {
			series, points = 2, 13
		}
		chart := `<c:chartSpace xmlns:c="` + c + `">`
		for j := 0; j < series; j++ {
			chart += `<c:ser><c:val><c:numLit>`
			for k := 0; k < points; k++ {
				chart += fmt.Sprintf(`<c:pt idx="%d"><c:v>50</c:v></c:pt>`, k)
			}
			chart += `</c:numLit></c:val></c:ser>`
		}
		parts[fmt.Sprintf("word/charts/chart%d.xml", i+1)] = chart + `</c:chartSpace>`
	}
	parts["word/document.xml"] = doc + `</w:body></w:document>`
	parts["word/_rels/document.xml.rels"] = rels + `</Relationships>`
	var out bytes.Buffer
	z := zip.NewWriter(&out)
	for name, raw := range parts {
		f, err := z.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Write([]byte(raw)); err != nil {
			t.Fatal(err)
		}
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func formalRealAssetsFixture(t *testing.T, v *model.ManagementTraitsFormalVersion) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, v.AssetKey)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	workbook, err := os.ReadFile(filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-content-v1.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	for name, b := range map[string][]byte{"content.xlsx": workbook, "template.docx": formalSyntheticTemplate(t)} {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	a, err := loadManagementTraitsFormalAssets(root, v.AssetKey)
	if err != nil || a.templateSHA == "" {
		t.Fatal("synthetic assets rejected", err)
	}
	v.ContentSHA, v.WorkbookSHA, v.TemplateSHA, v.BindingSHA = a.contentSHA, a.workbookSHA, a.templateSHA, a.bindingSHA
	v.IdentitySHA = formalVersionIdentity(*v)
	return root
}

func formalSchemaMetadata(t *testing.T) managementTraitsSchemaMetadata {
	t.Helper()
	d, err := managementTraitsFormalSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	var m managementTraitsSchemaMetadata
	for _, table := range d.Tables {
		for _, col := range table.Columns {
			row := managementTraitsSchemaColumnRow{Table: table.Name, Name: col.Name, ColumnType: col.Type, Nullable: "NO"}
			if strings.Contains(col.Type, "char") || col.Type == "longtext" {
				row.Charset, row.Collation = "utf8mb4", "utf8mb4_bin"
			}
			m.Columns = append(m.Columns, row)
		}
		for _, idx := range table.Indexes {
			for i, column := range idx.Columns {
				m.Indexes = append(m.Indexes, managementTraitsSchemaIndexRow{Table: table.Name, Name: idx.Name, NonUnique: idx.NonUnique, Position: i + 1, Column: column})
			}
		}
	}
	for _, fk := range d.ForeignKeys {
		m.ForeignKeys = append(m.ForeignKeys, managementTraitsSchemaFKRow{Table: fk.Table, Name: fk.Name, Column: "version_id", Position: 1, ReferencedTable: fk.RefTable, ReferencedColumn: "id", UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"})
	}
	return m
}

func formalExpectSchema(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	m := formalSchemaMetadata(t)
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}).AddRow("el_mng_formal_version", "InnoDB").AddRow("el_mng_formal_approval", "InnoDB").AddRow("el_mng_formal_audit", "InnoDB"))
	cols := make([]any, 0)
	for _, x := range m.Columns {
		cols = append(cols, x)
	}
	indexes := make([]any, 0)
	for _, x := range m.Indexes {
		indexes = append(indexes, x)
	}
	fks := make([]any, 0)
	for _, x := range m.ForeignKeys {
		fks = append(fks, x)
	}
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name").WillReturnRows(managementRuntimeModelRows(t, cols...))
	mock.ExpectQuery("SELECT table_name AS table_name, index_name AS index_name").WillReturnRows(managementRuntimeModelRows(t, indexes...))
	mock.ExpectQuery("SELECT k.table_name AS table_name").WillReturnRows(managementRuntimeModelRows(t, fks...))
	mock.ExpectQuery("SELECT 1 AS invalid FROM el_mng_formal_approval").WillReturnRows(sqlmock.NewRows([]string{"invalid"}))
}

func TestManagementTraitsFormalPublicSameActorDualApproval(t *testing.T) {
	v, _ := formalRegistryFixture(t)
	root := formalRealAssetsFixture(t, &v)
	f, _, _, _, _ := managementRuntimeLoadFixture(t, false)
	db, mock := managementRuntimeDB(t)
	s := &ManagementTraitsFormalRegistry{db: db, environment: "local", assetRoot: root}
	approvals := make([]model.ManagementTraitsFormalApproval, 0, 2)
	for _, kind := range []string{"content", "psychometrics"} {
		formalExpectSchema(t, mock)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .*el_mng_formal_version.*FOR UPDATE").WithArgs(v.ID, 1).WillReturnRows(managementRuntimeModelRows(t, v))
		rows := sqlmock.NewRows([]string{"id"})
		if len(approvals) > 0 {
			rows = managementRuntimeModelRows(t, approvals[0])
		}
		mock.ExpectQuery("SELECT .*el_mng_formal_approval").WillReturnRows(rows)
		mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*FOR SHARE").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
		mock.ExpectExec("INSERT INTO `el_mng_formal_approval`").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO `el_mng_formal_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		got, err := s.Change(context.Background(), v.ID, v.IdentitySHA, 1, "approve", kind, "", 7)
		if err != nil {
			t.Fatal("real file/public approval rejected", err)
		}
		approvals = got.Approvals
		if got.CanGenerate || got.Version.State != "draft" {
			t.Fatal("approval auto-activated")
		}
	}
	if len(approvals) != 2 || approvals[0].ActorID != 7 || approvals[1].ActorID != 7 || approvals[0].Kind == approvals[1].Kind || approvals[0].ID == approvals[1].ID {
		t.Fatal("independent same-actor signatures missing")
	}
	for _, action := range []string{"activate", "revoke"} {
		formalExpectSchema(t, mock)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .*el_mng_formal_version.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, v))
		mock.ExpectQuery("SELECT .*el_mng_formal_approval").WillReturnRows(managementRuntimeModelRows(t, approvals[0], approvals[1]))
		if action == "activate" {
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*FOR SHARE").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
		}
		mock.ExpectExec("UPDATE `el_mng_formal_version`").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO `el_mng_formal_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		got, err := s.Change(context.Background(), v.ID, v.IdentitySHA, v.Epoch, action, "", "withdrawn", 7)
		if err != nil {
			t.Fatal("public transition rejected", err)
		}
		if action == "activate" && (!got.VersionReady || got.CanGenerate) {
			t.Fatal("version activation or pipeline capability wrong")
		}
		if action == "revoke" && (got.VersionReady || !got.RevokedHistoricalAdminReadAllowed || got.BlockedReason != "version_revoked") {
			t.Fatal("revocation policy wrong")
		}
		if !got.Version.CreatedAt.Equal(v.CreatedAt) || got.Version.IdentitySHA != v.IdentitySHA {
			t.Fatal("activation overwrote immutable identity/create time")
		}
		v = got.Version
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsFormalPublicRegisterDraft(t *testing.T) {
	v, _ := formalRegistryFixture(t)
	root := formalRealAssetsFixture(t, &v)
	f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
	db, mock := managementRuntimeDB(t)
	s := &ManagementTraitsFormalRegistry{db: db, environment: "local", assetRoot: root}
	formalExpectSchema(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_mng_exam_profile.*FOR SHARE").WillReturnRows(managementRuntimeModelRows(t, profile))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*FOR SHARE").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectExec("INSERT INTO `el_mng_formal_version`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `el_mng_formal_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	got, err := s.Register(context.Background(), v.VersionCode, profile.ExamID, v.AssetKey, 7)
	if err != nil || got.Version.State != "draft" || got.Version.Epoch != 1 || got.Version.CreatedBy != 7 || got.VersionReady || len(got.Approvals) != 0 || got.Approvals == nil || got.BlockedReason != "dual_approval_required" {
		t.Fatal("public registration rejected or auto approved", err)
	}
	if got.Version.SourceSnapshot != formalSourceSnapshot(f.Bundle, profile) || got.Version.TemplateSHA != v.TemplateSHA {
		t.Fatal("registered wrong immutable source/assets")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsFormalRepeatApprovalZeroWrites(t *testing.T) {
	v, _ := formalRegistryFixture(t)
	root := formalRealAssetsFixture(t, &v)
	f, _, _, _, _ := managementRuntimeLoadFixture(t, false)
	approval := model.ManagementTraitsFormalApproval{ID: "original-signature", VersionID: v.ID, IdentitySHA: v.IdentitySHA, Kind: "content", ActorID: 7, CreatedAt: v.CreatedAt}
	db, mock := managementRuntimeDB(t)
	s := &ManagementTraitsFormalRegistry{db: db, environment: "local", assetRoot: root}
	formalExpectSchema(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*el_mng_formal_version.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, v))
	mock.ExpectQuery("SELECT .*el_mng_formal_approval").WillReturnRows(managementRuntimeModelRows(t, approval))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*FOR SHARE").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectCommit()
	got, err := s.Change(context.Background(), v.ID, v.IdentitySHA, 1, "approve", "content", "", 8)
	if err != nil || len(got.Approvals) != 1 || got.Approvals[0].ID != approval.ID || got.Approvals[0].ActorID != 7 {
		t.Fatal("duplicate approval replaced original proof", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsFormalGuardConfigurationClassification(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("el_mng_formal_version").AddRow("el_mng_formal_approval").AddRow("el_mng_formal_audit"))
	formalExpectSchema(t, mock)
	protected, err := CheckManagementTraitsLegacyScope(context.Background(), db, ManagementTraitsLegacyScopeRequest{AllLegacy: true})
	if err != nil || protected {
		t.Fatal("configuration registry broke unrelated legacy reads", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

	db2, m2 := managementRuntimeDB(t)
	m2.ExpectQuery("SELECT table_name AS table_name FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("el_mng_formal_version"))
	if _, err = CheckManagementTraitsLegacyScope(context.Background(), db2, ManagementTraitsLegacyScopeRequest{AllLegacy: true}); err == nil {
		t.Fatal("partial configuration installation ignored")
	}
	if err := m2.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsFormalSchemaDrift(t *testing.T) {
	d, _ := managementTraitsFormalSchemaContract()
	for _, kind := range []string{"valid", "missing_column", "nullable", "type", "collation", "missing_index", "index_order", "fk_rule"} {
		t.Run(kind, func(t *testing.T) {
			m := formalSchemaMetadata(t)
			switch kind {
			case "missing_column":
				m.Columns = m.Columns[1:]
			case "nullable":
				m.Columns[0].Nullable = "YES"
			case "type":
				m.Columns[0].ColumnType = "varchar(32)"
			case "collation":
				m.Columns[0].Collation = "utf8mb4_general_ci"
			case "missing_index":
				m.Indexes = m.Indexes[1:]
			case "index_order":
				m.Indexes[0].Position = 2
			case "fk_rule":
				m.ForeignKeys[0].DeleteRule = "CASCADE"
			}
			err := validateManagementTraitsFormalSchema(d, m)
			if (err == nil) != (kind == "valid") {
				t.Fatal("schema gate mismatch", err)
			}
		})
	}
}

func TestManagementTraitsFormalSchemaAbsentAndEnvironmentClosed(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
	if checkManagementTraitsFormalSchema(context.Background(), db) != ErrManagementTraitsFormalClosed {
		t.Fatal("missing schema not closed")
	}
	for _, env := range []string{"", "production", "staging", "LOCAL"} {
		s := &ManagementTraitsFormalRegistry{db: db, environment: env}
		rows, err := s.List(context.Background())
		if err != ErrManagementTraitsFormalClosed || rows == nil {
			t.Fatal("environment not fail-closed")
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
