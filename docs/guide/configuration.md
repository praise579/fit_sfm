# 配置指南

本页说明 `config.yaml` 中各大配置区块的职责与典型写法，帮助你按使用场景快速配置 fit_sfm（Singbox Subscribe Convert）。

> 本页**不逐字段罗列类型与默认值**。`config/config.yaml` 的注释是配置字段的唯一事实源（SSOT），字段级类型、默认值与注释说明请直接查看该文件对应区块。
>
> 配置文件位置、加载优先级、环境变量覆盖与热重载机制见 [usage.md](usage.md)（「配置加载顺序」「环境变量覆盖」两节）。

## 配置总览

配置文件是一个 YAML 文件，顶层按职责划分区块，形如：

```yaml
server:        # 服务器监听与超时
auth:          # 访问认证
subscription:  # 节点订阅来源
templates:     # 模板列表（一个模板 = 一个客户端/版本）
default_template:  # 未指定模板时使用的模板 ID
cache:         # 本地缓存目录与文件
cloudflare:    # 刷新时同步清理 Cloudflare CDN 缓存
logging:       # 日志
```

最简可运行配置至少需要满足：`auth.password` 非空、`subscription.url` 非空、`templates` 中至少一个模板 `enabled: true`、`default_template` 指向一个真实存在的模板。程序加载配置时会做上述校验，不满足会直接报错退出。

复制示例配置作为自定义配置的步骤见 [installation.md](installation.md)（从源码编译运行 → 创建并修改配置文件）。

## server：服务器配置

`server` 区块控制 HTTP 服务监听端口与连接超时。

```yaml
server:
  port: 9000
  read_timeout: 15  # 秒
  write_timeout: 15 # 秒
  idle_timeout: 60  # 秒
```

要点：

- `port` 是服务监听端口，也可通过启动参数 `-p` 或环境变量 `SERVER_PORT` 覆盖（优先级更高）。
- 三个超时字段单位均为秒，用于 HTTP 服务器的读写与空闲超时。

> 字段级说明见 [`config/config.yaml`](../../config/config.yaml) 中 `server` 区块注释。

## auth：访问认证

`auth` 区块是整个服务唯一的访问凭据，保护配置下发与手动刷新两类接口。

```yaml
# 认证配置
auth:
  password: "your_secure_password"
```

要点：

- `password` 为必填项，留空时程序启动校验失败。
- 所有需要返回数据的接口都校验该密码：`GET /`（主接口）与 `GET /refresh`（手动刷新）均要求在 URL 查询参数中携带 `password`，不匹配时返回 `401` 与 `Password Error`。
- 密码会出现在订阅 URL 中，建议使用足够随机的高强度字符串；若含 `&`、`=`、空格等特殊字符，需先做 URL 百分号编码。
- 可通过环境变量 `PASSWORD` 覆盖，便于容器/进程注入，避免明文写入配置文件。

本服务不提供多用户、多 Token 体系；如需按用户区分，可在前端再加一层反向代理鉴权。

> 字段级说明见 [`config/config.yaml`](../../config/config.yaml) 中 `auth` 区块注释。

## subscription：节点订阅配置

`subscription` 区块定义上游节点来源：服务定时拉取该地址，解析出全部 sing-box 节点（`outbounds`），再注入到各个模板中。

```yaml
# 节点订阅
subscription:
  url: "https://your-subscription-url"
  timeout: 30          # 秒
  refresh_interval: 2  # 分钟
```

要点：

- `url` 为必填，指向节点订阅地址。除 `http(s)://` 远程地址外，也支持 `file://` 本地文件协议（把订阅文件放在本机，便于离线联调）。
- `refresh_interval` 为自动刷新间隔，单位**分钟**，须大于 0；到达间隔后程序会自动重新拉取节点与模板并热重载。
- `timeout` 为拉取请求超时，单位秒，不填时默认 30 秒；它同时用作刷新流程中调用 Cloudflare API 的请求超时。
- 手动刷新请调用 `GET /refresh?password=<密码>`，节点与各模板会并发拉取并重新载入内存。
- 可通过环境变量 `SUBSCRIPTION_URL`、`REFRESH_INTERVAL` 覆盖。

> 字段级说明见 [`config/config.yaml`](../../config/config.yaml) 中 `subscription` 区块注释。

## templates：模板配置

`templates` 是本服务的核心：同一批订阅节点，通过不同模板渲染出适配不同客户端 / 不同 sing-box 版本的配置。它是一个映射，键为模板 ID（用于 URL 参数引用），值为该模板的配置。

```yaml
# 模板列表
templates:
  # 模板 ID：default（默认；本仓库自带的本地官方模板，macOS + sing-box 1.14）
  default:
    url: "file://templates/sing-box-macos-1.14.json"
    name: "macOS 精简版"
    no_node: "🎯 全球直连"
    enabled: true

  # 模板 ID：openwrt（远程模板源示例）
  openwrt:
    url: "https://example.com/templates/openwrt.json"
    name: "OpenWRT"
    no_node: "🎯 全球直连"
    enabled: true

  # 模板 ID：ios（另一远程示例，带独立更新间隔）
  ios:
    url: "https://example.com/templates/ios.json"
    name: "iOS"
    no_node: "🎯 全球直连"
    enabled: true
    update_interval: 3600  # 可选，模板更新间隔（秒）

# 默认模板
default_template: "default"
```

要点：

- 每个模板的键（如 `default`、`openwrt`、`ios`）即模板 ID；请求时通过 `?template=<ID>` 指定，不指定时使用 `default_template`。
- `url`：模板文件地址，**同一列表可混排** `https://` 远程地址与 `file://` 本地文件（本地相对路径以运行工作目录为基准）。模板本质是含 pongo2 占位符的 JSON 文本，语法见 [templates.md](templates.md)。
- 仓库根 `templates/` 是随 git 版本化的**官方本地模板库**，`config/config.yaml` 的默认 `default` 模板即指向其中 `sing-box-macos-1.14.json`；clone 后在仓库根运行（`run` 或 `run -d <仓库根>`）即离线可用。改官方模板直接走 git，无需另存。
- `name`：模板显示名称，仅用于日志/标识，不影响请求参数。
- `no_node`：无节点兜底出口。当节点列表为空、或模板用 `NotesName` 过滤器筛选无结果时，用它保证 `outbounds` 非空（示例为 `🎯 全球直连` 直连出口）。
- `enabled`：是否启用。禁用的模板不会被加载，请求时返回 `Template ... not found or not enabled`。
- `update_interval`：可选，单位秒，控制该模板缓存多久后视为过期并重新拉取；不配置时默认按 1 小时判断（代码默认）。
- 每个模板有独立的本地缓存文件（按 `template_<模板ID>.json` 命名），多个模板并发更新、互不影响；`enabled` 开关可在配置热重载后动态生效。

> 字段级说明见 [`config/config.yaml`](../../config/config.yaml) 中 `templates` 区块注释。

## default_template：默认模板

`default_template` 是一个顶层标量，值为 `templates` 中的某个模板 ID：

```yaml
default_template: "default"
```

- 当请求 URL 未携带 `template` 参数时，使用该模板渲染。
- 该 ID 必须存在于 `templates` 中且对应模板已启用，否则配置校验失败。
- 可通过环境变量 `DEFAULT_TEMPLATE` 覆盖。

> 字段级说明见 [`config/config.yaml`](../../config/config.yaml) 中 `default_template` 行注释（含在文件顶部整体示例中）。

## cache：缓存配置

`cache` 区块控制节点与模板的本地磁盘缓存位置。缓存使服务在订阅源临时不可用时仍能继续对外提供配置。

```yaml
# 缓存配置
cache:
  directory: "./storage/cache"
  node_file: "node.json"
  template_file: "template.json"
```

要点：

- `directory`：缓存目录，相对路径以运行工作目录为基准；目录不存在时程序会自动创建。可用环境变量 `CACHE_DIR` 覆盖。
- `node_file`：节点缓存文件名，最终路径为 `目录/node_file`。
- `template_file`：模板缓存文件名（旧格式，保留兼容）；当前版本各模板实际按 `template_<模板ID>.json` 自动命名。
- 服务会监控缓存目录中的文件变化，缓存文件被更新后自动重新加载，无需手动重启（热重载）。
- Docker 部署时把该目录挂载为卷即可持久化保存。

> 字段级说明见 [`config/config.yaml`](../../config/config.yaml) 中 `cache` 区块注释。

## cloudflare：CDN 缓存清理配置

当订阅地址通过 Cloudflare CDN 加速时，源站更新后 CDN 上可能仍是旧缓存，客户端拿不到最新配置。配置本区块后，每次调用 `/refresh` 会**同步清理 Cloudflare 缓存**，让新配置立即生效。

```yaml
# Cloudflare 配置
cloudflare:
  enabled: false   # 是否启用 Cloudflare 缓存清理
  purge_url: ""    # Cloudflare 缓存清理 API 地址
  api_token: ""    # Cloudflare API Token（推荐）——在 Cloudflare 控制台创建
  api_key: ""      # Cloudflare API Key（可选）——与 api_email 一起使用
  api_email: ""    # Cloudflare 账户邮箱（可选）——与 api_key 一起使用
```

要点：

- `enabled` 为总开关；`purge_url` 在启用时必填，格式为 `https://api.cloudflare.com/client/v4/zones/{zone_id}/purge_cache`。
- **认证二选一**：
  - **API Token 方式（推荐）**：只配置 `api_token`。请求使用 `Authorization: Bearer <token>` 认证，Token 可在 Cloudflare 控制台按最小权限创建，便于单独授权与吊销。
  - **Global API Key 方式**：配置 `api_key` + `api_email`。请求使用 `X-Auth-Key` / `X-Auth-Email` 头认证，权限范围是整个账户，安全性较低。
  - 若两者都未配置而 `enabled: true`，刷新时会报 `cloudflare authentication not configured`。程序优先使用 `api_token`，未配置时才回退到 `api_key + api_email`。
- **触发时机与行为**：仅当 `cloudflare.enabled: true` 时，`GET /refresh?password=<密码>` 会先清理 CDN 缓存（`purge_everything`，即全量清理），再并发拉取节点与模板。清理失败会在刷新响应中体现为错误项，但不会中断其他刷新任务。
- **权限要求**：API Token 需具备 `Zone → Cache Purge → Purge` 权限，且 Zone Resources 包含目标域名；权限不足时 Cloudflare 返回 `403` / `Authentication error`。
- **获取方式**：Zone ID 在 Cloudflare Dashboard 的域名「概述」右侧 API 区域；API Token / Global API Key 均在 `My Profile → API Tokens` 下创建或查看。

> 字段级说明见 [`config/config.yaml`](../../config/config.yaml) 中 `cloudflare` 区块注释。

## logging：日志配置

`logging` 区块控制日志输出格式、文件与轮转。

```yaml
# 日志配置
logging:
  production: true
  file: "./storage/log/server.log"
  level: "info"  # debug, info, warn, error
  max_size: 10   # MB
  max_backups: 3
  max_age: 7     # 天
```

要点：

- `production`：是否生产模式。开启后日志以 JSON 结构输出（便于采集）；关闭则为更易读的文本格式。
- `file`：日志文件路径（相对路径以工作目录为基准，目录不存在时自动创建）；`level` 可选 `debug` / `info` / `warn` / `error`。排查 Cloudflare 清理请求细节、节点与模板加载情况时可临时调为 `debug`。
- `max_size` / `max_backups` / `max_age` 控制单文件最大体积（MB）、保留的旧文件数与保留天数，用于日志轮转。

> 字段级说明见 [`config/config.yaml`](../../config/config.yaml) 中 `logging` 区块注释。
