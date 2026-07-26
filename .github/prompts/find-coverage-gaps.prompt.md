---
mode: agent
description: "找出指定模块的测试盲区：分析 handler 函数的所有分支，对比 business-branches.md 矩阵，输出待补测试清单"
tools: ["codebase", "read_file", "grep_search", "semantic_search", "file_search", "create_file"]
---

# 找出测试盲区

为 ${input:module_name} 模块生成"测试盲区报告"。

## 执行流程

### Step 1：确定模块范围

根据 module_name 找出涉及的文件：

| 输入 | 范围 |
|------|------|
| handler 名（如 mbti） | internal/handler/mbti.go + mbti_*.go |
| 文件路径 | 该单一文件 |
| "all" | 所有 internal/handler/*.go |

### Step 2：解析所有函数和分支

读取目标文件，对每个**导出的 Handler 函数**（首字母大写）执行：

1. 列出函数签名
2. 解析函数体中的分支：
   - `if` 条件
   - `switch` / `case`
   - `for range` 中的过滤逻辑
   - `return` 不同结果的路径
3. 列出函数的所有可能输入维度：
   - 请求体字段（每个字段的可能值）
   - URL 参数 / Query 参数
   - 数据库前置状态
   - 配置/环境变量

### Step 3：生成业务分支矩阵（笛卡尔积）

将输入维度做笛卡尔积，列出每个组合的期望行为。

**注意**：完整笛卡尔积可能爆炸，按以下规则裁剪：
- 标识"等价类"（如 `name = "abc"` 和 `name = "xyz"` 等价）
- 标识"必失败"组合（如 `paperID = ""`）只测一个代表
- 保留所有边界值（0, 最大值, 负数, 空字符串）
- 保留所有"前端可能发的异常输入"（§5.1.2）

### Step 4：与现有矩阵比对

读取 `docs/business-branches.md`：

- 找到该模块的现有矩阵
- 标记每行：
  - ✅ 已在矩阵中且已测试
  - ❌ 已在矩阵中但未测试
  - 🆕 矩阵中没有，本次新发现

### Step 5：与现有测试比对

读取该模块的 `_test.go` 文件：

- 提取所有测试函数名 → 推断每个测试覆盖的分支
- 在矩阵中标记 ✅
- 找出真正未测试的分支

### Step 6：与已知陷阱比对

读取 `.github/copilot-instructions.md` §5.1-5.7，检查矩阵是否包含已知陷阱：

| 陷阱 | 检查项 |
|------|------|
| §5.1.1 struct 嵌套 | 矩阵是否有"前端嵌套 JSON 字段"分支 |
| §5.1.2 空字符串 | 每个 int/bool 字段是否有 `""` 分支 |
| §5.2.3 时间解析 | 时间字段是否有"多种格式"分支 |
| §5.3.5 Save 覆盖 | Update 是否有"create_time 保留"分支 |
| §5.6.11 路径校验 | 文件操作是否有"路径穿越"分支 |

### Step 7：生成报告

输出格式：

```markdown
# 🔍 测试盲区报告 — ${input:module_name}

## 📊 概览

| 指标 | 数值 |
|------|:---:|
| 函数数 | X |
| 业务分支总数（含新发现） | X |
| 已覆盖 ✅ | X (X%) |
| 未覆盖 ❌ | X (X%) |
| 新发现 🆕 | X |
| 高风险 🔥 | X |

## 🔥 P0 - 必须立即补测试（X 项）

### 1. 函数 X.Y - 分支描述
- **风险**：{为什么是 P0}
- **触发条件**：{...}
- **期望行为**：{...}
- **建议测试代码**：
  \`\`\`go
  func TestXY_BranchName(t *testing.T) {
      // ...
  }
  \`\`\`

## ⚠️ P1 - 建议补测试（X 项）

...

## 🆕 新发现的分支（X 项）

> 这些分支在 business-branches.md 中没有，需追加到矩阵：

```markdown
| #N | 字段A | 字段B | 期望行为 | 测试 |
|:--:|---|---|------|:---:|
| ... | ... | ... | ... | ❌ |
```

## 🛠️ 同步更新建议

1. 在 docs/business-branches.md 中追加 X 行（已生成上方 Markdown 块）
2. 在 docs/regression-tests.md 中将 X 个待补测试加入 backlog
3. 更新 docs/coverage-history.md 记录本次扫描

## 🔄 下次迭代

```
请按此报告补充测试：
- P0 全部补充
- P1 至少补 X 项
完成后重新执行 /find-coverage-gaps ${input:module_name} 验证
```
```

### Step 8：询问用户

输出报告后询问：

1. 是否要立即生成 P0 测试代码？
2. 是否要将矩阵更新写入 business-branches.md？
3. 是否要将待补项追加到 regression-tests.md backlog？

## 实施提示

- 不要一次性生成所有测试代码（避免 Copilot 输出超长）
- 只在用户确认 P0 项后再生成代码
- 利用现有 jwt_test.go / response_test.go / scoring_test.go 的风格作为模板
