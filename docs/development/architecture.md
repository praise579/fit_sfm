# 架构说明

> 本文面向想要理解或修改本项目的开发者。简明架构另见根目录 [CLAUDE.md](../../CLAUDE.md)。

## 一句话定位

**sing-box 订阅转换服务**：周期抓取远程节点订阅与多份 sing-box 模板，缓存到本地目录，接收 HTTP 请求后用 pongo2 模板引擎把节点套入指定模板，渲染出适配不同客户端/版本的 sing-box 配置文件。

## 进程内调用链

```
main.go
  └─ cmd.Execute            // cobra 根命令
      ├─ run                // run 子命令：定位配置文件、启动 Server、监控配置文件热重启
      └─ run_server.Server  // 组装服务
           ├─ pkg/logger     构建 zap 日志
           ├─ internal/fetcher.Init   注册抓取回调
           ├─ internal/handler.Init   初始化渲染引擎与 NotesName 过滤器
           ├─ internal/fetcher        启动时的首次抓取 + 定时刷新
           ├─ internal/watcher.Start  监控缓存目录，文件变化触发 reload
           └─ http.Server           注册路由后监听
                ├─ "/"        handler.HandleRequest    订阅转换主接口
                ├─ "/health"  handler.HandleHealth      健康检查
                └─ "/refresh" handler.HandleRefresh     手动刷新（可触发 Cloudflare purge）
```

关键入口：

- `cmd/run.go` — 配置文件定位与加载顺序、配置变更热重启（radovskyb/watcher）。
- `cmd/run_server.go` — `Server` 装配与生命周期；路由在 `run_server.go` 用 `http.NewServeMux` 注册。
- `main.go` — 通过 `//go:embed config/config.yaml` 把默认示例配置嵌入二进制。

## 配置与加载

- 配置结构定义于 `global/config.go`（`Config` 及各级 `*Config`）。
- 加载顺序（高 → 低）：`-c` 指定文件 > 工作目录 `config/config-dev.yaml` > 工作目录 `config.yaml` > 内嵌默认 `config/config.yaml`。详见 [usage.md](../guide/usage.md)。
- 加载后支持**环境变量覆盖**，随后 `Validate()` 强校验（端口、密码、订阅 URL、模板与 `default_template` 必填且引用存在、至少一个模板 enabled）。
- `default_template` 决定不显式指定模板时使用的模板；请求可显式选择其它启用模板。

## 核心模块职责

| 模块 | 职责 |
|---|---|
| `internal/fetcher` | 按 URL 抓取**节点订阅**与**各模板文件**到缓存目录；支持 `file://` 本地协议与 cache-buster；提供立即/按名/全部抓取入口与存在性/修改时间查询 |
| `internal/handler` | HTTP 处理与 pongo2 渲染；注册 `NotesName` 过滤器做节点名筛选；健康检查、手动 refresh、Cloudflare 缓存清理 |
| `internal/watcher` | fsnotify 监听缓存目录；节点/模板缓存文件变化时分别触发 `ReloadData` / `ReloadTemplateByName` |
| `global` | 全局配置结构、加载/校验/环境覆盖、进程根目录等环境信息 |
| `pkg/logger` | zap 封装：全局句柄 + 按需「stderr + 文件(可 JSON)」双写日志 |
| `pkg/safe_close` | 优雅关闭协调器：等所有受监管子 goroutine 退出后再返回 |
| `pkg/convert` | 通用类型/结构转换工具 |
| `pkg/fileurl` | 文件/URL/路径处理 |
| `pkg/util` | array / converter / crypto / hash / machine / password / random / time / validator 等小工具 |

## 数据流与缓存

1. **抓取**：启动后 `fetcher` 首次抓取节点订阅与全部启用模板，写入缓存目录（节点文件 `node.json`，各模板 `template_<name>.json`）；随后按 `subscription.refresh_interval` 与模板级 `update_interval` 周期刷新。
2. **渲染**：请求到达 `/`，handler 读缓存节点 + 目标模板，pongo2 渲染 `{{ Nodes }}`、`{{ "关键词" | NotesName }}` 等变量。
3. **热更新**：
   - *数据热更*：`internal/watcher` 监听缓存目录文件变化 → 触发 reload。
   - *配置热更*：`cmd/run.go` 监控配置文件本身变化 → 重新加载并重启服务。
4. **缓存清理**：手动 `/refresh` 触发重抓；配置了 `cloudflare.enabled` 时按 `purge_url` 清理 CDN 缓存（Token/Key 二选一）。

## 模板渲染机制

- 引擎：pongo2。核心变量与过滤器见 [templates.md](../guide/templates.md)。
- 每个模板对应一份远端 JSON 模板文件（可按客户端 sing-box 版本区分），经 `NotesName` 过滤器按关键词把节点名筛选进 `outbounds`。

## 进程生命周期

- `Server` 使用 `pkg/safe_close` 管理关闭：收到关闭信号后停止接收新请求，`http.Server.Shutdown` 优雅等待，最后 `logger.Sync()` 落盘缓冲日志。
- 相关设计取舍记录在 [decisions.md](../decisions.md)。
