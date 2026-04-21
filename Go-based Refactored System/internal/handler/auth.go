package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(s *service.AuthService) *AuthHandler { return &AuthHandler{svc: s} }

func (h *AuthHandler) CaptchaImage(c *gin.Context) {
	u, img, err := h.svc.GenerateCaptcha(c.Request.Context())
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "操作成功", "captchaOnOff": true, "uuid": u, "img": img})
}

// Login POST /login — RuoYi 响应把 token 放在顶层字段
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 500, "msg": "参数错误"})
		return
	}
	tok, err := h.svc.Login(c.Request.Context(), req, c.Request)
	if err != nil {
		c.JSON(200, gin.H{"code": 500, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "操作成功", "token": tok})
}

func (h *AuthHandler) GetInfo(c *gin.Context) {
	lu := c.MustGet("loginUser").(*model.LoginUser)
	c.JSON(200, gin.H{
		"code":        200,
		"msg":         "操作成功",
		"user":        lu.User,
		"roles":       lu.Roles,
		"permissions": lu.Permissions,
	})
}

func (h *AuthHandler) GetRouters(c *gin.Context) {
	lu := c.MustGet("loginUser").(*model.LoginUser)
	routers, err := h.svc.BuildMenus(lu.UserID)
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, routers)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	_ = h.svc.Logout(c.Request.Context(), c.GetHeader("Authorization"))
	response.AjaxOK(c, nil)
}
