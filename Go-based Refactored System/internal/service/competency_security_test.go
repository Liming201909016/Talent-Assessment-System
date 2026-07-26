package service

import (
	"reflect"
	"strings"
	"testing"
	"time"

	jwtpkg "github.com/talent-assessment/refactored/pkg/jwt"
)

const competencyTestSecret = "competency-test-secret-at-least-32-bytes"

func TestValidateAssessmentMode(t *testing.T) {
	tests := []struct {
		name       string
		assessment string
		scoring    string
		wantErr    bool
		competency bool
	}{
		{"legacy pair", AssessmentTypeLegacy, ScoringModeLegacy, false, false},
		{"competency pair", AssessmentTypeCompetency, ScoringModeCompetencyAverage, false, true},
		{"competency with legacy scoring", AssessmentTypeCompetency, ScoringModeLegacy, true, false},
		{"legacy with competency scoring", AssessmentTypeLegacy, ScoringModeCompetencyAverage, true, false},
		{"unknown assessment type", "unknown", ScoringModeLegacy, true, false},
		{"empty assessment type", "", ScoringModeLegacy, true, false},
		{"unknown scoring mode", AssessmentTypeLegacy, "unknown", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCompetency, err := ValidateAssessmentMode(tt.assessment, tt.scoring)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateAssessmentMode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotCompetency != tt.competency {
				t.Fatalf("ValidateAssessmentMode() competency = %v, want %v", gotCompetency, tt.competency)
			}
		})
	}
}

func TestValidateCompetencyReportAudience(t *testing.T) {
	tests := []struct {
		name     string
		audience string
		wantErr  bool
	}{
		{"frontline employee", CompetencyReportAudienceFrontlineEmployee, false},
		{"leader", CompetencyReportAudienceLeader, false},
		{"empty", "", true},
		{"unknown", "manager", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCompetencyReportAudience(tt.audience); (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCompetencyReportAudience() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCompetencyDimensionIDs(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr bool
	}{
		{"one dimension", []string{"dimension-01"}, []string{"dimension-01"}, false},
		{"multiple dimensions preserve order", []string{"dimension-02", "dimension-01"}, []string{"dimension-02", "dimension-01"}, false},
		{"empty list", []string{}, nil, true},
		{"nil list", nil, nil, true},
		{"empty id", []string{"dimension-01", ""}, nil, true},
		{"blank id", []string{"dimension-01", "  "}, nil, true},
		{"duplicate id", []string{"dimension-01", "dimension-01"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateCompetencyDimensionIDs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCompetencyDimensionIDs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ValidateCompetencyDimensionIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompetencyToken_RoundTrip(t *testing.T) {
	claims := CompetencyTokenClaims{
		Purpose:         CompetencyTokenPurposePaper,
		ParticipantType: CompetencyParticipantCandidate,
		ParticipantID:   "candidate-1",
		ExamID:          "exam-1",
		PaperID:         "paper-1",
	}
	token, err := CreateCompetencyToken(competencyTestSecret, claims, 15*time.Minute)
	if err != nil {
		t.Fatalf("CreateCompetencyToken() error = %v", err)
	}

	got, err := ParseCompetencyToken(competencyTestSecret, "Bearer "+token, CompetencyTokenPurposePaper)
	if err != nil {
		t.Fatalf("ParseCompetencyToken() error = %v", err)
	}
	if got.ParticipantID != claims.ParticipantID || got.ExamID != claims.ExamID || got.PaperID != claims.PaperID {
		t.Fatalf("parsed claims mismatch: got %+v, want %+v", got, claims)
	}
	if got.ExpiresAt <= time.Now().Unix() {
		t.Fatalf("token expiration was not set in the future: %d", got.ExpiresAt)
	}
}

func TestCreateCompetencyToken_ValidatesRequiredClaims(t *testing.T) {
	tests := []struct {
		name   string
		claims CompetencyTokenClaims
		ttl    time.Duration
	}{
		{"missing purpose", CompetencyTokenClaims{ParticipantType: CompetencyParticipantCandidate, ParticipantID: "p", ExamID: "e"}, time.Minute},
		{"invalid purpose", CompetencyTokenClaims{Purpose: "admin", ParticipantType: CompetencyParticipantCandidate, ParticipantID: "p", ExamID: "e"}, time.Minute},
		{"invalid participant type", CompetencyTokenClaims{Purpose: CompetencyTokenPurposeParticipant, ParticipantType: "admin", ParticipantID: "p", ExamID: "e"}, time.Minute},
		{"missing participant", CompetencyTokenClaims{Purpose: CompetencyTokenPurposeParticipant, ParticipantType: CompetencyParticipantCandidate, ExamID: "e"}, time.Minute},
		{"missing exam", CompetencyTokenClaims{Purpose: CompetencyTokenPurposeParticipant, ParticipantType: CompetencyParticipantCandidate, ParticipantID: "p"}, time.Minute},
		{"paper purpose missing paper", CompetencyTokenClaims{Purpose: CompetencyTokenPurposePaper, ParticipantType: CompetencyParticipantCandidate, ParticipantID: "p", ExamID: "e"}, time.Minute},
		{"non-positive ttl", CompetencyTokenClaims{Purpose: CompetencyTokenPurposeParticipant, ParticipantType: CompetencyParticipantCandidate, ParticipantID: "p", ExamID: "e"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := CreateCompetencyToken(competencyTestSecret, tt.claims, tt.ttl); err == nil {
				t.Fatal("CreateCompetencyToken() expected error, got nil")
			}
		})
	}
}

func TestParseCompetencyToken_RejectsInvalidTokens(t *testing.T) {
	validClaims := CompetencyTokenClaims{
		Purpose:         CompetencyTokenPurposePaper,
		ParticipantType: CompetencyParticipantTester,
		ParticipantID:   "tester-1",
		ExamID:          "exam-1",
		PaperID:         "paper-1",
	}
	validToken, err := CreateCompetencyToken(competencyTestSecret, validClaims, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	noExpirationToken, _ := jwtpkg.Create(competencyTestSecret, map[string]any{
		"purpose":          CompetencyTokenPurposePaper,
		"participant_type": CompetencyParticipantTester,
		"participant_id":   "tester-1",
		"exam_id":          "exam-1",
		"paper_id":         "paper-1",
	})
	expiredToken, _ := jwtpkg.Create(competencyTestSecret, map[string]any{
		"purpose":          CompetencyTokenPurposePaper,
		"participant_type": CompetencyParticipantTester,
		"participant_id":   "tester-1",
		"exam_id":          "exam-1",
		"paper_id":         "paper-1",
		"exp":              time.Now().Add(-time.Minute).Unix(),
	})

	tests := []struct {
		name    string
		token   string
		secret  string
		purpose string
	}{
		{"missing token", "", competencyTestSecret, CompetencyTokenPurposePaper},
		{"malformed token", "not-a-token", competencyTestSecret, CompetencyTokenPurposePaper},
		{"wrong secret", validToken, "wrong-secret", CompetencyTokenPurposePaper},
		{"missing expiration", noExpirationToken, competencyTestSecret, CompetencyTokenPurposePaper},
		{"expired token", expiredToken, competencyTestSecret, CompetencyTokenPurposePaper},
		{"wrong purpose", validToken, competencyTestSecret, CompetencyTokenPurposeParticipant},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCompetencyToken(tt.secret, tt.token, tt.purpose)
			if err == nil {
				t.Fatal("ParseCompetencyToken() expected error, got nil")
			}
			leaksToken := tt.token != "" && strings.Contains(err.Error(), tt.token)
			leaksSecret := tt.secret != "" && strings.Contains(err.Error(), tt.secret)
			if leaksToken || leaksSecret {
				t.Fatalf("error leaks token or secret: %v", err)
			}
		})
	}
}

func TestCompetencyTokenClaims_MatchesBinding(t *testing.T) {
	claims := CompetencyTokenClaims{
		Purpose:         CompetencyTokenPurposePaper,
		ParticipantType: CompetencyParticipantCandidate,
		ParticipantID:   "candidate-1",
		ExamID:          "exam-1",
		PaperID:         "paper-1",
	}

	if err := claims.ValidateBinding("candidate-1", "exam-1", "paper-1"); err != nil {
		t.Fatalf("ValidateBinding() valid binding error = %v", err)
	}
	for _, tc := range []struct {
		name, participantID, examID, paperID string
	}{
		{"wrong participant", "candidate-2", "exam-1", "paper-1"},
		{"wrong exam", "candidate-1", "exam-2", "paper-1"},
		{"wrong paper", "candidate-1", "exam-1", "paper-2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := claims.ValidateBinding(tc.participantID, tc.examID, tc.paperID); err == nil {
				t.Fatal("ValidateBinding() expected error, got nil")
			}
		})
	}
}
