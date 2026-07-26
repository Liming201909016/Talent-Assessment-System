---
applyTo: "Go-based Refactored System/internal/**/*.go"
description: "架构专家：审查 Go 后端代码的架构质量、分层规范、耦合度"
---

# 架构审查规则

## 1. 分层规范

本项目采用扁平 handler 架构（非 DDD），但仍需遵循基本分层：

```
handler/   → HTTP 请求处理 + 业务逻辑（当前项目 handler 承担了 service 职责）
model/     → 数据结构定义（business.go = 业务表，system.go = 系统表）
router/    → 路由注册
config/    → 配置加载
middleware/ → JWT 验证、CORS、日志等中间件
pkg/       → 通用工具包（jwt、redis、response、captcha、db）
```

### 检查要点
- handler 中不应出现基础设施代码（数据库连接、Redis 初始化）
- model 中不应出现业务逻辑
- pkg/ 下的包不应 import internal/ 下的任何包（依赖方向：内 → 外）
- router.go 只负责路由注册，不包含业务逻辑

## 2. Handler 职责与大小

### 单一职责
- 每个 handler 文件对应一个业务模块（qu.go = 题目，exam.go = 考试）
- 如果一个 handler 文件超过 800 行，考虑拆分（如 tester.go → tester.go + tester_list.go + tester_excel.go + tester_score.go）

### 检查 handler 膨胀信号
- 同一个文件中混合了 CRUD + 导出 + 报告生成 + 统计 → 需拆分
- 函数超过 100 行 → 提取子函数
- 重复的数据库查询模板 → 提取到 helper

## 3. 数据库访问模式

### 必须遵守
- 使用 GORM 参数化查询，不拼接 SQL
- tester 相关操作必须查 `el_tester` 表（禁用双表 el_tester_profile + el_tester_exam）
- `Save()` 覆盖所有字段 → 更新用 `Updates()` 或先读原记录再 Save（§5.3.5）
- 批量插入用 `CreateInBatches`（§5.3.6）

### 检查 N+1 查询
- 循环中执行 DB 查询 → 改为批量查询 + map 关联
- 嵌套 for + db.Where → 改为 JOIN 或 IN 子查询

## 4. 错误处理模式

```go
// ✅ 正确：检查错误并返回
if err := db.Create(&record).Error; err != nil {
    response.Rest(c, 1, err.Error(), nil)
    return
}

// ❌ 错误：忽略错误
db.Create(&record)
```

- 每个 DB 操作必须检查 error
- 使用 response.Rest() / AjaxOK() / Table() 统一返回格式
- 不在 handler 中 panic

## 5. 配置管理

- 禁止硬编码 IP、端口、密钥、连接串
- 所有环境差异通过 application.yml / .env.* 注入
- 检查代码中的魔数（magic number）→ 提取为常量或配置

## 6. 代码重复

- 相同的分页解析逻辑出现在多个 handler → 提取公共函数
- 相同的时间格式化出现多处 → 统一 helper
- 相同的权限检查逻辑 → 提取 middleware

## 审查输出格式

```
📐 [架构] 文件 - 问题描述
   影响：耦合度增加 / 可维护性降低 / 扩展困难
   建议：具体的重构方案

📏 [规范] 文件 - 违反的编码约定
   规则：§X.X
   修复：...
```
