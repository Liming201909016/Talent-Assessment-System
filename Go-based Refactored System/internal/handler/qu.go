package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// QuHandler 题目（el_qu）管理 —— 对应 /exam/api/qu/qu/*
type QuHandler struct{ db *gorm.DB }

func NewQuHandler(db *gorm.DB) *QuHandler { return &QuHandler{db: db} }

// 前端 DataTable 请求：{current, size, params: {content, repoIds, quType}}
// 注意：前端 quType 可能传 "" 空字符串或 int，用 interface{} 接收避免 JSON 解析失败
type quPagingReq struct {
	Current int         `json:"current"`
	Size    int         `json:"size"`
	Title   string      `json:"title"`
	Content string      `json:"content"`
	RepoID  string      `json:"repoId"`
	QuType  interface{} `json:"quType"`
	Level   interface{} `json:"level"`
	Params  struct {
		Content string      `json:"content"`
		RepoIds []string    `json:"repoIds"`
		QuType  interface{} `json:"quType"`
	} `json:"params"`
}

func toIntPtr(v interface{}) *int {
	switch val := v.(type) {
	case float64:
		i := int(val)
		return &i
	case int:
		return &val
	case string:
		if val == "" {
			return nil
		}
		i, err := strconv.Atoi(val)
		if err != nil {
			return nil
		}
		return &i
	}
	return nil
}

// POST /exam/api/qu/qu/paging
func (h *QuHandler) Paging(c *gin.Context) {
	var req quPagingReq
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}

	// 兼容 DataTable 的 params 嵌套和直接传参
	content := req.Content
	if content == "" {
		content = req.Params.Content
	}
	repoID := req.RepoID
	if repoID == "" && len(req.Params.RepoIds) > 0 {
		repoID = req.Params.RepoIds[0]
	}
	quType := toIntPtr(req.QuType)
	if quType == nil {
		quType = toIntPtr(req.Params.QuType)
	}
	level := toIntPtr(req.Level)

	// 如指定 repoId，通过 el_qu_repo 关联过滤
	var q *gorm.DB
	if repoID != "" {
		q = h.db.Table("el_qu AS q").
			Joins("INNER JOIN el_qu_repo AS r ON r.qu_id = q.id").
			Where("r.repo_id = ?", repoID)
	} else {
		q = h.db.Table("el_qu AS q")
	}
	if content != "" {
		q = q.Where("q.content like ?", "%"+content+"%")
	}
	if req.Title != "" {
		q = q.Where("q.content like ?", "%"+req.Title+"%")
	}
	if quType != nil {
		q = q.Where("q.qu_type = ?", *quType)
	}
	if level != nil {
		q = q.Where("q.level = ?", *level)
	}
	var total int64
	q.Count(&total)
	var rows []model.Qu
	q.Select("q.*").Order("CAST(REPLACE(q.content, 'V', '') AS UNSIGNED), q.update_time desc").
		Offset((req.Current - 1) * req.Size).Limit(req.Size).Scan(&rows)

	// 对齐 Java QuServiceImpl.paging：每条记录附带 answerList
	type quWithAnswers struct {
		model.Qu
		AnswerList []model.QuAnswer `json:"answerList"`
		Sort       int              `json:"sort"`
	}
	result := make([]quWithAnswers, len(rows))
	for i, qu := range rows {
		result[i].Qu = qu
		h.db.Where("qu_id = ?", qu.ID).Order("id").Find(&result[i].AnswerList)
		result[i].Sort = i + 1
	}

	response.Rest(c, gin.H{"records": result, "total": total, "current": req.Current, "size": req.Size})
}

// POST /exam/api/qu/qu/list  对齐 Java QuController.list：返回全部题目（参数被忽略）
func (h *QuHandler) List(c *gin.Context) {
	var rows []model.Qu
	h.db.Order("update_time desc").Find(&rows)
	response.Rest(c, rows)
}

// POST /exam/api/qu/qu/detail {id}
func (h *QuHandler) Detail(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	var qu model.Qu
	if err := h.db.Where("id = ?", id).First(&qu).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	var answers []model.QuAnswer
	h.db.Where("qu_id = ?", id).Find(&answers)
	// 关联题库
	var repoIDs []string
	h.db.Model(&model.QuRepo{}).Where("qu_id = ?", id).Pluck("repo_id", &repoIDs)
	// 对齐 Java：返回扁平对象（前端 form.vue 直接读 data.content / data.answerList）
	response.Rest(c, gin.H{
		"id":         qu.ID,
		"quType":     qu.QuType,
		"level":      qu.Level,
		"image":      qu.Image,
		"content":    qu.Content,
		"title":      qu.Title,
		"analysis":   qu.Analysis,
		"remark":     qu.Remark,
		"createTime": qu.CreateTime,
		"updateTime": qu.UpdateTime,
		"answerList": answers,
		"repoIds":    repoIDs,
	})
}

// Java QuType.RADIO = 1
const quTypeRadio = 1

// POST /exam/api/qu/qu/save
// 请求包含 qu 主体 + answers + repoIds（参考前端 saveData）
// 对齐 Java QuServiceImpl.save + checkData：
//   - content 非空；repoIds 至少 1 个
//   - 客观题 answers 非空；每项必须标注 is_right；至少 1 个正确项
//   - 单选题正确项不能 >1
//   - 保存后对每个受影响 repo 调用 refreshStat（刷新 radio/multi/judge_count）
func (h *QuHandler) Save(c *gin.Context) {
	// 前端 answerList 的 isRight 是 boolean（true/false），需要兼容转换
	type answerInput struct {
		ID       string `json:"id"`
		QuID     string `json:"quId"`
		IsRight  any    `json:"isRight"` // 接收 boolean 或 int
		Image    string `json:"image"`
		Content  string `json:"content"`
		Analysis string `json:"analysis"`
		Score    int    `json:"score"`
	}
	var body struct {
		model.Qu
		Answers []answerInput `json:"answerList"`
		RepoIDs []string      `json:"repoIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.RestErr(c, "参数错误: "+err.Error())
		return
	}

	// 转换 answerInput → model.QuAnswer（boolean → int8）
	answers := make([]model.QuAnswer, len(body.Answers))
	for i, a := range body.Answers {
		ir := int8(0)
		switch v := a.IsRight.(type) {
		case bool:
			if v {
				ir = 1
			}
		case float64:
			ir = int8(v)
		case int:
			ir = int8(v)
		}
		answers[i] = model.QuAnswer{
			ID: a.ID, QuID: a.QuID, IsRight: ir, Image: a.Image,
			Content: a.Content, Analysis: a.Analysis, Score: a.Score,
		}
	}

	// 1. 校验
	if strings.TrimSpace(body.Content) == "" {
		response.RestErr(c, "题目内容不能为空！")
		return
	}
	if len(body.RepoIDs) == 0 {
		response.RestErr(c, "至少要选择一个题库！")
		return
	}
	if len(answers) == 0 {
		response.RestErr(c, "客观题至少要包含一个备选答案！")
		return
	}
	trueCount := 0
	for _, a := range answers {
		if a.IsRight != 0 && a.IsRight != 1 {
			response.RestErr(c, "必须定义选项是否正确项！")
			return
		}
		if a.IsRight == 1 {
			trueCount++
		}
	}
	if trueCount == 0 {
		response.RestErr(c, "至少要包含一个正确项！")
		return
	}
	if body.QuType == quTypeRadio && trueCount > 1 {
		response.RestErr(c, "单选题不能包含多个正确项！")
		return
	}

	now := time.Now()
	// 收集需要刷新统计的 repo：新关联 + 旧关联（旧的要算在内，数量可能变少）
	affectedRepos := map[string]bool{}
	for _, rid := range body.RepoIDs {
		if rid != "" {
			affectedRepos[rid] = true
		}
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		qu := body.Qu
		isNew := qu.ID == ""
		if isNew {
			qu.ID = strconv.FormatInt(nextID(), 10)
			qu.CreateTime = &now
		} else {
			// 保留原有 create_time
			var orig model.Qu
			if err := tx.Select("create_time").Where("id = ?", qu.ID).Take(&orig).Error; err == nil && orig.CreateTime != nil {
				qu.CreateTime = orig.CreateTime
			} else {
				qu.CreateTime = &now // 原值为空时用当前时间
			}
		}
		qu.UpdateTime = &now
		if isNew {
			if err := tx.Create(&qu).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(&qu).Error; err != nil {
				return err
			}
			// 收集旧关联的 repo
			var oldRepoIDs []string
			tx.Model(&model.QuRepo{}).Where("qu_id = ?", qu.ID).Pluck("repo_id", &oldRepoIDs)
			for _, rid := range oldRepoIDs {
				affectedRepos[rid] = true
			}
			// 删除旧的答案和题库关联
			tx.Where("qu_id = ?", qu.ID).Delete(&model.QuAnswer{})
			tx.Where("qu_id = ?", qu.ID).Delete(&model.QuRepo{})
		}
		// 插入 answers
		for i := range answers {
			a := answers[i]
			if a.ID == "" {
				a.ID = strconv.FormatInt(nextID()+int64(i), 10)
			}
			a.QuID = qu.ID
			if err := tx.Create(&a).Error; err != nil {
				return err
			}
		}
		// 插入题库关联
		for i, rid := range body.RepoIDs {
			if rid == "" {
				continue
			}
			qr := model.QuRepo{
				ID:     strconv.FormatInt(nextID()+int64(100+i), 10),
				QuID:   qu.ID,
				RepoID: rid,
				QuType: qu.QuType,
				Sort:   i + 1,
			}
			if err := tx.Create(&qr).Error; err != nil {
				return err
			}
		}
		// 刷新每个受影响 repo 的统计
		for rid := range affectedRepos {
			if err := refreshRepoStat(tx, rid); err != nil {
				return err
			}
		}
		body.Qu = qu
		return nil
	})
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, body.Qu)
}

// POST /exam/api/qu/qu/delete {ids:[...]}
func (h *QuHandler) Delete(c *gin.Context) {
	var b struct {
		IDs []string `json:"ids"`
	}
	_ = c.ShouldBindJSON(&b)
	if len(b.IDs) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	affectedRepos := map[string]bool{}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// 先收集受影响 repo
		var oldRepoIDs []string
		tx.Model(&model.QuRepo{}).Where("qu_id IN ?", b.IDs).
			Distinct("repo_id").Pluck("repo_id", &oldRepoIDs)
		for _, rid := range oldRepoIDs {
			affectedRepos[rid] = true
		}
		if err := tx.Where("id IN ?", b.IDs).Delete(&model.Qu{}).Error; err != nil {
			return err
		}
		tx.Where("qu_id IN ?", b.IDs).Delete(&model.QuAnswer{})
		tx.Where("qu_id IN ?", b.IDs).Delete(&model.QuRepo{})
		for rid := range affectedRepos {
			if err := refreshRepoStat(tx, rid); err != nil {
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

// refreshRepoStat 对齐 Java RepoMapper.refreshStat：
//
//	UPDATE el_repo a SET
//	  radio_count=(SELECT COUNT(0) FROM el_qu_repo WHERE repo_id=a.id AND qu_type=1),
//	  multi_count=(SELECT COUNT(0) FROM el_qu_repo WHERE repo_id=a.id AND qu_type=2),
//	  judge_count=(SELECT COUNT(0) FROM el_qu_repo WHERE repo_id=a.id AND qu_type=3)
//	WHERE a.id=?
func refreshRepoStat(tx *gorm.DB, repoID string) error {
	if repoID == "" {
		return nil
	}
	sql := `UPDATE el_repo a SET
		radio_count=(SELECT COUNT(0) FROM el_qu_repo WHERE repo_id=a.id AND qu_type=1),
		multi_count=(SELECT COUNT(0) FROM el_qu_repo WHERE repo_id=a.id AND qu_type=2),
		judge_count=(SELECT COUNT(0) FROM el_qu_repo WHERE repo_id=a.id AND qu_type=3)
		WHERE a.id=?`
	return tx.Exec(sql, repoID).Error
}

// 公共工具：从 body 或 query 提取 id
func bindID(c *gin.Context) string {
	if id := c.Query("id"); id != "" {
		return id
	}
	var b struct {
		ID string `json:"id"`
	}
	_ = c.ShouldBindJSON(&b)
	return b.ID
}
