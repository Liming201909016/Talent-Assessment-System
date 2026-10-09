package service

import (
	"context"
	"crypto/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type managementTraitsRuntimePaperRecords struct {
	Paper           model.Paper
	Snapshot        model.ManagementTraitsPaperSnapshot
	Questions       []model.ManagementTraitsPaperQuestionSnapshot
	LegacyQuestions []model.PaperQu
	Buckets         []model.PaperQuAnswer
}

type ManagementTraitsRuntimePublicQuestion struct {
	ID               string                          `json:"id"`
	DisplayOrder     int                             `json:"displayOrder"`
	Content          string                          `json:"content"`
	Options          []ManagementTraitsRuntimeOption `json:"options"`
	SelectedOptionID *string                         `json:"selectedOptionId"`
}
type ManagementTraitsRuntimePaperDetail struct {
	PaperID    string                                  `json:"paperId"`
	ExamID     string                                  `json:"examId"`
	Title      string                                  `json:"title"`
	PaperToken string                                  `json:"paperToken,omitempty"`
	ServerTime time.Time                               `json:"serverTime"`
	StartedAt  time.Time                               `json:"startedAt"`
	Deadline   time.Time                               `json:"deadline"`
	ReminderAt time.Time                               `json:"reminderAt"`
	State      int                                     `json:"state"`
	Answered   int                                     `json:"answered"`
	Questions  []ManagementTraitsRuntimePublicQuestion `json:"questions"`
}

func buildManagementTraitsRuntimePaper(profile model.ManagementTraitsExamProfile, bundle model.ManagementTraitsDefinitionBundle, exam model.Exam, owner managementTraitsRuntimeOwner, paperID string, now time.Time, order []string, budget int) (managementTraitsRuntimePaperRecords, error) {
	var r managementTraitsRuntimePaperRecords
	if validateManagementTraitsRuntimeProfile(profile, bundle, budget) != nil || bundle.Status != "candidate-current-source" || exam.ID != profile.ExamID || exam.State != 0 || exam.AssessmentType != "legacy" || exam.ScoringMode != "legacy" || exam.TotalTime != 25 || owner.ExamID != exam.ID || !managementTraitsOpaqueID(owner.ID) || !managementTraitsOpaqueID(paperID) || owner.PaperID != "" || owner.EndTime != nil || owner.DelFlag != 0 || (owner.Kind != "candidate" && owner.Kind != "tester") || (owner.Kind == "tester" && owner.Status != "0") || now.Before(*profile.FrozenAt) || (exam.StartTime != nil && now.Before(*exam.StartTime)) || (exam.EndTime != nil && !now.Before(*exam.EndTime)) || len(order) != 140 {
		return r, ErrManagementTraitsRuntimeInvalid
	}
	manifest, err := DecodeManagementTraitsManifest([]byte(bundle.ScoringManifest), budget)
	if err != nil {
		return r, ErrManagementTraitsRuntimeInvalid
	}
	mapping, err := DecodeManagementTraitsMapping([]byte(profile.MappingSnapshot), manifest, budget)
	if err != nil {
		return r, ErrManagementTraitsRuntimeInvalid
	}
	var fields managementTraitsRuntimeFieldWire
	if managementTraitsDecodeStrict([]byte(profile.FieldContract), budget, &fields) != nil {
		return r, ErrManagementTraitsRuntimeInvalid
	}
	identity := managementTraitsProjectIdentity(fields, managementTraitsOwnerIdentity(owner))
	if !managementTraitsIdentityLegal(fields, identity) {
		return r, ErrManagementTraitsRuntimeInvalid
	}
	ic, err := managementTraitsCanonicalBytes(identity)
	if err != nil {
		return r, ErrManagementTraitsRuntimeInvalid
	}
	ec, err := managementTraitsCanonicalBytes(managementTraitsRuntimeEvidence{Schema: "mng-current-source-evidence-v1", Source: fields.Source, RepoCode: fields.RepoCode, CapturedAt: *profile.FrozenAt, ManifestSHA: bundle.ScoringManifestSHA, MappingSHA: profile.MappingSHA})
	if err != nil {
		return r, ErrManagementTraitsRuntimeInvalid
	}
	now = now.Truncate(time.Second)
	deadline := now.Add(25 * time.Minute)
	profileID := profile.ExamID
	r.Paper = model.Paper{ID: paperID, ExamID: exam.ID, UserID: owner.ID, Title: exam.Title, State: 1, TotalTime: 25, CreateTime: &now, UpdateTime: &now, LimitTime: &deadline}
	r.Snapshot = model.ManagementTraitsPaperSnapshot{PaperID: paperID, ExamID: exam.ID, ProfileExamID: &profileID, BundleID: bundle.ID, Source: "new_creation", EvidenceSnapshot: string(ec.JSON), EvidenceSHA: ec.SHA256, MappingSnapshot: profile.MappingSnapshot, MappingSHA: profile.MappingSHA, ScoringManifestSHA: bundle.ScoringManifestSHA, ParticipantType: owner.Kind, ParticipantID: owner.ID, ParticipantSnapshot: string(ic.JSON), FieldContract: profile.FieldContract, IdentitySource: "submitted_snapshot", SourceCapturedAt: *profile.FrozenAt, StartedAt: now, LimitTime: &deadline, CreatedAt: now}
	r.Questions = make([]model.ManagementTraitsPaperQuestionSnapshot, 0, 140)
	r.LegacyQuestions = make([]model.PaperQu, 0, 140)
	r.Buckets = make([]model.PaperQuAnswer, 0, 700)
	byID := make(map[string]ManagementTraitsMappedQuestion, 140)
	for _, q := range mapping.Questions {
		byID[q.SourceQuestionID] = q
	}
	seen := make(map[string]bool, 140)
	for i, id := range order {
		q, ok := byID[id]
		if !ok || seen[id] {
			return managementTraitsRuntimePaperRecords{}, ErrManagementTraitsRuntimeInvalid
		}
		seen[id] = true
		mq := manifest.Questions[q.Number-1]
		pqID := uuid.NewString()
		options, err := managementTraitsCanonicalBytes(q.Options)
		if err != nil {
			return managementTraitsRuntimePaperRecords{}, ErrManagementTraitsRuntimeInvalid
		}
		frozen, err := managementTraitsCanonicalBytes(managementTraitsFrozenQuestion{Schema: "mng-frozen-question-v1", ManifestSHA: bundle.ScoringManifestSHA, MappingSHA: profile.MappingSHA, Question: q, DimensionKey: mq.DimensionKey, Reverse: mq.Reverse, PaperQuestionID: pqID, DisplayOrder: i + 1})
		if err != nil {
			return managementTraitsRuntimePaperRecords{}, ErrManagementTraitsRuntimeInvalid
		}
		r.Questions = append(r.Questions, model.ManagementTraitsPaperQuestionSnapshot{ID: uuid.NewString(), PaperID: paperID, PaperQuestionID: pqID, SourceQuestionID: id, Number: q.Number, DisplayOrder: i + 1, DimensionKey: mq.DimensionKey, Reverse: mq.Reverse, Content: q.Content, OptionsSnapshot: string(options.JSON), ScoringSnapshotSHA: frozen.SHA256, CreatedAt: now})
		r.LegacyQuestions = append(r.LegacyQuestions, model.PaperQu{ID: pqID, PaperID: paperID, QuID: id, QuType: 1, Sort: i + 1})
		for _, o := range q.Options {
			r.Buckets = append(r.Buckets, model.PaperQuAnswer{ID: uuid.NewString(), PaperID: paperID, QuID: id, AnswerID: o.SourceOptionID, Sort: o.DisplayOrder, Score: o.Raw})
		}
	}
	if _, err := ValidateManagementTraitsStoredInput(bundle, r.Snapshot, r.Questions, budget); err != nil {
		return managementTraitsRuntimePaperRecords{}, ErrManagementTraitsRuntimeInvalid
	}
	return r, nil
}

func persistManagementTraitsRuntimePaper(tx *gorm.DB, r managementTraitsRuntimePaperRecords) error {
	if caps := managementTraitsSchemaCapabilities(tx); caps != nil {
		for _, q := range r.LegacyQuestions {
			if !managementTraitsSchemaIDFits(caps.Columns, "el_paper_qu.qu_id", q.QuID) || !managementTraitsSchemaIDFits(caps.Columns, "el_paper_qu_answer.qu_id", q.QuID) {
				return ErrManagementTraitsRuntimeInvalid
			}
		}
		for _, b := range r.Buckets {
			if !managementTraitsSchemaIDFits(caps.Columns, "el_paper_qu_answer.qu_id", b.QuID) || !managementTraitsSchemaIDFits(caps.Columns, "el_paper_qu_answer.answer_id", b.AnswerID) {
				return ErrManagementTraitsRuntimeInvalid
			}
		}
	}
	if tx.Create(&r.Paper).Error != nil || tx.CreateInBatches(r.LegacyQuestions, 100).Error != nil || tx.CreateInBatches(r.Buckets, 100).Error != nil || tx.Create(&r.Snapshot).Error != nil || tx.CreateInBatches(r.Questions, 100).Error != nil {
		return ErrManagementTraitsRuntimeInvalid
	}
	return nil
}

func (s *ManagementTraitsRuntimeService) CreatePaper(ctx context.Context, c ManagementTraitsRuntimeClaims) (ManagementTraitsRuntimePaperDetail, error) {
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return ManagementTraitsRuntimePaperDetail{}, err
	}
	if !c.valid() || c.Purpose != ManagementTraitsRuntimeParticipantPurpose || c.ExpiresAt <= time.Now().Unix() {
		return ManagementTraitsRuntimePaperDetail{}, ErrManagementTraitsRuntimeToken
	}
	paperID := ""
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exam model.Exam
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", c.ExamID).Take(&exam).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var owners []managementTraitsRuntimeOwner
		if tx.Table("el_"+c.ParticipantType).Clauses(clause.Locking{Strength: "UPDATE"}).Select("id,exam_id,paper_id,COALESCE(del_flag,-1) AS del_flag,name,COALESCE(gender,'') AS gender,COALESCE(telephone,'') AS telephone,COALESCE(affiliation,'') AS affiliation,COALESCE(post,'') AS post,age,degree,major,stu_flag,end_time").Where("id = ? AND exam_id = ?", c.ParticipantID, c.ExamID).Limit(2).Find(&owners).Error != nil || len(owners) != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		owner := owners[0]
		owner.Kind = c.ParticipantType
		if owner.ID != c.ParticipantID || owner.ExamID != c.ExamID || owner.DelFlag != 0 {
			return ErrManagementTraitsRuntimeInvalid
		}
		if owner.PaperID != "" {
			paperID = owner.PaperID
			return nil
		} // release owner lock BEFORE locking existing paper
		if owner.Kind == "tester" {
			var te model.Tester
			if tx.Select("id,status").Where("id = ?", owner.ID).Take(&te).Error != nil || te.Status == nil {
				return ErrManagementTraitsRuntimeInvalid
			}
			owner.Status = *te.Status
		}
		var profile model.ManagementTraitsExamProfile
		var bundle model.ManagementTraitsDefinitionBundle
		if tx.Where("exam_id = ?", exam.ID).Take(&profile).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var bundleErr error
		bundle, bundleErr = loadManagementTraitsWriteBundle(ctx, tx, profile.BundleID)
		if bundleErr != nil || validateManagementTraitsRuntimeProfile(profile, bundle, s.maxJSONBytes) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		m, err := DecodeManagementTraitsManifest([]byte(bundle.ScoringManifest), s.maxJSONBytes)
		if err != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		mapping, err := DecodeManagementTraitsMapping([]byte(profile.MappingSnapshot), m, s.maxJSONBytes)
		if err != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		order := make([]string, len(mapping.Questions))
		for i, q := range mapping.Questions {
			order[i] = q.SourceQuestionID
		}
		if ShuffleCompetencyQuestionIDs(order, rand.Reader) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		now := time.Now()
		if c.ExpiresAt <= now.Unix() {
			return ErrManagementTraitsRuntimeToken
		}
		r, err := buildManagementTraitsRuntimePaper(profile, bundle, exam, owner, uuid.NewString(), now, order, s.maxJSONBytes)
		if err != nil {
			return err
		}
		if err := persistManagementTraitsRuntimePaper(tx, r); err != nil {
			return err
		}
		q := tx.Table("el_"+owner.Kind).Where("id = ? AND exam_id = ? AND (paper_id IS NULL OR paper_id = '') AND end_time IS NULL", owner.ID, owner.ExamID).Updates(map[string]any{"paper_id": r.Paper.ID, "update_time": now})
		if q.Error != nil || q.RowsAffected != 1 {
			return ErrManagementTraitsRuntimeInvalid
		}
		paperID = r.Paper.ID
		return nil
	})
	if err != nil {
		return ManagementTraitsRuntimePaperDetail{}, ErrManagementTraitsRuntimeInvalid
	}
	// Paper detail takes the SAME paper lock as answer/submit. No lock inversion
	// with the owner-row update in submit, and no deadline reset on retries.
	pc := c
	pc.Purpose = ManagementTraitsRuntimePaperPurpose
	pc.PaperID = paperID
	return s.PaperDetail(ctx, pc)
}

func (s *ManagementTraitsRuntimeService) PaperDetail(ctx context.Context, c ManagementTraitsRuntimeClaims) (ManagementTraitsRuntimePaperDetail, error) {
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return ManagementTraitsRuntimePaperDetail{}, err
	}
	if !c.valid() || c.Purpose != ManagementTraitsRuntimePaperPurpose || c.ExpiresAt <= time.Now().Unix() {
		return ManagementTraitsRuntimePaperDetail{}, ErrManagementTraitsRuntimeToken
	}
	var out ManagementTraitsRuntimePaperDetail
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		paper, snapshot, _, _, owner, questions, err := s.loadRuntimePaper(ctx, tx, c.PaperID)
		if err != nil || c.ValidateBinding(owner.Kind, owner.ID, paper.ExamID, paper.ID) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		if validateManagementTraitsRuntimeBuckets(tx, paper.ID, questions, s.maxJSONBytes) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		now := time.Now()
		if c.ExpiresAt <= now.Unix() {
			return ErrManagementTraitsRuntimeToken
		}
		expiry := snapshot.LimitTime.Add(30 * time.Minute).Unix()
		if expiry <= now.Unix() {
			expiry = now.Add(5 * time.Minute).Unix()
		}
		pc := c
		pc.ExpiresAt = expiry
		token, err := CreateManagementTraitsRuntimeToken(s.secret, pc, now)
		if err != nil {
			return err
		}
		out = ManagementTraitsRuntimePaperDetail{PaperID: paper.ID, ExamID: paper.ExamID, Title: paper.Title, PaperToken: token, ServerTime: now, StartedAt: snapshot.StartedAt, Deadline: *snapshot.LimitTime, ReminderAt: snapshot.StartedAt.Add(20 * time.Minute), State: paper.State, Questions: make([]ManagementTraitsRuntimePublicQuestion, 0, 140)}
		sort.Slice(questions, func(i, j int) bool { return questions[i].DisplayOrder < questions[j].DisplayOrder })
		for _, q := range questions {
			options, err := managementTraitsRuntimePublicOptions(q.OptionsSnapshot, s.maxJSONBytes)
			if err != nil {
				return err
			}
			out.Questions = append(out.Questions, ManagementTraitsRuntimePublicQuestion{ID: q.PaperQuestionID, DisplayOrder: q.DisplayOrder, Content: q.Content, Options: options, SelectedOptionID: q.SelectedOptionID})
			if q.RawAnswer != nil {
				out.Answered++
			}
		}
		return nil
	})
	if err != nil {
		return ManagementTraitsRuntimePaperDetail{}, ErrManagementTraitsRuntimeInvalid
	}
	return out, nil
}

func (s *ManagementTraitsRuntimeService) ProfileDetail(ctx context.Context, examID string) (model.ManagementTraitsExamProfile, error) {
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return model.ManagementTraitsExamProfile{}, err
	}
	if !managementTraitsOpaqueID(examID) {
		return model.ManagementTraitsExamProfile{}, ErrManagementTraitsRuntimeInvalid
	}
	var p model.ManagementTraitsExamProfile
	var b model.ManagementTraitsDefinitionBundle
	if s.db.WithContext(ctx).Where("exam_id = ?", examID).Take(&p).Error != nil || s.db.WithContext(ctx).Where("id = ?", p.BundleID).Take(&b).Error != nil || validateManagementTraitsRuntimeProfile(p, b, s.maxJSONBytes) != nil {
		return model.ManagementTraitsExamProfile{}, ErrManagementTraitsRuntimeInvalid
	}
	return p, nil
}

func managementTraitsRuntimeSourceTrusted(status string) bool {
	return status == "candidate-current-source" || status == "retired"
}
