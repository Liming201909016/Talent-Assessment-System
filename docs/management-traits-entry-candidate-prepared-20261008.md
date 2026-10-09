# 旧管理入口单组件候选 — 2026-10-08

## 状态与未完成项

**PREPARED FOR INDEPENDENT REVIEW，非发布批准。** 单入口隔离源码与生产模式资源包已经生成；完整编译工厂语义/依赖闭包证明没有完成，独立 CodeReviewer 尚未执行，真实浏览器自动分流未验。本轮不部署；旧 URL 的线上永久反馈仍开放。原 staging 部署批准保留，统一 reissue 的真实 MySQL 门禁不受本包改变。

### 返回的工件

- [单入口源码](../Go-based%20Refactored%20System/bin/mng-admin-entry-20261008/candidate-ui/src/views/user/exam/index.vue#L259-L379)。
- [原始字节 source_patch.diff](../scripts/test/results/mng-admin-entry-20261008/review/source_patch.diff)。
- [artifact-manifest.json：393资源与完整SHA](../scripts/test/results/mng-admin-entry-20261008/review/artifact-manifest.json)。
- [候选资源归档](../Go-based%20Refactored%20System/bin/mng-admin-entry-20261008/candidate-dist.tar.gz)。
- [406旧输入与已上线资源清单](../scripts/test/results/mng-admin-entry-20261008/sourcehash-input.json)、[仅单SFC变化](../scripts/test/results/mng-admin-entry-20261008/sourcehash-after.json)。

| 对象 | SHA-256 / 实证 |
|---|---|
| 旧线上首页 | 52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860 |
| 新入口SFC | 47e0fcc5dc7e04feb01032d9f705d15c5a8467f296e53c07569b4249288875e8 |
| 候选首页 | bad70356750d846704a06d3515566833415b6bb3f3d22c2fc6bfa9ca6a6255bc |
| 候选tar.gz | c59161887fb00f31a11c0651c7f6e259b5660386a2a7338f77e9f61a38493084；7,631,433 bytes |
| 本次测试 | 86 pass / 0 fail / 0 skip，native exit 0 |
| 本次生产模式构建 | native exit 0，Build complete，393文件 |

## 精确基准与范围

原已上线 identity candidate-ui 保持只读。其 dist-faithful 的全部393文件逐大小/SHA匹配原发行清单；本轮开始与08:33:28Z最终正常公网GET均HTTP200/16155bytes/52eec完整SHA，未探测39生产。

406输入来自已发布清单：394 src、6 public、6配置；相对更早输入原已批准candidate身份切片仍是唯一旧变化。本次候选在此基准上**只改变**管理入口SFC。旧源码215 Go逐SHA再次匹配恢复清单；当前工作树 frontend src/public、internal Go与原candidate-ui在本轮前后快照完全相同。sourcehash-input 中 goBaseline 字段误用了前端文件数量，**不能作为Go计数**；准确215核验在 artifact-manifest 的 baselineGoFilesVerified。

模板文本与全部style内容/属性不变；source位置偏移来自script增长，不算样式变化。生产配置、锁文件、public、字体、品牌、旧模板输入不改；没有 npm install/update、配置切换或后端构建。产物由旧输入的独立cwd构建；不冒称原webpack绝对上下文相同或所有hash路径不变。

### 影响清单

| 消费方 | 处理 |
|---|---|
| 隔离入口 created/retry/watch/旧列表callback | 同步修改；可信server分流及当前scope/销毁屏障 |
| 正常 getInfo | 无需修改；复用旧login API的真实GET返回code200/user/permissions |
| 旧管理特质API | 无需修改；a90b1f67fa834fdefd4bc11d0eea91dcdc86fa965bb9098cb43a7b55492f9d07保持 |
| 旧TEST结果组件 | 无需修改；eb97854254e938735693c8e31a299a0ee919d0ff42b46c62eba2d4d537c96028保持，仍原手填report ID/旧TEST动作 |
| router与已上线candidate身份组件 | 无需修改；逐SHA保持原baseline，跳转params合同不变 |
| 旧tester开放/封闭列表 | 无需修改API；仍原参数/两类调用，仅未确认或迟到时拒绝加载/回写 |
| 新reissue/draft/formal/后端/DDL | 不包含；原双连接门禁、原批准与未部署状态保持 |

## 本次逻辑

只依据Detail合法当前ID、完整legacy两轴、精确同一00201/00202题库与**严格boolean**决定新旧，不使用query/meta/session/缓存role当授权或唯一分类。repoCode缺失可读取实际repoList.repoCode/code；两alias或多个题库不一致拒绝，不把undefined/null默认legacy。strictfalse真实历史合同走旧链，无profile探测；stricttrue核冻结字段白名单、fresh正常getInfo正安全数值userId=1或正ID且wildcard，再读原profile/detail同exam/nonempty frozenAt，进入已有ManagementTraitsResults。未知/失败拒旧且沿原error/retry UI；权限文案固定，不回显SQL/token/rawerror。

001/003完整legacy分类保留旧列表且不新增权限探测；00401完整competency/scoring分类保留专属跳转。没有新增public函数/API、props、报告组件、新filters或客户报告功能。

## TDD与工具失败保留

[实际隔离SFC测试](../scripts/test/mng-admin-entry-isolated-20261008.js)只加载此入口，使用原已有API导出；没有新reissue mocks/exports，不执行当前工作树167/560测试充当隔离证据。

- [有效RED](../scripts/test/results/mng-admin-entry-20261008/red-tests.json)：13pass/73fail/native1，queryless frozen原入口回旧list。
- 首[GREEN尝试](../scripts/test/results/mng-admin-entry-20261008/green-tests.json)：85pass/1fail/native1；唯一为Vue/JS realm空数组prototype比较，实际空值相同。改用Array.from后仍严格断言为空；为实际Element标签补测试stub，未抑制Vue警告或改产品。
- [最终GREEN](../scripts/test/results/mng-admin-entry-20261008/green-final-tests.json)：86pass/0fail/0skip/native0，模板实际编译；metadata/别名/身份/权限/重试、开放封闭、三阶段跨exam/销毁迟到均覆盖。仅合成API，不是真实服务器授权验收。
- [构建](../scripts/test/results/mng-admin-entry-20261008/build.json)：Node16.20.2，Webpack4.47.0/VueLoader15.11.1/VueTemplateCompiler2.6.12/BabelLoader8.4.1/VueCLI4.4.6/Terser4.8.1/TerserPlugin2.3.8，与旧工具版本相同；环境相关键清单为空、生产配置/lock SHA同；原2体积warning保留，无error。
- AST证明器首错误把合并chunk别名当唯一，第二错误把同名嵌套Vuex commit当require，均工具错误；修为重复实例保留及词法绑定后第三轮超过120秒仍无mapping/factory结果，手动Ctrl-C停止。未记录native完成退出码，**不是PASS或已确认timeout根因**。该文件已达三编辑，不继续修/复制绕预算。
- 准备helper review-artifacts 模式因未安装diff模块、Git换行stderr先后失败，三编辑停止，不可把此mode称可重放GREEN。最终按用户指定工件直接生成，使用命令局部core.autocrlf=false比较原始字节（不修改共享Git配置），原始diff退出1=存在差异；打包/完整SHA核验native0。另一次样式比较包含源码start/end偏移误判，失败保留，按真实content/attrs比较后通过；无产品修正。

## 已关闭与仍欠的编译证据

[被动运行加载器库存](../scripts/test/results/mng-admin-entry-20261008/runtime-resource-inventory.json)使用真实候选inline runtime，117动态chunk生成172个JS/CSS请求（117/55），路径与SHA全部存在，179个gzip解压与原文件匹配，1589 factory/2020重复实例保留，publicPath=/；没有网络请求/业务factory执行。**这只是资源存在性，不能代替完整工厂AST、context targets、依赖边与按计划加载闭包证明，也不是Chrome业务验收。**

独立CodeReviewer由主协调者接续：审阅本单SFC与diff，再对这份锁定SHA候选完成完整编译闭包证明；当前模式不派发代理、不自审冒独立PASS。此次只可复审、不release ready；后续全reissue管理UI仍待原双MySQL门禁，不将旧TEST页面描述成新客户报告功能已上线。

本轮SSH/SQL/主数据写/DDL/上传/部署/restart/真人身份答案或PDF/39生产操作均0；公开GET产生正常访问日志，不声称系统绝对零写。原用户浏览器保持既有可读结果页，不操作其会话。