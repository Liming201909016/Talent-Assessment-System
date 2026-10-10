# Production 部署只读审计（2026-10-09）

## 1. 范围与边界

- 目标：`39.106.61.48 / iZ0yosjdcen2p4Z`
- 基准：当前production发行标记及staging已核验基线。
- 本轮仅执行只读HTTP、浏览器、系统、文件权限、端口、日志和SQL查询。
- 未修改代码、配置、权限、数据库或防火墙；未重启任何服务；临时MySQL client由trap删除。
- 浏览器只验证匿名登录页加载；未使用管理员凭据，未执行业务写链。

## 2. 当前可用性结论

| 项目 | 实测结果 |
|---|---|
| Go服务 | `active`，PID=`2370974`，`NRestarts=0` |
| 后端/进程SHA | `f850575b1dfa6eac3f5b4533148baf715eabc8afa13ccd32c7655d0507a7d600`，磁盘与进程一致 |
| 前端index SHA | `593d4a20d890d73bf47f9a519a22fffbdb4b539afd5e1cc1dcc2fdebff73cfde`，393文件 |
| Nginx | 配置语法PASS；root、favicon、代理health均200 |
| MySQL / Redis | `mysqladmin ping=alive`；`redis-cli PING=PONG` |
| 浏览器 | 页面标题“人才综合素质评估系统”，登录页显示正常，无控制台错误 |
| 当前PID日志 | 自2026-10-09 22:09:12启动后ERROR/FATAL=0、panic=0、permission denied=0 |
| OOM | 最近24小时0 |

结论：当前没有持续性500或后端故障。Nginx日志尾部的8条5xx和842条permission记录包含UF-055修复前请求，不代表当前仍失败。

## 3. 发现清单

### P0 — 应优先处理

#### P0-1：生产登录使用明文HTTP

- Nginx监听80/8090，未监听443。
- 浏览器`secureContext=false`，登录页及账号密码提交路径使用HTTP。
- 响应缺少HSTS；HTTP环境本身也无法安全启用HSTS。
- 风险：同链路攻击者可窃取或篡改登录凭据和会话。
- 建议：配置受信TLS证书和443，HTTP统一301到HTTPS；完成后核验登录、下载、代理、WebSocket（如有）及Secure Cookie。

#### P0-2：应用配置文件包含秘密标记但权限为0644

- `configs`顶层5个文件均可被本机其他用户读取。
- `application.yml`检测到3个`password/secret/token/dsn`标记；`application-production.yml`检测到4个；未输出任何值。
- 风险：任一本机低权限账号可读取数据库连接或应用秘密。
- 建议：先列服务实际读取身份，再将秘密迁移到0600环境文件或受控Secret Store；至少立即把含秘密配置收紧到root或专用应用组可读。修改前必须验证服务重启和回滚。

#### P0-3：Go服务以root运行且无systemd沙箱

- production有效unit：`User=root`、`Group=root`。
- `NoNewPrivileges/PrivateTmp/ProtectSystem/ProtectHome/ProtectKernel*/RestrictSUIDSGID/MemoryDenyWriteExecute`均未启用。
- `systemd-analyze security`总体暴露评分=`9.6 UNSAFE`。
- 仓库服务定义原本指定`User=liming`、`Group=liming`，production有效unit与其漂移。
- 风险：应用漏洞可直接获得整机root权限。
- 建议：先盘点server需要写入的报告、tmp、private路径；创建专用低权限用户和最小写目录，再逐项启用systemd hardening。不可直接改User后重启，否则700目录会导致业务中断。

#### P0-4：生产管理员凭据仍需轮换

- 当前项目状态已有只读BCrypt证据：管理员仍匹配仓库历史默认候选。
- 本轮未打印、读取或重试该凭据。
- 风险：已知默认候选可导致后台接管。
- 建议：用户单独批准后执行一次性轮换，验证新登录、旧密码失效、会话撤销和恢复方案。

### P1 — 高优先级治理

#### P1-1：主机防火墙规则过宽

UFW对Anywhere及IPv6开放：

- 22、80、8090；
- 3306、10301、8088、9001；
- 39000–40000整段端口。

外部实测当前仅22、80、8090可达；3306、8092、10301、631、8088、9001和抽样39000/39500/40000被云侧网络拦截。但3306、8092在主机分别监听`*`，主机防火墙并未形成最小权限第二道防线。

建议：先确认宝塔、FTP/被动端口及8088/9001是否仍有业务所有者；随后收紧UFW和云NSG。数据库应只绑定内网/localhost，Go 8092应只绑定127.0.0.1并仅由Nginx代理。

#### P1-2：SSH允许root直接登录

- `sshd -T`：`PermitRootLogin yes`。
- 公网22可达。
- 建议：先验证具备sudo的普通运维账号和密钥逃生链，再关闭root直登和密码认证，保留受控应急流程。

#### P1-3：数据库存在409条孤儿试卷

- `el_paper`找不到父`el_exam`：409条。
- 状态分布：state0=199、state2=210。
- 关联`el_paper_qu`=20,566；关联`el_paper_qu_answer`=42,963。
- candidate和MBTI直接引用计数均为0。
- 风险：历史查询、统计或清理逻辑可能遇到缺失owner；这与此前worker反复记录`owner missing`一致。
- 建议：先输出409条的创建时间、产品归属、报告/用户引用和备份可恢复性；得到数据删除或恢复授权后再处理，禁止直接删除。

#### P1-4：磁盘和备份增长

- 根盘40GB，已用30GB，使用率80%，可用约7.8GB。
- backups=38项/1.7GB；tmp/uploadPath=502MB/782文件。
- inode仅14%，内存可用约5.5GB；无swap。
- 建议：建立备份目录索引、保留周期和恢复价值分级；先识别失败尝试备份与正式恢复点，逐项批准后清理。上传目录须先建立DB引用图，不可按文件年龄直接删除。

### P2 — 中期加固

#### P2-1：HTTP安全头缺失

根响应只有基本`Server`和`Content-Type`，未观察到CSP、`X-Content-Type-Options`、`X-Frame-Options`、`Referrer-Policy`、`Permissions-Policy`。建议在TLS上线后按前端实际资源逐步启用，先report-only验证CSP。

#### P2-2：敏感路径被SPA fallback返回200

`/.env`、配置路径、server、backups、`.git/config`和swagger路径均返回与index相同的16,155-byte SPA页面，没有直接文件泄漏证据，但会干扰安全扫描和监控。建议为点文件、配置、备份和二进制路径增加显式404/deny规则。

#### P2-3：备份与上传文件权限依赖父目录阻断

- backups根目录0755，38个直接子目录均0700；其内有81个0644文件，但当前因子目录0700无法被普通用户遍历。
- private目录0700，内部无world-readable文件。
- uploadPath目录0700，782个文件为world-readable mode，但同样依赖父目录0700保护。
- 建议：收紧backups根目录为0700，并把备份/上传敏感文件统一0600；同时避免再次使用会保留父目录mode的`cp -a ... /`恢复方式。

#### P2-4：CUPS及管理面监听范围可缩小

- CUPS监听0.0.0.0/[::]:631，BT Panel监听0.0.0.0:10301；外网目前不可达。
- 建议：若无业务需要，停用CUPS；管理面只绑定管理网段或localhost+隧道。

## 4. 数据与功能一致性

- `foreign_key_checks=1`。
- active paper=0。
- 00401当前10维/90题、结果0、报告0，符合历史清理后的状态。
- 管理特质14张sidecar表、2个005 repo、3个definition bundle、2个exam profile、2个paper snapshot、2个result run。
- 00501/00502当前均`state=0`，与客户此前操作一致。
- 全题库与staging的00301/00302身份、repo元数据及definition bundle差异仍是独立NO-GO事项，本轮未改。

## 5. 建议执行顺序

1. TLS/HTTPS与管理员密码轮换。
2. 收紧配置秘密权限；设计专用应用用户并在staging演练systemd降权。
3. 收紧UFW/云NSG、MySQL及8092绑定地址；关闭无用端口/服务。
4. 对409条孤儿试卷做引用闭包和恢复/删除决策。
5. 制定备份和uploadPath保留策略，把磁盘水位降至可控范围。
6. 补安全头和敏感路径显式deny。

所有P0/P1修改均应遵循：staging演练 → production零写preflight → 单次备份/变更 → 公网与业务验收 → 明确回滚收据。
