package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

func managementTraitsIdentityService(db *gorm.DB, cfg *config.Config, runtimeServices ...*service.ManagementTraitsRuntimeService) *service.ManagementTraitsRuntimeService {
	secret := ""
	if cfg != nil {
		secret = cfg.Jwt.Secret
	}
	if len(runtimeServices) > 0 {
		if runtimeServices[0].MatchesDependencies(db, secret, 1<<20) {
			return runtimeServices[0]
		}
		return service.NewManagementTraitsRuntimeService(nil, "", 1<<20)
	}
	return service.NewManagementTraitsRuntimeService(db, secret, 1<<20)
}

func managementTraitsIdentityCandidateBody(raw []byte) (service.ManagementTraitsCandidateIdentityRequest, bool) {
	var req service.ManagementTraitsCandidateIdentityRequest
	if len(raw) > 16<<10 || !utf8.Valid(raw) {
		return req, false
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return req, false
	}
	fields := map[string]*string{"id": &req.ID, "examId": &req.ExamID, "name": &req.Name, "gender": &req.Gender, "telephone": &req.Telephone, "affiliation": &req.Affiliation, "post": &req.Post}
	textFields := map[string]**string{"degree": &req.Degree, "major": &req.Major}
	intFields := map[string]**int{"age": &req.Age, "stuFlag": &req.StuFlag}
	seen := make(map[string]bool, len(fields))
	for d.More() {
		token, err := d.Token()
		key, ok := token.(string)
		field, exists := fields[key]
		textField, textExists := textFields[key]
		intField, intExists := intFields[key]
		if err != nil || !ok || (!exists && !textExists && !intExists) || seen[key] {
			return req, false
		}
		seen[key] = true
		token, err = d.Token()
		if err != nil || token == nil {
			return req, false
		}
		if exists || textExists {
			value, ok := token.(string)
			if !ok {
				return req, false
			}
			if exists {
				*field = value
			} else if value != "" {
				*textField = &value
			}
			continue
		}
		var value string
		switch typed := token.(type) {
		case string:
			value = typed
			if value == "" {
				continue
			}
		case json.Number:
			value = string(typed)
		default:
			return req, false
		}
		number, err := strconv.ParseInt(value, 10, 32)
		if err != nil || strconv.FormatInt(number, 10) != value {
			return req, false
		}
		n := int(number)
		*intField = &n
	}
	token, err = d.Token()
	if err != nil || token != json.Delim('}') {
		return req, false
	}
	if _, err := d.Token(); err != io.EOF {
		return req, false
	}
	return req, seen["examId"]
}

// true means the response is complete; no legacy count, Save or token write may follow.
// The body is restored byte-for-byte for the nonnew legacy binder.
func (h *CandidateHandler) TryManagementTraitsCandidateSave(c *gin.Context) bool {
	if !h.managementTraitsRuntime.AssemblyEnabled() {
		return false
	}
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, (16<<10)+1))
	if err != nil || len(raw) > 16<<10 {
		response.RestErr(c, "参数错误")
		return true
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	var lookup struct {
		ID        string  `json:"id"`
		ExamID    string  `json:"examId"`
		Name      string  `json:"name"`
		Telephone *string `json:"telephone"`
	}
	if json.Unmarshal(raw, &lookup) != nil {
		response.RestErr(c, "参数错误")
		return true
	}
	if lookup.ExamID == "" {
		return false
	}
	protected, err := service.ManagementTraitsIdentityScope(c.Request.Context(), h.db, lookup.ExamID, "candidate", lookup.ID)
	if err != nil {
		response.RestErr(c, "管理特质身份校验失败，请重新确认测评和身份")
		return true
	}
	req := service.ManagementTraitsCandidateIdentityRequest{ID: lookup.ID, ExamID: lookup.ExamID, Name: lookup.Name}
	if lookup.Telephone != nil {
		req.Telephone = *lookup.Telephone
	}
	strict, ok := managementTraitsIdentityCandidateBody(raw)
	if ok {
		req = strict
	} else if protected {
		response.RestErr(c, "参数错误：管理特质仅允许配置的身份字段")
		return true
	} else {
		// If a profile appears between scope probing and the locked service
		// admission, invalid raw JSON must still never become a new registration.
		req.InvalidPayload = true
	}
	token := c.GetHeader("X-Management-Traits-Token")
	if token == "" {
		token = c.GetHeader("Authorization")
	}
	view, handled, err := h.managementTraitsRuntime.TryRegisterCandidateIdentity(c.Request.Context(), req, token)
	if err != nil {
		response.RestErr(c, "管理特质身份校验失败，请重新确认测评和身份")
		return true
	}
	if !handled {
		return false
	}
	response.Rest(c, view)
	return true
}

// Called only after the existing password comparison, before legacy exam rebinding.
func (h *TesterHandler) TryManagementTraitsTesterLogin(c *gin.Context, examID string, tester *model.Tester, password string) bool {
	if !h.managementTraitsRuntime.AssemblyEnabled() {
		return false
	}
	if tester == nil {
		response.AjaxErr(c, "管理特质身份校验失败")
		return true
	}
	if examID == "" {
		protected, err := service.CheckManagementTraitsLegacyScope(c.Request.Context(), h.db, service.ManagementTraitsLegacyScopeRequest{TesterIDs: []string{tester.ID}})
		if err != nil || protected {
			response.AjaxErr(c, "管理特质身份校验失败，请重新确认测评和身份")
			return true
		}
		return false
	}
	view, handled, err := h.managementTraitsRuntime.TryTesterIdentity(c.Request.Context(), examID, tester.ID, password)
	if err != nil {
		response.AjaxErr(c, "管理特质身份校验失败，请重新确认测评和身份")
		return true
	}
	if !handled {
		return false
	}
	response.AjaxOK(c, view)
	return true
}
