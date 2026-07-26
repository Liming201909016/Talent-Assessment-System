---
mode: agent
description: "安全审计：对指定模块执行完整的安全审计，输出漏洞报告和修复方案"
tools: ["codebase", "read_file", "grep_search", "semantic_search", "file_search", "replace_string_in_file"]
---

# 安全审计

对 ${input:target} 执行完整的安全审计。

## 执行流程

### Step 1：确定审计范围

- 如果 target 是模块名（qu/exam/tester/...）→ 审计该模块所有 handler + 路由 + 前端
- 如果 target 是文件路径 → 审计该文件
- 如果 target 是 "all" → 审计所有 handler 文件

### Step 2：加载安全规则

读取 #file:.github/instructions/security-review.instructions.md

额外加载项目已知的安全修复记录（从 `docs/project-memory.md` 的"已修复的 Bug"部分）：
- Bug #7: List 泄露 password → json:"-"（已修复，验证是否回退）
- Bug #8: exam State 接受无效值 → 0-3 范围校验（已修复）
- Bug #9: PdfUpload 任意文件读取 → 路径限制（已修复）

### Step 3：逐项检查

#### A. 注入攻击
搜索代码中的危险模式：
```
grep: fmt.Sprintf.*SELECT|INSERT|UPDATE|DELETE
grep: db.Raw.*\+.*input|db.Exec.*\+.*input
grep: exec.Command
grep: os.Open.*\+.*user
```

#### B. 认证与授权
- 检查 router.go 中哪些路由在 JWT middleware 外（应仅限 login/captcha/公开API）
- 检查是否有路由绕过 JWT 的风险
- 检查 admin 与普通用户的权限隔离

#### C. 输入验证
- 所有 `c.ShouldBindJSON` 后是否有字段验证
- 文件上传是否限制了类型和大小
- 分页参数是否有上限
- ID 参数是否校验为正整数

#### D. 敏感数据
- 搜索密码、密钥、token 是否出现在日志或响应中
- 检查 .env 文件是否在 .gitignore 中
- 检查错误响应是否暴露内部信息

#### E. HTTP 安全
- CORS 配置是否过于宽松
- 文件下载的 Content-Type 是否正确
- 是否设置了安全响应头（X-Content-Type-Options, X-Frame-Options）

### Step 4：输出漏洞报告

```
═══════════════════════════════════════
🔒 安全审计报告 — ${input:target}
═══════════════════════════════════════

🔴 高危（立即修复）
1. [A03-注入] 文件:行号
   描述：...
   PoC：...
   修复方案：...

🟡 中危（尽快修复）
2. [A07-认证] 文件:行号
   描述：...
   修复方案：...

🟢 低危（建议修复）
3. [A05-配置] 文件:行号
   描述：...
   建议：...

📊 总计：X 个漏洞（高危 X, 中危 X, 低危 X）
```

### Step 5：自动修复

询问用户是否要自动修复高危漏洞。确认后：
1. 逐个修复
2. 修复后重新扫描验证
3. 更新 project-memory.md 记录修复
