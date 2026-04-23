package handler

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// MbtiHandler MBTI 职业性格测验独立模块
// 使用 el_mbti_answer 表，与现有 el_paper_qu_answer 完全隔离
type MbtiHandler struct {
	db            *gorm.DB
	reportHandler *MbtiReportHandler
}

func NewMbtiHandler(db *gorm.DB) *MbtiHandler { return &MbtiHandler{db: db} }

// SetReportHandler 注入报告生成器（在 router 中设置，避免循环依赖）
func (h *MbtiHandler) SetReportHandler(rh *MbtiReportHandler) { h.reportHandler = rh }

// mbtiAnswer 对应 el_mbti_answer 表
type mbtiAnswer struct {
	ID         string     `gorm:"column:id;primaryKey" json:"id"`
	PaperID    string     `gorm:"column:paper_id"      json:"paperId"`
	QuID       string     `gorm:"column:qu_id"         json:"quId"`
	ScoreA     int        `gorm:"column:score_a"        json:"scoreA"`
	ScoreB     int        `gorm:"column:score_b"        json:"scoreB"`
	Sort       int        `gorm:"column:sort"           json:"sort"`
	Answered   int8       `gorm:"column:answered"       json:"answered"`
	CreateTime *time.Time `gorm:"column:create_time"    json:"createTime"`
}

func (mbtiAnswer) TableName() string { return "el_mbti_answer" }

// ===================== POST /exam/api/mbti/paper-detail =====================
// 加载 48 题 + 已填分值（答题页用）
func (h *MbtiHandler) PaperDetail(c *gin.Context) {
	var b struct {
		PaperID string `json:"paperId"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}

	// 获取 paper 信息（计算剩余时间）
	var paper struct {
		ExamID     string     `gorm:"column:exam_id"`
		TotalTime  int        `gorm:"column:total_time"`
		State      int        `gorm:"column:state"`
		CreateTime *time.Time `gorm:"column:create_time"`
	}
	if err := h.db.Table("el_paper").Where("id = ?", b.PaperID).First(&paper).Error; err != nil {
		response.RestErr(c, "试卷不存在")
		return
	}

	// 获取考试关联的题库 code
	var repoCode string
	h.db.Table("el_exam_repo AS er").
		Joins("INNER JOIN el_repo AS rp ON rp.id = er.repo_id").
		Where("er.exam_id = ?", paper.ExamID).Limit(1).Pluck("rp.code", &repoCode)

	// 查 el_qu + el_qu_answer (通过 el_qu_repo 关联到 00301 题库)
	type quRow struct {
		QuID    string `gorm:"column:qu_id"   json:"quId"`
		Content string `gorm:"column:content" json:"content"`
		Title   string `gorm:"column:title"   json:"title"`
		Sort    int    `gorm:"column:sort"    json:"sort"`
	}
	var quList []quRow
	h.db.Table("el_qu_repo AS qr").
		Joins("INNER JOIN el_qu AS q ON q.id = qr.qu_id").
		Joins("INNER JOIN el_exam_repo AS er ON er.repo_id = qr.repo_id").
		Where("er.exam_id = ?", paper.ExamID).
		Select("qr.qu_id, q.content, q.title, qr.sort").
		Order("qr.sort ASC").
		Scan(&quList)

	// 查每题的 A/B 选项文字
	type ansOpt struct {
		QuID    string `gorm:"column:qu_id"`
		Content string `gorm:"column:content"`
	}
	var ansOpts []ansOpt
	quIDs := make([]string, len(quList))
	for i, q := range quList {
		quIDs[i] = q.QuID
	}
	if len(quIDs) > 0 {
		h.db.Table("el_qu_answer").Where("qu_id IN ?", quIDs).
			Select("qu_id, content").Order("qu_id, id ASC").
			Scan(&ansOpts)
	}
	optMap := map[string][2]string{} // quId → [optionA, optionB]
	for _, a := range ansOpts {
		pair := optMap[a.QuID]
		if pair[0] == "" {
			pair[0] = a.Content // 第一条 = A
		} else {
			pair[1] = a.Content // 第二条 = B
		}
		optMap[a.QuID] = pair
	}

	// 查已填答案
	var existing []mbtiAnswer
	h.db.Where("paper_id = ?", b.PaperID).Find(&existing)
	ansMap := map[string]mbtiAnswer{}
	for _, a := range existing {
		ansMap[a.QuID] = a
	}

	// 组装返回
	type quOut struct {
		QuID     string `json:"quId"`
		Content  string `json:"content"`
		Title    string `json:"title"`
		OptionA  string `json:"optionA"`
		OptionB  string `json:"optionB"`
		ScoreA   int    `json:"scoreA"`
		ScoreB   int    `json:"scoreB"`
		Answered bool   `json:"answered"`
		Sort     int    `json:"sort"`
	}
	result := make([]quOut, len(quList))
	for i, q := range quList {
		opts := optMap[q.QuID]
		out := quOut{
			QuID: q.QuID, Content: q.Content, Title: q.Title,
			OptionA: opts[0], OptionB: opts[1], Sort: q.Sort,
		}
		if ans, ok := ansMap[q.QuID]; ok && ans.Answered == 1 {
			out.ScoreA = ans.ScoreA
			out.ScoreB = ans.ScoreB
			out.Answered = true
		}
		result[i] = out
	}

	// 计算剩余时间
	leftSeconds := 99999
	if paper.TotalTime > 0 && paper.CreateTime != nil {
		elapsed := int(time.Since(*paper.CreateTime).Seconds())
		leftSeconds = paper.TotalTime*60 - elapsed
		if leftSeconds < 0 {
			leftSeconds = 0
		}
	}

	response.Rest(c, gin.H{
		"quList":      result,
		"leftSeconds": leftSeconds,
		"state":       paper.State,
		"repoCode":    repoCode,
	})
}

// ===================== POST /exam/api/mbti/fill-answer =====================
// 保存单题 AB 赋分（校验和=5）
func (h *MbtiHandler) FillAnswer(c *gin.Context) {
	var b struct {
		PaperID string `json:"paperId"`
		QuID    string `json:"quId"`
		ScoreA  int    `json:"scoreA"`
		ScoreB  int    `json:"scoreB"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if b.PaperID == "" || b.QuID == "" {
		response.RestErr(c, "paperId 或 quId 为空")
		return
	}
	if b.ScoreA < 0 || b.ScoreB < 0 || b.ScoreA > 5 || b.ScoreB > 5 {
		response.RestErr(c, "分值必须在 0-5 之间")
		return
	}
	if b.ScoreA+b.ScoreB != 5 {
		response.RestErr(c, "AB 选项分值之和必须为 5")
		return
	}

	// Upsert: 存在则更新，不存在则插入
	var existing mbtiAnswer
	err := h.db.Where("paper_id = ? AND qu_id = ?", b.PaperID, b.QuID).First(&existing).Error
	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		h.db.Create(&mbtiAnswer{
			ID:      strconv.FormatInt(time.Now().UnixNano(), 10),
			PaperID: b.PaperID, QuID: b.QuID,
			ScoreA: b.ScoreA, ScoreB: b.ScoreB,
			Answered: 1, CreateTime: &now,
		})
	} else {
		h.db.Model(&existing).Updates(map[string]interface{}{
			"score_a": b.ScoreA, "score_b": b.ScoreB, "answered": 1,
		})
	}
	response.Rest(c, true)
}

// ===================== POST /exam/api/mbti/submit =====================
// 交卷：更新 paper 状态 + 计算 MBTI 类型
func (h *MbtiHandler) Submit(c *gin.Context) {
	var b struct {
		PaperID string `json:"paperId"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}

	// 更新 paper 状态为已交卷 + 计算用时
	now := time.Now()
	var paper struct {
		CreateTime *time.Time `gorm:"column:create_time"`
	}
	h.db.Table("el_paper").Where("id = ?", b.PaperID).Select("create_time").Take(&paper)
	userTime := 0
	if paper.CreateTime != nil {
		userTime = int(now.Sub(*paper.CreateTime).Minutes())
		if userTime < 1 {
			userTime = 1
		}
	}
	h.db.Table("el_paper").Where("id = ? AND state = 0", b.PaperID).
		Updates(map[string]interface{}{"state": 2, "update_time": &now, "user_time": userTime})

	// 计算分数
	scores, mbtiType := h.calcMbtiScores(b.PaperID)

	// 持久化 MBTI 类型到 el_tester（通过 paper_id 关联）
	scoresJSON, _ := json.Marshal(scores)
	h.db.Table("el_tester").Where("paper_id = ?", b.PaperID).
		Updates(map[string]interface{}{
			"mbti_type":   mbtiType,
			"mbti_scores": string(scoresJSON),
			"end_time":    &now,
			"update_time": &now,
		})
	// 同时更新 el_candidate（如果存在）
	h.db.Table("el_candidate").Where("paper_id = ?", b.PaperID).
		Updates(map[string]interface{}{
			"end_time":    &now,
			"update_time": &now,
		})

	response.Rest(c, gin.H{
		"type":   mbtiType,
		"scores": scores,
	})

	// 异步生成 PDF 报告（不阻塞响应）
	if h.reportHandler != nil {
		go h.reportHandler.GenerateReportByPaperID(b.PaperID)
	}
}

// ===================== POST /exam/api/mbti/score =====================
// 查看得分（报告页用，不改状态）
func (h *MbtiHandler) Score(c *gin.Context) {
	var b struct {
		PaperID string `json:"paperId"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}

	scores, mbtiType := h.calcMbtiScores(b.PaperID)

	response.Rest(c, gin.H{
		"type":   mbtiType,
		"scores": scores,
	})
}

// ===================== 计分逻辑 =====================
// E总分 = 1A+5A+9A+13A+17A+21A+25A+29A+33A+37A+41A+45A (题号%4==1 的 A分)
// I总分 = 1B+5B+9B+13B+17B+21B+25B+29B+33B+37B+41B+45B
// S总分 = 2A+6A+10A+14A+18A+22A+26A+30A+34A+38A+42A+46A (题号%4==2 的 A分)
// N总分 = 同上 B分
// T总分 = 3A+7A+11A+15A+19A+23A+27A+31A+35A+39A+43A+47A (题号%4==3 的 A分)
// F总分 = 同上 B分
// J总分 = 4A+8A+12A+16A+20A+24A+28A+32A+36A+40A+44A+48A (题号%4==0 的 A分)
// P总分 = 同上 B分
func (h *MbtiHandler) calcMbtiScores(paperID string) (map[string]int, string) {
	// 读取所有答题记录 + 题目编号
	type row struct {
		Content string `gorm:"column:content"`
		ScoreA  int    `gorm:"column:score_a"`
		ScoreB  int    `gorm:"column:score_b"`
	}
	var rows []row
	h.db.Table("el_mbti_answer AS ma").
		Joins("INNER JOIN el_qu AS q ON q.id = ma.qu_id COLLATE utf8mb4_general_ci").
		Where("ma.paper_id = ? AND ma.answered = 1", paperID).
		Select("q.content, ma.score_a, ma.score_b").
		Scan(&rows)

	// 解析题号 V1~V48 → 数字 1~48
	scores := map[string]int{"E": 0, "I": 0, "S": 0, "N": 0, "T": 0, "F": 0, "J": 0, "P": 0}
	for _, r := range rows {
		num := 0
		if strings.HasPrefix(r.Content, "V") {
			num, _ = strconv.Atoi(r.Content[1:])
		}
		if num < 1 || num > 48 {
			continue
		}
		mod := num % 4
		switch mod {
		case 1: // E-I 维度：题号 1,5,9,...,45
			scores["E"] += r.ScoreA
			scores["I"] += r.ScoreB
		case 2: // S-N 维度：题号 2,6,10,...,46
			scores["S"] += r.ScoreA
			scores["N"] += r.ScoreB
		case 3: // T-F 维度：题号 3,7,11,...,47
			scores["T"] += r.ScoreA
			scores["F"] += r.ScoreB
		case 0: // J-P 维度：题号 4,8,12,...,48
			scores["J"] += r.ScoreA
			scores["P"] += r.ScoreB
		}
	}

	// 确定 4 字母类型（同分选 I/N/F/P）
	mbtiType := ""
	if scores["E"] > scores["I"] {
		mbtiType += "E"
	} else {
		mbtiType += "I"
	}
	if scores["S"] > scores["N"] {
		mbtiType += "S"
	} else {
		mbtiType += "N"
	}
	if scores["T"] > scores["F"] {
		mbtiType += "T"
	} else {
		mbtiType += "F"
	}
	if scores["J"] > scores["P"] {
		mbtiType += "J"
	} else {
		mbtiType += "P"
	}

	return scores, mbtiType
}
