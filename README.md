<!-- 徽章待发布策略确定后补充（CI/Docker/GitHub Release 地址未写死前不加链接） -->

# Singbox Subscribe Convert

**sing-box 订阅转换服务**：周期抓取远程订阅节点，用多份 sing-box 模板渲染出适配不同客户端/版本的配置文件，通过 HTTP 暴露给客户端订阅。

把**一份订阅、多种客户端**的重复劳动交给一个常驻服务：

- 一次接入多个订阅源与多份模板（OpenWRT / iOS / 不同 sing-box 版本各配一份）；
- 客户端订阅同一 URL，指定模板即得对应版本配置；
- 节点与模板本地缓存 + 定时刷新 + 文件变更热更，无需手动干预；
- 内置节点名筛选、健康检查、手动刷新接口与可选的 Cloudflare 缓存清理。

## 快速开始

以下为**源码运行**的最小示例（Docker / Docker Compose / 交叉编译见 [安装与部署](docs/guide/installation.md)）：

```bash
# 1. 获取源码（需要 Go 1.24+）
git clone https://github.com/praise579/fit_sfm.git && cd fit_sfm

# 2. 编译
go build -o singbox-subscribe-convert .

# 3. 准备配置（从内嵌示例复制一份再改）
cp config/config.yaml config/my-config.yaml

# 4. 编辑 config/my-config.yaml，至少修改 auth.password 与 subscription.url

# 5. 启动
./singbox-subscribe-convert run -c config/my-config.yaml
```

启动后默认监听 `9000` 端口。用支持订阅的客户端填入转换接口地址即可：

```
http://<host>:9000/?password=<你的密码>
```

## 文档

| 主题 | 入口 |
|---|---|
| 安装与部署（Docker / Compose / 源码） | [docs/guide/installation.md](docs/guide/installation.md) |
| 配置说明（字段级见 [`config/config.yaml`](config/config.yaml)） | [docs/guide/configuration.md](docs/guide/configuration.md) |
| 命令行与运行 | [docs/guide/usage.md](docs/guide/usage.md) |
| 模板编写与多版本 | [docs/guide/templates.md](docs/guide/templates.md) |
| HTTP API 参考 | [docs/guide/api.md](docs/guide/api.md) |
| 常见问题 | [docs/guide/troubleshooting.md](docs/guide/troubleshooting.md) |
| 架构与开发 | [docs/development/architecture.md](docs/development/architecture.md) |

## 特性一览

- **订阅 + 模板分离**：节点订阅只拉一份，模板按客户端/版本各一份。
- **本地缓存与自动刷新**：`subscription.refresh_interval`、模板级 `update_interval` 周期刷新，支持 `file://` 本地协议。
- **模板变量**：`{{ Nodes }}` 插入全部节点；`{{ "关键词" | NotesName }}` 按名称筛选节点。
- **多模板**：请求可指定模板；`default_template` 兜底。
- **可观测与运维**：`/health` 健康检查、`/refresh` 手动刷新、可选 Cloudflare 缓存清理、优雅关闭。

## 上游与致谢

本项目 fork 自 [haierkeys/singbox-subscribe-convert](https://github.com/haierkeys/singbox-subscribe-convert)（[Apache-2.0](LICENSE)），继承其完整 git 历史，原作者为 **HaierKeys**，特此致谢。独立仓库无法通过 GitHub fork 网络自动关联，故此处手动声明来源。

相对上游，fit_sfm 的主要改动：

- 模块路径由 `github.com/haierkeys/singbox-subscribe-convert` 统一为 `github.com/praise579/fit_sfm`，同步 import 与构建注入；
- 文档体系重组为 README 登录页 + `docs/`（guide / development / decisions）；
- 依许可合规重写 `pkg/logger`、`pkg/safe_close`（原声明源自 GPL-3.0 上游，Apache 下不可沿用）；
- 官方本地模板库 `templates/` 及示例中的远程 ios/openwrt 模板源取自 [haierkeys/free-network-tool](https://github.com/haierkeys/free-network-tool)，一并致谢。

原版代码见上方上游仓库链接。

## 相关

- [贡献指南](CONTRIBUTING.md)　|　[变更记录](CHANGELOG.md)　|　[安全问题](SECURITY.md)
- 文档索引：[docs/](docs/README.md)
- 许可：[Apache-2.0](LICENSE)

## 免责声明

本项目仅供个人学习、技术研究之用。使用者因以任何形式使用本项目（包括但不限于部署、运行、修改、分发或将其接入其它服务）而造成的财产损失、数据泄露、法律风险及后果等，均由使用者个人自行承担，作者与各贡献者概不负责。请在使用前了解并遵守所在地相关法律法规。
