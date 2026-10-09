# Runtime Validation Report

> **历史报告提示（2026-10-09）**：本文件主体记录2026-07至2026-10-01的特定验收批次，其中`PARTIAL/FAIL`只代表当时版本和范围。当前项目状态统一读取[project-status.md](project-status.md)；最新00401、MBTI、00501/00502 staging范围结论已由2026-10-09后续真实验证覆盖。历史失败仍保留作回归来源，不删除、不反向改写为当时PASS。

## 2026-10-01 FB-198部署后完整复测

### 结论

**总体为PARTIAL PASS：FB-198业务变更与核心构建、单测、浏览器、真实PDF和环境健康全部通过；完整测试发现1项活动模板视觉契约失败。**

| 范围 | 结果 | 证据 |
|---|---|---|
| Go全量 | PASS | `go test ./... -v -count=1`退出0；FB-198测试通过 |
| Go构建 | PASS | Windows build退出0；staging Linux后端SHA=`ee4566e7...` |
| 前端Vitest | PASS | 26文件、165项全部通过 |
| 前端production build | PASS WITH WARNINGS | 构建成功；仅asset/entrypoint体积2类既有warning |
| 当前题本/身份契约 | PASS | `COMPETENCY_PHASE1_CONVERTER_CONTRACT_TEST_PASS`、`COMPETENCY_PHASE1_IDENTITY_TEST_PASS` |
| 参与者浏览器回归 | PASS | 390/768/1440响应式、答案保存/刷新、认证/交卷门禁3/3通过；API为浏览器route mock，未写staging数据 |
| 260915模板契约 | PASS | 75控件/58唯一Tag/12图表/0外链 |
| v2重算验证器契约 | PASS | `COMPETENCY_V2_RECOMPUTE_CONTRACT_TEST_PASS` |
| FB-198真实报告 | PASS | 优势自律性/成就导向/计划执行，待发展敬业奉献/逻辑思维；五段全文匹配DB；A4 10页、DB/文件/SHA一致 |
| 活动v2模板视觉层级契约 | **FAIL** | 本地模板和精确下载的staging模板均在`test_fb194_overview_visual_hierarchy`失败；真实概览环图中心显示`60.94`，缺少契约要求的`总体得分`标签 |
| Staging终验 | PASS | talent-assessment/nginx/mysql active；内外health正常；关键错误0、Nginx 5xx=0、短时会话0、临时文件0 |

旧工作区任务“Full Regression (7 suites)”已过期：其引用的`chain-batch.js`、`business-rules-test.js`、`exam-fields-test.js`、`requirement-tests.js`、`browser-e2e-v2.js`、`exam-form-candidate-test.js`、`screenshot-chain-test.js`在活动测试目录均不存在，7项均为`MODULE_NOT_FOUND`。本轮改用当前受维护的Go/Vitest、`scripts/test/package.json`契约入口、3条参与者Playwright及v2模板/报告门禁；没有把失效旧任务记作业务失败。

**Generated**: 2026-07-26T10:45:00+08:00  
**Target**: staging `http://20.200.136.133` — 胜任力答题页 UI / 移动端适配

## Summary

| Step | Status | Exit Code | Details |
|------|--------|-----------|---------|
| Startup/readiness | PASS | 0 | `GET /prod-api/health` → HTTP 200 `{"status":"ok"}` |
| Browser capability | PASS | 0 | Playwright Chromium 可启动；使用临时 Node.js 20（系统 Node.js 16 不满足当前 Playwright 最低版本） |
| Responsive E2E | PASS | 0 | 390×844、768×1024、1440×900 三视口均无横向溢出 |
| Answer persistence E2E | PASS | 0 | 第9题选择→保存请求→统计9/31→题号已答→刷新恢复选择 |
| Auth/submit guard E2E | PASS | 0 | 缺token阻断API；未答交卷定位首道未答；补答后确认交卷、清理token并跳转完成页 |

**Overall**: PASS（限本次已部署胜任力答题页 UI E2E 范围）

## Environment

- Docker: UNAVAILABLE — 本机未找到 Docker CLI；本轮直接验证已运行的 staging，不需要本地容器基础设施。
- Node.js: system `v16.20.2`；通过 `npx node@20` 执行测试。
- Playwright: AVAILABLE — Chromium headless 启动检查退出码 0。
- infra-tier: staging 真实 nginx/静态资源与健康端点；答题API使用浏览器 route mock，避免改写正式参与者数据。
- browser-tier: PRIMARY（Playwright Chromium）。

## E2E Flows

### 1. Responsive layout

文件：`scripts/test/competency-mobile-ui-test.js`

- 手机390px：1列选项、5列题号、选项48px、题号44px、无横向溢出。
- 平板768px：3列选项、13列题号、无横向溢出。
- 桌面1440px：5列选项、21列题号、无横向溢出。

### 2. Answer save and reload

文件：`scripts/test/competency-answer-flow-e2e.js`

- 打开40题试卷并定位第9道未答题。
- 选择“非常符合”，断言仅发送一次保存请求且载荷正确。
- 断言已答/未答更新为9/31，题号状态切换为已答。
- 刷新页面后重新读取试卷，选择值保持，且未产生重复保存请求。

### 3. Authentication and submit guard

文件：`scripts/test/competency-submit-guard-e2e.js`

- 缺少paper token时返回上一页，试卷详情请求数为0。
- 尚有2题未答时点击交卷，提交请求数为0并定位第2题。
- 补答第2、3题后确认交卷，只发送一次manual提交。
- 提交成功后跳转 `/exam/thank-you`，session token被移除。

## Final Evidence

```text
competency-mobile-ui-test.js=0
competency-answer-flow-e2e.js=0
competency-submit-guard-e2e.js=0
STAGING_HEALTH={"status":"ok"} HTTP=200
```

## Fix Loop

- 答案统计首次断言受Element图标与空白文本影响；改为读取统计数字节点后通过。
- 交卷测试跨页面重载模拟状态三次超时；经用户确认，改为同一页面完成剩余题再交卷，最终通过。
- 上述均为测试代码问题，未修改生产业务代码。

## Known Gaps

- 本轮目标是刚部署的胜任力答题页 UI，不代表全系统管理端、传统001/002/003或正式报告链的完整浏览器回归。
- 为保护已到期的真实参与者结果，答题详情、保存和提交接口均在浏览器侧模拟；本轮没有写 staging 数据库。
- 后端真实持久化、并发提交、到期Worker和传统链已有历史 staging 验收记录，本轮未重复执行。
