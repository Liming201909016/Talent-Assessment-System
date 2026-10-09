package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	driver "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

func TestManagementTraitsReissueAPIClosedWithoutInstalledStorage(t *testing.T) {
	s := NewManagementTraitsRuntimeService(nil, "test-only", 1<<20)
	if _, err := s.ListReportReissues(context.Background(), "paper"); err == nil {
		t.Fatal("uninstalled storage returned success")
	}
}

func reissueTestContent(t *testing.T) ManagementTraitsTestContent {
	t.Helper()
	b, e := os.ReadFile(filepath.Join("..", "..", "..", "docs", "260929管理特质测评-优化", "260928测评内容+数据图.xlsx"))
	if e != nil {
		t.Fatal(e)
	}
	c, e := LoadManagementTraitsTestContent(b)
	if e != nil {
		t.Fatal(e)
	}
	return c
}

func reissueTestSchema(t *testing.T) managementTraitsSchemaMetadata {
	t.Helper()
	d, e := managementTraitsReissueContract()
	if e != nil {
		t.Fatal(e)
	}
	m := managementTraitsSchemaMetadata{}
	for _, n := range []string{"id", "paper_id", "exam_id"} {
		m.Columns = append(m.Columns, managementTraitsSchemaColumnRow{Table: "el_mng_result_run", Name: n, ColumnType: "varchar(64)", Nullable: "NO", Charset: "utf8mb4", Collation: "utf8mb4_0900_ai_ci"})
	}
	for _, tb := range d.Tables {
		for _, c := range tb.Columns {
			r := managementTraitsSchemaColumnRow{Table: tb.Name, Name: c.Name, ColumnType: c.Type, Nullable: "NO"}
			if c.Type == "longtext" || bytes.Contains([]byte(c.Type), []byte("char")) {
				r.Charset = "utf8mb4"
				r.Collation = "utf8mb4_bin"
				if c.Name == "run_id" || c.Name == "paper_id" || c.Name == "exam_id" {
					r.Collation = "utf8mb4_0900_ai_ci"
				}
			}
			m.Columns = append(m.Columns, r)
		}
		for _, i := range tb.Indexes {
			for n, c := range i.Columns {
				m.Indexes = append(m.Indexes, managementTraitsSchemaIndexRow{Table: tb.Name, Name: i.Name, Column: c, Position: n + 1, NonUnique: i.NonUnique})
			}
		}
	}
	for _, f := range d.ForeignKeys {
		for n, c := range f.Columns {
			m.ForeignKeys = append(m.ForeignKeys, managementTraitsSchemaFKRow{Table: f.Table, Name: f.Name, Column: c, Position: n + 1, ReferencedTable: f.RefTable, ReferencedColumn: f.RefColumns[n], UpdateRule: "RESTRICT", DeleteRule: "RESTRICT"})
		}
	}
	return m
}

func reissueExpectSchema(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	m := reissueTestSchema(t)
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}).AddRow("el_mng_report_reissue", "InnoDB").AddRow("el_mng_reissue_audit", "InnoDB"))
	cols := sqlmock.NewRows([]string{"table_name", "column_name", "column_type", "is_nullable", "character_set_name", "collation_name"})
	for _, r := range m.Columns {
		cols.AddRow(r.Table, r.Name, r.ColumnType, r.Nullable, r.Charset, r.Collation)
	}
	mock.ExpectQuery("SELECT table_name AS table_name, column_name AS column_name").WillReturnRows(cols)
	idx := sqlmock.NewRows([]string{"table_name", "index_name", "non_unique", "seq_in_index", "column_name", "prefix_length"})
	for _, r := range m.Indexes {
		idx.AddRow(r.Table, r.Name, r.NonUnique, r.Position, r.Column, r.Prefix)
	}
	mock.ExpectQuery("SELECT table_name AS table_name,index_name AS index_name").WillReturnRows(idx)
	fk := sqlmock.NewRows([]string{"table_name", "constraint_name", "column_name", "ordinal_position", "referenced_table_name", "referenced_column_name", "update_rule", "delete_rule"})
	for _, r := range m.ForeignKeys {
		fk.AddRow(r.Table, r.Name, r.Column, r.Position, r.ReferencedTable, r.ReferencedColumn, r.UpdateRule, r.DeleteRule)
	}
	mock.ExpectQuery("SELECT k.table_name AS table_name,k.constraint_name AS constraint_name").WillReturnRows(fk)
}

func reissueFixture(t *testing.T, code string) (managementTraitsResultValidationFixture, managementTraitsRuntimeRecords, model.Paper, model.ManagementTraitsExamProfile, managementTraitsRuntimeOwner) {
	t.Helper()
	f, r, p, profile, owner := managementRuntimeLoadFixture(t, false)
	if code == "00202" {
		m, e := DecodeManagementTraitsManifest([]byte(f.Bundle.ScoringManifest), 1<<20)
		if e != nil {
			t.Fatal(e)
		}
		mp, e := DecodeManagementTraitsMapping([]byte(f.Paper.MappingSnapshot), m, 1<<20)
		if e != nil {
			t.Fatal(e)
		}
		m.Questionnaire, m.Versions, _ = managementTraitsRuntimeVersions(code)
		mp.Questionnaire = m.Questionnaire
		mc, e := CanonicalManagementTraitsManifest(m)
		if e != nil {
			t.Fatal(e)
		}
		mp.ManifestSHA = mc.SHA256
		pc, e := CanonicalManagementTraitsMapping(m, mp)
		if e != nil {
			t.Fatal(e)
		}
		f.Bundle.Questionnaire = m.Questionnaire
		f.Bundle.ProductVersion = m.Versions.Product
		f.Bundle.QuestionVersion = m.Versions.Question
		f.Bundle.ScoringVersion = m.Versions.Scoring
		f.Bundle.NormVersion = m.Versions.Norm
		f.Bundle.ScoringManifest = string(mc.JSON)
		f.Bundle.ScoringManifestSHA = mc.SHA256
		f.Paper.ScoringManifestSHA = mc.SHA256
		f.Paper.MappingSHA = pc.SHA256
		f.Paper.MappingSnapshot = string(pc.JSON)
		var field managementTraitsRuntimeFieldContract
		if json.Unmarshal([]byte(f.Paper.FieldContract), &field) != nil {
			t.Fatal("field")
		}
		field.RepoCode = code
		field.ManifestSHA = mc.SHA256
		field.MappingSHA = pc.SHA256
		fc, _ := managementTraitsCanonicalBytes(field)
		f.Paper.FieldContract = string(fc.JSON)
		ec, _ := managementTraitsCanonicalBytes(managementTraitsRuntimeEvidence{Schema: "mng-current-source-evidence-v1", Source: field.Source, RepoCode: code, CapturedAt: f.Paper.SourceCapturedAt, ManifestSHA: mc.SHA256, MappingSHA: pc.SHA256})
		f.Paper.EvidenceSnapshot = string(ec.JSON)
		f.Paper.EvidenceSHA = ec.SHA256
		for i := range f.Questions {
			q := &f.Questions[i]
			c, _ := managementTraitsCanonicalBytes(managementTraitsFrozenQuestion{"mng-frozen-question-v1", mc.SHA256, pc.SHA256, mp.Questions[q.Number-1], q.DimensionKey, q.Reverse, q.PaperQuestionID, q.DisplayOrder})
			q.ScoringSnapshotSHA = c.SHA256
		}
		profile.MappingSnapshot = f.Paper.MappingSnapshot
		profile.MappingSHA = f.Paper.MappingSHA
		profile.FieldContract = f.Paper.FieldContract
		var e2 error
		r, e2 = buildManagementTraitsRuntimeRecords(f.Bundle, f.Paper, f.Questions, r.Run.ID, *r.Run.SubmittedAt, 1<<20)
		if e2 != nil {
			t.Fatal(e2)
		}
	}
	return f, r, p, profile, owner
}

func reissueExpectFrozen(t *testing.T, mock sqlmock.Sqlmock, f managementTraitsResultValidationFixture, r managementTraitsRuntimeRecords, p model.Paper, profile model.ManagementTraitsExamProfile, owner managementTraitsRuntimeOwner) {
	t.Helper()
	managementRuntimeExpectLoad(t, mock, f, r, p, profile, owner)
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
	mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
	qs := make([]any, len(f.Questions))
	for i := range qs {
		qs[i] = f.Questions[i]
	}
	mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot.*ORDER BY v_number").WillReturnRows(managementRuntimeModelRows(t, qs...))
}

func reissueFixtureArchive(t *testing.T, f managementTraitsResultValidationFixture, r managementTraitsRuntimeRecords, c ManagementTraitsTestContent) []byte {
	t.Helper()
	dto, e := BuildManagementTraitsTestReport(f.Bundle, f.Paper, f.Questions, r.Run, r.Dimensions, r.Modules, r.Receipt, c, 1<<20)
	if e != nil {
		t.Fatal(e)
	}
	b, e := json.Marshal(managementTraitsFrozenReissue{"mng-frozen-reissue-archive-v1", managementTraitsFrozenReissueKind, dto, f.Paper, f.Bundle, f.Questions})
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func TestManagementTraitsReissueAPISchemaDrift(t *testing.T) {
	d, e := managementTraitsReissueContract()
	if e != nil {
		t.Fatal(e)
	}
	if validateManagementTraitsReissueSchema(d, reissueTestSchema(t)) != nil {
		t.Fatal("valid schema refused")
	}
	for _, kind := range []string{"column", "nullable", "collation", "index", "foreign", "extra"} {
		t.Run(kind, func(t *testing.T) {
			m := reissueTestSchema(t)
			switch kind {
			case "column":
				m.Columns = m.Columns[1:]
			case "nullable":
				m.Columns[4].Nullable = "YES"
			case "collation":
				m.Columns[4].Collation = "utf8mb4_bin"
			case "index":
				m.Indexes[0].Prefix = 1
			case "foreign":
				m.ForeignKeys[0].DeleteRule = "CASCADE"
			case "extra":
				m.ForeignKeys = append(m.ForeignKeys, m.ForeignKeys[0])
			}
			if validateManagementTraitsReissueSchema(d, m) == nil {
				t.Fatal("drift accepted")
			}
		})
	}
}

func TestManagementTraitsReissueAPITransactionsAndArchive(t *testing.T) {
	templateSHA := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	for _, code := range []string{"00201", "00202", "00501", "00502"} {
		for _, mode := range []string{"commit", "render-failure", "audit-rollback", "concurrent-reuse"} {
			t.Run(code+"-"+mode, func(t *testing.T) {
				f, r, p, profile, owner := reissueFixture(t, code)
				c := reissueTestContent(t)
				raw := reissueFixtureArchive(t, f, r, c)
				db, mock := managementRuntimeDB(t)
				reissueExpectSchema(t, mock)
				managementRuntimeExpectSchema(t, mock)
				reissueExpectFrozen(t, mock, f, r, p, profile, owner)
				mock.ExpectQuery("SELECT .*el_mng_report_reissue").WillReturnRows(sqlmock.NewRows([]string{"id"}))
				mock.ExpectCommit()
				mock.ExpectExec("REISSUE_RENDER_OUTSIDE_TX").WillReturnResult(sqlmock.NewResult(0, 0))
				root := t.TempDir()
				old := []byte("LEGACY-DO-NOT-DELETE")
				if os.WriteFile(filepath.Join(root, "historical.pdf"), old, 0600) != nil {
					t.Fatal("fixture")
				}
				pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 2048)...)
				if mode != "render-failure" {
					reissueExpectFrozen(t, mock, f, r, p, profile, owner)
					if mode == "concurrent-reuse" {
						id := "22222222-2222-4222-8222-222222222222"
						key, e := writeManagementTraitsTestPDF(root, id, pdf, nil)
						if e != nil {
							t.Fatal(e)
						}
						existing := model.ManagementTraitsReportReissue{ID: id, RunID: r.Run.ID, PaperID: p.ID, ExamID: p.ExamID, Kind: managementTraitsFrozenReissueKind, Status: "completed", DataSnapshot: string(raw), DataSHA: managementTraitsReissueSHA(raw), TemplateSHA: templateSHA, ContentSHA: c.SourceSHA(), FileKey: key, FileSHA: managementTraitsReissueSHA(pdf), FileBytes: int64(len(pdf)), CreatedBy: 1, CreatedAt: time.Now()}
						mock.ExpectQuery("SELECT .*el_mng_report_reissue").WillReturnRows(managementRuntimeModelRows(t, existing))
						reissueExpectGenerateProof(t, mock, existing)
					} else {
						mock.ExpectQuery("SELECT .*el_mng_report_reissue").WillReturnRows(sqlmock.NewRows([]string{"id"}))
						mock.ExpectExec("INSERT INTO `el_mng_report_reissue`").WillReturnResult(sqlmock.NewResult(1, 1))
					}
					if mode == "audit-rollback" {
						mock.ExpectExec("INSERT INTO `el_mng_reissue_audit`").WillReturnError(errors.New("private credential sentinel"))
						mock.ExpectRollback()
					} else {
						mock.ExpectExec("INSERT INTO `el_mng_reissue_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
						mock.ExpectCommit()
					}
				}
				s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
				got, reused, e := s.GenerateReportReissueWithTemplateSHA(context.Background(), r.Run.ID, 1, root, c, templateSHA, func(ctx context.Context, dto ManagementTraitsTestReportData) ([]byte, error) {
					if dto.Questionnaire != f.Bundle.Questionnaire || dto.Participant.Name != owner.Name || !dto.TestOnly || len(dto.Dimensions) != 13 {
						t.Fatal("wrong frozen DTO")
					}
					if db.Exec("REISSUE_RENDER_OUTSIDE_TX").Error != nil {
						t.Fatal("render in transaction")
					}
					if mode == "render-failure" {
						return nil, errors.New("private renderer path")
					}
					return pdf, nil
				})
				fail := mode == "render-failure" || mode == "audit-rollback"
				if fail && e == nil || !fail && (e != nil || got.Status != "completed" || got.TemplateSHA != templateSHA || reused != (mode == "concurrent-reuse")) {
					t.Fatalf("mode %s failed: %v", mode, e)
				}
				b, _ := os.ReadFile(filepath.Join(root, "historical.pdf"))
				if !bytes.Equal(b, old) {
					t.Fatal("old PDF changed")
				}
				files, _ := os.ReadDir(root)
				want := 2
				if fail {
					want = 1
				}
				if len(files) != want {
					t.Fatal("orphan/missing PDF", len(files))
				}
				if !fail {
					snapshot := []byte(got.DataSnapshot)
					if !bytes.Equal(snapshot, raw) || !bytes.Contains(snapshot, []byte(owner.Name)) {
						t.Fatal("original fixture identity or archive lost")
					}
					reissueExpectSchema(t, mock)
					mock.ExpectBegin()
					mock.ExpectQuery("SELECT .*el_mng_report_reissue.*WHERE id = ").WillReturnRows(managementRuntimeModelRows(t, got))
					mock.ExpectExec("INSERT INTO `el_mng_reissue_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
					_, download, e := s.ReadReportReissueWithTemplateSHAs(context.Background(), got.ID, 1, root, "download", map[string]bool{templateSHA: true})
					if e != nil || !bytes.Equal(download, pdf) {
						t.Fatal("archived download failed", e)
					}
					for _, bad := range []string{"path", "hash", "status", "snapshot", "size"} {
						t.Run(bad, func(t *testing.T) {
							x := got
							switch bad {
							case "path":
								x.FileKey = "../escape.pdf"
							case "hash":
								x.FileSHA = "bad"
							case "status":
								x.Status = "generating"
							case "snapshot":
								x.DataSnapshot = "{}"
							case "size":
								x.FileBytes++
							}
							if _, e := s.readReissueFile(x, root); e == nil {
								t.Fatal("bad archived output accepted")
							}
						})
					}
				}
				if e := mock.ExpectationsWereMet(); e != nil {
					t.Fatal(e)
				}
			})
		}
	}
}

func TestManagementTraitsReissueAPIMetadataEmptyAndSchemaAbsent(t *testing.T) {
	for _, installed := range []bool{false, true} {
		db, mock := managementRuntimeDB(t)
		if installed {
			reissueExpectSchema(t, mock)
			mock.ExpectQuery("SELECT id,run_id,paper_id,exam_id,kind,status,template_sha,content_sha,file_sha,file_bytes,created_by,created_at FROM `el_mng_report_reissue`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		} else {
			mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
		}
		s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
		rows, e := s.ListReportReissues(context.Background(), "paper")
		if installed && (e != nil || rows == nil || len(rows) != 0) || !installed && !errors.Is(e, ErrManagementTraitsReissueClosed) {
			t.Fatal("metadata/schema contract", e)
		}
		if e := mock.ExpectationsWereMet(); e != nil {
			t.Fatal(e)
		}
	}
}

func TestManagementTraitsReissueAPIRejectsUnprovenSource(t *testing.T) {
	for _, mode := range []string{"missing", "historical", "incomplete", "owner", "revoked", "score"} {
		t.Run(mode, func(t *testing.T) {
			f, r, p, profile, owner := reissueFixture(t, "00201")
			db, mock := managementRuntimeDB(t)
			managementRuntimeExpectSchema(t, mock)
			if mode == "missing" {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT .*el_mng_result_run").WillReturnRows(sqlmock.NewRows([]string{"id"}))
				mock.ExpectRollback()
			} else {
				switch mode {
				case "historical":
					f.Paper.Source = "legacy_verified_snapshot"
				case "incomplete":
					r.Run.Status = "incomplete"
				case "owner":
					owner.ID = "foreign"
				case "revoked":
					f.Bundle.Status = "revoked"
				case "score":
					r.Run.OverallScore = managementTraitsResultValidationDecimal(big.NewRat(51, 1))
				}
				// Ordered full expectations intentionally catch any extra writer. Only
				// the read prefix consumed before rejection is required here.
				mock.MatchExpectationsInOrder(false)
				managementRuntimeExpectLoad(t, mock, f, r, p, profile, owner)
				mock.ExpectRollback()
			}
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			if _, e := s.QualifyReportReissue(context.Background(), r.Run.ID, reissueTestContent(t)); e == nil {
				t.Fatal("unproven source accepted")
			}
			if mode == "missing" {
				if e := mock.ExpectationsWereMet(); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}

// Real answer values and original stored score metadata, synthetic identity and
// source IDs/texts only. Never claim these fixture SHAs are original-source SHAs.
func TestManagementTraitsReissueAPIRealAnswersMaskedIdentity(t *testing.T) {
	file := filepath.Join("..", "..", "..", "scripts", "test", "results", "mng-reissue-api-local-20261008", "real-source-audit.json")
	b, e := os.ReadFile(file)
	if os.IsNotExist(e) {
		t.Skip("explicit readonly audit receipt required")
	}
	if e != nil {
		t.Fatal(e)
	}
	var a struct {
		Status string `json:"status"`
		Source []struct {
			Overall, Norm, Level                                                  string
			OwnerCount, OtherOwnerCount, Questions, Dimensions, Modules, Receipts int
			EvidenceSHA, EvidenceByteSHA                                          string
		}
		Answers []struct {
			V, Raw, Reverse, Final, OptionsCount, SelectedCount, SelectedMatches, LegacyMatches int
			Dimension                                                                           string
			SubmittedMatches                                                                    bool
		}
		Dimensions []struct {
			Key, Score, Norm, Level string
			Count, Answered, Sum    int
		}
		Modules []struct {
			Key, Score string
			Count      int
		}
	}
	if json.Unmarshal(b, &a) != nil || a.Status != "READ_COMPLETED" || len(a.Source) != 1 || len(a.Answers) != 140 || len(a.Dimensions) != 13 || len(a.Modules) != 4 {
		t.Fatal("missing audit facts")
	}
	src := a.Source[0]
	if src.OwnerCount != 1 || src.OtherOwnerCount != 0 || src.Questions != 140 || src.Dimensions != 13 || src.Modules != 4 || src.Receipts != 1 || src.EvidenceSHA != src.EvidenceByteSHA {
		t.Fatal("source qualification metadata")
	}
	f, r, _, _, _ := reissueFixture(t, "00201")
	for i, row := range a.Answers {
		q := &f.Questions[i]
		if row.V != q.Number || row.Dimension != q.DimensionKey || (row.Reverse == 1) != q.Reverse || row.Raw < 1 || row.Raw > 5 || row.Final != func() int {
			if q.Reverse {
				return 6 - row.Raw
			}
			return row.Raw
		}() || row.OptionsCount != 5 || row.SelectedCount != 1 || row.SelectedMatches != 1 || row.LegacyMatches != 1 || !row.SubmittedMatches {
			t.Fatal("original answer metadata inconsistent")
		}
		raw, final, selected := row.Raw, row.Final, fmt.Sprintf("source-o-%d-%d", q.Number, row.Raw)
		q.RawAnswer = &raw
		q.FinalScore = &final
		q.SelectedOptionID = &selected
	}
	// Independently derive each rational score from raw values, not stored scores.
	overall := new(big.Rat)
	for _, d := range a.Dimensions {
		sum, count := 0, 0
		for _, row := range a.Answers {
			if row.Dimension == d.Key {
				sum += row.Final
				count++
			}
		}
		score := new(big.Rat).Sub(new(big.Rat).Mul(big.NewRat(25, 1), big.NewRat(int64(sum), int64(count))), big.NewRat(25, 1))
		cached, _ := decimal.NewFromString(d.Score)
		if sum != d.Sum || count != d.Count || d.Answered != count || !decimal.NewFromBigRat(score, 6).Equal(cached) {
			t.Fatal("independent dimension score differs")
		}
		overall.Add(overall, score)
	}
	overall.Quo(overall, big.NewRat(13, 1))
	stored, _ := decimal.NewFromString(src.Overall)
	if !decimal.NewFromBigRat(overall, 6).Equal(stored) {
		t.Fatal("independent overall differs")
	}
	r, e = buildManagementTraitsRuntimeRecords(f.Bundle, f.Paper, f.Questions, r.Run.ID, *r.Run.SubmittedAt, 1<<20)
	if e != nil {
		t.Fatal(e)
	}
	c := reissueTestContent(t)
	dto, e := BuildManagementTraitsTestReport(f.Bundle, f.Paper, f.Questions, r.Run, r.Dimensions, r.Modules, r.Receipt, c, 1<<20)
	if e != nil || dto.Overall.Score != "58.64" {
		t.Fatal("masked real answers DTO failed", e)
	}
	for i, d := range a.Dimensions {
		cached, _ := decimal.NewFromString(d.Score)
		if !r.Dimensions[i].Score.Equal(cached) || *r.Dimensions[i].Level != d.Level {
			t.Fatal("stored dimension mismatch")
		}
	}
	for i, m := range a.Modules {
		cached, _ := decimal.NewFromString(m.Score)
		if r.Modules[i].ModuleKey != m.Key || !r.Modules[i].Score.Equal(cached) || r.Modules[i].DimensionCount != m.Count {
			t.Fatal("stored module mismatch")
		}
	}
	if out := os.Getenv("MNG_REISSUE_MASKED_DTO_OUTPUT"); out != "" {
		if !filepath.IsAbs(out) {
			t.Fatal("absolute synthetic output required")
		}
		dto.Participant = ManagementTraitsTestReportParticipant{Name: "SYNTHETIC IDENTITY - REAL ANSWERS", Telephone: "13800000000"}
		raw, e := json.Marshal(dto)
		if e != nil {
			t.Fatal(e)
		}
		file, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		_, e = file.Write(raw)
		closeErr := file.Close()
		if e != nil || closeErr != nil {
			t.Fatal("safe DTO write")
		}
	}
	t.Log("real answers 140/140; reverse 40; persisted 13/4 scores matched; synthetic identity only; overall 58.64")
}

// Regression MT-REISSUE-1062 in docs/regression-tests.md. The failed INSERT
// sees a winner committed after this transaction's earlier consistent read.
func TestBugMTReissue1062FreshWinner(t *testing.T) {
	for _, mode := range []string{"winner", "missing-winner", "wrong-run", "wrong-paper", "wrong-exam", "wrong-data", "wrong-template", "wrong-content", "wrong-kind", "incomplete-winner", "missing-file", "hash-drift", "size-drift", "path-tamper", "missing-generate-audit", "audit-failure", "revoked", "source-drift", "other-report-key", "audit-1062", "deadlock-1213", "string-error", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			f, records, p, profile, owner := reissueFixture(t, "00201")
			content := reissueTestContent(t)
			raw := reissueFixtureArchive(t, f, records, content)
			db, mock := managementRuntimeDB(t)
			reissueExpectSchema(t, mock)
			managementRuntimeExpectSchema(t, mock)
			root := t.TempDir()
			pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("SYNTHETIC WINNER\n"), 150)...)
			id := "33333333-3333-4333-8333-333333333333"
			key, e := writeManagementTraitsTestPDF(root, id, pdf, nil)
			if e != nil {
				t.Fatal(e)
			}
			winner := model.ManagementTraitsReportReissue{ID: id, RunID: records.Run.ID, PaperID: p.ID, ExamID: p.ExamID, Kind: managementTraitsFrozenReissueKind, Status: "completed", DataSnapshot: string(raw), DataSHA: managementTraitsReissueSHA(raw), TemplateSHA: managementTraitsReportTemplateSHA, ContentSHA: content.SourceSHA(), FileKey: key, FileSHA: managementTraitsReissueSHA(pdf), FileBytes: int64(len(pdf)), CreatedBy: 2, CreatedAt: time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)}
			old := []byte("IMMUTABLE LEGACY PDF")
			if os.WriteFile(filepath.Join(root, "historical.pdf"), old, 0600) != nil {
				t.Fatal("fixture")
			}
			reissueExpectFrozen(t, mock, f, records, p, profile, owner)
			mock.ExpectQuery("SELECT .*el_mng_report_reissue").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			mock.ExpectCommit()
			reissueExpectFrozen(t, mock, f, records, p, profile, owner)
			mock.ExpectQuery("SELECT .*el_mng_report_reissue").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			duplicate := &driver.MySQLError{Number: 1062, Message: "Duplicate entry 'PRIVATE_SECRET_SENTINEL' for key 'el_mng_report_reissue.uk_mng_reissue_input'"}
			switch mode {
			case "other-report-key":
				duplicate.Message = "Duplicate entry 'PRIVATE_SECRET_SENTINEL' for key 'el_mng_report_reissue.PRIMARY'"
			case "deadlock-1213":
				duplicate.Number = 1213
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "audit-1062" {
				mock.ExpectExec("INSERT INTO `el_mng_report_reissue`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO `el_mng_reissue_audit`").WillReturnError(duplicate)
			} else if mode == "string-error" {
				mock.ExpectExec("INSERT INTO `el_mng_report_reissue`").WillReturnError(errors.New(duplicate.Message))
			} else {
				mock.ExpectExec("INSERT INTO `el_mng_report_reissue`").WillReturnError(fmt.Errorf("wrapped insert: %w", duplicate))
			}
			mock.ExpectRollback()
			fallback := mode != "other-report-key" && mode != "audit-1062" && mode != "deadlock-1213" && mode != "string-error" && mode != "cancelled"
			if fallback {
				if mode == "revoked" {
					managementRuntimeExpectLoad(t, mock, f, records, p, profile, owner)
					revoked := f.Bundle
					revoked.Status = "revoked"
					mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, revoked))
					mock.ExpectRollback()
				} else {
					if mode == "source-drift" {
						managementRuntimeExpectLoad(t, mock, f, records, p, profile, owner)
						mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs(f.Bundle.ID).WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
						mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WillReturnRows(managementRuntimeModelRows(t, f.Paper))
						mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WillReturnRows(managementRuntimeModelRows(t, f.Bundle))
						questions := append([]model.ManagementTraitsPaperQuestionSnapshot{}, f.Questions...)
						questions[0].CreatedAt = questions[0].CreatedAt.Add(time.Second)
						qs := make([]any, len(questions))
						for i := range qs {
							qs[i] = questions[i]
						}
						mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot.*ORDER BY v_number").WillReturnRows(managementRuntimeModelRows(t, qs...))
					} else {
						reissueExpectFrozen(t, mock, f, records, p, profile, owner)
					}
					if mode != "source-drift" {
						switch mode {
						case "wrong-run":
							winner.RunID = "foreign-run"
						case "wrong-paper":
							winner.PaperID = "foreign-paper"
						case "wrong-exam":
							winner.ExamID = "foreign-exam"
						case "wrong-data":
							winner.DataSnapshot += " "
							winner.DataSHA = managementTraitsReissueSHA([]byte(winner.DataSnapshot))
						case "wrong-template":
							winner.TemplateSHA = "foreign-template"
						case "wrong-content":
							winner.ContentSHA = "foreign-content"
						case "wrong-kind":
							winner.Kind = "other"
						case "incomplete-winner":
							winner.Status = "generating"
						case "missing-file":
							if os.Remove(filepath.Join(root, key)) != nil {
								t.Fatal("fixture")
							}
						case "hash-drift":
							winner.FileSHA = managementTraitsReissueSHA([]byte("drift"))
						case "size-drift":
							winner.FileBytes++
						case "path-tamper":
							winner.FileKey = "../historical.pdf"
						}
						if mode == "missing-winner" {
							mock.ExpectQuery("SELECT .*el_mng_report_reissue").WillReturnRows(sqlmock.NewRows([]string{"id"}))
						} else {
							mock.ExpectQuery("SELECT .*el_mng_report_reissue").WillReturnRows(managementRuntimeModelRows(t, winner))
						}
						if mode == "winner" || mode == "missing-generate-audit" || mode == "audit-failure" {
							if mode == "missing-generate-audit" {
								mock.ExpectQuery("SELECT .*el_mng_reissue_audit").WillReturnRows(sqlmock.NewRows([]string{"id"}))
							} else {
								reissueExpectGenerateProof(t, mock, winner)
							}
							if mode == "winner" {
								mock.ExpectExec("INSERT INTO `el_mng_reissue_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
								mock.ExpectCommit()
							} else if mode == "audit-failure" {
								mock.ExpectExec("INSERT INTO `el_mng_reissue_audit`").WillReturnError(errors.New("PRIVATE_SECRET_SENTINEL"))
								mock.ExpectRollback()
							} else {
								mock.ExpectRollback()
							}
						} else {
							mock.ExpectRollback()
						}
					} else {
						mock.ExpectRollback()
					}
				}
			}
			s := NewManagementTraitsRuntimeService(db, "test-only", 1<<20)
			if mode == "cancelled" {
				db.Callback().Create().After("gorm:create").Register("test:cancel-duplicate", func(tx *gorm.DB) {
					if tx.Statement.Table == "el_mng_report_reissue" {
						cancel()
					}
				})
			}
			rendered := 0
			got, reused, err := s.GenerateReportReissue(ctx, records.Run.ID, 1, root, content, func(context.Context, ManagementTraitsTestReportData) ([]byte, error) { rendered++; return pdf, nil })
			if mode == "winner" {
				if err != nil || !reused || got.ID != winner.ID || got.DataSHA != winner.DataSHA || got.FileSHA != winner.FileSHA || got.CreatedBy != 2 || !got.CreatedAt.Equal(winner.CreatedAt) {
					t.Fatal("fresh committed winner was not returned")
				}
			} else if err == nil || reused || got.ID != "" || !errors.Is(err, ErrManagementTraitsReissueInvalid) || bytes.Contains([]byte(err.Error()), []byte("PRIVATE_SECRET_SENTINEL")) {
				t.Fatal("unsafe fallback or private error exposure")
			}
			if rendered != 1 {
				t.Fatal("renderer retried")
			}
			b, _ := os.ReadFile(filepath.Join(root, "historical.pdf"))
			if !bytes.Equal(b, old) {
				t.Fatal("legacy changed")
			}
			files, _ := os.ReadDir(root)
			want := 2
			if mode == "missing-file" {
				want = 1
			}
			if len(files) != want {
				t.Fatal("loser orphan or winner removed", len(files))
			}
			if mode != "missing-file" {
				b, _ = os.ReadFile(filepath.Join(root, key))
				if !bytes.Equal(b, pdf) {
					t.Fatal("winner changed")
				}
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func reissueExpectGenerateProof(t *testing.T, mock sqlmock.Sqlmock, r model.ManagementTraitsReportReissue) {
	t.Helper()
	a := model.ManagementTraitsReissueAudit{ID: "44444444-4444-4444-8444-444444444444", ReportID: r.ID, PaperID: r.PaperID, ExamID: r.ExamID, ActorID: r.CreatedBy, Action: "generate", CreatedAt: r.CreatedAt}
	mock.ExpectQuery("SELECT .*el_mng_reissue_audit").WithArgs(r.ID, r.PaperID, r.ExamID, "generate", 2).WillReturnRows(managementRuntimeModelRows(t, a))
}
