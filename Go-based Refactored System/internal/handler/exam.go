package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/pdfgen"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// ExamHandler 对应 /exam/api/exam/exam/*
type ExamHandler struct {
	db      *gorm.DB
	cfg     *config.Config
	pdfPool *pdfgen.Pool // 可为 nil（pdfgen.enabled=false 时）

	// R4: 全局熔断 - 连续失败超阈值则在冷却窗口内不重试
	cbMu              sync.Mutex
	cbConsecutiveFail int
	cbOpenUntil       time.Time
}

func NewExamHandler(db *gorm.DB, cfg *config.Config) *ExamHandler {
	h := &ExamHandler{db: db, cfg: cfg}
	if cfg.PdfGen.Enabled {
		p, err := pdfgen.NewPool(cfg.PdfGen.ChromePath, cfg.PdfGen.PoolSize)
		if err != nil {
			slog.Error("[exam] pdfgen pool init failed", "error", err)
		} else {
			h.pdfPool = p
			slog.Info("[exam] pdfgen pool ready", "size", cfg.PdfGen.PoolSize)
		}
		// R5: 幂等添加 pdf_partial 列（标记前端数据未完全加载就生成的 PDF）
		ensurePdfPartialColumn(db)
	}
	// FB-041: 这两个迁移与 pdfgen 无关，必须无条件执行，否则封闭模式 MBTI 交卷与
	// 导出会因 el_tester 缺列而沉默失败。
	ensureTesterMbtiColumns(db)
	return h
}

// ensurePdfPartialColumn 幂等 ALTER TABLE 添加 pdf_partial 列。
// MySQL 8.0+ 支持 IF NOT EXISTS；老版本会失败但忽略（已存在时报错也无害）。
func ensurePdfPartialColumn(db *gorm.DB) {
	for _, tbl := range []string{"el_candidate", "el_tester"} {
		// 先查是否已存在
		var count int64
		db.Raw(`SELECT COUNT(*) FROM information_schema.COLUMNS
				WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = 'pdf_partial'`, tbl).
			Scan(&count)
		if count > 0 {
			continue
		}
		sql := "ALTER TABLE " + tbl + " ADD COLUMN pdf_partial TINYINT NOT NULL DEFAULT 0 COMMENT '1=报告生成时数据不完整'"
		if err := db.Exec(sql).Error; err != nil {
			slog.Warn("[exam] add pdf_partial column failed", "table", tbl, "error", err)
		} else {
			slog.Info("[exam] added pdf_partial column", "table", tbl)
		}
	}
}

// ensureTesterMbtiColumns FB-041 修复：39 生产 el_tester 缺 mbti_type/mbti_scores 列，
// 导致：
//  1. 封闭 MBTI 交卷时 UPDATE el_tester 因 Unknown column 整条原子失败，end_time
//     不写入 → 列表显示"进行中"，报告按钮不可用。
//  2. 封闭模式导出 export-raw-data SQL 因同样原因失败，xlsx 只剩标题+表头。
//
// 处理方式：启动时幂等 ALTER TABLE 补齐列；若本次刚刚补上列，再一次性回填存量
// 已完成但 end_time=NULL 的封闭 MBTI 行（基于 el_paper.update_time）。
func ensureTesterMbtiColumns(db *gorm.DB) {
	specs := []struct {
		name string
		ddl  string
	}{
		{"mbti_type", "ALTER TABLE el_tester ADD COLUMN mbti_type VARCHAR(8) NULL COMMENT 'MBTI 4-letter type'"},
		{"mbti_scores", "ALTER TABLE el_tester ADD COLUMN mbti_scores TEXT NULL COMMENT 'MBTI 8 dimension scores JSON'"},
	}
	added := false
	for _, s := range specs {
		var count int64
		db.Raw(`SELECT COUNT(*) FROM information_schema.COLUMNS
				WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'el_tester' AND COLUMN_NAME = ?`, s.name).
			Scan(&count)
		if count > 0 {
			continue
		}
		if err := db.Exec(s.ddl).Error; err != nil {
			slog.Warn("[exam] add column on el_tester failed", "column", s.name, "error", err)
			continue
		}
		slog.Info("[exam] added column on el_tester", "column", s.name)
		added = true
	}
	if !added {
		return
	}
	// 仅在本次刚刚补齐列后执行一次回填，避免每次重启都扫表。
	// 用 el_paper.update_time 作为 end_time 兜底（paper 交卷时已写 update_time）。
	res := db.Exec(`UPDATE el_tester t
		JOIN el_paper p ON p.id = t.paper_id
		SET t.end_time = COALESCE(p.update_time, NOW()),
		    t.update_time = NOW()
		WHERE t.paper_id IS NOT NULL
		  AND p.state = 2
		  AND t.end_time IS NULL`)
	if res.Error != nil {
		slog.Warn("[exam] backfill el_tester.end_time failed", "error", res.Error)
		return
	}
	slog.Info("[exam] backfilled el_tester.end_time for finished closed MBTI rows",
		"affected", res.RowsAffected)
}

// Close 关闭后端资源（chromedp pool 等）。由 main 在收到 SIGTERM 后调用。
func (h *ExamHandler) Close() {
	if h.pdfPool != nil {
		h.pdfPool.Close()
	}
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
	q = q.Where("assessment_type IS NULL OR assessment_type <> ? OR publish_status = 1", service.AssessmentTypeCompetency)
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
			"assessmentType": e.AssessmentType, "scoringMode": e.ScoringMode,
			"competencyReportAudience":        e.CompetencyReportAudience,
			"competencyProductVersion":        e.CompetencyProductVersion,
			"competencyScoringVersion":        e.CompetencyScoringVersion,
			"competencyContentVersion":        e.CompetencyContentVersion,
			"competencyReportTemplateVersion": e.CompetencyReportTemplateVersion,
			"publishStatus":                   e.PublishStatus,
			"timeLimit":                       e.TimeLimit != 0, "showPdf": e.ShowPdf != 0,
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
	dimensionIDs := make([]string, 0)
	dimensions := make([]model.ExamCompetencyDimension, 0)
	if e.AssessmentType == service.AssessmentTypeCompetency {
		if err := h.db.Model(&model.ExamCompetencyDimension{}).
			Where("exam_id = ?", id).
			Order("display_order ASC").
			Find(&dimensions).Error; err != nil {
			response.RestErr(c, "查询胜任力维度配置失败")
			return
		}
		for _, dimension := range dimensions {
			dimensionIDs = append(dimensionIDs, dimension.DimensionID)
		}
	}
	// 对齐 Java ExamSaveReqDTO：扁平字段 + repoList + departIds
	response.Rest(c, gin.H{
		"id": e.ID, "title": e.Title, "content": e.Content,
		"openType": e.OpenType, "joinType": e.JoinType, "isOpen": e.IsOpen,
		"answerType": e.AnswerType, "level": e.Level, "state": e.State,
		"assessmentType": e.AssessmentType, "scoringMode": e.ScoringMode,
		"competencyReportAudience":        e.CompetencyReportAudience,
		"competencyProductVersion":        e.CompetencyProductVersion,
		"competencyScoringVersion":        e.CompetencyScoringVersion,
		"competencyContentVersion":        e.CompetencyContentVersion,
		"competencyReportTemplateVersion": e.CompetencyReportTemplateVersion,
		"publishStatus":                   e.PublishStatus,
		"publishedAt":                     e.PublishedAt, "publishedBy": e.PublishedBy,
		"timeLimit": e.TimeLimit != 0, "showPdf": e.ShowPdf != 0,
		"startTime": fmtExamTime(e.StartTime), "endTime": fmtExamTime(e.EndTime),
		"createTime": e.CreateTime, "updateTime": e.UpdateTime,
		"totalScore": e.TotalScore, "totalTime": e.TotalTime,
		"qualifyScore": e.QualifyScore, "pdfPath": e.PdfPath,
		"requiredFields": e.RequiredFields,
		"repoList":       reposOut, "departIds": departs,
		"repoCode": repoCode, "repoIds": repoIDs,
		"dimensionIds": dimensionIDs, "competencyDimensions": dimensions,
		"stuFlag": e.StuFlag,
	})
}

// Java 常量对齐：JoinType.REPO_JOIN=1, OpenType.DEPT_OPEN=2
const (
	joinTypeRepoJoin = 1
	openTypeDeptOpen = 2
)

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	values := make(map[string]int, len(left))
	for _, value := range left {
		values[value]++
	}
	for _, value := range right {
		if values[value] == 0 {
			return false
		}
		values[value]--
	}
	return true
}

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
		ID                              string           `json:"id"`
		Title                           string           `json:"title"`
		Content                         string           `json:"content"`
		OpenType                        int              `json:"openType"`
		JoinType                        int              `json:"joinType"`
		IsOpen                          int              `json:"isOpen"`
		AnswerType                      int              `json:"answerType"`
		Level                           int              `json:"level"`
		State                           int              `json:"state"`
		TotalScore                      int              `json:"totalScore"`
		TotalTime                       int              `json:"totalTime"`
		QualifyScore                    int              `json:"qualifyScore"`
		PdfPath                         string           `json:"pdfPath"`
		RequiredFields                  string           `json:"requiredFields"`
		StuFlag                         int8             `json:"stuFlag"`
		AssessmentType                  string           `json:"assessmentType"`
		ScoringMode                     string           `json:"scoringMode"`
		CompetencyReportAudience        string           `json:"competencyReportAudience"`
		CompetencyProductVersion        string           `json:"competencyProductVersion"`
		CompetencyScoringVersion        string           `json:"competencyScoringVersion"`
		CompetencyContentVersion        string           `json:"competencyContentVersion"`
		CompetencyReportTemplateVersion string           `json:"competencyReportTemplateVersion"`
		DimensionIDs                    []string         `json:"dimensionIds"`
		RepoList                        []model.ExamRepo `json:"repoList"`
		DepartIDs                       []string         `json:"departIds"`
	}
	if err := json.Unmarshal(rawBytes, &body); err != nil {
		slog.Info("exam-save: unmarshal error", "value", err)
		response.RestErr(c, "参数错误")
		return
	}
	if body.AssessmentType == "" {
		body.AssessmentType = service.AssessmentTypeLegacy
	}
	if body.ScoringMode == "" {
		if body.AssessmentType == service.AssessmentTypeCompetency {
			body.ScoringMode = service.ScoringModeCompetencyAverage
		} else {
			body.ScoringMode = service.ScoringModeLegacy
		}
	}
	isCompetency, err := service.ValidateAssessmentMode(body.AssessmentType, body.ScoringMode)
	if err != nil {
		response.RestErr(c, "测评类型与计分模式不匹配")
		return
	}
	versions := service.CompetencyVersionSet{}
	if isCompetency {
		phase1Profile := service.NormalizePhase1CompetencyConfiguration()
		if body.TotalTime <= 0 {
			response.RestErr(c, "胜任力测评必须配置大于0的答题时长")
			return
		}
		if body.CompetencyReportAudience == "" {
			body.CompetencyReportAudience = phase1Profile.ReportAudience
		}
		if len(body.DimensionIDs) == 0 {
			body.DimensionIDs = append([]string(nil), phase1Profile.DimensionIDs...)
		}
		versions = service.CompetencyVersionSet{
			ProductVersion: body.CompetencyProductVersion, ScoringVersion: body.CompetencyScoringVersion,
			ContentVersion: body.CompetencyContentVersion, ReportTemplateVersion: body.CompetencyReportTemplateVersion,
		}
		if versions.ProductVersion == "" {
			versions.ProductVersion = phase1Profile.Versions.ProductVersion
		}
		if versions.ScoringVersion == "" {
			versions.ScoringVersion = phase1Profile.Versions.ScoringVersion
		}
		if versions.ContentVersion == "" {
			versions.ContentVersion = phase1Profile.Versions.ContentVersion
		}
		if versions.ReportTemplateVersion == "" {
			versions.ReportTemplateVersion = phase1Profile.Versions.ReportTemplateVersion
		}
		if err := service.ValidatePhase1CompetencyConfiguration(body.CompetencyReportAudience, body.DimensionIDs, versions); err != nil {
			response.RestErr(c, "00401一期固定配置只能使用基层员工、十个A/B维度和已确认版本")
			return
		}
		body.RepoList = make([]model.ExamRepo, 0)
	} else {
		body.CompetencyReportAudience = ""
		body.DimensionIDs = make([]string, 0)
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
	var reportAudience *string
	if body.CompetencyReportAudience != "" {
		reportAudience = &body.CompetencyReportAudience
	}
	exam := model.Exam{
		ID: body.ID, Title: body.Title, Content: body.Content,
		OpenType: body.OpenType, JoinType: body.JoinType, IsOpen: body.IsOpen,
		AnswerType: body.AnswerType, Level: body.Level, State: body.State,
		TotalScore: body.TotalScore, TotalTime: body.TotalTime, QualifyScore: body.QualifyScore,
		PdfPath: body.PdfPath, RequiredFields: body.RequiredFields, StuFlag: body.StuFlag,
		AssessmentType: body.AssessmentType, ScoringMode: body.ScoringMode,
		CompetencyReportAudience: reportAudience,
		StartTime:                startTime, EndTime: endTime, TimeLimit: timeLimit, ShowPdf: showPdf,
	}
	if isCompetency {
		service.ApplyCompetencyVersions(&exam, versions)
	} else {
		service.ClearCompetencyVersions(&exam)
	}

	// 1. 计算总分（仅题库组卷）
	if !isCompetency && body.JoinType == joinTypeRepoJoin {
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
	err = h.db.Transaction(func(tx *gorm.DB) error {
		isNew := exam.ID == ""
		var original model.Exam
		wasPublishedCompetency := false
		selectedDimensions := make([]model.CompetencyDimension, 0)
		enabledQuestionCounts := make(map[string]int)
		phase1QuestionInventory := make(map[string]service.Phase1QuestionTypeCounts)
		if isNew {
			exam.ID = strconv.FormatInt(nextID(), 10)
			exam.CreateTime = &now
			if isCompetency {
				exam.PublishStatus = 0
			} else {
				exam.PublishStatus = 1
			}
		} else {
			if err := tx.Where("id = ?", exam.ID).Take(&original).Error; err != nil {
				return err
			}
			if original.CreateTime != nil {
				exam.CreateTime = original.CreateTime
			} else {
				exam.CreateTime = &now
			}
			wasPublishedCompetency = original.AssessmentType == service.AssessmentTypeCompetency && original.PublishStatus == 1
			if wasPublishedCompetency {
				if !isCompetency || original.CompetencyReportAudience == nil ||
					exam.CompetencyReportAudience == nil ||
					*original.CompetencyReportAudience != *exam.CompetencyReportAudience {
					return errors.New("已发布胜任力测评不能修改报告版本")
				}
				if service.CompetencyVersionSetFromExam(original) != service.CompetencyVersionSetFromExam(exam) {
					return errors.New("已发布胜任力测评不能修改产品、评分、内容或模板版本")
				}
				var existingDimensionIDs []string
				if err := tx.Model(&model.ExamCompetencyDimension{}).
					Where("exam_id = ?", exam.ID).
					Pluck("dimension_id", &existingDimensionIDs).Error; err != nil {
					return err
				}
				if !sameStringSet(existingDimensionIDs, body.DimensionIDs) {
					return errors.New("已发布胜任力测评不能修改测评维度")
				}
			}
			if original.AssessmentType != service.AssessmentTypeCompetency && isCompetency {
				exam.PublishStatus = 0
				exam.PublishedAt = nil
				exam.PublishedBy = nil
			} else if !isCompetency {
				exam.PublishStatus = 1
				exam.PublishedAt = original.PublishedAt
				exam.PublishedBy = original.PublishedBy
			} else {
				exam.PublishStatus = original.PublishStatus
				exam.PublishedAt = original.PublishedAt
				exam.PublishedBy = original.PublishedBy
			}
		}
		if isCompetency && !wasPublishedCompetency {
			if err := tx.Where("id IN ? AND status = ?", body.DimensionIDs, 0).
				Order("display_order ASC").
				Find(&selectedDimensions).Error; err != nil {
				return err
			}
			if len(selectedDimensions) != len(body.DimensionIDs) {
				return errors.New("所选测评维度不存在或已停用")
			}
			phase1QuestionInventory, err = loadPhase1QuestionInventory(tx, body.DimensionIDs)
			if err != nil {
				return err
			}
			if err := service.ValidatePhase1QuestionInventory(body.DimensionIDs, phase1QuestionInventory); err != nil {
				return errors.New("00401一期题本必须包含每维8道维度题和1道效度题")
			}
			for _, dimension := range selectedDimensions {
				enabledQuestionCounts[dimension.ID] = phase1QuestionInventory[dimension.ID].Dimension
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
			if err := tx.Where("exam_id = ?", exam.ID).Delete(&model.ExamRepo{}).Error; err != nil {
				return err
			}
			if err := tx.Where("exam_id = ?", exam.ID).Delete(&model.ExamDepart{}).Error; err != nil {
				return err
			}
		}

		// 3a. 仅题库组卷保存 exam_repo
		if !isCompetency && body.JoinType == joinTypeRepoJoin {
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

		// 3b. 胜任力草稿保存所选维度；已发布测评只校验，不重建快照。
		if !wasPublishedCompetency {
			if err := tx.Where("exam_id = ?", exam.ID).Delete(&model.ExamCompetencyDimension{}).Error; err != nil {
				return err
			}
			if isCompetency {
				associations := make([]model.ExamCompetencyDimension, 0, len(selectedDimensions))
				for _, dimension := range selectedDimensions {
					associations = append(associations, model.ExamCompetencyDimension{
						ID:                 strconv.FormatInt(nextID(), 10),
						ExamID:             exam.ID,
						DimensionID:        dimension.ID,
						DimensionCode:      dimension.Code,
						DimensionName:      dimension.Name,
						VIRDLevel:          dimension.VIRDLevel,
						ApplicableCategory: dimension.ApplicableCategory,
						CoreMeaning:        dimension.CoreMeaning,
						DisplayOrder:       dimension.DisplayOrder,
						QuestionCount:      enabledQuestionCounts[dimension.ID],
						CreateTime:         &now,
					})
				}
				if err := tx.CreateInBatches(associations, 100).Error; err != nil {
					return err
				}
			}
		}

		// 3c. 仅部门开放保存 exam_depart
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
	var exams []model.Exam
	if err := h.db.Select("id, assessment_type").Where("id IN ?", b.IDs).Find(&exams).Error; err != nil {
		response.RestErr(c, "查询测评类型失败")
		return
	}
	competencyIDs := make([]string, 0)
	legacyIDs := make([]string, 0)
	for _, exam := range exams {
		if exam.AssessmentType == service.AssessmentTypeCompetency {
			competencyIDs = append(competencyIDs, exam.ID)
		} else {
			legacyIDs = append(legacyIDs, exam.ID)
		}
	}
	competencyReportPaths := make([]string, 0)
	if len(competencyIDs) > 0 {
		if err := h.db.Model(&model.CompetencyReport{}).Where("exam_id IN ?", competencyIDs).
			Where("pdf_path <> ''").Pluck("pdf_path", &competencyReportPaths).Error; err != nil {
			response.RestErr(c, "查询胜任力报告文件失败")
			return
		}
	}

	// FB-021: 删除前检查是否有关联数据，避免孤儿
	// 业务规则：有 tester / candidate / paper 关联的 exam 不可直接删
	var testerCount, candidateCount, paperCount int64
	if len(legacyIDs) > 0 {
		if err := h.db.Table("el_tester").Where("exam_id IN ? AND (del_flag IS NULL OR del_flag = '0')", legacyIDs).Count(&testerCount).Error; err != nil {
			response.RestErr(c, "检查测评者关联失败")
			return
		}
		if err := h.db.Table("el_candidate").Where("exam_id IN ? AND (del_flag IS NULL OR del_flag = '0' OR del_flag = 0)", legacyIDs).Count(&candidateCount).Error; err != nil {
			response.RestErr(c, "检查考生关联失败")
			return
		}
		if err := h.db.Table("el_paper").Where("exam_id IN ?", legacyIDs).Count(&paperCount).Error; err != nil {
			response.RestErr(c, "检查试卷关联失败")
			return
		}
		if testerCount > 0 || candidateCount > 0 || paperCount > 0 {
			response.RestErr(c, fmt.Sprintf("无法删除：含 %d 个测评者、%d 个考生、%d 份试卷。请先清理关联数据", testerCount, candidateCount, paperCount))
			return
		}
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if len(competencyIDs) > 0 {
			if err := deleteCompetencyExamChain(tx, competencyIDs); err != nil {
				return err
			}
		}
		if len(legacyIDs) > 0 {
			if err := tx.Where("exam_id IN ?", legacyIDs).Delete(&model.ExamRepo{}).Error; err != nil {
				return err
			}
			if err := tx.Where("exam_id IN ?", legacyIDs).Delete(&model.ExamDepart{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", legacyIDs).Delete(&model.Exam{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	reportHandler := NewCompetencyReportHandler(h.db, h)
	for _, reportPath := range competencyReportPaths {
		reportHandler.removeReportFile(reportPath)
	}
	response.Rest(c, true)
}

func deleteCompetencyExamChain(tx *gorm.DB, examIDs []string) error {
	var paperIDs []string
	if err := tx.Model(&model.Paper{}).Where("exam_id IN ?", examIDs).Pluck("id", &paperIDs).Error; err != nil {
		return err
	}
	if len(paperIDs) > 0 {
		if err := tx.Exec("DELETE FROM el_competency_report_audit WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM el_competency_report WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM el_competency_group_result WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM el_competency_validity_result WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM el_competency_dimension_result WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM el_competency_result WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM el_paper_qu_answer WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM el_paper_qu WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM el_mbti_answer WHERE paper_id IN ?", paperIDs).Error; err != nil {
			return err
		}
	}
	if err := tx.Exec("DELETE FROM el_user_exam WHERE exam_id IN ?", examIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM el_candidate WHERE exam_id IN ?", examIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM el_tester WHERE exam_id IN ?", examIDs).Error; err != nil {
		return err
	}
	if len(paperIDs) > 0 {
		if err := tx.Where("id IN ?", paperIDs).Delete(&model.Paper{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Exec("DELETE FROM el_exam_competency_question WHERE exam_id IN ?", examIDs).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id IN ?", examIDs).Delete(&model.ExamCompetencyDimension{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id IN ?", examIDs).Delete(&model.ExamCompetencyGroup{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id IN ?", examIDs).Delete(&model.ExamRepo{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id IN ?", examIDs).Delete(&model.ExamDepart{}).Error; err != nil {
		return err
	}
	return tx.Where("id IN ?", examIDs).Delete(&model.Exam{}).Error
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
