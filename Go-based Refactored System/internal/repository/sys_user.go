package repository

import (
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

type SysUserRepo struct{ db *gorm.DB }

func NewSysUserRepo(db *gorm.DB) *SysUserRepo { return &SysUserRepo{db: db} }

func (r *SysUserRepo) FindByUsername(username string) (*model.SysUser, error) {
	var u model.SysUser
	if err := r.db.Where("user_name = ? AND del_flag = '0'", username).First(&u).Error; err != nil {
		return nil, err
	}
	u.Admin = u.UserID == 1
	return &u, nil
}

func (r *SysUserRepo) FindByID(id int64) (*model.SysUser, error) {
	var u model.SysUser
	if err := r.db.Where("user_id = ? AND del_flag = '0'", id).First(&u).Error; err != nil {
		return nil, err
	}
	u.Admin = u.UserID == 1
	return &u, nil
}

func (r *SysUserRepo) UpdateLogin(id int64, ip string) error {
	return r.db.Table("sys_user").Where("user_id = ?", id).
		Updates(map[string]any{"login_ip": ip, "login_date": gorm.Expr("NOW()")}).Error
}

// GetRoleKeys 取用户所拥有的 role.role_key 集合
func (r *SysUserRepo) GetRoleKeys(userID int64) ([]string, error) {
	var keys []string
	if userID == 1 {
		return []string{"admin"}, nil
	}
	err := r.db.Raw(`
		SELECT r.role_key FROM sys_role r
		JOIN sys_user_role ur ON ur.role_id = r.role_id
		WHERE r.status = '0' AND r.del_flag = '0' AND ur.user_id = ?`, userID).Scan(&keys).Error
	return keys, err
}

// GetMenuPerms 菜单权限集合（perms 字段）
func (r *SysUserRepo) GetMenuPerms(userID int64) ([]string, error) {
	var perms []string
	if userID == 1 {
		return []string{"*:*:*"}, nil
	}
	err := r.db.Raw(`
		SELECT DISTINCT m.perms FROM sys_menu m
		LEFT JOIN sys_role_menu rm ON m.menu_id = rm.menu_id
		LEFT JOIN sys_user_role ur ON rm.role_id = ur.role_id
		LEFT JOIN sys_role r ON r.role_id = ur.role_id
		WHERE m.status = '0' AND r.status = '0' AND ur.user_id = ? AND m.perms <> ''`, userID).Scan(&perms).Error
	return perms, err
}

type SysMenuRepo struct{ db *gorm.DB }

func NewSysMenuRepo(db *gorm.DB) *SysMenuRepo { return &SysMenuRepo{db: db} }

// SelectMenuTreeByUserID 返回菜单列表（不含 F 按钮类），按 parent_id,order_num 排序。
func (r *SysMenuRepo) SelectMenuTreeByUserID(userID int64) ([]*model.SysMenu, error) {
	var menus []*model.SysMenu
	if userID == 1 {
		err := r.db.Where("menu_type IN ('M','C') AND status = '0'").
			Order("parent_id, order_num").Find(&menus).Error
		return menus, err
	}
	err := r.db.Raw(`
		SELECT DISTINCT m.* FROM sys_menu m
		LEFT JOIN sys_role_menu rm ON m.menu_id = rm.menu_id
		LEFT JOIN sys_user_role ur ON rm.role_id = ur.role_id
		LEFT JOIN sys_role r ON r.role_id = ur.role_id
		WHERE ur.user_id = ? AND m.menu_type IN ('M','C') AND m.status = '0' AND r.status = '0'
		ORDER BY m.parent_id, m.order_num`, userID).Scan(&menus).Error
	return menus, err
}
