package handler

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

const (
	competencyImportMaxFileSize  = int64(10 << 20)
	competencyImportMaxUnzipSize = int64(64 << 20)
	competencyImportMaxRows      = 10000
)

type CompetencyImportHandler struct {
	db *gorm.DB
}

func NewCompetencyImportHandler(db *gorm.DB) *CompetencyImportHandler {
	return &CompetencyImportHandler{db: db}
}

// ImportTemplate downloads the fixed ten-column, one-question-per-row template.
func (h *CompetencyImportHandler) ImportTemplate(c *gin.Context) {
	file := excelize.NewFile()
	defer file.Close()
	sheet := "胜任力题目"
	index, err := file.NewSheet(sheet)
	if err != nil {
		response.RestErr(c, "生成导入模板失败")
		return
	}
	file.DeleteSheet("Sheet1")
	file.SetActiveSheet(index)

	for column, header := range service.CompetencyImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(column+1, 1)
		file.SetCellValue(sheet, cell, header)
	}
	examples := [][]string{
		{
			"1", "逻辑思维", "维度题", "A1-01-EXAMPLE-F", "9001", "面对复杂事项时，我会先拆解再逐项处理。",
			"分解能力", "正向", "启用", "示例数据-正式导入前请修改或删除",
		},
		{
			"1", "逻辑思维", "维度题", "A1-01-EXAMPLE-R", "9002", "我通常凭感觉就判断问题所在。",
			"结构化验证", "反向", "启用", "示例数据-正式导入前请修改或删除",
		},
		{
			"1", "逻辑思维", "效度题", "P1-VAL-EXAMPLE", "9001", "我在工作中从未出现过任何失误。",
			"极端美化", "正向", "启用", "示例数据-正式导入前请修改或删除",
		},
	}
	for column := range service.CompetencyImportHeaders {
		explainCell, _ := excelize.CoordinatesToCellName(column+1, 2)
		file.SetCellValue(sheet, explainCell, service.CompetencyImportInstructions[column])
		for row, example := range examples {
			exampleCell, _ := excelize.CoordinatesToCellName(column+1, row+3)
			file.SetCellValue(sheet, exampleCell, example[column])
		}
	}
	file.SetColWidth(sheet, "A", "E", 18)
	file.SetColWidth(sheet, "F", "G", 42)
	file.SetColWidth(sheet, "H", "J", 24)

	fileName := "competency-question-import-template.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := file.Write(c.Writer); err != nil {
		response.RestErr(c, "生成导入模板失败")
	}
}

// ImportPreview validates all rows and returns their normalized values and file
// digest. It performs no database writes.
func (h *CompetencyImportHandler) ImportPreview(c *gin.Context) {
	data, rows, digest, err := readCompetencyImportFile(c)
	_ = data
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	validation, err := h.validateImportRows(rows)
	if err != nil {
		response.RestErr(c, "校验导入数据失败")
		return
	}
	response.Rest(c, gin.H{
		"sha256":       digest,
		"successCount": len(validation.ValidRows),
		"errorCount":   len(validation.Errors),
		"successRows":  validation.ValidRows,
		"errorRows":    validation.Errors,
	})
}

// Import re-uploads and re-validates the file, then writes the complete batch in
// one transaction. Any row or database error leaves the question table unchanged.
func (h *CompetencyImportHandler) Import(c *gin.Context) {
	expectedHash := strings.ToLower(strings.TrimSpace(c.PostForm("expectedHash")))
	if expectedHash == "" {
		response.RestErr(c, "expectedHash 为空，请先执行导入预览")
		return
	}
	_, rows, digest, err := readCompetencyImportFile(c)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(digest)) != 1 {
		response.RestErr(c, "导入文件与预览文件不一致")
		return
	}
	validation, err := h.validateImportRows(rows)
	if err != nil {
		response.RestErr(c, "校验导入数据失败")
		return
	}
	if len(validation.Errors) > 0 {
		response.RestErr(c, "导入数据存在错误，请重新预览并修正")
		return
	}

	now := time.Now()
	questions := make([]model.Qu, 0, len(validation.ValidRows))
	for _, row := range validation.ValidRows {
		questionType := row.QuestionType
		questionCode := row.QuestionCode
		dimensionID := row.DimensionID
		itemNo := row.DimensionItemNo
		observationPoint := row.ObservationPoint
		direction := row.Direction
		questions = append(questions, model.Qu{
			ID:                     fmt.Sprintf("%d", nextID()),
			QuType:                 quTypeRadio,
			Content:                row.Content,
			CreateTime:             &now,
			UpdateTime:             &now,
			Remark:                 row.Remark,
			QuestionCode:           &questionCode,
			DimensionID:            &dimensionID,
			DimensionItemNo:        &itemNo,
			ObservationPoint:       &observationPoint,
			ScoringDirection:       &direction,
			CompetencyQuestionType: &questionType,
			QuestionStatus:         row.Status,
		})
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(questions, 100).Error
	})
	if err != nil {
		response.RestErr(c, "导入题目失败，已回滚全部数据")
		return
	}
	response.Rest(c, gin.H{"importedCount": len(questions), "sha256": digest})
}

func readCompetencyImportFile(c *gin.Context) ([]byte, [][]string, string, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil {
		return nil, nil, "", errors.New("请上传导入文件")
	}
	if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".xlsx") {
		return nil, nil, "", errors.New("只支持xlsx格式文件")
	}
	if fileHeader.Size <= 0 || fileHeader.Size > competencyImportMaxFileSize {
		return nil, nil, "", errors.New("文件必须大于0且不超过10MiB")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, "", errors.New("读取导入文件失败")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, competencyImportMaxFileSize+1))
	if err != nil || len(data) == 0 || int64(len(data)) > competencyImportMaxFileSize {
		return nil, nil, "", errors.New("读取导入文件失败或文件超过10MiB")
	}
	workbook, err := excelize.OpenReader(bytes.NewReader(data), excelize.Options{
		UnzipSizeLimit:    competencyImportMaxUnzipSize,
		UnzipXMLSizeLimit: 16 << 20,
	})
	if err != nil {
		return nil, nil, "", errors.New("解析xlsx文件失败")
	}
	defer workbook.Close()
	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, "", errors.New("xlsx文件没有可读工作表")
	}
	rows, err := workbook.GetRows(sheets[0])
	if err != nil {
		return nil, nil, "", errors.New("读取xlsx工作表失败")
	}
	if len(rows) > competencyImportMaxRows+1 {
		return nil, nil, "", errors.New("单次导入最多10000道题")
	}
	sum := sha256.Sum256(data)
	return data, rows, hex.EncodeToString(sum[:]), nil
}

func (h *CompetencyImportHandler) validateImportRows(rows [][]string) (service.CompetencyImportValidation, error) {
	var dimensions []model.CompetencyDimension
	if err := h.db.Order("display_order ASC").Find(&dimensions).Error; err != nil {
		return service.CompetencyImportValidation{}, err
	}
	refs := make([]service.CompetencyImportDimension, 0, len(dimensions))
	for _, dimension := range dimensions {
		refs = append(refs, service.CompetencyImportDimension{
			ID: dimension.ID, Order: dimension.DisplayOrder, Name: dimension.Name, Status: dimension.Status,
		})
	}

	var codes []string
	if err := h.db.Model(&model.Qu{}).
		Where("question_code IS NOT NULL").
		Pluck("question_code", &codes).Error; err != nil {
		return service.CompetencyImportValidation{}, err
	}
	existingCodes := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		existingCodes[code] = struct{}{}
	}
	type itemRow struct {
		DimensionID  string `gorm:"column:dimension_id"`
		QuestionType string `gorm:"column:competency_question_type"`
		ItemNo       int    `gorm:"column:dimension_item_no"`
	}
	var items []itemRow
	if err := h.db.Model(&model.Qu{}).
		Select("dimension_id, competency_question_type, dimension_item_no").
		Where("dimension_id IS NOT NULL AND competency_question_type IS NOT NULL AND dimension_item_no IS NOT NULL").
		Find(&items).Error; err != nil {
		return service.CompetencyImportValidation{}, err
	}
	existingItems := make(map[string]struct{}, len(items))
	for _, item := range items {
		existingItems[fmt.Sprintf("%s:%s:%d", item.DimensionID, item.QuestionType, item.ItemNo)] = struct{}{}
	}
	return service.ValidateCompetencyImportRows(rows, refs, existingCodes, existingItems), nil
}
