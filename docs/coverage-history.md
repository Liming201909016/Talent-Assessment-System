# 覆盖率与验收历史

> 记录可重复验证的测试数量和关键业务门禁。代码覆盖率百分比仅在实际生成coverage报告后填写，不推测。

| 日期 | Go全量 | 前端Vitest | staging关键门禁 | 未关闭项 |
|------|--------|------------|-----------------|----------|
| 2026-07-24 | 通过 | 8文件72项 | 001/002/003迁移、创建配置、核心安全分流 | 胜任力运行链、题库、结果、导出 |
| 2026-07-25 12:00 | 通过 | 9文件76项 | 创建/发布/答题/提交/删除、结果排序详情 | 00401维护、动态导出、容量链 |
| 2026-07-25 19:00 | 通过 | 13文件91项 | 00401展示/编辑/导入、维度维护 | 动态导出、Worker并发、容量链 |
| 2026-07-25 22:50 | 通过 | 14文件93项 | 三Sheet动态导出、Worker部分/零答和并发提交 | 100份真实容量链、阶段6负例 |
| 2026-07-25 23:20 | 通过 | 14文件93项 | 48维度×384题×100份容量链；安全/快照负例；001/002/003 smoke | 数据库恢复演练；正式报告/PDF延期 |
| 2026-07-26 10:45 | 未重复运行 | 15文件95项（前序已通过） | 胜任力答题页3条Playwright E2E：响应式、保存恢复、认证/交卷门禁，最终退出码均0 | 全系统管理端浏览器回归未在本轮执行；正式报告/PDF延期 |
| 2026-07-26 12:10 | 未重复运行 | 15文件96项 | FB-067完整胜任力结果通过路径参数打开测试报告；production build成功 | 尚未部署staging；正式报告/PDF延期 |
| 2026-07-26 13:45 | 通过；全仓语句10.5%（handler 8.3%、service 33.8%） | 17文件100项；语句/行58.9%、分支93.87%、函数37.5%（Vitest配置范围） | temp-v1临时报告Schema/PDF/下载/审计/删除；管理端7流程E2E；FB-068～071 | 客户正式文案与正式题库仍为外部依赖；production未部署 |
| 2026-07-26 17:31 | 通过；Windows build通过 | 17文件101项；production build成功 | REQ-048/049结果开始/完成时间、答题时长和维度得分合计本地RED→GREEN | 尚未部署staging；继续实施报告产品缺口 |
| 2026-07-26 18:04 | 通过；Windows/Linux build通过 | 17文件103项；production build成功 | 00401参考模板、REQ-057/058/059/061；FB-072真实PDF分页RED→GREEN；9页PDF生成/下载/审计/视觉验收 | temp-v1仍非正式人才文案；SC-012双受众实证待补 |
| 2026-07-26 18:18 | 未重复运行（前序通过） | 17文件103项；production build成功 | 5个独立配置×40题完整作答；3基层+2领导；5份A4×9页PDF；45页程序/视觉检查；FB-073 RED→GREEN | temp-v1仍非正式人才文案；SC-012要求同答案双受众，尚未由本轮不同分数样本关闭 |
| 2026-07-26 18:55 | 通过；Windows/Linux build通过 | 17文件104项；production build成功 | 00401双示例模板；2题预览/导入；386题九列导出；完整40题结果双端点三Sheet一致；临时数据清零 | temp-v1正式文案和SC-012仍待后续处理；production未部署 |
| 2026-07-26 19:10 | 通过；Windows build通过 | 未变更（前序17文件104项通过） | FB-074 RED→GREEN；缺失报告文案错误精确标识contentVersion/audience/dimension/level；REQ-076关闭 | SC-012同答案双受众PDF实证仍待后续处理；本切片未部署staging/production |
| 2026-07-26 19:15 | 未重复运行（前序通过） | 未变更（前序17文件104项通过） | SC-012同答案40/40双受众真实PDF：计分/5维度/顺序一致，A4 9页×2，规范化文本9/9一致，文案精确匹配12/12，临时数据/孤儿清零 | 客户正式题库与正式文案仍为外部依赖；production未部署 |
| 2026-07-26 19:45 | 未变更 | 18文件107项；production build成功 | UF-003/FB-075详情分流RED→GREEN；staging专用结果页9类按钮真实E2E通过，传统控件隐藏9/9，下载PDF 615336 bytes | 详情入口本地修复待staging部署；production未部署 |
| 2026-07-26 20:55 | 未变更 | 18文件107项；production build成功 | FB-075前端部署staging；真实测评管理按标题查询→主“详情”→`CompetencyResults`；排序/详情/40题审计/测试报告/PDF生成下载/返回全部通过，传统控件隐藏9/9 | staging前端index SHA-256=`d9ffb611ccd8de4fc4ae348ba57e58a66a934565ddb7b3ae4383745fbd7abf15`；production未部署 |
| 2026-07-26 21:08 | 未变更 | 18文件109项；production build成功 | UF-004/FB-076旧`exam/users`URL和首页最近测评分流RED→GREEN；staging旧URL+测评管理主详情双入口、专用页9类操作全部通过；传统控件隐藏9/9、传统generate-report调用0次 | staging前端index SHA-256=`c552833bd2149a3a5ae68f1522e9bee4c2c2b58c4bf26cfa1a80ba6f5fd0a5f1`；部署后相关错误0；production未部署 |
| 2026-07-26 21:26 | Go全量通过；Windows build通过 | 18文件112项；production build成功 | 00401结果页本地对齐传统详情布局；新增姓名/电话/完成状态筛选、查询/重置、完整答卷选择、批量生成/下载、查看/答题详情/下载；后端与前端先RED后GREEN | 尚未部署staging，更新后的全按钮E2E脚本待部署后执行；production未部署 |
| 2026-07-26 22:20 | 前序全量通过；Linux build通过 | 前序18文件112项通过；production build成功 | 00401最新后端/前端部署staging；旧URL与主详情、查询/重置/完整性、排序/维度、5维度/40题详情、查看、批量生成/下载、行下载、返回全部真实E2E通过；传统generate-report调用0次 | 后端SHA-256=`c9adf6df61a12fbb7aab607cfb4727f5f2ff88a866d372f006697e43b507d74d`；前端index=`fd7696c5b56302033e4e70fd7264d6697afd2a01ee19ef737be027f09e587950`；关键错误0；production未部署 |
| 2026-07-26 22:33 | 未变更 | 19文件114项；production build成功 | FB-077本地RED→GREEN：HTTP 200 JSON错误Blob不再保存为伪PDF；正常application/pdf仍原样返回；行下载显示后端错误 | 本地index SHA-256=`543fbdb8ad21c7faa58cff1bca44c48f0ecbe435bfdb437ded5fca515709320c`；尚未部署staging；production未部署 |
| 2026-07-26 22:40 | Go全量通过；Windows build通过 | 未变更（19文件114项） | FB-078本地RED→GREEN：低权限后台用户访问胜任力结果分页、逐题详情和管理员报告数据均在查询前返回HTTP 403；管理员/exam:list/exam:export矩阵通过 | FB-077/078均尚未部署staging；production未部署 |
| 2026-07-26 22:50 | Go全量通过；Windows build通过 | 19文件115项；production build成功 | FB-079本地前后端RED→GREEN：内部报告API仅接受`X-Internal-Token`，正确query token也返回401；前端请求query仅含paperId | 本地index SHA-256=`4b6f85721b137cec28fb7943e781df81c21015aa150c1f9bd1e628f62e958724`；FB-077～079尚未部署staging；历史nginx日志已有24行旧query请求，部署后需确认新增为0；production未部署 |
| 2026-07-26 22:59 | Go全量通过；Windows build通过 | 未变更（19文件115项） | FB-080本地RED→GREEN：同paperId 8并发临界区最大并发数=1；稳定64分片索引有界；查询/渲染/落盘/替换/审计全程串行 | `go test -race`因本机`CGO_ENABLED=0`未执行；普通并发专项、Go全量和build通过；FB-077～080尚未部署staging；production未部署 |
| 2026-07-26 23:05 | 未变更 | 19文件117项；production build成功 | FB-081本地RED→GREEN：列表旧响应和详情旧响应均不得覆盖最新筛选/人员；列表请求冻结query快照；仅最新请求控制loading | 本地index SHA-256=`a055a9b08951551ea770a8d4fb276bddee77963d06b0e4d77f1d8c9794f22217`；FB-077～081尚未部署staging；production未部署 |
| 2026-07-26 23:14 | 未变更 | 19文件119项；production build成功 | FB-082本地RED→GREEN：批量生成/下载启动时冻结完整答卷目标；运行中清空或替换表格选择不改变处理对象、进度及成功数 | 本地index SHA-256=`f22ba02e9bf143a05dad544a8dfdc66f9cffa2dc04469742094e85aa4a6618a2`；FB-077～082尚未部署staging；production未部署 |
| 2026-07-26 23:29 | Go全量通过；Windows/Linux build通过 | 19文件122项；production build成功 | FB-083～087本地RED→GREEN：JSON绑定、成功审计原子性、唯一/RFC5987文件名、E2E真实计数、移动端全屏单列详情；最终E2E脚本语法通过 | 后端SHA-256=`fcfaa85819702b8f9ab333e1f4ef834fe4bd858464098f740d7dc1cb29247348`；前端index=`fa12099adef2e656a9d12a47338a7f810c5ea18ac57ae0dfd7952f7523ee3787`；FB-077～087待staging部署验收；production未部署 |
| 2026-07-26 23:40 | 同上 | 同上 | FB-077～087部署staging；全按钮+API负例+390×844移动E2E通过；低权限三端点403；同paper双并发重生成均completed；实例1、当前PDF1、同paper文件1；真实内部token新增query日志0 | 数据库备份SHA-256=`cf81d677926d7c261741ac5ba771219fbc5d5e58bb89e2182265fafa7988b40a`；服务/health正常、关键错误0、临时状态0；production未部署 |
| 2026-08-09 | Go全量通过；Windows build通过 | 23文件132项；production build成功（2个既有体积warning） | 通用胜任力产品/评分/内容/模板版本：草稿默认与校验、发布冻结、结果快照、报告精确匹配、legacy清空；真实Gin HTTP非法产品版本在数据库前拒绝 | 007迁移静态幂等检查通过；本地MySQL `127.0.0.1:23306` 拒绝连接，真实迁移与回填未验证；未部署staging/production |
| 2026-08-10 | 前序Go全量通过；Linux build通过 | 23文件132项通过；production build成功（2个既有体积warning） | 007在staging连续执行2次；8个版本列、9个胜任力配置、10个结果、7个报告实例完成兼容回填；分页/详情API及管理表单真实显示四类版本；不支持产品版本真实请求拒绝且写入0行 | 数据库备份SHA-256=`0f31bf3dea7e67292406f1732f19c982b2ea80624335f6d753e3722cfe30f11c`；后端=`44df29bcf09a55ab4a1d61b94d408d9eb649cdc40645ea5c48e75753cf063fc7`、前端index=`1e60ac06cfb4ff219428151d91b1a6f3231001ff8748ea0bb80228ea34ad5163`；service/nginx/mysql/内外health正常，关键错误0；production未部署 |
| 2026-08-10 17:10 | Go全量、`go vet`、Windows build通过 | 23文件137项通过；production build成功（2个既有体积warning） | I2固定一期配置本地RED→GREEN：基层员工、十维、四版本、每维8+1库存；前端隐藏可选受众/维度；一期运行时完成前前后端禁止发布 | 本切片仅本地完成，未部署staging/production；下一依赖为一期五档、一级聚合和效度运行时 |
| 2026-08-10 17:20 | Go全量、`go vet`、Windows build通过 | 未变更（前序23文件137项） | I3一期纯评分引擎RED→GREEN：二级L1-L5精确边界、十维`sum(scoreSum/8)`、总体25/32.5/40/45五档、顺序无关和不完整不出正式分 | 提交持久化接入仍保持❌，等待一级聚合与效度结果一起接入；发布门禁保持关闭；未部署staging/production |
| 2026-08-10 17:30 | Go全量、`go vet`、Windows build通过 | 未变更（前序23文件137项） | I4一级聚合纯函数RED→GREEN：固定通用能力/心理素养各5维、精确平均、L1-L5边界、不完整计数与畸形输入拒绝 | group snapshot/result持久化仍保持❌，等待效度算法后统一接入；发布门禁保持关闭；未部署staging/production |
| 2026-08-11 | Go全量、`go vet`、Windows build通过 | 未变更（前序23文件137项） | I5效度纯函数RED→GREEN：10道原始分正向累加、35/36边界、10/50极值、9/10未完成和8类畸形输入拒绝 | validity持久化与默认统计隔离仍保持❌，下一步统一接入90题发布/提交；发布门禁保持关闭；未部署staging/production |
| 2026-08-11 09:27 | Go全量、`go vet`、Windows build通过 | 23文件139项；production build成功（2个既有体积warning） | I6本地RED→GREEN：2组/10维/90题发布、80/10拆分、10+2+1+1结果、NULL不完整分、效度筛选与排名默认、扩展导出、发布权限、一期报告门禁；staging E2E脚本语法通过 | 尚未部署staging/production；一期正式报告渲染器仍关闭 |
| 2026-08-11 09:35 | Go全量、`go vet`、Linux build通过 | 23文件139项；production build成功（2个既有体积warning） | staging真实90题发布→组卷→全答→提交→结果→筛选→三Sheet导出；2/10/90快照、10+2+1+1结果、3/L3、30/weak、10/good、重复请求幂等、清理0、传统签名不变 | 一期正式报告渲染器仍关闭；timeout不完整统一运行时本轮未在staging重复造数；production未部署 |
| 2026-08-11 10:31 | 前序Go/构建结果不变 | 前序23文件139项不变 | staging三组负向链：88/90 timeout产生overall/二级/一级/效度NULL并拒绝报告；40/questionable显式可查但默认排名排除且导出正确；快照INSERT强制失败后发布事务零残留、同草稿可重试；最终清理与传统签名通过 | 独立常模/汇总统计端点及一期正式报告渲染器仍未实现；production未部署 |
| 2026-08-12 20:15 | Go全量、go vet、Windows/Linux build通过 | 24文件；新增模板管理3项；production build成功（2个既有体积warning） | 统一报告模板页；管理员权限；49/12/0严格校验；非法Tag拒绝且SHA不变；合法上传备份+生效；真实API下载；1440/390浏览器下载/上传/移动无溢出 | staging后端=`210aac516088e610e8b5be6307391e59062f0a0d3053006277843c902f72d639`，前端index=`1afdb1ce766546584a2c1c1e5d0d492bead56d54b98e0653df2394a38fa0a80c`；production未部署 |

## 当前已验证规模

- 胜任力维度：48。
- 测试源题：384，每维度8题。
- 单测评发布快照：48维度、384题。
- 容量链：100名参与者、100份试卷、38,400条试卷题。
- 随机性样本：100份持久化题序，100个不同SHA-256；100次刷新全部稳定。
- 前端：18个Vitest文件、107项测试；胜任力答题页3个Playwright E2E文件；管理端7流程E2E；胜任力结果页9类按钮E2E。
- Go：`go test ./... -count=1` 全量通过。
- 五报告样本：5个独立配置、200道完整作答、25条维度结果、5个报告实例、45页PDF，孤儿0。
- 双受众对照样本：同答案40题、2个报告对象、2份A4×9页PDF；整体/维度计分一致，受众文案精确匹配12/12。

## 仍需补充的量化数据

- 传统001/002/003完整浏览器答题链数量（当前已有真实API smoke；管理列表与公共页面已覆盖）。
- 客户正式文案与正式PDF覆盖率：等待客户内容后建立；temp-v1临时PDF链已覆盖。
