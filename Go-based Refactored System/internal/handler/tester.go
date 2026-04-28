package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// TesterHandler 对应 /exam/api/tester/*
// 基于 el_tester 单表（替代 Java 的 el_tester_profile + el_tester_exam 双表设计）
type TesterHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewTesterHandler(db *gorm.DB, cfg *config.Config) *TesterHandler {
	return &TesterHandler{db: db, cfg: cfg}
}

// GET /exam/api/tester (RuoYi TableDataInfo)
// GET /exam/api/tester/list
// 对齐 Java TesterController.list：直接查 el_tester 单表（Java 原始表）
func (h *TesterHandler) List(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize = capPageSize(pageSize)
	name := c.Query("name")
	idNumber := c.Query("idNumber")
	examID := c.Query("examId")
	examStatus := c.Query("examStatus") // 0=未测评 1=进行中 2=已完成

	q := h.db.Table("el_tester").Where("del_flag IS NULL OR del_flag = '0'")
	if name != "" {
		q = q.Where("name like ?", "%"+name+"%")
	}
	if idNumber != "" {
		q = q.Where("id_number like ?", "%"+idNumber+"%")
	}
	if examID != "" {
		q = q.Where("exam_id = ?", examID)
	}
	switch examStatus {
	case "0":
		q = q.Where("paper_id IS NULL")
	case "1":
		q = q.Where("paper_id IS NOT NULL AND end_time IS NULL")
	case "2":
		q = q.Where("end_time IS NOT NULL")
	}

	var total int64
	q.Count(&total)

	type testerRow struct {
		ID          string     `gorm:"column:id;primaryKey" json:"id"`
		PaperID     *string    `gorm:"column:paper_id"      json:"paperId"`
		ExamID      *string    `gorm:"column:exam_id"       json:"examId"`
		IDNumber    string     `gorm:"column:id_number"     json:"idNumber"`
		Name        string     `gorm:"column:name"          json:"name"`
		Age         *int       `gorm:"column:age"           json:"age"`
		Gender      *string    `gorm:"column:gender"        json:"gender"`
		Password    string     `gorm:"column:password"      json:"password"`
		Telephone   *string    `gorm:"column:telephone"     json:"telephone"`
		Affiliation *string    `gorm:"column:affiliation"   json:"affiliation"`
		Depart      *string    `gorm:"column:depart"        json:"depart"`
		Post        *string    `gorm:"column:post"          json:"post"`
		Degree      *string    `gorm:"column:degree"        json:"degree"`
		Major       *string    `gorm:"column:major"         json:"major"`
		StuFlag     *int       `gorm:"column:stu_flag"      json:"stuFlag"`
		Status      *string    `gorm:"column:status"        json:"status"`
		CreateTime  *time.Time `gorm:"column:create_time"   json:"createTime"`
		UpdateTime  *time.Time `gorm:"column:update_time"   json:"updateTime"`
		DelFlag     *int       `gorm:"column:del_flag"      json:"delFlag"`
		EndTime     *time.Time `gorm:"column:end_time"      json:"endTime"`
		PdfPath     *string    `gorm:"column:pdf_path"      json:"pdfPath"`
		PdfFlag     *int       `gorm:"column:pdf_flag"      json:"pdfFlag"`
		RepoCode    *string    `gorm:"column:repo_code"     json:"repoCode"`
		UserTime    *int       `gorm:"column:user_time"     json:"userTime"`
		AnswerNum   int        `gorm:"-"                    json:"answerNum"`
		Title       string     `gorm:"column:title"         json:"title"`
	}
	rows := make([]testerRow, 0)
	h.db.Table("el_tester AS t").
		Joins("LEFT JOIN el_exam AS ea ON ea.id = t.exam_id").
		Joins("LEFT JOIN el_paper AS pa ON pa.id = t.paper_id").
		Joins("LEFT JOIN el_exam_repo AS er ON er.exam_id = t.exam_id").
		Joins("LEFT JOIN el_repo AS rp ON rp.id = er.repo_id").
		Where("t.del_flag IS NULL OR t.del_flag = '0'").
		Scopes(func(db *gorm.DB) *gorm.DB {
			if name != "" {
				db = db.Where("t.name like ?", "%"+name+"%")
			}
			if idNumber != "" {
				db = db.Where("t.id_number like ?", "%"+idNumber+"%")
			}
			if examID != "" {
				db = db.Where("t.exam_id = ?", examID)
			}
			switch examStatus {
			case "0":
				db = db.Where("t.paper_id IS NULL")
			case "1":
				db = db.Where("t.paper_id IS NOT NULL AND t.end_time IS NULL")
			case "2":
				db = db.Where("t.end_time IS NOT NULL")
			}
			return db
		}).
		Select("t.id, t.paper_id, t.exam_id, t.id_number, t.name, t.age, t.gender, t.password, t.telephone, t.affiliation, t.depart, t.post, t.degree, t.major, t.stu_flag, t.status, t.del_flag, t.end_time, t.pdf_path, t.pdf_flag, t.update_time, COALESCE(pa.create_time, t.create_time) AS create_time, pa.user_time, ea.title, rp.code AS repo_code").
		Order("COALESCE(pa.create_time, t.create_time) desc").
		Offset((pageNum - 1) * pageSize).Limit(pageSize).
		Scan(&rows)
	// 填充 answerNum
	for i := range rows {
		if rows[i].PaperID != nil && *rows[i].PaperID != "" {
			var n int64
			// MBTI（repoCode 003）答题记录在 el_mbti_answer
			if rows[i].RepoCode != nil && strings.HasPrefix(*rows[i].RepoCode, "003") {
				h.db.Table("el_mbti_answer").Where("paper_id = ? AND answered = 1", *rows[i].PaperID).Count(&n)
			} else {
				h.db.Table("el_paper_qu").Where("paper_id = ? AND answered = 1", *rows[i].PaperID).Count(&n)
			}
			rows[i].AnswerNum = int(n)
		}
	}
	response.Table(c, rows, total)
}

// GET /exam/api/tester/:id
func (h *TesterHandler) Detail(c *gin.Context) {
	id := c.Param("id")
	var t model.Tester
	if err := h.db.Table("el_tester").Where("id = ?", id).First(&t).Error; err != nil {
		response.AjaxErr(c, "不存在")
		return
	}
	response.AjaxOK(c, t)
}

// GET /exam/api/tester/idNumber/:idNumber
func (h *TesterHandler) DetailByIDNumber(c *gin.Context) {
	idn := c.Param("idNumber")
	examID := c.Query("examId")
	// 前端可能传 idNumber 或 tester ID，兼容两种查法
	q := h.db.Table("el_tester").Where("(id_number = ? OR id = ?) AND (del_flag IS NULL OR del_flag = '0')", idn, idn)
	if examID != "" {
		q = q.Where("exam_id = ?", examID)
	}
	var t model.Tester
	if err := q.First(&t).Error; err != nil {
		response.AjaxErr(c, "不存在")
		return
	}
	response.AjaxOK(c, t)
}

// testerReq 接收创建/更新请求，对应 el_tester 单表字段。
type testerReq struct {
	ID string `json:"id"`

	// Profile 字段
	IDNumber    string      `json:"idNumber"`
	Name        string      `json:"name"`
	Age         interface{} `json:"age"`
	Gender      *string     `json:"gender"`
	Password    string      `json:"password"`
	Telephone   *string     `json:"telephone"`
	Affiliation *string     `json:"affiliation"`
	Depart      *string     `json:"depart"`
	Post        *string     `json:"post"`
	Degree      *string     `json:"degree"`
	Major       *string     `json:"major"`
	StuFlag     interface{} `json:"stuFlag"`
	Status      *string     `json:"status"`

	// Exam 参与信息
	ExamID  string  `json:"examId"`
	PaperID *string `json:"paperId"`
}

// 对齐 Java TesterServiceImpl.insertTester — 写 el_tester 单表
// POST /exam/api/tester
func (h *TesterHandler) Create(c *gin.Context) {
	var r testerReq
	if err := c.ShouldBindJSON(&r); err != nil {
		response.AjaxErr(c, "参数错误")
		return
	}
	if r.Name == "" {
		response.AjaxErr(c, "姓名不能为空")
		return
	}
	if r.IDNumber == "" || r.ExamID == "" {
		response.AjaxErr(c, "缺少 idNumber 或 examId")
		return
	}

	// 对齐 Java validateClosedExam：仅允许封闭测评(is_open=2)添加人员
	if msg := h.validateClosedExam(r.ExamID); msg != "" {
		response.AjaxErr(c, msg)
		return
	}

	// 唯一性检查：(id_number, exam_id) 不可重复
	var dup int64
	h.db.Table("el_tester").
		Where("id_number = ? AND exam_id = ? AND (del_flag IS NULL OR del_flag = '0')", r.IDNumber, r.ExamID).
		Count(&dup)
	if dup > 0 {
		response.AjaxErr(c, "该测评人员已存在当前测评")
		return
	}

	// 默认密码：身份证号后 4 位
	pwd := r.Password
	if pwd == "" {
		if len(r.IDNumber) >= 4 {
			pwd = r.IDNumber[len(r.IDNumber)-4:]
		} else {
			pwd = r.IDNumber
		}
	}

	now := time.Now()
	delZero := 0
	tester := model.Tester{
		ID:          strconv.FormatInt(nextID(), 10),
		ExamID:      &r.ExamID,
		IDNumber:    r.IDNumber,
		Name:        r.Name,
		Age:         toIntPtr(r.Age),
		Gender:      r.Gender,
		Password:    pwd,
		Telephone:   r.Telephone,
		Affiliation: r.Affiliation,
		Depart:      r.Depart,
		Post:        r.Post,
		Degree:      r.Degree,
		Major:       r.Major,
		StuFlag:     toIntPtr(r.StuFlag),
		Status:      r.Status,
		DelFlag:     &delZero,
		CreateTime:  &now,
		UpdateTime:  &now,
	}
	if r.PaperID != nil {
		tester.PaperID = r.PaperID
	}
	if err := h.db.Create(&tester).Error; err != nil {
		// Handle MySQL duplicate key error (concurrent safety via uk_idnumber_examid)
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "duplicate") {
			response.AjaxErr(c, "该测评人员已存在当前测评")
			return
		}
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, true)
}

// 对齐 Java TesterServiceImpl.updateTester — 写 el_tester 单表
// PUT /exam/api/tester
func (h *TesterHandler) Update(c *gin.Context) {
	var r testerReq
	if err := c.ShouldBindJSON(&r); err != nil || r.ID == "" {
		response.AjaxErr(c, "参数错误")
		return
	}

	// 对齐 Java validateClosedExam：如果传了 examId，必须是封闭测评
	if r.ExamID != "" {
		if msg := h.validateClosedExam(r.ExamID); msg != "" {
			response.AjaxErr(c, msg)
			return
		}
	}

	now := time.Now()
	updates := map[string]any{
		"update_time": &now,
	}
	if r.Name != "" {
		updates["name"] = r.Name
	}
	if a := toIntPtr(r.Age); a != nil {
		updates["age"] = *a
	}
	if r.Gender != nil {
		updates["gender"] = r.Gender
	}
	if r.Password != "" {
		updates["password"] = r.Password
	}
	if r.Telephone != nil {
		updates["telephone"] = r.Telephone
	}
	if r.Affiliation != nil {
		updates["affiliation"] = r.Affiliation
	}
	if r.Depart != nil {
		updates["depart"] = r.Depart
	}
	if r.Post != nil {
		updates["post"] = r.Post
	}
	if r.Degree != nil {
		updates["degree"] = r.Degree
	}
	if r.Major != nil {
		updates["major"] = r.Major
	}
	if sf := toIntPtr(r.StuFlag); sf != nil {
		updates["stu_flag"] = *sf
	}
	if r.PaperID != nil {
		updates["paper_id"] = *r.PaperID
	}
	if r.ExamID != "" {
		updates["exam_id"] = r.ExamID
	}
	if err := h.db.Table("el_tester").Where("id = ?", r.ID).Updates(updates).Error; err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, true)
}

// DELETE /exam/api/tester/:ids
// 对齐 Java TesterMapper.deleteTesterByIds：物理删除 el_tester
func (h *TesterHandler) Remove(c *gin.Context) {
	ids := c.Param("ids")
	if ids == "" {
		response.AjaxErr(c, "ids 为空")
		return
	}
	if err := h.db.Table("el_tester").Where("id IN (?)", splitCsv(ids)).Delete(nil).Error; err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, true)
}

// DELETE /exam/api/tester/logistic/:ids
// 对齐 Java TesterServiceImpl.deleteLogistic：软删除 el_tester（del_flag=1）
func (h *TesterHandler) Logistic(c *gin.Context) {
	ids := c.Param("ids")
	if ids == "" {
		response.AjaxErr(c, "ids 为空")
		return
	}
	if err := h.db.Table("el_tester").
		Where("id IN (?)", splitCsv(ids)).
		Updates(map[string]any{"del_flag": "1", "update_time": time.Now()}).Error; err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, true)
}

// POST /exam/api/tester/end-time
// 对齐 Java TesterController.updateEndTime：根据 idNumber(+examId) 找到 tester_exam 行，
// 设置 end_time = now。idNumber 兼具身份证号/识别码两种语义（Java 先按 id_number 再按 access_code 查；
// 新 schema 暂不区分 access_code）。
func (h *TesterHandler) EndTime(c *gin.Context) {
	var b struct {
		IDNumber string `json:"idNumber"`
		ExamID   string `json:"examId"`
	}
	if err := c.ShouldBindJSON(&b); err != nil || b.IDNumber == "" {
		response.AjaxErr(c, "没有识别码或身份证号")
		return
	}
	te, err := h.selectTesterByIdentifier(b.IDNumber, b.ExamID)
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	if te == nil {
		response.AjaxErr(c, "没有录入测评者信息")
		return
	}
	now := time.Now()
	te.EndTime = &now
	te.UpdateTime = &now
	if err := h.db.Table("el_tester").Where("id = ?", te.ID).Updates(map[string]any{"end_time": &now, "update_time": &now}).Error; err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, 1)
}

// GET /exam/api/tester/team-score  对齐 Java TesterController.teamScore（Java 实现为空 stub）
func (h *TesterHandler) TeamScore(c *gin.Context) {
	response.AjaxOK(c, nil)
}

// selectTesterByIdentifier 从 el_tester 查：先匹配 telephone，再匹配 id_number/id
func (h *TesterHandler) selectTesterByIdentifier(idNumber, examID string) (*model.Tester, error) {
	if idNumber == "" {
		return nil, nil
	}
	q := h.db.Table("el_tester").
		Where("(telephone = ? OR id_number = ? OR id = ?) AND (del_flag IS NULL OR del_flag = '0')", idNumber, idNumber, idNumber)
	if examID != "" {
		q = q.Where("exam_id = ?", examID)
	}
	var rows []model.Tester
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	// 优先返回 telephone 精确匹配
	for i := range rows {
		if rows[i].Telephone != nil && *rows[i].Telephone == idNumber {
			return &rows[i], nil
		}
	}
	if len(rows) > 1 {
		return nil, &testerAmbigErr{}
	}
	return &rows[0], nil
}

type testerAmbigErr struct{}

func (*testerAmbigErr) Error() string {
	return "同一个测评人员已关联多个测评，请传入examId"
}

func splitCsv(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
