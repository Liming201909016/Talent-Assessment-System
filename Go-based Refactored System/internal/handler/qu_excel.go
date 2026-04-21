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

// Qu Excel 列顺序（对齐 Java QuExportDTO @ExcelField sort）：
// 1=题目序号 2=题目类型 3=题目内容 4=整体解析 5=题目图片 6=题目视频
// 7=所属题库 8=是否正确项 9=选项内容 10=选项解析 11=选项图片 12=题目标题
var quExcelHeaders = []string{
	"题目序号", "题目类型", "题目内容", "整体解析", "题目图片", "题目视频",
	"所属题库", "是否正确项", "选项内容", "选项解析", "选项图片", "题目标题",
}

// POST /exam/api/qu/qu/import/template
// 返回 Java 同结构的 4 行示范模板
func (h *QuHandler) ImportTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "试题数据"
	idx, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	// Row1 表头
	for i, hh := range quExcelHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hh)
	}
	// Row2 说明
	explain := []string{
		"正式导入请删除此说明行：数字，相同的数字表示同一题的序列",
		"只能填写1、2、3、4；1=单选、2=多选、3=判断、4=主观",
		"问题内容",
		"整个问题的解析",
		"题目图片，完整URL，多个用逗号隔开，限10个",
		"题目视频，完整URL，只限一个",
		"已存在题库的ID，多个用逗号隔开",
		"只能填写0或1；0=否，1=是",
		"候选答案1",
		"这个项是正确的",
		"答案图片，完整URL，只限一个",
		"题目标题（选填）",
	}
	for i, v := range explain {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, v)
	}
	// Row3-Row5 示范数据（多选题）
	rows := [][]any{
		{"1", "2", "找出以下可以被2整除的数（多选）", "最基本的数学题", "", "", "", "1", "数字：2", "2÷2=1，对", "", "示范题"},
		{"1", "", "", "", "", "", "", "0", "数字：3", "3÷2=1.5，错", "", ""},
		{"1", "", "", "", "", "", "", "1", "数字：6", "6÷2=3，对", "", ""},
	}
	for ri, row := range rows {
		for ci, v := range row {
			cell, _ := excelize.CoordinatesToCellName(ci+1, 3+ri)
			f.SetCellValue(sheet, cell, v)
		}
	}

	fileName := "试题导入模板.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		response.RestErr(c, err.Error())
	}
}

// POST /exam/api/qu/qu/import-excel  (multipart file)
// 对齐 Java：sheet 标题作为新题库 title，不存在则创建；行 1 为说明，从行 2 起读数据
// 同 no 的多行合并为同一题的候选答案
func (h *QuHandler) ImportExcel(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil || fh == nil {
		response.RestErr(c, "请上传导入文件")
		return
	}
	fp, err := fh.Open()
	if err != nil {
		response.RestErr(c, "读取文件失败")
		return
	}
	defer fp.Close()
	f, err := excelize.OpenReader(fp)
	if err != nil {
		response.RestErr(c, "解析 Excel 失败："+err.Error())
		return
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		response.RestErr(c, "Excel 无可读工作表")
		return
	}
	sheetName := sheets[0]
	allRows, err := f.GetRows(sheetName)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	// Java: ImportExcel(file, 1, 0) → header 为第 1 行（索引 0），数据从第 2 行开始
	if len(allRows) < 2 {
		response.RestErr(c, "Excel 空表")
		return
	}

	// 题库 title = sheetName（与 Java 行为一致：getTitle 返回 sheet 名）
	title := sheetName

	// 分组数据
	type optRow struct {
		aIsRight string
		aContent string
		aAnalysis string
		aImage   string
	}
	type quBlock struct {
		no       string
		quType   int
		content  string
		analysis string
		qImage   string
		qVideo   string
		quTitle  string
		repoIDs  []string
		options  []optRow
	}
	blocks := map[string]*quBlock{}
	order := []string{}

	for _, row := range allRows[1:] {
		get := func(i int) string {
			if i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}
		no := get(0)
		if no == "" {
			continue
		}
		// 跳过模板示范/注释行：非数字 no 跳过
		if _, err := strconv.Atoi(no); err != nil {
			continue
		}
		block, ok := blocks[no]
		if !ok {
			qt := 1
			if n, err := strconv.Atoi(get(1)); err == nil {
				qt = n
			}
			repoIDs := []string{}
			for _, rid := range strings.Split(get(6), ",") {
				rid = strings.TrimSpace(rid)
				if rid != "" {
					repoIDs = append(repoIDs, rid)
				}
			}
			block = &quBlock{
				no:       no,
				quType:   qt,
				content:  get(2),
				analysis: get(3),
				qImage:   get(4),
				qVideo:   get(5),
				quTitle:  get(11),
				repoIDs:  repoIDs,
			}
			blocks[no] = block
			order = append(order, no)
		}
		block.options = append(block.options, optRow{
			aIsRight:  get(7),
			aContent:  get(8),
			aAnalysis: get(9),
			aImage:    get(10),
		})
	}

	if len(order) == 0 {
		response.RestErr(c, "您导入的数据似乎是一个空表格！")
		return
	}

	// 确保 sheet 标题对应的 repo 存在
	now := time.Now()
	var repo model.Repo
	err = h.db.Where("title = ?", title).First(&repo).Error
	if err == gorm.ErrRecordNotFound {
		repo = model.Repo{
			ID:         strconv.FormatInt(nextID(), 10),
			Code:       "",
			Title:      title,
			CreateTime: &now,
			UpdateTime: &now,
		}
		if e := h.db.Create(&repo).Error; e != nil {
			response.RestErr(c, e.Error())
			return
		}
	} else if err != nil {
		response.RestErr(c, err.Error())
		return
	}

	okN := 0
	err = h.db.Transaction(func(tx *gorm.DB) error {
		for _, no := range order {
			blk := blocks[no]
			if blk.content == "" || len(blk.options) == 0 {
				continue
			}
			quID := strconv.FormatInt(nextID(), 10)
			if err := tx.Create(&model.Qu{
				ID:         quID,
				QuType:     blk.quType,
				Image:      blk.qImage,
				Content:    blk.content,
				CreateTime: &now,
				UpdateTime: &now,
				Analysis:   blk.analysis,
			}).Error; err != nil {
				return err
			}
			// 关联当前 sheet repo
			if err := tx.Create(&model.QuRepo{
				ID:     strconv.FormatInt(nextID()+1, 10),
				QuID:   quID,
				RepoID: repo.ID,
				QuType: blk.quType,
			}).Error; err != nil {
				return err
			}
			// 关联 Excel 中指定的额外 repo
			for _, rid := range blk.repoIDs {
				if rid == repo.ID {
					continue
				}
				_ = tx.Create(&model.QuRepo{
					ID:     strconv.FormatInt(nextID()+int64(len(rid)), 10),
					QuID:   quID,
					RepoID: rid,
					QuType: blk.quType,
				}).Error
			}
			// 答案项
			for idx, opt := range blk.options {
				isRight := int8(0)
				if opt.aIsRight == "1" {
					isRight = 1
				}
				if err := tx.Create(&model.QuAnswer{
					ID:       strconv.FormatInt(nextID()+int64(idx+100), 10),
					QuID:     quID,
					IsRight:  isRight,
					Image:    opt.aImage,
					Content:  opt.aContent,
					Analysis: opt.aAnalysis,
				}).Error; err != nil {
					return err
				}
			}
			okN++
		}
		return nil
	})
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	// 更新 repo 统计（radio_count/multi_count/judge_count）
	_ = refreshRepoStat(h.db, repo.ID)
	response.Rest(c, gin.H{"message": fmt.Sprintf("导入 %d 题", okN)})
}

// POST /exam/api/qu/qu/export  (JSON body：title/quType 过滤，可缺省)
// 按 no 一题一块展开，首行含题目信息，其余行仅答案行
func (h *QuHandler) Export(c *gin.Context) {
	var req struct {
		Title     string      `json:"title"`
		QuTypeRaw interface{} `json:"quType"`
		RepoID    string      `json:"repoId"`
	}
	_ = c.ShouldBindJSON(&req)
	quType := toIntPtr(req.QuTypeRaw)

	q := h.db.Table("el_qu AS q").Select("q.id, q.qu_type, q.content, q.analysis, q.image")
	if req.Title != "" {
		q = q.Where("q.content like ?", "%"+req.Title+"%")
	}
	if quType != nil {
		q = q.Where("q.qu_type = ?", *quType)
	}
	if req.RepoID != "" {
		q = q.Joins("INNER JOIN el_qu_repo qr ON qr.qu_id = q.id").
			Where("qr.repo_id = ?", req.RepoID)
	}
	type quRow struct {
		ID       string `gorm:"column:id"`
		QuType   int    `gorm:"column:qu_type"`
		Content  string `gorm:"column:content"`
		Analysis string `gorm:"column:analysis"`
		Image    string `gorm:"column:image"`
	}
	var qus []quRow
	q.Scan(&qus)

	f := excelize.NewFile()
	defer f.Close()
	sheet := "试题"
	idx, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	for i, hh := range quExcelHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hh)
	}

	rowIdx := 2
	for i, qu := range qus {
		no := strconv.Itoa(i + 1)

		// 查该题所属题库 id 列表
		var repoIDs []string
		h.db.Table("el_qu_repo").Where("qu_id = ?", qu.ID).Pluck("repo_id", &repoIDs)
		var answers []model.QuAnswer
		h.db.Where("qu_id = ?", qu.ID).Find(&answers)
		if len(answers) == 0 {
			// 仍然写一行题目
			writeRow(f, sheet, rowIdx, []any{no, qu.QuType, qu.Content, qu.Analysis, qu.Image, "", strings.Join(repoIDs, ","), "", "", "", "", ""})
			rowIdx++
			continue
		}
		for j, a := range answers {
			if j == 0 {
				writeRow(f, sheet, rowIdx, []any{
					no, qu.QuType, qu.Content, qu.Analysis, qu.Image, "",
					strings.Join(repoIDs, ","),
					int(a.IsRight), a.Content, a.Analysis, a.Image, "",
				})
			} else {
				writeRow(f, sheet, rowIdx, []any{
					no, "", "", "", "", "",
					"",
					int(a.IsRight), a.Content, a.Analysis, a.Image, "",
				})
			}
			rowIdx++
		}
	}

	fileName := "导出的试题-" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		response.RestErr(c, err.Error())
	}
}

func writeRow(f *excelize.File, sheet string, row int, vals []any) {
	for i, v := range vals {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		f.SetCellValue(sheet, cell, v)
	}
}
