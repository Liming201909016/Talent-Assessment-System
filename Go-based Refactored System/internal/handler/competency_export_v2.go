package handler

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type phase1V2ExportData struct {
	Runs             []model.CompetencyResultRun
	Overalls         []model.CompetencyResultRunOverall
	Modules          []model.CompetencyResultRunModule
	Dimensions       []model.CompetencyResultRunDimension
	Validities       []model.CompetencyResultRunValidity
	Answers          []competencyExportAnswer
	Questions        []model.ExamCompetencyQuestion
	SourceDimensions []model.ExamCompetencyDimension
	StartedAt        map[string]*time.Time
}

type phase1V2ExportPaperStart struct {
	PaperID   string     `gorm:"column:paper_id"`
	StartedAt *time.Time `gorm:"column:started_at"`
}

func isPhase1V2ExportExam(exam model.Exam) bool {
	versions := service.CompetencyVersionSet{
		ProductVersion: exam.CompetencyProductVersion, ScoringVersion: exam.CompetencyScoringVersion,
		ContentVersion: exam.CompetencyContentVersion, ReportTemplateVersion: exam.CompetencyReportTemplateVersion,
	}
	return service.IsPhase1CompetencyVersionSet(versions) || service.IsPhase1V2VersionSet(versions)
}

func loadPhase1V2CompetencyExportData(db *gorm.DB, examID string) (phase1V2ExportData, error) {
	data := phase1V2ExportData{
		Runs: make([]model.CompetencyResultRun, 0), Overalls: make([]model.CompetencyResultRunOverall, 0),
		Modules: make([]model.CompetencyResultRunModule, 0), Dimensions: make([]model.CompetencyResultRunDimension, 0),
		Validities: make([]model.CompetencyResultRunValidity, 0), Answers: make([]competencyExportAnswer, 0),
		Questions: make([]model.ExamCompetencyQuestion, 0), SourceDimensions: make([]model.ExamCompetencyDimension, 0), StartedAt: make(map[string]*time.Time),
	}
	versions := service.Phase1V2VersionSet()
	if err := db.Where("exam_id = ? AND product_version = ? AND scoring_version = ? AND content_version = ? AND report_template_version = ? AND report_audience = ? AND status = ?",
		examID, versions.ProductVersion, versions.ScoringVersion, versions.ContentVersion, versions.ReportTemplateVersion, service.CompetencyReportAudienceFrontlineEmployee, "completed").
		Order("completed_at ASC, paper_id ASC").Find(&data.Runs).Error; err != nil {
		return data, err
	}
	if len(data.Runs) == 0 {
		return data, nil
	}
	if err := db.Where("exam_id = ?", examID).Order("snapshot_order ASC").Find(&data.Questions).Error; err != nil {
		return data, err
	}
	if err := db.Where("exam_id = ?", examID).Order("display_order ASC").Find(&data.SourceDimensions).Error; err != nil {
		return data, err
	}
	runIDs := make([]string, 0, len(data.Runs))
	paperIDs := make([]string, 0, len(data.Runs))
	for _, run := range data.Runs {
		runIDs = append(runIDs, run.ID)
		paperIDs = append(paperIDs, run.PaperID)
	}
	if err := db.Where("result_run_id IN ?", runIDs).Find(&data.Overalls).Error; err != nil {
		return data, err
	}
	if err := db.Where("result_run_id IN ?", runIDs).Order("result_run_id ASC, display_order ASC").Find(&data.Modules).Error; err != nil {
		return data, err
	}
	if err := db.Where("result_run_id IN ?", runIDs).Order("result_run_id ASC, display_order ASC").Find(&data.Dimensions).Error; err != nil {
		return data, err
	}
	if err := db.Where("result_run_id IN ?", runIDs).Find(&data.Validities).Error; err != nil {
		return data, err
	}
	starts := make([]phase1V2ExportPaperStart, 0, len(paperIDs))
	if err := db.Table("el_paper").Select("id AS paper_id, create_time AS started_at").Where("id IN ?", paperIDs).Scan(&starts).Error; err != nil {
		return data, err
	}
	for _, start := range starts {
		data.StartedAt[start.PaperID] = start.StartedAt
	}
	answerSelect := `pq.paper_id, r.participant_id, r.participant_name, r.participant_telephone,
		pq.sort, q.question_code, q.competency_question_type AS question_type, q.question_content,
		d.dimension_id, d.dimension_code, d.dimension_name, q.observation_point, q.scoring_direction,
		q.options_snapshot, pq.raw_answer, pq.final_score, pq.answered`
	if err := db.Table("el_paper_qu pq").Select(answerSelect).
		Joins("INNER JOIN el_competency_result_run r ON r.paper_id = pq.paper_id").
		Joins("INNER JOIN el_exam_competency_question q ON q.id = pq.exam_question_id AND q.exam_id = ?", examID).
		Joins("INNER JOIN el_exam_competency_dimension d ON d.id = q.exam_dimension_id AND d.exam_id = ?", examID).
		Where("pq.paper_id IN ? AND r.id IN ?", paperIDs, runIDs).Order("pq.paper_id ASC, pq.sort ASC").Scan(&data.Answers).Error; err != nil {
		return data, err
	}
	return data, nil
}

func buildPhase1V2CompetencyExportWorkbook(exam model.Exam, data phase1V2ExportData, isAdmin bool) (*excelize.File, error) {
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", "结果汇总")
	if _, err := file.NewSheet("逐题明细"); err != nil {
		file.Close()
		return nil, err
	}
	if _, err := file.NewSheet("题目字典"); err != nil {
		file.Close()
		return nil, err
	}
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"00A651"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		file.Close()
		return nil, err
	}
	definitions := service.Phase1V2Dimensions()
	moduleCodes := []string{service.CompetencyPhase1V2ModuleTask, service.CompetencyPhase1V2ModuleInterpersonal, service.CompetencyPhase1V2ModuleSelf}
	moduleNames := map[string]string{
		service.CompetencyPhase1V2ModuleTask: "任务管理类", service.CompetencyPhase1V2ModuleInterpersonal: "人际管理类", service.CompetencyPhase1V2ModuleSelf: "自我管理类",
	}
	summaryHeaders := []string{"受测者ID", "类型", "姓名", "手机号", "试卷ID", "v2结果RunID", "开始时间", "完成时间", "答题时长(分钟)", "答题数/总题数", "完整性", "提交类型", "产品版本", "计分版本", "内容版本", "报告模板版本", "报告对象", "总体得分", "总体等级", "总体常模", "总体常模比较"}
	for _, code := range moduleCodes {
		name := moduleNames[code]
		summaryHeaders = append(summaryHeaders, name+"-得分", name+"-等级", name+"-常模", name+"-常模比较")
	}
	summaryHeaders = append(summaryHeaders, "效度原始分", "效度状态")
	for _, definition := range definitions {
		prefix := definition.DisplayCode + " " + definition.Name
		summaryHeaders = append(summaryHeaders, prefix+"-得分", prefix+"-等级", prefix+"-常模", prefix+"-常模比较")
	}
	writeCompetencyHeaders(file, "结果汇总", summaryHeaders, headerStyle)
	overallByRun := make(map[string]model.CompetencyResultRunOverall, len(data.Overalls))
	moduleByRun := make(map[string]map[string]model.CompetencyResultRunModule)
	dimensionByRun := make(map[string]map[string]model.CompetencyResultRunDimension)
	validityByRun := make(map[string]model.CompetencyResultRunValidity, len(data.Validities))
	for _, row := range data.Overalls {
		overallByRun[row.ResultRunID] = row
	}
	for _, row := range data.Modules {
		if moduleByRun[row.ResultRunID] == nil {
			moduleByRun[row.ResultRunID] = make(map[string]model.CompetencyResultRunModule)
		}
		moduleByRun[row.ResultRunID][row.ModuleCode] = row
	}
	for _, row := range data.Dimensions {
		if dimensionByRun[row.ResultRunID] == nil {
			dimensionByRun[row.ResultRunID] = make(map[string]model.CompetencyResultRunDimension)
		}
		dimensionByRun[row.ResultRunID][row.DimensionID] = row
	}
	for _, row := range data.Validities {
		validityByRun[row.ResultRunID] = row
	}
	for index, run := range data.Runs {
		overall, exists := overallByRun[run.ID]
		modules := moduleByRun[run.ID]
		dimensions := dimensionByRun[run.ID]
		validity, validityExists := validityByRun[run.ID]
		if !exists || !validityExists || len(modules) != 3 || len(dimensions) != 10 || overall.IsComplete != 1 || validity.IsComplete != 1 {
			file.Close()
			return nil, fmt.Errorf("v2导出结果不完整：run=%s", run.ID)
		}
		telephone := run.ParticipantTelephone
		if !isAdmin {
			telephone = maskPhone(telephone)
		}
		values := []interface{}{
			run.ParticipantID, run.ParticipantType, run.ParticipantName, telephone, run.PaperID, run.ID,
			formatCompetencyExportTime(data.StartedAt[run.PaperID]), formatCompetencyExportTime(overall.SubmittedAt), overall.UserTime,
			fmt.Sprintf("%d/%d", overall.AnsweredQuestionCount, overall.TotalQuestionCount), competencyCompleteText(overall.IsComplete), overall.SubmitType,
			run.ProductVersion, run.ScoringVersion, run.ContentVersion, run.ReportTemplateVersion, "基层员工",
			v2Decimal(overall.OverallScore), phase1V2LevelLabel(pointerValue(overall.LevelCode)), v2Decimal(overall.NormScore), phase1V2OverallComparisonLabel(pointerValue(overall.NormComparisonCode)),
		}
		for _, code := range moduleCodes {
			module, ok := modules[code]
			if !ok || module.IsComplete != 1 {
				file.Close()
				return nil, fmt.Errorf("v2导出模块不完整：run=%s module=%s", run.ID, code)
			}
			values = append(values, v2Decimal(module.ModuleScore), phase1V2LevelLabel(pointerValue(module.LevelCode)), v2Decimal(module.NormScore), phase1V2ModuleComparisonLabel(pointerValue(module.NormComparisonCode)))
		}
		values = append(values, v2ValidityScore(validity.ValidityScore), phase1V2ValidityLabel(pointerValue(validity.ValidityStatus)))
		for _, definition := range definitions {
			dimension, ok := dimensions[definition.ID]
			if !ok || dimension.IsComplete != 1 {
				file.Close()
				return nil, fmt.Errorf("v2导出维度不完整：run=%s dimension=%s", run.ID, definition.ID)
			}
			values = append(values, v2Decimal(dimension.DimensionScore), phase1V2LevelLabel(pointerValue(dimension.LevelCode)), v2Decimal(dimension.NormScore), phase1V2DimensionComparison(dimension.DimensionScore, dimension.NormScore))
		}
		writeCompetencyRow(file, "结果汇总", index+2, values)
	}

	detailHeaders := []string{"受测者ID", "姓名", "手机号", "试卷ID", "v2结果RunID", "个人题序", "题目编号", "题型", "题干快照", "原维度编号", "原维度名称", "v2维度编号", "v2维度名称", "v2模块", "考察点", "计分方向", "原始选择值", "原始选择文本", "最终题目得分", "是否作答"}
	writeCompetencyHeaders(file, "逐题明细", detailHeaders, headerStyle)
	runByPaper := make(map[string]model.CompetencyResultRun, len(data.Runs))
	for _, run := range data.Runs {
		runByPaper[run.PaperID] = run
	}
	for index, answer := range data.Answers {
		run := runByPaper[answer.PaperID]
		telephone := run.ParticipantTelephone
		if !isAdmin {
			telephone = maskPhone(telephone)
		}
		definition, mappingErr := service.MapPhase1V1DimensionToV2(answer.DimensionID)
		if mappingErr != nil {
			file.Close()
			return nil, mappingErr
		}
		writeCompetencyRow(file, "逐题明细", index+2, []interface{}{
			run.ParticipantID, run.ParticipantName, telephone, answer.PaperID, run.ID, answer.Sort,
			answer.QuestionCode, answer.QuestionType, answer.QuestionContent, answer.DimensionCode, answer.DimensionName,
			definition.DisplayCode, definition.Name, phase1V2ModuleName(definition.ModuleCode), answer.ObservationPoint,
			competencyDirectionText(answer.ScoringDirection), optionalInt8(answer.RawAnswer), competencyRawAnswerText(answer.RawAnswer, answer.OptionsSnapshot), optionalInt8(answer.FinalScore), competencyAnsweredText(answer.Answered),
		})
	}

	dimensionBySnapshotID := make(map[string]model.ExamCompetencyDimension, len(data.SourceDimensions))
	for _, dimension := range data.SourceDimensions {
		dimensionBySnapshotID[dimension.ID] = dimension
	}
	dictionaryHeaders := []string{"快照顺序", "题目编号", "题型", "v2稳定维度ID", "v2维度编号", "v2维度名称", "v2模块", "维度内题号", "题干快照", "考察点", "计分方向", "选项快照", "源题ID", "源题更新时间"}
	writeCompetencyHeaders(file, "题目字典", dictionaryHeaders, headerStyle)
	for index, question := range data.Questions {
		sourceDimension, exists := dimensionBySnapshotID[question.ExamDimensionID]
		if !exists {
			file.Close()
			return nil, fmt.Errorf("v2题目字典缺少维度快照：%s", question.QuestionCode)
		}
		definition, mappingErr := service.MapPhase1V1DimensionToV2(sourceDimension.DimensionID)
		if mappingErr != nil {
			file.Close()
			return nil, mappingErr
		}
		writeCompetencyRow(file, "题目字典", index+2, []interface{}{
			question.SnapshotOrder, question.QuestionCode, pointerString(question.CompetencyQuestionType), definition.ID,
			definition.DisplayCode, definition.Name, phase1V2ModuleName(definition.ModuleCode), question.DimensionItemNo,
			question.QuestionContent, question.ObservationPoint, competencyDirectionText(question.ScoringDirection), question.OptionsSnapshot,
			question.SourceQuID, formatCompetencyExportTime(question.SourceUpdateTime),
		})
	}
	for sheet, columnCount := range map[string]int{"结果汇总": len(summaryHeaders), "逐题明细": len(detailHeaders), "题目字典": len(dictionaryHeaders)} {
		if err := file.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
			file.Close()
			return nil, err
		}
		lastColumn, err := excelize.ColumnNumberToName(columnCount)
		if err != nil {
			file.Close()
			return nil, err
		}
		if err := file.SetColWidth(sheet, "A", lastColumn, 18); err != nil {
			file.Close()
			return nil, err
		}
	}
	return file, nil
}

func phase1V2ModuleName(code string) string {
	return map[string]string{
		service.CompetencyPhase1V2ModuleTask: "任务管理类", service.CompetencyPhase1V2ModuleInterpersonal: "人际管理类", service.CompetencyPhase1V2ModuleSelf: "自我管理类",
	}[code]
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func v2Decimal(value *decimal.Decimal) interface{} {
	if value == nil {
		return ""
	}
	return value.StringFixed(2)
}

func v2ValidityScore(value *decimal.Decimal) interface{} {
	if value == nil {
		return ""
	}
	return value.StringFixed(0)
}

func phase1V2DimensionComparison(score, norm *decimal.Decimal) string {
	if score == nil || norm == nil {
		return ""
	}
	switch score.Cmp(*norm) {
	case 1:
		return "高于常模"
	case -1:
		return "低于常模"
	default:
		return "与常模持平"
	}
}
