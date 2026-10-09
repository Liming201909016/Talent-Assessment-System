# MT-NEW-DRAFT — 本地持久化草稿门禁

## 未验证 / 发布边界

**仅本地实现，ready for independent review，不是staging验收通过。** 新DDL未在任何数据库执行；真实MySQL首轮/重跑、并发、多实例、完整浏览器和目标环境未验证。独立CodeReviewer由主协调者另行调用，本模式未派发。正式报告/两端正式功能、formal registry后续、主HTTP/Worker竞态及正式内容批准仍未完成，不因本次全量GREEN关闭。

本次SSH、远端HTTP、SQL/DDL/DML、部署、重启和生产操作全部0；没有历史回填、关闭未答旧卷、重算、身份/答案/PDF写入。旧exam1776822816300709851及其历史2完整/1未答/profile0仅沿用户政策保留，本轮没有重新查询其数据。旧21项不回滚、不顺手格式化，既有正式版本源不改。

## 已决定的契约

- 采用独立`el_mng_exam_draft`，不增加共享`el_exam`字段、不复用competency或报告模式；所有正常新002由服务器真实`el_repo.code`判定。当前唯一正常创建入口是[Save](../Go-based%20Refactored%20System/internal/handler/exam.go#L410)，未找到测评clone/import创建路由；人员导入不是测评创建。
- 同一Save事务写exam、题库关联和草稿；任一步失败回滚。boolean缺失默认新版；显式false对新002为HTTP400，畸形值也400；客户端伪repoCode不决定产品。
- 单一真实00201/00202、140单选且其余题型0、legacy+legacy、25分钟、关闭参与者报告、受支持非空唯一身份字段；不把mixed或不支持配置偷偷重写为可用版本。
- 保留“保存确认→独立冻结确认”。取消冻结留下持久化draft，不能进入旧链。draft编辑可以改00201→00202，sidecar的canonical repo ID/code同事务更新，原create_time保留；不能切非002/competency回流。历史无sidecar的编辑不自动添加记录。
- 冻结仍读取实际题库关联和源题、验证bundle/profile，sidecar与profile/bundle同事务更新为frozen，标记不删除。未成功提交不返回profile；末次marker更新失败回滚bundle/profile。更新时间不使用profile的秒级截断，避免同秒创建/冻结逆序。

## 服务器动作矩阵

| 生命周期 | 配置编辑 | 人员准备 | 参与者登记/开卷 | 旧结果/报告 |
|---|---|---|---|---|
| draft | 管理员允许，仅合法002，保create_time | 管理员精确同exam、无paper范围允许tester新增/修改/导入及真实GET列表 | 拒绝；candidate Save不因frozen=false回旧链，tester缺profile拒绝，旧create-paper拒绝 | 不提供旧计算/报告 |
| frozen | 原只读/旧guard继续 | 旧接口拒绝，不能用准备例外操作已冻结实体 | 原专属新版身份/140题快照/续答链 | 原专属TEST结果；formal仍未交付 |
| legacy | 原历史行为，不自动转新版 | 原既有策略 | 原既有历史作答政策，不擅关闭 | 原历史读取/PDF保留 |

准备例外只限明确路由，JWT后还必须user_id=1或`*:*:*`；角色名称不是授权。无exam筛选的全库集合不因任一未冻结draft放开；相同GET examId重复收集可归一，不同exam/既有paper或跨人员闭包不授例外。sidecar按1000个ID一批探测，无新增逐exam N+1。原旧保护、capture、report audit和formal配置分类均保留。

## 新表 / 结构门禁（未执行）

[独立DDL](../scripts/sql/management_traits_003_new_draft.sql#L1)只创建一张表：`exam_id varchar(64)` PRIMARY、`repo_id varchar(64)`、`repo_code varchar(5)`、`lifecycle varchar(6)`、created_at/updated_at datetime(6)、可空frozen_at。exam_id动态继承实际el_exam.id字符集/collation，FK RESTRICT/RESTRICT；其他字符串utf8mb4_bin。既有schema中repo.id为utf8mb3，故不伪造与其不兼容的FK；实际repo来源由应用和冻结事务验证。

首轮/重跑核7列/type/NULL/collation、精确PRIMARY、唯一RESTRICT FK及引擎，不修漂移；没有旧ALTER、业务DML、backfill、DROP或FK关闭。运行时[独立门禁](../Go-based%20Refactored%20System/internal/service/management_traits_draft.go#L18)拒绝跨schema FK及结构漂移，不负缓存、不AutoMigrate、不加入启动强制11表合同。缺新表使新002创建503且首个DML前拒绝；原legacy Detail读为明确legacy，非002创建不要求新表。安装错误结构时保护拒绝，不能当缺表放行。

## API变化及全部直接消费方（C2）

Detail新增`isManagementTraits`严格bool和`managementTraitsLifecycle`（draft/frozen/legacy），保留`managementTraitsProfileFrozen`严格bool。已冻结002仍只投影经过原完整profile验证的身份字段；未冻结字段不被客户端当可登记合同。新002 Save返回当前ID、true/draft/false；原其他Save响应不改。

| 实际消费方 | 同步状态 / 理由 |
|---|---|
| [管理表单](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L127) | 同步：新002必选不可取消、发送true、保存验证draft DTO、取消保留、重读draft可编辑且只切合法002；历史不自动勾选 |
| [考生](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L250) | 同步：当前ID＋strict bool＋明确lifecycle；draft提示等待管理员且不展示字段/不保存；unknown关闭；frozen仍configured-only；legacy必须明确false/legacy |
| [封闭登录](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/tester.vue#L72) | 同步：普通002 draft/缺失/不一致meta拒绝；原专属登录服务器再查真实profile，不靠UI授权 |
| [准备页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/preview.vue#L111) | 同步：无专属凭据的002只有明确同exam legacy可走原开始；draft/未知/冻结却无凭据不走旧create |
| [旧人员详情](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L294) | 同步：明确draft直达UpdateExam，不加载旧参与者/报告；提示封闭人员在管理页精确筛选准备 |
| [API wrapper](../Go-based%20Refactored%20System/ruoyi-ui/src/api/exam/exam.js#L7)、[独立API](../Go-based%20Refactored%20System/ruoyi-ui/src/api/managementTraits.js#L56) | 无需修改：原样返回新DTO/原严格字段白名单请求；不扩展公开身份写字段 |
| [传统结果](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result.vue#L353)、[002结果](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L229) | 无需修改：元数据新增键被忽略，实际paper/旧计算由原guard拒绝新draft/frozen；历史读取仍原链 |
| [00401结果](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/competencyResults.vue#L197) | 无需修改：competency不进入002生命周期；原字段不变 |
| [列表主入口](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/index.vue#L278)、[首页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/index.vue#L154) | 无需修改：未缓存的新draft进入旧详情后被服务器DTO消费重定向；已有冻结/探测错误保持原安全分流，不放旧报告 |
| [tester管理](../Go-based%20Refactored%20System/ruoyi-ui/src/views/tester/tester/index.vue#L228)、[ExamSelect](../Go-based%20Refactored%20System/ruoyi-ui/src/components/ExamSelect/index.vue#L26)、[ExamNameSelect](../Go-based%20Refactored%20System/ruoyi-ui/src/components/ExamNameSelect/index.vue#L26) | 无需修改：调用paging/fetchList而非Detail/Save；沿实际examId准备，服务器精确scope决定授权 |
| [团队PDF](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/team.vue#L100) | 无需修改：只用pdfTeam，不消费新DTO；新版旧PDF路由仍保护 |
| Gin/SFC/API/DI/metadata tests | 已同步：合法夹具携新状态，旧strict负例、迟到响应、历史false、字段子集保留；只有已被新政策取代的opt-out期望变更 |

## 服务调用影响（C5）

没有修改任何现存公开函数签名。新增Prepare/Persist仅实际Save调用；ReadDraft由Save编辑/冻结调用；Lifecycle由公开Detail调用。`ManagementTraitsIdentityScope`的candidate handler、注册同手机号既有owner检查、TryTesterIdentity无需改调用签名，自动包含新draft；TryRegister在事务/legacy fallback前新增明确draft拒绝。`CheckManagementTraitsLegacyScope`的handler middleware、无exam tester身份和服务participant闭包自动增加保护，现存测试/真实DB opt-in调用不授准备例外。

`ManagementTraitsLegacyScopeRequest`只新增内部准备意图及输出指针，不是JSON DTO；只有旧route guard设置，随后[JWT后授权](../Go-based%20Refactored%20System/internal/router/router.go#L67)。`managementTraitsSchemaByteBudgets`的schema容量发布及guarded DB fixed容量两处无需变调用；只添加draft键的真实exam/repo父预算，其他键不变。

## 源修改索引

- [草稿模型](../Go-based%20Refactored%20System/internal/model/management_traits_draft.go#L6)、[schema/创建/读取](../Go-based%20Refactored%20System/internal/service/management_traits_draft.go#L18)：新sidecar职责。
- [Detail](../Go-based%20Refactored%20System/internal/handler/exam.go#L329)、[Save事务](../Go-based%20Refactored%20System/internal/handler/exam.go#L660)：真实来源、新建原子写及生命周期响应。
- [身份probe与draft拒绝](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_identity.go#L73)、[冻结绑定](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_profile.go#L81)：不以profile0允许新草稿legacy fallback。
- [服务旧scope](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L298)、[route准备例外](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_guard.go#L403)、[可选ID预算](../Go-based%20Refactored%20System/internal/service/management_traits_schema.go#L565)：所有旧读写路径保护与限定管理员准备。
- 五个前端源的精确修改位置见C2表。
- 既有测试修改：[公开Detail合同](../Go-based%20Refactored%20System/internal/handler/exam_management_traits_candidate_test.go#L15)、[可选缺表冻结夹具](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_profile_test.go#L106)、[metadata精确alias夹具](../Go-based%20Refactored%20System/internal/service/management_traits_schema_alias_test.go#L204)、[实际router DI SQL夹具](../Go-based%20Refactored%20System/internal/router/management_traits_di_test.go#L233)、[管理SFC](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-admin.spec.js#L151)、[身份SFC](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-participant.spec.js#L95)、[年龄合同夹具](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-delivery.spec.js#L30)。

## 本地TDD与最终验证

| 执行 | 实际结果 |
|---|---|
| 初始真实Save RED（6案例＋顶层） | native1，7fail事件；未持久化草稿/false/mixed/缺schema/失败事务均未满足新合同 |
| 初始实际SFC draft/unknown/missing生命周期 RED | native1，3fail，64未选择；不是编译错误 |
| GET精确draft列表新增RED | native1，1子项＋顶层fail，证明相同examId重复收集误拦；归一相同scope后GREEN |
| [新建Save](../Go-based%20Refactored%20System/internal/handler/exam_management_traits_draft_test.go#L34) | 缺boolean/true成功、伪客户端code不改变server002、false400、missing schema无DML、mixed无DML、marker INSERT失败ROLLBACK |
| [实际编辑Save](../Go-based%20Refactored%20System/internal/handler/exam_management_traits_draft_edit_test.go#L15) | 002→002同步canonical源/保create_time；切001 HTTP400、首个DML前回滚 |
| [身份与旧route](../Go-based%20Refactored%20System/internal/handler/management_traits_draft_gate_test.go#L37) | actual Candidate.Save零INSERT拒绝、完整11表schema下tester BEGIN后缺profile ROLLBACK、draft DTO、旧create-paper403、全库列表403、管理员精确准备204/低权403 |
| [公开FreezeProfile](../Go-based%20Refactored%20System/internal/service/management_traits_draft_freeze_test.go#L29) | 完整runtime schema＋实际140源题builder→profile/bundle/draft同事务；末次draft UPDATE故障回滚；source code不符无INSERT |
| 专项增强时 | 36pass事件/native0；随后全量包含最终批量scope/FK拒绝代码 |
| 最终go test -json ./... -count=1 | **6272pass事件/680顶层、0fail、9原环境skip、parse0、native0**；23实库开关只在child清除，stdout7267815 bytes/stderr0，不存raw日志 |
| 最终完整前端Vitest | **32文件435pass、0fail、native0**；原106身份/API用例保留并加3生命周期例，管理增加2例；不是覆盖率百分比 |
| Go服务构建、全包build、vet、独立Linux构建 | **四项native0，stdout/stderr各0**；Go Build task也完成 |
| npm run build:prod | **native0/Build complete**；仅原asset/entrypoint两项体积warning，stderr603 bytes/Browserslist通知，不升级依赖 |
| 18个相关源/测试编辑器诊断 | 0错误 |

第一轮集成的合法管理员上下文/旧冻结可选metadata夹具、全量DI与alias精确SQL、年龄DTO夹具失配均保留为测试同步失败，不冒称产品RED。初始全量Go6255pass但DI/alias失败，修正后全量6268pass；最终新编辑/公开draft断言及批量门禁收口为6272。前端首次194中17失败（新协议及被取代opt-out夹具），同步后194通过，最终全仓435。

都是实际GORM/sqlmock/Gin/真实SFC编译与DOM＋mock API，不是实库DML/实际MySQL锁/桌面浏览器用户操作。本轮尚未逐个注入全部列/索引/FK漂移、所有写点COMMIT失败、真实人员Excel导入、跨实例、活跃卷维护、140题完整UI或新PDF验收；已有全量测试不代替这些新表实证。

## 范围保护与后续

562项源码/测试/SQL起始SHA基线：最终19既有项变化、543完全相同；新增6 Go文件＋1 SQL共7项。既有formal源变化0；Legacy目录本轮本机不存在，仅报告操作0，不当逐文件SHA证明。旧脚本/收据/资产不覆盖，no new .github rules；所有文档在docs、SQL在scripts/sql、Go与运行产物在B区/bin。

一次长scope命令只返回中断符，随后缓冲输出出现exit1，未作命令成功依据；独立短命令实际得到562/19/543/formal0及DRAFT_SCOPE_FINAL_COMPLETE结束标记。Linux本地产物50081033 bytes、SHA dca06c3015803ba4b9e8a514a9ab22521d0906a1f8d63e9d26380fa80e09cac3，仅bin保存未上传。

下一由主协调者安排独立CodeReviewer。已有“先staging修复”批准不自动覆盖新增schema安装：必须说明并另获**staging新草稿表migration＋统一backend/frontend修复发版**许可；fresh只读预检、活跃/待处理卷窗口、受限完整备份与可用回滚、独立恢复库DDL首轮/重跑、真实应用门禁验证后才可主库安装和统一发行。存在nonowned活跃/待处理卷即停，不自动结束。回滚必须保留sidecar/标记/历史，不DROP旧表、清标记或整库还原强迫legacy回流。