---
applyTo: "**/*.sql,**/model/*.go,**/handler/*.go"
description: "数据库专家：审查 SQL 查询、Schema 设计、索引策略、数据一致性"
---

# 数据库审查规则

## 1. Schema 设计

### 本项目表结构要点
- **el_tester 单表设计**：替代 Java 的 el_tester_profile + el_tester_exam 双表，所有 tester 操作必须查 el_tester
- **el_mbti_answer**：MBTI 答题记录，collation 必须为 utf8mb4_0900_ai_ci（避免 JOIN 失败）
- **sys_* 表**：RuoYi 系统管理表，字段名用下划线命名

### 检查要点
- NOT NULL 约束是否合理（前端可能发空值）
- 默认值是否设置（del_flag 默认 0，status 默认 0）
- 字段长度是否足够（id_number ≤ 18 字符）
- 字符集统一 utf8mb4

## 2. 索引策略

### 必须有索引的列
- 所有外键列：exam_id, paper_id, tester_id, repo_id, qu_id
- 高频查询条件：id_number, telephone, state, del_flag
- 排序列：create_time, sort
- 唯一约束：(id_number, exam_id) 在 el_tester 中

### 索引检查
- WHERE 条件中的列是否有索引
- JOIN 关联列是否有索引
- 组合查询是否需要复合索引
- 是否有冗余索引（已被复合索引覆盖的单列索引）

## 3. 查询优化

### GORM 查询规范
```go
// ✅ 参数化查询
db.Where("exam_id = ? AND del_flag = ?", examID, 0).Find(&testers)

// ❌ 字符串拼接（SQL 注入风险）
db.Where(fmt.Sprintf("exam_id = %d", examID)).Find(&testers)

// ✅ 只查需要的列
db.Select("id, name, telephone").Find(&testers)

// ❌ SELECT * 查所有列（浪费带宽和内存）
db.Find(&testers)
```

### 分页查询
- COUNT 和 SELECT 共用 WHERE 条件
- OFFSET 大于 10000 时考虑 keyset pagination
- pageSize 上限 500

### 避免慢查询
- 不在 WHERE 条件的列上使用函数：`WHERE DATE(create_time) = '2026-04-21'` ← 索引失效
- 改用范围查询：`WHERE create_time >= '2026-04-21' AND create_time < '2026-04-22'`
- LIKE 前缀匹配可用索引：`LIKE 'abc%'`，模糊匹配不可：`LIKE '%abc%'`

## 4. 数据一致性

### 事务使用
- 涉及多表操作必须用事务：`db.Transaction(func(tx *gorm.DB) error { ... })`
- 考试创建（el_exam + el_exam_repo + el_exam_depart）必须在同一事务中
- 删除操作检查外键关联（删除 exam 前检查是否有 tester/paper 引用）

### 软删除
- 系统使用 del_flag 软删除（0=正常，2=删除）
- 查询时必须加 `del_flag = 0` 条件（或 GORM 的 `Scopes`）
- 注意：部分旧数据 del_flag 可能为 NULL（§5.1.2 相关）

## 5. SQL 脚本规范（§5.5）

- MySQL Safe Update Mode：DELETE/UPDATE 必须包含主键列条件
- 幂等脚本：`SET NAMES utf8mb4; BEGIN; DELETE; INSERT; COMMIT;`
- 文件编码：UTF-8 无 BOM（PowerShell 默认加 BOM 会导致 MySQL 报错）

## 6. 迁移安全

- ALTER TABLE 操作在大表上可能锁表 → 评估影响行数
- 新增列应有默认值，避免 NOT NULL 无默认值导致旧数据报错
- 索引创建用 `CREATE INDEX ... ALGORITHM=INPLACE` 减少锁

## 审查输出格式

```
🗄️ [数据库-严重] 问题描述
   影响：数据丢失 / 性能下降 / 一致性破坏
   当前 SQL：...
   优化 SQL：...

📊 [数据库-优化] 问题描述
   建议：添加索引 / 优化查询 / 调整 Schema
```
