package handler

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// Repo 题库表（qu_repo）CRUD —— 业务模块范例
// 业务模块统一用 ApiRest 包装。
type RepoHandler struct{ db *gorm.DB }

func NewRepoHandler(db *gorm.DB) *RepoHandler { return &RepoHandler{db: db} }

type quRepo struct {
	ID         string     `gorm:"column:id;primaryKey" json:"id"`
	Code       string     `gorm:"column:code"          json:"code"`
	Title      string     `gorm:"column:title"         json:"title"`
	RadioCount int        `gorm:"column:radio_count"   json:"radioCount"`
	MultiCount int        `gorm:"column:multi_count"   json:"multiCount"`
	JudgeCount int        `gorm:"column:judge_count"   json:"judgeCount"`
	Remark     string     `gorm:"column:remark"        json:"remark"`
	CreateTime *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time"   json:"updateTime"`
}

func (quRepo) TableName() string { return "el_repo" }

// POST /api/qu/repo/paging
func (h *RepoHandler) Paging(c *gin.Context) {
	var req struct {
		Current int    `json:"current"`
		Size    int    `json:"size"`
		Title   string `json:"title"`
		Params  struct {
			Title string `json:"title"`
		} `json:"params"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	// 兼容两种传参：顶层 title 或 params.title
	searchTitle := req.Title
	if searchTitle == "" {
		searchTitle = req.Params.Title
	}
	q := h.db.Model(&quRepo{})
	if searchTitle != "" {
		q = q.Where("title like ?", "%"+searchTitle+"%")
	}
	var total int64
	q.Count(&total)
	var rows []quRepo
	q.Order("update_time desc").Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	response.Rest(c, gin.H{"records": rows, "total": total, "current": req.Current, "size": req.Size})
}

// POST /api/qu/repo/list
func (h *RepoHandler) List(c *gin.Context) {
	var rows []quRepo
	h.db.Order("update_time desc").Find(&rows)
	response.Rest(c, rows)
}

// POST /api/qu/repo/detail?id=xxx  or JSON body {"id":"xxx"}
func (h *RepoHandler) Detail(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		var b struct{ ID string `json:"id"` }
		_ = c.ShouldBindJSON(&b)
		id = b.ID
	}
	var r quRepo
	if err := h.db.Where("id = ?", id).First(&r).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	response.Rest(c, r)
}

// POST /api/qu/repo/save
func (h *RepoHandler) Save(c *gin.Context) {
	var r quRepo
	if err := c.ShouldBindJSON(&r); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if r.ID == "" {
		// 由 DB 默认值或调用方生成；若无则自行生成
		r.ID = strconv.FormatInt(nextID(), 10)
		if err := h.db.Create(&r).Error; err != nil {
			response.RestErr(c, err.Error())
			return
		}
	} else {
		if err := h.db.Save(&r).Error; err != nil {
			response.RestErr(c, err.Error())
			return
		}
	}
	response.Rest(c, r)
}

// POST /api/qu/repo/remove
func (h *RepoHandler) Remove(c *gin.Context) {
	var b struct {
		IDs []string `json:"ids"`
	}
	_ = c.ShouldBindJSON(&b)
	if len(b.IDs) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	if err := h.db.Model(&quRepo{}).Where("id IN ?", b.IDs).Delete(&quRepo{}).Error; err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, true)
}

// 原子 ID 生成：纳秒时间戳 + 原子计数器，避免并发冲突
var idCounter int64

func nextID() int64 {
	return time.Now().UnixNano() + atomic.AddInt64(&idCounter, 1)
}

// POST /exam/api/repo/batch-action
// 对齐 Java QuRepoService.batchAction：批量添加/移除题目到题库
func (h *RepoHandler) BatchAction(c *gin.Context) {
	var b struct {
		QuIDs   []string `json:"quIds"`
		RepoIDs []string `json:"repoIds"`
		Remove  *bool    `json:"remove"`
	}
	_ = c.ShouldBindJSON(&b)
	if len(b.QuIDs) == 0 || len(b.RepoIDs) == 0 {
		response.RestErr(c, "quIds 或 repoIds 为空")
		return
	}
	if b.Remove != nil && *b.Remove {
		// 移除
		h.db.Where("repo_id IN ? AND qu_id IN ?", b.RepoIDs, b.QuIDs).Delete(&quRepoRow{})
	} else {
		// 添加：对每个 quId 先删旧关联再重建
		for _, qid := range b.QuIDs {
			var qu struct{ QuType int `gorm:"column:qu_type"` }
			h.db.Table("el_qu").Where("id = ?", qid).Select("qu_type").Take(&qu)
			h.db.Where("qu_id = ?", qid).Delete(&quRepoRow{})
			for _, rid := range b.RepoIDs {
				h.db.Create(&quRepoRow{
					ID:     strconv.FormatInt(nextID(), 10),
					QuID:   qid,
					RepoID: rid,
					QuType: qu.QuType,
				})
			}
		}
	}
	// 重排 sort + 刷新统计
	for _, rid := range b.RepoIDs {
		var rows []quRepoRow
		h.db.Where("repo_id = ?", rid).Order("sort ASC").Find(&rows)
		for i := range rows {
			rows[i].Sort = i + 1
			h.db.Save(&rows[i])
		}
		_ = refreshRepoStat(h.db, rid)
	}
	response.Rest(c, true)
}

type quRepoRow struct {
	ID     string `gorm:"column:id;primaryKey"`
	QuID   string `gorm:"column:qu_id"`
	RepoID string `gorm:"column:repo_id"`
	QuType int    `gorm:"column:qu_type"`
	Sort   int    `gorm:"column:sort"`
}

func (quRepoRow) TableName() string { return "el_qu_repo" }
