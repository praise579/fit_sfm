# sing-box 订阅转换服务 —— 构建与运维入口
#
# 四组目标：
#   构建        交叉编译各平台产物到 build/<os>_<arch>/
#   开发        本机运行、测试、清理
#   本地 Docker 在开发机上编译并起容器，不推送任何远程仓库
#   发布        构建镜像并推送到 Docker Hub
#
# 产物布局 build/<os>_<arch>/<P_BIN> 是 Dockerfile 的硬依赖（见其 COPY 行），勿改。

# ------------------------------------------------------------------ 变量

# 项目标识
P_NAME = singbox-subscribe-convert
P_BIN  = sb-sub-c

# 模块路径，供 ldflags 注入；从 go.mod 读取，避免与 module 名两处漂移
REPO = $(eval REPO := $$(shell go list -f '{{.ImportPath}}' .))$(value REPO)

# 版本注入源。仓库无 tag 时 GitTag 退化为短 hash（--always），不会注入空串
GitTag         = $(shell git describe --tags --always)
GitVersion     = $(shell git log -1 --format=%h)
GitVersionDesc = $(shell git log -1 --format=%s)
BuildTime      = $(shell date +%FT%T%z)

LDFLAGS = -ldflags '-X ${REPO}/global.Version=$(GitTag) -X "${REPO}/global.GitTag=$(GitVersion) / $(GitVersionDesc)" -X ${REPO}/global.BuildTime=$(BuildTime)'

# 工具链（关闭 CGO，产物可直接跑在 alpine 上）
CGO     = CGO_ENABLED=0
goBuild = go build ${LDFLAGS}
goRun   = go run ${LDFLAGS}

# 目录
projectRootDir = $(shell pwd)
sourceDir      = $(projectRootDir)
buildDir       = $(projectRootDir)/build

# 发布：目标镜像仓库与 tag 前缀
DockerHubUser = haierkeys
DockerHubName = singbox-subscribe-convert
ReleaseTagPre = release-v
DevelopTagPre = develop-v

# 本地 Docker：镜像 / 容器 / 端口独立取值，与线上镜像和容器名互不干扰。
# 端口须与本地 config.yaml 的 server.port 一致，否则映射出来的端口无人监听。
LocalImage     = $(DockerHubName):local
LocalContainer = sb-sub-c-local
LocalPort      = 1900

.PHONY: all build-all run test clean \
        docker-build docker-up docker-down docker-rebuild docker-logs docker-clean \
        push-release push-dev \
        build-macos-amd64 build-macos-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64 \
        gox-linux gox-all

all: test build-all

# ------------------------------------------------------------------ 构建

# 交叉编译各平台，产物均在 build/<os>_<arch>/
build-macos-amd64:
	$(CGO) GOOS=darwin GOARCH=amd64 $(goBuild) -o $(buildDir)/darwin_amd64/${P_BIN} -v $(sourceDir)
build-macos-arm64:
	$(CGO) GOOS=darwin GOARCH=arm64 $(goBuild) -o $(buildDir)/darwin_arm64/${P_BIN} -v $(sourceDir)
build-linux-amd64:
	$(CGO) GOOS=linux GOARCH=amd64 $(goBuild) -o $(buildDir)/linux_amd64/${P_BIN} -v $(sourceDir)
build-linux-arm64:
	$(CGO) GOOS=linux GOARCH=arm64 $(goBuild) -o $(buildDir)/linux_arm64/${P_BIN} -v $(sourceDir)
build-windows-amd64:
	$(CGO) GOOS=windows GOARCH=amd64 $(goBuild) -o $(buildDir)/windows_amd64/${P_BIN}.exe -v $(sourceDir)

# 本地出全部平台产物（串行，较慢）
build-all:
	$(MAKE) build-macos-amd64
	$(MAKE) build-macos-arm64
	$(MAKE) build-linux-amd64
	$(MAKE) build-linux-arm64
	$(MAKE) build-windows-amd64

# CI 专用：GitHub Actions 用 gox 并行出多平台（.github/workflows/go-release-docker.yml）。
# 本机未安装 gox，本地构建请用 build-all。
gox-linux:
	$(CGO) gox ${LDFLAGS} -osarch="linux/amd64 linux/arm64" -output="$(buildDir)/{{.OS}}_{{.Arch}}/${P_BIN}"
gox-all:
	$(CGO) gox ${LDFLAGS} -osarch="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64" -output="$(buildDir)/{{.OS}}_{{.Arch}}/${P_BIN}"

# ------------------------------------------------------------------ 开发

# 本机直接运行。须在仓库根执行（配置按 config.yaml → config/config.yaml 顺序查找）
run:
	$(goRun) -v $(sourceDir)

test:
	go test ./...

clean:
	rm -rf $(buildDir)

# ------------------------------------------------------------------ 本地 Docker
# 只在本机操作，不推送任何远程仓库。
# 本地镜像为 linux/arm64（Apple Silicon 原生，免 QEMU 模拟）；
# amd64 的回归验证由下方 push-* 承担。背景见 docs/decisions.md。

# 编译 linux/arm64 并构建本地镜像
docker-build: build-linux-arm64
	docker build --platform linux/arm64 -t $(LocalImage) -f Dockerfile .

# 后台起容器；挂仓库根 config.yaml（只读）与 storage/
docker-up:
	docker run -d --name $(LocalContainer) \
		-p $(LocalPort):$(LocalPort) \
		-v $(projectRootDir)/config.yaml:/$(P_NAME)/config.yaml:ro \
		-v $(projectRootDir)/storage:/$(P_NAME)/storage \
		$(LocalImage)

# 停并删除本地容器（未运行时静默跳过）
docker-down:
	-docker stop $(LocalContainer)
	-docker rm $(LocalContainer)

# 改代码后一键：停容器 → 重编译并重打镜像 → 重新起
docker-rebuild:
	$(MAKE) docker-down
	$(MAKE) docker-build
	$(MAKE) docker-up

# 跟踪容器日志
docker-logs:
	docker logs -f --tail=100 $(LocalContainer)

# 清理本项目镜像与 none 镜像
docker-clean:
	$(call dockerImageClean)

# ------------------------------------------------------------------ 发布（推 Docker Hub）
# 均为 linux/amd64，与线上一致。

push-release: build-linux-amd64
	$(call dockerImageClean)
	docker build --platform linux/amd64  -t  $(DockerHubUser)/$(DockerHubName):latest -f Dockerfile .
	docker tag  $(DockerHubUser)/$(DockerHubName):latest $(DockerHubUser)/$(DockerHubName):$(ReleaseTagPre)$(GitTag)
	docker push $(DockerHubUser)/$(DockerHubName):$(ReleaseTagPre)$(GitTag)
	docker push $(DockerHubUser)/$(DockerHubName):latest

push-dev: build-linux-amd64
	$(call dockerImageClean)
	docker build --platform linux/amd64 -t $(DockerHubUser)/$(DockerHubName):dev-latest -f Dockerfile .
	docker tag $(DockerHubUser)/$(DockerHubName):dev-latest $(DockerHubUser)/$(DockerHubName):$(DevelopTagPre)$(GitTag)
	docker push $(DockerHubUser)/$(DockerHubName):$(DevelopTagPre)$(GitTag)
	docker push $(DockerHubUser)/$(DockerHubName):dev-latest

# ------------------------------------------------------------------ 辅助

define dockerImageClean
	@echo "docker Image Clean"
	bash scripts/docker_image_clean.sh
endef
