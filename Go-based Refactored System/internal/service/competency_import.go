package service

import (
	"fmt"
	"strconv"
	"strings"
)

var CompetencyImportHeaders = []string{
	"维度序号", "维度名称", "题目类型", "题目编号", "维度内题号",
	"题目内容", "考察点", "计分方向", "启用状态", "备注",
}

var CompetencyImportInstructions = []string{
	"填写当前有效维度的正整数顺序", "必须与维度主数据名称完全一致", "只能填写维度题或效度题", "全局唯一，例如A1-01-Q01",
	"同一维度和题目类型内正整数且唯一", "题干，不能为空", "考察点，不能为空", "只能填写正向或反向", "只能填写启用或停用", "可空",
}

const CompetencyImportDefaultRemark = "AI测试题-未信效度验证"

type CompetencyImportDimension struct {
	ID     string
	Order  int
	Name   string
	Status int8
}

type CompetencyImportRow struct {
	RowNumber        int    `json:"rowNumber"`
	DimensionID      string `json:"dimensionId"`
	DimensionOrder   int    `json:"dimensionOrder"`
	DimensionName    string `json:"dimensionName"`
	QuestionType     string `json:"questionType"`
	QuestionCode     string `json:"questionCode"`
	DimensionItemNo  int    `json:"dimensionItemNo"`
	Content          string `json:"content"`
	ObservationPoint string `json:"observationPoint"`
	Direction        string `json:"direction"`
	Status           int8   `json:"status"`
	Remark           string `json:"remark"`
}

type CompetencyImportRowError struct {
	RowNumber int      `json:"rowNumber"`
	Messages  []string `json:"messages"`
}

type CompetencyImportValidation struct {
	ValidRows []CompetencyImportRow      `json:"validRows"`
	Errors    []CompetencyImportRowError `json:"errors"`
}

func ValidateCompetencyImportRows(
	rows [][]string,
	dimensions []CompetencyImportDimension,
	existingCodes map[string]struct{},
	existingItems map[string]struct{},
) CompetencyImportValidation {
	result := CompetencyImportValidation{
		ValidRows: make([]CompetencyImportRow, 0),
		Errors:    make([]CompetencyImportRowError, 0),
	}
	if !competencyImportHeaderMatches(rows) {
		result.Errors = append(result.Errors, CompetencyImportRowError{
			RowNumber: 1,
			Messages:  []string{"表头必须与胜任力题目导入模板完全一致"},
		})
		return result
	}

	dimensionsByOrder := make(map[int]CompetencyImportDimension, len(dimensions))
	for _, dimension := range dimensions {
		dimensionsByOrder[dimension.Order] = dimension
	}
	seenCodes := make(map[string]struct{})
	seenItems := make(map[string]struct{})

	for index, source := range rows[1:] {
		rowNumber := index + 2
		if rowNumber == 2 && competencyImportInstructionRowMatches(source) {
			continue
		}
		values := make([]string, len(CompetencyImportHeaders))
		isEmpty := true
		for column := range values {
			if column < len(source) {
				values[column] = strings.TrimSpace(source[column])
			}
			if values[column] != "" {
				isEmpty = false
			}
		}
		if isEmpty {
			continue
		}

		messages := make([]string, 0)
		dimensionOrder, orderErr := strconv.Atoi(values[0])
		if orderErr != nil || dimensionOrder < 1 {
			messages = append(messages, "维度序号必须是正整数")
		}
		dimension, dimensionExists := dimensionsByOrder[dimensionOrder]
		if !dimensionExists {
			messages = append(messages, "维度不存在")
		} else {
			if dimension.Status != 0 {
				messages = append(messages, "维度已停用")
			}
			if values[1] != dimension.Name {
				messages = append(messages, "维度名称与主数据不一致")
			}
		}

		questionType := ""
		switch values[2] {
		case "维度题":
			questionType = "dimension"
		case "效度题":
			questionType = "validity"
		default:
			messages = append(messages, "题目类型只能是维度题或效度题")
		}

		questionCode := values[3]
		if questionCode == "" {
			messages = append(messages, "题目编号不能为空")
		} else {
			if _, exists := existingCodes[questionCode]; exists {
				messages = append(messages, "题目编号已存在")
			}
			if _, exists := seenCodes[questionCode]; exists {
				messages = append(messages, "题目编号在文件内重复")
			}
			seenCodes[questionCode] = struct{}{}
		}

		itemNo, itemErr := strconv.Atoi(values[4])
		if itemErr != nil || itemNo <= 0 {
			messages = append(messages, "维度内题号必须是正整数")
		} else if dimensionExists {
			itemKey := fmt.Sprintf("%s:%s:%d", dimension.ID, questionType, itemNo)
			if _, exists := existingItems[itemKey]; exists {
				messages = append(messages, "维度内题号已存在")
			}
			if _, exists := seenItems[itemKey]; exists {
				messages = append(messages, "维度内题号在文件内重复")
			}
			seenItems[itemKey] = struct{}{}
		}

		if values[5] == "" {
			messages = append(messages, "题目内容不能为空")
		}
		if values[6] == "" {
			messages = append(messages, "考察点不能为空")
		}
		direction := ""
		switch values[7] {
		case "正向":
			direction = CompetencyDirectionForward
		case "反向":
			direction = CompetencyDirectionReverse
		default:
			messages = append(messages, "计分方向只能是正向或反向")
		}
		if questionType == "validity" && direction == CompetencyDirectionReverse {
			messages = append(messages, "效度题计分方向必须是正向")
		}
		var status int8
		switch values[8] {
		case "启用":
			status = 0
		case "停用":
			status = 1
		default:
			messages = append(messages, "启用状态只能是启用或停用")
		}
		remark := values[9]
		if remark == "" {
			remark = CompetencyImportDefaultRemark
		}

		if len(messages) > 0 {
			result.Errors = append(result.Errors, CompetencyImportRowError{RowNumber: rowNumber, Messages: messages})
			continue
		}
		result.ValidRows = append(result.ValidRows, CompetencyImportRow{
			RowNumber:        rowNumber,
			DimensionID:      dimension.ID,
			DimensionOrder:   dimensionOrder,
			DimensionName:    dimension.Name,
			QuestionType:     questionType,
			QuestionCode:     questionCode,
			DimensionItemNo:  itemNo,
			Content:          values[5],
			ObservationPoint: values[6],
			Direction:        direction,
			Status:           status,
			Remark:           remark,
		})
	}
	if len(result.ValidRows) == 0 && len(result.Errors) == 0 {
		result.Errors = append(result.Errors, CompetencyImportRowError{RowNumber: 0, Messages: []string{"Excel没有题目数据"}})
	}
	return result
}

func competencyImportInstructionRowMatches(row []string) bool {
	if len(row) < len(CompetencyImportInstructions) {
		return false
	}
	for index, expected := range CompetencyImportInstructions {
		if strings.TrimSpace(row[index]) != expected {
			return false
		}
	}
	return true
}

func competencyImportHeaderMatches(rows [][]string) bool {
	if len(rows) == 0 || len(rows[0]) < len(CompetencyImportHeaders) {
		return false
	}
	for index, expected := range CompetencyImportHeaders {
		if strings.TrimSpace(rows[0][index]) != expected {
			return false
		}
	}
	return true
}
