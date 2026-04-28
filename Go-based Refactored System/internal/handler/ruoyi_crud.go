package handler

// ruoyi_crud.go — RuoYi 系统管理 CRUD（增删改查详情）
// 补全列表页面的新增/编辑/删除/详情功能

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

// ===================== User CRUD =====================

func (h *RuoYiSystemHandler) UserDetail(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		// GET /system/user/ — 新增表单需要的 posts + roles 下拉
		var roles []sysRole
		h.db.Where("del_flag = '0'").Order("role_sort").Find(&roles)
		var posts []sysPost
		h.db.Order("post_sort").Find(&posts)
		response.AjaxOK(c, gin.H{"roles": roles, "posts": posts})
		return
	}
	var u SysExamUser
	if err := h.db.Where("user_id = ?", userId).First(&u).Error; err != nil {
		response.AjaxErr(c, "用户不存在")
		return
	}
	var roles []sysRole
	h.db.Where("del_flag = '0'").Order("role_sort").Find(&roles)
	var posts []sysPost
	h.db.Order("post_sort").Find(&posts)
	var roleIds []int64
	h.db.Table("sys_user_role").Where("user_id = ?", userId).Pluck("role_id", &roleIds)
	response.AjaxOK(c, gin.H{
		"data": gin.H{
			"userId": u.ID, "userName": u.UserName, "nickName": u.RealName,
			"deptId": u.DepartID, "status": u.State, "roleIds": roleIds,
		},
		"roles": roles, "posts": posts,
	})
}

func (h *RuoYiSystemHandler) UserAdd(c *gin.Context) {
	var b struct {
		UserName string  `json:"userName"`
		NickName string  `json:"nickName"`
		Password string  `json:"password"`
		DeptId   string  `json:"deptId"`
		Status   string  `json:"status"`
		RoleIds  []int64 `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&b); err != nil || b.UserName == "" {
		response.AjaxErr(c, "参数错误")
		return
	}
	// 检查用户名唯一
	var cnt int64
	h.db.Table("sys_user").Where("user_name = ?", b.UserName).Count(&cnt)
	if cnt > 0 {
		response.AjaxErr(c, "用户名已存在")
		return
	}
	pwd := b.Password
	if pwd == "" {
		pwd = "123456"
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)

	now := time.Now()
	st := "0"
	if b.Status == "1" {
		st = "1"
	}
	// sys_user.user_id 是 auto_increment，不需要手动指定
	insertData := map[string]any{
		"user_name": b.UserName, "nick_name": b.NickName, "real_name": b.NickName,
		"password": string(hashed), "status": st,
		"create_time": now, "update_time": now,
	}
	// dept_id 是 bigint，空字符串会导致 MySQL 报错
	if b.DeptId != "" {
		insertData["dept_id"] = b.DeptId
	}
	result := h.db.Table("sys_user").Create(insertData)
	if result.Error != nil {
		response.AjaxErr(c, result.Error.Error())
		return
	}
	// 获取 auto_increment 的 user_id
	var userId int64
	h.db.Table("sys_user").Where("user_name = ?", b.UserName).Pluck("user_id", &userId)
	for _, rid := range b.RoleIds {
		h.db.Table("sys_user_role").Create(map[string]any{
			"user_id": userId, "role_id": rid,
		})
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) UserEdit(c *gin.Context) {
	var b struct {
		UserId   string  `json:"userId"`
		NickName string  `json:"nickName"`
		DeptId   string  `json:"deptId"`
		Status   string  `json:"status"`
		RoleIds  []int64 `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&b); err != nil || b.UserId == "" {
		response.AjaxErr(c, "参数错误")
		return
	}
	now := time.Now()
	updates := map[string]any{"update_time": now}
	if b.NickName != "" {
		updates["real_name"] = b.NickName
	}
	if b.DeptId != "" {
		updates["depart_id"] = b.DeptId
	}
	if b.Status != "" {
		st := 0
		if b.Status == "1" {
			st = 1
		}
		updates["state"] = st
	}
	h.db.Table("sys_user").Where("user_id = ?", b.UserId).Updates(updates)
	if len(b.RoleIds) > 0 {
		h.db.Table("sys_user_role").Where("user_id = ?", b.UserId).Delete(nil)
		for _, rid := range b.RoleIds {
			h.db.Table("sys_user_role").Create(map[string]any{
				"id": strconv.FormatInt(nextID(), 10), "user_id": b.UserId, "role_id": rid,
			})
		}
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) UserDelete(c *gin.Context) {
	ids := strings.Split(c.Param("ids"), ",")
	// FB-026: 保护 admin 账号（user_id=1）不可被删除
	for _, id := range ids {
		if id == "1" {
			response.AjaxErr(c, "不允许删除超级管理员")
			return
		}
	}
	// FB-027: 不允许删除自己
	if lu, ok := c.Get("loginUser"); ok {
		if user, ok := lu.(*model.LoginUser); ok && user != nil {
			currentID := strconv.FormatInt(user.UserID, 10)
			for _, id := range ids {
				if id == currentID {
					response.AjaxErr(c, "不允许删除当前登录用户")
					return
				}
			}
		}
	}
	h.db.Table("sys_user").Where("user_id IN ?", ids).Delete(nil)
	h.db.Table("sys_user_role").Where("user_id IN ?", ids).Delete(nil)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) UserChangeStatus(c *gin.Context) {
	var b struct {
		UserId string `json:"userId"`
		Status string `json:"status"`
	}
	_ = c.ShouldBindJSON(&b)
	// FB-028: 不允许禁用 admin
	if b.UserId == "1" && b.Status == "1" {
		response.AjaxErr(c, "不允许禁用超级管理员")
		return
	}
	st := 0
	if b.Status == "1" {
		st = 1
	}
	h.db.Table("sys_user").Where("user_id = ?", b.UserId).Update("state", st)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) UserResetPwd(c *gin.Context) {
	var b struct {
		UserId   string `json:"userId"`
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&b)
	// FB-029: 限制 admin 密码仅 admin 本人能改
	if b.UserId == "1" {
		if lu, ok := c.Get("loginUser"); ok {
			if user, ok := lu.(*model.LoginUser); ok && user != nil && user.UserID != 1 {
				response.AjaxErr(c, "只有超级管理员本人能修改其密码")
				return
			}
		}
	}
	// FB-030: 拒绝弱密码
	if b.Password == "" {
		b.Password = "Cp@1234"
	}
	if len(b.Password) < 8 {
		response.AjaxErr(c, "密码长度不能少于 8 位")
		return
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(b.Password), bcrypt.DefaultCost)
	h.db.Table("sys_user").Where("user_id = ?", b.UserId).Update("password", string(hashed))
	response.AjaxOK(c, nil)
}

// ===================== Role CRUD =====================

func (h *RuoYiSystemHandler) RoleDetail(c *gin.Context) {
	id := c.Param("roleId")
	var r sysRole
	if err := h.db.Where("role_id = ?", id).First(&r).Error; err != nil {
		response.AjaxErr(c, "角色不存在")
		return
	}
	response.AjaxOK(c, r)
}

func (h *RuoYiSystemHandler) RoleAdd(c *gin.Context) {
	var b struct {
		RoleName string  `json:"roleName"`
		RoleKey  string  `json:"roleKey"`
		RoleSort int     `json:"roleSort"`
		Status   string  `json:"status"`
		MenuIds  []int64 `json:"menuIds"`
	}
	_ = c.ShouldBindJSON(&b)
	now := time.Now()
	id := nextID()
	h.db.Create(&sysRole{RoleID: id, RoleName: b.RoleName, RoleKey: b.RoleKey, RoleSort: b.RoleSort, Status: b.Status, CreateTime: &now})
	for _, mid := range b.MenuIds {
		h.db.Table("sys_role_menu").Create(map[string]any{"role_id": id, "menu_id": mid})
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) RoleEdit(c *gin.Context) {
	var b struct {
		RoleId   int64   `json:"roleId"`
		RoleName string  `json:"roleName"`
		RoleKey  string  `json:"roleKey"`
		RoleSort int     `json:"roleSort"`
		Status   string  `json:"status"`
		MenuIds  []int64 `json:"menuIds"`
	}
	_ = c.ShouldBindJSON(&b)
	h.db.Model(&sysRole{}).Where("role_id = ?", b.RoleId).Updates(map[string]any{
		"role_name": b.RoleName, "role_key": b.RoleKey, "role_sort": b.RoleSort, "status": b.Status,
	})
	if b.MenuIds != nil {
		h.db.Table("sys_role_menu").Where("role_id = ?", b.RoleId).Delete(nil)
		for _, mid := range b.MenuIds {
			h.db.Table("sys_role_menu").Create(map[string]any{"role_id": b.RoleId, "menu_id": mid})
		}
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) RoleDelete(c *gin.Context) {
	ids := strings.Split(c.Param("roleIds"), ",")
	// FB-031: 保护 admin 角色（role_id=1）不可被删除
	for _, id := range ids {
		if id == "1" {
			response.AjaxErr(c, "不允许删除超级管理员角色")
			return
		}
	}
	h.db.Table("sys_role").Where("role_id IN ?", ids).Update("del_flag", "2")
	h.db.Table("sys_role_menu").Where("role_id IN ?", ids).Delete(nil)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) RoleMenuTreeselect(c *gin.Context) {
	roleId := c.Param("roleId")
	var menus []sysMenu
	h.db.Where("status = '0'").Order("parent_id, order_num").Find(&menus)
	tree := buildMenuTree(menus, 0)
	var checkedKeys []int64
	h.db.Table("sys_role_menu").Where("role_id = ?", roleId).Pluck("menu_id", &checkedKeys)
	response.AjaxOK(c, gin.H{"menus": tree, "checkedKeys": checkedKeys})
}

// ===================== Menu CRUD =====================

func (h *RuoYiSystemHandler) MenuDetail(c *gin.Context) {
	id := c.Param("menuId")
	var m sysMenu
	if err := h.db.Where("menu_id = ?", id).First(&m).Error; err != nil {
		response.AjaxErr(c, "菜单不存在")
		return
	}
	response.AjaxOK(c, m)
}

func (h *RuoYiSystemHandler) MenuAdd(c *gin.Context) {
	var m sysMenu
	_ = c.ShouldBindJSON(&m)
	m.MenuID = nextID()
	now := time.Now()
	m.CreateTime = &now
	h.db.Create(&m)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) MenuEdit(c *gin.Context) {
	var m sysMenu
	_ = c.ShouldBindJSON(&m)
	h.db.Model(&sysMenu{}).Where("menu_id = ?", m.MenuID).Updates(map[string]any{
		"menu_name": m.MenuName, "parent_id": m.ParentID, "order_num": m.OrderNum,
		"path": m.Path, "component": m.Component, "menu_type": m.MenuType,
		"visible": m.Visible, "status": m.Status, "perms": m.Perms, "icon": m.Icon,
	})
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) MenuDelete(c *gin.Context) {
	id := c.Param("menuId")
	h.db.Where("menu_id = ?", id).Delete(&sysMenu{})
	response.AjaxOK(c, nil)
}

// ===================== Dept CRUD =====================

func (h *RuoYiSystemHandler) DeptDetail(c *gin.Context) {
	id := c.Param("deptId")
	var d sysDept
	if err := h.db.Where("dept_id = ?", id).First(&d).Error; err != nil {
		response.AjaxErr(c, "部门不存在")
		return
	}
	response.AjaxOK(c, d)
}

func (h *RuoYiSystemHandler) DeptExcludeChild(c *gin.Context) {
	excludeId := c.Param("deptId")
	var depts []sysDept
	h.db.Where("del_flag = '0' AND dept_id != ?", excludeId).Order("parent_id, order_num").Find(&depts)
	response.AjaxOK(c, depts)
}

func (h *RuoYiSystemHandler) DeptAdd(c *gin.Context) {
	var d sysDept
	_ = c.ShouldBindJSON(&d)
	d.DeptID = nextID()
	now := time.Now()
	d.CreateTime = &now
	h.db.Create(&d)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) DeptEdit(c *gin.Context) {
	var d sysDept
	_ = c.ShouldBindJSON(&d)
	h.db.Model(&sysDept{}).Where("dept_id = ?", d.DeptID).Updates(map[string]any{
		"dept_name": d.DeptName, "parent_id": d.ParentID, "order_num": d.OrderNum,
		"leader": d.Leader, "phone": d.Phone, "email": d.Email, "status": d.Status,
	})
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) DeptDelete(c *gin.Context) {
	id := c.Param("deptId")
	h.db.Table("sys_dept").Where("dept_id = ?", id).Update("del_flag", "2")
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) DeptRoleTreeselect(c *gin.Context) {
	roleId := c.Param("roleId")
	var depts []sysDept
	h.db.Where("del_flag = '0' AND status = '0'").Order("parent_id, order_num").Find(&depts)
	tree := buildDeptTree(depts, 0)
	var checkedKeys []int64
	h.db.Table("sys_role_dept").Where("role_id = ?", roleId).Pluck("dept_id", &checkedKeys)
	response.AjaxOK(c, gin.H{"depts": tree, "checkedKeys": checkedKeys})
}

// ===================== Config CRUD =====================

func (h *RuoYiSystemHandler) ConfigDetail(c *gin.Context) {
	id := c.Param("configId")
	var cfg sysConfig
	if err := h.db.Where("config_id = ?", id).First(&cfg).Error; err != nil {
		response.AjaxErr(c, "参数不存在")
		return
	}
	response.AjaxOK(c, cfg)
}

func (h *RuoYiSystemHandler) ConfigAdd(c *gin.Context) {
	var cfg sysConfig
	_ = c.ShouldBindJSON(&cfg)
	cfg.ConfigID = nextID()
	now := time.Now()
	cfg.CreateTime = &now
	h.db.Create(&cfg)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) ConfigEdit(c *gin.Context) {
	var cfg sysConfig
	_ = c.ShouldBindJSON(&cfg)
	h.db.Model(&sysConfig{}).Where("config_id = ?", cfg.ConfigID).Updates(map[string]any{
		"config_name": cfg.ConfigName, "config_key": cfg.ConfigKey,
		"config_value": cfg.ConfigValue, "config_type": cfg.ConfigType, "remark": cfg.Remark,
	})
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) ConfigDelete(c *gin.Context) {
	ids := strings.Split(c.Param("configIds"), ",")
	h.db.Where("config_id IN ?", ids).Delete(&sysConfig{})
	response.AjaxOK(c, nil)
}

// ===================== Dict Type CRUD =====================

func (h *RuoYiSystemHandler) DictTypeDetail(c *gin.Context) {
	id := c.Param("dictId")
	var dt sysDictType
	if err := h.db.Where("dict_id = ?", id).First(&dt).Error; err != nil {
		response.AjaxErr(c, "字典类型不存在")
		return
	}
	response.AjaxOK(c, dt)
}

func (h *RuoYiSystemHandler) DictTypeAdd(c *gin.Context) {
	var dt sysDictType
	_ = c.ShouldBindJSON(&dt)
	dt.DictID = nextID()
	now := time.Now()
	dt.CreateTime = &now
	h.db.Create(&dt)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) DictTypeEdit(c *gin.Context) {
	var dt sysDictType
	_ = c.ShouldBindJSON(&dt)
	h.db.Model(&sysDictType{}).Where("dict_id = ?", dt.DictID).Updates(map[string]any{
		"dict_name": dt.DictName, "dict_type": dt.DictType, "status": dt.Status, "remark": dt.Remark,
	})
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) DictTypeDelete(c *gin.Context) {
	ids := strings.Split(c.Param("dictIds"), ",")
	h.db.Where("dict_id IN ?", ids).Delete(&sysDictType{})
	response.AjaxOK(c, nil)
}

// ===================== Dict Data CRUD =====================

func (h *RuoYiSystemHandler) DictDataDetail(c *gin.Context) {
	code := c.Param("dictCode")
	type row struct {
		DictCode  int64  `gorm:"column:dict_code" json:"dictCode"`
		DictSort  int    `gorm:"column:dict_sort" json:"dictSort"`
		DictLabel string `gorm:"column:dict_label" json:"dictLabel"`
		DictValue string `gorm:"column:dict_value" json:"dictValue"`
		DictType  string `gorm:"column:dict_type" json:"dictType"`
		Status    string `gorm:"column:status" json:"status"`
		Remark    string `gorm:"column:remark" json:"remark"`
	}
	var r row
	if err := h.db.Table("sys_dict_data").Where("dict_code = ?", code).Scan(&r).Error; err != nil {
		response.AjaxErr(c, "字典数据不存在")
		return
	}
	response.AjaxOK(c, r)
}

func (h *RuoYiSystemHandler) DictDataAdd(c *gin.Context) {
	var b struct {
		DictSort  int    `json:"dictSort"`
		DictLabel string `json:"dictLabel"`
		DictValue string `json:"dictValue"`
		DictType  string `json:"dictType"`
		Status    string `json:"status"`
		Remark    string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&b)
	now := time.Now()
	h.db.Table("sys_dict_data").Create(map[string]any{
		"dict_sort": b.DictSort, "dict_label": b.DictLabel, "dict_value": b.DictValue,
		"dict_type": b.DictType, "status": b.Status, "remark": b.Remark, "create_time": now,
	})
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) DictDataEdit(c *gin.Context) {
	var b struct {
		DictCode  int64  `json:"dictCode"`
		DictSort  int    `json:"dictSort"`
		DictLabel string `json:"dictLabel"`
		DictValue string `json:"dictValue"`
		DictType  string `json:"dictType"`
		Status    string `json:"status"`
		Remark    string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&b)
	h.db.Table("sys_dict_data").Where("dict_code = ?", b.DictCode).Updates(map[string]any{
		"dict_sort": b.DictSort, "dict_label": b.DictLabel, "dict_value": b.DictValue,
		"dict_type": b.DictType, "status": b.Status, "remark": b.Remark,
	})
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) DictDataDelete(c *gin.Context) {
	codes := strings.Split(c.Param("dictCodes"), ",")
	h.db.Table("sys_dict_data").Where("dict_code IN ?", codes).Delete(nil)
	response.AjaxOK(c, nil)
}

// ===================== Post CRUD =====================

func (h *RuoYiSystemHandler) PostDetail(c *gin.Context) {
	id := c.Param("postId")
	var p sysPost
	if err := h.db.Where("post_id = ?", id).First(&p).Error; err != nil {
		response.AjaxErr(c, "岗位不存在")
		return
	}
	response.AjaxOK(c, p)
}

func (h *RuoYiSystemHandler) PostAdd(c *gin.Context) {
	var p sysPost
	_ = c.ShouldBindJSON(&p)
	p.PostID = nextID()
	now := time.Now()
	p.CreateTime = &now
	h.db.Create(&p)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) PostEdit(c *gin.Context) {
	var p sysPost
	_ = c.ShouldBindJSON(&p)
	h.db.Model(&sysPost{}).Where("post_id = ?", p.PostID).Updates(map[string]any{
		"post_code": p.PostCode, "post_name": p.PostName, "post_sort": p.PostSort, "status": p.Status,
	})
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) PostDelete(c *gin.Context) {
	ids := strings.Split(c.Param("postIds"), ",")
	h.db.Where("post_id IN ?", ids).Delete(&sysPost{})
	response.AjaxOK(c, nil)
}

// ===================== Notice CRUD =====================

func (h *RuoYiSystemHandler) NoticeDetail(c *gin.Context) {
	id := c.Param("noticeId")
	var n sysNotice
	if err := h.db.Where("notice_id = ?", id).First(&n).Error; err != nil {
		response.AjaxErr(c, "公告不存在")
		return
	}
	response.AjaxOK(c, n)
}

func (h *RuoYiSystemHandler) NoticeAdd(c *gin.Context) {
	var n sysNotice
	_ = c.ShouldBindJSON(&n)
	n.NoticeID = nextID()
	now := time.Now()
	n.CreateTime = &now
	h.db.Create(&n)
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) NoticeEdit(c *gin.Context) {
	var n sysNotice
	_ = c.ShouldBindJSON(&n)
	h.db.Model(&sysNotice{}).Where("notice_id = ?", n.NoticeID).Updates(map[string]any{
		"notice_title": n.NoticeTitle, "notice_type": n.NoticeType,
		"notice_content": n.NoticeContent, "status": n.Status,
	})
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) NoticeDelete(c *gin.Context) {
	ids := strings.Split(c.Param("noticeIds"), ",")
	h.db.Where("notice_id IN ?", ids).Delete(&sysNotice{})
	response.AjaxOK(c, nil)
}
