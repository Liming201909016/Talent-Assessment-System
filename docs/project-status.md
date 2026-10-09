# 项目当前状态（后续开发与测试的唯一现状入口）

> 更新时间：2026-10-09  
> 本文件只保存**当前有效结论**；过程、失败和纠正历史保留在 [project-memory.md](project-memory.md)。发生冲突时，按“本文件 → project-memory 顶部最新纠正 → 日期更晚的限定事实 → 历史记录”读取。

## 1. 环境边界

| 环境 | 当前结论 |
|---|---|
| local | Go全量、Windows/Linux backend build、前端35文件611项及production build通过；Linux候选SHA=`753fad7a6134139b11ed3285c418da092c160b4fe81b9baf53dbf70f6d3cf0fc`，前端index=`abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4`。仍有大量未提交改动，不能用Git提交号代表已部署字节 |
| staging | `20.200.136.133 / vm-ubuntu-go-dev`；当前后端SHA `f2940fc5ea51edffc4f325df1f461f3ba4e86868df3aa0594af95764e880d61e`，前端index SHA `593d4a20d890d73bf47f9a519a22fffbdb4b539afd5e1cc1dcc2fdebff73cfde` |
| production | 🟡 2026-10-09受控发布完成且服务健康；真实管理员验收发现FB-219令005运行时Schema门禁在MySQL5.7上误拒绝，启用操作409且state保持1。本地补丁GREEN，production待发布复验 |

[生产发布 - 2026-10-09] `39.106.61.48 / iZ0yosjdcen2p4Z`已部署Git HEAD `227bff9`对应封存产物。发布前创建数据库、应用树和系统配置完整备份`/opt/talent-assessment/backups/production_release_backup_20261009_3e3692ce14634021`并校验。最终MySQL 5.7主库有14张`el_mng_*`表、005 repo/exam/profile/snapshot/run/revision/reissue各2、reissue audit12；两个exam均`state=1`可见禁用，00501 question v1、00502当前question v2。双环境变量均为`production`；共享DOCX SHA=`05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c`，内容XLSX SHA=`b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c`。匿名直连管理特质详情401，最近10分钟应用error journal为空。未执行认证管理员浏览器操作或客户启用切换。

staging最后独立终验：PID `22668`、`NRestarts=0`，talent-assessment/nginx/mysql均active，内外health均200，active paper=0，应用fatal/panic/permission错误0，Nginx 5xx=0。

## 2. 产品状态矩阵

| 产品/链路 | 当前状态 | 已验证范围 | 尚未覆盖/限制 |
|---|---|---|---|
| 00401一期胜任力 | 🟢 STAGING GREEN（范围限定） | 2组、10维、90题、恢复、90答、提交幂等、结果、筛选、三Sheet导出、批准v2 DTO、A4 10页PDF、审计及exact cleanup | production内容/发布仍需独立批准；本轮未做真实参与者浏览器全链和并发/容量重跑 |
| 003 MBTI | 🟢 STAGING GREEN（API+报告） | 48题保存、ESTJ计分回读、完整版/简版PDF、16+16模板、匿名模板门禁、exact cleanup | 本轮未做参与者浏览器逐步交互和16种人格逐类型报告全矩阵 |
| 00501/00502管理特质 | 🟢 STAGING GREEN / 🟢 PRODUCTION已发布（范围限定） | 名称与V67/V96、各140/700、保留结果50/13维/4模块、共用模板管理、TEST报告；production程序、14表、2 repo、2份禁用测评、2 run、4 PDF及模板已验收 | production未执行认证管理员浏览器/客户启用切换；formal仍不在本次范围 |
| 00101传统测评 | 🟢 STAGING只读冒烟 | Detail及有效XLSX导出 | 未做本轮临时新卷完整交卷回归 |
| 00201/00202历史管理特质 | 🟡 兼容保护有效 | Detail可读；冻结历史和002源数据保持 | 通用原始导出被管理特质legacy guard按设计403；是否需要专属导出属于产品决策，不应放宽guard绕过 |

## 3. 当前有效技术决策

1. 00501=`管理特质测验基层员工新版`，00502=`管理特质测验干部新版`；只对00502覆盖V67/V96，002与冻结历史不改。
2. 00501/00502共用一份DOCX；当前模板SHA=`05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c`。
3. 管理特质TEST运行时只在`REPORT_EFFECTIVE_ENV`与`MNG_TEST_REPORT_ENV`归一化后同为`local`、`staging`或`production`时启用；空值、不一致、`prod`及畸形值关闭。production必须显式使用`production/production`，不新增开关；005导入后以现有`el_exam.state=1`可见但禁用，由客户通过现有状态操作决定启用。
4. MBTI报告转换统一复用`pkg/libreofficepdf`，有90秒context、隔离workspace和失败关闭；禁止把DOCX冒充PDF。
5. 00401/MBTI/005测试必须使用唯一标记、写前基线、finally精确清理及基线前后比较；不可再硬编码“staging必须空库”。
6. 正式备份、客户材料、业务PDF、审计证据和`scripts/test/results`不是垃圾文件；清理前必须先做所有权、引用、大小、年龄和回滚价值检查。

## 4. 已关闭的矛盾

| 历史矛盾 | 当前结论 |
|---|---|
| 005题库/模板仍为LOCAL GREEN | 已被2026-10-09真实staging发布、浏览器/API和数据核验覆盖，现为STAGING GREEN（范围限定） |
| 当前工作树management-traits发布为PARTIAL | 该条是发布当时阶段结论；后续真实认证浏览器、模板、报告及完整主链验收已补齐，当前按本文件矩阵读取 |
| 00401 staging必须为空才能跑E2E | 错误；当前脚本接受非空基线并要求清理后精确恢复 |
| MBTI直接执行LibreOffice且失败可返回DOCX | 已由FB-218修正为共享有界转换、失败关闭 |
| “完整staging GREEN”等于全系统/production GREEN | 错误；production后来经独立评估、恢复副本演练、完整备份及受控发布单独批准并验收，不能由staging结论自动推出 |

## 5. 仍需处理的遗漏事项

### P0 — 外部批准/上线门禁

1. **00401 production正式内容批准**：本次只部署已批准的additive结构，没有导入00401正式内容；仍需客户/负责人确认精确内容SHA、模板SHA和适用环境。
2. **管理特质正式报告/formal assembly**：production当前只发布带TEST标注的005与reissue基线；formal registry/approval未包含，不得把TEST报告称正式报告。
3. **005 production认证验收**：程序、数据库、资产、匿名门禁和健康检查已通过；仍需客户使用真实管理员账号确认列表显示、模板元数据/下载、报告view/download以及一次`state 1→0→1`受控切换。

### P1 — 可在后续独立测试中关闭

1. **005全新staging生命周期**：新建唯一标记draft→人员→freeze→140答→提交→结果→报告→exact cleanup；当前主要依赖保留基线验证。
2. **真实浏览器用户链**：00401和MBTI本轮以真实API/DB/PDF为主；需补真实参与者浏览器入口、刷新恢复、提交、原生下载事件。005的历史native事件盲区也应统一复测。
3. **MBTI类型矩阵**：至少为16种类型各验证模板可加载和DOCX/PDF转换；当前完整链只得到ESTJ，模板只验证存在性16+16。
4. **001/002兼容回归**：若下一发行会改公共paper/exam/report代码，应重跑临时001/002/003新卷完整链，而非只读Detail。
5. **最新FB-218真实失败注入复验**：同步和异步失败关闭补丁已有RED/GREEN、Go全量并随production后端SHA上线；production未故意触发LibreOffice故障，仍缺远端真实失败注入证据。
6. **工作树清理**：生产字节已由SHA封存并关联已推送提交，但本地仍有本次范围外的未提交删除/编辑器改动；不得误提交或据此判断生产漂移。

### P2 — 测试与治理改进

1. 本轮未重算Go/Vue覆盖率；应建立新的coverage快照并做漂移比较。
2. `project-memory.md`为697,706 bytes、2,254行，继续作为时间线；后续先读本文件，只在需要历史证据时查时间线，避免旧`PARTIAL/BLOCKED`误当当前状态。
3. staging `/opt/talent-assessment/tmp/uploadPath`含历史业务/报告目录，本轮未删除；需另做文件与数据库引用图后才能提出清理清单。
4. `/opt/talent-assessment/backups`历史备份不在本轮清理授权内；如需释放空间，应先输出年龄/大小/关联发行/恢复演练价值，再由用户逐项批准。
5. `scripts/test/results`含发布和审计证据，不应按普通测试垃圾整体清空；可另制定保留周期和索引。

## 6. 后续任务读取规则

1. 先读本文件确定当前状态和待办。
2. 再读 [project-memory.md](project-memory.md) 顶部最新记录；只有调查历史根因时才读取对应日期段。
3. 功能分支以 [business-branches.md](business-branches.md) 为准；修复回归以 [regression-tests.md](regression-tests.md) 为准；业务链以 [business-chains.md](business-chains.md) 为准。
4. 任一新验证必须写明：环境、源码/产物SHA、是否真实DB/浏览器、创建数据、cleanup结果、未覆盖项。
5. 旧记录中的`PASS/GREEN/已部署`只对其原始日期和明确范围有效，不自动继承到当前版本。
