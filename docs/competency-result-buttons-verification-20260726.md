# UF-003 / FB-075 胜任力结果按钮分析与 E2E 验收

**日期**：2026-07-26  
**截图样本**：测评 `1785060945494990462`，答卷 `fc743d2b-0b72-49b2-803f-f285d62730ed`，受测者“五配置受测者05”  
**环境**：staging `20.200.136.133` 已部署入口分流 + 专用胜任力结果页真实 API/PDF

## 1. 错误根因

测评管理列表的“详情”按钮无条件跳转传统参与者/报告页面。胜任力测评没有传统题库 `repoCode`，但传统页面的“批量生成报告”调用 `/exam/api/exam/exam/generate-report`。传统生成器依赖 `repoCode=001/002/003` 选择报告模板，收到空值后返回“不支持的 repoCode”。

这不是报告文件损坏，而是**测评类型分流遗漏**：胜任力详情错误进入了传统业务链。

本地修复后：

- `assessmentType=competency` 的“详情”进入专用 `CompetencyResults`；
- `assessmentType=legacy` 继续进入原 `ListExamUser`；
- 胜任力页只调用 `/exam/api/competency/*` 结果与报告接口。

## 2. 截图按钮判定

| 截图控件 | 胜任力判定 | 处理结果 |
|---|---|---|
| 查询 / 重置 / 姓名电话状态筛选 | 属于传统参与者详情页 | 胜任力专用页使用排序指标、方向和维度筛选并自动查询，不再进入该页 |
| 批量生成报告 | **错误**，调用传统生成器，复现截图报错 | 不在胜任力页展示；使用每行“PDF→生成临时PDF” |
| 批量下载 | 通用候选人/测评人员 `pdf_path` 打包，不是胜任力报告实例契约 | 不在胜任力页展示 |
| 批量下载简版 | 仅 MBTI；截图中禁用是正确的 | 胜任力页不展示 |
| 批量删除 | 传统人员停用操作，不是胜任力结果操作 | 胜任力页不展示，避免绕过整链规则 |
| 删除报告 | 只清人员 `pdf_path/pdf_flag`，不会同步删除胜任力报告实例/审计 | 胜任力页不展示，避免状态不一致 |
| 查看团队报告 / 下载团队报告 | 传统团队报告链，胜任力未定义团队报告 | 胜任力页不展示 |
| 查看 | 传统 001/002/003 报告组件，空 `repoCode` 会错误分流 | 替换为胜任力“测试报告” |
| 答题详情 | 传统题库/答案结构 | 替换为胜任力“详情”：5条维度得分 + 40条逐题审计 |
| 下载 | 传统按数据库路径下载 | 替换为胜任力报告实例下载接口 |

因此截图页面并非“修每个传统按钮”即可安全复用；正确方案是按测评类型进入不同页面。

## 3. RED → GREEN

- RED：`competency-detail-routing.spec.js` 首次运行 1 失败 / 2 通过，失败项明确证明详情入口没有 `assessmentType=competency` 分流。
- GREEN：专项 3/3 通过。
- 前端全量：18 文件、107 项全部通过。
- production build：成功；仅保留既有 asset/entrypoint size 两项 warning。

## 4. Staging 全按钮 E2E

脚本：`scripts/test/staging-competency-result-buttons-e2e.js`

使用保留的完整 40 题领导人员版结果，真实执行：

1. 在测评管理按完整标题查询保留测评，并点击主“详情”；
2. 验证路由进入 `CompetencyResults`，未进入传统 `ListExamUser`；
3. 排序指标：提交时间 → 整体分 → 维度分；
4. 排序方向：升序、降序；
5. 维度选择：切换到 D02 人际交往；
6. 详情：验证 5 条维度结果；
7. 逐题审计：验证 40 条题目；
8. 测试报告：打开真实 `temp-v1` 报告页并验证免责声明；
9. PDF 生成：真实接口 HTTP 200；
10. PDF 下载：真实接口 HTTP 200、`application/pdf`、615336 bytes、文件头 `%PDF`；
11. 返回：回到测评管理页；
12. 浏览器 console：无相关 error。

输出：

```text
COMPETENCY_RESULT_BUTTONS_E2E_PASS
buttons=back|sort-metric|sort-direction|dimension-selector|detail-dimensions|detail-questions|test-report|pdf-generate|pdf-download
legacy_controls_hidden=9/9
download_pdf_bytes=615336
```

传统专属控件隐藏检查 9/9：批量生成、批量下载、批量下载简版、批量删除、删除报告、查看团队报告、下载团队报告、传统答题详情、完整版。

短时管理员 Redis 会话、状态文件和远端辅助文件均已清理。E2E 复用现有报告实例，不创建或删除业务测评数据。

## 5. 当前边界

- “详情”入口分流修复已部署 staging；前端 index SHA-256 为 `d9ffb611ccd8de4fc4ae348ba57e58a66a934565ddb7b3ae4383745fbd7abf15`，备份为 `/opt/talent-assessment/dist.bak.fb075.20260726_205341`。
- 已从真实测评管理主“详情”入口进入专用胜任力结果页，并完成全部有效按钮 E2E；截图中的旧入口行为不再复现。
- staging 后端、nginx、MySQL 均为 active，内外 health 正常，执行窗口关键错误为 0；短时会话和部署/E2E 临时文件已清理。
- production 未部署。

## 6. UF-004 / FB-076 旧标签页与直达 URL 补充验收

用户在 FB-075 发布后继续操作部署前已打开的 `/#/exam/exam/users/:examId/...` 标签页，仍可触发传统批量生成器。FB-075只保护测评管理主入口，没有保护旧URL、书签、浏览器历史和首页最近测评入口。

FB-076增加双重分流：传统详情组件先读取测评类型，胜任力在加载传统参与者前使用 `replace` 进入 `CompetencyResults`；首页最近测评也按 `assessmentType` 分流。专项测试由2失败/3通过转为5/5，全量前端为18文件109项。

staging真实E2E结果：

- 截图旧URL自动进入专用结果页；
- 测评管理主“详情”也进入同一专用结果页；
- 排序指标、方向、维度、详情、逐题审计、测试报告、PDF生成、PDF下载、返回共9类操作全部通过；
- 传统专属控件隐藏9/9；
- `/exam/api/exam/exam/generate-report` 调用次数为0；
- PDF下载615336 bytes；
- 部署后相关日志错误为0。

FB-076 staging前端index SHA-256为 `c552833bd2149a3a5ae68f1522e9bee4c2c2b58c4bf26cfa1a80ba6f5fd0a5f1`，备份为 `/opt/talent-assessment/dist.bak.fb076.20260726_210610`。已打开且尚未刷新的旧标签页仍运行浏览器内存中的旧JavaScript，刷新一次后即执行新分流。

## 7. 00401 传统风格结果页最终 staging 验收

22:20 已同时部署最新后端和前端。结果页保留胜任力整体分、评价均值、维度分、开始/完成时间和答题时长，同时对齐传统详情页的筛选、批量工具栏、表格选择和行按钮布局。

- 姓名、电话、完成状态筛选由后端参数化执行，COUNT 与分页数据使用相同条件；
- 查询、重置、整体分/维度分排序及升降序均真实触发分页接口；
- 仅完整答卷可选择；批量生成、批量下载和行下载均调用胜任力报告实例接口；
- “查看”打开 `temp-v1` 报告页；“答题详情”核对 5 条维度和 40 条逐题审计；
- 截图旧 URL 与测评管理主“详情”两个入口均进入专用结果页；
- 传统 `/exam/api/exam/exam/generate-report` 调用次数为 0。

真实输出：

```text
COMPETENCY_RESULT_BUTTONS_E2E_PASS
routes=stale-url|management-detail
buttons=back|query|reset|completion-filter|sort-metric|sort-direction|dimension-selector|answer-detail-dimensions|answer-detail-questions|view-report|batch-generate|batch-download|row-download
legacy_controls_hidden=9/9
download_pdf_bytes=615336
```

部署证据：

- 数据库备份：`/opt/talent-assessment/backups/element_before_00401_results_20260726_221318.sql.gz`，12,530,802 bytes，SHA-256 `866314c71451bf6afac151b07a3bff8e04ef9e5d38a406522989cc33d7e83a34`；
- 应用备份：`/opt/talent-assessment/server.bak.00401_results.20260726_221318`、`/opt/talent-assessment/dist.bak.00401_results.20260726_221318`；
- 后端 SHA-256：`c9adf6df61a12fbb7aab607cfb4727f5f2ff88a866d372f006697e43b507d74d`；
- 本地、远端和公网前端 index SHA-256：`fd7696c5b56302033e4e70fd7264d6697afd2a01ee19ef737be027f09e587950`；
- `talent-assessment`、nginx、MySQL 均 active，内外 health 为 ok，部署窗口关键错误 0；
- 短时 Redis 会话、远端部署/E2E 文件和本地临时包均清理为 0；production 未部署。

## 8. FB-077～087 安全、并发和移动端收口

23:31 起已同时部署累计后端与前端修复，并执行增强后的真实 staging 验收：

- API负例：畸形结果分页JSON返回HTTP 200/code=1/“参数格式错误”；旧query token返回HTTP 401；
- 下载：批量与行下载均验证`application/pdf`，下载615305 bytes；响应`filename*`使用百分号编码且不含`+`空格编码；
- 权限：将短时会话降级为仅`system:user:list`后，结果分页/逐题详情/管理员报告数据依次返回`403|403|403`；
- 并发：同一paper同时2次`force=true`均返回HTTP 200/code=0/completed，总耗时4752ms；最终报告实例1、状态completed、当前PDF存在，同paper目录PDF文件1；
- 内部令牌：部署前真实token query日志基线24行；部署后总计27行中3行为固定`legacy-query-token`主动负例，扣除后仍为24，证明新Chromium内部请求没有把真实token写入query。服务重启日志确认内部token自动生成1次，即完成轮换；
- 移动端：390×844页面无页面级横向溢出，姓名筛选宽度≥240px；详情弹窗为全屏且身份信息至少4个单列行；
- E2E证据纠正：真实检查的传统专属控件为6项，输出已由硬编码9/9改为动态`6/6`。

最终E2E输出：

```text
COMPETENCY_RESULT_BUTTONS_E2E_PASS
buttons=api-negative-guards|back|query|reset|completion-filter|sort-metric|sort-direction|dimension-selector|answer-detail-dimensions|answer-detail-questions|view-report|batch-generate|batch-download|row-download|mobile-layout
legacy_controls_hidden=6/6
download_pdf_bytes=615305
```

部署证据：

- 数据库备份：`/opt/talent-assessment/backups/element_before_fb077_087_20260726_233123.sql.gz`，12,530,655 bytes，SHA-256 `cf81d677926d7c261741ac5ba771219fbc5d5e58bb89e2182265fafa7988b40a`；
- 应用备份：`/opt/talent-assessment/server.bak.fb077_087.20260726_233123`、`/opt/talent-assessment/dist.bak.fb077_087.20260726_233123`；
- 后端 SHA-256：`fcfaa85819702b8f9ab333e1f4ef834fe4bd858464098f740d7dc1cb29247348`；
- 本地/远端/公网前端 index SHA-256：`fa12099adef2e656a9d12a47338a7f810c5ea18ac57ae0dfd7952f7523ee3787`；
- service/nginx/mysql active，内外health正常，部署窗口关键错误0，短时会话和本地/远端临时文件均为0；production未部署。
