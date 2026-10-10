# 阅读入口与冲突裁决（2026-10-09）

- **当前状态唯一入口**：[project-status.md](project-status.md)。后续开发/测试先读该文件；本文件保留完整时间线、失败证据与纠正记录，不再承担“一眼判断当前状态”的职责。
- 冲突优先级：`project-status.md` → 本文件顶部最新纠正 → 日期更晚且范围更具体的事实 → 较早历史记录。
- 历史标题中的`PASS/GREEN/已部署/已完成`只对其明确环境、版本、功能和日期有效；不得扩大为全系统或production结论。
- 历史`PARTIAL/BLOCKED/未部署`若已被后续同范围真实证据关闭，保留作过程证据，但不再是当前状态。禁止删除失败历史或覆写原收据。
- 当前待办已按P0/P1/P2整理在[遗漏事项](project-status.md#5-仍需处理的遗漏事项)；不要从下方两千余行历史中重新猜当前待办。

# 2026-10-10 UF-057 production旧PDF覆盖完成（LibreOffice 7.4/24.2兼容）

- 用户要求提高效率，不再重复90题，复用staging真实生命周期收据并直接覆盖历史PDF。staging后端SHA=`ec4c85d7...`真实90答/10页PDF已PASS：时长隐藏、Page前缀0、数字页码居中、`计划执行：`粗体、audit2、cleanup=`0|0|0|0`、基线`10|24|48|24`前后一致。随后仅为production LibreOffice 7.4增加wrapper兼容时，Azure staging SSH连续两次超时，未上传或改动。
- 两环境模板SHA=`52e0020c...`及微软雅黑/宋体文件一致；LibreOffice为staging 24.2、production 7.4。坐标证据定位：模板`footer2`页码段落缺居中导致物理第2页数字1靠左；嵌套DrawingML/VML内容控件使7.4把后续PAGE字段显示为文字并忽略动态label粗体。RED后改为从PAGE指令定位最近内层段落、渲染后移除页脚及selected-item冗余wrapper；相关报告测试53项通过。
- production确定性DOCX（零DB）实证10页：封面无页码，物理2–10页数字1–9中心约297.8/298.8（页面中心297.7），PAGE字面量0，`计划执行：`PDF XML为粗体。首次用中间binary重生成在第3编号页校验失败并`ROLLBACK=completed`；保留失败证据。最终Linux SHA=`c321bf35...`发布，binary备份=`/opt/talent-assessment/backups/uf057-production-lo74-server-20261010001507`，PID2375402/NRestarts0。
- 当前绑定基线不是旧收据`b9493c1c...`，DSN只读preflight确认覆盖前为SHA=`7cd40b29...`/716626 bytes、report/current=1/1、audit13、active paper0。最终仅force重生成授权report=`103d9a0d-0113-4aef-8962-434645783477`，精确备份=`/opt/talent-assessment/backups/uf057-report-20261010001515`；新PDF SHA=`cf1385c5...`/758202 bytes、10页、时长隐藏、Page前缀0、页码居中、计划执行粗体、认证下载/DB/文件一致。generate后audit14，最终下载验收后audit15；内外HTTP200、关键日志0、payload清理PASS。

# 2026-10-09 UF-056 staging动态GREEN并发布production（仅模板，零数据库访问）

- 重构失败的命令传输方式后，staging零写preflight确认旧v2 SHA=`a814c36e...`、服务active/NRestarts0及LibreOffice工具完整；DOCX解包只变`word/document.xml`。备份=`/opt/talent-assessment/backups/uf056-template-20261009230941`，安装SHA=`52e0020c...`后真实完成2组/10维/90题/90答/v1-v2结果/报告生成/LibreOffice转换/10页PDF下载；首页文本无`时长：`或`分钟`，审计2，cleanup=`0|0|0|0`，业务基线`10|24|48|24`前后一致。
- 用户明确要求发布production、替换模板且本次不涉及数据库。production严格设置`databaseAccess=false/databaseWrites=false`，旧活动模板SHA=`a814c36e...`；上传候选后再次证明ZIP有效、隐藏字段合同保留、语义差异仅`word/document.xml`，建立0700备份=`/opt/talent-assessment/backups/uf056-template-20261009232349`并原子替换为611455-byte SHA=`52e0020c...`。
- 终验：service active、PID=`2370974`、NRestarts=`0`、服务器内及公网root/API=`200/200`，公网root 16155 bytes、health=`{"status":"ok"}`，近5分钟panic/fatal/permission=0；SSH/SFTP会话关闭。没有重启、程序/前端/字体/Schema/题库/Nginx/systemd改动，也没有数据库读取或写入。保留的历史PDF未重生成，仍保留旧版式；后续新生成或明确重生成的v2报告使用新模板。

# 2026-10-09 production 00401真实浏览器保留数据链与UF-056本地候选

- 用户要求模拟真实production场景并保留测试业务数据。唯一marker=`TEST-BROWSER-00401-20261009223722`、exam=`1791556668267388349`、candidate=`1791556773728566030/TEST223722`、paper=`9174f98e-181f-486e-bbc7-0118bc39ced1`。真实浏览器完成登记、90题作答、前5题后刷新恢复、90/90二次确认提交、管理员结果列表、10维/90题详情、报告生成及PDF下载。
- 最终只读闭包为candidate/paper/题/已答/维/组/效度/总体=`1/1/90/90/10/2/1/1`，v2=`1/1/3/10/1`，completed报告1；报告ID=`103d9a0d-0113-4aef-8962-434645783477`，717300 bytes，SHA=`145bc06b...`，10页。报告审计因真实重复生成/查看/下载增长到7条，属于保留数据；不得再用固定`audit=2`断言。短期管理员Redis会话删除后remaining=0，远端临时脚本、本地token文件已删除，SSH已关闭；业务闭包完整保留。
- UF-056定位：production与staging活动v2模板同SHA=`a814c36e...`；production PDF确实显示`时长：1分钟`，而页码居中、综合表现模板字重及微软雅黑/宋体嵌入均正确。先RED后生成本地候选SHA=`52e0020c...`，仅隐藏可见时长并保留隐藏字段合同；专项、v2 Word、handler及build GREEN。三轮staging动态转换尝试均在转换前因命令传输/清理顺序失败，按门禁停止；未替换任何远端模板，production仍存在该版式问题。

# 2026-10-09 production部署完整只读审计（可用但安全/运维需加固）

- 当前运行GREEN：server/process=`f850575b...`、index=`593d4a20...`/393文件，Go active/PID2370974/NRestarts0，Nginx配置PASS，root/favicon/API/8092均200，MySQL alive、Redis PONG；真实匿名浏览器登录页无console error。当前PID自22:09:12启动后ERROR/FATAL/panic/permission均0；旧5899应用ERROR和Nginx permission记录不代表当前持续故障。
- P0安全事实：登录仍是HTTP且无443/secure context；5个configs顶层文件全部0644，两个application配置检测到3/4个秘密字段标记；effective systemd以root运行且所有主要hardening关闭，`systemd-analyze security=9.6 UNSAFE`，并与仓库unit的liming用户定义漂移；管理员仍匹配历史默认候选的既有P0未关闭。
- P1网络/数据/容量：UFW对Anywhere开放22/80/8090以及3306/10301/8088/9001/39000–40000，云侧当前只让22/80/8090可达，但MySQL/Go仍绑定所有网卡；SSH允许root直登。数据库有409条无父exam的paper（state0=199/state2=210），关联paper_qu20566、answer42963，candidate/MBTI直接引用0。根盘80%/余7.8GB，backup38项1.7GB，uploadPath502MB/782文件。
- P2：根响应缺CSP/XCTO/XFO/Referrer/Permissions安全头；敏感URL由SPA fallback返回200/16155而非实际泄漏；backups根755且内部81个0644文件依赖子目录700保护，upload 782文件同样依赖父目录700；CUPS/BT Panel监听所有网卡但云侧不可达。
- 本轮全程只读，未改代码、配置、权限、DB、防火墙或服务。完整脱敏证据与修复顺序见`docs/production-deployment-audit-20261009.md`。

# 2026-10-09 staging基准对production程序/模板/Schema/题库只读复核

- 两环境同一脚本实时采集且全程零写，production每次SSH均在收据返回后关闭。服务、进程文件一致性和内外health均PASS；前端index=`593d4a20...`且均393文件。后端按已批准兼容边界保持staging=`f2940fc5...`、production=`f850575b...`（staging基线+FB-219），不是未解释漂移。
- 模板目录各9文件；00101/00102/00201/00202/00301 XLSX、00401 v1/v2 DOCX、005内容XLSX和共用DOCX的文件名、大小、SHA逐项完全一致。00401文案66/124、包1/1及四项内容/包SHA也完全一致。
- 00401 Schema外键SHA相同。字段级差异仅MySQL8移除整数display width、`DEFAULT_GENERATED`元数据；staging另有`el_competency_migration`标记表。production仅多`el_qu.idx_qu_type_level_id(qu_type,level,id)`安全索引。未发现业务列、默认值、顺序或外键缺失。
- 00401题库精确一致：10维、90题、分布`90|90|80|10|62|18|10|90`、dimension SHA=`ac6d4290...`、question SHA=`db1554e7...`。production旧00401运行历史为0；staging保留10测评/24试卷/24结果/28报告，这是环境业务历史差异，不应覆盖production。
- 全题库扩展核验发现不能声称全库一致：共享00101/00102/00201/00202/00501/00502题数及题干+选项语义SHA一致，但00101/00102/00501/00502 repo元数据SHA不同；production `00301`的48题语义SHA等于staging `00302`，而staging自身`00301`是另一套48题，production缺`00302`；production另有`10201`一题。传统题总数staging/production=`1135/858`，未关联题=`122/0`。005 definition bundle均3行但语义SHA不同，需单独列字段差异和历史引用闭包后才能决定同步。
- 结论：00401及全部9份部署模板GREEN；全程序允许FB-219差异；全Schema无业务缺口；“全部传统/005题库按staging覆盖production”当前NO-GO，必须先确认00301/00302产品身份、repo元数据和bundle差异，不能直接全量复制。

# 2026-10-09 UF-055 production静态站点500已恢复

- 用户五问确认production根地址刚刚开始全站500、无特定数据条件。公网探针证明root/favicon500，但API proxy health200；后端active、8092=200、PID=`2370974`、NRestarts=0，故不是Go服务或数据库故障。
- `namei`实证dist/index自身755/644，但`/opt`和`/opt/talent-assessment`均0700。根因追到已禁止复用的旧清理controller：备份父目录先被递归chmod700，失败rollback再`cp -a ... /`保留mode；此前“根目录标准0755”只验证并修复了`/`，遗漏`/opt`和应用目录。[纠正 - 2026-10-09] 上述旧结论并不代表完整静态路径权限已恢复。
- 用户批准后只执行两个`chmod 755`，断言失败自动恢复原mode；worker `www`读index、服务器内三HTTP和公网三HTTP全部PASS。公网root=200/16155、favicon=200/26900、API health=200/15；真实浏览器显示系统登录页。未重启、未改代码/配置/文件内容/数据库。

# 2026-10-09 00401旧测试闭包已删除（production与staging题库精确一致）

- 用户再次明确“直接删除”。弃用失败controller后新增主键常量版`scripts/tools/production-phase1-history-delete-v2-20261009.sh`：DELETE内零子查询、保留SQL_SAFE_UPDATES、不开FOREIGN_KEY_CHECKS、不TRUNCATE；production只输入一次root密码并复用同一SSH/SFTP会话完成全流程。
- 删除前新建并验证完整数据库及5份报告文件备份=`/opt/talent-assessment/backups/phase1_history_delete_v2_20261009220803`。事务按外键顺序删除旧9测评、22试卷、19结果、147维度结果、5报告、11审计、1406试卷题、548冻结题、旧454题、旧48维及3条owned导出日志；报告文件同步删除。其他已核为0的闭包表无写入。
- 最终production只读收据：旧competency exam/paper/result/report=`0/0/0/0`，active paper0；当前10维/90题、分布`90|90|80|10|62|18|10|90`，dimension SHA=`ac6d4290...`、question SHA=`db1554e7...`；与当前staging同一时点六项完全一致。传统001/002/003/005题共858保留，v1/v2文案66/124、包1/1、两模板SHA不变。
- 服务active、PID=`2370974`、NRestarts=0、server/process=`f850575b...`、health200；根目录标准0755、mysql临时写PASS。production认证会话已关闭。UF-054关闭。

# 2026-10-09 00401旧454题清理BLOCKED并完整恢复（禁止复用失败controller）

- 用户明确授权：完整数据库备份后删除旧00401的454题及全部相关测试历史，同时保留当前10维90题和001/002/003/005传统题。只读闭包为旧48维、454题、9测评、22试卷、1406试卷题、548冻结题、19结果、147维度结果、5报告、11审计和3条导出日志；active paper=0，其他交叉引用均0。
- 多轮controller均先创建并校验全量gzip数据库备份及5份报告文件副本；已知有效备份包括`phase1_history_delete_20261009214800`、`...214949`、`...215404`。删除事务分别被MySQL safe-update及服务器临时目录权限拒绝；未得到一次完整DELETE COMMIT。
- 调查确认production根目录`/`曾为0700，导致mysql用户无法穿越到`/tmp`。用户批准修复为标准0755后，mysql临时写PASS，并从`...215404`精确恢复一度被不完整全库restore清空的5条`el_competency_report`；11条audit始终保留。
- 回滚逻辑中的`cp -a report-files/. /`又把根目录模式覆盖回0700；最终已再次修复`/`为0755，mysql临时写PASS，服务active/health200。最终只读收据：server/process=`f850575b...`、active paper0、当前10维90题及SHA不变、旧9 exam/22 paper/19 result/5 report全部恢复、v1/v2文案66/124和包1/1不变。
- 按“三轮失败停止”门禁，本轮停止。`scripts/tools/production-phase1-history-delete-20261009.sh`不得再次执行；下一步必须重写为逐表预计算主键直接DELETE的新controller，并先在恢复副本演练，不得继续现场试错。

# 2026-10-09 00401 production全量staging基线同步完成（仅保留FB-219兼容补丁，GREEN）

- 用户在精确staging后端因MySQL5.7校验失败并完整回滚后，明确选择“staging程序 + FB-219补丁”。最终候选因此锁定：后端继续使用已验证FB-219 SHA=`f850575b...`；前端使用staging精确index=`593d4a20...`/393文件；00401 v1/v2模板使用staging精确SHA=`54b167fc.../a814c36e...`；v1文案66条和批准包1条来自staging只读导出，题库与Schema无需写入。
- 第二次受控发布创建新全量备份=`/opt/talent-assessment/backups/phase1_staging_sync_20261009212420`，controller结构验收PASS。随后真实E2E首次到报告生成阶段发现公开配置`PHASE1_WORD_REPORT_ENABLED`未启用；本轮候选、卷、结果、报告、审计、文件及Redis均精确清理，baseline完全恢复。
- 经引用点核验后新增独立systemd drop-in，仅设置非秘密公开开关`PHASE1_WORD_REPORT_ENABLED=true`；重启后PID=`2367815`、NRestarts=0。最终完整production 00401 E2E PASS：发布2组/10维/90题、90答、提交幂等、v1/v2结果、筛选、三Sheet`2/91/91`、正式报告approved、10页PDF、审计2，exact cleanup全部0且所有表baseline完全一致。
- 最终只读收据：server/process=`f850575b...`、index=`593d4a20...`、393前端文件、active paper0、10维/90题及两项题库SHA保持不变、00401历史0、v1/v2文案=`66/124`、包=`1/1`、文案SHA=`c52b2b19.../35cf08ee...`、模板=`54b167fc.../a814c36e...`、renderer=true、最近10分钟fatal/panic/permission=0。00401 production正式报告由RED转GREEN。

# 2026-10-09 00401 production全量staging替换已自动回滚（真实E2E发现MySQL5.7兼容阻断）

- staging SSH恢复后，同一只读inventory实测：server/process=`f2940fc5...`、index=`593d4a20...`、MySQL8.0.46、10维/90题及题库SHA与production完全相同、v1/v2文案=`66/124`、包=`1/1`、模板SHA分别=`54b167fc.../a814c36e...`。production对应v1为0/0/absent，v2文案SHA与staging一致但模板不同；production不存在00401版本元组或repo关系历史，故授权的00401历史删除闭包实际为0。
- Schema逐行对比确认：staging仅多`el_competency_migration`标记表；其余列差异为MySQL5.7整数display width和MySQL8 `DEFAULT_GENERATED`元数据，FK SHA相同。production另有`el_qu.idx_qu_type_level_id(qu_type,level,id)`性能索引；未删除该生产安全增量，也未执行DDL。题库内容SHA相同，未重复写题库。
- 受控controller先做active paper=0、00401历史=0、内容SHA和payload门禁，随后创建全量备份=`/opt/talent-assessment/backups/phase1_staging_sync_20261009211916`，再替换staging精确后端/前端/v1+v2模板并导入66条v1文案及production环境批准包。controller验收PASS：PID2366553、server/process=`f2940fc5...`、index=`593d4a20...`、模板=`54b167fc.../a814c36e...`、健康200。
- 随后的真实production 00401完整E2E在创建测评入口被`管理特质保护校验失败，请稍后重试`拒绝；exact cleanup和全表baseline恢复均PASS。该失败证明用户已知并接受的FB-219回退仍会让staging FB-218后端在production MySQL5.7环境阻断公共exam保存，因此不能将controller结构验收当业务GREEN。
- 按四阶段门禁立即从上述全量备份恢复数据库、server、dist和全部export templates。回滚后PID=`2366975`、NRestarts=0、server/process=`f850575b...`、index=`abf93dd1...`、健康200、active paper0、10维/90题不变、00401历史0、v1文案/包/模板继续0/0/absent、v2继续124/1及模板`f9859993...`。production当前状态与尝试前一致；release marker未移动。

# 2026-10-09 00401 production全量staging替换预检BLOCKED（staging SSH不可达，零写）

- 用户将此前“仅同步00401且保留较新production程序”的边界改为：完整使用staging后端和前端替换production，同时仅删除00401历史；接受production的FB-219及其他较新程序随整套程序回退。该高风险范围已经结构化确认，但尚未执行。
- staging `20.200.136.133:22`的只读inventory上传首次及唯一重试均在SSH命令启动前connect timeout；staging远端脚本、SQL、上传、备份、替换、restart均为0。按两轮门禁停止继续重试，未用旧状态冒充实时staging事实。
- 本地找到并逐字节核验与当前staging标记精确匹配的既存封存物：backend `Go-based Refactored System/bin/server-staging-fb218-v2-linux-amd64` SHA=`f2940fc5ea51edffc4f325df1f461f3ba4e86868df3aa0594af95764e880d61e`；frontend `Go-based Refactored System/bin/staging-dist.tar.gz` SHA=`a74789a88be767bb35fdabb738e214dad7bea7d9be0c4b09ff2bf974defa4ece`，393文件，内部`index.html` SHA=`593d4a20d890d73bf47f9a519a22fffbdb4b539afd5e1cc1dcc2fdebff73cfde`。这只证明封存字节匹配已有标记，不能替代当前staging DB/Schema/模板实时inventory。
- 新增只读工具`scripts/tools/phase1-live-readonly-inventory-20261009.sh`，只输出非秘密运行时SHA、health、MySQL版本、00401 10维/90题分布与内容SHA、Schema SHA、报告包/模板及历史范围计数；临时MySQL client为0600并由trap删除。本次production运行PASS：PID=`2355899`、NRestarts=`0`、server/process=`f850575b...`、index=`abf93dd...`、MySQL5.7.44、active paper0、10维/90题及`90|90|80|10|62|18|10|90`分布、runtime refs0、v1文案/包/模板均0/absent、v2为124/1且模板SHA=`f9859993...`。production临时payload已清除，未备份、DDL、DML、替换或restart。
- 结论：完整替换需要先取得同一时点的staging实时Schema/内容/模板收据并比较；SSH网络恢复前不能满足“以staging为基准检查后最终更新”，因此本轮fail-closed停在零写preflight。production继续保持原标记，不移动release marker。

# 2026-10-09 00401 production正式报告同步PARTIAL（v2已安装，v1写前阻断）

- 用户明确选择仅同步00401、保留较新的production程序和其他产品，允许沿用staging具名批准，并保留但不迁移旧00401历史数据。production先完成完整备份并安装staging精确v2资产：124条正式文案、production approved包、v2 DOCX SHA=`f98599939e3bf7923abf8bd457e30259cde3fa68d6c4e9dc9d3288de0a2eae16`；备份=`/opt/talent-assessment/backups/phase1_report_20261009_8af8c5dd2963418f`。后端/进程SHA保持`f850575b1dfa6eac3f5b4533148baf715eabc8afa13ccd32c7655d0507a7d600`且未重启。
- 随后完整production 00401 E2E确认默认新卷查询v1版本元组，而production仅有v2包；报告验收RED，但finally精确删除本轮候选、卷、结果、run、报告、审计、文件及Redis token，所有业务表基线恢复。v2安装本身通过范围内验证并保留。
- staging SSH两次在远端命令启动前超时。后续从仓库权威材料确定性重建v1：66条内容（5 overall、7 template、2 group、50 dimension、2 validity），内容SHA=`3060bf06f3f52715c7cf9b05f277e4ccd723a7f571785c5a26918cc98d8dbb42`；精确staging v1模板存在于`docs/competency-phase1-report-fb170-pie-labels-outside.docx`，536894 bytes/SHA=`54b167fcc02737ba44e0432a930e1f5657db57cec87d9063f1ee7c111c4932bf`，与B区旧副本SHA=`a6031417...`明确不同。新增production批准门禁SQL `competency_017_v1_report_content_production_approval.sql`，绑定题本SHA=`f33b878e...`、内容SHA、Liming双职责具名批准、production环境及完整免责声明。
- v1部署未发生：第一轮controller把`state=0`误当全局active paper，第二轮又推测了不存在的`el_question/phase_order`结构；两次均在创建备份和业务写之前fail-closed，payload自动清除。只读诊断确认真实active paper=`state=1`且为0，服务PID=`2355899`、`NRestarts=0`、进程SHA不变、`REPORT_EFFECTIVE_ENV=production`。最终停止继续写入并清除payload/probe/receipt；production v1模板仍不存在，未新增备份、未导入v1文案、未重启。下一步必须先用实际表`el_competency_dimension.id`精确十ID和真实`el_qu/el_qu_repo/el_repo`关系完成一次全量零写preflight，再单次备份/部署/10页PDF E2E，不得继续猜Schema。

# 2026-10-09 00401 production正式报告审批包准备完成（未上线）

- 用户明确指定00401数据以staging为准，并确认客户在staging已测试通过。production待审批候选因此锁定为staging精确版本，不另造文案、规则或模板分支：内容源工作簿SHA=`edb9efd27ec86bc34db3a796c2022a99495fd9ec52e8a6580cd7404b2ab933b5`、内容语义SHA=`329409e408f10ec7f048a757e83c963b7736f66e41e8d601ff4397fd8545b48c`、124条内容、DOCX SHA=`f98599939e3bf7923abf8bd457e30259cde3fa68d6c4e9dc9d3288de0a2eae16`。
- 审批包位于[00401-production-formal-report-approval-package-20261009.md](00401-production-formal-report-approval-package-20261009.md)，已列版本元组、不可变SHA、staging真实10页PDF证据、免责声明、独立审批栏、preflight/备份/恢复副本演练/发布/验收/回滚门禁。014 SQL文本INSERT计数实测124；审批包Markdown诊断0。
- 该决定只关闭“采用哪套数据”的候选选择，不伪造production内容负责人/心理测量负责人姓名或时间，也不自动继承staging的`effective_environment='staging'`。本轮production写入、上传、重启均0；正式报告继续fail-closed，须取得production环境审批和上线授权后另行生成专用批准SQL并受控发布。

# 2026-10-09 production受控发布完成（范围限定GREEN）

- 精选提交已推送至GitHub，最终controller修正提交=`227bff9`。首个正式尝试在任何备份/写入前因payload成员比较错误失败；第二个尝试完成全备份并启动新程序，但错误地用port 80校验前端，acceptance失败后自动rollback=`ROLLBACK_OK=1`。回滚后主库MNG表0/005 repo0，server/process=`03397e0f…`、index=`f5cd615b…`、服务健康，证明数据库和运行时恢复成功。
- 纠正controller只把前端验收目标改为实际应用vhost `127.0.0.1:8090`，重新封存、合同GREEN并使用新stamp `3e3692ce14634021`发布。完整数据库、应用和系统配置备份=`/opt/talent-assessment/backups/production_release_backup_20261009_3e3692ce14634021`；controller输出`PRODUCTION_RELEASE_PASS=1`且exit0。
- 最终后端/进程SHA=`753fad7a6134139b11ed3285c418da092c160b4fe81b9baf53dbf70f6d3cf0fc`，前端index=`abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4`；服务active、PID2354770、NRestarts0、nginx 7进程，8092/8090 health及8090 root均200，最近10分钟error journal为空。
- 主库验证为MNG表14、005 repo2、state1 exam2、profile/snapshot/run/revision/reissue各2、reissue audit12；00501=`mng-00501-db-current-v1`，00502当前=`mng-00502-db-current-v2`。REPORT/MNG环境均为production，匿名直连管理特质详情401。DOCX/XLSX SHA分别为`05c55e77…`/`b0498249…`。未执行认证管理员浏览器链或客户状态切换，formal 002及competency 009/014/015继续明确排除。

[补充 - 2026-10-09] 真实管理员浏览器验收发现FB-219：MySQL5.7统一`utf8mb4_general_ci`被旧代码硬编码的MySQL8 narrow-edge签名拒绝，00501启用409。新增回归先RED后GREEN，只放宽到父子相同的合法utf8mb4排序规则；全Go和build通过，提交`f6d0719`。用户明确不走CI/CD，手工补丁先备份旧binary再发布SHA=`f850575b…`，PID2355899/NRestarts0/三HTTP200。00501 state 1→0→1、模板元数据、结果列表和reissue PDF真实浏览器均通过，最终双005 state1且新PID无management schema reject。只读BCrypt比较另确认admin仍匹配仓库已知历史默认候选；未打印hash或改密码，已列为需授权轮换事项。

[补充 - 2026-10-09 production完整E2E] 00501完整140答链最终单次PASS：13维/4模块、524150-byte PDF view/download同SHA、candidate/paper/run/reissue/file逐项清理，前后基线完全一致。首次清理器因合法underscore ID和production Python3.8不支持dict union两处工具假设失败，均在唯一marker下精确恢复后重跑GREEN。客户随后从外部管理员IP将00501/00502均设为state0进行中，未由测试覆盖。

[补充 - 2026-10-09 00401 production内容与E2E] 首次E2E在exam Save被“所选测评维度不存在或已停用”拒绝且cleanup/baseline均0漂移，证明此前仅发布Schema而无固定A/B内容。用户明确批准仅导入10维+90题、正式报告继续关闭。工作簿SHA=`828c4267…`；完整备份=`/opt/talent-assessment/backups/phase1_content_20261009_8b17bb3f57794f2c`。为兼容production旧48维唯一name/order约束，旧主数据order精确移至101–148，8重名追加“（历史）”；9个旧测评的冻结关联/结果逐表SHA不变，migration009未执行/无marker。官方API导入90题并通过80/10、62/18/10方向、逐维9题、导出回读和重复导入失败关闭。完整E2E最终PASS：90答、同卷恢复、提交幂等、10维/2组/效度/总体、v1=`10|2|1|1`、v2=`1|1|3|10|1`、筛选、三Sheet`2/91/91`；formal报告无审批包，三接口受控关闭且零写；删除后所有表基线逐项一致。

# 2026-10-09 production 005策略与迁移包（静态GREEN，MySQL5.7动态演练待批准）

- 用户明确：005发布production、保留TEST标注、不新增开关、客户决定是否启用；同时要求生成并评估production迁移包、清理Git范围、完成FB-218。
- `ManagementTraitsTestRuntimeEnabled`经RED→GREEN改为仅允许匹配的`local/local`、`staging/staging`或`production/production`；空/单边/不一致/`prod`/畸形仍关闭。生产包将两个005测评设为现有`state=1`可见禁用，客户通过已有状态操作启用。
- staging只读导出精确005闭包和4份PDF/两份模板；生成`scripts/data/production-migration-20261009`，顶层22项SHA通过。包含00401 007/008/010/011/012/013与MNG 001/003/004；排除formal002及competency009/014/015；有exact rollback和customer activation SQL。
- MNG 001父列门禁由`varchar<=64`收紧为三个父ID必须`varchar(64) NOT NULL`且字符集/排序规则一致；合同先RED后GREEN。Go全量、Windows build、Linux候选SHA`07d088b…`、前端35/602和production build均通过；FB-218异步失败关闭已包含在Linux候选。
- 尚未运行会创建/删除production服务器临时恢复Schema的MySQL5.7 rehearsal；未stage/commit/push，未production备份/DDL/DML/上传/重启。当前仍NO-GO。

[补充 - 2026-10-09] 用户授权临时Schema演练。审查迭代关闭客户state入口、rollback drift、restore routing、资产DB-file绑定、00502冻结profile旧文案、checksum CRLF、140/700唯一顺序等阻断。第二轮MySQL5.7.44 restored-copy全部GREEN并自动cleanup0；生产`element`仍MNG表0/005 repo0/active paper0，服务active/health200、server/process SHA仍`03397e0f…`、NRestarts0。最终本地Linux候选`753fad7a…`、前端index`abf93dd1…`；Go全量及前端35/611通过。未commit/push，未正式部署。

# 2026-10-09 GitHub与production发布前完整评估（NO-GO，零生产写）

- 用户要求提交GitHub后完整发布生产，并选择精选可复现提交、生产保留现有数据/迁移结构、复用系统Noto/WQY；随后明确选择在production启用005 TEST并复制staging基线。后两项相互冲突：复制基线包含业务DML和合成TEST人员/答卷/结果/报告，不是“仅结构”。按更高风险解释停止写操作。
- Git当前`master`相对`origin/master` ahead3，HEAD=`5217ce6558eb7876d9e4cc37f1a617296deb2590`；工作树407项（tracked modified74、untracked271、deleted64）。大量关键Go/Vue、迁移、模板和文档未跟踪，另有超长LibreOffice profile生成目录导致Git扫描warning；当前不能形成已审阅、可重现且无意外删除的精选提交，故未stage/commit/push。
- production安全凭据弹窗连接成功，只读preflight PASS：主机`iZ0yosjdcen2p4Z`、PID1195022/NRestarts0/active；8092/8090/80均200；server/proc SHA=`03397e0f…`、index=`f5cd615b…`、372前端文件；MySQL5.7.44、54表、可用磁盘约9.4GB、内存约5.8GB。`APP_ENV=production`，REPORT/MNG双环境变量未显式设置。
- production数据现状：active paper0、到期state0卷517、到期active competency0；66 exam/1980 paper/1448 candidate/33 tester/4862 MBTI answer；competency表6、result19、report3，但四版本列0、result-run/current六表0；management-traits表0、005 repo0；paper缺exam孤儿409。字体/LO/Chrome齐全，不上传Windows字体。
- 独立评估与CodeReviewer给出阻断：MBTI异步转换失败仍曾发布DOCX路径（本轮先RED后本地GREEN修复，尚未重新全量/staging/production验证）；现有production部署配置与当前工作树不可复现；005 production显式开关、MySQL5.7管理特质schema、staging基线最小/全量DML、TEST模板和回滚均无生产批准包；00401 production内容SHA/迁移范围仍未最终批准。生产未创建备份、未运行DDL/DML、未上传、未重启。

# 2026-10-09 主要新版链staging完整测试与临时垃圾清理（GREEN）

- 用户选择范围为00401一期胜任力、003 MBTI、00501/00502管理特质，001/002只读冒烟；垃圾仅清理明确临时产物。测试前完整数据库备份=`/opt/talent-assessment/backups/fulltest-20261009-104034-ba76ee88`，root0700/files0600，数据库gzip SHA=`7a89783c7ec8b3cd194ec393b46dd77e36f608b4056510215cd95f6ecfe26908`且SHA/gzip复核通过。
- 00401真实staging完整链PASS：新建/发布2组10维90题、重复发布、同paper恢复、90题真实保存、manual提交幂等、10维/2组/效度/总体、筛选、三Sheet导出、已批准v2强类型DTO、生成并下载10页PDF、审计2；exact cleanup=`0|0|0|0`，运行时四项基线前后同为`10|24|48|24`。同步修复旧验收脚本的空库假设、过期XLSX文本断言、content package=1假设及遗漏v2 result-run清理，产品代码未因此改变。
- MBTI真实staging完整链首次发现FB-218：48题保存、ESTJ计分回读后，旧直接LibreOffice命令两次超过180/240秒。先新增RED，后改为复用已用于胜任力的`libreofficepdf.Client`、90秒context、隔离临时workspace，并禁止转换失败时把DOCX冒充PDF成功。最终后端SHA=`f2940fc5ea51edffc4f325df1f461f3ba4e86868df3aa0594af95764e880d61e`、PID=`22668`；复测48/48、ESTJ、完整版/简版PDF、16+16模板、匿名模板401/403均PASS，exact cleanup四项0，MBTI基线`73|1491|1352|2653`前后相同。
- 00501/00502保留基线各1份completed/140/140/50/13维/4模块，真实管理员生成新模板SHA报告、view/download均200，字节647985/648159；完成验证后精确删除本轮revision2/audit/PDF并恢复两份revision1 current指针，报告总数恢复3。模板管理此前真实元数据/下载/上传继续有效。
- 00101只读detail及9658-byte XLSX导出PASS；00201 detail PASS，通用原始数据导出被管理特质legacy guard按设计403，不放宽保护。全Go任务、前端35文件602项通过；Docker未安装，Node16.20.2、Playwright1.59.1及Chromium可用，浏览器/API轴正常执行。
- 本地清理仅删除B区logs、前端空coverage、测试screenshots及根空screenshots，共54文件/13616966 bytes；不删results、release/evidence、客户材料或正式备份。staging只删两次失败MBTI运行拥有的2个旧LibreOffice profile，所有本轮`/tmp`前缀残留0；应用tmp业务上传目录未动。最终三服务active、内外health200、active paper0、legacy完整dump SHA前后同为`e30e69817b8fa0945bc36df0da4300bfe08aa40967ac75168183ef577843cad5`、fatal/panic/permission0、Nginx5xx0。production未访问。

# 2026-10-09 00501/00502题库、共用模板与程序完整staging发布（GREEN）

- 用户授权仅staging完整发布；production未访问。发布前active paper=0，创建受限全量回滚备份`/opt/talent-assessment/backups/mng005-template-20261009-101929-e50ad044`，包含数据库、server、393文件前端、configs/private、systemd及精确005回滚DML，目录0700/文件0600且`SHA256SUMS`复核通过。
- staging当前PID=`19566`、`NRestarts=0`，talent-assessment/nginx/mysql均active，内外health均200。磁盘与进程server SHA=`5bf13ca9e317a91069463b3a064dbc8722b091676c80b78cfbc620efd55416bf`；index SHA=`593d4a20d890d73bf47f9a519a22fffbdb4b539afd5e1cc1dcc2fdebff73cfde`，前端393文件；模板SHA=`05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c`。
- 现有00501/00502原位更名为“管理特质测验基层员工新版”“管理特质测验干部新版”；只改00502 V67/V96为用户确认文本。两库继续各140题/700答案；00201/00202全源数据receipt前后同SHA=`13ddd0eef9e9278d2f5b221ca327a5ebdaa1491f086ab91d4edd0e4029c80e5e`，管理特质历史表receipt前后同SHA=`9e700cd9e953c7a256bbe42a21c2a4377303fcdc013fe8c28a58a306d5236c0d`，冻结题420、结果3、报告3均未变。
- 真实管理员浏览器验证共用模板元数据/下载/同文件上传/刷新均HTTP200；DOCX 530193 bytes、90控件/6图表/5数字标签/0外链、valid=true，上传生成自动备份`management-traits-002-test-only-v2.docx.20261009_102458_000.bak`且生效SHA不变。`#/qu/template`真实页面显示共用卡片、校验通过、SHA及上传下载按钮；匿名模板接口401。
- 发布过程中两次在远端写入前被精确门禁拦截；一次完整切换后因PowerShell stdin末尾CR触发自动rollback，server/template/data均恢复，但前端已完成新393文件切换。随后从全量备份准备旧前端回滚目录并完成后端/template/data发布；最终payload清零、应用fatal/panic/permission错误0。该过程证明回滚路径实际生效，但不构成production批准。

# 2026-10-09 00501/00502题库名称与管理版题干纠正（本地GREEN）

- 用户确认00501/00502显示名分别为“管理特质测验基层员工新版”“管理特质测验干部新版”；00502 V67最终采用“当我接手具有挑战性的工作时，我通常能鼓励大家创新，并提出创新的解决方案。”，V96采用客户管理版“当下属反对我的某个决定或者工作安排时，我会保持冷静和理性来应对。”。
- [纠正 - 2026-10-09] 先前“V67/V96不修数据库”的决定继续约束00201/00202和历史冻结数据；本次新决定只修005 fresh创建定义。00501保持00201已核140/140原文；00502仅覆盖V67/V96，其余138题、题序、700选项、分值、方向和公式继续复制00202。
- fresh fixture SQL及opt-in baseline writer两条创建路径已同步；合同先后锁定两条RED，最终exit0；受影响Go package、全Go测试和server build均exit0，诊断0。未写local/staging/production数据库；用户选择本地完成后再单独确认staging，production未授权。

# 2026-10-09 00501/00502共用报告模板管理（本地GREEN）

- 用户确认00501/00502共同管理一份模板，不拆成两个文件，也不增加预览、版本列表或手工回滚。既有报告模板页新增一张共用卡片，支持元数据、下载、选择DOCX、二次确认上传、loading、失败保留文件与刷新。
- 后端新增TEST门禁内的模板元数据/下载/上传路由；20MiB上限，拒绝损坏ZIP、外部关系、宏/嵌入对象及不兼容90控件/6业务图/5数值标签契约。上传先写自动备份，再同目录原子替换；模板读取和替换共享RWMutex。
- 新TEST与reissue报告持久化当前实际模板SHA。历史报告读取只信任当前模板和服务器自动备份文件的实际SHA，继续返回原PDF而不重渲染；旧固定SHA公开方法保留为兼容包装。production关闭TEST assembly时不注册模板管理路由。
- TDD RED为后端新符号缺失及前端4项失败；最终Go全量/build/vet均0，前端35文件602项及production build通过。通用lint脚本受既有`.eslintignore`阻断，直接lint仍有目标文件既有格式债务。浏览器因本机8092后端未启动停在登录验证码500；未真实上传、未staging/production部署。

# 2026-10-09 MNG005 staging真实浏览器验收（PARTIAL：仅native event）

- 用户确认的真实admin页清page/context mocks后getInfo200；00501/00502旧URL均自动进入新管理页，各1completed/140/140/50.00/13维4模块且旧list403=0。普通历史002旧URL保旧页/list200；冻结兼容002旧URL进新页58.64。快速reload一次触发公共重复提交保护，页面retry真实恢复，不改代码。
- 两产品原reissue0；UI各生成1份、重复复用同ID/同file、view/download均HTTP200。最终旧TEST2＋独立reissue2=4份保留；reissue audit generate/reuse/view/download=2/3/2/3。无认证独立context两读取401且无PDF bytes；无低权账号、token/storageState输出。
- 新PDF SHA分别4d40da09…/2b28aba0…，DB/private/API/local一致、private regular0600。与已hardened retained baseline逐页文本9/9、144-DPI像素9/9、drawing signature全同；canonical140/100/40/13/4、identity-v2 COMPLETE、PDF object privacy findings0/attachments0。当前PyMuPDF对旧baseline也产生相同坐标提取漂移，旧绝对坐标子断言不兼容当前库，不是产品布局变化。
- integrated browser native download event 60秒未观测；真实下载HTTP200字节通过同一认证UI响应内存安全留证，故状态精确为PARTIAL_NATIVE_DOWNLOAD_EVENT_UNOBSERVED。SSH只读PID15554/d30c…/c4f3…/health200，保护DB5e80140d…、冻结002 58.642639、应用错误/Nginx5xx0；无代码、部署、restart、答案或production操作。证据目录mng005-staging-browser-acceptance-20261009。

# 2026-10-09 staging rollback canonical schema evidence（reviewer blocking关闭）

- 原rollback证据仅记录三张新增表0行，未证明结构在切换中不漂移。脚本新增确定性canonical receipt：3表engine/collation、29列name/type/null/default/extra/order、20 index-column的name/unique/order/column、7 FK-column的name/ordered cols/ref table+cols/update/delete；HEX编码与NULL marker消除文本歧义。
- 回归先RED确认原脚本缺canonical字段和PRE/OLD/POST采集；GREEN后Node合同与bash syntax均0。fresh preflight active0/新增表rows0；完整staging rollback复跑PID14293→15461→15554、恰2次restart。PRE/OLD/POST三份6328-byte receipt逐字节相同且SHA均`1a9f16e…`，匹配003/004迁移预期schema。
- 旧恢复演练`1ebadfd6…`只含两张reissue表且缺engine/collation/default/extra，明确不可exact比较，不后验造证。本次实时三阶段证据关闭reviewer缺口；旧失败和成功演练全部保留。
- 最终PID15554/NRestarts0，三服务active、内外health200；server d30c…/index c4f3…/393 manifest cfe7…；配置ed9837…、受保护DB5e8014…、PDF3d2095…、冻结002 58.642639保持；state1/overdue/additive rows/auth residue/app errors/Nginx5xx全0。完整见management-traits-staging-rollback-drill-20261009.md；仍仅staging，不授权production。

# 2026-10-09 staging发布包分离与完整rollback drill（PASS）

- 仅20.200.136.133，production0。runtime 400文件归档SHA`e077c400…`与acceptance 11文件归档SHA`4a5122df…`物理分离；展开393 dist逐文件＋所有binary/config/SQL/script/archive扫描，高置信秘密0。JSEncrypt marker为生成代码而非PEM body。
- 受限备份实核旧server4179…/index52eec…/393 manifest8633…；成功轮PID13961→14209→14293、restart2，旧health200/legacy200+401/auth reissue404，新d30…/c4f…/393 manifestcfe…/legacy200+401/reissue unauth401+auth409。首轮错误用unauth401判断旧路由，失败后新发行恢复并额外启动1次，证据保留；不计入成功轮2次。
- 配置ed9837…、受保护DB5e8014…、私有PDF3d2095…、冻结002 58.642639、三新增表0全程不变；三服务active、内外health200、errors/Nginx5xx0。最终runtime testbinary0；受限backup内历史5个均0600/顶层0700；payload/上传脚本/短时Redis认证及header0。完整见management-traits-staging-rollback-drill-20261009.md；staging PASS不授权production。

# 2026-10-09 release candidate 210616 reviewer evidence addendum（本地证据收口）

- 当前候选`release-candidate-20261008T210616Z`明确替代并保留旧`203459`；candidate ZIP/manifest、旧收据均未改。新C区`evidence-addendum.json` SHA=`a8c19cbf…`，binding root=`c0b5d8d3…`，状态仅`GREEN_LOCAL_EVIDENCE_BOUND_NOT_DEPLOYMENT_APPROVAL`。
- 从原打包命令恢复并重放exact输入：cmd/internal/pkg全部Go、go.mod/sum、前端src/public及4构建配置，共654文件；当前重放aggregate=`0e1e9d1…`与candidate manifest精确相同，未拿变更后源码追认。Linux SHA=`331862e2…`、Windows同源gate SHA=`0d8589d1…`均绑定该aggregate。
- DB证据仅脱敏投影：before/after同receipt SHA=`913e3c87…`；20 obsolete表与20 orphan关系全0；00501/00502 retained计数各1/1/1/1/1/1/140/140/700/1/13/4/1，五类report chain均0；generation=`1791483691965`且只留hash IDs。legacy copy/manifest/input/mapping/field-contract hashes均绑定，无raw ID/PII/secret/identity-field hash。
- 首次HTTP gate保留`HEALTH_TIMEOUT_120S` receipt；日志有listening marker但不能证明probe前未ready，故只归类`unattributed startup/probe timing, later success`。后续同源Windows gate health/legacy200、participant/admin/reissue/formal404；WSL registry distro0，Linux ELF仅`go version -m`静态确认Go1.26.2/CGO0/linux-amd64/trimpath，未冒Linux runtime PASS。
- 仅evidence/docs；无source/test/remote/DB/deploy。JSON解析、binding root、654计数、source replay、敏感值扫描、Markdown链接、candidate/旧receipt immutability和diagnostics均通过。

# 2026-10-09 FB-217 localhost解析与数字绑定（本地GREEN）

- `resolveServerRuntimeControls`现接收可注入纯resolver；仅精确小写localhost触发一次固定2秒`LookupIPAddr`，结果必须非空且全部loopback，优先127.0.0.1否则排序IPv6，并用`JoinHostPort`绑定规范数字IP，避免监听阶段二次DNS。
- 显式IP不触发resolver；其他hostname、localhost大小写/空白变体、解析错误、空结果、非loopback或混合结果均在DB初始化前拒绝。空host默认仍`:port`且只允许Worker默认启用；disable=true继续要求精确local+loopback。
- RED为resolver参数缺失编译失败；最终focused/full Go、Windows build、vet、diagnostics均0。独立复审补强空白变体和IPv6确定性；无服务启动、DB/远端/部署。

# 2026-10-09 通用本地安全启动控制（本地GREEN，待独立review后打包）

- server新增host绑定：空/0.0.0.0保持原`:port`，显式IP或localhost经纯resolver校验并用`net.JoinHostPort`；域名、host:port等畸形值启动前拒绝。
- `LOCAL_DISABLE_BACKGROUND_WORKERS=true`仅在`APP_ENV`精确local且bind host为loopback时允许，其他组合启动前fail-closed；同时控制competency与management expiry Worker。默认false，staging/production旧监听和Worker行为不变。
- 管理特质TEST route/runtime仍只由`REPORT_EFFECTIVE_ENV`+`MNG_TEST_REPORT_ENV`独立门禁；local candidate可用APP_ENV=local/127.0.0.1/disable=true并把两个TEST环境设production来省略TEST routes。无HTTP请求控制路径。
- focused RED已实证缺字段/resolver/router helper编译失败；focused GREEN、全Go package GREEN、Windows server build0、vet0、diagnostics0。五个改动Go文件规范化后与gofmt一致；首次因terminal继承GOOS=linux产生Windows exec格式错误，进程级GOOS=windows复跑通过并恢复原值。
- 未启动服务、未远端/DB/部署，也未构建新release candidate package；按确认顺序先交独立CodeReviewer，PASS后再产出可复现候选包。

# 2026-10-09 production-like candidate HTTP gate（启动前安全阻断）

- 仅本机评估release candidate `cc6665ec…`＋393-file dist；404项SHA全核0错、Linux ELF/50128192 bytes。生产门禁源码合同13/13通过，但不是candidate HTTP实跑。
- 安全组合确定为APP_ENV=local＋REPORT_EFFECTIVE_ENV=production＋MNG_TEST_REPORT_ENV=production；前者避免production overlay，后两者令MNG TEST assembly关闭，必须准确标为production-like gate而非完整production config启动。
- actual `cmd/server`仅拼`":port"`，ServerCfg/环境无host/address，immutable candidate无法绑定127.0.0.1；WSL/TCP proxy/firewall均不能产生所需loopback listener证据。router另无条件启动competency expiry worker。按用户边界在启动前停止，未短暂暴露0.0.0.0、未启动18093/23318、未DB/HTTP candidate操作。
- 现有23316 SSH、23317 Memurai、18092 backend、18080 frontend仍各唯一127.0.0.1 listener，现backend health200；三个既有private artifact存在/nonreparse且未输出内容。完整限定见production-candidate-local-http-gate-20261009.md；当前runtime gate BLOCKED，不得把static GREEN或当前dev health冒作candidate PASS。

# 2026-10-09 FB-216 管理特质TEST运行时生产fail-closed（本地GREEN）

- 新单一门禁`ManagementTraitsTestRuntimeEnabled`只信任进程级`REPORT_EFFECTIVE_ENV`与`MNG_TEST_REPORT_ENV`，trim+lower后须同为local或同为staging；不使用APP_ENV，空/单边/不一致/prod/production/畸形全部false，HTTP无法覆盖。
- disabled assembly不构造管理特质runtime、不执行启动Schema precheck、不注册participant/admin/result/TEST report/reissue/formal routes、不启动expiry worker；精确前缀在JWT前404。旧002 paper/report路由和legacy guard继续注册，candidate/tester旧链通过disabled identity service直通，Exam Detail不探测sidecar。
- RED为缺helper/factory及production仍注册18路由；GREEN focused/affected包、全Go6501/0/14、build/vet/diagnostics0。未改运行配置、未重启当前005 local服务、未DB/远端/部署；production正式能力未来需独立assembly。

# 2026-10-09 MNG005 local readiness harness收口（GREEN_LOCAL）

- 唯一旧RED是C区harness exact expected receipt漏`legacyMarkerRejected=true`；producer实际拒绝旧固定marker，符合active generation/evidence安全决策，故不删producer字段。fresh RED exit1→仅补测试期望→GREEN exit0；完整7字段无其他遗漏。
- fresh focused 005前端208/208、全前端34files/598tests、diagnostics0。Go/Vue产品源码、SQL、模板、runtime和DB未改；复用同日SHA锁定全量证据（Go6484/0/14、coverage44.3%；前端coverage60.25/95.53/39.13/60.25）。旧RED verdict保留，新`production-readiness-local-rerun-20261009-040307`为GREEN_LOCAL。
- 未browser E2E/staging/production/shared DB/部署；本地GREEN不授权生产，外部门禁和环境skip继续明示。

# 2026-10-09 MNG005 reset head canonical bytes（FB-215，本地GREEN）

- `scanResetJournal`对legacy outcome-prefix及新head-SHA-prefix两种head文件名均重新按writer相同`MarshalIndent("", "  ")+newline`序列化，原始字节必须精确相等；空白、字段重排和内容篡改全部拒绝，历史文件原地保留且fail-closed，不自动改写/移动。focused FB-212～215、完整helper package、reset helper build及server build通过；无remote/DB写/新reset sequence。当前launcher没有reset-journal只读命令，故未冒充actual scan。

# 2026-10-09 MNG005 reset journal final hardening（代码完成，actual待核）

- FB-212～214补齐exact filename/schema/sequence/status/count/link门禁；pending恢复由数量比较升级为排序`nameHash/contentSHA/bytes` multiset，拒绝同数量替换；唯一最高next orphan outcome可补写缺失immutable head，多orphan拒绝。现存sequence1～3旧head只读兼容原命名，新head绑定内容SHA。focused/helper/actual standalone结果待本轮终验追加；无remote/staging/shared DB。

# 2026-10-09 MNG005 reset journal v2（代码完成，actual待核）

- reset helper由mutation后单receipt＋可覆盖master改为intent-before-mutation、outcome-after-postcondition及immutable chain-head；所有active v2状态由目录扫描重建，旧`reset-chain.json`和`reset-NNNNNN.json`只读排除，首v2 intent记录legacy master SHA。
- pending intent只在当前状态精确等于持久pre或expected-post时自动写`recovered-no-mutation`/`recovered`；混合状态失败关闭。pre保存reports/audits/files及retained runs/products，文件名只保存SHA。
- report-only v4不再保存raw examId，commitment和Python oracle消费examIdHash；递归scanner新增19位ID和UUID拒绝。
- Actual：generation `1791483691965`一次有效report-only成功，sequence1 pre-reset、sequence2 post-reset，后者删除2 reports/2 files；sequence3 standalone删除全0并保留2 runs/products。新report-only core SHA `86f2a365…`，PDF SHA `b1777ded…`/`bf92f1ad…`。未prepare/答题/远端写。

# Project Memory

## 2026-10-09 MNG005 reset receipt hash chain（FB-207）

- reset helper新增无raw ID/secret的append-only indexed JSON receipt与master hash chain，字段含sequence/mode/active generation、实际deleted report/audit/file、retained run/product、manifest/evidence/helper binary/source SHA及UTC timestamp。report-only不再硬编码cleanStart，v3 final receipt绑定pre/post reset receipt SHA并解析真实launcher stdout。
- focused RED为类型/helper缺失编译失败；GREEN后Go测试、helper build、Node及PowerShell syntax通过。首次actual预跑发现Go map JSON不保证status首字段，launcher安全过滤因此丢弃输出；已改ordered struct，失败链节保留。无prepare/280、remote/staging/production。

## 2026-10-09 MNG005 generation-aware local baseline complete

- 仅本机isolated copy；staging shared write/deploy/production均0。新active generation=`1791483691965`，00501/00502各由正常synthetic admin登录和真实UI完成140题、manual submit，创建时v2 hash-only evidence含exam/run/title/name/gender/phone hashes及`expectedReports=0`；LocalAppData按generation独立SID-only evidence和DPAPI private v3 manifest绑定。旧固定marker baseline目录/记录保留，仍无创建evidence且不可复用；runtime只指向新active generation。
- prepare前两次失败分别精确清理1/2条新generation链，旧baseline/fixture不动；第三次两链成功。随后发现safe manifest缺product/question version，透明补齐这两个DB可独立验证字段后、在首次private capture前重新计算safe manifest SHA；独立创建evidence字节/SHA未改。actual inspector通过20个global orphan=0、两链各1/1/1/1/1/1/140/140/700/1/13/4/1、五类report数组空。
- 两轮report-only均`answerClicks/answerSaveRequests/startRequests/submitRequests=0`，每轮00501/00502生成→重复复用→view200→native download200；score50、13维/4模块。两轮独立PDF oracle均identity-v2 COMPLETE、9页/TEST/客户文案/AST 140题100正40反、xref1866/streams82、credential findings0；每轮结束精确reset。最终额外reset两次均reports0/audits0/completedRuns2/deletedReports0，active inspector仍reports expected zero、orphan0，服务health正常。

## 2026-10-09 MNG005 retained baseline independent evidence correction

- [纠正 - 2026-10-09] FB-204的DB/orphan检查有效，但其“private manifest已独立绑定baseline创建身份”结论有误：capture从后续DB回算身份，且只保存了一个未解析的后续report receipt SHA。现存baseline没有创建时独立receipt或当时受保护的SHA，因此不得复用，也不得以当前DB生成的新attestation追认。
- FB-205改为v3：创建时hash-only evidence含product/exam/run/title/三identity field hashes及`expectedReports=0`；成功prepare同一launcher调用把证据复制到LocalAppData SID-only文件并将SHA绑定到DPAPI runtime。report evidence不存在时不声明report绑定。现存两条记录未删除、DB未写，actual inspect必须以`BASELINE_EVIDENCE_MISSING_NOT_REUSABLE`拒绝，后续另建新baseline。

## 2026-10-09 MNG005 retained baseline exact private manifest / orphan closure

- FB-204完成：工作区外SID-only DPAPI manifest持久化两条retained 00501/00502的raw synthetic candidate/paper/snapshot/run/receipt及五类report ID数组；launcher原子写、stdin供Go inspector，raw child IDs不进CLI/stdout/workspace。workspace safe receipt只存ID/hash、版本/source、mapping/input/evidence/snapshot/field-contract/identity hashes和计数。
- actual copy DB只读：obsolete精确20表全0；20条parameterless global LEFT JOIN orphan关系全0。两retained链各1/1/1/1/1/1/140/140/700/1/13/4/1，revision/current/audit及reissue/audit均明确0。package test/helper build/all build/vet、actual partial inspector、normal InspectRuntime全0；DB写/delete/remote/restart0。
- 历史限制保留：obsolete删除前未保存每个child raw ID，无法后验枚举已删ID；global orphan0仅关闭parent-absent orphan风险，不证明不存在另一条完整且无法归属的非孤儿链。

## 2026-10-09 MNG005 partial cleanup inspector fail-closed

- FB-203修复：obsolete父锚点0不再直接成功；20张相关表按exact exam/bundle外键路径逐表读计数，任一残留失败。00501/00502现存基线身份由baseline manifest输入，exact身份及1/1/1/1/1/1/140/140/700/1/13/4/1计数任一漂移失败；全0才返回LOCAL_PARTIAL_BASELINE_ABSENT。无LIKE/FK关闭/写库/远端/restart。

## 2026-10-08T15:43Z 005共享staging保留TEST基线完成

- 仅20.200.136.133 staging共享element；production0、在线应用部署/替换/restart0。最终PID2746/NRestarts0/backend4179…/front52eec…/healthok，active paper0；005管理UI未发布。
- 前一wrapper失败实际无残留。首续跑harness因wrapper遗漏APP_ENV=production而仅分类database connection并exact defer清0；私有失败流保留backup。原binary7bbc8ec2…不改，第二次补online config overlay且REPORT_EFFECTIVE_ENV仍staging后exit0。
- 保留00501/00502各repo1/140/140/700、exam/profile/candidate/paper/snapshot1、raw3答案140、completed run1/13维/4模块/receipt1、TEST report/current/generate-audit1，总体50.000000；report总数1/1/1→3/3/3。私有PDF d7fd78bb…/647985与b392cff6…/648159均root0600并同DB。
- DataSnapshot原字节SHA分别26334021…/10bf4c3a…，DB自算同；00201/00202 source stream SHA写前后159b225f…/a071422a…不变，旧报告文件5bdaebeb…不变，nonowned005冲突0。
- 新safe receipt v2 SHAe490f838…含每产品exam/run/report/title/PDF/DataSHA/身份字段hash及identityCommitment，不含原始run/report ID/人员明文；只读升级前后DB 2/2/2/2不变。
- hardened oracle PASS：canonical AST 140题/100正40反/13维4模块，raw3独立全50；两PDF各A4九页/36段/5环13柱1常模线/4星/TEST/外链附件凭据0。临时upload/schema/user/runroot0；root0700受限backup、ownership-complete及私有reportroot保留。完整限定见management-traits-005-shared-baseline-staging-20261008.md。

## 2026-10-08T15:24Z 005共享staging保留基线：网络阻断于上传前，业务写0

- 用户明确授权共享`element`持久TEST基线且非production；fresh只读核`vm-ubuntu-go-dev`、active卷0、005占用0、00201/00202各1/140/140/700、管理特质11表、现有report revision/current/audit=1/1/1、reissue/draft/formal表0。在线PID2746/NRestarts0/healthok、backend4179…/front52eec…；冻结002仍58.642639。服务器TEST XLSX/Word SHA为b0498249…/05c55e77…，LO24.2.7.2。
- 新受限备份`mng005_shared_baseline_c51130cb775ba019`：root0700/files0600；DB gzip SHA068cd193…、application f998399c…、report assets 97fb3b30…，gzip/tar/SHA校验通过。三次命令的非业务工具失败保留：root目录普通读取、root glob展开、最终CRLF破坏`wc -l`；均发生在对应文件生成或完整校验之后，不冒一次全绿。
- 新opt-in一次性handler test harness不监听/不Worker/内部消费DSN，计划clone独立005→canonical frozen profile→生产identity/create/140 raw3/submit/现有report链；失败自动exact-owned清理，成功才保留。默认SKIP测试编译0、vet0、Linux test binary 52,260,546 bytes/SHA7bbc8ec2…；回滚脚本bash-n0，但gofmt门禁仍未关闭，未上传。
- 实际上传首次与唯一重试均TCP/22连接前timeout，SCP0/远端命令0；因此共享库repo/exam/person/paper/run/report/PDF写入0、部署/restart/cache/config/production0。受限备份保留，rollback未执行。当前状态`BLOCKED_NETWORK_BEFORE_UPLOAD`，不得称005在线UI或持久基线完成；完整限定见`docs/management-traits-005-shared-baseline-staging-20261008.md`。

## 2026-10-08T16:20Z 005 identity-v2 重跑受登录检测阻断；预算门禁与回收已完成

- 仅测试/隔离运行辅助代码与C区文档；产品service/source逻辑、staging共享库、production和部署均0。PDF扫描默认单object由8MiB收敛为16MiB，总object text+raw+decoded为128MiB；10,688,445-byte合法流通过，16MiB+1与总量+1结构mutation均拒绝。identity-v2公式12个绑定分量逐项mutation均改变commitment；canonical AST合同保持GREEN。
- 新E2E工具已具备fresh 00501/00502创建/冻结/各140次真实UI raw3/提交/生成复用/view/native download，先写不可变`e2e-core-receipt.json`，再写含product/formula/script/core receipt/title/exam/run/report/3字段/PDF hash的完整v2 commitment，最后才把hash锁定manifest交给copy-only cleanup。未取得新的成功产物：三次有界执行均停在外部Chromium正常登录后的脚本admin检测，用户确认页面为后台首页，但工具未取得verified getInfo，未继续生成PDF；不伪造JWT或绕过登录，不把旧PDF补写为完整commitment。
- 三次尝试实际各建立了一个00501 frozen/candidate/paper前置链；精确恢复工具只匹配3条`MNG005-V2-<13位时间>-A`、00501、profile/candidate/paper各3、report/audit0后事务清理，exam从74回71、paper1492回1489，未碰其他行。三个空/marker-only结果目录清理；没有新PDF可保留。
- 最终`VerifyRuntime`再次GREEN：copyOnly/source SELECT1142/crossSchemaFK0/Redis ownership/reports schema/loopback均通过，DB Check为81表/71exam/1489paper/14管理特质表/2个005 repo；18080/18092/23316/23317四listener均127.0.0.1。一次中间Verify失败是launcher向旧strict boot decoder无条件增加新字段导致`boot input`，改为仅v2 cleanup发送后恢复，不是DB隔离失败。
- 最终Node合同为canonical AST GREEN、12 commitment mutations GREEN、16MiB/128MiB budget mutation GREEN；PowerShell parser0、E2E syntax0、helper build/vet0、runtime测试16/16。历史两PDF与blocked receipt保持不改，runIdHash仍缺；任务状态BLOCKED_AUTH_DETECTION，不报告新PDF SHA/pages或identity-v2 COMPLETE。

## 2026-10-08T14:43Z 管理特质005 AST/PDF深扫完成；identity v2既有证据阻断

- 仅C区，无UI/DB/report/service/remote；两PDF SHA不变。AST exporter v2绑定真实identity+scoring源码及四node位置/hash，mutation拒绝alternate mapping/模块漂移并接受benign strings。
- PDF深扫两份共1866 xref object、82 stream；object/raw/decoded bytes=`254372/1030825/31595162`，附件0、encrypted/needsPass0、credential findings0。单个decoded font最大10,688,445 bytes，需16MiB per-object、64MiB total配置；8MiB门禁真实拒绝。
- identity v2公式已实现且未来E2E写runIdHash；旧receipt没有runIdHash，业务记录已清理、PDF无UUID，existing模式fail-closed且不造final commitment。完整关闭只能未来获准重跑E2E后在cleanup前保存run hash。

## 2026-10-08T14:33Z 管理特质005 PDF oracle四个blocking关闭

- 仅C区Go AST工具、E2E测试源码hash-only模式、Python oracle/测试及证据文档；未重跑UI/DB/report generation/services，PDF字节SHA保持`0850bc50…`/`9d925561…`。
- `mng005-canonical-contract.go`用`go/parser`从真实`ManagementTraitsDimensions()`消费的私有catalog严格导出13维/140题/100正40反和模块`self3/interpersonal4/task3/development3`；source SHA=`4b417188…`、function SHA=`51980057…`，不再由Python regex/重复表提供评分映射。
- E2E固定fixture现可对既有receipt/PDF生成hash-only identity contract；00501/00502 commitment=`eb8af41a…`/`7d5d5bde…`，oracle从PDF标签值重算3个field hash和总commitment，每份11位手机号精确1个，不把明文写入commitment receipt。
- 三份JSON递归扫描含Unicode key规范化、最多2层/64KiB Base64 UTF-8解码；两PDF正文、11项metadata、合计1866个xref object string及embedded attachment name/content扫描，credential findings均0、attachments0。Node exporter合同、Python AST/oracle和四文件diagnostics均0。

## 2026-10-08T14:31Z 管理特质005 PDF oracle审阅warning关闭

- 仅oracle/docs；未UI、DB操作、报告生成或服务操作。旧oracle缺独立4模块值/privacy verdict的单独断言RED已保留；GREEN从13维精确有理数按canonical映射独立聚合，self/interpersonal/task/development各一次且均50，不复制UI模块数。
- 两份PDF身份严格匹配E2E固定synthetic allowlist，姓名+手机号标记各2、11位手机号各1；PDF和3份安全收据JWT/credential发现0。AST/oracle/diagnostics均0，PDF SHA保持`0850bc50…`/`9d925561…`。

## 2026-10-08T14:18Z 管理特质005双产品客户模板报告E2E完成

- 仅local回环＋独立副本；copy-only 004 DDL首/重跑通过，owned后端因本任务能力更新由PID20036替换为9540，Redis12324/Vue7912/隧道19132保留。reissue未登录401，旧report/genericPDF仍503；无staging共享element写、部署、生产或历史PDF修改。
- 00501/00502各新建冻结synthetic测评、真实candidate UI 140×raw3并manual完成；00501在70题reload恢复。独立canonical mapping确认40反向一次后13维/4模块/总体精确50，两UI实际均50.00。
- 真实管理UI完成详情13/4、生成/幂等复用/查看/原生下载。修复purpose合同bug：后端中文固定用途被前端错误要求英文TEST；实际RED→精确常量GREEN，相关171pass/build0。
- 两PDF分别653766/653523 bytes，SHA `0850bc50…`/`9d925561…`；独立oracle均A4九页、36客户段、13维、5环+13柱+1常模线、4星、TEST、0 placeholder/外链，Word回归保护footer policy。DB原DataSnapshot字节SHA、文件SHA/bytes及审计闭合。
- exact cleanup后两个exam/report/audit/runtime私有PDF全0；下载synthetic证据保留。005 fixtures各1/140/140/700、draft/reissue schema、四本机服务保留；冻结002 58.642639和五SHA不变。完整失败与收据见[本地debug报告](management-traits-005-local-debug-20261008.md)。

## 2026-10-08 CodeReviewer 最后 atomic commit blocking 关闭

- `Write-AtomicProtectedJson`现先验证既有目标private ACL/nonreparse，再创建同目录temp；temp在commit前完成private ACL/nonreparse、DPAPI解密、JSON解析和全部预期字段验证。commit只使用`File.Move`首次创建或`File.Replace`替换，成功后仅设置`committed`并return；finally仅在未提交时静默删除temp，无生产目标postcheck/Set-Acl。
- Windows真实合同证明首次Move和既有Replace后的目标均保持仅当前SID ACL；commit前注入失败保持旧SHA且temp为0，字段失败和不安全既有目标均在hook前拒绝。parser均0、合同native0；真实`VerifyRuntime`通过且`runtime.dpapi` SHA保持`45FA2E5D5C9756A7B89D6455C7788CAFC7EB6916ABE083B218A8D6BEA4CE29E3`。无DB写/UI/remote/restart。

## 2026-10-08T13:20Z runtime.dpapi 原子写与Redis绑定不可变完成

- [纠正 - 2026-10-08] 13:11“两个blocking关闭”只完成Go owned-runtime gate；当时正常launcher仍自动刷新已持久化绑定且直接写密文目标。现正常路径只在五绑定全部缺失时首次绑定；部分缺失或任何stale值均拒绝且不写。人工显式`RebindRuntime`才可在live ownership/private门禁后只轮换绑定、不轮换秘密，本次未对真实runtime调用。
- Windows PowerShell原子DPAPI模式已验证：同目录CreateNew密文temp、Flush(true)、仅SID ACL/nonreparse、解密JSON/字段预校验；已有目标以`File.Replace`无backup提交，首次以无覆盖Move提交，失败保持旧SHA并精确清temp。PS5 `File.Replace` null backup继续使用NullString。首次合成Replace IOException保留，独立探针及后续完整合同通过。
- launcher可安全dot-source；FB-202实际Windows inspector默认skip，显式模式只传非秘密PID/port/exe/config path/SHA。package test/vet、合同、actual inspector、VerifyRuntime均0；真实密文SHA不变，backend PID20036、Redis PID12324、direct/proxy health200、NOAUTH保持。未重启、未远端/DB写/280 UI/部署。[完整纠正](management-traits-005-local-debug-20261008.md)。

## 2026-10-08T13:11Z 本机Redis进程／配置强绑定完成

- 仅local dev runtime，未部署、未远端DB写、未重跑280 UI。CodeReviewer两个blocking已用同一Go gate关闭：所有短时命令及正常backend启动前均要求127.0.0.1 exact port/PID、listener owner、Memurai executable、工作区外private/nonreparse config、完整config SHA、唯一安全指令及constant-time requirepass一致，并执行认证PING/DBSIZE；不再只凭密码可连判定owned Redis。
- RED为新binding/verifier符号缺失编译失败；GREEN package test/vet、verify/inspect/backend build、PS parser均0。首actual verify因PowerShell参数传递安全拒绝，改为整数PID内联＋config路径子进程环境后通过。实际Redis PID12324/127:23317、config SHA4c9b12f0…、listener/executable/ACL/nonreparse/hash/auth全部match、DBSIZE1。
- 精确停止旧owned backend PID2736并以新gate重启PID20036/127:18092，Redis/Vue/隧道保留。重启后verify0，direct/proxy/captcha200、图像有效、匿名Redis NOAUTH；不登录、不报告。收据及限定见[本地debug报告](management-traits-005-local-debug-20261008.md)。

## 2026-10-08T12:27Z 本机隔离Redis／前后端启动完成（仅启动验收）

- 未提交管理员登录/真实答题/报告E2E，005注册/新schema/formal审批/报告操作关闭；本轮新入口未独立review，前阶段用户交接PASS不扩大。源element全表dataSHAe6c7251a…不同10:34c27fdfb8…，原因未归因、不归因本地启动或回滚；历史整库同不能冒当前。首次SSH255连接前timeout、唯一重试0/stderr0；12:27:32Z主PID2746/restarts0/4179/52eec/health及冻结002分值hash保持。
- 已有本机Memurai4.1.2/API7.2.5独立PID12324/127:23317、Go PID2736/127:18092、VueCLI PID7912/127:18080启动；原SSH PID19132/127:23316保留。Redis认证/单DB0/初始DBSIZE0，不拷源session，验证码后4新键；不共享缓存/装服务/改公网或服务器config。入口http://127.0.0.1:18080/#/login，原账号密码由用户正常输入，不reset/伪登录。
- C区原launcher新六runtime模式、三编辑停止；B区ignored bin/mng005-runtime新Go及真实VueCLI外部config wrapper/tmp自有目录。原Go config/main/Redis/JWT/业务/vueconfig/package/tasks不改，原511源码/测试按上阶段afterSHA全同。前端先滤秘密只注公开代理，DB/JWT只stdin/DPAPI；Memurai密码文件工作区外私有明文/仅SID，不假称为密文。三文件ACL/nonreparse、三个owned进程argv三secret匹配0，原argv暴露P1历史保留。
- Go运行/inspect build两0/vet两0、PSparser0/JSsyntax0/diagnostics0；前端开发编译24491ms成功。实际CLI5＋boot拒绝6/native0，原stdin回归0；独立live direct/proxy/captcha200/图像valid/RedisNOAUTH/native0。actualbrowser本地登录DOM图像可见、resourcehosts127:18080，response事件空不冒全网络捕获。首RESP美元展开NOAUTHfalse与首captchaEnabled错误native1按真实captchaOnOff修正，失败保留；测试三编辑停止、不改API。文档patch重复path第一次工具拒绝未落盘，合并单header后重试。
- 运行后copy completed1/58.642639/profile1及manifest66277c8b/input7474e6aa/mapping原字节dc7faac6/field051dbe83同原；源SELECT1142/crossFK0。原Worker仅copy，新dev raw日志关闭不当生产全错误0；不要求copy启动后全dataSHA同源，不重跑6400/598/coverage。本机三终端及隧道保留，remote只读/deploy-restart-sharedDML-DDL-production-005注册-历史删除0。[范围、C4、收据与终端](management-traits-005-local-debug-20261008.md)。

## 2026-10-08T10:34Z owned debug凭据P1修正／轮换与DPAPI同步完成

- [纠正 - 2026-10-08] 下方“凭据只工作区外DPAPI/未输出”不能证明进程argv没有暴露：原恢复CREATE USER含随机密码经mysql -e执行，历史风险保留。本次仅该owned7081fbec31e3d105专属账号双host轮换，原positive_app不改；恢复脚本密码SQL改stdin并关闭xtrace，未重跑恢复或改副本数据/授权。
- 真实Git Bash合成mock回归RED secretInArgvtrue/stdin0/native1→GREEN false/stdin2/native0；非真实MySQL执行证明。远端精确ownership/root0700-0600/双host八项copy-only grants/roleproxy0及本地SID ACL均核验，再先旧DPAPI密文备份＋新待同步密文；随机32hex只SSH/MySQLstdin，不.env/argv/secretstdout。
- 远端轮换成功后File.Replace PS5 null变空路径失败，待同步密文保留，不再次ALTER/回滚旧pwd；NullString传.NET null后仅消费原pending，新Go连接native0、旧密文BLOCKEDnative1、原Check再次0/源SELECT1142/crossFK0。经验：PS5 .NET nullable string参数用NullString而非裸$null；失败后先验证既有pending，不重跑密码轮换。真实grant格式是backtick/原生排序，batch转义须--raw再精确比较。
- 独立10:34:38Z native0：78表原全部dataSHA c27fdfb8…同源/副本，源完整schema/data和原基线同，旧assets/PDF/cache同；PID2746/restarts0/healthok/4179/52eec保持，general_log当前0非历史无审计证明。23316保留/应用不启动/源业务写0/生产0，旧失败收据不覆盖。[完整风险、失败与新终验](management-traits-005-local-debug-20261008.md)。

## 2026-10-08T10:21Z 普通002恢复核验／独立副本本机连接完成（后台未启动）

- 最新用户限定恢复仅Go普通002，保005及冻结新版002；独立staging副本允许测试写入，不改共享业务库/既有账号/env/Redis/网络、不部署。当前普通002Save/旧评分报告导出路径已保留，无需重复回滚。发现tester仍URL/session提前return跳过metadata，真实SFC RED13fail→最小修复GREEN433；只有该SFC和单测变，511保护项509同/缺失0，其他Go/原005/身份strict名单0变。
- 聚焦后端440pass/0fail/0skip，原005最终35/0/0；server/allbuild/vet0，前端一次34files598/0/0/build0；未重跑全6400/coverage/真实全业务E2E，独立CodeReviewer仍主安排。原框架warning及历史失败保留。
- 实际staging owned副本talent_mng005_local_7081fbec31e3d105恢复78表/71exam/1489paper/11mng/跨库FK0，005真实repo0未注册。新专属账号仅localhost/127及副本权限，主element SELECT真实1142拒绝；本地127.0.0.1:23316 SSH隧道已保留，Go真实Ping/CURRENT_USER/SHOWGRANTS/native0。凭据仅工作区外DPAPI+仅本用户ACL，未输出/下载原身份秘密、不覆盖env。
- 当前源全表data-only有序dump与副本SHA c27fdfb8e235ed36923e68297419a3f4f6731ec49564620f438bb57e11140b5a相同；源库全schema/data及assets/PDF/cache前后摘要同。新受限backup mng005_local_7081fbec31e3d105/root0700/files0600、gzip/SHA51d0d45c…有效，原backup不改、仅私有restore流去实例SET。新库/新用户保留，未DROP/迁移/seed/主写。
- fresh指定1791298091700970647源及副本completed1/58.642639，manifest66277c8b…、input7474e6aa…、mapping原字节dc7faac6…、fieldcontract051dbe83…同；未重算/换码/删除旧PDF。10:21:06Z主PID2746/NRestarts0/healthok/backend4179…/front52eec…保持。
- 完整后台仍需Redis且固定监听全网卡，未擅用共享Redis或启动应用；本轮仅DB连接就绪，不冒local health/真实登录PASS。主下一独立review→缓存隔离/回环启动；005注册及发布仍后续。[完整范围、SHA和复用连接入口](management-traits-005-local-debug-20261008.md)。

## 2026-10-08 MT-005-ISOLATION 本地产品隔离GREEN（未注册/未部署）

- 最新005政策已完成本地切片，限定替代下方“政策尚未实施/新002强制草稿”的历史：普通002缺flag/false原新建，显式true提示005；005真实DB码强制单140题25min合法字段/持久draft，缺表仅关闭005。既有002draft同产品编辑/冻结保留，普通002不得新冻结升级；已有冻结profile优先只读返回，不改现场1791298091700970647或58.64。
- 新00501/00502明确staff/leader及独立mng-00501-v1/mng-00502-v1和各自db-current题本namespace；规范化schema/domain/精确13维4模块/反向一次/常模/客户205与Word引擎原样复用。修改前后合成002manifest/mapping四goldenSHA同，12个原核心源码逐SHA同；不拿本地fixture当现场真实hash。
- 客户端集中LEGACY002/NEW005/FROZEN_COMPAT002＋draft/unknown分流：严格server bool和sameexam授权冻结，code/URL/session不授权限；005未知/false无draft禁旧，普通多002legacy保留。普通002旧表单无新opt-in，005新版默认锁定；既有配置→005必须新建。实际repoList不加假条目，真实005未注册不能称可创建。
- 有效RED来源两005拒绝/Save九子项失败/candidate两错误解锁；最终全Go6400 PASS事件/0fail/11原环境skip/native0，server/allbuild/vet0/输出0；追加最新Go35/0/0。最终前34files586/0/0/build0，当前编译UI28桌面手机syntheticcasePASS/pageErrors0/forbidden0/closedtrue。全量Go在最后追加跨码测试前，追加后专项闭合，不冒另一次全6402。非coverage/realMySQL/真实人员报告/正式批准。
- 两browser失败原因实采为测试predicate挂载前null.__vue__，只补就绪守卫后GREEN；JSON空白解析及fixture旧政策/缺真实repoCode失败均保留。21既有源码/测试改＋4新B区文件；C区本地验证器/浏览器与必要docs。507原路径缺失0，没有历史删除/归档；正常构建bin/dist与既有测试临时生命周期另列，不说垃圾清理已完成。
- SSH/远端/DB占用核查/DDL/DML/注册题库/旧成绩或PDF迁移/发布0；独立CodeReviewer尚未执行（主安排），真实005注册/来源审核/实际环境门禁下一独立任务，旧混合发布继续挂起。完整[实施、文件行号、失败与机器证据](management-traits-005-isolation-local-20261008.md)。

## 2026-10-08 管理特质005独立编码：已确认项目决策（替代冲突旧政策）

- 用户明确采用方案并作为项目决策：00501基层员工新版、00502干部新版；后续新增005采用新版评分/客户模板/新版管理。00201/00202所有原功能及正常新建保持。此最新确认替代历史“新建002禁旧、强制新版草稿/新默认”政策；下方历史不删，政策尚未实施为代码调整。
- 已冻结新版002（含1791298091700970647、58.642639／两位58.64）按原冻结来源与显式源映射兼容保留：不改码、退legacy、重算、删PDF、批量UPDATE或重写hash；不整体回滚已需兼容的局部故障修复。旧报告重出仍只接受可信原来源，不伪造历史。
- 尚未执行的002混合staging发布计划在项目文档层面挂起，后续改为明确005及现场新002兼容scope；不是停止后台或取消活跃任务。本轮未查运行进程/远端状态、未kill/部署/SQL/SSH，既有已部署事实与失败门禁保留。
- 005题库/评分/内容/模板/route版本绑定须先核数据库占用；本轮仅473个本地Go/Vue/JS文件未见005编码，不能断言空闲。现有引擎、模板、renderer、report/reissue两表及draft/approval组件可后续评估复用，不当垃圾，不全局替换产品列/hash/source命名空间。正式内容/心理测量/production批准未扩大。
- C区[权威决策与限定清理盘点](management-traits-product-code-decision-20261008.md)：A已确认可删0；B构建/输出/历史辅助脚本均候选，原源码/SHA/proof/失败收据保留至稳定及保留审查；C客户原件、字体、源码/env、旧PDF/现场新002、远端备份及业务表本轮禁删。只读盘点native0，删除/移动/归档0；末尾“1”由主会一次澄清，不解释为删除批准。

## 2026-10-08T08:34Z MT-ADMIN-ENTRY-ISOLATED：单SFC源码与393候选已准备，完整闭包未PASS

- 仅已上线52eec identity candidate-ui基准单user/exam/index.vue隔离修改，406输入仅该SFC变化；215旧Go恢复清单逐SHA同、原工作树front src/public/internal及旧candidate-ui前后快照同。旧APIa90b1f67…/旧TEST结果eb978542…/router/login/已上线candidate d1afa…保持；template/style/public/config/lock无变，不带newreissue/draft/formal或backend/DDLs。sourcehash-input的goBaseline字段误用前端数，不作Go依据；准确215另在候选manifest。
- 严格同ID/legacy两轴/精确002/server booltrue＋fields→正常getInfo code200正安全数字ID1或正IDwildcard→原同exam frozenprofile→已有TEST结果页；strictfalse真实旧合同保旧/URL-meta-session无鉴别权，未知拒旧/retry，non002/00401原逻辑，seq初始0及cross/destroy屏障。仅隔离actual SFC新86专项，RED13pass73fail→GREEN86/0/native0，不冒当前167/full560。首85/1是跨realm空数组prototype工具断言，真实为空；失败保留。
- 新prodmode build0/393，Node16.20.2与原compiler版本同/原2warning；SFC47e0fcc5…、indexbad70356…、tar7631433bytes/c5916188…。被动runtime117动态chunk/172请求117JS55CSS/missing0/179gzip/1589factory2020instances，只库存存在性，不AST/context/依赖plan或真实browser。完整证明器两工具问题纠正后第三轮超120s仍无映射结果，CtrlC停止/native完成退出未采，三编辑停止，不复制/改弱PASS。准备helper reviewmode亦三编辑后仍失败（依赖/Git换行）；直接原始字节diff/资源打包native0，不称该mode可重放GREEN。
- [可复审源码、source_patch.diff、artifact-manifest与包](management-traits-entry-candidate-prepared-20261008.md)已返回，仅PREPARED_FOR_INDEPENDENT_REVIEW，完整闭包/独立CodeReviewer❌，未发布；主协调者安排review及完整候选证明。08:33:28Z公网200/16155bytes仍52eec，SSH/SQL/DDL/上传/deploy/restart/真人PDF/39production0；用户当前读恢复页不动，旧URL永久反馈仍开放，原staging批准/reissue双MySQL门禁保持。

## 2026-10-08T08:15Z MT-ADMIN-DIRECT：指定浏览器只读恢复，永久旧URL未修上线

- fresh原admin页getInfo200/code200/正数字idOne与wildcard；公开Detail200/code0/sameexam1791298091700970647/legacy两轴/00201/strictfrozen true/三个字段键，无lifecycle；profile200/code0/same/frozenAt。实际旧tester-list HTTP403/code1/响应键code-msg-success；compiled旧入口先URL/session意图helper而没有serverfrozen分支，正确旧保护不放宽。空表不是no-completed，新结果权限403未出现。
- 已仅手动导航原page专属management-traits-results/同exam，真实list200/code0/success/count1/completed1/140140/overall58.642639；点击只读详情200/同run-paper-exam/13dim4mod。旧TEST DOM总体原有327050/5577，独立两位58.64，**不冒UI显示58.64**。保留新route/details，报告generate-view-download/reissues请求0/续答0/截图PII0，SSH-SQL-DDL-deploy-restart-production0，不声称HTTP日志零写。
- 已有local入口09790e74…与实际bool-only永久SFC无需重复修；fresh只跑两专项167pass/0fail/0pending/native0，API5be56c保持，560/build仅既有证据不重冒。公网08:15:56index200/16155bytes/52eec完整SHA仍旧。current reissue服务实现/测试/schema已读未改，不重跑1062或盖用户改动。
- 永久旧URL自动分流仍未发布，手动恢复不关该反馈；统一reissue实库失败门禁保留。主独立CodeReviewer下一只old identity candidate-ui基线管理SFC→已有TEST结果页的最小授权/frozen切片及资源闭包，不带新报告UI/wrapper/draft/formal/DDL/backendrestart。当前无该最小发行包/reviewPASS，不打包整dist。[本轮安全证据及责任人](management-traits-reissue-staging-release-20261008.md#L3)。

## 2026-10-08T07:57Z MT-REISSUE-1062：LOCAL GREEN，实库复验SSH双timeout／未发布

- 仅reissue服务目标bug：原run普通读先于paper UPDATE建立RR view，报告Find仍consistent read；原report INSERT错误丢sentinel。typed MySQLError1062仅精确报告输入unique尾名触发，一次全ROLLBACK后fresh reuse-only事务；完整原来源/current bundle共享锁/同run-paper-exam归档字节/DataSHA/template/content/成功generate审计及PDFsize-SHA全match，再原子reuse审计。其他1062/audit1062/1213/字符串错误/取消/撤销/任何漂移失败关闭；无重render/sleep/锁或隔离级别变化、仅清自有loser key，不winner/旧PDF/current/result写。
- 新23分支sqlmock有效RED5pass19fail→专项111/0/2LOskip/native0；首全包120s限额134.755s/native1而0fail事件、timeoutstack未保留，不断定唯一cause。真实发现全部服务tests四批＋所有其他packages、子限额仍120s，6383pass/694top/0fail/11原skip/parse0/native0；Windowsserver/allbuild/vet0，最后精确隔离73/0/1skip及build/vet/Linux/externalcompile各0。非coverage/realMySQL，不重前端560/LO。
- 118隔离运行输入仅服务1变化，current/candidate精确SHA261b90bb…；新green overlay映射同当前tests3a40e200…，旧overlay测试copy/原manifest只是历史，不直接继续原publish。编辑器format与EOF造成rawSHA门禁两fail，规范化code同、最后exactPASS；不要拿formatted editor和native不同版本互盖。原ab07083e发行binary/DDLd5c5…保持，新Linux3b2b1290…仅验证、未更新install-package/发行verdict。ignored partial-source模块gopls有undefined诊断，原受控overlay实际build/vet0，不补全/修改go.mod绕它。
- 07:42:04＋07:43:43strictSSH连接前255timeout/stdout0/67bytes同SHA，唯一重试后停止。远端restore/GRANT/DROP/upload/mainDML-DDL/deploy/restart/production0，随机owned计划仅local，不称cleanup实证；07:28PID2746/4179/52eec/主两表0/58.64仍最后历史，未fresh核。原实库1062的**具体键尚未实采**，RR路径为代码/SQLmock证明，不冒真实时序闭环。
- 原external fixture增强有界双pool/RR/两个persistence readview屏障及安全exact-key类别，保留report1/audit4，补generate1/reuse1/same-data-file/files1/orphan0/source不变，已编译未远端执行。C区bug-only helper无publish，复用原protectedbackup/exact随机库finally，green绑定最终binary df0478c8…＋source/testSHA；新工具三编辑/服务测试三编辑/隔离三编辑停止。主先独立CodeReviewer再fresh恢复库RED键与GREEN复验，批准保留/发布仍NO-GO，详情[本轮发布报告](management-traits-reissue-staging-release-20261008.md#L3)。

## 2026-10-08T07:55Z 胜任力独立生产发布启动：当前备份与5.7结构演练完成，未切换

- 用户确认仅新版一期90题/v2报告、排除新版管理特质，旧454题/48D/9配置/19结果/3PDF保留并行；另明确批准当前受限备份/自有恢复库首重跑后DROP。生产发布批准保留，不等旧48D并行运行与独立包门禁已完成，不跑009/014/015/MNG、不重算或清旧卷。
- 首Posh整载荷>68536bytes/BeginExecute拒绝/remote0，模块等待CtrlC止；改24000字符分块/精确0600payload SHA后真实339f2046959a45d2备份成功。恢复后的beforefacts夹具SQL引号丢失1064/42000、12before/DDL0，独立fixedclassification实证；脚本3编辑停止，用户另准一次hex修正/一次演练，失败及备份保留不覆盖。
- 唯一新47dc18ffbb16498e：MySQL5.7现库备份→恢复自有库→007008010011012013首/重跑12次exit0，结构投影SHA两轮e2084a01aa5e2f22d74f3384d506af0a83b065e8a27dbcf9f56c1d2ef6cd6ff5同，14表旧列投影SHA同（007空scoring标签明确期望、非canonical全库证明），48维保持/MNG0。精确cleanup0，独立07:55:25两ownedschema残留0。
- 新current备份root0700/files0600，DB14864794/SHA19d30e3bb4c75a2a324e349240eecf409360a55ae33b617157d60c9a6ddcacd2，application500260278/SHA1a47de9329c6d1cb7dbd5b4d9c43cb579785cc250336a97b7d5215580fed3f78；manifest0，首failDBf65ab…同size不同SHA不猜原因。秘密归档仅远端受限不下载，payload永久留证不glob删。独立GTID_OFF/dump实例语句和存储定义固定匹配0；主54表/四product列0/48-454-19-3/PID1195022/NRestarts0/两runtimeSHA/health保持，主DDL/DML/replace/restart0，ownSSH0。
- [完整授权/失败/收据/限定结果](competency-production-release-preparation-20261008.md)。独立驱动审阅指出GTID泛化/crossschema解析/完整signature/timeout清理欠项，当前实际未超时及cleanup0不抹风险；首次重跑通过不等全迁移/运行PASS。code/front独立closure、新旧身份题本并行适配、production精确内容批准、writer/故障回滚/2旧expiry策略仍欠，整体NO-GO未上线。仅C区2新驱动/docs，无B区/原SQL改，脚本预算用完不自行继续变更。

## 2026-10-08T07:28Z reissue发布失联恢复：真实演练失败，主库未安装／未切换

- **BLOCKED，未发布/未生成真人新报告**。用户交接历史DTO兼容独立CodeReviewer PASS/full560/build0为既有权威，不重新复审或重跑全量；但真实恢复库双连接门禁失败不能由本地PASS替代。07:22:40演练native1，07:23:27安全诊断：err0=false/err1=true、mysql_1062、reports1/audits1、未同report复用。未取得具体冲突键/完整根因，不猜测、不削弱断言继续发布；三次原演练失败全部保留，本轮不第四重放或改产品。
- fresh只读strict liming/knownpem/Batch/Conn10/Attempts1首次native0、stderr0；07:28:42Z PID2746，backend磁盘+/proc均4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c，front52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860；三active/healthok，两报告环境staging。全state1=0、overduecompetency=0、mainreissue两表0、指定completed1/58.642639（显示58.64）保持。
- exact owned恢复库/库授权0、payload/next/server pending路径均不存在、deploy receipt及main签名不存在；本机只有既存node9316/1764、未匹配reissue脚本，远端限定进程仅本次bash15915，未发现发布进程。不kill/广清/恢复主库，不冒全系统无后台活动。受限backup mng_reissue_76f547bc8ae0833b永久保留，ownership/四SHA/gzip/0700-0600门禁0，DBgzip773072035e5f422863ac4f21b151c86cca12075b2e09f406364ca90967355b26；旧rows/schema/score/profile/465PDF/private/config/cache与备份baseline同。
- [限定纠正] 下方“备份/真实MySQL两轮尚未开始”已过时：真实备份及三次恢复演练确已执行，后两首/重跑结构签名同1ebadfd6…，清库0/errors0；**演练失败、主发布仍未开始**。最新fixedUI实际消费3/3新输入/markers存在/393资源/index5b47f939…，原fixed模块收据仍BLOCKED/81未归因，closure/rehearsal/deployment verdict文件均不存在，不臆造packagegate.ready；本轮不重开review或强行写PASS。
- 新候选backendab07083ee2a972ae66343110bdab9b80310d77d934cd3e6eb3f1cd77cde25908仍local，current声明前端三源SHA匹配，不打包工作树draft/formal。此次SSH仅SELECT/摘要/文件状态；DDL/upload/restart/真人PDF/答案或旧卷写/production均0。页面新UI与生成查看下载SKIP，原因未部署，不称保存已验；nativeNode0与wrapper成功分开记录。恢复receipt/机器阶段见[本轮发布报告](management-traits-reissue-staging-release-20261008.md#L3-L17)。

## 2026-10-08T07:11Z 生产数据库变更与stage只读对比完成：general_ci窄边阻断已证

- 用户本轮聚焦DB变更/必要stage对比，仅只读采集与本地审查，不迁移/restart/backup写/恢复副本/主业务写。新安全Posh提示root生产会话0＋stage既有liming/key/strict；三批双端SQL均exit0/10秒SELECT/READ ONLY，secret只远端内存pipe，不保存人员答案。末生产07:10:56Z PID1195022/active/NRestarts0/healthok，自有连接关闭0。
- fresh生产5.7.44-log/54表614列188索引列20FK30关联列；stage8.0.46/78表909列334索引列49FK68关联列。stage额外24表、共享新增14列＝007八+008五+013一；formal002/draft003/reissue004两端均无。起末结构/聚合及rawSHA保持：production834378935768917bb2f50f999ce859bc6789d02015cc34c1d1d8d2892bcca461/stageafd19a7020c8772d2cc25783fc32e7a6af5c126a83c388e5492f78ec6720c7bd，不冒全库字节未变。
- [补充纠正] 旧7月“生产胜任力源题0”不是当前：实查454源题/48D/9配置/19结果/3报告；stage90题/10AB/24结果/28报告，禁止同步或跑009重置、015stage批准。生产源ID全部ASCII<=32、MNG三父DDL条件3，但当前management_traits_schema.go窄边例外明确仅0900，生产general_ci的64→32不支持；仅建001不足runtime可用。原生既有Drift/unverified_base/general_ci测试native0/2pass父子/0fail0skip/parse0，仅本地拒绝合同，不远端runtime验收；editor未发现保留。
- 主业务孤儿聚合：生产result/report缺paper0、snapshot缺exam0；state0缺exam199仍保留、2到期competency未交。生产MBTI两字段较stage宽，超stage长度计数均0但不授权缩窄。stage双列维度item重复8由新题型索引区分，不当数据错误。四production独有优化索引保留，QuartzRESTRICT/NOACTION不迁移、0900不拷5.7。
- 007/008含回填/索引切换/多DDL且008/011重跑会DROP自身FK后重建，不能假全事务回滚或无写no-op；014开头无PK DELETE需单独治理，当前不修/执行。stage8不能替5.7恢复副本首/重跑/故障/锁/运行writer验收。完整[审查/逐迁移适用性与证据](database-change-production-stage-20261008.md)，仅C区新2只读SQL/PS collector/JS comparer/docs；parse/nodecheck/实际比对0、diagnostics0，Go业务/原SQL0改，NO-GO待scope/兼容方案/当前backup/独立恢复授权。

## 2026-10-08 MT-ADMIN-STAGING-BOOL 旧DTO兼容本地GREEN（待独立复审，未发布）

- [限定纠正] 下条真实historical strictfalse/无lifecycle-newflag RED已本地解除；完整新schema lifecycle不是已发布旧bool合同的必需。仅管理入口三行：合法ID同exam＋完整legacy两轴＋实际002＋strictbooleanfalse，允许新标记缺失或一致legacy/false；任意存在unknown/null/undefined/畸形/矛盾拒绝，不unknown默认false。stricttrue无lifecycle的字段/profile/fresh权限原门禁、合法draft编辑、非002和00401原分流保持。
- 实际两DTO路由投影永久SFC回归：historical旧list1/profile0/newUI0，即使URL mngTest也不覆server；指定new stricttrue新页1/旧list0/profile1，仅本地mock不是线上认证。有效RED114/109pass5fail/native1→专项167/0/0/native0→full33files560/0/0/native0（+35，原39保留），prodmode build0/原2体积warning/两diagnostics0。首相对路径ENOENT工具失败独立保留，测试三编辑/产品一编辑。
- 原生终验internal Go实际222/聚合ffe5ab6a…前后同、API5be56c…及候选ab07083e…保持，不重build Go或跑6272/6359；不误报236内部数。只一SFC产品＋既有测试和必要docs，依赖/客户资产/字体不变；本地dist非最小发行包。
- 独立CodeReviewer及当前发行闭包由主安排，真实MySQL两表/双连接、真实browser/真人PDF未验；本轮SSH/HTTP/SQL/DDL/部署/restart/真人和旧PDF操作0。不自行宣reviewPASS或继续发布；原统一staging授权保留。[完整政策/失败/数值退出/源与产物保护](management-traits-reissue-staging-release-20261008.md#L3-L15)。

## 2026-10-08T04:35Z reissue统一staging接续：隔离后端GREEN／真实旧DTO入口RED，未发布

- 用户明确统一两独立表＋五API＋admin分流staging20.200.136.133及随后指定真实完成结果TEST报告生成/查看下载；批准保留，不draft/formal/production/旧PDF覆盖。freshstrictSSH首次255timeout、唯一重试0；PID2746/磁盘进程4179…/front52eec…/三activehealthok，全state1=0、旧11表且reissue/draft/formal0，指定examcompleted1/58.642639。
- 从已发布identity candidate-overlay建立118运行源隔离闭包；新报告model/schema/service/handler、audited纯依赖/文案adapter及五路由追加，不复用current draft loader/guard。专项两次49pass0fail1明确LOskip/native0；首次全build因bin overlay输入被扫描为独立包exit1，仅工具复制原go.mod/go.sum形成嵌套模块边界后allbuild/vet/Linux各0/输出0。候选49991517bytes/SHAab07083ee2a972ae66343110bdab9b80310d77d934cd3e6eb3f1cd77cde25908仅local。
- [限定纠正] 下方“false历史保旧/按actual DTO补labels”未覆盖已发布旧bool合同：真实公开Detail两exam各200code0；历史177682…strictfalse/legacy两轴且无lifecycle/newflag，当前e925…SFC默认created旧list0/entryBlockedtrue，REDnative1；指定179129…stricttrue本地adapter分流PASS，非真实admin/browser授权验收。原Detail确无两个新字段，不能以补字段夹具冒线上合同或夹带draft后端发布。
- 业务源码未修；统一发布在该兼容门禁停止。远端backup写/upload/DDL/replace/restart/真人生成下载/production0；主库未安装两表，旧历史不重发。CodeReviewer及新前端发行闭包待主协调者；不要重复索取已批准scope，也不自行改PASS或部署整个工作树。[完整报告/失败/源码SHA/后续责任](management-traits-reissue-staging-release-20261008.md)。

## 2026-10-08 MT-ADMIN-REVIEW-3：三项UI finding源码/单测GREEN，浏览器停止未PASS

- [限定纠正] 下方“unknown/矛盾/迟到全部收口”仅旧测试覆盖，CodeReviewer三项confirmed本轮有效SFC RED27fail/100pass/native1。仅当前两个管理SFC：same-exam完整assessment/scoring+实际repo classifier再分流、002明确legacy false、不强制非002新字段；报告每row reference/requestSeq/promise/epoch、postgenerate force、精准生成ID/失败保已生成可retry；旧list query快照与scope/seq及destroy守卫。
- 首五相关199pass0；两旧正向夹具按actual DTO补scoringMode/legacy bool与labels，不删负向。最终33files525pass/0fail/native0（486+39）/fresh frontendbuild0/诊断0。原40新SFC保留、现79/API53不改；222Go聚合a2c6b44f…前后相同、API5be56c…保持，Go6359/backend/DB/remote未跑。
- 浏览器首1440在legacy route双探测repeat-submit失败/native1；actual AppMain route.path keyed，组件watcher仅nextTick后跳过inactive/destroyed，公共guard不改。driver三编辑后最终120000ms boundedstop/native null/SIGTERM/parent1/stdoutstderr0，无summary；1440截图/合成下载仅部分产物、390未证，不能沿旧两case PASS。终验ownedNode/Chrome0，旧Node9316/1764保留；不第四改/增预算，独立review须主协调者。本轮只source/unit/build ready，不release ready；[完整失败/调用影响/前后SHA](management-traits-admin-ui-local-20261008.md)。

## 2026-10-08 MT-ADMIN-UI 本地管理入口与客户模板报告操作完成（未部署）

- 仅两个管理Vue＋API及对应测试/docs/C区本地mock测试；旧URL按当前exam服务器strict frozen true→fresh getInfo数字正ID1或正ID全局权限→profile同exam进入专属页。false历史保旧/draft原编辑/unknown字段或失败不旧fallback，URL/session不解锁，迟到/销毁scope保护。结果表原Element UI风格/20条本地分页、0与NULL/精确HALF_UP、13维4模块、独立行报告三动作，无手填ID。
- 实际后端list只有examId与200上限/run事实，无姓名手机号/逐题答案/服务端查询分页；页面明确未提供，不造参数/API或以ID冒姓名。报告新五reissue接口lazy按行资格/meta，nested report/reused/purpose、明确revision选择/生成后该行刷新；404409503只报告服务不可用，结果不清，旧TEST/current/PDF不回退覆盖。仍TEST/不可人才决策，未正式批准不叫正式。
- 有效新SFC/API RED31fail50pass/exit1；全量resume回归1fail与两个本地browserTimeout保留。真实bundle证实旧入口与结果页只读Detail POST在1000ms触公共repeat-submit；用户额外1轮批准后只切已有fetchManagementTraitsExamConfig隔离client，不清全局session缓存/改公共拦截器/加sleep。后续只读POST消费者须核对这条真实拦截边界。
- 最终全前端33files486pass/0fail（435→+51）/生产buildnative0/诊断0；实际当前bundle mock1440/390两case：旧直达自动分流、13+4、生成/Blob查看/native合成下载、503结果保留、scroll=client1440/390、pageerror0/forbidden0/closedtrue/native0。仅合成API/PDF流程，不是实际服务器报告/PDF正文或DB验收。internal222Go前后聚合SHAffe5ab6adc89a43f04e81ad077068339799a2a9f9efa4239ca582547890a9870同；formatter/reissue/formal/draft不动，不重跑6359。
- 本轮SSH/SQL/DDL/部署/restart/生产访问/真人身份答案/PDF/客户资产/候选原tab/Git写0。当前dist是整个未部署工作树，不最小staging发行包；线上指定URL未切。独立CodeReviewer须主协调者调用，后续统一新API/两表/backend/front发布另批准。[完整影响、失败和本地证据](management-traits-admin-ui-local-20261008.md)。

## 2026-10-08T02:57Z 39生产Posh-SSH只读续检完成：仍NO-GO

- [限定纠正] 02:41默认密钥255是历史；用户选择Posh、安全Get-Credential后strict会话0/root/iZ0yosjdcen2p4Z连接成功，仅只读现状，不部署/restart/迁移/新备份/业务清理/权限修改；结束关闭自有连接、凭据不保存。首多行send扁平化本地ParserError/remote0，单行后成功。
- 实际/proc及磁盘backend SHA03397e0faf24a21fb6da4e76ba0776226ab87bf2a3c52f72b475eeed4791e44a、index f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136仍7月发行；PID1195022/root/active/NRestarts0、激活7月27日15:58:34CST，终验02:56:45Z保持/内外healthok。根盘74%/可用10501028KiB，内存7658/可用5698MiB；3306/8092全地址监听不等公网可达，TLS/限源未验。配置0644/backup父0755仅记录不chmod。
- 默认mysql socketexit1后仅远端内存DSN→匿名管道只读成功、不印secret。实际MySQL5.7.44-log/element54表；66exam=57legacy+9competency，1980paper=state0 517/state1 0/state2 1463。旧进行中state0不能混新MNGstate1；517全部到期/最近24h新增0，316legacy+2competency关联现存exam，独立LEFT JOIN确认199缺exam/原因未查不删除。Worker新版首扫2旧competency策略未定，不把0未到期当维护许可。
- 当前8旧competency表/48维度/19结果/3报告/20FK总数，四exam版本列0/MNG表0，不整包升级或跑staging009重置/批准SQL。配置凭据全局28权限含SELECT/CREATE/ALTER各1，无写探针。限定depth2仅7月SQLgzip，SHA283d227047aef5fefaa315f628c1da59ca13b67583f56806579a73528ebee08d/gzip/整manifest/两distgzip-tar均0，子0700/files0600；旧backup非当前恢复点、不称新恢复演练。
- [完整现状/失败/门禁](production-preflight-20261008.md)仅C区更新；scope/当前backup与恢复/5.7兼容/旧卷策略/正式功能验收仍待，NO-GO。30min指定journalcritical0非全业务无错，SQL只读COMMIT非业务DML；不沿检查授权创建备份、上线或清199卷。

## 2026-10-08T02:56Z 指定管理员页真实只读复核：认证有效／旧入口未分流，完成结果存在

- 完整读取当前磁盘记忆实际1873行（不是交接所述3776行）及reissue本地合同；原生SHA核对三前端文件与本次读取一致。只复用指定admin页81675ce1…，page/context mocks各清一次并刷新一次；candidate页不操作、不复制或落盘凭据。
- 02:55:33Z正常getInfo HTTP200/code200，admin角色、userId1、wildcard均true；刷新后store管理员/wildcard仍true。公开Detail HTTP200/code0、精确exam1791298091700970647、repo00201、实际isOpen数值1、managementTraitsProfileFrozen严格true；正常管理员profile/detail HTTP200/code0、同exam及frozenAt有效。不能沿上轮401归因本轮空表。
- 当前DOM仍ListExamUser传统“测评详情”、0行，未到ManagementTraitsResults。刷新resource timing实采旧tester/tester-list HTTP403；profile请求不在自动入口链，后续仅人工只读核验。页面compiled入口函数实读含已知意图/管理员双门禁再profile分流，没有冻结metadata直达分支；当前route无mngTest query/meta且此exam session已知标记false。本地同入口也仍依managementTraitsExamKnown，另有未部署draft分支；不能以minified函数缺原符号名认定功能缺失，不称完整线上源码等价。
- 02:56:32Z正常只读POST专属results/list HTTP200/code0、精确scope共1条/completed1/140完整1；仅该行GET results/detail HTTP200/code0，run/paper/exam绑定一致、140/140、13维4模块、总体显示58.64/qualified。原“暂无数据”是旧列表读取失败，不能作为no-results证据；不放宽正确旧保护/角色门禁。
- 当前本地结果组件及API仍旧TEST生成＋手输已知report ID，无report-reissues消费；新五API/两表未迁移未部署沿本地合同记录，本轮不探测新接口、不读报告PDF或触发查看/下载审计。报告是否已有及数量本轮UNVERIFIED。刷新有404控制台事件但未归属具体资源，不冒业务404；response listener未采到事件，HTTP刷新证据来自resource timing及必要fetch实际响应。
- 本轮仅本记忆追加；业务/UI/脚本/配置0修改，SSH/SQL/部署/restart/报告生成查看下载/人员答案写0，不声称HTTP访问日志零写或DB零增量实证。页面保留指定旧URL；未运行编译/测试（无代码变更）。下一仅待明确授权管理员入口最小分流修复及独立reissue UI/发布范围，不沿“已登录”扩大为实现或部署批准。

## 2026-10-08T02:51Z 指定管理页只读检查：会话401／旧列表403，结果未验

- 复用用户原管理员tab，清page/context既存route mocks后，02:50:56Z正常GET getInfo实际HTTP401/code401；首页和本地权限缓存不当有效管理员认证。随后仅按用户明确指定URL导航，停留exam/users/1791298091700970647/1指定管理页；候选tab不操作、凭据不导出、不重新登录或伪造会话。
- 本次导航的资源元数据：公开exam/detail200、旧tester/tester-list403；DOM仍传统“测评详情”，行数0/“暂无数据”，非management-traits-results。403产生未捕获Axios页面错误。不能把空表解释为无人完成，不能据本轮401确定403唯一根因；未请求专属results/detail/qualification或生成、查看/下载PDF。
- 当前磁盘原生只读核对：user/exam组件SHA6373a0f0…，分流依managementTraitsExamKnown及本地管理员判定后读profile；002 frozen投影没有独立直达分支（draft分支另有）。结果组件SHAeb978542…仍Run ID表格、手输已知report ID；API SHAa90b1f67…仅旧reports/generate-test与view/download，无report-reissues消费。上述为当前本地源码事实，不冒当前线上完整源码等价证明或已验专属页。
- 本轮仅本记忆追加，业务/UI/SQL/配置0修改；不SSH/DDL/DML/报告生成/人员答案操作/部署/restart/生产探测，不声称已核DB零增量。编译测试未重跑；下一先由用户正常登录恢复真实getInfo200，再只读检查专属结果、13维4模块及报告入口；本轮不改UI、不扩大旧身份发布授权。

## 2026-10-08T02:41Z 39生产只读预检：公网部分完成／SSH认证阻断

- 用户确认39.106.61.48为production、选择先只查现状/发布范围未定；20.200.136.133仍staging。fresh80首页200但title“没有找到站点”/health404；8090首页200/15799bytes/SHA f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136，health200/statusok；HTTPS一次5秒连接timeout/curl28。首页字节同7月记录，不推完整前端/后端未变。
- knownhost/root/default既有认证唯一SSH255/authentication_denied/stdout0，无可复用Posh会话或当前安全认证引用；用户说明此前提供过，本次限定查找未恢复认证，不猜口令/轮试密钥/恢复旧秘密。主机/backend/DB/活跃卷/备份回滚UNVERIFIED。
- Git变更360，未定候选包/不build或测试、不整包发布；最新reissue本地完成仍未部署、formal准入未关闭，旧staging身份批准不扩大。C区[检查报告](production-preflight-20261008.md)及本记忆，PARTIAL/暂不放行；remote命令/上传/重启/DDL/DML0，HTTP正常访问日志不冒绝对零系统写。

## 2026-10-08 MT-REISSUE-API：真实来源核验与本地后端完成（未部署）

- 用户明确只1/2，不UI/draft/formal/Worker/full560/自动staging或生产；已有外围代码保留。新独立报告+审计两表DDL仅C区工件未执行，原旧表/current/pdf_path/结果零修改，远端只有READ ONLY SQL/有界已知归档计数。
- [纠正 - 2026-10-08] 用户newexam1791298091700970647现在真实state2/completed1、new_creation/submitted_snapshot1/140raw/13dim4mod1receipt、唯一同paper/exam/participant/提交时间candidate1；不沿身份保存历史猜未完成。原manifest/mapping/evidence实际字节SHA匹配。独立Rat比13/4存储成绩全同，总体58.642639→58.64。原身份JSON未下载，source仅匿名PK域hash/必要分值与元数据；真人新HTTP资格/生成未执行，不说原实名报告已出。
- 旧exam1776822816300709851仍两完整140卷+一state0，snapshot/run0；真实旧字段无题干/V/当时身份快照。已知backups maxdepth2中2026-04-22之前SQLgzip0，非全外部归档穷尽。两旧卷不能重发，拒当前源库补证/伪snapshot；Ed25519审计签名只证明原bytes签署，原adapter不改。
- 新五管理员API沿原JWT组，仅runId/paperId/reportId，strict未知/重复/null/空串/path/approval拒，正ID1或正wildcard。可信冻结loader复用原140/13/4/receipt/owner/时间及bundle共享锁→205纯核心→002Word/LO；原snapshotJSON字符串与raw/selected/final归档，不新签钥或正式审批。独立UUID私有reissues/0600，同输入unique幂等、事务外render、二次paper锁复读及metadata/audit同TX；失败只清自己孤儿，旧PDF/current保留。授权归档读取只验原archive/file、不重渲染或读当前人员；撤销禁newgenerate。
- 新表在旧容量guard默认会因未知ID列失败，真实专项RED已证；仅每操作schema通过后复制capacity扩八ID键，singleton不改。真实HTTP夹具另遗漏Take的绑定LIMIT1及Read末次EOF；三故障修正后实际200/PDF字节/RFC5987/安全头通过，原失败保留。
- 专项49pass/0fail/1PDF opt-in skip；匿名真实分值1pass＋实际LO1pass均native0。真实答案/合成身份TEST PDF650342bytes/A4九页/SHA8d3622647f612c08b4fb197960266fe13067cfa3043cafb29efe080b637877bc，36/36所选客户原文匹配，不冒原identity/mappingSHA或六图全像素/实名报告。默认全Go6359pass/693top/0fail/11明确环境skip/parse0/native0，27测试开关仅child清；server/allbuild/vet各0/stdoutstderr0、诊断0。
- MySQL两表首轮/重跑、真实双连接竞争、所有故障/符号链接组合、原名新HTTP200、独立CodeReviewer、staging/production/前端未验。新API仅local ready，不沿原4179/52eec身份发行自动部署。完整合同/影响/失败及作品见[本轮报告](management-traits-reissue-api-local-20261008.md)。

## 2026-10-08T01:43Z 最小身份 staging 已部署／单合成保存与清理 PASS

- [限定纠正] 10-07T15:56 SSH双timeout为历史；10-08strict首次0、fresh全state1/pendingcompetency0且切换前再次0。沿用户最终独立source/artifact CodeReviewer PASS/ALLOW发布，不重复review/proof/build、不打包整工作树，draft/formal/reissue/DDL/production0。
- 隔离backend49903310/SHA4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c在线/proc同；一次正常主重启PID2006→2746，nginx/mysql PID保持。前端393公网逐大小SHA同manifest，index52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860；旧8baa/353c备份保留。backend先可信Detail再frontatomic，未承诺零停机。
- 受限backup mng_identity_5d517cc02d6c8396 root0700/files0600，current完整DBgzipSHA e84f74c07e6093ac224f1d4d982dbeb3e6836b4c4f48bf895b52d9ef5c6abc15/application809bfa…/systemee6b…，gzip/tarcompare/SHA及metadata校验；未下载secrets，exactuploads0/rollback0。
- 用户exam1791298091700970647公开Detail200code0/sameid/stricttrue/namegendertelephone，新普通URL DOM三字段/保存可用；只读，不改原浏览器未保存输入、不提交真人。正常独立admin登录后仅单synthetic配置明确冻结，普通UI真实保存200code0，examId＋三身份字段/extra0；SQLcandidate1/身份及字段子集一致/no-paper-noPDF，非paper快照验收。
- 精确PK/FK-SafeUpdate清理native0/owned全部0，复用原bundle保留；恢复exam71/paper1488/candidate1349/profile1/bundle1，不清他人/不假11逐0。driver原terminalexit0，contexts关闭；最终三healthy/PID2746/旧465逐SHA/私有基线/源各140140700/cache/config/Schema/原用户整profile及contract051dbe…保持，11表15FK/draft0。publication＋singleidentityacceptancecompleted，完整draft/正式功能仍未部署。[完整限定收口](management-traits-candidate-identity-local-20261007.md#L3)、[最终收据](../scripts/test/results/mng-identity-publish-20261008/acceptance-verdict.json)。
- 新经验：PowerShell here-string管道node的stdin不是可继续接收terminal答案的交互通道；首loginhandoff未送达/业务0，direct-file Node启动后ask确认正常。异步原terminal的LASTEXITCODE不能从另一个sync terminal取；本轮回原terminal实采0，不把blank冒0。原local语法失败/publishstderr90及旧失败全保留，不改产品迁就工具。

## 2026-10-07T15:56Z 最小身份 staging 发布：SSH 双超时，远端执行0

- [限定纠正] 下方15:07/14:50门禁为历史；用户最新主上下文明确交接独立CodeReviewer最终PASS ready_source_and_artifact / ALLOW minimalstagingpublish（两源及闭包/393产物）。本worker按该交接执行、不冒自己独立review、不覆盖旧NOT_EXECUTED、不重卡raw393/old21。最小staging发布与0活跃一次restart批准保留。
- fresh strict liming/knownkey/Conn10首次＋唯一重试连接前timeout，native255/255、10030/10028ms、stdout0、各stderr67bytes/SHA dbc4779f…；父native1。远端shell未启动，当前PID/健康/活跃/待处理/配置与DB未核验，停止第三次SSH/HTTP造数。
- 本轮本地不可变产物native0：candidatebin49903310/SHA4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c；candidate-ui/dist-faithful393各大小SHA精确匹配、额外0；manifest1649a1c7ca4e166d290daa3a8e934d4704a3c1937e066f4fd17b8515c733974f、index52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860。未重建/打包currentworktree或带draft/formal/reissue。
- backup/upload/replace/restart/DDL/DML/production0；synthetic0/cleanup_required0非cleanupPASS。原用户页只读快照未刷新改填保存，真实public新投影/子集/保存200及SQL清理未验、反馈开放。网络恢复后沿原批准fresh预检续作，无需重复部署审批。[报告](management-traits-candidate-identity-local-20261007.md#L3)、[收据](../scripts/test/results/mng-identity-minimal-20261007/publish-ssh-blocked-20261007-155646.json)。

## 2026-10-07T15:07Z 最小身份资源归因：public七项已证／完整编译仍BLOCKED，未发布

- [纠正] 14:50“394src/public”实际为394src；恢复所用source-scope641项只有394src、0public，输入清单本身遗漏public，不是路径过滤不匹配。另一个旧build-inputs406＝394src＋6public＋6配置。隔离public缺失；5copy资源＋2HTMLgzip精确解释386→393，首页模板另缺。六public/六配置当前旧SHA全同，fresh远端七路径SHA全同；非字体/手工JS/客户上传，不copy私人数据。
- 两隔离真实npm build:prod --dest dist-faithful native0/各393/原2warning，Node16.20.2；baseline index4880…/candidate52ee…仍非旧353c…，347新/不同路径＋345旧路径缺。旧/基准各1589模块ID，仅374shared/1215替换，shared335raw不同；同ID不能冒同模块，重复428要保留。数量门禁关闭不等编译等价PASS；严格模块图/AST证明未完成。
- 新C区driver三编辑停止。原路径只读映射构建180000ms后null/SIGTERM/ETIMEDOUT/manifest0、原源不变；后续readonly服务load0/realpathSync.native undefined，但timeout唯一cause未证。不第四改/复制或增加budget。两源用户转交独立PASS、本轮SHA b052…/d1af…匹配记新receipt USER_SUPPLIED_PASS_HASH_MATCHED，旧NOT_EXECUTED不覆写，BreakGlass不派发。
- 15:07:28Z strictSSH首次0/stderr0：PID2006/backend8baa/index353c/393/三activehealthok/state1全0/overduecompetency0/draft表0。candidatebin4179…保持。backup写/上传/替换/restart/DDL/DML/production/用户页操作0；synthetic save200/SQL清理未执行，反馈开放。既有最小staging批准保留，下一需新有界驱动修正或独立模块图证明，不泛重新审批部署。[完整收口](management-traits-candidate-identity-local-20261007.md#L3)、[限定receipt](../scripts/test/results/mng-identity-minimal-20261007/artifact-bounded-verdict.json)。

## 2026-10-07T14:50Z MT-CANDIDATE最小隔离补丁：精确8baa基准已恢复，source-review ready／未发布

- 用户明确批准仅旧Detail可信冻结bool/字段＋candidate身份切片staging，安全0活跃时一次重启，不DDL/draft/formal/reissue/production；原授权保留，不再询问相同scope。BreakGlass禁止派发agent，独立review未执行。
- 215Go＋394前端src/public发布清单逐SHA恢复，17旧字节从对应代码文件VSCodeHistory精确匹配，19新增Go overlay排除；原工作树业务/测试/config零修改。Linux baseline49902830/SHA8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676与fresh线上**完整binary一致**，解除了精确后端源码缺口，不回滚旧21或拿GitHEAD冒源。
- 隔离仅Detail+strings import和草稿前candidate冻结/当前ID/子集/迟到屏障；API helper保持线上a90b1f67…，无lifecycle/isManagementTraits新依赖。有效GinRED8fail16pass/SFC35fail71pass→GREEN24pass/106pass/native0/0skip；23opt-in子进程清，buildall/vet/Linux三0/输出0，候选49903310/SHA4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c未上传。不复跑6272/435/560UI，不冒实库保存。
- **发布门禁仍欠**：独立CodeReviewer未执行；隔离UI两build0/原2warning，各386资源/sourceMaps0，baseline index659e…不等线上353c…，candidate e338…，两构建路径delta418；唯一src/public变化candidate已证，但Webpack上下文/模块资源差异未完成归因，不称393等价/onlycandidate编译包，不部署此dist。主先source review→原路径可复现构建或模块归因→fresh维护backup/一次重启/真实synthetic200验收，非重新授权。
- 14:44与14:50strictSSH首次各0：PID2006/backend8baa/front353c/三active保持、state1全0/draft表0；末次目标profile冻结1/三字段name,gender,telephone/healthok。只读字段名/布尔，不读人员PII；远端上传/备份写/替换/restart/DDL/DML/生产0，feedback未关闭。当前0活跃不替未来切换前复核。
- 工具C区scripts/tools，隔离源/产物B区ignored bin、收据C区results。重要新经验：确切发布SHA清单可与**仅相关源码URI**的VSCodeHistory字节配对，Go overlay以空路径排除新增未部署文件，并以原路径构建得到线上完全相同binary；不要只凭SHA清单不存在bytes就停止。前端隔离cwd构建不能凭源码相同宣称资源等价。本轮工具三编辑后停止追加，未派发/部署。完整交接见[身份修复报告最新段](management-traits-candidate-identity-local-20261007.md#L3)。

## 2026-10-07T14:30Z MT-CANDIDATE截图保存：fresh部署差异确认，未发布

- 用户收缩仅身份保存bug，报告/审批/draft/完整UI暂停。已有本地修复不重复编辑；fresh公开POST Detail200code0同ID/00201/name,gender,telephone，frozen/isManagementTraits/lifecycle三键缺失。SQL只读profile1/frozen1/同三字段合同SHA051dbe83…；draft表0。线上backend8baa7f87…/index353c9fb3…，本轮PID2006（不是旧17552），两次只读前后保持。
- 独立真实线上普通/my页modefalse/blockedfalse；合成值实际handleSave+确认，精确candidate/save在浏览器abort，捕获13键/额外8键，其中7空或null，与strict未知键/null固定错误合同吻合。姓名/电话/gender键真实，无sex/mobile aliases；remoteSaveSent=false，不称服务器保存200或线上已修，用户原页面/认证不动。
- fresh focused SFC/API109/0/exit0，全前端435/0/exit0（此前exit1未复现/原因未知）；Gin合同16＋配置真实create9事件均0fail/0skip/exit0，sqlmock非实库。production build0/原2体积warning；4源测试SHA保持，业务代码0改，不重跑Go全量/vet/560UI。
- 发布阻断：线上393资源/maps0，当前dist352同/2index差/39旧路径缺（hash命名差，非远端缺文件）；未证明onlycandidate。线上缺必需冻结投影，frontend-only不可用，当前candidate又依赖未部署draft三态；不能整包＋DDL凑合同。最小后续仅旧后端Detail既有IdentityScope/ProfileDetail冻结字段/strict bool投影＋对应candidate修复切片/已有API，不draft/formal/新建UI；须主确认后端最小范围及安全维护窗口，受控隔离构建/review后续。staging修复授权保留，backend/dist替换/restart/DDL/DML/生产0。[同一报告本轮限定事实](management-traits-candidate-identity-local-20261007.md#L3)。

## 2026-10-07 MT-MINIMAL-REISSUE：独立历史输入→客户Word→本地PDF完成（仅synthetic/TEST）

- [范围纠正] 用户最新收缩00201/00202客户V2.8报告路线：本轮只最小独立报告链，暂停draft/审批/Worker竞态/完整560UI外围推进；保留已有代码与历史政策记录，不删除/回滚，不沿此前staging授权部署。
- 新纯适配器接受原始历史审计JSON+detached Ed25519 attestation+应用可信根，独立legacy_verified_snapshot/audited_reissue；无DB/el_mng/profile迁入要求，不改原new_creation guard。签名仅证明审计签署字节，不替归档事实；未接真实审计根/历史loader/HTTP按钮，不冒真实历史资格。
- 复用原CalculateManagementTraits及抽出的客户205规则DTO核心/现00290Tag6图Word/通用LO，不00401评分或删旧PDFsave。原字节SHA/140raw方向final/13sum-count-Rat/norm/4模块Rat/705/13/身份原时间冻结为不暴露可变引用的独立快照；来源不足/签名未知/桶映射方向或身份时间不符拒绝。
- 实际最终synthetic00201 all50 PDF654393bytes/A4九页/LO26.2.5.2/SHA2aca74466b5ee018a78aa3e898591c21b9c1dae9139cf6a6a3a97e0028780f55；独立Fraction-XLSX-OPC-PDF140/40/13/4、36客户原文、5环13柱1常模折线PASS，75OPC中68非值部件保持；三客户原件与styles SHA保持，实看第三页/九页总览，不全几何或Word桌面验收。
- 首编译REDnative1→最终适配/身份248、Word11、既有报告消费方43、真实LO1均pass/0fail/0skip/native0；buildall/vet0/stdoutstderr0、诊断0。相关修改前基线10pass/明确LOskip1保留；未重跑全Go/前端，当前旧Frontendtaskexit1未调查，不沿6272/435冒本轮GREEN。
- 本轮remote/SQL/DDL/DML/历史真实PDF/前端修改/部署/restart/git操作0；既有B区仅报告DTO文件抽核心，新Go3，其他保护范围同SHA。两真实历史完成140卷仍缺原V映射/题干与人员完成时快照可信证明，10-06SQL仅历史事实；管理员单卷reissue是下一slice，不称全部报告替换完成。[实际作品/全部SHA/合同与未验](management-traits-minimal-reissue-local-20261007.md)。

## 2026-10-07 MT-NEW-DRAFT：独立持久化门禁本地GREEN（未部署，待独立review）

- [政策/技术纠正] 下条“草稿方案待协调者选择”由用户本轮明确独立sidecar＋保持两步确认替代；不是原子createfreeze。新002 server实际repo code决定draft，缺boolean默认新版/false400/mixed拒零DML；exam/link/draft同TX，取消freeze不回旧链。旧历史profile0不插标记、不关闭未答旧卷/不回填。
- 新el_mng_exam_draft 7列/PRIMARY exam_id/FK RESTRICT继承实际varchar64/collation；独立003 SQL仅C区artifact未执行。缺表只关闭新002创建、不startupfail/AutoMigrate；draft可合法002编辑及精确admin tester准备，participant登记/start拒绝；freeze真实source/profile/bundle与marker更新同TX，保留frozen记录。公开DTO strict isManagementTraits/lifecycle/frozen；当前ID、未知关闭，历史明确legacy/false原路径。
- 有效Save RED7fail事件、SFC3fail、精确GET列表RED2事件→GREEN；最终Go6272pass/680top/0fail/9原skip/parse0/native0，front32files435pass/native0，server/allbuild/vet/Linux四native0/stdoutstderr0，npm build:prod0/原2体积warning/18代码diagnostics0。Gin/sqlmock/真实编译SFC，不真实MySQL/完整浏览器/新DDL/远端验收；23optin只child清。
- 562源码基线19既有变化/543同，6Go＋1SQL新，formal既有源0变、旧21不rollback；本轮SSH/远端HTTP/SQL/DML/DDL/部署/restart/生产0。formal报告与两端后续/mainrace未完成。独立CodeReviewer须主调用；新增schema安装须另明确staging migration＋统一后前端发行许可，维护预检/backup/恢复DDL/rollback先行，不沿旧授权自动安装。
- [完整本地合同/消费方/失败/验证](management-traits-new-draft-local-20261007.md)。本轮只ready for review，不宣称staging用户反馈已关闭。

## 2026-10-07 MT-NEW-ONLY：新政策已确认，草稿持久化门禁待定（未实施/未发布）

- 用户本轮明确批准“先发布staging修复并验收”，20.200.136.133继续staging、不转production；发布范围含既有candidate修复及禁止新建旧版002。旧结果/PDF保留，既有未完旧卷仅核查、不自动关闭；不批准正式内容/正式报告启用或formal registry后续发布。
- [政策纠正 - 2026-10-07] 下方10-06“默认新版可手动取消”的新建政策被本轮“禁止新建旧版00201/00202”替代；原历史事实和收据保留，历史编辑不自动转换。
- 实读Save仍仅写legacy+legacy exam/repo；新建非competency publish_status=1，未接收或持久化managementTraitsTestOnly。前端Save后另行freeze，取消保留无profile配置；IdentityScope依数据库profile判新旧。仅请求true+UI禁用不能阻止未冻结配置走旧链，不能宣称完整server enforce。
- 当前Exam没有独立002草稿模式字段；现有profile严格要求冻结时间/映射/合同完整，不能把空草稿profile塞入现模型绕验证。仍需协调者明确持久化草稿模式方案，或确认改变保存/冻结协议后再TDD；不擅自复用competency版本列/publish_status含义、自动冻结或新增DDL。
- 本轮只读定位及两既有C区文档更新，业务/测试/SQL/环境0改；未运行测试/编译/浏览器/SSH/DB/部署/restart。已有candidate审阅PASS按用户交接记录，不当本轮独立复审；新门禁尚未实现。后续新门禁须独立review，发布须fresh活跃答卷安全预检；发现nonowned活跃/待处理卷则阻断并申请窗口，不自动结束旧卷。

## 2026-10-07 MT-CANDIDATE-FROZEN：增量失败关闭本地GREEN，独立复审未执行

- 实际读取普通入口源码/旧verdict；原new config只核字段串、不核server frozen/ID，queryless未知标记可回旧链。原生SFC12fail82pass/exit1，新增unknown8与URL/false跨scope4有效RED；仅candidate与既有SFC测试增量修改，不放宽backend白名单。
- 新解锁必须当前ID（字符串/正安全整数）+strict bool true+原字段非空唯一白名单；旧002必须当前ID+明确false。服务器repo002也施加检查，不能伪造URL001绕行；第二次读取未知/false不legacy回退，原异步屏障保持。
- 最终身份SFC/API106pass native0、全前端32files430pass0fail、最终前端build native0/原2体积warning、诊断0；非真实browser/DB/remote。BreakGlass禁止派agent，独立CodeReviewer仍欠；不formal后续/历史治理/SSH/SQL/部署/restart/环境门禁变更，真实身份/答案/PDF零操作。
- [纠正 - 2026-10-07] 下条“配置畸形关闭”未覆盖frozen/ID，本次仅补此边界。工具read_file/grep曾返回旧编辑器内容，原生磁盘另有14项RED；首patch覆盖后立即恢复全部14项/补实际ID布尔夹具并验证106pass，不以旧80或旧全Go数冒本轮。完整限定证据见[同日报告](management-traits-candidate-identity-local-20261007.md)。

## 2026-10-07 MT-CANDIDATE-FIELDS：开放普通入口身份子集本地GREEN（未发布）

- 用户明确先正式功能、暂不部署生产；旧版新建入口后续停用/历史只读。本任务只修身份登记，不继续formal slice/旧入口治理/harness，不改.env门禁、真实配置/身份/答案或旧PDF，remote写/部署/restart0。
- strict SSH首次exit0及公开Detail只读HTTP200code0：指定exam1791298091700970647真实配置与冻结合同均name,gender,telephone，profile1/frozen1；截图两字段与当前数据差异未归因，不能假去gender或补字段。不读取/输出实名手机号，合同hash见独立报告。
- 已证代码/合成链：普通queryless入口缺server新profile标记→旧固定rules＋全candidateForm→strict decoder拒未知键/null。线上当前Detail标记不存在实证，但当时失败body/路由未抓包，不推gender唯一cause。仅Detail新增完整验证后冻结字段/布尔投影＋候选组件server分流/ready失败关闭/迟到响应屏障，复用已有registerManagementTraitsCandidate白名单，strict拒绝不改。
- 有效SFC RED7fail2pass→最终身份SFC/API80pass native0；全前端一次32files404pass（工具无numericexit）；Gin真实Detail/Save与配置gender拒绝/两字段成功均local sqlmock。另lateidentity RED1→GREEN、repo分类失败RED2事件→最小Error传播GREEN。最终全Go6247pass/675顶层0fail9原skip/parse0/native0，相对6229增18/2；buildserver/buildall/vet三native0/输出0，frontendbuild0，diagnostics0。不是coverage/真实DB/远端修复验收，原53formal专项不改。
- 3866→3867SHA：两生产文件/三既有测试变＋新Go测试，原scripts/收据0变，formal/guard/router0变，Legacy操作0；已有修改不rollback。测试分支与反馈/回归/覆盖历史已更新，只标本地GREEN。独立CodeReviewer未执行，BreakGlass不再派agent，须主协调者安排；旧入口停用/formal后续未开始。[完整影响/限定结果](management-traits-candidate-identity-local-20261007.md)。

## 2026-10-06 MT-FORMAL slice1：业务三政策确证、本地版本管理后端GREEN（整体PARTIAL）

- 用户确证Q1=A同一授权系统账号分别内容/psychometrics两独立签署；Q2=B精确获批完整新版TEST run可追加独立formalPDF、不覆TEST原current/PDF、不转换legacy；Q3=B撤销禁newgenerate/activate、授权admin可读历史formal原PDF并显著revoked保audit。选择政策不批准205规则或模板，不复制Liming/2026-10-01候选批准。
- 本地新增version/approval/audit模型、Register/List/Change真实行锁事务及五JWT管理员API；独立三表SQL仅C区未执行、原001/11表不改。旧scope只把完整严格schema/FK/孤儿校验通过的三配置表分类排除，未知extra仍拒绝。独立MNG_FORMAL_REGISTRY_ENV只local、MNG_FORMAL_ASSET_DIR受控绝对目录，构造读取不写环境；正userId1或正userId且*:*:*，普通exam:list不授签。
- version identity含精确source bundle四轴/manifest/mapping/两资产/绑定SHA，不含人员/profile时限/run；审批服务器actor/time、两职责允许同人、重复保原证据零写、不自动activate。state/epoch限定更新不覆盖create_time，audit同事务失败rollback；revoke终态/epoch推进。canGenerate恒false，generationBlockedReason=formal_report_pipeline_not_installed；versionReady不是完整run许可。Q2 run资格及formalPDF生成/下载、两端UI、全局current选择、实际MySQL/竞争/独立CodeReviewer未完成，禁止部署。
- 新专项53pass/14顶层0fail/native0，实际公开Register/双签/activate/revoke+真实受控文件+sqlmock完整gate/GinHTTP401403400503404；全量6229pass/673顶层/0fail/9原skip/parse0/exit0，server/all-build/vet三native0/stdoutstderr0，新Go诊断0。child-only清23实库开关；无SSH/browser/remoteDDL/historywrite/PDF/前端测试，不把sqlmock当真实DB或审批。
- [纠正 - 2026-10-06] 原候选lo-compatible“零外链”仅OPC/chart范围不完整：实际正文16个LINK Excel.SheetMacroEnabled.12仍引用外部工作簿；新正式门禁真实拒绝，候选SHA未改。不去TEST警示/删LINK绕门禁，登记缺template只draft不可批准；补资产须新revision。生产校验source wire复用严格解码时每字段必有JSON tag，公开服务测试有效发现后修正。
- 原上阶段仅文档/三问题未答记录保留，已由本次policy+本地code授权限定取代。完整文件/API/CRUDE/C4/失败/验证：[slice1证据](management-traits-formal-registry-local-20261006.md)。本模式禁止再派agent，独立复审须主协调者，下一报告/两端UI未开始，不继续harness。
- 最终scope native0：215既有源仅router/guard变化、其余213逐SHA同；Legacy目录不存在/操作0，前端/harness未改、旧21不rollback；50doclinks有效，[安全收据](../scripts/test/results/mng-formal-registry-local-20261006-6591add72de5/verdict.json)。Windows环境变量单值装大JSON超过长度限制首次exit1，改stdin传输exit0，不凭外层PS成功冒GREEN。未执行任何DDL。

## 2026-10-06 MT-FORMAL todo1：正式报告/两端UI契约草案（等待业务选择）

- 用户最新明确选择“实现正式报告功能，并改造两端UI”，不继续harness。本阶段仅C区[实施契约草案](management-traits-formal-report-implementation.md)与[24组待实现分支](business-branches.md)及本记忆；业务代码/运行资产/客户材料/SQL脚本0改，无DB执行/远端/浏览器/部署/重启/历史迁移。契约尚非定稿、正式功能未实现，不沿旧TEST PASS宣称交付。
- 当前源码已核：报告TEST DTO/SHA/用途标题-label硬约束，HTTP MNG_TEST_REPORT_ENV只local/staging；revision UNIQUE(run_id,revision)、current PRIMARY(paper_id)没有用途维度。推荐新增独立正式版本登记/审批/启用和正式revision/current/audit职责，复用合格评分事实但不写TEST current、不ALTER001/旧11表，不靠删TEST警示或伪环境开放正式。
- 新el_mng_表会被旧守卫枚举；配置审批表没有paper身份列不能冒canonical report audit，必须先独立schema门禁及已知配置表分类、正式资源复合闭包，未知extra仍拒绝。当前runtimeAdmin只userId1/*:*:*，permission参数非真实细粒度授权；认证purpose实为management_traits_participant/paper，本轮未找到management_traits_test常量，报告用途与token用途分开。
- 一次性三项业务选择交主协调者：Q1系统内双职责审批允许同人/必须不同人/外部具名签署证据；Q2仅新正式测评或允许精确获批的既有new_creation完整TEST run追加独立正式PDF；Q3撤销后禁普通查看下载或允许授权管理员历史原PDF并显著撤销状态。此前Liming/2026-10-01仅候选测试，不复制为正式批准，不seed批准或替用户选不可逆schema。
- 范围仅00201/00202新建默认新版，历史old results/PDF保留，00401/MBTI不动；production保持关闭。必要业务答复和契约经主判断后再一逻辑切片RED→代码→GREEN/编译真实调用→回写，未经内容批准不启用正式报告。业务定稿前不生成可执行plan/tasks/checkpoints，不开展后续实现。

## 2026-10-06T14:29Z 冻结测试双模态竞态RED→GREEN／单stagingfreezePASS

- **整体仍NO-GO**：同批四完整UI/自然0-3-140/20min/续答5min401/expired409/四native-Python未完，mainraceUNPROVEN。源码确认正式报告仅TEST/local-staging，production拒绝，正式方案未确认/未实施是交付P0；新[生产准入/回滚清单](management-traits-production-readiness-20261006.md)不授权上线/生产探测。
- 新有界3轮取证仅两个既有C区测试；即时localSFC通过但远端f768在create0资源Timeout、6a5e在两confirm已passed后响应Timeout。新合成250ms save延迟实际SFC有效RED1：两次标题都是“提示”，真正“确认冻结 TEST”未点且仍显示，证明测试竞态，非产品APIbug；旧13ee/9b唯一cause不回推。插入$&替换SyntaxError保留/非有效行为RED。
- 到3次先停，用户结构化**另批准仅1次driver修正/短复验**：同exam保存response屏障＋精确冻结title；合同未第四改/不复制。原延迟夹具GREEN0，仍正常UI双确认、原120000/1500预算/guard/产品源码不变。JS替换携源正则$&时用callback，不能replacementstring否则展开match。
- cc784正常独立auth14:28:47HTTP200/admin，freshbaseline0；仅一00202candidate草稿/default-clear-reselect/保存取消/重新保存/独立freezeHTTP200；独立SQLfrozen1/minutes25/paper0/mappingf0f7387d…/manifest4a033e1b…/bundle candidate-current-source。无person/paper/answers/native；不four全验。三个远端短批各finally/final0、context全关，原集成route未动/其auth不fresh核。
- 14:29:13freshfinal0/11逐0/private0/15FK/旧465-source-config-runtime-metadata-schema-cache同/PID17552/backend8baa/front353c/三health保持；受限新backupSHA80f10898…保留，无deploy/restart/sharedcfg/prod/history写。相关4SFC202native0、641源逐SHA同/Go215同；只测试修正无需productbuild，不重跑6176/Python。有效RED与原fail保留。[完整本轮事实](management-traits-default-staging-result-20261006.md)。

## 2026-10-06T14:06Z 新测试预算3/3停止：完整563UI/native3，最终freeze Timeout，finally0

- **完整仍BLOCKED**：自然0/3/140与20min/续答自然5min/expired409、四native/Python未通过，mainraceUNPROVEN/formal-prod关。两永久C区测试各3次apply_patch后停止，无第四次/复制绕预算/B区业务改动/部署/restart/历史迁移。下条inprogress为较早进度，本条收口active0。
- 第1actual compiled-driver RED1→3组GREEN0、相关4SFC202pass；单b274真实满分00202报告五阶段/native666790/f8ef08…独立DB-private/dataSHA匹配且finally0。新fullf12四default/人员/freeze、563UI真实保存（560完整+3partial）、三manual/三native/report五stage全部passed；后续Promise Error被共享active误记report-download。不能据最后标签称下载失败，也不追溯旧b254原因；具体自然/续答底层cause未采。
- 第2异步归属RED1→GREEN0：error自身safeStage及固定category，不存rawerror。单零答13ee停freeze Timeout/未开卷0答/finally0。第3实际freeze已选夹具RED1checkedfalse→GREEN0checkedtrue，仅等精确配置加载且未选才click；最后9b仍freeze-00202-candidate Timeout/HTTP未采、cases0，未进入自然，三轮停止，不把localGREEN当远端修复。原旧两批及所有新fail保留。
- full三自然原1500sec deadline14:11:59/14:12:49/14:12:53Z，13:51提前finally不是自然PASS。正常admin续答issue/同paper140/原deadline/URL清已证，expiresAt1791294767到期前结束，5min401与expired409未验。Pythonvenv3.14.4已配置，实际事件oracle gate1/NOTEXECUTED，不运行四oracle或冒ASTPASS。
- 四本轮批各cleanup/final0，14:06:20 freshSSH首次0/7exactroots0/11-private0/15FK/旧465与393manifest/source/config/runtime/schema/cache同，历史3-2-1-0/newside0；PID17552/backend8baa/front353c及三health保持。692保护文件仅两测试变，215Go/frontend src/旧receipt0，ownedprocess0，原两node保留。14:05原集成getInfo200/admin/storage0但homefalse，保留当前页/不说首页，不导出credential。
- 新经验：并发observer的失败须绑定自己的stage，不读最后一个共享facts.active；waitForURL不是认证证据，用户确认后只检查自有context实际页面getInfo。exact最终freeze的具体子动作仍欠，不以已有common坑猜原因。[全部真实结论与收据](management-traits-default-staging-result-20261006.md)。

## 2026-10-06 新有界测试诊断：单00202报告PASS，完整批待结束

- 用户新授权最多3轮、仅测试驱动补证，不部署/restart/历史迁移。仅两个既有C区脚本第1轮：实际编译驱动本地RED1→三合同GREEN0，相关4SFC/API202pass/exit0，两syntax/diagnostics0；固定安全报告stage/class/HTTP、paper→run精确行、模态关闭屏障及normal独立context页面getInfo代替silent900000URL。不猜旧report Error或旧login不同window根因；旧两summary原样。
- 集成4304030e…13:29getInfo200/admin；独立窗口即时通知/结构化确认，凭据不出浏览器。单b274c34b667c auth200/freshSSHbaseline0，仅1新00202candidate full140UI/满分/manualcompleted、五报告stage全部passed/native666790/SHAf8ef08e7…，DB+private及dataRaw独立SHA匹配；13:43:34cleanup0/13:43:36final0/owned11-private0/旧465-source-schema-cache-runtime配置/PID17552/backend8baa/front353c同，窗口关。
- 四组合后续新批f12bb9277e40已正常auth200/baseline0/四新root；自然/四native/Python仍未完成，不宣称产品DONE。publicationcompleted/diagnoseinprogress/naturalnotstarted/cleanupnotstarted，仅diagnose active。本次venv3.14.4已显式配置，尚未实际oracle。详情[本轮进度](management-traits-default-staging-result-20261006.md)。

## 2026-10-06T12:46Z 新默认前端staging发布完成；完整UI阻断且安全收口

- 用户明确Azure subscription1a55…/rg-positive-2026/vm-ubuntu-go-dev/fdpo.onmicrosoft.com是20.200.136.133 staging环境确认，不NSG/防火墙/IP/账号授权。已有frontend-only统一发布批准实际执行，不portal取秘密、不后端build/替换/restart/production；下条SSHblocked及“新默认未发布”为旧事实，由本条限定解除。
- 首strict SSH0/vm-ubuntu-go-dev/liming；只读首次CRLF exit2，整段LF后0。641旧scope仅approved form不同/Go215同；fresh Vue生产build0/原warning保留。11:36:59新front原子exchange，393公网每asset200/大小SHA一致，backup/public/final各native0。index3b83…→353c9fb3a0f30540c8ca091885b67bf458249c1023eca909f78c83988bf27de2；backend8baa…/PID17552不变，任何restart/reload0。
- 新受限backup /opt/talent-assessment/backups/mng_default_20261006_1135 root0700/files0600，旧tar626104c9…/manifest7a5c71c0…、新manifest6cbf324d…，gzip/tarcompare/SHA/回滚目录有效永久保留；exactupload0。仅新C区driver+必要docs，无B区/旧receipt修改，不旧21rollback。
- **完整产品仍PARTIAL/BLOCKED**：首批MTHb254aecaad41正常独立login11:49:57getInfo200；4新UI默认不点击TEST/保存取消独立freeze、4tester draft准备及4freeze通过，三个140UI＋partial3共423保存、两manualcompleted；一00201tester真实native661148/b1e10de1…，不是四报告/独立SQL-PDForaclePASS。第二报告Error cause未留仍UNVERIFIED，readonly归档report1/固定failure事件0不证明产品无错。自然0/3/140原1500秒尚未到期便finally，未验自然/20min/竞态。11:56:35exact4PKcleanup0/11/private全0，11:56:40final0/旧全baseline同，原failure保留。
- 三次测试错误修正后停止：CRLF边界、独立文档/阶段增强、嵌套String.raw正则`${}`syntax1→exact文本匹配syntax0。新clear-reselect/resume-expiry/expired409仅代码未实验，不称已解决原Error。最终MTH9d23f20f068b normal-home900000ms TimeoutError/native1，即使用户确认不冒getInfo200；cases0/baseline0/资源0、不称cleanupPASS，独立browserclosed。Python仅明确配置venv3.14.4，四oracleNOTEXECUTED。
- 12:46:38freshendSSH0：三activehealth/PID17552/backend8baa/front353c保持，11逐0/private0/15FK/旧465/source/schema/cache及新旧manifest/backupSHA同。历史3/2complete/1state0/0state1，snapshot-run-revision0不迁移；641/215逐SHA无漂移，ownednode/browser/debugger0，原两个node不kill。发布completed/UIblocked/natural-native未完/首批cleanupcompleted/最终不需要，active0/formal-prod关/mainraceUNPROVEN。[完整实际结果](management-traits-default-staging-result-20261006.md)。

## 2026-10-06 新默认frontend-only staging已获批准；fresh SSH双timeout阻断

- 用户明确“批准，完成所有业务功能”，本轮批准仅00201/00202新建默认TEST前端统一发布至20.200.136.133 staging及随后完整UI/自然/native验收；不后端替换/restart、历史迁移、00401/MBTI或production。既有另确认发布的待授权记录由本条限定替代，不重复索取相同发布批准；用户转交独立CodeReviewer PASS为交接事实，不冒本轮自行派发审阅。
- 本轮严格既有liming/key/Strict/Batch/ConnectTimeout10/ConnectionAttempts1首次及唯一重试均连接前timeout，nativeExit255/255、10033ms/10028ms、stdout0、stderr各67bytes/SHA dbc4779f…；父Node退出1。远端hostname/id命令未启动，不归因为认证失败或产品bug，两次上限后停止。
- 前端发布blocked；四新UI、自然3例/native4、续答及历史只读本轮均notstarted。候选包fresh SHA/inputs门禁及远端11表/private/旧465/源/Schema/cache/三PID/backend/front本轮未核验；所有旧PASS及8baa…/3b83…仅历史，不冒本轮fresh事实。未执行备份/上传/dist切换/SQL/browser/服务restart，新owned0、cleanup_required0（不是远端cleanupPASS）。
- 仅新增C区[安全通信收据](../scripts/test/results/mng-default-staging-20261006-ssh-blocked-2a61c8/ssh-readonly.json)及两个既有文档追加；业务源/测试driver/旧receipt不改，无Go全量或Python命令。所需用户动作是恢复本机到该staging VM的TCP/22可达性；通信恢复后续用原批准，从fresh基线和候选门禁继续，不通过HTTP绕行造数。

## 2026-10-06 P1清空同库重选限定local GREEN（未发布，独立复审欠）

- [纠正 - 2026-10-06] 下条“清空后默认”及81项GREEN未覆盖00201/00202清空后同库重选。用户转交CodeReviewer确证P1；实际SFC新增4回归，patch前83pass/2fail/exit1，两code均true,false,false，期望true,false,true。仅clear分支响应式重置row.repoCode一行；admin85/85、六相关224/224/native0，连续同code显式取消仍false，历史false/冻结true/非002/00401原断言全保留。
- 完整前端任务只跑一次32files389pass0fail；接口无numeric exit不伪填。npm build:prod原生0/Build complete，原两体积warning及Browserslist提示保留；两代码diagnostics0。本轮682项SHA仅form/admin测试变，215Go逐SHA同/永久driver0/其他前端src0，无Go重跑、SQL、远端、部署、restart、旧21rollback或格式清理。
- 浏览器本轮未GREEN：隔离模块首缺；复用已有模块后clear等待三次超时，最后actual DOM已unchecked/class el-checkbox，两个注入等待选择器native stdin实核中文→002????。是PowerShell→Node传输编码故障，三次同类停止；cases0/contexts-server closed，不第四次或弱化产品断言。永久driver不改，原六casePASS仅历史，新的远端完整UI/native仍欠。
- 独立CodeReviewer工具不可用，须主协调者派发精确本次一行/四回归复审，不以selfReview冒PASS；新默认未部署/正式批准未解除。local1/local2的SFC全前端build完成，browser replay blocked单列；stage3notstarted须review＋另确认frontend-only发布，stage4本轮文档完成。有效RED与所有失败独立ignored收据保留。[本轮完整纠正](management-traits-new-default-local-20261006.md#L3-L12)。

## 2026-10-06T09:16Z 新建002默认新版本地完成；fullUI inspect启动根因GREEN

- 用户已明确“新建测评默认新版，历史保留”，只00201/00202；本轮**新默认未部署/远端未验，四新完整UI/native未重跑、主race UNPROVEN、formal/production关闭**。不历史迁移/重算/回填/旧PDF生成或删除，不改00401/MBTI/Go/共享cfg，不上传或restart；下一统一frontend-only staging须另确认，不沿先前完整发行自动追加发布。
- 仅form.vue新建管理员/无route编辑ID及已保存ID/legacy+legacy/join1/单一002精确code/140单选其他0默认勾选，复用25分钟/showPdf false；字段子集保留、非法字段不自动过滤，清空/切非002清除，同库手动取消保留，另一002重新默认。历史异步回填不触发、frozen保持true/只读；Save与独立freeze确认原合同不变。API后端仍按真实冻结/快照/令牌，不用UI默认当授权。
- 有效SFC RED75/6/exit1→81/0/exit0；全前端既有任务32files385pass（此前363，接口未暴露nativeexit不伪填）；production build native0/两体积warning＋Browserslist年龄提示保留。真实本地构建Chromium6case/0pageerror/exit0：四code×开放封闭新建不点TEST就默认选/正常Save→取消freeze，历史oldfalse与frozen true/35min保留。全部API仅localmock，不冒真实DB/staging。index ac7684e2dce710c9bc97fa72686ffaea35aed9e99d261929ddbac739cd2d8493，线上3b83…未替换。Go三既有guard/router/service编辑器253pass/0fail，仅本地，215Go逐SHA无变化，不重跑6176。
- [纠正 - 2026-10-06] 上条exitnull cause未证现已限定定位：第四人员新增200确实成功，取证 `node -e loader inspect ...` 在实际Node16进入交互debugger，240008ms/SIGTERM/ETIMEDOUT、stdout263、SQL未启动，不是人员新增/SQL锁/maxBuffer产品bug。仅--仍失败，最终固定非保留mng-evidence首位置参数＋loader剥离，合成原生baseline/inspect0及四ID合同GREEN；原240/120预算、SQL/权限/清理不变。runner安全补signal/errorcode/elapsed与evidence-stage，旧失败保留。
- 09:16:06Z corrected真实evidence仅已清理四PK只读inspect：首次SSH0/父0/1407ms/stderr0、FK_ENABLED1，tester/paper/report/bundle0行。新本地诊断目录，原uf054 summary/20files未写；没有新造数或cleanup/备份/权限操作，不冒全健康/465/PIDfresh检查。最初诊断误入Node调试器并未连接SSH；固定哨兵后的本次才真实SQL。
- 本轮786项SHA仅一个frontend src、一个既有SFC测试、一个fullUIrunner变化；两个新C区测试和必要docs。本地browser/context/server关闭、两既存node布尔owned/debugger均false，不kill旧会话。独立CodeReviewer禁止派发，只有安全自审不冒独立PASS。旧641/发行front基线保持，后续新发布获批后建立独立newscope，不重置旧receipt。[完整范围/失败/后续门禁](management-traits-new-default-local-20261006.md)。

## 2026-10-06T08:54Z 只读现状核对：完整发布已完成，最新 full UI 已失败收尾

- [纠正 - 2026-10-06] 下条“新版UI实跑仍未收口”已过时。最新 uf054-ui-37f80e5d48fd/summary.json 于08:51:49Z结束：ui blocked、natural/native notstarted、cases/native均空；cleanup completed、ownedResidual0、ownedBrowserContextsClosed true。四份新UI/native与本批自然三例均未通过，不能沿旧UF050四功能UI或UF051单native样本改写本批结论。
- 四人员收据已写，primary202实际HTTP200/code200/exactScope true；随后的evidence-inspect记录exit null，summary保留AssertionError及旧active标签prepare-owned-tester-primary202。因此不能据标签认定第四人员新增失败或产品bug，inspect终止原因未证，不新修测试/业务或重跑。
- 实读完整发行原生exit0/DEPLOYED_STAGING：后端8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676、完整393前端保留、正常主重启1次；最后08:51独立final nativeExit0/PID17552/private0/主计数回78/70/1488/134494/295328/1349/27，旧465/源/cache/schema等同。上述远端事实是原收据，本轮未新SSH/SQL。诊断d61006081401目录无summary，不当完整suite；其08:21清理证据沿上一条记录，不重复清理。
- 当前仅两node进程，命令行内安全匹配fullUI runner/final watcher均false，不输出命令行、不kill；最新batch已finally收口，没有发现需接续的owned清理。原集成页使用完整pageID清page/context mocks后08:54:22真实getInfo HTTP200/code200/admin true，凭据仅浏览器内存；短pageID首次not-found保留，不新建管理员会话。
- 当前表单仍显式勾选002 TEST链，保存后另行冻结；profile冻结只读、TEST结果含管理员续答/生成/查看/下载，不提供正式报告/旧API回退/历史重算。新建默认新版与历史两完整卷新评分或仅换版式不是同一政策，尚不能把“新版替换”自行解释为历史迁移或全应用重写。
- 本轮仅该既有记忆追加；不发布/造数/代码修改/历史迁移/测试重放。发布完成与最新UI验收失败分别报告；旧21来源未知保留但不重新阻断已获完整scope批准的发行。下一由主协调者一次确认002替换政策，再决定有界测试诊断或独立历史方案。

## 2026-10-06 完整候选 staging 发布完成；新版UI实跑仍未收口

- **未验先列：本轮四组合完整UI/native4份/自然三例尚未完成，主HTTP/Worker争用UNPROVEN，formal/production不启用。** 用户最新明确“完整的发布，然后开始新版测试，包括完整UI”，覆盖此前仅最小插桩diff的发行范围限制；旧21来源未知保留，但不再以无法证明onlyinstrumentation阻断经审查/验证的当前完整candidate，也不宣称旧新等价。旧历史exam1776822816300709851只读保留，不迁移/recompute/backfill/生成或删除旧PDF。
- 07:52fresh严格SSH首次0、state1全0/expired00401待处理0/11侧表逐0/private0/15FK，历史3卷=state2两+state0一；源码UF053 ended641项逐SHA稳定。freshWindows/Linuxserver/buildall/vet各0/输出0，6176pass/659top/0fail/9skip复用同源码；fresh前端任务32files363pass，已有VTU deprecated提示保留。当前7非测试旧漂移及UF053源实读安全自审，不冒独立CodeReviewer；主提供此前UF053独立PASS不覆盖旧磁盘NOTEXECUTED或扩大其审阅scope。
- 07:56:42–07:57:39完整发行nativeexit0/stderr0。newbackend49902830/SHA8baa7f87…已运行，PID2002完全排空→17552；一次主stop/start，nginx/mysql不重启。完整393前端与currentdist/远端/公网逐byteSHA全匹配，index3b83b976…保留；05c55e77…TESTdocx及b0498249…内容205规则由真实应用账号读取校验，不覆盖。remoteconfigs/env/dropins/unit不改、observer默认off、DDL/GRANT/cache/globalauthpolicy/生产0。
- 受限正式backend_82155be9457ad1bf备份root0700/files0600：完整singletransaction/routines/triggers/events全库SHAb034d5ba…、旧binary原uid/gid/mode、完整dist/configs、unit/dropins及两上传树。gzip/SHA/application-system tarcompare及rollbackmetadata能力核验成功，未本轮新恢复演练、不下载secretarchive。pre/stopped/post三个actual应用账号full11/159/67/21/cached4/CRUD/source140700/TESTgatePASS；health及管理401/参与者401/空body400正确，general001002003MBTI004五只读HTTP200。legacy12SHA051795bd…/旧465逐路径SHA/immutable/Schema不变，payload五exact上传清0，rollback未需要。
- 发布后原页07:57getInfo401停造数；用户正常login后07:59getInfo200/admin。独立native窗口另正常login不复制token。首批MTHf3ee9513051f四次同add route取消freeze保留id，实际仅1root重复保存，人员准备前Timeout；精确清0/旧基线同。测试流程修新文档+distinctroot，第二批MTH785b4387200a正常四不同root后primarytester选择器Timeout，实际未createperson/paper/freeze，四root清0、08:11final11/private0/PID17552/旧465/runtime/source/cache/schema全同，旧失败不改。
- 达3次编辑后用户另批准“1轮仅测试定位/修正及完整UI复验，不再发布”。唯一诊断MTHd61006081401/APIprepdraft1791274764808276191（非完整创建UI）：真实精确query200/code200/rows0、该ownoption可见；弹窗原嵌套has selector0、相对标签1，证明测试定位失败，不放guard或修业务。按新预算改精确相对CSS、补必填stuFlag及封闭前name-only；08:21exact诊断清0/旧465与current12/runtime同。最终手动登录等待15min超时，尚未baseline/业务0；仅内存正常首页自动检测替代terminal确认，仍getInfo200前不造数、原等待预算不延。
- 三个C区新driver及一C区新发行报告，运行产品源/前端源码修改0，不回滚既有changes/Git提交/push/现代化workflow操作0。详见[完整发行及实际失败记录](management-traits-full-staging-release-20261006.md)，后续须以新batch最终summary和exactfinally结果更新，不沿旧四UI/native单样本宣称本轮通过。

## 2026-10-06 UF053发布获批：历史资格只读完成，最小整包来源门禁BLOCKED

- **未发布/未restart/未生成历史PDF；主race仍UNPROVEN，管理UI401未验证。** 用户本次明确批准仅stagingbackend最小单卷观测＋一次维护restart＋未来专属合成卷；不生产/前端/cache/权限表/历史snapshot-profile-run创建、重算回填或旧PDF删除。泛“是否可以”仅资格核查，旧完整报告不能自动重生成。
- actual路由examId1776822816300709851/isOpen1；真实SQL标题匹配1、legacy/legacy/25min、00201单库140题。有效candidate3/有paper3/end非NULL2/旧pdf_path非空2；三卷candidate同paper+exam各1/tester0，state2两卷各140/140/700桶140勾选，state0一卷0/140。profile0、三卷snapshot/run/revision逐0、全库11sidecar逐0。旧PDF仅引用存在性未核字节。新版接口必须explicit run＋new_creation冻结输入、13维4模块receipt，**现接口不能直接生成历史新版**；完整答案不证明旧V/题本/人员/版本证据，需单独历史兼容/迁移确认或独立重新测评，不补造当前源映射。
- 原page清routes后07:37:44getInfo401；未伪auth/继续业务API，SSH只读metadata/聚合成功。首次选项JOIN误用pq.id导致NULL，保留为无效审计；实读组卷桶qu_id=源ID后改pq.qu_id真实SQL：完成280题每题5不同1..5/一勾选/所选score与actual一致，未答140题勾选0，420题新版raw/final/examQuestion指针均NULL。桶标记80/200及40/100仅历史事实，不称新版13维资格。
- 07:39:43fresh主PID2002/backend8fb264e…/front3b83b976…/三activehealthok、未到期state1=0/expired00401=0/11逐0/private0、两报告ENVstaging/观测键未设。没有upload/备份写入/stop或配置操作，维护restart预算未用，新owned0/cleanup_required0非cleanupPASS。
- 本轮独立UF053after641/641同、Linux49902830/8baa7f87…同；原6176/0/9/build-vet0仅复用。旧21current稳定但原hash0匹配，7非测试含后端handlers/公共LO，before自身不能证明等于线上8fb源码。旧本地8fb二进制匹配线上，新旧buildinfo同脏revision不证明语义；可达Git16源码版本及LF/CRLF/BOM均不匹配，旧application归档后端.go条目0/latest证据只有checker源和旧bin。**最小scope门禁未解，禁止夹带旧未知业务变更发布**；UF053磁盘verdict/final仍独立reviewNOTEXECUTED，本模式不派agent或冒reviewPASS。
- 一次SSH SQL成功后archivepipeline退出1/stderr50/原因未采；后续独立Perlarchiveprobe0/stderr0，旧失败保留。仅C区新限定报告＋两既有docs追加，源/前端/旧收据不改；主待办historycompleted/publishblocked/reportnotstarted/active0。[真实结论与门禁](management-traits-history-release-preflight-20261006.md)。后续由协调者取得原源码或明确额外scope并独立review；已有最小发布批准保留，未解决不能只改PASS标签或再批准泛发布绕过。

## 2026-10-06 UF053：A 本地最小观测 GREEN；未发布，主竞争仍 UNPROVEN

- **未验先列：主HTTP/Worker真实竞争/staging新观测/真实MySQL及独立CodeReviewer未验；race detector CGO0不可执行exit2。** 用户askQuestions明确A“先增加最小脱敏观测并本地验证；staging后端发布及重启再单独确认”；本轮SSH/部署/restart/远端造数/共享cfg/SET0，不沿旧批准扩大操作。BreakGlass禁止再派发agent，仅自审，不称独立reviewPASS。
- 实际HTTP认证后SubmitParticipant固定http_participant，scanExpiry通过private typedcontext给Submit标expiry_worker，其他Submit是trusted_internal；main同一runtime/pool。两个新process键MNG_RACE_OBSERVE_ENV（精确local/staging）＋MNG_RACE_OBSERVE_PAPER_SHA256（domain mng-race-paper-v1+NUL+canonicalv4UUID的SHA256）构造一次读取，默认关闭，仅exacthash匹配；目标归属另需owned门禁，语法不证明synthetic。无rawUUID/人员ID/hash值/PII/token/SQL/rawerror输出，随机opaque session/resource/attempt同资源稳定、每attempt独立，无新密钥。
- 正常9阶段/固定10容量，锁内仅monotonic时间与数组赋值、原Transaction返回后原slogemit；paperUPDATE→bundleSHARE/SQL/clock/期限/Workerinterval/事务/API/返回类型/权限不改。entry不是socketrecv；wait是锁SQL调用起点而非实证DBwait；process_ns只同session比较、不能混减P_Stimer；日志时间不是事件发生时间。nilerr且连续body+transaction正常返回才committedcreated/reused，panic/失败/缺终态failed不声称DBrollback成功。观测能力不PASSrace。
- 新compileRED→首18pass3fail（新metadata时间/SQL匹配夹具）→23pass；增强时括号编译失败为无效行为RED，保留；有效未完成attempt误报commit RED0pass1fail/exit1→emitter最小收紧→最终service25/Gin7事件PASS。覆盖真实ScanExpiry及两入口并行、完整/复用零写/多失败、默认关闭/配置不可变/容量/隐私/事务外sink；Gin真实签发token沿原handler验证caller不能伪造，DBsqlmock，不主实库HTTPrace。
- 本轮默认全Go **6176pass/659顶层/0fail/9原skip/parse0/exit0**，新增32事件/7顶层；Windowsserver/buildall/vet/Linuxcandidate编译各numericexit0/stdoutstderr0，源码diagnostics0。原生测试替代editor未发现；全量Raw Output仅byteSHA不回显/落盘。新格式规范化相等，CRLF/EOF不顺手重写；终验首次失败是新报告158超145行链接，纠正链接后再验证，不业务/source漂移。
- 修改仅5既有service源＋新private观测/两新Go测试/C区本地验证入口与必要既有docs；无Legacy/front/API/SQL/模型/配置文件/规则/旧receipt覆盖，旧21历史漂移未知保留，不rollback。scope与候选完整SHA见最终本地收据；原native/四UI/自然三例/恢复库race与旧失败不改。下一CodeReviewer独立只读→用户另批仅stagingbackend＋一次维护restart＋exclusiveownedtargethashoptin/精确验证清理，不自动执行、不再25min碰运气。[完整影响与门禁](management-traits-four-real-verification-20261003.md#L3)。
- [最终收口](../scripts/test/results/uf053-local-observation-20261006/verdict.json)exit0：639→641快照/633原项SHA同，5既有service改动＋本轮测试增强、新增private观测/Gin测试；Go212+3=215、front变化0/旧receipt30同，Legacy快照0仅操作0。driver syntax0/24新doclinks有效/8Go规范化format同，rawCRLF/EOF不改。Linux仅本地候选49902830bytes/SHA8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676，未上传，线上8fb…/front3b83…/PID2002仅最后历史未fresh复核。本地设计/实现/记录完成，不更新workflow；主race/部署/独立review继续未完成。

## 2026-10-06T05:59:06Z UF052：主竞态只读能力核对完成，调用归属仍欠（B）

- **主HTTP/Worker race仍UNPROVEN，不新增数据/25min复跑/native/tests或部署。** 首试strict SSH exit0、22项只读查询各exit0/errno NULL、stderr0；MySQL8.0.46/performance_schema1，现有观察连接global SELECT/PROCESS true（账号不输出；未验证应用账号权限）。使用既有sudo/socket认证，不新取凭据或配置。主PID2002前后同，未fresh核健康/旧465，不能沿旧终验冒本轮全系统SHA。
- statements/transactions current+history、global/thread instrumentation、statements_digest均YES；两history_long均NO；statement/sql159与com33全部enabled/timed，transaction1/1/1，history每thread各10。threads43/foreground8/instrumented43/history43，locks/waits0/0；statement current/history/long6/51/0、transaction9/62/0。仅元数据/计数/自身投影与锁WHERE1=0，不读取实际他人SQL_TEXT/参数/DIGEST_TEXT/LOCK_DATA；自身留存COMMITTED行只证明投影可读，不作为新事务证据，0锁不证明Worker未运行。
- 实际main将同runtime传router/Worker，guarded GORM保留同pool；Worker特有scan先归还连接，再Submit开启事务，两条提交共用paper UPDATE→bundle SHARE/相同SQL。threadID/connectionID无HTTP URI/goroutine身份，扫描digest不能跨pool复用标后续Worker claim；10条history会被逐题/事务SQL覆盖，不启long、不碰performance_schema SET/GRANT。**当前只能准备精确owned锁时序观察，不能保证唯一caller归属，故A门禁未满足，不再自然等待碰运气；不是产品bug。**
- 可执行下一边界：未来仅exact synthetic paper/PRIMARY RECORD/owned UUID的服务端LOCK_DATA过滤→精确waitpair→thread/event/transaction/timer与allowlisted normalized digest，HTTP采send/receive/reused；缺UUID/事件/唯一caller即UNPROVEN，跨表非原子/轮询漏瞬态不当无竞争。若必须主级实证，B需另明确固定安全caller/stage instrumentation＋backend staging发布/restart；专属协调行锁亦需单独写授权且是人为争用，不冒自然。C可明确验收豁免此额外主观测，保留恢复库race PASS＋主自然HTTP PASS与主race UNPROVEN；当前无权限缺口，不申请多余共享SET。
- 617scope/Go212及旧UF050全部17JSON逐SHA保持；只新增C区安全收据/方案及两既有docs，不改源/配置/规则/账本/workflow。一次本地模板syntax1发生在SSH/文件创建前，仅修引号后执行，失败保留在方案中。[能力收据](../scripts/test/results/uf052-race-readonly-da33f5a3ebc5/attempt-1.json)、[最小方案](../scripts/test/results/uf052-race-readonly-da33f5a3ebc5/minimal-plan.json)、[完整限定结论](management-traits-four-real-verification-20261003.md#L3)。旧单native648182 PASS与四native失败仍各自历史，不回填。

## 2026-10-06T05:51:15Z UF051单样本真实native落盘PASS；主竞态证据仍UNPROVEN

- **未验先列：主HTTP/Worker实际重叠仍未证；旧四report native0/4历史不改，新证据仅1份00201candidate。** 旧browser的04:46:15.806/04:47:19.905/04:47:37.244是观察响应时间，没有request发送时间/reused/Worker开始；SQL只存秒级submitted，未采created/source。当前源码两路径共用Source=submission/服务Submit(manual)，timeout仅锁后时间到期，不是Worker标记。05:27只读精确三owned paper journal窗口命中0/parse0，不证明Worker未运行；恢复库真实进程/双pool锁竞争PASS仍仅服务层，不替主HTTP。
- 能力先验：集成dataURL纯syntheticBlob anchor的8秒download事件timeout/saveAs0/精确Downloads标记0；独立已有Playwright1.41.2/Chromium121/Node16同类Download事件＋saveAs成功606bytes/SHA dbd7770588a64ca9a33386d6dd685c21f298b20ad8a875d2ccd6ad9b8dce9829。工具链差异实证，具体VSCode拦截内部未核，不称产品bug/不改业务迁就工具。两PDF都不是APIbytes写盘或SCP。
- 原页05:24getInfo401后停业务请求/保留首页；用户明确选择独立窗口正常登录及仅1新native样本。首独立窗口05:47无凭据，业务0/未SSHbaseline，失败保留；仅补该窗口自身正常Admin-Token Cookie兼容，不复制原页token/会话文件。第二窗口05:50:07getInfo200/code200/admin，用户正常login，凭据仅浏览器内存。freshStrictSSH首次0/备份9ebee6a…gzip/SHA/主11逐0及旧基线同后才写。
- MTH5c4639715720只1新00201candidate，正常API配置/freeze/登记/组卷/140save/manual completed，明确不是新完整参与者UI；started05:50:17→deadline06:15:17原1500秒。实际UI生成/下载按钮Enter，05:51:03首次Download事件＋saveAs1，report626546ba-6f1d-4a59-9711-7289918843ba PDF648182bytes/SHA **e10282ec1ecd7f8867a5ac86f6cbaea51bea0d6074b8a00666fba5f5dbc26e30**，HTTP200/applicationpdf/%PDF；DB/private0600独立大小SHA精确同，SQL140/score50/1run13dim4mod1receipt/source submission/created=submitted。不用新1份冒旧四份或自然race。
- exact-PK备份root0700/files0600/gzip/SHA cf24286fc69ae99405b08633a0b9cba4f6333976fbdea010c38dbc7921fff6f5保留；原清理器仅收窄arity/title/count到1，FK/SafeUpdate/共享引用/UUID-SHA守卫不变，05:51:09commit0/owned0/private0；05:51:14独立final0/主11逐0/旧465-current12-source4-runtime-config-schema-cache同/PID2002/三healthy/backend8fb…front3b83…保持。独立context/process关闭，原集成admin首页/context1保留，新会话清于context关闭不logout原页；617scope/Go212及旧全部JSON SHA同。
- 两C区测试helper，syntax0/editor0，真实driver最终nativeexit0；新样本源代码/API/guard/env/cfg/worker参数/业务timer/时钟/期限/DDL/部署/restart/production操作0，未重跑560UI或全Go。工具读取两错误猜路径及关闭已消失synthetic页失败保留，不当业务错误；独立contextpages1无synthetic页实核。单样本native门禁关闭，完整产品仍PARTIAL；主race若要求同层实证，需另明确安全成功attempt/锁时序观测边界，不改后台logging或重跑25分钟碰运气。[完整限定收口](management-traits-four-real-verification-20261003.md#L3)、[新summary](../scripts/test/results/uf051-product-native-5c4639715720/summary.json)。
- [最终补充 - 05:54:18Z] 原集成页清routes后真实getInfo200/code200/admin/home/participantStorage0，无credential导出；独立最终断言exit0、两syntax0、617/212逐SHA同及原UF050全部17JSON同、24doclinks有效、新收据JWT字面量0。[五阶段限定收口](../scripts/test/results/uf051-native-capability-20261006-a71c90/final-scope-verification.json)。

## 2026-10-06T04:52:13Z UF050发布/四功能UI/自然3例/finally0；native与主竞态仍PARTIAL

- **未通过先列：native4report各2×20sec eventtimeout/saveAs0；主HTTP/Worker时间重叠/锁竞争未证明；00202candidate首140细粒度数组丢失，仅执行/刷新/SQL聚合。** 不用SCP/唯一run冒native/race，正式内容/production/mainrestart/ImportData状态入口未扩。四功能UI4/4不含单独native门禁，整体PARTIAL。
- 发布仅front：SSH read首次native0，8资产strict6module/status-only/scope/hashrefs/index/gzip及公网393全PASS；旧index98547b68…→新3b83b976…，受限backup uf050_20261006_041149/root0700/files0600/tarSHAcab27a29805b69479f41176de2c7f138a709cd73de1ac590fdea0ed752f5e25f、manifest2de67af…/旧393即时回滚目录验证保留，实际rollback0。server8fb264e…/PID2002/三active、config/source/schema/cache/旧465保持，upload0，无nginxreload/任何restart/DDL/production。
- 正常UI4draft/exact未冻结scope新增刷新四次200/payload字符串0→全部人员准备后4freeze/2bundle，SQL4status0/nonNULL，四正确密码login200code200五键，UF050正常手工新增STAGING VERIFIED，不回填旧状态/放guard/改ImportData。六卷正常开卷/140700/原1500sec；四全答563含partial3，用按钮键盘Enter非fetchloop/不冒鼠标click，423细response保留。三个manual正常submit、四admin13维4模块/score50/TESTgenerate-view200/四私有PDF与DB-SCP大小SHA一致仅安全合成留档。
- 真25min自然140/0/3原deadline04:46:15/04:47:19/04:47:36Z，UI自动submit04:46:15.806/04:47:19.905/04:47:37.244各HTTP200code0并完成页；20min提示三可见。SQL三timeout、140completed50、0/3incomplete overall+13dim+4mod全NULL/noPDF；六各唯一1run13dim4mod1receipt，期限/答案未改。三原genuinepaper凭据后续save各409；incomplete admingenerate实际4次409（helper重复观察6receipt/3paper未去重导致多发各1次，不删除/伪称2次）。主Worker竞争仍UNPROVEN，不回填恢复库历史。
- 04:51:08六owned页清storage/原凭据globals/timers关闭；04:51:24原4PK/FK事务cleanup首次exit0/只四SHA核过UUIDprivatePDFunlink；04:51:28独立final0/11逐0/private0/15FK/主计数回原/旧465-current12-source4-runtime-config-metadata-schema-cache同，617scope/212Go及旧收据SHA同。04:52:13全部真实PK再核0/旧393rollbackSHAvalid/upload0；readonly连接SafeUpdate默认0，cleanup SESSION1，不改global。04:52:11原getInfo200/admin/home/storage0/context1/own密码global清，pending空/owned后台0，备份/合成证据保留。
- 工具经验：同page新的长browser调用可使旧deferredID不可取得但观察回调仍收尾，须专用page并等待期间不再该page长执行；收据按paper去重后才发负向请求，保留重复raw证据。集成页鼠标稳定性不可当业务bug，可用正常按钮键盘Enter，必须真实网络响应并准确标activation。没有改业务timer/clock/时长/DBdeadline/为了native写APIbytes文件。脚本三syntax0/editor0、real六case断言exit0，旧8/184/264/build复用不6144重跑。[最终完整scope](management-traits-four-real-verification-20261003.md#L3)、[summary](../scripts/test/results/uf050-ui-f61006041501/summary.json)。

## 2026-10-06T04:33Z UF-050前端发布及真实status/login完成；自然suite仍活跃

- **pending先列：自然25min0/3/140截至04:33尚未到期；完整UI3/4、native三份各2timeout/saveAs0，finally未执行，owned4exam4tester2candidate6paper仍待清。** 不称整体DONE、SCP不native，主HTTP/Worker竞态未证明。自然原期限04:46:15/04:47:19/04:47:36Z、各1500秒未改，有界browser deferred观察实际运行。
- 04:10:49SSH首次native0/旧tar12472320/SHA98b81de…/PID2002；离线严格6module/status-only及scope/hashrefs/index-map/旧新gzip全PASS，8/393完整归因exit0，不靠385同冒PASS。617原scope仅原SFC/212Go保持，旧source21不再阻当前已通过版本发布。
- 04:13fresh受限旧dist backup root0700/files0600/sha/gzip/tar-compare及非front/旧465/source/schema/cache/DBbackup适用PASS。04:14:02仅dist原子renameat2 exchange成功/旧目录回滚保留/rootroot755；公网393全byteSHA/独立final04:14:10exit0，front98547b68…→3b83b976…/server8fb264e…/PID2002/三active不变/upload0，backend/config/env/DDL/cache/nginxreload/任何restart/production0。
- 原页清routes后getInfo200；新MTHf61006041501全部tester在任一freeze前精确draft查询→正常UI4新增及同scope刷新200，payload字符串0；独立SQL4status0/nonNULL，四正确密码登录200/code200/五键，UF050正常手工新增闭环真实验证，ImportData未扩。4freeze/2bundle，六正常UI卷，三个manual完整各140save/恢复/完成/1-13-4-1/50，三个natural140/0/3未交。正常按钮用键盘Enter非fetchloop，563保存，423细粒度response保留/一candidate140聚合保留但原细数组helper失败丢失不补造。
- 三normaladmin结果13维4模块/TESTgenerate-view200；native两次20秒eventtimeout每份，native文件0。三privateSCP本地合成副本与DB/privateSHA、648173/648017/648043保持，明确非native；safe progress含准确PK无凭据。工具driver错误（repo下拉标签、draft强探测、native/stability/reload readiness）保留，不改业务。仅C区三个helper和必要docs，脚本check/editor0；本地原8/184/264/build复用。后续同一运行继续等待真实期限→第四report/native有界→SQL状态/negative→所有PK/SHAprivate精确finally及旧基线收口。[完整进度与证据](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-06T04:03:51Z UF-050有界read gate：两次连接超时，未部署

- **未验先列：八项完整编译差异归因、frontend发布、真实新增status字符串0及正常tester登录均未完成；phase1仍阻断。** 本轮只接续已批准frontend-only staging范围，不重复申请发布授权、不开展fullsuite或业务造数。
- 固定组合命令同一次SSH验证/opt/talent-assessment真实cwd、非symlink dist、hostname/user/PID/indexSHA并流式tar旧dist；首次04:03:41Z及唯一重试04:03:51Z均原生exit255、signal/error null、stdout0，安全stderr明确Connection timed out（各67bytes/SHA dbc4779f6c9efb426d2ec093088dd86963fbccd91ec0d5dbd98b3a5e08743a72）。Strict/Batch/ConnectTimeout10/ConnectionAttempts1保持；Node exit1，不能以外层PowerShell成功冒充SSH0。远端路径验证及tar均未取得结果，连接预算已耗尽，停止不另握手或HTTP绕行。
- [首次实读收据](../scripts/test/results/uf050-staging-20261006/read-gate-1791259410993/attempt-1.json)、[唯一重试收据](../scripts/test/results/uf050-staging-20261006/read-gate-1791259410993/attempt-2.json)为本轮新证据；上一轮失败原因仍UNVERIFIED，不回填为本轮双timeout。393/385/8仅沿既有历史，八候选tester JS/CSS、app/index及gzip未fresh读取，不称all UF050或unexpected功能已证；旧SHA/current源码反向恢复不能替代完整编译门禁。
- 本轮script patch0/业务源码改动0、远端upload/backup/distSwitch/backend/cache/env/nginx操作0、新资源0/cleanup_required0（非cleanup PASS）。管理员最后真实401仍为03:56:09Z历史，不重试登录/造auth，正常登录需求保留；旧465/source/cache/metadata/PID/backend/当前front本轮均未fresh核。只新增本地安全runtime JSON两份及两既有docs简洁记账，无old-dist tar。后续仍用既有发布授权；先恢复read gate、全部八差异通过才可front-only发布，未知功能差异须scope裁决。

## 2026-10-06 UF-050获准frontend-only staging发布：上传前BLOCKED

- **未发布/未真实复验，todo1未完成**。用户明确批准仅20.200.136.133 staging前端；不重复索取批准，不替换backend/重启任何服务/访问production。03:56:09Z原page清page/context routes后真实getInfo401/code401、participantStorage0，首页保留不等于有效认证；未login/logout/reset/伪JWT或保存凭据。
- fresh严格liming/key/Batch/Strict/Conn10/Attempts1握手exit0、hostname vm-ubuntu-go-dev、PID2002、三服务active，backend8fb264e…/旧front98547…保持。当前617项逐SHA只一SFC变化，212Go及旧11receipt不变；去唯一status默认行恢复原源码SHA3814e0ff…。本地buildexit0复用原收据，不重跑全量。
- 实读线上393与本地393资源：385逐SHA不变；8候选差异为tester JS/CSS、app/index及gzip。app只引用变化、index规范资源引用并排序映射后相同；CSS观察为Vue scope ID，tester还含webpack module ID/顺序变化。**完整严格编译差异门禁未完成**：最终专用测试exit1/SSH read-only comparison failed，初版未保留失败读取阶段、exit/尝试细节，不推断为双timeout。无新一轮连接重跑、没有包/上传/备份/切换；工具缺diff与整包模块重排断言失败保留，不装依赖/改业务。
- 新C区[单用途发布差异测试](../scripts/test/uf050-frontend-release-test.js)补后续安全传输exit/attempt/分类SHA收据；本轮Node语法/diagnostics另核。旧frontendonly脚本已读，因包含无故障回滚的删除与nginx reload未执行。新资源0/cleanup_required0，不称remote cleanupPASS；旧465/source/cache/完整DBbackup/11表/Schema本轮未fresh终验。四完整UI0/4、natural/race/native仍未验，原RED/summary不改。
- [本轮失败收据](../scripts/test/results/uf050-staging-20261006/blocked.json)、[完整交接](management-traits-four-real-verification-20261003.md#L3)。主协调者续用已有授权：SSH稳定及原页正常getInfo200→完整编译差异门禁/当前非前端基线/受限备份与回滚→仅前端原子切换→最小owned UI字符串0/SQL/login五键及exactcleanup；此前完整suite不重跑，不能跳过门禁发布。

## 2026-10-06 UF-050最小本地GREEN（未发布，远端验收不关闭）

- 未验：真实DB往返/staging修复验证、四完整UI0/4、自然25min0/3/140/race/native；无SSH/browser/远端写/deploy/主restart/production。原suite2blocked/3-4notstarted/5completed cleanup0/pending空、原四statusNULL/login200-code500保留，不改旧summary。
- 根因已本地复现：人员SFC reset遗漏status，既存testerReq/model `*string`与Create原样映射令缺字段为NULL；strict identity只接受字符串`"0"`（`"1"`停用），仍拒NULL/未知。仅reset一行默认`status: "0"`，不服务器默认全部启用或回填旧行。编辑整体详情覆盖默认值，0/1/NULL/2原状态提交保留；再次新增才重置0，candidate不动。
- 原实际SFC单项freshRED1fail/exit1；补保留状态测试后业务patch前3pass5fail/exit1→actual SFC8pass，相关5文件184pass/0fail/exit0。Vue production build fresh exit0/两既有体积warning＋Browserslist年龄通知保持；原backend显式0成功、NULL/1等拒绝与Gin/router guard专项264事件/38顶层/0fail0skip/parse0/exit0。无Go修改不重跑6144/Go build，编辑器未发现测试、nativeVitest替代，两文件diagnostics0。
- 212Go逐SHA与freshbefore及原suite相同；617项scope仅一frontend src改变，旧11receipt变化0，21历史漂移未归因；Legacy snapshot0文件仅操作0。原02:26:11旧465/source/PID/健康及finally0是历史，不伪fresh远端验证。新增/编辑一SFC一测试及六既有docs，安全运行收据C区；首次范围校验ANSI格式exit1保留、仅去ESC后同断言通过。
- ImportData独立新建也未赋status、更新不写status，模板/导出无状态列；不扩张为所有入口已修，不顺手改导入/backend/candidate。下次同类定位须查真实nullable model/入口映射，不能相信规则默认值说明。后续只需明确批准此次frontend-only staging发布，无后端替换或服务restart；本轮不执行。[完整消费者/证据](management-traits-four-real-verification-20261003.md#L3)、[local verdict](../scripts/test/results/uf050-local-20261006/verdict.json)。

## 2026-10-06T02:26:45Z 完整UI实际续跑：draft scoped200，新UF-050 NULL状态登录BLOCKED，finally0

- **四完整UI0/4、自然25min0/3/140/race/native仍未验**，停止于新真实业务block，不交“下一执行者”计划。全部4配置正常UI draft保存→2tester scope精确选择/查询→4正常新增及同exam刷新HTTP200/code200→全部准备后4freeze200。旧unfiltered403路径不再执行、守卫不变，UF-049 scope已远端限定通过。
- 正常tester登录第二有界点击02:24:14.563Z HTTP200/code500/身份拒绝，第一响应未保留。02:25:16独立SQL4新tester statusNULL/del0/nameOwned/密码存在/paper-endNULL；2bundle各2owned profile、foreign refs0。当前UI reset无status/Create持久化r.Status/identity拒nil或非0，正常UI创建状态合同断链；未直接采服务器内部拒绝stage，不夸唯一cause。新actual SFC回放payload.statusundefined vs '0' **1fail/0pass/exit1**，原3case未选择；真实UI/SQL合同REDexit1，无GREEN/业务修改，不重跑179/250/build。
- 00201candidate正常登记/准备/开始1卷140题但0click/0答/run0/report0，started1791253388/deadline1791254888原1500秒；server1791253516未到期。0/3卷登录失败未创建，native生成/监听/saveAs均0，未声明race或自然到期。新markerMTHa61006021801，准确4exam/4profile/2独占bundle/4tester/1candidate/1paper闭包，未改时间/状态/权限/密码/DDL/缓存/主restart/deploy/production。
- fresh02:19:07SSH首次0/无retry，getInfo200/admin/wildcard；actual8fb264e/front98547/PID2002/3health/11逐0/private0/15FK/source140140700/旧465/current12-source4-runtime-cache同，完整backup9ebee6a…gzipSHA复用。02:26:06原四PK清理器首次0/FK-SafeUpdate1/foreignrefs0，02:26:11独立final0/所有owned及11child/private0/旧465及完整前后hash同/PID2002/3active/cache755；212Go逐SHA变化0，旧21归因保留。02:26:45原adminhome/getInfo200/storage0/testglobals清/owned页0/timer销毁，无后台任务仍活。
- stage1completed/2blocked/3-4notstarted/5completed，本轮cleanup重开后实际收口，不沿旧PASS冒充；manage_todo_list未提供，仅safeJSON状态。只一新C区取证入口、一既有SFC新增RED及六既有文档，旧失败receipt不改。[完整证据](management-traits-four-real-verification-20261003.md#L3)、[本批summary](../scripts/test/results/mng-premature-timeout-a61006021801/summary.json)、[新SFC RED](../scripts/test/results/mng-premature-timeout-a61006021801/sfc-status-red.json)。

## 2026-10-06 UF-049归因纠正：403正确，测试人员准备scope本地GREEN

- 未验：纠正后的远端完整UI/natural/race/native仍未跑，stage2 blocked/PARTIAL、3/4 notstarted、5原cleanupPASS不变；无deploy/主restart/主库或权限写/owned新记录，业务源码0改。原admin02:13:03.553Z只读getInfo200/admin/wildcard/home/storage0。
- actual owned SQL：candidate profile1，但新tester所属测评profile0/无paper。原GET tester/list无examId而AllLegacy保护全库，正确403；queryParams与新增form独立，只在弹窗选exam不改变列表scope。权限缺失/必须新专用人员API未证。正确测试顺序：全部tester在任何owned freeze前准备；查询先选择精确owned未冻结exam→查询→新增→同scope刷新，然后freeze并走专用参与者/结果，冻结后旧列表仍拒绝。
- 新实际SFC回放原步骤RED0pass2fail/exit1→只纠正测试查询scopeGREEN3pass；相关5文件179pass0fail/exit0。新可复用browser driver只筛选精确owned未冻结scope，403/跨exam/frozen stop，无fallback；本轮未远端执行，Nodecheck/4非法输入合同0、diagnostics0。原Go guard/真实handler/router/service专项250事件/36顶层/0fail0skip/parse0/exit0；既有Go Build＋补充serverbuild/buildall/vet均0，非完整产品验收。
- [纠正 - 2026-10-06] 下条“业务安全列表修复后才能继续”范围过宽：本失败源于验收全库旧列表路径错误，可用已有exact draft筛选，不修改guard/角色/Schema。冻结后人员维护是另一合同，未新增功能。原两JSON msg问号不当原始中文body；历史403/code1/RED保留，固定文案仅当前源码/本地Gin证据。
- [完整归因/消费者/未验](management-traits-four-real-verification-20261003.md#L3)、[本地RED](../scripts/test/results/mng-personnel-local-20261006/red.json)、[限定verdict](../scripts/test/results/mng-personnel-local-20261006/verdict.json)。本轮只两个测试文件及六既有文档/ignored证据，不业务修复/远端BUGresolved/完成workflow；21旧漂移不解除。

## 2026-10-06T02:03:12Z 四UI/natural/native接续：实际人员列表403，finally0

- **BLOCKED/PARTIAL**：四完整UI0/4、自然25min0/3/140/竞争及native落盘未验。新page4304030e…清routes后getInfo200/admin/wildcard；freshSSH首次0/11逐0/private0/15FK/主78/70/1488/134494/295328/1349/27/旧465/source两个140140700/runtime/cache同，已有backup9ebee6a…gzipSHA复用。actual仍8fb264e…/front98547…/PID2002/报告staging，APPENVproduction仅配置；不deploy/restart/DDL/clock/共享权限/cache改写。
- 唯一MTH6c260106a001实际2exam/1profile1bundle/1candidate1tester/1paper。00201candidate正常管理配置/name-only冻结/登记/准备/开始后140真实radio click及140save200，自动翻题/末题140；导航第1题一般，刷新detail200/140/samepaper/全部raw3选项保持。started1791251844/deadline1791253344原1500秒，02:01:53SQLserver1791252113仍未到期/state1/140snapshot raw3与140legacy answered/run0/report0；未交卷，不称score50或timeoutPASS。
- UF-049/MT-ADMIN-PERSONNEL-LIST-403：正常封闭新增tester UI成功＋SQL新PK1，随后旧unfiltered GET tester/list刷新403；相同正常Bearer02:01:00再核403/code1/固定“新版实体禁止旧接口”。current守卫collection AllLegacy按设计failclosed，前端新增后仍无catch调用旧列表；是管理闭环冲突，不能放开密码/guard或绕路继续。captured-real-response最小UI RED0pass1fail/exit1，未修业务，两个00202及0/3卷未创建。源21历史归因保留，current源码不冒充部署语义。
- 02:01:57 exactcleanup只内存原器4→2/count2/精确两个标题，FK/SafeUpdate1/SESSION64MiB/非ownshared拒绝，commit0。02:02:02独立根PK/person/paper/link与11child全0/private0、旧465/current12-source4-runtime-config-metadata-schema-cache同/PID2002/三健康，pending空；无PDF/上传/备份删除。原期限未改，owned页面销毁止timer后及时清，不等自然到期。
- 02:02:16原index/getInfo200/admin/wildcard/storage0；02:03:12context仅原1页/owned新页0/blank0，密码global清空/无secret文件。close后工具snapshot page-not-found由独立pages0证实实际关闭。仅C区新finally脚本一patch/Nodecheck0/editor0及必要账本；212业务Go逐SHA不变，无业务build/6144回放。工具selector/keepalive/必填stuFlag/URL回调错误保留，不归业务根因。
- [完整限定证据](management-traits-four-real-verification-20261003.md#L3)、[机器阶段状态](../scripts/test/results/mng-ui-natural-native-20261006/summary.json)、[fresh真实RED](../scripts/test/results/mng-ui-natural-native-20261006/ui-business-red.json)、[独立final](../scripts/test/results/mng-ui-natural-native-20261006/final-attempt1.json)。下一由后端/前端负责者修安全管理闭环后另获明确staging部署批准；本轮不替换binary或主restart，不把PARTIAL写完整PASS。

## 2026-10-05T15:11:53Z 当前源码本地 Go build/regression PASS（旧21项未归因保留）

- 本轮未验：前端、E2E、真实MySQL/LibreOffice、race/coverage、staging/production及旧新语义归因；不称完整产品PASS。只验证当前本机源码，没有修复、回滚、格式化、Git变更、SSH/browser或部署。
- Go1.26.2 windows/amd64/CGO0；100个测试文件实际读取23个环境键，进程中均未设置。FB185_MYSQL_DSN/MNG_GUARD_AUDIT_MYSQL_DSN/MNG_SOURCE_LOCK_MYSQL_DSN在Process/User/Machine均不存在；原生子进程清除全部测试开关/产物路径，未改用户环境或.env。真实DB/staging opt-in未启用，连接观察非回环0仅为单时点，不冒称全程抓包。
- 既有Go Build任务正常完成但接口未暴露数字exit；补充同命令独立服务构建实采exit0/stdoutstderr各0，同Windows PE产物50588160bytes/SHA459668b86c8f1992e6ac94826d115b2cfcf9663d3fbe51431545f1dfde18dc51。编辑器1474pass/0fail仅发现范围；随后默认全仓JSON运行一次，6144pass/652通过顶层/0fail/9skip（8顶层＋1子项），10包PASS/9包无测试SKIP，parse0/exit0；go build ./...及go vet ./...各exit0/stdoutstderr0。与历史6144/652/9计数差量全0，非覆盖率。
- 新current快照212/212逐SHA不变，其中漂移21/21稳定，聚合d1b5ecd6a45d6092fcb3bb5c6638af9d100f755bda542d233bbe578b06ee06b6；go.mod/go.sum/tasks/.env与两个旧基线文件也保持。旧212仅191匹配、旧21仍0匹配，历史FAIL_UNATTRIBUTED21不重置；本轮源码diagnostics0。
- 只新增ignored C区运行收据并更新三既有文档，未新增永久验证脚本。任意测试Output不落盘/不回显，保留非Output JSON事件、原stdout字节SHA及安全skip原因，独立再次计数吻合。[本轮verdict](../scripts/test/results/mng-current-local-20261005-4b6a7edf5d4a/verdict.json)、[起始SHA](../scripts/test/results/mng-current-local-20261005-4b6a7edf5d4a/source-start.json)、[结束SHA](../scripts/test/results/mng-current-local-20261005-4b6a7edf5d4a/source-end.json)、[完整范围与九skip](management-traits-four-real-verification-20261003.md#L3)。完成仅当前local验证，不自动续作或部署。

## 2026-10-05 源码漂移只读复核：21项稳定，差异性质仍未确定

- 对既有漂移清单中的21个文件重新计算SHA，21/21与漂移记录的currentSHA完全一致；本次未观察到继续变化。7个非测试源码、14个测试源码；未修改、回滚或编译这些文件，未访问远端。
- 逐文件将当前字节构造LF、CRLF以及各自UTF-8 BOM变体，对比原beforeSHA：0/21匹配。因此不能将全部差异归结为纯统一换行/BOM转换；混合换行、格式调整或语义差异均未定，不能据此认定行为变化。
- Git HEAD与索引字节及其LF/CRLF变体对比原beforeSHA同样0/21匹配。旧SHA清单不包含旧源码字节，Git HEAD不能冒充已验证源码；尚未取得精确匹配旧SHA的文本，无法给出可信的旧新语义差异或修改者归因。原基线及失败证据保留，不重置源码门禁。记录入口为[21项漂移清单](../scripts/test/results/mng-premature-timeout-c90f8a172e64/local-source-drift.json)。

## 2026-10-05T14:34:36Z 唯一提前自声明timeout主HTTP边界PASS，清理0（整体PARTIAL）

- **[限定纠正 - 14:38:12Z] 本地全SHA收口失败1次，不能称当前212全部不变。** 14:33:52实际212匹配后，21个Go文件mtime14:34:16–18发生SHA漂移，本任务没有编辑它们，来源未归因/未回滚；原失败及[完整漂移清单](../scripts/test/results/mng-premature-timeout-c90f8a172e64/local-source-drift.json)保留。下文212不变仅14:33:52的已核历史。新增[14:38:12远端只读复核](../scripts/test/results/mng-premature-timeout-c90f8a172e64/remote-after-local-drift.json)exit0，main11/private0/旧465/current12/source/runtime/Schema/cache/PID2002/三健康全保持；单项远端PASS不因本地漂移伪称源码总门禁PASS，不部署或编译这些外部变化。
- **仍未验：主systemd重启、自然25min主HTTP到期/不完整/竞争、完整参与者UI、native下载及formal/production。** 本次仅已批准单项；不重放140save/报告/atomic/Worker，不部署、restart、DDL、业务代码或共享配置/权限/cache修改。没有业务build，不称compiled。
- 原page7d2235c5…清page/context routes；14:32:30真实getInfo200/code200/admin/wildcard，再正常admin唯一MTHc90f8a172e64/00201tester配置(isOpen2/state0/name/25min)、正常新增及正确密码登录五键、freeze/组卷/有效paper凭据detail200。密钥/随机密码仅页内内存；没有伪JWT/Redis/reset/token文件。已授权的SSH strict liming/knownkey/Conn10首次0、无retry；fresh主11逐0/15FK/private0/旧465/source4/current12/runtime/cache与最近基线同，完整backup9ebee6a…gzipSHA复核后复用。
- **仅1次POST participant/submit**，14:33:02.576–.611Z，真实HTTP400/code400/body“仅允许manual提交，超时由服务器判定”/successfalse/datanull。请求完整paperId＋timeout/实际X-Management-Traits-Token，无adminBearer；同凭据此前真实detail验证purpose及归属，不把无凭据或缺参数400当PASS。实际handler先拒submitType、尚未进入participant解析；这是有效凭据请求的manual-only入口拒绝，不声称400路径再次校验token，也不是0答manual409。
- SQL前后serverEpoch1791210764/1791210798，原started/frozen1791210751、deadline1791212251，余1487/1453秒；25min未到期。paperstate1/owner.endNULL、140快照/140legacy题/700桶、answered/submitted/checked各0；run/dim/module/receipt/revision/current/audit全0→0，八组整行SHA完全同，profile/冻结时间/原deadline/答案/人员无变化，private0。完整[HTTP浏览器receipt](../scripts/test/results/mng-premature-timeout-c90f8a172e64/http-browser-receipt.json)、[前SQL](../scripts/test/results/mng-premature-timeout-c90f8a172e64/before-attempt1.json)、[后SQL](../scripts/test/results/mng-premature-timeout-c90f8a172e64/after-attempt1.json)。
- 14:33:22原FK精确清理器仅内存收窄4→1/单tester标题/count1，SafeUpdate1/FK开启/SESSION64MiB/owned共享引用0，commit0。14:34:36独立exam1791210750722429072/tester1791210750804032163/paper4f6b5fd3-063c-439c-affc-053d342fb75f各PK及全部child0、pending0；无PDF/上传创建，不删文件。主11逐0/private0/15FK、主78/70/1488/134494/295328/1349/27、旧465及源/当前12/runtime/Schema/metadata/cache全同、PID2002/3active/healthok/backend8fb264e…保持。
- 14:33:56原home/session/getInfo200/admin/wildcard保留，ownedGlobals absent/participantStorage0/监听器新增0/新页0。仅新增C区[专用证据入口](../scripts/test/management-traits-premature-timeout-evidence.js)与ignored结果、三个必要文档；Nodecheck0/editor0，212业务Go逐SHA不变；模板插值冲突在远端执行前纠正，helper两patch，没有业务RED/GREEN修复授权或源代码改动。
- [限定纠正 - 2026-10-05] 下条14:23历史401及businessRequests0完整保留；本次正常认证和真实唯一HTTP400＋SQL零写已关闭该单项，不关闭其他产品/正式门禁。[完整限定证据](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-05T14:23:18Z 主HTTP最小欠项已定位；正常会话401，未创建记录

- **BLOCKED_AUTH/产品PARTIAL**：清原page/context routes后，真实getInfo14:22:06及14:23:18均HTTP401/code401；cookie/store相等、既有凭据/首页保留不等于有效认证。末次adminfalse/wildcardfalse、参与者storage0/publichealth200。浏览器response事件数组未捕获，不声称事件观察PASS；实际fetch返回及console401保留。无login/logout/reset/伪JWT/Redis/凭据文件或新page。
- 完整报告已包含未完整manual409（10-03/10-04）及真实adminresume自然5min等历史，故不重跑。选定另一已设计最小边界：有效owned参与者未到期时POST participant/submit自行声明submitType=timeout；真实handler合同400/code400/“仅允许manual提交，超时由服务器判定”/successfalse/datanull。**本轮业务请求0、HTTP400及SQL差量未验**，不把源码合同或401当该业务PASS。
- SSH首次1/无retry/exit0，14:23:05fresh只读：主11逐0/15FK/private0、78/70/1488/134494/295328/1349/27，current12/source4/旧465/runtime/资产metadata与最新bf051325…after全同；cache target/full与最新及approved-after同/rootroot755。PID2002/backend8fb264e…/front98547…保持、三服务active/healthok。仅流式cmp，无远端证据文件/备份/恢复/DML/DDL/部署/主restart/共享配置权限操作。
- 新资源0/cleanup_required0，**不是执行清理PASS**；原fail与报告/atomic/Worker旧case不重放、fixture预算不动、业务源码0改，无需build全套。下一只需用户在原页正常登录后真实getInfo200/admin，再freshbaseline+exactcleanup能力门禁后执行这一单独owned HTTP边界；主systemd停机需另明确批准，完整UI/natural主HTTPexpiry/formal-prod仍未验。[脱敏receipt](../scripts/test/results/mng-http-boundary-20261005-142318/receipt.json)、[精确scope](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-05T14:14:25Z 新有界观察层1轮GREEN：atomic＋恢复库Worker尾部完成，产品仍PARTIAL

- 主HTTP acceptance/主systemd重启/fullUI/coverage/formal-prod仍未测；此前50/zero报告与独立DataSHA不重跑。新授权fixture1/3、runner0/3，原context210/外test240/child30不变，旧3轮FAIL及37.33秒RED保留；限定todo2完成，不改workflow状态。
- 真实根因：runtime构造gorm.Open在同pool新registry，外层callbacks无效；构造显式保留Logger，GORM processor及Raw.Scan均向其Trace交真实error。仅C区fixture改无参数/mutex/context精确Trace，typed1644＋45000＋恰1moduleINSERT门禁保留；Worker真实首扫及PK事务结束屏障同步修正，watcher defer清空，无callback注册/业务代码变更。
- 唯一restore bf05132543452c00/positive_app：140保存34645ms，真实midmodule1644/45000；4表增量0、paperstate1/submitted0、独立ownerNULL、答案/桶/frozen/profile/person完整facts SHA20b372b4…相同。同卷正常retry completed140/score50/1run13dim4mod1receipt，duplicate reused/零增量/facts不变。原NULL标量负向仍sql_null_to_time，nullable struct实际GREEN。
- Worker真实独立child PID28610首扫未到期→有意Kill/Wait→自然跨原deadline1791209521→新PID28616首扫1023ms/父核2279ms/all expired settled；140completed、0/3incomplete正式NULL/no-render，各SQL1/13/4/1，LoadValidatedRun receipt关联复核、deadline不重置。真实RunExpiry/manual独立2pool/同PK锁竞争唯一timeout1/13/4/1；Worker锁已观察，事务结束屏障后cancel/join，重复同run；非主HTTP/systemd。
- 父实际1pass0fail0skip/125.76秒/exit0，五case标记PASS；secondchild1pass/1.03秒，firstchild有意被杀不计PASS。fresh Node/contract/test-c/bootstrapbuild/两vet六exit0、editor0，finally恢复原Go env；fixtureSHAc86739a1…/testbin d792e823…，runner/bootstrap/restored脚本原样；212业务Go逐SHA不变/aggregate33b04187…。
- fresh14:07:18SSH首次0/无retry，current备份9ebee6a…gzipSHA通过后复用；最终14:14:25只读0/restore-trigger-grants-upload0/two child PIDs gone、主11逐0/private0/15FK、主78/70/1488/134494/295328/1349/27、465/current12/source4/runtime/资产metadata保持/PID2002/3activehealth/backend8fb264e…不变。cache0:0:755/targetstat本轮前后同、整树stat匹配approved permission-after；freshbefore仅target，不冒称新增整树before。正式备份/原失败/新GREEN root0700/files0600永久保留。
- 经验：独立gorm.Open不共享callback registry；测试只读观察可走显式保留的Logger.Trace，Raw.Scan用Recorder后还会调用currentLogger。不得从父registry callback缺失推断SQL未执行。取消Worker前用exact-PK事务完成屏障，不用run已存在推断Worker已结束。[完整scope及准确行号](management-traits-four-real-verification-20261003.md#L3)、[GREEN](../scripts/test/results/mng-lifecycle-remaining-bf05132543452c00/transport-6-execute.json)、[当前终验](../scripts/test/results/mng-observer-scope-20261005-140715/final-readonly.json)、[机器summary](../scripts/test/results/mng-observer-scope-20261005-140715/summary.json)。

## 2026-10-05T13:58:26Z atomic/Worker新接续：原NULL扫描cause确证；最终夹具FAIL、3轮停止

- SSH首次exit0/无retry；fresh主11逐0/private0/15FK/PID2002/3activehealth、backend8fb264e…/cache0:0:755；完整备份9ebee6a…gzipSHA通过。current12表051795bd…/source259d817a…/465逐路径/完整runtime匹配，不重跑已PASS HTTPzero/DataSHA或浏览器。
- 第1轮仅安全观测并保留原失败：唯一restore e5e7a1c701a3340a，sql.NullTime独立真实NULL，原Scan(**time.Time)为*fmt.wrapError/allowlisted sql_null_to_time/contextnone，4.01s/exit1。原unknown已限定确证为测试NULL扫描，非身份service/业务bug；旧210.08秒save cause仍未证。[真实RED](../scripts/test/results/mng-lifecycle-remaining-e5e7a1c701a3340a/safe-failure-evidence.json)。
- nullable结构修正及140完整atomic/Worker增强共用新3轮夹具预算（观测1/修正2/扫描门禁3），最终3/3、runner1/3；不第四轮/副本/删断言/增210或240。final restore51fba58578d2f855真实fill140/34411ms，通过失败提交到新增errno1644观测门禁未满足，37.33s/exit1、0pass1fail0skip；GREEN0、正常retry/完整facts/Worker均未执行，child0。[最终FAIL](../scripts/test/results/mng-lifecycle-remaining-51fba58578d2f855/safe-failure-evidence.json)。
- 新测试观测踩坑：RuntimeService构造重新gorm.Open同pool但独立callback registry，观察挂外层db不能观测实际s.db的Create/Row；本次driver观察失效，future child/race latch亦同缺口。source已证registry不同，但实际事务errno未取得，不把runtime_invalid称触发器或业务bug实证。三轮上限停止，后续需新有界scope修真正观察层，不能复用现fixture称Worker可运行。
- finalNodecheck/contract/test-c/bootstrapbuild/两vet各0，GOOS/GOARCH/CGO按原值finally恢复、两diagnostics0；无6144/front/race/coverage新证据。两恢复库/trigger/grants/uploads0/cleanup errors0，13:58:26终验exit0/11逐0/private0/465+source4+current12+runtime/cache完整stat保持、PID2002/3activehealth；正式backup/失败manifest/root0700/files0600保留，无mainDML/deploy/restart/共享配置权限/production。[新终验](../scripts/test/results/mng-atomic-worker-20261005-134523/final-readonly.json)。
- 仅C区fixture/runner及必要账本；runner移除假当前401、不重开历史HTTPzero/DataSHA。currentoverall PARTIAL/FAIL，atomic完整/Worker与主HTTP acceptance仍未验，不完成workflow。工具清单形状cmp1及PS中文native路径ENOENT1保留，分别只读纠正0；客户三原件实际SHA不变。完整[范围/新旧cause/停止交接](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-05T13:32:39Z atomic/Worker新授权接续：SSH两次连接前timeout，未执行修正

- 完整读取本记忆/最新四真实报告、生命周期fixture/runner/bootstrap/restored/final脚本、runtime-validation及相关规则；两历史receipt目录全部JSON读取/解析。新scope仅atomic＋恢复库Worker，测试两文件各最多3轮新预算，实际修改0/使用0，context210/外限240不变，不重复询问已有授权。
- StrictHostKeyChecking=yes/BatchMode=yes/ConnectTimeout10/ConnectionAttempts1、既有liming密钥，仅20.200.136.133 staging。fresh只读基线初试及唯一重试均连接前timeout/exit255/远端stdout空，脚本未启动；两次上限后停止，不HTTP绕行。当前11表/private/restore/GRANT/465/源/12表/备份SHA/cache/3health/PID均未本轮复核，08:31:50是最后已核历史，不回填当前PASS。
- 原unknown调用点实际为fixture直接Table(el_candidate).Select(end_time).Scan(&endTime)，目的类型**time.Time；不是TryRegisterCandidateIdentity或service身份读取。已读锁定GORM1.25.12 Scan/ScanRows源码，但未取得原error类型/driver errno/本轮RED，故NULL扫描、锁、context、trigger残留等原因全部未证，不能据源码猜测修正。旧rollback部分零写及ctxnone实证保留，完整owner验证/retry/Worker仍未完成。
- 新恢复库/上传/触发器/GRANT/子Worker进程0，cleanup_required0，不称远端cleanup0。业务/fixture/runner/预算/共享配置权限/主DDL/deploy/mainrestart/浏览器/报告/production均无操作；build/6144/front/race未重跑。仅C区失败receipt及五既有账本追加，[本轮失败证据](../scripts/test/results/mng-atomic-worker-20261005-133239/ssh-readonly.json)，整体BLOCKED，不更新workflow完成态。SSH恢复后由测试执行者继续fresh只读→安全cause→有效RED→最小fixture修正，不要求重复授权。

## 2026-10-05T08:31:50Z 唯一正常HTTP零分报告PASS／独立DataSHA新样本PASS／精确清理0

- 原同host page7d2235c5…清page/context mocks后08:21:40真实getInfo200/admin1/wildcard/cookieStoreMatch，不沿历史401跳过。prefix browserPage工具not-found改用实际ID，无新admin/reset/JWT/Redis/logout/tokenfile。finally08:31:50原index/getInfo200，owned密码/参与者token全局清空，participant storage不存在、新page0。
- 仅MTH4f13a5e47af1/00201tester一exam。actual封闭isOpen2（旧交接0不合法），legacy/join1/state0/25min/name；正常owned新增在freeze前/正确密码五键；无paper前提freeze允许已有人员。actual manifest方向＋sourceOptionId/raw选择100正raw1/40反raw5，140HTTPsave/readback/manualcompleted/1run13dim4mod1receipt、每题final1/13维4模块非NULL精确0，独立Fraction/常模/receiptPASS；源V67/V96及版本不改。
- 唯一UIgenerate确认08:26:02.421–08:26:08.471Z HTTP200/code0/attempt1/revision1/createdBy1；UIview/download各200/applicationpdf/%PDF/660779bytes，SHA **e8c74497fa0580544e3b2a74f480a0ab76b717d889eafa7e9585a9e4fd026c2d**与DB/private0600/独立scp精确同，1generate+2读取download审计。native事件20秒未捕获，不称文件落盘；整个参与者UI未测，内部DOCX/OPC未重capture，不重跑四卷。
- 新唯一report persisted data_snapshot原UTF8 9790bytes由既有SQL器无损读取，Node crypto与Python hashlib独立SHA **0198cffd460892ed3edfff03d73ff368ff03fb89458809ce390d2719146e6df9**匹配DB/HTTPDTO，未重序列化、不改json:"-"或API、不拿服务自check当独立证明。hash只证明字节完整性，算术/文本另验证；旧四恢复文件UNVERIFIED不追溯改写。既有函数内存复用exit0/LO24.2九页/8嵌入字体/五灰环各360/360-whitehole1.0-blue0/五0.00-12/10pt/36客户全文PASS，第三页contact实看，非全PDF几何或OPC新证据。
- fresh baseline12表051795bd…/465旧PDF/运行资产与最近http_772ccd9f3e56完全一致，fullbackup9ebee6a…gzip/SHA实核后复用；新http_4f13a5e47af1 evidence受限保存currentbefore/source/cache，不重复全backup/restoremain。exactcleanup仅内存4→1/tester标题，SafeUpdate1/FK开启/range64MiB/主键闭包，08:29:50exit0/所有owned根0/11表逐0/private0，只删除新SHA PDF；正式备份全保留。
- 首次终验exit1仅本轮owned pre-generate.cursor0644，采游标命令未设umask077；确认exactroot文件无symlink后只收紧0600，sharedpermissions0，失败receipt保留。08:31:27新终验exit0/证据0700-0600errors0/manifestb22d768d…；主78/70/1488/134494/295328/1349/27、15FK、465/source4/12表/backend-front-config-template-unit/cache完整stat相同，PID2002/backend8fb264e…/cache0:0:755/3activehealthok/实际两reportstaging，新增reportfailure0。
- 当前todo1 authenticatedHTTPzero＋DataSHA＋cleanup **completed**；todo2 atomic完整/Worker **未完成且三轮停止**，rollback participantunknown/ctxnone需新有界授权但本轮不问、不改fixture/budget/代码/部署/主restart/共享权限/production。仅C区合成证据与四既有账本更新，不重跑6144/Go编译/front/coverage、不修改workflow状态或称wholeDONE。[完整本轮与失败保留](management-traits-four-real-verification-20261003.md#L3)。
- [限定纠正 - 2026-10-05] 下条08:15的当前HTTP401由本次正常有效会话和新唯一零分链限定解除，DataSHA欠项只对本新report关闭；旧恢复样本/atomicWorker/正式产品批准保留原边界。后续证据游标同样必须umask077，不能仅主baseline脚本设过就假定另一sudo shell继承。

## 2026-10-05T08:15:43Z SSH恢复／七项生命周期服务PASS；atomic夹具unknown停止，非产品DONE

- 正常admin08:13:29Z实际getInfo401/code401/credential0；旧page不存在后同host复用，不login/logout/reset/JWT伪签/Redis/凭据文件。HTTPzero、PDF/view/download/UI本轮SKIP；原四文件oracleGREEN/144段仅历史，独立DataSHA仍UNVERIFIED。无deploy/权限/业务/主DDL/oracle/production变化，主PID2002及8fb264e…保持。
- current完整备份http_772ccd9f3e56 SHA9ebee6a…与当前12表/465基线无漂移，复核gzip/SHA后复用，不反复全backup。两个唯一恢复库1ac0b50628a4d637/0ca26bf185d9b25f：crossschema0、既存11表完整gate、真实CURRENT_USER positive_app。临时精确库GRANT五权限最终REVOKE0；root只恢复/独立trigger DDL，不冒应用验证，未授SUPER/改global binlog。
- 首轮优先剩余case：到期140保存34968ms/ctxnone、独立双pool manual与Worker可信Submit一次created一次reused/同run1dim13mod4receipt1/score50timeout；到期0/3incomplete/正式NULL/render0/expired拒、<=300s真实resume及purpose/crossIDs/签名解析expiry、exam.end已有deadline不变/新拒、retired已有继续/新拒、revoked答案增量0/render0/revision0均PASS。只是服务集成，不是主HTTP/RunExpiry扫描竞争/自然5min/UI。
- 首轮50.42秒exit1，stage owned failure trigger/mysql1419/contextnone；终验log_bin1/trust_function_creators0。只改独立建夹具，第二轮不重复上述case，root预置精确synthetic snapshot marker/run PK闭包trigger。实际失败提交及SQLrun/dim/mod/receipt零增量、paperstate1、submitted0通过；candidate.end_time读取stage rollback participant/unknown/contextnone/3895ms，整体3.90秒exit1，原因UNVERIFIED，不猜NULL/GORM。fixture三次编辑上限达，停止不增budget/改业务。后续retry、完整atomic、RunExpiry进程断电重启/主systemd/HTTP均未执行，新WorkerPID0。
- 原context210秒/外限240秒未改变；原210.08秒actualsave缺cause仍UNVERIFIED，本轮35秒全答不能回推历史timeout。fresh专项test/bootstrap build/vet四exit0/editor0，未重跑6144/全量或前端。安全stage/class/context合同先RED→GREEN；未知Error不格式化，literal无JWT/凭据/BOM。
- 两finally库DROP0/GRANT0/精确三上传项与output0、cleanup errors0；永久root0700/files0600 evidence lifecycle_1ac0…/0ca26…全部manifest/gzip/SHA复核。08:15:43Z终验SSHexit0：11表逐0/15FK/private0/旧465逐SHA、当前12表051795bd…/source4表/backend/front/config/templates/unit/权限及sharedcachestat全相同，主78/70/1488/134494/295328/1349/27，PID2002/3active/healthok、report staging/APPENVproduction。其他active0/expiredcompetency0不是WorkerPASS；原备份与失败保留。
- 完整逐case/完整SHA/范围与后续见[最新四真实报告](management-traits-four-real-verification-20261003.md#L3)。当前整体FAIL/PARTIAL，不完成父workflow任务；下一先安全查证rollback participant unknown并获新有界夹具授权，再atomic/Worker尾部，正常getInfo200才唯一HTTPzero/adminresume/UI。工具只读引号1064、安全regex假阳性、网页抓取失败均保留，不归业务bug。下方SSH阻断由本条限定解除，HTTP/atomic/Worker未解除。

## 2026-10-05 新有界oracle切片完成；HTTP401/SSH双timeout，生命周期仍BLOCKED

- 本次新明确**1轮**修正已用完；仅综合zero-restored验证器chartScore/chartNorm十二位期望改format(decimal,'.12f')，不改Fraction/Decimal比较/等级/阈值或业务。原未修改完整REDexit1/首份50PASS后零fail；精确RED0E-12对真实Go0.000000000000、数值相等而词法FAIL。唯一patch后完整4case/0error/exit0/RESTORED_PASS_NOT_HTTP，四份Fraction/13维4模块/140与40反一次/OPC/六图/每份36原文共144全部PASS。原三轮失败receipt不覆盖。
- 独立输出scripts/test/results/mng-bounded-20261005-aa67b4600284仅九个既有合成文件逐SHA复制；RED44c900f5…/GREEN019f7530…保留。零分旧服务器LO24.2 PDF660827bytes/SHA53a6993bc642ac9ebb9befd59f3f22b1d0116e25678e0b1161b7a9a6fb8c5bb8/A4九页，五灰环360/360/whitehole1.0/blue0，五0.00总体12pt/模块10pt/bbox保留，第三页实看。四OPC各75/68原字节，非值XML/star/footer保持；全PDF几何及crossengine不夸全等。**DataSHA仍UNVERIFIED_MODEL_JSON_OMITS_SNAPSHOT**，不拿服务自check当独立验证。
- 指定samepage dc5c14cb…首页保留、page/context routes清空，真实Admin-Token/store相等；07:39:20Z实际getInfo401/code401，admin/wildcardfalse。无login/logout/reset/JWT/Redis/secretfiles/参与者storage，未创建HTTP合成实体。真实零分HTTP/报告/view/download/UI均SKIP，不以旧恢复actor1替代认证。
- 既有liming/knownkey/Strict/Batch/Conn10/Attempts1只读诊断两次连接前timeout/各255，远端shell/SQL未启动；到此连接预算停止。当前backend/PID/3health/11表/465/源/非ownbaseline/共享0755/备份SHA未复核，02:48:56Z仅最后已核历史。无新备份/restore/Worker/mainrestart/upload/DML/权限/业务patch/deploy/DDL/production，新资源0→cleanup_required0，**不是远端cleanup0**。
- 本地原test.logSHA0e32188035f946a16bb94e65b9bf26f12b40c2c86075aebc29b0115d0d0305c8只有四PDF/到期0与3/actualsave210.08sFAIL，无contextcause或phase时间。原共享ctx210/外限240已读但不能据耗时断定timeout，故不擅改预算或重跑3分钟，fixture本轮0改。expiryfull/race/end/retired/revoked/resume/offlineWorker/failureatomic本轮SKIP；旧0/3及普通mixed只历史。
- 选定venv3.14.4/PyMuPDF1.28.2/openpyxl3.1.5，AST0/实际四文件GREEN0/editor0；Pylance MCPloadhooknull失败保留。仅C区Python＋必要docs/账本，未Go编译/全量。oracle切片COMPLETE，整体PARTIAL/BLOCKED；下一需正常getInfo200＋SSH恢复后current backup/baseline及最少HTTPzero/finally，生命周期先安全cause/phase取证。详见[本轮完整证据](management-traits-four-real-verification-20261003.md#L3)。
- [限定纠正 - 2026-10-05] 历史三轮综合oracleFAIL已被新明确一次patch四文件GREEN限定解除，旧失败不删；历史HTTP50与主基线PASS不代表本轮401/SSHtimeout下的新验证。正式内容/production批准未解除。

## 2026-10-05T02:48:56Z A单目录权限APPLIED_VERIFIED；唯一正常HTTP报告PASS，finally0

- 用户结构化明确批准A，仅/var/spool/libreoffice/uno_packages/cache/uno_packages rootroot0700→0755。fresh realpath/no-symlink/空目录/owner-mode/父stat/前shared快照及Perlaccess-default四ENODATA全PASS；原metadata/SHA保存在root0700/files0600。共享chmod1目录/0文件/chown0/递归0/新ACL0/非rootwrite0，父cache仍755且limingw0。两转换均通过，rollback未需要；ctimeintentional不可恢复。改前EACCES→改后原shared路径limingopendir成功，已知读取遍历阻断解除；原stderr直接essential路径的严格历史因果标签仍UNVERIFIED，不将所有DeploymentException归同一原因。
- 同纯syntheticDOCXdadc9610…、原实际LO参数及独立profile，输入由liming自己创建0700workspace/0600DOCX，不把root证据目录当input；HOME/XDG/PATH/LANG/TMPallowlist与matchingtransient有效serviceprops。normal02:40:16–21/matched02:40:22–27各一次90sec，**defaultshared0override/rootLOruns0，两exit0/PDF552631/9非空/7嵌入字体/TEST**；PDFSHA8e9c8fff…/67f43760…，textSHA2333523b…相同，不强PDF字节同SHA。两stderr68bytes/5dacd356…、DeploymentException/permission/generic类别0。
- 权限永久证据lo_permission_20261005_TgMkLSzP（正式parent下），初始manifest592c3ced…/独立final9462113f…，原metadata及manifestSHA校验0/permissionserrors0。除了目标mode/ctime，cache完整快照相同；HTTP后仍与permissionafter完全相同、所有父stat不变/stamp及newsharedfiles0；不称整个LO系统零写入。
- 原page不存在后同host复用dc5c14cb…；起初actualgetInfo401/credential0，02:41后actual200/admin1/wildcard/cookieStoreMatch/home。代理未login/logout/reset/伪JWT/Redis/凭据文件，不推断会话恢复原因。auth恢复后新current完整singletransaction/routines/triggers/events/gzip备份http_772ccd9f3e56 root0700/files0600，DBSHA9ebee6a155ca17c7814085d7f839c8c1d59f2ef9ee3baa8dd5ea5f9f7902241f，永久保留。
- 唯一**MTH772ccd9f3e56/00201candidate/name**正常配置/freeze140700/匿名五键身份＋四typed空串/25-20时间/140HTTPraw3保存readback/manualcompleted/正常adminresult PASS。实际SQL140raw3final3/overall50/1run13dim4mod1receipt。唯一generate02:45:41.352–47.267Z **HTTP200/code0/5914.7ms/attempt1/revision1**；新游标后reportfailure0/JSONparse0。normaladminview/download各200/%PDF，与DB/liming0600/独立下载副本相同SHA **c4239113ffb9763bcaf6621973cad5f71173e8a6e94ca3ec13dd25d8ad6cfd9e**、648199bytes。
- 原稳定四卷SQL器/oracle仅内存收窄首份all50=1份，原源码SHA不改：独立Fraction13/4/140/40反一次/常模/receipt/dataSHA/客户36段全部PASS；LO24.2A4九页全非空，5蓝半环＋13维柱线图/第三页/九页contact实看。**不是零分灰环/四组合144段/全UI/native下载/完整产品PASS**；zero-restored三轮FAIL及210sec生命周期夹具/expiryWorkerrestart-race原样，未重跑6144/build/deploy/主restart/生产。
- exact exam1791168237804362417/candidatea716a011-f178-4bbe-9ac5-1eced9fcf7c5/paperdb74ee5d-85fd-43e5-a0f9-1cd64a06f530/run6e721e25-8e24-40a9-b4d4-63de0bc875aa/bundle36f24809-63a7-4cb1-b847-f07642959fe3/reportaf707807-6b4c-432b-be57-704b0ef7c565。原HTTPcleanup仅内存4→1/titlecount1，SafeUpdate1/FK开启/SESSIONrange64MiB/PK闭包；02:48:09cleanup0，仅新exactSHA私有PDF删除。02:48:56finalSSH0：11表逐0/owned全0/private0/LOlive0/phase1tmp0/owneduploadpayload0/transientnot-found；主78/70/1488/134494/295328/1349/27及15FK，旧12SHA051795bd…/源4表/465逐PDFSHA/所有运行资产前后同。PID2002/8fb264e…/两reportstaging/APPENVproduction保持，3activehealthok/general001002003MBTI004五只读HTTP业务PASS。
- HTTPfinalmanifest9a450ae7…/backupgzipSHA复核及证据0700/0600errors0；正常admin最后getInfo200/home/cookieStoreMatch保留、本轮testglobal清空/sessionStorage写0。C区两新single-scopeprobe/SSHtransport、bash-n0/Nodecheck0/diagnostics0，无新framework/旧脚本改写。工具错误保留：两Node嵌套syntax在SSH前、Perlreadonlyliteral在chmod前、browser错token字段只内存纠正不重复登记、dynamicimport在view/download前失败后内存SHA自检＋serverSHA复核。完整SHA/新合成receipt/PDF/SQL/图片/限定结论见[四真实报告顶部](management-traits-four-real-verification-20261003.md#L3)。权限切片完成后停止，不重复授权/部署。
- [限定纠正 - 2026-10-05] 下条A待授权/RX未验及历史普通用户LO失败已由本轮明确批准、default两转换与唯一真实HTTP报告PASS限定解除；不是四组合或生命周期全验，旧失败及严格stderr因果门槛保留。

## 2026-10-05T02:31:16Z LO最小权限方案只读核定（未repair；待精确选择）

- 两次有界SSH只读exit0，未运行LO/root基线/新probe/HTTP写、未chmod/chown/setfacl/配置改/部署/restart。当前cache整树仅父root0755与子root0700两个目录、文件/包/插件/registry/stamp0；全部父0755，子limingr0/x0/w0。与上一cache-after-targeted完整metadata/SHA snapshot cmp SAME；不称历史root执行零系统写入。
- getfacl/setfacl/getfattr均缺，未install；已安装Perl SYS_getxattr只读核父/子各access/default ACL全部ENODATA，确无扩展ACL。必要系统bootstrap/registry rdb root0644/soffice.bin0755均limingread1，无需文件权限变化。动态子不属dpkg/exactstatoverride0，core/common4:24.2.7-0ubuntu0.24.04.6限定四maintainerscript cache引用0，包默认子mode仍未核。
- 候选A仅exact /var/spool/libreoffice/uno_packages/cache/uno_packages 0700→0755、UID/GID0保持，目录1/文件0，不递归/不给父w；可解除已知读取遍历阻断，但增加group/otherrx不是主体最小。回滚仅0700，ctime不可恢复，执行前需重核空/owner/mode/ACL漂移即stop。B仅limingrx ACL主体最小但工具缺/需单独批准、mask会改变numericgroupclass；C每公共Client子进程owned0700workspace/shared-cache加-env:UNO_SHARED_PACKAGES_CACHE=fileURI，不改shared/systemd，需独立代码测试构建/受限发布授权。
- 已保存roottrace父stamp两O_RDWR|CREAT|EXCL成功、limingEACCES，不能证明stamp必需或可忽略，**RX充分性未验**；不授父cachew。上轮C单变量成功支持候选，UserInstallation已有隔离不替shared cache；严格stderr路径仍0/rootcauseUNVERIFIED，未HTTPGREEN。002/00401共公共Client、MBTI三调用独立exec同liming，A/B全共享影响；C仅公共Client，MBTI不自动覆盖。未知系统定时/未来消费者未穷尽，无全系统扫描。
- 02:31:16Z当前PID2002/backend8fb264e…不变、3svcactive/healthok/private0/LOlive0/sharedsnapshotSAME。11表/owned0/source/465旧PDF/旧12表SHA沿02:12:56Z历史，本轮没SQL/逐PDF复核；buildtest未跑，业务源码0。仅两个既有C区docs回写，精确方案/回滚/消费方表见四真实报告新节。**等待A/B/C明确选择及exact共享operation批准，泛“继续”不授权permissions/HTTPwrites**。

## 2026-10-05T02:12:56Z 有界LO syscall三次完成：共享缓存EACCES实证，严格根因UNVERIFIED／未修

- 本轮只诊断，root／liming／最后command-local缓存redirect三个syntheticLO退出 **0／1／0**，PDF552631／0／552631bytes；inputdadc9610…同SHA，75秒＋5秒grace／16MiBcap均未触，core禁用、只file/process syscall、无read/write缓冲、未attach主PID。失败stderr195bytes／b6ecc0ff…与前轮完全相同；stderr明确资源path0，未达到用户“stderr指同一essential资源”严格CONFIRMED门槛，**根因标签UNVERIFIED**，不猜Java/HOME/sandbox，不修业务／配置／权限或HTTP重跑。
- 已验证阻断：`/usr/lib/libreoffice/share/uno_packages/cache/uno_packages` root openat O_RDONLY|O_NONBLOCK|O_CLOEXEC|O_DIRECTORY=11，liming同path/flags=-1EACCES；真实symlink目标 **/var/spool/libreoffice/uno_packages/cache/uno_packages root:root0700**，limingread0/traverse0、所有父0755。birth/ctime/mtime2026-10-04T06:24:13.038304998Z早于本轮，不归因给具体历史操作者。stamp.sys的O_RDWR|O_CREAT|O_EXCL root成功、limingEACCES及随后ENOENT不能当独立缺文件根因，也不能以此要求共享父cache可写。
- 公开lounorc/fundamentalrc实际UNO_SHARED_PACKAGES_CACHE绑定已读；最后只对子命令增加`-env:UNO_SHARED_PACKAGES_CACHE=file://<ownedworkspace>/shared-cache`，自己的liming0700目录预检可写，同合成输入／profile及其他环境参数，成功0/PDF552631、DeploymentException0、原shared目录调用0。是诊断对照不是修改运行配置／业务GREEN；相同PDF大小不称byteSHA相同。三次预算耗尽，不再probe。
- **系统路径零写入说法禁止**：root成功基线LO自身短暂两次创建共享stamp.sys，cache父mtime/ctime02:07:42Z更新，finalstamp0；无主动sharedchmod/chown／repair，但root阶段未有完整sharedcachebefore字节基线，不能称整个LO系统树不变。targeted前后完整sharedcache字节及mode/owner/size/mtime/ctimesnapshot相同。动态uno_packages不在dpkg清单、exactstatoverride无条目，父cache/libsymlink属libreoffice-core；core/common4:24.2.7-0ubuntu0.24.04.6，getfacl不存在、未install。未验证动态目录包规定默认mode，不将0755说成已核默认；下一任何该exact目录read/traverse修复必须单独授权，不chmod-R／放宽父可写／改HOME或部署。
- 永久证据 **lo_syscall_20261005_roBjxbvP** 在正式parent下root0700/files0600，rawtrace/stdoutstderr不下载／不打印；manifestSHA **9e1b15b0a225f39c4e0da6e04b2734fdc682e893dd6430ba181ec31b6cad75cd**、校验0／权限errors0。roottracef84d3c6e…533960bytes、liming735df912…129477、target61fd51d3…537048，实际脚本/安全资源receipt/final执行脚本与receipt保留，原8BorrQx7／所有fullbackup永久保留。两个C区小探针bash-n0/editor0，安全解析6assert6PASS（私有/转义/合成secret标记／无root成功可选ENOENT不输出）；不改B区或oracle/budget，不重跑Go全量/build。
- archived final02:12:56Z/SSHexit0：三个exactworkspace0／专属process0／旧payload0／旧transientnot-found，遗留exam1791105048344426522整owned0／11表逐0／private0／15FK。主 **78/70/1488/134494/295328/1349/27**、12表051795bd969650d0e55da676083078255331c66405b5fff0ee4ca9d0a7641d0b、源4表、465旧PDF逐路径SHA、backend/front/config/templates/dropins/unit全本轮前后不变，不抹外部既存漂移。PID2002前后同、healthok／3active／APPENVproduction＋两报告staging／8fb264e…未变；最新fullbackup8da18335…重新SHA/gzipPASS／parent三归档PASS。无DML/session/reset/mainrestart/deploy／production。
- 工具失败保留：Node插值syntax1在SSH前；缺stamp的namei exit1；dpkg postinstgrep无匹配xargs123提前停，后续安全补查0；首packageformatliteral不当版本实读，标准dpkg-query-W纠正。详细资源flags／完整SHA／安全终验和下一授权边界见四真实报告最新节。
- [纠正 - 2026-10-05] 上条“系统LO树不可读0／标准路径可遍历”未跟随share/cache目录symlink，不能证明/var/spool真实共享扩展缓存正常；现追加actual目标EACCES实证。上条底层资源未定位现已限定补充，严格DeploymentException路径关联仍UNVERIFIED，未repair／产品DONE。

## 2026-10-05T01:55:31Z 遗留exact清理PASS；有界LO三环境复现，深层根因仍UNVERIFIED

- **不是业务修复/产品DONE**：仅已授权staging遗留清理＋runtime诊断；三环境各一次LO，root exit0/PDF552631bytes，普通liming及matching transient均exit1/PDF0，stderr各195bytes/同SHAb6ecc0ff320848795c8f84e1e5b82e2c4867d9e80f4ef0305ebb7287514d3d54。固定分类DeploymentException/terminate/Unspecified Application Error实证；原09:12请求具体exitcode未留存，不能回推为1。底层触发原因仍UNVERIFIED，不猜HOME/字体/权限/sandbox，不修改业务/配置/模板/共享权限。
- SSH首次exit0/vm-ubuntu-go-dev/liming；actualREPORT_EFFECTIVE_ENV/MNGstaging、APPENVproduction旧cfg，online8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93/front98547保持。当前PID2002/启动01:36:55Z早于本轮首次01:43:14Z；不冒称昨日PID9059仍在，原因未查，本轮不重启/部署且probe前后PID2002相同。
- 新完整current单事务/routines/triggers/events备份cleanup_20261005_6jpcfVUo root0700/files0600/gzipSHA PASS，DB8da183355cca1c2eb379ff5827abd0f079bbb42b883364d21cc75b03e7783611。首次工具停止前另备份GHU38Ugd/009784afa7f8caa83ba185b874881e983db15d6bd92f0386db1344584833ed2e，两份及所有正式parent永久保留，不恢复main。
- 精确exam1791105048344426522/candidatea3889750-1307-4fed-bc27-e89613d2f518/paperc6d52955-aeec-4abe-8468-c5f350f1df84/runea660428-f2b5-4445-b6e7-9d8b2351e00d/bundle7a20628b-d061-4606-9874-db38aa198f80及marker/createdat真实核验；bundle外ref0。锁内PK/owner/count事务删除13dim4mod1receipt1run140qs1snapshot700桶140pq1candidate1paper1profile1bundle1link1exam，commit/SSHexit0；11侧表逐0/owned全0/revision-current-audit0/privatePDF0。无普通exam-delete绕保护/关闭FK/SafeUpdate/DDL。
- 01:54:44Zfinalreadonly0、01:55:16Zfull12cmp0：主78/70/1488/134494/295328/1349/27，15RESTRICT FK，两source各140140700；非own排除精确owned的12表前后SHA、源4表及465旧PDF逐SHA/运行资产均相同。清理后及probe后完整12表SHA051795bd969650d0e55da676083078255331c66405b5fff0ee4ca9d0a7641d0b匹配09:09写前基线，不删除正常旧卷、不抹更早81aeaa漂移。
- 实际service User/Groupliming、HOME/home/liming、PATH含/usr/bin、LO/usr/bin/libreoffice24.2.7.2；ProtectHome/PrivateTmp/PrivateDevices/NoNewPrivileges/RestrictNamespaces均no、ProtectSystemno、无路径屏蔽。两失败与root同参数/新0700workspace/0600纯syntheticDTO DOCX；固定SHA dadc9610b59be8deb4eabb497c4a7ec5bc5f2b09de0c81d77d3992e4d2583049。同安全环境与受限属性的独立transient匹配，不等于原HTTP进程完整复刻；root也有javaldx警告，不能单独当根因。
- 原09:12:20–24UTC窗口只有loexec分类1；新01:51:45–01:52:05UTCkernel AppArmor/userns/segfault/OOM分类均0；coredumpctl不存在，未取得core实证。标准LO路径liming可读/目录可遍历、系统LO不可读files0，HOME四目录可写；不是权限修复依据。新证据lo_diag_20261005_8BorrQx7/root0700/子目录0700/files0600，原stderr只私有保留、console仅固定类/hash/length。
- 三ownedworkspace清0/上传payload0/transientnot-found，新syntheticPDF不持久化DB，正式备份保留；三服务active/healthok。无浏览器/登录/注销/reset/JWT/Redis/凭据文件/新测评/报告生成/评分重算/6144或oracle/budget重跑/production。C区两个小脚本bash-n0/editor0，Go/build未重跑（运行源码未改）。三个既有主报告追加完整证据。
- 工具失败保留：PS末行CR使只读SQL1064/exit1、本地Node引号exit1/SSH前；首次清理脚本误把dimension/module派生ID当UUID，备份后停止、SQL未执行/owned根仍1，经真实源码run-d-/run-m-＋精确PK纠正后成功；冻结稿预期marker未找到未上传，改源码已核纯DTO合成稿。final终端chmod仅owned归档目录误设600/新指纹644，按目录/文件分开纠正并01:55:31Z权限errors0，sharedpermissionchanges0。
- 下一仅有界只读syscall/bootstrap取证候选；如需新执行/修共享配置、权限、依赖或logger，另获用户确认。本轮root-vs-liming失败机制已复现，底层根因尚未确证，不能直接repair或重复HTTP生成。

## 2026-10-04T09:14:04Z MT-REPORT-DIAG-01 staging诊断已发布；409定位LO子进程，finally SSH阻断

- **清理未完成，不能称0或DONE**：首次exact-one清理SSHexit255/无远端输出，唯一有界重试exit255/ssh_timeout；按2次失败停止。09:14:04.534Z原正常admin只读results/detail仍200/ownrun存在，最终SQL/11表归零/旧基线cmp未能执行。下一SSH恢复后优先精确清理，不能重新生成或以通用exam-delete绕profile保护。
- 新fresh Linux backend build0/SHA8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93；freshchecker build/vet0/SHA6d17e1be8d8bbc8b7606e088f5e38f598034dd9c752c94f66c83fa8260954b92。existing backend-only部署exit0/verified/noRollback，旧PID4106排空四项0→新PID9059，3fresh positive_app11/159/67/21/cached4/CRUD/205/privateprobe/旧链HTTP全PASS。只backend替换，DDL/front/template/config0；上传4项SHA核后finally清0。local6144/652/0/9为前阶段证据，未重复全量。
- 正式parent保留，新backend_1455f3ab79235854/root0700/files0600当前fullDB gzip/SHA6fb3065f92e86df3cda44dbaac2d538dab7e476bdb5fa56041070a5e1f7d139d；HTTP写前http_1455f3ab7923新fullDB SHAe163aa312556938e5af0984d3d3cbe6162abac52c503474208d34098206543c3，均singletransaction/routines/triggers/events/gzip通过。当前旧12表051795bd969650d0e55da676083078255331c66405b5fff0ee4ca9d0a7641d0b非历史81aeaa；写前主70/1488/134494/295328/1349/27，465旧PDF。部署前后currentbaseline相同，但复现后最终cmp因SSH阻断UNVERIFIED，不删除/还原非own新增旧卷。
- 原page不存在后同hostforceNewfalse复用新983629c4-fb41-4913-af01-54786ac2d78a；清routes、正常getInfo200/admin1/wildcard/cookieStoreMatch发布前后及finally均true。无login/logout/reset/伪JWT/Redis/tokenfile；本轮identity/paper凭据仅browsermemory，finally全部testglobals清空，原认证首页保留。
- 唯一MTH1455f3ab7923/00201candidate，正常配置/name+四typed空串/freeze140700/原25-20时间、140HTTPsave/readback、completed run1/13dim/4mod/receipt1/raw3final3各140/总体50均PASS。仅POSTgenerate-test一次：09:12:21.096Z—09:12:21.765Z，HTTP409/669.1ms/固定校验错误。freshunitjournal游标+精确窗口09:12:21.742Z **stage=render class=lo_command_exec_exit**；不猜权限/timeout/数据错误。确认LO命令非零退出类别，具体退出码/原因UNVERIFIED；revision/current/audit0/privatePDF0，未重试或业务fix。
- **未清理exact主键**：exam1791105048344426522，candidatea3889750-1307-4fed-bc27-e89613d2f518，paperc6d52955-aeec-4abe-8468-c5f350f1df84，runea660428-f2b5-4445-b6e7-9d8b2351e00d，bundle7a20628b-d061-4606-9874-db38aa198f80。复用http-owned-cleanup仅内存4→1、精确标题candidate与count1，执行SHAaac49148c86dd3a477551143e8789ad3778ea4341f01a9af671bf5f6aba7de13；FK/SafeUpdate/原PK闭包保持，无新持久化工具。两失败receipt保留。
- 下一：SSH恢复后接续已有清理授权，不再索取部署/登录；另行有界授权才可诊断实际liming/systemd环境LO非零退出（exitcode/安全stderr类别/权限与sandbox事实），不能直接修配置或业务。未execasuser/权限改写/更深logger/oracle/生命周期budget/四560重跑/production。详情与证据见四真实/部署报告顶部。

## 2026-10-04 MT-REPORT-DIAG-01 仅本地最小诊断

- 仅GenerateTestReport失败fixed stage/class，复用slog，success零日志/download nil不增；不改公开签名/DI/config/API权限/SQL/锁/事务/缓存/90秒timeout/模板评分/cleanup。LO保留原固定中文Error及typedcause/category，不保留命令output。未知Error不格式化，无SQL/参数/DSN/PII/path/ID/token/rawerror输出到新增诊断。
- 有效RED1pass14fail/parse0/exit1→最终三包77pass0fail0skip/parse0/exit0；真实service/sqlmock、secretwrapped/nested/unknown、实际坏Word与409/503/200、LOtypedcause及目录清理。两个write分类失败批次保留：实际Windows ENOTDIR PathError也IsNotExist，安全分类优先typedpath（permission仍优先），不是改业务返回/放宽门禁。最新本agent实跑全Go6144pass/652顶层/0fail/9原skip/parse0/exit0，相对昨日6090/646增54事件/6顶层；fresh Windowsbuild/vet各0/零error，14文件限定diff-check0。
- 独立只读CodeReviewerPASS/无blocker；更深层schema/load已丢cause只记validation，原staging409根因仍UNVERIFIED。未补marshal/全OS/竞争cause完整矩阵，不称coverage/真实DB或产品DONE。无SSH/browser/mainDB/生产/部署、oracle/budget变化。
- [时间纠正] owned-sql提交15:51:29/39/40/41CST→07:51:29/39/40/41UTC，下一旧窗口核证须覆盖07:51—07:53UTC；本轮仅读取本地receipt，不假称重新查journal。已有534f0abb…/560save/4run52dim16mod4receipt/cleanup0及465不变仍历史。
- 下agent必须从当前source fresh Linux build完整SHA后受限backend-only staging发布，一份exact-owned临时all50复现，采新UTCstage/class/finally精确清理/旧baseline不变；若validation/unknown仍不足，申请该阶段深层诊断，不自行业务fix/oracle/budget扩展。详见[完整影响/合同/交接](management-traits-four-real-verification-20261003.md#L3)。
- 终端经验补充：sync timeout转opaque terminal后，新同步调用cwd/变量不等于原会话；一次错误根目录build/vet退出1，不能误当sourceFAIL。用原terminal ID单次读实际6144/652/0/9及build/vet0证据，后续命令显式Go cwd。历史工具失败保留。

## 2026-10-04T07:54:06Z 正常会话四HTTP评分PASS；真实generate409阻断，finally0

- 本轮新existingpage正常getInfo200/admin1/wildcard/cookieStoreMatch，清page/context mocks后使用原正常会话，无login/logout/reset/JWT伪签/Redis/secret文件。最终同页home/getInfo200保留；本轮浏览器合成密码/token globals及MTstorage清空。
- 不重复部署：online534f0abb…、front98547b68…、TEST05c55e77…实际不变，最新actual positive_app11/159/67/21/4queries日志只读复用。新当前完整backup=http_afc22a7e8edf/root0700/files0600/DB8b4863124d7e20b35e9ec0a3b8022805484ee02da9f0aa92fa4bddbed0a593fd，gzip/SHA通过永久保留；本轮旧12表81aeaa…/465逐SHA/所有运行资产cmp相同，不抹旧漂移。
- MTHafc22a7e8edf四正常admin配置/freeze、candidate配置name/typed空串五键、tester真实正确密码五键/140700/固定shuffle/25-20点/4missing409均PASS。560HTTPsave/readback、4completed/52dim16mod4receipt、repeat与mixed首次双HTTP并发same run/create+reusePASS；原SPECS+Fraction独立raw/final反向40次、选项、sum/count/六位缓存/等级/常模/receipt时间全部PASS，50/0/100/565195/11154。UI仅一candidate140恢复/同答案save-next/同卷deadline/submit完成页及admin1行/13dim4mod，非四全UI/新开始按钮全链。
- **首份00201candidate50正常admin generate-test HTTP409/610.78905ms**，SQLrevision/current/audit0/privatePDF0，journal只有请求状态无底层原因，不猜timeout/renderer/权限。依businesserrorstop，其余generate/view/download/真实PDF字节SHA/9页/灰环/144客户段及生命周期全部SKIP。不修业务、不重跑或修改第三轮失败oracle/210秒fixture，不复制新测试绕预算。
- 首candidate工具误带adminBearer导致business1，源码确认其被当participantcredential；仅内存helper改为前端实际匿名首次注册合同后PASS，无业务修复。PS嵌套引号progress本地exit1、不存在路径read失败保留，不影响后续实际只读证据。
- C区仅新增exact四receipt清理入口，bash-n0/diag0/实际cleanup0；4exam4link2candidate2tester4paper560pq2800桶560qs4snapshot4profile2bundle及4run52dim16mod4receipt清0，11表逐0/private0。07:54:06ZfinalSSH0、主78/70/1487/134354/294628/1348/27、15FK、source各140140700、3svcactive/healthok/窗口critical0（不掩generate409）、旧465/12表/runtimeSHA不变。无上传payload/生产访问/DDL/restart。
- 证据scripts/test/results/MTHafc22a7e8edf/owned-sql.json SHAe8616a780003e1469497e75883cf433f15d975342fe706c0bb0b575912823c37、score/http/browser/cleanup/final receipts。整体BLOCKED于真实HTTPgenerate409，综合oracle历史三轮FAIL未修及生命周期预算未解除；下一独立有界诊断需明确授权，不重复索取网页登录或部署。详见四真实/部署报告顶部；历史记录原样保留。

## 2026-10-04T06:31:12Z 零分backend已发布；恢复库PARTIAL、HTTP/UI仍BLOCKED

- 实际原page不存在，同host复用登录页getInfo401/credential0，finally仍401/页面保留；无login/logout/reset/自签adminJWT/Redis/secret文件。用户需要正常重新login；没有因401跳过SSH发布。
- 当前源码fresh Linux build0、新staging server SHA534f0abb7f5e90a9eb8da5606763afa0bf15e1fd2994a0418f82cb503e4bf6d5；checker a5c56290…fresh build0。旧abea56…bin/mtime/numeric1000/1000/755永久保留。backend_31a1dce2c6b90249 root0700/files0600当前fullmysqldump gzip/SHA12d2ded728d3ba42aee40807b96ae897c93146ea5fe8c889459ee2bfb5f96646通过；原3backup不删。旧PID2005 drain四项0→atomicserver→stoppedfreshgate→PID4106/postgate/exit0，无rollback。3fresh positive_app11/159/67/21/4queries/CRUD/205/002各140/700/private0700全PASS；DDL/front/template/config执行0，production未访问。
- 上轮backend-only内存helper未永久落盘，查证后C区新增小范围入口复用原部署合同，不假称找到旧helper。新运行源码无改动，昨日6090/0/9/build/vet仅复用不重跑；本轮freshbackend/checker/隔离夹具build/vet0。
- 当前fullbackup仅恢复mng_lifecycle_test_8a88d9d20f42d4f7，crossschemaFK0，真实public service/renderer/目标LO24.2，非authenticatedHTTP。四560save/4run52dim16mod4receipt、重复/mixed双连接唯一、4实际PDF三SHA通过；zero660827bytes/SHA53a6993bc642ac9ebb9befd59f3f22b1d0116e25678e0b1161b7a9a6fb8c5bb8、9页，实际5灰环各360/360、白洞1.0、蓝pixels0、五0.00字号12/10pt，第三页实看。全部TEST合成数据，主report0，不把恢复库actor1当HTTPadmin。
- 到期零/3答真实incomplete、0/140与3/140、正式overall/dim/mod NULL、timeoutreceipt及render前拒报告/expiredfill拒绝PASS。第3到期140样本填答在210.08秒actualsave失败、整体exit1；夹具上下文210秒但底层err未记录，不确认为逻辑bug。failfast后到期full/race、end、retired/revoked、resumeUI、Worker/restart、failureatomic均SKIP。旧昨日边界不称今日重跑；恢复库服务不覆盖HTTP/UI。
- 新oracle3次有界：DataSnapshot实际json:"-"、score在numericanchor不是SDT、最后Decimal零字面量0E-12不同Go0.000000000000真实查证；第1份50报告9页/36客户全文/精确Fraction/OPC PASS后停止。其余全量144段/非零OPC独立门禁未关闭。达到三轮不再编辑，不改source/阈值；独立原annulus对zero实际灰矢量采样PASS仅限定像素，不覆盖综合oracleFAIL。PylanceMCP启动失败，AST0/实际venv3.14.4验证；PS引号与GBK读取失败保留。
- finally专属restoreDROP0、payload0、主11表逐0/private0/15FK；06:31:12Z终验exit0、主78/70/1487/134354/294628/1348/27、当前12表81aeaa…/465逐SHA/frontconfigsunit不变、3svcactive/healthok/startupcritical0。只本轮currentbaseline不变，不抹旧drift。证据backup/lifecycle_8a88d9d20f42d4f7长期保留；本地synthetic archive34381785…/全部PDF/失败receipt在scripts/test/results/mng-zero-staging-8a88d9d20f42d4f7，无凭据文件。
- 最新状态DEPLOYEDPASS、RESTOREDPARTIAL/exit1、HTTPUIBLOCKED、综合oracleFAIL、限定zero灰环PASS；主代理下一需用户正常login及新有界验证器/生命周期预算收口，不重复deploy或现代化taskmetadata。完整SHA/每case与失败历史见四真实报告最新节。

## 2026-10-03 MT-ZERO-RING-01 精确零分灰底环（仅本地完成）

- 用户“继续”只批准精确零分浅灰底环，保留0.00原字号/font/坐标及所有非零样式。本轮无SSH/deploy/DB/userbrowser/restart。实际原图/模板确认余量透明为根因，不是数字遮挡；有限运行时样式例外只改五环idx1直接填充为模板已有E7E6E6，线noFill保持。big.Rat.Sign判零，.004显示0.00不触发，不epsilon/null零/全余量改灰/13详情新增环。
- 原客户/source-layout/optimized/compatible/TEST资产不改。TEST完整SHA05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c及mng-test-lo-template-v2保持；评分四轴/content/DTO值/SQL不改，策略标MT-ZERO-RING-01。既有生成UUID+revision+FileSHA仍管理新呈现；本轮无DB revision、旧下载不重渲染。
- 有效GoRED5pass4fail→专项11pass/0/0；原四DataRaw逐SHA核验离线Go重放13事件PASS。真实旧全零PDF新graygate RED5missing/exit1保留。八代表Go→LO26.2.5.2真实九页GREEN；全零/等值零5灰环、mixed2、nearzero0且不假称sixvisible，非零OPC/style原样。
- 四冻结客户全文回放各9页/六图/36段完整可提取，共144/144；全零5环216DPI RGB231/230/230每个360/360桶、whitehole100%、bluepixels0，五0.00/12pt整体10pt模块保持。same-engine零/.004数字bbox/font/size/color精确相同；组合图真实数据标签约0.028pt纵移不冒称全PDF全等。旧LO24.2与本地26.2字体坐标差异明示，不声称跨引擎一致或全文无遮挡/章节顺序全验。
- 新全零666519bytes/SHA2045b29f9263928bc81a742809f554175f0edcaaf3161a2f22d2cfd689e79536；receipt scripts/test/results/management-traits-zero-ring-frozen-pdf/1cb36f/contract.json SHAc55271f8b151603207023e0aab02459663a43d58868a6169a38e9a5c151d3d29。75OPC中68原字节不变，余7仅文本/数值/zero-fill；11297protectedfiles SHA不变。主代理实看零分第三页/九页contact。
- 原source-layout10/compatible17（117子项）原方法未改/隔离短输出PASS。原CLI会覆盖历史产物，重放用新oracle upstream-safe隔离输出，不直接旧CLI；长路径LO exit0无PDF的11fail及冻结长文案几何首次fail保留，后者按五锚点实测统一20pt只平移采样区，不降低灰环门禁。
- 主代理全Go6090pass/646顶层/0fail/9原环境skip/parse0/exit0，较6081增加9事件/1顶层；build/vet各0且stdoutstderr0bytes，独立两轮review PASS、源码diagnostics0/diffcheck0。未coverage/race/frontend。新Go测试3次编辑停止，不顺手EOF或改无关格式。详见四真实报告顶部完整影响和SHA。
- [限定补充 - 2026-10-03] 下方staging零分FOURBLOCKED历史未被本轮localGREEN解除；staging仍需未来另行发布/验收，生命周期expiry/incomplete/offlineWorker/restart/race未验。原560save/4run52dim16module/144text及DB-downloadSHA PASS保留，不冒称本轮远端实时事实；完成后停止，不自动发布或开始下一逻辑。

## 2026-10-03T15:12:21Z 四真实评分/PDF完成，零分环图可见性BLOCKED；清理0

- 用户正常重新login后原page/context清routes，真实getInfo200/userId1/wildcard/cookieStoreMatchtrue；finally仍200/home/adminCookie保留，本轮test globals及子页MT storage清空，子页关闭。不新admin/login/logout/reset/自签JWT/Redis/secret文件。
- 沿最新backend receipt，在线abea56b32f5159211b44f065facfc740ee961b313c4a51b377f915284fc3a2cf/parser73d72d03…/index98547b68…/TEST05c55e77…/内容b0498249…/DDL7ff…均核一致，读取最新actual positive_app gate，不build/deploy/DDL/restart。APP_ENVproduction仅配置选择，两个report env仍staging/private0700。最新完整backup fb63c316…gzip/SHA通过永久保留。
- 新当前baseline root0700=four_20261003_145100_b5bc510880c1（原正式backup下），12表SHA81aeaa3b04607b6e6562f30ace2905d91a13cce6cc4f79c421c78616494545d4/旧PDF465逐路径SHA/configdistdropin前后精确相等；不能从历史f57/旧dump称全期间无漂移，14:46之前既存漂移不修。
- MTRa05eb3af5e49四组合全部正常admin创建/freeze、candidate配置name-only/tester正确密码身份五键、140/700组卷/重复同卷固定序/1500secdeadline/1200sec提醒/缺题409。真实560save读回，重复和mixed双并发manual同run；4run52dim16module4receipt，SQLraw/final/selected/sum/count/6位缓存/未舍入等级/topbottom/时间一致，独立Fraction总体50/0/100/565195/11154→50.00/0.00/100.00/50.67，常模705/13/学习53.75/创新50正确。原nested tester实际失败已限定解除。
- 四服务器LO24.2 A4九页PDF各648247/657576/666442/648380bytes；authenticated view/download8次byteSHA与DB/private0600/scp副本精确一致。每份36所需客户段→XLSX+DTO+PDF逐字PASS，共144，非每报告all205。4revision/current及audit闭包/旧pdf_path不写/孤儿0。完整SHA/合成PDF/SQL/36页图片见四真实报告顶部与results/MTRa05eb3af5e49。
- **FOURBLOCKED**：00201tester全0分第三页五环图区空白仅0标签，可见环形矢量5/0/5/5；原TEST chart1～5余量点a:noFill且线noFill，value-only正确0/100，零分没有可见环。普通oracle exit0/严格six-visible gate真实exit1。不是评分错误/随机LO漏图，需独立模板余量可见性切片，本轮不修/不替模板/部署；不把全部4PDF生成或cleanupPASS称FOURPASS。
- 实际UI仅一个candidate saveNext/组件重新load固定答案/140submit→thankyou，以及原admin结果1行/13维4模块/iframe查看/下载callback HTTP200；native download文件事件timeout明示，不四完整UI。8边界PASS：wrongpurpose401/foreignowner401/foreignoption409零写、retired新卷409且已有save1不改deadline、revoked fill/resume/generate409/无孤儿新file、真实adminresume<=300秒且detail不延凭据、自然5min后401。paper到期/incomplete/offlineWorker startup/restart/expiry-submit竞争未验，视觉失败停止且不改source时间凑pastdeadline，不伪造token或重启。
- 初始测试20字符识别码超varchar18/1406账号0→仅测试输入17字符；五键id误作participantId多创建1无paper synthetic账号也清0，UI初始化/异步路由/native下载工具错误保留。首次cleanup主键IN触发1175/8MiB rangeoptimizer，事务自动回滚exit1；仅同连接SESSION64MiB、SafeUpdate1/FK开启后重放exit0，无共享config变化。
- finally精确4exam4link6candidate2tester6paper840pq4200桶840qs6snapshot4profile2bundle及结果/报告audit清理；只删除4 SHA核过私有PDF，源题/客户/V67V96不写。15:10:54cleanup0及15:12:21finalreadonly0：11表逐0、所有ownedlegacy0/private0/15FK、主78/70/1487/134354/294628/1348/27、源各140/140/700、旧465+12表当前SHA/config保持、三服务active/healthok/窗口critical0。正式备份保留；其他production未访问。
- [纠正 - 2026-10-03] 下方14:46“原admin401/四组合业务0”已由本轮正常会话和四真实评分PDF限定解除；原历史失败仍保留。最新阻断是六图零分可见性，生命周期仍PARTIAL，不称产品全部完成。只新增C区测试/清理脚本和文档，业务源码/模板/配置变更0。

## 2026-10-03T14:46:05Z backend-only staging已发布，四组合仍受原admin401阻断

- SSH真实恢复vm-ubuntu-go-dev/liming/exit0；只20.200.136.133/knownkey/Strict/Batch/Conn10/Attempts1。原集成浏览器同page清routes后getInfo发布前后401，cookie/store相等、首页及原凭据保留；不重新login/logout/reset/新admin/自签JWT/Redis/令牌文件。四组合全部freeze0/save0/submit0/run0/PDF0，正常tester登录及13维4模块/六图全文/UI/expiry/resume/revoke/Worker/concurrency均SKIP认证前提，不以缺token401当业务PASS。
- 本轮fresh Linuxbackend build0/SHAabea56b32f5159211b44f065facfc740ee961b313c4a51b377f915284fc3a2cf/49839141bytes已实际在线；parserSHA73d72d03…同交接，核验器freshoverlaybuild0/SHAa5c56290…16442455bytes。旧PID27605 stop→PID/cgroup/8092/positive_app连接0→atomic仅server→freshgate→PID30638post，deployexit0/verified/noRollback。无DDL/front/template/config变更，不重跑6081/front355，不改业务源码/生产。
- 发布前**既存漂移**：current12表SHA81aeaa3b04607b6e6562f30ace2905d91a13cce6cc4f79c421c78616494545d4非历史f57；旧465PDF共同路径变更0/新增2/缺失2且SHA排序多重集合不同。原因未查明，不能称仅rename/自13:55不变。原正式三归档SHA/gzipPASS永久保留，未用旧dump还原main。
- 新root0700回滚证据=/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/backend_ce78e7e96e76bfd6；当前完整single-transaction/routines/triggers/events gzip12794331bytes/SHAfb63c316ddea0a299e5e24c98428e530c1bbe0251b4ee7dcf68b1ea413d89d07/验证PASS/0600，旧bin原字节mtime及numericmetadata1000/1000/755独立保留，不能从副本600猜mode。没有restore或删除正式备份。
- pre/gate/post三fresh实际positive_app生产CheckRuntimeSchema首次cached/11-159-67-21/4metadata/emptycanonicalguard/CRUD/002各140-700ASCII均PASS，实际TESTcontent205规则/private0700写关删files0。health200/admin三401/合法participant四401/空body400/general+001+002+003MBTI+00401五只读业务成功。APPENVproduction仅配置选择，REPORT_EFFECTIVE_ENV/MNG_TEST_REPORT_ENVstaging保持。
- 本轮**当前baseline前后**legacy12表/PDF465逐路径SHA/完整distconfigsdropinsunit/Schema签名精确相同，index98547b68…和TEST05c55e77…/b0498249…不变；14:46:05Z主78/70/1487/134354/294628/1348/27、11sidecar逐0/15FK、三服务active/healthok、启动critical0。当前无marked业务数据/privatePDF；exact四上传文件及/tmp/mng_backend_scope_ce78e7e96e76bfd6删除remaining0，备份及rollback保留。
- 工具失败记录：首次Node Bash插值syntax1在SSH前；strict历史baseline不相等exit1是真实漂移；最后只读here-doc末行CRLF ROOT命令exit127，LF Node stdin重放BACKEND_EVIDENCE_CLEAN_FINAL_EXIT0，未再次发布/重启。部署本体和主终验各exit0。后续远端脚本stdin须归一末行CRLF，不能仅处理here-string内部。
- 下一只需用户在原页正常重新登录→getInfo200/admin→正常tester登录→四组合真实PDF/生命周期/ownedcleanup。已授权backend部署完成，不重复deploy/确认，不用本地或startupPASS覆盖四组合BLOCKED。详细完整SHA/四项SKIP见两最新报告顶部。
- [纠正 - 2026-10-03] 下方14:37未发布/SSH双timeout已由本轮实际YES限定解除，admin401仍未解除。下方13:55旧指纹与465不变为历史，不作为本轮起始事实；只承诺本轮currentbaseline操作前后相等。

## 2026-10-03T14:37:18Z 已授权backend-only续作双阻断（未发布）

- 完整读本记忆/四真实/最新部署报告和规则后实际接续；SSH knownkey/Strict/Batch/Conn10初次及一次有界重试均连接前timeout、exit255，远端shell/SQL/备份/上传/stop/drain/install/restart未开始。下方13:55:33Z的11表0/465及指纹不变/三服务/备份SHA是历史，本轮无法复核，不能冒称当前PASS。
- 原用户admin同page首页保留、page/context route清除，cookie/store真实存在且相等，实际getInfo用原cookie/store三次均401；不能以有cookie或首页UI当合法认证，不推定JWT到期或Redis失效根因。未新auth/logout/reset/自签JWT/Redis/storageState文件/输出保存token；终验homePreserved/既有token保留true。四组合全部0save/0submit/0run/0PDF及生命周期/UI SKIP：合法admin与SSH双前提阻断，无marked数据产生/cleanup_required0，非remote cleanup PASS。
- 当前源码fresh Linux backend49839141bytes/SHAabea56b32f5159211b44f065facfc740ee961b313c4a51b377f915284fc3a2cf；freshchecker16442455bytes/SHAa5c5629098898a67311e74c01beb6bfd0ebfef556880c21a0f0e418bab439646，build0/无error，仅本地bin保留未上传。ignored bin overlay映射永久scripts/tools核验器到cmd main，不复制或更改任何业务源码。parserSHA73d72d03b3a899ac6982e4bb9783e425f8e6491ff99712ae05706accf79a115d。
- 本轮原生nested/Gin回归42pass/7顶层/0fail0skip/parse0/exit0；Windowsbuild/vet ./.../Linuxoverlayvet各0/输出0bytes。全6081/0/9及review PASS仍为前轮，未重跑。公网14:37:18.359Z health200/ok只证HTTP健康。
- 本地index98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51、TESTdocx05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c、xlsx b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c重新核SHA一致，不frontendbuild/部署/模板改写/DDL/生产操作。BACKEND_DEPLOYEDNO/FOURBLOCKED明确记录于两报告顶部；不重复索取已授权部署，只需SSH恢复及同页用户正常有效认证后重新baseline→backendonly受限backup/drain/atomicrestart/freshactualgate→testerlogin→四PDF→exactcleanup。
- [纠正 - 2026-10-03] 下方“管理员前提已解除”仅13:55历史有效；本轮原token真实401，此刻不再满足四组合写入前提。旧0ee9在线信息未本轮SSH复核，也不包含新parser修复；不以本地fresh build称已发布或业务验收通过。

## 2026-10-03 MT-REAL-TESTER-SCOPE 本地解析器收口（真实staging待fresh）

- 仅普通coding，完整读记忆/四真实报告/Go规则；本轮无SSH/browser/真实DB/DDL/deploy/restart，不操作原admin会话。13:55:33Z主11表0/465PDF/指纹不变是历史交接，未本轮远端复核，实际四组合仍BLOCKED。
- 接管磁盘已有SELECT词法scope和真实完整11表warm-registry→LoginForm隔离回归；先复跑正确密码auditDriver1/五键/token绑定/旧写0、错误密码lookup1/scope0、孤儿复合1行→nested0均PASS，不伪称本轮新RED。既有RED账本记录HTTP200/business500/auditDriver0/domainReject1/otherErrors0/write0保留。
- 新安全补充先有效RED14子10pass4fail＋顶层fail/exit1：canonical二次resolve误绑、correlated父scope预算、ID先type后超长driver成功errnil。仅schema parser两轮apply_patch：resolveColumn带owner，类型基线＋当前type→id OR逐段更新，checkValue接原词避免canonical被再当alias；不改权限/Schema/bind/ASCII/UTF8/容量值/guardSQL/孤儿拒绝。mixedcandidate36/tester4两顺序及各自overflow6子也通过，不承诺任意SQL泛化/UNION独立scope/所有Update WHERE覆盖。
- 最终原生专项42pass/7顶层/0fail0skip；四包5369pass/275顶层/0fail4skip；全Go6081pass/645顶层/0fail9原环境skip，JSON解析0/exit0。Go Build任务exit0无error，vet0且stdout/stderr空；独立review PASS，补mixed后终审7pass＋Gin3pass/0fail0skip，文件SHA不变。editor无测试发现，不当editorGREEN；非coverage、未front/PDF/race。
- 001SHA保持7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8。生产规范化gofmt一致，测试service2拼接空格/EOF与既有handler2拼接差异保留，按编辑上限不再format。全部共享callback/mapExpr/CASE/Rawsource/owner/guard/profile/create/write/load/results/resume/report消费方审查，仅parser需要同步改，其余理由及准确证据见docs/management-traits-four-real-verification-20261003.md顶部本地交接。
- 下一先fresh当前Linuxbackend/validator完整SHA，不复用旧0ee9server；只读baseline→正常admin认证→exact-owned封闭testerfreeze/login→四真实完整评分/PDF/生命周期→finally主键清理/旧465与指纹不变。原stagingFAIL未解除，不自动发布/新DB写或称产品DONE。

## 2026-10-03T13:55:33Z 四组合真实续验BLOCKED（封闭身份guard查询拒绝，已清理）

- 用户正常admin集成浏览器原page/home保留，实际auth.js Admin-Token仅浏览器内存；page/context unroute后GET getInfo HTTP200/success/admin/wildcard均true，不打印token/PII、不logout/reset/自签Redis/JWT。此前缺凭据阻断已限定解除。
- 四owned配置MTR0fa2fc8f9d真实创建/freeze均PASS；源码封闭isOpen2不是交接0。两code各140mapping、manifest/mappingSHA与上阶段相同；00201 candidate五键身份＋typed空串/140题700项＋重复同paper同题序deadline/25min20min＋未完整manual409 PASS。保存0/成功交卷0/run0/PDF0；参与者UI与13/4/评分oracle/报告及生命周期全部未验。
- 00201 tester正常新增status0/del0且正确密码后login HTTP200业务拒绝身份；实际13:51:25Z journal runtime_guard.go:249 management traits data rejected，audit→revision→run nested participant参数查询rows0。TryTesterIdentity在guarded s.db走scope；SQL参数检查将无前缀参数按外层table解析，嵌套scope风险已定位（源码推断，backend隔离RED未补），不冒称DBUnknownColumn或已修。按actualbugstop停止其余两身份，不改业务/DDL/模板/配置、不重建部署重启。详情docs/management-traits-four-real-verification-20261003.md；无凭据browser probe保存仅syntax验，不重试失败登录。
- finally exact主键事务/SQL_SAFE_UPDATES1/FK启用，先确认2bundle仅4ownprofile引用；4exam4link2tester1candidate1paper140pq700桶140qs1snapshot4profile2bundle清理，exit0/residual0；11表最终逐0/privatePDF0，无新PDF删除。13:55:33Z SSHexit0，主78/70/1487/134354/294628/1348/27、15FK、source280/1400、旧12表f57f35…/465PDF逐SHA均不变；正式备份3SHA/root0700/DB9ab00b04…保留；server0ee9…/front98547…不变，三服务active/healthok。窗口guard拒绝1不能称业务errors0。
- 新脚本归scripts/db、probe及脱敏JSON归scripts/test/results；无secret/token/DSN文件，无storageState文件或新adminsession。HTTP-origin crypto.randomUUID不可用工具错误在首个写入前，改getRandomValues后成功，不归业务bug。下一由backend负责guardedDB嵌套SQL RED→最小修复→fresh复验；不得拿本次cleanPASS称FULLPASS或继续未授权业务fix。

## 2026-10-03T13:20:41Z 四组合真实续验凭据BLOCKED（只读终验PASS）

- 完整读取记忆/部署/本地实施报告及实际auth/runtime/report合同后续作。当前本地专用凭据键0、已知env/YAML无管理员引用；远端进程同类key0，受限/home/liming/root/configs深度2文件名和已知配置键名核查未找到管理员密码来源。SQL仅启用未删除user_id1计数1，不读取hash/PII。用户补充SSH私钥已实际用于liming认证，但不是网页管理员密码；不新建/重置账号、猜旧密码、自签JWT/伪造Redis session。未captcha/login、无失败次数写入。
- 新server0ee9b326…/index98547b68…完整SHA与部署相同；REPORT_EFFECTIVE_ENV及MNG_TEST_REPORT_ENV均staging，三服务active、启动仍21:10:51CST，未重部署/DDL/重启或重新跑fresh应用账号fullgate。公网12实际合同PASS/0fail/exit0：health200、六管理缺auth401、四合法participant缺token401、空body400；仅sentinel，无业务写。
- 最后只读13:20:41Z/SSHexit0：正式root0700备份三SHA OK/DB9ab00b04…；旧12表f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e及465旧PDF逐SHA不变，11新表逐表0/15RESTRICT FK/private liming0700 files0/healthok。没有owned测试实体/文件/会话，cleanup_required0而非清理完成。
- 四组合freeze/create/560save/submit/13dim4mod/PDF/UI及resume/expiry/revoke/Worker/concurrency全部SKIP（缺正常管理员登录凭据），不将只读PASS称完整验收。首个Node嵌套模板syntaxexit1在SSH前，LF stdin重跑成功；无业务修复、源题/V67V96/客户/旧PDF/production改动。详情及四组合表见部署报告顶部。
- 下一只需用户给既有受限网页管理员凭据文件绝对路径及username/password键名，不聊天粘贴密码；内存消费真实captcha/login/getInfo、不能用SSH/root/JWT配置代替认证。部署YES限定状态保持，不重复部署授权。

## 2026-10-03T13:13:15Z staging TEST部署YES（最小验收，四组合未执行）

- DEPLOYED=YES仅部署/正确协议及旧链只读验收；四真实组合/PDF/expiry/revoke/resume/Worker并发仍未验，主11sidecar每表0，无profile/参与者/paper/run/report创建。管理员安全明文凭据/有效登录会话本轮未取得，专用进程env key为空；下一实际captcha/login，不使用旧脚本口令或伪造JWT/Redis session，不重复部署确认。
- 正式旧失败checker/实际router/source已读，纠正工具participant完整JSON缺token401、空body独立400；pre允许11但强完整gate；原uid/gid/mode独立保存，mtime恢复不猜私有600副本。有效RED3fail→GREEN3pass；真实注册JWT/handler7子＋1顶层PASS，metadata实际独立文件行为PASS，fresh工具build/vet/bash-n0。业务111Go文件聚合SHAe0d6f6f61efe2b58840edc2352074d6c35ce9166a460b3dfed2a42306faff644前后相等，无业务/API/001SQL改动。
- 复用前次fresh且完整SHA相同server0ee9b326c459611babb6951c63291abcdfc27491e8782d3dd9178e4bb56d8483、index98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51、dist-tare0b37707e750510d8b0789b3f7560668ff59d39cef100ee18a27c1dfa65c101e；不重跑6039/front355/无关build。fresh checker16431599bytes/SHAe6bbf1890cc40c76c0916beca7bf0ec6d274adc095e24a5175e1f29e4e96b0eb。
- liming/knownkey/Strict/Conn10目标仅staging20.200.136.133；真实APP_ENVproduction＋REPORT_EFFECTIVE_ENVstaging。旧PID24732、新guard27481各stop→PID/cgroup/8092/positive_app connections0，guard TESTdisabled先完整gate再排空；既存11表只读first/repeat/final签名3fb28c5c220c063795e77b3e27b8eed8ace01b0dc7ee82614a556af8f1c3c0d1相同，DDL执行0、15RESTRICT FK不变。5fresh checker实际positive_app fullgate11/159/67/21/4queries/CRUD/002140/700ASCII均PASS，不root替validator。
- 最终PID27605、启动2026-10-03T13:10:51Z、三服务active；四MNG_TEST key真实生效独立staging/原报告§4绝对路径，private liming0700/files0、TEST两文件liming0600，真实读205规则/写关删探针PASS。新envSHA7c56d3ab4cab53633b0b4e52628be69f7af7e29f208707023a75ecde8e2cb082 root:liming0640、新dropinSHA96dff874996f27cf15335e84cf0aa6cd3aeea23dadd2241a04d73d55d5918367 root0644；共享configs/旧模板归档compare0。
- 内部health200、admin三401、四participant合法body401、空body400，general/00101/00201/00301MBTI/00401五读HTTP200业务成功。公网health200/ok、index200同SHA、四合法401/空400，393front逐文件SHA全部同fresh，无浏览器四组合/真实报告声称。主计数78/70/1487/134354/294628/1348/27，旧12表f57f35…及465PDF逐SHA不变，启动窗口关键错误0。
- 原正式backup112853_fd24d4ee35a9/3SHA和失败deploy_3461565581060643永久保留。新证据deploy_24bde3635e559ab7 root0700/files0600含checker源码/bin、实际script、DDL未执行副本、schema/gate/drain/post/checksums/metadata/cleanup/final。即时旧dist.mng-rollback-24bde3635e559ab7保留；有新数据停止审查，不DROP11表/全库restore。exact上传/tmp/mng_deploy_staging_24bde3635e559ab7及本地两编译副本/空目录清0，候选bin保留，无secret/token文件。
- 最终远端13:12:40Z exit0、公网/源码guard13:13:15.810Z exit0。一处本地终验JS引号syntaxexit1在SSH前，无远端操作；LF直接脚本重跑成功，不重启回滚或业务改码。完整四组合/guardcapture/profilecreate精确路由、安全登录前提/marked清理交接见docs/management-traits-staging-deployment-20261003.md顶部F，不执行本阶段。
- [纠正 - 2026-10-03] 原13:01:44Z DEPLOYEDNO为前次真实失败并已回滚历史，证据保留；本轮重新授权续作已完成YES限定部署验收，不覆盖历史失败、不称完整产品DONE或production上线。

## 2026-10-03 MT-GUARD-AUDIT本地必需修复（actual待fresh复验）

- 用户明确普通coding且仅当前逻辑bug，已授权完整report TEST staging但本轮禁止SSH/mainwrite/deploy；不modernize任务、不重复索取逻辑修复授权。已完整读记忆/最新报告/规则及真实守卫、DDL、报告生成/下载audit写入口。
- actual audit五列只有id/report_id/actor_id/action/created_at，真实report_id引用revision.id，生成/下载写r.ID不是paper主键。仅guard生产一轮＋新测试三轮；canonical audit精确五列识别，报告canonical出现要求完整11表并复用原strict gate，global LEFT JOIN audit→revision→run核paper/exam复合身份，孤儿/空/不相关scope不裸成功；scope绑定按真实run/paper/exam/owner闭包，1000绑定批次，unknown report-only仍拒绝。原七core/capture历史兼容、租约/路由/签名/API/DDL15FK不改。
- 有效RED10pass35fail1skip exit1→主代理新回归62pass/10顶层0fail1skip exit0；四包5327pass0fail3skip、review PASS。主代理全量6039pass/638通过顶层0fail8skip，解析错误0/exit0，fresh Windows build/vet ./...均0且零error输出；5977基线增加62事件/10顶层及1明确环境skip。新外部owned恢复库audit回归未设置DSN明确skip，不访问真实库；原7skip保持，八项完整名称见报告。
- fresh本地Linux server49828320bytes/SHA0ee9b326c459611babb6951c63291abcdfc27491e8782d3dd9178e4bb56d8483；service审计测试二进制18524853bytes/SHAd1b1c3f7b6e5dc4d9c33644460ef1a0fc99c80435aefeba276db5bdcf2fe69cd，两编译exit0，仅ignored bin保留未上传。源码变化后必须重建，不复用旧actualalias验证器/旧server作为fresh证明。
- 外部用例只接受mng_guard_audit_test_<16hex>与匹配DSN/crossschemaFK0/11sidecar空，真实strict gate后transaction内synthetic关联fixture，ROLLBACK核11表0；不CREATE/DROP/ALTER/旧表写/文件会话。fresh Linux测试二进制必须保留cwd payload/Go-based Refactored System/internal/service及payload/scripts/sql/management_traits_001_runtime.sql相对层次，不能在任意/tmp直接运行后假称gate失败。
- 两Go诊断0/scoped diffcheck0；001 SHA7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8不变。生产规范化gofmt一致，新测试三处拼接空格及EOF差异三轮停止，不再格式修整、不碰无关脏文件。
- 风险仍明示：每report-enabled scope四metadata＋global auditanti-join成本未测；原gate单进程不保护外部/多实例写。fresh actual恢复/capture/真实审计闭包、四组合HTTP/PDF仍未验，不能本地GREEN称staging或可部署。详细新receipt/真实影响清单/远端仅交接顺序见最新本地报告顶部。
- [纠正 - 2026-10-03] 下方actual新阻断及“需再授权修复”为上阶段事实，本轮已获明确授权并完成本地bug修复；上阶段actual失败仍保留，未由本轮复验解除。

## 2026-10-03 新别名代码actual复验（完整Schema PASS，整体BLOCKED）

- 本轮仅授权staging受控复验，不主element DDL/部署/逻辑修复。SSH strict liming/knownkey/Batch/timeout10通过，MySQL8.0.46、APP_ENVproduction＋REPORT_EFFECTIVE_ENVstaging，三服务active。备份112853_fd24d4ee35a9数据库SHA9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1、三归档SHA/gzip/tar及当前application/config/templates tar-compare全通过；旧12表指纹与465PDF无漂移，旧备份适用且永久保留。
- fresh Linux validator SHA5cdfd07eeadf0fee60fab43f964d6cd441550907f09a96833b722a80f4e0674d、原service.test SHA6976307da8b0c206c66f41f2a45a12c46ee72da787749dc95889be133a77a0af，build/vet0。实际config.Load/JWT内部消费，root既有socket仅改exact owned临时DSN，SELECT DATABASE精确核验，无credential/PII/token输出/文件或grants变更；不是应用账号授权/服务重启验收。
- 唯一恢复库mng_verify_20261003_4407cd0a1cb24799，旧/新跨schema FK0；001first/repeat11表15RESTRICT FK、签名1020f16a37f84f9fbecdfd8c425969def679af26ed2eb4fb0fdea4a1b460d4e6相等，额外FK逐列签名ef9c1c7b409e4aa413e70de853d27b9818e09b6e339de8b4ffe1a6a178c2318a相等。生产完整CheckRuntimeSchema首次/重复/post fresh全部PASS，真实执行Columns均小写正确、GORM行数11/159/67/21、errorfalse；同实例第二次缓存PASS仅4metadata查询。无diagnostic替actual/忽略case/放宽validator。
- gate后两002真实生产FreezeProfile从备份原文各700行→140题/700选项通过、源capacity实际兼容，marked临时exam/link/profile/bundle按主键清理；实际配置密钥token sign/parse内存roundtrip PASS，未发HTTP或创建会话。
- **新阻断**：完整11表＋owned capture时public ManagementTraitsIdentityScope返回拒绝，post exit1；actual capture/table/extra-column三个投影标签正确且rows1/12/41，errorfalse。源码与actualDDL证据：report_audit仅id/report_id/actor_id/action/created_at，legacy guard capabilities只认可paper/exam/participant等或run_id/result_run_id，不识别report_id-only extra，查询保护数据前失败关闭。后续独立AllLegacy true断言未执行，不称guard PASS；本阶段不改逻辑/删audit/降七表/遮蔽fixture。需后续明确逻辑修复授权下RED完整11表/审计保护闭包→最小安全修复→fresh复验，**不能mainDDL/部署**。
- full gate通过后独立空mng_source_lock_test_20261003_e680e18840cd49ef原双连接test实际4pass/0fail/0skip、2.31s、exit0；commit/rollback/MVCC-currentread通过，第二连接1205及释放后成功；临时TCP→unix relay/child-onlyDSN，无凭据文件，fixture表0。
- 正式backup子目录actualalias_4407cd0a1cb24799保存first-repeat完整签名/actual日志/post失败/receipt EXIT1 remaining0 errors0；actualalias_lock_e680e18840cd49ef receipt EXIT0 remaining0 errors0。两上传/tmp/mng_actualalias_20261003_*确切owned目录已清0，本地唯一临时validator源码/两二进制/空目录已精确清理；不删除其他任务产物或正式备份，无token/DSN文件。final只读UTC12:10:16Z exit0、主mng0、旧摘要f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e、465PDF与server/index/config/templates全不变，计数67/70/1487/134354/294628/1348/27。
- 现场unit Userliming/WorkingDirectory/opt/talent-assessment；实际application.yml＋application-production.yml；exportTemplates liming0750、tmp/uploadPath与/data/uploadPath0755；Nginx root dist、alias仅/data/uploadPath/profile。四MNG_TEST key未设置，新TEST模板/内容目标与建议private根/opt/talent-assessment/private/management-traits-test-reports均不存在，未创建。下一先修guard并fresh复验→freshLinux/server/frontend完整SHA（stale不可用）→guard/drain停写→staging主DDLfirstrepeat→新绝对TEST文件/私有0700根及显式staging环境→四真实组合/expiry/revoke/resume/Worker/PDF/SQL→marked cleanup，当前不执行。
- 本轮未重跑5977/0/7或前端355；独立actual source-lock4/0/0不改默认7skip。不以完整Schema PASS或final cleanup exit0覆盖post实际失败。详细路径/SHA/行标签及交接见最新本地报告顶部。
- 清理工具经验：两次apply_patch Delete未实际落盘，保护性检查停止；最后按用户explicit exact临时source清理授权核所有权/两二进制SHA，仅删除本次三个临时产物及空目录，不终端改写源码内容。实测local temp0、专用DSN环境不存在，docdiffcheck0/diagnostics0；不能仅凭编辑工具“success”报告已删除。
- [纠正 - 2026-10-03] 下方actualSchema失败/列索引FK未跑已由本次fresh完整PASS限定解除；overall仍BLOCKED于完整11表capture/旧守卫，历史失败保留，不能称第一阶段全部PASS或可部署。

## 2026-10-03 MT-SCHEMA-ALIASES本地修复（actual待fresh复验）

- actual TABLE_NAME/ENGINE大写→小写tag扫描11rows/有效0来自上阶段；本轮3生产文件7SQL仅补小写AS，Schema22＋capture/guard4共26输出全部显式alias（新增23）。不改验证/签名/cache/过滤/JOIN/参数/排序/DDL。其余大写字段是同类风险本地模拟，不冒称远端已观察。
- 新驱动夹具按执行SQL生成标签/独立完整SQL去alias相等，真实GORM无alias零值、有alias填充。有效RED3pass12fail→最终15pass/3顶层0fail0skip；首次GREEN4/11是新COALESCE解析/guard查询fixture错误，按实际纠正，不放宽validator。生产一轮、新测试三次编辑停止格式追加。
- 14既有testfiles65prefix同步，仅SQL字符串/DI计数prefix；85测试文件其余字节保持。CodeReviewer只读PASS；列/type/NULL/capacity/索引/FK、cache失败粘滞/fresh刷新保持；scalar COUNT不改。
- 本轮无SSH/DB读写/DDL/部署/前端/模板/PDF，不触A区/共享旧表源题/私有PDF/正式备份/旧EOF。四包聚焦5265pass/258顶层/0fail/2skip；隔离终端UTF-8全量5977pass/628顶层/0fail/7skip，解析错误0/exit0，build/vet均0且零error输出。较5962/625增加15事件/3顶层，非coverage。开工后新增11:43:59Z只读receipt重新读取保留，不以本地GREEN称actual完整Schema/staging验收。
- 001 SQL SHA保持7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8；三生产文件规范化gofmt相等，新测试唯一EOF末换行缺失按三次编辑上限保留，不宣称rawgofmt全绿；scoped diffcheck0/diagnostics0。
- 下一主代理fresh Linux validator SHA→唯一owned恢复库DDLfirst/repeat→生产完整11表15FK/索引/列/旧引用严格gate PASS＋actual26输出标签→旧摘要/465PDF不变→exactcleanup；未PASS不得mainDDL/部署。详见最新报告顶部，旧actualFAIL历史不删。
- [纠正 - 2026-10-03] 下方11:43:59Z“源码仍无alias”是修复前事实；本轮local已补齐且全量通过，远端fresh gate仍未复验。七默认skip完整名称见最新报告，不用前轮sourceLock实证改成本轮0skip。

## 2026-10-03 第一阶段交接只读复核（11:43:59Z，仍BLOCKED）

- 新请求按最新磁盘receipt续作，不重复已经达到三次上限的Schema演练。实际SSH strict known key/BatchMode/ConnectTimeout10成功，hostname/user匹配vm-ubuntu-go-dev/liming；APP_ENVproduction、REPORT_EFFECTIVE_ENVstaging，三服务active/healthok，READONLY_RECHECK_EXIT=0。
- 既有112853_fd24d4ee35a9备份root0700、所有文件0600；三归档SHA/gzip/tar复核通过。两后续演练first/repeat签名、legacy前后及PDF前后清单cmp均通过；两cleanup receipt均EXIT1/REMAINING0/ERRORS0。五确切owned schema0、三payload路径0、main mng0。
- 生产Schema源码仍无显式metadata别名，真实Schema FAIL未解除；本次没有新建库/备份、DDL、源码修复、部署或主库写入，没有重跑Go测试。下一步由backend worker按RED→最小投影别名修复→本地全量验证后再fresh真实gate复验；不得把只读终验exit0称第一阶段PASS。

## 2026-10-03 SSH恢复／第一阶段真实演练阻断（actual Schema扫描）

- SSH真实恢复hostname=vm-ubuntu-go-dev/liming；APP_ENVproduction但REPORT_EFFECTIVE_ENVstaging，三服务active、MySQL8.0.46、内外health200/ok。下方SSH timeout是先前阶段事实，现已解除；本阶段没有main迁移/部署/production操作。
- 新root0700/文件0600完整备份=/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9。DB含routines/triggers/events、single-transaction gzip12793472bytes/SHA9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1；旧server/dist/configs/alltemplates及完整tmp/uploadPath/PDF/system归档gzip/tar/SHA通过，465PDF清单前后完全一致，备份永久保留。未chmod共享旧目录。
- 三次有界独立恢复（首次失败后两诊断复验）；后两DDLfirst/repeat均11表15RESTRICT FK、列/索引/FK签名不变SHA1020f16a37f84f9fbecdfd8c425969def679af26ed2eb4fb0fdea4a1b460d4e6，但fresh actual CheckRuntimeSchema全部FAIL，不能PASS或推进mainDDL。
- 已证实MySQL驱动metadata列标签TABLE_NAME/ENGINE大写，生产tag小写且SELECT没有AS：GORM rows11但name/engine有效0；独立只读显式小写alias诊断有效11。生产门禁原样保留返回data rejected；后续列/索引/FK完整校验尚未执行。需backend worker先RED大小写真实标签再最小别名修复及全量/fresh真实gate复验，不能放宽Schema/ALTER旧表/skip。本轮不改业务或原测试/DDL。
- 原source-lock测试fresh Linux testbinary实际跑两次，各4pass事件/0fail/0skip，commit/rollback/MVCC-currentread；普通第二连接UPDATE真实1205、释放后成功。独立空库+root临时localhost→unixrelay，无创建用户/grants/env/service变更，不等于完整submit/report并发验收。
- 旧12表dump SHA前后f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e；计数67/70/1487/134354/294628/1348/27不变，主mng0。两个002各140/700且ASCII，qu最大4/5、option12/14，V错误0，repoJOIN140/140与examJOIN7/6无1267。旧server/index SHA与备份一致；465PDF字节不变。
- 五确切owned schema最终remaining0、cleanup errors0；三上传资源/唯一临时目录精确删除0，临时cmd源码删除，正式备份保留。新无凭据scripts/db/management-traits-staging-phase1.sh保存流程，bash-n0；二进制仅独立验证器/test，不部署main。后续终验曾有只读SQL括号1064及PS末行CRLFexit127，需最终清洁只读复核receipt，失败如实记录。完整证据/后续部署门禁见最新本地报告顶部。
- 本地5962/0/7和355前端/模板合同是上轮基线，本阶段不重新声称全量已跑。新增实际source-lock环境通过不回写旧全量7skip为0；完整staging仍BLOCKED，不称全部DONE。
- [终验补充 - 2026-10-03T11:40:41Z] 去CRLF的tr native引号方案实际syntaxexit2、未写；最终Node spawnSync直接stdinLF后PHASE1_FINAL_READONLY_CLEAN_EXIT=0，五确切schema0/remotepayload0/mainmng0/三服务active/healthok/backup三SHA OK。临时源码删除初次未落盘、不空保护停止；最终按exactcleanup授权删除已核本任务源码/空目录及四临时二进制，localremaining0。doc/scriptdiffcheck0/diagnostics0；actualSchema仍FAIL，未自批修复或部署。

## 2026-10-03 已授权staging第一阶段预检阻断（10:12:59Z）

- 已完整读取记忆/最新本地报告/001 SQL及实际config、Schema/source-lock测试。只执行第一阶段只读SSH预检；既有liming/known key/StrictHostKeyChecking=yes/BatchMode/ConnectTimeout10/ConnectionAttempts1，两次均连接前timeout、exit255，远端命令未启动；一次初试＋一次有界重试后停止。
- 公网health实际HTTP200、{"status":"ok"}；不能代替主机身份、当前REPORT_EFFECTIVE_ENV、MySQL/mng0tables、父列或002 source/JOIN复核。用户归属为staging，历史APP_ENVproduction+REPORT_EFFECTIVE_ENVstaging本阶段未重新读取，不凭APP_ENV称prod；此前SSH restored交接被本次实际不可连接结果限定纠正。
- 完整受限备份、唯一ownership临时schema恢复、DDL first/repeat、fresh actual CheckRuntimeSchema、真实双connection source-lock、exact cleanup均未开始，无新备份path/SHA或清理PASS。未远端写/主element迁移/oldtable ALTER/env/service/dist替换/production；未改shared backup权限或删正式备份，只追加既有报告和本记忆。
- 本地5962pass0fail7skip/build/vet是上轮基线，本阶段未重跑。主代理恢复SSH后从readonly预检重启已授权第一阶段，backup失败stop；第一阶段未PASS，不进入第二阶段。第二阶段仍需fresh Linux/front build完整SHA及guard/drain→mainDDL。详见management-traits-local-implementation-20261003.md顶部阶段receipt。

## 2026-10-03 002旧列兼容Schema guard本地完成

- 仅精确旧opaque边支持source qu/answer NOTNULL varchar64→旧paper_qu_answer.qu_id/answer_id NOTNULL varchar32，均utf8mb4_0900_ai_ci；仅repo.id NOTNULL varchar64 utf8mb3_general_ci→两repo_id NOTNULL varchar64 utf8mb4_0900_ai_ci。其他边、新11表15FK/索引/NULL/collation/UUID36下限严格不放宽；无共享旧表/DDL/env/前端/模板/API修改。
- 缓存typed字节预算+ASCIIOnly；窄旧目标先约束源父实际ID再沿引用传播，31/32 ASCII通过/33拒绝；兼容Unicode拒绝，repo64 ASCII=64字符/64字节，不将mb3的192最大编码字节当64字符容量。正常既有UTF8 opaque字节合同保持，不扩大generic Unicode。mapping/options/selected/Raw source/loaded/checked CASE均消费同门禁；组卷全部源ID在首个INSERT前预检，原JOIN/LIMIT/hash/140冻结/批量策略不改。
- 主代理真实baseline5349pass0fail7skip→新RED45fail→最终Go5962pass625顶层0fail7同环境skip，新增613pass/10顶层；聚焦5250pass0fail2环境skip，解析错误0，build/vet/diffcheck0。真实sqlmock140/700writer及公开完整run只读通过，非法writer零INSERT。独立只读作用域复审PASS；不等于实际MySQL/目标环境验收。
- 本轮无SSH/DB读写/DDL/部署，用户当前staging源数据/旧JOIN实证仅按交接使用，未本轮复查；7skip保留，未删failedtest。Schema2轮/writer1轮/新测试3次编辑后停止试改，源码全部apply_patch，不创建现代化framework/task。详见本地报告顶部。
- 最终格式只读检查：两生产Go忽略CRLF与gofmt一致；新测试唯一EOF末换行缺失，三次编辑上限后不再修，非完整gofmt GREEN。代码/文档diffcheck0，八文件diagnostics0；不影响已执行5962pass/build/vet证据。
- [纠正 - 2026-10-03] 下方“所有真实引用必须同charset/collation且子>=父容量”的历史规则，现在仅上述精确旧边按actual ASCII目标预算兼容；新FK及其余引用不变。远端11表15FK安装、完整staging仍待main后续独立授权/执行。

## 2026-10-03 最新本地汇总／SSH阻断 receipt（仅文档收口）

- **PARTIAL，远端不能完成**：本轮20.200.136.133 staging既有liming／密钥、BatchMode、ConnectTimeout10三次SSH timeout，各exit255，remote command未启动；public health HTTP200 ok。远端备份／临时Schema／DDL／部署／datasetcleanup均未开始，11表15FK真实未验，production未操作；本次仅五既有文档追加，不重新SSH或改代码／格式。
- 最新独立BuildValidator交接：Go5349pass／0fail／7环境skip，前31files355pass／0fail，Windows／Linux build和vet GREEN。事件含子项，非覆盖率、非all E2E。Linux SHA前缀0aa1e764…是格式only前构建，不能用于当前部署，部署前fresh source重建完整SHA。
- 本轮主代理已真实跑原source10／0fail-error-skip exit0；LO-compatible17／0fail-error-skip，68.156秒exit0，实际SVG／pixel／六图／36长文本／13详情。最新production bundle SHA98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51；续答mock47PASS／0fail／0skip，四组合560save／4download／8submit GREEN，仅mock API／PDF，不是real DB或真实签名／服务器报告验收。
- sourcebundle事务锁、锁后expiry复核401零写、Worker keyset10batch绕fail、singleton DI／真实引用capacity、管理员5min续答及冻结deadline不改已本地实现并unitPASS。下方历史“未实现／停止”保留，由此限定纠正为codeclosed；环境仍未验，不写全部DONE。7skip完整名称与显式enabled本地PDF PASS区分见最新本地报告。
- 格式scope79文件／68格式调整、非EOF实质格式diff0，规范化代码／注释未改；仍CRLF79 raw差异及score测试3处额外EOF空行，三轮停止，不宣称rawgofmt GREEN，不再改格式。
- staging全功能、测试部署及确切临时Schema创建／清理已授权；恢复SSH→readonly preflight→受限完整备份→临时Schema演练→guard排空→DDL first/repeat→fresh build TEST scope部署→四真实组合PDF/SQL及expiry/revoke/concurrency/resume→marked IDs清理／终验。不能跳过、不重新要求无谓部署确认、不碰production；详见[最新receipt](management-traits-local-implementation-20261003.md#L3)。

## 2026-10-03 两安全项新授权完成：共享DI与真实引用容量（本地GREEN）

- 未验证边界：未访问真实DB/SSH、未执行DDL或部署；前端303仅用户交接基线，未改/未重跑。全量7skip不算环境通过，其中新增来源锁MySQL用例属于并行工作，不归本切片成果。
- candidate/tester/runtime handler可选注入同一实际pool/secret/JSON预算服务，默认各handler生命周期一次；router预检后发布，main HTTP与Worker共享。错库/错密钥/错预算/显式nil注入失败关闭，初始空配置不能通过后续改cfg复活；fresh装配重新读Schema。
- Schema保留11表所有列/索引/15FK严格合同，增加真实legacy引用列容量/NULL/charset/collation签名；exam/profile、candidate/tester、多态participant、repo/qu/relation/option引用目标不得小于父容量。exam仍外部opaque，不强制UUID；生成父paper/pq/candidate/answer桶需要36..64。实际ID按UTF-8字节和缓存父容量核验，含query/JOIN/写入/加载值/JSON映射与选项；Raw源加载显式校验，公开LoadValidatedRun前置同Schema。
- 私有GORM registry共用原pool，不污染legacy DB；s.db构造后不替换，atomic发布只读容量。GORM1.25.12 Set返回clone0，直接保存会污染后续请求；Set后Session{}使用clone2复制Settings并隔离Statement，不能用NewDB:true（clone1丢Settings，门禁失效）。已真实RED→GREEN；非法UTF8夹具不能经过JSON编码后假称仍非法，合法U+FFFD不得禁用。
- 主代理独立focus2546pass/0fail/0skip；最终全量5290pass/0fail/7skip（605通过顶层，事件含子项），JSON解析错误0，TEST/BUILD/VET退出0；独立最终review PASS，关键10Go diagnostics0。真实Gin两请求及candidate/tester/runtime六请求完整预检querycount=1、错误粘滞/fresh刷新、错注入拒绝均通过。
- 本轮schema生产3/3、runtime构造3/3（含两次经证据修正Session）、新增references测试3/3；停止新增改码。此前局部receipt中的“缓存/actual引用未接”已由本条限定完成纠正；并行write/report/撤销/resume不由本条宣称完成，完整产品/真实环境仍需独立验收。详细消费方和证据见最新本地报告顶部。

## 2026-10-03 新授权父键容量局部修复 receipt（整体未完，明确停止）

- 用户重新授权父键容量／撤销并发／跨请求schema缓存＋独立管理员限时续答；本worker实际仅容量切片，一轮生产validator＋一轮兼容矩阵：paper/pq varchar36..64、exam1..64；11表列/索引/15FK/collation/cache严格不变，不改DDL或原70短容量RED。旧兼容1/32正例与新需求冲突，保留case改为明确拒绝，不skip/delete。
- 独立RED71fail exit1→全量4642pass/0fail/6skip exit0，UTF-8 JSON解析错误0；go build -o bin/server.exe ./cmd/server及go vet ./...退出0，两Go诊断0。事件含子项，不推测覆盖率。原6环境skip未变；前端303仅用户转交基线，未重跑。
- **仍不可部署**：实际exam/profile引用ID与缓存父列长度检查未接；bundle撤销竞争、统一paper/bundle事务锁、共享actual cfg/DB singleton DI、管理员短purpose续答签发/消费尚未实现/验证。短期在已验证容量切片后明确停止，不以all-Go-green代表安全闭环，不无谓scalar churn。
- resume政策已明确：不按手机号匿名恢复；有效且唯一candidate/paper/exam绑定；retired已有卷允许、revoked拒绝；到期不得继续答，凭据只看状态/完成，不resetdeadline、不unfreeze。新endpoint未注册，不提供假路径给前端。合同字段及P0 case清单见management-traits-local-implementation-20261003.md顶部，后续主代理再实现并委派frontend。
- 无前端/客户源/Word/TESTlabel/font/star/footer、DB/SSH/SQL/部署改动。下方71fail属于前轮历史，容量失败已局部解除，其余安全阻断保留。

## 2026-10-03 002后端身份／Schema续作（BLOCKED，三轮停止）

- 本轮仅backend18Go文件：configured九字段含age/degree/major/stuFlag、typed空串、取消固定name+telephone；新candidate配置写入/tester配置投影；已有paper复读冻结140/700/有效唯一owner，profile草稿/资料变化不覆盖frozen/绑定/deadline，身份五键及八API签名不变。candidate仍需有效participant凭据，过期／遗失凭据匿名续登未实现。
- 11表15FK DDL/models未改。新真实sqlmock metadata矩阵覆盖全部列/索引/FK/error/cache；畸形bigint与text collation拒绝。身份完整门禁在事务前，但每HTTP新service缓存不跨请求，七表scope探测尚在。
- 独立review CHANGES REQUESTED：bundle非锁定读取有撤销→写提交竞争；paper/pq父键可短于UUID36；身份cache实例寿命warning。新增父键容量RED保留，不skip/delete。worker审计identity HTTP5轮/service4纠正、旧HTTP测试3轮/schema行为测试3轮，已违反/达到三轮上限，停止追加修复，需用户明确新有界授权，不拆文件绕过。
- 主代理UTF-8可靠全量4571pass/71fail/6skip，TEST_EXIT1、解析错误0；唯一失败顶层TestBugMTSchemaGeneratedParentCapacity（70短容量子项）。BUILD_EXIT0/VET_EXIT0。6skip原5+MNG_REAL_LO未启用；初次有解析错误计数作废。聚焦GREEN不代表全量GREEN。详细changedfiles/contract/evidence见management-traits-local-implementation-20261003.md本轮backend顶部。
- 未前端/SSH/真实DB写/DDL/部署/production，客户源和旧PDF不动。不能部署/DONE。下方前轮“身份/配置完全未做”现由此限定代码切片部分覆盖，未解除并发撤销/容量/cache/安全凭据恢复及真实环境门禁。

## 2026-10-03 002完整授权续作（后端本地PARTIAL，未部署）

- 本轮explicit政策已定：25min/20min提示服务器时间；未到期缺题reject，到期140completed否则incomplete无正式分/报告；exam.end只禁新开，existing原deadline；startup/restart worker立即扫不reset；retired禁新开但已有按frozen继续；source审核撤销禁新写/new report；仅admin显式generate，不auto。当前库原文00202 V67/V96仍独立db-current不覆盖客户/旧源；205原词仅TEST，formal/prod禁止，staging先备份/迁移/测试部署/临时finally清理由主代理执行，本轮没SSH/DB写/部署。
- 已新增11表显式MySQL5.7/8 DDL、15 RESTRICT FK，完整列/type/NULL/PK/unique/5非唯一read-expiry索引/FK/collation缓存门禁，公共入口事务前每service实例一次；全缺失关闭、部分或错误拒绝、重启刷新，不AutoMigrate/回填。DDL父exam/paper/pq id的collation必须一致，不一致停止、不改旧键；真实MySQL首次/重复执行未验。
- 后端原八503接真实freeze/profile/create/detail/fill/submit/result服务；新增140题＋700桶批量原子组卷，真实owner、V/显示序分离、冻结25/20元数据；既有卷释放owner锁再锁paper避免锁倒置；retired续答/审核撤销区分；incomplete详情仅审计、报告仍completed-only。新Worker接主服务启动/取消，只扫new_creation，首扫立即，同Submit，不触legacy、不auto报告。
- 新独立TEST revision/current/audit模型与admin generate-test/view/download，显式run/report ID、mode/test_title/test_label、UUID私有不可覆盖文件、90秒同context、事务外独立Word→共享LO，末次同paper锁复读DTO/source，metadata/current/audit原子；失败新文件清理、旧PDF保留；ID下载同句柄有界size/SHA/DTO复核，路径不回显。HTTP报告ENV仅明确local/staging，unset/production拒绝。
- MT-LABEL-02真实PDF无TEST先RED；v1流式用途标记推动原大标题裁切，保留失败副本。最终TEST v2由精确a986候选确定性新建page-relative用途drawing（原88＋2＝90），去新增run后原document字节一致、其余OPC不变；客户原件/原候选SHA不变。v2 SHA05c55e77...，530193bytes，独立configs文件；旧v1不覆盖。
- 真实205原词DTO（明确synthetic/sqlmock）→Go→LO完整原文PDF A4 9页653016bytes/SHA3eca22f7...；TEST/不可人才决策可提取且封面实图可见、原大标题恢复、六图页可见。全文layout假失败因左栏“测评结果”插入右栏跨行，raw读序26段+总体全文逐字GREEN；不改客户字、不删除固定标签。非真实数据库受测者、不宣称全九页/目标环境验收。
- 最终Go2569pass/0fail/5既有skip（含子事件，MNG_REAL_LO_TEST=1）；Windows build/vet0，Linux本地build0但最终包需索引补强后重建；原source-layout10/0/0/0实际通过，原兼容17/optimized23未重跑防覆盖证据。前端既有26文件165pass/build:prod0，未新增前端；详细测试及消费方见management-traits-local-implementation-20261003.md顶部。
- **P0未完不能部署**：前端002profile/答题/admin链和旧result2分流未实施；身份admission仍阻end后/retired重登录、事务内7表probe未接11表缓存；configured字段仅现有5项且强制name+telephone，age等/空字符串完整合同未完；本机无mysql/mysqld和3306/23306，真实DDL/并发/四组合HTTP/UI/restart未验。已有valid token恢复通过不是完整登录闭环。无新Excel/批量/历史重算/高级current，不放宽004。
- [最终构建补充] Schema五非唯一索引补齐后Linux amd64本地新产物49683564 bytes/SHA e7f8e40a1194bf77cbf4f4e699616a53b5f1efc874614c5730a7d23e0c7b5065，FINAL_LINUX_BUILD_EXIT=0；关键Go/报告memory诊断0、代码diff与文档diff0。仍只是不可部署的本地PARTIAL候选，没有运行身份/UI/DB环境门禁证据。
- [纠正 - 2026-10-03] 下方“没有DDL／八handler全部503／无report core”是上一轮历史；上述代码已落地并本地GREEN，但不等于完整用户授权链完成或可部署。图表旧修复仍保留；本轮未创建现代化状态框架，无外部agent完成声明。测试经验记录在既有docs，不建新规则。

## 2026-10-03 MT-WORD-RUNTIME-01独立有界图表修复

- 用户重新授权继续诊断；一轮最小生产代码修复，仅002独立Word数字槽校验：实际总体/任务anchor有4个w:t，第4个为空白，其余模块3个；manifest只绑定第2个值。六业务title/name及rId12～16/rId22→chart1～6正确，不是key/OPC错误。兼容3/4并核额外空白，只替值，不改模板SHA/样式/数据语义/测试用途代码/205条文案或运行门禁。
- 保留TestManagementTraitsTestWordIndependentValueOnly并增强不同分值六图/五槽、chart非值XML/正文非文本XML/其他OPC字节不变。原生RED charts exit1→GREEN；聚焦1813pass/0fail/0skip；Go全量2525pass事件/0fail/5既有环境skip、明确BUILD_EXIT=0，诊断0。编辑器编辑后未发现测试，不冒称编辑器GREEN；PS输出JSON须设UTF-8，默认乱码解析失败的统计作废重跑。
- 本轮原候选17项含真实PDF及source-layout10项均真实通过；optimized23未重跑。真实Go混合分值及全50稿经本机LO26.2.5.2输出A4 9页；全50第3页五环图＋13维柱线图实际可见，混合稿五标签25/75/100/0/28.85及13维常模正确，可见占位符/U+FFFD为0。候选a986ba0f...、客户DOCXc82c2dc0...、XLSXb0498249...SHA均保持，不用变模板解决失败。
- [纠正 - 2026-10-03] 下方“charts失败、无新PDF/未调用LO”为上一轮停止事实，本次仅限定图表bug已解除；整体002仍PARTIAL、DB/HTTP/UI/持久化/目标环境未验、formal/运行未启用。新发现：既有DOCX测试标记不在PDF文本层显示，按only-chart范围不改其定位逻辑，需独立确认；不能称报告可交付。LO输出prefix警告而实际全50转换exit0，不修改外部配置。
- 无SSH/SQL/DDL/DB写/客户原件或活动模板变更/部署。细节与消费方清单见management-traits-local-implementation-20261003.md顶部；本轮只两Go文件与本bug相关文档，临时产物在ignored tmp。

## 2026-10-03 002本地部分实现／渲染失败停止

- 新授权：当前题库原文冻结且00202V67/V96独立db-current版本，不改旧源题；完整002测试链获批，保旧result/PDF、不自动历史重算；staging备份迁移测试范围部署及验后临时数据清理获批，worker本轮仅local，production禁止；现有文案只测试口径、formal关闭。未决时段/用时/离线/退役/epoch等不默认启用。
- 实际router已安装既有ManagementTraitsLegacyScopeGuard于JWT前。真实Setup三旧公开写入口403、两管理读取未登录401，RED三失败→GREEN4pass事件；候选身份既有分流未改，不等于完整新链接通。
- 新SHA锁定客户XLSX内容加载器提取195维度条件文案＋5总体评价＋5三段总体建议＝205逻辑规则；强DTO复核run/13维/4模块/receipt及冻结身份/提交时间，仅纯适配。新两测试本地PASS，工作簿SHA b0498249...，不修客户文本、不作formal批准。
- 独立002 value-only Word实现未完成：新真实模板契约在charts阶段失败，三轮上限后停止。候选实测88唯一SDT/5数字槽/6图；ZIP含合法空目录、body先logo表后标题。保留失败测试，未生成新PDF、未调用LO或开放报告。
- 基线0021806/0/0；最终Go全量2512pass事件/1fail/5既有skip，ALL_TEST_EXIT=1；唯一失败TestManagementTraitsTestWordIndependentValueOnly。Go Build及明确复验BUILD_EXIT=0、前端26文件165项PASS/build:prod退出0（2既有warning）、诊断0。不能标记全绿或可部署。
- 无SSH/SQL/DDL/DB写/客户原件模板修改/前端代码变更/部署。本机未找到mysql/mysqld命令且3306/23306监听0；create-paper、完整Schema签名、真实新HTTP/UI及report持久化/ID下载仍未实现。详见management-traits-local-implementation-20261003.md，主代理需接管阻断，不把既有17+10+23上游证据称本轮重跑。

## 2026-10-02 原稿保真002 LibreOffice兼容第三模式

- 用户在严格原稿模式显示失败后明确允许numeric框/legend/pageflow必要兼容调整；仍保原件/font字号/color/fixed字/rPr/star/footer，不重启删星/PAGE-only。新增显式libreoffice-compatible第三模式，独立lo-compatible-template/manifest/demo/PDF；strict/optimized及旧62产物hash不变，未DB/GoVue/runtime/活动配置/部署。
- mode缺失RED10项1fail9errors、strict实际PDF7项57失败子项；新完整17项GREEN、旧10+23回归0fail/error/skip，独立scope review PASS。主代理独立结构10项通过；实际PDF须加--visual-boundaries，不能把仅模式flag10项当完整渲染。五组数字5/5，几何/像素避碰25+25，36长文本、两稿各13详情建议同页，原sample59.45保持只demo替值。
- [完成补充] 主代理独立完整真实PDF17项于64.233秒通过，CONTRACT_TESTS=17 FAILED=0 ERRORS=0 SKIPPED=0退出0；并非只读子代理证据或仅结构10项。
- 兼容templateSHAa986ba0f.../demo09ed491a...，原稿c82c2dc0...不变；Word10/11页88SDT close0不变，LO9/9。仍概览及第7页留白/star总页数原样；目标/WordPDF/正式内容/运行接入批准未关闭，只staging候选测试。详细证据lo-compatible-validation.json和verification顶部。

## 2026-10-02 客户原稿优先派生002程序模板（当前选择）

- 用户重申客户DOCX为模板权威，并明确选择客户原稿版式优先；覆盖此前删星/PAGE-only与标签外移/扩框/legend空间/分页优化决定。原布局/固定文字/样式/star/footer保留，仅绑定清外链必要转换；原件不改。不冲突的post/submittedAt和13维平铺语义保留，批准仍仅staging候选测试，不部署/正式启用。
- 新source-layout显式模式输出独立management-traits-002-source-layout-template.docx/source-layout-manifest.json/source-layout-synthetic-demo.docx于generated候选目录，未接runtime/活动配置。旧optimized默认及五产物hash不变。新增10合同主代理实际PASS 0error/fail/skip，旧23回归通过、独立review PASS；88SDT/5slots/6charts及原文样式/布局/star/footer/media完整白名单验证。source sample59.45不被cache59.66替换，空模板示例非评分基准。
- 原稿主代理重新实测660351 bytes/SHA c82c2dc0...不变；新附件无完整新落盘路径，未声称bytewise相同。新模板SHA465358b1.../demo efa4c253...；Word原稿/模板/demo 9/9/10页关闭SHA不变，LO8/9页。原框数字拆行/人际裁切/legend截短/演示尾字孤页复现，按原稿优先未再优化，不能宣称视觉验收通过；目标环境/客户验收/正式内容及runtime门禁保持。

## 2026-10-02 管理特质用户决策（仅staging候选测试）

- 本轮问答确认产品正式名“管理特质”；职务使用岗位post、测评日期使用提交时间submittedAt；暂不新增年龄/用时/总体常模位置；13维图接受平铺、不恢复四模块层级。postdate与flat仅语义确认，不代表运行实现，也不改变四模块评分聚合。
- 删除星形、仅显示中文等级；页脚仅当前页码、不显示总页数。star删除/pageonly尚未在候选执行；目标服务器PDF作为主要排版验收依据、暂不固定页数，本轮未执行该验收，不宣称实际模板已变动。
- 内容及心理测量负责人均Liming；用户先给20291001，经明确核对选择2026-10-01。对象范围明确为“仅staging候选测试使用，未决内容不视为正式批准”；不称已正式双批准，不选择production，不授权模板激活、部署、迁移或运行接入。
- 错字/频率/统计声明及V67/V96不擅自修订，当前questionnaires差异不改DB。仅更新本文件与[候选验证记录顶部决策表](management-traits-word-candidate-verification-20261001.md#L3-L28)，无代码/DOCX/JSON/其他账本/新规则/DB/SSH/部署；下方旧pending保留，仅本轮明确选择的同事项待确认状态被覆盖，未决内容与正式启用门禁不解除。

## 2026-10-02 管理特质002候选标签避碰/模块分页切片完成

- 用户明确选择继续优化本地候选版式。仅既有两Python脚本/generated产物：左模块标签左96pt、右模块右100pt、总体下108pt，环图矩阵/组合图/字体字号配色业务数据原件不变；13详情单模块cantSplit/keepLines，末行keepNext=0防全链粘连；建议局部同页、取消摘要后强制分页。无DB/GoVue/配置/活动模板/部署。
- 前次SHA精确重建后4视觉测试48子项RED；主代理完整实际PDF复跑23tests0fail/error/skip退出0，独立有界审阅PASS。五组数字5/5、SVG包络25标签无交叠、144DPI标签环色像素0；候选/demo各13详情和13完整建议同页，摘要与详情共享第4页，不再摘要孤页。
- 最新Word只读10/11页88控件关闭SHA不变，LO26.2.5.2为9/9页；Word PDF/段落分页未验。概览下部及第7页两完整详情后留白仍有，不全局压缩；页数/星级页脚分类/字段语义/正式内容/目标环境/运行接入仍待确认，不能激活。
- 最新candidateSHA1f236407.../demoSHAbe2ec711...；原件c82c2dc0...不变。详细证据layout-slice-validation-20261002.json及候选验证文档顶部；下方前次相交/摘要孤页/跨页及旧SHA页数均属历史，不删除。

## 2026-10-02 管理特质002候选局部排版门禁通过

- 用户明确批准继续调整仅候选标签框/图例/分页粘连；修改既有两个候选脚本及generated产物，未修改客户原件/字体/字号/配色/业务数据/GoVue/DB/活动模板/配置或部署。五标签框原中心保留，宽120点高总体64/模块44、边距0；legend仅manualLayout，总体/两栏摘要局部keepLines/keepNext。
- 扩展RED17顶层10 failures；主代理真实LO重跑18 tests/0fail/0error/0skip退出0，独立CodeReviewer PASS。0/25/75/100/28.85每组5/5完整数字，图例完整不撞100标签，36长文本完整且无尾字孤页。旧停止记录保留，本轮批准的新范围已解除该三项局部阻断，不宣称整体视觉全绿。
- Word实际只读候选9页/demo10页88控件，关闭SHA不变；LO26.2.5.2为8/9页。实图仍有环图文字与环体相交、摘要页稀疏、责任心详情跨页；不顺手改全局布局。正式版式/页数/字段语义/星级页脚/内容批准/目标服务器/运行链待确认，候选仍未激活。
- 当前candidateSHA8b929da1...、demoSHAcbdcf73c...，原稿SHA c82c2dc0.../XLSX b0498249...保持。详见docs/management-traits-word-candidate-verification-20261001.md顶部10月2日更新；历史SHA不再代表当前产物。

## 2026-10-01 管理特质002 Word候选（视觉未通过，停止试改）

- 用户同意评估收敛后授权开工，仅本地Word候选；新增scripts/tools/build-management-traits-002-word-candidate.py及scripts/test/management-traits-002-word-candidate-contract-test.py，产物docs/generated/management-traits-word-candidate-20261001/。不改DB/源题/GoVue/活动模板/运行配置/客户原件，不部署。详细证据见docs/management-traits-word-candidate-verification-20261001.md。
- 88实际SDT＋5对象数字槽、6独立anchor、literal零外链/公式；不新增年龄/用时/总体常模位置，post/日期语义及层级分类/页脚/星级仍待审。候选源SHA c82c2dc0...不变。首次inline造成排布漂移，已改为原坐标anchor并保留背景；结构通过不等于视觉通过。
- 最终主代理14结构/数据测试通过，但真实0/25/75/100分PDF四子项均4/5完整数字token，VISUAL_CONTRACT_EXIT=1；总体28.85拆行、人际数字裁切、图例截短与100标签重叠，长文案第4页仅尾字。三轮视觉修复后按规则停止，等待确认候选标签框/图例空间/分页局部版式调整，不准交付/激活。
- 最终Word只读候选9页/demo10页、LO26.2.5.2为8/9页，关闭SHA不变；未批准页数或目标环境验收。候选SHA76b58e95...，demoSHA6685578e...；初版Word10/11与LO9/9已被替代，不混用。本轮无Go全量重跑；脚本诊断0。正式报告/历史重算/部署仍关闭。

## 2026-10-01 管理特质002生命周期/来源门禁设计补充

- 用户在记录差异、不修DB/程序逻辑不变后要求继续；本轮只在设计§18收敛生命周期/真实来源门禁，不新增代码/DDL或查询环境。现库存量原文与客户管理修订稿作为两种待选择来源，不替用户选择、不共用题本版本/SHA、不放宽138/140文本差异。
- 明确四门禁：S2B-D数据自洽、可信DB读取/历史证据及唯一真实participant、执行策略/锁内状态时限及旧写保护、正式内容/Word批准；verified布尔或合法SHA不能代替真实加载/审核。draft-reviewed-retired、profile冻结、paper完成/不完整、historical证据reviewed、run不可变和capture排空保护均为逻辑提案，未加物理状态或默认实现。
- 结果一致零写复用、冲突拒绝不修、纯重渲染不建run、证据不足只可另批保旧PDF；来源submission/historical_recompute与paper来源不混用，人员submitted_snapshot/captured_at_recompute区分。时间/窗口/离线、来源载体、审批/历史策略等仍待一次性选择。
- 独立审阅发现已开卷遇retired续答/交卷分支遗漏，已补为待确认，并明确退役与可信性审核撤销不等价、不得换已冻结题本。未修改程序逻辑/客户原件/DB/部署；本地文档链接及格式检查用于设计一致性，不冒称运行验证。

## 2026-10-01 管理特质002题本差异处理决定

- 用户要求记录差异、不做数据库修正、程序逻辑保持不变；核对报告§8追加决定。00202 V67/V96文本差异及Excel K6标签错字保留，不更新DB/客户原件/源题/题序/选项/公式或引用守卫。
- 两文本差异不改变V/维度/正向/raw/公式，同一原始答案的算术不变；不能宣称措辞对受测者作答/测量无影响。不修DB不等于批准连缀文本或等价，138/140仍如实记录；后续明确题本来源/版本前不放宽S2B/S2C精确文本/hash匹配、不伪造mapping/历史证据。本步仅文档，无DB/代码/重算/部署操作。

## 2026-10-01 管理特质002 staging实际源题核对完成

- SSH恢复后用户继续，实际element只读事务查询四源表metadata与两002题库/1400选项行，查询退出0；无DB/配置/代码/客户原件更新、无重算部署。完整证据见docs/management-traits-staging-questionnaire-verification-20261001.md。XLSX SHA仍b0498249...；原始ASCII/HEX SELECT行按LF拼接指纹afec65695a52914197770a53c3d476e394a68c8c5ae0a493bf18105cd2392619，仅内存保留不作历史证明。
- 当前00201 repo2019583026766479361/00202 repo2019583796408680449各唯一、140单选/140源ID/140关系，sort连续1～140，700唯一选项且每题raw1～5对应不符合→很符合；两库共享题ID0。q.content=V号，实际陈述在q.title，不在选项content；源选项无sort，未浏览器核序。
- 对Excel H全局V逐题，B先去维度内编号：00201题干140完全一致；00202为138一致，V67/V96仍使用基层版；现库两题本140条题干互相相同。客户管理B68连缀文本待确认、B97下属措辞未落库，未自动更新原题。
- dimension_id/scoring_direction/question_code/competency_question_type均NULL，不具新版目录。is_right每题唯一、100题raw5/40题raw1，后者V与客户40反向差异0；旧保存raw和公式一次反向不变。两sheet13公式右侧26/26与本地旧代码一致，140terms唯一/40反；K6左名实际人机敏感性是新发现标签错字，按名称初次miss已复核，不是公式差异。
- [纠正 - 2026-10-01] 下方“实际题本未查询”网络阻塞历史已由本次当前源库核对解除；仍未查历史paper/作答时备份或服务器代码hash，不能把当前源码/题本一致当任意历史证据。下一步需确认V67/标签修订与V96启用版本，不擅自改DB/放宽已引用题守卫。

## 2026-10-01 管理特质002 staging实际题本核对阻塞

- 用户明确选择staging实际数据库只读核对00201/00202。已确认当前模型列与查询路径，并以既有liming/SSH密钥尝试元数据及题库查询；首次ConnectTimeout=10超时，第二次ConnectTimeout=10/ConnectionAttempts=2仍超时，SSH退出255。连接未建立，远端mysql/SELECT未启动，没有实际题本查询数据。
- 实际两题本数量/题序/题干/五选项分值与客户XLSX逐题比对仍未执行；不得用旧schema.sql/历史140题或S1-S2D synthetic fixture代替当前数据库证据。尤其el_qu.content与el_qu_answer.content的实际语义及qa.score实时存在性须先元数据查询确认；全局V按qu_repo.sort对照Excel H列，不能用题干维度内题号推定。
- 未改DB/网络策略/凭据/客户文件/运行模板、未部署/重算。恢复SSH后按只读事务先查四源表列及唯一repo code，再LEFT JOIN导出源题/关联/选项进行逐题核对；也可由用户提供同四表的staging只读导出替代连接。
- [纠正 - 2026-10-01 SSH重试恢复] 用户要求继续尝试SSH后，使用既有用户/密钥及BatchMode、ConnectTimeout=10、ConnectionAttempts=2连接成功，hostname=vm-ubuntu-go-dev、远端用户liming、SSH_EXIT=0。此前网络阻塞已解除；本次只验证握手/主机/用户，尚未执行数据库查询或题本比对，未改远端配置或部署。

## 2026-10-01 管理特质002 S2D只读结果一致性本地完成

- 用户确认仅local suppliedrun/dim/module一致性，无DB/HTTP/重算/部署。新增service management_traits_result_validation.go/test.go一个边界ValidateManagementTraitsStoredResult：completed及13/4数量→S2C→S1原始答案精确重建，要求140/140，核run真实声明身份/四评分版本/Q/manifest及inputhash/count和overall、13维及4模块全部事实/缓存；不修改S1-S2C/模型/旧入口、不建writer、不修记录。
- 维度integer sum/count必须与原始答案相符；decimalFromRat既有helper从big.Rat生成6位缓存、Equal数值比较允许等值不同指数而拒额外有效精度/NULL/两位常模；等级和排序/聚合源仍是精确有理数。完整key/order/name/module/成员数/childRunID及子表内唯一ID均核对，返回Rat/切片独立、数组重排不变。
- RED新函数undefined退出1；GREEN主代理原生Go8顶层+567子项=575pass/0fail/0skip，编辑器484passed/0failed为不同计数口径。全量1918pass/0fail/5既有skip，ALL_TEST_EXIT=0 TEST_EXIT=0 BUILD_EXIT=0，独立CodeReviewer PASS、诊断0；skip与S2C一致不计环境验证。设计§17/分支/覆盖历史同步。
- completed字符串/自洽不证明真实DB提交/授权/历史来源/版本批准；Source/CreatedAt/SubmittedAt/UserTimeSeconds及S2C证据/人员/采集/时间门禁仍延期。正式报告/重算入口不开放，staging/production不变；下一步仍需用户确认独立生命周期或实际题本核对切片。

## 2026-10-01 管理特质002 S2C严格解码/只读适配本地完成

- 用户确认仅本地strictstoredJSON+S2A→S2B只读适配，不DB/HTTP/重算/部署。新增service management_traits_decode.go/test.go：DecodeManagementTraitsManifest/Mapping仅接S2B canonical存储domain结构，ValidateManagementTraitsStoredInput只接调用者传入bundle/paper/140题行；不改S1/S2A/S2B/旧入口，不建writer。
- 严格tag全部必需含false、exactcase，拒duplicate含escaped同名/unknown/missing/null、类型/小数指数整数/overflow、非法UTF8/BOM/trailing、孤立UTF16代理；合法代理对和U+FFFD保留，fixednorm/dims/policy核验。对象排版/题目选项重排可接受，SHA由规范化字节重建不取任意排版字符串SHA，输出独立且输入不变。预算每字段+manifest/mapping/140options总量减法预检，调用者正数注入，上游读取前限额延期。
- adapter核bundleID/四版本/Q/SHA及paperBundle/SHA/nullable同examprofile；各题paper/唯一ID/V、方向/题干/选项/raw/合法source选择/final一次反向/评分SHA；已答三指针齐全、未答全nil、不用submittedAt判断。返回S2B输入及S1raw答案不评分。Source/Status/evidence/IdentitySource/ParticipantSnapshot/FieldContract/所有时间明确未校验，不能当真实人员归属/历史/审批/提交许可。
- RED缺API退出1；GREEN九顶层含子365/0，主代理编辑器365/0、独立CodeReviewer PASS、Go全量/S1S2C联合和Windowsserver构建ALL_TEST_EXIT=0 S1_S2C_TEST_EXIT=0 BUILD_EXIT=0。全量5既有环境用例跳过（MySQL并发/客户模板上传/FB169-170活动模板/客户模板LO页数），不计环境通过；S2C无skip，诊断0。两新文件内容gofmt对齐无BOM/CRLF保留。
- 设计§16/分支/覆盖历史同步；真实DB加载/历史证据/版本批准/人员资料及采集JSON/时限状态/合法run复用缓存/写事务/旧写保护仍待下一独立切片，staging/production未变。

## 2026-10-01 管理特质002 S2B纯校验/hash本地完成

- 用户确认纯manifest/题本映射/快照输入校验和stablehash，不导入/DB/HTTP/历史重算/部署。新增service management_traits_contract.go/test.go三个边界：CanonicalManagementTraitsManifest/Mapping/Input；不改S1、S2A、旧handler/router/model或依赖，不接运行入口。
- manifest验证140唯一V及S1维度方向、五项原始分/文字/展示序、UTF-8原文、staff/leader和四轴语法；mapping exactSHA/题本/逐V文本选项匹配及实体内唯一opaque源ID；input exactSHA/声明paper/exam/participant/candidate-tester、唯一V/pqid/展示序及同题selected/raw一致，未答空选择raw0，输出V序S1raw答案不反向。语法不证明受支持/批准，声明身份不证明真实DB归属。
- Canonical固定struct JSON/V序/raw序，保留原文/空白与DisplayOrder、stdlibJSON escaping、无indent/newline；domain manifest/mapping/input/frozen-question-v1区分；评分SHA含S1固定norm/policy，排除content/template/人员呈现/时间；数组重排稳定、变化传播、input不变/输出独立。逐题SHA含全mappingSHA，任一映射变动传播到各题hash是显式v1作用域；synthetic题干fixture不等于客户原文/历史证据。
- RED为先建test时新符号undefined退出1；GREEN10顶层含子项172/0，主代理编辑器172/0、S1S2B联合及Go全量和Windowsserver构建ALL_TEST_EXIT=0 S1_S2B_TEST_EXIT=0 BUILD_EXIT=0，独立CodeReviewer PASS、诊断0。newfile实际gofmt内容已补齐，CRLF仍保留；未宣称rawgofmt-l通过。设计§15/分支/覆盖历史记录同步。
- 待后续：实际题本/源库对照、历史独立证据/真实participant、执行版本与批准门禁、JSON严格解码与限额、S2A存储适配、既有run复用及事务/旧写保护；本轮hash有效不可当作历史重算或正式报告许可，staging/production未变。

## 2026-10-01 管理特质002 S2A本地存储模型完成

- 用户确认S2A仅本地结构/测试，不执行DB迁移/运行接入/历史重算/部署。本轮只新增internal/model/management_traits.go及management_traits_test.go：七个独立plain模型（评分bundle/newprofile/paper快照/question快照/run/dimension/module）+TableName，11有序复合唯一索引仅GORM声明。无关系/hook/init/AutoMigrate/SQL/writer/报告内容表，旧model/handler/router/service不变。
- 新profile与历史逐卷证据分开：ProfileExamID可空，paper快照独立EvidenceSnapshot/SHA、mapping、participant_type/id/字段和身份来源，不以user_id101证明归属。bundle/run仅四评分轴无content/template；题目V/display_order分开，选项/题干JSONlongtext、raw/final可空；维度整数sum/count保存精确源，分值*decimal.Decimal decimal18,6/等级/用时可空，模块不含未确认等级/常模比较。
- RED七模型undefined导致build failed；GREEN原生go test九顶层全部PASS，MODEL_TEST_EXIT=0 ALL_TEST_EXIT=0 BUILD_EXIT=0，独立CodeReviewer PASS、诊断0。编辑器runTests未发现新测试，未冒称编辑器通过，原生Go测试为执行证据。新文件无BOM，内容忽略CRLF与gofmt一致；不改S1换行。
- 设计§14补充最小存储合同及后续门禁，业务分支/覆盖历史同步。真实表/索引/FK/NULL/default/collation未安装或核验；状态/来源/hash/JSON/cardinality/immutable/唯一真实人员及缓存一致性仍需后续入口/writer校验，不把声明视为保护已生效。capture保护标记独立待实现，DDL应用前必须先补旧写/删除保护。

## 2026-10-01 管理特质002 S1纯身份/精确评分本地完成

- 用户明确确认仅S1本地实现、不HTTP/DB/正式报告/历史重算/部署。新增service的management_traits_identity.go/scoring.go及scoring_test.go：staff/leader逻辑身份不推定00201/00202映射、不激活版本、不导入题干；13语义维度/140唯一V、100正40反、已确认常模与防御性深拷贝。旧standScore2、router/handler/model/配置/题本保持不变，新符号仅被新测试调用。
- 精确big.Rat按原始1～5正向/反向6-raw一次转换，维度25×均分−25、模块成员维度等权、总体13维等权、综合常模705/13；等级90/70/30/10、排序用未舍入值，顶3/底3按固定维度同分序，全同分重叠且无等级过滤。未答raw必须0；有效不完整输入保留13维/4模块身份和计数/合计，所有正式分/等级/常模为nil/空、选择集合非nil空，不把缺题反向成6。纯函数不决定manual/timeout运行语义。
- RED为新类型/目录符号缺失导致service编译失败；GREEN9顶层测试含子项47通过/0失败，独立三个新文件执行覆盖100%语句（不是service/系统覆盖率），Go全量和Windows构建明确ALL_TEST_EXIT=0/S1_TEST_EXIT=0/BUILD_EXIT=0，诊断0、独立CodeReviewer PASS。业务分支与覆盖历史已登记；本任务不是旧bug修复，未登记为既有安全问题已修。
- 格式限制如实保留：新文件UTF-8无BOM但编辑通道保留CRLF，gofmt -l仍列出（与LF输出差异，非编译错误）；未用终端改文件，未宣称格式门禁全绿。TaskExecutor因自身场景task文件前置不能执行测试任务，已改常规代理仅建测试，未创建无关现代化任务文件。
- S1不包括实际题本/历史证据校验、版本hash、持久化、参与者权限/交卷/时限、中文报告标签/条件文案、Word模板/导出，旧代码风险仍未修复；没有staging/production变更，下一步等待用户选择独立切片。

## 2026-10-01 管理特质002设计R2修订（仅文档）

- 按完整评估修订`docs/management-traits-word-design-20261001.md`：评分product/question/scoring/norm及scoring_manifest_sha与content/template呈现轴分开，渲染hash另建；换文案/模板仅新增report revision，不迫使重算。exam profile只约束新组卷，历史paper逐卷证据、可空profile；旧PDF capture先独立保护，不伪造140题证据。
- 身份按paper+exam反查candidate/tester恰一有效关联并核历史证据，新卷原子绑定真实participant；禁止以paper.user_id101或ID长度判断。所有新查看/生成/Excel/批量目标固定显式run/report及manifest，current使用拟定最新成功attempt_seq+selection_epoch，后启动失败不阻断先启动成功；均为待确认提案，非已实现。
- 94改为候选目录而非必需数，age/userTime/overall.norm先确认位置；欠佳标签/物理选项5→1含义、总体建议三段槽及顶底摘要独立。增加旧Save/Update/凭据/匿名管理写/双PDF/导出门禁、四精确POST的参与者token合同、后台权限与资源提案；时段/到期/离线/报告触发/字段位置等仍待确认，运行入口关闭。
- 独立初审找出历史新版分流仍用profile和四参与者JWT豁免描述矛盾，已修正为paper标记/快照+显式run及四精确POST豁免；模板GET read/候选POST write/激活activate拆开，模块固定顺序明确待确认。本轮仅修改设计/记忆，未改业务/规则/客户原件/DB/部署，静态一致性不等于代码修复或批准。
- 最终独立复验为设计一致性PASS WITH PENDING CONFIRMATIONS，未发现剩余内部阻断；31个本地链接及行范围有效、文档诊断0、scoped diff check通过。未重新运行业务编译/测试，本轮只改文档；客户决策/运行验收/正式启用依然待完成。

## 2026-10-01 管理特质002完整第二轮复核

- 在`docs/management-traits-code-gap-analysis-20261001.md`追加§7～10：材料坐标/14项增量风险/全消费方清单/12组验收矩阵；保留首轮历史边界。三原件SHA再次一致，本地Go Build完成、旧002相关8项测试8/8；直接提取exam.vue原handSelect作内存mock复现answers累积（高→另一题→低仍带三项）。未改业务/测试/规则/客户原件，未HTTP写入/查DB/远端/部署，新版及完整E2E仍未验证。
- 原稿表1已存在姓名/性别/单位/职务/联系方式/日期；表4已有顶3/底3名称+摘要两栏，但0内容控件。年龄/用时仍无位置，职务到岗位映射待确认。XLSX等级L4=欠佳不是00401薄弱；题本C:G物理顺序为5→1含义，导入不能按列号赋1→5；综合评价D2:D6多段建议无动态维度替换槽，不默认插入底3。
- 新源码风险：candidate Save新struct全字段覆盖旧关联（完成且无请求ID会被拒，非空ID绕过该前置）；匿名paper前缀包含paging/save/delete；tester匿名详情/登录返回password且明文比较，匿名PUT可更新身份/密码/关联；CreatePaper与LoginForm状态口径冲突（实际常量0启用1禁用2就绪3过期），不得套规则旧状态表。
- 进一步确认匿名ShowPdf路径+匿名PdfUpload仅任意人员登记检查不是调用者授权；报告删除/ZIP是JWT门禁而非匿名，handler业务权限仍不足，已排除复审误判。服务端/candidate时间格式尾000为字面量零导致同秒同名碰撞，tester使用.000并非此缺陷。生成器接受incomplete、双Excel权限/披露不一致、Prefix路径检查不充分；均静态证据未实际利用。Qu.Save已拒已引用源题修改，不能称任意可改，但仍无完整题本/身份冻结。

## 2026-10-01 管理特质002客户方案与代码差距评估

- 新增`docs/management-traits-code-gap-analysis-20261001.md`，按V2.6完整方案形成MT-01～036证据矩阵；原方案SHA仍为`e306e791...`，只读提取214非空段。只新增评估文档/记忆，未改Go/Vue/SQL/客户原件、未访问DB/远端、未部署或重算；本轮未重新编译测试，不把前轮旧评分16/16当新版验收。
- 本轮进一步确认002两答题页Timer均注释，无20分钟提示；preview是测前准备不是答后预览。FillAnswer的002分支不强制一个合法选项，可将未知选择写成answered=1/actual_score=0；HandExam无140全答/可信时限检查，前端还先单独写end_time。现有Worker仅competency，不覆盖002。以上为源码证据，未做真实漏洞/并发调用。
- 开放requiredFields当前语义为勾选显示且必填、未勾选隐藏，而后端无条件姓名/手机号必填且不按配置校验其余字段。旧Excel总体48/36/24/12与报告58/52/45/35不一，维度>=4与>4也不同；导出答题数固定140。新版需独立结果事实及统一消费目标，不能只换Word。
- 运维边界：backup.sh仅有cron注释，未核验实际定时启用；restore-mbti/db-sync只覆盖同步和MBTI回灌，不是常态全库自动恢复。历史临时Schema恢复演练已有记录，不得误报从未恢复。SPSS/SAS专属契约未找到；团体/千人/矛盾已排除，35分钟延期，自动报告/可配置平台及扩展运维范围待确认。

## 2026-10-01 管理特质002设计再评估（待修订）

- 全面再审结论为CHANGES REQUESTED，不否定已完成的六项局部安全约束闭环。当前草案仍有四项P1运行/Schema前阻断：计分输入hash包含内容/模板全部版本导致纯重渲染生命周期不清；exam_id唯一profile与逐paper历史证据绑定冲突；传统createPaperTx把UserID写为101不能作为candidate/tester身份依据；多run/current slot及预览/PDF/Excel目标版本选择未明确。设计未在本轮改写，缺口仍待修订。
- 其他重要待补：94必需Tag中原稿无位置的年龄/用时/总体常模需位置审阅或改可选，多段建议协议/固定定义权威源需定稿；API需method/权限/资源超时矩阵；manual未到期缺答与到期incomplete需明确，新版前端不应像现examClick.doHandler先单独写人员end_time再handExam；离线自动提交Worker范围未获明确确认。generation为最新启动优先，后启动失败可能阻止先启动成功切current，应明确策略。
- 12类表和94键均是草案不是不可变目标，推荐评分事实与报告呈现版本解耦、新组卷profile与历史paper证据解耦、人员按paper+exam反向关联和证据解析；不把全套存储前置为S1依赖。S1纯身份/精确评分可以确认后先行，S2/S5不能按现稿直接启用。仅静态复审，未执行DB/运行测试/代码或模板修改/部署。

## 2026-10-01 管理特质002实施设计草案

- 新增`docs/management-traits-word-design-20261001.md`，仅设计文档，未编码/创建表接口/修模板/查DB/部署。拟定独立002 sidecar（不占用competency版本字段）、完整paper题目/选项快照、并行run/report与历史PDF捕获；版本/物理表/API均标拟新增待确认。Word合同草案推导94必需文本键+6业务图表，不是客户文件已有字段；模块等级/比较、文案修订、页数、激活交互与正式批准仍不默认实施。
- 只读代码复核确认旧002 PaperQu/PaperQuAnswer无题干/选项文字快照；标准分V号读取当前qu_repo.sort，不能从旧paper所谓“冻结题干”恢复历史身份。历史重算需独立旧题本/映射/选项证据，证据不足拒绝；新测评冻结V号与展示序分离，保存及交卷同paper锁，140完整后原子写13维/4模块/run。
- 独立设计审阅发现`CandidateHandler.LogicDeletePdfByIds`直接清candidate/tester文件与指针，绕过人员删除守卫；设计已纳入DELETE/PUT路由及前端删除报告保护。历史捕获须先立迁移标记、阻断新写并排空在途生成/上传/异步压缩，再读同句柄字节核验，不把捕获成功当原始题本证明。
- 草案补齐run/report复合paper/exam身份约束、bundle/mapping与冻结选项输入hash、严格run复用交叉校验、report revision/current DB代际条件切换及文件SHA复用门禁、模板清理后全OPC零公式/引用/externalData。初期有新版历史的删除拒绝策略为待确认建议，不套00401整链物理删除。推荐下一本地实现切片S1纯身份/评分与RED→GREEN，HTTP/DB/正式报告入口保持关闭，需用户确认设计后推进。

## 2026-10-01 管理特质002规则决策确认

- 用户通过六组选择确认同时改造00201/00202新版评分及个人Word报告，保留两题本差异；本轮不含团体报告、千人扩容、矛盾预警。完整权威记录见`docs/management-traits-word-assessment-20261001.md`§9；此前只读评估中同事项“待确认”状态由本条覆盖，原客户材料不改写。
- 精确计算、最终HALF_UP两位，等级和排序用未舍入值；学习力/创新性常模53.75/50、综合精确705/13（显示54.23）。四模块所属维度等权，总体13维等权；顶3/底3同分按模板顺序，自信→情绪→自律→社会→领导→敏感→合作→计划→责任→决断→进取→学习→创新，全同分允许两组重叠。
- Excel等级列为动态文案权威，Word错档按等级纠正；不输出无分布依据百分位。错字/频率冲突、V67文本、模块等级/比较规则、固定文本统计措辞、最终名称及具名内容/测量批准仍未定，不擅自改写或冒充批准。
- Word仅作可维护模板，结果仅交PDF；历史报告保留，新版正式报告要求140/140，历史新版重算先限staging指定样本。先修复六图再共同确定Word/服务器PDF页数，当前不承诺9/10页；25分钟沿方案，35分钟提示暂不实施。自动生成触发、版本/存储/历史访问设计、DB题本对照与样本仍待后续。本步仅记录决策，未编码/修模板/写DB/重算/部署，规则确认不等于部署授权。

## 2026-10-01 管理特质002 Word模板模式只读评估

- 客户三材料位于`docs/260929管理特质测评-优化/`，完整评估=`docs/management-traits-word-assessment-20261001.md`。XLSX51421 bytes/SHA=`b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c`；报告DOCX660351 bytes/SHA=`c82c2dc0cdcaec413866268c557561d5c53abc41ec02d94dc6aea615edca2f48`；方案DOCX36305 bytes/SHA=`e306e7918f3d9b9c6c97c08dd076679c20c58c6c55891174cd7d83c8852ae923`。本轮未修改客户原件/业务代码/运行模板/DB，未部署。
- 产品为00201/00202，不能误用00401契约。工作簿6Sheet，两题本各140题、100正40反；13公式与现有`standScore2`逐条一致，140题各归属一次、反向标记差异0；两题本去空白仅V67/V96题干不同。完整摘要/评价/建议为65/65/65，另13定义及5总体评价+5总体建议；工作簿没有原生图表或独立数据图Sheet，不含Word引用的260908图表工作簿。
- 新规则为13维百分制`25×均分−25`及13维等权综合、四类3/4/3/3归组、90/70/30/10五档、最高3/最低3及完整建议。方案要求均分先两位舍入但Word样例对应精确映射，必须确认，不可直接套00401已确认口径。方案学习/创新常模53.75/50，而Word正文55/53.75；方案综合常模705/13=54.230769…正确。Word自律正文55.50/图55、创新正文54.55/图50、总体标题59.45/缓存59.659315/图13维均值59.4477766；多条等级评价/建议错档，不能以原样例为评分金标准。线性映射与常模均值不能推导前10%等真实百分位。
- Word原稿20表/2节、0内容控件、6图表/6外链，六图均位于Word分组graphicFrame；正文3inline/13anchor。Word16只读实开9页、无修订/批注且原SHA不变；后续Word PDF导出尝试超时，无Word PDF验收证据。本机LibreOffice26.2.5.2原稿转换626842-byte A4八页，六张业务图全部缺失（5环图+13维柱线图），责任心结果和合作性建议跨页；UTF-8直接提取与画面确认中文正常，PowerShell初次乱码仅输出解码。PDF及视觉证据在`docs/generated/management-traits-assessment-20261001/`，不能当正式报告。目标服务器LibreOffice未测。
- 现有002双生成链为服务端Chromium和浏览器PdfLoader回写同一人员pdf_path，旧报告会覆盖；13维查询重算依赖可变题库顺序V映射，旧未答题置0使反向项变6，旧报告与Excel档位不同。新版需独立版本/结果报告快照、统一Word生成及导出事实源；仅复用通用LibreOffice基础设施，不放宽00401或MBTI业务门禁。既有评分相关测试16/16通过，不等于新版链通过；DB题干/选项/题序未比对，千人并发/预警/团体报告未验证。下一步等待一次性规则确认，不擅自实现。

## 2026-10-01 FB-198部署后完整复测

- 完整复测为PARTIAL PASS。Go全量和build通过；前端26文件165项及production build通过（仅既有asset/entrypoint体积warning）；当前题本转换/身份契约2/2、参与者Playwright 3/3、260915模板契约和v2重算验证器契约通过。参与者浏览器使用staging真实bundle并mock答题API，未写数据库。
- 旧“7套完整回归”任务已经过期，其引用的7个活动目录JS全部不存在并返回`MODULE_NOT_FOUND`；当前权威入口是`scripts/test/package.json`及现存专项脚本，不得继续把旧任务写作可执行完整回归。
- 新发现1项独立模板回归：`competency-v2-report-format-contract-test.py`对本地SHA=`f9859993...`和staging活动SHA=`a814c36e...`均在FB-194环图中心层级断言失败。真实PDF概览环图清晰显示`60.94`，但缺少契约要求的灰色小字`总体得分`。FB-198最高3/最低2本身已通过真实PDF验证；该模板视觉问题未在本轮修复，需独立反馈/RED→GREEN，不能将完整复测报告为全绿。
- 最终staging三服务active、内外health正常、应用关键错误0、Nginx 5xx=0、短时会话及远端/本地测试临时文件0。Production未修改。

## 2026-10-01 UF-048 / FB-198 优势与待发展改为纯分值排序

- 用户确认v2报告“胜任力综合表现”不再把良好/优秀限定为优势、其余限定为待发展；十维统一按精确分值排序，最高3项为优势、最低2项为待发展，同分继续按固定维度顺序截取。跨等级后的五个槽位复用各维度对应等级的完整已批准表现评估文案，不新增或重新批准内容包，模板仍保持3+2槽位。
- RED在十维全部低于良好时得到空优势列表；GREEN后所有等级组合均固定3项优势和2项待发展。`BuildPhase1V2ReportData`只改变选择项文案查找类型，不改DTO、评分、总体建议低分2/3维、Word字段或图表。聚焦选择器/DTO测试16/16、Go全量测试和Go build通过，编辑器诊断0；当前仅本地完成，未部署staging/production。
- [FB-198 staging部署阻塞 - 2026-10-01] Linux产物48854107 bytes/SHA=`ee4566e7a698ff592acaeab40c5986974787be27929cd1c2bc002d1adddcd300`。公网`/prod-api/health`为`ok`，但客户端公网IP仍为`20.239.176.250`，TCP/22检查失败，SSH以既有`liming`和密钥三次连接均在远端命令执行前超时。未创建部署前备份、未上传/替换后端、未重生成报告；staging和production均未修改。恢复后必须继续只读预检→完整数据库/旧后端/模板/目标旧PDF备份→部署→真实报告最高3/最低2及全文文案核验→健康/日志/临时文件终验。
- [纠正 / FB-198 staging完成 - 2026-10-01] SSH恢复后预检三服务active、内外health正常，旧后端SHA=`ac929c5868b09b2da3202f4fc11271036d2d16bbe6dca13e082c8ebf133f51cb`。部署前受限备份=`/opt/talent-assessment/backups/fb198_score_ranking_20261001_103638`，数据库gzip 12438508 bytes/SHA=`6698537b852165424a1b4db8f6b852f60d08fe85be8c0b735b7a3670b19bf365`且gzip通过，同时保存旧后端、v2模板SHA=`a814c36e3759c8ff2cf5c17148e0939b477f17d4d3f430bb6a22f530022f9f5a`和目标旧PDF SHA=`92e76fc2beb9356fc6b050422777804f64eb5443335697f004c24a86f20a5e95`，权限0700/0600。新后端48854107 bytes/SHA=`ee4566e7a698ff592acaeab40c5986974787be27929cd1c2bc002d1adddcd300`。
- [FB-198真实报告终验] paper=`24504b9c-1874-4bbd-af09-f9d0d83abb16`、run=`828092d3-0dfb-4bd4-b13f-7f564d7283e6`的十维分值为53.125/68.75/59.375/71.875/56.25/65.625/53.125/65.625/75/40.625。新规则选出优势`自律性/成就导向/计划执行`、待发展`敬业奉献/逻辑思维`；53.125同分时逻辑思维按固定顺序优先于合作意识。坐标感知提取分别核验左右两栏，五个维度名、顺序及对应等级的完整批准`dimension`文案全部逐字匹配数据库。报告实例保持completed，PDF 814637 bytes/SHA=`e981f5de4c351c24856c8803eb797a714e355cadb6bdebb4e3b1832f417cf5cb`，认证下载/数据库/文件一致，A4 10页。v2计数保持18/18/54/180/18；三服务及内外health正常，应用关键错误/Nginx 5xx/短时会话/临时文件均0；production未修改。

## 2026-09-19 v2真实报告数据与格式复核

- [FB-191新客户模板本地接入 - 未部署] 用户新附件已落为`docs/260918/competency-frontline-report-template-draft-v2.docx`，727538 bytes/SHA=`dfdabc09641492f9cf7f9bb5bda077424e2f760693fa4f69d8a12e8089c7793a`。原件ZIP/XML完整，Word 16只读打开10页/9表/19 inline/20 shapes且无写回；结构为75控件/58唯一Tag/12图表/9表/4节，但Word重存恢复了12条空目标外链、12个externalData及17个公式，且把等级刻度媒体关系从既有`rId27`重排为`rId28`，不能直接作为运行模板。清理后的58字段稿通过原客户契约；FB-191先RED复现修复器硬编码关系ID失败，再改为按OPC目标`word/media/image17.png`解析关系。最终本地运行模板652906 bytes/SHA=`eb88e00b29056f25ba13d7a8146d6287322c8e8afa13188803cdbb1059f9074e`，连续两次构建SHA一致，60字段/12图表/零外链契约通过。真实v2 value-only填充后Word 16为10页，LibreOffice 26.2为A4 10页、详情2/2/2/2/2、动态文本齐全、未解析字段0，逐页无重叠/裁切/空白/孤立标题；Go全量和build通过。仅替换本地活动模板，staging仍保持SHA=`86e302...`，production未修改。
- [纠正 / FB-191客户文件已修复 - 2026-09-20] 用户随后明确要求修改客户模板本身。修改前原件已逐字节保存为`docs/260918/competency-frontline-report-template-draft-v2.before-fb191.docx`，大小727538 bytes、SHA仍为`dfdabc09641492f9cf7f9bb5bda077424e2f760693fa4f69d8a12e8089c7793a`。原路径现替换为修复后的客户可维护模板，652906 bytes/SHA=`eb88e00b29056f25ba13d7a8146d6287322c8e8afa13188803cdbb1059f9074e`，与本地活动运行模板字节一致。直接文件契约确认60字段、12图表、零外链；Word 16只读打开10页/13物理表/19 inline/20 shapes且无写回。该动作只修改本地客户模板并保留原件备份，未部署staging或production。
- [纠正 / FB-191 staging部署完成 - 2026-09-20] 用户授权部署staging。只读预检三服务active、内外health正常、根分区可用约51.8GB；部署前受限备份位于`/opt/talent-assessment/backups/fb191_customer_template_20260920_094310`，目录root:root 0700，数据库gzip 12382658 bytes且`gzip -t`及清单全部通过，并保存旧后端、旧模板、同paper旧v1/v2 PDF，文件均0600。仅替换v2模板，后端保持SHA=`1f6cedac48830fca389dcfa249274c4233a227c13673fabe77b8a8d018f9e70f`；staging模板现为652906 bytes/SHA=`eb88e00b29056f25ba13d7a8146d6287322c8e8afa13188803cdbb1059f9074e`，包内60唯一Tag/77控件/12图表，external links/externalData/formulas均0。
- [FB-191 staging真实报告终验] 同一paper=`0aae19a9-cfb2-4fc4-b4bd-493bbc11056e`复用run=`b8741b37-21b4-4ac8-97b2-42c1b2d84788`并force重生成report=`bc9a222a-509d-4c3b-b328-c82b411b84da`；最终PDF 826984 bytes/SHA=`46166d00f7f9fa326ce4f256a79cbe282a2c800ec75691e17668e24bde32106a`，认证下载/数据库/文件SHA一致。目标LibreOffice 24.2输出A4 10页，详情2/2/2/2/2，时长、空优势、批准效度/免责声明齐全，未解析字段0；10页逐页复核无空白、重叠、裁切或孤立标题。得分事实保持总体37.812500、模块34.375/39.0625/42.708333、效度22/good，v2表数量仍15/15/45/150/15；旧v1 PDF SHA=`fad1b708...`不变，v1/v2 completed各1份，current指针正确。三服务active、内外health正常、临时会话/文件0、应用关键错误0、Nginx 5xx=0；production未修改。
- [UF-042纠正 / 组合图漏验 - 2026-09-20] 用户截图指出客户Word中的十维柱线组合图在系统PDF未显示。重新取得当前staging真实PDF第4/5页确认图确实缺失；客户原始备份直接经本机LibreOffice 26.2转换也缺失，因此不是FB-191部署回归。OOXML实证：`chart2.xml`有2系列、42个缓存值且业务键/关系完整，但图表位于`mc:Choice Requires="wpg"`内的Word 2010 `wpg:graphicFrame`；LibreOffice选择该Choice后只渲染同组等级图片，忽略graphicFrame。`mc:Fallback`有两个静态VML图片而无chart2关系，不能随value-only运行时更新。此前“10页逐页无空白/重叠/裁切”不等于所有图表可见，FB-191验收遗漏了组合图像素存在性。后续修复必须将等级图片与动态chart2拆为LibreOffice可识别的顶层对象，同时保持客户样式，并新增目标PDF组合图可见性门禁；当前仅完成根因检查，尚未修改模板或远端环境。
- [纠正 / UF-042与FB-192 staging完成 - 2026-09-20] 修复器从客户原始备份重新确定性构建，不再把该区域误替换成横向总体等级图；保留原70×447竖向等级图和原`chart2.xml`全部柱/线样式、标签、颜色、坐标轴及业务数据，仅把Word 2010组合对象拆为两个普通inline drawing。修复后的客户文件和本地/staging运行模板为633575 bytes/SHA=`0d079104547ae6d2f6b53bccfe625c8de3b3710b00a8ecb67b653878b0ca9520`；FB-191版本逐字节备份为`docs/260918/competency-frontline-report-template-draft-v2.before-fb192.docx`。模板连续两次构建SHA一致，Word 16空模板/填充稿均10页且无写回，本机LibreOffice 26.2为A4 10页并通过组合图文本可见性门禁。
- [FB-192部署与真实终验] 部署前备份=`/opt/talent-assessment/backups/fb192_comparison_chart_20260920_101837`，数据库gzip 12382729 bytes且gzip/manifest通过，同时保存旧后端、旧模板和同paper旧v1/v2 PDF，权限0700/0600。后端未变；真实paper=`0aae19a9-cfb2-4fc4-b4bd-493bbc11056e`重生成原report=`bc9a222a-509d-4c3b-b328-c82b411b84da`，PDF 830889 bytes/SHA=`eb063b560af8a8ace12527e6f83a0bbc60a26a3b0e39d0aabdb4527c18ed64b7`，认证下载/数据库/文件一致。目标LibreOffice 24.2第4页清晰显示竖向等级图、10个维度名称、10根实际分值柱/标签和常模折线；十个分值34.375/34.375/37.5/12.5/53.125/46.875/31.25/56.25/43.75/28.125及十个常模逐项与数据库一致。全报告A4 10页、详情2/2/2/2/2、逐页总览无空白/重叠/裁切；v1 PDF SHA=`fad1b708...`不变，v1/v2 completed各1份，current指针正确。production未修改。
- [UF-043 / FB-193 staging完成 - 2026-09-20] 用户对比截图指出“胜任力综合表现”生成稿丢失动态维度名且整段粗体。根因是DTO已有DimensionName但Word适配器只传RuleText；模板单个内容控件含混合run，通用替换仅保留首个粗体run。RED由`TestBugFB193_V2OverviewIncludesDynamicDimensionNames`和`TestBugFB193_V2OverviewPreservesLabelAndBodyStyles`分别复现。修复后适配值为`DimensionName：RuleText`，仅五个优势/待发展槽按中文冒号拆成粗体名称run和常规正文run；空状态及其余60字段行为不变。客户模板无需再变，仍为SHA=`0d079104...`。
- [FB-193部署与真实终验] 部署前备份=`/opt/talent-assessment/backups/fb193_overview_style_20260920_105927`，数据库gzip 12383246 bytes且gzip/manifest通过，保存旧后端、模板、基线v1/v2及目标paper旧v1，权限0700/0600。新后端48794226 bytes/SHA=`55d0a28f20d0663f9af6e7f805318f7a392d14f076e971b524393c5179c71ea6`，模板不变。选择真实paper=`af33d5c6-7c50-48c5-b468-8fae28c0cefc`、run=`674ae744-b1e4-4eee-bcf8-7890ce32449f`，其7个良好/优秀维度选出优势`敬业奉献/持续学习/合作意识`，其余3维选出最低待发展`沟通表达/逻辑思维`；新v2 report=`e8aa0480-4b21-4b33-897e-5e781079bb81`，PDF 805338 bytes/SHA=`13c21d3d680b6bdc719e601cf9cf69f0083fa5dd2e6e580a03b1831aabb4adda`。五段`维度名：批准正文`经分栏提取与数据库逐句一致，实图确认仅名称粗体、正文常规；A4 10页、组合图第4页、详情2/2/2/2/2。目标旧v1报告512060 bytes/SHA=`73923af3...`保留，v1/v2各1份且current指向v2；v2总数仍15/15/45/150/15。三服务和内外health正常、应用关键错误/Nginx 5xx/短时会话/临时文件均0；production未修改。
- [UF-044 / FB-194概览页美化 - 本地未部署] 用户红框聚焦总体摘要框、环图中心和十维图等级条。按模板职责仅改固定版式：摘要卡增加`00B050`左侧28单位强调边、`F1F8F4`浅底及120/180 dxa内边距；环图中心从拥挤的“总体评价+分值”改为9pt灰色`总体得分`和16pt绿色动态分值，独立行距避免重叠/换行；客户70×447等级图升级为180×650抗锯齿圆角绿色/橙色色阶，显示70/30/10边界，并把显示宽度由368223扩至650000 EMU、对应缩减chart宽度保持总宽和页数。修复前模板已逐字节备份为`docs/260918/competency-frontline-report-template-draft-v2.before-fb194.docx`，SHA=`0d079104...`；新客户/活动模板640430 bytes/SHA=`3b88616faad6b250fec0cbc7002cce2502a0eb5648e4db1b968322d8eb0037b7`，连续构建一致。Word空/填充稿均10页且无写回，本机LibreOffice 26.2 A4 10页，组合图可见性、60字段、12图表、详情分页、Go全量/build和实图均通过。staging仍为旧模板SHA=`0d079104...`，production未修改。
- [纠正 / FB-194 staging部署完成 - 2026-09-20] 用户授权部署。预检三服务active、内外health正常；部署前受限备份=`/opt/talent-assessment/backups/fb194_overview_beauty_20260920_120030`，数据库gzip 12387343 bytes且gzip/manifest通过，保存旧后端、旧模板和目标旧v2 PDF，权限0700/0600。仅替换模板为640430 bytes/SHA=`3b88616faad6b250fec0cbc7002cce2502a0eb5648e4db1b968322d8eb0037b7`，后端保持SHA=`55d0a28f...`。真实paper=`af33d5c6-7c50-48c5-b468-8fae28c0cefc`重生成同一v2 report，PDF 810033 bytes/SHA=`62847134bcf70468261674765fb1bd4364530dc9fe4773719db9d06f52111159`，认证下载/数据库/文件一致。目标LibreOffice 24.2实图确认浅绿左强调摘要卡、清晰居中的`总体得分 76.88分`、高清加宽等级条；总体76.875、模块77.5/75/77.083333保持，组合图和3优势/2待发展正常。A4 10页、详情2/2/2/2/2、十页接触表无空白/重叠/裁切；模板60唯一Tag/77控件/零外链/180×650等级图。三服务/内外health正常、应用关键错误/Nginx 5xx/会话/临时文件均0；production未修改。
- [UF-045 / FB-195 staging完成 - 2026-09-20] UI原卡片固定调用v1模板路径和75字段/内嵌工作簿注册表，因此有效v2附件被拒为`未支持Tag dimension.cooperation.score`。新增独立管理员v2路由`template-v2`/download/upload，严格读取正文+所有页眉的60字段、解析12业务图表并以value-only替换器拒绝公式/引用/externalData/外链；v1路由和门禁不放宽。前端卡片切换到活动v2文件名、契约文案和专用API。RED为后端缺函数编译失败、前端3/4失败；GREEN后Go全量、Windows/Linux build、前端全量/build通过。
- [FB-195部署与真实上传] 部署前备份=`/opt/talent-assessment/backups/fb195_v2_ui_upload_20260920_123109`，数据库gzip 12387699 bytes且gzip/manifest通过，同时保存旧后端、v1/v2模板和dist，权限0700/0600。生效后端48826932 bytes/SHA=`befb206ece1b81ed981cdc8774e71b23e40b42865ab19bc16d7a2dc645c4413d`，前端index SHA=`2d4ba6c5ac1e745d736f8c2894cb99406d4890123bb18fb0c0d8bb001c8e9bc9`。使用UI同源真实multipart接口上传用户附件685597 bytes/SHA=`f98599939e3bf7923abf8bd457e30259cde3fa68d6c4e9dc9d3288de0a2eae16`成功，返回valid=true、60/60字段、12图表、外链清理0和备份`competency-phase1-report-v2.docx.20260920_123206_000.bak`；元数据刷新和认证下载SHA一致。v1模板继续SHA=`54b167fc...`。同一真实paper随后重生成PDF 809120 bytes/SHA=`1b91d5912f4e55cee32eea74245274315ae45631fb4a6a038b1054cacac7009e`，A4 10页、组合图第4页、详情2/2/2/2/2及优势/待发展通过。production未修改。
- [UF-046 / FB-196 staging完成 - 2026-09-20] 用户确认五档规则：优秀`S>=90`深绿`#00A651`、良好`70<=S<90`绿色`#38B86A`、合格`30<=S<70`浅绿`#A8D889`、薄弱`10<=S<30`浅橙`#F2A45F`、不足`S<10`深橙`#E88937`，常模线保持模板橙色。实现只在`chart.dimension.comparison`分值系列重建10个`c:dPt`填充，不触碰折线和其他样式；FB-179同步只对这一区域允许动态差异。边界RED/GREEN覆盖90/89.99/70/69.99/30/29.99/10/9.99/0/100，实际DOCX十点色及折线样式门禁通过，Word/LibreOffice A4 10页。
- [FB-196部署与真实终验] SSH短时超时期间公网health持续200，恢复后先备份再部署。备份=`/opt/talent-assessment/backups/fb196_five_band_colors_20260920_143843`，数据库gzip 12388048 bytes且gzip/manifest通过，保存旧后端、UI上传后的v2模板和旧PDF，权限0700/0600。新后端48832760 bytes/SHA=`791ac73b22b155a59d347cedcd96b7bdff391ed11a2eb551afb6acc4451f2724`；模板保持用户UI上传SHA=`f9859993...`。真实paper=`af33d5c6-...`十维为合格68.75/62.5/68.75、良好75/81.25/75/87.5/87.5/71.875、优秀90.625；PDF实图同档同色，橙色常模线保持。新PDF 809137 bytes/SHA=`4225757b1a9e8ee9934a2c248979783f89674918e9707bbb302f0816c55a524d`，A4 10页、组合图第4页、详情2/2/2/2/2，v2数据仍15/15/45/150/15。三服务/内外health正常、应用关键错误/Nginx 5xx/会话/临时文件均0；production未修改。
- [UF-047 / FB-197新版导出本地完成 - 2026-09-22] 用户确认契约：现有三Sheet同步升级、仅导出completed v2、不回退v1、完整输出且暂无客户样表。根因是`competency_export.go`仍查询v1唯一结果、两组及1～5/10～50分。新增`competency_export_v2.go`：一期v1/v2产品配置均路由到v2专用导出；查询严格匹配四个v2版本、基层对象和completed状态，批量加载run/overall/module/dimension/validity、开始时间、答案及快照，无N+1；子表必须完整1/3/10/1否则失败关闭。结果汇总输出人员/时间/run/四版本/受众、总体+三模块+十维的两位百分制、中文等级、常模和比较、效度；逐题明细/题目字典通过显式v1→v2映射增加稳定ID、A/B/C编号和模块并保留原始选择/方向/最终题分。无v2结果时仅输出三Sheet表头。通用胜任力旧导出保持；两个现有下载入口同步使用同一新版工作簿。RED为新类型/构建器缺失；GREEN后聚焦导出、Go全量和Windows/Linux build通过，Linux后端48912549 bytes/SHA=`eb649326ce4ba04e2ae50b0dc6b55f43e39a9cd236411f9f745589bfede4b5a9`。未改数据库或前端，尚未部署staging/production。
- [纠正 / FB-197 staging部署与真实导出 - 2026-09-22] 最终Linux后端因headers-only查询顺序补强重建为48912525 bytes/SHA=`ac929c5868b09b2da3202f4fc11271036d2d16bbe6dca13e082c8ebf133f51cb`。部署前备份=`/opt/talent-assessment/backups/fb197_v2_export_20260922_120723`，数据库gzip 12413615 bytes/SHA=`e7e82c37bd42ccfeb2c6a26211d87c93825f7112c35b55c60c1dbafd01e9d080`且gzip/manifest通过，旧后端及v2模板一并保存，权限0700/0600。真实exam=`1786519110160592439`有3个completed v2 run；`export-raw-data`与`export-raw-answers`均返回41389-byte相同XLSX、SHA=`b4ae03e644236255c3f12eab7c6a8a5d32fde049c5834a202d674e70fc2b09af`。永久门禁`scripts/test/competency-v2-export-xlsx-test.py`验证三Sheet、汇总3行×75列、逐题270×20、字典90×14、四版本/受众/中文等级/百分制范围，并与数据库3 overall+9 module+30 dimension=42项事实逐项一致。真实无v2 run exam=`1786520226890178516`导出8491-byte工作簿，三Sheet行数均1（仅表头，75/20/14列），证明不回退v1。终验v2数据为16/16/48/160/16，三服务和内外health正常、应用关键错误/Nginx 5xx/会话/临时文件均0；production未修改。
- 复核对象为staging报告`bc9a222a-509d-4c3b-b328-c82b411b84da`、paper=`0aae19a9-cfb2-4fc4-b4bd-493bbc11056e`、run=`b8741b37-21b4-4ac8-97b2-42c1b2d84788`；未重新生成报告、未写数据库、production未修改。PDF仍为819286 bytes、SHA=`2e7cf492d5380b9842e005714739847ff5aa3e0a3fbc0b41d05df46f5bed2a50`、A4 10页。
- [数据准确性通过] 冻结答案为90/90，其中80维度题、10效度题；维度final_score总和201、效度raw总和22，80题正反向复核差异0。十维持久化公式差异0，总体由十维独立重算为37.812500，三模块独立重算为34.375000/39.062500/42.708333，均与run一致；paper/exam/participant及四版本绑定全部一致。
- [PDF数据通过] PDF展示总体37.81/合格/低于常模，模块34.38/39.06/42.71，十维两位HALF_UP分值与等级均匹配；10条表现评估、总体评价/建议、2条待发展及3条模块比较共17/17逐字匹配已批准数据库文案；无未解析Tag和旧样例值。字体均嵌入，文本替换字符0。
- [格式问题] ①测评仅配置`name,gender,telephone`，封面仍显示空年龄/单位/岗位；②冻结时长为1分钟，但封面只渲染出残缺的“时”，`1分钟`在PDF文本和画面中均缺失；③报告概览页左侧优势项为空且未显示“暂无优势项”，整页墨迹率仅1.86%；④总体等级刻度图在LibreOffice中缩成窄竖条，标签逐字换行且没有清晰分值定位；⑤详情分页从模板的每页2维漂移为2/3/2/2/1，第7物理页底部出现孤立的“自我管理类结果及建议”标题，第10页明显稀疏。
- [内容契约缺口] v2 DTO已持有`Validity.DisplayText`和`Disclaimer`，但58字段Word适配只绑定`validity.status`，没有绑定详细效度文案或批准免责声明。当前PDF仅显示“有效性：有效”，并使用模板固定的保密/使用/注意事项；若产品要求输出批准包中的两段文字，需要新增模板Tag及value-only绑定，不能由程序改写客户样式。
- [FB-190本地修复完成 / 未部署] 按模板优先原则，运行模板新增`validity.text/report.disclaimer`两项稳定Tag（总计60字段），封面表格改为3列并扩大日期/时长空间，手机号左对齐紧凑显示；等级图替换为五档横向刻度，并同步修复DrawingML与LibreOffice VML fallback；十维详情增加确定性2/2/2/2/2分页边界。运行时读取exam.required_fields，只删除未配置的完整个人信息行；零优势/零待发展分别输出明确空状态；批准效度文案和免责声明按DTO精确写入。模板构建连续两次SHA一致=`a6306762991229c9e073ab57c0e197b51317c3609b6a09f673015eda3791741a`。RED为Go缺requiredFields编译失败及模板仅58字段；GREEN后Go全量、build、模板契约通过，真实生产渲染器→本机LibreOffice 26.2生成754708-byte A4 10页PDF，时长/空优势/批准效度/批准免责声明均可提取，未解析字段0，详情页每页2维。当前未部署staging，production未修改。
- [FB-190 staging部署阻塞 - 2026-09-19] 用户已授权部署staging；Linux后端已构建为48788179 bytes/SHA=`1f6cedac48830fca389dcfa249274c4233a227c13673fabe77b8a8d018f9e70f`，模板为622617 bytes/SHA=`a6306762991229c9e073ab57c0e197b51317c3609b6a09f673015eda3791741a`。公网`/prod-api/health`仍HTTP 200且客户端公网IP仍为`20.239.176.250`，但TCP/22检查失败、SSH连续三次`ConnectTimeout=20`超时。遵守先备份后部署纪律，远端备份命令未启动，后端/模板未上传或替换，真实报告未重生成；production未修改。SSH恢复后继续：只读预检→完整数据库/旧后端/旧模板/旧v1+v2 PDF备份→部署后端和模板→force重生成paper `0aae19a9-cfb2-4fc4-b4bd-493bbc11056e`→认证下载→数据、SHA、A4 10页、2/2/2/2/2和逐页视觉终验。
- [纠正 / FB-190 staging部署完成 - 2026-09-19] SSH恢复后完成只读预检和受限备份`/opt/talent-assessment/backups/fb190_v2_report_format_20260919_170144`：数据库gzip 12737576 bytes/SHA=`b0db636ca0c193f61cc09ece433266b2fa02046107dd542db44fccbc5ec369d9`且`gzip -t`通过，旧后端、旧模板、同paper旧v1/v2 PDF均为root:root 0600。首次部署模板在目标LibreOffice 24.2仍呈现2/3/2/2/1，证明表内`pageBreakBefore`无效；未回退已正确的其他修复，而是按真实反馈将三张详情表分段并插入表间分页，模板连续构建SHA一致。最终生效后端SHA=`1f6cedac48830fca389dcfa249274c4233a227c13673fabe77b8a8d018f9e70f`，模板622799 bytes/SHA=`86e302989e3a9b976a750189c68b318157c5a4cb3b9bbc07a42a58cb76a540c8`。
- [FB-190真实报告终验] paper=`0aae19a9-cfb2-4fc4-b4bd-493bbc11056e`复用run=`b8741b37-21b4-4ac8-97b2-42c1b2d84788`并force重生成同一report=`bc9a222a-509d-4c3b-b328-c82b411b84da`；最终PDF 828445 bytes/SHA=`d659fe3ec4a9d6a3b9c8f59625da2d103d05d0e9a64209e5063594ad11099423`，认证下载SHA一致，A4 10页，时长/空优势/批准效度/批准免责声明存在，未解析字段0，详情严格2/2/2/2/2且逐页无标题孤立、重叠、裁切或空白页。v2数量保持15/15/45/150/15，得分事实仍为总体37.812500、模块34.375/39.0625/42.708333、效度22/good；同paper旧v1 PDF SHA=`fad1b708...`仍存在，v1/v2 completed各1份，current指针仍指向v2。三服务active、内外health正常、临时会话/发布文件0、应用关键错误0、Nginx 5xx=0；production未修改。


## 2026-09-18 260918客户修订材料开工前复核

- 客户修订材料位于`docs/260918/`：题本/规则工作簿56389 bytes、SHA-256=`edb9efd27ec86bc34db3a796c2022a99495fd9ec52e8a6580cd7404b2ab933b5`；开发方案966161 bytes、SHA-256=`1e44a90e821b9781b4b43762948e450ad07f9bf7420ed6d842cd6dbca9ffa991`；报告模板692662 bytes、SHA-256=`8f8c3e937dff3f2cdaf2d9fac2c935d06500dacb894364eebbb17673dbe35cb8`。三份260915基线均保留，复核未修改客户文件、运行模板、数据库或远端环境。
- [规则已修正] 方案与工作簿已统一显示名称为“数字应用”；维度和总体五档边界明确为`S>=90 / 70<=S<90 / 30<=S<70 / 10<=S<30 / S<10`，L1=优秀、L5=不足；三模块及总体常模比较改为连续、互斥区间，原人际50分重叠已消除。题本Sheet逐行与260915完全一致；等级评价50条正文也未改，只规范了L1-L5表头。
- [新增规则] 三模块报告概览按模块得分从高到低排列；优势最多3项，只取良好/优秀维度；待发展最多2项，若全部维度均为良好/优秀则不列。总体使用建议另按总体等级引用最低2项或最低3项维度，不与报告概览的2个待发展槽共用数量规则。
- [用户确认] 维度同分时按模板固定顺序截取：逻辑思维→计划执行→数字应用→成就导向→持续学习→沟通表达→合作意识→求真务实→自律性→敬业奉献；模块同分时按任务管理→人际管理→自我管理展示。报告页眉效度文字使用“有效/存疑”，内部状态继续使用`good/questionable`。
- [样例算术仍冲突] 新增“图表”Sheet仍保留求真务实/自律性`78.25`、自我管理`75.08`、总体`68.15`，违反已确认的“全程精确、最终ROUND_HALF_UP两位”；客户修订Word模板则已经使用正确的`78.13/75.00/68.13`，图表缓存为精确`78.125/68.125`。开工口径继续以精确公式和Word模板为准，工作簿“图表”Sheet必须明确为非验收样例或再修正。
- [用户确认] 正式验收采用精确公式和Word值`78.13/75.00/68.13`；Excel“图表”Sheet中的旧值只视为非验收样例缓存，不参与评分或测试期望。新版四类标识确定为`competency-frontline-phase1-v2 / competency-phase1-scoring-v2 / competency-phase1-content-v2 / competency-phase1-report-v2`，旧v1结果和报告只读保留。
- [客户模板契约] 修订模板完整保留75个内容控件/58个唯一稳定Tag、12个语义图表业务键、9表/4节/30个详情不可拆行，Microsoft Word 16只读实开为10页/9表/19 inline/20 shapes、无批注和修订。客户只改了两处动态可见示例文案及图表截图中的`78.25->78.13`，内容控件和图表业务键未丢失。
- [模板再次清理必需] 客户用Word编辑图表后重新引入12条空目标External OLE关系、12个`externalData`和图表公式/引用；现有FB-172契约按设计失败。使用已验证的字面量物化与外链清理逻辑生成测试副本后，除客户有意更新的`image18.png`外，其余75/58、12业务键、精确值、零外链/零公式、分页保护全部通过；客户原件未覆盖。
- [LibreOffice分页回归] 客户修订稿及仅清理图表后的测试副本在本机LibreOffice 26.2.5.2均输出A4 12页，12页均有文本；经验证的260915开发初稿在同一全新临时/profile环境仍为A4 10页。根因是客户重存后39个正文段落从精确280行高恢复为400、9个段落从自动252恢复为360，并有一个表格精确行高397恢复为593，导致十维详情从5页扩为7页；不是外链清理造成。正式接入前需再次做一次仅恢复跨渲染器紧凑度的布局清理，并复验Word/LibreOffice均10页。
- [程序开工差距] 当前程序仍固定旧A/B十维身份、通用能力/心理素养2组、维度/组1～5分和总体10～50分；`el_competency_result`与效度结果以paper为主键，维度/组结果以paper+快照唯一，报告按paper+内容+模板唯一且读取单一旧结果。260918必须先新增并行`result_run`及版本化总体/模块/维度/效度表，再实现新稳定身份映射、3模块百分制评分、常模比较与规则文案、新75/58模板契约和独立value-only渲染器；旧结果、旧PDF和旧渲染路径保持只读兼容。
- [程序契约差距] 当前Word模板校验和填充只处理`word/document.xml`，而260918模板有4个动态控件位于`word/header3.xml`；当前渲染器还会重写旧饼图/雷达/环图标签、坐标、网格和标题，不符合新模板“客户负责样式、程序负责值”。新版本必须校验并填充正文及页眉，运行时只更新控件文本和图表字面量数据，不应用旧版图表规范化或LibreOffice坐标校准。
- [本地基线] 当前脏工作区含大量既有未提交修改，后续必须在现状上做最小增量、不得覆盖用户改动。开工前基线已验证：Go全量测试通过、Go build通过、前端Vitest 26文件165项通过；未执行staging/production部署。
- [FB-173第一切片完成 / DB未应用] 新增MySQL 5.7+幂等迁移`scripts/sql/competency_011_result_runs.sql`和五个并行模型：`el_competency_result_run`、`el_competency_result_run_overall`、`el_competency_result_run_module`、`el_competency_result_run_dimension`、`el_competency_result_run_validity`。run按`paper_id+scoring_version`唯一，冻结四类版本、报告对象、人员快照、来源和状态；总体/效度与run一对一，模块/维度按稳定身份和显示顺序唯一，分数与常模均使用DECIMAL(18,6)。
- [FB-173兼容边界] 011没有ALTER/UPDATE/INSERT/DELETE四张旧结果表，不插入、不回填任何run，不改变001～010迁移、旧模型主键、当前提交/查询/报告路径或历史PDF。新增RESTRICT外键只连接011表与既有paper/exam；未来接入写入前必须同步扩展整链删除顺序。
- [FB-173验证] RED为五个新模型未定义导致model测试编译失败；GREEN后FB-173聚焦测试、完整model/schema包、Go全量测试和Go build均通过，相关诊断0、gofmt检查通过、011为UTF-8无BOM、scoped git diff check通过。当前本机未执行真实MySQL首次/二次迁移，故数据库幂等、表/索引/外键/collation实证仍保持未验证；未部署staging/production。
- [FB-174第二切片完成 / DB未应用] 新增纯身份定义`internal/service/competency_v2_identity.go`、产品版本作用域目录/映射模型和MySQL 5.7+幂等迁移`scripts/sql/competency_012_v2_dimension_catalog.sql`。v2固定十个语义ID与Tag键一致，A/B/C仅为版本作用域显示编号；顺序为逻辑思维、计划执行、数字应用、成就导向、持续学习、沟通表达、合作意识、求真务实、自律性、敬业奉献，模块成员固定为任务5/人际2/自我3。
- [FB-174旧答卷映射] v1旧ID按稳定来源身份显式一对一映射：a1-01→逻辑、a1-03→计划、a1-02→数字、b1-04→成就、a1-04→学习、a1-05→沟通、b1-05→合作、b1-02→求真、b1-03→自律、b1-01→敬业；空值和未知ID失败关闭，不按名称或显示顺序猜测。012只新增两张表及10+10条不可变种子，不修改旧维度、题目、发布快照或结果。
- [FB-174启用边界与验证] 四个v2标识已定义，但`ValidateExecutableCompetencyVersions`仍明确拒绝v2，第二切片没有接入新建测评、评分、历史重算或报告。RED为身份函数、两个模型和012缺失导致编译失败；GREEN后FB-174聚焦测试、Go全量测试、Go build、诊断、gofmt、无BOM和scoped diff check均通过。012尚未真实MySQL首次/重复执行，未部署staging/production；下一切片应只实现v2百分制纯评分并继续保持运行入口关闭。
- [FB-175第三切片完成 / 尚未接入] 新增纯函数评分器`internal/service/competency_v2_scoring.go`，输入为冻结v1答题维度身份和已完成正反向处理的1～5最终题分；先按FB-174显式映射为v2语义维度，再以`25 × (题分和/8) - 25`计算精确`big.Rat`维度分，总体为十维精确算术平均。样例十维结果保持总体`545/8=68.125`，未做中间、持久化或显示舍入。
- [FB-175等级与失败边界] v2独立语义等级码为`excellent/good/qualified/weak/insufficient`，连续区间固定为`>=90 / [70,90) / [30,70) / [10,30) / <10`。80题数量、题型、v1来源ID、旧显示顺序和已答题最终分均严格校验；任一维度不满8答时该维度及总体无正式分值/等级，其他独立完整维度可保留精确分值；输入行顺序不影响输出固定顺序。
- [FB-175验证与启用边界] 开始前复核用户/自动化改动后的`competency_v2_identity.go`，诊断0且FB-174身份回归通过，无需修复该文件。FB-175 RED为新评分函数、结果类型及等级常量缺失导致编译失败；GREEN后FB-175聚焦测试、FB-174/175与v1评分兼容测试、Go全量测试、Go build及相关诊断均通过。v2仍被`ValidateExecutableCompetencyVersions`拒绝，未接入运行时、result_run写入、历史重算、模块/常模或报告，未部署staging/production；下一切片应只实现三模块精确聚合与常模比较。
- [FB-176第四切片完成 / 尚未接入] 开工前直接读取260918客户XLSX“三类模块评价”和“总体评价”Sheet，确认连续区间：任务`<56 / [56,60) / [60,75) / >=75`，人际`<50 / [50,56) / [56,70) / >=70`，自我`<58 / [58,63) / [63,75) / >=75`，总体`<55 / [55,63) / [63,70) / >=70`；常模固定为任务58、人际53.75、自我60、总体57.75。比较码固定为`below_norm/at_norm/above_norm/standout`，总体最高档单独为`superior`。
- [FB-176精确聚合] 新增纯函数`internal/service/competency_v2_modules.go`。样例三模块精确分为任务`135/2=67.5`、人际`475/8=59.375`、自我`75`，因此比较结果依次为`above_norm/above_norm/standout`；总体`545/8=68.125`对应`above_norm`。输入维度顺序不影响模块固定输出顺序；报告按分数排序和同分规则留给后续选择器切片。
- [FB-176失败边界与验证] 外部修改后的`competency_scoring_test.go`先复核FB-175聚焦测试通过。模块聚合严格核对十维语义ID、稳定键、显示编号/顺序、模块归属、题数、精确分值和等级；任一子维度不完整时该模块不产生分数、等级、常模或比较码，其他模块保持独立。RED为聚合/比较函数、类型和常量缺失导致编译失败；GREEN后FB-176聚焦测试、FB-174～176与v1评分兼容测试、Go全量测试、Go build及诊断均通过。仍未接入运行时、数据库写入、历史重算或报告，未部署staging/production；下一切片应实现模块展示排序、优势最多3项、待发展最多2项及总体建议最低2/3维选择器。
- [FB-177第五切片完成 / 尚未接入] 开工前重新读取被外部修改的`competency_v2_modules.go`，诊断0；最终FB-174～177链式回归通过。另直接读取260918客户XLSX确认：优势仅取良好/优秀且最多3项，不足3项按实际数量，零项返回空；待发展取低于良好的最低项且最多2项，全部良好/优秀时返回空；总体优秀/良好/合格建议引用最低2维，薄弱/不足引用最低3维，和报告待发展槽独立。
- [FB-177确定性选择] 新增纯函数`internal/service/competency_v2_selectors.go`。模块按精确分降序，同分固定任务→人际→自我；优势按精确分降序、待发展与总体建议按精确分升序，维度同分均按模板固定顺序。样例模块顺序为自我→任务→人际，优势为数字应用→求真务实→自律性，待发展及合格总体建议为成就导向→沟通表达；薄弱/不足总体建议再加入计划执行。空优势/待发展使用非nil空切片。
- [FB-177失败边界与验证] 初始RED为选择器函数/类型缺失导致编译失败；补强测试后第二次RED证明“来自另一组结果、内部自身合法的模块行”可与当前维度混用。最终选择器会从当前维度重新聚合模块并逐字段比较，拒绝跨run混配，同时严格拒绝不完整或身份/顺序/分值/等级/常模不一致输入。FB-177聚焦测试、FB-174～177与v1兼容测试、Go全量测试、Go build及诊断均通过。仍未接入报告DTO、规则文案、运行时或数据库，未部署staging/production；下一切片应实现v2报告DTO及版本化规则文案精确匹配。
- [FB-178第六切片完成 / 尚未接入] 开工前复核外部修改后的`competency_v2_selectors.go`，诊断0且FB-176/177回归通过。新增`internal/service/competency_v2_report.go`，DTO schema固定为`competency-phase1-report-data-v2`，显式携带`result_run_id`、四类v2版本、基层员工受众、总体/模块/十维、优势、待发展、总体建议维度、效度状态和统一免责声明。评分层仍保存精确`big.Rat`，仅DTO展示使用正数ROUND_HALF_UP两位：总体68.125→68.13、人际59.375→59.38、求真/自律78.125→78.13。
- [FB-178规则文案键] 复用版本化`el_competency_report_text`承载结构但本切片不写数据库种子。v2新增内容类型`module_comparison/overall_advice/strength/development`，连同既有`overall/dimension/validity`全部按`content_version + audience + content_type + semantic identity + level/comparison code`精确匹配；优势/待发展短文案与十维完整表现评估分开，禁止跨版本、受众、状态或条件回退。仅对选择器实际选中的优势/待发展要求行，因此空分类可返回非nil空切片。
- [FB-178失败边界与验证] 文案包内任一v2活动行临时、正文/免责声明空白、精确键重复或免责声明不一致即整体失败；DTO还重新验证十维精确分与总体值，并由选择器重新聚合模块，拒绝混合run数据。RED为DTO构建器、类型和v2内容类型缺失导致编译失败；GREEN后FB-178聚焦测试、FB-174～178与v1报告兼容测试、Go全量测试、Go build及诊断均通过。未新增/导入客户规则文案，未接入DB查询、运行时端点或模板渲染，v2仍不可执行且未部署；下一切片应清理260918客户模板并实现正文+页眉value-only渲染器。
- [FB-179第七切片完成 / 尚未接入] 260918客户模板原件保持692662 bytes/SHA-256=`8f8c3e937dff3f2cdaf2d9fac2c935d06500dacb894364eebbb17673dbe35cb8`；原件契约RED明确命中`word/charts/_rels/chart1.xml.rels`外链。扩展既有确定性构建器的`--already-bound`模式，避免重复创建客户已保留的75个控件，并允许封面空段已清理；运行模板生成到`configs/export-templates/competency-phase1-report-v2.docx`，620371 bytes/SHA-256=`f02ed8a5d05f9826aa4f0cb3b7ff4265124a86ddb615ff42ab335b2e0ec3a3dd`。契约实证为75控件/58唯一Tag、12语义图表键、零外链/公式/引用，客户260918媒体逐文件字节保留；原260915构建路径也重新生成并通过契约。
- [FB-179 value-only渲染器] 新增独立`internal/handler/competency_report_v2_word.go`，不调用旧版饼图/雷达/环图规范化、LibreOffice坐标校准或内嵌工作簿重写。它要求完整58字段和12图表数据，替换正文及所有页眉内容控件首个文本节点、清空其余节点，并按`wp:docPr title`语义键只改现有`c:val`中的数字字面量；图表类型、标签、颜色、字体、坐标、轴线和所有非值XML保持不变，其他ZIP部件字节不变。缺失/未知字段或图表、重复/未知图表键、公式/引用/externalData以及任意`.rels`外链均失败关闭。
- [FB-179验证] 初始RED为渲染函数与语义图表解析函数缺失；补强RED又复现图表关系外链被接受，最终增加全OPC关系扫描。真实渲染DOCX为620648 bytes/SHA-256=`8eefda32eb39e39451870fbe5aceb1d0337a73522c7e80bbf0041a7c6f6b73d4`，Microsoft Word 16只读打开为10页/9表/19 inline/20 shapes；LibreOffice转换PDF为745309 bytes/SHA-256=`0f230cbda7b1e936c89b0a8386360e26b38683aec844e21b8605e54c166740f8`、10页A4。FB-179、完整handler、Go全量和build通过，诊断0。当前仅清空未用预定义槽的值，不改变/发明布局；DTO到58字段/12图表的适配、运行时转换、数据库写入和部署仍未接入，v2保持不可执行。下一切片应绑定result_run与报告实例/当前展示指针，再统一接入DTO→渲染运行链。
- [FB-180第八切片完成 / DB未应用] 新增MySQL 5.7+幂等迁移`scripts/sql/competency_013_report_result_run_binding.sql`。既有`el_competency_report`仅新增可空`result_run_id`，旧报告保持NULL且脚本无UPDATE/INSERT/DELETE，不回填、不覆盖旧实例或PDF；保留原`paper_id+content_version+template_version`唯一键，并新增`result_run_id+content_version+template_version+audience`唯一键。`CompetencyReport.ResultRunID`使用`*string`及`omitempty`，旧API序列化不新增空字段。
- [FB-180当前展示与一致性] 新表`el_competency_report_current`以`paper_id+audience`为主键、`report_id`唯一，只保存当前展示选择，不删除历史版本。报告到run使用`(result_run_id,paper_id)→(id,paper_id)`复合RESTRICT外键，当前指针使用`(report_id,paper_id,audience)→(id,paper_id,audience)`复合RESTRICT外键，从数据库层阻止跨答卷run和跨受众报告误绑定；为复合外键新增候选唯一键。DDL通过information_schema守卫，重跑时已有外键则跳过约束列字符集对齐。
- [FB-180删除链与验证] 整链删除同步扩展为报告审计→当前指针→报告实例→run效度/模块/维度/总体→result_run→旧结果→答卷，避免013/011外键阻断。RED为`CompetencyReportCurrent`缺失导致model编译失败及删除链缺6步；GREEN后FB-180/FB-048聚焦测试、完整model+handler、Go全量测试和Go build通过，诊断0。013尚未真实MySQL首次/重复执行，当前生成/下载API和前端仍按旧paper版本查询，不写`result_run_id`或当前指针；未部署staging/production。下一切片应完成DTO→58字段/12图表适配并接入v2运行时读写，之后才评估解除v2执行门禁。
- [FB-181第九切片完成 / DB未应用] 新增`CompetencyRuntimeService.FindPhase1V2FormalReportData`：仅在发现同paper且评分版本精确为v2的completed run后读取总体/3模块/10维/效度，按`score_sum`重建精确有理数并交叉校验持久化六位分数、等级、模块、维度常模、总体常模、比较码、效度边界、提交元数据、人员身份及paper/exam一致性；发现非法v2 run后失败关闭，不回退v1。v2批准包继续要求双批准、双SHA、免责声明，并新增当前`APP_ENV`精确匹配。
- [FB-181适配与转换] 新增DTO→Word适配器，固定输出58字段和12个语义图表：显示字段用两位值，图表使用从持久化事实重建的精确分数和经校验的十维常模；总体评价与低分维度替换后的使用建议共同进入模板现有规则控件，优势/待发展不足槽清空。`phase1WordReportRenderer`新增独立v2模板路径，v2只执行FB-179 value-only渲染和既有文档转换器，不调用v1图表校准，也不允许回退到读取旧paper结果的Chromium页面。
- [FB-181实例与兼容] 生成端优先探测精确v2 run；完成时报告实例绑定`result_run_id`，报告元数据、paper+audience当前指针和成功审计同事务提交，不更新candidate/tester旧`pdf_path`。强制重生成在新文件成功前保留原completed实例，任一步失败仍可下载旧PDF。单份下载有当前指针时校验报告→run→版本/受众→批准环境，无指针时保持原v1路径。为兼容011～013尚未应用的数据库，运行时先探测表/列，缺失时不查询新表并在所有v1报告SELECT/INSERT中省略`result_run_id`；批量下载仍按旧v1版本键，留待独立切片。
- [FB-181验证与边界] RED为运行时重建器、正式数据类型和58/12适配器缺失导致service/handler编译失败；GREEN后聚焦FB-181、完整service/handler/config、Go全量和Go build通过，相关诊断0，`git diff --check`与`gofmt -d`通过。真实v2模板已经适配器填充并传入捕获转换器，输出DOCX含实际总体评价/建议。011/012/013仍未执行，未创建/重算任何v2 run，未导入或批准v2文案，真实MySQL事务和LibreOffice PDF尚未执行，执行门禁仍关闭；staging/production均未部署。下一切片应实现历史/新答卷v2 result_run原子写入及v2文案导入批准，之后执行数据库迁移与staging真实报告验收；批量下载当前指针切换另行处理。
- [FB-182兼容性纠正] 最终审查发现FB-180扩展的整链删除在011/013未应用时会无条件访问不存在的current/run表，导致既有v1胜任力测评删除事务失败。RED精确报告6张可选表均无存在性守卫；修复后current表独立检查，run子表同时要求parent存在并逐表检查，随后仍按效度→模块→维度→总体→run顺序删除，缺表时继续完整v1链。FB-182/FB-048、Go全量和build通过；未在真实pre-/partial-migration MySQL执行删除。
- [FB-183身份与并发纠正] 最终审查继续发现两项运行边界：run虽校验paper/exam但未校验paper.user_id，且单份下载未与force重生成共享锁。RED分别为身份校验符号缺失导致service编译失败、Download源码无锁。修复后同一次paper查询读取`exam_id/user_id/user_time`，要求run的exam/participant同时匹配；Download从选择当前实例到审计和流式输出全程持有既有paper分片锁，重生成只能在下载完成后删除旧文件。聚焦FB-181～183通过；批量ZIP仍为独立旧链，不在本次并发保证范围。
- [FB-184第十切片完成 / DB未应用] 新增单卷v2结果运行构建和历史重算服务：只接受已完成且四类版本精确为v1的一期结果，从`el_paper_qu`联结发布时冻结题目/维度快照读取90题；维度使用持久化`final_score`，同时以`raw_answer+scoring_direction`复核其一致性，效度使用10题原始值。计算继续复用精确`big.Rat`评分、三模块和常模，持久化为六位decimal；样例总体68.125、人际59.375、求真务实78.125均通过纯构建实测。
- [FB-184原子性与幂等] 新建run、overall、3 module、10 dimension和validity在同一事务写入；提交/重算均先锁paper，按`paper_id+v2 scoring_version`查询。已有run只在状态、版本、受众、人员快照、提交元数据、总体/模块/维度/效度精确值全部重新校验通过后返回`reused=true`，不执行Save/Update/Delete；损坏或未完成run失败关闭且不覆盖。011五表全不存在时新v1提交保持兼容；出现部分表、关键列、唯一索引或RESTRICT外键不完整时事务失败关闭。
- [FB-184入口、验证与边界] 新增后台单卷`POST /exam/api/competency/results/recompute-v2`，仅管理员ID=1或全局权限可调用；普通`exam:list`用户实测403。新完成且结果完整的一期v1提交在011完整存在时同事务创建v2 run；已提交历史答卷由单卷入口重算。RED为构建器和handler符号缺失；GREEN后FB-184聚焦4项、Go全量测试、Go build、相关诊断和scoped diff check通过。011～013仍未真实MySQL执行，故真实事务回滚、并发重复请求、数据写入和staging运行尚未验证；未批量重算、未生成v2报告、未导入批准文案、未解除v2新建测评门禁，未部署staging/production。
- [FB-184复审阻断 - 2026-09-19] 只读复审结论记录于`docs/competency-v2-result-run-review-20260919.md`：当前方案的同事务写入、paper锁、精确评分、raw/final复核、不可变复用和管理员权限方向正确，但尚不可执行011～013或staging重算。P0包括：构建器未拒绝非基层受众，冻结题目/维度JOIN未验证同一exam，每次提交在paper锁事务内执行5表+6列+3索引+6约束约20次information_schema探测，以及事务回滚/并发幂等/损坏run/部分结构仍只有源码字符串断言而无可执行证据。
- [FB-185待办与额外边界] 需先补RED：非基层受众、跨exam快照、90题输入异常、已有run零写入与子表损坏、module_id/source漂移、子表插入失败ROLLBACK、结构签名矩阵及双连接并发。P1另有：011门禁未核对全部列/顺序唯一键/FK签名，run未冻结user_time而报告读取可变paper值，HTTP可能返回原始DB错误，012 no-op种子重跑不能识别目录/映射漂移。2026-09-19重新执行Go全量和build均通过，但只证明已有测试和编译，不解除上述阻断。
- [FB-185A身份链修复完成 - 2026-09-19] RED新增非基层受众构建和跨exam冻结快照两项测试，初始因loader缺少exam参数而编译失败，且旧构建器会接受leader受众。GREEN后`buildPhase1V2ResultRunRecords`只接受`frontline_employee`；`loadPhase1V2AnswerInputs`由已锁paper传入exam_id，并在question/dimension两个冻结快照JOIN中分别约束同一exam。聚焦2/2、result-run/runtime 3/3、Go全量、build、诊断及独立代码复审通过。未执行迁移、数据库写入或部署；FB-185其余结构门禁、事务行为测试、并发、module_id/source、时长冻结和错误映射仍为阻断。
- [FB-185B语义身份修复完成 - 2026-09-19] RED证明source校验器及创建/复用/正式读取共享入口缺失；模块fixture补齐合法module_id后，篡改module_id必须被拒绝。GREEN要求持久化模块`module_id`与固定module_code一致，source仅允许`submission/historical_recompute`，并接入run创建、已有run幂等复用及正式报告读取。特意验证historical_recompute可复用submission创建的合法run且不改写原source。聚焦、FB-181/183/184/185A/B链、Go全量、build、诊断和两轮独立复审通过。事务回滚、Schema签名/缓存、真实并发、user_time冻结与错误映射仍待后续切片。
- [FB-185C～H本地加固完成 - 2026-09-19] C用真实writer+sqlmock证明module插入失败时BEGIN后ROLLBACK且无COMMIT，合法已有run只读校验并零写入复用；D将每次交卷事务内约20次Migrator探测替换为事务前、每服务实例一次的information_schema完整签名缓存，校验五表全字段定义、必要唯一键、六外键和013可选复合索引；迁移后必须重启服务刷新缓存。E为提交、超时自动提交、管理员重算和v2报告生成统一稳定错误响应并在服务端记录内部原因。F在尚未应用的011 overall新增user_time，提交/历史重算冻结、复用校验、报告读取均使用该值。G为012增加精确10+10 tuple及作用域总数漂移门禁，使用MySQL 5.7兼容的条件化不存在表查询而非不可PREPARE的SIGNAL。H使011在013复合报告FK存在时跳过受约束paper_id重复ALTER，保持重跑安全。关联测试、Go全量、build和最终独立复审通过。
- [FB-185I环境门禁 - 2026-09-19] 已新增`TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun`：配置`FB185_MYSQL_DSN`后会创建独立临时库、应用011、以两个连接竞争同一paper锁，并断言一次创建/一次复用及run/overall/module/dimension/validity=`1/1/3/10/1`后清库。本机未配置该专用测试DSN，测试明确SKIP；禁止把编译通过描述成真实并发通过。011～013真实首次/重复执行及staging重算仍未进行，production未修改。
[完成补充 - 2026-09-19] 上述环境门禁已在staging MySQL 8.0.46关闭：双连接真实并发得到一次创建、一次复用和1/1/3/10/1；011～013首次及重复执行均成功。15份完整v1答卷首次批次创建15份v2 run，第二批全部复用，最终run/overall/module/dimension/validity=`15/15/45/150/15`。production未修改。
[FB-186/187 staging实证 - 2026-09-19] 首次历史重算发现Schema预检真实兼容缺陷：MySQL拒绝`AS precision/scale`，修复后又发现information_schema投影到`Table/Name/...`未显式映射，导致现存列和PRIMARY索引被误报缺失。两项均先补RED，再用安全别名及全部GORM column tag修复；重启清缓存后15份重算通过。此类information_schema Raw扫描不得依赖隐式字段名推导。
[v2文案草稿 - 2026-09-19] `generate-competency-v2-content.py`从260918客户工作簿确定性生成124条精确键：50维度完整评价、20优势、30待发展、5总体、5总体建议、12模块比较、2效度。效度与免责声明来自staging已批准v1包并在JSON中标明来源；内容SHA=`329409e408f10ec7f048a757e83c963b7736f66e41e8d601ff4397fd8545b48c`，工作簿SHA=`edb9efd27ec86bc34db3a796c2022a99495fd9ec52e8a6580cd7404b2ab933b5`。staging只导入124条`status=1,is_temporary=0`和draft包；因缺少本v2精确哈希的具名内容/心理测量双批准，不得激活或生成正式报告。
[v2批准与真实报告 - 2026-09-19] 用户明确批准并提供内容负责人LIming、心理测量负责人Ruiling。015仅对上述精确双SHA执行staging批准，124条文案全部`status=0,is_temporary=0`。staging因`APP_ENV=production`复用生产形配置触发FB-188；现以`REPORT_EFFECTIVE_ENV=staging`独立约束报告批准环境，未改变配置文件选择，production未修改。真实v2报告绑定独立run并通过认证下载，PDF 819286 bytes/SHA=`2e7cf492d5380b9842e005714739847ff5aa3e0a3fbc0b41d05df46f5bed2a50`，A4恰10页；同paper旧v1报告/PDF SHA=`fad1b708dff42389eb1da4d94a34ea52b0e5393a09cb223a0fba6f5ec59a2b99`仍存在。
[v2完成后复核 / FB-189 - 2026-09-19] 本地Go全量测试和build再次通过，编辑器诊断0。staging三服务active、health正常，环境仍为`APP_ENV=production + REPORT_EFFECTIVE_ENV=staging`；8张目标表/8个RESTRICT外键、10+10目录映射、15/15/45/150/15运行数据、124条活动文案和精确双SHA批准均保持不变，应用账号`positive_app`对`element`的SELECT/INSERT/UPDATE/DELETE/CREATE/ALTER权限仍存在。当前v2 PDF仍为819286 bytes、SHA=`2e7cf492...`、A4 10页，current指针绑定run `b8741b37-...`，同paper v1/v2 completed报告各1份且文件SHA均匹配；应用关键错误和Nginx 5xx均为0。复核发现工具脚本仍假定报告未批准，FB-189先RED后移除该过期默认断言；重跑得到15份全部复用、零创建、数量不变。production未修改。

## 2026-09-18 260915客户新版基层报告需求只读评估

- 客户材料位于`docs/260915/`，共3份：题本/等级/总体评价XLSX、开发方案DOCX、基层员工表格版报告样例DOCX。本轮仅只读评估，未修改业务代码、数据库、运行模板或任何远端环境。
- 新题本仍为90题：80道维度题+10道效度题；与当前`phase1-questions.csv`按去空白题干逐题比对为90/90完全命中，题型、计分方向和考察点差异均为0。变化不是题目内容，而是10个维度的身份、顺序和分组：新口径为任务管理5项、人际管理2项、自我管理3项，编码依次使用A1-01～A1-05、B1-01～B1-02、C1-01～C1-03；当前系统仍为通用能力5项+心理素养5项及A/B编码。
- 新计分改为百分制：维度分=`25×八题平均分−25`，三类模块为所属维度百分制平均，总体为10维百分制平均；维度和总体均采用`≤10 / (10,30] / (30,70] / (70,90] / >90`五档。效度仍为10题原始分合计、`≤35`良好、`>35`存疑。当前系统持久化和报告主口径仍为维度/一级1～5、总体10～50，属于评分版本变更，不能只换模板。
- 新工作簿的50条维度诊断与当前CSV按维度名称+L1～L5逐条去空白比对50/50相同，5条总体诊断5/5相同；主要内容变化是等级标签、分值尺度、三类模块常模和报告展示。模块常模为任务58、人际53.75、自我60，总体常模57.75。
- 新报告样例经Microsoft Word 16只读打开为10页、9表、19 inline、20 shapes、4节；但本机LibreOffice 26.2.5.2转换为A4 13页，封面个人信息被拆至第2页、阅读指南跨到第4页、详情延长至第13页。该稿不能未经兼容改造直接作为服务器运行模板。
- 新报告包含12图表：总体环图1、10维得分+常模组合图1、10个维度环图；当前模板为一级3D饼图1、十维雷达图1、10个维度环图。新样例还有12条指向客户本机D盘Excel的外链，上传前必须继续执行零外链清理。
- 新样例内容控件为38处/34个唯一Tag；相对当前运行模板缺15个既有业务Tag，并错误重复`a1-02.diagnosis`和`a1-04.score/level/diagnosis`。其中沟通表达块误复用了持续学习Tag，`a1-05`三项Tag完全缺失。当前严格上传契约会拒绝该文件，不能直接部署。
- 新增报告能力包括：三类模块得分/常模比较、总体常模比较、按条件选择最多3项优势和2项待发展、百分制图表及表格式十维详情。现有两一级维度快照/结果表可以容纳3行，但版本、固定分组、评分、报告DTO、文案匹配、模板字段与图表物理映射均硬编码为当前两组口径，必须按新产品/评分/内容/模板版本并行实现，不能覆盖历史一期版本。
- 客户材料仍有待确认冲突：①开发方案称L1=`>90`、L5=`≤10`，工作簿标题行却把L1写成不足0-10、L5写成优秀90-100；②维度50分归入合格，同时人际常模比较区间既写50-55持平又写50及以下低于，50重叠；③样例78.25说明可能先将4.125四舍五入为4.13再映射百分制，而公式若用精确均值应为78.125/78.13；④优势/待发展遇并列、少于目标数量及效度存疑时的展示规则未完整定义；⑤“有效性：有效”的显示文案与现有good/questionable映射未明确。
- [用户决策 - 2026-09-18] 产品入口采用覆盖当前版本，不再让新建测评继续选择旧一期；但数据审计采用新版本并存方式：新建稳定内部维度ID，旧结果/PDF只读保留，历史答卷基于原始答案生成独立新版结果和新版PDF，不原地覆盖旧记录。历史重算范围仅staging，production不在本轮范围。
- [用户决策 - 2026-09-18] 新评分全程精确计算，仅最终展示保留两位；报告验收要求Microsoft Word与服务器LibreOffice均固定10页；本轮只做基层员工版，260915材料继续视为待内容负责人和心理测量负责人双重批准。
- [实施边界 - 2026-09-18] 等级L1/L5方向、边界50分归属及优势/待发展规则仍待客户书面确认。确认前只实施新稳定身份、版本和导入/迁移基础，不实现或启用新版正式评分、历史重算和报告生成，避免临时规则形成正式结果。
- [二轮实施审查 - 2026-09-18] 原“先身份/版本、再评分报告”的方向正确，但步骤需进一步前置版本化结果存储。当前总体结果以`paper_id`为主键，二级/一级结果以`paper_id+快照ID`唯一，效度结果也以`paper_id`为主键；同一答卷无法同时保存旧版和新版结果。历史重算前必须先新增独立`result_run`及其版本化总体/模块/维度/效度明细，或完成等价的复合主键改造。为降低旧链回归和外键风险，首选新增并行表，不原地改变001～010旧表主键。
- [二轮实施审查 - 2026-09-18] 新稳定维度身份不能直接写入现有主表覆盖旧A/B身份：现表的code/name/display_order均全局唯一，现有题目、发布快照、结果、导出和模板都引用旧身份。应先建立产品版本作用域的维度目录及`旧维度→新稳定维度`映射，再让新测评发布快照和历史重算读取该目录；旧快照保持只读。
- [二轮实施审查 - 2026-09-18] 报告实例虽然按`paper_id+content_version+template_version`唯一，但未绑定评分运行批次；下载和生成仍先按`paper_id`读取唯一旧结果，且生成成功会覆盖candidate/tester的`pdf_path`。新版并存需让报告绑定`result_run_id`并增加当前展示指针，生成新PDF只写新路径，不删除旧PDF。
- [二轮实施审查 - 2026-09-18] 当前还存在一个独立旧版缺陷：评分总体最低档内部码为`not_qualified`，Word标签映射却使用`unqualified`，最低档可能显示空文本。该问题应单独按RED→GREEN修复，不与260915版本改造混做。
- [模板职责确认 - 2026-09-18] 260915报告采用“客户负责样式、程序负责值”的长期边界。开发方先一次性修复客户样例的内容控件、遗漏/错误Tag、12条图表外链、图表业务键和易漂移分页结构；之后客户可维护字体、字号、颜色、边框、间距、固定文字及契约允许范围内的图表样式，程序生成时只替换内容控件值及图表数据，不重建数据标签、颜色、字体、坐标、线型、图例、坐标轴或图表类型。
- [模板安全边界 - 2026-09-18] 无需改代码的模板变化限于：动态控件之外的固定文字，以及不改变业务结构的字体/颜色/边框/间距/图表样式。必须由开发介入的结构变化包括：新增或删除维度/模块、删除/复制/改名动态控件、改变动态字段含义、改变图表类型、数据系列/数据点数量或顺序、图表业务键、动态模块顺序及跨页结构。所有客户重存稿仍需经过Tag唯一性、零外链、图表数据槽、Word+LibreOffice 10页和逐页视觉门禁。
- [现状差距 - 2026-09-18] 当前旧一期渲染器不完全满足上述边界：除更新图表缓存外，还会规范化一级饼图标签、雷达图坐标轴/标签、环形图粗体/对齐并重建居中标题，以及应用LibreOffice坐标校准。260915新版本必须采用独立的value-only渲染路径；旧版本为保持历史兼容继续保留原行为，不在本次顺手重构。
- [条件区块确认 - 2026-09-18] 用户允许程序按业务条件隐藏模板中预先定义的完整行、段落或重复槽位，用于未配置个人字段、效度提示和数量不足的优势/待发展项；该能力只控制显隐，不得修改区块的字体、颜色、边框、间距或图表样式，也不得在运行时发明新布局。
- [报告数据分类确认 - 2026-09-18] 260915报告内数据正式分为三类：A=模板固定内容，每份报告完全相同，包括标题、定义、免责声明、栏目名称，由客户在模板中直接维护，程序不替换；B=评估系统数据，直接读取冻结结果或按已批准评分规则计算，包括姓名、分数、等级、排名和图表数据，由系统字段/图表数据绑定写入；C=规则文案，按已批准条件和内容版本精确匹配，包括表现评估、使用建议、优势项和待发展项，由规则文案绑定写入，缺少精确匹配时报告生成失败，不跨版本或条件回退。A类不设置动态绑定；B/C类绑定不得由客户删除、复制、改名或改变含义。
- [B类数据审查 - 2026-09-18] 现有90题题干、方向和考察点无需修改，原始作答、正反向计分、人员冻结信息、提交时间、作答时长和效度10题原始分均可复用；但题目当前绑定旧A/B维度身份，260915必须通过产品版本作用域映射到新稳定内部维度ID和A/B/C显示编号，不得覆盖旧快照。
- [B类样例算术冲突 - 2026-09-18] 260915样例中的求真务实和自律性均写78.25，该值无法由8道1～5整数题按`25×精确均分−25`得到；对应可达值是78.125，最终两位为78.13。按样例其余分数重算，自我管理精确为75，总体精确为68.125、最终两位为68.13，不是样例的75.08和68.15。样例明显使用了“先把4.125写成4.13再转百分制”的中间舍入，违反用户已确认的全程精确规则，不能作为评分验收值。
- [B类常模比较冲突 - 2026-09-18] 十维常模可确定性聚合为任务58、人际53.75、自我60、总体57.75；但客户比较区间按整数书写，不能覆盖实际可出现的小数（如任务60.5、人际69.5、自我62.5、总体69.5），且人际50同时落入“50-55持平”和“50及以下低于”。样例的人际59.38按材料应为略高却写接近/持平，总体68.15按材料应为略高却写优于；连续区间及显示词必须书面定稿后编码。
- [B类绑定冲突 - 2026-09-18] 新报告样例38个内容控件只覆盖人员、日期和部分维度明细；总体得分/等级、3模块得分/等级/常模比较、效度状态、优势3项、待发展2项及作答时长没有完整B类绑定。沟通表达块误复用`dimension.competency-a1-04.*`，缺少应有的沟通表达绑定；概览优势还复用了详细诊断Tag，短摘要与完整表现评估会互相覆盖。12个图表只有样例缓存和本机Excel外链，没有稳定业务键，当前程序无法安全更新新图表数据。
- [B类程序修正范围 - 2026-09-18] 当前程序固定旧10维身份、2组、维度1～5、总体10～50和旧图表，报告DTO无常模/比较码/优势待发展选择/实际名次，结果和效度以paper单版本保存，报告不绑定评分run且强制重生成会替换人员pdf_path并删除旧实例文件。260915必须新增独立版本分支、3模块百分制精确评分、版本化常模与选择结果、并行result_run结果表、报告绑定result_run及新value-only模板契约；旧路径保持只读兼容。
- [B类排名边界 - 2026-09-18] 260915样例实际呈现的是“与常模比较”，没有个人名次或百分位值；当前程序也只有按分数排序，没有持久化名次。若“排名”要进入报告，必须另行确认参照人群、统计时间点、并列规则、样本门槛和是否冻结，否则本轮不实现实际排名字段。
- [B类优势待发展冲突 - 2026-09-18] 等级评价Sheet首行与样例说明支持“优势最多3项且只取良好/优秀、待发展最多2项且全为良好/优秀时不列”，但总体评价Sheet在薄弱胜任/尚未胜任的使用建议中要求最低3个维度，和报告仅2个待发展槽位冲突；同分并列、边界截断和稳定排序仍无规则。开发前必须确认“报告综合表现固定2项”与“总体使用建议低分3项”是否为两个不同输出，不能让同一选择结果兼任。
- [B类名称冲突 - 2026-09-18] 题本、等级评价、常模Sheet和报告样例使用“数字应用”，开发方案常模段出现“数据应用”。程序稳定内部ID应与显示名称解耦；正式显示词仍需客户确认统一为“数字应用”或“数据应用”。
- [FB-172 260915模板初稿完成 - 2026-09-18] 新初稿位于`docs/260915/competency-frontline-report-template-draft-v2.docx`，619782 bytes、SHA-256=`23d605ea9eed37eaefe78f00d2b23826f7346cee475b4750990c077d0526e029`；客户原报告样例保持692319 bytes、SHA-256=`407e5302140e3ecb31052cf82e7445de57fa958c0e259ce976efb9ac7d4dc44a`。确定性构建器与契约测试分别位于`scripts/tools/build-260915-competency-report-template.py`和`scripts/test/competency-260915-template-contract-test.py`，连续构建SHA一致。
- [FB-172模板契约 - 2026-09-18] 初稿共有75个内容控件/58个唯一稳定Tag：覆盖封面与详情页眉人员字段、提交日期/时长/效度状态、总体得分/等级/常模比较、三模块得分/等级/常模比较、模块摘要、总体使用建议、3个优势槽、2个待发展槽和10维得分/等级/表现评估；修正了沟通表达误用持续学习Tag。新稳定键使用语义身份（如`dimension.communication.*`），不把A/B/C显示编号当内部ID。
- [FB-172图表与样例值 - 2026-09-18] 12图表保留客户图表类型、系列、颜色和样式，增加12个稳定业务键；移除12条本机Excel关系、12个externalData及全部图表工作簿公式，保留并更新内置数值/分类缓存。按全程精确、最终两位规则将求真务实/自律性改为78.13、自我管理改为75.00、总体改为68.13；图表缓存保留精确78.125/68.125。
- [FB-172版式验证 - 2026-09-18] 仅删除封面标题与个人信息区之间4个确认无文字/绘图的空段，段落/行高缩放为原值70%，并给十维详情30行加cantSplit、标题/得分行加keepNext；客户媒体文件名和字节逐项不变。Microsoft Word 16实开为10页、9表、20 shapes、19 inline且无修复；LibreOffice 26.2.5.2输出819881-byte A4 PDF，10页均非空、十维齐全、旧错误示例值为0，PDF SHA-256=`b10228ae44c838604a86079876df4b51b01e3d8eda9426eff9c768bbfe66547f`。该文件仅为未接入程序的本地初稿，未替换运行模板、未部署staging/production。

## 2026-09-16 单机300人同时测评容量评估

- 评估场景经用户确认为300名独立评测者同时登录，00401一期90题/20分钟，答题高峰仅答题与交卷、不生成PDF。完整评估见`docs/single-node-300-concurrent-assessment-20260916.md`。
- 当前链路为Nginx→单Go进程→同机MySQL/Redis；生产配置MySQL连接池maxOpen=50/maxIdle=10，Chrome PDF池2路，LibreOffice全局单槽。每题即时保存且无服务端轮询；300人共产生27000个保存事务，20分钟平均22.5次/秒，主要风险是300人集中开始、同步翻题和集中交卷造成的连接池与磁盘IO突发。
- 已有最接近实证仅为staging 100人、10并发、每份384题组卷与详情：创建总耗时5.190秒，p95=0.758秒，写入38400条试卷题；未覆盖300人登录、27000次逐题保存、集中交卷和Worker兜底，因此不能宣称300并发已验证。
- 正式单机推荐基线为8 vCPU、16GiB、200GiB NVMe/高性能SSD、200Mbps服务端公网带宽；最低4 vCPU/8GiB/100Mbps仅可作为压测起点。当前前端核心入口+答题页资源实测未压缩2448398 bytes、已有预压缩副本641919 bytes/客户端，但Nginx是否实际启用gzip/gzip_static需部署时核验。
- 云采购口径补充：选择非突发/非共享核的8核16GiB x86_64实例、60～80GiB系统SSD、200GiB独立高性能数据盘（至少3000 IOPS/125MB/s）、固定公网IP和200Mbps有保障的BGP公网带宽；若静态资源正式接入并预热CDN，源站带宽可降至100Mbps。仅开放HTTPS及限源SSH，8092/3306/6379不开放公网；每日快照外还必须有异机MySQL备份和恢复演练。
- 上线门禁为300个唯一身份的真实节奏压测：60秒登录、30秒开卷、20分钟27000次保存、10～30秒集中交卷，并单独测试300人离线后Worker兜底。答题窗口必须错峰PDF、批量导出、备份和升级。未执行代码修改、远端部署或真实300人压测。

## 2026-09-03 UF-041 / FB-171 个人信息紧凑双列排布

- 用户以两张报告截图要求个人信息字段依次、紧凑、保持双列、上部对齐且无空行，并明确选择“姓名独占首行”：姓名整行；其余已配置身份字段按客户模板中的出现顺序每行两列；奇数末项留在左列；时间/时长固定在最后一行双列。
- [纠正 - 2026-09-03] FB-167此前确认的“每个身份字段独占整行、模板决定位置”已被本次明确规则部分覆盖。模板仍决定身份字段顺序和单元格样式，但运行时现在决定紧凑双列配对；姓名全宽和时间/时长同排为固定产品规则。
- FB-171 RED由`TestBugFB171_Phase1ProfileUsesCompactDoubleColumns`复现：`name,gender,telephone`实际4行，目标3行。GREEN后为姓名全宽、性别/手机号双列、时间/时长双列；`name,age,gender,telephone`为姓名、年龄/性别、手机号左列、时间/时长，保持模板顺序和奇数末项规则。
- 实现仅修改`filterPhase1WordProfileTable`的文档内表格重排：先按模板顺序收集已配置内容控件，再重建身份行；不改requiredFields过滤、字段值、报告计分、数据库、API、模板文件或图表。FB-130/166/167/168/171相关回归5/5、编辑器测试622/622、Go全量和build均通过，相关诊断0。
- 使用当前staging生效FB-170模板SHA=`54b167fc...`经新运行时填充并由LibreOffice 26.2.5.2生成A4 11页PDF；`name,telephone`实图为姓名整行、手机号下一行左列、时间/时长最后一行双列，字段连续、上部对齐且无空行；图表样式保留测试通过。
- [staging部署完成] 部署前完整备份=`/opt/talent-assessment/backups/fb171_profile_double_columns_20260903_091619`，目录root:root 0700、文件0600；数据库gzip 12669320 bytes/SHA-256=`3229c686b12640d2790f8b66a5a5974e301d7c889b1fcd6fd52f7c2430dc8b50`且gzip/manifest全部通过，同时保存旧后端、当前模板、前端index和凯迪旧报告。新后端SHA-256=`1b0938cea960cbde12fe676bd3805af0145dadcd4acf7744ea81eac2a6f2d921`；模板保持`54b167fc...`、前端保持`cd395b2f...`，数据库结构未变。
- 真实paper=`287347a7-d8e1-4606-9ba0-f352f2e848af`的requiredFields=`name,age,telephone,gender`，强制重生成原实例`d785dc6d-4074-45d2-b71c-9f93e27d7366`：completed、A4 10页、537434 bytes、SHA-256=`5cf60f689fc4197b31d1c6bef8b806632be81250c3618b95dd1183f0a61cfef5`。数据库/服务器文件/本地副本大小SHA一致，10页均非空、十维名称10/10、未解析字段0。
- staging真实PDF实图确认：姓名独占首行；年龄/性别同排双列；手机号作为奇数末项保留左列；时间/时长最后一行双列；各行上部对齐、连续且无空行。终验talent-assessment/nginx/mysql均active，内外health正常，FB-171远端临时文件/短时会话/旧后端残留均0，部署窗口应用关键错误0、最近1500条Nginx 5xx=0。production未修改。

## 2026-09-02 UF-040 / FB-170 饼图数值移至图外

- 用户明确要求“不要数字压着饼图”。当前staging模板541430 bytes/SHA-256=`1e9b88ee4973d890d55ce206b6b1f173c3defb8ba8ad5827f9886787f0ae727b`使用点级`bestFit`、系列级`inEnd`及白色`bg1`标签；FB-169部署后的真实报告因此将左绿2.90、右青3.18显示在色块上。
- FB-170 RED由`TestBugFB170_Phase1GroupPieLabelsStayOutsideChart`固定：不允许`inEnd`，必须保留两个独立点布局，点标签必须使用可见前景色`tx1`。仅改`outEnd`、删除点布局或恢复旧坐标均未通过LibreOffice实图，根因包括3D饼图布局覆盖及白字在图外白底不可见。
- 最终候选`docs/competency-phase1-report-fb170-pie-labels-outside.docx`为536894 bytes/SHA-256=`54b167fcc02737ba44e0432a930e1f5657db57cec87d9063f1ee7c111c4932bf`：点级保留`bestFit`和两个独立图外坐标，系列级改为`outEnd`，点标签由`bg1`改为`tx1`。与staging源模板相比仅`word/charts/chart1.xml`变化。
- 候选经真实Go填充及LibreOffice 26.2.5.2生成A4 11页PDF：3.75位于绿色饼图左侧、3.50位于青色饼图右侧，均有短引导线且不覆盖色块；雷达十个12pt分值、十维名称和五层网格保持清晰。11页均非空，未解析占位符0。
- Word 16只读打开通过，候选SHA不变；上传契约及相邻FB-169/170测试57/57、Go全量测试和build均通过，相关诊断0。
- [staging部署完成] 部署前完整备份=`/opt/talent-assessment/backups/fb170_pie_labels_outside_20260903_000037`，目录root:root 0700、文件0600；数据库gzip 12668675 bytes/SHA-256=`f4fb7c7eb39409a3f35791ac7f4f5c8f1232cdf7a4a302fd5841747691170453`且gzip/manifest全部通过，同时保存后端、旧模板、前端index和凯迪旧报告。仅替换一期模板，生效SHA-256=`54b167fcc02737ba44e0432a930e1f5657db57cec87d9063f1ee7c111c4932bf`、536894 bytes；后端保持`d1623d1a...`、前端保持`cd395b2f...`，数据库结构未变。
- 真实paper=`287347a7-d8e1-4606-9ba0-f352f2e848af`强制重生成原实例`d785dc6d-4074-45d2-b71c-9f93e27d7366`：completed、A4 10页、537625 bytes、SHA-256=`9d1ab77ca7095a0223a4b5cefeb0183b4d7934415cef793fc63752755e48e960`。数据库/服务器文件/本地副本大小SHA一致，10页均非空、十维名称10/10、未解析字段0。
- staging真实PDF实图确认左绿2.90和右青3.18均位于饼图外侧并带短引导线，不覆盖色块；左绿仍对应通用能力、右青仍对应心理素养。雷达十个12pt分值、十维名称和五层网格保持清晰。服务器结构为`inEnd=0/outEnd=1/点布局=2/外链=0`。
- 终验talent-assessment/nginx/mysql均active，内外health正常，FB-170远端临时文件/短时会话/旧模板残留均0，部署窗口应用关键错误0、最近1500条Nginx 5xx=0。production未修改；FB-159前端仍未部署。

## 2026-09-02 UF-039 / FB-169 图表标签样式由模板控制

- 用户反馈在Word模板中调近一级饼图数值位置、调大雷达图分值字体后，强制生成报告仍沿用旧格式。触发paper=`287347a7-d8e1-4606-9ba0-f352f2e848af`；当前staging上传模板541430 bytes/SHA-256=`1e9b88ee4973d890d55ce206b6b1f173c3defb8ba8ad5827f9886787f0ae727b`。
- 结构实证：上传稿chart1仍为3D饼图，两个点分别有独立manualLayout坐标，系列位置已设`inEnd`；chart2主系列分值字体为12pt、隐藏系列为11pt。旧生成代码会把chart1整个`dLbls`重建为固定`outEnd`，并把chart2主系列重建为固定8pt，因此模板修改必然丢失。
- FB-169先取得双RED：饼图两个手工坐标均消失，雷达12pt变8pt。GREEN后只在原标签块内删除饼图静态样例文字、统一动态`0.00`与显示属性，并隐藏雷达重复系列；不再覆盖饼图位置或雷达主系列字体。左右业务映射、颜色、动态值、雷达维度名称和0-5五层网格继续由既有硬约束保护。
- 精确上传稿经真实Go填充后，两个饼图点布局与雷达主系列字体逐字节保留；LibreOffice 26.2.5.2生成A4 11页PDF，饼图3.75/3.50位于图内且贴近扇区，雷达十个分值以12pt清晰显示，五层网格与十个维度名完整。相邻FB-160/161/163/169、Go全量及build通过，相关诊断0。
- 已知限制：当前上传模板和本地基线原件均可由Word 16打开，但两者经过现有完整填充链后生成的内部DOCX均被Word拒绝；LibreOffice PDF正常。该行为并非FB-169引入，内部DOCX当前不作为交付物。此处为部署前状态，已由下方staging部署与真实paper终验关闭。
- [staging部署完成] 发布前备份=`/opt/talent-assessment/backups/fb169_chart_template_styles_20260902_172943`，目录0700、文件root:root 0600；数据库gzip 12668845 bytes/SHA-256=`82a0e68ec300a7c50b7ab80f4e5f0726b9d87c5fd1ad7ff504ee59b56010daa0`且manifest/gzip校验通过，同时保存旧后端、当前模板、前端index和目标旧PDF。新后端SHA-256=`d1623d1a02e3d1b57c91779f5b8fb3bcc60e3744217a313cc59b91c2be332514`；模板保持`1e9b88ee...`、前端保持`cd395b2f...`，数据库结构未变。
- 真实paper=`287347a7-d8e1-4606-9ba0-f352f2e848af`（凯迪，效度good）force重生成completed报告537404 bytes/SHA-256=`0541e66d084d1c64167c8d4c30874fbbe9f158f083447df01eb13baba4f2913f`、A4 10页。DB/服务器/API副本大小SHA一致，10页均非空、十维名称10/10、未解析内部字段0；一级图左绿2.90=通用能力、右青3.18=心理素养，分值贴近图内；雷达十个12pt分值清晰且五层网格、十个名称完整。
- 终验talent-assessment/nginx/mysql均active，内外health正常，FB-169远端临时文件/会话/旧后端及本地产物均0，应用关键错误0、最近1500条Nginx 5xx=0。production未修改；FB-159前端仍未随本次后端发布。

## 2026-09-01 UF-038 / FB-168 个人信息行间距一致

- 用户截图指出FB-167模板的个人信息行间距不一致。staging精确审计：姓名/年龄/性别/手机号/岗位/时间为before=156、after=156、line=144；单位和时长仅line=144，造成视觉节奏不同。
- 为保持用户已选择的方案B，最终不由后端覆盖段落样式；行距继续属于客户Word模板职责。FB-168 RED命中单位/时长spacing不同；最终候选`docs/competency-phase1-report-fb168-spacing.docx`将姓名、年龄、性别、手机号、单位、岗位、时间、时长全部统一为before/after=156、line=144，同时保留模板控制的顺序、位置和字体。
- 候选532193 bytes/SHA-256=`a60314174cdf755740cd03dff4dfbfdb94cc298fea3e82846125d50cc29d8502`。上传契约、FB-130/131/160～168、Go全量及build通过；本地LibreOffice 26.2.5.2生成10页，个人信息实图为姓名→手机号→时间/时长且间距一致，页面其他内容不变。当前未部署staging/production。
- [staging模板部署完成] 仅替换模板，后端保持SHA-256=`51a1bc72db19b8521755716b799d85289077f0a5bb5c46aab1d196d930a44a86`不变。部署前备份=`/opt/talent-assessment/backups/fb168_profile_spacing_20260901_163756`，保存旧后端、旧模板、小鱼及全字段旧PDF，全部0600；旧模板SHA=`bbb44705...`。生效模板SHA-256=`a60314174cdf755740cd03dff4dfbfdb94cc298fea3e82846125d50cc29d8502`。
- 服务器DOCX结构审计确认participant.name/age/gender/telephone/affiliation/post及result.submittedAt/userTime共8个单元格仅1种spacing：before=156、after=156、line=144。模板外部关系/externalData均0。
- 真实小鱼报告重生成554883 bytes/SHA-256=`bf2da30db0a8a86058d03de9003d4ce33ffd949dfc0873cb6d30ea57e8ead0e9`、A4 11页；全字段报告重生成562945 bytes/SHA-256=`9384f1b97b922b1c81d70b52749e1e5dcbfa6d18aaf2f17b7ec3936a71922bc6`、A4 12页。两份个人信息实图行距一致，DB/文件大小SHA一致，所有物理页非空，页码为PAGE-only且无分数页码。
- 终验三服务active、内外health正常、FB-168临时文件/会话0、应用关键错误0、最近1500条Nginx 5xx=0。production未修改。

## 2026-09-01 FB-167 个人信息由模板决定顺序和位置

- 用户在“程序固定顺序”和“模板决定顺序/位置”之间明确选择B。FB-166已部署的程序固定顺序因此需要纠正：模板应继续承担字段顺序、行列位置和样式控制；后端只按requiredFields删除未配置控件、删除空行，并在一行只剩一个字段时跨两列消除空洞。
- FB-167有效RED使用当前模板真实顺序`姓名→年龄→性别→手机号`，旧硬编码会输出`姓名→性别→年龄→手机号`。GREEN不再把字段放入后端固定列表，而是按模板行/单元格顺序原位过滤；时间/时长固定行继续保留并对齐段落间距。测试明确覆盖配置`affiliation/post`时单位和岗位Tag仍存在。
- 为让当前默认模板直接满足示例，同时保持后续可编辑性，基于staging精确模板SHA=`8f491f4a...`制作候选`docs/competency-phase1-report-fb167-template-order.docx`：姓名/年龄/性别/手机号/单位/岗位六个控件各自为模板中的独立全宽行，时间/时长仍为最后一行两列。客户以后在Word中移动这些字段行或把两个字段放入同一行，生成程序将跟随模板。
- 初版候选因Windows原地ZIP更新产生CRC错误已废弃；最终使用全新ZIP条目重建，532197 bytes/SHA-256=`bbb447052b24be72f8b950d2a06d6ac15c87286074b92e37a279da88e74933c6`。上传契约、FB-130/131/160～167、Go全量和build通过；Word 16只读打开11页且SHA不变；本地LibreOffice 26.2.5.2以`name,telephone`生成10页，实图为姓名→手机号→时间/时长，符合示例。
- 当前仅本地完成，尚未部署staging/production。正确发布必须同时部署新后端和候选模板；只部署其一会导致模板语义与程序行为不一致。
- [staging同步部署完成] 部署前备份=`/opt/talent-assessment/backups/fb167_template_order_20260901_161412`：数据库gzip 12667890 bytes/SHA-256=`2ee475685bcc858c0473ff66ea477d892e43b50e1ae418f5dee772201fb3c206`且gzip校验通过，并保存旧后端、旧模板和小鱼/小米旧PDF，全部0600。部署后后端SHA-256=`51a1bc72db19b8521755716b799d85289077f0a5bb5c46aab1d196d930a44a86`，模板SHA-256=`bbb447052b24be72f8b950d2a06d6ac15c87286074b92e37a279da88e74933c6`。
- 真实小鱼paper=`ff08cb88-...`（配置name,gender,telephone）重生成554883 bytes/SHA-256=`8abaaad7a2fab87e8019851e2cacc4ed4cb4d34a4014208bbe7777eacdee0151`、A4 11页，实图为姓名→性别→手机号→时间/时长，无空行；正文页码1～10。真实全字段paper=`93a26b3b-...`（name,gender,age,telephone,affiliation,post）重生成562912 bytes/SHA-256=`f6adab637add64c91e7387c0f4771c9a88f453408db8de0c7280329ea698c8db`、A4 12页，实图依模板顺序显示姓名→年龄→性别→手机号→单位→岗位→时间/时长，单位/岗位值均为21。两份DB/文件大小SHA一致，物理页全部非空且无总页数分数。
- 当前模板静态个人信息行顺序经服务器ZIP审计为name|age|gender|telephone|affiliation|post|submittedAt+userTime；外部关系和externalData均0。终验三服务active、内外health正常、FB-167临时文件/会话0、应用关键错误0、最近1500条Nginx 5xx=0。production未修改。

## 2026-09-01 UF-036 / FB-166 一期报告个人信息紧凑排布

- 用户截图要求个人信息按配置依次、紧凑、上部对齐且无空行；目标示例为姓名、手机号逐行在左侧，最后一行时间在左、时长在右。真实小鱼paper=`ff08cb88-...`的测评配置为`name,gender,telephone`，现象不是字段配置错误。
- 根因是`filterPhase1WordProfileTable`只删除未配置单元格，保留字段继续占原模板两列槽位：姓名左、性别左、手机号右，视觉顺序断裂；时间与时长还继承不同段前/段后间距，基线不齐。
- FB-166首个RED为`name,gender,telephone`过滤后仅3行而非目标4行；增强RED进一步命中时间单元格spacing=`before/after 156`、时长仅line=144。GREEN重新收集配置字段，按`name,gender,age,telephone,affiliation,post`稳定顺序逐项生成跨两列的连续行；未配置字段不生成行；固定时间/时长行保留两列，并把时长段落间距对齐时间。
- 使用staging精确模板SHA=`8f491f4a...`及`name,telephone`测试数据经本地LibreOffice 26.2.5.2实图验证：姓名→手机号→时间/时长，字段连续无空行，时间与时长同一基线，页面其余总体评价内容不变。FB-130/166、Go全量和build通过，相关诊断待最终检查。当前仅本地完成，未部署staging/production。
- [staging部署完成] 部署前备份=`/opt/talent-assessment/backups/fb166_profile_layout_20260901_153046`：数据库gzip 12667952 bytes/SHA-256=`144da96537b3634091a2108d8ea1b9907004456670db336ec6818c97c8ca7e24`且gzip校验通过，并保存旧后端、当前模板和小鱼/小米旧PDF，全部0600。仅部署后端，模板保持SHA-256=`8f491f4ac09b74d8784e08dae79efa8343ff1ded03b25a033b6774e39b58d5f6`不变；新后端SHA-256=`cfb8899f40a41abaef45fe8cb046312b65e8a6a586f3fc97622c9d83acdf169f`。
- 真实小鱼paper=`ff08cb88-...`、配置`name,gender,telephone`强制重生成：PDF 554887 bytes/SHA-256=`53ba9e66813a4689aad921b056bdbe13cb386216334a301d089cbae3063db121`，LibreOffice 24.2、A4 11页。实图确认姓名→性别→手机号逐行左对齐、无空行，时间与时长同排同基线；总体评价及后续内容未错位。正文页码1～10、总页数分数0，数据库实例与文件大小/SHA一致。
- 终验talent-assessment/nginx/mysql均active，内外health正常，模板外部关系/externalData均0，FB-166临时文件和短时会话0，应用关键错误0、最近1500条Nginx 5xx=0。production未修改。

## 2026-09-01 UF-035 / FB-165 Word重存模板外链上传兼容

- 用户反馈从报告模板页下载后，在Microsoft Word中任意修改并保存，再上传即报“模板不得包含外部Excel链接”；截图显示当前生效模板本身为51控件/12图表/0可见占位符。根因是Word重存可能恢复chart关系中的`TargetMode=External`及chart的`externalData`，而旧上传链在任何清理前直接执行零外链拒绝；报告生成链虽已有清理能力，但上传链未复用。
- 修复不放宽安全规则：`UploadPhase1Template`读取文件后，先只针对`word/charts/_rels/*.rels`删除外部Relationship、针对`word/charts/chart*.xml`删除`externalData`；随后仍执行内容控件、图表、页码、外链等全部严格校验，最终只安装清理后的零外链DOCX。响应增加`removedExternalArtifacts`供前端/审计识别；原本零外链的模板返回原始字节，不因无操作清理改变SHA。
- FB-165 RED因`sanitizePhase1WordTemplateUpload`缺失而编译失败；GREEN夹具模拟Word恢复本机`C:/Users/customer/chart-data.xlsx`关系和`externalData`，清理计数2、最终零外链且完整上传契约通过。相邻FB-162/164、Go全量及Linux build通过；待部署后端48,443,747 bytes/SHA-256=`6336348f14efcd33fc71e55d24425e5b206fa6b8e1162e4095b19916135022ec`。
- [staging阻塞] 2026-09-01公网`/prod-api/health`仍返回ok，但TCP/22及SSH连续超时，无法执行远端备份、后端替换和真实上传验证；未上传或修改staging/production。SSH恢复后固定步骤：只读预检→数据库/后端/模板备份→部署后端→用模拟Word重存稿真实上传→确认响应清理计数、服务器模板零外链和契约有效→清理/health/log终验。
- [纠正 / staging部署完成] SSH随后恢复。部署前备份=`/opt/talent-assessment/backups/fb165_word_upload_20260901_140653`：数据库gzip 12667491 bytes/SHA-256=`20893d5e51d98f060fd646ec88114f2088b04326889e5c42133403b8f8711904`且gzip校验通过，并保存旧后端SHA=`8c0da167...`和旧模板SHA=`33f87b94...`，全部0600。最终后端SHA-256=`6336348f14efcd33fc71e55d24425e5b206fa6b8e1162e4095b19916135022ec`；部署过程未直接替换模板。
- 真实认证上传使用当前模板构造Word重存等价稿：输入含1条`TargetMode=External`和1个`externalData`；API返回code=0、`removedExternalArtifacts=2`、`valid=true`。生效模板536952 bytes/SHA-256=`8f491f4ac09b74d8784e08dae79efa8343ff1ded03b25a033b6774e39b58d5f6`，外部关系0、externalData 0、NUMPAGES 0、PAGE字段1；说明系统不是接受风险外链，而是安全删除后才保存。
- 使用真实小鱼paper=`ff08cb88-...`强制重生成并下载：LibreOffice 24.2、A4 11页、554657 bytes、SHA-256=`99791618d5b2920cbc7b1ef9882650e0ccbfa9df001fbefbb4f26c8a6a17bdf8`；正文页码1～10、总页数分数0，数据库实例与文件大小/SHA一致。终验三服务active、内外health正常、模板零外链、FB-165临时文件/会话0、应用关键错误0、最近1500条Nginx 5xx=0；production未修改。

## 2026-08-30 UF-034 当前staging模板页码只读核验

- 用户截图页脚为`10 / 12`。SSH恢复后确认截图对应当前11:17新生成的小鱼/小米报告：两份PDF物理页数均为11，但可提取页码均为`1 / 12`至`10 / 12`；封面不编号，正文实际10页，因此分母12错误。
- staging模板在2026-08-30 11:17:17又被上传替换，当前552679 bytes、SHA-256=`cd712b8e9599a7a28861fdc1ab6cebec05b959385aa467a40ef86720ce9e1bff`，不同于FB-163部署模板`034b1967...`。当前模板`word/footer1.xml`明确包含`PAGE`和`NUMPAGES`字段；Word打开空模板为10页/2节，字段缓存为PAGE=1、NUMPAGES=10，但staging LibreOffice 24.2填充长文案后输出11页并将总页数显示为12。
- 根因是客户新上传模板重新引入跨Word/LibreOffice不可靠的`NUMPAGES`，不是PDF阅读器或数据库页数。此前FB-125已验证稳健方案为仅保留当前页`PAGE`，不显示总页数。当前任务只读检查，未修改或替换11:17客户模板；production未修改。
- [FB-164方案1修复与staging部署] 用户确认采用仅显示当前正文页码。新增上传门禁拒绝任何footer中的`NUMPAGES`并提示仅保留`PAGE`；运行时对历史模板兜底，按Word字段边界删除分隔符和总页数字段，保留原PAGE run及格式。RED为两个新符号未定义，GREEN聚焦页码/模板图表、Go全量和Linux build通过。
- 修复候选基于staging精确SHA=`cd712b8e...`，仅移除footer1的` / NUMPAGES`并按既有兼容规则规范20个重复VML ID，不改可见正文/图表。最终模板547224 bytes、SHA-256=`33f87b94fd586d6e73ed0479d13172a74ffde11a3e3b6f16b581d4b64d18dfc4`；Word 16只读打开10页/2节且唯一页脚字段为PAGE，关闭后SHA不变；本地LibreOffice夹具10个物理页、正文页码1～9、总页数分数0。
- staging备份=`/opt/talent-assessment/backups/fb164_page_only_20260830_114428`：数据库gzip 12643659 bytes/SHA-256=`810dd0111defbfdd21ba910d31389afde3a6fc8a2a1b96a5ac4029d8e7280f88`且gzip校验通过，同时保存旧后端、旧模板和小鱼/小米旧PDF，全部0600。部署后后端SHA-256=`8c0da1677c49c4be7b361acb709abe47dc5227268c5ced658d6c7a9f25cf3faa`，模板SHA-256=`33f87b94...`。
- 真实小鱼paper=`ff08cb88-...`重生成554793 bytes/SHA-256=`7ea9c621...`；小米paper=`0aae19a9-...`重生成554904 bytes/SHA-256=`68b9429d...`。两份均LibreOffice 24.2、A4 11个物理页，正文页码严格为1～10且`当前/总页数`分数计数0；数据库实例与下载文件大小/SHA一致。末页实图仅显示`10`，无`/12`。
- 终验talent-assessment/nginx/mysql均active，内外health正常，FB-164临时文件和短时会话0，应用关键错误0、最近1500条Nginx 5xx=0。production未修改。

## 2026-08-29 客户修正版一期报告模板只读检查

- 客户文件=`docs/competency-phase1-report-fa6c59a3.docx`，实际557663 bytes、SHA-256=`ac4d9515335ba92137c80f1f66b079cb2d2868f62d59eb01b97c2257491840d2`；文件名中的`fa6c59a3`不是当前文件哈希。与工作区运行模板不相同，检查期间未替换模板。
- 上传契约通过：DOCX 82个ZIP部件、重复部件0、读取/XML错误0、51个内容控件/49个唯一Tag、仅两个一级分Tag各重复2次、12图表、0外链、0可见占位符。Microsoft Word 16只读打开成功，为11页、5表、15个inline和14个shape，关闭后SHA不变。
- 客户文件自身的chart1仍为idx0=`通用能力`+青色`#30C0B4`、idx1=`心理素养`+绿色`#75BD42`，因此静态模板本身仍与用户确认的“左绿通用能力、右青心理素养”相反；它不能配合旧后端直接解决FB-163。
- 使用FB-163本地修正版运行时填充同一客户文件后，最终图表为idx0右青=`心理素养`、idx1左绿=`通用能力`，真实LibreOffice 26.2.5.2生成A4 11页PDF。11页均有文本、无未解析占位符；饼图和雷达页实图确认左绿/右青映射正确、雷达十维名称/数值/五层网格完整。
- 客户原chart2有两套分类轴且均delete=0，但值轴仍缺一个min=0和两个majorUnit；现有运行时会统一修正为0-5、主单位1，并规范化标签。结论：模板结构可用且与FB-163新后端组合后输出正确，但模板静态映射未独立修正；当前未部署staging/production，也未替换正式模板。
- [纠正 / 修复并部署] 已从客户原件生成`docs/competency-phase1-report-fb163-fixed.docx`，原件不覆盖。只交换chart1分类与样例值缓存，使idx0右青=`心理素养/3.70`、idx1左绿=`通用能力/3.75`；另将20个重复VML ID按其唯一spid规范化，不改可见内容或布局。最终模板551069 bytes、SHA-256=`034b196740b179ff034010c47cfcfb5be1c7240e3c50269ab8a91f063868a14d`，Word 16只读打开11页且哈希不变，上传契约、FB-131及FB-160～163门禁通过。
- staging部署前备份=`/opt/talent-assessment/backups/fb163_pie_mapping_20260829_193000`：数据库gzip 12642298 bytes/SHA-256=`e28d4b08380147f669488557dfc6aa0e7a49fcc8c7ac59f3842e379b8d45ef84`且gzip校验通过，并保存旧后端、旧模板和待重生成旧PDF，文件均0600。部署后后端SHA-256=`18561c8daf5f8cde8ebe33fe29c30ad9d385ef76dc48f997227cf7713bb16a83`，模板SHA-256=`034b1967...`。
- 真实paper=`ff08cb88-10b8-4b01-a90d-457784a3a980`数据库一级结果为通用能力3.100000、心理素养3.250000；force重生成并下载成功，最终PDF 548338 bytes、SHA-256=`a01c06920f608a15a9ff70954c47891e178039c3d9e50dfa8c4b83554d179ae3`，LibreOffice 24.2、A4 11页。实图确认左侧绿色3.10对应通用能力、右侧青色3.25对应心理素养；雷达图十维名称、分值和五层网格保持正常。数据库实例/下载哈希大小一致，最近两条审计为regenerate/download成功。
- 终验talent-assessment/nginx/mysql均active，内外health正常，临时发布文件和短时会话均0，部署窗口应用关键错误0、最近1500条Nginx 5xx=0。production未修改；本次未发布前端FB-159。

## 2026-08-29 FB-163 一级饼图位置/颜色业务映射修正

- 用户确认正确口径为左侧绿色=`通用能力`、右侧青色=`心理素养`。只读检查发现活动模板物理点固定为idx0右青`#30C0B4`、idx1左绿`#00B050`，但运行时原先按`通用能力/心理素养`写入idx0/idx1，业务含义确实反向。
- RED精确复现运行时值为`[3.10,3.25]`；GREEN改为idx0写心理素养3.25、idx1写通用能力3.10，并同步分类缓存与可选内嵌工作簿的业务键、名称和值。正文表格仍按“通用能力/心理素养”展示，不受图表物理点顺序影响。
- 真实活动V1模板渲染回归同时验证分类、数值和两点RGB；聚焦4/4、Go全量及build通过，相关诊断0。现已随上述最终客户模板部署staging并用真实报告验证；production未修改。其他既有报告需force重生成后才会更新。

## 2026-08-25 测试脚本与垃圾文件治理清理

- 经用户确认执行四类清理：删除明确生成物/依赖缓存；归档FB-040历史验收；归档旧staging/39服务器部署、字体/PDF诊断及错放SQL；删除全部含明文凭据构造模式的历史脚本。当前活跃VS Code任务、CI入口、一期转换/CSV契约、移动答题E2E、结果按钮E2E和模板回归脚本均保留。
- 已删除test根目录4张FB040截图、无断言debug脚本、重复的test版PM截图脚本、`scripts/test/node_modules`和Python缓存；2个FB040 test脚本移入`test/archive`，35个历史工具移入`tools/archive`。extensionless Go覆盖率产物已删除，并新增`.gitignore`规则防止重新进入Git。
- 安全清理共删除49个含明文凭据构造模式的历史脚本：先删除1个旧生产部署脚本，后续文件名扫描再确认并删除48个（47个未跟踪归档脚本及已跟踪旧`db-sync-v2.sh`）。删除后同类标记计数0。历史凭据是否仍有效无法由代码库确认，服务器侧仍应按安全流程轮换。
- 当前Go/Vue业务改动中的批量ZIP、模板图表、测评者筛选、结果列表与移动选项符号均有真实调用/路由/测试引用，未发现可证明的死业务代码，因此未删除生产代码。清理后test根目录PNG=0、node_modules不存在、`__pycache__`不存在、Go tmp/bin/dist产物均为空；`git diff --check`通过。
- 第二轮按全仓文件名引用图继续收口：额外将36个零外部引用的一次性API录制/差异、Excel/PDF比较、dump/extract、旧MBTI生成、截图、模板定向变换和单次验证工具移入`tools/archive`；保留有CI/文档/测试引用的转换器、CSV校验、Word模板构建器，以及当前staging容量/安全/导出/Worker/传统链验收器。工具根目录现仅保留上述可复用入口。
- 终验：活跃test/tool根文件分别47/35，JS 35个与Python 32个语法检查通过；Go全量和前端26文件165项通过；明文凭据构造标记0，test根PNG=0，node_modules/pycache/Go coverage=0，`git diff --check`通过。当前形成50个已跟踪删除记录（覆盖率、敏感/过时脚本及归档迁移），未提交、未部署。
- 第三轮将依赖只读Legacy Redis、固定本机端口/旧admin口令的7套旧回归，以及local/server acceptance、production-demo和旧performance baseline共11个脚本移入test archive；其8个无引用截图/日志fixture一并归档。`.vscode/tasks.json`移除已失效任务，改为当前参与者E2E和一期模板契约任务，并修复原乱码/非法JSON；迭代prompt、README、UX规则同步改为Go/Vitest和当前Playwright入口，活跃文档不再引用旧suite。
- 测试包原`npm test`固定失败且声明不存在的`index.js`，现改为一期转换/身份契约入口，并增加参与者E2E入口。任务JSON、package/package-lock均可解析，新模板契约任务真实通过。第三轮后活跃test/tool根文件为36/35，旧suite活跃引用0。
- [纠正 - 2026-08-25] 上述第二轮“50个已跟踪删除记录”为中间状态；第三轮继续归档14个已跟踪旧suite/工具后，最终为64个已跟踪删除、28个已跟踪修改。Go全量、前端165项、活跃脚本语法和专项任务均通过；未提交、未部署。

## 2026-08-25 UF-033 最新staging模板饼图/雷达图只读审查

- 客户反馈上传新稿后饼图和雷达图仍异常。本轮严格只读检查：未替换模板、未调用正式报告生成接口、未写数据库。staging当前生效模板580362 bytes、mtime=`2026-08-24 16:30:51 +08:00`、SHA=`896b59e5f17608b5a890c4d72eb5155963ce6b1b487015f83ef92ea3db510214`；即时前版备份SHA=`3014de49...`，FB-155已验证基线SHA=`18a4e608...`。
- 饼图根因确认：当前chart1相对已验证基线新增手工富文本数据标签`3.7/3.75`，运行时只更新数值缓存，不会替换该手工文本；同时分类顺序变为`心理素养/通用能力`，与运行时固定`通用能力/心理素养`相反，格式从`0.00`退回`General`。真实小鱼报告正文一级分为`3.10/3.25`、小米为`2.65/2.38`，但两份饼图都显示样例`3.75/3.70`，证明不是单份数据问题。
- 雷达图根因确认：当前chart2第一个分类轴由基线`delete=0`变为`delete=1`。后端FB-155运行时仍把两个值轴修正为min=0/max=5/majorUnit=1，因此五层网格存在，但LibreOffice输出完全隐藏十个维度名称；每点数值继续显示并与折线/相邻数字重叠。两份16:31真实报告均复现。
- 额外门禁盲区：当前V1模板有12条`TargetMode=External`图表关系，chart1明确暴露作者本机`E:\...\260805数据图表.xlsx`路径；现有上传校验只对V2拒绝外链，V1仅计数仍判有效。模块分页门禁通过；结构门禁失败项为零外链、chart1两位小数及原生标签居中。当前模板最终上传相对即时备份只改`document.xml/settings/core/app/chart7`，说明饼图/雷达图异常已在更早一轮客户重存中引入，不是最后一次chart7编辑单独造成。
- 使用现有Go渲染器从精确模板生成测试DOCX，再在staging LibreOffice 24.2.7.2隔离`/tmp`转换为A4 10页PDF，未落报告实例；结果与两份真实报告一致。建议最小修复为：保留客户正文/分页和最后chart7改动；chart1去除手工标签、恢复`通用能力/心理素养`顺序与`0.00`；chart2恢复分类轴名称并调整数值标签避免压线；清除全部V1外部Excel关系并把零外链加入上传强制门禁。具体雷达图是否保留每点分值需客户确认后再写RED测试和修改。
- [客户示例确认与本地GREEN] 用户提供的目标图明确保留3D饼图、青色/绿色两块及外侧两位小数；雷达图选择“维度名称+分值”。FB-160～162先RED为4个缺失符号编译失败，再实现运行时规范化：chart1固定`通用能力/心理素养`顺序、清手工标签、动态`0.00`外侧值；chart2恢复两个分类轴、保留五层网格、仅主系列显示8pt两位小数并隐藏重复标签块；V1/V2上传统一拒绝外链，运行时同时清关系和`externalData`。
- 候选从当前staging SHA=`896b59e5...`定向修改，仅处理chart1/chart2及12个图表外链，保留客户最新正文、分页和chart7；候选544237 bytes/SHA=`fa6c59a3588388d3cf52b836683d7c91c813a12507b580e21c783662e044666b`。零外链、两位小数、模块流和生产上传契约通过；Microsoft Word 16只读打开10页/5表/15 inline且关闭后SHA不变。
- 候选经新运行时填充后，由staging LibreOffice 24.2.7.2隔离转换为A4 10页、513914 bytes、PDF SHA=`9f2e6764...`。页面实图：3D饼图显示动态`3.75/3.50`并与正文一致；雷达图显示十个维度名、十个分值和五层网格，未见压线/相邻数字重叠。Go聚焦、新旧兼容、全量测试和build通过，相关诊断0。当前候选和后端均未部署staging/production，未提交；正式发布仍需用户明确授权。
- [staging部署与真实终验] 预检旧后端SHA=`eacc49db...`、旧模板=`896b59e5...`。完整备份=`/opt/talent-assessment/backups/fb160_162_chart_fix_20260825_134212`：数据库gzip 12285040 bytes/SHA=`8ce4f62b23a09d0d9cb1a7ca254ed1b19bb74f8568bf49756f09a3af4f7e451d`且gzip校验通过，并保存旧后端、旧模板和小鱼/小米旧PDF，全部0600。
- 生效后端SHA=`30e7d74223d25e5b1928f5af38dd4d23f0e535eb0adf4b163da2b740a57e4ae6`，模板SHA=`fa6c59a3588388d3cf52b836683d7c91c813a12507b580e21c783662e044666b`。真实小鱼paper=`ff08cb88-...`重生成557353 bytes/SHA=`36aff8a8...`：饼图3.10/3.25；真实小米paper=`0aae19a9-...`重生成558040 bytes/SHA=`25c61048...`：饼图2.65/2.38；均与同页正文一致。
- 两份真实LibreOffice 24.2报告均为A4 11页且11/11非空；雷达图均显示十个维度名、十个两位小数分值和五层完整网格，逐图复核无数字重叠、压线、裁切。数据库记录与下载副本大小/SHA一致。同期FB-158真实API“是”8/8、“否”19/19通过。
- 终验talent-assessment/nginx/mysql均active，内外health正常，短时会话0、远端发布/验证文件0、应用关键错误0、最近1500条Nginx 5xx=0。production未修改，前端FB-159未随本次发布；当前代码未提交。

## 2026-08-25 FB-159 胜任力结果列表删除评价均值

- 用户截图明确要求删除胜任力测评结果列表中的“评价均值”列。本切片仅删除`competencyResults.vue`的该表格列及行绑定，不删除后端`evaluationAverage`字段，不改变计分、数据库、排序参数、详情、三Sheet导出或报告中的评价均值用途。
- RED为前端26文件中1项失败/164项通过，命中列表仍含`label="评价均值"`；GREEN为26文件165项全部通过。production build通过，仅保留既有2个asset/entrypoint体积warning；目标文件诊断0。当前未部署staging/production、未提交。
- [staging部署 - 2026-09-03] 前端已原子切换，生效`index.html` SHA=`b4a82e212d7b199eb703bd942ff4217d93c3ab44d8a00b2c1103cb9cbce3e525`，共379个文件；Nginx实际返回的两个胜任力结果列表资源均含页面标识且不含“评价均值”。报告页仍保留该指标（全量资源中1个文件），符合本切片边界。受限备份=`/opt/talent-assessment/backups/fb159_remove_evaluation_average_20260903_093152`；后端SHA=`1b0938ce...`、报告模板SHA=`54b167fc...`均未改变。终验三服务active、内外health正常、外部index HTTP 200且SHA一致、应用关键错误0、Nginx 5xx=0、发布临时文件和旧dist目录均为0。production未修改，代码未提交。

## 2026-08-25 UF-032 / FB-158 测评者是否学生筛选本地修复

- 截图页面为封闭测评者管理。前端`queryParams`已发送数值`stuFlag=1/0`和`telephone`，但后端`TesterHandler.List`只读取姓名、身份证、测评和答题状态，COUNT与带JOIN的行查询均忽略是否学生和手机号，因此选择“是/否”仍返回混合数据。staging只读数据分布为有效tester中`stu_flag=0` 19条、`stu_flag=1` 8条、NULL 0条。
- 最小修复新增共享`applyTesterListIdentityFilters`：手机号使用精确匹配；`stuFlag="1"/"0"`分别绑定数值1/0；空值不增加条件。COUNT查询和行查询分别以无前缀/`t.`前缀调用同一过滤器，避免总数与当前页不一致。其他消费者未传这两个参数时SQL不变；传统测评详情与统计页面无需同步修改。
- TDD证据：RED因共享过滤器未定义而编译失败；GREEN专项验证是、否、全部、手机号以及COUNT/行查询双调用。Go全量测试和build通过，两个改动文件诊断0。当前仅本地完成，未部署staging/production、未提交。

## 2026-08-25 UF-031 / FB-157 胜任力报告批量下载ZIP本地修复

- 00401结果页旧`batchDownloadReports()`逐个调用单份PDF接口并逐次`saveAs`，因此“批量下载”会触发多个浏览器下载；这不是后端报错或近期回归，而是初始实现未提供胜任力报告实例的批量ZIP契约。
- 新增管理员权限保护的`POST /exam/api/competency/reports/batch-download`，请求一次提交`paperIds`。后端去空、去重并限制最多100份；批量读取结果和completed报告实例，逐项保持单份下载相同的冻结版本/一期内容批准门禁，并复用允许目录路径校验。任一报告缺失、未批准、版本无效或文件无效时，在发送下载头前整体拒绝。
- ZIP先写0600系统临时文件，成功后为每份报告在同一事务写download审计，再以`application/zip`、RFC5987中文文件名和Content-Length响应；finally删除临时ZIP。包内文件名为`姓名-paperId-胜任力测试报告.pdf`，姓名和paperId均移除路径分隔符，paperId保证同名人员文件唯一。前端只发一次请求并只调用一次`saveAs`，单份PDF下载保持不变。
- TDD证据：RED为前端26文件中2项失败、161项通过，失败均为archive API调用0次。GREEN为Go聚焦2/2（实写并实读含2份PDF的ZIP、同名唯一、无路径分隔符，以及空/重复/100份上限）、Go全量和build通过；前端26文件164项和production build通过，build仅有既有2个asset/entrypoint体积warning；相关诊断0。当前未部署staging/production、未提交。
- [staging部署完成 - 2026-08-25] 本次统一发布FB-156与FB-157。预检确认`20.200.136.133`的talent-assessment/nginx/mysql均active、内网health正常、根分区可用约49.1GiB；旧后端SHA=`fd371a2a...`、旧index SHA=`d9678225...`。
- 部署前完整备份=`/opt/talent-assessment/backups/fb156_fb157_20260825_113333`，数据库gzip 12284962 bytes/SHA=`947c97188eebab1fac9c186ed09e59f0c497a3185dd4f18ed02c0d94f8798609`且`gzip -t`通过；旧后端和旧dist归档SHA分别为`fd371a2a...`、`c1089a38...`，目录0700、文件0600。
- 生效后端SHA=`eacc49db780e87a1bc67f3b03dbfa4f99627f895f8eeea769eb7245d22e0fa02`，前端index SHA=`cd395b2fa7cbecd46fb4679f018614158c6a08a9aa556c29197faaf031d378be`；公网index原始字节15874 bytes且SHA一致。真实staging Chromium在390/768/1440三视口通过FB-156，其中手机五项left/right/width完全一致为29/361/332，scrollWidth=clientWidth=390。
- FB-157使用两份既有completed报告执行真实管理员认证POST；单一响应为881097-byte `application/zip`，含2个唯一文件名和2份`%PDF`有效内容，下载审计原子增加2。短时Redis会话0，远端上传/解压/验收文件0，本地dist/发布包/截图/验收脚本0；三服务active、内外health正常，部署窗口应用关键错误0、最近1000条Nginx 5xx=0。production未修改，当前未提交。

## 2026-08-25 UF-030 / FB-156 手机端五级选项左对齐本地修复

- 用户截图为staging微信内置浏览器中的00401一期答题页；五个选项卡左边界不一致。真实Chromium RED量化为390px下left=`29,39,39,39,39`，根因是Element UI全局规则对相邻`el-radio.is-bordered`增加10px `margin-left`，当前组件只重置普通margin，未精确覆盖该高优先级相邻选择器。
- 修复仅修改`competencyExam.vue`视觉层：卡片显式`box-sizing:border-box;width:100%;min-width:0`，精确将相邻bordered-radio的margin-left归零，并固定内部radio input不伸缩；不改答题数据、保存、自动下一题、倒计时或提交逻辑。
- GREEN真实Chromium：390×844五项均left=29、right=361、width=332、高48px，页面scrollWidth=clientWidth=390；768×1024保持3列、每项210×52；1440×900保持5列、每项203.6×52。题号导航和操作区触控门禁继续通过，截图人工确认五项边框与圆点纵向整齐。
- 前端全量26文件162项、production build通过（仅既有asset/entrypoint体积2个warning），相关文件诊断0。当前未部署staging/production、未提交；staging仍为RED截图所示旧前端。

## 2026-08-24 UF-029 / FB-155 雷达图五层网格staging修复

- 用户截图指出一期报告“各维度得分情况”雷达图仅显示最外层虚线多边形。线上模板chart2实证包含两套`radarChart + catAx + valAx`：两个值轴虽都有`majorGridlines`和max=5，但`majorUnit`均缺失，且其中一个轴没有显式min=0；真实good报告PDF按原样复现只有一个外框。
- FB-155先取得结构RED：生成后的两个值轴未同时满足min=0、max=5、majorUnit=1和majorGridlines。修复在运行时按业务图表映射定位雷达图（兼容V1物理chart2和V2业务键），逐个值轴冻结0-5范围及1分主单位，不修改客户Word模板、十维数据、折线、标签或颜色。
- 本轮预检确认模板又由此前`8500de32...`更新为当前SHA=`18a4e608629700eeb9ecdda24e485df07fcb2021cf1449c029bcdb0420f780c7`、577678 bytes；候选和最终部署均基于这一精确线上模板。部署前完整备份=`/opt/talent-assessment/backups/fb155_radar_grid_20260824_123500`：数据库gzip 12634797 bytes/SHA=`bbff79ea236c40ebd8cc1f1e4ebee833cdcae9dbfd5bcdc9c4629a04c9143a83`，并保存旧后端、当前模板和两份旧PDF，权限0700/0600且gzip校验通过。
- 最终后端SHA=`fd371a2a7c655c7717ff15640c1b798eca241b2a6ab4be2b23b7b0c1a34463e2`，模板保持`18a4e608...`。real good paper=`4bb5506b-3ba5-4c43-b7e7-2a117287bc3b`重生成A4 11页、507919 bytes、SHA=`02df96a759377df71ba25e8f5190bc2ddd1b64b7c20e93250bf92e2bdb8c2e3e`；real questionable paper=`658dc083-6216-4373-93b3-a7b1d188ec44`重生成A4 11页、507290 bytes、SHA=`6ac8f284b1d2b17fd68d2329a0baacbf09752ffcac96556b23df4f749591c62d`。数据库与服务器文件大小/SHA一致。
- 两份真实雷达页均由staging LibreOffice 24.2.7.2渲染出五层完整同心十边形，维度名、分值和蓝色折线完整；FB-154环形图180-DPI中心门禁两份均保持10/10。22个物理页接触表逐页复核无整页空白、重叠、裁切或模块错序。Word专项38/38、Go全量/build、诊断通过；三服务active、内外health正常、应用关键错误和最终窗口Nginx 5xx均0。production未修改。

## 2026-08-24 UF-028 / FB-154 环形图非粗体与目标渲染器居中验证

- 用户截图要求一期报告二级维度环形图内数字居中且不加粗。现网模板SHA=`41a92df65a78c4c5b76b96861d90a913b4e2900b7a67e336038f460397699ecc`中chart5/6/7/10仍含可见标签`b="1"`，十图字体属性不一致；旧坐标校准针对粗体字形，不能直接复用到统一非粗体后的字形度量。
- FB-154先取得双RED：生成DOCX格式回归命中缺显式非粗体/混合粗体；当前本地LibreOffice PDF的180-DPI门禁无法在一个圆环孔内找到分值。修复在运行时对chart3–12统一写入`b="0"`、垂直`anchor="ctr"`和水平`algn="ctr"`，兼容V1富文本和暂停使用的V2无文本属性图表；Word原模板不改。
- SSH恢复后读取目标环境为LibreOffice 24.2.7.2，并下载精确线上模板。候选DOCX仅上传到`/tmp`，未替换应用；经过四轮真实目标渲染反馈校准，最终候选由staging LibreOffice生成A4 10页，180-DPI十图中心偏差强制门禁10/10通过（横纵绝对偏差均不超过2px），且生成DOCX十图均无`b="1"`。
- 本地一期Word专项37/37、Go全量、Go build和Python像素脚本语法均通过，相关文件诊断0。当前未部署后端、未重生成数据库中的正式报告、未提交；production未修改。
- [staging部署与逐页终验 - 2026-08-24] 用户选择保留staging当前模板。预检时模板已由先前`41a92df...`更新为SHA=`8500de32b5ae07d426e8c6a3ccfd53da41cf9fd4c3561b81513dacd21113f78f`、576646 bytes，因此重新下载精确线上模板并从零执行LibreOffice 24.2.7.2门禁；没有回退到Git内旧模板。
- 部署前完整备份=`/opt/talent-assessment/backups/fb154_chart_center_20260824_114502`，数据库gzip 12634547 bytes/SHA=`3542bb137aa3d4b72092eeda28cbb2a47fc7fd0f41df2348f4cd3578612379d2`，同时保存旧后端SHA=`6a97abb2...`、当前模板及good/questionable旧PDF，目录/文件权限0700/0600且gzip校验通过。
- 真实长文案与两组不同分值暴露短夹具盲区：原生`dLbl`即使写delete/showVal仍被LibreOffice按分值重排，chart7还显示第二数据点标签。最终方案删除chart3–12整个原生`dLbls`，以固定位置的非粗体两位小数标题覆盖圆心；不改客户Word模板、图表数据、颜色或环宽。最终后端SHA=`e59422408c40a20867ee5ca9e18f5794f3773fa75e56af0fe68e8531ebbe23bd`，模板保持`8500de32...`。
- real good paper=`4bb5506b-3ba5-4c43-b7e7-2a117287bc3b`重生成A4 11页、507678 bytes、SHA=`13860c30876e54fd4aa8a9ec9661420e4843b8b1b6c22597ef1a616b10dddcac`；real questionable paper=`658dc083-6216-4373-93b3-a7b1d188ec44`重生成A4 11页、507049 bytes、SHA=`6911df8efa757fc82d69699a428d2d6f36f63da9229bb57f298186bbf5a4de11`。数据库、服务器文件和API下载大小/SHA一致；good提示整段缺失，questionable完整提示保留。
- 两份真实PDF的180-DPI十图中心偏差门禁均10/10通过，横纵绝对偏差均≤2px。22个物理页逐页均有文本且为A4；两份11页接触表逐页人工复核无额外数字标签、整页空白、重叠、裁切、模块错序或字体加粗。Word专项37/37、Go全量/build和诊断通过。三服务active、内外health正常、应用关键错误0；11:48服务重启瞬间出现1次本机health 502，最终12点发布/重生成窗口Nginx 5xx=0。production未修改。

## 2026-08-21 UF-027 / FB-153 效度提示条件展示staging终验

- 用户确认一期报告只在效度存疑时显示“提示：”整段文字；效度良好时连同“提示：”前缀整段隐藏。
- 根因是Word/PDF主路径和Vue/Chromium兜底路径都直接渲染`validityText`，未读取冻结结果中的`validity.status`决定段落可见性。
- FB-153先取得双RED：Word渲染因缺条件状态协议失败，Vue实测good时提示节点仍存在。修复后Word渲染携带内部效度状态，good时删除包含`validity.notice`控件的唯一完整段落，questionable时保留并替换正式文案；Vue仅在`questionable`时渲染节点。正式文案快照和模板Tag继续保留，不改数据库、评分或客户Word模板。
- Word专项36/36、前端全量26文件161项、Go全量、Go build和production前端build均通过；新增编辑器错误0。
- [staging部署阻塞 - 2026-08-21] 用户明确要求部署staging。最终Linux后端48,358,202 bytes/SHA=`6a97abb2c0c587373727c7a2f7efefdbc9cc24005e8a6031a8c701e07b1d01ee`，前端index SHA=`d96782252174dbdf11372bf3007d6c1abce40763a6981b3f768edd439276db60`，379文件；前端归档7,466,612 bytes/SHA=`c4250abd580ec0fe5f3f9eefe1076a2d29a1763005d2af2cc3a8bccc99235a50`，均保存在ignored `tmp/`等待恢复。
- staging预检连续三次失败：SSH ConnectTimeout=15/20秒均超时，TCP/22、80、443均不可达，公网health也超时；同机到1.1.1.1:443正常，客户端公网IP仍为`20.239.176.250`，证明是目标主机/网络不可达而非本机断网。当前Azure CLI两个可用订阅均查不到`20.200.136.133`，无法代为启动VM或调整网络。遵守“先备份再部署”，远端备份命令未启动、应用未上传/替换、报告未重生成；production未修改。主机恢复后继续：只读预检→完整数据库/后端/dist备份→部署后端和前端→用good与questionable真实报告验证整段隐藏/保留→清理/health/log终验。
- [staging恢复并完成 - 2026-08-21] TCP/22、80恢复后，预检确认`talent-assessment`、nginx、MySQL均active且内网health正常。部署前备份位于`/opt/talent-assessment/backups/fb153_validity_notice_20260821_212256`：数据库gzip 12,605,849 bytes/SHA=`0aa70b981a1f6a74c25787ccbefa6a85908730305bedf519d782e2f855d3eecd`，并保留旧后端、dist、模板及两份验收PDF；旧后端SHA=`228bca0060e0aae7c39f8f1a3ac5de1af8b4da1c579ac9481f6a153c1d830311`。
- staging已部署后端SHA=`6a97abb2c0c587373727c7a2f7efefdbc9cc24005e8a6031a8c701e07b1d01ee`和前端index SHA=`d96782252174dbdf11372bf3007d6c1abce40763a6981b3f768edd439276db60`。真实good paper=`af1ebf1b-1b8a-442e-83dd-b7a64541760c`强制重生成11页PDF 514,178 bytes/SHA=`469fc4c716f240b91ea1ade60b60f8c606b0df6b5fd3bc6bab5ed6cf58fa970a`，全文无“提示/效度/掩饰/真实想法”；真实questionable paper=`5c636031-521b-4be9-a3a1-47ade0a166e2`强制重生成11页PDF 516,720 bytes/SHA=`5964e2440a169821800f65e7ef4aef3dfea76d18801b7500331c2f7143f3e90d`，包含完整“提示：该受测者存在掩饰真实想法的可能性…”段落。数据库实例状态、SHA和大小与下载结果一致。
- 部署前端在真实Chromium的1440×900和390×844两种视口均通过：good无`.phase1-validity`节点，questionable显示精确整段文字。内外health均`{"status":"ok"}`，三服务active，部署后应用关键错误0、Nginx 5xx为0，远端临时文件清零；production未修改。

## 2026-08-18 UF-026 / FB-152 最新模板重部署与图表居中终验

- 用户再次保存的客户模板原始SHA=`565c640349cccbb8a0e6a91d64f3f59c4cd85d54f498a8ac063ca151e5c65284`、538720 bytes。结构RED确认一级/二级之间的叠加显式分页再次出现；最小修复只删除该分页，保留本次标签字体和版式调整。最终模板SHA=`19c0f1d4c6474781f98d8761d72bedaaed9b51a45e2e03e8c5b582b1d6ab9b2f`、531944 bytes，Tag边界、零外链、模块流和两位小数四项门禁通过。
- Word 16真实只读打开最终模板为10页、5表、15 inline、5 shapes，关闭后SHA不变；工作区正式运行模板已同步为同一SHA。本地LibreOffice短夹具为A4 10页，但后续真实长文案验收证明短夹具不能替代staging校准。
- staging首轮真实paper=`4bb5506b-3ba5-4c43-b7e7-2a117287bc3b`生成A4 11页，180-DPI像素门禁9/10失败；第二轮仅chart11仍偏移`x=8px,y=-3px`。根据真实PDF反馈更新chart3–12的LibreOffice转换前校准，第三轮最终后端SHA=`228bca0060e0aae7c39f8f1a3ac5de1af8b4da1c579ac9481f6a153c1d830311`。
- 当前数据库实例、服务器文件和API下载均为512403 bytes、SHA=`3bd9b279ae8a63618b3ac1e12fba602ede226dbe8a23a9a5c1ab39cd3ecf72d3`，LibreOffice 24.2、A4 11页。180 DPI下chart3–12十图的数字字形中心与圆环白洞中心横纵偏差全部不超过2px，10/10通过；汇总、一级图和分析区均为实际`通用能力3.10 / 心理素养2.75`。
- 11页逐页总览无整页空白、重叠、裁切或模块错序。部署前正式备份保留于`/opt/talent-assessment/backups/fb152_latest_template_20260818_164414`，数据库gzip 12582798 bytes、SHA=`bbfcd8f3854645ad3148ad81fc342fdc8f412e8d4b1bc8b406d7cabc4420804d`。
- staging与本地过程文件、临时后端、PDF、截图及短时会话均清零；三服务active、内外health正常、应用关键错误和Nginx 5xx均0。production未修改。

## 2026-08-18 UF-025 / FB-151 历史报告复用本地修复

- 用户指定报告`456-4bb5506b-3ba5-4c43-b7e7-2a117287bc3b-胜任力临时测试报告`与staging当前模板不一致。只读核验实际实例PDF创建于2026-08-13 23:14:47，LibreOffice 24.2、A4 12页、538378 bytes、SHA=`63b7ff4f9f0bfd82039851cf788b1066cc845d7acc444a980632a070122cb8ec`，页脚仍显示旧`/13`；不是当前模板SHA=`50b238cc...`的10页产物。
- 根因是结果页单份和批量“生成报告”均发送`force:false`，后端对completed实例按设计直接reuse；用户点击生成并未真正重渲染。最初怀疑模板示例值不一致，已由实例时间、文件和运行时同源替换证据排除，FB-150标记superseded。
- FB-151测试先取得4项RED，修复后两处显式生成均发送`force:true`。聚焦19/19、前端全量26文件160项、production build、模板分值契约、Go一级动态控件/3D图表测试和真实本地LibreOffice页数集成门禁均通过，编辑器诊断0。
- 遵照用户“先测试、不部署”，未部署staging/production，也未重生成指定paper。staging真实GREEN需用户后续明确授权部署前端后再对该paper执行生成、下载和模板逐页对比。
- [纠正 - 2026-08-18] 用户随后明确授权部署并删除过程文件。staging部署前完整备份=`/opt/talent-assessment/backups/fb151_force_regenerate_20260818_160704`，数据库gzip 12582859 bytes/SHA=`3c4fdc5e7301c07404151e186b2a6c40c8eabd39495e51f3a96d42fb3b3ddd2e`，并保存旧dist、后端、模板和旧报告，目录/文件0700/0600。
- 新前端index SHA=`9e51b0b4c19201c4ff37ef57152a7d79a3905b2764eb5cec03d40993c60599c3`，本地/远端/公网原始字节一致。指定paper经真实API `force=true`重生成LibreOffice 24.2 A4 10页、502212 bytes、SHA=`b80c58a42e929517fcaf4712547bb2ed11786ac7dfd9bc6fbd26e7c7fdfb02ca`；数据库实例、服务器文件、API下载三方一致，regenerate成功审计1。
- 新PDF汇总表、一级饼图和分析区均显示同一实际值`通用能力3.10 / 心理素养2.75`，旧静态`3.75/3.70`计数0。十页逐页总览无整页空白、重叠、裁切或模块错序。自动环形图像素脚本因该paper部分两位小数在窄标签框中换行而无法识别字形，未作为本次格式一致性通过证据；人工总览确认图表位于对应单元格内。
- staging和本地本任务过程文件、临时dist回滚目录及短时Redis会话均清零；三服务active、内外health正常、应用关键错误和Nginx 5xx均0。正式备份保留，production未修改。

## 2026-08-18 UF-023 / FB-144 手机端考生标题本地删除

- 用户确认只删除考生公开流程在手机浏览器顶部显示的系统名称，管理后台继续保留系统标题。
- `App.vue`按路由`hideSystemTitle`元数据返回空浏览器标题；考生入口、信息、准备、001/002/003/00401答题、结果及完成共13个路由启用该标记，管理端路由不启用。
- FB-144先取得RED 2失败/1通过，修复后专项3/3；前端全量26文件160项通过，production build退出码0，仅保留既有2个资源体积warning；相关文件编辑器诊断0。当前仅本地完成，未部署staging或production。
- [纠正 - 2026-08-18] 上述“未部署staging”已失效。FB-151统一发布的前端index SHA=`9e51b0b4c19201c4ff37ef57152a7d79a3905b2764eb5cec03d40993c60599c3`已包含FB-144；本次重建本地index得到同一SHA，远端379文件且公网原始字节一致，无需重复覆盖。真实Chromium经考生登录→准备→00401答题后`document.title`为空；管理标题保留由同一部署包的专项3/3覆盖。production未修改。

## 2026-08-18 UF-022 / FB-141 全测评断点续答本地完成

- 用户确认断点续答适用于001/002/003/00401全部测评；具体测评ID和考生ID未知。本切片只处理断点续答，手机浏览器标题删除按“一次一个逻辑变更”留待下一步。
- 已核验既有后端和准备页会恢复同一份进行中试卷及已保存答案；MBTI页面原本已定位第一道未答。缺口只在前端恢复位置：传统整页停留顶部、传统单题页和00401停留第1题。
- FB-141先取得RED 3失败/1通过，再实现GREEN 4/4：传统整页在渲染后滚动到第一道未答；传统单题页和00401直接打开第一道未答；MBTI既有行为保持。以“前10题已答、第11题未答”验证，已答数量和答案状态不变。
- 前端全量25文件157项通过，production build退出码0，仅保留既有2个资源体积warning；改动文件编辑器新增错误0。当前仅本地完成，未部署staging或production。
- [纠正 - 2026-08-18] 上述“未部署staging”已失效。FB-151统一发布的同一index SHA=`9e51b0b4...`已包含FB-141。本次前端全量26文件160项和production build再次通过；staging临时00401考生真实API答10题后重新登录，登录返回paperId、restore返回paperId均保持`0ce1312a-...`，paper-detail为已答10/未答80且第11题未答。真实Chromium按正常登录→准备→答题链打开第11题，导航活动项为“第11题，未答”。测试tester/paper/paper_qu/result清理计数`0/0/0/0`，临时浏览器脚本和profile已删除，production未修改。

## 2026-08-17 FB-125 Word→LibreOffice PDF版式兼容修复

- staging当前V1模板生成的真实PDF已逐页复核：物理12页，第3页几乎空白，第5页雷达图说明与逻辑思维卡片严重重叠，页脚缓存显示`5 / 13`，与物理页数不一致。
- 根因不是报告数据或图表缓存，而是Microsoft Word与LibreOffice对浮动锚点定位参数、分节、显式分页符和`NUMPAGES`字段缓存的解释不同。原V1与LibreOffice重存稿均含16个`wp:anchor`和15个`wp:inline`，但重存会改写对象定位和分页结构：原V1为2个显式分页符，重存稿为2个`nextPage`分节加3个显式分页符，重叠明显消失但产生冗余空白页。
- 最终模板基于LibreOffice重存稿，移除“特别说明”后及“一级维度测评结果分析”前的两个冗余显式分页符，将页脚从`PAGE / NUMPAGES`改为仅`PAGE`，并将十个维度的浮动标题/定义组合图形改为普通流式标题和定义段落。这样维度标题、定义、得分与诊断按正文顺序分页，不再依赖绝对锚点。
- staging部署前完整备份位于`/opt/talent-assessment/backups/fb125_pdf_layout_20260817_173005`，数据库gzip为12,229,020 bytes、SHA-256=`f719fde5d930ea5e85eed7f28f05076366d97d70760bfc3c95b2835f3b89c347`，同时保留原V1模板SHA=`3b6a83fd...`和首版候选SHA=`2ef8572c...`，目录/文件权限为0700/0600。
- staging最终生效模板与工作区正式模板SHA-256均为`42866f2768bf35115831ce0c24deb7aeff0f14a62fb441305e9dffe0faaca297`、452,788 bytes。模板API实证为schema-v1、49内容控件、12图表、0可见占位符、valid=true，下载SHA一致且保持no-store/no-cache。
- 使用完整一期paper=`93a26b3b-047d-4ad9-890a-3b4fe6d042e4`强制重生成报告成功：LibreOffice 24.2、A4 9页、606,276 bytes、PDF SHA-256=`95fb9b0d24be92306a7ffb70566028dc22c1441c9d2dd419e7cadb1a7318faec`。数据库实例、服务器文件和API下载三方大小/哈希一致；九页均有内容、十维标题10/10、内部业务键/占位符/错误`/13`页数均为0。逐页96DPI总览确认无原雷达说明与卡片重叠、无整页空白、无下一维度标题提前或空色块。
- 本地模板专项36项、Go全量和Go build通过；staging临时LibreOffice目录、短时会话和远端验收文件均为0，talent-assessment/nginx/mysql active，内外health正常，部署窗口关键应用错误和Nginx 5xx均为0。production未修改。Microsoft Word桌面外观仍需用户下载新SHA版本后最终确认。
- [FB-126历史报告补充] 用户指出的`123-658dc083-6216-4373-93b3-a7b1d188ec44-胜任力临时测试报告`实际是2026-08-13生成并持久化的旧PDF；模板替换不会自动重写已完成实例。旧文件已先备份为`fb125_pdf_layout_20260817_173005/123-658dc083-old-before-fb126.pdf`，SHA=`42bc4ae3...`、535448 bytes、A4 12页。仅对该paper执行force重生成后，新报告为LibreOffice 24.2、A4 9页、601397 bytes、SHA=`74c52f8af6322d48968a66fcd4a37f38bc02a4b86eb40139cc4155962b7f20fe`；数据库、服务器文件和API下载三方一致，regenerate成功审计1，九页非空、十维10/10且无原重叠/空白/错序/`/13`。其他历史报告未批量重生成，production未修改。
- [FB-127图表入格补充] 用户截图进一步发现首个逻辑思维环形图仍浮在图示说明区。OOXML实证为chart3虽位于表格单元格中但仍使用`wp:anchor`和段落偏移，chart4–chart12均已是单元格内`wp:inline`。新增`TestBugFB127_Phase1DimensionChartsAreInlineInsideScoreCells`先RED命中chart3，再将chart3转换为原得分单元格内的inline；最终模板SHA=`37caebca5ae3b0cf5b986cf1f797e0d7be16bab6189c7154a5c24f3c322aa869`、452773 bytes，部署前版本已备份为`competency-phase1-report-before-fb127.docx`。
- FB-127使用同一用户paper force重生成后，报告为LibreOffice 24.2、A4 9页、601416 bytes、SHA=`948eb7ca30a1bfb491e7be6c2786cc845d515165d7d4f19b02d9944081b69c87`；实例/文件/下载三方一致，累计regenerate审计2。120DPI截图确认逻辑思维3.50图已进入左侧得分单元格，数字应用、计划执行、持续学习等图表也都在各自表格内。结构专项、模板专项37项、Go全量和build通过；临时文件/会话0、三服务和内外health正常、关键日志与Nginx 5xx均0。production未修改。
- [FB-127后模板全页审查] 120DPI逐页审查确认当前9个物理页均非空，无新重叠、裁切、图表越格、内部业务键或错误总页数；49内容控件、12图表、0可见占位符和chart3–12全部inline+in-cell保持通过。最后一页内容利用率较低，但客户原样例末页也留白，是否强制恢复10个物理页属于版式选择，未擅自增加分页。
- 审查发现三个尚未修正的明确问题：①一级得分汇总行仍是静态`3.75/3.70`，而该paper数据库、图表和分析区为`3.50/3.60`；②chart1仍为`pie3DChart`，把两个独立1–5分误画为构成比，违反2026-08-10已确认口径；③该测评requiredFields仅`name,telephone`，模板仍显示年龄/性别/单位/岗位标签并留空。分别登记FB-128/129/130，当前未修改模板或staging，需按独立RED→GREEN切片处理。
- [FB-128完成] 一级汇总行的静态`3.75/3.70`已替换为可重复`group.general_ability.score`和`group.psychological_quality.score`内容控件；两个Tag均由1处增为2处，运行时同步填充汇总行和分析区。最终模板SHA=`70bcf95368b3387c7090a7ba848971a614723e24f0bf62c8f9f4baff7abcdf2b`、452806 bytes，部署前FB-127模板已备份为`competency-phase1-report-before-fb128.docx`。
- 用户paper=`658dc083-6216-4373-93b3-a7b1d188ec44`重生成后，一级汇总行、图表标签和分析区均为实际`3.50/3.60`；PDF为LibreOffice 24.2、A4 9页、601414 bytes、SHA=`dc144195ddaf7b0367d74245d72bdb3d44494c9f664c0bbc7525ddad90079c39`，实例/文件/下载三方一致。RED→GREEN、相邻FB-127/V1契约、LibreOffice转换、Go全量和build通过；临时文件/会话0、三服务和内外health正常、关键日志/Nginx 5xx为0。production未修改。FB-129/130仍保持RED，未与本切片混改。
- [FB-129完成] 一级图表由误导性的`pie3DChart`改为横向clustered bar：两个分值使用共同0–5值轴、主刻度1、两位小数标签、独立绿色/青色条，不显示百分比和冗余图例。最终模板SHA=`6f43de7c05bf5335a891ad6c1af694b9599954e48acb241fe1a131f506de3736`、452965 bytes，部署前FB-128模板已备份为`competency-phase1-report-before-fb129.docx`。
- 用户paper再次重生成后，图表显示通用能力3.50、心理素养3.60的独立0–5长度，汇总行和分析区保持一致；PDF为LibreOffice 24.2、A4 9页、564100 bytes、SHA=`3e25629470cf7220eeb78eae101e39ef2f9b09a2dc7ee4dfd72259083ce3b49b`，实例/文件/下载三方一致，累计regenerate审计4。FB-127～129、模板契约39项、Go全量和build通过；临时文件/会话0、服务/health/日志通过。production未修改。FB-130仍为下一独立切片。
- [FB-130本地GREEN / staging阻塞] `phase1WordPayload.Meta`已接入`requiredFields`。Word填充前按六个稳定字段键裁剪个人信息表：删除未配置字段的单元格或整行，单字段行跨两列，姓名/手机号及固定时间/时长保留；空requiredFields继续保留六项兼容行为。RED命中年龄标签残留，GREEN后用户同款`name,telephone`本地LibreOffice报告为A4 9页，只显示姓名、手机号、时间和时长，无年龄/性别/单位/岗位。
- FB-130本地模板/handler专项40项、Go全量通过；Linux后端已构建为48,340,473 bytes、SHA=`1099a2aa37f41033c7991d054ee0efeeb66aaaf0ace57d235c1c29a8becce2fb`，模板继续使用FB-129 SHA=`6f43de7c...`。部署前SSH以ConnectTimeout 10/20秒连续三次超时，TCP/22不可达，但公网health仍ok。遵守“先备份再部署”，数据库备份命令未启动、后端未上传/替换、报告未重生成；production未修改。SSH恢复后顺序固定为：完整备份→部署后端（模板无需变）→force重生成用户paper→验证仅姓名/手机号及三方哈希→清理/health/日志。
- [UF-017 / FB-131 Word打不开最终根因纠正] 当前FB-129模板ZIP、45个部件和全部XML都完整，LibreOffice可转换。初步发现chart1新增2个Schema错误，但修成Office 2019 SDK零错误后，本机Word 16仍拒绝，证明SDK零错误不是Word可打开的充分条件。逐版Word COM二分实证：Git HEAD V1 SHA=`3b6a83fd...`可打开；LibreOffice重存基线SHA=`a341e3a3...`首次失败，之后FB-125～131所有基于该稿的候选均失败。LO重存相对Word基线改写39个部件、新增6个、删除51个，是全包兼容损坏根因。
- 最终方案放弃LO重存稿，从用户曾确认可打开的Git V1基线出发，仅做定向变换：删除冗余分页、10个维度标题/定义改流式段落、chart3入格、一级汇总动态控件、chart1独立0–5条形图、footer4仅PAGE。最终包相对基线仅`word/document.xml`和`word/charts/chart1.xml`内容变化，无部件增删；模板SHA=`0899d49768cf1463ecb97544d2ef7d11332ef6ae42133a10e89501d9e7a33fac`、531586 bytes。
- 本机Microsoft Word 16.0 build 19127真实打开最终空模板与运行时填充DOCX成功：Word物理8页、9表、16 inline、5 shapes，关闭不保存后SHA均不变；从staging API下载的同SHA模板也真实打开成功。Office 2019 SDK仍有24项原V1兼容提示，但chart1新增错误0；这些提示不阻止Word，真实COM结果优先。LibreOffice填充报告继续为A4 9页、仅PAGE页码、无错误总页数。
- staging备份=`/opt/talent-assessment/backups/fb130_131_20260818_092103`，数据库gzip 12230193 bytes/SHA=`1ce92faf...`，旧后端=`515801f1...`、旧坏模板=`6f43de7c...`。已部署FB-130后端SHA=`1099a2aa...`及FB-131模板SHA=`0899d497...`。用户paper重生成PDF为529952 bytes、SHA=`fed9ccb0a440e9bff414a5f91b75f189721bbe0edfbfbcb6f3d173fd9aad4c15`、A4 9页；仅显示姓名/手机号/时间/时长，一级独立图和二级图表正常，实例/文件/下载一致，regenerate审计5。临时文件/会话/LO工作区0，三服务和内外health正常，关键日志/Nginx 5xx为0。production未修改。
- [UF-018 / FB-132附件修复与系统替换] 用户第二次提交的同名附件实际SHA=`36f3fe9472590b3ff4d4c3c5f1504eb2d894f2f9f371798054e63048e098ff96`、542140 bytes，仍不是官方`0899d497...`。Word 16返回“文件可能已经损坏”；上传门禁拒绝重复`dimension.competency-a1-05.diagnosis`；附件为52控件/49唯一字段/5表，chart1回退pie3D。与官方文本差异主要是内容控件内示例诊断文案，运行时生成时会被正式冻结文案覆盖，不需要迁移；附件旧版布局和图表反而会回退FB-125～131修复。
- 原附件已备份为`tmp/user-submitted-template-36f3fe94.docx`，随后用官方Word原生模板修复同名`docs/competency-phase1-report-0899d497.docx`。修复文件SHA=`0899d49768cf1463ecb97544d2ef7d11332ef6ae42133a10e89501d9e7a33fac`、531586 bytes；Word真实打开8页/9表/16 inline/5 shapes且无写回，上传契约51控件/49字段/12图表/0占位符通过，LibreOffice填充为A4 9页。
- staging替换备份=`/opt/talent-assessment/backups/template_user_repair_20260818_104725`，备份和生效SHA均=`0899d497...`（等内容原子替换）。模板API返回51/49/12/0、valid=true、no-store及正确MIME；下载回本机Word 16实开成功。用户paper强制重生成PDF 529952 bytes、SHA=`ba768e17ae309fdb5959441f68be98c0711e2f887dc807be028965242c53946a`、A4 9页，实例/文件/下载一致。临时文件/会话/LO工作区0，三服务和内外health正常，关键日志/Nginx 5xx为0。production未修改。
- [UF-019 / FB-133模板审查] 新附件`docs/胜任力测评报告模板.docx`实际SHA=`b606aedc39c12a7162df25036a65951e7acbceb2a0b728dfec9f951e9567e3bc`、542105 bytes，不是系统模板。Word 16实开提示“文件可能已经损坏”；Office 2019校验24项WPS图表扩展错误；服务端上传门禁拒绝重复`dimension.competency-a1-05.diagnosis`。结构为52控件/49唯一字段/5表/15 inline，chart1回退pie3D；相对官方改20部件并删除footer2-4/header3-4。
- 附件的10个诊断控件填入长示例文本，其中沟通表达被复制为2个同Tag；这些示例在系统运行时会被数据库正式文案覆盖，不能作为迁移附件的理由。LibreOffice可宽容生成A4 10页，但与系统9页门禁不符：第4页仅半个一级分析框且大面积空白，第6-10页回到旧式布局，并显示内部`competency-a1-0x诊断`文字。结论为不合格，未修改工作区正式模板、未上传/替换staging、未重生成报告；系统仍保持Word验证通过的`0899d497...`。
- [FB-134仅修Tag] 按用户“只修复Tag，不修改文字和布局”要求，先备份`docs/胜任力测评报告模板.before-tag-fix.docx`，再解除第二个重复`dimension.competency-a1-05.diagnosis`内容控件但原样保留其`sdtContent`。修复文件为`docs/胜任力测评报告模板.docx`，SHA=`f0b49aff0d2c881fce27b564aa82fbc482a5dddc2ec0a0a9609c1a63e1355dba`、534729 bytes。
- 修复前后只有`word/document.xml`变化，ZIP部件无增删；可见文字SHA完全一致，文本长度3398、段落141、表格5、行42、单元格61、绘图20均不变。内容控件52→51、唯一字段49，合法重复仅两个一级得分字段；服务端上传门禁通过。本机Word 16正常打开且关闭后SHA不变。该附件仍保留其旧5表/旧pie3D/10页布局及第二份静态沟通表达诊断文字，因为用户明确禁止改文字和布局；本次未替换工作区正式模板、未部署staging，系统仍为`0899d497...`。
- [客户模板进一步审查] 用户确认业务内容由Tag替换且当前只作候选后，对Tag修复稿做运行时填充、图表/关系、分页和页脚深审。49个唯一业务字段齐全且顺序整体正确；chart3-12均在表格单元格内，页脚仅PAGE。仍有四类结构问题：①`dimension.competency-a1-04.diagnosis`把静态`【诊断】`包在控件内，运行时替换后持续学习模块丢失该标签（实测PDF确认）；②chart1仍为旧3D饼图；③11个图表关系指向原作者本机`C:\Users\...\数据图表.xlsx`，PDF依赖缓存可生成，但Word编辑/刷新不可移植并泄露本机路径；④4个显式分页符+2个nextPage分节导致Linux A4 10页，第4页明显稀疏。另有24项WPS图表扩展Schema提示，但Tag修复后Word 16已实开，属于版本兼容风险而非当前打开阻塞。该候选未部署，系统模板仍为`0899d497...`。
- [FB-135持续学习Tag边界本地完成] 备份=`docs/胜任力测评报告模板.before-boundary-fix.20260818_113235.docx`，SHA=`f0b49aff...`。将静态`【诊断】`从`dimension.competency-a1-04.diagnosis`控件内移到同位置控件前，并将Alias从错误的a1-03同步为a1-04；不改其他Tag。修复后候选SHA=`a85373277889117883e8b6c15d1ed0bd7cf5839d858dfa41389453b884ae8a05`、534733 bytes。
- 边界修复前后可见文字SHA完全一致，3398字符、141段落、5表、42行、61单元格、20绘图均不变。新增Python回归先RED后GREEN；Word 16打开且无写回；服务端上传门禁及重复Tag负向门禁通过；真实LibreOffice填充PDF显示`【诊断】competency-a1-04 诊断`。本切片未处理11个外部Excel关系或一级3D饼图，等待客户选择；未替换系统模板、未部署staging。
- [FB-136客户选择与外链清理] 用户选择“清除外链”而非内嵌Excel，并确认保留客户3D一级饼图。清理前备份=`docs/胜任力测评报告模板.before-external-link-fix.20260818_113921.docx`，SHA=`a8537327...`。删除chart2–12共11条TargetMode=External关系及11个对应`c:externalData`节点；不改图表缓存、样式、得分点、正文或布局。
- 当前客户候选`docs/胜任力测评报告模板.docx` SHA=`434495c26dbb7ea28662168384467139fff522d8412a5131bc96ff82825496d7`、533117 bytes。外链测试RED→GREEN；修复前后可见文字SHA、141段落、5表、20绘图完全一致；Word 16打开且无写回，服务端上传契约通过，LibreOffice缓存图表仍生成A4 10页，持续学习运行时标签正确。一级3D饼图按客户选择保留，不再作为本轮待修项；候选仍未部署，系统模板保持`0899d497...`。
- [纠正 - 2026-08-18 全新分页复核] 先前把客户候选“第4页内容较少/明显稀疏”列为问题不准确，用户明确确认不存在。基于当前SHA=`434495c2...`重新从零填充并以120DPI渲染10页：全部物理页非空；物理第4页含一级得分、图表及通用/心理素养分析，物理第5页为心理素养分析延续，属于客户10页编排，不再列缺陷。旧结论保留在历史记录中，本条为正式纠正。
- 全新复核仍发现两个实际跨页点：逻辑思维标题/定义在物理第6页末，得分与诊断在第7页开头；自律性标题/定义/得分在第9页，诊断在第10页开头。其他八个维度模块的标题、定义、得分和诊断均在同一物理页。是否修复这两个跨页需客户确认；可通过局部keep-with-next或表格行分页控制处理，不应再以“第4页较少”为理由重排整份模板。
- [FB-137全部维度模块分页优化完成] 用户要求完成优化后，先备份`docs/胜任力测评报告模板.before-page-flow-fix.20260818_115714.docx`，SHA=`0af9370b...`。首轮只处理逻辑思维/自律性虽解决Linux两处跨页，但Word端仍可能拆其他模块；最终将门禁扩展到全部10维度：每个模块的标题/定义、得分、诊断三行均`cantSplit`，标题行和得分行的单元格直接段落均`keepNext`，不修改组合图内部段落。
- 最终客户候选`docs/胜任力测评报告模板.docx` SHA=`ba98522523235bd66fc47cdd553a347a0c7ddd5a426541ecfc5f50ca2f410884`。优化前后可见文字SHA一致，3398字符、142段落、5表、42行、61单元格、20绘图均不变；Tag边界和零外链测试保持GREEN，Word 16打开且无写回，上传契约通过。LibreOffice为A4 10物理页，逻辑思维完整移到第7页、自律性完整移到第10页，十个模块均不跨页；Word导出为11物理页，合作意识完整位于最后一页，十个模块同样不拆分。Word/LO页数差异来自排版引擎，不是空白页或内容丢失。候选尚未替换系统模板、未部署staging。
- [staging发布阻塞 - 2026-08-18] 用户明确要求部署并测试。最终关闭Word后候选SHA再次确认=`ba98522523235bd66fc47cdd553a347a0c7ddd5a426541ecfc5f50ca2f410884`、533187 bytes；Tag边界、零外链、全部10维模块同页、上传契约、Word 16实开（11页且无写回）和LibreOffice A4 10页门禁全部通过。部署前数据库/模板备份SSH以ConnectTimeout=20连续两次超时，第三次只读握手仍超时；TCP/22不可达但公网health正常。遵守先备份纪律，备份命令未启动、模板未上传/替换、真实报告未重生成，staging仍使用`0899d497...`，production未修改。SSH恢复后继续：备份→原子替换`ba985225...`→模板API/Word下载实开→force重生成用户paper→10页逐页/三方SHA/清理终验。
- [FB-138客户模板staging分页兼容完成 - 2026-08-18] SSH恢复后首次部署全模块候选`ba985225...`，真实用户paper在staging LibreOffice 24.2生成12页：物理第5页仅有一级“心理素养”说明尾行，下一显式分页又强制二级结果从第6页开始。该结果未通过门禁，模板立即回滚至`0899d497...`，用户报告恢复9页；production未修改。
- 三份相同测试数据的无控制/两模块/全模块DOCX在staging 24.2均为9页，证明固定短夹具不能替代真实长文案验收。新增RED回归精确拒绝一级分析表与“二级维度测评结果及建议”之间的叠加显式分页；最终仅删除该空分页段落，保留10模块全部`cantSplit + keepNext`、可见文字、Tag、图表和其余客户分页结构。
- 最终客户模板SHA-256=`9bf1cbb77b2cb6f23d37d878fb2d0e9da7664f296a90b2f50cc8c62a423c216d`、533147 bytes。Tag边界、零外链、全模块流式结构测试通过；本机Word 16与staging API下载稿均真实打开为11页、5表、14 inline、6 shapes，关闭后SHA不变。模板API为51控件/49字段/12图表/0占位符、valid=true、no-store和正确DOCX MIME。
- staging发布前备份=`/opt/talent-assessment/backups/customer_template_sparse_fix_20260818_123253`，数据库gzip 12231339 bytes/SHA=`c9a83d7e62583491b3d94e9bf6cb5a37f255988c06c0764375b12803fc4444d1`，旧模板SHA=`0899d497...`。真实paper=`658dc083-6216-4373-93b3-a7b1d188ec44`强制重生成后为LibreOffice 24.2、A4 11页、500691 bytes、PDF SHA=`8ae5a38662efc849d33c9722829707e33ca8b960edb74352a78d50065855e2f1`；数据库、服务器文件和下载副本三方一致。120DPI逐页总览确认11页均有内容、无重叠/裁切/整页空白，10个维度模块均未拆页。短时会话、LO profile和客户模板远端临时文件均为0，三服务及内外health正常，关键错误0；production未修改。
- [FB-139环形图分值居中完成 - 2026-08-18] 用户截图指出环形图内数字未居中。chart3–12可见标签均保留手工`x/y`和Office 2013扩展`w/h`，但十图的标签框中心`(x+w/2,y+h/2)`均未统一落在图心；chart3首个RED水平偏移`-0.005068372027874496`。新增结构回归后，只将每图可见标签改为`x=-w/2,y=-h/2`，不改图表缓存、颜色、数据、字体、正文或分页。
- 最终模板SHA-256=`4bcf5aceb62b0424af5f89dd315f8575725f615ee1f1d8a827dc943a73b923fe`、533152 bytes。十图标签居中、Tag边界、零外链和模块分页四项门禁通过；本机正式模板和staging API下载稿均由Word 16真实打开为11页/5表/14 inline/6 shapes，关闭后SHA不变。模板API继续为51/49/12/0、valid=true、no-store和正确MIME。
- staging部署前备份=`/opt/talent-assessment/backups/customer_chart_label_center_20260818_125701`，数据库gzip 12231411 bytes/SHA=`6e4c2172b014470f62f8603ff87778b8ab0657ce5a43b03fbaef9b7dfd9e28b9`，旧模板SHA=`9bf1cbb7...`。同一真实paper重生成后为LibreOffice 24.2、A4 11页、500686 bytes、PDF SHA=`11aed2b94619ca8469de8d33104aff98a74543a40180d7c4160ffb37df2e29f8`；数据库/服务器/下载三方一致。180DPI逐页总览确认10个环形图内分值均水平、垂直居中且无分页/重叠回归。production未修改。
- [纠正 / UF-021 / FB-140 - 2026-08-18] 上述FB-139“10图均居中”结论有误。原因是仅验证OOXML标签框`x=-w/2,y=-h/2`并人工查看缩略总览，没有测量LibreOffice输出像素。用户复核指出多个图仍偏移后，新增真实PDF像素门禁：180 DPI自动识别10个圆环和内部数字字形框，要求水平/垂直中心差均不超过2px。旧PDF十图全部失败，偏移范围为x=`0～15px`、y=`-3.5～20.5px`。
- 正确方案不再修改客户Word模板（Word中本来居中），而是在后端仅对LibreOffice转换前的已填充DOCX应用chart3–12独立校准；Graph和Word模板不受影响。校准值来自同一真实报告的像素反馈，并通过三轮收敛，最终真实11页PDF十图均通过2px强制门禁。Go聚焦、全量测试和Windows/Linux构建通过。
- staging部署前备份=`/opt/talent-assessment/backups/fb140_pdf_chart_center_20260818_131740`，数据库gzip 12231459 bytes/SHA=`d48ee559b46297ae612e72b9c84386cf6a136c4ffb694e82e702a7e6e133d339`，旧后端SHA=`1099a2aa...`，模板SHA=`4bcf5ace...`。最终后端SHA=`bb932884ac09d9d3c163ceb7f573c7cbbdc6d2d78262f43b59692073656fc74f`；真实paper生成LibreOffice 24.2 A4 11页、500688 bytes、PDF SHA=`70821dd0ca2f5ae4d80d83f8fa474c2621d40639b3f2e1d2086b3787b7e9981f`，数据库/服务器/下载三方一致。180 DPI像素测试10/10通过，逐页总览无分页或重叠回归；远端临时文件、LO profile和短时会话均0，三服务、内外health和关键日志通过。production未修改。
- [最新客户模板重新校验 / FB-147～149 - 2026-08-18] 用户重新保存的模板SHA=`e85ccec5...`、542612 bytes。RED门禁发现三个回归：一级/二级区之间的显式分页再次出现，chart7由单元格内inline退回anchor，chart3–12缺少显式`0.00`标签格式。两个一级得分Tag仍各有2个，汇总表并未丢失动态控件。
- 最小修复仅删除该叠加分页、将chart7恢复inline，并给chart3–12可见标签增加`0.00`；可见文字3398字符及SHA保持不变。最终模板SHA=`50b238cc30d0480d534fb2291842248b9306f0878fb095f8008236aeaa894cd6`、535238 bytes，Word 16真实打开10页/5表/14 inline/6 shapes且无写回；Tag边界、零外链、模块流、动态一级表格、两位小数和上传契约51/49/12/0均通过。
- staging备份=`/opt/talent-assessment/backups/customer_template_jump_20260818_151823`，数据库gzip 12231731 bytes/SHA=`247be7cc...`，旧后端=`bb932884...`、旧模板=`4bcf5ace...`。模板与首轮后端已部署；真实paper=`658dc083-6216-4373-93b3-a7b1d188ec44`生成LibreOffice 24.2 A4 10页、499767 bytes、SHA=`ff14d0dd...`，实例/文件/下载三方一致。一级汇总表为真实`3.50/3.60`，分析区一致，静态`3.75/3.70`计数0，图表及正文分值均为两位小数。
- 首轮真实PDF像素门禁仍失败，证明新模板保存后旧FB-140校准值不能复用；基于真实长文案像素反馈生成第二轮后端SHA=`9cdfa3f1c91ee81d3765a0f5c63145fdeb8276d04008c94b5b520b7b26da85b1`并部署。SSH恢复后同一paper重生成LibreOffice 24.2 A4 10页、499772 bytes、PDF SHA=`37e115246d366b03276b7ec88d32df172c837434371364b1a2973cda3018c5e3`；数据库实例、服务器文件和API下载三方一致，180 DPI十图像素门禁10/10通过。十页逐页总览无整页空白、重叠、裁切或模块错序；短时会话、任务临时文件和LO profile均0，三服务active、内外health正常、关键应用错误和Nginx 5xx均0。production未修改。

## 2026-08-17 staging统一发布与回归

- 用户确认仅发布staging（`20.200.136.133`），production未修改。发布前Go全量、go vet、Linux build、前端24文件153项及production build通过；前端仅2个既有资源体积warning。
- 发布前完整备份位于`/opt/talent-assessment/backups/release_verify_20260817_20260817_162021`，目录0700、文件0600；数据库gzip通过完整性检查，SHA-256=`d6572b07f4743171b31c18f4b3dd91cf0ea094e45bc998f1260f7a0511e48ff3`，同时备份后端、前端和V1稳定模板。
- staging生效版本：后端SHA-256=`515801f10a183d053edf9acab06cf2a9eb477463ee6316b9b18a875cb339d370`，前端index SHA-256=`4be781a62ee7895b2b610ff897b9424df52dfc03235ce1e245ae603d6e8edded`，V1 Word模板SHA-256=`3b6a83fd4a2fddf7c0a47c1eda5e2e4141b7d0d72fd9431980928be598e86b92`；公网index哈希一致，内外health正常。
- staging真实测试通过：模板元数据/下载为schema-v1、no-store、正确SHA；完整一期答卷强制重生成LibreOffice 24.2 A4 12页PDF，API/数据库size与SHA一致；结果页面列表`34.13`、10维详情和10个报告分值均为两位小数；真实封闭测评新增空身份证人员按手机号识别、默认密码后4位并清理0；传统00101/00201/00301各完成2题组卷→答题→交卷→结果→标准分，整链清理0。
- 终验talent-assessment/nginx/mysql均active，发布与浏览器短时会话0、LibreOffice临时工作区0、临时发布文件0、传统烟测和人员验收数据0、应用关键错误0、Nginx最近5xx=0。production未部署。
- 用户再次提供旧错误截图后，通过公网`http://20.200.136.133/prod-api/exam/api/tester`重新实测：真实封闭测评`1786520226890178516`、身份证空、手机号非空新增返回code=200，身份证保持空、默认密码为手机号后4位；测试行和短时会话清理为0。截图中的“缺少 idNumber 或 examId”不再存在于当前代码路径。

## 2026-08-13 FB-123/124 测评管理分值与封闭人员新增 staging修复

- FB-123将胜任力管理列表的整体分、所选维度分、评价均值，以及详情中的整体分、一级得分、得分合计、维度分统一显示两位小数；空值显示`—`，不改变数据库DECIMAL精度，也不把题数、逐题原始值和逐题计分值改成小数。Vue/Chromium报告原有`format(...).toFixed(2)`保持，并补齐报告“得分合计”的格式化；Word报告继续使用`StringFixed(2)`。
- FB-124统一手工新增与Excel导入契约：封闭测评人员身份证号可空；身份证存在时按`id_number+exam_id`识别，身份证为空时按`telephone+exam_id`识别；默认密码优先取手机号后4位。手机号与身份证同时为空仍在数据库写入前拒绝。
- 本地RED→GREEN后Go全量、go vet、Linux build、前端24文件153项和production build通过。staging备份位于`/opt/talent-assessment/backups/fb123_124_20260813_235807`，数据库gzip SHA-256=`4309492e0627381efd8592e9d52ec122f95fc052eb4fb55e8e21201470ba1104`。
- staging后端SHA-256=`03551509cd8a85c7d99c2ef16b2e3630a2cdc0ee2da53a13b71ab0471c324a81`，前端index SHA-256=`8e9cf6661e84c2004e2c881ac278cc6707520e56930410f02bcaff07d85e01a3`。真实封闭测评`1786520226890178516`新增空身份证人员成功，手机号/默认密码核对通过且清理为0；真实一期结果浏览器显示列表整体分`34.13`、10维详情得分和10个报告得分全部两位小数。production未修改。
- 公网index与远端哈希一致，公网health正常；talent-assessment/nginx/mysql均active，FB-123/124短时会话0、验收人员0、部署临时文件0、关键日志0。

## 2026-08-13 一期Word模板V2透明业务键（本地）

- 在不修改数据库结构的前提下新增代码字段注册表：75个稳定业务键，其中49个必需、26个可选；当前人员/结果/正式文案DTO作为数据源，得分百分比、满分和距满分差值在运行时派生。可重复字段允许在Word多处使用，不可重复字段仍由上传门禁拒绝重复。
- 12个图表增加稳定业务键：`chart.group.overview`、`chart.dimension.radar`及10个`chart.dimension.<维度ID>`；业务键保存在Word图表对象替代文字标题，运行时通过document关系解析实际图表部件，不再依赖`chart1.xml`～`chart12.xml`物理编号。V1无业务键模板继续使用旧物理映射。
- V2内嵌Excel仅含`FieldDictionary`和`ChartData`：前者自描述75个字段与schema版本，后者用稳定业务键承载2个一级和10个二级图表数据；运行时同步更新ChartData与图表缓存。上传门禁校验业务字段、12图表键、公式区域、package关系、1个内嵌工作簿和0外链，并向管理页面返回透明契约元数据。
- 确定性生成器连续运行SHA一致；候选模板SHA-256=`0f1a23a895df3417bf9a1e939ab101728c27ae26b0a7a6961234e47da65539fd`。V1/V2聚焦、负向矩阵、Go全量、Windows build、前端全量/production build通过；本机LibreOffice真实转换为A4 12页。当前仅本地完成，未切换staging，production未修改。

## 2026-08-13 一期Word模板V2 staging部署完成

- staging部署前完整备份位于`/opt/talent-assessment/backups/phase1_v2_contract_20260813_230941`，目录0700、文件0600；数据库`element.sql.gz`通过`gzip -t`且SHA-256=`953cdb431fef76f43b801c1c360ba0be64333c72fed5402b4d240256bc40fc20`，同时备份旧后端、旧前端和旧模板。旧前端即时回滚目录保留为`/opt/talent-assessment/dist.pre-phase1-v2-20260813_230941`。
- staging生效后端SHA-256=`a0f96c8fc947b2ced89e9b86122098ff0b3ec5fce5952a84f374beb73a9bc231`，前端index SHA-256=`2bce116eb5357872c4315d40ba05aee06cf2bfbe22835f222fb72e777a1c5055`，V2模板SHA-256=`0f1a23a895df3417bf9a1e939ab101728c27ae26b0a7a6961234e47da65539fd`；公网index与远端原始字节哈希一致，公网health正常。
- 模板管理真实API返回`schema-v2 / registeredFields=75 / usedFields=49 / businessCharts=12 / embeddedWorkbooks=1 / externalLinks=0 / visibleTokens=0 / valid=true`。远端DOCX独立结构审计为1个内嵌工作簿、12个唯一业务图表标题、12个package关系和0外链。
- 使用完整一期paper=`93a26b3b-047d-4ad9-890a-3b4fe6d042e4`强制重生成报告成功；reportId=`60f14a12-bef8-42d3-a22a-831ea260b2c2`，LibreOffice 24.2、A4 12页、539250 bytes、PDF SHA-256=`da7c7e9a92985be04f784d81bb1d0849165cfa05f3f362228e07bfaaf7801a68`。人员文本和未解析占位符检查通过，API下载与数据库size/SHA一致。
- 模板管理页面使用的三个真实接口已复验：元数据返回V2有效契约；下载返回正确DOCX MIME、562496 bytes且SHA-256与生效模板一致；上传同一合法V2模板后备份数2→3、返回备份名`competency-phase1-report.docx.20260813_231848_000.bak`并保持生效SHA不变。前端专项4/4覆盖页面加载、V2契约展示、下载文件名、上传确认/刷新和失败保留文件。
- 终验talent-assessment/nginx/mysql均active，运行时临时工作区0、V2短时Redis会话0、部署窗口关键journal错误0、Nginx最近5xx=0；上传、验收和本地发布临时文件已清理。production未修改。

## 2026-08-13 FB-122 Microsoft Word模板兼容修复

- 用户真实Microsoft Word打开staging下载的旧V2模板（562496 bytes、SHA-256=`0f1a23a895df3417bf9a1e939ab101728c27ae26b0a7a6961234e47da65539fd`）时报“Word在试图打开文件时遇到错误”。下载接口、ZIP和LibreOffice均正常，说明原验收遗漏桌面Word的严格OPC检查。
- 根因已定位：生成器把`<Default Extension="xlsx">`追加到`[Content_Types].xml`末尾，位于全部`<Override>`之后；实际位置lastDefault=8448、firstOverride=389，违反OPC要求的Default先于Override。LibreOffice容忍该结构，Microsoft Word拒绝整个DOCX。
- FB-122先RED后GREEN：内嵌工作簿改为显式`/word/embeddings/competency-phase1-chart-data.xlsx` Override；上传门禁新增Content Types顺序、显式Override和禁止全局xlsx Default检查。修复后lastDefault=333、firstOverride=389。
- 修复已部署staging：后端SHA-256=`5b73b91dda1af522987d05b71d04d7c60b70144c54b7f24782bf3a002159afec`，模板562521 bytes、SHA-256=`a2387516f20c18037dca84b3e17cd7eb04ce60640004d8d88c0118f8912b0793`；备份位于`/opt/talent-assessment/backups/fb122_word_compat_20260813_234503`。真实下载与结构门禁通过，完整一期报告仍由LibreOffice 24.2生成A4 12页。当前执行环境未安装Microsoft Word COM，最终桌面Word打开结果需用户删除旧下载并重新下载后确认；production未修改。
- [纠正 - 2026-08-14] 上述首次FB-122修复无效，用户确认SHA=`a2387516...`仍无法由Microsoft Word打开。Open XML SDK显示V1/V2均有源模板遗留的202条兼容性警告，不能据此隔离；LibreOffice重存会大幅改写图表公式/部件，也不能直接作为可维护模板。V2内嵌Excel方案现已暂停，不再宣称staging可用。
- staging已回退为此前经客户Microsoft Word编辑并多次上传成功的V1内容控件模板：557442 bytes、SHA-256=`3b6a83fd4a2fddf7c0a47c1eda5e2e4141b7d0d72fd9431980928be598e86b92`，schema-v1、49控件、12图表、校验有效；完整一期报告仍生成A4 12页。坏V2备份位于`/opt/talent-assessment/backups/fb122_v2_suspended_20260814_001235`。
- 为防止浏览器继续返回或用户误开同名旧坏文件，下载接口增加`Cache-Control: no-store, no-cache, must-revalidate`、`Pragma: no-cache`；页面下载文件名包含SHA前8位，当前应为`competency-phase1-report-3b6a83fd.docx`。缓存修复后端SHA-256=`515801f10a183d053edf9acab06cf2a9eb477463ee6316b9b18a875cb339d370`，前端index SHA-256=`4be781a62ee7895b2b610ff897b9424df52dfc03235ce1e245ae603d6e8edded`。production未修改。
- [用户确认 - 2026-08-14] 从staging重新下载的`competency-phase1-report-3b6a83fd.docx`已可由Microsoft Word正常打开，V1回退闭环完成。V2内嵌Excel继续暂停，不得重新部署或向客户交付；后续优化以当前V1模板为稳定基线。

## 2026-08-11 客户 V1 题本与报告模板复核

- 新版客户题本 `260810基层员工胜任力测评题本+等级评价+总体评价V1.xlsx` SHA-256=`f33b878e6fa3f3b8496a838c1a8e648dda29b1e75a41a21a73e90362a978b42f`。实读为3个Sheet、10个A/B维度、80道维度题、10道效度题；维度题62正向/18反向，效度题10道正向；90个题号和题干均唯一。题目、方向、10×5维度文案和5条总体文案与既有候选JSON逐字段一致。
- 新版题本把维度单元格从名称改为“编号+换行+名称”，当前转换器按整个单元格匹配名称，真实执行报`unknown phase-1 dimension: A1-01 / 逻辑思维`；正式导入前必须单独修复并回归转换契约。效度题的“题目类型”列仍为空，必须继续按效度量表层级明确识别，不能当普通维度题。
- `胜任力测评报告样例.docx` SHA-256=`71648879883952f2199df0ec1f973bcd462d7bb9674f0960b4a6c11da7c4d053`；随附PDF实证为A4 10页。结构为封面、阅读说明、个人/总体/效度提示、一级维度、十维概览、10个二级维度诊断建议。Word可提供2个一级说明、效度存疑提示和特别说明，但没有独立的“效度良好”正式提示，最终免责声明/使用边界仍需批准确认。
- 当前一期Vue十页框架与物理页数一致，但“十维能力全景”实际为横向柱状图而非决策基线要求的10轴雷达图；一级页未显示等级；二级详情未展示Excel中的完整定义；封面未显示测评名称。正式内容导入前需作为独立实现切片处理。
- [FB-108本地修复 - 2026-08-12] 转换器现同时支持旧“名称”和V1“编号+换行+名称”维度单元格；新版格式会对A/B编号与名称执行精确配对，错配在生成任何候选/导入产物前确定性拒绝。默认来源切换为260810 V1题本，工具版本升至1.3.0、材料日期升至2026-08-10。RED真实报`unknown phase-1 dimension`；GREEN契约测试、转换器语法检查和既有一期身份/导入回归均通过。本切片未覆盖或重新生成现有候选JSON/XLSX，未写数据库、未部署。
- [规范化CSV人工审阅包 - 2026-08-12] 用户确认采用4个规范化业务CSV、Windows Excel兼容UTF-8 BOM，并在人工修订后替换90题和正式报告文案。转换器1.4.0新增`--csv-output-dir`与`--csv-only`，已从260810 V1题本生成`phase1-questions.csv` 90行、`phase1-dimensions.csv` 10行、`phase1-dimension-levels.csv` 50行、`phase1-overall-levels.csv` 5行，目录为`scripts/data/competency-phase1-csv/`。实际产物BOM/表头/行数/SHA和62正向+18反向+10效度均验证通过；既有候选JSON/XLSX未覆盖。人工检查期间不得修改题目编号、题型、一级/二级维度ID/编号、顺序、等级编号等身份列；正式替换前另行实现CSV导入预览、全量校验、事务替换及备份，不直接把人工CSV写入数据库。
- [CSV长期导入契约优化 - 2026-08-12] 上述4文件审阅包已扩展为6文件`competency-phase1-csv-v1`：新增`phase1-package.csv`承载四版本、双来源SHA、双批准、环境和最终免责声明；新增`phase1-report-static-texts.csv`承载2条一级说明与good/questionable效度提示。题目/维度/文案文件增加稳定顺序、审核状态/备注；维度和总体等级增加内部等级码与精确上下界，包含标志用Excel稳定的1/0。效度good内容保持空白并标`missing_required`，package保持draft，不伪造批准或免责声明。转换器升至1.5.0；新增校验器支持manual-review和approved-import两模式，检查6文件集、BOM、表头、身份、唯一性、题型、精确方向、边界、审核状态、公式注入和正式内容源SHA。当前manual-review通过，内容源SHA=`c5f343938587f5680328749083caf3b3ea62703be9899d7261f253303cd18289`；approved-import按设计因draft失败。本轮未写数据库、未替换现有候选JSON/XLSX、未部署。
- [计分与CSV一致性审计 - 2026-08-12] 核心后端计分与确认后的CSV精确边界一致：单题正向raw/反向6-raw，二级保存8题转换后`score_sum`并除以8，一级为5个二级平均，总体为10个二级平均之和，效度仅累加10道效度题raw。原说明“各维度原始得分”应改称“转换后计分总和”；“效度所有题目”应明确为10道效度题；总体“65%以下尚未胜任”与50%–65%薄弱胜任重叠，确定应为低于50%，当前代码/CSV已按`>=45/>=40/>=32.5/>=25/<25`修正。8题整数分使二级实际可达均分以0.125递增：L1 1–1.625、L2 1.75–2.625、L3 2.75–3.5、L4 3.625–4.25、L5 4.375–5；统一理论边界仍适用于一级0.025步长。定向计分测试18项通过。当前剩余冲突在报告显示：Vue把L1-L5显示为起步/发展/胜任/熟练/卓越，且总体显示较弱/待提升，与CSV二级、一级和总体正式标签不一致，需独立RED→GREEN修复。完整证据见`docs/phase1-scoring-data-consistency-audit-20260812.md`。
- [CSV驱动客户样例报告模板 - 2026-08-12] 用户选择实际系统Vue/PDF模板，缺失效度良好提示和最终免责声明继续保留门禁。新增确定性目录生成器，将已校验CSV的一级/二级名称、一级说明、核心含义、定义、二级/一级/总体正式标签及区间生成到前端运行目录；Vue不再维护这些业务名称和描述。模板按客户样例实现封面、阅读说明/十维矩阵、个人/总体/效度、一级结果、10轴雷达、5页每页2维定义与诊断，共10页。旧起步/发展/胜任/熟练/卓越及较弱/待提升标签已移除；一期后端DTO补齐examTitle/requiredFields/startedAt/userTime/generatedAt。RED为前端2失败和后端meta 6项缺失；GREEN为前端专项11/11、全量23文件142项、后端相关125项、Go全量/Build及production build通过。真实Chromium 1440和390视口均10页/5详情/10卡/2组/1雷达且无溢出、console错误和效度分泄露；移除模板内额外硬编码免责声明并重新构建后，真实PDF为A4 10页、378868 bytes并已清理。当前CSV仍draft，缺失正式内容时报告门禁保持关闭；未部署staging/production。详见`docs/phase1-report-template-implementation-20260812.md`。
- [效度良好提示建议稿 - 2026-08-12] 用户明确要求先生成建议稿，后续由客户修改CSV和报告模板。`phase1-report-static-texts.csv`的good行已写入“本次测评作答效度良好……”建议文案，来源标记`AI建议稿-待客户确认`、状态`pending_review`、备注要求客户及心理测量负责人确认或修改；转换器1.5.1同步维护同一默认建议，防止重新生成时丢失。CSV manual-review通过，新内容源SHA=`7ffe7fa7a2145a6de4fd8d349adbaf19a011c333557cd457494c75c5f478cf9d`，前端目录已重新生成并通过新鲜度检查。package仍为draft，最终免责声明、双批准、环境和正式内容SHA仍为空，approved-import和正式报告门禁没有放开；未写数据库、未部署。
- [客户DOCX报告模板权威口径 - 2026-08-12] 用户明确`胜任力测评报告样例.docx`是客户提供的完整报告样例，不是单纯视觉参考；它同时规定固定格式、固定文本、条件变化文本、人员/时间/分数/等级变量及其展示位置。DOCX内示例日期、人员和分数只作变量位置样例，不得固化。后续客户修改业务内容时维护CSV；修改格式、固定说明、页面顺序或视觉时维护DOCX；程序需分别同步到内容版本和Vue/PDF模板，再执行DOM及10页A4验收。详细分类见`docs/phase1-report-template-contract-20260812.md`。本轮只纠正文档与项目口径，未改业务代码、数据库或部署环境。
- [一期候选数据staging导入与模拟 - 2026-08-12] 用户选择staging、候选测试身份和Go+Vue+测试修复范围。导入前CSV与staging逐字段核对90题和10维一致；完整备份位于`/opt/talent-assessment/backups/phase1_90_import_20260812_113308`，数据库12,116,623 bytes、gzip/清单通过。FB-109修复备份权限0755/0644为0700/0600，并纠正root-owned 0700目录下普通glob失效为`sudo find`。候选导入器以确定性ID和UTF-8 UNHEX事务连续执行两次，最终66条临时非活动文案（5 overall+7 template+2 group+50 dimension+2 validity）及1个draft包；内容SHA=`3060bf06f3f52715c7cf9b05f277e4ccd723a7f571785c5a26918cc98d8dbb42`。真实90题draft门禁链、临时激活的DTO→Chromium→下载及超时/存疑/发布回滚负向链全部完成并清理。完整客户文本引出FB-110英文标题被强制大写、FB-111本地11页、FB-112 Linux服务端16/20页；均RED→GREEN，最终staging DOM为10页/5详情/10卡/2组/1雷达，打印逐页无有效内容溢出，Windows与Linux服务端PDF均A4 10页。最终候选文案恢复`is_temporary=1,status=1`、包恢复draft，测评/结果/报告/审计0，传统签名不变，服务健康，production未修改。证据见`docs/phase1-candidate-staging-import-verification-20260812.md`。
- [一期staging独立E2E复测 - 2026-08-12] 写入前完整备份位于`/opt/talent-assessment/backups/phase1_90_import_20260812_120247`，数据库/后端/前端SHA清单、gzip及0700/0600权限均通过。重新执行真实90题正向链、88/90超时不完整、效度40存疑、发布回滚、桌面/移动浏览器报告和Linux服务端临时批准报告链，全部通过；浏览器PDF为A4 10页、363233 bytes，服务端下载同为10页。终验候选状态=`10|80|10|66|7|66|1|0|0|0|0`，传统三组签名不变，服务均active、短时会话和关键错误为0，临时脚本/PDF已清理，内容包恢复draft，production未修改。
- [一期代码提交与本地产物清理 - 2026-08-12] 已清理本地后端构建目录、运行日志、临时目录、前端dist、Go覆盖率文件以及脚本日志/测试结果/截图；保留node_modules依赖。提交前Go全量测试、Go Build、前端23文件142项及production build通过；暂存差异通过格式和硬编码凭据扫描。用户`.vscode/settings.json`明确排除，未纳入提交；一期代码、迁移、CSV契约、客户材料、测试和验收文档形成提交`feat: complete phase-one competency assessment`，未push。
- [UF-006 / FB-113 staging修复 - 2026-08-12] 真实结果页exam=`1786508394008352244`的完整一期答卷因遗留产品版本判断无法勾选，且公共Axios拦截器把非200业务码降级为字符串`error`。按RED→GREEN移除一期永久禁用，仅允许完整答卷选择；报告生成API显式使用`rejectWithBusinessMessage`保留后端门禁文本，其他请求错误契约不变；批量结果显示首个失败原因。专项18/18、前端全量23文件143项及production build通过。最终staging前端index SHA-256=`eeb891fce85829aeb4b12fb993877e93868f12b2738533e89be118589b66ff54`，备份为`/opt/talent-assessment/dist.bak.fb113_final.20260812_124807`。真实浏览器确认复选框可用、选中后按钮启用、请求命中后端且UI准确显示“一期正式报告内容尚未完成双重批准”；短时Redis会话0，talent-assessment/nginx active，公网health正常。候选内容包仍为draft，未创建正式报告或绕过双批准；production未修改。
- [一期客户测试报告生成 / FB-114 - 2026-08-12] 用户明确选择先在staging直接批准候选内容并生成报告，内容/心理测量批准人均为`Liming`，生效环境`staging`，采用用户确认的最终免责声明。写入前完整备份位于`/opt/talent-assessment/backups/phase1_90_import_20260812_130732`；事务精确更新1个候选包和66条候选文案，终验为approved、两个批准时间非空、两个SHA均64位、66条文案全部正式活动且免责声明一致。首次真实生成发现FB-114：批准免责声明已冻结到DTO/实例但Vue仍显示目录样例特别说明；RED断言失败后改为优先渲染`reportText.disclaimer`，无批准文本时才回退样例。专项11/11、前端全量23文件143项及production build通过；staging index SHA-256=`49c36ae620952906876801eca66613ccd7b7270e059a9f0e65c8b434ff3da394`，前端备份=`/opt/talent-assessment/dist.bak.fb114.20260812_132809`。强制重生成报告ID=`60f14a12-bef8-42d3-a22a-831ea260b2c2`，状态completed、A4 10页、725716 bytes、SHA-256=`6e77fe779dc15f799d008ff0e9589a93ebc293d755edf6cc249530600672d2bc`；API下载、远端文件、数据库三方哈希/大小一致，PDF文本包含受测者`之端是`及完整批准免责声明。审计为generate=1、regenerate=1、download=2且全部成功，candidate PDF标志已更新；production未修改。当前批准仅用于staging客户测试，CSV人工审阅源仍保留draft/pending_review，客户确认后再决定是否固化为正式源文件或退回修订。
- [FB-115 客户模板视觉重构 - 2026-08-12] 用户反馈生成报告与权威DOCX/PDF差异较大。按10页同DPI逐页对照取得RED后，复用DOCX内嵌原始封面插画，重做封面标题/装饰、阅读说明连续正文与双色矩阵、总体五档环绕图、一级得分表/饼图/分析框、居中10轴雷达和客户式跨页维度详情流；纠正旧契约“每页固定两个维度”为定义/得分/诊断按原稿连续跨页。首次部署因屏幕`min-height:1123px`压过打印高度生成19页，补充打印min-height后恢复10页；逐页图像复核又修正信息密度及重复页眉页码。最终前端index SHA-256=`f5f886ede1bbc6735b4449b0e35d97350c315611d998d3277eca880e5fe12219`，回滚备份=`/opt/talent-assessment/dist.bak.fb115_header.20260812_141215`。最终报告ID=`60f14a12-bef8-42d3-a22a-831ea260b2c2`，completed、A4 10页、1068584 bytes、SHA-256=`deed692f984a69ff5a01eee4397beb51cbf9995eb9816e80c10f1ff114939d66`；十维名称、完整批准免责声明和实际可见页面均通过验证。专项12/12、前端全量23文件144项及production build通过；66条正式活动文案与staging双批准保持不变，短时会话0，三服务active，公网health正常，production未修改。视觉已显著接近客户样例，但最终接受度仍由客户测试确认。
- [FB-116 Word模板 + Microsoft Graph改造 - 2026-08-12] 用户选择Microsoft Graph。已从客户权威DOCX确定性生成运行模板`configs/export-templates/competency-phase1-report.docx`（SHA-256=`7808bd325d51e4967c0bb128358bbcdb1153f3312aac3465cf60f157090bf991`），保留Word固定正文/版式，加入49个必需动态占位符并映射12个原生图表；客户可调整字体、分页、表格、图表样式和固定文本，误删占位符或改变图表数据点数量会拒绝生成。后端以冻结正式DTO填充人员、时间、总体、两组、十维、效度、免责声明和图表缓存；Graph客户端使用client credentials上传临时DOCX、有限退避下载并验证`%PDF-`、成功/失败均清理远端文件，错误不泄露令牌/密钥；生成/下载API、报告实例、SHA和审计契约不变，Graph失败可显式回退Chromium。RED为缺Word/Graph符号编译失败；GREEN为专项、真实模板契约、Graph mock成功/非PDF清理/暂态重试、Go全量、go vet、Windows/Linux build。部署前完整备份=`/opt/talent-assessment/backups/phase1_word_graph_20260812_154957`；staging最终后端SHA-256=`3e394ec03fadd085b98bc35e1b406a8761250de568a0e6699bead70bbe048980`，模板49占位符/12图表、三服务和内外health通过，Word功能默认关闭时真实Chromium强制重生成仍为A4 10页且会话0。用户暂时无法提供Microsoft 365组织租户的tenant/client secret/drive凭据，因此未修改systemd共享环境、未启用Graph、未取得真实Graph PDF；production未修改。
- [FB-116纠正 / 参考MBTI本地转换 - 2026-08-12] 用户确认切换为“Word模板填充→服务器LibreOffice→Chromium兜底”，Microsoft Graph降为可选转换器，因此tenant/client secret/drive不再是当前主路径阻塞。新增独立`pkg/libreofficepdf`，每次转换使用0700临时工作区、0600 DOCX、隔离LibreOffice profile、单槽顺序队列、20MiB/50MiB限制、`%PDF-`验证、上下文超时和finally清理；命令输出不进入API错误。Windows实测必须使用可等待的`soffice.com`且profile URI为`file:///C:/...`，不能直接依赖`soffice.exe`启动器。生成器对客户Word做确定性的LibreOffice垂直间距兼容，并只将阅读说明正文缩小10%；当前运行模板SHA-256=`42647b8c7782932227ed18da59541f63730c7c3ba7c915c153f15302ff1738f5`，仍为49占位符/12图表。真实LibreOffice 26.2.2.2以生产长度总体、效度、免责声明和十维诊断文案转换为A4恰好10页；单测覆盖有效PDF、非PDF、命令失败、不泄露输出、路径穿越、排队取消和临时目录清理，Go全量、go vet、Windows/Linux build通过，Linux产物SHA-256=`5d880a60bbaaeab5681e79d90d44733bc49990b09e325b1701c6aa31e6322203`。staging公网health正常，但TCP/22连续预检超时，因此尚未备份/上传/启用LibreOffice主路径，FB-116继续为STAGING BLOCKED；production未修改。
- [UF-007 / FB-117 Word内容控件模板 staging完成 - 2026-08-12] 用户截图证明把内部`{{...}}`长字段直接暴露给客户会造成换行、下划线/图形遮挡和固定图片难维护。用户确认Microsoft Word内容控件方案。生成器现将49个字段转换为页面正常示例值+隐藏稳定Tag，运行时按Tag填充并拒绝缺失/重复控件，同时兼容旧可见token模板；复用源DOCX ZIP元数据后连续两次生成SHA一致。最终模板SHA-256=`10b323f0b9a223ae5ac31d12b89b7d9db940dba0caa9a9ee0b1a4cf5b112aedb`，契约为49 Tag/49唯一/12图表/0可见token；最终Linux后端SHA-256=`bc8b8ae2b0d168ce8dc04a567eca7f390b3fc8d290235d7a8eb1db0a2a456f94`。staging备份=`/opt/talent-assessment/backups/phase1_word_libreoffice_20260812_163958`，0700/0600、数据库gzip和SHA均通过。真实完整paper强制生成报告ID=`60f14a12-bef8-42d3-a22a-831ea260b2c2`，LibreOffice 24.2、A4 10页、532074 bytes、SHA-256=`006801a23298c3f870e2779832bba1605dcda18c3c87b4239575fe907a693384`，关键文本、未解析字段、API/文件/DB哈希全部通过；临时工作区和短时会话0、三服务active、公网health正常、关键日志0。production未修改。
- [一期Word模板重新部署 - 2026-08-12] 用户在Microsoft Word中继续调整模板后要求重新部署。修改后本地模板SHA-256=`ab9949d48aaf8db9548cf0b2db8030afacce51d5a8aa9489f07324692e7ec9e2`、561724 bytes；专项契约测试通过，仍为49个唯一内容控件、12个图表和0个可见token，本地LibreOffice可正常转换。staging旧模板备份位于`/opt/talent-assessment/backups/phase1_word_template_redeploy_20260812_171922`（目录0700、文件0600）；部署后远端模板哈希与本地一致。对paper=`93a26b3b-047d-4ad9-890a-3b4fe6d042e4`强制重生成报告成功：LibreOffice 24.2、A4 11页、527210 bytes、SHA-256=`8b244e26dba9afb886034e7bbb5d5797753474446fce16548c337452bf34a1e5`，人员文本和未解析字段检查通过，API/下载/数据库哈希一致。模板契约49/49/12/0、临时工作区0、短时会话0、三服务active、公网health正常、关键日志0。此次仅重新部署用户调整后的模板，后端未变，production未修改。
- [一期Word模板Git发布与staging部署 - 2026-08-12] 用户随后继续保存模板，先提交并push，之后明确要求部署staging。Git/staging版本SHA-256=`be6de11c063b3c7b90654206a4d7d9928bb906ea0bf625e5f915c7cb0783b052`、568208 bytes；49个内容控件及真实模板填充契约专项测试通过。部署前旧模板备份位于`/opt/talent-assessment/backups/phase1_word_template_git_release_20260812_175604`（目录0700、文件0600）；部署后远端模板哈希与Git版本一致。对paper=`93a26b3b-047d-4ad9-890a-3b4fe6d042e4`强制重生成成功：LibreOffice 24.2、A4 11页、522969 bytes、SHA-256=`233601d49464bcb3d4e9541e30e53de32afedeaa588f27f2cab74aaabb3c42c1`，人员文本、未解析字段、API/下载/数据库哈希全部通过。远端模板契约49/49/12/0、临时工作区0、短时会话0、三服务active、公网health正常、关键日志0。production未修改。
- [一期Word模板再次发布 - 2026-08-12] 用户再次调整模板后要求发布。发布前契约门禁发现`dimension.competency-a1-03.diagnosis`重复2次且`dimension.competency-a1-04.diagnosis`缺失；根据重复Tag之前最近的可见标题，确认第二个重复控件位于“持续学习”块。保留原模板到本地ignored临时目录后，仅修正该隐藏Tag，不改可见版式、图片或文字。修复后模板SHA-256=`3b6a83fd4a2fddf7c0a47c1eda5e2e4141b7d0d72fd9431980928be598e86b92`、557442 bytes，契约为49 Tag/49唯一/重复0/12图表/0可见token，专项填充和LibreOffice转换通过。staging旧模板备份=`/opt/talent-assessment/backups/phase1_word_template_rerelease_20260812_181538`；真实完整paper强制重生成报告为LibreOffice 24.2、A4 12页、537496 bytes、SHA-256=`b065e0b450c5994bad525cc8132f3cd66a0a1c84918e03679bdfef64475eb6a9`，人员文本、未解析字段及API/下载/数据库哈希一致。临时工作区/短时会话0，三服务active、公网health正常、关键日志0。production未修改。
- [一期Word模板管理页面 staging完成 - 2026-08-12] 用户确认合并到既有“报告模板”页面，并采用“严格校验后自动生效”。页面顶部新增00401一期卡片，显示文件名、大小、修改时间、SHA-256、49内容控件/12图表/可见token状态，支持下载、选择DOCX、二次确认上传、loading和失败后保留文件；下方16种MBTI表格保持不变。后端新增管理员/全局权限专用元数据、下载、上传接口；上传限制20MiB，验证DOCX正文、精确49唯一Tag、chart1–12数据点及零可见token，在同目录临时写入并保留timestamped 0600备份，报告生成与上传共享RW锁。RED为后端缺校验/安装符号、前端缺API与页面能力；GREEN为后端契约/重复Tag/备份替换/权限/路由测试及前端3项专项、前端全量、Go全量、go vet、Windows/Linux build和production前端build。staging部署备份=`/opt/talent-assessment/backups/phase1_template_management_20260812_195444`；后端SHA-256=`210aac516088e610e8b5be6307391e59062f0a0d3053006277843c902f72d639`，前端index=`1afdb1ce766546584a2c1c1e5d0d492bead56d54b98e0653df2394a38fa0a80c`。首轮真实上传发现服务以liming运行而模板目录root:root 755不可写；经用户确认仅将`export-templates`改为`liming:liming 750`。随后真实API完成元数据、DOCX下载、重复Tag拒绝且SHA不变、合法上传备份0→1及立即生效；真实浏览器`/qu/template`完成下载请求、上传成功刷新，1440×900和390×844无横向溢出。最终模板SHA保持`3b6a83fd...`，短时会话/临时工作区0、三服务active、公网health正常、关键日志0。production未修改。
- [UF-009 / FB-118 一期个人信息首次保存 staging完成 - 2026-08-13] staging截图与用户5问确认：新建一期胜任力测评只选姓名/性别/手机号，首次进入考生页却显示默认六项，二次编辑保存后才一致。根因位于前端`handleSave()`：新测评尚无ID，每次保存前调用`applyPhase1Profile()`都会把`requiredFieldsList`重置为姓名/性别/年龄/手机号/单位/岗位；首次持久化因此写入六项，而编辑时已有ID所以不再重置。修复为仅`handleAssessmentTypeChange('competency')`调用`applyPhase1Profile(true)`初始化默认六项；保存和详情回填调用默认false，只固定受众/十维/四版本/时长，不覆盖用户字段子集。FB-118专项先RED后5/5 GREEN，前端全量和production build通过（仅2个既有体积warning），编辑器诊断0。staging前端备份=`/opt/talent-assessment/backups/fb118_required_fields_20260813_134536`，部署index SHA-256=`15a66ff8234068ba0384faa995f251c6175db90ba5644c8ab32b57db94e3446c`。真实浏览器新建一期测评后，首次保存请求和Detail均为`name,gender,telephone`，考生信息页立即只渲染姓名/性别/手机号；最终临时测评和短时会话0、三服务active、公网health正常、关键日志0。production未修改。
- [一期准备页测评描述更新 - 2026-08-13] 按用户提供的最终文本替换一期胜任力准备页短描述：说明评估工作情景行为表现与内在倾向，增加“作答说明”及无对错/按第一反应两条规则、严格保密与使用范围、感谢语，并以红色固定提示90题全部必答。旧实现按十维`questionCount`只合计80道维度题，漏掉10道效度题；现一期固定产品不再动态求和。专项测试先RED后GREEN，前端全量和production build通过（2个既有体积warning）。staging前端备份=`/opt/talent-assessment/backups/phase1_description_20260813_142414`，index SHA-256=`07206680f9a71a3ab3c5174d5104061fd3aa5934f9a63cb4c7562cc71bd562a4`；真实测评`1786588375737209899`准备页完整文本和90题提示通过，旧80题/短描述均不存在。短时会话0、三服务active、公网health正常、关键日志0。production未修改。
- [UF-010 / FB-119～121 答题交互与微信入口 staging完成 - 2026-08-13] 用户要求一期答题保存后自动下一题、隐藏内部题号和“五级量表”，并反馈微信扫码白屏。自动跳题实现先捕获当前题与索引，只有保存成功且仍停留原索引时向后移动一题；失败清空本次值并停留，最后一题受边界保护且绝不自动交卷。参与者页面已删除内部code/type节点和样式。微信Android UA直接重建的hash URL当时可加载，说明服务/API本身正常；为规避微信扫码器和内置WebView对二维码fragment的兼容差异，二维码统一改为无hash的`/exam-entry.html?examId=...`，由仅使用ES5语法的静态页校验参数后跳转SPA candidate/tester路由。专项测试先4项RED后17/17 GREEN，前端全量和production build通过。staging备份=`/opt/talent-assessment/backups/fb119_121_answer_mobile_20260813_144204`，index SHA-256=`588f7fb1f95ad61fa4780c265ce7492bc917fcd23c1d1f2e92a7fae72489f62b`，中转页SHA-256=`28af89d9d1417f2c0aa166a3c7f6d8a540ca6e59c8fa0eef9c61cdadf54a15c7`。staging真实静态资源E2E验证成功保存9→10、失败停留、40题末题停留且submit=0、标签隐藏与刷新恢复；真实测评`1786588375737209899`经微信Android WebView UA中转入口加载考生信息，控制台/请求错误0、390×844无溢出。三服务active、公网health正常。production未修改。
- [一期Word内嵌Excel模板 staging切换 - 2026-08-13] 基于当前模板生成`competency-phase1-report-embedded.docx`。生成器创建一个有效XLSX，Sheet1承载2个一级得分，Sheet2承载10维雷达数据及10组环形图“得分/距5分差值”；12个chart关系由外部OLE文件改为内部package并统一指向该工作簿，所有公式同步引用内嵌文件，外部本机路径归零。运行时在更新`chartN.xml`缓存的同时，使用excelize更新B2:B3、B4:B13和B33:C42，避免Word刷新后恢复旧示例值。RED为缺候选文件和`phase1ChartWorkbookPath`；GREEN覆盖49/12上传门禁、工作簿可读、12关系/公式、缓存与Excel双写。生成器连续两次SHA一致；模板SHA-256=`231385fc3f1082a8096e59b84b3210ee4e95c48aacf4c71d1baf9be1003ff20f`、559108 bytes。staging切换备份=`/opt/talent-assessment/backups/phase1_embedded_excel_switch_20260813_162033`；后端SHA-256=`cd2aad3edc5bf8e2a48cbb1afd9cf950d4bc4db2864f5d106cba69103b94259f`。远端实证为1个内嵌工作簿、12个package关系、0外链；完整paper强制重生成报告ID=`60f14a12-bef8-42d3-a22a-831ea260b2c2`，LibreOffice 24.2、A4 12页、537663 bytes、SHA-256=`27005cc70f08c6bfdbe1909e311675de8901d7f26e00177da103c71e3d10bad1`，API/下载/DB哈希一致。真实页面总览发现页脚仍写`/13`而物理页为12、第4/11页偏稀疏、部分维度模块跨页；已写入客户维护说明作为下一轮版式修订项。短时会话/临时工作区0、三服务active、公网health正常、关键日志0。production未修改。

## 2026-08-11 一期正式报告框架（本地）

- 已新增独立的 `competency-phase1-report-data-v1` 强类型契约：固定总分/50、一级/二级维度各/5，严格要求完整的2个一级维度、10个按固定顺序的二级维度及效度good/questionable；受测者DTO不含效度原始分、35分阈值或旧generic `evaluationAverage`。
- 已新增固定十页一期Vue分支：封面、阅读说明、个人/总体/效度、一级维度、十维全景，以及5页每页2个二级维度。旧generic四档动态报告保留在独立else分支，未改其数据和布局契约。
- 已新增010报告框架迁移和内容包模型，仅创建双重批准门禁表，不插入任何approved内容包。批准必须精确匹配四类一期版本与基层员工对象，并同时具备内容批准、测量批准、两个SHA-256、适用环境和免责声明。
- 一期生成数据和下载入口已接入稳定批准门禁；当前没有approved内容包，仍返回“一期正式报告内容尚未完成双重批准”，不会创建报告实例、渲染PDF或下载历史文件。候选JSON中的 `productionContentApproved=false` 未被冒充为正式批准。
- RED证据：Go最初因内容包模型/批准校验/一期快照/十页DTO缺失而编译失败，Schema测试因010不存在而失败，Vue因无一期分支失败。GREEN证据：报告相关Go三包通过；Go全量通过；Go Build通过；前端23文件140项通过；production build完成；相关编辑器诊断0。
- [未执行/未部署] 010尚未对本地或staging数据库执行；本切片未导入正式文案、未创建approved内容包、未开放一期报告生成/下载、未部署staging或production。
- [纠正 / staging部署完成 - 2026-08-11] 上述“010尚未执行、未部署staging”已失效；production仍未修改。部署前完整备份位于`/opt/talent-assessment/backups/phase1_90_import_20260811_145819`，目录/数据库归档权限为0700/0600；`element.sql.gz`为12,116,097 bytes，`gzip -t`通过，SHA-256=`6e21f8760fcf0e2f261a38fa771ed88f8334a8073e2815107af7e824274f0891`。应用替换前备份为`server.bak.20260811_145954`和`dist.bak.20260811_145954`，旧后端/index SHA-256=`71f9bb0df79ed0e5a21d39a484db75a1e45957ee133820604dfe06ceb5d178ab`/`2b4555c78cea87db227629ce91c0b91691459061cc43965469126c76c9bcbdc3`。
- 010迁移在staging连续执行两次成功；内容包表为17列，主键、五版本/受众唯一索引和状态查询索引签名正确，approved/任何内容包行数为0。迁移后传统题、题库关系和答案三组签名与备份一致。
- 部署后后端/index SHA-256=`9dc345fda54cc8bc3f975532254a45a20732b45b0208b0cd89d18d9a2be1e9f2`/`00bded900e719b99022f76bbf6e9bcdf450404aff15b5cbec02f7d774e930784`，本地、远端和公网index一致；生产bundle包含一期隔离分支，远端真实Chromium可加载登录页DOM。
- staging完整正向链再次通过：发布2组/10维/90题并幂等，90题选项快照、试卷恢复、全答、重复提交、10+2+1+1结果、维度/一级3/L3、总体30/weak、效度10/good、筛选和三Sheet导出全部正确。报告数据、生成、下载三个入口均精确返回“一期正式报告内容尚未完成双重批准”，Content-Type非PDF，内容包/报告实例/审计=`0|0|0`。
- 三组负向链再次通过：超时88/90且总体/相关一级二级/效度正式分为NULL；效度40/questionable且默认排名排除；强制发布快照失败后组/题/关联事务回滚为0并可重试。全部验收finally清理成功。
- 终验维度/维度题/效度题/测评/总体/一级/效度/内容包/报告/审计=`10|80|10|0|0|0|0|0|0|0`，新增孤儿和测试触发器=`0|0|0|0`，传统签名不变，临时Redis会话0，talent-assessment/nginx/mysql均active，内外health正常，部署窗口应用关键错误和Nginx 5xx均0；远端部署/验收临时文件已清理。正式内容仍未导入，报告生成/下载继续关闭。
- [追加测试 - 2026-08-11] 使用本机Puppeteer Core加载staging真实production bundle，并仅mock内部报告数据响应。1440×900桌面和390×844移动端均真实渲染恰好10个报告页、5个双维详情页、10个维度卡和2个一级组；十个A/B编号/名称、总体30/50和效度提示均可见，无横向溢出、浏览器console错误0。页面与网络请求均未出现效度原始分、35/36阈值或`validityScore`，内部token未进入API query，输出`STAGING_PHASE1_REPORT_FRAMEWORK_UI_PASS`。
- [追加测试 - 2026-08-11] 在staging按主键临时插入四版本与受众精确匹配、两个SHA/批准人/时间均填写但`approval_status=draft`的内容包；重新执行完整90题真实链后，报告数据/生成/下载仍全部被同一双批准错误拒绝，报告实例/审计均0。EXIT trap按主键删除draft包及远端脚本。终验内容包/胜任力测评/总体结果/报告/审计=`0|0|0|0|0`、短时Redis会话0、三服务active、内外health正常、追加窗口关键日志0。
- [FB-105/106修复与staging实证 - 2026-08-11] 继续测试发现两个已获批路径缺陷：`FormalReportData`在内容包通过双批准后固定返回“内容尚未导入”，导致一期强类型构建器永远不可达；修复后又由真实E2E发现下载入口在一期专属批准门禁后仍调用只支持generic的版本校验器，已生成PDF返回JSON而非文件。两项均先登记并取得RED，再实现GREEN。
- FB-105现从DB读取精确版本/受众的总体、2组说明、10个当前L1-L5维度文案、当前效度提示和免责声明，拒绝临时文案、缺项及批准包免责声明不一致，然后调用`BuildPhase1ReportTextSnapshot`与`BuildPhase1FormalReportData`输出`competency-phase1-report-data-v1`。报告生成handler独立读取持久化结果，不再假定generic map中的result是model值。FB-106使一期下载在专属批准校验通过后跳过generic-only版本校验，generic历史路径保持原校验。
- 本地证据：FB-105/106定向RED→GREEN；Go全量、`go vet`、Windows/Linux build通过；前端既有23文件140项保持通过。FB-106最终部署前完整备份为`/opt/talent-assessment/backups/phase1_90_import_20260811_152747`，数据库归档12,116,520 bytes、`gzip -t`通过、SHA-256=`b3a292ee6d9cd7086c64aea36d26b91fd6d03b59359a3f5a5514c0da3d01ac6a`，旧后端SHA-256=`98e573fe738c20e094e171c3babb3437232f2ba3e03fbb5f7dee6c5ac513f54a`；最终后端SHA-256=`b971a72bb5e9dea36334bd6a9ee622e446e6b395ddbce7b8427529ed93677290`。
- staging临时approved实证仅安装当前测试结果所需的14条英文框架文案和1个测试批准包：真实发布/90题作答/10+2+1+1结果后，报告数据返回10页/2组/10维DTO且不含`validityScore`；Chromium生成completed报告，下载为`application/pdf`且`pdfinfo`恰为10页，成功生成/下载审计2条。finally删除生成PDF、下载副本、测试报告/结果/测评、14条文案和批准包。
- 清理approved测试内容后再次执行默认正向门禁及超时不完整、效度存疑、发布回滚三组负向链，全部通过。最终内容包/一期正式文案/胜任力测评/总体/一级/效度/报告/审计/测试触发器=`0|0|0|0|0|0|0|0|0`，测试PDF和短时Redis会话0，传统签名不变，三服务active、内外health正常、修复窗口关键日志0。正式内容仍未获批，正常staging报告入口继续关闭；production未修改。
- [P0/P1启动 / FB-107本地修复 - 2026-08-11] 扩展一期内容包批准负向矩阵，现覆盖四版本、受众、draft/retired、两类批准人及时间、双SHA、环境和免责声明。正式文案矩阵首次RED为24项中4项失败，证明空白总体文案、空白效度文案和多行免责声明不一致会被接受；实现后构建器对必需文本统一执行trim校验，并拒绝非空免责声明冲突。定向24/24、Go全量、Go Build、`go vet ./...`和无改写gofmt检查均通过，相关文件诊断0。该修复仅收紧一期正式文案校验；generic报告路径和公开契约无需修改。正式内容、双批准资料和production授权仍未提供，本轮未部署staging/production。

## 2026-08-11 一期90题统一运行时（本地）

- 发布已冻结2个一级组、10个A/B维度和80道dimension+10道forward validity题；确认的五级选项文字进入快照，效度全局序号强制完整覆盖1–10。
- 提交已按80/10拆分并原子写10条二级、2条一级、1条效度、1条总体结果；不完整总体分改为数据库NULL，重复提交保持幂等。
- 管理端可筛选/查看效度和一级结果；分数排名默认完整+效度良好；三Sheet导出增加一级分、效度分/状态和题型。
- 发布仅允许管理员或全局权限账号。一期正式报告L1-L5/一级/效度渲染器尚未实现，后端和前端报告入口继续关闭，但测评发布/答题/结果链已解除临时门禁。
- 本地证据：Go全量、go vet、Windows build通过；前端23文件139项及production build通过；Python staging E2E脚本语法通过。尚未部署staging/production。
- [纠正 / staging部署完成 - 2026-08-11] 上述“尚未部署staging”已失效；production仍未修改。独立备份位于`/opt/talent-assessment/backups/phase1_90_import_20260811_093237`，数据库gzip SHA-256=`2d9090cbc62d8b5b3265f76d48b19f50b38b9fbe130f02f83deacb343b5eb7d1`且`gzip -t`通过；旧后端SHA-256=`eb0766cd5a98535c876457eb012f7bc1780f9e39bc4dc09ea75f80f4fe2d65a7`，传统题/关系/答案签名已保存并在终验确认不变。
- 008迁移连续执行两轮成功，`el_competency_result.overall_score=decimal(18,6)|YES`；部署后后端SHA-256=`71f9bb0df79ed0e5a21d39a484db75a1e45957ee133820604dfe06ceb5d178ab`，远端/本地/公网index SHA-256=`2b4555c78cea87db227629ce91c0b91691459061cc43965469126c76c9bcbdc3`。
- staging真实完整链通过：发布2组/10维/90题（80 dimension+10 validity）并幂等；90个确认选项快照、试卷恢复、全答提交与重复提交通过；结果恰为10二级+2一级+1效度+1总体，十维和两组均3/L3、总体30/weak、效度10/good；good筛选1、questionable筛选0；三Sheet导出包含一级、效度和题型；一期正式报告仍按设计拒绝。
- 验收finally清理`exam|paper|candidate|result=0|0|0|0`。终验维度/维度题/效度题/测评/总体/一级/效度结果=`10|80|10|0|0|0|0`，三类新增孤儿=`0|0|0`，传统签名不变，临时Redis会话0，服务/nginx/mysql均active，内外health正常，部署窗口关键日志0。
- [三组staging负向验收 - 2026-08-11] 独立备份位于`/opt/talent-assessment/backups/phase1_90_import_20260811_103057`，数据库gzip SHA-256=`6a76a96682d8620b2b62674bf82f169dd7bd47a6b2d03dae7018166f545d9308`且可解压；当前后端哈希与部署版本一致，传统题/关系/答案签名终验不变。
- 超时不完整真实链：故意留空`A1-01-Q01`和`P1-VAL-Q01`并将limit_time置于过去，经manual入口由服务端可信转换timeout。结果为88/90、维度79/80、overall/evaluation均NULL、恰1条二级和1条一级NULL、效度9/10+incomplete+NULL；管理不完整+效度未完成筛选返回1，默认分数排名返回0，正式报告明确拒绝。
- 效度存疑真实链：80道维度题raw=3、10道效度题raw=4，得到十维/两组均3/L3、总体30/weak、效度40/questionable。显式all和questionable筛选均返回1，默认overallScore排名返回0，三Sheet导出包含效度原始分、效度状态和questionable。
- 发布回滚真实链：临时`BEFORE INSERT`触发器对题目快照强制抛出`PHASE1_RUNTIME_FORCED_SNAPSHOT_FAILURE`。发布拒绝后publish status/group/question/group links=`0|0|0|0`，证明组快照和维度更新随事务回滚；删除触发器后同一草稿成功发布2组/10维/90题。最终三组统一输出`STAGING_PHASE1_NEGATIVE_RUNTIME_PASS`，清理exam/result/group/validity/trigger=`0|0|0|0|0`，短时会话0、孤儿0、服务和内外health正常、关键日志0。production未修改。

## 2026-06-24

- 项目规则要求维护本文件；本次任务开始时文件不存在，已补建。
- MBTI 完整版 PDF 方框字根因：LibreOffice/Linux 对 DOCX 中复杂 `w14:*` 特效、风险字体族、以及仅有 `w:hint="eastAsia"` 的静态正文 run 进行字体 fallback/subset 时不稳定；文本层可提取正常，但渲染层出现方框。
- 已验证生产 ESTP 样本：旧文件 `ESTP_ESTP_20260624193941.pdf` 有方框；部署 FB-044 后强制重生成 `ESTP_ESTP_20260624201230.pdf`，第 4-7 页渲染图未见方框。
- 后端兜底链路：`replaceDocumentFields` 全局剥离 `w14:textFill` / `w14:props3d`；归一化风险字体族；为 hint-only 东亚字体补齐 `Noto Sans CJK SC`；同时归一化 `styles.xml` / `fontTable.xml` 中的不稳定 CJK 字体声明。
- 生产后端当前已部署 md5：`a8110e41c837a99365e6246d4a910a44`，服务 `talent-assessment` active，`/health` 返回 HTTP 200。
- 生产验证脚本：`scripts/tools/recheck-estp-production.ps1` 可用于同一 ESTP 记录强制重生成并下载第 4-7 页渲染图。
[纠正 - 2026-08-25] 上述历史辅助脚本已在测试脚本治理中删除，不再作为可执行入口；如需复验ESTP，须基于当前API重新建立无凭据脚本。

## 2026-07-24

- 胜任力测验需求基线已形成：客户可从 48 个维度任意组合；每个所选维度的全部启用题目进入测评；每位受测者首次组卷时对全题池独立纯随机，题序随后固定。
- 胜任力题目统一使用五级量表；正向题按 1/2/3/4/5 计分，反向题按 5/4/3/2/1 计分；一道题只属于一个维度，考察点仅作题目说明。
- 维度得分为该维度题目最终分的平均值；整体得分为本次各维度得分之和；全部题目必答，限时到期自动提交；到期未答题仅按已答题计算并标记未完整作答；受测者默认不能查看结果。
- 参考文件 `docs/2600724胜任力测验维度.xlsx` 已核验：Sheet1 恰有 48 个连续维度，无缺号、重号、重名或缺失核心含义；基层通用 10 项、管理通用 38 项。
- 胜任力报告采用 1.00–5.00 等距四档；正式分级解读、典型行为和发展建议由客户提供；参考表第 42 项正式名称由“权利动机”修正为“权力动机”。
- 最终需求基线位于 `docs/competency-assessment-requirements.md`；业务规则已确认，当前仅待客户交付 48 维度正式报告文案。
- 胜任力题目采用 AI 分批生成与人工审核：每批 4 个维度，每维度首轮 8 道中性职场题（默认 6 正向、2 反向），共规划 12 批；AI 初稿不得跳过内容审核和预测试直接用于人才决策。
- 第 1 批 D01–D04 共 32 道候选题已完成独立内容审查，审阅稿为 `docs/competency-question-bank-review-batch-01.md`；48 维度进度台账为 `docs/competency-question-bank-progress.md`，当前状态为第 1 批待人工审核、其余未生成。
- 第 2 批 D05–D08 共 32 道候选题已完成独立内容审查，审阅稿为 `docs/competency-question-bank-review-batch-02.md`；D05–D08 已在进度台账中标记为待人工审核，当前累计完成 8 个维度、64 道候选题。
- 第 2 批预测试重点：D05 与适应性区分；D06 与主动性/创新性/认知动机区分；D07 按跨部门协作机会检查岗位公平性；D08 与逻辑思维/问题解决/归纳总结区分。
- 第 3 批 D09–D12 共 32 道测试题已完成一次独立内容审查并修订；D11、D12 后续应按战略信息接触和团队影响机会检查岗位公平性。
- [纠正 - 2026-07-24] 上述“仅第 1/2 批完成、其余未生成”的阶段状态已过时。按用户“仅测试用、一次完成”的要求，D01–D48 共 12 批、384 道测试题现已全部生成；每维度 8 道（6 正向、2 反向），全部明确标记为未经过信效度验证。
- 全量最小机械校验已通过：12 个批次文件、48 个维度、384 个唯一题号，无完全重复题干；进度台账 `docs/competency-question-bank-progress.md` 已更新为 V2.0。D13–D48 未逐批做深度内容审查，仅供系统测试与后续人工审阅，不得直接用于正式人才决策。
- 胜任力落地实施设计已形成，基线文档为 `docs/competency-assessment-implementation-plan.md`；当前尚未修改业务代码、数据库或生产环境。
- 已确认架构：新增显式 `assessment_type=competency` / `scoring_mode=competency_average`，现有 001/002/003 行为冻结；48 维度为独立主数据，不建成 48 个题库。
- 已确认历史一致性方案：测评发布时冻结维度和全部启用题目的元数据/选项分值快照；个人组卷只读发布快照，使用安全随机源执行全题池 Fisher–Yates 并固化题序。
- 已确认计分事实来源：胜任力使用专属维度结果和整体结果表保存 DECIMAL 成绩，现有整数 `obj_score` / `user_score` / `el_user_exam.max_score` 不承载胜任力小数成绩。
- 已确认到期提交采用前端触发与后端过期 Worker 双保险，提交服务使用行锁、唯一索引和状态检查保证幂等；受测者接口使用精确匿名路径和参与者/试卷令牌，管理结果与报告保持管理员权限。
- 数据库变更必须使用 `scripts/sql/` 下兼容 MySQL 5.7 的显式幂等迁移，不使用 `AutoMigrate` 或 `ORDER BY RAND()`；下一步只实施阶段 1，按 1A 安全分流基线 → 1B Schema、48 维度、题目导入与纯计分引擎执行，完成验证后再等待确认。
- 胜任力阶段 1A 已启动：`docs/business-branches.md` 已补充显式类型分流、精确匿名路由、参与者/试卷令牌和旧 paper API 隔离矩阵；共识别 41 个待覆盖业务分支（40 个 P0、1 个 P1）及 6 个验收门禁，业务代码尚未修改。
- 胜任力阶段 1A 首个切片已完成 RED→GREEN：RED 时精确参与者 POST 路由和内部报告 GET 路由共 6 项测试失败，Service 因缺少类型/令牌实现而 build failed；实现后相关测试 66 项通过。
- 新增 `internal/service/competency_security.go`：只允许 legacy+legacy、competency+competency_average 两种类型组合；胜任力参与者/试卷令牌使用 HS512、强制 exp、purpose、participant/exam/paper 绑定，错误不回显 token 或 secret。
- `internal/middleware/middleware.go` 已增加 method+path 精确匿名规则：仅 4 个参与者 POST 和内部报告 GET；不同 method、后缀路径和管理/结果接口保持后台 JWT 拦截。真实 httptest 验证精确路径返回 204、后缀路径返回 401。
- 2026-07-24 验证：`Go: Build` 通过；`Go: Test All` 通过；4 个改动 Go 文件诊断为 0。阶段 1A 尚未完成：旧 Paper/标准分 API 的胜任力隔离需等待阶段 1B Schema 增加 `assessment_type` 后接入，API quick regression 也保留到该切片完成后执行。
- 胜任力阶段 1B Schema 首个切片已完成 RED→GREEN：RED 时 `Exam` 缺少 5 个字段且迁移文件不存在（2 项失败）；新增模型字段和迁移后 2 项专项测试通过。
- 新增 `scripts/sql/competency_001_schema.sql`：兼容 MySQL 5.7，以 `information_schema + PREPARE` 幂等增加 `assessment_type`、`scoring_mode`、`publish_status`、`published_at`、`published_by` 和 `idx_exam_assessment_publish`；只新增结构，不删除或覆盖数据，旧记录默认 legacy+legacy+已发布。
- `model.Exam` 已同步上述 5 个字段；新增静态测试验证字段 tag、MySQL 5.7 幂等片段，并禁止 `ADD/CREATE ... IF NOT EXISTS`、`AutoMigrate`、DROP。
- 新增本地工具 `scripts/tools/apply-local-migration.js`，从本地配置读取 DSN、使用现有 mysql2 执行指定迁移且不打印连接串或密码；`node --check` 通过。
- [未验证 - 2026-07-24] 本地迁移实际执行失败：配置端点 `127.0.0.1:23306` 返回 `ECONNREFUSED`，因此本地数据库尚未新增字段/索引，也未取得 SQL 查询结果。`Go: Build` 与 `Go: Test All` 均通过，但在迁移成功前不可启动使用新增 `Exam` 模型的运行时服务。
- [新增确认 - 2026-07-24] 胜任力报告分为基层员工版和领导人员版。选择环节固定在测评配置阶段，发布后冻结；两个版本共用同一 Vue/PDF 模板、模块、得分、等级、维度顺序和图表，仅总体评价与发展建议文案不同。
- 报告对象使用独立字段 `competency_report_audience`（`frontline_employee` / `leader`），不复用历史 `stu_flag`；报告生成时从结果/报告快照读取版本，不能临时选择或回退到另一版本。
- 已将 `competency_report_audience` 纳入尚未执行的 `competency_001_schema.sql` 和 `model.Exam`；新增白名单校验。RED 时模型/迁移缺字段共 2 项失败，GREEN 后专项测试 35 项通过，`Go: Build` 与 `Go: Test All` 全部通过，相关文件诊断为 0。
- 正式文案匹配键已确定：总体评价按“报告对象 + 总体等级 + 内容版本”，发展建议按“报告对象 + 维度 + 得分等级 + 内容版本”。目标版本文案缺失时生产报告必须失败，不得自动编造或使用另一版本。
- [实现 - 2026-07-24] 创建/编辑测评表单已增加“传统测评/胜任力测评”类别。选择胜任力后，可选择基层员工版或领导人员版，并从 48 维度中多选；传统题库组卷控件自动隐藏，发布后报告版本和维度只读。
- 新增 `el_competency_dimension` / `el_exam_competency_dimension` 模型和 `scripts/sql/competency_002_dimensions.sql`；迁移包含 48 条幂等主数据、唯一约束和 D42“权力动机”，不删除数据且不覆盖管理员维护的 status。
- 新增管理员 API `POST /exam/api/competency/dimensions/list`；测评 Save 入口校验类型、报告版本、至少一个且不重复的维度，并在事务中拒绝不存在/停用维度、批量保存关联；Detail/Paging 返回类型、报告版本、发布状态和已选维度。
- 已发布胜任力测评的报告版本和维度集合在后端禁止修改；前端同步只读。草稿可编辑并恢复选择；删除无业务关联的测评会清理维度配置关联。
- 创建配置切片 RED：缺少维度校验函数、模型和 002 迁移导致 service/model build failed。GREEN：专项 Go+Vue 测试 45 项通过；前端全量 Vitest 6 文件 67 项通过；Go Build、Go Test All、前端 production build 均通过。
- 前端 production build 仍有 3 个既有 warning：`paper/paper/index.vue` 引用不存在的 `listCaptures` 导出，以及 asset/entrypoint size 超限；本切片未引入新的编译错误。
- [未验证 - 2026-07-24] 因本地 MySQL `127.0.0.1:23306` 仍拒绝连接，001/002 迁移、维度 API、测评保存/详情数据库往返和浏览器真实操作尚未验证；不得部署或启动依赖新增表结构的后端。
- [环境纠正 - 2026-07-24] 用户确认 `pg-azure-dev01.postgres.database.azure.com` 为误提供地址，本项目继续使用单机 MySQL，目标环境为 staging，不迁移 PostgreSQL。
- 工作区规则仍将 `20.200.136.133` 定义为 production，不能作为 staging 使用。2026-07-24 只读连通性检查显示该地址 TCP/22 不可达，SSH `ConnectTimeout=10` 超时；未登录、未执行 SQL、未上传或部署。继续 staging 验证前需取得正确的 staging 主机/IP 和 SSH 用户/密钥对应关系。
[纠正 - 2026-07-24] 用户明确确认上述环境归属已变更：`20.200.136.133` 当前为 staging，允许用于本次单机 staging 迁移和验证；仍禁止将本次变更发布到其他 production 环境。此前 SSH 超时事实仍有效，需要查证用户/密钥并按重试纪律处理。
- staging 历史脚本确认 SSH 用户/密钥为 `liming` + `~/.ssh/vm-ubuntu-go-dev_key.pem`，不是 `vm-positive-dev-key.pem`；使用正确组合和 `ConnectTimeout=20` 重试仍超时。
- staging HTTP 状态已验证：`http://20.200.136.133/` 可访问登录页，`/prod-api/health` 返回 `{"status":"ok"}`，说明应用在线但管理端口 22 被网络策略阻断。
- 当前 Azure CLI 已登录，但当前账号可访问的全部订阅均未找到公网 IP `20.200.136.133`，不能使用 Azure Run Command；该 VM 属于其他账号/订阅。
- 当前客户端公网 IP 为 `20.239.176.250`；继续 staging 迁移需在该 VM 所属 NSG/防火墙临时允许来源 `20.239.176.250/32` 访问 TCP/22，或由有权限的账号执行 Run Command。
- staging Linux 后端产物已重建：`Go-based Refactored System/bin/server-linux`，44,505,510 bytes，SHA-256 `5981f1277b68bfcd6bc8b4965dc89f02b52ebc5451f97443334971dee723680c`；前端 production build 已生成 dist。尚未上传或部署。
- [纠正 - 2026-07-24 20:13 +08:00] staging SSH 已恢复：`20.200.136.133:22` TCP 可达，使用 `liming` + `~/.ssh/vm-ubuntu-go-dev_key.pem` 免交互认证成功；远端主机名 `vm-ubuntu-go-dev`，远端用户 `liming`。此前 SSH 超时状态已失效，可继续 staging 备份、迁移和部署流程。
- [staging 发布 - 2026-07-24] 预检通过：talent-assessment/nginx/mysql 均 active，MySQL 8.0.45，根分区 61G、已用 6.6G（11%）；迁移前 element 有 46 表、60 个测评，胜任力字段/表均为 0。
- staging MySQL 备份成功并通过 gzip 校验：`/opt/talent-assessment/backups/element_before_competency_20260724_201912.sql.gz`，12,081,209 bytes，SHA-256 `b9ddc6146ece2cb5672f3fb2ef67e7d93206126ed721c78bf7ecc15c7d8476c0`。
- staging 001/002 迁移成功：el_exam 新增 6 个字段，`idx_exam_assessment_publish` 2 列，新增 2 张胜任力表；48 维度序号 1-48、code/name 各 48 唯一，D42=权力动机；60 个既有测评全部为 legacy+legacy+NULL audience+published，无非法组合。
- [纠正 - 2026-07-24] 初次最终检查发现 staging `el_exam.id=utf8mb4_0900_ai_ci` 与新关联表 `exam_id=utf8mb4_general_ci` JOIN 报 MySQL 1267。`competency_002_dimensions.sql` 已改为动态继承 el_exam.id 字符集/collation，兼容 MySQL 8 staging 和 MySQL 5.7 general_ci；重跑后 `exam_id=utf8mb4_0900_ai_ci`，JOIN/孤儿检查通过。
- staging 应用部署成功：远端应用备份为 `/opt/talent-assessment/server.bak.20260724_202321` 和 `/opt/talent-assessment/dist.bak.20260724_202321`；后端 SHA-256 `5981f1277b68bfcd6bc8b4965dc89f02b52ebc5451f97443334971dee723680c`，前端 index SHA-256 `678b36d53d012bfef4fff13f6c6b98d8236007ad2620f47060303f0b5aa91d44`，与本地产物一致；service/nginx active，内外 health 均返回 `{"status":"ok"}`，登录页 HTTP 200。
- staging 真实 API 验证通过：维度列表 48；临时创建基层员工版+2维度草稿，Detail 回填一致；编辑为领导人员版+1维度，Detail 与 SQL 一致（competency|competency_average|leader|draft）；删除后 exam+关联残留 0，孤儿关联 0，临时 Redis 会话已清理。
- staging 浏览器验证通过：管理端新建页可切换胜任力，报告版本显示基层员工版/领导人员版，48 维度按 VIRD 分组；实测选择领导人员版及 D01/D02 后计数为 2。旧测评管理页仍加载共 60 条，页面可见 001 学生/职场、002 基层/管理干部、003 MBTI 等既有记录。
- staging 最终检查：TEMP_EXAMS=0、TEMP_ASSOCIATIONS=0、DIMENSIONS=48、DISPATCH_COLUMNS=6；部署后 journal 未发现 panic/fatal/unknown column/missing table/error。Go 全量测试在 collation 修复后再次通过。
- [阶段 1B 题目/计分 - 2026-07-24] 已新增 `competency_003_questions.sql`，以 MySQL 5.7 兼容的 information_schema + PREPARE 方式幂等扩展 `el_qu` 的 question_code、dimension_id、dimension_item_no、observation_point、scoring_direction、question_status，并创建 3 个唯一/查询索引；dimension_id 动态继承维度主表 id 的 charset/collation。`model.Qu` 使用指针字段保留既有题 NULL 语义。
- [阶段 1B 题目/计分 - 2026-07-24] 已新增无 HTTP/DB 依赖的纯计分引擎：正向 raw、反向 6-raw；按维度 ID 分组；未答题不进入分子/分母；零已答维度 score 为 nil 且不计入整体；使用 `math/big.Rat` 保持精确分数，最终显示才保留两位；总体评价按有效维度平均值进行 1/2/3/4/5 边界四档分级。
- RED 证据：新增测试前 service build 因计分符号未定义失败，model 测试因 003 迁移不存在且 Qu 缺 6 字段失败。GREEN 证据：题目 Schema + 计分专项 6 项通过；`Go: Build` 通过；最新 `Go: Test All` 退出码 0。
- staging 迁移前完整备份已成功：`/opt/talent-assessment/backups/element_before_competency_20260724_231128.sql.gz`，12,085,196 bytes，SHA-256 `0d9eed2810a57875d807ad6537f0befcc9d09b043d1c5d81720f56e818ef11d7`。001/002/003 和验证脚本已上传 `/tmp`。
- [未验证 - 2026-07-24] staging 003 迁移连续 3 次 SSH 连接超时，均发生在远端命令启动前；因此 `el_qu` 尚未取得 6 字段、3 索引、collation 和旧题 NULL 状态的真实 SQL 证据。不得部署包含新 Qu 字段的后端，待 SSH 恢复后执行 `/tmp/staging-competency-migrate.sh`。
- [纠正 - 2026-07-24 23:20 +08:00] 上述 SSH 阻塞已解除。TCP/22 可达并成功登录 `vm-ubuntu-go-dev`；`competency_003_questions.sql` 已在 staging 执行，并再次完整重跑 001/002/003 验证幂等性。
- staging 003 实际 SQL 证据：`el_qu` 题目字段 6 个；三个索引共 5 个索引列；`el_qu.dimension_id` 与 `el_competency_dimension.id` 均为 `utf8mb4_general_ci`；既有题带胜任力元数据记录数为 0；第二次执行输出 `MIGRATION_RERUN_OK` 且所有计数不变。
- 迁移后运行验证：`talent-assessment=active`、`nginx=active`、内部 health 返回 `{"status":"ok"}`；最终检查 PUBLIC_HEALTH 正常、登录页 HTTP 200、临时测评/关联均为 0，近期日志无 panic/fatal/unknown column/missing table/error。003 仅改变数据库 Schema，尚未部署本地新增后端二进制。
- [阶段 1A 收口 - 2026-07-24] 已新增旧 Paper API 类型守卫：通过 `el_paper` 与 `el_exam` 的单次索引 JOIN 读取 assessment_type/scoring_mode，严格调用 `ValidateAssessmentMode`；记录不存在、查询失败、非法组合、胜任力分别返回受控错误，查询失败不得回退 legacy。
- 守卫已覆盖旧 `CreatePaper`、`paper-detail`、`paperQu-detail`、`qu-detail`、`fill-answer`、`hand-exam`、`paper-result`、paper/tester/candidate 三类标准分入口。胜任力创建在 `createPaperTx` 前拒绝；答题在空答案快速成功和写事务前拒绝；交卷在写事务前及事务内聚合前双重检查；标准分在固定公式查询前拒绝。
- RED 证据：实现前专项测试 10 个入口均缺守卫、守卫文件不存在、两个写入口未在事务前拒绝。GREEN 证据：运行时模式 + 入口顺序专项测试通过；相关 5 个 Go 文件诊断为 0；`Go: Build` 和最新 `Go: Test All` 均退出码 0。
- [未验证 - 2026-07-24] API quick regression `chain-batch.js` 未进入测试链：它固定访问本机 127.0.0.1:8092，captcha 响应为 null，读取 uuid 时抛 TypeError。当前本地 MySQL 端口此前不可用，未为此次单个切片启动或部署服务；旧 001/002/003 的真实 API 回归保留为 ⚠️。
- [阶段 1B 题目导入基础 - 2026-07-24] 已实现管理员三接口：GET 导入模板、POST 导入预览、POST 正式导入。模板固定九列、一题一行、无合并单元格；预览返回规范化成功行、全部失败行及文件 SHA-256，且不执行数据库写入。
- 导入校验覆盖：维度序号 1-48、维度存在/启用、名称精确一致、题号与维度内题号在文件及数据库中唯一、题干/考察点必填、正向/反向、启用/停用；不包含每维度 7-8 题限制。空备注默认 `AI测试题-未信效度验证`。
- 正式导入要求重传同一 xlsx 和 expectedHash，使用 constant-time 比较并重新执行完整校验；任一行错误不写入；全部有效时单事务 `CreateInBatches(100)` 写 `el_qu`，不创建五套固定 `el_qu_answer`。上传压缩文件限制 10MiB、解压限制 64MiB、单次最多 10000 题。
- RED 证据：service 缺导入表头/数据结构/校验器，handler 缺导入 Handler，专项测试编译失败。GREEN 证据：导入 service/handler 专项全部通过；模板实读验证九列、示例及零合并；HTTP 验证缺文件、错误扩展名和 SHA-256 不一致；相关文件诊断 0；`Go: Build` 通过，最新 `Go: Test All` 退出码 0。
- [未验证 - 2026-07-24] 本切片未部署 staging；正式导入的真实 MySQL 批量写入/回滚尚未执行。下一待办“验证数据库唯一约束”可与 staging 临时导入及清理一并验证，发布纪律要求先累计本地变更再确认统一部署。
- [staging 唯一约束验证 - 2026-07-24] 使用事务临时题目验证两个索引：重复 `question_code=ZZ-CONSTRAINT-CODE-A` 被 MySQL 1062 拒绝，命中 `el_qu.uk_qu_question_code`；重复 `(dimension_id, dimension_item_no)=(competency-d01,900003)` 被 MySQL 1062 拒绝，命中 `el_qu.uk_qu_dimension_item`。
- 验证前后 `el_qu` 均为 855 行，临时 ID/题号残留为 0；验证脚本带 EXIT cleanup 且 DELETE 使用主键 ID 条件。随后 staging PUBLIC_HEALTH 正常、登录页 HTTP 200、临时测评/关联为 0、近期错误日志为空。
- [阶段 1B 启用题数校验 - 2026-07-24] 维度列表现通过一次 `GROUP BY dimension_id` 查询返回每维度 `questionCount`，无 N+1；创建/编辑胜任力草稿时在事务内一次查询所选维度启用题数，任一为 0 即以“维度编号 + 名称”拒绝，并在草稿关联中保存当前 `question_count`。
- 前端维度选择器显示每维度启用题数和已选总题数；维度停用或启用题数为 0 时不可选择。RED：Go 因缺 `validateEnabledQuestionCounts` 编译失败，Vue 两项测试因缺总题数/禁选逻辑失败。GREEN：后端专项通过，前端 Vitest 7 文件 69 项全过，Go Build/Go Test All 通过，前端 production build 成功（仅保留既有 listCaptures 和体积 3 个 warning），相关文件诊断 0。
- staging 只读数据核对：48 个维度全部为零启用题，总启用题数为 0，因为现有 855 道题均为 legacy 且尚未导入 384 道胜任力测试题。因此本切片部署后会按设计禁止创建新的胜任力草稿，必须先部署导入接口并导入测试题后再验证正向保存链。
- [未验证 - 2026-07-24] 本切片尚未部署 staging；题数查询数据库失败的注入测试和 staging 正向创建仍待后续统一部署/导入时验证。

## 2026-07-25

- [胜任力核心闭环完成] 新增 004 运行时迁移：不可变发布题目快照、`el_paper_qu` 原始/最终分字段、维度结果、整体结果及到期扫描索引；所有关联列动态继承被关联列 charset/collation。迁移在 staging 成功并幂等重跑：runtime_tables=3、paper_question_columns=3、runtime_indexes=9。
- 发布服务在行锁事务中严格重校验测评、维度、启用题和题目元数据，批量冻结题目与五级选项 JSON，更新维度题数、发布时间/人员和理论总分；重复发布只读返回原快照摘要。创建页增加独立“发布并冻结题目”按钮。
- 参与者闭环：tester 登录/candidate 保存为胜任力签发绑定参与者令牌；首次组卷锁参与者，使用 crypto/rand Fisher–Yates 洗牌并批量固化题序；重复进入返回同一试卷；独立答题页只渲染一个复杂题卡，支持即时保存、进度、第一道未答、倒计时和失败重试。
- 提交闭环：手工未答拒绝；完整/到期提交使用同一行锁事务和纯计分函数，写唯一整体结果与 N 条维度结果，更新 paper 和 participant end_time；重复提交返回首次结果。可配置 Worker 每 30 秒默认分批扫描到期试卷并调用同一提交服务，shutdown 时停止。
- 管理结果/报告：管理员分页、详情和报告数据 API 只读已保存结果；报告读取 `report_audience` 快照，共用一份 Vue 报告布局。客户正式文案未交付时明确 `reportTextReady=false` 和“正式解读文案待配置”，不自动编造或跨 audience 回退；内部报告 GET 仍需常量时间校验内部令牌。
- RED→GREEN：运行时模型/004 迁移/安全随机算法最初均缺失导致 model/service build fail；实现后 Go Build 和 Go Test All 退出码 0。前端 Vitest 8 文件 72 项通过；production build 成功，仍仅有既有 listCaptures 和资源体积 3 个 warning。
- staging 部署前备份：`/opt/talent-assessment/backups/element_before_competency_20260725_003626.sql.gz`，12,085,769 bytes，SHA-256 `46b605ecd7f26c3afde54e434c4ad14d927eca48814ef615a98b9e860296abc4`。
- staging 累计部署成功：后端 SHA-256 `f2b3c1635728c20d72967c795970c5562805255e32993a3910a4d17fdf0adcb0`，前端 index SHA-256 `e1e6503e86c1ec1f8f9c4917745e4fb9d81d19239d20392fb371ed4f6787e6d9`；talent-assessment/nginx active，health 正常；远端备份 `server.bak.20260725_003922` / `dist.bak.20260725_003922`。
- staging 真实全链通过：临时 D01/D02 各 2 题 → 创建 leader 草稿 → 发布 2 维度/4 题 → 重复发布幂等 → candidate 令牌 → 首次随机组卷 → 重复进入同 paper → 刷新题序固定且集合完整 → 未答完手工交卷拒绝 → 全答 → 完整提交 → 重复提交幂等 → 管理结果/报告。已知答案结果 D01=5、D02=1、overall=6、evaluationAverage=3、level=good、audience=leader；SQL 计数 snapshot/paperQu/dimensionResult/result=4/4/2/1；最终 cleanup_remaining=0。
- staging 最终检查：PUBLIC_HEALTH 正常、登录页 HTTP 200、TEMP_EXAMS=0、TEMP_ASSOCIATIONS=0、48 维度、近期关键错误为空。尚未导入 384 道 AI 测试题；正式报告文案仍未交付，因此当前报告仅为测试布局/数据能力，不通过生产文案验收。
- [测试题导入准备 - 2026-07-25] 已从 12 批审阅稿生成 `scripts/data/competency-test-questions.xlsx`：48 维度、384 题、384 唯一题号、每维度 8 题、每维度 6 正向 + 2 反向；文件 77,890 bytes，本地 SHA-256 `b84979d3784d5d42b3e0aaee344a82a11d89cbe99a5076773f21de7173422ab0`。所有行备注为“AI测试题-未信效度验证”。
- 已新增 `scripts/tools/staging-competency-import-questions.py`：远端内部生成短时管理员会话，经正式预览 API 强制校验 384 成功/0 错误，再以同一 SHA-256 调正式导入 API；导入后核对 384 题、48 维度、288 正向、96 反向、每维度 8 题，并删除临时 Redis 会话。检测到 1-383 条部分数据时拒绝继续，检测到完整 384 条时只读幂等返回。
- [阻塞 - 2026-07-25 09:43 +08:00] 测试题尚未实际导入。执行前备份连续两次 SCP/SSH 超时，TCP/22 检查为不可达，第三次 SSH 复测仍超时。遵循数据库变更纪律，未在无导入前备份的情况下通过 HTTP 绕过执行写入。staging 公网 health 仍为 `{"status":"ok"}`、登录页 HTTP 200。SSH 恢复后顺序固定为：完整备份 → 上传 XLSX/导入器 → 预览 → 正式导入 → SQL/维度 API/业务链验证。
- [纠正 - 2026-07-25 09:49 +08:00] 上述 SSH 阻塞已解除，主机 `vm-ubuntu-go-dev`、talent-assessment 和 MySQL 均 active。导入前完整备份成功：`/opt/talent-assessment/backups/element_before_competency_20260725_094936.sql.gz`，12,109,890 bytes，SHA-256 `33b517b6488b9dc4f24b468a8a24030c14eb49ec043b807b97bf58f6dde64b42`。
- [测试题正式导入 - 2026-07-25] staging 预览 API 返回 success=384、errors=0；正式导入 API 返回 imported=384，文件 SHA-256 与本地一致：`b84979d3784d5d42b3e0aaee344a82a11d89cbe99a5076773f21de7173422ab0`。再次运行导入器返回 `COMPETENCY_QUESTIONS_ALREADY_IMPORTED`，未重复写入。
- 最终 SQL 证据：测试题总数/唯一题号/维度数/正向/反向/启用/测试备注为 `384|384|48|288|96|384|384`；维度分布为 `48|8|8|384`（48 维度，每维度最少/最多均 8）；关联 `el_qu_answer` 行数为 0，确认未重复保存五级固定选项。
- 导入后完整运行时链再次通过：D01/D02 各以已有 8 题 + 2 临时题发布，共 20 快照题；重复发布/恢复同 paper/固定完整题序/未答手工拦截/20 题作答/提交/结果/报告全部通过。统一选择 raw=3 后 D01=3、D02=3、overall=6、evaluationAverage=3、good；SQL snapshot/paperQu/dimensionResult/result=`20|20|2|1`；cleanup_remaining=0。最终 health 正常、登录页 200、临时测评/关联为 0、近期错误为空。
- [数据结构与数据全面核验 - 2026-07-25] 新增只读检查器 `scripts/tools/staging-competency-data-integrity-check.sh` 和本地源文件规范化检查器 `scripts/tools/verify-competency-import-source.js`。staging 全部检查通过，未执行数据修改。
- Schema 证据：维度表 10 个目标字段、测评维度表 12 个、el_qu 扩展 6 个、el_exam 分流 6 个、3 张运行时表、el_paper_qu 3 个答题字段、3 个 DECIMAL(18,6) 成绩字段均符合类型/可空性；18 个必需索引存在，抽查的 7 个关键唯一/复合索引列顺序及唯一性准确；两组关键 JOIN 列 charset/collation 一致。
- 维度数据证据：48 行，ID/code/name/order 均 48 唯一，顺序 1-48，全部启用，无空必填；ID/code 与顺序一致，D42=权力动机；VIRD 分布为 20/10/10/8（Versatility/Integrity/Resilience/Drive），适用类别为基层通用 10、管理通用 38。
- 题目数据证据：384 行、384 唯一题号、48 维度、288 正向/96 反向、全部启用且全部带测试题标记；每维度恰好 8 题，题号前缀/后缀与维度及维度内题号完全一致；必填缺失、孤儿、维度内重复、重复题干、固定答案行、题库关联行均为 0。855 条 legacy 题无胜任力元数据，el_qu 总数 1239。
- 源文件与数据库逐字段比对：按“维度序号、名称、题号、维度内题号、题干、考察点、方向、状态、备注”UTF-8 HEX 规范化排序后，本地 XLSX 与 staging 数据库 SHA-256 均为 `3e7f887453b3aeab81b24dba148e2f151a37bccf31fa9e86a293ed3417a47f5d`，证明 384 行九字段内容一致。
- 运行时一致性证据：测评维度、发布快照、维度结果、整体结果孤儿数均为 0；测评类型组合、答题 raw/final 范围、维度/整体结果计数与等级取值均无非法记录。最终 health 正常、登录页 200、临时测评/关联为 0、近期关键错误为空。
- 核验过程中首次 VIRD 预期误写为 20/16/4/8，实际 002 主数据定义为 D21-D30 Integrity、D31-D40 Resilience，即 20/10/10/8；已修正检查器后全量通过。该失败是检查脚本预期错误，不是数据库数据错误。
- [需求与数据库设计重审 - 2026-07-25] 核心主数据→源题→发布快照→个人答题→维度/整体结果的分层设计合理，当前 staging Schema、18 个关键索引、关联列 collation、384 题数据及孤儿检查均通过；但“测试闭环完成”不等于“生产需求全部完成”，正式上线仍有 P0 门禁。
- P0 代码缺陷：参与者提交接口接受客户端 `submitType=timeout`，`Submit` 只对 manual 检查未答题，未验证 timeout 请求确已到 `limit_time`，可提前提交不完整答卷；正式上线前必须由服务端可信判断到期，参与者不能自声明 timeout。
- P0 删除一致性：通用 Paper 删除未处理/禁止 `el_competency_dimension_result` 和 `el_competency_result`，且 001-004 无外键；删除胜任力试卷后可产生结果孤儿并破坏历史快照。已发布测评/胜任力试卷的删除、归档、匿名化和保留年限必须先确认并实现专用策略。
- P0 历史身份快照：结果表没有 participant_type/participant_id 和报告基本信息快照；candidate/tester 可物理删除，Worker 也靠两表探测参与者类型。需确认人员删除/匿名化政策，并在结果或报告实例冻结参与者类型及报告所需基本信息。
- P0 报告仍是测试能力：数据库尚无 `el_competency_report_text`、`el_competency_report`、`content_version`；正式客户文案未交付，当前固定 `reportTextReady=false`。无法满足正式文案版本冻结、历史重生成和报告实例审计。
- P0 功能缺口：REQ-050 的任一维度排序尚未实现，结果分页当前只按 submitted_at；REQ-051 的胜任力动态导出尚未接入现有导出链。当前 384 道题仍是“AI测试题-未信效度验证”，不能用于正式人才决策。
- P1 Schema 建议：评估为历史实体增加 RESTRICT 外键；增加 `(exam_id,snapshot_order)` 和胜任力范围内题序唯一保障；增加 `(exam_id,submitted_at,paper_id)`；持久化 DECIMAL 时避免 Go float64；迁移除索引名外核对列顺序/唯一性；对已有运行时表执行结构漂移检查。当前无外键但 staging 孤儿为 0，属于“现状正确、约束不足”。
- 另发现发布语义偏差：维度名称/VIRD/类别/核心含义在草稿 Save 时复制，Publish 仅更新题数和 snapshot_time；若草稿保存后、发布前主数据变化，发布快照会保留旧草稿值。需确认期望是“草稿时冻结”还是“发布时冻结”；现有需求/实施计划写的是发布时冻结。
- 待业务确认：时长必须为正、允许不限时还是采用默认值（当前 total_time<=0 隐式 2 小时）；测评结束时间与个人倒计时优先级；不完整答卷是否可生成正式报告/进入排名；优势与待发展维度数量和并列规则；正式报告图表；数据保留与删除；题库正式版本、信效度验收；报告文案首个 content_version 与版本生命周期。
- [业务确认 - 2026-07-25] 客户拥有自己的正式题库，不考虑正式题库版本管理；当前 384 道 AI 题继续只作系统测试。正式报告文案阶段暂时搁置，不作为当前核心测评链门禁。
- [业务确认 - 2026-07-25] 允许删除已发布的胜任力测评，采用整链事务性物理删除：报告实例/文件引用、整体/维度结果、答案/试卷题、试卷、candidate/tester 测评记录、发布题目/维度快照和测评本身全部删除；任一步失败整体回滚，删除后不可恢复且必须零孤儿。
- [业务确认 - 2026-07-25] 胜任力个人答题时长必须显式配置且大于 0，不允许不限时或隐式 2 小时默认值；个人到时必须结束并自动提交，不能继续超时作答。测评结束时间只限制尚未开始者创建新试卷，与已经开始试卷的个人答题时长无关。
- [业务确认 - 2026-07-25] 未完整作答的答卷可以保留管理端结果和审计信息，但不得生成正式报告。
- [运行边界与删除修复 - 2026-07-25] FB-045～050 已按 RED→GREEN 完成：参与者 HTTP 只能发送 manual，服务端锁定试卷后按真实 limit_time 转 timeout；提前 timeout 和无 limit_time 均拒绝。胜任力 Save/Publish 强制 total_time>0，移除隐式 2 小时；首次组卷在创建前检查 exam.end_time，已有试卷恢复不受结束时间截断。
- 正式报告数据入口现强制 `is_complete=1`，不完整答卷仍可通过管理结果详情审计，但无法进入正式报告页面/内部渲染数据。前端倒计时到零仍显示超时重试提示，但请求发送 manual，由服务端可信判定是否到期。
- 已发布胜任力测评删除改为同一事务整链物理删除：维度/整体结果 → 答案/试卷题 → user_exam/candidate/tester → 试卷 → 发布题目/维度快照 → 配置关联 → 测评。Legacy 删除仍保留有关联即拒绝；通用 paper/candidate/tester 物理或逻辑删除会拒绝胜任力实体，必须从所属测评整链删除。
- 发布时现重新读取当前启用维度主数据，按当前 display_order 排列并冻结 code/name/VIRD/category/core meaning/order/题数/snapshot_time，修复草稿保存后主数据变化仍冻结旧值的问题；题目在排序前先做必填元数据校验，避免异常数据 panic。
- 验证：FB-045～050 专项全部通过；Go Build 通过；最新 Go Test All 退出码 0；前端 Vitest 8 文件 73 项通过；production build 成功（仍只有既有 3 个 warning）；相关文件诊断 0。上述新行为尚未部署 staging，整链真实删除和到期负例需部署后验证。
- [staging 边界与删除发布 - 2026-07-25] 部署前数据库备份：`/opt/talent-assessment/backups/element_before_competency_20260725_112239.sql.gz`，12,130,320 bytes，SHA-256 `77d76a7b572f711f3f10f3f4b9b8bf20b03b42930d261d7d4e324b924977a97b`。应用备份为 `server.bak.20260725_112421` / `dist.bak.20260725_112421`。
- staging 累计部署哈希与本地一致：后端 `d2a0e1461e23d097e33ce2461d2cf0937aa7fcf760198a39453dc05611d7860b`，前端 index `8d18943f1b6c0610c4eaf65358b875dd0b683a8a3e0b7659115da894a6b925da`；talent-assessment/nginx active，health 正常。
- staging 真实边界验证通过：totalTime=0 保存被拒且未写 exam；已结束测评发布后无法创建新试卷；参与者自声明 timeout 被拒；通用 candidate/paper 删除被拒并提示删除所属测评；将真实试卷 limit_time 调至过去后由服务端转换为不完整 timeout 结果；管理结果仍可读，正式报告数据明确拒绝“不完整作答”。
- staging 整链物理删除通过：删除前 snapshot/paper/paperQu/dimensionResult/result/candidate=`8|1|8|1|1|1`，调用测评删除后测评、维度/题目快照、试卷、试卷题、结果和 candidate 总残留为 0，脚本 cleanup_remaining=0。
- 部署后完整主链再次通过：发布 20 题、恢复同 paper、固定完整题序、未答手工拒绝、完整答题提交、结果/报告一致；D01/D02=3/3、overall=6、average=3、good，db_counts=`20|20|2|1`，cleanup=0。全库结构/数据完整性检查再次全部通过，384 题规范化哈希不变，最终临时测评/关联为 0、近期关键错误为空。
- [剩余代码盘点 - 2026-07-25] 核心创建/发布/答题/提交/删除链已完成并部署，但管理闭环仍未全部实现：胜任力维度仅有 list，缺启停/维护；题目仅有模板/预览/导入，缺专用分页、详情、编辑、启停及前端管理页；通用 Qu.Save 强制题库和答案，不适合作为胜任力题目维护入口。
- 结果管理仍缺前端页面；前端 `fetchCompetencyResults` / `fetchCompetencyResultDetail` 只有 API 封装无消费方。后端 ResultPaging 只按 submitted_at 排序，尚未实现 REQ-050 的整体分/任一维度升降序。
- 胜任力动态导出尚未接入 ExportRawAnswers/ExportRawData，现有实现仍按 001/002/003 题库逻辑；REQ-051 未完成。通用 GenerateReport 的路由选择仍只支持 repoCode 001/002/003，胜任力 PDF 生成未接入；正式报告文案和报告实例表按用户决定暂缓。
- 后续技术债：结果未冻结 participant_type/participant_id/基本信息；Worker 每份过期试卷探测 tester/candidate，存在 N+1；数据库无外键、无 `(exam_id,snapshot_order)` 唯一约束和 `(exam_id,submitted_at,paper_id)` 索引；DECIMAL 模型仍经 float64 持久化。另需完成 48维度/384题单卷容量、100份随机试卷、Worker与前端并发提交验收。
- 当前胜任力累计改动仍处于工作区未提交状态；staging 已部署不等于版本库已形成可追踪提交，正式交付前需整理提交。
- [结果管理本地完成 - 2026-07-25] 新增管理员胜任力结果页和专用隐藏路由；测评列表仅对 `assessment_type=competency` 显示“胜任力结果”，传统测评继续使用原测试记录/导出/统计入口。结果页支持提交时间、整体分、所选已测维度分升降序，显示 candidate/tester 姓名、手机号、完成度、得分和提交信息，并提供维度得分与逐题审计详情。
- 结果分页后端现使用 `sortBy + sortDirection + dimensionId` 白名单契约；稳定次序追加 paper_id，页大小限制 1-500。维度值通过参数化 JOIN，且先验证维度属于当前测评；examId 必填。candidate/tester 通过同一查询 LEFT JOIN 返回人员类型和基本信息，不在结果循环中执行 N+1 查询。
- 结果管理验证：排序白名单/非法字段/非法方向/缺维度/缺 exam 上下文专项 Go 测试通过；结果页面 Vitest 新增 3 项，全量前端为 9 文件 76 项通过；Go Build、Go Test All、前端 production build 均成功，相关文件编辑器诊断为 0。production build 保留既有资源体积 warning，无新增编译错误。
- [未部署 - 2026-07-25] 本次结果管理改动尚未发布 staging；真实 MySQL 的整体分/维度分分页次序、candidate/tester 返回和浏览器交互仍需按发布纪律在用户确认统一发布后验证。正式 production 未执行任何操作。
- [纠正 - 2026-07-25 12:00 +08:00] 上述“结果管理尚未发布 staging”已失效。用户确认后已部署到 `20.200.136.133` staging；正式 production 仍未执行任何操作。
- staging 部署前完整数据库备份：`/opt/talent-assessment/backups/element_before_competency_20260725_115947.sql.gz`，12,130,319 bytes，SHA-256 `dcc6f51e37bf8d1d6df6edab30a3159134034dd8d50cd09f643febb2d1b7ffb4`。应用备份为 `server.bak.20260725_120032` / `dist.bak.20260725_120032`。
- staging 部署哈希与本地产物一致：后端 `b93d6a00e983af5fbfef087c7731196a59c127c1bff9e6e51e66322b8c9303f8`，前端 index `69ecb86946aa3d99af7b5b574ef746044c91ee6fe060317cdbf25be821b53e87`；talent-assessment/nginx active，内外 health 均为 `{"status":"ok"}`，登录页 HTTP 200。
- 新增 staging 结果管理验证器 `scripts/tools/staging-competency-results-verify.py` 和短时浏览器会话辅助脚本。真实创建 1 个临时胜任力测评、D01/D02、3 名 candidate、3 份各 16 题完整答卷：提交时间降序 High→Mixed→Low；整体分升序 2→6→8；D01 维度分降序 5→4→1；三行均返回 candidate/name/telephone；结果详情为 2 个维度、16 条逐题审计。
- staging 浏览器验证通过：专用路由加载正确测评标题和 3 条结果；整体分升序后首行为 Result Low；切换维度分自动选择 D01，降序后首行为 Result High 且显示维度分 5；详情弹窗显示 overall=8、D01=5、D02=3，逐题审计标签实测 16 行。验证截图确认表格、正反向、原始值和计分值均正常展示。
- 验证后已删除短时 Cookie/Redis 会话及临时测评整链数据：`cleanup_remaining=0`、`temp_result_exams=0`、`orphan_competency_results=0`；部署后 15 分钟关键日志 `panic/fatal/unknown column/missing table=0`。剪贴板中的短时令牌已覆盖失效。
- [题库功能只读检查 - 2026-07-25] 传统题库已具备题库分页/列表/详情/新增编辑/删除、题目分页/详情/新增编辑/删除、批量加入或移出题库、Excel 导入/导出/模板等后端能力；题库删除会拒绝仍被测评引用的记录，并事务清理 `el_qu_repo`。相关 Go 专项回归 14 项通过，编辑器诊断为 0。
- 当前传统题目管理与胜任力题目共用 `el_qu`，但传统 `QuHandler.Paging/List/Detail/Save/Delete` 和 `RepoHandler.BatchAction` 未按 `dimension_id` 分流。结合已验证的 384 道胜任力题无 `el_qu_answer` / `el_qu_repo`，全局传统题目页会把胜任力题混入结果；前端又固定读取 `answerList[0]`、`answerList[1]` 和 `content.substring(1)`，存在页面运行时错误与错误编号展示风险。传统保存还会通过 `Save()` 覆盖未在表单返回的胜任力扩展字段，传统删除可直接删除胜任力源题，属于 P0 数据边界缺口。
- 传统题目分页每行分别查询答案和首个题库，20 行页面约产生 41 次 DB 查询；导出同样每题查询题库和答案，存在明显 N+1。题库分页未应用 pageSize 上限，批量题库操作逐题/逐关联写库且忽略多处 DB error，批量量大时性能和一致性不足。
- 前端题目新增/编辑表单把选择器绑定到 `repoId`，但校验规则检查空的 `repoIds`，且转换为数组发生在校验通过之后；新增保存可能被前端错误拦截。全局题库多选筛选后端只采用首个 repo ID。题库列表隐藏多选/判断计数，却把 `radioCount` 标成“题目数量”，展示语义不准确。
- staging 公网只读探测确认传统导入组件硬编码 `/dev-api/exam/api/qu/qu/import-excel` 返回 HTTP 405，而实际 `/prod-api/...` 返回 JWT 401，证明 production build 的传统题目导入地址错误。胜任力题目目前仅有模板/预览/正式导入 API，缺分页、详情、单题编辑、启停、删除和管理页面。
- [未验证 - 2026-07-25 13:06 +08:00] staging 公网 health 正常，但 SSH 连续三次不可达；此前结果验证的短时会话已清理，浏览器停留在登录页，无法完成本轮真实题库列表/编辑/导入页面操作和只读 SQL 复核。未创建、修改或删除任何题库数据，未部署任何变更。
- [staging 管理员密码重置 - 2026-07-25] 经用户明确确认，通过已认证的 `PUT /system/user/resetPwd` 将 user_id=1 的 admin 密码重置为用户指定值；接口返回 HTTP 200/code 200。随后注销旧会话，使用新密码和新验证码重新登录成功并进入系统首页，证明 BCrypt 持久化和登录校验链生效。项目记忆不记录密码明文；剪贴板中的旧密码已清除。SSH 仍不可达，因此未直接查询数据库哈希。
- [题库 staging 浏览器核验 - 2026-07-25] 新密码登录后，题库列表真实加载 6 个传统题库：00302/00301 各 48 题、00102/00101 各 90 题、00201/00202 各 140 题；默认每页 20，页面仅显示 `radioCount` 并标为“题目数量”。00302 题库内题目页真实加载 20 行，题号、题库名、A/B 选项和时间展示正常；已有题目的编辑页可回填题库、题号、题干和 2 个选项。
- [题库 P0 实证 - 2026-07-25] 全局题目页真实 API 返回 total=1239；第一页 20 行全部是胜任力题、全部 `dimensionId` 非空且 `answerList=[]`。Vue 随即连续抛出 `TypeError: Cannot read properties of undefined (reading 'content')`，表格无法呈现，证实传统/胜任力未隔离及 `answerList[0]` 固定访问已实际破坏页面，不再只是静态风险。
- [题目新增 P1 实证 - 2026-07-25] 在新增题目页面只设置有效 `repoId` 和题号后，直接调用表单校验得到 `valid=false`，错误为“必须选择一个题库！”，此时 `repoId` 有值而 `repoIds` 仍为空；未发保存请求、未写数据库。证明选择器/校验字段不一致会阻止正常新增。
- 本轮题库检查只读完成：没有点击保存、导入、删除或批量操作，没有创建/修改题库数据，也没有部署代码。管理员浏览器会话暂时保留用于后续逐项修复验证。
- [FB-051 本地修复 - 2026-07-25] 传统题目管理现以 `dimension_id IS NULL` 为明确边界：Paging/List/Detail/Export 不再返回胜任力源题；传统 Save 对请求中的胜任力元数据和既有胜任力题目 ID 双重拒绝；Delete 与 Repo.BatchAction 在任何写入前检查整批 ID，命中任一胜任力题即整批拒绝。统一错误为“胜任力题目请使用胜任力专用接口”。
- 前端传统题目列表已移除 `answerList[0/1]` 直接索引，空答案显示 `—`；题号只在确有 `V` 前缀时去前缀，异常数据不再导致整表崩溃或丢失首字符。
- FB-051 RED 证据：新测试初次运行同时报告 Paging/List/Detail/Save/Delete/BatchAction/Export 共 8 个隔离缺口，以及前端 3 个直接索引/安全格式化缺口。GREEN 后 FB-051 专项通过；题库相关测试 7 项通过；Go Test All、Go Build、前端 Vitest 9 文件 76 项和 production build 全部通过，相关文件诊断为 0。
- [未部署 - 2026-07-25] FB-051 仅在本地完成，尚未部署 staging；当前 staging 全局题目页仍会混入384道胜任力题并报错。按发布纪律等待后续累计变更后统一确认 staging 发布。题目新增 `repoId/repoIds` 校验问题和传统导入 `/dev-api` 路径问题尚未修复。
- [FB-052 本地修复 - 2026-07-25] 题目表单的题库选择器、`el-form-item` 和 rules 现统一绑定 `repoId`；提交时在任何答案/表单验证前执行 `syncRepoSelection()`，始终以当前 `repoId` 覆盖 `repoIds`。因此新增题目选择题库后可通过题库必填校验，编辑时切换题库也不会继续发送旧关联。
- FB-052 RED 证据：专项测试首次明确失败为 `rules.repoId=undefined` 且 `syncRepoSelection=undefined`。GREEN 后专项 2 项通过；前端全量 Vitest 10 文件 78 项通过；production build 成功；相关文件诊断为 0。
- [未部署 - 2026-07-25] FB-052 与 FB-051 均尚未部署 staging；真实新增/切换题库保存需统一发布后验证。传统导入 `/dev-api` 硬编码问题仍未修复。
- [FB-053 本地修复 - 2026-07-25] 通用 DataTable 的传统题目导入地址现使用 `VUE_APP_BASE_API + /exam/api/qu/qu/import-excel`，随 development/staging/production 构建自动切换；上传请求头加入当前 `Authorization: Bearer <token>`，不再使用硬编码 `/dev-api` 或匿名上传。
- FB-053 RED 证据：专项测试首次同时失败，实际 URL 为硬编码 `/dev-api/...` 且 `upload.headers=undefined`。GREEN 后专项 2 项通过；前端全量 Vitest 11 文件 80 项通过；production build 成功。直接扫描 production dist 得到 `/prod-api/exam/api/qu/qu/import-excel` 匹配 12 次、`/dev-api/exam/api/qu/qu/import-excel` 匹配 0 次，相关文件诊断为 0。
- [未部署 - 2026-07-25] FB-051～FB-053 均尚未统一部署 staging；传统 Excel 导入真实上传和数据库写入仍需发布后用临时题库/题目执行并清理。未修改 production。
- [FB-054 本地修复 - 2026-07-25] 传统题目分页已移除逐行答案/题库查询：主列表取得当前页 ID 后，一次 `IN` 查询全部答案、一次有序 LEFT JOIN 查询全部题库关联，再用 map 组装。默认20行页面数据库往返由约42次降为固定4次（COUNT、主列表、答案、题库关联）；空页只执行前两次。
- 题库关联批量查询按 `qu_id, sort, id` 稳定排序并保留首关联语义；题库已删除时用 `COALESCE` 继续输出 `[已删题库:<id>]`。COUNT、列表、关系查询错误现全部检查并返回受控错误，不返回部分数据；无答案题的 `answerList` 保持非 nil 空数组。
- FB-054 RED 证据：专项测试首次失败为缺少 `loadQuestionPageRelations`。GREEN 后专项通过；题库相关后端测试22项通过；Go Test All和Go Build通过，相关文件诊断为0。
- [未部署 - 2026-07-25] FB-051～FB-054 均尚未统一部署 staging；实际MySQL查询数量与响应内容需发布后验证。未修改production。
- [FB-055 本地修复 - 2026-07-25] 题库批量加入/移除现统一在一个GORM事务内执行。请求ID先去空/去重；批量校验全部题目均存在且为传统题、全部题库存在；加入时一次删除旧关联并 `CreateInBatches(100)` 重建，移除时一次删除目标组合。旧题库与请求题库均纳入受影响集合，按题库ID稳定顺序重排 `(sort,id)` 并刷新统计；所有查询、删除、插入、重排、统计错误均返回并触发整批回滚。
- FB-055 RED时缺少事务、ID规范化、批量查询/写入、受影响题库集合和事务统计刷新共7类证据；GREEN后专项通过。不存在题目/题库不再静默生成零类型关联或孤儿关联。
- [FB-056 本地修复 - 2026-07-25] 传统Excel导入的sheet题库查询/创建、额外题库ID校验、题目/答案/全部关联写入及sheet/额外题库统计刷新已纳入同一事务。额外题库不存在或任一写入/统计失败时，包含新建sheet题库在内的整批数据全部回滚；原先忽略的额外关联和统计错误已消除。
- FB-056 RED明确命中题库事务外创建、额外关联错误忽略和事务外统计刷新；GREEN后专项通过。
- [FB-057 本地修复 - 2026-07-25] 传统题目导出复用分页批量关系加载器，一次加载全部答案和全部有序题库关联，工作簿循环不再访问数据库。当前约855道传统题的导出关系查询由约1710次降为固定2次；主查询或关系查询失败时在写工作簿前返回受控错误，多题库ID顺序保持 `(sort,id)`。
- FB-054～FB-057连续专项全部通过；题库相关后端测试25项通过，相关文件诊断为0。[未部署] FB-051～FB-057仍需统一部署staging后执行真实批量加入/移除、Excel导入回滚和导出内容验证；未修改production。
- [FB-058 本地修复 - 2026-07-25] 传统题目删除现于事务前查询 `el_paper_qu`，任一所选题被历史试卷引用即整批拒绝并返回引用数。可删除时按答案→题库关联→题目顺序执行，关联读取、三个删除及题库统计刷新错误全部传播并回滚；受影响题库按ID稳定刷新。
- [FB-059 本地修复 - 2026-07-25] 编辑既有传统题目前同样检查 `el_paper_qu`；历史试卷引用题返回“题目已被试卷引用，不能修改”，查询失败返回受控错误，二者均发生在写事务和答案/题库关联替换之前。新增题目不执行历史引用检查。
- FB-058/059均先取得缺守卫的RED失败，再实现GREEN；FB-055～FB-059连续专项通过。[未部署] FB-051～FB-059等待统一staging发布和真实数据验证；未修改production。
- [FB-060 本地修复 - 2026-07-25] 题库分页现复用全局 `capPageSize`（默认10、最大200），并检查COUNT与列表查询错误；数据库失败不再伪装成成功空列表，空记录返回非nil数组。RED时5项边界/错误处理断言全部缺失，GREEN后专项通过。
- 本轮尝试统一部署FB-051～FB-060：Linux后端和production前端已成功重建（后端SHA-256 `5443fd2de74e6a3e1258947b47caa46bb8eaa9873bc60b668c55092c5520108f`，前端index SHA-256 `fd542f2c12cf81d615af5ff6fbfda648f0b581b2ef8176da7bf9eb879001e0e0`），但staging SSH在16:26和16:32两次均不可达，公网health仍正常。未上传、未备份、未部署；production未修改。
- [FB-061 本地修复 - 2026-07-25] 传统题目保存现先去空/去重题库ID，并在写事务内校验全部题库存在。编辑链严格检查原题读取、旧题库关联读取和答案/关联删除错误；任何失败均回滚，不再用当前时间掩盖原题读取失败，也不会留下部分替换。
- [FB-062 本地修复 - 2026-07-25] 题库保存后端新增标题trim/必填校验。新建由服务端初始化零统计与创建/更新时间；更新取消 `Save`，仅允许修改code/title/remark/update_time，客户端不能覆盖create_time和计算统计。RowsAffected=0返回“题库不存在”，成功后重读数据库真实行返回。
- FB-060～FB-062专项通过；FB-061/062均先取得RED再GREEN。[未部署] staging SSH仍不可达，FB-051～FB-062待统一部署；production未修改。
- [FB-063 本地修复 - 2026-07-25] 题目List、题目Detail的答案/题库关联/题库编码查询，以及题库List现全部检查数据库错误并返回具体受控失败；列表初始化为非nil空数组，详情任一关系读取失败都不再返回部分对象。RED时10项错误处理断言全部缺失，GREEN后FB-060～063专项通过。
- staging SSH于16:44再次不可达（公网health正常），因此FB-051～FB-063仍未上传、备份或部署；production未修改。
- [纠正 - 2026-07-25 17:10 +08:00] 上述FB-051～FB-063“未部署”状态已失效。SSH恢复后已重新构建并统一部署到 `20.200.136.133` staging；正式production仍未执行任何操作。
- staging部署前完整数据库备份：`/opt/talent-assessment/backups/element_before_competency_20260725_170919.sql.gz`，12,130,321 bytes，SHA-256 `011b921e26c39eebb23d3466698d22f3ff3900858ea88043279409edc64d3928`。应用备份为 `server.bak.20260725_171029` / `dist.bak.20260725_171029`。
- 部署哈希与本地一致：后端 `38a69e11ea84b16d7cebafe83f185439ca1cc8f9c74754f6be7a4eb052d3f0d9`，前端index `fd542f2c12cf81d615af5ff6fbfda648f0b581b2ef8176da7bf9eb879001e0e0`；talent-assessment/nginx active，内外health均为 `{"status":"ok"}`，登录页HTTP 200。
- [题库staging验收 - 2026-07-25] 真实API核对：传统题=855、胜任力题=384；全局传统题目分页total=855、第一页胜任力题数=0。真实浏览器全局题目页显示20行传统题，未再出现 `undefined.content`；新增题表单的有效repoId校验通过并同步覆盖旧repoIds；上传组件URL为 `/prod-api/exam/api/qu/qu/import-excel` 且包含Bearer认证头。
- 临时题库/题目真实链验证通过：题库客户端伪造radioCount=999后数据库返回0；新增传统题→重复/空ID批量加入归一化为单关联且sort=1→无效题库批量加入被拒并保持原关联不变→移除→重新加入；证明批量成功与失败回滚均正确。
- 传统Excel真实上传验证通过：有效xlsx导入1题并刷新关联统计；引用不存在额外题库的xlsx被拒，临时sheet题库和题目均回滚为0；最终导入题和题库通过正式API清理。历史试卷引用题的删除/编辑均被拒、前后题目/答案/试卷引用计数保持 `1|2|442`；胜任力源题传统删除被拒。
- 两轮验收均 `cleanup_remaining=0`；最终 `temp_qb_repos=0`、`temp_qb_questions=0`、`competency_orphans=0`、最近30分钟关键错误=0。短时Cookie/Redis会话/令牌和剪贴板均已清理。
- [UF-001 / 00401 staging完成 - 2026-07-25] 按用户确认采用“虚拟胜任力题库入口”方案，不把384道胜任力源题写入传统 `el_qu_repo`。题库分页动态插入 code=`00401`、title=`胜任力测验题库`、实时题数=384、`virtual=true`；页面禁用该行选择/传统编辑删除，并导航到专用只读胜任力题目页。
- 新增管理员 `POST /exam/api/competency/questions/paging`，仅查询 `dimension_id IS NOT NULL`，返回题号、维度、维度内题号、题干、考察点、方向、状态，支持维度/状态/题号/题干参数化筛选和稳定排序。前端专用页默认20条、维度可搜索、状态及关键词过滤，明确提示通过专用导入维护。
- 00401部署前备份：`/opt/talent-assessment/backups/element_before_competency_20260725_173607.sql.gz`，12,130,318 bytes，SHA-256 `272ba46d8821ce12fe7b16bd96ad06454ce867a6bbc6c4eaf748e81da597411e`；应用备份 `server.bak.20260725_173636` / `dist.bak.20260725_173636`。部署哈希：后端 `9df58e8be57837355f5088e83342f3c528e5b94de4c561b652a425b4d51b2cf1`，前端index `9423c8a362cd1fe8085122d4fe4e3112c5e2247fe09f4523ccd147e0edc57fb6`。
- staging API验证：题库总数7、00401唯一、名称正确、题数384；专用分页total=384/首屏20；D01+启用筛选total=8。浏览器验证题库列表首行00401/384且复选框禁用，点击进入 `/#/exam/competency/questions`，页面标题“00401 胜任力测验题库”、首屏20、总数384，首题D01-Q01及维度/考察点/方向/状态展示正确；截图已人工核验。
- 00401本地验证：Go全量、Go Build、前端Vitest 12文件82项均通过；UF-001专项由12项缺失RED转GREEN。终验service/nginx active、health正常、临时行0、最近15分钟关键错误0，短时Cookie/Redis会话/令牌/剪贴板已清理；production未修改。
- [00401题目编辑/启停本地完成 - 2026-07-25] 新增专用 `POST /exam/api/competency/questions/update`；独立request仅接收id、题干、考察点、方向、状态、备注，题号/维度/维度内题号/create_time不可提交或覆盖。后端校验空ID/题干/考察点、方向白名单及状态0/1，兼容JSON数字/字符串并拒绝空串、nil、小数和越界；更新限定 `dimension_id IS NOT NULL` 且只使用字段级Updates，已发布题目快照不受源题修改影响。
- 00401页面新增编辑按钮和弹窗：身份字段只读展示，可编辑题干/考察点/方向/状态/备注；保存按钮loading防重复，成功关闭/通知/刷新，失败保留数据重试。后端RED因契约/函数全缺失编译失败，前端RED 2项失败；实现后后端专项与前端4项专项均GREEN。
- [00401题目编辑/启停 staging完成 - 2026-07-25] 部署前备份 `/opt/talent-assessment/backups/element_before_competency_20260725_181343.sql.gz`，12,130,321 bytes，SHA-256 `397d54c730917c72d08c180c5bab9331c3db1b41ef277575e20a44705503cf4e`；应用备份 `server.bak.20260725_181415` / `dist.bak.20260725_181415`。部署哈希：后端 `d241326e764f9532ffd5e4ebb96f30e0ec80ecbc3c432cf5bcb599ffa40f45c1`，前端index `040449905873db53eb720714aab50d968c862ab280027a23ec149ddd84947c41`。
- staging临时源题API验证：题干/考察点/方向/状态/备注更新成功；客户端同时伪造questionCode/dimensionId/dimensionItemNo时，题号/维度/维度内题号/create_time保持不变；空字符串状态被拒且数据不变；legacy题ID被专用更新接口拒绝。临时题清理后 `cleanup_remaining=0`。
- staging浏览器验证：00401首行“编辑”按钮打开弹窗，身份区显示 `D01-Q01 / D01 沟通表达 / 1`，题干、考察点、正反向、启停、备注均正确回填，保存/取消按钮可见；未保存现有正式题。终验 `qedit_temp=0`、service/nginx active、health正常、最近15分钟关键错误0，短时Cookie/Redis会话/令牌/剪贴板已清理；production未修改。
- [00401导入UI本地完成 - 2026-07-25] 页面新增专用模板下载和导入弹窗：仅接受0-10MiB xlsx；选择/移除/变更文件会清除旧预览与SHA-256；预览展示成功/错误总数、全部错误消息或前20条规范化成功行；仅0错误且有digest时允许确认导入。正式导入复用同一File并传expectedHash，成功后刷新题目与维度题数；失败保留文件/预览重试，预览和导入均有loading防重复。
- 导入UI专项先RED失败3项（缺文件边界、预览、确认导入方法），实现后00401页面7项专项全GREEN；后端专用模板/预览/哈希复核/事务导入沿用已通过的测试和staging验证。
- [FB-064 - 2026-07-25] staging浏览器用系统下载模板生成唯一临时题，预览真实返回“可导入1行、错误1行”；错误行是模板第2行填写说明。根因：模板生成三行（表头/说明/示例），校验器从第2行起全部当题目。新增RED稳定复现后，模板生成与校验共享 `CompetencyImportInstructions`，且只精确跳过系统模板第2行；普通错误行不会被宽松忽略。专项service+handler测试GREEN，待重新部署staging后完成正式导入和清理验收。
- [00401导入UI staging完成 - 2026-07-25] 部署前数据库备份 `/opt/talent-assessment/backups/element_before_competency_import_ui_20260725_182806.sql.gz`，SHA-256 `e268cecc6b05a792621469ec6388b29fe1751e2d05a309a958baf4ac953ff646`。前端index SHA-256 `12031d4b2ae7ec75a258102f8092cf5276b3dad99e94a6c50705bc5039d1030a`；FB-064后端SHA-256 `750c7880789525d752498c2ff07f54c79614e8e972ddffeab734ab7a2e715ae6`。
- staging浏览器验收：00401显示“下载模板/导入题目”；模板真实下载HTTP 200、xlsx MIME、6512 bytes；临时唯一题 `UI-TEST-20260725-1835` 修复前预览复现说明行错误，修复部署后为成功1/错误0并启用确认；正式导入后列表384→385。按查询所得主键删除后列表恢复384，数据库胜任力题384、D01启用8、临时题0、临时关联0；backend/nginx active、health OK、最近10分钟panic/fatal/segmentation/import-failed均0。
- 两个短时Redis管理员会话、浏览器Cookie、本地临时模板/xlsx/部署包均已清理为0；production未修改。
- [验证差异 - 2026-07-25] 前端全量Vitest 12文件87项通过，前端production build和Go build通过；胜任力导入service+handler专项通过。Go全量当前被既有源码字符串测试 `TestBugFB060_RepositoryPagingCapsSizeAndChecksErrors` 阻塞：实现已改为00401虚拟行所需的 `physicalTotal/physicalRows`，测试仍只接受旧片段 `q.Count(&total).Error` / `Find(&rows).Error`。本次导入功能未修改该陈旧测试，需作为独立测试维护变更处理。
- [胜任力维度维护本地完成 - 2026-07-25] 新增管理员专用维度维护页和 `POST /exam/api/competency/dimensions/update`。稳定ID、编号和create_time不可改；可维护名称、VIRD层级、适用类别、核心含义、显示顺序和启停状态。后端使用独立request、UTF-8字符长度、固定VIRD/类别白名单、顺序1-48和状态0/1校验，字段级Updates，不使用Save或客户端model覆盖。
- 显示顺序调整采用单事务原子交换：目标位置已有维度时先放入源顺序对应的唯一负数临时位，再更新当前维度和原占位维度，避免 `uk_competency_dimension_order` 冲突；名称冲突和MySQL 1062返回受控提示。已发布测评继续读取维度/题目快照，主数据编辑只影响未来保存或发布。
- `competency_002_dimensions.sql` 的重复执行策略已由覆盖名称/VIRD/类别/核心含义/顺序改为已有ID no-op，防止后续幂等重跑抹掉管理员维护结果；新库仍完整初始化48条默认维度。
- 00401页面新增“维度维护”入口；维护页显示编号、顺序、名称、VIRD、类别、核心含义、启用题数和状态，支持关键词/VIRD/状态筛选、默认20行分页、移动端全屏编辑。状态变化必须二次确认并说明仅影响未来测评；保存失败保留表单重试。
- RED证据：后端因维度更新request/校验器不存在而编译失败，随后固定分类和顺序交换测试分别命中缺口；前端因维度维护页面不存在而suite失败。GREEN：后端维度专项通过，前端维度维护4项通过；前端全量13文件91项、Go全量、Windows Go Build和production build均通过。FB-060陈旧源码断言已仅同步局部变量名 `physicalTotal/physicalRows`，恢复全量测试，不改变业务行为。
- [胜任力维度维护 staging完成 - 2026-07-25] 部署前数据库备份 `/opt/talent-assessment/backups/element_before_competency_dimension_maintenance_20260725_190525.sql.gz`，SHA-256 `d8e2cccb7329cfb53a1cd5ff9a0a77e92965a73bec3606ee2d7c6ae9cb1b796b`。部署后后端SHA-256 `13c043d43cd52910b0c880f1df814603a11d8d0ddf4623952265c4c242b59453`，前端index SHA-256 `5f46301fc1dacae38bc79aee9cd3c9e9176593a1236a02508eab16cca5f063e5`，与本地产物一致。
- staging API真实验证：D01临时修改名称/核心含义/状态并将顺序1→2，D02原顺序2原子交换为1；伪造code/create_time未生效。重复D02名称和空字符串status均被拒；重跑新的002迁移后临时名称/顺序/状态保持，证明管理员维护不被幂等seed覆盖。finally调用正式API恢复D01/D02原值和顺序。
- staging浏览器验证：00401页面“维度维护”入口可达；页面总数48、默认20行，D01/D02字段和各8道启用题正确；关键词D42只返回“权力动机 / Drive内驱力 / 管理通用 / 8 / 启用”；编辑弹窗回填编号、名称、顺序、VIRD、类别、核心含义和状态，切换停用后出现“只影响未来创建或发布，历史不受影响”的二次确认，取消后未写数据。
- 终验：维度总数/唯一code/唯一name/唯一order=`48|48|48|48`，顺序1-48、停用数0；D01/D02/D42恢复为 `1/2/42 + status=0`，临时名称/含义0。短时Redis会话、浏览器Cookie、本地/远端临时文件均清理为0；talent-assessment/nginx active、health正常、最近15分钟panic/fatal/duplicate/unknown-column均0；production未修改。
- [胜任力动态导出本地完成 - 2026-07-25] 新增独立 `competency_export.go`，两个既有入口 `export-raw-data` / `export-raw-answers` 在显式 `competency+competency_average` 时调用同一个三Sheet构建器；legacy/001/002/003继续原模板/宽表逻辑。测评列表对胜任力同时显示“胜任力结果、导出汇总、导出原始答题”，两种导出均明确提示包含结果汇总、逐题明细和题目字典并使用同一文件名。
- `结果汇总`只读已保存整体/维度结果，动态维度列按发布快照display_order，包含人员、开始/完成、用时、完成率、完整性、整体分/均值/等级/报告对象/计分版本；不重新计算成绩。`逐题明细`读取个人固化题序、发布题干/维度/考察点/方向/选项快照及paper原始值/最终分；原始选择文本优先从options_snapshot解析。`题目字典`按snapshot_order输出一次发布题目完整元数据，不受个人随机题序影响。
- 动态工作簿支持48维度后超过Z列的列宽设置、冻结首行、空结果仍输出三张合法表头。非超级管理员手机号沿用既有脱敏策略；汇总入口对胜任力补齐 `*:*:* / exam:list / exam:export` 权限检查，查询任一步失败均在写响应前中止。
- RED证据：后端缺工作簿数据结构/构建器导致编译失败，前端2项因胜任力导出被v-else隐藏且缺专用文案/文件名失败。GREEN：动态导出后端3项、前端入口2项通过；Go全量和Windows Build通过，前端全量14文件93项和production build通过（保留既有3个warning），相关文件诊断0。待部署staging后使用真实三人、16题、2维度结果核对两个端点的三Sheet内容与持久化成绩一致。
- [staging阻塞 - 2026-07-25] 动态导出部署前备份连续3次通过 `liming + ~/.ssh/vm-ubuntu-go-dev_key.pem + ConnectTimeout=10` 连接 `20.200.136.133:22` 均超时，远端备份命令未启动。遵守“先备份再部署”，未上传、未替换应用、未创建临时验收数据；production未修改。staging公网 `/prod-api/health` 仍返回ok。
- 待部署产物已准备：Linux后端SHA-256 `9e3281126bada6a6481e9ff820e979b90fa72f96dc510b77a9019808538d9ba8`，前端index SHA-256 `5c95171c55c10989c4a2db05768a53425740a76638d6c85533f3c02fbe8b0ced`。SSH恢复后固定顺序：数据库备份→部署前后端→运行三人/16题/2维度临时结果链→下载并解析两个入口的三Sheet→核对2/6/8整体分和D01 1/4/5→整链清理与健康检查。
- [纠正 - 2026-07-25 22:48] 上述动态导出staging阻塞已解除。部署前备份 `/opt/talent-assessment/backups/element_before_competency_export_20260725_224808.sql.gz`，SHA-256 `f02b01f1855093ba9a927f7e86e946f3a3e263096bd9d7d03dad246bde8a7e63`。部署后后端/前端index SHA-256分别为 `9e3281126bada6a6481e9ff820e979b90fa72f96dc510b77a9019808538d9ba8` / `5c95171c55c10989c4a2db05768a53425740a76638d6c85533f3c02fbe8b0ced`，与本地产物一致。
- staging真实动态导出验收：临时测评 `RESULT-SORT-9c97b02dfc` 创建3名candidate、3份各16题完整答卷，持久化整体分Low/Mixed/High=`2/6/8`，D01=`1/4/5`，D02=`1/2/3`。两个既有GET端点均返回约12.9KiB xlsx及RFC5987文件名，标准库解析均为 `结果汇总/逐题明细/题目字典`；3条汇总、48条逐题、16条唯一题目字典，动态D01/D02列、完成率100%、原始值/文本/最终分和快照顺序全部正确；两个端点规范化工作簿内容完全一致。
- staging浏览器验收：临时胜任力测评“更多”菜单显示“修改/胜任力结果/导出汇总/导出原始答题”；点击导出汇总弹出明确的“结果汇总、逐题明细和题目字典”确认文案，取消未发下载写操作。
- 真实验收后通过正式测评整链删除清理，`cleanup_remaining=0`；临时exam/result=0、结果孤儿=0、胜任力源题仍384。短时Redis会话、远端xlsx/状态/验证器/部署文件和本地部署包均为0；talent-assessment/nginx active、health正常、最近20分钟panic/fatal/export-failed/unknown-column均0；production未修改。
- [到期Worker批量优化本地完成 - 2026-07-25] Worker启动后在context未取消时立即执行一次扫描，不再等待首个默认30秒周期；后续仍按ticker周期扫描并随context取消退出。过期试卷主查询现通过candidate/tester LEFT JOIN和CASE一次解析participant_type，同时严格限定competency+competency_average，删除原每份试卷1-2次COUNT归属查询；批次100时归属解析数据库往返由最多201次降为1次（提交事务本身不变）。
- Worker扫描按 `limit_time,id` 稳定排序并保留配置LIMIT；归属缺失或单份Submit失败记录paperId后继续，失败记录保持进行中并在下轮重试。RED证据命中缺candidate/tester批量JOIN、loop内N+1和启动首扫顺序；GREEN后Worker专项与100轮×384题安全随机容量测试通过，每轮384题无缺失/重复且产生多个不同排列。前端/Worker并发提交、部分/零作答真实持久化仍待staging验证。
- [到期Worker staging完成 - 2026-07-25] 部署前备份 `/opt/talent-assessment/backups/element_before_competency_worker_20260725_225938.sql.gz`，SHA-256 `dd851f660f222800a56af91951b238503419ff3c1dde8416ba155dc62c2b4b91`。部署后Linux后端SHA-256 `cb74abd36bb072130c86101ee89d34d8ed75a7fa94f5dda5059120513f5e1615`，与本地产物一致；前端未变更。
- staging真实并发/到期验收：同一份已答完且已过期试卷同时发起2次手工提交（服务端转可信timeout），最终仅1条整体结果+2条维度结果，两次调用均返回完成。随后创建candidate部分作答和tester零作答两份试卷，将limit_time置过去并重启服务；两份均在15秒内被启动首扫提交，短于默认30秒周期。
- candidate部分答卷持久化为 `answered=1/16, effective_dimensions=1, overall=3, is_complete=0, submit_type=timeout`，2个维度中1个score=NULL；tester零答卷为 `0/16, effective_dimensions=0, overall=0, incomplete timeout`，2个维度score均NULL，证明未答题未补最低分。测试使用正式整链删除，`cleanup_remaining=0`。
- 终验：临时exam/result=0、结果孤儿=0、进行中且已过期胜任力试卷=0、源题384；远端验证器/二进制临时文件和短时Redis会话清理为0。talent-assessment/nginx active、health正常、最近20分钟panic/fatal/worker-failed/owner-missing均0；production未修改。
- [FB-065 - 2026-07-25] 首次48维度×384题×100份试卷容量链在10并发创建candidate时命中MySQL 1062：`Candidate.Save` 使用 `time.Now().UnixMilli()` 作为主键，同毫秒请求会生成重复ID。验证器finally已整链清理，`cleanup_remaining=0`。新增RED同时命中缺nextID和仍使用UnixMilli；修复为项目既有原子 `nextID()` 后GREEN。该一行修复仅改变新candidate主键生成，不改变请求/响应或现有记录。
- [FB-065 staging完成 - 2026-07-25] 修复后二进制SHA-256 `5d96b345d8fe7020340b3b96cc2fefc860d856a25eada05ae82134a2bb3a6eb5` 已部署staging；并发容量链未再出现Duplicate entry。48×384×100容量、负例/越权/快照、传统001/002/003 smoke全部完成，最终临时测评=0、candidate/paper/result孤儿=0、源题384、服务/nginx active、health正常、近期panic/fatal/duplicate/unknown-column均0。
- [剩余待办权威清单 - 2026-07-25] 当前立即阻塞项只有：FB-065 已本地GREEN但尚未部署staging；因此48维度×384题×100份真实容量链尚未取得最终通过证据。首次失败数据已整链清理，部署前容量备份为 `/opt/talent-assessment/backups/element_before_competency_capacity_20260725_230644.sql.gz`，SHA-256 `a8bd421b5ed3ca22b15f5c7eb334c8c2bfef180de554ba0306206ea5ed091a95`；待部署后原样重跑容量验证器。
- [纠正 - 2026-07-25] 上述容量阻塞已解除。FB-065部署后后端SHA-256 `5d96b345d8fe7020340b3b96cc2fefc860d856a25eada05ae82134a2bb3a6eb5`。staging全量容量链发布48维度/384快照耗时0.098s；10并发创建100名candidate、100份试卷和38,400试卷题，总耗时5.190s，单链p50/p95/max=`0.479/0.758/0.837s`；100份持久化题序hash全部不同，第二次读取100份题序全部稳定。详情刷新p50/p95/max=`0.063/0.105/0.125s`。每份题数最小/最大均384，所有行绑定exam_question_id。整链删除后`cleanup_remaining=0`。这些是当前staging观测值，不声明生产SLA。
- 容量链之后仍需完成的staging验收：参与者非法值/外来题/已结束/已过期保存负例矩阵；发布后修改或停用源题不影响已有快照/历史试卷；未登录、错误purpose、跨participant/exam/paper绑定和管理端越权矩阵；001/002/003完整API与浏览器回归；数据库备份恢复/回滚演练。以上不需要新增业务功能，属于阶段6验收债务。
- [阶段6负例/安全/快照完成 - 2026-07-25] staging真实拒绝raw=0、外来paperQuestion、完成后继续写；过期保存由服务端触发可信timeout。安全矩阵拒绝缺token、paper token冒充participant purpose、跨paper、跨participant和未认证管理结果访问。发布后将D01-Q01源题内容/考察点/方向/状态临时修改，发布快照及历史试卷题干保持完全不变，随后源题按原字段恢复。临时测评整链删除，`cleanup_remaining=0`。因此上述阶段6债务已缩减为001/002/003完整回归和数据库恢复/回滚演练。
- [传统001/002/003 staging回归 - 2026-07-25] 分别复用00101/00201/00301现有题库创建临时legacy测评和2题试卷，真实执行create-paper、paper-detail、qu-detail、fill-answer、hand-exam、paper-result、stand-score；三类均通过，证明新增显式分流未误拦传统链。随后先删paper再删exam，`cleanup_remaining=0`。阶段6业务链剩余阻塞仅数据库恢复/回滚演练（破坏性操作需单独确认）。
- 已明确延期、不应在当前无客户材料时实施：48维度×4等级正式文案、基层/领导总体评价与发展建议、`content_version`、正式报告实例表及正式胜任力PDF。当前仅有测试报告布局和报告数据，`reportTextReady=false`；通用GenerateReport仍不支持胜任力。正式题库由客户提供，现有384题继续仅作系统测试。
- 数据库/历史审计技术债：结果未冻结participant_type/participant_id/报告基本信息；模型DECIMAL仍用float64；无外键；缺建议的 `(exam_id,snapshot_order)` 唯一约束和更贴合提交时间查询的索引。实施前需单独设计迁移和兼容策略，不与容量验收混做。
- [数据库技术债只读实证 - 2026-07-25] staging四张胜任力关系/结果表外键数=0；`el_competency_result`身份快照字段数=0；三项成绩列均为DECIMAL(18,6)，但Go模型仍用float64；结果表仅有PRIMARY(paper_id)和 `(exam_id,overall_score,paper_id)`；当前 `(exam_id,snapshot_order)` 重复组=0。现状数据一致，但约束/历史身份/精确类型仍需专门迁移设计。
- 项目治理待办：当前胜任力累计改动尚未形成可追踪Git提交；`docs/coverage-history.md` 与规则要求的 `docs/business-chains.md` 仍不存在；`business-branches.md` 早期仍有已被后续staging证据取代的⚠️条目，需单独做账本收口。production仍未部署，必须等待用户明确“部署生产/上线”。
- [治理收口 - 2026-07-25] 已新增 `docs/business-chains.md` 作为传统/胜任力端到端链索引，新增 `docs/coverage-history.md` 记录Go/前端测试与staging规模快照；已用真实staging证据关闭legacy API、参与者负例、快照不可变和容量链旧盲区。仍未创建Git提交，production未部署。
- [仍需用户确认的破坏性步骤 - 2026-07-25] 数据库恢复/回滚演练需要创建临时schema、将备份恢复进去、校验后DROP临时schema；根据破坏性操作纪律必须在执行前单独获得确认。正式报告/PDF继续受客户文案缺失阻塞，不属于可自行完成项。
- [数据库恢复演练完成 - 2026-07-26] 经用户授权，将备份 `element_before_competency_capacity_20260725_230644.sql.gz` 恢复到临时schema `element_restore_verify`，耗时59秒。恢复库/当前库表数均51、维度48、胜任力源题384、运行时快照/结果均0，计数一致；随后DROP临时schema并确认不存在，现有`element`未被覆盖。
- [数据库加固005 staging完成 - 2026-07-26] 新增幂等迁移 `competency_005_hardening.sql`：结果冻结participant type/id/name/telephone/age/gender/affiliation/post/degree/major；增加 `(exam_id,snapshot_order)` 唯一索引、`(exam_id,submitted_at,paper_id)` 查询索引和11个非多态关系外键。迁移重复执行通过，现有结果身份回填空记录0；外键真实拒绝无效exam关联且测试行残留0。participant_id动态继承candidate.id的utf8mb4_0900_ai_ci，修复首次验证发现的collation冲突。
- Go成绩模型改用 `shopspring/decimal` 对应DECIMAL(18,6)，提交直接从big.Rat构造六位decimal，不再经过float64；JSON保持数值输出，动态导出写六位精确字符串。新提交事务读取人员主数据并冻结完整基本信息。三份已知答卷再次验证overall=`2/6/8`、D01=`1/4/5`、结果排序/详情正常；临时修改candidate姓名和手机号后，结果表仍保留原快照，随后人员原值恢复。最终运行验证清理为0。
- 005最终部署前备份 `/opt/talent-assessment/backups/element_before_competency_identity_full_20260726_000731.sql.gz`，SHA-256 `971d8f48d64eec09df50c8c1b88dce926f370f34d74ac855727586ef54476246`；后端SHA-256 `a61da43a212a308483e600a81be8fcb23caa63b09565c08596dce3c620d943c4`。Go全量测试通过，staging service/health正常，production未修改。
- [剩余外部依赖 - 2026-07-26] 当前可执行开发与staging验收待办已完成。无法自行关闭的仅有：客户正式题库与48维度分级/两类报告文案未交付，因此正式报告实例、content_version和胜任力生产PDF继续延期；production发布必须等待用户明确“部署生产/上线”。Git工作区包含累计后端改动及用户 `.vscode/settings.json` 变更，且仓库 `.gitignore` 明确排除docs/scripts/前端，因此未自动提交或推送，避免把用户编辑和不可追踪迁移拆散。
- [UF-002 / FB-066 开始测评修复 - 2026-07-26] 用户截图显示“00401 ABC 测评”准备页仍使用传统90题指导语，点击开始连续弹出英文 `competency exam is not published`。实际数据ID为exam `1785027744745375431`、candidate `1785027772270618331`；根因是该胜任力测评publish_status=0、快照=0，但OnlinePaging仍公开草稿且准备页未做发布门禁。
- FB-066先RED：后端测试命中在线列表无publish过滤和英文错误；前端测试命中缺canStart/startDisabledReason。修复后未发布胜任力草稿不出现在在线列表；直接准备URL显示中文“尚未发布”并禁用开始；后端sentinel映射为可操作中文；loading期间重复点击直接返回；胜任力指导语按配置维度题数动态显示，不再固定90题。Go全量、前端15文件94项、production build通过（保留既有3个warning）。
- 经用户明确授权，部署前备份 `/opt/talent-assessment/backups/element_before_fb066_start_gate_20260726_091536.sql.gz`，SHA-256 `0b73ebcbe095f92b010a2a9883fe7384f5069f1400473e78d7fe3ca36be13431`；部署后端/前端SHA-256分别为 `d719e24da3bf983fa693692d219fda420d70c6d65b72377a8d8d80f8e1ebadc7` / `00560297289ddef4e48b08f9a6f6fa53f2b8ce5033d2097ca9f1813be64856aa`。00401 ABC已发布并冻结D01/D02/D05/D06/D32共5维度40题，发布前各维度均启用且各8题；发布后OnlinePaging返回published=1。浏览器准备页显示59分钟、40题五级量表指导语、开始按钮可用、无英文通知。验证未提前创建试卷，paper数仍0，用户点击后才开始个人倒计时。production未修改。
- [胜任力答题页视觉优化 - 2026-07-26] 本地重构 `competencyExam.vue` 的视觉层级，不改答题、保存、计时或提交接口：顶部改为题号/进度/已答未答/倒计时状态卡；五级选项在桌面端由纵向窄列改为5列等宽卡片，移动端回落单列；操作区只保留1个实心主CTA“交卷”，下一题为描边次操作；题号导航增加当前/已答/未答图例、44px移动触控区、键盘焦点、aria-label/aria-current和保存状态aria-live。桌面内容宽度由960px提升为1180px，选项有效横向利用率由截图约15%提升到约100%，页面颜色控制为主色/成功色/灰阶3类语义色，图标统一Element UI一套。
- 视觉优化验证：新增组件结构与无障碍回归测试，首次RED为缺 `.exam-kicker`；实现后前端15文件95项全部通过，目标文件编辑器诊断0；`npm run build:prod`成功，保留既有3个warning（paper页面缺listCaptures导出、asset/entrypoint体积）。本轮只完成本地代码和构建，未部署staging或production；真实浏览器新版截图需等待用户确认统一发布staging后验收。
- [移动端兼容验收 - 2026-07-26] 新增 `scripts/test/competency-mobile-ui-test.js`，通过Playwright真实Chromium拦截只读试卷详情，在390×844手机和768×1024平板视口验证响应式布局。手机：viewport/client/scroll宽均390，无横向溢出，五级选项1列、题号5列、选项48px、题号44px、操作区370px；平板：宽均768，无溢出，选项3列、题号13列、选项52px、题号40px、操作区720px。根据首张真实截图进一步将手机操作区调整为“上一题/下一题”同排、“第一道未答”次行、交卷独立主CTA，复测 `COMPETENCY_MOBILE_UI_PASS`。最终前端15文件95项通过、production build成功、编辑器诊断0；截图产物在ignored的 `scripts/test/screenshots/competency-mobile-390.png`。仍未部署staging/production。
- [答题页UI与移动端适配staging部署 - 2026-07-26] 用户明确选择部署到 `20.200.136.133` staging。部署前前端备份为 `/opt/talent-assessment/dist.bak.competency_mobile_ui.20260726_100334`，备份index SHA-256 `00560297289ddef4e48b08f9a6f6fa53f2b8ce5033d2097ca9f1813be64856aa`。新dist共401文件/11,675,995 bytes，归档SHA-256 `b5c4a2262e73c4250d54b67232b501169f5a61ae4429f075a3f98c00c8f26e55`；原子目录切换、root:root和a+rX权限、nginx -t/reload完成，部署后本地/远端/公网原始index SHA-256均为 `e43daefeda97c5869c948fd002d3221190a4a86d5595f14b277f5026ac53d9fc`。
- staging真实静态资源经Chromium三视口验收：390×844手机无横向溢出、1列选项/5列题号、48px选项/44px题号；768×1024平板无溢出、3列选项/13列题号；1440×900桌面无溢出、5列选项/21列题号，结果 `COMPETENCY_MOBILE_UI_PASS`。真实准备页仍显示00401 ABC、59分钟、40题且无加载错误；点击时原参与者短时认证已失效，系统正确提示重新登录/填写信息，因此未改动答案或试卷。验收试卷详情仅浏览器侧route mock，不写staging数据库。production未修改。
- [真实试卷状态补充 - 2026-07-26] participant `1785027772270618331` 实际已在09:20:23创建paper `ff9fb841-c7d2-4eb3-93c9-11813f83e350`，作答8/40；个人limit_time为10:19:23，Worker于10:19:30按可信timeout自动提交。最终paper state=2，result submit_type=timeout、is_complete=0、answered/total=8/40，candidate end_time=10:19:30。该结果符合既定规则：不完整答卷保留管理审计但不得生成正式报告。当前不能继续原试卷；任何重置或重测会涉及现有结果处理，必须取得用户明确选择后执行。
- [用户决定 - 2026-07-26] 对上述到期不完整答卷，用户选择“保留结果，停止本次测评”。未删除或重置试卷/结果，未创建新参与者或新试卷。
- [胜任力答题页E2E - 2026-07-26] 针对已部署staging静态资源新增并执行3条Playwright Chromium流程：`competency-mobile-ui-test.js`覆盖390/768/1440三视口；`competency-answer-flow-e2e.js`覆盖第9题保存、9/31统计、题号状态和刷新恢复；`competency-submit-guard-e2e.js`覆盖缺token不发详情请求、未答交卷定位、补答、单次manual提交、完成页跳转和token清理。最终三个脚本退出码均0，staging health HTTP 200。首次答案统计文本断言及交卷跨重载模拟失败均仅修正测试代码，未改生产业务代码。
- 本轮为保护已到期且用户决定保留的真实8/40答卷，答题API使用浏览器route mock，只加载staging真实静态资源，不写数据库；因此结论限定为胜任力答题页UI E2E通过，不扩张为全系统浏览器回归。完整证据见 `docs/runtime-validation-report.md`。
- [已完成测评与报告检查 - 2026-07-26] staging只读盘点共有1163份传统已完成试卷，其中961份已有pdf_path、202份缺报告；按题库分布：无repo 1/1，00101 339/294，00102 724/592，00201 22/17，00202 23/17，00301 31/17，00302 23/23（格式为已完成/已有报告）。批量生成202份会产生PDF文件和数据库写入，未在无明确范围下执行。
- 00401 ABC当前唯一结果paper `ff9fb841-c7d2-4eb3-93c9-11813f83e350`为state=2、timeout、is_complete=0、8/40，完整结果数0，candidate pdf_path为空/pdf_flag=0，服务器无胜任力报告目录。`FormalReportData`先执行`validateCompetencyFormalReport`，不完整结果返回`ErrCompetencyIncompleteReport`；FB-047专项测试1项通过。用户明确选择“只检查00401，不生成”，因此未调用生成接口、未写数据库、未创建报告文件。
- [00401报告链只读审查 - 2026-07-26] 当前胜任力“报告”只实现完整答卷的测试报告数据和Vue页面：提交时冻结整体/维度结果与人员基本信息；`AdminReportData -> FormalReportData`先拒绝is_complete!=1，再返回结果、维度和逐题审计；页面固定`reportTextReady=false`并显示“正式解读文案待配置”。未实现`el_competency_report_text`、`el_competency_report`、content_version、胜任力PDF落盘/下载或pdf_path更新。
- 通用`POST /exam/api/exam/exam/generate-report`的chromedp路由仅支持repoCode 001/002，003明确转MBTI专用链，其他（包括无传统repoCode的competency/00401）返回“不支持的 repoCode”；因此不能用于00401。另发现测试报告入口参数错误：路由定义要求`/exam/competency/report/:paperId`，`competencyReport.vue`读取`$route.params.paperId`，但`competencyResults.vue#showReport`却通过query传paperId；这会阻断即使完整答卷的测试报告导航。该轮仅检查未改代码；修复入口和实现正式PDF应作为两个独立逻辑变更。
- [剩余待办权威清单 - 2026-07-26 12:05] 当前可立即处理的功能缺陷：①胜任力结果页测试报告入口把paperId放在query，而命名路由要求params；②前端`paper/paper/index.vue`调用并导入不存在的`listCaptures`，production build持续警告，相关截图/抓拍入口存在运行时风险。两项应分别按bug RED→GREEN处理。
- 当前验证/治理债务：①`business-branches.md`仍保留多条已被后续staging证据覆盖的旧⚠️/❌，需逐项核证收口；其中“competency切回legacy”代码已清空报告版本/维度并恢复repo控件，账本状态过时；②仍缺Go语句覆盖率、前端行/分支覆盖率、传统001/002/003完整浏览器回归；③Git有49个可见未提交改动，且docs/scripts/整个前端目录被.gitignore排除，当前部署成果无法由普通git提交完整追踪；`.vscode/settings.json`为用户改动，整理提交时必须排除或单独处理；④`shopspring/decimal`被业务代码直接导入但go.mod仍标为indirect，需单独执行依赖整理并验证。
- 当前外部依赖/授权项：客户正式题库与48维度四等级、基层/领导两类正式文案未交付；在材料确认前继续延期content_version、文案表、报告实例、正式胜任力PDF和下载/重生成审计。production仍未部署，必须等待用户明确上线授权。另有202份传统已完成试卷缺PDF，属于运营批处理待办，需指定范围、备份和分批方案后执行。
- 本轮现状验证：当前Go全量测试通过、Go Build通过、前端15文件95项通过。早期本地DB/SSH/迁移/容量/恢复演练等“未验证/阻塞”记录均已有后续纠正，不再是当前待办；005已完成身份快照、decimal、索引和外键加固，不应重复实施。
- [FB-067 本地修复 - 2026-07-26] 胜任力结果页原先用`query.paperId`跳转命名路由，但路由定义和报告页均要求`params.paperId`。先新增专项测试，修复前真实收到`query`而断言`params`失败；随后仅将`showReport`改为`params: { paperId: row.paperId }`。专项4项、前端全量15文件96项通过，production build成功，相关文件诊断0。构建仍保留既有3个warning（`listCaptures`缺失及asset/entrypoint体积），与本次修复无关。本切片尚未部署staging/production；真实完整00401答卷仍不存在，因此浏览器报告数据链需后续用完整测试答卷验收。
- [剩余待办集中完成 - 2026-07-26] 用户确认：正式胜任力报告先用明确标记的临时测试文案实现；不批量生成202份传统缺失PDF；不部署production；完成后统一部署staging；调整Git追踪策略。FB-068移除无采集端/接口/表/弹窗支撑的“考试截图”死入口，production build warning由3个降为2个。FB-071使Navbar WebSocket仅在`VUE_APP_SOCKET_ENABLED=true`时启动，解决Go环境无/ws路由导致的持续连接错误。
- 新增MySQL 5.7兼容幂等迁移`competency_006_reports.sql`：文案、报告实例、审计三表，paper+content_version唯一实例，4个外键；`temp-v1`共392条（两类audience、8条总体评价、48维度×2 audience×4等级），全部明确“不可作为人才决策依据”。报告服务按冻结audience+总体等级+维度+等级精确匹配，缺失或跨版本/受众不得回退；实例冻结文案和成绩JSON快照。
- 新增管理员PDF生成/下载链：完整答卷门禁→唯一实例→内部token报告页→Chromium PDF→允许目录落盘→SHA-256/大小/状态→candidate/tester兼容pdf字段→生成/复用/重生成/下载审计。下载使用`filepath.Rel`阻止路径逃逸和RFC5987文件名。整链删除新增审计→实例顺序并在事务成功后删除允许目录内PDF。FB-070修复共享PDF打印器固定跳过第2页的问题：competency从`2-`打印，001/002保留`3-`旧行为。
- staging部署前备份：`/opt/talent-assessment/backups/element_before_competency_reports_20260726_131710.sql.gz`，SHA-256 `b43f2bc139d4f946a49a55a2ae7c7dbe324a683ebee024e5eefc694429f95b20`；应用备份时间戳`20260726_131710`。006迁移两次执行后3表/392文案/4外键一致。最终后端SHA-256 `c7277be5e13595cdd2648ab839f339eaa467dd3ce0058a78a4efcfc8d2530657`，前端index `c14fb47237785edbd227fa498d49f876c25f5835127abb4474b00de44f404565`。
- staging完整验证：临时3人×16题结果链通过；Result High生成PDF、下载文件`%PDF`有效、下载SHA与实例一致，pdftotext包含“临时测试报告”和“不可作为人才决策依据”，生成/下载审计各1；管理端浏览器7流程（dashboard、测评列表、00401、题目、结果、测试报告、试卷列表）通过且console无错误；整链清理remaining=0。终验temp-v1=392、报告实例/审计/RESULT-SORT测评/报告文件/孤儿均0，service/nginx/mysql active，health ok，panic/fatal/unknown-column/report-failed均0。
- 覆盖率基线：全仓Go语句10.5%（handler 8.3%、service 33.8%）；前端Vitest配置范围语句/行58.9%、分支93.87%、函数37.5%。最终Go全量和Windows build通过，前端17文件100项通过，production build成功（只余asset/entrypoint体积2个warning）。FB-069通过无BOM重建`header_icons.go`解除全仓coverage插桩失败。
- Git追踪策略已调整：纳入Go、前端源码、docs、.github、SQL/data/db及安全Python/Shell和4个已审查E2E脚本；继续忽略依赖、构建/覆盖率/截图/日志、`.vscode`和可能含硬编码凭据的既有JS/PS1/archive脚本。敏感扫描后仅测试中的固定`Bearer token`命中，为非真实值。本地提交`eaf8c23 feat: complete competency assessment workflow`已创建，未push；用户`.vscode/settings.json`仍单独未提交。
- 当前剩余项仅为外部/授权：客户正式题库和正式文案到位后，用正式content version替换temp-v1并做正式PDF验收；202份传统缺失PDF按用户决定不生成；production按用户决定不部署。当前00401真实8/40不完整答卷继续保留审计，不生成报告。
- [全部可执行盲区收口 - 2026-07-26] `business-branches.md`中的⚠️/❌/🔥已归零：staging临时停用D48后创建测评被拒且零写入，随后恢复status=0；发布D01临时测评后leader→frontline报告版本修改被拒并整链删除；关闭sql.DB注入证明启用题数查询错误向上传播；HTTP测试覆盖0字节与10MiB+1文件；staging临时BEFORE INSERT触发器强制导入失败，API返回已回滚、题号残留0，finally删除触发器。最终D48=启用、触发器/RESULT-SORT/NEG/ROLLBACK题/报告实例/审计均0。
- 新增安全验收器`staging-competency-config-negative-e2e.js`和`staging-competency-import-rollback-e2e.js`，均不包含凭据并通过环境变量读取短时状态/SSH配置。最终Go全量再次通过；staging service/nginx/mysql active、health ok，panic/fatal/unknown-column均0。当前没有可自行执行且已确认范围内的功能或验证待办。
- [全系统待办纠正 - 2026-07-26] 上述“没有可自行执行待办”仅适用于胜任力核心及已确认验收范围；全仓复核发现`router.go`仍暴露56条统一成功占位路由（38条直接stub + 18条`stubGroup`生成）。其中13条是前端实际使用的用户资料/角色授权/注册P1闭环，25条是缓存刷新、在线强退、任务/日志、代码生成等可选管理P2/P3能力，18条user repo/wrong-book通用CRUD未被当前前端调用，宜删除而非实现。占位路由返回成功但不持久化，属于真实功能缺口，已加入`business-branches.md`。
- 当前Git与环境复核：`origin/master...HEAD=0/0`，`shopspring/decimal`已是直接依赖；仅用户`.vscode/settings.json`未提交。staging后端/nginx/mysql active、health ok，部署哈希仍为后端`c7277be5e13595cdd2648ab839f339eaa467dd3ce0058a78a4efcfc8d2530657`、前端`c14fb47237785edbd227fa498d49f876c25f5835127abb4474b00de44f404565`。正式题库/文案、202份传统PDF、production上线仍分别受外部材料或运营授权约束。
- [胜任力需求落地审计 - 2026-07-26] 按V1.2基线逐项核对REQ-001～076（含031A/B、064A，共79项）：68项已实现、7项部分实现、1项未实现、3项因客户正式内容延期。核心测评链PASS；正式生产报告仍需客户题库/文案和内容验收。P1产品缺口集中为：管理结果页缺开始时间/时长、维度详情未显示scoreSum、报告封面缺姓名/测评名且日期取提交时间、阅读说明不完整、基本信息未按requiredFields过滤、维度核心含义/图表缺失、缺文案错误不含具体匹配键；SC-012尚缺同答案双audience PDF实证。完整矩阵见`docs/competency-requirements-audit-20260726.md`。
- staging当前00401已存在两份结果：原8/40 timeout不完整答卷继续保留；新增一份40/40 manual完整答卷，overall=12.875000、evaluationAverage=2.575000、level=average。00401仍为5维度/40快照题、frontline_employee、59分钟；temp-v1=392，快照/结果/报告孤儿均0。
- [REQ-048/049本地完成 - 2026-07-26] 管理结果分页新增`startedAt=p.create_time`并保留`userTime`；结果页展示开始时间、完成时间、答题时长，维度详情展示`scoreSum`。后端与前端专项均先RED后GREEN；Go全量、Windows build、前端17文件101项和production build通过，编辑器诊断0。该切片尚未部署staging/production；下一项为REQ-057/058报告封面与阅读说明。
- [00401测试报告模板完成 - 2026-07-26] 参考`260715适应力测验报告模板初稿V1.1.docx`的绿色封面、阅读说明、个人信息、总体评价、图表和逐维度解读结构，重做胜任力Chromium PDF模板。报告数据新增测评标题、requiredFields、开始/用时、实际渲染时间及发布时冻结的维度核心含义；页面按requiredFields动态显示身份，说明1.00–5.00与不可跨维度组合比较，绘制总体等级刻度、维度柱状图和逐维度圆环。
- [FB-072 staging完成 - 2026-07-26] 首版真实PDF发现运行时模板组件标题未进入PDF文本，固定1040px内容高度造成18页及交替空白；按RED→GREEN改为直接DOM标题、打印高度自适应、850px封面和紧凑维度页。最终00401完整答卷`9f38be0d-5bb8-4bd1-9de2-fc51ec96b2cc`生成9页PDF，614072 bytes，SHA-256`5c7d9a8d3c6a084f0a7f7eea1295000dbf89608506afcf0a955b67ca41297d22`；包含测评名、阅读说明、总体评价、五维图表、核心含义、发展提示和临时免责声明。本地副本为`docs/generated/00401-competency-test-report.pdf`。
- staging本次部署前数据库备份：`/opt/talent-assessment/backups/element_before_competency_20260726_175528.sql.gz`，12,148,739 bytes，SHA-256`badfb9a6c9be97a3faa4a5163db45b3c7fba3d8b66830fdc55bb2bf34a8e069e`；最终后端SHA-256`dc2d4514b89cb20fae8b58813dc8b563e0e17bdc9f7621ac4d313e6917df0e32`，前端index SHA-256`98bbd8afc3f7b783608d7bf085dc66b9e89df029caa365b4c30e37e11dbd907e`。报告实例completed/temp-v1，服务/nginx/mysql/内外health正常，近期关键错误0、短时会话0；production未修改。
- [五套完整测评与报告验收 - 2026-07-26] 按用户确认创建并永久保留5个独立胜任力配置，均采用D01/D02/D05/D06/D32、5维度/40题；报告对象为3份基层员工版+2份领导人员版。5名受测者完成200/200道作答，得到25条维度结果；评价均值依次1/2/3/4/5，整体分依次5/10/15/20/25，覆盖较低/一般/良好/较高四档。5个报告实例及10条成功生成/下载审计完整，孤儿0。
- [FB-073 staging完成 - 2026-07-26] 首次五配置执行发现requiredFields显示姓名/性别/年龄/手机号/单位/岗位时，总体评价页溢出，单份报告变为10页；失败批次已自动整链清理。按RED→GREEN压缩总体页打印行高、间距、圆环和说明区后，5份报告均为A4 9页。45页逐页文本、A4尺寸、空白页和视觉拼版检查全部通过，无裁切/重叠/错位；本地文件和完整证据位于`docs/generated/five-competency-reports/`与`docs/five-competency-reports-verification-20260726.md`。
- 五报告部署前备份：`/opt/talent-assessment/backups/element_before_five_reports_20260726_181237.sql.gz`，12,152,141 bytes，SHA-256`babfba097d39fd6cd83f35135c0e97eb4aa65deb8e779ad3698d4f00662e154c`。最终前端index SHA-256`c826c72dcd3ecb5bf6b71798209979aa6101d3d9c2e02c41ab8a88aedb73dbe3`；service/nginx/mysql及内外health正常，近期关键错误0、短时会话0。Production未修改。
- [00401导入导出staging完成 - 2026-07-26] 专用导入模板由“表头+说明+1示例”调整为“表头+说明+2示例”，D01示例分别覆盖正向/反向，备注明确正式导入前修改或删除。新增管理员题目导出端点和页面按钮，导出全部胜任力源题，九列与导入契约完全一致，稳定按维度顺序/维度内题号/主键排列；空题库仍输出合法表头，查询失败在响应头前返回受控错误。
- 真实staging round-trip：下载4行模板→预览success=2/errors=0→同文件SHA正式导入2题→题目导出386行并核对两示例方向/状态→完整五维40/40答卷通过两个结果导出端点得到相同的3 Sheet（汇总1、明细40、字典40），持久化overall=5、average=1。随后按主键删除两示例，胜任力源题恢复384、短时会话/临时文件0、近期关键错误0、内外health正常。完整记录见`docs/00401-import-export-verification-20260726.md`。
- 本次部署前备份`/opt/talent-assessment/backups/element_before_00401_io_20260726_185406.sql.gz`，12,530,477 bytes，SHA-256`e26ca039c49d12b357025aa1c55dbf4673d2b79093d5a7e8ada24482c70e195c`；最终后端SHA-256`dcc7e3b59b6564a697354288dbb26a83e36be5463a6899183e747b62e0ab5fe5`，前端index SHA-256`b43a905991328291d6068a456e5a75faf3118b695ca2d68b87a456adbf95fdd0`。Go全量与Windows/Linux build通过，前端17文件104项及production build通过；production未修改。
- [REQ-076 / FB-074本地完成 - 2026-07-26] 报告总体或维度文案精确匹配缺失时，错误现完整输出`contentVersion`、冻结`audience`、`dimension`（总体固定overall）和目标`level`，继续禁止跨版本/受众回退。RED测试先证明旧错误仅区分总体/维度，GREEN后专项5项、Go全量和Windows build通过，编辑器诊断0。需求审计更新为75/79已实现、2项部分、2项客户内容延期；当前下一项为SC-012同答案双受众真实PDF严格对比。本切片未部署staging或production。
- [SC-012同答案双受众staging完成 - 2026-07-26] 创建frontline/leader两个仅报告对象不同的五维40题测评，按题号提交完全相同的40/40原始答案。两边持久化overall=14.375、evaluationAverage=2.875、level=average，D01/D02/D05/D06/D32顺序及5项题数/得分合计/维度分/等级完全一致；temp-v1总体2/2、维度10/10均精确命中各自audience匹配行。两份真实PDF均为A4 9页，归一化逐页文本9/9一致；72DPI逐页像素差异最大0.004539，仅来自受众标签和时间文字，第2、4～9页像素完全一致。验收器finally整链删除，exam/result/report=0、结果/报告孤儿0、短时会话和远端脚本0；service/nginx active、health ok、执行窗口关键错误0。SC-012在temp-v1下关闭，客户正式文案到位后需对正式content version复验；production未部署。
- [UF-003 / FB-075本地修复与按钮E2E - 2026-07-26] 截图“不支持的repoCode”根因是测评列表“详情”无视assessmentType，把competency送入依赖001/002/003 repoCode的传统参与者/报告页。本地改为competency详情进入`CompetencyResults`，legacy保持原`ListExamUser`；RED为1失败/2通过，GREEN专项3/3、前端全量18文件107项、production build成功（仅既有2个体积warning）。staging直接进入专用页，使用保留的领导版40题结果真实操作排序指标/方向/维度、5维详情、40题审计、测试报告、PDF生成/下载和返回共9类按钮，全部通过；下载PDF 615336 bytes，传统专属控件隐藏9/9，console无相关错误。短时Redis会话/状态/远端文件已清理。入口分流尚未部署staging，因此截图路径在部署前仍会复现；production未部署。
- [纠正 - 2026-07-26 20:55] 上述“入口分流尚未部署staging”已失效。FB-075前端已定向部署至`20.200.136.133` staging，未部署后端；远端备份`/opt/talent-assessment/dist.bak.fb075.20260726_205341`，部署后本地/远端/公网index SHA-256均为`d9ffb611ccd8de4fc4ae348ba57e58a66a934565ddb7b3ae4383745fbd7abf15`，dist为401文件。真实E2E从测评管理按完整标题查询保留测评并点击主“详情”，确认进入`CompetencyResults`，随后9类按钮全部通过、传统控件隐藏9/9、PDF下载615336 bytes。短时Redis会话、本地/远端状态、辅助脚本和部署包均清理为0；talent-assessment/nginx/mysql active、内外health正常、执行窗口关键错误0；production未部署。
- [UF-004 / FB-076 staging完成 - 2026-07-26] FB-075仅修测评列表主入口，部署前已打开的传统`exam/users`标签页、书签/历史直达和首页最近测评仍可绕过分流。新增传统详情组件加载前类型守卫：先fetch exam detail，competency用replace进入`CompetencyResults`且不加载传统参与者；首页最近测评同步按assessmentType分流。专项RED为2失败/3通过，GREEN 5/5；前端全量18文件109项、production build成功（既有2个体积warning）。staging部署备份`/opt/talent-assessment/dist.bak.fb076.20260726_210610`，本地/远端/公网index SHA-256=`c552833bd2149a3a5ae68f1522e9bee4c2c2b58c4bf26cfa1a80ba6f5fd0a5f1`。真实E2E覆盖截图旧URL和测评管理主详情双入口，专用页9类操作全部通过、传统控件隐藏9/9、传统generate-report调用0、PDF 615336 bytes；部署后相关错误0，短时会话与临时文件清理为0；production未部署。已打开且不刷新的旧标签仍运行内存旧JS，需刷新一次触发新分流。
- [00401结果页传统风格对齐本地完成 - 2026-07-26] 用户确认采用“布局与按钮风格一致、保留胜任力字段”，按钮范围为查询/重置、查看/答题详情/下载、批量生成/下载胜任力PDF。结果页现使用传统详情的顶部筛选+右对齐批量工具栏+表格选择+文本行按钮结构；新增姓名/电话/完整性参数化筛选，COUNT与分页行查询共用条件，非法完成状态拒绝；默认20行。仅完整答卷可选并执行批量报告，逐份调用胜任力报告实例API，显示进度与成功/失败数，查询/翻页清空旧选择。后端专项RED为编译失败/缺3字段，前端RED为3失败/6通过；GREEN后专项Go通过、前端9/9，全量Go通过、Windows build通过、前端18文件112项和production build通过（既有2个体积warning），最终本地index SHA-256=`fd7696c5b56302033e4e70fd7264d6697afd2a01ee19ef737be027f09e587950`。更新后的staging全按钮E2E脚本覆盖查询/重置/完整性/排序/详情/查看/批量生成下载/行下载，但本切片尚未部署staging；production未部署。
- [纠正 - 2026-07-26 22:20] 上述“尚未部署staging”已失效。最新结果筛选后端和传统风格前端已同时部署至`20.200.136.133` staging。部署前数据库备份`/opt/talent-assessment/backups/element_before_00401_results_20260726_221318.sql.gz`为12,530,802 bytes，SHA-256=`866314c71451bf6afac151b07a3bff8e04ef9e5d38a406522989cc33d7e83a34`；应用备份为`server.bak.00401_results.20260726_221318`和`dist.bak.00401_results.20260726_221318`。最终后端SHA-256=`c9adf6df61a12fbb7aab607cfb4727f5f2ff88a866d372f006697e43b507d74d`，本地/远端/公网index SHA-256=`fd7696c5b56302033e4e70fd7264d6697afd2a01ee19ef737be027f09e587950`。真实E2E覆盖旧URL和测评管理主详情、姓名查询、完整状态、重置、整体/维度升降序、5维度/40题详情、查看报告、完整答卷选择、批量生成、批量下载、行下载和返回，全部通过；PDF 615336 bytes，传统专属控件隐藏9/9，传统generate-report调用0。短时会话及本地/远端临时文件清零；service/nginx/mysql active、内外health ok、部署窗口关键错误0；production未部署。
- [FB-077本地修复 - 2026-07-26 22:33] staging只读负例确认未生成报告下载返回HTTP 200、application/json、code=1且非PDF；旧前端Blob拦截直接放行会保存伪PDF。新增下载API专项测试先RED为1失败/1通过，实际错误为Promise返回JSON Blob；修复后只接受`application/pdf`，非PDF通过Blob.text或FileReader解析后端消息并reject，行下载捕获后显示消息，批量链按既有catch计失败且不调用saveAs。专项2/2及结果页9项通过，前端全量19文件114项、production build成功，相关文件诊断0；本地index SHA-256=`543fbdb8ad21c7faa58cff1bca44c48f0ecbe435bfdb437ded5fca515709320c`。本切片尚未部署staging；production未部署。
- [FB-078本地修复 - 2026-07-26 22:40] 胜任力结果分页、逐题详情和管理员报告数据原先只有JWT登录认证，未校验管理员/exam:list/exam:export。RED低权限用户调用分页时直接进入nil service并panic，证明未在查询前拒绝；实现统一`canAccessExamResults`与HTTP守卫后，三个端点均返回403。允许矩阵覆盖user_id=1、*:*:*、exam:list、exam:export，非相关权限拒绝；传统与胜任力PDF生成/下载复用同一权限口径，行为不变。内部报告渲染改为内部token验证后直接调用无公开权限入口的响应helper，未被后台JWT守卫误拦。专项权限/内部token/报告错误分类全部通过，Go全量和Windows build通过，相关文件诊断0。FB-077/078均尚未部署staging；production未部署。
- [FB-079本地修复 - 2026-07-26 22:50] 胜任力内部报告API原先将token放进query，staging安全计数确认nginx access log已有24行`internal/report-data?...token=`历史请求（未输出令牌值）。后端RED使用正确query token直接进入nil service并panic，前端RED显示token仍在params；修复后前端query仅含paperId、token只放`X-Internal-Token`，后端移除query回退并对正确query token返回401。报告页面hash中的`_internal`用于本机Chromium页面状态，不会随HTTP请求发送给nginx；API网络与访问日志不再包含token。专项前后端GREEN，Go全量、Windows build、前端全量19文件115项和production build通过，相关文件诊断0；本地index SHA-256=`4b6f85721b137cec28fb7943e781df81c21015aa150c1f9bd1e628f62e958724`。FB-077～079尚未部署staging；历史日志不会自动消失，部署后需按新执行窗口确认新增query-token日志为0，并根据部署配置确认/轮换内部令牌；production未部署。
- [FB-080本地修复 - 2026-07-26 22:59] 同一paperId的并发报告生成原先可同时执行“查实例→写generating→Chromium渲染→新文件落盘→覆盖pdf_path”，存在唯一键失败、重复重渲染和孤儿PDF。RED测试因缺同卷锁/有界索引而编译失败；实现后单例`CompetencyReportHandler`持有固定64分片mutex，以无分配FNV-1a将paperId稳定映射，锁在参数规范化后、FormalReportData和实例查询前获取，直到审计与最终实例读取后释放。8个同paper并发测试最大临界区并发=1，索引稳定且范围0-63；锁存储固定，不随paper数量增长。普通专项、Go全量和Windows build通过，文件诊断0；`go test -race`因本机CGO关闭提示requires cgo而未执行，不计为代码失败。该方案覆盖当前单进程部署；未来若水平扩展多实例需升级为数据库/分布式互斥。FB-077～080尚未部署staging；production未部署。
- [FB-081本地修复 - 2026-07-26 23:05] 结果页快速切换查询/排序/分页时，先发的慢响应可覆盖后发的新结果；快速打开不同人员详情同样可把旧详情写入当前弹窗。两个延迟Promise测试RED均稳定复现：列表实际变回older、详情实际变回older（2失败/9通过）。实现后列表和详情使用独立单调request sequence，只有最新序号可写rows/total/detail和关闭对应loading；无维度的非法排序也递增序号以使在途请求失效；发送列表请求时复制query快照，避免响应请求读取后续修改的响应式对象。专项11/11、前端全量19文件117项及production build通过，文件诊断0；本地index SHA-256=`a055a9b08951551ea770a8d4fb276bddee77963d06b0e4d77f1d8c9794f22217`。FB-077～081尚未部署staging；production未部署。
- [FB-082本地修复 - 2026-07-26 23:14] 批量生成/下载循环原先每轮读取实时`selectedRows.length`和索引，用户运行中清空或替换选择会提前终止或改处理对象，且成功/失败统计使用变化后的长度。两个延迟Promise测试RED均稳定复现：原选2条但运行中改选后生成/下载都只调用1次（2失败/11通过）。修复后任务启动时先对完整答卷过滤并浅复制为`targetRows`，后续目标、循环长度、进度和统计只读该快照；表格选择仍可变化但不会影响已启动任务。专项13/13、前端全量19文件119项和production build通过，文件诊断0；本地index SHA-256=`f22ba02e9bf143a05dad544a8dfdc66f9cffa2dc04469742094e85aa4a6618a2`。FB-077～082尚未部署staging；production未部署。
- [FB-083～087本地收口 - 2026-07-26 23:29] FB-083分页畸形JSON由旧“examId为空”改为查询前“参数格式错误”；FB-084将报告completed元数据、人员pdf状态和成功审计纳入同一事务，回滚删除新文件；FB-085前端文件名加入paperId且服务端filename*以`%20/%2B`编码；FB-086 E2E隐藏控件统计由错误硬编码9/9改为实际数组动态6/6，并为行下载补PDF MIME断言；FB-087在<768px下标题/工具栏换行、筛选全宽、详情全屏且身份单列。所有项先RED后GREEN；Go全量、Windows/Linux build、前端全量19文件122项、production build和E2E语法均通过，相关文件诊断0。最终待部署后端SHA-256=`fcfaa85819702b8f9ab333e1f4ef834fe4bd858464098f740d7dc1cb29247348`，前端index=`fa12099adef2e656a9d12a47338a7f810c5ea18ac57ae0dfd7952f7523ee3787`。FB-077～087待staging统一部署和真实按钮/安全/移动验收；production未部署。
- [纠正 / FB-077～087 staging完成 - 2026-07-26 23:40] 上述“待staging部署”已失效。部署前数据库备份`element_before_fb077_087_20260726_233123.sql.gz`为12,530,655 bytes、SHA-256=`cf81d677926d7c261741ac5ba771219fbc5d5e58bb89e2182265fafa7988b40a`；应用备份`server.bak.fb077_087.20260726_233123`/`dist.bak.fb077_087.20260726_233123`。本地/远端后端SHA-256=`fcfaa85819702b8f9ab333e1f4ef834fe4bd858464098f740d7dc1cb29247348`，本地/远端/公网index=`fa12099adef2e656a9d12a47338a7f810c5ea18ac57ae0dfd7952f7523ee3787`。增强E2E通过API负例、旧URL/主详情、筛选/排序、5维/40题详情、查看、批量生成下载、行下载PDF MIME、返回和390×844移动布局，传统控件真实动态6/6、PDF 615305 bytes。低权限短时会话调用分页/详情/管理员报告数据为403|403|403；同paper双并发force重生成均completed（4752ms），最终实例1、completed、PDF615305、同paper文件1；成功审计累计31。真实内部token query日志部署前24，部署后总27且新增3条全部是固定假token主动负例，真实token新增0；服务重启自动生成内部token一次完成轮换。短时Redis会话、发布/E2E临时文件清零，service/nginx/mysql active、内外health ok、关键错误0；production未部署。

## 2026-07-27

- [FB-088～095 / 13条管理路由本地完成] 用户资料5条、角色授权7条和公开注册1条已由成功占位路由替换为真实实现；未处理缓存、monitor/tool或通用dead stub。
- 个人资料更新只读取登录上下文用户ID，严格校验昵称、邮箱、手机号和sex，并检查邮箱/手机号重复；密码修改改用PUT JSON正文，校验旧bcrypt和8～72位字母数字新密码，写入bcrypt后以Redis SCAN批量失效该用户全部会话，不使用KEYS或输出token。
- 头像限制2MiB且必须真实解码为JPEG/PNG，使用crypto/rand英文文件名，保存于配置upload.path/profile的允许子目录；DB更新失败删除新文件。`RuoYiSystemHandler`构造器现显式接收cfg。
- 用户/角色授权接口按既有RuoYi权限字符串执行后端检查，同时允许user_id=1、admin角色和`*:*:*`；user_id=1/role_id=1受保护。用户角色、角色用户和数据范围关系使用事务、ID去重、启用对象校验、批量写入和受控错误；授权列表分页上限200且空集合非nil。
- 注册要求`sys.account.registerUser=true`，验证码通过Redis Lua原子读取删除保证一次性；用户名仅2～20位字母数字下划线，密码强度/确认值/唯一性均校验，bcrypt写入普通启用`00`系统用户。启用的`common`角色存在时同事务授权，不存在时只创建用户。
- 新增MySQL 5.7幂等迁移`system_001_user_name_unique.sql`：先列出并阻断重复用户名，再通过information_schema+PREPARE新增`uk_sys_user_user_name`，不删除或修改现有数据。
- RED证据：首次专项因8组新符号全部未定义而build failed；新增密码传输测试随后确认旧实现仍从URL query读取密码并失败。GREEN证据：FB-088～095专项通过，敏感接口真实HTTP 403门禁测试通过，前端Vitest 19文件122项通过，Go全量和Windows Go build通过，相关文件编辑器诊断0。
- [未验证 - 2026-07-27] 本次未部署、未提交；唯一索引迁移未在本地/staging MySQL执行，真实MySQL事务、Redis会话失效、头像文件/DB回滚和完整Vue浏览器链仍需部署前备份后在staging验证。production未修改。
- [FB-096/097 缓存管理闭环本地完成 - 2026-07-27] `ConfigByKey` 使用 `sys_config:` 精确键执行一小时 read-through；Redis读写失败均回退并返回数据库值，记录不存在继续返回空字符串，其他数据库错误返回受控失败。字典读取同样容忍Redis故障、检查DB错误并保持空数组。
- 配置与字典type/data新增、编辑、删除现检查JSON和全部数据库结果；编辑/删除在事务内先读取旧key/type，提交后精确失效旧/新缓存键。两个refresh路由已由真实handler替换，仅接受对应`system:config:edit`/`system:dict:edit`（admin和wildcard兼容），使用分批SCAN反复清除各自前缀，不触碰login/captcha；Redis失败不返回伪成功。
- RED证据：FB-096/097首次专项因`configCacheKey`、`dictCacheKey`、精确失效、SCAN和refresh handler均未定义而build failed。GREEN证据：专项测试通过，覆盖key helper、缓存hit/miss/1h TTL、Redis回退、DB not-found/错误、精确与前缀失效、无KEYS、路由零stub、权限、Redis刷新失败和空字典数组；`go test ./... -v -count=1`与`go build -o bin/server.exe ./cmd/server`均退出码0。尚未部署、未提交。
- [FB-098/099 无真实能力模块退役 - 2026-07-27] 已完整移除假在线/假强退、定时任务与调度日志、固定模拟缓存监控、全部tool/gen，以及18条user/repo与wrong-book生成stub路由；删除`stubGroup`和`internal/handler/stub.go`。四个无消费重复wrapper（`src/api/user/repo.js`、`src/api/user/book.js`、`src/views/user/repo.js`、`src/views/user/book.js`）已删除，`el_user_book`模型/表/数据和sys_job/gen历史表均未改动。
- 操作日志、登录日志继续保留真实只读list路由、API wrapper、页面和菜单；删除/清空/导出按钮、wrapper和后端stub路由已移除。服务监控、配置/字典刷新、表单构建和Swagger保持不变。
- 新增MySQL 5.7幂等事务迁移`system_002_retire_unsupported_modules.sql`，仅用`menu_id`主键显式IN更新审计写按钮1041/1042/1044/1045、109/1046-1048、110/1049-1054、113、115/1055-1060为`status=1, visible=1`；保留审计查询按钮1040/1043，无DELETE/DROP/TRUNCATE，不删除历史业务表。
- RED证据：后端专项列出26条monitor/tool退役路由、18条通用stub路由、4类stub符号和缺失迁移；前端专项3失败/1通过，命中模块文件、隐藏路由、审计写按钮/wrapper仍存在。GREEN证据：专项Go通过并以有效管理员JWT真实请求9个代表端点均返回404；前端专项4/4通过。全量Go测试、Windows Go build、前端Vitest 20文件126项和production build均通过。SQL迁移仅生成并做静态约束测试，按“不部署”要求未在数据库执行；未部署、未提交。
- [FB-100 本地修复 - 2026-07-27] 部署前契约审查发现后端头像返回 `/profile/avatar/...`，但前端登录和上传完成后会错误拼接 `VUE_APP_BASE_API`，形成 `/prod-api/profile/...` 并绕过nginx静态alias。新增专项测试先RED复现，再将两个消费点改为保留后端静态URL；专项1/1、前端全量21文件127项和production build通过，最终前端index SHA-256为`f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136`。
- [本地最终复核 - 2026-07-27] 主代理独立重跑Go全量、Windows build、Linux amd64 build、前端全量和production build，均退出码0；前端仅保留既有asset/entrypoint体积warning。`go mod tidy`已将sqlmock、miniredis、mysql driver和pdfcpu按真实直接导入归入direct require。`ruoyi_system.go`两个Linux指标Scanner均检查`Err()`，相关诊断归零；生产router零Stub注册。
- [staging阻塞 - 2026-07-27] 最终待部署后端SHA-256为`beddeb0dcad7adb30d119410f293d3b6502b8e18c44f4640789dcaf8e0e350cd`；前端包已在ignored tmp目录准备。当前客户端公网IP仍为`20.239.176.250`，但`20.200.136.133:22`连续多次SSH/TCP检查超时，远端预检、数据库备份、system_001/002迁移、应用部署和staging完整验收均尚未开始；公网`/prod-api/health`仍为HTTP 200 `{"status":"ok"}`。需恢复TCP/22后继续，production未部署。
- 新增无凭据staging验收器`scripts/test/staging-administration-e2e.py`，覆盖注册/验证码/登录、资料/密码/双会话失效、头像字节读取、角色授权/数据范围、配置字典缓存、退役路由/菜单和审计只读，并带管理员摘要保护、原配置恢复及主键清理；本地仅完成Python AST语法验证，尚未在staging执行。
- [staging 管理闭环部署完成 - 2026-07-27] TCP/22恢复后完成只读预检；部署前完整数据库备份位于`/opt/talent-assessment/backups/admin_closure_20260727_103957`，`element.sql.gz`为12,182,806字节且gzip校验通过，SHA-256为`6dabd12c356a9ef91e8ecde59e52ec8b890669925684a66f5f63bb31dcee3e81`。旧后端和前端归档哈希分别为`fcfaa85819702b8f9ab333e1f4ef834fe4bd858464098f740d7dc1cb29247348`、`e5a4f0db5f477d404d6a4dbd6edb08b6b4ad0fa4f1c2dcaeecb3a24b6a2b6488`。
- `system_001_user_name_unique.sql`和`system_002_retire_unsupported_modules.sql`均在staging连续执行两轮通过：重复用户名组为0，唯一索引定义为`0|user_name`，23个退役菜单全部`status=1,visible=1`，审计查询按钮1040/1043保持启用。
- 实际nginx将`/profile/` alias到`/data/uploadPath/profile/`，应用生产配置写`/opt/talent-assessment/tmp/uploadPath/profile`；部署时建立受控符号链接将应用目录映射到该nginx目录。管理E2E已验证PNG上传后匿名读取字节完全一致、伪图片拒绝和文件清理。
- [FB-101 staging RED→GREEN - 2026-07-27] 首轮管理E2E在`POST /system/dict/type`返回HTTP 403。根因是JWT中间件将整个`/system/dict/`前缀匿名放行，导致管理handler拿不到`loginUser`；专项测试先RED命中四个管理路由和两个应保留的公开读取路由，再移除宽泛前缀，仅按method+path匿名GET字典type读取和POST batch读取。专项40项和Go全量通过，修复后管理E2E全链GREEN。
- staging最终后端SHA-256为`703f2eb988faebe36d81db64a7ae0c026c5e905d3b82225ab7b9f0f43a8d62c1`，前端index SHA-256为`f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136`；talent-assessment、nginx、mysql均active，本地8092和nginx `/prod-api/health`均返回`{"status":"ok"}`。
- staging管理E2E通过全部10组证据并完成清理：注册门禁/验证码重放、资料当前用户隔离、密码双会话失效、头像、角色替换/列表/单批授权/状态/范围、配置缓存、字典缓存、退役路由404/菜单隐藏/审计只读、admin摘要保护。临时用户、角色、关系、字典、头像、会话均为0，注册配置精确恢复，admin摘要不变。
- staging 00401导入导出E2E通过：模板4行、正反向示例各1、preview成功2/错误0、导入2、题目导出386行、结果导出3个sheet且两个端点内容一致，示例题清理为0。同答案双受众报告E2E通过：40/40答案一致、5维计分事实一致、文案精确匹配2个总体和10个维度、双方均为9页A4、归一化文字9/9一致、最大像素差0.004539，考试/结果/报告和会话清理为0。
- 最终审计：临时用户/角色/字典/导入示例题/管理E2E会话均为0；部署后服务journal无error，最近nginx访问无5xx。根目录旧版7套测试任务因依赖本机127.0.0.1:8092、8089和已不存在的Legacy Redis工具而退出1，未作为staging验证证据；staging专用API/PDF验收和浏览器登录页加载均已通过。production未部署。
- [00401对传统测评兼容性复核 - 2026-07-27] staging现有60个legacy测评、8个competency测评、1461份legacy试卷和7份competency试卷。模式组合非法记录为0；legacy测评关联胜任力维度配置为0，competency测评关联传统题库为0，带`dimension_id`的胜任力源题关联`el_qu_repo`为0；132,014条legacy试卷题的`exam_question_id`均为空，280条competency试卷题均有发布快照ID。
- 使用00101、00201、00301各一份既有完成答卷执行全程只读staging API回归：测评详情、传统试卷详情、传统试卷结果均HTTP 200业务成功；三个动态导出分别生成9,658/9,564/9,056字节有效XLSX；003专用MBTI详情返回48题。传统试卷详情端点请求competency paper返回HTTP 200业务码1，跨类型防护生效。验收仅创建短时Redis管理员会话并已清理为0，未新增、修改或删除测评数据。
- `CompetencyExpiryWorker`查询以`assessment_type=competency AND scoring_mode=competency_average`内连接硬限制；staging仍有298份历史legacy过期进行中试卷，证明胜任力Worker未将其自动提交。复核窗口服务journal无error、nginx最近请求无5xx。结论：当前00401对001/002/003及MBTI核心链未发现行为或数据污染；剩余低风险盲区仅为传统与胜任力大批量导出同时执行时的资源竞争，及未来正式胜任力PDF若接入传统批量下载时的文件分类规则。
- [生产发布范围确认 - 2026-07-27] 用户确认生产初始化48个可维护维度和00401虚拟题库，客户通过UI负责导入正式题目；发布包不包含staging的384道AI测试题。生产开放完整胜任力链，并导入带“临时测试/不可作为人才决策依据”警示的`temp-v1`双受众文案。
- [39.106.61.48生产只读预检 - 2026-07-27] SSH TCP/22和root/Posh-SSH认证可用；主机根分区40G/剩余约12G，内存7.5GiB/可用约5.6GiB。talent-assessment、MySQL 5.7.44、Redis和宝塔Nginx进程运行；Go后端监听8092且内部health正常，Chrome/Chromium/LibreOffice/PDF工具、Noto/WQY中文字体、32+32个MBTI模板和5个导出模板齐全。生产实际unit以root运行`/opt/talent-assessment/server`，APP_ENV=production。
- 生产当前流量存在切换阻塞：旧`:8088`使用`/www/wwwroot/dist`并代理无监听的8091，API为502；Go候选`:8090`使用`/opt/talent-assessment/dist`并代理8092，浏览器登录页、health和captcha正常。目标无端口公网80从当前客户端不可达；服务器UFW已允许80，但云安全组/上游未放通，且80当前不是业务vhost。上线前必须放通云侧TCP/80并将业务vhost指向`/opt/talent-assessment/dist + 8092`，同时建立`/profile/`上传目录映射。
- 生产数据库当前46表、54个传统测评、578题、1865试卷、158104试卷题、349119答案；MBTI表已有4814行。胜任力六个exam字段、六个题目字段、三个试卷题字段、八张表和全部外键均不存在；48维度、胜任力题/测评/结果/报告和temp-v1均不存在。`el_paper_qu`约65.8MiB，004 ALTER须先在MySQL 5.7恢复副本演练。重复用户名组0、唯一索引未建；23个退役菜单ID全部存在且语义与staging一致。MBTI表、el_repo utf8mb4_general_ci及两个优化索引已存在，因此不整包执行旧deploy-39 Schema。
- 生产当前没有`/data/backup`或`/opt/talent-assessment/backups`目录；上线前必须创建受限目录并完成全库+routines+triggers备份、gzip测试和SHA-256，备份失败不得迁移。完整差异矩阵和冒烟样本见`docs/production-release-39.106.61.48-20260727.md`。
- [生产发布候选准备 - 2026-07-27] commit `d743d063f1281632f42a5630bffec08831533802`重新通过Go全量、Linux build、前端21文件127项和production build。ignored发布包位于`Go-based Refactored System/tmp/prod-release-d743d06/`，后端SHA-256=`03397e0faf24a21fb6da4e76ba0776226ab87bf2a3c52f72b475eeed4791e44a`，前端归档=`aaff21f4995048a1a4e0b37b77d5d3a733af2b75759d72f826f0f25104bbd82a`，index=`f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136`；包含001-006和system001/002八个迁移，XLSX/JSON/CSV题目数据文件数为0。当前仅完成只读预检和包准备，尚未备份、迁移、上传或替换生产应用。
- [生产备份与迁移演练完成 - 2026-07-27] 完整备份位于`/opt/talent-assessment/backups/release_00401_20260727_153001`；数据库gzip 14,274,653字节、SHA-256=`283d227047aef5fefaa315f628c1da59ca13b67583f56806579a73528ebee08d`，另含两套dist、后端、Nginx、unit和菜单快照且均为0600。授权的临时Schema恢复耗时13.812秒，第一轮完整迁移20.353秒、第二轮幂等0.172秒；传统核心摘要不变，48维度、8表、15外键、392文案、唯一索引和23菜单全部通过，题目/测评/报告实例为0；临时Schema已DROP为0。
- [00401生产Schema与8090发布完成 - 2026-07-27] 生产主Schema迁移耗时16.109秒，传统核心摘要`54|578|1955|158104|349119|1424|33`前后不变；48维度唯一性完整，00401题目0、胜任力测评0、temp-v1=392、报告实例/审计=0|0、非法模式/孤儿=0。应用后端/index最终SHA-256分别为`03397e0faf24a21fb6da4e76ba0776226ab87bf2a3c52f72b475eeed4791e44a`/`f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136`；8090 vhost补齐profile alias，service、health、captcha和登录页正常。
- 生产只读冒烟通过001/002/003详情、试卷结果和有效XLSX导出（9730/9460/18788字节）、MBTI 48题、48维度题数0、唯一00401虚拟入口题数0、6644字节导入模板、退役路由404和菜单隐藏；短时会话清零，部署后journal无error。客户导入SOP位于`docs/00401-production-import-guide.md`。按用户选择旧8088保持不动，先以`http://39.106.61.48:8090/`验收；目标无端口80仍受云安全组/上游未放通阻塞，尚未切换。
- [生产浏览器验收 - 2026-07-27] 8090管理后台首页和题库管理真实加载；列表唯一00401行显示“胜任力测验题库”、题数0、总题库7。专用页显示维度维护、导出题目、下载模板、导入题目并明确“暂无胜任力题目”；维度维护页显示总数48，D01字段、题数0和启用状态正常。验收仅执行只读导航，未导入、编辑、启停或创建测评。
- [生产15分钟观察 - 2026-07-27] 服务15:58:34启动，16:13:40复核仍active，8092/8090 health均ok；部署后journal关键错误0，精确部署日期/时间窗Nginx新增5xx=0。阶段性数据库终验仍为48维度、胜任力题/测评0、temp-v1=392、报告实例/审计0、非法模式0。生产`/tmp/prod-release-d743d06`目录及一个同名前缀归档仍在，删除需单独破坏性操作确认；正式备份目录必须保留。
- [纠正 - 2026-07-27] 用户随后明确确认删除上述两个临时上传项；`/tmp/prod-release-d743d06`和同名tar.gz残留数为0。正式备份目录仍存在，清理后service active、8090 health ok。
- [生产旧8088停用 - 2026-07-27] 经用户明确确认，从宝塔Nginx主配置移除旧8088 server块；修改前备份为`/opt/talent-assessment/backups/release_00401_20260727_153001/nginx.conf.before-disable-8088.20260727_162121`，SHA-256=`c18bc433bdfc5771133ae31f453896e20cbfe89cfd686fe02cb46918df774511`。配置检查和reload通过，Nginx配置8088引用0，本机/公网8088不可达；公网8090首页200、health ok，服务active。旧`/www/wwwroot/dist`及备份未删除。

## 2026-07-28

- [UF-005 / FB-102 本地修复] 生产00401开放测评二维码白屏根因是胜任力测评不关联物理题库，在线列表`repoCode`为空，而candidate路由要求必填`:repoCode`；二维码URL因此止于stuFlag且Vue不匹配组件。新增统一resolver将competency映射为虚拟code `00401`，开放/封闭二维码和直接导航共用该值，legacy继续保留真实repoCode。
- RED专项3/3失败；GREEN专项3/3、前端全量22文件130项、production build均通过，构建index SHA-256=`a1c2a83d5f480f0277188273a96060a95651631ffeac47c329ddbbf64e60f2c8`。8089本地开发服务器真实打开`/#/my/exam/candidate/local-competency/0/00401`并显示考生信息表单；因本地8092未运行，detail请求返回500，此错误只反映本地后端未启动，不影响路由匹配证据。本切片尚未部署staging或production。
- [VIRD新材料需求分析 - 2026-07-28] 已只读核验`260728VIRD测评方案.docx`、`260728-VIRD维度评价+总体评价.xlsx`和`260728VIRD测验报告示例V2.3.docx/pdf`，完整分析见`docs/00401-vird-requirements-gap-analysis-20260728.md`；本轮未修改Go/Vue/SQL、未写数据库、未部署。
- 新评价表包含48个连续唯一维度、48条完整定义、240条维度1～5分行为描述，以及基层岗位/管理岗位各5档总体诊断和发展建议共10组；与现有主数据仅D03核心含义、D42名称、D43核心含义三项不同。D42新材料写“权利动机”，与此前客户确认并已落地的“权力动机”冲突，未覆盖现有值。
- 新材料候选规则与V1.2存在实质冲突：维度从四档改五档；总体从有效维度均值四档改为总分占满分比例五档；五级选项文字变化；报告改为VIRD品牌、五档仪表、雷达图、完整定义、表现评估及总体发展建议。维度区间存在4个数学空档，总体末档“65%以下”与上一档重叠，均须客户确认后才能编码。
- 新报告样例还存在内容错配：7维度总分24.13/满分35=68.94%，总体标签“合格胜任者”和4条发展建议均匹配基层岗位合格档，但诊断段实际取自工作簿“优秀胜任者”行；正式映射不能照抄样例，需按岗位+等级精确确认。
- 当前00401的48维度、全量取题、发布快照、独立随机、正反向计分、自动提交、双受众、结果/导出、PDF生成下载审计均可复用；主要Gap为新评分版本、正式内容Schema/版本冻结和V2.3报告模板。生产00401正式题目仍为0，新材料不含题目明细。

## 2026-08-09

- [260807基层员工第一期需求与Gap分析] 已只读核验`docs/260807胜任力开发资料/`的开发方案、题本/评价工作簿和DOCX/PDF报告样例；完整分析见`docs/00401-phase1-requirements-gap-analysis-20260809.md`。本轮未修改Go/Vue/SQL，未导入题目，未写数据库，未部署或连接远端环境。
- 260807材料把第一期收缩为固定基层员工10个二级维度，跳过选维度；一级分组为通用能力（逻辑思维、数字应用、计划执行、持续学习、沟通表达）和心理素养（敬业奉献、求真务实、自律性、成就导向、合作意识）。该一期顺序与现有48维度全局顺序不同；成就导向/合作意识与现有D41成就动机/D35合作性不是同名，需确认稳定ID映射。
- 题本经逐行复核为80道维度题+10道效度题=90题。每个二级维度8题；逻辑思维、数字应用、计划执行、持续学习、沟通表达、敬业奉献、求真务实、自律性、合作意识均为6正向+2反向（第7/8题反向），成就导向为8道正向；合计62正向、18反向。效度10题均未填写方向或题目类型，方案所称“伪装维度”也没有独立ID/名称。
- 新候选计分为：二级维度8题平均；两个一级维度各取5个二级维度平均；总体为10个二级维度分相加、满分50；二级/一级区间边界为1.7/2.7/3.5/4.3且原文写“上包含”；效度10题候选总分>35存疑、<=35良好。DOCX末档“65%以下”与上一档重叠，XLSX明确总体档位为45以上、40-45、32.5-40、25-32.5、25以下；边界包含关系和效度题方向仍须客户确认。
- 新报告样例为A4、10个物理页（封面+正文第1～9页），包含阅读说明、个人信息/总体评价/效度提示、两个一级维度、10轴二级雷达图和10个维度定义/得分/五档诊断建议。样例一级维度图将两个独立1～5得分画成近似构成比，存在误导风险；第9页留白较多但无额外空白页。
- 当前00401核心答题链可复用，但一期P0 Gap为固定10维产品配置与顺序、题目类型、一级结果、效度结果、三套五档标签/算法、客户题本导入转换、评分/内容版本发布冻结及正式10页报告。现有四档`competency-v1`、动态维度选择、九列严格导入、`temp-v1`和每维一页模板均不能直接满足260807一期。
- 本轮未重新核验生产实时数据；生产题目0、胜任力测评0、`temp-v1`392条等仅是2026-07-27最后一次已验证基线，不能表述为2026-08-09实时状态。
- [260807一期题本候选转换完成] 新增本地工具`scripts/tools/convert-competency-phase1-workbook.js`，将客户多级表头XLSX确定性转换为`scripts/data/competency-phase1-candidate-20260807.json`。源文件SHA-256=`2debf83ffcbda000a016f4ea591b688afca7dcfd26730f5c963b131e8477e0b9`，候选JSON SHA-256=`153ad2c45e8c30e43ab2c86f33472568fd4fc76d7d082044b2a0553fa0baa162`；`--check`重现校验和独立结构断言均通过。该工具当前命中根目录`.gitignore`的`scripts/tools/*.js`规则，仍是未纳入Git跟踪的本地工具；本轮未改变忽略策略。
- [纠正 - 2026-08-09] 上述转换工具忽略状态已失效。根目录`.gitignore`已增加精确例外`!scripts/tools/convert-competency-phase1-workbook.js`；转换器现为Git可见的未跟踪文件，其他`scripts/tools/*.js`继续保持忽略。候选JSON同为Git可见的未跟踪文件；本轮未暂存、提交或推送。
- 候选JSON已验证为90题（80维度+10效度）、62正向+18反向、10个效度方向空值、8个同名映射+2个待确认映射、50条维度分档文案+5条总体文案，并显式保持`importReady=false`、`databaseWritten=false`。当前有6个阻塞项和3个警告项；未修改Go/Vue/SQL，未导入题目、未写数据库、未连接远端或部署。
- [通用胜任力版本化基础本地完成 - 2026-08-09] 四类默认版本为产品`competency-generic-v1`、评分`competency-v1`、内容`temp-v1`、报告模板`competency-report-v1`。版本标识只允许小写字母、数字、点、下划线和连字符，最长32字符；产品、评分、报告模板必须由当前程序支持，内容版本允许保留合法的未来标识。
- 冻结链已建立：胜任力草稿保存时规范化版本；发布时与题目快照一并冻结且发布后不可修改；交卷时复制到`el_competency_result`；正式报告按结果中的内容版本和模板版本精确读取，不跨版本回退。传统测评清空四个胜任力版本字段。
- 新增幂等迁移`scripts/sql/competency_007_versions.sql`，增加Exam、Result、Report所需版本列，并只对空版本的既有胜任力数据回填当前兼容版本；legacy Exam保持空版本。模型与迁移静态测试已通过。
- 本地验证：Go全量测试通过，Windows Go build通过；前端Vitest 23文件132项通过，production build通过并保留2个既有资源体积warning；真实Gin HTTP测试确认不支持的产品版本在进入数据库事务前被拒绝。
- [未验证 - 2026-08-09] 本地MySQL `127.0.0.1:23306`返回`ECONNREFUSED`，因此007尚未取得真实执行、重复执行和数据回填查询证据。迁移必须在可用数据库环境执行并核对列、胜任力回填及legacy空值后，才能标记完成。
- 本版本化切片未部署staging/production，未暂存、未提交、未推送；未导入260807的90道题，也未实现效度方向、新五档算法或一级维度聚合。
- [纠正 / 通用胜任力版本化基础 staging 完成 - 2026-08-10] 上述 007 与 staging 未验证状态已失效。目标仅为用户确认的 `20.200.136.133` staging；production 未部署。SSH 初始因 TCP/22 超时受阻，放通当前客户端公网 IP `20.239.176.250` 后，以 `liming` 和既有密钥登录 `vm-ubuntu-go-dev` 成功。
- 部署前完整数据库备份为`/opt/talent-assessment/backups/element_before_competency_20260810_100113.sql.gz`，12,219,312 bytes，gzip校验通过，SHA-256=`0f31bf3dea7e67292406f1732f19c982b2ea80624335f6d753e3722cfe30f11c`。
- `competency_007_versions.sql`在staging连续执行两次成功。新增版本列总数8；9个competency exam、10个result和7个report实例版本缺口均为0；60个legacy exam非空版本数为0；业务行数迁移前后保持exam/result/report=`69/10/7`。9个胜任力配置均回填`competency-generic-v1 / competency-v1 / temp-v1 / competency-report-v1`。
- 应用部署备份为`server.bak.20260810_100311`和`dist.bak.20260810_100311`，旧后端/index SHA-256分别为`703f2eb988faebe36d81db64a7ae0c026c5e905d3b82225ab7b9f0f43a8d62c1`和`f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136`。部署后后端/index SHA-256分别为`44df29bcf09a55ab4a1d61b94d408d9eb649cdc40645ea5c48e75753cf063fc7`和`1e60ac06cfb4ff219428151d91b1a6f3231001ff8748ea0bb80228ea34ad5163`，本地、远端和公网index一致；dist为`root:root 0755`。
- staging真实验证：管理员分页返回9/9胜任力配置均为四类兼容版本；真实详情返回同一版本集；不支持产品版本`unsupported-v9`在写库前返回受控错误，临时标题数据库残留0；Chromium真实打开既有胜任力编辑页并显示产品/评分/内容/模板标签及“发布后与题目快照一并冻结”提示。短时Redis会话残留0。
- 最终状态：talent-assessment/nginx/mysql均active，内外health均`{"status":"ok"}`，部署窗口panic/fatal/unknown-column/report-failed计数0；远端发布临时文件已清理，数据库和应用备份保留。此次部署仍不包含260807的90题正式导入、效度规则、五档算法、一级维度结果或10页正式报告模板。
- [260807一期数据结构本地完成 - 2026-08-10] 新增 `scripts/sql/competency_008_phase1_structures.sql`：源题与发布快照增加可空 `competency_question_type`；已有 `dimension_id` 非空源题和既有发布快照兼容回填为 `dimension`；新增测评一级分组快照、维度到分组的可空多对一关联、一级结果、效度结果，以及整体结果的维度题总数/已答数。一级得分/等级、效度得分/状态均保持 NULL 可表达，未写死一期分组名称或维度映射。
- 008 只对冻结版本 `competency-generic-v1 + competency-v1` 的历史结果以现有总题数兼容回填维度题计数，不重算分数；当前九列导入显式写 `dimension`，发布快照复制题型，当前提交继续按原 `competency-v1` 计算并将现有计数同时写入维度题计数。传统题目读写隔离扩展为 `dimension_id IS NOT NULL OR competency_question_type IS NOT NULL`，可防止未来 `dimension_id` 为空的效度题进入传统页面/API。
- 新增表全部使用 RESTRICT 外键并已同步完整链删除顺序：报告 → 一级/效度/二级/整体结果 → 试卷子表/人员/试卷 → 题目快照 → 二级维度快照 → 一级分组快照 → 测评。迁移动态对齐 exam/paper/group 外键列字符集与 collation；原维度题号唯一索引升级为 `(dimension_id, competency_question_type, dimension_item_no)`。
- TDD 证据：初始聚焦 RED 为 64 通过/4 失败，模型单测因三个新模型未定义而编译失败；实现后最终聚焦测试 89/89。Go 全量测试退出码0，Windows与Linux amd64构建退出码均0，`go vet ./...`退出码0，所有改动Go文件编辑器诊断0，scoped `git diff --check`退出码0。
- [未验证 - 2026-08-10] 本机 `mysql/mysqld/docker` 均不可用，WSL未安装，因此008尚未取得真实MySQL 5.7/8首次执行、第二次幂等执行、SQL回填查询和真实HTTP/数据库结果证据。本切片未连接或部署staging/production；未导入260807的90题，未实现效度方向/阈值、一级聚合、新五档、D41/D35映射或10页报告。
- [纠正 / 008 staging 数据库迁移完成 - 2026-08-10] 上述 008 数据库未验证状态已失效。目标仅为用户确认的 `20.200.136.133` staging（`vm-ubuntu-go-dev`）；MySQL 8.0.46、talent-assessment和nginx预检均active，根分区迁移前可用约51.7GiB。迁移前核心计数为exam/qu/question-snapshot/result=`69/1239/344/10`，008目标列/表/外键均为0。
- 完整数据库备份为`/opt/talent-assessment/backups/element_before_competency_20260810_114814.sql.gz`，12,219,416 bytes、0600，gzip校验通过，SHA-256=`0061da2e26bbeba2c4c5bd233efc8dbddb90741d64fb38d6b8cbcf45128dfba2`。本地与远端008 SHA-256一致：`b34ee8d491243fccb0626f2742cc325f7d3d5027bec842b07d32377089edd96b`。
- `competency_008_phase1_structures.sql`在staging连续执行两次成功，第一轮5,248ms、第二轮2,884ms，两轮SQL错误数均0。两轮后均为5个目标列、3张新表、5个新外键；核心计数持续为`69/1239/344/10`，新增分组/一级结果/效度结果/维度分组链接均为0，证明迁移未伪造尚未计算的数据。
- 题型回填证据：源题`dimension/validity/NULL/非法=384/0/855/0`，发布快照=`344/0/0/0`；10条`competency-generic-v1 + competency-v1`历史结果的维度题计数差异为0，实际分布为`40/4`、`40/8`、`40/40`、`104/0`、`104/104`（总题/已答与维度题/已答维度题一一一致），未重算得分。
- 结构终验证据：三张新表列数为`9/15/9`；题号唯一索引精确为`UNIQUE(dimension_id,competency_question_type,dimension_item_no)`且临时替代索引残留0；题型、分组和结果查询索引签名全部正确。5个新外键均为RESTRICT/RESTRICT，原15个胜任力外键保留，总数20；exam/group/paper五组关联列collation全部匹配，五类孤儿计数均为0。
- 迁移后talent-assessment/nginx/mysql均active，内部和公网health均`{"status":"ok"}`，公网登录页可加载；近20分钟应用panic/fatal/unknown-column/missing-table/segmentation计数0。此次仅迁移数据库，没有上传或部署后端/前端，production未修改；仍未导入260807的90题，也未实现效度方向/阈值、一级聚合、新五档、D41/D35映射或10页报告。
- 当前staging `/tmp` 保留本任务创建的迁移SQL及两份执行日志共3个文件，不含凭据；因删除文件属于破坏性清理，未在无单独确认时删除。正式数据库备份必须保留。
- [一期A/B维度身份口径确认 - 2026-08-10] 用户确认总体维度数量本期暂不确定；总体正文“40维”与图片可数“34维”均不得作为完整池结论。一期稳定身份优先并用A/B替换D体系：`A1-01`逻辑思维、`A1-02`数字应用、`A1-03`计划执行、`A1-04`持续学习、`A1-05`沟通表达、`B1-01`敬业奉献、`B1-02`求真务实、`B1-03`自律性、`B1-04`成就导向、`B1-05`合作意识；A层为通用能力、B层为心理素养，适用对象为基层员工。
- 用户确认旧胜任力历史无保留负担，选择“本地改造 + `20.200.136.133` staging全量重置”；传统001/002/003必须保留，production不动。安全顺序固定为：完整数据库/PDF备份和传统摘要 → 停止写流量 → 应用既有整链删除每个胜任力测评并核验PDF路径 → 009预检/锁/一次性替换 → 009第二次no-op → 十维/零D/零旧题文案/传统摘要复核 → 恢复写流量。
- [一期A/B身份本地实现完成 - 2026-08-10] `competency_002_dimensions.sql`的新库种子已改为且仅为上述一期十维；新增staging-only `competency_009_phase1_identity_reset.sql`，要求同会话显式设置staging和写流量已停止变量、获取会话级锁、检查测评/快照/试卷题/答案/错题本/结果/报告依赖为零，再用持久化marker在事务内清胜任力源题关系/答案、源题、旧`temp-v1`文案和退役主数据并插入十维。009不直接删除exam/paper/candidate/tester/result/report实例，也不推断34/40维。
- [SQL兼容性纠正 - 2026-08-10] MySQL 5.7官方可预处理语句列表不包含`SIGNAL`；最初的动态`PREPARE SIGNAL`方案已纠正。009使用MySQL 5.7可预处理的`SELECT JSON_EXTRACT(无效JSON)`作为条件失败门，并以确定性错误标识区分未显式授权、迁移锁失败和残留依赖；移除了mysql2执行器不支持的`DELIMITER`/临时存储过程。
- 维度维护API/UI当前只接受`通用能力/心理素养 + 基层员工 + 顺序1-10`，稳定ID/code继续不可编辑；九列导入说明和示例改为A/B，顺序只要求正整数并必须命中当前维度主数据，不再硬编码1-48，保留未来已确认维度顺序的扩展能力。
- 候选转换器升级为`competency-phase1-candidate-v2 / 1.1.0`，所有维度题/效度关联/评价文案均绑定A/B ID/code，80道维度题使用`A1-01-Q01`类稳定前缀；退役D映射字段和`MAP-001`已移除。源XLSX SHA-256仍为`2debf83ffcbda000a016f4ea591b688afca7dcfd26730f5c963b131e8477e0b9`，新候选SHA-256=`0210adbe649a56c6af64de8ec24f1320d4a2d82fe0803373c766ec16f2c5c732`；90题、80维度/10效度、62正向/18反向、50条维度分档、5条总体文案保持不变。候选仍为`importReady=false/databaseWritten=false`，剩5个阻塞和3个警告。
- TDD证据：RED为Go聚焦71通过/8失败、前端132通过/1失败、候选身份首维失败；GREEN为聚焦Go 79/79、Go全量退出码0、前端Vitest 23文件133项全通过、Windows和Linux amd64构建、`go vet ./...`、前端production build、候选身份与字节级确定性检查全部通过。相关编辑器诊断0，scoped `git diff --check`为0；候选身份脚本已精确解除忽略并加入CI定向作业。
- [未验证 / staging阻塞 - 2026-08-10] 009尚未在真实MySQL 5.7/8执行，首次替换、残留依赖阻断、第二次no-op、传统摘要和报告PDF残留均未取得数据库/文件证据。当前客户端公网IP仍为`20.239.176.250`；`20.200.136.133:22` TCP不可达，使用既有`liming`和密钥只读握手超时，但公网`/prod-api/health`仍返回`{"status":"ok"}`。因此本轮未登录、未备份、未停服务、未删除、未迁移、未部署，production未修改。
- 本身份切片仍不导入90题，不实现效度方向/阈值、一级聚合、新五档、固定一期产品选择限制或10页正式报告。
- [纠正 / 一期A/B身份 staging 重置完成 - 2026-08-10] 上述 SSH 与009未验证状态已失效。`20.200.136.133` staging（MySQL 8.0.46）完整备份位于`/opt/talent-assessment/backups/phase1_ab_reset_20260810_141544`，目录0700；数据库归档SHA-256=`50da89e91d7559fe589bae3544bcbf3f4b0f6c06577ecaaec069e965f0ea303c`，7个报告PDF归档SHA-256=`82c24632fe9dff91c0a619a313ea5a1095ce58d240d05c6df255fac0ca50091e`，备份清单复核通过。
- 重置前有9个胜任力测评、384道源题、48个D维度、392条旧报告文案、62条二级快照、344条题目快照、528条胜任力试卷题、66条二级结果、10条整体结果、7个报告实例和756条审计。停止Nginx写流量后，通过既有整链删除API一次删除9个测评；全部运行/报告依赖归零，7个原PDF路径和胜任力报告目录PDF均归零，临时Redis管理员会话归零。
- 009脚本上传SHA-256=`428e0fae36fe8b274bb5d07c08dde7298b6f51652f64faded5247e1dfa2441e1`。同一MySQL会话显式设置staging与停写授权后，首次执行`apply_reset=1`，得到marker=1、十个A/B身份逐字段匹配、D=0、源题=0、报告文案=0、运行依赖=0；第二次执行`apply_reset=0`，marker与十维完整行签名不变。十组传统数据签名在整链删除、009和部署后均与基线完全一致。
- staging部署产物与本地一致：后端SHA-256=`2018167ca515fc9c5f2ac957e9e75720fc2727febf6e15e2c9595c1b845655df`，前端归档SHA-256=`fdb120e61ddb667e25578889e3f483e2410fa6fbf68c69a3472a29057c8e8f03`，index SHA-256=`da918edb9307c03202e925839b8d97e4f3e7653eb58fc6e33c6564b56b3ecef5`；dist为root:root 0755，talent-assessment/nginx/mysql均active，内外health正常，部署窗口后端关键错误0。
- 真实Nginx API验证维度10、A/B code 10、每维题数0，00401题库total/records均0；真实Chromium验证维度维护表10行、题库0行、维度下拉10项，API失败、console error和page error均0。所有reset/acceptance/UI临时Redis会话与本地令牌状态文件已清理为0；production未修改。
- [验证边界 - 2026-08-10] `competency_002_dimensions.sql`尚未在全新真实Schema执行；009缺显式授权和残留依赖两类失败门已有静态测试但未做真实数据库负向执行。远端YAML为CRLF，安全解析配置键前必须先移除`\r`；否则会把存在的JWT配置误判为空。
- [00401 一期 P0/P1 决策闭环 - 2026-08-10] 用户确认：2026-08-07一期方案完全覆盖00401旧胜任力产品逻辑，覆盖范围不含传统001/002/003；产品固定为基层员工、固定`frontline_employee`和十个A/B维度，隐藏维度/报告对象选择并由后端强制校验。其余P0和P1全部采用建议口径，P2继续延期。权威记录为`docs/00401-phase1-customer-decisions-20260810.md`。
- 一期固定90题全部启用和必答并统一随机混排；80道维度题按原正反向规则计分，B1-04八题全部正向；10道效度题全部标记`validity`并按原始1～5正向累计，不建立伪装维度。效度`<=35`良好、`>35`存疑、未答完为未完成，且不影响二级/一级/总体得分。
- 二级/一级采用上包含五档边界`1.7/2.7/3.5/4.3`，一级L3统一“中分”；总体按十个二级维度分之和使用`>=45 / >=40 / >=32.5 / >=25 / <25`五档。默认时长20分钟且可配置为任意正值；新五级选项文字随评分版本冻结。
- 效度存疑完整答卷仍可生成带显著警示的报告，但默认不进入排名/常模/汇总决策统计，并允许关联重测且保留历史。报告采用绿色视觉和A4十页目标但修正一级构成比误导及分页问题；受测者不显示效度原分/阈值，管理端与导出显示；个人信息继续按requiredFields动态展示，一期默认六项。
- 四类一期版本名确认为`competency-frontline-phase1-v1 / competency-phase1-scoring-v1 / competency-phase1-content-v1 / competency-phase1-report-v1`。允许后续在staging将90题作为候选测试内容导入，但不等于production正式内容批准；production仍需内容负责人、心理测量负责人审批和单独上线授权。
- 本轮仅记录决策，未改Go/Vue/SQL、未修改转换器或候选JSON、未导入90题、未写数据库、未部署。候选JSON继续保持`importReady=false`，直到后续RED→GREEN实现并重新生成。当前剩余客户交付为批准人/日期、效度良好与存疑最终文案、最终免责声明、正式文案源文件及production上线授权，不再存在P0/P1算法选择歧义。
- [纠正 / 00401一期90题staging正式导入完成 - 2026-08-10] 上述“未导入90题”状态已失效。staging导入前独立完整备份位于`/opt/talent-assessment/backups/phase1_90_import_20260810_161855`，包含数据库、旧后端、旧前端及传统题/题库关系/答案签名；数据库归档`gzip -t`通过，导入后三组传统签名全部不变。
- 一期候选已升级为`competency-phase1-import-v3 / converter 1.2.0`，`importReady=true`、blockers/warnings均为0；确定性十列工作簿SHA-256=`828c4267e6c7ad387a73ddb0e923b461d5a336ef225bd7414216c0def814de9f`。正式预览API返回90成功/0错误，正式导入返回90；重复预览返回0成功/90错误，重复正式导入被拒且行数保持90。
- staging最终数据为总题/唯一题号/维度题/效度题/维度正向/维度反向/效度正向/启用=`90/90/80/10/62/18/10/90`；10个A/B维度各8道维度题+1道关联效度题；胜任力题的`el_qu_answer`和`el_qu_repo`关联均为0，胜任力测评仍为0。
- FB-103按RED→GREEN修复维度启用题数误计效度题，维度API现在10维合计80题；FB-104按RED→GREEN修正导入弹窗“九列模板”为“十列模板”。Go全量、`go vet`、Linux build和前端23文件135项、production build全部通过。
- staging最终部署后端SHA-256=`eb0766cd5a98535c876457eb012f7bc1780f9e39bc4dc09ea75f80f4fe2d65a7`，前端归档=`1670a327e7cc1cead0b68b36a64eaead59e4b0509f4f1723b3dcded941c514a0`，index=`7c912375ed212515dafeedc4b717c75fcd8ce38ae0a98cc0a98231a5ca639366`。真实导出十列表头+90行并与导入文件按题号逐行一致；真实Chromium首屏含维度题/效度题标签、总数90、十列提示，console/request错误均0。
- 终验service/nginx/mysql均active、health ok、应用关键错误0、短时`phase1-*` Redis会话0、远端临时上传文件已清理。完整证据见`docs/00401-phase1-90-import-verification-20260810.md`。本切片仍不实现固定一期测评运行时、新五档、一级聚合、效度结果算法或正式十页报告；production未修改。
- [00401一期固定产品配置本地完成 - 2026-08-10] 用户选择先完成固定配置，并明确在新五档、一级聚合和效度算法尚未实现时禁止发布。后端固定报告对象`frontline_employee`、十个A/B维度及顺序、四类一期版本；空固定字段按一期画像补齐，leader、缺失/换序维度或任一其他版本在数据库事务前拒绝。
- 草稿保存新增一次题型分组查询，要求十维全部启用且每维恰有8道启用`dimension`和1道启用`validity`，未知题型为0；关联草稿的`question_count`仍只保存维度题数8。发布服务识别一期版本后立即返回`ErrPhase1CompetencyRuntimePending`，不读取或写入快照，避免旧`competency-v1`算法冒充一期评分。
- 前端隐藏领导人员版和维度选择器，显示只读“基层员工·10个A/B维度·90题”及四版本；切换胜任力自动写固定字段、新建默认20分钟和六项个人信息；切回legacy继续清空一期字段并恢复题库控件。发布按钮禁用且显示“一期五档、一级维度和效度运行时完成后方可发布”。
- TDD证据：后端RED为8个一期符号未定义，前端RED为2项固定画像缺失；GREEN后聚焦Go与前端通过，Go全量、`go vet`、Windows build、前端23文件137项和production build通过，编辑器诊断0。该切片未部署staging/production；下一独立任务为一期五档计分、一级聚合或效度结果算法。
- [00401一期五档纯评分引擎本地完成 - 2026-08-10] 新增`CalculatePhase1CompetencyResult`，只接收80道`dimension`输入并强制十个A/B身份、固定顺序、每维8题；维度分使用精确`big.Rat(scoreSum/8)`，总体分为十个精确维度分之和，不提前四舍五入，也不复用旧四档所需的维度均值。
- 二级L1-L5精确使用上包含边界1.7/2.7/3.5/4.3；总体使用`<25 / >=25 / >=32.5 / >=40 / >=45`，内部码为`not_qualified/weak/qualified/good/excellent`。可信超时的不完整输入保留80题及已答计数，但不输出正式维度分、总体分或等级；缺行、混入validity、未知维度或顺序不匹配直接拒绝。
- TDD证据：RED为一期等级常量和评分函数均未定义；GREEN覆盖二级边界、总体边界、十维精确总分30、反序输入仍按A/B输出、不完整无正式分和畸形输入拒绝。Go全量、`go vet`和Windows build通过，编辑器诊断0。
- 本切片只完成无DB依赖纯评分引擎，未接入Submit持久化，未修改现有历史`CalculateCompetencyResult`；原因是一期提交必须同时正确拆分80道维度题/10道效度题并写一级与效度结果，不能先写半套正式结果。I3运行时接入仍标❌，发布门禁继续关闭；未部署staging/production。
- [00401一期一级维度聚合纯函数本地完成 - 2026-08-10] 新增`CalculatePhase1GroupResults`：无论十个二级结果输入顺序如何，固定输出`general_ability/通用能力`（A1-01～A1-05）和`psychological_quality/心理素养`（B1-01～B1-05），顺序1/2，子维度集合不可替换。
- 完整一级分为五个精确二级`big.Rat`的平均值，并复用一期二级上包含L1-L5边界；不提前四舍五入。任一子维度不完整时，一级结果保留总/有效维度数和总/已答题数，但`Score=nil`、level为空且`IsComplete=false`；另一组不受影响。缺失、重复、未知身份、顺序错误、题数/已答数非法、完整但无分或分数/等级不一致均拒绝。
- TDD证据：RED为固定一级码和`CalculatePhase1GroupResults`未定义；GREEN覆盖反序输入固定输出、两组精确均值3/L3、单维未答后的5/4与40/39计数、1.7/2.7/3.5/4.3边界和5类畸形输入。Go全量、`go vet`、Windows build通过，编辑器诊断0。
- 本切片仍只实现无DB依赖聚合纯函数，未创建`el_exam_competency_group`快照或写`el_competency_group_result`；必须等待效度算法后与一期90题发布/提交事务统一接入。I4运行时接入仍为❌，发布门禁继续关闭；未部署staging/production。
- [00401一期效度计算纯函数本地完成 - 2026-08-11] 新增`CalculatePhase1ValidityResult`，严格接收10道唯一、有序、`validity + forward`题，只累加原始1～5分，不执行反向转换，也不读取/修改二级、一级或总体结果。
- 10题完成时`V<=35`返回`good`、`V>35`返回`questionable`；35/36边界和10/50极值均有测试。可信超时未完成时保留总题/已答计数，`Score=nil`、status=`incomplete`、`IsComplete=false`。题数非10、混入dimension、raw越界、反向元数据、空/重复题号或顺序错误均拒绝。
- TDD证据：RED为效度输入/结果类型、三状态常量和计算函数未定义；GREEN后聚焦效度测试通过，Go全量、`go vet`和Windows build通过，编辑器诊断0。
- 五档、一级聚合和效度三套纯函数现已齐备，但仍未接入DB。下一独立切片应统一完成：90题发布快照（含两个一级快照）→ Submit按题型拆分80/10 → 写10条二级、2条一级、1条效度和1条总体结果 → 默认统计排除不完整/存疑 → 解除发布门禁。I3/I4/I5运行时分支继续为❌；未部署staging/production。

## 2026-10-03T12:45:57Z MT-GUARD-AUDIT实际恢复库阶段PASS（仅本阶段）

- 唯一mng_guard_audit_test_729c47a8d0194b63从正式SHA9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1备份恢复；三归档SHA/gzip/tar/前后tarcompare全PASS，cross-schemaFK0，主element未写/DDL0，不部署/重启/改配置或生产逻辑。SSH strict liming/knownkey/Conn10实际通过，APP_ENVproduction＋REPORT_EFFECTIVE_ENVstaging。
- 未改001 first/repeat/post11表15RESTRICT FK完整签名相等，SHA3fb28c5c220c063795e77b3e27b8eed8ace01b0dc7ee82614a556af8f1c3c0d1（含FK逐列细节，非旧签名算法）；三个fresh生产CheckRuntimeSchema实际PASS，同实例两次仅四metadata；真实driver小写列标签及GORM行数11/159/67/21。无diagnostic替gate、无mock。
- 两code实际public FreezeProfile各140/700，原manifest/mappingSHA保持；actual config.Load在真实运行目录消费配置/实际secret，子测试仅内存env注入exact ownedroot socketDSN和secret，原fixture cwd保持；子cwd无配置的env-only提示如实保留，不当server启动验收。真实token roundtrip不打印/落盘，无身份HTTP/会话。
- 原source只有TestManagementTraitsGuardAuditMySQLExternal；新增独立TestManagementTraitsGuardAuditMySQLStagingExternal真实post5pass/0fail/0skip（3.67s），包含原canonical-audit-rollback子用例；7/11core＋owned capture的public IdentityScope/CandidateIDs/PaperIDs/AllLegacy均PASS，audit.report_id→revision.id→run复合paper/exam/participant实际闭包/rollback11表0。unknown/partial/孤儿revision-run/复合漂移真实failclosed；故障只在owned事务SESSIONFKchecks0造行、复原1再rollback，不改FK元数据/main。
- 原source-lock真实另行空mng_source_lock_test_729c47a8d0194b63，fresh原test4pass/0fail/0skip、2.65s，普通第二connectionUPDATE1205与释放后成功、MVCC-currentreadPASS。两exact库/remote payload/唯一local temp最终0，cleanup errors0，正式证据backup下guard_actual_729c47a8d0194b63永久保留。
- 旧12表指纹f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e/465PDF/server/index/config/templates/unit均不变，67/70/1487/134354/294628/1348/27；三服务active/healthok、finalSSHexit0。新TEST两目标/私有目录仍不存在，四MNG_TEST key仍未设；下一guard/drain→主DDL→TEST部署→四组合HTTP/UI/PDF/Worker验收尚未执行。
- 本地最终全量6039pass/638顶层/0fail/9环境skip/parse0/exit0，build/vet/临时bootstrapvet0/零error。新增StagingExternal本地无DSN明确skip，不把actualPASS折算成默认0skip；前端/PDF/race未重跑。fresh Linux server49828320/SHA0ee9b326c459611babb6951c63291abcdfc27491e8782d3dd9178e4bb56d8483仍仅本地；本次真实测试binary19895312/SHA96b7570d650c6cc958f868b37765248813c3308f1c2f4a3167f92094137480f6已exact清理，旧bin test不当本轮fresh。
- 同次可复用资产保留：scripts/db/management-traits-guard-staging-verify.sh、scripts/test/management-traits-guard-staging-real-test.js、scripts/test/fixtures/management-traits-staging-bootstrap.go.txt及新增Go实库test。bootstrapfixture与实际main规范化逐字节一致，--verify-assets/nodecheck/bash-n实跑PASS；JSwrapper尚未整轮重放，不假称已跑。重放显式--owned-staging/fresh编译/exactrandom库/knownkey/SHA/finally，无secret依赖，不再手写temporary recoding。详见[报告末尾actual阶段receipt](management-traits-local-implementation-20261003.md)。
- [纠正 - 2026-10-03] 本文件顶部fresh actual待验及历史完整capture阻断已由本次恢复库production public函数与audit闭包PASS限定解除；原失败保留，不能据此称已部署/完整产品DONE。本阶段没有实际失败后的逻辑修复。

## 2026-10-08T13:05Z 管理特质005隔离副本双产品真实E2E完成

- 用户授权仅写 `talent_mng005_local_7081fbec31e3d105`。真实应用reload的 `/getInfo` 为HTTP200/code200/admin/wildcard；`VerifyRuntime`逐次证明copyOnly、源element SELECT 1142、跨库FK0、loopback和reports disabled。没有共享staging/production写入或部署。
- 修正本地 verifier 将“私有Redis必须DBSIZE=0”错误当成持续隔离条件：RED为新helper缺失编译失败，GREEN聚焦Go测试通过；运行中redisKeys=1也能在数据库/网络/认证隔离全部通过时验证成功，负数元数据仍拒绝。
- 副本安装draft migration 003，并由哈希锁定、schema/stamp硬绑定SQL创建可复用独立00501/00502 fixture。两者各repo1/关系140/题140/选项700，独立PK与ownership marker；00201/00202各1/140/140/700未变。
- 真实本地管理UI分别Save为draft并显式冻结00501/00502；真实参与者UI分别完成140题和手工交卷。00501额外验证答70题后reload仍为70并定位第71题。完成态每产品均profile1/candidate1/paper1/snapshot140/legacy140/bucket700/run1/dimension13/module4/receipt1；00501 staff总体29.444146，00502 leader总体50.000000，版本命名空间分别为mng-00501/00502独立版本。
- 完成态证据后，精确exam-ID cleanup按实读RESTRICT FK顺序删除两个临时测评全链；两个exam所有上述瞬态计数归0，005 fixture与draft schema保留。冻结002 exam1791298091700970647仍completed1/profile1/overall58.642639及manifest/input/mapping/field-contract SHA全同；源SELECT继续1142。
- 未生成报告；本地开发入口继续禁用report/pdf/template。MySQL隧道、独立Redis、后端和Vue前端保留运行；没有DROP副本、删除fixture或关闭他人进程。完整证据与失败记录见 `docs/management-traits-005-local-debug-20261008.md`。

## 2026-10-03T13:01:44Z staging主DDL已安装，TEST部署失败已回滚

- **DEPLOYED=NO**：用户明确主staging backup/DDL/deploy授权后实际执行两phase；旧PID2001停服→PID/cgroup/8092/positive_app连接0→newguardPID24499且TESTdisabled/mng0/health→再全排空→main001first/repeat11表15FK，完整签名3fb28c5c220c063795e77b3e27b8eed8ace01b0dc7ee82614a556af8f1c3c0d1，3fresh应用账号CheckRuntimeSchema11/159/67/21/4queries均PASS，CRUD权限/002140/700ASCII只读PASS。旧表/源题不ALTER/DML/backfill。
- postchecker空body调用participant/paper-detail错误期待401，actualjournal20:56:48 HTTP400（body校验先于token）。exit1/restart-postcheck，按任何error停未改业务/失败断言或重试。新front路由、general/00401/MBTI HTTP、四组合/PDF/expiry/revoke/resume/Worker未执行，不称全部TEST链PASS。
- 旧后端/dist/env恢复；exact新增两个TEST资产、env/dropin、空private根、393-file失败候选目录和/tmp/mng_deploy_staging_3461565581060643均清0，候选/源码/script/DDL归档到正式backup/deploy_3461565581060643永久保留。无secret/token文件或git；**11主表保留且每表0**，不autoDROP/全库restore。
- rollback把server owner设liming造成Uid/Gid漂移；证据server.before统一chmod600不能作原mode参考，第二mode终验失败如实保留。最终从未变正式application.tar.gz实读server为root/root0755，恢复owner/mode/原时间；整application tarcompare0、3正式归档SHA OK、旧12表f57f35…／465PDF逐SHA不变。final表78/70/1487/134354/294628/1348/27，三服务active/healthok、finalSSH0。旧serveree4566e7…／index2d4ba6c5…仍在线。
- 本次fresh本地Go6039pass/0fail/9skip/parse0/vet0/build0，front31files355pass/build0；Linux0ee9b326…、index98547b68…、dist-tare0b37707…仅候选。完整SHA/路径及阶段receipt见docs/management-traits-staging-deployment-20261003.md；正式备份仍112853_fd24d4ee35a9（DB9ab00b04…）。下一核验要以11空表既存状态续作；合法paperId缺token401与空body400分开，metadata从正式归档核验，不从私有600副本猜原mode。停止后不自行再部署。
- [纠正 - 2026-10-03] 上方“主mng0/mainDDL尚未执行”已由本次main11表15FK限定替代；应用部署仍NO/已回滚，整产品验收仍未完成。失败与元数据工具执行记录不删除，final cleanup0不等于deploymentPASS。
- 最后公网health200/statusok、index200/SHA2d4ba6c5…证明旧版回滚，非freshroute验收；local checker exactSHA删源码/二进制/空目录remaining0。apply_patch Delete未落盘须按实际验证，不仅凭success。部署script＋3docs诊断0/scopeddiff0；正式backup与失败receipt保留，不再发布。

# 2026-10-09 Production release candidate local package (NOT_GO)

- Local-only candidate
elease-candidate-20261008T203459Z: deterministic Linux CGO0 server SHA $(@{schema=production-release-candidate-v1; createdUtc=2026-10-08T20:37:26.9358535Z; decision=NOT_GO; deploymentPerformed=False; git=; toolchain=; builds=; package=; featureGates=; environmentVariableNames=System.Object[]; testReceipts=System.Object[]; migrations=System.Object[]; blockers=System.Object[]}.builds.go.build1Sha256) (two builds identical) and frontend 393-file aggregate $(@{schema=production-release-candidate-v1; createdUtc=2026-10-08T20:37:26.9358535Z; decision=NOT_GO; deploymentPerformed=False; git=; toolchain=; builds=; package=; featureGates=; environmentVariableNames=System.Object[]; testReceipts=System.Object[]; migrations=System.Object[]; blockers=System.Object[]}.builds.frontend.build1AggregateSha256) (two builds identical), source maps 0.
- ZIP SHA $(@{schema=release-candidate-verification-v1; verifiedUtc=2026-10-08T20:40:20.5581958Z; archive=; extraction=; binary=; migrations=; goReproducible=True; frontendReproducible=True; sourceMaps=0; verdict=NOT_GO}.archive.sha256); all 404 extracted SHA entries verified. Secret scan high-confidence findings 0; one compiled synthetic marker remains and is accepted only with production MNG gate false.
- Management-traits TEST assets and all 9 reviewed migrations excluded/NOT_APPROVED. No remote/DB/deployment/startup. Overall NOT_GO; asset, migration, runtime and explicit production approvals remain blockers.

[Correction - 2026-10-09] The immediately preceding production release candidate memory entry was malformed by PowerShell interpolation. Verified facts:
- Candidate release-candidate-20261008T203459Z is local-only and NOT_GO.
- Linux server two-build SHA:
cc6665ecb134ec60b469a26871ed2ca27c9289dc9ee79d73c8d1838fe2baa354
; frontend two-build 393-file aggregate:
6111144547dd6397764ec217c25b6738adf6cf78be7a7956e469f4054b6c0abe
.
- ZIP SHA:
e342b297a2c0425a0582830b1414396c154f139ee9b68d187cbe8448ce1ae79f
;
404
 extracted SHA entries verified.
- Secret scan high-confidence findings 0; one compiled synthetic marker remains under the production-disabled management-traits gate.
- Management-traits TEST assets and all nine reviewed migrations are excluded/NOT_APPROVED. No remote access, DB action, startup, or deployment occurred.

## 2026-10-09 当前工作树 management-traits staging 发布（PARTIAL）

- 仅发布到 `20.200.136.133` / `vm-ubuntu-go-dev`；未访问或修改 production `39.106.61.48`。发布源为当前工作树 HEAD `5217ce6558eb7876d9e4cc37f1a617296deb2590`，没有使用旧 `release-candidate-20261008T210616Z`。
- 当前源码门禁：Go 全量通过（环境型实库测试按原条件跳过）、前端 34 文件 598 项通过、Go build/vet 通过、前端 `build:prod` 通过（保留 2 个既有体积 warning）。两次 Linux CGO0 构建一致，server SHA-256=`d30c8e40e0c99dc525ad42b1269581665dd72acb7e1bfd47958ac0efd1561fc7`；前端 393 文件，index SHA-256=`c4f3b8f76b740244bd6b1d10ac6e24e4650f15367f1c427e4b46055297e21555`。
- 发布前受限备份为 `/opt/talent-assessment/backups/mng_current_d0e8202eafb14b08`，数据库、旧 server、旧 dist、配置/systemd 归档均完成 SHA-256 与 gzip 校验。恢复库先后两次执行 draft 003 和 reissue 004，结构重跑通过；真实两个应用连接在 REPEATABLE-READ 下验证一份创建、一份复用、精确 `uk_mng_reissue_input` 1062、1 report/4 audits/无孤儿，临时库最终为 0。
- 主库只新增 `el_mng_exam_draft`、`el_mng_report_reissue`、`el_mng_reissue_audit`，脚本均执行两次；最终 management-traits 表14、RESTRICT FK18、三张新增表行数均0。未安装 formal registry/approval 表，未 backfill、未改旧结果或旧 PDF。
- 后端只重启1次，最终 PID13210；talent-assessment/nginx/mysql active，内外 health 200，磁盘和进程 server SHA一致，公网前端393文件逐SHA通过。环境为 `APP_ENV=production`、`REPORT_EFFECTIVE_ENV=staging`、`MNG_TEST_REPORT_ENV=staging`；最近应用关键错误0、Nginx 5xx=0。
- 保留事实：00501/00502各1个 completed run、13维、4模块；冻结002 completed run=1；普通00201/00202 repo=2；私有PDF=3。受保护原 management-traits 表发布前后 dump SHA一致，active paper=0；临时payload和演练Schema最终0。
- 验收判定仅为 **STAGING PARTIAL**：匿名 legacy/admin/reissue 边界均401且健康/数据/产物通过，但当前集成浏览器没有真实已登录会话，故没有伪造token或storageState；新reissue报告生成/查看/下载、同run并发HTTP复用、临时005 draft/person创建冻结与精确清理未执行。独立CodeReviewer也未在本BreakGlass阶段执行。production继续 **NO-GO**。

[纠正 - 2026-10-09] 上述`STAGING PARTIAL`是该次发布当时的阶段判定。后续同日已完成真实管理员浏览器、模板元数据/下载/上传、00501/00502保留结果、TEST报告生成/view/download、数据及报告精确清理，并发布后续MBTI修复版本；当前状态改按[project-status.md](project-status.md)读取。未补齐的全新005 draft→freeze→140答生命周期、formal/production批准仍保留为P1/P0，不把后续范围GREEN扩大为全部功能或production GREEN。
