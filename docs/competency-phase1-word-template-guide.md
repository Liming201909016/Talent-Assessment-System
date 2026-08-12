# 00401 一期 Word 报告模板维护说明

## 1. 文件与渲染链

- 客户源文件：`docs/260807胜任力开发资料/胜任力测评报告样例.docx`
- 运行模板：`Go-based Refactored System/configs/export-templates/competency-phase1-report.docx`
- 模板生成器：`scripts/tools/build-competency-phase1-word-template.py`
- PDF转换：Microsoft Graph（Microsoft 365 Word转换引擎）
- 故障回退：可配置回退到原Vue/Chromium报告

当前运行模板SHA-256：`7808bd325d51e4967c0bb128358bbcdb1153f3312aac3465cf60f157090bf991`。

## 2. 客户可修改内容

客户可以直接用Microsoft Word调整：

- 字体、字号、颜色、行距和段落间距；
- 页边距、分页符、表格尺寸和单元格样式；
- 封面、背景、装饰图形和图片；
- 12个原生Word图表的颜色、线型、图例和位置；
- 固定标题、阅读说明、维度定义、等级说明等固定内容；
- 各动态字段所在位置。

客户修改后，应将DOCX交回并提供版本号。部署前由程序执行占位符、图表、页数和PDF验证。

## 3. 禁止修改的动态标记

运行模板包含49个唯一动态占位符。占位符由双花括号包围，例如：

- `{{report.date}}`
- `{{participant.name}}`
- `{{participant.age}}`
- `{{participant.gender}}`
- `{{participant.telephone}}`
- `{{participant.affiliation}}`
- `{{participant.post}}`
- `{{result.submittedAt}}`
- `{{result.userTime}}`
- `{{overall.level}}`
- `{{overall.diagnosis}}`
- `{{validity.notice}}`
- `{{report.disclaimer}}`
- `{{group.general_ability.score}}`
- `{{group.general_ability.level}}`
- `{{group.general_ability.description}}`
- `{{group.psychological_quality.score}}`
- `{{group.psychological_quality.level}}`
- `{{group.psychological_quality.description}}`
- `{{dimension.<维度ID>.score}}`
- `{{dimension.<维度ID>.level}}`
- `{{dimension.<维度ID>.diagnosis}}`

维护要求：

1. 不得修改、翻译或删除占位符名称；
2. 一个占位符内部不要局部改变字体或颜色，避免Word将其拆成多个XML文本节点；
3. 可以移动整个占位符到其他段落或表格单元格；
4. 维度定义是客户维护的固定Word正文，不由程序覆盖；
5. 保存格式必须为DOCX，不接受DOC、PDF或WPS私有格式。

## 4. 原生图表映射

| Word图表 | 动态数据 |
|---|---|
| chart1 | 通用能力、心理素养两个一级得分 |
| chart2 | 十个二级维度雷达图得分 |
| chart3–chart12 | 十个维度的“当前得分 / 距5分差值”环形图 |

客户可以修改图表样式，但不能删除数据系列或改变数据点数量：

- chart1：2个数值；
- chart2：10个数值；
- chart3–chart12：每个2个数值。

## 5. Microsoft Graph配置

只使用环境变量注入，不把密钥写入Git或YAML：

- `PHASE1_WORD_REPORT_ENABLED=true`
- `PHASE1_WORD_REPORT_TEMPLATE_PATH=./configs/export-templates/competency-phase1-report.docx`
- `PHASE1_WORD_REPORT_FALLBACK_CHROMIUM=true`
- `MSGRAPH_TENANT_ID`
- `MSGRAPH_CLIENT_ID`
- `MSGRAPH_CLIENT_SECRET`
- `MSGRAPH_DRIVE_ID`
- `MSGRAPH_REPORT_FOLDER=talent-assessment-reports`
- `MSGRAPH_TIMEOUT_SECONDS=90`

Graph应用应只获得指定报告Drive/Site所需的最小文件读写权限。程序使用客户端凭据取得令牌，上传临时DOCX，下载Word转换的PDF，并在成功或失败后删除远端临时文件。日志和API错误不会输出令牌或客户端密钥。

## 6. 发布验收

每次客户调整模板后必须完成：

1. DOCX ZIP结构可打开；
2. 49个唯一占位符完整且全部可映射；
3. 12个图表的数据点数量未变化；
4. 真实结果生成后无未替换的双花括号；
5. Graph返回文件头为`%PDF-`且大小在限制内；
6. A4页数和客户预期一致；
7. 姓名、人员字段、时间、总体、一级、十维、效度和免责声明正确；
8. Graph临时文件已清理；
9. 报告实例SHA-256、生成/重生成/下载审计正确；
10. Graph故障时Chromium回退路径可用。
