# 项目当前状态（后续开发与测试的唯一现状入口）

> 更新时间：2026-10-10
> 本文件只保存**当前有效结论**；过程、失败和纠正历史保留在 [project-memory.md](project-memory.md)。发生冲突时，按“本文件 → project-memory 顶部最新纠正 → 日期更晚的限定事实 → 历史记录”读取。

## 1. 环境边界

| 环境 | 当前结论 |
|---|---|
| local | UF-057最终Linux候选SHA=`c321bf35af790a855a342995f39427f72a874dcc494fcd984c5142760f2cec87`；相关报告测试53项、Python控制器语法和build通过。仍有范围外未提交改动，不以整个工作树代表已部署字节 |
| staging | `20.200.136.133 / vm-ubuntu-go-dev`；当前后端SHA=`ec4c85d73fd2e1e8acb719b137b0b59309bc9de2f8a8e3f80653b7f00e5ad5`。真实90答/10页PDF已验证时长隐藏、无Page前缀、页码居中、计划执行粗体、cleanup0；随后Azure SSH不可达，未部署仅为LibreOffice 7.4增加的wrapper兼容增量 |
| production | 🟢 后端/进程SHA=`c321bf35af790a855a342995f39427f72a874dcc494fcd984c5142760f2cec87`（保留FB-219并含UF-057），前端index=`593d4a20d890d73bf47f9a519a22fffbdb4b539afd5e1cc1dcc2fdebff73cfde`；指定00401历史报告已覆盖并通过10页PDF、居中数字页码、时长隐藏、计划执行粗体及绑定下载验收 |

[生产发布 - 2026-10-09] `39.106.61.48 / iZ0yosjdcen2p4Z`已部署封存release；完整备份=`/opt/talent-assessment/backups/production_release_backup_20261009_3e3692ce14634021`。真实管理员验收发现并修复FB-219，补丁提交=`f6d0719`、binary备份=`/opt/talent-assessment/backups/fb219_binary_20261009_7d5fbd1559914437`。最终MySQL 5.7主库有14张`el_mng_*`表、005 repo/exam/profile/snapshot/run/revision/reissue各2、reissue audit12；两个exam均恢复`state=1`。共享DOCX SHA=`05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c`，内容XLSX SHA=`b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c`。真实管理员状态往返、模板、结果和647985-byte reissue PDF读取均通过。

[生产E2E补充 - 2026-10-09] 00501真实创建唯一候选、140答、提交、13维/4模块、reissue PDF view/download同SHA，baseline逐表恢复。用户随后由外部管理员会话将00501/00502均改为`state=0`进行中，属于客户控制状态。00401经用户明确批准导入staging已验收工作簿SHA=`828c4267e6c7ad387a73ddb0e923b461d5a336ef225bd7414216c0def814de9f`：新增A/B 10维+90题；旧48维主数据order归档至101–148，8个重名追加“（历史）”，旧冻结测评/结果字节不变；完整备份=`/opt/talent-assessment/backups/phase1_content_20261009_8b17bb3f57794f2c`。随后真实00401完成2组/10维/90题发布、90答、提交幂等、v1/v2结果、筛选、三Sheet导出并exact cleanup。用户随后授权沿用staging具名批准并同步正式报告：production已受控安装124条v2文案、approved production v2包及精确v2模板SHA=`f98599939e3bf7923abf8bd457e30259cde3fa68d6c4e9dc9d3288de0a2eae16`，备份=`/opt/talent-assessment/backups/phase1_report_20261009_8af8c5dd2963418f`；但新建测评仍绑定v1版本元组，完整报告E2E因此保持RED且已精确清理。v1精确模板SHA=`54b167fcc02737ba44e0432a930e1f5657db57cec87d9063f1ee7c111c4932bf`和66条确定性内容已在本地重建；两轮production写前preflight均在备份/写入前失败，最终确认服务PID=`2355899`、`NRestarts=0`、v1模板仍不存在、payload/probe均清零。当前不得将00401正式报告称为production GREEN。

[全量staging替换回滚 - 2026-10-09] staging SSH恢复并完成实时差异盘点；题库10维/90题及内容SHA已与production一致，Schema仅有MySQL版本元数据、staging迁移标记表及production额外安全索引差异。受控替换创建全量备份=`/opt/talent-assessment/backups/phase1_staging_sync_20261009211916`并通过文件/DB结构验收，但真实00401创建测评被staging旧后端的管理特质MySQL5.7保护校验拒绝。已立即完整回滚并验证：PID=`2366975`、NRestarts=0、后端/进程=`f850575b...`、前端=`abf93dd1...`、健康200、v1仍0/0/absent、v2保持124/1及原模板。当前状态矩阵和P0正式报告RED结论不变。

[00401正式报告GREEN - 2026-10-09] 用户选择staging基线仅保留FB-219兼容补丁后重新受控发布，完整备份=`/opt/talent-assessment/backups/phase1_staging_sync_20261009212420`。最终server/process=`f850575b...`、staging前端index=`593d4a20...`/393文件、v1/v2文案=`66/124`、approved包=`1/1`、模板SHA=`54b167fc.../a814c36e...`，公开renderer开关已通过独立systemd drop-in启用。真实production 00401完整链含10页正式PDF、审计2、三Sheet导出及exact cleanup全部PASS；PID=`2367815`、NRestarts=0、最近fatal/panic/permission=0。

[00401旧测试历史清理 - 2026-10-09] 完整备份=`/opt/talent-assessment/backups/phase1_history_delete_v2_20261009220803`；已删除旧9测评/22试卷/19结果/5报告/454题/48维完整测试闭包。production当前仅保留10维90题，分布和dimension/question SHA与staging实时收据完全一致；传统858题、正式文案/包/模板保持。PID=`2370974`、NRestarts=0、health200、MySQL临时写PASS。

[staging基准只读复核 - 2026-10-09] production前端及393文件、全部9份模板、00401 10维90题/正式内容均与staging精确一致；后端仅保留已批准FB-219差异。00401字段级Schema差异只有MySQL8/5.7元数据，staging多迁移标记表，production多安全索引，FK一致。扩展到全题库后发现传统题总数staging/production=`1135/858`：六个共享repo题数及语义SHA一致，但production缺staging `00301`那套48题及`00302`身份（production `00301`内容实际匹配staging `00302`），另有production专有`10201`一题；四个repo元数据及005 definition bundle语义也不同。故00401范围GREEN，但全题库覆盖同步未获准且当前NO-GO。

[UF-055 production静态站点恢复 - 2026-10-09] 公网root/favicon曾500而API health200；根因是旧失败rollback把`/opt`与应用父目录mode保留为0700，Nginx worker无法穿越。用户批准仅将两目录修复为0755，无文件/DB/配置/程序改动且无重启。最终公网root/favicon/API=`200/200/200`，真实浏览器登录页加载；后端PID=`2370974`、NRestarts=0不变。

[production部署只读审计 - 2026-10-09] 当前runtime/Nginx/MySQL/Redis/匿名浏览器均GREEN，当前PID启动后应用错误0；但安全与运维NO-GO项仍在：HTTP登录无TLS、含秘密配置0644、Go以root运行且systemd评分9.6 UNSAFE、root SSH、UFW规则过宽、409条孤儿paper、磁盘80%及备份/上传增长。完整证据见[production-deployment-audit-20261009.md](production-deployment-audit-20261009.md)。本轮零写，未改现场。

[00401 production保留浏览器链 - 2026-10-09] marker=`TEST-BROWSER-00401-20261009223722`的真实候选/试卷/90答/v1-v2结果/正式报告/PDF按用户要求保留；刷新恢复、提交确认、管理员列表与90题详情、浏览器PDF 200均通过。PDF为10页/717300 bytes/SHA=`145bc06b...`，审计7条；短期管理员会话、远端临时脚本、本地token及SSH均已清理。该保留PDF生成于UF-056模板发布前，仍显示`时长：1分钟`；它是历史文件，不代表当前活动模板状态。

[UF-056模板发布 - 2026-10-09] staging先安装候选并以真实90答、LibreOffice转换、10页PDF验证首页时长隐藏，cleanup0且业务基线不变。用户批准production仅替换模板且不涉及数据库后，production活动v2模板从`a814c36e...`原子更新为611455-byte SHA=`52e0020c...`，只变`word/document.xml`；备份=`/opt/talent-assessment/backups/uf056-template-20261009232349`。服务PID=`2370974`、NRestarts0、内外root/API 200、关键日志0，零数据库访问/写入。保留旧PDF未重生成，仍含历史时长文本；新生成或明确重生成的v2报告使用新模板。

[UF-057历史报告覆盖 - 2026-10-10] 用户批准直接覆盖保留paper=`9174f98e-181f-486e-bbc7-0118bc39ced1`的旧PDF。根因是LibreOffice 7.4把嵌套内容控件中的PAGE字段当文本且忽略动态标签run粗体；renderer在渲染后仅移除冗余wrapper、统一PAGE字段并居中内层段落。production同机确定性DOCX先验证10页、封面无页码、正文数字1–9居中、Page字面量0及`计划执行：`粗体，再发布后端SHA=`c321bf35...`，binary备份=`/opt/talent-assessment/backups/uf057-production-lo74-server-20261010001507`。指定report=`103d9a0d-0113-4aef-8962-434645783477`从SHA=`7cd40b29...`/716626 bytes覆盖为`cf1385c5...`/758202 bytes；精确备份=`/opt/talent-assessment/backups/uf057-report-20261010001515`，report/current保持1/1，audit最终15，认证下载与绑定一致。PID=`2375402`、NRestarts0、内外HTTP 200、关键日志0、active paper0、临时payload已清理。

staging最后独立终验：PID `22668`、`NRestarts=0`，talent-assessment/nginx/mysql均active，内外health均200，active paper=0，应用fatal/panic/permission错误0，Nginx 5xx=0。

## 2. 产品状态矩阵

| 产品/链路 | 当前状态 | 已验证范围 | 尚未覆盖/限制 |
|---|---|---|---|
| 00401一期胜任力 | 🟢 STAGING GREEN / 🟢 PRODUCTION GREEN | production已安装10维90题、v1/v2正式文案/包/模板；真实浏览器完成登记、刷新恢复、90答、提交、管理员结果/详情、10页正式PDF及浏览器下载响应；UF-056模板经staging真实动态报告验收后发布production | 保留旧PDF未重生成，仍是历史版式；真正浏览器native download事件及并发/容量未重跑 |
| 003 MBTI | 🟢 STAGING GREEN（API+报告） | 48题保存、ESTJ计分回读、完整版/简版PDF、16+16模板、匿名模板门禁、exact cleanup | 本轮未做参与者浏览器逐步交互和16种人格逐类型报告全矩阵 |
| 00501/00502管理特质 | 🟢 STAGING GREEN / 🟢 PRODUCTION完整TEST链GREEN | production新建唯一候选、140答、提交、13维/4模块、PDF view/download及exact cleanup通过；客户当前将两测评设为进行中 | formal仍不在本次范围 |
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

1. **管理特质正式报告/formal assembly**：production当前只发布带TEST标注的005与reissue基线；formal registry/approval未包含，不得把TEST报告称正式报告。
2. **生产管理员凭据轮换**：当前admin仍匹配仓库历史默认候选；需用户明确授权后单独轮换并验证登录/回滚。
3. **生产TLS/HTTPS**：当前登录页仅HTTP、无443且浏览器非secure context；须配置受信证书、HTTPS及HTTP重定向后重验登录/下载/Cookie。
4. **秘密配置与root运行**：application配置含秘密标记但为0644；有效service以root运行且hardening评分9.6 UNSAFE。先在staging设计0600秘密注入、专用用户写目录与systemd沙箱，再单次生产迁移。

### P1 — 可在后续独立测试中关闭

1. **005全新staging生命周期**：新建唯一标记draft→人员→freeze→140答→提交→结果→报告→exact cleanup；当前主要依赖保留基线验证。
2. **真实浏览器用户链**：00401参与者入口、刷新恢复、90答、提交、管理员结果/详情和浏览器PDF响应已关闭；仍缺真正浏览器native download事件。MBTI仍需补参与者浏览器链；005历史native事件盲区也应统一复测。
3. **MBTI类型矩阵**：至少为16种类型各验证模板可加载和DOCX/PDF转换；当前完整链只得到ESTJ，模板只验证存在性16+16。
4. **001/002兼容回归**：若下一发行会改公共paper/exam/report代码，应重跑临时001/002/003新卷完整链，而非只读Detail。
5. **最新FB-218真实失败注入复验**：同步和异步失败关闭补丁已有RED/GREEN、Go全量并随production后端SHA上线；production未故意触发LibreOffice故障，仍缺远端真实失败注入证据。
6. **工作树清理**：生产字节已由SHA封存并关联已推送提交，但本地仍有本次范围外的未提交删除/编辑器改动；不得误提交或据此判断生产漂移。
7. **传统题库身份差异决策**：先确认00301/00302在production的产品身份和是否需要双库并存，再逐字段比较00101/00102/00501/00502 repo元数据及3条005 definition bundle；未完成引用闭包和客户决策前禁止以staging全量覆盖production。
8. **生产网络面收紧**：UFW对全网允许3306/10301/8088/9001/39000–40000且root SSH开启；云侧当前拦截多数端口，但应核实所有者后同步收紧UFW/NSG，并让MySQL和Go 8092只绑定内网或localhost。
9. **409条孤儿试卷闭包**：state0=199/state2=210，关联paper_qu20566、answer42963；先核产品、时间、报告/用户引用和恢复价值，未经批准不得删除。
10. **磁盘与保留策略**：根盘80%/余7.8GB，backups38项1.7GB，uploadPath502MB/782文件；先建立备份索引和DB文件引用图，再逐项批准清理。

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
