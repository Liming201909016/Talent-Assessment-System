# 环境发布标记

> 当前标记更新时间：2026-10-09。每次增量发布必须先读取本文件，并在验收后追加新标记；不得覆盖历史标记。

## staging — 当前基线

| 项目 | 已验证值 |
|---|---|
| 主机 | `20.200.136.133 / vm-ubuntu-go-dev` |
| 后端 SHA-256 | `f2940fc5ea51edffc4f325df1f461f3ba4e86868df3aa0594af95764e880d61e` |
| 前端 index SHA-256 | `593d4a20d890d73bf47f9a519a22fffbdb4b539afd5e1cc1dcc2fdebff73cfde` |
| 00401 源内容 | A/B 10维、90题；工作簿 SHA=`828c4267e6c7ad387a73ddb0e923b461d5a336ef225bd7414216c0def814de9f` |
| 00401 验收 | 2组/10维/90答/结果/三Sheet导出/正式10页报告 GREEN |
| 005 验收 | 00501/00502各140题、结果、模板、TEST/reissue报告 GREEN |
| MBTI 验收 | 48答、ESTJ、完整版/简版PDF GREEN |
| 环境状态 | 服务健康；该标记不自动批准production |

## production — release `20261009`

| 项目 | 已验证值 |
|---|---|
| 主机 | `39.106.61.48 / iZ0yosjdcen2p4Z` |
| 代码提交 | 主发布`1753a919...`；controller`227bff9`；FB-219=`f6d0719` |
| 后端/进程 SHA-256 | `f850575b1dfa6eac3f5b4533148baf715eabc8afa13ccd32c7655d0507a7d600` |
| 前端 index SHA-256 | `abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4` |
| 完整发布备份 | `/opt/talent-assessment/backups/production_release_backup_20261009_3e3692ce14634021` |
| FB-219 binary备份 | `/opt/talent-assessment/backups/fb219_binary_20261009_7d5fbd1559914437` |
| 00401内容备份 | `/opt/talent-assessment/backups/phase1_content_20261009_8b17bb3f57794f2c` |
| 005数据 | 14张sidecar表、2 repo、2基线run、模板及4 PDF；客户于16:27通过正式管理员操作将两个005设为`state=0`进行中 |
| 005完整E2E | 00501新建唯一候选、140答、提交、13维/4模块、reissue PDF view/download同SHA；exact cleanup及基线恢复 PASS |
| 00401源内容 | 增量安装A/B 10维+90题，工作簿 SHA=`828c4267e6c7ad387a73ddb0e923b461d5a336ef225bd7414216c0def814de9f`；旧48维仅主数据归档为order101–148，8个重名追加“（历史）”，冻结测评/结果字节不变 |
| 00401完整E2E | 新建/发布2组10维90题、同卷恢复、90答、提交幂等、10维/2组/效度/总体、v1/v2结果、筛选、三Sheet导出 PASS；formal报告因未发布014/015而按设计fail-closed；exact cleanup及全表基线恢复 PASS |
| 00401正式报告 | 未发布正式内容/模板审批包；production保持安全关闭 |
| 服务 | `active`、`NRestarts=0`、8092/8090/root HTTP 200 |

## 增量发布使用规则

1. 发布前记录目标环境当前后端、进程、前端、Schema/content版本及关键数据计数。
2. 只打包“当前标记 → 候选标记”的差异；禁止重新拼装全工作树。
3. 数据变更必须绑定输入文件SHA、完整备份、所有权范围和可执行回滚。
4. 验收收据至少包含：环境、候选SHA、真实HTTP/DB证据、创建资源、cleanup、未覆盖项。
5. 成功后追加新标记；失败只写历史时间线，不移动当前标记。
