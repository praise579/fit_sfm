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

### 脚本整理

- 删除 4 个与本项目无关的残留脚本：`start.sh`、`start_console.sh`、`stop.sh`、`docker_redeploy.sh`。它们随上游快照一次性导入，全仓库零引用，且内含他项目标识符（`goapi_starfission`、`./new/apiRun`、端口 `8000-8002`、`configs/`、镜像 `xxx/xxxxx`），与本项目实际取值（`sb-sub-c`、端口 `9000`、`config/`）完全不符。
- 在用脚本收归 `scripts/`：`entrypoint.sh`、`docker_image_clean.sh`（均经 `git mv`，保留历史）。
- `Dockerfile` 的 `COPY entrypoint.sh` 同步为 `COPY scripts/entrypoint.sh`；`Makefile` 内调用路径同步为 `scripts/docker_image_clean.sh`。
- `scripts/docker_image_clean.sh` 增加一行工作目录锚定（`cd "$(dirname "$0")/.."`）：该脚本以「当前目录名」作为待清理的项目名，搬入子目录后若从 `scripts/` 直接执行会认错项目，此行保持行为等价。
- 新增 `make docker-clean` 转发目标，使镜像清理可独立调用（此前仅被 `push-online`/`push-dev` 间接调用）。
- 修复 `push-online` / `push-dev` 依赖的不存在目标 `build-linux` → `build-linux-amd64`。该断链使这两个目标此前直接报 *No rule to make target*，`docker_image_clean.sh` 的唯一调用路径也因此从未跑通。

### Makefile 重构

- 按「构建 / 开发 / 本地 Docker / 发布」重新分节，每个目标补一行说明；删除全部历史遗留注释、他项目残留块（`build2`）、调用不存在 define 的 `# $(call checkStatic)`。
- 删除死代码：`include .env`、恒同分支的 `platform` 死 `ifeq`、未被引用的 `goClean` / `goGet` / `cfgDir` / `cfgFile`、`build-macos-amd64` 中未定义的 `$(bin)`，以及仅打印一行横幅的 `init` define。
- 修复 `run` 的 `$(goRun)-v`：少一个空格导致 `-v` 被粘进 ldflags 参数内部，既污染 `BuildTime` 的注入值，又使 `-v` 从未真正生效。
- `GitTag` 取值加 `--always` 兜底：仓库无 tag 时 `global.Version` 不再注入空串，退化为短 hash（`push-release` 的镜像 tag 随之从空的 `release-v` 变为带短 hash）。
- 新增本地 Docker 目标：`docker-build` / `docker-up` / `docker-down` / `docker-rebuild` / `docker-logs`。本地镜像为 `linux/arm64`，端口取 `LocalPort`（默认 1900，对齐本地 `config-docker.yaml` 的 `server.port`）；取舍理由见 [decisions.md](docs/decisions.md)。
- `push-online` 更名为 `push-release`（语义即「正式发布」），`push-dev` 不变；文档同步。
- 新增 `.dockerignore`：构建上下文排除 `.git` / `.ua` / `storage` / `config.yaml` 与三个非 Linux 平台产物，**保留 `build/linux_*`**（Dockerfile COPY 的来源，不可排除）。
- 删除被 git 跟踪却无人消费的 `.env`（原仅含一个拼错的、无消费者的 `RUNTIME_ENVIROMENT`）及 `.gitignore` 中空转的 `!.env` 规则。

### 修复

- 修复本地 Docker 容器端口访问不到：`docker-up` 把宿主 `config.yaml` 挂成了容器内 `/singbox-subscribe-convert/config-docker.yaml`，而该文件名不在程序的配置查找列表（`config/config-dev.yaml` → `config.yaml` → `config/config.yaml`）内，容器因此**静默回落内嵌默认配置**（`server.port: 9000`、示例订阅地址），与映射出的 `LocalPort`（1900）对不上，表现为 1900 端口无人监听。现改为挂载仓库根 `config-docker.yaml` 到 `/singbox-subscribe-convert/config/config.yaml`，与服务器部署的 `config/` 目录约定一致。
- `docker-up` 增加宿主文件存在性校验：bind mount 源缺失时 Docker 会静默创建目录，使容器内配置落点变成目录，故障更难定位。

