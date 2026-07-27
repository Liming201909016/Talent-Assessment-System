# 39.106.61.48 生产发布准备报告（2026-07-27）

## 1. 发布目标

- 目标环境：`39.106.61.48`
- 目标入口：`http://39.106.61.48/`
- 发布提交：`d743d063f1281632f42a5630bffec08831533802`
- 题库编码：`00401`
- 初始化48个胜任力维度，允许客户维护。
- 不导入staging的384道AI测试题；客户通过00401页面导入正式题目。
- 开放完整胜任力测评链，并初始化带显著测试警示的`temp-v1`临时报告文案。

## 2. 当前生产只读基线

### 2.1 主机与服务

| 项目 | 已验证状态 |
|---|---|
| 主机 | `iZ0yosjdcen2p4Z`，Linux 5.4，x86_64 |
| 根分区 | 40G，总使用26G，剩余约12G |
| 内存 | 7.5GiB，可用约5.6GiB |
| Go服务 | `talent-assessment=active`，`/opt/talent-assessment/server`，root运行 |
| MySQL | active，5.7.44-log，端口3306 |
| Redis | 进程运行，监听127.0.0.1:6379 |
| Nginx | 宝塔Nginx进程运行；不是systemd管理 |
| Chrome/PDF | google-chrome、chromium、LibreOffice、pdfinfo、pdftotext均存在 |
| 中文字体 | Noto Sans CJK SC和文泉驿字体存在 |
| 应用后端 | 8092监听，直接`/health`返回`{"status":"ok"}` |

### 2.2 当前流量路由

当前生产存在两套前端/代理：

1. 旧入口 `:8088`
   - 静态目录：`/www/wwwroot/dist`
   - `/prod-api/`代理到`127.0.0.1:8091`
   - 当前8091无监听，因此API health/captcha为502。
2. Go候选入口 `:8090`
   - 静态目录：`/opt/talent-assessment/dist`
   - `/prod-api/`代理到`127.0.0.1:8092`
   - 浏览器登录页、health和captcha均正常。

目标无端口URL尚有阻塞：

- 当前客户端访问公网80端口失败。
- 服务器UFW已允许80，但云安全组/上游网络仍未放通。
- Nginx当前80端口仅有状态server，不是人才测评业务vhost。
- 正式切换前必须放通云侧TCP/80，并新增或调整业务vhost监听80、指向`/opt/talent-assessment/dist`和8092。

### 2.3 应用与目录

| 项目 | 当前状态 |
|---|---|
| 生产后端SHA-256 | `e437897b28c211d9e01b877c5eaf0b12f26b3cc4b7c4879eedfbde97e9061853` |
| `/opt`前端index SHA-256 | `30b6ed5bf6d7f48d92c4f2fe24ac3793c1142a406897874335b0fa8a734e6d2b` |
| 旧8088前端index SHA-256 | `96c0a9c012d0e47e008ed29d08a2ae5ef135631a739fc8ecebfb4c296389282d` |
| MBTI完整模板 | 32个文件 |
| MBTI简版模板 | 32个文件 |
| 导出模板 | 5个文件 |
| `/profile/`生产目录 | 尚未创建/映射 |
| 发布备份目录 | `/data/backup`和`/opt/talent-assessment/backups`当前均不存在 |

上线时必须先创建受限备份目录，并核对应用上传目录与Nginx `/profile/` alias的一致映射。

## 3. 数据库差异矩阵

### 3.1 生产数据规模

| 表 | 行数/估算 | 大小 |
|---|---:|---:|
| `el_exam` | 54 | 0.06 MiB |
| `el_qu` | 578 | 0.14 MiB |
| `el_paper` | 1,865 | 0.64 MiB |
| `el_paper_qu` | 158,104真实行；information_schema估算148,477 | 65.80 MiB |
| `sys_user` | 12 | 0.02 MiB |
| `sys_menu` | 91 | 0.02 MiB |

传统数据基线：

- `el_paper_qu_answer=349,119`
- `el_candidate=1,424`
- `el_tester=33`
- 已完成试卷1,441，进行中试卷514。
- MBTI答题表已存在，记录4,814。
- candidate/tester中已有PDF标记分别为1,168/9。

### 3.2 胜任力差异

| 对象 | 生产现状 | 发布动作 |
|---|---|---|
| `el_exam`六个分流字段 | 0/6 | 执行001 |
| `el_qu`六个胜任力字段 | 0/6 | 执行003 |
| `el_paper_qu`三个胜任力字段 | 0/3 | 执行004 |
| 八张胜任力表 | 0/8 | 执行002、004、006 |
| 胜任力外键 | 0 | 执行005、006 |
| 48维度 | 表不存在 | 002初始化48条 |
| 胜任力题目 | 字段不存在 | 发布后保持0，客户导入 |
| 胜任力测评/结果/报告 | 不存在 | 发布后保持0 |
| `temp-v1`文案 | 表不存在 | 006初始化392条 |

### 3.3 系统迁移

- 重复用户名组为0，可以执行用户名唯一索引迁移。
- `uk_sys_user_user_name`当前不存在。
- 23个退役菜单ID全部存在，名称/路径/权限与staging确认的在线用户、定时任务、缓存监控、代码生成及审计写按钮一致，可以执行退役迁移。
- 审计查询按钮不在退役集合内。

### 3.4 MBTI基础Schema

生产已具备：

- `el_mbti_answer`及4,814条数据。
- `el_repo=utf8mb4_general_ci`。
- `idx_qu_type_level_id`和`idx_repo_qu_type_qu_id`。

因此本次不执行整个`deploy-39.106.61.48-schema.sql`，避免重复执行其中的可选全表字符集转换；只执行胜任力和系统最小迁移集。

### 3.5 DDL风险

`el_paper_qu`约66MiB、158,104行，004迁移会增加三列和索引。预计适合低峰期维护窗口，但必须先在生产备份恢复的MySQL 5.7临时Schema演练并记录实际耗时。演练未完成前不得直接迁移生产主Schema。

## 4. 最小迁移顺序

1. `competency_001_schema.sql`
2. `competency_002_dimensions.sql`
3. `competency_003_questions.sql`
4. `competency_004_runtime.sql`
5. `competency_005_hardening.sql`
6. `competency_006_reports.sql`
7. `system_001_user_name_unique.sql`
8. `system_002_retire_unsupported_modules.sql`

迁移明确不包含：

- 384道AI测试题XLSX/JSON。
- staging数据库dump。
- staging测评、人员、答卷、结果或报告实例。
- 00401物理`el_repo`行；00401由代码动态生成。

## 5. 发布候选

本地门禁：

- Go全量测试通过。
- 前端Vitest：21个文件、127项全部通过。
- Linux后端构建通过。
- 前端production build通过，仅保留既有资源体积warning。

发布包位置（本地ignored临时目录）：

`Go-based Refactored System/tmp/prod-release-d743d06/`

| 文件 | SHA-256 |
|---|---|
| `server-linux` | `03397e0faf24a21fb6da4e76ba0776226ab87bf2a3c52f72b475eeed4791e44a` |
| `dist.tar.gz` | `aaff21f4995048a1a4e0b37b77d5d3a733af2b75759d72f826f0f25104bbd82a` |
| 解压后`index.html` | `f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136` |

包内包含8个最小迁移脚本和manifest；题目数据文件数为0。

## 6. 上线前阻塞门禁（执行前状态）

1. 云安全组放通公网TCP/80，使目标无端口URL可访问。
2. 确认业务vhost切换方案：监听80，静态目录`/opt/talent-assessment/dist`，API代理8092，并配置`/profile/`。
3. 取得生产临时Schema恢复/迁移演练授权。
4. 完成完整备份演练；当前生产没有现成发布备份目录。
5. 低峰维护窗口开始前再次确认当前进行中业务和514份历史进行中试卷状态。
6. 明确上线时旧8088入口的处理：保留只读跳转、切换到新Go服务或下线，不能继续代理无监听的8091。

## 7. 上线后无写入冒烟

- service/MySQL/Redis/Nginx及内外health。
- 登录页、captcha、管理员登录。
- 001样本：exam `1777519143742934858`，paper `1781537059361742297`。
- 002样本：exam `1779242722762934339`，paper `1779242771945149303`。
- 003样本：exam `1777340619080514997`，paper `1782622430279909945`。
- 三类测评详情、试卷详情、结果和小规模导出。
- MBTI专用详情返回48题。
- 00401唯一虚拟入口、题数0、48维度、模板下载有效。
- 空题状态下创建/发布受控拒绝。
- `temp-v1=392`，报告实例/审计为0。
- 不执行正式导入，不创建测试测评、人员、答卷或报告。

## 8. 客户导题约束

客户导题前必须另做数据库备份。推荐先每个维度少量试导，验证模板预览、题号唯一、维度内序号、正反向、启停和导出回读后，再分批导入正式全量题库。导题完成并签字验收后，方可创建和发布正式胜任力测评。

## 9. 实际执行结果

### 9.1 完整备份

生产备份目录：

`/opt/talent-assessment/backups/release_00401_20260727_153001`

| 备份 | 字节 | SHA-256 |
|---|---:|---|
| `element.sql.gz` | 14,274,653 | `283d227047aef5fefaa315f628c1da59ca13b67583f56806579a73528ebee08d` |
| `server.before` | 44,475,629 | `e437897b28c211d9e01b877c5eaf0b12f26b3cc4b7c4879eedfbde97e9061853` |
| `dist-opt.before.tar.gz` | 7,179,852 | `7a49be853c0ef4379c25c5b03229cfaa2e747fbe61d213592a4232ae8b4081fb` |
| `dist-8088.before.tar.gz` | 6,923,402 | `c96ffaaf7603f0ef7c880b0f4985bb6b8c6b4e2b2099d11212482fd6f045dfc5` |

同时备份了Nginx主配置、8090 vhost、systemd unit和23行菜单快照。备份文件模式为600，数据库gzip完整性检查通过。

### 9.2 MySQL 5.7迁移演练

经授权创建临时Schema `element_release_verify_20260727`，恢复生产备份后执行完整迁移两轮：

- 恢复耗时：13.812秒。
- 第一轮迁移：20.353秒，其中004运行时迁移14.887秒。
- 第二轮幂等重跑：0.172秒。
- 传统核心行数摘要迁移前后均为`54|578|1955|158104|349119|1424|33`。
- 48维度、8张表、15个外键、392条temp-v1、用户名唯一索引、23个退役菜单全部符合预期。
- 胜任力题目、测评、报告实例均为0。
- 演练结果：`PRODUCTION_MIGRATION_REHEARSAL_PASS`。
- 临时Schema已DROP并确认不存在。

### 9.3 生产主Schema迁移

在完整备份和演练通过后执行相同迁移集：

- 总耗时：16.109秒。
- 004运行时迁移：10.846秒。
- 传统核心行数摘要保持不变。
- 维度：`48|48唯一code|48唯一name|48唯一order`。
- 00401题目：0。
- 胜任力测评：0。
- temp-v1临时文案：392。
- 报告实例/审计：`0|0`。
- 外键：15。
- 退役菜单：23。
- 结果：`PRODUCTION_SCHEMA_MIGRATION_PASS`。

### 9.4 应用部署

按用户选择，先部署并验收8090，不切换公网80：

- 后端SHA-256：`03397e0faf24a21fb6da4e76ba0776226ab87bf2a3c52f72b475eeed4791e44a`。
- 前端index SHA-256：`f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136`。
- 旧dist保留于`/opt/talent-assessment/dist.pre00401.20260727_155834`。
- `/opt/talent-assessment/tmp/uploadPath/profile`已创建，8090 vhost增加`/profile/`静态alias。
- Nginx配置检查通过并reload。
- `talent-assessment=active`，8092 health、8090 health、8090 captcha和登录页均正常。

### 9.5 生产只读冒烟

- 001详情、试卷详情、结果、导出通过；导出9,730字节。
- 002详情、试卷详情、结果、导出通过；导出9,460字节。
- 003详情、试卷详情、结果、导出通过；导出18,788字节。
- MBTI专用详情返回48题。
- 48维度全部可查询且题数为0。
- 00401虚拟入口唯一、题数0、`virtual=true`。
- 00401导入模板为有效XLSX，6,644字节。
- online/job/cache/codegen代表端点均为404。
- `getRouters`不暴露退役菜单。
- 短时Redis管理员会话已清理。
- 最终数据库：非法模式0、孤儿0、胜任力题0、胜任力测评0、报告实例/审计0。
- 部署后服务journal无error或关键失败。

### 9.6 生产浏览器验收

- 管理后台首页在8090真实加载成功。
- 题库管理列表显示唯一`00401 胜任力测验题库`，题目数量为0，共7个题库入口。
- 00401专用页显示“维度维护、导出题目、下载模板、导入题目”，空题库明确显示“暂无胜任力题目”。
- 维度维护页显示总数48；D01起始记录、名称、VIRD、适用类别、核心含义、题数0和启用状态均正常。
- 浏览器验收仅执行只读导航，没有导入、编辑、启停或创建测评。

### 9.7 部署后观察

服务于15:58:34启动；16:13:40（超过15分钟）复核结果：

- `talent-assessment=active`。
- 8092和8090 health均返回`{"status":"ok"}`。
- 自部署启动以来journal关键错误数为0。
- 2026-07-27部署时间窗内Nginx新增5xx为0。
- 阶段性数据库终验仍为48个唯一维度、胜任力题目0、胜任力测评0、temp-v1文案392、报告实例/审计0、非法模式0。

生产`/tmp/prod-release-d743d06`目录和一个同名前缀归档仍存在。它们仅是上传发布包，不是运行文件；删除属于破坏性操作，等待单独确认。正式备份目录不在清理范围内。

### 9.8 尚未完成的入口切换

生产新版本当前入口为：

`http://39.106.61.48:8090/`

按用户选择暂不切换80。目标`http://39.106.61.48/`仍需：

1. 云安全组/上游网络放通TCP/80。
2. 配置业务vhost监听80并指向`/opt/talent-assessment/dist + 8092`。
3. 切换后重复health、captcha、登录页及传统只读冒烟。

旧8088入口未修改，仍代理无监听的8091，其API 502属于部署前既有状态。

## 10. 本次发布回滚点

### 应用回滚

若8090新版本出现应用级故障：

1. 停止`talent-assessment`。
2. 从备份目录恢复`server.before`到`/opt/talent-assessment/server`，权限恢复为可执行、root:root。
3. 将当前dist移出，使用`dist-opt.before.tar.gz`恢复`/opt/talent-assessment/dist`。
4. 从`talent-assessment.conf.before`恢复8090 vhost。
5. 执行宝塔Nginx二进制配置检查并reload。
6. 启动`talent-assessment`。
7. 复核后端SHA-256恢复为`e437897b28c211d9e01b877c5eaf0b12f26b3cc4b7c4879eedfbde97e9061853`，再执行传统只读冒烟。

旧8088未在本次切换，因此不需要为应用回滚修改`/www/wwwroot/dist`；其备份仍保留供独立恢复。

### 菜单回滚

若退役菜单需要恢复，必须根据备份目录中的`menu-targets.before.tsv`按`menu_id`逐行恢复原`status/visible`，不能按父级或名称模糊更新。

### 数据库回滚边界

本次迁移只新增表、列、索引、外键和维度/temp-v1种子，并更新23个已确认菜单。应用回滚时默认保留新增Schema，避免在有新数据后执行危险DROP。

只有满足以下条件时才允许整库恢复：

- 已确认主库业务数据被错误修改，单纯应用/菜单回滚无法恢复。
- 明确接受恢复点之后的生产写入全部丢失。
- 再次获得数据库恢复授权。
- 使用`element.sql.gz`恢复前先停止所有写入，并再次验证gzip与SHA-256。
