# 002 Word候选：本地结构完成，视觉验收未通过

## 2026-10-02 当前兼容切片：原稿内容保真＋独立LibreOffice兼容版

用户在严格原稿模式显示缺陷说明后，明确选择修复LibreOffice显示兼容，允许数字框、图例、分页必要局部调整。**原件/字体字号/配色/固定正文/星形/页脚仍保留**，不恢复删星/PAGE-only决定；兼容版独立生成，严格原稿版及历史优化版不覆盖。未授权活动安装、运行接入或部署。

- [兼容程序模板候选](generated/management-traits-word-candidate-20261001/management-traits-002-lo-compatible-template.docx)、[绑定manifest](generated/management-traits-word-candidate-20261001/lo-compatible-manifest.json)、[合成演示PDF](generated/management-traits-word-candidate-20261001/lo-compatible-synthetic-demo.pdf)、[验证证据](generated/management-traits-word-candidate-20261001/lo-compatible-validation.json)。新显式`--libreoffice-compatible`解耦原稿文本绑定与局部版式；原例59.45、原rPr、星形及页脚保持，填充演示才替换全部绑定值。
- 缺模式RED为10项1fail/9errors；严格原稿实际PDF7项57失败子项。兼容版完整GREEN为17项0失败/错误/跳过，旧严格10项和优化23项回归通过。主代理独立结构10项通过，独立审阅scope gate PASS；完整实际PDF执行须带`--visual-boundaries`，仅模式开关不能冒充渲染验收。
- 主代理随后独立执行兼容模式加`--visual-boundaries`：64.233秒，`CONTRACT_TESTS=17 FAILED=0 ERRORS=0 SKIPPED=0`，退出0；几何/像素、图例、完整长文案及详情分页均实际执行，不是只读取既有证据。
- 五组0/25/75/100/28.85各5/5数字、25几何及25像素避碰、完整图例、36长文案、每稿13详情和13建议同页通过本机门禁。88SDT、5数字槽、六图保持，旧62个根目录产物SHA不变。
- 新模板SHA=`a986ba0f3c5985486cf5f2a781346beb4140d2f2e79952a25c3beb57a273eb46`，演示SHA=`09ed491a2665c2145f2771312066201fb2b9c2c7252967b4aecc7989bb456fa0`；原稿c82c2dc0…不变。Word只读10/11页88控件、关闭SHA不变；LO26.2.5.2为9/9页，不宣称跨引擎同页。
- **未验收：**概览下部与第7页留白仍有；原星形与总页数页脚仍按客户稿保留，含义/跨引擎页脚需最终审阅。目标服务器、Word PDF、正式内容及运行链未验，不称已激活。staging候选测试批准范围不变。

## 2026-10-02 最新决定：客户原稿版式优先及程序模板派生

用户再次指定客户原稿为权威来源，并明确选择“客户原稿版式优先”：保留原布局、固定文字、星形和页脚，只加入动态绑定、清外链与必要转换兼容。**覆盖下方相冲突的删星/PAGE-only及标签、图例、分页优化决定**；历史优化文件与验证记录保留，不作为本次选定模板。字段post/submittedAt及13维平铺等不冲突的语义决定保留；staging候选测试边界不变，不代表正式批准或部署授权。

### 输出及验证

- [程序绑定派生模板](generated/management-traits-word-candidate-20261001/management-traits-002-source-layout-template.docx)、[专属绑定manifest](generated/management-traits-word-candidate-20261001/source-layout-manifest.json)、[合成填充演示](generated/management-traits-word-candidate-20261001/source-layout-synthetic-demo.docx)。均为独立本地派生文件，未安装活动目录或接入报告程序；演示不是正式测量结果。
- 构建器新增显式`--source-layout`模式，默认旧优化模式不变。原稿模式不扩框、不外移标签、不改legend/pPr分页、不强制摘要粗体样式，保留原例数字59.45；88 SDT＋5数字槽＋六图字面量绑定及原坐标转换，原固定文字/样式/星形/页脚/媒体受完整白名单验证。
- 主代理独立测试：**10 tests / 0 failures / 0 errors / 0 skips，退出0**；旧优化完整23项回归亦通过，旧五项产物哈希保持。独立审阅PASS，仅本地派生模式。
- 工作区原稿重新实测660351字节，SHA=`c82c2dc0cdcaec413866268c557561d5c53abc41ec02d94dc6aea615edca2f48`不变。附件未提供完整新落盘文件，不能声称新附件已逐字节比对；本次使用已核验的工作区同名原件。
- 新模板SHA=`465358b1ad88bb9442029c2c03609ec1816c64e2d8e9b3021019859a521ff471`；演示SHA=`efa4c253b4e0727c2a15cad1ad0ba91fb4f10945f173087e6caf4735906f0f82`。Word原稿/派生/演示只读9/9/10页，关闭SHA不变；本机LO派生8页、演示9页。

### 不能宣称通过的项目

原框数字拆行、人际裁切、图例截短与演示尾字孤页在本机PDF复现；本模式依据客户原稿优先保留，**不是完整视觉验收通过**。原稿示例分数冲突继续仅作位置样例，生成结果必须全部按同一事实源替换。必要兼容拆组仍须目标服务器/客户审阅；原版一致性合同不等于跨引擎像素一致或正式可用。数据库、Go/Vue、运行配置、活动模板、部署及正式启用均未改变。

## 2026-10-02用户决策

本表记录本轮问答的明确选择，参照[设计§8](management-traits-word-design-20261001.md#L179-L224)区分字段语义、呈现选择与实现状态。**对象范围：仅staging候选测试使用，未决内容不视为正式批准。**负责人和日期的确认不等于正式内容及心理测量双批准，不构成production选择、模板激活、运行接入或部署授权。

| 事项 | 用户明确决策 | 状态及批准边界 |
|---|---|---|
| 产品正式名 | 管理特质 | 名称已确认；本切片仅记录，未改候选或系统名称 |
| 职务 / 测评日期（postdate） | 职务使用岗位`post`；测评日期使用提交时间`submittedAt` | 仅确认字段语义，不代表运行绑定、采集或显示逻辑已实现 |
| 新字段位置 | 暂不新增年龄、用时、总体常模位置 | 不新增位置，不将这些候选目录项强制设为必需字段 |
| 13维图分类（flat） | 接受13维平铺，不恢复四模块层级 | 仅确认展示语义；不改变四模块评分聚合，也未修改运行实现 |
| 等级呈现（star） | 删除星形，仅显示中文等级 | 决策已记录；星形删除尚未在候选执行 |
| PDF主要排版验收与页数 | 以目标服务器PDF作为主要排版验收依据，暂不固定页数 | 验收口径已确认；本轮未访问目标服务器或执行PDF验收，不宣称Word与服务器同页 |
| 页脚（pageonly） | 仅显示当前页码，不显示总页数 | 决策已记录；仅当前页码尚未在候选执行 |
| 内容负责人 / 心理测量负责人 | 两者均为Liming | 仅确认负责人，不称已正式双批准 |
| 确认日期 | 用户先给`20291001`，经明确核对选择`2026-10-01` | 记录核对后的日期；本决策记录日期为2026-10-02，不伪造正式批准时间 |
| 对象范围 | 仅staging候选测试使用，未决内容不视为正式批准 | 不选择production，不授权部署、迁移、运行接入或正式报告启用 |
| 未决文字及题本差异 | 错字、频率、统计声明、V67/V96不擅自修订；当前questionnaires差异不改DB | 保留现状，不以本轮选择视为内容批准、文本等价或数据库修正授权 |

### 未执行清单与历史覆盖规则

- **star删除、pageonly尚未在候选执行**；本轮未修改DOCX、PDF、JSON、构建器、测试脚本或活动模板，不宣称实际模板已变动。
- **postdate与flat确认仅语义，不运行实现**；未接入字段读取、绑定、采集、图表渲染或报告链。
- 目标服务器PDF主要排版验收尚未执行；暂不固定页数不是排版已验收，也不是激活许可。
- 未决内容、错字/频率/统计声明及V67/V96继续保留；当前questionnaires差异不改DB，未执行数据库查询、修改、迁移、SSH或部署。
- 下方旧pending及历史“待确认/未批准”记录原样保留；**仅本表已明确决策的同事项由本表覆盖其待确认状态**，不覆盖历史测试证据、不将未执行变更记为完成，也不解除未决内容、正式双批准、运行接入或启用门禁。

日期：2026-10-01。范围为用户同意的Word候选切片；无数据库、运行入口、活动模板、客户原件或部署变更。

## 2026-10-02 后续更新：标签避碰与模块分页本地通过

**仍未批准：**最终版式/跨引擎页数、星级/页脚/分类层级及字段语义、正式内容、目标服务器与运行接入。概览下部、第7页两项完整详情后仍有留白，未全局压缩。本轮消除的是标签压环、摘要独占页和详情拆分，不代表所有页面密度最优或可激活。下方前次“仍存在”仅为历史状态。

用户明确选择继续优化本地候选。仅调整五个标签位置（左模块左移96pt、右模块右移100pt、总体下移108pt）；原环图/组合图位置、字体/字号/颜色/数据不变。13详情表按模块加cantSplit/keepLines，前两行keepNext、末行明确终止；建议标题/完整段落局部粘连，不串联全部维度。取消摘要后详细分析的强制分页，使摘要与详情共享物理第4页。

- 精确重建前次candidate `8b929da1…` / demo `cbdcf73c…`并核SHA后，新增视觉RED：4顶层、48失败子项、0 errors，退出1。一次版式实现后主代理完整复跑 **23 tests / 0 failures / 0 errors / 0 skips，退出0**；独立有界审阅PASS。
- 五组0/25/75/100/28.85仍各5/5完整数字；实际SVG五环包络与25标签框交叠0，144DPI标签框内环色像素0。图例完整、36长文案完整；候选和演示各13详情块/13完整建议同页，摘要不再独占稀疏页。
- Word只读候选10页/演示11页、88控件，关闭不保存SHA不变；LO26.2.5.2均9页。Word PDF及段落分页未验证，不宣称跨引擎一致。
- 当前候选DOCX SHA=`1f236407713143d0e945d60391e9179bc761c62f0a42e750b50075c98f029273`，演示DOCX SHA=`be2ec7119fd5a5589a05d4dd6c0d5ddb0e026cb913873321e8b97775faa88154`，原件SHA不变。前次所有SHA及页数保留为历史。
- [本切片详细证据](generated/management-traits-word-candidate-20261001/layout-slice-validation-20261002.json)、[当前演示全页](generated/management-traits-word-candidate-20261001/demo-review-contactsheet.png)。未改DB、Go/Vue、配置、活动模板或部署；脚本测试通过不代替运行链验证。

## 2026-10-02 更新：局部显示门禁通过，整体候选仍待审阅

**未关闭：**环图文字与环体相交的原中心布局、摘要页稀疏、责任心详情跨页仍存在；最终版式/页数、字段语义、固定星级及页脚、内容批准、目标服务器和运行链均未验收。本轮通过不等于候选已可激活。下文10月1日失败保留为历史；当前MT-WORD-01仅所列数字裁切/图例/尾字孤页已通过本地验证。

用户明确同意仅调整标签框宽高/内边距、图例空间和分页粘连。构建器保持原中心，五数字框宽120点、高总体64/模块44点、四内边距为0；比较图仅改legend manualLayout；总体诊断keepLines及两摘要栏局部keepLines/keepNext。不改字体/字号/配色、图形类型、系列文字/数据、客户原件、Go/Vue或数据库。白名单测试保留原数据/样式/其他段落属性约束。

- 扩展RED：17顶层、10 failures、0 errors；最终主代理与独立审阅复跑：**18 tests / 0 failures / 0 errors / 0 skips，退出0**。独立审阅PASS。
- 实际0/25/75/100/28.85五组PDF均完整数字5/5，未扩大原数字计数区域或降低计数；图例完整且不与100标签框相交；36条动态长文案完整，尾字孤页消除。第4页现为完整摘要表但仍稀疏。
- Word只读候选9页/演示10页、88控件，关闭不保存SHA不变。LO26.2.5.2候选8页/演示9页；不承诺跨引擎同页或最终页数。
- 当前候选DOCX SHA=`8b929da1fd8e78068240dc57c7ea6df14518ad3b24c9b0b54af5f6b0299fe1ba`，演示DOCX SHA=`cbdcf73c04279d95eb1f8b365cd6be236f431c6497d00fa69f6ecd015a65b2df`。原稿及工作簿SHA仍与下文来源相同；下文旧候选SHA只作历史。
- [本轮生成/验证记录](generated/management-traits-word-candidate-20261001/final-validation-20261002.json)、[演示全页实图](generated/management-traits-word-candidate-20261001/demo-review-contactsheet.png)、[候选全页实图](generated/management-traits-word-candidate-20261001/candidate-review-contactsheet.png)。原图链接已更新为当前产物，不作为旧失败画面证据。

## 1. 阻断与未验证项（优先阅读）

**结论：PARTIAL / NOT ACCEPTED。** 不能交付或激活当前候选。

- 实际PDF总体28.85拆行为“2”及“8.85”，人际模块数字底部裁切；组合图图例截短且与100分标签重叠。
- 长文案演示第4页仅上一页尾字“人。”，分页验收未通过。
- 实际0/25/75/100四组PDF测试均只能找到4个完整环图分值token，要求5个；门禁保持RED，未删断言或伪造微小分数掩盖0分。
- 同一文件视觉修复已进行三轮，按规则停止继续试改，等待用户选择后续处理方式。
- Word和本机LibreOffice页数不同；页数目标未批准，不能仅凭页数判定内容完整。目标服务器、正式内容批准、全页视觉、性能、活动模板及真实报告链均未验收。

## 2. 本地产物与已验证能力

- [构建器](../scripts/tools/build-management-traits-002-word-candidate.py)：仅处理确定SHA的客户原稿，输出限于本地候选目录，拒绝原件作为目标；未调用一期硬编码修复入口。
- [契约与真实PDF边界测试](../scripts/test/management-traits-002-word-candidate-contract-test.py)：原稿先RED，候选结构GREEN，真实渲染仍RED。
- [候选DOCX](generated/management-traits-word-candidate-20261001/management-traits-002-candidate.docx)、[位置manifest](generated/management-traits-word-candidate-20261001/field-position-manifest.json)、[synthetic演示DOCX](generated/management-traits-word-candidate-20261001/management-traits-002-synthetic-demo.docx)、[演示PDF](generated/management-traits-word-candidate-20261001/management-traits-002-synthetic-demo.pdf)。仅用于审阅，演示不是正式结果或评分金标准。

88个实际SDT：人员/日期6、总体等级/诊断2、13维五字段65、顶底摘要12、原位置总体三段建议3。另有总体及四模块5个逐对象数字槽，位置路径在manifest中；不把94候选注册键当必需控件数。

六图从分组graphicFrame提升为普通独立anchor，按来源矩阵定位；五环、比较图及固定五档背景恢复同页。五环各1系列×2点，比较图2系列×13点。图表引用物化为字面量，全包外链/公式/externalData/numRef/strRef/multiLvlStrRef为0，20表及原12媒体部件保留。比较图两层分类改成13个叶子literal，原层次保存在manifest，**层级显示未宣称原样保留**。

未新增年龄、时长、总体常模位置或模块等级；职务→post、测评日期→submittedAt仅候选语义。固定星级、页脚总页数、固定定义/统计文字不擅自纠正。动态长文案来自客户工作簿原文，冲突文字原样保留。

## 3. 可重复证据

- 原稿结构RED：3项失败，退出1（无控件、分组图、外链/公式）。
- 最终主代理原生Python验证：14项结构/绑定/来源/确定性/样式及媒体/边界数据测试通过；加入实际PDF后共15个顶层测试，4个边界子项失败、0 errors、0 skips；`CONTRACT_TESTS=15 FAILED=4 ERRORS=0 SKIPPED=0`、`VISUAL_CONTRACT_EXIT=1`。
- 真实PDF：每个score=0/25/75/100均`full_numeric_tokens=4 expected=5`。不是未执行或环境跳过。
- 两脚本编辑器诊断0；语法编译通过。未修改Go/Vue，因此本轮未重跑其全量编译/测试。
- 最终Word只读打开：候选9页、长文案演示10页，关闭不保存SHA不变；LibreOffice26.2.5.2分别8页/9页。初版曾为Word10/11及LO9/9，已被最终产物替代，不能混用。
- 最终候选SHA：`76b58e95b8e5acf58335768efaecca9cc10202c7683ac9da1056877d42cec18b`。
- 最终演示DOCX SHA：`6685578ec61b0d6fa2775f461f4132e0dfc726da5f309f708cf48e11e9cb3d0d`。
- 客户原稿SHA保持：`c82c2dc0cdcaec413866268c557561d5c53abc41ec02d94dc6aea615edca2f48`。工作簿SHA保持`b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c`。

实图：[概览页裁切/换行](generated/management-traits-word-candidate-20261001/demo-review-3.png)、[尾字孤页](generated/management-traits-word-candidate-20261001/demo-review-4.png)。

## 4. 下一步须确认

建议允许**只对候选**调整数字标签框宽高/内边距、组合图图例空间，以及概览段落分页粘连，保留字体/配色/业务数据/原件不变；再以当前RED真实PDF门禁验证。是否允许该局部版式调整由用户确认，不继续无限尝试严格沿用原框坐标。所有运行接入及部署继续关闭。