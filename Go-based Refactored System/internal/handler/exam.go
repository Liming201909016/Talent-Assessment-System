package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// ExamHandler 对应 /exam/api/exam/exam/*
type ExamHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewExamHandler(db *gorm.DB, cfg *config.Config) *ExamHandler {
	return &ExamHandler{db: db, cfg: cfg}
}

type examPagingReq struct {
	Current int         `json:"current"`
	Size    int         `json:"size"`
	Title   string      `json:"title"`
	State   interface{} `json:"state"`
	IsOpen  interface{} `json:"isOpen"`
	Params  struct {
		Title     string   `json:"title"`
		RepoIds   []string `json:"repoIds"`
		StartTime string   `json:"startTime"`
		EndTime   string   `json:"endTime"`
	} `json:"params"`
}

// POST /exam/api/exam/exam/paging
func (h *ExamHandler) Paging(c *gin.Context) {
	var req examPagingReq
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	req.Size = capPageSize(req.Size)
	q := h.db.Model(&model.Exam{})
	searchTitle := req.Title
	if searchTitle == "" {
		searchTitle = req.Params.Title
	}
	if searchTitle != "" {
		q = q.Where("title like ?", "%"+searchTitle+"%")
	}
	if st := toIntPtr(req.State); st != nil {
		q = q.Where("state = ?", *st)
	}
	if op := toIntPtr(req.IsOpen); op != nil {
		q = q.Where("is_open = ?", *op)
	}
	var total int64
	q.Count(&total)
	var rows []model.Exam
	q.Order("update_time desc").Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	// 附加 repoCode + stuFlag（对齐 Java ExamDTO 聚合字段）
	result := h.enrichExamRows(rows)
	response.Rest(c, gin.H{"records": result, "total": total, "current": req.Current, "size": req.Size})
}

// POST /exam/api/exam/exam/online-paging (在线测评列表)
// 对齐 Java ExamMapper.online：按 open_type 过滤可见考试
//
//	open_type=1(全部公开) 或 open_type=3 → 所有人可见
//	open_type=2(部门限定) → 需通过 el_exam_depart + sys_user 匹配（简化：匿名端点无 userId，暂不过滤部门）
//
// 注意：不按 is_open 过滤！is_open 是测评人员来源方式（1=开放/2=封闭），不是可见性控制
func (h *ExamHandler) OnlinePaging(c *gin.Context) {
	var req examPagingReq
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	q := h.db.Model(&model.Exam{})
	// 对齐 Java：open_type IN (1, 2, 3) — 即不做 open_type 限制，显示全部考试
	// Java 原始 SQL：WHERE (ex.open_type=1 OR ex.open_type=3 OR uc.user_id=#{userId})
	// 简化处理：匿名端点无法获取 userId，直接返回全部（与 admin 登录行为一致）
	searchTitle := req.Title
	if searchTitle == "" {
		searchTitle = req.Params.Title
	}
	if searchTitle != "" {
		q = q.Where("title like ?", "%"+searchTitle+"%")
	}
	var total int64
	q.Count(&total)
	var rows []model.Exam
	q.Order("update_time desc").Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	result := h.enrichExamRows(rows)
	response.Rest(c, gin.H{"records": result, "total": total, "current": req.Current, "size": req.Size})
}

// fmtExamTime 将 *time.Time 格式化为前端 el-date-picker 需要的 "yyyy-MM-dd HH:mm"
func fmtExamTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02 15:04")
}

// enrichExamRows 对齐 Java ExamDTO：为每条 exam 附加 repoCode/repoIds/stuFlag
func (h *ExamHandler) enrichExamRows(rows []model.Exam) []gin.H {
	result := make([]gin.H, len(rows))
	for i, e := range rows {
		r := gin.H{
			"id": e.ID, "title": e.Title, "content": e.Content,
			"openType": e.OpenType, "joinType": e.JoinType, "isOpen": e.IsOpen,
			"answerType": e.AnswerType, "level": e.Level, "state": e.State,
			"timeLimit": e.TimeLimit != 0, "showPdf": e.ShowPdf != 0,
			"startTime": fmtExamTime(e.StartTime), "endTime": fmtExamTime(e.EndTime),
			"createTime": e.CreateTime, "updateTime": e.UpdateTime,
			"totalScore": e.TotalScore, "totalTime": e.TotalTime,
			"qualifyScore": e.QualifyScore, "pdfPath": e.PdfPath,
		}
		// repoCode: 取第一个关联题库的 code
		var repoCode string
		h.db.Table("el_exam_repo AS er").
			Joins("INNER JOIN el_repo AS rp ON rp.id = er.repo_id").
			Where("er.exam_id = ?", e.ID).
			Limit(1).Pluck("rp.code", &repoCode)
		r["repoCode"] = repoCode
		// repoIds
		var repoIDs []string
		h.db.Model(&model.ExamRepo{}).Where("exam_id = ?", e.ID).Pluck("repo_id", &repoIDs)
		r["repoIds"] = repoIDs
		// stuFlag: 从 DB 读取
		r["stuFlag"] = e.StuFlag
		result[i] = r
	}
	return result
}

// POST /exam/api/exam/exam/detail
func (h *ExamHandler) Detail(c *gin.Context) {
	id := bindID(c)
	if id == "" {
		response.RestErr(c, "id 为空")
		return
	}
	var e model.Exam
	if err := h.db.Where("id = ?", id).First(&e).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	var repos []model.ExamRepo
	h.db.Where("exam_id = ?", id).Find(&repos)
	// 为每个 repo 项补充 repoCode（前端编辑表单需要）
	type repoWithCode struct {
		model.ExamRepo
		RepoCode string `json:"repoCode"`
	}
	repoCodeMap := map[string]string{}
	if len(repos) > 0 {
		var repoIDs2 []string
		for _, r := range repos {
			repoIDs2 = append(repoIDs2, r.RepoID)
		}
		type rc struct {
			ID   string `gorm:"column:id"`
			Code string `gorm:"column:code"`
		}
		var rcs []rc
		h.db.Table("el_repo").Where("id IN ?", repoIDs2).Find(&rcs)
		for _, r := range rcs {
			repoCodeMap[r.ID] = r.Code
		}
	}
	var reposOut []repoWithCode
	for _, r := range repos {
		reposOut = append(reposOut, repoWithCode{ExamRepo: r, RepoCode: repoCodeMap[r.RepoID]})
	}
	var departs []string
	h.db.Model(&model.ExamDepart{}).Where("exam_id = ?", id).Pluck("depart_id", &departs)
	var repoIDs []string
	for _, r := range repos {
		repoIDs = append(repoIDs, r.RepoID)
	}
	var repoCode string
	h.db.Table("el_exam_repo AS er").
		Joins("INNER JOIN el_repo AS rp ON rp.id = er.repo_id").
		Where("er.exam_id = ?", id).Limit(1).Pluck("rp.code", &repoCode)
	// 对齐 Java ExamSaveReqDTO：扁平字段 + repoList + departIds
	response.Rest(c, gin.H{
		"id": e.ID, "title": e.Title, "content": e.Content,
		"openType": e.OpenType, "joinType": e.JoinType, "isOpen": e.IsOpen,
		"answerType": e.AnswerType, "level": e.Level, "state": e.State,
		"timeLimit": e.TimeLimit != 0, "showPdf": e.ShowPdf != 0,
		"startTime": fmtExamTime(e.StartTime), "endTime": fmtExamTime(e.EndTime),
		"createTime": e.CreateTime, "updateTime": e.UpdateTime,
		"totalScore": e.TotalScore, "totalTime": e.TotalTime,
		"qualifyScore": e.QualifyScore, "pdfPath": e.PdfPath,
		"requiredFields": e.RequiredFields,
		"repoList":       reposOut, "departIds": departs,
		"repoCode": repoCode, "repoIds": repoIDs,
		"stuFlag": e.StuFlag,
	})
}

// Java 常量对齐：JoinType.REPO_JOIN=1, OpenType.DEPT_OPEN=2
const (
	joinTypeRepoJoin = 1
	openTypeDeptOpen = 2
)

// POST /exam/api/exam/exam/save
//
// 业务逻辑与 ExamServiceImpl.save 对齐：
//  1. calcScore：当 joinType=REPO_JOIN 时按 repoList 计算 totalScore = Σ(radio/multi/judge count×score)。
//  2. 状态修复：timeLimit=false 且 state=2 时，state 置 0。
//  3. 仅当 joinType=REPO_JOIN 才保存 exam_repo；仅当 openType=DEPT_OPEN 才保存 exam_depart。
func (h *ExamHandler) Save(c *gin.Context) {
	// 先读 raw JSON
	rawBytes, _ := io.ReadAll(c.Request.Body)
	var rawMap map[string]interface{}
	json.Unmarshal(rawBytes, &rawMap)

	// 用扁平 struct 接收（避免 model.Exam 嵌入导致的 JSON tag 冲突）
	var body struct {
		ID             string           `json:"id"`
		Title          string           `json:"title"`
		Content        string           `json:"content"`
		OpenType       int              `json:"openType"`
		JoinType       int              `json:"joinType"`
		IsOpen         int              `json:"isOpen"`
		AnswerType     int              `json:"answerType"`
		Level          int              `json:"level"`
		State          int              `json:"state"`
		TotalScore     int              `json:"totalScore"`
		TotalTime      int              `json:"totalTime"`
		QualifyScore   int              `json:"qualifyScore"`
		PdfPath        string           `json:"pdfPath"`
		RequiredFields string           `json:"requiredFields"`
		StuFlag        int8             `json:"stuFlag"`
		RepoList       []model.ExamRepo `json:"repoList"`
		DepartIDs      []string         `json:"departIds"`
	}
	if err := json.Unmarshal(rawBytes, &body); err != nil {
		slog.Info("exam-save: unmarshal error", "value", err)
		response.RestErr(c, "参数错误")
		return
	}

	// 处理冲突字段：startTime/endTime/timeLimit/showPdf
	parseTime := func(raw interface{}) *time.Time {
		if raw == nil {
			return nil
		}
		s, ok := raw.(string)
		if !ok || s == "" {
			return nil
		}
		for _, layout := range []string{"2006-01-02 15:04", "2006-01-02 15:04:05", time.RFC3339} {
			if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
				return &t
			}
		}
		return nil
	}
	startTime := parseTime(rawMap["startTime"])
	endTime := parseTime(rawMap["endTime"])
	var timeLimit int8
	if v, ok := rawMap["timeLimit"]; ok && v != nil {
		switch vv := v.(type) {
		case bool:
			if vv {
				timeLimit = 1
			}
		case float64:
			timeLimit = int8(vv)
		case string:
			if vv == "true" || vv == "1" {
				timeLimit = 1
			}
		}
	}
	var showPdf int8
	if v, ok := rawMap["showPdf"]; ok && v != nil {
		switch vv := v.(type) {
		case bool:
			if vv {
				showPdf = 1
			}
		case float64:
			showPdf = int8(vv)
		case string:
			if vv == "true" || vv == "1" {
				showPdf = 1
			}
		}
	}

	// 映射到 model.Exam
	exam := model.Exam{
		ID: body.ID, Title: body.Title, Content: body.Content,
		OpenType: body.OpenType, JoinType: body.JoinType, IsOpen: body.IsOpen,
		AnswerType: body.AnswerType, Level: body.Level, State: body.State,
		TotalScore: body.TotalScore, TotalTime: body.TotalTime, QualifyScore: body.QualifyScore,
		PdfPath: body.PdfPath, RequiredFields: body.RequiredFields, StuFlag: body.StuFlag,
		StartTime: startTime, EndTime: endTime, TimeLimit: timeLimit, ShowPdf: showPdf,
	}

	// 1. 计算总分（仅题库组卷）
	if body.JoinType == joinTypeRepoJoin {
		obj := 0
		for _, item := range body.RepoList {
			if item.RadioCount > 0 && item.RadioScore > 0 {
				obj += item.RadioCount * item.RadioScore
			}
			if item.MultiCount > 0 && item.MultiScore > 0 {
				obj += item.MultiCount * item.MultiScore
			}
			if item.JudgeCount > 0 && item.JudgeScore > 0 {
				obj += item.JudgeCount * item.JudgeScore
			}
		}
		exam.TotalScore = obj
	}

	// 2. 状态修复：非限时 + state=2 → state=0
	if exam.TimeLimit == 0 && exam.State == 2 {
		exam.State = 0
	}

	now := time.Now()
	err := h.db.Transaction(func(tx *gorm.DB) error {
		isNew := exam.ID == ""
		if isNew {
			exam.ID = strconv.FormatInt(nextID(), 10)
			exam.CreateTime = &now
		} else {
			// 保留原有 create_time（Save 会覆盖所有字段）
			var orig model.Exam
			if err := tx.Select("create_time").Where("id = ?", exam.ID).Take(&orig).Error; err == nil && orig.CreateTime != nil {
				exam.CreateTime = orig.CreateTime
			} else {
				exam.CreateTime = &now // 原值为空时用当前时间
			}
		}
		exam.UpdateTime = &now
		if isNew {
			if err := tx.Create(&exam).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(&exam).Error; err != nil {
				return err
			}
			tx.Where("exam_id = ?", exam.ID).Delete(&model.ExamRepo{})
			tx.Where("exam_id = ?", exam.ID).Delete(&model.ExamDepart{})
		}

		// 3a. 仅题库组卷保存 exam_repo
		if body.JoinType == joinTypeRepoJoin {
			seen := map[string]bool{}
			for i := range body.RepoList {
				r := body.RepoList[i]
				if r.RepoID == "" {
					continue
				}
				if seen[r.RepoID] {
					return gorm.ErrDuplicatedKey
				}
				seen[r.RepoID] = true
				r.ID = strconv.FormatInt(nextID()+int64(i), 10)
				r.ExamID = exam.ID
				if err := tx.Create(&r).Error; err != nil {
					return err
				}
			}
		}

		// 3b. 仅部门开放保存 exam_depart
		if body.OpenType == openTypeDeptOpen {
			for i, did := range body.DepartIDs {
				if did == "" {
					continue
				}
				d := model.ExamDepart{
					ID:       strconv.FormatInt(nextID()+int64(200+i), 10),
					ExamID:   exam.ID,
					DepartID: did,
				}
				if err := tx.Create(&d).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		if err == gorm.ErrDuplicatedKey {
			response.RestErr(c, "不能选择重复的题库！")
			return
		}
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, exam)
}

// POST /exam/api/exam/exam/delete
func (h *ExamHandler) Delete(c *gin.Context) {
	var b struct {
		IDs []string `json:"ids"`
	}
	_ = c.ShouldBindJSON(&b)
	if len(b.IDs) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}

	// FB-021: 删除前检查是否有关联数据，避免孤儿
	// 业务规则：有 tester / candidate / paper 关联的 exam 不可直接删
	var testerCount, candidateCount, paperCount int64
	h.db.Table("el_tester").Where("exam_id IN ? AND (del_flag IS NULL OR del_flag = '0')", b.IDs).Count(&testerCount)
	h.db.Table("el_candidate").Where("exam_id IN ? AND (del_flag IS NULL OR del_flag = '0' OR del_flag = 0)", b.IDs).Count(&candidateCount)
	h.db.Table("el_paper").Where("exam_id IN ?", b.IDs).Count(&paperCount)
	if testerCount > 0 || candidateCount > 0 || paperCount > 0 {
		response.RestErr(c, fmt.Sprintf("无法删除：含 %d 个测评者、%d 个考生、%d 份试卷。请先清理关联数据", testerCount, candidateCount, paperCount))
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id IN ?", b.IDs).Delete(&model.Exam{}).Error; err != nil {
			return err
		}
		// 配置型关联表（无业务数据），可一并删除
		tx.Where("exam_id IN ?", b.IDs).Delete(&model.ExamRepo{})
		tx.Where("exam_id IN ?", b.IDs).Delete(&model.ExamDepart{})
		return nil
	})
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, true)
}

// POST /exam/api/exam/exam/state  {id, state}
func (h *ExamHandler) State(c *gin.Context) {
	var b struct {
		ID    string   `json:"id"`
		IDs   []string `json:"ids"`
		State int      `json:"state"`
	}
	_ = c.ShouldBindJSON(&b)

	// 校验 state 值有效性：0=启用 1=禁用 2=就绪 3=过期
	if b.State < 0 || b.State > 3 {
		response.RestErr(c, "无效的状态值")
		return
	}

	// 对齐 Java BaseStateReqDTO：支持 ids[] 批量，也兼容单个 id
	ids := b.IDs
	if len(ids) == 0 && b.ID != "" {
		ids = []string{b.ID}
	}
	if len(ids) == 0 {
		response.RestErr(c, "id 为空")
		return
	}
	if err := h.db.Model(&model.Exam{}).Where("id IN ?", ids).
		Update("state", b.State).Error; err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, true)
}

// POST /exam/api/exam/exam/review-paging
// 对齐 ExamServiceImpl.reviewPaging：列出含主观题(has_saq=1)的考试 + 参考人数 + 待阅数
func (h *ExamHandler) ReviewPaging(c *gin.Context) {
	var req struct {
		Current int `json:"current"`
		Size    int `json:"size"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	type row struct {
		model.Exam
		ExamUser    int `gorm:"column:exam_user"    json:"examUser"`
		UnreadPaper int `gorm:"column:unread_paper" json:"unreadPaper"`
	}
	var total int64
	h.db.Model(&model.Exam{}).Where("has_saq = 1").Count(&total)
	var rows []row
	h.db.Table("el_exam AS ex").
		Select(`ex.*,
			(SELECT COUNT(DISTINCT user_id) FROM el_paper WHERE exam_id=ex.id) AS exam_user,
			(SELECT COUNT(0) FROM el_paper WHERE exam_id=ex.id AND state=1) AS unread_paper`).
		Where("ex.has_saq = 1").
		Order("ex.update_time DESC").
		Offset((req.Current - 1) * req.Size).Limit(req.Size).
		Find(&rows)
	response.Rest(c, gin.H{
		"records": rows, "total": total,
		"current": req.Current, "size": req.Size,
	})
}
