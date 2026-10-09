# 002 新建默认新版、历史保留：本地完成（2026-10-06）

## 最新纠正：新默认前端已发布，完整业务验收仍阻断

11:36:59Z仅staging前端原子发布，freshbuild0/393公网逐大小SHA/index353c9fb3…/受限备份和回滚有效，backend8baa…/PID17552不变、服务restart0。下方“未发布/SSH阻断”保留为历史，由本次实际限定解除。

首批4正常默认UI配置/4draft人员/4freeze、三140保存/两manual/一native通过，第二报告Error未确证cause，自然尚未到期即exactcleanup0/final0。最终批原15min正常首页等待timeout/native1、业务0，不能以用户确认当真实getInfo200；四完整UI、自然/native4及独立oracle未通过，不再试改或新登录。12:46freshend0/11-private0/旧465-source-schema-cache/历史3-2-1/641-Go215同，formal-prod关/mainraceUNPROVEN。[完整实际证据与阶段表](management-traits-default-staging-result-20261006.md)。

## 最新发布接续：已获批准，SSH连接前阻断

- **本轮未发布，四新完整UI/自然3例/native4/续答/历史只读均未开始。** 用户已明确批准仅20.200.136.133 staging前端统一发布，不后端替换、不重启任何服务、不生产；下方“须另确认发布”为历史，本次授权保留。用户转交独立CodeReviewer PASS，不将其写为本执行者的新审阅。
- fresh严格SSH首次及唯一重试均连接前timeout：**native exit255/255**、10033/10028ms、stdout0；stderr各67bytes及相同SHA，authDenied=false。父Node exit1；远端只读hostname/id命令没有启动，两次上限后停止，不再握手或经HTTP绕行造数。[原生通信证据](../scripts/test/results/mng-default-staging-20261006-ssh-blocked-2a61c8/ssh-readonly.json)。
- 候选dist SHA/inputs一致性门禁和远端当前baseline未执行；旧backend8baa…/front3b83…、11表/private0、旧465、源140/700、Schema/cache/PID均是历史证据，不能当本轮fresh核验。无备份、上传、切换、SQL、浏览器业务或restart，新资源0/cleanup_required0，不声称远端清理PASS。
- 当前唯一active阶段是前端发布blocked，其后UI及自然/native为notstarted；无后台验收任务或新owned待清理。恢复本机到staging TCP/22可达性后沿已获批准续作fresh基线→候选门禁→受限备份/前端原子发布→完整真实验收，不重复申请相同发布许可。业务源/driver/旧收据未修改。

## P1限定纠正：清空后同库重选（本轮仅local GREEN）

- **未完成先列：独立 CodeReviewer 未执行，当前工具无该 agent 调用能力；发布前必须由主协调者派发独立只读复审。新默认仍未部署，远端完整UI/native、正式内容批准及主race不因本轮解除。** 本地浏览器窄路径未完成：初次缺隔离模块，随后三次清空等待断言超时，均保留失败；不能声称本轮六case GREEN。
- 用户转交的 CodeReviewer 确证P1：原清空分支残留 `row.repoCode`，真实新建00201/00202在未手动取消时 `select→clear→same select` 得到 `true,false,false`。下方旧“清空后默认”/81项GREEN仅为前轮已测范围，不包括本缺失组合；历史记录与旧RED/收据不删除。
- [真实SFC新增回归](../Go-based%20Refactored%20System/ruoyi-ui/tests/unit/management-traits-admin.spec.js#L181-L199)先执行产品未修改版本：**83pass/2fail/native exit1**，两个失败均精确为上述布尔序列；另外两个连续同库显式opt-out测试已通过。[有效RED](../scripts/test/results/mng-default-clear-p1-20261006-321f35cf/red-valid.json)保留；前两次JSON解析错误不计业务RED，第二次错误收据也保留。
- [唯一产品patch](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L802-L806)：仅clear分支增加响应式 `row.repoCode=''`，清除选择身份；不改默认资格、保存/冻结、历史回填、冻结只读或非002/00401门禁，不引入watcher或以空code绕过资格。连续同code显式取消仍false；清空后合法同库重新true。[专项GREEN](../scripts/test/results/mng-default-clear-p1-20261006-321f35cf/focused-green.json)：admin **85/85**（原81保留＋4新增），既有六相关文件 **224/224/native exit0**。
- 完整前端Vitest仅执行一次既有任务：**32files/389pass/0fail**；任务接口未暴露numeric exit，不伪填0，也不为取数字重复全量。[npm生产构建](../scripts/test/results/mng-default-clear-p1-20261006-321f35cf/build.json)真实 **exit0/Build complete**，原两体积warning及Browserslist年龄提示保留，无配置/依赖升级；source/test编辑器diagnostics0。本地index SHA `353c9fb3a0f30540c8ca091885b67bf458249c1023eca909f78c83988bf27de2`，未上传。
- 本地浏览器复用原六case夹具、仅内存加入四新建的clear/same-reselect步骤，永久driver未编辑。[最后失败](../scripts/test/results/mng-default-clear-p1-20261006-321f35cf/browser-final-attempt.json)实际clear DOM为 `002管理特质TEST链 / class=el-checkbox`（已取消勾选），cases0/closed true；同native stdin只读探针确证注入中文被转换为 `002????`，两个等待选择器均错误。因此为测试传输编码失败，不作为产品SFC回归失败；三次同类失败后停止，不第四次试跑、不扩预算、不称六case通过。原前轮六case历史PASS保留，本轮不替代远端验收。
- [SHA范围终验](../scripts/test/results/mng-default-clear-p1-20261006-321f35cf/scope.json)：682项仅本表单及既有admin测试变化，**215Go逐SHA不变**、其他前端源码/永久driver变化0；产品规范化diff精确一行。未Go测试/build、SQL、远端数据、部署、重启、Legacy操作、无关格式/旧21rollback；inspect已修driver不再改。必要记忆/交接仅追加本纠正，C区ignored新收据不覆盖旧RED。
- 主待办限定：local1修复/RED→GREEN completed；local2 SFC专项/全前端/build completed，浏览器窄路径blocked单列；stage3 notstarted（须独立CodeReviewer审查本新增行/四回归及用户另确认frontend-only staging发布）；stage4本轮文档completed。无待办工具，未调用task生命周期，不将全部新默认/fullUI写为完成。

## 未验与发布边界

- **新默认 UI 尚未部署 staging，不能称远端已验证默认入口。** 本轮没有上传、切换前端、替换后端或任何服务重启；production 未连接。
- 新批四组合完整参与者/结果/PDF/native 尚未重新验收。原 uf054 批次仍 blocked、cases/native 空、finally completed/ownedResidual0；不回填旧失败，不重跑560题或25分钟以掩盖本地入口尚未发布。
- 主 HTTP/Worker 真实争用仍 UNPROVEN；正式内容/正式报告门禁不解除，当前新版仍 TEST。
- 当前 BreakGlass 模式禁止派发其他 agent，本轮仅源码自审，不冒独立 CodeReviewer PASS。

## 已确认政策与唯一产品增量

用户已明确选择 **“新建测评默认新版，历史保留”**。范围仅00201/00202；历史评分、旧PDF、00401与MBTI不改，不实现历史迁移/重算/回填或全应用重写。

[表单默认入口](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L790-L817)只在真实新建（route无编辑ID且postForm无已保存ID）、有现有管理员权限、legacy+legacy/joinType1、单一有ID物理00201/00202题库、140单选及其他题型0时勾选既有 TEST 控件。复用原选择方法固定25分钟和关闭受测者报告；不改API结构，不把UI选择作为后端资格证明。

个人字段原子集保留，非法身份证/部门配置仍由既有合同拒绝，不自动清除。清空或切非002清除新建选择；切另一合法002重新默认，同库再次事件保留手动取消。历史详情回填不触发默认；原冻结配置仍只读/selected true，探测失败仍禁止保存。保存仍是原Save，后续冻结必须另外确认；取消只保留draft，不替用户冻结。

[提示文案](../Go-based%20Refactored%20System/ruoyi-ui/src/views/exam/exam/form.vue#L126-L129)明确新建默认、可取消、历史不自动切换、TEST与显式冻结限制。没有改题库名称/代码映射、源题、V67/V96或正式内容批准。

## 影响清单

| 消费方 | 本轮处理与理由 |
|---|---|
| form.vue RepoSelect change→repoChange | **同步修改**：仅创建事件默认选中；按真实子组件change先于input顺序，用e.id校资格，不等错误的旧row.repoId |
| handleManagementTraitsSelection/validateManagementTraitsConfiguration | **无需修改**：复用原25分钟、字段合同和140单库限制；未放宽用户字段或冻结资格 |
| fetchData、历史编辑/已冻结编辑 | **无需修改**：没有加入watcher或详情回填自动选中；route/postForm双ID及原只读门禁保护 |
| handleSave/submitForm、saveData、freezeManagementTraitsProfile | **无需修改**：签名/请求/响应不变，冻结依然单独确认，取消/失败不自动重试 |
| API/Go profile、人员、组卷、提交、Worker、结果、报告与续答 | **无需修改**：仍只由真实profile/快照/令牌决定新版归属；默认只改变前端创建选择 |
| 00401产品配置/MBTI/历史评分和PDF | **无需修改**：精确code及非competency门禁；本轮Go215逐SHA无变化，仅一个frontend src改变 |
| admin SFC回归 | **同步修改**：原“002选择不默认”是旧政策，换为新政策可执行断言；新增历史、资格、取消、冻结确认分支，不删除历史执行收据 |
| fullUI取证启动与新C区合同 | **同步修改**：下述Node CLI测试故障，非产品行为；不改SQL/超时/权限/清理器 |

没有共享工具签名、API返回结构、公共配置/环境变量或SQL变更。

## RED→GREEN 与本地真实浏览器

| 层级 | 已验证事实 | 边界 |
|---|---|---|
| 原未修改产品SFC RED | 隔离夹具75pass/6fail，native exit1；新建两个002、切换、取消及冻结确认缺默认失败 | [有效RED](../scripts/test/results/mng-new-default-local-20261006/red-isolated.json)保留；首73/8含两测试夹具错误不混计 |
| 实际SFC GREEN | **81/81，native exit0** | [GREEN](../scripts/test/results/mng-new-default-local-20261006/green-admin.json)；外部API mocked，不是真实DB |
| 既有全前端任务 | **32文件385项通过**，原基线32/363 | 任务接口未暴露numeric exit；不伪填。含人员/参与者/结果/报告/续答/00401回归，不是全产品E2E |
| 本地生产构建 | **native exit0 / Build complete** | [构建](../scripts/test/results/mng-new-default-local-20261006/frontend-build.json)；原两体积warning与Browserslist年龄通知保留，无依赖升级 |
| 当前构建真实Chromium | **6case、0pageerror、native exit0**：两code×开放/封闭四创建，不点击TEST勾选就已选；正常保存→取消独立冻结；历史old false/冻结true两编辑 | [新浏览器收据](../scripts/test/results/mng-new-default-browser-e81796d102da/summary.json)明确LOCAL_BUILT_UI_MOCK_API；不冒真实登录、DB、staging/default远端验收 |
| 原Go历史隔离回归 | 编辑器**253pass/0fail**，三既有guard/service/router测试文件 | 工具未提供skip/native exit，不伪填；不是实库；本轮Go源码未修改，不重跑6176全量或Go build |
| 语法/编辑器 | 三JS语法numeric0；五代码文件diagnostics0 | Vue/SFC通过实际Vitest及production构建 |

本地构建index SHA **ac7684e2dce710c9bc97fa72686ffaea35aed9e99d261929ddbac739cd2d8493**，不是线上3b83…。截图实看新建勾选与历史未勾选/35分钟保留；新建截图采于取消框淡出阶段，不能将截图中残影当冻结已完成。未扩张为手机完整布局审阅或全表单视觉重设计。

既有UI清单已由原测试覆盖：创建/编辑、精确draft人员准备、单独冻结、参与者登记或封闭登录、准备/开卷/答题/同卷恢复、管理员续答、结果/13维4模块、显式TEST生成/查看/下载。**本轮新增真实浏览器仅配置六case**；新默认部署后的四组合完整新批仍欠，不新增冻结后人员维护或未经确认的业务功能。

## fullUI runner exit:null 已证根因与修正

原08:51失败发生在第四人员新增成功之后；active仍prepare-owned-tester-primary202是旧阶段标签，不是该新增失败。新只读诊断仍原240秒上限，**240008ms / null / SIGTERM / ETIMEDOUT**，263bytes输出；没有inspect子收据。纯本地原生Node16复现显示实际进入Debugger attached并等待，说明首位置参数inspect被当Node调试命令，SQL尚未启动。

首次只加 `--` 仍被真实Node16同样解析，失败保留；没有增加预算。最终[启动器](../scripts/test/management-traits-full-ui-staging-20261006.js#L18-L39)使用固定非保留位置参数 `mng-evidence`，loader剥离哨兵后保留mode/mark/四精确ID；增加固定evidence阶段、signal、error.code和elapsedMs，错误正文只hash，不输出凭据。SQL、240秒外限/120秒SSH、两次连接超时重试规则及旧清理器未改。

[真实启动合同](../scripts/test/management-traits-full-ui-harness-contract-test.js)仅替换子payload为合成argv，不SSH：baseline0/inspectnull RED→最终 **baseline0/inspect0/整体0**，四ID保留。中间沙盒缺facts与仅--仍失败收据全部保留，三次测试编辑后停止追加。

**09:16:06Z只读真实复验**：使用修正后的真实evidence函数，仅对原已清理四主键inspect，独立新本地目录，不覆盖旧收据；首次SSH native0、父0、1407ms、stderr0，返回 **FK_ENABLED=1**，tester/paper/report/bundle查询均无行。无新业务数据、备份写入或清理。本次不是完整全库健康/旧465/配置/进程hash复核，也不是本轮四case验收。[实际复验](../scripts/test/results/mng-new-default-local-20261006/inspect-corrected.json)。

## 范围收口与失败保留

[本轮SHA终验](../scripts/test/results/mng-new-default-local-20261006/scope-final.json)基线786项：仅一表单源码、一既有SFC测试、一既有fullUI脚本变化；215Go无变化。两个新C区测试脚本与本报告按归属存放，产物在ignored results。Legacy快照0文件，只承诺操作0，不伪称不存在目录的整树字节核验。原uf054 summary仍08:51:49结束/blocked/native0/ownedResidual0、目录20文件，未写原收据；未Git提交/push或workflow状态操作。

工具失败保留：PowerShell非ASCII筛选正则变问号，发生在真实RED收据落盘后的显示阶段；第二次写wx拒EEXIST，旧RED未覆盖。首SFC夹具共享repo污染/clearValidate stub缺失已用第三次测试编辑纠正；新浏览器首四创建通过后历史标签括号被复用正则当分组，定位timeout，只改测试标签匹配后六case通过。原三个构建warning类别不抹除。

本轮浏览器/context/server均关闭；进程布尔扫描仅两既存node，本轮default-browser、contract、inspect标记和debugger均false，不kill旧worker/用户会话。

## 下一步需一次明确确认

汇总本地完成后，**另确认只发布本次前端到20.200.136.133 staging，不替换后端、不重启任何服务**。然后先当前基线/受限旧dist备份与回滚/完整差异核验→统一前端发布→真实默认创建最小往返→四新组合完整UI/native＋独立SQL/PDF oracle→精确finally。

现有fullUI保留原641/source及旧发行前端基线门禁，当前新default源码与旧基线不同会停止；须在新前端获批且发布实证后建立**新的独立批准scope**，不能重置旧after/旧收据或把显式勾选旧suite称新默认已验。自然三例已有历史实证，不为默认入口重复25分钟；主race仍另列UNPROVEN，正式报告/production不自动开放。