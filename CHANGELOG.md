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
