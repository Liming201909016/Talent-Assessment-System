package handler

import (
	"bytes"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
	"github.com/xuri/excelize/v2"
)

type competencyQuestionExportRow struct {
	DimensionOrder   int    `gorm:"column:dimension_order"`
	DimensionName    string `gorm:"column:dimension_name"`
	QuestionType     string `gorm:"column:competency_question_type"`
	QuestionCode     string `gorm:"column:question_code"`
	DimensionItemNo  int    `gorm:"column:dimension_item_no"`
	Content          string `gorm:"column:content"`
	ObservationPoint string `gorm:"column:observation_point"`
	ScoringDirection string `gorm:"column:scoring_direction"`
	QuestionStatus   int8   `gorm:"column:question_status"`
	Remark           string `gorm:"column:remark"`
}

// Export downloads all 00401 source questions in the same ten-column format accepted by import.
func (h *CompetencyImportHandler) Export(c *gin.Context) {
	rows := make([]competencyQuestionExportRow, 0)
	err := h.db.Table("el_qu q").
		Select(`d.display_order AS dimension_order, d.name AS dimension_name,
			q.competency_question_type, q.question_code, q.dimension_item_no, q.content, q.observation_point,
			q.scoring_direction, q.question_status, q.remark`).
		Joins("INNER JOIN el_competency_dimension d ON d.id = q.dimension_id").
		Where("q.dimension_id IS NOT NULL").
		Order("d.display_order ASC, q.competency_question_type ASC, q.dimension_item_no ASC, q.id ASC").
		Scan(&rows).Error
	if err != nil {
		slog.Error("competency question export query failed", "error", err)
		response.RestErr(c, "查询胜任力题目失败")
		return
	}
	file, err := buildCompetencyQuestionExportWorkbook(rows)
	if err != nil {
		response.RestErr(c, "生成胜任力题目导出文件失败")
		return
	}
	defer file.Close()
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		response.RestErr(c, "生成胜任力题目导出文件失败")
		return
	}
	fileName := "00401-competency-questions.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Length", strconv.Itoa(buffer.Len()))
	if _, err := c.Writer.Write(buffer.Bytes()); err != nil {
		slog.Error("competency question export response failed", "error", err)
	}
}

func buildCompetencyQuestionExportWorkbook(rows []competencyQuestionExportRow) (*excelize.File, error) {
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", "胜任力题目")
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"409EFF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		file.Close()
		return nil, err
	}
	for column, header := range service.CompetencyImportHeaders {
		cell, err := excelize.CoordinatesToCellName(column+1, 1)
		if err != nil {
			file.Close()
			return nil, err
		}
		if err := file.SetCellValue("胜任力题目", cell, header); err != nil {
			file.Close()
			return nil, err
		}
		if err := file.SetCellStyle("胜任力题目", cell, cell, headerStyle); err != nil {
			file.Close()
			return nil, err
		}
	}
	for index, row := range rows {
		values := []interface{}{
			row.DimensionOrder, row.DimensionName, competencyQuestionTypeText(row.QuestionType), row.QuestionCode, row.DimensionItemNo,
			row.Content, row.ObservationPoint, competencyDirectionText(row.ScoringDirection),
			competencyQuestionStatusText(row.QuestionStatus), row.Remark,
		}
		for column, value := range values {
			cell, err := excelize.CoordinatesToCellName(column+1, index+2)
			if err != nil {
				file.Close()
				return nil, err
			}
			if err := file.SetCellValue("胜任力题目", cell, value); err != nil {
				file.Close()
				return nil, err
			}
		}
	}
	if err := file.SetPanes("胜任力题目", &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.SetColWidth("胜任力题目", "A", "E", 18); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.SetColWidth("胜任力题目", "F", "G", 42); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.SetColWidth("胜任力题目", "H", "J", 24); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

func competencyQuestionTypeText(questionType string) string {
	if questionType == "validity" {
		return "效度题"
	}
	return "维度题"
}

func competencyQuestionStatusText(status int8) string {
	if status == 1 {
		return "停用"
	}
	return "启用"
}
