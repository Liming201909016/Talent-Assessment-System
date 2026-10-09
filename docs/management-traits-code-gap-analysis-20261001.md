# 管理特质002：客户需求与现有代码差距评估

日期：2026-10-01。范围：00201/00202，客户V2.6开发方案为主线，结合工作簿、V2.8报告及已确认决定。**本次只做需求理解和代码评估，不实施、不查写数据库、不重算、不部署。**

## 1. 验证边界与结论

**未验证项先列：**
- 未对数据库中的两套002题本逐题核对；140题、五选项、原始分和V号是否与客户材料一致，仍未验证。
- 未运行002新版API/浏览器链、计时/并发/越权测试、真实数据库事务或目标服务器Word转换；没有新版功能通过证据。
- 本轮未重新编译或执行测试：仅新增本评估文档及更新项目记忆，未改业务代码。前轮旧评分16/16通过只证明旧公式回归。
- 定时备份实际启用、全库自动恢复、个人数据静态加密、当前远端PDF匿名暴露、微信/浏览器全兼容均未在本轮运行核验。

**第二轮完整复核补充：**三份客户原件SHA重新核验一致；本地Go构建完成，8项既有002评分/等级测试通过。直接提取整页答题的现有方法进行内存模拟，复现单选请求携带历史选项的数组累积问题；没有调用HTTP或DB。其余新增安全/事务/文件风险为源码证据，未在真实环境复现。完整证据及修正见§7～§10；上述“本轮未测试”保留为第一轮历史边界。

**总判定：旧002已有测评基础链，但不满足客户新版。此次不是“换一个Word模板”，也不需要从零重做整个系统。**应保留传统通用入口基础，新增002专属题本/评分事实和报告契约，并补齐该链的完整性、身份、计时及历史保护边界。00401已有实现仅是复用参考，不能计作002完成。

## 2. 需求理解与有效口径

### 2.1 材料依据

- [客户开发方案V2.6](260929管理特质测评-优化/260928管理特质测验开发方案修改稿V2.6.docx)：本轮只读提取全部正文/表格，共214个非空段落；36305 bytes；SHA-256为 `e306e7918f3d9b9c6c97c08dd076679c20c58c6c55891174cd7d83c8852ae923`，与前轮一致。
- [题本及条件文案工作簿](260929管理特质测评-优化/260928测评内容+数据图.xlsx)、[报告样例V2.8](260929管理特质测评-优化/260928管理潜质测评报告模板修改稿V2.8.docx)：前轮已完成材料交叉核对，本轮不重复转换或改写。
- [已确认规则](management-traits-word-assessment-20261001.md#L287)优先于客户文件中被明确纠正的条款；[设计草案](management-traits-word-design-20261001.md)不是现有能力，也不是已通过的实施设计。

### 2.2 核心业务目标

1. 两套140题问卷分别保留；原始选项1～5，按13维正反向公式计算。
2. 维度百分制为 `25×精确均分−25`；总体为13维等权平均；四模块按所属维度等权平均，不能以四模块简单平均替代总体。
3. 全程精确，最终HALF_UP两位；等级与最高/最低排序使用未舍入值；五档边界为90/70/30/10。
4. 固定常模，学习53.75、创新50、总体精确705/13；常模辅助比较不等于真实百分位。
5. 最高3项、最低3项，按已确认模板顺序截断同分，全同分允许重叠；动态摘要/评价/建议按Excel等级列精确匹配。
6. Word负责可维护样式，系统负责值；交付PDF，不新增结果DOCX下载；六张业务图必须真实可见。
7. 新版正式报告仅140/140；保留旧PDF。历史新版重算仅限将来指定的staging样本，不自动覆盖或批量执行。
8. 25分钟沿方案；20分钟剩5分钟提示是方案需求，但当前未实现。

### 2.3 不得扩大实施范围

团体报告、千人扩容、矛盾维度预警已明确不含本轮；超过35分钟提示延期。自动生成报告触发仍待选择。SPSS/SAS、全库自动备份恢复、任意公式编辑、自定义导出虽在方案内，但是否纳入本次002个人报告改造尚未单独确认，列为完整方案差距，不自动承诺实现。

未关闭的材料事项：管理版V67、建议错字/频率冲突、模块等级/比较规则、固定统计措辞、最终名称/页数、具名内容与测量批准。不得从00401复制区间或把样例错误固化为验收值。

## 3. 需求—代码差距矩阵

状态含义：**已有基础**仅指静态链路存在；**部分**指缺少客户要求；**缺少新版**指查读的002链未接入该能力；**待确认/排除**不纳入当前交付。没有用完成率百分比掩盖关键链路差距。

### 3.1 配置、入口与身份

| ID | 客户需求 | 当前实现证据 | 状态与差距 |
|---|---|---|---|
| MT-01 | 两题本与测评对象独立 | [旧13维公式](../Go-based%20Refactored%20System/internal/handler/tester_score.go#L137-L181)、[002路由](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L236-L249) | 部分：有002题库分流，无专属冻结题本/评分/内容/模板版本链；不能以开放/封闭或人员ID长度代表题本身份。 |
| MT-02 | 姓名、年龄、性别、手机号、单位、岗位必填/选填 | [配置字段](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L184-L202)、[前端规则](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L166-L186)、[后端保存](../Go-based%20Refactored%20System/internal/handler/candidate.go#L61-L100) | 部分：目前勾选=显示且必填，未勾选=隐藏，不是“显示但选填”；后端无条件要求姓名/手机号，未按配置校验其余字段。 |
| MT-03 | 手机号各号段识别 | [开放页验证](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/candidate.vue#L142-L159)、[后端入口](../Go-based%20Refactored%20System/internal/handler/candidate.go#L94-L100) | 部分：前端接受1开头11位，无号段枚举；后端只有非空校验，不能据此认定格式契约完整。 |
| MT-04 | 封闭导入、身份证或手机号验证、后4位默认密码 | [创建身份](../Go-based%20Refactored%20System/internal/handler/tester.go#L215-L242)、[身份查找](../Go-based%20Refactored%20System/internal/handler/tester.go#L463-L492)、[导入](../Go-based%20Refactored%20System/internal/handler/tester_excel.go#L183-L225) | 已有基础：两身份支持、默认优先手机号后4位；需统一导入说明/真实必填与格式校验。登录标签写手机号但后端兼容身份证，需交互验收。 |
| MT-05 | 手机/微信二维码、电脑链接 | [链接与码生成](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/list.vue#L199-L246)、[中转页](../Go-based%20Refactored%20System/ruoyi-ui/public/exam-entry.html#L20-L47) | 已有基础：无需安装、开放/封闭路由存在；本轮未扫码或验证002实际配置。 |
| MT-06 | 二维码美化及量表变化自动更新 | [二维码生成](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/list.vue#L199-L236) | 部分：固定黑白码，打开时按测评记录生成；同一exam链接可保持稳定，不应把“更新码”误设计成每改内容换URL。缺自定义美化和版本/失效规则。 |
| MT-07 | 测评时段及25分钟 | [表单](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L168-L221)、[组卷时限](../Go-based%20Refactored%20System/internal/handler/paper.go#L319-L339) | 部分：可自由配置时长/时段，缺002专属25分钟约束；窗口与已开始个人时限的优先关系需写明和验收。 |

### 3.2 题本、作答与交卷

| ID | 客户需求 | 当前实现证据 | 状态与差距 |
|---|---|---|---|
| MT-08 | 按顺序140题、每题固定5项 | [组卷](../Go-based%20Refactored%20System/internal/handler/paper.go#L262-L339) | 部分：按当前题库sort取配置数量，非002固定140；缺V1～V140唯一连续、单选题型、5项分值1～5以及两题本匹配门禁。 |
| MT-09 | 题干、选项、题序历史稳定 | [单题读取](../Go-based%20Refactored%20System/internal/handler/paper.go#L500-L549)、[计分V映射](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L1565-L1600) | 缺少新版：试卷保存ID/排序/选项关联，但文字读取当前源题；V号读取当前题库sort。旧paper不能视为完整冻结题本证据。 |
| MT-10 | 指导语、滚动/逐题浏览 | [整页答题](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/exam.vue#L1-L50)、[单题页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/examClick.vue#L1-L105) | 已有基础：两个answerType都能承载002；需核对新版指导语、五项文字和移动端实际布局，不直接套00401新组件。 |
| MT-11 | 25分钟倒计时、20分钟提示 | [整页Timer被注释](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/exam.vue#L1-L17)、[单题Timer被注释](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/examClick.vue#L5-L18) | 缺少：两页都未启用计时，未发现剩5分钟提示；数据库limit_time不等于执行时限。 |
| MT-12 | 离开后恢复答案及页面状态 | [恢复/创建分支](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/preview.vue#L215-L258)、[首个未答定位](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/exam.vue#L186-L202) | 部分：已保存答案及同卷恢复有基础，恢复到首个未答而非精确退出位置；查卷失败后可继续创建，创建和绑定人员分两次请求。服务端幂等及保存失败/刷新需实测。 |
| MT-13 | 作答后预览答卷 | [准备页分流](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/preview.vue#L179-L275)、[交卷交互](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/examClick.vue#L337-L404) | 缺少：preview是测前准备，不是作答后核查；整页可滚动重看、单题可导航，但未找到独立“预览答卷”入口/模式。 |
| MT-14 | 每题必须且只能选一项 | [FillAnswer](../Go-based%20Refactored%20System/internal/handler/paper.go#L552-L652) | 部分且高风险：UI radio限制单选；002后端遍历数组，未强制恰一个合法选项，未知ID可得到actual_score=0且answered=1，多项取遍历最后命中分值；缺原始1～5门禁。 |
| MT-15 | 全答后交卷、正式报告140/140 | [前端检查](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/examClick.vue#L378-L404)、[HandExam](../Go-based%20Refactored%20System/internal/handler/paper.go#L669-L730) | 部分且高风险：前端检查未答，后端只检查状态，不强制140完整；不能把前端检查当报告资格事实。 |
| MT-16 | 时限执行、离线到期处理 | [FillAnswer/HandExam](../Go-based%20Refactored%20System/internal/handler/paper.go#L574-L730)、[Worker范围](../Go-based%20Refactored%20System/internal/service/competency_worker.go#L12-L73) | 缺少002：旧保存/交卷未检查limit_time；已有Worker仅扫描competency。002到期如何保存不完整及离线兜底范围仍需确认，正式报告不可生成。 |
| MT-17 | 保存/交卷/完成状态一致 | [交卷事务](../Go-based%20Refactored%20System/internal/handler/paper.go#L681-L722)、[前端先写end_time](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/examClick.vue#L337-L350)、[paper身份](../Go-based%20Refactored%20System/internal/handler/paper.go#L319-L338) | 部分且高风险：无共享paper行锁/条件状态更新；前端先写人员end_time且不等待，再交卷；paper.user_id硬编码101，不能当真实candidate/tester。 |

### 3.3 评分、文案、Word报告及导出

| ID | 客户需求 | 当前实现证据 | 状态与差距 |
|---|---|---|---|
| MT-18 | 13维正反向公式 | [standScore2](../Go-based%20Refactored%20System/internal/handler/tester_score.go#L137-L181) | 已有旧公式：与客户材料公式一致；缺答V默认0导致反向项6，且float64/四位中间舍入，不可直接作为新版精确评分器。 |
| MT-19 | 百分制、五档、13维综合、四模块 | [旧评分](../Go-based%20Refactored%20System/internal/handler/tester_score.go#L137-L181)、[旧综合档位](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L241-L272) | 缺少新版：旧维度1～5，总体13均分之和，非0～100；未接入四模块和新档位。不能把新版值全局塞回旧函数。 |
| MT-20 | 常模、精确排序、最高3/最低3 | [旧逐维结果](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L163-L190)、[维度清单](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L345-L359) | 缺少新版：无13维常模版本事实、稳定选择器及同分协议；四模块得分聚合已确认，但模块常模展示/比较区间尚未批准，不能当既定合同；不把线性映射写成前10%人口排名。 |
| MT-21 | 65摘要/65评价/65建议及总体文案 | [旧字典读取](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L677-L705)、[旧分档取字典](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L736-L761) | 部分：旧字典有定义/评价，但无已批准205条件文案的002版本契约；缺题本/内容/等级精确匹配、缺项失败关闭及审批证据。 |
| MT-22 | Word可维护、填值、六图、PDF | [当前002生成](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L236-L271)、[转换基础设施](../Go-based%20Refactored%20System/pkg/libreofficepdf/client.go#L55-L121)、[材料转换证据](management-traits-word-assessment-20261001.md#L157-L174) | 缺少002接入：现在是Vue/Chromium；LO可复用但不是002渲染器。客户稿0内容控件、六外链图，本地转换六图全缺，须先修兼容和字段合同。94键是草案不是客户既有合同。 |
| MT-23 | 即时算分、自动报告 | [交卷整数合计](../Go-based%20Refactored%20System/internal/handler/paper.go#L694-L722)、[普通挂载生成PDF](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L623-L625) | 部分：交卷保存原始整数总和，新版13维/4模块并未原子持久化；打开旧报告页自动截图上传不等于提交后可靠自动Word生成。新触发待确认。 |
| MT-24 | 单份/指定/批量报告导出 | [管理入口](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L455-L562)、[旧生成](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L191-L331) | 已有旧基础：新版预览、单份、批量需共同选定结果/报告版本；不得读取或混用旧人员PDF指针。 |
| MT-25 | 历史报告保留、可靠重生成 | [先删旧再更新DB](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L279-L331)、[浏览器上传](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L863-L943) | 不满足：服务端/浏览器两链覆盖同一指针，旧文件删除后DB更新失败无法保证旧PDF可用；异步压缩还会改文件字节。需新旧隔离、迁移门禁及修订选择。 |
| MT-26 | 报告与Excel数据一致 | [Excel等级](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L1218-L1256)、[旧PDF等级](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L241-L272)、[模板导出](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L1259-L1334) | 不满足新版且旧口径不一：Excel总体阈值48/36/24/12，页面58/52/45/35；维度边界>=4与>4也不同。Excel答题数还固定写140，无真实完整性。须读同一冻结事实源。 |
| MT-27 | 自定义导出数据/图表/重点提示 | [固定列导出](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L1259-L1334) | 部分：当前固定模板，未找到002动态选择列/报告区块契约；本轮实施范围待确认，不让其阻塞固定个人PDF。 |

### 3.4 管理、运维与扩展

| ID | 客户需求 | 当前实现证据 | 状态与差距 |
|---|---|---|---|
| MT-28 | 量表模板上传及题目/选项维护 | [通用导入](../Go-based%20Refactored%20System/internal/handler/qu_excel.go#L114-L178)、[通用保存](../Go-based%20Refactored%20System/internal/handler/qu.go#L271-L345) | 部分：传统CRUD/Excel可复用；没有完整002题本+公式+文案+模板包导入。通用“正确项”语义也不能代替Likert契约。 |
| MT-29 | 自定义类别、计分、因子公式、解释、建议、预警 | [硬编码公式](../Go-based%20Refactored%20System/internal/handler/tester_score.go#L147-L164)、[字典](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L677-L705) | 部分：题库/选项/字典可维护，公式仍硬编码；未找到002任意公式编辑器/批准发布链。固定已确认评分与“任意可配置平台”必须分开确认。 |
| MT-30 | 实时进度和统计图 | [人员列表](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L47-L104)、[逐人已答查询](../Go-based%20Refactored%20System/internal/handler/tester.go#L147-L160) | 部分：请求时状态/已答数存在，非持续实时推送；需确认刷新频率、统计分母及并发查询成本，不把静态列表称完整实时仪表盘。 |
| MT-31 | 按时间/测试号/姓名/编号/性别/年龄/单位查询 | [前端筛选](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L1-L104)、[后端筛选](../Go-based%20Refactored%20System/internal/handler/tester.go#L47-L72) | 部分：姓名/电话/状态及部分身份条件存在；显示年龄/性别/时间不等于支持条件查询。缺时间范围、性别、年龄、单位等完整组合；测试号定义待确认。 |
| MT-32 | SPSS/SAS兼容文件 | [现有XLSX入口](../Go-based%20Refactored%20System/internal/router/router.go#L206-L211) | 未找到专属实现：Go/Vue/脚本/文档检索未见.sav/.sas7bdat契约或验收。XLSX可另行导入统计软件不等于原生兼容交付，需确认格式/变量字典。 |
| MT-33 | 自动备份/还原 | [备份脚本](../Go-based%20Refactored%20System/deploy/backup.sh#L1-L9)、[MBTI恢复范围](../scripts/db/restore-mbti.sh#L18-L88)、[同步回灌范围](../scripts/db/db-sync.sh#L94-L123) | 部分：存在mysqldump脚本及MBTI恢复/同步工具，无已核验的常态全库自动还原闭环；cron注释不是实际启用证据。项目记忆已有授权临时Schema恢复演练，不能误报从未恢复。 |
| MT-34 | 隐私、认证、加密、安全存储 | [匿名路径](../Go-based%20Refactored%20System/internal/middleware/middleware.go#L33-L83)、[保存和交卷](../Go-based%20Refactored%20System/internal/handler/paper.go#L552-L730)、[PDF落盘](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L279-L305) | 部分且高风险：后台JWT等基础存在；传统paper整段匿名，所读handler未校验参与者绑定令牌。随机paperId不是授权。当前PDF以0644保存且文件名含姓名；实际外部暴露/TLS/静态加密需环境核验。 |
| MT-35 | 更新扩展、维护文档、浏览器兼容 | [方案评估](management-traits-word-assessment-20261001.md)、[设计草案](management-traits-word-design-20261001.md) | 部分：已形成分析/草案，不能据此认定版本升级/回滚/浏览器矩阵通过。仍需002专属测试、运行手册和验收证据。 |
| MT-36 | 团体报告、千人、矛盾预警、>35分钟提示 | [已确认排除/延期](management-traits-word-assessment-20261001.md#L293-L305)、[开放旧团体](../Go-based%20Refactored%20System/internal/handler/candidate.go#L471-L516)、[封闭空返回](../Go-based%20Refactored%20System/internal/handler/tester.go#L458-L460) | 排除/延期：旧团体接口存在不等于完整新版能力；00401的100份/10并发组卷实证不是002千人答题证明。 |

## 4. 必须优先处理的链路缺口

此处P0指**新版正式启用前阻断**，不是声称本轮已复现生产事故。

| 优先级 | 缺口 | 最小必要改造 | 必要验收 |
|---|---|---|---|
| P0 | 题本身份/V号/原始答案不可靠 | 002专属身份、题序与V号分离、完整冻结140题/五项/方向；拒绝错题本 | 两套源库逐题核对；缺/重/错V、外来选项、多选/零值失败关闭 |
| P0 | 新评分与结果事实缺失 | 独立精确13维/4模块/总体与等级，持久化同一结果事实，不覆盖旧函数 | 全1/3/5、反向、边界、四模块和13维总体、HALF_UP、同分排序 |
| P0 | 身份/完整性/时限/事务不闭环 | 确定真实参与者，绑定令牌；同paper序列化保存和交卷，原子更新人员状态 | 绕过UI缺答、多项、跨人/跨卷、过期保存、并发提交、回滚、失败重试 |
| P0 | Word合同及正式文案未可用 | 确认字段位置，205文案版本/批准；六图兼容、值绑定、正式报告门禁 | 全条件文本、缺项拒绝、Word实开、目标LO六图可见、长文案分页 |
| P0 | 历史PDF会被旧链覆盖 | 独立报告修订/选择，旧链隔离；迁移前阻断写入/排空压缩，保留旧文件 | 新生成失败旧PDF可下载；双链竞争、单/批/Excel同版本与SHA |
| P1 | 倒计时/提示/答卷预览及查询缺口 | 明确20分钟提示、作答后预览、选填字段、筛选合同并逐片接入 | 手机/桌面、断点、网络失败、组合查询、显示/导出一致 |
| 待确认 | 可配置平台、自动报告、运维/统计格式扩展 | 分开确认范围和格式；不前置团体/千人等已排除功能 | 独立验收计划，不借本次个人Word报告默认扩容 |

## 5. 对已有设计的约束

此前设计仍为CHANGES REQUESTED，当前报告不自动改写它。四项需先解决：
1. **评分与呈现版本解耦**：纯重渲染不应被不相关模板/文案变化强迫重算。
2. **新exam配置与历史paper证据分开**：exam唯一profile不能代表每份旧paper的证据。
3. **真实participant来源明确**：传统paper.user_id=101不能用作身份校验依据。
4. **共享目标选择合同**：预览/PDF/Excel/批量明确选择同一run与报告修订，而非各自解析“当前”。

物理表数量、94必需键和离线Worker方案都是草案；不能在审查通过前当作实施既定前置。尤其年龄/时长/总体常模无原稿位置的键，须确认新增位置或改为可选。

## 6. 建议推进顺序及验收门禁

1. **只读数据核对**：取得用户指定环境/样本后，对00201/00202的140题、五项、分值、题序、源文本逐字段核验；历史不足证据明确拒绝重算。
2. **纯身份/精确评分切片**：不开放HTTP/DB正式入口，先RED→GREEN验证已确认公式、聚合、常模和选择器。
3. **运行闭环**：冻结题本、参与者归属、保存/交卷/可信到期/140完整性和结果写入；不照抄00401参与者身份假设。
4. **Word与内容切片**：客户位置/文案冲突/批准闭环后，修复六图、字段合同及目标转换；再定页数，不先承诺9/10页。
5. **消费方统一**：管理员查看、单份/批量PDF、Excel及历史访问读取相同事实源；隔离旧双生成链与删除入口。
6. **真实验收后再授权发布**：两题本×开放/封闭，覆盖全答/缺答/错误选项/重复/到期/断网/并发/权限/转换失败/旧PDF保留；仅获staging授权后执行相应验证，生产另行授权。

**下一步不是直接编码整套设计。**先关闭题本数据库对照和上述四项设计冲突，再选择一个独立本地实现切片。此次没有修改客户原件、Go/Vue/SQL、数据库或部署环境。

## 7. 第二轮材料交叉复核：需求不能仅按方案标题理解

### 7.1 本轮新增直接材料定位

下述段落号按`word/document.xml`的`//w:body//w:p`文档顺序计数，包含表格/文本框内段落；不是Word页码。只读XML提取，不在客户原件增加Tag或修改文本。

| 位置 | 已验证内容 | 需求判断 |
|---|---|---|
| 工作簿“管理特质综合表现”C2:G2；“维度评价”D2:H2；“维度发展建议”C2:G2 | L1优秀、L2良好、L3合格、L4欠佳、L5不足 | 002显示必须为“欠佳”，不能借00401的“薄弱”；内部码和显示词分开。 |
| 工作簿“基层员工题本”及“管理人员题本”C1:G1 | 很符合→比较符合→一般→不太符合→不符合 | 工作簿物理列序与方案UI选项1→5相反；必须按选项含义绑定原始分，不能按列号直接赋1～5。 |
| 两题本H68/B68及H97/B97 | V67/V96的两版题干不同；管理版B68有连缀表述 | 不能去掉题干开头的维度内题号后，用它替代全局V号；V号取明确H列，保留两题本身份。 |
| 工作簿“综合评价”A2:D6 | 五档总体标签、完整评价、每档三段总体建议 | 总体标签与维度标签不同；总体建议独立于顶3/底3。D3称“一至两个”、D4称“一个”优先提升维度，但没有动态替换槽；不擅自把底3插入这些正文。 |
| Word表1/正文P1～P12 | 姓名、性别、单位、职务、联系方式、测评日期 | 原稿已有六项位置，职务是否映射系统岗位字段仍需确认；没有年龄/作答时长位置，不能按94键草案直接增加。 |
| Word表4/正文P49～P58 | 高分三项与低分三项两栏，含“维度名：摘要” | 六项真实槽位存在，但没有内容控件；须绑定维度名称与对应等级摘要，不以原稿样例维度为固定数据。 |
| Word表6起的13维详情 | 每维名称/得分/等级/常模、说明、测评结果 | 定义和结果文本必须分清；固定定义维护来源与Excel定义权威关系需定稿，不能同时由两处独立修改。 |
| Word正文P28 | 固定说明称等级为“标准化的相对评价” | 线性映射和固定分段不证明人群标准化；即使程序不计算百分位，固定说明仍需内容/心理测量审阅。 |

**原件核验结果：**XLSX51421、报告DOCX660351、方案DOCX36305 bytes，三份完整SHA与§2和前轮材料清单一致。客户文件没有被修改。原稿已有岗位/高低栏位置的事实，优先于“这些位置完全不存在”的推断；仍不能把已见位置等同于已绑定控件。

### 7.2 需要明确区分的四层

- **题本事实**：140题、全局V号、两套措辞、选项含义/原始分和方向；数据库真实现状尚未逐题比对。
- **计分事实**：13维精确百分制、四模块、总体、等级、常模和顶底选择；不得依赖每次查询时的当前题库关系。
- **内容规则**：195条维度条件文本+10条总体条件文本；机械完整不等于具名批准。两套题本也不证明已有两套不同受众文案。
- **呈现规则**：Word固定文本/版式、动态系统值和条件文案位置；模板修订不应强迫重算答题事实。总体建议、六项摘要、13维详细评价/建议是不同输出槽。

## 8. 第二轮新增代码风险与纠正

以下G编号是评估问题，不是已修复bug编号。P0为新版正式启用前阻断，所有问题均未修复；明确区分源码、内存模拟和真实环境证据。

| ID/级别 | 新发现与实际影响 | 源码证据 | 验证层次/必要验收 |
|---|---|---|---|
| G01/P0 | 整页单选保存复用`multiValue`数组并push，改答同一题后新请求同时带旧/新选项；后端未拒多项，分值由遍历命中顺序决定，不能保证最后选择生效。 | [handSelect](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/exam.vue#L207-L236)、[002保存分支](../Go-based%20Refactored%20System/internal/handler/paper.go#L605-L631) | 已用现有方法内存模拟复现请求数组累积；未测DB。需真实“高→低/低→高/反向项改答/刷新”的最终答案与分值断言。 |
| G02/P0 | 同手机号进行中人员再次Save，以新struct覆盖旧记录，未复制paper/end/pdf字段；已有完成记录默认无ID会被拒，但提供非空ID可绕过该前置拒绝，后续仍按手机号找旧记录。 | [Candidate.Save](../Go-based%20Refactored%20System/internal/handler/candidate.go#L81-L138) | 源码确认，不声称正常无ID重填能覆盖已完成人员。需测试有/无ID、未完成/已完成、并发重复填写及关联保留。 |
| G03/P0 | paper整段匿名规则覆盖管理paging/save/delete；Save接收model并全字段写入，能改变传统paper的状态/关联；Delete清答卷及人员指针。 | [匿名规则](../Go-based%20Refactored%20System/internal/middleware/middleware.go#L58-L60)、[管理路由](../Go-based%20Refactored%20System/internal/router/router.go#L295-L311)、[Save/Delete](../Go-based%20Refactored%20System/internal/handler/paper.go#L137-L210) | 未调用真实匿名管理请求；必须拆精确公开参与者接口和后台接口，再做401/403及零写入验证。 |
| G04/P0 | tester匿名详情直接返回带password的model；登录明文比较且返回password。匿名PUT人员更新还可改密码、paper和exam关联。登录产生JWT不等于后续旧paper接口验证该JWT。 | [详情](../Go-based%20Refactored%20System/internal/handler/tester.go#L174-L189)、[密码字段](../Go-based%20Refactored%20System/internal/model/business.go#L190-L211)、[登录](../Go-based%20Refactored%20System/internal/handler/tester_score.go#L526-L610)、[匿名PUT](../Go-based%20Refactored%20System/internal/middleware/middleware.go#L99-L104)、[更新](../Go-based%20Refactored%20System/internal/handler/tester.go#L327-L378) | 源码确认，未读取真实密码/利用接口。后台bcrypt不能计作tester保护；需独立凭据迁移及身份更新白名单，不在本次擅自执行。 |
| G05/P0 | 状态口径冲突：建卷常量0启用/1禁用/2就绪/3过期，LoginForm注释却为0未开始/1启用/2过期/3禁用，实际只拒2/3。state=1可过登录但建卷拒绝；开始时间未被该登录检查。 | [实际常量](../Go-based%20Refactored%20System/internal/handler/paper.go#L23-L32)、[建卷门禁](../Go-based%20Refactored%20System/internal/handler/paper.go#L238-L253)、[登录状态](../Go-based%20Refactored%20System/internal/handler/tester_score.go#L550-L574) | 以真实代码常量和数据库配置核对，不引用过时规则表。CreatePaper只有state门禁，无直接开始/结束时刻/参与者门禁，需全入口一致性验收。 |
| G06/P1 | 后台stateUrl缺`/api`，与注册路径不符；State处理函数仅校验0～3并更新，未核对题本完整/合法状态迁移。 | [前端URL](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/index.vue#L185-L190)、[原样post](../Go-based%20Refactored%20System/ruoyi-ui/src/api/common.js#L17-L20)、[后端State](../Go-based%20Refactored%20System/internal/handler/exam.go#L920-L951) | 静态前后端不一致；未验证远端有无额外兼容代理，不能仅凭此断言现网404。 |
| G07/P0 | ShowPdf匿名返回服务器路径，匿名下载仅证明路径被任意人员登记，不证明调用者有权限或该测评允许查看。 | [ShowPdf](../Go-based%20Refactored%20System/internal/handler/paper.go#L889-L909)、[下载/路径登记检查](../Go-based%20Refactored%20System/internal/handler/candidate.go#L702-L765) | 源码确认授权缺口；未下载真实他人PDF。应按报告ID+调用者/测评权限鉴权，路径登记只作文件引用校验。 |
| G08/P0 | 批量下载和报告删除有JWT门禁，但handler未校验具体业务权限；删除同时按candidate/tester两表操作同一ID列表，忽略DB错误，且未保护新002历史。 | [路由](../Go-based%20Refactored%20System/internal/router/router.go#L315-L332)、[精确匿名列表](../Go-based%20Refactored%20System/internal/middleware/middleware.go#L48-L56)、[删除实现](../Go-based%20Refactored%20System/internal/handler/candidate.go#L263-L285)、[candidate ZIP](../Go-based%20Refactored%20System/internal/handler/candidate.go#L769) | **纠正复审误判：不能称candidate整段匿名，也不能称此删除/ZIP匿名可达。**JWT与对象授权分开，需低权限/跨测评/同ID双类型/删除失败测试。 |
| G09/P0 | Prefix目录包含检查不严格；上传文件名含未经安全处理的人员名/标题；tester旧PDF直接删除。安全目录字符串相同前缀不等于位于目录内，符号链接边界也未解决。 | [candidate上传落盘](../Go-based%20Refactored%20System/internal/handler/candidate.go#L582-L634)、[下载边界](../Go-based%20Refactored%20System/internal/handler/candidate.go#L702-L746)、[tester上传](../Go-based%20Refactored%20System/internal/handler/tester_score.go#L438-L487) | 校验缺陷静态确认，实际越界取决于目录/权限/链接拓扑；未做攻击调用。应核验绝对路径、Rel和打开时边界、生成唯一文件名。现有安全规则的Prefix示例亦不充分，本轮未改规则。 |
| G10/P0 | 服务端和candidate浏览器持久化的时间格式末尾`000`是字面量零，不是毫秒；路径未含paper/revision。同名同标题同秒可碰撞，若oldPath==saved，服务端写后删除会删本次新文件。 | [服务端路径/删除](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L279-L310)、[candidate路径](../Go-based%20Refactored%20System/internal/handler/candidate.go#L607-L634) | 静态机制确认，未并发复现。tester使用`.000`再去点，**不是同一格式缺陷**；仍需唯一revision路径及同目标互斥，Chrome池不是业务锁。 |
| G11/P0 | 报告生成不校验paper完成与140合法作答；页面20秒强制ready并标incomplete，重试器最终把incomplete接受为成功，仍更新PDF标记。 | [生成输入](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L191-L235)、[强制ready](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/result2.vue#L589-L616)、[接受不完整](../Go-based%20Refactored%20System/internal/handler/exam_report_gen.go#L91-L103) | 源码确认；**作答完整性与渲染完整性是两个门禁**，不能只看文件>1024字节或PDF生成成功。需加载失败/缺文案/图表延迟测试。 |
| G12/P0 | 双Excel入口授权/脱敏不一致：RawAnswers检查权限，RawData对002未执行对应handler权限检查，固定模板直接输出身份证/手机号；回退也应统一披露策略。 | [RawAnswers权限](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L162-L194)、[RawData分流](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L856-L902)、[明文列](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L1299-L1305) | JWT仍存在，不称匿名导出；未下载真实数据。需低权限/管理员、两入口、模板/回退全部交叉验收。 |
| G13/P1 | 单题导航的保存/详情响应无序列控制，可由旧响应覆盖当前题；整页页内值已改变不等于保存已成功；交卷未等待全部在途保存。 | [单题保存/加载](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/examClick.vue#L411-L507)、[整页保存](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/exam.vue#L207-L236) | 异步控制流静态确认，未浏览器慢网注入；须重试/快速导航/连续改答/马上交卷的持久化断言。 |
| G14/P1 | 002导出仍逐人查询并重算；部分查询错误被忽略，产生缺分/空表仍输出文件；RawData还把查询错误细节回给客户端。 | [逐人查询](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L375-L448)、[模板重算及吞错](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L1311-L1333)、[错误回显](../Go-based%20Refactored%20System/internal/handler/exam_pdf.go#L940-L944) | 未测吞吐/故障注入，不能写预计提升倍数。新版应批量读取冻结结果、响应前完整验证，不回退混入旧版。 |

**源题维护纠正：**[Qu.Save历史引用守卫](../Go-based%20Refactored%20System/internal/handler/qu.go#L296-L306)已拒绝修改被试卷引用的源题；不是“任何管理员都能随意编辑旧题”。但这不等于完整快照：V映射仍读取可变题库关系，选项/文字仍读源表，而且不能证明所有写路径和现库约束完整。未来应保留此守卫，不用新增快照为理由放宽旧边界。

## 9. 全链影响清单与复用边界

此清单用于后续改造范围判断，**本轮没有对这些消费方实施修改**。

| 环节 | 已有消费方/数据流 | 后续002改造判断 |
|---|---|---|
| 配置/源题 | Exam.Save/State、题库关联、Qu.Save/Delete/导入 | 需同步：002专属题本/完整性及冻结边界；无需改变001/003评分及00401版本字段，理由为产品作用域不同。 |
| 参与者创建/登录/重填 | Candidate.Save/Update、Tester.LoginForm/DetailByIDNumber/Update、准备页建卷后绑定 | 需同步：真实participant归属、最小响应、关联保留和幂等建卷；不是仅新增结果表就能保护历史。 |
| 保存/交卷 | 两传统答题页、FillAnswer、HandExam、人员EndTime、Paper.Save/Delete | 需同步：合法单选、共享锁/时限/原子完成与旧公共写入口隔离；不能只修前端。 |
| 结果读取 | Paper/Tester/Candidate标准分、AnswerDetail、result2及人员列表 | 需同步：新版读取冻结run/身份/题本事实；旧标准分无需改尺度，理由为历史消费者仍依赖1～5。 |
| 模板与正式内容 | 模板管理页、002模板入口、文案匹配、DOCX绑定、LibreOffice | 需新增002合同并接入：现00401/MBTI渲染业务无需放宽，理由为字段、图表和等级不相同；通用转换器可复用。 |
| 生成/下载/删除 | 服务端GenerateReport、两PdfPersistence、ShowPdf/PdfUpload、双ZIP、直接报告删除、异步压缩 | 需同步保护并选定目标：新旧报告隔离、同revision文件、权限/失败恢复；漏一个旧入口仍会覆盖或解除关联。 |
| Excel | ExportRawData/RawAnswers、002模板/回退、字段披露 | 需同步：同run精确事实/版本/完整性及授权；三Sheet是否采用需客户确认，不照搬00401。 |
| 迁移/保留/运维 | 旧PDF捕获、历史paper证据、DB迁移/备份、报告文件/字体/转换器 | 需明确批准样本和环境后验证：不同历史paper分别证明身份；备份包括DB、PDF、模板、内容/代码版本，不能只备份数据库。 |

**架构判断：**独立002评分事实与报告呈现分层是必要边界，但不是本轮授权大重构。现有通用入口、五级选项显示、QRCode、XLSX和转换器可复用；参与者认证、状态更新、指针覆盖不能原样当作安全基础。配置、计分、内容和呈现应各有版本职责，至少能区分“重算”与“纯重渲染”，物理表数量仍待设计审阅。

## 10. 验收矩阵、当前证据与最终判定

### 10.1 必须安排的验收分支

以下不是已执行测试；每行都需落实到测试索引/真实证据后才能标完成。本轮未进入bug修复，没有生成RED测试文件或修改测试账本。

| 组 | 关键分支 | 当前证据/缺口 |
|---|---|---|
| A 题本 | 00201/00202、140唯一V号、100正40反、5项含义/分值、V67/V96、缺行/重行/错方向/错题本 | 材料已核验；数据库/发布/组卷对应未验证。 |
| B 配置/身份 | 开放/封闭、显示必填/显示选填/隐藏、手机号/身份证、非空请求ID、重复重填、多人同名、跨测评、状态0～3、开始/结束窗口 | 代码缺口已列；真实HTTP/DB未验证。 |
| C 保存 | 正反向项高→低/低→高、未知/外来/空/重复/多项、快速改答、失败重试、保存中交卷 | 原整页方法数组累积内存模拟已复现；真正HTTP/DB未验证。 |
| D 续答/到期 | 刷新/重登同卷、精确退出位置或首未答、服务器与客户端时钟差、20分钟提醒、25分钟到期、离线/服务重启 | 续答基础已有；新版语义及Worker范围待确认/验证。 |
| E 提交 | 手工139/140拒绝、完整提交、可信到期不完整、双连接并发、保存/提交竞争、子表失败回滚、请求重试幂等 | 仅140/140正式报告资格已确认；002真实事务/并发未验证。 |
| F 评分 | 10/11/12/13题分母、全原始1/3/5、逆向极值、0～100、90/70/30/10边界、精确等级/排序、最终HALF_UP、模块3/4/3/3、13维等权、全部同分重叠 | 旧8项测试通过，**不覆盖新版**；原始全1/全5不是正反向处理后的全最低/全最高，测试应分别构造。 |
| G 内容 | 65摘要+65评价+65建议+5总体+5总体建议、欠佳标签、两题本文案作用域、缺项/错版本/未批准、总体多段建议、固定统计措辞 | 材料数量已确认；正式内容批准/绑定未完成。 |
| H Word/PDF | Word实开、六图逐图可见、原生系列/点数/顺序、零外链、Word重存、正文/页眉/页脚、多段/混合粗体、长姓名/单位/文案、A4/字体/无孤立标题 | 前轮原稿LO六图缺失；修复稿/目标LO未验证，页数待定。 |
| I 历史/文件 | 旧PDF保留、生成失败旧文件可读、碰撞/同名、压缩排空、上传/生成/下载/删除竞争、同ID双人员类型、迁移/回滚 | 原始历史证据和指定样本未提供；未迁移。 |
| J 权限 | 匿名paper管理拒绝、tester密码不返回、低权限/跨人/跨测评、路径同前缀兄弟目录/符号链接、PDF公开策略、双导出同披露规则 | 静态风险确认，无真实安全矩阵通过证据。 |
| K 管理/导出 | 组合筛选、COUNT与列表、版本目标一致、两入口/模板/回退、无新版结果/不完整结果、批量部分失败/上限/取消、PDF/Excel逐值一致 | 固定旧链存在，新版待实现/验收。 |
| L 性能/运维 | 独立人员节奏登录/开卷/保存/集中提交，PDF错峰/排队取消，导出N+1、backup调度/异机副本、DB+PDF+模板恢复演练、回滚 | 本轮未压测或环境检查；不承诺千人/SLA、不以历史00401容量替代。 |

### 10.2 本轮实际完成的验证

- **材料**：三份原件字节数与完整SHA重新核验一致；只读抽取等级表、两题本差异、总体建议和Word表1/4/详情位置。
- **编译**：本地`Go: Build`任务完成；实际执行`go build -o bin/server.exe ./cmd/server`，无编译错误。产物在现有bin目录，未部署。
- **既有测试**：`TestStandScore2_AllMiddle/Dimensions/ReverseScoring/AllFour`、`TestMngScoreLevel/TotalLevel/Diagnosis`、`TestEvalExpr_StandScore2Formula`，8 passed、0 failed。这些验证的是旧公式/旧等级，不是新版评分或安全链。
- **现有方法内存模拟**：从[原handSelect](../Go-based%20Refactored%20System/ruoyi-ui/src/views/paper/exam/exam.vue#L207-L236)直接提取方法，mock仅捕获fillAnswer参数。q1高→q2低→q1低三次调用得到数组依次为`[q1-high]`、`[q1-high,q2-low]`、`[q1-high,q2-low,q1-low]`，断言通过，证据标记`ACTUAL_HANDSELECT_SHARED_ARRAY_REPRODUCED`。前次PowerShell美元符号转义失败后已修正重跑，最终退出0；未调用HTTP/DB，也未改原方法。
- **尚未执行**：全量Go/Vitest、真实002端到端/数据库题本对照/安全复现/新Word/目标服务器转换/压测；不将本地编译与8项旧测试写成全系统通过。

### 10.3 最终评估结论

**材料足够建立新版评分和报告的主要需求，但不足以直接激活正式报告；现有代码足够提供旧基础链，不足以安全接入新版。**

必须优先落实的不是“把Word转成PDF”一项，而是：①合法且冻结的题本/答案，②真实参与者与服务端提交门禁，③新评分及结果事实，④批准文案与可运行Word合同，⑤新旧报告完整性/生命周期及全消费方一致性。副产品扩展按原确认范围排除或另行确认。

最终仍待客户/环境确认：V67及建议错字/频率、模块等级/比较与顺序、岗位显示映射/年龄/时长新增位置、最终名称/统计说明/页数、自动报告触发、到期不完整/离线兜底、内容与测量批准、新Excel合同、历史样本/数据对照环境和执行授权。**评分纯函数可独立准备，但这些事项未关闭前不可启用正式新版链或批量历史重算。**