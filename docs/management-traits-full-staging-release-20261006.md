# 完整候选 staging 发布与新版 UI 验证（2026-10-06）

## 12:46Z 新默认frontend-only后续实际完成，验收仍PARTIAL

新默认353c9fb3…已于11:36:59Z仅前端staging原子发布，393公网大小SHA全部同；backend8baa…/PID17552保持，不新后端build/替换/服务restart。受限新backup与旧dist完整回滚资源保留。

首批正常四default draft/人员/freeze，三个140真实UI/两个manual/一native；第二report步骤Error cause未证，全部owned精确cleanup0。最终独立窗口原15min首页等待TimeoutError，未auth/baseline/造数，不称cleanup执行PASS。自然三例/native四/独立Pythonoracle未完成，formal-prod及mainrace不解除；旧fail不改。12:46鲜活终验SSH0/新manifest/旧465/source/schema/cache/历史聚合/641-Go215同。[详细实证](management-traits-default-staging-result-20261006.md)。

## 09:16Z 最新限定纠正：完整发行已完成，新默认仅本地完成

[纠正 - 2026-10-06] 下方“四组合新版UI尚在执行”是旧进度；uf054-ui-37f80e5d48fd于08:51:49Z已blocked收尾、cases/native空、cleanup completed/ownedResidual0。第四人员实际200/code200成功，随后取证子进程首位置参数inspect误入Node16调试器，不能以旧active标签判人员失败。固定非保留哨兵后原生合同GREEN；本轮真实取证仅原已清理四PK只读inspect，首次SSH0/FK1/0行，不重跑整个suite。

用户已选“新建测评默认新版，历史保留”；新表单默认002本地SFC81/全前端385/构建0/六本地mock-API浏览器case通过，**尚未部署staging，不冒线上default已验**。原完整发行后端8baa…/front3b83…与一次主restart记录保留，本轮没有发布或restart。下一只frontend-only staging统一汇总发布须另确认；原641及旧receipt不重置，之后建立独立newscope。[完整本轮变更与失败](management-traits-new-default-local-20261006.md)。

## 当前未验项

- **发布已于07:57:39Z完成；本轮四组合新版 UI 尚在执行，不能沿用旧批次结果作为本轮通过。**
- 主 HTTP/Worker 真实争用仍 UNPROVEN；本次观测默认关闭，不通过多重重启或改期限凑证据。
- 旧 21 个文件的历史修改来源仍未知；最新授权为当前完整已验证候选，不再要求证明其仅为最小插桩差异，也不声称与旧构建语义等价。
- 本执行者完成的是当前完整范围只读安全自审，不是新的独立 CodeReviewer 审阅。协调者提供的前阶段 UF053 独立 PASS 不修改磁盘旧 NOT_EXECUTED 收据，也不扩张其审阅范围。

## 授权与范围

用户明确“完整的发布，然后开始新版测试，包括完整的 UI”。目标仅 20.200.136.133 staging；生产不连接。当前后端、完整前端及必要运行资源作为统一 candidate 核验；线上前端、TEST 模板/内容已一致时保留，并记录完整 manifest，而非无谓改为另一个版本。

历史测评 1776822816300709851 不执行迁移、回填、重算、旧报告生成或删除。仅未来新专属合成测评测试。既有远端配置/密钥/权限表/MySQL global/cache 不改；主服务正常维护重启限一次。

## 当前源码只读安全自审

旧漂移清单 7 个非测试文件均实读当前完整内容；两个构建命令不进入 server，执行器不上传它们。重点如下：

| 当前文件 | 已检查的发布安全边界 |
|---|---|
| [candidate handler](../Go-based%20Refactored%20System/internal/handler/candidate.go#L57-L61) | 新身份优先分流；旧 PDF 持久化压缩在请求冻结 gate 内同步结束，不遗留响应后异步覆盖 |
| [tester handler](../Go-based%20Refactored%20System/internal/handler/tester.go#L193-L215) | status 仍为可空字符串，创建按请求保存；没有把 NULL/未知默认准入，不改密码/旧人员状态 |
| [TEST 模板构建](../Go-based%20Refactored%20System/internal/handler/management_traits_test_template.go#L20-L24) | 来源 SHA 锁定，独立 TEST 包，客户原件/星级/页脚不替换；ZIP 使用新 header |
| [诊断分类](../Go-based%20Refactored%20System/internal/service/management_traits_report_diagnostic.go#L28-L36) | 固定 stage/class，未知错误不格式化、无 token/SQL/PII 输出 |
| [公共 LO 客户端](../Go-based%20Refactored%20System/pkg/libreofficepdf/client.go#L77-L118) | 有界大小、单槽、独立 profile/0700/0600、直接 exec 参数、固定公开错误及受控 cause，未改 shared cache |
| [TEST 内容 CLI](../Go-based%20Refactored%20System/cmd/mng-test-content/main.go#L23-L61) | 精确内容校验，输出与源不同，已有不同文件拒绝覆盖；本包不执行 |
| [TEST 模板 CLI](../Go-based%20Refactored%20System/cmd/mng-test-template/main.go#L25-L58) | 原件不覆盖、不同已有输出拒绝；本包不执行 |

[完整旧接口 gate](../Go-based%20Refactored%20System/internal/handler/management_traits_runtime_guard.go#L26-L67)仍包含人员/PDF/删除/导出/旧答案写入口；[路由安装](../Go-based%20Refactored%20System/internal/router/router.go#L62-L65)在 JWT 前。不能仅因 handler 内有旧兼容分支就称新版 gate 被移除。旧 legacy 路径的 prefix 路径校验等已知风险不在本轮修复、不夸全系统安全消除。

UF053 当前 [观测配置](../Go-based%20Refactored%20System/internal/service/management_traits_race_observation.go#L37-L54)缺两键默认关闭；私有来源 marker、事务后固定有界日志保持。[提交事务](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_write.go#L193-L287)继续 paper UPDATE→bundle SHARE、锁后真实时间和原子结算，不接受客户端 timeout。[到期扫描](../Go-based%20Refactored%20System/internal/service/management_traits_expiry.go#L72-L98)只处理 new_creation；旧历史卷没有 snapshot，不进入这条 Worker。

自审未发现本次候选需新增功能/放宽身份、guard 或路径门禁的理由；此结论不证明旧来源、历史等价、主实库争用或新的独立审阅。

## 本轮已核证据

07:52:33Z 严格 SSH 首试 exit0：主 PID2002、三服务 active/health ok、state1 全0/到期00401待处理0/11侧表逐0/private0/15 RESTRICT FK；历史卷计数3/完成2/未开始1/进行中0，观测键未设。后端8fb264e…、前端3b83b976…。

当前641项与 UF053 ended 快照逐 SHA 一致。07:53 Windows/Linux server、全包 build、vet 四项均数字 exit0，stdout/stderr 均0；同源码真实全Go6176pass/659顶层/0fail/9skip复用，不称本轮重跑/coverage。

证据目录：[本轮起始范围](../scripts/test/results/mng-full-release-20261006/scope-start.json)、[只读维护门禁](../scripts/test/results/mng-full-release-20261006/fresh-readonly-preflight-attempt-1.json)、[候选构建](../scripts/test/results/mng-full-release-20261006/candidate-artifacts.json)。

失败保留：新 preflight driver 首次 Node 语法检查因 Bash `${…}` 与 JS 模板冲突 exit1；在 SSH 之前修正后真实执行 exit0。一次错误 middleware 路径读取失败后按实际 handler 文件核实，未修改业务或推断不存在的守卫。Node16 已实核，公网验证改用内置 HTTP，未安装包。

## 完整发行实际完成

07:56:42–07:57:39Z完整发行原生 exit0/stderr0。新后端49902830bytes/SHA **8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676** 已运行；旧主PID2002完整排空后，新PID **17552**。只一次正常主服务 stop/start，Nginx/MySQL未重启，rollback未使用。检查前与最终完整scope稳定。

393/393前端文件与本地 dist、SSH实际目录及公网HTTP200原字节SHA全部一致；index仍 **3b83b976eaec6b260060ae1368d53edbf77b674a2e1efef5de04d0e7d9638f97**。这是完整发行内“已符合当前版本，保留”而非只验index或只发最小观测。两个 TEST 资源通过真实应用账号读205规则/精确SHA，旧configs/env/dropins/unit全部不变，观测默认关闭。

受限服务器备份 `/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/backend_82155be9457ad1bf` 保留全库（singletransaction/routines/triggers/events）、旧binary及uid/gid/mode、完整dist/configs、unit/dropins、两上传树；目录0700、文件0600，全部gzip/SHA和application/system tar差异核对成功。全库SHA **b034d5ba1d531e5aea842b414c62fe28ab4eac4dd2114121e0d03ff907e10ab9**。不下载秘密归档，不全库恢复；已有恢复演练是历史证据，本轮只核gzip和回滚资源/metadata，不能称新恢复演练完成。

pre/stopped/post三次fresh应用账号Schema首次/缓存 **11/159/67/21及4metadata查询**、空完整canonical scope、CRUD权限、两题本140/700 ASCII和TEST资产全部PASS。内网health200；管理缺认证401、四完整参与者缺token401、空body400合同PASS；general/00101/00201/00301/00401五只读链HTTP200业务成功。DDL/SQL业务写/GRANT/共享配置/历史生成/生产访问0。

数据库legacy12表SHA **051795bd969650d0e55da676083078255331c66405b5fff0ee4ca9d0a7641d0b** 前后相同，465旧PDF逐路径SHA不变，11侧表逐0/private0/15FK；5项exact上传全部删除remaining0，正式备份保留。当前scope641稳定。

前端本轮正常Vitest任务 **32文件/363tests通过**（不等于real UI），已有Vue测试方法覆盖deprecated提示如实保留，未为了发布改测试/业务。Go全量6176/0/9仍同源码复用，四fresh编译/vet0。报告和测试driver只属C区，运行源码改动0。

机器证据：[完整发行manifest](../scripts/test/results/mng-full-release-20261006/complete-release-manifest.json)、[393公网与远端一致](../scripts/test/results/mng-full-release-20261006/full-frontend-retained.json)、[实际备份/排空/安装/验收](../scripts/test/results/mng-full-release-20261006/complete-candidate-publish-1.json)、[已发布限定结论](../scripts/test/results/mng-full-release-20261006/deployment-verdict.json)、[上传精确清理](../scripts/test/results/mng-full-release-20261006/payload-exact-cleanup-1.json)。

## 新 UI 测试前提

07:57:54原集成页清routes后getInfo401，未绕认证造数。用户正常重新登录，07:59:32真实getInfo200/code200/admin。原窗口保留不导出token；为四份真正native Download/saveAs使用已有Playwright1.41.2/Chromium独立窗口再次正常登录，凭据仅浏览器内存。新driver创建单独随机MTH scope，不拿旧批次140/0/3或旧native失败当新通过。