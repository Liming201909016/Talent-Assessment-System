# MT-FORMAL slice1 本地版本管理后端

日期：2026-10-06。**PARTIAL：正式PDF/两端UI/真实MySQL/双连接竞争/独立CodeReviewer未验证或未实现；SQL未执行、远端/部署/重启/历史迁移0。**

## 已确认政策与实现边界

- Q1=A：同一正常授权账号分别签content/psychometrics，两个不可覆盖的独立记录；重复职责签署返回原证据，不更新签署主体/日期。
- Q2=B：后续精确获批完整新版TEST run可追加独立formalPDF；本次注册从可信冻结profile抽取bundle/四评分轴/manifest/mapping，**不复制exam、个人字段、时间或run作为审批身份**。本次没有实现run→formalPDF，不改变TEST/legacy链。
- Q3=B：version撤销终态、epoch推进、audit保留；后续授权管理员读原历史PDF需显著revoked。本次只提供政策投影，formalPDF读取接口未安装。
- 未将205规则本身自动批准。合成fixture的actor7只是sqlmock单元数据，不是真实审批/seed；不创建管理员或签生产token。

## 影响及实际文件

| 文件 | 本次内容 | 影响处理 |
|---|---|---|
| [正式模型](../Go-based%20Refactored%20System/internal/model/management_traits_formal.go#L5) | version/approval/audit三实体 | 新增；旧模型/11表无需修改 |
| [正式registry](../Go-based%20Refactored%20System/internal/service/management_traits_formal_registry.go#L28) | 真实Register/List/Change，双签/activate/revoke同version行锁事务 | 新增；不调用TEST report writer、不生成PDF |
| [正式schema](../Go-based%20Refactored%20System/internal/service/management_traits_formal_schema.go#L14) | 全列类型NULL/bin collation、唯一键顺序、RESTRICT关系和孤儿门禁 | 新增独立合同；不存在关闭正式，部分安装失败关闭 |
| [正式资产门禁](../Go-based%20Refactored%20System/internal/service/management_traits_formal_assets.go#L17) | 受控key/同句柄预算读取、205原内容来源及规范digest、独立DOCX检查 | 新增；原XLSX/DOCX/TEST常量和客户文件不改 |
| [管理HTTP](../Go-based%20Refactored%20System/internal/handler/management_traits_formal_registry.go#L16) | 五接口，真实loginUser actor、严格扁平字符串JSON、固定中文错误 | 新增；普通exam:list不批准、未新增匿名豁免 |
| [router装配](../Go-based%20Refactored%20System/internal/router/router.go#L45-L47)及[注册](../Go-based%20Refactored%20System/internal/router/router.go#L233-L235) | 单独registry实例与formal前缀 | 同步修改，旧RuntimeHandler签名不改 |
| [旧守卫](../Go-based%20Refactored%20System/internal/service/management_traits_runtime_guard.go#L280-L315) | 正式三表完整严格校验后分类为非paper配置 | 同步修改，未知extra及原canonical report闭包仍拒绝 |
| [独立DDL](../scripts/sql/management_traits_002_formal_registry.sql#L1) | MySQL5.7/8 CREATE IF NOT EXISTS，3表/2 FK，无批准seed、无旧表ALTER | 新增C区；**未执行**，DDL自动提交，重跑不修漂移 |
| [事务测试](../Go-based%20Refactored%20System/internal/service/management_traits_formal_registry_test.go)、[公开服务与schema测试](../Go-based%20Refactored%20System/internal/service/management_traits_formal_behavior_test.go)、[资产测试](../Go-based%20Refactored%20System/internal/service/management_traits_formal_assets_test.go)、[HTTP测试](../Go-based%20Refactored%20System/internal/handler/management_traits_formal_registry_test.go) | 合成结构/实际候选读取、公开双签及显式启用撤销、rollback/幂等/权限与输入 | 新增；旧测试无需同步改 |

CRUDE：Create=Register/version、Change/approval/audit；Read=List与Change锁定version/approval、source bundle；Update=Change仅state/epoch，create_time/正文/hash不覆盖；Delete/Import/Export/Login=没有registry写入口；List=GET有界最新100及批量审批；Detail=同List投影，不手填路径。旧candidate/tester的Login/CRUD经legacy scope间接消费新表分类，保留原保护。

## API与配置调用点（C2/C4）

前缀：/exam/api/management-traits/formal。GET /versions；POST /versions/register、/versions/approve、/versions/activate、/versions/revoke。新接口当前消费者是测试，前端下一slice新增封装；现有前端/旧结果DTO无需修改，不能宣称UI已可用。

登记仅versionCode/examId/assetKey；审批versionId/expectedIdentitySha/expectedEpoch/kind（content或psychometrics）；启用前三项；撤销前三项加reason。epoch为规范正整数字符串。客户端actor/approvedAt/path/重复/未知/null/空串拒绝，时间只来自服务器。所有操作JWT与正userId1或正userId且*:*:*；没有伪细粒度permission支持。

新增两环境键只在registry构造读取一次：MNG_FORMAL_REGISTRY_ENV必须精确local；MNG_FORMAL_ASSET_DIR是服务器绝对受控目录。消费方=构造→gate/environment、assetsMatch/Register→资产读取；router装配及公开服务/HTTP测试同步验证，其他配置/TEST/Worker无需修改。未写任何.env/shared配置；staging/production/未设置均关闭，不以REPORT_EFFECTIVE_ENV或APP_ENV伪批准。

目录每assetKey下固定content.xlsx及可选template.docx；客户端只给48位小写ASCIIkey，拒路径/符号链接/大文件。原candidate内容可以登记draft；缺template可登记但canApprove/canActivate=false。模板必须88文本绑定、六语义图表/五数字槽、有限ZIP/XML、无TEST用途/宏/嵌入对象/外链/公式/字段代码；当前尚无内容修订上传器/正式渲染器/Word与LO视觉验收，不能当完整模板验收完成。

版本identitySHA绑定版本code、local、code、bundle、规范source、内容digest/原XLSX SHA/模板SHA/bindingSHA/key。变文件不原地更新identity，审批/启用重读并精确比；必须新revision重新双签。登记缺template后补文件不会复活该旧draft，需新版本。按明确versionId+epoch操作，不提供自动“最新版本”回退；环境+code全局当前版本选择及generation current在后续报告切片实现。

响应分离versionReady与canGenerate：版本双签并显式active可versionReady=true，**canGenerate恒false / generationBlockedReason=formal_report_pipeline_not_installed**，避免UI误认为报告功能已接通。撤销返回version_revoked及历史管理员可读政策，不冒充实际PDF权限校验。

## 新发现与失败历史

原lo-compatible候选正文实际16条LINK Excel.SheetMacroEnabled.12外部工作簿字段；原“零外链”仅OPC/chart关系范围，不含正文字段，不能直接作为正式资产。正式结构门禁真实拒绝，原文件SHA未修改。未删字段/生成正式模板/去TEST警示来绕过审批。所有字段代码暂严格拒绝（包括PAGE），后续正式模板合同需单独处理允许字段；本次不称候选可激活。

TDD：先新增registry测试→undefined模型/方法编译RED exit1；实现后sqlmock LIMIT参数缺失导致事务测试fail，补实际GORM参数后通过。真实候选测试首先相对路径错误、随后正式门禁拒绝；只读实际LINK证据后纠正测试为拒绝，不降低生产门禁。公开服务双签有效失败暴露新source wire缺JSON tag，补显式tag不改既有严格解码；最终GREEN。

## 已执行验证

- 新正式专项53pass事件/14顶层，0fail，原生exit0：包括公开Register、Change同账号两签、真实文件读取、显式activate/revoke、create_time/identity保留、重复签原证据零写、audit失败rollback及scope配置分类；Gin真实httptest 401/403/400/503/404响应，不真实登录/实库HTTP。
- 全量go test -json ./... -count=1：**6229pass、673顶层、0fail、9原环境skip、parse0、exit0**。child-only清23个既有测试开关键，GOOS=windows/GOARCH=amd64/CGO_ENABLED=0；不改变真实进程配置。不是coverage/race/真实MySQL。
- 既有Go Build任务完成；补原生go build -o bin/server.exe ./cmd/server、go build ./...、go vet ./...：**三exit0、stdout/stderr各0**。未Linux构建/前端构建/浏览器/PDF渲染，不拿旧证据当本轮PASS。
- 新生产Go及两现有改动文件diagnostics均0。真实SQL/3表first-repeat/FK/并发/性能未验证，SQL未执行。mode禁止派发其他agent，只有自审，无独立CodeReviewer PASS。
- 最终[本地收据](../scripts/test/results/mng-formal-registry-local-20261006-6591add72de5/verdict.json)：开工215既有Go源码仅router/guard两项授权变化、其余213逐SHA同；Legacy目录本机不存在，文件操作0，不能称Legacy全字节比对。50文档链接/行范围通过、原DDL0执行。首次scope传大JSON到Windows环境变量超长度限制失败，stdin复验exit0，首次无artifact，不隐藏失败；未修改原21漂移或harness。

下一：主协调者只读复审slice1→正式资产/预览完整合同→报告独立DTO/storage/current与Q2/Q3真实读取→管理UI→参与者UI。保持一逻辑slice，不自动部署或迁移。