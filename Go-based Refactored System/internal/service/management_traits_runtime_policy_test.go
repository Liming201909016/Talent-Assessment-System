package service

import (
	"testing"
	"time"
)

func TestManagementTraitsRetiredContinuesFrozenPaperRevokedRejects(t *testing.T) {
	f, records, paper, profile, owner := managementRuntimeLoadFixture(t, false)
	f.Bundle.Status = "retired"
	if managementTraitsRuntimeMetadata(f.Bundle, f.Paper, profile, paper, owner, 1<<20) != nil {
		t.Fatal("retired interrupted already-frozen paper")
	}
	if _, err := validateManagementTraitsRuntimeResult(f.Bundle, f.Paper, f.Questions, records, 1<<20); err != nil {
		t.Fatal("retired invalidated immutable run", err)
	}
	f.Bundle.Status = "review-revoked"
	if managementTraitsRuntimeMetadata(f.Bundle, f.Paper, profile, paper, owner, 1<<20) == nil {
		t.Fatal("revoked source trusted")
	}
	if _, err := buildManagementTraitsRuntimeRecords(f.Bundle, f.Paper, f.Questions, "new-run", paper.CreateTime.Add(time.Minute), 1<<20); err == nil {
		t.Fatal("revoked source permitted new result")
	}
}
