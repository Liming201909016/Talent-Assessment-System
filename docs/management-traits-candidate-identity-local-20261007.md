# MT-CANDIDATE-FIELDS — 本地身份登记修复

## 2026-10-08T01:43Z 最小 staging 发布与单一合成身份验收完成

**未替真人保存、未发布生产、未安装草稿/正式审批/重发功能；本轮仅最小身份切片 staging PASS。** 用户实际信息与原页面未保存输入不读不写；新独立公开页只读检查。旧失败/NOT_EXECUTED为历史，未覆盖；用户交接的独立 CodeReviewer 最终 source＋artifact PASS/ALLOW 为本次发布依据，未重复审阅、模块证明或构建。

- **发行与安全**：strict SSH首次成功，fresh PID2006/backend8baa/index353c/三服务healthy；全产品state1与待结算competency均0，切换前再次fresh0。仅隔离server-candidate与candidate-ui/dist-faithful，49903310bytes及393文件逐大小/SHA一致；原工作树未打包、业务源0改、build0。正常主service stop/start一次，PID2006→2746；nginx/mysql PID保持、无restart/reload。后端先通过真实Detail同ID/严格true/三字段，再前端原子exchange/rootroot755；rollback实际0，不承诺零停机。
- **精确旧→新SHA**：后端8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676→**4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c**，/proc/2746/exe同；index353c9fb3a0f30540c8ca091885b67bf458249c1023eca909f78c83988bf27de2→**52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860**。公网393每项HTTP200/大小/SHA精确匹配candidate manifest；index按manifest真实值核验，不用旧硬编码。
- **备份/回滚**：服务器受限目录 `/opt/talent-assessment/backups/mng_identity_5d517cc02d6c8396` root0700/files0600，当前全库singletransaction/routines/triggers/events gzip SHA **e84f74c07e6093ac224f1d4d982dbeb3e6836b4c4f48bf895b52d9ef5c6abc15**；application归档809bfa0d2e6be0cfa0d37ef08532f59fe4628bb6fc58bfd81fd7b239b9cf5f42、system归档ee6b1a3d0bdb7acb9c3d436867c85dff260818939acc1a258ed17ed3d37275bd。gzip/tarcompare/manifest SHA及原binary numeric metadata实际校验；即时旧dist目录保留，未下载secret备份。exact上传两文件和owned tmp目录最终0，正式备份保留。
- **用户页只读**：新public tab普通/my URL无mngTest，公开POST Detail200/code0/同1791298091700970647/managementTraitsProfileFrozen严格true/requiredFields=name,gender,telephone；DOM仅姓名、手机号、性别，保存按钮disabled=false/loading=false。不去填真人/保存/刷新原用户页，不去掉合法gender。旧13keys/8extra已由新链替代，后端strict白名单不放宽。
- **真实合成保存**：正常独立Chromium管理员getInfo200/code200/admintrue，正常UI仅1新00201开放TEST配置及明确两步冻结，配置三字段。普通考生UI实际填写合成身份并保存 **HTTP200/code0**，body四键examId＋name/gender/telephone，extra0；原五键身份响应，token只浏览器内存，不留response原文或storageState。SQL仅布尔/字段名/哈希：新candidate恰1、身份匹配1、非配置字段空1、paper/pdf空1；冻结contract字段三项一致。没有startpaper、paper snapshot、倒计时、作答或报告，不能把profile合同称paper身份快照。
- **精确清理/保护**：唯一新exam1791423812678221128/candidate0cd04eae-9011-4e9f-9486-eca044a1dd36，FK开启/SafeUpdate1/主键条件清理native0；owned exam/candidate/profile/paper逐0。复用既有bundle24603e07-b74e-4348-9560-0e06f786c987，保留其原用户引用不删除；主计数恢复exam71/paper1488/candidate1349/profile1/bundle1，而非全库11表归零。合成driver原terminal实采native exit0，contexts关闭。最终SSH0/三healthy/新PID与两SHA保持，旧465逐路径PDF/私有文件基线/源各140-140-700/cache/config/metadata/Schema及原用户整profile SHA均前后相同；其field_contract SHA仍051dbe83d91a184e273ae6ce72748952c44cad4e4163ec82dcf6bf3c83aa2b6b，11表15FK、draft表0。
- **工具失败保留**：首次只读inline模板syntax1在SSH前，纠正同命令0；首次pipe登录确认未送达Node、业务写0，关闭exact owned进程后改direct-file窗口，用户正常登录并真实验收0；成功publish收据保留stderr90bytes/SHA，未保留rawstderr/未推断cause。另一同步terminal退出码为空不是native0，最终回原terminal明确IDENTITY_SYNTHETIC_NATIVE_EXIT=0。没有业务试改或新包构建。

最终文档核验另有一次native1：把完整build收据文件SHA误与manifest数组SHA比较；只读格式探针实际compact数组SHA精确为1649a1c7ca4e166d290daa3a8e934d4704a3c1937e066f4fd17b8515c733974f，纠正验证口径后native0/新增链接均有效/两个脚本syntax0/owned测试Node0。既有manifest与产物没有修改，不是发布包漂移或重新打开模块门禁。

状态：source/artifact沿最终独立PASS；publication **completed**；single synthetic identity acceptance/SQL/cleanup **completed**。只关闭此staging身份故障，不声称客户自己的信息已保存，不声称完整禁止新建旧版002或完整正式产品完成。用户可在原页面刷新加载新资源后自行重新保存。

证据：[最终验收收据](../scripts/test/results/mng-identity-publish-20261008/acceptance-verdict.json)、[发布](../scripts/test/results/mng-identity-publish-20261008/deployment-verdict.json)、[公网393＋Detail](../scripts/test/results/mng-identity-publish-20261008/public-verdict.json)、[真实UI HTTP](../scripts/test/results/mng-identity-publish-20261008/synthetic-http.json)、[SQL](../scripts/test/results/mng-identity-publish-20261008/synthetic-sql-verdict.json)、[清理闭环](../scripts/test/results/mng-identity-publish-20261008/synthetic-summary.json)、[远端终验](../scripts/test/results/mng-identity-publish-20261008/final-1791423821973-1.json)。新C区[发行驱动](../scripts/tools/mng-identity-publish-20261008.js)、[单一合成测试](../scripts/test/management-traits-identity-save-staging-20261008.js)，Node语法/编辑器diagnostics0；未修改业务/旧harness/旧收据/用户Git。

## 2026-10-07T15:56Z 最小 staging 发布执行：SSH 双超时阻断，未启动远端命令

**发布 BLOCKED_SSH；真实合成保存未验证。** 首次及唯一重试均连接前超时，native exit255/255，10030/10028ms，stdout0；父进程 native exit1。没有远端预检结果，不把最后 PID2006/backend8baa/front353c 或旧健康与活跃计数当本轮 fresh 事实。没有创建服务器备份、上传、替换、重启或任何 SQL 写入；restartCount=0、DDL/DML/production=0。原用户页面只读快照，未刷新、改输入或保存。

- [限定纠正 - 2026-10-07] 下方15:07及14:50的资源门禁阻断/独立复审未执行是当时历史。用户最新主上下文明确交接：独立 CodeReviewer 最终 **PASS ready_source_and_artifact / ALLOW minimalstagingpublish**，涵盖两源码及补证闭包与实际393产物。本worker依据该独立审阅交接继续发布流程，不自称重新执行review、不重新阻断raw393/old21，也不覆盖旧NOT_EXECUTED收据。已有最小staging发布与0活跃时一次后端重启批准保留。
- 本轮仅原生核已批准不可变产物：隔离server-candidate 49903310bytes，SHA **4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c**；candidate-ui/dist-faithful实际393文件逐大小/SHA精确匹配既有manifest，额外文件0，native exit0。manifest SHA **1649a1c7ca4e166d290daa3a8e934d4704a3c1937e066f4fd17b8515c733974f**，index SHA **52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860**。未重建/打包当前工作树，未夹带draft/formal/reissue。
- 固定liming/既有USERPROFILE密钥、Strict=yes/Batch=yes/Conn10/Attempts1，最多首次+1retry，两个stderr各67bytes/SHA dbc4779f6c9efb426d2ec093088dd86963fbccd91ec0d5dbd98b3a5e08743a72。两次上限后停止，没有HTTP绕行造数或第三次SSH。
- 合成实体0、cleanup_required0，**不是执行清理PASS**；公开新冻结投影/新页面子集/后台保存200/SQL与清理均未验。反馈与回归仍不关闭。source_completed；publication blocked；acceptance not_started，不修改现代化workflow状态。

[本次安全执行收据](../scripts/test/results/mng-identity-minimal-20261007/publish-ssh-blocked-20261007-155646.json)。下一恢复TCP/22后沿原批准，从fresh主机/三服务/真实活跃与Worker待处理0预检继续，再受限备份与回滚、最小产物切换/一次重启及专属合成验收；不要求重复批准同scope。

## 2026-10-07T15:07Z 资源有界归因：七项已查明，393恢复；完整编译门禁仍阻断，未发布

**ALL_ARTIFACT_GATE 未通过；发布、重启、合成身份保存均未执行。** 不是再次请求最小 staging 发布批准；原批准保留。用户本轮转交精确两源独立 CodeReviewer PASS，本轮原生核 SHA 仍匹配 Go `b0526da8…`、candidate `d1afa257…`；此事实记为 `USER_SUPPLIED_PASS_HASH_MATCHED`，本 BreakGlass 没有派发或自称执行独立复审，不修改旧收据的 `NOT_EXECUTED`。

- **七项缺口实证**：恢复工具读取旧 source-scope（641项），实际其中394个前端src、0个public；输入清单本身没有public，不能归因为路径过滤不匹配。隔离两目录因此均缺 public；394 是 src 输入数量，不包含 public。这纠正上节“394个src/public”表述。另一个真实旧 build-inputs 收据共406项＝394src＋6public＋6配置；六public（含首页模板）和六配置当前 SHA 均精确匹配旧收据。未从猜测的截图/字体/客户上传补资源，也未复制私人文件。
- 完整七个发布路径是 `exam-entry.html`、其gzip、favicon、`html/ie.html`、其gzip、robots和`static/tester.xlsx`。五个静态源直接来自已核 public，两个gzip由原 CompressionPlugin 生成。15:07:28Z fresh 线上七项逐 SHA 同旧manifest；无字体、运行手工JS或用户上传资源。详细路径及SHA依据见[旧完整manifest](../scripts/test/results/mng-default-release-20261006/new-dist-manifest.json)、[新限定收据](../scripts/test/results/mng-identity-minimal-20261007/artifact-bounded-verdict.json)。
- 新 C区[有界验证器](../scripts/tools/mng-identity-artifact-gate-20261007.js)只补齐 B区隔离输入，原两生产补丁/原工作树保持。真实 `build:prod` 脚本为 `vue-cli-service build`；实际两次 `npm run build:prod -- --dest dist-faithful` native0/Build complete，Node16.20.2，两份393文件，各原两体积warning，未升级依赖或更改生产配置。package/package-lock/Babel/Vue配置SHA同旧；原公开输入缺失是已证原因，但不是所有编译差异的唯一原因。
- **数量恢复不等价**：[基准构建](../scripts/test/results/mng-identity-minimal-20261007/baseline-faithful-build.json)首页SHA488025e5…、[候选构建](../scripts/test/results/mng-identity-minimal-20261007/candidate-faithful-build.json)52eecf04…仍不同旧353c…；[完整路径比较](../scripts/test/results/mng-identity-minimal-20261007/faithful-baseline-comparison.json)347项新/不同、345旧路径缺失。旧/基准各1589 unique module ID，但只374同ID、1215分别新增/缺失；重合335项raw不同，其中同ID在不同构建不能假定同模块。合法重复实例428也不能用覆盖map丢弃。未取得严格图映射/AST等价证明，不能称只有candidate变化或“buildmatches393”。
- 第三轮验证器改为原 Webpack cwd＋只读虚拟输入，输出仍仅bin：[真实失败收据](../scripts/test/results/mng-identity-minimal-20261007/baseline-context-build.json)180000ms后native null/SIGTERM/ETIMEDOUT、资源0、stdoutSHA01ba…/stderr16840bytes SHA05655…、原源码SHA保持。候选此策略构建未执行。有界只读探针exit0/服务加载成功但 `fs.realpathSync.native` 为undefined；未取得其与timeout的严格关联，不猜唯一cause或增加预算。三次验证器编辑已用完，停止第四改/复制脚本绕预算，所有失败保留。
- 后端候选完整SHA仍 **4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c**；原源审阅PASS不能代替编译包门禁。原24Gin/106SFC及build/vet只沿已有同SHA收据，不重跑全仓、不称本轮新测试GREEN。
- fresh strict SSH首次exit0/stderr0/1548ms：hostname/user正确，PID2006，backend8baa…/index353c…/dist393，三服务active/healthok；全产品state1=0、overdue competency=0、draft表0。仅只读SQL，不读人员值/答案。原用户页面未操作，未打开新页面/刷新/保存；真实synthetic保存200、SQL清理、backup/rollback和新发行公网资源验收均NOT_EXECUTED。

todo1资源数量/静态缺口归因完成，完整artifact门禁BLOCKED；todo2发布NOT_EXECUTED；todo3真实保存NOT_EXECUTED。远端backup写/上传/替换/restart/DDL/DML/生产0。下一需新有界驱动修正或独立模块图证明，不需要重新批准同一发布范围；不得上传本轮未证明的393包。下方14:50历史及旧失败收据保持。

## 2026-10-07T14:50Z 最小 staging 补丁：精确后端基准恢复，隔离 GREEN，ready for source review

**未发布；独立复审与前端编译资源归因未完成。已有最小 staging 修复/部署授权保留，不需重新批准同一范围。** BreakGlass 禁止派发其他代理；此处不是独立 CodeReviewer PASS，也不是部署可用性 PASS。原工作树业务源/原测试/配置未修改，远端上传、备份写入、替换、重启、DDL/DML、生产访问均0。

### 精确来源与最小隔离

- 已恢复 UF053 发布时215个 Go 文件及最新默认前端394个 src/public 输入；17个旧字节版本从**对应代码文件**的 VS Code 本地历史取得，并逐 SHA 匹配清单，其余当前字节匹配。没有扫描凭据/个人数据历史或使用 Git HEAD 代替线上源。19个不属于发布基准的新增 Go 文件在 overlay 明确排除；不回滚或删除原文件。
- 原生 Linux 基准 build exit0、stdout/stderr0、49902830bytes，SHA **8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676**，与 fresh 在线后端**完整字节一致**。此前“无法取得精确旧源”的发布门禁已限定解除；旧21历史漂移不重新归因或回滚。
- 隔离工作区位于 [构建 overlay](../Go-based%20Refactored%20System/bin/mng-identity-minimal-20261007/candidate-overlay.json)，恢复工具位于 [C区工具](../scripts/tools/mng-identity-minimal-stage-20261007.js)。不改原 Go/Vue 工作树，所有生成源码/构建产物留在忽略的 B区 bin；证据留在 C区测试结果目录。
- [唯一 Go 运行改动](../Go-based%20Refactored%20System/bin/mng-identity-minimal-20261007/candidate/internal/handler/exam.go#L306-L367)：仅 Detail 及 strings import，从既有 `ManagementTraitsIdentityScope`/完整 `ProfileDetail` 投影同exam冻结字段与严格 bool；分类/profile失败关闭，旧/non002显式false；不含草稿生命周期、正式审批、报告或重发改造。原白名单、路由、令牌、审计及历史数据路径保持。
- [唯一前端运行改动](../Go-based%20Refactored%20System/bin/mng-identity-minimal-20261007/candidate-ui/src/views/paper/exam/candidate.vue#L216-L433)：复用草稿改造前已验证切片，可信当前ID/strict bool/字段子集、unknown禁止旧payload、加载与迟到响应屏障；明确旧false保历史链。不消费 `isManagementTraits`/`managementTraitsLifecycle`，不需要新表。既有 [API helper](../Go-based%20Refactored%20System/bin/mng-identity-minimal-20261007/candidate-ui/src/api/managementTraits.js) SHA a90b1f67fa834fdefd4bc11d0eea91dcdc86fa965bb9098cb43a7b55492f9d07，与线上源码基准一致，未修改。

### TDD 与 fresh 证据

- 精确旧源＋既有严格身份测试实际 RED：Gin8fail/16pass事件（不是编译失败），SFC/API35fail/71pass，两个native exit1；strict decoder原拒绝及configured注册原成功均保持。原 RED 和失败收据保留。
- 只在隔离副本装入两侧修复后 GREEN：Gin **24pass/0fail/0skip/parse0/exit0**（8公开Detail＋6strict＋1body＋9configured create）；SFC/API **106pass/0fail/0pending/exit0**。106为草稿协议前的身份回归，不冒当前109含draft或全435；合法三字段、未选gender/选gender空值、unknown/null/别名拒绝、跨exam/URL spoof/迟到响应均保留。Gin/sqlmock与真实编译SFC/mockAPI，非真实MySQL身份保存。
- candidate Go全包build/vet/Linux server三native exit0，stdout/stderr均0；候选49903310bytes，SHA **4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c**，仅本地未上传。子进程清23个opt-in键，不改用户环境；未重跑6272/560UI/完整前端。
- 前端baseline/candidate两fresh Vue生产构建均exit0/Build complete，保留原2体积warning；两份各386资源，source map0。baseline index **659e5bf0894c1aa9b30d46052e14b30619e6f58111e681b2e1b8d9478387888c**，**不等于**线上353c9fb3…；candidate index e338aff8df48d409a091c89c26eff8339baf11e73aadb99c07df3831f858706e。两隔离构建按路径418项差异（含增删命名），唯一src/public字节变化为candidate.vue；具体Webpack路径/模块ID/资源命名差异原因未完成实证，不能称393资源等价或仅candidate编译差异。不得部署此dist作为已通过包。
- 14:44:14Z与14:50:45Z严格SSH首次均exit0/stderr0：PID2006/线上backend8baa…/index353c…保持，talent-assessment/nginx/mysql三active，state1全0、draft表0，后一次目标profile frozen1、requiredFields name,gender,telephone及health statusok。SQL只读、不输出姓名手机号，不提交用户截图中的真实身份。最后0活跃是单时点，不替未来维护切换前TOCTOU复核。
- 新工具/两个隔离业务文件诊断0。工具历史保留：一次PowerShell嵌套引号解析失败发生在SSH前；一次历史URI前缀匹配未找到，纠正后精确SHA匹配成功。未以失败返回空伪造缺源码。

### 主协调者接续（不是新授权请求）

1. `CodeReviewer` 只读复审精确两运行差异及恢复/测试收据；本worker未派发、不自审冒独立通过。
2. 前端构建负责者完成原路径上下文的可复现基准或模块级等价归因、线上393资产完整验证，再确认候选编译差异只来自身份切片；**当前386资源候选不满足发布门禁**。
3. 已批准维护发布负责者复核所有活跃/启动即结算卷为0、受限全库/原binary/dist/config/unit/权限备份与回滚有效后，仅一次正常后端重启，先真实公开Detail冻结bool/字段合同，后前端切换；不重启nginx/mysql、不执行DDL/安装draft。
4. 正常管理员认证下仅独占synthetic identity注册200及SQL精确核验/精确清理；原用户页面只刷新读取，不能提交其真人值代验收。真实保存验收本轮未执行，用户反馈仍未关闭。

收据：[精确恢复清单](../scripts/test/results/mng-identity-minimal-20261007/recovered-source-manifest.json)、[基准binary精确匹配](../scripts/test/results/mng-identity-minimal-20261007/baseline-linux-build.json)、[RED](../scripts/test/results/mng-identity-minimal-20261007/red-tests.json)、[GREEN](../scripts/test/results/mng-identity-minimal-20261007/green-tests.json)、[两运行文件差异](../scripts/test/results/mng-identity-minimal-20261007/candidate-source-diff.json)、[候选build](../scripts/test/results/mng-identity-minimal-20261007/candidate-builds.json)、[未通过的前端归因门禁](../scripts/test/results/mng-identity-minimal-20261007/frontend-build-diff.json)。本切片 ready for **source review**，整体发布未完成，不能要求用户仅刷新就称线上修复。

## 2026-10-07T14:30Z 只解决截图保存故障：线上部署差异确认，最小发布阻断

**未发布，线上反馈未关闭；没有向服务器发送任何身份保存请求。** 本轮仅诊断/复验本bug，暂停报告、审批、草稿与完整UI后续；未重复修改已存在的业务修复。todo1诊断完成、todo2已有本地修复复验完成、todo3发布门禁阻断（收口后active=0）。独立CodeReviewer仍由主协调者安排。

- 真实公开路由为[POST Detail](../Go-based%20Refactored%20System/internal/router/router.go#L214-L220)。本轮HTTP200/code0/success=true、当前ID一致、repoCode=00201、requiredFields=name,gender,telephone；managementTraitsProfileFrozen、isManagementTraits、managementTraitsLifecycle三键全部不存在。严格SSH首次exit0：在线backend SHA8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676、index SHA353c9fb3a0f30540c8ca091885b67bf458249c1023eca909f78c83988bf27de2、PID2006（非沿历史17552）；两次独立只读命令前后PID/SHA保持，stderr0。
- SQL只读事务核exam存在1/isOpen1/三字段，profile存在1/frozen_at非空1；冻结requiredFields数组准确为name/gender/telephone，field_contract SHA051dbe83d91a184e273ae6ce72748952c44cad4e4163ec82dcf6bf3c83aa2b6b；el_mng_exam_draft表存在数0。未读姓名电话等人员值，不查询答案或修改任何配置。
- 独立同host页面加载真实线上资源；普通/my入口实际managementTraitsMode=false、examBlocked=false，显示姓名/手机号/性别。用Synthetic/13800000000/0合成值走真实handleSave→正常确认，仅在浏览器abort精确candidate/save请求，examId匹配；捕获13键，额外8键affiliation/age/degree/depart/idNumber/major/post/stuFlag，其中前7键为空或null。第一次直接submitForm探针未经过handleSave，examId为空，只算辅助观察；根因证据采用第二次完整保存确认路径。两次均remoteSaveSent=false，最后unroute/reload清合成输入，用户原页面/账号保留，无JWT导出。
- 根因是线上缺冻结元数据且旧candidate仍走全表单；[strict decoder](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_identity.go#L43-L62)拒未知键/null，[固定错误](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_identity.go#L133-L139)与截图一致。gender本身已在真实合同内，不删除其必填、不用sex/mobile aliases、不放宽后端或以URL解锁。不是本轮真实远端保存错误复现：写请求被主动拦截，不能宣称远端保存200或反馈已解决。

### 本轮实际验证（不沿历史计数）

- 原生Vitest两文件109pass/0fail/0pending/exit0：实际编译SFC覆盖普通server-frozen字段子集、strict当前ID/boolean、unknown/跨exam关闭、历史明确legacy/false与draft拒绝；实际API封装[精确字段payload](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-api.spec.js#L77-L91)覆盖三字段且丢弃其他键（包括空值）。SFC使用Element stub/API mock，不冒完整浏览器或实库。
- 当前全前端435pass/0fail/0pending/exit0；此前全量exit1本轮未复现，不能追溯其原因。JSON的numTotalTestSuites=86是suite计数，不当86文件；focused同字段5也不当5文件。原VTU提示保留。
- 真实Gin/sqlmock Detail/strict/body合同16pass事件/0fail/0skip/parse0/exit0；另[ConfiguredCreate](../Go-based%20Refactored%20System/internal/handler/management_traits_identity_chain_test.go#L94-L145)9pass事件（1顶层+8子项）/0fail/0skip/exit0，三字段成功、未配置gender/配置gender空值ROLLBACK及五键响应均通过。编辑器发现范围27pass/0fail单列，不替原生计数。无Go业务修改，不重跑全Go/build/vet/6272或560UI。
- fresh vue-cli production build exit0/Build complete，原2体积warning；index SHA33f1ec0d0bcad7f32ba802896ba16a0aeb719559b05f667ebf018e91ad7a3dd2，**仅本地未上传**。candidate/API/两测试四文件执行前后SHA相同；没有删除此前报告/草稿代码或格式化无关文件。

### 最小发布门禁与必要后续范围

线上dist393文件、source map文件0；与当前本地构建按原路径逐SHA比较：352同、2不同（index及gzip）、39原路径本地缺失（内容哈希命名资源已变化，不代表线上丢文件）。未完成新旧编译模块等价归因，不能据352同宣称onlycandidate。当前[candidate](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L247-L267)和[新版配置加载](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L359-L368)已带未部署的lifecycle/isManagementTraits依赖；当前[Detail](../Go-based%20Refactored%20System/internal/handler/exam.go#L328-L350)调用草稿生命周期，整包还含其他未部署功能。禁止整包发布或执行草稿DDL来凑合同。

**frontend-only不满足线上合同。** 必需最小范围：以在线旧后端为基线，仅Detail使用已有IdentityScope/ProfileDetail完整验证后投影当前exam的冻结requiredFields及严格frozen bool（错误关闭、历史显式false，不公开profile/PII）；前端只取对应candidate冻结bool/当前ID/字段子集修复切片，复用线上已有API白名单，排除新draft/formal及新建配置UI变更。若选择发布当前candidate，则还须同Detail准确三态元数据，但不得顺带安装draft表。本轮未取得可证明与在线旧后端对应的完整源码隔离构建，不拼凑旧21SHA、不变门禁标签或绕测试。需主明确确认最小后端补丁范围及安全维护窗口，再受控overlay/独立review/定向测试/受限备份/发布验收；现有staging修复授权保留，不重新请求同一前端授权。

本轮DDL/DML/业务实体创建/备份写入/上传/dist切换/backend替换/service restart/production访问均0；只读SQL与浏览器请求拦截不是远端写验收，cleanup_required=0而不是执行清理PASS。线上仍存在同一故障条件，下一不能只要求用户刷新。

## 2026-10-07 新建旧版禁用政策与staging授权（尚未实施/发布）

**新建门禁未实现；草稿持久化契约未定，因此本轮未发布。** 用户已明确批准先发布staging修复并验收（20.200.136.133不转生产），范围包括candidate修复与新建旧版002禁用。既有结果/PDF保留、未完旧卷只核查不自动关闭；不启用正式报告，不发布未完成formal registry功能，不执行其DDL。无需重新索取同一发布授权。

新政策覆盖10-06“默认新版、允许手动取消”，但不覆盖历史编辑/历史测评策略。只禁复选框或新增不持久化的请求boolean不构成完整服务端门禁：

- [Save请求与写入](../Go-based%20Refactored%20System/internal/handler/exam.go#L405-L449)、[新建事务](../Go-based%20Refactored%20System/internal/handler/exam.go#L589-L610)及[exam写入](../Go-based%20Refactored%20System/internal/handler/exam.go#L669-L681)没有独立002草稿模式，新建非competency仍为legacy+legacy/published；唯一运行创建路径搜索命中该Save，router对应POST save。未发现另一个业务exam创建handler；测试夹具创建不计业务入口。
- [前端保存/独立冻结](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L742-L779)当前取消冻结后保留已保存配置；[服务器身份分流](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L108-L143)依真实profile存在性，不读请求意图boolean。故只要求创建请求true不能证明未冻结草稿禁止旧链。
- [Exam字段](../Go-based%20Refactored%20System/internal/model/business.go#L68-L103)无独立002模式；[现profile校验](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_profile.go#L24-L52)要求完整冻结时间/映射/合同，[profile读取](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_create.go#L270-L283)同样严格。不得写半成品profile冒合法新版或放宽校验。

需协调者先确认：保留两步确认和取消留草稿时，以独立持久化模式识别新草稿并封闭旧入口（需存储/API/所有读写影响分析）；或者明确改变为创建与冻结原子完成（改变现确认/取消语义，不能自动选择）。本轮不擅自选schema、复用competency字段、自动freeze或用时间/ID推断新旧。

本轮只更新本报告及项目记忆，未改业务/测试/SQL/环境；未执行编译、测试、真实DB、浏览器、SSH、部署或restart。既有106/430/Gin15等仍为历史验证，不冒本轮fresh。新门禁完成后由协调者安排独立CodeReviewer，随后fresh staging预检/受限备份/回滚及验收；若有nonowned活跃或待处理卷，先阻断申请维护窗口，不能自动结束。必要backend restart仅在安全预检通过后说明并执行，不沿UF053旧预算冒授权。

## 2026-10-07 CodeReviewer warning：仅补本地契约回归

- 审阅状态：用户转交CodeReviewer PASS；非冻结标记断言warning已补，不代表本轮重新执行独立CodeReviewer。此前“未执行”记录保留为历史。
- [唯一测试改动](../Go-based%20Refactored%20System/internal/handler/exam_management_traits_candidate_test.go#L77-L83)：实际Gin响应经json.Unmarshal解码为map[string]any；legacy、absent_schema、non002三个成功分支必须有managementTraitsProfileFrozen键、动态类型为bool且严格false，requiredFields仍为name,gender,telephone。缺键/null/字符串/数字/true不再通过；frozen严格true及错误响应不回退断言保留。现有后端已输出准确布尔值，本轮是测试合同补强，不是业务根因修复；没有修改产品制造RED。
- 实跑go test -json ./internal/handler -run '^(TestBugManagementTraitsCandidatePublicDetail|TestBugManagementTraitsCandidateStrictRejectedKeys|TestBugManagementTraitsIdentityConfiguredBodyContract)$' -count=1：15pass事件（3顶层＋12子项）、0fail、0skip、parse0、native exit0、stderr0；三个非冻结成功案例均通过新增键/类型/false条件。聚焦测试完成编译；未重跑全Go、430前端、SFC或服务构建，不以历史结果冒fresh。
- 修改前526个Go/前端源码及测试/依赖声明SHA基线，最终仅上述测试文件变化，其余525项相同；测试诊断0。无签名/API/消费方变更，无业务源码/环境文件/Git/SQL/远端opt-in/SSH/部署/重启操作；未验真实DB、浏览器或远端。生产部署仍未授权、目标未确认，由主协调者后续核对。
- 验证工具历史：首条长终端命令未取得结果后中断；缩短后实际聚焦exit0。新同步会话读取旧内存基线失败exit1，回原终端复核取得BASELINE_FILES=526、CHANGED_COUNT=1及唯一目标测试路径；不把该会话失败当代码RED或隐去。

## 2026-10-07 续作：MT-CANDIDATE-FROZEN（仅本地）

**独立CodeReviewer未执行：当前BreakGlass模式禁止派发其他代理。远端/真实MySQL/完整浏览器未验；未部署、无SSH或数据操作。** 原普通入口实现及旧verdict已实际读取，不将旧收据当本轮执行证据。正式PDF/两端功能、生产目标与旧版历史政策不在本切片，TEST警示和所有后端白名单保持。

[纠正 - 2026-10-07] 原“配置畸形关闭”不完整：新版loadManagementTraitsConfig只检查requiredFields，不检查服务器冻结标记/响应测评ID；普通002入口也把缺失标记视为旧版。因此可在URL显式mngTest时仅凭字段串解锁，或未知标记回退旧全表单。真实SFC基线12fail/82pass/exit1确认，非仅字符串自审。

- [普通入口](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L246-L262)：以URL或服务器任一002分类施加标记检查，必须当前测评ID匹配且标记为布尔；只有明确false进入既有历史链。伪造非002 URL不能绕过服务器002分类，缺失/null/字符串/数字标记全部关闭。
- [新版解锁](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L341-L361)：当前测评ID匹配（字符串或正安全整数）＋managementTraitsProfileFrozen严格true＋原有非空/唯一/白名单字段检查后才解除examBlocked；第二次读取false/缺失仍关闭，不回退legacy。loading、失败、代际屏障沿原实现保持。
- [回归](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-participant.spec.js#L93-L174)：既有14项冻结/ID测试保留，新增8项未知标记及4项URL/false跨测评拒绝。成功夹具补齐实际Detail布尔/ID合同，旧false成功路径不删除。

影响：candidate组件与该SFC测试同步修改；api/managementTraits白名单封装、公开Detail布尔投影、Candidate.Save/strict decoder、guard/router、管理端与其他详情消费者无需修改（协议已存在，本次仅消费收紧），后端生产/SQL/模型/配置0改。不存在函数签名或API结构变更。

本轮有效RED：原磁盘12fail/82pass；新增未知标记8fail；最后URL/false跨测评4fail/60未选择，均exit1。中间工具陈旧编辑器内容与原生磁盘不一致，首次测试patch覆盖了14项磁盘测试并造成4项缺合同夹具失败；已从实际读取内容恢复全部14项及ID/布尔夹具，不删除负向断言。此工具失败不称新业务RED。

最终实际身份SFC/API106pass/0fail/native0；全前端现有任务32files430pass/0fail（未暴露数字exit）；最终vue-cli build native0/Build complete，保留原2项体积warning及Browserslist通知；编辑器两文件0错误。实际Gin公开Detail/strict身份专项15pass事件/0fail0skip/parse0/native0，未重跑全Go/vet。两源文件执行前后SHA精确相同，父验证exit0；完整保护目录SHA清单本轮未采集。SFC编译脚本/模板＋Element stub和mockAPI、Gin/sqlmock，不冒称真实浏览器/实库/远端身份保存通过。[本轮独立收据](../scripts/test/results/mng-candidate-frozen-local-20261007/verdict.json)保留，旧verdict不覆盖。生产与测试各3轮编辑后停止，不继续其他逻辑功能。

## 未验证与边界

远端修复未发布、未验收；正式报告后续切片、旧版新建入口停用未在本任务实施。禁止部署、重启、改环境门禁、改真实配置/冻结合同/身份/答案或历史PDF。独立CodeReviewer未执行（BreakGlass禁止再派代理）。

## 分诊与只读事实（修改前）

五问：开放考生身份保存失败；历史是否可用未定；仅指定测评已确认；截图报告性别必填及配置外字段参数错误；目标exam1791298091700970647，标题仅以SQL布尔相等核验。不保存或输出真实姓名手机号。

2026-10-07严格SSH第一次exit0/stderr0，只读事务：exam存在1/isOpen1/state0/titleMatches1；当前requiredFields=name,gender,telephone；profile存在1/frozen1，冻结字段也是[name,gender,telephone]，合同SHA051dbe83d91a184e273ae6ce72748952c44cad4e4163ec82dcf6bf3c83aa2b6b。截图name+mobile子集不等于当前持久化数据，原因/修改时间未证。不改配置；用合成name+telephone夹具验证合法子集。

在线后端8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676；前端index353c9fb3a0f30540c8ca091885b67bf458249c1023eca909f78c83988bf27de2。无远端写入。

## 已核源码链与根因

- 管理配置使用requiredFieldsList字符串键，watch写requiredFields逗号串；模型required_fields同名映射，没有inquiry数字ID或sex/mobile别名转换。合法性别键gender、手机telephone。
- freeze将requiredFields复制到field_contract；严格服务从冻结合同取配置。公开Detail原来只返回可变exam.required_fields，没有可识别的新profile标记。
- candidate.vue只凭route mngTest或本地participant token分新旧；在线列表/二维码普通入口无此标记。因此未获token的新版考生走旧rules和saveData(candidateForm)，而非已有registerManagementTraitsCandidate白名单函数。
- 旧rules仍含固定gender；全旧表单携idNumber/depart/null等，strict decoder在service前拒绝，正是配置外字段参数错误路径。性别本身不是该参数错误唯一原因；本次当前真实合同选了gender，要求填写符合合同。

## 修改前C2/C5影响清单

不改签名/既有字段含义。仅Detail对具有可信冻结profile的002返回额外managementTraitsProfileFrozen=true及冻结requiredFields；其他字段/后台权限不扩张。元数据失败不伪装legacy。候选字段来自服务器，不从route/storage/复选框推断。

| 消费方 | 处理与理由 |
|---|---|
| candidate.vue created/config/showField/rules/submit | 同步修改：等待server配置，识别新标记，复用现有白名单；异步旧响应不能跨exam生效；配置失败禁止提交 |
| exam/form.vue fetchData | 无需修改：新增标记被忽略，冻结字段与只读profile一致；原单独admin profile探测/保存冻结不改 |
| user/exam/index.vue | 无需修改：详情新键不参与旧人员分流，既有guard不变 |
| paper/exam/tester.vue、preview.vue | 无需修改：既有新token/显式intent路径；不新增封闭身份字段写入 |
| paper/exam/result.vue、result2.vue | 无需修改：既有paper保护及TEST结果隔离不变 |
| exam/competencyResults.vue | 无需修改：00401不进入002新profile投影 |
| api/exam/exam.js、api/managementTraits.js fetchManagementTraitsExamConfig | 无需修改：既有详情协议容纳新增键；独立client无admin凭据 |
| api/managementTraits.js registerManagementTraitsCandidate | 无需修改：现有唯一字段白名单已经足够；不复制工具函数，不对象spread |
| handler Candidate.Save / strict decoder / identity service | 无需修改：保留null/未知键/别名/未配置非空字段拒绝，测试实跑 |
| model / freeze / import-export / formal registry / TEST gate / router | 无需修改：不改存储合同、配置审批、权限、报告或环境 |
| participant SFC测试、新增真实Detail/Gin测试 | 同步修改：queryless、子集/性别、pending/failure/异步/legacy及服务零写拒绝 |
| 既有API/admin/resume/legacy与全量Go/frontend测试 | 无需改期望，实际回归；6229历史基线不覆写 |

## 实施前分支矩阵

queryless server-frozen name+telephone / 三字段gender / 无profile旧002 / 显式新模式pending-failure / 当前exam切换慢响应 / 配置畸形或缺失 / 未配置gender与未知字段后端零写拒绝 / 公开响应仅新增标记与字段名。先RED再业务patch。

## 本地完成（远端用户反馈不关闭）

1. [服务端Detail](../Go-based%20Refactored%20System/internal/handler/exam.go#L306-L348)：分类查询错误明确失败；仅精确002+legacy作用域使用已有IdentityScope、完整Schema/ProfileDetail校验，投影冻结requiredFields和managementTraitsProfileFrozen布尔。不新增公开profile/权限接口，不输出field_contract/manifest/mapping/人员/资产路径。
2. [考生组件](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L216-L348)：先等配置，普通入口遇服务端标记改走原TEST身份链；配置失败/畸形禁止提交，重试保留；route切exam重置旧表单，configRequest屏障隔离配置/恢复/身份迟到响应，旧资料回填不再打印个人信息。
3. [提交](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L391-L424)：复用原registerManagementTraitsCandidate，不复制白名单工具、不发旧全对象；name+telephone无gender规则，选gender仍必须填写。不改变strict decoder接受范围；旧测评成功路径仍saveData。
4. [SFC回归](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-participant.spec.js#L98-L153)、[现有API精确payload](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-api.spec.js#L77-L85)、[新Gin Detail/拒绝回归](../Go-based%20Refactored%20System/internal/handler/exam_management_traits_candidate_test.go)、[已有真实注册矩阵补四例](../Go-based%20Refactored%20System/internal/handler/management_traits_identity_chain_test.go#L98-L109)均已同步。其他消费方按上述清单无需改，原期望回归保持。

## 验证证据

| 执行 | 实际结果 |
|---|---|
| 修改前实际SFC subset/queryless/pending/error/late-config | exit1，7fail/2pass/25未选择；非编译失败 |
| 修改前实际Gin Detail+现有strict拒绝 | exit1，7pass事件/6fail事件；parse0。失败含缺冻结投影/错误回退；原strict拒绝已正确 |
| 补充迟到identity响应 | 有效RED exit1/1fail/37未选择→加请求代际后GREEN |
| 补充服务端repo分类读取失败 | 有效RED exit1/0pass2fail事件→检查Error后GREEN |
| 最终SFC+真实API封装两文件 | 80pass/0fail，exit0；真实SFC脚本+模板编译，Element子组件stub，不冒完整浏览器或ElementUI全校验证据 |
| 前端全量已有Vitest任务（一次） | 32files/404pass/0fail；任务工具未暴露numeric exit，不伪填；保留原VTU deprecated提示 |
| 最终go test -json ./... -count=1 | 6247pass事件/675通过顶层/0fail/9原skip/parse0/exit0；子进程清23实库开关，未连接真实测试DB |
| go build -o bin/server.exe ./cmd/server、go build ./...、go vet ./... | 三项numeric exit0，stdout/stderr均0字节 |
| 前端真实vue-cli production build | exit0/Build complete；构建提示及Browserslist年龄通知保留，不升级依赖 |
| 编辑器诊断 | 六代码/测试文件0错误 |

最初6246全量完成后，又补repo分类失败RED及最小错误传播，所以重新跑最终6247；不是用旧源码全量冒最终GREEN。原6229/53正式版本专项不改，差量18事件/2顶层，本任务不关闭formal PDF/生产门禁。一次apply_patch重复路径拒绝未修改文件，合并同路径后正常应用；本次API测试参数数组在执行前按Vitest单参数约定包装，未放宽断言。

公开线上只读POST Detail进一步核实HTTP200/code0/success=true，requiredFields=name,gender,telephone、repoCode00201，**managementTraitsProfileFrozen键不存在**。这是当前接口断点实证，不是用户当时浏览器路由或失败原始body的抓包。未知的是具体哪个旧键/null触发及截图时配置变化；不宣称gender唯一根因，也不追溯改历史证据。

## 安全/范围收口

3866→3867保护快照，既有源仅两生产文件及三个测试文件变动，新增一个Go测试；scripts原脚本/旧收据变化0、formal/guard/router变化0、Legacy变化0。新增本安全收据在快照后，不覆盖旧receipt。SQL仅只读事务和字段名/hash/布尔投影，无身份值；公开HTTP仅配置读取，不保存用户信息。

- profile冻结合同、真实exam、已有身份/答案及所有历史PDF未写；没有造TEST数据或reset，没有DDL、部署、restart、.env/环境门禁操作。
- 参数strict拒绝继续有效：idNumber:null/depart/grade/sex/mobile在解码前拒绝；未配置非空gender在真实事务内ROLLBACK/零INSERT，选gender空值仍拒绝；姓名+手机号及配置三字段均正常成功，身份响应仍原五键。
- 本地全量/HTTP为Gin+sqlmock，不是真实MySQL写入、远端UI或完整浏览器保存验收。没有复跑560题harness。公开Detail多了完整profile验证的元数据查询，queryless转换还沿已有独立client重读配置；性能负载未测试，不顺手改缓存/抽象。
- 本地切片completed；远端修复未部署/未验证、截图当时字段差异未归因、独立CodeReviewer未执行。下一由主协调者安排只读CodeReviewer；任何staging/production发布须另明确授权，正式后续与旧版新建入口停用另续，不自动执行。

机器证据：[verdict](../scripts/test/results/mng-candidate-fields-local-20261007/verdict.json)。