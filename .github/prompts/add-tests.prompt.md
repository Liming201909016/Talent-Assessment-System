---
mode: agent
description: "生成单元测试：为指定的 Go handler 或 pkg 生成完整的单元测试"
tools: ["codebase", "read_file", "grep_search", "file_search", "create_file", "replace_string_in_file"]
---

# 生成单元测试

为 ${input:target_file} 生成完整的单元测试。

## 执行流程

### Step 1：分析被测文件

1. 读取目标文件，提取所有公开函数签名
2. 分析每个函数的：
   - 输入参数和返回值
   - 数据库依赖（GORM 调用）
   - 外部依赖（Redis、文件系统）
   - 业务规则和边界条件

### Step 2：加载测试规约

读取 #file:.github/instructions/go-test.instructions.md 获取：
- 命名规范（Test{函数名}_{场景}）
- 项目特定陷阱（空字符串、时间格式、Save覆盖、el_tester单表）
- 响应格式要求（Rest/AjaxOK/Table）
- 测试结构模板

### Step 3：确定测试策略

对每个函数确定：
- 正常路径测试
- 错误路径测试
- 边界条件测试
- 业务规则测试

### Step 4：生成测试文件

在同目录下创建 `{filename}_test.go`，包含：

1. **测试辅助函数**
   - `setupTestRouter()` — 创建 gin 测试路由
   - `makeRequest()` — 发送 HTTP 请求的 helper
   - Mock DB 设置

2. **Table-driven tests** — 每个函数的多场景测试

3. **必覆盖场景**（按优先级）：
   | 优先级 | 场景类型 | 说明 |
   |--------|---------|------|
   | P0 | 正常流程 | 传入有效数据，验证返回正确 |
   | P0 | 必填校验 | 缺少必填字段，验证返回错误 |
   | P0 | 权限校验 | 未携带 token，验证 401 |
   | P1 | 空字符串 | int/bool 字段传 ""（§5.1.2） |
   | P1 | 时间格式 | 多种时间格式输入（§5.2.4） |
   | P1 | 业务规则 | 模块特定的业务约束 |
   | P2 | 空列表 | 查询无数据返回 [] 非 null |
   | P2 | 并发 | atomic 计数器、map 竞态 |

### Step 5：验证测试可编译

确认：
- import 路径正确
- 使用 `testing` 标准库
- mock 数据符合数据库约束
- 测试数据不依赖外部环境

### Step 6：输出

1. 创建测试文件
2. 列出测试覆盖矩阵（函数 × 场景）
3. 提示运行命令：`cd Go-based\ Refactored\ System && go test ./internal/handler/ -run TestXxx -v`
