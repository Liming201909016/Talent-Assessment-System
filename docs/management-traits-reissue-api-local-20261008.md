# 002 独立报告后端切片 — 2026-10-08

## 1. 未验与边界（先列）

- **未部署、未执行新DDL、未生成或下载原实名报告**。staging正常管理员对新接口的真实200、真实MySQL首次/重复安装与跨连接竞争未验；本轮HTTP200为本地httptest＋sqlmock。管理员UI、草稿、formal审批、Worker观测和整560UI均未推进；生产操作0。
- **两旧140题完成卷仍不能重发**：缺原V→题干/维度映射、原身份完成快照及可信历史载体。旧表实际没有题干/V字段；已知备份目录maxdepth2中、2026-04-22之前的SQL gzip计数0。这不是全服务器/所有外部归档穷尽证明，不无限扩查、不拿当前题库补证。
- 本地PDF使用**真实答案＋合成身份和合成题本身份**；不是原实名完整报告，不把合成mapping/identity SHA冒充服务器原SHA。原客户Word桌面、所有PDF几何/像素、staging LO及独立CodeReviewer未验。本模式不派发agent，不自审冒独立PASS。
- 存量Ed25519历史审计适配器保持不变，未配置真实审计根或接受客户来源JSON。本slice采用用户允许的**可信新冻结结果loader→既有纯评分/205文案核心**，不是伪签名或历史迁入new_creation。

## 2. 真实来源只读结论

[最终只读收据](../scripts/test/results/mng-reissue-api-local-20261008/real-source-audit-final.json)由[限定审计入口](../scripts/tools/management-traits-reissue-readonly-audit-20261008.js#L1)执行：strict/Batch/ConnectTimeout10/ConnectionAttempts1，首次SSH native0，SQL显式READ ONLY；既有sudo/socket认证在服务器内部消费，无env/密码/JWT/DSN、姓名/手机号正文输出。

| 对象 | 本轮实际事实 | 资格结论 |
|---|---|---|
| 历史exam1776822816300709851 | 3卷、2个state2各140题已答、1个state0；snapshot/run0；两完成卷各唯一candidate、tester0 | 两完成卷来源证据不足；未答卷另缺完整性；全部不重发 |
| 新exam1791298091700970647 | 1卷state2、1个completed staff run、1份new_creation/submitted_snapshot；唯一同paper/exam/participant及原提交时间candidate1、其他tester0 | 找到1份合格冻结**候选来源**，不是凭用户曾保存身份猜测完成 |
| 新卷原始答案 | 140唯一V，五项快照、每题一checked、一selectedMatches、一legacyMatches，提交时间一致；40反向一次 | 原raw/final与固定13维目录核对，完整140/140 |
| 成绩与提交事实 | 13维＋4模块＋1receipt；原时间匹配，冻结先于开卷；overall58.642639、norm54.230769、qualified | 独立Rat重建13维sum/count、六位缓存及4模块全匹配；显示58.64 |
| 原存储字节 | manifest61331bytes、mapping74362bytes、field421bytes；evidence/mapping/manifest实际SHA2均匹配原存储摘要 | 三原字节hash通过；identity仅输出原JSON摘要，不下载或mask后称原hash |
| 全库限定新版完成聚合 | staff/leader completed共1 | 不造新完整卷，不提交、重算或修改任何用户记录 |

首次元数据SQL也核实旧el_paper_qu只有ID/源题ID/题序/原答分及可空新版指针；el_paper_qu_answer只有桶标记/score/abc等，无题干/V JSON。历史两匿名paper hash与10-06审计相同；本次未下载旧PDF、未读取当前源题正文。完整新原始私有JSON没有下载至本地，故**本轮没有对真人调用新HTTP资格loader或新生成服务**，最终资格仍由运行时严格loader逐次决定。

## 3. 本地实现与API合同

已有JWT命名空间为 `/exam/api/management-traits`，只追加以下路由，不改变旧响应/匿名规则：

| 方法＋相对路径 | 参数 | 合同 |
|---|---|---|
| GET /report-reissues/qualification | 唯一runId | 真实只读loader校验后返回eligible/kind/purpose；没有迁移/写结果 |
| GET /report-reissues | 唯一paperId | 最多100条独立报告元数据，created_at DESC/id DESC；空数组[]；不读完整身份archive |
| POST /report-reissues/generate | JSON仅runId | strict未知/重复/null/空串拒；actor取正常loginUser；返回report/reused及TEST用途 |
| GET /report-reissues/view | 唯一reportId | inline PDF，先完整文件及归档校验，再审计，最后发送响应 |
| GET /report-reissues/download | 唯一reportId | attachment PDF、RFC5987、Content-Length、nosniff/no-store |

[注册与handler](../Go-based%20Refactored%20System/internal/handler/management_traits_reissue_api.go#L18)复用既有管理员口径：正user_id1或正user_id且全局权限；普通exam:list不授权。所有权限在DB前，客户端无path/assetsSHA/actor/approval/evidence输入。[唯一旧代码增量](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime.go#L43)仅RegisterRoutes追加调用；原八参与者、TEST、formal及公开入口不改。

[可信桥接](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_api.go#L44)调用既有loadTestReportData：同paper锁、唯一owner、完整140快照和旧桶交叉核、13/4/receipt/原时间、原冻结字段/身份、已存成绩重建一致；bundle共享锁检查撤销/immutable版本与manifest。不会联结当前el_qu重建历史。保存原snapshot内JSON字符串、bundle与140题原raw/final/selected快照及报告DTO，不替换原身份、不回填新结果。

[生成与归档](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_api.go#L88)：首次事务读原输入与已有同hash报告→事务外90秒Word/LO→UUID O_EXCL 0600新文件→第二次同paper锁重新读取并比较完整归档字节→metadata和generate审计同TX。UNIQUE(run_id,data_sha,template_sha,content_sha)＋paper锁幂等，同输入复用已验旧文件；另一请求先提交时仅清理本次自有候选文件，追加reuse审计。任何失败不写旧current/pdf_path、不删除旧files；旧TEST与新私有reissues子目录分离。

[ID读取](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_api.go#L128)只读已保存archive/PDF：状态、绑定、数据SHA、UUID相对key、允许私有根/符号链接、同句柄大小/%PDF/文件SHA校验；不重渲染、不读取当前人员。source撤销阻止新生成，管理员仍可读取已归档TEST PDF；不宣称formal撤销UI已实现。

[独立模型](../Go-based%20Refactored%20System/internal/model/management_traits_reissue.go#L6)新增el_mng_report_reissue（15列）及el_mng_reissue_audit（7列），两个三列复合RESTRICT FK：report→原run(id,paper_id,exam_id)，audit→report(id,paper_id,exam_id)。无额外current表。[可选完整结构门禁](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_schema.go#L38)检查全部列/NULL/type/字符集、实际父键collation、索引顺序/前缀/唯一性、FK列顺序/规则与InnoDB；缺失/畸形503仅新能力关闭，不startup失败或AutoMigrate。[独立DDL工件](../scripts/sql/management_traits_004_report_reissues.sql)只在C区，**执行0**。

新表查询最初被既有十一表ID容量守卫拒绝，专项真实失败保留。[修正](../Go-based%20Refactored%20System/internal/service/management_traits_reissue_api.go#L22)仅在新结构完整验证后，为当前操作复制原capacity并添加八个新ID预算；不修改singleton指针或旧schema合同、不跳过守卫。

## 4. 影响清单

| 消费方 | 处理及理由 |
|---|---|
| 现有RegisterRoutes、router命名空间/JWT | 同步修改仅前者追加注册；router无需修改，既有组已认证 |
| 新service/model/schema/handler及两新Go测试 | 同步实现；没有旧消费者，未来UI须另slice接新DTO |
| 原TEST Generate/Read、current/revision/audit与旧pdf_path | 无需修改：新表新文件路径、原签名/返回保持，全Go回归通过 |
| 原可信loadTestReportData、BuildManagementTraitsTestReport、205核心 | 无需修改：新桥只在其原new_creation门禁之后复用，未下调历史/评分/版本约束 |
| Ed25519 audited_reissue及原三测试/离线作品 | 无需修改：无真实历史根，原synthetic证明不冒历史真实性 |
| 002 Word、通用LO、客户原件/两运行资产 | 无需修改：仍原TEST/6图/90Tag/SHA，未去警示、未套00401评分或旧覆盖save |
| draft/formal/Worker/frontend/.env/旧SQL/Legacy | 无需修改：明确超出scope；未删除、回滚或部署已有外围代码 |

本轮未改任何公共函数签名/配置键/旧API结构/数据库字段含义。新增方法只被新handler和测试调用；新列表隐藏data_snapshot/data_sha/file_key，不输出人员身份或源原文。没有新批准记录、签钥、bootstrap制度或正式启用。

## 5. 真实验证、失败保留与作品

- 初始RED：service缺方法编译失败；真实Gin未注册路由404而期望401，native1，不冒纯编译RED为行为RED。
- 首完整专项18pass/9fail：新表被旧ID容量守卫拒绝；复制当前操作预算后47pass/native0。后追加真实HTTP测试先错误单次Read/EOF，再发现Take绑定LIMIT=1遗漏；只修测试，所有失败收据保留，未改产品迁就fixture。该测试三次故障修正后停止新增试改。
- [最终专项](../scripts/test/results/mng-reissue-api-local-20261008/focused-accepted.json)：49pass/0fail/1明确PDF opt-in skip，native0；包含两code正常/外部render失败/审计rollback/模拟另一请求先提交复用、归档篡改拒绝、结构漂移、缺表/空数组、源拒绝、正常Gin权限/strict参数及真实httptest服务器view/download200＋原字节/文件名/安全头。
- 历史/撤销/owner/损坏成绩的负向新测试只断言拒绝且不配置任何writer；部分提前拒绝用例的未消费读取后缀不作全SQL路径证据。原完整loader负向矩阵随全Go回归通过。**真实跨连接并发、所有BEGIN/COMMIT故障点及全部符号链接组合仍未验**，不把mock模拟并发当MySQL竞争实证。
- [真实匿名评分](../scripts/test/results/mng-reissue-api-local-20261008/masked-real-score.json)1pass/native0；[实际本地LO](../scripts/test/results/mng-reissue-api-local-20261008/masked-real-lo.json)1pass/native0/0skip；[独立PDF文本及原SHA收据](../scripts/test/results/mng-reissue-api-local-20261008/masked-pdf-verdict.json)：36/36所选客户原文、58.64、TEST及合成身份可提取，9页。
- **[本地实际PDF](../scripts/test/results/mng-reissue-api-local-20261008/masked-real-report.pdf)**：真实答案/合成身份，仅TEST；650342bytes，SHA **8d3622647f612c08b4fb197960266fe13067cfa3043cafb29efe080b637877bc**。不是原实名报告或线上生成200；未声称六图像素/全页无碰撞完成。
- [默认全Go](../scripts/test/results/mng-reissue-api-local-20261008/all-go-tests.json)：6359pass事件/693通过顶层/0fail/11skip/parse0/native0。27个测试开关仅子进程清除，不改用户环境；11skip完整名称在收据，包含新PDF opt-in与既有历史离线PDF/实库/模板环境项，均不计环境通过。不是覆盖率，也不是6272→6359全部来自本slice。
- [server build](../scripts/test/results/mng-reissue-api-local-20261008/build-server.json)、[all build](../scripts/test/results/mng-reissue-api-local-20261008/build-all.json)、[vet](../scripts/test/results/mng-reissue-api-local-20261008/vet.json)全部native0/stdoutstderr0；Go Build任务完成，新文件diagnostics0，审计Node语法0。前端/全560/production/Linux发行包没有重跑或发布。
- 两运行资产重新SHA核验仍content=b0498249…、TEST template=05c55e77…；未修改客户原件。原工作树大量已有改动保留，不以Git HEAD冒开工基线或声称全仓逐SHA保护。

## 6. 收口与下一步

来源核验及本地后端slice完成；真实原名PDF未出，新表未装、新接口未部署。staging backend4179/frontend52eec为前次发行事实，本轮未替换或fresh重新证明其完整运行字节。所有远端写/POSTsubmit/人员保存/重算/DDL/上传/restart/生产/git操作0；远端读取没有临时凭据文件，也没有待清理用户记录。

下一由CodeReviewer对本切片独立只读review；后续UI、staging新表安装及统一受控发布需另明确scope，不沿身份补丁或此前外围批准自动部署。历史可信归档若未来提供，另接受控服务端loader，不开放客户路径/JSON自证或给旧卷伪造snapshot。