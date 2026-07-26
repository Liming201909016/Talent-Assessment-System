---
mode: agent
description: "模块优化：分析模块代码质量，提出改进建议并实施"
tools: ["codebase", "read_file", "grep_search", "semantic_search", "file_search", "replace_string_in_file", "create_file", "run_in_terminal"]
---

# 模块优化

对 ${input:module_name} 模块进行全面分析和优化。

## 执行流程

### Step 1：加载上下文

1. 读取 `docs/project-memory.md` — 获取模块当前状态和已知问题
2. 读取 `.github/copilot-instructions.md` — 获取编码规约

### Step 2：代码质量扫描

读取模块的所有相关文件，检查：

#### 代码异味（Code Smells）
- [ ] 函数过长（> 100 行）
- [ ] 文件过大（> 800 行）
- [ ] 重复代码（相似代码块出现 3+ 次）
- [ ] 过深嵌套（> 4 层 if/for）
- [ ] 魔数（未命名的常量值）
- [ ] 未使用的代码（dead code）
- [ ] 过长的参数列表（> 5 个参数）

#### 健壮性检查
- [ ] 错误处理是否完整（每个 DB 操作都检查 error）
- [ ] 空值处理（nil slice → JSON null 问题）
- [ ] 并发安全（全局变量、map 读写）
- [ ] 资源泄漏（未关闭的文件、连接）

#### 可维护性检查
- [ ] 命名是否清晰（函数名体现行为，变量名体现含义）
- [ ] 是否有适当的注释（复杂业务逻辑、非显而易见的设计决策）
- [ ] 是否遵循项目约定（响应格式、数据库表、时间处理）

### Step 3：生成优化方案

按影响分级输出：

```
═══════════════════════════════════════
🔧 ${input:module_name} 模块优化方案
═══════════════════════════════════════

📌 必要优化（影响正确性或安全性）
1. ...

📈 推荐优化（提升性能或可维护性）
2. ...

💡 可选优化（代码整洁度提升）
3. ...
```

### Step 4：确认后实施

对每项优化：
1. 说明修改内容和影响范围
2. 等待用户确认
3. 实施修改
4. 编译验证：`cd 'Go-based Refactored System' && go build ./cmd/server`

### Step 5：回写记忆

将优化记录更新到 `docs/project-memory.md`
