package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

type CompetencyQuestionPageRow struct {
	model.Qu
	DimensionCode  string `gorm:"column:dimension_code" json:"dimensionCode"`
	DimensionName  string `gorm:"column:dimension_name" json:"dimensionName"`
	DimensionOrder int    `gorm:"column:dimension_order" json:"dimensionOrder"`
}

type CompetencyQuestionUpdateRequest struct {
	ID               string      `json:"id"`
	Content          string      `json:"content"`
	ObservationPoint string      `json:"observationPoint"`
	ScoringDirection string      `json:"scoringDirection"`
	QuestionStatus   interface{} `json:"questionStatus"`
	Remark           string      `json:"remark"`
}

type CompetencyDimensionUpdateRequest struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	VIRDLevel          string      `json:"virdLevel"`
	ApplicableCategory string      `json:"applicableCategory"`
	CoreMeaning        string      `json:"coreMeaning"`
	DisplayOrder       interface{} `json:"displayOrder"`
	Status             interface{} `json:"status"`
}

var errCompetencyDimensionConflict = errors.New("competency dimension name or order conflict")

const competencyPhase1DimensionCount = 10

func normalizeCompetencyDimensionInteger(value interface{}, field string) (int, error) {
	switch number := value.(type) {
	case int:
		return number, nil
	case int8:
		return int(number), nil
	case float64:
		if number != float64(int(number)) {
			return 0, fmt.Errorf("%s必须是整数", field)
		}
		return int(number), nil
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(number))
		if err != nil {
			return 0, fmt.Errorf("%s必须是整数", field)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("%s必须是整数", field)
	}
}

func validateCompetencyDimensionUpdate(request *CompetencyDimensionUpdateRequest) error {
	request.ID = strings.TrimSpace(request.ID)
	request.Name = strings.TrimSpace(request.Name)
	request.VIRDLevel = strings.TrimSpace(request.VIRDLevel)
	request.ApplicableCategory = strings.TrimSpace(request.ApplicableCategory)
	request.CoreMeaning = strings.TrimSpace(request.CoreMeaning)
	if request.ID == "" {
		return fmt.Errorf("id 为空")
	}
	for _, field := range []struct {
		value, label string
		limit        int
	}{
		{request.Name, "维度名称", 100},
		{request.VIRDLevel, "能力层级", 100},
		{request.ApplicableCategory, "适用对象", 50},
		{request.CoreMeaning, "核心含义", 500},
	} {
		if field.value == "" {
			return fmt.Errorf("%s不能为空", field.label)
		}
		if utf8.RuneCountInString(field.value) > field.limit {
			return fmt.Errorf("%s长度不能超过%d个字符", field.label, field.limit)
		}
	}
	validLayer := map[string]bool{
		"通用能力": true,
		"心理素养": true,
	}
	if !validLayer[request.VIRDLevel] {
		return fmt.Errorf("能力层级不合法")
	}
	if request.ApplicableCategory != "基层员工" {
		return fmt.Errorf("适用对象不合法")
	}
	order, err := normalizeCompetencyDimensionInteger(request.DisplayOrder, "显示顺序")
	if err != nil || order < 1 || order > competencyPhase1DimensionCount {
		return fmt.Errorf("显示顺序必须是1到10的整数")
	}
	status, err := normalizeCompetencyDimensionInteger(request.Status, "维度状态")
	if err != nil || (status != 0 && status != 1) {
		return fmt.Errorf("维度状态只能是启用或停用")
	}
	return nil
}

func normalizeCompetencyQuestionStatus(value interface{}) (int, error) {
	switch status := value.(type) {
	case int:
		return status, nil
	case int8:
		return int(status), nil
	case float64:
		if status != float64(int(status)) {
			return 0, fmt.Errorf("题目状态只能是启用或停用")
		}
		return int(status), nil
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(status))
		if err != nil {
			return 0, fmt.Errorf("题目状态只能是启用或停用")
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("题目状态只能是启用或停用")
	}
}

func validateCompetencyQuestionUpdate(request *CompetencyQuestionUpdateRequest) error {
	request.ID = strings.TrimSpace(request.ID)
	request.Content = strings.TrimSpace(request.Content)
	request.ObservationPoint = strings.TrimSpace(request.ObservationPoint)
	request.ScoringDirection = strings.TrimSpace(request.ScoringDirection)
	if request.ID == "" {
		return fmt.Errorf("id 为空")
	}
	if request.Content == "" {
		return fmt.Errorf("题目内容不能为空")
	}
	if request.ObservationPoint == "" {
		return fmt.Errorf("考察点不能为空")
	}
	if request.ScoringDirection != "forward" && request.ScoringDirection != "reverse" {
		return fmt.Errorf("计分方向只能是正向或反向")
	}
	status, err := normalizeCompetencyQuestionStatus(request.QuestionStatus)
	if err != nil || (status != 0 && status != 1) {
		return fmt.Errorf("题目状态只能是启用或停用")
	}
	return nil
}

// CompetencyDimensionHandler 管理胜任力测评创建时使用的维度主数据。
type CompetencyDimensionHandler struct {
	db *gorm.DB
}

func NewCompetencyDimensionHandler(db *gorm.DB) *CompetencyDimensionHandler {
	return &CompetencyDimensionHandler{db: db}
}

// List 返回全部维度及启停状态，按当前主数据顺序排列。
// 管理员创建测评时只允许选择 status=0 的维度。
func (h *CompetencyDimensionHandler) List(c *gin.Context) {
	rows := make([]model.CompetencyDimension, 0)
	if err := h.db.Order("display_order ASC").Find(&rows).Error; err != nil {
		response.RestErr(c, "查询胜任力维度失败")
		return
	}
	counts, err := loadEnabledQuestionCounts(h.db, nil)
	if err != nil {
		response.RestErr(c, "查询胜任力维度题数失败")
		return
	}
	type dimensionItem struct {
		model.CompetencyDimension
		QuestionCount int `json:"questionCount"`
	}
	result := make([]dimensionItem, 0, len(rows))
	for _, dimension := range rows {
		result = append(result, dimensionItem{
			CompetencyDimension: dimension,
			QuestionCount:       counts[dimension.ID],
		})
	}
	response.Rest(c, result)
}

// Update 维护未来测评使用的维度主数据。稳定ID、编号和创建时间不可修改；
// 已发布测评继续读取发布快照，不受主数据后续变化影响。
func (h *CompetencyDimensionHandler) Update(c *gin.Context) {
	var request CompetencyDimensionUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if err := validateCompetencyDimensionUpdate(&request); err != nil {
		response.RestErr(c, err.Error())
		return
	}
	displayOrder, _ := normalizeCompetencyDimensionInteger(request.DisplayOrder, "显示顺序")
	status, _ := normalizeCompetencyDimensionInteger(request.Status, "维度状态")
	now := time.Now()
	var updated model.CompetencyDimension
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var existing model.CompetencyDimension
		if err := tx.Where("id = ?", request.ID).Take(&existing).Error; err != nil {
			return err
		}
		var conflictCount int64
		if err := tx.Model(&model.CompetencyDimension{}).
			Where("id <> ? AND name = ?", request.ID, request.Name).
			Count(&conflictCount).Error; err != nil {
			return err
		}
		if conflictCount > 0 {
			return errCompetencyDimensionConflict
		}
		var orderOwner model.CompetencyDimension
		if existing.DisplayOrder != displayOrder {
			err := tx.Where("id <> ? AND display_order = ?", request.ID, displayOrder).Take(&orderOwner).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err == nil {
				if err := tx.Model(&model.CompetencyDimension{}).Where("id = ?", orderOwner.ID).
					Update("display_order", -existing.DisplayOrder).Error; err != nil {
					return err
				}
			}
		}
		updates := map[string]interface{}{
			"name": request.Name, "vird_level": request.VIRDLevel,
			"applicable_category": request.ApplicableCategory, "core_meaning": request.CoreMeaning,
			"display_order": displayOrder, "status": status, "update_time": &now,
		}
		result := tx.Model(&model.CompetencyDimension{}).Where("id = ?", request.ID).Updates(updates)
		if result.Error != nil {
			var mysqlError *mysql.MySQLError
			if errors.As(result.Error, &mysqlError) && mysqlError.Number == 1062 {
				return errCompetencyDimensionConflict
			}
			return result.Error
		}
		if orderOwner.ID != "" {
			if err := tx.Model(&model.CompetencyDimension{}).Where("id = ?", orderOwner.ID).
				Update("display_order", existing.DisplayOrder).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ?", request.ID).Take(&updated).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.RestErr(c, "胜任力维度不存在")
		return
	}
	if errors.Is(err, errCompetencyDimensionConflict) {
		response.RestErr(c, "维度名称或显示顺序已存在")
		return
	}
	if err != nil {
		response.RestErr(c, "更新胜任力维度失败")
		return
	}
	response.Rest(c, updated)
}

// QuestionPaging 返回00401虚拟题库中的胜任力源题，只读且不复用传统题目编辑接口。
func (h *CompetencyDimensionHandler) QuestionPaging(c *gin.Context) {
	var req struct {
		Current      int         `json:"current"`
		Size         int         `json:"size"`
		DimensionID  string      `json:"dimensionId"`
		Status       interface{} `json:"status"`
		QuestionCode string      `json:"questionCode"`
		Content      string      `json:"content"`
		Params       struct {
			DimensionID  string      `json:"dimensionId"`
			Status       interface{} `json:"status"`
			QuestionCode string      `json:"questionCode"`
			Content      string      `json:"content"`
		} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if req.Current < 1 {
		req.Current = 1
	}
	req.Size = capPageSize(req.Size)
	dimensionID := strings.TrimSpace(req.DimensionID)
	if dimensionID == "" {
		dimensionID = strings.TrimSpace(req.Params.DimensionID)
	}
	questionCode := strings.TrimSpace(req.QuestionCode)
	if questionCode == "" {
		questionCode = strings.TrimSpace(req.Params.QuestionCode)
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		content = strings.TrimSpace(req.Params.Content)
	}
	status := toIntPtr(req.Status)
	if status == nil {
		status = toIntPtr(req.Params.Status)
	}

	query := h.db.Table("el_qu q").
		Joins("INNER JOIN el_competency_dimension d ON d.id = q.dimension_id").
		Where("q.dimension_id IS NOT NULL")
	if dimensionID != "" {
		query = query.Where("q.dimension_id = ?", dimensionID)
	}
	if status != nil {
		query = query.Where("q.question_status = ?", *status)
	}
	if questionCode != "" {
		query = query.Where("q.question_code LIKE ?", "%"+questionCode+"%")
	}
	if content != "" {
		query = query.Where("q.content LIKE ?", "%"+content+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.RestErr(c, "查询胜任力题目总数失败")
		return
	}
	rows := make([]CompetencyQuestionPageRow, 0)
	if err := query.Select("q.*, d.code AS dimension_code, d.name AS dimension_name, d.display_order AS dimension_order").
		Order("d.display_order ASC, q.dimension_item_no ASC, q.question_code ASC").
		Offset((req.Current - 1) * req.Size).Limit(req.Size).Scan(&rows).Error; err != nil {
		response.RestErr(c, "查询胜任力题目列表失败")
		return
	}
	response.Rest(c, gin.H{"records": rows, "total": total, "current": req.Current, "size": req.Size})
}

// QuestionUpdate 只修改未来发布使用的源题可编辑字段；已发布测评继续读取不可变快照。
func (h *CompetencyDimensionHandler) QuestionUpdate(c *gin.Context) {
	var request CompetencyQuestionUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if err := validateCompetencyQuestionUpdate(&request); err != nil {
		response.RestErr(c, err.Error())
		return
	}
	status, _ := normalizeCompetencyQuestionStatus(request.QuestionStatus)
	now := time.Now()
	updates := map[string]interface{}{
		"content": request.Content, "observation_point": request.ObservationPoint,
		"scoring_direction": request.ScoringDirection, "question_status": status,
		"remark": request.Remark, "update_time": &now,
	}
	result := h.db.Model(&model.Qu{}).
		Where("id = ? AND dimension_id IS NOT NULL", request.ID).
		Updates(updates)
	if result.Error != nil {
		response.RestErr(c, "更新胜任力题目失败")
		return
	}
	if result.RowsAffected == 0 {
		response.RestErr(c, "胜任力题目不存在")
		return
	}
	var question model.Qu
	if err := h.db.Where("id = ? AND dimension_id IS NOT NULL", request.ID).Take(&question).Error; err != nil {
		response.RestErr(c, "查询更新后的胜任力题目失败")
		return
	}
	response.Rest(c, question)
}
