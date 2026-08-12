# 00401 一期 90 题 staging 导入与验收记录（2026-08-10）

## 未完成项与范围限制

- production 未连接、未修改、未部署；本次目标仅为 `20.200.136.133` staging。
- 本次只导入并验证一期题本源题，不实现一期固定测评配置、新五档计分、一级聚合、效度结果算法或正式十页报告。
- 导入前的三次备份命令尝试分别因远端重定向权限、变量展开和 PowerShell 引号传递失败；均发生在业务数据写入前。最终独立脚本备份成功并通过 gzip 校验。
- 首次应用切换在后端替换后因误判前端目录为 `ruoyi-ui/dist` 中断；当时数据库尚未导入。确认实际目录为 `/opt/talent-assessment/dist` 后完成前端切换并恢复全部服务。
- 首轮浏览器短时会话因临时 Redis 登录对象缺少完整 `user` 而回到登录页；修正测试会话对象后真实 Chromium 验收通过，全部短时会话已删除。

## 导入基线

- staging 主机：`20.200.136.133 / vm-ubuntu-go-dev`
- 导入前状态：10 个 A/B 维度、0 道胜任力源题、0 个胜任力测评、009 marker=1。
- 导入工作簿：`scripts/data/competency-phase1-import-20260810.xlsx`
- 工作簿 SHA-256：`828c4267e6c7ad387a73ddb0e923b461d5a336ef225bd7414216c0def814de9f`
- 契约：10 列，包含显式“题目类型”。

## 安全备份

完整备份目录：

`/opt/talent-assessment/backups/phase1_90_import_20260810_161855`

包含：

- `element.sql.gz`：数据库、routines、triggers、events；`gzip -t`通过。
- `backend.before`：部署前后端。
- `frontend-dist.before.tar.gz`：部署前前端。
- `SHA256SUMS`：三个备份产物校验清单。
- `traditional-qu.sha256`、`traditional-qu-repo.sha256`、`traditional-qu-answer.sha256`：传统题、题库关系、答案全行有序签名。

导入后重新计算上述三组传统签名，结果全部与备份基线一致：`TRADITIONAL_SIGNATURES=UNCHANGED`。

## 本地质量门禁

- 候选转换器 `--check`：JSON/XLSX 字节级确定性通过。
- 候选身份测试：90题、80维度题、10效度题、62正向维度题、18反向维度题、10正向效度题、90唯一题号全部通过。
- Go 聚焦测试、Go 全量测试、`go vet ./...`、Linux amd64 build通过。
- 前端 Vitest：23个文件、135项测试全部通过。
- 前端 production build通过；仅有既有 asset/entrypoint size warning。
- Python staging验收脚本：AST和编辑器诊断通过。
- Node Chromium验收脚本：`node --check`通过。

导入前额外关闭两个真实缺口：

1. FB-103：维度启用题数原来会把关联效度题计入，导致每维显示9题。先RED后修复为只统计 `competency_question_type=dimension`，GREEN并进入Go全量。
2. FB-104：导入弹窗仍显示“九列模板”。先RED后改为“十列模板”，GREEN并进入前端全量。

## staging 部署

最终部署哈希：

- 后端：`eb0766cd5a98535c876457eb012f7bc1780f9e39bc4dc09ea75f80f4fe2d65a7`
- 前端归档：`1670a327e7cc1cead0b68b36a64eaead59e4b0509f4f1723b3dcded941c514a0`
- 前端 `index.html`：`7c912375ed212515dafeedc4b717c75fcd8ce38ae0a98cc0a98231a5ca639366`

后端、Nginx和MySQL均为active；`http://127.0.0.1:8092/health`返回`{"status":"ok"}`；Nginx配置检查通过。

## 正式API导入证据

通过真实管理员JWT调用正式接口：

1. `POST /exam/api/competency/questions/import-preview`
   - success=90
   - errors=0
   - 响应SHA-256与导入文件一致。
2. `POST /exam/api/competency/questions/import`
   - imported=90
   - 响应SHA-256与导入文件一致。
3. 相同文件重复预览
   - success=0
   - errors=90，全部暴露为既有题号错误。
4. 相同文件重复正式导入
   - 被“导入数据存在错误”拒绝。
   - 数据库源题总数仍为90。

## 数据库与API验收

最终源题统计：

| 指标 | 结果 |
|------|------|
| 总数 / 唯一题号 | 90 / 90 |
| 维度题 / 效度题 | 80 / 10 |
| 维度题正向 / 反向 | 62 / 18 |
| 效度题正向 | 10 |
| 启用题 | 90 |
| A/B维度数 | 10 |
| 每维度维度题 | 8 |
| 每维度关联效度题 | 1 |
| `el_qu_answer`关联 | 0 |
| `el_qu_repo`关联 | 0 |
| 胜任力测评 | 0 |

管理员API验证：

- 维度接口返回10维，启用维度题合计80，不把10道效度题计入维度题数。
- 题目分页返回total=90，题型分布80/10。
- 题目导出返回10列表头+90数据行；按题号规范化后与导入工作簿逐行一致。

## 真实Chromium验收

真实staging静态资源和真实后端API下：

- 页面标题为“00401 胜任力测验题库”。
- 分页总数显示90；首屏20行包含18道维度题和2道效度题。
- 表格存在“题目类型”列，真实显示“维度题”和“效度题”标签。
- 导入对话框显示“专用十列模板”，不再出现“专用九列模板”。
- browser console errors=0，failed requests=0。

## 最终状态与清理

- `talent-assessment/nginx/mysql=active/active/active`。
- 部署窗口应用journal的`panic|level=ERROR|fatal`计数为0。
- `login_tokens:phase1-*`短时会话残留为0。
- 上传到远端`/tmp`的后端、前端、工作簿、验收脚本和解压目录均已清理。
- production未修改。

结论：00401一期90题已在staging正式导入；导入内容、题型隔离、方向、启用状态、维度分布、关系隔离、重复导入保护、导出回读和真实UI均通过验收。
