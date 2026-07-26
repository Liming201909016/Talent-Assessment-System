# Runtime Validation Report

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
