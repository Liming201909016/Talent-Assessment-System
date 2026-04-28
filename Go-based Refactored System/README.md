# Talent Assessment - Go Refactored Backend

Go 后端，1:1 平移 Legacy Java（RuoYi-Vue 基线）的 talent-assessment 业务系统。

## 要求

- Go 1.22+
- 能访问同一 MySQL（默认通过 SSH 隧道 `127.0.0.1:13306`）
- 能访问同一 Redis（默认 `127.0.0.1:6379` db=1）

## 快速启动（本地）

```powershell
# 1. 配置
Copy-Item .env.example .env.local
# 按需修改 .env.local

# 2. 安装依赖
go mod tidy

# 3. 启动
go run ./cmd/server

# 或构建后运行
go build -o bin/server.exe ./cmd/server
./bin/server.exe
```

监听端口默认 `8092`（与 Java 8091 并行，方便对比）。

## 兼容性约定（关键）

- JWT：HS512，secret 与 Java 完全一致 (`abcdefghijklmnopqrstuvwxyz`)；claim 使用 `login_user_key: <uuid>`
- Redis：db=1；键 `login_tokens:<uuid>` / `captcha_codes:<uuid>` 复用 Java 格式（值结构尽量兼容）
- 密码：BCrypt（`golang.org/x/crypto/bcrypt`），与 Spring `BCryptPasswordEncoder` 同格式 `$2a$10$...`
- 响应体：沿用 Java 两种包装：
  - RuoYi 核心 `AjaxResult{code,msg,data}`
  - 业务 `ApiRest{code:0,msg,data,success}`

## 目录

```
cmd/server/              入口
internal/
  config/                 配置加载
  middleware/             gin 中间件（JWT/CORS/日志/恢复）
  router/                 路由注册
  handler/                HTTP 处理器
  service/                业务服务
  repository/             数据访问（GORM）
  model/                  DB 实体 & DTO
pkg/
  jwt/                    HS512 token 创建解析
  redis/                  redis 客户端
  db/                     gorm 数据库
  response/               响应包装
  captcha/                math 验证码
configs/                  application.yml
deploy/                   Dockerfile / nginx.conf / compose
```

详见 `docs/` 中的迁移报告。

## 测试与质量

| 命令 | 用途 |
|------|------|
| `make test` 或 `go test ./...` | 运行所有 Go 单元测试 |
| `make test-pkg` | 仅 pkg 层测试 |
| `make coverage` | 生成 HTML 覆盖率报告 |
| `make lint` | golangci-lint 静态分析 |
| `cd ruoyi-ui && npm test` | 前端 Vitest 单元测试 |
| `node scripts/test/chain-batch.js` | API 集成测试（需本地环境） |
| `node scripts/test/ux-chain-runner.js D` | 业务链 + 多专家点评（需浏览器环境） |

VS Code 中：`Ctrl+Shift+P` → `Tasks: Run Task` → 选择测试任务。

## 团队协作

新成员入职先读：

- [docs/team-onboarding.md](../docs/team-onboarding.md) — 30 分钟上手
- [.github/copilot-instructions.md](../.github/copilot-instructions.md) — Copilot 工作规则
- [docs/project-memory.md](../docs/project-memory.md) — 项目知识库

测试质量管理：

- [docs/business-branches.md](../docs/business-branches.md) — 业务条件分支矩阵
- [docs/regression-tests.md](../docs/regression-tests.md) — 回归测试 backlog
- [docs/coverage-history.md](../docs/coverage-history.md) — 覆盖率趋势
- [docs/user-feedback-log.md](../docs/user-feedback-log.md) — 用户反馈追踪

交付准备：

- [docs/demo-checklist.md](../docs/demo-checklist.md) — 客户演示前检查清单
