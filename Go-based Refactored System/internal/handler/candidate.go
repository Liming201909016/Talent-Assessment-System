package handler

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// CandidateHandler 对齐 Java CandidateController：/exam/api/candidate/*
// 用于开放考试（is_open=1）的考生管理，与 tester（封闭考试）平行。
type CandidateHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewCandidateHandler(db *gorm.DB, cfg *config.Config) *CandidateHandler {
	return &CandidateHandler{db: db, cfg: cfg}
}

// Candidate 对齐 el_candidate
type Candidate struct {
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	PaperID     *string    `gorm:"column:paper_id"      json:"paperId"`
	ExamID      string     `gorm:"column:exam_id"       json:"examId"`
	Name        string     `gorm:"column:name"          json:"name"`
	Age         *int       `gorm:"column:age"           json:"age"`
	Password    string     `gorm:"column:password"      json:"-"`
	Status      *string    `gorm:"column:status"        json:"status"`
	Telephone   *string    `gorm:"column:telephone"     json:"telephone"`
	Affiliation *string    `gorm:"column:affiliation"   json:"affiliation"`
	Post        *string    `gorm:"column:post"          json:"post"`
	Degree      *string    `gorm:"column:degree"        json:"degree"`
	Major       *string    `gorm:"column:major"         json:"major"`
	StuFlag     *int       `gorm:"column:stu_flag"      json:"stuFlag"`
	Gender      *string    `gorm:"column:gender"        json:"gender"`
	EndTime     *time.Time `gorm:"column:end_time"      json:"endTime"`
	PdfPath     *string    `gorm:"column:pdf_path"      json:"pdfPath"`
	DelFlag     *int       `gorm:"column:del_flag"      json:"delFlag"`
	PdfFlag     *int       `gorm:"column:pdf_flag"      json:"pdfFlag"`
	CreateTime  *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime  *time.Time `gorm:"column:update_time"   json:"updateTime"`
}

func (Candidate) TableName() string { return "el_candidate" }

// POST /exam/api/candidate/save
func (h *CandidateHandler) Save(c *gin.Context) {
	var b struct {
		ID          string      `json:"id"`
		ExamID      string      `json:"examId"`
		Name        string      `json:"name"`
		Age         interface{} `json:"age"`
		Gender      *string     `json:"gender"`
		Telephone   *string     `json:"telephone"`
		Affiliation *string     `json:"affiliation"`
		Post        *string     `json:"post"`
		Degree      *string     `json:"degree"`
		Major       *string     `json:"major"`
		StuFlag     interface{} `json:"stuFlag"`
		IdNumber    *string     `json:"idNumber"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	delZero := 0
	ca := Candidate{
		ExamID:      b.ExamID,
		Name:        b.Name,
		Age:         toIntPtr(b.Age),
		Gender:      b.Gender,
		Telephone:   b.Telephone,
		Affiliation: b.Affiliation,
		Post:        b.Post,
		Degree:      b.Degree,
		Major:       b.Major,
		StuFlag:     toIntPtr(b.StuFlag),
		DelFlag:     &delZero,
	}
	if b.Name == "" {
		response.RestErr(c, "姓名不能为空")
		return
	}
	if b.ExamID == "" || b.Telephone == nil || *b.Telephone == "" {
		response.RestErr(c, "缺少 examId 或 telephone")
		return
	}

	// 检查同手机号是否已完成测评（不可重复答题）
	var completedCount int64
	h.db.Table("el_candidate").
		Where("exam_id = ? AND telephone = ? AND end_time IS NOT NULL AND (del_flag IS NULL OR del_flag = 0)", b.ExamID, *b.Telephone).
		Count(&completedCount)
	if completedCount > 0 && b.ID == "" {
		// 新建时拒绝（如果已有已完成记录）
		response.RestErr(c, "该手机号已完成测评，不可重复答题")
		return
	}

	var existing Candidate
	err := h.db.Where("exam_id = ? AND telephone = ?", b.ExamID, *b.Telephone).First(&existing).Error
	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		ca.ID = strconv.FormatInt(time.Now().UnixMilli(), 10)
		ca.CreateTime = &now
		ca.UpdateTime = &now
		if err := h.db.Create(&ca).Error; err != nil {
			response.RestErr(c, err.Error())
			return
		}
	} else if err != nil {
		response.RestErr(c, err.Error())
		return
	} else {
		ca.ID = existing.ID
		ca.CreateTime = existing.CreateTime // 保留原有 create_time
		if ca.CreateTime == nil {
			ca.CreateTime = &now // 原值为空时用当前时间
		}
		ca.UpdateTime = &now
		if err := h.db.Save(&ca).Error; err != nil {
			response.RestErr(c, err.Error())
			return
		}
	}
	response.Rest(c, ca)
}

// PUT /exam/api/candidate
func (h *CandidateHandler) Update(c *gin.Context) {
	var b struct {
		ID          string      `json:"id"`
		PaperID     *string     `json:"paperId"`
		ExamID      string      `json:"examId"`
		Name        string      `json:"name"`
		Age         interface{} `json:"age"`
		Gender      *string     `json:"gender"`
		Telephone   *string     `json:"telephone"`
		Affiliation *string     `json:"affiliation"`
		Post        *string     `json:"post"`
		Degree      *string     `json:"degree"`
		Major       *string     `json:"major"`
		StuFlag     interface{} `json:"stuFlag"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	now := time.Now()
	updates := map[string]interface{}{"update_time": &now}
	if b.PaperID != nil {
		updates["paper_id"] = *b.PaperID
	}
	if b.Name != "" {
		updates["name"] = b.Name
	}
	if b.Age != nil {
		updates["age"] = toIntPtr(b.Age)
	}
	if b.Gender != nil {
		updates["gender"] = *b.Gender
	}
	if b.Telephone != nil {
		updates["telephone"] = *b.Telephone
	}
	if b.Affiliation != nil {
		updates["affiliation"] = *b.Affiliation
	}
	if b.StuFlag != nil {
		updates["stu_flag"] = toIntPtr(b.StuFlag)
	}
	h.db.Model(&Candidate{}).Where("id = ?", b.ID).Updates(updates)
	response.Rest(c, 1)
}

// DELETE /exam/api/candidate/:ids
func (h *CandidateHandler) Remove(c *gin.Context) {
	ids := splitCsv(c.Param("ids"))
	if len(ids) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	h.db.Where("id IN ?", ids).Delete(&Candidate{})
	response.Rest(c, 1)
}

// DELETE /exam/api/candidate/logistic/:ids
func (h *CandidateHandler) Logistic(c *gin.Context) {
	ids := splitCsv(c.Param("ids"))
	if len(ids) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	h.db.Model(&Candidate{}).Where("id IN ?", ids).Update("del_flag", 1)
	response.Rest(c, 1)
}

// DELETE /exam/api/candidate/logicDeletePdfByIds/:ids
func (h *CandidateHandler) LogicDeletePdfByIds(c *gin.Context) {
	ids := splitCsv(c.Param("ids"))
	if len(ids) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	h.db.Model(&Candidate{}).Where("id IN ?", ids).Update("pdf_flag", 0)
	response.Rest(c, 1)
}

// POST /exam/api/candidate/info
func (h *CandidateHandler) Info(c *gin.Context) {
	var b struct {
		ID string `json:"id"`
	}
	_ = c.ShouldBindJSON(&b)
	var ca Candidate
	if err := h.db.Where("id = ?", b.ID).First(&ca).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	response.Rest(c, ca)
}

// POST /exam/api/candidate/tester-info
func (h *CandidateHandler) TesterInfo(c *gin.Context) {
	var b struct {
		PaperID string `json:"paperId"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	type row struct {
		Candidate
		UserTime   *int       `gorm:"column:user_time"   json:"userTime"`
		CreateTime *time.Time `gorm:"column:create_time" json:"createTime"`
	}
	var r row
	err := h.db.Table("el_candidate AS ca").
		Select("ca.*, pa.create_time, pa.user_time").
		Joins("LEFT JOIN el_paper pa ON pa.id = ca.paper_id").
		Where("ca.paper_id = ?", b.PaperID).
		Take(&r).Error
	if err != nil {
		// 封闭模式考生在 el_tester，兜底查询
		err2 := h.db.Table("el_tester AS ca").
			Select("ca.*, pa.create_time, pa.user_time").
			Joins("LEFT JOIN el_paper pa ON pa.id = ca.paper_id").
			Where("ca.paper_id = ?", b.PaperID).
			Take(&r).Error
		if err2 != nil {
			response.RestErr(c, "不存在")
			return
		}
	}
	response.Rest(c, r)
}

// POST /exam/api/candidate/tester-list
func (h *CandidateHandler) TesterList(c *gin.Context) {
	var b struct {
		Name      string `json:"name"`
		Status    string `json:"status"`
		Telephone string `json:"telephone"`
		ExamID    string `json:"examId"`
	}
	_ = c.ShouldBindJSON(&b)

	type row struct {
		ID          string     `gorm:"column:id"          json:"id"`
		ExamID      string     `gorm:"column:exam_id"     json:"examId"`
		PaperID     *string    `gorm:"column:paper_id"    json:"paperId"`
		RepoCode    *string    `gorm:"column:repo_code"   json:"repoCode"`
		Name        string     `gorm:"column:name"        json:"name"`
		Age         *int       `gorm:"column:age"         json:"age"`
		Telephone   *string    `gorm:"column:telephone"   json:"telephone"`
		Affiliation *string    `gorm:"column:affiliation" json:"affiliation"`
		Gender      *string    `gorm:"column:gender"      json:"gender"`
		Post        *string    `gorm:"column:post"        json:"post"`
		Degree      *string    `gorm:"column:degree"      json:"degree"`
		Major       *string    `gorm:"column:major"       json:"major"`
		StuFlag     *int       `gorm:"column:stu_flag"    json:"stuFlag"`
		PdfPath     *string    `gorm:"column:pdf_path"    json:"pdfPath"`
		PdfFlag     *int       `gorm:"column:pdf_flag"    json:"pdfFlag"`
		DelFlag     *int       `gorm:"column:del_flag"    json:"delFlag"`
		EndTime     *time.Time `gorm:"column:end_time"    json:"endTime"`
		CreateTime  *time.Time `gorm:"column:create_time" json:"createTime"`
		UserTime    *int       `gorm:"column:user_time"   json:"userTime"`
		AnswerNum   int        `gorm:"-"                  json:"answerNum"`
	}
	q := h.db.Table("el_candidate AS ca").
		Select("ca.id, ca.exam_id, ca.paper_id, o.code AS repo_code, ca.name, ca.age, ca.telephone, ca.affiliation, ca.gender, ca.post, ca.degree, ca.major, ca.stu_flag, ca.pdf_path, ca.pdf_flag, ca.del_flag, ca.end_time, pa.create_time, pa.user_time").
		Joins("LEFT JOIN el_paper pa ON pa.id = ca.paper_id").
		Joins("LEFT JOIN el_exam_repo er ON er.exam_id = ca.exam_id").
		Joins("LEFT JOIN el_repo o ON er.repo_id = o.id").
		Where("(ca.del_flag = '0' OR ca.del_flag IS NULL)")
	if b.Name != "" {
		q = q.Where("ca.name like ?", "%"+b.Name+"%")
	}
	if b.Status != "" {
		q = q.Where("ca.status = ?", b.Status)
	}
	if b.Telephone != "" {
		q = q.Where("ca.telephone = ?", b.Telephone)
	}
	if b.ExamID != "" {
		q = q.Where("ca.exam_id = ?", b.ExamID)
	}
	q = q.Order("pa.create_time DESC")
	var rows []row
	q.Scan(&rows)
	for i := range rows {
		if rows[i].PaperID != nil && *rows[i].PaperID != "" {
			var n int64
			if rows[i].RepoCode != nil && strings.HasPrefix(*rows[i].RepoCode, "003") {
				h.db.Table("el_mbti_answer").Where("paper_id = ? AND answered = 1", *rows[i].PaperID).Count(&n)
			} else {
				h.db.Table("el_paper_qu").Where("paper_id = ? AND answered = 1", *rows[i].PaperID).Count(&n)
			}
			rows[i].AnswerNum = int(n)
		}
	}
	response.Rest(c, rows)
}

// POST /exam/api/candidate/stand-score
func (h *CandidateHandler) StandScoreCandidate(c *gin.Context) {
	var b struct {
		PaperID  string `json:"paperId"`
		RepoCode string `json:"repoCode"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	rows, err := queryPaperQuContentForCandidate(h.db, b.PaperID)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	var out map[string]float64
	if strings.HasPrefix(b.RepoCode, "002") {
		out = standScore2(rows)
	} else {
		out = standScore1(rows)
	}
	response.Rest(c, out)
}

// GET /exam/api/candidate/team-score?examId=&isOpen=
func (h *CandidateHandler) TeamScore(c *gin.Context) {
	examID := c.Query("examId")
	isOpenStr := c.DefaultQuery("isOpen", "1")
	isOpen, _ := strconv.Atoi(isOpenStr)
	if examID == "" {
		response.RestErr(c, "examId 为空")
		return
	}
	type idRow struct {
		PaperID  *string    `gorm:"column:paper_id"`
		EndTime  *time.Time `gorm:"column:end_time"`
		RepoCode *string    `gorm:"column:repo_code"`
	}
	var people []idRow
	if isOpen == 1 {
		h.db.Table("el_candidate AS ca").
			Select("ca.paper_id, ca.end_time, o.code AS repo_code").
			Joins("LEFT JOIN el_exam_repo er ON er.exam_id = ca.exam_id").
			Joins("LEFT JOIN el_repo o ON er.repo_id = o.id").
			Where("ca.exam_id = ? AND (ca.del_flag = '0' OR ca.del_flag IS NULL)", examID).
			Scan(&people)
	} else {
		h.db.Table("el_tester AS t").
			Select("t.paper_id, t.end_time, o.code AS repo_code").
			Joins("LEFT JOIN el_exam_repo er ON er.exam_id = t.exam_id").
			Joins("LEFT JOIN el_repo o ON er.repo_id = o.id").
			Where("t.exam_id = ? AND (t.del_flag IS NULL OR t.del_flag = 0)", examID).
			Scan(&people)
	}
	var scores []map[string]float64
	for _, p := range people {
		if p.PaperID == nil || *p.PaperID == "" || p.EndTime == nil {
			continue
		}
		rows, err := queryPaperQuContentForCandidate(h.db, *p.PaperID)
		if err != nil {
			continue
		}
		code := ""
		if p.RepoCode != nil {
			code = *p.RepoCode
		}
		if strings.HasPrefix(code, "002") {
			scores = append(scores, standScore2(rows))
		} else {
			scores = append(scores, standScore1(rows))
		}
	}
	response.Rest(c, scores)
}

// POST /exam/api/candidate/end-time
func (h *CandidateHandler) EndTimeCandidate(c *gin.Context) {
	var b struct {
		PaperID string `json:"paperId"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.PaperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	now := time.Now()
	// 先查 el_candidate
	var ca Candidate
	if err := h.db.Where("paper_id = ?", b.PaperID).First(&ca).Error; err == nil {
		ca.EndTime = &now
		ca.UpdateTime = &now
		h.db.Save(&ca)
		response.Rest(c, 1)
		return
	}
	// 兜底查 el_tester（封闭模式）
	res := h.db.Table("el_tester").Where("paper_id = ?", b.PaperID).
		Updates(map[string]interface{}{"end_time": &now, "update_time": &now})
	if res.RowsAffected > 0 {
		response.Rest(c, 1)
		return
	}
	response.RestErr(c, "不存在")
}

// POST /exam/api/candidate/pdf-persistence
func (h *CandidateHandler) PdfPersistence(c *gin.Context) {
	paperID := c.PostForm("paperId")
	if paperID == "" {
		response.RestErr(c, "paperId 为空")
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		response.RestErr(c, "未上传文件")
		return
	}
	var ca Candidate
	useTester := false
	if err := h.db.Where("paper_id = ?", paperID).First(&ca).Error; err != nil {
		// 兜底查 el_tester（封闭模式）
		if err2 := h.db.Table("el_tester").Where("paper_id = ?", paperID).First(&ca).Error; err2 != nil {
			response.RestErr(c, "不存在")
			return
		}
		useTester = true
	}
	if ca.PdfPath != nil && *ca.PdfPath != "" {
		_ = os.Remove(*ca.PdfPath)
	}
	var name, title string
	tbl := "el_candidate"
	if useTester {
		tbl = "el_tester"
	}
	h.db.Table(tbl+" AS ec").
		Select("ec.name, ep.title").
		Joins("JOIN el_paper AS ep ON ec.paper_id = ep.id").
		Where("ec.paper_id = ?", paperID).
		Row().Scan(&name, &title)

	ts := time.Now().Format("20060102150405000")
	day := time.Now().Format("20060102")
	profile := h.cfg.Upload.Profile
	if profile == "" {
		profile = h.cfg.Upload.Path
	}
	if profile == "" {
		profile = "./tmp"
	}
	pdfDir := filepath.Join(filepath.Dir(profile), "pdf", day)
	_ = os.MkdirAll(pdfDir, 0o755)
	fname := fmt.Sprintf("%s_%s_%s.pdf", name, title, ts)
	saved := filepath.Join(pdfDir, fname)

	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			continue
		}
		dst, err := os.OpenFile(saved, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			src.Close()
			continue
		}
		io.Copy(dst, src)
		src.Close()
		dst.Close()
	}
	one := 1
	now := time.Now()
	ca.PdfPath = &saved
	ca.PdfFlag = &one
	ca.UpdateTime = &now
	if useTester {
		h.db.Table("el_tester").Where("paper_id = ?", paperID).Updates(map[string]interface{}{
			"pdf_path": saved, "pdf_flag": 1, "update_time": &now,
		})
	} else {
		h.db.Save(&ca)
	}
	response.Rest(c, 1)
}

// POST /exam/api/candidate/pdf-upload
// 对齐 Java ExamController.pdfUpload：根据文件路径下载 PDF
func (h *CandidateHandler) PdfUpload(c *gin.Context) {
	var b struct {
		File string `json:"file" form:"file"`
	}
	_ = c.ShouldBind(&b)
	log.Printf("[pdf-upload] file=%q, contentType=%s", b.File, c.ContentType())
	if b.File == "" {
		response.RestErr(c, "file 为空")
		return
	}
	// 安全检查：仅允许服务器上传目录下的文件
	allowedDir := filepath.Clean(h.cfg.Upload.Path)
	if allowedDir == "" || allowedDir == "." {
		allowedDir = filepath.Clean("./tmp")
	}
	clean := filepath.Clean(b.File)
	log.Printf("[pdf-upload] allowedDir=%q, clean=%q, hasPrefix=%v", allowedDir, clean, strings.HasPrefix(clean, allowedDir))
	if !strings.HasPrefix(clean, allowedDir) {
		response.RestErr(c, "文件路径不合法")
		return
	}
	f, err := os.Open(clean)
	if err != nil {
		response.RestErr(c, "文件不存在")
		return
	}
	defer f.Close()
	fname := filepath.Base(clean)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(fname))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	http.ServeContent(c.Writer, c.Request, fname, time.Now(), f)
}

// POST /exam/api/candidate/batch-download
func (h *CandidateHandler) BatchDownload(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if len(ids) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	var paths []string
	h.db.Model(&Candidate{}).Where("id IN ? AND pdf_path IS NOT NULL AND pdf_path != ''", ids).Pluck("pdf_path", &paths)
	if len(paths) == 0 {
		response.RestErr(c, "没有可下载的 PDF")
		return
	}
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=batch-download.zip")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	zw := zip.NewWriter(c.Writer)
	defer zw.Close()
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		w, err := zw.Create(filepath.Base(p))
		if err != nil {
			f.Close()
			continue
		}
		io.Copy(w, f)
		f.Close()
	}
}

// queryPaperQuContentForCandidate 与 tester_score.go 的 queryPaperQuWithContent 结构相同
func queryPaperQuContentForCandidate(db *gorm.DB, paperID string) ([]paperQuContent, error) {
	var rows []paperQuContent
	err := db.Table("el_paper_qu AS pq").
		Select("eq.content AS content, pq.is_right AS is_right, pq.answered AS answered, pq.actual_score AS actual_score").
		Joins("LEFT JOIN el_qu AS eq ON pq.qu_id = eq.id").
		Where("pq.paper_id = ?", paperID).
		Order("pq.sort ASC").
		Find(&rows).Error
	return rows, err
}
