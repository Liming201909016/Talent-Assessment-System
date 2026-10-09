package service

import (
	"context"
	"errors"
	"time"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const managementTraitsResumeTTL = 5 * time.Minute

func managementTraitsResumeID(id string) bool {
	if !managementTraitsOpaqueID(id) {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// IssueCandidateResume is an administrator-only credential operation, not an
// identity search or a paper creation path. All database operations are reads.
func (s *ManagementTraitsRuntimeService) IssueCandidateResume(ctx context.Context, examID, participantID, paperID string) (ManagementTraitsIdentityResponse, error) {
	var out ManagementTraitsIdentityResponse
	if !managementTraitsResumeID(examID) || !managementTraitsResumeID(participantID) || !managementTraitsResumeID(paperID) {
		return out, ErrManagementTraitsRuntimeInvalid
	}
	if s == nil || s.db == nil {
		return out, ErrManagementTraitsRuntimeClosed
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return out, err
	}
	err := s.db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		paper, snapshot, bundle, _, owner, questions, err := s.loadRuntimePaper(ctx, tx, paperID)
		if err != nil || paper.ID != paperID || paper.ExamID != examID || snapshot.ExamID != examID || owner.Kind != "candidate" || owner.ID != participantID || owner.PaperID != paperID || paper.State != 1 || owner.EndTime != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		if snapshot.LimitTime == nil || !time.Now().Before(*snapshot.LimitTime) {
			return ErrManagementTraitsRuntimeExpired
		}
		// The initial loader reads under the paper lock; a current locking read
		// additionally closes the review-revocation race during token issuance.
		var lockedBundle model.ManagementTraitsDefinitionBundle
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", bundle.ID).Take(&lockedBundle).Error != nil || !managementTraitsRuntimeSourceTrusted(lockedBundle.Status) {
			return ErrManagementTraitsRuntimeInvalid
		}
		var profile model.ManagementTraitsExamProfile
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("exam_id = ?", examID).Take(&profile).Error != nil || profile.ExamID != examID || validateManagementTraitsRuntimeProfile(profile, lockedBundle, s.maxJSONBytes) != nil || !profile.FrozenAt.Equal(snapshot.SourceCapturedAt) || profile.MappingSHA != snapshot.MappingSHA || profile.MappingSnapshot != snapshot.MappingSnapshot || profile.FieldContract != snapshot.FieldContract || managementTraitsRuntimeMetadata(lockedBundle, snapshot, profile, paper, owner, s.maxJSONBytes) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		// Separate locking reads are required: MySQL does not lock both source
		// tables through the existing UNION owner projection. Do not ignore a
		// deleted/disabled or competing tester row on this exact paper.
		owners := make([]managementTraitsRuntimeOwner, 0, 3)
		testers := make([]managementTraitsRuntimeOwner, 0, 3)
		columns := "id, exam_id, paper_id, COALESCE(del_flag,-1) AS del_flag, end_time"
		if tx.Table("el_candidate").Clauses(clause.Locking{Strength: "UPDATE"}).Select("'candidate' AS kind, "+columns).Where("paper_id = ?", paperID).Limit(3).Find(&owners).Error != nil || tx.Table("el_tester").Clauses(clause.Locking{Strength: "UPDATE"}).Select("'tester' AS kind, COALESCE(status,'') AS status, "+columns).Where("paper_id = ?", paperID).Limit(3).Find(&testers).Error != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		owners = append(owners, testers...)
		if validateManagementTraitsRuntimeOwner(owners, "candidate", participantID, examID, paperID) != nil || owners[0].EndTime != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var exam model.Exam
		if tx.Select("id, is_open, assessment_type, scoring_mode, total_time").Where("id = ?", examID).Take(&exam).Error != nil || exam.ID != examID || exam.IsOpen != 1 || exam.AssessmentType != "legacy" || exam.ScoringMode != "legacy" || exam.TotalTime != 25 || validateManagementTraitsRuntimeBuckets(tx, paperID, questions, s.maxJSONBytes) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		var identity managementTraitsRuntimeIdentity
		if managementTraitsDecodeStrict([]byte(snapshot.ParticipantSnapshot), s.maxJSONBytes, &identity) != nil {
			return ErrManagementTraitsRuntimeInvalid
		}
		now := time.Now()
		expires := now.Add(managementTraitsResumeTTL).Unix()
		if expires > snapshot.LimitTime.Unix() {
			expires = snapshot.LimitTime.Unix()
		}
		if expires <= now.Unix() {
			return ErrManagementTraitsRuntimeExpired
		}
		token, err := CreateManagementTraitsRuntimeToken(s.secret, ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimePaperPurpose, ParticipantType: "candidate", ParticipantID: participantID, ExamID: examID, PaperID: paperID, ExpiresAt: expires}, now)
		if err != nil {
			return err
		}
		out = ManagementTraitsIdentityResponse{ID: participantID, ExamID: examID, PaperID: paperID, Name: identity.Name, ParticipantToken: token}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrManagementTraitsRuntimeExpired) {
			return ManagementTraitsIdentityResponse{}, ErrManagementTraitsRuntimeExpired
		}
		return ManagementTraitsIdentityResponse{}, ErrManagementTraitsRuntimeInvalid
	}
	return out, nil
}
