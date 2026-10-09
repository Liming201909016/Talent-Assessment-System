# 002 客户V2.8最小报告链 — 2026-10-07

## 1. 未完成与边界（先列）

- **真实历史两份140题答卷未重发、未证明来源资格**。本轮无SQL/SSH/HTTP，沿用10-06只读审计仅作为历史事实：有完整原始分和700桶，不等于具备历史V→源ID→题干/维度映射及人员完成时快照。禁止查当前源题补造当时题本，旧13维分也不可反推新原始分。
- **管理员历史reissue按钮/数据库loader/产品上线未实施**，也未实施完整正式审批或全参与者UI。已有draft/formal/Worker/前端改动不删除、不回滚。本轮仅完成独立适配器与实际本地PDF切片。
- 全Go/前端全量、真实MySQL、staging/production、Word桌面、跨LO引擎、全页几何避碰、独立CodeReviewer均未验。终端既有Frontend任务exit1保留，未调查或重跑，不能称前端全绿。历史6272/435/9skip不是本轮结果。
- 当前根目录未发现Legacy目录，不能声称本轮读取了原Java生成算法；仅核当前Go旧保存/组卷代码与既有历史审计。Legacy写操作0。

## 2. 实际实现与输入信任合同

[独立适配器](../Go-based%20Refactored%20System/internal/service/management_traits_audited_reissue.go#L126)只接原始UTF-8 JSON字节、detached Ed25519审计签名、应用持有的审计公钥、原SHA锁定内容包和预算。**无HTTP入口、无签名器、无配置键、无DB依赖**；公钥不能来自请求。测试随机合成公钥不作为真实历史信任根部署。

源策略精确`legacy_verified_snapshot`，报告策略`audited_reissue`；既有`new_creation/submitted_snapshot/completed`产品门禁不变。不是把旧记录塞进新版profile/bundle/run表再渲染。签名证明审计者签署字节，**不证明真实归档事实**；真实审计者须先查原始载体，缺证不签。任意合法SHA/签名/布尔均不是正式内容批准。

源JSON必须完整包含：

- `schema/policy/synthetic/repoCode/evidenceKind/evidenceReference`：00201→staff，00202→leader，original historical_export审计引用，不接受new_creation替代。
- `manifest`：原题干、唯一V1～140、固定13维与方向、原五项文字/原始分/显示序、明确历史题本版本。评分与常模轴精确使用已确认百分制版本，不复用00401计分。
- `mapping/input`：原sourceQuestionID/sourceOptionID、每卷paperQuestionID/显示序、唯一140原始选择、answered/raw1～5；沿用S2B固定V/维度/方向/哈希及源选项一致性校验，不接受任意dimension map。
- `owner`：paper/exam/participant type+ID与input完全一致；historical_completion_snapshot身份、字段白名单、原开始/提交时间。ParseInLocation支持RFC3339与本地秒格式；缺时/倒序拒绝，不用现在时间代替。用时由这两个已声明原始时间推导，不新增客户模板用时槽。
- `buckets`：每卷题恰五项、各raw1～5/sourceID唯一、一项checked且与actual raw相同、一项is_right且正向raw5/反向raw1。已答不足、重复V、未知维度、桶方向漂移、映射缺失、身份不符均拒绝。

[旧组卷代码](../Go-based%20Refactored%20System/internal/handler/paper.go#L340-L409)复制选项的score/is_right、桶qu_id是源题ID而非paper_qu.id；[旧002保存](../Go-based%20Refactored%20System/internal/handler/paper.go#L611-L631)将所选score写actual_score。[旧模型](../Go-based%20Refactored%20System/internal/model/business.go#L154-L185)没有题干或V号快照，因此完整桶仍缺历史映射。上述100正40反为固定目录及合成验证，不冒称本轮真实历史SQL。

适配器复用`CalculateManagementTraits`精确Rat，不增加新计分算法；raw反向一次、140完整后13维等权总体/4模块成员维度等权、705/13综合常模、未舍入等级和排名。源字节/签名/审计根SHA、mapping/manifest/input SHA、原时间及推导用时、140raw/reverse/final、13维sum/count/exact/norm/level、4模块count/exact和DTO形成不可变快照，所有公开读取返回独立副本。

[共享客户文案构建器](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L175)只抽取原条件选择与DTO填充；原函数签名、既有DTO JSON与用途不变。两个来源边界分别验证后才调用此私有核心。客户205规则仍SHA锁定；36段是本样本实际选择数，不说每份输出205段。

## 3. 影响清单（全部已核）

| 消费方/资产 | 处理与理由 |
|---|---|
| [原BuildManagementTraitsTestReport](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L153) | 同步修改，仅抽取文案核心；new_creation等前置不改、公共签名不改 |
| [原runtime loader](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go#L153) | 无需修改，继续原事务/owner/子表/来源门禁；生成/下载回归已跑 |
| [原报告DTO测试](../Go-based%20Refactored%20System/internal/service/management_traits_test_report_test.go#L40) | 无需修改，字段及结果合同不变，已重跑 |
| [原下载测试](../Go-based%20Refactored%20System/internal/service/management_traits_report_download_test.go#L30) | 无需修改，原metadata/file/DTO一致性不变，已重跑 |
| [原身份测试](../Go-based%20Refactored%20System/internal/service/management_traits_identity_chain_test.go#L390) | 无需修改，原optional身份投影不变，已重跑 |
| [002 Word renderer](../Go-based%20Refactored%20System/internal/handler/management_traits_test_word.go#L118) | 无需修改，继续完整90Tag/6图/SHA与TEST校验；未去16 LINK正文或绕formal gate |
| [通用LO](../Go-based%20Refactored%20System/pkg/libreofficepdf/client.go#L89) | 无需修改，原隔离profile/队列/90s/%PDF/临时清理；实际转换已跑 |
| 00401/旧评分/旧save/所有HTTP API/旧pdf_path/模型/SQL/环境/前端 | 无需修改；只技术参考，禁止复用会删除旧PDF的save。指纹保护；零新DDL/接口/依赖 |

## 4. RED→GREEN与真实验证

- 修改前相关基线native0：10pass/0fail，`TestBugManagementTraitsPDFVisibleTestLabel`明确未opt-in跳过1。23既有测试开关仅子进程清除、开始时present空；新产物开关只向专属子进程注入。
- 新适配器首先写测试、实现前编译RED native1；[原RED证据](../scripts/test/results/mng-reissue-20261007-final-7bd362/adapter-red.json)含原输出SHA，未伪装行为断言RED。
- 最终[适配器/身份/原DTO](../scripts/test/results/mng-reissue-20261007-final-7bd362/adapter-green.json)：248pass/0fail/0skip；[真实LO](../scripts/test/results/mng-reissue-20261007-final-7bd362/real-lo.json)：1pass；[Word相邻回归](../scripts/test/results/mng-reissue-20261007-final-7bd362/word-regression.json)：11pass；[共享消费方](../scripts/test/results/mng-reissue-20261007-final-7bd362/shared-consumers.json)：43pass。各native0/解析0，303为含子项pass事件，不是303个独立顶层测试或覆盖率。
- `go build ./...`和`go vet ./...`各native0/stdoutstderr0；初始Go Build任务完成，未暴露numeric exit，不冒填。新Go/JS/Python编辑器诊断0，Node语法检查0。
- [适配器测试](../Go-based%20Refactored%20System/internal/service/management_traits_audited_reissue_test.go#L51)覆盖两code×0/50/100、反向1/5一次、immutable返回；[不足来源矩阵](../Go-based%20Refactored%20System/internal/service/management_traits_audited_reissue_test.go#L79)覆盖签名/外来根/篡改/预算/严格JSON空串null类型重复case及身份/时间/映射/桶；[旧guard保持](../Go-based%20Refactored%20System/internal/service/management_traits_audited_reissue_test.go#L121)与tester/空快照已验。
- [真实最小链测试](../Go-based%20Refactored%20System/internal/handler/management_traits_audited_reissue_pdf_test.go#L21)从原字节+独立签名→服务适配器→当前002客户Word→实际LO，无DB/profile/run迁移。75 OPC部件中68个非值部件逐字节不变；原Word非文本/图表非值样式由原回归校验；星级/页脚原政策不改。
- [独立Python oracle](../scripts/test/management-traits-reissue-pdf-oracle.py#L22)复用既有独立SPECS/Fraction，不调用Go算术；实际逐项校验140/40/13/4/705/13、205规则精确条件选择的36段原文、OOXML6图数据及PDF五环矢量/13柱/常模12段折线。原三件整文件与styles XML的SHA记录于[独立证据](../scripts/test/results/mng-reissue-20261007-final-7bd362/independent-oracle.json)，均保持原SHA。
- 本机已有LO26.2.5.2、Python已显式配置本workspace .venv3.14.4；未安装、更改权限或环境。真实Python oracle native0，不以AST或服务自检查当PASS。
- 原一次PowerShell内嵌JS引号解析失败，未启动Go/写证据；用literal stdin重试43pass/native0。第一份合成PDF及原收据保留，不覆盖；补四模块exact证据后完整重验最终独立目录。

## 5. 最终可查看样本及来源

- **[实际PDF](../scripts/test/results/mng-reissue-20261007-final-7bd362/report.pdf)**：synthetic/TEST，654393 bytes、A4 **9页（本样本实测，非强制页数合同）**，LO26.2.5.2。
- PDF SHA：`2aca74466b5ee018a78aa3e898591c21b9c1dae9139cf6a6a3a97e0028780f55`。
- **[原始来源](../scripts/test/results/mng-reissue-20261007-final-7bd362/source.json)** SHA：`d941fc9c99965050a324e5d6e21d51ec4ad13ec5b49e2fc6137ac4b77938dab5`；完整题干为synthetic，人员无真人PII，不冒充真实历史卷。
- **[不可变评分/报告快照](../scripts/test/results/mng-reissue-20261007-final-7bd362/snapshot.json)** SHA：`2acdbc6b8017d742ed0b97609488495817d8965c06d63e36a412da23c0645726`。
- [DOCX](../scripts/test/results/mng-reissue-20261007-final-7bd362/report.docx) SHA：`af61d9f56b8fde46339e5327fe4688b4f5a3a8742f105f2bcba03b478a7474c8`；运行模板SHA仍`05c55e77...`，内容仍`b0498249...`。
- [渲染侧车](../scripts/test/results/mng-reissue-20261007-final-7bd362/render-receipt.json)、[最终九页总览](../scripts/test/results/mng-reissue-20261007-final-7bd362/all-pages.png)、[第三页六图](../scripts/test/results/mng-reissue-20261007-final-7bd362/page-03.png)。已实看总览与第三页；未声称每个字符的全页无重叠/无遮挡或Word桌面等价。

## 6. 安全收口与下一slice

[源码范围证据](../scripts/test/results/mng-reissue-20261007-final-7bd362/scope.json)：既有B区仅上述报告数据文件变化；新增service适配器/单位测试/handler集成测试3 Go文件；其他纳入保护的Go/前端/config/Word/XLSX均同SHA。新增C区两离线验证脚本及本文，项目记忆/分支账本追加。旧未提交改动全部原样保留；不全Go格式化、不git commit/push。

本地PDF以O_EXCL独立落盘，仅ignored测试目录；通用转换器删除的只是本次自有临时workspace，不旧PDF/旧资产。remote DDL/DML/backfill/history user writes/部署/restart/共享配置权限更改均0，正式运行环境仍关闭。

下一slice须先取得真实历史的原V映射/题干与选项版本/原身份及时间的可靠载体，建立应用固定审计根或等价可信loader再接管理员单卷reissue；证据缺失明确拒绝。管理员按钮/历史查询/正式产品授权不是本轮交付，不在本轮自动继续。