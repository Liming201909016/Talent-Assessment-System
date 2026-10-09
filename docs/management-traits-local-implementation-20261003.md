# 002本地实现与阻断交接（2026-10-03）

## 2026-10-05T01:55:31Z staging精确清理与有界LO现场诊断（运行源码未修改）

**底层根因仍UNVERIFIED，未业务修复/完整验收。** 接续已授权exact exam1791105048344426522清理，新受限完整currentDB备份SHA **8da183355cca1c2eb379ff5827abd0f079bbb42b883364d21cc75b03e7783611** 后按PK/归属/marker/createdat及actualFK事务删除，13dim4mod1receipt1run140qs1snapshot700bucket140pq与各1candidate/paper/profile/bundle/link/exam；11侧表逐0/owned0/privatePDF0/15FK。主78/70/1488/134494/295328/1349/27，两source各140140700，465旧PDF/源题/运行资产与非own数据指纹一致，完整12表最终SHA051795bd969650d0e55da676083078255331c66405b5fff0ee4ca9d0a7641d0b匹配09:09写前。

只新增C区[精确清理脚本](../scripts/db/management-traits-diag-owned-cleanup-20261005.sh#L1)及[LO合成runtime探针](../scripts/tools/management-traits-lo-runtime-probe-20261005.sh#L1)，各bash-n0/editor0、真实cleanup0/probe脚本0。真实LO三次exit **0/1/1**：root有效PDF552631bytes，liming及matching systemd无PDF/同195bytes stderr SHA **b6ecc0ff320848795c8f84e1e5b82e2c4867d9e80f4ef0305ebb7287514d3d54**，fixedDeploymentException/terminate/Unspecified Application Error；root也有javaldx警告。没有足够证据把原因归为权限/HOME/sandbox/字体/Java，原HTTP具体exitcode未保存，不回填本次1。系统LO可读、HOME四目录可写、实际sandbox关键项no，无kernel AppArmor/segfault/OOM命中，coredumpctl unavailable。

实际服务PID2002/启动01:36:55Z早于本轮首次SSH，源码SHA8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93、front/template/content保持；不重启主服务。新workspace/profile/payload/transientunit最终0，原rawstderr仅受限backup/lo_diag_20261005_8BorrQx7保留，目录0700/files0600最终errors0，正式备份保留。三服务active/healthok；本轮无浏览器、凭据、SQL新fixture、报告生成、评分重算、production/现代化task状态/6144全量或oracle/budget重跑。运行Go代码未变，**未重跑Go/build/前端**，不冒称新编译验收。

首次只读PS换行1064/本地Node引号失败、派生dim/modID非UUID造成清理工具备份后停止且未执行SQL、冻结稿marker未核过未上传、owned证据目录/新指纹权限收口纠正均保留，未掩盖成业务GREEN。完整路径/行数/主键/三环境安全事实与后续授权边界见[本轮主报告](management-traits-four-real-verification-20261003.md#L3)。下一仅只读bootstrap/syscall取证候选；共享配置/权限/依赖或业务logger改动必须另获用户批准，不能直接fix。

## MT-GUARD-AUDIT：真实 report_id 闭包修复（2026-10-03，仅本地）

**未验证先列明**：本轮没有SSH、真实数据库连接/写入、DDL、main element写入或部署；实际恢复库fresh复验仍未执行，不能称staging PASS或解除部署门禁。本机mysql/mysqld未找到、3306/23306无监听，专用审计DSN/Schema环境变量未设置；新增真实MySQL用例明确skip。前端/PDF/race/覆盖率未重跑。新测试文件仍有三处gofmt拼接空格差异及缺EOF换行，已达三次编辑上限停止，不碰其他文件格式。

### 根因、改动及完整11表保护闭包

- 上一实际恢复库完整Schema PASS11/159/67/21、11表15FK与两题本140/700是上阶段事实，本轮未重新查询。上阶段capture拒绝源于audit-only能力未识别，不能用本地测试覆盖历史实际失败。
- [实际audit DDL](../scripts/sql/management_traits_001_runtime.sql#L191-L200)严格只有id/report_id/actor_id/action/created_at；外键report_id→report_revision.id。生成[写入](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go#L168-L178)使用本次revision主键，下载[审计](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go#L228-L235)使用加载的r.ID；不是paper_id、run_id或旧报告主键。
- 仅[守卫生产文件](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L280-L465)和[新增回归](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard_audit_test.go)变动；无签名/API/模型/DDL/依赖变更，不删守卫、不添加冗余FK或伪paper_id。001 SHA保持7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8。
- [能力识别](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L337-L355)：仅精确canonical audit五列接受report_id关联，任意未知report_id-only marker仍拒绝。表/列ASCII白名单、长度和重复检查；数据值始终绑定。
- [完整gate与孤儿门禁](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L361-L384)：任一canonical revision/current/audit出现时，要求原11表全在，原样复用checkRuntimeSchema完整列/type/NULL/charset/collation/PK/索引/15FK/旧引用容量检查；不缓存或放宽validator。audit LEFT JOIN revision→run同时匹配run/paper/exam，missing revision、missing run或复合身份漂移一律受控错误；empty/AllLegacy/不相关scope也先检查，不以INNER JOIN吞孤儿后返回false,nil。
- [scope闭包](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L417-L426)：report_id IN revision.id，revision.run_id IN已解析paper/exam/participant作用域的run.id；每批最多1000绑定值。仍先解析legacy paper/owner/question/PDF存储关联，不推断repoCode或源题身份；不把推断exam扩大为整exam目标。
- 11表图：bundle→profile/snapshot间接；profile→exam；snapshot→paper/exam/owner；question snapshot→paper/paperQuestion；run→paper/exam/owner；dimension/module/receipt→run；revision→run及paper/exam；current→revision/paper；audit→revision→run。未改源题写入门禁/冻结锁/运行时写事务或旧PDF路径。
- 七core历史/capture兼容路径及全缺失legacy行为保留；完整报告合同出现后不能降为七表，部分安装/类型/FK漂移失败关闭。合法同exam历史兄弟卷不会因另一paper的报告被扩大保护。

### 真实影响清单（所有项均明确处理，无待定签名）

| 真实消费方 | 处理 | 理由/本轮证据 |
|---|---|---|
| CheckManagementTraitsLegacyScope | 同步修改 | canonical audit能力、完整gate、全局孤儿检查及真实revision/run闭包 |
| 新DDL形状/SQL/GORM回归及外部用例 | 同步新增 | 从未改001读取11表列和15FK，audit五列不捏造；独立SQL oracle与真实GORM执行 |
| [身份participant回退](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L121-L136)、[candidate登记](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L330)、[tester登录](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L444) | 无需修改 | 同一守卫签名/返回；public identity capture真实unit通过，不冒称HTTP数据库E2E |
| [Candidate Save分流](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_identity.go#L104-L147)、[Tester login分流](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_identity.go#L161-L180) | 无需修改 | 继续受控身份错误/新旧分流；原handler/router回归保持 |
| [全局旧路由inventory](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_guard.go#L27-L93)、[实际守卫调用](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_guard.go#L375-L413) | 无需修改 | exam save/state/delete/teamPDF/upload/generate/export；paper CRUD/create/fill/hand/review/result/score；candidate/tester CRUD/end/PDF/import/login/collection；MBTI save/submit/report/download仍同403/503失败关闭，不更改路由/JWT/旧DTO |
| [异步旧写租约交接](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_guard.go#L418-L430)及freeze lease | 无需修改 | 原租约/排空机制不变，不移除保护 |
| 新freeze/profile/create/detail/fill/submit/result/Worker、GenerateTestReport/ReadTestReportPDF | 无需修改 | 继续原完整Schema/来源/身份/时限/TEST门禁和事务；audit writer外键事实相同，不改权限或新链数据 |
| 原schema/fullgate/guard/identity/handler/router测试 | 无需修改 | 本轮未改既有fixture/断言，四包回归通过，保留原RED历史 |
| 001 SQL/15FK、models、前端/模板/客户原件、001/003/00401及历史002源/结果/PDF | 无需修改 | 无数据/API/业务计算/文件语义变更；不对历史002整库禁写，不碰无关脏工作区 |

### RED→GREEN与独立验证（事件含子项，非覆盖率）

| 证据 | 实际结果 |
|---|---|
| 先写真实11表/audit五列测试，再改生产代码 | 有效RED **10pass/35fail/1skip、exit1**；初次CRLF解析夹具失败已排除，不作为有效RED |
| 首轮最小生产修复 | 45pass/0fail/1skip、exit0；补错误/批次回归后最终62pass |
| 主代理独立新增测试：go test ./internal/service -run 'TestBugManagementTraitsGuardAudit\|TestManagementTraitsGuardAuditMySQLExternal' -count=1 -json -timeout120s | **62pass/10通过顶层/0fail/1skip、解析错误0、exit0** |
| 四包ManagementTraits/BugMTSchema聚焦（worker） | **5327pass/268通过顶层/0fail/3skip、解析错误0、exit0** |
| 独立CodeReviewer | **PASS、无blocking finding**；自行guard/schema3279pass/0fail/1skip、handler/router152pass/0fail，exit0 |
| 主代理最终全量：go test ./... -count=1 -json -timeout 180s | **6039pass/638通过顶层/0fail/8skip、解析错误0、exit0**；基线5977/628/0/7增加62pass事件/10顶层及1明确外部环境skip |
| Go Build任务及主代理fresh go build -o bin/server.exe ./cmd/server；go vet ./... | **AUDIT_MAIN_BUILD_EXIT=0 / AUDIT_MAIN_VET_EXIT=0**，零error输出，AUDIT_MAIN_VALIDATION_END完整返回 |
| 本地fresh Linux amd64 server及service测试二进制 | **两编译exit0**，仅bin产物，无上传或部署；完整SHA见下方 |
| 两Go编辑器诊断/scoped diffcheck | **0错误/exit0**；生产规范化gofmt一致，测试格式残余如顶部，不宣称rawgofmt全绿 |

新增回归具体：完整空安装/AllLegacy空/capture、public identity capture；paper/exam/candidate/tester/question/PDF和恶意paper字符串绑定；同exam历史兄弟精度；孤儿revision/run/复合错配在empty/all/mismatch均拒绝；query/row/scan失败；type/NULL/collation/indexprefix/FK target/action/order；audit-only/缺revision-current-receipt/未知report-only/缺额外重复注入列/非requested table；1001 owner分500+500+1 typed绑定；完整gate四查询错误和scoped审计错误。

**全量8个skip完整保留**：TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun；TestPhase1WordTemplateCandidateUploadContract；TestBugFB170_Phase1GroupPieLabelsStayOutsideChart；TestPhase1CustomerWordTemplateLibreOfficeProducesExpectedPages；TestBugFB169_Phase1ChartGenerationPreservesTemplateLabelStyles/uploaded_staging_template；TestBugManagementTraitsPDFVisibleTestLabel；TestManagementTraitsSourceLockMySQLExternalUpdate；新增TestManagementTraitsGuardAuditMySQLExternal。原7项环境条件未改，新项因MNG_GUARD_AUDIT_MYSQL_DSN缺失明确skip，不以既有上阶段source-lock实证把本轮skip改成PASS。

本轮fresh本地Linux产物（GOOS=linux/GOARCH=amd64/CGO_ENABLED=0，不是已运行服务器版本）：
- [bin/server-guard-audit-linux](../Go-based%20Refactored%20System/bin/server-guard-audit-linux)：49828320 bytes，SHA256 **0ee9b326c459611babb6951c63291abcdfc27491e8782d3dd9178e4bb56d8483**，go build退出0。
- [bin/service-guard-audit-linux.test](../Go-based%20Refactored%20System/bin/service-guard-audit-linux.test)：18524853 bytes，SHA256 **d1b1c3f7b6e5dc4d9c33644460ef1a0fc99c80435aefeba276db5bdcf2fe69cd**，go test -c退出0。
- 两源码SHA256分别为guard **5b83a097bad5140b6d494f337212a4cc7414008d432c34ab2593c562c184194c**，新回归 **df29795f9d0e4f943d1521ea507f89c7310629d31860954a401ceb866b88ae83**；后续源码如有变化必须重建并重新记录SHA，不复用本条为fresh证明。只读file_search因ignored bin未找到，list_dir已确认两个产物存在。

### fresh恢复库复验指引（仅交接，当前不运行）

1. 当前源重新构建Linux amd64 server及service测试二进制，记录完整SHA，禁止复用上一actualalias已删除二进制或旧server SHA；本轮本地fresh产物和SHA见下方最终receipt。后续上传仅独立验证器payload，不替换main服务。
2. 核实staging主机/REPORT_EFFECTIVE_ENV及既有备份适用性，维护窗口内只创建**唯一**mng_guard_audit_test_<16小写hex>恢复库。恢复旧表，不含USE/CREATE DATABASE重定向，cross-schema FK必须0；运行未改001 first/repeat，比较11表15FK及完整签名。主element仍不写、不部署。
3. 外部测试要求MNG_GUARD_AUDIT_MYSQL_SCHEMA精确等于上述库名，MNG_GUARD_AUDIT_MYSQL_DSN数据库名相同，不允许multiStatements/interpolateParams；真实密钥/DSN由既有安全渠道在子进程内存消费，禁止打印、落文件或进git。用既有socket/root受控复验不是应用账号授权验证。
4. 源码环境从Go根运行：`go test ./internal/service -run '^TestManagementTraitsGuardAuditMySQLExternal$' -count=1 -v -timeout 120s`。fresh测试二进制运行：`service-guard-audit-linux.test -test.run '^TestManagementTraitsGuardAuditMySQLExternal$' -test.v -test.timeout 120s`。必须非skip、exit0，不能将默认skip当真实验证。
5. **二进制cwd必需**：测试读取../../../scripts/sql/management_traits_001_runtime.sql。独立payload保留层次：cwd为payload/Go-based Refactored System/internal/service；未改SQL放payload/scripts/sql/management_traits_001_runtime.sql，二进制放payload/bin；从该cwd用payload/bin的绝对路径执行。核SQL完整SHA7ff621…后再跑，不能修改夹具去paper_id以隐藏问题。
6. 外部用例先真实完整checkRuntimeSchema与empty/all-empty，再选择恢复库既有candidate/paper/exam关系；**仅在可回滚事务**插入synthetic bundle/snapshot/run/revision/audit，实际paper/exam/candidate/AllLegacy和public identity闭包均protected且无error。ROLLBACK后11sidecar行数全0；不写旧表、PDF/会话，不CREATE/ALTER/DROP，不插formal。测试fixture只证明真实关联/SQL，不等于140题/评分/报告内容完整E2E。
7. 另由fresh原actual validator复验完整11表＋owned capture的public identity和AllLegacy断言（两者必须执行），记录actual列标签/行数及所有gate结果，确认上一ACTUAL_IDENTITY_CAPTURE_BLOCKED解除。此独立external用例不创建capture表，也不禁用FK制造真实孤儿；孤儿/漂移当前证据仍是本地真实SQL/GORM夹具。
8. 任一步失败立即停止main迁移/部署。所有临时对象仅按确切ownership receipt清理；ROLLBACK/11表0核验后精确DROP本次恢复库与已列payload，正式备份永久保留；旧12表摘要、465PDF、server/index/config/template/health前后相同再记录结论。本轮没有执行这些远端步骤。

### 残余已知风险

- canonical报告安装后每次旧scope检查增加原完整gate四metadata查询和一次全局audit anti-join，之后按scope有界子查询；没有N+1逐audit查询，但大审计表EXPLAIN/延迟未测，不称性能验收。
- 原freeze/mutation gate仅单进程，外部写或多实例仍有check/write竞争；本bug不扩张到分布式协调，不以unit/审阅声称真实并发安全。
- 生产一轮/新测试三轮后停止；不继续格式修整或拆文件绕过上限。本轮未编辑其他业务文件或原DDL，五个既有docs仅追加事实。当前仅本地bug GREEN，真实恢复库及完整staging报告链仍待fresh验证。

[纠正 - 2026-10-03] 下方“需再授权逻辑修复”是上阶段停止边界；用户本次已明确授权此必需修复，本轮已完成限定本地实现，不再索要重复授权。上阶段actual失败记录保留，尚不能把本地GREEN写成actual guard PASS或可部署。

## 新别名代码受控actual复验（2026-10-03T12:10:16Z，整体BLOCKED）

**阻断先列明**：首次及重复DDL后的生产完整`CheckRuntimeSchema`实际PASS，但后续`ManagementTraitsIdentityScope`捕获保护链实际返回拒绝，演练退出1。本阶段不能称整体第一阶段PASS、不能执行主element DDL或部署；独立legacy guard的最终protected=true断言因前一断言失败未执行。没有修改生产逻辑/原测试/001 SQL，没有删除或放宽失败门禁。两题本真实来源冻结、实际密钥token roundtrip与独立双连接来源锁已通过，不扩大为四组合HTTP/UI/PDF或完整运行验收。

### 备份适用性与fresh实际装配

- 只读SSH strict已知密钥/BatchMode/ConnectTimeout10/ConnectionAttempts1成功，hostname/user=vm-ubuntu-go-dev/liming；MySQL8.0.46，APP_ENV=production、REPORT_EFFECTIVE_ENV=staging，三服务active/healthok。
- 沿用受限正式备份 `/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9`；目录root0700，数据库SHA **9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1**，三件SHA256SUMS、gzip及两tar遍历通过。完整application归档与当前server/dist/configs/两套模板`tar --compare`通过，无内容或归档元数据漂移；12旧表摘要及465份PDF清单与备份相等。无新备份、无共享chmod，原备份永久保留。
- fresh临时Go验证器从当前源码构建Linux amd64，SHA **5cdfd07eeadf0fee60fab43f964d6cd441550907f09a96833b722a80f4e0674d**；原service测试二进制SHA **6976307da8b0c206c66f41f2a45a12c46ee72da787749dc95889be133a77a0af**，两build exit0、验证器vet exit0/diagnostics0。临时源码仅通过apply_patch建立，位于B区ignored tmp且已最终删除；未更改任何业务源/原测试。
- 验证器内部读取当前服务进程环境，调用实际`config.Load()`消费基础及production覆盖配置和JWT密钥；DSN先核原element/local3306，再仅转为**exact owned恢复库、root既有unix socket认证**，不创建用户/grants/cnf/env文件、不输出DSN/密钥/PII/token。`SELECT DATABASE()`精确断言临时根名；使用真实MySQL pool/GORM及生产RuntimeService，不mock、不以diagnostic SELECT替gate。该root隔离复验**不是应用账号授权验证或已运行服务重启验收**。

### 首次/重复DDL与生产完整门禁实证

- 唯一恢复库 `mng_verify_20261003_4407cd0a1cb24799`；同备份恢复成功，所有恢复旧/新FK的`referenced_table_schema<>DATABASE()`计数0，**没有旧FK仍指向element**。restore没有CREATE/DROP DATABASE或USE重定向；不改旧列。
- 未改001 SQL SHA保持 **7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8**；首次及重复均11表/15个RESTRICT-RESTRICT FK，完整列/NULL/charset/collation/索引/FK签名cmp相等，SHA **1020f16a37f84f9fbecdfd8c425969def679af26ed2eb4fb0fdea4a1b460d4e6**；额外FK逐列/列序签名SHA **ef9c1c7b409e4aa413e70de853d27b9818e09b6e339de8b4ffe1a6a178c2318a**，first/repeat相等。
- 每个fresh Linux实例原样执行生产`CheckRuntimeSchema`；真实SQL pool观察实际执行查询的`Rows.Columns()`，真实GORM logger只记录行数/error，不重复诊断查询、不截取或变更校验输入。首次、重复及post三个fresh实例均完整PASS；各同实例第二次调用PASS且metadata查询仍恰4次，缓存生效。

| 生产实际查询位置（完整SQL见原文件） | 实际驱动列标签 | actual GORM rows / error |
|---|---|---|
| [table查询](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L403) | table_name,engine | **11 / false** |
| [column查询](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L422) | table_name,column_name,column_type,is_nullable,character_set_name,collation_name | **159 / false** |
| [index查询](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L425) | table_name,index_name,non_unique,seq_in_index,column_name,prefix_length | **67 / false** |
| [FK查询](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L428) | table_name,constraint_name,column_name,ordinal_position,referenced_table_name,referenced_column_name,update_rule,delete_rule | **21 / false**，复合FK含21列行、15约束 |

首查询精确为`SELECT table_name AS table_name, engine AS engine FROM information_schema.tables WHERE table_schema = DATABASE() AND LEFT(table_name, 7) = 'el_mng_' ORDER BY table_name`。实际11rows不再被大小写扫描丢成有效0；生产validator完整返回nil，严格全部字段/索引/FK/旧引用capacity保持，不是单表count PASS。

### Gate后实际来源/token/capture：新阻断，不继续试改

- 仅在owned恢复库新建两marked临时exam/link，调用真实生产`FreezeProfile`，内部源loader/私有容量guard原SQL实际各700rows/errorfalse，再解码冻结manifest/mapping为140题/700选项；随后按确切主键删除profile/bundle/link/exam。源题/两共享repo没有更新。
- 00201 manifest SHA **66277c8bd4f2d51fc2ab51020625e6fe27b8df10ae02381bdb023b5773e97e0d**、mapping SHA **dc7faac623b10bc42c4e39b493518469ef0c14e586b35e6d0f7ddaafbd1e58bf**；00202 manifest SHA **4a033e1bf8f9a39dc66b92ef36ecd38bfd0ed9f598a4c295be1b32cd72ee5e2d**、mapping SHA **f0f7387dbc9ab9ff95205262baab1b2185b56288e0bcc9f866ac1c114e0889b6**。00202保持db-current原文，不改V67/V96或声称等于客户修订稿。
- 实际配置密钥签发/解析participant-purpose token PASS；token只在进程内存，未输出/写盘、未发业务HTTP、不创建会话。
- full gate PASS后，owned库临时新增且明确defer删除`el_mng_owned_capture`，只含synthetic participant/paper sentinel，不含人员资料。实际identity capture投影标签table_name/rows1；同public identity链调用legacy table投影标签table_name/rows12，legacy extra-column投影标签table_name,column_name/rows41；均errorfalse。七projection共26列标签均实际执行观察过。
- **实际public身份保护断言失败：VALIDATOR_BLOCKED_STAGE=ACTUAL_IDENTITY_CAPTURE_BLOCKED**。后续独立AllLegacy protected断言未执行，不伪称完整guard成功。生产源码[capabilities检查](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L327-L340)只识别paper/exam/profile/participant/candidate/tester/paper_question/pdf_path，或有core时run_id/result_run_id；实际DDL签名中的`el_mng_report_audit`只有id/report_id/actor_id/action/created_at，没有任何认可列。完整11表安装后它成为未知extra，检查在查询保护数据前失败关闭；这与metadata alias已填充不同，是**新增完整安装兼容阻断**。本阶段只复核，不补report_id跳过、不删audit表、不降为七表、不改capture fixture遮蔽问题。
- 原alias回归的capture测试只模拟capture元数据，不覆盖完整11表中的report_audit额外保护列；5977/0/7本地GREEN不能覆盖本次actual失败。下一backend worker先完整11表＋capture/空数据/审计关联的RED，再最小安全处理审计→revision→run/paper保护闭包，之后fresh复验；需另一次用户明确逻辑修复授权，本阶段停止。

### 独立原双连接测试与清理

- 完整Schema已PASS后，在另一个唯一空库 `mng_source_lock_test_20261003_e680e18840cd49ef`独立补跑**未修改**原`TestManagementTraitsSourceLockMySQLExternalUpdate`；2.31s、**4pass事件/0fail/0skip、exit0**，commit/rollback/revocation-after-MVCC-snapshot三子项通过，第二连接真实UPDATE收到1205、释放后成功，locking-current-read拒绝旧MVCC审核。
- 既有socket认证的root临时localhost随机TCP relay仅用于满足原测试localhost边界；DSN只注入子进程环境，进程退出后不残留，未建用户/grants/env文件。原test defer删除唯一fixture表，查询table remaining0后由EXIT DROP exact owned库。
- 主演练evidence `/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/actualalias_4407cd0a1cb24799`：原日志/first-repeat签名/post失败/cleanup均保留root私有目录，receipt **EXIT1 / TEMP_SCHEMA_REMAINING0 / CLEANUP_ERRORS0 / MAIN_MNG0 / PAYLOAD_REMAINING0 / PDF_COUNT465**。
- 独立锁证据同backup下 `actualalias_lock_e680e18840cd49ef`：receipt **EXIT0 / TEMP_SCHEMA_REMAINING0 / CLEANUP_ERRORS0 / PAYLOAD_REMAINING0**。未创建的主流程lock库 `mng_source_lock_test_20261003_4407cd0a1cb24799`也精确确认不存在，不能称其测试已执行。
- 两确切上传目录 `/tmp/mng_actualalias_20261003_4407cd0a1cb24799`、`/tmp/mng_actualalias_20261003_e680e18840cd49ef`均只删除本次列明文件后rmdir，无glob/rm-rf。本地唯一validator源码/两二进制/空目录亦最终精确清理；没有local/remote凭据或token文件、没有uploads/业务会话。本轮不清理其他任务产物或正式备份。
- 最终只读SSH **ACTUAL_ALIAS_FINAL_READONLY_EXIT0**；旧12表摘要保持 **f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e**，table/exam/paper/paper_qu/paper_qu_answer/candidate/tester计数 **67/70/1487/134354/294628/1348/27**；主mng0、465PDF清单逐文件相等、application/config/templates归档无漂移；server SHA ee4566e7a698ff592acaeab40c5986974787be27929cd1c2bc002d1adddcd300，index SHA 2d4ba6c5ac1e745d736f8c2894cb99406d4890123bb18fb0c0d8bb001c8e9bc9；三服务active/healthok。
- 本轮不重跑Go全量5977/0/7或前端355；fresh Linux validator/test build及vet是本轮执行。独立锁4/0/0不把默认七skip改成0。无主库DDL/数据写、服务重启、配置变更、客户原件/旧PDF变化或production部署。

### 下一deploy安全顺序及现场路径事实（计划，不执行）

1. **先解除本次audit-only guard阻断并fresh完整复验**，否则止于此。不是仅fullSchema PASS即可部署；不要用本次已删临时二进制作为服务产物。
2. 再核当前现场指纹/备份适用性；从最终fresh源码重建Linux server和前端完整包、记录完整SHA，本地早先0aa1/e7f8等stale产物不可复用。现场仍旧server ee4566e7…、旧index 2d4ba6c5…，并非新版已发布。
3. 实际unit `/etc/systemd/system/talent-assessment.service`，User=liming、WorkingDirectory=/opt/talent-assessment；配置文件实际为 `/opt/talent-assessment/configs/application.yml`＋`application-production.yml`，APP_ENV=production选择配置但REPORT_EFFECTIVE_ENV=staging限定报告批准。旧代码仍在运行，需先安全guard/排空旧生成上传压缩与停写，再staging主DDL first/repeat及完整gate，然后TEST配置/后端/前端，重启刷新缓存；本阶段不自动执行这些步骤。
4. 实际exportTemplates=/opt/talent-assessment/configs/export-templates，liming:liming0750；upload.path=/opt/talent-assessment/tmp/uploadPath，0755。Nginx仅观察到root /opt/talent-assessment/dist和alias /data/uploadPath/profile/；两个upload根均0755，**不能把它们或tmp作为新增私有报告根**，不chmod共享目录。
5. 实际MNG_TEST_REPORT_ENV/DIR/CONTENT_PATH/TEMPLATE_PATH四key在运行进程全部未设置；两个新TEST目标文件及新private目录实际不存在。后续部署才准备独立绝对目标 `/opt/talent-assessment/configs/export-templates/management-traits-002-test-only-v2.docx`（SHA05c55e77…）、`/opt/talent-assessment/configs/export-templates/management-traits-002-test-content-v1.xlsx`（SHAb0498249…）、新增不被静态alias覆盖的 `/opt/talent-assessment/private/management-traits-test-reports`（liming专有0700/文件0600，**仅建议目标，尚未创建**）。显式MNG_TEST_REPORT_ENV=staging，不凭APP_ENV猜测；精确新文件路径注入，不覆盖旧模板。
6. guard/drain→staging主DDL→fresh TEST后端/模板原文/前端→权限与重启→四真实组合及expiry/revoke/resume/Worker/PDF全文六图SQL→marked IDs/私有新文件/会话finally清理→旧摘要/465PDF/health终验。当前overall BLOCKED；不因用户既有staging部署授权而自行越过本阶段“只验证”范围。

- 本地清理执行补充：两次apply_patch Delete均未实际删除临时源码，两个保护性检查各exit1停止、未扩展清理范围；最终按用户exact temporary source清理授权，先核唯一目录仅main.go/verifier/service.test及两二进制完整SHA、已读取源码所有权，再精确删除这三个临时产物和空目录，**ACTUAL_ALIAS_LOCAL_TEMP_REMAINING=0 / LOCAL_DEDICATED_DSN_PRESENT=False**。没有用终端改写源码内容。最终两文档diffcheck0/diagnostics0；失败工具检查不掩盖、不记为最初已清理成功。

[纠正 - 2026-10-03] 下方“actual完整Schema仍FAIL/尚未执行列索引FK”现由本轮fresh真实PASS限定解除，历史记录保留；**新完整11表guard/capture阻断未解除**，所以不得将该纠正写成staging全部PASS或可部署。

## MT-SCHEMA-ALIASES：元数据扫描兼容修复（2026-10-03，仅本地）

**未验证先列明**：本轮不SSH、不真实DB读写/DDL/部署；actual完整Schema仍待fresh演练，不能称staging验收。前端/PDF/race/覆盖率未跑，7环境skip不算PASS；不处理旧EOF/CRLF，不动A区、共享旧表/源题、客户文件/私有PDF/正式备份。开工后报告和memory新增11:43:59Z只读交接已重新读取并原样保留，历史actualFAIL不删除。

### 根因与最小变更

- 上阶段actual TABLE_NAME/ENGINE大写，11行与小写GORM tags不匹配、有效0；独立显式alias诊断有效11。本轮只补SQL小写AS，不把validation改为忽略大小写。其余元数据大写标签属于本地模拟的同类风险，未声称远端逐项观察。
- 3生产文件**7条投影**：完整Schema四查询22输出字段（已有3个表达式alias），identity capture及legacy guard三查询4输出；**新增23个AS，26个输出全部显式映射**。WHERE/JOIN/参数/ORDER BY/批次、签名、cache/失败粘滞/fresh刷新/atomic capacity发布均保持。

| 生产位置（精确替换行） | 输出别名（均匹配原tag） |
|---|---|
| [tables](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L403) | table_name、engine |
| [columns](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L422) | table_name、column_name、column_type、is_nullable、character_set_name、collation_name |
| [statistics](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L425) | table_name、index_name、non_unique、seq_in_index、column_name、prefix_length |
| [FK](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L428) | table_name、constraint_name、column_name、ordinal_position、referenced_table_name、referenced_column_name、update_rule、delete_rule |
| [identity capture](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L72) | table_name |
| [legacy table](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L274)、[legacy columns](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L313) | table_name；table_name、column_name |

### 消费方闭环（无签名变更）

- **同步修改**：checkRuntimeSchema七投影所在3文件；14既有测试的65个SQL-prefix/DI计数prefix。数据、断言、错误/skip合同不改；worker baseline字节核验14/14仅授权替换、85测试文件其余字节不变。
- **无需修改（同gate/返回/事务/校验合同不变）**：Setup；FreezeProfile；CreatePaper/PaperDetail/ProfileDetail；FillAnswer/Submit/SubmitParticipant；LoadValidatedRun；ListRuns/ResultDetail；TryRegisterCandidateIdentity/TryTesterIdentity；IssueCandidateResume；GenerateTestReport/ReadTestReportPDF；ScanExpiry/scanExpiry。它们直接消费修正后的扫描，权限/时限/来源/TEST边界不改。
- **无需修改（仅SQL输出名称，不变结构/API/文件）**：001 DDL、模型、旧001/003/00401、前端/模板/客户源/PDF/配置。scalar COUNT(*)按位置扫描不受struct标签影响，保持。
- 既有测试同步精确清单：service [create86](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_create_test.go#L86)、[guard28起10处](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard_test.go#L28)、[identity21起8处](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity_test.go#L21)、[resume19/32～34共4处](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_resume_test.go#L19-L34)、[schema behavior24起11处](../Go-based%20Refactored%20System/internal/service/management_traits_schema_behavior_test.go#L24)、[schema17起5处](../Go-based%20Refactored%20System/internal/service/management_traits_schema_test.go#L17)；handler [DI36/293/399](../Go-based%20Refactored%20System/internal/handler/management_traits_di_test.go#L36)、[resume209～212/579/581](../Go-based%20Refactored%20System/internal/handler/management_traits_handler_resume_test.go#L209-L212)、[identity chain26](../Go-based%20Refactored%20System/internal/handler/management_traits_identity_chain_test.go#L26)、[guard33](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_guard_test.go#L33)、[identity53起9处](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_identity_test.go#L53)；router [DI26/226/233](../Go-based%20Refactored%20System/internal/router/management_traits_di_test.go#L26)、[routes27/43](../Go-based%20Refactored%20System/internal/router/management_traits_routes_test.go#L27)、[retired setup64](../Go-based%20Refactored%20System/internal/router/retired_modules_test.go#L64)。行号未变，无其他断言/格式/EOF修改。

### RED→GREEN及严格证据

- [新测试](../Go-based%20Refactored%20System/internal/service/management_traits_schema_alias_test.go#L18)根据**执行SQL**生成驱动标签：无AS返回大写原列，有AS才返回实际alias。独立原完整SQL作oracle，去alias后整句必须相等；sqlmock1.5.2 Rows持有标签slice，真实GORM Scan消费，不从tags推fixture。
- [驱动对照](../Go-based%20Refactored%20System/internal/service/management_traits_schema_alias_test.go#L112)：无AS真实扫描1行/两字段零值，有AS正确填充；[完整gate](../Go-based%20Refactored%20System/internal/service/management_traits_schema_alias_test.go#L142)两正例（含旧列精确兼容）、六反例（type/NULL/collation/indexprefix/FK action/order）、两调用只查一次、失败不发布capacity；[capture](../Go-based%20Refactored%20System/internal/service/management_traits_schema_alias_test.go#L196)三投影读取/保护命中。
- 修改生产前有效RED：**3pass/12fail/0skip、exit1**；合法gate被拒及capture TABLE_NAME扫描失败。首次GREEN4pass/11fail因新COALESCE解析顺序/guard查询形状夹具错误，按实际源码纠正，不改生产validation。最终新回归 **15pass/3顶层/0fail/0skip、解析错误0、exit0**。
- 独立CodeReviewer **PASS、无finding**；七投影/四metadata查询真实GORM测试、严格规则和MySQL5.7/8语法通过静态复审。没有再运行测试或称staging通过。
- 四包聚焦go test ./internal/service ./internal/handler ./internal/model ./internal/router -run 'ManagementTraits|BugMTSchema' -count=1 -json -timeout180s：**5265pass/258顶层/0fail/2skip、解析错误0、exit0**，较基线5250新增15通过事件/3顶层；skip为PDFVisibleTestLabel及SourceLockMySQLExternalUpdate。
- 最终隔离终端原始UTF-8 JSON：go test ./... -count=1 -json -timeout180s，**5977pass/628通过顶层/0fail/7skip、解析错误0、exit0**；较旧5962/625新增15事件/3顶层。go build -o bin/server.exe ./cmd/server及go vet ./...均**exit0、零error输出**。前一次PowerShell聚合终端当时未返回最终统计，不用其不完整buffer作证；Node spawnSync独立完整重跑得到ALIAS_ISOLATED_VALIDATION_END。计数含顶层/子项，不是覆盖率。
- 七环境skip原样：TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun、TestPhase1WordTemplateCandidateUploadContract、TestBugFB170_Phase1GroupPieLabelsStayOutsideChart、TestPhase1CustomerWordTemplateLibreOfficeProducesExpectedPages、TestBugFB169_Phase1ChartGenerationPreservesTemplateLabelStyles/uploaded_staging_template、TestBugManagementTraitsPDFVisibleTestLabel、TestManagementTraitsSourceLockMySQLExternalUpdate。本轮未配置真实DB测试DSN，不将既有staging sourceLock PASS折算为默认全量0skip。
- 编辑器测试发现No tests found，原生Go为证据；Go Build任务已完成。三生产文件规范化内容与gofmt相等，新测试差异仅EOF末换行缺失（只读确认onlyFinalNewline=true），不宣称raw gofmt全绿。生产一轮修复、新测试三次编辑后停止，不处理旧EOF/继续格式试改；scoped diffcheck0、18Go及5文档diagnostics0，001 SQL SHA保持7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8。
- [纠正 - 2026-10-03] 下方11:43:59Z“当前源码仍无显式别名”是修复前事实，本轮已仅在本地补齐并全量验证；远端fresh完整gate未复验，原实际Schema FAIL/主库mng0/465PDF保护及备份receipt不由本地GREEN覆盖。

### 下一actual rehearsal validator目标（本轮未执行）

主代理fresh源码重建独立Linux validator并记录完整SHA；唯一owned恢复库001 first/repeat后运行**生产CheckRuntimeSchema**，不能用diagnostic查询代替gate。必须完整11表/全部type-NULL-charset-collation/PK-unique-5读取索引/15FK列序与RESTRICT/精确旧引用capacity首次与重复均PASS；记录actual Rows.Columns的26投影标签、gate返回、DDL签名前后、旧表/465PDF摘要及exact cleanup。fresh实例刷新、同实例失败粘滞不绕过。未PASS不得mainDDL/部署；旧备份/sourceLock两次各4pass记录保持，本轮默认7skip仍不是环境验收。

## 第一阶段交接只读复核（2026-10-03T11:43:59Z，仍BLOCKED）

**未解除：实际Schema门禁失败，不能推进主DDL或部署。** 当前源码metadata查询仍未加显式别名；三次演练上限已到，本次不重新建库或重跑失败演练，不改生产代码/测试/DDL。

- 实际SSH退出0，hostname/user=vm-ubuntu-go-dev/liming；strict known key、BatchMode、ConnectTimeout10、ConnectionAttempts1。实际APP_ENV=production、REPORT_EFFECTIVE_ENV=staging，三服务active、内部health返回ok。
- 既有112853_fd24d4ee35a9受限备份root0700/所有文件0600，三件SHA全部OK、gzip及两个tar遍历通过；正式备份保留，未新增备份或改变共享权限。
- 两后续演练first/repeat完整签名cmp通过；各自legacy-before/after及PDF-before/after清单cmp通过。cleanup receipt均EXIT=1 REMAINING=0 ERRORS=0，exit1仍是Schema失败，不是清理失败。
- 五个确切receipt对应临时schema实时remaining0；三个确切payload路径不存在；主element mng表0。READONLY_RECHECK_EXIT=0，仅代表本次只读复核通过，不代表actual Schema或完整002测试链PASS。
- 本次无主库写、DDL、部署、环境/服务/dist替换或production操作，未重跑Go/build/前端/PDF测试。后续backend worker先真实大小写列标签RED→最小投影别名修复及全量Go/build/vet，再独立fresh Schema演练；通过后才进入既有guard/drain与staging部署清单。

## SSH恢复后第一阶段实际演练 receipt（2026-10-03，BLOCKED：真实Schema扫描拒绝）

**未完成先列明**：第一阶段不是PASS，不能进入主element迁移或部署。最新代码实际CheckRuntimeSchema在完整临时库首次/重复DDL后均失败；原因已由真实MySQL驱动与GORM扫描定位，不是SSH、DDL安装或旧列兼容预算。没有修改生产Schema代码/原测试/001 SQL，没有放宽任何门禁；正式新链HTTP/UI/PDF和重启Worker仍未执行。

### 实际主机、备份与安装证据

- SSH恢复：既有liming/known key，StrictHostKeyChecking=yes、BatchMode=yes、ConnectTimeout10、ConnectionAttempts1；hostname=vm-ubuntu-go-dev，三服务active，MySQL8.0.46；APP_ENV=production但REPORT_EFFECTIVE_ENV=staging。内部和公网health均HTTP200/ok。主element新增mng表0。
- 新备份=/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9，root:root0700，所有文件0600/权限错误0；没有chmod共享旧目录。全库single-transaction/quick/routines/triggers/events/hex-blob，gzip测试通过，无CREATE/DROP DATABASE或USE重定向；数据库12793472 bytes，SHA256=9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1。
- application.tar.gz SHA256=079a718a96d1fd43d0fcf037d3796362684d2211400d975258f0d929c90d5fbd，包含旧server/dist/configs/两套MBTI模板；files-and-system.tar.gz SHA256=21fa940fd5800f70e33c047c42e13148ddeeebe8a1a2240c466c765a81ff998d，完整归档tmp、/data/uploadPath、Nginx和unit。两归档gzip及tar遍历验证通过，三件SHA256SUMS复验全部OK。465份PDF逐文件SHA清单只保存在root私有备份内，未回显人员/全路径；归档前后及演练后PDF清单完全一致。备份永久保留。
- 首次独立库mng_verify_20261003_dce3d6ae80b64301恢复成功、DDL首次成功，11表15RESTRICT FK；actual gate失败后EXIT清理remaining0。首次失败日志保留在备份根目录，不覆盖。
- 第二/第三次有界独立演练分别在rehearsal_baccb22d7fd9478f、rehearsal_b0dcc6a8e1f34f38证据子目录；每次从同一全库备份恢复到全新唯一库，first/repeat均11表15RESTRICT FK，完整列/NULL/collation/索引/FK签名两次完全一致，SHA256=1020f16a37f84f9fbecdfd8c425969def679af26ed2eb4fb0fdea4a1b460d4e6。DDL原文件SHA256=7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8；主element从未applyDDL。

### 已证实的新阻断（不可用sqlmock GREEN覆盖）

- [实际门禁查询](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L403)写SELECT table_name, engine，但MySQL8真实驱动Columns返回大写TABLE_NAME/ENGINE；[目标结构](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L49-L52)tag是小写table_name/engine。实际查询rows11、GORM populated_innodb0；仅在独立诊断SELECT加显式小写AS后rows11/populated_innodb11。生产门禁仍原样执行并返回management traits data rejected。
- 运行日志显示拒绝发生于首个table metadata扫描后的检查，尚未运行后续完整列/索引/FK校验。因此不能把11/15或人工签名一致写成actual完整Schema PASS；其他projection是否同样受大小写影响尚未由生产完整门禁确认。
- 本轮没有修改/自批准放宽原Schema规则，没有把诊断别名查询替代生产CheckRuntimeSchema。达到三次真实演练后停止；需后续backend worker先RED复现真实大写列标签，再最小别名修复、核全部消费方/原fixture，全量Go/build/vet及fresh临时库门禁复验。不要ALTER旧表或用忽略case/skip/tablecount来绕过。

### 真实双连接、旧数据和精确清理

- 原TestManagementTraitsSourceLockMySQLExternalUpdate源文件未改，fresh Linux service测试二进制SHA256=b38bbe377dc62e782462abb8d07b09f64b4696b2c3c9378f70facc08ca923321。在两个独立空mng_source_lock_test_20261003_*库各跑一次，commit/rollback/revocation-after-MVCC-snapshot三子项全部PASS（4pass事件/0fail/0skip，2.61s/2.60s）；真实第二连接普通UPDATE在shared lease期间收到1205，释放后成功，locking current read拒绝已撤销来源。没有创建用户、修改grants或MySQL配置；root进程临时localhost随机端口relay到unix socket，只使原TCP限定测试消费socket认证，测试门禁原样。
- 最新验证器从fresh源码交叉编译，build退出0；最终诊断二进制SHA256=ded0c1f9030ce6b2a54256304bbaf4d4b9fccc0f3de7608572600e7f9e5c3eca。临时cmd源码在完成后删除；业务代码/公开API/原测试/配置/前端没有改动。新无凭据操作脚本为[scripts/db/management-traits-staging-phase1.sh](../scripts/db/management-traits-staging-phase1.sh#L1)，bash -n实际exit0；它是第一阶段备份/独立演练执行器，不是正式部署脚本。
- 12旧表有序全行/结构dump指纹前后完全一致：f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e。表数/exam/paper/paper_qu/paper_qu_answer/candidate/tester前后=67/70/1487/134354/294628/1348/27。旧server SHA256保持ee4566e7a698ff592acaeab40c5986974787be27929cd1c2bc002d1adddcd300；旧dist/index保持2d4ba6c5ac1e745d736f8c2894cb99406d4890123bb18fb0c0d8bb001c8e9bc9。
- 两repo实际只读JOIN：00201/00202各140关系/140题/700选项/700唯一选项，sort1..140；question ID最大字节4/5，option12/14，非ASCII计数0、Vcode错配0；旧exam-repo JOIN7/6，无1267。旧repo MB3→MB4及两个64→32签名只读再次确认，不改源题。
- 五个确切已创建schema（首次恢复1+两后续恢复2+两锁fixture2）均由各次EXIT trap仅按owned receipt DROP，最终逐个查remaining0；fixture自己的表也已先由原test defer删除。唯一上传目录/tmp/mng_phase1_20261003_091ed97fcf06内verifier/service.test/runtime.sql精确删除并rmdir，remote payload remaining0；诊断中间上传名亦不存在。正式备份/旧PDF/共享源题不删。
- 两次完整后续演练exit1是实际Schema FAIL，不是整体GREEN；其中cleanup errors0、remaining0。后续只读终验脚本曾有括号错误（1064）、PowerShell末行CRLF产生exit127，数据查询及清理结果已执行；须用最终无CRLF只读复核退出码作为终验receipt，不能隐藏这些工具执行失败。
- [终验补充 - 2026-10-03T11:40:41Z] 初次去CRLF管道因native参数引号剥离使tr表达式错误，远端syntax exit2，没有写操作；临时源码首次删除未实际落盘，局部清理按“不空则停”拒绝误删。最终用Node spawnSync直接stdin发送LF脚本，PHASE1_FINAL_READONLY_CLEAN_EXIT=0：五确切库remaining0、remote payload0、main mng0、三服务active、内部healthok、三件备份SHA全部OK。仅本任务四个本地二进制及已确认临时源码/空目录按用户exactcleanup授权最终删除，LOCAL_TEMP_BINARY_REMAINING=0、LOCAL_TEMP_SOURCE_REMAINING=False。doc/script diffcheck0，编辑器diagnostics0。所有先前失败日志/正式备份保留，actual Schema FAIL未被此次清理exit0覆盖。

下一部署清单：先修复并真实复验actual Schema扫描→第一阶段重新PASS→fresh Linux/server/front完整SHA→现有旧写guard/在途生成上传压缩排空及停写→staging主DDL first/repeat→TEST范围配置/后端/前端部署→四真实组合及expiry/revoke/concurrency/resume/Worker/PDF六图全文SQL→marked确切IDs/文件/会话finally清理与传统摘要终验。production不在授权范围，本agent没有进入主库或部署步骤。

## 已授权staging第一阶段预检（2026-10-03，BLOCKED，未执行数据库演练）

**未完成先列明**：本阶段两次有界SSH预检均在连接建立前超时，退出255，远端命令未启动。未能重新确认主机身份、当前REPORT_EFFECTIVE_ENV、MySQL版本、mng表数量、父列/旧列兼容签名或实际002 JOIN/source。受限全库备份、临时库恢复、DDL首次/重复、fresh Go实际Schema gate及来源锁MySQL测试均未开始；没有可以报告为本阶段PASS的数据库证据。

- 执行边界：用户已授权本阶段仅staging只读预检→新root0700/文件0600完整备份→唯一ownership临时schema恢复→001首次/重复及11表15FK实际门禁→独立source-lock fixture→exact cleanup。主element迁移、旧表ALTER、环境/服务/dist替换、完整result/submit并发和production不在本阶段。本次不创建framework/task或临时库，也不修改共享备份目录权限。
- 已完整读取project-memory、本报告、001 SQL，并读取实际配置加载器、Schema兼容实现、source-lock MySQL测试与相关测试/安全/数据库规则。当前5962pass/0fail/7skip与build/vet GREEN是上一轮本地已验证基线，**不是本阶段重新运行**。
- 两次使用既有liming、已知SSH密钥、StrictHostKeyChecking=yes、BatchMode=yes、ConnectTimeout=10、ConnectionAttempts=1；各返回`ssh: connect to host 20.200.136.133 port 22: Connection timed out`及exit255。一次初试＋一次有界重试后停止，不继续无界重试，不调整网络或host key。
- 独立公网GET /prod-api/health实际返回HTTP200、`{"status":"ok"}`；最终检查UTC=2026-10-03T10:12:59.8198480Z。HTTP健康不代替SSH身份、Schema/签名或报告环境确认。
- 用户指定20.200.136.133为staging。历史有效环境为APP_ENV=production＋REPORT_EFFECTIVE_ENV=staging，本阶段因SSH未建立而**未复核当前值**，不能只凭APP_ENV称production，也不能仅凭历史称当前报告环境已确认。此前SSH restored交接不等于本阶段可连接；本阶段实际网络结果优先。
- 备份路径/SHA、temp first/repeat签名、锁测试pass/fail/skip、legacy before/after摘要及temp remaining0：**未取得**，不得填写推测值。本次未发起远端写操作；未创建临时目标，不能将“未创建”写为“已完成清理”。正式既有备份未删除。
- 后续主代理：先恢复SSH并严格核host/env，再从本阶段只读预检起执行全部已授权步骤；备份验证失败即停。第一阶段真实门禁尚未关闭，不能进入第二阶段。第二阶段仍需fresh Linux/前端重建完整SHA、guard/drain后再主库DDL，不以当前HTTP200或旧本地GREEN作为可部署证据。

## 002旧列Schema guard兼容切片（2026-10-03，本地GREEN）

**未验证先列明**：本轮没有SSH、真实DB读写、DDL、部署、环境或前端/模板改动。7个既有环境skip完整保留；sqlmock执行的SQL不是MySQL验收。当前结论仅为旧列兼容代码/测试完成，不宣称11表15FK已实际安装或完整staging链完成。编辑器未发现新Go测试、工作区构建任务返回未找到，随后使用原生Go执行全部测试及同一构建命令成功；未创建现代化scenario/task文件或使用TaskExecutor。

**格式限制**：最终只读检查确认两个生产Go忽略CRLF后与gofmt一致；新增测试唯一差异是文件末尾缺一个换行（末尾EOF-only），不是缩进或代码差异。该测试已三次编辑，按上限不再改，不宣称新增测试完整gofmt通过。最终代码/文档diffcheck仍exit0，八修改文件diagnostics均0。

### 精确兼容边界及来源

- 用户本轮提供的staging只读实证：MySQL8.0.46，两002源库各140题/700选项，V/raw/反向/唯一性正确；repo/relation/question/answer实际MAX ID长度19/13/5/14且全ASCII；源question/answer id为NOT NULL varchar64 utf8mb4_0900_ai_ci，旧答案桶qu_id/answer_id为NOT NULL varchar32同collation；repo.id为NOT NULL varchar64 utf8mb3_general_ci，两repo_id引用为utf8mb4_0900_ai_ci。旧JOIN计数140/140、7/6且无1267。这些是用户交接证据，**本轮没有远端重复查询**。
- 仅对白名单旧边允许64→32：el_qu.id→el_paper_qu_answer.qu_id、el_qu_answer.id→el_paper_qu_answer.answer_id，必须精确类型/NULL/charset/collation；31/33/35/63等缩窄子列和63→32父列漂移拒绝，不对任意短列fallback。
- 仅对repo.id及其两既有64字符repo_id引用允许精确MB3-general→MB4-0900差异。兼容只接ASCII opaque ID；64 ASCII字符=64字节，不以MB3最大192字节当64字符容量。正常既有opaque-ID UTF8字节合同不扩大，既有小父列8等正常兼容保持。
- 缓存携带字节容量和ASCII限制，窄旧目标预算先约束源父ID，再沿真实引用传播到mapping/options/selected和旧桶。源加载逐实际question/option核验，原SQL/JOIN/LIMIT701及hash/140冻结构建不变；全部组卷源ID在首个INSERT前预检，checked CASE表达式参数在BEGIN前检查。无CAST或IN策略重写。
- 新11表全部字段、索引、15FK/RESTRICT/charset/collation仍严格；paper/pq及其他生成UUID父键36下限保持，原70短paper/pq反例保留并通过。无共享旧表/DDL/env/模板/前端/legacy公开API改动。

### C5消费方闭环（真实grep/read确认）

| 消费方 | 处理 | 依据 |
|---|---|---|
| managementTraitsSchemaReferenceCapacities→validateManagementTraitsSchema/checkRuntimeSchema | 同步修改 | 精确旧边例外；完整元数据先验证，缓存发布仍once/atomic，不任意fallback |
| managementTraitsSchemaByteBudgets→checkRuntimeSchema/managementTraitsSchemaGuardDB | 同步修改 | 内部typed容量带ASCII位；传播到真实所有引用，不改变外部签名 |
| managementTraitsSchemaIDFits→statement/records/source/load | helper同步修改，source/load调用处无需改码 | 已有Raw逐行检查、JSONmapping/options、loaded selected/owner调用保留；同helper得到收紧的目标预算，公开LoadValidatedRun真实正例通过 |
| CheckStatement/CheckRecords→GuardDB Query/Row/Create/Update回调 | 同步修改 | 同pool私有registry不污染legacy；checked clause.Expr参数也核验；读/写/JSON均保留字段和participant限制 |
| NewManagementTraitsRuntimeService→GuardDB(nil) | 无需改码 | nil内部容量及现有atomic指针调用兼容，实际全量DI回归通过 |
| persistManagementTraitsRuntimePaper→CreatePaper及writer测试 | 同步修改 | 所有LegacyQuestions/Buckets源ID先核真实qu_id/answer_id目标，再执行已有批量SQL；API/事务/返回不变 |
| profile freeze/源构建、loadRuntimePaper、FillAnswer/Submit、bucket验证、结果/报告读取 | 无需改码 | 保持已有调用链，只消费更严格缓存预算；聚焦5250与全量5962实际通过，完整公开run读和140/700writer正例通过 |
| 原Schema/容量/缓存/索引/FK测试、001/003/00401、旧模型/DDL/公开API、前端/模板 | 无需改码 | 既有测试与源码保持，原Go全量通过；本轮不声称前端/模板重新运行 |

### 真实RED→GREEN及本地证据

| 阶段 | 结果 |
|---|---|
| 修改前go test ./... -count=1 -json -timeout180s | 5349pass、615通过顶层、0fail、7skip、解析错误0、exit0 |
| 新测试先写，原生RED | 555pass/45fail/0skip、解析错误0、exit1；精确旧metadata正例被拒 |
| 第一轮实现后 | 594pass/6fail/0skip；测试误把正常31父列列为漂移、漏GORM SAVEPOINT，另发现checked CASE参数遗漏；失败如实保留记录 |
| 第二轮Schema联合回归 | 3087pass/0fail/0skip，新增600pass；修正测试父63→32反例/精确3个SAVEPOINT，checked参数纳入门禁 |
| 最终新增测试 | **613pass事件，10通过顶层，0fail/0skip**；增加loaded/selected/JSON/checked Unicode反例及公开完整run只读正例 |
| 最终002四包聚焦ManagementTraits或BugMTSchema | **5250pass、255通过顶层、0fail、2环境skip、解析错误0、exit0** |
| 最终Go全量go test ./... -count=1 -json -timeout180s | **5962pass、625通过顶层、0fail、7既有skip、解析错误0、exit0**；较5349基线增加613pass事件/10顶层，不是覆盖率 |
| go build -o bin/server.exe ./cmd/server | **LEGACY_COMPAT_BUILD_EXIT=0**，无error输出 |
| go vet ./... | **LEGACY_COMPAT_VET_EXIT=0**，无error输出 |
| 三Go增量git diff --check | **LEGACY_COMPAT_DIFF_CHECK_EXIT=0** |
| 独立只读作用域复审 | **PASS，无新增blocking finding**；正常Unicode字节合同是既有约束且明确范围外，不为本compat扩大或改旧API |

新增测试覆盖实际Raw source700行、两code稳定manifest/mapping及140冻结，31/32通过/33拒绝、Unicode/非法UTF8/空ID拒绝；repo64 ASCII通过/65和Unicode拒绝；原DB Unicode查询不受私有callback影响；非法writer只有BEGIN/ROLLBACK且零INSERT，正向实际140/700批量writer；loaded/model/JSON/selected/CASE预算及公开完整run50分只读事务；新表逐列/type/NULL/collation、所有索引前缀/FK动作漂移拒绝及70个UUID短父键拒绝。没有删除或skip任何原失败测试。

7skip仍是：TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun、TestPhase1WordTemplateCandidateUploadContract、TestBugFB170_Phase1GroupPieLabelsStayOutsideChart、TestPhase1CustomerWordTemplateLibreOfficeProducesExpectedPages、TestBugFB169_Phase1ChartGenerationPreservesTemplateLabelStyles/uploaded_staging_template、TestBugManagementTraitsPDFVisibleTestLabel、TestManagementTraitsSourceLockMySQLExternalUpdate。本轮未关闭其环境门禁。

轮次：Schema生产2轮、writer1轮、新测试3次编辑（含初始RED及两次补正/补强），停止继续试改；没有终端写源码。仅两managementtraits生产Go、一新同目录Go测试及既有文档事实追加。下方全引用必须父子charset一致/子容量>=父的历史表述由本条**仅对白名单旧opaque边**限定纠正，其他严格边不变。下一远端环境执行由main另行授权，不在本轮进行。

## 最新汇总 receipt：本地已实现，staging 仍 PARTIAL（2026-10-03）

**未完成先列明：无法完成远端验收。** 本轮 staging 20.200.136.133 使用既有 liming／密钥、BatchMode=yes、ConnectTimeout=10 的三次 SSH 均超时，每次退出255，连接未建立、远端命令未启动；公网 health HTTP200、返回 ok。未启动远端备份、临时Schema演练、DDL、部署或dataset清理；11表／15FK的真实安装及首次／重复执行未验。完整功能已获授权，但目前只完成本地实现与限定验证，不能标记全部DONE或all E2E。production未操作。本次收口仅追加既有五文档，不重跑业务测试、不改代码／格式、不新增rules或报告框架。

### 最新已执行证据（按来源及作用域分开）

| 证据来源 | 最新真实结果 | 不得扩大为 |
|---|---|---|
| 本轮主代理原模板合同 | source-layout 10项／0fail／0error／0skip，exit0；LibreOffice-compatible完整17项／0fail／0error／0skip，68.156秒，exit0；真实SVG、pixel、六图、36长文本及13详情已跑 | 不是目标服务器或真实DB报告验收 |
| 最新独立BuildValidator交接 | Go全量5349pass／0fail／7skip；前端31files／355pass／0fail；Windows／Linux build及vet GREEN | Go计数含顶层／子项，不是覆盖率；7skip不算环境通过 |
| Linux构建交接 | SHA前缀0aa1e764…属于格式only之前的构建 | 不作为当前部署产物；部署前必须从fresh source重建并记录完整SHA |
| 最新本地production bundle＋mock浏览器 | 新续答47cases：47PASS／0fail／0skip；原四组合：560save／4download／8submit GREEN；bundle SHA256=98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51 | production bundle是本地构建类型，不是生产部署；API／PDF响应mock，不是real DB、服务端签名或真实PDF生成 |
| 格式整改最终receipt | scope79文件，其中68文件格式调整；非EOF实质格式差异0，规范化代码／注释未改 | CRLF导致79文件raw差异仍在，score测试另有3处EOF额外空行；三轮上限已停止，不能宣称raw gofmt GREEN，不再格式改文件 |

mock结果已读取现有产物：[续答47项](../scripts/test/results/management-traits-resume-mock.json)、[四组合及bundle](../scripts/test/results/management-traits-frontend-mock.json)。旧失败mock产物及旧失败报告保留，不删除或用新结果覆盖历史结论。

### 本地代码已闭合，环境验证仍未闭合

- source bundle事务锁及统一锁序、报告前后事务复核已实现并有真实unit／sqlmock行为PASS；不能用它代替真实MySQL双连接撤销竞争。
- 保存／参与者提交在paper锁与source锁后重新核凭据到期；锁等待期间过期返回HTTP401、ROLLBACK、不写答案／不复用结果。可信内部Worker提交不冒充参与者凭据。
- expiry Worker使用keyset、有界10batch并绕过失败项继续扫描；singleton DI／实际引用capacity已接；均本地测试通过，真实重启、排空及并发未验。
- 管理员续答上限5分钟且不越过原deadline；恢复同candidate／paper／exam及冻结身份、题序、答案，不重置deadline或解冻profile，不按手机号匿名恢复；HTTP与前端不得将短凭据升级成长凭据。前端355项及续答mock47项通过，不等于真实联合E2E。
- [纠正 - 2026-10-03] 下方“前端未实施／source竞争未修／缓存引用未接／管理员resume未实现”属于此前切片阻断，现已在本地代码与单测范围闭合；原4571／71／6等失败记录保留。真实环境门禁仍未解除，不能把历史代码blocker关闭解释为可直接部署或授权全部完成。

7个环境skip名称保留：TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun、TestPhase1WordTemplateCandidateUploadContract、TestBugFB170_Phase1GroupPieLabelsStayOutsideChart、TestPhase1CustomerWordTemplateLibreOfficeProducesExpectedPages、TestBugFB169_Phase1ChartGenerationPreservesTemplateLabelStyles/uploaded_staging_template、TestBugManagementTraitsPDFVisibleTestLabel、TestManagementTraitsSourceLockMySQLExternalUpdate。PDFVisibleTestLabel在最新默认全量中未启环境而skip；另一次显式enabled真实本地PDF已PASS，两次执行口径不混算。来源锁MySQL和其他skip仍不计环境验证。

### 已授权但未执行的远端续作（不得假终结）

1. 恢复SSH后先只读preflight：实际主机／服务／MySQL／LO／字体、父列容量与collation、私有目录及旧应用摘要；不打印秘密，不改网络或共享配置。
2. 完整受限DB（含routines／triggers）、旧后端／dist／模板／旧PDF备份；gzip及SHA验证，失败即停，正式备份保留。
3. 在已授权临时Schema恢复和演练DDL首次／重复执行，验证11表／15FK／索引／NULL／collation／拒绝与回滚；只创建清理确切临时目标，不覆盖production。
4. 验证guard并排空在途旧生成／上传／压缩、停止写流量；随后staging DDL first／repeat，确认完整结构后才启用新应用。
5. 从fresh source重建最终Linux／前端产物、记录完整SHA，按既有顺序仅部署TEST scope；formal／production不启用。
6. 两题本×candidate／tester四真实组合，完成HTTP／UI／PDF／SQL逐项对照、原PDF保留、expiry／revoke／concurrency／resume及重启Worker；mock和unit不能替代。
7. 按本次marked确切IDs和受控新文件路径清理临时数据／会话／Schema，遵循FK删除顺序；共享源题／bundle、旧PDF和正式备份不删。核残留／孤儿0、传统摘要、health／日志后再给最终验收结论。

上述staging测试部署及临时Schema创建／清理已有用户授权，不再要求无谓重复部署确认；当前阻断是SSH未恢复。本次文档收口没有再次执行SSH或任何远端写操作。

## 两安全项新授权最终验收：共享DI＋真实引用容量（主代理，本地GREEN）

**未验证／跳过先列出**：未连接真实DB、SSH，未执行迁移或部署；未重跑前端303、真实MySQL并发或目标PDF验收。全量7个环境skip仍保留，不把sqlmock称为真实数据库。来源锁/续答/write/report并行改动不属于下列成果，不由本报告替并行代理验收。整体产品不可据此直接上线。

### 已落地及消费方闭环

| 消费方／文件 | 本轮处理 | 原因／保持合同 |
|---|---|---|
| [candidate](../Go-based%20Refactored%20System/internal/handler/candidate.go#L25)、[tester](../Go-based%20Refactored%20System/internal/handler/tester.go#L19)、[身份DI](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_identity.go#L18) | 同步修改 | optional variadic构造参数保旧调用兼容；默认每handler只建一次，身份请求复用持有服务；实际DB/secret/预算不一致或显式nil注入失败关闭 |
| [runtime handler](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime.go#L24)、[router](../Go-based%20Refactored%20System/internal/router/router.go#L15)、[main](../Go-based%20Refactored%20System/cmd/server/main.go#L25) | 同步修改 | actual配置创建单实例；router在发布前预检；candidate/tester/runtime/PDF及Worker复用；无generic global |
| [service构造](../Go-based%20Refactored%20System/internal/service/management_traits_runtime.go#L22)、[Schema](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L111) | 同步修改 | 同pool私有callback registry、immutable DB句柄、atomic只读容量；完整失败缓存与fresh装配刷新；严格11表/15FK合同保留 |
| [身份标量查询](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L98) | 必需最小接线 | 拒绝超长ID后安全Scan.Error，不再Row().Scan nil panic；身份响应五字段不变 |
| [公开run加载](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_load.go#L130)、[实际Raw源加载](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_source.go#L32) | 必需最小接线 | public LoadValidatedRun先统一Schema；Raw.Scan后核真实relation/question/option字节预算，冻结前失败关闭；签名不变 |
| [Schema引用行为](../Go-based%20Refactored%20System/internal/service/management_traits_schema_references_test.go)、[review回归](../Go-based%20Refactored%20System/internal/service/management_traits_schema_behavior_test.go#L350)、schema/load/handler/router既有夹具 | 同步修改／新增 | 实际metadata与真实Gin/sqlmock调用，不用源码name代替能力；补真实23legacy列，不删除/skip原case |
| [handler DI回归](../Go-based%20Refactored%20System/internal/handler/management_traits_di_test.go#L110)、[router DI回归](../Go-based%20Refactored%20System/internal/router/management_traits_di_test.go#L121) | 新增 | 两request与三个入口六request完整预检querycount=1、错注入、配置不复活、失败粘滞/fresh、legacy DB无callback污染 |
| 前端303、DDL/model、公开API返回、旧001/003/00401评分/报告、客户源与模板 | 无需修改 | 本轮只依赖寿命与既有引用边界收紧；不改业务语义、数据库结构或文件；不把未重跑前端当本轮通过 |
| 并行write/report/撤销/管理员resume | 不编辑／不评价完成 | 避免覆盖其他代理工作，只计当前全量现状，不混算本轮实现 |

### 真实容量规则

- 元数据查询读取真实父及引用列，而不是仅三个父id或索引名字。引用列varchar容量必须覆盖父最大值；FK额外核列容量、charset/collation、列序及RESTRICT动作，原列/type/NULL/index严格矩阵不弱化。
- exam是外部opaque ID，仍允许父varchar1..64，不引入UUID假设。paper/pq及实际生成candidate/answer桶父键需容纳UUID36；sidecar GORM varchar64仍严格。profile_exam_id预算沿profile.exam_id→exam.id传播；多态participant按真实candidate/tester父列限制。
- 运行时按UTF-8**字节**预算检查真实参数和加载值，不以字符数冒充ID合同；覆盖生成SQL及JOIN别名、结构化条件/updates、实际model读取、profile/paper mapping与options JSON。Raw源投影显式校验relation/question/option，超过缓存父容量在冻结前拒绝。
- 缓存只在实例上；GORM私有registry共享连接池但不污染原DB。DB句柄不在首次请求替换；Set后Session{}独立Statement且复制Settings，避免拒绝错误污染下一请求。合法U+FFFD保留，非法UTF8夹具使用真实非法字节。

### RED→GREEN与最终独立证据

| 阶段／命令 | 真实结果 |
|---|---|
| 新引用矩阵RED | 159失败事件；额外source/participant边界8失败事件，exit1 |
| 跨request DI RED | handler5失败顶层；两request预检2次、配置变更复活、无共享注入；错注入后续RED22失败事件，exit1 |
| 独立review RED | Row nilpanic、JOIN绕过、加载ID/Raw source/public loader/DB替换34失败事件；零值Dest加载21子项RED |
| 主代理最终focus：go test ./internal/service ./internal/handler ./internal/router -run 'TestBugMTSchema\|TestManagementTraitsSchema\|TestManagementTraitsDDLModelContract\|TestBugManagementTraitsDI' -count=1 -json -timeout 120s | **2546pass／0fail／0skip，PARSE_ERRORS=0，EXIT=0** |
| 主代理最终go test ./... -count=1 -json -timeout 120s | **5290pass／0fail／7skip，605通过顶层，PARSE_ERRORS=0，EXIT=0**；事件含子项，不是覆盖率百分比 |
| Go Build任务＋go build -o bin/server.exe ./cmd/server | **SAFETY_BUILD_EXIT=0**，无error输出 |
| go vet ./... | **SAFETY_VET_EXIT=0**，无error输出 |
| 独立最终review | **PASS**，前次六finding在此限定范围闭合 |
| 关键10Go文件diagnostics | **0错误** |

7skip：candidate模板上传、FB170、FB185I真实MySQL并发、客户模板LO页数、FB169 uploaded_staging_template、MT PDFVisibleTestLabel未启MNG_REAL_LO_TEST、并行新增SourceLockMySQLExternalUpdate未配置真实环境。本轮未新增skip，不声称真实DB/目标环境GREEN。

### 轮次及历史纠正

- schema生产累计3/3；service构造累计3/3；引用新测试3/3，既有behavior测试含真实非法UTF8夹具纠正3/3；DI生产最多2/3。本轮不再追加代码试改。
- 中间8fail与随后28fail不是最终结果：前者Set clone0污染Statement、非法UTF8经JSON编码转合法replacement；第一次改NewDB:true导致clone1丢Settings。按当前GORM源码查证后最终Session{} clone2复制Settings且隔离错误，原非法UTF8夹具恢复真正ff字节，全部断言GREEN。没有删除RED、没有放宽合法引用预算。
- [纠正 - 2026-10-03] 下方“缓存与actual引用未接”及进度中“本轮验收失败停止”是先前/中间状态；上述两项已经实际编码并主代理独立验证完成。保留历史不删除；其他产品链／并行安全项／真实环境不由本轮GREEN宣称完成。

## 新授权有界修复 receipt：父键容量局部GREEN，其余安全项停止未完

**先列阻断：本轮没有完成全部新授权，不可部署。** 来源撤销事务锁／统一锁序、跨HTTP请求singleton DI及管理员续答凭据均未实现；实际profile/exam引用ID对缓存父列长度的运行时检查亦未接入。Go全量0fail只表示现存测试通过，不能解除这些安全门禁。短期执行在容量切片后明确停止，不拆文件绕过三轮、不继续scalar试改。

### 授权及执行边界

- 已完整读取project-memory、本报告、Go安全／架构／性能／数据库／测试／业务链／覆盖规则和当前相关service/handler/router；Execution扩展返回none，且提示无active scenario。本轮不创建scenario/task目录、不调用任务生命周期。
- 用户重新授权三个安全切片及独立resume功能；此次实际仅修容量，生产文件和对应兼容测试各**一轮**，未达到三轮上限。客户源、Word、TEST label、font、star/footer、前端五字段／冻结元数据／11路由不改，无DB/SSH/DDL/部署。
- 父键规则：el_paper/el_paper_qu必须varchar(36..64)，el_exam继续varchar(1..64)；三父列NOT NULL、utf8mb4及相同collation。11表全部列、索引、15 FK签名及动作、错误／失败粘滞缓存保持原严格检查。不同合法varchar长度不要求相等。

### 实际改动／影响

- [生产validator](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L114-L125)：新增生成UUID父键36字符下限，不改DDL/模型、不自动修父表。
- [既有兼容矩阵](../Go-based%20Refactored%20System/internal/service/management_traits_schema_behavior_test.go#L150-L175)：原三个父键全设1/32仍期待成功，与本轮明确要求冲突；保留全部case，短容量现在断言ErrManagementTraitsRuntimeInvalid，64仍通过。没有skip/delete或放宽正错误。
- [原容量RED](../Go-based%20Refactored%20System/internal/service/management_traits_schema_capacity_behavior_test.go#L10)：不改；paper/pq各1..65全部保留；[external exam矩阵](../Go-based%20Refactored%20System/internal/service/management_traits_schema_capacity_behavior_test.go#L33)及混合长度/全部collation/FK/错误/cache回归保留。
- 消费方无签名或响应变化：CheckRuntimeSchema及freeze/create/detail/fill/submit/results/report/identity仍使用同validator，**无需同步改码（理由：只收紧元数据，非法结构失败关闭）**；原签名矩阵测试**同步修改**，其他测试**无需修改（全量实跑通过）**。长生命周期DI和actual-ID检查仍是未完项，不写成无需修改。

### 本轮独立验证

| 实际命令／证据 | 结果 |
|---|---|
| go test ./internal/service -run '^TestBugMTSchemaGeneratedParentCapacity$' -count=1 -json -timeout 60s | RED：70短容量子项＋顶层71fail，CAPACITY_RED_EXIT=1 |
| go test ./... -count=1 -json -timeout 120s | **4642pass／0fail／6skip**，CAPACITY_FULL_EXIT=0，PARSE_ERRORS=0；带Test字段事件，含顶层/子项 |
| go build -o bin/server.exe ./cmd/server | CAPACITY_BUILD_EXIT=0，无error输出 |
| go vet ./... | CAPACITY_VET_EXIT=0，无error输出 |
| 两改动Go编辑器diagnostics | 0错误 |

6skip仍为candidate上传模板、FB170、FB185I真实MySQL并发、客户LO页数、FB169 uploaded_staging_template、ManagementTraitsPDFVisibleTestLabel未设MNG_REAL_LO_TEST；未改skip条件，不计环境通过。前端303是用户转交基线，本worker没有重跑或改前端，不把它记成本轮独立验证。实际DB查询未执行，不声称DDL／数据库数据状态通过。

### 剩余安全切片与resume前端交接合同（均未实现）

1. **缓存＋实际引用长度**：缓存经严格校验的父列容量，运行入口检查实际exam/profile引用ID长度；handler/router注入同一actual cfg/DB RuntimeService，identity与runtime/PDF复用，fresh装配刷新，避免global随机缓存。当前每identity请求新service、七表scope仍存在；没有新跨HTTP行为证据。
2. **撤销并发**：现有bundle非锁定读取仍保留。必须让新开／保存／提交／report两次事务复核与撤销操作共享事务锁，统一paper/bundle锁顺序、所有审批query检查error；要证明BEGIN、锁等待／释放、撤销先赢时无写／无新PDF、writer先赢时commit后才撤销及失败ROLLBACK。不能仅SQL出现FOR UPDATE即关闭本项；真实MySQL仍禁止访问。
3. **开放重登录**：用户选择管理员签发短时purpose凭据，不按手机号匿名恢复。新管理员endpoint及participant消费均尚未注册，前端不能调用假设路径；待主代理实现后再定实际路径并委派前端，旧11路由/五字段不改。
4. **待实现请求／响应字段合同**：管理员签发请求必须显式examId＋candidateId＋paperId，服务端查有效candidate、跨表唯一owner及精确paper/exam绑定；JWT管理员／全局范围与TEST范围门禁。响应需purpose受限凭据、expiresAt、原deadline及允许能力（续答或仅状态／完成）；凭据不放URL/log，签发不得更新deadline／冻结身份／profile。
5. **必须新增case**：未登录401、普通exam权限403、非TEST范围拒绝、缺／重复／未知字段、跨exam/paper/candidate、零/多owner、deleted、retired已有卷允许、revoked拒绝、凭据过期/错误purpose、签发后撤销、并发签发与提交、丢失旧token的授权恢复、到期139/140只查看状态/完成且Fill拒绝、完成卷不得重新答、原140题/700桶/字段/题序/deadline零改写。当前没有这些新case的RED/GREEN，不冒称resume可用。

下方71fail、旧前端未实施及此前各阶段记录均保留为历史；此receipt只纠正当前现存测试的容量失败状态，不纠正其他安全项为已完成。

## 本轮仅后端身份／Schema续作：BLOCKED，按三轮上限停止

**未完成项先列明：不得部署或标记完整002链DONE。**

1. 来源bundle读取仍是非锁定MVCC读取，审核撤销与保存／报告提交存在竞争窗口；顺序负向测试不能替代撤销并发序列化证据。
2. Schema父键容量门禁接受不足36字符的paper/paper_qu主键。新增真实元数据测试保留RED：70个短容量子项失败，连同顶层共71个fail事件；没有删除或skip该测试。
3. 身份HTTP每请求构造RuntimeService，完整11表门禁已移至事务前，但缓存不能跨身份请求复用；既有七表作用域存在性探测仍保留，不得称全入口已统一缓存。
4. 身份HTTP／service身份文件修复轮次审计达到／超过三轮上限，Schema行为测试文件已三轮；停止继续改码，等待用户明确授权新的有界修复。此前worker超限是流程违规，不以后续GREEN掩盖，也不拆文件绕过。
5. candidate恢复仍要求有效participant凭据；匿名同手机号、过期凭据拒绝，没有新增安全凭据恢复接口。若“重登录”包含凭据已丢失／过期的开放身份，此分支仍未完成，不能按手机号自行放行。
6. 真实MySQL首次／重复DDL、并发撤销／双连接、四组合HTTP/UI、实际重启Worker、目标环境报告尚未验证。本轮不访问SSH、不真实DB写、不部署、不改前端；formal／production仍关闭。

### 本轮已实现的后端切片

- 明确政策：end与retired仅约束新开卷，已有paper身份接续复读冻结paper／140题／700桶／唯一有效owner；当前profile草稿或字段变更不能覆盖已冻结身份、绑定、题序及deadline。审核revoked顺序调用拒绝，但上述并发窗口未关闭。
- configured-only扩展为name/gender/telephone/affiliation/post/age/degree/major/stuFlag九项，取消无条件name+telephone；typed空字符串先解码为空值，已配置required为空仍拒绝。新candidate按配置显式写入，tester主数据按配置投影，未配置姓名不在身份响应泄露。idNumber/depart没有作为新增采集字段擅自支持。
- owner审核要求跨candidate/tester恰一、exam/paper绑定、del_flag=0及tester status精确0；frozen loader不因可变资料或当前profile改写既有快照。报告DTO兼容可选冻结字段，旧五字段JSON保留严格验证。
- Schema新增完整真实sqlmock metadata矩阵：全部必需列缺失／类型NULL／索引顺序唯一性前缀／15FK签名动作／query-row-scan错误／并发缓存／失败粘滞与fresh实例刷新；拒绝畸形bigint宽度及text collation漂移。11表15FK DDL和model tags未修改。
- freeze与旧写共享排他lease到commit，追加序列化行为测试；非本任务旧链不重构。

### 本任务changedfiles（不把工作区其他未跟踪文件算成本轮）

生产9文件：[身份服务](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L194)、[冻结服务](../Go-based%20Refactored%20System/internal/service/management_traits_runtime.go#L27)、[冻结加载](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_load.go#L35)、[owner](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_owner.go#L31)、[组卷](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_create.go#L44)、[profile](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_profile.go#L50)、[严格解码](../Go-based%20Refactored%20System/internal/service/management_traits_decode.go#L74)、[报告DTO](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L75)、[HTTP身份](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_identity.go#L26)。另1生产文件：[Schema](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L105-L132)。

测试8文件：[身份回归](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity_test.go#L83)、[service链](../Go-based%20Refactored%20System/internal/service/management_traits_identity_chain_test.go#L15)、[freeze lease](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_profile_test.go#L14)、[HTTP身份回归](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_identity_test.go#L58)、[HTTP链](../Go-based%20Refactored%20System/internal/handler/management_traits_identity_chain_test.go#L17)、[真实tester配置登录](../Go-based%20Refactored%20System/internal/handler/management_traits_configured_tester_test.go#L16)、[Schema行为](../Go-based%20Refactored%20System/internal/service/management_traits_schema_behavior_test.go#L39)、[父键容量RED](../Go-based%20Refactored%20System/internal/service/management_traits_schema_capacity_behavior_test.go#L10)。共18个Go文件；文档追加本报告、project-memory、business-branches、regression-tests、coverage-history五文件。

### 前端接口交接（无签名变更）

- actualauthentry是CandidateHandler.Save和TesterHandler.LoginForm，原路径与response消费方已grep。candidate Rest code0；tester Ajax code200；data仍仅id/examId/paperId/name/participantToken五键。开放保存可仅发送配置字段；typed未配置空串可接收，configured required空值拒绝；有卷恢复须提供participant凭据，不用旧人员update重绑。
- 保持原八API请求／响应签名。create/detail给paperId、serverTime、startedAt、deadline、reminderAt、questions及paperToken；fill返回answered；submit返回runId/status/complete/answered/submittedAt/reused。
- profile的fieldContract/mappingSnapshot仍JSON字符串；result详情仍PascalCase，big.Rat分数为有理数字符串或null，不误当列表decimal。报告仍显式runId generate-test、reportId view/download。前端由并行主代理负责，本轮不评价其完成状态。
- 同步修改：身份HTTP→service→字段/profile/create/load/owner→报告适配及其测试。无需修改：八API路由和DTO签名、001/003/00401评分／报告、DDL/model tags、客户原件、历史PDF。来源锁／长生命周期缓存仍未完成，不标无需修改。

### 实际验证与计数

| 验证 | 证据 |
|---|---|
| worker身份行为RED→GREEN | 真实service/Gin/sqlmock入场、字段、冻结恢复、无旧写；聚焦2611pass/0fail/1skip exit0，随后分包复验exit0；非真实MySQL |
| workerSchema metadata RED→GREEN | RED1516pass/192fail exit1；GREEN1858pass/0fail/0skip exit0，15顶层，含子事件 |
| 父键容量新增RED | 主代理全量实际执行，70个paper/pq短容量子项失败，顶层TestBugMTSchemaGeneratedParentCapacity失败；保留待修 |
| 主代理最终全量 | go test ./... -count=1 -json -timeout120s：**4571pass／71fail／6skip，exit1，JSON解析错误0**；含顶层／子事件，不是独立用例数 |
| 主代理构建／vet | go build -o bin/server.exe ./cmd/server：FINAL_BUILD_EXIT=0；go vet ./...：FINAL_VET_EXIT=0；Go Build任务亦完成 |
| 初次统计 | PS默认解码产生JSON解析错误，初次4559pass计数作废；改UTF-8后重跑取得上行可靠计数 |
| 独立复审 | CHANGES REQUESTED：撤销竞争、UUID容量blocking；每请求新runtime缓存warning |

6skip为原五个环境用例加TestBugManagementTraitsPDFVisibleTestLabel（本轮未设MNG_REAL_LO_TEST）；不复用上一轮LO成功为本轮运行证据。未跑前端、未宣称覆盖率百分比或完整链GREEN。

## 完整授权续作：后端基础链与TEST报告（本地PARTIAL，未部署）

**仍未完成／未验证，先列出：**

1. **P0 前端未实施**：新增002创建/冻结入口、开放/封闭参与者新route、25/20服务器时间展示、管理结果run选择及显式生成/查看/下载按钮、旧准备/旧result2直达分流均未完成。现有前端26文件165项及production build通过只代表基线。
2. **P0 身份接续未完成**：现有identity admission仍在登录/登记时检查测评窗口，导致end之后或retired后的重登录未形成完整闭环；已持有效token的CreatePaper恢复路径已实际验证不受exam.end截断。身份入口仍有事务内七表probe，不是本轮11表完整缓存门禁。不要据新服务CreatePaper通过把登录闭环记为完成。
3. **P0 configured字段未完整保真**：当前身份合同仅支持name/gender/telephone/affiliation/post，现有登记还强制name+telephone；配置age/degree/major/stuFlag或其他子集尚未实现。未知配置会拒绝，不会静默改成已有五项。空字符串、已开卷资料恢复、不得改模型的完整矩阵须接续。
4. **真实MySQL／HTTP/UI E2E未验**：本机mysql/mysqld命令未找到，3306/23306无监听；没有执行DDL或任何业务DB写。sqlmock是真实Go事务执行但不是MySQL。首次／二次迁移、父键collation、FK真实拒绝、双连接竞争、两题本×candidate/tester实际140答题和重启Worker终验均未完成。
5. **环境／视觉未全验**：未SSH、staging、production或全九页视觉验收；未运行race/golangci-lint或宣称gofmt通过。原17项兼容PDF/optimized23未重跑，防止覆盖既有输出；原source-layout10已重跑通过。新v2只实际查看封面及六图页，保持原star/footer选择，不改固定文案。
6. 管理列表当前上限200，超过200拒绝，没有分页UI；新版Excel/批量ZIP/历史重算/高级current选择不在本轮实现，不回退旧导出或历史题本猜测。新报告总预算90秒；新Worker默认30秒/100批次，尚未添加独立运维配置项或真实压力证据。

**整体不能标DONE、不能部署。** 本轮不是只规划或只改标记：已新增显式DDL、完整Schema门禁、140组卷writer/公开接线、结果读、Worker及独立报告持久化与ID读取；以下记录实现证据，不将可编译/单测GREEN扩大为完整运行E2E。

### 本轮用户明确政策（覆盖同事项pending，保留旧历史）

| 项目 | 本轮明确决定 | 当前实现／边界 |
|---|---|---|
| 时间 | 25分钟，20分钟提示；服务器时间 | 新paper冻结started/deadline；详情返回serverTime/reminderAt；UI未接 |
| 未到期 | 缺题拒绝，140完成才提交 | 同paper锁私有事务通过；HTTP已接，真实MySQL未验 |
| 到期 | 140完整则completed，否则incomplete，无正式分及报告 | 现有结果构建＋新公开Submit；incomplete管理详情新增RED→GREEN；报告仍completed-only |
| exam.end | 仅禁新开卷；已开按冻结deadline | 公开CreatePaper→既有paper恢复sqlmock零写验证；身份重登录仍是未完成项 |
| offline/restart | Worker启动立即扫，原deadline不重置 | 新独立Worker接主服务启动/取消；只扫new_creation sidecar；没有真实服务重启测试 |
| retired | 禁新开卷，已开原冻结题本继续 | 新开严格candidate-current-source；existing metadata/结果允许retired；审核撤销另拒绝 |
| source审核撤销 | 禁关联新写及new report，不删除旧文件 | 可信来源只有candidate-current-source/retired；review-revoked及未知拒绝；末次报告事务复读 |
| 报告触发 | 仅管理员显式generate，不auto | 独立generate-test；Submit/Worker不调用报告；普通exam:list拒绝 |
| 来源 | 当前库原文；00202 V67/V96独立db-current身份 | 复用服务端700行SQL／Vsort/title/raw严格构建，不收client manifest；客户修订稿不自选 |
| 内容与环境 | 205原词仅TEST，formal关闭；production不操作 | 独立SHA固定原文副本；report mode=test，必有title/label；HTTP要求MNG_TEST_REPORT_ENV精确local或staging，production/unset拒绝 |
| 发布 | staging备份→迁移→测试部署→marked temporary finally cleanup已授权，由主代理执行 | 本轮仅local，未连接或修改远端；正式备份必须保留 |

### Upstream Artifacts Consumed（本次追加）

- [项目记忆](project-memory.md)与本报告既有段落：现有S1-S2D、八503、Word slots修复及原件SHA。
- [设计](management-traits-word-design-20261001.md)：§4冻结身份／V与显示序、§5精确事实、§7事务、§8六图及value-only、§9私有报告、§12精确权限、§18来源/政策；高级current提案不擅自实施。
- [业务链](business-chains.md)及[分支](business-branches.md)：保留001/003/00401边界。
- [原兼容候选manifest](generated/management-traits-word-candidate-20261001/lo-compatible-manifest.json)：88原控件、5数字槽、6图与原style/star/footer。
- [客户原文XLSX](260929管理特质测评-优化/260928测评内容+数据图.xlsx)：精确SHA的205原词TEST包，不作formal批准。

### Evidence Mapping（本次追加）

| 上游合同 | 实际代码／测试 | 判定 |
|---|---|---|
| §4/5 全140冻结、V与个人序分离、四版本 | [新组卷](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_create.go)、[四组合测试](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_create_test.go) | 两code×两身份构建140/700；严格新开窗口/retired/重复顺序负向；不是四组合真实API |
| §7/12 先结构后tx、同paper锁、不可变run | [Schema](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go)、[公开保存提交](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_write.go) | 11表全部字段类型/NULL、PK/唯一键、5非唯一查询索引、15 FK/列序/RESTRICT、关联collation；每实例一次；缺失/错误缓存；不AutoMigrate |
| §18 已开卷不随exam.end/retired重置 | [公开恢复集成测试](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_resume_test.go) | schema→exam/owner tx→释放owner锁→paper详情tx→绑定token；旧deadline/20分钟点不变 |
| §8 原稿固定字不改，只测试标记例外 | [确定性builder](../Go-based%20Refactored%20System/internal/handler/management_traits_test_template.go)、[字节保护](../Go-based%20Refactored%20System/internal/handler/management_traits_test_template_test.go) | 只加TEST浮动SDT drawing；去除此run后document原字节一致，其他OPC原字节一致；原88控件不放宽，新测试副本严格90 |
| §9 report/run身份、私有新文件、tx外LO、审计同tx | [报告core](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go)、[双事务](../Go-based%20Refactored%20System/internal/service/management_traits_report_pipeline_test.go)、[ID读取](../Go-based%20Refactored%20System/internal/service/management_traits_report_download_test.go) | 真实sqlmock证明render在commit后、末次复读DTO、revision/current/audit原子提交；失败回滚且新文件清理，历史sentinel原样；损坏PDF拒绝 |
| §12 管理鉴权／不收路径／formal隔离 | [新报告HTTP](../Go-based%20Refactored%20System/internal/handler/management_traits_report_runtime.go)、[路由](../Go-based%20Refactored%20System/internal/handler/management_traits_report_routes_test.go)、[环境矩阵](../Go-based%20Refactored%20System/internal/handler/management_traits_report_env_test.go) | 401/403/400实际httptest，local/staging真205包通过，production/unset拒绝；没有真实MySQL支持的HTTP 200 E2E |

### 逐文件实现与消费方（本轮）

- [Schema与完整DDL](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go)、[SQL](../scripts/sql/management_traits_001_runtime.sql)：11表8评分/receipt＋3报告层；显式MySQL5.7/8 CREATE，15 RESTRICT FK；不ALTER旧表、不回填、不种approved；父exam/paper/paper_qu的charset/collation须相同，否则明确失败，不自动改旧键。
- [报告模型](../Go-based%20Refactored%20System/internal/model/management_traits_report.go)：revision/current/audit独立，mode/test_title/test_label单独字段、私有file_key不序列化；run+paper+exam复合FK及current+paper绑定；不更新candidate/tester旧pdf_path。
- [RuntimeService](../Go-based%20Refactored%20System/internal/service/management_traits_runtime.go)：FreezeProfile公开先完整Schema缓存，再调用现有锁内源库冻结；原CandidateRuntimeEnabled仍false且无调用作开关，不能拿它当完整UI已启用。
- [组卷服务](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_create.go)：140题纯构建，复用既有安全shuffle，批量写legacy Paper/PaperQu/700 buckets及新快照，同事务绑定真实owner。既有卷先释放owner锁再锁paper，避免与Submit更新owner构成锁倒置。返回冻结题干/选项，无方向/评分/密码；serverTime/deadline/reminderAt由服务器。
- [Profile校验](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_profile.go)、[加载](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_load.go)、[结果构建](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_result.go)：retired不否定已冻结输入，未知/revoked拒绝；新增incomplete管理只读详情，13/4计数/NULL仍校验。报告独立loader仍completed-only。
- [公开保存／提交](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_write.go)：Schema前置后调用原同paper锁事务；不自动报告、不修损坏run。既有私有事务测试保持。
- [管理结果](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_results.go)：显式exam元数据列表、显式run完整详情；最多200无分页，列表不证明report-ready。
- [HTTP接线](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime.go)：原四participant POST、freeze/profile detail/result list/detail接真实服务；请求继续严格重复/未知/null/多选拒绝，purpose token只存本请求context，安全错误不返回DB或秘密。旧router守卫不再修改。
- [Worker](../Go-based%20Refactored%20System/internal/service/management_traits_expiry.go)、[主服务启动](../Go-based%20Refactored%20System/cmd/server/main.go)：只通过new_creation sidecar JOIN扫描，首扫立即、30秒/100条、同Submit；取消退出。不存在DDL时缓存关闭，不去扫描旧298类legacy过期卷。
- [报告core](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go)：完整可信DB模型→205原词DTO→tx外renderer→O_EXCL UUID.pdf→同paper锁复读同DTO→新revision/current/audit提交；最后成功提交为最小current，不新增attempt/epoch/高级历史选择；失败仅删本次新文件。下载同句柄有界读取/size/SHA/DTO/source/ID全部复核。
- [报告HTTP](../Go-based%20Refactored%20System/internal/handler/management_traits_report_runtime.go)：仅管理员或wildcard的POST generate-test、GET view/download；明确run/report ID；90秒同context预算覆盖DB／队列／LO／文件／提交；RFC5987/no-store/nosniff。既有LO可执行配置只用于复用基础设施，不改00401规则。
- [TEST模板builder](../Go-based%20Refactored%20System/internal/handler/management_traits_test_template.go)、[构建CLI](../Go-based%20Refactored%20System/cmd/mng-test-template/main.go)：精确原候选SHA，只加page-relative浮动用途drawing，保原字/rPr/style/media/star/footer；不同现存目标拒绝覆盖，同内容可只读复用。
- [独立渲染器及回归](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word.go)、[六图测试](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word_test.go)：改为TEST v2精确SHA，严格90＝原88＋2用途Tag；删除原错误cover递归定位，其他value-only/六图/五槽不放宽。
- [原文复制CLI](../Go-based%20Refactored%20System/cmd/mng-test-content/main.go)：原XLSX SHA/205规则验证后复制为明确TEST运行副本，不改客户、不生成formal或approved。当前所有runtime副本在configs/export-templates，候选/客户源保留。
- 新TDD与集成文件：schema、create、resume、policy、expiry、incomplete_detail、report_runtime、report_pipeline、report_download及handler test_label/test_template/report_routes/report_env。测试都在Go代码区；SQL在scripts/sql；文档仅docs。

消费方闭环：新运行handler→service/schema/writers、主服务→新Worker、新报告HTTP→core/独立Word/共享LO、renderer→TEST v2配置、测试→这些真实函数均**同步修改／新增并已编译验证**。原Candidate/Tester身份入口**尚需同步修改，作为P0明确未完成**；前端**尚需frontend角色同步修改**。001/003/00401、旧002评分/结果/Excel/PDF、原SQL迁移、客户原件、原88字段候选与旧router守卫**无需修改（理由：本轮新scope/新run/report及测试副本独立）**，全量未新增失败。没有改动用户其他既有脏文件。

### API与运行配置（现有注册前缀）

全部前缀为 /exam/api/management-traits：
- POST participant/create-paper：examId＋participant purpose header；恢复／新建。
- POST participant/paper-detail：paperId＋paper purpose header。
- POST participant/fill-answer：paperId/paperQuestionId/optionId＋paper token。
- POST participant/submit：paperId/submitType=manual＋paper token。
- POST profile/freeze：管理员，body只有examId；GET profile/detail：管理员examId。
- POST results/list：管理员examId；GET results/detail：管理员runId。
- POST reports/generate-test：管理员runId；GET reports/view、reports/download：管理员reportId；路径参数/formal字段拒绝。

新增运行环境变量不含秘密：MNG_TEST_REPORT_ENV必须精确local/staging；MNG_TEST_REPORT_DIR必须服务器私有绝对目录、不得被Nginx alias覆盖，0700/文件0600；MNG_TEST_CONTENT_PATH、MNG_TEST_TEMPLATE_PATH可指定文件，默认指向独立TEST配置；模板/原文仍精确SHA白名单。下载不依赖当前模板文件。生产／未配置报告环境拒绝，不通过APP_ENV猜测。未在本机或任何服务器设置这些运行变量来激活应用。

### Test Results（最新，不混用旧计数）

Go pass/fail/skip为带Test字段事件，含顶层/子项，不是独立顶层数量。

| 已实际执行 | 结果 |
|---|---|
| TEST可见用途初始真实Go→LO RED | 缺TEST，LABEL_RED_EXIT=1 |
| TEST v1真实用途短文案GREEN | 三用途可见、557285 bytes；随后实图发现本轮封面标题下移，v1不作为最终运行模板 |
| Schema/DDL、create、Worker、report core首次RED | 对应缺失符号编译失败，各exit1，后实现测试通过 |
| retired政策RED | 已冻结retired被拒，exit1；新旧区分后GREEN |
| incomplete管理RED | completed-only拒合法139/140，exit1→GREEN；报告仍拒绝不完整 |
| TEST环境RED | unset/production/prod/STAGING四项在真实工作簿下通过，exit1；现六环境矩阵GREEN |
| 非唯一读取索引RED | schema签名漏idx_mng_profile_bundle，exit1；五个查询索引/前缀补齐GREEN |
| 完整客户全文真实Go→LO | layout读序把左栏“测评结果”插入右栏跨行正文产生假缺字；raw读序26段＋总体三段逐字GREEN；未删除标签或改源词 |
| TEST v2真实Go→LO＋原件字节保护＋六图value-only | 3项PASS，LABEL_V2_GREEN_EXIT=0；A4 9页653016 bytes；完整文案和三用途齐全；实际封面标题恢复，六图页可见 |
| 公开CreatePaper恢复完整sqlmock tx | schema先读／exam-owner tx／paper详情tx／140题／700桶／purpose token／deadline及20分钟点保持，通过；零新写 |
| report完整双事务／audit rollback／ID下载／损坏文件 | 全部PASS；外部render发生在第一commit之后；revision3原子current/audit；失败新文件残留0、历史sentinel字节不变；下载匹配，corrupt拒绝 |
| 原候选source-layout-contract | 10 tests / 0fail / 0error / 0skip，13.269s，SOURCE10_EXIT=0 |
| 最终 go test ./... -count=1 -json -timeout 120s（MNG_REAL_LO_TEST=1） | **2569 pass / 0 fail / 5既有skip**，VERIFIED_FINAL_TEST_EXIT=0 |
| go build -o bin/server.exe ./cmd/server；go vet ./... | VERIFIED_FINAL_BUILD_EXIT=0、VERIFIED_FINAL_VET_EXIT=0 |
| npm test（既有工作区任务） | **26文件165项PASS**，Duration12.79s |
| npm run build:prod | FRONTEND_BUILD_EXIT=0，仍既有asset/entrypoint体积warning；不含新002前端 |
| Linux amd64本地交叉编译 | 最终FINAL_LINUX_BUILD_EXIT=0，49683564 bytes，SHA e7f8e40a1194bf77cbf4f4e699616a53b5f1efc874614c5730a7d23e0c7b5065；只本地候选，不是已部署版本；此前49683300-byte构建保留为历史 |
| 关键Go文件／本报告和memory最终诊断、代码／文档diff check | 均0错误；SCOPED_DIFF_EXIT=0、FINAL_DOC_DIFF_EXIT=0；不宣称gofmt或golangci-lint通过 |

5个既有skip仍为MySQL双连接FB185I、客户模板candidate upload、FB170、客户模板LO页数、FB169 uploaded_staging_template；不计环境验收通过。未运行真实MySQL、浏览器新链、race或17/23兼容视觉合同，不扩大结果。

### 产物与原件保护

- 最终运行TEST v2模板：[配置文件](../Go-based%20Refactored%20System/configs/export-templates/management-traits-002-test-only-v2.docx)，530193 bytes，SHA **05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c**；90唯一SDT＋原5数字槽/6图。v1明确留作本轮历史失败副本，不覆盖。
- TEST内容：[运行副本](../Go-based%20Refactored%20System/configs/export-templates/management-traits-002-test-content-v1.xlsx)，原字节SHA **b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c**；205规则只测试，不是批准包。
- 最终完整原文PDF：[本地TEST v2](../Go-based%20Refactored%20System/tmp/mng-full-content-local-20261003-0115/management-traits-TEST-full-content-v2.pdf)，653016 bytes/A4 9页，SHA **3eca22f70b7a09ffb17fe7fffb7e51df5cb5840551b6399769e5c5e8d39b16f9**。数据为明确synthetic sqlmock冻结fixture＋真实客户205文案，不是实际数据库人员测评。
- 同ignored tmp目录保留初版诊断／v1全文稿／对应截图及synthetic DTO，均新名称，不覆盖旧证据。真实LO converter的系统temp/profile均由defer清理；sqlmock rollback测试t.TempDir自动清理；没有需要清理的staging临时业务行。
- 原兼容候选仍 **a986ba0f3c5985486cf5f2a781346beb4140d2f2e79952a25c3beb57a273eb46**；客户DOCX仍 **c82c2dc0cdcaec413866268c557561d5c53abc41ec02d94dc6aea615edca2f48**；客户XLSX仍b049...。原件、字体／固定text／star/footer及原88字段合同未改。

### Staging命令与顺序计划（仅交接，未执行）

**先关闭上述P0身份／字段／前端未完项及完整代码门禁；不能现在执行迁移或部署。** 主代理之后按如下顺序，生产不访问：

1. 只读确认staging目标归属／SSH主体、MySQL5.7或8、LO/字体、服务、private目录不被静态alias公开；不打印cnf/env/token。核验三个父id的类型、长度、utf8mb4/collation。此DDL要求父collation一致；不一致先单独设计，不ALTER生产旧键。
2. 受限完整数据库routines/triggers备份与旧后端/dist/模板/旧002PDF、001/003/00401摘要；gzip完整性＋SHA＋0700/0600；失败停止。恢复副本验证后才主库迁移。备份保留。
3. 先确认旧写guard实际生效并排空在途旧PDF写／上传／压缩。维护窗口停止写流量。
4. 同一受控mysql会话，以私有defaults-extra-file读取凭据（变量为主代理路径，不含密码）：

	```sh
	mysql --defaults-extra-file="$MYSQL_TEST_CNF" "$TEMP_SCHEMA" < scripts/sql/management_traits_001_runtime.sql
	mysql --defaults-extra-file="$MYSQL_TEST_CNF" "$TEMP_SCHEMA" < scripts/sql/management_traits_001_runtime.sql
	```

	新DDL只有CREATE，无数据回填；MySQL DDL会自动commit，部分安装不得启用；执行器遇首错停止，不能用--force。临时库先验证11表所有列/NULL/精度、PK/唯一键/5查询索引/15 FK/RESTRICT/collation、父摘要、第二次不变、错误case，随后按既有备份纪律对staging同样首次/重复执行。临时Schema清理仅确切marked temporary授权目标。
5. 重建本机发布包，不以此文档中前次Linux hash替代最终hash：

	```sh
	GOOS=linux GOARCH=amd64 go build -o bin/server-mng-test-linux ./cmd/server
	```

	仅测试范围上传后端→两个明确TEST配置→完成的前端dist→权限修复；MNG_TEST_REPORT_ENV=staging、独立绝对private目录及精确路径注入；迁移后重启刷新Schema缓存。禁止formal/production。
6. 通过正常candidate/tester入口执行两题本×两身份：新测评／精确配置→管理员freeze→140组卷→重登同paper/题序/deadline→单选保存→缺题manual拒绝→140完整→submit/重试→run/13/4/receipt→管理员generate-test→reportId view/download，核HTTP、文件/DB SHA、完整原词/TEST标记/六图实际可见、旧PDF不变。负向含零/多owner、跨exam/paper/run、tokenpurpose/expiry、revoked/retired、end后existing续答、子表/audit回滚、同卷并发、部分/零答超时及重启首扫、未配置/production环境拒绝。
7. finally按本次生成的**确切临时主键清单**：audit→current→report revision→receipt→dimension/module→run→question snapshot→paper snapshot→legacy buckets/qu→owner→paper→profile→本次临时exam；共享源题/bundle不盲删。SQL DELETE含主键条件；新UUID PDF仅在DB引用移除后按受控私有路径清理，旧PDF/正式备份不删；短时会话及temp schema清零。终验孤儿0、传统摘要保持、health/日志；报告完整证据后等待下一授权。

### 本轮学习／边界

- TEST用途只写DOCX metadata或unsupported Choice不够；必须真实PDF文本及实图可见。流式标记会推动原浮动封面标题，最终使用专门page-relative TEST drawing，原稿所有对象不改。
- PDF表格layout读序会插入左栏固定标签，不能直接拼全文误判缺字；保留用途layout检查，全文用raw读序逐字比对及实图。
- 新卷create恢复不能持owner锁再等待paper锁，否则与submit的paper→owner更新倒置。释放恢复事务后再同paper锁读取。
- 11表完整缓存需要PK/unique/FK之外的expiry/read索引；Raw扫描全部显式column tag。主服务Worker与handler各自实例缓存，不声称跨全进程仅一次探测。
- 不把fresh valid token恢复的通过扩张为完整重登录；身份admission窗口／字段和前端是明确P0未完成。无后台同步子代理工具模式可用，本轮未伪称frontend已委派完成。

开始UTC：2026-10-03T00:43:54Z；最终构建/文档核验UTC：2026-10-03T01:33:37Z，约2983秒。整体请求仍PARTIAL，不能以本轮backend GREEN记完整[DONE]。

---

## 独立有界修复补充：MT-WORD-RUNTIME-01（2026-10-03）

**限定图表bug：本地GREEN；整体仍PARTIAL / 不可部署、不可启用。** 用户重新授权继续诊断后，仅一轮生产代码修复使原失败测试通过。下方原2512/1/5及三轮停止记录保留为历史，不代表本次最终状态。

**未关闭项先列明**：真实DB/HTTP/UI/报告持久化及目标服务器验收仍未执行；formal和运行入口未启用。实际Go输出PDF文本层没有“测试报告”，尽管DOCX中含该标记；按本轮仅图表范围保留原标记定位逻辑，不以结构断言宣称可见用途标记已通过。LibreOffice输出既有 `Could not find platform independent libraries <prefix>` 警告，全50稿实际转换退出0；未修改外部工具配置。编辑器测试入口在编辑后返回未发现测试，以下GREEN以真实原生Go命令为证据。

### 根因与最小修复

- [渲染器数字槽分支](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word.go#L108-L112)原先硬编码 `len(texts)==3`。真实候选总体与任务模块分别有4个 `w:t`，第4个是原稿尾随空白；其余三个模块为3个。候选manifest的numeric_text_slots明确指向第2个文本节点，并未要求节点总数为3。
- 六个原生anchor的 `wp:docPr name/title` 正确对应 `chart.overall`、self/task/development/interpersonal模块、dimension.comparison；关系rId12～16及rId22依次指向chart1～6。失败发生在总体数字标签，而非业务key、关系、图表类型或系列点数。未按类型猜映射、未重编号OPC。
- 只将数字槽节点数校验改为3或4；4节点时仍要求末节点为纯空白，继续只替换第2个值。未知键、重复键、六图/五槽基数、系列/点序、模板SHA及formal拒绝不放宽。
- [保留原测试并增强](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word_test.go#L34)：总体28.85、自我25、人际75、任务100、发展0；逐项断言六图全部数值及常模、五独立数字槽、每个chart除数值外XML字节一致、正文除文本内容外XML字节一致、其他所有OPC部件字节一致。原测试名和失败路径保留。可选 `MNG_TEST_WORD_OUTPUT` 仅用于本地测试产物，不是运行开关。

### 消费方与改动范围

- grep确认 `renderManagementTraitsTestWord` 的代码消费方仅为同包Word测试：渲染器与测试同步修改；没有函数签名或API结构变化。
- service DTO/205逻辑文案、实际router、旧生成链、00401/MBTI、活动模板及客户原件：无需同步修改，本轮未改变它们的数据、函数、配置或运行门禁。未覆盖工作区其他既有修改。
- 本轮改动源文件为上述两个Go文件；文档仅本报告、project-memory、business-branches、regression-tests和 [progress-details](progress-details.md)。临时DOCX/PDF/截图/profile放项目ignored tmp下；既有合同的测试产物留在原generated证据子目录。无SSH/SQL/DDL/DB写入/部署/激活。

### 本轮真实验证

| 命令 / 检查 | 实际结果 |
|---|---|
| 原测试及增强RED：`go test ./internal/handler -run TestManagementTraitsTestWordIndependentValueOnly -count=1 -v -timeout 60s` | 两次均FAIL，`charts`，exit1；先RED后修代码 |
| 002聚焦：`go test ./internal/handler ./internal/service ./internal/model ./internal/router -run ManagementTraits -count=1 -json -timeout 120s` | 1813 pass / 0 fail / 0 skip，`FOCUSED_GREEN_EXIT=0` |
| Go Build工作区任务及 `go build -o bin/server.exe ./cmd/server` | 任务完成；明确 `BUILD_EXIT=0` |
| `go test ./... -count=1 -json -timeout 120s` | 2525 pass / 0 fail / 5既有skip，`ALL_TEST_EXIT=0`；事件计数含子项，不与旧2512机械比较为新增13独立测试 |
| 原候选脚本 `--libreoffice-compatible --visual-boundaries` | 17 tests / 0 fail / 0 error / 0 skip，98.458s，exit0；真实五边界PDF/像素/长文案门禁保持 |
| 同脚本 `--source-layout-contract` | 10 tests / 0 fail / 0 error / 0 skip，12.959s，exit0 |
| 两Go文件编辑器诊断 | 0错误 |
| 可选产物测试补充后的最终全量及构建复验 | `FINAL_ALL_EVENTS={pass:2525,fail:0,skip:5}`；`FINAL_ALL_TEST_EXIT=0`、`FINAL_BUILD_EXIT=0`；文档增量check=0，全50稿第3页18个50.00（五数字槽＋十三柱值）、可见占位符0 |
| 首次全量输出统计 | PowerShell非UTF-8解码导致JSON解析失败，计数作废；设置UTF-8后完整重跑取得上述证据 |
| 首次 `python` 调用 | PATH入口exit9009，未运行合同；改用已存在虚拟环境python.exe后真实17/10通过 |

全量五项skip仍为下方列出的既有环境用例，不计真实MySQL/旧上传模板/目标LO验收通过。未重跑optimized23，不引用其上游结果作为本轮证据。

### 实际Go → 本机LibreOffice PDF

- [差异化Go输出](../Go-based%20Refactored%20System/tmp/management-traits-word-fix-20261003/management-traits-runtime-test.pdf)：A4 9页、542964 bytes；SHA `f8f3142a2a14b8c2d1984b5fb3f4681ee37db7036db4ad206c340be2ae654948`。五标签实际提取25.00/75.00/100.00/0.00/28.85，13维分值50.00与常模均显示；可见业务占位符0、U+FFFD=0。0分环的分值扇区为空属于数据/原样式，不据此误报缺失业务绑定。
- [全50分Go输出](../Go-based%20Refactored%20System/tmp/management-traits-word-fix-20261003/management-traits-runtime-all50.pdf)：真实函数二次生成，经隔离profile的LibreOffice26.2.5.2转换，`LO_ALL50_EXIT=0`；A4 9页、542627 bytes；SHA `2aa0708b46624f4e4a30ef3c1879ea435f0478c1b682fda0dfb865cab0ac9c5e`。第3页PNG实际查看：总体＋四模块五环图、13根分值柱及常模线均可见，不以ZIP六部件代替可见性。仅核六图页，不宣称全报告视觉或目标环境验收。
- 候选SHA仍 `a986ba0f3c5985486cf5f2a781346beb4140d2f2e79952a25c3beb57a273eb46`；客户DOCX仍 `c82c2dc0cdcaec413866268c557561d5c53abc41ec02d94dc6aea615edca2f48`，客户XLSX仍 `b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c`。没有变模板绕过失败。

---

**状态：BLOCKED / PARTIAL，不可部署。** 全量Go测试为2512 pass事件、1 fail、5 skip；新002 Word渲染测试在`charts`阶段失败，已达单文件三轮修复上限并停止。没有新PDF、真实SQL、LibreOffice、浏览器完整链或staging验证证据。不能将编译通过称为完整实现。

## 授权边界

- 用户10月3日明确批准当前题库原文冻结；00202 V67/V96不得冒称客户管理修订稿、不修改旧源题。
- 批准完整002测试链和独立新版run/report；旧结果/PDF保留，不自动历史重算。
- staging备份、迁移、测试范围部署及临时数据验后清理已授权，但本worker仅本地工作；主代理独立准备staging。本轮未SSH、SCP、数据库写入或部署，production不动。
- 客户现有文案仅测试报告口径；未决错字/频率/统计措辞不自动修订，formal门禁保持关闭。
- 时间/窗口/退役/epoch/批量/历史高级路径不按推测启用。本轮没有增加运行开关。

## Upstream Artifacts Consumed

- [项目记忆](project-memory.md)：既有S1-S2D、原稿优先及LibreOffice兼容候选边界。
- [设计R2及生命周期合同](management-traits-word-design-20261001.md)：§3-8版本/来源/精确事实/字段/六图，§12权限与旧写保护，§18未决运行政策。
- [业务链](business-chains.md)：旧001/002/003与00401隔离职责。
- [兼容候选manifest](generated/management-traits-word-candidate-20261001/lo-compatible-manifest.json)：88唯一SDT、5普通数字槽、6图及源样式合同。
- [客户工作簿](260929管理特质测评-优化/260928测评内容+数据图.xlsx)：精确SHA锁定的原始条件文案。

## Evidence Mapping

| 上游合同 | 本轮实现/证据 | 结论 |
|---|---|---|
| 设计§12：旧scope保护必须接实际router | [实际安装](../Go-based%20Refactored%20System/internal/router/router.go#L51)、[真实Setup HTTP回归](../Go-based%20Refactored%20System/internal/router/management_traits_routes_test.go#L17) | 三旧公开写入口RED→403 GREEN；两管理GET未登录401 |
| 设计§5/8：精确事实、按等级匹配现有文案、post/submittedAt | [SHA锁定内容](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L35)、[强DTO](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L124) | 两测试通过；仅纯适配器，未接HTTP/可信DB读取事务 |
| 候选88/5/6＋仅替值、原星形页脚/样式保留 | [独立渲染实现](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word.go#L61)、[失败契约](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word_test.go#L34) | 新符号缺失RED后仍charts失败；不得渲染/启用 |
| 设计§18：未决政策不得默认启用 | 八个既有handler及公开freeze/fill/submit仍关闭 | 本轮没有解除503或formal门禁 |

## 本轮改动与影响清单

1. [router.go](../Go-based%20Refactored%20System/internal/router/router.go#L51)：仅把已有`ManagementTraitsLegacyScopeGuard`装在实际JWT前。未改Setup签名或其他路由。
2. [management_traits_routes_test.go](../Go-based%20Refactored%20System/internal/router/management_traits_routes_test.go#L17)：真实Setup→HTTP执行三受保护旧写端点及两管理员未登录读取。使用sqlmock，不连接真实数据库。
3. [management_traits_test_report.go](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L35)：只接受确切客户工作簿SHA，读取65摘要＋65评价＋65建议＋5总体评价＝200条、再5总体三段建议＝205逻辑规则；不修订原文，题本无切换，内容包不能激活formal。
4. [报告DTO构建](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L124)：重新复核完整run/13维/4模块/receipt、冻结身份/字段、输入SHA、提交时点及精确评分；按精确等级选原始文案、仅展示两位、图值12位。新增API只被新测试引用；现有loader/HTTP未接入，不能当真实数据库授权证明。
5. [DTO与源合同测试](../Go-based%20Refactored%20System/internal/service/management_traits_test_report_test.go#L17)：SHA/205条/50.00评分事实、冻结姓名、三段建议、13/4/3/3；损坏hash/sum/身份JSON及incomplete拒绝。输入是标明synthetic的存储fixture，不把Word样例59.45当评分期望。
6. [独立Word渲染器](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word.go#L61)：未完成、未接任何HTTP消费者。不调用00401渲染器，候选SHA白名单，XML位置仅替值，封面既有标题位置追加测试用途标记；五数字槽和六图合同尚未通过，不能声称rPr/六图/PDF最终已验证。
7. [Word测试](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word_test.go#L34)：保留真实失败，不skip、不降低88/5/6、不删除RED。尚未产生可接受输出。

调用/消费方：
- 实际router：同步修改，安装已存在守卫；旧handlers：无需修改，受保护对象在handler前被拒。
- candidate/tester身份分流：无需修改，本轮沿已有实现；既有schema失败和legacy兼容行为仍由原测试覆盖。
- 新内容/DTO：仅新service测试消费，不改变现有API返回结构。
- 新Word：仅新handler测试消费，无运行路由或活动模板安装。
- 001/003/00401/MBTI评分、结果、活动模板：无需修改，不新增任何产品版本/分值变化；全量测试除新Word失败外没有新增失败事件。5项既有环境skip仍保留。
- 前端：无需同步当前未接入DTO；未实现002新入口，不声称已实现完整UI链。

## Test Results

计数口径为`go test -json`带Test字段的pass/fail/skip事件，包含顶层及子项，不是独立顶层用例数。

| 实际执行 | 结果 |
|---|---|
| 基线 `go test ./internal/handler ./internal/service ./internal/model -run ManagementTraits -count=1 -json` | 1806 pass / 0 fail / 0 skip |
| 真实router RED `go test ./internal/router -run TestBugManagementTraitsRealRouter -count=1 -json` | exit1；三旧写端点逃逸403保护 |
| 同命令GREEN | `ROUTER_GREEN_EXIT=0`；4 pass事件 / 0 fail / 0 skip；3旧端点403、2新管理GET401 |
| 内容/DTO RED | 新符号undefined、exit1 |
| `go test ./internal/service -run 'TestManagementTraitsTest(Content|Report)' -count=1 -v` | `REPORT_DTO_EXIT=0`；两顶层PASS |
| `go test ./internal/handler -run TestManagementTraitsTestWord -count=1 -v` | `WORD_GREEN_EXIT=1`；一项FAIL，安全阶段`charts` |
| 工作区Go Build任务，随后 `go build -o bin/server.exe ./cmd/server`明确复验 | `BUILD_EXIT=0`；server49929216 bytes |
| 工作区Go Test All任务及 `go test ./... -count=1 -json` | `ALL_TEST_EXIT=1`；2512 pass / 1 fail / 5 skip |
| 工作区Frontend Vitest任务 `npm test` | 26文件 / 165项PASS，Duration32.18s |
| `npm run build:prod` | `FRONTEND_BUILD_EXIT=0`，2个既有资源体积warning |
| 6个本轮Go文件编辑器诊断 | 0错误；不代表Word业务测试通过 |
| 已跟踪router增量 `git diff --check` | exit0；未将CRLF格式提示称为raw gofmt通过 |

全量唯一失败：`TestManagementTraitsTestWordIndependentValueOnly`。原有5 skip：
- `TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun`：未配置专用MySQL环境。
- `TestPhase1WordTemplateCandidateUploadContract`。
- `TestBugFB170_Phase1GroupPieLabelsStayOutsideChart`。
- `TestPhase1CustomerWordTemplateLibreOfficeProducesExpectedPages`。
- `TestBugFB169_Phase1ChartGenerationPreservesTemplateLabelStyles/uploaded_staging_template`。

本地原始日志位于项目ignored logs目录，本轮无秘密环境变量值输出。系统查询未找到mysql/mysqld命令，3306/23306监听数0；不能把sqlmock称为真实SQL验证。

## Remaining Blockers

- **HIGH / 当前代码失败**：Word渲染器charts阶段失败。已三轮，须由主代理重新定位/确认后继续；不可绕过模板SHA或六图断言。
- **HIGH / 核心链未实现**：create-paper writer、profile/detail和结果list/detail真实接口、参与者前端、独立report revision持久化/授权ID下载尚未接通；八新handler仍503。
- **HIGH / Schema未完成**：没有新增002显式DDL、完整列/索引/FK/collation签名，不可执行AutoMigrate或借七模型宣称表已安装。
- 真实两题本×candidate/tester、140题×五档/极值、并发提交/双连接/rollback、旧PDF不覆盖、六图实际PDF/全文完整性门禁尚未取得完整本地证据。
- 时间窗口对已开卷、用时/截止顺序与离线结算政策仍未确认；当前私有事务的25分钟/manual→timeout演算不构成启用批准。
- 本轮未运行旧17+10+23候选Python合同；先前主代理GREEN证据是上游，不冒充本轮重跑。
- 不创建modernize任务/上下文目录；本workspace无该会话目录，交接和经验按用户约束只放docs。

## Staging交接必要清单（未执行）

1. 先关闭上述代码/政策/Schema阻断，Go全量必须0fail、明确skip原因；重新运行实际Word→LibreOffice及完整PDF文本/图像门禁。
2. 主代理预检staging身份、服务/字体/LibreOffice/MySQL版本及0700私有PDF目录，不查看/打印凭据；production不访问。
3. 受限全库含routines/triggers备份、gzip完整性、SHA；旧后端/dist、002旧PDF与现用模板备份、001/003/00401源/结果摘要。
4. 先部署旧写保护、排空在途生成/上传/压缩；明确迁移窗口和回滚顺序。
5. `scripts/sql`显式MySQL5.7/8 DDL，经首次/重复执行、完整签名/FK/collation/NULL/unique实证；不ALTER旧分值、不改旧源题、不回填历史run。
6. 仅临时标记测试范围部署新后端/独立002测试配置/前端，保持formal关闭；不是正式激活。
7. 两题本×candidate/tester真实新测评→冻结→140答题→提交→13/4/run→管理员测试PDF→按report ID查看下载，逐项核SHA、正文/六图/等级、并发重试/rollback/身份cross/原PDF保留。
8. 验后按确切临时主键清理测试记录/文件/短时会话（正式备份不删），复核临时残留0、孤儿0、传统摘要不变、health/日志；主代理报告，production仍不部署。

## Learnings

- 仅在隔离引擎测试守卫不能证明实际Setup安装，必须真实装配HTTP回归。
- 该候选ZIP有10个合法空目录条目，严格路径校验须区分尾随目录slash和路径逃逸；原件不重建。
- 真实正文以logo表开头，标题是后续直接段落，不是body第一子节点；不得按未查证的首节点假设定位。
- 样式保护采用字节位置替值的独立002实现仍须所有图表合同及实际PDF验收，本轮charts失败不可交付。

开始：2026-10-03T00:09:40Z；最终只读核验：2026-10-03T00:27:20Z（约1060秒）。未写[DONE]，任务仍阻断。

## 受控恢复库 MT-GUARD-AUDIT actual阶段 PASS（2026-10-03T12:45:57Z）

**边界先列明**：本阶段仅恢复库/生产公开service函数/真实MySQL守卫验证PASS；不是新版HTTP/UI/PDF整体验收，不是已部署。未执行main element DDL、服务替换/重启、配置/grants变更、前端构建或部署，未访问其他production。应用账号连接与启动、guard/drain维护窗口、四真实受测组合/报告与Worker仍待下一阶段；性能/race未测。本地默认全量新增一个明确实库环境skip，共9skip，不能折算成0skip。测试新增Go的规范化gofmt差异仅EOF末换行，未再格式修改。

### 源码核实、授权范围与永久证据

- 完整读取本报告及project-memory后续作，原源码只有[TestManagementTraitsGuardAuditMySQLExternal](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard_audit_test.go#L481)，没有用户提到的StagingExternal名称。本轮新增[独立真实验证用例](../Go-based%20Refactored%20System/internal/service/management_traits_guard_staging_test.go#L59) `TestManagementTraitsGuardAuditMySQLStagingExternal`；不重命名/改写旧用例，不修改生产逻辑/原001 SQL，不用现代化任务框架。
- SSH实际vm-ubuntu-go-dev/liming，strict known key/BatchMode/ConnectTimeout10/ConnectionAttempts1；MySQL8.0.46-0ubuntu0.24.04.4，实际APP_ENVproduction＋REPORT_EFFECTIVE_ENVstaging。正式受限备份沿用 /opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9，永久保留；root0700/top-level文件0600已重新核验。
- 三归档完整SHA/gzip/tar均PASS：DB **9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1**；application **079a718a96d1fd43d0fcf037d3796362684d2211400d975258f0d929c90d5fbd**；files-and-system **21fa940fd5800f70e33c047c42e13148ddeeebe8a1a2240c466c765a81ff998d**。恢复前后两应用/文件系统归档均tar --compare退出0；没有覆盖旧备份或chmod共享目录。
- 唯一恢复库 **mng_guard_audit_test_729c47a8d0194b63**；唯一另行空锁fixture库 **mng_source_lock_test_729c47a8d0194b63**。正式证据保留 /opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/guard_actual_729c47a8d0194b63，含restore/DDL first-repeat/完整first-repeat-post签名/实际schema-post-lock日志/receipt/payload-SHA256SUMS/payload-cleanup-receipt。日志不含秘密/token/PII/DSN。
- 本轮Linux bootstrap **8451164 bytes/SHA d9863f7e15594c47b813412d35ac82738408620348d91104320298929578c795**；fresh service测试二进制 **19895312 bytes/SHA 96b7570d650c6cc958f868b37765248813c3308f1c2f4a3167f92094137480f6**。均本地build exit0，上传仅独立payload，远端SHA逐件OK，未上传/替换main server。两临时二进制已最终exact删除，不复用旧service.test为本轮证明。
- 当前源重新构建的本地Linux server **49828320 bytes/SHA 0ee9b326c459611babb6951c63291abcdfc27491e8782d3dd9178e4bb56d8483**，build exit0；只是本地准备产物，不能称已运行环境版本。后续任何生产源码变化必须重新构建核完整SHA。

### 实际生产完整Schema、metadata shape及来源

- 恢复路由检查无CREATE/DROP DATABASE/USE，完整备份只恢复到owned库；所有旧/新cross-schemaFK实际0。未改legacy父列或将旧FK指向element。
- 未改001 SHA **7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8**，首次/重复均11表15RESTRICT/RESTRICT约束；完整列/type/NULL/charset/collation、索引/前缀、逐FK列/序/目标/动作签名first=repeat=post，SHA **3fb28c5c220c063795e77b3e27b8eed8ace01b0dc7ee82614a556af8f1c3c0d1**。本签名加入FK逐列细节，与上阶段1020f旧签名算法不同，不作为结构变化。
- 首次、重复及post各fresh实例执行实际NewManagementTraitsRuntimeService→CheckRuntimeSchema，全部nil/PASS；每个实例连续两次调用仍仅四次metadata查询。Rows.Columns来自**该生产查询返回的实际sql.Rows**，GORM Trace直接记录同查询行数；没有额外诊断SELECT替代validator、没有标签转换或mock。

| 生产查询 | 实际驱动columns | actual GORM rows |
|---|---|---|
| tables | table_name,engine | 11 |
| columns | table_name,column_name,column_type,is_nullable,character_set_name,collation_name | 159 |
| indices | table_name,index_name,non_unique,seq_in_index,column_name,prefix_length | 67 |
| FK | table_name,constraint_name,column_name,ordinal_position,referenced_table_name,referenced_column_name,update_rule,delete_rule | 21（15约束的复合列） |

- post实际公开FreezeProfile两次：恢复库当前原文00201和00202分别140题/700选项；manifest SHA **66277c8bd4f2d51fc2ab51020625e6fe27b8df10ae02381bdb023b5773e97e0d / 4a033e1bf8f9a39dc66b92ef36ecd38bfd0ed9f598a4c295be1b32cd72ee5e2d**；mapping SHA **dc7faac623b10bc42c4e39b493518469ef0c14e586b35e6d0f7ddaafbd1e58bf / f0f7387dbc9ab9ff95205262baab1b2185b56288e0bcc9f866ac1c114e0889b6**。仅owned新marked exam/link/profile/bundle按确切主键清理；两共享repo/源题不改，00202仍db-current原文。
- bootstrap内部从运行进程读取环境，实际在/opt/talent-assessment调用config.Load消费application及production覆盖配置/实际JWT秘密；核原DSN确为element/local3306后只将子进程DSN换exact owned库、既有root unix socket，无用户/grants/env/cnf文件。子测试cwd保持payload/Go-based Refactored System/internal/service供原DDL fixture读取，config.Load输出“本cwd无配置，env only”提示；实际秘密由bootstrap内存注入，未生成临时配置副本。不是应用账号授权/完整server启动验收。
- 实际配置密钥CreateManagementTraitsRuntimeToken→ParseManagementTraitsRuntimeToken内存roundtrip PASS；没有输出/落盘token，不发身份HTTP、不创建业务会话。

### public capture/旧Scope与审计闭包：非skip真实PASS

- owned capture只有participant_type/participant_id/paper_id与synthetic sentinel；真实**ManagementTraitsIdentityScope＋CandidateIDs＋PaperIDs＋AllLegacy**均protected=true,error=nil。完整11canonical＋capture及七core＋capture两形状分别全部执行PASS；七表形状只暂时RENAME四个空owned extras到非el_mng前缀，再exact恢复，最终完整签名与初始相等，不删除/弱化生产gate或遮蔽canonical audit。
- 原外部用例作为子用例canonical-audit-rollback原样调用：从恢复库既有candidate/paper/exam真实关系选owner，事务中插synthetic bundle/snapshot/run/revision/audit；audit.report_id→revision.id→run.id复合paper/exam及candidate/AllLegacy/公开identity闭包全部保护，ROLLBACK后11表逐表count0。不是完整评分数据或正式报告。
- 新实库负向missing-revision/missing-run/identity-drift各在**同一owned连接的回滚事务**设置SESSION FOREIGN_KEY_CHECKS=0仅造故障数据，不改FK元数据；恢复checks=1后ROLLBACK。空scope/AllLegacy/无关paper均返回ErrManagementTraitsRuntimeInvalid，protected=false，不能吞孤儿或裸成功；每案回滚后11表0。
- unknown report_id-only owned表及临时缺receipt的partial安装，真实公开CheckManagementTraitsLegacyScope均失败关闭；exact DROP unknown/RENAME receipt恢复后fresh完整Schema再次PASS。未在main或任何其他库制造故障。
- 主post运行 **5pass事件/0fail/0skip，3.67s**（主用例＋原canonical子用例＋3负向子用例），首/重复schema-only各1pass，整个GUARD_ACTUAL_REMOTE_EXIT0；最终actual post未跳过。不以本地sqlmock或名称存在代替执行。
- 原TestManagementTraitsSourceLockMySQLExternalUpdate fresh测试二进制独立复验 **4pass/0fail/0skip，2.65s，exit0**；commit/rollback/MVCC-currentread，真实第二连接普通UPDATE1205/释放后成功。仅子进程随机localhost→unix relay；原测试自清表后空库count0，未改测试逻辑或grants。

### 最终清理、旧系统不变及本地回归

- 主演练receipt **EXIT=0 TEMP_SCHEMA_REMAINING=0 CLEANUP_ERRORS=0 MAIN_MNG=0 PDF_COUNT=465**；schema与lockschema最终exact DROP=0。上传 /tmp/mng_guard_staging_729c47a8d0194b63 核owner liming0700/完整SHA/恰4文件后仅删bootstrap/service.test/未改SQL/SHA清单及空目录，payload remaining0。
- 本地唯一tmp/mng-guard-staging-20261003下main/manifest/两二进制按actual源码及SHA核所有权清0；apply_patch Delete返回后仍以实际磁盘为准，最终精确删除、目录不存在。既有bin服务器候选和其他任务tmp不删。没有秘密/token文件/环境残留；正式备份和原历史失败日志保留。
- 旧12表dump指纹 **f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e**，465PDF逐文件清单及server/dist/config/templates/unit/archive完全不变；主表数/exam/paper/paper_qu/paper_qu_answer/candidate/tester仍 **67/70/1487/134354/294628/1348/27**。实际server **ee4566e7a698ff592acaeab40c5986974787be27929cd1c2bc002d1adddcd300**、index **2d4ba6c5ac1e745d736f8c2894cb99406d4890123bb18fb0c0d8bb001c8e9bc9**。三服务active、health ok；GUARD_FINAL_REMOTE_EXIT0，UTC12:45:57Z。
- 本地最终go test ./... -count=1 -json -timeout180s实际 **6039pass/638通过顶层/0fail/9skip/解析错误0/exit0**；build、vet ./...、临时bootstrap vet全部exit0/零error输出。前8skip完整保持，仅新增StagingExternal在本地缺DSN明确skip；远端实库非skip与本地默认skip两口径分开。
- 新增测试源码/脚本诊断0，asset scoped diffcheck0；没有生产代码修复、DDL改动或失败后试改。JS wrapper只执行node --check及--verify-assets，实际远端首轮由等价手动编译/上传入口执行下面保存的同一bootstrap和shell；**不声称wrapper已经整轮实库重放**。

### 同次实证资产保留与下一阶段指引（不自动部署）

- [真实Go测试](../Go-based%20Refactored%20System/internal/service/management_traits_guard_staging_test.go#L59)SHA **1725e78aeab8d4516b8f64994e3d83545873db185cb02fa248e9f2f092636e4b**；canonical原测试未改。原guard/identity/Schema/FreezeProfile/token/锁函数和所有API无需同步改码：签名及行为不变，仅被验证。
- [同次恢复/DDL/实库/指纹脚本](../scripts/db/management-traits-guard-staging-verify.sh#L1)SHA **eb5227a3765a29176613c54695ab47e6bf735d3b9826fd84e9d92fab540980f2**，实际以LF stdin执行/bash-n0；[exact bootstrap source fixture](../scripts/test/fixtures/management-traits-staging-bootstrap.go.txt#L1)规范化SHA **ad1a0b29baf7dfcdc333d3d96155514c7e353e748e9a2a38e6274985b2e5b92a**，与已执行临时main规范化逐字节相等；fixture只是测试数据，不是B区活动源码。
- [可复用显式重放入口](../scripts/test/management-traits-guard-staging-real-test.js#L1)SHA **ec39c6bc8a2ab60a9caf2e2c11a750774e01897e87e3c03df92325b9ccbd0a14**：--verify-assets已实际PASS；--owned-staging才允许受控实库重放，重新build/copy、随机16hex exact库、源码env/matching schema/multiStatements=false/interpolateParams=false/root既有socket/原cwd保持，finally只删known own文件；没有secret或依赖新增。以后复用此资产，不再临时重新手写验证器。
- **待deploy配置/路径**：实际unit /etc/systemd/system/talent-assessment.service、Userliming、WorkingDirectory/opt/talent-assessment；基础configs/application.yml＋application-production.yml，APP_ENVproduction不能当报告环境。四MNG_TEST key在运行进程仍全部未设置，以下目标本轮实际仍不存在：
	- MNG_TEST_REPORT_ENV=staging（显式TEST，formal/production不开放）。
	- MNG_TEST_REPORT_DIR=/opt/talent-assessment/private/management-traits-test-reports，后续新建liming专有0700/新PDF0600，不得被Nginx alias覆盖，不使用共享tmp/uploadPath或/data/uploadPath（两者实测liming0755）。
	- MNG_TEST_TEMPLATE_PATH=/opt/talent-assessment/configs/export-templates/management-traits-002-test-only-v2.docx；精确SHA **05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c**。
	- MNG_TEST_CONTENT_PATH=/opt/talent-assessment/configs/export-templates/management-traits-002-test-content-v1.xlsx；精确SHA **b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c**，205原词仅TEST。export-templates实测liming0750，不覆盖旧模板。
- **下一阶段由部署负责人执行**：重新核现场与备份适用性/fresh Linux及fresh前端完整SHA→实际guard/drain在途旧生成/上传/压缩与停写→已授权staging主001 first/repeat/完整gate→独立TEST配置/后端/前端/权限/重启cache→两code×candidate/tester四真实HTTP/UI/PDF/SQL＋expiry/revoke/concurrency/resume/Worker→marked新数据/文件/会话finally cleanup与465旧PDF/旧指纹终验。当前不执行这些步骤；phase PASS不代表产品整体DONE或production上线批准。

[纠正 - 2026-10-03] 报告顶部“fresh实际恢复/完整capture guard待验”和上阶段ACTUAL_IDENTITY_CAPTURE_BLOCKED，现由本次fresh恢复库真实public capture、AllLegacy和audit闭包PASS**限定解除**。历史失败记录不删除；main迁移/部署/完整四组合E2E仍未执行，不把限定纠正扩大为已上线。

## 主 staging 部署实际执行，最终回滚（2026-10-03T13:01:44Z）

**DEPLOYED=NO／失败先列出**：新应用重启后核验器以空对象调用participant/paper-detail却期待401；实际HTTP400（生产先校验paperId），exit1。按用户任何error停止要求未修业务／断言、未再次部署。general/001/002/00401/MBTI HTTP只读及fresh前端浏览器、四组合/PDF/Worker均未执行，不称完整TEST链PASS。

- 已实际完成两次old/new PID/cgroup/8092/MySQL应用连接排空、新守卫TEST禁用运行、**main element未改001 first/repeat：11空表15RESTRICT FK**、三fresh应用账号positive_app生产fullgate均PASS，metadata11/159/67/21、缓存4查询、CRUD权限/两002ASCII140/700实证。无共享旧表ALTER/DML/backfill、源题写或其他production操作。
- 新独立TEST资产/env/dropin/private和fresh dist曾安装；HTTP失败后旧server/dist共享设置恢复、新运行资产/目录/上传exact清0。**不自动DROP11主表**；每表0行，最终总表67→78，旧exam/paper/pq/bucket/candidate/tester仍70/1487/134354/294628/1348/27。
- rollback owner/mode实证偏差已依据**正式application归档**恢复旧server root:root0755；不能用被私有证据chmod0600的server.before副本当原mode。初次本地只读命令引号syntax失败及两元数据终验失败保留，不掩盖。最终整application tarcompare0、3备份SHA OK、旧12表f57f35…/465PDF不变、三服务active/healthok、finalSSH0。
- 正式证据=/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/deploy_3461565581060643；失败receipt、exact源码/执行script/DDL、签名、旧app/config备份及失败新front/template归档、cleanup/rollback/finalreceipt永久保留。完整路径、全部SHA、配置key、阶段和未验项见[部署实证报告](management-traits-staging-deployment-20261003.md)。
- 本次fresh Go6039/0/9 parse0、vet0、Linux build0；前端31files355pass/build0。候选server SHA0ee9b326…／index98547b68…只是本次构建和失败部署候选，不是最终在线版本；在线仍旧ee4566e7…／2d4ba6c5…。不把final cleanup0覆盖部署exit1。
- 下一从**11空表已存在**状态复验，不重复假定mng0，不删表／还原全库。先独立修核验请求（合法paperId缺token401与空body400分开）和archive metadata回滚处理，再有界续部署；本轮明确停止，不自动执行四组合。

[纠正 - 2026-10-03] 上一阶段“mainDDL未执行”现由本轮11表15FK实际安装限定纠正；“尚未完成TEST应用部署/整体验收”仍成立。旧应用已恢复，不能称已部署或上阶段全量闭环完成。