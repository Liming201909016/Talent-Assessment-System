# 项目当前状态（后续开发与测试的唯一现状入口）

> 更新时间：2026-10-10
> 本文件只保存**当前有效结论**；过程、失败和纠正历史保留在 [project-memory.md](project-memory.md)。发生冲突时，按“本文件 → project-memory 顶部最新纠正 → 日期更晚的限定事实 → 历史记录”读取。

## 1. 环境边界

| 环境 | 当前结论 |
|---|---|
| local | UF-058可见动态字段wrapper修复已GREEN：专项、Go全量、build及production LibreOffice 7.4零DB临时转换通过；同一候选已完成staging/production发布验收。仍有范围外未提交改动，不以整个工作树代表已部署字节 |
| staging | 🟢 `20.200.136.133 / vm-ubuntu-go-dev`；后端/进程SHA=`6c878c8874ae7b08acc2640f18dff0742cf1778bbd699e284b38c430acbfe124`，前端index=`0e8e02aa654006ce27593227d696494768ed9d2304e447d746f6d25323ccc3cb`。FB-231人员页空examId零请求已发布，发布后tester/list 403=0；PID3394/NRestarts0、health200 |
| production | 🟡 `39.106.61.48 / iZ0yosjdcen2p4Z`；后端/进程SHA=`6c878c8874ae7b08acc2640f18dff0742cf1778bbd699e284b38c430acbfe124`，前端index=`0e8e02aa654006ce27593227d696494768ed9d2304e447d746f6d25323ccc3cb`/393文件，00401 v2模板=`ef92bfba2026ac2b0609a6b619a41fa29ff8931a9e31569b28fd2f8ace2e24dd`；运行健康，但新生成00401报告因production缺少`逻辑思维/good/development`正式内容而失败关闭，待补数据后复验 |

[UF-068 production新版结果服务不可用 - 2026-10-10 23:40] 发布后用户看到“新版结果服务暂不可用”。production日志连续三次给出受控内部原因：`competency-phase1-content-v2/frontline_employee/development/competency-logical-reasoning/good`内容缺失。服务、答案和已持久化分数未丢失；FB-222代码按设计对待发展短评缺行失败关闭。当前00401既有历史PDF仍保持，但涉及该精确维度/等级组合的新报告生成与批量准备为RED。尚未取得触发exam/paper ID，未补内容、未重生成报告、未写数据库。下一步必须先以production内容完整性RED覆盖十维×所需等级，再受控补齐正式`development`内容并复验单份生成与批量下载。

[FB-220～FB-231 production累计发布 - 2026-10-10 23:37] 用户批准“本轮全部staging变更”并明确00401“只换模板、不重生成历史报告”。写前只读门禁确认MySQL5.7.44、server/process=`13d5f07e...`、index=`593d4a20...`/393、模板=`52e0020c...`、state1及未过期state0均0、内外health200。当前staging模板`d4daf935...`未通过总体环图中心语义合同，用户明确选择已验收`ef92bfba...`；其余候选为staging运行精确字节。完整回滚备份=`/opt/talent-assessment/backups/production-fb220-fb231-20261010233739`，一次停启后server/process=`6c878c88...`、index=`0e8e02aa...`/393、模板=`ef92bfba...`，PID=`2394050`、NRestarts0、8092/8090/root均200、验证码code200且浏览器120×40、关键日志/Nginx5xx/tester-list请求均0、payload残留0。758份历史PDF数量及清单SHA=`fa8925a8...`写前后相同，DB写0、历史报告重生成0、rollback未触发。

[生产发布 - 2026-10-09] `39.106.61.48 / iZ0yosjdcen2p4Z`已部署封存release；完整备份=`/opt/talent-assessment/backups/production_release_backup_20261009_3e3692ce14634021`。真实管理员验收发现并修复FB-219，补丁提交=`f6d0719`、binary备份=`/opt/talent-assessment/backups/fb219_binary_20261009_7d5fbd1559914437`。最终MySQL 5.7主库有14张`el_mng_*`表、005 repo/exam/profile/snapshot/run/revision/reissue各2、reissue audit12；两个exam均恢复`state=1`。共享DOCX SHA=`05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c`，内容XLSX SHA=`b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c`。真实管理员状态往返、模板、结果和647985-byte reissue PDF读取均通过。

[生产E2E补充 - 2026-10-09] 00501真实创建唯一候选、140答、提交、13维/4模块、reissue PDF view/download同SHA，baseline逐表恢复。用户随后由外部管理员会话将00501/00502均改为`state=0`进行中，属于客户控制状态。00401经用户明确批准导入staging已验收工作簿SHA=`828c4267e6c7ad387a73ddb0e923b461d5a336ef225bd7414216c0def814de9f`：新增A/B 10维+90题；旧48维主数据order归档至101–148，8个重名追加“（历史）”，旧冻结测评/结果字节不变；完整备份=`/opt/talent-assessment/backups/phase1_content_20261009_8b17bb3f57794f2c`。随后真实00401完成2组/10维/90题发布、90答、提交幂等、v1/v2结果、筛选、三Sheet导出并exact cleanup。用户随后授权沿用staging具名批准并同步正式报告：production已受控安装124条v2文案、approved production v2包及精确v2模板SHA=`f98599939e3bf7923abf8bd457e30259cde3fa68d6c4e9dc9d3288de0a2eae16`，备份=`/opt/talent-assessment/backups/phase1_report_20261009_8af8c5dd2963418f`；但新建测评仍绑定v1版本元组，完整报告E2E因此保持RED且已精确清理。v1精确模板SHA=`54b167fcc02737ba44e0432a930e1f5657db57cec87d9063f1ee7c111c4932bf`和66条确定性内容已在本地重建；两轮production写前preflight均在备份/写入前失败，最终确认服务PID=`2355899`、`NRestarts=0`、v1模板仍不存在、payload/probe均清零。当前不得将00401正式报告称为production GREEN。

[全量staging替换回滚 - 2026-10-09] staging SSH恢复并完成实时差异盘点；题库10维/90题及内容SHA已与production一致，Schema仅有MySQL版本元数据、staging迁移标记表及production额外安全索引差异。受控替换创建全量备份=`/opt/talent-assessment/backups/phase1_staging_sync_20261009211916`并通过文件/DB结构验收，但真实00401创建测评被staging旧后端的管理特质MySQL5.7保护校验拒绝。已立即完整回滚并验证：PID=`2366975`、NRestarts=0、后端/进程=`f850575b...`、前端=`abf93dd1...`、健康200、v1仍0/0/absent、v2保持124/1及原模板。当前状态矩阵和P0正式报告RED结论不变。

[00401正式报告GREEN - 2026-10-09] 用户选择staging基线仅保留FB-219兼容补丁后重新受控发布，完整备份=`/opt/talent-assessment/backups/phase1_staging_sync_20261009212420`。最终server/process=`f850575b...`、staging前端index=`593d4a20...`/393文件、v1/v2文案=`66/124`、approved包=`1/1`、模板SHA=`54b167fc.../a814c36e...`，公开renderer开关已通过独立systemd drop-in启用。真实production 00401完整链含10页正式PDF、审计2、三Sheet导出及exact cleanup全部PASS；PID=`2367815`、NRestarts=0、最近fatal/panic/permission=0。

[00401旧测试历史清理 - 2026-10-09] 完整备份=`/opt/talent-assessment/backups/phase1_history_delete_v2_20261009220803`；已删除旧9测评/22试卷/19结果/5报告/454题/48维完整测试闭包。production当前仅保留10维90题，分布和dimension/question SHA与staging实时收据完全一致；传统858题、正式文案/包/模板保持。PID=`2370974`、NRestarts=0、health200、MySQL临时写PASS。

[staging基准只读复核 - 2026-10-09] production前端及393文件、全部9份模板、00401 10维90题/正式内容均与staging精确一致；后端仅保留已批准FB-219差异。00401字段级Schema差异只有MySQL8/5.7元数据，staging多迁移标记表，production多安全索引，FK一致。扩展到全题库后发现传统题总数staging/production=`1135/858`：六个共享repo题数及语义SHA一致，但production缺staging `00301`那套48题及`00302`身份（production `00301`内容实际匹配staging `00302`），另有production专有`10201`一题；四个repo元数据及005 definition bundle语义也不同。故00401范围GREEN，但全题库覆盖同步未获准且当前NO-GO。

[UF-055 production静态站点恢复 - 2026-10-09] 公网root/favicon曾500而API health200；根因是旧失败rollback把`/opt`与应用父目录mode保留为0700，Nginx worker无法穿越。用户批准仅将两目录修复为0755，无文件/DB/配置/程序改动且无重启。最终公网root/favicon/API=`200/200/200`，真实浏览器登录页加载；后端PID=`2370974`、NRestarts=0不变。

[production部署只读审计 - 2026-10-09] 当前runtime/Nginx/MySQL/Redis/匿名浏览器均GREEN，当前PID启动后应用错误0；但安全与运维NO-GO项仍在：HTTP登录无TLS、含秘密配置0644、Go以root运行且systemd评分9.6 UNSAFE、root SSH、UFW规则过宽、409条孤儿paper、磁盘80%及备份/上传增长。完整证据见[production-deployment-audit-20261009.md](production-deployment-audit-20261009.md)。本轮零写，未改现场。

[00401 production保留浏览器链 - 2026-10-09] marker=`TEST-BROWSER-00401-20261009223722`的真实候选/试卷/90答/v1-v2结果/正式报告/PDF按用户要求保留；刷新恢复、提交确认、管理员列表与90题详情、浏览器PDF 200均通过。PDF为10页/717300 bytes/SHA=`145bc06b...`，审计7条；短期管理员会话、远端临时脚本、本地token及SSH均已清理。该保留PDF生成于UF-056模板发布前，仍显示`时长：1分钟`；它是历史文件，不代表当前活动模板状态。

[UF-056模板发布 - 2026-10-09] staging先安装候选并以真实90答、LibreOffice转换、10页PDF验证首页时长隐藏，cleanup0且业务基线不变。用户批准production仅替换模板且不涉及数据库后，production活动v2模板从`a814c36e...`原子更新为611455-byte SHA=`52e0020c...`，只变`word/document.xml`；备份=`/opt/talent-assessment/backups/uf056-template-20261009232349`。服务PID=`2370974`、NRestarts0、内外root/API 200、关键日志0，零数据库访问/写入。保留旧PDF未重生成，仍含历史时长文本；新生成或明确重生成的v2报告使用新模板。

[UF-057历史报告覆盖 - 2026-10-10] 用户批准直接覆盖保留paper=`9174f98e-181f-486e-bbc7-0118bc39ced1`的旧PDF。根因是LibreOffice 7.4把嵌套内容控件中的PAGE字段当文本且忽略动态标签run粗体；renderer在渲染后仅移除冗余wrapper、统一PAGE字段并居中内层段落。production同机确定性DOCX先验证10页、封面无页码、正文数字1–9居中、Page字面量0及`计划执行：`粗体，再发布后端SHA=`c321bf35...`，binary备份=`/opt/talent-assessment/backups/uf057-production-lo74-server-20261010001507`。指定report=`103d9a0d-0113-4aef-8962-434645783477`从SHA=`7cd40b29...`/716626 bytes覆盖为`cf1385c5...`/758202 bytes；精确备份=`/opt/talent-assessment/backups/uf057-report-20261010001515`，report/current保持1/1，audit最终15，认证下载与绑定一致。PID=`2375402`、NRestarts0、内外HTTP 200、关键日志0、active paper0、临时payload已清理。

[UF-058字体漂移本地修复 - 2026-10-10] production只读复核确认活动模板仍为SHA=`52e0020c...`且免责声明模板run显式微软雅黑；当前report/current实际绑定PDF为758202 bytes/SHA=`cd6dbce4...`。根因是LibreOffice 7.4对保留`w:sdt` wrapper的已填充值采用SimSun。RED后最小修复为：填充前继续严格验证全部Tag合同，填充后对可见动态字段移除wrapper并保留内部run/`w:rPr`，隐藏时长合同不变。专项、Go全量、build通过；production同机仅`/tmp`隔离转换的确定性probe为10页、SimSun=0、免责声明MicrosoftYaHei、Page前缀0、计划执行粗体1/1。当前仅LOCAL GREEN，未部署staging/production、未覆盖历史PDF。

[UF-058 staging发布验收 - 2026-10-10] SSH恢复后fresh核验确认后端于09:20:31切换为Linux候选SHA=`13d5f07e...`，server/process一致；旧后端SHA=`ec4c85d7...`保存在`/opt/talent-assessment/backups/uf058-staging-server-20261010092029/server.before`。复用completed v2 report=`7426a5dc-3020-4917-9f65-0b0f14b86cb0`重生成后为814516 bytes/SHA=`e9eb405c...`，report/current=1、audit=24；PDF 10页、免责声明仅MicrosoftYaHei字体子集、SimSun=0、Page前缀0、`计划执行：`粗体1/1。模板仍为SHA=`52e0020c...`；服务active、PID=`3245`、NRestarts0、内外health200、active paper0、关键日志0、临时残留0。当前升级为STAGING GREEN；production仍未部署或覆盖。

[UF-058 production发布验收 - 2026-10-10] 用户明确批准后，production零写预检确认旧server/process=`c321bf35...`、模板=`52e0020c...`、active paper0、report/current=1及当前PDF=`cd6dbce4...`。后端单次切换为staging同候选SHA=`13d5f07e...`，旧后端备份=`/opt/talent-assessment/backups/uf058-production-server-20261010094424`。指定report=`103d9a0d-0113-4aef-8962-434645783477`强制重生成并备份至`/opt/talent-assessment/backups/uf058-production-report-20261010095539`；新PDF=709795 bytes/SHA=`89896d6f...`，10页、免责声明MicrosoftYaHei、SimSun0、Page前缀0、计划执行粗体2/2、report/current=1、audit=22，控制器认证下载与DB/文件绑定一致。服务active、PID=`2382320`、NRestarts0、内外及公网health200、关键日志0、active paper0、上传残留0。UF-058当前PRODUCTION GREEN。

[全项目待办评估 - 2026-10-10] 已完成代码、前端、测试、CI、运维和数据治理只读审计，完整清单见[project-backlog-assessment-20261010.md](project-backlog-assessment-20261010.md)。本地实测：Go全量测试exit0/总覆盖率45.7%、server build通过；Vue 35文件/611测试通过，覆盖率`60.25/95.53/39.13/60.25`，production build通过。新增最高优先级代码安全项为匿名`online-paging`暴露部门限定考试及任意Origin CORS同时允许credentials；CI Go 1.24与`go.mod` 1.26不一致，前端测试/lint/coverage和gosec允许失败不阻断。当前工作树100条变更（63 deleted/12 modified/25 untracked），仍不得作为发布源。

[production孤儿试卷保留例外 - 2026-10-10] 用户选择“保留并登记例外”。production零写预检确认409条paper/14个缺失父exam，state0/state2=`199/210`、active0、paper_qu20566（answered1467）、paper_qu_answer42963、409条均绑定现存sys_user，candidate/tester/MBTI引用均0；旁链另有exam_repo24、user_exam19、user_book43。最早完整备份`release_00401_20260727_153001/element.sql.gz`也未找到14个父exam（0/14），不构造tombstone、不删除历史。14组脱敏指纹=`26836df2...e21b`，最终只读门禁status/baseline PASS、drift0、databaseWrites0、cleanup completed。完整证据见[production-paper-orphan-governance-20261010.md](production-paper-orphan-governance-20261010.md)。

[production磁盘治理 - 2026-10-10] 用户批准仅清理可再生系统数据，不触碰业务备份/uploadPath/MySQL/home/root历史部署。已清空Snap cache、删除4个disabled Snap revision、执行APT clean并将journal压缩至256MB；根盘从80%（used=`31,839,608,832`、avail=`8,116,183,040` bytes）降至72%（used=`28,536,225,792`、avail=`11,419,566,080`），释放`3,303,383,040` bytes。server SHA=`13d5f07e...`、PID=`2382320`、NRestarts0不变；Go/MySQL active，内外及公网root/health200，最近关键日志0，Chrome/LibreOffice可用。完整证据见[production-disk-governance-20261010.md](production-disk-governance-20261010.md)。

[005 formal slice2 - 2026-10-10] 已新增local-only正式资产只读preview API，仅接受00501/00502及受控assetKey；返回205规则、88 Tag、6图、5数字槽、四项SHA、非nil集合及稳定阻断原因，路径不泄露，读取前/中/后身份与大小复核。RED为缺服务符号/HTTP404；GREEN为formal专项35项、Go全量及server build。该能力零DB写、零批准、零PDF、零远端；客户原始V2.8模板仍未通过正式门禁，formal整体继续P0。完整证据见[management-traits-formal-assets-preview-local-20261010.md](management-traits-formal-assets-preview-local-20261010.md)。

[FB-224～FB-226 / UF-059～UF-061 staging发布 - 2026-10-10] 已单次发布后端SHA=`e221ab4512724e9c7748f4d78c88618930bb3e29fdd64eb99a5c0c66550cffd0`、前端index SHA=`16455aefeec9df8fda97298b32f0cfe90d1fa83860b0f9433e500e81e0c3d86f`（393文件）及客户基准00401 v2模板SHA=`75b79b93e36a3a8dee83abb1d822faa5fae49a500597389c5bc5c63f2c3cd6dc`。认证元数据真实返回00401 60个全可选可重复字段、005 95个可填字段（90必需+5可选），临时Redis会话精确清理。指定00401历史报告经真实API强制重生成，绑定PDF 820207 bytes/SHA=`7150a3a85a987a34a0a371acd5bbbf93f27f1708295c5fe418816b2f74605b3a`、A4 10页，文本无“时长/分钟/页 共”，10页视觉抽检未见破版。最终PID=`25353`、NRestarts=0、内外health200、active paper0、关键日志0、临时残留0；production未修改。部署脚本首次回滚因`ERR` trap递归触发start-limit且Windows ZIP路径分隔导致空dist，已从精确旧后端/旧模板恢复服务并用POSIX tar恢复新前端，随后非递归回滚脚本完成后端/模板发布；该失败和修复已保留。

[FB-226 staging精确重新发布 - 2026-10-10 21:36] 按用户要求对当前已验收三项字节重新发布，不引入新源码或内容变更。写前确认后端/进程、index、模板SHA分别仍为`e221ab45...`/`16455aef...`/`75b79b93...`，三服务active、active paper0。使用POSIX tar建立可实际解包的完整前端回滚包并核验index SHA后，原子重装三项资产并重启一次；最终PID=`26356`、NRestarts=0、内外及公网health200、393文件、关键日志0、临时残留0。认证元数据再次验证00401=60、005=95/必需90、repeatable=true，临时Redis会话EXISTS0；数据库写0，production未修改。

[纠正 - 2026-10-10 21:41 / staging浏览器部署检查] 上述运行时、SHA和认证API收据有效，但“STAGING GREEN”结论不完整。真实浏览器两次加载`http://20.200.136.133/#/`均进入登录页，但验证码图片固定破损；控制台为`ERR_INVALID_URL`。资源证据：活动bundle实际请求`GET /captchaImage`并收到16290-byte `text/html`首页，页面最终设置`data:image/gif;base64,undefined`；同浏览器直接请求`GET /prod-api/captchaImage`返回HTTP200 JSON，含`captchaOnOff=true`、1112字符PNG base64及uuid。由于验证码开启，当前网页登录链为RED；不得以health200或伪造会话API通过替代浏览器登录验收。production未受影响，本轮检查零写。

[FB-227 staging登录恢复 - 2026-10-10 21:48] 根因是隔离发布worktree缺少被忽略的`.env.production`且构建未显式注入API前缀。保持源码逻辑不变，以`VUE_APP_BASE_API=/prod-api`重建并通过616项测试、production build、bundle前缀门禁；仅原子替换前端为index SHA=`b0576cc20270f7d9a70d85da6f6c9562c868417ba511c25a07695009095de983`/393文件，回滚备份=`/opt/talent-assessment/backups/fb227-staging-frontend-20261010214846`。真实浏览器无Token验证码120×40有效；无效旧Token精确请求`/prod-api/getInfo`并在401后回登录页，Token清除、page error0，无avatar TypeError/405/重复提交通知。后端SHA保持`e221ab45...`，PID1535/NRestarts0，服务重启0、DB写0、active paper0；production未修改。

[FB-228 00401批量ZIP恢复 - 2026-10-10 22:03] 用户真实操作仍报paper=`24504b9c-1874-4bbd-af09-f9d0d83abb16`报告未生成。只读证据确认legacy result为v1，但frontline current精确指向completed v2 report/run；单份下载已走current，批量加载器却只按legacy tuple查找。RED后改为批量加载current/report/run并逐项验证绑定与环境批准，无current才走原legacy兼容。专项、Go全量和全build通过；后端原子发布为SHA=`6c878c8874ae7b08acc2640f18dff0742cf1778bbd699e284b38c430acbfe124`，备份=`/opt/talent-assessment/backups/fb228-staging-backend-20261010220324`。真实认证batch-download返回693998-byte ZIP/1份820207-byte PDF，PDF SHA=`99c40ab80a33d74b5742dba11c4259fa82dd70b83141c92993c74b1b3dca143c`；download audit19→20，临时会话/ZIP清0。PID3394/NRestarts0/health200；production未修改。

[FB-229/FB-230页码、总体环图及中心语义 - 2026-10-10] STAGING GREEN。最终活动模板=608305 bytes/SHA=`ef92bfba2026ac2b0609a6b619a41fa29ff8931a9e31569b28fd2f8ace2e24dd`；普通PAGE、inline环图及无边框中心语义层同时兼容Word/LibreOffice。指定report重生成PDF=813847 bytes/SHA=`c6fbd4bb...`、A4 10页；目标实图显示物理2/3页码、环图中心“总体评价 / 60.94 分”、十维图。PID3394/NRestarts0、health200、active0、日志/temp0；production未修改。

[FB-231测评人员页无范围403 - 2026-10-10] STAGING GREEN。真实日志旧版六次unfiltered tester/list403；前端空examId现零请求并提示先选测评，精确examId路径不变。专项8、全前端617及build通过；前端only发布index=`0e8e02aa...`/393，tester chunk与本地同SHA，发布后tester/list403=0。备份=`/opt/talent-assessment/backups/fb231-staging-frontend-20261010232709`；PID3394/NRestarts0、health200、DB写0/restart0。preflight发现活动00401模板已独立漂移为`d4daf935...`，本轮严格保留未覆盖；production未修改。

[UF-059批量下载staging验证 - 2026-10-10] 胜任力结果页原先对尚无completed报告的完整答卷直接请求ZIP，必然收到“报告尚未生成”。现改为对冻结选择逐份`force:false`幂等准备，全部成功后单次下载ZIP；任一准备失败则不请求ZIP。RED为前端611通过/1失败，GREEN为35文件613项及production build；代码已随FB-226候选发布staging，真实批量选择链本轮未另造数据重跑，production未发布。

[UF-060管理结果百分制staging发布 - 2026-10-10] 管理列表/详情原读v1旧整体合计及1–5维度均分，现对completed v2 run使用已持久化百分制overall/3模块/10维/效度，分页排序同步优先v2；无run的generic/legacy回退旧口径，v1报告数据合同不变。专项、affected、Go全量/build、前端35文件614项/build通过；代码已部署staging并参与指定v2报告真实重生成，列表/详情浏览器值本轮未单独复核，production未发布。

[UF-061待发展项Excel短评staging验证 - 2026-10-10] 工作簿“等级评价”G/I/K列已被内容生成器导入`development`，但v2报告DTO误用`dimension`完整文案。现最低两项精确查询同维度/等级的`development`短评并在缺行时失败关闭；最高3/最低2选择、优势项、维度详情及评分不变。聚焦4项、service、Go全量/build通过；代码已部署staging，指定历史v2报告真实重生成和PDF视觉抽检通过，production未发布。

[UF-062客户最新模板评估 - 2026-10-10] 客户原件`最新胜任力报告模板.docx`保持不变（720337 bytes/SHA=`83bb2c89...`）；已另建“剔除时长”候选（648445 bytes/SHA=`e1f4f807...`），仅`word/document.xml`变化，保留一个隐藏`result.userTime`合同，LibreOffice本地转换为12页A4 PDF且提取文本无“时长/分钟”。当前直接替换仍明确NO-GO：对比图位于Word专用`mc:AlternateContent/wpg`组合，页脚含`NUMPAGES`；既有修复器因新版关系结构无法定位对比图而失败。候选仅59/60且缺`validity.text`曾是阻断项，现已被FB-224本地新政策取消：已注册内容控件存在则填充，缺失则忽略，`report.disclaimer`/`validity.text`同样可选。图表、外链、可见占位符、`NUMPAGES`继续严格；组合图祖先结构仍是FB-223门禁缺口。活动本地/远端模板SHA=`52e0020c...`均未替换，未访问数据库或部署。

[FB-224 v2模板内容控件可选 - 2026-10-10] 用户确认仅放宽00401 v2内容控件缺失校验，全部60个已注册字段均可缺失；模板中存在的已知Tag按原value-only逻辑填充/替换，不存在则忽略且不重建，未知Tag继续拒绝。个人信息过滤现从任一现存participant控件定位表，整表无participant控件时直接跳过。12图、图表可写性、外链、可见占位符和`NUMPAGES`门禁不变。RED命中缺`report.disclaimer`；GREEN聚焦测试、handler、Go全量及build通过，Python合同脚本语法通过且客户候选仍因组合图失败。当前仅LOCAL GREEN，未部署远端。

[FB-225客户基准新模板 - 2026-10-10] 已从客户原件派生并发布staging模板`最新胜任力报告模板-修复后-staging候选.docx`，617264 bytes/SHA=`75b79b93e36a3a8dee83abb1d822faa5fae49a500597389c5bc5c63f2c3cd6dc`；不补`validity.text`，保留59个唯一已注册Tag，12图value-only、十维对比图解组、三份页脚PAGE-only、隐藏时长合同。staging真实重生成报告为A4 10页（受真实数据/环境分页影响，不以本地9页作固定合同）、820207 bytes/SHA=`7150a3a8...`，文本无“时长/分钟/页 共”，抽检封面、总体图、十维图和末页正常。production模板未修改。

[FB-226模板语义字段清单与重复Tag - 2026-10-10] 报告模板管理页为00401及00501/00502两张卡新增“模板字段”入口；弹窗可搜索并说明同Tag多处同值。staging认证元数据真实返回00401 60项全部可选/可重复、005 95项（90必需+5可选）全部可重复；临时管理员会话已清理。重复Tag的实际填值行为由本地00401/005构造模板回归验证；本轮未远端上传合成重复Tag模板，避免替换005活动模板。前端已发布staging，production未发布。

staging最后独立终验：PID `22668`、`NRestarts=0`，talent-assessment/nginx/mysql均active，内外health均200，active paper=0，应用fatal/panic/permission错误0，Nginx 5xx=0。

## 2. 产品状态矩阵

| 产品/链路 | 当前状态 | 已验证范围 | 尚未覆盖/限制 |
|---|---|---|---|
| 00401一期胜任力 | 🟢 STAGING GREEN / 🟡 PRODUCTION既有链可用、新报告生成RED | production运行资产、10维90题、v1/v2包及既有历史PDF保持；UF-058既有报告验收仍有效 | production缺`逻辑思维/good/development`正式内容；命中该组合的新报告生成/批量准备失败关闭。待补数据后复验；native download事件及并发/容量未重跑 |
| 003 MBTI | 🟢 STAGING GREEN（API+报告） | 48题保存、ESTJ计分回读、完整版/简版PDF、16+16模板、匿名模板门禁、exact cleanup | 本轮未做参与者浏览器逐步交互和16种人格逐类型报告全矩阵 |
| 00501/00502管理特质 | 🟢 STAGING GREEN / 🟢 PRODUCTION完整TEST链GREEN | production新建唯一候选、140答、提交、13维/4模块、PDF view/download及exact cleanup通过；客户当前将两测评设为进行中 | formal仍不在本次范围 |
| 00101传统测评 | 🟢 STAGING只读冒烟 | Detail及有效XLSX导出 | 未做本轮临时新卷完整交卷回归 |
| 00201/00202历史管理特质 | 🟡 兼容保护有效 | Detail可读；冻结历史和002源数据保持 | 通用原始导出被管理特质legacy guard按设计403；是否需要专属导出属于产品决策，不应放宽guard绕过 |

## 3. 当前有效技术决策

1. 00501=`管理特质测验基层员工新版`，00502=`管理特质测验干部新版`；只对00502覆盖V67/V96，002与冻结历史不改。
2. 00501/00502共用一份DOCX；当前模板SHA=`05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c`。
3. 管理特质TEST运行时只在`REPORT_EFFECTIVE_ENV`与`MNG_TEST_REPORT_ENV`归一化后同为`local`、`staging`或`production`时启用；空值、不一致、`prod`及畸形值关闭。production必须显式使用`production/production`，不新增开关；005导入后以现有`el_exam.state=1`可见但禁用，由客户通过现有状态操作决定启用。
4. MBTI报告转换统一复用`pkg/libreofficepdf`，有90秒context、隔离workspace和失败关闭；禁止把DOCX冒充PDF。
5. 00401/MBTI/005测试必须使用唯一标记、写前基线、finally精确清理及基线前后比较；不可再硬编码“staging必须空库”。
6. 正式备份、客户材料、业务PDF、审计证据和`scripts/test/results`不是垃圾文件；清理前必须先做所有权、引用、大小、年龄和回滚价值检查。

## 4. 已关闭的矛盾

| 历史矛盾 | 当前结论 |
|---|---|
| 005题库/模板仍为LOCAL GREEN | 已被2026-10-09真实staging发布、浏览器/API和数据核验覆盖，现为STAGING GREEN（范围限定） |
| 当前工作树management-traits发布为PARTIAL | 该条是发布当时阶段结论；后续真实认证浏览器、模板、报告及完整主链验收已补齐，当前按本文件矩阵读取 |
| 00401 staging必须为空才能跑E2E | 错误；当前脚本接受非空基线并要求清理后精确恢复 |
| MBTI直接执行LibreOffice且失败可返回DOCX | 已由FB-218修正为共享有界转换、失败关闭 |
| “完整staging GREEN”等于全系统/production GREEN | 错误；production后来经独立评估、恢复副本演练、完整备份及受控发布单独批准并验收，不能由staging结论自动推出 |
| 409条孤儿paper必须删除才算治理完成 | 用户已批准保留历史例外；已完成全表动态引用闭包、备份恢复可能性核验和固定指纹漂移门禁。active必须保持0，新增或漂移重新打开P1 |

## 5. 仍需处理的遗漏事项

### P0 — 外部批准/上线门禁

1. **00401 production待发展正式内容缺口（UF-068）**：当前缺少`competency-phase1-content-v2/frontline_employee/development/competency-logical-reasoning/good`，导致新报告生成显示“新版结果服务暂不可用”。先补production内容完整性RED和精确备份，再导入已批准短评并验证单份生成、批量准备/ZIP及PDF；禁止用空文案或跨等级回退绕过。
2. **管理特质正式报告/formal assembly**：slice1版本登记和slice2资产preview仅LOCAL GREEN；production当前只发布带TEST标注的005与reissue基线，formal registry/approval未安装，正式候选未批准，DTO/storage/PDF/UI未实现。不得把preview通过、TEST或reissue称正式报告。
3. **生产管理员凭据轮换**：当前admin仍匹配仓库历史默认候选；需用户明确授权后单独轮换并验证登录/回滚。
4. **生产TLS/HTTPS**：当前登录页仅HTTP、无443且浏览器非secure context；须配置受信证书、HTTPS及HTTP重定向后重验登录/下载/Cookie。
5. **秘密配置与root运行**：application配置含秘密标记但为0644；有效service以root运行且hardening评分9.6 UNSAFE。先在staging设计0600秘密注入、专用用户写目录与systemd沙箱，再单次生产迁移。
6. **匿名部门考试授权边界**：`OnlinePaging()`为匿名路由且当前返回`open_type=2`部门限定考试；必须先写匿名/同部门/跨部门RED测试，再改为匿名仅公开、部门型按认证用户关系过滤。
7. **生产CORS来源边界**：当前任意Origin通过且同时允许credentials；按local/staging/production配置精确allowlist并补跨站认证读写测试。

### P1 — 可在后续独立测试中关闭

1. **005全新staging生命周期**：新建唯一标记draft→人员→freeze→140答→提交→结果→报告→exact cleanup；当前主要依赖保留基线验证。
2. **真实浏览器用户链**：00401参与者入口、刷新恢复、90答、提交、管理员结果/详情和浏览器PDF响应已关闭；仍缺真正浏览器native download事件。MBTI仍需补参与者浏览器链；005历史native事件盲区也应统一复测。
3. **MBTI类型矩阵**：至少为16种类型各验证模板可加载和DOCX/PDF转换；当前完整链只得到ESTJ，模板只验证存在性16+16。
4. **001/002兼容回归**：若下一发行会改公共paper/exam/report代码，应重跑临时001/002/003新卷完整链，而非只读Detail。
5. **最新FB-218真实失败注入复验**：同步和异步失败关闭补丁已有RED/GREEN、Go全量并随production后端SHA上线；production未故意触发LibreOffice故障，仍缺远端真实失败注入证据。
6. **传统题库身份差异决策**：先确认00301/00302在production的产品身份和是否需要双库并存，再逐字段比较00101/00102/00501/00502 repo元数据及3条005 definition bundle；未完成引用闭包和客户决策前禁止以staging全量覆盖production。
7. **生产网络面收紧**：UFW对全网允许3306/10301/8088/9001/39000–40000且root SSH开启；云侧当前拦截多数端口，但应核实所有者后同步收紧UFW/NSG，并让MySQL和Go 8092只绑定内网或localhost。
9. **磁盘长期保留策略**：2026-10-10已通过系统缓存清理将根盘80%降至72%，当前低于建议75% warning但仍高于长期目标70%。业务backups/uploadPath未清理；后续先建立DB文件引用图、恢复价值与保留周期，再逐项批准。孤儿试卷逻辑删除不会自动缩小InnoDB表空间，不能作为本项解决方案。
10. **CI质量门禁可信化**：CI/PR gate从Go 1.24对齐`go.mod` 1.26；取消前端依赖、测试、覆盖率、lint和gosec的无条件`continue-on-error`，至少P0/P1安全问题必须阻断。
11. **后端正确性与性能**：统一路径边界校验、分页上限、GORM Error/RowsAffected处理；消除Candidate/Tester列表和试卷答案装载N+1；JWT占位密钥在非local环境启动时fail-closed。
12. **参与者失败恢复**：MBTI与00401补加载/保存/提交错误状态和重试；通用DataTable使用catch/finally；答题页不再依赖未知浏览器历史回退。
13. **并发与race证据**：在CGO=1环境执行`go test -race ./...`，并以真实MySQL双连接验证submit/expiry、reissue、报告生成/下载和模板替换竞争。
14. **00401客户最新模板staging发布与上传门禁补强**：客户基准新模板已LOCAL GREEN，待staging SSH恢复后执行零写预检，单次发布Linux后端SHA=`b30e08de...`与模板SHA=`75b79b93...`并生成真实报告验收；当前未发布。另需为`validatePhase1V2WordTemplateUpload()`增加“`chart.dimension.comparison`不得位于`mc:AlternateContent/wpg`组合”门禁和RED/GREEN回归，防止未来上传未解组模板。当前严格格式脚本对活动SHA=`52e0020c...`在FB-194旧视觉断言出现`StopIteration`，应拆分运行兼容与版本化视觉门禁。

### P2 — 测试与治理改进

1. 2026-10-10本地快照：Go statements 45.7%（历史44.3%，+1.4个百分点）；Vue statements/branches/functions/lines=`60.25/95.53/39.13/60.25`。Go handler32.7%、config3.0%、repository/db/redis0%、pdfgen1.6%为优先盲区；后续建立按包漂移和新增代码覆盖门禁。
2. `project-memory.md`为697,706 bytes、2,254行，继续作为时间线；后续先读本文件，只在需要历史证据时查时间线，避免旧`PARTIAL/BLOCKED`误当当前状态。
3. staging `/opt/talent-assessment/tmp/uploadPath`含历史业务/报告目录，本轮未删除；需另做文件与数据库引用图后才能提出清理清单。
4. `/opt/talent-assessment/backups`历史备份不在本轮清理授权内；如需释放空间，应先输出年龄/大小/关联发行/恢复演练价值，再由用户逐项批准。
5. `scripts/test/results`含发布和审计证据，不应按普通测试垃圾整体清空；可另制定保留周期和索引。

## 6. 后续任务读取规则

1. 先读本文件确定当前状态和待办。
2. 再读 [project-memory.md](project-memory.md) 顶部最新记录；只有调查历史根因时才读取对应日期段。
3. 功能分支以 [business-branches.md](business-branches.md) 为准；修复回归以 [regression-tests.md](regression-tests.md) 为准；业务链以 [business-chains.md](business-chains.md) 为准。
4. 任一新验证必须写明：环境、源码/产物SHA、是否真实DB/浏览器、创建数据、cleanup结果、未覆盖项。
5. 旧记录中的`PASS/GREEN/已部署`只对其原始日期和明确范围有效，不自动继承到当前版本。
