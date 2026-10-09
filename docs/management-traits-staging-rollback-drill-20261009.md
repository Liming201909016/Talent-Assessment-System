# 管理特质 staging 发布包分离与完整回滚演练（2026-10-09）

## Reviewer schema signature addendum（attempt 3，PASS）

Reviewer指出原证据只证明三张新增表行数为0，没有证明rollback前、旧runtime运行时、恢复新runtime后三阶段的表结构完全相同。脚本现对`el_mng_exam_draft`、`el_mng_report_reissue`、`el_mng_reissue_audit`生成确定性canonical receipt：表engine/collation；列name/type/null/default/extra/order；索引name/unique/order/columns；外键name/ordered columns/referenced table+columns/update/delete。文本字段以HEX表示，NULL default有独立标记，避免分隔符和NULL/字符串歧义。

新回归先RED确认原脚本缺全部canonical字段和PRE/OLD/POST三阶段采集；补丁后Node合同与`bash -n`均exit0。执行前只读preflight为active paper=0、三表行数=0。用户授权的完整staging rollback重新执行，保留前两次演练目录，使用新attempt 3目录；PID顺序`14293 → 15461 → 15554`，正常stop/start恰2次。

三份6328-byte canonical receipt逐字节相同，均为SHA-256 `1a9f16e82a3facd55d418d87bf85771c4d6d23eb5edc88c531b8d96c30ddbe27`，且与003/004迁移安装后的预期schema常量匹配。每份包含3条table、29条column、20条index-column和7条FK-column记录。机器结论见[attempt 3 verdict](../scripts/test/results/mng-release-package-separation-20261009/rollback-drill-attempt3-schema-verdict.json)，完整运行输出见[attempt 3 output](../scripts/test/results/mng-release-package-separation-20261009/rollback-drill-attempt3-schema-output.txt)，三阶段原始receipt分别为[PRE](../scripts/test/results/mng-release-package-separation-20261009/full-rollback-drill-20261009-attempt3-schema/pre.additive-schema.canonical)、[OLD](../scripts/test/results/mng-release-package-separation-20261009/full-rollback-drill-20261009-attempt3-schema/old.additive-schema.canonical)、[POST](../scripts/test/results/mng-release-package-separation-20261009/full-rollback-drill-20261009-attempt3-schema/post.additive-schema.canonical)。

旧恢复演练receipt SHA=`1ebadfd6818723bc83465ed4c85d43ff0a9da6e5f397e1e53537ef082ae2ec3c`只含两张reissue表，且缺表engine/collation及列default/extra，故明确标记`HISTORICAL_REHEARSAL_EXACT_COMPARABLE=0`，不做后验伪证明。该缺口由本次安全完整rollback的PRE/OLD/POST实时采集关闭。

最终独立只读核验见[attempt 3 final](../scripts/test/results/mng-release-package-separation-20261009/post-rollback-attempt3-schema-final.txt)：PID15554、`NRestarts=0`、三服务active、内外health200；当前server/index/393-file manifest恢复，配置/受保护DB/PDF SHA及冻结002分数`58.642639`保持，state1/两个overdue/三表行数/临时认证残留/应用错误/Nginx5xx均0。

## 结论

**PASS（仅 staging）。** 本次只操作 `20.200.136.133`；production 未访问。当前 staging 已恢复并保持新发行：后端 SHA-256 `d30c8e40e0c99dc525ad42b1269581665dd72acb7e1bfd47958ac0efd1561fc7`，首页 SHA-256 `c4f3b8f76b740244bd6b1d10ac6e24e4650f15367f1c427e4b46055297e21555`，前端 393 文件，三服务 active，内外 health 200。

完整机器结论见[回滚演练 verdict](../scripts/test/results/mng-release-package-separation-20261009/rollback-drill-verdict.json)，SHA-256 `d00015666825b59eda5f2a47188ddba510071a82c0cd0f99f4ef710d68b9702c`；独立最终状态见[post-rollback-final.txt](../scripts/test/results/mng-release-package-separation-20261009/post-rollback-final.txt)，SHA-256 `6bd222f67da4809549047edb477412d6a8ddf6f786a0415f79d1c7bd0f652da3`。

## 1. 运行包与验收包分离

原统一工作目录未删除或改写；新证据目录为[scripts/test/results/mng-release-package-separation-20261009](../scripts/test/results/mng-release-package-separation-20261009)。

| 包 | 内容边界 | 文件数 | 字节 | SHA-256 |
|---|---|---:|---:|---|
| runtime-package.tar.gz | 单一 Linux server、展开后的 393 个 dist 文件、dist SHA 清单、2 个运行资产、003/004 两个已安装迁移 | 400 | 23,036,161 | `e077c4003d14b441048b19e78d3b56eac5775316f8ffff864fe5b376efb4c450` |
| acceptance-package.tar.gz | `reissue-service.test`、2 个 SQL fixture、2 个资产 fixture、overlay 元数据及4个验收编排脚本 | 11 | 11,126,700 | `4a5122dfe13fa0ee05e0fdacc14060a8ab472a619186fbac170da483f610e3a9` |

[package-manifest.json](../scripts/test/results/mng-release-package-separation-20261009/package-manifest.json) SHA-256 为 `5383d3bb49f39b45b529ef1fcfee1eb1b96b8da8bba92b5721d6d3008c2fc617`。两个包内的 `SHA256SUMS` 均实际复核通过。

扫描覆盖 runtime 400 个实际文件及其归档、acceptance 11 个实际文件及其归档；同时以 Latin-1 与 UTF-16LE 视图扫描二进制字符串、配置、SQL、脚本和前端资源。两个包高置信秘密命中均为 0。宽泛词扫描只命中 Go 二进制中的 `password/secret` 符号及前端 JSEncrypt 生成 PEM 的代码标记；后者不含 Base64 PEM key body，不是嵌入私钥。验收二进制允许的 synthetic 测试字面量未形成凭据/DSN/JWT/私钥高置信命中。

## 2. 演练前门禁

- active paper=`0`；到期 competency worker=`0`；到期 management-traits worker=`0`；8092 已建立连接=`0`。
- draft/reissue 三张新增表总行数=`0`。
- 受限备份 `/opt/talent-assessment/backups/mng_current_d0e8202eafb14b08` 为 `0700`，既有四项 `SHA256SUMS` 全部通过。
- 备份中的旧后端实际 SHA=`4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c`，旧首页实际 SHA=`52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860`，不是仅按预期值推断。
- 旧前端 393 文件，完整路径/字节 manifest SHA=`8633c5805ff42fde40831256606ea939650416adbeccd81ce0127c54e6aec55d`。
- 既有发布 payload 目录=`0`；运行路径下 test binary=`0`。

## 3. 完整 rollback → reapply

成功轮 PID 顺序为 `13961 → 14209 → 14293`，正常 service stop/start 恰好 2 次：

1. 原子切换到备份旧后端/旧前端，启动 PID `14209`。
2. 核验磁盘和 `/proc/<pid>/exe` 均为旧 server SHA，旧 index/full manifest 匹配；health 200；三个服务 active；双报告环境仍为 staging；配置 SHA 未变。
3. 旧 legacy detail HTTP 200、旧 management-traits results 匿名边界 401；两分钟、仅 Redis 的临时管理员会话验证 reissue qualification 为 404，证明旧二进制未注册新路由。临时会话和 header 文件在 finally 清零，未登录浏览器、未写数据库。
4. 原子切回演练前保存的精确新后端/新前端，启动 PID `14293`。
5. 新 legacy detail HTTP 200、results 匿名边界 401；新 reissue 匿名边界 401、临时管理员请求 409（业务输入无效但路由存在），不是 404。
6. 新后端磁盘与进程 SHA、393 前端文件及完整 manifest 全部恢复；最近应用严重错误与 Nginx 5xx 均为 0。

第一次演练在旧阶段因匿名请求先被全局 JWT 中间件返回401，而脚本错误期望路由404，实际失败并立即恢复新发行；该轮额外执行1次恢复启动，证据保留且不计入随后成功轮的2次重启。第二轮改用短时内部认证准确验证旧404/新409，未削弱断言。

## 4. 数据与环境不变量

rollback 前、旧版本运行时、新版本恢复后均保持：

- configs + systemd unit/drop-in SHA=`ed9837af63bce10ac4005a6e56fafc4565ff267e7a988078e292b63e7da4d53f`。
- 排除三张新增空表后的管理特质受保护数据 dump SHA=`5e80140dbc2308c60f38b2f485941f91a83b113fea62f20a6faaed974b9a8645`。
- 管理特质私有 PDF manifest SHA=`3d209586a67cce13c4a93ed34c43ceb9f8a0b11d64aa8549131d726b5bd75066`。
- 冻结002 completed overall score=`58.642639`。
- draft/reissue 三表总行数=`0`；未重跑迁移、未DROP表、未写业务数据。

最终 runtime 路径 test binary=`0`；历史受限 backup 内有5个验收二进制，全部文件 `0600` 且对应顶层备份目录 `0700`，模式错误=`0`。`/tmp/mng_current_*`、本轮上传脚本、短时Redis登录键和 `/run` header/login 文件均清零。两个演练证据目录留在原受限备份下，顶层均为 `0700`。

## 5. 范围限定

本轮没有重新部署发行包、没有执行迁移、没有生成或查看报告、没有浏览器认证、没有生产访问。结论只证明 staging 对当前旧→新代码/前端回滚及精确恢复可行，并关闭发布包混入验收二进制的证据缺口；不构成 production 发布批准。
