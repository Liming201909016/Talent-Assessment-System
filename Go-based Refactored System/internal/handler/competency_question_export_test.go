package handler

import (
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/service"
	"github.com/xuri/excelize/v2"
)

func TestCompetencyQuestionExportWorkbook_UsesImportCompatibleColumns(t *testing.T) {
	rows := []competencyQuestionExportRow{
		{DimensionOrder: 1, DimensionName: "逻辑思维", QuestionType: "dimension", QuestionCode: "A1-01-Q01", DimensionItemNo: 1, Content: "反向题", ObservationPoint: "逻辑验证", ScoringDirection: "reverse", QuestionStatus: 0, Remark: "备注一"},
		{DimensionOrder: 1, DimensionName: "逻辑思维", QuestionType: "validity", QuestionCode: "P1-VAL-Q01", DimensionItemNo: 1, Content: "效度题", ObservationPoint: "极端美化", ScoringDirection: "forward", QuestionStatus: 0, Remark: "备注二"},
	}
	file, err := buildCompetencyQuestionExportWorkbook(rows)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sheet := file.GetSheetList()[0]
	for index, header := range service.CompetencyImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		assertWorkbookCell(t, file, sheet, cell, header)
	}
	assertWorkbookCell(t, file, sheet, "A2", "1")
	assertWorkbookCell(t, file, sheet, "C2", "维度题")
	assertWorkbookCell(t, file, sheet, "D2", "A1-01-Q01")
	assertWorkbookCell(t, file, sheet, "H2", "反向")
	assertWorkbookCell(t, file, sheet, "I2", "启用")
	assertWorkbookCell(t, file, sheet, "C3", "效度题")
	assertWorkbookCell(t, file, sheet, "D3", "P1-VAL-Q01")
	assertWorkbookCell(t, file, sheet, "H3", "正向")
	assertWorkbookCell(t, file, sheet, "I3", "启用")
}

func TestCompetencyQuestionExportWorkbook_EmptyDataKeepsHeaders(t *testing.T) {
	file, err := buildCompetencyQuestionExportWorkbook([]competencyQuestionExportRow{})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sheet := file.GetSheetList()[0]
	assertWorkbookCell(t, file, sheet, "A1", service.CompetencyImportHeaders[0])
	assertWorkbookCell(t, file, sheet, "A2", "")
}

func TestCompetencyQuestionExport_QueryFailureStopsBeforeResponseHeaders(t *testing.T) {
	source := readSourceFile(t, "competency_question_export.go")
	body := extractFunctionBody(t, source, "func (h *CompetencyImportHandler) Export(")
	for _, required := range []string{"q.dimension_id IS NOT NULL", "q.competency_question_type", "d.display_order ASC, q.competency_question_type ASC, q.dimension_item_no ASC, q.id ASC", "Scan(&rows).Error", "if err != nil", "查询胜任力题目失败"} {
		if !strings.Contains(body, required) {
			t.Errorf("question export missing %q", required)
		}
	}
	if strings.Index(body, "if err != nil") > strings.Index(body, `c.Header("Content-Type"`) {
		t.Fatal("query error must be handled before writing xlsx response headers")
	}
}
