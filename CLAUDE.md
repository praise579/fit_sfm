# CLAUDE.md

面向 Claude Code / AI 助手与本仓库协作者的项目速览。更完整的用户与开发文档在 [`docs/`](docs/README.md)；阅读顺序建议：本文 → [架构说明](docs/development/architecture.md)。

## 项目概述

**sing-box 订阅转换服务**：周期抓取远程节点订阅与多份 sing-box 模板并缓存到本地，通过 HTTP 请求把节点套入指定模板，用 pongo2 渲染出适配不同客户端/版本的 sing-box 配置文件。Go 1.24 编写，模块路径 `github.com/praise579/fit_sfm`。

## 常用命令

```bash
go build ./...              # 编译（含测试）
go test ./...               # 跑全部测试（make test 等价）
go vet ./...                # 静态检查

# 运行（开发）
go run . run -c config/config.yaml

# 交叉编译（make build-<os>-<arch>）
make build-linux-amd64      # Linux AMD64
make build-linux-arm64      # Linux ARM64
make build-macos-amd64      # macOS Intel
make build-macos-arm64      # macOS Apple Silicon
make build-windows-amd64    # Windows AMD64
make build-all              # 以上全部平台
```

产物在 `build/<os>_<arch>/sb-sub-c`。Air 热重载开发可用 `.air.toml`。

## 配置加载顺序（优先级从高到低）

1. `-c` / `--config` 命令行参数指定文件
2. 工作目录下 `config/config-dev.yaml`（本地开发覆盖，已 gitignore）
3. 工作目录下 `config.yaml`（本地覆盖，已 gitignore，勿提交）
4. `config/config.yaml`（仓库示例，通过 `//go:embed` 作为内嵌默认）

加载后支持环境变量覆盖（`SERVER_PORT`/`PASSWORD`/`SUBSCRIPTION_URL`/`DEFAULT_TEMPLATE`/`CACHE_DIR`/`REFRESH_INTERVAL`）并执行强校验。详见 [usage.md](docs/guide/usage.md)。

## 目录结构

```
main.go            // 入口：//go:embed config/config.yaml 内嵌默认配置
cmd/               // cobra 命令层：root / run / run_server(Server 装配) / version
global/            // Config 结构体与加载校验、进程根目录 env、版本变量
internal/fetcher   // 抓取节点订阅与模板到缓存目录（支持 file://、cache-buster）
internal/handler   // HTTP 处理 + pongo2 渲染 + NotesName 过滤器 + refresh/Cloudflare purge
internal/watcher   // fsnotify 监听缓存目录，文件变化触发 reload
pkg/convert        // 类型/结构转换
pkg/fileurl        // 文件/URL/路径处理
pkg/logger         // zap 封装（全局句柄 + stderr/文件双写）
pkg/safe_close     // 优雅关闭协调器
pkg/util           // 杂项小工具
templates/         // 官方本地模板库（如 sing-box-macos-1.14.json；默认模板 file:// 指向此处，随 git 版本化）
config/config.yaml // 配置示例（内嵌，字段注释即权威说明）
```

## 架构（简要）

```
main.go → cmd/run.go(定位配置/热重启) → cmd/run_server.go (Server)
   ├─ internal/fetcher   首次抓取 + 周期刷新 → cache/
   ├─ internal/handler   Init 渲染引擎
   ├─ internal/watcher.Start  监听 cache/ → handler.ReloadData/ReloadTemplateByName
   └─ http.Server        路由 /(渲染) /health /refresh
```

调用链、数据流、缓存命名（`node.json`、`template_<name>.json`）与热更机制见 [architecture.md](docs/development/architecture.md)。

## 模板渲染机制

pongo2 模板引擎。核心变量：
- `{{ Nodes }}` — 插入全部节点完整 JSON
- `{{ "关键词" | NotesName }}` — 按关键词筛选节点名称列表
- 模板按客户端/版本拆分多份，`default_template` 兜底，请求可选模板

完整语法与示例见 [templates.md](docs/guide/templates.md)、[api.md](docs/guide/api.md)。

## 文档地图

- 使用/部署：`docs/guide/installation.md` `configuration.md` `usage.md` `templates.md` `api.md` `troubleshooting.md`
- 开发：`docs/development/architecture.md`
- 决策：`docs/decisions.md`　变更：`CHANGELOG.md`　贡献：`CONTRIBUTING.md`

## 改动时的伴随要求

- 改变配置字段、接口行为或架构时，同步更新：`config/config.yaml` 注释（SSOT）→ 对应 `docs/guide` 页 → 必要时在 `docs/decisions.md` 补一条「背景/决策/后果」。
- 用户可见的行为变化记入 `CHANGELOG.md`。
- 保持代码注释为简体中文；改动请先跑 `go test ./...` 与 `go vet ./...`。
