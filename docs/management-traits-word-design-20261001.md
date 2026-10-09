# 002管理特质：版本化结果、题目冻结与Word模板实施设计

日期：2026-10-01。修订：**R2（完整评估后修订），仍待设计/业务确认；没有新增代码、表、接口、配置或运行模板。**

依据：[全方位评估及用户确认记录](management-traits-word-assessment-20261001.md#L287)、[完整代码差距评估](management-traits-code-gap-analysis-20261001.md#L137)。文档归C区。本文所有`mng-*`版本、`el_mng_*`表、字段及API均为**拟新增设计**，不是当前系统已存在的能力。R2取代旧稿同事项的版本/hash、历史profile、人员身份、94必需键和current代际提案；保留历史审查记录，不把设计修订称为运行风险已修复。

## 1. 边界、阻塞与可先行工作

已确认业务：00201/00202两题本、13维百分制及四模块、精确计算与最终两位、顶3/底3、Excel条件文案、PDF交付、历史保留、140题完整门禁。团体/千人扩容/矛盾预警/35分钟提示不在本轮。

没有关闭的决定：模块等级/比较文字、内容错字与频率、V67最终题干、名称/固定统计说明、最终页数、具名正式内容批准、历史样本、自动生成触发。

本设计将这些事项与基础开发解耦：纯评分和版本基础可先经批准实现；含未确认文字的包保持draft；正式内容/模板/环境未获批准时不得生成正式报告。**本文不是可直接执行的最终实施计划**，未关闭事项不会被默认值悄悄补齐。

推荐第一实现切片为本地纯身份/评分契约及测试，入口保持关闭，不前置全套12类存储。任何数据库或运行接入之前，须完成§12保护门禁：删除以及Save/Update解除关联、匿名管理写、旧PDF双写均需覆盖。每个切片单独完成、验证、等待确认，不一次实现整个改造。

## 2. 实际现状与必须保持的接口边界

| 现有证据 | 对设计的约束 |
|---|---|
| [Exam模型](../Go-based%20Refactored%20System/internal/model/business.go#L68-L100)及[legacy保存分流](../Go-based%20Refactored%20System/internal/handler/exam.go#L419-L452) | 002仍为legacy，独立sidecar承载版本；不占用会被清空的competency版本字段 |
| [传统组卷](../Go-based%20Refactored%20System/internal/handler/paper.go#L222-L420)按当前题库sort及LIMIT取题 | 新版必须验证唯一002题本及全140题；公式V序和个人展示序分离 |
| [PaperQu/选项关联模型](../Go-based%20Refactored%20System/internal/model/business.go#L154-L185)没有文字快照 | 新增自己的冻结题干/选项，不能宣称旧paper已有冻结题本 |
| [当前V映射](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L1565-L1600)来自当前关联sort | 新版加载不得再调用该函数推断公式题号 |
| [交卷事务](../Go-based%20Refactored%20System/internal/handler/paper.go#L669-L731)没有显式FOR UPDATE | 新版锁定paper再读答案、写结果；答案保存必须同锁，事务存在不等于并发互斥 |
| [服务端002生成](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L191-L323)与[浏览器回写](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L863-L945)共用人员PDF指针 | 新版统一独立Word链；保留旧端点结构但阻止对新版paper进行旧截图写入 |
| [匿名路径规则](../Go-based%20Refactored%20System/internal/middleware/middleware.go#L35-L62)、[ShowPdf](../Go-based%20Refactored%20System/internal/handler/paper.go#L889-L909)及[旧下载](../Go-based%20Refactored%20System/internal/handler/candidate.go#L657-L744) | 新报告不返回服务器路径、不借公开旧接口下载；管理鉴权和paper归属校验前置 |
| [LibreOffice转换器](../Go-based%20Refactored%20System/pkg/libreofficepdf/client.go) | 复用基础设施；不改MBTI/00401业务契约，不宣称批量容量已验证 |

本轮没有连接DB或远端：题本对应关系、实际索引/collation、现存题库是否漂移、静态alias是否覆盖PDF目录均未实证。

## 3. 版本与身份：独立于题库序号、中文名和人员类型

### 3.1 版本元组（拟定，需设计确认）

| 轴 | 拟定值 | 变化原因 |
|---|---|---|
| product | `mng-traits-v2` | 新百分制/报告产品 |
| question | `mng-staff-20260928-v1` / `mng-leader-20260928-v1` | 两套题干独立；未来修订V67必须变更包hash，不能覆盖旧包 |
| scoring | `mng-percent-scoring-v1` | 正反向、除数、五档、模块权重及同分规则 |
| norm | `mng-norm-20260928-v1` | 13常模及综合精确值 |
| content | `mng-content-20260928-v1` | 205条条件文案及正式审核后的修订 |
| template schema | `mng-word-schema-v1` | 字段、对象及六图结构合同 |
| template layout | `mng-word-layout-v1` + 每份SHA | 客户字体/分页编辑形成新revision，不用业务版本代替文件hash |

**版本职责拆分：**上表是系统manifest，不是评分输入元组。评分事实只绑定product/question/scoring/norm及其评分manifest SHA；content、template schema/layout属于呈现轴，在生成报告时单独选择并冻结，不进入评分输入hash。含内容/模板的综合包必须提供独立`scoring_manifest_sha`，不能用整个综合ZIP/JSON SHA迫使换模板后重算。

生命周期：答案、题本或评分规则变更才建立新评分事实；norm变化建立新norm版本run（不改旧run）；内容修订或模板样式变化只建立新report revision，复用原completed run。新内容包/模板需验证支持该run的product/question/scoring/norm，版本兼容缺失失败关闭，不能直接拿“当前活动稿”跨版本渲染历史结果。

设计建议00201→staff、00202→leader；以精确repo code及题本核验确认，不按`isOpen`、`stuFlag`、人员ID长度或名称猜测。该对应仍须读DB与客户基线核对；一场新版测评不允许混合题本或多个002题库。不改变001/003/00401配置与分流。

### 3.2 稳定维度键（拟定键；顺序、题数、常模来自已核验材料）

| order | key | 名称 | module | 题数 | 常模 |
|---:|---|---|---|---:|---:|
| 1 | self_confidence | 自信心 | self | 12 | 57.50 |
| 2 | emotional_stability | 情绪稳定性 | self | 10 | 55.00 |
| 3 | self_discipline | 自律性 | self | 10 | 56.25 |
| 4 | sociality | 社会性 | interpersonal | 10 | 51.25 |
| 5 | leadership | 领导性 | interpersonal | 10 | 52.50 |
| 6 | interpersonal_sensitivity | 人际敏感性 | interpersonal | 13 | 50.00 |
| 7 | cooperation | 合作性 | interpersonal | 10 | 57.50 |
| 8 | planning | 计划性 | task | 12 | 53.75 |
| 9 | responsibility | 责任心 | task | 10 | 58.75 |
| 10 | decisiveness | 决断性 | task | 12 | 53.75 |
| 11 | proactiveness | 进取性 | development | 10 | 55.00 |
| 12 | learning | 学习力 | development | 11 | 53.75 |
| 13 | innovation | 创新性 | development | 10 | 50.00 |

排序同分使用上表order。顶组降序、底组升序，各自截前3；同分均用order升序。全同分两组都为1/2/3，保留两个区块但不推导“绝对优势/短板”标签。四模块拟按模板固定self→interpersonal→task→development展示，仍待确认；不新增按分排序行为。维度同分顺序已确认，不与模块展示顺序混淆。

## 4. 140题身份冻结与历史证据

### 4.1 不可变题本包

生成规范化审阅包时保留原文及来源Sheet/单元格；人工改动形成新hash，不覆盖源附件。每题包含：question_version、V号1～140、稳定维度key、方向、题干、五个选项及原始分1～5；选项排列和分值分别记录，不能按A/B/C位置猜分。

绑定DB源题时需要140条显式`source_qu_id→V号`映射、题干/选项/分值逐项核验及来源证据hash。规范化允许Unicode/空白差异核对，原文保留；实质题干差异、未知选项分、重复/缺失V均阻断，不自动覆盖数据库题目。两题本不同V67/V96不得互换。

### 4.2 新测评与新paper

推荐新增独立“准备/冻结新版题本”动作，固定一场测评的版本元组、全题快照及requiredFields，不沿用00401发布状态。冻结后已绑定内容不再随源题、题库sort或profile编辑改变；有paper时拒绝改变题本及版本。未确认自动触发前保持管理员显式操作。

组卷事务锁已验证的participant及冻结profile，建立legacy paper/paper_qu/答案关联的同时建立002 paper快照和140题快照，原子绑定人员；重复创建返回原卷，全部成功才COMMIT。新版使用§12精确参与者API，旧答题入口只对明确未纳入新版的paper保留兼容；新版/迁移中paper在旧创建、保存、交卷、人员绑定和PDF写接口查询/写入前拒绝，不继续复用匿名旧接口作为安全合同。新版详情读取冻结文字/选项，新版保存使用冻结原始分，不读当前QuAnswer决定分值。

题目快照保存：paper_question_id、source_qu_id、V号、display_order、dimension_key、方向、题干、选项JSON（source answer ID、文字、原始分、顺序）、source_hash。V号是公式身份，display_order是屏幕顺序，两者即便相等也不能合并。

### 4.3 历史paper

旧记录没有题干快照，当前题库内容一致**不能证明作答时一致**。必须提供独立历史证据（经核验的备份/旧题本版本、稳定题ID映射及选项分值），覆盖全部140题和两题本差异。证据不足时返回`mapping_unverifiable`，保留原报告，禁止以当前sort、模糊题干或最新题本猜测。

首次获准历史重算：限定staging白名单paper；锁paper，检查完成及140答，保存逐paper证据包和完整输入快照，再原子写新版run。**历史证据不绑定或覆盖exam唯一新组卷profile**：同exam下不同历史paper可以各自指向经审核的证据版本；profile只约束将来新组卷。历史身份/requiredFields若只能取当前人员和测评值，必须注明`captured_at_recompute`，不能假称提交时冻结；是否可正式出报告由内容/身份审核门禁决定。

人员身份解析：传统paper.user_id=101不具证明力。历史卷必须用`paper_id + exam_id`分别查candidate/tester的有效关联，并核对独立历史证据；恰一条匹配才候选接受，零条/跨exam/多条/跨类型歧义均拒绝，不用“candidate优先”、ID长度或stuFlag猜测。新paper在组卷事务冻结participant_type/id及人员字段，并核验人员关联指向同paper/exam；此后旧人员Update不得重绑。纸卷公共user_id不再参与002身份判定；与其他产品无关的user_id语义保持不变。

### 4.4 输入与完整性

140题身份完整与140题全答是不同校验：前者缺失/重复/错版直接报数据错误；后者超时未答标incomplete，不生成正式run分数或报告。禁止把缺失题置0再反向得到6；禁止截断非法分数掩盖错误。

`actual_score`在旧002是选中选项原始分；新版不能再反向两次。以checked选项和冻结option分重建raw，并与actual_score交叉核验，一题只能选一项，raw必须整数1～5；方向处理一次后最终分仍1～5。UI不暴露方向/评分规则。

## 5. 精确评分与持久化事实

维度整数题分和为$t_i$，题数$n_i$；精确分：

$$S_i=\frac{25(t_i-n_i)}{n_i},\quad M_g=\frac{\sum_{i\in g}S_i}{|g|},\quad O=\frac{\sum_{i=1}^{13}S_i}{13}.$$

用有理数计算、比较和选择；模型展示缓存用DECIMAL(18,6)，但缓存不能代替精确源。维度保存整数sum/count，模块与总体由13个精确维度重建并校验缓存；norm包保存精确十进制/分数，综合705/13不得从已显示54.23逆推。

最终文字使用HALF_UP两位；图表由相同有理数在XML边界输出12位小数（作为图形近似，不再用于分级/聚合）。文字、Excel均复用同一显示格式，不把float64/四位旧均分作为评分源。

五档L1～L5按90/70/30/10已确认边界；内部拟定语义码`excellent/good/qualified/weak/insufficient`仅在002命名空间使用。模块等级和常模比较文字未定，本设计只持有模块score，不设置module.level/compare默认值。

维度中文标签固定优秀/良好/合格/**欠佳**/不足，总体中文标签按Excel总体五档匹配；内部weak不借用00401“薄弱”。顶底摘要、13维评价/建议、总体评价/多段建议分别匹配，不能把底3自动插入总体建议文字；题本物理选项列为5→1含义，原始分按文本身份绑定，不按列序猜测。

标准分旧API继续返回旧1～5结构；新版独立强类型DTO包含schema、paper/run、全版本、人员/字段快照、13维、4模块、总体、顶3/底3和源hash。管理响应不回服务器路径；集合为空时使用[]而非null。

## 6. 数据模型方案（拟新增，未执行DDL）

为避免扩展通用Exam/Paper大表和误用00401，推荐sidecar及并行结果；精确物理表名、列宽和索引需Schema切片确认。配置目录/题本定义可用一份不可变JSON bundle，避免再建重复13维目录表。

| 拟新增表 | 核心信息 | 键/关系 |
|---|---|---|
| el_mng_definition_bundle | product/question/scoring/norm评分manifest、SHA、状态；不含活动content/template | PK id；评分版本组合唯一；审核后正文/hash不可更新 |
| el_mng_exam_profile | 新组卷exam、评分bundle、冻结题ID/选项映射、字段采集合同、时限 | PK exam_id；RESTRICT→exam及bundle；冻结后不可原地改配置，有paper后不换绑定 |
| el_mng_paper_snapshot | paper/exam、nullable新组卷profile、独立评分bundle/逐paper证据、mapping hash、人员及字段、来源/时间/保护状态 | PK paper_id；RESTRICT→paper/bundle；new_creation需profile同exam，historical_evidence无profile依赖；UNIQUE(paper_id,exam_id)供复合FK |
| el_mng_paper_question_snapshot | paper、paper_question、V/display_order、原题/选项/方向、提交raw/final | UNIQUE(paper_id,V)和(paper_id,paper_question_id)；RESTRICT→paper快照；paper_question复合关联须校验同paper |
| el_mng_result_run | paper/exam、评分/常模版本、输入hash、状态、总体精确源引用/六位缓存、时间/人员来源 | UNIQUE(paper_id,scoring_version,norm_version)；RESTRICT复合→paper快照；完成后不可变 |
| el_mng_result_dimension | run、dimension_key/order、sum/count、缓存score/norm、等级 | UNIQUE(run_id,key)、UNIQUE(run_id,order)；RESTRICT→run；恰13条 |
| el_mng_result_module | run、module_key/order、成员数、六位score缓存 | UNIQUE(run_id,key)、UNIQUE(run_id,order)；RESTRICT→run；恰4条 |
| el_mng_content_package | 内容版本、适用题本受众、来源/内容SHA、双批准、环境、draft/approved/retired | 版本+受众唯一；批准绑定精确SHA，禁止推测批准人 |
| el_mng_report_text | package、类型、dimension/overall、L等级、正文、source坐标/审核状态 | UNIQUE(package_id,type,identity,level)；总体identity用overall而非NULL；RESTRICT→package |
| el_mng_report | source、paper/exam、nullable run、content package、template schema/layout/SHA、revision、文件及DTO快照 | PK id；RESTRICT→run/paper；source区分word_v2与legacy_capture；legacy_capture不强制新版run |
| el_mng_report_current | paper、报告对象/新版展示槽、report_id | PK(paper_id,slot)；RESTRICT复合→report同paper/slot，阻止跨卷指针 |
| el_mng_report_audit | actor、paper/report、动作、成功/错误码、时间及request ID | RESTRICT→相关报告；不含密码、答案全文、路径或token |

12类表是逻辑职责划分，不是已批准建表数量；review若可安全合并可在Schema切片调整，但不得丢掉冻结/版本/审批/历史不变量。

participant_type/id为来源信息，不直接FK到candidate/tester多态行；冻结字段不依赖人员后来修改。状态incomplete与数据异常区分；完整run的13+4子项必须同事务写入。legacy_capture的迁移保护记录不得伪造bundle或140题：可先登记report及独立保护标记，只有证据核验后才创建评分paper快照；物理marker可独立表或受控保护实体，Schema切片定稿，不能要求capture先通过评分证据门禁。

身份链补充：新组卷profile的exam_id必须等于paper.exam_id；历史来源不要求profile。paper_question必须属于快照同paper。run设置UNIQUE(id,paper_id,exam_id)，report通过(run_id,paper_id,exam_id)复合FK绑定；report设置UNIQUE(id,paper_id,slot)供current关联。底层旧表若无所需复合唯一键，Schema切片需评估只增索引，不能假定已有；同事务加载器核验paper.exam_id、快照participant_type/id的真实paper+exam关联及paper_question.paper_id，**不得校验participant_id等于旧paper.user_id**。legacy_capture的NULL run不免除paper/exam及历史人员归属审核。

评分输入hash合同：用V号升序的规范化UTF-8序列，包含评分输入schema、paper/exam、participant_type/id、product/question/scoring/norm、**scoring_manifest_sha**、mapping正文SHA、每题评分快照hash、冻结选项身份/原始分、已选选项身份、raw及方向。内容版本、模板hash/布局、PDF字段位置及捕获时间不进入评分hash；每题评分hash也不能嵌入这些呈现轴。固定数字格式/JSON键序和空值语义。重复请求从冻结输入重建hash，并核验13维sum/count/等级/缓存、4模块/总体及身份；任一失配拒绝，不覆盖已有run。同(paper,scoring,norm)遇不同题本或输入hash是冲突，不能覆盖原run，纠错重测/新运行身份另审。

报告渲染hash另包含run ID+评分hash、经审核人员呈现快照SHA、content package ID/SHA、模板schema/layout及确切文件SHA、renderer版本/转换器与字体环境签名、完整DTO SHA。报告字段或样式变化只产生新revision；生成启动后不再读取活动内容/模板变化替换其捕获输入。

MySQL5.7+显式幂等迁移，UTF-8无BOM；不AutoMigrate、不改旧整数分、不回填所有旧paper。外键列动态继承父表collation；不依赖5.7忽略的CHECK保障业务约束。物理schema签名需核对列、索引列序与唯一性、FK来源目标；缺表旧链兼容，部分安装新版入口失败关闭。

## 7. 锁、事务、删除与原报告保护

### 7.1 新版交卷

只对paper快照明确标记新版的002执行：事务前检查一次结构可用性→事务锁paper→验证身份/state→读冻结140题及答案→严格校验→生成13维/4模块/总体→写run与子项→冻结提交raw/final和时长→更新paper及完成状态→COMMIT。

答题保存也先锁同paper，再检查进行中/真实时限及冻结题归属；避免交卷读取后仍保存答案。重复提交验证已有run及input hash后只读返回，不重写run；任一子项/状态写入失败全部ROLLBACK。旧001/003/旧002继续原路径，不在本切片修其一般并发问题。

提交语义提案：参与者只能manual请求，锁后由服务器时钟判定。未到期139/140拒绝且保持进行中；未到期140/140完成；到期已全答按timeout完成；到期缺答冻结审计/incomplete、正式分NULL、不补最低分。保存先检查同一可信截止时间，到期拒绝答案写入并进入统一到期结算；前端不单独写人员end_time，不以本地倒计时证明到期。25分钟与20分钟提醒用同一个冻结deadline计算，重登不重置。测评窗口对已开卷的优先关系、离线Worker/扫描延迟和上述到期结算策略仍列待业务确认；未确认前相关运行入口关闭，不借只扫描00401的Worker。报告自动生成触发仍不实施。

### 7.2 历史PDF保护

推荐受控legacy_capture流程把既有PDF复制到新版私有不可变目录并核验SHA/大小，登记来源和原路径证据；不生成新版成绩或解读。记录旧PDF原始来源与捕获时间，不称为Word新版。

在捕获完成前，禁止对目标历史paper执行会删除旧PDF的重生成/旧持久化操作。捕获后旧路径也不由本轮清理。新Word报告使用独立report/current，不更新candidate/tester的旧pdf_path；旧浏览器上传对已纳入新版的paper受控拒绝，防止再次覆盖。

捕获顺序必须先建立持久化迁移保护标记、阻断新写及文件删除，再排空已经在途的生成/上传/旧链异步压缩任务；无法证明排空时只捕获副本而不宣称稳定迁移成功。保护标记可作为paper快照迁移状态，不虚构题本已冻结。排空后从同一打开文件句柄读取字节、计算SHA并写私有副本，核验读前/读后原文件元数据及副本SHA，登记完成才解除捕获中的维护状态；保护失败或捕获失败不删除原件。

对尚未迁移的旧002继续旧行为；进入迁移范围的paper用明确标记分流，不能仅按repo以002开头就全量切换。

### 7.3 删除兼容

现有[Paper删除](../Go-based%20Refactored%20System/internal/handler/paper.go#L172-L210)、[Exam删除](../Go-based%20Refactored%20System/internal/handler/exam.go#L739-L820)、[Tester删除](../Go-based%20Refactored%20System/internal/handler/tester.go#L386-L420)、[Candidate删除](../Go-based%20Refactored%20System/internal/handler/candidate.go#L195-L230)不会自动处理这些表。

另有[LogicDeletePdfByIds](../Go-based%20Refactored%20System/internal/handler/candidate.go#L263-L285)直接清理candidate/tester文件与指针，不经人员删除路径。DELETE/PUT两个路由及前端“删除报告”按钮都需保护：整批先解析人员→paper，任一含迁移保护/新版历史即整批拒绝，发生在任何cleanup或字段更新前。不能只加Paper.Delete守卫后声称旧PDF已安全保留。

推荐初期保守保护：已存在002快照/run/report时，paper和有关联exam删除在任何写/文件清理前拒绝，并告知有历史报告；无paper的draft profile允许同事务先删profile再删exam。已有新版历史的人员删除/软删除亦先拒绝，避免匿名化/保留政策未定时清掉身份或文件。不直接把00401整链物理删除政策套给002。

更广泛的删除/匿名化/保留年限须单独确认；不为绕过FK而级联删除。schema安装前就需上线对应守卫，并在真实数据库确认拒绝零写入、旧无sidecar路径兼容。

## 8. Word合同R2：94键目录候选，不是94个必需键

**94仅是原稿推导的字段目录上限，不是客户文件实测控件数或已批准必需数。**按完整评估材料定位，先将每个实际动态位置与业务值对应，再形成必需/可选/重复白名单manifest；候选模板和注册表同步验证后才冻结schema。固定定义在展示上可留Word，但其与Excel定义的权威维护规则仍待定，不由程序运行时改写客户定义。

| 字段组 | Tag规则（拟定） | 唯一键数 |
|---|---|---:|
| 人员 | participant.name/age/gender/telephone/affiliation/post | 6 |
| 时间 | result.submittedAt / result.userTime | 2 |
| 总体 | overall.score / level / norm / diagnosis / advice | 5 |
| 模块 | module.self/interpersonal/task/development.score | 4 |
| 十三维 | dimension.{§3.2 key}.score/level/norm/diagnosis/advice | 65 |
| 顶3/底3 | overview.high.1/2/3.name/text；overview.low.1/2/3.name/text | 12 |
| 合计 | 6+2+5+4+65+12 | 94 |

结构候选分级：原稿表1有name/gender/telephone/affiliation/职务/测评日期位置；post到职务映射、submittedAt到测评日期语义须确认。age/userTime/overall.norm列为**位置未确认、不得强制必需**；exam.title/report.generatedAt也保持可选审阅。其他score/level/diagnosis/advice及六摘要槽按实际位置审阅，其中多段advice不能由一个计数代替协议验收。未配置字段不强迫采集，显示/隐藏完整行与选填规则尚待确认。表格94合计只计候选注册键，不计必需数。

正文、页眉、页脚及文本框中的真实文本控件都必须扫描；重复scalar键只按明确重复白名单允许（例如封面/页眉name、图中score），不能要求每个Tag只出现一次。完整详情、建议、概览块有独立结构标记与次数门禁，防止复制整块仍因Tag重复被放行。

摘要槽name/text使用独立控件，保留维度名/冒号粗体及正文常规；静态栏目标签位于控件外。多段总体建议的候选协议为每档原文三段分别绑定预定义段落槽，运行时只填值不创建段落布局；若改用换行协议须先验证Word/LO样式及完整段落，二者不能混用。逻辑advice注册键可映射三槽，但manifest必须声明基数/来源段号；正文缺段失败关闭，不把底3强塞进总体建议。协议与模板位置仍待客户审阅。

### 8.1 六图绑定

| 业务key | 类型 | 绑定数据 |
|---|---|---|
| chart.overall | doughnut | [O,100−O] |
| chart.module.self | doughnut | [M_self,100−M_self] |
| chart.module.interpersonal | doughnut | [M_interpersonal,100−M_interpersonal] |
| chart.module.task | doughnut | [M_task,100−M_task] |
| chart.module.development | doughnut | [M_development,100−M_development] |
| chart.dimension.comparison | bar+line | 分值13点、常模13点，分类严格按§3.2顺序 |

通过Word图表对象替代文字title存业务key，以OPC关系解析真实部件；五环图各1系列2点，组合图2系列各13点。多层分类中的模块名可以是模板固定说明，但不能成为第14～17个分值点。现有客户六图缓存清理时需对此实测，不照物理chart编号写死。

一次性适配拆分六个分组graphicFrame，保留固定背景/等级图和样式；原稿不覆盖。运行时不重建图表、不覆盖配色、坐标轴或字体；未确认动态分档配色前颜色仍由模板控制。

五环图旧富文本手工分数必须改成能绑定的普通SDT标签（允许重复score键）或逐对象绑定的数值文本槽；若用后者，schema manifest须记录精确对象key/文本节点边界。不能只更新c:val却保留旧59.45等富文本。选择最终方案前用Word/目标LO真实样例确定可靠性，不把旧数字的全局正则替换用于运行时。

### 8.2 模板生命周期

建议上传走验证→候选保存→本地/目标转换预览→确认激活，不把结构通过等同排版通过或直接替换活动稿。激活方式仍为设计建议，需确认管理交互。

原文件、候选文件、活动revision分别记录SHA；合法Word重存恢复图表外链时可在候选清理，再复验合同，错误计数与候选hash可见。未知外链/宏/OLE/关系路径逃逸/重复ZIP/损坏XML/缺键/错图/旧样例残留拒绝。限制20MiB压缩、64MiB解压为拟定门槛，需测试验证并注入配置，不能仅靠后缀。

候选清理后全OPC关系复验：图表数据只允许字面量分类与数值，残留externalData、公式c:f、numRef/strRef、危险嵌入关系均拒绝。清理外链不能只删Relationship留下可刷新引用，也不能访问外部工作簿求值；能从完整缓存确定性物化时只修改候选，缓存不完整则失败。固定数值占位作为空模板例值可以存在，但填充后不得遗留未绑定动态样例标题；不能把空模板所有数字都当残留拒绝。

包保存为新的ZIP条目，不复制不匹配CRC的FileHeader。只修改必要部件，保留媒体与固定内容。页脚建议仅PAGE，不固定“共10页”；该显示变化需模板审阅确认。Word与目标LO六图全可见及长文案完整后才定页数，不先压缩文字到9页。

## 9. 统一报告运行链与安全交付

生成提案：管理员显式生成/批量生成→权限与paper范围→读取completed run与已批准内容包→捕获活动模板字节/SHA→生成冻结DTO及绑定值→临时DOCX→现有LibreOffice→校验PDF/实际页数→私有目录新文件→短事务登记新revision、审计和current→返回report ID。

每次成功force生成保留旧completed revision。默认复用必须精确匹配run、内容SHA、模板SHA与合同；若模板变更只提示旧报告已过期，不后台重写。转换在DB行锁事务外进行，单进程分片同卷锁控制；最终事务重核paper/run存在性及revision代际，防止生成期间删除或旧慢请求覆盖新指针。

revision/current策略R2提案（待确认）采用**最新成功的启动序号**，替代旧“最新启动优先”。启动短事务分配唯一单调attempt_seq，只增加分配计数，不推进current的last_success_seq；失败不推进成功水位。完成时锁(paper,slot)并校验完整输入、有效策略epoch及文件SHA，仅当attempt_seq>last_success_seq且epoch有效才切current并更新last_success_seq/RowsAffected；若较晚请求已经成功，较早完成的文件仅登记历史，不能回切。较晚请求失败不阻止较早成功成为current。显式选择历史report时管理员原子增加selection_epoch，该时刻前启动的生成不可覆盖其选择；首次空current有可锁定占位。兼容批准撤销/删除保护在完成时再次检查；未明确选择策略前运行入口不启用。同卷锁仅是单进程优化，DB序号/epoch才保证多实例/重启一致性。复用仍核验版本、目录、PDF签名、size/SHA，不能只查completed。

统一目标合同：`result_run_id`和`report_id`不是可互换“当前”。首次结果页显式选completed run，生成请求固定run ID及content package/template revision ID，服务器返回解析后的target manifest；查看/下载只用report ID，不重算；Excel请求携带显式runIds（单卷亦如此）。列表默认“当前”只能经同一解析器返回run/report及selection_epoch，并固定成上述ID后发送，批量开始前冻结每项manifest；缺新版run/缺report明确逐项或整批报错，不退旧版。legacy_capture仅供旧PDF查看，不能充当新版run或新版Excel源。若current report指向另一run，UI显示不一致并要求选择，不能悄悄换run；报告按ID下载、已冻结导出不因current后来变化而改变内容。

报告目录配置独立于公开profile静态alias，不硬编码IP/路径；目录0700、文件0600，鉴权流式读取。最终字节SHA在任何压缩完成之后计算，completed后不允许异步改文件。失败仅清理本次未登记候选，旧文件/current保持；清理器先核对DB引用，再按受控策略处理孤儿，未授权不删除历史文件。

API草案（拟新增；路径以实现前router审阅为准）：管理前缀`/exam/api/management-traits`下提供results/detail、reports/generate、reports/list、reports/download、reports/batch-download、template/info/download/candidate/activate。全部JWT及后台权限，不加匿名prefix。历史重算另为管理员+staging+paper白名单的专属POST，错误不暴露DB细节。

下载按report ID取冻结文件，客户端不能指定filesystem path。列表返回可下载ID及大小/SHA，不返路径。旧pdf-upload/show_pdf对已迁移paper应先分流/拒绝新版文件访问；参加者是否可看报告沿测评已有设置和明确身份鉴权，不因旧路径匿名就放行新版。

批量ZIP建议最多100份、去空去重、请求版本固定；发送头前全部预检、任何文件/权限/版本错误整体拒绝。包内姓名+paper+report保证唯一，不带目录分隔符；审计记录“开始流式提供”而非声称客户端已保存。候选临时ZIP按finally清理。

## 10. 前端、导出与消费方影响清单

以下是**未来需要同步修改/无需修改**的完整职责清单，不是本轮实际改动清单；每个实现切片仍需重新列出真实调用点。

| 消费方 | 处理 | 原因/合同 |
|---|---|---|
| [测评创建编辑](../Go-based%20Refactored%20System/internal/handler/exam.go)及管理页 | 同步修改 | 新版profile显式绑定/冻结，002精确识别，不改变legacy总体类型 |
| [paper组卷/详情/保存/交卷](../Go-based%20Refactored%20System/internal/handler/paper.go) | 同步修改 | 有新版快照才分支；源题改动不改变冻结选项；同paper锁与run事务 |
| [旧标准分tester](../Go-based%20Refactored%20System/internal/handler/tester_score.go)及[candidate](../Go-based%20Refactored%20System/internal/handler/candidate.go) | 无需改变旧返回结构；同步保护新版PDF入口 | 新版分值走独立API，避免旧页面把百分值当1～5 |
| [报告批量生成](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go) | 同步修改 | 按明确迁移标记/逐paper快照和显式run分流至Word，历史新版不要求新组卷profile；捕获中/缺run拒绝，不回退Chromium |
| [行内查看/生成/下载](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue)及[result2](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue) | 同步修改 | 新版查看为只读PDF预览，未生成时提示；旧详情URL不能对新版paper截图回写 |
| candidate/tester持久化、单份/批量下载、ShowPdf | 同步修改分流与保护 | 不依赖人员pdf_path寻找新版；旧历史入口只读受保护文件 |
| Candidate.Save/Update、Tester.Login/Detail/Update、准备页人员绑定、Paper.Save、Exam.State/题库关联 | 同步修改或在新版作用域拒绝旧写入口 | 防关联覆盖/重绑、密码回显与匿名管理写；RESTRICT只保护删除不足以保护历史，详见§12门禁 |
| [原始数据/原始答题/AnswerDetail](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go)及[导出前端](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/index.vue) | 同步修改 | 新版取run与题目快照，不按当前sort重算，显式标注缺run/不可重算，不混旧尺度 |
| 两套002 Excel模板及无模板回退 | 同步修改新版模板合同 | 保留旧文件；新文件与columns版本绑定；迁移/新版paper按显式run导出，不要求历史存在profile，缺run或新模板拒绝而非回退旧导出 |
| [模板管理页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/template/index.vue)、API、[router](../Go-based%20Refactored%20System/internal/router/router.go) | 同步修改 | 增加002专属卡片与候选/激活权限、元数据，不共享00401校验器 |
| paper/exam/tester/candidate删除及LogicDeletePdfByIds两个路由/前端删除报告按钮 | 同步修改保护 | RESTRICT安装前先守卫；捕获标记也保护，整批预检先于文件清理，不能伪成功 |
| 001心理/003MBTI/00401、通用转换器 | 无需修改业务行为 | 本轮保持接口/评分/模板，增加相关兼容回归，不顺手重构 |

Excel合同先定义事实列：人员/时间/题本及评分版本、run、完整性、总体分/等级/常模、4模块score、13维score/level/norm；逐题数据使用冻结V号/题干/选择文字/raw/direction/final。不新增未经批准的百分位/模块等级/矛盾预警。客户尚无新版Excel列布局确认，设计不擅自复制00401三Sheet作为最终002合同。

## 11. 分段门禁与下一切片建议（非最终任务执行计划）

| 切片 | 可验证产物 | 启用条件 |
|---|---|---|
| S1 纯身份/评分 | 两题本定义、13稳定键/4模块、精确评分/五档、顶3底3；边界及畸形输入测试 | 设计确认后本地实现，不接HTTP/DB、不开放报告 |
| S2 题本/结果结构 | 最小评分bundle/profile/逐paper证据/snapshot/run结构、旧写保护、幂等schema与事务测试 | DB题本对照及生命周期确认；不前置全12表，DDL实际执行另授权 |
| S3 Word候选 | 位置驱动必需/可选manifest（94目录非强制）、六图、零外链、动态标题、Word/LO实开及长文案 | 客户字段/段落/固定内容边界审阅；候选不替换活动模板，先确定页数 |
| S4 内容包 | 205条件文案、来源坐标、hash、draft/批准门禁 | 错字频率/定义及固定说明定稿；批准人和精确包hash另确认 |
| S5 运行整链 | 新版组卷/保存/交卷结果，生成、历史保护、预览、下载、导出 | 模块/导出合同与profile启用时点确认；无自动生成默认 |
| S6 staging验收 | 两产品×两人员模式真实整链，历史白名单、DB→PDF→Excel一致 | 备份、样本及明确staging授权；production单独授权 |

测试必须包含：140唯一/100正40反、10/11/12/13分母、0/100与10/30/70/90、精度差异、模块非等权总体、防跨题本/跨paper/跨run、同分全重叠、缺文案/错版/未批准、六图像素存在及标签真实更新、长文案跨页、候选Word重存、转换取消、并发交卷零重复、子项失败ROLLBACK、下载/force生成并发、旧PDF保留、伪路径/匿名拒绝及001/003/00401回归。

按仓库规则：先分支矩阵/RED再代码/GREEN，Go测试放代码区、脚本测试放scripts/test、辅助工具放scripts/tools、SQL放scripts/sql、规范化数据放scripts/data。模板初稿/审阅证据放docs，批准后运行模板才放代码区运行配置目录。本轮不创建新规则文件。

### 设计必须确认而不能默认实施的选择

1. 同意独立sidecar及并行run/report，保留旧整数分及旧API尺度。
2. 同意冻结profile动作及精确00201/00202题本对应，经DB核验后才启用。
3. 同意初期拒绝有新版历史的paper/exam/人员删除；匿名化和保留年限另定。
4. 同意历史证据不足则不重算，legacy_capture仅保存旧PDF，不冒称原题本重算。
5. 审阅94候选目录中的实际必需/可选位置、六图及多段协议，确认“候选预览后激活”；不批准不存在位置的必需字段。
6. 确认§9最新成功序号/显式历史选择epoch策略、统一run/report目标合同，以及§12权限/资源提案；未选择事项不暗设默认。

本轮只完成设计文档。S1是推荐首个本地实现切片；当前未创建代码、schema或候选模板，未执行DB查询/迁移/报告生成/部署。

## 12. R2运行接入前保护与API合同（提案，未实现）

### 12.1 旧链保护不是删除保护一项

依据[14项代码风险](management-traits-code-gap-analysis-20261001.md#L163)，采用明确paper/exam迁移标记判定作用域。旧接口探测失败不得回退legacy；schema缺失只有明确未纳入新版实体才兼容，部分结构安装失败关闭。

| 门禁 | 新版/迁移范围必须处理 | 不扩大本轮的边界 |
|---|---|---|
| 公开入口/后台权限 | 精确method+path匿名列表；paper paging/save/delete必须后台鉴权；新参与者令牌purpose/participant/exam/paper/expiry验证 | 不以此文档授权全仓修复；共享中间件的旧001/003影响需另列调用清单并回归 |
| 人员与历史关联 | 新participant创建/恢复同事务保留旧paper/end/pdf；拒绝旧Save/Update改新版绑定或凭据；歧义归属失败关闭 | 不能把原匿名tester PUT视为安全基础；明文tester密码的迁移/强制重置需独立设计及授权，不复制后台bcrypt假称已有 |
| 答案输入 | 每次构造新数组、恰一冻结选项、raw1～5；拒绝旧FillAnswer对新版写；前端等待在途保存，不先写end_time | 保留旧评分函数尺度；修复旧整页通用缺陷若影响其他产品须单独RED→GREEN |
| 版本/状态 | 新冻结profile及paper受旧Exam.Save/State、Paper.Save/删除及关联变更保护；状态统一按实际0启用/1禁用/2就绪/3过期解析 | 不根据旧Login注释猜状态；数据库实际状态与时段核验完成前不启用 |
| 文件与报告 | 旧双生成/双上传/ShowPdf/下载/ZIP/直接删除及压缩全覆盖；新路径随机唯一、不含原始姓名路径；先完成后计算SHA，严格Rel包含+符号链接/打开边界 | 直接报告删除和ZIP有JWT但还需业务权限，不误称匿名；路径登记不是请求者授权；新文件绝不先删旧 |
| 输出完整性 | completed run140合法答题、DTO/批准文本/六图独立门禁；禁止接受incomplete渲染为成功；双Excel统一权限/字段披露/错误处理 | 不把PDF大小或生成成功当图表/文案完整证据；不把旧模板回退当新版导出 |

### 12.2 拟新增method、权限与最小请求

统一前缀`/exam/api/management-traits`，后台路径不加入匿名前缀。所有ID经归属和作用域检查，业务错误不回显DB/路径/密码。权限名为新合同候选，menu/角色映射未创建。

| Method/后缀 | 调用者与拟权限 | 关键请求/响应 |
|---|---|---|
| POST participant/create-paper | 精确POST豁免后台JWT，强制已验证participant令牌 | examId/participantToken；事务恢复/创建paper，返回绑定paperToken |
| POST participant/paper-detail | 精确POST豁免后台JWT，强制paperToken | 冻结题目/选择/可信deadline；不返回方向/分值/密码 |
| POST participant/fill-answer | 精确POST豁免后台JWT，强制paperToken | paperQuestionId/optionId；单选保存，时限/锁/状态先于写入 |
| POST participant/submit | 精确POST豁免后台JWT，强制paperToken | manual；返回完成/incomplete或未答拒绝，不在响应自动生成报告 |
| POST results/list、GET results/detail | JWT+management-traits:result:read及测评范围 | list显式exam/分页上限；detail显式runId，不返回服务器路径 |
| POST reports/generate | JWT+management-traits:report:generate及测评范围 | runId/contentPackageId/templateRevisionId/force；冻结manifest，返回reportId/status |
| GET reports/list、GET reports/download | JWT+management-traits:report:read及测评范围 | 列表paperId、下载reportId；鉴权提供PDF，不接filesystem路径 |
| POST reports/select-current | JWT+management-traits:report:select及测评范围 | reportId/expectedSelectionEpoch，条件切换/冲突拒绝，审计历史选择 |
| POST reports/batch-download | JWT+management-traits:report:read及全部测评范围 | 固定reportIds≤100，全量预检后ZIP，不在循环里重新选current |
| POST exports/results | JWT+management-traits:export及测评范围 | runIds及确认的columns合同；无对应run不回退旧分/旧脱敏策略 |
| GET template/info、GET template/download | JWT+management-traits:template:read | 读revisionId，返回元数据/模板，不接任意路径 |
| POST template/candidate | JWT+management-traits:template:write | 上传返回候选SHA/合同及预览状态，不自动激活；read权限不能上传 |
| POST template/activate | JWT+management-traits:template:activate | candidateId/expectedActiveRevision；实际Word/目标LO门禁合格才切活动指针 |
| POST results/recompute | 管理员专权+staging+paper白名单+历史证据审核 | paperId/evidenceId/scoring+norm版本；不推测题本，不更新旧结果 |

仅上述四个参与者精确POST豁免后台JWT，handler仍强制对应purpose/身份绑定令牌；同路径其他method、后缀及后台路径不豁免。正式参与者报告读入口暂不开放；若需沿showPdf设置提供读取，须单独明确身份权限和精确路由，不共用后台下载匿名放行。管理员ID=1/全局权限可在同一授权helper明确定义兼容，其余不凭JWT登录即放行；资源范围核验不由权限字符串替代。

### 12.3 资源与验证约束

资源提案必须配置注入并在实现切片测试后定稿：分页默认20/上限200；批量下载100；DOCX20MiB/解包64MiB；复用通用转换器PDF50MiB上限。生成总预算涵盖排队+填充+转换+文件/登记，候选基线120秒，排队占用同一deadline；任何子阶段不得重置预算，取消后不切current。批量生成不作为一个长同步请求，前端按冻结目标逐项发起并展示失败；未批准后台队列/自动报告Worker不在此处新增。ZIP发送前预检并生成受限临时文件，磁盘总量和超时也配置限制，超限整体拒绝，finally清理本次临时物。

测试证明前不承诺上述数值可支撑容量。文件下载审计记录提供开始/服务端结束，不声称客户端已保存。传输TLS、静态alias、密钥管理、tester凭据升级、备份恢复是启用前环境/安全核验，不以本地代码和文档完成替代。

## 13. R2四项设计冲突的闭环与仍待确认事项

| 原阻断 | R2设计处理 | 必须取得的实施证据 |
|---|---|---|
| content/template进入评分hash | §3/6评分manifest与报告渲染hash分离，内容/模板变化只建revision | 相同评分输入更换content/template不建run；更换scoring/norm不覆盖旧run |
| exam唯一profile约束历史paper | §4/6新组卷profile与逐paper历史证据分离；capture可先独立保护不伪造评分快照 | 同exam两历史paper不同证据分别可审核；无140题证据仍可保旧PDF但拒重算 |
| paper.user_id101身份假设 | §4/6真实paper+exam双表唯一归属，新卷事务绑定participant快照，拒歧义 | 零/多/跨exam/跨类型关联拒绝；合法candidate/tester无需与101相等 |
| 多run/current消费目标不清 | §9显式run/report及批量manifest冻结、selection_epoch；§12选择端点和权限 | UI/PDF/Excel同目标，current变化不改已启动任务，旧capture不冒充新版 |

**“闭环”只指本稿已提出明确可测试合同，不代表用户已批准或代码已修复。**四项原矛盾不再作为正文的默认行为。仍待确认：字段/多段/固定定义合同，模块展示顺序/等级/比较，最新成功current策略，时段/到期/离线结算，采集必填/选填、参与者报告权限、最终名称/页数、内容双批准、历史样本及环境/Schema执行授权。R2不得未经这些确认直接作为S2/S5启用依据。

## 14. S2A本地模型合同（已实现声明，未创建数据库结构）

用户确认S2A仅本地结构与测试。本轮新增[七个模型](../Go-based%20Refactored%20System/internal/model/management_traits.go)，不生成或执行DDL，不建立仓储/writer，不接自动迁移、HTTP或S1→存储适配。§6的完整报告层仍是提案，不因七个Go类型存在成为真实数据库结构。

| 模型 / 拟表 | 已声明的最小存储合同 | 未执行的约束 |
|---|---|---|
| ManagementTraitsDefinitionBundle / el_mng_definition_bundle | 四评分轴、题本逻辑身份、longtext评分manifest及char(64)SHA；四版本有序唯一索引 | 未导入/批准题本，无hash/JSON/题数运行校验，版本必须题本作用域化留后续验证 |
| ManagementTraitsExamProfile / el_mng_exam_profile | exam主键、bundle/映射快照及SHA、字段采集合同、时长及nullable冻结时间 | profile只供新组卷；时段/冻结不可变、父表关联未执行 |
| ManagementTraitsPaperSnapshot / el_mng_paper_snapshot | paper主键及paper+exam候选唯一键；nullable ProfileExamID；独立bundle/历史证据与映射JSON/SHA、真实participant身份和人员字段/来源/时点、nullable截止时间 | 历史可无profile；不可把捕获当前身份冒充提交快照；证据审核/唯一真实归属未实现 |
| ManagementTraitsPaperQuestionSnapshot / el_mng_paper_question_snapshot | V号与display_order独立；题干/选项、方向、评分hash；paper+V/paperQuestion/order三个唯一声明；选择/raw/final/提交时间可空 | 同paper题目关联、140唯一连续/5项/1～5/合法单选及一次冻结提交未实现 |
| ManagementTraitsResultRun / el_mng_result_run | paper+scoring+norm唯一声明；id+paper+exam复合候选键；四评分轴、输入hash、participant、来源/状态/完成计数；nullable总体score/norm/level、用时秒/提交时间 | 没有content/template轴；合法已存在run只读复用/损坏拒绝及事务不可变未实现 |
| ManagementTraitsResultDimension / el_mng_result_dimension | run+维度及run+order唯一声明；整数sum/count精确源、nullable decimal score/norm及level | 13维身份、子项与run一致性、缓存由有理数重建未实现 |
| ManagementTraitsResultModule / el_mng_result_module | run+模块及run+order唯一声明；成员数和nullable decimal score | 恰4模块/成员聚合未执行；不新增未确认模块等级/常模比较字段 |

所有ID/版本显式varchar(64)，hash为char(64)，快照为longtext，成绩为`*decimal.Decimal`/decimal(18,6)，无float分值。JSON可空字段不omitempty，能区分NULL、零、未知用时。Snapshot/Run保存原始事实身份，模型中没有UserID101假设，不包含关系对象或自动hook。只保留字面量TableName方法。

11个有序复合唯一索引由GORM schema.Parse/ParseIndexes本地解析验证，**没有任何索引或FK被安装到MySQL**。没有数据库签名/collation证据，也没有实体不可变性/数据合法性生效证据；仅靠struct不能实现这些保证。

### 后续Schema/运行接入门禁

1. 显式MySQL5.7+DDL需另立切片；真实父表ID长度/字符集/collation核验后才能定稿，不能按本地varchar标签推断兼容现库。
2. RESTRICT及复合FK列序、父候选键、所需查询索引、NOT NULL/默认值/状态长度、MySQL索引字节限制仍待DDL设计；本轮未加这些运行声明以避免冒称生效。
3. DDL应用前须先实施旧Save/Update/删除/双PDF写保护；capture保护标记独立且尚未实现，不要求历史PDF先具140题评分快照。
4. hash规范化、source/status/identity_source白名单、选项JSON、时长/身份/来源验证、完整140+13+4和缓存复核均是后续入口/writer责任；S2A不新增状态机或writer。
5. 后续结果写入应直接从精确有理数生成六位decimal缓存，保存维度sum/count重建源，不能把decimal截断后用于等级/排序。未知与未答值的NULL语义需在真实首次/重复迁移、事务回滚和并发测试关闭。

本轮RED为七模型未定义导致model编译失败；GREEN九个顶层合同测试全部通过，Go全量及Windows构建退出码0，独立代码审阅PASS。编辑器暂未发现测试，使用Go原生命令验证，不将发现失败误写为测试通过。未访问数据库、未改旧模型、未建报告/内容表、未部署。

## 15. S2B纯校验与规范hash合同（本地已实现，未接运行入口）

新增[纯合同实现](../Go-based%20Refactored%20System/internal/service/management_traits_contract.go)及[独立测试](../Go-based%20Refactored%20System/internal/service/management_traits_contract_test.go)。S2B只接收强类型Go输入，不读取附件/数据库、不解析模型中的longtext、不生成批准包或持久化记录。测试题干与版本均为明确synthetic fixture，不冒充客户正式题本。

| 边界 | 已执行校验 | 规范化内容 |
|---|---|---|
| CanonicalManagementTraitsManifest | staff/leader；四版本小写语法且≤64字节；140唯一V1～140，逐题维度/方向与S1目录相同；UTF-8非空题干；五项raw1～5及展示序唯一，文字按raw明确为不符合/不太符合/一般/比较符合/很符合 | domain=`mng-scoring-manifest-v1`；V升序，option按raw升序但保留DisplayOrder，题干原文及空白保留；包含固定13维/精确常模/公式/聚合/等级/顶底选择策略；不含内容/模板/审批/时间 |
| CanonicalManagementTraitsMapping | manifest重新校验并校验确切SHA及题本身份；140逐V题干/选项raw/文字/展示序精确匹配；源题ID与源选项ID各自在实体类型内全局唯一、非空UTF-8/≤64字节/无首尾空白或控制字符 | domain=`mng-source-mapping-v1`；V/raw升序，冻结来源ID与原文，引用manifest SHA；不查询实际来源行 |
| CanonicalManagementTraitsInput | mapping重新校验并核验确切manifest/mapping SHA、题本身份、paper/exam/participant ID及candidate/tester类型；140唯一V/paperQuestion/displayOrder，逐V源题匹配；已答必须同题选项ID与raw一致，未答只允许空选择+raw0 | domain=`mng-score-input-v1`，逐题`mng-frozen-question-v1`SHA；输入包含四评分轴/manifest/mapping SHA/纸卷和人员声明/原始选择；输出V升序S1原始答案，不做第二次反向或计算分数 |

JSON使用固定struct字段顺序、标准encoding/json转义、无缩进/末尾换行；SHA为这些确切UTF-8字节的SHA-256小写64位hex。只重排数组副本，不改调用者输入；返回JSON/答案切片为独立数据。任何校验失败返回零结果，不提供部分合法hash。版本合法语法**不等于受支持、来源有效或已批准**；没有硬编码激活版本。

DisplayOrder不是V号；固定模块数组顺序仅S1计算合同，不视为已批准报告排序。逐题SHA引用整个mapping SHA，故任一映射变化会传播到各冻结题hash及输入hash，这是此v1合同的显式作用域；它不证明数据库外键或历史证据。呈现人员姓名/地址、内容版本、Word样式/模板SHA、捕获时间不进入评分输入。实际题本题干修订则进入manifest并需独立题本版本审核。

后续必须验证：客户题本逐题数据库对照、历史证据与真实participant归属、版本执行白名单/批准、JSON重复键/未知字段/限额解码、S2A字段与canonical数据的安全适配、输入hash与合法既有run复用、真实事务和删除/旧写保护。当前只验证调用者声明自洽，不授权其身份；不能用synthetic hash覆盖客户材料原SHA。

RED为测试先建时契约符号未定义；GREEN10顶层测试、含子项172通过/0失败，S1+S2B联合、Go全量与Windows server构建退出码均0，独立CodeReviewer PASS。无SQL/DB/HTTP/题本导入/历史重算/报告或部署。本轮新文件实际Go内容格式已对齐；CRLF保留，未将原始gofmt -l写作通过。

## 16. S2C严格存储解码与只读模型适配（本地已实现）

新增[严格解码及适配实现](../Go-based%20Refactored%20System/internal/service/management_traits_decode.go)和[测试](../Go-based%20Refactored%20System/internal/service/management_traits_decode_test.go)。只消费调用者提供的模型/字节，不查询或写数据库，不建立writer/运行接口，不返回正式分数。

| API | 已验证合同 | 明确不证明的事项 |
|---|---|---|
| DecodeManagementTraitsManifest | 仅接受S2B规范manifest存储结构及domain；逐层精确JSON tag及全部必需字段（含false）；重建S2B后核对固定维度/常模/policy及schema | 原始public manifest JSON不是存储结构；合法版本语法不等于执行支持/批准，题干不是历史来源证明 |
| DecodeManagementTraitsMapping | 仅接受S2B规范mapping存储结构/domain；严格字段/类型及确切manifest关联；再执行S2B逐V选项和来源ID校验 | 不查询源题、不核验历史备份、actual DB唯一归属 |
| ValidateManagementTraitsStoredInput | bundle ID/四版本/题本及确切评分SHA；paper bundle/manifest/mapping SHA及可空同exam profile声明；140模型行的paper/ID/V、方向/题干/选项、选择raw/final及逐题评分SHA；返回S2B校验输入和S1原始答案 | 不校验Source/Status、EvidenceSnapshot/SHA、IdentitySource、人员资料/采集合同JSON、所有时间/提交元数据；这些不被默认为已审核，亦不决定完成/到期/权限 |

严格解码拒绝：重复字段（包括转义后同名）、未知字段、缺字段、大小写别名、任意null、错误root/值类型、整数小数或指数写法/溢出、尾随对象/垃圾、非法UTF-8、BOM、孤立UTF-16代理项、过深输入。合法代理对与真实U+FFFD可接受，原文保留，不静默替换坏字符。出错统一返回不含原始JSON/人员ID/秘密的错误及零结果。私有schema reader只用于本次明确的struct/选项数组，不作为公共通用JSON框架。

允许对象字段顺序/排版空白、题目及选项数组重排；SHA仍基于S2B重新生成的规范字节，不基于带排版空白的存储字符串。固定规范维度/items/policy顺序不能随意更改。接收字节与模型数据不被修改，输出独立；没有把模型数据在原地排序。

预算由调用者显式提供正数`maxJSONBytes`：独立decode按字段字节限制；模型适配同时检查manifest+mapping+140个options JSON的**合计预算**，以剩余量减法预检防溢出。当前已有字符串的分配及上游数据库/HTTP读取限额不在此函数保证范围，未来入口需在读取前限额；预算数字未作为全局配置默认启用。

已答语义：SelectedOptionID/RawAnswer/FinalScore三个指针齐全，raw1～5及同题合法选项，final必须等于raw或6−raw（按冻结方向一次处理）。未答三个指针全nil；混合nil、零raw或空已答选择拒绝。SubmittedAt是否为空不用于推定已答或提交状态，时间门禁延期。ProfileExamID非空只核验其等于paper.ExamID，不加载profile，也不以user_id101核人。逐题SHA重建包含完整mapping SHA，沿S2B已有作用域。

剩余接入门禁：真实paper/exam/participant和源题加载与授权、历史证据/身份来源审核、执行版本/审批、时间/状态/采集合同及人员资料合法性、合法已存在run复用与持久化缓存复核、旧Save/Update/删除/PDF写保护、真实DDL/FK/collation和事务。S2C通过不可当作正式重算/报告许可。

RED缺三个新API退出1；GREEN九顶层/含子项365通过0失败，独立CodeReviewer PASS，主代理Go全量/S1～S2C联合测试及Windows构建退出码均0。全量5个既有环境用例SKIP（真实MySQL并发、客户模板上传、FB-169/170活动模板、客户模板LibreOffice页数），未作为通过证据；S2C无SKIP。两新文件无BOM，Go内容与gofmt一致但CRLF保留。未修改S1/S2A/S2B或旧业务入口、客户原件、SQL或运行环境。

## 17. S2D只读结果一致性复核（本地已实现）

新增[结果复核器](../Go-based%20Refactored%20System/internal/service/management_traits_result_validation.go#L19)及[测试](../Go-based%20Refactored%20System/internal/service/management_traits_result_validation_test.go#L203)。输入为调用者提供的bundle/paper/140题快照/run/13维/4模块及JSON预算；不读写DB、不建立writer、不修正损坏结果，不改变既有评分或运行入口。

`ValidateManagementTraitsStoredResult`先要求run.Status=completed且子项13/4齐全，再调用S2C验证存储输入、S1从140原始答案重建精确分数。任一合法但未答输入仍拒绝作为完整结果；成功返回独立的精确S1结果，不返回数据库模型缓存作为计分源。

| 对象 | 逐项复核 |
|---|---|
| run | 合法ID、paper/exam/participant类型与ID、题本、四评分轴、评分manifest SHA及重新生成的input SHA、140/140计数、overall精确等级/六位分值缓存/常模缓存 |
| 13维 | 每行合法且实体内唯一ID、同run、恰13个唯一固定key、key对应order/name/module/count/answered/sum、精确等级及score/norm六位缓存 |
| 4模块 | 每行合法且实体内唯一ID、同run、恰4个唯一固定key及计算order、成员数3/4/3/3、成员维度等权聚合的六位缓存 |

缓存比较复用已存在`decimalFromRat`，从原始答案得到的有理数直接按6位构造decimal，以数值Equal比较。允许等值decimal的不同指数/末尾零，拒绝不同数值、缺值、额外有效精度、两位常模54.23冒充六位54.230769。原始整数sum/count也须与答案重建完全一致，不能仅让存储sum与其自身缓存互相自洽。等级/顶底选择来自精确S1值，不由六位缓存再聚合/舍入/排序。

允许questions/dimensions/modules数组重排，不修改调用者结构/指针；每次返回新Rat/切片，调用者修改结果不污染未来验证。ID唯一范围是各子表自身，不要求维度/模块两类行ID跨表不同。失败统一零结果及不含人员/题干/数据库细节的错误，无legacy回退或Save修复。

**completed字符串及自洽校验不证明真实交卷或授权。**run.Source/CreatedAt/SubmittedAt/UserTimeSeconds及S2C明确延期的证据/人员/采集/审批/时限元数据仍不校验；调用者必须另做真实归属、来源/状态/时间/执行版本批准及事务门禁，不能仅调用此函数启用正式报告。输入hash匹配同样不证明历史题本证据成立。

RED先建测试时新函数undefined退出1；GREEN原生Go事件为8顶层及其子项575次pass（其中567子项），0失败/0跳过；编辑器显示484通过/0失败，二者计数口径分别记录不混用。主代理全量Go事件1918 pass/0 fail/5既有环境skip，TEST_EXIT=0 BUILD_EXIT=0，独立CodeReviewer PASS；5项skip与S2C记录一致，不计环境通过。本轮没有数据库/HTTP/历史重算/报告生成/部署。

## 18. 生命周期与来源门禁收敛（设计提案，未实现/激活）

本节延续用户“不修数据库、不改变程序逻辑”的决定，只定义后续运行前的条件，不新建状态机代码、字段、DDL或配置。依据[实际题本及处理决定](management-traits-staging-questionnaire-verification-20261001.md#L78)。S1～S2D是算术/数据自洽验证，不能单独构成以下来源/权限/生命周期门禁。

### 18.1 已知文本基线与执行范围

当前staging基线：00201对应客户基层原文140/140；00202对应客户管理原文138/140，V67/V96仍为基层措辞。保留源库，禁止按“算术一样”隐式采用客户两项文本，也禁止删除两项文本匹配检查。

未来新题本有两条**待选择路径**，本节不替用户作选择：
- 现库存量原文作为独立评分题本来源，逐字冻结并记录当前捕获时点；即使身份为leader，也不能把该manifest命名或描述为客户管理修订稿全文。
- 客户管理修订稿作为独立题本来源，待V67定稿后建立新来源/映射和冻结方式，不能宣称现库已匹配；不覆盖旧源题是既定边界。

来源选择将影响question版本/manifest SHA/映射，不影响已确认的固定评分公式。两种文本不能共用同一不可变题本版本或hash。当前源库SELECT指纹只证明当次当前数据，不证明作答时文本、客户批准或源ID永续不变。K6材料标签错字不按模糊名称映射维度，稳定维度身份继续依据已确认目录。

### 18.2 四类门禁职责分离

| 门禁 | 输入事实与核验责任 | 失败行为 |
|---|---|---|
| 数据自洽 | S2B/C：完整题本/映射/原始选择与hash；S2D：完整run及精确事实/缓存 | 拒绝，不修复缓存、不回退legacy、不追加“通过”状态 |
| 真实来源与身份 | 经可信数据库读取的paper.exam_id、有效candidate/tester纸卷+测评关联、原题/选项；历史须另有作答时期来源证据 | 零/多/跨exam/跨类型歧义拒绝；不用101、ID长度、stuFlag或candidate优先兜底 |
| 执行策略与状态 | 精确评分版本执行支持、题本来源选择及审核、schema可用、旧写保护、调用者权限、锁内state/deadline | 未知/缺项/撤销/部分安装失败关闭；不凭版本语法或字符串completed放行 |
| 正式内容与呈现 | 指定run及其来源审核、具名内容/测量批准与环境、确切内容/模板revision、六图及输出完整性 | 不生成/激活正式报告，旧completed PDF保留；审核撤销不删除历史文件 |

可信DB读取须由未来仓储/加载器在一致事务内执行参数化查询，检查所有查询错误并记录同次读的证据版本；用户传入一个“verified=true”或合法SHA不能代替此过程。新paper身份在组卷时原子绑定，后续读一致性核验冻结身份及有效关联，人员更新/解除关联需先受旧写守卫约束。人员合法性/采集字段与脱敏规则另定合同，不以评分hash记录所有个人资料。

### 18.3 状态与来源合同提案

下表是逻辑状态，不强制立刻增加物理列或将多个对象状态塞入一个status字段。题本版本永久不可变；审核撤销/退役属授权元数据变化，需独立记录，不改评分manifest正文或SHA。

| 对象 | 逻辑过程 | 可执行条件 / 不可执行边界 |
|---|---|---|
| 题本包 | draft → reviewed；reviewed → retired | reviewed须精确版本/SHA与题本来源审阅证据；只允许指定版本进入新组卷，不因hash有效自动review。retired默认禁止新组卷，历史读取/重渲染是否准许须显式审核策略，不跨版回退 |
| 新组卷profile | draft → frozen | 冻结评分包/140题映射/采集合同/时限，在事务内复核源库；frozen后不原地换题本或改变个人deadline规则。需变更则独立新配置/测评，不重绑已有paper |
| 新paper | created → answering → completed 或 incomplete | 创建必须锁有效participant/profile并防重复；保存/提交锁同paper，完成状态不可由旧Paper.Save或人员EndTime单独写。140合法答题才有completed正式结果；incomplete保留审计，不生成正式分或报告 |
| 历史评分证据 | proposed → reviewed 或 rejected | reviewed覆盖指定paper作答时期140题、题ID/V/选项分值及真实人员归属，绑定证据SHA/审核主体/时间/范围。当前源库匹配或旧PDF捕获不能单独使其reviewed |
| 结果run | 新事务创建完整completed或审计incomplete；completed不可变 | 结果及13+4子项同事务，与paper/人员完成事实一致；生成中/失败尝试不作为正式run复用。损坏run拒绝而不是原地Save“纠正” |
| 历史PDF捕获 | protected → draining → captured；失败保留protected | 捕获可在评分证据不足时先保护旧PDF；排空已在途生成/上传/压缩后核字节/SHA。failed不删除原件、不解除保护去恢复旧覆盖行为；恢复/解除由单独维护决定 |

run.Source提案只区分`submission`与`historical_recompute`，paper.Source只区分`new_creation`与`historical_evidence`；不同对象的枚举不能混用。来源合法值不证明事件实际发生：submission核锁内提交事实，historical_recompute还核环境/白名单/证据授权。合法completed run可以被不同合法请求复用且原source不变，不为把历史重算请求写成新提交而覆盖source。

IdentitySource提案为`submitted_snapshot`与`captured_at_recompute`：前者要有提交时点人员快照来源；后者明确是重算时捕获，不能仅因为当前人员恰一关联就冒称提交时字段。历史证据不能提供原时点人员信息时，正式报告是否可以使用后捕获身份须单独审批，评分自洽通过不自动同意。

题本**退役与审核撤销分开**：retired禁止新组卷，不自动等同评分事实被否定；已开卷paper遇题本退役时，继续按原冻结题本续答/交卷还是暂停，仍须确认。审核撤销若涉及题本可信性，则已开卷续答/提交、已有run再使用及新报告生成分别核显式策略，不能把停止新生成直接解释为删除历史或允许继续答题。状态处理必须读取冻结绑定及可信授权元数据，不能用最新活动包替换已开卷题本。该分支未确认前新版运行接入保持关闭，现有旧测评行为本轮不变。

### 18.4 时间与交卷条件（仍待业务确认，不启用）

采用服务器可信时钟和冻结deadline；前端只显示剩余时间，20分钟提示对应冻结25分钟deadline前5分钟。计时精度、用时整秒/舍入、测评窗口对已开卷优先关系、截止瞬间保存顺序、离线Worker/重启结算延迟均需定稿，不在此凭猜测写代码。

建议锁内语义沿§7：未到期140完整可提交，未到期缺答拒绝且不写人员end_time；到期由服务器判定为timeout，完整140可完成、缺答写incomplete审计而非正式分。参与者不能自行声明timeout提前绕过必答。到期保存拒绝新答案；若先持有锁的合法保存与提交竞争，后取得锁的操作必须重新读取时间/状态，不沿用锁前判断。

已完成请求重试须核原提交元数据和S2D，再只读返回；不能用本次请求时间重新生成时长/SubmittedAt。真实deadline、SubmittedAt、UserTimeSeconds应具备可解释同源时间合同；当前S2D明确未核这些字段，后续门禁必须补上，不能用已有函数名推定时间已验证。

### 18.5 结果复用与历史路径

| 情况 | 后续允许行为提案 |
|---|---|
| 同paper/scoring/norm，原始输入/身份/题本及完整结果均一致 | 生命周期/来源策略及授权通过后调用S2D只读复用，零UPDATE/INSERT/DELETE，原source/时间保持 |
| 同唯一键但input hash、题本、身份、子项或缓存不同 | 返回明确冲突/损坏，不覆盖；纠错重测或新运行身份须另审 |
| 仅content/template变化 | 复用原run建立独立report revision，不创建评分run、不换旧结果事实 |
| scoring/norm变化 | 只有显式支持及证据/授权满足才创建新并行run，旧run/PDF只读保留；不是该对象任意未来版本均可执行 |
| 历史证据不足，但旧PDF可访问 | 可按另批capture策略保存旧PDF副本；拒绝新版评分/正式报告，不以当前源库或PDF逆推原始答案/题干 |
| 历史run本次data自洽通过，但来源审核或批准撤销 | 保留行及文件；阻止需要该授权的重算/新生成/激活，历史授权下载另按保留/隐私政策判断，不物理删除或伪装为缺run回退 |

历史流程必须是指定staging白名单paper及显式评分目标，无白名单/独立历史证据即拒绝。来源审核材料至少包含paper/exam、question版本/确切题干与选项、140显式源ID/V映射、证据来源及取得时点、人员来源、审核主体/决定及SHA；**这些为未来证据合同需求，不是现有EvidenceSnapshot已经有这些内容**。JSON字段协议、签名/审计载体、复核人及授权范围仍须审阅，当前模型的longtext不自动执行批准。

### 18.6 后续最小实施顺序与确认清单

1. 确认确切题本来源选择、源库不修改边界及历史/新卷范围；未选来源不生成可执行manifest种子或自动忽略两题。
2. 定稿最小来源/身份/采集/时间/批准载体合同及其测试矩阵，只将相应条件注入后续门禁；不继续给每个未决策字段随意默认值。
3. 确认旧写/删除/PDF覆盖保护的切片及影响清单，RED→GREEN后才能进入真实Schema安装；不得因本地七模型可编译执行AutoMigrate。
4. 明确MySQL幂等DDL、RESTRICT/FK/collation、可信加载/事务writer及失败回滚/双连接并发验收；执行数据库迁移或staging运行写验证均需单独授权。
5. 正式内容/Word合同与批准独立推进，未满足六图/模板位置/环境/具名批准就只保留draft，不生成正式报告。

待一次性确认：新测评采用现库原文还是客户修订稿；冻结后配置变更方式；已开卷遇退役/审核撤销的续答/交卷及retired历史使用/下载策略；历史人员后捕获使用；到期/窗口/用时与离线结算；来源审核载体/主体及范围；独立旧写保护实施范围。选择尚未取得，本节全部维持提案。本次只修改文档，未改程序逻辑、源题库或运行环境。