package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	jwtpkg "github.com/talent-assessment/refactored/pkg/jwt"
)

const (
	AssessmentTypeLegacy     = "legacy"
	AssessmentTypeCompetency = "competency"

	ScoringModeLegacy            = "legacy"
	ScoringModeCompetencyAverage = "competency_average"

	CompetencyReportAudienceFrontlineEmployee = "frontline_employee"
	CompetencyReportAudienceLeader            = "leader"

	CompetencyTokenPurposeParticipant = "competency_participant"
	CompetencyTokenPurposePaper       = "competency_paper"

	CompetencyParticipantCandidate = "candidate"
	CompetencyParticipantTester    = "tester"
)

var errInvalidCompetencyToken = errors.New("invalid competency token")

// ValidateAssessmentMode validates the only supported assessment/scoring pairs.
// The bool result is true only for the competency flow.
func ValidateAssessmentMode(assessmentType, scoringMode string) (bool, error) {
	switch {
	case assessmentType == AssessmentTypeLegacy && scoringMode == ScoringModeLegacy:
		return false, nil
	case assessmentType == AssessmentTypeCompetency && scoringMode == ScoringModeCompetencyAverage:
		return true, nil
	default:
		return false, fmt.Errorf("invalid assessment type and scoring mode combination")
	}
}

// ValidateCompetencyReportAudience validates the report wording profile chosen
// while configuring a competency exam. It does not alter questions or scoring.
func ValidateCompetencyReportAudience(audience string) error {
	switch audience {
	case CompetencyReportAudienceFrontlineEmployee, CompetencyReportAudienceLeader:
		return nil
	default:
		return errors.New("invalid competency report audience")
	}
}

// ValidateCompetencyDimensionIDs validates the dimension selection at the HTTP
// entry boundary. Database existence/status/question-count checks follow in the
// save transaction.
func ValidateCompetencyDimensionIDs(dimensionIDs []string) ([]string, error) {
	if len(dimensionIDs) == 0 {
		return nil, errors.New("at least one competency dimension is required")
	}
	seen := make(map[string]struct{}, len(dimensionIDs))
	result := make([]string, 0, len(dimensionIDs))
	for _, rawID := range dimensionIDs {
		id := strings.TrimSpace(rawID)
		if id == "" {
			return nil, errors.New("competency dimension id is required")
		}
		if _, exists := seen[id]; exists {
			return nil, errors.New("duplicate competency dimension")
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

// CompetencyTokenClaims binds a short-lived token to one participant and exam.
// Paper-purpose tokens are additionally bound to one paper.
type CompetencyTokenClaims struct {
	Purpose         string
	ParticipantType string
	ParticipantID   string
	ExamID          string
	PaperID         string
	ExpiresAt       int64
}

func (c CompetencyTokenClaims) validate() error {
	if c.Purpose != CompetencyTokenPurposeParticipant && c.Purpose != CompetencyTokenPurposePaper {
		return errInvalidCompetencyToken
	}
	if c.ParticipantType != CompetencyParticipantCandidate && c.ParticipantType != CompetencyParticipantTester {
		return errInvalidCompetencyToken
	}
	if c.ParticipantID == "" || c.ExamID == "" {
		return errInvalidCompetencyToken
	}
	if c.Purpose == CompetencyTokenPurposePaper && c.PaperID == "" {
		return errInvalidCompetencyToken
	}
	return nil
}

// ValidateBinding prevents participant, exam, or paper tokens from being reused
// against another resource.
func (c CompetencyTokenClaims) ValidateBinding(participantID, examID, paperID string) error {
	if c.ParticipantID != participantID || c.ExamID != examID || c.PaperID != paperID {
		return errInvalidCompetencyToken
	}
	return nil
}

// CreateCompetencyToken creates an expiring HS512 token for the competency flow.
func CreateCompetencyToken(secret string, claims CompetencyTokenClaims, ttl time.Duration) (string, error) {
	if secret == "" || ttl <= 0 {
		return "", errInvalidCompetencyToken
	}
	if err := claims.validate(); err != nil {
		return "", err
	}
	now := time.Now()
	claims.ExpiresAt = now.Add(ttl).Unix()
	return jwtpkg.Create(secret, map[string]any{
		"purpose":          claims.Purpose,
		"participant_type": claims.ParticipantType,
		"participant_id":   claims.ParticipantID,
		"exam_id":          claims.ExamID,
		"paper_id":         claims.PaperID,
		"iat":              now.Unix(),
		"exp":              claims.ExpiresAt,
	})
}

// ParseCompetencyToken verifies signature, expiration, purpose, and required
// resource-binding claims. Returned errors never include the token or secret.
func ParseCompetencyToken(secret, token, expectedPurpose string) (CompetencyTokenClaims, error) {
	if secret == "" || strings.TrimSpace(token) == "" {
		return CompetencyTokenClaims{}, errInvalidCompetencyToken
	}
	raw := strings.TrimSpace(token)
	if strings.HasPrefix(raw, "Bearer ") {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
	}
	claims, err := jwtpkg.Parse(secret, raw)
	if err != nil {
		return CompetencyTokenClaims{}, errInvalidCompetencyToken
	}

	parsed := CompetencyTokenClaims{
		Purpose:         stringClaim(claims, "purpose"),
		ParticipantType: stringClaim(claims, "participant_type"),
		ParticipantID:   stringClaim(claims, "participant_id"),
		ExamID:          stringClaim(claims, "exam_id"),
		PaperID:         stringClaim(claims, "paper_id"),
		ExpiresAt:       int64Claim(claims, "exp"),
	}
	if err := parsed.validate(); err != nil || parsed.Purpose != expectedPurpose {
		return CompetencyTokenClaims{}, errInvalidCompetencyToken
	}
	if parsed.ExpiresAt <= time.Now().Unix() {
		return CompetencyTokenClaims{}, errInvalidCompetencyToken
	}
	return parsed, nil
}

func stringClaim(claims map[string]any, key string) string {
	value, _ := claims[key].(string)
	return value
}

func int64Claim(claims map[string]any, key string) int64 {
	switch value := claims[key].(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	case int:
		return int64(value)
	default:
		return 0
	}
}
