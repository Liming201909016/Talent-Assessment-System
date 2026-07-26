---
mode: agent
description: "迭代循环：运行测试 → 专家审查 → 修复问题 → 回归验证 → 更新记忆的完整迭代流程"
tools: ["codebase", "read_file", "grep_search", "semantic_search", "file_search", "replace_string_in_file", "create_file", "run_in_terminal"]
---

# 迭代循环

执行一次完整的质量迭代：测试 → 审查 → 修复 → 验证 → 记录。

## 前置条件

- Go 后端已启动（port 8092）
- 前端 dev server 已启动（port 8089）
- 数据库 SSH tunnel 已连接（port 23306）

## 执行流程

### Phase A：运行测试套件

#### A1. Go 单元测试
```bash
cd "Go-based Refactored System" && go test ./... -v -count=1
```

#### A2. API 集成测试（快速回归）
```bash
node scripts/test/chain-batch.js
node scripts/test/business-rules-test.js
```

#### A3. 需求验证
```bash
node scripts/test/requirement-tests.js
```

#### A4. 收集结果
- 记录通过/失败数量
- 记录失败的测试名称和错误信息
- 如果有失败测试，进入 Phase B 分析

### Phase B：分析失败原因

对每个失败的测试：
1. 读取测试代码，理解测试预期
2. 读取被测 handler 代码
3. 定位 root cause（是代码 bug 还是测试数据问题）
4. 分类：`[BUG]` 代码问题 / `[TEST]` 测试需更新 / `[ENV]` 环境问题

### Phase C：专家审查（可选）

如果 Phase A 全部通过，执行预防性审查：

1. 检查最近修改的文件（`git diff --name-only HEAD~5`）
2. 对修改过的文件运行安全审查（参考 #file:.github/instructions/security-review.instructions.md）
3. 对修改过的文件运行性能审查（参考 #file:.github/instructions/performance-review.instructions.md）

### Phase D：修复

按优先级修复：
1. **P0 — 安全漏洞**：立即修复
2. **P1 — 功能 Bug**：修复代码，编译验证
3. **P2 — 优化建议**：记录到 project-memory.md，不立即修复

每个修复后：
- `go build ./cmd/server` 编译验证
- 不自动部署（遵循部署纪律）

### Phase E：回归验证

修复完成后，重新运行 Phase A 的测试：
- 之前失败的测试应通过
- 之前通过的测试不应回退

### Phase F：更新记忆

将本次迭代结果写入 `docs/project-memory.md`：

```markdown
### YYYY-MM-DD 迭代记录

#### 测试结果
- Go 单元测试：X/Y 通过
- API 集成测试：X/Y 通过
- 需求验证：X/Y 通过

#### 发现并修复
- [BUG] ...
- [安全] ...

#### 待优化（下次迭代）
- [P2] ...
```

### Phase G：输出迭代报告

```
═══════════════════════════════════════
📊 迭代报告 — YYYY-MM-DD
═══════════════════════════════════════

✅ 测试：X/Y 通过（修复前 A/B）
🔧 修复：X 项（P0: X, P1: X）
📝 待办：X 项（记录在 project-memory.md）
⏱️ 耗时：~X 分钟

下次迭代建议重点：
1. ...
2. ...
```
