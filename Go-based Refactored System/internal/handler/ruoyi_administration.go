package handler

import (
	"bytes"
	"context"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/pkg/redisx"
	"github.com/talent-assessment/refactored/pkg/response"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const maxAvatarBytes = 2 << 20

var (
	phonePattern    = regexp.MustCompile(`^1[3-9][0-9]{9}$`)
	usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{2,20}$`)
)

type userProfileUpdateRequest struct {
	NickName    string `json:"nickName"`
	Email       string `json:"email"`
	Phonenumber string `json:"phonenumber"`
	Sex         string `json:"sex"`
}

type registerRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	Code            string `json:"code"`
	UUID            string `json:"uuid"`
}

type userRoleRelation struct {
	UserID int64 `gorm:"column:user_id"`
	RoleID int64 `gorm:"column:role_id"`
}

func (userRoleRelation) TableName() string { return "sys_user_role" }

type roleDeptRelation struct {
	RoleID int64 `gorm:"column:role_id"`
	DeptID int64 `gorm:"column:dept_id"`
}

func (roleDeptRelation) TableName() string { return "sys_role_dept" }

func currentLoginUser(c *gin.Context) (*model.LoginUser, bool) {
	value, ok := c.Get("loginUser")
	if !ok {
		return nil, false
	}
	login, ok := value.(*model.LoginUser)
	return login, ok && login != nil
}

func hasSystemPermission(login *model.LoginUser, permission string) bool {
	if login == nil {
		return false
	}
	if login.UserID == 1 {
		return true
	}
	for _, role := range login.Roles {
		if role == "admin" {
			return true
		}
	}
	for _, candidate := range login.Permissions {
		if candidate == "*:*:*" || candidate == permission {
			return true
		}
	}
	return false
}

func requireSystemPermission(c *gin.Context, permission string) (*model.LoginUser, bool) {
	login, ok := currentLoginUser(c)
	if !ok || !hasSystemPermission(login, permission) {
		response.AjaxForbidden(c, "没有权限执行该操作")
		return nil, false
	}
	return login, true
}

func validateUserProfileUpdate(req userProfileUpdateRequest) error {
	req.NickName = strings.TrimSpace(req.NickName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phonenumber = strings.TrimSpace(req.Phonenumber)
	if req.NickName == "" || utf8.RuneCountInString(req.NickName) > 30 {
		return errors.New("用户昵称不能为空且不能超过30个字符")
	}
	for _, r := range req.NickName {
		if unicode.IsControl(r) {
			return errors.New("用户昵称包含非法字符")
		}
	}
	address, err := mail.ParseAddress(req.Email)
	if err != nil || address.Address != req.Email || len(req.Email) > 50 || strings.ContainsAny(req.Email, "\r\n") {
		return errors.New("邮箱格式不正确")
	}
	parts := strings.Split(req.Email, "@")
	if len(parts) != 2 || parts[0] == "" || !strings.Contains(parts[1], ".") {
		return errors.New("邮箱格式不正确")
	}
	if !phonePattern.MatchString(req.Phonenumber) {
		return errors.New("手机号码格式不正确")
	}
	if req.Sex != "0" && req.Sex != "1" && req.Sex != "2" {
		return errors.New("性别参数不正确")
	}
	return nil
}

func validateStrongPassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return errors.New("密码长度必须为8至72位")
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("密码必须同时包含字母和数字")
	}
	return nil
}

func validateRegisterRequest(req registerRequest) error {
	if !usernamePattern.MatchString(req.Username) {
		return errors.New("用户名只能包含字母、数字和下划线，长度为2至20位")
	}
	if err := validateStrongPassword(req.Password); err != nil {
		return err
	}
	if req.Password != req.ConfirmPassword {
		return errors.New("两次输入的密码不一致")
	}
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.UUID) == "" {
		return errors.New("验证码不能为空")
	}
	return nil
}

func parseUniquePositiveIDs(raw string) ([]int64, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("ID不能为空")
	}
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, errors.New("ID格式不正确")
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			return nil, errors.New("ID格式不正确")
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func parseOptionalUniquePositiveIDs(raw string) ([]int64, error) {
	if strings.TrimSpace(raw) == "" {
		return make([]int64, 0), nil
	}
	return parseUniquePositiveIDs(raw)
}

func avatarDestination(uploadRoot, profileURL, filename string) (string, string, error) {
	if uploadRoot == "" || profileURL == "" || filepath.Base(filename) != filename {
		return "", "", errors.New("头像存储配置不正确")
	}
	parsed, err := url.Parse(profileURL)
	if err != nil || parsed.IsAbs() || parsed.RawQuery != "" || parsed.Fragment != "" || !strings.HasPrefix(parsed.Path, "/") {
		return "", "", errors.New("头像访问路径配置不正确")
	}
	for _, segment := range strings.Split(filepath.ToSlash(parsed.Path), "/") {
		if segment == ".." || segment == "." {
			return "", "", errors.New("头像访问路径配置不正确")
		}
	}
	cleanURL := filepath.ToSlash(filepath.Clean(parsed.Path))
	if cleanURL == "/" || strings.Contains(cleanURL, "..") {
		return "", "", errors.New("头像访问路径配置不正确")
	}
	absoluteRoot, err := filepath.Abs(uploadRoot)
	if err != nil {
		return "", "", errors.New("头像存储配置不正确")
	}
	profileRoot := filepath.Join(absoluteRoot, filepath.FromSlash(strings.TrimPrefix(cleanURL, "/")))
	destination := filepath.Join(profileRoot, "avatar", filename)
	rel, err := filepath.Rel(profileRoot, destination)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", errors.New("头像保存路径不安全")
	}
	return destination, cleanURL + "/avatar/" + filename, nil
}

func randomAvatarName(extension string) (string, error) {
	buffer := make([]byte, 16)
	if _, err := cryptoRand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer) + extension, nil
}

func (h *RuoYiSystemHandler) UserProfileUpdate(c *gin.Context) {
	login, ok := currentLoginUser(c)
	if !ok {
		response.AjaxUnauthorized(c, "未登录")
		return
	}
	var req userProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AjaxErr(c, "参数格式错误")
		return
	}
	req.NickName = strings.TrimSpace(req.NickName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phonenumber = strings.TrimSpace(req.Phonenumber)
	if err := validateUserProfileUpdate(req); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	var user model.SysUser
	if err := h.db.Where("user_id = ? AND del_flag = '0'", login.UserID).First(&user).Error; err != nil {
		response.AjaxErr(c, "当前用户不存在")
		return
	}
	var duplicate int64
	if err := h.db.Model(&model.SysUser{}).Where("user_id <> ? AND del_flag = '0' AND email = ?", login.UserID, req.Email).Count(&duplicate).Error; err != nil {
		response.AjaxErr(c, "校验邮箱失败")
		return
	}
	if duplicate > 0 {
		response.AjaxErr(c, "邮箱已被其他用户使用")
		return
	}
	if err := h.db.Model(&model.SysUser{}).Where("user_id <> ? AND del_flag = '0' AND phonenumber = ?", login.UserID, req.Phonenumber).Count(&duplicate).Error; err != nil {
		response.AjaxErr(c, "校验手机号码失败")
		return
	}
	if duplicate > 0 {
		response.AjaxErr(c, "手机号码已被其他用户使用")
		return
	}
	result := h.db.Model(&model.SysUser{}).Where("user_id = ? AND del_flag = '0'", login.UserID).Updates(map[string]any{
		"nick_name": req.NickName, "email": req.Email, "phonenumber": req.Phonenumber,
		"sex": req.Sex, "update_time": time.Now(),
	})
	if result.Error != nil || result.RowsAffected != 1 {
		response.AjaxErr(c, "修改个人信息失败")
		return
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) UserProfileUpdatePwd(c *gin.Context) {
	login, ok := currentLoginUser(c)
	if !ok {
		response.AjaxUnauthorized(c, "未登录")
		return
	}
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AjaxErr(c, "参数格式错误")
		return
	}
	oldPassword := req.OldPassword
	newPassword := req.NewPassword
	if oldPassword == "" || newPassword == "" {
		response.AjaxErr(c, "旧密码和新密码不能为空")
		return
	}
	if oldPassword == newPassword {
		response.AjaxErr(c, "新密码不能与旧密码相同")
		return
	}
	if err := validateStrongPassword(newPassword); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	var user model.SysUser
	if err := h.db.Where("user_id = ? AND del_flag = '0'", login.UserID).First(&user).Error; err != nil {
		response.AjaxErr(c, "当前用户不存在")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)) != nil {
		response.AjaxErr(c, "旧密码不正确")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		response.AjaxErr(c, "密码加密失败")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.SysUser{}).Where("user_id = ? AND del_flag = '0'", login.UserID).
			Updates(map[string]any{"password": string(hash), "update_time": time.Now()})
		if result.Error != nil || result.RowsAffected != 1 {
			return errors.New("更新密码失败")
		}
		return invalidateUserSessions(c.Request.Context(), login.UserID)
	}); err != nil {
		response.AjaxErr(c, "修改密码失败")
		return
	}
	response.AjaxOK(c, nil)
}

func invalidateUserSessions(ctx context.Context, userID int64) error {
	if redisx.Client == nil {
		return errors.New("Redis不可用")
	}
	var cursor uint64
	for {
		keys, next, err := redisx.Client.Scan(ctx, cursor, redisx.LoginTokenKey+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			values, err := redisx.Client.MGet(ctx, keys...).Result()
			if err != nil {
				return err
			}
			toDelete := make([]string, 0)
			for i, value := range values {
				raw, ok := value.(string)
				if !ok {
					continue
				}
				var login model.LoginUser
				if json.Unmarshal([]byte(raw), &login) == nil && login.UserID == userID {
					toDelete = append(toDelete, keys[i])
				}
			}
			if len(toDelete) > 0 {
				if err := redisx.Client.Del(ctx, toDelete...).Err(); err != nil {
					return err
				}
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func (h *RuoYiSystemHandler) UserProfileAvatar(c *gin.Context) {
	login, ok := currentLoginUser(c)
	if !ok {
		response.AjaxUnauthorized(c, "未登录")
		return
	}
	if h.cfg == nil {
		response.AjaxErr(c, "头像存储配置不可用")
		return
	}
	fileHeader, err := c.FormFile("avatarfile")
	if err != nil || fileHeader.Size <= 0 {
		response.AjaxErr(c, "请选择头像文件")
		return
	}
	if fileHeader.Size > maxAvatarBytes {
		response.AjaxErr(c, "头像文件不能超过2MiB")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.AjaxErr(c, "读取头像文件失败")
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, maxAvatarBytes+1))
	if err != nil || len(raw) == 0 || len(raw) > maxAvatarBytes {
		response.AjaxErr(c, "头像文件不能超过2MiB")
		return
	}
	_, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png") {
		response.AjaxErr(c, "头像必须是真实的JPEG或PNG图片")
		return
	}
	extension := ".png"
	if format == "jpeg" {
		extension = ".jpg"
	}
	filename, err := randomAvatarName(extension)
	if err != nil {
		response.AjaxErr(c, "生成头像文件名失败")
		return
	}
	destination, imageURL, err := avatarDestination(h.cfg.Upload.Path, h.cfg.Upload.Profile, filename)
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		response.AjaxErr(c, "创建头像目录失败")
		return
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		response.AjaxErr(c, "保存头像文件失败")
		return
	}
	_, writeErr := output.Write(raw)
	closeErr := output.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(destination)
		response.AjaxErr(c, "保存头像文件失败")
		return
	}
	result := h.db.Model(&model.SysUser{}).Where("user_id = ? AND del_flag = '0'", login.UserID).
		Updates(map[string]any{"avatar": imageURL, "update_time": time.Now()})
	if result.Error != nil || result.RowsAffected != 1 {
		_ = os.Remove(destination)
		response.AjaxErr(c, "更新头像失败")
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "操作成功", "imgUrl": imageURL})
}

func (h *RuoYiSystemHandler) UserAuthRoleDetail(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:user:query"); !ok {
		return
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil || userID <= 0 {
		response.AjaxErr(c, "用户ID格式不正确")
		return
	}
	var user model.SysUser
	if err := h.db.Where("user_id = ? AND del_flag = '0'", userID).First(&user).Error; err != nil {
		response.AjaxErr(c, "用户不存在")
		return
	}
	type roleWithFlag struct {
		model.SysRole
		Flag bool `json:"flag"`
	}
	roles := make([]roleWithFlag, 0)
	var allRoles []model.SysRole
	if err := h.db.Where("del_flag = '0' AND status = '0'").Order("role_sort, role_id").Find(&allRoles).Error; err != nil {
		response.AjaxErr(c, "查询角色失败")
		return
	}
	assigned := make([]int64, 0)
	if err := h.db.Table("sys_user_role").Where("user_id = ?", userID).Pluck("role_id", &assigned).Error; err != nil {
		response.AjaxErr(c, "查询用户角色失败")
		return
	}
	assignedSet := make(map[int64]struct{}, len(assigned))
	for _, id := range assigned {
		assignedSet[id] = struct{}{}
	}
	for _, role := range allRoles {
		_, flag := assignedSet[role.RoleID]
		roles = append(roles, roleWithFlag{SysRole: role, Flag: flag})
	}
	c.JSON(200, gin.H{"code": 200, "msg": "操作成功", "user": user, "roles": roles})
}

func (h *RuoYiSystemHandler) UserAuthRoleUpdate(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:user:edit"); !ok {
		return
	}
	userID, err := strconv.ParseInt(c.Query("userId"), 10, 64)
	if err != nil || userID <= 0 {
		response.AjaxErr(c, "用户ID格式不正确")
		return
	}
	if userID == 1 {
		response.AjaxErr(c, "不允许修改超级管理员的角色")
		return
	}
	roleIDs, err := parseOptionalUniquePositiveIDs(c.Query("roleIds"))
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	if containsInt64(roleIDs, 1) {
		response.AjaxErr(c, "不允许分配超级管理员角色")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := validateActiveUser(tx, userID); err != nil {
			return err
		}
		if err := validateActiveRoles(tx, roleIDs); err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&userRoleRelation{}).Error; err != nil {
			return errors.New("清理用户角色失败")
		}
		return createUserRoleRelations(tx, userID, roleIDs)
	}); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) RoleChangeStatus(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:role:edit"); !ok {
		return
	}
	var req struct {
		RoleID int64  `json:"roleId"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RoleID <= 0 || (req.Status != "0" && req.Status != "1") {
		response.AjaxErr(c, "参数格式错误")
		return
	}
	if req.RoleID == 1 {
		response.AjaxErr(c, "不允许修改超级管理员角色")
		return
	}
	result := h.db.Model(&model.SysRole{}).Where("role_id = ? AND del_flag = '0'", req.RoleID).
		Updates(map[string]any{"status": req.Status, "update_time": time.Now()})
	if result.Error != nil {
		response.AjaxErr(c, "修改角色状态失败")
		return
	}
	if result.RowsAffected != 1 {
		response.AjaxErr(c, "角色不存在")
		return
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) RoleDataScope(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:role:edit"); !ok {
		return
	}
	var req struct {
		RoleID    int64   `json:"roleId"`
		DataScope string  `json:"dataScope"`
		DeptIDs   []int64 `json:"deptIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RoleID <= 0 || !strings.Contains("12345", req.DataScope) || len(req.DataScope) != 1 {
		response.AjaxErr(c, "参数格式错误")
		return
	}
	if req.RoleID == 1 {
		response.AjaxErr(c, "不允许修改超级管理员角色")
		return
	}
	deptIDs, err := normalizePositiveIDs(req.DeptIDs)
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	if req.DataScope != "2" {
		deptIDs = make([]int64, 0)
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := validateActiveRoles(tx, []int64{req.RoleID}); err != nil {
			return err
		}
		if err := validateActiveDepts(tx, deptIDs); err != nil {
			return err
		}
		result := tx.Model(&model.SysRole{}).Where("role_id = ? AND del_flag = '0' AND status = '0'", req.RoleID).
			Updates(map[string]any{"data_scope": req.DataScope, "update_time": time.Now()})
		if result.Error != nil || result.RowsAffected != 1 {
			return errors.New("修改角色数据范围失败")
		}
		if err := tx.Where("role_id = ?", req.RoleID).Delete(&roleDeptRelation{}).Error; err != nil {
			return errors.New("清理角色部门失败")
		}
		if len(deptIDs) == 0 {
			return nil
		}
		relations := make([]roleDeptRelation, 0, len(deptIDs))
		for _, deptID := range deptIDs {
			relations = append(relations, roleDeptRelation{RoleID: req.RoleID, DeptID: deptID})
		}
		if err := tx.CreateInBatches(relations, 100).Error; err != nil {
			return errors.New("保存角色部门失败")
		}
		return nil
	}); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, nil)
}

type authorizedUserRow struct {
	UserID      int64      `gorm:"column:user_id" json:"userId"`
	UserName    string     `gorm:"column:user_name" json:"userName"`
	NickName    string     `gorm:"column:nick_name" json:"nickName"`
	Email       string     `gorm:"column:email" json:"email"`
	Phonenumber string     `gorm:"column:phonenumber" json:"phonenumber"`
	Status      string     `gorm:"column:status" json:"status"`
	CreateTime  *time.Time `gorm:"column:create_time" json:"createTime"`
}

func (h *RuoYiSystemHandler) RoleAllocatedList(c *gin.Context) {
	h.roleAuthorizationList(c, true)
}

func (h *RuoYiSystemHandler) RoleUnallocatedList(c *gin.Context) {
	h.roleAuthorizationList(c, false)
}

func (h *RuoYiSystemHandler) roleAuthorizationList(c *gin.Context, allocated bool) {
	if _, ok := requireSystemPermission(c, "system:role:list"); !ok {
		return
	}
	roleID, err := strconv.ParseInt(c.Query("roleId"), 10, 64)
	if err != nil || roleID <= 0 {
		response.AjaxErr(c, "角色ID格式不正确")
		return
	}
	if err := validateActiveRoles(h.db, []int64{roleID}); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageSize = capPageSize(pageSize)
	q := h.db.Table("sys_user AS u").Where("u.del_flag = '0' AND u.status = '0'")
	if allocated {
		q = q.Where("EXISTS (SELECT 1 FROM sys_user_role ur WHERE ur.user_id = u.user_id AND ur.role_id = ?)", roleID)
	} else {
		q = q.Where("u.user_id <> 1 AND NOT EXISTS (SELECT 1 FROM sys_user_role ur WHERE ur.user_id = u.user_id AND ur.role_id = ?)", roleID)
	}
	if value := strings.TrimSpace(c.Query("userName")); value != "" {
		q = q.Where("u.user_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(c.Query("phonenumber")); value != "" {
		q = q.Where("u.phonenumber LIKE ?", "%"+value+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.AjaxErr(c, "查询授权用户总数失败")
		return
	}
	rows := make([]authorizedUserRow, 0)
	if err := q.Select("u.user_id, u.user_name, u.nick_name, u.email, u.phonenumber, u.status, u.create_time").
		Order("u.user_id").Offset((pageNum - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		response.AjaxErr(c, "查询授权用户列表失败")
		return
	}
	response.Table(c, rows, total)
}

func (h *RuoYiSystemHandler) RoleAuthCancel(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:role:remove"); !ok {
		return
	}
	var req struct {
		RoleID int64 `json:"roleId"`
		UserID int64 `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RoleID <= 0 || req.UserID <= 0 {
		response.AjaxErr(c, "参数格式错误")
		return
	}
	if err := h.removeRoleUsers(req.RoleID, []int64{req.UserID}); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) RoleAuthCancelAll(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:role:remove"); !ok {
		return
	}
	roleID, err := strconv.ParseInt(c.Query("roleId"), 10, 64)
	if err != nil || roleID <= 0 {
		response.AjaxErr(c, "角色ID格式不正确")
		return
	}
	userIDs, err := parseUniquePositiveIDs(c.Query("userIds"))
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	if err := h.removeRoleUsers(roleID, userIDs); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) removeRoleUsers(roleID int64, userIDs []int64) error {
	if roleID == 1 || containsInt64(userIDs, 1) {
		return errors.New("不允许修改超级管理员用户或角色")
	}
	return h.db.Transaction(func(tx *gorm.DB) error {
		if err := validateActiveRoles(tx, []int64{roleID}); err != nil {
			return err
		}
		if err := validateActiveUsers(tx, userIDs); err != nil {
			return err
		}
		if err := tx.Where("role_id = ? AND user_id IN ?", roleID, userIDs).Delete(&userRoleRelation{}).Error; err != nil {
			return errors.New("取消用户授权失败")
		}
		return nil
	})
}

func (h *RuoYiSystemHandler) RoleAuthSelectAll(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:role:add"); !ok {
		return
	}
	roleID, err := strconv.ParseInt(c.Query("roleId"), 10, 64)
	if err != nil || roleID <= 0 {
		response.AjaxErr(c, "角色ID格式不正确")
		return
	}
	userIDs, err := parseUniquePositiveIDs(c.Query("userIds"))
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	if roleID == 1 || containsInt64(userIDs, 1) {
		response.AjaxErr(c, "不允许修改超级管理员用户或角色")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := validateActiveRoles(tx, []int64{roleID}); err != nil {
			return err
		}
		if err := validateActiveUsers(tx, userIDs); err != nil {
			return err
		}
		existing := make([]int64, 0)
		if err := tx.Table("sys_user_role").Where("role_id = ? AND user_id IN ?", roleID, userIDs).Pluck("user_id", &existing).Error; err != nil {
			return errors.New("查询已有授权失败")
		}
		existingSet := make(map[int64]struct{}, len(existing))
		for _, id := range existing {
			existingSet[id] = struct{}{}
		}
		relations := make([]userRoleRelation, 0, len(userIDs))
		for _, userID := range userIDs {
			if _, exists := existingSet[userID]; !exists {
				relations = append(relations, userRoleRelation{UserID: userID, RoleID: roleID})
			}
		}
		if len(relations) > 0 {
			if err := tx.CreateInBatches(relations, 100).Error; err != nil {
				return errors.New("保存用户授权失败")
			}
		}
		return nil
	}); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AjaxErr(c, "参数格式错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if err := validateRegisterRequest(req); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	var configValue string
	if err := h.db.Table("sys_config").Where("config_key = ?", "sys.account.registerUser").Pluck("config_value", &configValue).Error; err != nil {
		response.AjaxErr(c, "读取注册配置失败")
		return
	}
	if !strings.EqualFold(strings.TrimSpace(configValue), "true") {
		response.AjaxErr(c, "当前系统未开放用户注册")
		return
	}
	if h.cfg == nil {
		response.AjaxErr(c, "注册配置不可用")
		return
	}
	if h.cfg.Captcha.Enabled {
		if err := consumeCaptcha(c.Request.Context(), req.UUID, req.Code); err != nil {
			response.AjaxErr(c, err.Error())
			return
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.AjaxErr(c, "密码加密失败")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.SysUser{}).Where("user_name = ?", req.Username).Count(&count).Error; err != nil {
			return errors.New("校验用户名失败")
		}
		if count > 0 {
			return errors.New("用户名已存在")
		}
		now := time.Now()
		user := struct {
			UserID     int64      `gorm:"column:user_id;primaryKey;autoIncrement"`
			UserName   string     `gorm:"column:user_name"`
			NickName   string     `gorm:"column:nick_name"`
			UserType   string     `gorm:"column:user_type"`
			Password   string     `gorm:"column:password"`
			Status     string     `gorm:"column:status"`
			DelFlag    string     `gorm:"column:del_flag"`
			Sex        string     `gorm:"column:sex"`
			CreateBy   string     `gorm:"column:create_by"`
			CreateTime *time.Time `gorm:"column:create_time"`
			UpdateTime *time.Time `gorm:"column:update_time"`
		}{UserName: req.Username, NickName: req.Username, UserType: "00", Password: string(hash), Status: "0", DelFlag: "0", Sex: "2", CreateBy: "register", CreateTime: &now, UpdateTime: &now}
		if err := tx.Table("sys_user").Create(&user).Error; err != nil {
			if isDuplicateDatabaseError(err) {
				return errors.New("用户名已存在")
			}
			return errors.New("创建用户失败")
		}
		var common model.SysRole
		err := tx.Where("role_key = ? AND status = '0' AND del_flag = '0'", "common").First(&common).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return errors.New("查询默认角色失败")
		}
		if err := tx.Create(&userRoleRelation{UserID: user.UserID, RoleID: common.RoleID}).Error; err != nil {
			return errors.New("分配默认角色失败")
		}
		return nil
	}); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, nil)
}

func consumeCaptcha(ctx context.Context, uuidValue, answer string) error {
	if redisx.Client == nil {
		return errors.New("验证码服务不可用")
	}
	key := redisx.CaptchaKey + strings.TrimSpace(uuidValue)
	script := redis.NewScript(`local value = redis.call('GET', KEYS[1]); if value then redis.call('DEL', KEYS[1]); end; return value`)
	value, err := script.Run(ctx, redisx.Client, []string{key}).Text()
	if err == redis.Nil {
		return errors.New("验证码已失效")
	}
	if err != nil {
		return errors.New("验证码服务不可用")
	}
	value = strings.Trim(value, "\"")
	if !strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(answer)) {
		return errors.New("验证码错误")
	}
	return nil
}

func normalizePositiveIDs(values []int64) ([]int64, error) {
	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			return nil, errors.New("ID格式不正确")
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func containsInt64(values []int64, target int64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func validateActiveUser(tx *gorm.DB, userID int64) error {
	return validateActiveUsers(tx, []int64{userID})
}

func validateActiveUsers(tx *gorm.DB, userIDs []int64) error {
	if len(userIDs) == 0 {
		return errors.New("用户ID不能为空")
	}
	var count int64
	if err := tx.Model(&model.SysUser{}).Where("user_id IN ? AND status = '0' AND del_flag = '0'", userIDs).Count(&count).Error; err != nil {
		return errors.New("校验用户失败")
	}
	if count != int64(len(userIDs)) {
		return errors.New("用户不存在或已停用")
	}
	return nil
}

func validateActiveRoles(tx *gorm.DB, roleIDs []int64) error {
	if len(roleIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&model.SysRole{}).Where("role_id IN ? AND status = '0' AND del_flag = '0'", roleIDs).Count(&count).Error; err != nil {
		return errors.New("校验角色失败")
	}
	if count != int64(len(roleIDs)) {
		return errors.New("角色不存在或已停用")
	}
	return nil
}

func validateActiveDepts(tx *gorm.DB, deptIDs []int64) error {
	if len(deptIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Table("sys_dept").Where("dept_id IN ? AND status = '0' AND del_flag = '0'", deptIDs).Count(&count).Error; err != nil {
		return errors.New("校验部门失败")
	}
	if count != int64(len(deptIDs)) {
		return errors.New("部门不存在或已停用")
	}
	return nil
}

func createUserRoleRelations(tx *gorm.DB, userID int64, roleIDs []int64) error {
	if len(roleIDs) == 0 {
		return nil
	}
	relations := make([]userRoleRelation, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		relations = append(relations, userRoleRelation{UserID: userID, RoleID: roleID})
	}
	if err := tx.CreateInBatches(relations, 100).Error; err != nil {
		return errors.New("保存用户角色失败")
	}
	return nil
}

func isDuplicateDatabaseError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate entry") || strings.Contains(message, "error 1062")
}
