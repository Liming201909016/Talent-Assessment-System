# ErrorFixer Progress Details

## 2026-10-09 — Staging rollback canonical schema evidence

- Root cause：原drill只核三张additive表行数0，未保存或比较table/column/index/FK metadata；旧rehearsal receipt又只覆盖两张reissue表并缺engine/collation/default/extra，不能后验证明三阶段结构。
- RED/Fix：新增Node合同先RED列出全部缺失字段；rollback脚本改用新attempt目录，按有序HEX canonical格式保存PRE/OLD/POST三份0600 receipt，每阶段立即与迁移预期SHA比较，最后三方比较；历史partial receipt显式标记不可exact比较。Node合同与bash syntax均0。
- Re-validation：fresh active paper0/additive rows0；授权staging完整复跑PID14293→15461→15554，恰2次restart。三receipt各6328 bytes、3 tables/29 columns/20 index columns/7 FK columns、byte-identical、SHA均`1a9f16e82a3facd55d418d87bf85771c4d6d23eb5edc88c531b8d96c30ddbe27`且expected match。旧/新路由边界、app errors、Nginx5xx均通过。
- Final：PID15554/NRestarts0，三服务active、内外health200；当前server/index/front恢复；config/protected DB/PDF/frozen002保持；state1/overdue/additive rows/auth residue/runtime test binary/errors/5xx均0。无待决策；仅staging证据，不批准production。

## 2026-10-09 — FB-217 localhost resolver hardening

- Root cause：`resolveServerRuntimeControls`按字符串信任大小写任意localhost并把hostname交给监听器，既未验证完整DNS结果，也会在监听时发生第二次不受控解析。
- RED/Fix：focused测试先因缺resolver参数编译失败；现注入`serverIPResolver`纯函数，仅精确小写localhost执行一次固定2秒解析，要求非空且全部loopback，优先127.0.0.1否则排序IPv6并以`net.JoinHostPort`绑定数字地址。显式IP保持零DNS；错误、空、非loopback、混合及hostname变体全部fail-closed。
- Re-validation：`cmd/server` focused package exit0；`go test ./... -count=1` exit0；Windows server build exit0；`go vet ./...` exit0；diagnostics0。第一次focused GREEN尝试继承`GOOS=linux`产生Win32执行格式错误，显式Windows目标后通过并恢复环境。
- Independent review：复审补强了localhost空白变体拒绝和“无127.0.0.1时优先排序IPv6”的明确实现；补强后再次执行focused/full/build/vet全绿。无待决策；未启动服务、未访问DB/远端、未部署。

## 2026-10-09 — Generic safe local server controls

- Root cause：server只构造`:port`，无法为本机候选绑定loopback；competency Worker在router中无条件启动，management Worker在main中仅受TEST runtime gate控制，没有统一、fail-closed的本地禁用门禁。
- Fix：新增纯`resolveServerRuntimeControls`，host只接受显式IP或归一化localhost并用`net.JoinHostPort`；默认空/0.0.0.0保持`:port`。新增仅`APP_ENV=local`+loopback可用的Worker禁用配置，并通过router options同时控制competency与management Worker；TEST route gate保持独立。
- RED：focused测试先因`ServerCfg.Host`/`DisableBackgroundWorkers`、resolver及router worker helper缺失而编译失败。
- Re-validation：focused `cmd/server`+`internal/router` GREEN；全Go所有package GREEN（service 114.448s）；Windows server build=0、`go vet ./...`=0、diagnostics=0，五个改动Go文件规范化后与gofmt一致。当前terminal原有GOOS=linux导致首次Windows执行报格式错误，明确以进程级GOOS=windows复跑并恢复原环境。
- Decision：源码已完成，未启动服务、未访问远端/DB、未部署。按用户顺序，新可复现candidate package须在独立CodeReviewer通过后再构建，本轮未提前打包。

## 2026-10-09 — MT-005 harness receipt / local readiness GREEN

- Root cause：producer合同已正确增加并断言`legacyMarkerRejected=true`，但独立harness仍用旧6字段deep equality，形成测试合同遗漏；不是产品缺陷。
- Fix：只在C区harness exact expected object加入该安全字段；producer、Go/Vue、SQL、模板均未改。完整7字段逐项比对确认无其他遗漏。
- Re-validation：修改前RED exit1；修改后harness exit0，focused 005前端208/208、全前端598/598、diagnostics0。旧RED verdict不覆盖；新rerun verdict以SHA绑定同日Go/build/coverage/static/runtime/oracle证据并判定GREEN_LOCAL。
- Decision：无需产品决策；旧marker必须fail closed。未部署，staging/production/browser E2E及环境opt-in测试仍是外部门禁。

## 2026-10-09 — MT-005 reset head canonical bytes（FB-215）

- Root cause：head scan只做严格JSON解码和字段/文件名前缀校验；legacy outcome-prefix不绑定raw SHA，因此同语义空白或字段重排可通过。
- Fix：两种文件名都用writer相同canonical serializer重建head，并要求raw bytes精确相等；不canonical或内容篡改均fail-closed，历史文件不改写/不移动。
- Re-validation：FB-215先RED；随后FB-212～215、完整helper package、reset helper build、server build均GREEN。无remote/DB写/新reset sequence；当前无reset-journal只读命令，actual scan未执行。

## 2026-10-09 — MT-005 reset journal final hardening（FB-212～FB-214）

- Root cause：scanner只按prefix/后缀局部校验，未精确约束schema、filename sequence、outcome删除算术及head内容SHA；pending只看DB/file数量；outcome已落盘而head失败会形成永久orphan。
- Fix：exact v2 decoder/filename regex/full link-count contract；新intent持久化排序后的pre/post私有文件fingerprint multiset；恢复读取实际root并精确比较。唯一最高next有效outcome可在scan时create-exclusive补写head，多个orphan拒绝。旧sequence1～3历史head按原outcome-prefix只读兼容，新head使用内容SHA前缀。
- RED/GREEN：FB-212～214先因fingerprint helper缺失编译RED；实现后focused验证、helper build、Node contract及actual standalone reset结果在本轮终验追加。不访问remote/staging/shared DB。

## 2026-10-09 — MT-005 reset journal v2（FB-208～FB-211）

- Root cause：旧流程先提交DB删除和文件删除，再写单个reset receipt及覆盖式`reset-chain.json`；落盘失败无法证明操作意图或恢复实际删除计数。report-only v3另把raw examId复制到公开receipt。
- Fix：新增immutable intent/outcome/head三类v2文件，全部create-exclusive且file fsync；intent在mutation前写入pre/expected-post counts与hash-only file identifiers，outcome在postcondition后写入。链状态只扫描不可变文件，旧master/receipt只读排除并以legacyHeadSHA桥接；pending只允许精确pre/post恢复。report-only v4及identity/oracle改用examIdHash。
- RED/GREEN：新测试先因旧test内容重复及v2符号缺失RED；实现后focused package GREEN，覆盖O_EXCL、legacy不变、rewrite拒绝、known-state recovery、mixed fail-closed及injected outcome-write failure保留pending。
- Re-validation：actual report-only、standalone reset、Python oracle和实际receipt递归隐私扫描尚待本轮后续执行；不接触staging/shared DB。

[后续实际结果] secure launcher在active generation `1791483691965`只执行一次有效report-only：participant answer-save/start/submit均0，生成两份报告后post-reset删除reports2/files2；v2链sequence1/2。随后standalone sequence3为reports/audits/files删除0、retained runs/products=2/2，outcome SHA `7c13e70e…`、previous outcome `3e6403d0…`。report-only v4 core SHA `86f2a365…`，两PDF SHA `b1777ded…`/`bf92f1ad…`；无prepare、无UI答题、无staging/shared DB操作。首次直接Node16缺模块、第二次直接Node22因stdin credential合同有界终止，均未进入有效report-only；secure launcher一次成功。

## 2026-10-09 — MT-005-RESET-RECEIPT-CHAIN（FB-207）

- Root cause：reset helper只输出不完整stdout，E2E丢弃结果并硬编码`cleanStart:true`；没有持久化序号、模式、实际删除/保留计数或manifest/evidence/helper SHA，也没有可验证hash chain。
- RED/GREEN：两个focused tests先因receipt/helper缺失编译失败；实现append-only `reset-NNNNNN.json`、master chain、完整字段/隐私门禁，report-only v3绑定pre/post SHA并解析actual stdout。focused Go、helper build、Node/PowerShell syntax通过。
- Actual预跑：reset已执行并写首链节，但Go map的JSON字段顺序使launcher过滤器未转发该行，E2E按设计失败而未生成报告；改用ordered output struct，失败链节保留。后续只运行report-only与standalone reset，不prepare/280、不remote。

## 2026-10-09 — MT-005 active generation baseline（FB-206）

- Root cause：E2E prepare已生成动态marker，但launcher/Go/reset仍绑定旧固定path/marker，且report-only只有`answerClicks`，无法证明HTTP participant writes为0。
- RED/GREEN：新增generation regression先因helper/field缺失编译RED；实现v2 safe/evidence generation、private v3 generation、DPAPI active pointer、dynamic reset及三类HTTP request counter。旧固定marker明确拒绝为reusable generation。
- 实际执行：首次prepare在00501返回409，exact cleanup1；第二次完整两链后发现counter只监听admin context，exact cleanup2；修正每个participant context后第三次成功。成功generation的safe manifest在首次private capture前补齐遗漏的product/question version，证据SHA未改；此修正及失败均不隐藏。
- 终验：active `1791483691965`两链exact counts/orphan0/report0；report-only双轮零participant writes且PDF oracle双GREEN；最终reset两次均0并重检业务链不变。远端/部署/production=0。

## 2026-10-09 — MT-005-RETAINED-EVIDENCE-BINDING（FB-205）

- Root cause：FB-204的DPAPI manifest在baseline创建后从当前DB回算身份，并绑定一个未解析且属于后续reset/regenerate链的report receipt SHA；该SHA不能证明baseline创建时的身份。现存prepare目录只有含raw ID的mutable manifest，没有创建时独立receipt或受保护原始SHA绑定。
- Fix：v3 manifest必须读取受保护私有副本`baseline-evidence.json`，同时验证文件SHA以及product/exam/run/title/三身份字段hash和`expectedReports=0`；删除无语义校验的report SHA绑定。未来prepare在submit响应时生成hash-only evidence，launcher仅在同一次成功prepare调用后复制到LocalAppData SID-only文件并把SHA写入DPAPI runtime。
- Current retained baseline：未删除、未写DB；因缺少原始独立创建receipt，明确不可复用，`InspectPartialBaseline`现在以`BASELINE_EVIDENCE_MISSING_NOT_REUSABLE` fail closed。多个后续report receipts虽exam/run hash相同，但name hash漂移且报告链已reset，不能后补为创建证据。
- RED/GREEN：FB-205先因独立evidence类型/加载/验证符号缺失编译失败；matching fixture、name drift、report-count mismatch、missing receipt四分支已实现。最终验证结果记录于本节后续复验条目。

## 2026-10-09 — MT-005-RETAINED-BASELINE-MANIFEST（FB-204）

- Root cause：FB-203仍以workspace manifest的exam/title/code和聚合计数为主，未持久化candidate/paper/snapshot/run/receipt原始ID，未绑定product/question/scoring/norm/source及mapping/input/evidence/identity hashes；obsolete父链删除前也未保存全部child ID，无法事后枚举任意已删ID。
- RED：四个sqlmock测试先因manifest/orphan/query symbols缺失编译失败；真实执行路径覆盖parent-absent全局orphan=1拒绝、candidate/identity hash漂移拒绝、mapping/input漂移拒绝、意外report row拒绝、exact rows通过。不是map-only测试。
- Fix：launcher从exact safe manifest两条exam/title/code一次只读查询完整raw synthetic ownership，将candidate/paper/snapshot/run/receipt及report/current/audit/reissue IDs写入工作区外当前SID-only DPAPI manifest；后续只由stdin加载，CLI/控制台/workspace不含raw child IDs。workspace receipt仅保留ID/hash、版本、source、mapping/input/evidence/snapshot/field-contract hashes、counts及报告ID hashes。
- Closure：参数化exact obsolete 20表均0；另执行20条无参数全局LEFT JOIN orphan查询，全部0。00501/00502各1/1/1/1/1/1/140/140/700/1/13/4/1，product/question/scoring/norm/questionnaire/source及全部immutable hashes精确匹配。revision/current/audit和reissue/audit五类ID数组均明确为空。
- Historical limitation：obsolete chain删除前旧manifest没有child IDs，不能事后证明“每个已删ID”；全局orphan=0只关闭父缺失child残留风险，不证明不存在另一条完整、非孤儿且无法归属的链。本轮无DB写/delete/remote/service restart。
- Re-validation：package test、helper build、`go build ./...`、package vet、PowerShell parser、actual `InspectPartialBaseline`和独立`InspectRuntime`均exit0；normal inspector仍显示两产品reports=0/audits=0。无待决策。
- Final review：重新执行diagnostics、safe receipt raw-ID/UUID/19位ID扫描、`gofmt -d`、focused package test与package vet；先发现并修正两份Go文件的纯格式差异，随后全部通过（`FINAL_CODE_REVIEW_CHECKS=PASS`）。本轮scope内未发现新的blocking；workspace其余既有未提交改动未纳入本次结论。

## 2026-10-09 — MT-005-PARTIAL-CLEANUP-INSPECTOR（FB-203）

- Root cause：旧只读检查器用父表JOIN取得ownership；父记录为0时直接返回`LOCAL_PARTIAL_BASELINE_ABSENT`，没有逐表核残留，也没有确认现存00501/00502基线未漂移。
- RED：三个focused tests先因新检查边界未定义而编译失败；分别锁定parent0/child1失败、all0成功、retained count错误失败。
- Fix：[runtime inspector](../Go-based%20Refactored%20System/bin/mng005-runtime/main.go)对obsolete exam/bundle逐个参数化精确查询20张相关表，不使用LIKE/FK关闭/写操作；现存两个baseline身份由launcher传入不可变baseline manifest，随后按exact exam/title/code核1/1/1/1/1/1/140/140/700/1/13/4/1。
- Launcher仅为`InspectPartialBaseline`注入已存在ownership manifest路径；普通`InspectRuntime`协议不变。未远端、未写DB、未重启服务。
- Re-validation：focused package、inspector build、actual read-only inspector及普通InspectRuntime结果见本轮最终回复；无用户决策。

## 2026-10-08T14:43Z — MT-005 三项复审（2项实现，identity full commitment被既有证据阻断）

- RED：[三阻断收据](../scripts/test/results/mng005-report-e2e-20261008/code-review-three-blockers-red.json)先锁定两PDF SHA，并实证旧产物缺AST node位置/hash、完整versioned commitment及encryption/stream counters。
- AST修复：[exporter](../scripts/tools/mng005-canonical-contract.go)现在同时解析真实identity/scoring源码，严格核`ManagementTraitsDimensions`三语句build/同变量return、13维catalog、`CalculateManagementTraits`唯一四模块聚合loop，并全文件拒绝含维度值/符号的其他mapping-like composite。Node合同对decoy identifier、partial alternate map、模块顺序漂移均拒绝，unrelated benign strings通过；四AST node SHA=`51980057…/beea5a89…/676f1bfe…/31436abb…`。
- PDF修复：[oracle](../scripts/test/management-traits-005-report-pdf-oracle.py)拒needs_pass/encrypted；逐一扫描1866 xref object text、82 raw/decoded streams、trailer/catalog/metadata及附件，按UTF-8/UTF-16LE/BE/latin1和raw ASCII模式扫描。真实最大decoded stream为10,688,445 bytes，默认8MiB gate按设计先拒；配置16MiB单对象、64MiB总量后GREEN：object/raw/decoded=`254372/1030825/31595162` bytes，attachments0、credential findings0、encrypted/needsPass0。
- identity v2公式已加明确marker并绑定formula/product/exam/run/report/title/producer/safe-receipt/PDF/3 field hashes；未来正常E2E会从真实detail写`runIdHash`。但既有browser receipt只有reportIdHash，DB/report已按前序清理，PDF stream无UUID，且本任务禁止UI/DB/report/service/remote；[blocked receipt](../scripts/test/results/mng005-report-e2e-20261008/identity-commitment-v2-blocked.json)因此精确拒绝`00501:runIdHash`、`00502:runIdHash`，records为空，不编造final commitment。Node合同对此fail-closed行为PASS。
- 未修改两PDF，SHA仍`0850bc50…`/`9d925561…`；四代码diagnostics0、Node双syntax/Python AST0。决策：若必须关闭full commitment，只能在未来允许的新E2E生成前先保存runIdHash；现有已清理证据不可后补。

## 2026-10-08T14:33Z — MT-005-PDF-ORACLE-FOUR-BLOCKERS（GREEN）

- RED：[四阻断收据](../scripts/test/results/mng005-report-e2e-20261008/code-review-blocking-red.json)在修改前精确缺`canonicalSourceSHA/identityCommitmentSHA/pdfObjectCredentialScan/recursiveReceiptCredentialScan`，PDF SHA锁定未变。
- 根因：旧oracle复制Python维度表/硬编码synthetic allowlist，只扫PDF提取正文和拼接JSON文本；没有绑定真实Go catalog、E2E fixture来源、JSON嵌套/Base64或PDF metadata/xref/附件对象。
- 修复：[Go AST exporter](../scripts/tools/mng005-canonical-contract.go)严格解析真实私有catalog并确认`ManagementTraitsDimensions()`消费它，输出[canonical contract](../scripts/test/results/mng005-report-e2e-20261008/canonical-contract.json)；[E2E hash-only模式](../scripts/test/management-traits-005-report-e2e-local-20261008.js)输出[identity commitment](../scripts/test/results/mng005-report-e2e-20261008/identity-commitment.json)，未来正常receipt也带hash，不输出明文。
- 复验：[Node合同](../scripts/test/mng005-oracle-contract-test.js)exit0；[Python oracle](../scripts/test/management-traits-005-report-pdf-oracle.py)AST/oracle exit0。Go source/function SHA=`4b417188…`/`51980057…`，模块3/4/3/3；commitments 00501=`eb8af41a…`、00502=`7d5d5bde…`；每份身份field hash3/手机号1。
- 隐私扫描：3 JSON findings0（Unicode key、Base64 depth2/64KiB自测）；两PDF 22 metadata values、1866 xref objects、0 attachments/findings0。PDF SHA仍`0850bc50…`/`9d925561…`；四文件diagnostics0。未UI/DB/report generation/services，无待决策。

## 2026-10-08T14:31Z — MT-005-PDF-ORACLE-REVIEW（两个warning GREEN）

- 根因：原oracle只独立推导13维/总体，模块只复制浏览器receipt数量；也未证明PDF只含批准synthetic身份或安全收据无凭据值。
- RED：[独立收据](../scripts/test/results/mng005-report-e2e-20261008/code-review-warning-red.json)明确两份均缺`independentModuleScores/privacyVerdict`。修复仅[oracle](../scripts/test/management-traits-005-report-pdf-oracle.py)：140×raw3/40反向一次→13维精确Rat→canonical四模块等权；四键各一次且self/interpersonal/task/development均50。
- privacy：实际E2E源固定synthetic姓名/性别/手机号allowlist；PDF可提取身份精确匹配、单位/职务空，11位手机号仅allowlist。PDF与browser/db-audit/cleanup三安全JSON扫描JWT样式及JWT/Bearer/Authorization/password/口令/密钥，findings=0；每份身份标记2、phone1。
- 复验：venv3.14.4 AST0，focused oracle两PDF PASS，diagnostics0；PDF SHA前后保持`0850bc502a61009ed47335c75b20384fa19f2bbd9c082275d6064d186c1eed87`/`9d925561bcd85470e074fd4b35e30937cfc60147cc4d40116c219ff5dab64aec`。未UI/DB读写/报告生成/服务操作，无待决策。

## 2026-10-08 — MT-DEBUG-ATOMIC-COMMIT（最后CodeReviewer blocking GREEN）

- 根因：atomic helper在Move/Replace后仍对生产目标执行`Set-Acl`和断言，任何postcommit失败都会形成“调用报错但新目标已生效”；既有目标也未在temp创建前验证private ownership。
- 修复：[launcher](../scripts/tools/mng005-local-debug.ps1)先验既有目标，再完整预验temp；commit后仅置位并return，finally只清理未提交temp。[合同](../scripts/test/mng005-runtime-binding-contract-test.ps1)覆盖Move/Replace真实ACL、解密/字段、目标先验和commit前注入旧SHA/temp0。
- 复验：parser 0/0、合同native0且10项状态全true；真实VerifyRuntime exit0，runtime SHA前后同`45FA2E5D5C9756A7B89D6455C7788CAFC7EB6916ABE083B218A8D6BEA4CE29E3`。无DB写/UI/remote/restart；无需决策。

## 2026-10-08T13:20Z — MT-DEBUG-RUNTIME-BINDING（后续2 blocking + warning GREEN）

- 根因：13:11版launcher正常同步会自动改写既有Redis绑定，且DPAPI密文直接写目标；旧文件可能在失败中损坏，stale owned身份也可能被无声接受。另Go verifier仅fake inspector覆盖，无实际Windows实现门禁。
- 修复：[launcher](../scripts/tools/mng005-local-debug.ps1)新增同目录私有密文temp、Flush/ACL/nonreparse/解密JSON预校验、Replace/首次Move及精确清理；正常路径只首次绑定，partial/stale均拒绝，只有显式`RebindRuntime`可更新五个绑定且不动秘密。可dot-source，不含生产故障注入开关。[PS合同](../scripts/test/mng005-runtime-binding-contract-test.ps1)先缺helper RED，后覆盖原子失败、清理、stale、首次、partial与rebind GREEN。[Go FB-202](../Go-based%20Refactored%20System/bin/mng005-runtime/main_test.go)默认skip，显式本机仅接收非秘密身份元数据并调用实际Windows inspector。
- 验证：package test/vet、PS合同、actual inspector、VerifyRuntime均0；首次合成Replace曾报IOException，后独立Replace探针、跟踪合同及普通合同均PASS，失败保留。真实runtime密文SHA前后同，backend PID20036和Redis PID12324未重启；direct/proxy200、NOAUTH true。最终源码SHA：launcher `5682a83fae527851c606b075856a5dacace5e9392287a4f5f1567aebb1314e9f`、PS合同 `5b749ff7b7f8b2ac49bd345e566f695b860de4b39cde25f599b80a5be9cfd3fe`、Go test `388dbf8cf8dc4134ad57efa6a369f0bc5e94df94a6fd2333ac0fbfd36939c66d`。无额外决策；不远端/DB写/280 UI/部署。

## 2026-10-08T13:11Z — MT-DEBUG-REDIS-OWNERSHIP（两个blocking本地GREEN）

- 根因：原launcher仅保存Redis端口/密码，Go入口只做认证PING；无法证明127.0.0.1:23317监听者是本任务Memurai、使用工作区外受保护的精确配置，也未保证正常backend启动复用verify门禁。
- RED：`TestBugFB200_RedisRuntimeRequiresExactOwnedProcessAndConfig`、`TestBugFB201_RedisRuntimeRequiresAuthenticatedLiveState`先因binding/verifier缺失编译失败。GREEN拒绝wrong host/port/PID/listener PID/executable/private ACL/config SHA/requirepass/auth/DBSIZE，保留运行中DBSIZE非负合同；package test/vet/build与PS parser均0。
- 修复：[Go runtime](../Go-based%20Refactored%20System/bin/mng005-runtime/main.go)在任何DB操作模式或router/Worker前核listener/process/config/secret/live state；[launcher](../scripts/tools/mng005-local-debug.ps1)从Get-NetTCPConnection+CIM取得当前owned身份并写入当前用户DPAPI。秘密不进argv/output，main不读取命令行。
- actual首verify安全拒绝暴露PowerShell参数传递工具缺口，改为整数PID内联＋config路径仅子进程环境后通过。Redis PID12324、config SHA4c9b12f0…及listener/executable/ACL/nonreparse/hash/auth全match、DBSIZE1。
- 仅旧owned backend PID2736停止并重建启动PID20036；Redis/Vue/隧道不动。重启后verify/live均0，direct/proxy/captcha200、NOAUTH true；不登录/报告/UI重跑/远端DB写/部署。[收据](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/runtime-redis-ownership-20261008.json)。无需用户/协调者额外决策；交独立review复核。

## 2026-10-08T10:34Z — MT-DEBUG-CREDENTIAL-ARGV（已轮换／DPAPI同步／源数据保持）

- 根因：[原恢复脚本](../scripts/db/mng005-local-debug-restore.sh#L56-L65)将随机密码置于CREATE USER的mysql-e argv；历史已执行，不宣称无暴露。最小stdin heredoc＋set+x修正，原双host/八项copy-only GRANT不变；[合成真实Bash回归](../scripts/test/mng005-mysql-stdin-contract-test.js)RED native1/argvtrue/stdin0→GREEN0/argvfalse/stdin2。
- [专属安全轮换](../scripts/tools/mng005-local-credential-rotate.ps1)：精确ownership/备份SHA/root权限/两host/grants/roleproxy0、当前用户DPAPI/SIDACL/现有23316门禁；先旧密文backup与新pending密文，密码仅SSH/MySQLstdin，positive_app原整行/grants只远端内存核前后同，无源业务写/全局grant/env/cache/deploy/restart/appstart。
- 初次格式gate拒绝DDL0，真实backtick/排序及--raw修正后PASS。一次本地字符串SyntaxError remote0。远端真实轮换和前后校验完成，PS5裸$null使File.Replace失败；新pending保留，NullString第三编辑修正，只验证和同步既有pending、无第二次轮换/旧密码回滚。新及活动Go连接0、旧密文BLOCKED1、源1142、crossFK0；原launcher/Go核验器/前端598不改不重跑。
- [独立10:34:38终验](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/credential-rotation-final-1791455678358.json)native0/stderr0：78表source-copy SHA c27fdfb8…同，源完整schema/data/原资产PDF/cache基线同，双host原八grant/roleproxy0/PID2746/restarts0/healthok/backend4179/front52eec保持；当前general_log0不回推历史。原备份/收据保留，应用未启动、独立review由主安排。无需扩大范围决策；历史风险继续保留。[完整说明](management-traits-005-local-debug-20261008.md)。

## 2026-10-08T07:57Z — MT-REISSUE-1062（LOCAL GREEN，REAL PENDING／未发布）

- 未完：strict SSH首次＋唯一重试均连接前255timeout，remote0；新真实MySQL/具体冲突键/修复后1report-generate1-reuse1-file1-orphan0尚未实采，独立CodeReviewer由主执行，不能发布。07:28线上健康/4179/52eec/PID2746/主两表0/58.64仅历史，不伪fresh。
- 根因路径：run普通读在paper UPDATE前建RR view，serialized loser报告Find仍旧view；原INSERT1062不恢复。仅[服务](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_api.go#L129-L262)捕获typed1062精确报告input unique，原TXROLLBACK后一次fresh reuse-onlyTX，完整来源/同tuple/rawSHA/template-content/generated审计/PDFsize-SHA再核，reuse审计失败关闭；other1062/audit/1213/取消/撤销等不恢复，render1/只清loser/无旧写。
- [回归](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_api_test.go#L516-L693)23分支RED5/19→专项111/0/2skip；测试第三编辑修source-drift读后缀夹具，不改断言。首full native1/134.755s/0fail事件、无timeoutstack；保持120s、全discover service四批＋所有其他packages6383/0/11skip/694top/parse0/native0，Windows/allbuild/vet0。最后exact隔离73/0/1skip/build-vet-Linux-external0，[完整收据](../scripts/test/results/mng-reissue-stage-20261008/bug1062-scope-final.json)。
- 118 runtime只target服务1变；current/candidate SHA261b90bb…逐字节同，green overlay tests3a40…同当前，原ab07083e发行bin/DDL保留，新Linux3b2b…仅local。编辑器格式与EOFraw门禁失败已exact同步；bin subset gopls standalone缺依赖不改go.mod，overlay真实编译为权威。旧overlay/manifest/旧tests拷贝不可冒本bugfresh发行输入。
- 原真实fixture加两RR连接/两个paperlock前readview屏障及仅固定key类别；新C区[bug-only工具](../scripts/tools/mng-reissue-bug1062-20261008.js)final external binary绑定df0478c8…/source-testSHA，协议保护/exact随机库finally/无publish。全部新负向/rawsecret/fileaudit证明为local，不实库PASS。无task lifecycle调用/部署包/前端/主库/生产操作；[详细变更与责任人](management-traits-reissue-staging-release-20261008.md#L3)。

## 2026-10-08 — MT-ADMIN-STAGING-BOOL（LOCAL GREEN，待主独立复审／未发布）

- 根因：当前管理SFC误把未部署draft合同字段当已发布strictbool旧Detail的必需。仅[产品入口三行](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L314-L334)，旧false前提严格同exam/合法ID/legacy两轴/实际002，标记缺失或一致可选legacy/false允许；存在unknown/畸形/矛盾仍关闭。新stricttrue、draft及freshgetInfo权限/字段/profile原门禁不动，不放后端guard。
- 先[实际DTO SFC回归](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-admin-ui.spec.js#L232-L285)有效RED109pass5fail/native1；专项167/0/0/full33files560/0/0/native0，prodmode buildnative0/原2warning/两diagnostics0。首相对路径ENOENT仅工具失败保留；测试三编辑、产品一次；原39矩阵/force/迟到/旧list所有权保留。
- internal Go实际222/聚合ffe5ab6a…前后同，API5be56c…、隔离Go candidateab07083e…保持，不重build Go/跑旧6272或6359。仅此SFC＋既有测试/docs，依赖/字体/客户资产不改；SQL/SSH/远端HTTP/备份上传/部署/restart/真人PDF/旧卷旧PDF0，browser不跑。
- 独立复审能力由主协调者安排，本worker未调用或冒PASS；真实MySQL两表/竞争/发行闭包/正常admin browser未验，source ready非release ready。原staging授权保留，不补新授权或自行继续部署。[完整政策/失败/保护/运行收据](management-traits-reissue-staging-release-20261008.md#L3-L15)。无task lifecycle调用。

## 2026-10-08 — MT-ADMIN-REVIEW-3（源码/单测GREEN，浏览器STOPPED）

- 三根因：类型/生命周期未核全就old返回；metadata loading早退使生成后新刷新丢失且旧响应覆盖；旧list未绑定query/exam/seq。仅[旧入口](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L303-L362)/[列表](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L389-L407)及[结果metadata](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/managementTraitsResults.vue#L157-L207)，两单测真实DTO＋39回归，本地driver和必要账本，不API/Go/DB/remote。
- 实际RED27fail100pass/native1、首相关199pass0、full任务523pass2fail→真实夹具同步后33files525pass0fail/native0，fresh build0、diagnostics0。222Go前后聚合a2c6b44f…同，API5be56c…保持，旧40/API53/续答/00401原分支保留。原风格/TEST警示/姓名手机缺口未变。
- 未完先列：独立CodeReviewer不可用须主调用；浏览器第一legacy keyed双探测触repeat-submit/native1，nextTick停用watcher保护后第二有界120000ms/null/SIGTERM/parent1、stdoutstderr0/no-summary。driver三编辑停止、390和完整两case不称GREEN，终验owneddriver/Chrome0，旧Node9316/1764不kill；不增预算/第四修/旧PASS回填。
- 仅ready for source independent review，整体验证PARTIAL，不release ready；主需独立复审及新有界browser根因预算。无task lifecycle/任务目录/SSH/部署。[完整移交/前后SHA/失败证据](management-traits-admin-ui-local-20261008.md)、[停止收据](../scripts/test/results/mng-admin-ui-local-7e74014c951c/bounded-stop.json)。

## 2026-10-06 — MT-DEFAULT-CLEAR-P1（local GREEN，独立复审blocked）

- 根因：clear未重置row.repoCode，未手动取消的00201/00202同库重选被误判为连续选择。先真实SFC新增4项，RED83pass/2fail/native1；唯一产品增量为[清空选择身份一行](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L803)，[回归](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-admin.spec.js#L181-L199)保留连续同库显式取消false。
- 验证：admin85/85、既有六相关224/224/native0，全前端任务一次32files389pass0fail（任务接口未暴露numeric exit），npm build:prod native0/Build complete，两代码diagnostics0；原warning保留。682项SHA仅form/admin测试变化、215Go同/永久driver0，无Go/SQL/远端/部署或旧21rollback。
- 未完成：本地浏览器clear等待三次同类失败后停止，actual DOM已unchecked；native stdin探针确证注入中文→002????，cases0/资源closed，不称六case通过。独立CodeReviewer调用能力缺，返回blocked供主协调者重新派发精确一行＋4项回归审阅；本worker未调用生命周期。新默认未发布、正式内容/远端完整UI门禁保持。
- local1/local2的SFC与build完成，浏览器blocked单列；stage3notstarted须review＋用户frontend-only staging发布确认，stage4记录完成。旧失败不删，首两JSON解析错误不当有效RED。[完整收据及限定纠正](management-traits-new-default-local-20261006.md#L3-L12)。

## 2026-10-03 主代理最终纠正：两项安全门禁完成

- 下方“验收失败后停止”是worker中间状态，保留原记录。本次主代理查证GORM1.25.12：Set为clone0，Session NewDB clone1不复制Settings；最终Session{} clone2保留容量指针并隔离Statement/错误。runtime文件第3轮收敛；schema文件仍3轮，未再修改。
- 非法ff夹具恢复为真实非法UTF8字节，未禁合法U+FFFD、未删除/skip断言。注入实例实际DB/secret/预算与cfg不一致及显式nil均失败关闭，DI第二轮GREEN。
- 最终主代理focus2546/0/0、全量5290/0/7、JSON解析错误0；build/vet0、关键10Go diagnostics0；独立最终review PASS。两项实际完成，不只局部报告；未真实DB/SSH/DDL/部署、未前端303复跑，並行write/report/resume单独验收。
- worker此前未遵守“docs只返回主代理写”，本文件有中间交接写入；主代理保留事实并追加纠正，不覆盖并行代码，不将此前流程违规隐去。详见[当前完整证据](management-traits-local-implementation-20261003.md#L3)。

## 2026-10-03 — 实际引用容量／私有 registry 第3轮（已编码，验收失败后停止）

- 执行用户补充授权：仅 schema/runtime/identity/load/source 与既有 behavior/load tests；schema 本次只有一次生产补丁，累计第3轮已用完；references test 累计3轮，本次零修改。不触 handler/router/main/write/report/DDL/models/API，不调用生命周期工具，无真实DB/SSH/部署。此追加仅为 ErrorFixer 必需的交接记录，未改其他业务文档或账本。
- RED：新增7个顶层 behavior tests，原生产实现34个失败事件、0通过、exit1；其中 Row.Scan nil panic、JOIN参数绕过、Raw源加载容量、constructor共用registry／once替换DB、public loader绕过schema均真实复现。加载fixture改为零值Dest后独立复验21个子项全部行为失败，exit1。没有删除原测试或新增Skip。
- 实际修复：[构造器](../Go-based%20Refactored%20System/internal/service/management_traits_runtime.go#L22)创建同pool私有callback registry；[schema](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L216)只atomic发布capacities、不替换s.db，并增加Built Query/JOIN参数及AfterQuery真实模型/映射/options预算检查；[身份标量](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L98)改安全Scan.Error；[公共loader](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_load.go#L130)前置schema，UNION owner加载显式校容量；[Raw source](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_source.go#L32)对实际Scan所得relation/question/option ID核容量；load test两个public fixture补原完整schema expectations。
- 公开路径盘点：FreezeProfile/ProfileDetail/CreatePaper/PaperDetail/FillAnswer/Submit/ListRuns/ResultDetail/LoadValidatedRun/TryRegisterCandidateIdentity/TryTesterIdentity/IssueCandidateResume/GenerateTestReport/ReadTestReportPDF；既有CheckRuntimeSchema通过同私有registry接收capacities，LoadValidatedRun补遗漏前置。不是新增HTTP路由或目标环境验收。
- 最终focus：219pass/8fail/0skip，exit1；full：5248pass/8fail/7skip，UTF-8 JSON parse errors0，exit1。go build -o bin/server.exe ./cmd/server、go vet ./...、scoped git diff --check全部exit0；七个Go文件编辑器diagnostics0。计数包含顶层与子项事件，未称覆盖率或全绿。
- 剩余失败为3个顶层／8事件：ReviewJoinParametersFailClosed四个子项均在非法查询后的合法sentinel查询失败（并非非法JOIN放行）；原ActualExamQueryAndProfileWrite在合法查询处失败；ReviewLoadedIDsAndJSONBudgets/question_options/ff仍返回nil。前两组需检查constructor Set返回的GORM statement克隆/错误污染；最后一项fixture把非法UTF-8字符串经json.Marshal编码为合法U+FFFD，不能以禁止合法U+FFFD修复或弱化契约。未进行第4轮试改，也未声称以上推断已修好。
- 结论：本地PARTIAL且不可交付／部署。需主代理审查剩余根因并取得新有界修复授权；保持references test与第3轮schema现状供审阅。无主动回滚或清理用户改动。

## 2026-10-03 — 新授权容量局部receipt，剩余安全项明确停止

- 根因：schema父varchar仅检查<=64，paper/pq实际生成UUID36但接受1..35；旧兼容测试还把三父列改1/32期待成功，与用户重新授权的严格下限冲突。
- 一轮修复：[validator](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L114-L125)仅paper/pq加36下限；[兼容矩阵](../Go-based%20Refactored%20System/internal/service/management_traits_schema_behavior_test.go#L150-L175)保留短容量case转明确拒绝。原容量70RED不改，无skip/delete，11表15FK/type/collation/error/cache不弱化。
- 原生RED71fail exit1→全量4642pass/0fail/6skip exit0，PARSE_ERRORS0；go build -o bin/server.exe ./cmd/server及go vet ./...退出0；两Go diagnostics0。原6skip不变；未前端303复跑、真实DB/SSH/部署。
- 未完：实际profile/exam引用ID对cache父容量检查、非锁定bundle撤销竞争、统一paper/bundle事务锁、actual cfg/DB共享service DI及管理员purpose短时resume签发/消费。短期在已验证切片后明确停止，不claim安全闭环或可部署。
- resume已确认合同及P0 case登记[最新本地报告](management-traits-local-implementation-20261003.md#L3)，未注册endpoint、不交假路径给前端、不手机号匿名恢复、不resetdeadline、不unfreeze，到期只状态/完成。正式文件/客户源/Word/label/font/star/footer不动，生命周期由主代理负责。

## 2026-10-03 — MT-WORD-RUNTIME-01

- 目标：重新授权后的独立有界002图表渲染修复；不调用任务生命周期、不建任务目录、不修改其他产品、客户原件、模板SHA或运行激活。
- 根因：数字标签误要求恰好3个w:t；真实总体/任务为4个，尾随空白不属于分值槽。manifest绑定第二节点；六图title/name与rId→OPC路径实核正确。
- 修改：[独立渲染器](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word.go#L110-L112)仅兼容3/4节点且第4必须空白；[原测试增强](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word_test.go#L34)核混合分值六图/五槽及非值XML/其他OPC字节保持。仅一轮生产代码修复，原测试名称未变。
- 验证：原生RED charts exit1→GREEN；002聚焦1813pass/0fail/0skip；Go全量2525pass/0fail/5既有skip，build退出0，诊断0；原候选17项实际PDF＋10项source合同0fail/error/skip。编辑器编辑后未发现测试，不当作GREEN。
- 实际输出：真实Go全50DOCX→LO26.2.5.2 exit0，A4 9页/542627-byte PDF，第3页五环图＋13维柱线图可见；混合值25/75/100/0/28.85正确、可见占位符与替换字符0；source/candidate/XLSX SHA未变。
- 未关闭：既有“测试报告”标记DOCX存在但PDF文本不可见，only-chart范围不修改定位逻辑；真实DB/HTTP/UI/持久化/目标环境未验，formal/运行未启用。LO prefix警告仍保留；无SSH/SQL/DDL/DB写/部署。整体PARTIAL，不可宣称可交付。
- 影响/文件清单和完整证据见[本地报告顶部](management-traits-local-implementation-20261003.md#L3)；project-memory/business-branches/regression-tests只追加本bug记录，其他既有修改未覆盖。