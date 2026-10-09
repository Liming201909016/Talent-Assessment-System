package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

// Both stored identities and requested destinations must be supplied by callers.
// IdentifierExamID scopes only identifier lookups, never existing object IDs.
type ManagementTraitsLegacyScopeRequest struct {
	// Only exact admin preparation routes may request this exception. The
	// middleware enforces normal JWT/admin before executing the handler.
	DraftPreparation     bool
	DraftPreparationUsed *bool
	// AllLegacy is only for collections whose effective handler exam filter is absent.
	AllLegacy           bool
	ExamIDs             []string
	PaperIDs            []string
	CandidateIDs        []string
	TesterIDs           []string
	PaperQuestionIDs    []string
	TesterIdentifiers   []string
	CandidateTelephones []string
	PDFPaths            []string
	IdentifierExamID    string
}

type managementTraitsLegacySet map[string]struct{}

func (s managementTraitsLegacySet) add(values ...string) bool {
	changed := false
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := s[v]; !ok {
			s[v] = struct{}{}
			changed = true
		}
	}
	return changed
}

func (s managementTraitsLegacySet) values() []string {
	out := make([]string, 0, len(s))
	for v := range s {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

type managementTraitsLegacyScope struct {
	exams, papers, candidates, testers, questions managementTraitsLegacySet
	targetExams                                   managementTraitsLegacySet
}

type managementTraitsLegacyPredicate struct {
	sql  string
	args []any
}

func managementTraitsLegacyPredicates(column string, values []string) []managementTraitsLegacyPredicate {
	rows := make([]managementTraitsLegacyPredicate, 0, len(values))
	for _, v := range values {
		rows = append(rows, managementTraitsLegacyPredicate{column + " = ?", []any{v}})
	}
	return rows
}

// Batches contain at most 1000 bound values, not one round trip per object.
func managementTraitsLegacyQueries(predicates []managementTraitsLegacyPredicate, visit func(string, []any) (bool, error)) (bool, error) {
	for len(predicates) > 0 {
		parts, args := make([]string, 0), make([]any, 0)
		n := 0
		for n < len(predicates) && len(args)+len(predicates[n].args) <= 1000 {
			parts = append(parts, "("+predicates[n].sql+")")
			args = append(args, predicates[n].args...)
			n++
		}
		if n == 0 {
			return false, ErrManagementTraitsRuntimeInvalid
		}
		found, err := visit(strings.Join(parts, " OR "), args)
		if err != nil {
			return false, ErrManagementTraitsRuntimeInvalid
		}
		if found {
			return true, nil
		}
		predicates = predicates[n:]
	}
	return false, nil
}

func managementTraitsLegacyFresh(predicates []managementTraitsLegacyPredicate, seen map[string]bool) []managementTraitsLegacyPredicate {
	out := make([]managementTraitsLegacyPredicate, 0, len(predicates))
	for _, p := range predicates {
		key := p.sql + fmt.Sprintf("%q", p.args)
		if !seen[key] {
			seen[key] = true
			out = append(out, p)
		}
	}
	return out
}

func managementTraitsLegacyOwnerPredicates(s managementTraitsLegacyScope, req ManagementTraitsLegacyScopeRequest, kind string, initial bool) []managementTraitsLegacyPredicate {
	ids := s.candidates
	if kind == "tester" {
		ids = s.testers
	}
	p := managementTraitsLegacyPredicates("id", ids.values())
	p = append(p, managementTraitsLegacyPredicates("paper_id", s.papers.values())...)
	p = append(p, managementTraitsLegacyPredicates("exam_id", s.targetExams.values())...)
	if !initial {
		return p
	}
	columns, identifiers := []string{"telephone"}, req.CandidateTelephones
	if kind == "tester" {
		columns, identifiers = []string{"id_number", "telephone"}, req.TesterIdentifiers
	}
	unique := make(managementTraitsLegacySet)
	unique.add(identifiers...)
	for _, column := range columns {
		for _, v := range unique.values() {
			atom := managementTraitsLegacyPredicate{column + " = ?", []any{v}}
			if req.IdentifierExamID != "" {
				atom.sql += " AND exam_id = ?"
				atom.args = append(atom.args, req.IdentifierExamID)
			}
			p = append(p, atom)
		}
	}
	paths := make(managementTraitsLegacySet)
	for _, v := range req.PDFPaths {
		paths.add(strings.ReplaceAll(v, "\\", "/"))
	}
	p = append(p, managementTraitsLegacyPredicates("BINARY REPLACE(pdf_path, CHAR(92), '/')", paths.values())...)
	return p
}

// Resolve the transitive stored paper/exam/owner closure without status filters:
// deleted owners and broken/rebound legacy links still carry protection evidence.
func managementTraitsLegacyClosure(db *gorm.DB, s managementTraitsLegacyScope, req ManagementTraitsLegacyScopeRequest) error {
	_, err := managementTraitsLegacyQueries(managementTraitsLegacyPredicates("id", s.questions.values()), func(where string, args []any) (bool, error) {
		rows := make([]struct {
			PaperID string `gorm:"column:paper_id"`
		}, 0)
		if err := db.Raw("SELECT paper_id FROM el_paper_qu WHERE "+where, args...).Scan(&rows).Error; err != nil {
			return false, err
		}
		for _, r := range rows {
			s.papers.add(r.PaperID)
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	seen := []map[string]bool{{}, {}, {}}
	for initial := true; ; initial = false {
		paper := managementTraitsLegacyPredicates("id", s.papers.values())
		paper = append(paper, managementTraitsLegacyPredicates("exam_id", s.targetExams.values())...)
		groups := [][]managementTraitsLegacyPredicate{
			managementTraitsLegacyFresh(paper, seen[0]),
			managementTraitsLegacyFresh(managementTraitsLegacyOwnerPredicates(s, req, "candidate", initial), seen[1]),
			managementTraitsLegacyFresh(managementTraitsLegacyOwnerPredicates(s, req, "tester", initial), seen[2]),
		}
		if len(groups[0])+len(groups[1])+len(groups[2]) == 0 {
			return nil
		}
		for i, predicates := range groups {
			table, projection := "el_paper", "id, exam_id"
			if i == 1 {
				table, projection = "el_candidate", "id, exam_id, paper_id"
			}
			if i == 2 {
				table, projection = "el_tester", "id, exam_id, paper_id"
			}
			_, err := managementTraitsLegacyQueries(predicates, func(where string, args []any) (bool, error) {
				rows := make([]struct {
					ID      string `gorm:"column:id"`
					ExamID  string `gorm:"column:exam_id"`
					PaperID string `gorm:"column:paper_id"`
				}, 0)
				if err := db.Raw("SELECT "+projection+" FROM "+table+" WHERE "+where, args...).Scan(&rows).Error; err != nil {
					return false, err
				}
				for _, r := range rows {
					s.exams.add(r.ExamID)
					switch i {
					case 0:
						s.papers.add(r.ID)
					case 1:
						s.candidates.add(r.ID)
						s.papers.add(r.PaperID)
					case 2:
						s.testers.add(r.ID)
						s.papers.add(r.PaperID)
					}
				}
				return false, nil
			})
			if err != nil {
				return err
			}
		}
	}
}

func managementTraitsLegacyRuntimePredicates(s managementTraitsLegacyScope, columns map[string]bool) []managementTraitsLegacyPredicate {
	p := make([]managementTraitsLegacyPredicate, 0)
	for _, item := range []struct {
		column string
		values managementTraitsLegacySet
	}{
		{"paper_id", s.papers}, {"exam_id", s.targetExams}, {"profile_exam_id", s.exams},
		{"candidate_id", s.candidates}, {"tester_id", s.testers}, {"paper_question_id", s.questions},
	} {
		if columns[item.column] {
			p = append(p, managementTraitsLegacyPredicates(item.column, item.values.values())...)
		}
	}
	if columns["participant_id"] {
		for _, item := range []struct {
			kind string
			ids  managementTraitsLegacySet
		}{{"candidate", s.candidates}, {"tester", s.testers}} {
			for _, id := range item.ids.values() {
				atom := managementTraitsLegacyPredicate{"participant_id = ?", []any{id}}
				if columns["participant_type"] {
					atom = managementTraitsLegacyPredicate{"participant_type = ? AND participant_id = ?", []any{item.kind, id}}
				}
				p = append(p, atom)
			}
		}
	}
	return p
}

func managementTraitsLegacyExists(db *gorm.DB, table string, p []managementTraitsLegacyPredicate) (bool, error) {
	return managementTraitsLegacyQueries(p, func(where string, args []any) (bool, error) {
		rows := make([]struct {
			Protected int `gorm:"column:protected"`
		}, 0, 1)
		err := db.Raw("SELECT 1 AS protected FROM "+table+" WHERE "+where+" LIMIT 1", args...).Scan(&rows).Error
		return len(rows) > 0, err
	})
}

// CheckManagementTraitsLegacyScope never infers runtime scope from a repo code,
// source question, status, frozen_at or cache. Callers hold a mutation lease
// through all legacy work; freeze callers hold the exclusive gate through commit.
func CheckManagementTraitsLegacyScope(ctx context.Context, db *gorm.DB, req ManagementTraitsLegacyScopeRequest) (bool, error) {
	if db == nil || ctx == nil {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	db = db.WithContext(ctx)
	core := map[string]bool{
		(model.ManagementTraitsDefinitionBundle{}).TableName():      true,
		(model.ManagementTraitsExamProfile{}).TableName():           true,
		(model.ManagementTraitsPaperSnapshot{}).TableName():         true,
		(model.ManagementTraitsPaperQuestionSnapshot{}).TableName(): true,
		(model.ManagementTraitsResultRun{}).TableName():             true,
		(model.ManagementTraitsResultDimension{}).TableName():       true,
		(model.ManagementTraitsResultModule{}).TableName():          true,
	}
	tables := make([]struct {
		Name string `gorm:"column:table_name"`
	}, 0)
	const metadata = "SELECT table_name AS table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name,7) = 'el_mng_' ORDER BY table_name"
	if db.Raw(metadata).Scan(&tables).Error != nil {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	count, extras := 0, make([]string, 0)
	present := make(map[string]bool)
	formalTables := managementTraitsFormalTables()
	formalCount := 0
	draftPresent := false
	for _, table := range tables {
		if !strings.HasPrefix(table.Name, "el_mng_") || len(table.Name) > 64 || present[table.Name] {
			return false, ErrManagementTraitsRuntimeInvalid
		}
		// Metadata names are interpolated only after an ASCII identifier allowlist.
		for _, c := range table.Name {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
				return false, ErrManagementTraitsRuntimeInvalid
			}
		}
		present[table.Name] = true
		if table.Name == "el_mng_exam_draft" {
			draftPresent = true
		} else if formalTables[table.Name] {
			formalCount++
		} else if core[table.Name] {
			count++
		} else {
			extras = append(extras, table.Name)
		}
	}
	if draftPresent && CheckManagementTraitsDraftSchema(ctx, db) != nil {
		return false, ErrManagementTraitsDraftClosed
	}
	if formalCount != 0 {
		if formalCount != len(formalTables) || checkManagementTraitsFormalSchema(ctx, db) != nil {
			return false, ErrManagementTraitsRuntimeInvalid
		}
	}
	if count != 0 && count != len(core) {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	if len(tables) == 0 {
		return false, nil
	}
	sort.Strings(extras)
	capabilities := make(map[string]map[string]bool)
	for _, name := range extras {
		capabilities[name] = make(map[string]bool)
	}
	_, err := managementTraitsLegacyQueries(managementTraitsLegacyPredicates("table_name", extras), func(where string, args []any) (bool, error) {
		rows := make([]struct {
			Table  string `gorm:"column:table_name"`
			Column string `gorm:"column:column_name"`
		}, 0)
		if err := db.Raw("SELECT table_name AS table_name, column_name AS column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND ("+where+")", args...).Scan(&rows).Error; err != nil {
			return false, err
		}
		for _, r := range rows {
			if capabilities[r.Table] == nil || r.Column == "" || len(r.Column) > 64 || capabilities[r.Table][r.Column] {
				return false, ErrManagementTraitsRuntimeInvalid
			}
			for _, c := range r.Column {
				if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
					return false, ErrManagementTraitsRuntimeInvalid
				}
			}
			capabilities[r.Table][r.Column] = true
		}
		return false, nil
	})
	if err != nil {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	auditTable := (model.ManagementTraitsReportAudit{}).TableName()
	for name, columns := range capabilities {
		known := false
		for _, c := range []string{"paper_id", "exam_id", "profile_exam_id", "participant_id", "candidate_id", "tester_id", "paper_question_id", "pdf_path"} {
			known = known || columns[c]
		}
		if (columns["run_id"] || columns["result_run_id"]) && count == len(core) {
			known = true
		}
		if name == auditTable {
			// Only the canonical audit has this indirect identity. Arbitrary
			// report_id-only markers remain unknown and fail closed.
			if len(columns) != 5 || !columns["id"] || !columns["report_id"] || !columns["actor_id"] || !columns["action"] || !columns["created_at"] {
				return false, ErrManagementTraitsRuntimeInvalid
			}
			known = true
		}
		if !known {
			return false, ErrManagementTraitsRuntimeInvalid
		}
	}
	if present[auditTable] || present[(model.ManagementTraitsReportRevision{}).TableName()] || present[(model.ManagementTraitsReportCurrent{}).TableName()] {
		// Keep seven-core historical/capture installations compatible. Once
		// canonical reports exist, require the complete current contract and
		// reuse the strict validator on this DB; no constructor, shared cache,
		// callback registry or inferred runtime-service dependencies are needed.
		d, err := managementTraitsSchemaContract()
		if err != nil {
			return false, ErrManagementTraitsRuntimeInvalid
		}
		for _, table := range d.Tables {
			if !present[table.Name] {
				return false, ErrManagementTraitsRuntimeInvalid
			}
		}
		if (&ManagementTraitsRuntimeService{db: db}).checkRuntimeSchema(ctx) != nil {
			return false, ErrManagementTraitsRuntimeInvalid
		}
		// FK metadata alone cannot prove old/imported rows are not orphans.
		// Check globally, including empty and mismatched scopes, before any
		// successful unprotected return. Composite identity drift is an orphan.
		rows := make([]struct {
			Protected int `gorm:"column:protected"`
		}, 0, 1)
		if db.Raw("SELECT 1 AS protected FROM el_mng_report_audit a LEFT JOIN el_mng_report_revision r ON r.id = a.report_id LEFT JOIN el_mng_result_run u ON u.id = r.run_id AND u.paper_id = r.paper_id AND u.exam_id = r.exam_id WHERE r.id IS NULL OR u.id IS NULL LIMIT 1").Scan(&rows).Error != nil || len(rows) != 0 {
			return false, ErrManagementTraitsRuntimeInvalid
		}
	}
	if req.AllLegacy {
		protectedTables := make([]string, 0, 4+len(extras))
		if draftPresent {
			protectedTables = append(protectedTables, "el_mng_exam_draft")
		}
		if count == len(core) {
			protectedTables = append(protectedTables,
				(model.ManagementTraitsExamProfile{}).TableName(),
				(model.ManagementTraitsPaperSnapshot{}).TableName(),
				(model.ManagementTraitsResultRun{}).TableName(),
				(model.ManagementTraitsPaperQuestionSnapshot{}).TableName(),
			)
		}
		protectedTables = append(protectedTables, extras...)
		for _, table := range protectedTables {
			rows := make([]struct {
				Protected int `gorm:"column:protected"`
			}, 0, 1)
			if db.Raw("SELECT 1 AS protected FROM "+table+" LIMIT 1").Scan(&rows).Error != nil {
				return false, ErrManagementTraitsRuntimeInvalid
			}
			if len(rows) > 0 {
				return true, nil
			}
		}
		return false, nil
	}
	s := managementTraitsLegacyScope{
		exams: make(managementTraitsLegacySet), papers: make(managementTraitsLegacySet),
		candidates: make(managementTraitsLegacySet), testers: make(managementTraitsLegacySet),
		questions: make(managementTraitsLegacySet), targetExams: make(managementTraitsLegacySet),
	}
	s.exams.add(req.ExamIDs...)
	s.targetExams.add(req.ExamIDs...)
	s.papers.add(req.PaperIDs...)
	s.candidates.add(req.CandidateIDs...)
	s.testers.add(req.TesterIDs...)
	s.questions.add(req.PaperQuestionIDs...)
	if err := managementTraitsLegacyClosure(db, s, req); err != nil {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	if draftPresent {
		rows, err := managementTraitsInstalledDrafts(ctx, db, s.exams.values())
		if err != nil {
			return false, err
		}
		for _, d := range rows {
			if !req.DraftPreparation || d.Lifecycle != "draft" || len(s.targetExams) != 1 || len(s.exams) != 1 || len(s.papers) != 0 || req.DraftPreparationUsed == nil {
				return true, nil
			}
			*req.DraftPreparationUsed = true
		}
	}
	identityColumns := map[string]bool{"paper_id": true, "exam_id": true, "participant_id": true, "participant_type": true}
	if present[auditTable] {
		found, err := managementTraitsLegacyQueries(managementTraitsLegacyRuntimePredicates(s, identityColumns), func(where string, args []any) (bool, error) {
			return managementTraitsLegacyExists(db, auditTable, []managementTraitsLegacyPredicate{{"report_id IN (SELECT id FROM el_mng_report_revision WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE " + where + "))", args}})
		})
		if err != nil || found {
			return found, err
		}
	}
	if count == len(core) {
		found, err := managementTraitsLegacyExists(db, (model.ManagementTraitsExamProfile{}).TableName(), managementTraitsLegacyPredicates("exam_id", s.exams.values()))
		if err != nil || found {
			return found, err
		}
		for _, item := range []struct {
			table   string
			columns map[string]bool
		}{
			{(model.ManagementTraitsPaperSnapshot{}).TableName(), identityColumns},
			{(model.ManagementTraitsResultRun{}).TableName(), identityColumns},
			{(model.ManagementTraitsPaperQuestionSnapshot{}).TableName(), map[string]bool{"paper_id": true, "paper_question_id": true}},
		} {
			found, err := managementTraitsLegacyExists(db, item.table, managementTraitsLegacyRuntimePredicates(s, item.columns))
			if err != nil || found {
				return found, err
			}
		}
	}
	for _, table := range extras {
		if table == auditTable {
			continue // The actual revision -> run closure was checked above.
		}
		columns := capabilities[table]
		p := managementTraitsLegacyRuntimePredicates(s, columns)
		if columns["pdf_path"] {
			paths := make(managementTraitsLegacySet)
			for _, v := range req.PDFPaths {
				paths.add(strings.ReplaceAll(v, "\\", "/"))
			}
			p = append(p, managementTraitsLegacyPredicates("BINARY REPLACE(pdf_path, CHAR(92), '/')", paths.values())...)
		}
		for _, column := range []string{"run_id", "result_run_id"} {
			if !columns[column] {
				continue
			}
			if count != len(core) {
				return false, ErrManagementTraitsRuntimeInvalid
			}
			found, err := managementTraitsLegacyQueries(managementTraitsLegacyRuntimePredicates(s, identityColumns), func(where string, args []any) (bool, error) {
				return managementTraitsLegacyExists(db, table, []managementTraitsLegacyPredicate{{column + " IN (SELECT id FROM el_mng_result_run WHERE " + where + ")", args}})
			})
			if err != nil || found {
				return found, err
			}
		}
		found, err := managementTraitsLegacyExists(db, table, p)
		if err != nil || found {
			return found, err
		}
	}
	return false, nil
}

// Process-local only. Multiple application instances require a shared gate.
type managementTraitsLegacyGateState struct {
	mu                      sync.Mutex
	cond                    *sync.Cond
	readers, writersWaiting int
	exclusive               bool
}

var managementTraitsLegacyGate = func() *managementTraitsLegacyGateState {
	g := &managementTraitsLegacyGateState{}
	g.cond = sync.NewCond(&g.mu)
	return g
}()

type ManagementTraitsLegacyLease struct {
	gate     *managementTraitsLegacyGateState
	released bool // guarded by gate.mu, including the Fork/Release race
}

func LockManagementTraitsRuntimeFreeze() func() {
	g := managementTraitsLegacyGate
	g.mu.Lock()
	g.writersWaiting++
	for g.exclusive || g.readers > 0 {
		g.cond.Wait()
	}
	g.writersWaiting--
	g.exclusive = true
	g.mu.Unlock()
	var once sync.Once
	return func() { once.Do(func() { g.mu.Lock(); g.exclusive = false; g.cond.Broadcast(); g.mu.Unlock() }) }
}

func LockManagementTraitsLegacyMutation() *ManagementTraitsLegacyLease {
	g := managementTraitsLegacyGate
	g.mu.Lock()
	for g.exclusive || g.writersWaiting > 0 {
		g.cond.Wait()
	}
	g.readers++
	g.mu.Unlock()
	return &ManagementTraitsLegacyLease{gate: g}
}

func (l *ManagementTraitsLegacyLease) Release() {
	if l == nil || l.gate == nil {
		return
	}
	g := l.gate
	g.mu.Lock()
	if !l.released {
		l.released = true
		g.readers--
		g.cond.Broadcast()
	}
	g.mu.Unlock()
}

// Register a handoff BEFORE releasing the request lease/sending the response.
// A live lease can extend its own read ownership even with a writer queued.
func (l *ManagementTraitsLegacyLease) Fork() func() {
	if l == nil || l.gate == nil {
		return nil
	}
	g := l.gate
	g.mu.Lock()
	if l.released {
		g.mu.Unlock()
		return nil
	}
	g.readers++
	g.mu.Unlock()
	child := &ManagementTraitsLegacyLease{gate: g}
	return child.Release
}
