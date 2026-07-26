---
description: '提示词优化指令：自动读取项目记忆和规则，增强提示词的上下文和约束'
---

# 提示词优化指令

当此指令被附加到对话中时，执行以下增强流程：

## 1. 上下文加载（必须先做）

在处理用户请求之前，先读取：
- `docs/project-memory.md`（项目事实、模块状态、已知问题）
- `.github/copilot-instructions.md`（三区结构、Go编码规约 §5.1-§5.7、文件归属规则）
- 用户记忆中的 api-contracts / deployment-discipline / test-strategy

## 2. 意图分析

识别用户请求涉及的：
- **模块**（qu/exam/paper/tester/candidate/mbti/auth/system/前端/部署）
- **操作类型**（新增/修 bug/重构/测试/部署/分析）
- **关键文件**（根据下方映射表推断 handler/model/router/前端组件路径）
- **已知陷阱**（从 Go编码规约和项目记忆中匹配）

## 3. 自动补充约束

在执行任务时，自动应用以下约束：

### 文件路径映射
- qu → handler/qu.go, qu_excel.go | model/business.go | views/qu/
- exam → handler/exam.go, exam_pdf.go | model/business.go | views/exam/
- tester → handler/tester.go, tester_*.go | model/business.go | views/exam/
- mbti → handler/mbti.go, mbti_report.go | model/business.go | views/exam/mbtiExam.vue
- auth → handler/auth.go | views/login.vue
- system → handler/ruoyi_system.go, ruoyi_crud.go | model/system.go | views/system/

### 响应格式
- qu/exam/paper/candidate → response.Rest()
- tester CRUD → response.AjaxOK()
- tester 列表 → response.Table()
- 系统管理 → response.AjaxOK()

### 数据库表
- tester → **el_tester**（禁用 el_tester_profile + el_tester_exam 双表）
- mbti → el_mbti_answer
- exam → el_exam + el_exam_repo + el_exam_depart

### 高频编码规约
1. 禁止 request struct 嵌入 model struct（§5.1.1）
2. 前端空字符串 "" → Go int/bool 报错，用 interface{} 或 *int（§5.1.2）
3. 必须 time.ParseInLocation，不用 time.Parse（§5.2.3）
4. Save() 覆盖所有字段 → 更新用 Updates()（§5.3.5）
5. 文件路径校验防穿越（§5.6.11）
6. Content-Disposition 中文用 RFC 5987（§5.6.12）

## 4. 执行前确认

如果任务涉及部署，必须：
- 只做本机编译验证，不自动 scp/ssh
- 等用户明确说"部署"才执行
- 多个修复积累后统一发布

## 5. 执行后回写

任务完成后，检查是否需要更新 `docs/project-memory.md`：
- 新发现的事实 → 新增记录
- 旧记录有误 → 追加纠正（保留原记录）
- 模块状态变更 → 更新状态
