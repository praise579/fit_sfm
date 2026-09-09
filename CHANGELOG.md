# Changelog

本项目的显著变更记录，格式参照 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### 开源就绪（文档与工程校准）

- 模块路径 `github.com/haierkeys/singbox-subscribe-convert` 更名为 `github.com/praise579/fit_sfm`，同步全部 import 与构建注入。
- 许可合规：重写 `pkg/logger` 与 `pkg/safe_close`，去除源自 mosdns 的 GPL-3.0 版权头，仓库整体维持 Apache-2.0；补最小单测。
- 校准构建与 CI：
  - 修复 `make build-all` 引用的拼错目标 `build-winmdows-amd64` → `build-windows-amd64`；
  - `make test` 由占位改为真实执行 `go test ./...`；
  - 移除遗留死目标 `old-gen` / `gen`；
  - `.github/workflows/test.yml` 重写为 push/PR 触发的 vet + build + test。
- 文档体系重组：
  - 新增 `docs/`（guide / development / decisions）；
  - `README.md` 精简为登录页；原安装/配置/模板/API/FAQ 详述迁移至 `docs/guide/`；
  - `config/config.yaml` 注释补全为配置字段唯一事实源（SSOT）；
  - 新增 `CHANGELOG.md`、`CONTRIBUTING.md`、`SECURITY.md` 与 `.github` 协作模板；
  - 根 `CLAUDE.md` 重写并增加文档地图；
  - `.gitignore` 追加根 `config.yaml`、`.ua/`（本地开发覆盖配置不再有误提交风险）。
  - README 底部新增「上游与致谢」，声明 fork 自 `haierkeys/singbox-subscribe-convert`（Apache-2.0）并致谢模板源 `haierkeys/free-network-tool`。
  - README 末尾新增「免责声明」，提示本项目仅供个人学习研究、使用风险由使用者自行承担。

### 官方本地模板库

- 新增仓库根 `templates/`：随 git 版本化的官方模板 `sing-box-macos-1.14.json`（适用于 macOS + sing-box 1.14），改模板走 git 可评审、随发布分发。
- 示例 `config/config.yaml` 模板列表改为**远程 / 本地混排**：`default` 默认模板改为指向仓库内 `file://templates/sing-box-macos-1.14.json`（`default_template` 不变）；原 `key=default` 的远程 OpenWRT 模板改名 `openwrt` 保留；`ios`、`1.13-ios` 与 `subscription.url` 保持远程示例。
- `Dockerfile` 同步 `COPY templates/` 进镜像，容器默认配置开箱可用。
- 修正兜底标记：各模板示例 `no_node` 由 `🎯 全球直连` 改为 `DIRECT`（官方模板真实存在的直连出站 tag），并同步 `global/config.go` 的兜底默认值与全部文档；避免 `NotesName` 筛选无结果时渲染出引用不存在出站、sing-box 校验失败的配置。
- 说明：官方模板改动不即时热更（watcher 只监控缓存目录副本），`/refresh` 或到点定时自动拉取后生效。

