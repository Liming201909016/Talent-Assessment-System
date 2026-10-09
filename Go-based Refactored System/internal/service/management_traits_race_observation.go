package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

type managementRaceCaller uint8

const (
	managementRaceParticipant managementRaceCaller = iota + 1
	managementRaceWorker
	managementRaceInternal
)

type managementRaceWorkerKey struct{}
type managementRaceObservationKey struct{}

// Immutable per service; no shared callback, registry, per-paper map or I/O
// inside a transaction. Opt-in is limited to one canonical random v4 paper.
type managementRaceObservationConfig struct {
	target   string
	session  string
	resource string
	origin   time.Time
	logger   *slog.Logger
}

func loadManagementRaceObservationConfig() managementRaceObservationConfig {
	return newManagementRaceObservationConfig(os.Getenv("MNG_RACE_OBSERVE_ENV"), os.Getenv("MNG_RACE_OBSERVE_PAPER_SHA256"), rand.Reader)
}

func newManagementRaceObservationConfig(environment, target string, random io.Reader) managementRaceObservationConfig {
	if environment != "local" && environment != "staging" {
		return managementRaceObservationConfig{}
	}
	b, err := hex.DecodeString(target)
	if err != nil || len(b) != sha256.Size || hex.EncodeToString(b) != target {
		return managementRaceObservationConfig{}
	}
	var nonce [32]byte
	if _, err := io.ReadFull(random, nonce[:]); err != nil {
		return managementRaceObservationConfig{}
	}
	return managementRaceObservationConfig{target: target, session: hex.EncodeToString(nonce[:16]), resource: hex.EncodeToString(nonce[16:]), origin: time.Now(), logger: slog.Default()}
}

type managementRaceObservationEvent struct {
	stage     string
	processNS int64
}

type managementRaceObservation struct {
	attempt  string
	resource string
	session  string
	caller   string
	origin   time.Time
	entryUTC string
	logger   *slog.Logger
	events   [10]managementRaceObservationEvent
	count    int
}

func (s *ManagementTraitsRuntimeService) startRaceObservation(ctx context.Context, paperID string, caller managementRaceCaller) (context.Context, *managementRaceObservation) {
	// Do not inherit an earlier attempt into unrelated operations.
	if ctx != nil && ctx.Value(managementRaceObservationKey{}) != nil {
		ctx = context.WithValue(ctx, managementRaceObservationKey{}, (*managementRaceObservation)(nil))
	}
	if s == nil || ctx == nil || s.raceObservation.target == "" {
		return ctx, nil
	}
	id, err := uuid.Parse(paperID)
	if err != nil || id.Version() != 4 || id.Variant() != uuid.RFC4122 || id.String() != paperID {
		return ctx, nil
	}
	hash := sha256.Sum256([]byte("mng-race-paper-v1\x00" + paperID))
	if hex.EncodeToString(hash[:]) != s.raceObservation.target {
		return ctx, nil
	}
	name := ""
	switch caller {
	case managementRaceParticipant:
		name = "http_participant"
	case managementRaceWorker:
		name = "expiry_worker"
	case managementRaceInternal:
		name = "trusted_internal"
	default:
		return ctx, nil
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return ctx, nil
	}
	cfg := s.raceObservation
	obs := &managementRaceObservation{attempt: hex.EncodeToString(nonce[:]), resource: cfg.resource, session: cfg.session, caller: name, origin: cfg.origin, entryUTC: time.Now().UTC().Format(time.RFC3339Nano), logger: cfg.logger}
	obs.mark("entry")
	return context.WithValue(ctx, managementRaceObservationKey{}, obs), obs
}

func managementRaceMark(ctx context.Context, stage string) {
	if ctx != nil {
		obs, _ := ctx.Value(managementRaceObservationKey{}).(*managementRaceObservation)
		obs.mark(stage)
	}
}

func (o *managementRaceObservation) mark(stage string) {
	if o == nil || o.count == len(o.events) {
		return
	}
	switch stage {
	case "entry", "settlement_entered", "transaction_entered", "paper_lock_wait", "paper_lock_acquired", "bundle_lock_wait", "bundle_lock_acquired", "body_returned", "transaction_returned":
		o.events[o.count] = managementRaceObservationEvent{stage: stage, processNS: time.Since(o.origin).Nanoseconds()}
		o.count++
	}
}

// Called only by the entry defer after Transaction has returned. Failed
// commit/rollback/BEGIN is deliberately not interpreted as a DB final state.
func (o *managementRaceObservation) emit(result ManagementTraitsRuntimeSubmission, err error) {
	if o == nil {
		return
	}
	outcome := "failed"
	if err == nil && o.count >= 2 && o.events[o.count-1].stage == "transaction_returned" && o.events[o.count-2].stage == "body_returned" {
		outcome = "committed_created"
		if result.Reused {
			outcome = "committed_reused"
		}
	}
	for i := 0; i < o.count; i++ {
		e := o.events[i]
		o.logger.Info("management_traits_settlement_observation", "session", o.session, "resource", o.resource, "attempt", o.attempt, "caller", o.caller, "stage", e.stage, "process_ns", e.processNS, "elapsed_ns", e.processNS-o.events[0].processNS, "entry_utc", o.entryUTC, "outcome", outcome)
	}
}
