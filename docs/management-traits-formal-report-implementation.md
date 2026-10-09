# 002管理特质正式报告与两端UI：实施契约

日期：2026-10-06。状态：**POLICY CONFIRMED；slice1正式版本登记/审批/启用/撤销后端本地已实现；正式PDF和两端UI尚未实现。**

最新政策：Q1=A，同一授权系统账号允许分别完成内容与心理测量两项独立签署；Q2=B，精确匹配获批题本/评分/常模的完整新版TEST run可追加独立正式PDF，TEST原PDF/current不覆盖、legacy历史不转换；Q3=B，撤销禁止新生成/新启用，但授权管理员可继续view/download历史formal原PDF，明确revoked并保文件及audit。用户选择政策不构成205条内容或模板的正式批准。

[本地slice1证据与实际限制](management-traits-formal-registry-local-20261006.md)：6229pass/0fail/9原skip、server/all-build/vet原生exit0；五管理API、独立三表DDL（未执行）及旧守卫配置表严格分类。只完成版本管理后端，不称正式报告交付。真实候选仍有16条外部LINK字段，正式门禁拒绝，未改候选/TEST资产。

## 1. 范围基线与授权

用户已选择“实现正式报告功能，并改造两端UI”，不是继续harness；本次明确授权第一个完整本地逻辑slice版本登记/审批/启用/撤销及必要schema/守卫/API/测试。B区存Go代码，C区存文档与SQL；不修改客户材料、TEST运行资产或环境配置，不执行SQL、浏览器、SSH、部署、重启、历史迁移。原仅文档授权属于上阶段，已由本次本地code授权限定替代。

后续本地功能范围仅00201/00202新建默认新版；历史old results/PDF原样保留，00401、MBTI、001不动。正式功能的开发授权不等于具体内容批准，也不等于启用或发布生产。一次一个逻辑切片，契约经主协调者判定及必要业务答复后才实施下一步。

发现方法：完整读取项目记忆、既有客户决策与设计，读取当前报告HTTP/service/Word/model/001 SQL，并搜索真实调用点。两种题本、管理端和参与者端均在范围内；下文只记录本地源码合同，数据库实际结构沿既有证据，**本轮没有重新查询任何数据库或环境**。

依据：[用户评分/交付决定](management-traits-word-assessment-20261001.md#L287-L316)、[候选负责人及批准边界](management-traits-word-candidate-verification-20261001.md#L37-L61)、[原R2设计](management-traits-word-design-20261001.md)。已读取feature-inventory及creating-implementation-plan方法；因业务阻塞未关闭，不生成独立可执行plan、tasks、checkpoint或现代化工作流。

## 2. 当前真实能力与技术约束

| 现状证据 | 实际结论 | 对正式功能的约束 |
|---|---|---|
| [报告HTTP](../Go-based%20Refactored%20System/internal/handler/management_traits_report_runtime.go#L18-L59) | 只有generate-test/view/download；MNG_TEST_REPORT_ENV只接受local/staging | 不改环境伪装、不复用TEST路由返回正式PDF；新增正式专属入口，默认关闭 |
| [TEST内容加载](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L15-L101) | SHA锁定工作簿；205逻辑规则；类型不提供approved字段 | 现资产只能作为新正式候选的来源，不自动变成正式批准包；内容修订新版本新SHA |
| [TEST DTO](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L147-L206) | TestOnly=true；要求new_creation/submitted_snapshot/completed及完整事实复核 | 正式DTO独立purpose及版本身份，不将TestOnly=false或删标签当正式实现 |
| [Word门禁](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word.go#L126-L143) | 硬编码TEST模板SHA、DTO schema、来源SHA及两个TEST标签 | 保留TEST模板与函数的现有合同；正式渲染读取已批准独立模板及其绑定manifest |
| [报告持久化](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go#L96-L109) | TEST revision/current/audit同事务；current按paper_id upsert | 正式生成不能调用此writer，否则会替换TEST current |
| [报告模型](../Go-based%20Refactored%20System/internal/model/management_traits_report.go#L7-L47)、[实际DDL文本](../scripts/sql/management_traits_001_runtime.sql) | revision唯一(run_id,revision)，current主键paper_id；test_title/test_label非空列 | 增加mode不能产生两个独立current；不ALTER这些已运行表，不改001 |
| [鉴权](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime.go#L169-L192) | 真正允许userId=1或*:*:*；传入permission字符串未用于细粒度授权 | 不宣称已支持审批角色权限；正式审批要校验真实登录主体及经用户确定的负责人授权 |
| [令牌](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_security.go#L13-L16) | purpose是management_traits_participant / management_traits_paper | management_traits_test不是这里已存在的令牌purpose；报告用途与认证purpose必须分开 |
| [旧链守卫](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L260-L376) | 枚举所有el_mng_表，未知无身份列extra失败关闭，canonical审计间接闭包特判 | 新版本登记/审批表必须先接独立完整结构门禁及守卫分类，不能“新建表就安全”或删掉未知拒绝 |

数据库侧现有11表不包含正式审批登记模型。旧设计中的12类职责是提案，不可冒充当前已实现的正式内容表。当前TEST标题为“管理特质 TEST 测试报告”，标签为“仅供系统测试，不可作为人才决策依据”；正式功能不得覆盖这两个常量或原资产。

## 3. 用户故事与可验收要求

### US1（P0）：版本提交、批准、显式启用

管理员登记精确正式候选资产；内容负责人和心理测量负责人审阅同一版本；批准后管理员另行启用，上传或单次审批均不得自动启用。

Given缺任何批准、资产改变或未启用，When尝试正式生成，Then拒绝且无正式PDF/revision/current写入。Given全部门禁满足且使用明确版本，When正常授权生成，Then记录批准证据与确切资产身份，不读取最新活动稿偷偷替换。

### US2（P0）：正式/TEST互不覆盖

管理员从人员结果行选择报告用途并查看版本历史，不手填reportId。Given已存在TEST报告，When生成正式报告，ThenTEST报告字节和current不变；正式失败时也不回退TEST或传统PDF。

### US3（P1）：管理端及参与者端明确状态

管理端显示人员、完成状态、报告用途、版本、可操作状态与拒绝原因；参与者保留可信25分钟/20分钟提示、保存/恢复/交卷行为，技术ID移入管理员审计详情。正式标识只能依据后端可信冻结配置，不依据URL或本机storage。

### 功能要求（本功能内唯一REQ编号）

| REQ | 合同 | 验收结果 |
|---|---|---|
| REQ-001 | 00201/00202新建新版；历史/00401/MBTI不变 | 历史结果、旧PDF、旧模板/源题零覆盖，无自动迁移 |
| REQ-002 | 正式与TEST报告用途、资产、DTO、接口、存储current独立 | 两种报告可独立查找，不因另一种生成改变当前选择 |
| REQ-003 | 正式版本不可变身份绑定双批准与环境；启用单独授权动作 | 空批准、错SHA、错受众、错环境、未启用均零正式产物 |
| REQ-004 | 审批主体来自正常认证；不seed批准、不填假人/假日期 | 请求中的任意批准人名不能替代登录用户及授权证据 |
| REQ-005 | 复用合格不可变run、13维/4模块/receipt和真实来源，不重新计分 | 换内容/模板只新增报告revision，run/input/答案/期限不变 |
| REQ-006 | 正式模板和内容独立登记并严格校验 | 205规则精确键齐全、固定说明纳入审阅、六图/五数字槽/长文本及零分灰环通过；无危险外链/宏/关系逃逸 |
| REQ-007 | 生成末次锁内复核批准/启用代际及run事实，失败保旧 | 审批撤销/版本切换/并发/转换失败不推进正式current，不留下未登记文件 |
| REQ-008 | 管理端完整版本/审批/启用/报告行操作闭环 | 操作后刷新，取消零请求，失效审批明确原因；不要求手工输入UUID |
| REQ-009 | 参与者状态来自后端，保持现有采集、时限、恢复及匿名精确路径边界 | TEST警示保留；正式条件不满足时不显示“正式已启用”，不泄露评分方向/内部版本/密码 |
| REQ-010 | 下载只用明确正式reportId及同一已登记字节 | 鉴权、资源归属、批准撤销策略、同句柄size/SHA、PDF MIME校验；无路径输入/旧链回退 |
| REQ-011 | production保持关闭，本地验证与部署授权分离 | 任意本地批准不使production可执行；本轮远端/发版操作0 |
| REQ-012 | 分支先登记、代码与测试逐切片闭环 | slice1真实handler/service测试及编译证据；组件及整链测试在后续切片 |

## 4. 关键实体与推荐最小隔离方案（设计，不是DDL批准）

1. **正式版本登记**：不可变revision ID，产品/题本/评分/常模四轴与scoringManifestSHA、mapping适用性、受众/两code覆盖范围；contentVersion/contentSHA/sourceWorkbookSHA、templateSchema/templateRevision/templateSHA/bindingManifestSHA、固定文字审阅清单及渲染器合同身份。仅审批/启用授权元数据可追加事件，正文/hash不可原地更新。
2. **双批准记录**：内容/心理测量两种职责分别明确动作，绑定同一完整版本身份摘要及适用环境、真实签署/登记主体、服务器事件时间；外部签署方式另见Q1。负责人曾确认为Liming，但2026-10-01只用于候选测试，不复制为正式批准时间。
3. **启用选择**：按环境+00201/00202分别选择已批准版本，带单调授权epoch；approved与active分离。缺版本/未批准/撤销时不可启用。批准两个code须各有真实匹配的四轴/题本SHA，不按leader名称推断适用性。
4. **正式报告revision/current/audit**：新增独立正式专表职责，引用已有run的(id,paper_id,exam_id)复合候选键；正式current只引用同paper的正式revision，TEST继续现表。正式revision冻结版本/批准证据摘要、DTO/DataSHA、文件SHA/大小、真实actor/时间；审计有明确资源身份。
5. **用途与资格显示**：Q2已选择既有合格新版TEST run可追加独立正式报告，无需以“新正式测评绑定”排除该政策。新建默认新版仍不等于默认正式；无approved+active版本及完整run资格时只提供TEST验证，正式选项禁用且说明原因。保存和冻结保持两动作，不回填历史profile/snapshot。

推荐新增专属正式表而不是向TEST表插入formal：满足两个current、保留test_title/test_label及现有11表签名，不影响已运行TEST下载。复用评分bundle/run及其可信加载与精确校验**不等于复用TEST内容类型/DTO/writer**。是否允许已存在TEST run生成正式报告由Q2确定，技术可复用不能代替业务资格。

现有守卫会扫描全部el_mng_表：版本/审批/活动选择这类非paper实体须明确列入已知正式配置表白名单，并严格校验完整正式schema、关系和行语义；不可给其伪造paper_id避开未知extra。正式报告/audit以真实paper/exam/run复合身份接闭包，未知表仍fail-closed。完整新增结构缺失只关闭正式功能；部分安装/孤儿应失败关闭，不放宽TEST既有11表门禁，不使无关传统读取被新登记表意外拒绝。

slice1已定三配置表el_mng_formal_version/approval/audit及独立C区DDL；CRUDE/列类型索引/FK见本地证据。正式报告实体/current的DDL后续单独设计，不ALTER旧表/001、不AutoMigrate、不seed批准。运行模板以后归B区运行配置，原件/候选审阅归C区docs。迁移真实执行仍须另行授权与备份，不借DDL事务声称可回滚（MySQL DDL自动提交）。

## 5. 正式生成与版本状态合同

版本流：登记草稿→结构校验及候选预览→两职责批准→显式启用。修改资产创建新revision且双批准重新开始；退役/撤销保留历史事件和文件，不原地删除。正式内容来源仍是用户确定的Excel等级列、Word固定样式；创新性频率/错字、统计声明、V67/V96及星形/页脚采用最终批准资产，不由AI全局替换。暂不新增年龄/用时/总体常模位置、模块等级/比较、百分位、团体或自动报告。

审批绑定四评分轴/scoringManifestSHA、两类呈现SHA、绑定manifest、受众及环境的完整摘要；文件上传前后SHA、规范规则内容SHA与原XLSX SHA分别记录，不能混称。正式模板从权威客户候选形成独立draft，不从活动TEST包删两个警示就冒正式批准。未批准的预览只能是明确“候选草稿”的合成预览，不给真实参与者出正式报告。

推荐生成链：认证/资源资格→显式run+formalVersionRevision捕获→真实来源/完成/140/13/4/receipt复核→同一冻结版本构造正式DTO→事务外Word/LO→UUID独占私有文件→末次事务锁paper、版本授权记录及正式current，重核身份/hash/环境/批准/epoch→原子revision+current+audit。所有锁按固定顺序；新增有界生成deadline涵盖排队/转换/持久化，取消不切current，不改原业务时钟或25分钟期限。

推荐current初期使用同卷串行生成及DB条件代际提交，不新增高级手工历史current选择API：版本切换/撤销推进授权epoch，在途旧epoch失败不推进current；已完成PDF按明确reportId读，历史可列出但不是悄悄换成当前版本。并发不同版本/同版本生成要有真实双连接证据，不只依赖单进程mutex。正式缺任何门禁时给受控错误，不返回partial成功，不回退TEST。

正式环境配置与MNG_TEST_*完全独立；本阶段仅local，未来staging需明确部署授权；production默认硬关闭，批准记录的environment不是运行环境配置也不是上线授权。确切新配置键及所有调用点在实现切片列C4清单后再定义。

## 6. 接口及两端UI合同（全部拟新增，不是当前路由）

正式版本登记/审批/启用/撤销、版本状态读取已采用五个独立管理路由；预览、报告按run列表、正式生成/查看/下载尚待实现。均正常JWT与操作/资源授权，不加入匿名prefix、不改已有generate-test/view/download含义；登记请求不接approved=true或客户端批准actor。Q1已采用同一授权系统账号分别双签，具体当前管理员边界见§8。

报告列表响应建议含reportId/runId/paperId/reportKind/versionRevision/status/createdAt/canView/canDownload/blockedReason，不含文件路径；权限在服务端每次复核。正式生成请求固定runId+versionRevisionId+expectedActivationEpoch；下载固定reportId。版本状态应返回受控可操作能力与原因，不能只靠布尔“已批准”遮掉题本不匹配。字段及路径在实现前再确认消费者C2，不改旧结果模型JSON形状。

管理端：既有模板管理页增加002版本卡片及登记/审核/启用明确分组，不共享00401上传门禁；结果页以姓名/完成度/完成时间/用时/报告状态为主，技术ID进入审计详情。正式、TEST分区及各自版本历史；正式不可用时显示具体门禁，TEST只能显式测试操作。取消确认零写入；成功刷新，失败保留表单；页面最多一个主CTA、行操作≤3，其余更多，键盘可达、状态不只靠颜色。

参与者：登记→登录→准备→答题→完成共用可信用途状态；TEST场景保留警示。正式场景只在后端冻结正式资格成立时采用客户批准指导语，不以mngTest缺失当正式。原匿名认证purpose、五键身份/资源绑定、requiredFields子集、空串、140五选一、即时保存、第一未答、原deadline、20分钟提示、手动缺题拒绝、到期incomplete NULL、管理员≤5分钟续答规则保持，不扩匿名恢复或重置期限。参与者报告下载不在本请求明确授权范围，继续关闭，不擅自打开showPdf。

目标验收：390/768/1440视口不横向溢出（表格允许受控滚动）；触控≥44px、字体≥12px；一主CTA，Element UI图标一套，桌面结果默认20行、数字右齐。当前源码结果页7列、20行、操作180px、手填报告ID是待优化事实；屏幕利用率/颜色对比/新UI截图未测，不填估计PASS。

## 7. CRUDE及消费方影响清单（未来切片，不是本輪改动）

不更改既有三TEST表结构或字段含义；对其当前实际入口已核：Create=GenerateTestReport/persistManagementTraitsTestReport；Read=ReadTestReportPDF及保护闭包；Update=TEST current upsert；Delete=旧链被guard阻止，不提供业务报告删除；List=无报告列表；Detail=按ID读取；Import/Export/Login=没有报告表导入/导出/login writer，但candidate/tester身份读取通过guard间接查询audit→revision→run。新增正式表的全部同类入口需独立清单，不能依此表推定已完整实现。

| 已核消费方/入口 | 后续处理 | 理由 |
|---|---|---|
| [TEST service](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go)、[model](../Go-based%20Refactored%20System/internal/model/management_traits_report.go)、[HTTP](../Go-based%20Refactored%20System/internal/handler/management_traits_report_runtime.go)、[TEST Word](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word.go) | 无需修改TEST公开签名/语义 | 正式独立路径；TEST SHA/标题/purpose恒定；相关pipeline/download/env/source-lock回归必须保留 |
| [schema](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go)、[legacy guard](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go) | 同步修改独立正式门禁/已知表分类及闭包 | 新extra会被枚举；不能忽略任意表或破坏canonical audit五列 |
| [runtime HTTP及注册](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime.go)、[router](../Go-based%20Refactored%20System/internal/router/router.go) | 同步注册新管理接口 | 正式认证不能由旧候选admin函数接受任意permission字符串来冒充审批授权 |
| [JS API](../Go-based%20Refactored%20System/ruoyi-ui/src/api/managementTraits.js)、[结果页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/managementTraitsResults.vue) | 同步新增版本/报告列表/正式封装与行操作 | 三TEST报告函数的实际消费方为结果页；旧函数不换含义，不手填report ID |
| [配置页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue)、[模板页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/template/index.vue) | 同步用途/版本资格及审批卡片 | 正式未批准禁用，不自动修改历史profile；00401与MBTI既有块不变 |
| [candidate](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue)、[tester](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/tester.vue)、[准备](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/preview.vue)、[答题](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/managementTraitsExam.vue) | 同步可信用途显示，保持答题流程 | 当前警示/intent/storage固定TEST；不能只批量删除警示 |
| [测评列表](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/index.vue)、[人员详情入口](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue)、[首页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/index.vue)、[旧result2](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue) | 同步仅新版用途探测消费者，旧保护保留 | 真实调用managementTraitsExamKnown/Intent；新正式入口不能落旧截图PDF回写 |
| [结果读取](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_results.go)、[令牌](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_security.go)及评分/run事实 | 无需修改旧返回结构/认证purpose/评分 | 新UI显示数据用专属投影，新增资格状态不覆盖既有immutable事实 |
| [001 SQL](../scripts/sql/management_traits_001_runtime.sql)及现有11表 | 无需修改 | 新正式DDL未来独立脚本；旧索引与签名保持 |
| 旧PDF/旧结果、00401/MBTI渲染/导出、既有harness | 无需修改 | 本次不迁移、重算、生成旧PDF或继续harness；兼容验收不等于本轮重跑 |

## 8. 已确认业务政策（原三项选择已关闭）

**Q1 已确认A**：同一授权系统账号允许兼任，必须分别签署content/psychometrics两条独立审批记录。actor取真实loginUser.UserID，时间取服务器；不索密码、不预填历史日期、不seed批准。slice1授权边界沿现有管理特质管理员口径，仅正UserID=1或正UserID且*:*:*；普通exam:list不授签署权。未实现外部签署或非管理员细粒度负责人授权，不能声称已支持。此前Liming/2026-10-01仍仅候选测试证据。

**Q2 已确认B**：既有new_creation/submitted_snapshot/completed且140/140新版TEST run，在其精确题本/评分/常模及mapping明确获批时可追加独立formal结果/PDF，不重算、不转换原TEST用途、不写TEST current、不覆TEST文件；legacy历史不纳入。审批绑定source bundle/manifest/mapping/版本与呈现资产，不绑定某个受测者或profile身份字段。完整可信run/13维4模块/receipt及报告生成资格验证在后续报告切片实施；不能用本slice的versionReady替代run验证。

**Q3 已确认B**：撤销为不可重新启用终态，禁止newgenerate/activate；保留审批/audit/原文件。授权管理员可继续view/download明确reportId的历史原formalPDF，UI显著revoked，字节不重写。slice1只实现version撤销和历史管理员可读的政策投影，尚未注册formalPDF读接口；不宣称下载已经实现。退役与来源可信性撤销不合并。

正式内容尚未定稿不是让AI代选或再问几十项文案：现有Excel/Word可登记为draft，未决错字/频率/统计依据列为版本审阅阻断；最终客户材料及真实双批准才能关闭。登记上传能力可在后续实现，但本阶段不生成正式资产或批准记录。

## 9. 后续切片顺序及成功标准（不是可执行plan）

Q1～Q3已确认；本地版本登记/双签/显式启用/撤销、独立schema/守卫分类和五管理API已在slice1实施。下一由主协调者只读复审后继续独立正式资产接收/清理/预览与完整模板验收、正式DTO/Word/持久化/ID读取→管理端版本与报告行UI→参与者用途状态UI→本地合成与真实运行验收。每次先RED/矩阵、后代码/GREEN、编译/真实调用、回写后停止等待；不一次性做完，不调用现代化任务状态工具。

REQ映射：slice1覆盖003/004/011的版本管理边界及012；schema分类涉及001/002但报告独立current尚未实现；正式报告链005/006/007/010、管理UI008、参与者UI009仍未完成。53个新pass事件是本地测试，不是系统覆盖率或真实MySQL证据。

可测成功标准：未批准正式生成产物0；正式/TEST互相覆盖0；未授权审批成功0；两code×开放/封闭四流程通过且历史资产变更0；完整报告恰13维/4模块/6图、所需批准原文逐字一致；0/3答到期无正式报告；三个视口及键盘流程通过；正式失败无current推进/未登记文件；production执行0。本轮均为未来验收条件，不是已通过结果，不宣称新覆盖率或完整产品DONE。

## 10. 本阶段记录

上阶段仅写三份C区草案、未编译或改业务的事实保留为历史；本次slice1新增Go模型/服务/HTTP/测试，新增C区独立SQL并注册router/守卫分类。全Go6229pass/0fail/9skip（673顶层），server/all-build/vet各exit0且输出0，未DB/远端/发版/浏览器，不重跑harness。独立CodeReviewer本模式禁止派发，尚未独立复审；正式PDF/两端UI/真实schema及并发验证仍未完成，禁止部署。