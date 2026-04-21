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
