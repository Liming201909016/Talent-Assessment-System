package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/pkg/response"
	"gorm.io/gorm"
)

// ===================== Sys Depart =====================

type SysDepartHandler struct{ db *gorm.DB }

func NewSysDepartHandler(db *gorm.DB) *SysDepartHandler { return &SysDepartHandler{db: db} }

type SysDepart struct {
	ID       string `gorm:"column:dept_id;primaryKey" json:"id"`
	DeptType *int   `gorm:"column:dept_type"          json:"deptType"`
	ParentID string `gorm:"column:parent_id"          json:"parentId"`
	DeptName string `gorm:"column:dept_name"          json:"deptName"`
	DeptCode string `gorm:"column:dept_code"          json:"deptCode"`
	Sort     int    `gorm:"column:sort"               json:"sort"`
}

func (SysDepart) TableName() string { return "sys_depart" }

func (h *SysDepartHandler) Paging(c *gin.Context) {
	var req struct {
		Current int `json:"current"`
		Size    int `json:"size"`
		Params  struct {
			DeptName string `json:"deptName"`
		} `json:"params"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	q := h.db.Model(&SysDepart{}).Where("parent_id = '0'")
	if req.Params.DeptName != "" {
		q = q.Where("dept_name like ?", "%"+req.Params.DeptName+"%")
	}
	var total int64
	q.Count(&total)
	var rows []SysDepart
	q.Order("sort ASC").Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	response.Rest(c, gin.H{"records": rows, "total": total, "current": req.Current, "size": req.Size})
}

func (h *SysDepartHandler) List(c *gin.Context) {
	var rows []SysDepart
	h.db.Order("sort ASC").Find(&rows)
	response.Rest(c, rows)
}

func (h *SysDepartHandler) Detail(c *gin.Context) {
	id := bindID(c)
	var d SysDepart
	if err := h.db.Where("dept_id = ?", id).First(&d).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	response.Rest(c, d)
}

func (h *SysDepartHandler) Save(c *gin.Context) {
	var d SysDepart
	if err := c.ShouldBindJSON(&d); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if d.ID == "" {
		d.ID = strconv.FormatInt(nextID(), 10)
		// 自动 deptCode = 父 code + sort
		if d.DeptCode == "" {
			var parent SysDepart
			if h.db.Where("dept_id = ?", d.ParentID).First(&parent).Error == nil {
				var maxSort int
				h.db.Model(&SysDepart{}).Where("parent_id = ?", d.ParentID).
					Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)
				d.Sort = maxSort + 1
				d.DeptCode = parent.DeptCode + strconv.Itoa(d.Sort)
			}
		}
		h.db.Create(&d)
	} else {
		h.db.Save(&d)
	}
	response.Rest(c, gin.H{"id": d.ID})
}

func (h *SysDepartHandler) Delete(c *gin.Context) {
	var b struct {
		IDs []string `json:"ids"`
	}
	_ = c.ShouldBindJSON(&b)
	if len(b.IDs) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	h.db.Where("dept_id IN ?", b.IDs).Delete(&SysDepart{})
	response.Rest(c, true)
}

func (h *SysDepartHandler) Tree(c *gin.Context) {
	var all []SysDepart
	h.db.Order("sort ASC").Find(&all)

	type node struct {
		SysDepart
		Children []*node `json:"children"`
	}
	nodeMap := map[string]*node{}
	var roots []*node
	for _, d := range all {
		n := &node{SysDepart: d}
		nodeMap[d.ID] = n
	}
	for _, n := range nodeMap {
		if n.ParentID == "0" || n.ParentID == "" {
			roots = append(roots, n)
		} else if p, ok := nodeMap[n.ParentID]; ok {
			p.Children = append(p.Children, n)
		}
	}
	response.Rest(c, roots)
}

func (h *SysDepartHandler) Sort(c *gin.Context) {
	var b struct {
		ID   string `json:"id"`
		Sort int    `json:"sort"` // 0=上移 1=下移
	}
	_ = c.ShouldBindJSON(&b)
	var cur SysDepart
	if err := h.db.Where("dept_id = ?", b.ID).First(&cur).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	var sibling SysDepart
	var err error
	if b.Sort == 0 {
		err = h.db.Where("parent_id = ? AND sort < ?", cur.ParentID, cur.Sort).Order("sort DESC").First(&sibling).Error
	} else {
		err = h.db.Where("parent_id = ? AND sort > ?", cur.ParentID, cur.Sort).Order("sort ASC").First(&sibling).Error
	}
	if err == nil {
		cur.Sort, sibling.Sort = sibling.Sort, cur.Sort
		h.db.Save(&cur)
		h.db.Save(&sibling)
	}
	response.Rest(c, true)
}

// ===================== Sys Role =====================

type SysRoleHandler struct{ db *gorm.DB }

func NewSysRoleHandler(db *gorm.DB) *SysRoleHandler { return &SysRoleHandler{db: db} }

type SysRole struct {
	ID       int64  `gorm:"column:role_id;primaryKey" json:"id"`
	RoleName string `gorm:"column:role_name"          json:"roleName"`
	RoleKey  string `gorm:"column:role_key"           json:"roleKey"`
	RoleSort int    `gorm:"column:role_sort"          json:"roleSort"`
	Status   string `gorm:"column:status"             json:"status"`
}

func (SysRole) TableName() string { return "sys_role" }

func (h *SysRoleHandler) Paging(c *gin.Context) {
	var req struct {
		Current int `json:"current"`
		Size    int `json:"size"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	var total int64
	h.db.Model(&SysRole{}).Count(&total)
	var rows []SysRole
	h.db.Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	response.Rest(c, gin.H{"records": rows, "total": total, "current": req.Current, "size": req.Size})
}

func (h *SysRoleHandler) List(c *gin.Context) {
	var rows []SysRole
	h.db.Find(&rows)
	response.Rest(c, rows)
}

// ===================== Sys Config =====================

type SysConfigHandler struct{ db *gorm.DB }

func NewSysConfigHandler(db *gorm.DB) *SysConfigHandler { return &SysConfigHandler{db: db} }

type SysConfig struct {
	ID        int64  `gorm:"column:config_id;primaryKey" json:"id"`
	SiteName  string `gorm:"column:site_name"             json:"siteName"`
	FrontLogo string `gorm:"column:front_logo"            json:"frontLogo"`
	BackLogo  string `gorm:"column:back_logo"             json:"backLogo"`
	CopyRight string `gorm:"column:copy_right"            json:"copyRight"`
}

func (SysConfig) TableName() string { return "sys_config" }

func (h *SysConfigHandler) Detail(c *gin.Context) {
	var cfg SysConfig
	if err := h.db.First(&cfg).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	response.Rest(c, cfg)
}

func (h *SysConfigHandler) Save(c *gin.Context) {
	var cfg SysConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.RestErr(c, "参数错误")
		return
	}
	if cfg.ID == 0 {
		h.db.Create(&cfg)
	} else {
		h.db.Save(&cfg)
	}
	response.Rest(c, gin.H{"id": cfg.ID})
}

// ===================== Sys User (exam module) =====================

type SysExamUserHandler struct{ db *gorm.DB }

func NewSysExamUserHandler(db *gorm.DB) *SysExamUserHandler { return &SysExamUserHandler{db: db} }

type SysExamUser struct {
	ID         string     `gorm:"column:user_id;primaryKey" json:"id"`
	UserName   string     `gorm:"column:user_name"          json:"userName"`
	RealName   string     `gorm:"column:real_name"          json:"realName"`
	Password   string     `gorm:"column:password"           json:"-"`
	Salt       string     `gorm:"column:salt"               json:"-"`
	RoleIDs    string     `gorm:"column:role_ids"           json:"roleIds"`
	DepartID   string     `gorm:"column:depart_id"          json:"departId"`
	State      *int       `gorm:"column:state"              json:"state"`
	CreateTime *time.Time `gorm:"column:create_time"        json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time"        json:"updateTime"`
}

func (SysExamUser) TableName() string { return "sys_user" }

type SysUserRole struct {
	ID     string `gorm:"column:id;primaryKey" json:"id"`
	UserID string `gorm:"column:user_id"       json:"userId"`
	RoleID string `gorm:"column:role_id"       json:"roleId"`
}

func (SysUserRole) TableName() string { return "sys_user_role" }

func (h *SysExamUserHandler) Paging(c *gin.Context) {
	var req struct {
		Current int `json:"current"`
		Size    int `json:"size"`
		Params  struct {
			UserName string `json:"userName"`
			RealName string `json:"realName"`
		} `json:"params"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	q := h.db.Model(&SysExamUser{})
	if req.Params.UserName != "" {
		q = q.Where("user_name like ?", "%"+req.Params.UserName+"%")
	}
	if req.Params.RealName != "" {
		q = q.Where("real_name like ?", "%"+req.Params.RealName+"%")
	}
	var total int64
	q.Count(&total)
	var rows []SysExamUser
	q.Order("create_time DESC").Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&rows)
	response.Rest(c, gin.H{"records": rows, "total": total, "current": req.Current, "size": req.Size})
}

func (h *SysExamUserHandler) List(c *gin.Context) {
	var rows []SysExamUser
	h.db.Find(&rows)
	response.Rest(c, rows)
}

// RuoYiUserList 对齐 RuoYi 前端 system/user/list 的 GET 请求格式
// 前端用 GET + query params: pageNum, pageSize, userName, phonenumber, status, deptId, beginTime, endTime
func (h *SysExamUserHandler) RuoYiUserList(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageNum <= 0 { pageNum = 1 }
	if pageSize <= 0 { pageSize = 10 }

	userName := c.Query("userName")
	phonenumber := c.Query("phonenumber")
	status := c.Query("status")
	deptId := c.Query("deptId")

	q := h.db.Model(&SysExamUser{})
	if userName != "" {
		q = q.Where("user_name like ?", "%"+userName+"%")
	}
	if phonenumber != "" {
		q = q.Where("real_name like ?", "%"+phonenumber+"%")
	}
	if status != "" {
		q = q.Where("state = ?", status)
	}
	if deptId != "" {
		q = q.Where("depart_id = ?", deptId)
	}
	var total int64
	q.Count(&total)
	var rows []SysExamUser
	q.Order("create_time DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows)

	// RuoYi TableDataInfo 格式
	type userRow struct {
		UserID     string     `json:"userId"`
		UserName   string     `json:"userName"`
		NickName   string     `json:"nickName"`
		Dept       *gin.H     `json:"dept"`
		Phonenumber string    `json:"phonenumber"`
		Status     string     `json:"status"`
		CreateTime *time.Time `json:"createTime"`
	}
	result := make([]userRow, len(rows))
	for i, u := range rows {
		st := "0"
		if u.State != nil && *u.State == 1 { st = "1" }
		result[i] = userRow{
			UserID: u.ID, UserName: u.UserName, NickName: u.RealName,
			Phonenumber: "", Status: st, CreateTime: u.CreateTime,
		}
		if u.DepartID != "" {
			var deptName string
			h.db.Table("sys_dept").Where("dept_id = ?", u.DepartID).Pluck("dept_name", &deptName)
			if deptName != "" {
				d := gin.H{"deptName": deptName}
				result[i].Dept = &d
			}
		}
	}
	response.Table(c, result, total)
}

func (h *SysExamUserHandler) Detail(c *gin.Context) {
	id := bindID(c)
	var u SysExamUser
	if err := h.db.Where("user_id = ?", id).First(&u).Error; err != nil {
		response.RestErr(c, "不存在")
		return
	}
	// 加载 roles
	var roles []SysRole
	h.db.Table("sys_role AS r").
		Joins("INNER JOIN sys_user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", id).
		Find(&roles)
	response.Rest(c, gin.H{"user": u, "roles": roles})
}

func (h *SysExamUserHandler) Save(c *gin.Context) {
	var b struct {
		ID       string   `json:"id"`
		UserName string   `json:"userName"`
		RealName string   `json:"realName"`
		Password string   `json:"password"`
		DepartID string   `json:"departId"`
		Roles    []string `json:"roles"`
	}
	_ = c.ShouldBindJSON(&b)
	now := time.Now()
	if b.ID == "" {
		// create
		id := strconv.FormatInt(nextID(), 10)
		u := SysExamUser{
			ID:         id,
			UserName:   b.UserName,
			RealName:   b.RealName,
			DepartID:   b.DepartID,
			CreateTime: &now,
			UpdateTime: &now,
		}
		if b.Password != "" {
			u.Password = b.Password // 考试模块用户密码与 Java 保持一致，明文存储
		}
		h.db.Create(&u)
		h.saveRoles(id, b.Roles)
		response.Rest(c, gin.H{"id": id})
	} else {
		var u SysExamUser
		if err := h.db.Where("user_id = ?", b.ID).First(&u).Error; err != nil {
			response.RestErr(c, "不存在")
			return
		}
		u.RealName = b.RealName
		u.DepartID = b.DepartID
		if b.Password != "" {
			u.Password = b.Password
		}
		u.UpdateTime = &now
		h.db.Save(&u)
		h.saveRoles(u.ID, b.Roles)
		response.Rest(c, gin.H{"id": u.ID})
	}
}

func (h *SysExamUserHandler) saveRoles(userID string, roleIDs []string) {
	h.db.Where("user_id = ?", userID).Delete(&SysUserRole{})
	for _, rid := range roleIDs {
		h.db.Create(&SysUserRole{
			ID:     strconv.FormatInt(nextID(), 10),
			UserID: userID,
			RoleID: rid,
		})
	}
}

func (h *SysExamUserHandler) Delete(c *gin.Context) {
	var b struct {
		IDs []string `json:"ids"`
	}
	_ = c.ShouldBindJSON(&b)
	if len(b.IDs) == 0 {
		response.RestErr(c, "ids 为空")
		return
	}
	h.db.Where("user_id IN ? AND user_name != 'admin'", b.IDs).Delete(&SysExamUser{})
	h.db.Where("user_id IN ?", b.IDs).Delete(&SysUserRole{})
	response.Rest(c, true)
}

func (h *SysExamUserHandler) State(c *gin.Context) {
	var b struct {
		IDs   []string `json:"ids"`
		State int      `json:"state"`
	}
	_ = c.ShouldBindJSON(&b)
	h.db.Model(&SysExamUser{}).
		Where("user_id IN ? AND user_name != 'admin'", b.IDs).
		Update("state", b.State)
	response.Rest(c, true)
}

func (h *SysExamUserHandler) ResetPwd(c *gin.Context) {
	var b struct {
		ID       string `json:"id"`
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.ID == "" || b.Password == "" {
		response.RestErr(c, "参数错误")
		return
	}
	h.db.Model(&SysExamUser{}).Where("user_id = ?", b.ID).Update("password", b.Password)
	response.Rest(c, true)
}

// POST /exam/api/sys/user/update  对齐 Java：仅更新当前登录用户的密码
func (h *SysExamUserHandler) Update(c *gin.Context) {
	var b struct {
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.Password == "" {
		response.Rest(c, true)
		return
	}
	// 从 context 取 loginUser ID（如不存在则不操作）
	lu, _ := c.Get("loginUser")
	if lu == nil {
		response.RestErr(c, "未登录")
		return
	}
	// loginUser 是 *model.LoginUser，取 UserID
	type hasUserID interface{ GetUserID() int64 }
	if u, ok := lu.(hasUserID); ok {
		h.db.Model(&SysExamUser{}).Where("user_id = ?", u.GetUserID()).Update("password", b.Password)
	}
	response.Rest(c, true)
}

// POST /exam/api/sys/user/reg  对齐 Java：学员注册
func (h *SysExamUserHandler) Reg(c *gin.Context) {
	var b struct {
		UserName string `json:"userName"`
		RealName string `json:"realName"`
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.UserName == "" || b.Password == "" {
		response.RestErr(c, "参数错误")
		return
	}
	var cnt int64
	h.db.Model(&SysExamUser{}).Where("user_name = ?", b.UserName).Count(&cnt)
	if cnt > 0 {
		response.RestErr(c, "用户名已存在，换一个吧！")
		return
	}
	now := time.Now()
	id := strconv.FormatInt(nextID(), 10)
	u := SysExamUser{
		ID:         id,
		UserName:   b.UserName,
		RealName:   b.RealName,
		Password:   b.Password,
		RoleIDs:    "student",
		CreateTime: &now,
		UpdateTime: &now,
	}
	h.db.Create(&u)
	h.saveRoles(id, []string{"student"})
	var roles []string
	h.db.Table("sys_role AS r").
		Joins("INNER JOIN sys_user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", id).
		Pluck("r.role_name", &roles)
	response.Rest(c, gin.H{"id": id, "userName": b.UserName, "realName": b.RealName, "roles": roles})
}

// POST /exam/api/sys/user/quick-reg  对齐 Java：快速注册（存在则直接登录，不存在则注册）
func (h *SysExamUserHandler) QuickReg(c *gin.Context) {
	var b struct {
		UserName string `json:"userName"`
		RealName string `json:"realName"`
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&b)
	if b.UserName == "" {
		response.RestErr(c, "参数错误")
		return
	}
	var u SysExamUser
	if err := h.db.Where("user_name = ?", b.UserName).First(&u).Error; err == nil {
		// 已存在 → 直接返回（与 Java 一致：不校验密码）
		var roles []string
		h.db.Table("sys_role AS r").
			Joins("INNER JOIN sys_user_role ur ON ur.role_id = r.id").
			Where("ur.user_id = ?", u.ID).
			Pluck("r.role_name", &roles)
		response.Rest(c, gin.H{"id": u.ID, "userName": u.UserName, "realName": u.RealName, "roles": roles})
		return
	}
	// 不存在 → 走注册
	h.Reg(c)
}
