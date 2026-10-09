# 管理特质005产品隔离 — 2026-10-08 本地切片

## 未验证与边界（先列）

- **仅本地代码/合成测试完成，未部署。** 未查数据库005占用、未注册真实005题库、未复制题目、未执行DDL/DML/批量UPDATE、未访问SSH/服务器；不能称005真实业务创建或正式报告已可用。
- 指定现场002测评1791298091700970647及58.642639／58.64未访问、重算或迁移；本轮没有该记录fresh数据库/hash证据。兼容由原冻结合同回归、旧核心源码SHA与合成canonical黄金SHA验证，不用合成结果冒充真人实证。
- 独立CodeReviewer尚未执行，本模式不派发其他agent；正式内容/心理测量批准、production报告、实际MySQL/竞争/安装/来源审核均未扩大。
- 清理仍为**延期保留**：没有确认可安全删除项，不删除/移动/归档历史文件、三处恢复bin、客户原件、旧PDF或失败收据。正常编译重建本地dist/覆盖本地server编译输出；既有测试按自身临时工件生命周期清理，不等于历史垃圾清理。
- 首次浏览器及JSON解析失败保留，详见下文；不将构建成功等同全产品验收。旧002混合发布计划继续挂起。

## 实施与影响清单（公开签名/响应结构保持）

本轮21个既有Go/Vue/JS文件变化、4个新B区文件；测试/验证入口位于相应单元测试目录与C区scripts/test。只实现同一产品隔离切片，不注册数据、不全局替换002。

| 文件/位置 | 同步修改与行为 |
|---|---|
| [后端集中产品分类](../Go-based%20Refactored%20System/internal/service/management_traits_product.go#L4) | 精确LEGACY002/NEW005/OTHER；code只代表配置身份，不代表冻结授权 |
| [版本白名单](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_security.go#L102-L128) | 00501→staff/mng-00501-v1/mng-00501-db-current-v1；00502→leader/mng-00502-v1/mng-00502-db-current-v1；原002四版本不变，不能混受众或只换码沿用hash |
| [草稿入口](../Go-based%20Refactored%20System/internal/service/management_traits_draft.go#L124-L192) | 真实DB repo code决定005必需草稿；普通002缺flag/false原新建，显式true拒并提示005；既有002draft可编辑同系列、不跨005/退旧；缺表仅关闭005新建 |
| [公开详情](../Go-based%20Refactored%20System/internal/handler/exam.go#L328-L378) | 005复用真实生命周期/profile投影，无标记005拒绝而非legacyfalse；普通002明确false，已有冻结002保原字段/profile |
| [冻结事务](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_profile.go#L69-L110) | 原冻结profile首先只读复用；新冻结必须持久draft，普通002不能隐性升级；原002draft兼容；实际repo ID查询140/700源，保原子marker更新 |
| [正式登记结构门禁](../Go-based%20Refactored%20System/internal/service/management_traits_formal_schema.go#L170) | 仅精确追加005两码；不复制批准、不启用正式PDF，sourceMatch沿真实版本/受众/SHA核验 |
| [客户端集中分类](../Go-based%20Refactored%20System/ruoyi-ui/src/utils/managementTraitsProduct.js#L3) | NEW005为产品意图；同exam+服务器stricttrue才新冻结/旧FROZEN_COMPAT002；合法draft单列；005false/unknown关闭，旧bool-only002兼容，普通多002旧配置保留 |
| [配置表单](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L133-L149)及[默认守卫](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L398-L402) | 普通002显示旧版、隐藏新建opt-in、不改25分钟默认/旧字段；005新版默认锁定，基层/干部按实际码；保两步确认/取消保draft；历史→005拒绝；005未知重读只读关闭 |
| [列表版本标签](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/index.vue#L217-L225) | 按服务器005码显示基层员工新版/干部新版，不伪造repo IDs或客户端题库选项 |
| [开放登记](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L246-L270)、[配置复核](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L354-L369) | 005无冻结或unknown不旧保存；真实冻结字段子集；兼容旧002严格bool-only，存在新标记须一致；迟到/同ID屏障保留 |
| [封闭登录](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/tester.vue#L58-L107)、[准备门禁](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/preview.vue#L110-L122) | 配置pending禁登录；005合法冻结走新登录；无冻结005不旧开卷；旧002明确false原路径 |
| [管理员旧URL分流](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L326-L356)、[结果资格](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/managementTraitsResults.vue#L123-L140) | 005和冻结兼容002复用原专属结果页；同exam/精确code/真实管理员/profile门禁，URL/query/session不赋权限；unknown拒旧，00401/MBTI原分类 |

无需修改（理由）：原002标准分/导出/双PDF、评分13维4模块/反向一次/精确常模705/13、canonical JSON schema/hash domain、decode/runtime source/load/create/result validation、审计历史重出、报告存储及客户Word/LO渲染器均技术复用；没有改公开函数签名或新增响应字段。旧历史audited-reissue仍仅002，不将原历史源宣称005来源。原shared四个runtime版本消费者profile/load/result/formalSourceMatch同步通过精确005版本白名单；它们自身无需改（原参数/字节校验与来源门禁保持）。原API wrapper/token/router/public路径无需改（共享端点，服务器资格仍逐请求验证）。

## 来源与旧002保护

- 开工原生扫描507个internal/前端src/unit Go/Vue/JS文件，005出现0；仅本地定义，无数据库空闲结论。
- 005合成fixture使用独立repo及题项/选项ID，实际public FreezeProfile读取该repo的140题/700项；非硬编码原002 repo ID。缺题/源不合法仍拒绝。真实source SQL、题干及规范化核心字节未改。
- 合成00201原manifest SHA＝639f42caf091d6572b4e4b4df61af18d3295a4dc19b812d0e23f6cfcfe64619f，mapping SHA＝f7b97a615d0134706285f3ef65edf0c4eb8a05993d151390ea5e8ebec2b6b3f6；00202为1e7128bf52fce6cdad541a653965b1149814afd7b76a820e7622d25a7591e88e／569a16b609ddf0595cf3403f02f11500620a255f71c294b1994a079689a9c936。修改前实采、修改后黄金断言通过；不是现场真人hash。
- 原canonical/scoring/identity/source/load/create/result-validation/历史audited-reissue/reissue/report服务及相关handler共12个保护源码逐SHA同开工基线；原507路径缺失0。当前实现变化仅必要产品版本/草稿/元数据及UI分流。原002源文件未全局换005，客户模板和205原文未修改。

## TDD与实际验证

| 验证 | 真实结果 |
|---|---|
| 实现前Go来源RED | 两005子项拒绝，顶层FAIL，native1；同时实采上述002黄金SHA |
| 实现前实际Save RED | 005五子项失败/一个mixed拒绝通过；普通002四场景失败/两显式true拒绝通过，native1 |
| 实现前实际candidate SFC RED | 未冻结005错误解锁旧身份两fail，冻结005两pass，native1 |
| 管理特质专项 | Go3230 PASS事件/0fail/6环境skip/native0；前端8files418/0/0/native0（随后另加3项测试） |
| 全Go | [6400 PASS事件/0fail/11skip/parse0/native0](../scripts/test/results/mng-005-isolation-local-20261008/go-all-1791452162855.json)；含子项，非覆盖率 |
| 追加最终Go | [35/0/0/native0](../scripts/test/results/mng-005-isolation-local-20261008/go-final-1791452484857.json)，包括最后扩展的002draft跨005拒绝、005实际冻结及原冻结只读复用；不把它改写成另一次全Go6402 |
| Go构建/vet | [server build0](../scripts/test/results/mng-005-isolation-local-20261008/go-build-1791452168296.json)、[all build0](../scripts/test/results/mng-005-isolation-local-20261008/go-build-all-1791452177795.json)、[vet0](../scripts/test/results/mng-005-isolation-local-20261008/go-vet-1791452186087.json)，stdout/stderr均0bytes |
| 最终前端全量 | [34files586pass/0fail/0pending/native0](../scripts/test/results/mng-005-isolation-local-20261008/front-all-1791452508901.json)；旧560基线未删保护，冲突新建政策测试转为005，不修测试绕绿色 |
| 最终production构建 | [vue-cli-service build native0／Build complete](../scripts/test/results/mng-005-isolation-local-20261008/front-build-1791452534106.json)；Browserslist提示保留，不升级依赖或扩大配置范围 |
| 当前编译UI浏览器 | [28/28 PASS，pageErrors0，非本机/未授权请求0，contexts/server closed](../scripts/test/results/mng-005-isolation-local-20261008/browser-1791452534517/summary.json)：1440/390×普通002两码/冻结002兼容/冻结005两码/无冻结005两码×candidate/admin。所有API合成，无真实身份、报告生成或数据库验收 |

11个环境skip完整名称见全Go收据；实库、真实LO等opt-in仅在子进程清除36个测试环境键，不修改用户环境。既有环境未验不能列为PASS。

失败不隐藏：首次新分类模块不存在是编译RED；首次读取错误的service草稿测试路径失败，后核对实际handler测试；初版补丁块次序失败后原补丁纠正。一次登录SFC多余括号诊断修正并复核无错误。首次Vitest JSON定位未兼容空白，exit1/解析失败收据保留；更正解析后408pass/6fail明确为3旧新建政策fixture＋3正向fixture缺真实repoCode，调整后负向保护保留且全量GREEN。两次浏览器失败收据为组件未挂载时测试predicate读取null.__vue__；第二轮实采stack后仅增加DOM就绪守卫，28场景及所有请求/权限断言未弱化，最终新编译bundle通过。没有覆写旧失败或复制测试绕过预算。

初次文档终验native1为两个新helper行范围超界，已改为真实定义行；初版安全汇总对go-build文件名前缀误选go-build-all收据，不影响上述独立server-build真实退出0，最终汇总以精确阶段名匹配。原失败汇总保留、不覆盖。

## 下一任务（不在本轮执行）

1. 主协调者安排CodeReviewer独立只读复审本切片及最终SHA。
2. 另行确认真实005占用/独立题库注册及来源审核范围，再制定备份和独立ID复制策略；保持原002 repo/源题/成绩/PDF/hash。没有005实际条目时真实选择器不出现虚假新版选项。
3. 明确005 source/内容/模板版本及审核证据后再执行受控真实环境验证；本轮TEST技术复用不等于正式内容批准。
4. 发布须重新确认005及现场冻结002兼容的独立scope，不沿旧002混合计划部署；历史垃圾仍待精确所有权/引用/保留核验。