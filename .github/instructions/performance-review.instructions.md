---
applyTo: "**/*.go"
description: "性能专家：审查 Go 代码的性能问题，包括 N+1 查询、内存泄漏、缓存缺失"
---

# 性能审查规则

## 1. 数据库查询性能

### N+1 查询（最常见）
- 循环中执行 `db.Where().Find()` → 改为批量 `WHERE id IN (?)` + map 关联
- 嵌套 for 循环中有 DB 调用 → 合并为 JOIN 查询
- 标志性代码模式：
```go
// ❌ N+1 问题
for _, exam := range exams {
    var testers []Tester
    db.Where("exam_id = ?", exam.ID).Find(&testers)  // N 次查询
}

// ✅ 批量查询
var allTesters []Tester
db.Where("exam_id IN ?", examIDs).Find(&allTesters)  // 1 次查询
testerMap := groupBy(allTesters, "ExamID")
```

### 缺失索引
- WHERE 条件中的列应有索引
- 高频查询列：exam_id, paper_id, tester_id, repo_id, id_number, telephone
- JOIN 的关联列必须有索引
- ORDER BY 的列如果不在索引中，大数据量时会全表排序

### 分页性能
- 大偏移量分页 `OFFSET 10000 LIMIT 10` 性能差 → 考虑 keyset pagination
- `pageSize` 必须有上限（建议 ≤ 500），防止 `pageSize=999999`
- COUNT(*) 和 SELECT 应尽量共用 WHERE 条件

## 2. 内存使用

### 大数据集处理
- 导出 Excel 时加载全部数据到内存 → 如果数据量可能超过 10000 行，使用流式写入
- 批量插入用 `CreateInBatches(records, 100)`（§5.3.6）而非一次性 Create
- 文件上传/下载用 `io.Copy` 流式传输，不 `ioutil.ReadAll`

### 内存泄漏信号
- 未关闭的 `*os.File`、`*sql.Rows`、`io.ReadCloser`
- goroutine 泄漏：启动的 goroutine 没有退出条件
- 缓存没有 TTL 或 size limit

## 3. 并发性能

### 连接池
- 数据库连接池参数是否合理（MaxOpenConns, MaxIdleConns, ConnMaxLifetime）
- Redis 连接池是否配置

### goroutine 安全
- 全局 map 的并发读写 → 用 `sync.Map` 或 `sync.RWMutex`
- 全局计数器 → 用 `atomic.AddInt64`

## 4. HTTP 性能

### 响应大小
- 列表 API 是否返回了不必要的大字段（如完整的题目内容在列表中）
- 文件下载是否设置了正确的 Content-Length
- 是否启用了 Gzip 压缩

### 缓存策略
- 字典数据（dict）、菜单（getRouters）等低频变化数据应缓存
- 静态资源（模板文件）应设置 Cache-Control

## 5. 编译优化

- 检查是否有 `fmt.Sprintf` 用于日志但日志级别被禁用（浪费 CPU）
- 字符串拼接循环中用 `strings.Builder` 替代 `+`
- `json.Marshal/Unmarshal` 对性能敏感路径考虑 `json-iterator`

## 审查输出格式

```
⚡ [性能-严重] 文件:行号 - 问题描述
   影响：QPS 下降 / 内存溢出 / 响应延迟
   当前：~Xms / ~X MB
   优化后：~Xms / ~X MB
   方案：具体代码修改

⏱️ [性能-优化] 文件:行号 - 可优化项
   建议：...
```
