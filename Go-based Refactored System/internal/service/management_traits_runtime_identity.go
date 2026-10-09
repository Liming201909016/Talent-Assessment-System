package service

import (
	"context"
	"crypto/subtle"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ManagementTraitsCandidateIdentityRequest struct {
	ID             string  `json:"id"`
	ExamID         string  `json:"examId"`
	Name           string  `json:"name"`
	Gender         string  `json:"gender"`
	Telephone      string  `json:"telephone"`
	Affiliation    string  `json:"affiliation"`
	Post           string  `json:"post"`
	Age            *int    `json:"age"`
	Degree         *string `json:"degree"`
	Major          *string `json:"major"`
	StuFlag        *int    `json:"stuFlag"`
	InvalidPayload bool    `json:"-"`
}

type ManagementTraitsIdentityResponse struct {
	ID               string `json:"id"`
	ExamID           string `json:"examId"`
	PaperID          string `json:"paperId"`
	Name             string `json:"name"`
	ParticipantToken string `json:"participantToken"`
}

type managementTraitsIdentitySchema struct {
	present  bool
	run      bool
	draft    bool
	captures []string
}

// No Migrator.HasTable: a failed metadata query must never mean "legacy".
func probeManagementTraitsIdentitySchema(ctx context.Context, db *gorm.DB) (managementTraitsIdentitySchema, error) {
	var result managementTraitsIdentitySchema
	if db == nil || ctx == nil {
		return result, ErrManagementTraitsRuntimeInvalid
	}
	tablesToProbe := []string{
		(model.ManagementTraitsDefinitionBundle{}).TableName(),
		(model.ManagementTraitsExamProfile{}).TableName(),
		(model.ManagementTraitsPaperSnapshot{}).TableName(),
		(model.ManagementTraitsResultRun{}).TableName(),
		(model.ManagementTraitsPaperQuestionSnapshot{}).TableName(),
		(model.ManagementTraitsResultDimension{}).TableName(),
		(model.ManagementTraitsResultModule{}).TableName(),
	}
	counts := make([]int64, len(tablesToProbe))
	for i, table := range tablesToProbe {
		if db.WithContext(ctx).Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", table).Scan(&counts[i]).Error != nil || counts[i] < 0 || counts[i] > 1 {
			return result, ErrManagementTraitsRuntimeInvalid
		}
	}
	// Capture tables have no local model yet. Discover them rather than treating
	// their absence in the Go type registry as proof that no protection exists.
	var tables []struct {
		Name string `gorm:"column:table_name"`
	}
	if db.WithContext(ctx).Raw("SELECT table_name AS table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name, 7) = 'el_mng_' AND (table_name LIKE '%capture%' OR table_name = 'el_mng_exam_draft') ORDER BY table_name").Scan(&tables).Error != nil {
		return result, ErrManagementTraitsRuntimeInvalid
	}
	result.captures = make([]string, 0, len(tables))
	for _, table := range tables {
		if table.Name == "el_mng_exam_draft" {
			if result.draft {
				return result, ErrManagementTraitsRuntimeInvalid
			}
			result.draft = true
			continue
		}
		if !strings.HasPrefix(table.Name, "el_mng_") || !strings.Contains(table.Name, "capture") || len(table.Name) > 64 {
			return result, ErrManagementTraitsRuntimeInvalid
		}
		for _, c := range table.Name {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
				return result, ErrManagementTraitsRuntimeInvalid
			}
		}
		result.captures = append(result.captures, table.Name)
	}
	var core int64
	for _, count := range counts {
		core += count
	}
	if core != int64(len(tablesToProbe)) && (core != 0 || len(result.captures) != 0) {
		return result, ErrManagementTraitsRuntimeInvalid
	}
	result.present, result.run = core == int64(len(tablesToProbe)), counts[3] == 1
	return result, nil
}

func managementTraitsIdentityHasProfile(ctx context.Context, db *gorm.DB, examID string) (bool, error) {
	var count int64
	if db.WithContext(ctx).Model(&model.ManagementTraitsExamProfile{}).Select("count(*)").Where("exam_id = ?", examID).Scan(&count).Error != nil || count < 0 || count > 1 {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	return count == 1, nil
}

// true means the caller must not perform a legacy identity write. This includes
// an existing protected identity when the requested destination is legacy.
func ManagementTraitsIdentityScope(ctx context.Context, db *gorm.DB, examID, kind, participantID string) (bool, error) {
	if !managementTraitsOpaqueID(examID) || (kind != "candidate" && kind != "tester") || (participantID != "" && !managementTraitsOpaqueID(participantID)) {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	schema, err := probeManagementTraitsIdentitySchema(ctx, db)
	if err != nil {
		return false, err
	}
	if schema.draft {
		if CheckManagementTraitsDraftSchema(ctx, db) != nil {
			return false, ErrManagementTraitsDraftClosed
		}
		d, err := managementTraitsInstalledDraft(ctx, db, examID)
		if err != nil {
			return false, err
		}
		if d != nil {
			return true, nil
		}
	}
	if !schema.present {
		return managementTraitsIdentityParticipantScope(ctx, db, kind, participantID)
	}
	return managementTraitsIdentityScopeWithSchema(ctx, db, schema, examID, kind, participantID)
}

func managementTraitsIdentityParticipantScope(ctx context.Context, db *gorm.DB, kind, participantID string) (bool, error) {
	if participantID == "" {
		return false, nil
	}
	req := ManagementTraitsLegacyScopeRequest{}
	if kind == "candidate" {
		req.CandidateIDs = []string{participantID}
	} else {
		req.TesterIDs = []string{participantID}
	}
	protected, err := CheckManagementTraitsLegacyScope(ctx, db, req)
	if err != nil {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	return protected, nil
}

func managementTraitsIdentityScopeWithSchema(ctx context.Context, db *gorm.DB, schema managementTraitsIdentitySchema, examID, kind, participantID string) (bool, error) {
	protected, err := managementTraitsIdentityHasProfile(ctx, db, examID)
	if err != nil || participantID == "" {
		return protected, err
	}
	var owners []managementTraitsRuntimeOwner
	if db.WithContext(ctx).Table("el_"+kind).Select("id, exam_id, paper_id").Where("id = ?", participantID).Limit(2).Find(&owners).Error != nil || len(owners) > 1 {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	paperID := ""
	if len(owners) == 1 {
		if owners[0].ID != participantID || !managementTraitsOpaqueID(owners[0].ExamID) {
			return false, ErrManagementTraitsRuntimeInvalid
		}
		p, err := managementTraitsIdentityHasProfile(ctx, db, owners[0].ExamID)
		if err != nil {
			return false, err
		}
		protected, paperID = protected || p, owners[0].PaperID
	}
	for _, table := range append([]string{"el_mng_paper_snapshot"}, schema.captures...) {
		var count int64
		q := db.WithContext(ctx).Table(table).Where("(participant_type = ? AND participant_id = ?)", kind, participantID)
		if paperID != "" {
			q = q.Or("paper_id = ?", paperID)
		}
		if q.Select("count(*)").Scan(&count).Error != nil || count < 0 {
			return false, ErrManagementTraitsRuntimeInvalid
		}
		protected = protected || count > 0
	}
	if schema.run {
		var count int64
		q := db.WithContext(ctx).Model(&model.ManagementTraitsResultRun{}).Where("participant_type = ? AND participant_id = ?", kind, participantID)
		if paperID != "" {
			q = q.Or("paper_id = ?", paperID)
		}
		if q.Select("count(*)").Scan(&count).Error != nil || count < 0 {
			return false, ErrManagementTraitsRuntimeInvalid
		}
		protected = protected || count > 0
	}
	fullProtected, err := managementTraitsIdentityParticipantScope(ctx, db, kind, participantID)
	if err != nil {
		return false, ErrManagementTraitsRuntimeInvalid
	}
	return protected || fullProtected, nil
}

type managementTraitsIdentityAdmissionData struct {
	Exam    model.Exam
	Profile model.ManagementTraitsExamProfile
	Bundle  model.ManagementTraitsDefinitionBundle
}

func (s *ManagementTraitsRuntimeService) managementTraitsIdentityAdmission(ctx context.Context, tx *gorm.DB, examID string, open int) (managementTraitsIdentityAdmissionData, error) {
	var admission managementTraitsIdentityAdmissionData
	var exam model.Exam
	if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", examID).Take(&exam).Error != nil {
		return admission, ErrManagementTraitsRuntimeInvalid
	}
	if exam.ID != examID || exam.IsOpen != open || exam.AssessmentType != "legacy" || exam.ScoringMode != "legacy" {
		return admission, ErrManagementTraitsRuntimeInvalid
	}
	var profiles []model.ManagementTraitsExamProfile
	if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("exam_id = ?", examID).Limit(2).Find(&profiles).Error != nil || len(profiles) != 1 || profiles[0].ExamID != examID {
		return admission, ErrManagementTraitsRuntimeInvalid
	}
	profile := profiles[0]
	var bundle model.ManagementTraitsDefinitionBundle
	if tx.Where("id = ?", profile.BundleID).Take(&bundle).Error != nil {
		return admission, ErrManagementTraitsRuntimeInvalid
	}
	return managementTraitsIdentityAdmissionData{exam, profile, bundle}, nil
}

func (s *ManagementTraitsRuntimeService) managementTraitsNewIdentityFields(admission managementTraitsIdentityAdmissionData, now time.Time) (managementTraitsRuntimeFieldWire, error) {
	var fields managementTraitsRuntimeFieldWire
	e, p, b := admission.Exam, admission.Profile, admission.Bundle
	if e.State != 0 || (e.StartTime != nil && (e.StartTime.IsZero() || now.Before(*e.StartTime))) || (e.EndTime != nil && (e.EndTime.IsZero() || !now.Before(*e.EndTime))) || (e.StartTime != nil && e.EndTime != nil && !e.StartTime.Before(*e.EndTime)) || b.Status != "candidate-current-source" || validateManagementTraitsRuntimeProfile(p, b, s.maxJSONBytes) != nil || p.FrozenAt.After(now) || managementTraitsDecodeStrict([]byte(p.FieldContract), s.maxJSONBytes, &fields) != nil {
		return fields, ErrManagementTraitsRuntimeInvalid
	}
	return fields, nil
}

func managementTraitsOwnerIdentity(owner managementTraitsRuntimeOwner) managementTraitsRuntimeIdentity {
	return managementTraitsRuntimeIdentity{Name: owner.Name, Gender: owner.Gender, Telephone: owner.Telephone, Affiliation: owner.Affiliation, Post: owner.Post, Age: owner.Age, Degree: owner.Degree, Major: owner.Major, StuFlag: owner.StuFlag}
}

func managementTraitsProjectIdentity(fields managementTraitsRuntimeFieldWire, identity managementTraitsRuntimeIdentity) managementTraitsRuntimeIdentity {
	allowed := make(map[string]bool, len(fields.RequiredFields))
	for _, f := range fields.RequiredFields {
		allowed[f] = true
	}
	if !allowed["name"] {
		identity.Name = ""
	}
	if !allowed["gender"] {
		identity.Gender = ""
	}
	if !allowed["telephone"] {
		identity.Telephone = ""
	}
	if !allowed["affiliation"] {
		identity.Affiliation = ""
	}
	if !allowed["post"] {
		identity.Post = ""
	}
	if !allowed["age"] {
		identity.Age = nil
	}
	if !allowed["degree"] {
		identity.Degree = nil
	}
	if !allowed["major"] {
		identity.Major = nil
	}
	if !allowed["stuFlag"] {
		identity.StuFlag = nil
	}
	return identity
}

func managementTraitsIdentityLegal(fields managementTraitsRuntimeFieldWire, identity managementTraitsRuntimeIdentity) bool {
	if !managementTraitsRuntimeFields(fields.RequiredFields, identity) {
		return false
	}
	allowed := make(map[string]bool, len(fields.RequiredFields))
	for _, f := range fields.RequiredFields {
		allowed[f] = true
	}
	for key, value := range map[string]string{"name": identity.Name, "gender": identity.Gender, "telephone": identity.Telephone, "affiliation": identity.Affiliation, "post": identity.Post} {
		if !utf8.ValidString(value) || len(value) > 255 || strings.ContainsRune(value, 0) || (!allowed[key] && value != "") {
			return false
		}
	}
	for key, p := range map[string]*string{"degree": identity.Degree, "major": identity.Major} {
		if p != nil && (!allowed[key] || !utf8.ValidString(*p) || len(*p) > 255 || strings.ContainsRune(*p, 0)) {
			return false
		}
	}
	if identity.Age != nil && (!allowed["age"] || *identity.Age <= 0 || int64(*identity.Age) > 2147483647) {
		return false
	}
	if identity.StuFlag != nil && (!allowed["stuFlag"] || (*identity.StuFlag != 0 && *identity.StuFlag != 1)) {
		return false
	}
	return true
}

func (s *ManagementTraitsRuntimeService) managementTraitsIdentityResponse(owner managementTraitsRuntimeOwner, now time.Time, expiry int64) (ManagementTraitsIdentityResponse, error) {
	token, err := CreateManagementTraitsRuntimeToken(s.secret, ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: owner.Kind, ParticipantID: owner.ID, ExamID: owner.ExamID, ExpiresAt: expiry}, now)
	if err != nil {
		return ManagementTraitsIdentityResponse{}, err
	}
	return ManagementTraitsIdentityResponse{ID: owner.ID, ExamID: owner.ExamID, PaperID: owner.PaperID, Name: owner.Name, ParticipantToken: token}, nil
}

// handled=false is the only legacy fallback permission. An error, including a
// schema probe failure, must abort the handler before any old write.
func (s *ManagementTraitsRuntimeService) TryRegisterCandidateIdentity(ctx context.Context, req ManagementTraitsCandidateIdentityRequest, participantToken string) (ManagementTraitsIdentityResponse, bool, error) {
	var response ManagementTraitsIdentityResponse
	if s != nil && s.assemblyDisabled {
		return response, false, nil
	}
	if s == nil {
		return response, true, ErrManagementTraitsRuntimeInvalid
	}
	if !managementTraitsOpaqueID(req.ExamID) || (req.ID != "" && !managementTraitsOpaqueID(req.ID)) {
		return response, true, ErrManagementTraitsRuntimeInvalid
	}
	schema, err := probeManagementTraitsIdentitySchema(ctx, s.db)
	if err != nil {
		return response, err != nil, err
	}
	if schema.draft {
		if CheckManagementTraitsDraftSchema(ctx, s.db) != nil {
			return response, true, ErrManagementTraitsDraftClosed
		}
		d, err := managementTraitsInstalledDraft(ctx, s.db, req.ExamID)
		if err != nil || (d != nil && d.Lifecycle == "draft") {
			return response, true, ErrManagementTraitsRuntimeInvalid
		}
	}
	if !schema.present {
		protected, err := managementTraitsIdentityParticipantScope(ctx, s.db, "candidate", req.ID)
		if err != nil || protected {
			return response, true, ErrManagementTraitsRuntimeInvalid
		}
		return response, false, nil
	}
	handled, err := managementTraitsIdentityScopeWithSchema(ctx, s.db, schema, req.ExamID, "candidate", req.ID)
	if err != nil {
		return response, true, err
	}
	if !handled {
		if req.Telephone != "" {
			var owners []managementTraitsRuntimeOwner
			if s.db.WithContext(ctx).Table("el_candidate").Select("id, exam_id, paper_id").Where("exam_id = ? AND telephone = ?", req.ExamID, req.Telephone).Order("id").Limit(1).Find(&owners).Error != nil {
				return response, true, ErrManagementTraitsRuntimeInvalid
			}
			for _, owner := range owners {
				protected, err := ManagementTraitsIdentityScope(ctx, s.db, req.ExamID, "candidate", owner.ID)
				if err != nil || protected {
					return response, true, ErrManagementTraitsRuntimeInvalid
				}
			}
		}
		return response, false, nil
	}
	if s.secret == "" || s.maxJSONBytes <= 0 || req.InvalidPayload {
		return response, true, ErrManagementTraitsRuntimeInvalid
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return response, true, err
	}
	now := time.Now()
	identity := managementTraitsRuntimeIdentity{Name: req.Name, Gender: req.Gender, Telephone: req.Telephone, Affiliation: req.Affiliation, Post: req.Post, Age: req.Age, Degree: req.Degree, Major: req.Major, StuFlag: req.StuFlag}
	resumeExpiry := int64(0)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		admission, err := s.managementTraitsIdentityAdmission(ctx, tx, req.ExamID, 1)
		if err != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var owners []managementTraitsRuntimeOwner
		q := tx.Table("el_candidate").Select("id, exam_id, paper_id, name, gender, telephone, affiliation, post, age, degree, major, stu_flag, end_time, COALESCE(del_flag,-1) AS del_flag").Clauses(clause.Locking{Strength: "UPDATE"}).Where("exam_id = ? AND telephone = ?", req.ExamID, req.Telephone)
		if req.Telephone == "" {
			q = tx.Table("el_candidate").Select("id, exam_id, paper_id, name, gender, telephone, affiliation, post, age, degree, major, stu_flag, end_time, COALESCE(del_flag,-1) AS del_flag").Clauses(clause.Locking{Strength: "UPDATE"}).Where("1 = 0")
		}
		if req.ID != "" {
			q = q.Or("id = ?", req.ID)
		}
		if participantToken != "" {
			claims, e := ParseManagementTraitsRuntimeToken(s.secret, participantToken, ManagementTraitsRuntimeParticipantPurpose, time.Now())
			if e == nil && claims.ExamID == req.ExamID && claims.ParticipantType == "candidate" {
				q = tx.Table("el_candidate").Select("id, exam_id, paper_id, name, gender, telephone, affiliation, post, age, degree, major, stu_flag, end_time, COALESCE(del_flag,-1) AS del_flag").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", claims.ParticipantID)
			}
		}
		if q.Limit(2).Find(&owners).Error != nil || len(owners) > 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		now = time.Now()
		expiry := now.Add(30 * time.Minute).Unix()
		if len(owners) == 1 {
			owner := owners[0]
			owner.Kind = "candidate"
			claims, err := ParseManagementTraitsRuntimeToken(s.secret, participantToken, ManagementTraitsRuntimeParticipantPurpose, now)
			if err != nil || claims.ValidateBinding("candidate", owner.ID, req.ExamID, "") != nil || owner.ExamID != req.ExamID || owner.DelFlag != 0 || (req.ID != "" && req.ID != owner.ID) {
				return ErrManagementTraitsRuntimeToken
			}
			if owner.PaperID == "" {
				fields, e := s.managementTraitsNewIdentityFields(admission, now)
				if e != nil || owner.EndTime != nil || !managementTraitsIdentityLegal(fields, identity) || !reflect.DeepEqual(identity, managementTraitsProjectIdentity(fields, managementTraitsOwnerIdentity(owner))) {
					return ErrManagementTraitsRuntimeInvalid
				}
			} else {
				resumeExpiry = claims.ExpiresAt
			}
			response, err = s.managementTraitsIdentityResponse(owner, now, claims.ExpiresAt)
			return err
		}
		if req.ID != "" || participantToken != "" {
			return ErrManagementTraitsRuntimeToken
		}
		fields, e := s.managementTraitsNewIdentityFields(admission, now)
		if e != nil || !managementTraitsIdentityLegal(fields, identity) {
			return ErrManagementTraitsRuntimeInvalid
		}
		owner := managementTraitsRuntimeOwner{Kind: "candidate", ID: uuid.NewString(), ExamID: req.ExamID, Name: req.Name, Gender: req.Gender, Telephone: req.Telephone, Affiliation: req.Affiliation, Post: req.Post, Age: req.Age, Degree: req.Degree, Major: req.Major, StuFlag: req.StuFlag}
		response, err = s.managementTraitsIdentityResponse(owner, now, expiry)
		if err != nil {
			return err
		}
		// Explicit columns only: never insert password, token, paper, end or PDF.
		record := struct {
			ID          string
			ExamID      string
			Name        string
			Gender      string
			Telephone   string
			Affiliation string
			Post        string
			Age         *int
			Degree      *string
			Major       *string
			StuFlag     *int
			DelFlag     int
			CreateTime  time.Time
			UpdateTime  time.Time
		}{owner.ID, owner.ExamID, owner.Name, owner.Gender, owner.Telephone, owner.Affiliation, owner.Post, owner.Age, owner.Degree, owner.Major, owner.StuFlag, 0, now, now}
		if tx.Table("el_candidate").Create(&record).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		return nil
	})
	if err != nil {
		if err != ErrManagementTraitsRuntimeToken {
			err = ErrManagementTraitsRuntimeInvalid
		}
		return ManagementTraitsIdentityResponse{}, true, err
	}
	if response.PaperID != "" {
		if err := s.managementTraitsValidateIdentityResume(ctx, &response, "candidate", resumeExpiry); err != nil {
			return ManagementTraitsIdentityResponse{}, true, err
		}
	}
	return response, true, nil
}

// The handler passes the existing tester ID after identifier lookup, before
// its legacy update_time/exam_id write. Password is rechecked inside the lock.
func (s *ManagementTraitsRuntimeService) TryTesterIdentity(ctx context.Context, examID, participantID, password string) (ManagementTraitsIdentityResponse, bool, error) {
	var response ManagementTraitsIdentityResponse
	if s != nil && s.assemblyDisabled {
		return response, false, nil
	}
	if s == nil || !managementTraitsOpaqueID(participantID) {
		return response, true, ErrManagementTraitsRuntimeInvalid
	}
	handled, err := ManagementTraitsIdentityScope(ctx, s.db, examID, "tester", participantID)
	if err != nil || !handled {
		return response, err != nil, err
	}
	if s.secret == "" || s.maxJSONBytes <= 0 || password == "" || len(password) > 4096 {
		return response, true, ErrManagementTraitsRuntimeToken
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return response, true, err
	}
	now := time.Now()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		admission, err := s.managementTraitsIdentityAdmission(ctx, tx, examID, 2)
		if err != nil {
			return err
		}
		var testers []model.Tester
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", participantID).Limit(2).Find(&testers).Error != nil || len(testers) != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		now = time.Now()
		te := testers[0]
		if te.ID != participantID || te.ExamID == nil || *te.ExamID != examID || te.DelFlag == nil || *te.DelFlag != 0 || te.Status == nil || *te.Status != "0" || subtle.ConstantTimeCompare([]byte(te.Password), []byte(password)) != 1 {
			return ErrManagementTraitsRuntimeToken
		}
		value := func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		}
		identity := managementTraitsRuntimeIdentity{Name: te.Name, Gender: value(te.Gender), Telephone: value(te.Telephone), Affiliation: value(te.Affiliation), Post: value(te.Post), Age: te.Age, Degree: te.Degree, Major: te.Major, StuFlag: te.StuFlag}
		if value(te.PaperID) == "" {
			fields, e := s.managementTraitsNewIdentityFields(admission, now)
			identity = managementTraitsProjectIdentity(fields, identity)
			if e != nil || te.EndTime != nil || !managementTraitsIdentityLegal(fields, identity) {
				return ErrManagementTraitsRuntimeInvalid
			}
		}
		response, err = s.managementTraitsIdentityResponse(managementTraitsRuntimeOwner{Kind: "tester", ID: te.ID, ExamID: examID, PaperID: value(te.PaperID), Name: identity.Name}, now, now.Add(30*time.Minute).Unix())
		return err
	})
	if err != nil {
		if err != ErrManagementTraitsRuntimeToken {
			err = ErrManagementTraitsRuntimeInvalid
		}
		return ManagementTraitsIdentityResponse{}, true, err
	}
	if response.PaperID != "" {
		if err := s.managementTraitsValidateIdentityResume(ctx, &response, "tester", time.Now().Add(30*time.Minute).Unix()); err != nil {
			return ManagementTraitsIdentityResponse{}, true, err
		}
	}
	return response, true, nil
}

func (s *ManagementTraitsRuntimeService) managementTraitsValidateIdentityResume(ctx context.Context, response *ManagementTraitsIdentityResponse, kind string, expiry int64) error {
	// Never hold owner locks while acquiring the paper lock (submit uses paper -> owner).
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		paper, snapshot, _, _, owner, questions, err := s.loadRuntimePaper(ctx, tx, response.PaperID)
		if err != nil || owner.Kind != kind || owner.ID != response.ID || paper.ExamID != response.ExamID || validateManagementTraitsRuntimeBuckets(tx, paper.ID, questions, s.maxJSONBytes) != nil || expiry <= time.Now().Unix() {
			return ErrManagementTraitsRuntimeInvalid
		}
		var frozen managementTraitsRuntimeIdentity
		if managementTraitsDecodeStrict([]byte(snapshot.ParticipantSnapshot), s.maxJSONBytes, &frozen) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		response.Name = frozen.Name
		return nil
	})
}
