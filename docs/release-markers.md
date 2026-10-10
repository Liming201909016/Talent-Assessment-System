# 环境发布标记

> 当前标记更新时间：2026-10-10。每次增量发布必须先读取本文件，并在验收后追加新标记；不得覆盖历史标记。

## staging — 当前基线

| 项目 | 已验证值 |
|---|---|
| 主机 | `20.200.136.133 / vm-ubuntu-go-dev` |
| 后端 SHA-256 | `6c878c8874ae7b08acc2640f18dff0742cf1778bbd699e284b38c430acbfe124` |
| 前端 index SHA-256 | `0e8e02aa654006ce27593227d696494768ed9d2304e447d746f6d25323ccc3cb` |
| 00401 源内容 | A/B 10维、90题；工作簿 SHA=`828c4267e6c7ad387a73ddb0e923b461d5a336ef225bd7414216c0def814de9f` |
| 00401 验收 | 2组/10维/90答/结果/三Sheet导出/正式10页报告 GREEN |
| 005 验收 | 00501/00502各140题、结果、模板、TEST/reissue报告 GREEN |
| MBTI 验收 | 48答、ESTJ、完整版/简版PDF GREEN |
| 环境状态 | 服务健康；该标记不自动批准production |

### production — FB-220～FB-231累计发布 `20261010233739`

| 项目 | 已验证值 |
|---|---|
| 发布边界 | 用户批准本轮全部staging变更；00401只换模板，不重生成历史报告；Schema/题库/正式内容不变 |
| 后端/进程 | SHA=`6c878c8874ae7b08acc2640f18dff0742cf1778bbd699e284b38c430acbfe124` |
| 前端 | index SHA=`0e8e02aa654006ce27593227d696494768ed9d2304e447d746f6d25323ccc3cb`；393文件 |
| 00401 v2模板 | 608305 bytes；SHA=`ef92bfba2026ac2b0609a6b619a41fa29ff8931a9e31569b28fd2f8ace2e24dd` |
| 模板漂移裁决 | staging活动`d4daf935...`实际未通过总体环图中心语义合同；用户明确选择已验收`ef92bfba...`，未推广未知漂移字节 |
| 回滚备份 | `/opt/talent-assessment/backups/production-fb220-fb231-20261010233739`；旧server/template及完整POSIX `dist.before.tar.gz`写前核验 |
| 历史报告保护 | 非backup范围758份PDF数量不变；逐文件清单SHA写前后=`fa8925a8167016cf0feedd8552320548490ed8c9a01e7b242e7267ac75ace263`；重生成0 |
| 运行验收 | PID=`2394050`、NRestarts0；8092 health/8090 root/API均200；captcha code200、浏览器120×40；关键日志/Nginx5xx/tester-list请求均0；payload残留0 |
| 数据边界 | MySQL5.7.44；DB写0；state1及未过期state0均0；Schema/题库/内容未改 |
| 发布结论 | PRODUCTION GREEN（精确发布与匿名入口范围）；认证后业务流程沿既有local/staging验收，本轮未使用管理员凭据复跑 |
| 源码结项 | 提交=`da0cc81`，已推送`origin/master`；UF-068发布后业务纠正见下节 |

#### 发布后业务门禁纠正 `202610102340`

| 项目 | 已验证值 |
|---|---|
| 运行资产 | 上述server/index/template、健康、验证码、历史PDF保护收据继续有效 |
| 新发现阻断 | production缺`competency-phase1-content-v2/frontline_employee/development/competency-logical-reasoning/good`；日志连续三次命中 |
| 影响 | 命中该维度/等级的新报告生成及批量准备RED；不是服务宕机或答案/分数丢失 |
| 当前结论 | 精确部署GREEN；00401 production新报告业务链PARTIAL/RED，待正式内容补齐后重新验收 |
| 操作边界 | 本轮只读日志；DB写0、报告重生成0、服务操作0 |

### staging — FB-231测评人员页空范围403 `20261010232709`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 仅前端测评人员页；空examId不调用旧全量接口，精确examId流程不变；后端/DB/模板/production不变 |
| 前端 | index SHA=`0e8e02aa654006ce27593227d696494768ed9d2304e447d746f6d25323ccc3cb`；393文件；tester lazy chunk SHA=`f229a1131e012a2dc2488ab42214b636cfa9b3aaa999d8cbd8a7db2aa65ed367` |
| 回滚备份 | `/opt/talent-assessment/backups/fb231-staging-frontend-20261010232709/dist.before.tar.gz` |
| 测试 | RED空范围API调用1→GREEN0；人员专项8/8；前端35文件617项；production build通过 |
| 远端验收 | root/health200；发布后tester/list请求及403=0；PID3394、NRestarts0；service restart0；DB写0 |
| 独立漂移 | preflight发现活动00401模板SHA=`d4daf935b97e303561976f5e437c1c873b7e27030e9a8e7b6b7ba5152bebd3b1`，与此前标记不同；本前端only发布原样保留，未追溯或覆盖 |
| 限制 | 无真实管理员浏览器DOM收据；远端部署chunk与本地测试字节精确一致，待用户刷新后人工复核 |
| 发布边界 | STAGING GREEN（FB-231范围）；production未修改 |

### staging — FB-230总体环图中心语义 `20261010230534`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 00401 v2模板中心语义层及指定报告重生成；后端、前端、Schema、题库、评分与production不变 |
| 最终模板 | 608305 bytes；SHA=`ef92bfba2026ac2b0609a6b619a41fa29ff8931a9e31569b28fd2f8ace2e24dd`；Word实开；59/60可选字段、12图、零外链/公式 |
| 模板备份 | `/opt/talent-assessment/backups/fb230-staging-template-20261010230534` |
| 报告备份 | `/opt/talent-assessment/backups/fb230-staging-report-20261010230551` |
| 最终报告 | paper=`24504b9c-1874-4bbd-af09-f9d0d83abb16`；report=`7426a5dc-3020-4917-9f65-0b0f14b86cb0`；813847 bytes/SHA=`c6fbd4bb55f4f88d2c46a104051e5e9e833cd2313dadb0942229c5f6966a57d2`；A4 10页 |
| 视觉/文本验收 | 环图中心显示`总体评价`、动态实际值`60.94`和`分`；物理2/3页码及十维图保持；下载PDF文本label=1、score-unit匹配=2 |
| 运行验收 | backend/process=`6c878c88...`、front index=`b0576cc...`、PID3394、NRestarts0、health200、active0、关键日志0、temp0 |
| 发布边界 | STAGING GREEN；production未修改，不自动批准production |

### staging — FB-229页码与总体环图模板 `20261010224819`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 00401 v2模板及指定历史报告重生成；后端、前端、Schema、题库、评分与production不变 |
| 最终模板 | 608015 bytes；SHA=`b8d6766290e4ba3b23ac9c98bfd76ce169b90c28f884d88b8f60fc6d8f6e77c7`；Word实开；59/60可选字段、12图、零外链/公式 |
| 模板备份 | 原活动模板：`/opt/talent-assessment/backups/fb229-staging-template-20261010224608`；中间模板：`/opt/talent-assessment/backups/fb229-staging-template-v2-20261010224819` |
| 报告备份 | 原报告：`/opt/talent-assessment/backups/fb229-staging-report-20261010224628`；中间报告：`/opt/talent-assessment/backups/fb229-staging-report-v2-20261010224840` |
| 最终报告 | paper=`24504b9c-1874-4bbd-af09-f9d0d83abb16`；report=`7426a5dc-3020-4917-9f65-0b0f14b86cb0`；812318 bytes/SHA=`f677860b6a4dd6a6011c25908ef5f05b0fc600a4e9d572df0f3c4f17788751f6`；A4 10页 |
| 视觉验收 | LO24.2：物理第2/3页显示`第 1 页`/`第 2 页`；报告概览总体环图可见；十维对比图保持可见 |
| 运行验收 | backend/process=`6c878c88...`、front index=`b0576cc...`、PID3394、NRestarts0、health200、active paper0、关键日志0、temp0 |
| 过程纠正 | 首轮模板恢复环图但LO24.2仍抑制section first footer；新增目标引擎RED后移除该节first/titlePg。首轮generate实际成功但wrapper误判code0，失败证据保留 |
| 发布边界 | STAGING GREEN；production未修改，不自动批准production |

### staging — FB-228批量ZIP current-v2绑定 `20261010220324`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 仅后端批量报告加载器：current→completed report→completed v2 run；无current保持legacy；前端、模板、Schema、评分和production不变 |
| 后端/进程 SHA-256 | `6c878c8874ae7b08acc2640f18dff0742cf1778bbd699e284b38c430acbfe124`；50275180 bytes |
| 旧后端备份 | `/opt/talent-assessment/backups/fb228-staging-backend-20261010220324/server.before`；SHA=`e221ab4512724e9c7748f4d78c88618930bb3e29fdd64eb99a5c0c66550cffd0` |
| 测试 | FB-228 RED→GREEN；相邻archive/download测试、Go全量、全build通过 |
| 真实验收 | paper=`24504b9c-1874-4bbd-af09-f9d0d83abb16`；HTTP200 `application/zip`；693998-byte ZIP/1 entry；PDF 820207 bytes/SHA=`99c40ab80a33d74b5742dba11c4259fa82dd70b83141c92993c74b1b3dca143c` |
| 审计与清理 | download audit19→20；临时Redis会话EXISTS0；远端ZIP不存在 |
| 服务 | PID=`3394`、NRestarts0、health200；active paper0；rollback未触发 |
| 发布边界 | STAGING GREEN；production未修改，不自动批准production |

### staging — FB-227登录入口恢复 `20261010214846`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 仅重建并原子替换前端；显式`VUE_APP_BASE_API=/prod-api`；产品源码逻辑、后端、模板、Schema、数据及production不变 |
| 前端 | index SHA=`b0576cc20270f7d9a70d85da6f6c9562c868417ba511c25a07695009095de983`；393文件；bundle含`/prod-api` |
| 回滚备份 | `/opt/talent-assessment/backups/fb227-staging-frontend-20261010214846/dist.before.tar.gz`，写前核验旧index SHA=`16455aef...` |
| 本地验证 | Vue 35文件/616项通过；production build通过 |
| 浏览器无Token | 登录页正常；请求`/prod-api/captchaImage`；验证码图120×40、无invalid image |
| 浏览器旧Token | 请求`/prod-api/getInfo`返回401后清Token并回登录页；pageErrors0，无avatar TypeError、405或重复提交通知 |
| 运行验收 | backend/process=`e221ab45...`、PID=`1535`、NRestarts0、root/health/captcha200、active paper0 |
| 操作边界 | 数据库写0；后端服务重启0；production未修改 |

### staging — FB-226精确重新发布 `20261010213613`

| 项目 | 已验证值 |
|---|---|
| 发布性质 | 对上一标记的后端、前端、00401 v2模板做exact-byte republish；无新源码、Schema、题库、配置或内容差异 |
| 后端/进程 SHA-256 | `e221ab4512724e9c7748f4d78c88618930bb3e29fdd64eb99a5c0c66550cffd0` |
| 前端 | index SHA=`16455aefeec9df8fda97298b32f0cfe90d1fa83860b0f9433e500e81e0c3d86f`；393文件 |
| 00401 v2模板 | SHA=`75b79b93e36a3a8dee83abb1d822faa5fae49a500597389c5bc5c63f2c3cd6dc` |
| 新回滚备份 | `/opt/talent-assessment/backups/fb226-staging-republish-20261010213613`；server/template精确副本、完整POSIX `dist.before.tar.gz`及SHA清单；tar内index已写前核验 |
| 元数据复验 | 00401=`60`；005=`95`/required=`90`；全部repeatable；临时Redis管理员会话cleanup completed/EXISTS0 |
| 服务验收 | active、PID=`26356`、NRestarts=0、talent-assessment/nginx/mysql active、内外及公网health200、active paper0、关键日志0、临时残留0 |
| 数据边界 | 数据库写0；未重新生成报告；沿用上一标记的真实报告验收收据 |
| 发布边界 | STAGING GREEN；production未修改，不自动批准production |

#### 发布后浏览器门禁纠正 `202610102141`

| 项目 | 已验证值 |
|---|---|
| 页面加载 | `http://20.200.136.133/#/`静态页面、CSS、JS和背景图可加载，跳转登录页 |
| 阻断 | 验证码图片为`data:image/gif;base64,undefined`，naturalWidth/Height=`0/0`，控制台`ERR_INVALID_URL` |
| 请求证据 | bundle请求`/captchaImage`得到HTTP200 `text/html`首页（16290 bytes）；正确的`/prod-api/captchaImage`得到HTTP200 JSON、`captchaOnOff=true`、PNG base64长度1112及uuid |
| 判定 | 🔴 浏览器登录NO-GO；上方服务/SHA/API收据仍有效，但不得再称完整STAGING GREEN |
| 操作边界 | 只读检查；未登录、未改远端、未改数据库、未改production |

### staging — FB-224～FB-226 / UF-059～UF-061增量 `20261010212040`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 后端可选/重复模板字段、批量报告准备、v2百分制投影、待发展短评、客户模板兼容；前端模板字段弹窗及批量准备；00401 v2模板替换。Schema/题库/005模板/production不变 |
| 后端/进程 SHA-256 | `e221ab4512724e9c7748f4d78c88618930bb3e29fdd64eb99a5c0c66550cffd0`；50238404 bytes |
| 前端 | index SHA=`16455aefeec9df8fda97298b32f0cfe90d1fa83860b0f9433e500e81e0c3d86f`；393文件 |
| 00401 v2模板 | 617264 bytes；SHA=`75b79b93e36a3a8dee83abb1d822faa5fae49a500597389c5bc5c63f2c3cd6dc` |
| 旧资产 | backend=`13d5f07e...`；index=`593d4a20...`；template=`52e0020c...` |
| 运行备份 | `/opt/talent-assessment/backups/fb226-staging-20261010212040`；server/template精确旧字节及SHA清单；首次脚本的dist zip为空，不作为前端回滚源 |
| 报告备份 | `/opt/talent-assessment/backups/fb226-staging-report-20261010212600`；report/current/audit SQL及旧PDF副本 |
| 元数据验收 | 认证接口00401=`60`全可选/可重复；005=`95`（90必需+5可选）全可重复；临时Redis会话cleanup completed/EXISTS0 |
| 报告验收 | paper=`24504b9c-1874-4bbd-af09-f9d0d83abb16`；report=`7426a5dc-3020-4917-9f65-0b0f14b86cb0`；820207 bytes/SHA=`7150a3a85a987a34a0a371acd5bbbf93f27f1708295c5fe418816b2f74605b3a`；A4 10页；时长/分钟/总页提示0；视觉抽检通过 |
| 服务验收 | active、PID=`25353`、NRestarts=0、talent-assessment/nginx/mysql active、内外health200、active paper0、关键日志0、远端临时残留0 |
| 失败/纠正 | 首次回滚trap递归导致start-limit，Windows ZIP路径导致空dist；确认旧SHA后以POSIX tar恢复前端和旧服务，再用非递归rollback完成新发布。失败证据保留，不影响最终运行SHA |
| 发布边界 | STAGING GREEN；production未修改，不自动批准production |

### staging — UF-058 renderer增量 `20261010092031`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 仅后端renderer可见动态字段wrapper兼容；模板、前端、Schema、题库、字体及配置不变 |
| 后端/进程 SHA-256 | `13d5f07e5d157db5a379c014b43c058094800fc01c7d5768ea1be419b7d43415`；50198351 bytes |
| 旧后端备份 | `/opt/talent-assessment/backups/uf058-staging-server-20261010092029/server.before`；SHA=`ec4c85d73fd2e1e8acb719b137b0b59309bc9de2f8a8e3f80653b7f00e5ad5` |
| 模板 | v2 SHA=`52e0020c5a6f39535d020bf41b18d9412c096ab43ce6608964b0101b955f30b9`，未变 |
| 报告验收 | report=`7426a5dc-3020-4917-9f65-0b0f14b86cb0`、paper=`24504b9c-1874-4bbd-af09-f9d0d83abb16`；814516 bytes/SHA=`e9eb405c0f9cf474187b27901ad6811d1ca9439334cf026ea40d66cff61c4237` |
| PDF合同 | 10页、免责声明MicrosoftYaHei、SimSun=0、Page前缀0、`计划执行：`粗体1/1 |
| 报告回滚 | `/opt/talent-assessment/backups/uf058-staging-report-20261010092417`；覆盖前PDF SHA=`4d98dc3bd7f9efde04a6901779c2f867cf3cef51d05c38f189d3e3b09854494d`，report/current/audit SQL副本齐全 |
| 服务验收 | active、PID=`3245`、NRestarts=0、内外health200、active paper0、关键日志0、临时残留0；report/current=1、audit=24 |
| 发布边界 | STAGING GREEN；不自动批准production，production未修改 |

### production — UF-058 renderer增量 `20261010094424`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 仅后端renderer及指定00401历史报告重生成；模板、前端、Schema、题库、字体及配置不变 |
| 后端/进程 SHA-256 | `13d5f07e5d157db5a379c014b43c058094800fc01c7d5768ea1be419b7d43415`；50198351 bytes |
| 旧后端备份 | `/opt/talent-assessment/backups/uf058-production-server-20261010094424/server.before`；SHA=`c321bf35af790a855a342995f39427f72a874dcc494fcd984c5142760f2cec87` |
| 模板 | v2 SHA=`52e0020c5a6f39535d020bf41b18d9412c096ab43ce6608964b0101b955f30b9`，未变 |
| 报告验收 | report=`103d9a0d-0113-4aef-8962-434645783477`、paper=`9174f98e-181f-486e-bbc7-0118bc39ced1`；709795 bytes/SHA=`89896d6f54dc56d19a8ac7f0310bf216da9f8a248e502a9b8a906aabdfa5bd98` |
| PDF合同 | LibreOffice 7.4；10页、免责声明MicrosoftYaHei、SimSun=0、Page前缀0、`计划执行：`粗体2/2，数字页码居中 |
| 报告回滚 | `/opt/talent-assessment/backups/uf058-production-report-20261010095539`；覆盖前PDF SHA=`cd6dbce48c1947abaf0ed21e02ba96eed1d433beb1002b2676157dbfda04825f`，report/current/audit SQL副本齐全 |
| 绑定与审计 | report/current=1；audit20→generate21→认证下载22；下载字节与DB/文件SHA一致 |
| 服务验收 | active、PID=`2382320`、NRestarts=0、内外及公网health200、active paper0、关键日志0、payload0 |
| 发布边界 | PRODUCTION GREEN；未改其他产品数据或资产 |

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
| 00401正式报告v2 | 已安装124条正式文案、production approved包及v2模板SHA=`f98599939e3bf7923abf8bd457e30259cde3fa68d6c4e9dc9d3288de0a2eae16`；备份=`/opt/talent-assessment/backups/phase1_report_20261009_8af8c5dd2963418f` |
| 00401正式报告v1 | 默认新卷绑定v1；66条正式文案/approved包和精确模板SHA=`54b167fcc02737ba44e0432a930e1f5657db57cec87d9063f1ee7c111c4932bf`尚未安装，完整报告E2E为RED；最近两轮preflight均在备份/写入前失败，production无新增备份/业务写，v1模板仍不存在 |
| 服务 | `active`、`NRestarts=0`、8092/8090/root HTTP 200 |

## production — release `20261009-phase1-staging-sync`

| 项目 | 已验证值 |
|---|---|
| 主机 | `39.106.61.48 / iZ0yosjdcen2p4Z` |
| 兼容边界 | staging基线，仅保留用户确认的FB-219 MySQL5.7兼容补丁 |
| 后端/进程 SHA-256 | `f850575b1dfa6eac3f5b4533148baf715eabc8afa13ccd32c7655d0507a7d600` |
| 前端 index SHA-256 | `593d4a20d890d73bf47f9a519a22fffbdb4b539afd5e1cc1dcc2fdebff73cfde`；393文件 |
| 完整发布备份 | `/opt/talent-assessment/backups/phase1_staging_sync_20261009212420` |
| 00401题库 | 10维/90题，dimension SHA=`ac6d4290...`、question SHA=`db1554e7...`，与staging完全一致且本轮无需重写 |
| 00401正式内容 | v1/v2文案=`66/124`、approved包=`1/1`；文案SHA=`c52b2b19.../35cf08ee...` |
| 00401模板 | v1=`54b167fcc02737ba44e0432a930e1f5657db57cec87d9063f1ee7c111c4932bf`；v2=`a814c36e3759c8ff2cf5c17148e0939b477f17d4d3f430bb6a22f530022f9f5a` |
| Renderer | `PHASE1_WORD_REPORT_ENABLED=true`，独立systemd drop-in，process环境已核验 |
| 完整E2E | 2组/10维/90题/90答/提交幂等/v1+v2结果/筛选/三Sheet/approved 10页PDF/审计2 PASS；exact cleanup和全表baseline恢复PASS |
| 服务 | PID=`2367815`、`NRestarts=0`、健康200、最近10分钟fatal/panic/permission=0 |

### production — 00401历史清理增量 `20261009220803`

| 项目 | 已验证值 |
|---|---|
| 完整备份 | `/opt/talent-assessment/backups/phase1_history_delete_v2_20261009220803`（数据库+5份报告文件） |
| 删除闭包 | 旧9测评、22试卷、19结果、5报告、11审计、454题、48维及冻结/试卷引用 |
| 当前00401 | 10维、90题；分布=`90|90|80|10|62|18|10|90` |
| staging一致性 | dimension SHA=`ac6d4290...`、question SHA=`db1554e7...`，production/staging一致 |
| 保护范围 | 传统001/002/003/005题858保留；v1/v2文案66/124、包1/1及模板不变 |
| 服务 | PID=`2370974`、`NRestarts=0`、server/process=`f850575b...`、health200 |

### production — UF-055静态站点权限修复 `2026-10-09`

| 项目 | 已验证值 |
|---|---|
| 原因 | `/opt`及`/opt/talent-assessment`为0700，Nginx worker无法穿越，root/favicon 500；API proxy保持200 |
| 变更 | 用户批准两目录改为0755；不改文件、DB、Nginx配置或程序，不重启 |
| 公网验收 | root=`200/16155 bytes`、favicon=`200/26900 bytes`、API health=`200/15 bytes` |
| 浏览器验收 | 页面标题“人才综合素质评估系统”，进入登录页 |
| 服务 | PID=`2370974`、`NRestarts=0`，与修复前一致 |

### staging / production — UF-056 v2模板增量 `20261009232349`

| 项目 | 已验证值 |
|---|---|
| 变更边界 | 仅`competency-phase1-report-v2.docx`；不改程序、前端、字体、Schema、题库、Nginx或systemd |
| 原模板 | SHA=`a814c36e3759c8ff2cf5c17148e0939b477f17d4d3f430bb6a22f530022f9f5a`，685600 bytes |
| 新模板 | SHA=`52e0020c5a6f39535d020bf41b18d9412c096ab43ce6608964b0101b955f30b9`，611455 bytes；解包只变`word/document.xml` |
| staging动态验收 | 2组/10维/90题/90答/v1-v2结果/LibreOffice/10页PDF；首页时长隐藏、审计2、cleanup0、基线无漂移 |
| staging备份 | `/opt/talent-assessment/backups/uf056-template-20261009230941` |
| production备份 | `/opt/talent-assessment/backups/uf056-template-20261009232349` |
| production数据边界 | `databaseAccess=false`、`databaseWrites=false`；旧报告未重生成 |
| production验收 | service active、PID=`2370974`、NRestarts0、内外root/API=`200/200`、公网root 16155 bytes、health ok、关键日志0 |

## 增量发布使用规则

1. 发布前记录目标环境当前后端、进程、前端、Schema/content版本及关键数据计数。
2. 只打包“当前标记 → 候选标记”的差异；禁止重新拼装全工作树。
3. 数据变更必须绑定输入文件SHA、完整备份、所有权范围和可执行回滚。
4. 验收收据至少包含：环境、候选SHA、真实HTTP/DB证据、创建资源、cleanup、未覆盖项。
5. 成功后追加新标记；失败只写历史时间线，不移动当前标记。
