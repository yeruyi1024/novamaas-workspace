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

- 2026-09-06：参考上游问题 [QuantumNous/new-api#6166](https://github.com/QuantumNous/new-api/issues/6166) 与待合并修复 [QuantumNous/new-api#6174](https://github.com/QuantumNous/new-api/pull/6174)，通过本仓库 [#11](https://github.com/yeruyi1024/novamaas-workspace/pull/11) 实现阿里百炼 Wan3 视频任务结果对整数、小数及数字字符串时长的兼容解析，解决任务可提交但轮询无法更新的问题。本仓库额外拒绝非数值、非有限值和超出本机 `int` 范围的输入，并补充 `relaykit` 数值解析测试、阿里任务适配器回归测试、独立模块构建及真实任务轮询验证。
- 2026-09-05：同步 [QuantumNous/new-api#6653](https://github.com/QuantumNous/new-api/pull/6653) 的提交 `362c9d666ab4dddbe789cba2c1acc4217573d6a5`，新增独立的 Volc Native（渠道类型 61）和火山方舟原生 `/api/v3` 图片、视频任务接口。本仓库补充了双向渠道隔离、任务凭据延续、令牌访问限制、请求边界校验、完整任务响应、前端国际化、测试和使用文档。上游 PR 在同步时仍处于未合并状态，因此本改动会在独立分支和本仓库 PR 中验证后再决定是否进入主分支。

This independently maintained fork starts at exactly the upstream commit above and retains its Git history. Initial changes cover fork documentation, build versioning and GitHub Actions; application source code is unchanged. Future upstream improvements will be reviewed and integrated selectively, with their source commits and validation recorded.

## 同步上游

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
