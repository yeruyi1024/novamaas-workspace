<div align="center">

<img src="./web/public/logo.png" alt="NovaMaaS logo" width="88" />

# 星枢 MaaS 平台

**NovaMaaS — 面向 AI Token 供应聚合、商业分销与算力资源运营的统一基础设施**

[![Build, package and publish](https://github.com/yeruyi1024/novamaas-workspace/actions/workflows/build.yml/badge.svg)](https://github.com/yeruyi1024/novamaas-workspace/actions/workflows/build.yml)
[![License: AGPL v3](https://img.shields.io/badge/license-AGPL--3.0-3157d5.svg)](LICENSE)
[![Upstream](https://img.shields.io/badge/upstream-QuantumNous%2Fnew--api-7357d9.svg)](https://github.com/QuantumNous/new-api)

[快速开始](#快速开始) · [主要能力](#主要能力) · [与上游差异](#novamaas-与上游差异) · [项目文档](#项目文档)

</div>

> [!IMPORTANT]
> **项目来源与致谢**
>
> 星枢 MaaS 平台（NovaMaaS）基于 **[QuantumNous/new-api](https://github.com/QuantumNous/new-api)** 建设，初始基线严格对应上游 **[v1.0.0-rc.26](https://github.com/QuantumNous/new-api/releases/tag/v1.0.0-rc.26)** 和提交 **`8f6961c675932f406260ff0c218bc2aa0603e9b2`**，并完整保留上游 Git 历史。
>
> 感谢 **QuantumNous、new-api 项目作者及所有贡献者**提供的开源基础与持续投入。NovaMaaS 将遵循上游 AGPL-3.0 许可及附加声明，持续保留原作者署名、项目来源和许可证信息。

**NovaMaaS Workspace is an independently maintained fork of [QuantumNous/new-api](https://github.com/QuantumNous/new-api), initially based on exactly `v1.0.0-rc.26`. The repository retains its upstream Git history, license notices, attribution, and provenance records.**

## 快速开始

克隆本仓库并使用 Docker 构建当前源码。默认使用 SQLite；数据保存在 `novamaas-data` 卷中。

```bash
git clone https://github.com/yeruyi1024/novamaas-workspace.git
cd novamaas-workspace
docker build -t novamaas:local .
docker run -d --name novamaas --restart unless-stopped \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v novamaas-data:/data \
  novamaas:local
```

打开 [http://localhost:3000](http://localhost:3000) 完成初始化。升级前请备份数据卷。仓库根目录的 `docker-compose.yml` 仍使用上游镜像 `calciumion/new-api:latest`；运行本仓库请使用上述命令，或在 Compose 中明确替换为本仓库镜像。

也可从 [GitHub Packages](https://github.com/yeruyi1024/novamaas-workspace/pkgs/container/novamaas-workspace) 复制已发布的固定标签，使用 `ghcr.io/yeruyi1024/novamaas-workspace:build_YYYYMMDDTHHMMSSZ_multiarch` 替代上方的 `novamaas:local`。镜像权限、版本标签与生产部署说明见 [构建文档](docs/BUILD.zh_CN.md)。

## 主要能力

- **统一模型网关**：接入多家模型供应商，提供协议转换、模型路由、权限、额度与用量管理。
- **多模态与视频任务**：支持火山方舟、DoubaoVideo 等渠道的视频任务、模型映射、请求审计与媒体交付；通过 Webhook 推送任务状态和响应。
- **素材与对象存储**：私有素材库、下游 Action API，以及对 Base64 图片和视频输入的临时存储转换。
- **商业运营**：渠道成本与利润视图、客户消费查询、正式记账和月度对账凭证。
- **客户端与身份**：浏览器管理界面、独立微信小程序及移动登录能力。

已交付范围、上线条件和后续路线图见 [平台能力说明](docs/PLATFORM_OVERVIEW.zh_CN.md)。

## NovaMaaS 与上游差异

以下是相对 [QuantumNous/new-api](https://github.com/QuantumNous/new-api) 需要长期维护的主要运行时差异；完整 PR 台账保留在下方，供上游同步时核对。

| 领域 | 主要区别 |
| --- | --- |
| 模型与视频协议 | 火山原生渠道、视频任务审计与交付，以及 Moonshot/Kimi 兼容入口和用量结构。 |
| 素材与存储 | 私有素材库、下游 Action API、渠道副本同步、Base64 媒体暂存和请求快照异步归档。 |
| 计费与对账 | 渠道成本快照、利润视图、正式消费账本和月度凭证。 |
| 用户与客户端 | 手机号身份、微信小程序登录、多媒体对话及受控模型能力。 |
| 运行观测 | 当前节点资源、分类流量、活跃请求与连接数；有界纯内存历史及低开销采集。 |

<details>
<summary>查看完整的长期运行时差异台账</summary>

此处不是完整 Changelog，而是 NovaMaaS 相对 `QuantumNous/new-api` 的**关键、长期运行时差异**清单，用于上游同步时判断哪些能力必须保留、重做或移除。PR 只有同时满足以下条件才收录：

1. 改变核心产品运行时能力，例如供应商接入、请求/响应协议、模型路由、计费、权限、数据兼容或租户能力。
2. 上游主线尚无等价实现，或 NovaMaaS 在上游实现之外保留了有意义的行为与安全边界。
3. 差异具有长期维护价值；未来同步上游时忽略它，可能导致功能回退、兼容性故障或业务语义变化。

CI/CD、镜像发布、构建环境、首页展示、文档整理、测试补充、内部重构、依赖升级、临时排障，以及未合并或已被替代的方案不进入本表，除非它们同时改变上述核心运行时边界。

<!-- novamaas-pr-ledger:start -->
| 关键差异 PR | 日期 | 类型 | 领域 | 关键变化 | 与上游关系 |
| --- | --- | --- | --- | --- | --- |
| [#79](https://github.com/yeruyi1024/novamaas-workspace/pull/79) | 2026-10-17 | `fix/feat` | Anthropic 用量 / 历史 Token 核算 | OpenAI → Anthropic 输出普通输入与缓存读写分离，保留原始计费快照；管理员可按客户、Key、模型、协议及时间只读核算历史 Token 与两种历史价格，在明确的旧版原生缓存写入字段假设下，仅重算无写入请求的读取重叠，导出记录依据且不自动调整钱包。 | 协议规则选择性同步上游提交 `0ed497f066`；管理员历史核算为 NovaMaaS 下游专属，上游同步时须保留协议语义与内部计费分离、未知历史凭证不猜测及只读核算边界。 |
| [#74](https://github.com/yeruyi1024/novamaas-workspace/pull/74) | 2026-10-16 | `feat/perf` | 节点资源 / 分类流量 / 系统信息 | Root 面板展示 CPU、内存、网卡、请求与连接历史，按 HTTP 入站/出站、SQL、Redis 及固定业务类别统计；流式分片计数即时可见，默认 24 小时有界内存历史，重启清空，启动时可完整旁路采集。 | NovaMaaS 下游专属；上游无等价的分类流量与纯内存历史组合实现，同步时须保留实际读写字节、Root 权限、无监控持久化、关闭旁路及原有性能保护边界。 |
| [#72](https://github.com/yeruyi1024/novamaas-workspace/pull/72) | 2026-10-14 | `feat/fix` | 请求快照 / 对象存储 / 历史兼容 | Ali、DoubaoVideo、VolcNative 原始与上游 JSON 经节点持久暂存和有限 Worker 异步归档，数据库保留对象索引及冻结计费证据；独立存储生命周期、管理员永久读取入口、历史双读与核验后 CAS 清理，有界并发迁移及轻量无大正文排序查询。 | 既有视频任务审计、存储与计费修正能力的 NovaMaaS 下游扩展；上游无等价实现，同步时须保留节点交接恢复、独立快照生命周期、鉴权读取、证据优先级及验证后清理边界。 |
| [#69](https://github.com/yeruyi1024/novamaas-workspace/pull/69) | 2026-10-12 | `feat` | PC 生图 / OSS 交付 / 图片权限 | PC 生图复用现有 Relay，创作页面公开访问，模型调用与个人记录需登录；生成文件独立保存到私有 OSS，30 天后自动清理，签名地址不超过文件保留期限，到期仍保留生成及审计记录；手动入素材库的副本独立保留，并支持联合删除；跨用户图片巡查限 Root 或授权管理员，请求快照沿用主分支普通管理员读取权限。 | NovaMaaS 下游专属；上游无等价的生图持久记录、私有 OSS 交付及授权巡查组合，后续同步须保留公开页面与私有模型调用的边界、结果归属、权限隔离、30 天文件清理与独立素材库副本的边界。 |
| [#61](https://github.com/yeruyi1024/novamaas-workspace/pull/61) | 2026-10-09 | `fix` | 任务计价日志 / 历史数据兼容 | 提交、结算、退款和调账日志统一保存并展示精确分档倍率；历史提交日志仅在用户、任务与计价身份一致时读取原任务证据补充返回值，空倍率快照覆盖旧值，保持原日志、账目和余额不变。 | #57 的 NovaMaaS 下游扩展；上游无等价有效调账投影及旧任务计价证据恢复，同步时须保留基础倍率与分档倍率分离、空快照优先级及只读恢复边界。 |
| [#60](https://github.com/yeruyi1024/novamaas-workspace/pull/60) | 2026-10-09 | `fix` | SQLite / 历史成本迁移 | 旧成本调整表先新增可空事件列，再创建唯一索引；保留历史数据并支持重复启动，避免 SQLite 添加唯一列导致启动失败。 | #59 的 NovaMaaS 下游兼容修复；上游无等价成本调整表，同步时须保留旧库升级与事件唯一性约束。 |
| [#59](https://github.com/yeruyi1024/novamaas-workspace/pull/59) | 2026-10-09 | `feat` | 模型调账 / 历史成本 / 利润核算 | 调账按模型分档价格及独立渠道成本折扣追加原快照成本更正，销售差额为零也可修复成本且不额外扣钱包；原月份日志及利润汇总读取有效成本，后续折扣重算沿用修正基数；冻结成本证据、原子确认及依赖撤销，保留旧批次兼容。 | #57、#31 的 NovaMaaS 下游扩展；上游无等价实现，同步时须保留不可变成本凭证、累计计价取整、同事务调账、成本不重复扣钱包及有效基数重算边界。 |
| [#57](https://github.com/yeruyi1024/novamaas-workspace/pull/57) | 2026-10-09 | `feat` | 有效消费账目 / 模型计费调账 / Seedance 2.5 | 保留原消费、退款及成本凭证，以修正分录投影用户日志、统计和原月份账目；新增按当前模型价格与历史分辨率、视频输入、用量重算，按净差额原子扣退余额并支持幂等确认及逆序撤销；历史对账单作废后重建，独立审计防重复汇总；两种 Seedance 2.5 模型名接入分辨率和视频输入分档。 | #55、#56 账务与财务权限能力的 NovaMaaS 下游扩展；上游无等价实现，同步时须保留不可变原证据、有效金额读模型、钱包差额事务及分档计费边界。 |
| [#56](https://github.com/yeruyi1024/novamaas-workspace/pull/56) | 2026-10-09 | `feat` | 月度对账 / 永久凭证存储 / 财务权限 | 新账单证据按 `billing/user-{用户HMAC标识}/{记账月份}/{账单UUID}/…` 归档，与 assets 复用同一用户标识，保留不可覆盖上传及历史对象读取；跨账户对账须同时具备管理员角色与财务核算权限，财务人员可查看调账记录，调账写操作仍限超级管理员。 | #55 永久账单归档与调账能力的 NovaMaaS 下游扩展；上游无等价实现，同步时须保留按客户与账期归档、历史对象读取及财务权限边界。 |
| [#55](https://github.com/yeruyi1024/novamaas-workspace/pull/55) | 2026-10-08 | `feat` | 月度对账 / 永久凭证 | PDF 与 Excel 共用冻结正式账本并永久归档，PDF 和清单绑定 Excel 指纹；模型明细按实际消费组和历史费率拆行并冻结，PDF 与 Excel 移除原价列；超级管理员可预览、幂等执行及撤销固定单价钱包任务的费率纠错，保留原证据并仅在管理员日志显示；新账单将差额归入原消费/退款账期，防止操作月份重复计入；同一客户同月最多一份有效账单，管理员先作废原版再按最新账务创建草稿，保留历史导入凭证、旧文件与确认，新版重新下发确认。 | #22、#43 正式对账与品牌能力的 NovaMaaS 下游扩展；上游无等价双文件联合确认、账单作废后重建与历史费率差额纠错，同步时须保留不可变证据及原范围边界。 |
| [#54](https://github.com/yeruyi1024/novamaas-workspace/pull/54) | 2026-10-08 | `fix` | 用户短信核验 / 正式对账 | 已登录管理员直接发送目标手机号验证码，保存时原子核验；历史补录仅覆盖正式起点及之后的消费/退款，保持起点、起始序号和钱包余额不变，起点前消费不阻止零消费正式单。 | #50 短信认证与既有正式对账能力的 NovaMaaS 下游调整；上游没有等价的管理员短信保存核验和正式账本补录链，同步时需保留时间边界及防重复补录校验。 |
| [#52](https://github.com/yeruyi1024/novamaas-workspace/pull/52) | 2026-10-07 | `fix` | 微信手机号认证 | 保留服务端数字 errcode 与兑换阶段，区分授权码无效和其他拒绝；仅明确 access_token 失效时刷新重试一次，阻止微信错误进入密码登录页面。 | #50、#51 的 NovaMaaS 下游扩展；上游没有等价手机号兑换链，同步时需保留诊断脱敏和不可盲重试授权码的边界。 |
| [#51](https://github.com/yeruyi1024/novamaas-workspace/pull/51) | 2026-10-07 | `feat` | 登录认证 / 微信身份联合 | 浏览器通过微信网站应用可信 UnionID 登录唯一已绑定平台账号，拒绝未绑定、失效或冲突身份且不自动注册；统一绑定安全确认与登录入口，移除通用日志财务列表列。 | #50 的 NovaMaaS 下游扩展；上游当前没有等价的微信网站应用与小程序已绑定身份联合登录边界，后续同步需保留或重新评估。 |
| [#50](https://github.com/yeruyi1024/novamaas-workspace/pull/50) | 2026-10-06 | `feat` | 用户 / 短信认证 / 微信小程序 | 通过短信核验 +86 手机号绑定与换绑，提供短信登录和密码登录后的短信替代 MFA；微信服务端手机号仅匹配已核验平台账号，经确认建立身份绑定，并支持资料展示、受控解绑及会话失效。 | #25、#33、#41 的 NovaMaaS 下游扩展；上游当前没有等价的短信手机号核验、微信手机号匹配确认与受控解绑组合实现。 |
| [#48](https://github.com/yeruyi1024/novamaas-workspace/pull/48) | 2026-10-07 | `feat/fix` | 素材审核 / 媒体任务通知 | 保留素材投递租约接管防护，按类别隔离个人与系统 Webhook；任务状态和素材审核事件经持久队列异步投递，管理员控制类别开关、全量地址及客户手册下载，并支持即时连通性测试。 | #46 的 NovaMaaS 下游扩展；上游当前没有等价的跨素材与媒体任务通知、系统级订阅和配置控制组合实现。 |
| [#47](https://github.com/yeruyi1024/novamaas-workspace/pull/47) | 2026-09-30 | `feat` | 财务核算 / 通用日志 | 有财务核算查看权限时，逐条日志关联不可变成本快照及追加调整，展示营业额、成本与利润；无权限视图不返回这些字段，缺失快照的历史记录不推断成本。 | #31 财务核算能力的 NovaMaaS 下游展示与权限补全；上游当前没有等价的渠道成本快照和逐条日志财务视图。 |
| [#46](https://github.com/yeruyi1024/novamaas-workspace/pull/46) | 2026-09-29 | `feat` | 素材审核 / 状态通知 | 持续复核已可用素材的上游审核状态，同步明确拒绝结果并拦截不可用素材；用户在平台配置 HTTPS Webhook，状态与通知事务保存，支持失败重试及版本去重。 | NovaMaaS 下游专属；扩展现有素材同步机制，上游当前没有等价的延迟复审对账与客户状态推送组合实现。 |
| [#43](https://github.com/yeruyi1024/novamaas-workspace/pull/43) | 2026-09-28 | `feat` | 计费 / 对账 PDF / 供应商测试 | 新增运营主体品牌配置，将裁切后的 Logo 固定在新版对账单快照，并统一普通与视频测试报告的 PDF 页头；管理员可设置过去或远期的记账起点，历史用量仍需核验导入。 | NovaMaaS 下游专属；上游当前没有等价的正式对账快照、运营主体品牌及供应商报告统一输出能力。 |
| [#41](https://github.com/yeruyi1024/novamaas-workspace/pull/41) | 2026-09-27 | `feat` | 微信小程序 / 登录认证 / 多媒体对话 | 增加经微信服务端验证的手机号归属与系统分配号码过渡、独立账号密码及 2FA 登录回退；图片输入、语音转写和合成经后台模型能力配置、用户鉴权与现有渠道计费链处理。 | NovaMaaS 下游专属；上游当前没有等价的小程序手机号安全归属、独立移动登录回退和受控多媒体对话组合实现。 |
| [#39](https://github.com/yeruyi1024/novamaas-workspace/pull/39) | 2026-09-24 | `feat` | 素材库 / 渠道请求日志 / 任务日志 | 素材组搜索分页、素材预览下载和上游审核拒绝后的不可用状态及视频请求反馈；渠道测试与同步请求采用有界异步 MySQL 日志及脱敏详情，任务日志支持用户名和模型检索。 | NovaMaaS 下游专属；上游当前没有等价的素材审核状态联动、渠道请求诊断及此任务日志检索组合实现。 |
| [#37](https://github.com/yeruyi1024/novamaas-workspace/pull/37) | 2026-09-23 | `feat` | 素材库 / 下游兼容 API | 新增平台签发并加密保存的用户级素材库 AK/SK，在根路径和 `/api/v3/` 提供火山方舟同形态的 HMAC-SHA256 V4 Action API；支持公网 URL 安全导入、租户隔离、下游密钥自助管理和大规模素材分页检索。 | NovaMaaS 下游专属；上游当前没有面向下游客户的平台素材库 AK/SK、同路径 Action API、URL 导入与租户隔离组合实现。 |
| [#36](https://github.com/yeruyi1024/novamaas-workspace/pull/36) | 2026-09-22 | `feat` | 素材库 / 对象存储 / 视频渠道 | 新增网关自有永久素材库、租户权限与签名预览，按渠道维护上游副本和同步任务；支持 Volcengine Action AK/SK、Bearer 及 YooFang REST Bearer SK，并在 DoubaoVideo 和火山原生请求中将我方素材 ID 实时翻译为对应渠道 ID。 | NovaMaaS 下游专属；上游当前没有等价的自有素材库、多渠道副本同步、加密渠道凭据与请求时 ID 映射组合实现。 |
| [#35](https://github.com/yeruyi1024/novamaas-workspace/pull/35) | 2026-09-21 | `feat` | 渠道诊断 / 视频任务可观测性 | 新增复用渠道代理与 HTTP 配置的分阶段网络探测，并为 Doubao Video、火山原生和阿里百炼持久化请求体读取、请求准备、临时存储转换、上游请求及总耗时，在任务日志中分开展示。 | NovaMaaS 下游专属；上游当前没有等价的渠道 DNS/TCP/TLS/TTFB 探测与视频请求全链路指标组合实现。 |
| [#34](https://github.com/yeruyi1024/novamaas-workspace/pull/34) | 2026-09-21 | `fix` | 计费 / 财务核算 | 按来源日志识别异步任务实时快照与历史回填快照，阻止重复成本凭证；汇总和明细以最早的不可变快照为准，历史更正继续使用追加式调整。 | #31 的 NovaMaaS 下游正确性修复；上游当前没有等价的渠道成本快照与历史回填核算能力。 |
| [#33](https://github.com/yeruyi1024/novamaas-workspace/pull/33) | 2026-09-21 | `feat` | 用户 / 微信小程序 | 新增独立 Taro 微信小程序客户端、小程序专属外部身份与可轮换移动会话，复用现有权限和计费规则；AppSecret、OpenID 与 session_key 不进入客户端，公开状态仅暴露登录就绪状态。 | NovaMaaS 下游专属；上游当前没有等价的微信小程序身份、移动会话与完整客户端组合实现。 |
| [#32](https://github.com/yeruyi1024/novamaas-workspace/pull/32) | 2026-09-20 | `feat` | 模型协议 / Kimi KVV | 为 OpenAI 渠道的 `/moonshot/v1/chat/completions` 增加 Kimi 兼容透传模式，保留上游 Kimi 请求与 JSON/SSE 扩展字段、原始 usage 和模型映射语义，不向普通 OpenAI 上游伪造 KVV 能力。 | #28 Moonshot 兼容入口的 NovaMaaS 下游扩展；上游当前没有等价的 OpenAI 渠道 Kimi 透传模式。 |
| [#31](https://github.com/yeruyi1024/novamaas-workspace/pull/31) | 2026-09-20 | `feat` | 计费 / 财务核算 | 新增渠道级上游成本折扣、不可变成本快照与追加式调整、历史回填重算、财务核算权限，以及日志、首页和账单中的营业额、成本与利润视图；未配置成本折扣时按成本等于营业额处理。 | NovaMaaS 下游专属；上游当前没有将渠道成本配置、历史成本证据、权限隔离和利润报表组合起来，同时保持客户售价与钱包扣费不变的等价实现。 |
| [#30](https://github.com/yeruyi1024/novamaas-workspace/pull/30) | 2026-09-18 | `fix` | 模型协议 / Moonshot 兼容 | 补齐 Kimi Chat `prompt_tokens_details`、Responses 缓存与推理明细，以及 Messages `cache_creation` 分档字段；区分缺失值与显式零值，不生成未知用量。 | #28 的下游兼容性补充；上游当前没有跨 OpenAI、Responses 与 Anthropic Messages 输出 Kimi 用量结构的等价实现。 |
| [#28](https://github.com/yeruyi1024/novamaas-workspace/pull/28) | 2026-09-18 | `feat` | 模型协议 / Moonshot 兼容 | 新增同域名 `/moonshot` Chat Completions、Responses、Anthropic Messages 与模型列表入口，复用 OpenAI 上游并输出 Kimi 兼容结构；无法真实复现的 Kimi 专属语义明确拒绝。 | NovaMaaS 下游专属；上游当前没有以 OpenAI 渠道为数据源、按独立路径输出 Kimi 格式并拒绝伪造专属能力的等价实现。 |
| [#26](https://github.com/yeruyi1024/novamaas-workspace/pull/26) | 2026-09-11 | `feat` | 视频任务 / 请求审计 | 为 DoubaoVideo、火山原生和阿里百炼分别归档客户端原始请求与实际上游请求，并在管理员日志详情中对照展示；Base64 暂存场景记录转换后的地址。 | NovaMaaS 下游专属；上游当前没有视频任务双请求快照、暂存后正文审计与管理员对照查看的等价实现。 |
| [#25](https://github.com/yeruyi1024/novamaas-workspace/pull/25) | 2026-09-10 | `feat` | 用户 / 登录认证 | 为用户增加可维护且全局唯一的手机号，并支持手机号密码登录与认证版本失效；同步管理界面与登录文案。 | NovaMaaS 下游专属；上游当前没有等价的手机号身份字段、唯一性保护及密码登录组合能力。 |
| [#23](https://github.com/yeruyi1024/novamaas-workspace/pull/23) | 2026-09-10 | `feat` | 对象存储 / 火山方舟视频 | 将 Base64 暂存接入 DoubaoVideo，并把火山原生与 DoubaoVideo 的暂存范围从图片扩展到 MP4、WebM、MOV 视频输入；保留签名 URL、重试复用、任务清理和请求审计边界。 | NovaMaaS 下游专属；上游当前没有等价的 DoubaoVideo Base64 暂存及双渠道视频 Data URI 对象存储转换能力。 |
| [#22](https://github.com/yeruyi1024/novamaas-workspace/pull/22) | 2026-09-10 | `feat/perf` | 计费 / 客户对账 / 永久凭证 | 新增持久钱包结算与小时账本、正式记账起点及企业主体、管理员下发和客户确认、受控历史核验导入；以私有 OSS 固化明细、版本化 PDF 与清单，禁止清理使用日志。 | NovaMaaS 下游专属；上游当前没有等价的正式消费账本、历史导入确认与不可变月度凭证组合能力。 |
| [#21](https://github.com/yeruyi1024/novamaas-workspace/pull/21) | 2026-09-09 | `feat/fix/perf` | 视频任务 / 日志 / 登录与审计 | 为阿里百炼增加任务详情实时拉取；公开视频默认直连资源方并保留 `/content` 代理兼容模式；登录会话调整为 24 小时；新增系统公告、视频生成与违规统计、强制知晓及设备指纹审计；修复普通用户使用日志字段丢失。 | NovaMaaS 下游专属；其中用量统计修复同步上游提交 [`8c8c4153d`](https://github.com/QuantumNous/new-api/commit/8c8c4153d4b80d54352d21593de41aa9a6178f7e)。 |
| [#20](https://github.com/yeruyi1024/novamaas-workspace/pull/20) | 2026-09-08 | `perf` | 日志 / 任务审计 | 将任务请求体从 `logs.other` 与 `tasks.properties` 迁移至独立归档表，保留管理员按需查看，并通过可恢复批处理清理历史热表载荷。 | NovaMaaS 下游专属；上游当前没有独立请求体归档、按需审计读取与历史迁移组合能力。 |
| [#19](https://github.com/yeruyi1024/novamaas-workspace/pull/19) | 2026-09-08 | `feat` | 视频任务 / 计费 | 为 Doubao Seedance 2.0 增加稳定公开模型名，使其可映射到不同上游模型 ID，同时复用 720p、1080p、4K 与视频输入计费倍率。 | NovaMaaS 下游专属；上游当前没有该稳定公开别名及其参数计费映射。 |
| [#18](https://github.com/yeruyi1024/novamaas-workspace/pull/18) | 2026-09-08 | `fix` | 对象存储 / 数据兼容 | 使用 GORM 方言感知条件引用存储策略 `key` 列，修复 MySQL 1064 错误，并增加 SQLite、MySQL、PostgreSQL 查询回归测试。 | NovaMaaS 下游专属；属于 #17 对象存储能力的兼容性修复，上游当前没有等价的存储策略实现。 |
| [#17](https://github.com/yeruyi1024/novamaas-workspace/pull/17) | 2026-09-08 | `feat` | 对象存储 / 火山方舟 | 新增系统级存储 Profile 与用途 Policy，以私有阿里云 OSS 签名地址兼容火山原生 Base64 媒体请求，并实现任务终态清理、请求审计分流及管理员访问控制。 | NovaMaaS 下游专属；上游当前没有等价的火山 Base64 暂存、通用对象存储策略与审计分流组合实现。 |
| [#14](https://github.com/yeruyi1024/novamaas-workspace/pull/14) | 2026-09-07 | `feat` | 视频任务 / 使用日志 | 为 DoubaoVideo 增加 302 重定向与 600 秒服务端代理模式，记录 DoubaoVideo、原生 Ark、阿里百炼视频任务请求体，并在鉴权日志详情中按数据可用性提供视频下载和格式化 JSON 查看入口。 | NovaMaaS 下游专属；上游当前没有等价的视频交付模式与任务请求审计组合实现。 |
| [#11](https://github.com/yeruyi1024/novamaas-workspace/pull/11) | 2026-09-06 | `fix` | 阿里百炼 | 兼容 Wan3 任务结果中的整数、小数和数字字符串时长，恢复异步任务状态更新并增加适配器回归测试。 | 对齐上游 [#6166](https://github.com/QuantumNous/new-api/issues/6166) / [#6174](https://github.com/QuantumNous/new-api/pull/6174)，并增加非法值和溢出保护。 |
| [#8](https://github.com/yeruyi1024/novamaas-workspace/pull/8) | 2026-09-06 | `feat` | 火山方舟 | 为 Volc Native 增加仅改写顶层 `model` 的模型映射，平台侧继续使用公开别名完成权限、计费和日志。 | #1 的下游增强；上游 [#6653](https://github.com/QuantumNous/new-api/pull/6653) 尚未覆盖该映射能力。 |
| [#1](https://github.com/yeruyi1024/novamaas-workspace/pull/1) | 2026-09-05 | `feat/fix` | 火山方舟 | 选择性引入 Volc Native 渠道，并补齐任务凭据延续、取消状态、响应关闭、路由隔离、权限约束和多语言支持。 | 来源为仍未合并的上游 [#6653](https://github.com/QuantumNous/new-api/pull/6653) / [#4705](https://github.com/QuantumNous/new-api/issues/4705)，NovaMaaS 追加安全与兼容加固。 |
<!-- novamaas-pr-ledger:end -->

维护方式：台账由作者和评审按长期运行时差异标准人工维护，不作为 PR 或 CI 的合并门禁。记录使用真实 PR 链接，不再维护容易过期的审核状态；上游同步类 PR 仍需更新 [UPSTREAM.md](UPSTREAM.md)。

</details>

## 项目文档

| 文档 | 用途 |
| --- | --- |
| [平台能力与路线图](docs/PLATFORM_OVERVIEW.zh_CN.md) | 已交付能力、详细边界、路线图与上游维护方式 |
| [节点资源与分类流量监控](docs/design/RUNTIME_MONITORING.zh_CN.md) | 分类定义、内存窗口、采样与关闭配置、性能数据及验收截图 |
| [UPSTREAM.md](UPSTREAM.md) | 上游基线、同步记录、来源提交与维护策略 |
| [上游版本选择性同步](docs/UPSTREAM_SYNC.zh_CN.md) | GitHub Actions 创建 WIP PR、人工启动 AI 接管、确认计划后适配与 CI 验证 |
| [docs/BUILD.zh_CN.md](docs/BUILD.zh_CN.md) | 本地构建、CI、安装包、GHCR 与腾讯云 CCR 发布说明 |
| [docs/VOLC_NATIVE.zh_CN.md](docs/VOLC_NATIVE.zh_CN.md) | 火山方舟原生 API 渠道、任务接口和兼容性边界 |
| [docs/ASSET_LIBRARY_API.zh_CN.md](docs/ASSET_LIBRARY_API.zh_CN.md) | 素材库、下游 AK/SK、火山 Action API 兼容与大列表性能边界 |
| [媒体任务与素材审核 Webhook](docs/MEDIA_TASK_WEBHOOKS.zh_CN.md) | 个人与系统通知、实时测试、品牌化 PDF 手册、异步投递 |
| [对账单一期实施说明](docs/design/BILLING_STATEMENTS_IMPLEMENTATION.zh_CN.md) | 实际交付范围、记账/归档机制、迁移和部署边界 |
| [对账单 PDF/Excel 导出说明](docs/design/BILLING_STATEMENT_EXPORTS.zh_CN.md) | 附件字段支持情况、历史实际费率拆分、双文件指纹与作废后创建版本规则 |
| [历史消费与调账实施记录](docs/design/BILLING_CORRECTIONS.zh_CN.md) | 有效金额展示、渠道/模型计费调账、钱包差额、历史账单作废重建与 Seedance 2.5 分档 |
| [对账单验收指南](docs/design/BILLING_STATEMENTS_ACCEPTANCE.zh_CN.md) | 历史查询、正式记账、创建草稿、下发与客户确认操作 |
| [历史核验与统计说明](docs/design/BILLING_HISTORY_REVIEW.zh_CN.md) | 历史导入确认边界、空单防护、MySQL 聚合与 PDF 固化 |
| [对账产品方案](docs/design/BILLING_STATEMENTS_PRD.zh_CN.md) / [技术方案](docs/design/BILLING_STATEMENTS_TECH.zh_CN.md) / [数据专项方案](docs/design/BILLING_STATEMENTS_DATA_PIPELINE.zh_CN.md) | 设计基线与后续演进建议；不代表全部能力已实现 |
| [LICENSE](LICENSE) | AGPL-3.0 许可证与适用条款 |
| [NOTICE](NOTICE) | 上游声明、署名和附加许可说明 |
| [THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md) | 第三方依赖许可证信息 |

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
| `STORAGE_CREDENTIAL_ENCRYPTION_KEY` | Stable AES-GCM master key for static object-storage credentials; must be identical on every node and across restarts | Falls back to `CRYPTO_SECRET`, then `SESSION_SECRET` |
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
