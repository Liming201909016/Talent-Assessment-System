package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
	"gorm.io/gorm"
)

// paperQuContent 试卷题目内容查询结果（共享 struct）
type paperQuContent struct {
	Content     string
	IsRight     int8
	Answered    int8
	ActualScore int
}

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
	if err := h.db.Save(&exam).Error; err != nil {
		slog.Error("pdf-team: save pdf_path failed", "error", err)
	}
	response.Rest(c, gin.H{"pdfPath": savePath})
}

// exportPRow 导出用人员行
type exportPRow struct {
	Name        string     `gorm:"column:name"`
	IDNumber    string     `gorm:"column:id_number"`
	Gender      *string    `gorm:"column:gender"`
	Age         *int       `gorm:"column:age"`
	Telephone   *string    `gorm:"column:telephone"`
	Affiliation *string    `gorm:"column:affiliation"`
	Depart      *string    `gorm:"column:depart"`
	Post        *string    `gorm:"column:post"`
	Degree      *string    `gorm:"column:degree"`
	Major       *string    `gorm:"column:major"`
	StuFlag     *int       `gorm:"column:stu_flag"`
	UserTime    *int       `gorm:"column:user_time"`
	PaperID     *string    `gorm:"column:paper_id"`
	EndTime     *time.Time `gorm:"column:end_time"`
	CreateTime  *time.Time `gorm:"column:create_time"`
	MbtiType    *string    `gorm:"column:mbti_type"`
	MbtiScores  *string    `gorm:"column:mbti_scores"`
}

// POST /exam/api/exam/exam/export-raw-data  {examId}
// 基于 Excel 模板导出全体测评人员原始答题数据 + 维度得分
func (h *ExamHandler) ExportRawData(c *gin.Context) {
	examID := c.PostForm("examId")
	if examID == "" {
		examID = c.Query("examId")
	}
	if examID == "" {
		var b struct {
			ExamID string `json:"examId"`
		}
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

	// 取人员列表
	var rows []exportPRow
	isOpen := exam.IsOpen == 1
	if isOpen {
		h.db.Table("el_candidate AS t").
			Joins("LEFT JOIN el_paper AS pa ON pa.id = t.paper_id").
			Select(`t.name, '' as id_number, t.gender, t.age, t.telephone, t.affiliation,
				'' as depart, t.post, t.degree, t.major, t.stu_flag,
				pa.user_time, t.paper_id, t.end_time, pa.create_time`).
			Where("t.exam_id = ? AND (t.del_flag IS NULL OR t.del_flag = 0)", examID).
			Order("pa.create_time DESC").
			Scan(&rows)
	} else {
		h.db.Table("el_tester AS t").
			Joins("LEFT JOIN el_paper AS pa ON pa.id = t.paper_id").
			Select(`t.name, t.id_number, t.gender, t.age, t.telephone, t.affiliation,
				t.depart, t.post, t.degree, t.major, t.stu_flag,
				pa.user_time, t.paper_id, t.end_time, pa.create_time,
				t.mbti_type, t.mbti_scores`).
			Where("t.exam_id = ? AND (t.del_flag IS NULL OR t.del_flag = 0)", examID).
			Order("pa.create_time DESC").
			Scan(&rows)
	}

	// 尝试基于模板导出
	templateFile := h.findExportTemplate(repoCode)
	if templateFile != "" {
		if isMbti {
			h.exportMbtiByTemplate(c, exam, repoCode, templateFile, rows)
		} else if isMng {
			h.exportMngByTemplate(c, exam, repoCode, templateFile, rows)
		} else {
			h.exportPsyByTemplate(c, exam, repoCode, templateFile, rows)
		}
		return
	}
	// 无模板时退化为原始导出
	h.exportFallback(c, exam, repoCode, isMng, isMbti, rows)
}

// findExportTemplate 查找导出模板文件
// 优先使用配置路径，退化到默认路径
func (h *ExamHandler) findExportTemplate(repoCode string) string {
	searchDirs := []string{}
	if h.cfg.Upload.ExportTemplates != "" {
		searchDirs = append(searchDirs, h.cfg.Upload.ExportTemplates)
	}
	searchDirs = append(searchDirs, "./configs/export-templates", "../configs/export-templates")

	// 先精确匹配 repoCode
	for _, dir := range searchDirs {
		pattern := filepath.Join(dir, repoCode+".*xlsx")
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			slog.Info("export: template found", "path", matches[0])
			return matches[0]
		}
	}
	// 退化到前缀匹配（如 00302 → 003）
	if len(repoCode) >= 3 {
		prefix := repoCode[:3]
		for _, dir := range searchDirs {
			pattern := filepath.Join(dir, prefix+"*.*xlsx")
			matches, err := filepath.Glob(pattern)
			if err == nil && len(matches) > 0 {
				slog.Info("export: template fallback", "path", matches[0], "prefix", prefix)
				return matches[0]
			}
		}
	}
	return ""
}

// exportMbtiByTemplate 使用 MBTI 模板填充数据
// 模板结构: Row0=标题(合并), Row1=列头, 数据从 Row2 开始
// 列: A=日期, B=时间, C=序号, D=姓名, E=身份证号, F=性别, G=年龄, H=手机号,
//
//	I=单位/学校, J=部门, K=岗位, L=学历, M=专业, N=是否学生, O=答题用时, P=答题状态,
//	Q=MBTI类型, R=外向E, S=内向I, T=感觉S, U=直觉N, V=理性T, W=感性F, X=判断J, Y=感知P
func (h *ExamHandler) exportMbtiByTemplate(c *gin.Context, exam model.Exam,
	repoCode, templateFile string, rows []exportPRow) {

	f, err := excelize.OpenFile(templateFile)
	if err != nil {
		slog.Error("export: open template failed", "template", templateFile, "error", err)
		response.RestErr(c, "打开模板失败")
		return
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	// 替换标题
	f.SetCellValue(sheet, "A1", exam.Title+"测验结果统计表")

	// 删除模板中的示例数据行（Row2+）
	existingRows, _ := f.GetRows(sheet)
	if len(existingRows) > 2 {
		for i := len(existingRows); i > 2; i-- {
			f.RemoveRow(sheet, i)
		}
	}

	// 对开放考试补充 MBTI 分数（el_candidate 表没有 mbti_type/mbti_scores 列）
	for i := range rows {
		if rows[i].MbtiType == nil || *rows[i].MbtiType == "" {
			if rows[i].PaperID != nil && *rows[i].PaperID != "" && rows[i].EndTime != nil {
				h.fillMbtiFromPaper(&rows[i])
			}
		}
	}

	dataStartRow := 3 // Excel row 3 (1-indexed), after title + header
	for i, r := range rows {
		row := dataStartRow + i

		gender := formatGender(r.Gender)
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
		userTime := 0
		if r.UserTime != nil {
			userTime = *r.UserTime
		}

		// A=日期, B=时间
		if r.CreateTime != nil {
			f.SetCellValue(sheet, cellName(1, row), r.CreateTime.Format("2006-01-02"))
			f.SetCellValue(sheet, cellName(2, row), r.CreateTime.Format("15:04:05"))
		}
		f.SetCellValue(sheet, cellName(3, row), i+1)                     // 序号
		f.SetCellValue(sheet, cellName(4, row), r.Name)                  // 姓名
		f.SetCellValue(sheet, cellName(5, row), r.IDNumber)              // 身份证号
		f.SetCellValue(sheet, cellName(6, row), gender)                  // 性别
		f.SetCellValue(sheet, cellName(7, row), derefInt(r.Age))         // 年龄
		f.SetCellValue(sheet, cellName(8, row), derefStr(r.Telephone))   // 手机号
		f.SetCellValue(sheet, cellName(9, row), derefStr(r.Affiliation)) // 单位/学校
		f.SetCellValue(sheet, cellName(10, row), derefStr(r.Depart))     // 部门
		f.SetCellValue(sheet, cellName(11, row), derefStr(r.Post))       // 岗位
		f.SetCellValue(sheet, cellName(12, row), derefStr(r.Degree))     // 学历
		f.SetCellValue(sheet, cellName(13, row), derefStr(r.Major))      // 专业
		f.SetCellValue(sheet, cellName(14, row), stu)                    // 是否学生
		f.SetCellValue(sheet, cellName(15, row), userTime)               // 答题用时
		f.SetCellValue(sheet, cellName(16, row), status)                 // 答题状态

		// MBTI 维度 (Q-Y, cols 17-25)
		if r.MbtiType != nil {
			f.SetCellValue(sheet, cellName(17, row), *r.MbtiType)
		}
		if r.MbtiScores != nil {
			var ms map[string]int
			if json.Unmarshal([]byte(*r.MbtiScores), &ms) == nil {
				dimKeys := []string{"E", "I", "S", "N", "T", "F", "J", "P"}
				for j, k := range dimKeys {
					f.SetCellValue(sheet, cellName(18+j, row), ms[k])
				}
			}
		}
	}

	fileName := exam.Title + "-原始数据.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+
		"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		slog.Error("export: write failed", "error", err)
	}
}

// psyScoreLevel 心理特质分数等级: ≥7 高分, 4.5~7 中分, <4.5 低分
func psyScoreLevel(score float64) string {
	if score >= 7 {
		return "高分"
	} else if score >= 4.5 {
		return "中分"
	}
	return "低分"
}

// psyHealthLevel 心理健康水平（简化版）
func psyHealthLevel(scores map[string]float64) string {
	sum := 0.0
	count8 := 0
	for _, v := range scores {
		sum += v
		if v > 8 {
			count8++
		}
	}
	avg := sum / float64(len(scores))
	if avg <= 7 && count8 <= 1 {
		return "心理状态良好"
	} else if avg <= 7 && count8 <= 3 {
		return "心理状况较好"
	} else if avg <= 7 && count8 <= 5 {
		return "心理状态尚可"
	} else if avg > 7 && count8 <= 6 {
		return "心理状态较差"
	}
	return "心理状态很差"
}

// exportPsyByTemplate 心理特质模板导出（001）
// 00101(职场版): C0日期 C1时间 C2姓名 C3身份证号 C4性别 C5年龄 C6手机号 C7单位 C8岗位 C9答题时间 C10答题数 C11健康水平 C12~C35=12维度(分+级)
// 00102(学生版): C0日期 C1时间 C2姓名 C3身份证号 C4性别 C5年龄 C6手机号 C7学校 C8专业 C9学历 C10答题时间 C11答题数 C12健康水平 C13~C36=12维度(分+级)
func (h *ExamHandler) exportPsyByTemplate(c *gin.Context, exam model.Exam,
	repoCode, templateFile string, rows []exportPRow) {

	f, err := excelize.OpenFile(templateFile)
	if err != nil {
		slog.Error("export: open template failed", "template", templateFile, "error", err)
		response.RestErr(c, "打开模板失败")
		return
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", exam.Title+"测验结果统计表")

	// 删除模板示例数据行（保留前3行: 标题+表头+维度头）
	existingRows, _ := f.GetRows(sheet)
	if len(existingRows) > 3 {
		for i := len(existingRows); i > 3; i-- {
			f.RemoveRow(sheet, i)
		}
	}

	// 根据 repoCode 确定是职场版还是学生版
	isStu := repoCode == "00102"
	dims := []string{"焦虑", "抑郁", "心理失衡", "敌意", "恐惧", "身体不适", "认知衰退", "情绪化", "挫折感", "自我否定", "怀疑感", "职业倦怠"}

	dataStartRow := 4 // Excel row 4 (1-indexed)
	for i, r := range rows {
		row := dataStartRow + i
		gender := formatGender(r.Gender)
		userTime := 0
		if r.UserTime != nil {
			userTime = *r.UserTime
		}

		// 日期 & 时间
		if r.CreateTime != nil {
			f.SetCellValue(sheet, cellName(1, row), r.CreateTime.Format("2006-01-02"))
			f.SetCellValue(sheet, cellName(2, row), r.CreateTime.Format("15:04:05"))
		}
		f.SetCellValue(sheet, cellName(3, row), r.Name)                // 姓名
		f.SetCellValue(sheet, cellName(4, row), r.IDNumber)            // 身份证号
		f.SetCellValue(sheet, cellName(5, row), gender)                // 性别
		f.SetCellValue(sheet, cellName(6, row), derefInt(r.Age))       // 年龄
		f.SetCellValue(sheet, cellName(7, row), derefStr(r.Telephone)) // 手机号

		var dimStartCol int
		if isStu {
			// 学生版: C7学校, C8专业, C9学历, C10答题时间, C11答题数, C12健康水平, C13+维度
			f.SetCellValue(sheet, cellName(8, row), derefStr(r.Affiliation)) // 学校
			f.SetCellValue(sheet, cellName(9, row), derefStr(r.Major))       // 专业
			f.SetCellValue(sheet, cellName(10, row), derefStr(r.Degree))     // 学历
			f.SetCellValue(sheet, cellName(11, row), fmt.Sprintf("%d 分钟", userTime))
			f.SetCellValue(sheet, cellName(12, row), 90) // 答题数（90题固定）
			// C13=健康水平, C14+维度
			dimStartCol = 14
		} else {
			// 职场版: C7单位, C8岗位, C9答题时间, C10答题数, C11健康水平, C12+维度
			f.SetCellValue(sheet, cellName(8, row), derefStr(r.Affiliation)) // 单位
			f.SetCellValue(sheet, cellName(9, row), derefStr(r.Post))        // 岗位
			f.SetCellValue(sheet, cellName(10, row), fmt.Sprintf("%d 分钟", userTime))
			f.SetCellValue(sheet, cellName(11, row), 90) // 答题数
			// C12=健康水平, C13+维度
			dimStartCol = 13
		}

		// 计算维度分
		if r.PaperID != nil && *r.PaperID != "" && r.EndTime != nil {
			qus, err := queryPaperQuContent(h.db, *r.PaperID)
			if err == nil {
				scores := standScore1(qus)
				// 健康水平
				healthCol := dimStartCol - 1
				f.SetCellValue(sheet, cellName(healthCol, row), psyHealthLevel(scores))
				// 12维度 × 2列(分数+等级)
				for j, d := range dims {
					v := scores[d]
					f.SetCellValue(sheet, cellName(dimStartCol+j*2, row), math.Round(v*100)/100)
					f.SetCellValue(sheet, cellName(dimStartCol+j*2+1, row), psyScoreLevel(v))
				}
			}
		}
	}

	fileName := exam.Title + "-原始数据.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+
		"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		slog.Error("export: write failed", "error", err)
	}
}

// mngScoreLevel 管理特质分数等级（5级）
func mngScoreLevel(score float64) string {
	if score >= 4.0 {
		return "高分"
	} else if score >= 3.0 {
		return "较高分"
	} else if score >= 2.0 {
		return "中等分"
	} else if score >= 1.0 {
		return "较低分"
	}
	return "低分"
}

// mngTotalLevel 管理特质总分等级
func mngTotalLevel(total float64) string {
	if total >= 48 {
		return "高分"
	} else if total >= 36 {
		return "较高分"
	} else if total >= 24 {
		return "中等分"
	} else if total >= 12 {
		return "较低分"
	}
	return "低分"
}

// mngDiagnosis 管理特质测评诊断
func mngDiagnosis(total float64) string {
	if total >= 48 {
		return "高潜管理者"
	} else if total >= 36 {
		return "较高潜管理者"
	} else if total >= 24 {
		return "中等管理者"
	} else if total >= 12 {
		return "低潜管理者"
	}
	return "非管理型"
}

// exportMngByTemplate 管理特质模板导出（002）
// 00201/00202: C0日期 C1时间 C2姓名 C3身份证号 C4性别 C5年龄 C6手机号 C7单位 C8部门
// C9岗位 C10职务 C11答题时间 C12答题数 C13得分等级 C14测评诊断 C15总分
// C16~C41=13维度(分+级)
func (h *ExamHandler) exportMngByTemplate(c *gin.Context, exam model.Exam,
	repoCode, templateFile string, rows []exportPRow) {

	f, err := excelize.OpenFile(templateFile)
	if err != nil {
		slog.Error("export: open template failed", "template", templateFile, "error", err)
		response.RestErr(c, "打开模板失败")
		return
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", exam.Title+"测验结果统计表")

	existingRows, _ := f.GetRows(sheet)
	if len(existingRows) > 3 {
		for i := len(existingRows); i > 3; i-- {
			f.RemoveRow(sheet, i)
		}
	}

	dims := []string{"社会性", "进取性", "领导性", "计划性", "人际敏感性", "自信心", "责任心", "学习力", "创新性", "情绪稳定性", "自律性", "决断性", "合作性"}

	dataStartRow := 4
	for i, r := range rows {
		row := dataStartRow + i
		gender := formatGender(r.Gender)
		userTime := 0
		if r.UserTime != nil {
			userTime = *r.UserTime
		}

		if r.CreateTime != nil {
			f.SetCellValue(sheet, cellName(1, row), r.CreateTime.Format("2006-01-02"))
			f.SetCellValue(sheet, cellName(2, row), r.CreateTime.Format("15:04:05"))
		}
		f.SetCellValue(sheet, cellName(3, row), r.Name)
		f.SetCellValue(sheet, cellName(4, row), r.IDNumber)
		f.SetCellValue(sheet, cellName(5, row), gender)
		f.SetCellValue(sheet, cellName(6, row), derefInt(r.Age))
		f.SetCellValue(sheet, cellName(7, row), derefStr(r.Telephone))
		f.SetCellValue(sheet, cellName(8, row), derefStr(r.Affiliation)) // 单位
		f.SetCellValue(sheet, cellName(9, row), derefStr(r.Depart))      // 部门
		f.SetCellValue(sheet, cellName(10, row), derefStr(r.Post))       // 岗位
		f.SetCellValue(sheet, cellName(11, row), "")                     // 职务（暂无此字段）
		f.SetCellValue(sheet, cellName(12, row), fmt.Sprintf("%d 分钟", userTime))
		f.SetCellValue(sheet, cellName(13, row), 140) // 答题数（140题固定）

		if r.PaperID != nil && *r.PaperID != "" && r.EndTime != nil {
			qus, err := queryPaperQuContent(h.db, *r.PaperID)
			if err == nil {
				scores := standScore2(qus)
				// 总分
				total := 0.0
				for _, d := range dims {
					total += scores[d]
				}
				total = math.Round(total*100) / 100

				f.SetCellValue(sheet, cellName(14, row), mngTotalLevel(total)) // 得分等级
				f.SetCellValue(sheet, cellName(15, row), mngDiagnosis(total))  // 测评诊断
				f.SetCellValue(sheet, cellName(16, row), total)                // 总分

				// 13维度 × 2列
				for j, d := range dims {
					v := scores[d]
					f.SetCellValue(sheet, cellName(17+j*2, row), math.Round(v*10000)/10000)
					f.SetCellValue(sheet, cellName(17+j*2+1, row), mngScoreLevel(v))
				}
			}
		}
	}

	fileName := exam.Title + "-原始数据.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+
		"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		slog.Error("export: write failed", "error", err)
	}
}

// fillMbtiFromPaper 从 el_mbti_answer 实时计算 MBTI 分数（用于 el_candidate 没有缓存分数的场景）
// 逻辑与 mbti.go calcMbtiScores 一致：按题号(V1~V48) mod 4 分配维度
func (h *ExamHandler) fillMbtiFromPaper(r *exportPRow) {
	if r.PaperID == nil {
		return
	}
	type answerRow struct {
		Content string `gorm:"column:content"` // V1~V48
		ScoreA  int    `gorm:"column:score_a"`
		ScoreB  int    `gorm:"column:score_b"`
	}
	var answers []answerRow
	h.db.Table("el_mbti_answer AS ma").
		Joins("INNER JOIN el_qu AS q ON q.id = ma.qu_id COLLATE utf8mb4_general_ci").
		Where("ma.paper_id = ? AND ma.answered = 1", *r.PaperID).
		Select("q.content, ma.score_a, ma.score_b").
		Scan(&answers)

	if len(answers) == 0 {
		return
	}

	scores := map[string]int{"E": 0, "I": 0, "S": 0, "N": 0, "T": 0, "F": 0, "J": 0, "P": 0}
	for _, a := range answers {
		num := 0
		if strings.HasPrefix(a.Content, "V") {
			fmt.Sscanf(a.Content[1:], "%d", &num)
		}
		if num < 1 || num > 48 {
			continue
		}
		mod := num % 4
		switch mod {
		case 1: // E-I
			scores["E"] += a.ScoreA
			scores["I"] += a.ScoreB
		case 2: // S-N
			scores["S"] += a.ScoreA
			scores["N"] += a.ScoreB
		case 3: // T-F
			scores["T"] += a.ScoreA
			scores["F"] += a.ScoreB
		case 0: // J-P
			scores["J"] += a.ScoreA
			scores["P"] += a.ScoreB
		}
	}

	pick := func(a, b string) string {
		if scores[a] >= scores[b] {
			return a
		}
		return b
	}
	mbtiType := pick("E", "I") + pick("S", "N") + pick("T", "F") + pick("J", "P")
	r.MbtiType = &mbtiType
	scoresJSON, _ := json.Marshal(scores)
	s := string(scoresJSON)
	r.MbtiScores = &s
}

// exportFallback 原始程序化导出（无模板）
func (h *ExamHandler) exportFallback(c *gin.Context, exam model.Exam,
	repoCode string, isMng, isMbti bool, rows []exportPRow) {

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
		gender := formatGender(r.Gender)
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
		userTime := 0
		if r.UserTime != nil {
			userTime = *r.UserTime
		}
		baseData := []any{
			i + 1, r.Name, r.IDNumber, gender, derefInt(r.Age), derefStr(r.Telephone),
			derefStr(r.Affiliation), derefStr(r.Depart), derefStr(r.Post),
			derefStr(r.Degree), derefStr(r.Major), stu, userTime, status,
		}
		for j, v := range baseData {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}

		dimVals := make([]any, len(dims))
		if r.PaperID != nil && *r.PaperID != "" && r.EndTime != nil {
			if isMbti {
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
				qus, err := queryPaperQuContent(h.db, *r.PaperID)
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

	for i := range allHeaders {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 14)
	}

	fileName := exam.Title + "-原始数据.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+
		"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		response.RestErr(c, err.Error())
	}
}

// cellName 简写: col(1-based), row(1-based) → "A1"
func cellName(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

func formatGender(g *string) string {
	if g == nil {
		return ""
	}
	switch *g {
	case "0":
		return "男"
	case "1":
		return "女"
	default:
		return *g
	}
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// capPageSize 限制分页大小，防止恶意大请求
func capPageSize(size int) int {
	if size <= 0 {
		return 10
	}
	if size > 200 {
		return 200
	}
	return size
}

// queryPaperQuContent 统一的试卷题目查询（消除 3 处重复）
// 被 tester_score.go, candidate.go, exam_pdf.go 共用
func queryPaperQuContent(db *gorm.DB, paperID string) ([]paperQuContent, error) {
	var rows []paperQuContent
	err := db.Table("el_paper_qu AS pq").
		Select("eq.content AS content, pq.is_right AS is_right, pq.answered AS answered, pq.actual_score AS actual_score").
		Joins("LEFT JOIN el_qu AS eq ON pq.qu_id = eq.id").
		Where("pq.paper_id = ?", paperID).
		Order("pq.sort ASC").
		Find(&rows).Error
	return rows, err
}
