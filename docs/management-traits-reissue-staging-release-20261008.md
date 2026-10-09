# 002 独立报告 staging 发布接续 — 2026-10-08

## 2026-10-09 真实管理员浏览器验收：PARTIAL（仅native download event未观测）

- 复用用户确认登录的集成浏览器，先清page/context routes并刷新；真实`getInfo` HTTP200。00501/00502从旧`exam/users` URL均自动进入新版结果页，旧`tester-list` 403为0；各1条completed、140/140、50.00，分数详情各13维/4模块。普通历史002旧URL继续传统页且`tester-list` HTTP200；冻结兼容002旧URL进入新版页并显示58.64。一次快速reload触发公共重复提交保护，页面“重试入口探测”真实恢复，未放宽保护。
- 两产品初始reissue均0；经UI各生成1份，generate HTTP200。重复UI生成均HTTP200、明确“已复用”，同report ID、同file SHA/bytes；view/download均HTTP200 `application/pdf`。最终旧TEST namespace仍2份，独立reissue namespace新增2份，合计4份且均保留。新审计为generate/reuse/view/download=`2/3/2/3`；没有手填report ID字段。
- 00501新PDF 647985 bytes/SHA=`4d40da09…`，00502为648159 bytes/SHA=`2b28aba0…`。DB `data_sha`均等于归档原字节SHA；两私有reissue文件regular/0600、size/SHA与DB/API/本地合成下载一致。独立无认证context对同一view/download均401且响应无`%PDF-`字节；未创建低权限账号、未输出token或storageState。
- 当前PyMuPDF对既有byte-preserved hardened baseline也把柱图/常模线坐标整体报告在新Y区间，原绝对坐标子断言因此失败；不是新PDF布局漂移。两新PDF与各自已通过hardened oracle的retained baseline逐页文本9/9、144-DPI像素9/9及drawing signature全部精确相同；当前递归PDF object扫描每份899 xref/40 stream、credential findings0、附件0、未加密。fresh canonical合同仍140题/100正/40反/13维/4模块，identity-v2 COMPLETE。
- 集成浏览器真实下载按钮已产生HTTP200，但其native `download` event在60秒内未出现；已按既有工具层限制从同一认证UI响应内存捕获PDF字节并通过loopback保存，未导出浏览器凭据。因此业务/API/文件/PDF验收均PASS，整体精确标记`PARTIAL_NATIVE_DOWNLOAD_EVENT_UNOBSERVED`，不冒native事件PASS。
- SSH只读终验PASS：PID15554，server=`d30c8e40…`、index=`c4f3b8f…`、health200；保护的管理特质DB SHA仍`5e80140d…`，冻结002仍58.642639，baseline runs/dims/modules=`2/26/8`，draft0，应用严重错误0、Nginx5xx0。无部署/restart/代码修改/答案修改/production访问。机器证据见[acceptance verdict](../scripts/test/results/mng005-staging-browser-acceptance-20261009/acceptance-verdict.json)。

## 2026-10-09 发布包分离与完整回滚演练：PASS

- 仅`20.200.136.133` staging；production未访问。runtime包与acceptance包已物理分离：runtime归档SHA=`e077c400…`、400文件；acceptance归档SHA=`4a5122df…`、11文件。逐文件（含393个展开dist文件、二进制字符串、配置、SQL、脚本）高置信秘密命中均0；JSEncrypt完整marker命中经核为生成PEM代码、无PEM Base64 body。
- 受限备份实际旧server=`4179fd3f…`、旧index=`52eecf04…`、393文件full manifest=`8633c580…`。成功演练PID`13961→14209→14293`，恰2次正常stop/start；旧health200、legacy detail200/results401、经短时内部认证reissue404；恢复新版本后新server=`d30c8e40…`、index=`c4f3b8f…`、393文件manifest=`cfe77d95…`，reissue匿名401/认证409、legacy仍200/401。
- rollback前/旧运行/新恢复后三个空新增表、受保护DB SHA=`5e80140d…`、私有PDF manifest=`3d209586…`、配置SHA=`ed9837af…`、冻结002=`58.642639`全不变；三服务active、双staging环境、应用严重错误/Nginx5xx均0。
- 第一轮因全局JWT先返回401而错误期待旧未知路由404，失败后立即恢复新发行并保留证据；第二轮用2分钟Redis-only临时管理员会话准确取得旧404/新409，最终无Redis/header残留。任务总计另有首轮恢复启动1次，但成功drill restart count严格为2。
- 最终runtime test binary0；历史受限backup中的5个test binary均0600且顶层目录0700；旧payload、本轮上传脚本和短时认证文件均清零。完整证据见[回滚演练报告](management-traits-staging-rollback-drill-20261009.md)。结论仅为staging PASS，不批准production。

> **项目决策更新（2026-10-08）**：采用[005独立产品政策](management-traits-product-code-decision-20261008.md)：00501基层员工新版、00502干部新版；原002全部功能及正常新建保留，现场已冻结新版002按原来源映射兼容，不改码/退legacy/重算/删PDF。下文旧002混合发布scope尚未执行部分在计划层面挂起，旧“批准继续有效／沿原scope接续”不再作为继续发版依据；后续须明确005与现场兼容范围及原安全门禁。保留已部署事实、候选、失败收据和待部署两表/draft/approval组件；此更新不停止后台、不取消活跃任务、不回滚源码，也不批准正式内容或production。本轮无SQL/SSH/部署，运行状态未重新核验。

## 2026-10-09 当前工作树统一 staging 发布：PARTIAL

- 发布目标仅 `20.200.136.133`，production未访问。当前HEAD为`5217ce6558eb7876d9e4cc37f1a617296deb2590`；旧候选未复用。新server SHA=`d30c8e40e0c99dc525ad42b1269581665dd72acb7e1bfd47958ac0efd1561fc7`，index SHA=`c4f3b8f76b740244bd6b1d10ac6e24e4650f15367f1c427e4b46055297e21555`，393个前端文件逐SHA通过。初始宽泛secret扫描命中4个URL/password属性或JSEncrypt私钥头代码字面量；脱敏上下文复核后，要求完整私钥块或带引号静态凭据的精确扫描为0，不把初始命中删除或伪称首次即0。
- 受限备份 `/opt/talent-assessment/backups/mng_current_d0e8202eafb14b08` 完成并校验。恢复库draft003/reissue004各执行两次；真实双连接得到create1/reuse1、精确input unique 1062、report1/audit4、source unchanged，临时Schema清零。
- 主库只安装draft和reissue三表，重跑通过，最终新增表0行；未安装formal/approval。后端仅重启一次，最终PID13210；三服务active，内外health正常，报告环境双staging，最近关键错误及Nginx 5xx均0。
- retained 00501/00502各1run/13dim/4module，冻结002 run1、普通002 repo2保持；受保护旧sidecar dump发布前后SHA一致，私有PDF3。临时payload/演练库清零，备份永久保留。
- 未完成：当前浏览器无真实登录会话，按安全要求未伪造token/storageState，因此reissue生成/查看/下载、HTTP并发复用及临时005 draft/person全链未验；独立CodeReviewer未执行。判定 **STAGING PARTIAL / PRODUCTION NO-GO**。

## 08:15Z MT-ADMIN-DIRECT：当前只读入口恢复，永久分流仍未发布

**未发布先列**：本轮只恢复指定管理员浏览器的可读结果页，不替换前端或后端、不安装两表、不重跑1062演练。统一reissue实库门禁仍以此前失败/PENDING记录为准；没有新独立源码/产物review，也没有最小frontend-only可发布包，不把工作树dist或全reissue UI用于这次403修复。

- 五问沿用户已给信息：staging旧管理URL直达；是否曾正常未知；指定冻结00201管理员；接口403/传统表空；exam1791298091700970647。不索取或保存姓名手机号、JWT、答案或报告。
- 08:13:52Z当前页正常GET getInfo HTTP200/code200，正安全数字userId且idOne/adminRole/wildcard均true；公开POST Detail HTTP200/code0，同exam、legacy/legacy、00201、严格布尔frozen=true，字段键name/gender/telephone，无lifecycle字段。08:14:10Z管理员profile/detail HTTP200/code0/sameExam/frozenAt存在。
- 同页实际失败GET `/prod-api/exam/api/tester/tester-list?pageNum=1&pageSize=20&examId=1791298091700970647`：HTTP403/code1，响应字段仅code/msg/success。已上线入口实际compiled方法先依URL/session意图helper和cached管理员helper再查profile，未消费服务器frozen标记；无意图旧地址回旧list。403来自正确旧保护，不是本轮新结果API权限不足；原“暂无数据”不能当无人完成。
- 已在用户指定原page仅手动router导航至 `http://20.200.136.133/#/exam/management-traits-results/1791298091700970647`。08:14:56Z正常专属POST results/list HTTP200/code0/success=true，精确同exam、1条且completed1、140/140、overall58.642639；真实UI表格显示1条completed/140/140。正常点击只读“详情”，GET results/detail HTTP200，同run/paper/exam、13维4模块。旧TEST页面总体直接显示精确有理数327050/5577，本轮独立换算两位为58.64；**不声称旧UI已显示58.64**，不扩为分值显示修复。
- 页面保留该专属路由和只读详情。该页资源记录报告generate/view/download/reissues请求0，未点击续答/报告动作、不保存截图/PII；只读HTTP自身正常访问日志不冒绝对系统零写。SSH/SQL/DDL/部署/restart/真人写/生产访问均0。
- 已有[本地入口修复](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L302-L367)和[永久实际SFC回归](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-admin-ui.spec.js#L232-L285)足够覆盖本次已核布尔-only合同，不重复编辑产品或新造RED。08:15本轮原生只复跑管理SFC/API两文件：167pass/0fail/0pending/native0；当前入口原生SHA09790e74c6a53d1da97229805c562a6365074b4cb2ceff08d6b95a358a34b920与既有修复记录一致，API5be56c…保持。全560/build0仅之前记录，本轮不冒fresh构建/全量。
- 08:15:56Z公网首页HTTP200/16155bytes/SHA52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860，仍旧发行。当前两个reissue服务实现/测试及schema已读current，未修改或以历史源码覆盖用户新改动。
- 主协调者下一安排独立CodeReviewer：从已上线identity candidate-ui原始基线提取**仅管理入口SFC**到已存在旧TEST结果页的最小server-frozen/fresh授权切片，保持old API/profile与旧TEST结果组件，不引入newreissue wrapper/draft/formal；完成独立source及compiler产物闭包门禁后，沿已有staging授权考虑单前端发布，无backend restart/DDL。本BreakGlass未派agent、未制作/批准该包；手动导航只是当前读恢复，旧URL自动分流反馈仍开放。

## 07:57Z MT-REISSUE-1062：LOCAL GREEN，真实修复复验待SSH／独立复审，仍未发布

**未完成先列**：真实MySQL新复验未启动，独立CodeReviewer未执行，统一发布及真人新报告仍禁止。07:42:04和07:43:43 strict/Batch/Conn10/Attempts1均连接前timeout/native255、stdout0、stderr67bytes同SHA；首次加唯一重试后停止，不第三次SSH或HTTP绕行。[首次](../scripts/test/results/mng-reissue-stage-20261008/bug1062-preflight-red-1791445324660.json)、[重试](../scripts/test/results/mng-reissue-stage-20261008/bug1062-preflight-red-1791445423067.json)。本轮远端恢复/GRANT/DROP/上传/主DDL-DML/部署/restart/生产操作均0；两个随机owned计划仅本地收据，不是已创建库或cleanup PASS。当前线上事实仍以07:28历史终验为最后证据，不把PID2746/4179/52eec/主两表0/旧58.64说成本轮fresh实核。

### 根因与最小修复

- 原真实失败仍是一成功/一1062、reports1/audits1，原收据只记录errno，**实际冲突键未取得**。代码已核：普通run读取先建立read view→paper UPDATE等待→bundle SHARE当前读，但随后报告Find仍快照读；原代码把report INSERT错误一律丢为sentinel，没有回滚后加载winner的分支。新增实际服务SQLmock证实该报告输入唯一键失败路径；这不是宣称已取得原实库read view/constraint时序。
- [产品修改](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_api.go#L129-L262)：仅报告INSERT捕获errors.As至MySQLError且Number1062、Message尾部精确匹配已核schema的`uk_mng_reissue_input`（带table或不带table两形式）；不使用raw err.Error字符串token匹配、不接受任意1062/ErrDuplicatedKey翻译猜测。事务返回并ROLLBACK后，一次fresh **reuse-only**事务重走原paper UPDATE→bundle SHARE/完整来源和140/13/4资格，归档原字节必须不变；不再次render/INSERT report、不sleep/poll或改隔离级别/锁。
- winner必须精确同run/paper/exam、DataSnapshot原字节/DataSHA、template/content及完整归档；原读取器核私有路径、regular file、PDF头、完整大小与SHA。成功generate审计须恰一且绑定同report/paper/exam/创建actor/有效时间，然后reuse审计与返回同事务完成。审计失败拒绝；返回winner原ID/actor/时间/file metadata，不以loser替换。所有复用途径共用验证，既有archive只读接口未改。
- record的CreatedAt和随机ID不进入输入SHA；源DTO只有原提交时间，源paper/bundle/questions的原时间保持，未新签钥/重新签原档/规范化丢字段。失败或复用仅清闭包保存的**本次loser key**，不按返回winner key删除；旧PDF/current/pdf_path/result无writer变化。

### TDD及真实执行证据

- [持久回归](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_api_test.go#L516-L693)23分支：精确winner、缺winner、跨run/paper/exam、归档字节/template/content/kind/status漂移、缺PDF/hash/size/path漂移、缺generate审计/reuse审计失败、当前读source撤销/末次归档漂移、其他主键1062、audit1062、1213、字符串伪异常和ctx取消。每次只render1，旧file及winner保留、loser无孤儿；**仅sqlmock/本地文件**。
- 首test缺gorm导入build失败不是行为RED；补导入后[有效RED](../scripts/test/results/mng-reissue-stage-20261008/bug1062-local-red-behavior.json)5pass19fail/native1（含顶层）。第一GREEN尝试109pass2fail/2skip，source-drift夹具在核心加载阶段提前拒绝且剩expectation；仅修夹具为末次原始归档读取漂移，revoked改成实际bundle SHARE revoked，未放断言。测试本文件累计三编辑后停止。最终[完整专项](../scripts/test/results/mng-reissue-stage-20261008/bug1062-focused-green.json)111pass/0fail/2明确LO环境skip/native0。
- 首全仓带120s包限额命令134.755秒/native1、5423pass/0fail事件/10skip；未保留timeout stack，不能据elapsed确定唯一cause或称全绿。保持120s子限额，按[真实发现的全部服务顶层名称](../scripts/test/results/mng-reissue-stage-20261008/bug1062-full-discovery-1791445681844.json)四分区＋其余所有packages执行（无删/skip测试），[全量汇总](../scripts/test/results/mng-reissue-stage-20261008/bug1062-full-green-1791445927090.json)6383pass/694通过顶层/0fail/11原环境skip/parse0/所有child native0。较旧6359增24事件/1顶层，不是覆盖率或实库通过。
- Windows server/full build/vet各native0/输出0；[最终精确隔离复验](../scripts/test/results/mng-reissue-stage-20261008/bug1062-scope-final.json)专项73pass/0fail/1LOskip、fullbuild/vet/Linuxserver/externaltest编译各0/输出0。真实HTTP权限/incomplete/来源/path旧矩阵保留；不前端/full560/真实LO重跑。

### 精确隔离与后续责任

- 118个已批准运行输入只有这一服务变化；当前与[隔离源](../Go-based%20Refactored%20System/bin/mng-reissue-stage-20261008/source/internal/service/management_traits_reissue_api.go)逐字节同SHA`261b90bba4fabb6a8b74c66b5a629dc862561f8ac0ecdf5ae8d437e980f620cb`。新green overlay把**同一现有测试文件**映射至当前SHA`3a40e2006f2e9ef10cfdf467b1644628c341e9dadb9c7992f1b2c11e7655205d`，不复制另一个不同版本或夹带draft/formal。同步时编辑器format与末尾CRLF造成两次byte gate失败，原生gofmt/trim比较证实无逻辑差异，最后精确同步PASS；隔离文件三编辑后停止。旧overlay中的旧测试拷贝/旧source-candidate清单是**历史**，后续不得直接拿它们当新发行闭包。
- 原发布server候选`ab07083e…`及原binary-candidate收据保留未覆盖；仅本地验证Linux SHA`3b2b1290f8c856b6d6c6ada4bc1152754a3b1f6b0f889b12458b4cb93f455f83`，最终外部测试binary SHA`df0478c8b92e76659c995ea658c2f38812d835ab0bbb7ae51e0982134dc40bb4`。不构建可发布整包、不更新install-package或deployment/rehearsal verdict；DDL SHA`d5c5f3d1b8fbcbc1ca61a8d1e8efbcfa16187c511934fc8cb9dbac861005bc12`保持。源和测试根目录diagnostics0；ignored bin部分源gopls单独当嵌套模块报undefined/import错误，实际受控overlay全build/vet通过，不为消除编辑器partial-module诊断补未批准源/改go.mod。
- 影响清单：Generate公开签名/DTO/handler调用点无需修改（合同不变）；List/Read/Qualify及原shared loader/guard无需修改（原来源与归档门禁保持）；model/schema/DDL/旧current/pdf_path无需修改（零结构/旧写变化）；对应服务测试、隔离服务源/显式green测试映射、真实fixture及必要账本同步修改已完成。新增typed classifier和winner validator为private，仅Generate本文件调用。
- [既有真实fixture](../scripts/test/fixtures/management-traits-reissue-staging-external.go.txt)增加两个独立连接的REPEATABLE-READ校验与**render外部结束后、paper lock前两个普通run读视图屏障**；仅固定操作/errno/精确键类别，无SQL/人名/JSON输出。保持报告1、总audit4原门禁，并要求generate1/reuse1、同report/data/file SHA、files1/orphan0/source原run不变。夹具已Linux编译，**尚未远端运行**。
- [C区本bug有界入口](../scripts/tools/mng-reissue-bug1062-20261008.js)无publish mode，复用原受限备份/保护协议，随机独占恢复库、原两轮DDL签名、应用账号/两个pool及exact finally；green绑定最终external binary/source/test SHA。主代理先独立安全复审，然后SSH恢复后按原批准完成exact restored库RED具体键与GREEN至少双connection验证；不得直接调用旧install publish或覆盖旧失败。当前仅ready_source_local；实际新key/复用、remote audit计数/孤儿/主保护fresh均PENDING。原统一staging授权保留，不重复审批范围，发布门禁未关闭。

## 07:28Z 失联发布恢复：BLOCKED，服务健康，主两表未安装／版本未切换

**未完成先列**：真实恢复库双连接幂等演练失败；统一发布、新管理页面和真人独立TEST报告生成/查看/下载均未完成。本轮仅恢复实际状态，不重复CodeReviewer/full560/build。用户交接兼容切片独立PASS继续有效，但不替代真实MySQL门禁。不会把“no response returned”解释为未操作或成功。

- 07:28:42Z staging strict SSH首次成功，native0/signal null/error null/stderr0，父Node native0；仅已知liming/pem、Strict/Batch/Conn10/Attempts1。实际PID2746；backend磁盘和进程SHA均`4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c`，front index SHA`52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860`。三服务active/healthok、两报告环境staging；全state1=0、overduecompetency=0。
- READ ONLY实际mainreissue两表0；指定exam completed1、overall58.642639（显示58.64）。旧rows/schema/source/score/profile/PDF/private/config/cache与受限备份baseline匹配。没有覆盖旧report_current/pdf_path/答案，不重算旧卷。
- 精确owned恢复库`mng_reissue_test_76f547bc8ae0833b`和授权均0；payload/next/server pending路径均不存在，deploy-receipt及main首/重跑签名不存在。本机仅既存node9316/1764，未匹配reissue脚本；远端限定进程投影仅本次bash15915。没有第二份发布、kill或广泛临时清理，不用该限定投影宣称系统全部后台任务为0。
- 既有backup ownership核对、四SHA及DBgzip完整性通过，目录0700/文件0600违规0。DBgzip SHA`773072035e5f422863ac4f21b151c86cca12075b2e09f406364ca90967355b26`，永久保留；本轮未重新备份/上传/执行DDL/重启或修改配置。完整证据：[fresh只读恢复收据](../scripts/test/results/mng-reissue-stage-20261008/recovery-readonly-1791444527298-1.json)。
- **实际发布前阻断**：[最后演练](../scripts/test/results/mng-reissue-stage-20261008/restored-rehearsal-1791444160717-1.json)native1，而非成功wrapper误报；首/重跑结构签名相同`1ebadfd6818723bc83465ed4c85d43ff0a9da6e5f397e1e53537ef082ae2ec3c`，finally owned schema0/errors0。[安全双连接诊断](../scripts/test/results/mng-reissue-stage-20261008/actual-mysql-concurrency-diagnostic.json)明确一成功/一失败、MySQL1062、reports1/audits1、未同report/dataSHA复用。具体冲突键与完整根因尚未取得，不把SQL错误号当完整根因；不删除失败、改弱断言、第四次演练或跳过后发布。
- local backend候选SHA仍`ab07083ee2a972ae66343110bdab9b80310d77d934cd3e6eb3f1cd77cde25908`。fixed前端3/3声明输入真实消费且当前声明源SHA匹配，markers存在、393资源/index`5b47f939ec03a9a00ed0ee860be4e19e1cae158dfdda542207ae7f2ff83f2b3d`；[fixed模块收据](../scripts/test/results/mng-reissue-stage-20261008/ui-module-proof-candidate-fixed.json)仍BLOCKED/81未归因，closure/rehearsal/deployment verdict及approved-artifacts不存在。只是既有磁盘状态，不新复审、不推定81为81业务改动、不伪造packagegate.ready。
- **历史限定纠正**：下方“备份/上传/真实两轮迁移未开始”为早期事实。失联调用确已完成受限备份、上传和恢复演练及exact清理，但主库DDL/主服务切换未发生；07:24终验与本次fresh核验均旧版本健康。未安装主两表，所以真人reissue新报告未创建，页面新UI/保存/查看/下载SKIP；没有正常登录要求或伪SESSION操作，没有本地真人PDF/PII/token/storageState文件、生产访问0。
- 当前事实阶段：兼容源码步骤completed（沿用户交接独立PASS）；actual-stage2-publish blocked/preflight-readonly completed，mainDDL/backend/front switch notstarted；report acceptance notstarted，owned temporary cleanup completed（依据原收据和本次零残留实核）。本会话无manage-todo能力，不改现代化scenario任务、不把blocked写completed。[机器恢复汇总](../scripts/test/results/mng-reissue-stage-20261008/recovery-state-20261008-0728.json)。

## 本地兼容修复接续：LOCAL GREEN，待主协调者独立复审，未发布

**未验先列**：本轮没有执行独立 CodeReviewer、当前隔离发行闭包归因、真实管理员浏览器、真实 MySQL 首次/重复两表迁移或跨连接竞争。下方原发布阻断及失败收据保持历史；本轮只关闭旧 DTO 的本地 SFC 兼容失败，不宣布 release PASS。此前统一 staging 授权保留，后续部署及重启由主协调者接续。

- 根因：已发布旧 Detail 的 strict frozen=false 是可信旧路由事实；本地未部署 draft schema 的两个字段被错误当必需。仅[管理入口](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L314-L334)三行产品修改：合法非空无外侧空白字符串/正安全整数 ID 必须匹配当前 exam；存在的新版标记严格校类型和值；strictfalse 可接受标记缺失或可选一致的 legacy/false。
- **实际政策**：完整 legacy/legacy、真实002 classifier、同exam、明确布尔false缺一不可。lifecycle 缺失只在此严格前提下可接受，不把 unknown 默认为false；任意存在的 unknown/null/undefined lifecycle 或畸形 newMode 拒绝，true newMode 缺 draft 或与 legacy 矛盾拒绝。合法明确 draft/true 仍经原 fresh getInfo 回编辑；stricttrue 无lifecycle仍经字段白名单、当前管理员及同exam frozen profile进入专属页。URL mngTest 不是服务器标记或权限，不能把可信false改成新，也不能解锁未知。
- [既有实际SFC单测](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-admin-ui.spec.js#L232-L285)固化两份实际脱敏路由投影，不依赖ignored收据、不增姓名/字段数 API。历史177682…旧list恰1/profile0/新UI0，当前非管理员getInfo夹具不授新版能力；指定179129…stricttrue专属replace1/旧list0/profile1，未改变58.64结果组件。后二者只本地适配，不冒线上认证或结果重验。
- 有效 RED：114tests/109pass/5fail/0skip/native1；GREEN 专项两文件167pass/0fail/0skip/native0；完整前端33文件560pass/0fail/0skip/native0，较原525新增35。原39回归完整保留：仅把原strictfalse无标记负例按新明确政策改成unknown lifecycle负例；report metadata force/所有权、旧list query/迟到/destroy矩阵未删改。
- 首次测试收据相对路径越过根目录 ENOENT/native1 是工具失败，非行为 RED；之后按实际层级修正并有效复现。测试文件本轮三次编辑、产品一次，不第四试改。原生 child 的JSON计数和输出完整保留于[RED](../scripts/test/results/mng-reissue-stage-20261008/compat-local-red.json)、[专项GREEN](../scripts/test/results/mng-reissue-stage-20261008/compat-local-focused.json)、[全前端](../scripts/test/results/mng-reissue-stage-20261008/compat-local-full.json)。describe suite计数不冒文件数，实际testResults=33。
- [生产模式构建](../scripts/test/results/mng-reissue-stage-20261008/compat-local-build.json)原生exit0/Build complete，2既有体积warning保留；两代码编辑器diagnostics0。仅本地当前dist重建，不当最小隔离可发布包；没有修改依赖、字体、客户资产或 Go 候选。
- [保护范围收据](../scripts/test/results/mng-reissue-stage-20261008/compat-local-scope.json)：实数internal Go=222（非236全树口径），聚合ffe5ab6a…前后相同；API5be56c…及隔离Linux候选ab07083e…保持。源入口e925…→09790e74…；候选后端未重建，未把当前draft/formal后端混入。Go全量/6272/6359、本轮SQL及真实并发未跑，不声称其PASS。
- 影响：只此SFC产品同步修改；created/retry调用签名、旧列表消费者、结果组件、API权限、Go安全后端无需修改（合同/原分流及所有权保持）。必要回归和文档同步更新。SSH/远端HTTP/DB/DDL/备份写/upload/部署/restart/真人身份答案/PDF/旧卷和旧PDF操作均0；不启动或反复运行120秒browser诊断。

## 当前结论：BLOCKED，尚未发布

用户已批准仅 20.200.136.133 staging 的两张独立报告表、五个 reissue API、管理页分流及随后指定完成结果的新 TEST PDF 生成/查看/下载。该批准继续有效；不包含草稿、正式审批、生产、历史补证或旧 PDF 覆盖。

本轮发布前发现真实 DTO 与待发布管理入口不兼容，**未执行主库 DDL、备份写入、上传、服务替换/重启、报告生成或下载**。真实 MySQL 首次/重复迁移、恢复演练、跨连接竞争、隔离前端完整编译归因、真实管理员浏览器验收均未开始。不能把隔离后端 GREEN 当统一发行完成。

## 1. staging 真实只读预检

- 2026-10-08T04:32:48Z strict SSH 首次连接超时/native255；唯一重试 04:32:50Z native0/stderr0，hostname 门禁确认 vm-ubuntu-go-dev。不调整网络/密钥/权限。
- 线上 PID2746，磁盘及实际进程后端 SHA256 均为 `4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c`；首页 SHA256 为 `52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860`。三服务 active，health 为 `{"status":"ok"}`。
- READ ONLY SQL：全产品 state1=0；el_mng_ 当前十一表，无独立 reissue/draft/formal 表。指定 exam1791298091700970647 completed=1，overall_score=58.642639。这是预检时事实，不替代未来切换前安全检查。
- [首次失败与重试收据](../scripts/test/results/mng-reissue-stage-20261008/preflight-1791433958849-1.json)、[成功预检](../scripts/test/results/mng-reissue-stage-20261008/preflight-1791433968893-2.json)。无人员正文、凭据、snapshot JSON 导出。

## 2. 隔离后端候选已建立并本地验证

[C区准备入口](../scripts/tools/mng-reissue-stage-20261008.js)只消费此前已发布最小身份 candidate-overlay，不包装当前整工作树。原线上运行加载器、守卫、路由及身份 Detail 基准保留；新运行闭包包含独立 model/schema/service/handler、原 audited adapter 依赖、共享纯文案构建器，仅在原注册函数追加五 reissue 路由。原树业务文件未改。

[源码候选清单](../scripts/test/results/mng-reissue-stage-20261008/source-candidate.json)记录全部118个运行源路径/SHA及九个新增/覆盖输入（含两个测试）。已扫描排除 `CheckManagementTraitsDraftSchema`、`managementTraitsFormalTables`、`ManagementTraitsExamLifecycle` 依赖；该扫描不是独立 CodeReviewer 或完整权限/实际数据库闭包证明。

- 专项两次均49pass/0fail/1skip/native0；明确跳过 `TestManagementTraitsReissueAPIMaskedRealAnswersLibreOffice`，本轮未启用真实 LO。最新[测试事件](../scripts/test/results/mng-reissue-stage-20261008/test-events-1791434045571.json)。不重跑或冒用全工作树6359测试、旧实库/PDF成果。
- 初次 `go build ./...` native1：复制的 overlay 输入在 bin 子树被当独立包扫描。仅准备工具增加复制原 go.mod/go.sum 的嵌套模块边界，原项目依赖/运行源码不改；原失败 [构建收据](../scripts/test/results/mng-reissue-stage-20261008/compile-1791433984519.json)保留。
- 同一 overlay 的全包 build、vet、Linux server build 随后各native0/stdoutstderr0：[全包构建](../scripts/test/results/mng-reissue-stage-20261008/compile-1791434045572.json)、[vet](../scripts/test/results/mng-reissue-stage-20261008/compile-1791434052682.json)、[Linux构建](../scripts/test/results/mng-reissue-stage-20261008/compile-1791434059734.json)。
- [二进制候选](../scripts/test/results/mng-reissue-stage-20261008/binary-candidate.json)：49991517bytes，SHA256 `ab07083ee2a972ae66343110bdab9b80310d77d934cd3e6eb3f1cd77cde25908`，仅本机，未上传。

## 3. 新发布阻断：历史002管理入口合同失配

04:35:44Z 对历史 exam1776822816300709851 和指定新 exam1791298091700970647 分别正常公开 Detail POST，均 HTTP200/code0。仅 DTO 在内存供**当前 SFC 默认 created 链**使用；管理员 getInfo/profile/旧列表用本地 adapter，不发真实认证或列表请求。因此不是线上新页面浏览器成功或正常管理员授权验收。

| 分支 | 真实服务器DTO | 当前SFC实际行为 | 结论 |
|---|---|---|---|
| 历史00201 | 同exam、legacy/legacy、strict frozen=false；lifecycle/newMode 两字段均缺失 | legacy list调用0、entryBlocked=true、未分流 | **RED：违反保留历史入口要求** |
| 指定新版00201 | 同exam、legacy/legacy、strict frozen=true；同样仅bool投影 | replace=ManagementTraitsResults、旧list0、profile adapter1 | 当前SFC分流通过，非真实管理员浏览器验收 |

当前入口 SHA256 `e925c048c35c2e15ebe8dddd241da1ad42441d5fc22230120b8fafe36c191020`。根因定位到[false分支](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L333-L337)：要求 lifecycle=legacy 且 newMode=false，否则按draft要求拒绝；而[原已发布Detail](../Go-based%20Refactored%20System/bin/mng-identity-minimal-20261007/candidate/internal/handler/exam.go#L368-L377)只有可信 frozen 布尔，不返回两个字段。旧单测正向夹具主动补了两字段，未覆盖此次真实旧合同。

[脱敏实际DTO/SFC收据](../scripts/test/results/mng-reissue-stage-20261008/frontend-compatibility.json) native1，历史pass=false、新版pass=true；原错误保持。不为了发布改后端加入草稿生命周期、放宽 unknown/矛盾门禁或直接发布整dist。尚未修改该产品分支。

## 4. 收口与推荐下一责任人

- 发布批准保留，不重新询问同一范围；先由前端 TaskExecutor 针对**已发布严格布尔合同**补持久 RED、最小兼容修复，并保持已返回 lifecycle/newMode 的矛盾拒绝、unknown拒绝、历史旧链及新冻结分流。
- CodeReviewer 对该切片与新的隔离发行闭包只读复审；本 BreakGlass 不派发代理、不自审冒独立PASS。
- 通过后才按原批准继续隔离前端构建/归因、fresh维护预检、受限备份/恢复库两轮DDL及实际新结构守卫/竞争验证、主库两表安装、后端前端统一发行、指定页面及真实新版 TEST PDF 原生下载、独立SQL与旧PDF/score/current保留核验。
- 真人报告仅应用私有目录/DB内归档；下载若执行，只存用户安全目录，不放Git或可搜索results。当前真人生成/下载0，不存在须删除的真人报告或新恢复库；旧两完成历史卷不生成。
- 本轮仅新增C区工具/报告/脱敏运行收据及必要账本/记忆，新增B区 ignored 隔离源码与构建产物；原业务/测试源码、客户模板、配置、旧收据、用户浏览器、生产、Git提交/push均未操作。