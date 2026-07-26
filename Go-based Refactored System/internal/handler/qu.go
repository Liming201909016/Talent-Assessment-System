package handler

import (
	"fmt"
	"sort"
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
		RepoID  string      `json:"repoId"`
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

type quWithAnswers struct {
	model.Qu
	AnswerList []model.QuAnswer `json:"answerList"`
	Sort       int              `json:"sort"`
	RepoName   string           `json:"repoName"`
	RepoID     string           `json:"repoId"`
}

type questionPageRepo struct {
	QuID   string `gorm:"column:qu_id"`
	RepoID string `gorm:"column:repo_id"`
	Title  string `gorm:"column:title"`
}

func loadQuestionPageRelations(db *gorm.DB, questionIDs []string) (map[string][]model.QuAnswer, map[string][]questionPageRepo, error) {
	answersByQuestion := make(map[string][]model.QuAnswer, len(questionIDs))
	repositoriesByQuestion := make(map[string][]questionPageRepo, len(questionIDs))
	for _, questionID := range questionIDs {
		answersByQuestion[questionID] = make([]model.QuAnswer, 0)
		repositoriesByQuestion[questionID] = make([]questionPageRepo, 0)
	}
	if len(questionIDs) == 0 {
		return answersByQuestion, repositoriesByQuestion, nil
	}

	answers := make([]model.QuAnswer, 0)
	if err := db.Where("qu_id IN ?", questionIDs).
		Order("qu_id ASC, id ASC").Find(&answers).Error; err != nil {
		return nil, nil, err
	}
	for _, answer := range answers {
		answersByQuestion[answer.QuID] = append(answersByQuestion[answer.QuID], answer)
	}

	repositories := make([]questionPageRepo, 0)
	if err := db.Table("el_qu_repo AS qr").
		Joins("LEFT JOIN el_repo AS rp ON rp.id = qr.repo_id").
		Where("qr.qu_id IN ?", questionIDs).
		Select("qr.qu_id, qr.repo_id, COALESCE(rp.title, '') AS title").
		Order("qr.qu_id ASC, qr.sort ASC, qr.id ASC").
		Scan(&repositories).Error; err != nil {
		return nil, nil, err
	}
	for _, repository := range repositories {
		repositoriesByQuestion[repository.QuID] = append(repositoriesByQuestion[repository.QuID], repository)
	}
	return answersByQuestion, repositoriesByQuestion, nil
}

// POST /exam/api/qu/qu/paging
func (h *QuHandler) Paging(c *gin.Context) {
	var req quPagingReq
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	req.Size = capPageSize(req.Size)

	// 兼容 DataTable 的 params 嵌套和直接传参
	content := req.Content
	if content == "" {
		content = req.Params.Content
	}
	repoID := req.RepoID
	if repoID == "" {
		repoID = req.Params.RepoID
	}
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
			Where("r.repo_id = ? AND q.dimension_id IS NULL", repoID)
	} else {
		q = h.db.Table("el_qu AS q").Where("q.dimension_id IS NULL")
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
	if err := q.Count(&total).Error; err != nil {
		response.RestErr(c, "查询题目总数失败")
		return
	}
	rows := make([]model.Qu, 0)
	if err := q.Select("q.*").Order("CAST(REPLACE(q.content, 'V', '') AS UNSIGNED), q.update_time desc").
		Offset((req.Current - 1) * req.Size).Limit(req.Size).Scan(&rows).Error; err != nil {
		response.RestErr(c, "查询题目列表失败")
		return
	}

	// 对齐 Java QuServiceImpl.paging：每条记录附带 answerList + repoName
	questionIDs := make([]string, 0, len(rows))
	for _, question := range rows {
		questionIDs = append(questionIDs, question.ID)
	}
	answersByQuestion, repositoriesByQuestion, err := loadQuestionPageRelations(h.db, questionIDs)
	if err != nil {
		response.RestErr(c, "查询题目关联数据失败")
		return
	}
	result := make([]quWithAnswers, len(rows))
	for i, qu := range rows {
		result[i].Qu = qu
		result[i].AnswerList = answersByQuestion[qu.ID]
		result[i].Sort = i + 1
		repo := questionPageRepo{}
		if repositories := repositoriesByQuestion[qu.ID]; len(repositories) > 0 {
			repo = repositories[0]
		}
		if repo.RepoID != "" && repo.Title == "" {
			result[i].RepoName = "[已删题库:" + repo.RepoID + "]"
		} else {
			result[i].RepoName = repo.Title
		}
		result[i].RepoID = repo.RepoID
	}

	response.Rest(c, gin.H{"records": result, "total": total, "current": req.Current, "size": req.Size})
}

// POST /exam/api/qu/qu/list  对齐 Java QuController.list：返回全部题目（参数被忽略）
func (h *QuHandler) List(c *gin.Context) {
	rows := make([]model.Qu, 0)
	if err := h.db.Where("dimension_id IS NULL").Order("update_time desc").Find(&rows).Error; err != nil {
		response.RestErr(c, "查询题目列表失败")
		return
	}
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
	if err := h.db.Where("id = ? AND dimension_id IS NULL", id).First(&qu).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	answers := make([]model.QuAnswer, 0)
	if err := h.db.Where("qu_id = ?", id).Find(&answers).Error; err != nil {
		response.RestErr(c, "查询题目答案失败")
		return
	}
	// 关联题库
	repoIDs := make([]string, 0)
	if err := h.db.Model(&model.QuRepo{}).Where("qu_id = ?", id).
		Pluck("repo_id", &repoIDs).Error; err != nil {
		response.RestErr(c, "查询题目关联失败")
		return
	}
	// 取首个题库的 code（用于前端按题库类型显示不同字段）
	var repoCode string
	if len(repoIDs) > 0 {
		if err := h.db.Model(&model.Repo{}).Where("id = ?", repoIDs[0]).
			Pluck("code", &repoCode).Error; err != nil {
			response.RestErr(c, "查询题库编码失败")
			return
		}
	}
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
		"repoCode":   repoCode,
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
	body.RepoIDs = normalizeBatchIDs(body.RepoIDs)
	if hasCompetencyQuestionMetadata(body.Qu) {
		response.RestErr(c, errCompetencyQuestionDedicatedAPI.Error())
		return
	}
	if body.ID != "" {
		if err := rejectCompetencyQuestionIDs(h.db, []string{body.ID}); err != nil {
			response.RestErr(c, err.Error())
			return
		}
		if err := rejectPaperReferencedQuestionIDs(h.db, []string{body.ID}); err != nil {
			if err == errPaperReferencedQuestion {
				response.RestErr(c, "题目已被试卷引用，不能修改")
			} else {
				response.RestErr(c, "检查题目试卷引用失败")
			}
			return
		}
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
		if err := validateImportRepositoryIDs(tx, body.RepoIDs); err != nil {
			return err
		}
		qu := body.Qu
		isNew := qu.ID == ""
		if isNew {
			qu.ID = strconv.FormatInt(nextID(), 10)
			qu.CreateTime = &now
		} else {
			// 保留原有 create_time
			var orig model.Qu
			if err := tx.Select("create_time").Where("id = ?", qu.ID).Take(&orig).Error; err != nil {
				return err
			}
			if orig.CreateTime != nil {
				qu.CreateTime = orig.CreateTime
			} else {
				qu.CreateTime = &now
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
			if err := tx.Model(&model.QuRepo{}).Where("qu_id = ?", qu.ID).
				Pluck("repo_id", &oldRepoIDs).Error; err != nil {
				return err
			}
			for _, rid := range oldRepoIDs {
				affectedRepos[rid] = true
			}
			// 删除旧的答案和题库关联
			if err := tx.Where("qu_id = ?", qu.ID).Delete(&model.QuAnswer{}).Error; err != nil {
				return err
			}
			if err := tx.Where("qu_id = ?", qu.ID).Delete(&model.QuRepo{}).Error; err != nil {
				return err
			}
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
	if err := rejectCompetencyQuestionIDs(h.db, b.IDs); err != nil {
		response.RestErr(c, err.Error())
		return
	}
	var paperReferenceCount int64
	if err := h.db.Table("el_paper_qu").Where("qu_id IN ?", b.IDs).
		Count(&paperReferenceCount).Error; err != nil {
		response.RestErr(c, "检查题目试卷引用失败")
		return
	}
	if paperReferenceCount > 0 {
		response.RestErr(c, fmt.Sprintf("无法删除：题目已被试卷引用（%d条）", paperReferenceCount))
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// 先收集受影响 repo
		var oldRepoIDs []string
		if err := tx.Model(&model.QuRepo{}).Where("qu_id IN ?", b.IDs).
			Distinct("repo_id").Pluck("repo_id", &oldRepoIDs).Error; err != nil {
			return err
		}
		if err := tx.Where("qu_id IN ?", b.IDs).Delete(&model.QuAnswer{}).Error; err != nil {
			return err
		}
		if err := tx.Where("qu_id IN ?", b.IDs).Delete(&model.QuRepo{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id IN ?", b.IDs).Delete(&model.Qu{}).Error; err != nil {
			return err
		}
		sort.Strings(oldRepoIDs)
		for _, repoID := range oldRepoIDs {
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
