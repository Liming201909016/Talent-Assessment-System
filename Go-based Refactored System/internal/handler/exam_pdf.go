package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

	baseDir := h.cfg.Upload.Path
	if baseDir == "" {
		baseDir = "./tmp"
	}
	pdfDir := filepath.Join(baseDir, "pdf", "team")
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

// POST /exam/api/exam/exam/export-raw-answers {examId}
// 导出考生逐题答题原始记录（按考生→题号排序）
// 列：考生ID、姓名、手机号、身份证号、题号、题目内容、考生答案、是否答对、得分
//
// 安全：
//   - 仅管理员/拥有 exam:list 权限的用户可导出
//   - 手机号、身份证号自动脱敏（仅显示后 4 位 + ****）
//   - 操作记录到 sys_oper_log
func (h *ExamHandler) ExportRawAnswers(c *gin.Context) {
	// === 权限校验 (#6) ===
	luVal, ok := c.Get("loginUser")
	if !ok {
		response.AjaxUnauthorized(c, "")
		return
	}
	lu, _ := luVal.(*model.LoginUser)
	if lu == nil {
		response.AjaxUnauthorized(c, "")
		return
	}
	// 超级管理员或具有相关权限
	isAdmin := lu.UserID == 1
	hasPerm := isAdmin
	if !isAdmin {
		for _, p := range lu.Permissions {
			if p == "*:*:*" || p == "exam:list" || p == "exam:export" {
				hasPerm = true
				break
			}
		}
	}
	if !hasPerm {
		response.Ajax(c, 403, "无权导出原始答题记录", nil)
		return
	}

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
	h.exportRawAnswersWide(c, lu, isAdmin, exam)
}

// exportRawAnswersWide 按客户模板（260429）输出宽格式：
// A=考生ID, B=姓名, C=手机号, D..= 1题得分..N题得分, 后续 = 维度得分
// 题数 N 按试卷实际题数（按 sort 升序）；维度按 repoCode：
//   - 001 → 12 维（standScore1）
//   - 002 → 13 维（standScore2）
//   - 003 (MBTI) → 4 维（E/I/S/N/T/F/J/P 四对，取 aggregateMbtiScores 各 8 个 score）
//   - 其他 → 不输出维度列
func (h *ExamHandler) exportRawAnswersWide(c *gin.Context, lu *model.LoginUser, isAdmin bool, exam model.Exam) {
	examID := exam.ID

	// 1) 题库 code + id（取首个，与已有逻辑一致）
	var repoCode string
	var repoID string
	h.db.Table("el_exam_repo er").
		Joins("INNER JOIN el_repo r ON r.id = er.repo_id").
		Where("er.exam_id = ?", examID).
		Limit(1).
		Select("r.code, r.id").
		Row().Scan(&repoCode, &repoID)
	isMbti := strings.HasPrefix(repoCode, "003")

	// 2) 考生列表
	type personRow struct {
		ID        string  `gorm:"column:id"`
		Name      string  `gorm:"column:name"`
		Telephone *string `gorm:"column:telephone"`
		PaperID   *string `gorm:"column:paper_id"`
	}
	var persons []personRow
	if exam.IsOpen == 1 {
		h.db.Table("el_candidate").
			Select("id, name, telephone, paper_id").
			Where("exam_id = ? AND (del_flag IS NULL OR del_flag = 0 OR del_flag = '0')", examID).
			Order("name ASC, id ASC").
			Scan(&persons)
	} else {
		h.db.Table("el_tester").
			Select("id, name, telephone, paper_id").
			Where("exam_id = ? AND (del_flag IS NULL OR del_flag = '0')", examID).
			Order("name ASC, id ASC").
			Scan(&persons)
	}

	// 3) 收集所有 paperId，统计题数（取最大 sort 作为列数）
	paperIDs := make([]string, 0, len(persons))
	for _, p := range persons {
		if p.PaperID != nil && *p.PaperID != "" {
			paperIDs = append(paperIDs, *p.PaperID)
		}
	}
	quCount := 0
	if len(paperIDs) > 0 && repoID != "" {
		// 题号 = el_qu_repo.sort（题库内题目编号，1-based，与题库管理界面"题目编号"列一致）
		var maxSort int
		h.db.Table("el_qu_repo").
			Where("repo_id = ?", repoID).
			Select("COALESCE(MAX(sort), 0)").
			Row().Scan(&maxSort)
		quCount = maxSort
	}
	// MBTI: 兜底从 el_mbti_answer 算（题号从 V1..VN 解析）
	if isMbti && quCount == 0 && len(paperIDs) > 0 {
		type mr struct {
			Content string `gorm:"column:content"`
		}
		var ms []mr
		h.db.Table("el_mbti_answer ma").
			Joins("INNER JOIN el_qu eq ON eq.id COLLATE utf8mb4_general_ci = ma.qu_id").
			Where("ma.paper_id IN ?", paperIDs).
			Select("DISTINCT eq.content AS content").
			Scan(&ms)
		for _, m := range ms {
			if n := parseMbtiQuNum(m.Content); n > quCount {
				quCount = n
			}
		}
	}

	// 4) 维度名（决定后续列）
	var dimNames []string
	switch {
	case strings.HasPrefix(repoCode, "001"):
		dimNames = []string{"焦虑", "抑郁", "心理失衡", "敌意", "恐惧", "身体不适", "认知衰退", "情绪化", "挫折感", "自我否定", "怀疑感", "职业倦怠"}
	case strings.HasPrefix(repoCode, "002"):
		dimNames = []string{"社会性", "进取性", "领导性", "计划性", "人际敏感性", "自信心", "责任心", "学习力", "创新性", "情绪稳定性", "自律性", "决断性", "合作性"}
	case isMbti:
		dimNames = []string{"E", "I", "S", "N", "T", "F", "J", "P"}
	}

	// 5) 生成 Excel
	f := excelize.NewFile()
	defer f.Close()
	sheet := "原始答题"
	idx, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	// 表头
	headers := make([]string, 0, 3+quCount+len(dimNames))
	headers = append(headers, "考生ID", "姓名", "手机号")
	for i := 1; i <= quCount; i++ {
		headers = append(headers, strconv.Itoa(i))
	}
	headers = append(headers, dimNames...)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "微软雅黑", Size: 11, Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4169E1"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	for i, hd := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hd)
	}
	if len(headers) > 0 {
		lastCol, _ := excelize.ColumnNumberToName(len(headers))
		f.SetCellStyle(sheet, "A1", lastCol+"1", headerStyle)
	}
	// 列宽
	f.SetColWidth(sheet, "A", "A", 14)
	f.SetColWidth(sheet, "B", "B", 10)
	f.SetColWidth(sheet, "C", "C", 14)
	if quCount > 0 {
		startCol, _ := excelize.ColumnNumberToName(4)
		endCol, _ := excelize.ColumnNumberToName(3 + quCount)
		f.SetColWidth(sheet, startCol, endCol, 5)
	}
	if len(dimNames) > 0 {
		startCol, _ := excelize.ColumnNumberToName(4 + quCount)
		endCol, _ := excelize.ColumnNumberToName(3 + quCount + len(dimNames))
		f.SetColWidth(sheet, startCol, endCol, 9)
	}

	// 6) 数据行
	for i, p := range persons {
		row := i + 2
		tel := derefStr(p.Telephone)
		if !isAdmin {
			tel = maskPhone(tel)
		}
		f.SetCellValue(sheet, mustCell(1, row), p.ID)
		f.SetCellValue(sheet, mustCell(2, row), p.Name)
		f.SetCellValue(sheet, mustCell(3, row), tel)

		if p.PaperID == nil || *p.PaperID == "" {
			continue
		}

		// 6.1 题目得分（按题库编号 qu_repo.sort 列位）
		// 不同题库取分字段不同：
		//   001 (心理特质): is_right 存"考生选择 0/1"，actual_score 始终=1（仅表示题目分值），需导出 is_right
		//   002 (管理特质): actual_score 存"考生选项分 1~5"
		//   其他: 默认 actual_score
		type pqRow struct {
			QuRepoSort  int  `gorm:"column:qu_repo_sort"`
			ActualScore int  `gorm:"column:actual_score"`
			IsRight     int8 `gorm:"column:is_right"`
		}
		var pqs []pqRow
		if repoID != "" {
			h.db.Table("el_paper_qu pq").
				Joins("INNER JOIN el_qu_repo qr ON qr.qu_id = pq.qu_id AND qr.repo_id = ?", repoID).
				Where("pq.paper_id = ?", *p.PaperID).
				Select("qr.sort AS qu_repo_sort, pq.actual_score, pq.is_right").
				Scan(&pqs)
		}
		usePsy := strings.HasPrefix(repoCode, "001")
		for _, pq := range pqs {
			// qu_repo.sort 1-based → 列号 = 3 + sort
			if pq.QuRepoSort < 1 || pq.QuRepoSort > quCount {
				continue
			}
			col := 3 + pq.QuRepoSort
			var v int
			if usePsy {
				v = int(pq.IsRight)
			} else {
				v = pq.ActualScore
			}
			f.SetCellValue(sheet, mustCell(col, row), v)
		}

		// 6.2 MBTI 题目得分（覆盖 el_mbti_answer，记 score_a）
		if isMbti {
			type maRow struct {
				QuID    string `gorm:"column:qu_id"`
				Content string `gorm:"column:content"`
				ScoreA  int    `gorm:"column:score_a"`
			}
			var mas []maRow
			h.db.Table("el_mbti_answer ma").
				Joins("INNER JOIN el_qu eq ON eq.id COLLATE utf8mb4_general_ci = ma.qu_id").
				Where("ma.paper_id = ?", *p.PaperID).
				Select("ma.qu_id, eq.content, ma.score_a").
				Scan(&mas)
			for _, m := range mas {
				n := parseMbtiQuNum(m.Content)
				if n < 1 || n > quCount {
					continue
				}
				f.SetCellValue(sheet, mustCell(3+n, row), m.ScoreA)
			}
		}

		// 6.3 维度得分
		if len(dimNames) > 0 {
			dimStartCol := 4 + quCount
			switch {
			case strings.HasPrefix(repoCode, "001"):
				qus, err := queryPaperQuContent(h.db, *p.PaperID)
				if err == nil {
					sc := standScore1(qus)
					for j, d := range dimNames {
						v := sc[d]
						f.SetCellValue(sheet, mustCell(dimStartCol+j, row), math.Round(v*100)/100)
					}
				}
			case strings.HasPrefix(repoCode, "002"):
				qus, err := queryPaperQuContent(h.db, *p.PaperID)
				if err == nil {
					sc := standScore2(qus)
					for j, d := range dimNames {
						v := sc[d]
						f.SetCellValue(sheet, mustCell(dimStartCol+j, row), math.Round(v*10000)/10000)
					}
				}
			case isMbti:
				type maRow2 struct {
					QuID    string `gorm:"column:qu_id"`
					Content string `gorm:"column:content"`
					ScoreA  int    `gorm:"column:score_a"`
					ScoreB  int    `gorm:"column:score_b"`
				}
				var mas []maRow2
				h.db.Table("el_mbti_answer ma").
					Joins("INNER JOIN el_qu eq ON eq.id COLLATE utf8mb4_general_ci = ma.qu_id").
					Where("ma.paper_id = ?", *p.PaperID).
					Select("ma.qu_id, eq.content, ma.score_a, ma.score_b").
					Scan(&mas)
				rows := make([]mbtiAnswerRow, 0, len(mas))
				for _, m := range mas {
					rows = append(rows, mbtiAnswerRow{Content: m.Content, ScoreA: m.ScoreA, ScoreB: m.ScoreB})
				}
				dims, _, _ := aggregateMbtiScores(rows)
				for j, d := range dimNames {
					f.SetCellValue(sheet, mustCell(dimStartCol+j, row), dims[d])
				}
			}
		}
	}

	// 7) 输出
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		slog.Error("export-raw-answers: build xlsx failed", "error", err)
		response.RestErr(c, "生成 Excel 失败："+err.Error())
		_ = h.recordOperLog(c, lu, exam, len(persons), 1, err.Error())
		return
	}
	fileName := exam.Title + "-原始答题记录.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+
		"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	if _, err := c.Writer.Write(buf.Bytes()); err != nil {
		slog.Error("export-raw-answers: write response failed", "error", err)
	}
	_ = h.recordOperLog(c, lu, exam, len(persons), 0, "")
}

// mustCell 包装 CoordinatesToCellName，错误退化为 A1（防御）
func mustCell(col, row int) string {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return "A1"
	}
	return cell
}

// _legacyExportRawAnswers 旧版（长格式）已弃用，保留下方代码不再使用。
func (h *ExamHandler) _legacyExportRawAnswers(c *gin.Context, lu *model.LoginUser, isAdmin bool, exam model.Exam) {
	examID := exam.ID
	_ = examID
	// 一次性 JOIN 查询所有考生的所有题目和答案
	type answerRow struct {
		PersonID    string  `gorm:"column:person_id"`
		Name        string  `gorm:"column:name"`
		Telephone   *string `gorm:"column:telephone"`
		IDNumber    *string `gorm:"column:id_number"`
		PaperID     string  `gorm:"column:paper_id"`
		Sort        int     `gorm:"column:sort"`
		QuContent   string  `gorm:"column:qu_content"`
		QuType      int     `gorm:"column:qu_type"`
		Answered    int8    `gorm:"column:answered"`
		Answer      string  `gorm:"column:answer"`
		IsRight     int8    `gorm:"column:is_right"`
		ActualScore int     `gorm:"column:actual_score"`
		Score       int     `gorm:"column:score"`
	}

	var rows []answerRow
	isOpen := exam.IsOpen == 1

	// 检测是否为 MBTI 测评
	var repoCode string
	h.db.Table("el_exam_repo er").
		Joins("INNER JOIN el_repo r ON r.id = er.repo_id").
		Where("er.exam_id = ?", examID).
		Limit(1).
		Pluck("r.code", &repoCode)
	isMbti := strings.HasPrefix(repoCode, "003")

	// del_flag 兼容子句（int 0、'0'、NULL 三种都视为未删除）(#9)
	delFlagOK := "(c.del_flag IS NULL OR c.del_flag = 0 OR c.del_flag = '0')"
	delFlagOKT := "(t.del_flag IS NULL OR t.del_flag = 0 OR t.del_flag = '0')"

	// 1. 普通题目（el_paper_qu）— 所有测评类型都会有
	if isOpen {
		h.db.Table("el_candidate AS c").
			Select(`c.id AS person_id, c.name, c.telephone, '' AS id_number, c.paper_id,
				pq.sort, eq.content AS qu_content, pq.qu_type, pq.answered,
				pq.answer, pq.is_right, pq.actual_score, pq.score`).
			Joins("INNER JOIN el_paper_qu AS pq ON pq.paper_id = c.paper_id").
			Joins("INNER JOIN el_qu AS eq ON eq.id = pq.qu_id").
			Where("c.exam_id = ? AND c.paper_id IS NOT NULL AND c.paper_id != '' AND "+delFlagOK, examID).
			Order("c.name ASC, c.id ASC, pq.sort ASC").
			Scan(&rows)
	} else {
		h.db.Table("el_tester AS t").
			Select(`t.id AS person_id, t.name, t.telephone, t.id_number, t.paper_id,
				pq.sort, eq.content AS qu_content, pq.qu_type, pq.answered,
				pq.answer, pq.is_right, pq.actual_score, pq.score`).
			Joins("INNER JOIN el_paper_qu AS pq ON pq.paper_id = t.paper_id").
			Joins("INNER JOIN el_qu AS eq ON eq.id = pq.qu_id").
			Where("t.exam_id = ? AND t.paper_id IS NOT NULL AND t.paper_id != '' AND "+delFlagOKT, examID).
			Order("t.name ASC, t.id ASC, pq.sort ASC").
			Scan(&rows)
	}

	// 2. MBTI 增量数据（el_mbti_answer）— 仅 003 题库
	if isMbti {
		type mbtiRow struct {
			PersonID  string  `gorm:"column:person_id"`
			Name      string  `gorm:"column:name"`
			Telephone *string `gorm:"column:telephone"`
			IDNumber  *string `gorm:"column:id_number"`
			PaperID   string  `gorm:"column:paper_id"`
			QuID      string  `gorm:"column:qu_id"`
			QuContent string  `gorm:"column:qu_content"`
			ScoreA    int     `gorm:"column:score_a"`
			ScoreB    int     `gorm:"column:score_b"`
			Answered  int8    `gorm:"column:answered"`
		}
		var mbtiRows []mbtiRow
		if isOpen {
			h.db.Table("el_candidate AS c").
				Select(`c.id AS person_id, c.name, c.telephone, '' AS id_number, c.paper_id,
					ma.qu_id, eq.content AS qu_content, ma.score_a, ma.score_b, ma.answered`).
				Joins("INNER JOIN el_mbti_answer AS ma ON ma.paper_id = c.paper_id").
				Joins("INNER JOIN el_qu AS eq ON eq.id COLLATE utf8mb4_general_ci = ma.qu_id").
				Where("c.exam_id = ? AND c.paper_id IS NOT NULL AND c.paper_id != '' AND "+delFlagOK, examID).
				Order("c.name ASC, c.id ASC, eq.content ASC").
				Scan(&mbtiRows)
		} else {
			h.db.Table("el_tester AS t").
				Select(`t.id AS person_id, t.name, t.telephone, t.id_number, t.paper_id,
					ma.qu_id, eq.content AS qu_content, ma.score_a, ma.score_b, ma.answered`).
				Joins("INNER JOIN el_mbti_answer AS ma ON ma.paper_id = t.paper_id").
				Joins("INNER JOIN el_qu AS eq ON eq.id COLLATE utf8mb4_general_ci = ma.qu_id").
				Where("t.exam_id = ? AND t.paper_id IS NOT NULL AND t.paper_id != '' AND "+delFlagOKT, examID).
				Order("t.name ASC, t.id ASC, eq.content ASC").
				Scan(&mbtiRows)
		}

		// 批量加载题目选项
		quIDSet := map[string]bool{}
		for _, r := range mbtiRows {
			quIDSet[r.QuID] = true
		}
		quIDs := make([]string, 0, len(quIDSet))
		for id := range quIDSet {
			quIDs = append(quIDs, id)
		}
		type quAnswerRow struct {
			QuID    string `gorm:"column:qu_id"`
			Content string `gorm:"column:content"`
		}
		var quAnswers []quAnswerRow
		if len(quIDs) > 0 {
			h.db.Table("el_qu_answer").
				Where("qu_id IN ?", quIDs).
				Select("qu_id, content").
				Scan(&quAnswers)
		}
		answerByQu := map[string][]string{}
		for _, qa := range quAnswers {
			answerByQu[qa.QuID] = append(answerByQu[qa.QuID], qa.Content)
		}

		for _, m := range mbtiRows {
			opts := answerByQu[m.QuID]
			optA, optB := "", ""
			if len(opts) >= 1 {
				optA = opts[0]
			}
			if len(opts) >= 2 {
				optB = opts[1]
			}
			// 解析题号 V1-V48 → 1-48 (#3 防御：非 V 开头或解析失败保持 0)
			sortNum := parseMbtiQuNum(m.QuContent)
			selected := ""
			if m.Answered == 1 {
				if m.ScoreA > m.ScoreB {
					selected = fmt.Sprintf("A (%d/%d)", m.ScoreA, m.ScoreB)
				} else if m.ScoreB > m.ScoreA {
					selected = fmt.Sprintf("B (%d/%d)", m.ScoreA, m.ScoreB)
				} else {
					selected = fmt.Sprintf("(平分 %d/%d)", m.ScoreA, m.ScoreB)
				}
			} else {
				selected = "(未作答)"
			}
			rows = append(rows, answerRow{
				PersonID:    m.PersonID,
				Name:        m.Name,
				Telephone:   m.Telephone,
				IDNumber:    m.IDNumber,
				PaperID:     m.PaperID,
				Sort:        sortNum,
				QuContent:   fmt.Sprintf("%s | A: %s | B: %s", m.QuContent, optA, optB),
				QuType:      5,
				Answered:    m.Answered,
				Answer:      selected,
				IsRight:     0,
				ActualScore: m.ScoreA + m.ScoreB,
				Score:       0,
			})
		}
	}

	// 统一排序：按姓名 → 考生ID → 题号（保证 MBTI 行也参与排序）
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Name != rows[j].Name {
			return rows[i].Name < rows[j].Name
		}
		if rows[i].PersonID != rows[j].PersonID {
			return rows[i].PersonID < rows[j].PersonID
		}
		return rows[i].Sort < rows[j].Sort
	})

	// 生成 Excel
	f := excelize.NewFile()
	defer f.Close()
	sheet := "原始答题记录"
	idx, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	headers := []string{"考生ID", "姓名", "手机号", "身份证号", "试卷ID", "题号", "题型", "题目内容", "考生答案", "是否答对", "实得分", "题目满分"}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "微软雅黑", Size: 11, Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4169E1"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	lastCol, _ := excelize.ColumnNumberToName(len(headers))
	f.SetCellStyle(sheet, "A1", lastCol+"1", headerStyle)

	// 列宽
	widths := []float64{20, 12, 14, 20, 20, 6, 6, 50, 30, 8, 8, 8}
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	// 数据行（手机号/身份证脱敏 #1，但超管全量显示）
	for i, r := range rows {
		row := i + 2
		quTypeText := quTypeName(r.QuType)
		isRightText := "—"
		if r.Answered == 1 {
			if r.IsRight == 1 {
				isRightText = "是"
			} else {
				isRightText = "否"
			}
		}
		answer := r.Answer
		if r.Answered != 1 {
			answer = "(未作答)"
		}
		tel := derefStr(r.Telephone)
		idn := derefStr(r.IDNumber)
		if !isAdmin {
			tel = maskPhone(tel)
			idn = maskIDNumber(idn)
		}
		values := []any{
			r.PersonID,
			r.Name,
			tel,
			idn,
			r.PaperID,
			r.Sort,
			quTypeText,
			r.QuContent,
			answer,
			isRightText,
			r.ActualScore,
			r.Score,
		}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	// 先写入 buffer 再返回（#5 避免错误时返回半截文件）
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		slog.Error("export-raw-answers: build xlsx failed", "error", err)
		response.RestErr(c, "生成 Excel 失败："+err.Error())
		_ = h.recordOperLog(c, lu, exam, len(rows), 1, err.Error())
		return
	}

	fileName := exam.Title + "-原始答题记录.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+
		"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	if _, err := c.Writer.Write(buf.Bytes()); err != nil {
		slog.Error("export-raw-answers: write response failed", "error", err)
	}

	// 操作日志（#7）
	_ = h.recordOperLog(c, lu, exam, len(rows), 0, "")
}

// parseMbtiQuNum 从题目内容（如 V1, V48）解析数字编号；解析失败返回 0
func parseMbtiQuNum(content string) int {
	if !strings.HasPrefix(content, "V") || len(content) < 2 {
		return 0
	}
	n := 0
	if _, err := fmt.Sscanf(content[1:], "%d", &n); err != nil {
		return 0
	}
	if n < 0 || n > 9999 {
		return 0
	}
	return n
}

// maskPhone 手机号脱敏：138****5678
func maskPhone(s string) string {
	if len(s) < 7 {
		return s
	}
	r := []rune(s)
	if len(r) < 11 {
		return s
	}
	return string(r[:3]) + "****" + string(r[len(r)-4:])
}

// maskIDNumber 身份证号脱敏：保留前 6 位（行政区划）和后 4 位
func maskIDNumber(s string) string {
	r := []rune(s)
	if len(r) < 12 {
		return s
	}
	return string(r[:6]) + strings.Repeat("*", len(r)-10) + string(r[len(r)-4:])
}

// recordOperLog 写入 sys_oper_log（#7）
func (h *ExamHandler) recordOperLog(c *gin.Context, lu *model.LoginUser, exam model.Exam, rowCount int, status int, errMsg string) error {
	now := time.Now()
	param := fmt.Sprintf(`{"examId":"%s","examTitle":"%s","rowCount":%d}`,
		exam.ID, exam.Title, rowCount)
	if len(param) > 1900 {
		param = param[:1900]
	}
	if len(errMsg) > 1900 {
		errMsg = errMsg[:1900]
	}
	username := ""
	if lu.User != nil {
		username = lu.User.UserName
	}
	row := map[string]interface{}{
		"title":          "测评导出原始答题",
		"business_type":  3, // 3=查询/导出
		"method":         "ExamHandler.ExportRawAnswers",
		"request_method": "POST",
		"operator_type":  1, // 1=后台用户
		"oper_name":      username,
		"oper_url":       c.Request.URL.Path,
		"oper_ip":        c.ClientIP(),
		"oper_param":     param,
		"json_result":    "",
		"status":         status,
		"error_msg":      errMsg,
		"oper_time":      &now,
	}
	return h.db.Table("sys_oper_log").Create(&row).Error
}

// quTypeName 题型枚举映射
func quTypeName(t int) string {
	switch t {
	case 1:
		return "单选"
	case 2:
		return "多选"
	case 3:
		return "判断"
	case 4:
		return "简答"
	case 5:
		return "MBTI"
	default:
		return fmt.Sprintf("类型%d", t)
	}
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
//
// content 字段语义：以"题库编号"(el_qu_repo.sort) 为准生成 V{n}，
// 而非直接取 el_qu.content（题目文本编码）。
// 原因：心理量表公式中的 V1..V90 指题库编号；目前 4 个题库 V编码 == 题库编号 (1:1)，
// 但代码语义上以 qu_repo.sort 为准更严谨。
func queryPaperQuContent(db *gorm.DB, paperID string) ([]paperQuContent, error) {
	type row struct {
		Sort        int  `gorm:"column:sort"`
		IsRight     int8 `gorm:"column:is_right"`
		Answered    int8 `gorm:"column:answered"`
		ActualScore int  `gorm:"column:actual_score"`
	}
	var raws []row
	err := db.Table("el_paper_qu AS pq").
		Select("qr.sort AS sort, pq.is_right AS is_right, pq.answered AS answered, pq.actual_score AS actual_score").
		Joins("INNER JOIN el_paper p ON p.id = pq.paper_id").
		Joins("INNER JOIN el_exam_repo er ON er.exam_id = p.exam_id").
		Joins("INNER JOIN el_qu_repo qr ON qr.qu_id = pq.qu_id AND qr.repo_id = er.repo_id").
		Where("pq.paper_id = ?", paperID).
		Order("qr.sort ASC").
		Find(&raws).Error
	if err != nil {
		return nil, err
	}
	out := make([]paperQuContent, 0, len(raws))
	for _, r := range raws {
		out = append(out, paperQuContent{
			Content:     fmt.Sprintf("V%d", r.Sort),
			IsRight:     r.IsRight,
			Answered:    r.Answered,
			ActualScore: r.ActualScore,
		})
	}
	return out, nil
}

// GET /exam/api/exam/exam/answer-detail?paperId=xxx
// 返回某个考生的逐题答题详情：题号 + 选项 + 考生选了哪个
//
// 权限：管理员或拥有 exam:list / exam:export
// 数据源：el_paper_qu_answer (含 checked 标记)
func (h *ExamHandler) AnswerDetail(c *gin.Context) {
	luVal, _ := c.Get("loginUser")
	lu, _ := luVal.(*model.LoginUser)
	if lu == nil {
		response.AjaxUnauthorized(c, "")
		return
	}
	isAdmin := lu.UserID == 1
	if !isAdmin {
		ok := false
		for _, p := range lu.Permissions {
			if p == "*:*:*" || p == "exam:list" || p == "exam:export" {
				ok = true
				break
			}
		}
		if !ok {
			response.Ajax(c, 403, "无权查看", nil)
			return
		}
	}

	paperID := c.Query("paperId")
	if paperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}

	// 1) paper + 考生 + 题库
	type paperInfo struct {
		ExamID    string  `gorm:"column:exam_id"`
		Title     string  `gorm:"column:title"`
		RepoID    string  `gorm:"column:repo_id"`
		RepoCode  string  `gorm:"column:repo_code"`
		RepoTitle string  `gorm:"column:repo_title"`
		IsOpen    int8    `gorm:"column:is_open"`
		Name      string  `gorm:"column:name"`
		Telephone *string `gorm:"column:telephone"`
	}
	var pi paperInfo
	// candidate 兜底 tester
	row := h.db.Table("el_paper p").
		Joins("LEFT JOIN el_exam e ON e.id = p.exam_id").
		Joins("LEFT JOIN el_exam_repo er ON er.exam_id = e.id").
		Joins("LEFT JOIN el_repo r ON r.id = er.repo_id").
		Joins("LEFT JOIN el_candidate c ON c.paper_id = p.id").
		Where("p.id = ?", paperID).
		Limit(1).
		Select(`p.exam_id, e.title, r.id AS repo_id, r.code AS repo_code, r.title AS repo_title,
			COALESCE(e.is_open, 0) AS is_open, c.name, c.telephone`).
		Take(&pi)
	if row.Error != nil || pi.Name == "" {
		// 兜底 tester
		_ = h.db.Table("el_paper p").
			Joins("LEFT JOIN el_exam e ON e.id = p.exam_id").
			Joins("LEFT JOIN el_exam_repo er ON er.exam_id = e.id").
			Joins("LEFT JOIN el_repo r ON r.id = er.repo_id").
			Joins("LEFT JOIN el_tester t ON t.paper_id = p.id").
			Where("p.id = ?", paperID).
			Limit(1).
			Select(`p.exam_id, e.title, r.id AS repo_id, r.code AS repo_code, r.title AS repo_title,
				COALESCE(e.is_open, 0) AS is_open, t.name, t.telephone`).
			Take(&pi).Error
	}
	if pi.ExamID == "" {
		response.RestErr(c, "试卷不存在")
		return
	}

	// 2) 题目 + 题号 (qu_repo.sort)
	type qrow struct {
		QuID    string `gorm:"column:qu_id"`
		QuSort  int    `gorm:"column:qu_sort"`
		Content string `gorm:"column:content"`
		Title   string `gorm:"column:title"`
	}
	var qrows []qrow
	q := h.db.Table("el_paper_qu pq").
		Joins("INNER JOIN el_qu eq ON eq.id = pq.qu_id").
		Where("pq.paper_id = ?", paperID).
		Select("pq.qu_id, eq.content, COALESCE(eq.title,'') AS title")
	if pi.RepoID != "" {
		q = q.Joins("INNER JOIN el_qu_repo qr ON qr.qu_id = pq.qu_id AND qr.repo_id = ?", pi.RepoID).
			Select("pq.qu_id, qr.sort AS qu_sort, eq.content, COALESCE(eq.title,'') AS title").
			Order("qr.sort ASC")
	} else {
		q = q.Order("pq.sort ASC")
	}
	q.Scan(&qrows)

	// 3) 选项 + checked （一次查全，按 qu_id 分组）
	quIDs := make([]string, 0, len(qrows))
	for _, r := range qrows {
		quIDs = append(quIDs, r.QuID)
	}
	type pqaRow struct {
		QuID     string `gorm:"column:qu_id"`
		AnswerID string `gorm:"column:answer_id"`
		Abc      string `gorm:"column:abc"`
		IsRight  int8   `gorm:"column:is_right"`
		Checked  int8   `gorm:"column:checked"`
		Sort     int    `gorm:"column:sort"`
	}
	var pqas []pqaRow
	if len(quIDs) > 0 {
		h.db.Table("el_paper_qu_answer").
			Where("paper_id = ? AND qu_id IN ?", paperID, quIDs).
			Select("qu_id, answer_id, abc, is_right, checked, sort").
			Order("sort ASC").
			Scan(&pqas)
	}
	// 选项内容（从 el_qu_answer 拿）
	ansIDs := make([]string, 0, len(pqas))
	for _, a := range pqas {
		ansIDs = append(ansIDs, a.AnswerID)
	}
	type ansRow struct {
		ID      string `gorm:"column:id"`
		Content string `gorm:"column:content"`
	}
	var ans []ansRow
	if len(ansIDs) > 0 {
		h.db.Table("el_qu_answer").Where("id IN ?", ansIDs).Select("id, content").Scan(&ans)
	}
	ansMap := make(map[string]string, len(ans))
	for _, a := range ans {
		ansMap[a.ID] = a.Content
	}

	// 4) 组装响应
	type optionDTO struct {
		Abc     string `json:"abc"`
		Content string `json:"content"`
		IsRight int8   `json:"isRight"`
		Checked int8   `json:"checked"`
		Score   int    `json:"score"` // MBTI: 考生分配给该选项的分数 (0~10)
	}
	type qDTO struct {
		Sort    int         `json:"sort"`
		VCode   string      `json:"vCode"`
		Title   string      `json:"title"`
		Options []optionDTO `json:"options"`
	}
	optsByQu := make(map[string][]optionDTO)
	for _, a := range pqas {
		optsByQu[a.QuID] = append(optsByQu[a.QuID], optionDTO{
			Abc:     a.Abc,
			Content: ansMap[a.AnswerID],
			IsRight: a.IsRight,
			Checked: a.Checked,
		})
	}

	// 4.1) MBTI 题库：覆盖 score (从 el_mbti_answer 取 score_a/score_b)
	isMbti := strings.HasPrefix(pi.RepoCode, "003")
	if isMbti && len(quIDs) > 0 {
		type maRow struct {
			QuID     string `gorm:"column:qu_id"`
			ScoreA   int    `gorm:"column:score_a"`
			ScoreB   int    `gorm:"column:score_b"`
			Answered int8   `gorm:"column:answered"`
		}
		var mas []maRow
		h.db.Table("el_mbti_answer").
			Where("paper_id = ? AND qu_id IN ?", paperID, quIDs).
			Select("qu_id, score_a, score_b, answered").
			Scan(&mas)
		mbtiByQu := make(map[string]maRow, len(mas))
		for _, m := range mas {
			mbtiByQu[m.QuID] = m
		}
		// 重新组装：按 mbti score 给两个选项填 score 和 checked（score 大的视为已选）
		for _, r := range qrows {
			m, ok := mbtiByQu[r.QuID]
			opts := optsByQu[r.QuID]
			if !ok || len(opts) < 2 {
				continue
			}
			opts[0].Score = m.ScoreA
			opts[1].Score = m.ScoreB
			if m.Answered == 1 {
				opts[0].Checked = 0
				opts[1].Checked = 0
				if m.ScoreA > m.ScoreB {
					opts[0].Checked = 1
				} else if m.ScoreB > m.ScoreA {
					opts[1].Checked = 1
				}
				// 平分：两个都不标 checked
			}
			optsByQu[r.QuID] = opts
		}
	}

	questions := make([]qDTO, 0, len(qrows))
	for _, r := range qrows {
		questions = append(questions, qDTO{
			Sort:    r.QuSort,
			VCode:   r.Content,
			Title:   r.Title,
			Options: optsByQu[r.QuID],
		})
	}

	response.AjaxOK(c, gin.H{
		"paperId":   paperID,
		"examId":    pi.ExamID,
		"examTitle": pi.Title,
		"repoCode":  pi.RepoCode,
		"repoTitle": pi.RepoTitle,
		"isMbti":    isMbti,
		"name":      pi.Name,
		"telephone": derefStr(pi.Telephone),
		"questions": questions,
	})
}
