# Production 孤儿试卷保留例外治理（2026-10-10）

## 1. 未执行与边界

- **未删除、更新、插入任何 production 数据。**
- 未重启服务，未修改 production 文件、配置、权限、网络或防火墙。
- 未输出原始 exam/paper/user ID、姓名、电话、密码、token、DSN 或其他 PII/秘密。
- 本轮仅执行 MySQL 5.7 `SET SESSION TRANSACTION READ ONLY` 查询和备份流式只读扫描。
- 用户明确选择：**保留并登记例外**，不采用删除或构造父 exam。

## 2. 环境与范围

| 项目 | production | staging/local |
|---|---|---|
| 目标 | `39.106.61.48 / iZ0yosjdcen2p4Z` | 本轮未查询 |
| 数据库 | MySQL `5.7.44-log` / `element` | 不适用 |
| 目标对象 | `el_paper` 中父 `el_exam` 不存在的历史记录及其引用闭包 | 不适用 |
| 写入 | 0 | 0 |
| 服务操作 | 0 | 0 |

## 3. Production 实测闭包

### 3.1 主链

| 指标 | 实测 |
|---|---:|
| 孤儿 paper | 409 |
| 缺失父 exam | 14 |
| state=0 | 199 |
| state=2（已完成） | 210 |
| state=1（活动） | 0 |
| paper_qu | 20,566 |
| answered paper_qu | 1,467 |
| paper_qu_answer | 42,963 |
| 仍可解析到现有源题的 paper_qu | 18,990 |
| 绑定现存 sys_user 的 paper | 409 |
| candidate/tester/MBTI 引用 | 0 / 0 / 0 |
| 创建时间范围 | 2022-06-16 ～ 2025-10-10 |

14 个父 exam 分组的脱敏确定性指纹为：

`26836df28929fb71c2fa4b901a8cc2dc0a1dfd14eb61d1b4fcfad27af1a7e21b`

该指纹由排序后的 `SHA256(exam_id):paper_count:state0:state2` 行以 LF 连接后计算，不保存原始 exam ID。

### 3.2 产品线索

| repo 线索 | paper_qu | 涉及 paper |
|---|---:|---:|
| 00101 | 16,200 | 180 |
| 00102 | 810 | 9 |
| 无法由当前源题关系解析 | 3,558 | 220 |

结论：这些记录主要是传统测评历史，不属于当前 00401、MBTI 或 005 sidecar 运行链。

### 3.3 旁链

所有含 `paper_id` 的当前表已从 `information_schema` 动态枚举。除以下两张快照表外，其他 paper 引用全部为 0：

- `el_paper_qu`：20,566 行 / 409 paper；
- `el_paper_qu_answer`：42,963 行 / 409 paper。

所有含 `exam_id` 的当前表也已动态枚举。父 exam 缺失的非零旁链为：

| 表 | 行数 | 缺失 exam 数 |
|---|---:|---:|
| `el_exam_repo` | 24 | 24 |
| `el_user_exam` | 19 | 12 |
| `el_user_book` | 43 | 3 |

其中 14 个 paper 父 exam 都有一条 `el_exam_repo` 关系；另有 10 个没有 paper 的缺失 exam 关系。`el_user_book` 中还有两个不属于这 14 个 paper 父 exam 的缺失 exam 分组。因此，旧的“只删 paper 及题快照”方案不能形成引用闭包。

### 3.4 外键与恢复来源

- production `el_paper` 当前没有指向 `el_exam` 的数据库外键，因此历史父记录删除后仍可留下 paper。
- 最早可用完整 production 全库备份：`release_00401_20260727_153001/element.sql.gz`，14,274,653 bytes。
- 对该备份流式扫描 `el_exam` INSERT：14 个缺失父 exam 命中 **0/14**。
- 结论：当前可用备份不能恢复原父 exam。构造 tombstone 只能猜测缺失字段和业务语义，未获批准且不采用。

## 4. 风险判断

### 保留风险

- 全部 paper 均非活动状态，当前 modern competency/management-traits Worker 使用现存 exam JOIN，不会把这 409 条当活动新链处理。
- 传统历史查询若假定父 exam 必然存在，仍可能显示缺失标题/owner；需要以例外监控防止数量继续增长。
- 当前旁链还存在 24 条 `el_exam_repo`、19 条 `el_user_exam`、43 条 `el_user_book` 父 exam 缺失，不能把“409 paper 保留”误写成“所有 exam 引用完整”。

### 删除风险

- 会删除 210 份已完成试卷、1,467 条已作答题快照以及现存系统用户的历史记录。
- 当前最早备份也不能恢复父 exam，删除后只能恢复孤儿快照，不能恢复完整业务语义。
- 逻辑删除约 63,938 条 paper 子记录不会自动缩小 InnoDB 表空间；预计只占相关大表约 12%～13% 的行量。若不执行高风险表重建/`OPTIMIZE TABLE`，对根盘 80% 基本没有即时收益。因此孤儿治理与磁盘治理必须分开。

## 5. 已批准的保留例外

用户于 2026-10-10 选择“保留并登记例外”。当前决策如下：

1. 409 条 paper、20,566 条 paper_qu、42,963 条 paper_qu_answer保留。
2. 与缺失 exam 相关的 `el_exam_repo`、`el_user_exam`、`el_user_book` 暂不删除或改写。
3. 不构造虚假的父 exam，不根据 paper.title 猜测 exam 配置。
4. 活动孤儿必须持续为 0；任何新增孤儿或基线漂移均重新打开 P1。
5. 该例外不代表 Schema 健康，只代表经用户批准保留历史不完整数据。
6. 磁盘 80% 不通过删除该批数据处理，另开独立治理任务。

## 6. 防回归门禁

保留例外的只读门禁应至少断言：

- orphan paper=`409`；缺失父 exam=`14`；活动 orphan=`0`；
- state0/state2=`199/210`；
- paper_qu/paper_qu_answer=`20566/42963`；
- candidate/tester/MBTI 引用均为 0；
- 409 条均仍绑定现存 sys_user；
- 14 组指纹保持 `26836df2...e21b`；
- 新增任意 orphan、活动 orphan、报告/新产品 sidecar 引用或指纹漂移必须失败并报警。

本轮新增只读工具：

- `scripts/tools/production-paper-orphan-readonly-preflight-20261010.sh`；
- `scripts/tools/production-paper-orphan-backup-coverage-readonly-20261010.sh`。

旧全局清理脚本 `scripts/sql/cleanup_orphans.sql` 已改为只读诊断，并明确禁止用于删除。

## 7. 验收收据

- production 主闭包预检：`status=passed`、`databaseWrites=0`、`cleanup=completed`。
- 最早全库备份扫描：`parentsFound=0/14`、`databaseWrites=0`、`cleanup=completed`。
- 临时客户端配置最终检查：`CLEANUP_EXIT=0`。
- production SSH 会话在任务收口时关闭。

## 8. 后续独立任务

下一项建议处理**磁盘 80%**：只读建立 backups/uploadPath/private/logs/MySQL 表空间索引和 DB 文件引用图，输出可释放空间与恢复价值；未经再次批准不删除任何文件或备份。
