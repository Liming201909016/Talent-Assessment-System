---
applyTo: "**/*.go"
description: "安全专家：编辑 Go 文件时自动检查 OWASP Top 10 安全风险"
---

# 安全审查规则（自动附加到 Go 文件编辑）

当编辑或审查 Go 代码时，检查以下安全风险：

## 1. 注入类（OWASP A03）

### SQL 注入
- 禁止字符串拼接 SQL：`fmt.Sprintf("SELECT ... WHERE name = '%s'", input)` ← 危险
- 必须用 GORM 参数化：`db.Where("name = ?", input)` 或 `db.Raw("... WHERE id = ?", id)`
- `db.Exec()` 中的 SQL 必须用 `?` 占位符

### 命令注入
- 禁止 `exec.Command("sh", "-c", userInput)`
- 如果必须调用外部命令，参数必须白名单校验

## 2. 路径穿越（OWASP A01）

- 文件上传/下载 API 必须校验路径在允许目录内
- 使用 `filepath.Clean()` + `strings.HasPrefix()` 校验：
```go
clean := filepath.Clean(userPath)
if !strings.HasPrefix(clean, allowedDir) {
    return errors.New("path not allowed")
}
```
- 本项目已知规则（§5.6.11）：上传/下载 API 必须校验路径在允许目录内

## 3. 认证与授权（OWASP A07）

- JWT token 必须验证签名和过期时间
- 本项目 JWT：HS512，密钥来自配置文件，30分钟过期
- 检查 middleware 是否正确拦截未认证请求
- 敏感操作（删除、批量操作）需要检查权限

## 4. 敏感数据暴露（OWASP A02）

- 密码字段不能出现在 API 响应中（除非业务需要，如 tester 的 password 列表显示）
- 日志中不打印密码、token、密钥
- 错误消息不暴露内部实现（不返回 stack trace）

## 5. 输入验证

- 前端可能发送空字符串 ""（§5.1.2），Go int/bool 会 unmarshal 失败
- 分页参数 pageNum/pageSize 必须有上限（防止 `pageSize=999999`）
- ID 参数必须校验为正整数

## 6. HTTP 安全头

- Content-Disposition 中文文件名用 RFC 5987 编码（§5.6.12）
- 下载响应设置 `Content-Type` 为准确的 MIME 类型
- CORS 配置不能为 `*`（生产环境）

## 7. 并发安全

- 全局计数器（如 nextID）必须用 `atomic` 或 `sync.Mutex`
- 避免 map 的并发读写（Go map 非线程安全）

## 审查输出格式

发现问题时按以下格式报告：

```
🔴 [严重] 文件:行号 - 问题描述
   风险：可能导致 XXX
   修复：建议的修复代码

🟡 [中等] 文件:行号 - 问题描述
   风险：...
   修复：...

🟢 [低] 文件:行号 - 问题描述
   建议：...
```
