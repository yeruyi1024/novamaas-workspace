# Upstream provenance / 上游来源

- Project / 项目：New API
- Repository / 仓库：https://github.com/QuantumNous/new-api
- Baseline release / 初始版本：`v1.0.0-rc.26`
- Release / 发布页：https://github.com/QuantumNous/new-api/releases/tag/v1.0.0-rc.26
- Commit / 原始提交：`8f6961c675932f406260ff0c218bc2aa0603e9b2`
- Fork repository / 本仓库：https://github.com/yeruyi1024/novamaas-workspace
- Fork initialization / 初始化日期：2026-09-05

本仓库从上述提交创建，保留其完整 Git 历史。初始化时未合入后续上游版本，未修改应用业务源码；改动为本分支 README、版本标识、构建文档及 GitHub Actions 配置。`VERSION` 中的 `-novamaas.*` 后缀用于区分本项目构建与上游发行版。

本项目由 NovaMaaS Workspace 独立维护，基于 QuantumNous/new-api 开源项目进行 fork 和本土化改造。结合我们的使用场景，上游变更幅度较大，仍有较多未关闭的 issue，现有版本难以直接满足本土化适配需求，因此我们选择固定版本作为维护起点。感谢上游作者及所有贡献者的工作。

后续会不定期评估并选择性同步上游中有益、兼容、适合本项目的改动。每次同步应记录上游提交、改动原因和验证结果；不会自动全量跟随上游主分支。

## 已选择同步的上游改动

- 2026-10-16：拉取官方 `main` 提交 [`d032204733`](https://github.com/QuantumNous/new-api/commit/d032204733db31208f10bac118b6e683aab3bf29) 进行对比，选择性同步上游提交 [`0ed497f066`](https://github.com/QuantumNous/new-api/commit/0ed497f066a68613375124303ef54f220267b334) 中 OpenAI → Anthropic 的输入用量转换规则：OpenAI 总输入扣除缓存读写后再返回 `input_tokens`，已有 Anthropic 语义则保持原值，避免重复扣减。本次仅修正 Anthropic 响应，不引入该上游提交的其他转换重构；保留原始计费用量、日志字段及其他协议响应。已通过 `relaykit` 全量测试与独立构建、OpenAI/Claude 适配器测试，以及缓存分离不改变计费和日志用量的回归测试。
- 2026-09-20：参考上游提交 [`7c044d7c5`](https://github.com/QuantumNous/new-api/commit/7c044d7c5c2d2beadf16b21910950f8f593bc3ef) 对真实 Kimi 思考模型名的保护，在现有 OpenRouter 兼容逻辑中同时检查公开模型名与模型映射后的上游名称。这样即使平台公开别名不是 `kimi-k2-thinking`，映射到该真实 ID 后也不会被错误裁剪为 `kimi-k2` 或被注入伪造的 reasoning 配置。该上游提交同时包含大范围模型修饰符与计费身份重构，本次未引入这些无关架构变化，只同步 Kimi 所需的最小兼容行为并增加映射回归测试。
- 2026-09-20：同步上游提交 [`6e10f9bc9`](https://github.com/QuantumNous/new-api/commit/6e10f9bc927a4eae889864a6ef601359d53526b9) 的 Kimi K3 动态工具加载修复。平台现在会保留对话中 `messages[].tools` 声明，只对 `content` 缺失的动态工具消息省略 `content` 键，将消息级工具纳入 Token 估算，并避免系统提示词注入改写这类消息；普通消息仍保留既有 `content: null` 兼容行为。合并时保留 NovaMaaS `/moonshot` 兼容入口的安全边界：使用普通 OpenAI 上游时继续明确拒绝无法真实复现的动态工具加载，原生 Moonshot 渠道则完整透传。已通过根模块全量 Go 测试、`relaykit` 全量测试与独立构建，以及前端类型检查和测试。
- 2026-09-09：同步上游提交 [`8c8c4153d`](https://github.com/QuantumNous/new-api/commit/8c8c4153d4b80d54352d21593de41aa9a6178f7e) 的日志用量统计修复，通过本仓库 [#21](https://github.com/yeruyi1024/novamaas-workspace/pull/21) 将 RPM/TPM 查询结果先扫描到独立结构，再赋值给完整统计对象，避免第二次扫描覆盖已汇总的 quota。该 PR 同时包含 NovaMaaS 下游专属的阿里百炼任务详情刷新、公开视频直连交付、24 小时登录会话、登录公告与违规确认审计，以及普通用户日志展示修复；已通过相关 Go 测试、前端测试、类型检查和 `relaykit` 独立构建。
- 2026-09-06：参考上游问题 [QuantumNous/new-api#6166](https://github.com/QuantumNous/new-api/issues/6166) 与待合并修复 [QuantumNous/new-api#6174](https://github.com/QuantumNous/new-api/pull/6174)，通过本仓库 [#11](https://github.com/yeruyi1024/novamaas-workspace/pull/11) 实现阿里百炼 Wan3 视频任务结果对整数、小数及数字字符串时长的兼容解析，解决任务可提交但轮询无法更新的问题。本仓库额外拒绝非数值、非有限值和超出本机 `int` 范围的输入，并补充 `relaykit` 数值解析测试、阿里任务适配器回归测试、独立模块构建及真实任务轮询验证。
- 2026-09-05：同步 [QuantumNous/new-api#6653](https://github.com/QuantumNous/new-api/pull/6653) 的提交 `362c9d666ab4dddbe789cba2c1acc4217573d6a5`，新增独立的 Volc Native（渠道类型 61）和火山方舟原生 `/api/v3` 图片、视频任务接口。本仓库补充了双向渠道隔离、任务凭据延续、令牌访问限制、请求边界校验、完整任务响应、前端国际化、测试和使用文档。上游 PR 在同步时仍处于未合并状态，因此本改动会在独立分支和本仓库 PR 中验证后再决定是否进入主分支。

This independently maintained fork starts at exactly the upstream commit above and retains its Git history. Initial changes cover fork documentation, build versioning and GitHub Actions; application source code is unchanged. Future upstream improvements will be reviewed and integrated selectively, with their source commits and validation recorded.

## 公司主分支同步

- 2026-10-16：PC 生图分支 rebase 到公司 `main` 提交 [`05915b6c6`](https://github.com/yeruyi1024/novamaas-workspace/commit/05915b6c6a2deda1553af22ead390cd4822211e9)（[#74](https://github.com/yeruyi1024/novamaas-workspace/pull/74) 节点资源与分类流量监控）。保留主分支监控接入及 Root 权限，生图请求快照继续沿用 `AdminAuth`，图片巡查仍需 Root 或其授权；原有生图逻辑和七种语言翻译均已核对保留。Go 全量回归、前端 123 个测试文件共 678 项测试、类型检查、受影响文件 lint/格式检查、根模块 vet/构建、`relaykit` 独立 vet/构建及前端生产构建通过。全库格式检查另有 16 个与公司 main 完全一致的既有文件问题，本次未改动。

- 2026-10-15：重新获取并核对公司 `main`，最新提交仍为 [`ce9e31f1e`](https://github.com/yeruyi1024/novamaas-workspace/commit/ce9e31f1ec17b39562fbd90c4990303a3cc2b86f)，PC 生图分支已完整包含该提交，无需重复合并。架构复核确认继续复用主分支 Relay、计费、请求快照、素材库及对象清理；合并生图与素材库重复的 OSS 签名实现，保留图片归属和存储用途校验，JPEG/WebP 缩略图沿用素材库压缩规则。存储、生图、controller、router 和 common 的无缓存回归、根模块全量 Go 测试及 vet/构建、`relaykit` 独立构建通过。MySQL 环境下浏览器验证模型读取、历史恢复、OSS 原图预览和直连下载通过，下载 PNG 为 1254 × 1254、2,173,905 字节。

- 2026-10-14：PC 生图分支合入公司主分支提交 [`ce9e31f1e`](https://github.com/yeruyi1024/novamaas-workspace/commit/ce9e31f1ec17b39562fbd90c4990303a3cc2b86f)，接入 [#72](https://github.com/yeruyi1024/novamaas-workspace/pull/72) 的请求快照异步归档、对象存储读取及历史兼容能力。合并保留生图存储用途和图片巡查权限；请求快照的查看与下载统一沿用主分支 `AdminAuth`，移除生图专属的快照授权限制和对应前端拦截。已通过 Go 全量回归及权限调整后的受影响包回归、根模块 vet/构建、`relaykit` 独立 vet/构建、前端全量测试及 93 项受影响回归、类型检查、受影响文件 lint/格式检查和生产构建。

## 同步上游

推荐使用 GitHub Actions 创建 Draft/WIP PR，再由维护者启动 AI 分析、确认计划并在同一 PR 中实施，操作步骤见 [上游版本选择性同步](docs/UPSTREAM_SYNC.zh_CN.md)。该机制固定目标版本和来源提交，保存功能处理计划、适配记录与最新提交的 CI 证据；不会自动合并主分支。评估目标与实际纳入内容分别记录，选择性同步不会覆盖上方初始 fork 基线。

以下命令保留为人工检查和迁入的基础操作：

```bash
git remote add upstream https://github.com/QuantumNous/new-api.git
git fetch upstream --tags
git switch -c sync/upstream-change main
# 审查改动后，将占位符替换为选定的真实提交：
git cherry-pick <upstream-commit>
```

已配置 `upstream` 时无需重复添加。推送同步分支并提交 PR，等待本仓库构建和测试通过后再合入。

## 许可与署名

保留上游 [LICENSE](LICENSE)、[NOTICE](NOTICE)、[THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md) 及源码中的版权和署名信息。本项目沿用上游 AGPL-3.0 许可及 NOTICE 中的附加声明。二进制包和镜像包含这些文件及本来源说明。

原有工作流原样保存在 [.github/upstream-workflows](.github/upstream-workflows)，仅用于参考，不由 GitHub Actions 执行。本仓库实际工作流位于 [.github/workflows](.github/workflows)。
