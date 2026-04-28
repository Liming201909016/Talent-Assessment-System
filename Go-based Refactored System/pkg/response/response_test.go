package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func runHandler(h gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	h(c)
	return w
}

func parseJSON(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("invalid JSON: %v\nbody: %s", err, body)
	}
	return m
}

// ===================== AjaxOK / AjaxErr =====================

func TestAjaxOK_Format(t *testing.T) {
	w := runHandler(func(c *gin.Context) {
		AjaxOK(c, gin.H{"id": 1, "name": "test"})
	})
	if w.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", w.Code)
	}
	m := parseJSON(t, w.Body.Bytes())
	if m["code"].(float64) != 200 {
		t.Errorf("code: want 200, got %v", m["code"])
	}
	if m["msg"] != "操作成功" {
		t.Errorf("msg: want '操作成功', got %v", m["msg"])
	}
	data, ok := m["data"].(map[string]any)
	if !ok {
		t.Fatal("data should be object")
	}
	if data["id"].(float64) != 1 {
		t.Errorf("data.id: want 1, got %v", data["id"])
	}
}

func TestAjaxOK_NilData(t *testing.T) {
	w := runHandler(func(c *gin.Context) { AjaxOK(c, nil) })
	m := parseJSON(t, w.Body.Bytes())
	if _, exists := m["data"]; exists {
		t.Error("data key should be omitted when nil")
	}
}

func TestAjaxErr_Format(t *testing.T) {
	w := runHandler(func(c *gin.Context) { AjaxErr(c, "数据库连接失败") })
	if w.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", w.Code)
	}
	m := parseJSON(t, w.Body.Bytes())
	if m["code"].(float64) != 500 {
		t.Errorf("code: want 500, got %v", m["code"])
	}
	if m["msg"] != "数据库连接失败" {
		t.Errorf("msg: want '数据库连接失败', got %v", m["msg"])
	}
}

func TestAjaxUnauthorized_DefaultMsg(t *testing.T) {
	w := runHandler(func(c *gin.Context) { AjaxUnauthorized(c, "") })
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", w.Code)
	}
	m := parseJSON(t, w.Body.Bytes())
	if m["code"].(float64) != 401 {
		t.Errorf("code: want 401, got %v", m["code"])
	}
	if m["msg"] == "" {
		t.Error("default msg should not be empty")
	}
}

func TestAjaxForbidden_CustomMsg(t *testing.T) {
	w := runHandler(func(c *gin.Context) { AjaxForbidden(c, "需要管理员权限") })
	if w.Code != http.StatusForbidden {
		t.Errorf("status: want 403, got %d", w.Code)
	}
	m := parseJSON(t, w.Body.Bytes())
	if m["msg"] != "需要管理员权限" {
		t.Errorf("msg: want '需要管理员权限', got %v", m["msg"])
	}
}

// ===================== Rest / RestErr (qu/exam/paper/candidate 用) =====================

func TestRest_Format(t *testing.T) {
	w := runHandler(func(c *gin.Context) {
		Rest(c, gin.H{"foo": "bar"})
	})
	m := parseJSON(t, w.Body.Bytes())
	if m["code"].(float64) != 0 {
		t.Errorf("code: want 0, got %v", m["code"])
	}
	if m["msg"] != "" {
		t.Errorf("msg: want empty, got %v", m["msg"])
	}
	if m["success"] != true {
		t.Errorf("success: want true, got %v", m["success"])
	}
	data := m["data"].(map[string]any)
	if data["foo"] != "bar" {
		t.Errorf("data.foo: want 'bar', got %v", data["foo"])
	}
}

func TestRestErr_Format(t *testing.T) {
	w := runHandler(func(c *gin.Context) { RestErr(c, "参数错误") })
	m := parseJSON(t, w.Body.Bytes())
	if m["code"].(float64) != 1 {
		t.Errorf("code: want 1, got %v", m["code"])
	}
	if m["msg"] != "参数错误" {
		t.Errorf("msg: want '参数错误', got %v", m["msg"])
	}
	if m["success"] != false {
		t.Errorf("success: want false, got %v", m["success"])
	}
}

// ===================== Table (tester 列表用) =====================

func TestTable_Format(t *testing.T) {
	rows := []map[string]any{
		{"id": 1, "name": "张三"},
		{"id": 2, "name": "李四"},
	}
	w := runHandler(func(c *gin.Context) { Table(c, rows, 2) })

	m := parseJSON(t, w.Body.Bytes())
	if m["code"].(float64) != 200 {
		t.Errorf("code: want 200, got %v", m["code"])
	}
	if m["total"].(float64) != 2 {
		t.Errorf("total: want 2, got %v", m["total"])
	}
	rowsResp, ok := m["rows"].([]any)
	if !ok {
		t.Fatalf("rows should be array, got %T", m["rows"])
	}
	if len(rowsResp) != 2 {
		t.Errorf("rows length: want 2, got %d", len(rowsResp))
	}
}

func TestTable_EmptyRows(t *testing.T) {
	// 关键测试：空列表必须返回 [] 而非 null（前端 el-table 兼容）
	rows := make([]map[string]any, 0)
	w := runHandler(func(c *gin.Context) { Table(c, rows, 0) })

	body := w.Body.String()
	if !contains(body, `"rows":[]`) {
		t.Errorf("empty rows should be [] not null. body: %s", body)
	}
	m := parseJSON(t, w.Body.Bytes())
	if m["total"].(float64) != 0 {
		t.Errorf("total: want 0, got %v", m["total"])
	}
}

func TestTable_NilRows_KnownIssue(t *testing.T) {
	// 已知陷阱：var rows []T → nil → JSON = null（前端会显示"暂无数据"）
	// 此测试记录该行为，用于回归保护：调用方必须用 make() 而非 var
	var rows []map[string]any // 显式 nil
	w := runHandler(func(c *gin.Context) { Table(c, rows, 0) })

	body := w.Body.String()
	if !contains(body, `"rows":null`) {
		t.Logf("Note: nil slice serializes to %s", body)
	}
	// 这是 Go encoding/json 的行为，我们记录但不修复
	// 修复方法在调用方：handler 中必须用 make([]T, 0)
}

// ===================== ApiRest struct 字段顺序与 JSON tag =====================

func TestApiRest_JSONTags(t *testing.T) {
	r := ApiRest{Code: 0, Msg: "", Data: "x", Success: true}
	b, _ := json.Marshal(r)
	s := string(b)
	for _, key := range []string{`"code":`, `"msg":`, `"data":`, `"success":`} {
		if !contains(s, key) {
			t.Errorf("missing JSON key %s in %s", key, s)
		}
	}
}

func TestApiRest_OmitEmptyData(t *testing.T) {
	// data 字段使用 omitempty，nil 时应不出现
	r := ApiRest{Code: 1, Msg: "err", Data: nil, Success: false}
	b, _ := json.Marshal(r)
	if contains(string(b), `"data":`) {
		t.Errorf("nil data should be omitted, got: %s", b)
	}
}

// helper
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
