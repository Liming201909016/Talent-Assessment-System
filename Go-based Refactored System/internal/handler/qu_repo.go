package handler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
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
	Virtual    bool       `gorm:"-"                    json:"virtual"`
}

func (quRepo) TableName() string { return "el_repo" }

const competencyVirtualRepoID = "competency-question-bank-00401"

func competencyVirtualRepo(questionCount int) quRepo {
	return quRepo{ID: competencyVirtualRepoID, Code: "00401", Title: "胜任力测验题库", RadioCount: questionCount, Virtual: true}
}

func countCompetencyQuestions(db *gorm.DB) (int, error) {
	var count int64
	if err := db.Table("el_qu").Where("dimension_id IS NOT NULL").Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

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
	req.Size = capPageSize(req.Size)
	// 兼容两种传参：顶层 title 或 params.title
	searchTitle := req.Title
	if searchTitle == "" {
		searchTitle = req.Params.Title
	}
	q := h.db.Model(&quRepo{})
	if searchTitle != "" {
		q = q.Where("title like ?", "%"+searchTitle+"%")
	}
	var physicalTotal int64
	if err := q.Count(&physicalTotal).Error; err != nil {
		response.RestErr(c, "查询题库总数失败")
		return
	}
	includeCompetency := searchTitle == "" || strings.Contains("00401", searchTitle) || strings.Contains("胜任力测验题库", searchTitle)
	competencyCount := 0
	if includeCompetency {
		var err error
		competencyCount, err = countCompetencyQuestions(h.db)
		if err != nil {
			response.RestErr(c, "查询胜任力题数失败")
			return
		}
	}
	total := physicalTotal
	if includeCompetency {
		total++
	}
	rows := make([]quRepo, 0, req.Size)
	pageStart := (req.Current - 1) * req.Size
	if includeCompetency && pageStart == 0 {
		rows = append(rows, competencyVirtualRepo(competencyCount))
	}
	physicalOffset := pageStart
	if includeCompetency {
		physicalOffset--
		if physicalOffset < 0 {
			physicalOffset = 0
		}
	}
	remaining := req.Size - len(rows)
	if remaining > 0 && int64(physicalOffset) < physicalTotal {
		physicalRows := make([]quRepo, 0, remaining)
		if err := q.Order("update_time desc").Offset(physicalOffset).
			Limit(remaining).Find(&physicalRows).Error; err != nil {
			response.RestErr(c, "查询题库列表失败")
			return
		}
		rows = append(rows, physicalRows...)
	}
	response.Rest(c, gin.H{"records": rows, "total": total, "current": req.Current, "size": req.Size})
}

// POST /api/qu/repo/list
func (h *RepoHandler) List(c *gin.Context) {
	rows := make([]quRepo, 0)
	if err := h.db.Order("update_time desc").Find(&rows).Error; err != nil {
		response.RestErr(c, "查询题库列表失败")
		return
	}
	response.Rest(c, rows)
}

// POST /api/qu/repo/detail?id=xxx  or JSON body {"id":"xxx"}
func (h *RepoHandler) Detail(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		var b struct {
			ID string `json:"id"`
		}
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
	r.Title = strings.TrimSpace(r.Title)
	if r.Title == "" {
		response.RestErr(c, "题库名称不能为空")
		return
	}
	now := time.Now()
	if r.ID == "" {
		r.ID = strconv.FormatInt(nextID(), 10)
		r.RadioCount = 0
		r.MultiCount = 0
		r.JudgeCount = 0
		r.CreateTime = &now
		r.UpdateTime = &now
		if err := h.db.Create(&r).Error; err != nil {
			response.RestErr(c, err.Error())
			return
		}
	} else {
		result := h.db.Model(&quRepo{}).Where("id = ?", r.ID).Updates(map[string]any{
			"code": r.Code, "title": r.Title, "remark": r.Remark, "update_time": &now,
		})
		if result.Error != nil {
			response.RestErr(c, result.Error.Error())
			return
		}
		if result.RowsAffected == 0 {
			response.RestErr(c, "题库不存在")
			return
		}
		if err := h.db.Where("id = ?", r.ID).Take(&r).Error; err != nil {
			response.RestErr(c, "查询题库失败")
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
	for _, id := range b.IDs {
		if id == competencyVirtualRepoID {
			response.RestErr(c, "00401为胜任力题库入口，不能通过传统题库接口删除")
			return
		}
	}

	// FB-024: 删除题库前检查是否被 exam 引用
	var examRefCount int64
	h.db.Table("el_exam_repo").Where("repo_id IN ?", b.IDs).Count(&examRefCount)
	if examRefCount > 0 {
		response.RestErr(c, fmt.Sprintf("无法删除：被 %d 个考试引用，请先移除考试中的题库关联", examRefCount))
		return
	}

	// FB-025: 删除题库时同时清理题目-题库关联（el_qu_repo），避免孤儿数据
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&quRepo{}).Where("id IN ?", b.IDs).Delete(&quRepo{}).Error; err != nil {
			return err
		}
		// 清理关联表
		return tx.Table("el_qu_repo").Where("repo_id IN ?", b.IDs).Delete(nil).Error
	}); err != nil {
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

func normalizeBatchIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
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
	b.QuIDs = normalizeBatchIDs(b.QuIDs)
	b.RepoIDs = normalizeBatchIDs(b.RepoIDs)
	if len(b.QuIDs) == 0 || len(b.RepoIDs) == 0 {
		response.RestErr(c, "quIds 或 repoIds 为空")
		return
	}
	if err := rejectCompetencyQuestionIDs(h.db, b.QuIDs); err != nil {
		response.RestErr(c, err.Error())
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var questions []struct {
			ID     string `gorm:"column:id"`
			QuType int    `gorm:"column:qu_type"`
		}
		if err := tx.Table("el_qu").Select("id, qu_type").
			Where("id IN ? AND dimension_id IS NULL", b.QuIDs).
			Find(&questions).Error; err != nil {
			return err
		}
		if len(questions) != len(b.QuIDs) {
			return fmt.Errorf("部分题目不存在或不是传统题目")
		}
		var repoCount int64
		if err := tx.Model(&quRepo{}).Where("id IN ?", b.RepoIDs).Count(&repoCount).Error; err != nil {
			return err
		}
		if int(repoCount) != len(b.RepoIDs) {
			return fmt.Errorf("部分题库不存在")
		}

		var oldRepoIDs []string
		if err := tx.Model(&quRepoRow{}).Where("qu_id IN ?", b.QuIDs).
			Distinct("repo_id").Pluck("repo_id", &oldRepoIDs).Error; err != nil {
			return err
		}
		affected := make(map[string]struct{}, len(oldRepoIDs)+len(b.RepoIDs))
		for _, repoID := range oldRepoIDs {
			affected[repoID] = struct{}{}
		}
		for _, repoID := range b.RepoIDs {
			affected[repoID] = struct{}{}
		}

		if b.Remove != nil && *b.Remove {
			if err := tx.Where("repo_id IN ? AND qu_id IN ?", b.RepoIDs, b.QuIDs).
				Delete(&quRepoRow{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Where("qu_id IN ?", b.QuIDs).Delete(&quRepoRow{}).Error; err != nil {
				return err
			}
			associations := make([]quRepoRow, 0, len(questions)*len(b.RepoIDs))
			for _, question := range questions {
				for _, repoID := range b.RepoIDs {
					associations = append(associations, quRepoRow{
						ID: strconv.FormatInt(nextID(), 10), QuID: question.ID,
						RepoID: repoID, QuType: question.QuType,
					})
				}
			}
			if err := tx.CreateInBatches(associations, 100).Error; err != nil {
				return err
			}
		}

		affectedRepoIDs := make([]string, 0, len(affected))
		for repoID := range affected {
			affectedRepoIDs = append(affectedRepoIDs, repoID)
		}
		sort.Strings(affectedRepoIDs)
		for _, repoID := range affectedRepoIDs {
			rows := make([]quRepoRow, 0)
			if err := tx.Where("repo_id = ?", repoID).
				Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
				return err
			}
			for i := range rows {
				if err := tx.Model(&quRepoRow{}).Where("id = ?", rows[i].ID).
					Update("sort", i+1).Error; err != nil {
					return err
				}
			}
			if err := refreshRepoStat(tx, repoID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.RestErr(c, err.Error())
		return
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
