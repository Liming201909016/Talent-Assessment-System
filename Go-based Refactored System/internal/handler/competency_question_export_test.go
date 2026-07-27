package handler

import (
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/service"
	"github.com/xuri/excelize/v2"
)

func TestCompetencyQuestionExportWorkbook_UsesImportCompatibleColumns(t *testing.T) {
	rows := []competencyQuestionExportRow{
		{DimensionOrder: 1, DimensionName: "沟通表达", QuestionCode: "D01-Q01", DimensionItemNo: 1, Content: "正向题", ObservationPoint: "结构表达", ScoringDirection: "forward", QuestionStatus: 0, Remark: "备注一"},
		{DimensionOrder: 1, DimensionName: "沟通表达", QuestionCode: "D01-Q02", DimensionItemNo: 2, Content: "反向题", ObservationPoint: "沟通调整", ScoringDirection: "reverse", QuestionStatus: 1, Remark: "备注二"},
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
	assertWorkbookCell(t, file, sheet, "C2", "D01-Q01")
	assertWorkbookCell(t, file, sheet, "G2", "正向")
	assertWorkbookCell(t, file, sheet, "H2", "启用")
	assertWorkbookCell(t, file, sheet, "C3", "D01-Q02")
	assertWorkbookCell(t, file, sheet, "G3", "反向")
	assertWorkbookCell(t, file, sheet, "H3", "停用")
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
	for _, required := range []string{"q.dimension_id IS NOT NULL", "d.display_order ASC, q.dimension_item_no ASC, q.id ASC", "Scan(&rows).Error", "if err != nil", "查询胜任力题目失败"} {
		if !strings.Contains(body, required) {
			t.Errorf("question export missing %q", required)
		}
	}
	if strings.Index(body, "if err != nil") > strings.Index(body, `c.Header("Content-Type"`) {
		t.Fatal("query error must be handled before writing xlsx response headers")
	}
}
