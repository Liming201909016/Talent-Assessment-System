package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AjaxResult 用于 RuoYi 核心接口（/login /getInfo /system/* 等）
// 前端兼容：code==200 视为成功
func Ajax(c *gin.Context, code int, msg string, data any) {
	m := gin.H{"code": code, "msg": msg}
	if data != nil {
		m["data"] = data
	}
	c.JSON(http.StatusOK, m)
}

func AjaxOK(c *gin.Context, data any) {
	Ajax(c, 200, "操作成功", data)
}

func AjaxErr(c *gin.Context, msg string) {
	Ajax(c, 500, msg, nil)
}

func AjaxUnauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = "认证失败，无法访问系统资源"
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": msg})
}

func AjaxForbidden(c *gin.Context, msg string) {
	if msg == "" {
		msg = "没有权限"
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": msg})
}

// ApiRest 业务模块（qu/paper/exam/tester 等）使用的额外包装
type ApiRest struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    any    `json:"data,omitempty"`
	Success bool   `json:"success"`
}

func Rest(c *gin.Context, data any) {
	c.JSON(http.StatusOK, ApiRest{Code: 0, Msg: "", Data: data, Success: true})
}

func RestErr(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, ApiRest{Code: 1, Msg: msg, Success: false})
}

// TableData 分页返回（RuoYi TableDataInfo 对齐）
type TableData struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Rows  any    `json:"rows"`
	Total int64  `json:"total"`
}

func Table(c *gin.Context, rows any, total int64) {
	c.JSON(http.StatusOK, TableData{Code: 200, Msg: "查询成功", Rows: rows, Total: total})
}
