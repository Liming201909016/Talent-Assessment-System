# User Feedback Log

| ID | Date | Feedback | Classification | Reproduction | Regression | Scope | Symptom | Trigger Data | Status | Test/Fix |
|----|------|----------|----------------|--------------|------------|-------|---------|--------------|--------|----------|
| UF-001 | 2026-07-25 | 题库列表看不到编号00401的胜任力测验题库及其题目 | 功能缺口 / UI交互 | admin进入“测评管理→题库管理” | 从未显示过 | 当前admin；功能设计对所有管理员一致 | 现有列表仅6个传统题库，无00401入口 | code=00401，名称=胜任力测验题库，384道胜任力题 | ✅ staging已修复 | `TestFeedbackUF001_CompetencyQuestionBank00401IsReachable`; `competency-question-list.spec.js`; staging API/browser 384题 |
| UF-002 | 2026-07-26 | 胜任力测评预览页点击“开始测评”后无法进入答题，重复提示英文 `competency exam is not published` | 业务规则 / UI交互 | 受测者进入准备页 → 点击“开始测评” | 第一次测试即失败 | 根因影响所有未发布胜任力草稿；具体样本为00401 ABC | 页面连续出现两条相同英文错误通知，未跳转答题页 | 实际examId=`1785027744745375431`，participantId=`1785027772270618331` | ✅ staging已修复 | FB-066；未发布草稿从在线列表隐藏、准备页禁用开始并中文提示、后端中文映射、防重复点击；该测评已获授权发布5维度40题，浏览器验收通过 |

## Classification Summary

- 功能缺口 / UI交互：1
- 业务规则 / UI交互：1（已修复）
