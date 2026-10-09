package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const (
	ManagementTraitsRuntimeParticipantPurpose = "management_traits_participant"
	ManagementTraitsRuntimePaperPurpose       = "management_traits_paper"
)

var (
	ErrManagementTraitsRuntimeToken   = errors.New("management traits authentication rejected")
	ErrManagementTraitsRuntimeInvalid = errors.New("management traits data rejected")
	ErrManagementTraitsRuntimeClosed  = errors.New("management traits candidate runtime is closed pending legacy write protection")
	ErrManagementTraitsRuntimeExpired = errors.New("management traits paper expired")
	ErrManagementTraitsRuntimeMissing = errors.New("management traits requires all answers before deadline")
)

type ManagementTraitsRuntimeClaims struct {
	Purpose         string
	ParticipantType string
	ParticipantID   string
	ExamID          string
	PaperID         string
	ExpiresAt       int64
}

func (c ManagementTraitsRuntimeClaims) valid() bool {
	if (c.ParticipantType != "candidate" && c.ParticipantType != "tester") || !managementTraitsOpaqueID(c.ParticipantID) || !managementTraitsOpaqueID(c.ExamID) {
		return false
	}
	switch c.Purpose {
	case ManagementTraitsRuntimeParticipantPurpose:
		return c.PaperID == ""
	case ManagementTraitsRuntimePaperPurpose:
		return managementTraitsOpaqueID(c.PaperID)
	default:
		return false
	}
}

func (c ManagementTraitsRuntimeClaims) ValidateBinding(kind, participantID, examID, paperID string) error {
	if !c.valid() || c.ParticipantType != kind || c.ParticipantID != participantID || c.ExamID != examID || c.PaperID != paperID {
		return ErrManagementTraitsRuntimeToken
	}
	return nil
}

// Expiration is explicit and never silently refreshed by retries.
func CreateManagementTraitsRuntimeToken(secret string, c ManagementTraitsRuntimeClaims, now time.Time) (string, error) {
	if secret == "" || !c.valid() || c.ExpiresAt <= now.Unix() {
		return "", ErrManagementTraitsRuntimeToken
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"purpose": c.Purpose, "participant_type": c.ParticipantType, "participant_id": c.ParticipantID,
		"exam_id": c.ExamID, "paper_id": c.PaperID, "exp": c.ExpiresAt,
	}).SignedString([]byte(secret))
	if err != nil {
		return "", ErrManagementTraitsRuntimeToken
	}
	return token, nil
}

func ParseManagementTraitsRuntimeToken(secret, raw, purpose string, now time.Time) (ManagementTraitsRuntimeClaims, error) {
	if secret == "" || len(raw) > 4096 || (purpose != ManagementTraitsRuntimeParticipantPurpose && purpose != ManagementTraitsRuntimePaperPurpose) {
		return ManagementTraitsRuntimeClaims{}, ErrManagementTraitsRuntimeToken
	}
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "Bearer ")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS512 {
			return nil, ErrManagementTraitsRuntimeToken
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS512"}), jwt.WithJSONNumber(), jwt.WithExpirationRequired(), jwt.WithTimeFunc(func() time.Time { return now }))
	if err != nil || token == nil || !token.Valid {
		return ManagementTraitsRuntimeClaims{}, ErrManagementTraitsRuntimeToken
	}
	// ParseInt also rejects fractions/exponents/non-finite/out-of-range numbers.
	n, ok := claims["exp"].(json.Number)
	exp, err := strconv.ParseInt(string(n), 10, 64)
	if !ok || err != nil || exp <= now.Unix() {
		return ManagementTraitsRuntimeClaims{}, ErrManagementTraitsRuntimeToken
	}
	c := ManagementTraitsRuntimeClaims{Purpose: stringClaim(claims, "purpose"), ParticipantType: stringClaim(claims, "participant_type"),
		ParticipantID: stringClaim(claims, "participant_id"), ExamID: stringClaim(claims, "exam_id"), PaperID: stringClaim(claims, "paper_id"), ExpiresAt: exp}
	// Missing paper_id is not interchangeable with a required explicit empty claim.
	if _, ok := claims["paper_id"].(string); !ok || !c.valid() || c.Purpose != purpose {
		return ManagementTraitsRuntimeClaims{}, ErrManagementTraitsRuntimeToken
	}
	return c, nil
}

func managementTraitsRuntimeVersions(code string) (string, ManagementTraitsScoringVersions, error) {
	q, v := "", ManagementTraitsScoringVersions{Product: "mng-traits-v2", Scoring: "mng-percent-scoring-v1", Norm: "mng-norm-20260928-v1"}
	switch code {
	case "00201":
		q, v.Question = ManagementTraitsQuestionnaireStaff, "mng-staff-db-current-v1"
	case "00202":
		q, v.Question = ManagementTraitsQuestionnaireLeader, "mng-leader-db-current-v1"
	case "00501":
		q, v.Product, v.Question = ManagementTraitsQuestionnaireStaff, "mng-00501-v1", "mng-00501-db-current-v1"
	case "00502":
		q, v.Product, v.Question = ManagementTraitsQuestionnaireLeader, "mng-00502-v1", "mng-00502-db-current-v2"
	default:
		return "", ManagementTraitsScoringVersions{}, ErrManagementTraitsRuntimeInvalid
	}
	return q, v, nil
}

func validateManagementTraitsRuntimeVersions(questionnaire string, v ManagementTraitsScoringVersions) error {
	legacy00502 := ManagementTraitsScoringVersions{Product: "mng-00502-v1", Question: "mng-00502-db-current-v1", Scoring: "mng-percent-scoring-v1", Norm: "mng-norm-20260928-v1"}
	if questionnaire == ManagementTraitsQuestionnaireLeader && v == legacy00502 {
		return nil
	}
	for _, code := range []string{"00201", "00202", "00501", "00502"} {
		q, expected, err := managementTraitsRuntimeVersions(code)
		if err == nil && questionnaire == q && v == expected {
			return nil
		}
	}
	return ErrManagementTraitsRuntimeInvalid
}
