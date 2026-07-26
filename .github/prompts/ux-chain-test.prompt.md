---
mode: agent
description: "业务链浏览器测试 + 多专家点评：模拟真实用户走完整业务链，同步进行 UX/业务/无障碍专家分析，输出迭代清单"
tools: ["codebase", "read_file", "grep_search", "semantic_search", "file_search", "create_file", "replace_string_in_file", "run_in_terminal"]
---

# 业务链测试 + 多专家点评

对 ${input:chain_name} 业务链执行端到端浏览器测试，同步进行多专家分析，输出可执行的迭代清单。

支持的链：Chain-A（考试生命周期）/ Chain-B（题库管理）/ Chain-C（开放考试答题）/ Chain-D（MBTI）/ Chain-E（封闭考试）/ Chain-F（系统管理）/ Chain-G（导出模板）/ Chain-H（报告管理）

## 执行流程

### Phase 1：加载业务链定义

1. 读取 `docs/business-chains.md` → 找到 ${input:chain_name} 的完整步骤
2. 读取 `docs/project-memory.md` → 获取相关模块状态
3. 读取以下专家指令：
   - #file:.github/instructions/ux-density.instructions.md（密度专家）
   - #file:.github/instructions/ux-visual.instructions.md（视觉专家）
   - #file:.github/instructions/business-chain.instructions.md（业务专家）
   - #file:.github/instructions/ui-ux-review.instructions.md（交互专家）

### Phase 2：环境检查

确认服务可用：
- Go 后端：`http://127.0.0.1:8092`（端口可达）
- 前端 dev server：`http://127.0.0.1:8089`（端口可达）
- 数据库 SSH tunnel：`23306` 端口可达
- 测试账号：admin / cp@1234 可登录

如不可用，提示用户先启动服务。

### Phase 3：Layer 1 — 单页操作项穷举

对业务链涉及的每个页面，自动枚举所有操作项：

```javascript
// 注入到 Puppeteer 页面执行
const operations = await page.evaluate(() => {
  const items = []
  // 所有按钮
  document.querySelectorAll('.el-button, button, a[href]').forEach(el => {
    items.push({
      type: el.tagName.toLowerCase(),
      text: el.innerText.trim(),
      classes: el.className,
      disabled: el.disabled || el.classList.contains('is-disabled'),
      visible: el.offsetParent !== null,
      bbox: el.getBoundingClientRect()
    })
  })
  // 所有表单输入
  document.querySelectorAll('input, select, textarea').forEach(el => {
    items.push({ type: 'input', ... })
  })
  // 所有表格行操作
  document.querySelectorAll('.el-table__row').forEach(row => { ... })
  return items
})
```

输出：每页操作项清单（按钮/链接/输入/分页/排序）

### Phase 4：Layer 2 — 业务链执行

按 `business-chains.md` 中定义的步骤顺序执行：

1. 启动 Puppeteer + viewport 1366x768（桌面）+ 375x667（移动端）
2. 登录 → 按步骤导航 → 操作 → 截屏
3. 每步验证：
   - HTTP 状态码 200
   - 页面无 console error
   - DOM 出现预期元素
   - 数据有变化（创建后列表 +1，删除后列表 -1）
4. 截屏保存到 `scripts/test/screenshots/${chain_name}/`

### Phase 5：Layer 3 — 多专家并行分析

对每张截图同时进行 4 个专家分析：

#### A. UX 密度专家
基于 ux-density.instructions.md 规则，给出量化指标：
- 默认行数 / 屏幕利用率 / 列数 / 平均列宽 / 操作列宽
- 建议改进的具体数字

#### B. UX 视觉专家
基于 ux-visual.instructions.md 规则：
- 颜色总数 / 按钮层级分布 / 主 CTA 数量 / 图标库数量
- 视觉混乱点定位

#### C. 业务专家
基于 business-chain.instructions.md 规则：
- 业务链断裂点（按钮 404 / 跳转失败 / 状态不同步）
- 操作顺序合理性
- 业务规则违反

#### D. 无障碍专家
- 图标按钮缺 aria-label
- 颜色对比度
- 键盘导航
- 触摸目标 ≥ 44px（移动端）

### Phase 6：汇总输出 — 迭代清单

按优先级排序输出，**格式必须能被下次会话的 Copilot 直接消费**：

```markdown
═══════════════════════════════════════
🧪 业务链测试报告 — ${input:chain_name}
═══════════════════════════════════════

## 📊 测试统计
- 链路步骤：X / Y 通过
- 单页操作项：X / Y 通过
- 截图数量：X 张
- 专家发现：P0 X / P1 X / P2 X / P3 X

## 🔴 P0 — 阻塞性问题（必须立即修复）

### 1. [业务-链断] Step X：操作名
**位置**：文件:行号
**症状**：截图 #X 显示 ...
**根因**：...
**修复方案**：
\`\`\`vue
// 修改 xxx.vue 第 X 行
- 旧代码
+ 新代码
\`\`\`
**验证方法**：重新执行 Step X 应通过

## 🟡 P1 — 影响体验（建议修复）

### 2. [UX-密度] 列表页屏幕利用率仅 58%
**当前指标**：默认 10 行，行高 56px，操作列 200px
**目标指标**：默认 20 行，行高 44px，操作列 140px
**修复方案**：
- pageSize 默认值改为 20
- 表格 size="mini"
- 操作列改用 icon button + tooltip
**预计效果**：单屏可见 10 → 18 条记录

### 3. [UX-视觉] MBTI 详情页使用 7 种按钮颜色
**当前**：primary 4 / success 3 / warning 2 / danger 2
**建议**：保留 1 个主操作 primary，其余 plain；危险用 danger
**截图证据**：screenshots/${chain_name}/03-detail.png

## 🟢 P2 — 优化建议

...

## 📁 截图清单
- 01-login.png — 登录页
- 02-list.png — 列表页
- 03-detail.png — 详情页
- ...

## 🔄 下次迭代输入

将以下内容粘贴给 Copilot 进行修复：

```
请修复 ${input:chain_name} 业务链的以下问题：
1. [P0] 文件:行 - 修复方案
2. [P1] 文件:行 - 修复方案
...
完成后重新执行 /ux-chain-test ${input:chain_name} 验证
```
```

### Phase 7：询问下一步

输出报告后询问用户：
- 是否要立即修复 P0 问题？
- 是否要将报告保存到 `scripts/test/screenshots/${chain_name}/report.md`？
- 是否要更新 `docs/project-memory.md` 记录本次迭代发现？

## 实施提示

如果 Puppeteer 脚本未实现自动枚举，可参考 `scripts/test/screenshot-chain-test.js` 现有框架，扩展添加：
- `getPageOperations(page)` — 枚举所有操作项
- `analyzeUXDensity(page)` — 计算密度指标
- `analyzeColorPalette(page)` — 提取使用的颜色列表

如未来需要，可创建 `scripts/test/ux-chain-runner.js` 作为新的复用框架。
