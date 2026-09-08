<div align="center">

<img src="./web/public/logo.png" alt="NovaMaaS logo" width="88" />

# 星枢 MaaS 平台

**NovaMaaS — 面向 AI Token 供应聚合、商业分销与算力资源运营的统一基础设施**

[![Build, package and publish](https://github.com/yeruyi1024/novamaas-workspace/actions/workflows/build.yml/badge.svg)](https://github.com/yeruyi1024/novamaas-workspace/actions/workflows/build.yml)
[![License: AGPL v3](https://img.shields.io/badge/license-AGPL--3.0-3157d5.svg)](LICENSE)
[![Upstream](https://img.shields.io/badge/upstream-QuantumNous%2Fnew--api-7357d9.svg)](https://github.com/QuantumNous/new-api)

[平台定位](#平台定位) · [能力版图](#能力版图) · [产品路线图](#产品路线图) · [与上游差异](#novamaas-与上游差异) · [版本与上游维护](#版本与上游维护) · [快速开始](#快速开始) · [项目文档](#项目文档)

</div>

> [!IMPORTANT]
> **项目来源与致谢**
>
> 星枢 MaaS 平台（NovaMaaS）基于 **[QuantumNous/new-api](https://github.com/QuantumNous/new-api)** 建设，初始基线严格对应上游 **[v1.0.0-rc.26](https://github.com/QuantumNous/new-api/releases/tag/v1.0.0-rc.26)** 和提交 **`8f6961c675932f406260ff0c218bc2aa0603e9b2`**，并完整保留上游 Git 历史。
>
> 感谢 **QuantumNous、new-api 项目作者及所有贡献者**提供的开源基础与持续投入。NovaMaaS 将遵循上游 AGPL-3.0 许可及附加声明，持续保留原作者署名、项目来源和许可证信息。

**NovaMaaS Workspace is an independently maintained fork of [QuantumNous/new-api](https://github.com/QuantumNous/new-api), initially based on exactly `v1.0.0-rc.26`. The repository retains its upstream Git history, license notices, attribution, and provenance records.**

## 平台定位

星枢 MaaS 平台面向需要长期运营 AI 服务的组织，为异构模型供应、Token 资源、客户访问与商业结算提供统一的管理基础。平台以 new-api 成熟的多模型网关能力为技术起点，在兼容现有生态的基础上，逐步构建从 **Token 聚合与分销** 到 **算力纳管与租赁** 的完整供应网络。

NovaMaaS 的核心目标不是增加孤立功能，而是将供应、产品、租户和结算连接为可运营的业务闭环：

- 对供应侧，统一接入和管理不同模型供应商、Token 库存及未来的算力资源。
- 对运营侧，提供路由、权限、定价、额度、用量与结算能力。
- 对分销侧，支持直客、渠道合作伙伴、企业团队等多种商业交付模式。
- 对使用侧，以统一接口屏蔽上游差异，降低 AI 应用与团队的接入成本。

## 能力版图

| 能力层 | 建设内容 | 当前阶段 |
| --- | --- | --- |
| 统一 AI 网关 | 多供应商接入、协议转换、模型路由、失败重试与访问控制 | 已具备基础能力 |
| Token 聚合 | 统一管理多来源 Token、渠道、模型能力、额度和使用策略 | 持续增强 |
| Token 分销 | 面向客户、团队和合作渠道封装访问能力，并支持计量与费用管理 | 持续增强 |
| 多租户商业化 | 租户隔离、组织权限、产品定价、渠道策略、账单与经营分析 | 产品路线图 |
| 算力纳管 | 统一登记、分组、监控和调度异构算力资源 | 产品路线图 |
| 算力租赁 | 将可调度算力封装为可分配、可计量、可结算的供应产品 | 产品路线图 |

> [!NOTE]
> “持续增强”与“产品路线图”用于区分已经交付的基础能力和后续建设方向，不代表尚未发布的功能已经可用于生产环境。正式能力范围以对应版本说明和实际界面为准。

### 对象存储与火山 Base64 兼容增强

系统设置新增通用“存储配置（Profile）+ 用途策略（Policy）”模型，首个驱动为阿里云 OSS。火山原生渠道可按渠道及模型开启 Base64 媒体暂存：平台校验并上传 `content[].image_url.url` 中的 JPEG、PNG 或 WebP Data URI，并在发往上游的请求副本中替换为限时签名 HTTPS 地址。任务日志保存并展示实际发送给上游的转换后请求；使用日志为管理员保留原始 Base64 请求并标记“已被临时存储转换”，非管理员无法查看请求体。

- OSS Bucket 应保持私有，不需要设置公共读；平台使用 OSS V4 签名地址提供临时访问，最长有效期为 168 小时。
- 静态 AccessKey 在数据库中使用 AES-GCM 加密。生产部署应显式配置稳定的 `STORAGE_CREDENTIAL_ENCRYPTION_KEY`；未配置时会依次尝试复用 `CRYPTO_SECRET`、`SESSION_SECRET`，三者均缺失则拒绝保存静态凭证。
- 对象路径包含用途前缀、UTC 日期、不可逆用户标识和任务 ID，实现用户与任务隔离；Data URI 声明的 MIME 类型必须与解码后的文件特征一致，并受单文件、单请求总量和文件数限制。
- 为保证 SQLite、MySQL 和 PostgreSQL 默认部署下都能完整记录原始请求，当前视频任务请求体仍执行 2 MiB 审计写入上限；超限请求会在上传前明确拒绝，不会以丢弃或改写 Base64 日志换取继续执行。
- 任务成功、失败或确认取消后会触发清理；上传失败、进程中断和上游状态不确定时由数据库租约重试与最长保留期限兜底，成功删除后的对象账本墓碑保留 30 天再分批清理。建议同时在 OSS 配置生命周期规则，按 `temporary/relay-media/` 前缀做更长周期的灾难兜底清理。
- 本次只新增 `storage_profiles`、`storage_credentials`、`storage_policies`、`storage_objects` 四张表，不修改既有数据库表字段。Profile 的服务商类型已为腾讯云 COS 和 S3 兼容存储（包括 MinIO）预留，当前尚未启用对应驱动；后续素材库可复用同一存储层并按用途生成临时授权地址。
- 火山原生和 DoubaoVideo 任务日志支持直接查询任务信息，分别复用 `/api/v3/contents/generations/tasks/{taskID}` 和 `/v1/video/generations/{taskID}`；普通用户只能查询自己的任务，管理员可从任务日志跨用户诊断。

配置顺序：先在“系统设置 → 存储 → 对象存储”创建并测试 OSS 配置，再启用“中转媒体临时存储”策略，最后在目标火山原生渠道的高级设置中开启“Base64 媒体暂存”。阿里云侧最小权限需覆盖目标业务前缀以及 `temporary/relay-media/healthcheck/` 测试前缀的上传、签名读取和删除；接口行为参考[阿里云 OSS Go SDK V2 文档](https://help.aliyun.com/zh/oss/developer-reference/manual-for-go-sdk-v2/)、[V4 预签名下载文档](https://help.aliyun.com/en/oss/developer-reference/v2-presign-download)和[生命周期规则文档](https://help.aliyun.com/zh/oss/user-guide/lifecycle-rules-based-on-the-last-modified-time/)，火山请求格式参考[火山方舟原生内容生成接口](https://docs.volcengine.com/docs/82379/1520757?lang=zh)。

## 产品路线图

NovaMaaS 将围绕供应聚合、商业运营和算力资源三个方向持续演进。路线图不绑定未经验证的交付日期，每项能力会在完成实现、测试与版本记录后正式发布。

| 阶段 | 建设目标 | 重点能力 |
| --- | --- | --- |
| 第一阶段：平台基础 | 建立可独立维护、可追溯、可发布的 NovaMaaS 基线 | 上游来源治理、本土化适配、统一构建、版本标识、前端产品化呈现 |
| 第二阶段：Token 供应网络 | 将分散的模型和 Token 供应转化为统一资源池 | 供应接入、库存管理、智能路由、额度策略、用量观测与分销能力 |
| 第三阶段：多租户商业化 | 支撑企业客户、内部团队和渠道合作伙伴的规模化运营 | 租户隔离、组织权限、产品目录、差异化定价、渠道分润与账单结算 |
| 第四阶段：算力资源平台 | 将 GPU 等算力资源纳入统一供应与交易体系 | 算力纳管、资源监控、任务调度、容量编排、算力租赁与统一计费 |

## NovaMaaS 与上游差异

此处不是完整 Changelog，而是 NovaMaaS 相对 `QuantumNous/new-api` 的**关键、长期运行时差异**清单，用于上游同步时判断哪些能力必须保留、重做或移除。PR 只有同时满足以下条件才收录：

1. 改变核心产品运行时能力，例如供应商接入、请求/响应协议、模型路由、计费、权限、数据兼容或租户能力。
2. 上游主线尚无等价实现，或 NovaMaaS 在上游实现之外保留了有意义的行为与安全边界。
3. 差异具有长期维护价值；未来同步上游时忽略它，可能导致功能回退、兼容性故障或业务语义变化。

CI/CD、镜像发布、构建环境、首页展示、文档整理、测试补充、内部重构、依赖升级、临时排障，以及未合并或已被替代的方案不进入本表，除非它们同时改变上述核心运行时边界。

<!-- novamaas-pr-ledger:start -->

| 关键差异 PR | 日期 | 类型 | 领域 | 关键变化 | 与上游关系 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| [#19](https://github.com/yeruyi1024/novamaas-workspace/pull/19) | 2026-09-08 | `feat` | 视频任务 / 计费 | 为 Doubao Seedance 2.0 增加稳定公开模型名，使其可映射到不同上游模型 ID，同时复用 720p、1080p、4K 与视频输入计费倍率。 | NovaMaaS 下游专属；上游当前没有该稳定公开别名及其参数计费映射。 | PR 审核中 |
| [#18](https://github.com/yeruyi1024/novamaas-workspace/pull/18) | 2026-09-08 | `fix` | 对象存储 / 数据兼容 | 使用 GORM 方言感知条件引用存储策略 `key` 列，修复 MySQL 1064 错误，并增加 SQLite、MySQL、PostgreSQL 查询回归测试。 | NovaMaaS 下游专属；属于 #17 对象存储能力的兼容性修复，上游当前没有等价的存储策略实现。 | PR 审核中 |
| [#17](https://github.com/yeruyi1024/novamaas-workspace/pull/17) | 2026-09-08 | `feat` | 对象存储 / 火山方舟 | 新增系统级存储 Profile 与用途 Policy，以私有阿里云 OSS 签名地址兼容火山原生 Base64 媒体请求，并实现任务终态清理、请求审计分流及管理员访问控制。 | NovaMaaS 下游专属；上游当前没有等价的火山 Base64 暂存、通用对象存储策略与审计分流组合实现。 | PR 审核中 |
| [#14](https://github.com/yeruyi1024/novamaas-workspace/pull/14) | 2026-09-07 | `feat` | 视频任务 / 使用日志 | 为 DoubaoVideo 增加 302 重定向与 600 秒服务端代理模式，记录 DoubaoVideo、原生 Ark、阿里百炼视频任务请求体，并在鉴权日志详情中按数据可用性提供视频下载和格式化 JSON 查看入口。 | NovaMaaS 下游专属；上游当前没有等价的视频交付模式与任务请求审计组合实现。 | PR 审核中 |
| [#11](https://github.com/yeruyi1024/novamaas-workspace/pull/11) | 2026-09-06 | `fix` | 阿里百炼 | 兼容 Wan3 任务结果中的整数、小数和数字字符串时长，恢复异步任务状态更新并增加适配器回归测试。 | 对齐上游 [#6166](https://github.com/QuantumNous/new-api/issues/6166) / [#6174](https://github.com/QuantumNous/new-api/pull/6174)，并增加非法值和溢出保护。 | PR 审核中 |
| [#8](https://github.com/yeruyi1024/novamaas-workspace/pull/8) | 2026-09-06 | `feat` | 火山方舟 | 为 Volc Native 增加仅改写顶层 `model` 的模型映射，平台侧继续使用公开别名完成权限、计费和日志。 | #1 的下游增强；上游 [#6653](https://github.com/QuantumNous/new-api/pull/6653) 尚未覆盖该映射能力。 | 已合并 |
| [#1](https://github.com/yeruyi1024/novamaas-workspace/pull/1) | 2026-09-05 | `feat/fix` | 火山方舟 | 选择性引入 Volc Native 渠道，并补齐任务凭据延续、取消状态、响应关闭、路由隔离、权限约束和多语言支持。 | 来源为仍未合并的上游 [#6653](https://github.com/QuantumNous/new-api/pull/6653) / [#4705](https://github.com/QuantumNous/new-api/issues/4705)，NovaMaaS 追加安全与兼容加固。 | 已合并 |

<!-- novamaas-pr-ledger:end -->

维护约束：每个面向 `main` 的 PR 都必须在 PR 描述中二选一标记“关键差异”或“常规变更”，并说明判断理由。只有符合上述三个条件的“关键差异”PR 才在标记区域新增一行，使用真实 PR 编号和链接，说明类型、影响领域、关键变化、与上游的关系及当前状态；“常规变更”不得为自身新增账本行。上游同步类 PR 还必须同步更新 [UPSTREAM.md](UPSTREAM.md)。`.github/workflows/build.yml` 会校验分类是否唯一，并对“关键差异”验证当前 PR 的账本行。

## 版本与上游维护

NovaMaaS 采用“固定基线、定期评估、选择性合并、完整记录”的长期维护方式：

1. 定期检查 QuantumNous/new-api 的新版本、重要修复和兼容性改进。
2. 对候选改动进行代码审查、依赖分析和数据库兼容性评估，不自动全量追随上游主分支。
3. 通过独立分支和 Pull Request 合并上游改动，并完成 Go、前端、数据库与容器构建验证。
4. 每次上游同步都在 [UPSTREAM.md](UPSTREAM.md) 中记录来源提交、合并原因、适配内容和验证结果。
5. 正式能力随 NovaMaaS 版本统一发布，在版本说明中建立“上游提交—NovaMaaS 版本—构建产物”的对应关系。

当前 PR 合并 CI 使用 `build_<UTC 合并时间>_<架构>` 标识单架构镜像，并使用 `build_<UTC 合并时间>_multiarch` 标识 GHCR 多架构镜像。产品版本由 [VERSION](VERSION) 管理；发布时应确保源码版本、发布说明、安装包和容器镜像之间可以相互追溯。

## 快速开始

### 使用容器镜像

下面展示 PR 合并构建的标签格式。请将 `YYYYMMDDTHHMMSSZ` 替换为 CI Summary 或 [Packages](https://github.com/yeruyi1024/novamaas-workspace/pkgs/container/novamaas-workspace) 中的实际 UTC 合并时间戳；生产部署应选择经过验证的固定标签或镜像 digest。

```bash
NOVAMAAS_IMAGE=ghcr.io/yeruyi1024/novamaas-workspace:build_YYYYMMDDTHHMMSSZ_multiarch
docker pull "$NOVAMAAS_IMAGE"
docker run -d --name novamaas --restart unless-stopped \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v novamaas-data:/data \
  "$NOVAMAAS_IMAGE"
```

部署完成后访问 `http://localhost:3000`，按照初始化页面完成管理员和数据库配置。SQLite 数据保存在挂载的数据卷中；升级或迁移前必须先备份数据库。

### 从源码构建

```bash
git clone https://github.com/yeruyi1024/novamaas-workspace.git
cd novamaas-workspace
docker build -t novamaas:local .
docker run --rm -p 3000:3000 -v novamaas-data:/data novamaas:local
```

完整的构建、安装包、双架构镜像和腾讯云 CCR 配置说明见 [NovaMaaS 构建文档](docs/BUILD.zh_CN.md)。

## 项目文档

| 文档 | 用途 |
| --- | --- |
| [UPSTREAM.md](UPSTREAM.md) | 上游基线、同步记录、来源提交与维护策略 |
| [docs/BUILD.zh_CN.md](docs/BUILD.zh_CN.md) | 本地构建、CI、安装包、GHCR 与腾讯云 CCR 发布说明 |
| [docs/VOLC_NATIVE.zh_CN.md](docs/VOLC_NATIVE.zh_CN.md) | 火山方舟原生 API 渠道、任务接口和兼容性边界 |
| [LICENSE](LICENSE) | AGPL-3.0 许可证与适用条款 |
| [NOTICE](NOTICE) | 上游声明、署名和附加许可说明 |
| [THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md) | 第三方依赖许可证信息 |

## 构建与发布

- 普通 PR 和手动工作流只执行源码检查，不发布镜像。
- PR 合并到 `main` 后构建 Linux amd64/arm64 安装包和容器镜像。
- GitHub-hosted runner 将 amd64/arm64 多架构镜像发布至 [GitHub Container Registry](https://github.com/yeruyi1024/novamaas-workspace/pkgs/container/novamaas-workspace)；self-hosted runner 将 Linux amd64 镜像发布至 `ccr.ccs.tencentyun.com/nova-proj/nova-maas`。
- PR 合并构建使用 `build_<UTC 合并时间>_<架构>` 标签；发布 Release 或推送 Git Tag 不触发此工作流，正式 Release 应使用独立的版本标签方案。
- GHCR 使用 GitHub 自动提供的 `GITHUB_TOKEN`；腾讯云凭据仅保存在 GitHub Actions Secrets 中，不得写入源码、文档或普通 Variables。

## 许可与合规

本项目沿用上游 [AGPL-3.0 许可证](LICENSE)，保留 [NOTICE](NOTICE)、[第三方许可](THIRD-PARTY-LICENSES.md) 和原作者署名。任何部署和商业化使用都必须遵守许可证、上游服务条款及所在地区适用的法律法规。

使用第三方模型、Token、支付、算力或其他上游资源时，运营方必须自行取得合法授权，并承担备案、内容安全、实名、日志留存、税务和数据合规等责任。

## 上游项目原始说明 / Original upstream documentation

以下保留上游说明及署名，其中的仓库、发布页和镜像地址指向上游。使用本分支请以上面的地址和构建说明为准。

<details>
<summary><strong>展开查看 QuantumNous/new-api 原始 README</strong></summary>

<br />

<div align="center">

![new-api](/web/public/logo.png)

# New API

🍥 **Next-Generation LLM Gateway and AI Asset Management System**

<p align="center">
  <a href="./README.zh_CN.md">简体中文</a> |
  <a href="./README.zh_TW.md">繁體中文</a> |
  <strong>English</strong> |
  <a href="./README.fr.md">Français</a> |
  <a href="./README.ja.md">日本語</a>
</p>

<p align="center">
  <a href="https://raw.githubusercontent.com/Calcium-Ion/new-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/Calcium-Ion/new-api?color=brightgreen" alt="license">
  </a><!--
  --><a href="https://github.com/Calcium-Ion/new-api/releases/latest">
    <img src="https://img.shields.io/github/v/release/Calcium-Ion/new-api?color=brightgreen&include_prereleases" alt="release">
  </a><!--
  --><a href="https://hub.docker.com/r/CalciumIon/new-api">
    <img src="https://img.shields.io/badge/docker-dockerHub-blue" alt="docker">
  </a>
  <a href="https://atomgit.com/QuantumNous/new-api" target="_blank">
    <img alt="AtomGit G-Star" src="https://atomgit.com/QuantumNous/new-api/star/badge.svg"/>
  </a>
</p>

<p align="center">
  <a href="https://trendshift.io/repositories/20180" target="_blank">
    <img src="https://trendshift.io/api/badge/repositories/20180" alt="QuantumNous%2Fnew-api | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/>
  </a>
  <br>
  <a href="https://hellogithub.com/repository/QuantumNous/new-api" target="_blank">
    <img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=539ac4217e69431684ad4a0bab768811&claim_uid=tbFPfKIDHpc4TzR" alt="Featured｜HelloGitHub" style="width: 250px; height: 54px;" width="250" height="54" />
  </a><!--
  -->
  <a href="https://atomgit.com/QuantumNous/new-api" target="_blank">
    <img alt="AtomGit G-Star" src="https://atomgit.com/QuantumNous/new-api/star/new_badge.svg" width="250" height="55" />
  </a>
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-key-features">Key Features</a> •
  <a href="#-deployment">Deployment</a> •
  <a href="#-documentation">Documentation</a> •
  <a href="#-help-support">Help</a>
</p>

</div>

## 📝 Project Description

> [!IMPORTANT]
> - This project is intended solely for lawful and authorized AI API gateway, organization-level authentication, multi-model management, usage analytics, cost accounting, and private deployment scenarios.
> - Users must lawfully obtain upstream API keys, accounts, model services, and interface permissions, and must comply with upstream terms of service and applicable laws and regulations.
> - Users should ensure their use complies with upstream terms of service and applicable laws and regulations.
> - When providing generative AI services to the public, users should comply with applicable regulatory requirements and fulfill all filing, licensing, content safety, real-name verification, log retention, tax, and upstream authorization obligations required by their jurisdiction.

---

## 🤝 Trusted Partners

<p align="center">
  <em>No particular order</em>
</p>

<p align="center">
  <a href="https://www.cherry-ai.com/" target="_blank">
    <img src="./docs/images/cherry-studio.png" alt="Cherry Studio" height="80" />
  </a><!--
  --><a href="https://github.com/iOfficeAI/AionUi/" target="_blank">
    <img src="./docs/images/aionui.png" alt="Aion UI" height="80" />
  </a><!--
  --><a href="https://bda.pku.edu.cn/" target="_blank">
    <img src="./docs/images/pku.png" alt="Peking University" height="80" />
  </a><!--
  --><a href="https://www.compshare.cn/?ytag=GPU_yy_gh_newapi" target="_blank">
    <img src="./docs/images/ucloud.png" alt="UCloud" height="80" />
  </a><!--
  --><a href="https://www.aliyun.com/" target="_blank">
    <img src="./docs/images/aliyun.png" alt="Alibaba Cloud" height="80" />
  </a><!--
  --><a href="https://io.net/" target="_blank">
    <img src="./docs/images/io-net.png" alt="IO.NET" height="80" />
  </a>
</p>

---

## 🙏 Special Thanks

<p align="center">
  <a href="https://www.jetbrains.com/?from=new-api" target="_blank">
    <img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" alt="JetBrains Logo" width="120" />
  </a>
</p>

<p align="center">
  <strong>Thanks to <a href="https://www.jetbrains.com/?from=new-api">JetBrains</a> for providing free open-source development license for this project</strong>
</p>

---

## 🚀 Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the project
git clone https://github.com/QuantumNous/new-api.git
cd new-api

# Edit docker-compose.yml configuration
nano docker-compose.yml

# Start the service
docker-compose up -d
```

<details>
<summary><strong>Using Docker Commands</strong></summary>

```bash
# Pull the latest image
docker pull calciumion/new-api:latest

# Using SQLite (default)
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v ./data:/data \
  calciumion/new-api:latest

# Using MySQL
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" \
  -e TZ=Asia/Shanghai \
  -v ./data:/data \
  calciumion/new-api:latest
```

> **💡 Tip:** `-v ./data:/data` will save data in the `data` folder of the current directory, you can also change it to an absolute path like `-v /your/custom/path:/data`

</details>

---

🎉 After deployment is complete, visit `http://localhost:3000` to start using!

> [!WARNING]
> When operating this project as a public generative AI service or API resale service, users should first complete all required filing, licensing, content safety, real-name verification, log retention, tax, payment, and upstream authorization obligations.

📖 For more deployment methods, please refer to [Deployment Guide](https://docs.newapi.pro/en/docs/installation)

---

## 📚 Documentation

<div align="center">

### 📖 [Official Documentation](https://docs.newapi.pro/en/docs) | [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/QuantumNous/new-api)

</div>

**Quick Navigation:**

| Category | Link |
|------|------|
| 🚀 Deployment Guide | [Installation Documentation](https://docs.newapi.pro/en/docs/installation) |
| ⚙️ Environment Configuration | [Environment Variables](https://docs.newapi.pro/en/docs/installation/config-maintenance/environment-variables) |
| 📡 API Documentation | [API Documentation](https://docs.newapi.pro/en/docs/api) |
| ❓ FAQ | [FAQ](https://docs.newapi.pro/en/docs/support/faq) |
| 💬 Community Interaction | [Communication Channels](https://docs.newapi.pro/en/docs/support/community-interaction) |

---

## ✨ Key Features

> For detailed features, please refer to [Features Introduction](https://docs.newapi.pro/en/docs/guide/wiki/basic-concepts/features-introduction)

### 🎨 Core Functions

| Feature | Description |
|------|------|
| 🎨 New UI | Modern user interface design |
| 🌍 Multi-language | Supports Simplified Chinese, Traditional Chinese, English, French, Japanese |
| 🔄 Data Compatibility | Fully compatible with the original One API database |
| 📈 Data Dashboard | Visual console and statistical analysis |
| 🔒 Permission Management | Token grouping, model restrictions, user management |

### 💰 Authorized Usage Accounting and Billing

- ✅ Internal top-up and quota allocation for lawful authorized scenarios (EPay, Stripe)
- ✅ Organization-level per-request, usage-based, and cache-hit cost accounting
- ✅ Cache billing statistics for OpenAI, Azure, DeepSeek, Claude, Qwen, and supported models
- ✅ Flexible billing policies for internal management or authorized enterprise customers

### 🔐 Authorization and Security

- 😈 Discord authorization login
- 🤖 LinuxDO authorization login
- 📱 Telegram authorization login
- 🔑 OIDC unified authentication
- 🔍 Key quota query usage (with [new-api-key-tool](https://github.com/Calcium-Ion/new-api-key-tool))

### 🚀 Advanced Features

**API Format Support:**
- ⚡ [OpenAI Responses](https://docs.newapi.pro/en/docs/api/ai-model/chat/openai/create-response)
- ⚡ [OpenAI Realtime API](https://docs.newapi.pro/en/docs/api/ai-model/realtime/create-realtime-session) (including Azure)
- ⚡ [Claude Messages](https://docs.newapi.pro/en/docs/api/ai-model/chat/create-message)
- ⚡ [Google Gemini](https://doc.newapi.pro/en/api/google-gemini-chat)
- 🔄 [Rerank Models](https://docs.newapi.pro/en/docs/api/ai-model/rerank/create-rerank) (Cohere, Jina)

**Intelligent Routing:**
- ⚖️ Channel weighted random
- 🔄 Automatic retry on failure
- 🚦 User-level model rate limiting

**Format Conversion:**
- 🔄 **OpenAI Compatible ⇄ Claude Messages**
- 🔄 **OpenAI Compatible → Google Gemini**
- 🔄 **Google Gemini → OpenAI Compatible** - Text only, function calling not supported yet
- 🚧 **OpenAI Compatible ⇄ OpenAI Responses** - In development
- 🔄 **Thinking-to-content functionality**

**Reasoning Effort Support:**

<details>
<summary>View detailed configuration</summary>

**OpenAI series models:**
- `o3-mini-high` - High reasoning effort
- `o3-mini-medium` - Medium reasoning effort
- `o3-mini-low` - Low reasoning effort
- `gpt-5-high` - High reasoning effort
- `gpt-5-medium` - Medium reasoning effort
- `gpt-5-low` - Low reasoning effort

**Claude thinking models:**
- `claude-3-7-sonnet-20250219-thinking` - Enable thinking mode

**Google Gemini series models:**
- `gemini-2.5-flash-thinking` - Enable thinking mode
- `gemini-2.5-flash-nothinking` - Disable thinking mode
- `gemini-2.5-pro-thinking` - Enable thinking mode
- `gemini-2.5-pro-thinking-128` - Enable thinking mode with thinking budget of 128 tokens
- You can also append `-low`, `-medium`, or `-high` to any Gemini model name to request the corresponding reasoning effort (no extra thinking-budget suffix needed).

</details>

---

## 🤖 Model Support

> For details, please refer to [API Documentation - Gateway Interface](https://docs.newapi.pro/en/docs/api)

| Model Type | Description | Documentation |
|---------|------|------|
| 🤖 OpenAI-Compatible | OpenAI compatible models | [Documentation](https://docs.newapi.pro/en/docs/api/ai-model/chat/openai/createchatcompletion) |
| 🤖 OpenAI Responses | OpenAI Responses format | [Documentation](https://docs.newapi.pro/en/docs/api/ai-model/chat/openai/createresponse) |
| 🎨 Midjourney-Proxy | [Midjourney-Proxy(Plus)](https://github.com/novicezk/midjourney-proxy) | [Documentation](https://doc.newapi.pro/api/midjourney-proxy-image) |
| 🎵 Suno-API | [Suno API](https://github.com/Suno-API/Suno-API) | [Documentation](https://doc.newapi.pro/api/suno-music) |
| 🔄 Rerank | Cohere, Jina | [Documentation](https://docs.newapi.pro/en/docs/api/ai-model/rerank/creatererank) |
| 💬 Claude | Messages format | [Documentation](https://docs.newapi.pro/en/docs/api/ai-model/chat/createmessage) |
| 🌐 Gemini | Google Gemini format | [Documentation](https://docs.newapi.pro/en/docs/api/ai-model/chat/gemini/geminirelayv1beta) |
| 🔧 Dify | ChatFlow mode | - |
| 🎯 Custom upstream | Supports configuring legally authorized upstream endpoints | - |

### 📡 Supported Interfaces

<details>
<summary>View complete interface list</summary>

- [Chat Interface (Chat Completions)](https://docs.newapi.pro/en/docs/api/ai-model/chat/openai/createchatcompletion)
- [Response Interface (Responses)](https://docs.newapi.pro/en/docs/api/ai-model/chat/openai/createresponse)
- [Image Interface (Image)](https://docs.newapi.pro/en/docs/api/ai-model/images/openai/post-v1-images-generations)
- [Audio Interface (Audio)](https://docs.newapi.pro/en/docs/api/ai-model/audio/openai/create-transcription)
- [Video Interface (Video)](https://docs.newapi.pro/en/docs/api/ai-model/audio/openai/createspeech)
- [Embedding Interface (Embeddings)](https://docs.newapi.pro/en/docs/api/ai-model/embeddings/createembedding)
- [Rerank Interface (Rerank)](https://docs.newapi.pro/en/docs/api/ai-model/rerank/creatererank)
- [Realtime Conversation (Realtime)](https://docs.newapi.pro/en/docs/api/ai-model/realtime/createrealtimesession)
- [Claude Chat](https://docs.newapi.pro/en/docs/api/ai-model/chat/createmessage)
- [Google Gemini Chat](https://docs.newapi.pro/en/docs/api/ai-model/chat/gemini/geminirelayv1beta)

</details>

---

## 🚢 Deployment

> [!TIP]
> **Latest Docker image:** `calciumion/new-api:latest`

### 📋 Deployment Requirements

| Component | Requirement |
|------|------|
| **Local database** | SQLite (Docker must mount `/data` directory)|
| **Remote database** | MySQL ≥ 5.7.8 or PostgreSQL ≥ 9.6 |
| **Container engine** | Docker / Docker Compose |
| **System architecture** | 64-bit only (amd64 / arm64); 32-bit systems are not supported |

### ⚙️ Environment Variable Configuration

<details>
<summary>Common environment variable configuration</summary>

| Variable Name | Description | Default Value |
|--------|------|--------|
| `SESSION_SECRET` | Authentication signing secret; must be identical on every node | - |
| `SESSION_COOKIE_SECURE` | `false`/unset disables the refresh/logout OriginGuard for local HTTP dev proxies; `true` enables the Secure cookie and strict Origin checks | `false` |
| `SESSION_COOKIE_TRUSTED_URL` | Required with Secure mode: comma-separated exact HTTPS Origins allowed to call refresh/logout; not a relay CORS allowlist | - |
| `TRUSTED_PROXIES` | Unset/blank trusts loopback, RFC 1918 and IPv6 ULA with a startup warning; `none` trusts no proxies; an explicit proxy IP/CIDR list replaces the defaults | `127.0.0.0/8, ::1, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7` |
| `USER_SESSION_ACTIVE_LIMIT` | Maximum active login Sessions per user | `50` |
| `USER_SESSION_ISSUANCE_LIMIT` | Maximum Sessions created per user within the issuance window, including revoked Sessions | `100` |
| `USER_SESSION_ISSUANCE_WINDOW_SECONDS` | Per-user Session issuance window; clamped to the revoked retention period when configured higher | `86400` |
| `USER_SESSION_REVOKED_RETENTION_DAYS` | Days to retain revoked Session rows for audit and issuance accounting | `7` |
| `USER_SESSION_HOURLY_ALERT_THRESHOLD` | Global Sessions created per hour that triggers an alert only; it never blocks login | `5000` |
| `CRYPTO_SECRET` | HMAC secret for cache keys; nodes sharing Redis must use the same effective value | Defaults to `SESSION_SECRET` |
| `SQL_DSN` | Database connection string | - |
| `REDIS_CONN_STRING` | Redis connection string | - |
| `RELAY_IDLE_CONN_TIMEOUT` | Idle keep-alive timeout for relay HTTP clients, seconds. Defaults to Go standard library behavior; set `0` to disable | `90` |
| `STREAMING_TIMEOUT` | Streaming timeout (seconds) | `300` |
| `STREAM_SCANNER_MAX_BUFFER_MB` | Max per-line buffer (MB) for the stream scanner; increase when upstream sends huge image/base64 payloads | `64` |
| `MAX_REQUEST_BODY_MB` | Max request body size (MB, counted **after decompression**; prevents huge requests/zip bombs from exhausting memory). Exceeding it returns `413` | `32` |
| `AZURE_DEFAULT_API_VERSION` | Azure API version | `2025-04-01-preview` |
| `ERROR_LOG_ENABLED` | Error log switch | `false` |
| `PYROSCOPE_URL` | Pyroscope server address | - |
| `PYROSCOPE_APP_NAME` | Pyroscope application name | `new-api` |
| `PYROSCOPE_BASIC_AUTH_USER` | Pyroscope basic auth user | - |
| `PYROSCOPE_BASIC_AUTH_PASSWORD` | Pyroscope basic auth password | - |
| `PYROSCOPE_MUTEX_RATE` | Pyroscope mutex sampling rate | `5` |
| `PYROSCOPE_BLOCK_RATE` | Pyroscope block sampling rate | `5` |
| `HOSTNAME` | Hostname tag for Pyroscope | `new-api` |

📖 **Complete configuration:** [Environment Variables Documentation](https://docs.newapi.pro/en/docs/installation/config-maintenance/environment-variables)

</details>

### 🔧 Deployment Methods

<details>
<summary><strong>Method 1: Docker Compose (Recommended)</strong></summary>

```bash
# Clone the project
git clone https://github.com/QuantumNous/new-api.git
cd new-api

# Edit configuration
nano docker-compose.yml

# Start service
docker-compose up -d
```

</details>

<details>
<summary><strong>Method 2: Docker Commands</strong></summary>

**Using SQLite:**
```bash
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v ./data:/data \
  calciumion/new-api:latest
```

**Using MySQL:**
```bash
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" \
  -e TZ=Asia/Shanghai \
  -v ./data:/data \
  calciumion/new-api:latest
```

> **💡 Path explanation:**
> - `./data:/data` - Relative path, data saved in the data folder of the current directory
> - You can also use absolute path, e.g.: `/your/custom/path:/data`

</details>

<details>
<summary><strong>Method 3: BaoTa Panel</strong></summary>

1. Install BaoTa Panel (≥ 9.2.0 version)
2. Search for **New-API** in the application store
3. One-click installation

📖 [Tutorial with images](./docs/BT.md)

</details>

### ⚠️ Multi-machine Deployment Considerations

> [!WARNING]
> - All nodes must use the same primary database and the same `SESSION_SECRET`; otherwise Access Tokens, refresh sessions, and temporary authentication flows cannot be verified consistently.
> - Nodes connected to the same Redis must also use the same `CRYPTO_SECRET`, or their cache-key digests will differ and shared entries cannot be reused consistently.

The database is authoritative for login Sessions and for the per-user active/issuance limits. Redis Session entries are short-lived caches whose TTL follows `SYNC_FREQUENCY` (60 seconds by default) and never exceeds the Session's remaining lifetime.

| Redis topology | Session propagation | Rate limiting |
| --- | --- | --- |
| Shared Redis | Revocations and version publications normally propagate immediately | Redis limits are shared across nodes |
| Independent Redis per node | Nodes converge from the database within the effective `SYNC_FREQUENCY`; a newly rotated token may receive a temporary 401 on a node with stale cache | Each node has its own allowance, so aggregate capacity can reach roughly the configured limit multiplied by the node count |
| No Redis | Every Session validation reads the database | In-memory limits are independent per node |

A shorter `SYNC_FREQUENCY` reduces the independent-Redis staleness window but causes one additional primary-key Session lookup per active SID, per node, per TTL. These guarantees make Session authentication bounded-stale across the supported topologies; rate limits and other Redis-backed control-plane caches remain topology-dependent.

See [User authentication and login sessions](./docs/authentication.md) for the token, Origin-check and PAT contracts.

### 🔄 Channel Retry and Cache

**Retry configuration:** `Settings → Operation Settings → General Settings → Failure Retry Count`

**Cache configuration:**
- `REDIS_CONN_STRING`: Redis cache (recommended)
- `MEMORY_CACHE_ENABLED`: Memory cache

---

## 🔗 Related Projects

### Upstream Projects

| Project | Description |
|------|------|
| [One API](https://github.com/songquanpeng/one-api) | Original project base |
| [Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy) | Midjourney interface support |

### Supporting Tools

| Project | Description |
|------|------|
| [new-api-key-tool](https://github.com/Calcium-Ion/new-api-key-tool) | Key quota query tool |
| [new-api-horizon](https://github.com/Calcium-Ion/new-api-horizon) | New API high-performance optimized version |

---

## 💬 Help Support

### 📖 Documentation Resources

| Resource | Link |
|------|------|
| 📘 FAQ | [FAQ](https://docs.newapi.pro/en/docs/support/faq) |
| 💬 Community Interaction | [Communication Channels](https://docs.newapi.pro/en/docs/support/community-interaction) |
| 🐛 Issue Feedback | [Issue Feedback](https://docs.newapi.pro/en/docs/support/feedback-issues) |
| 📚 Complete Documentation | [Official Documentation](https://docs.newapi.pro/en/docs) |

### 🤝 Contribution Guide

Welcome all forms of contribution!

- 🐛 Report Bugs
- 💡 Propose New Features
- 📝 Improve Documentation
- 🔧 Submit Code

---

## 📜 License

This project is licensed under the [GNU Affero General Public License v3.0 (AGPLv3)](./LICENSE).

Additional terms under AGPLv3 Section 7 apply. Modified versions must preserve
the author attribution notice `Frontend design and development by New API
contributors.` in the appropriate legal notices and in any prominent about,
legal, footer, or attribution location presented by the user interface.

Modified versions that present a user interface must also preserve a visible
link to the original project: <https://github.com/QuantumNous/new-api>.

This is an open-source project developed based on [One API](https://github.com/songquanpeng/one-api) (MIT License).

If your organization's policies do not permit the use of AGPLv3-licensed software, or if you wish to avoid the open-source obligations of AGPLv3, please contact us at: [support@quantumnous.com](mailto:support@quantumnous.com)

---

## 🌟 Star History

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=Calcium-Ion/new-api&type=Date)](https://star-history.com/#Calcium-Ion/new-api&Date)

</div>

---

<div align="center">

### 💖 Thank you for using New API

If this project is helpful to you, welcome to give us a ⭐️ Star！

**[Official Documentation](https://docs.newapi.pro/en/docs)** • **[Issue Feedback](https://github.com/Calcium-Ion/new-api/issues)** • **[Latest Release](https://github.com/Calcium-Ion/new-api/releases)**

<sub>Built with ❤️ by QuantumNous</sub>

</div>

</details>
