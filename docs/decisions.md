# 决策记录（decisions）

技术决策的「背景 / 决策 / 后果」单文件记录。新决策追加到末尾；如某条被推翻，更新该条状态而不删除（保留演进脉络）。按文档约定（见根 CLAUDE.md），难以逆转或易被后续质疑的设计取舍都应在此留痕。

---

## 2026-09 · 许可：维持 Apache-2.0，重写两处 GPL 源文件

**背景**：仓库 LICENSE 为 Apache-2.0，但 `pkg/logger` 与 `pkg/safe_close` 文件头声明代码源自 mosdns（GPL-3.0）。以 Apache-2.0 分发含 GPL-3.0 文件的整体仓库违反许可条款（GPL 文件不能改授 Apache）。

**决策**：不对全仓库转 GPL-3.0（会连带约束自研代码），而是**干净重写这两个包**——两者体量小（zap 封装 / 优雅关闭协调器），保持导出 API 不变、行为等价，去掉 GPL 版权头，为两包补最小单测。仓库整体维持宽松的 Apache-2.0，后续贡献亦不接纳 GPL 传染代码。

**后果**：两个文件实现与上游已无直接派生态，可安全随 Apache-2.0 分发；代价是本次开源就绪多了一个小的代码改动面。

---

## 2026-09 · module 路径对齐公开仓库

**背景**：`go.mod` 模块名为 `github.com/haierkeys/singbox-subscribe-convert`，与公开仓库 `github.com/praise579/fit_sfm` 不一致，导致 import 路径、构建注入（ldflags）与仓库名错位。

**决策**：module 路径统一为 `github.com/praise579/fit_sfm`，同步修改全部源码 import 与 Makefile/CI 中的构建注入变量。

**后果**：import 与仓库名一致，`go install`/引用方不再困惑；需注意与上游 fork 的关系说明（README 链接均已改指新仓库）。

---

## 2026-09 · 文档结构：README 登录页 + docs/ 纯 Markdown

**背景**：根目录 `README.md`（约 780 行）与 `BUILD-FROM-SOURCE.md`（约 990 行保姆级教程）内容大量重叠，且缺少开发者/决策类文档。目标是面向开源交付。

**决策**：
1. 采用**纯 Markdown 目录**（根 README + `docs/guide` + `docs/development`），不引入 MkDocs/Docusaurus 等站点生成器，降低构建与维护成本；
2. `README.md` 瘦身为登录页（简介/特性/快速开始/文档导航），详细内容去重后迁入 `docs/guide/`；
3. 配置字段的**唯一事实源（SSOT）为 `config/config.yaml` 注释**，`docs/guide/configuration.md` 只按场景组织并指向该文件，避免双源漂移；
4. 语言用简体中文，代码/命令/配置项/模板变量保留原文。

**后果**：单一入口可维护、零构建依赖；配置变更只需改 yaml 注释一处。镜像名与徽章链接因发布策略未定暂不写死（源码构建为主）。

---

## 2026-09 · 镜像/徽章发布策略暂缓

**背景**：CI（`go-release-docker.yml`）已具备按 tag 构建 ghcr/Docker Hub 镜像与附件的流程，但发布者账号、是否推 Docker Hub、徽章地址未最终确定。

**决策**：本次文档以「从源码构建」为主要安装方式，不写死任何具体镜像名与徽章链接；待首次正式 Release 确定发布渠道后再补齐 README 顶部徽章与镜像示例。

**后果**：README 短期内少几个链接/徽章，避免在文档中承诺尚未就绪的发布渠道；补记时属于纯文档追加，无重构成本。
