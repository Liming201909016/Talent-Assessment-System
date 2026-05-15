package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/redisx"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

type DictHandler struct{ db *gorm.DB }

func NewDictHandler(db *gorm.DB) *DictHandler { return &DictHandler{db: db} }

type sysDictData struct {
	DictCode  int64  `gorm:"column:dict_code;primaryKey" json:"dictCode"`
	DictSort  int    `gorm:"column:dict_sort"            json:"dictSort"`
	DictLabel string `gorm:"column:dict_label"           json:"dictLabel"`
	DictValue string `gorm:"column:dict_value"           json:"dictValue"`
	DictType  string `gorm:"column:dict_type"            json:"dictType"`
	CSSClass  string `gorm:"column:css_class"            json:"cssClass"`
	ListClass string `gorm:"column:list_class"           json:"listClass"`
	IsDefault string `gorm:"column:is_default"           json:"isDefault"`
	Status    string `gorm:"column:status"               json:"status"`
	Remark    string `gorm:"column:remark"               json:"remark"`
}

func (sysDictData) TableName() string { return "sys_dict_data" }

// GET /system/dict/data/type/:dictType  — 读取字典数据（带 Redis 缓存）
func (h *DictHandler) DataByType(c *gin.Context) {
	dictType := c.Param("dictType")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	key := redisx.SysDictKey + dictType

	if b, err := redisx.Client.Get(ctx, key).Bytes(); err == nil && len(b) > 0 {
		var rows []sysDictData
		if json.Unmarshal(b, &rows) == nil {
			response.AjaxOK(c, rows)
			return
		}
	}
	var rows []sysDictData
	h.db.Where("dict_type = ? AND status = '0'", dictType).Order("dict_sort").Find(&rows)
	if b, err := json.Marshal(rows); err == nil {
		redisx.Client.Set(ctx, key, b, time.Hour)
	}
	response.AjaxOK(c, rows)
}

// POST /system/dict/data/batch  — 批量读取多个字典（一次 RTT 替代 N 次）
// Body: {"types": ["el_anxiety_stu", "el_depression_stu", ...]}
// Resp: {"el_anxiety_stu": [...], "el_depression_stu": [...], ...}
// 用途：报告页 result.vue/result2.vue 启动时聚合 12/13 个 dict 调用
func (h *DictHandler) BatchData(c *gin.Context) {
	var req struct {
		Types []string `json:"types"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Types) == 0 {
		response.AjaxOK(c, map[string][]sysDictData{})
		return
	}
	if len(req.Types) > 50 {
		req.Types = req.Types[:50]
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	result := make(map[string][]sysDictData, len(req.Types))
	missing := make([]string, 0)

	// 1. 先批量从 Redis 取
	for _, t := range req.Types {
		key := redisx.SysDictKey + t
		if b, err := redisx.Client.Get(ctx, key).Bytes(); err == nil && len(b) > 0 {
			var rows []sysDictData
			if json.Unmarshal(b, &rows) == nil {
				result[t] = rows
				continue
			}
		}
		missing = append(missing, t)
	}

	// 2. 缺失的一次性 SQL 查（IN 子句）
	if len(missing) > 0 {
		var rows []sysDictData
		h.db.Where("dict_type IN ? AND status = '0'", missing).
			Order("dict_type, dict_sort").Find(&rows)
		// 按 dict_type 分组
		grouped := make(map[string][]sysDictData, len(missing))
		for _, r := range rows {
			grouped[r.DictType] = append(grouped[r.DictType], r)
		}
		// 写回 Redis 并填充结果（包含空数组以缓存"该字典不存在"避免穿透）
		for _, t := range missing {
			rs := grouped[t]
			if rs == nil {
				rs = []sysDictData{}
			}
			result[t] = rs
			if b, err := json.Marshal(rs); err == nil {
				redisx.Client.Set(ctx, redisx.SysDictKey+t, b, time.Hour)
			}
		}
	}

	response.AjaxOK(c, result)
}
