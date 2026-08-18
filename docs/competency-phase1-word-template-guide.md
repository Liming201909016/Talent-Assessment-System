# 00401 一期 Word 报告模板维护说明

## 1. 文件与渲染链

- 客户源文件：`docs/260807胜任力开发资料/胜任力测评报告样例.docx`
- 运行模板：`Go-based Refactored System/configs/export-templates/competency-phase1-report.docx`
- 模板生成器：`scripts/tools/build-competency-phase1-word-template.py`
- 默认PDF转换：服务器LibreOffice（与MBTI报告相同模式）
- 可选PDF转换：Microsoft Graph（Microsoft 365 Word转换引擎）
- 故障回退：可配置回退到原Vue/Chromium报告

运行模板由生成器确定性生成。当前staging模板SHA-256为`3b6a83fd4a2fddf7c0a47c1eda5e2e4141b7d0d72fd9431980928be598e86b92`。

## 2. 客户可修改内容

客户可以直接用Microsoft Word调整：

- 字体、字号、颜色、行距和段落间距；
- 页边距、分页符、表格尺寸和单元格样式；
- 封面、背景、装饰图形和图片；
- 12个原生Word图表的颜色、线型、图例和位置；
- 固定标题、阅读说明、维度定义、等级说明等固定内容；
- 各动态字段所在位置。

客户修改后，应将DOCX交回并提供版本号。部署前由程序执行内容控件、图表、页数和PDF验证。

管理员可进入“测评管理 → 报告模板”，在页面顶部的“00401 一期胜任力报告模板”卡片中：

- 查看当前文件名、大小、修改时间、SHA-256和模板契约状态；
- 下载当前生效模板；
- 选择DOCX并点击“上传并生效”。系统严格校验后自动备份旧模板并替换，校验失败不会影响当前生效模板。

## 3. 动态字段内容控件

运行模板包含49个Microsoft Word纯文本内容控件。页面显示“张三”“30”“合格胜任”“此处显示诊断与发展建议”等正常示例值，不再显示内部双花括号字段。

稳定字段键保存在内容控件的隐藏`Tag`元数据中，例如：

- `report.date`
- `participant.name`
- `overall.level`
- `overall.diagnosis`
- `validity.notice`
- `group.general_ability.score`
- `dimension.<维度ID>.score`
- `dimension.<维度ID>.level`
- `dimension.<维度ID>.diagnosis`

维护要求：

1. 可以直接修改内容控件中文字的字体、字号、颜色和段落位置；
2. 不要使用“删除内容控件”命令，不要修改内容控件属性中的`Tag`；
3. 可以选中整个内容控件并移动到其他段落或表格单元格；
4. 维度定义是客户维护的固定Word正文，不由程序覆盖；
5. 保存格式必须为DOCX，不接受DOC、PDF或WPS私有格式。

如需核对Tag：在Microsoft Word启用“开发工具”选项卡，选中字段后打开“属性”。普通编辑不需要显示Tag。

## 4. 固定图片和图形调整

- 打开“开始 → 选择 → 选择窗格”，从对象列表选中封面图片、背景图形或装饰线；
- 优先选择“固定在页面上的位置”，并设置明确的文字环绕方式，避免随正文漂移；
- 可以替换图片或调整大小、透明度和层级，动态字段和图表不依赖图片文件名；
- 保存DOCX后必须在staging重新生成PDF验收，不能只依据Word编辑视图判断分页。

## 5. 原生图表映射

| Word图表 | 动态数据 |
|---|---|
| chart1 | 通用能力、心理素养两个一级得分 |
| chart2 | 十个二级维度雷达图得分 |
| chart3–chart12 | 十个维度的“当前得分 / 距5分差值”环形图 |

客户可以修改图表样式，但不能删除数据系列或改变数据点数量：

- chart1：2个数值；
- chart2：10个数值；
- chart3–chart12：每个2个数值。

### 内嵌Excel候选模板

已生成候选文件`configs/export-templates/competency-phase1-report-embedded.docx`。它在当前模板版式基础上：

- 内嵌一个`word/embeddings/competency-phase1-chart-data.xlsx`；
- 12个图表全部改为内部package关系，不再引用原作者电脑上的Excel路径；
- 报告生成时同时更新图表XML缓存和内嵌Excel单元格；
- Word中“编辑数据”可打开随DOCX携带的工作簿，刷新图表不会恢复外部示例文件；
- 图表的位置、大小和样式仍可在Word中调整。

> [纠正 - 2026-08-14] 内嵌Excel V2模板因Microsoft Word仍无法稳定打开，已暂停并从staging撤回。staging当前恢复为V1内容控件模板，SHA-256=`3b6a83fd4a2fddf7c0a47c1eda5e2e4141b7d0d72fd9431980928be598e86b92`。下载文件名包含SHA前8位，应为`competency-phase1-report-3b6a83fd.docx`；客户当前应基于此文件调整版式。V2候选不得重新上传为生效模板。

> [补充 - 2026-08-17] FB-125已将V1模板升级为LibreOffice兼容的流式布局，当前staging模板SHA-256=`42866f2768bf35115831ce0c24deb7aeff0f14a62fb441305e9dffe0faaca297`，下载文件名应为`competency-phase1-report-42866f27.docx`。十个维度标题与定义已由浮动组合图形改为普通流式段落，服务器真实报告恢复为A4 9页且无重叠、空白页和错误总页数。V2内嵌Excel仍暂停。

> [补充 - 2026-08-17 FB-127] 当前staging V1模板SHA-256=`37caebca5ae3b0cf5b986cf1f797e0d7be16bab6189c7154a5c24f3c322aa869`，下载文件名应为`competency-phase1-report-37caebca.docx`。十个维度环形图均已固定为对应得分表格单元格内的行内对象，避免LibreOffice将首图浮到图示说明区域。V2内嵌Excel继续暂停。

> [补充 - 2026-08-17 FB-128] 当前staging V1模板SHA-256=`70bcf95368b3387c7090a7ba848971a614723e24f0bf62c8f9f4baff7abcdf2b`，下载文件名应为`competency-phase1-report-70bcf953.docx`。一级得分汇总行和结果分析区已使用同一组动态字段，避免模板示例分值与真实报告不一致。V2内嵌Excel继续暂停。

> [补充 - 2026-08-17 FB-129] 当前staging V1模板SHA-256=`6f43de7c05bf5335a891ad6c1af694b9599954e48acb241fe1a131f506de3736`，下载文件名应为`competency-phase1-report-6f43de7c.docx`。两个一级维度改为共同0–5刻度的独立横向得分条，不再使用构成比饼图。V2内嵌Excel继续暂停。

> [纠正并完成 - 2026-08-18 FB-131] 根因不是单纯Schema错误，而是LibreOffice重存改写整个Word包；即使SDK零错误仍可能被Word拒绝。最终模板从Word原生V1基线定向构建，当前staging SHA-256=`0899d49768cf1463ecb97544d2ef7d11332ef6ae42133a10e89501d9e7a33fac`，下载文件名应为`competency-phase1-report-0899d497.docx`。本机Word 16已真实打开空模板、填充DOCX和staging API下载文件，均无修复写回；LibreOffice报告保持A4 9页。

## 6. 转换配置

默认采用LibreOffice，不需要Microsoft 365组织租户或客户端密钥：

- `PHASE1_WORD_REPORT_ENABLED=true`
- `PHASE1_WORD_REPORT_TEMPLATE_PATH=./configs/export-templates/competency-phase1-report.docx`
- `PHASE1_WORD_REPORT_FALLBACK_CHROMIUM=true`
- `PHASE1_WORD_REPORT_CONVERTER=libreoffice`
- `LIBREOFFICE_PATH=libreoffice`
- `PHASE1_WORD_REPORT_TIMEOUT_SECONDS=90`

运行模板已针对服务器LibreOffice分页基线做确定性垂直间距兼容处理。客户调整模板后，必须以服务器实际转换结果验收，不能只依据Microsoft Word中的页数。

如后续具备Microsoft 365组织租户，可将转换器切换为`graph`。Graph密钥只使用环境变量注入，不写入Git或YAML：

- `PHASE1_WORD_REPORT_CONVERTER=graph`
- `MSGRAPH_TENANT_ID`
- `MSGRAPH_CLIENT_ID`
- `MSGRAPH_CLIENT_SECRET`
- `MSGRAPH_DRIVE_ID`
- `MSGRAPH_REPORT_FOLDER=talent-assessment-reports`
- `MSGRAPH_TIMEOUT_SECONDS=90`

Graph应用应只获得指定报告Drive/Site所需的最小文件读写权限。程序使用客户端凭据取得令牌，上传临时DOCX，下载Word转换的PDF，并在成功或失败后删除远端临时文件。日志和API错误不会输出令牌或客户端密钥。

## 7. 发布验收

每次客户调整模板后必须完成：

1. DOCX ZIP结构可打开；
2. 49个唯一内容控件Tag完整且全部可映射；
3. 12个图表的数据点数量未变化；
4. 客户模板无可见双花括号，真实结果生成后所有内容控件均被正式值填充；
5. 转换器返回文件头为`%PDF-`且大小在限制内；
6. A4页数和客户预期一致；
7. 姓名、人员字段、时间、总体、一级、十维、效度和免责声明正确；
8. LibreOffice本地临时目录或Graph远端临时文件已清理；
9. 报告实例SHA-256、生成/重生成/下载审计正确；
10. 当前转换器故障时Chromium回退路径可用。
