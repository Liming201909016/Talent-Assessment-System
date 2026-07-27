package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
)

// TestBugFB088_ProfileValidationRejectsInvalidAndAcceptsRuoYiFields
// 对应：docs/regression-tests.md #FB-088
func TestBugFB088_ProfileValidationRejectsInvalidAndAcceptsRuoYiFields(t *testing.T) {
	valid := userProfileUpdateRequest{NickName: "安全用户", Email: "safe@example.com", Phonenumber: "13812345678", Sex: "2"}
	if err := validateUserProfileUpdate(valid); err != nil {
		t.Fatalf("valid profile rejected: %v", err)
	}
	invalid := []userProfileUpdateRequest{
		{NickName: "", Email: "safe@example.com", Phonenumber: "13812345678", Sex: "0"},
		{NickName: "safe", Email: "not-an-email", Phonenumber: "13812345678", Sex: "0"},
		{NickName: "safe", Email: "safe@example.com", Phonenumber: "123", Sex: "0"},
		{NickName: "safe", Email: "safe@example.com", Phonenumber: "13812345678", Sex: "9"},
	}
	for i, req := range invalid {
		if err := validateUserProfileUpdate(req); err == nil {
			t.Fatalf("invalid profile %d accepted", i)
		}
	}
}

// TestBugFB089_PasswordPolicyRequiresMixedEightCharacters
// 对应：docs/regression-tests.md #FB-089
func TestBugFB089_PasswordPolicyRequiresMixedEightCharacters(t *testing.T) {
	for _, password := range []string{"short1", "onlyletters", "12345678"} {
		if err := validateStrongPassword(password); err == nil {
			t.Fatalf("weak password accepted: %q", password)
		}
	}
	if err := validateStrongPassword("secure123"); err != nil {
		t.Fatalf("strong password rejected: %v", err)
	}
}

func TestBugFB089_ProfilePasswordIsNotSentInURLQuery(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	handlerSource, err := os.ReadFile(filepath.Join(filepath.Dir(current), "ruoyi_administration.go"))
	if err != nil {
		t.Fatal(err)
	}
	passwordMethod := string(handlerSource)
	start := strings.Index(passwordMethod, "func (h *RuoYiSystemHandler) UserProfileUpdatePwd")
	if start < 0 {
		t.Fatal("password handler not found")
	}
	end := strings.Index(passwordMethod[start:], "func invalidateUserSessions")
	if end < 0 {
		t.Fatal("password handler not found")
	}
	passwordMethod = passwordMethod[start : start+end]
	if strings.Contains(passwordMethod, `c.Query("oldPassword")`) || strings.Contains(passwordMethod, `c.Query("newPassword")`) {
		t.Fatal("password handler reads secrets from URL query parameters")
	}
	apiPath := filepath.Join(filepath.Dir(current), "..", "..", "ruoyi-ui", "src", "api", "system", "user.js")
	apiSource, err := os.ReadFile(apiPath)
	if err != nil {
		t.Fatal(err)
	}
	api := string(apiSource)
	apiStart := strings.Index(api, "export function updateUserPwd")
	if apiStart < 0 {
		t.Fatal("Vue password API not found")
	}
	apiEnd := strings.Index(api[apiStart:], "// 用户头像上传")
	if apiEnd < 0 {
		t.Fatal("Vue password API not found")
	}
	api = api[apiStart : apiStart+apiEnd]
	if strings.Contains(api, "params: data") || !strings.Contains(api, "data: data") {
		t.Fatal("Vue password API still sends secrets in URL query parameters")
	}
}

// TestBugFB090_AvatarDestinationStaysInsideConfiguredProfile
// 对应：docs/regression-tests.md #FB-090
func TestBugFB090_AvatarDestinationStaysInsideConfiguredProfile(t *testing.T) {
	root := t.TempDir()
	path, urlPath, err := avatarDestination(root, "/profile", "avatar.jpg")
	if err != nil {
		t.Fatalf("avatar destination rejected: %v", err)
	}
	profileRoot := filepath.Join(root, "profile")
	rel, err := filepath.Rel(profileRoot, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatalf("avatar path escaped profile root: %q", path)
	}
	if urlPath != "/profile/avatar/avatar.jpg" {
		t.Fatalf("unexpected avatar URL: %q", urlPath)
	}
	if _, _, err := avatarDestination(root, "/../outside", "avatar.jpg"); err == nil {
		t.Fatal("path-traversal profile accepted")
	}
}

// TestBugFB091_SensitiveAdministrationRequiresExactPermission
// 对应：docs/regression-tests.md #FB-091
func TestBugFB091_SensitiveAdministrationRequiresExactPermission(t *testing.T) {
	cases := []struct {
		name string
		user *model.LoginUser
		perm string
		want bool
	}{
		{"admin id", &model.LoginUser{UserID: 1}, "system:role:edit", true},
		{"admin role", &model.LoginUser{UserID: 8, Roles: []string{"admin"}}, "system:role:edit", true},
		{"wildcard", &model.LoginUser{UserID: 8, Permissions: []string{"*:*:*"}}, "system:role:edit", true},
		{"exact", &model.LoginUser{UserID: 8, Permissions: []string{"system:role:edit"}}, "system:role:edit", true},
		{"unrelated", &model.LoginUser{UserID: 8, Permissions: []string{"system:user:list"}}, "system:role:edit", false},
		{"missing", nil, "system:role:edit", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasSystemPermission(tc.user, tc.perm); got != tc.want {
				t.Fatalf("hasSystemPermission()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestBugFB091_SensitiveHandlersRejectBeforeDatabaseAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &RuoYiSystemHandler{}
	handlers := []struct {
		name   string
		method string
		path   string
		call   gin.HandlerFunc
	}{
		{"user auth role detail", http.MethodGet, "/system/user/authRole/2", h.UserAuthRoleDetail},
		{"user auth role update", http.MethodPut, "/system/user/authRole?userId=2&roleIds=2", h.UserAuthRoleUpdate},
		{"role status", http.MethodPut, "/system/role/changeStatus", h.RoleChangeStatus},
		{"role data scope", http.MethodPut, "/system/role/dataScope", h.RoleDataScope},
		{"allocated list", http.MethodGet, "/system/role/authUser/allocatedList?roleId=2", h.RoleAllocatedList},
		{"unallocated list", http.MethodGet, "/system/role/authUser/unallocatedList?roleId=2", h.RoleUnallocatedList},
		{"cancel", http.MethodPut, "/system/role/authUser/cancel", h.RoleAuthCancel},
		{"cancel all", http.MethodPut, "/system/role/authUser/cancelAll?roleId=2&userIds=3", h.RoleAuthCancelAll},
		{"select all", http.MethodPut, "/system/role/authUser/selectAll?roleId=2&userIds=3", h.RoleAuthSelectAll},
	}
	for _, tc := range handlers {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("loginUser", &model.LoginUser{UserID: 9, Permissions: []string{"system:unrelated"}})
			tc.call(c)
			if w.Code != http.StatusForbidden {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestBugFB089_SessionInvalidationUsesScanNotKeys(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(current), "ruoyi_administration.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func invalidateUserSessions")
	if start < 0 {
		t.Fatal("session invalidation helper not found")
	}
	end := strings.Index(source[start:], "func (h *RuoYiSystemHandler) UserProfileAvatar")
	if end < 0 {
		t.Fatal("session invalidation helper not found")
	}
	helper := source[start : start+end]
	if !strings.Contains(helper, ".Scan(") || strings.Contains(helper, ".Keys(") {
		t.Fatal("user session invalidation must use SCAN and must not use KEYS")
	}
}

// TestBugFB092_RelationshipIDsAreStrictPositiveAndDeduplicated
// 对应：docs/regression-tests.md #FB-092
func TestBugFB092_RelationshipIDsAreStrictPositiveAndDeduplicated(t *testing.T) {
	ids, err := parseUniquePositiveIDs("3,2,3")
	if err != nil || len(ids) != 2 || ids[0] != 3 || ids[1] != 2 {
		t.Fatalf("unexpected normalized IDs: %v, err=%v", ids, err)
	}
	for _, raw := range []string{"", "0", "-1", "1,,2", "abc"} {
		if _, err := parseUniquePositiveIDs(raw); err == nil {
			t.Fatalf("invalid ID list accepted: %q", raw)
		}
	}
}

// TestBugFB093_RegistrationValidationKeepsVueContractAndStrongPassword
// 对应：docs/regression-tests.md #FB-093
func TestBugFB093_RegistrationValidationKeepsVueContractAndStrongPassword(t *testing.T) {
	valid := registerRequest{Username: "new_user", Password: "secure123", ConfirmPassword: "secure123", Code: "12", UUID: "uuid"}
	if err := validateRegisterRequest(valid); err != nil {
		t.Fatalf("valid registration rejected: %v", err)
	}
	invalid := []registerRequest{
		{Username: "中文", Password: "secure123", ConfirmPassword: "secure123", Code: "12", UUID: "uuid"},
		{Username: "new_user", Password: "weak", ConfirmPassword: "weak", Code: "12", UUID: "uuid"},
		{Username: "new_user", Password: "secure123", ConfirmPassword: "different1", Code: "12", UUID: "uuid"},
		{Username: "new_user", Password: "secure123", ConfirmPassword: "secure123", Code: "", UUID: "uuid"},
	}
	for i, req := range invalid {
		if err := validateRegisterRequest(req); err == nil {
			t.Fatalf("invalid registration %d accepted", i)
		}
	}
}

// TestBugFB094_ThirteenP1RoutesNoLongerUseSuccessStubs
// 对应：docs/regression-tests.md #FB-094
func TestBugFB094_ThirteenP1RoutesNoLongerUseSuccessStubs(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	routerPath := filepath.Join(filepath.Dir(current), "..", "router", "router.go")
	raw, err := os.ReadFile(routerPath)
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, marker := range []string{
		`sys.PUT("/user/profile", handler.AjaxStub`,
		`sys.PUT("/user/profile/updatePwd", handler.AjaxStub`,
		`sys.POST("/user/profile/avatar", handler.AjaxStub`,
		`sys.GET("/user/authRole/:userId", handler.AjaxStub`,
		`sys.PUT("/user/authRole", handler.AjaxStub`,
		`sys.PUT("/role/changeStatus", handler.AjaxStub`,
		`sys.PUT("/role/dataScope", handler.AjaxStub`,
		`sys.GET("/role/authUser/allocatedList", handler.TableStub`,
		`sys.GET("/role/authUser/unallocatedList", handler.TableStub`,
		`sys.PUT("/role/authUser/cancel", handler.AjaxStub`,
		`sys.PUT("/role/authUser/cancelAll", handler.AjaxStub`,
		`sys.PUT("/role/authUser/selectAll", handler.AjaxStub`,
		`r.POST("/register", handler.AjaxStub`,
	} {
		if strings.Contains(source, marker) {
			t.Errorf("P1 route still uses success stub: %s", marker)
		}
	}
	if !strings.Contains(source, "NewRuoYiSystemHandler(db, cfg)") {
		t.Error("RuoYiSystemHandler is not constructed with cfg")
	}
}

// TestBugFB095_UserNameUniqueMigrationChecksDuplicatesBeforeDDL
// 对应：docs/regression-tests.md #FB-095
func TestBugFB095_UserNameUniqueMigrationChecksDuplicatesBeforeDDL(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	migration := filepath.Join(filepath.Dir(current), "..", "..", "..", "scripts", "sql", "system_001_user_name_unique.sql")
	raw, err := os.ReadFile(migration)
	if err != nil {
		t.Fatalf("migration missing: %v", err)
	}
	source := strings.ToLower(string(raw))
	duplicateCheck := strings.Index(source, "group by user_name")
	alter := strings.Index(source, "alter table sys_user")
	if duplicateCheck < 0 || alter < 0 || duplicateCheck > alter {
		t.Fatal("migration must check duplicate user_name values before ALTER TABLE")
	}
	for _, destructive := range []string{"delete from sys_user", "update sys_user", "drop table"} {
		if strings.Contains(source, destructive) {
			t.Fatalf("destructive migration statement found: %s", destructive)
		}
	}
	if !strings.Contains(source, "information_schema.statistics") || !strings.Contains(source, "prepare") {
		t.Fatal("migration is not MySQL 5.7 idempotent")
	}
}
