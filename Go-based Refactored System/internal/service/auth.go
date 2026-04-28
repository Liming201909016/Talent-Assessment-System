package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/repository"
	"github.com/talent-assessment/refactored/pkg/captcha"
	jwtpkg "github.com/talent-assessment/refactored/pkg/jwt"
	"github.com/talent-assessment/refactored/pkg/redisx"
	"golang.org/x/crypto/bcrypt"
)

// CompareBCrypt 兼容 Spring BCryptPasswordEncoder ($2a$ / $2b$)；Go 的 crypto/bcrypt 可直接校验。
func CompareBCrypt(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

type AuthService struct {
	cfg      *config.Config
	userRepo *repository.SysUserRepo
	menuRepo *repository.SysMenuRepo
}

func NewAuthService(cfg *config.Config, ur *repository.SysUserRepo, mr *repository.SysMenuRepo) *AuthService {
	return &AuthService{cfg: cfg, userRepo: ur, menuRepo: mr}
}

// GenerateCaptcha 生成 math 验证码：返回 uuid + base64 png
func (s *AuthService) GenerateCaptcha(ctx context.Context) (uuidStr, imgB64 string, err error) {
	img, ans, err := captcha.GenerateMath()
	if err != nil {
		return "", "", err
	}
	u := strings.ReplaceAll(uuid.New().String(), "-", "")
	if err := redisx.Client.Set(ctx, redisx.CaptchaKey+u, ans, 2*time.Minute).Err(); err != nil {
		return "", "", err
	}
	return u, img, nil
}

// ValidateCaptcha 校验并删除；Java 侧一次性验证逻辑一致
func (s *AuthService) ValidateCaptcha(ctx context.Context, code, uuidStr string) error {
	if !s.cfg.Captcha.Enabled {
		return nil
	}
	key := redisx.CaptchaKey + uuidStr
	raw, err := redisx.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return errors.New("验证码已失效")
	}
	if err != nil {
		return err
	}
	redisx.Client.Del(ctx, key)
	// Java 侧经 Jackson 序列化存的可能带引号，这里做宽松处理
	raw = strings.Trim(raw, "\"")
	if !strings.EqualFold(strings.TrimSpace(raw), strings.TrimSpace(code)) {
		return errors.New("验证码错误")
	}
	return nil
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
	UUID     string `json:"uuid"`
}

// Login 校验验证码 + 用户密码，生成 JWT 并持久化 LoginUser
//
// FB-015: 增加登录失败次数限制
//   - 同一用户名连续失败 5 次 → 锁定 15 分钟
//   - 锁定期间任何登录请求都直接拒绝
func (s *AuthService) Login(ctx context.Context, req LoginReq, r *http.Request) (token string, err error) {
	// FB-015: 先检查失败次数（按用户名）
	if req.Username != "" {
		failKey := redisx.LoginFailKey + req.Username
		failCount, _ := redisx.Client.Get(ctx, failKey).Int()
		if failCount >= LoginMaxFailures {
			return "", errors.New("登录失败次数过多，请 15 分钟后再试")
		}
	}

	if err := s.ValidateCaptcha(ctx, req.Code, req.UUID); err != nil {
		s.recordLoginFailure(ctx, req.Username)
		return "", err
	}
	u, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		s.recordLoginFailure(ctx, req.Username)
		// FB-015 加固：不区分"用户不存在"和"密码错误"，避免泄露用户名是否存在
		return "", errors.New("用户名或密码错误")
	}
	if u.Status == "1" {
		return "", errors.New("对不起，您的账号已停用")
	}
	if !CompareBCrypt(u.Password, req.Password) {
		s.recordLoginFailure(ctx, req.Username)
		return "", errors.New("用户名或密码错误")
	}

	// 登录成功，清除失败计数
	if req.Username != "" {
		redisx.Client.Del(ctx, redisx.LoginFailKey+req.Username)
	}

	_ = s.userRepo.UpdateLogin(u.UserID, clientIP(r))

	roles, _ := s.userRepo.GetRoleKeys(u.UserID)
	perms, _ := s.userRepo.GetMenuPerms(u.UserID)

	tokUUID := strings.ReplaceAll(uuid.New().String(), "-", "")
	now := time.Now().UnixMilli()
	lu := &model.LoginUser{
		UserID:      u.UserID,
		DeptID:      u.DeptID,
		Token:       tokUUID,
		LoginTime:   now,
		ExpireTime:  now + int64(s.cfg.Jwt.ExpireMinutes)*60_000,
		IPAddr:      clientIP(r),
		Browser:     r.Header.Get("User-Agent"),
		Permissions: perms,
		Roles:       roles,
		User:        u,
	}
	if err := s.storeLoginUser(ctx, lu); err != nil {
		return "", err
	}
	claims := map[string]any{s.cfg.Jwt.LoginUserKey: tokUUID}
	return jwtpkg.Create(s.cfg.Jwt.Secret, claims)
}

// LoginMaxFailures FB-015: 登录失败最大次数（达到后锁定）
const LoginMaxFailures = 5

// LoginFailureLockDuration FB-015: 失败计数器有效期 = 锁定时长
const LoginFailureLockDuration = 15 * time.Minute

// recordLoginFailure 记录登录失败，达到 LoginMaxFailures 触发锁定
func (s *AuthService) recordLoginFailure(ctx context.Context, username string) {
	if username == "" {
		return
	}
	key := redisx.LoginFailKey + username
	cnt, _ := redisx.Client.Incr(ctx, key).Result()
	if cnt == 1 {
		// 首次失败，设置过期时间
		redisx.Client.Expire(ctx, key, LoginFailureLockDuration)
	}
}

func (s *AuthService) storeLoginUser(ctx context.Context, lu *model.LoginUser) error {
	b, _ := json.Marshal(lu)
	return redisx.Client.Set(ctx, redisx.LoginTokenKey+lu.Token, b, time.Duration(s.cfg.Jwt.ExpireMinutes)*time.Minute).Err()
}

// ParseAuth 从 Authorization 头解析 token → LoginUser
func (s *AuthService) ParseAuth(ctx context.Context, authHeader string) (*model.LoginUser, error) {
	if authHeader == "" {
		return nil, errors.New("missing token")
	}
	raw := authHeader
	if strings.HasPrefix(raw, s.cfg.Jwt.Prefix) {
		raw = strings.TrimPrefix(raw, s.cfg.Jwt.Prefix)
	}
	claims, err := jwtpkg.Parse(s.cfg.Jwt.Secret, raw)
	if err != nil {
		return nil, err
	}
	tokUUID, _ := claims[s.cfg.Jwt.LoginUserKey].(string)
	if tokUUID == "" {
		return nil, errors.New("no login_user_key")
	}
	b, err := redisx.Client.Get(ctx, redisx.LoginTokenKey+tokUUID).Bytes()
	if err != nil {
		return nil, errors.New("token expired")
	}
	var lu model.LoginUser
	if err := json.Unmarshal(b, &lu); err != nil {
		return nil, err
	}
	// 相差不足 20 分钟自动刷新
	if lu.ExpireTime-time.Now().UnixMilli() < 20*60*1000 {
		now := time.Now().UnixMilli()
		lu.LoginTime = now
		lu.ExpireTime = now + int64(s.cfg.Jwt.ExpireMinutes)*60_000
		_ = s.storeLoginUser(ctx, &lu)
	}
	return &lu, nil
}

// Logout 清除 Redis 登录记录
func (s *AuthService) Logout(ctx context.Context, authHeader string) error {
	raw := strings.TrimPrefix(authHeader, s.cfg.Jwt.Prefix)
	if raw == "" {
		return nil
	}
	claims, err := jwtpkg.Parse(s.cfg.Jwt.Secret, raw)
	if err != nil {
		return nil
	}
	tokUUID, _ := claims[s.cfg.Jwt.LoginUserKey].(string)
	if tokUUID == "" {
		return nil
	}
	redisx.Client.Del(ctx, redisx.LoginTokenKey+tokUUID)
	return nil
}

// BuildMenus 将 SysMenu 线性列表构造成前端路由树（RuoYi buildMenus）
func (s *AuthService) BuildMenus(userID int64) ([]map[string]any, error) {
	menus, err := s.menuRepo.SelectMenuTreeByUserID(userID)
	if err != nil {
		return nil, err
	}
	byID := map[int64]*model.SysMenu{}
	for _, m := range menus {
		byID[m.MenuID] = m
	}
	var roots []*model.SysMenu
	for _, m := range menus {
		if p, ok := byID[m.ParentID]; ok && m.ParentID != 0 {
			p.Children = append(p.Children, m)
		} else {
			roots = append(roots, m)
		}
	}
	return menusToRouter(roots), nil
}

func menusToRouter(items []*model.SysMenu) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, m := range items {
		path := m.Path
		// 顶级菜单（ParentID==0）path 必须以 / 开头，与 Java SysMenuServiceImpl.buildMenus 一致
		if m.ParentID == 0 && path != "" && path[0] != '/' {
			path = "/" + path
		}
		r := map[string]any{
			"name":      capFirst(m.Path),
			"path":      path,
			"hidden":    m.Visible == "1",
			"component": componentOf(m),
			"query":     m.Query,
			"meta": map[string]any{
				"title":   m.MenuName,
				"icon":    m.Icon,
				"noCache": m.IsCache == 1,
				"link":    nil,
			},
		}
		if len(m.Children) > 0 && m.MenuType == "M" {
			r["alwaysShow"] = true
			r["redirect"] = "noRedirect"
			r["children"] = menusToRouter(m.Children)
		}
		out = append(out, r)
	}
	return out
}

func componentOf(m *model.SysMenu) string {
	if m.Component != "" && m.MenuType != "M" {
		return m.Component
	}
	if m.ParentID == 0 && m.MenuType == "M" {
		return "Layout"
	}
	if m.Component == "" && m.MenuType == "M" {
		return "ParentView"
	}
	return m.Component
}

func capFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	return r.RemoteAddr
}
