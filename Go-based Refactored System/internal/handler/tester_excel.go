package handler

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ===== Tester Excel =====

// POST /exam/api/tester/importTemplate
// 对齐 Java importTemplate：下载含 标题/说明/表头 三行 + 性别/是否学生 下拉校验的 xlsx
func (h *TesterHandler) ImportTemplate(c *gin.Context) {
	headers := []string{"姓名", "手机号", "身份证号", "年龄", "性别", "密码", "单位/学校", "部门", "岗位", "学历", "专业", "职务", "是否学生"}
	notes := []string{"必填，如：张三", "必填，11位", "选填，18位", "选填，如：28", "必填，男/女", "选填，默认手机号后4位", "选填", "选填", "选填", "选填", "选填", "选填", "必填，是/否"}
	colWidths := []float64{12, 14, 20, 8, 8, 15, 16, 12, 12, 10, 12, 10, 10}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "测评人员导入模板"
	idx, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "微软雅黑", Size: 16, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "微软雅黑", Size: 11, Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    []excelize.Border{{Type: "left", Style: 1, Color: "000000"}, {Type: "right", Style: 1, Color: "000000"}, {Type: "top", Style: 1, Color: "000000"}, {Type: "bottom", Style: 1, Color: "000000"}},
	})
	noteStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "微软雅黑", Size: 9, Italic: true, Color: "808080"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	// Row 1 标题
	last := colName(len(headers))
	f.SetCellValue(sheet, "A1", "测评人员导入模板")
	f.MergeCell(sheet, "A1", last+"1")
	f.SetCellStyle(sheet, "A1", last+"1", titleStyle)
	f.SetRowHeight(sheet, 1, 28)

	// Row 2 说明
	for i, n := range notes {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, n)
	}
	f.SetCellStyle(sheet, "A2", last+"2", noteStyle)
	f.SetRowHeight(sheet, 2, 20)

	// Row 3 表头
	for i, n := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue(sheet, cell, n)
	}
	f.SetCellStyle(sheet, "A3", last+"3", headerStyle)
	f.SetRowHeight(sheet, 3, 22)

	// 列宽
	for i, w := range colWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	// 手机号(B列)和身份证号(C列)设为文本格式，防止Excel自动转科学计数法
	textStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 49}) // 49 = "@" = 文本
	f.SetColStyle(sheet, "B", textStyle) // 手机号
	f.SetColStyle(sheet, "C", textStyle) // 身份证号
	f.SetColStyle(sheet, "F", textStyle) // 密码

	// 数据校验：性别（第 5 列 E，模板列序：姓名→手机号→身份证号→年龄→性别）
	dv := excelize.NewDataValidation(true)
	dv.Sqref = "E4:E5000"
	_ = dv.SetDropList([]string{"男", "女"})
	dv.SetError(excelize.DataValidationErrorStyleStop, "输入错误", "性别只能填 男 或 女")
	f.AddDataValidation(sheet, dv)

	// 是否学生（第 13 列 M，因新增了"职务"列）
	dv2 := excelize.NewDataValidation(true)
	dv2.Sqref = "M4:M5000"
	_ = dv2.SetDropList([]string{"是", "否"})
	dv2.SetError(excelize.DataValidationErrorStyleStop, "输入错误", "是否学生只能填 是 或 否")
	f.AddDataValidation(sheet, dv2)

	// 冻结前 3 行
	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 3, TopLeftCell: "A4", ActivePane: "bottomLeft"})

	fileName := "测评人员导入模板.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		response.AjaxErr(c, err.Error())
	}
}

// POST /exam/api/tester/importData  (multipart file + changeExamId)
// 对齐 Java importData：从第 3 行起读取，按 id_number 判存/新建
func (h *TesterHandler) ImportData(c *gin.Context) {
	changeExamID := c.PostForm("changeExamId")
	if changeExamID == "" {
		response.AjaxErr(c, "请选择导入所属测评")
		return
	}
	// 校验封闭测评
	if msg := h.validateClosedExam(changeExamID); msg != "" {
		response.AjaxErr(c, msg)
		return
	}
	fh, err := c.FormFile("file")
	if err != nil || fh == nil {
		response.AjaxErr(c, "请上传导入文件")
		return
	}
	fp, err := fh.Open()
	if err != nil {
		response.AjaxErr(c, "读取文件失败")
		return
	}
	defer fp.Close()
	f, err := excelize.OpenReader(fp)
	if err != nil {
		response.AjaxErr(c, "解析 Excel 失败："+err.Error())
		return
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		response.AjaxErr(c, "Excel 无可读工作表")
		return
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	// Java：headerRow=2, dataStartRow=3（索引从 0）；Go 索引也是 0，第 4 行起是数据 → rows[3:]
	if len(rows) < 4 {
		response.AjaxErr(c, "未知错误，导入失败")
		return
	}

	insertN, updateN, skipN := 0, 0, 0
	now := time.Now()

	for _, row := range rows[3:] {
		get := func(i int) string {
			if i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}
		// 修复 Excel 数字列科学计数法问题
		fixNum := func(s string) string {
			if strings.Contains(s, "E+") || strings.Contains(s, "e+") {
				f, err := strconv.ParseFloat(s, 64)
				if err == nil {
					return strconv.FormatInt(int64(f), 10)
				}
			}
			// 去掉小数点（如 "13812345678.0" → "13812345678"）
			if idx := strings.Index(s, "."); idx > 0 {
				return s[:idx]
			}
			return s
		}
		name := get(0)
		tel := fixNum(get(1))
		idn := fixNum(get(2))
		if name == "" || tel == "" {
			skipN++
			continue
		}
		ageStr := get(3)
		genderText := get(4)
		pwd := get(5)
		affiliation := get(6)
		depart := get(7)
		post := get(8)
		degree := get(9)
		major := get(10)
		postTitle := get(11) // 职务列
		stuText := get(12)

		// 性别：男→0 女→1
		var gender *string
		switch genderText {
		case "男":
			s := "0"
			gender = &s
		case "女":
			s := "1"
			gender = &s
		}
		// 学生：是→1 否→0
		stuFlag := 0
		if stuText == "是" {
			stuFlag = 1
		}
		// 默认密码：手机号后 4 位（需求R01），如手机号为空则身份证后4位
		if pwd == "" {
			if len(tel) >= 4 {
				pwd = tel[len(tel)-4:]
			} else if len(idn) >= 4 {
				pwd = idn[len(idn)-4:]
			} else if idn != "" {
				pwd = idn
			} else {
				pwd = "0000"
			}
		}
		_ = postTitle // 职务：暂存但 el_tester 表无此列，后续可扩展
		var agePtr *int
		if ageStr != "" {
			if a, err := strconv.Atoi(ageStr); err == nil {
				agePtr = &a
			}
		}

		ptr := func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}

		err := h.db.Transaction(func(tx *gorm.DB) error {
			// upsert by (id_number 或 telephone, exam_id) — R01: 身份证号可为空，用手机号匹配
			identifier := idn
			identCol := "id_number"
			if identifier == "" {
				identifier = tel
				identCol = "telephone"
			}
			if identifier == "" {
				return nil // 无标识，跳过
			}
			var existing model.Tester
			err := tx.Table("el_tester").
				Where(identCol+" = ? AND exam_id = ? AND (del_flag IS NULL OR del_flag = '0')", identifier, changeExamID).
				First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				examIdStr := changeExamID
				tester := model.Tester{
					ID:          strconv.FormatInt(nextID(), 10),
					ExamID:      &examIdStr,
					IDNumber:    idn,
					Name:        name,
					Age:         agePtr,
					Gender:      gender,
					Password:    pwd,
					Telephone:   ptr(tel),
					Affiliation: ptr(affiliation),
					Depart:      ptr(depart),
					Post:        ptr(post),
					Degree:      ptr(degree),
					Major:       ptr(major),
					StuFlag:     &stuFlag,
					CreateTime:  &now,
					UpdateTime:  &now,
				}
				if e := tx.Create(&tester).Error; e != nil {
					return e
				}
				insertN++
			} else if err != nil {
				return err
			} else {
				updates := map[string]any{"name": name, "update_time": &now}
				if agePtr != nil { updates["age"] = agePtr }
				if gender != nil { updates["gender"] = gender }
				if pwd != "" { updates["password"] = pwd }
				if tel != "" { updates["telephone"] = tel }
				if affiliation != "" { updates["affiliation"] = affiliation }
				if depart != "" { updates["depart"] = depart }
				if post != "" { updates["post"] = post }
				if degree != "" { updates["degree"] = degree }
				if major != "" { updates["major"] = major }
				updates["stu_flag"] = stuFlag
				if e := tx.Table("el_tester").Where("id = ?", existing.ID).Updates(updates).Error; e != nil {
					return e
				}
				updateN++
			}
			return nil
		})
		if err != nil {
			skipN++
			continue
		}
	}

	response.AjaxOK(c, fmt.Sprintf("导入成功，新增 %d 条，更新 %d 条，跳过 %d 条", insertN, updateN, skipN))
}

// POST /exam/api/tester/export  (form or query: examId)
// 对齐 Java exportFile：输出含 title + 测评人员 为名的 xlsx
func (h *TesterHandler) Export(c *gin.Context) {
	examID := c.PostForm("examId")
	if examID == "" {
		examID = c.Query("examId")
	}
	if examID == "" {
		response.AjaxErr(c, "请选择要导出的测评")
		return
	}
	if msg := h.validateClosedExam(examID); msg != "" {
		response.AjaxErr(c, msg)
		return
	}

	type row struct {
		Title       string  `gorm:"column:title"`
		Name        string  `gorm:"column:name"`
		IDNumber    string  `gorm:"column:id_number"`
		Age         *int    `gorm:"column:age"`
		Gender      *string `gorm:"column:gender"`
		Telephone   *string `gorm:"column:telephone"`
		Affiliation *string `gorm:"column:affiliation"`
		Depart      *string `gorm:"column:depart"`
		Post        *string `gorm:"column:post"`
		Degree      *string `gorm:"column:degree"`
		Major       *string `gorm:"column:major"`
		StuFlag     *int    `gorm:"column:stu_flag"`
	}
	var rows []row
	h.db.Table("el_tester AS t").
		Joins("LEFT JOIN el_exam AS e ON e.id = t.exam_id").
		Select("e.title, t.name, t.id_number, t.age, t.gender, t.telephone, t.affiliation, t.depart, t.post, t.degree, t.major, t.stu_flag").
		Where("t.exam_id = ? AND (t.del_flag IS NULL OR t.del_flag = '0')", examID).
		Scan(&rows)
	if len(rows) == 0 {
		response.AjaxErr(c, "当前测评下没有可导出的测评人员")
		return
	}
	title := rows[0].Title
	if title == "" {
		title = "测评"
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := title + "-测评人员"
	// Excel sheet name 长度≤31，且不能含 :\/?*[]
	sheet = sanitizeSheetName(sheet)
	idx, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	headers := []string{"序号", "姓名", "手机号", "身份证号", "年龄", "性别", "单位/学校", "部门", "岗位", "学历", "专业", "是否学生"}
	for i, hh := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hh)
	}
	genderText := func(g *string) string {
		if g == nil {
			return ""
		}
		if *g == "0" {
			return "男"
		}
		if *g == "1" {
			return "女"
		}
		return *g
	}
	stuText := func(s *int) string {
		if s == nil {
			return ""
		}
		if *s == 1 {
			return "是"
		}
		return "否"
	}
	for i, r := range rows {
		ageVal := ""
		if r.Age != nil {
			ageVal = strconv.Itoa(*r.Age)
		}
		data := []any{
			i + 1, r.Name, derefStr(r.Telephone), r.IDNumber, ageVal, genderText(r.Gender),
			derefStr(r.Affiliation), derefStr(r.Depart),
			derefStr(r.Post), derefStr(r.Degree), derefStr(r.Major), stuText(r.StuFlag),
		}
		for j, v := range data {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheet, cell, v)
		}
	}

	fileName := title + "-测评人员.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		response.AjaxErr(c, err.Error())
	}
}

// validateClosedExam 对齐 Java：仅允许 is_open=封闭（Java CLOSED_EXAM_TYPE=2）
func (h *TesterHandler) validateClosedExam(examID string) string {
	var isOpen *int
	err := h.db.Table("el_exam").Where("id = ?", examID).Select("is_open").Row().Scan(&isOpen)
	if err != nil {
		return "所选测评不存在"
	}
	if isOpen == nil || *isOpen != 2 {
		return "测评者管理仅支持选择封闭测评"
	}
	return ""
}

// fixExcelNumber 将 Excel 科学计数法（如 "1.10101199001011E+17"）转为纯数字字符串
func fixExcelNumber(s string) string {
	if s == "" {
		return s
	}
	// 如果包含 E/e（科学计数法），转为整数字符串
	if strings.ContainsAny(s, "Ee") {
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return strconv.FormatInt(int64(f), 10)
		}
	}
	// 如果是浮点数（如 "13812345678.0"），去掉小数部分
	if strings.Contains(s, ".") {
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return strconv.FormatInt(int64(f), 10)
		}
	}
	return s
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func colName(n int) string {
	s, _ := excelize.ColumnNumberToName(n)
	return s
}

func sanitizeSheetName(s string) string {
	for _, ch := range []string{":", "\\", "/", "?", "*", "[", "]"} {
		s = strings.ReplaceAll(s, ch, "_")
	}
	if len([]rune(s)) > 31 {
		rs := []rune(s)
		s = string(rs[:31])
	}
	return s
}
