# 全项目待办评估（2026-10-10）

> 评估模式：只读代码、文档与本地构建/测试；未访问远端、未修改数据库、未部署。
>
> 状态裁决：当前事实以 [project-status.md](project-status.md) 为准；历史记录仅用于证据追溯。UF-055～UF-058、00401 production 核心链、005 production TEST 链不重复列为待修复事项。

## 1. 执行摘要

项目当前是“核心业务范围限定可用，但尚未达到全产品、全安全、全运维 GREEN”。

- 00401：staging/production 核心链 GREEN；native download、并发与容量仍未补验。
- MBTI：API、ESTJ 计分及报告链 GREEN；16 类型矩阵和参与者真实浏览器链未关闭。
- 00501/00502：production TEST 链 GREEN；formal 报告仍未批准或实现完整前端闭环。
- 001/002：仅有范围限定兼容证据，下一次修改公共 paper/exam/report 代码前需完整回归。
- production：业务可用，但 TLS、默认管理员凭据、秘密权限、root 服务、SSH/网络面、孤儿数据和磁盘治理仍为 NO-GO。
- 代码层新增确认了匿名部门考试暴露、任意 Origin CORS、路径边界、数据库错误吞掉、N+1、分页上限和前端失败恢复等债务。
- CI 当前不能证明前端、安全扫描已通过；Go 版本也与 go.mod 不一致。

## 2. 本轮实测基线

| 检查 | 结果 | 结论 |
|---|---:|---|
| Go 全量测试与覆盖率 | exit 0；总 statements 45.7% | 较历史 44.3% 上升 1.4 个百分点，但 handler 32.7%、config 3.0%、repository/db/redis 0%、pdfgen 1.6% 仍是主要盲区 |
| Go server build | exit 0 | 本地可编译 |
| Vue Vitest coverage | 35 files / 611 tests 全过 | statements 60.25%、branches 95.53%、functions 39.13%、lines 60.25%，与历史基线相同 |
| Vue production build | exit 0 | 当前磁盘代码可构建；result2.vue 的编辑器 CSS 诊断未阻断真实构建，但 v-for key/lint 债务仍需独立处理 |
| Git 工作树 | 100 条：63 deleted、12 modified、25 untracked | 当前工作树不可直接作为发布源；继续以封存 manifest/SHA 为发布依据 |

## 3. 优先级定义

- **P0**：安全暴露、凭据/传输风险、正式产品上线门禁；未关闭前不能称为全项目 production-ready。
- **P1**：高概率造成数据不一致、功能误导、性能退化或质量门禁失真，应在下一轮功能开发前后分批关闭。
- **P2**：架构、测试治理、UX 和可维护性债务，可按模块渐进偿还。
- **P3**：长期治理和体验优化，不阻断当前范围限定业务。

## 4. P0 待办

| ID | 待办 | 已验证事实/影响 | 前置批准 | 验收标准 | 工作量 |
|---|---|---|---|---|---|
| P0-01 | 005 formal 报告链 | 当前只有明确标记 TEST 的结果/报告；formal registry、内容批准、正式前端入口和权限闭环未完成 | 产品、内容、法务/交付 | formal 内容和模板批准；TEST/formal 路由、权限、数据、文件名、审计独立；真实浏览器生成/view/download/cleanup | L |
| P0-02 | 生产管理员凭据轮换 | 当前仍匹配仓库历史默认候选，存在后台接管风险 | 用户明确批准、维护窗口 | 新凭据登录通过、旧凭据失败、既有会话撤销、秘密不落日志/Git、回滚可用 | S |
| P0-03 | 生产 TLS/HTTPS | 当前登录和业务请求仅 HTTP，无 443/HSTS/Secure Cookie 基础 | DNS/证书/网络所有者 | 受信证书；80→443；登录、刷新、API、PDF、上传下载全链重验；Cookie 策略正确 | M |
| P0-04 | 秘密与服务最小权限 | secret-bearing 配置 0644；有效 Go service 以 root 运行；systemd 9.6 UNSAFE | 运维/安全、维护窗口 | staging 先完成 0600/Secret Store、专用用户、写目录清单和 sandbox；production 单次迁移；有效 unit 与仓库 SHA 一致 | L |
| P0-05 | 修复匿名部门考试暴露 | `OnlinePaging()` 是匿名路由，明确把 open_type=2 一并返回；部门限定考试可被非目标用户发现 | 产品确认部门可见性语义 | 匿名仅返回公开类型；部门型要求认证并按关系过滤；匿名/同部门/跨部门测试 | M |
| P0-06 | 收紧生产 CORS | `AllowOriginFunc` 对任意 Origin 返回 true，同时 `AllowCredentials=true`、headers=`*` | 确认 local/staging/production 前端域名 | 按环境 allowlist；拒绝未授权 Origin；认证读写跨站测试；预检和正式请求均覆盖 | S |

## 5. P1 待办

### 5.1 质量门禁

| ID | 待办 | 证据 | 验收标准 | 工作量 |
|---|---|---|---|---|
| P1-01 | 对齐 CI Go 版本 | go.mod=1.26；CI 和 PR gate=1.24 | CI 从 go.mod 读取或固定 1.26.x；输出 go version；test/build/vet 同提交通过 | S |
| P1-02 | 让前端与安全检查真正阻断 | npm ci、Vitest、coverage、ESLint、gosec 均允许失败继续 | 依赖/测试/覆盖率失败阻断；安全 P0/P1 阻断；报告保留但不吞退出码 | S-M |
| P1-03 | 建立 skip/race 门禁 | 环境型 skip 无 owner/到期日；race detector 未形成有效证据 | skip 台账含 owner/原因/最近运行/发布影响；CGO=1 执行 `go test -race ./...` | M |
| P1-04 | 真实 MySQL 并发与失败注入 | manual submit/expiry、reissue、报告生成/下载、模板替换等竞争仍未完全证明 | 双连接竞争；唯一完整 run；无半成品 current/孤儿文件；失败后重试与回滚一致 | L |

### 5.2 后端正确性与安全

| ID | 待办 | 证据 | 验收标准 | 工作量 |
|---|---|---|---|---|
| P1-05 | 统一路径边界校验 | candidate PDF 清理、上传、批量下载使用无目录边界的 `HasPrefix` | 封装 `filepath.Rel`/绝对路径校验；相邻前缀、`..`、盘符、符号链接测试；非法路径 fail-closed | M |
| P1-06 | 不再吞数据库错误 | Candidate Update/Remove/Logistic、多个 Paging/System CRUD 忽略 GORM Error/RowsAffected | 所有写操作检查 Error/RowsAffected；列表 DB 错误不得伪装空数据；文件+DB 失败顺序有补偿策略 | M-L |
| P1-07 | 所有分页统一上限 | OnlinePaging、UserExam、部门/角色接口只处理 size<=0，不限制超大正值 | 统一 `capPageSize()`；极大 pageNum 溢出测试；导出不走无限分页 | S-M |
| P1-08 | 消除列表与试卷创建 N+1 | Candidate/Tester 每行 COUNT；createPaperTx 每题查答案 | paper_id/qu_id 批量查询+GROUP BY/map；EXPLAIN 和查询次数测试 | M |
| P1-09 | JWT 默认值启动时 fail-closed | 默认配置为可预测占位符，加载逻辑未拒绝 production/staging 使用占位值 | 非 local 环境拒绝空值/占位值；密钥长度策略；启动门禁测试 | S |
| P1-10 | 文件写入原子性 | PDF copy/ZIP 的 Copy/Close 错误未完整传播 | 临时文件+fsync/close+rename；校验 magic/size；磁盘满和中断测试 | M |

### 5.3 前端业务恢复

| ID | 待办 | 证据 | 验收标准 | 工作量 |
|---|---|---|---|---|
| P1-11 | MBTI 网络失败恢复 | 加载、自动保存和提交缺少明确 catch、状态和重试 | 加载错误页；保存保留选择并可重试；提交未确认不得跳转；浏览器断网/超时验收 | M |
| P1-12 | 00401 答题失败恢复 | 初始 load 无 catch；保存失败清空选择；手工提交失败无提示 | 加载重试；保存失败不丢视觉选择；手工/超时提交均有受控状态和重试 | M |
| P1-13 | 通用 DataTable 错误恢复 | `getList()` 只有 then，失败后 loading 可永久保持 | catch+finally；错误提示/重试；删除和状态变更一致处理 | S |
| P1-14 | 固定业务回退与离开保护 | 多处 `$router.go(-1)` 依赖未知历史栈 | 无历史直达、token 失效、答题中后退均使用受控目标和确认 | M |

### 5.4 产品与数据验收

| ID | 待办 | 验收标准 | 工作量 |
|---|---|---|---|
| P1-15 | 005 fresh staging 生命周期 | 唯一 marker；draft→人员→freeze→140 UI 答题（含刷新）→submit→13维/4模块→报告→native download→失败注入→exact cleanup | L |
| P1-16 | MBTI 16 类型矩阵 | 16/16 类型计分、32/32 完整/简版转换、关键文本/字体/PDF/header/native download/cleanup | L |
| P1-17 | 四产品 native download | 页面真实点击、browser download event、saveAs、failure=null、SHA/MIME/Disposition/未认证拒绝 | M |
| P1-18 | 001/002/003 完整兼容回归 | 下一次公共代码发布前，新卷/答题/提交/结果/导出或受控拒绝；旧字节不变 | M-L |
| P1-19 | 409 孤儿 paper 闭包 | **2026-10-10 已关闭为用户批准保留例外**：active0，完整引用闭包与备份0/14恢复可能性已核，固定脱敏指纹门禁通过；新增或漂移重新打开P1 | 已完成 |
| P1-20 | 题库身份差异决策 | 明确 00301/00302 身份、repo 元数据和 005 bundle；禁止 staging 整包覆盖 production | M，需产品批准 |
| P1-21 | 磁盘/备份/uploadPath 治理 | 建索引和 DB 引用图；异机备份；MySQL 恢复演练；告警；盘占用目标 <70% | M-L，需清理批准 |
| P1-22 | 网络面与 SSH 收紧 | 8092/3306 仅 localhost/内网；UFW/NSG 双层最小化；root SSH 关闭且逃生链验证 | M，需运维批准 |

## 6. P2/P3 待办

### P2

1. 渐进拆分大 handler/service：优先拆路径校验、列表聚合查询、报告 I/O 和提交事务；不改变 API 契约。
2. 数据库连接增加 `ConnMaxIdleTime`、请求级 context/timeout、连接池和慢查询观测。
3. “删除报告”增加二次确认并使用危险按钮语义；先确认是删除还是停用。
4. 统一下载封装：状态码、MIME、magic、文件名、延迟 revoke、错误提示。
5. 报告生成后刷新结果状态；统一按钮权限显隐策略。
6. 00401 宽表移动端保留 5～7 个核心列，其余进入详情。
7. 历史账本增加 `current/historical/evidenceType/closedBy/remainingGaps`，避免旧 RED/PARTIAL 被误读。
8. 安全头补齐 CSP、Permissions-Policy；敏感路径显式 404/403。现有应用中间件已有 XCTO/XFO/Referrer，但 production Nginx/实际响应仍需按环境复验。
9. 当前 Go 低覆盖包建立目标：handler ≥45%、config/db/repository/redis 关键路径非零、pdfgen 失败路径有覆盖；只增不降。

### P3

1. 清理静默 `.catch(() => {})`、保留列表筛选/分页/排序返回状态。
2. 建立管理端口、CUPS、BT Panel、`/profile/` 的所有者与复核日期。
3. 统一 release manifest：source、binary、frontend、schema/content/template、service unit、Nginx、运行用户和回滚包 SHA。
4. 将 overall release 状态拆分为 runtime、security、data-governance、content-approval、rollback、release-approval 六个门。
5. 逐步清理 100 条工作树变更；不得直接删除未确认所有权的证据、备份或客户材料。

## 7. 推荐执行顺序

### 波次 A：立即安全门禁

1. P0-05 匿名部门考试 + P0-06 CORS（先 RED 测试，再修代码）。
2. P0-02 管理员凭据轮换。
3. P0-03 TLS。
4. P0-04 专用用户、秘密注入和 systemd hardening。

### 波次 B：质量门禁可信化

1. P1-01/P1-02 CI 修正。
2. P1-03 当前 skip/race 台账。
3. P1-05～P1-10 后端正确性与性能，按一个逻辑变更一批执行。
4. P1-11～P1-14 前端失败恢复。

### 波次 C：产品覆盖与数据闭环

1. P1-15 005 fresh lifecycle。
2. P1-16 MBTI 16 类型。
3. P1-17 native download。
4. P1-18 传统产品兼容回归。
5. P1-20～P1-22 数据、容量和网络治理；P1-19已按批准保留例外关闭。

### 波次 D：正式产品与长期治理

1. 产品批准后执行 P0-01 005 formal。
2. 渐进架构拆分、下载封装、账本结构化和发布可复现治理。

## 8. 当前发布判断

- **可继续运行**：当前已验证的 00401 核心链与 005 TEST 链。
- **不可宣称**：全项目 production-ready、formal-ready、全并发安全、全产品 E2E、全安全 GREEN。
- **不建议现在追加生产发布**：本轮没有待部署代码；应先选择一个逻辑变更，在本地 RED/GREEN、staging 验证后再单次申请 production 发布。
