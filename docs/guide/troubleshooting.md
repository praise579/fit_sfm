# 常见问题排查（Troubleshooting）

本文档整理服务使用中的常见问题，按「症状 → 可能原因 → 排查与解决」结构编写。内容与源码行为对齐（配置校验见 `global/config.go`、文件监控见 `internal/watcher/watcher.go`、刷新逻辑见 `internal/handler/handler.go`）。

通用排查提示：

- 日志文件默认位于 `./storage/log/server.log`（由配置 `logging.file` 指定），需要更详细输出时可将 `logging.level` 设为 `debug` 后重启服务。
- 配置文件查找优先级：命令行 `-c`/`--config` > `config/config-dev.yaml` > `config.yaml` > `config/config.yaml`。

---

## 1. 服务启动失败（启动后立即退出）

**症状**：执行启动命令后进程很快退出，未监听端口。

**可能原因**：

- 配置文件缺失或 YAML 语法错误。
- 配置校验不通过。启动时会执行 `Validate()`，常见必填项缺失：
  - `server.port` 不在 `1–65535` 范围；
  - `auth.password` 为空；
  - `cache.directory` 为空；
  - `subscription.url` 为空；
  - `subscription.refresh_interval` 不大于 0；
  - `templates` 为空，或没有任何模板 `enabled: true`；
  - `default_template` 为空，或其值不在 `templates` 中。
- 端口被占用。

**排查与解决**：

1. 查看启动输出与日志文件（`logging.file`），确认是否有 `Error loading config` / `validate config error` 字样。
2. 用 `-c <配置文件>` 显式指定配置，排除加载到了错误配置文件的情况。
3. 逐一核对上述必填项。
4. 检查端口占用：`lsof -i:<port>`（macOS/Linux），更换 `server.port` 后重试。

> 注意：启动时若远端文件拉取失败，服务**不会**因此退出（仅打印警告并可能以降级状态运行），因此“启动失败”更多是配置或端口问题。

---

## 2. 无法获取节点

**症状**：返回的配置中没有节点；日志出现拉取失败；`/health` 的 `has_data` 为 `false`。

**可能原因**：

- `subscription.url` 配置错误或不可访问。
- 网络不通、DNS 解析失败或被防火墙拦截。
- 远端返回非 200、内容为空，或返回的不是预期的 sing-box `outbounds` 节点文件。

**排查与解决**：

1. 检查 `subscription.url` 是否为可访问的订阅节点地址。
2. 在服务器本机直接访问该 URL，确认网络与远端状态。
3. 检查防火墙/安全组是否放行服务出网。
4. 调用手动刷新并观察结果：
   ```bash
   curl "http://<host>:<port>/refresh?password=<密码>"
   ```
   若返回的 `errors` 中包含 `Node Subscription: fetch error: ...`，按错误信息定位。
5. 确认节点文件格式：程序按 JSON 解析并读取其中的 `outbounds` 数组，每个节点需含字符串 `tag`，无 `tag` 或重复 `tag` 的节点会被忽略。
6. 查看 debug 日志确认实际请求地址：fetcher 会向 URL 追加 `_t=<时间戳>&_r=<随机串>` 缓存穿透参数，并以浏览器 User-Agent 请求。

> 只要本地缓存（`cache.directory` 下的节点文件）存在，即使远端暂时不可用，服务仍可基于缓存继续工作，`/health` 的 `has_data` 会保持 `true`。

---

## 3. 模板未生效 / 请求提示模板错误

**症状**：请求返回 `400`，响应体形如 `Template Error: ...` 或 `Template 'xxx' not found or not enabled`；或返回的仍是旧模板内容。

**可能原因**：

- 请求的模板 ID 在配置 `templates` 中不存在。
- 该模板 `enabled: false`。
- 模板 URL 不可访问，且本地无缓存可降级。
- 模板缓存未过期，`/` 请求不会强制更新。

**排查与解决**：

1. 确认 URL 中的 `template`（或兜底的 `type`）参数值是配置 `templates` 下的**键名**（如 `default`、`ios`），不是显示名称 `name`。
2. 确认该模板配置为 `enabled: true`。
3. 用 `/health` 查看已加载模板数量：
   ```bash
   curl "http://<host>:<port>/health"
   ```
   `has_template: false` 或 `template_count` 偏小时说明模板加载有问题。
4. 检查模板 URL 是否可访问。
5. 模板按 `update_interval`（秒，默认 1 小时）判断缓存是否过期；若想立即更新：
   - 调用 `/refresh?password=<密码>` 强制重拉全部模板；
   - 或请求主接口时加 `&refresh=1`（仅更新当前请求涉及的节点与全部启用模板）。
6. 若本地已存在模板缓存，远端下载失败时服务会降级使用本地缓存，此时返回内容可能不是最新，属预期行为。

---

## 4. 认证失败（返回 Password Error）

**症状**：请求返回 `401`，响应体为 `Password Error`。

**可能原因**：

- URL 中 `password` 参数与配置 `auth.password` 不一致。
- 密码含 `&`、`=`、空格、`#` 等特殊字符但未做 URL 编码，导致参数被截断或改变。
- 部署时修改了配置但未重启，或环境变量 `PASSWORD` 覆盖了预期值。

**排查与解决**：

1. 核对 URL 中的 `password` 与配置 `auth.password` 是否完全一致（区分大小写）。
2. 密码含特殊字符时先做 URL 编码，例如 `p@ss&word` 应写成 `p%40ss%26word`（或用 `curl --data-urlencode` 构造）。
3. 检查环境变量：`PASSWORD` 会覆盖配置文件中的 `auth.password`。
4. 修改密码后需重启服务生效。
5. 可用 `/health` 验证连通性（该接口无需鉴权），以区分“网络问题”与“鉴权问题”。

---

## 5. Docker 容器无法访问

**症状**：容器处于运行状态，但从宿主机或外部无法访问服务。

**可能原因**：

- 端口映射错误或未映射。
- 容器启动后崩溃并自动重启，或健康状态异常。
- 容器网络/防火墙策略限制。

**排查与解决**：

1. 确认容器在运行：`docker ps`，观察 `STATUS` 是否为 `Up`。
2. 确认端口映射正确，宿主端口指向容器内 `9000`（或你配置的 `server.port`）：
   ```bash
   docker ps --format "table {{.Names}}\t{{.Ports}}"
   ```
   预期看到形如 `0.0.0.0:<宿主端口>->9000/tcp`。
3. 查看容器日志定位启动问题：
   ```bash
   docker logs <容器名>
   ```
4. 配置与缓存目录务必通过 `-v` 挂载，否则容器重建后配置丢失：
   ```bash
   docker run -d \
     -p 9000:9000 \
     -v /path/to/config:/singbox-subscribe-convert/config \
     -v /path/to/storage:/singbox-subscribe-convert/storage \
     <镜像名>
   ```
5. 检查宿主机防火墙与云安全组是否放行对应端口。
6. 若修改了 `server.port`，需同步调整容器端口映射的右侧端口。

---

## 6. 更新不生效（修改后仍是旧内容）

**症状**：改了配置或节点/模板后，请求返回的内容没有变化。

**可能原因**：

- 修改的是 `config.yaml`，而服务**不会**监听配置文件变更——配置只在启动时加载，需重启才生效。
- 节点/模板的本地缓存文件尚未过期，主接口 `/` 在缓存有效期内不会重新拉取。
- 远端 CDN（如 Cloudflare）缓存了旧响应，或客户端本地缓存了订阅。
- 修改的是模板缓存文件，但写入了错误路径。

**排查与解决**：

1. **区分两类“修改”**：
   - 修改 `config.yaml`：必须重启服务（程序不监控配置文件）。
   - 修改节点/模板**内容**：服务通过 fsnotify 监控缓存目录（`cache.directory`），节点文件或 `template_<名>.json` 被写入后约 1 秒内自动重载，无需重启。
2. 手动触发刷新：
   ```bash
   curl "http://<host>:<port>/refresh?password=<密码>"
   ```
   该接口会清空内存与磁盘缓存并重新拉取全部启用模板与节点。
3. 请求主接口时带 `&refresh=1` 可让单次请求先更新再返回。
4. 确认写入的缓存路径正确：
   - 节点文件：`cache.directory` + `cache.node_file`；
   - 模板文件：`cache.directory` + `template_<模板名>.json`。
5. 若服务前置了 CDN：响应带 `Cache-Control: no-cache, no-store, must-revalidate` 等防缓存头，但 CDN 仍可能缓存；启用并调用 Cloudflare 清理（见第 7 节）或手动刷新 CDN。

---

## 7. Cloudflare 缓存清理失败

`/refresh` 仅在 `cloudflare.enabled: true` 时执行 Cloudflare 缓存清理。相关错误会出现在 `/refresh` 返回的 `errors` 列表中，或日志中 `cloudflare cache purge: ...` 行。

### 7.1 返回缺少认证头错误

**症状**：日志或错误信息提示缺少 `X-Auth-Key`、`X-Auth-Email` 或 `Authorization` 头；或提示 `cloudflare authentication not configured`。

**可能原因**：未配置任何认证方式。

**排查与解决**：在 `config.yaml` 的 `cloudflare` 下配置以下任一方式，然后重启：

```yaml
cloudflare:
  enabled: true
  purge_url: "https://api.cloudflare.com/client/v4/zones/<ZONE_ID>/purge_cache"
  # 方式一（推荐）：API Token
  api_token: "your_api_token"
  api_key: ""
  api_email: ""
```

```yaml
cloudflare:
  enabled: true
  purge_url: "https://api.cloudflare.com/client/v4/zones/<ZONE_ID>/purge_cache"
  # 方式二：Global API Key + 账户邮箱
  api_token: ""
  api_key: "your_global_api_key"
  api_email: "your_email@example.com"
```

代码按以下优先级选择认证头：

- 配置了 `api_token` → 请求头 `Authorization: Bearer <token>`；
- 否则同时配置了 `api_key` 与 `api_email` → 请求头 `X-Auth-Key` + `X-Auth-Email`；
- 两者都缺 → 报错 `cloudflare authentication not configured: either api_token or (api_key + api_email) is required`。

### 7.2 返回 Authentication error 或 403

**症状**：`/refresh` 返回 `cloudflare API returned status 403: ...` 或类似认证错误。

**可能原因**：

- API Token 权限不足或已过期。
- Zone Resources 未包含目标域名。
- Global API Key 对应的邮箱填写错误。

**排查与解决**：

1. 检查 API Token 是否有效、未过期。
2. 确认 Token 权限包含 **Zone → Cache Purge → Purge**，并在 Zone Resources 中选择了目标域名。
3. 使用 Global API Key 时，确认 `api_email` 为账户邮箱。
4. 核对 `purge_url` 中的 `<ZONE_ID>` 是否正确（Cloudflare 域名概览页可查看 Zone ID）。

### 7.3 Cloudflare 清理超时

**症状**：`/refresh` 报超时或长时间无响应。

**可能原因**：

- 服务器到 Cloudflare API 网络不稳定。
- 请求超时设置过短。

**排查与解决**：

1. 检查到 `api.cloudflare.com` 的网络连通性。
2. 调大 `subscription.timeout`（秒，默认 30），该值同时用于 Cloudflare 清理请求的超时。
3. 查看 debug 日志中的请求 URL 与响应详情。

### 7.4 需要查看 Cloudflare 清理详细日志

**排查与解决**：将日志级别改为 `debug` 并重启服务：

```yaml
logging:
  level: "debug"
```

重启后 `/refresh` 会打印完整的 Cloudflare 请求体、响应体与状态码，便于定位问题。

---

## 附：快速定位流程

| 现象 | 第一步动作 |
|------|-----------|
| 启动即退出 | 看启动输出 + 日志中 `Error loading config`/`validate config error` |
| 返回无节点 | `GET /health` 看 `has_data`，再 `/refresh` 看错误 |
| 模板报错 | `GET /health` 看 `has_template`/`template_count`，核对模板键名与 `enabled` |
| `Password Error` | 核对 `password` 参数与 `auth.password`（含 URL 编码） |
| 内容不更新 | 区分改的是 `config.yaml`（需重启）还是缓存文件（自动重载） |
| Cloudflare 失败 | 核对 `enabled`/`purge_url`/`api_token` 或 `api_key`+`api_email`，开 debug 日志 |
