# 2026-10-06 历史新版报告资格与 UF053 staging 发布预检

## 1. 未完成与最终待办

- **发布 BLOCKED**：当前整包相对线上 8fb264e… 的最小观测限定未获可追溯证明；旧 21 项源码漂移仍未归因。现有 UF053 磁盘收据明确独立审阅 NOT_EXECUTED，不能把用户交接中的 reviewer PASS 当作已独立核实的磁盘证据。
- **历史新版 PDF 未生成**：现有新版 TEST 接口不支持这些历史 legacy 卷；本轮没有授权历史迁移、重算、回填或覆盖报告。
- **管理 UI 未验证**：原 page/context 清除 routes 后，07:37:44.544Z 真实 getInfo HTTP401/code401；cookie 存在但不是有效认证。原首页保留，不登录/注销/reset、不伪造 JWT/Redis 会话或导出令牌。
- 主待办最终为：1 history **completed**；2 publish **blocked**；3 report **not_started**。执行时只有 history 或 publish-preflight 一个 active；最终 active=0。无可用 todo 接口，未声称调用不存在的接口，未修改现代化 workflow。

## 2. 本次批准的精确边界

用户于 2026-10-06 明确“批准发布”，承接 UF053：仅 staging 20.200.136.133 后端最小安全单卷观测、一次维护 restart，以及未来指定专属合成卷观测。既有授权保留，不重复索取同范围授权；但授权不解除源码来源/范围门禁，也不扩大为其他业务变更批准。

生产、前端、共享 cache/权限表、旧卷 snapshot/profile/result_bundle 创建、历史 DDL/DML/recompute/backfill、V67/V96 修订和旧 PDF 删除均禁止。询问“之前的测试是否可以生成新版报告”只授权资格核查，不授权生成或迁移。

## 3. 链接实际含义与 API 消费链

目标链接为 `/#/exam/exam/users/1776822816300709851/1/0422测管特基层-开放版`。

- [实际路由](../Go-based%20Refactored%20System/ruoyi-ui/src/router/index.js#L417-L422)将第一参数命名 examId、第二参数 isOpen；不是 paperId、participantId 或新版 runId。
- [组件入口](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L256-L303)先获取测评详情，再按已知新冻结 profile 和管理员资格分流；并不是凡 00201 历史卷都进入新版结果。
- [详情 API](../Go-based%20Refactored%20System/ruoyi-ui/src/api/exam/exam.js#L7-L9)是 POST /exam/api/exam/exam/detail，body.id；[真实 handler](../Go-based%20Refactored%20System/internal/handler/exam.go#L262-L309)读取 el_exam 及 el_exam_repo→el_repo，不按 URL 标题推定题库。
- [开放列表选择](../Go-based%20Refactored%20System/ruoyi-ui/src/views/user/exam/index.vue#L348-L372)调用 listTester；[API](../Go-based%20Refactored%20System/ruoyi-ui/src/api/tester/tester.js#L30-L36)是 GET /exam/api/tester/tester-list，而不是 candidate POST 同名接口。[真实 handler](../Go-based%20Refactored%20System/internal/handler/tester_list.go#L47-L112)查询 el_candidate，按 examId 筛选，人员通过 paper_id 与 paper 关联。
- 当前 UI 认证401后没有继续请求上述业务 API；以下真实数据取自 SSH 限定只读 SQL，不冒称已验证页面或生成按钮。

## 4. 真实历史数据资格

两次只读 SSH 首试各 exit0/stderr0：07:37 后元数据查询及 07:39:43Z 聚合与维护安全查询；使用既有 liming/key、StrictHostKeyChecking=yes、BatchMode=yes、ConnectTimeout10、ConnectionAttempts1，远端 hostname vm-ubuntu-go-dev。sudo/socket 只读消费既有认证，不输出 DSN/密码/用户 PII；查询显式 READ ONLY transaction。

| 事实 | 实际查询结果 |
|---|---|
| 精确 exam 主键 / 标题核验 | 1776822816300709851，行数1；用户给定标题精确匹配1 |
| 实际类型 | is_open=1、legacy/legacy、25分钟 |
| 实际题库 | 唯一查询行 code=00201，配置140单选题 |
| 有效开放人员 | 3个 candidate，3个有 paper 关联、2个 end_time 非NULL、2个非空旧 pdf_path |
| paper 状态 | state2 两卷；state0 一卷（不是正在作答 state1） |
| 同卷真实身份 | 三卷各 candidate 同 paper+exam 恰1，tester 同关联0；不使用 paper.user_id 推定归属 |
| 两完成卷 | 各140个唯一源题、140已答、140 actual_score 在1～5；700选项桶、140勾选 |
| 未作答卷 | 140个唯一源题、0已答、700选项桶、0勾选 |
| 新版链 | 目标 profile0；三卷各新版 paper_snapshot0/result_run0/report_revision0；全库11张 el_mng_ 表逐表0 |

paper UUID 不输出，取 `SHA256("history-paper-v1" + NUL + paperID)`：完成卷 225dd5bd8f33e8ddc257143848de054f008f38a2e20af3b011e3115beb7f0cbf、3c4c38e206e4e74db0b3b0c8126a8b858e42fbdcfc401c6d62d3630e9c5dff0a；state0 卷 4ee4c80f32b0f139c98cb60223dcb2e26b75719e336b4a0787a967ffe88f85da。

### 选项审计纠正与严格证据边界

首次选项聚合错误地将桶 qu_id 关联到 paper_qu.id，产生 NULL；**该审计无效，不当作历史选项缺失**。随后实读[旧组卷赋值](../Go-based%20Refactored%20System/internal/handler/paper.go#L377-L409)，确定桶 QuID=a.QuID，即源题 ID；第二次按 paper_id+paper_qu.qu_id 关联后真实 SQL 成功：

- 完成组280题：每题5桶、5个不同1～5分、恰1勾选、所选桶score与actual_score一致，全部280；未作答组140题每题5桶、不同1～5分，勾选0。
- 存储桶 is_right=1 每题恰1；完成组标记score1/5为80/200，未作答组40/100。它是历史桶中保留的标记事实，**不是已确认的新版 V/13维映射、题干或题本版本证据**。
- 三卷共420题 exam_question_id/raw_answer/final_score 均NULL。没有读取姓名/手机号/密码/旧PDF实际路径或当前源题正文，没有把当前140题源库当作旧卷原始题本证明。
- 旧 pdf_path 非空只证明数据库有旧文件引用；本轮未下载/逐SHA核验旧两PDF，不承诺其当前文件存在性或渲染能力。

**资格结论：现接口不能直接生成这些历史卷的新版报告。** 两完成卷的答案有完整性基础，但缺新版可信冻结题本/身份/版本/hash及独立 run；第三卷未答更不满足140/140。不能把旧标准分线性硬转成新版百分制，也不能仅用当前题库补造历史证据。

[Generate handler](../Go-based%20Refactored%20System/internal/handler/management_traits_report_runtime.go#L61-L90)要求明确 runId；[加载器](../Go-based%20Refactored%20System/internal/service/management_traits_report_runtime.go#L114-L156)必须读真实 run、paper snapshot/bundle/140题快照、13维/4模块及receipt；[纯适配门禁](../Go-based%20Refactored%20System/internal/service/management_traits_test_report.go#L153-L167)只接受 new_creation/submitted_snapshot/completed，[元数据门禁](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_load.go#L91-L127)明确非历史当前来源。此为源码合同，不是本轮实际 POST 的409响应；本轮没有调用生成/重算/下载接口。

“新版”若指新评分/新Word TEST链，以上缺口阻止现接口直接使用；若仅指旧评分换样式，也没有本轮已核可用的兼容适配或生成授权，不能回答已可生成。后续历史并行迁移/兼容方案须单独确认并审计原始来源，旧PDF保留；或另建独立新测评按新冻结流程重新作答。两种均未在本轮执行，不承诺历史迁移已实现。

## 5. staging 维护安全与来源门禁

07:39:43Z 实际：未到期 state1 卷0，过期进行中 competency+competency_average 卷0；11张管理特质sidecar逐0/private文件0。三服务 talent-assessment/nginx/mysql active，内部 health {"status":"ok"}。主PID2002，报告两环境为staging，APP_ENV=production仅既有配置选择；观测两键在当前进程均未出现。

线上 server SHA **8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93**，前端index SHA **3b83b976eaec6b260060ae1368d53edbf77b674a2e1efef5de04d0e7d9638f97**。这些是本轮fresh事实；没有维护restart，未新建目标合成卷或启用观测。安全计数0只证明该取样时刻，不承诺未来维护窗口无新增卷。

- [UF053既有verdict](../scripts/test/results/uf053-local-observation-20261006/verdict.json)为6176pass/659顶层/0fail/9skip、Windows/Linux build/vet exit0，属于上一轮本地验证，本轮未重跑，不称真实MySQL/main race PASS。
- 独立重新计算 ended snapshot：641/641当前SHA相同，候选49902830bytes、SHA **8baa7f87b6ae59279d7c233cf8aa41b38ecf7e32d20ac2c644ae3edf95f12676**与既有收据匹配；没有新增未许可源码变化。
- UF053相对本轮before的六既有文件差异为五service源码及本轮观测测试增强，新增private观测/Gin测试已有原收据。**这不证明 before 本身与线上已部署源码相同**。
- [旧21项漂移](../scripts/test/results/mng-premature-timeout-c90f8a172e64/local-source-drift.json)：21/21仍与漂移后的currentSHA一致，0/21匹配原beforeSHA；其中7个非测试文件，涉及candidate/tester handler、模板handler、report诊断、公共LO client及两个非server命令。当前server整包包含其中后端实现，不能把这组全部说成与发布无关的测试文件。
- 本地旧 server-mng-report-diag-linux 字节SHA精确等于线上8fb…；新旧 Go build info均同一VCS revision 5217ce…且 vcs.modified=true，**相同脏revision不能恢复构建时源码或证明语义相等**。
- 可达 Git 历史限定搜索：candidate8、tester7、LO client1，共16个版本；其余4文件无可达历史。原字节及LF/CRLF/BOM变体均无匹配旧非测试beforeSHA。不是穷尽所有外部备份/不可达对象，也不是已证明它们一定存在行为变化。
- 旧受限 application.tar.gz 实读含后端cmd/internal/pkg的.go源码条目0；最新8fb发布evidence中有checker源码、脚本、旧二进制、数据库/哈希/门禁日志，未发现完整server源码归档。第三次SSH中 SQL审计成功后 archive pipeline exit1/stderr50bytes，未取得原因、不计整体PASS；第四次独立只读archiveprobe使用既有Perl，exit0/stderr0，明确文件存在/source条目0，原失败保留。
- UF053 verdict和final-verification均明确 independentReview NOT_EXECUTED；结果目录未见独立review收据。本模式不能派发agent，不自审冒独立PASS。

**发布 BLOCKED 于来源/范围及独立review证据，不是候选编译失败或历史卷影响计数阻断。** 未执行上传、远端备份写入、server切换、配置写入、stop/restart/rollback；一次维护restart批准预算未使用。不能把授权擅自解释为批准旧未知业务变更后整包发布。

恢复条件：取得与8fb构建对应的原源码/可核归档，或由协调者先明确处理候选额外范围并完成独立CodeReviewer审阅，不能只改文档标签放行。门禁解除后续用本次既有最小发布授权，重新fresh安全计数→受限备份/rollback→仅后端一次维护restart→默认关闭或明确已核专属合成目标→验证。此处仅交接，不自动创建合成卷或再申请第二次restart。

## 6. 本轮范围

仅SSH只读/浏览器getInfo和本地文件/hash/buildinfo/Git历史只读检查；业务Go/前端/Legacy/SQL/配置修改0，远端DML/DDL/SET/GRANT0，历史生成/迁移/重算0，新合成实体0，后台等待/观察任务0，生产访问0。新证据文档属于C区，本报告及两既有文档只追加限定结论，不改旧收据/失败/21项基线。未生成新PDF，cleanup_required0不是执行远端清理PASS。