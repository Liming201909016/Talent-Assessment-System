---
mode: agent
description: "全专家审查：对指定模块运行安全、架构、性能、数据库、UI 五维审查"
tools: ["codebase", "read_file", "grep_search", "semantic_search", "file_search"]
---

# 全专家审查

对 ${input:module_name} 模块执行五维专家审查。

## 执行流程

### Step 1：加载上下文

1. 读取 `docs/project-memory.md` 获取模块当前状态
2. 读取 `.github/copilot-instructions.md` 获取编码规约

### Step 2：定位模块代码

根据模块名定位相关文件：

| 模块 | 后端文件 | 前端文件 |
|------|---------|---------|
| qu | internal/handler/qu.go, qu_excel.go | src/views/qu/ |
| exam | internal/handler/exam.go, exam_pdf.go | src/views/exam/ |
| tester | internal/handler/tester.go, tester_*.go | src/views/exam/ |
| paper | internal/handler/paper.go | src/views/exam/ |
| candidate | internal/handler/candidate.go | — |
| mbti | internal/handler/mbti.go, mbti_report.go | src/views/exam/mbtiExam.vue |
| auth | internal/handler/auth.go | src/views/login.vue |
| system | internal/handler/ruoyi_system.go, ruoyi_crud.go | src/views/system/ |

### Step 3：逐一执行专家审查

读取并应用以下指令文件的规则，对模块代码执行审查：

1. **安全审查** — 参考 #file:.github/instructions/security-review.instructions.md
   - SQL 注入、路径穿越、输入验证、敏感数据暴露

2. **架构审查** — 参考 #file:.github/instructions/architecture-review.instructions.md
   - 分层规范、Handler 职责、代码重复、错误处理

3. **性能审查** — 参考 #file:.github/instructions/performance-review.instructions.md
   - N+1 查询、内存使用、缓存策略、并发安全

4. **数据库审查** — 参考 #file:.github/instructions/database-review.instructions.md
   - 查询优化、索引策略、事务使用、数据一致性

5. **UI/UX 审查**（仅前端模块）— 参考 #file:.github/instructions/ui-ux-review.instructions.md
   - 表单交互、响应式、状态反馈、无障碍性

### Step 4：汇总输出

按严重程度分级输出所有发现：

```
═══════════════════════════════════════
📋 ${input:module_name} 模块 — 全专家审查报告
═══════════════════════════════════════

🔴 P0（必须修复）
1. [安全] ...
2. [性能] ...

🟡 P1（建议修复）
3. [架构] ...
4. [数据库] ...

🟢 P2（可优化）
5. [UX] ...
6. [架构] ...

📊 总计：X 项发现（P0: X, P1: X, P2: X）
```

### Step 5：确认下一步

询问用户：
- 是否要修复 P0 问题？
- 是否要逐一处理 P1 问题？
- 是否要记录到 project-memory.md？
