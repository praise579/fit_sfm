# 模板指南

模板（template）是 fit_sfm 的核心概念：节点订阅提供「相同的一批节点」，模板提供「面向不同客户端 / 不同 sing-box 版本的配置骨架」，二者在请求时合成最终的 sing-box 配置。这样一份订阅源可以适配 OpenWrt、iOS、Android 等多端多版本，解决不同 sing-box 版本（例如 1.12 起对旧版本不兼容）带来的配置碎片化问题。

模板本身是一个文本文件（内容基本是 JSON），内嵌少量 pongo2 占位符。模板的 URL 在 `config.yaml` 的 `templates.<ID>.url` 中配置，配置方式见 [configuration.md](configuration.md) 的 templates 区块；多模板如何服务多客户端/多版本，见下文「多模板配置与多版本客户端」。

> 以下示例并非纯 JSON——`{{ ... }}` 是 pongo2 占位符，**渲染完成后**输出才是合法 JSON。

## 渲染机制简述

- 服务为每个启用的模板维护一个 pongo2 模板实例。模板经 `url` 拉取后缓存在本地（`template_<模板ID>.json`），请求到达时用当前内存中的节点数据渲染，结果以 `application/json` 返回。
- 若本地还没有缓存文件、或缓存已过期（超过该模板的 `update_interval`，默认 1 小时），服务会先自动下载再渲染；下载失败但本地已有缓存时降级使用旧缓存。
- 触发渲染的请求为 `GET /?password=<密码>&template=<模板ID>`；不携带 `template` 时使用配置的 `default_template`。
- 模板可使用两个核心渲染元素：
  - 变量 `{{ Nodes }}`：把订阅的全部节点完整 JSON 插入模板；
  - 过滤器 `{{ "关键词" | NotesName }}`：按关键词筛选节点名称，用于生成选择器（selector / urltest）的节点列表。

## 模板变量 `{{ Nodes }}`

**作用**：将订阅中的所有节点以完整配置（JSON 对象，含服务器地址、端口、加密/认证方式等，其 `tag` 即节点名）插入到模板所在位置。多个节点以 `,` 连接，可直接作为 `outbounds` 数组的数组元素。

**典型位置**：放在 `outbounds` 数组末尾。注意它前面已写的最后一个元素之后要保留逗号：

```json
{
  "outbounds": [
    {
      "tag": "🚀 节点选择",
      "type": "selector",
      "outbounds": ["DIRECT"]
    },
    { "tag": "DIRECT", "type": "direct" },

    {{ Nodes }}
  ]
}
```

渲染后 `{{ Nodes }}` 会展开为类似下面的内容（示意），从而把整批节点作为出站加入配置：

```json
{ "type": "shadowsocks", "tag": "🇭🇰 香港 01", "server": "...", ... },
{ "type": "vmess", "tag": "日本专线", "server": "...", ... }
```

## 过滤器 `{{ "关键词" | NotesName }}`

**作用**：按关键词筛选节点**名称**（即节点 `tag`），返回一个由名称字符串组成的 JSON 数组（去掉了外层 `[ ]`，便于直接写在 `[ ... ]` 中间），用于填充 selector / urltest 等出站组的 `outbounds` 列表。

**基本语法**：

```json
"outbounds": [ {{ "关键词" | NotesName }} ]
```

**匹配语义**（与 `internal/handler` 中过滤器实现一致）：

- 关键词为节点名的**子串模糊匹配**（`strings.Contains`），例如 `"香港"` 可命中 `🇭🇰 香港 01`、`香港-IPLC`。
- **大小写敏感**：关键词需与实际节点名称保持一致，例如 `"HK"` 不会命中 `hk`。
- **多关键词用 `|` 分隔，逻辑为 OR**：`"香港|新加坡"` 返回名称含「香港」**或**「新加坡」的节点。
- 关键词会先去除首尾空格；使用空字符串 `""` 表示不过滤，返回**全部**节点名称。
- 同一节点即使命中多个关键词也只会出现一次（按 `tag` 去重）。
- **无匹配兜底**：筛选结果为空时，自动补入 `no_node` 值（取值来自 `default_template` 所指向模板的 `no_node`），保证 `outbounds` 非空、选择器可用。⚠️ **该值必须是当前渲染模板内真实存在的出站 tag**（示例为 `DIRECT` 直连出口），否则会引用不存在的出站、sing-box 校验失败。

### 场景示例

**场景 1：获取全部节点（手动切换）**

```json
{
  "tag": "🐸 手动切换",
  "type": "selector",
  "outbounds": [ {{ "" | NotesName }} ]
}
```

**场景 2：筛选特定地区节点**

```json
{
  "tag": "🇭🇰 香港节点",
  "type": "selector",
  "outbounds": [ {{ "香港" | NotesName }} ]
}
```

只返回名称含「香港」的节点，如 `🇭🇰 香港 01`、`香港专线`。

**场景 3：筛选多个地区（OR 逻辑）**

```json
{
  "tag": "🇭🇰 港新节点",
  "type": "selector",
  "outbounds": [ {{ "香港|新加坡" | NotesName }} ]
}
```

**场景 4：配合自动测速**

```json
{
  "tag": "♻️ 港新自动",
  "type": "urltest",
  "outbounds": [ {{ "香港|新加坡" | NotesName }} ],
  "url": "http://www.gstatic.com/generate_204",
  "interval": "10m",
  "tolerance": 50
}
```

`NotesName` 返回的名称与 `{{ Nodes }}` 插入节点对象的 `tag` 完全一致，因此可放心作为选择器的 `outbounds` 引用；若节点源中正好有与分组 tag 同名的节点，注意避免命名冲突。

## 完整模板示例

一个较完整的模板骨架（把「手动切换」「自动测速」「按地区分组」组合起来，最后用 `{{ Nodes }}` 注入全部节点）：

```json
{
  "outbounds": [
    {
      "tag": "🚀 节点选择",
      "type": "selector",
      "outbounds": ["🐸 手动切换", "♻️ 自动选择", "🇭🇰 香港节点", "🇯🇵 日本节点", "DIRECT"]
    },
    {
      "tag": "🐸 手动切换",
      "type": "selector",
      "outbounds": [ {{ "" | NotesName }} ]
    },
    {
      "tag": "♻️ 自动选择",
      "type": "urltest",
      "outbounds": [ {{ "" | NotesName }} ],
      "url": "http://www.gstatic.com/generate_204",
      "interval": "10m"
    },
    {
      "tag": "🇭🇰 香港节点",
      "type": "selector",
      "outbounds": [ {{ "香港" | NotesName }} ]
    },
    {
      "tag": "🇯🇵 日本节点",
      "type": "selector",
      "outbounds": [ {{ "日本" | NotesName }} ]
    },
    { "tag": "DIRECT", "type": "direct" },

    {{ Nodes }}
  ]
}
```

## 多模板配置与多版本客户端

同一批节点、不同客户端版本 → 配置多个模板、各自访问。模板源 `http(s)://` 远程与 `file://` 本地可**在同一列表混排**，仓库自带官方模板即以本地 `file://` 加入（需在仓库根运行才命中）：

```yaml
templates:
  # 默认模板 default：仓库自带本地官方模板（macOS + sing-box 1.14），
  # file:// 指向仓库根 templates/sing-box-macos-1.14.json，随 git 版本化
  default:
    url: "file://templates/sing-box-macos-1.14.json"
    name: "macOS 精简版"
    no_node: "DIRECT"
    enabled: true

  # openwrt / ios：远程模板源示例，与 default 混排在同一个列表
  openwrt:
    url: "https://example.com/templates/openwrt.json"
    name: "OpenWRT"
    no_node: "DIRECT"
    enabled: true

  ios:
    url: "https://example.com/templates/ios.json"
    name: "IOS"
    no_node: "DIRECT"
    enabled: true

default_template: "default"
```

对应得到不同访问地址（节点内容一致，只是格式按模板区分）：

```bash
# 默认模板（default_template 指向 default，即本地官方 macOS / sing-box 1.14）
curl "https://your-host/?password=xxx"

# 显式指定本地官方模板
curl "https://your-host/?password=xxx&template=default"

# 显式指定远程 OpenWRT 模板
curl "https://your-host/?password=xxx&template=openwrt"

# 显式指定远程 iOS 模板
curl "https://your-host/?password=xxx&template=ios"
```

模板特性：

- **独立缓存**：每个模板有独立的本地缓存文件（`template_<模板ID>.json`），互不影响。
- **并行更新**：刷新时多个模板并发拉取，提高效率。
- **动态加载**：可通过 `enabled` 启用 / 禁用模板，禁用后请求该模板返回 `Template ... not found or not enabled`。
- **热重载**：模板缓存文件变化后自动重新加载，无需重启服务。

## 使用提示

- **占位符摆放**：`{{ Nodes }}` 应放在 JSON 数组元素的位置（通常为 `outbounds` 末尾），并保证前面元素与它之间保留逗号；`{{ "..." | NotesName }}` 应放在某个出站组 `outbounds` 的 `[ ]` 内部。
- **命名一致**：`NotesName` 的大小写敏感子串匹配以节点 `tag` 为准，建议模板中的地区词与节点命名保持同一套习惯（如都用中文或都用 `HK`）。
- **多关键词**：用 `|` 组合多种命名，例如 `"香港|HK|Hong Kong"` 可兼容多种命名方式。
- **空结果兜底**：筛选不到节点时自动使用 `no_node` 直连出口，模板仍可用；若需要不同兜底出口，调整模板配置的 `no_node`（兜底取 `default_template` 所指向模板的 `no_node`；该值必须指向模板内真实存在的出站 tag，官方模板用 `DIRECT`）。
- **验证与排错**：把模板渲染后的输出存成文件并校验 JSON 合法性可快速定位逗号或引号问题；将 `logging.level` 设为 `debug` 可观察节点 / 模板加载与渲染报错详情。
- **模板更新**：修改模板文件内容后，调用 `GET /refresh?password=<密码>` 手动拉取更新，或等待其 `update_interval`（默认 1 小时）过期后自动更新。
