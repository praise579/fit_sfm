# 文档总览

本项目（[sing-box 订阅转换服务](../README.md)）的全部文档。简体中文撰写；代码、命令、配置项与模板变量保留原文。

## 使用 / 部署

| 文档 | 内容 |
|---|---|
| [安装与部署](guide/installation.md) | 系统要求、Docker / Docker Compose / 从源码构建、交叉编译 |
| [配置](guide/configuration.md) | 配置文件位置与加载顺序、各区块（server/auth/subscription/templates/cache/cloudflare/logging）说明；字段级说明见 [`config/config.yaml`](../config/config.yaml)（唯一事实源） |
| [使用方法](guide/usage.md) | CLI 参数、环境变量覆盖、运行产物目录 |
| [模板编写](guide/templates.md) | pongo2 渲染、`{{ Nodes }}`、`{{ "关键词" | NotesName }}`、多模板与多版本客户端 |
| [HTTP API](guide/api.md) | `/`、`/health`、`/refresh` 等接口与 Cloudflare 缓存清理 |
| [常见问题](guide/troubleshooting.md) | 症状 → 原因 → 解决 |

## 开发

| 文档 | 内容 |
|---|---|
| [架构说明](development/architecture.md) | 调用链、模块职责、数据流与缓存、热更机制 |
| [贡献指南](../CONTRIBUTING.md) | 构建 / 测试 / 提 PR 约定 |
| [决策记录](decisions.md) | 关键技术决策的「背景 / 决策 / 后果」 |

## 其它

- 面向 AI / 协作者的项目速览：[`CLAUDE.md`](../CLAUDE.md)
- 变更记录：[`CHANGELOG.md`](../CHANGELOG.md)
- 许可协议：[`LICENSE`](../LICENSE)（Apache-2.0）
