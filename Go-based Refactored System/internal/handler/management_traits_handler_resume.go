package handler

import (
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/service"
)

var managementTraitsResumeResourceID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func (h *ManagementTraitsRuntimeHandler) IssueCandidateResume(c *gin.Context) {
	if !managementTraitsRuntimeAdmin(c, "management-traits:candidate:resume") {
		return
	}
	body, ok := managementTraitsRuntimeBody(c, "examId", "participantId", "paperId")
	if !ok || !managementTraitsResumeResourceID.MatchString(body["examId"]) || !managementTraitsResumeResourceID.MatchString(body["participantId"]) || !managementTraitsResumeResourceID.MatchString(body["paperId"]) || len(c.Request.URL.Query()) != 0 {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：须指定测评、考生和试卷ID")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	data, err := h.svc.IssueCandidateResume(c.Request.Context(), body["examId"], body["participantId"], body["paperId"])
	managementTraitsRuntimeRespond(c, data, err)
}

// A detail read must not exchange a short credential for a longer one. Keep
// the same five claims/purpose and cap expiry to both input and server output.
func (h *ManagementTraitsRuntimeHandler) capPaperCredential(data service.ManagementTraitsRuntimePaperDetail, claims service.ManagementTraitsRuntimeClaims, err error) (service.ManagementTraitsRuntimePaperDetail, error) {
	if err != nil {
		return service.ManagementTraitsRuntimePaperDetail{}, err
	}
	now := time.Now()
	issued, err := service.ParseManagementTraitsRuntimeToken(h.secret, data.PaperToken, service.ManagementTraitsRuntimePaperPurpose, now)
	if err != nil || claims.ValidateBinding(issued.ParticipantType, issued.ParticipantID, data.ExamID, data.PaperID) != nil {
		return service.ManagementTraitsRuntimePaperDetail{}, service.ErrManagementTraitsRuntimeToken
	}
	if issued.ExpiresAt > claims.ExpiresAt {
		issued.ExpiresAt = claims.ExpiresAt
	}
	data.PaperToken, err = service.CreateManagementTraitsRuntimeToken(h.secret, issued, now)
	if err != nil {
		return service.ManagementTraitsRuntimePaperDetail{}, err
	}
	return data, nil
}
