# 00401 导入导出验收记录

**日期**：2026-07-26  
**环境**：staging `20.200.136.133`  
**范围**：00401 胜任力题库模板、题目导入、题目导出、胜任力测评结果导出

## 1. 结论

- 00401 导入模板已调整为 4 行：1 行表头、1 行填写说明、2 行示例题。
- 两行示例均属于 D01 沟通表达，分别覆盖正向和反向计分；题目编号为 `D01-EXAMPLE-F`、`D01-EXAMPLE-R`，维度内题号为 9001、9002。
- 模板下载、预览、正式导入、题目导出、结果汇总导出和原始答题导出均通过真实 staging 验证。
- 两道示例题仅作为临时验收数据，验收后已按主键删除；胜任力源题恢复为 384 道。
- Production 未部署。

## 2. 本地验证

| 项目 | 结果 |
|---|---|
| 模板 RED | 原模板只有 3 行，测试收到 `template rows/columns = 3/9` |
| 模板 GREEN | 4 行、9 列、无合并单元格；两道示例均通过导入校验；方向为 forward/reverse |
| 题目导出 RED | 后端缺导出数据结构/工作簿构建器，前端缺 `exportQuestions()` |
| 题目导出 GREEN | 导入兼容九列、正反向/启停中文值、空题库合法表头、查询失败先于响应头处理 |
| Go全量 | `go test ./... -count=1` 通过 |
| Go构建 | Windows、Linux构建通过 |
| 前端专项 | `competency-question-list.spec.js` 8/8 通过 |
| 前端全量 | 17 文件、104 项全部通过 |
| 前端构建 | production build 通过；仅保留既有 asset/entrypoint size 两项 warning |

## 3. Staging真实业务链

### 3.1 模板与导入

1. 管理员下载 `/exam/api/competency/questions/import-template`。
2. 返回 xlsx MIME、RFC 5987 文件名和 4 行模板。
3. 模板包含 2 行有效示例：正向 1 道、反向 1 道。
4. 预览结果：`successCount=2`、`errorCount=0`，返回 64 位 SHA-256。
5. 使用同一文件和预览 SHA-256 正式导入：`importedCount=2`。
6. 数据库胜任力题由 384 临时增加到 386。

### 3.2 题目导出

- 新增管理员端点：`GET /exam/api/competency/questions/export`。
- 导出工作表：`胜任力题目`。
- 列结构与导入模板完全一致：维度序号、维度名称、题目编号、维度内题号、题目内容、考察点、计分方向、启用状态、备注。
- 导出共 386 道题，包含刚导入的两道示例题。
- 正向/反向和启用状态与数据库一致。
- 排序为维度显示顺序、维度内题号、题目主键的稳定顺序。

### 3.3 测评结果导出

使用保留的完整测评 `1785060929295657251` 验证两个既有端点：

- `GET /exam/api/exam/exam/export-raw-data`
- `GET /exam/api/exam/exam/export-raw-answers`

两端点均返回相同规范化内容：

| Sheet | 数据行数 |
|---|---:|
| 结果汇总 | 1 |
| 逐题明细 | 40 |
| 题目字典 | 40 |

持久化结果核对：答题 `40/40`、整体分 `5.000000`、评价均值 `1.000000`；两个导出端点内容完全一致。

## 4. 清理与环境终验

- 两道临时示例题按查询所得主键删除，残留 0。
- 胜任力源题恢复为 384。
- 短时 Redis 管理会话残留 0。
- 远端临时二进制、前端归档和验收脚本已删除。
- 最近 20 分钟 `panic/fatal/unknown column/import failure/export failure` 计数为 0。
- `talent-assessment`、`nginx`、`mysql` 均 active。
- 内部和公网 health 均返回 `{"status":"ok"}`。

## 5. 备份与部署校验

- 数据库备份：`/opt/talent-assessment/backups/element_before_00401_io_20260726_185406.sql.gz`
- 备份大小：12,530,477 bytes
- 备份 SHA-256：`e26ca039c49d12b357025aa1c55dbf4673d2b79093d5a7e8ada24482c70e195c`
- 后端 SHA-256：`dcc7e3b59b6564a697354288dbb26a83e36be5463a6899183e747b62e0ab5fe5`
- 前端 index SHA-256：`b43a905991328291d6068a456e5a75faf3118b695ca2d68b87a456adbf95fdd0`
