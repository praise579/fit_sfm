# HTTP API 参考

本文档介绍 fit_sfm（sing-box 订阅转换服务）对外提供的全部 HTTP 接口。所有描述均以源码实际行为为准（`internal/handler/handler.go`、`cmd/run_server.go`）。

服务默认监听端口为 `9000`（配置文件 `server.port`），地址形如 `http://<host>:<port>/`。下文示例统一使用 `http://<host>:<port>/...` 的可替换写法，请将 `<host>`、`<port>` 替换为你的实际地址。

## 路由表

| 方法 | 路径 | 用途 | 是否需要鉴权 |
|------|------|------|---------------|
| 任意 | `/` | 订阅转换主接口：将节点套入模板渲染并返回 sing-box 配置 | 是（`password` 参数） |
| 任意 | `/health` | 健康检查，返回数据/模板加载状态 | 否 |
| 任意 | `/refresh` | 手动刷新：重新拉取节点与模板并重载；可选清理 Cloudflare 缓存 | 是（`password` 参数） |

路由使用 Go 标准库 `http.ServeMux` 注册（见 `cmd/run_server.go`）：

```go
mux.HandleFunc("/", handler.HandleRequest)        // 主要订阅转换接口
mux.HandleFunc("/health", handler.HandleHealth)   // 健康检查接口
mux.HandleFunc("/refresh", handler.HandleRefresh) // 手动刷新接口
```

补充说明：

- 主接口只处理路径恰好为 `/` 的请求；访问任何其它路径（如 `/foo`）会**在鉴权之前**直接返回 `404`（空响应体），避免无关请求进入鉴权日志。
- 三个接口均未限制 HTTP 方法，使用 GET 访问即可；返回内容由代码中实际写入的 `Content-Type` 决定。
- 除 `/health` 外，`/refresh` 与 `/` 都通过查询参数 `password` 与配置 `auth.password` 做**严格字符串相等**校验。

---

## 认证方式

受保护接口把密码放在查询参数中：

```
password=<auth.password>
```

校验失败时返回：

| 状态码 | 响应体 |
|--------|--------|
| `401 Unauthorized` | `Password Error` |

- 主接口失败响应会设置 `Content-Type: text/plain`；`/refresh` 失败响应未设置该头。
- 密码中的特殊字符（如 `&`、`=`、空格）需做 URL 编码，否则会被解析为额外参数导致校验失败。

---

## 主接口：`GET /`

将订阅节点套入指定模板并渲染，返回 sing-box 配置。

### 查询参数

| 参数名 | 必填 | 说明 |
|--------|------|------|
| `password` | 是 | 访问密码，需与配置 `auth.password` 一致 |
| `template` | 否 | 模板 ID（配置 `templates` 下的键名，如 `default`、`ios`）。缺省时回退到 `type` 参数，仍为空则使用 `default_template` |
| `type` | 否 | 兜底模板 ID：当 `template` 为空时，`type` 会作为模板名使用（历史兼容）。同时该原始值会以 `setType` 变量注入模板上下文 |
| `refresh` | 否 | 取值为 `1` 或 `true` 时，在响应前先强制拉取最新节点文件与全部启用模板并重载（不会清理 Cloudflare 缓存，也不删除本地磁盘缓存） |

> 说明：`type` 最初定位为“自定义类型参数，传递给模板”，实际代码中它还承担**模板选择回退**功能（`template` 为空时 `templateName = type`），并把原始值放入模板上下文变量 `setType`。

### 处理流程

1. 路径非 `/` → 直接 `404`。
2. 校验 `password`；失败 → `401`。
3. 若 `refresh=1` 或 `refresh=true`：拉取节点文件（失败仅记日志）并重载数据；拉取全部启用模板并重载。
4. 确定模板名：`template` → `type` → `default_template`。
5. `EnsureTemplate`：模板配置不存在/未启用时报错；本地缓存文件缺失则立即下载；缓存超过模板更新间隔（`update_interval`，单位秒，默认 1 小时）则尝试更新，下载失败但本地有缓存时降级使用本地文件。
6. 校验模板存在且 `enabled: true`。
7. 用模板上下文渲染并返回。

### 模板上下文变量

渲染时注入的变量（模板中可直接使用）：

| 变量 | 说明 |
|------|------|
| `Nodes` | 所有去重节点的完整 JSON 文本（已处理逗号连接，模板中直接以 `{{ Nodes }}` 插入 `outbounds` 数组） |
| `setType` | 请求参数 `type` 的原始值 |
| `nodeCount` | 当前已加载的去重节点数量 |
| `noNode` | 当前所用模板配置的 `no_node` 值 |

另注册有模板过滤器 `NotesName`（`{{ "关键词" | NotesName }}`），用于按关键词筛选节点名称列表，其匹配语义与用法见 [templates.md](templates.md)。

### 请求示例

```bash
# 使用默认模板
curl "http://<host>:<port>/?password=your_password"

# 指定模板
curl "http://<host>:<port>/?password=your_password&template=ios"

# 使用 type 作为模板回退
curl "http://<host>:<port>/?password=your_password&type=ios"

# 强制刷新后再返回（先拉取最新节点/模板）
curl "http://<host>:<port>/?password=your_password&template=ios&refresh=1"
```

### 成功响应

- 状态码：`200 OK`
- `Content-Type: application/json`
- 响应体：模板渲染后的内容。模板的输出即接口返回，外层没有额外包装；只要模板本身是合法 JSON，返回体就是可直接作为 sing-box 配置使用的 JSON。
- 附加响应头：

| 响应头 | 示例值 | 作用 |
|--------|--------|------|
| `Content-Type` | `application/json` | 标识返回内容类型 |
| `Profile-Update-Interval` | `6` | 建议客户端每隔 6 小时检查更新 |
| `Subscription-Userinfo` | `upload=0; download=0; total=<节点数>` | 向客户端上报流量/节点总量 |
| `Cache-Control` | `no-cache, no-store, must-revalidate` | 禁止中间层缓存 |
| `Pragma` | `no-cache` | 兼容 HTTP/1.0 的防缓存头 |
| `Expires` | `0` | 内容立即过期 |

### 错误响应

| 状态码 | 触发条件 | 响应体 |
|--------|----------|--------|
| `401 Unauthorized` | `password` 与配置不一致 | `Password Error` |
| `400 Bad Request` | `EnsureTemplate` 阶段失败（模板不存在/已禁用/下载失败等） | `Template Error: <具体错误>`，例如 `Template Error: template 'xxx' not found in configuration` |
| `400 Bad Request` | 模板配置不存在或未启用 | `Template 'xxx' not found or not enabled` |
| `500 Internal Server Error` | 模板未加载到内存 | `Template 'xxx' not loaded` |
| `500 Internal Server Error` | 模板渲染出错 | `Server Error: <具体错误>` |
| `404 Not Found` | 请求路径不是 `/` | 空响应体 |

---

## 健康检查：`GET /health`

无需鉴权，用于监控服务状态。

### 请求示例

```bash
curl "http://<host>:<port>/health"
```

### 响应

- `Content-Type: application/json`

```json
{
  "status": "ok",
  "has_data": true,
  "has_template": true,
  "node_count": 10,
  "template_count": 3
}
```

字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | string | `ok` 或 `degraded`。节点数据与模板均加载成功时为 `ok` |
| `has_data` | bool | 是否已加载节点数据 |
| `has_template` | bool | 是否已加载至少一个模板 |
| `node_count` | int | 已加载的去重节点数量 |
| `template_count` | int | 已加载的模板数量 |

### 状态码

- `200 OK`：`has_data` 与 `has_template` 均为 `true`。
- `503 Service Unavailable`：无节点数据或无模板（服务降级，如启动时远端拉取失败且本地无缓存）。

---

## 手动刷新：`GET /refresh`

重新拉取全部启用模板与节点订阅、清空并重建本地缓存，可选触发 Cloudflare 缓存清理。

### 查询参数

| 参数名 | 必填 | 说明 |
|--------|------|------|
| `password` | 是 | 访问密码，需与配置 `auth.password` 一致 |

鉴权失败返回 `401`，响应体为 `Password Error`。

### 执行流程（按代码实际顺序）

源码中 `/refresh` 的执行顺序如下：

1. **（可选）清理 Cloudflare 缓存**：若 `cloudflare.enabled: true`，先调用 `PurgeCloudflareCache()`（见下节）。
2. **清空本地缓存**：
   - 清空内存中的节点与全部模板；
   - 删除磁盘缓存文件：节点文件（`cache.directory/node_file`）与各模板文件（`cache.directory/template_<模板名>.json`）。
3. **并发重新拉取并重载**：
   - 节点订阅文件（`subscription.url`）；
   - 所有启用模板（`templates.*.url`）；
   - 每项成功后分别重新加载到内存。
4. 汇总所有任务结果，返回 JSON。

> 注意：`/refresh` 无论 Cloudflare 是否启用都会刷新文件；Cloudflare 步骤失败不会中止文件刷新，只会被计入返回的 `errors` 列表。

### Cloudflare 缓存清理所需配置

触发 `PurgeCloudflareCache` 需要同时满足：

1. `cloudflare.enabled: true`；
2. `cloudflare.purge_url` 非空（形如 `https://api.cloudflare.com/client/v4/zones/<ZONE_ID>/purge_cache`）；
3. 配置以下任一认证方式：
   - `cloudflare.api_token`（推荐），请求头使用 `Authorization: Bearer <token>`；
   - 或同时配置 `cloudflare.api_key` 与 `cloudflare.api_email`，请求头使用 `X-Auth-Key` + `X-Auth-Email`。

请求行为 `POST {purge_url}`，请求体固定为：

```json
{
  "purge_everything": true
}
```

请求超时沿用 `subscription.timeout`（默认 30 秒）。Cloudflare 返回非 2xx 状态码时视为失败，错误信息形如 `cloudflare API returned status <code>: <body>`。

### 请求示例

```bash
curl "http://<host>:<port>/refresh?password=your_password"
```

### 成功响应

状态码 `200 OK`，`Content-Type: application/json`：

```json
{
  "status": "success",
  "message": "Files refreshed successfully",
  "node_count": 10,
  "template_count": 3
}
```

### 失败响应

任一任务（拉取/重载/Cloudflare 清理）失败时，状态码 `500 Internal Server Error`，`Content-Type: application/json`：

```json
{
  "status": "error",
  "errors": [
    "Node Subscription: fetch error: <具体错误>",
    "cloudflare cache purge: <具体错误>"
  ]
}
```

---

## 状态码速查

| 状态码 | 主接口 `/` | `/health` | `/refresh` |
|--------|-----------|-----------|------------|
| `200 OK` | 返回渲染配置 | 服务正常 | 刷新成功 |
| `401 Unauthorized` | 密码错误 | —（无需鉴权） | 密码错误 |
| `400 Bad Request` | 模板不存在/未启用/模板处理失败 | — | — |
| `404 Not Found` | 路径非 `/` | — | — |
| `500 Internal Server Error` | 模板未加载或渲染失败 | — | 有任务失败 |
| `503 Service Unavailable` | — | 数据或模板未加载 | — |
