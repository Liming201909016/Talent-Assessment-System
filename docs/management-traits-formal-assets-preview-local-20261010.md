# 005 formal 正式资产只读预览（LOCAL GREEN）

日期：2026-10-10。状态：**slice2 LOCAL GREEN；不代表内容批准、正式PDF完成、数据库安装、staging/production发布。**

## 1. 本切片完成内容

新增管理员只读入口：

- `POST /exam/api/management-traits/formal/assets/preview`
- 请求只允许 `repoCode` 与 `assetKey`。
- `repoCode` 仅接受 `00501`、`00502`；旧002、00401、MBTI及其他产品均拒绝。
- 只读取 `MNG_FORMAL_ASSET_DIR/<assetKey>/content.xlsx` 与可选 `template.docx`，客户端不能提交路径。
- 仅当 `MNG_FORMAL_REGISTRY_ENV=local` 且资产根目录为绝对路径时可用。
- 不读取或写入正式登记数据库表，不登记版本、不批准、不启用、不生成PDF。

响应提供：

- 内容/模板是否存在；
- 205条规则计数；
- 工作簿SHA、规范内容SHA、模板SHA、binding SHA；
- 88个Tag、6张图、5个数字标签及计数；
- `readyForRegistration`、`readyForApproval`；
- 稳定且不含服务器路径的 `blockedReasons`。

所有数组字段均以非nil空数组返回，不产生JSON `null`。

## 2. 安全与失败关闭

资产读取保持并强化原严格门禁：

- 受控绝对目录、48位安全key；
- root/asset/file拒绝符号链接及非普通文件；
- 单文件最大20MiB；
- 同一句柄读取，并在打开前、打开后、读取后复核文件身份和大小；
- DOCX最多256部件、总解压64MiB、XML最多100000节点/128深度；
- 拒绝宏、ActiveX、嵌入对象、外部关系、关系逃逸、字段代码、TEST文案、图表公式和externalData；
- 精确要求88 Tag、6图、5数字标签及固定series/point合同。

preview把失败分类为受控reason，例如：

- `content_missing`
- `content_file_too_large`
- `content_identity_changed`
- `content_contract_invalid`
- `template_missing`
- `template_zip_invalid`
- `template_parts_limit`
- `template_uncompressed_limit`
- `template_xml_invalid`
- `template_relationship_invalid`
- `template_field_code_invalid`
- `template_test_marking_present`
- `template_contract_invalid`

reason不包含真实服务器路径、XML正文或客户内容。

## 3. 业务边界

- 缺模板但内容合法：可登记draft，`readyForRegistration=true`，`readyForApproval=false`。
- 内容或模板无效：只返回阻断原因，不自动修复、不自动登记。
- 候选完全合格：只证明结构可进入审阅，不等于客户内容批准或心理测量批准。
- 原Register/Approve/Activate仍会重新读取资产，不信任历史preview结果。
- TEST/reissue报告DTO、模板、current、路由和文件完全不变。
- 00401、MBTI、001、002历史链不变。

## 4. TDD与验证收据

### RED

- 服务编译失败：`PreviewAssets undefined`、`formalReadAssetChecked undefined`。
- HTTP真实请求命中404：preview路由不存在。
- RED exit=`1`。

### GREEN

- formal专项：35项通过，0失败。
- Go全量：全部包通过，exit=`0`。
- Go server build：通过，无编译错误。
- 编辑器诊断：本次Go文件0错误。
- `git diff --check`：exit=`0`。
- `gofmt -d`只报告既有Windows CRLF与gofmt LF的整文件行尾差异；新增代码缩进与import可编译，未使用终端改写文件绕过工作区编辑规则。

全量第一次在两个新增HTTP负向分支失败：环境未配置时输入错误被503优先遮蔽。修正为先校验005产品/key、再检查local环境；聚焦35项与最终全量均GREEN。失败历史保留，不用旧PASS冒最终结果。

## 5. 未完成与下一步

仍未完成：

1. 当前客户原始V2.8模板通过正式门禁；它仍有外部关系/字段等风险。
2. 可供客户双签的正式候选资产及批准包。
3. formal registry三表真实MySQL安装。
4. run→formal独立DTO、140/140及13维/4模块/receipt资格。
5. formal revision/current/audit独立存储。
6. 正式Word/LibreOffice/PDF生成、view/download。
7. 管理端版本/审批/报告UI及参与者用途UI。
8. staging、production及真实并发验收。

下一逻辑切片建议：**形成一份当前存在且能通过preview的005正式模板候选，但保持draft未批准**。在该切片前需确认页码字段政策（严格白名单PAGE/NUMPAGES或普通文本）以及00501/00502是两个独立version还是共享呈现批准+独立source binding。
