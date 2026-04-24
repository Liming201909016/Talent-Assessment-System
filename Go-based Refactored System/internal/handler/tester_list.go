package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// testerListRow 对齐 Java PaperResultDTO 的 tester-list 输出：
// 字段尽量对齐前端使用到的列。
type testerListRow struct {
	ID          string     `gorm:"column:id"          json:"id"`
	ExamID      string     `gorm:"column:exam_id"     json:"examId"`
	PaperID     *string    `gorm:"column:paper_id"    json:"paperId"`
	RepoCode    *string    `gorm:"column:repo_code"   json:"repoCode"`
	Name        string     `gorm:"column:name"        json:"name"`
	Age         *int       `gorm:"column:age"         json:"age"`
	Telephone   *string    `gorm:"column:telephone"   json:"telephone"`
	Affiliation *string    `gorm:"column:affiliation" json:"affiliation"`
	Gender      *string    `gorm:"column:gender"      json:"gender"`
	Post        *string    `gorm:"column:post"        json:"post"`
	Degree      *string    `gorm:"column:degree"      json:"degree"`
	Major       *string    `gorm:"column:major"       json:"major"`
	StuFlag     *int       `gorm:"column:stu_flag"    json:"stuFlag"`
	Password    *string    `gorm:"column:password"    json:"password"`
	PdfPath     *string    `gorm:"column:pdf_path"    json:"pdfPath"`
	PdfFlag     *int       `gorm:"column:pdf_flag"    json:"pdfFlag"`
	DelFlag     *int       `gorm:"column:del_flag"    json:"delFlag"`
	EndTime     *time.Time `gorm:"column:end_time"    json:"endTime"`
	CreateTime  *time.Time `gorm:"column:create_time" json:"createTime"`
	UserTime    *int       `gorm:"column:user_time"   json:"userTime"`
	AnswerNum   int        `gorm:"-"                  json:"answerNum"`
}

// GET /exam/api/tester/tester-list  (TableDataInfo)
// 对齐 Java CandidateServiceImpl.testerInfoList / TesterController.testerList：
// Java 直接查 el_candidate 表（即使是封闭考试场景也是如此），Go 保持一致。
//   SELECT ca.*, o.code repoCode, pa.create_time, pa.user_time
//   FROM el_candidate ca LEFT JOIN ...
//   WHERE ca.del_flag='0' [and filters]
//   ORDER BY pa.create_time DESC
func (h *TesterHandler) TesterList(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize = capPageSize(pageSize)
	name := c.Query("name")
	status := c.Query("status")
	telephone := c.Query("telephone")
	examID := c.Query("examId")
	examStatus := c.Query("examStatus") // 答题状态: 0=未测评 1=进行中 2=已完成

	where := func(q *gorm.DB) *gorm.DB {
		q = q.Where("(ca.del_flag = '0' OR ca.del_flag IS NULL)")
		if name != "" {
			q = q.Where("ca.name like ?", "%"+name+"%")
		}
		if status != "" {
			q = q.Where("ca.status = ?", status)
		}
		if telephone != "" {
			q = q.Where("ca.telephone = ?", telephone)
		}
		if examID != "" {
			q = q.Where("ca.exam_id = ?", examID)
		}
		switch examStatus {
		case "0": // 未测评
			q = q.Where("ca.paper_id IS NULL")
		case "1": // 进行中
			q = q.Where("ca.paper_id IS NOT NULL AND ca.end_time IS NULL")
		case "2": // 已完成
			q = q.Where("ca.end_time IS NOT NULL")
		}
		return q
	}

	var total int64
	where(h.db.Table("el_candidate AS ca")).Count(&total)

	var rows []testerListRow
	where(h.db.Table("el_candidate AS ca").
		Joins("LEFT JOIN el_paper AS pa ON pa.id = ca.paper_id").
		Joins("LEFT JOIN el_exam_repo AS er ON er.exam_id = ca.exam_id").
		Joins("LEFT JOIN el_repo AS o ON o.id = er.repo_id")).
		Select(`ca.id, ca.exam_id, ca.paper_id, o.code AS repo_code,
		ca.name, ca.age, ca.telephone, ca.affiliation, ca.gender, ca.post,
		ca.degree, ca.major, ca.stu_flag, ca.password, ca.pdf_path, ca.pdf_flag, ca.del_flag,
		ca.end_time, pa.create_time, pa.user_time`).
		Order("pa.create_time DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows)

	// 填充 answerNum：对每行若 paper_id 非空，统计已作答数
	for i := range rows {
		if rows[i].PaperID != nil && *rows[i].PaperID != "" {
			var n int64
			// MBTI（repoCode 003）答题记录在 el_mbti_answer
			if rows[i].RepoCode != nil && strings.HasPrefix(*rows[i].RepoCode, "003") {
				h.db.Table("el_mbti_answer").
					Where("paper_id = ? AND answered = 1", *rows[i].PaperID).
					Count(&n)
			} else {
				h.db.Table("el_paper_qu").
					Where("paper_id = ? AND answered = 1", *rows[i].PaperID).
					Count(&n)
			}
			rows[i].AnswerNum = int(n)
		}
	}

	response.Table(c, rows, total)
}
