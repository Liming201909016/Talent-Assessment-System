package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// PaperHandler 对齐 Java PaperController：/exam/api/paper/paper/*
type PaperHandler struct{ db *gorm.DB }

func NewPaperHandler(db *gorm.DB) *PaperHandler { return &PaperHandler{db: db} }

// Java 常量
const (
	paperStateING      = 0
	paperStateWaitOpt  = 1
	paperStateFinished = 2
	paperStateBreak    = 3

	examStateEnable     = 0
	examStateDisabled   = 1
	examStateReadyStart = 2
	examStateOverdue    = 3

	quTypeMulti = 2
	quTypeJudge = 3
)

var abcList = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L",
	"M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

// ===================== 分页 / CRUD =====================

// POST /exam/api/paper/paper/paging
func (h *PaperHandler) Paging(c *gin.Context) {
	var req struct {
		Current int `json:"current"`
		Size    int `json:"size"`
		Params  struct {
			Title  string      `json:"title"`
			ExamID string      `json:"examId"`
			UserID string      `json:"userId"`
			State  interface{} `json:"state"`
		} `json:"params"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	q := h.db.Model(&model.Paper{})
	if req.Params.Title != "" {
		q = q.Where("title like ?", "%"+req.Params.Title+"%")
	}
	if req.Params.ExamID != "" {
		q = q.Where("exam_id = ?", req.Params.ExamID)
	}
	if req.Params.UserID != "" {
		q = q.Where("user_id = ?", req.Params.UserID)
	}
	if st := toIntPtr(req.Params.State); st != nil {
		q = q.Where("state = ?", *st)
	}
	var total int64
	q.Count(&total)
	var rows []model.Paper
	q.Order("update_time desc").
		Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	response.Rest(c, gin.H{
		"records": rows, "total": total,
		"current": req.Current, "size": req.Size,
	})
}

// POST /exam/api/paper/paper/detail   {id}
func (h *PaperHandler) Detail(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	var p model.Paper
	if err := h.db.Where("id = ?", id).First(&p).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	response.Rest(c, p)
}

// POST /exam/api/paper/paper/save
func (h *PaperHandler) Save(c *gin.Context) {
	var p model.Paper
	if err := c.ShouldBindJSON(&p); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	now := time.Now()
	p.UpdateTime = &now
	if p.ID == "" {
		p.ID = strconv.FormatInt(nextID(), 10)
		p.CreateTime = &now
		if err := h.db.Create(&p).Error; err != nil {
			response.RestErr(c, err.Error())
			return
		}
	} else {
		// 保留原有 create_time
		var orig model.Paper
		if err := h.db.Select("create_time").Where("id = ?", p.ID).Take(&orig).Error; err == nil && orig.CreateTime != nil {
			p.CreateTime = orig.CreateTime
		} else {
			p.CreateTime = &now // 原值为空时用当前时间
		}
		if err := h.db.Save(&p).Error; err != nil {
			response.RestErr(c, err.Error())
			return
		}
	}
	response.Rest(c, gin.H{"id": p.ID})
}

// POST /exam/api/paper/paper/delete {ids:[]}
func (h *PaperHandler) Delete(c *gin.Context) {
	var b struct {
		IDs []string `json:"ids"`
	}
	_ = c.ShouldBindJSON(&b)
	if len(b.IDs) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	if err := h.db.Where("id IN ?", b.IDs).Delete(&model.Paper{}).Error; err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, true)
}

// ===================== 创建试卷（核心） =====================

// POST /exam/api/paper/paper/create-paper  {examId}
// 对齐 PaperServiceImpl.createPaper(examId)
func (h *PaperHandler) CreatePaper(c *gin.Context) {
	var b struct {
		ExamID string `json:"examId"`
	}
	if err := c.ShouldBindJSON(&b); err != nil || b.ExamID == "" {
		response.RestErr(c, "examId 为空")
		return
	}
	var exam model.Exam
	if err := h.db.Where("id = ?", b.ExamID).First(&exam).Error; err != nil {
		response.RestErr(c, "考试不存在！")
		return
	}
	switch exam.State {
	case examStateEnable:
		// ok
	case examStateOverdue:
		response.RestErr(c, "考试已结束！")
		return
	case examStateReadyStart:
		response.RestErr(c, "考试尚未开始！")
		return
	default:
		response.RestErr(c, "未到开放时间！")
		return
	}

	paperID, err := h.createPaperTx(&exam)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, gin.H{"id": paperID})
}

// createPaperTx：事务内按 exam_repo 规则随机抽题 + 落 paper / paper_qu / paper_qu_answer
func (h *PaperHandler) createPaperTx(exam *model.Exam) (string, error) {
	var paperID string
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// 1. exam_repo 规则
		var rules []model.ExamRepo
		if err := tx.Where("exam_id = ?", exam.ID).Find(&rules).Error; err != nil {
			return err
		}
		type chosenQu struct {
			qu    model.Qu
			rule  model.ExamRepo
			score int // 按类型取 radio/multi/judge score
		}
		var chosen []chosenQu
		excludes := []string{"none"}

		pick := func(rule model.ExamRepo, quType, count, score int) error {
			if count <= 0 {
				return nil
			}
			q := tx.Table("el_qu AS a").
				Joins("LEFT JOIN el_qu_repo b ON a.id = b.qu_id").
				Where("b.repo_id = ? AND a.qu_type = ?", rule.RepoID, quType).
				Where("a.id NOT IN ?", excludes)
			if exam.Level > 0 {
				q = q.Where("a.`level` = ?", exam.Level)
			}
			var list []model.Qu
			if err := q.Select("a.*").Order("RAND()").Limit(count).Find(&list).Error; err != nil {
				return err
			}
			for _, qu := range list {
				chosen = append(chosen, chosenQu{qu: qu, rule: rule, score: score})
				excludes = append(excludes, qu.ID)
			}
			return nil
		}

		for _, rule := range rules {
			if err := pick(rule, quTypeRadio, rule.RadioCount, rule.RadioScore); err != nil {
				return err
			}
			if err := pick(rule, quTypeMulti, rule.MultiCount, rule.MultiScore); err != nil {
				return err
			}
			if err := pick(rule, quTypeJudge, rule.JudgeCount, rule.JudgeScore); err != nil {
				return err
			}
		}
		if len(chosen) == 0 {
			return errors.New("规则不正确，无对应的考题！")
		}

		// 2. 落 paper
		now := time.Now()
		limit := now.Add(time.Duration(exam.TotalTime) * time.Minute)
		paper := model.Paper{
			ID:           strconv.FormatInt(nextID(), 10),
			UserID:       "101", // Java 原系统也硬编码管理员ID
			DepartID:     "105", // Java 原系统也硬编码部门ID
			ExamID:       exam.ID,
			Title:        exam.Title,
			TotalScore:   exam.TotalScore,
			TotalTime:    exam.TotalTime,
			UserScore:    0,
			QualifyScore: exam.QualifyScore,
			State:        paperStateING,
			HasSaq:       0,
			CreateTime:   &now,
			UpdateTime:   &now,
			LimitTime:    &limit,
		}
		if err := tx.Create(&paper).Error; err != nil {
			return err
		}
		paperID = paper.ID

		// 3. 批量落 paper_qu + paper_qu_answer
		var allPQ []model.PaperQu
		var allPQA []model.PaperQuAnswer
		baseID := nextID()
		for i, cq := range chosen {
			pq := model.PaperQu{
				ID:          strconv.FormatInt(baseID+int64(i), 10),
				PaperID:     paperID,
				QuID:        cq.qu.ID,
				QuType:      cq.qu.QuType,
				Answered:    0,
				IsRight:     0,
				Sort:        i,
				Score:       cq.score,
				ActualScore: cq.score,
			}
			allPQ = append(allPQ, pq)
			// 答案列表（按 Java：2 项时 is_right 反向放序，>2 项按 score 降序）
			var answers []model.QuAnswer
			if err := tx.Where("qu_id = ?", cq.qu.ID).Find(&answers).Error; err != nil {
				return err
			}
			if len(answers) == 2 {
				for j, a := range answers {
					ii := 1
					if a.IsRight == 1 {
						ii = 0
					}
					_ = j
					pqa := model.PaperQuAnswer{
						ID:       strconv.FormatInt(baseID+int64(10000+i*100+j), 10),
						PaperID:  paperID,
						QuID:     a.QuID,
						AnswerID: a.ID,
						Checked:  0,
						Sort:     ii,
						Abc:      abcList[ii],
						IsRight:  a.IsRight,
						Score:    a.Score,
					}
					allPQA = append(allPQA, pqa)
				}
			} else {
				sortedAnswers := make([]model.QuAnswer, len(answers))
				copy(sortedAnswers, answers)
				for a := 1; a < len(sortedAnswers); a++ {
					for b := a; b > 0 && sortedAnswers[b-1].Score < sortedAnswers[b].Score; b-- {
						sortedAnswers[b-1], sortedAnswers[b] = sortedAnswers[b], sortedAnswers[b-1]
					}
				}
				for j, a := range sortedAnswers {
					pqa := model.PaperQuAnswer{
						ID:       strconv.FormatInt(baseID+int64(10000+i*100+j), 10),
						PaperID:  paperID,
						QuID:     a.QuID,
						AnswerID: a.ID,
						Checked:  0,
						Sort:     j,
						Abc:      abcList[j],
						IsRight:  a.IsRight,
						Score:    a.Score,
					}
					allPQA = append(allPQA, pqa)
				}
			}
		}
		if err := tx.CreateInBatches(allPQ, 100).Error; err != nil {
			return err
		}
		if len(allPQA) > 0 {
			if err := tx.CreateInBatches(allPQA, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return paperID, err
}

// ===================== 试卷详情 / 单题详情 / 答卷结果 =====================

// POST /exam/api/paper/paper/paper-detail {id=paperId}
// 对齐 paperDetail：返回 paper 基本信息 + repo + radio/multi/judge 分组题目
func (h *PaperHandler) PaperDetail(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	var paper model.Paper
	if err := h.db.Where("id = ?", id).First(&paper).Error; err != nil {
		response.RestErr(c, "试卷不存在")
		return
	}
	var er model.ExamRepo
	var repo model.Repo
	if err := h.db.Where("exam_id = ?", paper.ExamID).Take(&er).Error; err == nil {
		if err2 := h.db.Where("id = ?", er.RepoID).Take(&repo).Error; err2 != nil {
			slog.Warn("paper: repo not found", "repoId", er.RepoID, "error", err2)
		}
	}
	var qs []model.PaperQu
	h.db.Where("paper_id = ?", id).Order("sort ASC").Find(&qs)

	radio, multi, judge := []model.PaperQu{}, []model.PaperQu{}, []model.PaperQu{}
	for _, q := range qs {
		switch q.QuType {
		case quTypeRadio:
			radio = append(radio, q)
		case quTypeMulti:
			multi = append(multi, q)
		case quTypeJudge:
			judge = append(judge, q)
		}
	}
	response.Rest(c, gin.H{
		"id":           paper.ID,
		"title":        paper.Title,
		"examId":       paper.ExamID,
		"totalScore":   paper.TotalScore,
		"totalTime":    paper.TotalTime,
		"qualifyScore": paper.QualifyScore,
		"state":        paper.State,
		"limitTime":    paper.LimitTime,
		"createTime":   paper.CreateTime,
		"repo":         repo,
		"radioList":    radio,
		"multiList":    multi,
		"judgeList":    judge,
	})
}

// paperQuAnswerView 对齐 PaperQuAnswerExtDTO
type paperQuAnswerView struct {
	ID       string `json:"id"`
	PaperID  string `json:"paperId"`
	QuID     string `json:"quId"`
	AnswerID string `json:"answerId"`
	Checked  int8   `json:"checked"`
	Sort     int    `json:"sort"`
	Abc      string `json:"abc"`
	IsRight  int8   `json:"isRight"`
	Content  string `json:"content"`
	Image    string `json:"image"`
	Score    int    `json:"score"`
}

// POST /exam/api/paper/paper/qu-detail {paperId, quId}
// 对齐 findQuDetail：返回单题 + 所有候选答案
func (h *PaperHandler) QuDetail(c *gin.Context) {
	var b struct {
		PaperID string `json:"paperId"`
		QuID    string `json:"quId"`
	}
	if err := c.ShouldBindJSON(&b); err != nil || b.PaperID == "" || b.QuID == "" {
		response.RestErr(c, "参数错误")
		return
	}
	var pq model.PaperQu
	if err := h.db.Where("paper_id = ? AND qu_id = ?", b.PaperID, b.QuID).First(&pq).Error; err != nil {
		response.RestErr(c, "题目不存在")
		return
	}
	var qu model.Qu
	if err := h.db.Where("id = ?", b.QuID).Take(&qu).Error; err != nil {
		response.RestErr(c, "题目不存在")
		return
	}
	var answers []paperQuAnswerView
	h.db.Table("el_paper_qu_answer AS pa").
		Select("pa.id, pa.paper_id, pa.qu_id, pa.answer_id, pa.checked, pa.sort, pa.abc, qa.is_right, qa.content, qa.image, qa.score").
		Joins("LEFT JOIN el_qu_answer qa ON pa.answer_id = qa.id").
		Where("pa.paper_id = ? AND pa.qu_id = ?", b.PaperID, b.QuID).
		Order("pa.sort ASC").
		Find(&answers)

	response.Rest(c, gin.H{
		"id":          pq.ID,
		"paperId":     pq.PaperID,
		"quId":        pq.QuID,
		"quType":      pq.QuType,
		"answered":    pq.Answered,
		"answer":      pq.Answer,
		"sort":        pq.Sort,
		"score":       pq.Score,
		"actualScore": pq.ActualScore,
		"isRight":     pq.IsRight,
		"content":     qu.Content,
		"image":       qu.Image,
		"analysis":    qu.Analysis,
		"answerList":  answers,
	})
}

// POST /exam/api/paper/paper/fill-answer {paperId, quId, answers:[], answer}
// 对齐 PaperServiceImpl.fillAnswer
func (h *PaperHandler) FillAnswer(c *gin.Context) {
	var b struct {
		PaperID string   `json:"paperId"`
		QuID    string   `json:"quId"`
		Answers []string `json:"answers"`
		Answer  string   `json:"answer"`
	}
	if err := c.ShouldBindJSON(&b); err != nil || b.PaperID == "" || b.QuID == "" {
		response.RestErr(c, "参数错误")
		return
	}
	// 未作答 → 直接返回
	if len(b.Answers) == 0 && strings.TrimSpace(b.Answer) == "" {
		response.Rest(c, true)
		return
	}

	// 找出题目所属题库 code（判定 002 管理特质）
	var repoCode string
	h.db.Table("el_qu_repo qr").
		Joins("INNER JOIN el_repo r ON r.id = qr.repo_id").
		Where("qr.qu_id = ?", b.QuID).
		Limit(1).
		Pluck("r.code", &repoCode)

	answerSet := map[string]bool{}
	for _, id := range b.Answers {
		answerSet[id] = true
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var list []model.PaperQuAnswer
		if err := tx.Where("paper_id = ? AND qu_id = ?", b.PaperID, b.QuID).Find(&list).Error; err != nil {
			return err
		}

		if strings.HasPrefix(repoCode, "002") {
			// 管理特质：取选中项的 score 作为 actualScore
			actualScore := 0
			for i := range list {
				it := &list[i]
				if answerSet[it.ID] {
					it.Checked = 1
					actualScore = it.Score
				} else {
					it.Checked = 0
				}
				if err := tx.Save(it).Error; err != nil {
					return err
				}
			}
			return tx.Model(&model.PaperQu{}).
				Where("paper_id = ? AND qu_id = ?", b.PaperID, b.QuID).
				Updates(map[string]any{
					"is_right": 1, "answer": b.Answer,
					"answered": 1, "actual_score": actualScore,
				}).Error
		}

		// 心理特质：所有项 checked 必须与 is_right 一致
		right := true
		for i := range list {
			it := &list[i]
			if answerSet[it.ID] {
				it.Checked = 1
			} else {
				it.Checked = 0
			}
			if it.IsRight != it.Checked {
				right = false
			}
			if err := tx.Save(it).Error; err != nil {
				return err
			}
		}
		rightFlag := int8(0)
		if right {
			rightFlag = 1
		}
		return tx.Model(&model.PaperQu{}).
			Where("paper_id = ? AND qu_id = ?", b.PaperID, b.QuID).
			Updates(map[string]any{
				"is_right": rightFlag, "answer": b.Answer, "answered": 1,
			}).Error
	})
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, true)
}

// POST /exam/api/paper/paper/hand-exam {id=paperId}
// 对齐 PaperServiceImpl.handExam
func (h *PaperHandler) HandExam(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var paper model.Paper
		if err := tx.Where("id = ?", id).First(&paper).Error; err != nil {
			return err
		}
		if paper.State != paperStateING {
			return errors.New("试卷状态不正确！")
		}
		// 客观分：SUM(actual_score) WHERE is_right=1 AND qu_type<4
		var objScore int
		tx.Table("el_paper_qu").
			Where("paper_id = ? AND is_right = 1 AND qu_type < 4", id).
			Select("COALESCE(SUM(actual_score), 0)").Scan(&objScore)
		paper.ObjScore = objScore
		paper.UserScore = objScore
		paper.SubjScore = 0
		now := time.Now()
		if paper.HasSaq == 1 {
			paper.State = paperStateWaitOpt
		} else {
			paper.State = paperStateFinished
			// 同步保存 user_exam 成绩（若存在 user_id / exam_id）
			if paper.UserID != "" && paper.ExamID != "" {
				_ = h.joinUserExamResult(tx, paper.UserID, paper.ExamID, objScore, objScore >= paper.QualifyScore)
			}
		}
		paper.UpdateTime = &now
		if paper.CreateTime != nil {
			dur := int(time.Since(*paper.CreateTime).Minutes())
			if dur < 1 {
				dur = 1
			}
			paper.UserTime = dur
		}
		return tx.Save(&paper).Error
	})
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, true)
}

// joinUserExamResult 对齐 UserExamService.joinResult：
// 存在则更新 max_score + passed；不存在则插入一行 try_count=1
func (h *PaperHandler) joinUserExamResult(tx *gorm.DB, userID, examID string, score int, passed bool) error {
	var ue model.UserExam
	err := tx.Where("user_id = ? AND exam_id = ?", userID, examID).First(&ue).Error
	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		passedFlag := int8(0)
		if passed {
			passedFlag = 1
		}
		ue = model.UserExam{
			ID:         strconv.FormatInt(nextID(), 10),
			UserID:     userID,
			ExamID:     examID,
			TryCount:   1,
			MaxScore:   score,
			Passed:     passedFlag,
			CreateTime: &now,
			UpdateTime: &now,
		}
		return tx.Create(&ue).Error
	}
	if err != nil {
		return err
	}
	ue.TryCount++
	if score > ue.MaxScore {
		ue.MaxScore = score
	}
	if passed {
		ue.Passed = 1
	}
	ue.UpdateTime = &now
	return tx.Save(&ue).Error
}

// POST /exam/api/paper/paper/paper-result {id=paperId}
// 对齐 paperResult：paper + 题目+答案列表
func (h *PaperHandler) PaperResult(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	var paper model.Paper
	if err := h.db.Where("id = ?", id).First(&paper).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	type quRow struct {
		model.PaperQu
		Content string `gorm:"column:content" json:"content"`
		Image   string `gorm:"column:image"   json:"image"`
	}
	var qus []quRow
	h.db.Table("el_paper_qu AS pq").
		Select("pq.*, eq.content, eq.image").
		Joins("LEFT JOIN el_qu eq ON pq.qu_id = eq.id").
		Where("pq.paper_id = ?", id).
		Order("pq.sort ASC").
		Find(&qus)
	response.Rest(c, gin.H{
		"id":           paper.ID,
		"title":        paper.Title,
		"examId":       paper.ExamID,
		"userId":       paper.UserID,
		"totalScore":   paper.TotalScore,
		"qualifyScore": paper.QualifyScore,
		"userScore":    paper.UserScore,
		"objScore":     paper.ObjScore,
		"subjScore":    paper.SubjScore,
		"state":        paper.State,
		"userTime":     paper.UserTime,
		"quList":       qus,
	})
}

// POST /exam/api/paper/paper/paperQu-detail {id=paperId}
// 对齐 paperQuDetail：paper + 各类型题目的完整详情（含答案）
func (h *PaperHandler) PaperQuDetail(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	var paper model.Paper
	if err := h.db.Where("id = ?", id).First(&paper).Error; err != nil {
		response.RestErr(c, "试卷不存在")
		return
	}
	var er model.ExamRepo
	var repo model.Repo
	if err := h.db.Where("exam_id = ?", paper.ExamID).Take(&er).Error; err == nil {
		if err2 := h.db.Where("id = ?", er.RepoID).Take(&repo).Error; err2 != nil {
			slog.Warn("paper: repo not found", "repoId", er.RepoID, "error", err2)
		}
	}
	var qs []model.PaperQu
	h.db.Where("paper_id = ?", id).Order("sort ASC").Find(&qs)

	buildFull := func(pq model.PaperQu) gin.H {
		var qu model.Qu
		h.db.Where("id = ?", pq.QuID).Take(&qu)
		var answers []paperQuAnswerView
		h.db.Table("el_paper_qu_answer AS pa").
			Select("pa.id, pa.paper_id, pa.qu_id, pa.answer_id, pa.checked, pa.sort, pa.abc, qa.is_right, qa.content, qa.image, qa.score").
			Joins("LEFT JOIN el_qu_answer qa ON pa.answer_id = qa.id").
			Where("pa.paper_id = ? AND pa.qu_id = ?", pq.PaperID, pq.QuID).
			Order("pa.sort ASC").
			Find(&answers)
		return gin.H{
			"id": pq.ID, "paperId": pq.PaperID, "quId": pq.QuID,
			"quType": pq.QuType, "answered": pq.Answered, "answer": pq.Answer,
			"sort": pq.Sort, "score": pq.Score, "actualScore": pq.ActualScore,
			"isRight": pq.IsRight, "content": qu.Content, "title": qu.Title, "image": qu.Image,
			"analysis": qu.Analysis, "answerList": answers,
		}
	}
	radio, multi, judge := []gin.H{}, []gin.H{}, []gin.H{}
	for _, q := range qs {
		full := buildFull(q)
		switch q.QuType {
		case quTypeRadio:
			radio = append(radio, full)
		case quTypeMulti:
			multi = append(multi, full)
		case quTypeJudge:
			judge = append(judge, full)
		}
	}
	response.Rest(c, gin.H{
		"id":           paper.ID,
		"title":        paper.Title,
		"examId":       paper.ExamID,
		"totalScore":   paper.TotalScore,
		"totalTime":    paper.TotalTime,
		"qualifyScore": paper.QualifyScore,
		"state":        paper.State,
		"repo":         repo,
		"radioList":    radio,
		"multiList":    multi,
		"judgeList":    judge,
	})
}

// POST /exam/api/paper/paper/training {id}
// 练习模式展示：与 paper-result 同形即可（Java 行为一致）
func (h *PaperHandler) Training(c *gin.Context) { h.PaperResult(c) }

// POST /exam/api/paper/paper/show_pdf {id=paperId}
// 对齐 showPdf：返回 {pdfPath, pdfFlag} 简化视图
func (h *PaperHandler) ShowPdf(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	var r struct {
		PdfPath string `gorm:"column:pdf_path"`
		PdfFlag *int   `gorm:"column:pdf_flag"`
	}
	if err := h.db.Table("el_tester").
		Where("paper_id = ?", id).
		Select("pdf_path, pdf_flag").
		Take(&r).Error; err != nil {
		// 兜底查 el_candidate（开放模式）
		h.db.Table("el_candidate").
			Where("paper_id = ?", id).
			Select("pdf_path, pdf_flag").
			Take(&r)
	}
	response.Rest(c, gin.H{"pdfPath": r.PdfPath, "pdfFlag": r.PdfFlag})
}

// ===================== 阅卷 + 标准分 =====================

// POST /exam/api/paper/paper/review-paper
// 对齐 PaperServiceImpl.reviewPaper：提交阅卷（主观题打分）
func (h *PaperHandler) ReviewPaper(c *gin.Context) {
	var req struct {
		ID     string `json:"id"`
		QuList []struct {
			ID          string `json:"id"`
			ActualScore int    `json:"actualScore"`
		} `json:"quList"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == "" {
		response.RestErr(c, "参数错误")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新每道题的 actual_score + is_right=1
		for _, q := range req.QuList {
			tx.Model(&model.PaperQu{}).Where("id = ?", q.ID).
				Updates(map[string]any{"actual_score": q.ActualScore, "is_right": 1})
		}
		// 2. SUM 主观分（qu_type >= 4 对齐 Java subjective）
		var subjScore int
		tx.Table("el_paper_qu").
			Where("paper_id = ? AND qu_type >= 4", req.ID).
			Select("COALESCE(SUM(actual_score), 0)").Scan(&subjScore)
		// 3. 更新 paper
		var paper model.Paper
		if err := tx.Where("id = ?", req.ID).First(&paper).Error; err != nil {
			return err
		}
		paper.SubjScore = subjScore
		paper.UserScore = paper.ObjScore + subjScore
		paper.State = paperStateFinished
		now := time.Now()
		paper.UpdateTime = &now
		if err := tx.Save(&paper).Error; err != nil {
			return err
		}
		// 4. 同步 user_exam
		if paper.UserID != "" && paper.ExamID != "" {
			_ = h.joinUserExamResult(tx, paper.UserID, paper.ExamID,
				paper.UserScore, paper.UserScore >= paper.QualifyScore)
		}
		return nil
	})
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, true)
}

// POST /exam/api/paper/paper/stand-score  {id=paperId}
// 对齐 PaperServiceImpl.criterionScore：返回 int[12]（原始分，无常模换算）
func (h *PaperHandler) PaperStandScore(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	var qs []struct {
		Content string `gorm:"column:content"`
		IsRight int8   `gorm:"column:is_right"`
	}
	h.db.Table("el_paper_qu AS pq").
		Select("eq.content, pq.is_right").
		Joins("LEFT JOIN el_qu eq ON pq.qu_id = eq.id").
		Where("pq.paper_id = ?", id).
		Order("pq.sort ASC").
		Find(&qs)

	vars := make(map[string]float64, 91)
	for i := 1; i <= 91; i++ {
		vars[fmt.Sprintf("V%d", i)] = 0
	}
	for _, q := range qs {
		if q.IsRight == 1 {
			vars[q.Content] = 1
		}
	}

	exprs := []string{
		"1-V1+1-V36+V41+V56+1-V61+V81+1-V15",
		"V12+V26+V38+1-V49+V70+1-V77+1-V86+V21",
		"1-V11+1-V31+1-V47+V57+V75+V84+1-V19",
		"V3+V35+V46+V54+V63+1-V78+V14",
		"1-V4+1-V28+1-V40+V51+1-V72+V87+1-V23",
		"1-V2+V30+1-V44+V59+V74+1-V89+V17",
		"V7+1-V33+V42+V52+V68+1-V80+V20",
		"V9+1-V37+V32+1-V53+1-V66+V85+1-V16",
		"1-V10+1-V27+V48+1-V58+1-V64+1-V82+V18",
		"V5+1-V34+1-V43+1-V50+1-V67+1-V79+1-V24",
		"V6+1-V39+V25+V55+1-V73+V83+1-V90+V65+V71+1-V22",
		"1-V8+V29+V45+1-V60+1-V69+V76+V88+V62+V13",
	}
	scores := make([]int, 12)
	for i, expr := range exprs {
		v, err := evalExpr(expr, vars)
		if err != nil {
			continue
		}
		s := int(v + 0.5)
		if s < 0 {
			s = 0
		} else if s > 10 {
			s = 10
		}
		scores[i] = s
	}
	response.Rest(c, scores)
}
