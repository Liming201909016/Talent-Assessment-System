---
applyTo: "**/*_test.go"
description: "Go 测试专家：生成单元测试时自动应用项目编码规约和测试规范"
---

# Go 单元测试生成规则

## 1. 测试文件位置

- 测试文件与被测文件在同一目录：`handler/exam.go` → `handler/exam_test.go`
- 使用 `package handler` （同包测试，可访问内部函数）
- 文件名格式：`{被测文件名}_test.go`

## 2. 测试命名规范

```go
// 函数命名：Test{函数名}_{场景}
func TestCreateExam_Success(t *testing.T) { ... }
func TestCreateExam_MissingTitle(t *testing.T) { ... }
func TestCreateExam_InvalidTimeFormat(t *testing.T) { ... }

// Table-driven tests 用于多场景
func TestCreateExam(t *testing.T) {
    tests := []struct {
        name     string
        input    map[string]interface{}
        wantCode int
        wantMsg  string
    }{
        {"正常创建", validInput, 0, ""},
        {"标题为空", noTitle, 1, "title is required"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { ... })
    }
}
```

## 3. 本项目特定的测试注意事项

### 数据库
- 测试用 SQLite in-memory 或 mock，不连真实数据库
- tester 相关必须查 `el_tester` 表（非双表）
- 如果用 sqlmock，GORM 的 `Table("el_tester")` 调用需要 mock 对应的 SQL

### JSON 请求体
- 前端可能发送 `""` 空字符串给 int/bool 字段（§5.1.2）→ 测试必须覆盖此场景
- request struct 禁止嵌入 model struct（§5.1.1）→ 测试用独立 struct

### 时间处理
- 必须用 `time.ParseInLocation`（§5.2.3）→ 测试覆盖 UTC vs Local 差异
- 测试多种时间格式输入：`2026-04-21 12:00:00` 和 `2026-04-21T12:00:00.000Z`

### 响应格式
- qu/exam/paper/candidate → `response.Rest()` → 检查 `{code:0, msg:"", data:{...}, success:true}`
- tester CRUD → `response.AjaxOK()` → 检查 `{code:200, msg:"操作成功", data:{...}}`
- tester 列表 → `response.Table()` → 检查 `{code:200, rows:[...], total:N}`

### GORM 陷阱
- `Save()` 覆盖所有字段（§5.3.5）→ 测试验证 create_time 不被覆盖
- `var rows []T` → nil → JSON = `null` → 测试必须验证空列表返回 `[]` 非 `null`

## 4. 测试结构模板

```go
package handler

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    // 注册被测路由
    return r
}

func TestXxx_Success(t *testing.T) {
    router := setupTestRouter()

    body, _ := json.Marshal(map[string]interface{}{
        "field": "value",
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/path", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", w.Code)
    }

    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)

    if resp["code"].(float64) != 0 {
        t.Errorf("expected code 0, got %v", resp["code"])
    }
}
```

## 5. 必须覆盖的测试场景（按优先级）

| 优先级 | 场景 | 示例 |
|--------|------|------|
| P0 | 正常流程 | 创建成功、查询有数据 |
| P0 | 权限校验 | 未登录被拒绝 |
| P0 | 输入验证 | 必填字段为空、格式错误 |
| P1 | 边界条件 | 空列表、超长字符串、零值 |
| P1 | 业务规则 | validateClosedExam、scoreA+scoreB=5 |
| P2 | 错误路径 | DB 连接失败、文件不存在 |
| P2 | 并发安全 | 计数器原子性、map 竞态 |
