---
description: "Bug 驱动测试规则：修复 bug 时强制先写 RED 测试再改代码"
---

# Bug 驱动测试规则

**核心规则**：修复任何 bug 必须先写一个能复现该 bug 的失败测试（RED），再改代码使其通过（GREEN）。

## 1. 触发条件

当用户消息包含以下任一关键词时，强制启动本流程：

- "修 bug" / "修复" / "fix" / "解决"
- "出错" / "报错" / "异常" / "失败"
- "不能" / "无法" / "失效"
- "应该是 X 但实际是 Y"

## 2. 强制执行的 5 步流程

### Step 1：理解 bug

不要立即写代码。先问清楚：

- **复现路径**：什么操作触发？
- **预期行为**：应该怎样？
- **实际行为**：实际怎样？
- **环境**：local / staging / production？
- **触发数据**：什么样的输入会触发？

如果用户已提供这些信息，跳到 Step 2。否则向用户提问。

### Step 2：定位代码

读取相关 handler / 组件代码，标识 bug 在哪个分支。

### Step 3：写 RED 测试

**先写测试，不改代码**。测试应：

- 命名：`TestBug{编号}_{简短描述}`
- 注释：链接到 docs/regression-tests.md 中的 backlog 编号
- 模拟 bug 触发条件
- 断言期望行为
- **运行测试 → 必须失败**（这证明测试真在测 bug）

```go
// TestBug{N}_{Description}
// 对应：docs/regression-tests.md #{N}
// 复现：{操作步骤}
// 期望：{预期行为}
// 实际：{当前 bug 行为}
func TestBug{N}_{Description}(t *testing.T) {
    // GIVEN: {触发条件}
    // WHEN:  {操作}
    // THEN:  {期望结果}
    ...
}
```

### Step 4：在账本中登记

在 `docs/regression-tests.md` 添加一行：

```markdown
| {N} | {Bug 描述} | {文件} | 🔴 RED | {测试位置} |
```

### Step 5：修复代码 → 验证 GREEN

修改代码后：

1. 重新运行测试 → 必须通过（🟢 GREEN）
2. 运行所有相关测试 → 不能破坏其他测试
3. 在 regression-tests.md 中将状态改为 🟢 GREEN
4. 在 business-branches.md 中找到对应分支 → 标记 ✅
5. 提交时 commit message 引用：`fix: ... (closes regression #{N})`

## 3. 例外情况

### 不需要写测试的 bug

- 纯 UI 视觉问题（色差、间距）→ 改用截图测试或 UX-视觉专家审查
- 配置文件错误 → 验证步骤记录在 docs/deployment-guide.md
- 第三方依赖问题 → 升级依赖即可

但**所有这些例外必须在 regression-tests.md 标 ⚪ N/A 并写明理由**。

## 4. 报告格式

执行完 5 步流程后，在响应末尾追加：

```
📋 Bug 修复报告
═══════════════════════════════════════
Bug 编号：#{N}
描述：{描述}
RED 测试：{测试名} ({文件})
GREEN 验证：✅ 通过
账本更新：
  - regression-tests.md: 状态 🔴 → 🟢
  - business-branches.md: 分支 ❌ → ✅

🔄 后续建议：
  - 是否要检查类似的 handler？{是/否}
  - 是否需要更新 project-memory.md？{是/否}
```

## 5. 反模式（必须避免）

❌ **直接改代码不写测试** — bug 会复发
❌ **写完代码再补测试** — 测试可能写得太宽松（看代码反推测试）
❌ **测试只测 happy path** — 必须测引发 bug 的具体条件
❌ **修复后忘了更新账本** — 失去回归保护可见性

## 6. 时间预算

写 RED 测试 + 登记账本 通常增加 10-20% 时间。这是测试盲区减少机制的"利息"，不是浪费。

## 7. 与既有规则的协同

- 优先于"快速修复"心理 — 即使是 1 行的小 bug 也要走完流程
- 与 coverage-aware 规则配合 — 修 bug 后还要检查同类分支
- 与 prompt-optimizer 配合 — 优化提示词时记得加入"先写测试"约束
