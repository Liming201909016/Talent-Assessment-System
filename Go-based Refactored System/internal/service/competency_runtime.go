package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	CompetencySubmitManual  = "manual"
	CompetencySubmitTimeout = "timeout"
)

var (
	ErrCompetencyNotPublished      = errors.New("competency exam is not published")
	ErrCompetencyAlreadyDone       = errors.New("competency paper is already completed")
	ErrCompetencyPaperExpired      = errors.New("competency paper has expired")
	ErrCompetencyDurationRequired  = errors.New("胜任力测评必须配置大于0的答题时长")
	ErrCompetencyExamEnded         = errors.New("测评已结束，不能开始答题")
	ErrCompetencyTimeoutNotReached = errors.New("试卷尚未到达提交时间")
	ErrCompetencyIncompleteReport  = errors.New("未完整作答，不能生成正式报告")
)

func validateCompetencyStart(exam *model.Exam, now time.Time) error {
	if exam.TotalTime <= 0 {
		return ErrCompetencyDurationRequired
	}
	if exam.EndTime != nil && !now.Before(*exam.EndTime) {
		return ErrCompetencyExamEnded
	}
	return nil
}

func validateCompetencyTimeout(limitTime *time.Time, now time.Time) error {
	if limitTime == nil {
		return ErrCompetencyDurationRequired
	}
	if now.Before(*limitTime) {
		return ErrCompetencyTimeoutNotReached
	}
	return nil
}

func validateCompetencyFormalReport(result model.CompetencyResult) error {
	if result.IsComplete != 1 {
		return ErrCompetencyIncompleteReport
	}
	return nil
}

type CompetencyRuntimeService struct {
	db           *gorm.DB
	cfg          *config.Config
	randomSource io.Reader
}

func NewCompetencyRuntimeService(db *gorm.DB, cfg *config.Config) *CompetencyRuntimeService {
	return &CompetencyRuntimeService{db: db, cfg: cfg, randomSource: rand.Reader}
}

type CompetencyPublishSummary struct {
	ExamID         string     `json:"examId"`
	DimensionCount int        `json:"dimensionCount"`
	QuestionCount  int        `json:"questionCount"`
	PublishedAt    *time.Time `json:"publishedAt"`
	AlreadyDone    bool       `json:"alreadyPublished"`
	CompetencyVersionSet
}

func (s *CompetencyRuntimeService) Publish(examID string, publishedBy *int64) (CompetencyPublishSummary, error) {
	var summary CompetencyPublishSummary
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var exam model.Exam
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", examID).Take(&exam).Error; err != nil {
			return err
		}
		isCompetency, err := ValidateAssessmentMode(exam.AssessmentType, exam.ScoringMode)
		if err != nil || !isCompetency || exam.CompetencyReportAudience == nil {
			return errors.New("测评不是有效的胜任力草稿")
		}
		if exam.TotalTime <= 0 {
			return ErrCompetencyDurationRequired
		}
		frozenVersions := CompetencyVersionSetFromExam(exam)
		if !IsPhase1CompetencyVersionSet(frozenVersions) {
			return errors.New("00401一期固定配置只能使用基层员工、十个A/B维度和已确认版本")
		}
		versions := frozenVersions
		if exam.PublishStatus == 1 {
			var dimensionCount, questionCount int64
			if err := tx.Model(&model.ExamCompetencyDimension{}).Where("exam_id = ?", examID).Count(&dimensionCount).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.ExamCompetencyQuestion{}).Where("exam_id = ?", examID).Count(&questionCount).Error; err != nil {
				return err
			}
			summary = CompetencyPublishSummary{ExamID: examID, DimensionCount: int(dimensionCount), QuestionCount: int(questionCount), PublishedAt: exam.PublishedAt, AlreadyDone: true, CompetencyVersionSet: versions}
			return nil
		}

		var dimensions []model.ExamCompetencyDimension
		if err := tx.Where("exam_id = ?", examID).Order("display_order ASC").Find(&dimensions).Error; err != nil {
			return err
		}
		if len(dimensions) == 0 {
			return errors.New("胜任力测评至少需要一个维度")
		}
		dimensionIDs := make([]string, 0, len(dimensions))
		associationByDimension := make(map[string]model.ExamCompetencyDimension, len(dimensions))
		for _, dimension := range dimensions {
			dimensionIDs = append(dimensionIDs, dimension.DimensionID)
			associationByDimension[dimension.DimensionID] = dimension
		}
		if err := ValidatePhase1CompetencyConfiguration(*exam.CompetencyReportAudience, dimensionIDs, versions); err != nil {
			return errors.New("00401一期固定配置只能使用基层员工、十个A/B维度和已确认版本")
		}
		var masterDimensions []model.CompetencyDimension
		if err := tx.Where("id IN ? AND status = 0", dimensionIDs).Find(&masterDimensions).Error; err != nil {
			return err
		}
		if len(masterDimensions) != len(dimensions) {
			return errors.New("所选胜任力维度不存在或已停用")
		}
		masterByID := make(map[string]model.CompetencyDimension, len(masterDimensions))
		for _, master := range masterDimensions {
			masterByID[master.ID] = master
		}
		sort.Slice(dimensions, func(i, j int) bool {
			return masterByID[dimensions[i].DimensionID].DisplayOrder < masterByID[dimensions[j].DimensionID].DisplayOrder
		})

		var questions []model.Qu
		if err := tx.Where("dimension_id IN ? AND question_status = 0", dimensionIDs).
			Find(&questions).Error; err != nil {
			return err
		}
		for _, question := range questions {
			if question.DimensionID == nil || question.QuestionCode == nil || question.DimensionItemNo == nil || question.ObservationPoint == nil || question.ScoringDirection == nil || question.CompetencyQuestionType == nil || question.Content == "" {
				return fmt.Errorf("题目 %s 的胜任力元数据不完整", question.ID)
			}
		}
		if err := validatePhase1ValidityQuestionOrder(questions); err != nil {
			return err
		}
		sort.Slice(questions, func(i, j int) bool {
			leftOrder := masterByID[*questions[i].DimensionID].DisplayOrder
			rightOrder := masterByID[*questions[j].DimensionID].DisplayOrder
			if leftOrder != rightOrder {
				return leftOrder < rightOrder
			}
			leftType := *questions[i].CompetencyQuestionType
			rightType := *questions[j].CompetencyQuestionType
			if leftType != rightType {
				return leftType == model.CompetencyQuestionTypeDimension
			}
			if *questions[i].DimensionItemNo != *questions[j].DimensionItemNo {
				return *questions[i].DimensionItemNo < *questions[j].DimensionItemNo
			}
			return *questions[i].QuestionCode < *questions[j].QuestionCode
		})
		counts := make(map[string]Phase1QuestionTypeCounts, len(dimensions))
		now := time.Now()
		groupSnapshots := make([]model.ExamCompetencyGroup, 0, 2)
		groupIDByDimension := make(map[string]string, len(dimensions))
		for _, definition := range Phase1CompetencyGroups() {
			groupID := uuid.NewString()
			groupSnapshots = append(groupSnapshots, model.ExamCompetencyGroup{
				ID: groupID, ExamID: examID, GroupCode: definition.Code, GroupName: definition.Name,
				DisplayOrder: definition.DisplayOrder, DimensionCount: len(definition.ChildDimensionIDs),
				QuestionCount: len(definition.ChildDimensionIDs) * 8, CreateTime: &now, SnapshotTime: &now,
			})
			for _, dimensionID := range definition.ChildDimensionIDs {
				groupIDByDimension[dimensionID] = groupID
			}
		}
		if err := tx.CreateInBatches(groupSnapshots, 100).Error; err != nil {
			return err
		}
		snapshots := make([]model.ExamCompetencyQuestion, 0, len(questions))
		seenCodes := make(map[string]struct{}, len(questions))
		for _, question := range questions {
			if question.DimensionID == nil || question.QuestionCode == nil || question.DimensionItemNo == nil || question.ObservationPoint == nil || question.ScoringDirection == nil || question.CompetencyQuestionType == nil || question.Content == "" {
				return fmt.Errorf("题目 %s 的胜任力元数据不完整", question.ID)
			}
			association, ok := associationByDimension[*question.DimensionID]
			if !ok {
				return fmt.Errorf("题目 %s 的维度不属于本测评", *question.QuestionCode)
			}
			if _, duplicate := seenCodes[*question.QuestionCode]; duplicate {
				return fmt.Errorf("题目编号 %s 重复", *question.QuestionCode)
			}
			seenCodes[*question.QuestionCode] = struct{}{}
			questionType := *question.CompetencyQuestionType
			if questionType != model.CompetencyQuestionTypeDimension && questionType != model.CompetencyQuestionTypeValidity {
				return fmt.Errorf("%s 题目类型非法", *question.QuestionCode)
			}
			if *question.ScoringDirection != CompetencyDirectionForward && *question.ScoringDirection != CompetencyDirectionReverse {
				return fmt.Errorf("%s 计分方向非法", *question.QuestionCode)
			}
			if questionType == model.CompetencyQuestionTypeValidity && *question.ScoringDirection != CompetencyDirectionForward {
				return fmt.Errorf("%s 效度题必须正向计分", *question.QuestionCode)
			}
			count := counts[*question.DimensionID]
			if questionType == model.CompetencyQuestionTypeDimension {
				count.Dimension++
			} else {
				count.Validity++
			}
			counts[*question.DimensionID] = count
			options, err := competencyOptionsJSON(*question.ScoringDirection)
			if err != nil {
				return err
			}
			snapshots = append(snapshots, model.ExamCompetencyQuestion{
				ID: uuid.NewString(), ExamID: examID, ExamDimensionID: association.ID, SourceQuID: question.ID,
				QuestionCode: *question.QuestionCode, DimensionItemNo: *question.DimensionItemNo,
				QuestionContent: question.Content, ObservationPoint: *question.ObservationPoint,
				CompetencyQuestionType: question.CompetencyQuestionType,
				ScoringDirection:       *question.ScoringDirection, OptionsSnapshot: options,
				SourceUpdateTime: question.UpdateTime, SnapshotOrder: len(snapshots) + 1, CreateTime: &now,
			})
		}
		if err := ValidatePhase1QuestionInventory(dimensionIDs, counts); err != nil {
			return err
		}
		for _, dimension := range dimensions {
			master := masterByID[dimension.DimensionID]
			groupID := groupIDByDimension[dimension.DimensionID]
			if err := tx.Model(&model.ExamCompetencyDimension{}).Where("id = ?", dimension.ID).
				Updates(competencyDimensionSnapshotUpdates(master, groupID, counts[dimension.DimensionID].Dimension, now)).Error; err != nil {
				return err
			}
		}
		if err := tx.CreateInBatches(snapshots, 100).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Exam{}).Where("id = ? AND publish_status = 0", examID).
			Updates(map[string]any{
				"publish_status": 1, "published_at": &now, "published_by": publishedBy, "total_score": 5 * len(dimensions),
				"competency_product_version": versions.ProductVersion, "competency_scoring_version": versions.ScoringVersion,
				"competency_content_version": versions.ContentVersion, "competency_report_template_version": versions.ReportTemplateVersion,
			}).Error; err != nil {
			return err
		}
		summary = CompetencyPublishSummary{ExamID: examID, DimensionCount: len(dimensions), QuestionCount: len(snapshots), PublishedAt: &now, CompetencyVersionSet: versions}
		return nil
	})
	return summary, err
}

func validatePhase1ValidityQuestionOrder(questions []model.Qu) error {
	seen := make(map[int]struct{}, 10)
	for _, question := range questions {
		if question.CompetencyQuestionType == nil || *question.CompetencyQuestionType != model.CompetencyQuestionTypeValidity {
			continue
		}
		if question.DimensionItemNo == nil || *question.DimensionItemNo < 1 || *question.DimensionItemNo > 10 {
			return errors.New("00401一期效度题全局序号必须为1至10")
		}
		if _, exists := seen[*question.DimensionItemNo]; exists {
			return errors.New("00401一期效度题全局序号不能重复")
		}
		seen[*question.DimensionItemNo] = struct{}{}
	}
	if len(seen) != 10 {
		return errors.New("00401一期效度题全局序号必须完整覆盖1至10")
	}
	return nil
}

func competencyDimensionSnapshotUpdates(master model.CompetencyDimension, groupID string, questionCount int, snapshotTime time.Time) map[string]any {
	return map[string]any{
		"group_id":            groupID,
		"dimension_code":      master.Code,
		"dimension_name":      master.Name,
		"vird_level":          master.VIRDLevel,
		"applicable_category": master.ApplicableCategory,
		"core_meaning":        master.CoreMeaning,
		"display_order":       master.DisplayOrder,
		"question_count":      questionCount,
		"snapshot_time":       &snapshotTime,
	}
}

func competencyOptionsJSON(direction string) (string, error) {
	labels := []string{"完全不符合", "比较不符合", "不确定", "比较符合", "完全符合"}
	type option struct {
		RawValue   int    `json:"rawValue"`
		Label      string `json:"label"`
		FinalScore int    `json:"finalScore"`
	}
	options := make([]option, 0, 5)
	for raw := 1; raw <= 5; raw++ {
		final, err := CalculateCompetencyQuestionScore(raw, direction)
		if err != nil {
			return "", err
		}
		options = append(options, option{RawValue: raw, Label: labels[raw-1], FinalScore: final})
	}
	data, err := json.Marshal(options)
	return string(data), err
}

type CompetencyPaperAccess struct {
	PaperID    string     `json:"paperId"`
	PaperToken string     `json:"paperToken"`
	State      int        `json:"state"`
	LimitTime  *time.Time `json:"limitTime"`
	Completed  bool       `json:"completed"`
}

type participantRow struct {
	ID      string
	PaperID *string
	EndTime *time.Time
}

func (s *CompetencyRuntimeService) CreateOrRestorePaper(claims CompetencyTokenClaims) (CompetencyPaperAccess, error) {
	if claims.Purpose != CompetencyTokenPurposeParticipant {
		return CompetencyPaperAccess{}, errors.New("参与者令牌用途错误")
	}
	var access CompetencyPaperAccess
	err := s.db.Transaction(func(tx *gorm.DB) error {
		table := participantTable(claims.ParticipantType)
		if table == "" {
			return errors.New("参与者类型错误")
		}
		var participant participantRow
		if err := tx.Table(table).Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id, paper_id, end_time").Where("id = ? AND exam_id = ? AND (del_flag IS NULL OR del_flag = 0)", claims.ParticipantID, claims.ExamID).Take(&participant).Error; err != nil {
			return err
		}
		if participant.PaperID != nil && *participant.PaperID != "" {
			var paper model.Paper
			if err := tx.Where("id = ? AND exam_id = ?", *participant.PaperID, claims.ExamID).Take(&paper).Error; err != nil {
				return err
			}
			access = CompetencyPaperAccess{PaperID: paper.ID, State: paper.State, LimitTime: paper.LimitTime, Completed: paper.State == 2 || participant.EndTime != nil}
			return nil
		}
		if participant.EndTime != nil {
			return ErrCompetencyAlreadyDone
		}
		var exam model.Exam
		if err := tx.Where("id = ?", claims.ExamID).Take(&exam).Error; err != nil {
			return err
		}
		isCompetency, err := ValidateAssessmentMode(exam.AssessmentType, exam.ScoringMode)
		if err != nil || !isCompetency || exam.PublishStatus != 1 {
			return ErrCompetencyNotPublished
		}
		now := time.Now()
		if err := validateCompetencyStart(&exam, now); err != nil {
			return err
		}
		var snapshots []model.ExamCompetencyQuestion
		if err := tx.Where("exam_id = ?", claims.ExamID).Order("snapshot_order ASC").Find(&snapshots).Error; err != nil {
			return err
		}
		if len(snapshots) == 0 {
			return ErrCompetencyNotPublished
		}
		ids := make([]string, len(snapshots))
		byID := make(map[string]model.ExamCompetencyQuestion, len(snapshots))
		for i, snapshot := range snapshots {
			ids[i] = snapshot.ID
			byID[snapshot.ID] = snapshot
		}
		if err := ShuffleCompetencyQuestionIDs(ids, s.randomSource); err != nil {
			return err
		}
		limit := now.Add(time.Duration(exam.TotalTime) * time.Minute)
		paper := model.Paper{ID: uuid.NewString(), UserID: claims.ParticipantID, ExamID: exam.ID, Title: exam.Title, TotalTime: exam.TotalTime, TotalScore: exam.TotalScore, State: 0, CreateTime: &now, UpdateTime: &now, LimitTime: &limit}
		if err := tx.Create(&paper).Error; err != nil {
			return err
		}
		paperQuestions := make([]model.PaperQu, 0, len(ids))
		for index, id := range ids {
			snapshot := byID[id]
			snapshotID := snapshot.ID
			paperQuestions = append(paperQuestions, model.PaperQu{ID: uuid.NewString(), PaperID: paper.ID, QuID: snapshot.SourceQuID, QuType: 1, Sort: index + 1, Score: 5, ExamQuestionID: &snapshotID})
		}
		if err := tx.CreateInBatches(paperQuestions, 100).Error; err != nil {
			return err
		}
		if err := tx.Table(table).Where("id = ? AND paper_id IS NULL", participant.ID).Updates(map[string]any{"paper_id": paper.ID, "update_time": &now}).Error; err != nil {
			return err
		}
		access = CompetencyPaperAccess{PaperID: paper.ID, State: paper.State, LimitTime: &limit}
		return nil
	})
	if err != nil {
		return CompetencyPaperAccess{}, err
	}
	ttl := 3 * time.Hour
	if access.LimitTime != nil && time.Until(*access.LimitTime) > 0 {
		ttl = time.Until(*access.LimitTime) + time.Hour
	}
	token, err := CreateCompetencyToken(s.cfg.Jwt.Secret, CompetencyTokenClaims{Purpose: CompetencyTokenPurposePaper, ParticipantType: claims.ParticipantType, ParticipantID: claims.ParticipantID, ExamID: claims.ExamID, PaperID: access.PaperID}, ttl)
	if err != nil {
		return CompetencyPaperAccess{}, err
	}
	access.PaperToken = token
	return access, nil
}

func participantTable(participantType string) string {
	if participantType == CompetencyParticipantTester {
		return "el_tester"
	}
	if participantType == CompetencyParticipantCandidate {
		return "el_candidate"
	}
	return ""
}

type CompetencyPaperQuestionView struct {
	ID       string          `json:"id"`
	Sort     int             `json:"sort"`
	Code     string          `json:"code"`
	Content  string          `json:"content"`
	Answered bool            `json:"answered"`
	RawValue *int8           `json:"rawValue"`
	Options  json.RawMessage `json:"options"`
}
type CompetencyPaperView struct {
	PaperID         string                        `json:"paperId"`
	State           int                           `json:"state"`
	ServerTime      time.Time                     `json:"serverTime"`
	LimitTime       *time.Time                    `json:"limitTime"`
	TotalCount      int                           `json:"totalCount"`
	AnsweredCount   int                           `json:"answeredCount"`
	UnansweredCount int                           `json:"unansweredCount"`
	Questions       []CompetencyPaperQuestionView `json:"questions"`
}

func (s *CompetencyRuntimeService) PaperDetail(claims CompetencyTokenClaims) (CompetencyPaperView, error) {
	if err := claims.ValidateBinding(claims.ParticipantID, claims.ExamID, claims.PaperID); err != nil {
		return CompetencyPaperView{}, err
	}
	var paper model.Paper
	if err := s.db.Where("id = ? AND exam_id = ? AND user_id = ?", claims.PaperID, claims.ExamID, claims.ParticipantID).Take(&paper).Error; err != nil {
		return CompetencyPaperView{}, err
	}
	type row struct {
		ID              string
		Sort            int
		Answered        int8
		RawAnswer       *int8
		QuestionCode    string
		QuestionContent string
		OptionsSnapshot string
	}
	var rows []row
	if err := s.db.Table("el_paper_qu pq").Select("pq.id,pq.sort,pq.answered,pq.raw_answer,q.question_code,q.question_content,q.options_snapshot").Joins("INNER JOIN el_exam_competency_question q ON q.id=pq.exam_question_id").Where("pq.paper_id = ?", paper.ID).Order("pq.sort ASC").Scan(&rows).Error; err != nil {
		return CompetencyPaperView{}, err
	}
	questions := make([]CompetencyPaperQuestionView, 0, len(rows))
	answered := 0
	for _, row := range rows {
		if row.Answered == 1 {
			answered++
		}
		questions = append(questions, CompetencyPaperQuestionView{ID: row.ID, Sort: row.Sort, Code: row.QuestionCode, Content: row.QuestionContent, Answered: row.Answered == 1, RawValue: row.RawAnswer, Options: json.RawMessage(row.OptionsSnapshot)})
	}
	return CompetencyPaperView{PaperID: paper.ID, State: paper.State, ServerTime: time.Now(), LimitTime: paper.LimitTime, TotalCount: len(rows), AnsweredCount: answered, UnansweredCount: len(rows) - answered, Questions: questions}, nil
}

func (s *CompetencyRuntimeService) FillAnswer(claims CompetencyTokenClaims, paperQuestionID string, rawValue int) (int, error) {
	if rawValue < 1 || rawValue > 5 {
		return 0, errors.New("选项值必须在1到5之间")
	}
	answeredCount := 0
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var paper model.Paper
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND exam_id = ? AND user_id = ?", claims.PaperID, claims.ExamID, claims.ParticipantID).Take(&paper).Error; err != nil {
			return err
		}
		if paper.State != 0 {
			return ErrCompetencyAlreadyDone
		}
		if paper.LimitTime != nil && !time.Now().Before(*paper.LimitTime) {
			return ErrCompetencyPaperExpired
		}
		type answerRow struct {
			ID           string
			Direction    string
			QuestionType string
		}
		var row answerRow
		if err := tx.Table("el_paper_qu pq").Select("pq.id,q.scoring_direction AS direction,q.competency_question_type AS question_type").Joins("INNER JOIN el_exam_competency_question q ON q.id=pq.exam_question_id").Where("pq.id = ? AND pq.paper_id = ?", paperQuestionID, paper.ID).Take(&row).Error; err != nil {
			return errors.New("题目不属于此试卷")
		}
		finalScore := rawValue
		switch row.QuestionType {
		case model.CompetencyQuestionTypeDimension:
			var err error
			finalScore, err = CalculateCompetencyQuestionScore(rawValue, row.Direction)
			if err != nil {
				return err
			}
		case model.CompetencyQuestionTypeValidity:
			if row.Direction != CompetencyDirectionForward {
				return errors.New("效度题必须正向计分")
			}
		default:
			return errors.New("题目类型错误")
		}
		if err := tx.Model(&model.PaperQu{}).Where("id = ? AND paper_id = ?", row.ID, paper.ID).Updates(map[string]any{"raw_answer": rawValue, "final_score": finalScore, "answered": 1}).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.PaperQu{}).Where("paper_id = ? AND answered = 1", paper.ID).Count(&count).Error; err != nil {
			return err
		}
		answeredCount = int(count)
		return nil
	})
	return answeredCount, err
}

type CompetencySubmitSummary struct {
	PaperID          string     `json:"paperId"`
	SubmittedAt      *time.Time `json:"submittedAt"`
	IsComplete       bool       `json:"isComplete"`
	AlreadySubmitted bool       `json:"alreadySubmitted"`
}

func (s *CompetencyRuntimeService) Submit(claims CompetencyTokenClaims, submitType string) (CompetencySubmitSummary, error) {
	if submitType != CompetencySubmitManual && submitType != CompetencySubmitTimeout {
		return CompetencySubmitSummary{}, errors.New("提交类型错误")
	}
	var summary CompetencySubmitSummary
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var paper model.Paper
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND exam_id = ? AND user_id = ?", claims.PaperID, claims.ExamID, claims.ParticipantID).Take(&paper).Error; err != nil {
			return err
		}
		now := time.Now()
		if submitType == CompetencySubmitManual && paper.LimitTime != nil && !now.Before(*paper.LimitTime) {
			submitType = CompetencySubmitTimeout
		}
		if submitType == CompetencySubmitTimeout {
			if err := validateCompetencyTimeout(paper.LimitTime, now); err != nil {
				return err
			}
		}
		var existing model.CompetencyResult
		if err := tx.Where("paper_id = ?", paper.ID).Take(&existing).Error; err == nil {
			summary = CompetencySubmitSummary{PaperID: paper.ID, SubmittedAt: existing.SubmittedAt, IsComplete: existing.IsComplete == 1, AlreadySubmitted: true}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		type scoreRow struct {
			PaperQuestionID string
			Sort            int
			Answered        int8
			RawAnswer       *int8
			FinalScore      *int8
			ExamDimensionID string
			ExamGroupID     *string
			DimensionID     string
			DimensionCode   string
			DimensionName   string
			DisplayOrder    int
			QuestionCode    string
			QuestionType    string
			DimensionItemNo int
			Direction       string
		}
		var rows []scoreRow
		if err := tx.Table("el_paper_qu pq").Select("pq.id AS paper_question_id,pq.sort,pq.answered,pq.raw_answer,pq.final_score,q.exam_dimension_id,d.group_id AS exam_group_id,d.dimension_id,d.dimension_code,d.dimension_name,d.display_order,q.question_code,q.competency_question_type AS question_type,q.dimension_item_no,q.scoring_direction AS direction").Joins("INNER JOIN el_exam_competency_question q ON q.id=pq.exam_question_id").Joins("INNER JOIN el_exam_competency_dimension d ON d.id=q.exam_dimension_id").Where("pq.paper_id = ?", paper.ID).Order("pq.sort ASC").Scan(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return errors.New("试卷没有题目")
		}
		if submitType == CompetencySubmitManual {
			for _, row := range rows {
				if row.Answered != 1 {
					return fmt.Errorf("第%d题尚未作答", row.Sort)
				}
			}
		}
		dimensionInputs := make([]CompetencyScoreInput, 0, 80)
		validityInputs := make([]Phase1ValidityInput, 0, 10)
		dimensionSnapshot := make(map[string]scoreRow)
		for _, row := range rows {
			switch row.QuestionType {
			case model.CompetencyQuestionTypeDimension:
				score := 0
				if row.FinalScore != nil {
					score = int(*row.FinalScore)
				}
				dimensionInputs = append(dimensionInputs, CompetencyScoreInput{DimensionID: row.DimensionID, DimensionCode: row.DimensionCode, DimensionName: row.DimensionName, DisplayOrder: row.DisplayOrder, QuestionType: CompetencyQuestionTypeDimension, Answered: row.Answered == 1, FinalScore: score})
				dimensionSnapshot[row.DimensionID] = row
			case model.CompetencyQuestionTypeValidity:
				rawValue := 0
				if row.RawAnswer != nil {
					rawValue = int(*row.RawAnswer)
				}
				validityInputs = append(validityInputs, Phase1ValidityInput{QuestionCode: row.QuestionCode, Order: row.DimensionItemNo, QuestionType: CompetencyQuestionTypeValidity, Direction: row.Direction, Answered: row.Answered == 1, RawValue: rawValue})
			default:
				return errors.New("试卷包含非法胜任力题型")
			}
		}
		sort.Slice(validityInputs, func(i, j int) bool { return validityInputs[i].Order < validityInputs[j].Order })
		calculated, err := CalculatePhase1CompetencyResult(dimensionInputs)
		if err != nil {
			return err
		}
		calculatedGroups, err := CalculatePhase1GroupResults(calculated.Dimensions)
		if err != nil {
			return err
		}
		calculatedValidity, err := CalculatePhase1ValidityResult(validityInputs)
		if err != nil {
			return err
		}
		dimensionResults := make([]model.CompetencyDimensionResult, 0, len(calculated.Dimensions))
		for _, dimension := range calculated.Dimensions {
			source := dimensionSnapshot[dimension.DimensionID]
			var score *decimal.Decimal
			if dimension.Score != nil {
				v := decimal.NewFromBigRat(dimension.Score, 6)
				score = &v
			}
			var level *string
			if dimension.Level != "" {
				v := dimension.Level
				level = &v
			}
			complete := int8(0)
			if dimension.IsComplete {
				complete = 1
			}
			dimensionResults = append(dimensionResults, model.CompetencyDimensionResult{ID: uuid.NewString(), PaperID: paper.ID, ExamDimensionID: source.ExamDimensionID, DimensionID: source.DimensionID, DimensionCode: dimension.DimensionCode, DimensionName: dimension.DimensionName, DisplayOrder: dimension.DisplayOrder, TotalQuestionCount: dimension.TotalQuestionCount, AnsweredQuestionCount: dimension.AnsweredQuestionCount, ScoreSum: dimension.ScoreSum, DimensionScore: score, LevelCode: level, IsComplete: complete, CreateTime: &now})
		}
		if err := tx.CreateInBatches(dimensionResults, 100).Error; err != nil {
			return err
		}
		var exam model.Exam
		if err := tx.Where("id = ?", paper.ExamID).Take(&exam).Error; err != nil {
			return err
		}
		if exam.CompetencyReportAudience == nil {
			return errors.New("报告版本未配置")
		}
		versions := CompetencyVersionSetFromExam(exam)
		if !IsPhase1CompetencyVersionSet(versions) {
			return errors.New("测评不是00401一期固定版本")
		}
		groupResults := make([]model.CompetencyGroupResult, 0, len(calculatedGroups))
		for _, group := range calculatedGroups {
			var examGroupID string
			for _, dimensionID := range group.ChildDimensionIDs {
				source := dimensionSnapshot[dimensionID]
				if source.ExamGroupID == nil || *source.ExamGroupID == "" {
					return fmt.Errorf("%s 一级维度快照不完整", group.GroupCode)
				}
				if examGroupID == "" {
					examGroupID = *source.ExamGroupID
				} else if examGroupID != *source.ExamGroupID {
					return fmt.Errorf("%s 子维度快照归组不一致", group.GroupCode)
				}
			}
			var groupScore *decimal.Decimal
			if group.Score != nil {
				v := decimal.NewFromBigRat(group.Score, 6)
				groupScore = &v
			}
			var groupLevel *string
			if group.Level != "" {
				v := group.Level
				groupLevel = &v
			}
			groupComplete := int8(0)
			if group.IsComplete {
				groupComplete = 1
			}
			groupResults = append(groupResults, model.CompetencyGroupResult{ID: uuid.NewString(), PaperID: paper.ID, ExamGroupID: examGroupID, GroupCode: group.GroupCode, GroupName: group.GroupName, DisplayOrder: group.DisplayOrder, TotalDimensionCount: group.TotalDimensionCount, EffectiveDimensionCount: group.EffectiveDimensionCount, TotalQuestionCount: group.TotalQuestionCount, AnsweredQuestionCount: group.AnsweredQuestionCount, GroupScore: groupScore, LevelCode: groupLevel, IsComplete: groupComplete, ScoringVersion: versions.ScoringVersion, CreateTime: &now})
		}
		if err := tx.CreateInBatches(groupResults, 100).Error; err != nil {
			return err
		}
		validityComplete := int8(0)
		if calculatedValidity.IsComplete {
			validityComplete = 1
		}
		validityStatus := calculatedValidity.Status
		var validityScore *decimal.Decimal
		if calculatedValidity.Score != nil {
			v := decimal.NewFromInt(int64(*calculatedValidity.Score))
			validityScore = &v
		}
		validityResult := model.CompetencyValidityResult{PaperID: paper.ID, TotalQuestionCount: calculatedValidity.TotalQuestionCount, AnsweredQuestionCount: calculatedValidity.AnsweredQuestionCount, ValidityScore: validityScore, ValidityStatus: &validityStatus, IsComplete: validityComplete, ScoringVersion: versions.ScoringVersion, CreateTime: &now, UpdateTime: &now}
		if err := tx.Create(&validityResult).Error; err != nil {
			return err
		}
		var overall *decimal.Decimal
		if calculated.OverallScore != nil {
			value := decimal.NewFromBigRat(calculated.OverallScore, 6)
			overall = &value
		}
		var average *decimal.Decimal
		if calculated.EvaluationAverage != nil {
			v := decimal.NewFromBigRat(calculated.EvaluationAverage, 6)
			average = &v
		}
		var evaluation *string
		if calculated.EvaluationLevel != "" {
			v := calculated.EvaluationLevel
			evaluation = &v
		}
		complete := int8(0)
		isComplete := calculated.IsComplete && calculatedValidity.IsComplete
		if isComplete {
			complete = 1
		}
		var participant struct {
			Name        string `gorm:"column:name"`
			Telephone   string `gorm:"column:telephone"`
			Age         *int   `gorm:"column:age"`
			Gender      string `gorm:"column:gender"`
			Affiliation string `gorm:"column:affiliation"`
			Post        string `gorm:"column:post"`
			Degree      string `gorm:"column:degree"`
			Major       string `gorm:"column:major"`
		}
		table := participantTable(claims.ParticipantType)
		if table == "" {
			return errors.New("参与者类型错误")
		}
		if err := tx.Table(table).Select("name, COALESCE(telephone, '') AS telephone, age, COALESCE(gender, '') AS gender, COALESCE(affiliation, '') AS affiliation, COALESCE(post, '') AS post, COALESCE(degree, '') AS degree, COALESCE(major, '') AS major").
			Where("id = ? AND paper_id = ?", claims.ParticipantID, paper.ID).Take(&participant).Error; err != nil {
			return err
		}
		result := model.CompetencyResult{PaperID: paper.ID, ExamID: paper.ExamID, TotalQuestionCount: len(rows), AnsweredQuestionCount: calculated.AnsweredQuestionCount + calculatedValidity.AnsweredQuestionCount, DimensionQuestionCount: calculated.TotalQuestionCount, AnsweredDimensionQuestionCount: calculated.AnsweredQuestionCount, EffectiveDimensionCount: calculated.EffectiveDimensionCount, OverallScore: overall, EvaluationAverage: average, EvaluationLevel: evaluation, ParticipantType: claims.ParticipantType, ParticipantID: claims.ParticipantID, ParticipantName: participant.Name, ParticipantTelephone: participant.Telephone, ParticipantAge: participant.Age, ParticipantGender: participant.Gender, ParticipantAffiliation: participant.Affiliation, ParticipantPost: participant.Post, ParticipantDegree: participant.Degree, ParticipantMajor: participant.Major, ReportAudience: *exam.CompetencyReportAudience, IsComplete: complete, SubmitType: submitType, ProductVersion: versions.ProductVersion, ScoringVersion: versions.ScoringVersion, ContentVersion: versions.ContentVersion, ReportTemplateVersion: versions.ReportTemplateVersion, SubmittedAt: &now, CreateTime: &now, UpdateTime: &now}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		userTime := 1
		if paper.CreateTime != nil {
			userTime = int(now.Sub(*paper.CreateTime).Minutes())
			if userTime < 1 {
				userTime = 1
			}
		}
		if err := tx.Model(&model.Paper{}).Where("id = ?", paper.ID).Updates(map[string]any{"state": 2, "user_time": userTime, "update_time": &now}).Error; err != nil {
			return err
		}
		if err := tx.Table(table).Where("id = ? AND paper_id = ?", claims.ParticipantID, paper.ID).Updates(map[string]any{"end_time": &now, "update_time": &now}).Error; err != nil {
			return err
		}
		summary = CompetencySubmitSummary{PaperID: paper.ID, SubmittedAt: &now, IsComplete: isComplete}
		return nil
	})
	return summary, err
}

func (s *CompetencyRuntimeService) ResultDetail(paperID string) (map[string]any, error) {
	var result model.CompetencyResult
	if err := s.db.Where("paper_id = ?", paperID).Take(&result).Error; err != nil {
		return nil, err
	}
	dimensions := make([]model.CompetencyDimensionResult, 0)
	if err := s.db.Where("paper_id = ?", paperID).Order("display_order ASC").Find(&dimensions).Error; err != nil {
		return nil, err
	}
	groups := make([]model.CompetencyGroupResult, 0)
	if err := s.db.Where("paper_id = ?", paperID).Order("display_order ASC").Find(&groups).Error; err != nil {
		return nil, err
	}
	var validity *model.CompetencyValidityResult
	var validityRow model.CompetencyValidityResult
	if err := s.db.Where("paper_id = ?", paperID).Take(&validityRow).Error; err == nil {
		validity = &validityRow
	} else if !errors.Is(err, gorm.ErrRecordNotFound) || IsPhase1CompetencyVersionSet(CompetencyVersionSetFromResult(result)) {
		return nil, err
	}
	type auditRow struct {
		Sort             int    `json:"sort"`
		QuestionCode     string `json:"questionCode"`
		QuestionContent  string `json:"questionContent"`
		DimensionCode    string `json:"dimensionCode"`
		DimensionName    string `json:"dimensionName"`
		ObservationPoint string `json:"observationPoint"`
		ScoringDirection string `json:"scoringDirection"`
		RawAnswer        *int8  `json:"rawValue"`
		FinalScore       *int8  `json:"finalScore"`
		Answered         int8   `json:"answered"`
	}
	audit := make([]auditRow, 0)
	err := s.db.Table("el_paper_qu pq").Select("pq.sort,q.question_code,q.question_content,d.dimension_code,d.dimension_name,q.observation_point,q.scoring_direction,pq.raw_answer,pq.final_score,pq.answered").Joins("INNER JOIN el_exam_competency_question q ON q.id=pq.exam_question_id").Joins("INNER JOIN el_exam_competency_dimension d ON d.id=q.exam_dimension_id").Where("pq.paper_id = ?", paperID).Order("pq.sort ASC").Scan(&audit).Error
	if err != nil {
		return nil, err
	}
	return map[string]any{"result": result, "groups": groups, "dimensions": dimensions, "validity": validity, "questions": audit, "reportTextReady": false, "reportTextMessage": "正式解读文案待配置"}, nil
}

func (s *CompetencyRuntimeService) FormalReportData(paperID string) (map[string]any, error) {
	var result model.CompetencyResult
	if err := s.db.Where("paper_id = ?", paperID).Take(&result).Error; err != nil {
		return nil, err
	}
	if err := validateCompetencyFormalReport(result); err != nil {
		return nil, err
	}
	versions := CompetencyVersionSetFromResult(result)
	if IsPhase1CompetencyVersionSet(versions) {
		var contentPackage model.CompetencyReportContentPackage
		if err := s.db.Where("product_version = ? AND scoring_version = ? AND content_version = ? AND template_version = ? AND audience = ?", versions.ProductVersion, versions.ScoringVersion, versions.ContentVersion, versions.ReportTemplateVersion, result.ReportAudience).Take(&contentPackage).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPhase1ReportContentNotApproved
			}
			return nil, err
		}
		if err := ValidatePhase1ReportContentApproval(contentPackage); err != nil {
			return nil, err
		}
		return s.phase1FormalReportData(paperID, result, contentPackage)
	}
	if err := ValidateFrozenCompetencyVersionSet(versions); err != nil {
		return nil, err
	}
	data, err := s.ResultDetail(paperID)
	if err != nil {
		return nil, err
	}
	dimensions, _ := data["dimensions"].([]model.CompetencyDimensionResult)
	dimensionLevels := make(map[string]string, len(dimensions))
	for _, dimension := range dimensions {
		if dimension.LevelCode == nil || *dimension.LevelCode == "" {
			return nil, errors.New("报告维度等级不完整")
		}
		dimensionLevels[dimension.DimensionID] = *dimension.LevelCode
	}
	if result.EvaluationLevel == nil || *result.EvaluationLevel == "" {
		return nil, errors.New("报告总体等级不完整")
	}
	texts := make([]model.CompetencyReportText, 0)
	if err := s.db.Where("content_version = ? AND audience = ? AND status = 0", result.ContentVersion, result.ReportAudience).
		Find(&texts).Error; err != nil {
		return nil, err
	}
	snapshotJSON, err := BuildCompetencyReportTextSnapshot(result.ContentVersion, result.ReportAudience, *result.EvaluationLevel, dimensionLevels, texts)
	if err != nil {
		return nil, err
	}
	var snapshot CompetencyReportTextSnapshot
	if err := json.Unmarshal([]byte(snapshotJSON), &snapshot); err != nil {
		return nil, err
	}
	data["reportTextReady"] = true
	data["reportTextMessage"] = snapshot.Disclaimer
	data["reportText"] = snapshot
	data["productVersion"] = result.ProductVersion
	data["scoringVersion"] = result.ScoringVersion
	data["contentVersion"] = result.ContentVersion
	data["reportTemplateVersion"] = result.ReportTemplateVersion
	var meta struct {
		ExamTitle      string     `gorm:"column:exam_title"`
		RequiredFields string     `gorm:"column:required_fields"`
		StartedAt      *time.Time `gorm:"column:started_at"`
		UserTime       int        `gorm:"column:user_time"`
	}
	if err := s.db.Table("el_paper p").
		Select("e.title AS exam_title, COALESCE(e.required_fields, '') AS required_fields, p.create_time AS started_at, p.user_time").
		Joins("INNER JOIN el_exam e ON e.id = p.exam_id").
		Where("p.id = ?", paperID).Take(&meta).Error; err != nil {
		return nil, err
	}
	type dimensionMeaningRow struct {
		DimensionID string `gorm:"column:dimension_id"`
		CoreMeaning string `gorm:"column:core_meaning"`
	}
	meaningRows := make([]dimensionMeaningRow, 0)
	if err := s.db.Table("el_exam_competency_dimension").
		Select("dimension_id, core_meaning").Where("exam_id = ?", result.ExamID).
		Find(&meaningRows).Error; err != nil {
		return nil, err
	}
	dimensionCoreMeanings := make(map[string]string, len(meaningRows))
	for _, row := range meaningRows {
		dimensionCoreMeanings[row.DimensionID] = row.CoreMeaning
	}
	data["meta"] = map[string]any{
		"examTitle": meta.ExamTitle, "requiredFields": meta.RequiredFields,
		"startedAt": meta.StartedAt, "userTime": meta.UserTime, "generatedAt": time.Now(),
		"dimensionCoreMeanings": dimensionCoreMeanings,
	}
	return data, nil
}

func (s *CompetencyRuntimeService) phase1FormalReportData(paperID string, result model.CompetencyResult, contentPackage model.CompetencyReportContentPackage) (map[string]any, error) {
	data, err := s.ResultDetail(paperID)
	if err != nil {
		return nil, err
	}
	groups, ok := data["groups"].([]model.CompetencyGroupResult)
	if !ok {
		return nil, errors.New("一期正式报告一级维度数据异常")
	}
	dimensions, ok := data["dimensions"].([]model.CompetencyDimensionResult)
	if !ok {
		return nil, errors.New("一期正式报告二级维度数据异常")
	}
	validity, ok := data["validity"].(*model.CompetencyValidityResult)
	if !ok || validity == nil || validity.ValidityStatus == nil {
		return nil, errors.New("一期正式报告效度数据异常")
	}
	if result.EvaluationLevel == nil || *result.EvaluationLevel == "" {
		return nil, errors.New("一期正式报告总体等级不完整")
	}
	dimensionLevels := make(map[string]string, len(dimensions))
	for _, dimension := range dimensions {
		if dimension.LevelCode == nil || *dimension.LevelCode == "" {
			return nil, errors.New("一期正式报告二级维度等级不完整")
		}
		dimensionLevels[dimension.DimensionID] = *dimension.LevelCode
	}
	texts := make([]model.CompetencyReportText, 0)
	if err := s.db.Where("content_version = ? AND audience = ? AND status = 0", result.ContentVersion, result.ReportAudience).
		Where("content_type IN ?", []string{CompetencyReportContentOverall, CompetencyReportContentGroup, CompetencyReportContentDimension, CompetencyReportContentValidity}).
		Find(&texts).Error; err != nil {
		return nil, err
	}
	snapshotJSON, err := BuildPhase1ReportTextSnapshot(result.ContentVersion, result.ReportAudience, *result.EvaluationLevel, *validity.ValidityStatus, dimensionLevels, texts)
	if err != nil {
		return nil, err
	}
	var snapshot Phase1ReportTextSnapshot
	if err := json.Unmarshal([]byte(snapshotJSON), &snapshot); err != nil {
		return nil, err
	}
	if snapshot.Disclaimer != contentPackage.Disclaimer {
		return nil, errors.New("一期正式报告免责声明与批准内容包不一致")
	}
	formal, err := BuildPhase1FormalReportData(result, groups, dimensions, *validity, snapshot)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(formal)
	if err != nil {
		return nil, err
	}
	output := make(map[string]any)
	if err := json.Unmarshal(payload, &output); err != nil {
		return nil, err
	}
	output["reportTextReady"] = true
	output["reportTextMessage"] = snapshot.Disclaimer
	output["productVersion"] = result.ProductVersion
	output["scoringVersion"] = result.ScoringVersion
	output["contentVersion"] = result.ContentVersion
	output["reportTemplateVersion"] = result.ReportTemplateVersion
	var meta struct {
		ExamTitle      string     `gorm:"column:exam_title"`
		RequiredFields string     `gorm:"column:required_fields"`
		StartedAt      *time.Time `gorm:"column:started_at"`
		UserTime       int        `gorm:"column:user_time"`
	}
	if err := s.db.Table("el_paper p").
		Select("e.title AS exam_title, COALESCE(e.required_fields, '') AS required_fields, p.create_time AS started_at, p.user_time").
		Joins("INNER JOIN el_exam e ON e.id = p.exam_id").
		Where("p.id = ?", paperID).Take(&meta).Error; err != nil {
		return nil, err
	}
	output["meta"] = map[string]any{
		"examTitle": meta.ExamTitle, "requiredFields": meta.RequiredFields,
		"startedAt": meta.StartedAt, "userTime": meta.UserTime, "generatedAt": time.Now(),
	}
	return output, nil
}

type CompetencyResultPageRequest struct {
	ExamID        string
	Current       int
	Size          int
	Name          string
	Telephone     string
	Completion    string
	Validity      string
	SortBy        string
	SortDirection string
	DimensionID   string
}

type CompetencyResultPageRow struct {
	model.CompetencyResult
	ParticipantID        string           `gorm:"column:participant_id" json:"participantId"`
	ParticipantName      string           `gorm:"column:participant_name" json:"participantName"`
	ParticipantTelephone string           `gorm:"column:participant_telephone" json:"participantTelephone"`
	ParticipantType      string           `gorm:"column:participant_type" json:"participantType"`
	StartedAt            *time.Time       `gorm:"column:started_at" json:"startedAt"`
	UserTime             int              `gorm:"column:user_time" json:"userTime"`
	DimensionScore       *decimal.Decimal `gorm:"column:sort_dimension_score" json:"sortDimensionScore"`
	ValidityScore        *decimal.Decimal `gorm:"column:validity_score" json:"validityScore"`
	ValidityStatus       *string          `gorm:"column:validity_status" json:"validityStatus"`
}

type competencyResultSort struct {
	OrderClause string
	DimensionID string
}

type competencyResultFilters struct {
	Name           string
	Telephone      string
	IsComplete     *int
	ValidityStatus string
}

func normalizeCompetencyResultFilters(name, telephone, completion, validity string) (competencyResultFilters, error) {
	filters := competencyResultFilters{
		Name:      strings.TrimSpace(name),
		Telephone: strings.TrimSpace(telephone),
	}
	switch strings.TrimSpace(completion) {
	case "", "all":
	case "complete":
		value := 1
		filters.IsComplete = &value
	case "incomplete":
		value := 0
		filters.IsComplete = &value
	default:
		return competencyResultFilters{}, errors.New("完成状态只能是complete或incomplete")
	}
	switch strings.TrimSpace(validity) {
	case "", "all":
	case CompetencyPhase1ValidityGood, CompetencyPhase1ValidityQuestionable, CompetencyPhase1ValidityIncomplete:
		filters.ValidityStatus = strings.TrimSpace(validity)
	default:
		return competencyResultFilters{}, errors.New("效度状态只能是good、questionable或incomplete")
	}
	return filters, nil
}

func applyCompetencyResultFilters(query *gorm.DB, filters competencyResultFilters) *gorm.DB {
	if filters.Name != "" {
		query = query.Where("r.participant_name LIKE ?", "%"+filters.Name+"%")
	}
	if filters.Telephone != "" {
		query = query.Where("r.participant_telephone LIKE ?", "%"+filters.Telephone+"%")
	}
	if filters.IsComplete != nil {
		query = query.Where("r.is_complete = ?", *filters.IsComplete)
	}
	if filters.ValidityStatus != "" {
		query = query.Where("EXISTS (SELECT 1 FROM el_competency_validity_result vr WHERE vr.paper_id = r.paper_id AND vr.validity_status = ?)", filters.ValidityStatus)
	}
	return query
}

func validateCompetencyResultSort(sortBy, direction, dimensionID string) (competencyResultSort, error) {
	if sortBy == "" {
		sortBy = "submittedAt"
	}
	if direction == "" {
		direction = "desc"
	}
	if direction != "asc" && direction != "desc" {
		return competencyResultSort{}, errors.New("排序方向只能是asc或desc")
	}
	directionSQL := strings.ToUpper(direction)
	switch sortBy {
	case "submittedAt":
		return competencyResultSort{OrderClause: "r.submitted_at " + directionSQL + ", r.paper_id " + directionSQL}, nil
	case "overallScore":
		return competencyResultSort{OrderClause: "r.overall_score " + directionSQL + ", r.paper_id " + directionSQL}, nil
	case "dimensionScore":
		dimensionID = strings.TrimSpace(dimensionID)
		if dimensionID == "" {
			return competencyResultSort{}, errors.New("按维度排序时必须选择维度")
		}
		return competencyResultSort{
			OrderClause: "dr.dimension_score IS NULL ASC, dr.dimension_score " + directionSQL + ", r.paper_id " + directionSQL,
			DimensionID: dimensionID,
		}, nil
	default:
		return competencyResultSort{}, errors.New("不支持的结果排序字段")
	}
}

func competencyRankingDefaults(sortBy, completion, validity string) (string, string) {
	if sortBy == "overallScore" || sortBy == "dimensionScore" {
		if strings.TrimSpace(completion) == "" {
			completion = "complete"
		}
		if strings.TrimSpace(validity) == "" {
			validity = CompetencyPhase1ValidityGood
		}
	}
	return completion, validity
}

func (s *CompetencyRuntimeService) ResultPaging(req CompetencyResultPageRequest) ([]CompetencyResultPageRow, int64, error) {
	req.ExamID = strings.TrimSpace(req.ExamID)
	if req.ExamID == "" {
		return nil, 0, errors.New("examId 为空")
	}
	if req.Current < 1 {
		req.Current = 1
	}
	if req.Size < 1 {
		req.Size = 10
	}
	if req.Size > 500 {
		req.Size = 500
	}
	sortSpec, err := validateCompetencyResultSort(req.SortBy, req.SortDirection, req.DimensionID)
	if err != nil {
		return nil, 0, err
	}
	req.Completion, req.Validity = competencyRankingDefaults(req.SortBy, req.Completion, req.Validity)
	filters, err := normalizeCompetencyResultFilters(req.Name, req.Telephone, req.Completion, req.Validity)
	if err != nil {
		return nil, 0, err
	}
	if sortSpec.DimensionID != "" {
		var dimensionCount int64
		if err := s.db.Model(&model.ExamCompetencyDimension{}).
			Where("exam_id = ? AND dimension_id = ?", req.ExamID, sortSpec.DimensionID).
			Count(&dimensionCount).Error; err != nil {
			return nil, 0, err
		}
		if dimensionCount == 0 {
			return nil, 0, errors.New("所选维度不属于此测评")
		}
	}
	base := applyCompetencyResultFilters(s.db.Table("el_competency_result r").Where("r.exam_id = ?", req.ExamID), filters)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]CompetencyResultPageRow, 0)
	selectClause := `r.*, p.create_time AS started_at, p.user_time,
			COALESCE(NULLIF(r.participant_id, ''), c.id, t.id, '') AS participant_id,
			COALESCE(NULLIF(r.participant_name, ''), c.name, t.name, '') AS participant_name,
			COALESCE(NULLIF(r.participant_telephone, ''), c.telephone, t.telephone, '') AS participant_telephone,
			COALESCE(NULLIF(r.participant_type, ''), CASE WHEN c.id IS NOT NULL THEN 'candidate' WHEN t.id IS NOT NULL THEN 'tester' ELSE '' END) AS participant_type,
			NULL AS sort_dimension_score,
			(SELECT vr.validity_score FROM el_competency_validity_result vr WHERE vr.paper_id = r.paper_id) AS validity_score,
			(SELECT vr.validity_status FROM el_competency_validity_result vr WHERE vr.paper_id = r.paper_id) AS validity_status`
	if sortSpec.DimensionID != "" {
		selectClause = strings.Replace(selectClause, "NULL AS sort_dimension_score", "dr.dimension_score AS sort_dimension_score", 1)
	}
	query := s.db.Table("el_competency_result r").
		Select(selectClause).
		Joins("INNER JOIN el_paper p ON p.id = r.paper_id").
		Joins("LEFT JOIN el_candidate c ON c.paper_id = r.paper_id").
		Joins("LEFT JOIN el_tester t ON t.paper_id = r.paper_id").
		Where("r.exam_id = ?", req.ExamID)
	query = applyCompetencyResultFilters(query, filters)
	if sortSpec.DimensionID != "" {
		query = query.Joins("LEFT JOIN el_competency_dimension_result dr ON dr.paper_id = r.paper_id AND dr.dimension_id = ?", sortSpec.DimensionID)
	}
	err = query.Order(sortSpec.OrderClause).
		Offset((req.Current - 1) * req.Size).
		Limit(req.Size).
		Scan(&rows).Error
	return rows, total, err
}

// StableDimensionResults is useful for deterministic report rendering.
func StableDimensionResults(rows []model.CompetencyDimensionResult) {
	sort.Slice(rows, func(i, j int) bool { return rows[i].DisplayOrder < rows[j].DisplayOrder })
}
