# User Feedback Log

## UF-057 — 2026-10-09 ✅ production旧报告已覆盖

- 分类：报告版式 / 已生成资产。五问：①production管理端打开`/#/exam/competency-results/1791556668267388349`并查看正式报告；②UF-056活动模板发布后复查发现；③当前明确00401 v2报告及paper=`9174f98e-181f-486e-bbc7-0118bc39ced1`；④首页仍显示数字`1`，页脚显示`page2`；⑤用户期望时长整项无任何残留、页脚居中只显示数字，并已明确批准覆盖生成的旧报告文件。
- [关闭 - 2026-10-10] 先以RED锁定隐藏时长、PAGE字段兼容、页码内层段落居中及`计划执行：`粗体。staging真实90答/10页PDF通过；production LibreOffice 7.4确定性DOCX进一步证明需移除冗余wrapper。最终后端SHA=`c321bf35...`，指定PDF覆盖为SHA=`cf1385c5...`/758202 bytes；封面无时长/页码，正文数字1–9居中，无Page前缀，计划执行粗体。report/current=1/1，audit最终15，备份=`/opt/talent-assessment/backups/uf057-report-20261010001515`，下载/DB/文件一致。

## UF-056 — 2026-10-09 ✅ production模板已发布

- 分类：报告版式 / 部署资产。五问：①production 00401完整答题后生成正式PDF；②production与staging活动v2模板SHA一致，不是单机漂移；③当前确认00401正式报告；④用户要求首页时长隐藏整项、页码保持居中、`胜任力综合表现`保持模板粗体；⑤触发试卷=`9174f98e-181f-486e-bbc7-0118bc39ced1`，来自保留测评`TEST-BROWSER-00401-20261009223722`。
- [定位 - 2026-10-09] production PDF实测包含`时长：1分钟`，故只有首页时长不符合。DOCX页脚及PDF坐标证明页码已居中；标题、优势项/待发展项和动态维度名按模板加粗，正文常规字重。PDF已嵌入微软雅黑粗体/常规及宋体，本问题不是相关字体缺失。
- [本地修复 - 2026-10-09] 先新增`TestBugUF056_Phase1V2CoverHidesDuration`并取得RED；从线上精确v2基线SHA=`a814c36e...`生成候选，仅隐藏可见标签/单位并保留隐藏`result.userTime`强合同。候选SHA=`52e0020c...`；专项1/1、v2 Word 5/5、handler 53/53及Go build GREEN。未替换staging/production模板；目标LibreOffice动态转换仍待重新设计验证流程后完成。
- [发布关闭 - 2026-10-09] 重构传输门禁后，staging真实90答与LibreOffice 10页PDF证明首页时长隐藏、审计2、临时闭包清理0、基线无漂移。用户随后明确批准production只替换模板且不涉及数据库；production解包差异仅`word/document.xml`，活动v2模板SHA=`52e0020c...`，备份=`/opt/talent-assessment/backups/uf056-template-20261009232349`。服务PID2370974/NRestarts0，内外root/API 200，关键日志0。未重生成旧报告，因此保留PDF仍是历史版式。

## UF-055 — 2026-10-09 ✅ production全站500已关闭

- 分类：部署/环境。五问：①直接访问production根地址`http://39.106.61.48:8090/`即失败；②此前可用、刚刚开始；③用户确认所有页面；④浏览器只显示`500 Internal Server Error`，暂无失败请求响应正文或Request ID；⑤无特定测评、试卷或用户数据条件。
- 期望：production根页面和API代理正常响应。实际：公网根地址返回500。该问题发生在本轮staging基准只读检查之后；该检查未执行远端文件写入、DDL、DML、服务重启或部署。
- 下一步：先从公网分别探测root与`/prod-api/health`，再只读检查Nginx错误日志、应用service/journal、监听端口、磁盘/权限和配置引用；定位前不改代码或配置。

[关闭 - 2026-10-09] 公网实测root/favicon=`500/500`而`/prod-api/health=200`，隔离为Nginx静态文件链；后端active、8092=200、PID=`2370974`、NRestarts=0。`namei`确认dist/index为755/644，但父目录`/opt`及`/opt/talent-assessment`被置为0700，Nginx worker `www`无法穿越。根因链为已禁止复用的第一代00401历史清理controller：先把备份目录全部chmod700，再由失败rollback执行`cp -a report-files/. /`，把备份中的父目录mode保留到真实`/opt`和应用目录；此前只纠正了`/`，漏检两个父目录。

[修复与验收] 用户明确批准后仅将`/opt`和`/opt/talent-assessment`由0700改为0755；未改文件、数据库、Nginx配置或程序，未重启服务。以worker `www`读取index通过；服务器内root/favicon/API均200；公网真实HTTP分别`200(16155 bytes)/200(26900 bytes)/200(15 bytes)`，浏览器真实打开并显示“人才综合素质评估系统”登录页。失败断言会自动回滚原mode，本次全部通过未触发回滚。

## UF-054 — 2026-10-09 ✅ production已关闭

- 分类：数据一致性 / 题库范围认知。五问：①admin打开production `/#/exam/repo`，确认题库列表显示总题目数554；②此前也是554，不是本次发布回归；③当前仅确认admin；④页面无报错、空白或网络异常，只是数量不符合预期；⑤用户预期全局题目总数为90。
- 已核代码事实：题库列表由`/exam/api/repo/paging`返回物理题库行，并额外插入虚拟00401行；00401行只统计`dimension_id IS NOT NULL OR competency_question_type IS NOT NULL`的胜任力题。此前production只读receipt已证明00401精确为90题，因此“全局应为90”与“只替换00401、其他产品不动”的既有范围存在破坏性歧义。
- 当前未修改代码、题库、关联、历史答卷或环境。下一步先只读拆分554的物理题/胜任力题/题库关联和被测评引用数量，再由用户明确是否真的删除所有非00401题目；未获明确授权前禁止把全局554删到90。

[执行补充] 用户已授权删除旧00401完整测试闭包。只读确认页面实际00401计数为544=当前90+旧454；其他传统题858。删除controller在完整备份后被safe-update/MySQL tmp权限门禁拦截并自动恢复；期间5条报告被不完整restore清空，随后从有效全备份精确恢复。最终9测评/22试卷/19结果/5报告、旧454题均仍在，当前90题及其他产品不变；`/`权限修复为0755、mysql tmp写和应用health通过。状态保持🔍 BLOCKED，不得把本轮称为已删除。

[关闭 - 2026-10-09] 用户再次明确直接删除后，改用逐表主键常量DELETE的新controller，完整备份=`/opt/talent-assessment/backups/phase1_history_delete_v2_20261009220803`。旧9测评/22试卷/19结果/5报告/11审计/454题/48维及其冻结和试卷引用全部删除；当前10维90题的数量、分布和两项SHA与staging实时收据完全一致，传统858题保留。服务、MySQL临时写和root目录权限终验通过。

## FB-218 — 2026-10-09 ✅ staging关闭

五问事实来自本轮真实链：staging MBTI，48题全部保存并完成计分，预期完整版报告在有界时间内返回PDF，实际旧链两次超过180/240秒。根因是MBTI独立使用无context的LibreOffice命令且转换失败会继续返回DOCX路径。RED先锁定有界共享客户端；GREEN改为复用`libreofficepdf.Client`、90秒context和隔离临时目录，失败固定业务错误且不更新PDF成功状态。部署后真实ESTJ完整版/简版PDF、16+16模板和匿名门禁PASS，临时业务链及两个旧profile清理0；production未访问。

## MT-ADMIN-ENTRY-ISOLATED — 2026-10-08 候选可复审，永久反馈未关闭

沿原五问与指定旧URL403/已有completed结果事实，不新索PII或操作原浏览器。只从52eec已上线隔离source修管理入口server布尔分流，旧结果页/API保留；实际入口RED13/73→GREEN86/0、生产模式build0、单SFCdiff与393资源候选已返回。完整编译闭包证明第三轮未完成后停止，独立CodeReviewer待主协调者，不伪reviewPASS或发布。08:33公网仍52eec；旧URL永久分流反馈🔥仍开放，原staging及reissue双MySQL门禁保持。[源码、包、失败和下一边界](management-traits-entry-candidate-prepared-20261008.md)。

## MT-ADMIN-DIRECT — 2026-10-08T08:15Z 只读入口临时恢复／永久分流仍开放

分类：部署/入口业务链。五问已给：指定staging旧URL直达、历史是否可用未知、冻结00201管理员、403/空表、同exam1791298091700970647。fresh真实getInfo200/code200及sameexam strictfrozen true；旧tester-list403/code1而专属list200/code0，根因为旧线上入口未按服务器冻结信息分流，保护本来正确。已在原浏览器手动切专属结果页，实际completed1/140140、详情13维4模块及绑定一致；overall两位58.64，旧UI显示原精确有理数，不冒新显示样式上线。

✅ 当前可读页恢复；🟢 既有local分流fresh167专项通过；❌ 旧URL自动分流未发布/反馈不永久关闭。无产品新改动/报告动作/真人写/DDL/部署/生产。主协调者安排独立最小old-TEST入口SFC及compiler闭包review，不带尚未安装的新reissue UI。[本轮证据与当前地址](management-traits-reissue-staging-release-20261008.md#L3)。

## MT-REISSUE-1062 — 2026-10-08 本地修复／staging反馈门禁未关闭

沿完整用户交接五问：已批准统一staging报告发布前、独占恢复库两个连接生成同完成run；预期一created一reused同报告/原审计文件完整，实际一SUCCESS一1062且report1/audit1；旧backend4179/front52eec/主两表0/58.64未发布保留。原errno收据没有具体键，不补造根因实证。代码/RR路径及SQLmock23分支RED→GREEN仅target reissue服务，其他unique/audit1062/1213等失败关闭，归档/生成审计/PDFsizeSHA匹配才复用。

专项111/0/2skip、全Go6383/0/11skip/build-vet0；新增真实双readview屏障fixture已编译未远端运行。07:42首次＋07:43唯一strictSSH重试连接前255timeout，remote0、无主写/部署/真人生成/production，实际修复验证及独立review仍开放；不因用户已批准发布绕过门禁。[完整证据与下阶段](management-traits-reissue-staging-release-20261008.md#L3)。

## MT-ADMIN-DIRECT — 2026-10-08 ✅ 本地完成／远端反馈未关闭

沿用户完整五问，不再索取真人身份。冻结002旧URL无query/session时未分流，实际SFC先RED；只按同exam服务器strict字段与fresh管理员权限分流，旧false历史保留/unknown不偷偷旧回退。结果页改为原Element UI查询/分页/分数详情＋独立客户模板报告（TEST）三动作、lazy元数据选版本，不手填ID、不改旧生成器/警示。

本地真实bundle发现第二次只读POST被公共1000ms防重复提交拦截；两Timeout保留，用户结构化额外1轮批准后只接现有隔离wrapper，不放公共guard或加等待。最终486前端/build0、两mock视口完整操作PASS/13维4模块/页面无溢出/503结果仍在。**未部署、远端未验**；姓名手机号/逐题未由当前后端提供，界面如实标明，不冒完整原功能等价。[完整范围、失败、证据](management-traits-admin-ui-local-20261008.md)。

## MT-ADMIN-DIRECT — 2026-10-08 调研／本地实施

五问沿用户交接：①指定旧管理URL直达；②历史是否曾正常未知，不猜commit；③冻结00201/00202管理员；④getInfo200而旧列表403/空表，专属列表已有完成结果；⑤同exam strict冻结标记，测试只用合成exam/run。分类：业务链/入口分流。目标是参考原Element UI管理页完成结果与客户模板报告操作；本轮不访问生产/staging、不写真实身份或报告，远端反馈仍开放。

## MT-CANDIDATE-MINIMAL — 2026-10-08 ✅ staging 最小身份反馈限定闭环

五问沿已给截图/指定测评/预期字段，不再次索取姓名手机号；gender确为合法配置，不隐藏或去必填。用户最终独立source/artifact PASS/ALLOW后，SSH恢复按原批准完成隔离最小backend＋front发布，主重启一次；公开Detail同ID/stricttrue/三字段及普通URL DOM实际验证。没有更新原用户资料/答案、没有代保存真人或刷新原输入。

正常独立管理员登录后，仅一新synthetic开放00201TEST正常UI创建/明确冻结；普通考生实际三字段保存HTTP200/code0/extra0，SQL身份记录1/配置子集一致；FK/SafeUpdate/exactPK清理0、共享bundle和原用户profile保留、旧465/source/config/Schema/cache保持。状态：**STAGING DEPLOYED＋SYNTHETIC SAVE VERIFIED**，当前最小参数错误切片关闭；客户自己的保存未代执行，可刷新原页加载新资源后自行重新保存。不把旧失败覆盖为PASS、不称草稿新建政策/正式报告/全部系统/生产已完成。[完整限定报告](management-traits-candidate-identity-local-20261007.md#L3)、[最终证据](../scripts/test/results/mng-identity-publish-20261008/acceptance-verdict.json)。

## MT-CANDIDATE-MINIMAL — 2026-10-07T15:07Z 🔴 未发布／身份保存仍未远端验收

已有最小staging授权和用户转交两源独立PASS保留，本轮SHA匹配、不改旧review收据。七个缺失资源已证为隔离public遗漏的5copy＋2gzip，406构建输入核定，两npm生产build各393/native0；但完整编译差异347新/不同＋345旧缺，原路径驱动180s未完成/三编辑停止，不把393或source PASS冒compiled PASS。fresh线上PID2006/旧8baa/353c/三healthy/全state1与overduecompetency0，远端写与原用户页操作0，synthetic save NOT_EXECUTED，不称问题已解决。[同一报告](management-traits-candidate-identity-local-20261007.md#L3)、[新限定收据](../scripts/test/results/mng-identity-minimal-20261007/artifact-bounded-verdict.json)。

## MT-CANDIDATE-MINIMAL — 2026-10-07T14:50Z 🟢 隔离旧协议GREEN／🔴 线上反馈仍开放

沿用户已给五问，不索取或提交真人姓名手机号；最小旧Detail冻结bool/身份字段＋candidate切片staging已批准，不重新要求相同授权。精确发布源已从对应源码历史恢复且Linux基准完整SHA等于在线8baa…；两侧有效RED→Gin24/SFC106 GREEN，原白名单与历史路径不改，无新draft表依赖。

当前 source-review ready，不冒独立CodeReviewer PASS。独立复审与前端编译资源归因仍欠：隔离386资产baseline index不等线上393资产index，不部署此dist。最后14:50严格只读SSH0/PID2006/backend8baa/front353c/三active/state1=0/profile frozen1/三字段/draft表0/healthok；远端替换/restart/DDL/DML/生产0，用户原身份保存成功未验。主协调者接续复审/构建归因/已批准维护发布与独占synthetic200验收，[完整证据](management-traits-candidate-identity-local-20261007.md#L3)。

## MT-CANDIDATE-FIELDS — 2026-10-07T14:30Z 🔴 线上未发布，部署/合同断点已实测

五问沿用户已给事实：普通开放考生保存→截图固定错误；是否历史回归未知；指定1791298091700970647/00201；姓名电话性别已填却被拒；预期仅配置字段。分类：部署/环境＋新旧身份合同断链。本轮不索取真人信息、不扩报告/审批/draft工作。

公开Detail200/三配置键正确但新版标记全缺，真实浏览器旧modefalse；合成handleSave+确认拦截13键，extra8（7空/null），abort未远端保存。SQL冻结合同确为三字段/profile1/frozen1/draft表0；不改用户exam/person。fresh本地109＋435前端、Gin16＋9事件及build0，旧exit1未复现；不是远端GREEN。frontend-only不满足合同，整包带未批准代码不能发布；主须确认仅旧后端Detail冻结投影/对应candidate切片与维护窗口，既有staging修复授权保留。[同一报告与完整边界](management-traits-candidate-identity-local-20261007.md#L3)。

## MT-NEW-DRAFT — 2026-10-07 ✅ 本地实现 / 🔴 远端未验证

沿已给五问和已选sidecar技术实施：新002保存/取消freeze明确draft，不能旧身份/开卷；精确admin人员准备正常，旧历史继续原策略、不自动关卷。真实Save/Gin与SFC有效RED→GREEN，最终Go6272/frontend435/build-vet0。[完整契约及证据](management-traits-new-draft-local-20261007.md)。新增DDL未执行、远端操作0、独立review待主安排，未关闭线上反馈或正式报告任务。

## MT-NEW-DRAFT — 2026-10-07 调研/本地修复中

五问沿明确交接：新002保存→取消冻结；既有历史非本次迁移；影响00201/00202新建；profile0导致旧链绕行；使用合成exam/repo，不输出用户姓名电话。分类：业务规则/持久化门禁。政策已确定独立sidecar＋两步确认，不再询问底层技术选择；远端反馈尚未关闭。

## MT-CANDIDATE-FROZEN — 2026-10-07 同一反馈增量闭环（仅本地）

- 沿前次五问，不重新索取姓名电话。磁盘实际普通入口实现存在，但新版字段加载未验证server frozen/当前ID；unknown可旧链回退。原生SFC12fail＋新增unknown8/URL跨scope4有效RED，随后最小收紧，不放宽backend whitelist。
- ✅ LOCAL GREEN：身份106、全前端430、前端build native0；❌ 独立CodeReviewer未执行/远端保存未验。真实配置/身份/答案/历史PDF/TEST警示不改；本次SSH/SQL/部署/restart0。[完整纠正和失败记录](management-traits-candidate-identity-local-20261007.md)。

## MT-CANDIDATE-FIELDS — 2026-10-07 本地GREEN，用户远端反馈未关闭

- 已完成普通入口server-frozen分流与白名单提交；不是把gender去必填或放宽后端。当前真实SQL/冻结合同仍三字段，公开Detail200缺新版标记已只读核实；截图时字段差异/具体失败body未抓取，不能称唯一gender根因。
- actual SFC有效RED7fail→最终身份SFC/API80pass/native0；全前端32/404、最终Go6247/0/9skip/native0、build/vet0。name+telephone DOM/rules/payload精确两字段，配置gender保留，unknown/null/未配置非空gender仍拒绝/零INSERT。
- 状态：✅ LOCAL GREEN / 🔴 REMOTE NOT DEPLOYED；真实配置/身份/答案/PDF零写，无部署/restart/.env或formal后续操作，不冒客户当前页面已修。分类新增业务链/身份字段分流1。[完整闭环与全部消费方](management-traits-candidate-identity-local-20261007.md)。下条调研历史保留。

## MT-CANDIDATE-FIELDS — 2026-10-07 调研中（用户身份保存）

- 五问与范围：开放考生保存；是否历史回归未定；指定exam1791298091700970647；截图性别必填及配置外字段参数错误；用户预期仅已选字段。当前只读SQL实际配置及冻结合同均name,gender,telephone，截图name+mobile与当前数据差异未归因，不修改真实配置。
- 分类：业务链/字段配置与新旧分流。普通queryless入口走旧全表单，严格后端拒绝未知键/null；不能仅将gender去必填或放宽后端。不记录实名手机号。
- 🔍 先写合成真实SFC与Gin RED；不发布、不改.env/guard/真实身份答案/PDF/正式后续功能。[修改前事实与全部消费方](management-traits-candidate-identity-local-20261007.md)。

## MT-FREEZE-CONFIRM-RACE — 2026-10-06 有界阻断诊断（测试竞态，局部关闭）

- 5问沿用户完整交接：①新00202draft→保存/独立freeze超时；②历史回归未定，不推commit；③受控synthetic/admin，非全用户生产；④旧Timeout无HTTP，新6a5e两confirmpassed后响应Timeout；⑤旧13ee/9b及本轮exactowned标记均保留、清理0。不重复询问已知授权。
- 分类：UI测试异步竞态。真实本地SFC合成save延迟250ms，两次confirm标题“提示”，真正冻结模态未确认，exit1；先RED再另获一次driver最小修正批准，同exam保存屏障＋精确冻结标题GREEN0。不是产品APIbug/报告genbug，不追溯所有旧timeout唯一原因。
- ✅ 本地GREEN＋单cc784staging正常UIfreeze200/独立SQL25min/paper0/hash有效、exactfinally/final0/旧465/PID同；原集成页未改，不导出凭据。❌ 完整四UI/自然/四native-Python未完，production正式报告方案P0未实施；只[准备准入/回滚](management-traits-production-readiness-20261006.md)，无发布授权。[实际证据及剩余门禁](management-traits-default-staging-result-20261006.md)。

## UF-053 — 2026-10-06 主竞争证据补强（本地批准 A；非新业务故障）

- 5问沿完整上下文：①主HTTP manual与Worker唯一caller/锁时序欠证；②不是已确认业务回归；③只补已批准独占合成卷观测，不全用户日志；④UF052同pool/P_S history10无法唯一归属，原raceUNPROVEN；⑤旧合成卷已清，当前不创建新远端目标。用户明确选择“先增加最小脱敏观测并本地验证；staging后端发布及重启再单独确认”，不重复提问、不把批准扩为deploy/SET。
- 本地service固定caller/opaqueattempt+resource/session/有界内存时序→事务后原slog；两个process opt-in键默认关闭、仅target hash匹配，公开签名/HTTP/SQL/锁/期限/调度/权限不改。service25+真实Gin7测试事件exit0，含新观测终态误报有效RED→GREEN；仅sqlmock，不主实库race。
- 状态：**local合同 GREEN / 主竞态 UNPROVEN / 远端发布待单独批准**；独立CodeReviewer未执行（BreakGlass不再派发）。旧四UI/自然三例/恢复库race/单native样本与旧失败收据保留，旧21历史漂移不回滚。[完整影响、配置和未验](management-traits-four-real-verification-20261003.md#L3)。

## UF-050 — 2026-10-06T04:52Z 正常手工新增staging反馈已真实关闭；整套PARTIAL

- ✅ 原UI→NULL→login失败已由仅front默认字符串0发布及本批四正常UI payload0/SQL非NULL0/正常password login200code200五键限定关闭；精确draft准备先于freeze，不宽guard/旧行/ImportData。旧失败保留，非production发布。
- ✅ 后续四功能UI/自然25min0-3-140/20min提示/四admin结果报告200及所有exactPK finally0实际完成；0/3正式NULL/noPDF，原凭据expire-save三409、incompletegenerate真实四409（重复观察导致额外两请求，不伪填两次）。原admin200/home/context1/密码global清、旧465/source/runtime/schema/cache/PID2002同。
- ❌ native4report各2timeout/saveAs0与主Worker实际競争未证明，一140细粒度丢失；四SCP副本不native，整体PARTIAL、pending资源空。[完整闭环及未验](management-traits-four-real-verification-20261003.md#L3)、[最终summary](../scripts/test/results/uf050-ui-f61006041501/summary.json)。

## UF-050 — 2026-10-06T04:33Z staging手工新增反馈闭环；产品suite仍在运行

- ✅ 仅手工UI新增状态合同：全部tester未freeze前精确query新增/刷新200，payload字符串0、独立SQL四status0/nonNULL；全部准备后freeze，四正确密码登录200/code200/五键。仅前端发布、后端/guard/Schema/旧状态不改，ImportData保持未覆盖，不将所有新增入口标修复。
- ⚠️ 整体PARTIAL：完整UI3/4、自然0/3/140原25min正在等待、native三份各双eventtimeout/saveAs0，精确ownedfinally尚未执行。旧RED/401/timeout及旧summary原样；SCP不冒native、唯一run不冒竞态。[当前完整范围与证据](management-traits-four-real-verification-20261003.md#L3)。

## UF-050 — 2026-10-06发布获准但上传前BLOCKED

- 🔴 远端反馈仍未关闭：仅前端staging授权明确；原页03:56:09Z真实getInfo401，发布编译差异最终测试exit1/SSH read-only comparison failed，原因/传输细节未保留，不称双timeout。未再重跑连接、未上传/切换/backend替换/restart/生产操作，未创建业务实体。
- 已核fresh主机liming/PID2002/三active/后端8fb264e/旧front98547；617项仅SFC默认行变化、212Go/旧11receipt不变；393资源385相同及部分编译引用对照通过，**完整门禁未通过**。真实新增payload0/SQL状态/login五键未验，四完整UI0/4/natural/native保持；不以本地GREEN或历史清理关闭反馈。
- [本轮失败收据](../scripts/test/results/uf050-staging-20261006/blocked.json)、[完整授权边界与接续](management-traits-four-real-verification-20261003.md#L3)。下一正常原页getInfo200与SSH稳定后续用已有批准，不伪JWT或重复要求发布批准。

## UF-050 — 2026-10-06本地最小修复完成，远端反馈仍待发布复验

- 状态：🟢 LOCAL GREEN / 🔴 staging验收未关闭；分类/5问沿下条实际证据，历史回归未定。原四statusNULL及loginHTTP200/code500、完整UI0/4/natural/native未验、suite2blocked/3-4notstarted/5cleanupPASS原样保留。
- 仅SFC reset补既有字符串`status: "0"`，不修改后端Create/Update/Login/guard/Schema/旧行；编辑整体回填并保存`"0"/"1"/NULL/"2"`原值，下一次新增才重置0。candidate及独立ImportData不改，不能称全部新增入口修复。
- fresh原RED1fail→扩展RED3pass5fail→actual SFC8pass/相关184pass；原backend状态拒绝及守卫264pass/0fail0skip，frontend production buildexit0/既有两体积warning和Browserslist通知；两diagnostics0。仅本地，不新DB往返/部署/restart/生产，不关闭远端状态。[完整范围](management-traits-four-real-verification-20261003.md#L3)、[GREEN证据](../scripts/test/results/uf050-local-20261006/verdict.json)。

## UF-050 — 2026-10-06系统验收新发现：正常UI创建tester为NULL状态，TEST登录拒绝

- 分类：业务链/数据状态合同；状态：🔴 已复现、本地RED，未修。5问事实：①exact draft查询→正常UI新增/刷新200→全部准备后freeze200→正常密码登录；②是否历史回归未定；③本批4新增tester均statusNULL，实际登录仅1actor两次有界；④第二真实HTTP200/code500/身份校验失败，无五键；⑤markerMTHa61006021801/exam1791253225699847023/person1791253280154118748，完整ownedSQL保留。不是客户生产反馈，也不是旧403或admin401。
- 新actual SFC回放原add payload.statusundefined，期望显式启用'0'，1fail/0pass/exit1（原3case未选择）；真实UI/SQL合同REDexit1。当前源码strict admission拒nil或非0，未直接采部署内部stage；不放开guard/改共享DDL或主库状态，不默认修复业务。
- 四完整UI/natural/native未验；已停止owned timers/pages，原FK精确清理4exam/4tester/1candidate/1paper及全部child0、old465/source/runtime/Schema/cache/PID同、原getInfo200/home/testglobals清，pending空。新UF-050不关闭；UF-049 scoped准备仅本批远端通过。
- [actual SFC RED](../scripts/test/results/mng-premature-timeout-a61006021801/sfc-status-red.json)、[正常UI及失败response](../scripts/test/results/mng-premature-timeout-a61006021801/ui-receipt.json)、[完整证据](management-traits-four-real-verification-20261003.md#L3)。

## UF-049 — 2026-10-06重新归因：正确守卫／验收路径纠正本地GREEN

- **状态：测试路径local GREEN，远端未复验；不是业务代码已修或生产反馈已关闭。** 5问沿原证据：失败在新增后刷新；回归未知；正常admin/wildcard；无examId GET403/code1；另一candidate frozen1与新增tester profile0。没有问答缺失导致猜测权限表缺项。
- [纠正 - 2026-10-06] 原记录把冲突推为必须新增专用人员列表过宽。query表单与新增form不同，验收只选弹窗exam；全库AllLegacy正确拒绝。改验收流程：所有tester先于任一owned freeze准备，查询选exact-owned未冻结exam再新增；冻结后专用参与者/结果，不读旧名单或取旧密码。
- 新实际SFC方法回放原路径0pass2fail/exit1→仅测试驱动scope纠正3pass，相关5文件179pass；Go原守卫/实际handler/router/service250事件/0fail0skip，buildall/vet0。browser helper已提供但远端未跑，旧403/RED/summary及cleanupPASS不覆写，旧msg问号不当原始中文body。
- 未新创建/部署/restart/主DB权限写；原管理员02:13:03.553ZgetInfo200/home/storage0。四完整UI/natural/native仍待完整suite，不自动修业务、补新API或放guard。[完整归因](management-traits-four-real-verification-20261003.md#L3)、[新RED](../scripts/test/results/mng-personnel-local-20261006/red.json)、[本地verdict](../scripts/test/results/mng-personnel-local-20261006/verdict.json)。

## UF-049 — 2026-10-06 系统验收发现：002 TEST人员管理刷新被旧守卫拒绝

- 分类：业务链/UI交互；状态：🔴 实际RED，未修。复现：正常admin创建/freeze00201candidate→正常创建closed00201tester配置→测评者管理新增合成人员→保存成功后刷新旧unfiltered list。不是用户生产反馈，来自本轮已批准staging验收。
- 预期：正常管理闭环应安全可用，不泄露新实体凭据、不调用受保护旧读取；实际：GET /prod-api/exam/api/tester/list?pageNum=1&pageSize=20 HTTP403/code1/固定新版实体旧接口拒绝，pageError。正常新增由UI成功及SQL tester1791252035462016165核证；等待回调URL全局缺失属于工具错误，不冒填POST状态。
- 是否回归：未确定；范围：本轮新candidate profile存在时该管理员无exam筛选列表，未夸所有用户/历史时段。触发markerMTH6c260106a001/两exam1791251809609132344与1791251878134699743；原admin getInfo200/admin/wildcard，非失效登录。当前源码AllLegacy保护与UI旧getList消费链已定位，远端源码语义逐字未核，guard不放宽。
- 测试：captured实际403最小UI RED0pass1fail/exit1；尚无本地GREEN。真实业务失败后停止新资源，exact PK finally与独立终验全0，原admin会话保留。四UI/natural/native未完成；修复及staging发布不在本轮自动执行。
- [真实RED](../scripts/test/results/mng-ui-natural-native-20261006/ui-business-red.json)、[完整实测与影响](management-traits-four-real-verification-20261003.md#L3)。

| ID | Date | Feedback | Classification | Reproduction | Regression | Scope | Symptom | Trigger Data | Status | Test/Fix |
|----|------|----------|----------------|--------------|------------|-------|---------|--------------|--------|----------|
| UF-001 | 2026-07-25 | 题库列表看不到编号00401的胜任力测验题库及其题目 | 功能缺口 / UI交互 | admin进入“测评管理→题库管理” | 从未显示过 | 当前admin；功能设计对所有管理员一致 | 现有列表仅6个传统题库，无00401入口 | code=00401，名称=胜任力测验题库，384道胜任力题 | ✅ staging已修复 | `TestFeedbackUF001_CompetencyQuestionBank00401IsReachable`; `competency-question-list.spec.js`; staging API/browser 384题 |
| UF-002 | 2026-07-26 | 胜任力测评预览页点击“开始测评”后无法进入答题，重复提示英文 `competency exam is not published` | 业务规则 / UI交互 | 受测者进入准备页 → 点击“开始测评” | 第一次测试即失败 | 根因影响所有未发布胜任力草稿；具体样本为00401 ABC | 页面连续出现两条相同英文错误通知，未跳转答题页 | 实际examId=`1785027744745375431`，participantId=`1785027772270618331` | ✅ staging已修复 | FB-066；未发布草稿从在线列表隐藏、准备页禁用开始并中文提示、后端中文映射、防重复点击；该测评已获授权发布5维度40题，浏览器验收通过 |
| UF-003 | 2026-07-26 | 胜任力测评点击“详情”进入传统测评详情页，批量生成报告报“不支持的 repoCode” | 业务分流 / UI交互 | admin进入“测评管理”→胜任力测评“详情”→勾选已完成记录→“批量生成报告” | 初次发现；代码从未按assessmentType分流详情入口，判定为遗漏分支 | 影响所有胜任力测评的“详情”入口；截图样本为领导人员版“五配置受测者05” | 完成0份、失败1份；两次通知均为“生成报告失败: 不支持的 repoCode:” | exam=`1785060945494990462`，paper=`fc743d2b-0b72-49b2-803f-f285d62730ed`，assessmentType=`competency`，repoCode为空 | ✅ staging已修复 | FB-075专项3/3、前端全量18文件107项、build通过；staging从测评管理按标题查询并点击主“详情”后正确进入专用页，9类按钮真实E2E通过、传统控件隐藏9/9、PDF 615336 bytes |
| UF-004 | 2026-07-26 | FB-075部署后，已打开的传统详情旧标签页仍可停留在胜任力不兼容页面并再次调用传统批量生成报告 | 业务分流 / 跨页面状态 | staging直接访问或保留`/#/exam/exam/users/:examId/...`旧标签页 | FB-075只修测评列表主入口，未覆盖首页最近测评和旧URL直达 | 所有胜任力测评的旧URL、书签、历史标签页及首页最近测评入口 | 仍显示9个传统专属控件，点击批量生成提示“不支持的 repoCode” | exam=`1785060945494990462`，participant=五配置受测者05，截图URL为传统`exam/users`路由 | ✅ staging已修复 | FB-076专项由2失败/3通过转5/5；前端全量18文件109项；旧URL与主详情双入口真实E2E、专用页9类操作、传统控件隐藏9/9、传统generate-report调用0次 |

| UF-005 | 2026-07-28 | 00401胜任力测评二维码在手机扫码后页面空白，不能直接进入考生信息页 | 业务分流 / 移动端入口 | 生产在线测评列表打开00401开放测评二维码→手机扫码 | 首次生产使用发现；胜任力测评从设计起不关联传统题库，二维码逻辑未覆盖该分支 | 影响全部`assessmentType=competency`且`isOpen=1`的开放测评；生产当前4条均满足 | 二维码URL以`/.../:stuFlag/`结尾，缺少路由必填`:repoCode`，Vue无法匹配组件，`#app`为空且未发API请求 | 生产样本exam=`1785206588788205768`；4条胜任力在线记录`repoCode`均为空 | ✅ 本地已修复，待部署 | FB-102统一将胜任力虚拟题库解析为00401；开放二维码、封闭二维码和直接导航共用回退。专项RED 3/3→GREEN 3/3，全量130项及production build通过，本地8089真实路由显示考生信息表单 |
| UF-006 | 2026-08-12 | 一期胜任力结果页的完整答卷无法勾选，“批量生成报告”永久禁用且无可操作提示 | UI交互 / 业务门禁 | admin进入一期胜任力结果页→尝试勾选唯一完整答卷→点击“批量生成报告” | 按钮从一期接入时即被前端硬编码关闭；后端十页渲染器后来完成，但前端门禁未同步移除；公共Axios拦截器还会丢弃后端业务错误文本 | 影响所有`competency-frontline-phase1-v1`结果页 | 完整答卷复选框不可选、批量生成按钮禁用；页面仍提示“模板接入前保持关闭” | exam=`1786508394008352244`，paper=`93a26b3b-047d-4ad9-890a-3b4fe6d042e4`，is_complete=1，overall=34.125，validity=good | ✅ staging已修复 | FB-113；完整答卷现可勾选并发出生成请求，报告API窄范围保留业务错误；真实浏览器E2E确认按钮启用且UI显示“一期正式报告内容尚未完成双重批准”。内容包仍为draft，未绕过正式门禁 |
| UF-007 | 2026-08-12 | 交付的Word模板显示长`{{...}}`内部占位符，字段换行、下划线和固定图片互相遮挡，客户无法正常维护 | UI视觉 / 模板可维护性 | Microsoft Word直接打开一期最终运行模板并显示格式标记 | FB-116将内部长占位符直接写进客户可见版式后产生的回归 | 影响所有维护一期Word模板的客户和实施人员，长字段和十维诊断最明显 | 个人信息字段跨行，维度字段挤压布局，固定图形难以调整；截图未显示运行时替换后的正式PDF | 当前运行模板；用户确认采用Microsoft Word内容控件，页面显示正常示例值、字段Tag隐藏 | ✅ staging已修复 | FB-117：49个可见占位符改为正常示例值+隐藏内容控件Tag；重复生成SHA一致；staging真实Word→LibreOffice报告为A4 10页、无未替换字段，模板49 Tag/12图表/0可见token |
| UF-008 | 2026-08-13 | 在报告模板页面上传修改稿时提示`内容控件Tag重复：dimension.competency-b1-03.score` | 模板契约 / 数据一致性 | staging“报告模板”→选择`260813胜任力测评报告模板修改稿（房）V2.2.docx`→上传并生效 | 当前生效模板校验通过；问题仅存在于新修改稿，属于Word编辑过程中复制或误设内容控件Tag | 仅影响该候选DOCX，不影响当前生效模板和已生成报告 | 候选文件中至少两个内容控件共用“自律性得分”Tag；服务端严格校验拒绝覆盖 | 候选文件未保存到工作区或服务器，截图只能确定重复Tag，无法确定另一个控件应归属哪个维度 | 🔍 已定位，待修订候选DOCX | 当前上传门禁按设计工作：失败前不替换模板。建议从当前有效模板重新修改，或在Word开发工具→属性中将错误控件Tag改回其所在维度的`.score` |
| UF-009 | 2026-08-13 | 一期胜任力测评只配置姓名、性别、手机号，首次进入考生信息页却显示姓名、手机号、年龄、性别、单位、岗位，必须二次修改保存后才一致 | 数据一致性 / 配置回填 | staging新建一期胜任力测评→个人信息仅选3项→首次保存→进入考生信息页 | 新建后首次即出现，不是已知历史回归；二次进入修改并保存后恢复 | 当前确认影响一期胜任力测评；尚未扩张到通用胜任力或传统测评 | 首次保存后的准备页多出年龄、单位、岗位；截图显示6项均必填 | staging一期胜任力测评；用户未提供examId | ✅ staging已修复 | FB-118：六项默认值只在首次切换类型时应用；真实页面首次保存请求和Detail均为`name,gender,telephone`，考生页立即只显示3项。临时测评与会话清理为0 |
| UF-010 | 2026-08-13 | 一期答题页选择答案后不自动下一题，暴露内部题号和“五级量表”，微信扫码后白屏 | UI交互 / 移动端入口 | staging“胜任力测验260813”：桌面答题；微信扫码打开开放测评二维码 | 当前测试首次反馈，是否历史回归不确定 | 当前确认影响一期胜任力测评；手机环境为微信内置浏览器 | 前89题选择保存后仍停留当前题；顶部显示`B1-02-Q02`和“五级量表”；微信扫码页面完全加载不出且无提示 | exam=`1786588375737209899`、截图题号=`B1-02-Q02`、微信Android WebView | ✅ staging已修复 | FB-119真实静态资源验证成功保存9→10，失败不跳，末题停留且submit=0；FB-120内部题号/量表标签不存在；FB-121无hash中转URL在微信Android UA下加载考生页且无溢出/错误 |
| UF-011 | 2026-08-13 | 从staging报告模板页面下载的V2 DOCX无法由Microsoft Word打开 | 模板兼容性 / 部署回归 | “测评管理→报告模板”下载DOCX→Microsoft Word打开 | V1模板此前可打开；两个V2版本均失败 | 影响V2内嵌Excel模板；V1不受影响 | Word提示“在试图打开文件时遇到错误” | V2首版SHA=`0f1a23a...`、Content Types修订版=`a2387516...`均失败 | ✅ V1回退已由用户确认 | staging恢复V1 SHA=`3b6a83fd...`；下载no-store且文件名`competency-phase1-report-3b6a83fd.docx`；2026-08-14用户确认Microsoft Word可正常打开，报告链12页通过。V2继续暂停 |
| UF-012 | 2026-08-13 | 胜任力结果列表、详情中的整体分和维度分小数位不统一，报告也应统一保留两位 | 显示格式 / 数据一致性 | 测评管理→胜任力结果→列表整体分、答题详情→维度得分；查看报告 | 页面直接输出API数值，3.75/4/4.625混合显示；报告大部分已格式化但得分合计仍直接输出 | 所有胜任力结果，截图为一期90题完整答卷 | 整体分显示38.125、40.75、35.5；维度分显示3.75、4、4.625 | staging一期结果列表与详情 | ✅ staging已修复 | FB-123：聚合分值统一两位，空值为`—`；真实浏览器列表34.13、10维详情和10个报告分值通过；逐题原始值/计分值及题数保持整数语义 |
| UF-013 | 2026-08-13 | 封闭模式添加测评人员时身份证号标注选填，但留空提交后端返回“缺少 idNumber 或 examId” | 业务契约 / 封闭测评人员 | 测评管理→测评者管理→新增→选择封闭测评→填写姓名/手机号、身份证留空→确定 | Excel导入已支持身份证为空时按手机号识别；手工新增仍要求身份证，前后端契约不一致 | 所有封闭测评的手工新增人员；Excel批量导入链不受该错误影响 | 红色错误“缺少 idNumber 或 examId” | 截图测评`260812胜任力测2`，姓名`小花`，手机号`12341234123` | ✅ staging已修复 | FB-124：手工新增与Excel导入统一；真实封闭测评新增空身份证人员成功，默认密码手机号后4位，清理后残留0 |
| UF-014 | 2026-08-17 | Word模板版式正常，但服务器生成PDF时雷达图说明与首个维度卡片重叠、后续区域出现大块空白和页码差异 | 模板兼容性 / PDF版式 | Microsoft Word打开V1模板正常→系统生成一期PDF→查看二级维度结果页 | V1模板可编辑，但Word与Linux LibreOffice对浮动文本框、组合图形、锚点和分页计算不同；此前只验证页数/文本/哈希，没有持续做视觉重叠门禁 | 当前一期Word→LibreOffice报告；截图对比Word第4页与PDF第5页 | 图示说明覆盖逻辑思维模块，卡片内容错位，空白框，Word显示4/9而PDF显示5/13 | staging复核paper=`93a26b3b-047d-4ad9-890a-3b4fe6d042e4`、report=`60f14a12-bef8-42d3-a22a-831ea260b2c2` | ✅ staging已修复，待用户Word确认 | FB-125最终模板SHA=`42866f27...`；staging LibreOffice 24.2真实生成A4 9页、606276 bytes，逐页无重叠/空白/标题错序，十维10/10且无错误总页数。模板管理下载有效；production未修改 |
| UF-015 | 2026-08-17 | 用户检查`123-658dc083-6216-4373-93b3-a7b1d188ec44-胜任力临时测试报告`时仍看到旧PDF版式问题 | 历史产物 / 模板生效范围 | staging结果页下载已在2026-08-13生成的报告→打开PDF | FB-125只替换模板并重生成指定验收paper，不会自动改写其他已完成历史PDF；该报告仍指向旧文件 | 影响模板更新前已生成且未执行force重生成的报告实例；新生成报告不受影响 | 数据库/文件SHA一致但旧文件仍为LibreOffice 24.2 A4 12页；第3页近空白、第5页重叠、页脚`/13` | paper=`658dc083-6216-4373-93b3-a7b1d188ec44`，exam=`1786588375737209899` | ✅ staging已修复，待用户复核 | FB-126仅force重生成该paper：旧SHA=`42bc4ae3...`/12页→新SHA=`74c52f8a...`/A4 9页/601397 bytes；实例、文件、下载三方一致，十维10/10，逐页无原空白/重叠/错序，成功regenerate审计1。其他历史报告未批量处理，production未修改 |
| UF-016 | 2026-08-17 | 重生成后的维度环形得分图显示在图示说明上方，没有落在对应维度结果表格左侧空白单元格内 | 模板兼容性 / 图表定位 | staging下载FB-126新报告→查看二级维度详情首块“逻辑思维” | 十个环形图中chart3仍是表格单元格内的`wp:anchor`浮动对象，chart4–chart12已经是`wp:inline`；LibreOffice将chart3相对段落偏移到图示说明区 | 实际只影响首个逻辑思维环形图；其余九图已在表格内 | 3.50环形图出现在图示说明区域；逻辑思维表格左侧得分单元格为空 | paper=`658dc083-6216-4373-93b3-a7b1d188ec44` | ✅ staging已修复，待用户复核 | FB-127将chart3改为其原表格单元格内的`wp:inline`；结构测试确认chart3–12全部inline+in-cell。重生成PDF A4 9页、601416 bytes、SHA=`948eb7ca...`，页面截图确认逻辑思维3.50图位于左侧得分格，其他维度图保持正常；production未修改 |
| UF-017 | 2026-08-18 | 当前报告模板无法使用Microsoft Word打开 | 模板兼容性 / OOXML Schema | 从staging模板管理下载FB-129版本→Microsoft Word打开 | 损坏首次发生在将Word原生V1交给LibreOffice重存：LO改写39个部件、新增6个、删除51个；之后所有候选即使SDK错误归零也被Word拒绝。FB-129 chart1另新增2个Schema错误，但不是唯一根因 | 影响LO重存链产生的FB-125～129模板；PDF生成仍可用 | 本机Word 16实证：Git V1基线可打开，LO重存稿及后续候选全部失败；SDK零错误模板仍失败，证明Schema零错误不是充分条件 | staging旧SHA=`6f43de7c...`；最终Word原生修正版SHA=`0899d497...` | ✅ STAGING GREEN | FB-131从用户已确认可打开的Git V1基线出发，只定向修改document/chart1/footer4，不再全包LO重存。空模板和填充DOCX均由Word 16真实打开且关闭后SHA不变；staging API下载同样Word打开成功。LibreOffice报告保持A4 9页，production未修改 |
| UF-018 | 2026-08-18 | 用户连续提交名为`competency-phase1-report-0899d497.docx`的“修复模板”请求检查并替换系统 | 模板版本 / 文件完整性 | 附件与staging模板比较→Word打开→上传门禁→图表/Tag审计→修复同名文件→原子替换系统→重生成报告 | 新附件仍只沿用官方文件名，实际SHA=`36f3fe94...`；Word提示损坏、沟通表达诊断Tag重复、chart1回退饼图、表格9→5。其可见修改主要是运行时会覆盖的示例诊断文案，无需迁移 | 仅影响附件；系统官方模板本身正常 | 原附件Word失败、上传门禁失败；修复后Word打开、契约通过、LibreOffice 9页 | 原附件542140 bytes；修复后531586 bytes/SHA=`0899d497...` | ✅ 已修复并替换系统 | 原附件备份到tmp后，同名附件已替换为官方Word原生修正版；staging原子替换并备份旧模板，API下载SHA/契约/MIME/cache通过，本机Word实开且无写回；用户paper重生成A4 9页、SHA=`ba768e17...`，production未修改 |
| UF-019 | 2026-08-18 | 用户提交`胜任力测评报告模板.docx`并要求先认真检查问题 | 模板审查 / 版本回退 | 文件身份→Word实开→上传门禁→Tag边界→外链→真实填充PDF逐页重审 | 原附件重复Tag和损坏已由FB-134修复；持续学习Tag边界与外链由FB-135/136修复；客户确认保留3D饼图；FB-137/138完成模块与跨版本分页兼容 | 最终客户模板已部署staging，production未修改 | 首次全模块候选在LibreOffice 24.2生成12页并出现稀疏页，立即回滚；删除一级/二级区之间的叠加显式分页后，真实报告为11个非空页面且10模块均不拆页 | 最终模板SHA=`9bf1cbb7...`；真实PDF SHA=`8ae5a386...` | ✅ staging已修复并部署 | Word 16、模板API 51/49/12/0、LibreOffice 24.2 A4 11页、三方PDF SHA、逐页视觉和清理/health均通过 |
| UF-020 | 2026-08-18 | 环形图中心的分值数字未居中 | 模板视觉 / 图表标签定位 | 查看staging一期Word报告PDF的二级维度得分图 | 客户模板十个图表继承了各不相同的手工标签坐标，标签框中心未对齐图心 | 当前客户模板chart3–12；不影响一级图表或分值数据 | 截图中`3.50`偏离白色圆心 | 真实paper=`658dc083-6216-4373-93b3-a7b1d188ec44` | ⚠️ 原结论被UF-021纠正 | FB-139只验证OOXML坐标，未验证LibreOffice像素；此前“10/10居中”不成立，最终闭环见FB-140 |
| UF-021 | 2026-08-18 | 再次检查发现部署生成PDF后仍有很多环形图数字未居中，Word中正常 | 部署/环境 / 图表渲染差异 | Word模板中查看居中→staging LibreOffice 24.2生成PDF→查看二级维度图表 | FB-139结构坐标测试通过但未测量PDF实际像素，属于验证盲区；Word正常，LibreOffice PDF异常 | 真实报告chart3–12共10图，旧PDF像素门禁10/10失败 | 无接口错误；旧PDF偏移x=`0～15px`、y=`-3.5～20.5px` | paper=`658dc083-6216-4373-93b3-a7b1d188ec44`，最终PDF SHA=`70821dd0...` | ✅ staging已修复并部署 | FB-140仅在LibreOffice转换前独立校准十图；最终11页PDF的180 DPI像素中心测试10/10通过，Word模板保持不变 |
| UF-024 | 2026-08-18 | 最新模板调整后要求分值保留两位小数、一级表格数字必须替换，并重新校验部署staging | 模板契约 / 数据一致性 / 部署 | 用户保存最新DOCX→本地完整门禁→staging替换→同一真实paper强制重生成 | 模板重存恢复显式分页和chart7 anchor，并使chart3–12丢失`0.00`；两个一级Tag仍各2处，旧截图中的静态汇总值不是当前模板控件缺失 | 最新客户模板及其LibreOffice运行时校准；production不受影响 | 最终PDF为10页，汇总表/分析区均为真实`3.50/3.60`、静态值0，十图像素中心10/10通过 | template=`50b238cc...`；paper=`658dc083-6216-4373-93b3-a7b1d188ec44`；PDF=`37e11524...` | ✅ staging已修复并终验 | FB-147～149 GREEN；三方SHA、逐页视觉、临时清理、服务/health和日志均通过；production未修改 |
| UF-025 | 2026-08-18 | staging生成报告与当前下载模板的字体、颜色、分页、表格布局及一级分值不一致 | 历史产物 / 模板生效范围 | staging当前模板已更新→结果页点击生成→下载`456-4bb5506b-...-胜任力临时测试报告`→与当前模板比较 | UI“生成报告”固定发送`force=false`，已有completed实例时后端直接复用历史PDF；最初的模板示例值假设已由实例时间和文件证据排除 | 影响模板更新前已有completed报告；新paper首次生成不受复用缺陷影响 | 旧PDF创建于2026-08-13，为A4 12页且页脚仍`/13`；修复后新PDF为当前模板A4 10页，汇总/饼图/分析均为`3.10/2.75` | paper=`4bb5506b-3ba5-4c43-b7e7-2a117287bc3b`；new PDF SHA=`b80c58a4...`；template=`50b238cc...` | ✅ staging已修复并部署 | FB-151前端=`9e51b0b4...`；DB/文件/API下载502212 bytes与SHA一致，十页视觉无空白/重叠/裁切/错序，过程文件和短时会话清零，production未修改 |
| UF-026 | 2026-08-18 | 模板再次更新，要求重新部署并重点检查环形图数字居中 | 模板版本 / PDF图表渲染 | 最新DOCX→结构门禁→本地LibreOffice→staging真实长文案paper→180-DPI十图像素门禁 | Word重存恢复一级/二级区间显式分页，并将图表标签字体改为约11pt，旧模板校准值失效 | 最新客户模板及LibreOffice转换路径；Word原模板不应用运行时校准 | 本地短夹具通过后，真实长文案首轮9/10失败、第二轮仅chart11失败；第三轮当前报告10/10通过 | template=`19c0f1d4...`；paper=`4bb5506b-3ba5-4c43-b7e7-2a117287bc3b`；PDF=`3bd9b279...` | ✅ staging已修复并部署 | FB-152；Word 16实开10页且无写回，模板四门禁、Go全量/Linux build、真实A4 11页、DB/文件/API一致、逐页视觉、清理/health/log均通过；production未修改 |
| UF-027 | 2026-08-21 | 效度良好时不显示提示整段文字，仅效度存疑时显示 | 报告条件展示 / Word与Vue一致性 | 生成一期报告→效度状态为good或questionable→查看总体评价页提示段 | 当前Word和Vue一期模板均无状态条件，始终渲染正式效度文案 | 一期Word/PDF主路径及Vue/Chromium兜底路径 | good仍显示“提示：本次测评作答效度良好……” | 用户截图；触发数据为完整答卷且效度状态good | ✅ staging已修复并终验 | FB-153：真实good PDF整段无提示，真实questionable PDF保留完整提示；部署Vue在桌面/手机Chromium分别验证good节点缺失、questionable精确文案存在；内外health正常，production未修改 |
| UF-028 | 2026-08-24 | 环形图中数字应居中且不加粗 | 报告视觉 / LibreOffice图表渲染 | 查看一期Word报告PDF二级维度详情页 | 原生数据标签位置随分值变化且继承混合粗体；单份短夹具不能覆盖真实长文案和不同分值 | 一期Word/PDF主路径的chart3–12十个环形图 | 截图中`2.75`、`2.13`偏离白色圆心且呈粗体，无接口错误 | 用户截图；真实good与questionable完整答卷 | ✅ staging已修复并逐页终验 | FB-154删除全部原生标签，只保留固定中心的非粗体两位小数；两份真实A4 11页报告均180-DPI 10/10通过，逐页无额外标签、空白、重叠或裁切；服务/health/最终窗口5xx正常，production未修改 |
| UF-029 | 2026-08-24 | 报告雷达图显示不完整，只有一个虚线框 | 报告视觉 / LibreOffice雷达图 | 查看一期Word报告PDF“各维度得分情况”雷达图 | 线上模板chart2含两套雷达轴；值轴未统一显式设置0-5和majorUnit=1，LibreOffice只画最外层 | 一期Word/PDF主路径chart2；维度名称、折线和分值仍存在 | 应有五层同心网格，实际只有一个最外层虚线多边形，无接口错误 | 用户截图；real good/questionable两份完整报告 | ✅ staging已修复并逐页终验 | FB-155统一两套值轴为0-5、主单位1并保留主网格线；两份真实雷达页均显示五层完整同心网格，22页逐页无空白/重叠/裁切，环形图10/10回归保持，production未修改 |
| UF-030 | 2026-08-25 | 手机端五个选项框左侧需要对齐 | UI视觉 / 移动端答题 | 微信内打开00401一期答题页→查看五级选项 | Element UI为相邻的`is-bordered`单选框默认增加10px左边距，第一项没有该边距 | 00401一期移动端五个选项；桌面/平板布局保持 | 五个卡片左边界为`29,39,39,39,39`，截图无接口错误 | staging=`20.200.136.133`，微信内置浏览器，90题答题页 | ✅ staging已部署并实测 | FB-156：真实staging Chromium 390px五项均left=29/right=361/width=332/height=48且无溢出；平板3列、桌面5列及触控尺寸保持；部署index原始字节SHA与本地一致 |
| UF-031 | 2026-08-25 | 批量下载报告需完善，目前逐个下载，不是压缩包 | UI交互 / 文件下载 | 00401胜任力结果页→勾选多份完整答卷→批量下载 | 该功能从实现起即在前端循环单份下载，并非近期回归 | 所有胜任力结果页批量下载；单份下载保持PDF | 浏览器连续触发多份PDF下载，没有一个统一ZIP；无后端错误信息 | 最小触发数据为至少2份已生成的完整报告 | ✅ staging已部署并实测 | FB-157：staging真实认证请求一次下载2份，返回单个881097-byte `application/zip`，ZIP内2份PDF有效且名称唯一，download审计增加2；短时会话和发布临时文件清零 |
| UF-032 | 2026-08-25 | 测评者管理“是否学生”筛选不能使用 | 业务筛选 / 测评者管理 | 测评者管理→是否学生选择“是”→查询 | 该页面已有筛选控件，但后端列表从实现起未读取`stuFlag`，不是近期回归 | 所有封闭测评者管理列表；“是/否”两个值均受影响 | 选择“是”后列表仍同时显示“是”和“否”，截图无接口错误 | staging当前有效`el_tester`分布为stu_flag=0共19条、stu_flag=1共8条、NULL=0 | ✅ staging已部署并实测 | FB-158：真实API“是”返回8/8学生，“否”返回19/19非学生；COUNT/rows一致，短时会话清零 |
| UF-033 | 2026-08-25 | 客户新上传一期报告模板后，饼图和雷达图显示仍有问题 | 模板回归 / LibreOffice图表 | 报告模板页面上传新稿→对完整一期答卷强制生成→查看一级饼图和十维雷达图 | 2026-08-24 14:47已验证模板后连续上传多稿；当前稿16:30:51生效，属于客户重存/上传后的回归 | 所有使用当前staging模板新生成的一期Word/PDF；已用两份不同真实结果复现 | 无接口错误；饼图固定显示3.75/3.70且与正文分数不一致，雷达图无维度名称、数字压线；另有外部Excel关系 | template旧SHA=`896b59e5...`; papers=`ff08cb88...`小鱼、`0aae19a9...`小米 | ✅ staging已修复并逐图终验 | FB-160/161/162：模板SHA=`fa6c59a3...`；小鱼饼图3.10/3.25、小米2.65/2.38均与正文一致；两份雷达十维名称+分值清晰、五层网格完整、11页非空、零外链 |
| UF-034 | 2026-08-30 | 新生成一期报告实际11页，但页脚显示`1～10 / 12` | 模板回归 / PDF页码 | 客户11:17上传新模板→强制生成小鱼/小米报告→查看末尾维度页 | 已验证模板曾移除不可靠总页数；11:17上传稿重新加入`PAGE / NUMPAGES`，属于模板上传回归 | 使用该模板新生成的一期报告；小鱼、小米均复现 | 修复前PDF物理11页但正文显示`1/12`至`10/12` | template旧SHA=`cd712b8e...`; papers=`ff08cb88...`、`0aae19a9...` | ✅ staging已修复并实测 | FB-164采用方案1：模板/运行时仅保留`PAGE`，上传门禁拒绝`NUMPAGES`。两份真实PDF均为A4 11页，正文页码精确`1～10`、分数页码0；DB/文件SHA一致，末页实图仅显示`10`。 |
| UF-035 | 2026-09-01 | 报告模板在Word中随意修改一个地方后就无法上传，提示模板不得包含外部Excel链接 | 模板兼容性 / 上传流程 | 报告模板→下载DOCX→Word修改任意内容并保存→上传 | 当前零外链模板可用，Word重存后稳定失败，属于编辑兼容性缺口 | 所有需要客户继续维护的一期Word模板 | 页面同时显示“模板校验失败：模板不得包含外部Excel链接” | staging；截图显示当前模板51控件/12图表/0可见占位符 | ✅ staging已修复并实测 | FB-165上传前定向清除图表外部关系及`externalData`，再执行全部严格契约并仅保存零外链DOCX。真实模拟Word重存稿由1关系+1 externalData清理为0/0，响应removed=2、valid=true；真实报告生成和服务终验通过。 |
| UF-036 | 2026-09-01 | 报告个人信息字段应依次紧凑排列、上部对齐且不产生空行 | 报告布局 / 动态字段 | 生成一期报告→查看个人信息区 | 模板采用固定两列槽位，运行时只删除未配置单元格、不重排保留字段 | 所有requiredFields子集；截图为小鱼`name,gender,telephone`，目标示例为姓名、手机号逐行及时间/时长同排 | 手机号被保留在右列，字段视觉顺序断裂并出现大块留白 | paper=`ff08cb88-...`; requiredFields=`name,gender,telephone` | ✅ staging已修复并实测 | FB-166真实小鱼报告显示姓名→性别→手机号逐行左对齐，时间/时长同排同基线，无中间空行；A4 11页、页码1～10，数据库/文件SHA一致。 |
| UF-037 | 2026-09-01 | 个人信息顺序和位置应由客户Word模板决定，单位/岗位配置后必须显示 | 模板可维护性 / 动态布局 | 客户在Word中移动个人信息字段→上传→生成报告 | FB-166部署版由后端固定姓名/性别/年龄/手机号/单位/岗位顺序，不能体现模板重排 | 一期报告全部requiredFields组合 | 模板仍能控制样式，但字段顺序被程序覆盖 | 用户明确选择方案B | ✅ staging已修复并实测 | FB-167后端+模板同步部署。小鱼子集按模板显示姓名→性别→手机号→时间/时长且无空行；全字段真实报告显示姓名→年龄→性别→手机号→单位→岗位→时间/时长，单位岗位值均存在。 |
| UF-038 | 2026-09-01 | 模板个人信息各行间距应保持一致 | 报告布局 / 模板样式 | Word打开模板或生成全字段报告→查看个人信息行 | FB-167拆分独立行时保留各原单元格段落属性，单位/时长原本缺段前段后 | 全字段及任意requiredFields子集 | 单位、岗位及其他字段之间的垂直间距不完全一致 | 旧模板SHA=`bbb44705...` | ✅ staging已修复并实测 | FB-168模板统一六个身份字段和时间/时长为before/after=156、line=144；服务器结构审计8个单元格仅1种spacing，真实小鱼/全字段报告实图通过，无空白页。 |
| UF-039 | 2026-09-02 | 模板中调近饼图数值位置、调大雷达图分值字体后，生成报告仍使用旧格式 | 模板可维护性 / 图表样式 | staging上传修改后的DOCX→对完整一期结果强制生成→查看一级饼图与十维雷达图 | 模板原本仅有格式瑕疵；FB-160/161运行时为修复图表兼容性而重建整个标签块，导致后续模板样式无法生效 | 所有使用一期Word模板生成的胜任力报告 | 无上传/生成错误；模板饼图手工坐标和雷达12pt字号被生成程序替换成固定outEnd/8pt | paper=`287347a7-d8e1-4606-9ba0-f352f2e848af`; template SHA=`1e9b88ee...` | ✅ staging已修复并实测 | FB-169后端部署后，真实凯迪报告A4 10页：饼图左绿2.90/右青3.18贴近图内，雷达十个12pt分值与十维名称、五层网格清晰；DB/文件SHA一致，服务与日志通过。 |
| UF-040 | 2026-09-02 | 饼图数值不能压在图形色块上 | 报告视觉 / 饼图标签定位 | staging强制生成一期报告→查看一级维度3D饼图 | FB-169按模板保留了`inEnd`和手工坐标，真实PDF证明两个分值进入扇区内部 | 所有使用当前一期模板新生成的报告 | 无接口错误；左绿2.90、右青3.18覆盖在饼图色块上 | paper=`287347a7-d8e1-4606-9ba0-f352f2e848af`; old template SHA=`1e9b88ee...` | ✅ staging已修复并实测 | FB-170模板SHA=`54b167fc...`已部署；真实凯迪报告A4 10页，左绿2.90/右青3.18分别位于饼图左右外侧并有短引导线，不覆盖色块；映射、雷达12pt分值、十维名称和五层网格保持正常，DB/文件SHA一致。 |
| UF-041 | 2026-09-03 | 报告个人信息应保持双列、依次紧凑排布、上部对齐且无空行 | 报告布局 / 动态字段 | 生成一期报告→查看个人信息区 | FB-167按此前产品决定让每个身份字段独占模板整行，无法满足新的双列紧凑要求 | 所有一期报告及任意requiredFields子集 | 无接口错误；字段纵向占行过多，右列顶部出现可利用空位 | 截图覆盖`name,gender,telephone`与`name,telephone`；用户确认“姓名独占首行”方案 | ✅ staging已修复并实测 | FB-171后端已部署；真实凯迪报告按`name,age,telephone,gender`显示姓名整行、年龄/性别双列、手机号左列、时间/时长双列，连续上对齐且无空行；A4 10页，DB/文件SHA一致。 |
| UF-022 | 2026-08-18 | 考生中途退出，重新登录后应保留已答题并从第一道未答题继续 | UI交互 / 跨页面状态 | 001/002/003/00401答题10题→退出→重新登录同一考生→恢复进行中试卷 | 未知；代码核验确认后端/准备页已恢复同一试卷，MBTI已定位第一道未答，传统与00401页面原先仍默认第一题 | 用户确认范围为全部001/002/003/00401；具体测评ID/考生ID未知 | 无错误信息；已答数据可恢复，但页面位置不连续 | 最小复现条件：同一参与者、同一进行中试卷、前10题已答、第11题未答 | ✅ staging已部署并实测 | FB-141：本地四类型回归4/4；staging真实00401同一考生答10题后重新登录，paperId不变、已答10/未答80，真实Chromium直接显示第11题；测试人员/试卷/题目/结果清理均为0 |
| UF-023 | 2026-08-18 | 手机端考生公开流程顶部显示“人才综合素质评估系统”，要求删除 | UI视觉 / 移动端浏览器标题 | 手机浏览器打开考生信息、准备、答题、结果或完成页面 | 现有全局`App.vue`始终使用`VUE_APP_TITLE`，此前未区分管理端和考生公开流程 | 用户确认仅考生公开流程删除，管理后台保留 | 浏览器顶部标题栏显示系统名称和站点IP，无接口错误 | 截图站点IP=`39.106.61.48`；具体测评ID未知 | ✅ staging已部署并实测 | FB-144：13个考生公开路由使用空标题，管理路由保持原标题；staging真实Chromium经登录→准备→00401答题的`document.title`为空，部署index SHA与本地测试包一致，production未修改 |
| UF-042 | 2026-09-20 | 客户截图中的十维得分/常模组合图在系统生成PDF中未显示 | 模板兼容性 / LibreOffice组合图 | 客户Word模板可见组合图→系统value-only填充→staging LibreOffice 24.2生成PDF→查看报告概览后续页 | 客户原始重存稿在本机LibreOffice 26.2同样缺图，说明不是FB-191部署新引入；此前只验页数、文本、分值和详情分页，未强制检查该组合图像素存在 | 影响当前v2模板生成的全部PDF；Word中可见，LibreOffice路径缺失 | 无API或数据错误；`chart2.xml`仍有2系列和完整数值，但PDF仅显示组合对象中的等级图片，柱线图整体缺失 | template=`0d079104...`；paper=`0aae19a9-...`；report=`bc9a222a-...` | ✅ staging已修复并实测 | FB-192将Word 2010组合对象拆为普通内联竖向等级图和普通内联动态chart2；真实LibreOffice 24.2 PDF第4页显示10个维度、10根分值柱及常模折线，分值/常模逐项与数据库一致。A4仍为10页，详情2/2/2/2/2，v1不变。 |
| UF-043 | 2026-09-20 | “胜任力综合表现”生成报告与客户模板差异较大：维度名称消失且整段变为粗体 | 模板值替换 / 样式保真 | 客户模板中优势/待发展为“粗体维度名：常规描述”→系统生成报告→对比同页 | v2接入以来即存在；单一内容控件的首个run为粗体，通用替换器只保留首个run样式 | 所有存在非空优势项或待发展项的v2报告；空状态不受影响 | 模板显示`数字应用：`等动态名称且仅名称粗体；生成稿仅显示规则正文且整段粗体。规则正文随实际分值等级变化属于正确业务差异 | 用户对比截图；真实paper=`af33d5c6-...`选出3项优势/2项待发展 | ✅ staging已修复并实测 | FB-193部署后真实PDF显示优势`敬业奉献/持续学习/合作意识`、待发展`沟通表达/逻辑思维`；五个名称加冒号均粗体，批准正文均常规字重且逐句与数据库一致。A4 10页、组合图和详情分页无回归。 |
| UF-044 | 2026-09-20 | 美化报告概览页，重点改善总体摘要框、环图中心和十维图左侧等级刻度 | 报告视觉 / 模板层级 | 查看v2报告“报告概览”页并对照用户红框 | 现有内容与数据正确，但客户初稿沿用弱层级边框、中心文字紧凑和70×447低分辨率等级图 | 所有v2报告概览页；不涉及评分、文案选择或详情页 | 总体摘要框缺少视觉锚点；环图中心“总体评价+分值”拥挤；竖向等级条偏窄且像素感明显 | 用户标注截图；模板SHA=`3b88616f...` | ✅ staging已部署并实测 | FB-194：真实概览页为绿色左强调摘要卡、9pt灰色`总体得分`+16pt绿色分值、180×650高清圆角色阶；A4 10页和全页接触表通过，组合图/优势待发展/详情页及数据无回归。 |
| UF-045 | 2026-09-20 | 有效的v2报告模板无法在UI上传 | 模板管理 / 版本分流 | 报告模板管理→选择`competency-phase1-report-v2.docx`→上传并生效 | v2运行链完成后，管理页仍固定一期v1路径和v1注册表，属于遗漏版本分支 | 所有v2模板；v1模板上传不受影响 | 后端实际拒绝：`模板包含未支持的内容控件Tag：dimension.cooperation.score` | 附件685597 bytes/SHA=`f98599939e3bf7923abf8bd457e30259cde3fa68d6c4e9dc9d3288de0a2eae16`，77控件/60唯一Tag/12图表/零外链/零公式 | ✅ staging已修复并实测 | FB-195新增独立v2元数据/下载/上传路由和UI API；真实multipart上传返回60/60字段、12图表、valid=true并生成原子备份，下载SHA一致。随后真实报告A4 10页、组合图和详情分页通过；v1模板SHA未变。 |
| UF-046 | 2026-09-20 | 十维柱形图相同等级的分数显示不同颜色 | 图表样式 / 动态业务规则 | 生成不同分数报告→查看十维分值柱颜色 | 模板按固定数据点位置保存颜色，运行时此前只替换数值，因此颜色不随分值变化 | 所有v2报告的十维柱形；常模折线及其他图表不受影响 | 同属合格档的柱形可能一深一浅，颜色无法表达等级 | 用户确认五档：优秀≥90、良好70–<90、合格30–<70、薄弱10–<30、不足<10；常模线保持橙色 | ✅ staging已修复并实测 | FB-196真实报告中68.75/62.50等合格柱统一浅绿，75～87.50良好柱统一绿色，90.63优秀柱为深绿；橙色常模线不变。边界测试覆盖90/70/30/10及下界。 |
| UF-047 | 2026-09-22 | 胜任力系统导出的数据表与新版评分/报告不一致 | 导出数据一致性 / 版本分流 | 测评管理→胜任力测评→导出汇总或导出原始答题→与v2报告比较 | v2 result_run完成后，导出仍读取旧`el_competency_result/group_result/dimension_result`，属于遗漏消费方 | 一期v2导出；通用胜任力与传统测评暂不改变 | 导出仍是旧整体分、评价均值、两组和旧维度身份，缺三模块、百分制、常模、比较及四版本 | 用户确认：三Sheet同步升级、仅completed v2、不回退v1、完整输出、无客户样表 | ✅ staging已修复并实测 | 两个真实导出入口返回字节一致工作簿；3个v2 run对应汇总3行、逐题270行、字典90行，75/20/14列完整，42项数据库事实逐项一致。无v2 run的真实测评三个Sheet均仅表头。 |
| UF-048 | 2026-10-01 | 报告优势项和待发展项不应按等级类别过滤，应统一按十维分值排序 | 业务规则 / 报告选择器 | 生成v2报告→查看“胜任力综合表现” | FB-177按旧材料将良好/优秀归入优势、其余归入待发展，导致列表数量随等级变化 | 所有基层员工v2报告；模板仍固定3个优势槽和2个待发展槽 | 无接口错误；展示项不一定是全体十维中的固定最高3项和最低2项 | 用户确认：跨等级统一按分值；最高3、最低2；同分沿用固定维度顺序；正文使用该维度对应等级的完整表现评估文案 | ✅ staging已修复并实测 | 真实十维分值报告按75/71.875/68.75选出自律性/成就导向/计划执行，按40.625/53.125选出敬业奉献/逻辑思维；53.125同分按固定顺序取逻辑思维。五段完整批准文案逐项匹配数据库，PDF A4 10页且DB/文件SHA一致。 |

## Classification Summary

- 业务规则 / 报告选择器：1（staging已修复）
- 功能缺口 / UI交互：1
- 业务规则 / UI交互：1（已修复）
- 业务分流 / UI交互：1（staging已修复）
- 业务分流 / 跨页面状态：1（staging已修复）
- 业务分流 / 移动端入口：1（本地已修复，待部署）
- UI交互 / 业务门禁：1（staging已修复）
- UI交互 / 文件下载：1（staging已修复）
- 业务筛选 / 测评者管理：1（staging已修复）
- 模板回归 / LibreOffice图表：1（staging已修复）
- UI视觉 / 模板可维护性：1（staging已修复）
- 模板契约 / 数据一致性：1（已定位，待修订候选DOCX）
- 数据一致性 / 配置回填：1（staging已修复）
- UI交互 / 移动端入口：1（staging已修复）
- 模板兼容性 / 部署回归：1（V1回退已确认，V2暂停）
- 显示格式 / 数据一致性：1（staging已修复）
- 业务契约 / 封闭测评人员：1（staging已修复）
- 模板兼容性 / PDF版式：1（staging已修复，待用户Word确认）
- 历史产物 / 模板生效范围：2（均已在staging修复，待用户复核）
- 模板兼容性 / 图表定位：1（staging已修复，待用户复核）
- 模板兼容性 / OOXML Schema：1（staging已修复并经Word实证）
- 模板版本 / 文件完整性：1（附件已修复并替换系统）
- 模板版本 / PDF图表渲染：1（staging已修复并部署）
- 模板审查 / 版本回退：1（附件不合格，未替换系统）
- UI交互 / 跨页面状态：1（本地已修复，未部署）
- 模板契约 / 数据一致性 / 部署：1（staging已修复并终验）
- 模板契约 / 数据一致性：1（已定位，待修订候选DOCX）
- 模板可维护性 / 图表样式：1（staging已修复并实测）
- 模板兼容性 / LibreOffice组合图：1（staging已修复并实测）
