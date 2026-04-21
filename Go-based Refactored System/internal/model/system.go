package model

import "time"

// SysUser 对齐表 sys_user（列名已在 tag 中声明）
type SysUser struct {
	UserID      int64      `gorm:"column:user_id;primaryKey;autoIncrement" json:"userId"`
	DeptID      *int64     `gorm:"column:dept_id"                           json:"deptId"`
	UserName    string     `gorm:"column:user_name"                         json:"userName"`
	NickName    string     `gorm:"column:nick_name"                         json:"nickName"`
	Email       string     `gorm:"column:email"                             json:"email"`
	Phonenumber string     `gorm:"column:phonenumber"                       json:"phonenumber"`
	Sex         string     `gorm:"column:sex"                               json:"sex"`
	Avatar      string     `gorm:"column:avatar"                            json:"avatar"`
	Password    string     `gorm:"column:password"                          json:"-"`
	Status      string     `gorm:"column:status"                            json:"status"`
	DelFlag     string     `gorm:"column:del_flag"                          json:"delFlag"`
	LoginIP     string     `gorm:"column:login_ip"                          json:"loginIp"`
	LoginDate   *time.Time `gorm:"column:login_date"                        json:"loginDate"`
	CreateBy    string     `gorm:"column:create_by"                         json:"createBy"`
	CreateTime  *time.Time `gorm:"column:create_time"                       json:"createTime"`
	UpdateBy    string     `gorm:"column:update_by"                         json:"updateBy"`
	UpdateTime  *time.Time `gorm:"column:update_time"                       json:"updateTime"`
	Remark      string     `gorm:"column:remark"                            json:"remark"`
	Admin       bool       `gorm:"-"                                        json:"admin"`
}

func (SysUser) TableName() string { return "sys_user" }

// SysRole 对齐表 sys_role
type SysRole struct {
	RoleID    int64  `gorm:"column:role_id;primaryKey" json:"roleId"`
	RoleName  string `gorm:"column:role_name"          json:"roleName"`
	RoleKey   string `gorm:"column:role_key"           json:"roleKey"`
	RoleSort  int    `gorm:"column:role_sort"          json:"roleSort"`
	DataScope string `gorm:"column:data_scope"         json:"dataScope"`
	Status    string `gorm:"column:status"             json:"status"`
	DelFlag   string `gorm:"column:del_flag"           json:"delFlag"`
}

func (SysRole) TableName() string { return "sys_role" }

// SysMenu 对齐表 sys_menu
type SysMenu struct {
	MenuID     int64   `gorm:"column:menu_id;primaryKey" json:"menuId"`
	MenuName   string  `gorm:"column:menu_name"          json:"menuName"`
	ParentID   int64   `gorm:"column:parent_id"          json:"parentId"`
	OrderNum   int     `gorm:"column:order_num"          json:"orderNum"`
	Path       string  `gorm:"column:path"               json:"path"`
	Component  string  `gorm:"column:component"          json:"component"`
	Query      string  `gorm:"column:query"              json:"query"`
	IsFrame    int     `gorm:"column:is_frame"           json:"isFrame"`
	IsCache    int     `gorm:"column:is_cache"           json:"isCache"`
	MenuType   string  `gorm:"column:menu_type"          json:"menuType"`
	Visible    string  `gorm:"column:visible"            json:"visible"`
	Status     string  `gorm:"column:status"             json:"status"`
	Perms      string  `gorm:"column:perms"              json:"perms"`
	Icon       string  `gorm:"column:icon"               json:"icon"`
	CreateBy   string  `gorm:"column:create_by"          json:"createBy"`
	Remark     string  `gorm:"column:remark"             json:"remark"`
	Children   []*SysMenu `gorm:"-" json:"children,omitempty"`
}

func (SysMenu) TableName() string { return "sys_menu" }

// LoginUser 登录上下文（写入 Redis 的值，同时挂到 gin.Context）
type LoginUser struct {
	UserID      int64    `json:"userId"`
	DeptID      *int64   `json:"deptId"`
	Token       string   `json:"token"`
	LoginTime   int64    `json:"loginTime"`
	ExpireTime  int64    `json:"expireTime"`
	IPAddr      string   `json:"ipaddr"`
	Browser     string   `json:"browser"`
	OS          string   `json:"os"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	User        *SysUser `json:"user"`
}
