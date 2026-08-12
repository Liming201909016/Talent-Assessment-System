package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type competencyExportPerson struct {
	PaperID               string           `gorm:"column:paper_id"`
	ParticipantID         string           `gorm:"column:participant_id"`
	ParticipantType       string           `gorm:"column:participant_type"`
	Name                  string           `gorm:"column:participant_name"`
	Telephone             string           `gorm:"column:participant_telephone"`
	StartedAt             *time.Time       `gorm:"column:started_at"`
	SubmittedAt           *time.Time       `gorm:"column:submitted_at"`
	UserTime              int              `gorm:"column:user_time"`
	TotalQuestionCount    int              `gorm:"column:total_question_count"`
	AnsweredQuestionCount int              `gorm:"column:answered_question_count"`
	OverallScore          *decimal.Decimal `gorm:"column:overall_score"`
	EvaluationAverage     *decimal.Decimal `gorm:"column:evaluation_average"`
	EvaluationLevel       string           `gorm:"column:evaluation_level"`
	IsComplete            int8             `gorm:"column:is_complete"`
	SubmitType            string           `gorm:"column:submit_type"`
	ReportAudience        string           `gorm:"column:report_audience"`
	ScoringVersion        string           `gorm:"column:scoring_version"`
	ValidityScore         *decimal.Decimal `gorm:"column:validity_score"`
	ValidityStatus        string           `gorm:"column:validity_status"`
}

type competencyExportAnswer struct {
	PaperID          string `gorm:"column:paper_id"`
	ParticipantID    string `gorm:"column:participant_id"`
	Name             string `gorm:"column:participant_name"`
	Telephone        string `gorm:"column:participant_telephone"`
	Sort             int    `gorm:"column:sort"`
	QuestionCode     string `gorm:"column:question_code"`
	QuestionType     string `gorm:"column:question_type"`
	QuestionContent  string `gorm:"column:question_content"`
	DimensionCode    string `gorm:"column:dimension_code"`
	DimensionName    string `gorm:"column:dimension_name"`
	ObservationPoint string `gorm:"column:observation_point"`
	ScoringDirection string `gorm:"column:scoring_direction"`
	OptionsSnapshot  string `gorm:"column:options_snapshot"`
	RawAnswer        *int8  `gorm:"column:raw_answer"`
	FinalScore       *int8  `gorm:"column:final_score"`
	Answered         int8   `gorm:"column:answered"`
}

type competencyExportData struct {
	Groups           []model.ExamCompetencyGroup
	Dimensions       []model.ExamCompetencyDimension
	Persons          []competencyExportPerson
	GroupResults     []model.CompetencyGroupResult
	DimensionResults []model.CompetencyDimensionResult
	Answers          []competencyExportAnswer
	Questions        []model.ExamCompetencyQuestion
}

func competencyExportIdentity(c *gin.Context) (*model.LoginUser, bool, bool) {
	value, ok := c.Get("loginUser")
	if !ok {
		response.AjaxUnauthorized(c, "")
		return nil, false, false
	}
	login, ok := value.(*model.LoginUser)
	if !ok || login == nil {
		response.AjaxUnauthorized(c, "")
		return nil, false, false
	}
	isAdmin := login.UserID == 1
	if !isAdmin {
		allowed := false
		for _, permission := range login.Permissions {
			if permission == "*:*:*" || permission == "exam:list" || permission == "exam:export" {
				allowed = true
				break
			}
		}
		if !allowed {
			response.Ajax(c, 403, "无权导出胜任力测评结果", nil)
			return nil, false, false
		}
	}
	return login, isAdmin, true
}

func loadCompetencyExportData(db *gorm.DB, examID string) (competencyExportData, error) {
	data := competencyExportData{
		Groups:     make([]model.ExamCompetencyGroup, 0),
		Dimensions: make([]model.ExamCompetencyDimension, 0), Persons: make([]competencyExportPerson, 0),
		GroupResults:     make([]model.CompetencyGroupResult, 0),
		DimensionResults: make([]model.CompetencyDimensionResult, 0), Answers: make([]competencyExportAnswer, 0),
		Questions: make([]model.ExamCompetencyQuestion, 0),
	}
	if err := db.Where("exam_id = ?", examID).Order("display_order ASC").Find(&data.Groups).Error; err != nil {
		return data, err
	}
	if err := db.Where("exam_id = ?", examID).Order("display_order ASC").Find(&data.Dimensions).Error; err != nil {
		return data, err
	}
	if err := db.Where("exam_id = ?", examID).Order("snapshot_order ASC").Find(&data.Questions).Error; err != nil {
		return data, err
	}
	personSelect := `r.paper_id,
		COALESCE(NULLIF(r.participant_id, ''), c.id, t.id, '') AS participant_id,
		COALESCE(NULLIF(r.participant_type, ''), CASE WHEN c.id IS NOT NULL THEN 'candidate' WHEN t.id IS NOT NULL THEN 'tester' ELSE '' END) AS participant_type,
		COALESCE(NULLIF(r.participant_name, ''), c.name, t.name, '') AS participant_name,
		COALESCE(NULLIF(r.participant_telephone, ''), c.telephone, t.telephone, '') AS participant_telephone,
		p.create_time AS started_at, r.submitted_at, p.user_time,
		r.total_question_count, r.answered_question_count, r.overall_score,
		r.evaluation_average, COALESCE(r.evaluation_level, '') AS evaluation_level,
		r.is_complete, r.submit_type, r.report_audience, r.scoring_version,
		vr.validity_score, COALESCE(vr.validity_status, '') AS validity_status`
	if err := db.Table("el_competency_result r").Select(personSelect).
		Joins("INNER JOIN el_paper p ON p.id = r.paper_id").
		Joins("LEFT JOIN el_candidate c ON c.paper_id = r.paper_id AND c.exam_id = r.exam_id").
		Joins("LEFT JOIN el_tester t ON t.paper_id = r.paper_id AND t.exam_id = r.exam_id").
		Joins("LEFT JOIN el_competency_validity_result vr ON vr.paper_id = r.paper_id").
		Where("r.exam_id = ?", examID).
		Order("r.submitted_at ASC, r.paper_id ASC").Scan(&data.Persons).Error; err != nil {
		return data, err
	}
	paperIDs := make([]string, 0, len(data.Persons))
	for _, person := range data.Persons {
		paperIDs = append(paperIDs, person.PaperID)
	}
	if len(paperIDs) == 0 {
		return data, nil
	}
	if err := db.Where("paper_id IN ?", paperIDs).
		Order("paper_id ASC, display_order ASC").Find(&data.DimensionResults).Error; err != nil {
		return data, err
	}
	if err := db.Where("paper_id IN ?", paperIDs).
		Order("paper_id ASC, display_order ASC").Find(&data.GroupResults).Error; err != nil {
		return data, err
	}
	answerSelect := `pq.paper_id,
		COALESCE(c.id, t.id, '') AS participant_id,
		COALESCE(c.name, t.name, '') AS participant_name,
		COALESCE(c.telephone, t.telephone, '') AS participant_telephone,
		pq.sort, q.question_code, q.competency_question_type AS question_type, q.question_content,
		d.dimension_code, d.dimension_name, q.observation_point, q.scoring_direction, q.options_snapshot,
		pq.raw_answer, pq.final_score, pq.answered`
	if err := db.Table("el_paper_qu pq").Select(answerSelect).
		Joins("INNER JOIN el_paper p ON p.id = pq.paper_id").
		Joins("INNER JOIN el_exam_competency_question q ON q.id = pq.exam_question_id").
		Joins("INNER JOIN el_exam_competency_dimension d ON d.id = q.exam_dimension_id").
		Joins("LEFT JOIN el_candidate c ON c.paper_id = pq.paper_id AND c.exam_id = p.exam_id").
		Joins("LEFT JOIN el_tester t ON t.paper_id = pq.paper_id AND t.exam_id = p.exam_id").
		Where("pq.paper_id IN ?", paperIDs).
		Order("pq.paper_id ASC, pq.sort ASC").Scan(&data.Answers).Error; err != nil {
		return data, err
	}
	return data, nil
}

func buildCompetencyExportWorkbook(exam model.Exam, data competencyExportData, isAdmin bool) (*excelize.File, error) {
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
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"409EFF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		file.Close()
		return nil, err
	}

	summaryHeaders := []string{"受测者ID", "类型", "姓名", "手机号", "试卷ID", "开始时间", "完成时间", "答题时长(分钟)", "答题数/总题数", "完成率", "完整性", "提交类型", "整体得分", "总体评价均值", "总体等级", "报告对象", "计分版本"}
	for _, dimension := range data.Dimensions {
		summaryHeaders = append(summaryHeaders, dimension.DimensionCode+" "+dimension.DimensionName)
	}
	groupScores := make(map[string]map[string]*decimal.Decimal)
	for _, result := range data.GroupResults {
		if groupScores[result.PaperID] == nil {
			groupScores[result.PaperID] = make(map[string]*decimal.Decimal)
		}
		groupScores[result.PaperID][result.ExamGroupID] = result.GroupScore
	}
	baseHeaders := append([]string(nil), summaryHeaders[:17]...)
	for _, group := range data.Groups {
		baseHeaders = append(baseHeaders, group.GroupName)
	}
	baseHeaders = append(baseHeaders, "效度原始分", "效度状态")
	summaryHeaders = append(baseHeaders, summaryHeaders[17:]...)
	writeCompetencyHeaders(file, "结果汇总", summaryHeaders, headerStyle)
	dimensionScores := make(map[string]map[string]*decimal.Decimal)
	for _, result := range data.DimensionResults {
		if dimensionScores[result.PaperID] == nil {
			dimensionScores[result.PaperID] = make(map[string]*decimal.Decimal)
		}
		dimensionScores[result.PaperID][result.DimensionID] = result.DimensionScore
	}
	for index, person := range data.Persons {
		row := index + 2
		telephone := person.Telephone
		if !isAdmin {
			telephone = maskPhone(telephone)
		}
		completion := "0.00%"
		if person.TotalQuestionCount > 0 {
			completion = fmt.Sprintf("%.2f%%", float64(person.AnsweredQuestionCount)*100/float64(person.TotalQuestionCount))
		}
		values := []interface{}{
			person.ParticipantID, person.ParticipantType, person.Name, telephone, person.PaperID,
			formatCompetencyExportTime(person.StartedAt), formatCompetencyExportTime(person.SubmittedAt), person.UserTime,
			fmt.Sprintf("%d/%d", person.AnsweredQuestionCount, person.TotalQuestionCount), completion,
			competencyCompleteText(person.IsComplete), person.SubmitType, optionalDecimal(person.OverallScore),
			optionalDecimal(person.EvaluationAverage), person.EvaluationLevel, person.ReportAudience, person.ScoringVersion,
		}
		for _, group := range data.Groups {
			values = append(values, optionalDecimal(groupScores[person.PaperID][group.ID]))
		}
		values = append(values, optionalDecimal(person.ValidityScore), person.ValidityStatus)
		for _, dimension := range data.Dimensions {
			values = append(values, optionalDecimal(dimensionScores[person.PaperID][dimension.DimensionID]))
		}
		writeCompetencyRow(file, "结果汇总", row, values)
	}

	detailHeaders := []string{"受测者ID", "姓名", "手机号", "试卷ID", "个人题序", "题目编号", "题型", "题干快照", "维度编号", "维度名称", "考察点", "计分方向", "原始选择值", "原始选择文本", "最终题目得分", "是否作答"}
	writeCompetencyHeaders(file, "逐题明细", detailHeaders, headerStyle)
	for index, answer := range data.Answers {
		telephone := answer.Telephone
		if !isAdmin {
			telephone = maskPhone(telephone)
		}
		writeCompetencyRow(file, "逐题明细", index+2, []interface{}{
			answer.ParticipantID, answer.Name, telephone, answer.PaperID, answer.Sort,
			answer.QuestionCode, answer.QuestionType, answer.QuestionContent, answer.DimensionCode, answer.DimensionName,
			answer.ObservationPoint, competencyDirectionText(answer.ScoringDirection), optionalInt8(answer.RawAnswer),
			competencyRawAnswerText(answer.RawAnswer, answer.OptionsSnapshot), optionalInt8(answer.FinalScore), competencyAnsweredText(answer.Answered),
		})
	}

	dimensionByID := make(map[string]model.ExamCompetencyDimension, len(data.Dimensions))
	for _, dimension := range data.Dimensions {
		dimensionByID[dimension.ID] = dimension
	}
	dictionaryHeaders := []string{"快照顺序", "题目编号", "题型", "维度编号", "维度名称", "维度内题号", "题干快照", "考察点", "计分方向", "选项快照", "源题ID", "源题更新时间"}
	writeCompetencyHeaders(file, "题目字典", dictionaryHeaders, headerStyle)
	for index, question := range data.Questions {
		dimension := dimensionByID[question.ExamDimensionID]
		writeCompetencyRow(file, "题目字典", index+2, []interface{}{
			question.SnapshotOrder, question.QuestionCode, pointerString(question.CompetencyQuestionType), dimension.DimensionCode, dimension.DimensionName,
			question.DimensionItemNo, question.QuestionContent, question.ObservationPoint,
			competencyDirectionText(question.ScoringDirection), question.OptionsSnapshot, question.SourceQuID,
			formatCompetencyExportTime(question.SourceUpdateTime),
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

func (h *ExamHandler) exportCompetencyWorkbook(c *gin.Context, lu *model.LoginUser, isAdmin bool, exam model.Exam) {
	data, err := loadCompetencyExportData(h.db, exam.ID)
	if err != nil {
		slog.Error("competency export query failed", "examId", exam.ID, "error", err)
		response.RestErr(c, "查询胜任力导出数据失败")
		return
	}
	file, err := buildCompetencyExportWorkbook(exam, data, isAdmin)
	if err != nil {
		response.RestErr(c, "生成胜任力导出文件失败")
		return
	}
	defer file.Close()
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		response.RestErr(c, "生成胜任力导出文件失败")
		return
	}
	fileName := exam.Title + "-胜任力结果明细.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName)+"; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Length", strconv.Itoa(buffer.Len()))
	if _, err := c.Writer.Write(buffer.Bytes()); err != nil {
		slog.Error("competency export response failed", "examId", exam.ID, "error", err)
		return
	}
	if lu != nil {
		_ = h.recordOperLog(c, lu, exam, len(data.Persons), 0, "")
	}
}

func writeCompetencyHeaders(file *excelize.File, sheet string, headers []string, style int) {
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		file.SetCellValue(sheet, cell, header)
		file.SetCellStyle(sheet, cell, cell, style)
	}
}

func writeCompetencyRow(file *excelize.File, sheet string, row int, values []interface{}) {
	for index, value := range values {
		cell, _ := excelize.CoordinatesToCellName(index+1, row)
		file.SetCellValue(sheet, cell, value)
	}
}

func formatCompetencyExportTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func optionalDecimal(value *decimal.Decimal) interface{} {
	if value == nil {
		return ""
	}
	return value.StringFixed(6)
}

func pointerString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func optionalInt8(value *int8) interface{} {
	if value == nil {
		return ""
	}
	return *value
}

func competencyRawAnswerText(value *int8, optionsSnapshot string) string {
	if value == nil {
		return ""
	}
	var options []struct {
		RawValue int    `json:"rawValue"`
		Label    string `json:"label"`
	}
	if json.Unmarshal([]byte(optionsSnapshot), &options) == nil {
		for _, option := range options {
			if option.RawValue == int(*value) {
				return option.Label
			}
		}
	}
	labels := map[int8]string{1: "非常不符合", 2: "不太符合", 3: "一般", 4: "比较符合", 5: "非常符合"}
	return labels[*value]
}

func competencyDirectionText(value string) string {
	if value == "reverse" {
		return "反向"
	}
	return "正向"
}

func competencyCompleteText(value int8) string {
	if value == 1 {
		return "完整"
	}
	return "不完整"
}

func competencyAnsweredText(value int8) string {
	if value == 1 {
		return "已作答"
	}
	return "未作答"
}
