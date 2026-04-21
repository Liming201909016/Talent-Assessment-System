package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"github.com/xuri/excelize/v2"
)

// POST /exam/api/exam/exam/pdf-upload  {examId}
// 对齐 Java PdfUpload：按 exam.pdf_path 回传 PDF 二进制
func (h *ExamHandler) PdfUpload(c *gin.Context) {
	examID := c.PostForm("examId")
	if examID == "" {
		examID = c.Query("examId")
	}
	if examID == "" {
		response.RestErr(c, "examId 为空")
		return
	}
	var exam model.Exam
	if err := h.db.Where("id = ?", examID).First(&exam).Error; err != nil {
		response.RestErr(c, "考试不存在")
		return
	}
	if exam.PdfPath == "" {
		response.RestErr(c, "当前测评未上传 PDF")
		return
	}
	fp, err := os.Open(exam.PdfPath)
	if err != nil {
		response.RestErr(c, "文件"+exam.PdfPath+"不存在")
		return
	}
	defer fp.Close()
	name := filepath.Base(exam.PdfPath)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+name)
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	_, _ = io.Copy(c.Writer, fp)
}

// POST /exam/api/exam/exam/pdf-team  (multipart file[] + examId)
// 对齐 Java PdfTeam：上传一个或多个 PDF，合并为 {title}_{yyyyMMdd}{examId}.pdf 存 {profile-parent}/pdf/team/
func (h *ExamHandler) PdfTeam(c *gin.Context) {
	examID := c.PostForm("examId")
	if examID == "" {
		response.RestErr(c, "examId 为空")
		return
	}
	var exam model.Exam
	if err := h.db.Where("id = ?", examID).First(&exam).Error; err != nil {
		response.RestErr(c, "考试不存在")
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		response.RestErr(c, "读取上传文件失败")
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		response.RestErr(c, "请上传 PDF 文件")
		return
	}

	profile := h.cfg.Upload.Profile
	if profile == "" {
		profile = "./tmp/profile"
	}
	parent := filepath.Dir(profile)
	pdfDir := filepath.Join(parent, "pdf", "team")
	if err := os.MkdirAll(pdfDir, 0o755); err != nil {
		response.RestErr(c, "创建目录失败："+err.Error())
		return
	}
	dateStr := time.Now().Format("20060102")
	fileName := fmt.Sprintf("%s_%s%s.pdf", exam.Title, dateStr, examID)
	savePath := filepath.Join(pdfDir, fileName)

	out, err := os.Create(savePath)
	if err != nil {
		response.RestErr(c, "创建文件失败："+err.Error())
		return
	}
	defer out.Close()
	// 依次写入（Java 行为：循环覆盖；最后一次为准）
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			response.RestErr(c, err.Error())
			return
		}
		_, _ = out.Seek(0, io.SeekStart)
		_, err = io.Copy(out, src)
		src.Close()
		if err != nil {
			response.RestErr(c, err.Error())
			return
		}
	}

	exam.PdfPath = savePath
	_ = h.db.Save(&exam).Error
	response.Rest(c, gin.H{"pdfPath": savePath})
}

// POST /exam/api/exam/exam/export-raw-data  {examId}
// 对齐 Java exportRawData：导出全体测评人员原始答题数据 + 维度得分
func (h *ExamHandler) ExportRawData(c *gin.Context) {
	examID := c.PostForm("examId")
	if examID == "" {
		examID = c.Query("examId")
	}
	if examID == "" {
		// 也尝试从 JSON body 读取
		var b struct{ ExamID string `json:"examId"` }
		_ = c.ShouldBindJSON(&b)
		examID = b.ExamID
	}
	if examID == "" {
		response.RestErr(c, "examId 为空")
		return
	}
	var exam model.Exam
	if err := h.db.Where("id = ?", examID).First(&exam).Error; err != nil {
		response.RestErr(c, "考试不存在")
		return
	}

	// 判断题库类型
	var repoCode string
	h.db.Table("el_exam_repo er").
		Joins("INNER JOIN el_repo r ON r.id = er.repo_id").
		Where("er.exam_id = ?", examID).
		Limit(1).
		Pluck("r.code", &repoCode)
	isMng := strings.HasPrefix(repoCode, "002")
	isMbti := strings.HasPrefix(repoCode, "003")

	// 取人员列表（与 tester-list 相同 schema；Java 按 is_open 分支，新系统统一）
	type pRow struct {
		Name        string  `gorm:"column:name"`
		IDNumber    string  `gorm:"column:id_number"`
		Gender      *string `gorm:"column:gender"`
		Age         *int    `gorm:"column:age"`
		Telephone   *string `gorm:"column:telephone"`
		Affiliation *string `gorm:"column:affiliation"`
		Depart      *string `gorm:"column:depart"`
		Post        *string `gorm:"column:post"`
		Degree      *string `gorm:"column:degree"`
		Major       *string `gorm:"column:major"`
		StuFlag     *int    `gorm:"column:stu_flag"`
		UserTime    *int    `gorm:"column:user_time"`
		PaperID     *string `gorm:"column:paper_id"`
		EndTime     *time.Time `gorm:"column:end_time"`
		MbtiType    *string `gorm:"column:mbti_type"`
		MbtiScores  *string `gorm:"column:mbti_scores"`
	}
	var rows []pRow
	isOpen := exam.IsOpen == 1
	if isOpen {
		// 开放考试：从 el_candidate 取人员（无 id_number, depart, mbti 列）
		h.db.Table("el_candidate AS t").
			Joins("LEFT JOIN el_paper AS pa ON pa.id = t.paper_id").
			Select(`t.name, '' as id_number, t.gender, t.age, t.telephone, t.affiliation,
				'' as depart, t.post, t.degree, t.major, t.stu_flag,
				pa.user_time, t.paper_id, t.end_time`).
			Where("t.exam_id = ? AND (t.del_flag IS NULL OR t.del_flag = 0)", examID).
			Order("pa.create_time DESC").
			Scan(&rows)
	} else {
		// 封闭考试：从 el_tester 取人员
		h.db.Table("el_tester AS t").
			Joins("LEFT JOIN el_paper AS pa ON pa.id = t.paper_id").
			Select(`t.name, t.id_number, t.gender, t.age, t.telephone, t.affiliation,
				t.depart, t.post, t.degree, t.major, t.stu_flag,
				pa.user_time, t.paper_id, t.end_time, t.mbti_type, t.mbti_scores`).
			Where("t.exam_id = ? AND (t.del_flag IS NULL OR t.del_flag = 0)", examID).
			Order("pa.create_time DESC").
			Scan(&rows)
	}

	var dims []string
	if isMbti {
		dims = []string{"MBTI类型", "外向E", "内向I", "感觉S", "直觉N", "理性T", "感性F", "判断J", "感知P"}
	} else if isMng {
		dims = []string{"社会性", "进取性", "领导性", "计划性", "人际敏感性", "自信心", "责任心", "学习力", "创新性", "情绪稳定性", "自律性", "决断性", "合作性"}
	} else {
		dims = []string{"焦虑", "抑郁", "心理失衡", "敌意", "恐惧", "身体不适", "认知衰退", "情绪化", "挫折感", "自我否定", "怀疑感", "职业倦怠"}
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "原始数据"
	idx, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	baseHeaders := []string{"序号", "姓名", "身份证号", "性别", "年龄", "手机号", "单位/学校", "部门", "岗位", "学历", "专业", "是否学生", "答题用时(分钟)", "答题状态"}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "微软雅黑", Size: 11, Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4169E1"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	allHeaders := append([]string{}, baseHeaders...)
	allHeaders = append(allHeaders, dims...)
	for i, hh := range allHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hh)
	}
	lastCol, _ := excelize.ColumnNumberToName(len(allHeaders))
	f.SetCellStyle(sheet, "A1", lastCol+"1", headerStyle)

	for i, r := range rows {
		row := i + 2
		gender := ""
		if r.Gender != nil {
			switch *r.Gender {
			case "0":
				gender = "男"
			case "1":
				gender = "女"
			default:
				gender = *r.Gender
			}
		}
		stu := "否"
		if r.StuFlag != nil && *r.StuFlag == 1 {
			stu = "是"
		}
		status := "未测评"
		if r.PaperID != nil && *r.PaperID != "" {
			if r.EndTime != nil {
				status = "已完成"
			} else {
				status = "进行中"
			}
		}
		ageVal := 0
		if r.Age != nil {
			ageVal = *r.Age
		}
		userTime := 0
		if r.UserTime != nil {
			userTime = *r.UserTime
		}
		baseData := []any{
			i + 1, r.Name, r.IDNumber, gender, ageVal, derefStr(r.Telephone),
			derefStr(r.Affiliation), derefStr(r.Depart), derefStr(r.Post),
			derefStr(r.Degree), derefStr(r.Major), stu, userTime, status,
		}
		for j, v := range baseData {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}

		// 维度分
		dimVals := make([]any, len(dims))
		if r.PaperID != nil && *r.PaperID != "" && r.EndTime != nil {
			if isMbti {
				// MBTI: 从 el_tester.mbti_type / mbti_scores 读取
				if r.MbtiType != nil {
					dimVals[0] = *r.MbtiType
				}
				if r.MbtiScores != nil {
					var ms map[string]int
					if json.Unmarshal([]byte(*r.MbtiScores), &ms) == nil {
						dimKeys := []string{"E", "I", "S", "N", "T", "F", "J", "P"}
						for j, k := range dimKeys {
							dimVals[j+1] = ms[k]
						}
					}
				}
			} else {
				qus, err := h.queryPaperQuForScore(*r.PaperID)
				if err == nil {
					var scores map[string]float64
					if isMng {
						scores = standScore2(qus)
					} else {
						scores = standScore1(qus)
					}
					for j, d := range dims {
						v, ok := scores[d]
						if ok {
							dimVals[j] = math.Round(v*100) / 100
						} else {
							dimVals[j] = 0
						}
					}
				}
			}
		}
		for j, v := range dimVals {
			cell, _ := excelize.CoordinatesToCellName(len(baseHeaders)+j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	// 列宽最小 3000 ≈ 12 字符
	for i := range allHeaders {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 14)
	}

	fileName := exam.Title + "-原始数据.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		response.RestErr(c, err.Error())
	}
}

// queryPaperQuForScore 与 TesterHandler.queryPaperQuWithContent 相同语义，放 ExamHandler 避免交叉依赖
func (h *ExamHandler) queryPaperQuForScore(paperID string) ([]paperQuContent, error) {
	var rows []paperQuContent
	err := h.db.Table("el_paper_qu AS pq").
		Select("eq.content AS content, pq.is_right AS is_right, pq.answered AS answered, pq.actual_score AS actual_score").
		Joins("LEFT JOIN el_qu AS eq ON pq.qu_id = eq.id").
		Where("pq.paper_id = ?", paperID).
		Order("pq.sort ASC").
		Find(&rows).Error
	return rows, err
}
