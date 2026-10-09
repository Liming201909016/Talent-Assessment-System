# 002 staging 主部署实证（2026-10-03）

## 2026-10-05T01:55:31Z 无新部署：遗留owned清理PASS，LO用户环境故障复现

**未完成：报告生成仍BLOCKED，底层根因UNVERIFIED。** 本轮只接续已授权staging精确清理及三环境有界runtime诊断，不业务fix、部署/restart、配置/模板/sharedpermissions变更，不浏览器/新会话/报告生成/评分重算/production/oracle/budget操作。详见[完整清理与LO实证](management-traits-four-real-verification-20261003.md#L3)。

- 首次严格SSH0/vm-ubuntu-go-dev/liming；实际REPORT_EFFECTIVE_ENV/MNG_TEST_REPORT_ENVstaging、APPENVproduction旧配置。online8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93/front98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51不变。当前PID2002/启动01:36:55Z发生在本轮前，不冒称昨日9059仍在；本轮probe前后PID2002同。
- 新删除前完整current备份parent/cleanup_20261005_6jpcfVUo/root0700/files0600，singletransaction/routines/triggers/events/gzip/SHA PASS，DB **8da183355cca1c2eb379ff5827abd0f079bbb42b883364d21cc75b03e7783611**。首次工具停止前GHU38Ugd/009784afa7f8caa83ba185b874881e983db15d6bd92f0386db1344584833ed2e及所有原正式备份保留。
- 原exam1791105048344426522/marker＋createdat/owner/BUNDLE外ref0核验，exactPK＋归属＋逐ROW_COUNT锁内事务清理，01:48:12Zcommit0/SSH0：13dim4mod1receipt1run140qs1snapshot700桶140pq1candidate1paper1profile1bundle1link1exam，11sidecar逐0/owned全0/privatePDF0/revision-current-audit0；SafeUpdate1/FK启用，未DROP/restore/绕profile删除保护。
- 01:54:44ZfinalSQL0：主78/70/1488/134494/295328/1349/27/15FK，source各140140700，3svcactive/healthok。非own排除精确owned的12表before/after、源4表/465旧PDF逐SHA/运行资产均一致；01:55:16Z完整12表051795bd969650d0e55da676083078255331c66405b5fff0ee4ca9d0a7641d0b匹配09:09写前及cleanup后，未删其他正常旧卷，不覆盖旧81aeaa漂移。
- 同纯syntheticDTO/all50 DOCX/SHA dadc9610b59be8deb4eabb497c4a7ec5bc5f2b09de0c81d77d3992e4d2583049/new0700profile/0600input/原LO参数：root **exit0/PDF552631bytes**，普通liming及matching transient **exit1/PDF0**，两stderr195bytes/SHA b6ecc0ff320848795c8f84e1e5b82e2c4867d9e80f4ef0305ebb7287514d3d54相同，固定DeploymentException/terminate/generic application error。root亦有javaldx警告，不能单独当原因；原09:12请求具体exitcode仍未保存，不能回填1。
- 实际HOME/home/liming/LO/usr/bin/libreoffice24.2.7.2/安全PATH/procallowlist复制，User/Groupliming；ProtectHome/PrivateTmp/PrivateDevices/NoNewPrivileges/RestrictNamespacesno、ProtectSystemno、无路径屏蔽。系统LO全部可读、HOME四目录可写；old/new精确窗口无AppArmor/segfault/OOM类命中，coredumpctl unavailable。没有确证HOME/权限/sandbox根因，未修配置/权限/Java/字体。
- 3workspace/上传payload/transientunit最终0，所有syntheticPDF不持久化DB；parent/lo_diag_20261005_8BorrQx7保留受限原stderr与脱敏摘要/实际脚本，最终目录0700/files0600/错误0。新两个C区脚本bash-n0/editor0，cleanup0/probe脚本0（LO0/1/1并非全部转换PASS）；工具失败与owned证据权限纠正完整保留于主报告。
- 下一仅有界只读bootstrap/syscall诊断候选；任何新增执行/业务logger或共享配置、权限、依赖修改须用户确认。本轮**CLEANUPPASS / LO_FAILURE_REPRODUCED / ROOTCAUSE_UNVERIFIED**，不称产品DONE、不重复部署。

## 2026-10-04T09:14:04Z 最小诊断backend DEPLOYEDYES；一卷409 LO退出；finally未清理

**失败/未完成：exact-one清理首次SSHexit255、唯一重试exit255/ssh_timeout，停止。09:14:04.534Z正常admin结果只读仍200/ownedrun存在；最终SQL和复现后旧12表/465PDF比较UNVERIFIED，不能清理0或DONE。LO实际非零退出类别已证实，具体exitcode/原因未查，不修配置或业务。**

- 当前sourcefresh Linux backend build0，在线完整SHA **8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93**；checker freshbuild/vet0，SHA **6d17e1be8d8bbc8b7606e088f5e38f598034dd9c752c94f66c83fa8260954b92**。原backend-only入口真实EXIT0/STAGEverified/noRollback，PID4106排空四项0→PID9059；3fresh positive_app11/159/67/21/cached4/CRUD/205/privateprobe/旧HTTP全PASS，DDL/front/template/config0，production未访问。4项payload核SHA后清0。
- 永久parent同历史；新backend_1455f3ab79235854 root0700/files0600完整当前DB gzip/SHA **6fb3065f92e86df3cda44dbaac2d538dab7e476bdb5fa56041070a5e1f7d139d**；HTTP写前http_1455f3ab7923新fullDB SHA **e163aa312556938e5af0984d3d3cbe6162abac52c503474208d34098206543c3**。都含singletransaction/routines/triggers/events，gzip/SHA通过；原backend元数据/mtime/字节保存，正式旧备份不删。
- 本次**当前**旧12表051795bd969650d0e55da676083078255331c66405b5fff0ee4ca9d0a7641d0b非历史81aeaa；写前70/1488/134494/295328/1349/27、旧PDF465、侧表逐0/private0。部署前后cmp相同；不覆盖既存漂移或删非own新增旧卷。index98547b68…/TEST05c55e77…/contentb0498249…/staging报告环境与原配置保持。
- 同host原正常admin会话getInfo200发布前后/finally，token仅browsermemory。唯一MTH1455f3ab7923/00201candidate，真实freeze140700、140save/readback、completed1run13dim4mod1receipt/raw3final3各140/总体50。仅1生成：09:12:21.096Z—09:12:21.765Z，HTTP409/669.1ms；精确unitjournal09:12:21.742Z **render / lo_command_exec_exit**。SQLrevision/current/audit0/privatePDF0，不重试或运行四560/生命周期/oracle。
- 未清理exactexam **1791105048344426522**，paper **c6d52955-aeec-4abe-8468-c5f350f1df84**，run **ea660428-f2b5-4445-b6e7-9d8b2351e00d**，其余candidate/bundle/FK闭包及内存收窄清理SHA详见[本轮完整报告](management-traits-four-real-verification-20261003.md#L3)。SSH恢复后先接续现有清理授权，不重新部署/索取登录/自签或绕删除保护；深层LO退出原因另需独立有界授权。
- [诊断receipt](../scripts/test/results/MTDIAG-1455f3ab79235854/generate-diagnostic-receipt.json)、[部署receipt](../scripts/test/results/MTDIAG-1455f3ab79235854/deploy-receipt.txt)、[清理重试失败](../scripts/test/results/MTDIAG-1455f3ab79235854/cleanup-retry-receipt.txt)、[浏览器finally](../scripts/test/results/MTDIAG-1455f3ab79235854/browser-final.json)保留。原认证/首页保留，本轮testglobals清空；publichealth200不替SSH最终门禁。历史local6144为前阶段不重复，业务source修改0。

## 2026-10-04T07:54:06Z 不重部署：四HTTP评分PASS，首份generate409，完整验收BLOCKED

**未完成先列：正常admin generate-test首份HTTP409/610.78905ms，根因UNVERIFIED；其余报告/认证PDF/灰环全文/生命周期SKIP。综合oracle历史三轮FAIL与210秒夹具FAIL未修、未改、未重跑，不能FOURPASS。** 新正常用户会话getInfo200/admin/wildcard已真实复用，token仅浏览器内存，最终首页/getInfo200保留，无reset/logout/伪JWT/Redis。

- 在线server534f0abb…/front98547b68…/TEST05c55e77…及两个staging报告环境实际核验；最新positive_app完整gate日志复用，不重新deploy、DDL、配置、前端、模板、build或restart。仅新当前完整备份http_afc22a7e8edf/root0700/files0600/DB SHA8b4863124d7e20b35e9ec0a3b8022805484ee02da9f0aa92fa4bddbed0a593fd，gzip/SHA通过，永久保留。
- MTHafc22a7e8edf两code两正常身份四配置/freeze/140700/固定shuffle/25-20点、560HTTPsave、4completed run52dim16mod4receipt、repeat及mixed双HTTP竞争唯一PASS。原SPECS/Fraction独立SQL核总体50/0/100/565195/11154→50.00/0.00/100.00/50.67/反向只一次/receipt时间一致；仅1candidate真实UI子集＋admin1行13/4详情，不称四全UI。
- 首份报告失败后停止，report/current/audit/privatePDF0，不重试/修业务；没有新PDF SHA/bytes/pages可给。旧pdf_path不写。精确4exam/全部owned依赖按PK/FK事务清理exit0，11侧表各0/private0；当前旧12表81aeaa…/465旧PDF逐SHA/dist/config/unit/servercmp全相等，主78/70/1487/134354/294628/1348/27、15FK、source各140140700，三服务active/healthok/窗口critical0、finalSSHexit0。
- 完整四卷ID/真实失败/各SKIP/工具纠正与安全finally见[本轮报告](management-traits-four-real-verification-20261003.md#L3)；证据[HTTP收据](../scripts/test/results/MTHafc22a7e8edf/http-receipt.json)、[最终只读](../scripts/test/results/MTHafc22a7e8edf/final-readonly-receipt.txt)。仅C区新增精确清理工具和文档，业务/失败oracle/夹具不变。下一独立有界诊断正常HTTP generate409需新授权，不重复部署或索取密码。

## 2026-10-04T06:31:12Z 最新零分backend-only：DEPLOYED_PASS，业务验收仍 PARTIAL

**未完成：正常admin会话缺失getInfo401；完整authenticated HTTP/UI未执行。** 恢复库测试整体exit1/210.08秒后续actualsave失败，offlineWorker/到期完整竞争/剩余生命周期未执行；新综合PDForacle三轮后仍FAIL，不称FOURPASS或144/144。详细每case/工具失败/三轮停止见[最新真实报告](management-traits-four-real-verification-20261003.md#L3)。

- 新在线server完整SHA **534f0abb7f5e90a9eb8da5606763afa0bf15e1fd2994a0418f82cb503e4bf6d5**；fresh构建exit0。仅替换backend，旧abea56…及numeric1000/1000/755/mtime保存到正式backup/backend_31a1dce2c6b90249。当前完整mysqldump gzip SHA **12d2ded728d3ba42aee40807b96ae897c93146ea5fe8c889459ee2bfb5f96646**，root0700/files0600，旧备份/rollback长期保留。
- PID2005 stop→PID/cgroup/8092/appDB0→atomic→fresh应用账号gate→PID4106/post，deployexit0/verified/noRollback；pre/gate/post positive_app完整11/159/67/21/4metadata/CRUD/205规则/源002各140/700ASCII/私有0700探针全PASS。DDL执行0、front/template/config修改0；index98547b68…、TEST05c55e77…/contentb0498249…/全配置SHA不变，不访问production。
- 隔离当前restore mng_lifecycle_test_8a88d9d20f42d4f7服务完成四560save/4run52dim16mod4receipt、4目标LO24.2真实PDF及service读/DB/file三SHA；非HTTP用户鉴权。真实全零PDF **660827bytes/SHA53a6993bc642ac9ebb9befd59f3f22b1d0116e25678e0b1161b7a9a6fb8c5bb8**，9页，五完整E7E6E6环各360/360、whitehole1.0、blue0、五0.00/原字号。zero实际门禁PASS仅限定像素，不覆盖新综合oracle失败或缺失HTTP链。
- 真实零/3作答到期incomplete/正式分NULL/no-render报告/expiredfill拒绝在隔离服务PASS；到期140样本填答失败后failfast。恢复库finallyDROP0/payload0、主业务写0/11表逐0/private0，证据长期归档lifecycle_8a88d9d20f42d4f7。06:31:12Z只读exit0，主78/70/1487/134354/294628/1348/27、15FK、当前旧12表81aeaa…/465PDF逐SHA/frontconfigunit不变，3svcactive/healthok/startupcritical0。
- 原page已不存在，同host正常登录页保留getInfo401，无伪造/重置/注销/令牌文件。下一用户正常login→getInfo200才能HTTP/UI验收；已发布zero不重复发布。需新有界验证器/预算授权由主代理继续，当前不自动关闭todo。

## 最新真实业务续验：DEPLOYED=YES 保持；FOUR=BLOCKED于零分环图（15:12:21Z）

- 原page正常用户登录现真实getInfo200/admin/wildcard；本轮不build/deploy/DDL/restart，在线backend完整SHA仍 **abea56b32f5159211b44f065facfc740ee961b313c4a51b377f915284fc3a2cf**，front/template/config/客户原件/production不改。
- 四组合实际560save、4completed run/52dimension/16module/4receipt、4独立TEST revision/current及服务器LO24.2九页PDF完成；正常两个tester登录通过。独立Fraction、144所需原文段及authenticated HTTP/DB/private/scp bytesSHA全PASS；各总体50/0/100/565195/11154，显示50.00/0.00/100.00/50.67。
- **未满足六图可见**：00201tester全0分报告五环图区空白、0标签和13维常模线仍存在；TEST余量noFill导致0/100正确数值无可见环。严格真实PDF gate exit1，本轮不改样式/数值掩盖失败。paper到期/incomplete、offlineWorker restart及expiry-submit race未验；UI只一个真实participant保存恢复提交和一个admin结果/PDF子集，native下载事件capture timeout明示。
- 新当前baseline root0700=four_20261003_145100_b5bc510880c1（正式backup下），12表81aeaa3b04607b6e6562f30ace2905d91a13cce6cc4f79c421c78616494545d4/465旧PDF逐SHA/configdistdropin前后相同，不拿历史f57或旧dump当当前。最新完整DB fb63c316…及原正式备份永久保留。
- 精确FK/主键finally：首次rangeoptimizer1175事务回滚后，仅当前清理连接SESSION预算64MiB重放exit0，SafeUpdate1/FK保持。4exam/6candidate/2tester/6paper及全部本轮侧表/旧题桶0、4私有PDF按SHA清0；11表逐0/15FK、源各140/700、主78/70/1487/134354/294628/1348/27、三服务active/healthok/窗口critical0。原admin首页/getInfo200保留，本轮secret globals/子页MTstorage清空。
- 完整四组ID/所有PDF SHA/评分/原文/真实视觉失败、8边界证据、工具失败与精确清理见[最新四真实报告](management-traits-four-real-verification-20261003.md#L3)。部署YES只保原状态，不覆盖真实视觉BLOCKED，不需再次部署或用户重新登录。

## 最新backend-only：DEPLOYED=YES，FOUR仍BLOCKED（2026-10-03T14:46:05Z）

**未验先列明：原用户同页真实getInfo发布前后均401，四组合/正常tester登录/560保存/13维4模块/PDF全文与六图/生命周期/UI均未执行。** 保留原cookie/store及首页，不login/logout/reset/创建管理员/自签JWT或Redis会话；不称已解除tester业务失败。

- SSH实际恢复vm-ubuntu-go-dev/liming/exit0。current parser SHA73d72d03…无漂移；fresh Linux后端build0，已生效完整SHA **abea56b32f5159211b44f065facfc740ee961b313c4a51b377f915284fc3a2cf**（49839141bytes）。已有核验器fresh overlay build0/SHAa5c5629098898a67311e74c01beb6bfd0ebfef556880c21a0f0e418bab439646，应用账号fullgate实际PASS，不root替validator。
- 发布前发现既存历史漂移：12表指纹实际81aeaa3b04607b6e6562f30ace2905d91a13cce6cc4f79c421c78616494545d4，非历史f57；旧PDF仍465，共同路径变更0/新增2/缺失2，SHA排序多重集合也不同，原因未查明。**不能声称自13:55起不变或仅重命名**；没有恢复旧dump/旧PDF。
- 新root0700 rollback子目录 **/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/backend_ce78e7e96e76bfd6**：旧backend字节/mtime/原numeric uid-gid-mode1000/1000/755、freshchecker源码/bin、清单/前后基线/receipt和完整当前DB均保留0600。当前全库含routines/triggers/events的gzip12794331bytes/SHA **fb63c316ddea0a299e5e24c98428e530c1bbe0251b4ee7dcf68b1ea413d89d07**，gzip/SHA PASS。原正式三备份SHA均OK永久保留，不还原到main。
- 旧PID27605 stop→PID/cgroup/8092/appDB连接均0→原子只替server→停止时freshgate→start PID30638→postgate，部署exit0/STAGEverified/无需rollback。DDL执行0，front/template/config变更0，不运行完整front部署流程；实际REPORT_EFFECTIVE_ENV及MNG_TEST_REPORT_ENVstaging、APP_ENVproduction维持。
- pre、gate、post三个fresh positive_app核验：11/159/67/21、4metadata、fullcached schema、emptyguard、CRUD、00201/00202各140/700ASCII bad0、TEST205规则/0700private写关删/files0全部PASS。health200、管理三401、合法participant四401/空body400、general+00101/00201/00301MBTI/00401五只读成功；不以此替四组合业务。
- **本轮当前baseline**前后12表摘要/PDF465逐路径SHA/dist+configs+dropins+unit逐SHA/Schema签名全相等；index98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51、TEST05c55e77…/b0498249…保持。主78/70/1487/134354/294628/1348/27、11sidecar逐0/15FK、三服务active/healthok、启动critical0。
- 四确切上传文件核SHA后删除/tmp/mng_backend_scope_ce78e7e96e76bfd6/rmdir remaining0；business实体创建0/private新PDF0，不称四组合cleanup已执行。BACKEND_FINAL_EXIT0、最终LF只读BACKEND_EVIDENCE_CLEAN_FINAL_EXIT0；此前工具syntax1/历史baseline1/末行CRLF127及重新只读终验过程见[最新四真实报告](management-traits-four-real-verification-20261003.md#L3)，不隐藏失败。
- 无业务源码改动/6081或front355重跑/生产访问。下一只需用户在原页正常重新登录，getInfo200后继续已有四组合授权，**不用再部署或再次确认部署**。完整SHA、证据、四组合SKIP表及历史漂移边界见[完整receipt](management-traits-four-real-verification-20261003.md#L3)。

[纠正 - 2026-10-03] 下方BACKEND_DEPLOYEDNO/SSHtimeout为前轮；本轮已YES限定最小部署验收。原admin401与原tester实际业务失败尚无完整四组合复验证据，不称FULLPASS。

## 最新backend-only续作：未发布／BLOCKED（2026-10-03T14:37:18Z）

**BACKEND_DEPLOYED=NO；FOUR=BLOCKED。** 本轮已有staging授权不变，但严格SSH初次＋一次有界重试都连接前timeout/exit255；没有上传/执行远端shell、备份、SQL、stop/drain、安装或重启。用户原admin同页面保留，cookie/store存在且一致，真实getInfo仍HTTP401；未重新认证、logout/reset、自签JWT/Redis会话或保存token。当前在线serverSHA、REPORT_EFFECTIVE_ENV/MNG环境、正式备份SHA/11表15FK/465PDF及三服务状态**本轮均未远端复核**；下方DEPLOYED=YES及0ee9为历史部署，不代表本次parser发布。

- 当前源码fresh Linux后端49839141bytes/SHA256 **abea56b32f5159211b44f065facfc740ee961b313c4a51b377f915284fc3a2cf**，build0；fresh独立checker16442455bytes/SHA256 **a5c5629098898a67311e74c01beb6bfd0ebfef556880c21a0f0e418bab439646**，build/overlayvet0。两产物仅本地ignored bin；没有业务源码改动或把工具门禁放宽。
- 原生nested scope/Gin LoginForm回归42pass/7顶层/0fail0skip/parse0/exit0；Windowsbuild0、vet ./...0且输出0bytes。全Go6081/0/9和review PASS属上一轮，未本轮重跑。
- 本地index98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51、TEST docx05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c、xlsx b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c核SHA不变，无frontend/模板发布或DDL。公网health14:37:18.359Z HTTP200/ok，不替完整gate。
- 四组合本轮全部freeze0/save0/submit0/run0/PDF0；生命周期及真实UI全部SKIP：SSH和正常admin前提双阻断。本次无owned数据/文件/新会话，cleanup_required0而非remote cleanup PASS。保留用户会话/正式备份和已有证据，不继续盲目SSH或写业务数据。

完整命令、SHA、边界及四组合未验表见[本轮真实验收receipt](management-traits-four-real-verification-20261003.md#L3)。恢复SSH及同页有效正常认证后，接续已授权backend-only流程，无需再次批准部署；旧0ee9不能当fresh候选。原实际tester失败保留，尚未由新后端staging实证解除。

## 最新真实续验：BLOCKED（13:55:33Z，管理员前提已解除）

复用用户现有正常admin页面、清route mocks后getInfo200且admin/wildcard均true；不重新登录/输出令牌。四个marked legacy配置及两题本freeze实际成功；00201candidate身份/140题700项创建/重复复用/25-20分钟合同/缺题409通过。**00201tester正常密码后身份HTTP200业务拒绝**，实际guarded runtime audit→revision→run作用域查询报management traits data rejected。按遇业务bug停止，未修业务、未继续00202身份/560答案/交卷/报告/UI/生命周期，不重复deploy/DDL/build/restart。

最终exact事务清理exit0，本次实体残留0、11sidecar逐0/private files0；旧12表f57f35…/465旧PDF逐SHA、3正式备份SHA、server0ee9…/front98547…均不变，15FK/两source280题1400项/主78/70/1487/134354/294628/1348/27，三服务active/healthok。管理员会话保留，不logout。不把部署YES或清理PASS称四组合FULLPASS；失败窗口guard错误1明示。

详细四组合ID/全部SHA/实际拒绝位置/未验表及backend交接见 [真实续验报告](management-traits-four-real-verification-20261003.md#L1)，[脱敏receipt](../scripts/test/results/management-traits-four-real-20261003.json)。下方“缺管理员凭据”是历史事实，现由现有正常登录前提限定纠正；业务阻断仍未解除。

## 真实四组合续验：BLOCKED（2026-10-03T13:20:41Z，缺网页管理员凭据）

**未验先列明**：四组合均未开始；freeze/create/140题保存/submit/13维4模块持久化/真实TEST PDF/浏览器/管理员resume/expiry/revoke/Worker及提交竞争均为 **SKIP：未取得正常管理员登录所需密码的安全来源**。部署YES保持，不代表完整验收PASS。用户补充的SSH私钥已用于真实liming认证，但不能替代网页管理员密码；没有新建或重置管理员、猜旧口令、自签管理员JWT或伪造Redis会话。尚未请求captcha或尝试login，不消耗现有管理员失败次数。

| 组合 | 冻结/创建 | 真实保存/提交 | 13维/4模块/PDF | 判定 |
|---|---|---|---|---|
| 00201 × candidate | 未执行 | 0 / 0 | 未执行 | SKIP：管理员凭据缺失 |
| 00201 × tester | 未执行 | 0 / 0 | 未执行 | SKIP：管理员凭据缺失 |
| 00202 × candidate | 未执行 | 0 / 0 | 未执行 | SKIP：管理员凭据缺失 |
| 00202 × tester | 未执行 | 0 / 0 | 未执行 | SKIP：管理员凭据缺失 |

- 已完整读取项目记忆、本部署报告、完整本地实施报告及实际auth/runtime/report/API合同和所需规则/runtime-validation技能。正常登录必须captchaImage→login(username/password/code/uuid)→getInfo核真实userId1或wildcard，token仅内存；SSH root权限/应用JWT秘密/数据库bcrypt摘要均不是这一登录证明。
- 安全发现仅输出键名/存在性：当前本地进程专用凭据键0；已知本地环境文件只有Word/Graph键，三个应用YAML无管理员/凭据文件引用键。远端真实进程同类键0；仅在/home/liming、/root、应用configs深度2查凭据相关文件名，未找到管理员密码文件，只有非凭据sudo标记及非秘密TEST env。实际应用配置键名复核无独立管理员引用；未全仓检索秘密值/历史记录/备份配置。只读SQL核user_id1启用且未删除计数1，没有读取或输出密码hash/PII。已有auth辅助工具的旧自行签会话方式明确不复用。
- 实际SSH strict known key/BatchMode/ConnectTimeout10/ConnectionAttempts1，hostname/user=vm-ubuntu-go-dev/liming；REPORT_EFFECTIVE_ENV及MNG_TEST_REPORT_ENV均staging，三服务active。server完整SHA0ee9b326c459611babb6951c63291abcdfc27491e8782d3dd9178e4bb56d8483、index完整SHA98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51与已部署版本相等。本阶段不重建/重部署/DDL/重启，启动仍21:10:51 CST；没有重新运行fresh应用账号完整Schema gate，沿用上阶段证据。
- 公网真实HTTP **12项PASS/0fail，执行exit0，13:18:08.379Z**：health200/statusok；freeze/results/resume/generate/view/download六管理缺auth401；create/detail/fill/submit四合法字符串body缺participant/paper token401；独立detail空对象400。只用固定不存在sentinel，未发送令牌，不打印响应体/人员/秘密；这些证据不能替代跨真实owner/purpose/expiry或正向四组合。
- 最终只读SSH **exit0，13:20:41Z**：正式backup root0700且三归档SHA清单通过，DB完整SHA9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1；旧12表摘要f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e与正式基线一致，**465旧PDF逐文件SHA全部不变**；11sidecar逐表0、15RESTRICT FK、私有根liming0700/files0，healthok。无本次owned实体/文件/会话产生，cleanup_required=0，不能称已执行数据清理。
- 首次本地Node脚本因Bash参数展开嵌入JS模板产生syntaxexit1，发生SSH启动前/远端未执行；改用LF stdin原脚本成功，不隐瞒该工具失败。没有业务bug修复、代码/源题/客户文件/V67/V96/旧PDF/生产环境改动；正式备份和此前失败证据保留。
- 下一步所需输入仅为**已有staging网页管理员凭据受限文件的绝对路径＋用户名/密码键名**（不要聊天粘贴密码），由进程内部正常login消费。凭据未补齐前不开始marked写入、不新增管理员，不重复部署授权；后续验收负责人接续本报告§F。

## 当前状态：DEPLOYED=YES（13:13:15Z，仅最小部署验收）

**未验先列出**：四组合真实HTTP/UI/140题提交/PDF、expiry/revoke/resume/Worker竞争仍未执行；没有创建profile、参与者、新paper、run或报告。本轮没有取得真实管理员登录凭据：当前进程的专用管理员环境key为空，不能假称安全凭据已可用。仅部署验收YES，不是完整产品验收YES。一次本地终验Node字符串引号语法错误发生在SSH启动前，未触远端；改为LF直接输入后终验exit0，不因此回滚或改变业务。

### A. 续作实证与边界

- 唯一目标仍为用户明确的20.200.136.133 staging，liming/既有known key/StrictHostKeyChecking=yes/BatchMode=yes/ConnectTimeout10/ConnectionAttempts1；未访问其他production。实际APP_ENV=production仅选择既有配置；REPORT_EFFECTIVE_ENV=staging，新增MNG_TEST_REPORT_ENV=staging。
- 完整读取本报告历史和项目记忆后，读正式失败证据中的checker源码、实际handler/router/JWT/参数binding顺序。participant先严格JSON再真实token验证；管理员JWT/权限先于body。四个合法字符串body无凭据精确401；空body独立400。**未改handler、JWT、接口返回或认证机制。**
- 仅核验工具修正：已有11表pre走完整真实应用账号gate；DDL first/repeat两步记录MAIN_DDL_REUSED_NO_EXECUTION，不执行SQL。回滚独立保存原numeric uid/gid/mode，恢复mtime，不从被chmod0600的证据副本猜mode。
- 业务源111个生产Go文件按有序路径＋字节聚合SHA前后相同：e0d6f6f61efe2b58840edc2352074d6c35ce9166a460b3dfed2a42306faff644。001 SQL不变；无旧表ALTER/DML/backfill、无旧题本改写；新表每表0保留。不制造capture主表或跳过认证。
- 本轮不重跑6039或前端355全量，不重建已验证server/front：可用server、index、dist归档和两个TEST资产完整SHA与下方前次fresh清单完全相同，经上传清单/生效文件/公网复核。只有更改过的工具checker重新编译/vet。

### B. RED→GREEN及实际执行文件

- [工具合同测试](../scripts/test/management-traits-deployment-contract-test.js#L1)：有效RED三失败/exit1→GREEN三通过/exit0（合法body401与空body400、既存11表、独立元数据恢复）。
- [实际注册路由测试夹具](../scripts/test/fixtures/management-traits-deployment-http_test.go.txt#L1)：使用真实RegisterRoutes/JWT/handler、nil DB，7子测试＋1顶层PASS，缺凭据不进入业务DB；临时副本go test及go vet exit0。不是mock HTTP返回值。
- 提取实际部署脚本回滚片段，在独立owned临时目录中原0755文件→证据副本0600→替换→恢复，原内容、numeric uid/gid/mode、mtime完全一致，ROLLBACK_METADATA_BEHAVIOR_PASS/exit0；临时目录精确清理。
- [部署工具](../scripts/db/management-traits-staging-deploy.sh#L1) SHA2f62096ca0fb563ff655484f9beb83afb1f48525fe1b04bea2eee77d2777126e，bash-n exit0；[永久核验器源码](../scripts/tools/management-traits-staging-check.go.txt#L1) SHA5d41d66092b57852154b88c0bf1b56cc4070524a47b9480fc1957579afa25eca。
- fresh Linux checker16431599bytes/SHAe6bbf1890cc40c76c0916beca7bf0ec6d274adc095e24a5175e1f29e4e96b0eb，build/vet exit0，零error。工具测试SHA942614f5071f2e496a559690b693064e5a03670f77818811aa567862f69d5935；Go夹具SHA5be18fcb1fd61b74825377d5e35a1c4e7b0c0f799189af8fc860e4fd2bea9a6a。

### C. guard/drain与真实应用账号gate

- 旧PID24732→stop，PID消失/cgroup空/8092=0/positive_app连接0；原子安装现有fresh server，TEST仍disabled，guard PID27481完整gate/health通过。再次stop该PID并核同四项归零，之后只重复只读签名和gate，没有重DDL。
- pre旧应用、pre新guard、两gate、post共5个fresh checker实例均使用actual config.Load/positive_app/真实JWT配置内部消费；CheckRuntimeSchema首次＋缓存PASS、每实例metadata11/159/67/21共4queries，空AllLegacy=false/error=nil、实际CRUD权限及002各140/700/ASCII bad0。没有root validator替应用账号。
- 三份first/repeat/final完整结构签名相等，SHA3fb28c5c220c063795e77b3e27b8eed8ace01b0dc7ee82614a556af8f1c3c0d1；11表15RESTRICT FK不变，DDL执行次数0。
- 最终PID27605，ActiveEnterTimestamp=2026-10-03 21:10:51 CST（13:10:51Z），User/Group=liming、WorkingDirectory=/opt/talent-assessment；talent-assessment/nginx/mysql全部active。

### D. 已生效资产与配置

| 生效对象 | 完整SHA256 | 权限/位置 |
|---|---|---|
| server | 0ee9b326c459611babb6951c63291abcdfc27491e8782d3dd9178e4bb56d8483 | /opt/talent-assessment/server；49828320bytes |
| index | 98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51 | /opt/talent-assessment/dist/index.html；公网同SHA |
| TEST docx | 05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c | 原§4绝对独立目标；liming:liming0600 |
| TEST内容xlsx | b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c | 原§4绝对独立目标；liming:liming0600；实际loader205规则 |
| 非秘密TEST env | 7c56d3ab4cab53633b0b4e52628be69f7af7e29f208707023a75ecde8e2cb082 | /opt/talent-assessment/configs/management-traits-staging-test.env；root:liming0640 |
| 新唯一dropin | 96dff874996f27cf15335e84cf0aa6cd3aeea23dadd2241a04d73d55d5918367 | /etc/systemd/system/talent-assessment.service.d/management-traits-test.conf；root0644 |

新私有根=/opt/talent-assessment/private/management-traits-test-reports，liming0700、不在Nginx alias；应用用户实写/关/删0600探针PASS、files0。四MNG_TEST key已从真实进程读到原§4值。原共享configs/application/旧模板tar-compare exit0；未输出秘密。front393文件逐文件SHA全部匹配本地fresh dist，无symlink，root:root0755，nginx-t PASS；仅静态与API验收，不称浏览器四组合已验。

### E. HTTP、SQL和保留证据

- 内部9合同：health200；管理results/freeze/download401；四参与者create/detail/fill/submit完整合法JSON缺token401；独立paper-detail空对象400。公网Nginx四参与者同401、空body400、health200/statusok、index200/SHA同表。
- general online-paging及既有00101/00201/00301 MBTI/00401五只读调用均HTTP200且业务code0或200；只打印类型/状态，不打印既有人员/IDs/响应体。没有生成报告、更新答题或建立会话。
- final SQL table/exam/paper/paper_qu/paper_qu_answer/candidate/tester=78/70/1487/134354/294628/1348/27；11sidecar逐表0。旧12表指纹仍f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e；**465PDF逐SHA完全不变**；3正式归档SHA仍全部OK。指定启动窗口journal关键错误0。
- 正式备份沿用原§2全部路径/SHA，永久保留。新证据=/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/deploy_24bde3635e559ab7（root0700/文件0600）：实际checker源码/二进制、deploy-executed.sh、runtime-not-reapplied.sql、payload清单、pre/guard/gate/post日志、三schema签名、server.metadata.before/server.before、原unit/dropins/configs、deployed-SHA256SUMS、frontend393清单、sidecars/counts/PDF、shared-config-compare、cleanup/final receipt均保留。原deploy_3461565581060643失败证据未删除。
- 即时回滚dist=/opt/talent-assessment/dist.mng-rollback-24bde3635e559ab7保留，旧server.before另有独立元数据及原mtime；原owner/mode实读root:root0755。回滚需stop/drain、sidecar为空、按保存原元数据恢复旧bin/front并移除仅新配置；**有新数据则停止评审，禁止DROP11表或全库restore**。
- 本次/tmp/mng_deploy_staging_24bde3635e559ab7九确切文件核SHA后清理0；本地仅两编译副本和其空owned目录清理0，fresh bin保留。没有凭据/token文件、Git提交/push。最终远端时间13:12:40Z exit0；公网/393文件/业务源guard终验13:13:15.810Z exit0。

### F. 下一验收负责人合同（本阶段不执行）

1. 不重新部署/DDL。先读本节、实际[API客户端](../Go-based%20Refactored%20System/ruoyi-ui/src/api/managementTraits.js)、[TEST表单](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L485)，核上述SHA、staging环境、11空表/465PDF及fresh实例完整gate；保持owner/private目录边界。
2. 管理员必须实际GET /prod-api/captchaImage→POST /prod-api/login，JSON严格使用username/password/code/uuid，再GET /getInfo核userId1或*:*:*；token只在内存Authorization: Bearer。安全来源：用户认可的私有secret/env或现有安全登录会话，由测试进程内部消费、不打印/不落盘。本轮未核定可用管理员明文凭据，缺失时仅补一次安全输入，不重复部署确认；**不要使用旧脚本硬编码口令、DB密码hash当登录密码、自己签JWT或伪造Redis管理员会话**。应用DB/JWT仍可在远端实际config.Load/进程env内部消费，无需输出或下载config。
3. 唯一marked临时数据，四组合=00201×candidate/tester、00202×candidate/tester。通过真实管理UI/既有POST /exam/api/exam/exam/save创建四个新legacy+legacy配置，joinType1、单物理题库140radio/其余0、totalTime25、state0、正确isOpen（candidate1/tester0）、非空合法requiredFields及showPdf=false；跟随当前表单/DTO，不套猜测状态。冻结前paper0，closed tester只建立本次owned账号，password在实际保存响应或本次owned DB记录内存消费，不读取/输出其他人凭据。
4. POST /exam/api/management-traits/profile/freeze严格只有examId，后台真实auth；GET /profile/detail?examId。冻结使用[独占租约](../Go-based%20Refactored%20System/internal/service/management_traits_runtime.go#L44)排空同进程旧写；[事务前提](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_profile.go#L59)要求单repo/140/25及无paper。核源原文和140/700/manifest/mapping，而非仅SHA字符串。capture不是注册HTTP路由：原owned capture证明在[恢复库测试](../Go-based%20Refactored%20System/internal/service/management_traits_guard_staging_test.go#L167)，只能在明确owned恢复库重放；**不要在主element造capture表、历史重算或批量捕获465PDF**。
5. candidate POST /exam/api/candidate/save只examId＋已配置九字段子集（name/gender/telephone/affiliation/post/age/degree/major/stuFlag），返回身份五键；tester POST /exam/api/tester/login只examId/idNumber/password，走真实密码验证，身份token由服务器颁发。恢复candidate必须原有效token或管理员POST /management-traits/admin/candidate/resume（examId/participantId/paperId），不可按手机号匿名恢复。
6. 下列均加公网/prod-api，POST body严格字符串且禁止未知/重复字段：/management-traits/participant/create-paper只examId＋participant-purpose token；paper-detail只paperId＋paper-purpose；fill-answer只paperId/paperQuestionId/optionId；submit只paperId/submitType=manual。token放X-Management-Traits-Token，不放URL；以真实140题返回id及options操作，保存560次，重复create/detail同paper/固定题序，未完整manual409、140全答submit/重复幂等，核冻结25分钟/20分钟提醒。
7. 管理POST /management-traits/results/list只examId；GET /results/detail?runId；POST /reports/generate-test只runId；GET /reports/view或download只reportId。核run13维/4模块/receipt、report revision/current/audit复合绑定，PDF认证字节/SHA/size/TEST用途/客户205原词/六图/实际目标LO与逐页视觉。禁止客户端路径、正式激活、旧pdf_path替换。四组合及expiry/revoke/短续答/Worker真实负例由下一负责人单独执行并登记，不以本轮401代表这些通过。
8. 所有创建ID维护marked ownership；finally仅按本次主键和真实FK依赖事务清理（新profile保护会拒绝旧通用删除，不假称普通exam delete能清），只删除本次私有PDF并核SHA/允许目录；共享bundle引用未归零不能删除。11表恢复0、临时旧关联0、465旧PDF/源题与旧摘要不变、退出真实会话；任何业务失败停止，不修业务适配错误probe。正式备份和本部署证据始终保留。

---

## 历史失败记录（13:01:44Z，保留，不代表当前在线版本）

## 最终状态：DEPLOYED=NO；主 DDL 已安装，应用已回滚

**失败／未验先列出**：部署后只读核验器向 participant/paper-detail 发送空对象，错误期待 HTTP401；实际生产路由先校验 paperId，返回 **HTTP400**，部署命令 exit1、阶段 restart-postcheck。这是核验请求与既有合同不一致，不是已证实业务缺陷。遵守用户“任何 error 停止并恢复”的要求，未修业务逻辑、未修改失败断言、未重新部署。新前端浏览器路由、旧链四类型 HTTP 读取、两题本×两身份完整 HTTP/UI/PDF/SQL、expiry/revoke/resume/Worker 验收均**未执行**；不称完整 TEST 链通过。

最终只读终验 **2026-10-03T13:01:44Z / exit0**。旧后端、旧 dist、共享配置和旧模板精确恢复；新增 TEST 环境、两个独立测试资产、空私有目录及上传/失败新 dist 全部 exact 清理。主 element 的 **11 张新表／15 个 RESTRICT FK 保留，每表行数0**，不自动 DROP、不回填、不改变旧表。其他 production 未访问。

## 1. 授权、现场及 fresh 本地门禁

- 唯一目标 20.200.136.133；真实 hostname/user=vm-ubuntu-go-dev/liming。SSH 使用已有密钥、StrictHostKeyChecking=yes、BatchMode=yes、ConnectTimeout10、ConnectionAttempts1；没有调整网络、host key、账号或 grants。
- 实际报告环境 REPORT_EFFECTIVE_ENV=staging；APP_ENV=production 是既有配置选择，不改变基础/production 覆盖配置。unit 实际 User/Group=liming、WorkingDirectory=/opt/talent-assessment、ExecStart=/opt/talent-assessment/server、KillMode=control-group、TimeoutStop=90s；后端真实监听8092。
- 原3个 dropin 为 cjk.conf、phase1-word.conf、report-effective-env.conf，未覆盖。Nginx root=/opt/talent-assessment/dist，API代理127.0.0.1:8092，唯一观察到的上传 alias=/data/uploadPath/profile/。新报告根不在静态 alias 下。
- LibreOffice **24.2.7.2 420(Build:2)**，fc-match 为 Microsoft YaHei/微软雅黑；磁盘61G、已用16G、可用46G。未更换 LO/字体或修改共享上传目录权限。
- 当前源码 fresh Linux build、临时 checker build/vet 均exit0；Go全量 **6039pass／638通过顶层／0fail／9skip／parse0／exit0**，vet ./... exit0。计数含子事件，非覆盖率。
- fresh 前端 **31files／355pass／0fail**；build:prod exit0，资源体积/Browserslist和已有测试方法替换提示保留，不更新依赖。新 dist **393个文件**；下列 SHA 恰与此前候选相同，但本次重新构建，不复用旧产物作为证明。

| 本次 fresh 候选 | 字节数 | 完整 SHA256 |
|---|---:|---|
| [Linux server](../Go-based%20Refactored%20System/bin/server-mng-staging-linux) | 49828320 | 0ee9b326c459611babb6951c63291abcdfc27491e8782d3dd9178e4bb56d8483 |
| [fresh index](../Go-based%20Refactored%20System/ruoyi-ui/dist/index.html) | 16155 | 98547b680cd977a4d78e77c74d0a7395aa4f2ae967c2900550106170e11dff51 |
| [fresh dist 归档](../Go-based%20Refactored%20System/bin/mng-staging-dist.tar.gz) | 7627825 | e0b37707e750510d8b0789b3f7560668ff59d39cef100ee18a27c1dfa65c101e |
| 临时 Linux checker（已清理，源码在远端证据保留） | 16431823 | 3d25ac5dc45d7f2863a4bfa3026822e116deb9bfc419ad19796ca81e17852490 |
| checker 源码（原字节归档，无凭据） | — | c525f8ffcc9d0262b347b8607fac0de8cda50e10daba7f3e51617d69a92fb14f |
| 当前 guard 源码 | 20101 | 5b83a097bad5140b6d494f337212a4cc7414008d432c34ab2593c562c184194c |

9个本地环境skip原样：TestBugFB185I_ConcurrentRecomputeCreatesOneCompleteRun；TestPhase1WordTemplateCandidateUploadContract；TestBugFB170_Phase1GroupPieLabelsStayOutsideChart；TestPhase1CustomerWordTemplateLibreOfficeProducesExpectedPages；TestBugFB169_Phase1ChartGenerationPreservesTemplateLabelStyles/uploaded_staging_template；TestBugManagementTraitsPDFVisibleTestLabel；TestManagementTraitsGuardAuditMySQLStagingExternal；TestManagementTraitsGuardAuditMySQLExternal；TestManagementTraitsSourceLockMySQLExternalUpdate。前阶段恢复库非skip实证不折算成本次默认0skip。

## 2. 正式备份与不可删除证据

沿用正式备份 **/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9**（root0700、私有文件0600），部署前后3个 SHA均OK，gzip全部通过；部署前 application/files-and-system tar --compare均exit0。

| 正式归档 | 完整 SHA256 |
|---|---|
| element.sql.gz（12793472 bytes） | 9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1 |
| application.tar.gz | 079a718a96d1fd43d0fcf037d3796362684d2211400d975258f0d929c90d5fbd |
| files-and-system.tar.gz | 21fa940fd5800f70e33c047c42e13148ddeeebe8a1a2240c466c765a81ff998d |

本次永久证据 **/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/deploy_3461565581060643**：
- deploy-receipt.txt：DEPLOYED=0 EXIT=1 STAGE=restart-postcheck ROLLBACK=restored-previous-app-schema-preserved DDL_PRESERVED=1；原失败保留，不被 final exit0覆盖。
- precheck.log、guard-running-disabled.log、drain.log、ddl-first/repeat.log、gate-first/repeat.log、schema-first/repeat.signature、postcheck.log、failure-http-status.log。
- server.before、unit.before、dropins.before.tar.gz、configs.before.tar.gz：只保存在该主机 root私有目录，未下载任何含配置秘密的备份。
- checker-source.go.txt、deploy-executed.sh、runtime-applied.sql、payload-SHA256SUMS、frontend-candidate.sha256、frontend-failed-candidate.tar.gz、test-template.failed-candidate.docx、test-content.failed-candidate.xlsx：保留本次实际失败候选与执行源码，不留运行激活。
- cleanup-receipt.txt、rollback.txt、rollback-owner-mode-final.txt、application-final.log、legacy-after.sha256、pdf-after.sha256、counts-final.txt、sidecars-final.txt、final-env-keys.txt、final-receipt.txt：终验/回滚资料。

## 3. 两阶段实际 guard/drain → 主 DDL

1. 使用实际配置加载器、运行进程环境和真实应用账号 positive_app，核 DB为element/local3306、JWT秘密非空（不输出）、11表未安装、四MNG_TEST key未设置。旧 PID2001。
2. systemctl stop 旧后端，**PID不存在、control-group空、8092监听0、positive_app MySQL连接0**。非仅“发SIGTERM”或等待固定秒数。
3. 原子替换为当前 fresh 新后端并启动 **PID24499**；TEST环境仍未配置、main mng0，应用账号预检PASS、health正常。没有创建profile、冻结题本或问卷写入。
4. 再次停止新守卫 PID24499，同样 **PID不存在／cgroup空／8092=0／应用DB连接0**。主 DDL 在单后端停止且两轮旧/新在途连接均归零后执行。
5. 原 [001 SQL](../scripts/sql/management_traits_001_runtime.sql) SHA **7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8** 不变，使用 root既有socket执行，仅 CREATE11新表／15FK；无ALTER共享旧表、DML/backfill或改原源题。
6. 首次、重复均成功；两份完整列/type/NULL/charset/collation/PK/索引前缀/FK逐列序目标与动作签名相等，SHA **3fb28c5c220c063795e77b3e27b8eed8ace01b0dc7ee82614a556af8f1c3c0d1**，与前阶段恢复库相同。
7. **首次、重复、启动后 post 三个 fresh 实际 RuntimeService**使用 positive_app调用生产 CheckRuntimeSchema两次均PASS；同实例metadata查询4次、实际GORM行数 **11/159/67/21**。不是人工count替代validator，也不是root account gate。
8. 完整canonical空安装 AllLegacy scope返回 protected=false/error=nil；应用账号 SELECT/INSERT/UPDATE/DELETE权限通过 SHOW GRANTS仅内存检查，不打印grants/secret、不更改授权。00201/00202均真实只读140questions/700options/ASCII bad0；不写旧源、不重新冻结或创建测试测评。

## 4. TEST 资产与配置（曾安装，最终已移除）

只新增以下4个非秘密运行 key，保留所有既有环境值：
- MNG_TEST_REPORT_ENV=staging
- MNG_TEST_REPORT_DIR=/opt/talent-assessment/private/management-traits-test-reports
- MNG_TEST_TEMPLATE_PATH=/opt/talent-assessment/configs/export-templates/management-traits-002-test-only-v2.docx
- MNG_TEST_CONTENT_PATH=/opt/talent-assessment/configs/export-templates/management-traits-002-test-content-v1.xlsx

非秘密 EnvironmentFile=/opt/talent-assessment/configs/management-traits-staging-test.env（root:liming0640）；唯一新增 dropin=/etc/systemd/system/talent-assessment.service.d/management-traits-test.conf（root0644）。均在回滚时移除、daemon-reload；最终运行进程四key全部未设置。**没有最终生效的新版TEST配置 SHA可报告**。

| 曾安装独立目标（旧模板未覆盖） | 完整 SHA256 | 应用账号实证 |
|---|---|---|
| management-traits-002-test-only-v2.docx | 05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c | liming实际读取、SHA相等；530193bytes |
| management-traits-002-test-content-v1.xlsx | b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c | liming实际读取、生产内容loader=205规则；51421bytes |

私有新根 liming0700，新文件探针0600，由liming真实创建/写入/关闭/删除后files0；不使用共享upload/tmp作为PDF根。未生成任何PDF。两个资产核SHA后转存正式证据，再精确删除运行副本，空private子目录/本次新父目录rmdir；最终这些运行路径均不存在。

fresh dist在新目录展开，root:root0755后原子切换；nginx -t通过。HTTP失败后旧dist原子恢复。失败候选393文件逐文件SHA与本地fresh dist一致、无symlink；归档到私有证据后按确切清单删除，再rmdir所有空目录，不rm-rf/删未知文件。新dist和旧回滚临时目录均0。

## 5. 实际 HTTP 失败及停止

新后端 post checker实际依序完成：

| 请求（无登录token，不写测评） | 实际结果 |
|---|---|
| GET /health | HTTP200 |
| POST /exam/api/management-traits/results/list | HTTP401 |
| POST /exam/api/management-traits/profile/freeze | HTTP401 |
| GET /exam/api/management-traits/reports/download?reportId=deployment-invalid | HTTP401 |
| POST /exam/api/management-traits/participant/paper-detail，body={} | **HTTP400；checker期待401，exit1** |

实际 journal **2026/10/03 20:56:48**：`400 ... POST /exam/api/management-traits/participant/paper-detail`。仅精确筛选这条假请求，不输出其他人的路径、JWT或PII。

[PaperDetail](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime.go#L123-L130)在token校验前要求合法且非空paperId；核验器使用 `{}`，没有进入缺token401分支。没有证据证明业务不安全；**不能放宽真实handler或简单把期待401改为400来宣称鉴权已验**。下一次核验应分开“空对象400”和“合法明确paperId、缺token401”；本轮按error-stop未做此改动或重试。

后续 general/001/002/00401/MBTI真实HTTP只读检查位于该失败断言之后，**未执行**。新前端仅完成上传SHA/393文件/Nginx配置检查，**未执行浏览器路由或公网fresh index验收**。四真实测评组合和报告阶段未开始，没有测评临时数据或短时登录会话。

## 6. 回滚、必要元数据恢复及最终实证

- 失败 trap停止新后端；确认每个sidecar为空后移除新env/dropin、恢复旧server和旧dist、重启旧应用。**11表不删除**，没有全库restore或旧数据修改。
- 首次自动回滚把server owner设为liming，与原正式归档root/root不符。只读tarcompare发现Uid/Gid差异；使用私有server.before复制件作为mode参考又取到0600（因证据目录防泄漏策略），只读终验再次失败，**这些失败未隐藏**。
- 直接读取未变更正式 application归档的server条目，实证 **-rwxr-xr-x root/root 48854107**；随后仅恢复运行旧server的 **root:root0755**与原时间。没有再次发布／改字节／重启新代码。最终 application归档 tar --compare **exit0**，不是忽略mode/content差异。
- 初次只读终验 Node模板字符串引号语法exit1发生于SSH前，远端未启动；后改用LF脚本输入完成终验。未修改生产或核验器源码以隐藏失败。
- 最终旧backend **/opt/talent-assessment/server** SHA **ee4566e7a698ff592acaeab40c5986974787be27929cd1c2bc002d1adddcd300**，root:root0755；旧index **/opt/talent-assessment/dist/index.html** SHA **2d4ba6c5ac1e745d736f8c2894cb99406d4890123bb18fb0c0d8bb001c8e9bc9**。PID24732运行旧应用，原3dropin恢复。
- 旧12表指纹 **f57f35a7af3dd3efd2acea6f27c51f7146fed610176e7fb93ba94588cf86d53e** 前后相等。最终table/exam/paper/paper_qu/paper_qu_answer/candidate/tester=**78/70/1487/134354/294628/1348/27**，只有table数因新11表从67→78；旧表行数/数据不变。
- 11表分别0行：definition_bundle、exam_profile、paper_question_snapshot、paper_snapshot、report_audit、report_current、report_revision、result_dimension、result_module、result_run、runtime_receipt；FK15均RESTRICT，mainDDL保留。
- **465旧PDF逐文件SHA清单完全一致**；3正式归档SHA仍OK，旧server/dist/configs/所有模板与application归档整体验证exit0。
- talent-assessment/nginx/mysql全部active，内部health **{"status":"ok"}**；最后UTC13:01:44Z，POST_ROLLBACK_FINAL_VERIFY_EXIT0。公网旧index/health附加验证见后续收口。
- exact上传 /tmp/mng_deploy_staging_3461565581060643=0，失败新dist目录=0，两新TEST运行资产=0，新env/dropin=0，私有probe/files=0；正式备份/失败候选归档全部保留。无credential/token落盘、git提交或push；无production操作。

## 7. 改动及下一步（不自动执行）

- 新增 [部署脚本](../scripts/db/management-traits-staging-deploy.sh#L1)，归属scripts/db；业务源、API、001 SQL、客户原件和其他脏文件无修改。临时核验器用apply_patch建立，build/vet后实际上传运行；已归档精确源码供审计，按本次授权清理，不转为新业务功能。
- 脚本保留本次失败请求和回滚历史，不冒称脚本已完整PASS。后续须先在新有界切片修核验请求、按正式archive保存/恢复server metadata，而非从被统一chmod0600的私有副本猜原mode。
- 重新部署不得假定main mng0：当前是完整 **11空表／15FK**，应first/repeat no-op并fresh应用账号完整gate，不DROP再建、不可回填。
- 下一部署负责人从现有保留备份和本receipt续作，复核source/fresh资产→guard/drain→已存在主DDL完整门禁→重新独立TEST配置/front→正确最小HTTP合同。通过后再交完整验收负责人执行四组合/真实PDF/expiry/revoke/resume/Worker与marked cleanup。当前不是“已部署YES”，也不是完整产品验收PASS。

### 最后本地／公网收口（已执行）

- apply_patch Delete返回后磁盘仍有临时核验源码；再次按已归档源码SHA c525f8ff…及Linux checker SHA3d25ac5d…核确切所有权后，仅删除这两个文件及空临时目录，**OWNED_LOCAL_CHECKER_TEMP_REMAINING=0**。没有修改源码内容、删除其他bin/tmp或凭据文件。
- 真实公网 GET /prod-api/health=**HTTP200，status=ok**；GET /index.html=**HTTP200，SHA256 2d4ba6c5ac1e745d736f8c2894cb99406d4890123bb18fb0c0d8bb001c8e9bc9**。证明公网仍为回滚旧版本，不是fresh前端上线或浏览器新route验收。
- 新部署脚本及3个文档 diagnostics0，scoped git diff --check exit0；没有进一步业务编译/部署重试，不把文档GREEN改写成DEPLOYED=YES。