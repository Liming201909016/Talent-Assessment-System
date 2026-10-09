# 管理特质002：生产准入与回滚准备（2026-10-06）

## 2026-10-09 本地准入重跑：GREEN_LOCAL，不等于production GO

同日旧RED机器判定保持不可变。唯一失败是C区harness测试漏掉producer新增且正确的`legacyMarkerRejected=true`期望；producer实际先构造旧固定marker并断言拒绝，符合active generation与旧baseline不可复用的既定安全决策。先重跑取得RED exit1，再只补exact expected field，GREEN exit0；7个字段完整deep equality，无其他expected-field遗漏。

fresh验证：005前端2文件208/208、全前端34文件598/598、diagnostics 0。Go/Vue产品源码、SQL、模板、数据库与runtime均未改；复用同日SHA锁定的全量证据：Go 6484/0/14、build/vet/Linux门禁0、Go coverage 44.3%，前端coverage 60.25/95.53/39.13/60.25，PDF oracle 2份18页/identity-v2 COMPLETE/credential findings0。新[rerun verdict](../scripts/test/results/production-readiness-local-rerun-20261009-040307/final-rerun-verdict.json)为`GREEN_LOCAL`。

外部门禁仍明确存在：本轮未跑browser E2E、staging、production、共享数据库及环境opt-in MySQL/LibreOffice测试；未部署。production仍须本文件G1～G6及用户明确上线授权，不能由本地GREEN自动放行。

## 1. 当前判定：NO-GO；准备文档不等于生产发布授权

**正式报告开放方案未确认、未实施，是业务交付 P0，不只是内容签字或删除 TEST 水印的问题。** 当前新建002默认选择的是 TEST 链；后端生成/查看/下载仅允许 local/staging。按真实 production 环境运行，现有新版报告入口拒绝请求，不能承诺客户可用正式报告。

- 最新完整 staging 验收仍 BLOCKED：full f12 保存563次、三手工报告/三 native 通过；未完成第四自然报告、自然0/3/140、20分钟提示、续答自然5分钟到期及 expired409、同批四份独立 Python oracle。跨批单份不能补为同批4/4。
- 旧13ee/9b停止在00202候选冻结 Timeout；旧安全收据没有具体子步骤/HTTP，原因未证，不能称报告生成 bug。本轮优先用既有驱动单一 freeze-only 探针取证，不盲重跑560题。
- 主 HTTP/Worker 实际竞争 UNPROVEN，观测默认关闭；默认全Go九项环境 skip 与 CGO0 race 未运行必须明示，不把其他00401环境欠项全部列为本业务P0。
- 本轮14:29:13 staging独立终验确认后端8baa7f87…、前端353c9fb3…、PID17552不变、三服务active/healthok；旧465/源/配置/Schema/cache同。393是既有完整manifest资源数，本轮不是重新逐公网下载393资产，更不是生产核验。
- 本轮授权只有问题解决、受控 staging TEST 与生产准备；**productionReleaseNotAuthorized=1**。不访问/写生产、不修改DNS、防火墙、NSG、共享配置或生产env，不额外发版/restart、不启观测。

证据：[最新限定验收](management-traits-default-staging-result-20261006.md)、[旧最终机器判定](../scripts/test/results/mng-default-diagnostic-20261006-a7efc65f/final-verdict.json)。本轮探针结果另行追加，不改原失败。

## 2. 必须逐项关闭的准入门禁

| Gate | 当前状态 | 准入所需证据/责任 |
|---|---|---|
| G1 当前版本完整 runtime/UI/PDF | 🔴 未完成 | 修精确阻断后，一次同批00201/00202×开放/封闭完整UI；原1500秒自然0/3/140、20分钟提示、续答5分钟401、有效原凭据到期保存409；四真实Download/saveAs、DB/private/原始DTO字节SHA、独立算术/九页文本图表及精确零分环验证全部通过；finally仅自有闭包0 |
| G2 正式业务能力与批准 | 🔴 未确认/未实施 | Liming分别确认正式205内容/Word精确SHA、内容/测量审批及日期、未决措辞/题本版本；再批准独立正式报告方案与实现测试。不能用 TEST 环境伪装生产、删除用途警示或加未批准 bypass |
| G3 历史兼容 | 保留政策已定；发布前复核待做 | 旧测评/原始答案/旧PDF只读原路保留；新建默认不触发旧纸迁移/重算/回填。历史证据不足不得用当前140/700源映射补造；任何历史新版方案另行批准 |
| G4 可复现发布包 | 待封包 | 冻结真实工作区输入、编译器/依赖锁、Windows/Linux构建、前端bundle、模板/内容/SQL逐SHA manifest及源→产物关联；独立review和具名准入。脏Git revision不证明包来源；旧21项追溯风险单独决策，不回滚当前已批准完整源码、不无限阻断现版本验证 |
| G5 竞争与环境证据 | ⚠️ 主race未证 | 恢复库服务层竞争PASS不是主HTTP/Worker实证。由负责人明确要求补安全独占观测或具名接受此额外证据缺口；不自动重启、修改共享MySQL开关或以唯一run冒竞态已证 |
| G6 生产目标与运维 | 🔴 未确认/未授权核查 | 用户确认目标主机/域名/应用与DB路径、运行账号、维护窗口、备份恢复负责人、回滚阈值与恢复点。39.106.61.48仅旧文档，不能推定当前生产；20.200.136.133已明确是 Azure vm-ubuntu-go-dev staging |

### 正式支路的真实代码依据

- [后端环境门禁及仅TEST路由](../Go-based%20Refactored%20System/internal/handler/management_traits_report_runtime.go#L18-L41)：只 local/staging；没有 activation/formal 参数。
- [production拒绝测试](../Go-based%20Refactored%20System/internal/handler/management_traits_report_env_test.go#L11-L32)：unset/production/prod/STAGING均关闭。
- [前端结果页](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/managementTraitsResults.vue#L3-L24)：明确不提供正式报告、旧API回退或历史重算。
- [新建配置警示](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L3-L10)：默认新版仍是TEST；保存与冻结独立。
- [持久化与读取校验](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go#L209-L255)：mode=test及固定TEST版本/标签；不是把env改为production即可开放。

正式方案需要确定：批准记录与适用环境、精确内容/模板版本绑定、生成与下载权限、已完成140题限制、不可变run/revision/current/audit、私有存储/下载校验与UI分流。不得自行承诺现代码支持尚未实现的正式能力。

## 3. 生产元数据待确认（不收集密码/令牌）

1. 当前目标主机/域名/端口、SSH账号与既有密钥引用、实际systemd服务与Nginx应用路径。
2. 数据库版本/schema/应用账号权限、各业务已安装迁移集合；不复制staging连接串或秘密。
3. 私有报告绝对根、目录归属/0700文件0600、允许访问账号；不添加公网alias。
4. 正式内容与模板精确SHA、批准人/日期/环境，题本来源政策（V67/V96差异保持，不能偷偷修DB）。
5. 维护窗口、停止新开/写入与恢复策略、备份保存位置/留存、恢复演练及RPO/RTO接受人。
6. 当前进行中/到期答卷的维护策略，覆盖002及00401：startup Worker会立即结算到期卷，不能只看002空闲或把deadline重置。续答凭据不得进入日志/链接收据。

## 4. 仅在明确“部署生产/上线”且G1–G6签署后执行

本节是执行清单，**本轮不执行任何步骤**；额外 staging 发布也须本地汇总后单独统一批准。

1. 核对批准的唯一目标、维护窗口和包manifest；实际production安全只读预检，确认所有应用活动、资源与版本/Schema；发现漂移停止。
2. 创建该目标受限当前全库备份（single-transaction、routines/triggers/events）、旧binary/dist、配置/unit/dropins、模板/内容、上传树及全部私有/历史PDF。gzip/tar/完整SHA与原uid/gid/mode核验；秘密归档留服务器受限目录，不入Git/记忆/本地公开收据。备份失败立即停止。
3. 先在独立恢复副本验证批准的迁移与新旧二进制兼容；明确DDL/锁时间及首次/重复签名。当前[002显式迁移](../scripts/sql/management_traits_001_runtime.sql)的staging 11表/15FK验证不替 production 实际核验；不得默认应用或AutoMigrate/ALTER旧源列。
4. 按维护策略排空在途答题、生成/下载与Worker；确认PID/cgroup/监听/应用连接排空，不盲重启Nginx/MySQL/Redis。活动跨业务答卷或旧冻结profile不兼容即停止。
5. 仅部署已批准代码/模板/内容/环境配置；后端→模板/配置→前端→权限→逐步验证。是否需要restart只按批准窗口执行，runtime Schema cache需新进程验证；不启主race观测默认键。
6. 使用真实应用账号完整Schema/权限/源140700与冻结合同预检，主表旧数据与PDF逐SHA；health、管理员权限/参与者purpose隔离、旧001/002/003/MBTI/00401只读兼容、正式独立报告生成/认证下载验收。任何新合成写入必须精确所有权且可清理，不用普通旧delete绕守卫。
7. 观察签署的时间窗口：health、固定error-class/HTTP5xx、队列/DB连接/磁盘/LO失败、正式报告拒绝及孤儿。只留allowlist计数/哈希；禁止密码、JWT、DSN、他人原始SQL或PII日志。阈值未经实测不得承诺SLA或零停机。
8. 独立终验与签字后恢复写流量；保留全备份、版本manifest和失败收据。仅清该批批准的上传/合成闭包，不删他人资产或正式备份。

## 5. 回滚/停止清单

- **立即停止/评估回滚**：包SHA不符、健康/Schema/权限失败、旧PDF或源数据漂移、核心登录答题拒绝、正式报告关闭/数据不一致、孤儿文件/跨卷关联、已签署5xx或资源阈值超限。不要通过放宽guard补救。
- **代码/前端/配置回滚**：先禁止新写并排空在途请求/Worker；恢复本次备份的旧binary/dist和获准配置及原uid/gid/mode/时间；验证新Schema下旧程序安全、历史PDF可读、各应用健康，再经负责人确认恢复写流量。
- **数据回滚不同于代码回滚**：新11表及已产生的新profile/run/report默认保留。禁止盲DROP新表、删revision或全库restore覆盖上线后合法数据；如数据已写，先保存当前快照、分析增量与外键/共享引用，取得单独恢复授权与明确恢复点。
- **文件回滚**：旧PDF/模板按原SHA保留，不以重生成冒恢复；私有新文件按DB绑定/UUID/大小SHA和允许目录处理，不glob删除，不把备份0600副本元数据当原部署mode。
- **撤销凭据/续答**：按审批的安全策略处理受影响会话；不全局登出/重置管理员，不复制staging secret，不改变业务deadline或解除冻结。
- **恢复验收**：准确记录已恢复/未恢复、停机时间与数据损失范围；重新核健康、旧链/冻结卷与报告、Schema/资产manifest，保留失败与回滚证据，不写“保障零停机”。

## 6. 本轮问题解决进度

**限定完成**：250ms合成保存延迟的真实本地SFC有效RED exit1，两confirm错误复用“提示”框；三轮后用户另批仅一次最小driver修正，同exam保存response屏障＋精确“确认冻结 TEST”标题，原夹具GREEN exit0。[RED](../scripts/test/results/mng-freeze-local-9dd3d91995e6/verdict.json)、[GREEN](../scripts/test/results/mng-freeze-local-b582a14691d0/verdict.json)。不称产品freeze APIbug或所有旧timeout唯一cause，旧完整失败原样。

单cc784已真实staging正常UI创建/保存取消/重新保存/独立freeze HTTP200，独立SQLfrozen=true/25min/paper0及mapping/manifestSHA有效；精确cleanup/final首次native0/owned0/11-private0、旧465-source-schema-cache-PID同。只关闭单freeze测试切片，G1完整四UI/自然/四native-Python仍未关闭。[短探针](../scripts/test/results/uf054-ui-cc784023556d/summary.json)、[独立SQL](../scripts/test/results/uf054-ui-cc784023556d/freeze-only-pass.json)、[终验](../scripts/test/results/uf054-ui-cc784023556d/final-1791296953908-attempt1.json)。

本地四SFC/API202pass/exit0、641批准源/Go215逐SHA保持；两JS语法/diagnostics0。本轮无产品源码变更，不新编译Go/前端或引用旧6176为新证据；未运行Python。所有独立窗口已关闭，原集成页未改，本轮没有新观察其auth。生产准备文档completed，不代表完整产品完工。

下一最小用户决策：正式报告开放方案与精确内容/Word正式批准；这是生产功能缺口，不是生产发布授权。所有准入均通过前保持NO-GO。