# Business Branches

## STAGING-FULL-VALIDATION — 2026-10-09 GREEN

| 分支 | 状态 |
|---|---|
| 00401发布/90题/恢复/提交/结果 | ✅ 2组、10维、80+10题、90/90保存、manual与重复提交 |
| 00401导出/正式v2报告 | ✅ 三Sheet结构、批准DTO、10页PDF、2审计；exact cleanup恢复非空基线 |
| MBTI答题/评分 | ✅ 48题全部保存、ESTJ评分与回读一致，无效AB和守卫生效 |
| MBTI完整版/简版报告 | ✅ 两类PDF真实下载；16+16模板；匿名模板拒绝 |
| MBTI LibreOffice超时/失败 | ✅ FB-218共享有界客户端，失败不再把DOCX冒充PDF |
| 00501/00502结果与报告 | ✅ 各1 completed/140/50/13维/4模块，新模板SHA报告view/download200 |
| 005测试报告清理 | ✅ 删除本轮2个revision/audit/PDF，恢复两份旧current指针，报告总数回3 |
| 00101只读兼容 | ✅ detail与9658-byte XLSX导出 |
| 00201只读兼容 | ✅ detail；通用export按管理特质保护边界403，不放宽 |
| production | ⚪ 未访问 |

## MNG005-PRODUCTION-TEST-POLICY — 2026-10-09 GREEN

| 分支 | 状态 |
|---|---|
| local/staging双环境声明一致 | ✅ 注册005 TEST运行时 |
| production双环境声明一致 | ✅ 注册005 TEST运行时；生产部署仍须显式配置两项现有环境声明 |
| 空值、单边、不一致、`prod`或畸形值 | ✅ 继续关闭，HTTP不可覆盖进程环境 |
| 生产导入初始状态 | ✅ 两个005测评以现有`state=1`可见禁用，不新增第三个功能开关 |
| 管理员将单个已冻结005由禁用改为进行中 | ✅ 专用`/management-traits/admin/exam/state`只更新现有state |
| 管理员将单个已冻结005由进行中改为禁用 | ✅ 同一专用操作，不放宽legacy guard |
| 匿名/普通权限、非005、未冻结005、非法state、额外字段、重复键 | ✅ 写前或事务内拒绝 |
| DB失败/并发删除profile或exam | ✅ 行锁、条件UPDATE、事务回滚、受控错误 |
| 管理端列表启用/禁用操作、确认、成功刷新、失败提示 | ✅ 仅005管理员显示；focused API/SFC 152项通过 |
| 正式报告 | ⚪ 本策略只开放保留TEST标注的TEST/reissue，不把formal称已批准 |

## MT-005-QUESTION-BANK-CONTENT — 2026-10-09 STAGING GREEN

| 分支 | 状态 |
|---|---|
| 00501题库名称 | ✅ `管理特质测验基层员工新版` |
| 00502题库名称 | ✅ `管理特质测验干部新版` |
| 00501题干 | ✅ 140题继续逐字复制已核140/140的00201基层题本，不做覆盖 |
| 00502普通题干 | ✅ 除V67/V96外138题继续逐字复制00202 |
| 00502 V67 | ✅ 使用用户确认修订文案“当我接手具有挑战性的工作时，我通常能鼓励大家创新，并提出创新的解决方案。” |
| 00502 V96 | ✅ 使用客户管理版“当下属反对我的某个决定或者工作安排时，我会保持冷静和理性来应对。” |
| 题序、选项、分值、方向、公式 | ✅ 仍从002源复制，本次零修改 |
| 原00201/00202、冻结答卷及历史报告 | ✅ 不修改、不重算 |
| fresh fixture SQL与opt-in baseline writer | ✅ 两条005创建路径使用同一名称/题干合同 |
| staging/production现有数据 | ✅ staging已更新并完整复测；production未授权且未访问 |

## MT-005-SHARED-TEMPLATE-MANAGEMENT — 2026-10-09 STAGING GREEN

用户确认00501/00502继续共用一份报告模板；模板管理页增加一张共用模板卡片，不拆分产品文件。上传成功后，新生成和重发报告使用新模板；历史PDF及其模板SHA保持不变。

| 分支 | 状态 |
|---|---|
| 正常管理员/全局权限读取模板元数据 | ✅ 一张共用卡片，返回00501/00502、SHA、大小、修改时间与固定契约计数 |
| 未登录或非管理员读取、下载、上传 | ✅ 文件读取前返回401/403 |
| 共用模板存在且契约有效/文件缺失/契约损坏 | ✅ 元数据区分存在、有效及受控校验错误 |
| 下载有效DOCX | ✅ DOCX MIME、Content-Length、RFC5987文件名及no-store头完整 |
| 上传空文件、非DOCX、0字节、20MiB超限、ZIP损坏 | ✅ 写盘前拒绝 |
| 上传含路径穿越、重复ZIP部件、外部关系或不兼容控件/图表 | ✅ 写盘前拒绝，不改变当前模板 |
| 上传兼容模板 | ✅ 先备份旧文件，再同目录原子替换并返回新SHA；00501/00502共用 |
| 上传失败 | ✅ 页面保留已选文件，允许修正后重试 |
| 上传与报告生成并发 | ✅ 同一handler的模板读取与替换由RWMutex互斥；单次生成绑定同一字节及SHA |
| 新模板生成TEST/reissue报告 | ✅ 报告记录实际模板SHA，复用和下载按记录SHA校验 |
| 模板替换后的历史TEST/reissue报告 | ✅ 自动备份SHA进入受控allowlist；历史PDF不重渲染且可继续读取 |
| production或TEST运行时门禁关闭 | ✅ router测试证明不注册三条模板管理路由 |
| 页面00401与MBTI模板管理 | ✅ 前端全量602项及production build通过；未做已登录浏览器回归 |

本地验证：Go全量、build、vet均exit 0；前端35文件602项及`build:prod`通过。后续staging真实管理员浏览器完成元数据、下载、同文件上传、刷新及页面卡片验收，模板SHA和自动备份均核实；TEST报告按实际模板SHA生成。production未访问。

## MT-ROLLBACK-ADDITIVE-SCHEMA — 2026-10-09 STAGING GREEN

| 分支 | 状态 |
|---|---|
| rollback前active paper或三新增表非空 | ✅ fail-closed，不切换；实际preflight均0 |
| 三目标表缺失任一表 | ✅ canonical capture前要求exact count=3 |
| 表engine/collation漂移 | ✅ canonical SHA变化并在该阶段失败 |
| 列name/type/null/default/extra/order漂移 | ✅ 29条有序column记录，NULL default独立标记 |
| 索引name/unique/order/columns漂移 | ✅ 20条有序index-column记录 |
| FK name/cols/ref table/cols/update/delete漂移 | ✅ 7条有序FK-column记录 |
| PRE/OLD/POST任一不等迁移预期 | ✅ 每阶段立即比较expected SHA并失败 |
| 三阶段彼此不一致 | ✅ 恢复新runtime后再做三方相等门禁 |
| 历史rehearsal receipt metadata不足 | ✅ 明确`exactComparable=false`，不后验宣称匹配 |
| 成功完整复跑 | ✅ PID14293→15461→15554、restart2、三SHA相同、最终健康及不变量全绿 |

## LOCAL-SAFE-SERVER-CONTROLS — 2026-10-09 LOCAL GREEN

| 分支 | 状态 |
|---|---|
| host空值或`0.0.0.0` | ✅ 保持原有`:port`监听 |
| host为显式IP | ✅ 只用`net.ParseIP`，不触发DNS，使用规范数字地址和`net.JoinHostPort` |
| host精确为小写`localhost`且解析结果非空、全部loopback | ✅ 单次2秒受限解析；优先127.0.0.1，否则排序IPv6；绑定规范数字IP，不二次DNS |
| localhost解析错误、空结果、任一非loopback或混合结果 | ✅ 在数据库初始化前拒绝 |
| 大小写、空白等localhost变体 | ✅ 作为非允许hostname拒绝且不触发DNS |
| host为域名、含端口或其他畸形值 | ✅ 启动前拒绝 |
| 未设置`LOCAL_DISABLE_BACKGROUND_WORKERS`/配置false | ✅ 胜任力及管理特质Worker保持默认启动 |
| `APP_ENV=local`且loopback绑定、disable=true | ✅ 两类Worker均不启动，HTTP路由不受该开关控制 |
| 非local、非loopback或大小写不精确的local、disable=true | ✅ 启动前失败关闭，不静默忽略 |
| local+loopback+disable=true且两个TEST环境均production | ✅ Worker均关闭，管理特质TEST routes继续由独立环境门禁省略 |
| HTTP query/header/body尝试控制host或Worker | ✅ 无请求输入消费路径 |

## MT-TEST-RUNTIME-ENV-GATE — 2026-10-09 LOCAL GREEN

| 分支 | 状态 |
|---|---|
| 两个可信显式环境trim+lower后同为local | ✅ 注册TEST routes并启动expiry worker |
| 两个可信显式环境trim+lower后同为staging | ✅ 注册TEST routes并启动expiry worker |
| unset、单边设置、local/staging不一致、prod/production、畸形值 | ✅ 零runtime构造、零Schema预检、零worker、TEST/formal routes不注册 |
| HTTP query/header伪造local/staging | ✅ 实际404，不能影响进程级可信环境判断 |
| old002 legacy paper/create/detail/fill/submit/report及legacy guard | ✅ 禁用TEST runtime时仍保持注册/保护；identity辅助使用显式disabled assembly而不探测sidecar |

## MT-005-HARNESS-EXACT-RECEIPT — 2026-10-09 LOCAL GREEN

✅ valid manifest复用；✅ partial manifest拒绝；✅ cookie token解析；✅ clean report cycle；✅ oversized identity拒绝；✅ 旧固定marker拒绝；✅ receipt状态精确。harness以完整7字段deep equality消费producer，新增/缺失字段均失败；本轮仅修漏期望，不删除producer安全字段或放宽旧marker门禁。

## MT-005-RESET-JOURNAL-FINAL-HARDENING — 2026-10-09 LOCAL GREEN / ACTUAL待核

| 分支 | 状态 |
|---|---|
| intent/outcome/head exact filename regex含6位零填充sequence及12位hashprefix；JSON拒unknown/trailing | ✅ FB-212 |
| schema/mode/generation/sequence/status、nonnegative counts、deleted=pre-post、retained runs/products=2/2及previous links精确 | ✅ FB-212 |
| pending pre/post按排序`nameHash+contentSHA+bytes`完整multiset比较；同数量替换拒绝；不保存raw filename | ✅ FB-213 |
| 唯一最高next orphan outcome在全部link/state校验后补immutable head；多个orphan拒绝 | ✅ FB-214 |
| 现存legacy sequence1～3保持只读可验；新head文件名绑定head内容SHA | ✅ compatibility regression；actual sequence4待核 |
| legacy outcome-prefix与新head-SHA-prefix都要求head原始字节精确等于canonical indent/newline；空白/重排/篡改拒绝且不改写历史文件 | ✅ FB-215；无新sequence |

## MT-005-RESET-JOURNAL-V2 — 2026-10-09 CODE GREEN / ACTUAL待核

| 分支 | 初始状态 |
|---|---|
| intent必须在DB mutation/file delete前以O_EXCL+fsync持久化 | ✅ focused test |
| intent含sequence/mode/generation/previousOutcomeSHA/legacy bridge/random nonce、pre与expected-post counts | ✅ focused test |
| outcome不可变引用intentSHA，actual deleted counts由pre-post差值，master不覆盖 | ✅ focused test |
| immutable chain-head、全量link/gap/duplicate/rewrite检查，legacy reset-chain只读 | ✅ focused test |
| pending精确pre→recovered-no-mutation、精确post→recovered、混合状态拒绝 | ✅ focused test |
| outcome写失败保留pending，下一reset先自动reconcile | ✅ focused test；actual injection待核 |
| active v2 receipt及文件名无19位raw ID/UUID/JWT/password/token/credential | ✅ synthetic test；actual递归扫描待核 |
| report-only receipt不含raw examId，commitment/oracle只消费examIdHash | ✅ source contract；actual receipt待核 |

## MT-005-RESET-RECEIPT-CHAIN — 2026-10-09

✅ reset stdout按JSON实际解析；✅ 删除硬编码`cleanStart`；✅ sequence/mode/generation、实际删除/保留计数、manifest/evidence/helper SHA、timestamp完整；✅ append-only indexed receipt及master hash chain；✅ report-only v3 final receipt绑定pre/post reset SHA；✅ 不完整evidence与旧链字节漂移失败；✅ receipt禁raw ID/UUID/secret。⏳ actual链证明继续执行：报告后post-reset应删除2，再standalone reset应删除0，保留两条run/product且旧run receipts字节不变。

## MT-005-ACTIVE-BASELINE-GENERATION — 2026-10-09

✅ 新generation marker/path/evidence/private manifest四者一致；✅ runtime只有一个immutable active generation；✅ 旧固定marker无evidence继续fail closed且不被覆盖；✅ prepare必须280 answer-save/2 start/2 submit且report0；✅ report-only必须answer click/save/start/submit全0；✅ 两轮报告均generate/view/download和duplicate reuse；✅ reset只清report/audit/private report PDF并保留两条完整业务链；✅ 第二次reset幂等；✅ exact failed-generation cleanup禁止LIKE/FK关闭并不触及旧baseline/fixture；✅ source DB SELECT仍1142拒绝。

## MT-005-RETAINED-EVIDENCE-BINDING — 2026-10-09

✅ 独立evidence SHA与LocalAppData SID-only私有副本一致；✅ product/exam/run/title/name/gender/phone hashes全部与DB现状一致；✅ 无report evidence时`expectedReports=0`且任一report row失败；✅ evidence缺失、name drift、SHA drift、runtime binding缺失均失败；✅ 后续reset/regenerate report receipt不能替代baseline创建receipt；⚠️ 现存baseline缺独立创建receipt，保留数据但不可复用，不做DB删除/追认。

## MT-005-RETAINED-BASELINE-MANIFEST — 2026-10-09（identity binding结论由FB-205纠正）

✅ private manifest必须exact两产品并含candidate/paper/snapshot/run/receipt raw IDs、版本/source/hash和报告ID数组；✅ raw IDs只在工作区外SID-only DPAPI并由stdin加载；✅ workspace只存hash；✅ 20条parameterless global LEFT JOIN任一count>0失败；✅ exact retained candidate或identity/mapping/input hash漂移失败；✅ revision/current/audit/reissue任一非预期row失败；✅ 当前两产品计数各1/1/1/1/1/1/140/140/700/1/13/4/1且五类报告链显式0。⚠️ obsolete旧链删除前未保存child IDs，不能后验枚举；global orphan0仅关闭parent absent residual，不证明完整非孤儿未知链不存在。DB写/delete/remote/restart均0。

## MT-005-PARTIAL-CLEANUP-INSPECTOR — 2026-10-09

✅ exact obsolete parent=0 but any directly queried child count>0 fails；✅ all obsolete parent/child counts=0 plus both exact retained baseline IDs/counts unchanged returns `LOCAL_PARTIAL_BASELINE_ABSENT`；✅ either retained baseline missing or any expected count changed fails closed；✅ no LIKE、no FK disable、no broad marker search、no database writes/remotes/service restart。

## MT-005-IDENTITY-V2-FRESH / PDF-OBJECT-BUDGET — 2026-10-08

✅ formula绑定product/formula source/producer source/core receipt/title/exam/run/report/3 identity fields/PDF；✅ 12分量任一mutation改变commitment；✅ core receipt必须在commitment前create-only落盘，commitment必须在cleanup前；✅ 16MiB单object与128MiB总object budget，真实10,688,445-byte流合法，单体/总量超1拒绝；✅ fresh结果目录唯一且旧PDF/blocked evidence不覆盖；✅ aborted exact-owned chain按标题shape+产品+计数+report0回收，最终copy-only/source1142/crossFK0/四loopback恢复。❌ 两产品140UI、生成复用/view/native download、新PDF oracle和runIdHash完整commitment本轮未取得；三次登录检测上限后停止，禁止伪JWT/旧PDF后验补证。

## MT-005-STRUCTURE-IDENTITY-PDF-DEEP-SCAN — 2026-10-08

✅ 双真实Go source SHA＋四AST node位置/hash；✅ exact dimension builder/catalog与production module aggregation；✅ decoy/partial alternate mapping及模块漂移拒绝、benign strings接受；✅ PDF needs_pass/encrypted fail-closed；✅ 全xref object、raw/decoded stream、metadata/trailer/catalog/attachment多编码有界扫描，当前findings0。⚠️ identity v2公式与未来E2E run hash采集已实现，但既有immutable receipt无run hash且业务记录已清理；existing模式必须BLOCKED且records空，禁止由exam/report/title猜run或生成final commitment。完整关闭需未来获准的新E2E在cleanup前写runIdHash。

## MT-005-PDF-ORACLE-FOUR-BLOCKERS — 2026-10-08

✅ canonical评分来源必须绑定真实Go source/function/catalog SHA，13维顺序、140唯一题、100正40反、四模块成员3/4/3/3任一漂移失败；✅ E2E fixture source、browser receipt、report hash、PDF SHA/bytes与hash-only identity commitment必须闭合；✅ PDF姓名/性别/手机号按标签提取后3 field hashes和总commitment精确匹配，每份11位手机号恰1；✅ JSON对象递归检查敏感key/value并有界解码Base64，PDF正文/metadata/xref object string/附件名与内容均扫描；✅ 任一credential finding、附件总量超10MiB、source/PDF/receipt漂移失败。当前Node/Python GREEN、JSON/PDF findings0；不含UI/DB/report generation/services重跑。

## MT-REISSUE-PURPOSE-CONTRACT — 2026-10-08

✅ qualification/generate的purpose必须精确等于后端固定TEST用途文案；✅ eligible=true+正确kind+精确purpose允许新reissue；✅ 任意其他purpose仍拒绝；✅ 旧报告生成器不fallback。真实00501 RED后最小SFC修复，相关171pass；00501/00502真实UI生成/幂等复用/view/download均GREEN，cleanup后report/audit/private file均0。

## MT-DEBUG-RUNTIME-BINDING — 2026-10-08 CodeReviewer 后续阻断

✅ 同目录私有cipher temp/CreateNew/Flush/ACL/nonreparse；✅ commit前必须DPAPI解密、JSON及预期字段通过；✅ 已有目标Replace无backup、首次Move无覆盖；✅ 任意失败旧target SHA不变且只删本次temp；✅ 正常模式五绑定全缺才首次写；✅ partial/stale/same分别拒绝/拒绝/接受且拒绝不写；✅ 仅显式operator `RebindRuntime`可换五绑定，三个秘密不变；✅ 默认Go测试不触碰本机，显式FB-202只传非秘密身份并验证实际Windows inspector；✅ verify/backend共用不可变绑定。实际密文不变、PID20036/12324和health/NOAUTH保持；真实rebind、restart、远端、DB写、280 UI均0。

## MT-DEBUG-REDIS-OWNERSHIP — 2026-10-08 CodeReviewer blocking

✅ host仅127.0.0.1；✅ exact config port23317或launcher配置值；✅ persisted PID>0且等于当前唯一listener owner；✅ Memurai实际exe与DPAPI绑定路径一致；✅ config普通文件/private ACL/nonreparse/完整SHA；✅ bind/protected-mode/port/databases/requirepass唯一合同且secret常量时间一致；✅ AUTH PING和DBSIZE成功，运行态非负key数允许；✅ verify/inspect/fixture/cleanup/backend启动共用门禁；✅ 错host/port/PID/listener/exe/ACL/hash/secret/auth/DBSIZE均测试拒绝。实际PID12324、DBSIZE1、backend新PID20036及live health通过；远端DB写/部署/UI重跑0。

## MT-DEBUG-CREDENTIAL-ARGV — 2026-10-08 专属副本凭据

✅ 含密码CREATE USER通过stdin非-e argv（真实Bash合成mock RED→GREEN）；✅ 关闭xtrace/保留双host八项copy-only授权；✅ 精确ownership、host名单、root权限、备份SHA及本地SID ACL门禁后才轮换；✅ 旧密文先备份/新pending先加密，失败不回滚已暴露密码、只同步原pending；✅ 新连接0/旧连接BLOCKED1/原Check0/源SELECT1142拒绝；✅ 78表全dataSHA/源schema-data/旧资产缓存/PID保持。⚠️ 历史argv暴露后风险不消除；本轮不是应用启动/E2E/独立CodeReviewer。详见[修正与独立终验](management-traits-005-local-debug-20261008.md)。

## MT-TESTER-METADATA — 2026-10-08 修复前

🔥 URL/session不得跳过当前metadata；❌ 普通002及001/003/00401带mngTest仍旧登录；❌ 005draft/unknown无登录；❌ 跨exam/非法ID/读取失败/pending均拒绝；❌ 合法冻结002不因客户端exam结束禁止后端续答。仅本地SFC回归，真实环境登录与独立review仍未验。

[完成补充] ✅ 上述实际SFC分支RED→GREEN，本地相关433/全598；✅ 原普通002Save/旧分值/guard/冻结resume聚焦440/0/0；✅ 独立恢复78表/crossschemaFK0/本地专属账号连接及主库读取拒绝。⚠️ 完整后台启动、真实登录/报告E2E、005注册/来源/迁移与独立CodeReviewer未验，不沿DB Ping关业务运行门禁。[完整范围](management-traits-005-local-debug-20261008.md)。

## MT-005-ISOLATION — 2026-10-08 本地实施前矩阵

【完成补充】上述矩阵本地Go/SFC及实际编译UI已验证，以下初始状态保留为历史。普通002正常新建、显式新链拒绝；005强制独立draft/source版本；原002draft同产品编辑、跨005拒绝及原冻结只读兼容；005unknown禁旧；真实管理员/同exam/frozen/profile与两API命名空间均✅。全Go6400/0/11skip、追加35/0/0、前端586/0/0、browser28/0/0；实际MySQL、真实005题库注册/源审核、独立CodeReviewer仍⚠️，不当环境通过。[完整行号与证据](management-traits-005-isolation-local-20261008.md)。

最新005独立产品决策替代新建002强制新版政策；以下先RED再实现，不删除历史账本。

| 分支 | 初始状态 |
|---|---|
| 普通00201/00202缺flag或false正常旧新建，不探测draft结构 | ❌ |
| 新002显式true拒绝并提示使用005；既有002draft保留编辑、不转换005 | ❌ |
| 真实00501/00502强制单库140单选/25分钟/合法字段/独立draft；缺schema关闭 | ❌ |
| mixed、伪repoCode、005 flag=false或畸形bool、旧配置改005，首DML前拒绝 | ❌ |
| 005 staff/leader明确新版product/question namespace；002原canonical SHA保持 | ❌ |
| 005冻结依真实repo ID/source 140/700；禁止新冻结无profile002，已有冻结兼容 | ❌ |
| 002旧入口严格false；002兼容stricttrue同exam；005未知/false无draft不回旧 | ❌ |
| 配置005新版锁定、002旧表单无隐性opt-in；历史编辑不换产品 | ❌ |
| 管理结果同exam/真实code/管理员/profile；URL/session不赋冻结权限 | ❌ |
| 00401/MBTI原分流、旧成绩/PDF/来源不写；本地合成SFC/sqlmock验证 | ❌ |
| 真实005注册/题库复制、实际MySQL、客户内容审核、浏览器/独立review | ⚠️ 后续独立任务；本轮不连接DB或远端 |

## MT-ADMIN-ENTRY-ISOLATED — 2026-10-08 单SFC隔离候选

✅ 实际同ID/legacy两轴/精确002/stricttrue→fresh正常GetInfo数字正ID管理员或wildcard→同exam冻结profile→旧专属TEST页；✅ strictfalse无新标记真实历史/URL意图无效；✅ undefined/null/未知repo/冲突alias/非法ID/权限失败/profile null拒旧并重试；✅ non002/00401原分类、跨scope/销毁/迟到旧list。仅本地86专项/build0，template/style及旧API结果页逐SHA保持。❌ 完整编译AST/context依赖闭包/独立review/真实browser自动入口；🔥 线上旧URL仍未永久发布，不以393资源/被动URL库存PASS代替。[限定证据与候选](management-traits-entry-candidate-prepared-20261008.md)。

## MT-ADMIN-DIRECT — 2026-10-08T08:15Z 当前staging读恢复／自动入口仍🔥

✅ 当前正常管理员getInfo200/code200、同exam严格冻结bool与profile200；✅ 已存在专属结果页手动进入，实际list200/code0/1completed/140140及detail200/13维4模块/同run-paper-exam。✅ 原旧tester-list403保护保持，空表不可推断无人完成。✅ 既有local实际SFC/API fresh167/0/native0；🔥 线上旧URL无意图标记自动分流仍未发布，不能把手动导航标该分支✅。❌ 最小old-TEST入口前端source/compiler独立review/发布待主，❌ 整reissue实库门禁保留；无报告动作/真人写/DDL/生产。[安全收口](management-traits-reissue-staging-release-20261008.md#L3)。

## MT-REISSUE-1062 — 2026-10-08 修复前矩阵

🔥 typed报告输入唯一键1062须完整ROLLBACK后fresh来源/同归档/生成审计/PDF大小SHA复核，仅复用精确winner；❌ 其他主键/审计1062、1213、普通字符串错误不恢复；❌ 缺winner/跨run-paper-exam/原始字节及模板内容漂移关闭；❌ 撤销或输入变化/缺PDF/hash-size-path异常/审计缺失或失败/取消关闭且仅清自己的loser文件；❌ 真实两pool修复复验仍待，不沿旧mock成功宣称实库通过。渲染不重试、不改锁顺序/隔离级别/DDL/旧current/pdf_path。

[完成补充] ✅ 上述23分支本地实际服务sqlmock/文件GREEN，包含源当前读撤销、原raw字节漂移、原winner创建actor时间保留、render1/旧winner与历史file不删/loser0。❌ 真实MySQL具体键/读视图及修复竞争，❌ 独立安全review/发布；SSH首次＋唯一重试均timeout/远端0，不沿本地GREEN关实库门禁。[完整证据](management-traits-reissue-staging-release-20261008.md#L3)。

## MT-ADMIN-STAGING-BOOL — 本轮兼容 LOCAL GREEN，未发布

✅ 同exam合法ID类型/完整legacy两轴/实际002/strictfalse，仅标记缺失或可选一致legacy/false进旧list恰1；✅ URL意图不推翻可信serverfalse；✅ missing/null/stringfalse/numericfrozen、跨exam/非法ID、存在unknown/null/undefined生命周期/畸形newMode/新版矛盾均关闭；✅ 合法draft/true保持fresh权限后编辑、stricttrue无lifecycle保持字段/profile/fresh权限后专属结果。✅ 两实际脱敏DTO持久SFC、原39所有权/force/旧list延迟矩阵保留。RED5fail→full560pass/native0/build0/diagnostics0非coverage；❌ 独立review/发行闭包/真实MySQL与browser未验，[完整限定收据](management-traits-reissue-staging-release-20261008.md#L3-L15)。下方原RED历史保留。

## MT-ADMIN-STAGING-BOOL — 2026-10-08 🔥 新发布兼容门禁RED

[限定纠正] 旧“false历史保旧”只覆盖人工补齐lifecycle/newMode的夹具。实际已发布Detail对历史002为同exam＋legacy/legacy＋strict frozen=false，两个新字段缺失；当前入口默认created拒绝、旧list0。真实公开DTO＋实际SFC REDnative1，指定frozentrue本地adapter分流PASS。🔥 该旧bool合同必须补持久回归并最小兼容；unknown/跨scope/已返回矛盾字段保持拒绝，不添加draft后端解锁。当前未修、未发布；[完整证据](management-traits-reissue-staging-release-20261008.md#L33-L45)。

## MT-ADMIN-REVIEW-3 — 2026-10-08 条件/请求所有权限定纠正

✅ 同exam完整legacy/legacy或competency/competency_average、实际repoCode/repoList classifier，unknown/跨exam/矛盾拒绝；✅ 合法001/003无新生命周期保持旧、00401保持专属、002仅明确legacy false旧/draft true新编辑/frozen true新版、strict权限不变。✅ 行metadata旧then/catch/finally丢弃、force生成第二req/精确生成ID、pending查看等待、刷新失败“已生成”可retry、替换同IDrow/同run跨exam/destroy失效。✅ 旧list开放封闭query快照、序号/当前exam/watch/销毁、迟到成功与失败不盖新spinner。真实SFC新增39、最终525pass/build0非覆盖率；[全部分支与失败](management-traits-admin-ui-local-20261008.md)。

❌ 当前bundle浏览器最终有界120s/null-SIGTERM无summary、390/完整两case未验；❌ 独立CodeReviewer主协调者后续，非release ready。旧两case PASS保留仅历史，不新增后端姓名手机号/逐题合同、不部署。

## MT-ADMIN-UI — 2026-10-08 本地收口（未部署）

✅ 实际SFC strict frozen直达/false历史/draft/unknown/当前exam/字段缺失/URL spoof/真实getInfo权限/迟到响应；✅ 200列表合同、本地状态查询重置分页、精确HALF_UP两位/0与NULL、13维4模块事实与同run/paper绑定；✅ lazy资格/metadata、unknown不是未生成、多revision选择、新嵌套DTO、行三报告动作、防重复/失败/404409503不旧回退与PDF/MIME/%PDF/RFC5987；✅ 当前生产bundle合成mock1440/390实测直达/13+4/生成查看native合成下载/503结果保留/页面scroll=client，两case/0pageerror/forbidden0/closedtrue/native0。

真实browser另发现只读Detail在旧入口/结果页1000ms重复被公共POST拦截；两失败保留，用户额外1轮批准后只切已有隔离配置wrapper、公共guard不改。最终全前端33文件486pass/0fail、生产buildexit0，Go222前后SHA同。[完整报告](management-traits-admin-ui-local-20261008.md)。❌ 远端新表/API发行、真实报告/客户最终视觉与独立review；❌ 姓名手机/逐题信息及服务端搜索分页后端未提供，不伪造。初始矩阵保留为历史。

## MT-ADMIN-UI — 2026-10-08 实施前矩阵（local only）

| 分支 | 初始状态 |
|---|---|
| 同exam服务器strict frozen true直达；无query/session也进入专属结果 | 🔥 待RED |
| strict false历史保旧；draft回配置；unknown/畸形/失败不回旧 | ❌ |
| 正常getInfo数字正ID管理员/全局权限；role-only/null/失败关闭 | ❌ |
| URL/meta不能解锁；跨exam迟到和销毁响应失效 | ❌ |
| 真实examId-only列表/200上限；本地状态查询重置分页；0分与NULL区分 | ❌ |
| 姓名/手机号和逐题信息未提供，明确空缺、不造API或参数 | ❌ |
| 13维4模块详情run/paper/exam一致；完整140/completed生成守卫 | ❌ |
| 单行lazy资格/独立metadata；未知不是未生成；多revision明确选择 | ❌ |
| generate/view/download仅report-reissues，防重复/失败/刷新该行 | ❌ |
| 404/409/503报告能力关闭不清结果，不旧生成器fallback | ❌ |
| PDF MIME+%PDF验证、RFC5987文件名、关闭释放URL | ❌ |
| Element UI查询/表格/返回/错误重试；390页面不溢出 | ❌ |

不新批量API/后端/DDL/draft/审批/Worker，不访问远端或真实身份/PDF。历史矩阵保留。

## MT-REISSUE-API — 2026-10-08 本地限定收口

✅ 服务/sqlmock两code完整冻结输入、外部render失败、二次锁复读、metadata/audit rollback、模拟并发winner复用及自有file清理；✅ 新结构全列/NULL/collation/索引前缀/FK漂移/缺失关闭、空元数据[]；✅ actual Gin strict请求/正常管理员/普通exam:list/非正wildcard、四GET权限及缺结构；✅ 实际httptest服务器view/download200/PDF原字节/RFC5987/安全头、archive状态/hash/path/size/snapshot篡改拒绝；✅ 真实只读140/13/4/owner及三原字节SHA、真实答案独立Rat和合成身份实际LO9页/36客户段。

⚠️ 新源拒绝用例部分提前返回，仅断言拒绝、不声称其未消费读取后缀已验；原完整loader矩阵全Go通过。❌ 真MySQL首轮/重跑与跨连接竞争、全部commit故障、所有symlink组合、原实名新HTTP、独立review、UI/远端验收。专项49/0/1skip、默认全Go6359/0/11skip不是覆盖率。旧current/pdf_path/DB结果写0，DDL/部署0；[完整证据](management-traits-reissue-api-local-20261008.md)。下表实施前状态保留。

## MT-REISSUE-API — 2026-10-08 实施前矩阵（仅本地后端）

| 分支 | 初始状态 |
|---|---|
| 正常JWT管理员／普通exam:list／未登录，权限先于数据库 | ❌ |
| 请求仅runId；空串/null/重复/未知/客户端路径或批准标记拒绝 | ❌ |
| 完整可信冻结00201/00202，唯一owner、140raw、13维4模块receipt | ❌ |
| 历史无证据／不存在／140题不足／评分损坏／撤销来源拒绝 | ❌ |
| 新独立表缺失／列索引FK漂移，仅报告能力关闭、不startup迁移 | ❌ |
| 事务外渲染，提交前同paper锁重新读取完整冻结字节 | ❌ |
| 同输入幂等／并发重复复用，新UUID私有文件，旧current/pdf_path零写 | ❌ |
| metadata＋generate审计原子，失败只清理本次孤儿文件 | ❌ |
| 元数据非nil数组，无身份/原始来源/path披露；最新确定性排序 | ❌ |
| ID查看下载，路径/符号链接/大小/%PDF/SHA/非法状态拒绝 | ❌ |
| 撤销禁生成；授权读取已归档PDF不重渲染、不读取当前人员 | ❌ |
| 真实来源只读审计、实际本地LO、未部署/DDL0边界 | ❌ |

范围依据用户已给完整分支要求，不新增UI/审批/签密钥。历史账本保留。

## MT-CANDIDATE-MINIMAL — 2026-10-07 精确线上旧协议隔离回放

✅ 精确8baa旧源码＋新回归有效RED后，Detail frozen true/冻结字段、legacy/absent/non002严格false、repo/schema/profile失败关闭及private键不泄露：Gin8事件GREEN；✅ strict未知键/null/别名6、body1、配置身份create9事件（含合法gender/空gender拒绝/未配置gender拒绝），总24/0/0。✅ 草稿协议前真实编译SFC/API106/0/0，包括当前ID/boolean、unknown关闭、URL spoof、旧false保留、字段子集与迟到响应。均local overlay/sqlmock，不新表或真实用户保存。

❌ 独立source-review、前端编译资源等价归因、受限备份/维护发布及真实synthetic注册200/SQL/cleanup；未关闭线上反馈。两隔离UI build0但386资产baseline index不等线上393资产index，不能写发布PASS。原工作树不改、已有测试和草稿账本不删除。[精确范围与收据](management-traits-candidate-identity-local-20261007.md#L3)。

## MT-MINIMAL-REISSUE — 2026-10-07 独立本地报告切片

用户最新范围只独立可靠输入→不可变快照→原002客户Word→LO，不推进旧draft/formal/Worker/fullUI，不删除其历史账本。新增服务边界，不新增HTTP handler/API/SQL。

| 分支 | 本轮证据 |
|---|---|
| 00201/staff、00202/leader，完整140、0/50/100、40反向一次 | ✅ 新适配器单位测试，两code六种合成分值 |
| 原始字节签名/应用信任根、unsigned/未知根/篡改/预算、strict重复/空串/null/类型/case | ✅ 新拒绝矩阵，不允许JSON自声明verified替代审计根 |
| V/维度/方向/原source题项映射/所选raw/五桶/一checked一right/重复缺题 | ✅ Canonical既有入口+新增桶跨核；签了无效输入仍拒绝 |
| historical owner与input绑定/身份白名单/原时间两格式/缺时倒序 | ✅ 不读当前人员、不猜原始日期；tester及candidate合成输入 |
| 13维sum-count-exact/norm、4模块exact、705/13、条件文案36段/immutable返回 | ✅ 新评分侧车与独立Fraction/XLSX校验，不新算法 |
| 原new_creation产品入口保持严格拒历史输入 | ✅ 专项显式拒绝，既有生成/下载/身份回归保持 |
| synthetic实际Word→LO/%PDF/SHA/六图/原件样式保护、旧pdf_path零写 | ✅ 实际最终654393bytes/A4九页，5环13柱1常模线，68/75非值OPC字节保持 |
| 两真实历史来源审计、可信loader与管理员单卷按钮、正式/目标环境验收 | ❌ 下一slice；本轮无SQL/HTTP/部署/真实用户写入，不把签名或合成PDF当历史资格 |

[完整结果/独立作品与范围](management-traits-minimal-reissue-local-20261007.md)。最终303为含子项PASS事件、非覆盖率。前端原exit1未重跑/未关闭。

## MT-NEW-DRAFT — 2026-10-07 本地收口

✅ 新建server002/缺flag默认/伪clientcode/false400/mixed零DML/缺schema零DML/marker插入失败rollback；✅ actual编辑切另一002同TX保create_time/切非002拒；✅ cancel保持draft/重读可编辑；✅ candidate登记与完整schema下tester缺profile拒/旧开卷403；✅ 精确admin准备与GET列表204、低权/全库403；✅ public freeze真实source＋profile/bundle/marker原子、source不符零INSERT/marker失败rollback；✅ Detail strict三态、真实SFC unknown/draft关/旧106保持。

❌ 真实MySQL新表首轮/重跑、完整browser、独立review/staging验收、所有metadata漂移/COMMIT/真实Excel准备/跨实例并发。Go6272/0/9skip、front435、build/vet0为local测试，不coverage。[完整消费方、矩阵与限定结果](management-traits-new-draft-local-20261007.md)。下方实施前账本保留，不把全量GREEN当新表环境通过。

## MT-NEW-DRAFT — 2026-10-07 实施前矩阵

独立持久化 sidecar，不复用旧字段、不回填历史，保存和冻结保持两次确认。所有验证仅local。

| 分支 | 初始状态 |
|---|---|
| 真实repo002新建、缺boolean默认新版、false/畸形boolean拒绝 | 🔥 RED待执行 |
| 单一140单选、25分钟、合法身份子集；mixed/伪code/切非002零DML | ❌ |
| exam/link/draft同事务；任一写失败ROLLBACK；缺schema新002关闭 | ❌ |
| draft配置编辑保create_time/同步canonical repo；取消freeze保持draft | 🔥 |
| draft匿名candidate登记/tester登录/旧create-paper拒绝 | 🔥 |
| 精确scope管理员tester准备和列表允许；跨scope/全库/frozen拒绝 | ❌ |
| freeze同事务绑定真实repo/profile/bundle并将draft标记为frozen、不删标记 | ❌ |
| Detail明确isManagementTraits/lifecycle/strict frozen/current ID；未知关闭 | ❌ |
| 历史legacy/profile0和已冻结旧新版不回填；非002原行为 | ❌ |
| 真实SFC草稿/未知/迟到响应/必选新版；旧106身份回归 | ❌ |
| 新表DDL仅artifact、strict metadata/FK/索引；远端执行0 | ❌ |

## MT-CANDIDATE-FROZEN — 2026-10-07 增量分支

- ✅ URL显式新版但服务器false/missing/null/string/number：禁止登记/旧保存，字段仍null/表单不显示；只有当前测评ID＋strict true才解锁。
- ✅ 普通002 missing/null/string/number/object/array不回退legacy；URL伪造001但服务器002同样关闭；服务器false跨测评/缺ID/畸形ID不解锁旧链。
- ✅ 普通入口第一次true、第二次false/缺失保持新版且关闭；安全数字/字符串匹配当前ID正例、新旧字段子集、加载失败与代际隔离原回归保持。
- ✅ 本地身份106/全前端430/build0/诊断0。❌ 独立CodeReviewer、实库及完整浏览器、远端保存；未发布，不代替正式功能/生产门禁。[本轮限定结论](management-traits-candidate-identity-local-20261007.md)。

## MT-CANDIDATE-FIELDS — 2026-10-07 本地分支收口

- ✅ 普通queryless server-frozen两字段/配置gender三字段：真实SFC DOM/rules与API严格payload；不从002 code猜新模式。当前真实三字段不改。
- ✅ Detail无profile/全缺结构/非002保持旧投影；完整冻结profile只出字段串+布尔，不出private信息；分类查询/metadata/profile损坏关闭，原严格decoder/未配置gender拒绝和零INSERT保持。
- ✅ configpending/failure/空或重复未知/sex-mobile别名、旧异步配置与身份响应隔离、旧002正常saveData；selected typed空串按原服务合同不变。
- ✅ 前端404及最终Go6247/0/9skip、本地build/vet0、范围SHA。❌ 真实远端保存GREEN/独立CodeReviewer/性能负载；未coverage或完整浏览器UI。formal/production、旧新建入口治理及历史保留后续不动。[完整证据](management-traits-candidate-identity-local-20261007.md)。下条实施前矩阵保留。

## MT-CANDIDATE-FIELDS — 2026-10-07 实施前矩阵

| 分支 | 修改前状态/验收 |
|---|---|
| queryless新版开放入口，服务器冻结name+telephone | 🔥 旧全表单路径；必须隐藏gender、rules只含两键、提交仅两字段+examId |
| 冻结name+gender+telephone | ❌ 必须显示/校验gender，不假改真实配置 |
| 当前profile与exam配置不符 | ❌ 公开响应仅可信冻结字段，不改exam记录 |
| 002无profile或全部结构不存在 | ❌ 旧身份路径保持，不从repoCode猜新版 |
| metadata/完整profile验证失败、字段缺失/重复/未知 | ❌ 失败关闭，不legacy回退/写入 |
| 配置等待/重复提交/路由exam切换迟到响应 | ❌ 未ready零请求，旧响应不得污染新exam |
| strict Candidate.Save未知键/null/别名/未配置非空gender | ❌ 实际Gin/sqlmock零INSERT拒绝，不放宽守卫 |
| 公共详情消费者/00401/MBTI/历史/formal/TEST环境 | ❌ 全回归/范围SHA；禁止部署与数据变更 |

## 2026-10-06 MT-FORMAL slice1：已定政策、实施前分支

[完成补充] 下表实施前状态保留；新正式专项53pass/14顶层0fail：FS01/03/04/06/07/09本地公开服务/Gin通过；FS02覆盖完整/全缺/部分/类型NULL索引FK及配置分类（真实MySQL/全孤儿故障未验）；FS05真实原候选16LINK/TEST模板/错误ZIP及key逃逸拒绝、缺模板草稿通过，符号链接实测/宏外链完整负向组合未验；FS08仅audit插入故障及回滚/旧epoch拒绝，BEGIN/COMMIT/所有写点故障矩阵待补；FS10原TEST全量回归与限定scope收口，未PDF/前端/production实际验收。全Go6229/0/9skip、build/vet0，不coverage。PDF及两端UI仍❌，不把版本测试关闭MT-F07～18。[完整限定证据](management-traits-formal-registry-local-20261006.md)。

Q1=A（同一授权账号分别双签），Q2=B（精确获批的完整新版TEST run可追加独立正式报告），Q3=B（撤销禁新生成，授权管理员仍可读取历史原PDF并标revoked）。这是业务政策，不是具体资产批准。本次仅本地正式登记/审批/启用/撤销；PDF与两端UI未接入。

| 分支 | 实施前状态 | 验证入口 |
|---|---|---|
| FS01 未配置/非local/production/无库 | ❌ | 正式独立关闭，不探测DB或改变TEST |
| FS02 全缺表/部分表/列类型NULL索引FK或孤儿漂移 | ❌ | 独立schema完整校验；旧守卫只排除严格已知配置表 |
| FS03 未认证/非授权账号/伪审批人时间/重复键null空串未知字段 | ❌ | 正常JWT上下文，严格扁平请求 |
| FS04 登记受控资产及可信冻结题本/错误code或源hash | ❌ | 登记复制评分/映射，不复制人员、exam冻结或run身份 |
| FS05 无正式模板/资产变动/TEST模板/路径逃逸 | ❌ | 草稿可见阻断；缺合格模板不能双签或启用 |
| FS06 单签/同人两职责/重复签/不同摘要或旧epoch | ❌ | 两条独立不可覆盖记录，重复返回原证据，审批不自动启用 |
| FS07 显式启用/未双签/重复启用/撤销/再次启用 | ❌ | 同version行锁+epoch；撤销终态，audit原子 |
| FS08 BEGIN/INSERT/UPDATE/audit/COMMIT失败 | ❌ | 同事务rollback，不返回伪成功，不改create_time |
| FS09 列表空数组/状态和阻断原因/审批记录读取 | ❌ | 返回结构化readiness，非nil集合，不返回文件路径 |
| FS10 原TEST/历史PDF/前端/production | ❌ | 范围指纹及旧回归；不生成PDF、不部署/DDL |

完整run资格、13/4/receipt、历史formalPDF读取及生成竞争仍属于后续切片，不用版本测试关闭MT-F07～18。

## 2026-10-06 MT-FORMAL：正式报告/两端UI待实现分支（todo1草案）

用户已选正式功能，不继续harness。本轮只文档；[实施契约与Q1～Q3](management-traits-formal-report-implementation.md)未定稿，不把已确认开发范围当内容批准。以下24组均**未实现/未测试**，不得沿TEST历史PASS关闭。正式/TEST独立current、不ALTER001/旧11表，历史old results/PDF、00401/MBTI不动。

| ID | 待实现条件分支 | 初始状态/验收重点 |
|---|---|---|
| MT-F01 | 正式版本全部结构缺失/部分安装/签名或FK漂移 | ❌ 正式关闭；旧11表合同保持，未知extra仍fail-closed |
| MT-F02 | 未登录/非授权登记/审批/启用者；同人双职责 | ⚠️ Q1；正常JWT actor，不接受客户端姓名当批准 |
| MT-F03 | 草稿/单批准/双批准未启用/approved且active | ❌ 前三者正式零产物；启用必须独立确认 |
| MT-F04 | 内容/template/binding/scoring SHA、版本、受众、环境任一不符 | ❌ 精确失败，不跨版本或两code回退 |
| MT-F05 | 批准后替换资产/并发批准或启用/旧epoch请求 | ❌ 新revision重批；代际冲突拒绝，原审批证据保留 |
| MT-F06 | TEST内容/DTO/模板/路由冒正式、URL/storage伪用途 | ❌ TEST警示/SHA/purpose语义不变，服务端资格决定 |
| MT-F07 | 新正式测评/既有新版TEST run/历史old run | ⚠️ Q2；历史始终不迁移不重算，TEST转正式不擅选 |
| MT-F08 | new_creation/submitted_snapshot/completed与真实身份不符 | ❌ 保留来源门禁，不拿completed字符串当许可 |
| MT-F09 | 140完整/139未到期/0或3到期incomplete/子项或receipt漂移 | ❌ 正式仅140及13/4/receipt自洽；不完整NULL无报告 |
| MT-F10 | 原评分bundle退役/审核撤销；正式内容版本退役/撤销 | ❌ 来源可信性与正式批准分开；不换冻结题本 |
| MT-F11 | 正式版本撤销后已有PDF读取 | ⚠️ Q3；文件/审计保留，禁新生成，下载策略待选择 |
| MT-F12 | 同run换content/template | ❌ 仅新增正式revision；run/答案/input/时限零变动 |
| MT-F13 | 同paper两种报告current/同种多revision/跨paper关联 | ❌ 独立正式current不写TEST current，复合身份防跨卷 |
| MT-F14 | LO/取消/文件写/末次复核/审计或COMMIT失败 | ❌ 原报告/current保持，本次未登记文件清理，不回退旧链 |
| MT-F15 | 生成中启用切版/撤销/并发同卷完成 | ❌ 锁内复核epoch，晚到旧请求不覆盖有效current |
| MT-F16 | ZIP/XML/宏/外链/关系逃逸/重复槽/缺规则/超预算 | ❌ 候选拒绝；Word重存不自动安装；无外网求值 |
| MT-F17 | 正式六图/五数字槽/长文本/精确零分灰环/未配置字段 | ❌ 实际PDF验证，不只大小/页数；不强加年龄等位置 |
| MT-F18 | 下载缺/跨作用域reportId、路径/符号链接、size/SHA/PDF错误 | ❌ 发送头前拒绝；不返回路径或保存JSON伪PDF |
| MT-F19 | 管理端新建默认新版/历史编辑/未批准正式选择/保存与冻结 | ❌ 未批准禁正式；保留字段子集/25分钟/两确认；历史不补选 |
| MT-F20 | 行报告列表/用途切换/生成取消/刷新失败/旧慢响应 | ❌ 不手填UUID；只操作明确run/report/版本，状态刷新可信 |
| MT-F21 | 开放/封闭登记准备答题完成用途显示/过期续答 | ❌ TEST警示保持；不延原deadline、无匿名恢复、无参与者报告开放 |
| MT-F22 | 请求空串/null/别名/重复键；时间多格式；create_time | ❌ 独立扁平request；服务端审批时间，多格式ParseInLocation；不Save覆盖 |
| MT-F23 | 空版本/报告列表，390/768/1440、键盘/焦点/失败重试 | ❌ 非nil数组、loading、可操作拒绝原因；≥44px触控/一主CTA |
| MT-F24 | local批准误使production启用/非002/00401/MBTI/旧PDF | ❌ production关；越范围零修改、旧资产字节保持 |

下一由主协调者一次收Q1～Q3，定稿后再分切片实现；本轮不更新回归bug、反馈、覆盖率账本或工作流完成态，不运行任何旧harness。

## 2026-10-06T14:29Z freeze-only分支局部关闭；完整及正式门禁保持

| 分支 | 实际状态 |
|---|---|
| 正常保存响应有延迟、两独立确认框 | ✅ 真实本地SFC250ms有效RED1→同exam保存屏障/精确冻结标题GREEN0，不改产品 |
| 单00202candidate default/clear-reselect/保存取消/正常freeze/独立SQL | ✅ cc784HTTP200、frozen25/paper0/mapping及manifestSHA有效；不是四组合 |
| 精确PK/FK共享引用清理及旧PDF/源/配置/Schema/cache/PID | ✅ 三短批finally/final0；14:29旧465/PID17552/backend8baa/front353c保持，独立context关闭 |
| 四完整UI/自然0-3-140/20min/自然5min401/expired409/四native-Python | ❌ 本轮未执行，旧完整PARTIAL不回填 |
| 正式报告/生产环境 | 🔥 P0方案未确认/未实施；现仅TEST且production拒绝，不删警示或伪装staging |

相关SFC/API202pass/641源-Go215同；三次上限后用户另批一次最小测试修正，[完整证据](management-traits-default-staging-result-20261006.md)、[生产准备](management-traits-production-readiness-20261006.md)。不访问生产/发布/restart/迁移。

## 2026-10-06T14:06Z 三轮最终分支（整体BLOCKED）

| 分支 | 实际结论 |
|---|---|
| 独立正常登录即时提示/用户确认/自有context getInfo200 | ✅ 四新批真实通过，无900000静默URL等待/凭据转移 |
| default/清空同库重选/四人员draft精确scope/freeze | ✅ 单b274及fullf12通过；⚠️ 最后零答9b freeze Timeout，不能称所有路径已修 |
| 四140参与者/三manual/三报告五stage/native | ✅ fullf12 563保存、三manual/native；❌ 第四自然140报告与全四独立oracle未完成 |
| 异步safeStage和已选freeze避免toggle | ✅ actual compiled-driver RED→GREEN，仅local；full旧Error底层cause仍UNVERIFIED |
| 真实20min/1500sec自然0-3-140/续答自然5min/expired409 | ❌ 到期前失败finally；resume issue/140同卷原deadline已证，不冒expiryGREEN |
| finally/历史/source/旧PDF/当前服务 | ✅ 四批cleanup/final0、14:06fresh7roots0/11-private0/旧465/393/schema/cache/PID同，历史3-2-1/newside0 |

预算3/3停止、active0，不部署/restart/历史迁移，formal-prod/mainrace边界保持。[完整限定证据](management-traits-default-staging-result-20261006.md)。

## 2026-10-06 新有界测试补证（完整批尚在执行）

- ✅ local真实编译driver的normal-context200选择/401拒绝、固定安全stage+409、精确目标paper/run行三合同RED→GREEN；原生baseline/inspect启动合同也0，相关SFC202pass。
- ✅ 单新00202candidate清空同库重选/default/freeze/full140UI/manual满分/13维4模块/报告五阶段/native1独立DB-private-dataSHA/finally0。旧第二report cause仍⚠️UNVERIFIED，不称产品bug已修。
- ❌ 四组合新批f12bb9277e40/原1500sec自然0-3-140/四native/Python oracle尚未完成，正式/production/mainrace边界不解除。[真实单批与进度](management-traits-default-staging-result-20261006.md)。

## 2026-10-06 新默认staging实际分支：已发布，完整验收阻断

| 分支 | 本轮实证 | 限定 |
|---|---|---|
| approved frontend-only原子发布/backup/393公网SHA | ✅ build/backup/release/public/final均native0，新index353c9fb3…、backend8baa/PID17552同/restart0 | 仅新建002默认TEST，不历史迁移/正式批准 |
| 新建002两code×candidate/tester不点击TEST默认勾选/取消独立freeze→draft人员→4freeze | ✅ 首批正常UI，人员payload字符串0/scoped200 | 清空同库重选本轮未真验，本地85证据仅历史 |
| 四完整参与者/管理报告/native | ⚠️ 三140真实UI＋3partial/两manualcompleted/一native661148 | 第二报告Error未证cause，00202tester未执行，不4/4/不独立oracle |
| 原1500秒自然0/3/140与20min提示 | ❌ 本轮未完成 | 首批到期前finally，不靠旧UF050改为本批通过 |
| 正常新窗口认证/续答自然5min/原凭据expired409 | ❌ 最终首页15min等待超时/业务0，其余未验 | 用户确认不是getInfo200，无token导出/重置 |
| 精确finally/历史保留/本地scope | ✅ 首批cleanup0、独立final0、12:46freshend0/11-private0/旧465/source/schema/cache/历史3-2-1/641-Go215同 | 最终无新资源不需要清理，不称最终cleanup执行PASS；mainrace/formal-prod仍未验 |

[完整本轮结果](management-traits-default-staging-result-20261006.md)。原分支历史不删除。

## 2026-10-06 MT-NEW-DEFAULT 本地分支收口（非staging）

下表实施前矩阵保留；新建两个002/精确code及题型资格/混合库拒/清空和切换/手动取消/route与saved编辑ID/异步历史null/frozen只读/权限及00401隔离/独立冻结确认均 ✅ 实际SFC81项通过。原非法身份字段及loading/error仍由既有回归覆盖，全前端32/385通过。

真实本地构建六UI case ✅：四种新建默认+历史/冻结编辑，全部API localmock；staging新default和新四完整参与者/PDF/native仍 ❌，不沿原显式勾选suite回填。fullUI取证启动Node reserved inspect ✅ 原生RED→固定哨兵GREEN；已清理四PK只读inspect首次0/FK1/0行，不是整批E2E。

[完整限定证据和影响清单](management-traits-new-default-local-20261006.md)。无Go/SQL/API/00401/MBTI/历史PDF改动。

## 2026-10-06 MT-NEW-DEFAULT：新建默认新版、历史保留（实施前矩阵）

仅00201/00202；新默认仍为 TEST，不解除正式内容批准。仅新建管理员、legacy+legacy、joinType1、单一物理题库140单选/其余题型0默认勾选；个人字段不替换，保存后冻结仍需单独确认。历史编辑不自动切换；00401/MBTI/历史评分PDF不改。本轮只本地，不自动发布。

| 条件分支 | 初始状态 | 验证方式 |
|---|---|---|
| 新建00201/00202合法配置 | ❌ | 实际SFC repoChange、25分钟/禁参与者报告/字段子集/零自动保存冻结 |
| 非002精确code、题数139/141、混合库、多选/判断/简答非0、无ID | ❌ | 实际SFC不默认，不扩大既有资格 |
| 清空/切非002/切另一002、同库手动取消 | ❌ | 清除新建选择；同库保留取消，另002重新默认 |
| route编辑ID/postForm已保存ID/异步历史null profile | ❌ | 不自动补选、不覆盖历史字段/时间、不自动冻结 |
| 冻结profile/loading/error及权限不足 | ❌ | 原只读门禁保持；frozen true不因新默认关闭 |
| 00401配置、未支持个人字段 | ❌ | 版本/维度不变；原字段校验仍拒绝、不自动过滤 |
| 新默认保存→取消/确认冻结 | ❌ | 原Save协议；确认前零freeze；cancel保留draft |
| staging新default/新四case完整UI/native | ❌ 未验 | 本地完成后另确认frontend-only staging发布；旧显式勾选suite不证明新默认 |

## 2026-10-06 UF053 本地竞争观测分支（local GREEN，主 UNPROVEN）

本切片是安全观测，不是业务 bug 修复或主竞态 PASS；仅本地代码/测试获准，发布/restart/远端造数/SET 均未获准。

| 分支 | 状态 | 必须验证 |
|---|---|---|
| 缺失/非法环境或目标、不匹配/非 canonical v4 UUID | ✅ | 默认关闭、零观测；Gin disabled 返回同409 |
| HTTP participant / scanner Worker / 其他 trusted internal | ✅ | 真实 Gin、ScanExpiry、Submit；不从 timeout/source/header 推断，participant 固定来源 |
| 同资源并行 attempt / 重用 parent context | ✅ | 真实两入口并行/32独立 attempts、resource/session 稳定、取消与 values 保留；非 MySQL race |
| BEGIN / paper UPDATE / bundle SHARE / body return / transaction return | ✅ | 原 SQL 顺序；只成功锁读取记 acquired；sink 断言在全部原 SQL 完成后才 emit |
| 新建/复用/失败回滚/BEGIN 或 COMMIT 失败/未完成终态 | ✅ | 连续 body+Transaction 正常返回才 committed；有效 panic 误报 RED→GREEN；不声称 failed 的 DB 终态 |
| stage 数量/词汇/配置随机源失败 | ✅ | 固定容量/未知不输出/配置 entropy failure关闭；attempt RNG 故障未注入，未夸全随机源矩阵 |
| 主 HTTP 与 Worker 自然重叠/真实锁等待 | ❌ UNPROVEN | 本地 sqlmock 与观测能力不替主服务实证 |

[精确影响/配置/时钟及未验边界](management-traits-four-real-verification-20261003.md#L3)。service25/Gin7原生测试exit0；默认全量与构建收口另见报告，不新增覆盖率百分比。

## 2026-10-06T04:52Z UF-050/四功能UI/自然三例最终分支（整体PARTIAL）

| 分支 | 当前实证 | 限定 |
|---|---|---|
| 正常手工新增status→scope刷新→freeze→正确password | ✅ 四payload字符串0/SQLnonNULL0/login200code200五键；frontstrict8delta/公网393已发布 | 仅UF050手工入口，ImportData未扩/guard及旧NULL状态不改 |
| 四身份功能UI140/恢复/完成/admin13-4/TEST报告 | ✅ 四功能链/563正常键盘按钮保存/四报告generate-view200 | 非fetchloop/非鼠标click；423细粒度保留，一140丢失不补造；native独立 |
| 原1500sec自然0/3/140及20min提示/到期状态 | ✅ UI自动三200完成页、三提示可见；SQL140completed50与0/3incomplete全部正式NULL/noPDF；六唯一1/13/4/1 | deadline/clock/duration未改；主HTTP-Worker实际竞态仍UNPROVEN |
| 到期写/不完整报告拒绝 | ✅ 三原genuinepaper凭据save409、实际四incompleteadmingenerate409零新报告 | duplicateobserver6/unique3导致额外两guard请求保留，不称两次或401冒409 |
| Native/主竞态/ImportData/mainrestart/formal-prod | ❌ 四native各双eventtimeout/saveAs0，竞态未证，其余未验/未扩 | SCP私有合成副本不是native，整体PARTIAL |
| exactPK/FK/privateUUID finally及旧系统 | ✅ cleanup0/独立final0/全ownPK0/11逐0/private0/旧465-source-runtime-config-schema-cache/PID2002同，617scope同，原admin200/context1/pending空 | cleanup SESSION SafeUpdate1/FK启用、不global/权限/DDL/restart，受限backup保留 |

[最终完整证据](management-traits-four-real-verification-20261003.md#L3)、[机器状态](../scripts/test/results/uf050-ui-f61006041501/summary.json)。

## 2026-10-06T04:33Z UF-050 staging实际分支进度（非最终）

| 分支 | 当前事实 | 边界 |
|---|---|---|
| frontend strict全部八资产/backup/原子切换/公网393 | ✅ 原生exit0；仅status逻辑及webpack引用，受限backup/旧dist回滚保留；server/PID2002/旧465/source/schema/cache保持 | 无backend/config/DDL/nginxreload/任何restart/production；回滚未实际演练 |
| exact draft新增→scope刷新→全部准备后freeze→正常password | ✅ 四UI新增payload字符串0/SQL非NULL0/登录200code200五键 | 正常手工新增UF050关闭；ImportData未覆盖，guard不放宽 |
| 四组合140正常UI/恢复/交卷/admin13-4/report | ⚠️ 三manual完整PASS；00201candidate140原deadline自然等待 | 563按钮键盘Enter保存，不冒鼠标click；423逐响应＋一140聚合证据 |
| 自然0/3/140、20min提示/竞态/native/finally | ⚠️ 原1500秒未变/观察活跃；三native各双eventtimeout；owned清理pending | 未验不标完成；SCP非native、unique不证明真实竞态 |

[完整进度](management-traits-four-real-verification-20261003.md#L3)、[安全PK及UI收据](../scripts/test/results/uf050-ui-f61006041501/browser-progress-0432.json)。

## 2026-10-06 UF-050已批准发布的上传前门禁

| 分支 | 本轮实证 | 限定 |
|---|---|---|
| staging身份/旧运行指纹 | ✅ SSH握手exit0/vm-ubuntu-go-dev/liming/PID2002/三active/backend8fb264e/front98547 | 非旧465/source/Schema/cache/DBbackup本轮完整复核 |
| frontend-only包完整差异 | ⚠️ 393对393，385逐SHA相同，app/index部分对照通过；最终strict测试exit1/SSH读取失败 | 完整module/gzip门禁未通过，不盲发布；失败原因未留详细传输，不能称timeout |
| 原管理员正常认证 | ❌ 03:56:09ZgetInfo401/code401 | 首页不当有效身份，不造token/reset |
| 正常新增payload字符串0/SQL非NULL0/正常tester登录五键 | ❌ 未发布、未创建、未真实复验 | 原8/184/264本地证据不替远端合同 |
| frontend切换/故障回滚/精确ownedcleanup | ⚪ 未上传/切换/新实体，rollback及cleanup不需要 | 不称真实发布/清理PASS；no backend replace/no restart/no生产 |
| fullUI/natural/race/native | ❌ 0/4保持、未验 | separatephase待主协调接续，不标完成 |

[本轮机器证据](../scripts/test/results/uf050-staging-20261006/blocked.json)、[完整交接](management-traits-four-real-verification-20261003.md#L3)。已有发布批准继续有效，恢复前提后不重复索取。

## 2026-10-06 UF-050最小默认状态分支（仅本地GREEN）

| 分支 | 本轮实证 | 限定 |
|---|---|---|
| exact draft正常新增→提交 | ✅ actual SFC payload字符串0；原undefined RED1fail→GREEN | 仅reset默认一行，不服务器默认所有Create启用 |
| 编辑既存0/1/NULL/2→保存→再次新增 | ✅ 四case真实handleUpdate/submitForm保留原状态，随后handleAdd才初始化0 | 不把NULL/未知编辑行自动激活；后端Update原不写状态 |
| tester显式0准入；NULL/1/未知拒绝与旧scope保护 | ✅ 原后端状态/Gin/router专项264pass事件/0fail0skip | SQLmock，不真实DB；后端guard原样，不宽NULL |
| ImportData独立新建未赋status | ⚠️ 源码已核，本次未改/未真实导入验证 | 独立未覆盖入口，不扩大授权为批量激活 |
| staging完整UI/natural0-3-140/race/native | ❌ 未发布、完整0/4，原suite2blocked/3-4notstarted/5completed | 原cleanup0/旧465-source-PID健康仅历史，未fresh远端复查 |

[本地证据及消费者](management-traits-four-real-verification-20261003.md#L3)、[verdict](../scripts/test/results/uf050-local-20261006/verdict.json)。candidate、源题、SQL/env、Legacy、权限及旧21漂移不修改。

## 2026-10-06 UF-050完整UI新真实阻断（非业务修复）

| 分支 | 实际证据 | 限定 |
|---|---|---|
| 全部tester准备在freeze前/exact draft查询及新增刷新 | ✅ 两封闭scope、4次POST和同scope刷新200；随后4freeze200 | UF-049路径远端限定通过，不放开全库或frozen旧列表 |
| 正常UI新增不含status→冻结后tester登录 | 🔥 UF-050：真实4statusNULL、登录HTTP200/code500/noidentity；actual SFC statusundefined RED1fail/exit1 | guard保留；当前源码要求显式0，部署内部stage未直接观察，无业务GREEN |
| candidate登记/准备/开始原25min | ✅ 140题卷原deadline-started1500秒、0答/run0/report0 | 不是140click完成或自然到期PASS |
| 四完整UI/natural0-3-140/race/native | ❌ 完整0/4、0/3卷未创建、native未尝试 | 新业务block后停止，不改状态/clock制造通过 |
| 本批finally/原会话 | ✅ 4exam/4tester/1candidate/1paper及闭包0、11逐0/private0/旧465-source-runtime-Schema-cache-PID同；212Go不变、getInfo200/owned页0 | stage5本轮重新执行completed，pending空，旧失败保留 |

[本轮完整事实](management-traits-four-real-verification-20261003.md#L3)、[机器阶段](../scripts/test/results/mng-premature-timeout-a61006021801/summary.json)。

## 2026-10-06 UF-049作用域纠正（测试路径local GREEN，远端待验）

| 分支 | 已验证范围 | 边界 |
|---|---|---|
| 新增只选form.examId、queryParams.examId空 | ✅ 实际SFC本地原步骤RED2fail；请求scope为空 | 实际历史403/code1保留；另一candidate frozen1、新tester profile0 |
| 00201/00202精确owned未冻结query→新增→刷新 | ✅ 测试驱动纠正后2正向＋1原路径负向观察，相关179pass | mock外部API只证明scope，不是远端HTTP200/E2E |
| 全库及frozen实体旧读取/Schema/身份闭包拒绝 | ✅ 原Go guard/实际handler/router/service250事件通过 | 不移除403、无allowall、角色权限/旧密码读取契约不变 |
| 全部tester准备先于任一owned freeze | ⚠️ 新browser helper已有精确scope/frozen/cross-exam stop断言；未远端续跑 | 冻结后用已有专用参与者/结果API，不借名单功能扩大范围 |
| 四完整UI/natural0-3-140/race/native | ❌ 未验；stage2blocked、3/4notstarted；stage5原PASS | 无业务修改/部署/主restart/DB写；21漂移仍未归因 |

[纠正 - 2026-10-06] 下方“必须提供新业务列表”的归因过宽；本次是测试全库旧路径误入，已有exact draft API足够准备人员。冻结后人员维护未获明确合同，不默认实现。[完整证据](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-06 UF-049 实际完整UI接续（BLOCKED，未修）

| 分支 | 本轮状态 | 边界 |
|---|---|---|
| fresh正常管理员/实际部署/清理前提 | ✅ getInfo200/admin、11逐0/private0/15FK/旧465/source/runtime/cache/PID2002/三健康/backup复核 | 沿实际8fb264e/front98547，不称current21已部署 |
| 00201candidate完整登记/开始/140click-save/导航刷新 | ✅ 140真实HTTP200/auto-next/140恢复、samepaper与原1500秒deadline | 未交卷、run0/report0，不称完整矩阵PASS |
| 正常封闭人员新增后旧unfiltered list刷新 | 🔥 UF-049：实际新tester PK1，GET tester/list403；UI RED exit1/未修 | 保留failclosed，需安全专用管理闭环，不放开旧密码列表/不绕守卫 |
| 四完整UI/自然25min0-3-140/race/native文件 | ❌ 完整UI0/4、自然/native未验 | 真实业务403即停止，两个00202与0/3卷未创建 |
| 同批PKfinally/原会话 | ✅ 2exam/1candidate/1tester/1paper及11child全0/private0/465-current12-source4-Schema-cache/PID保持；原getInfo200/仅1admin页 | 无secret文件/业务源码变动/部署/restart |

[真实RED](../scripts/test/results/mng-ui-natural-native-20261006/ui-business-red.json)、[完整scope](management-traits-four-real-verification-20261003.md#L3)。旧通过与失败全部保留，未验不标completed。

## 2026-10-05T14:34:36Z 有效参与者提前自声明timeout主HTTP单项关闭

**[14:38:12限定纠正]** 14:33:52的212源码SHA相同仅该时刻；随后本地21Go SHA漂移，最终源码校验exit1保留，本任务未编辑/回滚这些文件、来源未归因。[本地漂移](../scripts/test/results/mng-premature-timeout-c90f8a172e64/local-source-drift.json)单列⚠️，不是当前源码全GREEN；[fresh远端](../scripts/test/results/mng-premature-timeout-c90f8a172e64/remote-after-local-drift.json)仍exit0/11逐0/private0/源旧465runtimecacheSchema同/PID2002/三健康，该HTTP边界PASS不变，overallPARTIAL。

| 分支 | 本轮真实证据 | 限定 |
|---|---|---|
| 正常admin及owned00201tester/真实paper凭据 | ✅ getInfo200/admin/wildcard、正常新增/正确密码五键/freeze/create/detail200/140题700项 | 清page/context mocks；凭据仅页内，原adminhome/session保留 |
| 原25min未到期自声明submitType=timeout | ✅ 唯一完整POST HTTP400/code400/manual-only固定msg/successfalse/null | 实际Submit在participant解析前拒类型；同真实有效token此前detail已验，不声称400内再次验签；非unauth/缺参数400/0答manual409 |
| SQL时限/状态/完成时间/答案/结果报告零变化 | ✅ serverEpoch1791210764→1791210798<deadline1791212251，paperstate1/testerEndNULL，140/140/700，0答/0submitted/0checked，run-dim-mod-receipt-revision-current-audit全0→0，八hash同 | 不改source/profile/frozen/deadline、不保存140题或生成报告 |
| precisePK清理及主系统不变 | ✅ cleanupCOMMITTED0/14:34:36exactPK全0，主11逐0/private0/15FK/current12-source4-465-runtime-schema-metadata-cache同/PID2002/3健康/rootroot755 | FK/SafeUpdate开启；没有PDF/上传/文件删除/pending IDs，212业务Go不变 |
| 主systemd/自然25min主HTTPexpiry/完整UI/native下载/formal-prod | ❌ 本轮未测，整体PARTIAL | 不deploy/restart/权限/cache/业务代码修改，历史报告/atomic/Worker不重放 |

[完整单项证据](management-traits-four-real-verification-20261003.md#L3)、[真实HTTP及browser finally](../scripts/test/results/mng-premature-timeout-c90f8a172e64/http-browser-receipt.json)。下方14:23历史401保留，本条仅限定关闭这一已执行单项；不是完整产品PASS。

## 2026-10-05T14:23:18Z 主HTTP单项收窄（认证BLOCKED，非业务PASS）

| 分支 | 当前事实 | 限定 |
|---|---|---|
| 有效owned参与者未到期自声明submitType=timeout | ❌ 本轮未执行；真实handler合同400/code400/manual-only | 设计已有；缺真实主HTTP＋SQL零差量；旧缺题manual409不重跑 |
| 正常管理员前提 | ⚠️ 清page/context routes后两次getInfo401/code401，首页/既有cookie-store保持 | 不猜认证根因、不伪造/重置；response事件未捕获，不称事件观察PASS |
| 当前只读基线 | ✅ 14:23:05SSH1次exit0/11逐0/15FK/private0/current12-source4-旧465-runtime-metadata-cache全同/PID2002/三健康 | 无新owned数据，cleanup_required0不是执行cleanupPASS |
| 主systemd/fullUI/自然主HTTPexpiry/formal-prod | ❌ 未测 | 主restart另需明确停机批准，产品PARTIAL；已PASS报告/atomic/恢复库Worker不重放 |

[精确scope与停止点](management-traits-four-real-verification-20261003.md#L3)、[脱敏receipt](../scripts/test/results/mng-http-boundary-20261005-142318/receipt.json)。下一只需原页正常登录后真实getInfo200，单项执行前重新核基线及exact-cleanup能力；业务/fixture零修改。

## 2026-10-05T14:14:25Z atomic＋真实恢复库Worker余项关闭（限定PASS，产品PARTIAL）

| 分支 | 本轮真实状态 | 边界 |
|---|---|---|
| 真实故障INSERT/完整回滚/NULL/正常retry/duplicate | ✅ typed MySQL1644/45000/1次，四表增量0、完整facts SHA不变、ownerNULL；同140retry1/13/4/1/duplicate零增量 | 新有界fixture1/3，原FAIL及原NULL标量负向保留；仅恢复库 |
| RunExpiry独立进程startup/offline/restart | ✅ PID28610未到期首扫→Kill/Wait→自然到期→PID28616立即1023ms全部结算，原deadline不重置 | 第一被杀测试不计PASS；非主systemd重启 |
| Worker 140/0/3与报告门禁 | ✅ full140 completed50，0/3incomplete正式NULL，各SQL1/13/4/1，render0/auto-report0、receipt绑定复核 | 子Worker真正完成后再做幂等Scan，不以父Scan替child |
| RunExpiry/manual实际双pool竞争 | ✅ 真实Worker scan＋PK锁已观察，事务结束屏障后取消/join，唯一timeout1/13/4/1，重复同run | 非主HTTP/跨主服务进程竞争 |
| 主HTTP/systemd/fullUI/coverage/formal-prod | ❌ 本轮未测 | 此前报告/DataSHA PASS不重跑或撤销，产品总验收PARTIAL |
| 当前finally/安全边界 | ✅ 14:14:25exit0/库trigger-grants-upload-child0，主11逐0/private0/旧465/source/runtime/PID2002/3health/cache755保持 | cache targetstat前后同、整树匹配approved-after；未采本轮整树before，业务212Go不变 |

父Go1pass0fail0skip/125.76秒/exit0、五CASE标记PASS；todo2仅本limited scope完成，主协调者接收，不修改workflow状态。[完整真实证据](management-traits-four-real-verification-20261003.md#L3)、[summary](../scripts/test/results/mng-observer-scope-20261005-140715/summary.json)。

## 2026-10-05T13:58:26Z atomic/Worker余项接续（FAIL/PARTIAL）

| 分支 | 本轮实证 | 限定 |
|---|---|---|
| fresh只读前提 | ✅ SSH首次0、11逐0/private0/备份gzipSHA/current12/source4/旧465/runtime/PID/cache/三健康 | 不重跑已PASS报告/UI |
| 原NULL标量扫描cause | ✅ 真实sql_null_to_time/*fmt.wrapError＋独立SQLNULL，4.01s有效RED | nullable结构修正存在，最终未执行到该处，无GREEN |
| 140题故障errno/完整零写/retry/duplicate | 🔴 save140/34411ms后新增driver观察门禁FAIL，37.33s exit1 | outerdb与service独立registry，真实errno未取得；完整断言/retry未执行，非业务bug已证 |
| child Worker startup/restart/缺答NULL/竞争/deadline | ❌ actual child0/余项未执行 | Row观测也挂错registry，不能把future代码或可信Submit历史当RunExpiryPASS；mainHTTP未测 |
| 精确finally/主系统不变 | ✅ 两restore/trigger/grant/payload0，13:58:26final0/11逐0/private0/465及current指纹/PID/cache完整stat保持 | cleanupPASS不覆盖testFAIL；fixture3/3停止 |

[新终验](../scripts/test/results/mng-atomic-worker-20261005-134523/final-readonly.json)、[完整证据与停止边界](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-05T13:32:39Z atomic/Worker续验连接阻断

| 分支 | 本轮状态 | 边界 |
|---|---|---|
| fresh只读staging基线 | ⚠️ SSH初试及唯一重试各255/连接前timeout，远端脚本未启动 | 当前11表/private/restore/备份/465/源/cache/健康/PID未复核；不沿历史填PASS |
| atomic unknown根因/有效RED/零写回滚＋retry | ❌ 本轮未执行 | 原失败点是fixture直接end_time Scan(**time.Time)，cause未证；新3轮授权使用0/代码修改0 |
| Worker独立进程startup/restart/expiry-manual竞争 | ❌ 本轮未执行，子进程0 | 不是主systemd重启；旧可信Submit竞争不替此分支 |
| 本轮资源及清理 | ⚪ 新owned资源0/cleanup_required0 | 非当前远端cleanup0；停止第三次SSH/HTTP绕行，不部署/重跑报告 |

[本轮失败receipt](../scripts/test/results/mng-atomic-worker-20261005-133239/ssh-readonly.json)、[详细接续边界](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-05T08:31:50Z 唯一 authenticated HTTPzero 分支完成

| 分支 | 本轮已验证状态 | 限定 |
|---|---|---|
| 当前管理员真实Bearer/getInfo与保留会话 | ✅ 200/admin/wildcard/cookie-store相等，原页清mocks/finalindex200 | 无伪JWT/Redis/reset/logout/凭据落盘；不沿历史401跳过 |
| 00201 tester正常登录/140冻结/零分选项 | ✅ actual isOpen2/正常密码五键，manifest方向＋真实sourceOptionId，100raw1/40raw5→140final1 | 不用全raw1/随机显示序/旧isOpen0；源题/版本不改 |
| manual完整零分持久化 | ✅ HTTP140save/readback/completed，1run13dim4mod1receipt；13/4分非NULL精确0/独立Fraction | 不是incompleteNULL，不重跑生命周期或四560 |
| 管理generate/view/download | ✅ 唯一UI生成确认HTTP200/code0/attempt1，UIview/download200/PDF/660779bytes，与DB/private/scp同SHA | native下载事件timeout未落盘，未全参与者UI；审计generate1/read2 |
| 独立新report DataSHA | ✅ persisted原UTF8 9790bytes独立Node/Python SHA匹配DB/HTTP，绑定同owned run/paper/exam | 未重序列化/改模型或API/服务自check；只关闭本新report，不追溯旧四份 |
| 零分报告五灰环/第六图/36客户段 | ✅ LO24.2九页/8嵌入字体，各360/360-whitehole1.0-blue0、0.00字号12/10pt、36全文缺失0 | 内部DOCX/OPC本轮未capture，不称全PDF几何/所有全文视觉布局全验 |
| exact cleanup/currentbaseline/旧465 | ✅ 08:29cleanup0、08:31final0/11逐0/private0/源4及current12/runtime/cache相同，3health/PID2002 | 首次仅owned游标0644门禁FAIL保留，收紧0600后PASS；共享权限不改 |
| atomic完整/Worker尾部 | ⚠️ 原unknown/contextnone三轮停止仍未完成 | 本轮未尝试、不改预算/fixture、不重复问授权、非wholetaskDONE |

[真实HTTP/PDF/SQL与精确清理证据](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-05T08:15:43Z 恢复库七项服务分支关闭，atomic/Worker仍欠

| 分支 | 本轮状态 | 边界 |
|---|---|---|
| 到期140/manual与Worker可信Submit双pool | ✅ 140保存34968ms/总体50、created1/reused1、1run13dim4mod1receipt/timeout | 非主HTTP、非RunExpiry扫描竞争；25min不改 |
| 到期0/3不完整/继续fill/报告 | ✅ 正式分NULL/render0，expired及completed写拒绝 | 真实positive_app恢复库服务，主库无写 |
| resume<=300s/purpose/cross paper-candidate-exam/签名expiry | ✅ 服务子集、detail原deadline | ❌ admin HTTP/自然等待5min/UI未测 |
| exam.end / retired / revoked | ✅ 已有保持deadline/新拒、retired已有写成功、revoked答案及revision增量0/render0 | 不含source撤销与写入并发 |
| failureatomic | ⚠️ 实际trigger提交失败后run/dim/mod/receipt增量0、paperstate1/submitted0；end_time读取unknown/ctxnone→exit1 | 不能全PASS；人员NULL及retry未验，三次夹具编辑达限 |
| offlineWorker/独立进程restart/主systemd | ❌ 前序夹具失败未执行，新WorkerPID0 | 不以可信Submit双pool替实际Worker启动重启 |
| 原210.08秒actualsave原因 | ⚠️ 未取得历史cause，不猜timeout | 本轮context210/外限240保持，两个新失败ctxnone |
| HTTPzero/当前认证/独立DataSHA | ❌ getInfo401，HTTPzero/UI SKIP；DataSHA仍UNVERIFIED | 四文件历史GREEN不可替本轮HTTP |
| 恢复库及主基线finally | ✅ 两库/GRANT/payload0；11表逐0/private0/465/源/当前12表/runtime/权限不变，PID2002/3health | 本轮真实08:15:43Z只读exit0，不覆盖测试exit1 |

详见[完整逐case与安全失败证据](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-05 新有界验证切片（测试修正，不改handler）

| 分支 | 本轮状态 | 限定证据 |
|---|---|---|
| 综合oracle精确零十二位字面量 | ✅ 原0E-12词法RED→唯一format('.12f')→四文件GREEN/exit0 | Fraction/数值/等级/阈值不改，历史三轮FAIL保留 |
| 四文件Fraction/144段/OPC/六图 | ✅ 4case/0error，各9页/36段；零五灰环各360/360 | 旧服务器恢复库证据离线重放，非本轮HTTP/服务运行 |
| 独立报告DataSHA | ❌ 模型json不导出snapshot，保持UNVERIFIED | 不能用内部服务自check替独立证明 |
| 实际原admin getInfo/HTTPzero链 | ⚠️ getInfo401/code401，cookie/store相等、首页保留 | HTTP保存/submit/generate/view/download/UI本轮SKIP；不自签/reset/logout |
| 生命周期210秒actualsave根因/预算 | ⚠️ 旧本地日志无contextcause/phase时间；SSH两timeout255 | 不猜timeout、不盲加预算/重跑/改业务；fixture未改 |
| expiryfull/manual+Worker竞争/offlineWorker/restart/其他生命周期 | ❌ 本轮SKIP | 认证与SSH阻断；旧0/3与普通mixed不能替代 |
| 当前主11表/465/源/非own/权限/健康终验 | ⚠️ 本轮SSH未建立，当前未复核 | 新资源0/cleanup_required0，不称远端cleanup0；共享权限无操作 |

[本轮完整事实与旧失败限定纠正](management-traits-four-real-verification-20261003.md#L3)，oracle已完成但整体PARTIAL/BLOCKED；formal/prod仍禁止。

## 2026-10-04 MT-REPORT-DIAG-01：LOCALONLY诊断矩阵

| 分支 | 状态/证据 | 边界 |
|---|---|---|
| schema与load失败分别标记 | ✅ 真实service/sqlmock缺结构503sentinel、run查询mysql_1205，stage各schema/load | schema深层DBcause已丢失未重构，validation不等于已定位 |
| render未知/期限/真实LO坏输入 | ✅ stage=render，unknown/context_deadline/lo_input_name | 不是staging409根因或目标LO真实转换成功 |
| write真实OS失败 | ✅ WindowsPathError优先os_path、原返回sentinel/旧file保持 | 非权限chmod变更；permission按errors.Is隔离注入 |
| reload/revision/persist三表及两个事务commit失败 | ✅ 实际有序sqlmock/ROLLBACK或commit失败、新file清理、旧PDF保持 | SQL/锁/调用顺序不变，无真实MySQL写入 |
| wrapped/nested/unknown secret与任意stage | ✅ safe类型/数值类别，panic Error从未调用；任意stage不emit | 新增诊断日志合同，不代表全仓GORM日志安全审计 |
| success与download/helper nil诊断 | ✅ successful generate零日志；下载/helper原回归通过 | API/DTO/config/auth不变 |
| LO固定message/cause/类别及隔离目录 | ✅ local配置/输入/排队/命令typedcause/deadline/缺PDF/非PDF/清理 | 所有已知类别保留，chmod/read/write每条OS注入❌未全覆盖 |
| marshal/复读bytes漂移/竞争cause优先序 | ❌ 未新增故障注入 | 分支已固定标签，不能称覆盖率百分比提升或全分支覆盖 |
| 原staginggenerate409与oracle/生命周期 | 🔥 根因UNVERIFIED；仍BLOCKED | 本轮无远端操作、不修oracle/budget/业务 |

有效RED1/14→专项77/0/0，独立只读reviewPASS；全Go6144pass/652顶层/0fail/9原skip/parse0/exit0，Windowsbuild/vet各0/零error、限定diff-check0，非覆盖率。[日志合同与所有消费方](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-04T07:54:06Z 正常HTTP评分分支限定关闭；报告生成实际FAIL

| 分支 | 本轮实际状态 | 边界 |
|---|---|---|
| 原正常admin getInfo、两tester密码登录、candidate configured五键 | ✅ getInfo200/admin/wildcard、四正常身份 | 原会话保留，无伪造；首candidate工具误带adminheader拒绝后按实际匿名合同纠正 |
| 四140/700、固定shuffle/25-20点、缺题manual、560save/readback | ✅ 4missing409、560正常保存、4completed | 非mock；源V/raw映射，不按随机显示序 |
| 普通repeat与mixed首次双HTTP竞争唯一run | ✅ 4run52dim16mod4receipt，Fraction全事实PASS | 不是expiry-submit竞争 |
| 正常admin生成首份all50 TEST报告 | 🔥 HTTP409/610.78905ms，report/file0，底层根因未记录 | 不修业务、不重试，其余PDF/灰环/全文SKIP |
| 真实UI | ✅ 一个candidate恢复/save-next/submit及admin1行13/4详情 | 新开始准备按钮/四全UI/报告查看下载未验 |
| 生命周期/综合oracle | ❌ 本轮failfast；历史三轮oracleFAIL和210秒fixtureFAIL保持 | 不改预算、不等价新建绕限 |
| exact-owned cleanup/currentbaseline | ✅ cleanup0/11表各0/465逐SHA/旧12表及runtime相同 | 三服务健康，无部署重启/production |

证据：[本轮完整报告](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-04 零分发布/隔离恢复验收（未覆盖完整HTTP/UI）

| 分支 | 本轮实际状态 | 限定证据 |
|---|---|---|
| exactzero五环呈现、正常0.00字号 | ✅ 目标LO24.2隔离真实PDF9页，灰环5个各360/360，白洞1.0/blue0 | 主backend已发布；[限定灰环receipt](../scripts/test/results/mng-zero-staging-8a88d9d20f42d4f7/output/standalone-zero-gray.json)，不是HTTP |
| 4组合保存/普通重复/mixed双连接唯一 | ✅ 恢复库公共服务560save/4run52dim16mod4receipt/4PDF三SHA | 正常身份服务凭据仅内存；非UI/adminHTTP |
| 到期0/3已答，incomplete正式分NULL/报告render前拒绝/继续fill拒 | ✅ 恢复库真实SQL，0/140和3/140，timeoutreceipt | 只owned一致时间fixture；不写主库/原来源 |
| 到期140已答与到期交卷竞争 | ❌ 后续填答210.08秒actualsave失败，整体exit1 | 失败保留，原因未记录底层error，不以mixed普通竞争代替 |
| end/retired/revoked/resume/Worker实际restart | ❌ 前序失败停止；adminHTTP/UI另缺正常会话 | 昨日证据仅历史，不说今日重跑 |
| 四PDF精确Fraction/144段/所有OPC独立门禁 | ⚠️ 新oracle3轮后仍FAIL：第1份50/36段PASS，零分期望0E-12字面量错误 | 实际Go0.000000000000；不再改oracle或source/阈值 |
| 主11表0/465旧PDF/当前指纹/模板配置 | ✅ 最后SSHexit0、restore/payload0、3svc/healthok | 仅本轮currentbaseline；历史漂移保留 |

完整FAIL/PASS/SKIP与新backendSHA见[当前报告](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-03 MT-ZERO-RING-01（RED登记，GREEN由主代理收口）

[完成补充] 以下初始RED登记保留；本地现已✅：精确零/等值零、near-zero非零、25/50/75/100、mixed条件填充、比较图及正文非值XML/所有其他OPC/模板SHA保护。原生11专项与显式四冻结DTO重放13事件PASS；真实LO八代表及四客户全文9页/144提取段PASS、独立review PASS。零环灰色门禁每个360/360角度桶、白洞100%，数字五槽不移动。全文可提取不代表全章节顺序/无遮挡；cross-renderer布局差异明示。全Go6090/0fail/9原skip、build/vet0，非coverage；staging未发布该fix/生命周期❌保留，详见[完整本地报告](management-traits-four-real-verification-20261003.md#L3)。

| 分支 | 登记 | 有限合同 |
|---|---|---|
| 五语义TEST环图ChartScore精确零（含等值零） | 🔥 RED | 仅余量点idx1直接noFill→E7E6E6，0/100与五标签0.00不变；all-zero十三维同为0 |
| 非零near-zero(.004显示0.00)、25/50/75/100 | ❌ 待原生回归 | 全部非值chart XML与原模板精确相同，不用display/float/epsilon判断零 |
| mixed部分精确零、部分near-zero/正数 | ❌ 待原生回归 | 只允许对应零分语义环图局部填充，其他物理环图不变 |
| 比较图、正文非文本XML、其他OPC与模板SHA | ❌ 待原生回归 | 只原有数据/文字替换；字体/坐标/线条及模板-v2不变，原SHA输入锁不放宽 |
| 显式输出实际DOCX与LO/PDF oracle | ❌ 主代理后续视觉验证 | opt-in绝对目录、拒覆盖；本worker不运行Python/LO/browser/DB/SSH或全量门禁 |

## 2026-10-03 MT-REAL-TESTER-SCOPE 本地parser回归闭合（actual未验）

| 分支 | 本地证据 | 边界 |
|---|---|---|
| 完整11表warm私有callback、正确密码真实LoginForm→guard audit nested→五键/token绑定 | ✅ Gin/sqlmock auditDriver1/orphan1/旧写0 | 下方真实stagingFAIL保留，不能无callback DB替代 |
| wrong password先拒绝，复合revision/run孤儿存在 | ✅ lookup1/scope0；孤儿1行/nested0/写0 | 无权限/孤儿门禁放宽 |
| SELECT嵌套/括号/alias遮蔽/literal-comment问号/IN多值/500与501typed批次 | ✅ 实际guarded GORM回归 | 非通用SQL语法/性能验收 |
| canonical名碰撞、correlated父类型、ID/type逆序单原子 | ✅ 新有效RED→GREEN | 原容量/UTF8/ASCII/未知ID/数值拒绝保持 |
| candidate36/tester4混合OR，两原子顺序及各自超预算 | ✅ 六子case，合法到driver/非法零driver，独立终审PASS | 当前生产type→id原子合同；任意混合逆序不宣称 |
| 四真实完整链/PDF/生命周期与fresh实际backend | ❌ 本轮禁止SSH/browser/deploy，待交接者复验 | 不用6081本地PASS或历史cleanup覆盖actualFAIL |

最终专项42/0/0、四包5369/0/4、全量6081/0/9、build/vet0；非coverage百分比。所有共享消费者同步判定见[本地修复交接](management-traits-four-real-verification-20261003.md#L3)。

## 2026-10-03 MT-REAL-TESTER-SCOPE：真实四组合停止证据

| 分支 | 当前证据 | 边界 |
|---|---|---|
| 现有正常admin cookie/getInfo，清拦截 | ✅ HTTP200/admin与wildcard true | token只内存，会话保留，不自签 |
| 两code×两身份合法新配置及freeze | ✅ 四profile/两bundle，各140mapping，原manifest/mappingSHA一致 | isOpen开放1/封闭2，不使用交接0 |
| candidate仅配置name＋typed空串，140/700新卷、重复create、服务器25/20点 | ✅ 00201真实五键身份/同paper题序deadline | 保存/完整交卷/评分/PDF/UI未验 |
| 未到期缺题manual submit | ✅ 实际409，run0 | 仅此负向，不冒称其他矩阵PASS |
| TryTesterIdentity使用guarded s.db访问audit嵌套participant作用域 | 🔥 00201真实身份失败；guard249 data rejected | 不修改业务；nested SQL参数scope需backend隔离RED确认 |
| 00202身份及四组合完整答案/报告/生命周期 | ❌ 实际错误后STOP/SKIP | 不以freeze成功或mock代替 |
| 精确owned事务清理及历史不变 | ✅ residual0/11表逐0/旧12表指纹及465PDF不变 | 仅清理PASS，不能覆盖身份FAIL |

详见 [真实续验及交接](management-traits-four-real-verification-20261003.md#L1)。历史凭据阻塞与本地GREEN边界保留，不删旧账本。

## 2026-10-03 MT-DEPLOY-PROBE（工具合同，最小staging验收）

| 分支 | 实际证据 | 边界 |
|---|---|---|
| participant先严格参数binding再token；完整四body无token401／空body400 | ✅ 真实注册handler7子项及staging内部/公网；工具RED→GREEN | 未改变业务鉴权/返回；无token正向链未执行 |
| 已完整安装11表：pre严格完整应用账号gate，不重DDL | ✅ 5fresh实例11/159/67/21及缓存4queries；first/repeat只读签名相同 | 所有11表行0、15FK；不放宽部分结构 |
| 私有600证据副本恢复原server uid/gid/mode/mtime | ✅ 独立metadata保存＋实际文件行为测试 | 本次部署成功未触发真实回滚；原失败证据保留 |
| fresh SHA复用／atomic bin/front/env／两阶段guard drains | ✅ PID/cgroup/8092/应用DB连接0、393逐SHA、配置及private写读PASS | 不以编译代替业务验收 |
| 主旧12表／465旧PDF保留 | ✅ fingerprint/逐文件SHA、共享configs tarcompare0、11sidecar逐表0 | 不改旧源、不历史重算 |
| 四组合／profile创建／guardcapture／真实TEST PDF／Worker等 | ❌ 本阶段明确未执行；精确交接见报告顶部F | 不伪造管理员会话；安全登录凭据尚待内部供应 |

本轮DEPLOYEDYES仅部署与只读合同；前次DEPLOYEDNO及其他历史矩阵保留。报告：[当前部署证据](management-traits-staging-deployment-20261003.md#L3)。

## 2026-10-03 MT-GUARD-AUDIT：11表实际audit闭包

| 分支 | 状态/本地证据 | 边界 |
|---|---|---|
| 完整11表实际audit五列、空scope/AllLegacy空 | ✅ DDL-derived真实GORM回归 | 不伪造paper_id、不删audit |
| capture/public identity在完整11表可评估保护 | ✅ RED→GREEN | ❌ fresh恢复库actual capture尚未复验 |
| audit.report_id→revision.id→run.id→paper/exam/participant，question/PDF→stored owner | ✅ 七scope及绑定值回归 | 不根据repoCode或源题推定运行身份 |
| 同exam未触历史兄弟paper、全缺失/seven-core旧兼容 | ✅ 精确scope及原回归 | 不扩大整exam保护、不改原DTO/路由 |
| audit孤儿revision/run/复合身份错配，empty/all/mismatch scope | ✅ 全局LEFT JOIN拒绝，不裸false,nil | 真MySQL孤儿拒绝未做，不关闭FK造数据 |
| 报告canonical部分安装、列/type/NULL/collation/索引/FK漂移 | ✅ 原11表完整validator复用及负向测试 | 失败关闭，不改001/15FK/缓存规则 |
| 未知report-only表、额外/重复/注入列、意外表名、DB query/row/scan错误 | ✅ 受控拒绝与原ASCII检查 | 不暴露DB/凭据 |
| 1001 typedowner按500+500+1、最多1000绑定 | ✅ 实际参数/SQL匹配 | ❌ 大audit表EXPLAIN/延迟未验 |
| 外部真实恢复库回归事务ROLLBACK/零残留 | ❌ 环境缺失明确skip | 新TestManagementTraitsGuardAuditMySQLExternal，仅owned schema；不SSH/部署 |

新增62pass/10顶层/0fail/1skip、四包5327/0fail/3skip、review PASS。主代理最终全量6039pass/638顶层/0fail/8skip、解析错误0/exit0；Windows build/vet及fresh Linux server/test编译均0；非覆盖率，八skip不计环境通过。原freeze租约/异步handoff及所有旧写入口不变，完整消费方见[报告](management-traits-local-implementation-20261003.md#L3)。

## 2026-10-03 MT-SCHEMA-ALIASES：元数据驱动标签

| 分支 | 本地证据 | 边界 |
|---|---|---|
| 无AS的TABLE_NAME/ENGINE大写→struct零值；有AS→正确字段 | ✅ 独立驱动对照真实GORM扫描 | actual远端两标签来自上阶段，本轮不SSH |
| tables/columns/statistics/FK22字段＋capture/guard4字段严格alias | ✅ 从执行SQL生成标签，四gate查询全部消费、独立去alias整句一致 | 不改WHERE/JOIN/参数/排序、不casefold validation |
| 合法/精确旧兼容、两次gate只查一次 | ✅ 两正例/capacity发布 | 非actual完整MySQL gate |
| type/NULL/collation/indexprefix/FK action/order漂移 | ✅ 六反例拒绝、失败不发布capacity；旧严格矩阵保留 | 不放宽11表15FK/UUID/ASCII目标预算 |
| identity capture/legacy guard表列结构体 | ✅ 三查询alias、实际capture读取/保护命中 | scalar COUNT按位置安全、未改 |
| fresh actual恢复库first/repeat生产gate | ❌ 本轮未执行 | 下一主代理fresh validator，不能以本地GREEN解除staging阻断 |

新回归RED3/12/0→GREEN15/0/0（3顶层），review PASS；四包聚焦5265/0fail/2skip、258顶层；全量5977pass/628顶层/0fail/7skip，解析错误0/exit0、build/vet0。DDL SHA保持，不处理旧EOF，实际演练分支仍❌。见[详细receipt](management-traits-local-implementation-20261003.md#L3)。

## 2026-10-03 002旧opaque引用兼容（MT-SCHEMA-LEGACY-002）

| 分支 | 本地证据 | 边界 |
|---|---|---|
| exact source64→旧桶32、repo MB3-general64→MB4-0900引用64 | ✅ 实际metadata三种组合/失败粘滞/fresh刷新 | 只白名单旧边；其他容量/NULL/type/charset/collation漂移拒绝 |
| actual source/option31、32通过；33/Unicode/非法UTF8/空拒绝；repo64ASCII通过、65/Unicode拒绝 | ✅ Raw source700行、两code manifest/mapping及140冻结不变 | ASCII32/64字符=32/64字节；原JOIN/LIMIT不改、不CAST |
| writer首个INSERT前全部源ID预检；合法140题/700桶 | ✅ 非法只有BEGIN/ROLLBACK零INSERT，正向批量writer通过 | sqlmock，不是真实MySQL |
| loaded/JSONmapping-options/selected/checked CASE目标预算与Unicode拒绝 | ✅ 31/32/33及Unicode各路径实际调用 | checked参数BEGIN前核验；不改公开API或正常opaque字节合同 |
| public完整run只读、new11tables/15FK/索引/UUID36下限、原DB无callback污染 | ✅ 正向50分run、逐列漂移与70短UUID拒绝、原DB Unicode查询通过 | 真实安装/并发/远端环境❌未验 |

新增613pass/10顶层，全Go5962/0fail/7既有skip、focus5250/0fail/2环境skip，build/vet0；非coverage百分比。无DDL/env/template/frontend/sharedold/SSH/deploy改动，见[本轮证据](management-traits-local-implementation-20261003.md#L3)。

## 2026-10-03 最新分支收口（本地实现，环境待验）

| 分支 | 最新证据 | 未关闭边界 |
|---|---|---|
| sourcebundle锁／统一事务锁序／报告末次复核 | ✅ 真实unit／sqlmock PASS，历史非锁定竞争代码blocker已闭合 | ❌ 真实MySQL双连接撤销竞争未验 |
| paper及source锁等待后凭据到期 | ✅ 真实Gin／sqlmock HTTP401、ROLLBACK、零答案写入／不复用结果 | ❌ 真实网络／MySQL锁等待未验 |
| expiry Worker keyset、10batch有界、绕fail继续 | ✅ 本地行为测试PASS | ❌ 实际重启／离线／并发结算未验 |
| singleton DI／实际引用capacity／失败缓存fresh刷新 | ✅ 本地行为矩阵PASS | ❌ 11表15FK真实安装／DDL first-repeat未验 |
| admin5min续答／短凭据不升级／冻结绑定deadline不改 | ✅ 本地unit、前端355项、新47case mock PASS | ❌ 真实admin签发→受测者→SQL／PDF联合E2E未验，不匿名手机号恢复 |
| 两题本×两身份、完整保存／提交／下载 | ✅ 本地production bundle mock：560save／8submit／4download | ❌ 非真实DB／报告生成；不能称四真实组合通过 |
| 原source10及LO-compatible17视觉合同 | ✅ 主代理本轮真实SVG／pixel／六图／36text／13detail、exit0 | ❌ 目标服务器验收未完成 |

[纠正 - 2026-10-03] 下方未实现／RED标记保留历史，以上仅限定codeclosed与本地验证；环境仍PARTIAL。SSH三次timeout255、remote command未启动，publichealth200 ok；无远端备份／演练／DDL／部署／cleanup。恢复SSH后执行已授权完整顺序，见[最新receipt](management-traits-local-implementation-20261003.md#L3)，不假终结待办。

## 2026-10-03 实际依赖／引用门禁闭环（本地）

| 分支 | 证据 | 边界 |
|---|---|---|
| 默认handler两request、注入candidate/tester/runtime六request、Worker同实例 | ✅ Gin/sqlmock完整预检querycount=1及main实际装配 | 完整Schema缓存共享；既有动态身份scope/capture探测仍保留，不宣称所有SQL都只有一次 |
| 首次Schema错误/缺表粘滞，fresh实例刷新 | ✅ DI及schema行为矩阵 | 不自动迁移，不borrow全局缓存 |
| 注入错DB/secret/预算/显式nil；空cfg后改密钥不可复活 | ✅ 双request拒绝、无legacy写/错库SQL/token | 公开响应与旧构造器调用兼容 |
| 父最大容量→exam/profile/participant/source全部引用，列NULL/charset/collation/15FK | ✅ metadata扫描与各列矩阵 | exam外部opaque1..64；生成父UUID36..64；DDL未执行 |
| 实际query/JOIN/读/写/model加载/JSONmapping-options/Rawsource字节预算 | ✅ 缓存父容量拒绝超长、UTF8及跨profile引用；Raw冻结前拒绝 | 不将静态name/index存在作为引用能力 |
| 冷s.db不可变、atomic发布、拒绝后合法request继续、原DBregistry不污染 | ✅ ReviewImmutablePrivateRegistry/Join/IdentityCount及实际DI测试 | 未执行race，不以代码review代替race工具证据 |
| public LoadValidatedRun冷/failed schema | ✅ 先同实例门禁、闭合公开报告消费者 | 私有事务fixture不代表公开越权 |

最终focus2546/0/0，全量5290/0/7、build/vet0、review PASS；sqlmock不是MySQL。下方同事项未完成标记保留历史并由本条限定纠正；并行安全链和真实环境单独验收。

## 2026-10-03 新授权容量局部receipt（其他安全分支未关闭）

| 分支 | 当前证据 | 边界 |
|---|---|---|
| paper/pq父列varchar1..35及65拒绝、36..64允许 | ✅ 原70短容量RED＋全部边界保留，全量GREEN | 仅元数据，未执行真实DDL |
| exam外部父列1..64、混合长度、collation/15FK/type/index/error/cache | ✅ 原真实GORM metadata矩阵回归 | 实际profile/exam引用ID对cached父容量检查❌未接 |
| bundle撤销与写/提交/report完成事务锁、统一锁序 | 🔥 未实现，非锁定读仍存在 | 不以FOR UPDATE字符串或顺序负例当并发证明 |
| identity/runtime/PDF共享actual cfg/DB实例，跨HTTP缓存/fresh刷新 | ⚠️ 未实现 | HTTP每请求新service仍存在 |
| admin TEST范围签发purpose短凭据，唯一有效candidate/paper/exam | ❌ endpoint及行为case未新增 | 无手机号匿名恢复；既有五字段/11前端路由不动 |
| resume retired允许/revoked拒绝/到期只状态或完成/绝不resetdeadline | ❌ 新凭据链未实现 | 需求已确认，P0 case见最新本地receipt；不冒称可用 |

## 2026-10-03 后端身份／Schema续作（BLOCKED）

| 分支 | 状态 | 证据边界 |
|---|---|---|
| configured九字段、空串typed、required为空拒绝、不泄露未配置姓名 | ✅ Gin/sqlmock RED→GREEN | 不新增idNumber/depart采集 |
| end/retired只挡新开，existing按冻结paper恢复，draft/profile漂移不覆盖 | ✅ 身份冻结恢复行为测试 | candidate需有效凭据；过期凭据恢复❌ |
| owner恰一、绑定/del_flag/tester status；冻结字段/deadline无写 | ✅ service／实际Save/LoginForm测试 | 非真实并发数据库 |
| Schema列/索引/FK漂移及query-row-scan错误、失败缓存 | ✅ metadata实际扫描矩阵 | 每HTTP新实例缓存跨请求⚠️ |
| UUID36父键容量不足 | 🔥 RED70子项＋顶层，未修 | 全量exit1，三轮上限停止 |
| 审核撤销与写／report提交竞争 | 🔥 独立review blocking，未修 | 非锁定bundle读取；不能将顺序拒绝当并发证明 |
| freeze gate排他到commit | ✅ lease行为测试 | 取消等待未独立修复 |
| 真实MySQL/四组合HTTPUI/重启Worker | ❌ 未验证 | 不SSH/DBwrites/deploy；前端并行不归本任务 |

## 002本轮基础实施（2026-10-03）

| 分支 | 状态 | 边界 |
|---|---|---|
| 运行测试模板独立预定义TEST用途SDT；真实PDF三标记可见 | ✅ MT-LABEL-02 RED→GREEN，TEST v2实际封面无新裁切 | 客户原件、原候选88控件/5槽/6图不改；新副本90 |
| Schema全缺失、列/NULL/type/index/FK/collation漂移、查询错误、缓存；5非唯一读取索引 | ✅ schema sqlmock/合同及索引RED→GREEN | 事务前每实例探测，不自动迁移；真实DB部分安装未验 |
| 报告test标题/用途字段与immutable revision/current/audit独立存储 | ✅ 双事务/原子回滚/ID下载/坏文件/旧PDF保护测试 | 真实MySQL未验，不占旧pdf_path |
| 两题本×两身份140题/700桶、V序与展示序、retired/window新开拒绝、writer回滚 | ✅ 新组卷构建/事务测试 | 不是四组合真实HTTP/UI |
| end后有效token恢复、原paper/deadline/20min点保持，retired续答，revoked拒绝 | ✅ 公开CreatePaper恢复sqlmock＋policy测试 | ❌ identity重登录仍未完 |
| server到期139/140审计详情无分/无报告 | ✅ incomplete详情RED→GREEN＋既有submit事务 | MySQL/真正Worker到期并发未验 |
| Worker启动即扫/cancel/new_creation硬scope | ✅ 新Worker测试及主服务接线 | ❌ 实际offline/restart未验；不碰legacy |
| admin／精确run/report／拒formal与任意path／ENV local-staging生产拒绝 | ✅ httptest与真实205原文环境矩阵 | ❌ 真实DB的HTTP200报告链未验 |
| 205真实原文DTO→Go→LO，26段维度及总体全文／用途、六图值 | ✅ 完整原文PDF9页，raw读序无缺词；实际封面与六图页 | 不是实际DB用户或全页/目标环境视觉验收 |
| configured字段完整保真、end/retired重登录、前端profile/新答题/admin结果及旧直达分流 | ❌ P0未完，集中交接 | 不把既有前端165项当新版覆盖 |

## MT-WORD-RUNTIME-01独立图表绑定（2026-10-03）

| 分支 | 状态 / 真实测试 | 边界 |
|---|---|---|
| 固定SHA候选数字标签3节点；总体/任务4节点且末节点为空白 | ✅ 原失败TestManagementTraitsTestWordIndependentValueOnly RED charts→GREEN | 仅第二节点分值替换，标签/单位/额外空白不变；一轮代码修复 |
| 六业务title/OPC绑定；总体28.85、自我25、人际75、任务100、发展0；13维及常模 | ✅ 逐系列逐点及五数字槽断言 | 不按图类型推测，不改评分或205文案 |
| chart除数值外XML、正文除文本外XML、其他所有OPC部件原字节不变 | ✅ 增强原真实函数测试 | 模板SHA、原style/star/footer/媒体保持 |
| SHA不符、formal DTO、13维缺失、三段建议缺失 | ✅ 原负向真实调用保持 | 无模板放宽或API启用 |
| 未知/重复键、其他节点数量、4节点非空尾字、系列/点序不符 | ⚪ 固定精确模板SHA使修改包在input先拒绝；保留内部防线，未声称各内部分支单独覆盖 | 后续扩大可用模板范围须专门负向合同 |
| 真实Go全50DOCX→本机LO→第3页五环图＋13维柱线图 | ✅ A4 9页/实际截图；混合分值PDF五标签正确、可见占位符0 | 不是DB/HTTP/UI/目标环境或全页视觉验收 |
| 测试用途标记DOCX存在但PDF文本不可见 | ⚠️ 本轮实际PDF发现，未修改标记定位代码 | 仅图表范围；不可交付/启用，后续独立确认 |

[限定纠正] 下表“独立002模板88/5/6 charts失败”为前轮历史，本次图表局部已GREEN；整体运行链门禁和其他未完成项不解除。原候选17＋source-layout10本轮重跑全部通过，optimized23未重跑。

## 002实际路由保护（2026-10-03，本地）

| 分支 | 验证状态 | 边界 |
|---|---|---|
| 实际Setup中受保护paper/save、candidate/update、tester PUT在旧handler前403 | ✅ MT-HTTP-01 RED三失败→GREEN四pass事件 | 不依赖JWT豁免猜身份；未开放新运行链 |
| 新版管理profile/result未登录401 | ✅ 实际Setup HTTP测试 | 普通JWT资源范围另验 |
| sidecar全缺失、旧空值DTO和请求体还原 | ✅ 既有GuardAbsentAndFailure / ActualDTONullsAbsentSchema | sqlmock，非真实MySQL |
| 探测失败/部分结构、全批保护及旧PDF不写/不删 | ✅ 既有实际handler与scope矩阵 | 数据库结构未安装 |
| freeze/create-paper/新HTTP保存交卷/报告 | ❌ 运行入口仍关闭 | 时间窗口策略与完整Schema门禁未闭合 |
| 客户原始文案精确SHA/205条、validated run/snapshot DTO、incomplete/损坏拒绝 | ✅ 两本地service测试 | 纯适配，不等于可信DB来源/报告授权 |
| 独立002模板88/5/6 value-only和测试用途 | 🔥 新Word测试charts失败、三轮后停止 | 未产生PDF，不放宽断言、不接运行入口 |

## 002原稿内容保真LibreOffice兼容（2026-10-02）

✅ 新17项本地合同/真实PDF门禁：原文rPr字体颜色star/footer保持；允许框/legend/pageflow白名单；独立输出保护、88/5/6绑定及确定性、五组数字/几何/像素、36长文案和13详情建议同页。旧strict10及optimized23保持。❌ 目标服务器、Word PDF、正式内容/运行链/安装仍未验；页面留白及原星级/总页脚待审。见[当前兼容切片](management-traits-word-candidate-verification-20261001.md#L3)。

## 002 Word Candidate Contracts (2026-10-01)

仅离线候选，不是handler或正式报告链覆盖。见[本地验收](management-traits-word-candidate-verification-20261001.md)。

[完成补充 - 2026-10-02 后续版式切片] 标签压环/摘要独占页/详情跨页已本地✅：23测试0失败，SVG+144DPI五组标签避碰、两DOCX各13详情/建议同页、摘要与详情共享页。前次⚠️保留为历史；最终页数/整体客户视觉/目标环境/批准仍未完成，概览及末详情页留白仍待审。

[完成补充 - 2026-10-02] 下表两项MT-WORD-01历史RED已由批准的局部排版切片取得本地✅：五组0/25/75/100/28.85各5/5数字，图例全文及标签框分离、36长文本与无尾字孤页测试通过；18项0失败。保持整体版式/跨页/目标环境/批准为❌，不删除历史。新增⚠️可见盲区：文字与环体相交、摘要页稀疏、责任心详情跨页，待模板审阅决定。

| Branch | Coverage | Boundary |
|---|---|---|
| 来源SHA、拒绝源=目标、确定性ZIP及独立输出 | ✅ 结构测试 | 不导入/激活 |
| 88实际位置SDT及5数字槽、标签边界/摘要粗体、六图literal/媒体固定层 | ✅ 结构测试 | 不批准字段语义或图层最终外观 |
| synthetic字段/图分一致，0/25/75/100不伪造原值 | ✅ 数据测试 | 不等于实际PDF完整 |
| 0/25/75/100实际PDF五图数字完整 | 🔥 RED MT-WORD-01，四组均4/5 | 三轮后停止，未交付，待候选局部版式确认 |
| 总体拆行、人际裁切、图例重叠/截短、概览尾字孤页 | 🔥 MT-WORD-01，实图失败 | 不放宽测试或宣称修复 |
| 全页视觉/Word与目标服务器共同验收/内容批准 | ❌ 未完成 | 本机实开与转换不代表目标环境或正式报告 |

## 002 Management Traits S2D Read-only Result Consistency (2026-10-01)

RED新函数缺失→GREEN原生Go 8顶层+567子项pass；编辑器484/0为另一个计数口径。仅输入自洽复核，无DB、审批、真实提交或授权结论。

| Branch | Coverage | Deferred boundary |
|---|---|---|
| 双题本/candidate-tester、原始全3/全4及方向极值、精确聚合/常模/顶底 | ✅ TestManagementTraitsResultValidationValid | 不加载真实题本/受测者 |
| 数组乱序、输入深克隆不变、返回Rat和集合隔离 | ✅ TestManagementTraitsResultValidationReorderedDetached | 不写或修复模型 |
| run.Status、身份/四版本/Q/hash/计数及overall缓存/等级篡改拒绝 | ✅ TestManagementTraitsResultValidationRunRejects | completed不证明真实交卷 |
| 13维恰完整、key/ID/所属run/order/name/module/count/sum/score/norm/level | ✅ TestManagementTraitsResultValidationDimensionRejects | 不依赖已安装索引/FK |
| 4模块完整、key/ID/所属run/order/成员数/精确缓存 | ✅ TestManagementTraitsResultValidationModuleRejects | 不新增模块等级/比较 |
| S2C上游非法JSON/hash/绑定/预算及有效但不完整输入拒绝 | ✅ TestManagementTraitsResultValidationUpstreamRejects / Incomplete | 139/140不成为正式结果 |
| Source/time/审批/历史证据不被默认为已验证 | ⚪ S2D延期；TestManagementTraitsResultValidationDeferredLifecycle | 后续独立生命周期/来源门禁 |

## 002 Management Traits S2C Strict Decode / Read-only Adapter (2026-10-01)

RED缺API→GREEN九顶层/含子项365通过；仅调用者提供模型的自洽验证，不查DB/身份来源/批准/时间状态。

| Branch | Coverage | Deferred evidence |
|---|---|---|
| 规范存储manifest/mapping重排/排版往返、domain/policy/dimensions固定核验 | ✅ TestManagementTraitsDecodeRoundtrip / NormativeAndMappingTamper | 未导入客户题本或读取数据库 |
| duplicate/escaped duplicate、unknown/casefold、null/缺false、类型/整数溢出/小数指数、尾随/BOM/UTF-8/深度拒绝 | ✅ TestManagementTraitsDecodeStrictTokens / EveryFieldRequired | 上游读取前资源限制待接 |
| 孤立代理项拒绝、合法代理对及真实U+FFFD保留 | ✅ TestManagementTraitsDecodeUnicode | 不改变客户文本 |
| 模型bundle/paper/版本/hash/profile声明、140题绑定/方向/文本/选项与逐题SHA拒篡改 | ✅ TestManagementTraitsDecodeStoredRejects / StoredValidAndDetached | 真实源题/participant归属未查询 |
| 已答选择/raw/final齐全且一致、未答全nil；输入不变/输出独立、方向处理一次 | ✅ TestManagementTraitsDecodeStoredValidAndDetached / MixedRawAndLargeBudget / IncompleteAndDeferredMetadata | 返回raw输入不是已提交正式结果 |
| 正预算、逐字段及合计精确边界/超限拒绝 | ✅ TestManagementTraitsDecodeStrictTokens / StoredValidAndDetached / MixedRawAndLargeBudget | HTTP/DB读入限额尚未实施 |
| Source/Status/evidence/人员字段/采集合同/所有时点不被当作已审核 | ⚪ 明确S2C延期；TestManagementTraitsDecodeIncompleteAndDeferredMetadata验证不推定 | 后续独立生命周期/来源门禁 |

## 002 Management Traits S2B Pure Contract / Hash (2026-10-01)

先缺契约符号RED→GREEN10顶层/含子项172通过。均为typed纯边界，不代表DB来源/身份授权/历史证据有效。

| Contract branch | Coverage | Remaining boundary |
|---|---|---|
| Manifest题本/版本语法、140唯一V/维度/方向、UTF-8题干、五项raw/文字/展示序绑定 | ✅ TestManagementTraitsContractManifestRejects / ManifestPolicy / OpaqueIDsAndSyntaxLimits | 客户题干、执行版本/批准未验证 |
| Mapping确切hash及题本、逐V原文/选项匹配、源题/源选项ID唯一且合法 | ✅ TestManagementTraitsContractMappingRejects / OpaqueIDsAndSyntaxLimits | 未查询真实源题/历史备份 |
| Input纸卷/测评/participant声明、题本/hash、140题/展示序/源题及同题选择raw一致 | ✅ TestManagementTraitsContractInputRejects | 真实归属/时限/权限/DB关联待验证 |
| V/raw数组重排canonical不变，保留原文和展示序，输入不变/输出隔离 | ✅ TestManagementTraitsContractCanonicalDeterminism | [补充2026-10-01] S2C已实现本地评分JSON严格解码/预算；HTTP/DB读取前限额仍未接 |
| 固定policy/norm纳入SHA、呈现轴不存在、UTF-8 JSON字节及标准SHA复算 | ✅ TestManagementTraitsContractManifestPolicy / CanonicalRoundtripAndFrozenHash | hash不证明内容批准 |
| 版本/题干/映射ID/人员声明/作答/展示序变化传播hash，旧hash拒绝 | ✅ TestManagementTraitsContractHashSensitivity / LinkedHashChanges | 既有run复用校验/DB写入未实现 |
| 未答空选择raw0可规范化并交S1、不出正式分；同分规则保持 | ✅ TestManagementTraitsContractIncompleteAndTieDelegation | 手工/到期运行语义未接入 |

## 002 Management Traits S2A Model Contracts (2026-10-01)

仅Go声明+本地GORM元数据/JSON/decimal测试，RED缺七个模型→GREEN九个顶层测试。不是数据库约束或运行时分支已经覆盖。

| Contract | Coverage | Deferred evidence |
|---|---|---|
| 七个独立TableName、字符串单主键、无自增 | ✅ TestManagementTraitsTableNamesAndPrimaryKeys | 无MySQL表创建证据 |
| 显式column/json/type、nullable profile/答案/score/level/time | ✅ TestManagementTraitsExactFieldsTagsTypesAndNullableSchema | 实际DDL nullable/default尚未执行 |
| 11个唯一索引名称/列序/priority | ✅ TestManagementTraitsOrderedUniqueIndexes | 数据库索引/FK/collation待验 |
| 完整JSON键和nil→null、零用时区别、历史ProfileExamID可空 | ✅ TestManagementTraitsJSONCompleteKeysAndNilValues / NullableZeroAndHistoricalIdentity | 实际人员归属/来源审核未实现 |
| decimal数值序列化、精确JSON往返和driver.Value/Scan | ✅ TestManagementTraitsJSONExactDecimalNumbersAndRoundTrip / DecimalDriverValueScanWithoutDB | 未DB写入或把缓存代替精确源 |
| 无内容/模板评分轴、无user_id身份假设、模块无未批等级比较 | ✅ TestManagementTraitsNoRenderingAxesRelationsOrFloatFields | 版本/hash运行校验未实现 |
| 纯模型，无hook/init/关联/AutoMigrate写入 | ✅ TestManagementTraitsPlainModelsHaveOnlyLiteralTableNameMethods | 旧写保护、持久化/并发/删除仍待后续 |

## 002 Management Traits S1 Pure Identity / Scoring (2026-10-01)

仅本地纯service入口，不接HTTP/DB/模板；不是既有002安全问题修复。先取得缺符号编译RED，再实现GREEN：9个顶层测试、含子测试47通过/0失败；独立新文件语句覆盖100%，不是service全包或系统覆盖率。staff/leader仅逻辑题本身份，不推定repo映射或正式批准。

| Function | Branch | Priority | Coverage |
|----------|--------|----------|----------|
| ManagementTraitsDimensions | 13固定语义维度/顺序/常模；V1～140唯一、100正40反；10/11/12/13题分母 | P0 | ✅ TestManagementTraitsDimensions_IndependentCatalog |
| ManagementTraitsDimensions | 返回Items与Rat深拷贝，调用者修改不污染未来计算 | P0 | ✅ TestManagementTraitsDimensions_FreshDeepCopies |
| CalculateManagementTraits | staff/leader保留身份、相同答案规则；输入重排不改变结果 | P0 | ✅ TestManagementTraitsCalculate_AllThreeAndQuestionnaireIdentity / AllFourExactAggregation |
| CalculateManagementTraits | 140题原始全3/全4、正反向处理一次及精确分值 | P0 | ✅ TestManagementTraitsCalculate_AllThreeAndQuestionnaireIdentity / AllFourExactAggregation |
| CalculateManagementTraits | 方向调整后的全最低/全最高；模块成员等权与13维总体等权，不题数加权/四模块等权 | P0 | ✅ TestManagementTraitsCalculate_DirectionAdjustedExtremes / AllFourExactAggregation |
| CalculateManagementTraits | 顶3/底3全维度精确排序、固定同分顺序、不按等级过滤、全同分重叠 | P0 | ✅ TestManagementTraitsCalculate_ExactRankingWithoutGradeFilters / DirectionAdjustedExtremes |
| CalculateManagementTraits | forward/reverse未答及零答：所有正式分/等级/常模抑制，空选择集合非nil | P0 | ✅ TestManagementTraitsCalculate_IncompleteSuppressesAllFormalResults |
| CalculateManagementTraits | nil/139/141行、重复/越界V、已答raw0/6/负值、未答raw非0、未知题本拒绝 | P0 | ✅ TestManagementTraitsCalculate_InvalidInputBoundary |
| ManagementTraitsLevelForScore | 精确0/10/30/70/90/100及阈值前小数；nil/越界拒绝；不修改入参 | P0 | ✅ TestManagementTraitsLevelForScore_ExactBoundariesAndFinalFormatting |
| Display boundary | big.Rat.FloatString(2) HALF_UP；69.999显示70.00仍合格、68.125显示68.13 | P0 | ✅ TestManagementTraitsLevelForScore_ExactBoundariesAndFinalFormatting |
| Deferred runtime | HTTP/DB/身份令牌/交卷锁/时限/Word/正式内容/历史重算/部署 | — | ⚪ S1不实现；仍按评估列为未完成 |

## 00401 Phase-1 Word Template V2 Contract (2026-08-13)

| Function | Branch | Priority | Coverage |
|----------|--------|----------|----------|
| Field registry | required business field is present exactly once | P0 | ✅ |
| Field registry | optional registered field is omitted or used by the template | P1 | ✅ |
| Field registry | unknown field, missing required field, or illegal duplicate is uploaded | P0 | ✅ |
| Field registry | repeatable registered field is used in more than one content control | P1 | ✅ |
| Chart registry | V2 business chart key resolves through document relationship to any physical `chartN.xml` | P0 | ✅ |
| Chart registry | chart files are reordered while business chart keys remain stable | P0 | ✅ |
| Chart registry | missing, duplicated, or unknown business chart key is uploaded | P0 | ✅ |
| Embedded workbook | `FieldDictionary` explains every registered field and declares schema V2 | P0 | ✅ |
| Embedded workbook | `ChartData` exposes stable business keys and receives runtime score/cache updates | P0 | ✅ |
| Embedded workbook | a chart uses an external workbook or an unexpected data range | P0 | ✅ |
| Backward compatibility | V1 template has no business chart keys and continues using the legacy physical mapping | P0 | ✅ |
| Template metadata | management API reports schema version, registered/used fields, business charts, workbook, and external-link counts | P1 | ✅ |
| Microsoft Word compatibility | active staging DOCX opens normally in desktop Microsoft Word without repair | P0 | ✅ V1用户实证；V2暂停 |
| Result score display | aggregate overall/group/dimension/score-sum values are null | P1 | ✅ display `—` |
| Result score display | aggregate values have 0, 1, 2, or 3+ decimal places | P0 | ✅ display exactly two decimals without changing stored precision |
| Closed tester create | ID number present | P0 | ✅ use ID number as exam-scoped identifier |
| Closed tester create | ID number empty and telephone present | P0 | ✅ use telephone as exam-scoped identifier and default password source |
| Closed tester create | both ID number and telephone empty | P0 | ✅ reject before database write |
| Closed tester list | student filter is `1` / 是 | P1 | ✅ FB-158 staging GREEN: real API returns 8/8 rows with `stuFlag=1` |
| Closed tester list | student filter is `0` / 否 | P1 | ✅ FB-158 staging GREEN: real API returns 19/19 rows with `stuFlag=0` |
| Closed tester list | student filter is empty / 全部 | P1 | ✅ FB-158: no `stu_flag` condition is added |
| Closed tester list | telephone filter is provided | P1 | ✅ FB-158: exact telephone condition is shared by COUNT and rows |
| Word-to-PDF layout | Word desktop layout is valid but Linux LibreOffice reflows floating shapes/anchors | P0 | ✅ staging real-data PDF uses flow paragraphs for all 10 dimension titles/definitions; 9-page visual review has no overlap or title reordering |
| Word-to-PDF layout | generated PDF page count differs from fixed footer total | P1 | ✅ staging template removes unreliable NUMPAGES total and retains automatic PAGE |
| Word-to-PDF layout | section page transition and explicit page break are stacked | P0 | ✅ final staging template removes two redundant explicit breaks; all 9 generated pages are non-empty |
| Word-to-PDF layout | completed report was generated before a template replacement | P0 | ✅ persisted historical PDF remains unchanged by design; user-selected paper `658dc083...` was force-regenerated and revalidated without bulk rewriting other reports |
| Word-to-PDF layout | user explicitly clicks individual or batch generate after a template replacement | P0 | ✅ FB-151 staging GREEN: both actions send `force:true`; real historical paper regenerated from current template as A4 10 pages |
| Word-to-PDF layout | dimension doughnut chart is stored outside or floats relative to its score table cell | P0 | ✅ chart3–chart12 are inline children of their matching score cells; verified by structure test and staging screenshot |
| Word report data | first-level score appears in summary row, chart, and analysis block | P0 | ✅ repeatable group score controls keep summary/chart/analysis aligned; staging user report shows `3.50/3.60` throughout |
| Word report semantics | customer chooses the original 3D overview pie despite independent-score semantics | P0 | ✅ customer decision retained; chart1 values use two decimals and do not display percentages |
| Word report profile | requiredFields contains only a subset of the six profile fields | P1 | ✅ staging report filters unconfigured cells/rows, spans retained cells, and preserves time/duration |
| Word report pagination | approved full text flows to 9 physical pages instead of the original 10-page sample | P2 | ⚠️ no content loss or blank page; final page is sparse and page-count preference needs confirmation before forcing pagination |
| Microsoft Word compatibility | template is produced by broad LibreOffice re-save instead of targeted edits to Word-native package | P0 | ✅ final template is rebuilt from Word-open V1 baseline; staging API download and filled DOCX both open in Word 16 without repair/writeback |
| Customer template Tag boundary | static `【诊断】` label is inside the dynamic diagnosis content control | P1 | ✅ local customer candidate keeps label outside a1-04 control; text/layout metrics unchanged and runtime label verified |
| Customer template chart source | charts retain external local Excel relationships | P1 | ✅ customer chose link removal; 11 external relationships and 11 externalData nodes removed while caches/style/layout remain |
| Customer template pagination | four explicit page breaks plus two next-page sections | P1 | ✅ customer confirms the 10-page composition and page-4 continuation are intentional; all physical pages contain content |
| Customer template module flow | one dimension title/definition is separated from its score/diagnosis by a page break | P1 | ✅ all 10 three-row modules use cantSplit + title/score keepNext; Word and LibreOffice exports keep each module intact |
| Customer template cross-version pagination | a first-level analysis table overflows by one line while an explicit break immediately precedes the secondary section | P0 | ✅ remove only the stacked explicit break; staging LibreOffice 24.2 real-data output matches Word at 11 nonblank pages and keeps all 10 modules intact |
| Customer doughnut chart label | Word displays centred but LibreOffice 24.2 PDF interprets manual label layout differently | P0 | ✅ keep Word template unchanged; apply chart3–12 calibration only before LibreOffice conversion and require every rendered centre delta ≤2px at 180 DPI |
| Customer template re-save | data-label font size or manual layout changes while chart identity remains the same | P0 | ✅ FB-152 requires the current long-text staging PDF to pass all ten rendered centre checks before release |
| Customer template re-save | Word restores a section-adjacent page break or changes a dimension chart from inline to anchor | P0 | ✅ structural gates reject both regressions; latest template removes one break and restores only chart7 to inline |
| Customer chart score format | visible group/dimension score labels omit explicit number format | P0 | ✅ chart1 and chart3–12 require `0.00`; real staging PDF summary and analysis both show `3.50/3.60` |
| Customer template calibration version | template coordinates/layout change after LibreOffice offsets were measured | P0 | ✅ recalibrate against the same real long-text paper; final 180 DPI rendered-pixel gate passes 10/10 |
| Customer doughnut score typography | generated chart3–12 labels inherit mixed bold attributes or move outside the hole after font normalization | P0 | ✅ FB-154 STAGING GREEN: remove native labels and use fixed non-bold centre overlays; real good/questionable reports both pass 10/10 at 180 DPI |
| Customer radar grid | LibreOffice receives radar value axes without explicit 0-5 range and one-point major unit | P0 | ✅ FB-155 STAGING GREEN: both real reports display all five concentric 1-point score levels under LibreOffice 24.2.7.2 |
| Customer overview pie | template contains manual numeric rich labels or category order differs from runtime contract | P0 | ✅ FB-160 staging GREEN: customer 3D example retained; two real reports prove dynamic outside values and fixed order |
| Customer overview pie | fixed slice position/color must match business meaning | P0 | ✅ FB-163 staging GREEN: real report proves left green=`通用能力` 3.10 and right cyan=`心理素养` 3.25; V1 cache and optional embedded workbook share the corrected point order |
| Customer report page number | uploaded template contains `PAGE / NUMPAGES` and LibreOffice reflows the report | P1 | ✅ FB-164 staging GREEN: upload rejects `NUMPAGES`, runtime keeps PAGE only; two real 11-page PDFs display exactly `1～10` with no total-page fraction |
| Customer radar labels | any category axis is deleted or score labels overlap the polygon | P0 | ✅ FB-161 staging GREEN: two real reports show ten names, ten readable values and five grid levels |
| Customer chart label styles | customer changes pie-label coordinates/position or radar score typography in Word | P0 | ✅ FB-169 STAGING GREEN: real report retains close in-chart pie positions and 12pt radar values while preserving dynamic data, left-green/right-cyan business mapping, duplicate suppression and five grids |
| Customer overview pie label overlap | template places group values inside the colored pie surface | P1 | ✅ FB-170 STAGING GREEN: deployed template keeps two template-owned point layouts, uses `outEnd` plus visible `tx1`; real report renders left green 2.90/right cyan 3.18 outside with short leader lines and no overlap while preserving dynamic data, mapping and radar typography |
| V1 chart portability | chart relationships contain `TargetMode=External` | P1 | ✅ FB-162 staging GREEN: active template is zero-link; all schemas reject external links and runtime strips relationships/externalData |
| Word-resaved template upload | Word restores external chart relationships after any edit | P1 | ✅ FB-165 staging GREEN: real authenticated upload removed 2 external artifacts before strict validation and persisted a valid zero-link DOCX; already-clean files retain original bytes |
| Phase-1 report profile layout | configured identity fields occupy sparse original left/right slots | P1 | ✅ FB-166 historical GREEN; FB-171 supersedes full-width rows with compact double columns while retaining no gaps and aligned `时间｜时长` |
| Phase-1 report profile order | customer reorders personal-information controls in Word | P1 | ✅ FB-167/171: generated identity order follows template controls; runtime now owns row compaction and two-column pairing, while template styles and order remain authoritative |
| Phase-1 report profile spacing | personal-information rows inherit different paragraph spacing | P1 | ✅ FB-168 staging GREEN: deployed template gives all eight identity/time cells exactly one spacing definition; real subset/all-field reports visually pass |
| Phase-1 report profile double-column flow | configured identity fields should use both columns without blank rows | P1 | ✅ FB-171 STAGING GREEN: real `name,age,telephone,gender` report renders name full-width, age+gender paired, telephone left, and submittedAt+userTime paired with no blank rows |
| Phase-one validity notice | validity status is good | P0 | ✅ FB-153 GREEN: the whole “提示” paragraph is absent in Word/PDF and Vue reports |
| Phase-one validity notice | validity status is questionable | P0 | ✅ FB-153 GREEN: retain the whole “提示” paragraph and approved questionable text |
| 260915 template data binding | every visible B/C sample value has the correct stable content-control Tag | P0 | ✅ FB-172: 75 controls/58 tags cover profile/header, overall, three modules, five selection slots and ten dimension details; communication has its own stable identity |
| 260915 chart portability | all 12 charts use stable business keys, cached values only, and zero external relationships/formulas | P0 | ✅ FB-172: chart relationships, externalData and workbook formulas are removed; 12 chart types/series/data-point counts remain fixed and Word opens normally |
| 260915 cross-renderer pagination | customer Word 10-page design converts to exactly 10 nonblank A4 pages in local LibreOffice | P0 | ✅ FB-172 local: four safe blank cover paragraphs removed, spacing/row metrics reduced to 70%, detail modules cannot split; Word and LibreOffice both produce 10 pages |
| 260915 customer style ownership | template repair preserves all customer media and chart styles while only data caches change at runtime | P0 | ✅ FB-172: all customer media names/bytes remain identical; no color/font/border/chart-type replacement; future style edits remain Word-owned and contract-tested |

| Area | Branch | Status | Notes |
|------|--------|--------|-------|
| MBTI full report generation | document.xml contains static body runs with w14:textFill / w14:props3d | ✅ | Triggered by production tofu-box issue; now covered by FB-042 fallback |
| MBTI full report generation | document.xml contains risky static body font families such as HYYakuHei / 汉仪雅酷黑 | ✅ | Covered by FB-043 font-family normalization fallback |
| MBTI full report generation | document.xml contains East Asian static body runs with only w:hint and no explicit font family | ✅ | Triggered by production ESTP "功利型/凭借" tofu-box issue; covered by FB-044 |
| MBTI full report generation | styles.xml / fontTable.xml declare unstable CJK fonts used by body fallback | ✅ | Covered by FB-044 style/font-table normalization |

## RuoYi Administration Stub Inventory

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| User profile | update profile/password/avatar and query/update assigned roles | P1 | ✅ | FB-088～092: authenticated self-service, bcrypt/session invalidation, decoded safe avatar and permission-checked transactional role replacement |
| Role authorization | status/data scope, allocated/unallocated users, cancel/select authorization | P1 | ✅ | FB-091～092: exact RuoYi permissions, super-admin protection, bounded paging and transactional validated relationships |
| Public registration | `POST /register` | P1 | ✅ | FB-093/095: config gate, atomic one-time captcha, strict credentials, bcrypt, transactional optional common role and unique-index migration |
| Cache administration | config/dict refresh | P2 | ✅ | FB-096/097: real permission-checked SCAN refresh handlers replace both success stubs and preserve login/captcha keys |
| Monitoring administration | force logout, jobs/job logs, operlog/logininfor cleanup | P2 | ✅ | FB-098/099 retire online/job/jobLog modules and unsupported audit mutations; real server and audit list routes remain |
| Code generation | `/tool/gen/*` | P3 | ✅ | FB-098/099 remove the unsupported backend routes, frontend module, hidden edit route, and menu access |
| User repo/wrong-book generic CRUD | 18 generated GET/POST routes | P3 | ✅ | FB-098 removes all generated stub routes and four confirmed unconsumed wrappers; `el_user_book` model/table/data remain |

### Administration Stub Closure Branch Matrix (2026-07-27)

| Function | Branch | Priority | Coverage |
|----------|--------|----------|----------|
| Profile update | authenticated current user updates valid nickname/email/phone/sex | P0 | ✅ |
| Profile update | forged user ID, malformed JSON, invalid field format, duplicate email/phone, missing user or DB failure | P0 | ✅ |
| Profile password | correct old password and strong new password | P0 | ✅ |
| Profile password | wrong old password, weak/same password, malformed JSON, hash/update failure | P0 | ✅ |
| Profile avatar | valid JPEG/PNG within size limit is stored under configured profile directory | P0 | ✅ |
| Profile avatar | empty/oversized/fake image, path manipulation, file/DB failure | P0 | ✅ |
| Profile avatar | persisted `/profile/...` URL is read through the nginx static location without an API-prefix rewrite | P0 | ✅ | FB-100: frontend store and upload completion keep the backend URL unchanged; regression test rejects `/prod-api/profile/...` |
| User role assignment | authorized caller reads or transactionally replaces valid active roles | P0 | ✅ |
| User role assignment | unauthorized caller, missing user/role, duplicate/invalid role ID, protected admin mutation or DB rollback | P0 | ✅ |
| Role authorization | authorized caller changes status/data scope and assigns or removes users transactionally | P0 | ✅ |
| Role authorization | unauthorized caller, malformed input, protected admin role/user, invalid IDs or DB rollback | P0 | ✅ |
| Registration | enabled registration with valid one-time captcha, unique username and strong password | P0 | ✅ |
| Registration | disabled registration, invalid/replayed captcha, malformed/duplicate username, weak/mismatched password or DB rollback | P0 | ✅ |
| Config/dict cache | cache miss/hit, exact write invalidation and prefix refresh | P1 | ✅ | FB-096/097: one-hour config read-through; config and dict mutations invalidate old/new exact keys; prefix refresh uses batched SCAN |
| Config/dict cache | Redis unavailable or database query failure returns controlled behavior without false success | P1 | ✅ | FB-096/097: config/dict reads fall back from Redis to DB; non-not-found DB failures and refresh failures return controlled errors; empty dict is `[]` |
| Dictionary routing | public type/batch reads bypass JWT while all type/data management routes retain authenticated login context | P0 | ✅ | FB-101: method-specific public reads remain anonymous; broad `/system/dict/` prefix removed so management handlers receive the authenticated login context |
| Optional modules | retired job/code-generation menus and hidden routes are absent | P1 | ✅ | FB-098/099 source and frontend tests; menu IDs disabled by explicit primary-key migration |
| Audit pages | operation/login audit lists remain available while unsupported mutation controls/routes are absent | P1 | ✅ | Read-only list routes/wrappers/pages retained; delete/clean/export controls and routes removed |
| Dead generic routes | user repo/wrong-book generated stubs return 404 and historical data remains untouched | P1 | ✅ | Authenticated HTTP tests return 404; no table/model/data migration is performed |
| Stub inventory | production router has zero `Stub`/`AjaxStub`/`TableStub` registrations and no `_todo` success response | P0 | ✅ | Production route source has no stub caller/helper; `internal/handler/stub.go` deleted |

## Competency Assessment — Phase 1A Security and Dispatch Baseline

> Scope: explicit assessment dispatch, exact anonymous routes, participant/paper token validation, and legacy API isolation.  
> Rule: all planned branches start as uncovered. P0 branches must become ✅ before Phase 1A can be accepted.

### A. Assessment Type Dispatch

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Assessment dispatch | `assessment_type=legacy` and `scoring_mode=legacy` | P0 | ✅ | Covered by `TestValidateAssessmentMode/legacy_pair` |
| Assessment dispatch | `assessment_type=competency` and `scoring_mode=competency_average` | P0 | ✅ | Covered by `TestValidateAssessmentMode/competency_pair` |
| Assessment dispatch | competency type with legacy scoring mode | P0 | ✅ | Covered by `TestValidateAssessmentMode/competency_with_legacy_scoring` |
| Assessment dispatch | legacy type with competency scoring mode | P0 | ✅ | Covered by `TestValidateAssessmentMode/legacy_with_competency_scoring` |
| Assessment dispatch | unknown or empty assessment type on a new competency request | P0 | ✅ | Covered by unknown/empty cases in `TestValidateAssessmentMode` |
| Assessment dispatch | existing database row has no new type before migration backfill | P1 | ✅ | Staging migration backfilled all legacy rows and repeated idempotent execution without changing valid combinations |

### B. Exact Anonymous Route Matching

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Anonymous routing | exact `POST /exam/api/competency/participant/create-paper` | P0 | ✅ | Exact matcher and HTTP middleware call covered |
| Anonymous routing | exact `POST /exam/api/competency/participant/paper-detail` | P0 | ✅ | Covered by `TestCompetencyParticipantRoutesUseExactMethodAndPath` |
| Anonymous routing | exact `POST /exam/api/competency/participant/fill-answer` | P0 | ✅ | Covered by `TestCompetencyParticipantRoutesUseExactMethodAndPath` |
| Anonymous routing | exact `POST /exam/api/competency/participant/submit` | P0 | ✅ | Covered by `TestCompetencyParticipantRoutesUseExactMethodAndPath` |
| Anonymous routing | same participant path with a different HTTP method | P0 | ✅ | GET/PUT/DELETE/PATCH negative cases covered |
| Anonymous routing | participant path with an added suffix or child path | P0 | ✅ | Suffix negatives and HTTP 401 middleware call covered |
| Anonymous routing | competency dimensions/questions/exams management endpoints | P0 | ✅ | Covered by `TestCompetencyManagementRoutesRequireAdminJWT` |
| Anonymous routing | competency results/export/admin report endpoints | P0 | ✅ | Covered by `TestCompetencyManagementRoutesRequireAdminJWT` |
| Anonymous routing | exact internal report-data endpoint | P0 | ✅ | Exact GET plus method/path negatives covered; handler token check remains a later report-handler branch |
| Anonymous routing | existing login/captcha/MBTI/legacy participant paths | P0 | ✅ | Existing anonymous behavior retained by current middleware tests; rerun after change |
| Anonymous routing | existing tester/qu/system management paths | P0 | ✅ | Existing non-anonymous behavior retained by current middleware tests; rerun after change |

### C. Participant and Paper Token Validation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Participant token | valid HS512 signature, unexpired token, expected purpose and matching participant/exam | P0 | ✅ | Round-trip and binding tests pass |
| Participant token | token missing | P0 | ✅ | Covered by missing-token case |
| Participant token | malformed token | P0 | ✅ | Covered by malformed-token case |
| Participant token | wrong signature or non-HS512 algorithm | P0 | ✅ | Wrong-secret case plus existing JWT algorithm tests pass |
| Participant token | expiration claim missing | P0 | ✅ | Covered by missing-expiration case |
| Participant token | expired token | P0 | ✅ | Covered by expired-token case |
| Participant token | wrong `purpose` claim | P0 | ✅ | Covered by wrong-purpose case |
| Participant token | participant ID does not match request/database owner | P0 | ✅ | Covered by `ValidateBinding/wrong_participant` |
| Participant token | exam ID does not match requested exam | P0 | ✅ | Covered by `ValidateBinding/wrong_exam` |
| Paper token | valid token matches participant, exam, and paper | P0 | ✅ | Covered by round-trip and valid binding |
| Paper token | paper ID does not match request | P0 | ✅ | Covered by `ValidateBinding/wrong_paper` |
| Paper token | participant or exam claim differs from paper ownership | P0 | ✅ | Participant and exam mismatch cases covered |
| Token handling | token or signing secret appears in log/error response | P0 | ✅ | Rejection tests assert generic errors contain neither token nor secret; implementation logs neither |

### D. Legacy Paper API Isolation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Legacy `CreatePaper` | exam does not exist | P0 | ✅ | Controlled not-found mapping covered by guard tests; staging legacy smoke verified valid paths remain reachable |
| Legacy `CreatePaper` | exam is competency | P0 | ✅ | Runtime mode test plus source-order test prove rejection before `createPaperTx` |
| Legacy `CreatePaper` | exam is legacy 001/002/003 | P0 | ✅ | Staging created temporary 00101/00201/00301 exams and two-question papers |
| Legacy paper read APIs | paper does not exist | P0 | ✅ | Guard tests cover controlled not-found; staging valid reads confirm no false competency rejection |
| Legacy paper read APIs | paper belongs to competency exam | P0 | ✅ | Guard/order tests cover `paper-detail`, `paperQu-detail`, `qu-detail`, `paper-result`, and `stand-score` |
| Legacy paper read APIs | paper belongs to legacy exam | P0 | ✅ | Staging paper-detail/qu-detail succeeded for 00101/00201/00301 |
| Legacy `FillAnswer` | paper belongs to competency exam | P0 | ✅ | Guard executes before empty-answer success and before write transaction |
| Legacy `FillAnswer` | paper belongs to legacy exam | P0 | ✅ | Staging answered two questions for each 00101/00201/00301 temporary paper |
| Legacy `HandExam` | paper belongs to competency exam | P0 | ✅ | Guard executes before write transaction and is repeated inside transaction before aggregation |
| Legacy `HandExam` | paper belongs to legacy exam | P0 | ✅ | Staging submitted and read paper-result for all three legacy types |
| Legacy tester/candidate standard-score endpoints | paper belongs to competency exam | P0 | ✅ | Both handlers guard before repo lookup and fixed formula query |
| Legacy tester/candidate standard-score endpoints | paper belongs to legacy 001/002 | P0 | ✅ | Formula unit tests plus staging stand-score response for 00101/00201; 00301 generic endpoint remained reachable |
| Legacy guard query | database lookup fails | P0 | ✅ | Any non-not-found DB error maps to controlled assessment-mode error; no legacy fallback |

### E. Phase 1A Acceptance Gate

| Gate | Priority | Coverage | Evidence required |
|------|----------|----------|-------------------|
| All new P0 dispatch branches | P0 | ✅ | `TestValidateAssessmentMode` passes |
| All exact anonymous route branches | P0 | ✅ | Direct matcher and HTTP middleware tests pass |
| All participant/paper token rejection branches | P0 | ✅ | Token tests pass with no secret output |
| All legacy API competency guards | P0 | ✅ | Runtime mode tests and handler source-order tests prove guards precede legacy reads/writes/formulas |
| Legacy 001/002/003 regression | P0 | ✅ | Go full suite plus staging temporary create-paper/detail/fill/submit/result/stand-score chains passed; cleanup=0 |
| Build and test gate | P0 | ✅ | `Go: Build` and `Go: Test All` passed on 2026-07-24 |

### F. Competency Report Audience Version

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Report audience | competency exam selects `frontline_employee` | P0 | ✅ | Staging create/detail round-trip verified |
| Report audience | competency exam selects `leader` | P0 | ✅ | Staging draft edit/detail and DB row verified |
| Report audience | competency exam omits audience or sends an unknown value | P0 | ✅ | Empty and unknown values rejected by `TestValidateCompetencyReportAudience` |
| Report audience | legacy exam has no competency report audience | P0 | ✅ | Staging migration verified 60 legacy exams remain legacy+legacy+NULL audience+published |
| Report audience | report audience is changed after competency publication | P0 | ✅ | Published exam save guard rejects audience changes; published result copies audience snapshot |
| Report audience | report layout and module list for both versions | P0 | ✅ | One `competencyReport.vue` renders both audience values |
| Report audience | overall evaluation content lookup | P0 | ✅ | temp-v1 matches exact audience + evaluation level + content version; formal customer content remains external replacement work |
| Report audience | development advice content lookup | P0 | ✅ | temp-v1 matches exact audience + dimension + level + content version; no cross-audience/version fallback |
| Report audience | result score, dimension order, and charts across versions | P0 | ✅ | SC-012 staging used identical 40/40 answers: overall/5 dimension facts and order matched; two A4 9-page PDFs had 9/9 normalized text equality and matching chart data |
| Report audience | exact text rows used by both real PDF versions | P0 | ✅ | SC-012 matched 2/2 overall and 10/10 dimension texts to exact temp-v1 audience+dimension+level rows |
| Report audience | historical report regeneration | P0 | ✅ | Report data reads `el_competency_result.report_audience` snapshot |
| Result navigation | competency exam “detail” action | P0 | ✅ | FB-075 RED→GREEN routes competency to `CompetencyResults` and preserves legacy route; deployed staging E2E queried the retained exam and clicked the primary detail action, then passed 9 operation classes and hid 9 legacy controls |
| Result navigation | competency legacy `exam/users` stale URL, bookmark or existing tab | P0 | ✅ | FB-076 component fetches exam type before loading participants and replace-routes competency to `CompetencyResults`; staging stale-URL E2E hid legacy controls 9/9 and called legacy generate-report 0 times |
| Result navigation | dashboard recent competency exam action | P0 | ✅ | FB-076 explicitly routes competency to `CompetencyResults`, preserves legacy `exam/users`; dedicated RED→GREEN source regression passed |
| Participant QR navigation | open legacy exam has a physical repo code | P0 | ✅ | Existing QR path includes the required `:repoCode` segment |
| Participant QR navigation | open competency exam has no physical repo association and `repoCode` is empty | P0 | ✅ | FB-102 resolves virtual code 00401 before building the required candidate route; exact URL unit assertion and local 8089 browser route render passed |
| Participant QR navigation | closed competency exam uses tester QR with empty optional repo code | P0 | ✅ | FB-102 uses the same resolver and URL builder; exact tester URL assertion includes `/00401` |

### F2. Competency Product / Scoring / Content / Template Versions

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Legacy save | assessment is legacy | P0 | ✅ | Four competency version fields remain empty and legacy behavior is unchanged |
| Competency draft | all four version fields are omitted | P0 | ✅ | Save resolves current product/scoring/content/template defaults before writing |
| Competency draft | a version contains an invalid identifier | P0 | ✅ | Reject before the exam transaction writes any row |
| Competency draft | a valid future content version is supplied | P0 | ✅ | Preserve the explicit content version without falling back to `temp-v1` |
| Publish | current executable product/scoring/template versions are configured | P0 | ✅ | Publish freezes all four versions with dimension/question snapshots |
| Publish | product, scoring, or template version is unsupported by the running code | P0 | ✅ | Reject before creating snapshots or changing publish status |
| Published edit | any frozen version differs from the stored version | P0 | ✅ | Reject the save while still allowing unrelated editable metadata |
| Submit | a published version set exists | P0 | ✅ | Result freezes product/scoring/content/template versions from the exam, never process constants |
| Report | result references a content and template version | P0 | ✅ | Text lookup and report instance use the exact frozen versions with no cross-version fallback |
| Compatibility migration | existing competency exams/results/reports predate version columns | P0 | ✅ | Staging executed 007 twice; 8 columns exist, all competency gaps are 0, and all 60 legacy exams retain empty versions |
| Exam API / form | competency detail or paging is loaded and switched to legacy | P1 | ✅ | Return/display all versions; switching to legacy clears all four fields |

> 2026-08-09：上述 ✅ 均为本地自动化测试/构建证据；本切片未部署 staging/production。007 迁移只有静态检查证据，必须在可用 MySQL 环境再次执行并核对回填结果。

> [纠正 - 2026-08-10] 上述 007 未部署状态已失效。`20.200.136.133` staging 已完成备份、两次幂等迁移、后端/前端部署、真实 API 拒绝测试及真实管理表单版本摘要验证；production 仍未部署。

### F3. Phase-1 Question Type / Validity / First-Level Result Structures

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Source question | existing legacy or competency row predates question type | P0 | ✅ | Nullable `competency_question_type` preserves legacy rows; existing dimension-linked competency rows backfill to `dimension` |
| Published question | snapshot predates question type | P0 | ✅ | Nullable snapshot field and compatibility backfill preserve old papers; current required dimension/direction fields remain unchanged |
| Overall result compatibility | existing result predates dimension-question counters | P0 | ✅ | Static migration contract backfills dimension counts from existing total/answered counts without recalculating scores |
| First-level group snapshot | product has grouped dimensions | P0 | ✅ | Model/migration contract stores one exam-scoped group plus nullable many-dimension links without hard-coded phase-1 names |
| First-level group result | paper has group aggregation | P0 | ✅ | Model/migration contract keeps score/level nullable and records counts/scoring version; runtime guard confirms no calculation/write was added |
| Validity result | paper has validity questions | P0 | ✅ | Model/migration contract keeps score/status nullable and records counts/scoring version; runtime guard confirms no direction/threshold/write was added |
| Historical result | old product has no group or validity data | P0 | ✅ | Migration creates no synthetic group/validity rows and scoring contract tests keep current `competency-v1` behavior |
| Full-chain delete | new group/validity rows exist | P0 | ✅ | Source-order regression test verifies result children before paper, then dimensions before referenced group snapshots, in one transaction |
| Migration rerun | 008 already applied | P0 | ✅ | Staging MySQL 8.0.46 executed 008 twice; rows, columns, indexes and foreign keys remained identical |

> 2026-08-10：上述 ✅ 已通过聚焦测试 89/89、Go 全量测试、Windows/Linux amd64 构建及 `go vet ./...`；本机无 MySQL/Docker/WSL，因此 008 尚未真实执行一次/两次，迁移重跑分支继续保持 ❌，也未部署 staging/production。

> [纠正 - 2026-08-10] 上述 008 数据库未验证状态已失效。`20.200.136.133` staging 已完成完整备份、首次迁移和第二次幂等迁移；5 个目标列、3 张表、5 个新外键、索引签名、collation、回填和孤儿检查全部通过。仅执行数据库迁移，未部署应用，production 未修改。

### F3A. Versioned Result-Run Storage

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Result run | one paper is scored by v1 and later recomputed by v2 | P0 | ✅ FB-173 local | Parallel run rows are unique by paper + scoring version; original paper-keyed result models/tables are unchanged |
| Result run | run stores frozen versions and participant identity | P0 | ✅ FB-173 local | Run model freezes product/scoring/content/template versions, report audience, participant snapshot, source and status |
| Overall result | a run has one overall result | P0 | ✅ FB-173 local | One-to-one run key stores exact score, level, norm score, comparison code, counts and submission metadata |
| Module result | a run has three module results | P0 | ✅ FB-173 local | Unique run + module code and run + display order store stable identity, exact score, level and norm comparison data |
| Dimension result | a run has ten semantic dimension results | P0 | ✅ FB-173 local | Unique run + dimension ID and run + display order store stable identity, exact score, level and norm score |
| Validity result | a run has one validity result | P0 | ✅ FB-173 local | One-to-one run key stores raw validity sum, internal status and completeness |
| Compatibility | legacy 001～010 result tables contain historical rows | P0 | ✅ FB-173 static | Migration contains no ALTER/UPDATE/INSERT/DELETE for the four existing result tables and inserts no run data |
| Migration rerun | result-run tables or foreign keys already exist | P0 | ✅ STAGING | 011～013 first and repeated execution passed on MySQL 8.0.46; the 2026-09-19 recheck still finds all 8 expected tables and 8 RESTRICT foreign keys |

### F4. Phase-1 A/B Dimension Identity Reset

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Identity source | phase-1 dimension definitions are loaded | P0 | ✅ | Exactly 10 enabled identities exist in order: `A1-01` through `A1-05`, then `B1-01` through `B1-05`; names and customer core meanings match the 2026-08-07 material |
| Identity source | unresolved phase-2/3 matrix is inspected | P0 | ✅ | Migration seeds no dimension outside the confirmed phase-1 ten and makes no 34/40-dimension claim |
| Fresh installation | migration 002 initializes dimension master data | P0 | ❌ | New databases receive only the confirmed A/B phase-1 identities and no `D01-D48` row |
| Existing environment reset | old competency exams, papers, results, snapshots, reports or source questions exist | P0 | ✅ | Staging invoked the existing transactional full-chain delete for all 9 competency exams, reduced all runtime/report dependencies to zero, then migration 009 replaced 384 source questions, 392 report-text rows and 48 retired dimensions with the ten A/B identities |
| Legacy isolation | traditional 001/002/003 data exists during reset | P0 | ✅ | Ten pre-reset legacy signatures covering exams, papers, paper questions, candidates, testers, user exams, questions, answers, repository links and repositories remained byte-identical after delete, 009 and deployment |
| Report cleanup | competency report instances reference generated PDF files | P0 | ✅ | Seven referenced PDFs were backed up, removed by the application full-chain delete and individually verified absent; the competency report directory contained zero PDFs afterward |
| Temporary content | old `temp-v1` report text references the retired D identity set | P0 | ✅ | Migration 009 cleared all 392 old report-text rows and created no replacement scoring/report content |
| Reset preflight | environment authorization/write quiescence is absent, a competency exam/runtime/report dependency remains, or the database shape is unexpected | P0 | ❌ | Migration 009 requires explicit staging-only and stopped-write authorization in the same session, takes a migration lock, then aborts before deleting source/master data on any failed precondition; it never bypasses full-chain/PDF cleanup |
| Reset rerun | identity reset has already completed | P0 | ✅ | Staging first execution recorded `apply_reset=1`; the second recorded `apply_reset=0`, while the marker and full ten-row dimension signature remained unchanged |
| Candidate artifact | customer workbook is converted after identity confirmation | P0 | ✅ | Candidate JSON uses A/B dimension IDs/codes and A/B-prefixed dimension question codes, with no D identity mapping or `MAP-001` blocker |
| Dimension maintenance | administrator edits a confirmed A/B dimension | P0 | ✅ | API/UI use the two confirmed layers (`通用能力`/`心理素养`), category `基层员工`, and order range 1-10; stable ID/code remain immutable |
| Question import contract | administrator downloads or validates the current dimension-question template | P1 | ✅ | Guidance and examples use current A/B identities; validation matches a positive order to an existing dimension rather than assuming D01-D48 |

> 2026-08-10：用户确认可清除旧胜任力历史数据，并选择“本地改造 + `20.200.136.133` staging 全量重置”；传统 001/002/003 必须保留，production 不修改。本切片不导入 90 题，不实现效度方向/阈值、一级聚合或五档评分。

> 2026-08-10 本地验证：先取得Go 71通过/8失败、前端132通过/1失败和候选身份失败的RED证据；实现后聚焦Go 79/79、Go全量、前端23文件133项、Windows/Linux构建、`go vet ./...`、前端production build、候选身份与确定性检查均通过。002/009仅通过静态契约；真实MySQL首次执行、阻断分支、第二次no-op、传统摘要和PDF残留仍未验证，故对应分支保持❌。当前公网health为ok，但当前公网IP `20.239.176.250` 到staging TCP/22超时，未登录、未备份、未删除、未执行009、未部署；production未修改。

> [纠正 - 2026-08-10] 上述 staging 阻塞及009未验证状态已失效。已完成受限完整备份、Nginx停写、9个胜任力测评整链删除、7个PDF逐路径核验、009首次执行和第二次no-op、后端/前端部署、真实Nginx API及Chromium页面验收；最终为marker=1、A/B维度10、D/源题/旧文案/胜任力测评/运行依赖/PDF均0，传统十组签名不变。002全新Schema实跑和009缺授权/残留依赖的真实负向阻断尚未执行，故对应两项继续保持❌；production未修改。

### F4A. Phase-1 v2 Semantic Identity Mapping

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| v2 identity | ten customer dimensions are loaded | P0 | ✅ FB-174 | Semantic IDs/keys are independent from A/B/C display codes and fixed in the approved report order |
| v1 answer mapping | each v1 dimension identity is present | P0 | ✅ FB-174 | All ten old snapshot dimension IDs map one-to-one to the matching semantic v2 identity |
| v1 answer mapping | blank or unknown source identity is supplied | P0 | ✅ FB-174 | Mapping fails closed; it never guesses by display order or silently drops an answer |
| v2 grouping | mapped dimension belongs to a report module | P0 | ✅ FB-174 | Fixed membership is task 5 / interpersonal 2 / self 3 |
| v2 display | report code or display order changes | P0 | ✅ FB-174 | Display code/order are version-scoped metadata, not the stable dimension identity |
| v2 versions | v2 constants exist before scoring/report implementation | P0 | ✅ FB-174 | Exact four v2 identifiers are defined but executable validation continues to reject them until later slices are GREEN |
| Compatibility | v1 master dimensions, source questions and snapshots exist | P0 | ✅ STATIC / DB NOT APPLIED | 012 creates only version catalog/mapping tables; it does not alter/update/delete v1 master, questions or snapshots |
| Migration rerun | v2 identities/mappings already exist | P1 | ✅ STAGING | First and repeated MySQL execution passed; the 2026-09-19 recheck finds exactly 10 unique catalog identities/orders and 10 one-to-one v1→v2 mappings |

### F4B. Phase-1 v2 Exact Percentage Scoring

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Dimension score | all eight final item scores are present | P0 | ✅ FB-175 | Exact score is `25 × (sum / 8) - 25`; no intermediate rounding |
| Overall score | all ten dimensions are complete | P0 | ✅ FB-175 | Exact arithmetic mean of ten percentage scores; sample remains 545/8 = 68.125 |
| Display rounding | exact score has a third decimal digit | P0 | ✅ FB-175 | Scoring result retains `big.Rat`; this slice performs no display or persistence rounding |
| Level | score is below/at 10, 30, 70 or 90 | P0 | ✅ FB-175 | Continuous bands are `<10`, `[10,30)`, `[30,70)`, `[70,90)`, `>=90` |
| Historical identity | v1 answer rows arrive in arbitrary row order | P0 | ✅ FB-175 | Rows map by v1 dimension ID and output the approved semantic v2 order, never by name |
| Incomplete answer | any dimension has fewer than eight answers | P0 | ✅ FB-175 | Incomplete dimension and overall expose no formal score/level; counts remain available |
| Invalid input | count/type/identity/order/final score is invalid | P0 | ✅ FB-175 | Pure scorer fails closed and emits no partial formal result |
| Compatibility | v1 scorer remains callable | P0 | ✅ FB-175 | New implementation is isolated; existing v1 scoring functions and constants are unchanged |

### F4C. Phase-1 v2 Modules and Norm Comparisons

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Module aggregation | all five task dimensions are complete | P0 | ✅ FB-176 | Exact mean is 67.5; norm is 58; comparison is `above_norm` |
| Module aggregation | both interpersonal dimensions are complete | P0 | ✅ FB-176 | Exact mean is 59.375; norm is 53.75; corrected comparison is `above_norm` |
| Module aggregation | all three self dimensions are complete | P0 | ✅ FB-176 | Exact mean is 75; norm is 60; comparison is `standout` |
| Module order | scores are tied or input dimensions are reordered | P0 | ✅ FB-176 | Output identity/order remains task → interpersonal → self; report sorting belongs to a later selector slice |
| Task comparison | score is below/at 56, 60 or 75 | P0 | ✅ FB-176 | Bands are `<56`, `[56,60)`, `[60,75)`, `>=75` |
| Interpersonal comparison | score is below/at 50, 56 or 70 | P0 | ✅ FB-176 | Bands are `<50`, `[50,56)`, `[56,70)`, `>=70` |
| Self comparison | score is below/at 58, 63 or 75 | P0 | ✅ FB-176 | Bands are `<58`, `[58,63)`, `[63,75)`, `>=75` |
| Overall comparison | score is below/at 55, 63 or 70 | P0 | ✅ FB-176 | Bands are `<55`, `[55,63)`, `[63,70)`, `>=70`; norm is 57.75 |
| Incomplete module | any child dimension is incomplete | P0 | ✅ FB-176 | That module has no score, level or comparison; independent complete modules remain formal |
| Invalid aggregate input | identity/order/module/score/level is inconsistent | P0 | ✅ FB-176 | Aggregator fails closed instead of silently regrouping malformed results |

### F4D. Phase-1 v2 Report Selectors

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Module overview | three complete modules have different scores | P0 | ✅ FB-177 | Sort descending by exact score |
| Module overview | two or three modules tie | P0 | ✅ FB-177 | Tie order is task → interpersonal → self |
| Strengths | any combination of dimension levels | P0 | ✅ FB-198 STAGING | Ignore level categories and always select the three highest exact scores; real varied-score PDF selected self-discipline, achievement orientation and plan execution |
| Development | any combination of dimension levels | P0 | ✅ FB-198 STAGING | Ignore level categories and always select the two lowest exact scores; a real tie at 53.125 used fixed order and selected logical reasoning after dedication |
| Selected-item text | selected dimension is in any of the five levels | P0 | ✅ FB-198 STAGING | Real PDF contains all five corresponding complete approved performance texts; no cross-category fallback or new content package |
| Overall advice | overall is excellent/good/qualified | P0 | ✅ FB-177 | Select the lowest two dimensions independently of report development slots |
| Overall advice | overall is weak/insufficient | P0 | ✅ FB-177 | Select the lowest three dimensions independently of report development slots |
| Invalid selector input | a module/dimension is incomplete or metadata/level is inconsistent | P0 | ✅ FB-177 | Selector fails closed and does not emit partial report choices; module rows are recomputed to reject cross-source mixtures |

### F4E. Phase-1 v2 Report DTO and Versioned Rule Text

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| DTO versions | v2 result is projected | P0 | ✅ FB-178 | Freeze all four exact v2 identifiers, audience and result-run identity |
| DTO scores | exact rational has more than two decimals | P0 | ✅ FB-178 | DTO displays ROUND_HALF_UP two decimals while scorer remains exact; 78.125→78.13 and 68.125→68.13 |
| Rule lookup | exact active v2 row exists | P0 | ✅ FB-178 | Match content version + audience + content type + semantic identity + level/comparison code |
| Rule lookup | only v1/other-audience/retired/temporary row exists | P0 | ✅ FB-178 | No cross-version, audience, status or temporary fallback |
| Rule lookup | exact key is duplicated or content/disclaimer is blank/inconsistent | P0 | ✅ FB-178 | Fail closed with the exact missing/duplicate key |
| Module text | each module comparison is selected | P0 | ✅ FB-178 | Match `module_comparison + module code + comparison code`, then retain selector order |
| Dimension text | all ten dimension performance rows exist | P0 | ✅ FB-178 | Match `dimension + stable dimension ID + level` |
| Strength/development text | selected category row exists | P0 | ✅ FB-178 | Match separate `strength`/`development` content; never reuse full performance text |
| Empty category | selector returns zero strengths/developments | P0 | ✅ FB-178 | DTO returns non-nil empty slices and requires no unused category rows |
| Validity | status is good/questionable | P0 | ✅ FB-178 | Internal status remains frozen and display text comes from exact versioned validity row |

### F4F. Phase-1 v2 Value-Only Word Rendering

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Body controls | all 58 field keys are supplied | P0 | ✅ FB-179 | Replace every occurrence in `word/document.xml`, preserving XML structure and styles |
| Header controls | participant/validity controls are in header parts | P0 | ✅ FB-179 | Replace matching controls in every `word/header*.xml` part |
| Repeated control | a repeatable key occurs in body and/or header | P0 | ✅ FB-179 | Every occurrence receives the same frozen value |
| Optional slot | a predefined strength/development/profile value is absent | P0 | ✅ FB-179 | Empty only that control's value; do not invent layout or text |
| Overall chart | exact score is supplied | P0 | ✅ FB-179 | Update the existing two-point literal series only |
| Comparison chart | ten scores and ten norms are supplied | P0 | ✅ FB-179 | Update both existing literal value series only; keep chart type/labels/colors/axes |
| Dimension charts | ten exact score/remainder pairs are supplied | P0 | ✅ FB-179 | Update each business-keyed literal series without relying on physical chart number |
| Template style | customer changes formatting or chart presentation | P0 | ✅ FB-179 | All non-value XML and all unrelated ZIP parts remain byte-identical |
| Contract failure | field/chart key is missing/unknown, duplicate chart key exists, formula/reference/external link returns | P0 | ✅ FB-179 | Renderer fails closed and emits no DOCX |

### F4G. Versioned Report Instance Binding

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Legacy report | existing report predates result runs | P0 | ✅ STATIC / DB NOT APPLIED | `result_run_id` remains nullable; migration performs no backfill or overwrite |
| v2 report | report is generated from a result run | P0 | ✅ STATIC / NOT WIRED | Report binds one exact run; composite FK guarantees report paper equals run paper |
| Version identity | same run is rendered with the same content/template/audience | P0 | ✅ STATIC / DB NOT APPLIED | Unique run+content+template+audience prevents duplicate instances |
| Parallel versions | one paper has v1 and v2 reports | P0 | ✅ STATIC / DB NOT APPLIED | Existing paper+content+template identity is retained; no old report/PDF is replaced |
| Current display | a report is selected for paper+audience | P0 | ✅ STATIC / NOT WIRED | Separate pointer table stores one current report per paper+audience |
| Pointer integrity | pointer report belongs to another paper/audience | P0 | ✅ STATIC / DB NOT APPLIED | Composite FK rejects mismatched report/paper/audience |
| Migration rerun | column/index/table/FKs already exist | P1 | ⚠️ STATIC GREEN / REAL DB PENDING | Information-schema guards make every DDL step idempotent and skip alignment after constraints exist |
| Full-chain delete | exam owns v1/v2 reports and result runs | P0 | ✅ FB-180 | Delete audit → current pointer → reports → run children → run → legacy results → paper |
| Pre-migration delete | 011/013 optional tables do not exist yet or are partially applied | P0 | ✅ FB-182 | Guard the current pointer and each run child independently, require the run parent for child subqueries, and continue the complete v1 transactional deletion chain |

### F4H. Phase-1 v2 Report Runtime Wiring

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Result source | a completed exact v2 result run exists for the paper | P0 | ✅ LOCAL / DB NOT APPLIED | Read run/overall/module/dimension/validity rows, reconstruct exact scores from score sums and reject persisted drift |
| Result source | no v2 result run exists | P0 | ✅ FB-181 | Preserve the existing v1 report generation path without changing its result or instance identity; schema probes skip all 011/013 access before migration |
| Result source | v2 run exists but is incomplete, malformed or version-mismatched | P0 | ✅ FB-181 | Fail closed on run identity, versions, completion/submission metadata, exact score/module/norm/validity drift; never fall back to v1 |
| Content approval | exact v2 content package and rule rows are approved | P0 | ✅ LOCAL / DB NOT APPLIED | Require exact versions/audience, dual approval, SHA values, active environment and matching disclaimer |
| DTO adapter | complete v2 formal report data is supplied | P0 | ✅ FB-181 | Produce exactly 58 fields and 12 chart payloads with approved labels, overall assessment/advice, optional slots and validated exact chart values |
| Renderer dispatch | report kind is phase-1 v2 | P0 | ✅ FB-181 | Read the independent v2 template and invoke value-only rendering before the existing converter; no v1 chart calibration or Chromium fallback |
| Report write | v2 PDF completes | P0 | ✅ STATIC / DB NOT APPLIED | Bind the instance to the exact result run and atomically set the paper+audience current pointer without changing legacy participant `pdf_path` |
| Report reuse | completed v2 instance and file already exist | P1 | ✅ STATIC / DB NOT APPLIED | Reuse the same run-bound instance, refresh the current pointer and write one reuse audit; failed force regeneration preserves the prior completed PDF |
| Current download | paper+audience current pointer exists | P0 | ✅ STATIC / DB NOT APPLIED | Download the exact completed report after validating bound run, versions, audience, approval and environment |
| Legacy download | no current pointer or binding schema exists | P0 | ✅ FB-181 | Retain the v1 result/version/report lookup and omit the unapplied nullable column from legacy SQL |
| Runtime identity | result run participant differs from paper owner | P0 | ✅ FB-183 | Read paper exam/owner together and reject before DTO construction unless both equal the frozen run identity |
| Download concurrency | force regeneration replaces a report while the same paper is downloading | P0 | ✅ FB-183 | Serialize single download with the existing per-paper generation lock so old-path deletion cannot race an open attempt |

### F4I. Phase-1 v2 Result-Run Creation and Historical Recompute

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| New submission | all five 011 result-run tables exist and a complete v1 phase-1 paper is submitted | P0 | ✅ FB-184 | Persist run, overall, three modules, ten dimensions and validity in the same submission transaction |
| Pre-migration submission | none of the five 011 tables exists | P0 | ✅ FB-184 | Preserve the existing v1 submission without issuing result-row writes |
| Partial migration | tables, critical columns, unique indexes or RESTRICT constraints are incomplete | P0 | ✅ FB-184 | Fail closed inside the submission transaction; never commit a partial v1/v2 submission state |
| Historical recompute | completed exact v1 phase-1 paper has all 90 frozen answer rows | P0 | ✅ FB-184 | Reconstruct 80 final dimension inputs and ten raw validity inputs from immutable paper snapshots; verify raw/final direction consistency |
| Exact persistence | v2 scores contain eighth fractions | P0 | ✅ FB-184 | Store DECIMAL(18,6) from exact rationals, preserving 68.125/59.375/78.125 facts without float64 |
| Idempotency | paper already has a complete valid v2 run | P0 | ✅ FB-184 | Revalidate frozen identity and every derived score, then return the existing run without UPDATE/DELETE or duplicate child inserts |
| Existing corruption | run exists but a child row or exact value is missing/mismatched | P0 | ✅ FB-184 | Reject and preserve all existing rows; never repair by overwrite |
| Concurrency | submit/recompute requests target the same paper | P0 | ✅ FB-184 | Lock the paper row before checking paper+scoring-version uniqueness |
| Authorization | ordinary result viewer requests historical recompute | P0 | ✅ FB-184 | Return 403 before calling the service; only administrator/global permission may write |
| Scope | caller requests bulk recompute or v2 report generation | P1 | ⚪ | Bulk operation, content approval and report generation remain separate slices |

### F4J. Phase-1 v2 Result-Run Review Gaps

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Audience identity | completed v1 result has a non-frontline report audience | P0 | ✅ FB-185A | Reject before creating an immutable v2 run that the v2 report reader cannot consume |
| Snapshot ownership | paper question points to a question/dimension snapshot from another exam | P0 | ✅ FB-185A | Require both frozen snapshot rows to belong to the locked paper exam |
| Transaction rollback | run/overall succeeds but a later child insert fails | P0 | ✅ FB-185C | Executed real writer with injected module failure; observed rollback/no commit; valid reuse executed with zero writes |
| Concurrent idempotency | two independent connections recompute the same paper | P0 | ✅ FB-185I STAGING | Two real MySQL connections produced one create, one reuse and 1/1/3/10/1 child counts; the current 15-paper replay remains zero-create/all-reuse |
| Existing identity drift | module_id or run source is invalid while visible scores remain valid | P0 | ✅ FB-185B | Reject creation/reuse/formal read without mutating rows; allow historical recompute to reuse a valid submission-created run |
| Schema signature | expected object names exist but columns/index order/uniqueness/FK targets or actions differ | P0 | ✅ FB-185D | Cached preflight checks complete migration-shaped definitions and known optional 013 index before writes |
| Submission capacity | 300 papers submit together after 011 | P0 | ✅ STATIC FB-185D / LOAD PENDING | Schema metadata checks moved before the transaction and run once per service process; 300-user load evidence remains a separate environment gate |
| Duration audit | paper user_time changes after v2 run creation | P1 | ✅ FB-185F STAGING | 011 with frozen overall.user_time is applied and the v2 report runtime reads the immutable run value |
| Error disclosure | a result-run database operation fails | P1 | ✅ FB-185E | Client receives a stable domain error without SQL/table/index details; internal cause is logged |
| Staging verifier | approved v2 report generation is already enabled while a default recompute-only check runs | P1 | ✅ FB-189 | Default verifier checks recompute idempotency only; report regeneration/download remains explicit behind `--generate-report` |
| Report profile fields | exam requiredFields contains only name/gender/telephone | P1 | ✅ FB-190 STAGING | Formal data reads exam.required_fields and the v2 renderer removes only unconfigured full-width profile rows; real PDF shows exactly name/gender/telephone |
| Report duration | frozen user_time is 1 minute | P0 | ✅ FB-190 STAGING | Template cover table reserves independent date/duration widths; real PDF text and pixels contain the complete `时长：1分钟` |
| Historical empty selector fallback | pre-FB-198 rules produced no good/excellent strength | — | ⚪ SUPERSEDED BY FB-198 | Kept as historical evidence only; current score-only ranking always fills three strength and two development slots |
| Overview grade scale | overall score is rendered by LibreOffice | P1 | ✅ FB-190 STAGING | Real PDF shows a readable five-band horizontal scale; DrawingML and VML fallback geometry both pass target LibreOffice rendering |
| Detail pagination | approved text lengths differ from template samples | P1 | ✅ FB-190 STAGING | Target LibreOffice requires split detail tables plus real separators; final report has exactly 2/2/2/2/2 dimensions and no orphan module heading |
| Approved validity/disclaimer | v2 DTO contains DisplayText and Disclaimer | P1 | ✅ FB-190 STAGING | Real PDF includes the exact approved validity text and disclaimer through the 60-field contract without fallback |
| Customer template re-save | Word renumbers the grade-scale image relationship while retaining the same media target | P1 | ✅ FB-191 STAGING | Drawing is resolved by OPC target `word/media/image17.png`; repaired customer/local/staging templates are byte-identical, and target LibreOffice generated the final 10-page report |
| Comparison chart rendering | Word 2010 `wpg` group contains the grade image and `chart.dimension.comparison` | P0 | ✅ FB-192 STAGING | Split into two ordinary inline drawings; real target PDF shows the vertical grade scale, ten score bars/labels, ten categories and normal line on page 4 |
| Overview selected-item style | strengths/developments are non-empty | P0 | ✅ FB-193 STAGING | Real 3-strength/2-development report includes actual selected names; only `维度名：` is bold and approved level-specific text remains regular weight |
| Overview visual hierarchy | summary card, overall doughnut and comparison scale render together | P1 | ✅ FB-194 STAGING | One green accent system, compact `总体得分` centre and 180×650 scale pass real target-LibreOffice 10-page rendering without changing values or chart semantics |
| Template management version | admin uploads a valid v2 semantic-tag DOCX | P0 | ✅ FB-195 STAGING | Dedicated v2 metadata/download/upload routes target `v2TemplatePath`, enforce 60 fields/12 charts, and preserve legacy v1 file/validator unchanged |
| Comparison bar color | each dimension score crosses a five-band boundary | P1 | ✅ FB-196 STAGING | Exact five-band point fills make equal bands equal colors; real PDF confirms qualified/good/excellent colors and preserves the orange normal line |
| Phase-1 v2 export | exam has completed v2 runs, missing v2 runs, or legacy-only rows | P0 | ✅ FB-197 STAGING | Real 3-run export is 3×75/270×20/90×14 and matches 42 DB facts; real no-v2 exam is headers-only on all sheets with no v1 mixing/fallback |

### G. Competency Exam Creation Configuration

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Exam creation | assessment type is legacy | P0 | ✅ | Staging browser loaded 60 existing exams including 001/002/003 variants after migration/deploy |
| Exam creation | assessment type is competency | P0 | ✅ | Form conditionally shows report version/dimensions and hides repo controls; frontend build passes |
| Exam creation | competency selects frontline employee and one or more enabled dimensions | P0 | ✅ | Staging API created draft with 2 dimensions; detail and SQL verified |
| Exam creation | competency selects leader and one or more enabled dimensions | P0 | ✅ | Staging API edited draft to leader with 1 dimension; detail and SQL verified |
| Exam creation | competency report version is empty or unknown | P0 | ✅ | Frontend validation and backend whitelist tests pass |
| Exam creation | no dimension selected | P0 | ✅ | Frontend rules plus backend empty/nil tests pass |
| Exam creation | duplicate dimension IDs | P0 | ✅ | Covered by `TestValidateCompetencyDimensionIDs/duplicate_id` |
| Exam creation | selected dimension does not exist or is disabled | P0 | ✅ | Staging temporarily disabled D48; Save rejected and wrote zero exam rows, then D48 was restored |
| Exam creation | selected dimension has zero enabled questions | P0 | ✅ | Pure guard tests identify the first zero-count dimension by code/name before exam write |
| Exam creation | selected dimensions all have enabled questions | P0 | ✅ | Guard accepts positive counts and Save stores each count in draft association |
| Exam creation | enabled-question count query fails | P0 | ✅ | Closed-database injection test proves grouped query error propagates to the Save transaction |
| Dimension list | enabled question counts vary by dimension | P0 | ✅ | One grouped query returns each count; UI sums selected questions and disables zero-count/status-disabled dimensions |
| Exam creation | dimensions list is empty | P1 | ✅ | Selector unit test verifies explicit migration guidance empty state |
| Exam creation | edit competency draft | P0 | ✅ | Staging frontline→leader and 2→1 dimension round-trip verified |
| Exam creation | edit published competency exam | P0 | ✅ | Staging published D01 exam then rejected leader→frontline audience change; temporary exam was deleted |
| Exam creation | switch form from competency back to legacy before save | P1 | ✅ | `handleAssessmentTypeChange` clears audience/dimensions, restores legacy scoring/publish state and repo controls |
| Exam creation | save request fails | P1 | ✅ | Save uses loading with `finally`; form model is retained for retry |

### G2. Competency Exam Deletion

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Delete exam | legacy exam has participant/paper relations | P0 | ✅ | Preserve existing rejection behavior |
| Delete exam | published competency exam with full runtime chain | P0 | ✅ | Dedicated transaction deletes results, answers, paper, participants, snapshots and exam in dependency order |
| Delete exam | competency chain deletion fails midway | P0 | ✅ | Every delete error is returned to the outer transaction, causing rollback |
| Delete exam | deletion succeeds | P0 | ✅ | Staging deleted an 8-question published chain; exam/snapshots/paper/results/candidate remaining count was 0 |
| Direct delete | competency paper/candidate/tester deleted outside exam chain | P0 | ✅ | Generic physical/logical delete endpoints reject and instruct deletion through owning exam |

### H. Competency Question Metadata and Pure Scoring

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question metadata | competency question has code, one dimension, item number, observation point, direction, and status | P0 | ✅ | Model and migration static tests expose all required fields |
| Question metadata | legacy question has NULL competency metadata | P0 | ✅ | Staging SQL verified zero existing questions with competency metadata after migration |
| Question metadata | duplicate global question code | P0 | ✅ | Staging MySQL rejected duplicate with 1062 on `uk_qu_question_code`; transaction left zero rows |
| Question metadata | duplicate dimension item number | P0 | ✅ | Staging MySQL rejected duplicate with 1062 on `uk_qu_dimension_item`; transaction left zero rows |
| Question metadata | dimension reference collation differs by environment | P0 | ✅ | Staging verified both question and dimension IDs use `utf8mb4_general_ci` |
| Question scoring | forward raw value 1 through 5 | P0 | ✅ | Final scores are 1,2,3,4,5 |
| Question scoring | reverse raw value 1 through 5 | P0 | ✅ | Final scores are 5,4,3,2,1 |
| Question scoring | raw value outside 1 through 5 | P0 | ✅ | Reject |
| Question scoring | direction empty or unknown | P0 | ✅ | Reject |
| Dimension scoring | complete dimension | P0 | ✅ | Average all question final scores |
| Dimension scoring | timeout with unanswered questions | P0 | ✅ | Average answered questions only; mark incomplete |
| Dimension scoring | zero answered questions | P0 | ✅ | Score is NULL and dimension is excluded from overall score/evaluation |
| Overall scoring | multiple valid dimensions | P0 | ✅ | Sum exact dimension averages without early rounding |
| Overall scoring | no valid dimensions | P0 | ✅ | Overall score is 0; evaluation average/level are NULL |
| Overall scoring | same answers in different display order | P0 | ✅ | Identical dimension and overall results |
| Score level | boundary values 1.00, 2.00, 3.00, 4.00, 5.00 | P0 | ✅ | Map to low/average/good/high exactly |

### I. Competency Question Import

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import template | administrator downloads template | P0 | ✅ | HTTP test parses xlsx and verifies nine headers, four rows, and zero merged cells |
| Import template | template provides understandable sample data | P1 | ✅ | Contains two valid D01 examples covering forward and reverse scoring |
| File boundary | file missing, not xlsx, empty, or larger than 10 MiB | P0 | ✅ | HTTP tests cover missing, wrong extension, zero-byte and 10MiB+1 payloads |
| Import preview | valid rows | P0 | ✅ | Pure validation returns normalized row; source guard proves preview has no write calls |
| Import preview | header differs from the nine-column contract | P0 | ✅ | Explicit header error test passes |
| Row validation | dimension order outside 1-48 or dimension missing/disabled | P0 | ✅ | Invalid order, missing dimension, and disabled dimension tests pass |
| Row validation | dimension name differs from master data | P0 | ✅ | Exact-name mismatch test passes |
| Row validation | question code empty or duplicated in file/database | P0 | ✅ | Empty, file duplicate, and injected database duplicate tests pass |
| Row validation | dimension item number invalid or duplicated in file/database | P0 | ✅ | Invalid, file duplicate, and injected database duplicate tests pass |
| Row validation | question content or observation point empty | P0 | ✅ | Both required-field tests pass |
| Row validation | direction is not 正向/反向 | P0 | ✅ | Unknown direction test passes; values normalize to forward/reverse |
| Row validation | status is not 启用/停用 | P0 | ✅ | Unknown status test passes; values normalize to 0/1 |
| Row validation | a dimension has any question count, including outside 7-8 | P0 | ✅ | Single-row dimension validates without count warning |
| Formal import | uploaded SHA-256 differs from preview | P0 | ✅ | Real multipart HTTP test rejects changed file before DB access using constant-time comparison |
| Formal import | any validation error exists | P0 | ✅ | Validation gate precedes transaction; preview reports all row errors |
| Formal import | all rows valid | P0 | ✅ | Staging previewed and imported 384/384 rows, rerun detected complete existing set and data hash matched source |
| Formal import | database insert fails | P0 | ✅ | Staging temporary BEFORE INSERT trigger forced MySQL failure; API reported rollback, question residue=0, trigger removed |
| Question export | one or more competency questions exist | P0 | ✅ | Exports all source questions in stable dimension/item order using the same nine-column contract as import |
| Question export | no competency questions exist | P1 | ✅ | Returns a valid header-only workbook |
| Question export | database query or workbook generation fails | P1 | ✅ | Returns a controlled error before writing a partial xlsx response |

#### I1. Phase-1 90-Question Mixed Import

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import contract | workbook uses the new ten-column contract with 题目类型 | P0 | ✅ | Template/service/export tests share ten headers; staging export returned ten headers and 90 data rows; FB-104 keeps the dialog wording on the ten-column contract |
| Row validation | 题目类型 is 维度题 | P0 | ✅ | Local service tests normalize to `dimension`; staging persisted 80 rows with dimension+type+item isolation |
| Row validation | 题目类型 is 效度题 | P0 | ✅ | Local service tests require forward validity rows; staging persisted 10 rows, one associated with each A/B dimension and excluded them from enabled dimension-question counts via FB-103 |
| Row validation | 题目类型 is empty or unknown | P0 | ✅ | Local service tests reject empty/unknown values before database writes |
| Formal import | one file contains 80 dimension and 10 validity rows | P0 | ✅ | Staging preview returned 90/0, formal import wrote 90 in one transaction, repeated preview returned 90 errors and repeated import was rejected with row count unchanged |
| Candidate conversion | all confirmed P0/P1 decisions resolve the candidate blockers | P0 | ✅ | Candidate v3 is deterministic and import-ready with zero blockers/warnings, ten forward validity rows, B1-04 all-forward and four confirmed version names |
| Candidate import workbook | generated workbook matches the resolved candidate byte-for-byte | P0 | ✅ | `--check` and identity test verify byte reproduction, 90 unique rows, 62/18 dimension directions and 10 forward validity rows; staging imported SHA-256 matched |
| Question export/UI | mixed source questions are read back after import | P1 | ✅ | Staging export content matched the imported workbook by question code; API returned 80/10; real Chromium showed both type tags, total 90 and the ten-column dialog with zero console/request errors |

#### I2. Phase-1 Fixed Product Configuration

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Draft defaults | new competency draft omits fixed audience, dimensions and four versions | P0 | ✅ | Pure profile test returns frontline employee, canonical ten A/B IDs, 20-minute default and four confirmed phase-1 versions; Save applies omitted fixed fields |
| Draft validation | client submits leader audience, another dimension set/order or another version | P0 | ✅ | Table-driven service tests reject each mutation; handler guard runs before database transaction |
| Draft inventory | all ten dimensions are enabled with exactly 8 enabled dimension rows and 1 enabled validity row each | P0 | ✅ | Grouped-query sqlmock test returns 8/1; validator accepts all ten exact inventories and draft associations retain dimension-only count 8 |
| Draft inventory | a dimension is absent/disabled or has a count/type mismatch | P0 | ✅ | Existing enabled-master length guard rejects missing/disabled dimensions; inventory matrix rejects 7/1, 8/0 and unknown-type rows before exam writes |
| Frontend profile | administrator selects competency on a new or draft form | P0 | ✅ | Vue test verifies selectors are absent, fixed profile/version text is present, fixed values are applied and duration defaults to 20 minutes |
| Frontend legacy switch | administrator switches the unsaved form back to legacy | P1 | ✅ | Existing version-form test verifies all four versions are cleared; implementation also clears audience/dimensions and restores repository data |
| Publish gate | phase-1 scoring/group/validity runtime is complete | P0 | ✅ | Focused RED proved the old service/UI gate; unified runtime tests now require backend readiness and an enabled Vue publish action |

#### I3. Phase-1 Five-Level Scoring Engine

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question score | dimension question is forward or reverse with raw value 1–5 | P0 | ✅ | Existing exhaustive 1–5 forward/reverse table remains GREEN; invalid values and directions are rejected |
| Dimension score | each canonical A/B dimension has exactly 8 answered dimension questions | P0 | ✅ | Phase-1 pure engine requires 80 rows and computes each exact `scoreSum/8` rational without early rounding |
| Dimension score | any canonical dimension is missing, has not exactly 8 rows, contains a non-dimension type or has unanswered rows | P0 | ✅ | Missing/mixed/unknown identity inputs are rejected; trusted incomplete rows return counts but nil formal dimension/overall scores and no levels |
| Dimension level | exact score is at 1.7/2.7/3.5/4.3 boundaries or immediately above | P0 | ✅ | Boundary table covers exact and +0.01 values with upper-inclusive L1–L5 rational comparisons |
| Overall score | all ten canonical dimensions are complete | P0 | ✅ | Ten exact dimension averages sum to an exact rational in 10–50; test fixture produces 30 without evaluation-average reuse |
| Overall level | exact total crosses 25/32.5/40/45 boundaries | P0 | ✅ | Boundary table covers below/exact values for not-qualified/weak/qualified/good/excellent without percentage conversion |
| Input identity | dimensions are reordered but identities/order metadata are valid | P1 | ✅ | Reverse input order still emits canonical A/B order and deterministic total/level |
| Runtime integration | submit/publish flow uses phase-1 scoring result | P0 | ✅ | Staging真实90题链持久化十维均为3/L3，总体30/weak，重复提交不重复写入 |

#### I4. Phase-1 First-Level Group Aggregation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Group identity | canonical ten A/B dimension results are provided in any order | P0 | ✅ | Reverse-order test emits exactly `general_ability/通用能力` then `psychological_quality/心理素养`, each with its fixed five child IDs |
| Group score | all five child dimensions are complete | P0 | ✅ | Both groups average five exact dimension rationals to 3 without early rounding and classify as L3 |
| Group score | one or more child dimensions are incomplete | P0 | ✅ | General group retains 5/4 dimensions and 40/39 question counts with nil score/level; unaffected psychological group remains complete with exact score 4 |
| Group level | exact group average is at 1.7/2.7/3.5/4.3 boundaries | P0 | ✅ | Boundary matrix reuses the exact phase-1 dimension classifier for both groups |
| Input integrity | child dimension is missing, duplicated, unknown, has invalid order or a complete row has nil score | P0 | ✅ | Malformed-input matrix rejects all five cases; score/level completeness consistency is also checked |
| Runtime integration | publish freezes two group snapshots and submit persists two group results | P0 | ✅ | Staging真实发布冻结2组并绑定10维，提交写2条3/L3一级结果，清理后零残留 |

#### I5. Phase-1 Validity Calculation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Validity input | exactly 10 validity questions are all answered with raw values 1–5 | P0 | ✅ | Pure function sums original raw values only, requires forward metadata and returns complete integer score/status |
| Validity boundary | complete score is exactly 35 or 36 | P0 | ✅ | Explicit fixtures verify 35=`good`, 36=`questionable` with exact integer scores |
| Validity incomplete | one or more validity questions are unanswered | P0 | ✅ | 9/10 fixture preserves counts, keeps score nil and returns `incomplete` with `IsComplete=false` |
| Validity extremes | complete answers sum to 10 or 50 | P1 | ✅ | All-1 and all-5 fixtures classify good/questionable respectively |
| Input isolation | count is not 10, type is not validity, raw is outside 1–5, direction is not forward or identity/order is invalid | P0 | ✅ | Malformed matrix rejects count/type/raw low/raw high/reverse/blank/duplicate/order cases |
| Score independence | validity result is calculated alongside dimension/group results | P0 | ✅ | API accepts only `[]Phase1ValidityInput` and returns a separate value; no dimension/group object is consumed or mutated |
| Runtime integration | submit persists one validity result and management can filter good/questionable/incomplete | P0 | ✅ | Staging真实提交写10/good，good筛选1、questionable筛选0；管理详情与扩展导出已验证 |

#### I6. Phase-1 Unified 90-Question Runtime

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Publish | fixed draft has exactly 80 dimension and 10 forward validity questions | P0 | ✅ | Staging冻结2组/10维/90题，题型80/10，五个确认选项逐题匹配，重复发布幂等 |
| Publish | inventory/type/direction/group metadata is malformed or snapshot insert fails | P0 | ✅ | Staging通过临时BEFORE INSERT触发器强制题目快照写入失败；事务回滚后publish=0、group=0、question=0、group links=0，移除触发器后同一草稿可正常发布90题 |
| Answer | dimension question is forward/reverse; validity question is forward | P0 | ✅ | 本地正反向穷举保持GREEN；staging真实混合90题全部保存，效度raw=1并得到总分10 |
| Submit | all 90 questions are answered | P0 | ✅ | Staging数据库与API均验证10+2+1+1结果，total=90、dimension=80，完整性为1 |
| Submit | timeout leaves a dimension and a validity row unanswered | P0 | ✅ | Staging真实88/90：维度79/80、overall NULL、1条二级NULL、1条一级NULL、效度9/10+incomplete+NULL；默认排名0、正式报告拒绝 |
| Submit | request is repeated after successful commit | P0 | ✅ | Existing idempotent result lookup returns the frozen result without duplicate rows |
| Management | validity status is good, questionable or incomplete | P0 | ✅ | 管理筛选API真实验证good/questionable；前端139项测试覆盖状态、一级和阈值详情 |
| Statistics | score ranking sees an incomplete or validity-questionable result | P0 | ✅ | Staging分别验证timeout和40/questionable答卷均不进入默认overallScore排名；显式all和questionable筛选仍返回存疑答卷。独立常模/汇总统计端点尚未建设，继续作为产品增强 |
| Report | complete validity is good or questionable | P0 | ✅ | LIming/Ruiling approved the exact v2 content/workbook hashes for staging; 124 rows are active. A real historical v2 report generated and downloaded with SHA match as exactly 10 A4 pages while its v1 report/PDF remained intact. |

#### I7. Phase-1 Formal Report Framework

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Version dispatch | frozen versions are generic or phase-1 | P0 | ✅ | Generic keeps the existing dynamic four-level renderer; phase-1 selects an isolated fixed report-data contract without reusing evaluationAverage |
| Approval gate | phase-1 content package is missing, draft, retired or lacks either approval/hash/disclaimer | P0 | ✅ | Reject before report instance creation, Chromium rendering, PDF write or success audit with a stable not-approved error |
| Content completeness | approved package lacks overall/group/dimension/validity text, contains whitespace-only text, or has inconsistent disclaimers | P0 | ✅ | Current overall, 2 group descriptions, 10 current L1-L5 dimension texts, current validity notice and one consistent final disclaimer are required; FB-107 RED reproduced three gaps and GREEN rejects them without cross-version/audience fallback |
| Report data | complete good/questionable phase-1 result has 2 groups, 10 dimensions and validity | P0 | ✅ | Build a strong `competency-phase1-report-data-v1` DTO with `/50` overall, independent `/5` groups, ten dimensions, versions and participant-visible fields |
| Ten-page layout | phase-1 report framework is selected | P0 | ✅ | Fixed pages: cover, guide, person/overall/validity, groups, 10-axis radar, then five two-dimension detail pages |
| Phase-1 mobile options | five bordered radio cards render in one mobile column | P0 | ✅ FB-156 staging GREEN: real Chromium confirms five cards share left=29/right=361/width=332 at 390px; tablet/desktop layouts, touch targets and zero overflow remain unchanged |
| Validity privacy | participant report renders validity | P0 | ✅ | Show good/questionable notice but never expose raw validity score or the 35/36 threshold; management detail/export remain unchanged |
| Current candidate state | formal approvals/content source are absent | P0 | ✅ | Framework exists but `reportAvailable=false`; view/generate/download/batch actions remain disabled and direct backend requests are rejected |
| Dimension level labels | report renders L1-L5 for group or dimension results | P0 | ✅ | CSV catalog exposes separate secondary labels 差/较差/合格/较优秀/优秀 and group labels 低分/较低分/中分/较高分/高分; Vue uses separate mappings |
| Overall level labels | report renders excellent/good/qualified/weak/not_qualified | P0 | ✅ | Vue resolves formal CSV labels 优秀胜任/良好胜任/合格胜任/薄弱胜任/尚未胜任 from the generated catalog |
| CSV report catalog | phase-1 template renders names, definitions and labels | P0 | ✅ | Deterministic generator validates CSV first and emits the runtime catalog; freshness check prevents stale catalog use |
| Customer sample layout | phase-1 formal template renders ten A4 pages | P0 | ✅ | Chromium and pdfinfo verify cover, guide, personal/overall/validity, groups, ten-axis radar and five two-dimension pages as exactly 10 A4 physical pages |
| Missing formal text | validity-good notice or final disclaimer is absent | P0 | ✅ | Existing dual-approval/content-completeness gate continues to reject report data before Vue/PDF rendering; template does not invent fallback text |
| Customer DOCX fixed text | cover and reading-guide fixed copy is rendered | P0 | ✅ | Seven reviewed CSV rows drive report title, English title, slogan, three reading paragraphs and special notice; staging Chromium verifies every fixed excerpt |

#### I8. Phase-1 Customer Workbook Conversion

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Dimension identity cell | V1 workbook supplies `code + newline + name` in question and level sheets | P0 | ✅ | FB-108 parses both values, requires an exact A/B code-name pair, and preserves 90 questions plus 10×5 report texts |
| Dimension identity mismatch | supplied code and name belong to different confirmed dimensions | P0 | ✅ | FB-108 rejects with a deterministic code/name mismatch error before generating candidate or import artifacts |
| Validity classification | validity rows have an empty question-type cell under the validity layer | P0 | ✅ | Continue classifying all 10 rows from the explicit validity layer and force forward scoring metadata |
| Normalized CSV export | confirmed V1 workbook is exported for manual review | P0 | ✅ | Produces six UTF-8 BOM CSV files: package, 90 questions, 10 dimensions, 50 dimension-level texts, 5 overall texts and 4 static report texts |
| CSV-only mode | reviewer requests CSV artifacts without replacing candidate JSON/XLSX | P0 | ✅ | Writes only the requested CSV package; existing candidate JSON/XLSX hashes remain separately verifiable |
| Long-term import template | CSV package is retained as the future question/report source | P0 | ✅ | Includes package metadata, stable order, review workflow fields, machine level codes, exact 1/0 score boundaries and formula-injection rejection |
| Static report content | phase-1 report requires template/group/validity text | P0 | ✅ | Exports seven customer-DOCX template rows, two group rows and two validity rows; validity-good remains an explicitly labeled AI suggestion with `pending_review` |
| Approval metadata | CSV package is reviewed but dual approval is incomplete | P0 | ✅ | Approval stays draft with explicit blank approver/time/environment/disclaimer fields; approved-import validation rejects it until all fields and content SHA match |
| Candidate database import | validated draft CSV is loaded into staging for simulation | P0 | ✅ | Imported exactly 66 inactive temporary report-text rows and one draft package using deterministic IDs in one transaction; formal report gate remains closed |
| Candidate import rerun | the same CSV package is imported again | P0 | ✅ | Transaction SQL ran twice and retained 66 unique lookup rows plus one draft package |
| Candidate source parity | staging already contains the 90 questions and 10 dimensions | P0 | ✅ | Pre-import comparison matched all question code/type/dimension/item/content/observation/direction/status and dimension id/code/name/order/status fields |

### J. Competency Publish Snapshot

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Publish | exam missing, legacy, or invalid mode | P0 | ✅ | Locked exam is strictly dispatched; invalid mode exits transaction |
| Publish | competency draft has no dimensions | P0 | ✅ | Transaction rejects empty selection |
| Publish | selected dimension disabled or has zero enabled questions | P0 | ✅ | Revalidated in publish transaction with dimension-specific zero count error |
| Publish | source question metadata invalid | P0 | ✅ | Code/content/direction/dimension metadata validated before batch insert |
| Publish | valid draft | P0 | ✅ | Staging froze 2 dimensions, 4 questions, option JSON and publish audit in one transaction |
| Publish | dimension master changes after draft save but before publish | P0 | ✅ | Publish refreshes code/name/VIRD/category/core meaning/order from current enabled master data |
| Publish | repeated request after success | P0 | ✅ | Staging repeated publish returned existing 4-question summary without rebuilding |
| Publish | source question changes after publication | P0 | ✅ | Staging mutated D01-Q01 content/observation/direction/status after publish; snapshot and historical paper content remained byte-identical, then source restored |

### K. Competency Participant Paper and Answering

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Participant auth | valid candidate/tester participant token | P0 | ✅ | Candidate staging chain issued and accepted bound participant/paper tokens; tester shares same issuer |
| Participant auth | missing, expired, wrong-purpose or mismatched token | P0 | ✅ | Token unit matrix plus Handler binding checks reject before runtime service |
| Create paper | exam not published or has no snapshots | P0 | ✅ | Runtime requires publish status and nonempty snapshot set before insert |
| Start assessment UI | competency exam is still draft/unpublished | P0 | ✅ | Online list hides drafts; direct preparation URL disables start with Chinese publish guidance; backend maps sentinel error to Chinese |
| Create paper | first entry | P0 | ✅ | Staging verified complete four-question unique set and fixed 1-N order using crypto Fisher–Yates |
| Create paper | repeated entry with in-progress paper | P0 | ✅ | Staging returned the same paper ID and a newly signed paper token |
| Create paper | participant already completed | P0 | ✅ | Locked participant/paper state returns completed instead of creating another paper |
| Create paper | secure random source fails | P0 | ✅ | Injectable random-source unit test returns error; shuffle occurs before paper insert |
| Paper detail | valid paper token | P0 | ✅ | Staging repeated detail order matched; response hid dimension, direction and final score |
| Fill answer | raw value 1-5 on owned in-progress unexpired paper | P0 | ✅ | Staging saved four values and returned answered counts only |
| Fill answer | invalid value, foreign question, finished or expired paper | P0 | ✅ | Staging rejected raw=0, foreign paper question and finished writes; expired write triggered trusted timeout submit |
| Exam timing | competency total time is zero or negative | P0 | ✅ | Save/publish reject; no hidden default duration |
| Create paper | current time is after exam end time | P0 | ✅ | New paper path rejects before snapshot read, shuffle or insert |
| Create paper | paper started before exam end time and personal duration remains | P0 | ✅ | Existing paper restore path precedes end-time start guard and preserves personal limit time |
| Resume UI | traditional all-question page restores saved answers and scrolls to the first unanswered question | P0 | ✅ | UF-022 / FB-141 STAGING GREEN: deployed exact tested bundle; restored answer flags select the first unanswered card and scroll it into view after rendering |
| Resume UI | traditional single-question page opens the first unanswered question | P0 | ✅ | UF-022 / FB-141 STAGING GREEN: deployed exact tested bundle; flattened restored paper selects the first unanswered item before loading question detail |
| Resume UI | MBTI page opens the first unanswered question | P0 | ✅ | FB-141 STAGING GREEN: deployed exact tested bundle; regression suite confirms the existing first-unanswered behavior |
| Resume UI | 00401 competency page opens the first unanswered question while preserving saved answers | P0 | ✅ | UF-022 / FB-141 STAGING GREEN: real API restored the same paper with 10 answers and real Chromium opened question 11 |
| Mobile browser title | participant entry, preparation, answering, result and completion routes | P1 | ✅ | UF-023 / FB-144 STAGING GREEN: 13 public participant routes opt out; real staging Chromium returned an empty title through login, preparation and answering |
| Mobile browser title | administrator routes | P1 | ✅ | System title must remain available for the management interface |

### L. Competency Submit, Results, and Report Data

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Manual submit | unanswered question exists | P0 | ✅ | Staging rejected manual incomplete submit with first unanswered display order |
| Manual submit | all questions answered | P0 | ✅ | Staging atomically wrote one overall and two ordered dimension results |
| Timeout submit | partially answered or zero answered dimensions | P0 | ✅ | Staging Worker persisted partial 1/16 and zero 0/16 timeout results with nil zero-answer dimensions and no minimum-score fill |
| Timeout submit | participant claims timeout before personal limit | P0 | ✅ | HTTP rejects client timeout; service validates locked paper limit before timeout semantics |
| Submit | repeated/concurrent after completion | P0 | ✅ | Staging repeated submit returned `alreadySubmitted`; PK/unique indexes prevent duplicate result sets |
| Submit | successful completion | P0 | ✅ | Staging updated paper/participant and participant response contained no scores |
| Admin results | paging/detail with administrator JWT | P0 | ✅ | Staging detail and report data returned saved overall, two dimensions and four audit rows |
| Participant results | anonymous participant endpoint | P0 | ✅ | Router exposes no participant result route |
| Report data | frontline/leader audience snapshot | P0 | ✅ | Staging report used saved `leader` audience and the same saved score facts |
| Report data | formal text missing | P0 | ✅ | Report returns `reportTextReady=false` and explicit pending marker; no cross-audience fallback |
| Formal report | result is incomplete | P0 | ✅ | Formal report data validates `is_complete=1`; admin result detail remains available |

### M. Competency Result Management

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Result paging | sort by submitted time | P0 | ✅ | Staging API/browser returned High→Mixed→Low for fixed descending submission times; stable `paper_id` tie-breaker retained |
| Result paging | sort by overall score ascending/descending | P0 | ✅ | Staging API/browser ascending returned 2→6→8; whitelist rejects user SQL fragments |
| Result paging | sort by selected dimension ascending/descending | P0 | ✅ | Staging D01 descending returned 5→4→1; parameterized JOIN and measured-dimension ownership check active |
| Result paging | dimension sort requested without dimension ID | P0 | ✅ | Pure validation test rejects before query |
| Result paging | unknown sort field/direction | P0 | ✅ | Pure validation test rejects both cases before query |
| Result paging | participant is candidate or tester | P0 | ✅ | Staging returned candidate name/telephone/type for all three rows through one LEFT JOIN query without N+1 |
| Result paging | no results | P1 | ✅ | Non-nil result slice initialized; frontend renders explicit empty state |
| Result detail | open from result list | P0 | ✅ | Staging browser dialog showed overall=8, two ordered dimensions and all 16 per-question audit rows |
| Result authorization | authenticated user lacks administrator/exam:list/exam:export permission | P0 | ✅ | FB-078 returns HTTP 403 before binding/querying; administrator, global, exam:list and exam:export retain access; internal-token rendering remains independently authenticated |
| Result UI | legacy exam row | P0 | ✅ | Existing test-record/export/statistics entries remain in the non-competency branch |
| Result UI | competency exam row | P0 | ✅ | Dedicated `CompetencyResults` route and list command preserve `examId` context |
| Result UI | complete competency result opens test report | P0 | ✅ | Named route receives `params.paperId`; FB-067 RED→GREEN verifies the report page can read `$route.params.paperId` |
| Result UI | management list shows start/completion/duration and detail shows score sum | P1 | ✅ | Result paging projects `p.create_time AS started_at` and `user_time`; UI shows start/completion/duration and dimension `scoreSum`; RED→GREEN tests pass |
| Result filtering | name/telephone/completion filters are empty | P1 | ✅ | Empty/all normalization test keeps filters nil/empty and existing stable sort/pagination |
| Result filtering | name or telephone contains text | P1 | ✅ | Inputs are trimmed and applied through parameterized LIKE to frozen participant snapshots; count and rows share the helper |
| Result filtering | completion is complete/incomplete | P1 | ✅ | Pure test covers all/complete/incomplete and rejects unknown values; service maps to `is_complete=1/0` |
| Result paging | malformed JSON or wrong field type | P1 | ✅ | FB-083 checks `ShouldBindJSON` and returns “参数格式错误” before defaults or service queries |
| Result UI | query and reset buttons | P1 | ✅ | Component test verifies reset clears identity/completion/sorting, returns to page 1 and reloads |
| Result UI | overlapping list/detail requests return out of order | P1 | ✅ | FB-081 uses independent monotonic sequences for list/detail; only the latest response may update data/loading, and list requests receive a frozen query snapshot |
| Result UI | no complete row selected for batch report action | P1 | ✅ | Buttons bind disabled state to selected complete rows; `selectable` rejects incomplete rows |
| Result UI | complete phase-1 row exists after the ten-page renderer is implemented | P0 | ✅ | FB-113: complete rows are selectable; generation reaches the backend dual-approval gate, and the batch summary preserves its actionable business error |
| Phase-1 PDF | approved content package supplies a final disclaimer | P0 | ✅ | FB-114: template renders the frozen approved disclaimer and only falls back to the CSV sample notice when no approved disclaimer is present; real staging PDF text verified |
| Phase-1 PDF | report is compared with the authoritative customer DOCX/PDF | P0 | ✅ | FB-115: uses the original DOCX cover illustration and customer-style cover, matrix, overall orbit, group table/pie, centered radar, flowing dimension pages, typography and print density; final staging PDF is A4 10 pages and passed equal-DPI page review |
| Phase-1 PDF | customer uploads or adjusts the approved Word template | P0 | ✅ | FB-116/117: 49 hidden content controls and 12 native charts are customer-maintainable; local LibreOffice is enabled by default, Graph optional, Chromium fallback configured; staging real generation is GREEN |
| Phase-1 Word template | customer opens the maintainable DOCX in Microsoft Word | P0 | ✅ | UF-007/FB-117: normal sample values are visible, stable field keys remain in hidden content-control tags, and the final contract scan reports 49 unique tags with zero visible `{{...}}` tokens |
| Phase-1 template management | administrator opens the existing report-template page | P1 | ✅ | Staging `/qu/template` loads current name, size, time, SHA-256 and 49/12/0 contract above the retained MBTI table; desktop 1440×900 and mobile 390×844 browser checks pass with zero overflow |
| Phase-1 template management | administrator downloads the active template | P1 | ✅ | Real API returns configured DOCX with correct MIME and exact SHA; browser button sends the authenticated download request; component Blob/file-name test passes |
| Phase-1 template management | administrator uploads a valid DOCX | P0 | ✅ | Real API and browser upload the active DOCX, create one new 0600 backup, atomically replace it, refresh metadata and preserve exact SHA; template directory is limited to service user `liming:liming 750` |
| Phase-1 template management | upload is malformed, incomplete, duplicated, oversized or unauthorized | P0 | ✅ | Real duplicate-Tag upload is rejected while active SHA stays unchanged; unit tests cover missing/duplicate contract and administrator/global-vs-exam:list permissions; UI preserves selected file after errors |
| Phase-1 exam configuration | first save uses a custom subset of required participant fields | P0 | ✅ | UF-009/FB-118 staging GREEN: type switch initializes six defaults once; save/detail do not reapply defaults. Real create→first save→Detail→candidate page preserves and renders exactly name/gender/telephone |
| Phase-1 pre-exam page | participant reviews the assessment description before starting | P1 | ✅ | Shows the confirmed behavior/tendency guidance, two numbered answer rules, confidentiality statement and fixed red notice that all 90 questions are required; does not derive 80 from dimension-only counts |
| Phase-1 answer UI | one of the first 89 questions is answered and saved successfully | P0 | ✅ | Staging mocked-persistence E2E captures question/index, persists first, updates counts and advances exactly one; failed save stays for retry |
| Phase-1 answer UI | participant is viewing any question | P1 | ✅ | Internal question code and “五级量表” nodes/styles removed; staging browser confirms absence while progress, text, options and navigation remain |
| Phase-1 answer UI | the final question is answered successfully | P0 | ✅ | Staging E2E saves question 40, keeps it active and records zero submit requests; explicit submit confirmation remains |
| Phase-1 mobile entry | participant opens the generated QR URL inside WeChat | P0 | ✅ | QR targets `/exam-entry.html` without fragment; ES5 bridge validates parameters then redirects. WeChat Android UA real exam entry loads candidate page, no 401/console/request error or horizontal overflow |
| Phase-1 PDF | approved Word template is converted by the MBTI-style local LibreOffice path | P0 | ✅ | Staging LibreOffice 24.2 generated existing complete paper as exactly 10 A4 pages; producer, required text, unresolved tags, API/file/DB hash and cleanup all passed |
| Phase-1 PDF | local LibreOffice is unavailable, times out or returns a non-PDF file | P0 | ✅ | Unit tests reject queue cancellation, command failure and non-PDF without exposing command output, and clean temporary workspaces; Chromium fallback remains enabled and its prior real E2E is GREEN |
| Phase-1 Word chart data | template uses an embedded Excel workbook | P0 | ✅ | Candidate contains exactly one `word/embeddings/competency-phase1-chart-data.xlsx`; all 12 chart relationships use internal `relationships/package`, formulas reference the embedded file, and external links=0 |
| Phase-1 Word chart data | report generation updates chart values | P0 | ✅ | Focused test verifies one payload updates chart3 cache and embedded Sheet1 group cells, Sheet2 radar cells and ten doughnut score/remainder rows; real LibreOffice conversion succeeds as A4 12 pages |
| Phase-1 Word chart data | embedded workbook or chart relationship is missing/malformed | P0 | ✅ | Candidate passes the existing 49-control/12-chart upload contract; embedded workbook/relationship/formula contract tests reject structural drift. Active staging template remains unchanged pending explicit activation |
| Result UI | one or more complete rows selected for batch generation | P1 | ✅ | Component test verifies one generation call per selected paper, loading cleanup and summary feedback |
| Result UI | one or more complete rows selected for batch download | P1 | ✅ FB-157 staging GREEN | One authenticated request submits a frozen selection snapshot and produces one ZIP browser download; real staging ZIP contains both selected current PDFs under unique sanitized participant+paper names and adds two download audits. Missing/unapproved/invalid-path reports reject before download |
| Result UI | selection changes while batch generation/download is running | P1 | ✅ | FB-082 captures a filtered/shallow-copied complete-result snapshot at task start; loops, progress and totals never read live selectedRows afterward |
| Result UI | row view/answer-detail/download actions | P1 | ✅ | Source/component tests verify legacy labels/order while retaining competency report, detail and download APIs |
| Result UI | result list score columns | P1 | ✅ FB-159 staging GREEN: Nginx-served result-list resources contain no “评价均值”; overall score, selected dimension score, details, exports, reports and persisted scoring data retained |
| Result UI | mobile viewport under 768px | P2 | ✅ | FB-087 wraps heading/toolbars, makes filters full width, preserves table scrolling and opens a full-screen one-column detail dialog |

### M2. Temporary Competency Formal Report and PDF

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Report template | cover identifies exam, participant, audience and generation date | P1 | ✅ | Staging 00401 PDF cover shows exam, participant, frozen audience and actual render date |
| Report template | reading guide explains 1–5 scale, aggregation and comparison boundary | P1 | ✅ | PDF explains 1.00–5.00, dimension average, overall sum/evaluation mean and cross-combination boundary |
| Report template | participant fields follow exam requiredFields | P1 | ✅ | Template filters frozen identity fields by exam requiredFields; empty configuration shows all |
| Report template | overview and measured dimensions have printable score charts | P1 | ✅ | Staging PDF renders four-level overall scale and five measured-dimension CSS bars without external chart runtime |
| Report template | each measured dimension shows frozen core meaning and temporary interpretation | P1 | ✅ | Nine-page 00401 PDF contains all five published core meanings, exact temp-v1 interpretations and development prompts |
| Report template | six configured participant fields fit on the overview page | P1 | ✅ | FB-073 RED→GREEN; five retained staging reports stay at 9 pages without splitting overall interpretation |
| Report template | score bands 1–5 and both audiences render consistently | P1 | ✅ | Five independent configs cover averages 1/2/3/4/5, overall 5/10/15/20/25, three frontline and two leader PDFs |
| Report text | active temporary content version exists for audience, overall level and every measured dimension level | P0 | ✅ | 392 temp-v1 rows cover 2 audiences × overall/dimensions × 4 levels; exact matches are frozen into instance snapshot |
| Report text | any required text is missing or belongs to another audience/version | P0 | ✅ | Pure service test rejects missing dimension level and cross-audience fallback before rendering |
| Report text | required exact lookup key is missing | P1 | ✅ | Error identifies contentVersion, audience, dimension (or overall), and level |
| Report text | temporary content is rendered | P0 | ✅ | Staging PDF text contains “临时测试报告” and “不可作为人才决策依据” |
| Internal report authentication | token is supplied in URL/query instead of `X-Internal-Token` | P0 | ✅ | FB-079 sends only `paperId` in query, places the token in `X-Internal-Token`, and rejects even a correct query token with HTTP 401 |
| Generate report | result is incomplete | P0 | ✅ | Existing FB-047 formal report guard rejects before instance/PDF writes |
| Generate report | complete result has no prior instance | P0 | ✅ | Staging created instance, rendered Chromium PDF, persisted path/hash/size and completed status |
| Generate report | same version already completed and force=false | P0 | ✅ | Unique paper+version and existing-file guard return the same instance without duplication |
| Generate report | concurrent requests target the same paperId | P0 | ✅ | FB-080 uses a stable bounded 64-stripe lock on the singleton report handler and locks before instance lookup through final audit/read, so waiting force=false requests re-read and reuse completed output |
| Regenerate report | force=true for same content version | P0 | ✅ | Same instance is refreshed and regenerate audit action is recorded by the dedicated branch |
| Generate report | render or file write fails | P0 | ✅ | Handler marks instance failed and never sets participant pdf_flag success |
| Generate report | success audit insert fails after PDF metadata update | P0 | ✅ | FB-084 inserts success audit through the same GORM transaction as report metadata and participant PDF state; rollback removes the new file |
| Download report | completed instance path is inside configured upload root and file exists | P0 | ✅ | Staging downloaded application/pdf, SHA-256 matched instance, and audit count was 1 |
| Download report | same-name participants or spaces/plus signs in response filename | P1 | ✅ | FB-085 includes paperId in frontend names and encodes server filename* with `%20`/`%2B` RFC5987-compatible percent encoding |
| Download report | missing instance/file or path escapes upload root | P0 | ✅ | `filepath.Rel` allow-root guard and completed-instance lookup reject invalid paths |
| Download report | HTTP 200 response body is a JSON business error rather than a PDF | P0 | ✅ | FB-077 rejects every non-`application/pdf` Blob, parses the backend message with Blob.text/FileReader fallback, and prevents `saveAs`/success counting |
| Delete exam | competency report instances/audits/files exist | P0 | ✅ | Staging full-chain cleanup removed report audit/instance before paper and removed allowed-root PDF; remaining=0 |

### N. Legacy and Competency Question Management Isolation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Legacy question paging/list/export | source question has `dimension_id IS NULL` | P0 | ✅ | Traditional endpoints retain legacy questions only |
| Legacy question paging/list/export | competency source question has non-NULL `dimension_id` | P0 | ✅ | Excluded before count/list/export by parameter-free SQL predicate |
| Legacy question detail | competency source question ID supplied | P0 | ✅ | Legacy detail query includes `dimension_id IS NULL` and returns not found |
| Legacy question save | new request contains competency metadata | P0 | ✅ | Metadata presence is rejected before validation/transaction |
| Legacy question save | existing competency source question ID supplied | P0 | ✅ | ID guard runs before transaction and preserves all metadata |
| Legacy question delete | any selected ID is a competency source question | P0 | ✅ | Whole batch rejected before transaction |
| Legacy repo batch action | any selected ID is a competency source question | P0 | ✅ | Rejected before deleting/rebuilding associations |
| Legacy question UI | row has nil/empty answer list | P0 | ✅ | Safe formatter renders `—`; direct option indexing removed |

### O. Legacy Question Form Repository Selection

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question create | one repository selected through `repoId` | P0 | ✅ | Bound field validates and synchronizes `repoIds=[repoId]` before validation |
| Question create | no repository selected | P0 | ✅ | Bound `repoId` rule shows “必须选择一个题库！” and synchronization produces an empty array |
| Question edit | repository selection unchanged | P0 | ✅ | Synchronization preserves one matching `repoIds` value |
| Question edit | repository changed from old to new | P0 | ✅ | Stale `repoIds` is always overwritten by the current selection before validation/save |

### P. Legacy Question Import Upload Routing

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import upload | development build | P0 | ✅ | Uses `VUE_APP_BASE_API`; development resolves to `/dev-api` |
| Import upload | staging build | P0 | ✅ | Uses `VUE_APP_BASE_API`; staging resolves to `/stage-api` |
| Import upload | production build | P0 | ✅ | Production artifact contains `/prod-api/exam/api/qu/qu/import-excel` and zero `/dev-api` matches |
| Import upload | authenticated administrator | P0 | ✅ | Upload headers include current `Authorization: Bearer <token>` |

### Q. Legacy Question Paging Performance

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question paging | page has no records | P1 | ✅ | Returns `records=[]`; loader exits before relation queries |
| Question paging | page has one or more records | P1 | ✅ | All answers loaded by one ordered `qu_id IN ?` query |
| Question paging | question has one or more repository relations | P1 | ✅ | All ordered relations loaded once; first `(sort,id)` relation retained |
| Question paging | repository relation points to deleted repository | P1 | ✅ | `COALESCE` preserves `[已删题库:<id>]` marker behavior |
| Question paging | relation query fails | P1 | ✅ | Query error returns “查询题目关联数据失败” without partial rows |

### R. Legacy Repository Batch Association

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Batch request | duplicate/blank question or repository IDs | P0 | ✅ | IDs normalized before database work |
| Batch add | any question does not exist or is competency | P0 | ✅ | Guard plus transactional legacy query rejects entire batch before writes |
| Batch add | valid questions and repositories | P0 | ✅ | Replaces selected questions' associations atomically with `CreateInBatches` |
| Batch remove | valid question/repository pairs | P0 | ✅ | Deletes selected pairs atomically |
| Batch reorder | old and requested repositories affected | P0 | ✅ | Reorders every affected repository by `(sort,id)` |
| Batch statistics | association set changes | P0 | ✅ | Refreshes every old/requested repository in the same transaction |
| Batch operation | any read/write/reorder/stat refresh fails | P0 | ✅ | Every error returns from the transaction and rolls back the batch |

### S. Legacy Question Excel Import Atomicity

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import repository | sheet repository exists | P0 | ✅ | Reused inside the import transaction |
| Import repository | sheet repository does not exist | P0 | ✅ | Created inside the same transaction as questions |
| Import extra repositories | all referenced IDs exist | P0 | ✅ | All IDs validated, every relation insert error checked |
| Import extra repositories | any referenced ID does not exist | P0 | ✅ | Rejected before question insertion and transaction rolled back |
| Import questions | any question/answer/relation insert fails | P0 | ✅ | Repository and all imported rows roll back together |
| Import statistics | import succeeds | P0 | ✅ | Every sheet/extra repository refreshed before commit |

### T. Legacy Question Export Performance

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question export | no matching questions | P1 | ✅ | Generates header-only workbook; loader exits before relation queries |
| Question export | one or more matching questions | P1 | ✅ | Reuses two batch relation queries for all exported questions |
| Question export | relation query fails | P1 | ✅ | Returns controlled error before writing workbook rows |
| Question export | question has multiple repositories | P1 | ✅ | Preserves all repository IDs in `(sort,id)` order |

### U. Legacy Question Deletion Integrity

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question delete | selected question belongs to competency | P0 | ✅ | Existing dedicated-API guard rejects the whole batch |
| Question delete | selected traditional question is referenced by a paper | P0 | ✅ | Rejects before opening delete transaction and reports reference count |
| Question delete | selected traditional questions are unreferenced | P0 | ✅ | Deletes answers/relations before questions in one transaction |
| Question delete | any child delete or statistic refresh fails | P0 | ✅ | Returns error and rolls back all selected questions |

### V. Legacy Referenced Question Immutability

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question save | new traditional question | P0 | ✅ | No historical reference check required |
| Question save | existing unreferenced traditional question | P0 | ✅ | Reference guard passes and existing transaction remains available |
| Question save | existing question referenced by any paper | P0 | ✅ | Rejected before answer/repository replacement |
| Question save | paper reference lookup fails | P0 | ✅ | Returns controlled error and does not open write transaction |

### W. Legacy Repository Paging Boundary

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Repository paging | size is zero/negative | P1 | ✅ | Normalized by shared `capPageSize` |
| Repository paging | size exceeds global maximum | P1 | ✅ | Capped before OFFSET/LIMIT query |
| Repository paging | count query fails | P1 | ✅ | Returns “查询题库总数失败” |
| Repository paging | list query fails | P1 | ✅ | Returns “查询题库列表失败” without partial records |

### X. Legacy Question Save Consistency

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question save | repository IDs contain blanks/duplicates | P0 | ✅ | Normalized before required validation and writes |
| Question save | any repository ID does not exist | P0 | ✅ | Rejected inside transaction before question insert/update |
| Question edit | original question lookup fails | P0 | ✅ | Returns error; never replaces create time or continues |
| Question edit | old association lookup/delete fails | P0 | ✅ | Returns error and rolls back question/answers/relations |
| Question save | all repositories and child rows valid | P0 | ✅ | Commits question, answers, relations and statistics atomically |

### Y. Legacy Repository Save Integrity

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Repository create | title is blank | P0 | ✅ | Trimmed and rejected before database write |
| Repository create | valid title | P0 | ✅ | Server initializes zero counts and both timestamps |
| Repository update | client sends count/create-time fields | P0 | ✅ | Only code/title/remark/update_time are updated; row is reloaded for response |
| Repository update | repository ID does not exist | P0 | ✅ | RowsAffected=0 returns “题库不存在” |

### Z. Legacy Question and Repository Read Reliability

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question list | query succeeds or returns no rows | P1 | ✅ | Returns records/non-nil empty array |
| Question list | query fails | P1 | ✅ | Returns “查询题目列表失败” |
| Question detail | answer/repository/code relation query fails | P1 | ✅ | Returns specific controlled error, never partial detail |
| Repository list | query succeeds or returns no rows | P1 | ✅ | Returns records/non-nil empty array |
| Repository list | query fails | P1 | ✅ | Returns “查询题库列表失败” |

### AA. Question Bank Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Legacy/competency source isolation | P0 | ✅ | Staging traditional paging total=855; first page competency rows=0; browser table renders without option-index error |
| Question form repository synchronization | P0 | ✅ | Staging browser validates current repoId and overwrites stale repoIds |
| Environment-aware authenticated upload | P0 | ✅ | Browser upload URL is `/prod-api/.../import-excel` with Bearer header |
| Repository batch add/remove atomicity | P0 | ✅ | Duplicate IDs normalize; invalid repository batch rolls back and preserves existing association |
| Excel import atomicity | P0 | ✅ | Valid workbook imports; invalid extra repository rolls back sheet repository and question |
| Referenced/competency question guards | P0 | ✅ | Historical edit/delete and competency delete rejected; SQL counts unchanged |
| Cleanup and integrity | P0 | ✅ | Temporary rows=0, competency orphans=0, critical logs=0 |

### AB. Competency Question Bank Entry 00401

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Repository paging | administrator opens repository list | P0 | ✅ | Adds virtual code `00401`, title “胜任力测验题库”, and live competency question count |
| Repository paging | title filter matches/does not match 00401 | P1 | ✅ | Includes virtual row only when code/title filter matches |
| Repository action | administrator opens 00401 | P0 | ✅ | Routes to dedicated read-only competency question page, never traditional editor |
| Competency question paging | no filters | P0 | ✅ | Returns only `dimension_id IS NOT NULL` rows with dimension fields |
| Competency question paging | dimension/status/code/content filters | P0 | ✅ | Applies parameterized filters and stable dimension/item ordering |
| Competency question UI | 384 rows exist | P0 | ✅ | Paginated table displays code, dimension, content, observation point, direction, status |
| Virtual repository mutation | edit/delete/batch selected | P0 | ✅ | UI disables selection/edit semantics; backend rejects traditional delete for virtual ID |

### AC. Competency Question Bank 00401 Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Repository list visibility | P0 | ✅ | Browser first row `00401 / 胜任力测验题库 / 384`, total repositories=7 |
| Dedicated navigation | P0 | ✅ | Click opens `/#/exam/competency/questions`; virtual row checkbox disabled |
| Dedicated question data | P0 | ✅ | API/browser total=384, first page=20, first row D01-Q01 with complete metadata |
| Filter correctness | P0 | ✅ | API D01 + enabled filter returns exactly 8 rows |
| No traditional association pollution | P0 | ✅ | 00401 is virtual; existing legacy=855 and competency=384 physical source counts unchanged |
| Cleanup/health | P0 | ✅ | Temporary rows=0, service/nginx active, health OK, critical logs=0 |

### AD. Competency Question Edit and Status

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question update | ID is empty | P0 | ✅ | Rejected before database access |
| Question update | source question does not exist or is legacy | P0 | ✅ | `id + dimension_id IS NOT NULL` RowsAffected guard rejects without fallback |
| Question update | content is blank | P0 | ✅ | Rejected with specific required-field message |
| Question update | observation point is blank | P0 | ✅ | Rejected with specific required-field message |
| Question update | direction is not `forward/reverse` | P0 | ✅ | Rejected before write |
| Question update | status is not 0/1 or frontend sends empty string | P0 | ✅ | Flexible JSON input rejects empty/nil/fractional/out-of-range values |
| Question update | valid editable fields | P0 | ✅ | Updates only content/observation/direction/status/remark/update_time |
| Question update | identity/history fields supplied by client | P0 | ✅ | Independent request omits identity/history fields; published snapshots untouched |
| Question update | database update fails | P0 | ✅ | Returns controlled error; frontend catch keeps dialog open |
| Question UI | edit opens from row | P0 | ✅ | Pre-fills editable copy and shows immutable identity context |
| Question UI | save succeeds | P0 | ✅ | Closes dialog, notifies, and refreshes current page |
| Question UI | save in progress or fails | P1 | ✅ | Loading disables duplicate submit; failure keeps dialog open |

### AE. Competency Question Edit Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Editable field persistence | P0 | ✅ | Temporary source question persisted content/observation/direction/status/remark |
| Identity field protection | P0 | ✅ | Forged code/dimension/item fields ignored; create_time unchanged |
| Invalid and cross-flow requests | P0 | ✅ | Empty status and legacy question ID rejected without mutation |
| Browser edit flow | P0 | ✅ | Dialog opens with immutable identity and all editable fields pre-filled |
| Cleanup/health | P0 | ✅ | Temporary question=0, service/nginx active, health OK, critical logs=0 |

### AF. Competency Question Import UI

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import template | administrator clicks download | P0 | ✅ | Downloads dedicated nine-column xlsx with fixed filename |
| File selection | no file selected | P0 | ✅ | Preview/formal import disabled with explicit guidance |
| File selection | extension is not xlsx or file exceeds 10 MiB | P0 | ✅ | Rejected client-side; backend guard remains authoritative |
| Import preview | request succeeds with valid rows | P0 | ✅ | Shows counts/normalized rows and retains SHA-256 |
| Import preview | validation has any errors | P0 | ✅ | Shows every row message and disables formal import |
| Import preview | file changes after preview | P0 | ✅ | Clears preview/hash and requires a new preview |
| Formal import | preview is valid and administrator confirms | P0 | ✅ | Re-uploads same file with expectedHash, refreshes list/dimensions |
| Formal import | request fails | P0 | ✅ | Keeps dialog, file and preview state for retry |
| Duplicate submission | preview/import request is running | P1 | ✅ | Loading state disables repeated action |

### AG. Competency Question Import UI Staging Acceptance

| Scenario | Priority | Coverage | Verified result |
|----------|----------|----------|-----------------|
| Dedicated template download | P0 | ✅ | HTTP 200, xlsx MIME, 6512 bytes |
| Template instruction row | P0 | ✅ | Initial real preview reproduced FB-064; after deployment preview skips exact instruction row |
| Valid preview | P0 | ✅ | Temporary unique row: success=1, errors=0, digest retained and confirm enabled |
| Formal import | P0 | ✅ | Browser confirmation imported one row; list refreshed from 384 to 385 |
| Cleanup | P0 | ✅ | Deleted by queried primary key; DB/list returned to 384, temporary question/relation/session/files=0 |
| Service health | P0 | ✅ | Backend/nginx active, health OK, recent panic/fatal/segmentation/import-failed counts all 0 |

### AG2. Competency Question Import/Export Round-Trip Acceptance

| Scenario | Priority | Coverage | Verified result |
|----------|----------|----------|-----------------|
| Template examples | P1 | ✅ | Four-row template contains one valid forward and one valid reverse D01 example |
| Template preview | P0 | ✅ | Real staging preview returned success=2, errors=0 and a 64-character SHA-256 |
| Template import | P0 | ✅ | Same file/hash imported two rows atomically; source question count changed 384→386 |
| Question export | P0 | ✅ | Nine import-compatible columns exported all 386 rows and both temporary examples with correct direction/status |
| Result exports | P0 | ✅ | Existing five-dimension 40/40 result exported 1 summary, 40 details and 40 dictionary rows from both endpoints; normalized content matched |
| Cleanup and health | P0 | ✅ | Temporary examples/session/files=0, source questions restored to 384, health OK, recent critical errors=0 |

### AH. Competency Dimension Master Maintenance

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Dimension update | ID is empty or dimension does not exist | P0 | ✅ | Rejects before/without mutation; no create fallback |
| Identity boundary | client supplies code/create_time | P0 | ✅ | Stable dimension code, ID and create time remain unchanged |
| Required fields | name/VIRD/category/core meaning is blank | P0 | ✅ | Rejects with field-specific message before database write |
| Fixed classifications | VIRD/category is outside supported master values | P0 | ✅ | Rejects arbitrary classification values at HTTP boundary |
| Display order | value is empty, fractional, below 1 or above 48 | P0 | ✅ | Flexible JSON input rejects invalid order |
| Display order | requested 1-48 position is occupied | P0 | ✅ | Atomically swaps the two display orders without violating unique index |
| Status | value is empty, fractional or outside 0/1 | P0 | ✅ | Flexible JSON input rejects invalid state |
| Uniqueness | name conflicts with another dimension | P0 | ✅ | Rejects with controlled conflict message; original row unchanged |
| Valid update | descriptive fields/order/status are valid | P0 | ✅ | Field whitelist update, re-read and return persisted row |
| Historical behavior | source dimension changes after an exam was published | P0 | ✅ | Existing exam snapshots/results unchanged; only future save/publish uses master data |
| Migration rerun | administrator-maintained row already exists | P0 | ✅ | Seed rerun does not overwrite maintained fields or status |
| Maintenance UI | page loads all 48 dimensions | P0 | ✅ | Stable display order, enabled question count, local filters and 20-row paging visible |
| Maintenance UI | edit succeeds | P0 | ✅ | Dialog closes, success feedback and list refresh |
| Maintenance UI | status changes | P0 | ✅ | Explicit confirmation explains impact on new exams; published exams unaffected |
| Maintenance UI | request fails | P1 | ✅ | Keeps dialog values and allows retry; loading prevents duplicate submit |

### AI. Competency Dimension Maintenance Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Field update and identity protection | P0 | ✅ | D01 descriptive/status update persisted; forged code/create_time ignored |
| Occupied order swap | P0 | ✅ | D01 1→2 and D02 2→1 completed atomically, then restored 1/2 |
| Invalid requests | P0 | ✅ | Duplicate name and empty status rejected without partial mutation |
| Migration rerun protection | P0 | ✅ | Rerunning 002 retained temporary maintained name/order/status; restore followed |
| Browser navigation and list | P0 | ✅ | 00401 entry opens maintenance page; total 48, default page 20, D01/D02 metadata correct |
| Browser filtering and edit | P0 | ✅ | D42 filter returns one correct row; edit dialog pre-fills all fields and status-change confirmation is explicit |
| Cleanup and integrity | P0 | ✅ | 48 unique codes/names/orders, order 1-48, all enabled, temporary values/session/files=0 |
| Service health | P0 | ✅ | Backend/nginx active, health OK, recent panic/fatal/duplicate/unknown-column counts all 0 |

### AJ. Competency Dynamic Result Export

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Dispatch | assessment is legacy/001/002/003 | P0 | ✅ | Existing template and wide-export behavior remains unchanged |
| Dispatch | assessment is competency | P0 | ✅ | Both existing export endpoints call the same competency workbook builder |
| Authorization | login missing or user lacks export permission | P0 | ✅ | Rejects before any result/snapshot query |
| Exam validation | examId empty/not found or invalid competency mode | P0 | ✅ | Returns controlled error, never falls back to legacy export |
| Data source | submitted complete or timeout-incomplete result exists | P0 | ✅ | Reads persisted result, dimension result, paper answer and publish snapshot values; no score recalculation |
| Result summary | selected dimensions vary by exam | P0 | ✅ | Dynamic columns follow frozen display order and include persisted dimension/overall scores |
| Result summary | incomplete timeout result | P0 | ✅ | Completion rate/integrity retained; missing dimension score exported as blank |
| Question detail | answered/unanswered item | P0 | ✅ | Includes personal sort, snapshot code/content/dimension/observation/direction, raw value/text, final score and answered flag |
| Question dictionary | personal random orders differ | P0 | ✅ | Exports one stable publish snapshot dictionary ordered by snapshot order |
| Empty result | published exam has no saved result | P1 | ✅ | Returns workbook with three sheets and headers, not a false-success malformed file |
| Query failure | any summary/dimension/detail/dictionary query fails | P0 | ✅ | Aborts before response body and returns controlled error |
| Privacy | non-super-admin exports telephone | P0 | ✅ | Masks telephone consistently with existing export policy |
| Frontend entry | competency exam dropdown | P0 | ✅ | Shows result page plus summary/raw export actions; confirmation and download filename are explicit |

### AK. Competency Dynamic Export Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Deployment backup and artifact identity | P0 | ✅ | Database backup created; deployed backend/frontend hashes equal local artifacts |
| Real persisted result set | P0 | ✅ | Three candidates, 16 questions each, overall 2/6/8 and D01 1/4/5 |
| Summary endpoint workbook | P0 | ✅ | HTTP xlsx, three sheets, 3 summary rows, D01/D02 dynamic columns and persisted scores verified |
| Raw-answer endpoint workbook | P0 | ✅ | Same normalized three-sheet content as summary endpoint; 48 detail rows and 16 dictionary rows |
| Response contract | P0 | ✅ | RFC 5987 competency filename and xlsx MIME/size verified for both endpoints |
| Browser entry | P0 | ✅ | Competency dropdown shows result + both export actions; confirmation describes all three sheets |
| Cleanup and integrity | P0 | ✅ | Full-chain cleanup=0, temporary exam/results/session/files=0, result orphans=0, competency source questions=384 |
| Service health | P0 | ✅ | Backend/nginx active, health OK, recent panic/fatal/export-failed/unknown-column counts all 0 |

### AL. Competency Expiry Worker Batch and Concurrency

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Startup | overdue papers exist when service starts | P0 | ✅ | Runs one scan immediately instead of waiting the first interval |
| Scheduled scan | no overdue papers | P1 | ✅ | One bounded query, no writes, no error |
| Owner resolution | candidate/tester paper | P0 | ✅ | Resolves participant type in the batch scan query, no per-paper owner query |
| Owner resolution | owner missing | P0 | ✅ | Logs paper ID, leaves state unchanged for repair/retry, continues batch |
| Batch bound | overdue count exceeds configured size | P0 | ✅ | Stable `limit_time,id` order and configured LIMIT |
| Timeout submit | partially answered paper | P0 | ✅ | Same submit service persists partial averages and marks incomplete/timeout |
| Timeout submit | zero-answer dimension/paper | P0 | ✅ | No fake minimum scores; nil dimension score and zero effective dimensions retained |
| Submit failure | one paper fails | P0 | ✅ | Logs without exposing token/data, continues other papers, failed paper retries next scan |
| Concurrent submit | frontend and Worker submit same expired paper | P0 | ✅ | Row lock/idempotency yields exactly one overall result and one result set |
| Capacity | 100 independent shuffles of 384 questions | P1 | ✅ | Every in-memory order is complete/unique and sample contains multiple distinct permutations |
| Shutdown | context cancelled | P0 | ✅ | Immediate scan/ticker goroutine exits cleanly |

### AM. Competency Expiry Worker Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Deployment backup and identity | P0 | ✅ | Database backup created; deployed backend SHA-256 equals local Linux artifact |
| Concurrent expired submit | P0 | ✅ | Two simultaneous manual/timeout-compatible requests produced 1 overall + 2 dimension results |
| Candidate partial timeout | P0 | ✅ | 1/16 answered, effective dimensions=1, overall=3, one nil dimension, incomplete timeout |
| Tester zero-answer timeout | P0 | ✅ | 0/16 answered, effective dimensions=0, overall=0, two nil dimensions, incomplete timeout |
| Startup immediate scan | P0 | ✅ | Two newly expired papers submitted within 15 seconds after service restart, below 30-second interval |
| Cleanup and integrity | P0 | ✅ | Full-chain cleanup=0, temporary exam/results=0, result orphans=0, running expired competency papers=0, source questions=384 |
| Service health | P0 | ✅ | Backend/nginx active, health OK, recent panic/fatal/worker-failed/owner-missing counts all 0 |

### AN. Full 48-Dimension Capacity Chain

| Gate | Priority | Coverage | Planned assertion |
|------|----------|----------|-------------------|
| Publish | all 48 enabled dimensions selected | P0 | ✅ | Staging froze exactly 48 dimensions and 384 unique source questions in 0.098s |
| Paper creation | 100 participants start independently | P0 | ✅ | Created exactly 100 papers and 38,400 paper-question rows |
| Set integrity | each 384-question paper | P0 | ✅ | Every paper contained 384 unique question IDs/codes with no missing or duplicate rows |
| Random independence | compare 100 persisted orders | P0 | ✅ | Persisted orders produced 100/100 distinct SHA-256 hashes |
| Refresh stability | reload each paper detail | P0 | ✅ | All 100 orders remained byte-for-byte stable after second read |
| Runtime response | 384-question detail | P1 | ✅ | Create chain p50/p95/max=0.479/0.758/0.837s; refresh=0.063/0.105/0.125s on staging |
| Source isolation | capacity papers use published snapshot | P0 | ✅ | All 38,400 rows bound exam_question_id; snapshot count/unique source/code=384 |
| Cleanup | delete capacity exam | P0 | ✅ | Full-chain transaction removed all capacity data; cleanup_remaining=0 |
| Service health | after capacity cleanup | P0 | ✅ | Backend remained healthy; final service/log check follows stage-6 closeout |