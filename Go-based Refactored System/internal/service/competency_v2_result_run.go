package service

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	phase1V2ResultRunSourceSubmission          = "submission"
	phase1V2ResultRunSourceHistoricalRecompute = "historical_recompute"
	phase1V2ResultRunStatusCompleted           = "completed"
)

type phase1V2ResultRunRecords struct {
	Run        model.CompetencyResultRun
	Overall    model.CompetencyResultRunOverall
	Modules    []model.CompetencyResultRunModule
	Dimensions []model.CompetencyResultRunDimension
	Validity   model.CompetencyResultRunValidity
}

type phase1V2AnswerRow struct {
	Sort            int
	Answered        int8
	RawAnswer       *int8
	FinalScore      *int8
	DimensionID     string
	DimensionCode   string
	DimensionName   string
	DisplayOrder    int
	QuestionCode    string
	QuestionType    string
	DimensionItemNo int
	Direction       string
}

func validatePhase1V2ResultRunSource(source string) error {
	if source != phase1V2ResultRunSourceSubmission && source != phase1V2ResultRunSourceHistoricalRecompute {
		return errors.New("phase-1 v2 result-run source is invalid")
	}
	return nil
}

func phase1V2ResultRunError(err error) error {
	if err == nil || errors.Is(err, ErrPhase1V2ResultRunUnavailable) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrPhase1V2ResultRunUnavailable, err)
}

func loadPhase1V2AnswerInputs(tx *gorm.DB, paperID, examID string) ([]CompetencyScoreInput, []Phase1ValidityInput, error) {
	rows := make([]phase1V2AnswerRow, 0, 90)
	if err := tx.Table("el_paper_qu pq").
		Select("pq.sort,pq.answered,pq.raw_answer,pq.final_score,d.dimension_id,d.dimension_code,d.dimension_name,d.display_order,q.question_code,q.competency_question_type AS question_type,q.dimension_item_no,q.scoring_direction AS direction").
		Joins("INNER JOIN el_exam_competency_question q ON q.id=pq.exam_question_id AND q.exam_id = ?", examID).
		Joins("INNER JOIN el_exam_competency_dimension d ON d.id=q.exam_dimension_id AND d.exam_id = ?", examID).
		Where("pq.paper_id = ?", paperID).Order("pq.sort ASC").Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	if len(rows) != 90 {
		return nil, nil, errors.New("phase-1 v2 recompute requires exactly 90 frozen questions")
	}
	dimensions := make([]CompetencyScoreInput, 0, 80)
	validity := make([]Phase1ValidityInput, 0, 10)
	for _, row := range rows {
		if row.Answered != 1 {
			return nil, nil, fmt.Errorf("第%d题尚未作答", row.Sort)
		}
		switch row.QuestionType {
		case model.CompetencyQuestionTypeDimension:
			if row.RawAnswer == nil || row.FinalScore == nil {
				return nil, nil, fmt.Errorf("第%d题缺少原始答案或最终分", row.Sort)
			}
			expectedFinalScore, err := CalculateCompetencyQuestionScore(int(*row.RawAnswer), row.Direction)
			if err != nil || expectedFinalScore != int(*row.FinalScore) {
				return nil, nil, fmt.Errorf("第%d题冻结计分不一致", row.Sort)
			}
			dimensions = append(dimensions, CompetencyScoreInput{
				DimensionID: row.DimensionID, DimensionCode: row.DimensionCode, DimensionName: row.DimensionName,
				DisplayOrder: row.DisplayOrder, QuestionType: CompetencyQuestionTypeDimension, Answered: true, FinalScore: int(*row.FinalScore),
			})
		case model.CompetencyQuestionTypeValidity:
			if row.RawAnswer == nil {
				return nil, nil, fmt.Errorf("第%d题缺少原始答案", row.Sort)
			}
			validity = append(validity, Phase1ValidityInput{
				QuestionCode: row.QuestionCode, Order: row.DimensionItemNo, QuestionType: CompetencyQuestionTypeValidity,
				Direction: row.Direction, Answered: true, RawValue: int(*row.RawAnswer),
			})
		default:
			return nil, nil, errors.New("试卷包含非法胜任力题型")
		}
	}
	sort.Slice(validity, func(i, j int) bool { return validity[i].Order < validity[j].Order })
	return dimensions, validity, nil
}

func decimalFromRat(value *big.Rat) *decimal.Decimal {
	if value == nil {
		return nil
	}
	converted := decimal.NewFromBigRat(value, 6)
	return &converted
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func buildPhase1V2ResultRunRecords(runID string, legacy model.CompetencyResult, dimensionInputs []CompetencyScoreInput, validityInputs []Phase1ValidityInput, userTime int, source string, createdBy *int64, now time.Time) (phase1V2ResultRunRecords, error) {
	if runID == "" || legacy.IsComplete != 1 || legacy.SubmittedAt == nil || !IsPhase1CompetencyVersionSet(CompetencyVersionSet{
		ProductVersion: legacy.ProductVersion, ScoringVersion: legacy.ScoringVersion,
		ContentVersion: legacy.ContentVersion, ReportTemplateVersion: legacy.ReportTemplateVersion,
	}) {
		return phase1V2ResultRunRecords{}, errors.New("completed phase-1 v1 result is required")
	}
	if err := validatePhase1V2ResultRunSource(source); err != nil {
		return phase1V2ResultRunRecords{}, err
	}
	if legacy.ParticipantType != CompetencyParticipantCandidate && legacy.ParticipantType != CompetencyParticipantTester {
		return phase1V2ResultRunRecords{}, errors.New("phase-1 v2 participant type is invalid")
	}
	if legacy.ReportAudience != CompetencyReportAudienceFrontlineEmployee {
		return phase1V2ResultRunRecords{}, errors.New("phase-1 v2 requires the frontline employee report audience")
	}
	if userTime < 1 {
		return phase1V2ResultRunRecords{}, errors.New("phase-1 v2 result-run user time is invalid")
	}
	scores, err := CalculatePhase1V2CompetencyResultFromV1(dimensionInputs)
	if err != nil {
		return phase1V2ResultRunRecords{}, err
	}
	modules, err := CalculatePhase1V2ModuleResults(scores.Dimensions)
	if err != nil {
		return phase1V2ResultRunRecords{}, err
	}
	validity, err := CalculatePhase1ValidityResult(validityInputs)
	if err != nil {
		return phase1V2ResultRunRecords{}, err
	}
	if !scores.IsComplete || !validity.IsComplete || scores.OverallScore == nil || validity.Score == nil {
		return phase1V2ResultRunRecords{}, errors.New("complete phase-1 v2 inputs are required")
	}
	overallNorm, err := Phase1V2OverallNormComparison(scores.OverallScore)
	if err != nil {
		return phase1V2ResultRunRecords{}, err
	}
	versions := Phase1V2VersionSet()
	records := phase1V2ResultRunRecords{
		Run: model.CompetencyResultRun{
			ID: runID, PaperID: legacy.PaperID, ExamID: legacy.ExamID,
			ProductVersion: versions.ProductVersion, ScoringVersion: versions.ScoringVersion,
			ContentVersion: versions.ContentVersion, ReportTemplateVersion: versions.ReportTemplateVersion,
			ReportAudience: legacy.ReportAudience, ParticipantType: legacy.ParticipantType, ParticipantID: legacy.ParticipantID,
			ParticipantName: legacy.ParticipantName, ParticipantTelephone: legacy.ParticipantTelephone, ParticipantAge: legacy.ParticipantAge,
			ParticipantGender: legacy.ParticipantGender, ParticipantAffiliation: legacy.ParticipantAffiliation,
			ParticipantPost: legacy.ParticipantPost, ParticipantDegree: legacy.ParticipantDegree, ParticipantMajor: legacy.ParticipantMajor,
			Source: source, Status: phase1V2ResultRunStatusCompleted, CreatedBy: createdBy, CompletedAt: &now, CreateTime: &now, UpdateTime: &now,
		},
		Overall: model.CompetencyResultRunOverall{
			ResultRunID: runID, TotalQuestionCount: 90, AnsweredQuestionCount: 90,
			DimensionQuestionCount: scores.TotalQuestionCount, AnsweredDimensionQuestionCount: scores.AnsweredQuestionCount,
			EffectiveDimensionCount: scores.EffectiveDimensionCount, OverallScore: decimalFromRat(scores.OverallScore),
			LevelCode: stringPointer(scores.OverallLevel), NormScore: decimalFromRat(overallNorm.NormScore),
			NormComparisonCode: stringPointer(overallNorm.ComparisonCode), IsComplete: 1,
			SubmitType: legacy.SubmitType, SubmittedAt: legacy.SubmittedAt, UserTime: userTime, CreateTime: &now, UpdateTime: &now,
		},
		Modules:    make([]model.CompetencyResultRunModule, 0, len(modules)),
		Dimensions: make([]model.CompetencyResultRunDimension, 0, len(scores.Dimensions)),
		Validity: model.CompetencyResultRunValidity{
			ResultRunID: runID, TotalQuestionCount: validity.TotalQuestionCount, AnsweredQuestionCount: validity.AnsweredQuestionCount,
			ValidityScore:  func() *decimal.Decimal { value := decimal.NewFromInt(int64(*validity.Score)); return &value }(),
			ValidityStatus: stringPointer(validity.Status), IsComplete: 1, CreateTime: &now, UpdateTime: &now,
		},
	}
	for _, module := range modules {
		records.Modules = append(records.Modules, model.CompetencyResultRunModule{
			ID: uuid.NewString(), ResultRunID: runID, ModuleID: module.ModuleCode, ModuleCode: module.ModuleCode,
			ModuleName: module.ModuleName, DisplayOrder: module.DisplayOrder, TotalDimensionCount: module.TotalDimensionCount,
			EffectiveDimensionCount: module.EffectiveDimensionCount, ModuleScore: decimalFromRat(module.Score), LevelCode: stringPointer(module.Level),
			NormScore: decimalFromRat(module.NormScore), NormComparisonCode: stringPointer(module.ComparisonCode), IsComplete: 1, CreateTime: &now,
		})
	}
	norms := Phase1V2DimensionNormScores()
	for index, dimension := range scores.Dimensions {
		records.Dimensions = append(records.Dimensions, model.CompetencyResultRunDimension{
			ID: uuid.NewString(), ResultRunID: runID, DimensionID: dimension.DimensionID, DimensionCode: dimension.DisplayCode,
			DimensionName: dimension.DimensionName, DisplayOrder: dimension.DisplayOrder, TotalQuestionCount: dimension.TotalQuestionCount,
			AnsweredQuestionCount: dimension.AnsweredQuestionCount, ScoreSum: dimension.ScoreSum, DimensionScore: decimalFromRat(dimension.Score),
			LevelCode: stringPointer(dimension.Level), NormScore: decimalFromRat(norms[index]), IsComplete: 1, CreateTime: &now,
		})
	}
	return records, nil
}

func validateExistingPhase1V2ResultRunMetadata(existing, expected model.CompetencyResultRun) error {
	if err := validatePhase1V2ResultRunSource(existing.Source); err != nil {
		return err
	}
	if existing.Status != phase1V2ResultRunStatusCompleted || existing.PaperID != expected.PaperID || existing.ExamID != expected.ExamID ||
		existing.ProductVersion != expected.ProductVersion || existing.ScoringVersion != expected.ScoringVersion ||
		existing.ContentVersion != expected.ContentVersion || existing.ReportTemplateVersion != expected.ReportTemplateVersion ||
		existing.ReportAudience != expected.ReportAudience || existing.ParticipantType != expected.ParticipantType ||
		existing.ParticipantID != expected.ParticipantID || existing.ParticipantName != expected.ParticipantName ||
		existing.ParticipantTelephone != expected.ParticipantTelephone || existing.ParticipantGender != expected.ParticipantGender ||
		existing.ParticipantAffiliation != expected.ParticipantAffiliation || existing.ParticipantPost != expected.ParticipantPost ||
		existing.ParticipantDegree != expected.ParticipantDegree || existing.ParticipantMajor != expected.ParticipantMajor ||
		existing.ErrorMessage != "" || existing.CompletedAt == nil {
		return errors.New("existing phase-1 v2 result run does not match the frozen v1 result")
	}
	if (existing.ParticipantAge == nil) != (expected.ParticipantAge == nil) ||
		(existing.ParticipantAge != nil && *existing.ParticipantAge != *expected.ParticipantAge) {
		return errors.New("existing phase-1 v2 result run participant age does not match")
	}
	return nil
}

func validateExistingPhase1V2ResultRun(tx *gorm.DB, existing model.CompetencyResultRun, expected phase1V2ResultRunRecords) error {
	if err := validateExistingPhase1V2ResultRunMetadata(existing, expected.Run); err != nil {
		return err
	}
	var overall model.CompetencyResultRunOverall
	modules := make([]model.CompetencyResultRunModule, 0, 3)
	dimensions := make([]model.CompetencyResultRunDimension, 0, 10)
	var validity model.CompetencyResultRunValidity
	if err := tx.Where("result_run_id = ?", existing.ID).Take(&overall).Error; err != nil {
		return err
	}
	if err := tx.Where("result_run_id = ?", existing.ID).Order("display_order ASC").Find(&modules).Error; err != nil {
		return err
	}
	if err := tx.Where("result_run_id = ?", existing.ID).Order("display_order ASC").Find(&dimensions).Error; err != nil {
		return err
	}
	if err := tx.Where("result_run_id = ?", existing.ID).Take(&validity).Error; err != nil {
		return err
	}
	persistedScores, persistedModules, persistedValidity, err := phase1V2ReportInputsFromRunRows(overall, modules, dimensions, validity)
	if err != nil {
		return err
	}
	if overall.SubmitType != expected.Overall.SubmitType || overall.SubmittedAt == nil || !overall.SubmittedAt.Equal(*expected.Overall.SubmittedAt) || overall.UserTime != expected.Overall.UserTime ||
		!expected.Overall.OverallScore.Equal(decimal.NewFromBigRat(persistedScores.OverallScore, 6)) ||
		validity.ValidityScore == nil || !validity.ValidityScore.Equal(*expected.Validity.ValidityScore) ||
		persistedValidity != *expected.Validity.ValidityStatus || len(persistedModules) != len(expected.Modules) {
		return errors.New("existing phase-1 v2 result run values do not match recompute")
	}
	for index := range persistedScores.Dimensions {
		if !expected.Dimensions[index].DimensionScore.Equal(decimal.NewFromBigRat(persistedScores.Dimensions[index].Score, 6)) {
			return errors.New("existing phase-1 v2 dimension values do not match recompute")
		}
	}
	return nil
}

func persistPhase1V2ResultRun(tx *gorm.DB, legacy model.CompetencyResult, dimensionInputs []CompetencyScoreInput, validityInputs []Phase1ValidityInput, userTime int, source string, createdBy *int64, now time.Time) (model.CompetencyResultRun, bool, error) {
	var existing model.CompetencyResultRun
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("paper_id = ? AND scoring_version = ?", legacy.PaperID, CompetencyPhase1ScoringVersionV2).Take(&existing).Error
	if err == nil {
		expected, buildErr := buildPhase1V2ResultRunRecords(existing.ID, legacy, dimensionInputs, validityInputs, userTime, source, createdBy, now)
		if buildErr != nil {
			return model.CompetencyResultRun{}, false, buildErr
		}
		if validateErr := validateExistingPhase1V2ResultRun(tx, existing, expected); validateErr != nil {
			return model.CompetencyResultRun{}, false, validateErr
		}
		return existing, true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.CompetencyResultRun{}, false, err
	}
	records, err := buildPhase1V2ResultRunRecords(uuid.NewString(), legacy, dimensionInputs, validityInputs, userTime, source, createdBy, now)
	if err != nil {
		return model.CompetencyResultRun{}, false, err
	}
	if err := tx.Create(&records.Run).Error; err != nil {
		return model.CompetencyResultRun{}, false, err
	}
	if err := tx.Create(&records.Overall).Error; err != nil {
		return model.CompetencyResultRun{}, false, err
	}
	if err := tx.CreateInBatches(records.Modules, 100).Error; err != nil {
		return model.CompetencyResultRun{}, false, err
	}
	if err := tx.CreateInBatches(records.Dimensions, 100).Error; err != nil {
		return model.CompetencyResultRun{}, false, err
	}
	if err := tx.Create(&records.Validity).Error; err != nil {
		return model.CompetencyResultRun{}, false, err
	}
	return records.Run, false, nil
}

func (s *CompetencyRuntimeService) RecomputePhase1V2ResultRun(paperID string, createdBy int64) (model.CompetencyResultRun, bool, error) {
	if paperID == "" || createdBy <= 0 {
		return model.CompetencyResultRun{}, false, errors.New("paper id and operator are required")
	}
	ready, err := s.phase1V2ResultRunSchemaState()
	if err != nil {
		return model.CompetencyResultRun{}, false, phase1V2ResultRunError(err)
	}
	if !ready {
		return model.CompetencyResultRun{}, false, phase1V2ResultRunError(errors.New("phase-1 v2 result-run migration is not applied"))
	}
	var run model.CompetencyResultRun
	var reused bool
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var paper model.Paper
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", paperID).Take(&paper).Error; err != nil {
			return err
		}
		var legacy model.CompetencyResult
		if err := tx.Where("paper_id = ? AND exam_id = ? AND participant_id = ?", paper.ID, paper.ExamID, paper.UserID).Take(&legacy).Error; err != nil {
			return err
		}
		dimensions, validity, err := loadPhase1V2AnswerInputs(tx, paper.ID, paper.ExamID)
		if err != nil {
			return err
		}
		operator := createdBy
		run, reused, err = persistPhase1V2ResultRun(tx, legacy, dimensions, validity, paper.UserTime, phase1V2ResultRunSourceHistoricalRecompute, &operator, time.Now())
		return err
	})
	return run, reused, phase1V2ResultRunError(err)
}
