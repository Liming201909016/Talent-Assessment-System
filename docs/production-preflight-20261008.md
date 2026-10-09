# 生产部署前只读检查（2026-10-08）

## 判定：PARTIAL / 暂不放行部署

### 最新续检：02:53～02:57Z Posh-SSH已连接，现状核验完成，仍NO-GO

以下为本轮最新事实；原02:41 SSH阻断及未核验记录保留为历史，由本节限定纠正。用户仍仅批准生产只读现状，不部署/restart/迁移/创建备份/清理业务。

| 项目 | 本轮实测 |
|---|---|
| 连接 | Posh-SSH/root/严格主机信任检查成功；iZ0yosjdcen2p4Z/Linux x86_64；安全提示消费密码，不保存凭据 |
| 应用 | active/running，PID1195022，root:root，NRestarts0，激活时间7月27日15:58:34 CST |
| 根盘 | 总40910528 KiB、可用10501028 KiB、使用率74% |
| 内存 | 总7658 MiB、可用5698 MiB、swap0；起始负载0.03/0.02/0.00 |
| 进程/工具 | nginx6、mysqld1、redis-server1；mysql/mysqldump、LibreOffice、PDF及Chrome/Chromium命令存在，不等转换验收 |
| 内部健康 | 8092/health、8090/prod-api/health均statusok |
| 环境 | 实际进程APP_ENV=production；限定report/formal/reissue白名单未返回其他值，不穷尽环境变量 |
| 数据库 | 配置凭据真实读取MySQL5.7.44-log/element/54表；全局28权限含SELECT/CREATE/ALTER各1，未写权限探针 |

后端47965277字节、root:root0755，磁盘及实际/proc/PID/exe均SHA `03397e0faf24a21fb6da4e76ba0776226ab87bf2a3c52f72b475eeed4791e44a`；首页15799字节/root:root0755/SHA `f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136`。两者同7月生产记录，终验保持；完整前端资产未逐项核验。

**数据库差异及旧卷风险：**

- 66个测评=57 legacy+9 competency；1980答卷=state0 517/state1 0/state2 1463。
- 旧链进行中是state0，不能用新管理特质state1=0称无活动卷。补核517份state0均有deadline且全部过期，未到期0/最近24h新建0，创建时间DB原值2022-06-16～2026-09-03。
- state0关联现存测评：316 legacy、2 competency，均到期；LEFT JOIN独立确认199份找不到所属测评。原因未调查，不删除；新Worker首扫2份旧胜任力卷的影响需明确策略与兼容验证，当前不交卷、不改deadline。
- 仅8旧competency表，48维度/19结果/3报告；全schema外键20仅总数，不冒完整签名通过。四个exam版本列计数0、el_mng_前缀表0，未具备最新一期/v2及管理特质Schema，不能直接整包升级。
- 旧生产已有数据，不能套用staging009重置或staging批准SQL。迁移保留政策、MySQL5.7恢复副本演练、源码→产物范围及正式报告准入尚未关闭。

**备份、权限与终验：**

- 限定backup深度2仅找到7月27日数据库gzip：14274653字节、root:root0600；发布子目录0700。不是全机/外部备份穷尽，也不是当前发布恢复点。
- 数据库SHA `283d227047aef5fefaa315f628c1da59ca13b67583f56806579a73528ebee08d`/gzip校验0；原SHA256SUMS整清单0；两套旧dist归档gzip/tar各0。旧binary/config/unit/menu存在，但未新恢复演练、未证明旧包适合当前回滚，不下载secret归档。
- 3306/8092全地址监听、6379仅127.0.0.1；未观察443/8088/8091监听。全地址监听不证明云侧公网可达；TLS/限源未验。生产配置root:root0644、backup父0755仅记录，不擅自改权限。
- 02:56:45Z终验exit0、非空stderr0、PID/两SHA/health保持；30min journal指定critical分类0不等全业务/Nginx5xx通过。02:57:14Z备份校验transport0/stderr0。检查后关闭本轮自有SSH连接。
- 先默认mysql socket尝试exit1，后只在远端内存解析现有生产DSN→匿名管道→mysql，秘密不输出/不落盘；查询显式READ ONLY、10秒SELECT预算，仅聚合/metadata，不读取人员/答案/密码hash。只读COMMIT不是业务DML。
- 首多行send被扁平化导致本地here-string ParserError、remote0；改为单行后成功，原工具失败保留。

**放行前提：**确定发布scope和manifest/独立审阅→当前全库及运行资产/PDF备份和恢复演练→5.7迁移兼容/旧数据保留→517旧卷及199缺关联策略→正式功能验收和维护窗口。此次只读检查不授权这些写操作；未build/tests、未部署、未restart、DDL/DML及业务清理0。

- 用户确认生产目标为 `39.106.61.48`，并选择先只查现状、不确定发布范围。本次不部署、不重启、不迁移、不创建备份或业务数据。
- SSH默认密钥认证被拒绝，远端命令通道未建立；主机资源、后端版本、数据库结构/权限、活跃答卷、备份完整性和回滚点均未核验。
- 本次本机Git变更360项，未冻结候选范围或包，未运行构建/测试。管理特质正式能力和新补发API不能沿staging最小身份修复的批准直接生产发布。参见[项目记忆](project-memory.md)及[生产准入清单](management-traits-production-readiness-20261006.md)。

## 公网实测

正文复核时间为2026-10-08T02:41:28Z～02:41:29Z。有限GET、不跟随重定向、不登录、不携带认证信息。

| 地址 | 首次HTTP状态 | 正文复核/错误 | 结论 |
|---|---:|---|---|
| `http://39.106.61.48/` | 200 | 1326字节；title“没有找到站点” | 非业务首页 |
| `http://39.106.61.48/prod-api/health` | 404 | 首次146字节 | 无端口入口未命中应用健康接口 |
| `http://39.106.61.48:8090/` | 200 | 15799字节；title“人才综合素质评估系统 [Go]” | 静态业务入口可访问，未验登录及资源闭包 |
| `http://39.106.61.48:8090/prod-api/health` | 200 | 15字节；`{"status":"ok"}` | 该入口应用健康可用 |
| `https://39.106.61.48/prod-api/health` | 无响应 | curl exit28、5秒连接超时 | 本次HTTPS探测未通过，原因未归因 |

公开正文SHA-256：

- 无端口首页：`cdf9d8eee8c4fe967fac3aa9218a7227647ae7aaaa4221c688e1aab7a9180f69`。
- 8090首页：`f5cd615b7a8f968b4ffba6fef61953c067fb86519d1a75987128f797bbb93136`，与[7月实际发布记录](production-release-39.106.61.48-20260727.md)一致；仅证明首页字节，不证明后端或完整前端版本未变。
- 健康正文：`a29ee2b15c494311c52521766e44af56a3ad2248e7a8ab465e5206463c13d288`。

## SSH阻断与后续前提

- 本机OpenSSH可用、目标有known_hosts记录；按历史root用户和现有默认认证，仅尝试一次。参数为BatchMode/严格主机检查、ConnectTimeout10、ConnectionAttempts1。
- exit255，固定错误分类 `authentication_denied`，stdout为空。未接受新密钥、猜口令、轮试其他密钥或复用staging认证。
- 本次终端未发现已连接Posh-SSH会话、SSH配置文件、生产环境文件或相关进程环境键。限定历史连接引用扫描仅输出位置/认证类别，不输出秘密。
- 用户表示此前已提供连接信息；当前上下文没有可消费的生产认证引用。历史记录中明文凭据脚本已清理，不恢复旧秘密。
- 下一只需既有授权SSH的安全引用：用户与私钥绝对路径，或已连接Posh-SSH所在终端/会话编号，或受限凭据文件路径及键名。不要在聊天粘贴密码。

取得认证后继续只读核验主机身份、服务/监听/资源、binary/frontend SHA、非秘密环境门禁、MySQL与Schema签名、所有业务活跃/到期卷计数、备份及回滚兼容性。发布范围明确后再查候选包和迁移差异，不自动部署或创建备份。

远端命令、上传、替换、重启、DDL/DML、人员/答卷/PDF操作均为0。HTTP产生正常访问日志，不宣称远端文件系统绝对零变化。