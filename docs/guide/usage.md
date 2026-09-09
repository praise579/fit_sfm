# 使用指南

本页介绍 `run` 子命令的命令行选项、环境变量覆盖、配置文件的加载顺序，以及运行时产生的目录与文件。

> 示例中的可执行文件名 `singbox-subscribe-convert` 为 `go build -o singbox-subscribe-convert .` 的产物；若使用 Makefile 交叉编译产物或预编译二进制，文件名是 `sb-sub-c`，替换即可。

## 命令行选项

查看帮助与版本：

```bash
./singbox-subscribe-convert --help

./singbox-subscribe-convert version
```

`run` 子命令选项：

| 选项 | 说明 |
|------|------|
| `-c, --config <path>` | 指定配置文件路径 |
| `-d, --dir <path>` | 指定工作目录（切换后再解析相对路径与查找配置） |
| `-p, --port <port>` | （预留）原意覆盖服务端口，当前版本尚未接线、不生效 |
| `-m, --mode <dev\|prod>` | （预留）原意指定运行模式，当前版本尚未接线、不生效 |

常用示例：

```bash
# 运行服务（使用默认配置查找逻辑，见「配置加载顺序」）
./singbox-subscribe-convert run

# 指定配置文件
./singbox-subscribe-convert run -c /path/to/config.yaml

# 指定工作目录
./singbox-subscribe-convert run -d /path/to/workdir
```

> 注：`-p` / `-m` 目前为预留 flag（尚未生效）。调整端口请直接修改配置文件 `server.port`。

启动成功后，日志会输出监听端口、加载的节点数 / 模板数等信息，例如：

```
Starting server ... port 9000
Initial fetch complete ... node_count 10 template_count 2
Server is running ... addr :9000
```

> 提示：终端关闭后服务会停止。如需后台运行，可使用 `nohup ./singbox-subscribe-convert run -c config/my-config.yaml > output.log 2>&1 &`，或使用 `screen` / 系统守护进程（如 macOS 的 launchd）托管。

## 配置加载顺序

`run` 启动时会按以下优先级查找配置文件（优先级从高到低）：

| 优先级 | 路径 / 参数 | 说明 |
|--------|-------------|------|
| 1 | `-c` / `--config` 指定的路径 | 显式指定，最高优先级 |
| 2 | `config/config-dev.yaml` | 相对工作目录 |
| 3 | `config.yaml` | 相对工作目录（仓库根目录下的配置文件） |
| 4 | `config/config.yaml` | 相对工作目录，默认配置 |

说明：

- 上述相对路径均以运行工作目录为基准，可用 `-d` 切换。
- 若通过 `-c` 指定的配置文件不存在，程序会给出 Warning，然后继续按默认顺序查找。
- 若以上文件都不存在，程序会在工作目录下自动创建默认的 `config/config.yaml`，并打印提示，要求填写密码、节点订阅地址与模板地址后再启动。
- Docker 容器内的工作目录固定为 `/singbox-subscribe-convert/`，同样按上述顺序查找（即依次为 `/singbox-subscribe-convert/config/config-dev.yaml`、`/singbox-subscribe-convert/config.yaml`、`/singbox-subscribe-convert/config/config.yaml`）。
- 运行期间修改当前生效的配置文件，会被自动监听并在几秒内热重载，无需手动重启。

## 环境变量覆盖

以下环境变量可以在不修改配置文件的情况下覆盖对应设置。环境变量在配置加载后生效，优先于配置文件中的值。

| 环境变量 | 覆盖的配置项 | 示例 |
|----------|-------------|------|
| `SERVER_PORT` | `server.port` | `9000` |
| `PASSWORD` | `auth.password` | `your_password` |
| `SUBSCRIPTION_URL` | `subscription.url` | `https://your-subscription-url` |
| `DEFAULT_TEMPLATE` | `default_template` | `default` |
| `CACHE_DIR` | `cache.directory` | `./data/cache` |
| `REFRESH_INTERVAL` | `subscription.refresh_interval`（分钟） | `2` |

示例：

```bash
export SERVER_PORT=9000
export PASSWORD="your_password"
export SUBSCRIPTION_URL="sub_url"
export DEFAULT_TEMPLATE="default"
export CACHE_DIR="./data/cache"
export REFRESH_INTERVAL=2
```

也可以只在单次运行时传入：

```bash
SERVER_PORT=8080 PASSWORD=secret123 ./singbox-subscribe-convert run -c config/my-config.yaml
```

## 运行产物与目录

以下路径都来自配置文件，且相对路径以运行工作目录为基准，可在配置中自行调整：

- **缓存目录**：由 `cache.directory` 指定，默认示例为 `./storage/cache`。节点与模板的本地缓存文件存放在此（例如 `node.json` 以及模板缓存文件）。程序启动时会自动创建该目录。
- **日志文件**：由 `logging.file` 指定，默认示例为 `./storage/log/server.log`。日志目录不存在时程序会自动创建。日志级别由 `logging.level` 控制（`debug` / `info` / `warn` / `error`）。

其他说明：

- 程序会对缓存目录中的文件变化进行监控，文件更新后自动重新加载，无需手动重启。
- Docker 部署时，上述缓存与日志均落在挂载的 `/singbox-subscribe-convert/storage/` 卷内，宿主机对应目录即上文 installation.md 中的 `storage` 目录，可持久化保存。
- 服务健康状态可通过 `GET /health` 查看，手动刷新节点与模板可调用 `GET /refresh?password=<密码>`。接口的详细说明见 [api.md](api.md)。
