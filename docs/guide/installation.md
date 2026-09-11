# 安装部署指南

本项目（fit_sfm，Singbox Subscribe Convert）是一个 sing-box 订阅转换服务，支持多模板、多版本客户端。本文介绍三种部署方式：

- Docker 部署（推荐）
- Docker Compose 部署（推荐，含自动更新辅助）
- 从源码编译运行

> 命令中的可执行文件名 `singbox-subscribe-convert` 来自 `go build -o singbox-subscribe-convert .` 的编译产物；若使用仓库 Makefile 交叉编译或预编译产物，文件名为 `sb-sub-c`，替换即可。
>
> 由于 Docker 镜像的发布策略尚未定稿，本文不写死具体镜像名。涉及镜像的示例统一使用 `<your-image>:<tag>` 占位，实际镜像名以仓库 Release 或说明文档为准。

## 系统要求

| 资源 | 要求 |
|------|------|
| Go | 1.24.1 或更高版本（源码编译） |
| Docker | 20.10+，以及 Docker Compose 2.0+（Docker 部署） |
| 内存 | 至少 100MB 可用内存 |
| 磁盘 | 至少 50MB 可用磁盘空间 |

## Docker 部署

仓库根目录自带 `Dockerfile` 与 `docker-compose.yaml`。容器内工作目录为 `/singbox-subscribe-convert/`，程序会在此目录下自动查找配置文件（详见 [usage.md](usage.md) 的「配置加载顺序」）。

### 1. 准备宿主机数据目录与配置文件

容器将以下两个目录作为卷，分别存放配置与运行数据：

| 容器内路径 | 用途 |
|-----------|------|
| `/singbox-subscribe-convert/config/` | 配置文件（`config.yaml`） |
| `/singbox-subscribe-convert/storage/` | 缓存、节点、日志 |

在宿主机上创建对应目录：

```bash
mkdir -p /data/singbox-subscribe-convert/config
mkdir -p /data/singbox-subscribe-convert/storage
```

把修改好的 `config.yaml` 放到 `/data/singbox-subscribe-convert/config/config.yaml`。推荐让容器内外的配置路径保持一致，便于排查（配置文件的准备与修改见下文「从源码编译运行」）。

### 2. 运行容器

使用预构建镜像（镜像名以仓库实际发布为准）：

```bash
docker run -d \
  --name singbox-subscribe-convert \
  -p 9000:9000 \
  -v /data/singbox-subscribe-convert/storage:/singbox-subscribe-convert/storage/ \
  -v /data/singbox-subscribe-convert/config:/singbox-subscribe-convert/config/ \
  <your-image>:<tag>
```

停止并删除容器：

```bash
docker stop singbox-subscribe-convert
docker rm singbox-subscribe-convert
```

### 3. 使用 Docker Compose（推荐）

仓库自带的 `docker-compose.yaml` 定义了两个服务：

| 服务 | 说明 |
|------|------|
| `singbox-subscribe-convert` | 主服务，映射端口 `9000` |
| `watchtower` | 镜像更新辅助容器，按 `WATCHTOWER_SCHEDULE`（每半小时）检查带有 `com.centurylinklabs.watchtower.enable=true` 标签的容器 |

使用前请按实际镜像情况调整 `docker-compose.yaml` 中主服务的 `image` 字段：

- 若使用已发布的预构建镜像，填写实际镜像名与标签；
- 若尚未发布或想用本地镜像，可改为使用下文「从 Dockerfile 构建镜像」得到的镜像，或将 `image` 替换为 `build: .` 指向仓库 Dockerfile。

目录挂载说明：

| 宿主机路径 | 容器内路径 | 说明 |
|-----------|-----------|------|
| `/data/singbox-subscribe-convert/storage/` | `/singbox-subscribe-convert/storage/` | 缓存、节点、日志 |
| `/data/singbox-subscribe-convert/config/` | `/singbox-subscribe-convert/config/` | 配置文件 |

启动与管理：

```bash
# 启动所有服务（主服务 + watchtower）
docker-compose up -d

# 查看运行状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止并移除
docker-compose down
```

> 注意：`docker-compose.yaml` 中 `watchtower` 默认配置为 `WATCHTOWER_MONITOR_ONLY=true`（仅监控、不自动更新），如需自动更新可将其改为 `false`。

常用排查命令：

```bash
# 查看容器实时日志
docker logs -f singbox-subscribe-convert

# 进入容器调试
docker exec -it singbox-subscribe-convert /bin/sh
```

容器内的日志默认位于 `/singbox-subscribe-convert/storage/logs/c.log`。

### 4. 从 Dockerfile 构建镜像（可选）

Dockerfile 需要先准备对应平台已编译的二进制，再从仓库根目录构建：

```bash
# 先编译 Linux 二进制（产物位于 build/linux_amd64、build/linux_arm64）
make build-linux-amd64
make build-linux-arm64

# 构建镜像（-t 后的镜像名按需填写）
docker build --platform linux/amd64 -t <your-image>:<tag> -f Dockerfile .
docker build --platform linux/arm64 -t <your-image>:<tag> -f Dockerfile .
```

运行自建镜像：

```bash
docker run -d \
  --name singbox-subscribe-convert \
  -p 9000:9000 \
  -v /data/singbox-subscribe-convert/storage:/singbox-subscribe-convert/storage/ \
  -v /data/singbox-subscribe-convert/config:/singbox-subscribe-convert/config/ \
  <your-image>:<tag>
```

上述手工步骤在仓库 `Makefile` 中有对应目标，可按需取用：`make docker-build`（构建本地镜像）、`make docker-up` / `docker-down` / `docker-logs` / `docker-rebuild`（起停与查看）、`make docker-clean`（清理镜像）。

发布镜像到 Docker Hub 用 `make push-release`（正式）或 `make push-dev`（开发）。两者的镜像名与 tag 前缀在 `Makefile` 顶部的 `DockerHubUser` / `DockerHubName` / `ReleaseTagPre` / `DevelopTagPre` 中定义，**不需要** `.env`。

### 5. 本地开发：从源码构建并运行容器

改完代码要在容器里验证时，用这一组目标。它们**只在本机操作，不涉及任何远程仓库**：

```bash
make docker-build     # 编译 linux/arm64 并构建本地镜像 singbox-subscribe-convert:local
make docker-up        # 后台起容器 sb-sub-c-local，挂载仓库根 config-docker.yaml 与 storage/
make docker-logs      # 跟踪日志
make docker-down      # 停并删除容器
make docker-rebuild   # 一键：停容器 → 重编译重打镜像 → 重新起
```

两点约定：

- **本地镜像构建为 `linux/arm64`**（Apple Silicon 原生执行，免 QEMU 模拟）；面向发布的 `push-*` 仍为 `linux/amd64`，amd64 的回归验证放在推送前完成。取舍理由见 [decisions.md](../decisions.md)。
- **端口由 `Makefile` 的 `LocalPort` 决定（默认 1900）**，须与所挂载 `config-docker.yaml` 的 `server.port` 一致，否则映射出来的端口无人监听。
- **容器内的落点固定为 `/singbox-subscribe-convert/config/config.yaml`**，与服务器部署的 `config/` 目录约定一致。程序只按固定顺序查找 `config/config-dev.yaml` → `config.yaml` → `config/config.yaml`，挂成别的文件名不会被读取，而是**静默回落内嵌默认配置**（端口 9000、示例订阅地址），表现为映射出去的端口访问不到。

> 本地开发**不使用**仓库根的 `docker-compose.yaml` —— 那份是给服务器部署配合 watchtower 用的，改它会影响线上。

## 从源码编译运行

### 1. 克隆仓库

```bash
git clone https://github.com/praise579/fit_sfm.git
cd fit_sfm
```

### 2. 编译

在项目根目录直接编译本机平台的可执行文件：

```bash
go build -o singbox-subscribe-convert .
```

> 若编译时提示网络错误，可先配置 Go 模块代理再重试：
>
> ```bash
> go env -w GOPROXY=https://goproxy.cn,direct
> # 或
> go env -w GOPROXY=https://goproxy.io,direct
> ```

如需交叉编译，可使用仓库 `Makefile` 提供的目标（产物位于 `build/<平台>/`，文件名为 `sb-sub-c`）：

```bash
make build-all              # 编译所有平台
make build-linux-amd64      # Linux AMD64
make build-linux-arm64      # Linux ARM64
make build-macos-amd64      # macOS Intel
make build-macos-arm64      # macOS Apple Silicon
make build-windows-amd64    # Windows AMD64
```

### 3. 创建并修改配置文件

项目自带示例配置 `config/config.yaml`。复制一份作为自定义配置（保留原文件便于以后恢复）：

```bash
cp config/config.yaml config/my-config.yaml
```

用任意编辑器打开 `config/my-config.yaml`，**至少需要修改以下两项**：

| 配置项 | 修改为 |
|--------|--------|
| `auth.password` | 一个你自己知道的密码，例如 `mysecret123` |
| `subscription.url` | 你从服务商获取的节点订阅地址 |

如需多模板，可仿照 `config/config.yaml` 中已有的 `templates` 结构，为不同客户端（如 iOS、OpenWrt 等不同 sing-box 版本）配置各自的模板 URL 与名称，并设置 `default_template`。

### 4. 运行

```bash
# 使用指定的配置文件运行
./singbox-subscribe-convert run -c config/my-config.yaml
```

更完整的命令行选项、环境变量覆盖与配置加载顺序，参见 [usage.md](usage.md)。

> 提示：若不想自行编译，也可以关注仓库 GitHub Releases 页面发布的预编译二进制（支持 Linux / macOS / Windows 多平台）。
