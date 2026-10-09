# 普通002恢复核验与独立副本本地连接 — 2026-10-08

## 2026-10-09 reset head canonical-byte最终阻断关闭

- 仅本机ignored reset helper与回归：legacy outcome-prefix、新head-SHA-prefix文件名继续兼容，但head内容必须逐字节等于`json.MarshalIndent(..., "", "  ")`后追加单个newline；额外空白、同语义字段重排及内容变更均拒绝，历史文件保持原位且不自动修复/隔离移动。
- FB-215先复现legacy空白/重排被接受，再GREEN；FB-212～215、完整helper package、reset helper和server build均通过。未执行reset、未增加sequence、未读写DB、未remote。launcher当前没有只读reset journal检查命令，因此不报告actual sequence1～3已重扫。

## 2026-10-09 generation-aware active baseline与report-only双轮完成

- 范围严格为本机loopback/owned copy；无staging shared写、部署或production。active generation `1791483691965`使用新marker；旧固定marker baseline与其缺失evidence状态原样保留，runtime显式active pointer不会选择旧记录。
- prepare使用正常synthetic admin自动登录和真实UI，00501/00502各创建/冻结、candidate保存、create paper、140次fill answer、manual submit；创建时request计数恰280/2/2，report请求0。v2 evidence在任何报告前create-only落盘，`expectedReports=0`，随后同launcher复制到generation-specific LocalAppData SID-only文件并绑定DPAPI runtime SHA。
- 两次失败generation `1791482751739`/`1791482962778`分别以exact title IN(A,B)发现1/2条主键链，report0门禁后按FK顺序清理，owned residue0；无LIKE/FK关闭/旧baseline或fixture变更。成功后safe manifest遗漏两个版本字段，首次private capture前仅补product/question version并由DB exact验证；创建evidence未改。
- private v3 capture和inspector：20 orphan查询全0；每产品1 exam/profile/draft/candidate/paper/snapshot、140 snapshot questions、140 legacy questions、700 buckets、1 completed run、13 dimensions、4 modules、1 receipt；revision/current/audit/reissue/reissue-audit为空。
- report-only run1/run2均零answer click/save/start/submit，生成两份customer TEST PDF、重复生成复用、view/download均200。两轮PDF oracle均`PASS_LOCAL_SYNTHETIC_ONLY_WITH_IDENTITY_V2_COMPLETE`，合计xref1866/streams82、credential findings0。每轮reset后报告0；最终显式reset两次均reports0/audits0/completedRuns2/deletedReports0，重检closure/orphan/业务链不变。

## 2026-10-09 retained baseline exact inspector与global orphan闭包

- 仅本机owned copy只读；无DB写/delete、remote、service restart。launcher以现有safe manifest的两条exact exam/title/code查询一次copy DB，把raw synthetic candidate/paper/snapshot/run/receipt及revision/current/audit/reissue IDs写到工作区外当前SID-only、nonreparse、DPAPI的`retained-baseline.dpapi`；原子temp/Replace复用既有helper，stdin传Go，不经CLI/stdout。workspace [safe receipt](../scripts/test/results/mng005-partial-inspection-7081fbec31e3d105/inspection-receipt.json)只含hash。
- 参数化obsolete direct查询20表均0。另有恰20条parameterless global LEFT JOIN关系：profile→exam/bundle、draft→exam、snapshot→paper/exam/profile/bundle/participant、question snapshot→snapshot/paper_qu、run→snapshot/participant、dimension/module/receipt→run、revision/current/audit、reissue/audit；全部orphan count0。
- retained 00501/00502各exact 1 exam/profile/draft/candidate/paper/snapshot、140 snapshot questions、140 paper questions、700 buckets、1 run/13 dimensions/4 modules/1 receipt。product=`mng-00501-v1`/`mng-00502-v1`、question db-current、scoring/norm/questionnaire/source及mapping bytes/hash、manifest/input/evidence/participant snapshot/field contract、三identity hashes均直接按private manifest谓词/比较验证。
- 报告零闭包不是聚合猜测：private manifest的revision/current/audit/reissue/reissue-audit raw ID数组均为空，actual exact查询五数组也为空；safe receipt五个hash数组明确`[]`，normal InspectRuntime同样两产品reports=0、四audit count0。
- 历史限制：obsolete链清理前的pre receipt只有exam/bundle聚合，没有每个child raw ID；删除后不能重建或证明任意已删ID。global orphan0关闭“父不存在但child残留”风险，不证明另一条完整且非孤儿、又不在ownership manifest中的链不存在。
- TDD：四个FB-204 sqlmock测试不是map-only，实际执行query builder/scan/validation；orphan1、wrong candidate/identity hash、mapping/input drift、unexpected report row均拒绝，correct通过。package test、helper build、all build、vet、PS parser、actual inspector和normal inspector全exit0，diagnostics0。

## 2026-10-08T15:24Z 共享staging保留基线接续：上传前网络阻断

- 本轮不复用本地副本作为交付。用户已授权共享`element`保留00501/00502 TEST基线；fresh只读确认active卷0、005占用0、002来源各1/140/140/700、现有report三表1/1/1，且reissue/draft/formal表均未安装。在线PID2746、backend4179…、front52eec…、health和冻结002 58.642639保持。
- 已创建受限全库/应用与报告资产备份`mng005_shared_baseline_c51130cb775ba019`，DB/application/report-assets SHA分别为068cd193…/f998399c…/97fb3b30…；root0700/files0600且gzip/tar/SHA门禁通过。命令尾部三个工具权限/CRLF失败如实保留，不影响已核备份字节。
- 新one-shot Go test harness只运行生产service调用，不启动HTTP/Worker；计划直接创建canonical profile以避免共享安装draft/reissue/formal DDL，成功后在现有revision/current/audit和私有报告根保留两套TEST链，失败自动exact-owned清理。Linux binary SHA7bbc8ec2…、compile/vet0，rollback脚本bash-n0；gofmt差异仍未关闭。
- 上传首次与唯一重试均在SSH连接前timeout；SCP未成功、远端harness命令未启动，共享库写/文件创建/服务操作/部署/production均0。状态为`BLOCKED_NETWORK_BEFORE_UPLOAD`，不沿本地旧PDF或identity证据冒持久共享基线完成。完整记录见[共享staging执行记录](management-traits-005-shared-baseline-staging-20261008.md)。

## 2026-10-08T16:20Z identity-v2 fresh重跑三次有界停止；隔离数据已精确回收

- 目标仅本机loopback＋owned副本。写前`VerifyRuntime`通过copyOnly/source1142/crossFK0/report schema/Redis ownership。新E2E源码已把不可事后重建证据顺序固定为：两产品真实UI完成→下载PDF→不可变core receipt（含原始synthetic exam/run/report ID，仅内部证据）→hash-only v2 commitment→cleanup manifest→exact cleanup；commitment绑定product、formula SHA、producer script SHA、core receipt SHA、title/exam/run/report hashes、name/gender/phone hashes和PDF SHA。
- 三次运行上限均未取得业务完成：第一轮终端readline handoff未被子进程消费；第二轮marker在native filesystem可见但脚本未消费；第三轮改由脚本轮询自身浏览器getInfo，用户确认外部Chromium显示后台首页，脚本仍未得到verified admin结果。三轮都没有生成、查看或下载新报告/PDF；未伪JWT、复制token、reset账号或绕过权限，停止后不再第四次改/跑。
- 三次均在停止前各创建了一个00501 frozen/candidate/paper前置链。恢复工具先只读证明恰3条、标题严格`MNG005-V2-<13位时间>-A`、00501、profile/candidate/paper各3且report/audit0，再按FK顺序事务删除及清理孤立bundle；回收后exam74→71、paper1492→1489。三个仅空目录/登录marker的本轮结果目录也按exact路径清理；无新PDF可保存。
- PDF oracle默认门禁现为单object16MiB、object text+raw+decoded合计128MiB；真实已知10,688,445-byte合法decoded流进入合同正例，16MiB+1和总量+1均失败。identity公式12个独立分量逐项mutation均改变commitment；canonical两源码/四AST node合同保持GREEN。没有重跑旧PDF，历史SHA和blocked v2 receipt保持。
- 最终`VerifyRuntime`重新通过，DB check为81表/71exam/1489paper/14管理特质表/2个005 repo，四listener18080/18092/23316/23317均loopback；服务继续存活。中间一次Verify BLOCKED已证为launcher给旧strict decoder无条件发送新字段，限定到新cleanup模式后恢复，不是数据库/Redis隔离漂移。
- 最终验证：PowerShell parser0、E2E syntax0、canonical+mutation+budget合同native0、runtime helper build/vet0、runtime tests16/16。当前结论是`BLOCKED_AUTH_DETECTION`，不是identity-v2 complete；不返回不存在的新PDF SHA/bytes/pages。

## 2026-10-08T14:43Z 三项复审：AST/PDF扫描完成，full identity commitment证据不足

- 仅C区工具/测试/docs与既有PDF只读；无UI、DB、report generation、service、remote。两PDF SHA保持`0850bc502a61009ed47335c75b20384fa19f2bbd9c082275d6064d186c1eed87`/`9d925561bcd85470e074fd4b35e30937cfc60147cc4d40116c219ff5dab64aec`。
- Go exporter v2绑定identity+scoring源码SHA和dimension catalog/builder/scoring function/module aggregation四节点位置/hash；mutation合同拒绝decoy symbolic/partial alternate mapping和模块顺序漂移，benign unrelated strings接受。模块仍self3/interpersonal4/task3/development3。
- PDF configured oracle两份各933 xref objects/41 streams，合计object text254372、raw1030825、decoded31595162 bytes；metadata22、trailer/catalog各2、attachment0、needsPass/encrypted0、credential findings0。最大合法decoded font stream10,688,445 bytes，故8MiB单对象门禁真实fail、配置16MiB单对象/64MiB总量GREEN。
- identity formula v2已完整定义；当前producer/formula/safe receipt SHA为`3a01e826…/5e2d0b54…/cc6f722d…`。旧browser receipt只保存reportIdHash，未保存runIdHash；PDF没有UUID，DB/report已清理。existing模式准确BLOCKED两份runIdHash、records空，不生成final commitment。需要未来获准重跑E2E并在cleanup前保存run hash，不能后验推造。

## 2026-10-08T14:33Z CodeReviewer四个PDF oracle blocking关闭

- 修改前独立RED收据锁定两PDF SHA并确认旧oracle缺`canonicalSourceSHA/identityCommitmentSHA/pdfObjectCredentialScan/recursiveReceiptCredentialScan`。本轮不重跑UI、数据库、报告生成或服务，不修改两PDF。
- 新C区Go AST工具严格读取真实`management_traits_identity.go`，从私有catalog导出`ManagementTraitsDimensions()`实际消费的13维、140题、100正40反及模块成员；source SHA=`4b417188b8cb362f27620a3b02029d77b2f6b78016091b7a83c76560d0763179`、function SHA=`519800574134b675146579713428e682d82de61289ac45e791421d035f1c527b`。模块为self3/interpersonal4/task3/development3。
- E2E源码新增无Playwright副作用的`--identity-contract-existing`模式，读取既有receipt/PDF并只输出hash。00501/00502 commitment分别为`eb8af41a113ecbffe83414a8427527ec5d972bf88ae873855d46c15f6673b319`、`7d5d5bde663771f99090018f5540d4aefcf2f6d920def7d46f44e2605ce500cc`；oracle从PDF标签提取值后重算每份name/gender/phone三个field hash与总commitment，每份手机号集合恰1。
- JSON扫描按解析后key自然处理Unicode转义，递归对象/数组；疑似Base64最多2层、单次解码64KiB并仅扫描UTF-8高可打印内容。三安全收据findings0。两PDF扫描正文、各11项metadata、各933个xref object string和embedded attachment名称/内容；附件0、findings0、外链0。
- Node contract exit0、Python AST/oracle exit0、四文件diagnostics0。最终oracle contract SHA=`6672c8b6…`、identity commitment file SHA=`a13cd4e7…`；PDF SHA仍`0850bc502a61009ed47335c75b20384fa19f2bbd9c082275d6064d186c1eed87`/`9d925561bcd85470e074fd4b35e30937cfc60147cc4d40116c219ff5dab64aec`。

## 2026-10-08T14:31Z CodeReviewer 两个 PDF oracle warning 关闭

- 仅修改独立Python oracle和证据文档；未重跑UI、未读写数据库、未生成报告、未启动/停止服务，两个既有PDF字节未变。
- 独立RED收据确认旧oracle缺少每份报告的`independentModuleScores`和`privacyVerdict`。修复后从140个raw=3按40个反向题仅映射一次，先独立得到13个精确维度分，再按canonical Go模块映射以有理数等权聚合；`self/interpersonal/task/development`四键各一次且均为精确50，不再读取浏览器收据的模块数量作为证明。
- synthetic-only门禁从实际E2E源和安全收据固定两组已批准姓名/性别/手机号，PDF可提取身份必须精确匹配、单位/职务必须为空；两份11位手机号集合必须是固定synthetic allowlist子集。PDF和三份安全JSON收据均扫描JWT样式、JWT/Bearer/Authorization/password/口令/密钥，发现数0；每份synthetic身份标记2个（姓名+手机号），不输出真实私密数据。
- 配置的venv Python AST通过；聚焦oracle两份均`PASS_LOCAL_SYNTHETIC_ONLY`。模块值均为50，身份标记各2、手机号各1、privacy/receipt credential findings均0；编辑器诊断0。PDF SHA仍为`0850bc502a61009ed47335c75b20384fa19f2bbd9c082275d6064d186c1eed87`和`9d925561bcd85470e074fd4b35e30937cfc60147cc4d40116c219ff5dab64aec`。

## 2026-10-08T14:18Z 00501/00502客户模板报告真实本地E2E完成并清理

- 仅本机回环和独立副本：启动时原PID20036仍健康，未因上次`RunBackend` exit1误重启；为启用本任务明确批准的独立reissue能力，先双执行copy-only 004 DDL并验证幂等，再精确停止owned PID20036、重建后端并启动PID9540。隧道19132、Redis12324、Vue7912均保留，四listener仍127.0.0.1；direct/proxy health均ok。无staging共享`element`写、无远端服务操作、无部署/生产/历史PDF修改。
- runtime只开放五个`report-reissues`路由，未登录仍401；旧`reports`和generic PDF路由仍503。approved XLSX/DOCX分别强制SHA `b0498249…`/`05c55e77…`，LO固定本机`soffice.com`，私有运行报告根在owned `tmp/mng005-debug-.../reports/reissues`。004首轮/重跑均exit0，verify为copyOnly/source1142/crossFK0/reportsEnabled=true。
- 新建并冻结两个exact synthetic开放测评：00501 exam `1791466773929935301`、00502 exam `1791466835021877003`。两者均通过真实管理UI Save+独立freeze、真实candidate登记/准备/开卷、140次可见选项`3一般`保存和manual交卷；00501在70题时真实reload恢复70/第71题。两产品实际结果均completed、140/140、13维、4模块、总体50.00；独立canonical oracle以140×raw3、40反向一次推导13维及总体精确50，不复制数据库分值作期望。
- 真实管理UI详情均13/4；生成/重复生成、查看和下载均只用new reissue按钮。00501首次沿先前本轮生成结果复用，00502首次新生成；两者重复均复用同report。独立正常Chromium窗口由用户正常登录，无token/storageState/cookie导出；原生Download事件成功保存两PDF到ignored结果目录。
- 发现并修复真实产品bug：后端purpose固定为“仅供系统测试，不可作为人才决策依据”，前端错误要求包含英文`TEST`，导致HTTP200/eligible=true仍拒绝。实际浏览器RED；SFC回归RED 1fail→最小精确常量修复→管理UI/API 171pass。未放宽eligible/kind、未旧报告fallback。
- 00501 PDF 653766 bytes/SHA `0850bc502a61009ed47335c75b20384fa19f2bbd9c082275d6064d186c1eed87`；00502 PDF 653523 bytes/SHA `9d925561bcd85470e074fd4b35e30937cfc60147cc4d40116c219ff5dab64aec`。独立PyMuPDF/openpyxl oracle两份均A4 9页且每页非空、TEST/免责声明、13维名、36条精确客户文案、5环+13柱+1常模折线、4个客户星级字形、0 placeholder/U+FFFD/外链；现有value-only Word回归另确认footer policy和非值OPC不变。下载SHA/bytes与view响应、DB metadata及私有文件逐字节一致。
- 持久化前清理证据：每产品report1且原`data_snapshot` UTF-8字节SHA与DB SHA2一致；00501 audit generate/reuse/view/download=`1/3/2/2`（包含集成浏览器诊断重试），00502=`1/1/1/1`；FileVerified均true。随后hash锁定脚本按exact exam/title和RESTRICT顺序删除报告审计→报告→结果/卷/人员/测评，并删除两份runtime私有PDF；最终两个code report/audit/private file均0。
- 保留项：下载的两份synthetic PDF、browser/DB/oracle/cleanup receipts；00501/00502 fixture仍各repo1/关系140/题140/选项700，draft/reissue三表保留。冻结002仍completed1/58.642639及manifest/input/mapping/mapping原字节/field-contract五SHA全同。后端PID9540、Redis、前端和隧道继续运行。
- 失败保留：集成浏览器无法捕获Blob anchor Download事件，改用独立正常Chromium后成功；独立browser首轮登录后`getInfo`监听订阅过晚timeout，业务请求0，修harness后重跑GREEN；PDF oracle首轮误将SPECS模块键当维度名而失败，改从批准工作簿取13名称后GREEN。未执行production/staging写、正式批准、旧PDF重出、25分钟自然/race或真人数据。

## CodeReviewer 最后 atomic commit blocking 关闭

- 仅修launcher原子密文写和本地合同：既有目标在任何temp/序列化前验证private ACL与nonreparse；temp在commit前验证private ACL、nonreparse、DPAPI解密、JSON及全部预期字段。`File.Move`/`File.Replace`是最后可抛出的提交语句，之后仅`committed=true; return`；finally成功路径不执行操作，失败路径只静默清理temp。
- Windows真实合同验证首次Move和Replace结果均保持当前SID-only ACL；commit相邻hook抛错时旧SHA不变/temp0，缺字段及不安全目标均在hook前拒绝。两个PowerShell parser为0、合同native0；真实`VerifyRuntime`通过，`runtime.dpapi`前后SHA均为`45FA2E5D5C9756A7B89D6455C7788CAFC7EB6916ABE083B218A8D6BEA4CE29E3`。未执行DB写、UI、remote或restart。

## 2026-10-08T13:20Z CodeReviewer 后续 2 blocking + warning 真正关闭

- **纠正13:11记录**：当时launcher仍会在正常同步中重写Redis PID/executable/config path/config SHA，且`runtime.dpapi`直接写入；因此“两个blocking关闭”只覆盖Go运行门禁，不足以证明持久化原子性和绑定不可变。本节追加修复后，正常prepare/verify/inspect/backend路径只允许五项绑定全部缺失时首次绑定；任一部分缺失或已绑定值不符均失败关闭且不改密文。只有人工显式`RebindRuntime`可在既有目录/文件ACL、nonreparse和live ownership门禁通过后轮换五项绑定，三个秘密字段原样保留；本次未对真实runtime执行rebind。
- `runtime.dpapi`改为同目录唯一私有密文temp：CreateNew、UTF-8无BOM字节、Flush(true)、仅当前SID ACL、nonreparse、DPAPI解密/JSON/预期字段校验后，已有目标用`File.Replace(temp,target,null,true)`，首次创建用无覆盖`File.Move`；失败保留旧目标字节SHA并只删除本次temp，不产生明文或backup。launcher改为可安全dot-source的函数入口，生产模式没有注入失败开关。
- 有效RED：新PowerShell合同先因`Write-AtomicProtectedJson`不存在失败。GREEN同时证明注入commit前失败时target SHA不变、temp精确清理、same binding接受、stale拒绝不重写、显式rebind只换绑定、全缺失首次绑定、部分缺失拒绝、密文可解且ACL私有。首个未跟踪GREEN尝试曾在合成`File.Replace`报“Unable to remove the file to be replaced”；独立Replace探针、带跟踪完整合同及随后普通合同均通过，失败不覆盖。
- FB-202新增Windows显式opt-in实际inspector测试；默认suite跳过，launcher只经环境传PID/port/exe/config path/config SHA等非秘密元数据，测试真实listener/process/config ACL/nonreparse/read/SHA，不传Redis密码。最终Go package test/vet、PS合同、actual inspector、`VerifyRuntime`全部exit0；verify为copyOnly/source1142/crossFK0/redisOwnershipVerified，Redis PID12324、config SHA仍`4c9b12f09d7a154f89d4b0c32e03a228e793d44a000de99f12f0e1c192a9eafa`。
- 终验不重启：backend PID20036仍监听127.0.0.1:18092，Redis PID12324仍存活；direct/proxy health均HTTP200/status ok，匿名Redis仍NOAUTH。真实`runtime.dpapi`验证前后字节SHA相同。未跑280 UI、未远端、未部署、未写DB。
- 最终源码SHA-256：launcher `5682a83fae527851c606b075856a5dacace5e9392287a4f5f1567aebb1314e9f`；PowerShell合同 `5b749ff7b7f8b2ac49bd345e566f695b860de4b39cde25f599b80a5be9cfd3fe`；Go inspector测试 `388dbf8cf8dc4134ad57efa6a369f0bc5e94df94a6fd2333ac0fbfd36939c66d`。最终gofmt语义对比、PowerShell parser、编辑器诊断均无错误。

## 2026-10-08T13:11Z CodeReviewer 两个 Redis runtime blocking 本地关闭

- **仅本地dev runtime；未部署、未远端DB写、未重跑280项UI。** 两个阻断分别为：启动只证明Redis密码可用而没有证明监听进程/配置归属；短时verify与正常backend启动没有共享同一强校验。新增回归先因 `redisRuntimeBinding` / `verifyRedisRuntime` 不存在而编译失败，再以同一验证器覆盖verify/inspect/fixture/cleanup/backend全部入口。
- DPAPI runtime元数据新增并持久化 `redisHost/redisPort/redisPid/redisExecutablePath/redisConfigPath/redisConfigSHA`。launcher每次启动前用精确 `127.0.0.1:23317` listener、CIM进程路径、无秘密的config路径和当前完整配置SHA刷新/核验绑定；配置仍为工作区外仅当前SID ACL且非reparse，不输出配置正文、密码或进程命令行。
- Go入口在连接router/Worker前独立核验：host必须127.0.0.1、exact port/PID、netstat listener PID、实际进程可执行文件、私有配置ACL、非symlink普通文件、完整SHA、唯一bind/port/protected-mode/databases/requirepass合同；密码使用常量时间比较。随后必须认证PING和DBSIZE成功，DBSIZE可为非负运行态值。任一host/port/PID/listener/executable/ACL/SHA/secret/auth/DBSIZE错误均失败关闭。
- 实际运行第一次因PowerShell子进程参数传递不符合预期而安全拒绝，未启动第二后端；改为整数PID白名单内联、config路径仅子进程环境传递后通过。最终Redis PID12324、listener/executable/configSHA/ACL/nonreparse全部match，认证PING通过、DBSIZE1；config SHA=`4c9b12f09d7a154f89d4b0c32e03a228e793d44a000de99f12f0e1c192a9eafa`。
- 仅精确停止旧owned backend PID2736，重建并启动PID20036/127.0.0.1:18092；Redis、Vue和MySQL隧道未重启。package test、vet、verify/inspect/backend build、launcher parser均exit0；重启后verify再次exit0，direct/proxy/captcha均HTTP200、验证码有效、匿名Redis仍NOAUTH，未提交登录/生成报告。[机器收据](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/runtime-redis-ownership-20261008.json)。

## 2026-10-08T13:05Z 00501/00502 隔离副本真实浏览器闭环完成

### 失败、修正与未覆盖边界

- 首次手工 `fetch('/dev-api/getInfo')` 未加应用实际使用的 Bearer 头，返回 HTTP 401；该结果不能证明会话失效。清除旧本地 cookie 并由用户在当前页正常登录后，改为监听页面 reload 发出的真实 `/getInfo` 请求，实际 HTTP 200、业务码200、admin=true、`*:*:*`=true；未导出或打印 JWT。登录页浏览器快照工具曾显示浏览器既有“记住密码”值，后续没有读取、复制、落盘或在文档中记录该值。
- 首次 `VerifyRuntime` 因独立 Redis 已有合法验证码/登录键而返回 `LOCAL_RUNTIME_BLOCKED`；隔离库、源拒绝和跨库FK没有失败。新增 RED 测试先以缺失 `validateRuntimeRedisState` 编译失败，再令 verifier 接受非负私有 Redis DBSIZE，同时继续逐次核验 owned schema、CURRENT_USER、源 `element` SELECT=1142、跨库FK=0、Redis认证和回环绑定。最终 `VerifyRuntime` exit0，redisKeys=1；不再把“运行中缓存必须永远为空”当隔离条件。
- Playwright逐题串行驱动的前三次循环均只完成一题后等待下一响应超时；每个已收到的响应和页面进度一致，没有重复写或错题。改为浏览器页面内按“当前已答数变化→按钮重新可用”驱动真实选项按钮，两个产品最终都140/140并真实交卷完成。失败不覆盖。
- 本轮不生成、查看或下载报告：开发入口仍对 report/pdf/template 返回503，正式审批环境仍关闭。没有部署、重启 staging 应用、写共享 `element` 或访问 production。

### 副本结构与可复用005资产

- 写前强制门均通过：真实应用 `/getInfo` 管理权限通过；`VerifyRuntime` 为 copyOnly/source1142/crossSchemaFK0/loopback/reportsDisabled；冻结002 exam `1791298091700970647` 仍为 completed1、profile1、overall `58.642639`，manifest/input/mapping/field-contract SHA与本报告既有基线相同。
- 实读副本源产品均为唯一repo、140关系、140题、700选项；00501/00502均为0。系统只有题库/题目逐项保存接口，没有题库克隆接口。因此新增并实际执行哈希锁定的 [005 fixture SQL](../scripts/sql/management_traits_005_local_fixture_7081fbec31e3d105.sql)，只允许 schema `talent_mng005_local_7081fbec31e3d105`，以独立PK前缀从00201/00202复制，不修改002，不关闭外键，不做宽泛删除。
- 同次只在副本安装 [draft migration 003](../scripts/sql/management_traits_003_new_draft.sql)。执行后 `el_mng_exam_draft`=1张；00501与00502各为repo1/关系140/题140/选项700，ownership marker=`MNG005_LOCAL_FIXTURE:7081fbec31e3d105:v1`。00201/00202计数保持各1/140/140/700。
- 005 fixture 作为后续隔离本地调试可复用产品保留；两个临时测评、人员、试卷、快照、答案和结果不保留。

### 真实本地浏览器与数据库证据

| 产品 | 管理端 | 参与者链 | 完成态数据库证据 |
|---|---|---|---|
| 00501 | UI选择真实00501 repo；Save HTTP200/code0/lifecycle=draft；独立“确认冻结 TEST” HTTP200/code0；配置变只读 | 开放登记→准备页→创建同一paper→答70题→reload后仍为answered70/第71题→答满140→手工交卷 | profile1/candidate1/paper1/snapshotQuestion140/legacyQuestion140/answerBucket700/run1/dimension13/module4/receipt1；`mng-00501-v1` + `mng-00501-db-current-v1`，questionnaire=staff，completed，overall=`29.444146` |
| 00502 | UI选择真实00502 repo；Save HTTP200/code0/lifecycle=draft；独立冻结 HTTP200/code0；配置变只读 | 开放登记→准备页→创建paper→真实UI选项答满140→手工交卷 | 同样为1/1/1/140/140/700/1/13/4/1；`mng-00502-v1` + `mng-00502-db-current-v1`，questionnaire=leader，completed，overall=`50.000000` |

- 浏览器只返回状态、业务码、是否存在ID、进度和完成状态；文档不记录参与者ID、paper/run ID、JWT、姓名、手机号、答案顺序或快照JSON。
- 完成态证据采集后执行哈希锁定的 [精确 cleanup SQL](../scripts/sql/management_traits_005_local_e2e_cleanup_7081fbec31e3d105.sql)。脚本先要求两个精确exam标题、2 profiles、2 candidates、2 papers、2 completed runs/receipts且报告实例0，再按已实读RESTRICT FK子→父删除；未禁用FK、未按前缀/时间宽删。
- 清理后两个exact exam的profile/candidate/paper/snapshotQuestion/legacyQuestion/answerBucket/run/dimension/module/receipt均0；005 fixture仍各1/140/140/700，draft表保留。冻结002总体与五项SHA再次完全相同，源SELECT仍1142拒绝。

### 本地运行状态

- MySQL隧道、独立Redis、旧PID2736后端和Vue前端继续保留。为修 verifier 曾关闭其父terminal，但旧后端子进程仍持有18092；尝试启动新后端被端口占用明确拒绝，未形成第二监听。新 verifier/inspect/apply/cleanup均为短时独立进程且exit0。
- 辅助运行入口新增 `ApplyFixtures`/`CleanupE2E`，仅消费工作区外DPAPI配置并调用固定独立二进制；SQL文件名、schema、stamp和SHA均硬绑定。`server-local.exe`重建成功；当前监听中的旧进程未被替换或重启，但业务链使用的源码版本已包含005能力并完成上述真实E2E。

## 2026-10-08T12:27Z 本机隔离Redis与前后端启动完成（限定启动验收）

### 未验、失败与异常先列

- **未提交管理员登录、未跑真实答题/报告E2E、未注册005、未安装draft/formal/reissue新表。** 开发入口将report/pdf/template API关闭503，正式审批环境关闭；原业务报告源码没有退役或回滚。前一阶段独立CodeReviewer PASS仅按用户交接使用，本次新开发入口未独立复审；未重跑全6400/598/coverage。
- **源库全表摘要漂移，原因未归因**：12:27:32Z源element data-only摘要为 `e6c7251a7fb9a1f64f11587d1cdb8e13f1b9ef7bc2c47807b8a6c6ab0b88bf5b`，不等于10:34的 `c27fdfb8…`。不能宣称源库整库未变、认定具体修改者/表或由本地启动造成；没有回滚、清理或改历史证据。专属账号实际DATABASE/源SELECT1142拒绝/跨库FK0通过，隔离证据不替代整库字节因果审计。
- 首只读SSH255连接前timeout/stdout0，唯一重试exit0/stderr0，失败保留。首匿名Redis探针经PowerShell展开RESP美元符号，观测NOAUTH=false，不当认证缺失；直接脚本PING实际NOAUTH。首live测试错用不存在的captchaEnabled字段native1，实读HTTP与AuthHandler后只改为captchaOnOff，最终native0，不改API。集成浏览器response事件列表为空，不报告完整请求捕获PASS；resource timing主机集合只有本地，DOM验证码可见，真实HTTP另由脚本确认。
- 原密码argv暴露P1历史继续保留，不能称轮换后从未暴露。源摘要异常不以local health、copy hash或主PID相同覆盖。

### 当前开发入口与隔离边界

| 运行项 | 已核事实 |
|---|---|
| 前端 | http://127.0.0.1:18080/#/login ，PID7912；VueCLI开发编译成功24491ms，真实登录控件及验证码可见，未提交登录 |
| 后端 | http://127.0.0.1:18092/health ，PID2736，HTTP200/statusok，只连接owned副本 |
| 独立Redis | 本机已有Memurai4.1.2/API7.2.5，新PID12324/127.0.0.1:23317；不安装系统服务、不改原Redis服务/配置、不复用staging缓存或已有DB索引 |
| MySQL | 原隧道PID19132/127.0.0.1:23316保留；talent_mng005_local_7081fbec31e3d105，专属账号，源SELECT真实1142拒绝 |

- 四listener均127.0.0.1，未公网绑定。Redis protected-mode/随机密码/仅DB0，save为空/appendonly关闭；初始DBSIZE0，不复制源session。验证码验证后12:26实读4个新键，不保持“永远空缓存”假设。
- DB密码及三个运行密钥不进argv/工作区环境文件；已有环境文件仍不存在，未创建覆盖。凭据/新JWT/内部token/Redis密码工作区外DPAPI；Memurai密码配置为工作区外仅当前SID可读**私有明文配置**，不假称其是DPAPI密文。三私有文件ACL/owner/onlySID/nonreparse通过；三个owned进程argv匹配本轮三个秘密次数0，不当历史无暴露证明。
- 文件根B区tmp/mng005-debug-7081fbec31e3d105，upload/reports/export-templates/mbti-full/mbti-simple/redis各自有子目录与当前用户ACL；未拷旧PDF/真实人员文件/源缓存。前端子环境先滤MYSQL/REDIS/JWT/MNG/MSGRAPH/PDFGEN/PHASE1/UPLOAD/VUE_APP和SECRET/TOKEN/DSN/PASSWORD键，再只注公开代理配置；不传DB/JWT给npm/浏览器。
- 原config.Load仍运行，专属入口随后覆盖实际Config的DSN/Redis/JWT/上传/模板/renderer，实际库绑定/1142/FK检查在router及两个原Worker前完成。Graph/Chromium/正式审批/报告关闭，Navbar socket=false；内部源码检索未见SMTP/SMS发送入口，不称任意未来页面外部作用穷尽。GORM/HTTP/raw业务日志仅本dev进程不输出，避免复制PII/SQL泄漏，不改共享日志配置或称全业务错误0。
- 原expiry Worker保持且只能写copy；启动后不要求copy全部数据继续等于源库历史SHA。本机Redis/Go/Vue三终端有意保留：`ebcc3fcc-f72c-4e20-b32f-db8d00396048`、`a15ad16b-b142-4327-9a2a-e3f6aa972c28`、`c146a77c-43bf-455e-90b1-71ba2056aca0`；原隧道不关。未来仅停这些owned进程，不DROP副本/删备份/关闭他人进程。

### 文件归属与C4影响清单（全部闭合）

| 来源与消费方 | 同步修改 / 无需修改及理由 |
|---|---|
| [原launcher及新模式](../scripts/tools/mng005-local-debug.ps1#L1-L146) | 同步修改C区tools：追加RuntimePrepare/RunRedis/VerifyRuntime/InspectRuntime/RunBackend/RunFrontend及可选端口，原Prepare/Check/Tunnel参数/default/协议兼容，读密文前加SID ACL；三次编辑后停止 |
| [轮换脚本唯一Check调用](../scripts/tools/mng005-local-credential-rotate.ps1#L150)及原人工Check/Tunnel | 无需修改：旧参数兼容，原Check本轮native0，原密码stdin回归native0；未重跑Prepare/恢复/轮换 |
| [专属Go运行入口](../Go-based%20Refactored%20System/bin/mng005-runtime/main.go#L34) | 新增B区ignored bin纯运行代码：JSON stdin/精确copy/回环/独立文件根/新密钥/私有GORM/原router-worker装配。未新增共享SERVER_HOST字段、未改业务/API/模型/评分。初始server-local运行产物保留，后加inspect/旧密钥比较编译为独立server-inspect，不覆盖当前backend或冒两binary同字节 |
| [Vue开发wrapper](../Go-based%20Refactored%20System/bin/mng005-runtime/vue-runtime.config.js#L1) ← launcher的MNG_LOCAL_FRONTEND_PORT/MNG_LOCAL_BACKEND_PORT | 新增B区运行配置；两新公开键仅launcher/wrapper/本轮测试消费，验证端口整数/冲突，回环host/open=false/hostcheck及两代理覆盖，异常不回退远端 |
| VUE_CLI_SERVICE_CONFIG_PATH → [已安装真实CLI加载器](../Go-based%20Refactored%20System/ruoyi-ui/node_modules/@vue/cli-service/lib/Service.js#L304-L337) | 复用真实支持的外部config入口，wrapper再读原Vue配置；node_modules不改、不装包 |
| [原Vue配置](../Go-based%20Refactored%20System/ruoyi-ui/vue.config.js)、[package](../Go-based%20Refactored%20System/ruoyi-ui/package.json)、[API base](../Go-based%20Refactored%20System/ruoyi-ui/src/utils/request.js#L16-L21)、[socket开关](../Go-based%20Refactored%20System/ruoyi-ui/src/layout/components/Navbar.vue#L103-L109) | 无需修改：原开发/生产默认行为保持；只本新进程VUE_APP_BASE_API=/dev-api与VUE_APP_SOCKET_ENABLED=false，proxy只127:18092；共享任务/依赖/lock不改 |
| [原Go主入口](../Go-based%20Refactored%20System/cmd/server/main.go)、[config](../Go-based%20Refactored%20System/internal/config/config.go)、[Redis初始化/键](../Go-based%20Refactored%20System/pkg/redisx/redis.go)、JWT/Upload/PDF消费者 | 无需修改：实际Config对象仅dev入口覆盖，旧主监听/默认环境不改，物理实例隔离固定键；同router/service/旧业务保护及005策略保留 |
| [运行契约/live测试](../scripts/test/mng005-local-runtime-contract-test.js)、本报告与记忆 | 同步新增/更新C区test/docs；五实际CLI配置分支/六编译boot拒绝分支/真实健康验证码NOAUTH断言；三次测试编辑停止，不改旧回归或用mock Redis冒真实 |

### 实证与限定结论

- Go运行/inspect build两次0、vet两次0，PSparser0/JSsyntax0/四代码diagnostics0；前端为新CLI开发编译成功，不是另一次production build/全前端测试。11合同分支native0；原P1 bash回归native0/secretInArgvfalse/stdin2/grantsPreservedtrue。
- [真实live收据](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/runtime-live-1791462360506.json)：direct/proxy/captcha各HTTP200、图像valid、Redis匿名NOAUTH、native0，loginSubmitted=false/reportsGenerated=false。
- [运行后copy只读](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/runtime-inspect-639270591616934723.json)：completed1/58.642639/profile1，manifest66277c8b/input7474e6aa/mapping及原字节dc7faac6/fieldcontract051dbe83与原现场同，源1142拒绝；未读姓名/电话/答案/PDF。
- [原511文件SHA保护](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/runtime-source-1791462361946.json)：按上阶段两SFC afterSHA核验511/511同/缺失0，共享业务源无修改；本轮仅新B区运行代码/config及C区launcher/test/docs，未删历史。
- [ACL/argv/四监听](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/runtime-acl-639270592179834535.json)、[首SSH失败](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/runtime-source-stage-readonly-1791462399327.json)、[唯一重试/源漂移](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/runtime-source-stage-readonly-retry-1791462452114.json)均保留。12:27:32Z主PID2746/NRestarts0/active/healthok/backend4179fd3f/front52eec及冻结002分值hash同；**不称source全表SHA同**。
- remote仅只读，无staging部署/主restart/共享Redis写/账号轮换/共享MySQL业务DDL-DML/production/005注册/历史删除。LOCAL_RUNTIME_READY仅本启动切片；用户可在本地页用原账号正常登录，后续由对应开发/测试者在copy开展一个业务切片，新入口独立复审由主协调者安排。

## 2026-10-08T10:34:38Z P1密码argv风险修正与专属账号轮换

- **历史风险保留**：原恢复已执行，含随机密码的 `CREATE USER` SQL曾进入mysql的 `-e` argv；不能报告“没有发生暴露”。本次轮换阻止旧凭据继续复用，不消除历史进程观察、审计或日志风险；没有清除历史日志或证据。当前独立终验general_log=0，不推断原执行时的日志状态或其他审计插件。
- 只改[恢复脚本密码SQL发送](../scripts/db/mng005-local-debug-restore.sh#L56-L65)：关闭xtrace，含密码语句通过heredoc stdin送MySQL；双host、八项副本授权不变。未重跑恢复、重建库或覆盖副本。唯一消费方[原Prepare入口](../scripts/tools/mng005-local-debug.ps1)无需修改：已通过SSH bash stdin发送整段脚本；原Check/Tunnel、Go核验器、Go/Vue业务和前端598基线无需修改或重跑。
- [合成回归](../scripts/test/mng005-mysql-stdin-contract-test.js)：真实本机Git Bash执行原账号片段，以shell函数mock捕获参数及stdin，非真实MySQL。有效RED native1：secretInArgv=true/stdinStatements=0；补丁后GREEN native0：secretInArgv=false/stdinStatements=2/grantsPreserved=true。只使用合成标记，没有真实秘密输出。
- [专属轮换入口](../scripts/tools/mng005-local-credential-rotate.ps1)先核工作区外当前用户SID/仅用户ACL/非reparse/DPAPI绑定及现有23316；远端hostname/root、root0700/0600、精确三行ownership、备份SHA、库存在、只有localhost/127.0.0.1、原八项授权逐字、role/proxy=0均通过才执行。没有仅凭账号前缀推定所有权；positive_app整行及SHOW GRANTS只远端内存作前后相等检查，原账号不旋转、不改GRANT。
- 先保存旧DPAPI**密文**备份和新密码待同步**密文**，均在原工作区外用户私有目录，ACL仅当前用户。新密码crypto RNG生成32个hex字符，仅SSH stdin→远端Python内存→MySQL stdin；无密码argv、base64远端命令参数、环境dump或SQL回显。远端原owned凭据文件更新为root0600，不下载/打印秘密字节；本地env、JWT及共享cache不改。
- 失败如实保留：首次只读门禁误按单引号/权限排序比较，被真实MySQL backtick/原生排序拒绝，DDL0；校正后精确授权SHA `cec61505d0a52d24a5d8e7be25f6cb6528803e1bb9755cac129982e138d14599`。一次本地PowerShell字符串展开SyntaxError发生在SSH前。实际远端轮换及前后校验通过后，File.Replace的PowerShell5 `$null`绑定成空字符串失败；新密码仍已生效、待同步密文保留，没有重复轮换或回滚到旧密码。用NullString明确传.NET null，仅验证并原子同步该既有密文；入口第三次编辑后停止。
- 真实新凭据Go连接native0：SELECT DATABASE()/CURRENT_USER匹配副本、78表/71exam/1489paper/11mng、crossSchemaFK0、SHOW GRANTS仅USAGE+副本，源element SELECT真实1142拒绝。旧活动密文的连接native1/BLOCKED；只将其记为旧凭据拒绝，不伪报未采集的1045。同步后原Check再次native0，活动密文及备份ACL仍仅用户，pending已原子消费。原23316隧道保留，未新建监听或kill。
- [新独立终验收据](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/credential-rotation-final-1791455678358.json)exit0/stderr0：源/副本全部78表data-only SHA仍 `c27fdfb8e235ed36923e68297419a3f4f6731ec49564620f438bb57e11140b5a`；源完整schema/data与原source-before基线同，runtime资产/旧PDF/cache与原assets-after基线同。双host原八项授权/role0/proxy0，PID2746/NRestarts0/三服务active/healthok，backend4179fd3f…/front52eecf04…保持。原备份和全部旧收据不覆盖。
- 仅专属副本账号密码DDL及其受限凭据同步；源element业务DDL/DML0、原positive_app改动0、全局授权/配置0、迁移/seed/缓存写/应用启动/部署/restart/生产0。业务后台及Redis隔离仍未完成，独立CodeReviewer仍由主安排，不冒本次自核为独立审阅。

## 未验证项与保留边界

- **业务后台未启动，未进行真实本地登录/答题/报告E2E。** 本轮完成的是真实MySQL副本恢复、最小Go连接核验，以及登录SFC修复和本地回归，不把DB Ping冒充应用health。
- 当前完整应用初始化必须Redis Ping，缓存键前缀固定；主入口仅支持端口、监听全网卡。没有配置隔离Redis、没有安全回环监听配置，因此不复用staging Redis、不启动业务后台或前端代理。后续须独立配置这两项后才能启动全应用。
- 未执行005注册、题库复制、draft/formal/reissue迁移、合成管理员创建、真人报告生成或历史重算；副本保留源库当前结构。独立CodeReviewer仍由主协调者安排，本worker没有自行宣称审阅PASS。
- staging应用、共享业务库原记录/DDL/全局配置/原账号权限/Redis/防火墙/NSG零修改；生产未访问。没有git restore/clean、全量格式化、删除客户材料/旧PDF/失败证据。新owned副本、专属账号、备份和回环隧道保留用于持续调试；未来清理须按精确所有权另确认。

## 本轮代码与调用影响

开工保存[511项原生源码/单测SHA](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/source-before.json)，结束[509项相同、仅目标SFC与单测变化、缺失0](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/source-after.json)。所有既有Go业务、其他前端源码及原005切片保留，不还原21项未归因旧漂移。

| 消费位置 | 同步修改/无需修改及理由 |
|---|---|
| [封闭登录SFC](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/tester.vue#L59-L107) | 同步修改：移除URL/session提前return；始终读metadata；仅同exam服务器分类授权新版登录，pending/失败拒绝。合法冻结002/005跳过客户端旧窗口判断，由后端原续答门禁决定 |
| [实际SFC回归](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-participant.spec.js#L240-L267) | 同步修改：5普通产品带URL意图仍旧登录、2未冻结005、4跨exam/非法ID、stored-token pending/failure；保留已有冻结002过期后走后端资格测试，按实际metadata合同更新夹具 |
| [普通002与005保存分类](../Go-based%20Refactored%20System/internal/service/management_traits_draft.go#L135-L169)及[真实Save测试](../Go-based%20Refactored%20System/internal/handler/exam_management_traits_draft_test.go#L114) | 无需修改：当前005切片已恢复普通002缺flag/false旧保存；请求repoCode不能覆盖真实DB code；普通002显式新版请求仍拒绝并提示005 |
| [新旧配置表单](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L127-L147) | 无需修改：普通002无新建opt-in，005强制新版；既有002draft/frozen兼容，不能把旧配置转换005 |
| [旧路由保护完整清单](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_guard.go#L27-L90)、[作用域分流](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_guard.go#L372-L439) | 无需修改：按真实profile/paper/marker保护，不以002编码全禁；覆盖创建、保存、身份、答题、分数、报告、导出和列表，精确draft准备例外仍需真实管理员。未撤已上线身份strict白名单 |
| [旧002选项原始分](../Go-based%20Refactored%20System/internal/handler/paper.go#L611-L630)、[旧标准分金标准](../Go-based%20Refactored%20System/internal/handler/tester_score2_formula_test.go#L11-L45) | 无需修改：旧实际选项score保存及原13维公式保持，聚焦测试已执行；非002原评分同样保留 |
| [原002服务端报告](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L236-L269)、[旧导出列](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L300-L309)、[人员导入](../Go-based%20Refactored%20System/internal/handler/tester_excel.go#L111-L134) | 无需修改：原002 report/result2、13维导出与封闭导入路径仍在；新的draft保护只按实际实体作用域处理。未找到独立exam clone/import路由，不伪造接口；客户端复制/另存仍经原Save |
| API wrapper、token/identity/resume、规范化manifest/mapping/hash、旧PDF、001/003/00401/MBTI | 无需修改：公开签名、返回结构、版本合同不变；511项SHA及focused/full frontend验证保护，不声称本轮真实全业务E2E |

## TDD及本地证据

- 修复前实际SFC **420pass/13fail/native1**：[RED](../scripts/test/results/mng-005-isolation-local-20261008/front-focused-1791454506436.json)。修复后 **433pass/0fail/native0**：[相关8文件GREEN](../scripts/test/results/mng-005-isolation-local-20261008/front-focused-1791454557339.json)。新增12例，原冻结续答案例改为先读真实冻结metadata，不删负向。
- 后端普通002/005/旧guard/旧原始分及标准分/冻结resume聚焦 **440pass/0fail/0skip/parse0/native0**：[真实事件摘要](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/backend-focus.json)。另原005最终35/0/0通过；没有重复全6400或冒充本轮全Go。
- [server build](../scripts/test/results/mng-005-isolation-local-20261008/go-build-1791454773622.json)、[all build](../scripts/test/results/mng-005-isolation-local-20261008/go-build-all-1791454782115.json)、[vet](../scripts/test/results/mng-005-isolation-local-20261008/go-vet-1791454787665.json)各native0，stdout/stderr0。最小DB核验器build/vet0，PowerShell parser0；编辑器目标诊断0。
- 本轮只一次[前端全量34files598pass/0fail/0pending/native0](../scripts/test/results/mng-005-isolation-local-20261008/front-all-1791454812585.json)，[生产模式构建native0/Build complete](../scripts/test/results/mng-005-isolation-local-20261008/front-build-1791454836700.json)。Browserslist既有提示及测试既有warning保留，不升级依赖；构建dist仅本地产物，没有上传。
- 全新编译UI浏览器及覆盖率未执行；此前28场景浏览器为旧切片历史证据，不拿它证明本次封闭登录修复。
- 首次报告链接终验native1：Save测试链接结束行超出实际文件，仅缩为已核定义位置后重新验证；不是业务测试失败，不覆盖该失败。构建收据按JSON内精确phase选择，避免go-build文件名前缀误取go-build-all。

## 独立恢复副本：已实际执行

- 目标仅20.200.136.133 staging，strict/Batch/ConnectTimeout10/Attempts1，已核hostname vm-ubuntu-go-dev及liming；sudo只用于受限备份/新库恢复/新账号，不用root账号给本地应用连接。
- 唯一owned库：`talent_mng005_local_7081fbec31e3d105`。源库element没有执行DDL/DML。备份服务器内部受限目录 `/opt/talent-assessment/backups/mng005_local_7081fbec31e3d105`，ownership标记固定；目录0700、全部文件0600，未下载真实身份/答案/源库密码。
- 新single-transaction/quick/skip-lock备份SHA **51d0d45caf52cbad293c4cb7b7006f3b04ff1412a6af83b259b01095eeb2610c**，gzip/清单验证通过。仅备份表结构与数据，显式排除routines/triggers/events/GTID设置；原备份不改，私有恢复流移除实例SET并预检禁止数据库切换和qualified FK。FK session指令仅恢复副本连接，主库不关闭FK。
- 恢复 **78表、71exam、1489paper、11张el_mng表、跨库FK0**；原结构复制，没有新迁移。源库与副本00501/00502真实repo数量均 **0**，未创建假条目，不称新版可实际创建。
- 源库全表结构/数据dump摘要与运行资产/旧PDF/源缓存前后逐SHA相同；独立再次读取全部表data-only有序dump，源库/副本一致SHA **c27fdfb8e235ed36923e68297419a3f4f6731ec49564620f438bb57e11140b5a**。不是仅表数相同。
- 新专属随机账号只localhost及127.0.0.1，只有副本SELECT/INSERT/UPDATE/DELETE/CREATE/ALTER/INDEX/REFERENCES，无global/GRANT OPTION；没有改positive_app任何权限。副本允许后续测试写入，本轮连接核验是只读，不伪造已完成写测试。
- [实际恢复收据](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/restore-receipt.json)、[独立最终只读证据](../scripts/test/results/mng005-local-debug-7081fbec31e3d105/readonly-final.json)。10:21:06Z主PID2746/NRestarts0/active/healthok，磁盘backend4179fd3f…/front52eecf04…保持；本轮deploy/restart0，不沿先前历史硬假定PID/hash。

## 现场冻结002保留：本轮fresh实证

指定exam1791298091700970647，源库与副本均只有1个completed run，总体 **58.642639（两位58.64）**。源库未重算或改码，全部data SHA一致。

| 原冻结字段 | 源库及副本相同SHA |
|---|---|
| scoring_manifest_sha | 66277c8bd4f2d51fc2ab51020625e6fe27b8df10ae02381bdb023b5773e97e0d |
| input_sha | 7474e6aae40e55f32b6f4f2c415ca9da28857ec13ad668096b49525e732c463c |
| mapping_sha及原mapping_snapshot字节SHA | dc7faac623b10bc42c4e39b493518469ef0c14e586b35e6d0f7ddaafbd1e58bf |
| 原field_contract字节SHA | 051dbe83d91a184e273ae6ce72748952c44cad4e4163ec82dcf6bf3c83aa2b6b |

本轮只查数量/分值/hash元数据，不下载participant_snapshot、姓名电话、答案JSON、原PDF或JWT；不是合成golden替代现场hash。

## 本机连接及后续使用

- 已保持SSH隧道 **127.0.0.1:23316 → staging 127.0.0.1:3306**；本机listener核验仅127.0.0.1，没有公网MySQL或网络策略改动。
- [本地连接入口](../scripts/tools/mng005-local-debug.ps1)的 `Check` 已真实执行：`LOCAL_COPY_CONNECTED`、`schemaGrantOnly=true`、`originSelectDenied=true`、`crossSchemaFK=0`、native0。MySQL真实CURRENT_USER绑定专属账号，SHOW GRANTS只USAGE＋唯一副本授权，主element SELECT返回1142。
- 凭据在工作区外当前Windows用户DPAPI加密保存、ACL禁止继承且仅该用户；原本本地env不存在，本轮不创建/覆盖、无全局进程环境修改、结果目录没有凭据/DSN。连接核验器通过stdin临时解密消费，不输出密码或原错误。
- [受限恢复入口](../scripts/db/mng005-local-debug-restore.sh#L1-L77)及[最小Go核验器](../Go-based%20Refactored%20System/bin/mng005-dbcheck/main.go)仅本次辅助；后者源及二进制在B区ignored bin保留，不冒充正式应用入口。
- 当前隧道terminal引用721561fb-1f81-4e7f-91b1-c2393eb0adf5，仅供主协调者管理；不要关闭他人SSH或自动DROP此副本。完成本地开发后再确认精确清理。
- 下一由主协调者安排CodeReviewer只读复审本次两前端文件及辅助隔离方案；再独立处理本地Redis隔离及回环应用启动。005注册/来源审核/迁移仍是后续独立任务，不沿本次授权发布staging或生产。