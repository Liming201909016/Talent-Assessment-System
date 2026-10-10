package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
)

type ManagementTraitsFormalRegistryHandler struct {
	registry *service.ManagementTraitsFormalRegistry
}

func NewManagementTraitsFormalRegistryHandler(db *gorm.DB) *ManagementTraitsFormalRegistryHandler {
	return &ManagementTraitsFormalRegistryHandler{registry: service.NewManagementTraitsFormalRegistry(db)}
}

func (h *ManagementTraitsFormalRegistryHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/versions", h.List)
	group.POST("/assets/preview", h.PreviewAssets)
	group.POST("/versions/register", h.Register)
	group.POST("/versions/approve", h.Approve)
	group.POST("/versions/activate", h.Activate)
	group.POST("/versions/revoke", h.Revoke)
}

func (h *ManagementTraitsFormalRegistryHandler) PreviewAssets(c *gin.Context) {
	if _, ok := formalRegistryActor(c); !ok {
		return
	}
	body, ok := formalRegistryBody(c, "repoCode", "assetKey")
	if !ok {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：仅允许005产品代码及受控资产标识")
		return
	}
	data, err := h.registry.PreviewAssets(c.Request.Context(), body["repoCode"], body["assetKey"])
	if errors.Is(err, service.ErrManagementTraitsFormalInvalid) {
		managementTraitsRuntimeHTTPError(c, 400, "仅允许00501或00502及合法受控资产标识")
		return
	}
	formalRegistryRespond(c, data, err)
}

// Deliberately retain the existing management-traits administrator boundary.
// Ordinary exam/read permissions do not grant authority to sign either role.
func formalRegistryActor(c *gin.Context) (int64, bool) {
	value, exists := c.Get("loginUser")
	if !exists {
		managementTraitsRuntimeHTTPError(c, 401, "请先登录")
		return 0, false
	}
	u, ok := value.(*model.LoginUser)
	if ok && u != nil && u.UserID > 0 {
		if u.UserID == 1 {
			return u.UserID, true
		}
		for _, p := range u.Permissions {
			if p == "*:*:*" {
				return u.UserID, true
			}
		}
	}
	managementTraitsRuntimeHTTPError(c, 403, "仅授权管理员可管理正式版本及分别审批")
	return 0, false
}

func formalRegistryBody(c *gin.Context, names ...string) (map[string]string, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil || !utf8.Valid(raw) {
		return nil, false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	// The shared parser accepts strict exact-key, nonempty, bounded strings.
	return managementTraitsRuntimeBody(c, names...)
}

func formalRegistryRespond(c *gin.Context, data any, err error) {
	c.Header("Cache-Control", "no-store")
	if err != nil {
		status, message := 409, "正式版本校验失败：请核对资产、双签及版本状态"
		if errors.Is(err, service.ErrManagementTraitsFormalClosed) {
			status, message = 503, "正式版本登记未就绪：仅local配置且须安装独立完整结构"
		} else if errors.Is(err, service.ErrManagementTraitsFormalConflict) {
			message = "版本摘要或代际已变化，或已撤销；请刷新后重试"
		}
		managementTraitsRuntimeHTTPError(c, status, message)
		return
	}
	c.JSON(200, gin.H{"code": 0, "msg": "", "success": true, "data": data})
}

func (h *ManagementTraitsFormalRegistryHandler) List(c *gin.Context) {
	if _, ok := formalRegistryActor(c); !ok {
		return
	}
	if len(c.Request.URL.Query()) != 0 {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误")
		return
	}
	data, err := h.registry.List(c.Request.Context())
	formalRegistryRespond(c, data, err)
}

func (h *ManagementTraitsFormalRegistryHandler) Register(c *gin.Context) {
	actor, ok := formalRegistryActor(c)
	if !ok {
		return
	}
	body, ok := formalRegistryBody(c, "versionCode", "examId", "assetKey")
	if !ok {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：仅允许版本、已冻结测评及受控资产标识")
		return
	}
	data, err := h.registry.Register(c.Request.Context(), body["versionCode"], body["examId"], body["assetKey"], actor)
	formalRegistryRespond(c, data, err)
}

func (h *ManagementTraitsFormalRegistryHandler) Approve(c *gin.Context)  { h.change(c, "approve") }
func (h *ManagementTraitsFormalRegistryHandler) Activate(c *gin.Context) { h.change(c, "activate") }
func (h *ManagementTraitsFormalRegistryHandler) Revoke(c *gin.Context)   { h.change(c, "revoke") }
func (h *ManagementTraitsFormalRegistryHandler) change(c *gin.Context, action string) {
	actor, ok := formalRegistryActor(c)
	if !ok {
		return
	}
	names := []string{"versionId", "expectedIdentitySha", "expectedEpoch"}
	if action == "approve" {
		names = append(names, "kind")
	}
	if action == "revoke" {
		names = append(names, "reason")
	}
	body, ok := formalRegistryBody(c, names...)
	if !ok {
		managementTraitsRuntimeHTTPError(c, 400, "参数格式错误：审批主体和时间由服务器确定")
		return
	}
	epoch, err := strconv.ParseInt(body["expectedEpoch"], 10, 64)
	if err != nil || epoch <= 0 || strconv.FormatInt(epoch, 10) != body["expectedEpoch"] || (action == "approve" && body["kind"] != "content" && body["kind"] != "psychometrics") {
		managementTraitsRuntimeHTTPError(c, 400, "审批职责或版本代际格式错误")
		return
	}
	data, err := h.registry.Change(c.Request.Context(), body["versionId"], body["expectedIdentitySha"], epoch, action, body["kind"], body["reason"], actor)
	formalRegistryRespond(c, data, err)
}
