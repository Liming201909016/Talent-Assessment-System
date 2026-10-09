# 胜任力一期 v2 Result Run 实施复审

> 复审日期：2026-09-19  
> 复审范围：FB-184 单卷历史重算、新答卷提交接入、011结构门禁、管理员重算接口  
> 本轮性质：只读评估；未修改Go生产代码、数据库、环境配置或远端环境

## 1. 结论

**结论：方案方向正确、本地编译与现有测试通过，但FB-184尚不能进入011～013真实迁移或staging数据写入阶段。**

当前已具备以下正确基础：

1. v2 run及overall/module/dimension/validity写入位于同一GORM事务。
2. 新答卷提交和历史重算均先锁定paper行，再按`paper_id + scoring_version`判断幂等。
3. v2分数沿用精确`big.Rat`计算并以六位decimal持久化。
4. 历史重算读取冻结题目、维度、原始答案和最终题分，并复核正反向计分一致性。
5. 已有run不使用`Save/Update/Delete`覆盖；发现不一致时失败关闭。
6. 单卷重算接口仅允许管理员ID=1或全局权限。

但复审确认存在4项发布阻断和5项高优先级缺口。现有FB-184测试中的核心事务、幂等和结构门禁主要依靠源码字符串断言，不能作为真实数据库原子性或并发证据。

## 2. 发布阻断项（P0）

### P0-1：构建阶段未限制基层员工受众

`buildPhase1V2ResultRunRecords()`直接将旧结果的`ReportAudience`写入v2 run，但v2报告读取器只接受`frontline_employee`。

影响：若历史v1结果受众为`leader`或异常值，系统会创建一条状态为completed、不可修改、但永远不能生成v2报告的run。

修复要求：写入前明确要求`legacy.ReportAudience == CompetencyReportAudienceFrontlineEmployee`，并新增非基层受众RED测试。

### P0-2：冻结题目联结没有证明属于同一exam

`loadPhase1V2AnswerInputs()`通过`pq.exam_question_id -> q.id -> d.id`读取快照，但查询只限制`pq.paper_id`，没有同时限制或验证`q.exam_id`、`d.exam_id`等于已锁paper的`exam_id`。

影响：若历史数据或人工修复造成跨测评快照引用，仍可得到结构完整且看似有效的v2 run，破坏审计身份链。

修复要求：函数接收锁定的`examID`，在JOIN/WHERE中同时校验question和dimension快照归属；增加跨exam污染RED测试。

### P0-3：每次提交在事务和paper锁内执行约20次结构探测

`phase1V2ResultRunSchemaState()`每次调用执行5次`HasTable`、6次`HasColumn`、3次`HasIndex`、6次`HasConstraint`。`Submit()`在创建v1结果之后、提交事务之前调用该函数。

影响：300人集中交卷时约产生6,000次`information_schema`查询，并延长每个paper事务和数据库连接占用时间。它不会造成同paper互锁，但会形成数据库元数据与连接池突发负载，和现有300并发容量目标冲突。

修复要求：将完整结构校验移至启动/部署预检或有界缓存；请求事务只读取预检结果。迁移后必须重启服务或使缓存可控刷新。

### P0-4：原子性、并发与幂等没有可执行测试证据

`TestBugFB184_ResultRunWriteIsAtomicIdempotentAndSubmitWired`主要搜索`Transaction`、`FOR UPDATE`和模型名称等源码字符串，以下行为均未实际执行：

- 任一子表插入失败是否ROLLBACK且零部分行；
- 两个连接并发重算同一paper是否只生成一套run；
- 已有合法run是否零写入并返回`reused=true`；
- 缺少/篡改任一子表行是否拒绝且不修改；
- 011完全缺失和各种部分结构是否分别兼容/失败关闭。

影响：当前“原子、并发幂等、完整门禁”是静态实现判断，不是已验证事实。

修复要求：先补RED行为测试，再修正生产代码；至少包括sqlmock事务回滚、合法复用零写入、损坏run拒绝、结构状态矩阵。真实MySQL并发测试必须在应用011后于staging执行。

### 2.1 FB-185A关闭记录（2026-09-19）

P0-1与P0-2已按RED→GREEN关闭：构建器明确拒绝非`frontline_employee`受众；历史答案加载接收已锁paper的exam_id，并在冻结题目和冻结维度两个JOIN上同时约束该exam。聚焦2项、相关result-run/runtime 3项、Go全量和build均通过。此时尚未处理的P0-3、P0-4及第3节P1问题，已在后续2.2～2.3节继续收敛。

### 2.2 FB-185B关闭记录（2026-09-19）

P1-2已按RED→GREEN关闭：模块行增加`module_id`固定语义校验；run source统一限定为`submission/historical_recompute`，并通过共享验证接入创建、已有run复用和正式报告读取。测试同时证明历史重算可以复用由submission创建的合法不可变run，不要求source等于当前调用来源，也不改写原source。其余问题的后续处理见2.3节。

### 2.3 FB-185C～H关闭记录（2026-09-19）

- C：真实writer注入module写失败后执行ROLLBACK且无COMMIT；合法已有run复用无写操作。
- D：五表结构改为事务前一次缓存的完整information_schema签名，覆盖字段定义、唯一键、外键及013可选索引形状；应用迁移后必须重启进程刷新缓存。
- E：提交、超时自动提交、管理员重算和v2报告生成统一返回稳定领域错误，内部原因仅写服务日志。
- F：011 overall新增`user_time`冻结值，创建/复用校验并由v2报告读取，不再读取可变paper时长。
- G：012使用MySQL 5.7兼容的条件化不存在表查询，在预期tuple变化或作用域额外行时确定失败。
- H：011检测013的复合报告FK，存在时跳过已证明兼容的paper_id重复ALTER，仍对exam/participant完成对齐。

FB-185I双连接真实MySQL测试代码已完成，但本机没有`FB185_MYSQL_DSN`，运行结果为明确SKIP。该项与011～013真实首次/重复执行仍是进入staging重算前的环境门禁。

## 3. 高优先级缺口（P1）

### P1-1：011结构门禁只校验名称，不校验完整签名

当前门禁未校验全部必需列，也遗漏`uk_result_run_module_order`和`uk_result_run_dimension_order`；`HasIndex/HasConstraint`只证明同名对象存在，未证明列顺序、唯一性、引用目标和`RESTRICT`动作正确。

修复要求：采用一次性information_schema签名校验，精确核对表、全部运行时字段、主键、5个唯一约束、6个外键及其列/目标/动作。

### P1-2：已有run复用不校验`module_id`及来源枚举

正式报告重建校验`module_code/name/order/score`，但没有校验持久化`module_id`；run只要求source非空，没有限定为`submission/historical_recompute`。

影响：语义身份或来源被篡改后仍可能被视为合法幂等结果。

### P1-3：作答时长没有冻结到run

v2报告人员和分数读取run快照，但`UserTime`仍从可变的`el_paper.user_time`读取。run模型和011均没有对应字段。

影响：历史报告重新生成时，人员与分数属于run，时长却可能反映后续被修改的paper值，不满足完整审计快照。

建议：在run或overall冻结`user_time`，并由v2报告只读取该冻结值。此项需要独立迁移设计，不能直接修改已评审的011后静默上线。

### P1-4：新路径可能向客户端暴露原始数据库错误

提交与重算handler直接返回service错误文本。新路径中的GORM/MySQL错误可能包含表名、列名、索引名或SQL片段。

修复要求：service对外返回稳定领域错误；内部详细错误写安全日志，HTTP响应不暴露数据库结构。

### P1-5：012种子重跑不能识别目录漂移

012对目录和映射使用`ON DUPLICATE KEY UPDATE id=id`。若已有同ID记录的名称、顺序、模块或映射目标错误，重跑会静默保留错误值。

建议：迁移末尾增加精确签名校验，发现10维目录或10条映射漂移时使部署失败。

## 4. 测试补强顺序

1. **FB-185A：输入身份RED**  
   覆盖非基层受众、跨exam快照、89/91题、空答案、raw/final不一致、非法方向、重复效度顺序。
2. **FB-185B：幂等完整性RED**  
   覆盖合法run零写入，以及run/overall/module/dimension/validity任一字段或行损坏；包含`module_id`和source。
3. **FB-185C：事务回滚RED**  
   使用现有`go-sqlmock`，让第三类子表插入失败，断言`BEGIN -> writes -> ROLLBACK`且无`COMMIT`。
4. **FB-185D：结构门禁RED**  
   覆盖零表、每张缺表、缺列、错误唯一键列序、错误FK目标/动作。
5. **FB-185E：真实MySQL并发**  
   应用011/012后，用两个独立连接并发重算同一paper，断言一套run/1 overall/3 module/10 dimension/1 validity，并获得一次新建、一次复用。
6. **FB-185F：容量回归**  
   证明提交热路径不再逐请求访问information_schema；执行300人集中交卷压测。

## 5. 本轮验证证据

- `go test ./... -v -count=1`：通过。
- `go build -o bin/server.exe ./cmd/server`：通过，零错误。
- 复审未执行011～013、未连接真实MySQL、未调用重算API、未生成v2报告、未部署任何远端环境。

上述通过结果只证明当前代码可编译且已有测试为GREEN，不证明本报告列出的真实事务、并发、结构签名和容量分支。

## 6. 建议决策

**暂停执行011～013及staging历史重算。** 下一步应只做FB-185 RED→GREEN加固；完成后重新执行Go全量/build，再进入真实MySQL迁移首次/重复执行、单卷重算、并发重算和报告生成验收。

## 7. Staging执行结果（2026-09-19）

- 已在20.200.136.133完成数据库和旧后端备份；011～013首次及重复执行均成功，production未修改。
- FB-185I已在隔离MySQL数据库以双连接真实执行：一次创建、一次复用，子表数量为1/1/3/10/1；隔离库已删除。清理临时授权时曾误撤应用账号原有`element.*`权限，已立即恢复并以表查询和health确认恢复有效。
- staging首次历史批次对15份完整v1答卷创建15份v2 run，数量为run/overall/module/dimension/validity=`15/15/45/150/15`；第二次执行为创建0、复用15，数量不变。
- 首次重算真实发现FB-186和FB-187：MySQL保留别名及GORM投影映射缺失。两项均按RED→GREEN修复、全量测试/build通过、重新部署并以15份真实重算关闭。
- 已从客户工作簿确定性生成124条v2候选文案，JSON/SQL二次生成字节一致；staging仅导入`status=1`非临时文案和`draft`内容包。没有伪造内容负责人或心理测量负责人批准，正式v2报告生成继续按设计失败关闭。
- 用户随后明确提供独立具名批准：内容负责人LIming、心理测量负责人Ruiling，并批准内容SHA=`329409e408f10ec7f048a757e83c963b7736f66e41e8d601ff4397fd8545b48c`、工作簿SHA=`edb9efd27ec86bc34db3a796c2022a99495fd9ec52e8a6580cd7404b2ab933b5`用于staging。015事务将124条文案激活并把包标记为staging approved。
- 激活后发现FB-188：staging主机为复用生产形配置而设置`APP_ENV=production`，严格环境批准误判为production。新增`REPORT_EFFECTIVE_ENV`并保留`APP_ENV`回退，staging systemd drop-in仅设置`REPORT_EFFECTIVE_ENV=staging`；production未修改。
- 真实历史答卷`0aae19a9-cfb2-4fc4-b4bd-493bbc11056e`已生成v2报告`bc9a222a-509d-4c3b-b328-c82b411b84da`，绑定run `b8741b37-21b4-4ac8-97b2-42c1b2d84788`。认证下载为`application/pdf`，大小819286 bytes，SHA=`2e7cf492d5380b9842e005714739847ff5aa3e0a3fbc0b41d05df46f5bed2a50`，`pdfinfo`确认10页、A4。相同paper的v1报告记录和旧PDF仍存在，旧PDF SHA=`fad1b708dff42389eb1da4d94a34ea52b0e5393a09cb223a0fba6f5ec59a2b99`。
