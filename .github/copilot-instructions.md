# 工作区规则（全局生效）

## 一、三区结构（最高优先级）

本工作区包含三个区域，所有文件操作必须遵循归属规则：

### A区：Legacy Java System/【只读区】
- 原系统完整源代码（包含 Java 后端和 Vue 前端）
- 绝不创建、修改、删除此目录下任何文件
- 仅用于读取和分析

### B区：Go-based Refactored System/【纯代码区】
- 新 Go 后端项目，只存放可编译、可运行的代码和运行配置
- 不放文档、分析报告、测试脚本、SQL 脚本等非运行文件

### C区：工作区根目录下的辅助目录【文档、规则、脚本】
- .github/ — Copilot 规则文件
- docs/ — 所有文档、分析报告、API 契约、架构设计
- scripts/ — 脚本和数据（按子目录分类，见下方"脚本目录结构"）

### 文件归属判断
- 想放文档/报告？→ docs/
- 想放 .go 代码？→ Go-based Refactored System/ 下对应子目录
- 想放自动化测试脚本？→ scripts/test/
- 想放工作用脚本（分析、调试、数据生成、对比）？→ scripts/tools/
- 想放数据库运维脚本（同步、备份、Schema）？→ scripts/db/
- 想放 SQL 变更脚本？→ scripts/sql/
- 想放数据文件（模板、JSON、SQL 数据）？→ scripts/data/
- 想放规则/提示词模板？→ .github/
- 不确定放哪？→ 先问我，不要自行决定

## 二、基本原则

1. Legacy Java System/ 只读，任何情况下不修改
2. 一次只做一件事，完成后等待确认再继续
3. 规则是渐进式的：本文件会随项目推进逐步补充细节规则（如编码规范、技术栈、部署规则等），当前未明确的细节请先询问
4. 生成任何文件前，先确认它属于哪个区
5. 如果需要创建新的规则文件（如 .github/instructions/ 下的子规则），请先说明用途，经确认后再创建
6. **发布纪律**：不频繁发布到生产服务器（20.200.136.133）。所有变更先在本机完成开发和测试，积累到一定程度后，向用户确认再统一发布。禁止每改一处就立即部署。

## 三、运行环境规则

Go 项目开发和运行过程中产生的各类文件，按以下规则存放：

### 环境与构建
- Go 版本要求、环境变量说明 → Go-based Refactored System/README.md
- go.mod / go.sum → Go-based Refactored System/ 根目录
- 编译产物（二进制文件）→ Go-based Refactored System/bin/（不提交 Git）
- 本地开发用的环境变量文件 → Go-based Refactored System/.env.local（不提交 Git）

### 运行时产物
- 运行日志 → Go-based Refactored System/logs/（不提交 Git）
- 文件上传临时目录 → Go-based Refactored System/tmp/（不提交 Git）

### .gitignore 规则
Go-based Refactored System/.gitignore 中必须排除：bin/、logs/、tmp/、.env.local、.env.staging、.env.production、*.exe、*.log

### 部署环境
本项目存在三个运行环境，配置文件按环境区分：

| 环境 | 用途 | 配置文件 |
|------|------|----------|
| local | 本地开发调试 | Go-based Refactored System/.env.local |
| staging | 云端测试验证 | Go-based Refactored System/.env.staging |
| production | 正式生产环境 | Go-based Refactored System/.env.production |

- 云平台暂定 Azure，后续可能调整
- 所有 .env.* 文件不提交 Git（已在 .gitignore 中排除）
- 部署脚本和 IaC 配置（如 Dockerfile、docker-compose、Azure 资源模板等）→ Go-based Refactored System/deploy/
- 部署文档（架构图、部署步骤、环境说明）→ docs/

### 环境切换原则
- 代码中不出现任何环境相关的硬编码值（IP、端口、密钥、连接串等）
- 所有环境差异通过配置文件或环境变量注入
- 同一份代码、同一个二进制，通过不同配置文件切换环境

## 四、脚本目录结构（scripts/）

scripts/ 目录按用途分为 5 个子目录，严格按归属存放：

### 目录结构

```
scripts/
├── test/                  # 系统自动化测试套件
│   ├── package.json       # 测试依赖
│   ├── *.js               # 活跃的测试套件（仅保留正在使用的）
│   ├── fixtures/          # 测试固定数据（模板、截图基线、示例文件）
│   ├── screenshots/       # 测试截图输出（不提交 Git）
│   ├── results/           # 测试结果 JSON（不提交 Git）
│   └── archive/           # 已归档的历史测试脚本（不再运行，仅留档参考）
├── tools/                 # 工作用脚本（分析、调试、数据生成、对比）
│   ├── *.js / *.py        # 数据分析、调试、生成、清理、对比等一次性或辅助脚本
│   └── *.ps1              # PowerShell 对比/批处理脚本
├── db/                    # 数据库运维脚本（同步、备份、Schema 检查）
│   ├── db-sync*.sh        # 数据库同步
│   ├── restore-*.sh       # 数据恢复
│   ├── schema-check.sh    # Schema 检查
│   └── schema.sql         # Schema 定义
├── data/                  # 数据文件（SQL 数据、模板、JSON）
│   ├── *.sql              # 数据导入 SQL
│   ├── *.xlsx             # 模板文件
│   └── *.json             # 数据 JSON
└── sql/                   # SQL 变更脚本（DDL、DML）
    └── *.sql              # ALTER、CREATE、INSERT 等变更语句
```

### 归属判断规则

| 你要放的东西 | 放到 | 判断依据 |
|-------------|------|---------|
| 可 `node test/xxx.js` 运行的自动化测试 | test/ | 有断言、有通过/失败判定 |
| 分析数据、调试问题、生成数据的脚本 | tools/ | 运行一次就完成任务，不是测试 |
| 数据库同步、备份、恢复、Schema 操作 | db/ | 操作目标是数据库实例 |
| SQL INSERT/ALTER/CREATE 语句 | sql/ | 单纯的 SQL 语句文件 |
| 数据文件（xlsx/json/csv/批量 SQL 数据） | data/ | 被脚本或系统消费的数据 |
| 已经不用的旧版测试脚本 | test/archive/ | 曾经是测试，现在已被新版替代 |
| 测试产物（结果 JSON、截图） | test/results/ 或 test/screenshots/ | 运行测试自动生成的 |

### 禁止事项
- 不在 scripts/ 根目录直接放脚本文件（必须归入子目录）
- 不在 test/ 放非测试脚本（分析、调试、生成类脚本放 tools/）
- 不在 test/ 放 node_modules/（已在 .gitignore 排除）
- 不在 scripts/ 放部署配置（nginx 配置等放 Go-based Refactored System/deploy/）
- 不在 scripts/ 放日志文件（.log 文件应在 .gitignore 排除）

### .gitignore 补充
scripts/ 下应排除：
- scripts/test/node_modules/
- scripts/test/screenshots/
- scripts/test/results/
- scripts/logs/
- *.log

## 五、Go 编码规约（已验证的坑与规范）

以下规则来自实际开发中踩过的坑，必须遵守：

### 5.1 JSON / Struct

1. **禁止在 request struct 中嵌入 model struct**：Go `encoding/json` 在扁平化后若外层和嵌入层出现相同 JSON tag，两者都会被静默忽略。解决办法：使用独立的扁平 request struct，手动映射到 model。
2. **前端可能发送空字符串 `""`**：Java Jackson 会宽容地把 `""` 转成 `null`，Go 对 `int/bool` 类型会直接报错。使用 `interface{}` 或 `*int` 接收，再手动转换。

### 5.2 时间处理

3. **必须用 `time.ParseInLocation`**：`time.Parse()` 返回 UTC 时间，存入 DB 会偏移 8 小时。始终用 `time.ParseInLocation(layout, value, time.Local)`。
4. **前端时间格式不统一**：可能收到 `2026-04-21 12:00:00` 或 `2026-04-21T12:00:00.000Z`（RFC3339）。用多格式尝试解析。

### 5.3 GORM

5. **`Save()` 会覆盖所有字段**：包括 `create_time`。更新记录前必须先从 DB 读取原始 `create_time`，赋值后再 Save。或者用 `Updates()` 只更新指定字段。
6. **`CreateInBatches` 替代循环 INSERT**：大量记录用批量插入，避免 N 次网络往返。

### 5.4 文件处理

7. **ZIP 文件不要复制原始 FileHeader**：Go 的 `zip.Writer` 在复制 FileHeader 时会因 CRC 不匹配写入 Data Descriptor，LibreOffice 无法解析。创建新 `zip.FileHeader{Name: ..., Method: zip.Deflate}` 即可。
8. **PowerShell 默认 UTF-8 BOM**：`Set-Content` 会加 BOM，MySQL `source` 命令会报错。使用 `[System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))` 写无 BOM 的 UTF-8。
9. **excelize 模板不要用 SheetJS 预处理**：SheetJS (`xlsx` npm) 的 `aoa_to_sheet` 会重建 sheet 内部 XML 结构，导致 Go `excelize` 的 `SetCellValue` 写入后数据丢失。模板保留原始格式（含示例数据），由 Go 代码 `RemoveRow` 清理。
10. **模板文件用纯英文命名**：中文文件名在跨平台传输（scp/git）和 glob 匹配时可能出问题。用 `{repoCode}.{english-type}.xlsx` 格式。

### 5.5 SQL 脚本

9. **MySQL Safe Update Mode**：DELETE/UPDATE 必须包含主键列条件（`WHERE id = ...` 或 `WHERE id BETWEEN ...`），不能只用非索引列。
10. **幂等脚本模板**：`SET NAMES utf8mb4; BEGIN; DELETE（清理）; INSERT; UPDATE（统计）; COMMIT;`

### 5.6 安全

11. **文件路径校验**：上传/下载 API 必须校验路径在允许目录内，防止路径穿越。
12. **Content-Disposition 中文**：使用 RFC 5987 编码 `filename*=UTF-8''` + URL 编码。

### 5.7 部署

13. **前端 dist 权限**：`scp` 上传的文件属于 `liming` 用户，nginx 以 `www-data` 运行无法读取 → 部署后必须执行 `sudo chmod -R 755 dist && sudo chown -R root:root dist`。
14. **PowerShell `$HOST` 保留变量**：不要在脚本中赋值 `$HOST`，会报错。用 `$H` 或 `$SERVER` 代替。
15. **SSH 超时重试**：Azure VM SSH 可能间歇性超时，部署命令加 `-o ConnectTimeout=10` 并准备重试。
16. **部署顺序**：后端 → 模板/配置 → 前端 dist → 权限修复 → 验证。每步验证后再下一步。

## 六、项目记忆（AI 经验记录）

### 文件位置
docs/project-memory.md — AI 和开发者共同维护的项目经验记录。

### 核心机制：读 → 执行 → 验证 → 回写

每次任务执行遵循以下循环：

第一步【读取】：开始任何任务前，先完整读取 docs/project-memory.md，基于已有记录制定执行计划，跳过已验证的内容

第二步【执行】：基于记忆中的已知信息执行任务，减少不必要的重复验证

第三步【验证】：执行过程中如果发现记忆中的内容有误、过时、或不完整，立即标记

第四步【回写】：任务完成后，将以下内容更新到 docs/project-memory.md：
- 新增：本次新发现的事实、配置、结论
- 纠正：将发现有误的旧记录标记为已纠正，保留原记录并注明纠正原因和新结论
- 补充：对已有记录增加细节或补充说明
- 完成：将进行中的条目更新为已完成

### 纠正记录格式

当纠正旧记录时，不要删除原记录，而是追加纠正说明：

原记录内容...
[纠正 - 日期] 上述内容有误。实际情况：xxx。发现原因：xxx

### 优先级
- docs/project-memory.md 中的记录优先于 AI 自身的推测
- 如果记忆与实际执行结果冲突，以实际结果为准，并立即回写纠正
- 如果记忆中没有相关信息，正常执行后将结论写入记忆

### 应该记录的内容
- 环境配置已验证的值（Go版本、数据库连接、端口等）
- 云端部署已验证的值（Azure资源名、区域、SKU、连接串等）
- 技术方案已确认的结论（JWT算法、密钥来源等）
- 模块迁移完成状态和测试结果
- 踩过的坑和解决方案
- 部署流程中的关键步骤
- 各环境配置差异
- 已确认的废弃代码或无用接口

### 不应该记录的内容
- 临时调试信息
- 每次编译日志
- 可直接从代码读取的信息

## 七、测试盲区减少机制（持续生效）

为了让测试覆盖率随项目演进而**持续上升**而非衰减，AI 在以下时刻必须主动行动：

### 7.1 四个账本（必须维护）

| 文件 | 作用 | 谁更新 |
|------|------|--------|
| docs/business-branches.md | 业务条件分支矩阵（每个 handler 函数的所有 if/case 组合） | 触发点 1/2/3/5 |
| docs/regression-tests.md | 已修 bug 的回归测试 backlog 和索引 | 触发点 2 |
| docs/user-feedback-log.md | 用户反馈→根因→测试的闭环记录 | 触发点 3 |
| docs/coverage-history.md | 覆盖率历史快照（漂移检测基线） | 触发点 5 |

### 7.2 五个触发点（必须遵守）

| 触发点 | 何时 | 驱动文件 | 强制行为 |
|--------|------|---------|---------|
| 1. 新增功能 | 写新 handler/组件 | coverage-aware.instructions.md | 先列分支矩阵→写代码→写测试 |
| 2. 修 bug | 含"修复/fix/出错"关键词 | bug-driven-test.instructions.md | 先写 RED 测试→修代码→GREEN |
| 3. 用户反馈 | 含"用户反映/线上有问题" | feedback-triage.instructions.md | 5 问→登记→分支→测试 |
| 4. PR 提交 | git push / PR | .github/workflows/pr-quality-gate.yml | 自动评论质量摘要 |
| 5. 周期扫描 | 每周/手动 | /coverage-drift prompt | 漂移分析→Top 10 待办 |

### 7.3 优先级标记（在账本中使用）

| 标记 | 含义 | 处理优先级 |
|------|------|----------|
| 🔥 | 生产 bug 或已知会引发用户问题 | P0 - 立即 |
| ⚠️ | 当前实现行为可疑或不确定 | P1 - 需确认 |
| ❌ | 未覆盖但风险一般 | P2 - 排队 |
| ⚪ | 不需要测试（必须有理由） | — |
| ✅ | 已有测试 | — |

### 7.4 关键原则

1. **盲区可见 > 盲区清零** — 不需要全测，但必须知道盲区在哪
2. **修 bug 必先写 RED 测试** — 即使是 1 行小修复
3. **每次代码变动只多不少地偿还测试债务** — 不强求一次到位
4. **账本随项目增长** — 不删除历史，只追加和标注
5. **AI 主动提问** — 不等用户问"还有什么没测的"

### 7.5 与编码规约的协同

每条 §5.1-5.7 编码规约都对应一组业务分支：

- §5.1.2（空字符串）→ business-branches.md 必有"前端 '' 输入"分支
- §5.2.3（时间解析）→ 时间字段必有"多种格式"分支
- §5.3.5（Save 覆盖）→ Update 必有"create_time 保留"分支
- §5.6.11（路径校验）→ 文件操作必有"路径穿越"分支

