# 生产数据库变更审查与 staging 对比（2026-10-08）

## 1. 结论：只读比对完成，迁移发布仍 NO-GO

**尚未执行：** MySQL5.7恢复副本迁移首轮/重跑、DDL锁耗时与故障恢复、实际应用账号新Schema运行门禁、写入/并发/删除回归。staging MySQL8现状一致不替代5.7迁移演练。本次不修改业务源码或迁移脚本，不复制数据、不升级数据库、不部署/restart、不执行DDL/DML或恢复/清理。

主要阻断：

1. 当前管理特质窄引用兼容例外仅认可0900排序规则；生产为general_ci，建新表后仍有运行时拒绝风险，须先独立兼容修复/审阅或明确列迁移方案，不绕过门禁。
2. 生产旧D题库及历史已实际使用，不能套staging身份重置。已有454道胜任力源题、9个配置、19份结果、3份报告；所有历史数据/原PDF保留。
3. 生产缺14个共享表新增列及24张staging表；不能直接整包替换。发布范围仍未冻结，最小002支路与00401全产品升级不是同一迁移。
4. 两份旧到期胜任力卷、199份缺测评的旧卷处理策略及当前恢复点尚未关闭。

## 2. 实际采集与证据

07:05Z起始结构、07:08Z数据前提、07:10:54Z结束结构，均采用10秒SELECT预算、READ ONLY事务，未取人员/答案/报告正文。默认值仅SHA，密码/DSN仅远端内存匿名管道消费。

- Production：39.106.61.48，已授权root/Posh-SSH严格信任检查；SQL使用现有生产配置凭据。
- Staging：20.200.136.133，既有liming/key/Strict/Batch/ConnectTimeout10/Attempts1；sudo socket仅作metadata/聚合查询，不冒应用validator验收。
- 两轮结构及聚合完全相同；表大小未进入规范结构签名，原始TSV本轮也逐SHA相同。
- 本次不是备份，不保存真实业务行或secret；证据属C区ignored运行产物。

| 项目 | Production | Staging |
|---|---|---|
| MySQL | 5.7.44-log | 8.0.46-0ubuntu0.24.04.4 |
| Schema | element | element |
| 表/列 | 54 / 614 | 78 / 909 |
| 索引列元数据行数 | 188 | 334 |
| FK约束数/关联列行数 | 20 / 30 | 49 / 68 |
| 默认排序规则 | utf8mb4_general_ci | utf8mb4_0900_ai_ci |
| 测评legacy/competency | 57 / 9 | 61 / 10 |
| 维度D/A-B | 48 / 0 | 0 / 10 |
| 胜任力源题 | 454 | 90 |
| 胜任力结果/报告 | 19 / 3 | 24 / 28 |
| state0到期/未到期 | 517 / 0 | 301 / 0 |
| state0缺测评关联 | 199 | 0 |

原始结构SHA：production `834378935768917bb2f50f999ce859bc6789d02015cc34c1d1d8d2892bcca461`，staging `afd19a7020c8772d2cc25783fc32e7a6af5c126a83c388e5492f78ec6720c7bd`。

证据：[结构差异JSON](../scripts/test/results/db-schema-compare-20261008-070518/comparison.json)、[production原始metadata](../scripts/test/results/db-schema-compare-20261008-070518/production.tsv)、[staging原始metadata](../scripts/test/results/db-schema-compare-20261008-070518/staging.tsv)、[production数据前提](../scripts/test/results/db-schema-compare-20261008-070839/production.tsv)、[staging数据前提](../scripts/test/results/db-schema-compare-20261008-070839/staging.tsv)。

## 3. 逐类结构变更

### 3.1 共享表新增14列

- el_exam：competency_product_version、competency_scoring_version、competency_content_version、competency_report_template_version（4）。
- el_competency_result：product_version、content_version、report_template_version、dimension_question_count、answered_dimension_question_count（5）。
- el_competency_report：template_version、result_run_id（2）。
- el_exam_competency_dimension.group_id、el_exam_competency_question.competency_question_type、el_qu.competency_question_type（3）。

对应结构主体为007的8列、008的5列、013的1列。生产overall_score为NOT NULL，staging为NULL可表达；这是未完成答卷语义所需变化，不把0当NULL回填。字段顺序及整数显示宽度变化不等业务语义变化。

### 3.2 Staging新增24表

- 一期组/效度：el_exam_competency_group、el_competency_group_result、el_competency_validity_result（008）。
- el_competency_migration（009数据重置marker，**不纳入production通用补表**）。
- el_competency_report_content_package（010，空结构不代表正式批准）。
- el_competency_result_run及overall/module/dimension/validity五子体系（011）。
- el_competency_version_dimension、el_competency_dimension_mapping（012）。
- el_competency_report_current（013）。
- 管理特质001的11表：definition_bundle/exam_profile/paper_snapshot/paper_question_snapshot/result_run/result_dimension/result_module/runtime_receipt/report_revision/report_current/report_audit，均有el_mng_前缀。

Production没有独有表。formal registry002三表、draft003一表、reissue004两表在采集时两环境均不存在；这些是未安装独立工件，不能把本地文件存在或另一路staging发布批准当成本次真实迁移PASS。

### 3.3 索引及外键

- 生产保留且staging没有的四个优化索引：el_candidate.idx_exam/idx_paper、el_qu.idx_qu_type_level_id、el_qu_repo.idx_repo_qu_type_qu_id。**不删除、不为了“相同”整表重建**。
- 同名uk_qu_dimension_item：生产(dimension_id,dimension_item_no)，staging增加competency_question_type。差异工具以索引列逐行显示，不把旧第二列消失当整个索引应删除。
- 报告唯一键新增template_version（007）及run版本/跨卷复合候选键（013）；必须先建替代索引保护旧paper外键，再切旧索引，不在缺保护窗口写入。
- 共用表的FK差异：新增group及report→run复合引用；Quartz旧FK的RESTRICT/NO ACTION属于版本表述差异，不是本业务迁移理由、不调整。
- 新FK必须继承production父列的字符集/collation。禁止从MySQL8 dump复制0900到MySQL5.7，也不以全库排序规则转换解决局部升级。

## 4. 关键兼容与数据前提

### 4.1 管理特质父DDL通过条件不等runtime可用

真实生产el_exam/el_paper/el_paper_qu.id均为NOT NULL varchar64/utf8mb4/general_ci，001父键条件计数3；源题/选项id varchar64，旧el_paper_qu_answer.qu_id及answer_id仅varchar32，但全部相关源ID均ASCII且长度<=32。

然而[窄引用兼容判断](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L196-L203)要求父collation精确0900；[容量拒绝分支](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L254-L269)在例外不满足时拒绝。Production的general_ci不满足该例外，不能承诺仅安装001后可用。

原生执行既有 `TestBugMTSchemaLegacy002Drift/unverified_base/utf8mb4_general_ci`：native0，2pass事件（父测试+子测试）、0fail/0skip、JSON解析0、stderr0。测试证明既有拒绝合同，不是远端新Schema服务验收；其夹具含staging旧repo编码，不能称完整production实例已复演。编辑器未发现测试，已用原生精准子测试替代，不把“未发现”记GREEN。

下一必须先明确仅安全拓展生产general_ci窄边的代码方案或经批准的数据保留列扩容方案，再RED/GREEN+实际5.7独立恢复库验收；当前不改业务代码、旧列或守卫。

### 4.2 防止误套主数据与缩窄

- Production454源题、48D身份、9配置、19结果、3报告与stage不同是业务事实，不是数据应同步。旧历史改新版需独立规则，不按名称猜映射，不恢复stage dump。
- [当前002种子](../scripts/sql/competency_002_dimensions.sql#L67-L82)已改A/B十维；唯一name/order冲突可使no-op仍保留D身份，前面的重复ALTER也不是无风险补跑。已安装001～006不可当原7月不变工件无条件重放。
- [009](../scripts/sql/competency_009_phase1_identity_reset.sql#L1-L20)明确staging-only且删除源题/文案/维度；生产依赖非空，也未授权重置，**排除**。
- [014开头DELETE](../scripts/sql/competency_014_v2_report_content_draft.sql#L1-L4)无PK条件、会重建draft内容，不是纯结构迁移；需单独生产批准、幂等/审计和SafeUpdate方案，当前不执行、不修脚本。
- [015](../scripts/sql/competency_015_v2_report_content_staging_approval.sql#L1-L22)绑定staging与历史具名SHA，不能转抄为production批准；再次执行原approved对象也并非no-op成功。
- Production tester.mbti_type varchar8→stage varchar4、mbti_scores TEXT→varchar200，当前超过新长度的行数均0。该单时点不授权缩窄、也不证明未来安全；与本升级无关，保持生产原容量。
- Production题号旧唯一键重复0、报告paper/content重复0；结果/报告缺paper、发布快照缺exam聚合均0。stage旧dimension/item双列重复8符合新索引按题型区分，不当题本错误；不在stage降回旧双列索引。

## 5. 迁移适用性与演练建议（未执行）

| 脚本 | 若发布对应功能时的候选处理 |
|---|---|
| competency007 | 新增版本及受限旧版本回填；预计涉及现9配置/19结果/3报告，先保存SHA/NULL语义，实库SafeUpdate与首次/重跑待演练 |
| competency008 | 五列/三表/题型回填/索引切换/NULL语义；多DDL及临时DROP自身FK后恢复，失败不事务回滚，需维护窗口和部分状态恢复测试 |
| competency009 | 禁止加入production序列；身份并行保留方案另定 |
| competency010 | 结构可独立审阅；无批准行不开放正式报告 |
| competency011→012→013 | 并行结果/稳定目录/报告绑定；按依赖序演练，旧报告result_run_id保持NULL、不自动重算；重复011/008有DROP-FK/ALTER，不能凭IF NOT EXISTS称无写no-op |
| competency014/015 | 不是结构补齐，draft写入及staging批准均隔离，生产精确内容批准另定 |
| management_traits001 | 新11表/15FK；生产父collation可继承，但runtime窄边阻断先解决。空新表不为历史旧卷补造快照 |
| management_traits002/003/004 | 分别formal/draft/reissue，可选独立scope；004依赖001的run复合候选键，不能只建2表而省其依赖；两环境未安装不冒已验 |

生产el_paper_qu的数据+索引约87.44MiB，答案桶表约189.36MiB（information_schema分配量非精确磁盘归因）；这轮007/008不是必须ALTER全部答案桶。只允许批准的最小列/索引集合，别用整库diff同步所有字符集、MBTI字段或Quartz表。

必要真实演练应在**独立MySQL5.7生产当前备份恢复副本**：首轮/重跑及FK列签名、回填前后旧成绩/原始答案/PDF指针SHA、故障中途再执行、旧/新二进制读兼容、并发/回滚、元数据缓存重启、旧到期Worker策略。staging8可补充功能回归但不能替5.7语法/锁证据。创建恢复库、恢复业务数据和最终DROP都需明确授权，当前未执行。

## 6. 本轮变更及验证

- 新C区[只读结构SQL](../scripts/db/production-stage-schema-readonly-20261008.sql)、[聚合前提SQL](../scripts/db/production-stage-data-preconditions-20261008.sql)、[安全采集](../scripts/tools/production-stage-schema-collect-20261008.ps1)、[本地比对](../scripts/tools/production-stage-schema-compare-20261008.js)。采集三批双端真实exit0；最终PS解析0、Node语法/比对/完整性断言native0、文件diagnostics0。无Go业务改动，不执行全量build/tests或其他远端功能链。
- 终验07:10:56Z production active/PID1195022/NRestarts0/healthok、SSHtransport0；自有生产会话关闭0，未持久化密码。两端起末结构/聚合及原始TSV SHA保持，不冒全库数据字节不变。
- 权限/触发器/配置/业务表/客户资产/旧PDF/Legacy目录操作0；不生成部署计划或自动执行后续。检查完成后停在明确范围和独立恢复演练授权前。