# Production 磁盘治理（2026-10-10）

## 1. 结果

production 根盘已从 **80%** 降至 **72%**，可用空间从 **8,116,183,040 bytes** 增至 **11,419,566,080 bytes**。

- 实际释放：**3,303,383,040 bytes**（约 3.08 GiB）。
- 未重启应用、MySQL 或主机。
- 未删除或修改业务备份、uploadPath、private、MySQL 数据、客户 ZIP、旧 Java 部署目录或应用发行文件。
- 后端 SHA、PID、重启次数保持不变；内外健康与公网入口均通过。

## 2. 用户批准范围

用户在只读盘点后明确批准“执行推荐清理”：

1. 清空 Snap 下载缓存；
2. 删除 4 个 disabled Snap revision；
3. 清理 APT 包缓存；
4. 将 systemd journal 归档压缩至 256 MB；
5. 不触碰业务备份、uploadPath、MySQL、home 和 root 下历史部署资料。

## 3. 写前盘点

### 3.1 文件系统

| 指标 | 写前 |
|---|---:|
| 根盘总量 | 41,892,380,672 bytes |
| 已用 | 31,839,608,832 bytes |
| 可用 | 8,116,183,040 bytes |
| 使用率 | 80% |
| inode 使用率 | 14% |
| 已删除但仍打开 | 9 个 / 18,816 bytes |

### 3.2 主要占用

| 路径 | 大小 |
|---|---:|
| `/root` | 8,841,104 KiB |
| `/var` | 5,855,252 KiB |
| `/usr` | 5,275,656 KiB |
| `/www` | 4,533,904 KiB |
| `/opt` | 3,883,420 KiB |
| `/home` | 1,962,740 KiB |
| `/opt/talent-assessment/backups` | 1,881,796 KiB |
| `/opt/talent-assessment/tmp/uploadPath` | 515,728 KiB |
| `/var/log/journal` | 1,166,102,528 bytes |

### 3.3 已识别但未清理的业务或高风险数据

- `/root/deploy6`：约 8.0 GiB，含旧部署、源码、Git pack、归档和日志；未确认所有权及恢复价值。
- `/home`：约 1.9 GiB，主要为客户历史 ZIP；未清理。
- `/www/server/data/mysql-bin.000006`：约 798 MiB；未在未核复制/恢复策略前清理。
- 应用 backups：写前约 1.8 GiB，含 production release、phase1、UF-056/057/058 等恢复点；未清理。
- uploadPath：写前 817 文件 / 526,241,879 bytes；绝大部分超过 90 天，但未经 DB 文件引用闭包，不按年龄删除。
- `/tmp`：有约 286,776,301 bytes 的 30 天以上常规文件，但混有 socket、Tomcat、历史 SQL/报告探针等，未做全局清理。

## 4. 已执行清理

### 4.1 Snap 缓存

- 写前物理占用：3,546,935,296 bytes。
- 删除范围严格限定为 `/var/lib/snapd/cache` 的一级普通文件。
- 写后缓存普通文件数：0。

### 4.2 Disabled Snap revisions

已通过 `snap remove <name> --revision=<rev>` 删除：

- `chromium:3537`
- `cups:1238`
- `gnome-46-2404:164`
- `mesa-2404:1165`

写后 disabled revision 数：0。当前 active revisions 保留并核对存在，包括 Chromium 3554、CUPS 1262、GNOME 168、Mesa 1839 和 Snapd 28254。

### 4.3 APT 缓存

- 写前包缓存：522,063,872 bytes。
- 执行 `apt-get clean`。
- 写后 payload 文件数：0；保留 0-byte `lock` 文件。

首次控制脚本在 APT 阶段后因把合法的 0-byte `lock` 文件误判为残留而退出。此前 Snap 和 APT 清理已经完成，服务保持健康。随后将断言纠正为“忽略 lock，只禁止 payload 文件”，没有重复删除业务文件。

### 4.4 systemd journal

- 写前：1,166,102,528 bytes。
- 执行 `journalctl --vacuum-size=256M`。
- 写后：192,954,368 bytes。

## 5. 写后验收

| 项目 | 结果 |
|---|---|
| 根盘 | 72% |
| 已用 | 28,536,225,792 bytes |
| 可用 | 11,419,566,080 bytes |
| 后端 SHA | `13d5f07e5d157db5a379c014b43c058094800fc01c7d5768ea1be419b7d43415` |
| Go service | active |
| MySQL | active |
| PID | 2382320（未变） |
| NRestarts | 0 |
| 内部 8092 health | 200 |
| 本机 Nginx root | 200 |
| 本机 8090 API | 200 |
| 公网 root | 200 |
| 公网 8090 health | 200 |
| 最近 10 分钟 panic/fatal/permission denied | 0 |
| Google Chrome | 148.0.7778.96 |
| LibreOffice | 7.4.7.2 |
| 已删除但仍打开 | 10 个 / 18,816 bytes，可忽略 |

最终业务目录只读复核：

- backups：46 个顶层对象 / 1,926,959,104 bytes；未删除。
- uploadPath：验收时 819 文件 / 527,645,639 bytes；任务期间存在并发业务生成文件。只读检查发现最近约 4 分钟有 3 个约 0.7 MB 文件，仅记录路径指纹，未读取或删除。该增长解释了写前/写后计数差异，不属于本次清理。

## 6. 当前结论与后续门禁

- 本次 80% 磁盘告警已关闭到 **72%**，但尚未达到此前建议的 `<70%` 长期目标。
- 继续降至 70% 以下不需要立即删除业务数据：可另行确认 journald 永久上限、Snap/apt 定期缓存策略；任何 `/root`、`/home`、MySQL binlog、backups 或 uploadPath 清理都必须先完成所有权、恢复价值或引用闭包。
- 推荐告警阈值：warning 75%、critical 85%；当前 72% 低于 warning。
- 未改变数据库、应用、前端、Nginx、systemd unit、网络或防火墙配置。
