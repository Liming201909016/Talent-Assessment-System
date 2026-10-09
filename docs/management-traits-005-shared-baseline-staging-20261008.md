# 管理特质005共享staging保留基线执行记录 — 2026-10-08

## 2026-10-08T15:43Z 完成：共享TEST基线已保留

**COMPLETED_STAGING_RETAINED_TEST_BASELINE。** 仅`20.200.136.133` staging共享`element`已保留00501/00502完整TEST基线；production未访问，在线应用未替换、未部署、未重启，管理UI仍未发布005入口。最终在线`MainPID=2746`、`NRestarts=0`、backend SHA=`4179fd3f...`、front SHA=`52eecf04...`、health=`ok`，active paper仍为0。

- 首次续跑前exact inspect证明前一wrapper exit1后upload/runroot/reportroot均不存在，005 repo/exam/prefix行均0，业务残留0。检查SQL误用`el_mng_paper_question_snapshot.exam_id`产生一次1054，只读命令停止且无写；纠正为经paper snapshot关联。
- 原binary 52,260,546 bytes/SHA=`7bbc8ec2...`不变；其源码`gofmt -d`大差异经原生字节核验仅CRLF，规范化后与gofmt输出完全相等。rollback脚本Git Bash `bash -n`为0。
- 第一次实际harness以exit1安全结束，固定分类仅`database connection`；原因是wrapper遗漏在线服务使用的`APP_ENV=production`配置overlay，而报告门禁仍为`REPORT_EFFECTIVE_ENV=staging`。defer清理后repo/exam/paper/run/report均0；私有stdout/stderr及hash保存在受限backup，不下载、不输出原错误。未改产品或放宽断言。
- 第二次使用同binary、增加`APP_ENV=production`后exit0，stdout安全分类仅PASS；00501/00502各repo1/关系140/题140/选项700，exam/profile/candidate/paper/snapshot各1，legacy answer140、bucket700、raw3 snapshot140、completed run1、13维、4模块、receipt1、TEST revision/current/generate audit各1，总体均`50.000000`。
- 两个DataSnapshot按MySQL原始UTF-8字节独立核验：00501=`26334021...`、00502=`10bf4c3a...`，数据库`data_sha`与`SHA2(CAST(data_snapshot AS BINARY),256)`精确相同。报告总数从`1/1/1`增至`3/3/3`；旧报告仍1且文件643663 bytes/SHA=`5bdaebeb...`与数据库一致。
- 私有服务器PDF均root:root 0600、无symlink并与DB/file一致：00501 647985 bytes/SHA=`d7fd78bb...`，00502 648159 bytes/SHA=`b392cff6...`。`ownership-complete.private.json`保留；原始及升级safe receipt均复制到root0700 backup。
- safe receipt v2 SHA=`e490f838...`；每产品包含exam/run/report/title/PDF/DataSnapshot/三身份字段hash及最终`identityCommitment`，无原始run/report ID或人员明文。升级器是单独opt-in只读测试，远端前后repo/exam/run/report计数`2/2/2/2`不变；本地test/vet/Linux编译和最终Go Build均0。
- 当前hardened oracle复用canonical Go AST与PDF object/递归凭据扫描：140题/100正/40反、13维、4模块，raw3独立维度/模块/总体均50；两PDF各A4九页、36客户段、13维名、5环+13柱+1常模线、4星、TEST警示、外链0、附件0、凭据发现0、identity run hash与commitment均存在。完整安全判定见`scripts/test/results/mng005-shared-baseline-c51130cb775ba019/pdf-oracle.json`。
- 002来源stream SHA在写前后保持：00201=`159b225f...`、00202=`a071422a...`；当前005 nonowned conflict0。临时upload/schema/user均0，最终精确删除server runroot；受限backup及私有reportroot永久保留。
- 首次独立终验在普通用户`cd` root0700 backup时权限拒绝，已保留为工具失败；随后仅用`sudo bash -c`复核manifest成功。下一次尾部因UTF-8固定标签比较未打印具体断点，已由此前已输出的REPORT_META精确标签证明替代为`mode=test`计数，最终尾部exit0。上述失败均不改业务事实。

## 结论

**BLOCKED_NETWORK_BEFORE_UPLOAD。** 本轮没有把00501/00502写入共享业务库，没有上传或执行测试二进制，没有创建报告PDF，没有部署、替换或重启在线应用。阻断发生在本地向`20.200.136.133`上传一次性测试二进制之前：首次与唯一重试均为TCP/22连接超时，远端命令未启动。

已安全完成的前置事项：fresh只读元数据核验、受限备份、专用一锤子测试二进制的本地编译/vet、以及未执行的精确回滚脚本。网络恢复后可从fresh只读门禁重新开始；不得依据本记录宣称005在线UI、共享库基线或客户模板报告已经可用。

## fresh staging事实

- 目标主机在只读连接时核为`vm-ubuntu-go-dev`，数据库为共享业务库`element`，MySQL 8.0.46。
- 在线应用保持`MainPID=2746`、`NRestarts=0`、active、health=`ok`；后端SHA-256=`4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c`，前端index SHA-256=`52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860`。
- 活跃试卷0、到期但仍active试卷0；因此没有因插入TEST基线而碰撞当前进行中测评的风险，但本轮仍未执行写入。
- 00501/00502 repo数量0，预留repo ID占用0；未复用未知记录。
- 00201与00202各为repo1/关系140/题140/选项700；只读摘要分别为`b7998aa68f6b8234743dd7b0fda9dddcbd96e7817bae10d110e7e2ce5c4e5c9a`与`29fbe6d538b51ff8670a88f94537cea6ada48257c8ea57b9d428699366be9fc4`。
- 已有管理特质11表，现有report revision/current/audit=`1/1/1`；reissue表0、draft/formal表0。方案不安装reissue、draft或formal新表，只使用现有report revision/current/audit。
- 冻结002基线仍为completed1、总体`58.642639`、manifest SHA=`66277c8b...`、input SHA=`7474e6aa...`。
- 服务器已有TEST客户内容XLSX SHA=`b0498249...`、TEST Word模板SHA=`05c55e77...`，均为liming:liming 0600；LibreOffice 24.2.7.2可用。

## 受限备份

唯一备份目录：`/opt/talent-assessment/backups/mng005_shared_baseline_c51130cb775ba019`，root:root 0700，所有已生成文件由root保护为0600。ownership标记为`MNG005BASE_20261008_c51130cb775ba019`。

- 数据库gzip SHA-256=`068cd19387a8afebd656b5e65e81ea9e0e595dd76a32290f55faf5c6cee12668`
- 应用与报告输入资产归档SHA-256=`f998399c697a0098cbe62fdcd01b0ae1b46984b6c153abfc899c0521ec4f042e`
- 既有报告资产归档SHA-256=`97fb3b304869879421bb439cca39e1d8cb6cbf8b10aad9664303f104975ea97c`

`gzip -t`、两个tar可读性及SHA256SUMS校验均已通过。首次备份命令因普通用户无法穿越root 0700目录执行`gzip -t`而退出1；第二次恢复执行完成文件生成，但普通shell在root目录外展开`*`失败；第三次用root shell完成完整校验，末尾仅打印文件模式错误数的命令因CRLF破坏`wc -l`参数而退出1。该末尾工具错误发生在全部备份完整性检查和SHA输出之后，不影响备份文件；失败均保留，不改写为一次全绿命令。

## 一次性harness范围

新增opt-in Go测试二进制仅在环境门禁、精确hostname、共享库名、active=0、005占用0、reissue/draft/formal表0、两个owned目录0700全部通过后运行。它不监听端口、不启动Worker、不修改在线service/config/cache；以现有服务配置内部取得DSN，不输出DSN、密码、token或人员明文。

预定动作是：

1. 单事务从当前00201/00202复制到独立00501/00502 repo与独立主键，保存明确TEST marker和来源摘要，不更新002。
2. 不安装draft表；以同一canonical构建器创建冻结bundle/profile和两个固定TEST exam。
3. 通过生产`TryRegisterCandidateIdentity`、`CreatePaper`、140次`FillAnswer(raw=3)`、`SubmitParticipant`、`GenerateTestReport`执行链路。
4. 报告写现有revision/current/audit，每产品1份；服务器私有根保留PDF，另复制合成PDF到临时证据目录供下载和本地oracle。
5. 任一步失败自动按两个exact exam ID、两个exact repo ID和关联主键回收本轮TEST行及文件；成功才置为持久保留。

本地门禁：opt-in测试默认SKIP且包编译通过；`go vet ./internal/handler`通过；Linux amd64/CGO0测试二进制52,260,546 bytes，SHA-256=`7bbc8ec2ac940d7a10341cf14583b978673a738b027da851cb18db20af2e06e1`；回滚脚本`bash -n`通过。`gofmt -d`仍返回差异，因此当前harness是**编译/vet通过但格式门禁未关闭**的受控工件；网络阻断后没有为了格式单独重建或上传。

## 失败与清理状态

- 业务写入：0。
- 上传：0。两次SCP均在远端命令前连接超时，目标upload文件未由SCP成功创建。
- 测试binary执行：0。
- 新repo/exam/candidate/paper/run/report/audit/PDF：0。
- 在线应用部署/替换/restart：0；production访问与修改：0。
- 本地生成的Linux测试二进制位于ignored的B区bin，仅为可重建产物；受限staging备份永久保留。
- 回滚脚本已经生成但**未执行**；它要求显式`--execute-exact-owned-rollback`、exact hostname、root-owned 0700报告根、两个exact exam/repo/title/marker、2份completed run与2份TEST report全部满足才可删除。当前共享库无这些owned根，不应运行。

## 网络恢复后的单一后续

先fresh只读确认在线PID/SHA/health、active卷、005占用和report基线；再关闭harness格式门禁、重建并核SHA；随后重新上传并执行。成功后下载synthetic PDF和safe receipt到新的`scripts/test/results/`目录，运行既有 hardened PDF oracle，独立SQL核140 raw3、50分、13/4/receipt、revision/current/audit、文件0600/无symlink/九页/36段/6图，并比较002来源与在线服务前后不变。在线UI仍因未部署005代码而不可查看，最终报告必须继续明确“数据准备完成、UI未发布”。
