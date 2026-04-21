package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/response"
)

// Stub 用于尚未实现的接口，返回统一占位响应，便于前端流程不中断。
// 后续按模块逐个替换为真实实现。
func Stub(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, response.ApiRest{
			Code:    0,
			Msg:     "",
			Data:    gin.H{"_todo": name, "_note": "Go refactor stub — business logic pending"},
			Success: true,
		})
	}
}

// AjaxStub 返回 RuoYi AjaxResult 形态占位
func AjaxStub(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功", "data": nil, "_todo": name})
	}
}

// TableStub 返回分页占位
func TableStub(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "查询成功", "rows": []any{}, "total": 0, "_todo": name})
	}
}
