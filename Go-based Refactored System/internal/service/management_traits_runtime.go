package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

type ManagementTraitsRuntimeService struct {
	db               *gorm.DB
	secret           string
	maxJSONBytes     int
	assemblyDisabled bool
	schemaOnce       sync.Once
	schemaErr        error
	schemaCapacity   atomic.Pointer[managementTraitsRuntimeCapacity]
	raceObservation  managementRaceObservationConfig
}

func NewDisabledManagementTraitsRuntimeService() *ManagementTraitsRuntimeService {
	return &ManagementTraitsRuntimeService{assemblyDisabled: true}
}

func (s *ManagementTraitsRuntimeService) AssemblyEnabled() bool {
	return s != nil && !s.assemblyDisabled
}

func NewManagementTraitsRuntimeService(db *gorm.DB, secret string, maxJSONBytes int) *ManagementTraitsRuntimeService {
	s := &ManagementTraitsRuntimeService{secret: secret, maxJSONBytes: maxJSONBytes, raceObservation: loadManagementRaceObservationConfig()}
	if db != nil {
		guarded, err := managementTraitsSchemaGuardDB(db, nil)
		if err != nil {
			s.schemaErr = ErrManagementTraitsRuntimeInvalid
		} else {
			s.db = guarded.Set("mng:schema_capacity", &s.schemaCapacity).Session(&gorm.Session{})
		}
	}
	return s
}

// Dependency identity is checked at assembly; secrets are never exposed.
func (s *ManagementTraitsRuntimeService) MatchesDependencies(db *gorm.DB, secret string, maxJSONBytes int) bool {
	return s != nil && s.db != nil && db != nil && s.db.ConnPool == db.ConnPool && s.secret == secret && s.maxJSONBytes == maxJSONBytes
}

// Deliberately not configurable: no caller can bypass incomplete legacy guards
// by supplying a boolean, an environment flag or a valid scoring hash.
func (s *ManagementTraitsRuntimeService) CandidateRuntimeEnabled() bool { return false }

func (s *ManagementTraitsRuntimeService) FreezeProfile(ctx context.Context, examID string) (model.ManagementTraitsExamProfile, error) {
	if s == nil || s.db == nil {
		return model.ManagementTraitsExamProfile{}, ErrManagementTraitsRuntimeClosed
	}
	if err := s.CheckRuntimeSchema(ctx); err != nil {
		return model.ManagementTraitsExamProfile{}, err
	}
	release := LockManagementTraitsRuntimeFreeze()
	defer release()
	return s.freezeProfileTransaction(ctx, examID, time.Now)
}
