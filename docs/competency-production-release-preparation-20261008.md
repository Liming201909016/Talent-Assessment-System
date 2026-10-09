# 胜任力独立生产发布准备（2026-10-08）

## 范围与当前状态

用户明确本次目标为39.106.61.48生产，仅新版一期90题/v2报告；新版管理特质不发布。保留454旧题、48个D维度、9个配置、19份结果及3份报告，新版身份/题本并行新增，不执行009清空、不覆盖旧PDF。

已另批准当前受限备份、在自有临时MySQL5.7库恢复后执行结构迁移首轮/重跑、验证后仅DROP自有临时库。此准备不自动授权业务数据转换、staging批准转production、重算旧结果、清旧到期卷或服务切换。

**最新状态：备份完成、独立MySQL5.7结构首轮/重跑通过；生产尚未切换，整体仍NO-GO。** 本轮限定演练不等完整上线验收。独立前后端发布闭包、旧48D与新十维的运行时并行适配、生产精确题本/内容批准仍未完成，不能上传当前整个工作树或现有管理特质candidate。

## 固定准备流程

- [隔离演练脚本](../scripts/db/competency-production-rehearsal-20261008.sh)：严格主机/root校验，唯一ownership stamp，root0700/文件0600目录；现库全备份含routines/triggers/events与gzip/manifest校验，运行资产与上传树受限归档。预检若有存储定义或cross-schema语句，停止而不自行改写生产归档。
- 六个候选结构脚本仅在恢复库执行：007→008→010→011→012→013。排除001～006盲重跑、009、014、015、全部management_traits迁移。
- 恢复数据库字符集保持utf8mb4/general_ci，原脚本逐SHA。原14张核心表仅投影原列并保存SHA，人员/答案不回显或下载；007空scoring_version补兼容标签的投影例外独立明确，不放宽成绩、答案或PDF指针验证。
- 首轮/重跑结构签名相等、原事实SHA相等、48旧维度保持、管理特质表0、主PID及binary/frontend SHA不变才可称本轮结构演练通过。
- 任一失败停止，EXIT trap仅按精确随机库及ownership DROP自有恢复库，独立查询剩余0；正式备份及失败日志永久保留。
- [Posh驱动](../scripts/tools/competency-production-rehearsal-20261008.ps1)仅接现有生产session，Windows安全凭据不保存；载荷无秘密，分块≤24000字符并校SHA，日志/备份含秘密只留远端受限目录。

## 已发生的工具失败

首次载荷整体超过Posh-SSH命令包上限68536bytes，BeginExecute拒绝；未开始远端shell/备份/迁移。其本地目录426220caef02405e仅approved-inputs，没有payload/结果。模块失败后仍占用本地等待，Ctrl+C结束该未发送命令；没有中断远端迁移。

新运行339f2046959a45d2采用分块载荷和SHA校验，原脚本及旧失败不修改；shell最终bash-n0、PowerShell语法0、编辑器diagnostics0。新载荷receipt存在不等备份或演练成功。

339f2046959a45d2实际备份校验成功，但在恢复后的原事实基线SQL阶段停止，MySQL1064/SQLSTATE42000。独立只读验证发现el_competency_result.columns中`IF(scoring_version=,competency-v1,scoring_version)`，确认shell引号丢失；不是dump损坏或迁移DDL失败。此时已创建12个before文件、六迁移执行0，临时库精确DROP后独立查询0。生产54表/48维度/454题/19结果/3报告保持，PID1195022/两版本SHA/health同。

本次正式备份保留于生产受限目录competency_only_20261008_339f2046959a45d2，root0700；数据库14864794bytes/root0600/SHA `f65ab071074da0d164cba2c0cead4facbe45280e37cbd3135c6a9491a15dd264`，应用资产归档500260278bytes/root0600/SHA `1a47de9329c6d1cb7dbd5b4d9c43cb579785cc250336a97b7d5215580fed3f78`；独立manifest校验exit0。运行资产备份包含配置等秘密，未下载到本机。

三次脚本编辑后停止，用户另明确批准仅一次表达式纠正与一次备份/隔离演练；将空版本期望改为LENGTH=0及ASCII hex，不修改迁移或成绩/答案断言、不增600秒预算。新运行结果以最终收据为准，旧失败收据保留：[第一真实失败](../scripts/test/results/competency-production-rehearsal-20261008-339f2046959a45d2/receipt.json)、[独立备份/运行核验](../scripts/test/results/competency-production-rehearsal-20261008-339f2046959a45d2/independent-check.json)、[夹具错误和清理复核](../scripts/test/results/competency-production-rehearsal-20261008-339f2046959a45d2/failure-audit.json)。

## 发布尚欠的必要工作

1. 当前一期运行时固定competency-a1/b1十个源ID，[配置校验](../Go-based%20Refactored%20System/internal/service/competency_version.go#L23-L26)、[身份/题型契约](../Go-based%20Refactored%20System/internal/service/competency_version.go#L82-L103)；012版本目录并非源主表的自动替代。生产48D表有name/order唯一键，不能直接跑002插AB十维或009清空。用户已选并行保留，需要实现并验证版本作用域身份/题本接入，不靠按名字猜D映射。
2. 从可信来源构造只含competency改动的后端/前端闭包，审阅共享router/人员接口/通用入口，排除管理特质API/Worker/draft/formal/reissue/UI；编译成功不证明范围隔离。
3. 生产内容批准须绑定精确题本、文案、模板SHA及production用途，不能复制stage015签署。未获正式批准不把报告开放为正式人才决策能力。
4. 确认两旧到期competency卷处理及维护窗口，保留199缺测评历史卷；不自动交卷、清理或重设deadline。

## 最终实证

用户额外一次修正后的唯一新运行47dc18ffbb16498e，真实远端exit0：

- 新备份数据库14864794bytes/root0600/SHA `19d30e3bb4c75a2a324e349240eecf409360a55ae33b617157d60c9a6ddcacd2`；应用资产归档500260278bytes/root0600/SHA `1a47de9329c6d1cb7dbd5b4d9c43cb579785cc250336a97b7d5215580fed3f78`，发布子目录root0700。与首失败数据库gzip SHA不同，仅保存实际指纹，不断言差异原因或逻辑数据变化。
- 恢复到competency_verify_47dc18ffbb16498e后，007/008/010/011/012/013首轮及重跑共12次mysql退出0；SECONDS整数计时首轮008/011各1秒、其余0，不将0秒当精确零耗时。
- 两轮所采结构投影SHA均 `e2084a01aa5e2f22d74f3384d506af0a83b065e8a27dbcf9f56c1d2ef6cd6ff5`；14张原表旧列输出投影SHA各before/after相同。007空版本标签按明确期望规范化；新列回填的全量语义尚未逐值审计，原字段输出不是无歧义canonical全库事实证明。
- 48旧维度保持，恢复库MNG表0；EXIT清理后该库0、cleanup0。07:55:25Z独立只读查询两个精确恢复库总残留0、主element54表/48维度/454题/19结果/3报告、主库competency_product_version列0，主库迁移未执行。
- 独立确认实例GTID_MODE=OFF，当前成功dump内GTID_PURGED/SQL_LOG_BIN/CREATE存储定义固定标记计数0；manifest校验0。此实测限定解除本次实例级dump语句疑虑，不把驱动认为可安全用于任意GTID/定义环境。
- 终验SSHexit0，生产PID1195022/NRestarts0/active、后端03397…及首页f5cd…实际完整SHA保持，healthstatusok；自有SSH会话关闭0。不服务切换、不重启Nginx/MySQL、不创建新业务题本或批准内容、不清两旧到期卷/199缺关联卷。

证据：[成功运行收据](../scripts/test/results/competency-production-rehearsal-20261008-47dc18ffbb16498e/receipt.json)、[独立备份/清理/主库/实例终验](../scripts/test/results/competency-production-rehearsal-20261008-47dc18ffbb16498e/independent-final.json)。正式备份及失败证据保留在远端，未下载配置/备份秘密；两轮精确载荷文件0600永久留证，不glob删除其他资源。

### 审阅发现与证据边界

本轮独立只读驱动审阅发现：GTID参数与dump内容门禁不适用于任意实例、cross-schema检测非完整SQL语法解析、签名未包含默认值/引擎/索引前缀/被引用schema、600秒客户端timeout并非远端总时限/强杀清理保证。今回真实未超时、GTID_OFF/dump固定标记0/实际清库0；不以这些限定结果抹去可复用驱动的欠项，不称完整迁移门禁PASS。

后续完整恢复验收还须覆盖所有业务快照/结果/审计表、主库与恢复库对应事实、完整列/默认值/索引前缀/FKschema、二进制读写兼容和故障中断清理。当前只可报告“本次六脚本5.7首轮/重跑及所采投影检查通过”，不能据此生产DDL放行。脚本编辑预算用完，任何后续驱动改动需另明确有界授权，不能复制文件绕预算。

下一逻辑工作为胜任力新旧身份/题本并行适配及隔离发布包；用户已确认保留政策，不需重新问是否保留或是否发布。生产发布批准保留，但必须先补齐上述产品与运行门禁，再进入维护切换。