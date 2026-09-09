# 贡献指南

欢迎为本项目提交 Issue 与 PR。请先阅读 [README](README.md) 与 [文档索引](docs/README.md)。

## 环境要求

- Go 1.24+
- （可选）Docker，用于容器化验证

## 本地开发

```bash
git clone https://github.com/praise579/fit_sfm.git && cd fit_sfm

go build ./...                 # 编译
go test ./...                  # 测试（与 make test 等价）
go vet ./...                   # 静态检查

# 本地运行（需先准备一份配置，示例见 config/config.yaml）
go run . run -c config/config.yaml
```

非平凡逻辑（分支/循环/解析/转换等）请补一条最小单测（stdlib `testing` 即可），确保改动有可跑验证。

## 提交 PR 前

- [ ] `go vet ./...` 与 `go test ./...` 通过
- [ ] 遵循 Go 官方格式（`gofmt` / `goimports`）
- [ ] 代码注释与文档使用简体中文；代码标识符、命令、配置项保留原文
- [ ] 若改变配置字段 / HTTP 接口 / 架构：
  - 更新 `config/config.yaml` 注释（它是配置字段的唯一事实源，SSOT）；
  - 同步 `docs/guide/` 对应页面（见 [docs/README.md](docs/README.md)）；
  - 用户可见行为变更记入 [CHANGELOG.md](CHANGELOG.md)；
  - 对难以逆转的设计取舍，在 [docs/decisions.md](docs/decisions.md) 追加「背景 / 决策 / 后果」。
- [ ] 若修改了引用外部库或他人代码，注意许可头与仓库 Apache-2.0 的兼容性（本仓库不接纳 GPL 传染代码）。

## 提交流程

1. 先开 Issue 说明动机与方案（Bug / 功能皆可）。
2. 基于最新 `master` 切分支：`git checkout -b feat/xxx`。
3. 提交并推送后创建 Pull Request，关联对应 Issue。
4. 保持 PR 聚焦单一变更，方便 review 与回滚。

## 目录速览

各包职责见 [architecture.md](docs/development/architecture.md)。`internal/` 放核心业务，`pkg/` 放可复用工具，`global/` 放跨包共享的配置与环境信息。
