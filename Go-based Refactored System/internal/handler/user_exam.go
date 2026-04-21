package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// UserExamHandler 对应 /exam/api/user/exam/*
type UserExamHandler struct{ db *gorm.DB }

func NewUserExamHandler(db *gorm.DB) *UserExamHandler { return &UserExamHandler{db: db} }

type userExamPagingReq struct {
	Current int    `json:"current"`
	Size    int    `json:"size"`
	UserID  string `json:"userId"`
	ExamID  string `json:"examId"`
}

// POST /exam/api/user/exam/paging
func (h *UserExamHandler) Paging(c *gin.Context) {
	var req userExamPagingReq
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	q := h.db.Model(&model.UserExam{})
	if req.UserID != "" {
		q = q.Where("user_id = ?", req.UserID)
	}
	if req.ExamID != "" {
		q = q.Where("exam_id = ?", req.ExamID)
	}
	var total int64
	q.Count(&total)
	var rows []model.UserExam
	q.Order("update_time desc").Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	response.Rest(c, gin.H{"records": rows, "total": total, "current": req.Current, "size": req.Size})
}

// POST /exam/api/user/exam/my-paging — 当前登录用户的考试记录
func (h *UserExamHandler) MyPaging(c *gin.Context) {
	var req userExamPagingReq
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	// 当前用户从中间件存入的 LoginUser（sys_user）读取
	userID := ""
	if v, ok := c.Get("loginUser"); ok {
		if lu, ok := v.(*model.LoginUser); ok && lu.User != nil {
			userID = itoaInt64(lu.User.UserID)
		}
	}
	if userID == "" {
		userID = req.UserID
	}
	q := h.db.Model(&model.UserExam{}).Where("user_id = ?", userID)
	var total int64
	q.Count(&total)
	var rows []model.UserExam
	q.Order("update_time desc").Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	response.Rest(c, gin.H{"records": rows, "total": total, "current": req.Current, "size": req.Size})
}

func itoaInt64(v int64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}
