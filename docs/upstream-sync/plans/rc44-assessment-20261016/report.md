# rc.26 → rc.44：AI 差异评估与待确认计划

**当前阶段：AI 已完成静态差异分析，等待维护者确认选定项。301 项 decision 全部为 pending；未实施任何上游功能。**

- 上游：[v1.0.0-rc.44](https://github.com/QuantumNous/new-api/releases/tag/v1.0.0-rc.44)，固定 `633b4c5b190dc1af99a2814391354d86948e7943`。
- 范围：`8f6961c675932f406260ff0c218bc2aa0603e9b2..633b4c5b190dc1af99a2814391354d86948e7943`；初始 fork 基线保持不变。
- 下游：`b2cb1c2473f1db803ff5bdaafc67699288bde713`；当前产品版本 `v1.0.0-rc.26-novamaas.2` 保持不变。
- `v1.0.0-rc.26-novamaas.3` 只是获批实施后的建议版本，目前没有更新 VERSION 或 UPSTREAM.md。
- 共 301 个来源提交，175 项与下游修改路径重叠，38 项涉及受保护路径；这些是审查信号，不是行为兼容结论。
- 301 项均有逐项 AI 建议，共核验 547 处固定 Git blob 行证据；每个采用、迁移、升级或 covered 建议都有下游代码对照。
- AI 建议写入 plan.json 的 ai_analysis；人工审批字段、来源、机械证据和依赖字段均未更改。

## 建议分布（不是已批准的处理计划）

| 建议 | 数量 |
| --- | ---: |
| `merge`：整项采用建议 | 11 |
| `migrate`：选择性适配建议 | 81 |
| `upgrade`：依赖/协同升级建议 | 13 |
| `covered`：已有等价行为 | 3 |
| `defer`：暂缓，需补设计或证据 | 168 |
| `skip`：本批不采用建议 | 25 |

## 建议首批优先评审的小范围修复

以下 18 项仅供逐项选择，不构成默认采用。merge 建议已做只读补丁应用检查；补丁可应用不证明行为兼容。migrate 需要按下游现有实现适配，确认时必须补全精确源码和回归测试白名单。

| 来源条目 | 功能 | 建议 |
| --- | --- | --- |
| [commit-0f2a2075ab07](https://github.com/QuantumNous/new-api/commit/0f2a2075ab072ea7e20ffa5dd5d58dbf1b6b5b22) | 请求校验错误状态码 | `merge` |
| [commit-0bee5d441029](https://github.com/QuantumNous/new-api/commit/0bee5d4410296e972bf0076414ade786c2c799c8) | 阿里图片输出格式 | `merge` |
| [commit-2cf177ac487e](https://github.com/QuantumNous/new-api/commit/2cf177ac487e62c627c7d423b65735ba2481ef4f) | RawMessage 请求深拷贝 | `merge` |
| [commit-876903a8eb22](https://github.com/QuantumNous/new-api/commit/876903a8eb22c44e395c03da38f6701c650651ae) | 火山方舟模型列表端点 | `merge` |
| [commit-76f7dafd2b82](https://github.com/QuantumNous/new-api/commit/76f7dafd2b82fc305533ce6a3065b3de2d338bae) | 音频扩展名大小写 | `merge` |
| [commit-8e5e09166cf8](https://github.com/QuantumNous/new-api/commit/8e5e09166cf83731df1819fd1cb9c50ea2576ed6) | Ollama 终止帧工具调用 | `merge` |
| [commit-3e8c358da09b](https://github.com/QuantumNous/new-api/commit/3e8c358da09b4598a946712db8219ff44f197475) | 智谱 reasoning_effort 保留 | `merge` |
| [commit-13c8e90b8f4b](https://github.com/QuantumNous/new-api/commit/13c8e90b8f4b929aa88c2e9616efc462e0068d72) | Responses image_url 转 Chat 对象 | `merge` |
| [commit-ba0e478e061e](https://github.com/QuantumNous/new-api/commit/ba0e478e061ee0c824922d89c05620de878103cf) | Claude 工具调用伴随答案文本 | `merge` |
| [commit-692e8d6ee6a9](https://github.com/QuantumNous/new-api/commit/692e8d6ee6a9a1620c2d731cb51a1e3154a7042b) | 管理员内置身份解绑类型修复 | `migrate` |
| [commit-98d50d5383a3](https://github.com/QuantumNous/new-api/commit/98d50d5383a33432ff6c30b129461b170e5cbffc) | 页面刷新后的初始化状态重验 | `migrate` |
| [commit-06dd76a9d38e](https://github.com/QuantumNous/new-api/commit/06dd76a9d38eb0eaf8e900b1264b500250156c4d) | Playground 会话数据按用户隔离 | `migrate` |
| [commit-2506e1b980ec](https://github.com/QuantumNous/new-api/commit/2506e1b980ecf67c829ef20255a16fc61c47cb55) | 创建用户拒绝非标准角色值 | `migrate` |
| [commit-dfd3cd893044](https://github.com/QuantumNous/new-api/commit/dfd3cd8930447780a5733eaaba9a306df595251d) | 系统设置侧栏与 root 权限对齐 | `migrate` |
| [commit-b5b94bc685fd](https://github.com/QuantumNous/new-api/commit/b5b94bc685fd2251551df826dab3575ab262f6dc) | 订阅偏好如实显示 | `merge` |
| [commit-d61d6be7593f](https://github.com/QuantumNous/new-api/commit/d61d6be7593f99eacf9462e99ee13f0c432657f5) | 任务提交 HTTP 成功码兼容 | `migrate` |
| [commit-b7017c251bad](https://github.com/QuantumNous/new-api/commit/b7017c251badaacaab840646a959635d00665e2d) | MySQL系统任务无变更写入保租约 | `migrate` |
| [commit-d1c79d728822](https://github.com/QuantumNous/new-api/commit/d1c79d7288221768a5f7270649d691ac5a137def) | 周统计默认时间范围对齐 | `merge` |

## 暂缓组与必须保留的下游边界

- **任务插件替换与 moejs 系列**：先设计 Ali/Doubao/Volc 的素材 ID、对象暂存、双请求快照、历史任务、不可变成本、计费和通知的迁移，不整包覆盖现有 Go 适配器。
- **协议转换、@模型修饰符和计费身份、Responses WebSocket、Tokenkit**：先证明公开别名、上游名称、收费身份和原始 usage 的一致性，保留 Moonshot/Kimi KVV 与独立 relaykit 模块边界。
- **PAT、动作证明、密码协议与多 RP Passkey**：需要完整路由权限矩阵、滚动升级和旧凭据方案；保留下游手机号、微信、小程序身份和会话失效。
- **预扣信任、fixed/图片计费和定价管理架构**：需要业务确认收费契约与信用策略；保留数量上界、饱和审计、钱包事务、冻结账目和有效调账投影。
- **旧库约束/主键重建、日志导出和系统任务清理**：单独验证 SQLite、MySQL、PostgreSQL 的升级、重复启动、租约、权限和历史证据，不能以编译通过替代。
- **上游品牌、署名、仓库规则和发布工作流**：不直接同步；现有 new-api / QuantumNous 身份信息和下游 CI 治理保留。

## 分析与验证的边界

本轮逐项读取了固定来源的真实差异，并对照下游实现、已有测试、README 台账、UPSTREAM 和计费设计。大型重构只形成架构、依赖和风险评估；尚未完成各个业务分支的兼容性证明。未运行上游候选代码、未执行迁移、未变更数据库或业务源码。每项风险中的回归项目都是确认实施后要补的验证，不是已经通过的测试。

工具和计划校验的 CI 结果只证明本轮材料和流程；它不证明待选功能已经正确迁入。所有建议路径尚未授权。确认后再填写 decision、reason、依赖、evidence 和 ai_scope，提交并给出完整确认 SHA。

## 分领域分析

### rc44 任务插件与计费建议

仅分析与建议，未选择实际功能、未实施业务变更。比较范围为固定上游 rc26→rc44，逐项代码证据使用该项精确 source_sha；下游固定 `b2cb1c2473f1db803ff5bdaafc67699288bde713`。JSON覆盖 task_plugins 51 项及 billing_pricing_and_model_management 38 项，合计89项；每项至少一处实际 Git 对象行证据。全部采用/迁移/升级/covered建议都附下游对照证据。

建议分布：merge 1，migrate 7，upgrade 1，covered 1，skip 2，defer 77。defer大多是同一个尚未采用的任务插件/价格管理架构的后续能力与修复，并不等于认定这些功能没有价值。各项依赖保留，供维护者选择完整功能组。代码行证据及精确建议范围见 `plan.json 中各项 ai_analysis`。

#### 建议优先形成可独立验证的维护批次

| 来源提交 | 建议 | 具体收益 | 实施前后应验证 |
| --- | --- | --- | --- |
| ac381acf4b | migrate | 同日时间区间用&&，避免9–12点成为恒真并全天收费；跨夜保留|| | Go/前端时间语义，时区，非法/同端点范围，组合规则，解析生成往返；已有表达式不静默改写 |
| b5b94bc685 | merge | 如实展示仅订阅偏好，没有订阅时提示拒绝而非自动钱包 | 订阅/钱包四种偏好与有无有效订阅；七种语言；后台收费不变 |
| f064bffa2b | migrate | Ali额外n和prompt_extend在每次尝试前反映到预留，避免适配器之后才改变倍率 | n=0/缺失/超大uint、Extra及multipart绕过、重试跨渠道不串倍率、额度不足补预留、零价转付费、饱和与冻结成本证据；本批不启用fixed/img_cr/image_count语言升级 |
| 2bfb89c1b9 | migrate | 排序菜单不锁定页面滚动，避免布局抖动 | 键盘焦点、外部关闭、移动端；无需为一行样式写实现镜像测试 |
| d61d6be759 | migrate | 合法任务HTTP201/202进入解析，避免误报失败 | 200/201/202、4xx/5xx、空响应、解析错误；预扣/退款/归档保持现有语义；重写为本地Go适配器夹具 |
| 996adffe51 | migrate | 定价行打开编辑时保留文本选择，稳定列定义 | 编辑/删除当前与其他模型、打开关闭、批量调价；防止ref陈旧身份 |
| 0908d9e7b9 | upgrade | 单独升级go-epay 0.0.4→0.0.5并核验回调可用性 | 审阅依赖差异、验签、content-type、金额、订单幂等；此次只做静态建议，未运行新库 |
| 3827fd8307 | migrate | 表达式模型卡片补充遗漏的缓存读取单价 | 普通/特殊表达式、cr缺失/0、图片缓存不误标、沿用下游布局 |
| fc08b76f0f | migrate | 无法解析表达式时仍显示冻结matched_tier，不误报未匹配 | 普通relay和任务两分支、缺失tier的回落、不重算历史账务、饱和及调账信息 |

只有 b5b94bc685 整提交在当前只读 `git apply --check` 中通过；其余即使建议采纳，也需要AI适配已改动的下游代码。该结果仅说明补丁可应用，不等于语义或测试通过。

#### 可认定已覆盖的单项

`b22a43b6e8` 唯一运行时改动是在 `taskBillingOther` 添加 `is_task=true`。上游该提交 `service/task_billing.go:151`，冻结下游同文件`:161` 已写入同值。可建议 covered 并附证据。这个结论只覆盖任务账务日志标记，不代表整个插件任务表达式计费已迁入。

#### 暂缓的大功能组及阻塞

- **JS任务插件（51项）**：起点 eb48396d5f 删除旧 Ali/Doubao/Hailuo 等Go适配器，新增插件管理、协议和产物体系；后续还有moejs、受控fetch、OpenAI Images、共享模型独立价格和流式提交。下游素材ID映射、永久素材副本、临时对象存储、请求快照异步归档、冻结计费/成本、任务通知和原生Volc隔离目前均在Go路径上。未设计这些行为的迁移与旧任务兼容前，不建议引入基础架构，也不孤立导入依赖它的市场UI和版本号。自动时长-1只能作为显式协议能力，按最大秒数预留后真实结算；绝不能作为负数计费乘数。
- **统一模型/供应商与定价管理**：0c76e4dae7及后续引入模型价格快照、乐观版本、元数据清理、币种编辑、转换草稿等。需要先确认与下游财务授权、渠道成本、调账、历史账单证据的权限和事务关系。上游价格同步UI是模型价格数据同步，不是本次CI上游代码版本同步，二者不可混淆。
- **fixed/图片计费语言**：064ed943e1、f362c7c51b及后续扩展引擎、快照和前端解析器。当前计费契约为每百万Token的自包含表达式且历史证据被冻结。应单独批准新收费单位、升级方案和各入口失败/无usage语义。f064bffa2b的数量预留安全部分可先迁入旧引擎，避免为安全修复强制带入语言升级。
- **仅输入预扣策略**：9978ee1e25不仅增加配置，也把预计输出C置0。下游为付费组保留输出估算和持久钱包结算；这属于信用/并发额度策略，不建议在同步批次中默认放宽。需要业务先接受长输出、并发与不足额结算风险，再决定信任门槛和预扣倍数。
- **模型健康采样与费用来源视觉**：健康API由三条最近非空率变为小时序列/近期样本，费用来源经历多次改版。当前下游有紧凑三柱卡片、Badge费用展示、工具费标记和有效调账投影，需选择最终设计，不逐条引入中间态。04c64734cc的24槽样式本地不适用，可skip；上游moejs PGO发版工作流192802ee06也可skip，不覆盖现有CI。

#### 应持续保留的业务边界

Root AGENTS的bounded quota与Checked饱和审计、`relaykit`独立构建、三数据库兼容；README台账中 #1/#8/#11/#14/#17/#18/#19/#20/#21/#22/#23/#26/#31/#34/#35/#36/#37/#39/#43/#46/#47/#48/#55/#56/#57/#59/#60/#61/#72/#74 涉及的任务隔离、对象与素材、永久账务证据、成本权限、有效修正投影和有界资源观测。后续实施先冻结计划/路径与批准提交，再在同一WIP PR逐项适配、更新来源和跑CI；本轮建议不改变VERSION或UPSTREAM。


### Relay/协议/路由差异建议

- 固定上游：`v1.0.0-rc.44` / `633b4c5b190dc1af99a2814391354d86948e7943`。
- 对照下游：合并同步工具后的 main `b2cb1c2473f1db803ff5bdaafc67699288bde713`。评估范围仍始于 `8f6961c675932f406260ff0c218bc2aa0603e9b2`，不改变初始 fork。
- 共 75 个候选：建议 merge 9、migrate 25、defer 37、skip 3、covered 1。建议不等于批准，人工 decision 应全部保持 pending。
- 机器结果：`plan.json 中各项 ai_analysis`，每项均含真实固定 Git blob 行证据，所有采用、迁移和 covered 项都有下游对照；依赖只指向此前来源条目。没有实施、业务代码提交、推送或 Docker 启动。

已读取 README 长期差异台账、UPSTREAM 中 Kimi/Moonshot 来源边界，以及 billingexpr 的单一事实源、token归一化、预扣/结算、表达式版本与快照语义。逐个来源检查真实 diff 和当前相关实现；对 149 文件转换重构等大组只形成架构风险与阻塞评估，没有逐分支完成全部协议/账单 golden 矩阵，因此明确暂缓，不能将该分析当成兼容性验收。

#### 可优先选择的小范围完整采用

| 来源条目 | 功能 | 当前差异与必需验证 |
| --- | --- | --- |
| `commit-0f2a2075ab07` | 请求校验返回400且不重试 | 本地普通校验仍未显式400；保留413和Moonshot独立校验，证明未发送上游请求。 |
| `commit-0bee5d441029` | 阿里图片response_format | 从ImageRequest读取；URL/default不下载，b64_json沿用大小与SSRF限制。 |
| `commit-2cf177ac487e` | RawMessage批量深拷贝 | 验证嵌套字段与重试请求独立，空值与显式零不改变。 |
| `commit-876903a8eb22` | 火山兼容渠道模型列表端点 | 普通VolcEngine改/api/v3/models，保留special base，不触及Volc Native Action API。 |
| `commit-76f7dafd2b82` | 音频扩展名大小写 | 大写/小写同解析，保留时长上界与计费饱和。 |
| `commit-8e5e09166cf8` | Ollama终帧工具调用 | Done帧载荷发送在finish前，工具索引与usage只结束一次。 |
| `commit-3e8c358da09b` | 智谱reasoning_effort | 本地重建DTO漏字段；保留空值、模型映射和原生THINKING。 |
| `commit-13c8e90b8f4b` | Responses→Chat图片URL对象 | 字符串转{url}，已有对象保持；URL/DataURI/detail回归。 |
| `commit-ba0e478e061e` | Claude工具调用伴随文本 | 移除互斥条件，同时保留内容与tool_calls，验证工具结果排序。 |

这些是可优先评审的 merge 建议，尚未执行。用户仍需选择全部或其中一部分。上下文偏移、来源署名限制和回归测试均要在实际实施时重新检查。

#### 值得单独迁移的小修

- `commit-b518d0033b67`：响应头超时，保留头后长流；原提交带受保护README，迁移代码/环境样例而非整提交复制。非流式长生成要给足余量。
- `commit-81336fc69b8a`：拒绝未实现的Gemini `:countTokens`，避免错误当成生成请求及收费。
- `commit-977450a3e810`：内存限流返回标准429说明，可独立适配旧中间件，不等于已经修复reservation并发语义。
- `commit-feefe09f2781`：TLS配置克隆，需和同提交中混入的Waffo store校验拆开；不能把克隆隔离误当允许关闭证书校验。
- `commit-c2b7a9a9e0b5` / `commit-789c970199ea` / `commit-767a82b87647`：原生Claude消息配置、安全配置、thinking扩展和多beta头保真。适配时补Moonshot Messages→普通OpenAI的拒绝/诊断边界，不能伪造支持。
- `commit-5ac75628409d`：Chat→Claude同帧reasoning/text丢答案，适配旧状态机并补块索引/终止/usage矩阵。
- `commit-6e9de44a7dca` / `commit-4eb3b9160566`：工具结果图片提升到user消息。后者只建议迁移媒体修复，其同时携带的Responses流分段须单独评估。

#### 已有 Kimi 能力必须区别处理

- `commit-6e10f9bc927a` 的K3动态工具消息修复有完整现有实现和UPSTREAM记录，可建议 covered：消息级tools、content省略、Token估算、系统提示跳过及原生Moonshot透传。
- `commit-7c044d7c5c2d` 只参考引入了Kimi真实模型名后缀保护；其 @模型修饰符、独立路由身份和计费身份大重构仍未引入，建议 defer，绝不能整项 covered。
- 普通OpenAI上游的 `/moonshot` 兼容入口仍明确拒绝动态工具加载等不可真实复现语义；Kimi KVV透传及缓存/推理明细是需要保留的下游差异。

#### 大组暂缓及阻塞

1. **7137转换/工具/计费基础**：`0ed497f066a6`、`bbd97446c260`，后接推理修饰符、诊断、缓存明细、转换失败计费。必须验证所有协议直连/跨协议/多跳、原始usage与转换usage、明确零/缺失、预扣/结算/退款及下游财务凭证。当前证据不足以批准整包升级。
2. **Responses WebSocket**：`9fe0457ee1f4`起始，需合并评估`75f3d246e2b1`、`ae249f4ecbaf`、`fa3cc1c6b9b1`、`d04c118c8803`、`801456a061fb`。本地无该会话能力。每请求鉴权、多密钥撤销、取消/断连结算、错误关联、关闭帧、设置持久化缺一不可。
3. **请求策略/协议终态/性能与成功限流**：`8b2c7105f91f`、`73a471f3a8b1`、`a5b40663d330`、`6237d9d77eb0`及完成后context取消修复。先定义下游分类监控和现有重试/业务拒绝口径，再证明成功计数和异步终态采样恰好一次。
4. **Tokenkit**：`bb28f2ccad3f`新增独立模块与受保护CI/治理改动，还会改变预扣与缺失usage估算。需要模块化方案、流量校准及Kimi动态工具/音频边界测试，不能作为普通依赖升级。
5. **搜索工具与搜索收费**：`acb347a197ff`、`d10f11ffff20`。唯一query只是供应商账单数的近似；需明确工具客户售价、实报费用、成本快照及估算证据，不能混为同一金额。
6. **每渠道关闭TLS验证**：`c6741c36a9a1`需要具体产品需求和覆盖所有下游HTTP客户端的设计；默认严格、缓存隔离、代理和SSRF边界尚未验收。

#### 需要人工判断的具体差异

- `07578faf26d1`用 `max(url数量,b64数量)`计图片数。若不同对象代表独立的URL图与Base64图，这可能少计；若是同图两个交付表示则去重合理。没有供应商证据时不能替业务作决定。还必须保留本地MaxImageN和AddOtherRatio保护。
- `4b13d9388ed6`为兼容Gemini legacy schema直接删非字符串enum，会放宽输入域；后续`c10780e054e6`提供支持JSON Schema端点的无损保留方案，应优先评估最终能力。legacy降级须有诊断与确认。
- `bc823579294f`删除真实格式转换回归，不建议同步。迁移多模态/schema能力时更需要这些契约测试。
- `0aec08fee811`为致谢空提交，无文件功能差异；实际迁移工具图片修复时记录来源/作者，不生造运行时变更。
- `1d4328e97417`的README声明多项尚未迁入API且涉及受保护署名，应待实际能力确定后人工更新，避免文档先于实现。

#### 后续实施验证建议

先运行所选功能的确定性定向测试，再跑约定的七项CI。后端协议行为需测试普通/流式、原生/跨协议、模型映射/公开别名；计费改动同时覆盖standard/表达式、缓存明细、显式零、失败/取消、预扣与结算、quota饱和审计及客户/成本凭证。relaykit必须 `GOWORK=off go build ./...` 独立验证。渠道设置涉及数据库时须检查SQLite/MySQL/PostgreSQL序列化与默认兼容；前端新增说明/开关需全语言及表单往返测试。

白名单只是建议候选文件，并非审批授权。确认实现前，维护者应按最终选择核对精确路径（包括回归测试和必要locale），提交完整人工计划并给出审批SHA。当前尚未更改VERSION、UPSTREAM来源基线或任何人工decision。


### 身份安全、通用前端、文档构建依赖：AI 建议汇总

评估目标：QuantumNous/new-api `v1.0.0-rc.44` / `633b4c5b190dc1af99a2814391354d86948e7943`。下游运行时对照固定为 `b2cb1c2473f1db803ff5bdaafc67699288bde713`。范围是 inventory 中 identity_security_and_accounts 33 项、frontend_general 42 项、documentation_build_dependencies 36 项，共 111 项。所有内容仅为建议，不授权迁入；原 plan 的 decision 应全部保持 pending。

逐项记录见 `plan.json 中各项 ai_analysis`：34 migrate、11 upgrade、48 defer、18 skip。没有使用 covered：文件重叠、同名实现、功能部分相似不足以证明整项已等价覆盖。每项均保留固定源码路径/行号证据；全部拟采用项有真实下游对照。依赖只引用此前候选 ID。

| 范围 | migrate | upgrade | defer | skip |
| --- | ---: | ---: | ---: | ---: |
| 身份安全/账号 | 9 | 0 | 24 | 0 |
| 通用前端 | 23 | 0 | 18 | 1 |
| 文档/构建/依赖 | 2 | 11 | 6 | 17 |

#### 建议先确认的小范围修复

1. **优先处理当前用户可感知错误和隐私边界**：`692e8d6ee6a9` 内置身份解绑参数（本地传 github_id 等字段名，model.User.ClearBinding 接受 github 等路由值）；`98d50d5383a3` setup状态不跨刷新持久缓存；`06dd76a9d38e` Playground聊天/参数按用户隔离；`2506e1b980ec` CreateUser 标准角色白名单；`dfd3cd893044` 侧栏设置入口与后端 RootAuth 一致。
2. **低范围可靠性交互**：`33320407f4f5` 保存其他OAuth配置时不强制请求既有OIDC discovery；`bdef11750524` 高级自定义路由隐藏header后仍能Add split；`d4c26bfb8dcb` 已有多key渠道可修改选择策略；`b892a6b2e44d` ChunkLoadError只刷新一次；`5fe8917be728` 缓存zhTW不再被转成zhCN；`96275d224bfc` 日期选择器识别zhCN/zhTW；`93ea55f3c310` 秒级日期选择。
3. **可单独评审的显示/性能改善**：长提示换行 `c76452d22412` / `8112120673bf`，Android零blur兼容 `45094bdf41fb`（仍需实机），渠道刷新按钮 `2035a82aeb54`，API地址复制 `585c4610e5d7`，按用户组筛选 `956019109bb0`，小数输入步长 `e04c05c2b929`。图标懒加载 `9c3d3aeb3c30` 必须连同 Windows 分隔符修复 `00e8a00cb4ae`；按本地包结构适配，不能直接照搬价格/用户/额度列的memo重组。
4. **共同组件需成组确认**：抽屉portal容器 `387a40914853`；移动对话框 `c49972441e73` + `a5bf94022060`。后者用 CSS 的100dvh/safe-area变量，不需要凭空添加JS viewport hook。应直接采用包含fallback的最终方案。

这些建议大多为 migrate：当前下游前端虽同为React19/TypeScript/Rsbuild/Base UI，但组件实现与上游后续状态并不完全相同；部分上游新cell/security页/model-provider模块本地不存在。不存在“文件看起来相同所以直接覆盖”的推断。

#### 必须另行设计的身份安全组

- **动作绑定证明、统一安全操作与登录挑战**：`45c3fbe8aeb0`、`0973dc2b8f55`、`6f2333990613`。下游 JWT proof 已绑定用户、session、版本、method、scope；上游进一步绑定动作context并串联enrollment/OAuth/2FA。应保留本地已验证手机号、微信UnionID、miniapp flow purpose、session失效及业务事务原子性。设计之后才能选实现范围。
- **PAT和审计**：`d8cb177440ce`、`a8729b5c3709`、`6cbb1c7edba4`、`caca52f8dea0`、`4924361aef73`。下游还有资产、原生Volc、钱包/财务纠错和小程序路由。先列全路由scope矩阵、旧令牌过渡和撤销策略；API模型key、控制台token、miniapp session必须保持身份边界。管理员step-up需融合已有手机管理证明，不能只替换现有controller。
- **密码加密/存储升级**：`b80d633cf586`、`918427d8ab41`、`0973dc2b8f55`。RSA传输协议、可选开关、Argon2id与v2 envelope必须考虑多副本、滚动升级/回退、长Unicode密码及下游手机号密码登录。不是HTTPS替代品。
- **安全多RP Passkey**：`b6566f33d908`、`385d2dfd10d8` 及后续测试/配置修复。下游当前单RP；需要凭据RPID迁移、旧NULL值、移除预览/确认、Origin验证、域名并发与多DB方案。
- **旧GitHub绑定迁移证据**：`2906e4f779b7` 有明确安全价值，不能因为暂缓整套登录架构而忘记跟进。当前保留 legacy_id，需独立审查自动认领依据；不直接把上游的挑战对象搬进下游手机/微信身份体系。
- **Telegram统一OAuth**：`3e84ec0ab823` 不是对旧Bot Token/assertion配置的透明替换，需要迁移计划与回调验证。

#### 暂缓或不采用的显示/治理项

- API key quota/group/time cells和通用额度弹窗（`2bec370629aa`、`551bb63edf40`、`bd22e45a740a`、`950644c9d544`、`ea7cb0ba4e0f`、`973cf8ef4600`）与日志移动布局（`a5e41a893379`）须保留下游有效账务金额、冻结成本、纠错和管理员字段。`9310231b3c27` 只调整上游新增wrapText模式，而本地没有该属性，单独搬两行无意义。
- 分阶段渠道编辑/快捷选项/悬浮模型映射（`505805a4c3ab`、`ed0ad6e3ff9f`、`0cde9d94f6d3`、`db536ec8bcbb`）涉及本地专有渠道和真实模型/计费身份，先明确最终设计。渠道模型更新预览 `6b638788c645` 可以独立按当前hook适配；它与本次源码版本同步并非同一功能。
- 大型错误提示/翻译契约（`12be9975c0bf` 174文件、`128089fa7ddf` 288文件）需要覆盖下游短信、财务、任务、资产的错误码和历史日志，不在同一批次整包导入。Go现代化214文件 `ebe4c368f287` 建议skip，避免无业务收益的合并冲突。
- 上游README、AGENTS、Issue/PR模板、发布工作流和品牌资源保持skip，按专门治理PR处理。下游已移除上游release.yml，不重建它；当前Dockerfile.dev已使用正确完整ldflags模块路径。`7037ac15bd8a` 只建议修复开发镜像模块清单上下文，本次不启动Docker。
- 价格卡片 `bee45b58a3c0` 采用xl三列，但下游已定制lg三列，建议skip以保留当前布局。`93cf4ba6fb92` 是空差异重复提交，skip并指向实际js-yaml变更 `7f51a97226c4`，不声称下游已经覆盖。

#### 依赖建议与验证门槛

11个upgrade建议包括 Lobe图标、Yace、流式Markdown解析、TanStack路由插件、Gorilla WebSocket、AI SDK和Electron的5类小版本锁定依赖。它们须在下游现有依赖图内生成最小锁文件差异；不能覆盖整个上游锁文件。版本建议来自固定源码比较，未核实具体CVE/发布公告，不宣称任何升级已修复某漏洞或已验证兼容。

Electron39→41、Oxfmt0.57→0.72、混合CI/DOMPurify/KaTeX更新暂缓至专门依赖PR。前者需要跨平台打包与preload/签名/更新验证；formatter需验证版权头保护与幂等；DOMPurify直接依赖和override须一致，KaTeX需覆盖数学渲染和消毒边界。

实施前由维护者确认精确source/test白名单；JSON里的suggested_allowed_paths只是建议，不是最终审批范围。有新UI文本时须按项目i18n技能补七语言。采用现有fixture优先，新增测试只覆盖真实行为，不能复制实现细节。当前分析没有运行上游候选代码，没有启动Docker，也没有做迁移、版本更新、提交或push。

最终验证按选项组合进行：前端typecheck/test/build；root Go test/build；relaykit在GOWORK=off下独立test/build。身份/DB/财务变更还需专属行为与SQLite/MySQL/PostgreSQL验证，七项通用CI不能替代这些契约。React性能建议参考项目vercel-react-best-practices的包导入、localStorage schema及memo依赖原则，未宣称实测包体收益。


### rc44 数据库、日志与运行维护建议

仅建议，不选择实际功能或实施业务代码。覆盖 `database_logs_and_operations` 全部26项；下游固定 `b2cb1c2473f1db803ff5bdaafc67699288bde713`，上游证据分别固定到各项提交，撤销项另附父版本。54处Git对象行证据已检查。建议：migrate15、upgrade1、merge1、covered1、skip2、defer6。所有采用/升级/迁移/covered都附下游对照。

#### 建议维护批次

| 批次 | 来源 | 判断和验证重点 |
| --- | --- | --- |
| 数据库协议/驱动兼容 | 66031a09d9、6eb6f35ed2、74158715cd | PG关闭GORM显式PrepareStmt，与JSON Valuer返回string、Scanner接受string/bytes、GORM1.25.12与mysql1.5.7/pg1.5.9配套。保留#74 observed Dialector、DialFunc、实际字节观测和关闭旁路。三数据库读写/NULL/旧库/重启及PG事务池验证后采用，不能只改开关。 |
| SQLite写锁 | 1751f43ee0 | 本地旧_busy_timeout形式尚未按pure-Go驱动pragma写法修正；迁入WAL、busy_timeout和BEGIN IMMEDIATE，核验钱包与素材事务，不引入长事务竞争。 |
| 历史唯一性与无效DDL | 2b6f1dfefb、27ff6a8767、9a8674425c、007d69942d、043ff99a51 | 以最终定义识别旧Prefill/Token约束和索引，保留外键/复合/表达式/部分索引并事务回滚；归一等价元数据避免无意义重启DDL。绝不照抄删除migrateDBFast、AGENTS及下游成本/账本/存储迁移。SQLite/MySQL5.7+/PG9.6+新库、升级、二次启动、并发和回滚需要真实fixture。 |
| 系统任务MySQL错误锁丢失 | b7017c251b | 同秒相同state导致MySQL changed rows为0，本地UpdateSystemTaskState仍直接报锁丢失。按同一now重新核实租约，真实MySQL覆盖持锁/过期/换持有者；SQLite单测无法证明该修复。 |
| 日志隐私 | 057f71c233 | 优先迁入读取侧user/admin/root范围与RawMessage保整数，不强迫整体替换为上游LogOther写入层。本地仅命中部分键才清理，root_info/channel/reject_reason覆盖不完整；同时必须保留下游request_body/ref/available、财务权限、饱和审计和永久请求归档入口。 |
| 旧授权范围安全 | 043ff99a51 | 结构化加载Casbin字段，对不支持own/other范围显式deny，重建内置角色时保留受限行，避免“跳过限制后回落广泛admin授权”。保留下游财务/素材资源权限。此提交还包含PG约束处理，应分别审查、在同计划中清楚限制范围。 |
| 冷启动与公共内容 | 219c9e0634、36dbbf0f77、c79b74b683 | 公共内容弱ETag采用最终namespace+原内容版本，gzip/Vary/304一致。hint只能优化公共页，受保护路由仍服务器确认无hint的旧有效RefreshCookie；与短信/微信现有会话不冲突。status集中共享React Query，所有直接消费者一起迁移，权限仍以后端为准。 |
| 静态资源预算及压缩 | 8ca26d81df | 真实嵌入文件独立限流与gzip缓存，适配本地router.Use顺序，不能复制上游NoRoute/plugin dispatcher。验证未知/API路径仍限流，压缩q=0/HEAD/Vary、规范缓存键防URL别名无界缓存，保留#74实际传输字节。 |
| Root价格管理 | 34b1942123 | 模型广场关闭后仅pricing、真实Root鉴权通过可访问管理；Authorization存在不代表授权。过期/撤销/禁用账号/非Root/其他模块均须拒绝。 |
| 默认周范围 | d1c79d7288 | 一行默认周粒度30→29，可作为独立merge候选；不会改变正式账期或已存偏好，做适度界面验证即可。 |

整补丁 `git apply --check` 通过的项目有1751f43ee0、b7017c251b、219c9e0634、d1c79d7288、34b1942123。只有补丁可应用的证据，尚未实施/执行运行测试；复杂项仍建议显式迁移以补本地回归和保留下游边界。

#### 已覆盖或被取代

- `8c8c4153d4`：上游只有RPM/TPM独立Scan后赋值的改动。本地`model/log.go:798-807`已使用rateStatistics，保留statistics.Quota，可covered。该结论也由UPSTREAM同步来源记录支持，但判断依据是实际代码，不仅来源文字。
- `69a41eeadc`、`2bf0820f4b`：初稿及撤销成对跳过；采用后续最终安全实现，避免无意义正反应用。
- `9f506dd7f9`：依赖当前不存在的LogOther写入类型，先defer。无变化返回原JSON和敏感键复用原则可融入057f读取侧适配，不需要先搬整个上游写入体系。

#### 需要独立决定的功能

- `4fc9d1f1fa` Options自动主键修复：涉及去重最后值胜出、表交换和旧表备份，不能仅凭Go模型有primaryKey就断言旧库无问题，也不能直接对所有库运行。先取得旧表样本、重复值裁决和恢复路径；新价格管理目前也不存在。
- `90134d5b6b` 系统任务历史：只读筛选有价值，但提交还新增DELETE history。先界定哪些是可清理运行记录，哪些是下游审计/恢复证据，避免覆盖财务、存储任务的长期证据边界。
- `633b4c5b19` 通用日志导出：上游逐行原始logs→NDJSON→前端CSV；本地列表做历史任务价、有效调账和成本权限投影。必须定义与列表一致的导出合同、行级权限、跨节点租约、取消/续租、大整数/CSV公式和历史范围；365天默认搜索限制不得默默阻断永久对账。不能复制原始导出而得到不一致账目。
- `251b76d863` 更新提醒：固定访问QuantumNous/new-api Releases且版本parser不接受-novamaas后缀，需要分清下游发布版本、初始fork基线和上游评估目标，不能表示选择性同步完成了上游全量升级。
- `521cebf585` 首页引导改版：布局/焦点/折叠体验可独立选择，先保留本地下游首页能力。

本轮没有修改计划decision、VERSION、UPSTREAM或任何tracked业务文件，没有启动Docker、服务、数据库或发起推送。


## 逐项机械证据与 AI 建议

### Upstream synchronization assessment: rc44-assessment-20261016

This report contains mechanical evidence, any recorded AI advice and review prompts. Advice does not establish semantic compatibility or authorize implementation.

- Source: https://github.com/QuantumNous/new-api
- Target release: `v1.0.0-rc.44` (`633b4c5b190dc1af99a2814391354d86948e7943`)
- Evaluated range: `8f6961c675932f406260ff0c218bc2aa0603e9b2..633b4c5b190dc1af99a2814391354d86948e7943`
- Preserved initial fork baseline: `8f6961c675932f406260ff0c218bc2aa0603e9b2`
- Downstream base: `b2cb1c2473f1db803ff5bdaafc67699288bde713` (`v1.0.0-rc.26-novamaas.2`)
- Proposed downstream version: `v1.0.0-rc.26-novamaas.3`

Each source commit is a review item, not a proven feature boundary. Use feature/group names and dependencies to describe cross-commit work.
All decisions need a reason. `covered` also needs evidence; patch equivalence and old source references are evidence only.
`merge` and `upgrade` use source commits; approved AI scopes permit conflict adaptation. `migrate` uses a reviewed patch or an explicitly approved AI scope; no commands run from the plan.
Sensitive CI/policy/license/provenance/version paths, symlinks, submodules and changes to existing attribution lines cannot be applied automatically.

#### Historical follow-up outside this evaluated range

Previous skipped/deferred sources and partial migrations stay visible here. These are review evidence, not selected work; a later full adoption or documented coverage supersedes them.

No historical follow-up outside this range is recorded.

#### Current evaluated range

| Item | Source | Decision | Files | Downstream overlap | Review signals |
| --- | --- | --- | ---: | ---: | --- |
| commit-cae3676ec6f4 | [cae3676ec6f4](https://github.com/QuantumNous/new-api/commit/cae3676ec6f46ee5ef596443256f78c4e9b34ceb) | pending | 1 | 0 | review required |
| commit-ba2e9287bb7a | [ba2e9287bb7a](https://github.com/QuantumNous/new-api/commit/ba2e9287bb7a8002116c03daa4c457a330054871) | pending | 1 | 0 | review required |
| commit-e468b73915e5 | [e468b73915e5](https://github.com/QuantumNous/new-api/commit/e468b73915e5028e9849de62c5018a0faa203012) | pending | 10 | 3 | protected files |
| commit-692e8d6ee6a9 | [692e8d6ee6a9](https://github.com/QuantumNous/new-api/commit/692e8d6ee6a9a1620c2d731cb51a1e3154a7042b) | pending | 2 | 1 | review required |
| commit-ac381acf4bf4 | [ac381acf4bf4](https://github.com/QuantumNous/new-api/commit/ac381acf4bf41204b97bb26b4c58c83275877a2e) | pending | 10 | 7 | review required |
| commit-7037ac15bd8a | [7037ac15bd8a](https://github.com/QuantumNous/new-api/commit/7037ac15bd8a29f8ee3e2b74e784bcdb75d67d22) | pending | 1 | 0 | review required |
| commit-eb48396d5fe9 | [eb48396d5fe9](https://github.com/QuantumNous/new-api/commit/eb48396d5fe97d27772d0cd5e3ca8aa5caa4f3e9) | pending | 336 | 77 | protected files |
| commit-0f2a2075ab07 | [0f2a2075ab07](https://github.com/QuantumNous/new-api/commit/0f2a2075ab072ea7e20ffa5dd5d58dbf1b6b5b22) | pending | 1 | 1 | review required |
| commit-98d50d5383a3 | [98d50d5383a3](https://github.com/QuantumNous/new-api/commit/98d50d5383a33432ff6c30b129461b170e5cbffc) | pending | 1 | 0 | review required |
| commit-b80d633cf586 | [b80d633cf586](https://github.com/QuantumNous/new-api/commit/b80d633cf586b001cfbb4200bae93e65abe57c2b) | pending | 10 | 4 | review required |
| commit-8454082f930f | [8454082f930f](https://github.com/QuantumNous/new-api/commit/8454082f930f44593e92791c2581ffc63eb30a59) | pending | 2 | 0 | review required |
| commit-918427d8ab41 | [918427d8ab41](https://github.com/QuantumNous/new-api/commit/918427d8ab41f6adaa4113d0496f1f8621855b70) | pending | 9 | 7 | review required |
| commit-6c22550ea325 | [6c22550ea325](https://github.com/QuantumNous/new-api/commit/6c22550ea325d4fea0e0ece52412cb1e4449291c) | pending | 23 | 3 | review required |
| commit-66031a09d99f | [66031a09d99f](https://github.com/QuantumNous/new-api/commit/66031a09d99f2ac4e0b94e2c41f04ed691a79304) | pending | 5 | 4 | review required |
| commit-0bee5d441029 | [0bee5d441029](https://github.com/QuantumNous/new-api/commit/0bee5d4410296e972bf0076414ade786c2c799c8) | pending | 2 | 0 | review required |
| commit-dc4732cfed71 | [dc4732cfed71](https://github.com/QuantumNous/new-api/commit/dc4732cfed712b004d3d3d94a414d3ff127a93cc) | pending | 11 | 7 | review required |
| commit-b5b94bc685fd | [b5b94bc685fd](https://github.com/QuantumNous/new-api/commit/b5b94bc685fd2251551df826dab3575ab262f6dc) | pending | 8 | 7 | review required |
| commit-1751f43ee07e | [1751f43ee07e](https://github.com/QuantumNous/new-api/commit/1751f43ee07edc9eb0c56fd9b23586861b43df46) | pending | 1 | 0 | review required |
| commit-6eb6f35ed211 | [6eb6f35ed211](https://github.com/QuantumNous/new-api/commit/6eb6f35ed211b7459cae3b9f13286b9c93fc1bd6) | pending | 5 | 3 | review required |
| commit-b518d0033b67 | [b518d0033b67](https://github.com/QuantumNous/new-api/commit/b518d0033b670f5518b8a2f1cf8ea0142a9d1b8d) | pending | 5 | 4 | protected files |
| commit-74158715cde6 | [74158715cde6](https://github.com/QuantumNous/new-api/commit/74158715cde6d7b767ead23d9a2af64b7b58a588) | pending | 2 | 2 | review required |
| commit-69a41eeadc81 | [69a41eeadc81](https://github.com/QuantumNous/new-api/commit/69a41eeadc81adb08d04346512c76d62fd6203db) | pending | 2 | 1 | review required |
| commit-2bf0820f4b89 | [2bf0820f4b89](https://github.com/QuantumNous/new-api/commit/2bf0820f4b89530acf14d389ba3e8229211933fa) | pending | 2 | 1 | review required |
| commit-2b6f1dfefbe2 | [2b6f1dfefbe2](https://github.com/QuantumNous/new-api/commit/2b6f1dfefbe217fed31fc0726717cc7de6958e8e) | pending | 4 | 2 | protected files |
| commit-8c8c4153d4b8 | [8c8c4153d4b8](https://github.com/QuantumNous/new-api/commit/8c8c4153d4b80d54352d21593de41aa9a6178f7e) | pending | 1 | 1 | previous source reference |
| commit-27ff6a8767e7 | [27ff6a8767e7](https://github.com/QuantumNous/new-api/commit/27ff6a8767e728f879d52770c273d4f73214a430) | pending | 3 | 1 | review required |
| commit-67a0585d0f25 | [67a0585d0f25](https://github.com/QuantumNous/new-api/commit/67a0585d0f252dfca445c11b7600971b7eeb8eea) | pending | 6 | 2 | protected files |
| commit-b7017c251bad | [b7017c251bad](https://github.com/QuantumNous/new-api/commit/b7017c251badaacaab840646a959635d00665e2d) | pending | 2 | 1 | review required |
| commit-0ed497f066a6 | [0ed497f066a6](https://github.com/QuantumNous/new-api/commit/0ed497f066a68613375124303ef54f220267b334) | pending | 149 | 22 | protected files |
| commit-bbd97446c260 | [bbd97446c260](https://github.com/QuantumNous/new-api/commit/bbd97446c26092f2e7250af429096064b9e0f899) | pending | 18 | 2 | review required |
| commit-aece11d2f7f0 | [aece11d2f7f0](https://github.com/QuantumNous/new-api/commit/aece11d2f7f095a33052696c5d28d3656609a99e) | pending | 7 | 2 | review required |
| commit-d8ca0ed0bb59 | [d8ca0ed0bb59](https://github.com/QuantumNous/new-api/commit/d8ca0ed0bb596e910e1955461e7691e40d224d70) | pending | 1 | 1 | protected files |
| commit-73afad588ca7 | [73afad588ca7](https://github.com/QuantumNous/new-api/commit/73afad588ca7af07134fa423e8a33fdea6c855b2) | pending | 2 | 0 | review required |
| commit-057f71c2336c | [057f71c2336c](https://github.com/QuantumNous/new-api/commit/057f71c2336c3981187b732a9d06f65490e9a946) | pending | 25 | 13 | review required |
| commit-219c9e06341f | [219c9e06341f](https://github.com/QuantumNous/new-api/commit/219c9e06341f1b100e2c572a5f97c45f151fd280) | pending | 12 | 2 | review required |
| commit-9f506dd7f905 | [9f506dd7f905](https://github.com/QuantumNous/new-api/commit/9f506dd7f905c288b4a119a8197cd64b77eb4a3f) | pending | 6 | 3 | review required |
| commit-9df450fe54e1 | [9df450fe54e1](https://github.com/QuantumNous/new-api/commit/9df450fe54e1a874a5339b7c38a61014217f02c3) | pending | 36 | 9 | review required |
| commit-36dbbf0f77e7 | [36dbbf0f77e7](https://github.com/QuantumNous/new-api/commit/36dbbf0f77e710455e745048f4a32e8120ad3fd2) | pending | 3 | 1 | review required |
| commit-8f5ab8e4048a | [8f5ab8e4048a](https://github.com/QuantumNous/new-api/commit/8f5ab8e4048a90d88b20ae1e6d5228b04233d3b8) | pending | 1 | 1 | protected files |
| commit-32c261923a97 | [32c261923a97](https://github.com/QuantumNous/new-api/commit/32c261923a9786c64d2af087327ef057e7bde7e3) | pending | 6 | 5 | review required |
| commit-3a9f41ee85cc | [3a9f41ee85cc](https://github.com/QuantumNous/new-api/commit/3a9f41ee85cc369f5b8d7fe6e62ff4e7bf3a9ec8) | pending | 2 | 2 | review required |
| commit-7c044d7c5c2d | [7c044d7c5c2d](https://github.com/QuantumNous/new-api/commit/7c044d7c5c2d2beadf16b21910950f8f593bc3ef) | pending | 46 | 18 | previous source reference |
| commit-6b659fd61c50 | [6b659fd61c50](https://github.com/QuantumNous/new-api/commit/6b659fd61c50e35d559c41520a0fff7b8aea56a4) | pending | 11 | 2 | review required |
| commit-d5803532bdcc | [d5803532bdcc](https://github.com/QuantumNous/new-api/commit/d5803532bdccde3a2b1583291f51e92d3519b1c6) | pending | 1 | 1 | protected files |
| commit-eb99ab1b4034 | [eb99ab1b4034](https://github.com/QuantumNous/new-api/commit/eb99ab1b40343c3317bb47981cccdbb2b159a5fa) | pending | 5 | 1 | review required |
| commit-2cf177ac487e | [2cf177ac487e](https://github.com/QuantumNous/new-api/commit/2cf177ac487e62c627c7d423b65735ba2481ef4f) | pending | 2 | 0 | review required |
| commit-49ec46966825 | [49ec46966825](https://github.com/QuantumNous/new-api/commit/49ec4696682530781a036eab1ac195f0b04706c0) | pending | 4 | 3 | review required |
| commit-d8cb177440ce | [d8cb177440ce](https://github.com/QuantumNous/new-api/commit/d8cb177440ceaae422d5bfd96c258d47af4e0f1d) | pending | 129 | 42 | review required |
| commit-9a8674425c5a | [9a8674425c5a](https://github.com/QuantumNous/new-api/commit/9a8674425c5a43435a259b58bb928a55d26be990) | pending | 7 | 4 | protected files |
| commit-45c3fbe8aeb0 | [45c3fbe8aeb0](https://github.com/QuantumNous/new-api/commit/45c3fbe8aeb049f03c13e14298a40b87aea5bd87) | pending | 74 | 24 | protected files |
| commit-3e84ec0ab823 | [3e84ec0ab823](https://github.com/QuantumNous/new-api/commit/3e84ec0ab8239cf2277f8a10d45566630a9c10fa) | pending | 49 | 24 | review required |
| commit-a8729b5c3709 | [a8729b5c3709](https://github.com/QuantumNous/new-api/commit/a8729b5c3709cc01d88fc3f2db5b91347fc9129e) | pending | 12 | 3 | review required |
| commit-0973dc2b8f55 | [0973dc2b8f55](https://github.com/QuantumNous/new-api/commit/0973dc2b8f550de71b75fdd3805576d3ce6ccf42) | pending | 61 | 24 | review required |
| commit-3f8a50cf8877 | [3f8a50cf8877](https://github.com/QuantumNous/new-api/commit/3f8a50cf8877683669cd812240a0beaf7b171c32) | pending | 33 | 18 | protected files |
| commit-521cebf585ef | [521cebf585ef](https://github.com/QuantumNous/new-api/commit/521cebf585efc2e782dd9fb93d0f66752c8d3c32) | pending | 3 | 1 | review required |
| commit-6f2333990613 | [6f2333990613](https://github.com/QuantumNous/new-api/commit/6f2333990613bf3e9dd36f541fc380148c7b5175) | pending | 55 | 24 | review required |
| commit-0c76e4dae77a | [0c76e4dae77a](https://github.com/QuantumNous/new-api/commit/0c76e4dae77a279e015329b7478e6f02d6b62edd) | pending | 122 | 19 | review required |
| commit-3b4652269a6a | [3b4652269a6a](https://github.com/QuantumNous/new-api/commit/3b4652269a6a6d9e2c8650a84ee8c4e447e6473b) | pending | 2 | 0 | review required |
| commit-6e10f9bc927a | [6e10f9bc927a](https://github.com/QuantumNous/new-api/commit/6e10f9bc927a4eae889864a6ef601359d53526b9) | pending | 7 | 6 | previous source reference |
| commit-7bbe85bcb095 | [7bbe85bcb095](https://github.com/QuantumNous/new-api/commit/7bbe85bcb09546e0b89572bf97198fc94889be3d) | pending | 5 | 1 | review required |
| commit-387a40914853 | [387a40914853](https://github.com/QuantumNous/new-api/commit/387a40914853310d69adc2f52474134ced5f4811) | pending | 6 | 0 | review required |
| commit-5c7cca015525 | [5c7cca015525](https://github.com/QuantumNous/new-api/commit/5c7cca015525212a7ac2741da7f3b51fa2e30db0) | pending | 5 | 0 | review required |
| commit-6298b0f32384 | [6298b0f32384](https://github.com/QuantumNous/new-api/commit/6298b0f3238461b9629dfc1c00866f8325123aa1) | pending | 2 | 0 | review required |
| commit-92bc7ff73c5e | [92bc7ff73c5e](https://github.com/QuantumNous/new-api/commit/92bc7ff73c5ef215496d09d3e4b5769c9e35006b) | pending | 3 | 0 | review required |
| commit-210734bb73bc | [210734bb73bc](https://github.com/QuantumNous/new-api/commit/210734bb73bc3c6e37548af90360aa4226566fe2) | pending | 12 | 2 | review required |
| commit-bee45b58a3c0 | [bee45b58a3c0](https://github.com/QuantumNous/new-api/commit/bee45b58a3c0b77e8dc81e6b5aeb4474aa9058d1) | pending | 3 | 0 | review required |
| commit-75e533209490 | [75e533209490](https://github.com/QuantumNous/new-api/commit/75e533209490a8ef3a8b5e3d93e4dac03ba19bcf) | pending | 18 | 7 | review required |
| commit-99974a814f00 | [99974a814f00](https://github.com/QuantumNous/new-api/commit/99974a814f00a26dfa7431beb313e487529bbeed) | pending | 29 | 2 | review required |
| commit-984330920061 | [984330920061](https://github.com/QuantumNous/new-api/commit/984330920061e014159c7581b8fcdb726de2c247) | pending | 43 | 7 | review required |
| commit-eb76b136b85e | [eb76b136b85e](https://github.com/QuantumNous/new-api/commit/eb76b136b85ecba9d6ad19c714c5781e51815530) | pending | 12 | 4 | review required |
| commit-0e0ba152bdcc | [0e0ba152bdcc](https://github.com/QuantumNous/new-api/commit/0e0ba152bdcc6891f6053047ccf14d41b3cad60a) | pending | 29 | 2 | review required |
| commit-71c1fd7caad7 | [71c1fd7caad7](https://github.com/QuantumNous/new-api/commit/71c1fd7caad738db4d13aabbf28eeadb293d0cfe) | pending | 36 | 11 | review required |
| commit-a5e41a893379 | [a5e41a893379](https://github.com/QuantumNous/new-api/commit/a5e41a893379e49bd9e1d025e775c04df8961c95) | pending | 20 | 11 | review required |
| commit-2bec370629aa | [2bec370629aa](https://github.com/QuantumNous/new-api/commit/2bec370629aa72d74d3f598b953a4d33b8bcbf7d) | pending | 28 | 9 | review required |
| commit-8f72ecbbfd86 | [8f72ecbbfd86](https://github.com/QuantumNous/new-api/commit/8f72ecbbfd86ded5ee15373921533eed3d334a9a) | pending | 5 | 1 | review required |
| commit-551bb63edf40 | [551bb63edf40](https://github.com/QuantumNous/new-api/commit/551bb63edf4007d7c4b0930504faf6db87012d73) | pending | 2 | 0 | review required |
| commit-bd22e45a740a | [bd22e45a740a](https://github.com/QuantumNous/new-api/commit/bd22e45a740a7c02c328704dc419d8795237cb74) | pending | 1 | 0 | review required |
| commit-950644c9d544 | [950644c9d544](https://github.com/QuantumNous/new-api/commit/950644c9d54445bdd8796643e4e2e18e13167a4c) | pending | 3 | 0 | review required |
| commit-524455fac3c4 | [524455fac3c4](https://github.com/QuantumNous/new-api/commit/524455fac3c438321df4ae9ed4ffd11bc635cc4f) | pending | 23 | 10 | review required |
| commit-fff0635bb14b | [fff0635bb14b](https://github.com/QuantumNous/new-api/commit/fff0635bb14b8ec5f10df9e582a83591df841902) | pending | 1 | 1 | protected files |
| commit-ebe4c368f287 | [ebe4c368f287](https://github.com/QuantumNous/new-api/commit/ebe4c368f28787919ea478c0abf1ea7101978179) | pending | 214 | 46 | review required |
| commit-ea7cb0ba4e0f | [ea7cb0ba4e0f](https://github.com/QuantumNous/new-api/commit/ea7cb0ba4e0f82e2bfa5e55752eb68bdf902f71b) | pending | 17 | 1 | review required |
| commit-9bf328d97497 | [9bf328d97497](https://github.com/QuantumNous/new-api/commit/9bf328d9749751757d5d6b74088d514813a618bd) | pending | 5 | 0 | protected files |
| commit-4fc9d1f1fa77 | [4fc9d1f1fa77](https://github.com/QuantumNous/new-api/commit/4fc9d1f1fa77c0cfdd9719cb59ff9ecc9885d66a) | pending | 3 | 1 | review required |
| commit-c79b74b68358 | [c79b74b68358](https://github.com/QuantumNous/new-api/commit/c79b74b68358180c68440057596bdc34a99cc649) | pending | 11 | 3 | review required |
| commit-876903a8eb22 | [876903a8eb22](https://github.com/QuantumNous/new-api/commit/876903a8eb22c44e395c03da38f6701c650651ae) | pending | 1 | 0 | review required |
| commit-d52bdc0b4087 | [d52bdc0b4087](https://github.com/QuantumNous/new-api/commit/d52bdc0b4087d50d87eb06387e86b2be53e85cd4) | pending | 48 | 8 | protected files |
| commit-12be9975c0bf | [12be9975c0bf](https://github.com/QuantumNous/new-api/commit/12be9975c0bf01fa175a2bb3360607767c8ba5fb) | pending | 174 | 26 | protected files |
| commit-a20574136b27 | [a20574136b27](https://github.com/QuantumNous/new-api/commit/a20574136b2746e9afc4b268c99cf99b3f6cc68b) | pending | 3 | 0 | review required |
| commit-7cf9b473f61e | [7cf9b473f61e](https://github.com/QuantumNous/new-api/commit/7cf9b473f61eb7dc09f6f0782192f17595140a08) | pending | 2 | 0 | review required |
| commit-064ed943e1ac | [064ed943e1ac](https://github.com/QuantumNous/new-api/commit/064ed943e1ac40e3eaca1b58ffb7fa5dacb3fde3) | pending | 54 | 14 | protected files |
| commit-bdef11750524 | [bdef11750524](https://github.com/QuantumNous/new-api/commit/bdef117505247769268b209665fb3ad7554c3da7) | pending | 1 | 0 | review required |
| commit-b6566f33d908 | [b6566f33d908](https://github.com/QuantumNous/new-api/commit/b6566f33d908fee574c8a6da2a23c4fd5df7353c) | pending | 10 | 8 | review required |
| commit-3e1b8b153249 | [3e1b8b153249](https://github.com/QuantumNous/new-api/commit/3e1b8b15324941ff20c77616737d8dd984582d33) | pending | 16 | 7 | review required |
| commit-3cea2bf799a6 | [3cea2bf799a6](https://github.com/QuantumNous/new-api/commit/3cea2bf799a61e07ad53368538fdb7064b017d77) | pending | 4 | 0 | review required |
| commit-f362c7c51b86 | [f362c7c51b86](https://github.com/QuantumNous/new-api/commit/f362c7c51b86a97e1fc72e4d48280cf53a5d2cb9) | pending | 8 | 0 | review required |
| commit-f064bffa2b65 | [f064bffa2b65](https://github.com/QuantumNous/new-api/commit/f064bffa2b65dd2c6348b375f8b91bab41aabc19) | pending | 25 | 4 | review required |
| commit-f256e40bcb29 | [f256e40bcb29](https://github.com/QuantumNous/new-api/commit/f256e40bcb29292376edb5f0ba2e74ffa463aced) | pending | 3 | 0 | review required |
| commit-25ec832faa48 | [25ec832faa48](https://github.com/QuantumNous/new-api/commit/25ec832faa48031dd74f6fff0f58390e29a79f7c) | pending | 8 | 2 | review required |
| commit-39294418af4c | [39294418af4c](https://github.com/QuantumNous/new-api/commit/39294418af4c920950be1b8c1fed0ed0d2bea52e) | pending | 37 | 12 | protected files |
| commit-b3e279464c82 | [b3e279464c82](https://github.com/QuantumNous/new-api/commit/b3e279464c82d6b1bd9fca6ce27f67c4c9246fec) | pending | 10 | 0 | review required |
| commit-505805a4c3ab | [505805a4c3ab](https://github.com/QuantumNous/new-api/commit/505805a4c3ab55fbcf6227ec110226ff1c193e0b) | pending | 37 | 12 | review required |
| commit-d4c26bfb8dcb | [d4c26bfb8dcb](https://github.com/QuantumNous/new-api/commit/d4c26bfb8dcb7434a0bc1e0d6153e848d86c39b4) | pending | 3 | 2 | review required |
| commit-ab489ab8842b | [ab489ab8842b](https://github.com/QuantumNous/new-api/commit/ab489ab8842b26f3dbc2d7a384e7c0ab0f700e0e) | pending | 19 | 11 | review required |
| commit-251b76d8633e | [251b76d8633e](https://github.com/QuantumNous/new-api/commit/251b76d8633ec4eeb1064dcdb5eb806ecf1a8668) | pending | 19 | 9 | review required |
| commit-74629e29f83f | [74629e29f83f](https://github.com/QuantumNous/new-api/commit/74629e29f83f506bb523c5542aa851947bd423eb) | pending | 96 | 24 | protected files |
| commit-385d2dfd10d8 | [385d2dfd10d8](https://github.com/QuantumNous/new-api/commit/385d2dfd10d821b25c8a6766bd16eea248cb1652) | pending | 48 | 26 | review required |
| commit-129f21b6942e | [129f21b6942e](https://github.com/QuantumNous/new-api/commit/129f21b6942ed433fd7eea1239b4a24d8b58b48b) | pending | 2 | 0 | review required |
| commit-c9a110190c52 | [c9a110190c52](https://github.com/QuantumNous/new-api/commit/c9a110190c5241c24d9d66de340431f5e9873db6) | pending | 10 | 7 | review required |
| commit-be36cbb8facf | [be36cbb8facf](https://github.com/QuantumNous/new-api/commit/be36cbb8facf99d523197f9e18d02b7605331048) | pending | 2 | 0 | review required |
| commit-007d69942d24 | [007d69942d24](https://github.com/QuantumNous/new-api/commit/007d69942d24eeea9edd4665f812d07ed11b3b97) | pending | 2 | 0 | review required |
| commit-2ba61576146f | [2ba61576146f](https://github.com/QuantumNous/new-api/commit/2ba61576146f0583f789ee6845f2280cb86819ce) | pending | 6 | 1 | review required |
| commit-043ff99a51ec | [043ff99a51ec](https://github.com/QuantumNous/new-api/commit/043ff99a51ecad8229389ddd04f45f4b25a23ac6) | pending | 5 | 1 | review required |
| commit-33142f0aee7c | [33142f0aee7c](https://github.com/QuantumNous/new-api/commit/33142f0aee7c87b2110f25dbddcb861dbdc2c44e) | pending | 4 | 1 | review required |
| commit-7fd06381976c | [7fd06381976c](https://github.com/QuantumNous/new-api/commit/7fd06381976cbea126d1f20973548bf60f20e37d) | pending | 3 | 1 | review required |
| commit-815217ba648f | [815217ba648f](https://github.com/QuantumNous/new-api/commit/815217ba648ff800c4f583fbfa36ba286cee6b2d) | pending | 2 | 0 | review required |
| commit-d92612038410 | [d92612038410](https://github.com/QuantumNous/new-api/commit/d926120384109a9e55b0cd54a06bb6e3ae860a2c) | pending | 7 | 0 | review required |
| commit-d1c79d728822 | [d1c79d728822](https://github.com/QuantumNous/new-api/commit/d1c79d7288221768a5f7270649d691ac5a137def) | pending | 1 | 0 | review required |
| commit-2509e25fa060 | [2509e25fa060](https://github.com/QuantumNous/new-api/commit/2509e25fa0601d26f5b8fbb3194ec76689a91a77) | pending | 3 | 0 | review required |
| commit-76f7dafd2b82 | [76f7dafd2b82](https://github.com/QuantumNous/new-api/commit/76f7dafd2b82fc305533ce6a3065b3de2d338bae) | pending | 2 | 0 | review required |
| commit-04c64734cc7c | [04c64734cc7c](https://github.com/QuantumNous/new-api/commit/04c64734cc7c2e58a9efd0b247182330f3f52cce) | pending | 2 | 0 | review required |
| commit-9fe0457ee1f4 | [9fe0457ee1f4](https://github.com/QuantumNous/new-api/commit/9fe0457ee1f4b9de407a254500d54f5a8f41ee29) | pending | 52 | 20 | review required |
| commit-8529f209c859 | [8529f209c859](https://github.com/QuantumNous/new-api/commit/8529f209c85913de9ec98c5af49dcfce4a41361a) | pending | 59 | 29 | review required |
| commit-2bfb89c1b99d | [2bfb89c1b99d](https://github.com/QuantumNous/new-api/commit/2bfb89c1b99d5eb7a799ba5706813614bfb6189a) | pending | 1 | 0 | review required |
| commit-8e5e09166cf8 | [8e5e09166cf8](https://github.com/QuantumNous/new-api/commit/8e5e09166cf83731df1819fd1cb9c50ea2576ed6) | pending | 2 | 0 | review required |
| commit-62f8db775bcd | [62f8db775bcd](https://github.com/QuantumNous/new-api/commit/62f8db775bcdaa5da144d9c4a36f5ead82de3ac5) | pending | 2 | 0 | review required |
| commit-d3874db61fce | [d3874db61fce](https://github.com/QuantumNous/new-api/commit/d3874db61fce61b6c0f5c37523b21912a3f2658d) | pending | 2 | 0 | review required |
| commit-5caafd3d84dc | [5caafd3d84dc](https://github.com/QuantumNous/new-api/commit/5caafd3d84dc74c8b0d081524f57a534b3980cb0) | pending | 16 | 13 | review required |
| commit-fa90b2312cfb | [fa90b2312cfb](https://github.com/QuantumNous/new-api/commit/fa90b2312cfb686aaa6010f9280bffcf0bf6d391) | pending | 6 | 1 | review required |
| commit-81336fc69b8a | [81336fc69b8a](https://github.com/QuantumNous/new-api/commit/81336fc69b8a932e70c24a8dbf439bec94c98328) | pending | 1 | 1 | review required |
| commit-610334dbd760 | [610334dbd760](https://github.com/QuantumNous/new-api/commit/610334dbd760f1c5199a09ae5f860f11451f4c56) | pending | 3 | 1 | review required |
| commit-69a50029819a | [69a50029819a](https://github.com/QuantumNous/new-api/commit/69a50029819a26c53e6babd276d49cfe2f8880ad) | pending | 18 | 13 | review required |
| commit-7209b6db95ef | [7209b6db95ef](https://github.com/QuantumNous/new-api/commit/7209b6db95ef33bfb0f6344ebf61c1fec9af780c) | pending | 5 | 4 | review required |
| commit-57791e31c62b | [57791e31c62b](https://github.com/QuantumNous/new-api/commit/57791e31c62bdf97df855f3b7dc912046b3bc990) | pending | 9 | 2 | protected files |
| commit-ed0ad6e3ff9f | [ed0ad6e3ff9f](https://github.com/QuantumNous/new-api/commit/ed0ad6e3ff9f8785a19d93cba73cf06ef2219668) | pending | 11 | 8 | review required |
| commit-07578faf26d1 | [07578faf26d1](https://github.com/QuantumNous/new-api/commit/07578faf26d1a21c70bf0fbf29f12730a212c583) | pending | 2 | 0 | review required |
| commit-d82678a9708b | [d82678a9708b](https://github.com/QuantumNous/new-api/commit/d82678a9708b611d52386c01f04238ec01f07463) | pending | 3 | 1 | protected files |
| commit-0cde9d94f6d3 | [0cde9d94f6d3](https://github.com/QuantumNous/new-api/commit/0cde9d94f6d3b77d5258ca6b36959f895d26f2a4) | pending | 31 | 8 | review required |
| commit-8b2c7105f91f | [8b2c7105f91f](https://github.com/QuantumNous/new-api/commit/8b2c7105f91f2316f90eb86f9b1c6b98774c7105) | pending | 1 | 0 | review required |
| commit-73a471f3a8b1 | [73a471f3a8b1](https://github.com/QuantumNous/new-api/commit/73a471f3a8b1cdf5dfe371184dfc3579cdcbefcc) | pending | 71 | 23 | review required |
| commit-8a11b4305ed1 | [8a11b4305ed1](https://github.com/QuantumNous/new-api/commit/8a11b4305ed139b55ee91f914c6a19d59c783d72) | pending | 25 | 1 | review required |
| commit-a5b40663d330 | [a5b40663d330](https://github.com/QuantumNous/new-api/commit/a5b40663d3307fc5011c85b83eb18bbaa33dbad2) | pending | 40 | 13 | review required |
| commit-6237d9d77eb0 | [6237d9d77eb0](https://github.com/QuantumNous/new-api/commit/6237d9d77eb0544329b88ff96d71c1715af4268f) | pending | 3 | 0 | review required |
| commit-75f3d246e2b1 | [75f3d246e2b1](https://github.com/QuantumNous/new-api/commit/75f3d246e2b1d6a3ea56342d9bc82cb3e1d0b6e2) | pending | 3 | 0 | review required |
| commit-90134d5b6bff | [90134d5b6bff](https://github.com/QuantumNous/new-api/commit/90134d5b6bff1a66affa35c5bd8e90e3d74a93f8) | pending | 21 | 15 | review required |
| commit-a8ed7f7c5f24 | [a8ed7f7c5f24](https://github.com/QuantumNous/new-api/commit/a8ed7f7c5f24e56c2f1612126d82e9fec625056c) | pending | 2 | 0 | review required |
| commit-e2471403c779 | [e2471403c779](https://github.com/QuantumNous/new-api/commit/e2471403c77957818b939ad86bbeccd051d2d042) | pending | 2 | 0 | review required |
| commit-42cbe6206a62 | [42cbe6206a62](https://github.com/QuantumNous/new-api/commit/42cbe6206a62550896c8371d62a6f15bad5b3e05) | pending | 16 | 8 | protected files |
| commit-d8069436612e | [d8069436612e](https://github.com/QuantumNous/new-api/commit/d8069436612e08bd0d96169222b10031556d09a7) | pending | 9 | 7 | review required |
| commit-ae249f4ecbaf | [ae249f4ecbaf](https://github.com/QuantumNous/new-api/commit/ae249f4ecbaf3ae9b1fc62cf377ca5a764b9a002) | pending | 11 | 3 | review required |
| commit-fa3cc1c6b9b1 | [fa3cc1c6b9b1](https://github.com/QuantumNous/new-api/commit/fa3cc1c6b9b13f7bf25c3e896d439185d326ac9c) | pending | 14 | 8 | review required |
| commit-32ef6d216f9f | [32ef6d216f9f](https://github.com/QuantumNous/new-api/commit/32ef6d216f9fa0a2aa182bd8aad9a02daf3767c8) | pending | 14 | 8 | review required |
| commit-9e905a83e1d6 | [9e905a83e1d6](https://github.com/QuantumNous/new-api/commit/9e905a83e1d6c369b3aa43de4f04ab244cce3205) | pending | 5 | 2 | review required |
| commit-ed7c4e35d429 | [ed7c4e35d429](https://github.com/QuantumNous/new-api/commit/ed7c4e35d4297962991ac5f00eea504ceac0fbd0) | pending | 32 | 14 | review required |
| commit-e2bb05f70ed6 | [e2bb05f70ed6](https://github.com/QuantumNous/new-api/commit/e2bb05f70ed6b6725a83ff6f1c97dbd2c2f640d4) | pending | 2 | 0 | review required |
| commit-55a6cd2a44ea | [55a6cd2a44ea](https://github.com/QuantumNous/new-api/commit/55a6cd2a44eaf189bbd565dd2cdfa591d19b06d6) | pending | 2 | 0 | review required |
| commit-3524fe0b1579 | [3524fe0b1579](https://github.com/QuantumNous/new-api/commit/3524fe0b15794d8d19378827d36a7edc0b0e91ea) | pending | 4 | 0 | protected files |
| commit-2906e4f779b7 | [2906e4f779b7](https://github.com/QuantumNous/new-api/commit/2906e4f779b715f282ae11203211dca77051d5af) | pending | 18 | 7 | review required |
| commit-3e8c358da09b | [3e8c358da09b](https://github.com/QuantumNous/new-api/commit/3e8c358da09b4598a946712db8219ff44f197475) | pending | 1 | 0 | review required |
| commit-946c730ed008 | [946c730ed008](https://github.com/QuantumNous/new-api/commit/946c730ed00806650fef509c74a868dda2847f7c) | pending | 3 | 0 | review required |
| commit-ea336ed6654a | [ea336ed6654a](https://github.com/QuantumNous/new-api/commit/ea336ed6654a97c0c27e004581578231e0d92824) | pending | 7 | 2 | review required |
| commit-96b0da2273a1 | [96b0da2273a1](https://github.com/QuantumNous/new-api/commit/96b0da2273a187b3fd9621f77b10dcebd046e2f6) | pending | 6 | 1 | review required |
| commit-1ea24a050d5e | [1ea24a050d5e](https://github.com/QuantumNous/new-api/commit/1ea24a050d5e276584d744122810b8dc57b10f0e) | pending | 6 | 2 | protected files |
| commit-e04c05c2b929 | [e04c05c2b929](https://github.com/QuantumNous/new-api/commit/e04c05c2b929c0df6cc35227f860cf6e19d480c7) | pending | 8 | 1 | review required |
| commit-b0bf2580e249 | [b0bf2580e249](https://github.com/QuantumNous/new-api/commit/b0bf2580e2491869ee9fe861c12b86153c3fe5f4) | pending | 6 | 2 | protected files |
| commit-972aed197282 | [972aed197282](https://github.com/QuantumNous/new-api/commit/972aed1972820389ea0b603ca58f03f846fbf790) | pending | 11 | 1 | review required |
| commit-a89037d0b9b0 | [a89037d0b9b0](https://github.com/QuantumNous/new-api/commit/a89037d0b9b065da8582959018321abbc48db549) | pending | 3 | 1 | review required |
| commit-ef70d1b79998 | [ef70d1b79998](https://github.com/QuantumNous/new-api/commit/ef70d1b79998db9bb32b6b79205afbc4b9f8716b) | pending | 1 | 1 | review required |
| commit-b932e6d876cd | [b932e6d876cd](https://github.com/QuantumNous/new-api/commit/b932e6d876cda4fa49528b5cdb873d9fc5e31b6f) | pending | 21 | 8 | review required |
| commit-9c3d3aeb3c30 | [9c3d3aeb3c30](https://github.com/QuantumNous/new-api/commit/9c3d3aeb3c30ea09c7ad0f810c86205c4da558d7) | pending | 16 | 3 | review required |
| commit-33320407f4f5 | [33320407f4f5](https://github.com/QuantumNous/new-api/commit/33320407f4f5dcf5bc635ee6d73dcc31c8bcc031) | pending | 8 | 6 | review required |
| commit-4c34f25a49f1 | [4c34f25a49f1](https://github.com/QuantumNous/new-api/commit/4c34f25a49f1cf491479e3842b2251e8a6908851) | pending | 48 | 19 | review required |
| commit-2d7aef7414e8 | [2d7aef7414e8](https://github.com/QuantumNous/new-api/commit/2d7aef7414e8671b1ac7db336bc170eeb7cc5e2b) | pending | 2 | 0 | review required |
| commit-6b638788c645 | [6b638788c645](https://github.com/QuantumNous/new-api/commit/6b638788c645a6c9164c5c6394827e3f5e3a1abd) | pending | 12 | 7 | review required |
| commit-db536ec8bcbb | [db536ec8bcbb](https://github.com/QuantumNous/new-api/commit/db536ec8bcbbc63034d009a7257ecee0f3468abf) | pending | 17 | 8 | review required |
| commit-58a9eff4c842 | [58a9eff4c842](https://github.com/QuantumNous/new-api/commit/58a9eff4c8421a6bbb04855e561310d3eeabd107) | pending | 8 | 4 | review required |
| commit-3fdf9083dc6b | [3fdf9083dc6b](https://github.com/QuantumNous/new-api/commit/3fdf9083dc6bfc47fc2ad0de3fac327d749d778d) | pending | 7 | 5 | review required |
| commit-da2540dda132 | [da2540dda132](https://github.com/QuantumNous/new-api/commit/da2540dda1328040193f279a58c3e315083e416c) | pending | 1 | 0 | review required |
| commit-d904b9700015 | [d904b9700015](https://github.com/QuantumNous/new-api/commit/d904b9700015ad73285b0f0030c88eb6683fbeda) | pending | 13 | 7 | review required |
| commit-585c4610e5d7 | [585c4610e5d7](https://github.com/QuantumNous/new-api/commit/585c4610e5d7103480620ab61954db5a5feac35b) | pending | 9 | 7 | review required |
| commit-c49972441e73 | [c49972441e73](https://github.com/QuantumNous/new-api/commit/c49972441e73f05207427567a2d9ad7c620db4bd) | pending | 23 | 3 | review required |
| commit-03563a4a7916 | [03563a4a7916](https://github.com/QuantumNous/new-api/commit/03563a4a791629a811076f1cbcbecfbe2716eac2) | pending | 80 | 22 | protected files |
| commit-a5bf94022060 | [a5bf94022060](https://github.com/QuantumNous/new-api/commit/a5bf940220604e5b9d150e2338bf6528dfbb8be4) | pending | 9 | 0 | review required |
| commit-2fddb863f221 | [2fddb863f221](https://github.com/QuantumNous/new-api/commit/2fddb863f2219d9eacea5b899ee39b46f11adf33) | pending | 9 | 7 | review required |
| commit-3abbb8198e04 | [3abbb8198e04](https://github.com/QuantumNous/new-api/commit/3abbb8198e0468ac6e16eb0103c1d3642e2e7668) | pending | 3 | 1 | review required |
| commit-65d3a2171fd1 | [65d3a2171fd1](https://github.com/QuantumNous/new-api/commit/65d3a2171fd11c65ad2e2e1f1d65723926659c81) | pending | 4 | 0 | review required |
| commit-9978ee1e25a6 | [9978ee1e25a6](https://github.com/QuantumNous/new-api/commit/9978ee1e25a647bfe004e96c8719a2cb62c24732) | pending | 23 | 12 | review required |
| commit-47713bcb1749 | [47713bcb1749](https://github.com/QuantumNous/new-api/commit/47713bcb1749f81c3a48127c0150e4c6d4da35f1) | pending | 4 | 0 | review required |
| commit-9a0be8750a6d | [9a0be8750a6d](https://github.com/QuantumNous/new-api/commit/9a0be8750a6d736d9692535ed2cd68f8eec46529) | pending | 7 | 0 | protected files |
| commit-9c293e8c0237 | [9c293e8c0237](https://github.com/QuantumNous/new-api/commit/9c293e8c02371bda844af79e3500ff2d516d1dda) | pending | 1 | 0 | review required |
| commit-474ed66fb14b | [474ed66fb14b](https://github.com/QuantumNous/new-api/commit/474ed66fb14b0de30f82a1729b4f7cf3dab19d63) | pending | 2 | 0 | review required |
| commit-00e8a00cb4ae | [00e8a00cb4ae](https://github.com/QuantumNous/new-api/commit/00e8a00cb4aebb4e029b5ff32b8fe1e98f0b0bc5) | pending | 2 | 0 | review required |
| commit-e537dc380bb2 | [e537dc380bb2](https://github.com/QuantumNous/new-api/commit/e537dc380bb253d4f902ac033ab39dd2ca462b5e) | pending | 9 | 7 | review required |
| commit-c0cff23a3af8 | [c0cff23a3af8](https://github.com/QuantumNous/new-api/commit/c0cff23a3af8e68809647bd39ec204bedbea28af) | pending | 17 | 7 | review required |
| commit-2c175190c44a | [2c175190c44a](https://github.com/QuantumNous/new-api/commit/2c175190c44a95c6af4877e901dc5211ef706979) | pending | 3 | 0 | review required |
| commit-d61d6be7593f | [d61d6be7593f](https://github.com/QuantumNous/new-api/commit/d61d6be7593f99eacf9462e99ee13f0c432657f5) | pending | 3 | 1 | review required |
| commit-54eee488bed3 | [54eee488bed3](https://github.com/QuantumNous/new-api/commit/54eee488bed3ec6b9659b3a792194b18c2dfad27) | pending | 6 | 1 | review required |
| commit-6e9de44a7dca | [6e9de44a7dca](https://github.com/QuantumNous/new-api/commit/6e9de44a7dca08f44234ab9c623f6a93a174d0c4) | pending | 2 | 0 | review required |
| commit-4eb3b9160566 | [4eb3b9160566](https://github.com/QuantumNous/new-api/commit/4eb3b9160566dc4c1f4f0e9340de254f7febf056) | pending | 5 | 0 | review required |
| commit-b6809a52df88 | [b6809a52df88](https://github.com/QuantumNous/new-api/commit/b6809a52df88e8db5c8591b2d3788a05401271b7) | pending | 3 | 0 | review required |
| commit-0aec08fee811 | [0aec08fee811](https://github.com/QuantumNous/new-api/commit/0aec08fee811ec6136828fda790551b49e410301) | pending | 0 | 0 | review required |
| commit-9310231b3c27 | [9310231b3c27](https://github.com/QuantumNous/new-api/commit/9310231b3c27fea933e939b46cf26e0ce67192e3) | pending | 2 | 0 | review required |
| commit-8cb88ebd17a0 | [8cb88ebd17a0](https://github.com/QuantumNous/new-api/commit/8cb88ebd17a072691f39705ae72c786f8bb5c4e0) | pending | 2 | 0 | review required |
| commit-996adffe5165 | [996adffe5165](https://github.com/QuantumNous/new-api/commit/996adffe5165bd5e311e33a03a86b8aede1fe376) | pending | 2 | 0 | review required |
| commit-5401874c6387 | [5401874c6387](https://github.com/QuantumNous/new-api/commit/5401874c6387b52149b7e5b0f61d9dafea5f5b7a) | pending | 2 | 2 | protected files |
| commit-a46f045d6d14 | [a46f045d6d14](https://github.com/QuantumNous/new-api/commit/a46f045d6d146b1e7e17616974bbf37e32885850) | pending | 1 | 0 | review required |
| commit-6c14c0762dc7 | [6c14c0762dc7](https://github.com/QuantumNous/new-api/commit/6c14c0762dc7f3bc77d0c80585f974e65ac2a571) | pending | 1 | 0 | review required |
| commit-c76452d22412 | [c76452d22412](https://github.com/QuantumNous/new-api/commit/c76452d22412be45eb708f2618748f7aac46703d) | pending | 1 | 0 | review required |
| commit-d04c118c8803 | [d04c118c8803](https://github.com/QuantumNous/new-api/commit/d04c118c8803f49e0c9bab74dcf5b5efeab9464a) | pending | 2 | 1 | review required |
| commit-c2b7a9a9e0b5 | [c2b7a9a9e0b5](https://github.com/QuantumNous/new-api/commit/c2b7a9a9e0b548c2051a949fceabb59029adcb49) | pending | 2 | 0 | review required |
| commit-789c970199ea | [789c970199ea](https://github.com/QuantumNous/new-api/commit/789c970199ea527e6a26e071915f4a4cd2c64178) | pending | 1 | 0 | review required |
| commit-f5e54ccdb203 | [f5e54ccdb203](https://github.com/QuantumNous/new-api/commit/f5e54ccdb203e8d4469035e72cc48dfd25b4314d) | pending | 7 | 0 | review required |
| commit-87bb71e7fc83 | [87bb71e7fc83](https://github.com/QuantumNous/new-api/commit/87bb71e7fc83121fc995a048077e5c0e209b789f) | pending | 2 | 0 | review required |
| commit-3509ae7c194d | [3509ae7c194d](https://github.com/QuantumNous/new-api/commit/3509ae7c194dcc0011cc4ec20cd495b198a7ed6d) | pending | 6 | 0 | review required |
| commit-6cbb1c7edba4 | [6cbb1c7edba4](https://github.com/QuantumNous/new-api/commit/6cbb1c7edba4d634ba637a6e6f111115d6de9c9a) | pending | 29 | 16 | review required |
| commit-caca52f8dea0 | [caca52f8dea0](https://github.com/QuantumNous/new-api/commit/caca52f8dea0436d63bacc0dd3e0a73ba7b6b094) | pending | 79 | 32 | review required |
| commit-335ebcff71b8 | [335ebcff71b8](https://github.com/QuantumNous/new-api/commit/335ebcff71b843655bf2c64f7a9a33ca21b792ea) | pending | 16 | 1 | review required |
| commit-2ffc59590115 | [2ffc59590115](https://github.com/QuantumNous/new-api/commit/2ffc59590115bb99d5252ceaeb3e73f06dfcd1ae) | pending | 24 | 4 | protected files |
| commit-c71eefbcd194 | [c71eefbcd194](https://github.com/QuantumNous/new-api/commit/c71eefbcd194ea3caca8476e399ae580b275470c) | pending | 14 | 2 | protected files |
| commit-2506e1b980ec | [2506e1b980ec](https://github.com/QuantumNous/new-api/commit/2506e1b980ecf67c829ef20255a16fc61c47cb55) | pending | 2 | 1 | review required |
| commit-4924361aef73 | [4924361aef73](https://github.com/QuantumNous/new-api/commit/4924361aef733ed724cfdb77ccf35690d9629971) | pending | 34 | 13 | review required |
| commit-ae73ef8e25b9 | [ae73ef8e25b9](https://github.com/QuantumNous/new-api/commit/ae73ef8e25b957ea3291512b66664f418f33c180) | pending | 2 | 0 | review required |
| commit-2035a82aeb54 | [2035a82aeb54](https://github.com/QuantumNous/new-api/commit/2035a82aeb5414253a728bd937d4b8f97aa99b9b) | pending | 2 | 1 | review required |
| commit-feefe09f2781 | [feefe09f2781](https://github.com/QuantumNous/new-api/commit/feefe09f2781429e039c3f7ba88da7f29492e98c) | pending | 6 | 2 | review required |
| commit-d0cb7347c07b | [d0cb7347c07b](https://github.com/QuantumNous/new-api/commit/d0cb7347c07bc935187c077b960435f3f3b08595) | pending | 2 | 1 | review required |
| commit-8112120673bf | [8112120673bf](https://github.com/QuantumNous/new-api/commit/8112120673bf593eb62e3a297dafca32a4605eb6) | pending | 2 | 0 | review required |
| commit-dfd3cd893044 | [dfd3cd893044](https://github.com/QuantumNous/new-api/commit/dfd3cd8930447780a5733eaaba9a306df595251d) | pending | 1 | 1 | review required |
| commit-5fe8917be728 | [5fe8917be728](https://github.com/QuantumNous/new-api/commit/5fe8917be728a93a29072a9fbd65429ed30d2930) | pending | 2 | 0 | review required |
| commit-56758edf95ec | [56758edf95ec](https://github.com/QuantumNous/new-api/commit/56758edf95ec162033a6b73b553ac789404f87c9) | pending | 22 | 9 | review required |
| commit-1a4166d8e8ba | [1a4166d8e8ba](https://github.com/QuantumNous/new-api/commit/1a4166d8e8ba9802d2ca56fe8ecf0ed5404e80d5) | pending | 13 | 0 | review required |
| commit-b48b74ab77cf | [b48b74ab77cf](https://github.com/QuantumNous/new-api/commit/b48b74ab77cf289ccaaa8d99c067ce50cced4a82) | pending | 6 | 3 | protected files |
| commit-973cf8ef4600 | [973cf8ef4600](https://github.com/QuantumNous/new-api/commit/973cf8ef4600947a4270e95ada7916740fa8264c) | pending | 1 | 0 | review required |
| commit-78bd5b1cbf7b | [78bd5b1cbf7b](https://github.com/QuantumNous/new-api/commit/78bd5b1cbf7bffc462a515a6c6e27567ace7a4d8) | pending | 10 | 0 | review required |
| commit-8ca26d81dfff | [8ca26d81dfff](https://github.com/QuantumNous/new-api/commit/8ca26d81dfff9d0454e0f5d06394c93689fe69b1) | pending | 7 | 2 | review required |
| commit-956019109bb0 | [956019109bb0](https://github.com/QuantumNous/new-api/commit/956019109bb09eb94417fd69ec0a5af0d897ac5a) | pending | 4 | 1 | review required |
| commit-c6bac617d46a | [c6bac617d46a](https://github.com/QuantumNous/new-api/commit/c6bac617d46ad2d63fa7cb692942fe6d21e264f7) | pending | 13 | 7 | review required |
| commit-0908d9e7b911 | [0908d9e7b911](https://github.com/QuantumNous/new-api/commit/0908d9e7b91124e5808f20d2d4b9402c7f2f5fae) | pending | 3 | 2 | review required |
| commit-bb28f2ccad3f | [bb28f2ccad3f](https://github.com/QuantumNous/new-api/commit/bb28f2ccad3f6286a9dc4a0ad9e20f571f57533b) | pending | 35 | 11 | protected files |
| commit-0a532f7f7d1b | [0a532f7f7d1b](https://github.com/QuantumNous/new-api/commit/0a532f7f7d1b507cc6657d6bc140f73d4298e46a) | pending | 3 | 0 | review required |
| commit-0a03088e0e73 | [0a03088e0e73](https://github.com/QuantumNous/new-api/commit/0a03088e0e737e99bc7ca936d3d2276e8d636fb8) | pending | 22 | 1 | review required |
| commit-acb347a197ff | [acb347a197ff](https://github.com/QuantumNous/new-api/commit/acb347a197ff0e22f651be9813b4a1a867342941) | pending | 52 | 14 | review required |
| commit-45094bdf41fb | [45094bdf41fb](https://github.com/QuantumNous/new-api/commit/45094bdf41fb13a5d24bd3e8eb3fdfec6241c59f) | pending | 1 | 0 | review required |
| commit-04c597bf6cce | [04c597bf6cce](https://github.com/QuantumNous/new-api/commit/04c597bf6cce1eea76b9e67a4936e15fb54a5612) | pending | 4 | 0 | review required |
| commit-d10f11ffff20 | [d10f11ffff20](https://github.com/QuantumNous/new-api/commit/d10f11ffff20688b8e3e745268cc0036e896d068) | pending | 11 | 0 | review required |
| commit-0cce21db0283 | [0cce21db0283](https://github.com/QuantumNous/new-api/commit/0cce21db0283304d139bf22588f4dc31bdc5ae67) | pending | 2 | 0 | review required |
| commit-bc823579294f | [bc823579294f](https://github.com/QuantumNous/new-api/commit/bc823579294fbca0064b622c22477d3d04ca1bd8) | pending | 4 | 0 | review required |
| commit-b23471112d8f | [b23471112d8f](https://github.com/QuantumNous/new-api/commit/b23471112d8f7e5848f70f30b798ee417d1ec75c) | pending | 2 | 2 | review required |
| commit-192802ee0691 | [192802ee0691](https://github.com/QuantumNous/new-api/commit/192802ee0691900575a3c7a0cdb90b2d7e703035) | pending | 4 | 3 | protected files |
| commit-6370b2942416 | [6370b2942416](https://github.com/QuantumNous/new-api/commit/6370b29424168039e94d40d610191e7d2e65dbf4) | pending | 11 | 0 | review required |
| commit-7aa3531ef4c2 | [7aa3531ef4c2](https://github.com/QuantumNous/new-api/commit/7aa3531ef4c247ad4891c06cc8ed9d0ffb73fecc) | pending | 2 | 0 | review required |
| commit-e0483122c7d3 | [e0483122c7d3](https://github.com/QuantumNous/new-api/commit/e0483122c7d3693629f0e84473380a8c9506c56a) | pending | 3 | 0 | review required |
| commit-128089fa7ddf | [128089fa7ddf](https://github.com/QuantumNous/new-api/commit/128089fa7ddf41b575206a0feb28940d33793833) | pending | 288 | 77 | protected files |
| commit-1d4328e97417 | [1d4328e97417](https://github.com/QuantumNous/new-api/commit/1d4328e97417a043a161a0dd30a5b129be3ace49) | pending | 1 | 0 | protected files |
| commit-b66674ac23d7 | [b66674ac23d7](https://github.com/QuantumNous/new-api/commit/b66674ac23d79bfd35a90797ca1ac64fc74a2d8f) | pending | 12 | 0 | protected files |
| commit-29509776ef85 | [29509776ef85](https://github.com/QuantumNous/new-api/commit/29509776ef855965fb2cff3978578d891005945e) | pending | 18 | 2 | protected files |
| commit-66c3e142820e | [66c3e142820e](https://github.com/QuantumNous/new-api/commit/66c3e142820e513a328def88f2771f73479275bc) | pending | 3 | 0 | review required |
| commit-3f02244783e0 | [3f02244783e0](https://github.com/QuantumNous/new-api/commit/3f02244783e0bb96b1398601f156625936be2f15) | pending | 4 | 0 | review required |
| commit-ba3b2c015ad8 | [ba3b2c015ad8](https://github.com/QuantumNous/new-api/commit/ba3b2c015ad83e6710c18592b0e3e3ba9df9a67e) | pending | 7 | 0 | review required |
| commit-c8d234baa18b | [c8d234baa18b](https://github.com/QuantumNous/new-api/commit/c8d234baa18b9090b394fc1b8f384eec94f8bfef) | pending | 2 | 0 | review required |
| commit-c6741c36a9a1 | [c6741c36a9a1](https://github.com/QuantumNous/new-api/commit/c6741c36a9a1a729554ad94ee99f21602a214a23) | pending | 29 | 18 | review required |
| commit-1ca79d57e77d | [1ca79d57e77d](https://github.com/QuantumNous/new-api/commit/1ca79d57e77d91c855371b43bb26b151429e18a1) | pending | 10 | 0 | review required |
| commit-0f38a14a5e8e | [0f38a14a5e8e](https://github.com/QuantumNous/new-api/commit/0f38a14a5e8e322ea741478b0e49878be3130e1d) | pending | 7 | 2 | review required |
| commit-036ba376da7d | [036ba376da7d](https://github.com/QuantumNous/new-api/commit/036ba376da7d572a4db3e0be8553aadc28c3616a) | pending | 4 | 0 | review required |
| commit-13c8e90b8f4b | [13c8e90b8f4b](https://github.com/QuantumNous/new-api/commit/13c8e90b8f4b929aa88c2e9616efc462e0068d72) | pending | 2 | 0 | review required |
| commit-5ac75628409d | [5ac75628409d](https://github.com/QuantumNous/new-api/commit/5ac75628409de7e3adde7d36771700714af50853) | pending | 1 | 0 | review required |
| commit-977450a3e810 | [977450a3e810](https://github.com/QuantumNous/new-api/commit/977450a3e810ee9c7d7e98acce1dc4efad90b491) | pending | 4 | 0 | review required |
| commit-b892a6b2e44d | [b892a6b2e44d](https://github.com/QuantumNous/new-api/commit/b892a6b2e44df572698f6cf731ff381f1977a1f2) | pending | 2 | 0 | review required |
| commit-1fb6d07fee01 | [1fb6d07fee01](https://github.com/QuantumNous/new-api/commit/1fb6d07fee0171d5f4a0bd2d2ea2660572b08389) | pending | 8 | 1 | review required |
| commit-06dd76a9d38e | [06dd76a9d38e](https://github.com/QuantumNous/new-api/commit/06dd76a9d38eb0eaf8e900b1264b500250156c4d) | pending | 2 | 0 | review required |
| commit-96275d224bfc | [96275d224bfc](https://github.com/QuantumNous/new-api/commit/96275d224bfcf6b5545ca7e57b3bf90f7c3ad396) | pending | 7 | 1 | review required |
| commit-93ea55f3c310 | [93ea55f3c310](https://github.com/QuantumNous/new-api/commit/93ea55f3c3102f7bc11cd005570c7965f0f22a03) | pending | 4 | 1 | review required |
| commit-4b13d9388ed6 | [4b13d9388ed6](https://github.com/QuantumNous/new-api/commit/4b13d9388ed6f43118707534daad4903e5d22745) | pending | 2 | 0 | review required |
| commit-801456a061fb | [801456a061fb](https://github.com/QuantumNous/new-api/commit/801456a061fb9988957341c4eb5aa8ec20c40de1) | pending | 3 | 0 | review required |
| commit-c10780e054e6 | [c10780e054e6](https://github.com/QuantumNous/new-api/commit/c10780e054e6e518c9d0d8a903526663c9b57d75) | pending | 10 | 1 | review required |
| commit-42e694eb60fe | [42e694eb60fe](https://github.com/QuantumNous/new-api/commit/42e694eb60fe3191795c00d416571ced716eb9c4) | pending | 2 | 0 | review required |
| commit-3827fd83076b | [3827fd83076b](https://github.com/QuantumNous/new-api/commit/3827fd83076b9cbaa3f1d9f9d2957ae653573ed7) | pending | 2 | 0 | review required |
| commit-3f926c3ef497 | [3f926c3ef497](https://github.com/QuantumNous/new-api/commit/3f926c3ef49745be273cd8f753fb691abdd7b836) | pending | 1 | 0 | review required |
| commit-bbf60b49f6a4 | [bbf60b49f6a4](https://github.com/QuantumNous/new-api/commit/bbf60b49f6a4cab709e49e57d45f7911fc7dccae) | pending | 2 | 0 | review required |
| commit-b22a43b6e89e | [b22a43b6e89e](https://github.com/QuantumNous/new-api/commit/b22a43b6e89e5349d8cdbddc88d570b16e15fb1f) | pending | 1 | 1 | review required |
| commit-34b19421237f | [34b19421237f](https://github.com/QuantumNous/new-api/commit/34b19421237fc4695e919cdd799c1b7813023085) | pending | 2 | 1 | review required |
| commit-ba0e478e061e | [ba0e478e061e](https://github.com/QuantumNous/new-api/commit/ba0e478e061ee0c824922d89c05620de878103cf) | pending | 2 | 0 | review required |
| commit-fc08b76f0fb3 | [fc08b76f0fb3](https://github.com/QuantumNous/new-api/commit/fc08b76f0fb3f95e7cb81667e7bf350200147fb3) | pending | 2 | 1 | review required |
| commit-4bfcf2812198 | [4bfcf2812198](https://github.com/QuantumNous/new-api/commit/4bfcf2812198913156a3e88a43e746b61a7b6ef5) | pending | 2 | 0 | review required |
| commit-9948b1ef14ea | [9948b1ef14ea](https://github.com/QuantumNous/new-api/commit/9948b1ef14ea17fb23539386c0cb09018671c955) | pending | 2 | 0 | review required |
| commit-2ed058795bf0 | [2ed058795bf0](https://github.com/QuantumNous/new-api/commit/2ed058795bf0ce968134aa8db18b0dd86b54c30e) | pending | 8 | 1 | protected files |
| commit-2b85a8fc67bd | [2b85a8fc67bd](https://github.com/QuantumNous/new-api/commit/2b85a8fc67bdb6d8d41a25f60f734dd3b9917618) | pending | 1 | 1 | review required |
| commit-767a82b87647 | [767a82b87647](https://github.com/QuantumNous/new-api/commit/767a82b8764778b847462ebecdc231eacd1e282c) | pending | 6 | 0 | review required |
| commit-ba29f0645394 | [ba29f0645394](https://github.com/QuantumNous/new-api/commit/ba29f06453942d9b5c0b32330340b8a000692c4f) | pending | 1 | 0 | review required |
| commit-b8c2dc12660d | [b8c2dc12660d](https://github.com/QuantumNous/new-api/commit/b8c2dc12660dceb2a905008a672f8bc0260c98c8) | pending | 2 | 0 | review required |
| commit-c730add954e1 | [c730add954e1](https://github.com/QuantumNous/new-api/commit/c730add954e1b565656a9320444cac5fee12cfe4) | pending | 2 | 0 | review required |
| commit-82306b3c3594 | [82306b3c3594](https://github.com/QuantumNous/new-api/commit/82306b3c35945bc0ed2c567688afe28d8a281300) | pending | 2 | 0 | review required |
| commit-bf4e1ebfb885 | [bf4e1ebfb885](https://github.com/QuantumNous/new-api/commit/bf4e1ebfb88579930d75d852d2d2bca74cbd5f62) | pending | 2 | 0 | review required |
| commit-7e25797d1468 | [7e25797d1468](https://github.com/QuantumNous/new-api/commit/7e25797d14681e4e16c2b46831450e7d6a5ccc74) | pending | 2 | 0 | review required |
| commit-7f51a97226c4 | [7f51a97226c4](https://github.com/QuantumNous/new-api/commit/7f51a97226c43bc082a065bdc8b40b719c4bdcce) | pending | 1 | 0 | review required |
| commit-1c58f3d161fa | [1c58f3d161fa](https://github.com/QuantumNous/new-api/commit/1c58f3d161fa32755b372d9449f2d5e077d50a5d) | pending | 1 | 0 | review required |
| commit-3fe394cdba28 | [3fe394cdba28](https://github.com/QuantumNous/new-api/commit/3fe394cdba280c00bc566de493b5fc8399b0a7d1) | pending | 2 | 2 | review required |
| commit-9f048107edbb | [9f048107edbb](https://github.com/QuantumNous/new-api/commit/9f048107edbba2156691718126de7c4577614533) | pending | 2 | 0 | review required |
| commit-3c9fb2a0de16 | [3c9fb2a0de16](https://github.com/QuantumNous/new-api/commit/3c9fb2a0de168d88ce5bffbf44acecdc1cdcdd39) | pending | 1 | 0 | review required |
| commit-93cf4ba6fb92 | [93cf4ba6fb92](https://github.com/QuantumNous/new-api/commit/93cf4ba6fb92e3f755fdeb6af9a9484bfa748816) | pending | 0 | 0 | review required |
| commit-dd59854050f2 | [dd59854050f2](https://github.com/QuantumNous/new-api/commit/dd59854050f26f4f67bb3dac5274af5e17013462) | pending | 1 | 0 | review required |
| commit-633b4c5b190d | [633b4c5b190d](https://github.com/QuantumNous/new-api/commit/633b4c5b190dc1af99a2814391354d86948e7943) | pending | 36 | 20 | review required |

#### commit-cae3676ec6f4

- Title: feat: glm chanel /v1/responses (#7050)
- Feature/group: feat: glm chanel /v1/responses (#7050)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/zhipu_4v/adaptor.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 智谱 Responses 原生转发
- AI recommendation: `migrate` — 本地智谱适配器的 Responses 转换仍返回 not implemented；上游补齐请求直通和 /api/v1/responses。建议作为独立渠道能力迁移，并补齐 URL、普通/流式响应的回归，保留现有 special base 分支。
- AI risks: 新端点能力取决于智谱具体服务；不能把旧 /api/paas/v4 与 special plan 混用。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/channel/zhipu_4v/adaptor.go, relay/channel/zhipu_4v/adaptor_test.go
- AI evidence: [upstream:relay/channel/zhipu_4v/adaptor.go:72](https://github.com/QuantumNous/new-api/blob/cae3676ec6f46ee5ef596443256f78c4e9b34ceb/relay/channel/zhipu_4v/adaptor.go#L72) — 新增智谱 Responses 目标端点。
- AI evidence: [downstream:relay/channel/zhipu_4v/adaptor.go:104](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/zhipu_4v/adaptor.go#L104) — 本地对应转换尚未实现。

#### commit-ba2e9287bb7a

- Title: feat(ollama): passthrough Claude Messages and OpenAI Responses (#7051)
- Feature/group: feat(ollama): passthrough Claude Messages and OpenAI Responses (#7051)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/ollama/adaptor.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Ollama 多协议原生透传
- AI recommendation: `defer` — 该提交把 Claude→OpenAI→Ollama 的原有转换改成 Claude 原生端点，并新增 Responses/compact 透传；本地仍使用原生 /api/chat。尚无目标 Ollama 版本和端点支持矩阵，不能直接改变已部署渠道的 Claude 行为。
- AI risks: 较旧 Ollama 实例未提供这些端点时会从可用转成404；需与后续每渠道OpenAI开关分组评估。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/ollama/adaptor.go:45](https://github.com/QuantumNous/new-api/blob/ba2e9287bb7a8002116c03daa4c457a330054871/relay/channel/ollama/adaptor.go#L45) — 按客户端协议选择 Claude 原生处理。
- AI evidence: [downstream:relay/channel/ollama/adaptor.go:35](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/ollama/adaptor.go#L35) — 本地 Claude 请求仍通过 Ollama 原生 Chat 转换。

#### commit-e468b73915e5

- Title: docs: update PR template and remove PR Check workflow (#7053)
- Feature/group: docs: update PR template and remove PR Check workflow (#7053)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.agents/github/ISSUE.md`, `.agents/github/PR.md`, `.github/ISSUE_TEMPLATE/bug_report.yml`, `.github/ISSUE_TEMPLATE/bug_report_en.yml`, `.github/ISSUE_TEMPLATE/feature_request.yml`, `.github/ISSUE_TEMPLATE/feature_request_en.yml`, `.github/PULL_REQUEST_TEMPLATE.md`, `.github/PULL_REQUEST_TEMPLATE/en.md`, `.github/workflows/pr-check.yml`, `AGENTS.md`
- Overlap with downstream changes: .github/PULL_REQUEST_TEMPLATE.md, .github/workflows/pr-check.yml, AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游 PR/Issue 模板和检查治理
- AI recommendation: `skip` — 上游新增专用代理模板并删除 PR Check 工作流；本地下游明确要求现有 PR 模板、人工确认和 AI 披露。属于受保护治理文件，不进入自动功能同步。
- AI risks: 模板/工作流如需调整走专门维护PR，保留 new-api/QuantumNous 归属和下游审查规则。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:.agents/github/PR.md:2](https://github.com/QuantumNous/new-api/blob/e468b73915e5028e9849de62c5018a0faa203012/.agents/github/PR.md#L2) — 固定源码：Agent-only PR body. Humans use `.github/PULL_REQUEST_TEMPLATE.md`.
- AI evidence: [downstream:AGENTS.md:166](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/AGENTS.md#L166) — 下游对照：- Always use the repository PR template at `.github/PULL_REQUEST_TEMPLATE.md` when drafting the PR title/body. Preserve the template structure and fill in the relevant sections instead of replacing it with an ad hoc format.

#### commit-692e8d6ee6a9

- Title: fix(web): restore admin unbinding for built-in providers (#6987)
- Feature/group: fix(web): restore admin unbinding for built-in providers (#6987)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`
- Overlap with downstream changes: web/src/features/users/components/dialogs/user-binding-dialog.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 管理员内置身份解绑类型修复
- AI recommendation: `migrate` — 本地解绑配置把 github_id 等存储字段发给 binding_type；上游改为 github/discord/wechat 等路由标识。建议只迁入标识修复和回归用例，保留本地手机号及微信小程序解绑入口。
- AI risks: 必须覆盖所有内置提供者，确认手机号/微信身份清理仍走现有证明与审计。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx
- AI evidence: [upstream:web/src/features/users/components/dialogs/user-binding-dialog.tsx:105](https://github.com/QuantumNous/new-api/blob/692e8d6ee6a9a1620c2d731cb51a1e3154a7042b/web/src/features/users/components/dialogs/user-binding-dialog.tsx#L105) — 固定源码：key: 'github',
- AI evidence: [downstream:web/src/features/users/components/dialogs/user-binding-dialog.tsx:107](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/users/components/dialogs/user-binding-dialog.tsx#L107) — 下游对照：key: 'github_id',
- AI evidence: [downstream:model/user.go:996](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/user.go#L996) — ClearBinding白名单将路由值github映射到存储列github_id，不接受前端当前发送的github_id。

#### commit-ac381acf4bf4

- Title: fix(billing): 修复时间规则恒真表达式导致倍率全天生效 (#6934)
- Feature/group: fix(billing): 修复时间规则恒真表达式导致倍率全天生效 (#6934)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/lib/__tests__/time-rule-expr.test.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 时间规则区间计费修复
- AI recommendation: `migrate` — 建议优先迁移：同日时段必须使用&&，跨夜才使用||，并验证时间函数取值范围和解析往返。本地仍统一使用||，9–12点会成为恒真条件导致倍率全天生效；应只移植此修复，保留后端冻结表达式与账务审计。
- AI risks: 不应默默改写已保存的旧表达式；同起止值、跨夜、非法值、组合条件及七语言提示需验证。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/pricing/lib/__tests__/time-rule-expr.test.ts, web/src/features/pricing/lib/billing-expr.ts, web/src/features/system-settings/models/tiered-pricing-editor.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- AI evidence: [upstream:web/src/features/pricing/lib/billing-expr.ts:378](https://github.com/QuantumNous/new-api/blob/ac381acf4bf41204b97bb26b4c58c83275877a2e/web/src/features/pricing/lib/billing-expr.ts#L378) — 时间函数显式限定有效整数范围；代码：const TIME_FUNC_RANGES: Record<TimeFunc, [number, number]> = {
- AI evidence: [downstream:web/src/features/pricing/lib/billing-expr.ts:738](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/pricing/lib/billing-expr.ts#L738) — 本地所有时间区间仍使用或连接，存在同日恒真缺口；代码：return `${fn} >= ${s} || ${fn} < ${e}`

#### commit-7037ac15bd8a

- Title: fix(docker): add relaykit go.mod to dev build context (#7072)
- Feature/group: fix(docker): add relaykit go.mod to dev build context (#7072)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `Dockerfile.dev`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 开发镜像模块依赖上下文
- AI recommendation: `migrate` — 本地 Dockerfile.dev 在复制 relaykit/go.mod 前执行 go mod download，而 root go.mod 有本地replace；上游补模块清单复制，建议只迁这一行兼容修复。
- AI risks: 本轮不启动 Docker；实施时可先检查模块解析上下文，后续容器构建验证需按用户授权，relaykit仍须独立构建。
- AI dependencies: none identified
- AI suggested paths (not approval): Dockerfile.dev
- AI evidence: [upstream:Dockerfile.dev:17](https://github.com/QuantumNous/new-api/blob/7037ac15bd8a29f8ee3e2b74e784bcdb75d67d22/Dockerfile.dev#L17) — 固定源码：ADD relaykit/go.mod ./relaykit/go.mod
- AI evidence: [downstream:Dockerfile.dev:14](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/Dockerfile.dev#L14) — 下游对照：ADD go.mod go.sum ./

#### commit-eb48396d5fe9

- Title: feat(task): replace built-in task adaptors with a sandboxed JS plugin system (#7076)
- Feature/group: feat(task): replace built-in task adaptors with a sandboxed JS plugin system (#7076)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `CLAUDE.md`, `THIRD-PARTY-LICENSES.md`, `common/api_type.go`, `common/api_type_task_plugin_test.go`, `common/init.go`, `common/trusted_proxies.go`, `constant/channel.go`, `constant/channel_test.go`, `constant/context_key.go`, `constant/env.go`, `constant/task.go`, `constant/task_test.go`, `controller/billing_option_test.go`, `controller/channel-billing.go`, `controller/channel-test.go`, `controller/channel.go`, `controller/channel_pin_retry_test.go`, `controller/channel_task_plugin_bind_test.go`, `controller/channel_task_plugin_validation_test.go`, `controller/channel_upstream_update.go`, `controller/log.go`, `controller/model.go`, `controller/option.go`, `controller/plugin_endpoint_test.go`, `controller/plugin_native_e2e_test.go`, `controller/plugin_protocol.go`, `controller/plugin_protocol_limiter.go`, `controller/plugin_protocol_limiter_test.go`, `controller/plugin_protocol_sdk_test.go`, `controller/plugin_protocol_test.go`, `controller/relay.go`, `controller/relay_task_plugin_test.go`, `controller/task.go`, `controller/task_generic_test.go`, `controller/task_log_view_test.go`, `controller/task_plugin.go`, `controller/task_plugin_debug.go`, `controller/task_plugin_debug_test.go`, `controller/task_plugin_test.go`, `controller/video_proxy.go`, `controller/video_proxy_gemini.go`, `docs/plugin-api/README.md`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `dto/channel_constraints.go`, `dto/channel_constraints_test.go`, `dto/plugin_protocol.go`, `dto/task.go`, `dto/task_plugin.go`, `e2e/doc_parse_test.go`, `go.mod`, `go.sum`, `logger/logger.go`, `main.go`, `middleware/auth.go`, `middleware/body_cleanup.go`, `middleware/distributor.go`, `middleware/distributor_test.go`, `middleware/jimeng_adapter.go`, `middleware/kling_adapter.go`, `middleware/logger.go`, `middleware/task_artifact_access.go`, `middleware/task_artifact_access_test.go`, `middleware/task_plugin.go`, `middleware/task_plugin_origin_task_test.go`, `middleware/task_plugin_test.go`, `middleware/trusted_proxies.go`, `middleware/utils.go`, `model/ability.go`, `model/channel.go`, `model/channel_cache.go`, `model/channel_constraint.go`, `model/channel_constraint_test.go`, `model/log.go`, `model/log_format_test.go`, `model/main.go`, `model/option.go`, `model/option_task_plugin_test.go`, `model/pricing.go`, `model/pricing_usage_schema_test.go`, `model/task.go`, `model/task_cas_test.go`, `model/task_openai_video_test.go`, `model/task_plugin.go`, `model/task_plugin_channel_select_test.go`, `model/task_plugin_test.go`, `pkg/billingexpr/compile.go`, `pkg/billingexpr/compile_usage_test.go`, `pkg/billingexpr/expr.md`, `pkg/billingexpr/run.go`, `pkg/billingexpr/settle.go`, `pkg/billingexpr/task_usage_test.go`, `pkg/billingexpr/types.go`, `pkg/jsplugin/cli.go`, `pkg/jsplugin/cli_test.go`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/fixture.go`, `pkg/jsplugin/fixture_test.go`, `pkg/jsplugin/protocol_supports_test.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_disabled_factory_test.go`, `pkg/jsplugin/registry_master_enabled_test.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/request.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`, `pkg/jsplugin/utils.go`, `plugins/.oxfmtrc.json`, `plugins/.oxlintrc.json`, `plugins/alibaba_responses_test.go`, `plugins/builtin_plugins_test.go`, `plugins/doubao_responses_test.go`, `plugins/embed.go`, `plugins/google_responses_test.go`, `plugins/hailuo_responses_test.go`, `plugins/jimeng_responses_test.go`, `plugins/kling_responses_test.go`, `plugins/sora_responses_test.go`, `plugins/sunoapi_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/google/plugin.js`, `plugins/tasks/hailuo/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `plugins/tasks/vertex-ai/plugin.js`, `plugins/tasks/vidu/plugin.js`, `plugins/vertex_ai_responses_test.go`, `plugins/video_responses_test_helpers_test.go`, `plugins/vidu_responses_test.go`, `relay/channel/adapter.go`, `relay/channel/api_request.go`, `relay/channel/api_request_test.go`, `relay/channel/minimax/relay-minimax.go`, `relay/channel/replicate/adaptor.go`, `relay/channel/task/ali/adaptor.go`, `relay/channel/task/ali/adaptor_test.go`, `relay/channel/task/ali/constants.go`, `relay/channel/task/doubao/adaptor.go`, `relay/channel/task/doubao/constants.go`, `relay/channel/task/gemini/adaptor.go`, `relay/channel/task/gemini/billing.go`, `relay/channel/task/gemini/dto.go`, `relay/channel/task/gemini/image.go`, `relay/channel/task/hailuo/adaptor.go`, `relay/channel/task/hailuo/constants.go`, `relay/channel/task/hailuo/models.go`, `relay/channel/task/jimeng/adaptor.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/auth.go`, `relay/channel/task/jsplugin/auth_test.go`, `relay/channel/task/kling/adaptor.go`, `relay/channel/task/sora/adaptor.go`, `relay/channel/task/sora/adaptor_test.go`, `relay/channel/task/sora/constants.go`, `relay/channel/task/suno/adaptor.go`, `relay/channel/task/suno/models.go`, `relay/channel/task/vertex/adaptor.go`, `relay/channel/task/vidu/adaptor.go`, `relay/channel/volcengine/adaptor.go`, `relay/channel/zhipu_4v/adaptor.go`, `relay/common/relay_info.go`, `relay/common/relay_utils.go`, `relay/common/relay_utils_test.go`, `relay/constant/relay_mode.go`, `relay/plugin_protocol.go`, `relay/plugin_protocol_test.go`, `relay/relay_adaptor.go`, `relay/relay_adaptor_jsplugin_test.go`, `relay/relay_task.go`, `relay/relay_task_test.go`, `relay/task_platform_error_test.go`, `relay/task_platform_test.go`, `relaykit/dto/channel_settings.go`, `router/api-router.go`, `router/main.go`, `router/plugin-router.go`, `router/plugin_router_test.go`, `router/relay-router.go`, `router/retired_frontend_routes_test.go`, `router/task-plugin-protocol-router.go`, `router/task-router.go`, `router/task_plugin_options_router_test.go`, `router/task_plugin_protocol_router_test.go`, `router/task_router_test.go`, `router/video-router.go`, `router/video_router_test.go`, `router/web-router.go`, `service/authz/authz_test.go`, `service/authz/resources_task_plugin.go`, `service/channel_select.go`, `service/channel_select_test.go`, `service/codex_channel_models.go`, `service/task_artifact_access.go`, `service/task_artifact_access_test.go`, `service/task_artifact_store.go`, `service/task_artifact_store_test.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/task_plugin_audit.go`, `service/task_plugin_view.go`, `service/task_plugin_view_test.go`, `service/task_polling.go`, `service/task_polling_test.go`, `setting/billing_setting/tiered_billing.go`, `setting/billing_setting/tiered_billing_test.go`, `setting/system_setting/system_setting_old.go`, `setting/system_setting/task_artifact.go`, `setting/system_setting/task_artifact_store.go`, `setting/system_setting/task_artifact_store_test.go`, `setting/system_setting/task_artifact_test.go`, `setting/task_plugin.go`, `setting/task_plugin_test.go`, `setting/task_pricing_setting/config.go`, `setting/task_pricing_setting/config_test.go`, `types/task_artifact.go`, `web/.gitignore`, `web/bun.lock`, `web/package.json`, `web/rsbuild.config.ts`, `web/scripts/sync-i18n.mjs`, `web/src/components/ai-elements/code-block.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/__tests__/channel-type-options.test.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/pricing/__tests__/breakdown-tier-match.test.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/task-expr.test.ts`, `web/src/features/pricing/__tests__/task-matrix-display.test.ts`, `web/src/features/pricing/__tests__/task-matrix.test.ts`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-billing-mode-badge.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/pricing-columns.tsx`, `web/src/features/pricing/components/pricing-sidebar.tsx`, `web/src/features/pricing/constants.ts`, `web/src/features/pricing/hooks/use-pricing-data.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-mode.ts`, `web/src/features/pricing/lib/breakdown-tier-match.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/filters.ts`, `web/src/features/pricing/lib/task-expr.ts`, `web/src/features/pricing/lib/task-matrix-display.ts`, `web/src/features/pricing/types.ts`, `web/src/features/system-settings/__tests__/task-public-address.test.ts`, `web/src/features/system-settings/general/system-info-section.tsx`, `web/src/features/system-settings/general/task-public-address.ts`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-ratio-table-columns.tsx`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/task-pricing-matrix.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/site/index.tsx`, `web/src/features/system-settings/site/section-registry.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/task-plugins/__tests__/enabled-option.test.ts`, `web/src/features/task-plugins/__tests__/marketplace-panel.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace.test.ts`, `web/src/features/task-plugins/__tests__/plugin-card.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-detail-sheet.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-icon.test.ts`, `web/src/features/task-plugins/__tests__/plugin-url.test.ts`, `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/usage-schema-table.test.tsx`, `web/src/features/task-plugins/api.ts`, `web/src/features/task-plugins/components/javascript-viewer.tsx`, `web/src/features/task-plugins/components/marketplace-capabilities.tsx`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/marketplace-panel.tsx`, `web/src/features/task-plugins/components/marketplace-plugin-card.tsx`, `web/src/features/task-plugins/components/marketplace-sources-dialog.tsx`, `web/src/features/task-plugins/components/plugin-card.tsx`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/plugin-icon.tsx`, `web/src/features/task-plugins/components/plugin-metadata-card.tsx`, `web/src/features/task-plugins/components/plugin-sandbox.tsx`, `web/src/features/task-plugins/components/plugin-source-picker.tsx`, `web/src/features/task-plugins/components/plugin-url-import-field.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`, `web/src/features/task-plugins/components/source-diff.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/features/task-plugins/components/usage-schema-table.tsx`, `web/src/features/task-plugins/index.tsx`, `web/src/features/task-plugins/lib/host-protocols.ts`, `web/src/features/task-plugins/lib/marketplace.ts`, `web/src/features/task-plugins/lib/plugin-icon.ts`, `web/src/features/task-plugins/lib/plugin-url.ts`, `web/src/features/task-plugins/types.ts`, `web/src/features/usage-logs/__tests__/access.test.ts`, `web/src/features/usage-logs/__tests__/artifacts.test.ts`, `web/src/features/usage-logs/__tests__/mobile-layout.test.ts`, `web/src/features/usage-logs/__tests__/task-details.test.ts`, `web/src/features/usage-logs/api.ts`, `web/src/features/usage-logs/components/__tests__/usage-facts.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/columns/task-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/components/dialogs/task-details-dialog.tsx`, `web/src/features/usage-logs/components/plugin-author-link.tsx`, `web/src/features/usage-logs/components/task-artifacts.tsx`, `web/src/features/usage-logs/components/timing-metrics-cell.tsx`, `web/src/features/usage-logs/components/usage-logs-mobile-card.tsx`, `web/src/features/usage-logs/components/usage-logs-provider.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/usage-logs/constants.ts`, `web/src/features/usage-logs/lib/columns.ts`, `web/src/features/usage-logs/lib/query-params.ts`, `web/src/features/usage-logs/lib/task-artifacts.ts`, `web/src/features/usage-logs/lib/task-details.ts`, `web/src/features/usage-logs/lib/task-mobile-layout.ts`, `web/src/features/usage-logs/lib/utils.ts`, `web/src/features/usage-logs/types.ts`, `web/src/hooks/use-sidebar-data.ts`, `web/src/i18n/locales/_reports/_sync-report.json`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/__tests__/localized-text.test.ts`, `web/src/lib/admin-permissions.ts`, `web/src/lib/localized-text.ts`, `web/src/routeTree.gen.ts`, `web/src/routes/_authenticated/task-plugins/index.tsx`
- Overlap with downstream changes: THIRD-PARTY-LICENSES.md, common/api_type.go, constant/channel.go, constant/context_key.go, constant/task.go, controller/channel-test.go, controller/channel.go, controller/log.go, controller/model.go, controller/option.go, controller/relay.go, controller/task.go, controller/video_proxy.go, dto/task.go, go.mod, go.sum, logger/logger.go, main.go, middleware/distributor.go, model/ability.go, model/channel.go, model/channel_cache.go, model/log.go, model/log_format_test.go, model/main.go, model/option.go, model/task.go, model/task_cas_test.go, relay/channel/task/ali/adaptor.go, relay/channel/task/ali/adaptor_test.go, relay/channel/task/doubao/adaptor.go, relay/channel/task/doubao/constants.go, relay/common/relay_info.go, relay/constant/relay_mode.go, relay/relay_adaptor.go, relay/relay_task.go, relaykit/dto/channel_settings.go, router/api-router.go, router/relay-router.go, router/video-router.go, router/web-router.go, service/authz/authz_test.go, service/task_billing.go, service/task_billing_test.go, service/task_polling.go, service/task_polling_test.go, web/rsbuild.config.ts, web/scripts/sync-i18n.mjs, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/system-settings/general/system-info-section.tsx, web/src/features/system-settings/site/index.tsx, web/src/features/system-settings/site/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/features/usage-logs/api.ts, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/columns/task-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/components/usage-logs-mobile-card.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx, web/src/features/usage-logs/constants.ts, web/src/features/usage-logs/lib/query-params.ts, web/src/features/usage-logs/lib/utils.ts, web/src/features/usage-logs/types.ts, web/src/hooks/use-sidebar-data.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/admin-permissions.ts, web/src/routeTree.gen.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 异步任务插件架构
- AI recommendation: `defer` — 该提交以沙箱 JS 驱动替换现有 Go 任务适配器，并一起引入插件管理、任务协议、产物存储及用量表达式。下游 Ali、Doubao、Volc 的素材映射、暂存、异步请求归档、冻结账务和通知尚无完整迁移设计，须作为独立架构项目评估后实施。
- AI risks: 删除现有适配器会丢失下游业务；需历史任务双读与凭据兼容、三数据库迁移、沙箱资源及网络边界验证。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/engine.go:18](https://github.com/QuantumNous/new-api/blob/eb48396d5fe97d27772d0cd5e3ca8aa5caa4f3e9/pkg/jsplugin/engine.go#L18) — 新增 JS 沙箱运行时及有界调用；代码：DefaultCallTimeout = 5 * time.Second
- AI evidence: [downstream:relay/channel/task/doubao/adaptor.go:265](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/task/doubao/adaptor.go#L265) — 下游在实际提交前按用户和渠道翻译资产 ID，后续还归档上游请求；代码：data, _, err = assetLibraryService.ResolveRequestAssetIDs(data, info.UserId, info.GetChannelID())

#### commit-0f2a2075ab07

- Title: fix(relay): 请求参数校验错误返回 HTTP 400 (#6774)
- Feature/group: fix(relay): 请求参数校验错误返回 HTTP 400 (#6774)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/relay.go`
- Overlap with downstream changes: controller/relay.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 请求校验错误状态码
- AI recommendation: `merge` — 上游仅将常规请求校验错误显式设为400并跳过重试；本地对应分支仍使用默认错误状态。可优先完整采用这项小修，保留413分支和下游Moonshot单独校验。
- AI risks: 需验证普通格式、Moonshot入口的400返回、没有上游请求/重试，不能改变413约定。
- AI dependencies: none identified
- AI suggested paths (not approval): controller/relay.go, controller/relay_validation_test.go
- AI evidence: [upstream:controller/relay.go:120](https://github.com/QuantumNous/new-api/blob/0f2a2075ab072ea7e20ffa5dd5d58dbf1b6b5b22/controller/relay.go#L120) — 上游明确400及SkipRetry。
- AI evidence: [downstream:controller/relay.go:140](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/relay.go#L140) — 本地常规请求校验还未显式设置400。

#### commit-98d50d5383a3

- Title: fix(web): recheck setup status after page reload (#6968)
- Feature/group: fix(web): recheck setup status after page reload (#6968)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/routes/__root.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 页面刷新后的初始化状态重验
- AI recommendation: `migrate` — 本地 setup_status_checked 持久化到 localStorage；切换服务实例或恢复空库后刷新仍可能跳过初始化检查。迁入仅页面会话缓存、成功响应才标记的逻辑。
- AI risks: 覆盖失败后重试、刷新、已初始化与未初始化重定向，避免重复导航请求。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/routes/__root.tsx, web/src/routes/__tests__/setup-status.test.tsx
- AI evidence: [upstream:web/src/routes/__root.tsx:111](https://github.com/QuantumNous/new-api/blob/98d50d5383a33432ff6c30b129461b170e5cbffc/web/src/routes/__root.tsx#L111) — 固定源码：let setupStatusChecked = false
- AI evidence: [downstream:web/src/routes/__root.tsx:112](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/routes/__root.tsx#L112) — 下游对照：const SETUP_CHECKED_KEY = 'setup_status_checked'

#### commit-b80d633cf586

- Title: feat(auth): encrypt password login transport
- Feature/group: feat(auth): encrypt password login transport
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/password_crypto.go`, `controller/user.go`, `main.go`, `model/main.go`, `model/password_crypto.go`, `router/api-router.go`, `web/bun.lock`, `web/package.json`, `web/src/features/auth/api.ts`, `web/src/features/auth/lib/password-encryption.ts`
- Overlap with downstream changes: controller/user.go, main.go, model/main.go, router/api-router.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 密码传输加密与持久化密钥
- AI recommendation: `defer` — 新增 RSA-OAEP 密钥表和浏览器加密协议，涉及登录 DTO、主进程初始化和前端依赖。本地下游还支持手机号密码及多种身份入口；建议作为单独认证升级，结合后续可选开关与长密码协议一起设计。
- AI risks: 应用层加密不能替代 HTTPS；要核对多副本密钥一致性、滚动升级、失效重试及手机号登录。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:common/password_crypto.go:98](https://github.com/QuantumNous/new-api/blob/b80d633cf586b001cfbb4200bae93e65abe57c2b/common/password_crypto.go#L98) — 固定源码：func DecryptPassword(ciphertextBase64 string, keyID string) (string, error) {
- AI evidence: [downstream:controller/user.go:30](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/user.go#L30) — 下游对照：type LoginRequest struct {

#### commit-8454082f930f

- Title: feat(chat): add AQBot preset (#7079)
- Feature/group: feat(chat): add AQBot preset (#7079)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `setting/chat.go`, `web/src/features/chat/lib/chat-links.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: AQBot 客户端预设导出
- AI recommendation: `defer` — 新增 aqbot://providers 深链，并把 API key 放入配置参数。本地有自有聊天客户端入口，先确认产品是否支持该客户端及用户授权展示，再决定引入。
- AI risks: 自定义协议会把密钥交给外部应用；需要用户明确操作、正确 URL 编码和禁空值回归。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/chat/lib/chat-links.ts:193](https://github.com/QuantumNous/new-api/blob/8454082f930f44593e92791c2581ffc63eb30a59/web/src/features/chat/lib/chat-links.ts#L193) — 固定源码：if (url.includes('{aqbotConfig}')) {
- AI evidence: [downstream:web/src/features/chat/lib/chat-links.ts:152](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/chat/lib/chat-links.ts#L152) — 下游对照：export function resolveChatUrl({

#### commit-918427d8ab41

- Title: feat(auth): make password encryption opt-in #6743
- Feature/group: feat(auth): make password encryption opt-in #6743
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.env.example`, `common/constants.go`, `common/init.go`, `controller/misc.go`, `controller/user.go`, `main.go`, `web/src/features/auth/api.ts`, `web/src/features/auth/sign-in/components/user-auth-form.tsx`, `web/src/features/auth/types.ts`
- Overlap with downstream changes: .env.example, common/constants.go, controller/misc.go, controller/user.go, main.go, web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/auth/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 密码传输加密默认关闭开关
- AI recommendation: `defer` — 这是前项加密功能的默认关闭和状态下发补丁；本地尚无该协议，单独引入开关不能解决登录行为。随加密升级一并评审。
- AI risks: 开启后缺少密文必须失败；不能让旧客户端在多节点切换中不可登录。
- AI dependencies: commit-b80d633cf586
- AI suggested paths (not approval):
- AI evidence: [upstream:common/constants.go:63](https://github.com/QuantumNous/new-api/blob/918427d8ab41f6adaa4113d0496f1f8621855b70/common/constants.go#L63) — 固定源码：var PasswordLoginEncryptionEnabled = false
- AI evidence: [downstream:controller/user.go:30](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/user.go#L30) — 下游对照：type LoginRequest struct {

#### commit-6c22550ea325

- Title: feat(task): resolve channel-mapped aliases and case variants for plugin models
- Feature/group: feat(task): resolve channel-mapped aliases and case variants for plugin models
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/billing_option_test.go`, `controller/option.go`, `controller/plugin_protocol_test.go`, `docs/plugin-api/v1.md`, `middleware/task_plugin.go`, `middleware/task_plugin_model.go`, `middleware/task_plugin_test.go`, `model/channel_cache.go`, `model/pricing.go`, `model/pricing_usage_schema_test.go`, `model/task_model_alias.go`, `pkg/jsplugin/model_fold.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/routing.go`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/hailuo/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `plugins/tasks/vidu/plugin.js`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/relay_task.go`, `relay/relay_task_test.go`
- Overlap with downstream changes: controller/option.go, model/channel_cache.go, relay/relay_task.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 任务别名与插件模型匹配
- AI recommendation: `defer` — 新增跨渠道聚合的插件模型别名与大小写折叠，并改写顶层 model 后再匹配插件。此处不是现有 Go 任务别名的小补丁；应随插件架构评估，保留公开名用于权限、价格和日志，真实名只发上游。
- AI risks: 同名别名指向不同插件时可能歧义；不得覆盖下游公开模型与实际计费身份。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:model/task_model_alias.go:38](https://github.com/QuantumNous/new-api/blob/6c22550ea325d4fea0e0ece52412cb1e4449291c/model/task_model_alias.go#L38) — 新增依赖 RoutingGeneration 的别名解析器；代码：func ResolveTaskModelAlias(g *jsplugin.RoutingGeneration, name string) (TaskAliasTarget, bool) {
- AI evidence: [downstream:relay/channel/task/ali/adaptor.go:147](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/task/ali/adaptor.go#L147) — 现有 Ali 根据渠道映射的实际模型选择 Wan3 路径；代码：mappedInfo.UpstreamModelName = req.Model

#### commit-66031a09d99f

- Title: fix(model): disable PostgreSQL prepared statements for pooler compatibility
- Feature/group: fix(model): disable PostgreSQL prepared statements for pooler compatibility
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `go.mod`, `go.sum`, `model/gorm_logger.go`, `model/gorm_logger_test.go`, `model/main.go`
- Overlap with downstream changes: go.mod, go.sum, model/gorm_logger.go, model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: PostgreSQL连接池与GORM预处理兼容
- AI recommendation: `migrate` — 建议将PostgreSQL显式PrepareStmt关闭并升级GORM至1.25.12，连同6eb6f35ed2的JSON Valuer和74158715cd的驱动更新组成数据库兼容批次。下游仍在观测Dialector上启用PrepareStmt，不能用上游原始postgres.New覆盖资源采集封装。
- AI risks: 需要PgBouncer类事务池验证及三数据库重复启动；保留#74真实连接字节观测和关闭旁路，不跨入未评估GORM大版本。
- AI dependencies: none identified
- AI suggested paths (not approval): go.mod, go.sum, model/main.go, model/gorm_logger.go, model/gorm_logger_test.go, model/runtime_traffic.go, model/runtime_traffic_test.go
- AI evidence: [upstream:model/main.go:146](https://github.com/QuantumNous/new-api/blob/66031a09d99f2ac4e0b94e2c41f04ed691a79304/model/main.go#L146) — PG连接同时禁用隐式与显式预处理；代码：}), newGormConfig(false))
- AI evidence: [downstream:model/main.go:143](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/main.go#L143) — 下游PG观测连接仍启用GORM显式预处理，存在池代理兼容缺口；代码：db, err := gorm.Open(dialector, newGormConfig(true))
- AI evidence: [upstream:model/main.go:145](https://github.com/QuantumNous/new-api/blob/66031a09d99f2ac4e0b94e2c41f04ed691a79304/model/main.go#L145) — PG分支保留simple protocol且该调用使用newGormConfig(false)；代码：PreferSimpleProtocol: true,

#### commit-0bee5d441029

- Title: fix(ali): honor image response format (#5513) (#7048)
- Feature/group: fix(ali): honor image response format (#5513) (#7048)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/ali/adaptor_test.go`, `relay/channel/ali/image.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 阿里图片输出格式
- AI recommendation: `merge` — 本地从 Gin context 字符串读取response_format，上游改从已解析ImageRequest读取并附URL/Base64测试；可优先完整采用，修复b64_json被忽略。
- AI risks: 下载图片转换Base64仍须沿用SSRF与大小限制，默认/URL响应不应额外下载。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/channel/ali/image.go, relay/channel/ali/adaptor_test.go
- AI evidence: [upstream:relay/channel/ali/image.go:289](https://github.com/QuantumNous/new-api/blob/0bee5d4410296e972bf0076414ade786c2c799c8/relay/channel/ali/image.go#L289) — 格式从真实图片请求读取。
- AI evidence: [downstream:relay/channel/ali/image.go:287](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/ali/image.go#L287) — 本地仍依赖未必写入的Context字段。

#### commit-dc4732cfed71

- Title: feat(web): factory task plugins update only with the system
- Feature/group: feat(web): factory task plugins update only with the system
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/marketplace.test.ts`, `web/src/features/task-plugins/components/marketplace-plugin-card.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`, `web/src/features/task-plugins/lib/marketplace.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 内置插件更新策略
- AI recommendation: `defer` — 市场区分内置版本与自定义覆盖版本，内置插件只随系统更新。当前仓库没有插件市场或覆盖层，独立导入界面无可操作对象；待插件版本及回滚策略确定后纳入。
- AI risks: 需要与禁用、覆盖回退和系统发版生命周期一致。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/task-plugins/components/plugins-table.tsx:46](https://github.com/QuantumNous/new-api/blob/dc4732cfed712b004d3d3d94a414d3ff127a93cc/web/src/features/task-plugins/components/plugins-table.tsx#L46) — 界面识别陈旧的内置覆盖版本；代码：import { isStaleFactoryOverride } from '../lib/marketplace'

#### commit-b5b94bc685fd

- Title: fix(subscription): 无有效订阅时前端如实显示「仅用订阅」偏好 (#6222) (#7086)
- Feature/group: fix(subscription): 无有效订阅时前端如实显示「仅用订阅」偏好 (#6222) (#7086)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/wallet/components/subscription-plans-card.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 订阅偏好如实显示
- AI recommendation: `merge` — 建议优先合入。界面直接展示已保存的 subscription_only，且无有效订阅时明确请求将被拒绝；本地仍显示wallet_first并暗示自动用钱包，与仅订阅语义矛盾。整提交git apply --check通过，实施时仍需验证现有钱包与订阅组合。
- AI risks: 七语言文案和禁用选项必须一致；只修展示，不改变扣费偏好或后台钱包路径。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/wallet/components/subscription-plans-card.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- AI evidence: [upstream:web/src/features/wallet/components/subscription-plans-card.tsx:333](https://github.com/QuantumNous/new-api/blob/b5b94bc685fd2251551df826dab3575ab262f6dc/web/src/features/wallet/components/subscription-plans-card.tsx#L333) — 上游选择器展示真实保存偏好；代码：value={billingPreference}
- AI evidence: [downstream:web/src/features/wallet/components/subscription-plans-card.tsx:198](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/wallet/components/subscription-plans-card.tsx#L198) — 本地会把仅订阅偏好展示成钱包优先；代码：disablePref && isSubPref ? 'wallet_first' : billingPreference

#### commit-1751f43ee07e

- Title: fix(sqlite): enable WAL + working busy timeout + _txlock=immediate to stop concurrent write lockouts (#7030)
- Feature/group: fix(sqlite): enable WAL + working busy timeout + _txlock=immediate to stop concurrent write lockouts (#7030)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/database.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: SQLite并发写与WAL
- AI recommendation: `migrate` — 建议迁入glebarez/modernc能识别的_pragma=busy_timeout、WAL与_txlock=immediate。下游仍使用_busy_timeout=30000形式，未建立上游声明的写锁与busy_timeout契约；需验证旧库升级及现有账务多步事务后采用。
- AI risks: BEGIN IMMEDIATE提前持有写锁，需确认钱包/素材/迁移事务无不必要长锁；保持自定义SQLite路径、只读场景和三数据库分支。
- AI dependencies: none identified
- AI suggested paths (not approval): common/database.go, common/database_test.go, model/main_test.go
- AI evidence: [upstream:common/database.go:64](https://github.com/QuantumNous/new-api/blob/1751f43ee07edc9eb0c56fd9b23586861b43df46/common/database.go#L64) — 默认SQLite DSN改为驱动识别的pragma及immediate事务；代码：var SQLitePath = "one-api.db?_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_txlock=immediate"
- AI evidence: [downstream:common/database.go:44](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/common/database.go#L44) — 下游默认DSN仍是旧语法；代码：var SQLitePath = "one-api.db?_busy_timeout=30000"

#### commit-6eb6f35ed211

- Title: fix(model): return string from JSON column Valuers for pg simple protocol
- Feature/group: fix(model): return string from JSON column Valuers for pg simple protocol
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/channel.go`, `model/json_column_test.go`, `model/main.go`, `model/prefill_group.go`, `model/task.go`
- Overlap with downstream changes: model/channel.go, model/main.go, model/task.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: JSON列读写适配PG simple protocol
- AI recommendation: `migrate` — 建议和PG预处理关闭一起迁移：JSON列Valuer返回string而非[]byte，Scanner同时接受string/[]byte并保留nil。下游还在返回序列化字节，直接关闭预处理可能让渠道/任务JSON写入失败；任务结构含下游归档和凭据字段，须按现有结构适配测试。
- AI risks: 不能清空旧string列值或丢失下游TaskPrivateData；JSON操作仍走common包装，三数据库往返与nil语义都需验证。
- AI dependencies: commit-66031a09d99f
- AI suggested paths (not approval): model/channel.go, model/json_column_test.go, model/main.go, model/prefill_group.go, model/task.go
- AI evidence: [upstream:model/channel.go:172](https://github.com/QuantumNous/new-api/blob/6eb6f35ed211b7459cae3b9f13286b9c93fc1bd6/model/channel.go#L172) — ChannelInfo的Valuer返回文本JSON；代码：return string(b), nil
- AI evidence: [downstream:model/channel.go:170](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/channel.go#L170) — 下游ChannelInfo仍返回Marshal的[]byte；代码：return common.Marshal(&c)

#### commit-b518d0033b67

- Title: fix(relay): bound the wait for upstream response headers (fixes unbounded heap growth → OOM) (#6949)
- Feature/group: fix(relay): bound the wait for upstream response headers (fixes unbounded heap growth → OOM) (#6949)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.env.example`, `README.md`, `common/constants.go`, `common/init.go`, `service/http_client.go`
- Overlap with downstream changes: .env.example, README.md, common/constants.go, service/http_client.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游响应头等待超时
- AI recommendation: `migrate` — 本地relay Transport没有响应头超时，上游新增1800秒默认、0关闭和Duration溢出保护。建议选择性迁移配置与Transport改动；原提交含受保护README，不能整提交自动采用。
- AI risks: 非流式大模型通常生成完毕才返回头；需验证长生成、0禁用、异常大配置、代理及stream头后不截断。
- AI dependencies: none identified
- AI suggested paths (not approval): common/constants.go, common/init.go, service/http_client.go, .env.example, service/http_client_test.go
- AI evidence: [upstream:service/http_client.go:115](https://github.com/QuantumNous/new-api/blob/b518d0033b670f5518b8a2f1cf8ea0142a9d1b8d/service/http_client.go#L115) — 仅限制等待响应头并做时长夹取。
- AI evidence: [downstream:service/http_client.go:102](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/http_client.go#L102) — 本地Transport只设置空闲连接等超时，未设响应头上限。

#### commit-74158715cde6

- Title: fix initialize database
- Feature/group: fix initialize database
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `go.mod`, `go.sum`
- Overlap with downstream changes: go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 数据库驱动配套升级
- AI recommendation: `upgrade` — 建议在同一数据库兼容批次把mysql驱动1.4.3→1.5.7、postgres驱动1.5.2→1.5.9，与GORM1.25.12和JSON Valuer协同验证；只更新相关依赖，不照抄本提交无关eventstream直接依赖标记。
- AI risks: 驱动更新会影响迁移元数据与预处理语义，必须验证旧库、重复启动和观测Dialector组合。
- AI dependencies: commit-66031a09d99f, commit-6eb6f35ed211
- AI suggested paths (not approval): go.mod, go.sum
- AI evidence: [upstream:go.mod:60](https://github.com/QuantumNous/new-api/blob/74158715cde6d7b767ead23d9a2af64b7b58a588/go.mod#L60) — 上游锁定配套postgres驱动版本；代码：gorm.io/driver/postgres v1.5.9
- AI evidence: [downstream:go.mod:62](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/go.mod#L62) — 下游仍锁定旧PG驱动；代码：gorm.io/driver/postgres v1.5.2

#### commit-69a41eeadc81

- Title: fix(model): drop leftover prefill_groups unique constraints before AutoMigrate (#7100)
- Feature/group: fix(model): drop leftover prefill_groups unique constraints before AutoMigrate (#7100)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/main.go`, `model/prefill_group_migration.go`
- Overlap with downstream changes: model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 预设组唯一约束修复初稿
- AI recommendation: `skip` — 该初稿被紧接着的2bf0820f4b撤销，随后2b6f1dfefb以更完整事务和索引核验重做。本轮不应用再撤销；评估最终2b6f与007d的合并行为，保留本地迁移链。
- AI risks: 直接采用过时初稿会遗漏失败回滚、索引定义和旧命名覆盖。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:model/prefill_group_migration.go:22](https://github.com/QuantumNous/new-api/blob/69a41eeadc81adb08d04346512c76d62fd6203db/model/prefill_group_migration.go#L22) — 这是后续被撤销的初稿迁移函数；代码：func migratePrefillGroupUniqueIndex(db *gorm.DB) error {

#### commit-2bf0820f4b89

- Title: Revert "fix(model): drop leftover prefill_groups unique constraints before Au…" (#7101)
- Feature/group: Revert "fix(model): drop leftover prefill_groups unique constraints before Au…" (#7101)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/main.go`, `model/prefill_group_migration.go`
- Overlap with downstream changes: model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 撤销预设组迁移初稿
- AI recommendation: `skip` — 它仅撤销上一条69a41eeadc，而下游没有引入该初稿；对当前基线无独立功能。保留来源与被取代关系，不执行无意义的正反补丁。
- AI risks: 不要把撤销看作应删除最终修复或本地下游迁移。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:model/main.go:317](https://github.com/QuantumNous/new-api/blob/2bf0820f4b89530acf14d389ba3e8229211933fa/model/main.go#L317) — 撤销后migrateDB不再调用初稿辅助函数；代码：func migrateDB() error {
- AI evidence: [parent:model/prefill_group_migration.go:22](https://github.com/QuantumNous/new-api/blob/69a41eeadc81adb08d04346512c76d62fd6203db/model/prefill_group_migration.go#L22) — 该项父提交中的初稿文件被本次整体删除；代码：func migratePrefillGroupUniqueIndex(db *gorm.DB) error {

#### commit-2b6f1dfefbe2

- Title: fix(model): drop leftover prefill_groups unique constraints before AutoMigrate
- Feature/group: fix(model): drop leftover prefill_groups unique constraints before AutoMigrate
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `model/main.go`, `model/prefill_group_migration.go`, `model/prefill_group_migration_test.go`
- Overlap with downstream changes: AGENTS.md, model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: PostgreSQL预设组唯一约束迁移
- AI recommendation: `migrate` — 建议移植最终安全迁移并和007d69942d按定义识别旧索引的修正一起采用。只加入PrefillGroup迁移入口与验证，不同步删除migrateDBFast，不覆盖AGENTS或本地钱包、成本、存储、身份迁移。
- AI risks: 须在锁表事务中核验/替换目标唯一索引并失败回滚，保留复合/表达式/部分索引；SQLite/MySQL分支应无越界DDL。
- AI dependencies: commit-66031a09d99f, commit-6eb6f35ed211, commit-74158715cde6
- AI suggested paths (not approval): model/main.go, model/prefill_group_migration.go, model/prefill_group_migration_test.go
- AI evidence: [upstream:model/prefill_group_migration.go:130](https://github.com/QuantumNous/new-api/blob/2b6f1dfefbe217fed31fc0726717cc7de6958e8e/model/prefill_group_migration.go#L130) — 最终迁移独立核验约束与索引；代码：func migratePrefillGroupUniqueness(db *gorm.DB) error {
- AI evidence: [downstream:model/main.go:317](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/main.go#L317) — 下游有必须保留的历史成本迁移前置步骤；代码：if err := ensureCostAccountingAdjustmentEventKeyColumn(DB); err != nil {

#### commit-8c8c4153d4b8

- Title: fix(log): preserve quota in usage statistics (#7108)
- Feature/group: fix(log): preserve quota in usage statistics (#7108)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/log.go`
- Overlap with downstream changes: model/log.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: True
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 日志统计quota不被RPM扫描覆盖
- AI recommendation: `covered` — 该项唯一行为是把RPM/TPM扫描到独立结构再赋值，避免覆盖已累加quota。下游sumLogStatistics已使用rateStatistics并分别赋Rpm/Tpm，且来源文档记录#21同步；可判本单项完整覆盖，无需替换包含有效调账统计的本地实现。
- AI risks: 本项covered仅指扫描隔离，不能因此覆盖下游有效消费和财务统计实现。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:model/log.go:685](https://github.com/QuantumNous/new-api/blob/8c8c4153d4b80d54352d21593de41aa9a6178f7e/model/log.go#L685) — 上游把速率查询扫描到独立结构；代码：if err := rpmTpmQuery.Scan(&rateStat).Error; err != nil {
- AI evidence: [downstream:model/log.go:802](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/log.go#L802) — 下游同样隔离扫描，保持现有statistics.Quota；代码：if err := rateQuery.Scan(&rateStatistics).Error; err != nil {
- AI evidence: [downstream:model/log.go:806](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/log.go#L806) — 独立结果仅覆盖RPM字段；代码：statistics.Rpm = rateStatistics.Rpm
- AI evidence: [downstream:model/log.go:807](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/log.go#L807) — 独立结果仅覆盖TPM字段；代码：statistics.Tpm = rateStatistics.Tpm

#### commit-27ff6a8767e7

- Title: fix(model): migrate legacy token key constraints
- Feature/group: fix(model): migrate legacy token key constraints
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/main.go`, `model/token_migration.go`, `model/token_migration_test.go`
- Overlap with downstream changes: model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: PG旧Token Key约束迁移
- AI recommendation: `migrate` — 建议把Token Key唯一约束迁移按最终PG模式加入启动前置，解决GORM旧约束名与当前唯一索引不一致。下游Token.Key已使用uniqueIndex但没有该兼容迁移，需要在旧PG样本上验证，非PG保持原迁移。
- AI risks: 必须保留Key唯一性和历史数据，未知约束/外键依赖不能静默删除；失败必须事务回滚。
- AI dependencies: commit-66031a09d99f, commit-6eb6f35ed211, commit-74158715cde6, commit-2b6f1dfefbe2
- AI suggested paths (not approval): model/main.go, model/token_migration.go, model/token_migration_test.go
- AI evidence: [upstream:model/token_migration.go:129](https://github.com/QuantumNous/new-api/blob/27ff6a8767e728f879d52770c273d4f73214a430/model/token_migration.go#L129) — 独立PG Token key唯一性迁移；代码：func migrateTokenKeyUniqueness(db *gorm.DB) error {
- AI evidence: [downstream:model/token.go:17](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/token.go#L17) — 下游Token Key使用唯一索引，旧约束形态需兼容；代码：Key                string         `json:"key" gorm:"type:varchar(128);uniqueIndex"`

#### commit-67a0585d0f25

- Title: fix(docs): correct Video API links across localized READMEs (#7116)
- Feature/group: fix(docs): correct Video API links across localized READMEs (#7116)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `README.en.md`, `README.fr.md`, `README.ja.md`, `README.md`, `README.zh_CN.md`, `README.zh_TW.md`
- Overlap with downstream changes: README.md, README.zh_CN.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游多语言 README Video API 链接
- AI recommendation: `skip` — 仅更新上游 README 的文档链接，涉及受保护 README 和下游已有产品介绍。不能用功能迁移改写；维护者可另行确认需要的单个链接。
- AI risks: 上游文档路径也可能变化，另行编辑时应验证链接并保留所有归属和长期差异台账。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:README.en.md:284](https://github.com/QuantumNous/new-api/blob/67a0585d0f252dfca445c11b7600971b7eeb8eea/README.en.md#L284) — 固定源码：- [Video Interface (Video)](https://docs.newapi.pro/en/docs/api/ai-model/videos/sora/createvideo)
- AI evidence: [downstream:README.md:11](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/README.md#L11) — 下游对照：[![Upstream](https://img.shields.io/badge/upstream-QuantumNous%2Fnew--api-7357d9.svg)](https://github.com/QuantumNous/new-api)

#### commit-b7017c251bad

- Title: fix(model): do not treat no-op system task state writes as lock loss (#7135)
- Feature/group: fix(model): do not treat no-op system task state writes as lock loss (#7135)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/system_task.go`, `model/system_task_test.go`
- Overlap with downstream changes: model/system_task.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: MySQL系统任务无变更写入保租约
- AI recommendation: `migrate` — 建议优先迁移：相同state在同一秒再次保存时，MySQL RowsAffected=0并不代表未持锁。下游仍直接报ErrSystemTaskLockLost；应按同一now重新核验租约并补真实MySQL回归，而不是只运行SQLite测试。
- AI risks: 真正过期/换持有者仍必须失败；重复状态不能使过期任务续命，保留有界并发与历史账务系统任务。
- AI dependencies: none identified
- AI suggested paths (not approval): model/system_task.go, model/system_task_test.go, model/system_task_database_test.go
- AI evidence: [upstream:model/system_task.go:333](https://github.com/QuantumNous/new-api/blob/b7017c251badaacaab840646a959635d00665e2d/model/system_task.go#L333) — 零变更时额外查询相同条件核实租约；代码：var held int64
- AI evidence: [downstream:model/system_task.go:311](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/system_task.go#L311) — 本地UpdateSystemTaskState以下仍把0行写入直接判为锁丢失；代码：func UpdateSystemTaskState(taskID string, lockedBy string, state any) error {
- AI evidence: [downstream:model/system_task.go:327](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/system_task.go#L327) — UpdateSystemTaskState精确分支：if result.RowsAffected == 0 {

#### commit-0ed497f066a6

- Title: feat(relay): hosted-tool conversion fidelity, reasoning normalization, and billing usage integrity (#7137)
- Feature/group: feat(relay): hosted-tool conversion fidelity, reasoning normalization, and billing usage integrity (#7137)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.gitignore`, `controller/channel-test.go`, `controller/relay.go`, `controller/relay_count_tokens_test.go`, `model/channel.go`, `relay/channel/aws/adaptor.go`, `relay/channel/aws/relay_aws_test.go`, `relay/channel/claude/adaptor.go`, `relay/channel/claude/adaptor_test.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_claude_test.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/gemini/adaptor.go`, `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_responses.go`, `relay/channel/newapi/adaptor.go`, `relay/channel/openai/adaptor.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/chat_via_responses_test.go`, `relay/channel/openai/helper.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/responses_via_chat.go`, `relay/channel/sub2api/adaptor_test.go`, `relay/channel/vertex/adaptor.go`, `relay/channel/zhipu_4v/adaptor.go`, `relay/chat_completions_via_responses.go`, `relay/chat_completions_via_responses_test.go`, `relay/claude_handler.go`, `relay/common/conversion_diagnostics.go`, `relay/common/override.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relay/common/tool_usage.go`, `relay/compatible_handler.go`, `relay/convert_request_error.go`, `relay/convert_request_error_test.go`, `relay/gemini_handler.go`, `relay/helper/price.go`, `relay/helper/price_test.go`, `relay/helper/reasoning_suffix.go`, `relay/helper/reasoning_suffix_test.go`, `relay/responses_handler.go`, `relaykit/README.md`, `relaykit/dto/billing_usage.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/channel_settings_test.go`, `relaykit/dto/claude.go`, `relaykit/dto/gemini.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_response.go`, `relaykit/dto/reasoning_state.go`, `relaykit/dto/usage_merge.go`, `relaykit/dto/usage_merge_test.go`, `relaykit/reasonmap/reasonmap.go`, `relaykit/relayconvert/claude_default_max_tokens_test.go`, `relaykit/relayconvert/convmeta/meta.go`, `relaykit/relayconvert/convmeta/meta_test.go`, `relaykit/relayconvert/convmeta/options.go`, `relaykit/relayconvert/golden_test.go`, `relaykit/relayconvert/internal/claude_messages/citations.go`, `relaykit/relayconvert/internal/claude_messages/stream_billing_usage_test.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_responses_hosted_stream.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_responses_req.go`, `relaykit/relayconvert/internal/gemini_chat/grounding.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_req.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_responses_hosted_stream.go`, `relaykit/relayconvert/internal/oai_chat/citations.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_stream_resp_test.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp_test.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_stream_resp.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/claude/usage.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/internal/toolconv/decode.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/hosted_values.go`, `relaykit/relayconvert/internal/toolconv/model.go`, `relaykit/relayconvert/internal/toolconv/policy_test.go`, `relaykit/relayconvert/internal/toolconv/response.go`, `relaykit/relayconvert/internal/toolconv/response_artifacts.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/gemini.go`, `relaykit/relayconvert/reasoning/intent.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/reasoning/suffix.go`, `relaykit/relayconvert/reasoning/suffix_test.go`, `relaykit/relayconvert/request_compat.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`, `relaykit/relayconvert/response_compat.go`, `relaykit/relayconvert/response_registry.go`, `relaykit/relayconvert/response_registry_test.go`, `relaykit/relayconvert/terminal_stream_test.go`, `relaykit/relayconvert/testdata/golden/request/claude_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/request/claude_to_openai.golden.json`, `relaykit/relayconvert/testdata/golden/request/claude_to_openai_responses.golden.json`, `relaykit/relayconvert/testdata/golden/request/gemini_to_claude.golden.json`, `relaykit/relayconvert/testdata/golden/request/gemini_to_openai.golden.json`, `relaykit/relayconvert/testdata/golden/request/gemini_to_openai_responses.golden.json`, `relaykit/relayconvert/testdata/golden/request/openai_responses_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/request/openai_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/response/claude_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/response/gemini_to_claude.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_responses_to_claude.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_responses_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_responses_to_openai.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_to_claude.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_to_openai_responses.golden.json`, `relaykit/relayconvert/testdata/golden/stream/claude_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/stream/gemini_to_claude.golden.json`, `relaykit/relayconvert/text_converter_registry.go`, `relaykit/relayconvert/text_converter_registry_test.go`, `relaykit/relayconvert/tool_loss_policy_test.go`, `relaykit/types/conversion.go`, `router/relay-router.go`, `router/relay_router_test.go`, `service/billing_session.go`, `service/billing_usage.go`, `service/log_info_generate.go`, `service/quota.go`, `service/request_converter.go`, `service/response_converter.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/token_counter.go`, `setting/model_setting/global.go`, `setting/ratio_setting/model_ratio.go`, `setting/reasoning/suffix.go`
- Overlap with downstream changes: .gitignore, controller/channel-test.go, controller/relay.go, model/channel.go, relay/channel/openai/adaptor.go, relay/channel/openai/helper.go, relay/channel/openai/relay-openai.go, relay/chat_completions_via_responses.go, relay/chat_completions_via_responses_test.go, relay/claude_handler.go, relay/common/relay_info.go, relay/compatible_handler.go, relay/responses_handler.go, relaykit/dto/channel_settings.go, relaykit/dto/channel_settings_test.go, relaykit/dto/openai_request.go, router/relay-router.go, router/relay_router_test.go, service/billing_session.go, service/quota.go, service/text_quota.go, setting/ratio_setting/model_ratio.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 托管工具、推理意图和计费快照重构
- AI recommendation: `defer` — 149文件重构引入托管工具转换、reasoning state、usage合并、诊断和计费快照，同时改控制器及多个协议。下游有Moonshot/KVV、饱和审计、财务快照等额外边界，尚未完成跨协议/计费矩阵证明，暂不整包引入。
- AI risks: 需把后续7170、模型修饰符、缓存明细及失败流计费一起评估；README受保护，relaykit独立公共接口也会改变。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/claude/adaptor.go:33](https://github.com/QuantumNous/new-api/blob/0ed497f066a68613375124303ef54f220267b334/relay/channel/claude/adaptor.go#L33) — 原生Claude适配开始调用统一思考模型转换。
- AI evidence: [downstream:relay/channel/claude/adaptor.go:29](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/claude/adaptor.go#L29) — 本地原生Claude目前保持请求直通。

#### commit-bbd97446c260

- Title: fix(relay): follow-up billing integrity and conversion completions (#7170)
- Feature/group: fix(relay): follow-up billing integrity and conversion completions (#7170)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/claude/adaptor.go`, `relay/channel/claude/adaptor_test.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_gemini_usage_test.go`, `relay/channel/openai/helper.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relaykit/dto/billing_usage.go`, `relaykit/dto/billing_usage_test.go`, `relaykit/dto/usage_merge.go`, `relaykit/dto/usage_merge_test.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/toolconv/decode.go`, `relaykit/relayconvert/internal/toolconv/encode.go`
- Overlap with downstream changes: relay/channel/openai/helper.go, relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 托管工具计费完整性后续修复
- AI recommendation: `defer` — 新增Gemini→Claude请求/响应以及Claude→Gemini流状态，依赖7137的新usage/转换基础。本地Claude的Gemini转换仍未实现，不能将本后续补丁孤立套入旧接口。
- AI risks: Gemini message_start用量、缓存与工具调用快照要与7137整组对照并验证不重复扣费。
- AI dependencies: commit-0ed497f066a6
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/claude/adaptor.go:28](https://github.com/QuantumNous/new-api/blob/bbd97446c26092f2e7250af429096064b9e0f899/relay/channel/claude/adaptor.go#L28) — 新实现依赖统一转换服务。
- AI evidence: [downstream:relay/channel/claude/adaptor.go:23](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/claude/adaptor.go#L23) — 本地此协议路径仍是not implemented。

#### commit-aece11d2f7f0

- Title: feat(plugin): add MiniMax-H3 /v2 video generation to the hailuo task … (#7168)
- Feature/group: feat(plugin): add MiniMax-H3 /v2 video generation to the hailuo task … (#7168)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.md`, `plugins/hailuo_responses_test.go`, `plugins/tasks/hailuo/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/relay_task.go`, `service/task_polling.go`
- Overlap with downstream changes: relay/relay_task.go, service/task_polling.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: MiniMax H3 视频模型
- AI recommendation: `defer` — 新增 H3 的 v2 请求、4–15 秒时长、768P/2K 分辨率和轮询上下文。模型接入有价值，但本地仍使用 Hailuo Go 适配器，应先确定实际渠道需求与计价，再单独迁移，不能直接加入只适配 JS 的驱动。
- AI risks: H3 输入媒体计费随后由 73afad588c 修复，不能只取初始实现；需保留既有视频任务审计。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/hailuo/plugin.js:119](https://github.com/QuantumNous/new-api/blob/aece11d2f7f095a33052696c5d28d3656609a99e/plugins/tasks/hailuo/plugin.js#L119) — H3 请求明确校验模型专属时长；代码：function h3Duration(req) {
- AI evidence: [downstream:relay/channel/task/hailuo/adaptor.go:185](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/task/hailuo/adaptor.go#L185) — 本地完成状态解析仍通过 Go Hailuo 适配器；代码：func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {

#### commit-d8ca0ed0bb59

- Title: chore: let owners use human PR templates
- Feature/group: chore: let owners use human PR templates
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`
- Overlap with downstream changes: AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 项目所有者 PR 模板例外
- AI recommendation: `skip` — 只修改上游 AGENTS 对owner/agent模板的分流；下游已由明确规则统一使用本仓库 PR 模板。无需引入上游例外或改写受保护治理。
- AI risks: 人类确认不能由AI代勾，现有本地授权优先，治理调整另开专门PR。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:AGENTS.md:163](https://github.com/QuantumNous/new-api/blob/d8ca0ed0bb596e910e1955461e7691e40d224d70/AGENTS.md#L163) — 固定源码：- When the pull request is created for the project owner, use the ordinary human PR template: `.github/PULL_REQUEST_TEMPLATE.md` for Chinese requests or `.github/PULL_REQUEST_TEMPLATE/en.md` for English requests. Project-owner pull requests MUST NOT use `.agents/github/PR.md` unless the owner explicitly asks for it.
- AI evidence: [downstream:AGENTS.md:166](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/AGENTS.md#L166) — 下游对照：- Always use the repository PR template at `.github/PULL_REQUEST_TEMPLATE.md` when drafting the PR title/body. Preserve the template structure and fill in the relevant sections instead of replacing it with an ad hoc format.

#### commit-73afad588ca7

- Title: fix(plugin): account for MiniMax-H3 input media usage (#7171)
- Feature/group: fix(plugin): account for MiniMax-H3 input media usage (#7171)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/hailuo_responses_test.go`, `plugins/tasks/hailuo/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: MiniMax H3 输入媒体计费
- AI recommendation: `defer` — 补齐 H3 输入图片数与参考视频秒数，用最大值预留、实际用量结算。它修正 H3 首版的少计费风险，若未来接入 H3 应一并迁移，目前没有 H3 驱动和对应账务事实模型。
- AI risks: 图片/视频元数据必须有界，额度不足须在上游提交前拒绝；不能按输出时长替代输入媒体成本。
- AI dependencies: commit-eb48396d5fe9, commit-aece11d2f7f0
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/hailuo/plugin.js:47](https://github.com/QuantumNous/new-api/blob/73afad588ca7af07134fa423e8a33fdea6c855b2/plugins/tasks/hailuo/plugin.js#L47) — H3 用量包含输入视频秒数；代码：input_video_seconds: {

#### commit-057f71c2336c

- Title: fix(logs): isolate privileged metadata
- Feature/group: fix(logs): isolate privileged metadata
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.gitignore`, `controller/channel-test.go`, `controller/channel_test_internal_test.go`, `controller/log.go`, `controller/relay.go`, `controller/relay_error_log_test.go`, `model/log.go`, `model/log_format_test.go`, `model/log_other.go`, `model/log_other_test.go`, `relay/convert_request_error_test.go`, `service/billing_usage.go`, `service/channel_affinity.go`, `service/log_info_generate.go`, `service/midjourney.go`, `service/quota_saturation_test.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/task_plugin_audit.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/violation_fee.go`, `web/src/features/usage-logs/components/__tests__/reject-reason.test.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/types.ts`
- Overlap with downstream changes: .gitignore, controller/channel-test.go, controller/log.go, controller/relay.go, model/log.go, model/log_format_test.go, service/midjourney.go, service/task_billing.go, service/task_billing_test.go, service/text_quota.go, service/violation_fee.go, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 日志特权元数据隔离
- AI recommendation: `migrate` — 建议优先迁移读取侧按user/admin/root投影及遗留字段脱敏，使用RawMessage保留大整数，并保留下游request_body/ref/available、财务权限与饱和审计。当前用户格式化仅在命中部分键时清理map，未统一处理root_info/channel/reject_reason。不要为此强制替换所有写入接口为上游LogOther。
- AI risks: 格式错误JSON需非Root安全失败；逐角色防泄露，同时保留Owner可见计价字段、有效调账和管理员永久归档入口。
- AI dependencies: none identified
- AI suggested paths (not approval): model/log.go, model/log_other.go, model/log_other_test.go, model/log_format_test.go, controller/log.go, controller/relay_error_log_test.go, service/log_info_generate.go, service/quota_saturation_test.go, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/components/__tests__/reject-reason.test.tsx, web/src/features/usage-logs/types.ts
- AI evidence: [upstream:model/log_other.go:236](https://github.com/QuantumNous/new-api/blob/057f71c2336c3981187b732a9d06f65490e9a946/model/log_other.go#L236) — 读取侧以RawMessage隔离可见域，避免float64破坏大整数；代码：var values map[string]json.RawMessage
- AI evidence: [downstream:model/log.go:126](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/log.go#L126) — 本地靠部分敏感键字符串命中进入map清理，特权键覆盖不同；代码：containsAdminOnlyData := strings.Contains(otherText, `"admin_info"`) ||

#### commit-219c9e06341f

- Title: 优化匿名冷启动与公开内容接口的重复回源请求 (#7166)
- Feature/group: 优化匿名冷启动与公开内容接口的重复回源请求 (#7166)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/misc.go`, `controller/revalidated_response.go`, `docs/authentication.md`, `service/auth_session.go`, `web/src/features/about/api.ts`, `web/src/features/home/api.ts`, `web/src/features/legal/api.ts`, `web/src/lib/api.ts`, `web/src/lib/auth-session.ts`, `web/src/lib/session-hint.ts`, `web/src/routes/(auth)/sign-in.tsx`, `web/src/routes/_authenticated/route.tsx`
- Overlap with downstream changes: controller/misc.go, service/auth_session.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 匿名冷启动与公开内容重验证
- AI recommendation: `migrate` — 建议将公开内容弱ETag、会话hint优化及受保护路由resolveAuthentication一起迁入，并采用36dbbf0f77最终稳定ETag。hint仅减少公共页匿名刷新，不能作为鉴权结论；下游短信、微信、24小时会话仍复用真实刷新凭据和撤销检查。
- AI risks: 旧有效RefreshCookie无hint仍需在受保护路由恢复；清Cookie、跨标签、撤销、gzip及多副本304必须验证。
- AI dependencies: none identified
- AI suggested paths (not approval): controller/misc.go, controller/revalidated_response.go, service/auth_session.go, web/src/features/about/api.ts, web/src/features/home/api.ts, web/src/features/legal/api.ts, web/src/lib/api.ts, web/src/lib/auth-session.ts, web/src/lib/session-hint.ts, web/src/routes/(auth)/sign-in.tsx, web/src/routes/_authenticated/route.tsx, service/auth_session_test.go, web/src/lib/__tests__/auth-session.test.ts, controller/revalidated_response_test.go
- AI evidence: [upstream:web/src/lib/auth-session.ts:372](https://github.com/QuantumNous/new-api/blob/219c9e06341f1b100e2c572a5f97c45f151fd280/web/src/lib/auth-session.ts#L372) — 权限相关路由始终可向服务端重新确认会话；代码：export async function resolveAuthentication(): Promise<RefreshOutcome> {
- AI evidence: [downstream:service/auth_session.go:292](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/auth_session.go#L292) — 下游已有统一刷新Cookie签发入口，需在此维护hint生命周期；代码：func WriteRefreshCookie(c *gin.Context, rawToken string) {

#### commit-9f506dd7f905

- Title: refactor(logs): simplify LogOther projection and dedupe sensitive keys
- Feature/group: refactor(logs): simplify LogOther projection and dedupe sensitive keys
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/log_format_test.go`, `model/log_other.go`, `model/log_other_test.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/text_quota_test.go`
- Overlap with downstream changes: model/log_format_test.go, service/task_billing.go, service/task_billing_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: LogOther投影简化与敏感键复用
- AI recommendation: `defer` — 该项主要整理已类型化LogOther的序列化/投影，当前没有该写入类型；若33只移植读取侧保护，不必先搬写入层再做此重构。可将真实有价值的无变化返回原JSON、敏感键一致性作为33实现原则，不照搬整个提交。
- AI risks: 不可把写入完整JSON误当已脱敏API结果；测试不能依赖不存在的上游私有类型。
- AI dependencies: commit-057f71c2336c
- AI suggested paths (not approval):
- AI evidence: [upstream:model/log_other.go:19](https://github.com/QuantumNous/new-api/blob/9f506dd7f905c288b4a119a8197cd64b77eb4a3f/model/log_other.go#L19) — 统一历史敏感键定义；代码：var legacySensitiveLogOtherKeys = []string{

#### commit-9df450fe54e1

- Title: feat(task): give polling hooks a real query context, host HTTP classification, and bounded poll failures
- Feature/group: feat(task): give polling hooks a real query context, host HTTP classification, and bounded poll failures
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.env.example`, `common/init.go`, `constant/env.go`, `controller/plugin_protocol_test.go`, `controller/relay.go`, `controller/task_plugin_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `model/log_other.go`, `model/task.go`, `model/task_cas_test.go`, `pkg/jsplugin/cli.go`, `pkg/jsplugin/cli_test.go`, `plugins/hailuo_responses_test.go`, `plugins/jimeng_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/google/plugin.js`, `plugins/tasks/hailuo/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `plugins/tasks/vertex-ai/plugin.js`, `plugins/tasks/vidu/plugin.js`, `plugins/veo_poll_test.go`, `relay/channel/adapter.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/common/relay_info.go`, `relay/relay_task.go`, `service/task_billing_test.go`, `service/task_plugin_view_test.go`, `service/task_polling.go`, `service/task_polling_test.go`
- Overlap with downstream changes: .env.example, controller/relay.go, model/task.go, model/task_cas_test.go, relay/common/relay_info.go, relay/relay_task.go, service/task_billing_test.go, service/task_polling.go, service/task_polling_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 任务轮询失败上限与查询上下文
- AI recommendation: `defer` — 插件轮询获得真实任务上下文，统一 HTTP 分类并增加连续失败阈值及超时退款。现有下游轮询还处理请求快照归档和通知，临时 429/网络抖动达到阈值即失败退款是运营策略变化，应先确认终态和上游仍运行时的处理。
- AI risks: 错误退款会造成免费交付或重复结算；需逐状态测试、故障恢复及人工补偿边界。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:common/init.go:206](https://github.com/QuantumNous/new-api/blob/9df450fe54e1a874a5339b7c38a61014217f02c3/common/init.go#L206) — 新增连续轮询失败阈值配置；代码：constant.TaskPollMaxFailures = GetEnvOrDefault("TASK_POLL_MAX_FAILURES", 20)
- AI evidence: [downstream:service/task_polling.go:29](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/task_polling.go#L29) — 现有轮询服务有独立 Go 任务编排，不能整体替换；代码：type TaskPollingAdaptor interface {

#### commit-36dbbf0f77e7

- Title: fix: keep ETag valid across different JSON packages
- Feature/group: fix: keep ETag valid across different JSON packages
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/etag.go`, `controller/misc.go`, `controller/revalidated_response.go`
- Overlap with downstream changes: controller/misc.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 稳定公开内容ETag
- AI recommendation: `migrate` — 建议与219c9e0634一起迁入最终版本：弱ETag由命名空间与原内容生成，不受JSON编码器或map字段顺序差异影响。当前公开接口仍直接返回内容，需要把HTTP包络、gzip Vary及304回包语义一起测试。
- AI risks: ETag只适用于公共内容，不能复用到用户/管理员配置；更改包络应换命名空间。
- AI dependencies: commit-219c9e06341f
- AI suggested paths (not approval): common/etag.go, controller/misc.go, controller/revalidated_response.go, common/etag_test.go, controller/revalidated_response_test.go
- AI evidence: [upstream:common/etag.go:21](https://github.com/QuantumNous/new-api/blob/36dbbf0f77e710455e745048f4a32e8120ad3fd2/common/etag.go#L21) — ETag稳定基于命名空间与原内容；代码：func ETagFor(namespace, content string) string {
- AI evidence: [downstream:controller/misc.go:188](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/misc.go#L188) — 下游公告接口仍直接回包，没有公共ETag辅助入口；代码："data":    common.OptionMap["Notice"],

#### commit-8f5ab8e4048a

- Title: fix(ci): resolve release version from trigger tag
- Feature/group: fix(ci): resolve release version from trigger tag
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.github/workflows/release.yml`
- Overlap with downstream changes: .github/workflows/release.yml
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游 release 按触发 tag 定版本
- AI recommendation: `skip` — 修复的是上游 release.yml 的 git describe 取版本，而下游已移除该发布工作流并使用自己的 VERSION/镜像流程。不存在可直接采用的同一发布入口。
- AI risks: 不得重建上游发布流程或把目标上游tag当下游已升级版本，CI文件属于保护范围。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:.github/workflows/release.yml:28](https://github.com/QuantumNous/new-api/blob/8f5ab8e4048a90d88b20ae1e6d5228b04233d3b8/.github/workflows/release.yml#L28) — 固定源码：VERSION=${GITHUB_REF#refs/tags/}
- AI evidence: [downstream:.github/workflows/build.yml:146](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/.github/workflows/build.yml#L146) — 下游对照：base_version=$(cat VERSION)

#### commit-32c261923a97

- Title: fix(task): explain 503 when a plugin-claimed model has no channel
- Feature/group: fix(task): explain 503 when a plugin-claimed model has no channel
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `middleware/distributor.go`, `middleware/distributor_test.go`
- Overlap with downstream changes: i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, middleware/distributor.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件无渠道诊断
- AI recommendation: `defer` — 错误提示解释模型被任务插件认领但没有可用渠道，依赖插件路由层；后续 7209b6db95 又隐藏内部插件标识，应采用目标版本最终脱敏行为而非本中间态。
- AI risks: 早期提示暴露内部插件键及管理操作信息。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:middleware/distributor.go:217](https://github.com/QuantumNous/new-api/blob/32c261923a9786c64d2af087327ef057e7bde7e3/middleware/distributor.go#L217) — 分流错误使用插件专属消息；代码：return i18n.T(c, i18n.MsgDistributorNoAvailableChannelTaskPlugin, map[string]any{"Group": group, "Model": modelName, "Plugin": pinned.Plugin.Meta.Key})

#### commit-3a9f41ee85cc

- Title: fix: temp disable /messages/count_tokens
- Feature/group: fix: temp disable /messages/count_tokens
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `router/relay-router.go`, `router/relay_router_test.go`
- Overlap with downstream changes: router/relay-router.go, router/relay_router_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude count_tokens 临时路由禁用
- AI recommendation: `skip` — 这是回退7137新加的/messages/count_tokens路由；本地基线没有注册该公共工具端点，无需先引入再禁用。日后若设计官方计数能力应独立决定。
- AI risks: 不能把关闭同名公共路由误当成下游Moonshot兼容入口能力的授权变更。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:router/relay-router.go:89](https://github.com/QuantumNous/new-api/blob/3a9f41ee85cc369f5b8d7fe6e62ff4e7bf3a9ec8/router/relay-router.go#L89) — 上游只注释临时新增路由。
- AI evidence: [downstream:router/relay-router.go:123](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/relay-router.go#L123) — 本地Claude路由直接注册/messages，没有count_tokens。

#### commit-7c044d7c5c2d

- Title: feat(relay): explicit @ model modifiers and canonical billing identity
- Feature/group: feat(relay): explicit @ model modifiers and canonical billing identity
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel-test.go`, `controller/model.go`, `middleware/distributor.go`, `middleware/distributor_test.go`, `model/channel_cache.go`, `model/channel_satisfy.go`, `relay/channel/claude/adaptor_test.go`, `relay/channel/claude/relay_claude_test.go`, `relay/channel/openai/adaptor.go`, `relay/claude_handler.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relay/compatible_handler.go`, `relay/convert_request_error_test.go`, `relay/gemini_handler.go`, `relay/helper/model_mapped.go`, `relay/helper/model_modifier.go`, `relay/helper/price.go`, `relay/helper/price_test.go`, `relay/helper/reasoning_suffix.go`, `relay/helper/reasoning_suffix_test.go`, `relay/responses_handler.go`, `relaykit/relayconvert/internal/convdiag/collector.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/suffix.go`, `relaykit/relayconvert/reasoning/suffix_test.go`, `relaykit/relayconvert/request_compat.go`, `relaykit/relayconvert/request_registry.go`, `setting/model_setting/global.go`, `setting/model_setting/global_test.go`, `setting/ratio_setting/matching_test.go`, `setting/ratio_setting/model_ratio.go`, `setting/reasoning/suffix.go`, `setting/reasoning/suffix_test.go`, `web/src/features/system-settings/models/global-settings-card.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/channel-test.go, controller/model.go, middleware/distributor.go, model/channel_cache.go, relay/channel/openai/adaptor.go, relay/claude_handler.go, relay/common/relay_info.go, relay/compatible_handler.go, relay/helper/model_mapped.go, relay/responses_handler.go, setting/ratio_setting/model_ratio.go, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: True
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 显式模型修饰符与计费身份
- AI recommendation: `defer` — 本地仅按UPSTREAM记录迁入本提交的Kimi真实模型名保护；当前仍用FormatMatchingModelName及原有计费模型身份。@修饰符、路由身份和计费键是未迁入的大改，不能标记整个提交covered。
- AI risks: 必须保持公开别名、映射后模型、租户限制、OpenRouter Kimi真实thinking模型和财务成本模型一致；缺完整路由计费回归。
- AI dependencies: commit-0ed497f066a6
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/model.go:254](https://github.com/QuantumNous/new-api/blob/7c044d7c5c2d2beadf16b21910950f8f593bc3ef/controller/model.go#L254) — 上游改为独立路由匹配身份。
- AI evidence: [downstream:controller/model.go:242](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/model.go#L242) — 本地仍使用现有匹配模型名。
- AI evidence: [downstream:UPSTREAM.md:19](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/UPSTREAM.md#L19) — 记录明确仅参考Kimi保护，未引入模型修饰符和计费身份重构。

#### commit-6b659fd61c50

- Title: fix(relay): preserve reasoning effort without implicit remapping
- Feature/group: fix(relay): preserve reasoning effort without implicit remapping
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/deepseek/adaptor.go`, `relay/channel/openai/adaptor.go`, `relay/channel/volcengine/adaptor.go`, `relay/channel/xai/adaptor.go`, `relay/common/relay_info.go`, `relay/helper/model_modifier.go`, `relay/helper/reasoning_suffix.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/intent.go`, `relaykit/relayconvert/reasoning/intent_test.go`
- Overlap with downstream changes: relay/channel/openai/adaptor.go, relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 推理力度保真与后缀保留
- AI recommendation: `defer` — 此修复建立在统一reasoning Intent/@修饰符之后，增加不隐式重映射和多供应商后缀保护。本地OpenAI仍有原有后缀与OpenRouter方言处理，单独cherry-pick会混用两套推理状态。
- AI risks: 未经分组验证会改变显式none、嵌套reasoning、模型映射和原生DeepSeek/火山/XAI参数语义。
- AI dependencies: commit-0ed497f066a6, commit-7c044d7c5c2d
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/openai/adaptor.go:259](https://github.com/QuantumNous/new-api/blob/6b659fd61c50e35d559c41520a0fff7b8aea56a4/relay/channel/openai/adaptor.go#L259) — 上游区分原生直通与需要渲染的推理意图。
- AI evidence: [downstream:relay/channel/openai/adaptor.go:257](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/openai/adaptor.go#L257) — 本地保留Kimi名称的实现仍是旧OpenRouter分支。

#### commit-d5803532bdcc

- Title: docs: require expression pricing and consolidated tests
- Feature/group: docs: require expression pricing and consolidated tests
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`
- Overlap with downstream changes: AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 表达式定价与集中测试治理规则
- AI recommendation: `skip` — 仅向上游 AGENTS 增加内置模型表达式价格和小变更测试组织要求。本地下游计费安全/测试规则由当前治理控制，不通过同步覆盖。
- AI risks: 如要采用额外约束需维护者单独评审，不得顺带改动现存显式价格覆盖或计费语义。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:AGENTS.md:110](https://github.com/QuantumNous/new-api/blob/d5803532bdccde3a2b1583291f51e92d3519b1c6/AGENTS.md#L110) — 固定源码：**Built-in model pricing:** New built-in model prices MUST be defined as self-contained billing expressions in `setting/billing_setting/builtin_billing.go`, using real USD per million tokens. Do not add new built-in prices to the legacy model/completion/cache ratio tables. Preserve explicit administrator pricing overrides. Existing legacy prices are migrated only when explicitly requested. Verify published prices and cover applicable context-length thresholds and cache categories.
- AI evidence: [downstream:AGENTS.md:103](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/AGENTS.md#L103) — 下游对照：**Billing expression system:** When working on tiered/dynamic billing (expression-based pricing), MUST read `pkg/billingexpr/expr.md` first. It documents the design philosophy, expression language, full architecture, token normalization rules, quota conversion, and expression versioning. All billing expression changes must follow that document.

#### commit-eb99ab1b4034

- Title: feat(billing): add built-in expression pricing for gpt-6-astra
- Feature/group: feat(billing): add built-in expression pricing for gpt-6-astra
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/option.go`, `relay/channel/openai/constant.go`, `setting/billing_setting/builtin_billing.go`, `setting/billing_setting/builtin_billing_test.go`, `setting/billing_setting/tiered_billing.go`
- Overlap with downstream changes: controller/option.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型内置表达式默认价格
- AI recommendation: `defer` — 新增内置表达式回落及 gpt-6-astra 默认售价，同时保持管理员显式配置优先。这会改变未配置模型的实际收费，本地没有该默认层；应先确认售价策略与模型能力，再以单独计费配置迁移。
- AI risks: 不能直接把上游供应商价格当本平台销售价格；模型名、长上下文、缓存创建及已有覆盖配置都需确认。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:setting/billing_setting/tiered_billing.go:53](https://github.com/QuantumNous/new-api/blob/eb99ab1b40343c3317bb47981cccdbb2b159a5fa/setting/billing_setting/tiered_billing.go#L53) — 内置默认表达式尊重已配置旧比例；代码：if ratio_setting.HasConfiguredModelRatio(model) {
- AI evidence: [downstream:setting/billing_setting/tiered_billing.go:38](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/setting/billing_setting/tiered_billing.go#L38) — 本地计费模式主要来自现有显式配置；代码：func GetBillingMode(model string) string {

#### commit-2cf177ac487e

- Title: perf(common): 批量复制 RawMessage，优化请求深拷贝 (#7221)
- Feature/group: perf(common): 批量复制 RawMessage，优化请求深拷贝 (#7221)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/copy.go`, `relay/request_clone_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: RawMessage 请求深拷贝
- AI recommendation: `merge` — 本地DeepCopy依赖逐元素反射；上游对json.RawMessage使用bytes.Clone，保持副本独立并增加请求重试隔离测试。可优先完整采用这项不改变协议内容的优化。
- AI risks: 需证明嵌套RawMessage、空值、显式零及源/克隆后续修改均不相互污染。
- AI dependencies: none identified
- AI suggested paths (not approval): common/copy.go, relay/request_clone_test.go
- AI evidence: [upstream:common/copy.go:24](https://github.com/QuantumNous/new-api/blob/2cf177ac487e62c627c7d423b65735ba2481ef4f/common/copy.go#L24) — 按字节批量复制且保留独立存储。
- AI evidence: [downstream:common/copy.go:14](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/common/copy.go#L14) — 本地仍为默认深拷贝。

#### commit-49ec46966825

- Title: fix(relay): apply model-specific OpenAI chat capabilities (#7211)
- Feature/group: fix(relay): apply model-specific OpenAI chat capabilities (#7211)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel_test_request_test.go`, `relay/channel/openai/adaptor.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_request_zero_value_test.go`
- Overlap with downstream changes: relay/channel/openai/adaptor.go, relaykit/dto/openai_request.go, relaykit/dto/openai_request_zero_value_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: OpenAI 按模型能力保留参数
- AI recommendation: `migrate` — 下游目前对所有GPT-5统一清空temperature/top_p/logprobs；上游按映射后模型和有效effort分别决定，并测试显式零和developer角色。建议抽取能力表和适配，避免连带迁入完整reasoning重构。
- AI risks: 需与后续GPT-6 Sol/Luna补丁合并设计；不能凭未来前缀套用限制，保留Moonshot/KVV映射和nil/zero语义。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/dto/openai_request.go, relaykit/dto/openai_request_zero_value_test.go, relay/channel/openai/adaptor.go, controller/channel_test_request_test.go
- AI evidence: [upstream:relaykit/dto/openai_request.go:253](https://github.com/QuantumNous/new-api/blob/49ec4696682530781a036eab1ac195f0b04706c0/relaykit/dto/openai_request.go#L253) — 上游将能力规则分离到DTO层。
- AI evidence: [downstream:relay/channel/openai/adaptor.go:343](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/openai/adaptor.go#L343) — 本地仍统一清空GPT-5采样字段。

#### commit-d8cb177440ce

- Title: feat(security): add access token management and audit logs
- Feature/group: feat(security): add access token management and audit logs
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/audit.go`, `controller/billing_option_test.go`, `controller/channel_task_plugin_bind_test.go`, `controller/channel_test_internal_test.go`, `controller/model_list_test.go`, `controller/relay.go`, `controller/secure_verification.go`, `controller/subscription.go`, `controller/twofa.go`, `controller/user.go`, `controller/user_manage_test.go`, `go.mod`, `go.sum`, `middleware/audit.go`, `middleware/auth.go`, `middleware/auth_test.go`, `model/audit_log.go`, `model/audit_other.go`, `model/log.go`, `model/main.go`, `model/user.go`, `router/api-router.go`, `router/main.go`, `router/plugin_router_test.go`, `router/web-router.go`, `service/authz/authz_test.go`, `service/authz/resources_audit.go`, `web/src/components/layout/components/mobile-drawer.tsx`, `web/src/components/profile-dropdown.tsx`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/profile/__tests__/settings.test.tsx`, `web/src/features/profile/api.ts`, `web/src/features/profile/components/__tests__/login-session-utils.test.ts`, `web/src/features/profile/components/dialogs/access-token-dialog.tsx`, `web/src/features/profile/components/dialogs/change-password-dialog.tsx`, `web/src/features/profile/components/dialogs/delete-account-dialog.tsx`, `web/src/features/profile/components/dialogs/email-bind-dialog.tsx`, `web/src/features/profile/components/dialogs/telegram-bind-dialog.tsx`, `web/src/features/profile/components/dialogs/two-fa-backup-dialog.tsx`, `web/src/features/profile/components/dialogs/two-fa-disable-dialog.tsx`, `web/src/features/profile/components/dialogs/two-fa-setup-dialog.tsx`, `web/src/features/profile/components/dialogs/wechat-bind-dialog.tsx`, `web/src/features/profile/components/login-session-dialogs.tsx`, `web/src/features/profile/components/login-session-item.tsx`, `web/src/features/profile/components/login-session-utils.ts`, `web/src/features/profile/components/login-sessions-card.tsx`, `web/src/features/profile/components/passkey-card.tsx`, `web/src/features/profile/components/profile-security-card.tsx`, `web/src/features/profile/components/profile-settings-card.tsx`, `web/src/features/profile/components/sidebar-modules-card.tsx`, `web/src/features/profile/components/tabs/account-bindings-tab.tsx`, `web/src/features/profile/components/tabs/notification-tab.tsx`, `web/src/features/profile/components/two-fa-card.tsx`, `web/src/features/profile/hooks/index.ts`, `web/src/features/profile/hooks/use-access-token.ts`, `web/src/features/profile/hooks/use-two-fa.ts`, `web/src/features/profile/index.tsx`, `web/src/features/profile/lib/user-settings.ts`, `web/src/features/profile/types.ts`, `web/src/features/security/__tests__/page.test.tsx`, `web/src/features/security/api.ts`, `web/src/features/security/components/__tests__/access-token-card.test.tsx`, `web/src/features/security/components/__tests__/login-session-utils.test.ts`, `web/src/features/security/components/__tests__/privacy-card.test.tsx`, `web/src/features/security/components/access-token-card.tsx`, `web/src/features/security/components/account-action-card.tsx`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/security/components/dialogs/access-token-dialog.tsx`, `web/src/features/security/components/dialogs/change-password-dialog.tsx`, `web/src/features/security/components/dialogs/delete-account-dialog.tsx`, `web/src/features/security/components/dialogs/email-bind-dialog.tsx`, `web/src/features/security/components/dialogs/telegram-bind-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-backup-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-disable-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-setup-dialog.tsx`, `web/src/features/security/components/dialogs/wechat-bind-dialog.tsx`, `web/src/features/security/components/login-session-dialogs.tsx`, `web/src/features/security/components/login-session-item.tsx`, `web/src/features/security/components/login-session-utils.ts`, `web/src/features/security/components/login-sessions-card.tsx`, `web/src/features/security/components/passkey-card.tsx`, `web/src/features/security/components/privacy-card.tsx`, `web/src/features/security/components/two-fa-card.tsx`, `web/src/features/security/hooks/use-access-token.ts`, `web/src/features/security/hooks/use-two-fa.ts`, `web/src/features/security/index.tsx`, `web/src/features/system-settings/maintenance/config.ts`, `web/src/features/system-settings/maintenance/sidebar-modules-section.tsx`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/__tests__/viewer.test.tsx`, `web/src/features/usage-logs/audit/api.ts`, `web/src/features/usage-logs/audit/components/audit-detail-fields.tsx`, `web/src/features/usage-logs/audit/components/audit-detail-value.tsx`, `web/src/features/usage-logs/audit/components/audit-log-columns.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/audit/components/audit-log-filter-bar.tsx`, `web/src/features/usage-logs/audit/components/audit-log-viewer.tsx`, `web/src/features/usage-logs/audit/index.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/components/__tests__/log-type-filter.test.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`, `web/src/features/usage-logs/components/compact-date-time-range-picker.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/components/dialogs/log-detail-layout.tsx`, `web/src/features/usage-logs/constants.ts`, `web/src/features/usage-logs/index.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/users/components/__tests__/permissions.test.tsx`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/hooks/__tests__/sidebar-config.test.tsx`, `web/src/hooks/use-sidebar-config.ts`, `web/src/hooks/use-sidebar-data.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/__tests__/http-cache.test.ts`, `web/src/lib/admin-permissions.ts`, `web/src/lib/auth-session.ts`, `web/src/lib/http-client.ts`, `web/src/routeTree.gen.ts`, `web/src/routes/_authenticated/security/index.tsx`, `web/src/routes/_authenticated/usage-logs/audit.tsx`
- Overlap with downstream changes: controller/audit.go, controller/model_list_test.go, controller/relay.go, controller/secure_verification.go, controller/user.go, controller/user_manage_test.go, go.mod, go.sum, model/log.go, model/main.go, model/user.go, router/api-router.go, router/web-router.go, service/authz/authz_test.go, web/src/components/profile-dropdown.tsx, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/profile/components/__tests__/login-session-utils.test.ts, web/src/features/profile/components/login-session-utils.ts, web/src/features/profile/components/tabs/account-bindings-tab.tsx, web/src/features/profile/index.tsx, web/src/features/system-settings/maintenance/config.ts, web/src/features/system-settings/maintenance/sidebar-modules-section.tsx, web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/features/usage-logs/components/compact-date-time-range-picker.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/constants.ts, web/src/features/usage-logs/index.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/users/components/users-mutate-drawer.tsx, web/src/hooks/use-sidebar-config.ts, web/src/hooks/use-sidebar-data.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/admin-permissions.ts, web/src/lib/http-client.ts, web/src/routeTree.gen.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 访问令牌生命周期及结构化审计
- AI recommendation: `defer` — 一次变更新增访问令牌管理、审计模型和大量安全页重组；本地控制台 JWT/refresh 与旧 access_token 并存，且审计含下游财务与身份记录。建议拆出认证/审计迁移方案，不能按整提交覆盖。
- AI risks: 日志隐私、审计分类、管理员权限矩阵、历史数据和前端个人中心迁移均需兼容。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/access_token.go:19](https://github.com/QuantumNous/new-api/blob/d8cb177440ceaae422d5bfd96c258d47af4e0f1d/controller/access_token.go#L19) — 固定源码：func RevokeAccessToken(c *gin.Context) {
- AI evidence: [downstream:middleware/auth.go:194](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/middleware/auth.go#L194) — 下游对照：func setDashboardAuthContext(c *gin.Context, user *model.UserBase, identity service.AuthIdentity, useAccessToken bool) {

#### commit-9a8674425c5a

- Title: fix(db): avoid redundant schema migrations on restart
- Feature/group: fix(db): avoid redundant schema migrations on restart
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `go.mod`, `go.sum`, `model/main.go`, `model/migration_dialector.go`, `model/migration_dialector_test.go`, `model/user_session_migration_test.go`
- Overlap with downstream changes: AGENTS.md, go.mod, go.sum, model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 数据库重复启动避免无效DDL
- AI recommendation: `migrate` — 建议按实际字段比较等价MySQL decimal默认值与PG CHAR元数据，并升级SQLite GORM驱动；和13/18/20的版本组一起验证。迁移Dialector必须组合现有observedMySQL/Postgres，不能替换真实连接观测或删除下游迁移。
- AI risks: 只忽略等价元数据差异，真实类型/长度/null变化仍应迁移；三数据库旧库二次启动必须无无效DDL并保留唯一性。
- AI dependencies: commit-66031a09d99f, commit-6eb6f35ed211, commit-74158715cde6, commit-2b6f1dfefbe2, commit-27ff6a8767e7
- AI suggested paths (not approval): go.mod, go.sum, model/main.go, model/migration_dialector.go, model/migration_dialector_test.go, model/user_session_migration_test.go, model/runtime_traffic.go, model/runtime_traffic_test.go
- AI evidence: [upstream:model/migration_dialector.go:23](https://github.com/QuantumNous/new-api/blob/9a8674425c5a43435a259b58bb928a55d26be990/model/migration_dialector.go#L23) — 定制迁移只归一等价元数据比较；代码：func (m mysqlSchemaMigrator) MigrateColumn(value any, field *schema.Field, column gorm.ColumnType) error {
- AI evidence: [downstream:model/runtime_traffic.go:72](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/runtime_traffic.go#L72) — 下游PG通过观测DialFunc统计真实连接字节，需保留；代码：config.DialFunc = func(ctx context.Context, network, address string) (net.Conn, error) {

#### commit-45c3fbe8aeb0

- Title: fix(security): bind verification proofs to sessions and actions
- Feature/group: fix(security): bind verification proofs to sessions and actions
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `common/json.go`, `common/json_test.go`, `controller/channel.go`, `controller/oauth.go`, `controller/passkey.go`, `controller/passkey_test.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `controller/twofa.go`, `middleware/auth_test.go`, `middleware/logger.go`, `middleware/secure_verification.go`, `model/auth_flow.go`, `model/passkey.go`, `model/twofa.go`, `model/twofa_enrollment.go`, `model/user_authentication_test.go`, `oauth/discord.go`, `oauth/generic.go`, `oauth/github.go`, `oauth/linuxdo.go`, `oauth/oidc.go`, `router/api-router.go`, `service/auth_token.go`, `service/auth_token_test.go`, `service/oauth_verification.go`, `service/passkey/session.go`, `service/security_verification.go`, `service/twofa.go`, `web/AGENTS.md`, `web/src/features/auth/api.ts`, `web/src/features/auth/components/oauth-callback-screen.tsx`, `web/src/features/auth/constants.ts`, `web/src/features/auth/lib/__tests__/oauth-callback-mode.test.ts`, `web/src/features/auth/lib/__tests__/oauth-popup.test.ts`, `web/src/features/auth/lib/oauth-callback-mode.ts`, `web/src/features/auth/lib/oauth-popup.ts`, `web/src/features/auth/passkey/api.ts`, `web/src/features/auth/passkey/hooks/use-passkey-management.ts`, `web/src/features/auth/secure-verification/__tests__/verification-api.test.ts`, `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/api.ts`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/hooks/__tests__/channel-key-disclosure.test.tsx`, `web/src/features/channels/hooks/use-channel-key-disclosure.ts`, `web/src/features/profile/types.ts`, `web/src/features/security/__tests__/enrollment.test.tsx`, `web/src/features/security/api.ts`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/security/components/dialogs/two-fa-setup-dialog.tsx`, `web/src/features/security/components/passkey-card.tsx`, `web/src/features/security/components/two-fa-card.tsx`, `web/src/features/security/hooks/use-two-fa-setup.ts`, `web/src/features/security/hooks/use-two-fa.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/api.ts`, `web/src/lib/http-client.ts`, `web/src/lib/oauth.ts`, `web/src/lib/passkey.ts`, `web/src/lib/secure-verification.ts`, `web/src/lib/server-error-message.ts`, `web/src/routes/oauth/$provider.tsx`
- Overlap with downstream changes: AGENTS.md, controller/channel.go, controller/secure_verification.go, model/auth_flow.go, model/twofa.go, model/user_authentication_test.go, router/api-router.go, service/auth_token.go, web/src/features/auth/constants.ts, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/auth/secure-verification/types.ts, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/http-client.ts, web/src/lib/server-error-message.ts, web/src/routes/oauth/$provider.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 按会话与动作内容绑定安全证明
- AI recommendation: `defer` — 上游新增 VerificationOperation/ContextHash 并重构 OAuth、Passkey、2FA enrollment；本地证明已绑定会话/身份/方法/scope，但下游手机、微信 miniapp 流程独立扩展。建议优先设计完整的动作绑定迁移，而非照搬治理文件和整套安全页。
- AI risks: 必须保持一次消费与业务事务原子性，防跨目标、跨动作重放；受保护 AGENTS.md 不进入适配范围。
- AI dependencies: commit-d8cb177440ce
- AI suggested paths (not approval):
- AI evidence: [upstream:service/security_verification.go:52](https://github.com/QuantumNous/new-api/blob/45c3fbe8aeb049f03c13e14298a40b87aea5bd87/service/security_verification.go#L52) — 固定源码：ContextHash string `json:"context_hash"`
- AI evidence: [downstream:service/auth_token.go:160](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/auth_token.go#L160) — 下游对照：func VerifySecurityProof(raw string, identity AuthIdentity, requiredScope string, allowedMethods []string) (string, error) {

#### commit-3e84ec0ab823

- Title: feat(auth): migrate Telegram to unified OAuth
- Feature/group: feat(auth): migrate Telegram to unified OAuth
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel.go`, `controller/misc.go`, `controller/oauth.go`, `controller/option.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `controller/telegram.go`, `controller/telegram_test.go`, `controller/twofa.go`, `go.mod`, `go.sum`, `model/custom_oauth_provider.go`, `model/external_identity_claim.go`, `oauth/registry.go`, `oauth/telegram.go`, `oauth/types.go`, `router/api-router.go`, `service/security_verification.go`, `setting/system_setting/telegram.go`, `web/src/features/auth/api.test.ts`, `web/src/features/auth/api.ts`, `web/src/features/auth/components/oauth-providers.tsx`, `web/src/features/auth/hooks/use-oauth-login.ts`, `web/src/features/auth/lib/__tests__/oauth-popup.test.ts`, `web/src/features/auth/lib/oauth-callback-mode.ts`, `web/src/features/auth/lib/oauth-popup.ts`, `web/src/features/auth/secure-verification/__tests__/verification-api.test.ts`, `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/api.ts`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/auth/types.ts`, `web/src/features/security/__tests__/enrollment.test.tsx`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/system-settings/auth/index.tsx`, `web/src/features/system-settings/auth/oauth-section.tsx`, `web/src/features/system-settings/auth/section-registry.tsx`, `web/src/features/system-settings/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/server-error-message.ts`, `web/src/routes/oauth/$provider.tsx`
- Overlap with downstream changes: controller/channel.go, controller/misc.go, controller/option.go, controller/secure_verification.go, go.mod, go.sum, model/external_identity_claim.go, router/api-router.go, web/src/features/auth/components/oauth-providers.tsx, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/auth/secure-verification/types.ts, web/src/features/auth/types.ts, web/src/features/system-settings/auth/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/server-error-message.ts, web/src/routes/oauth/$provider.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Telegram 统一 OAuth 流程
- AI recommendation: `defer` — 从旧 Telegram assertion 接口转为统一 OAuth flow、配置验证和会话身份绑定；本地仍有独立 Telegram 登录/绑定，并有微信 miniapp 的独立用途。需有旧客户端迁移、回调和配置计划后再选。
- AI risks: 不可把 Telegram、微信网页登录与微信小程序 flow purpose 合并；旧 Bot Token 配置不等同新 OAuth 配置。
- AI dependencies: commit-45c3fbe8aeb0
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/oauth.go:71](https://github.com/QuantumNous/new-api/blob/3e84ec0ab8239cf2277f8a10d45566630a9c10fa/controller/oauth.go#L71) — 固定源码：flowPayload.Telegram = telegramFlow
- AI evidence: [downstream:model/auth_flow.go:24](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/auth_flow.go#L24) — 下游对照：AuthFlowPurposeTelegramAssertion = "telegram_assertion"

#### commit-a8729b5c3709

- Title: feat(security): require verification for access token management
- Feature/group: feat(security): require verification for access token management
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/security_enrollment_test.go`, `controller/user.go`, `router/api-router.go`, `service/security_verification.go`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/security/api.ts`, `web/src/features/security/components/__tests__/access-token-card.test.tsx`, `web/src/features/security/components/access-token-card.tsx`, `web/src/features/security/hooks/use-access-token.ts`, `web/src/lib/secure-verification.ts`
- Overlap with downstream changes: controller/user.go, router/api-router.go, web/src/features/auth/secure-verification/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 访问令牌签发和撤销二次验证
- AI recommendation: `defer` — 该补丁依赖新的操作绑定安全证明与访问令牌控制器；本地 GenerateAccessToken 仍在 user.go。建议与 PAT 安全系列共同迁移，不能只改前端弹窗。
- AI risks: 令牌签发/撤销需要绑定准确的操作上下文、限流和审计，保留浏览器与小程序会话隔离。
- AI dependencies: commit-d8cb177440ce, commit-45c3fbe8aeb0
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/access_token.go:25](https://github.com/QuantumNous/new-api/blob/a8729b5c3709cc01d88fc3f2db5b91347fc9129e/controller/access_token.go#L25) — 固定源码：if middleware.RequireSecurityProof(c, service.VerificationOperation{Scope: service.VerificationScopeAccessTokenGenerate}) == nil {
- AI evidence: [downstream:controller/user.go:422](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/user.go#L422) — 下游对照：func GenerateAccessToken(c *gin.Context) {

#### commit-0973dc2b8f55

- Title: feat(security): harden account binding and password changes
- Feature/group: feat(security): harden account binding and password changes
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.env.example`, `common/account_password.go`, `common/crypto.go`, `common/password_crypto.go`, `controller/audit.go`, `controller/auth_flow_test.go`, `controller/custom_oauth.go`, `controller/email_binding.go`, `controller/misc.go`, `controller/model_list_test.go`, `controller/oauth.go`, `controller/passkey.go`, `controller/secure_verification.go`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `controller/setup.go`, `controller/telegram_test.go`, `controller/user.go`, `controller/wechat.go`, `middleware/secure_verification.go`, `model/account_security.go`, `model/auth_flow.go`, `model/email_binding.go`, `model/user.go`, `router/api-router.go`, `service/account_security.go`, `service/auth_session.go`, `service/email_binding.go`, `service/security_verification.go`, `web/src/features/auth/api.ts`, `web/src/features/auth/constants.ts`, `web/src/features/auth/lib/password-encryption.ts`, `web/src/features/auth/secure-verification/__tests__/verification-api.test.ts`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/auth/sign-up/components/sign-up-form.tsx`, `web/src/features/auth/types.ts`, `web/src/features/profile/api.ts`, `web/src/features/profile/types.ts`, `web/src/features/security/__tests__/account-security.test.tsx`, `web/src/features/security/__tests__/enrollment.test.tsx`, `web/src/features/security/components/account-action-card.tsx`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/security/components/dialogs/change-password-dialog.tsx`, `web/src/features/security/components/dialogs/email-bind-dialog.tsx`, `web/src/features/security/components/dialogs/wechat-bind-dialog.tsx`, `web/src/features/security/hooks/use-account-security.ts`, `web/src/features/security/index.tsx`, `web/src/features/setup/setup-wizard.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/password-policy.ts`, `web/src/stores/auth-store.ts`
- Overlap with downstream changes: .env.example, controller/audit.go, controller/misc.go, controller/model_list_test.go, controller/secure_verification.go, controller/user.go, model/auth_flow.go, model/user.go, router/api-router.go, service/auth_session.go, web/src/features/auth/constants.ts, web/src/features/auth/secure-verification/types.ts, web/src/features/auth/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/features/users/components/users-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/stores/auth-store.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 账号绑定、密码存储与长密码升级
- AI recommendation: `defer` — 新增 Argon2id、密码 envelope v2、邮箱绑定与账号安全事务；本地存在 verified phone、UnionID 和小程序账号合并边界，需分阶段双格式验证/哈希切换及回滚计划。
- AI risks: 滚动升级不能生成旧节点无法验证的密码；绑定解绑不可移除手机号所有权和会话失效规则。
- AI dependencies: commit-b80d633cf586, commit-918427d8ab41, commit-45c3fbe8aeb0, commit-3e84ec0ab823, commit-a8729b5c3709
- AI suggested paths (not approval):
- AI evidence: [upstream:common/crypto.go:31](https://github.com/QuantumNous/new-api/blob/0973dc2b8f550de71b75fdd3805576d3ce6ccf42/common/crypto.go#L31) — 固定源码：if strings.HasPrefix(hash, "$argon2id$") {
- AI evidence: [downstream:model/auth_flow.go:25](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/auth_flow.go#L25) — 下游对照：AuthFlowPurposeWeChatMiniAppBind = "wechat_miniapp_bind"

#### commit-3f8a50cf8877

- Title: feat(audit): complete token and quota operation records
- Feature/group: feat(audit): complete token and quota operation records
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `constant/context_key.go`, `controller/audit.go`, `controller/subscription.go`, `controller/token.go`, `controller/token_test.go`, `controller/user.go`, `controller/user_manage_test.go`, `controller/user_quota.go`, `middleware/audit.go`, `model/log.go`, `model/quota_reserve.go`, `model/user_quota_adjustment.go`, `router/api-router.go`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/__tests__/viewer.test.tsx`, `web/src/features/usage-logs/audit/components/audit-detail-fields.tsx`, `web/src/features/usage-logs/audit/components/audit-log-columns.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/components/__tests__/quota-adjustment.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/lib/quota-audit-operation.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: AGENTS.md, constant/context_key.go, controller/audit.go, controller/user.go, controller/user_manage_test.go, model/log.go, router/api-router.go, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: API key 操作与额度变更审计
- AI recommendation: `defer` — 新增 tokenAuditParams、独立额度调整与审计字段；本地额度调整接入钱包账本、财务纠错和冻结成本证据。需把审计事件适配到下游事务链，禁止替换现有账务路径。
- AI risks: 不能因审计重构重复扣费、丢失不可变账本或暴露 API key；AGENTS.md 变更另行治理。
- AI dependencies: commit-d8cb177440ce
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/user_quota.go:19](https://github.com/QuantumNous/new-api/blob/3f8a50cf8877683669cd812240a0beaf7b171c32/controller/user_quota.go#L19) — 固定源码："requested_quota": req.Value,
- AI evidence: [downstream:router/api-router.go:123](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/api-router.go#L123) — 下游对照：billingCorrection.POST("/preview", middleware.RootAuth(), middleware.CriticalRateLimit(), controller.PreviewBillingCorrection)

#### commit-521cebf585ef

- Title: fix(dashboard): simplify completed setup guide
- Feature/group: fix(dashboard): simplify completed setup guide
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/dashboard/components/overview/__tests__/setup-guide.test.tsx`, `web/src/features/dashboard/components/overview/overview-dashboard.tsx`, `web/src/features/dashboard/index.tsx`
- Overlap with downstream changes: web/src/features/dashboard/components/overview/overview-dashboard.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 完成初始化后简化首页引导
- AI recommendation: `defer` — 调整首页SectionPageLayout、完成后折叠入口与焦点恢复，是体验改版而非运行时兼容修复。下游首页已有新增财务/资源能力，应等维护者选择新布局后单独适配，不在数据库安全修复批次顺带覆盖。
- AI risks: 保留Root配置提示、用户权限和下游首页数据；焦点与折叠状态需可访问性验证。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/dashboard/components/overview/overview-dashboard.tsx:462](https://github.com/QuantumNous/new-api/blob/521cebf585efc2e782dd9fb93d0f66752c8d3c32/web/src/features/dashboard/components/overview/overview-dashboard.tsx#L462) — 完成引导的折叠入口加入焦点回退；代码：const setupGuideToggleRef = useRef<HTMLButtonElement>(null)

#### commit-6f2333990613

- Title: feat(auth): unify login verification and secure account deletion
- Feature/group: feat(auth): unify login verification and secure account deletion
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/audit.go`, `controller/auth_flow_test.go`, `controller/auth_session_test.go`, `controller/login_verification.go`, `controller/model_list_test.go`, `controller/passkey.go`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `controller/twofa.go`, `controller/user.go`, `model/account_security.go`, `model/auth_flow.go`, `model/login_verification.go`, `model/twofa.go`, `model/user.go`, `model/user_session.go`, `router/api-router.go`, `service/auth_session.go`, `service/login_verification.go`, `service/passkey/session.go`, `service/security_verification.go`, `web/src/features/auth/hooks/use-auth-redirect.ts`, `web/src/features/auth/hooks/use-oauth-login.ts`, `web/src/features/auth/otp/__tests__/login-verification.test.tsx`, `web/src/features/auth/otp/components/otp-form.tsx`, `web/src/features/auth/otp/index.tsx`, `web/src/features/auth/secure-verification/__tests__/verification-api.test.ts`, `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/api.ts`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/auth/sign-in/components/user-auth-form.tsx`, `web/src/features/auth/sign-up/components/sign-up-form.tsx`, `web/src/features/auth/types.ts`, `web/src/features/profile/api.ts`, `web/src/features/profile/types.ts`, `web/src/features/security/__tests__/account-security.test.tsx`, `web/src/features/security/components/dialogs/delete-account-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-backup-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-disable-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/api.ts`, `web/src/lib/auth-session.test.ts`, `web/src/lib/passkey.ts`, `web/src/routes/oauth/$provider.tsx`, `web/src/stores/auth-store.ts`
- Overlap with downstream changes: controller/audit.go, controller/model_list_test.go, controller/user.go, model/auth_flow.go, model/twofa.go, model/user.go, router/api-router.go, service/auth_session.go, web/src/features/auth/otp/components/otp-form.tsx, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/auth/secure-verification/types.ts, web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/auth/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/routes/oauth/$provider.tsx, web/src/stores/auth-store.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 统一登录挑战与安全删除账号
- AI recommendation: `defer` — 上游把 2FA/Passkey 登录挑战统一，并提高 Passkey UserVerification、账号删除验证要求；本地登录会签发自己的 session bundle 且有手机号/微信流程，需要一致的跨入口策略。
- AI risks: 删除账号前验证、会话撤销与数据留存要一并审计；不能让已有小程序会话绕过或全部失效。
- AI dependencies: commit-45c3fbe8aeb0, commit-0973dc2b8f55
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/passkey.go:81](https://github.com/QuantumNous/new-api/blob/6f2333990613bf3e9dd36f541fc380148c7b5175/controller/passkey.go#L81) — 固定源码：selection.UserVerification = protocol.VerificationRequired
- AI evidence: [downstream:service/auth_token.go:37](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/auth_token.go#L37) — 下游对照：type AuthIdentity struct {

#### commit-0c76e4dae77a

- Title: feat(models): rework model/vendor management and pricing
- Feature/group: feat(models): rework model/vendor management and pricing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/model_management_test.go`, `controller/model_meta.go`, `controller/model_pricing_config.go`, `controller/model_sync.go`, `controller/ratio_config.go`, `controller/ratio_sync.go`, `controller/ratio_sync_test.go`, `controller/vendor_meta.go`, `model/ability.go`, `model/channel_cache.go`, `model/model_meta.go`, `model/model_metadata_sync.go`, `model/model_pricing_config.go`, `model/option.go`, `model/pricing_default.go`, `model/vendor_management.go`, `model/vendor_meta.go`, `router/api-router.go`, `setting/billing_setting/tiered_billing.go`, `setting/ratio_setting/model_ratio.go`, `web/src/components/data-table/core/badge-list-cell.tsx`, `web/src/components/data-table/layout/data-table-page.tsx`, `web/src/components/data-table/layout/mobile-card-list.tsx`, `web/src/components/data-table/toolbar/bulk-actions.tsx`, `web/src/components/data-table/toolbar/view-mode-toggle.tsx`, `web/src/components/layout/components/section-page-layout.tsx`, `web/src/components/lobe-icon-field.tsx`, `web/src/components/ui/__tests__/combobox.test.tsx`, `web/src/components/ui/combobox-input.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx`, `web/src/features/channels/components/dialogs/channel-test-dialog.tsx`, `web/src/features/channels/components/dialogs/edit-tag-dialog.tsx`, `web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/model-pricing/__tests__/pricing.test.ts`, `web/src/features/model-pricing/api.ts`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/model-pricing/pricing.ts`, `web/src/features/models/__tests__/metadata-editing.test.tsx`, `web/src/features/models/__tests__/metadata-sync.test.tsx`, `web/src/features/models/__tests__/model-deletion.test.tsx`, `web/src/features/models/__tests__/vendor-management.test.tsx`, `web/src/features/models/api.ts`, `web/src/features/models/components/data-table-bulk-actions.tsx`, `web/src/features/models/components/data-table-row-actions.tsx`, `web/src/features/models/components/dialogs/create-deployment-drawer.tsx`, `web/src/features/models/components/dialogs/missing-models-dialog.tsx`, `web/src/features/models/components/dialogs/model-delete-dialog.tsx`, `web/src/features/models/components/dialogs/price-sync-dialog.tsx`, `web/src/features/models/components/dialogs/sync-wizard-dialog.tsx`, `web/src/features/models/components/dialogs/upstream-conflict-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-mutate-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-operation-dialog.tsx`, `web/src/features/models/components/dialogs/vendors-management-dialog.tsx`, `web/src/features/models/components/dialogs/view-logs-dialog.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/models/components/model-connections.tsx`, `web/src/features/models/components/models-columns.tsx`, `web/src/features/models/components/models-dialogs.tsx`, `web/src/features/models/components/models-primary-buttons.tsx`, `web/src/features/models/components/models-provider.tsx`, `web/src/features/models/components/models-table.tsx`, `web/src/features/models/components/vendor-linked-models.tsx`, `web/src/features/models/components/vendors-table.tsx`, `web/src/features/models/constants.ts`, `web/src/features/models/index.tsx`, `web/src/features/models/lib/model-actions.ts`, `web/src/features/models/section-registry.tsx`, `web/src/features/models/types.ts`, `web/src/features/models/vendor-api.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/pricing-controls.test.tsx`, `web/src/features/pricing/components/loading-skeleton.tsx`, `web/src/features/pricing/components/model-billing-mode-badge.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`, `web/src/features/pricing/components/pricing-sidebar.tsx`, `web/src/features/pricing/components/pricing-toolbar.tsx`, `web/src/features/pricing/index.tsx`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/profile/components/language-preferences-card.tsx`, `web/src/features/subscriptions/components/dialogs/subscription-purchase-dialog.tsx`, `web/src/features/subscriptions/components/dialogs/user-subscriptions-dialog.tsx`, `web/src/features/subscriptions/components/subscriptions-mutate-drawer.tsx`, `web/src/features/system-settings/__tests__/pricing-sync.test.tsx`, `web/src/features/system-settings/auth/custom-oauth/components/preset-selector.tsx`, `web/src/features/system-settings/integrations/waffo-pancake-settings-section.tsx`, `web/src/features/system-settings/models/channel-selector-dialog.tsx`, `web/src/features/system-settings/models/conflict-confirm-dialog.tsx`, `web/src/features/system-settings/models/group-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/group-special-usable-editor.tsx`, `web/src/features/system-settings/models/model-pricing-core.ts`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-pricing-snapshots.ts`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/ratio-settings-card.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/features/system-settings/models/upstream-price-cells.tsx`, `web/src/features/system-settings/models/upstream-ratio-sync-columns.tsx`, `web/src/features/system-settings/models/upstream-ratio-sync-helpers.ts`, `web/src/features/system-settings/models/upstream-ratio-sync-table.tsx`, `web/src/features/system-settings/models/upstream-ratio-sync.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/plugin-sandbox.tsx`, `web/src/features/usage-logs/audit/components/audit-log-filter-bar.tsx`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/lobe-icon.tsx`, `web/src/routes/_authenticated/models/$section.tsx`, `web/vitest.config.ts`
- Overlap with downstream changes: model/ability.go, model/channel_cache.go, model/option.go, router/api-router.go, setting/ratio_setting/model_ratio.go, web/src/features/channels/components/dialogs/channel-test-dialog.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/features/pricing/index.tsx, web/src/features/system-settings/types.ts, web/src/features/users/components/users-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型/供应商与价格管理重构
- AI recommendation: `defer` — 122文件重构统一模型元数据、供应商、价格快照及并发保存，并将插件usage_schema纳入价格配置。它有管理价值，但不是现有上游来源同步页面；须先确认与下游财务权限、固定档位、成本折扣和调账页面的关系。
- AI risks: 删除模型可能清理渠道和价格；所有删除、并发冲突及跨数据库事务需验证，不能损坏历史冻结账单。
- AI dependencies: commit-eb48396d5fe9, commit-eb99ab1b4034
- AI suggested paths (not approval):
- AI evidence: [upstream:model/model_pricing_config.go:46](https://github.com/QuantumNous/new-api/blob/0c76e4dae77a279e015329b7478e6f02d6b62edd/model/model_pricing_config.go#L46) — 新增带乐观版本冲突检测的价格管理模型；代码：var ErrModelPricingConflict = errors.New("model pricing changed; reload before saving")

#### commit-3b4652269a6a

- Title: feat(ali): support wan3.0 all-in-one video models
- Feature/group: feat(ali): support wan3.0 all-in-one video models
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/alibaba_wan3_test.go`, `plugins/tasks/alibaba/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Wan3 全能视频协议
- AI recommendation: `defer` — 上游为 Wan3 新增统一 input.media、adaptive 比例、分辨率和自动时长标记。下游已具有独立 Wan3 Go DTO 与映射后路由，但未证明所有协议/计费语义等价，应先做能力矩阵，再挑缺口迁移；不能标记整体 covered。
- AI risks: 自动时长负值不能直接参与计费；需保留显式零值、请求冻结、Wan2.x 旧行为。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/alibaba/plugin.js:82](https://github.com/QuantumNous/new-api/blob/3b4652269a6a6d9e2c8650a84ee8c4e447e6473b/plugins/tasks/alibaba/plugin.js#L82) — 上游把 -1 自动时长转为独立标记进入后续计价；代码：function normalizeAutoDuration(req) {
- AI evidence: [downstream:relay/channel/task/ali/adaptor.go:755](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/task/ali/adaptor.go#L755) — 下游已有独立 Wan3 请求类型，不能用 JS 文件覆盖；代码：type wan3VideoRequest struct {

#### commit-6e10f9bc927a

- Title: fix(relay): preserve Kimi K3 dynamic tool loading messages
- Feature/group: fix(relay): preserve Kimi K3 dynamic tool loading messages
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/moonshot/constants.go`, `relay/chat_completions_via_responses.go`, `relay/chat_completions_via_responses_test.go`, `relay/compatible_handler.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_request_zero_value_test.go`, `relaykit/relayconvert/kitutil/json_test.go`
- Overlap with downstream changes: relay/channel/moonshot/constants.go, relay/chat_completions_via_responses.go, relay/chat_completions_via_responses_test.go, relay/compatible_handler.go, relaykit/dto/openai_request.go, relaykit/dto/openai_request_zero_value_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: True
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Kimi K3 动态工具消息
- AI recommendation: `covered` — 已对照本地K3模型列表、messages[].tools序列化、无content动态消息、工具Token估算和系统提示注入逻辑；UPSTREAM已记录这项完整动态工具修复。保留现有Moonshot facade明确拒绝伪造动态加载的下游边界，不重复迁入。
- AI risks: covered仅指本提交动态工具修复；不表示7c044模型修饰符大重构或所有Kimi专属能力已覆盖。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/dto/openai_request.go:123](https://github.com/QuantumNous/new-api/blob/6e10f9bc927a4eae889864a6ef601359d53526b9/relaykit/dto/openai_request.go#L123) — 上游识别动态工具消息并省略其content。
- AI evidence: [downstream:relaykit/dto/openai_request.go:119](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/dto/openai_request.go#L119) — 本地已有相同动态消息专用序列化路径。
- AI evidence: [downstream:UPSTREAM.md:20](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/UPSTREAM.md#L20) — 已记录迁入范围、Token估算、系统提示与原生Moonshot/facade边界及验证。
- AI evidence: [downstream:relay/chat_completions_via_responses.go:37](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/chat_completions_via_responses.go#L37) — 系统提示注入跳过Kimi动态工具声明。
- AI evidence: [downstream:relay/facade/moonshot/facade.go:177](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/facade/moonshot/facade.go#L177) — 普通OpenAI数据源仍明确拒绝不可复现的动态工具加载。

#### commit-7bbe85bcb095

- Title: refactor(json): route JSON helpers through a host-injectable codec
- Feature/group: refactor(json): route JSON helpers through a host-injectable codec
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/json.go`, `common/json_test.go`, `relaykit/dto/values.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/kitutil/json.go`
- Overlap with downstream changes: relaykit/dto/values.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 宿主可注入 JSON 编解码器
- AI recommendation: `defer` — 上游让common与relaykit DTO统一走可注入Codec，本地common和独立relaykit仍直接使用标准库。属于跨模块初始化/序列化语义调整，不能只替换部分调用。
- AI risks: 要保留下游IntValue溢出/字符串解析加强、显式零、ETag稳定性，并证明relaykit无根模块依赖；须与所选JSON引擎升级一并验证。
- AI dependencies: commit-0ed497f066a6
- AI suggested paths (not approval):
- AI evidence: [upstream:common/json.go:18](https://github.com/QuantumNous/new-api/blob/7bbe85bcb09546e0b89572bf97198fc94889be3d/common/json.go#L18) — 宿主在初始化时注入统一编解码器。
- AI evidence: [downstream:common/json.go:22](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/common/json.go#L22) — 本地common JSON封装仍直接使用标准库。

#### commit-387a40914853

- Title: fix(web): keep drawer popups interactive and shim storage in tests
- Feature/group: fix(web): keep drawer popups interactive and shim storage in tests
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/ui/__tests__/portal-container.test.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/components/ui/drawer.tsx`, `web/src/components/ui/portal-container.ts`, `web/src/components/ui/select.tsx`, `web/src/test-setup.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 抽屉内 Select/Combobox 弹层交互
- AI recommendation: `migrate` — 上游将弹层 portal 放入抽屉内容容器，解决 modal 指针/焦点边界。本地 shared drawer/select 和 legacy combobox 组合仍不同，应迁入容器约束思想并测试真实嵌套，而非覆盖组件。
- AI risks: 焦点回收、点击选项、Escape、滚动裁切与普通页面弹层都需回归；测试 shim 不应伪造可访问性。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/components/ui/drawer.tsx, web/src/components/ui/select.tsx, web/src/components/ui/combobox.tsx, web/src/components/ui/combobox-input.tsx, web/src/components/ui/portal-container.ts, web/src/components/ui/__tests__/drawer-popups.test.tsx
- AI evidence: [upstream:web/src/components/ui/drawer.tsx:93](https://github.com/QuantumNous/new-api/blob/387a40914853310d69adc2f52474134ced5f4811/web/src/components/ui/drawer.tsx#L93) — 固定源码：<PortalContainerContext.Provider value={content}>
- AI evidence: [downstream:web/src/components/ui/combobox.tsx:173](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/components/ui/combobox.tsx#L173) — 下游对照：<ComboboxPrimitive.Portal>

#### commit-5c7cca015525

- Title: fix(perf): return hourly success-rate series for model health bar
- Feature/group: fix(perf): return hourly success-rate series for model health bar
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `pkg/perf_metrics/metrics.go`, `pkg/perf_metrics/types.go`, `web/src/features/performance-metrics/types.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型健康小时序列
- AI recommendation: `defer` — 把最近三个非空成功率改为带时间戳的小时序列并展示24槽。当前下游是刻意紧凑的三柱状态，接入会改变API契约与卡片布局，且后续1ca79d57e7重新定义近期样本；建议选最终方案整体迁移而非孤立改接口。
- AI risks: 必须区分无流量和失败，保留按用户可用组过滤；不能影响下游Root节点资源监控。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/perf_metrics/metrics.go:190](https://github.com/QuantumNous/new-api/blob/5c7cca015525212a7ac2741da7f3b51fa2e30db0/pkg/perf_metrics/metrics.go#L190) — API改为时间戳成功率序列；代码：RecentSuccessSeries: recentSuccessSeries(modelBuckets[name]),
- AI evidence: [downstream:web/src/features/pricing/components/model-perf-badge.tsx:70](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/pricing/components/model-perf-badge.tsx#L70) — 本地状态显示明确为三个柱位；代码：...Array(Math.max(0, 3 - statusRates.length)).fill(null),

#### commit-6298b0f32384

- Title: fix(plugin): suppress factory layer when disabling an overridden task plugin
- Feature/group: fix(plugin): suppress factory layer when disabling an overridden task plugin
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件禁用及内置回退
- AI recommendation: `defer` — 禁用覆盖插件时同时抑制内置 factory 层，防止禁用后继续接单。该行为应作为未来插件启停的必需安全语义；当前没有插件覆盖层，不宜先引入孤立控制器。
- AI risks: 禁用对已存在任务的轮询能力与新请求路由应分别确认。
- AI dependencies: commit-eb48396d5fe9, commit-dc4732cfed71
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/task_plugin.go:481](https://github.com/QuantumNous/new-api/blob/6298b0f3238461b9629dfc1c00866f8325123aa1/controller/task_plugin.go#L481) — 禁用会同时停止可服务此键的层；代码：// Switching a key off must silence every layer that can serve it. The

#### commit-92bc7ff73c5e

- Title: fix(plugins): make sunoapi alias-safe and lock alias echo across built-ins
- Feature/group: fix(plugins): make sunoapi alias-safe and lock alias echo across built-ins
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/builtin_plugins_test.go`, `plugins/sunoapi_responses_test.go`, `plugins/tasks/sunoapi/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Suno 插件别名及响应回显
- AI recommendation: `defer` — Suno 按解析后的真实模型区分歌词/音乐，响应保留公开别名，并只透传有效请求头；是插件别名链修复。下游 Suno Go 模式与 `/moonshot` 透传不同，应随插件模型身份统一纳入。
- AI risks: 公开别名不得变成上游模型或计费身份；空 Header 不应干扰服务端默认。
- AI dependencies: commit-eb48396d5fe9, commit-6c22550ea325
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/sunoapi/plugin.js:230](https://github.com/QuantumNous/new-api/blob/92bc7ff73c5ef215496d09d3e4b5769c9e35006b/plugins/tasks/sunoapi/plugin.js#L230) — 按已解析模型决定歌词模式；代码：const declared = trimmed(ctx.upstreamModel || ctx.model).toLowerCase();

#### commit-210734bb73bc

- Title: refactor(task): remove the custom-plugin layer switch
- Feature/group: refactor(task): remove the custom-plugin layer switch
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/init.go`, `constant/task.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `model/option.go`, `model/option_task_plugin_test.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_disabled_factory_test.go`, `pkg/jsplugin/registry_master_enabled_test.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`
- Overlap with downstream changes: constant/task.go, model/option.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件配置简化
- AI recommendation: `defer` — 删除自定义覆盖层的独立开关，保留总开关与内置禁用集。该提交是先前插件配置契约的调整，并非现有配置清理；应和最终版本插件策略一起选择。
- AI risks: 旧环境变量或已保存选项的升级语义需记录；不能意外启用自定义插件。
- AI dependencies: commit-eb48396d5fe9, commit-6298b0f32384
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/routing.go:848](https://github.com/QuantumNous/new-api/blob/210734bb73bc3c6e37548af90360aa4226566fe2/pkg/jsplugin/routing.go#L848) — 最终有效插件直接合并覆盖层；代码：maps.Copy(effective, override)

#### commit-bee45b58a3c0

- Title: fix(web): switch the pricing card grid to three columns at xl
- Feature/group: fix(web): switch the pricing card grid to three columns at xl
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/loading-skeleton.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 价格卡片三列断点
- AI recommendation: `skip` — 上游从 2xl 提前到 xl 三列；下游已定制为 lg 三列，比上游更早切换。此提交会撤回下游已选布局，不作为上游功能缺口。
- AI risks: 若重新设计定价页密度，需按实际卡片内容/视口验证，而非把样式差异当缺陷。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/pricing/components/model-card-grid.tsx:76](https://github.com/QuantumNous/new-api/blob/bee45b58a3c0b77e8dc81e6b5aeb4474aa9058d1/web/src/features/pricing/components/model-card-grid.tsx#L76) — 固定源码：<div className='grid grid-cols-1 gap-3 sm:gap-4 md:grid-cols-2 xl:grid-cols-3'>
- AI evidence: [downstream:web/src/features/pricing/components/model-card-grid.tsx:76](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/pricing/components/model-card-grid.tsx#L76) — 下游对照：<div className='grid grid-cols-1 gap-3 sm:gap-4 md:grid-cols-2 lg:grid-cols-3'>

#### commit-75e533209490

- Title: feat(pricing): support site currency in pricing editors
- Feature/group: feat(pricing): support site currency in pricing editors
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/model-pricing/__tests__/editor-currency.test.tsx`, `web/src/features/model-pricing/currency.ts`, `web/src/features/model-pricing/pricing-amount-input.tsx`, `web/src/features/model-pricing/pricing-currency-selector.tsx`, `web/src/features/system-settings/models/model-pricing-core.ts`, `web/src/features/system-settings/models/model-pricing-inputs.tsx`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/task-pricing-matrix.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/stores/pricing-preferences-store.ts`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 定价编辑器站点币种
- AI recommendation: `defer` — 统一编辑器USD与站点币种换算并保存用户显示偏好，依赖新的模型定价面板。应先确认换算只发生在展示/输入边界，不改变历史账单、钱包quota和独立上游成本币种。
- AI risks: 汇率变动、空/零价格与精度舍入可能改变保存售价，需往返与业务账务测试。
- AI dependencies: commit-0c76e4dae77a
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/model-pricing/currency.ts:34](https://github.com/QuantumNous/new-api/blob/75e533209490a8ef3a8b5e3d93e4dac03ba19bcf/web/src/features/model-pricing/currency.ts#L34) — 新定价模块定义币种换算入口；代码：export function getSitePricingCurrency(

#### commit-99974a814f00

- Title: feat(plugins): extend plugin metadata and icon support
- Feature/group: feat(plugins): extend plugin metadata and icon support
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel.go`, `controller/channel_task_plugin_bind_test.go`, `controller/channel_task_plugin_validation_test.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `model/pricing.go`, `model/pricing_usage_schema_test.go`, `model/task_plugin.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/icon.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/request.go`, `plugins/embed.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/google/plugin.js`, `plugins/tasks/hailuo/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `plugins/tasks/vertex-ai/plugin.js`, `plugins/tasks/vidu/plugin.js`, `router/api-router.go`, `router/plugin_router_test.go`
- Overlap with downstream changes: controller/channel.go, router/api-router.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件元数据、默认地址与图标
- AI recommendation: `defer` — 扩展插件图标、排序与默认 BaseURL，默认地址在创建渠道时固定保存。当前无插件实体和管理接口，需先评审图标解码/体积边界与渠道目的地主机的审计策略。
- AI risks: 不能因插件升级静默改变渠道主机；保留下载地址、SSRF、代理和租户边界。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/channel.go:480](https://github.com/QuantumNous/new-api/blob/99974a814f00a26dfa7431beb313e487529bbeed/controller/channel.go#L480) — 插件默认目的地被固化为渠道属性；代码：// The plugin default is persisted onto the channel instead of being

#### commit-984330920061

- Title: feat(web): improve plugin management and marketplace
- Feature/group: feat(web): improve plugin management and marketplace
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace-plugin-logo.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace.test.ts`, `web/src/features/task-plugins/__tests__/plugin-card.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-detail-sheet.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-icon-file.test.ts`, `web/src/features/task-plugins/__tests__/plugin-icon-image.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-icon.test.ts`, `web/src/features/task-plugins/__tests__/plugin-meta-preview.test.ts`, `web/src/features/task-plugins/__tests__/plugins-table.test.tsx`, `web/src/features/task-plugins/__tests__/website.test.tsx`, `web/src/features/task-plugins/api.ts`, `web/src/features/task-plugins/components/javascript-viewer.tsx`, `web/src/features/task-plugins/components/marketplace-capabilities.tsx`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/marketplace-panel.tsx`, `web/src/features/task-plugins/components/marketplace-plugin-card.tsx`, `web/src/features/task-plugins/components/plugin-card.tsx`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/plugin-endpoints.tsx`, `web/src/features/task-plugins/components/plugin-icon.tsx`, `web/src/features/task-plugins/components/plugin-integrity-check.tsx`, `web/src/features/task-plugins/components/plugin-metadata-card.tsx`, `web/src/features/task-plugins/components/plugin-model-list.tsx`, `web/src/features/task-plugins/components/plugin-website-link.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`, `web/src/features/task-plugins/components/source-diff.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/features/task-plugins/components/usage-schema-table.tsx`, `web/src/features/task-plugins/lib/marketplace.ts`, `web/src/features/task-plugins/lib/plugin-icon-file.ts`, `web/src/features/task-plugins/lib/plugin-icon.ts`, `web/src/features/task-plugins/lib/plugin-meta-preview.ts`, `web/src/features/task-plugins/lib/plugin-website.ts`, `web/src/features/task-plugins/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/test-setup.ts`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件市场与管理交互
- AI recommendation: `defer` — 引入较完整的市场来源、完整性校验、源码预览和上传/安装交互；没有对应后端插件生命周期时不单独引入。应和受信来源、签名/哈希、权限及回滚策略一起设计。
- AI risks: 外部插件源码是可执行内容，不能把市场安装视为普通静态资源下载。
- AI dependencies: commit-eb48396d5fe9, commit-dc4732cfed71, commit-99974a814f00
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/task-plugins/api.ts:93](https://github.com/QuantumNous/new-api/blob/984330920061e014159c7581b8fcdb726de2c247/web/src/features/task-plugins/api.ts#L93) — 前端上传依赖插件后端 API；代码：export async function uploadTaskPlugin(

#### commit-eb76b136b85e

- Title: feat(channels): improve plugin channel setup and icons
- Feature/group: feat(channels): improve plugin channel setup and icons
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/provider-badge.tsx`, `web/src/components/ui/__tests__/combobox.test.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/channel-type-badge.test.tsx`, `web/src/features/channels/components/channel-type-badge.tsx`, `web/src/features/channels/components/channels-columns.tsx`, `web/src/features/channels/components/channels-table.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/__tests__/task-plugin-base-url.test.ts`, `web/src/features/channels/lib/task-plugin-base-url.ts`, `web/src/features/channels/types.ts`
- Overlap with downstream changes: web/src/features/channels/api.ts, web/src/features/channels/components/channels-table.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件渠道创建与图标
- AI recommendation: `defer` — 渠道表单使用插件图标、默认地址与任务类型选择，需要已存在的插件列表和元数据。本地渠道表单包含下游视频、素材、存储配置，应待插件方案批准后按字段迁移。
- AI risks: 整个表单替换会丢失下游渠道选项与本地国际化。
- AI dependencies: commit-eb48396d5fe9, commit-99974a814f00, commit-984330920061
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/channels/lib/task-plugin-base-url.ts:51](https://github.com/QuantumNous/new-api/blob/eb76b136b85ecba9d6ad19c714c5781e51815530/web/src/features/channels/lib/task-plugin-base-url.ts#L51) — 新增插件地址信任警告规则模块；代码：export function assessBaseUrlTrust(
- AI evidence: [downstream:web/src/features/channels/components/drawers/channel-mutate-drawer.tsx:621](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/drawers/channel-mutate-drawer.tsx#L621) — 本地下游复用渠道编辑器承载现有渠道配置；代码：export function ChannelMutateDrawer({

#### commit-0e0ba152bdcc

- Title: feat(pricing): improve pricing editors and log display
- Feature/group: feat(pricing): improve pricing editors and log display
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/model-pricing/__tests__/editor-layout.test.tsx`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/models/__tests__/metadata-editing.test.tsx`, `web/src/features/models/components/data-table-row-actions.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/models/components/models-columns.tsx`, `web/src/features/models/components/models-dialogs.tsx`, `web/src/features/models/components/models-provider.tsx`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/task-matrix-display.ts`, `web/src/features/pricing/lib/task-price-display.ts`, `web/src/features/pricing/types.ts`, `web/src/features/system-settings/components/settings-form-layout.tsx`, `web/src/features/system-settings/models/__tests__/task-pricing-copy.test.tsx`, `web/src/features/system-settings/models/model-pricing-inputs.tsx`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-ratio-form.tsx`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/ratio-settings-card.tsx`, `web/src/features/system-settings/models/task-pricing-matrix.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`
- Overlap with downstream changes: web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/features/usage-logs/components/columns/common-logs-columns.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型价格编辑与日志展示
- AI recommendation: `defer` — 在模型管理中接入价格操作，整理动态价、任务矩阵、编辑器和日志表达式展示；依赖统一模型管理和币种编辑。当前本地有特殊费用、有效调账与成本权限提示，应按最终展示模块逐处迁移。
- AI risks: 全量覆盖日志列会丢失下游管理员饱和、成本和历史调账证据。
- AI dependencies: commit-0c76e4dae77a, commit-75e533209490
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/models/components/data-table-row-actions.tsx:31](https://github.com/QuantumNous/new-api/blob/0e0ba152bdcc6891f6053047ccf14d41b3cad60a/web/src/features/models/components/data-table-row-actions.tsx#L31) — 新增价格编辑权限入口；代码：import { useCanEditModelPricing } from '@/features/model-pricing/api'

#### commit-71c1fd7caad7

- Title: feat(models): improve model listing, pricing and visibility filters
- Feature/group: feat(models): improve model listing, pricing and visibility filters
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/model_management_test.go`, `controller/model_meta.go`, `model/model_meta.go`, `model/pricing.go`, `web/src/features/model-pricing/__tests__/editor-currency.test.tsx`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/model-pricing/pricing.ts`, `web/src/features/models/__tests__/metadata-editing.test.tsx`, `web/src/features/models/__tests__/model-listing.test.tsx`, `web/src/features/models/components/data-table-bulk-actions.tsx`, `web/src/features/models/components/data-table-row-actions.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/models/components/model-connections.tsx`, `web/src/features/models/components/model-square-status.tsx`, `web/src/features/models/components/models-columns.tsx`, `web/src/features/models/components/models-table.tsx`, `web/src/features/models/lib/model-utils.ts`, `web/src/features/models/types.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/components/pricing-columns.tsx`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/price.ts`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/routes/_authenticated/models/$section.tsx`
- Overlap with downstream changes: web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型广场可见性与模型列表
- AI recommendation: `defer` — 新增可见/不可用/隐藏/部分可用状态、供应商筛选和模型连接/价格入口。依赖前述元数据管理，并改变公开可见性判断，需要保留租户组和本地财务权限。
- AI risks: 模型广场隐藏和管理员可编辑不能混为访问授权；批量操作不得误删运行时配置。
- AI dependencies: commit-0c76e4dae77a, commit-0e0ba152bdcc
- AI suggested paths (not approval):
- AI evidence: [upstream:model/model_meta.go:21](https://github.com/QuantumNous/new-api/blob/71c1fd7caad738db4d13aabbf28eeadb293d0cfe/model/model_meta.go#L21) — 模型元数据新增广场状态；代码：type ModelSquareState string

#### commit-a5e41a893379

- Title: feat(usage-logs): refine mobile layout and keep quick actions visible
- Feature/group: feat(usage-logs): refine mobile layout and keep quick actions visible
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/data-table/core/__tests__/pagination.test.tsx`, `web/src/components/data-table/core/pagination.tsx`, `web/src/components/data-table/layout/data-table-page.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-filter.test.tsx`, `web/src/features/usage-logs/components/common-log-mobile-card.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`, `web/src/features/usage-logs/components/compact-date-time-range-picker.tsx`, `web/src/features/usage-logs/components/logs-filter-toolbar.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`, `web/src/features/usage-logs/components/timing-metrics-cell.tsx`, `web/src/features/usage-logs/components/usage-logs-mobile-card.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/features/usage-logs/components/compact-date-time-range-picker.tsx, web/src/features/usage-logs/components/usage-logs-mobile-card.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 用量日志移动卡片及快捷操作
- AI recommendation: `defer` — 上游新增 CommonLogMobileCard、紧凑分页和工具栏；下游日志扩展成本、财务纠错、永久账本与审计信息，需先设计字段映射再迁入布局。
- AI risks: 移动卡片不能遗漏或误显示有效扣费/成本、请求ID和仅管理员可见内容。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/components/data-table/layout/data-table-page.tsx:210](https://github.com/QuantumNous/new-api/blob/a5e41a893379e49bd9e1d025e775c04df8961c95/web/src/components/data-table/layout/data-table-page.tsx#L210) — 固定源码：compactPagination?: boolean
- AI evidence: [downstream:web/src/features/usage-logs/components/usage-logs-table.tsx:80](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/usage-logs/components/usage-logs-table.tsx#L80) — 下游对照：export function UsageLogsTable({ logCategory }: UsageLogsTableProps) {

#### commit-2bec370629aa

- Title: feat(web): refine API key and user quota displays
- Feature/group: feat(web): refine API key and user quota displays
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/data-table/core/data-table-view.tsx`, `web/src/components/data-table/core/truncated-cell.tsx`, `web/src/components/group-badge.tsx`, `web/src/components/ui/table.tsx`, `web/src/features/keys/components/__tests__/api-key-group-cell.test.tsx`, `web/src/features/keys/components/__tests__/api-key-group-combobox.test.tsx`, `web/src/features/keys/components/__tests__/api-key-listing.test.tsx`, `web/src/features/keys/components/__tests__/auto-group-order-editor.test.tsx`, `web/src/features/keys/components/api-key-group-cell.tsx`, `web/src/features/keys/components/api-key-quota-cell.tsx`, `web/src/features/keys/components/api-key-timestamp-cell.tsx`, `web/src/features/keys/components/api-keys-cells.tsx`, `web/src/features/keys/components/api-keys-columns.tsx`, `web/src/features/keys/components/api-keys-table.tsx`, `web/src/features/keys/components/auto-group-visuals.tsx`, `web/src/features/keys/components/data-table-row-actions.tsx`, `web/src/features/users/components/__tests__/quota-display.test.tsx`, `web/src/features/users/components/user-quota-cell.tsx`, `web/src/features/users/components/users-columns.tsx`, `web/src/features/users/components/users-table.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/styles/index.css`
- Overlap with downstream changes: web/src/features/users/components/users-table.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/styles/index.css
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: API key 与用户额度展示重组
- AI recommendation: `defer` — 上游新增多个 quota/group/timestamp cell 组件并调整表格尺寸；本地 API key 额度仍内联，用户额度还展示钱包有效余额。建议待统一展示方案后按下游有效金额迁移。
- AI risks: 不可用上游原始 quota 覆盖本地更正后的财务展示；需同时覆盖桌面/移动端和权限。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/keys/components/api-key-quota-cell.tsx:43](https://github.com/QuantumNous/new-api/blob/2bec370629aa72d74d3f598b953a4d33b8bcbf7d/web/src/features/keys/components/api-key-quota-cell.tsx#L43) — 固定源码：export function ApiKeyQuotaCell(props: ApiKeyQuotaCellProps) {
- AI evidence: [downstream:web/src/features/keys/components/api-keys-columns.tsx:145](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/api-keys-columns.tsx#L145) — 下游对照：accessorKey: 'remain_quota',

#### commit-8f72ecbbfd86

- Title: feat(usage-logs): add searchable group filter
- Feature/group: feat(usage-logs): add searchable group filter
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/ui/combobox-input.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/usage-logs/components/__tests__/group-filter.test.tsx`, `web/src/features/usage-logs/components/__tests__/log-type-filter.test.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/common-logs-filter-bar.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 日志可搜索用户组筛选
- AI recommendation: `migrate` — 上游新增 admin/user 分别获取可选组的可搜索筛选。本地日志仍有旧筛选组件，可按当前租户和权限迁入组查找与键盘支持。
- AI risks: 组列表权限、自由文本/选择值切换、URL 恢复和清空行为需覆盖；不要复用管理员组给普通用户。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/components/ui/combobox-input.tsx, web/src/components/ui/combobox.tsx, web/src/features/usage-logs/components/__tests__/group-filter.test.tsx
- AI evidence: [upstream:web/src/features/usage-logs/components/common-logs-filter-bar.tsx:128](https://github.com/QuantumNous/new-api/blob/8f72ecbbfd86ded5ee15373921533eed3d334a9a/web/src/features/usage-logs/components/common-logs-filter-bar.tsx#L128) — 固定源码：const { data: adminGroups } = useQuery({
- AI evidence: [downstream:web/src/features/usage-logs/components/common-logs-filter-bar.tsx:112](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/usage-logs/components/common-logs-filter-bar.tsx#L112) — 下游对照：export function CommonLogsFilterBar<TData>(

#### commit-551bb63edf40

- Title: fix(keys): show desktop quota amounts side by side
- Feature/group: fix(keys): show desktop quota amounts side by side
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/__tests__/api-key-listing.test.tsx`, `web/src/features/keys/components/api-key-quota-cell.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: API key 桌面额度并排显示
- AI recommendation: `defer` — 此修复针对前项新增 ApiKeyQuotaCell；下游尚无该组件，不能单独搬补丁。与额度展示重组一起评审。
- AI risks: 窄屏、无限额度、极大金额和有效账务金额需保留可读性。
- AI dependencies: commit-2bec370629aa
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/keys/components/api-key-quota-cell.tsx:108](https://github.com/QuantumNous/new-api/blob/551bb63edf4007d7c4b0930504faf6db87012d73/web/src/features/keys/components/api-key-quota-cell.tsx#L108) — 固定源码：'grid w-full min-w-0 items-baseline gap-x-2 gap-y-1 text-xs',
- AI evidence: [downstream:web/src/features/keys/components/api-keys-columns.tsx:145](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/api-keys-columns.tsx#L145) — 下游对照：accessorKey: 'remain_quota',

#### commit-bd22e45a740a

- Title: style(keys): widen the desktop quota column
- Feature/group: style(keys): widen the desktop quota column
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/api-keys-columns.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: API key 额度列宽调整
- AI recommendation: `defer` — 只为新增额度 cell 扩宽 quota 列；本地仍内联渲染，不宜在未迁入该布局时机械设宽。
- AI risks: 应联合既有列表密度与移动卡片验证，避免横向滚动回退。
- AI dependencies: commit-2bec370629aa, commit-551bb63edf40
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/keys/components/api-keys-columns.tsx:141](https://github.com/QuantumNous/new-api/blob/bd22e45a740a7c02c328704dc419d8795237cb74/web/src/features/keys/components/api-keys-columns.tsx#L141) — 固定源码：minSize: 260,
- AI evidence: [downstream:web/src/features/keys/components/api-keys-columns.tsx:145](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/api-keys-columns.tsx#L145) — 下游对照：accessorKey: 'remain_quota',

#### commit-950644c9d544

- Title: fix(keys): preserve spacing after desktop quota content
- Feature/group: fix(keys): preserve spacing after desktop quota content
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/__tests__/api-key-listing.test.tsx`, `web/src/features/keys/components/api-key-quota-cell.tsx`, `web/src/features/keys/components/api-keys-table.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: API key 额度内容留白修正
- AI recommendation: `defer` — 调整新增 quota cell 的最大宽度并去掉表格额外留白，依赖前述额度组件链；本地不存在同组件。
- AI risks: 应以最终布局评审，避免先宽后窄中间态及金额截断。
- AI dependencies: commit-2bec370629aa, commit-551bb63edf40, commit-bd22e45a740a
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/keys/components/api-key-quota-cell.tsx:86](https://github.com/QuantumNous/new-api/blob/950644c9d54445bdd8796643e4e2e18e13167a4c/web/src/features/keys/components/api-key-quota-cell.tsx#L86) — 固定源码：props.variant === 'card' ? 'space-y-2.5' : 'max-w-45 space-y-1.5'
- AI evidence: [downstream:web/src/features/keys/components/api-keys-columns.tsx:145](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/api-keys-columns.tsx#L145) — 下游对照：accessorKey: 'remain_quota',

#### commit-524455fac3c4

- Title: feat(redemptions): add batch deletion and optional file exports
- Feature/group: feat(redemptions): add batch deletion and optional file exports
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/audit.go`, `controller/redemption.go`, `controller/redemption_batch_test.go`, `middleware/audit.go`, `model/redemption.go`, `router/api-router.go`, `web/src/features/redemption-codes/api.ts`, `web/src/features/redemption-codes/components/__tests__/bulk-delete.test.tsx`, `web/src/features/redemption-codes/components/__tests__/redemptions-export.test.tsx`, `web/src/features/redemption-codes/components/data-table-bulk-actions.tsx`, `web/src/features/redemption-codes/components/redemptions-export-dialog.tsx`, `web/src/features/redemption-codes/components/redemptions-mutate-drawer.tsx`, `web/src/features/redemption-codes/components/redemptions-table.tsx`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/audit.go, router/api-router.go, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 兑换码批量删除与文件导出
- AI recommendation: `defer` — 上游增加最多1000个 ID 的批量软删除、审计和可选导出；本地兑换码会进入资金及账务追溯，需确认批删权限、已使用记录可追溯性及导出策略再选择。
- AI risks: 批删不能破坏兑换与账本证据；导出包含敏感兑换码，需权限、可见性及审计验证。
- AI dependencies: commit-d8cb177440ce
- AI suggested paths (not approval):
- AI evidence: [upstream:model/redemption.go:244](https://github.com/QuantumNous/new-api/blob/524455fac3c438321df4ae9ed4ffd11bc635cc4f/model/redemption.go#L244) — 固定源码：func BatchDeleteRedemptions(ids []int) (int64, error) {
- AI evidence: [downstream:router/api-router.go:429](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/api-router.go#L429) — 下游对照：redemptionRoute := apiRouter.Group("/redemption")

#### commit-fff0635bb14b

- Title: docs: update project architecture and Go conventions
- Feature/group: docs: update project architecture and Go conventions
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`
- Overlap with downstream changes: AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 架构和 Go 规范文档重写
- AI recommendation: `skip` — 上游架构已包含JS插件/PAT等下游尚未采用的能力；直接更新 AGENTS 会形成不真实架构说明并覆盖项目约束。
- AI risks: 后续实际迁入能力后再写准确下游架构，保护当前数据库/relaykit与身份规则。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:AGENTS.md:16](https://github.com/QuantumNous/new-api/blob/fff0635bb14b8ec5f10df9e582a83591df841902/AGENTS.md#L16) — 固定源码：- **Extensions**: JavaScript task plugins executed by Sobek; Electron desktop wrapper
- AI evidence: [downstream:AGENTS.md:67](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/AGENTS.md#L67) — 下游对照：**relaykit module independence:** The `relaykit/` Go module MUST remain independently buildable.

#### commit-ebe4c368f287

- Title: refactor: modernize Go code conventions
- Feature/group: refactor: modernize Go code conventions
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/custom-event.go`, `common/etag.go`, `common/gopool.go`, `common/init.go`, `common/quota_math.go`, `common/redis.go`, `common/session_cookie.go`, `common/ssrf_protection.go`, `common/str.go`, `common/totp.go`, `common/utils.go`, `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/audit.go`, `controller/auth_flow_test.go`, `controller/channel-test.go`, `controller/channel.go`, `controller/channel_authz_test.go`, `controller/channel_upstream_update.go`, `controller/channel_upstream_update_test.go`, `controller/custom_oauth.go`, `controller/deployment.go`, `controller/email_binding.go`, `controller/group.go`, `controller/model.go`, `controller/model_management_test.go`, `controller/model_meta.go`, `controller/model_pricing_config.go`, `controller/model_sync.go`, `controller/oauth.go`, `controller/option.go`, `controller/passkey.go`, `controller/pricing.go`, `controller/ratio_sync.go`, `controller/redemption.go`, `controller/relay_error_log_test.go`, `controller/secure_verification.go`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `controller/subscription.go`, `controller/task_plugin.go`, `controller/telegram_test.go`, `controller/topup.go`, `controller/topup_waffo.go`, `controller/topup_waffo_pancake_test.go`, `controller/twofa.go`, `controller/uptime_kuma.go`, `controller/user.go`, `controller/wechat.go`, `dto/midjourney.go`, `dto/suno.go`, `middleware/audit.go`, `middleware/auth.go`, `middleware/rate-limit.go`, `middleware/task_artifact_access.go`, `model/audit_log.go`, `model/channel.go`, `model/channel_cache.go`, `model/channel_satisfy.go`, `model/channel_status_test.go`, `model/checkin.go`, `model/gorm_logger.go`, `model/json_column_test.go`, `model/log.go`, `model/log_format_test.go`, `model/log_other_test.go`, `model/main.go`, `model/model_pricing_config.go`, `model/option.go`, `model/passkey.go`, `model/perf_metric.go`, `model/prefill_group.go`, `model/pricing.go`, `model/quota_reserve.go`, `model/quota_reserve_test.go`, `model/redemption.go`, `model/redemption_test.go`, `model/subscription.go`, `model/task.go`, `model/task_cas_test.go`, `model/token.go`, `model/topup.go`, `model/twofa.go`, `model/usedata.go`, `model/user.go`, `model/user_auth_cache.go`, `model/user_authentication_test.go`, `model/user_cache.go`, `model/user_session.go`, `model/user_session_test.go`, `model/user_update_test.go`, `model/utils.go`, `oauth/registry.go`, `pkg/billingexpr/compile.go`, `pkg/billingexpr/run.go`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/routing.go`, `plugins/hailuo_responses_test.go`, `relay/channel/ali/dto.go`, `relay/channel/api_request_getbody_test.go`, `relay/channel/aws/dto.go`, `relay/channel/aws/relay-aws.go`, `relay/channel/claude/relay_claude_test.go`, `relay/channel/gemini/adaptor.go`, `relay/channel/gemini/adaptor_responses_test.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_gemini_usage_test.go`, `relay/channel/ollama/dto.go`, `relay/channel/ollama/relay-ollama.go`, `relay/channel/openai/adaptor.go`, `relay/channel/openai/chat_via_responses_test.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_image.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/tencent/dto.go`, `relay/channel/tencent/relay-tencent.go`, `relay/common/outbound_body_test.go`, `relay/common/override.go`, `relay/common/override_test.go`, `relay/common/relay_info.go`, `relay/common/relay_utils.go`, `relay/common/stream_status_test.go`, `relay/common/tool_usage_test.go`, `relay/convert_request_error_test.go`, `relay/helper/billing_expr_request.go`, `relay/helper/common.go`, `relay/helper/stream_scanner_test.go`, `relay/mjproxy_handler.go`, `relay/plugin_protocol.go`, `relay/relay_task.go`, `relay/request_clone_test.go`, `relaykit/dto/billing_usage.go`, `relaykit/dto/claude.go`, `relaykit/dto/gemini.go`, `relaykit/dto/notify.go`, `relaykit/dto/openai_image.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_response.go`, `relaykit/dto/pricing.go`, `relaykit/dto/ratio_sync.go`, `relaykit/dto/realtime.go`, `relaykit/relayconvert/internal/claude_messages/citations.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/req_helpers.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_stream_resp_test.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp_test.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_stream_resp.go`, `relaykit/relayconvert/internal/toolconv/decode.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/kitutil/mask.go`, `relaykit/relayconvert/kitutil/value.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/suffix.go`, `relaykit/relayconvert/response_registry_test.go`, `router/plugin-router.go`, `service/auth_session.go`, `service/auth_session_test.go`, `service/auth_token.go`, `service/authz/adapter.go`, `service/authz/registry.go`, `service/authz/resolver.go`, `service/channel_affinity.go`, `service/channel_affinity_template_test.go`, `service/email_binding.go`, `service/error_test.go`, `service/file_decoder.go`, `service/file_service.go`, `service/funding_source.go`, `service/group.go`, `service/http_client_transport_test.go`, `service/http_transport_sharded.go`, `service/log_info_generate.go`, `service/midjourney.go`, `service/notify-limit.go`, `service/quota.go`, `service/quota_saturation_test.go`, `service/rankings.go`, `service/system_instance.go`, `service/task_billing_test.go`, `service/task_plugin_audit.go`, `service/task_polling.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/token_counter.go`, `service/violation_fee.go`, `service/webhook.go`, `setting/auto_group.go`, `setting/billing_setting/tiered_billing.go`, `setting/billing_setting/tiered_billing_test.go`, `setting/config/config.go`, `setting/model_setting/claude.go`, `setting/model_setting/gemini.go`, `setting/model_setting/global.go`, `setting/operation_setting/channel_affinity_setting.go`, `setting/operation_setting/operation_setting.go`, `setting/operation_setting/tools.go`, `setting/operation_setting/tools_price_test.go`, `setting/ratio_setting/exposed_cache.go`, `setting/ratio_setting/model_ratio.go`, `setting/sensitive.go`, `setting/system_setting/task_artifact_store.go`, `setting/task_plugin.go`, `setting/user_usable_group.go`
- Overlap with downstream changes: common/redis.go, common/ssrf_protection.go, controller/audit.go, controller/channel-test.go, controller/channel.go, controller/channel_authz_test.go, controller/model.go, controller/option.go, controller/secure_verification.go, controller/user.go, model/channel.go, model/channel_cache.go, model/gorm_logger.go, model/log.go, model/log_format_test.go, model/main.go, model/option.go, model/task.go, model/task_cas_test.go, model/twofa.go, model/usedata.go, model/user.go, model/user_authentication_test.go, model/user_cache.go, model/user_update_test.go, relay/channel/openai/adaptor.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, relay/helper/common.go, relay/mjproxy_handler.go, relay/relay_task.go, relaykit/dto/openai_request.go, relaykit/relayconvert/kitutil/mask.go, service/auth_session.go, service/auth_session_test.go, service/auth_token.go, service/funding_source.go, service/http_client_transport_test.go, service/midjourney.go, service/quota.go, service/task_billing_test.go, service/task_polling.go, service/text_quota.go, service/violation_fee.go, service/webhook.go, setting/ratio_setting/model_ratio.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 214文件 Go 语法现代化
- AI recommendation: `skip` — 样例主要是 interface{}→any、reflect.TypeFor 等广泛语法整理，未提供需要本批次修复的运行时缺口。跳过全仓重排以降低下游账务/协议/身份变更的冲突噪声。
- AI risks: 以后功能迁入可就地采用必要写法；不能把纯重构算作已同步的业务能力。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:common/utils.go:191](https://github.com/QuantumNous/new-api/blob/ebe4c368f28787919ea478c0abf1ea7101978179/common/utils.go#L191) — 固定源码：func Interface2String(inter any) string {
- AI evidence: [downstream:common/utils.go:191](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/common/utils.go#L191) — 下游对照：func Interface2String(inter interface{}) string {

#### commit-ea7cb0ba4e0f

- Title: refactor(web): unify table cells and quota details
- Feature/group: refactor(web): unify table cells and quota details
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/activity-time-cell.tsx`, `web/src/components/data-table/core/data-table-view.tsx`, `web/src/components/group-badge.tsx`, `web/src/components/masked-value-display.tsx`, `web/src/components/quota-details-popover.tsx`, `web/src/components/ui/table.tsx`, `web/src/features/keys/components/__tests__/api-key-group-cell.test.tsx`, `web/src/features/keys/components/__tests__/api-key-group-combobox.test.tsx`, `web/src/features/keys/components/__tests__/api-key-listing.test.tsx`, `web/src/features/keys/components/api-key-quota-cell.tsx`, `web/src/features/keys/components/api-key-timestamp-cell.tsx`, `web/src/features/keys/components/api-keys-cells.tsx`, `web/src/features/redemption-codes/components/redemptions-columns.tsx`, `web/src/features/users/components/__tests__/quota-display.test.tsx`, `web/src/features/users/components/user-quota-cell.tsx`, `web/src/features/users/components/users-columns.tsx`, `web/src/features/users/components/users-table.tsx`
- Overlap with downstream changes: web/src/features/users/components/users-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 统一表格时间与额度明细单元格
- AI recommendation: `defer` — 新增 ActivityTimeCell/QuotaDetailsPopover 并替换多个用户、兑换码、API key 单元格；本地下游财务金额和既有表格布局不同。建议与额度展示组统一评审，避免在新金额逻辑前搬通用组件。
- AI risks: 原始 quota、有效余额、历史成本不能混成相同显示；可访问标签和复制操作须保留。
- AI dependencies: commit-2bec370629aa, commit-551bb63edf40, commit-bd22e45a740a, commit-950644c9d544
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/components/data-table/core/data-table-view.tsx:149](https://github.com/QuantumNous/new-api/blob/ea7cb0ba4e0f82e2bfa5e55752eb68bdf902f71b/web/src/components/data-table/core/data-table-view.tsx#L149) — 固定源码：withContainer={false}
- AI evidence: [downstream:web/src/features/keys/components/api-keys-columns.tsx:145](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/api-keys-columns.tsx#L145) — 下游对照：accessorKey: 'remain_quota',

#### commit-9bf328d97497

- Title: fix: preserve provider fields in Sora video queries
- Feature/group: fix: preserve provider fields in Sora video queries
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/README.md`, `docs/plugin-api/v1.md`, `plugins/tasks/sora/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Sora 查询保留供应商扩展字段
- AI recommendation: `defer` — 保留 Sora 上游原始查询 JSON，再由宿主覆盖可信状态与完成时间，避免转换时丢字段。方案依赖插件任务视图，可在未来驱动迁移时实现；现阶段不应新加整个 JS 适配层。
- AI risks: 必须由宿主控制状态/权限字段，不能让不可信供应商 JSON 覆盖公开任务身份。
- AI dependencies: commit-eb48396d5fe9
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/task/jsplugin/adaptor.go:818](https://github.com/QuantumNous/new-api/blob/9bf328d9749751757d5d6b74088d514813a618bd/relay/channel/task/jsplugin/adaptor.go#L818) — 宿主覆盖状态同时保留扩展字段；代码：rendered["status"] = host.Status

#### commit-4fc9d1f1fa77

- Title: fix(options): rebuild options table primary key and stop pricing writes resetting rows
- Feature/group: fix(options): rebuild options table primary key and stop pricing writes resetting rows
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/main.go`, `model/model_pricing_config.go`, `model/option_primary_key_migration.go`
- Overlap with downstream changes: model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 历史Options主键修复
- AI recommendation: `defer` — 上游自动识别无主键options表、去重后交换新表并备份旧行，同时修复新价格管理的默认行写入。本地模型已有primaryKey声明但未审计实际旧库，且无新model_pricing_config；不在同步中默认运行最后值胜出和表交换，需要三数据库真实旧表样本及恢复方案后单独决定。
- AI risks: 配置可能含计费/密钥/运行参数；去重选择、锁、交换失败和多节点读取都需验证，不能丢失历史值。
- AI dependencies: commit-9a8674425c5a, commit-0c76e4dae77a
- AI suggested paths (not approval):
- AI evidence: [upstream:model/option_primary_key_migration.go:42](https://github.com/QuantumNous/new-api/blob/4fc9d1f1fa77c0cfdd9719cb59ff9ecc9885d66a/model/option_primary_key_migration.go#L42) — 仅确认缺唯一键后才重建表，但涉及自动数据选择与交换；代码：return repairOptionPrimaryKey(locked)
- AI evidence: [downstream:model/option.go:24](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/option.go#L24) — 本地模型有主键声明，仍不能据此推定所有历史库物理形态；代码：Key   string `json:"key" gorm:"primaryKey"`

#### commit-c79b74b68358

- Title: fix(frontend): deduplicate /api/status requests (#7189)
- Feature/group: fix(frontend): deduplicate /api/status requests (#7189)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`, `web/src/hooks/use-status.ts`, `web/src/hooks/use-system-config.ts`, `web/src/lib/__tests__/status-query.test.tsx`, `web/src/lib/nav-modules.ts`, `web/src/lib/status-query.ts`, `web/src/main.tsx`, `web/src/routes/pricing/$modelId/index.tsx`, `web/src/routes/pricing/index.tsx`, `web/src/routes/rankings/index.tsx`
- Overlap with downstream changes: web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/hooks/use-system-config.ts, web/src/lib/nav-modules.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 前端全局status请求去重
- AI recommendation: `migrate` — 建议把getStatus统一到共享React Query缓存及状态映射，并让路由守卫在需要时等待有效刷新。下游虽有useStatus共享key，其他直接调用没有统一入口；迁移时必须保留本地品牌、短信微信就绪状态、业务导航和配置字段。
- AI risks: 缓存只能包含公开全局字段；陈旧隐藏模块不能变成服务端授权，错误/过期/并发要有稳定测试。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx, web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/hooks/use-status.ts, web/src/hooks/use-system-config.ts, web/src/lib/__tests__/status-query.test.tsx, web/src/lib/nav-modules.ts, web/src/lib/status-query.ts, web/src/main.tsx, web/src/routes/pricing/$modelId/index.tsx, web/src/routes/pricing/index.tsx, web/src/routes/rankings/index.tsx
- AI evidence: [upstream:web/src/lib/status-query.ts:43](https://github.com/QuantumNous/new-api/blob/c79b74b68358180c68440057596bdc34a99cc649/web/src/lib/status-query.ts#L43) — 所有status消费者使用同一查询键与策略；代码：export const STATUS_QUERY_KEY = ['status'] as const
- AI evidence: [downstream:web/src/hooks/use-status.ts:44](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/hooks/use-status.ts#L44) — 本地hook有查询缓存，但尚无覆盖所有消费者的统一status模块；代码：const status = await getStatus()

#### commit-876903a8eb22

- Title: fix: 修正火山方舟渠道获取模型列表的端点路径 (#7203)
- Feature/group: fix: 修正火山方舟渠道获取模型列表的端点路径 (#7203)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel_upstream_update.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 火山方舟模型列表端点
- AI recommendation: `merge` — 本地VolcEngine分支仍请求/v1/models，上游改为/api/v3/models且保留special base路径。可优先完整采用单点修复并补路径回归，不影响下游Volc Native Action API。
- AI risks: 必须区分VolcEngine兼容渠道与Volc Native独立API，special base既有/v1/models不可被覆盖。
- AI dependencies: none identified
- AI suggested paths (not approval): controller/channel_upstream_update.go, controller/channel_upstream_update_test.go
- AI evidence: [upstream:controller/channel_upstream_update.go:426](https://github.com/QuantumNous/new-api/blob/876903a8eb22c44e395c03da38f6701c650651ae/controller/channel_upstream_update.go#L426) — 上游修正普通火山兼容渠道模型列表端点。
- AI evidence: [downstream:controller/channel_upstream_update.go:411](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/channel_upstream_update.go#L411) — 本地此分支仍拼接/v1/models。

#### commit-d52bdc0b4087

- Title: feat(billing): add time-based pricing editor and expression previews
- Feature/group: feat(billing): add time-based pricing editor and expression previews
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `pkg/billingexpr/billingexpr_test.go`, `pkg/billingexpr/testdata/frontend_simulation.json`, `web/src/components/ui/combobox-input.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/model-pricing/__tests__/editor-currency.test.tsx`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/cached-price-cell.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/components/pricing-columns.tsx`, `web/src/features/pricing/hooks/use-billing-time.ts`, `web/src/features/pricing/lib/__tests__/billing-expression.test.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-expression/README.md`, `web/src/features/pricing/lib/billing-expression/condition-display.ts`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/parser.ts`, `web/src/features/pricing/lib/billing-expression/runtime.ts`, `web/src/features/pricing/lib/billing-expression/structure.ts`, `web/src/features/pricing/lib/billing-expression/types.ts`, `web/src/features/pricing/lib/billing-expression/visual.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/task-expr.ts`, `web/src/features/pricing/lib/tier-expr.ts`, `web/src/features/system-settings/models/__tests__/request-simulation.test.tsx`, `web/src/features/system-settings/models/__tests__/visual-billing-editor.test.tsx`, `web/src/features/system-settings/models/billing-time-fields.tsx`, `web/src/features/system-settings/models/draft-number-input.tsx`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/request-simulation.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/models/tier-price-fields.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/features/system-settings/models/visual-billing-document-editor.tsx`, `web/src/features/system-settings/models/visual-condition-tree.css`, `web/src/features/system-settings/models/visual-condition-tree.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 时间定价可视编辑及表达式预览
- AI recommendation: `defer` — 引入独立表达式解析、运行、可视文档与时钟预览，替换现有正则式展示。范围达48文件且后续固定请求计费继续扩展，建议作为独立功能组选定，先合ac381acf4b的明确正确性修复。
- AI risks: 前端模拟必须与Go表达式和冻结日志语义一致；不能改变未识别表达式、AST规则跟踪及后台实际扣费。
- AI dependencies: commit-ac381acf4bf4, commit-0e0ba152bdcc
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/pricing/lib/billing-expression/parser.ts:425](https://github.com/QuantumNous/new-api/blob/d52bdc0b4087d50d87eb06387e86b2be53e85cd4/web/src/features/pricing/lib/billing-expression/parser.ts#L425) — 新增独立表达式解析器模块；代码：export function compileBillingExpression(source: string): CompilationResult {

#### commit-12be9975c0bf

- Title: fix(web): unify server error notifications
- Feature/group: fix(web): unify server error notifications
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/AGENTS.md`, `web/src/components/layout/components/chat-presets-item.tsx`, `web/src/components/sign-out-dialog.tsx`, `web/src/features/about/index.tsx`, `web/src/features/auth/forgot-password/components/forgot-password-form.tsx`, `web/src/features/auth/hooks/use-email-verification.ts`, `web/src/features/auth/hooks/use-oauth-login.ts`, `web/src/features/auth/otp/components/otp-form.tsx`, `web/src/features/auth/reset-password-confirm/index.tsx`, `web/src/features/auth/sign-in/components/user-auth-form.tsx`, `web/src/features/auth/sign-up/components/sign-up-form.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/channel-type-badge.tsx`, `web/src/features/channels/components/channels-columns.tsx`, `web/src/features/channels/components/channels-table.tsx`, `web/src/features/channels/components/dialogs/balance-query-dialog.tsx`, `web/src/features/channels/components/dialogs/channel-test-dialog.tsx`, `web/src/features/channels/components/dialogs/codex-usage-dialog.tsx`, `web/src/features/channels/components/dialogs/edit-tag-dialog.tsx`, `web/src/features/channels/components/dialogs/fetch-models-dialog.tsx`, `web/src/features/channels/components/dialogs/multi-key-manage-dialog.tsx`, `web/src/features/channels/components/dialogs/ollama-models-dialog.tsx`, `web/src/features/channels/components/dialogs/tag-batch-edit-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/hooks/use-channel-key-disclosure.ts`, `web/src/features/channels/hooks/use-channel-mutate-form.ts`, `web/src/features/channels/hooks/use-channel-upstream-updates.ts`, `web/src/features/channels/index.tsx`, `web/src/features/channels/lib/channel-actions.ts`, `web/src/features/chat/hooks/use-active-chat-key.ts`, `web/src/features/dashboard/components/flow/flow-charts.tsx`, `web/src/features/dashboard/components/models/performance-overview.tsx`, `web/src/features/dashboard/components/overview/overview-dashboard.tsx`, `web/src/features/dashboard/components/overview/performance-health-panel.tsx`, `web/src/features/dashboard/components/overview/summary-cards.tsx`, `web/src/features/dashboard/components/users/user-charts.tsx`, `web/src/features/dashboard/lib/flow-selection.ts`, `web/src/features/home/hooks/use-home-page-content.ts`, `web/src/features/keys/components/api-keys-columns.tsx`, `web/src/features/keys/components/api-keys-delete-dialog.tsx`, `web/src/features/keys/components/api-keys-multi-delete-dialog.tsx`, `web/src/features/keys/components/api-keys-mutate-drawer.tsx`, `web/src/features/keys/components/api-keys-provider.tsx`, `web/src/features/keys/components/api-keys-table.tsx`, `web/src/features/keys/components/data-table-row-actions.tsx`, `web/src/features/keys/components/dialogs/cc-switch-dialog.tsx`, `web/src/features/model-pricing/__tests__/save-errors.test.tsx`, `web/src/features/model-pricing/api.ts`, `web/src/features/models/components/deployments-table.tsx`, `web/src/features/models/components/dialogs/create-deployment-drawer.tsx`, `web/src/features/models/components/dialogs/extend-deployment-dialog.tsx`, `web/src/features/models/components/dialogs/missing-models-dialog.tsx`, `web/src/features/models/components/dialogs/model-delete-dialog.tsx`, `web/src/features/models/components/dialogs/prefill-group-management-dialog.tsx`, `web/src/features/models/components/dialogs/rename-deployment-dialog.tsx`, `web/src/features/models/components/dialogs/sync-wizard-dialog.tsx`, `web/src/features/models/components/dialogs/update-config-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-mutate-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-operation-dialog.tsx`, `web/src/features/models/components/dialogs/vendors-management-dialog.tsx`, `web/src/features/models/components/dialogs/view-details-dialog.tsx`, `web/src/features/models/components/dialogs/view-logs-dialog.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/models/components/drawers/prefill-group-form-drawer.tsx`, `web/src/features/models/components/models-table.tsx`, `web/src/features/models/components/vendor-linked-models.tsx`, `web/src/features/models/components/vendors-table.tsx`, `web/src/features/models/index.tsx`, `web/src/features/models/lib/model-actions.ts`, `web/src/features/models/lib/vendor-actions.ts`, `web/src/features/models/vendor-api.ts`, `web/src/features/playground/api.ts`, `web/src/features/playground/hooks/use-chat-handler.ts`, `web/src/features/playground/hooks/use-playground-options.ts`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-details-performance.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/hooks/use-pricing-data.ts`, `web/src/features/profile/components/checkin-calendar-card.tsx`, `web/src/features/profile/components/language-preferences-card.tsx`, `web/src/features/profile/components/sidebar-modules-card.tsx`, `web/src/features/profile/components/tabs/notification-tab.tsx`, `web/src/features/profile/hooks/use-profile.ts`, `web/src/features/rankings/hooks/use-rankings.ts`, `web/src/features/redemption-codes/components/__tests__/bulk-delete.test.tsx`, `web/src/features/redemption-codes/components/__tests__/redemptions-mutate-drawer.test.tsx`, `web/src/features/redemption-codes/components/data-table-bulk-actions.tsx`, `web/src/features/redemption-codes/components/data-table-row-actions.tsx`, `web/src/features/redemption-codes/components/redemptions-delete-dialog.tsx`, `web/src/features/redemption-codes/components/redemptions-mutate-drawer.tsx`, `web/src/features/redemption-codes/components/redemptions-primary-buttons.tsx`, `web/src/features/redemption-codes/components/redemptions-table.tsx`, `web/src/features/security/components/__tests__/privacy-card.test.tsx`, `web/src/features/security/components/dialogs/telegram-bind-dialog.tsx`, `web/src/features/security/components/login-sessions-card.tsx`, `web/src/features/security/components/passkey-card.tsx`, `web/src/features/security/components/privacy-card.tsx`, `web/src/features/security/hooks/use-access-token.ts`, `web/src/features/security/hooks/use-account-security.ts`, `web/src/features/security/hooks/use-two-fa-setup.ts`, `web/src/features/setup/setup-wizard.tsx`, `web/src/features/subscriptions/components/dialogs/reset-subscriptions-dialog.tsx`, `web/src/features/subscriptions/components/dialogs/subscription-purchase-dialog.tsx`, `web/src/features/subscriptions/components/dialogs/toggle-status-dialog.tsx`, `web/src/features/subscriptions/components/dialogs/user-subscriptions-dialog.tsx`, `web/src/features/subscriptions/components/subscriptions-mutate-drawer.tsx`, `web/src/features/subscriptions/components/subscriptions-table.tsx`, `web/src/features/system-info/components/system-instances-panel.tsx`, `web/src/features/system-info/components/system-tasks-panel.tsx`, `web/src/features/system-settings/auth/custom-oauth/hooks/use-custom-oauth-mutations.ts`, `web/src/features/system-settings/auth/custom-oauth/hooks/use-custom-oauth-providers.ts`, `web/src/features/system-settings/auth/oauth-section.tsx`, `web/src/features/system-settings/content/announcements-section.tsx`, `web/src/features/system-settings/content/api-info-section.tsx`, `web/src/features/system-settings/content/faq-section.tsx`, `web/src/features/system-settings/content/uptime-kuma-section.tsx`, `web/src/features/system-settings/general/channel-affinity/cache-stats-dialog.tsx`, `web/src/features/system-settings/general/channel-affinity/index.tsx`, `web/src/features/system-settings/hooks/use-system-options.ts`, `web/src/features/system-settings/hooks/use-update-option.ts`, `web/src/features/system-settings/integrations/payment-settings-section.tsx`, `web/src/features/system-settings/integrations/waffo-pancake-settings-section.tsx`, `web/src/features/system-settings/maintenance/log-settings-section.tsx`, `web/src/features/system-settings/maintenance/performance-section.tsx`, `web/src/features/system-settings/maintenance/update-checker-section.tsx`, `web/src/features/system-settings/models/model-ratio-form.tsx`, `web/src/features/system-settings/models/ratio-settings-card.tsx`, `web/src/features/system-settings/models/upstream-ratio-sync.tsx`, `web/src/features/task-plugins/api.ts`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/marketplace-panel.tsx`, `web/src/features/task-plugins/components/marketplace-plugin-card.tsx`, `web/src/features/task-plugins/components/marketplace-sources-dialog.tsx`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/plugin-icon.tsx`, `web/src/features/task-plugins/components/plugin-sandbox.tsx`, `web/src/features/task-plugins/components/plugin-url-import-field.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`, `web/src/features/task-plugins/index.tsx`, `web/src/features/usage-logs/audit/api.ts`, `web/src/features/usage-logs/audit/components/audit-log-viewer.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`, `web/src/features/usage-logs/components/common-logs-stats.tsx`, `web/src/features/usage-logs/components/dialogs/user-info-dialog.tsx`, `web/src/features/usage-logs/components/task-artifacts.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/users/api.ts`, `web/src/features/users/components/data-table-row-actions.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`, `web/src/features/users/components/user-quota-dialog.tsx`, `web/src/features/users/components/users-delete-dialog.tsx`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/features/users/components/users-table.tsx`, `web/src/features/wallet/components/subscription-plans-card.tsx`, `web/src/features/wallet/hooks/use-affiliate.ts`, `web/src/features/wallet/hooks/use-billing-history.ts`, `web/src/features/wallet/hooks/use-creem-payment.ts`, `web/src/features/wallet/hooks/use-payment.ts`, `web/src/features/wallet/hooks/use-redemption.ts`, `web/src/features/wallet/hooks/use-topup-info.ts`, `web/src/features/wallet/hooks/use-waffo-pancake-payment.ts`, `web/src/features/wallet/hooks/use-waffo-payment.ts`, `web/src/hooks/use-notifications.ts`, `web/src/lib/__tests__/server-error-notifications.test.ts`, `web/src/lib/api.ts`, `web/src/lib/handle-server-error.ts`, `web/src/lib/http-client.ts`, `web/src/lib/query-client.ts`, `web/src/lib/secure-verification.ts`, `web/src/lib/server-error-message.ts`, `web/src/main.tsx`, `web/src/routes/(auth)/oauth.tsx`, `web/src/routes/_authenticated/chat2link.tsx`, `web/src/routes/oauth/$provider.tsx`
- Overlap with downstream changes: web/src/features/about/index.tsx, web/src/features/auth/otp/components/otp-form.tsx, web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/channels/api.ts, web/src/features/channels/components/channels-table.tsx, web/src/features/channels/components/dialogs/channel-test-dialog.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/hooks/use-channel-mutate-form.ts, web/src/features/dashboard/components/overview/overview-dashboard.tsx, web/src/features/dashboard/components/overview/summary-cards.tsx, web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/features/system-info/components/system-tasks-panel.tsx, web/src/features/system-settings/content/announcements-section.tsx, web/src/features/system-settings/hooks/use-update-option.ts, web/src/features/system-settings/maintenance/log-settings-section.tsx, web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/features/usage-logs/components/common-logs-stats.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx, web/src/features/users/api.ts, web/src/features/users/components/data-table-row-actions.tsx, web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/features/users/components/users-mutate-drawer.tsx, web/src/features/users/components/users-table.tsx, web/src/lib/http-client.ts, web/src/lib/server-error-message.ts, web/src/routes/oauth/$provider.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 统一服务端错误提示与取消去重
- AI recommendation: `defer` — 174 文件横切变更新增安全消息包装、取消识别和 toast 去重，同时更换大量业务调用。下游还含短信/财务/资产等错误码；先做错误契约与去重策略，后按模块迁移。
- AI risks: 不能把取消操作报成错误，也不能吞掉真实账务/认证失败；受保护 web/AGENTS.md 不照搬。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/lib/handle-server-error.ts:27](https://github.com/QuantumNous/new-api/blob/12be9975c0bf01fa175a2bb3360607767c8ba5fb/web/src/lib/handle-server-error.ts#L27) — 固定源码：const reportedErrors = new WeakSet<object>()
- AI evidence: [downstream:web/src/lib/handle-server-error.ts:25](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/lib/handle-server-error.ts#L25) — 下游对照：export function handleServerError(error: unknown) {

#### commit-a20574136b27

- Title: fix(alibaba): correct Wan model protocols and usage accounting
- Feature/group: fix(alibaba): correct Wan model protocols and usage accounting
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/alibaba_responses_test.go`, `plugins/alibaba_wan3_test.go`, `plugins/tasks/alibaba/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Wan 模型协议及计费用量
- AI recommendation: `defer` — 建立按 Wan 型号区分的能力表，统一原生与兼容协议的尺寸、时长和用量统计。下游已实现 Wan3 Go 专用解析和冻结计费，对输入+输出时长、audio 维度和旧模型仍需逐型号核实，适合独立后端迁移评估。
- AI risks: 此修复不能只搬模型名；计费维度改变会影响存量价格、成本重算和历史任务。
- AI dependencies: commit-eb48396d5fe9, commit-3b4652269a6a
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/alibaba/plugin.js:4](https://github.com/QuantumNous/new-api/blob/a20574136b2746e9afc4b268c99cf99b3f6cc68b/plugins/tasks/alibaba/plugin.js#L4) — 能力表为 Wan3 定义分辨率与最长时长；代码："wan3.0-video": { kind: "all", resolutions: ["480P", "720P", "1080P"], defaultResolution: "1080P", maxDuration: 30 },
- AI evidence: [downstream:relay/channel/task/ali/adaptor.go:560](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/task/ali/adaptor.go#L560) — 下游当前估算由 Go 适配器提供并接冻结账务链；代码：func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {

#### commit-7cf9b473f61e

- Title: fix(web): keep model pricing content in one scroll area
- Feature/group: fix(web): keep model pricing content in one scroll area
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型定价单滚动容器
- AI recommendation: `defer` — 将价格信息、当前计费和编辑表单放入一个滚动区域，依赖本地尚不存在的ModelPricingPanel及新Sheet接口。应随统一定价管理UI评估，不单独复制新页面布局。
- AI risks: 保留移动端可达性、固定保存按钮和已有本地批量调价入口。
- AI dependencies: commit-0c76e4dae77a, commit-0e0ba152bdcc
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/system-settings/models/model-pricing-sheet.tsx:118](https://github.com/QuantumNous/new-api/blob/7cf9b473f61eb7dc09f6f0782192f17595140a08/web/src/features/system-settings/models/model-pricing-sheet.tsx#L118) — Sheet新增滚动头区域；代码：scrollHeader?: ReactNode

#### commit-064ed943e1ac

- Title: feat(billing): support fixed per-request expression pricing
- Feature/group: feat(billing): support fixed per-request expression pricing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `pkg/billingexpr/billingexpr_test.go`, `pkg/billingexpr/compile.go`, `pkg/billingexpr/expr.md`, `pkg/billingexpr/fixed.go`, `pkg/billingexpr/run.go`, `pkg/billingexpr/settle.go`, `pkg/billingexpr/testdata/frontend_simulation.json`, `pkg/billingexpr/types.go`, `relay/helper/price.go`, `relay/helper/price_test.go`, `relay/relay_task.go`, `service/log_info_generate.go`, `service/quota.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/tiered_settle.go`, `setting/billing_setting/tiered_billing.go`, `setting/billing_setting/tiered_billing_test.go`, `web/src/features/pricing/__tests__/breakdown-tier-match.test.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/lib/__tests__/billing-expression.test.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-expression/README.md`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/parser.ts`, `web/src/features/pricing/lib/billing-expression/runtime.ts`, `web/src/features/pricing/lib/billing-expression/types.ts`, `web/src/features/pricing/lib/billing-expression/visual.ts`, `web/src/features/pricing/lib/breakdown-tier-match.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/tier-expr.ts`, `web/src/features/system-settings/models/__tests__/visual-billing-editor.test.tsx`, `web/src/features/system-settings/models/request-simulation.tsx`, `web/src/features/system-settings/models/tier-price-fields.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/features/system-settings/models/visual-billing-document-editor.tsx`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relay/relay_task.go, service/quota.go, service/text_quota.go, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 固定每请求表达式计费
- AI recommendation: `defer` — 新增fixed定价标记、AST校验、快照单位及多入口结算，扩展现有每百万Token表达式契约。该能力可能有价值，但改变真实收费单位，需选择产品策略并证明任务/文本/图片/Realtime入口及历史表达式兼容后单独迁移。
- AI risks: 不能在冻结原始账务和成本证据之外重解释旧表达式；fixed与tier混用、失败/零用量、重试和退款需确定。
- AI dependencies: commit-eb48396d5fe9, commit-d52bdc0b4087
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/billingexpr/fixed.go:12](https://github.com/QuantumNous/new-api/blob/064ed943e1ac40e3eaca1b58ffb7fa5dacb3fde3/pkg/billingexpr/fixed.go#L12) — 引擎识别固定请求定价并供宿主校验入口；代码：func UsesFixedPricing(expression string) bool {
- AI evidence: [downstream:pkg/billingexpr/types.go:49](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/pkg/billingexpr/types.go#L49) — 本地冻结快照尚无收费单位和固定价格字段；代码：type BillingSnapshot struct {

#### commit-bdef11750524

- Title: fix: restore add split button in advanced custom routes (#7289)
- Feature/group: fix: restore add split button in advanced custom routes (#7289)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 高级自定义路由 Add split 可见性
- AI recommendation: `migrate` — 上游把 Add split 从可隐藏的 header 内移至固定 footer；本地该按钮仍在 !hideHeader 分支，内嵌编辑时可能不可达。建议迁入结构修复。
- AI risks: 模型发现专用路由不可分割；保持路由匹配、fallback顺序、键盘可达与本地高级路由语义。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx, web/src/features/channels/components/__tests__/advanced-custom-routes.test.tsx
- AI evidence: [upstream:web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx:1186](https://github.com/QuantumNous/new-api/blob/bdef117505247769268b209665fb3ad7554c3da7/web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx#L1186) — 固定源码：<div className={cn('px-3 py-2', !hideHeader && 'border-t')}>
- AI evidence: [downstream:web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx:1182](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx#L1182) — 下游对照：{!hideHeader ? (

#### commit-b6566f33d908

- Title: feat: clarify passkey website configuration
- Feature/group: feat: clarify passkey website configuration
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/auth/__tests__/passkey-settings.test.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/hooks/use-update-option.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/system-settings/hooks/use-update-option.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Passkey 网站配置说明与状态刷新
- AI recommendation: `defer` — 上游增加网站/域名帮助和状态查询失效处理；建议与安全多 RP 方案一起设计，以免把单 RP 当前配置界面先改成临时形态。
- AI risks: 不得把 RP ID 与完整 Origin 混为一谈；域名变化对已有凭据的影响需显式提示。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/system-settings/hooks/use-update-option.ts:45](https://github.com/QuantumNous/new-api/blob/b6566f33d908fee574c8a6da2a23c4fd5df7353c/web/src/features/system-settings/hooks/use-update-option.ts#L45) — 固定源码：'passkey.rp_id',
- AI evidence: [downstream:setting/system_setting/passkey.go:14](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/setting/system_setting/passkey.go#L14) — 下游对照：RPID                 string `json:"rp_id"`

#### commit-3e1b8b153249

- Title: feat(plugins): display localized changelogs with English fallback
- Feature/group: feat(plugins): display localized changelogs with English fallback
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`, `web/src/components/ui/markdown.tsx`, `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-changelog-panel.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-changelog.test.ts`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/plugin-changelog-panel.tsx`, `web/src/features/task-plugins/lib/plugin-changelog.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件多语言变更记录
- AI recommendation: `defer` — 新增按语言读取 changelog 并回落英语，Markdown 相对链接按来源解析。当前没有插件市场页面，不引入只服务未启用能力的 YAML/Markdown 扩展；待市场方案确定再处理。
- AI risks: 外部 Markdown 链接需要安全协议限制；语言回落不得执行源码。
- AI dependencies: commit-eb48396d5fe9, commit-984330920061
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/components/ui/markdown.tsx:723](https://github.com/QuantumNous/new-api/blob/3e1b8b15324941ff20c77616737d8dd984582d33/web/src/components/ui/markdown.tsx#L723) — 相对链接根据插件来源解析后限制协议；代码：function addExternalLinkAttributes(html: string, baseUrl?: string): string {

#### commit-3cea2bf799a6

- Title: feat(relaykit): preserve cached input token breakdowns
- Feature/group: feat(relaykit): preserve cached input token breakdowns
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/dto/billing_usage.go`, `relaykit/dto/billing_usage_test.go`, `relaykit/dto/openai_response.go`, `relaykit/dto/usage_merge.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 缓存输入 token 模态明细
- AI recommendation: `migrate` — 本地InputTokenDetails只有标量缓存计数，缺cached_tokens_details；上游增加可区分缺失/显式零的指针明细并深克隆。建议仅迁移DTO及合并契约，适配本地旧usage链，不能直接假定新usage_merge文件已存在。
- AI risks: 不得把缓存模态子集再加到总缓存重复计费；本地Moonshot明细、财务快照和标准/表达式计费需一致。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/dto/openai_response.go, relaykit/dto/billing_usage.go, relaykit/dto/billing_usage_test.go, relaykit/dto/usage_merge.go, relaykit/dto/usage_merge_test.go
- AI evidence: [upstream:relaykit/dto/openai_response.go:263](https://github.com/QuantumNous/new-api/blob/3cea2bf799a61e07ad53368538fdb7064b017d77/relaykit/dto/openai_response.go#L263) — 上游增加可选缓存模态明细。
- AI evidence: [downstream:relaykit/dto/openai_response.go:256](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/dto/openai_response.go#L256) — 本地对应结构尚无缓存模态明细。

#### commit-f362c7c51b86

- Title: feat(billingexpr): support image cache and quantity variables
- Feature/group: feat(billingexpr): support image cache and quantity variables
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `pkg/billingexpr/billingexpr_test.go`, `pkg/billingexpr/compile.go`, `pkg/billingexpr/expr.md`, `pkg/billingexpr/fixed.go`, `pkg/billingexpr/run.go`, `pkg/billingexpr/settle.go`, `pkg/billingexpr/testdata/frontend_simulation.json`, `pkg/billingexpr/types.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 图片缓存和图片数量表达式变量
- AI recommendation: `defer` — 扩展img_cr/image_count等变量并按表达式使用情况拆分缓存及图片数量，依赖fixed和新版快照。必须保留当前未引用类别归入p/c的计费真值原则；待计费语言升级明确后整体迁移。
- AI risks: 缓存文本/图片重叠可能双计；数量与Token价不能重复乘n；旧表达式快照需向后兼容。
- AI dependencies: commit-064ed943e1ac
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/billingexpr/types.go:30](https://github.com/QuantumNous/new-api/blob/f362c7c51b86a97e1fc72e4d48280cf53a5d2cb9/pkg/billingexpr/types.go#L30) — Token参数新增图片缓存维度；代码：ImgCR float64 // image cache read tokens, separated only when explicitly priced
- AI evidence: [downstream:pkg/billingexpr/types.go:26](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/pkg/billingexpr/types.go#L26) — 本地已有图片输出变量，但尚无图片缓存和数量语义；代码：ImgO float64 // image output tokens

#### commit-f064bffa2b65

- Title: fix(billing): validate image quantities before reserving quota
- Feature/group: fix(billing): validate image quantities before reserving quota
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/model.go`, `common/model_test.go`, `relay/channel/ali/adaptor_test.go`, `relay/channel/ali/dto.go`, `relay/channel/ali/image.go`, `relay/channel/ali/image_wan.go`, `relay/channel/openai/image_stream_test.go`, `relay/channel/openai/relay_image.go`, `relay/common/relay_info.go`, `relay/convert_request_error_test.go`, `relay/helper/billing_expr_request.go`, `relay/helper/openai_image_request_test.go`, `relay/helper/price.go`, `relay/helper/valid_request.go`, `relay/image_handler.go`, `relaykit/dto/legacy_dalle_image.go`, `relaykit/dto/openai_image.go`, `service/billing_session.go`, `service/image_billing.go`, `service/log_info_generate.go`, `service/quota.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/tiered_settle.go`, `service/tiered_settle_test.go`
- Overlap with downstream changes: relay/common/relay_info.go, service/billing_session.go, service/quota.go, service/text_quota.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 图片数量校验与每次重试预留
- AI recommendation: `migrate` — 建议抽出安全修复迁移到现有计费引擎：在请求验证和每次上游尝试前确定有效n、parameters.n与prompt_extend，并补足预留。下游虽已有最大n和适配器侧上限，Ali额外参数仍在预扣之后改变倍率，不能认为整项已覆盖。本批不顺带启用fixed/img_cr/image_count语言升级。
- AI risks: 避免重复乘n和跨渠道重试倍率串用；数量0/超大uint、multipart额外参数、额度不足补预留、零价转付费及冻结成本证据都需测试。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/dto/openai_image.go, relaykit/dto/legacy_dalle_image.go, relay/channel/ali/dto.go, relay/channel/ali/image.go, relay/channel/ali/image_wan.go, relay/channel/ali/adaptor_test.go, relay/channel/openai/relay_image.go, relay/channel/openai/image_stream_test.go, relay/common/relay_info.go, relay/helper/billing_expr_request.go, relay/helper/openai_image_request_test.go, relay/helper/price.go, relay/helper/price_test.go, relay/helper/valid_request.go, relay/image_handler.go, service/billing_session.go, service/billing_session_test.go, service/image_billing.go, service/image_billing_test.go, service/quota.go, service/text_quota.go, service/text_quota_test.go, service/tiered_settle.go, service/tiered_settle_test.go, service/log_info_generate.go
- AI evidence: [upstream:service/image_billing.go:21](https://github.com/QuantumNous/new-api/blob/f064bffa2b65dd2c6348b375f8b91bab41aabc19/service/image_billing.go#L21) — 以实际发送图片数量在每次尝试前预留；代码：func PrepareImageBillingForRequest(c *gin.Context, info *relaycommon.RelayInfo, count int, promptExtend bool) *types.NewAPIError {
- AI evidence: [downstream:relay/channel/ali/image.go:64](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/ali/image.go#L64) — 本地Ali在适配器构造时才应用额外数量，现有上限不等于预留同步；代码：info.PriceData.AddOtherRatio("n", float64(imageRequest.Parameters.N))

#### commit-f256e40bcb29

- Title: feat(billing): add expression defaults for GPT image models
- Feature/group: feat(billing): add expression defaults for GPT image models
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `setting/billing_setting/builtin_billing.go`, `setting/billing_setting/builtin_billing_test.go`, `setting/billing_setting/tiered_billing.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: GPT图片默认表达式售价
- AI recommendation: `defer` — 为GPT图片内置文字/图片缓存及输出价格，并保留管理员覆盖。依赖未采用的内置默认价格和img_cr变量；必须先确认本平台销售策略及来源价格，不能由同步默认启用真实收费。
- AI risks: 上游代码中的价目并不等于已核验的当前供应商价格；历史销售与成本策略需隔离。
- AI dependencies: commit-eb99ab1b4034, commit-f362c7c51b86
- AI suggested paths (not approval):
- AI evidence: [upstream:setting/billing_setting/builtin_billing.go:8](https://github.com/QuantumNous/new-api/blob/f256e40bcb29292376edb5f0ba2e74ffa463aced/setting/billing_setting/builtin_billing.go#L8) — 默认表达式分别计文本缓存和图片缓存；代码："gpt-image-2":            `tier("standard", p * 5 + cr * 1.25 + img * 8 + img_cr * 2 + c * 30)`,

#### commit-25ec832faa48

- Title: feat(pricing): convert legacy prices into expression drafts
- Feature/group: feat(pricing): convert legacy prices into expression drafts
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/model_management_test.go`, `controller/model_pricing_config.go`, `model/legacy_dalle_pricing.go`, `model/model_pricing_config.go`, `model/model_pricing_conversion.go`, `router/api-router.go`, `setting/ratio_setting/cache_ratio.go`, `setting/ratio_setting/model_ratio.go`
- Overlap with downstream changes: router/api-router.go, setting/ratio_setting/model_ratio.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 旧价格转换为表达式草稿
- AI recommendation: `defer` — 提供预览接口把旧比例/按次价格推导为表达式草稿，保留旧DALL-E兼容规则。该转换应显式预览并经人工确认，不能在升级中自动写入；依赖新模型定价与固定/图片数量语义。
- AI risks: 价格转换无法无损推导的模型必须拒绝，别名、零值、缓存写入与通道协议需证据。
- AI dependencies: commit-0c76e4dae77a, commit-064ed943e1ac, commit-f362c7c51b86
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/model_pricing_config.go:21](https://github.com/QuantumNous/new-api/blob/25ec832faa48031dd74f6fff0f58390e29a79f7c/controller/model_pricing_config.go#L21) — 新增只预览的旧价格转换入口；代码：func PreviewModelPricingConversion(c *gin.Context) {

#### commit-39294418af4c

- Title: feat(web): extend expression pricing editors and usage details
- Feature/group: feat(web): extend expression pricing editors and usage details
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/task-matrix.test.ts`, `web/src/features/pricing/components/cached-price-cell.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/lib/__tests__/billing-expression.test.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-expression/README.md`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/parser.ts`, `web/src/features/pricing/lib/billing-expression/runtime.ts`, `web/src/features/pricing/lib/billing-expression/types.ts`, `web/src/features/pricing/lib/billing-expression/visual.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/task-expr.ts`, `web/src/features/pricing/lib/tier-expr.ts`, `web/src/features/system-settings/models/__tests__/request-simulation.test.tsx`, `web/src/features/system-settings/models/__tests__/task-pricing-copy.test.tsx`, `web/src/features/system-settings/models/__tests__/visual-billing-editor.test.tsx`, `web/src/features/system-settings/models/request-simulation.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/models/tier-price-fields.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/features/system-settings/models/visual-billing-document-editor.tsx`, `web/src/features/usage-logs/components/__tests__/usage-facts.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 表达式价格编辑与用量详情
- AI recommendation: `defer` — 前端增加图片缓存、按张/按请求单位、复杂表达式解析及日志明细，与后端fixed和img_cr升级成套。当前本地动态价格格式较旧，不应只更新展示让用户误以为可计费。
- AI risks: 前后端单位不一致会造成错误报价；历史计价证据和下游调账投影必须保留。
- AI dependencies: commit-d52bdc0b4087, commit-064ed943e1ac, commit-f362c7c51b86
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/pricing/components/cached-price-cell.tsx:70](https://github.com/QuantumNous/new-api/blob/39294418af4c920950be1b8c1fed0ed0d2bea52e/web/src/features/pricing/components/cached-price-cell.tsx#L70) — 缓存价区分文本和图片缓存；代码：entry.field === 'cacheReadPrice' || entry.field === 'imageCachePrice'

#### commit-b3e279464c82

- Title: feat(web): review legacy pricing conversion before applying drafts
- Feature/group: feat(web): review legacy pricing conversion before applying drafts
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/model-pricing/__tests__/editor-currency.test.tsx`, `web/src/features/model-pricing/__tests__/editor-layout.test.tsx`, `web/src/features/model-pricing/api.ts`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/model-pricing/pricing-conversion-dialog.tsx`, `web/src/features/model-pricing/pricing.ts`, `web/src/features/system-settings/models/model-pricing-core.ts`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-pricing-snapshots.ts`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 旧价格转换人工复核
- AI recommendation: `defer` — 增加旧配置与转换草稿对照、差异说明和确认应用，是旧价转换的必要交互。当前无对应转换API和新定价面板，应随98一起评估，禁止自动替换存量售价。
- AI risks: 批量转换应有乐观版本与审计，不能修改既有冻结账单或历史成本。
- AI dependencies: commit-0c76e4dae77a, commit-75e533209490, commit-25ec832faa48, commit-39294418af4c
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/model-pricing/pricing-conversion-dialog.tsx:44](https://github.com/QuantumNous/new-api/blob/b3e279464c82d6b1bd9fca6ce27f67c4c9246fec/web/src/features/model-pricing/pricing-conversion-dialog.tsx#L44) — 新增独立转换复核对话框；代码：export type PricingConversionPreview = {

#### commit-505805a4c3ab

- Title: feat(web): unify channel setup and refine model and settings editors
- Feature/group: feat(web): unify channel setup and refine model and settings editors
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`, `web/src/components/multi-select.tsx`, `web/src/components/multi-select/__tests__/selection.test.tsx`, `web/src/components/ui/__tests__/label.test.tsx`, `web/src/components/ui/label.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-editor.test.tsx`, `web/src/features/channels/components/__tests__/provider-selection.test.tsx`, `web/src/features/channels/components/__tests__/upstream-model-selection.test.tsx`, `web/src/features/channels/components/dialogs/configure-models-dialog.tsx`, `web/src/features/channels/components/dialogs/fetch-models-dialog.tsx`, `web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx`, `web/src/features/channels/components/drawers/channel-configuration.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/drawers/channel-provider-picker.tsx`, `web/src/features/channels/components/model-mapping-editor.tsx`, `web/src/features/channels/components/upstream-model-selection.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/hooks/use-channel-model-discovery.ts`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/keys/components/dialogs/cc-switch-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-mutate-dialog.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/system-settings/auth/__tests__/passkey-settings.test.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/general/channel-affinity/rule-editor-dialog.tsx`, `web/src/features/system-settings/integrations/waffo-settings-section.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道配置工作流与模型编辑重组
- AI recommendation: `defer` — 37 文件把渠道编辑拆成连接/模型/路由/请求等阶段，并涉及上游 task plugin 描述。本地下游有 Volc/Kimi、原生任务与凭据特殊编辑，不宜整页替换。
- AI risks: 需逐项保留敏感字段权限、专有渠道参数、历史配置兼容与提交payload；先完成渠道能力矩阵。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/channels/components/drawers/channel-configuration.tsx:39](https://github.com/QuantumNous/new-api/blob/505805a4c3ab55fbcf6227ec110226ff1c193e0b/web/src/features/channels/components/drawers/channel-configuration.tsx#L39) — 固定源码：statuses: Record<ChannelConfigurationSection, ChannelConfigurationStatus>
- AI evidence: [downstream:web/src/features/channels/components/drawers/channel-mutate-drawer.tsx:621](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/drawers/channel-mutate-drawer.tsx#L621) — 下游对照：export function ChannelMutateDrawer({

#### commit-d4c26bfb8dcb

- Title: fix(channels): allow editing multi-key selection strategy
- Feature/group: fix(channels): allow editing multi-key selection strategy
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/hooks/use-channel-mutate-form.ts`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/hooks/use-channel-mutate-form.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 已有多 key 渠道选择策略可编辑
- AI recommendation: `migrate` — 上游在已有多 key 渠道中显示 multi_key_type 并把 multi_key_mode 送入更新。本地下游仍仅在新建 multi_to_single 时显示对应字段，建议小范围适配。
- AI risks: 保留敏感字段编辑权限，修改轮询/随机策略不能覆盖已存 key 或泄漏；前后端更新契约需回归。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/hooks/use-channel-mutate-form.ts, web/src/features/channels/hooks/__tests__/multi-key-mode.test.ts
- AI evidence: [upstream:web/src/features/channels/hooks/use-channel-mutate-form.ts:98](https://github.com/QuantumNous/new-api/blob/d4c26bfb8dcb7434a0bc1e0d6153e848d86c39b4/web/src/features/channels/hooks/use-channel-mutate-form.ts#L98) — 固定源码：? { multi_key_mode: data.multi_key_type }
- AI evidence: [downstream:web/src/features/channels/components/drawers/channel-mutate-drawer.tsx:3228](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/drawers/channel-mutate-drawer.tsx#L3228) — 下游对照：multiKeyMode === 'multi_to_single' && (

#### commit-ab489ab8842b

- Title: feat(channels): add plugin extensions to provider and model selection
- Feature/group: feat(channels): add plugin extensions to provider and model selection
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/provider-selection.test.tsx`, `web/src/features/channels/components/channel-plugin-extensions.tsx`, `web/src/features/channels/components/dialogs/configure-models-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/drawers/channel-provider-picker.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/channel-plugin-extensions.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道供应商选择中的插件扩展
- AI recommendation: `defer` — 在现有供应商选择器中列出插件支持的渠道类型与模型扩展，依赖插件元数据与渠道绑定。应跟随插件渠道迁移，保留 NovaMaaS 渠道类型及本地参数。
- AI risks: 选项重排不能改变既有渠道数值 ID、隔离和权限。
- AI dependencies: commit-eb48396d5fe9, commit-99974a814f00, commit-eb76b136b85e
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/channels/components/channel-plugin-extensions.tsx:28](https://github.com/QuantumNous/new-api/blob/ab489ab8842b26f3dbc2d7a384e7c0ab0f700e0e/web/src/features/channels/components/channel-plugin-extensions.tsx#L28) — 新增渠道插件扩展组件；代码：export function ChannelPluginExtensions(props: {

#### commit-251b76d8633e

- Title: feat(web): add administrator update reminders
- Feature/group: feat(web): add administrator update reminders
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/layout/components/app-header.tsx`, `web/src/components/layout/components/public-header.tsx`, `web/src/components/layout/components/system-brand.tsx`, `web/src/features/system-settings/maintenance/update-checker-section.tsx`, `web/src/features/system-update/__tests__/releases.test.ts`, `web/src/features/system-update/__tests__/update-checking.test.tsx`, `web/src/features/system-update/api.ts`, `web/src/features/system-update/releases.ts`, `web/src/features/system-update/store.ts`, `web/src/features/system-update/system-update-action.tsx`, `web/src/features/system-update/system-update-dialog.tsx`, `web/src/features/system-update/use-system-update.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/components/layout/components/public-header.tsx, web/src/components/layout/components/system-brand.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 管理员版本更新提醒
- AI recommendation: `defer` — 上游提醒直接读取QuantumNous/new-api Releases，版本解析也不接受-novamaas后缀。该机制不能作为下游产品更新状态，更不能把部分上游同步提示成已完成升级；需要区分本仓库发布与上游评估目标后另设计。
- AI risks: 保留上游归属信息，但提醒来源和版本判断必须符合fork语义；不得创建平台上游同步操作页。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/system-update/api.ts:45](https://github.com/QuantumNous/new-api/blob/251b76d8633ec4eeb1064dcdb5eb806ecf1a8668/web/src/features/system-update/api.ts#L45) — 提醒固定读取上游发布来源；代码：'https://api.github.com/repos/QuantumNous/new-api/releases?per_page=100',
- AI evidence: [downstream:VERSION:1](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/VERSION#L1) — 下游使用独立novamaas版本号而非单一上游标签；代码：v1.0.0-rc.26-novamaas.2

#### commit-74629e29f83f

- Title: feat(plugins): enhance task streaming and model pricing
- Feature/group: feat(plugins): enhance task streaming and model pricing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `controller/billing_option_test.go`, `controller/model_management_test.go`, `controller/option.go`, `controller/relay.go`, `controller/relay_task_plugin_test.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `dto/channel_constraints.go`, `dto/task.go`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `middleware/distributor.go`, `middleware/distributor_test.go`, `middleware/task_plugin.go`, `middleware/task_plugin_test.go`, `model/channel_constraint.go`, `model/model_pricing_config.go`, `model/pricing.go`, `model/pricing_usage_schema_test.go`, `model/task_plugin_channel_select_test.go`, `pkg/billingexpr/expr.md`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/json_state.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`, `pkg/jsplugin/utils.go`, `plugins/alibaba_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/performance_test.go`, `relay/channel/task/jsplugin/submit_stream.go`, `relay/helper/stream_scanner.go`, `relay/helper/stream_scanner_test.go`, `relay/relay_task.go`, `relay/relay_task_test.go`, `service/channel_select.go`, `service/channel_select_test.go`, `service/task_billing.go`, `service/task_polling.go`, `setting/billing_setting/tiered_billing.go`, `web/src/features/model-pricing/__tests__/editor-layout.test.tsx`, `web/src/features/model-pricing/__tests__/pricing.test.ts`, `web/src/features/model-pricing/api.ts`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/model-pricing/pricing.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/cached-price-cell.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/plugin-pricing.ts`, `web/src/features/pricing/lib/task-price-display.ts`, `web/src/features/pricing/types.ts`, `web/src/features/system-settings/billing/index.tsx`, `web/src/features/system-settings/billing/section-registry.tsx`, `web/src/features/system-settings/models/__tests__/plugin-pricing.test.tsx`, `web/src/features/system-settings/models/__tests__/task-pricing-copy.test.tsx`, `web/src/features/system-settings/models/index.tsx`, `web/src/features/system-settings/models/model-pricing-core.ts`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-pricing-snapshots.ts`, `web/src/features/system-settings/models/model-ratio-form.tsx`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/ratio-settings-card.tsx`, `web/src/features/system-settings/models/task-plugin-pricing-editor.tsx`, `web/src/features/system-settings/models/task-pricing-matrix.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/task-plugins/__tests__/plugin-detail-sheet.test.tsx`, `web/src/features/task-plugins/__tests__/usage-schema-table.test.tsx`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/usage-schema-table.tsx`, `web/src/features/task-plugins/types.ts`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: AGENTS.md, controller/option.go, controller/relay.go, dto/task.go, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, middleware/distributor.go, relay/relay_task.go, service/task_billing.go, service/task_polling.go, web/src/features/system-settings/billing/section-registry.tsx, web/src/features/system-settings/models/index.tsx, web/src/features/system-settings/types.ts, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 任务提交流与插件独立定价
- AI recommendation: `defer` — 新增提交 SSE 增量、任务流式响应、插件维度表达式和共享模型定价，对约 96 文件的路由、用量、账务及前端同时改造。当前任务计费含成本快照和调账证据，必须先定义冻结契约及中断退款行为。
- AI risks: 同名模型多插件不能误取价格；断流/提交成功后退款和历史冻结表达式必须正确。
- AI dependencies: commit-eb48396d5fe9, commit-6c22550ea325, commit-064ed943e1ac
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/option.go:371](https://github.com/QuantumNous/new-api/blob/74629e29f83f506bb523c5542aa851947bd423eb/controller/option.go#L371) — 新增按插件命名空间存储计费表达式；代码：billing_setting.PluginBillingExprOption: variants,

#### commit-385d2dfd10d8

- Title: feat(auth): add safe multi-RP ID passkey support
- Feature/group: feat(auth): add safe multi-RP ID passkey support
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/audit.go`, `controller/login_verification.go`, `controller/misc.go`, `controller/option.go`, `controller/passkey.go`, `controller/passkey_test.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `model/option.go`, `model/passkey.go`, `model/passkey_option.go`, `model/user_authentication_test.go`, `router/api-router.go`, `service/account_security.go`, `service/passkey/service.go`, `service/passkey/session.go`, `service/security_verification.go`, `setting/system_setting/passkey.go`, `web/src/features/auth/passkey/__tests__/domain-verification.test.tsx`, `web/src/features/auth/passkey/api.ts`, `web/src/features/auth/passkey/assertion.ts`, `web/src/features/auth/passkey/components/passkey-domain-selector.tsx`, `web/src/features/auth/passkey/types.ts`, `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/api.ts`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/auth/sign-in/components/user-auth-form.tsx`, `web/src/features/system-settings/api.ts`, `web/src/features/system-settings/auth/__tests__/passkey-settings.test.tsx`, `web/src/features/system-settings/auth/index.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/hooks/use-update-option.ts`, `web/src/features/system-settings/types.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/lib/server-error-message.ts`
- Overlap with downstream changes: controller/audit.go, controller/misc.go, controller/option.go, controller/secure_verification.go, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, model/option.go, model/user_authentication_test.go, router/api-router.go, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/auth/secure-verification/types.ts, web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/system-settings/api.ts, web/src/features/system-settings/hooks/use-update-option.ts, web/src/features/system-settings/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/lib/server-error-message.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 多 RP ID Passkey 及旧凭据域名保护
- AI recommendation: `defer` — 上游新增凭据 RPID、兼容域名、删除影响预览及确认；本地只有单 RPID，并从 ServerAddress 推导。此能力有价值，但应连同凭据迁移、域名移除和后续修复成组实施。
- AI risks: RP/Origin 选择、旧 NULL RPID 凭据、域名变更并发与多数据库迁移必须有确定性回归。
- AI dependencies: commit-6f2333990613, commit-b6566f33d908
- AI suggested paths (not approval):
- AI evidence: [upstream:model/passkey.go:26](https://github.com/QuantumNous/new-api/blob/385d2dfd10d821b25c8a6766bd16eea248cb1652/model/passkey.go#L26) — 固定源码：RPID            *string        `json:"rp_id,omitempty" gorm:"column:rp_id;type:varchar(253)"`
- AI evidence: [downstream:setting/system_setting/passkey.go:14](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/setting/system_setting/passkey.go#L14) — 下游对照：RPID                 string `json:"rp_id"`

#### commit-129f21b6942e

- Title: fix(plugins): normalize invalid UTF-8 in JSON state
- Feature/group: fix(plugins): normalize invalid UTF-8 in JSON state
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/json_state.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件 JSON 状态无效 UTF-8
- AI recommendation: `defer` — 对跨 JS 边界的字符串做 UTF-8 标准化，保证 JSON 克隆及比较稳定。当前无插件 JSON 状态边界，不应单独添加新模块；若启用插件需纳入基础可靠性。
- AI risks: 需验证字符串归一化不会改变被签名原文或跨请求共享可变状态。
- AI dependencies: commit-eb48396d5fe9, commit-74629e29f83f
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/json_state.go:130](https://github.com/QuantumNous/new-api/blob/129f21b6942ed433fd7eea1239b4a24d8b58b48b/pkg/jsplugin/json_state.go#L130) — 插件状态检查非 UTF-8 字符串；代码：if !utf8.ValidString(key) || len(key) > budget.bytes {

#### commit-c9a110190c52

- Title: fix(channels): show built-in base URLs as placeholders
- Feature/group: fix(channels): show built-in base URLs as placeholders
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `constant/channel.go`, `controller/channel.go`, `controller/channel_test_internal_test.go`, `router/channel-router.go`, `router/channel_router_test.go`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/channel-actions.ts`, `web/src/features/channels/lib/channel-type-config.ts`
- Overlap with downstream changes: constant/channel.go, controller/channel.go, router/channel-router.go, router/channel_router_test.go, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/channel-type-config.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道内置 Base URL 占位提示
- AI recommendation: `migrate` — 上游增加仅列出内置默认地址的只读接口，表单留空时显示提示。本地下游扩展渠道类型和地址策略，应读取本地 ChannelBaseURLs 并沿用 ChannelRead 权限。
- AI risks: 占位值不能写回覆盖用户自定义baseURL；Moonshot路径、渠道编号和未提供默认地址需保留。
- AI dependencies: none identified
- AI suggested paths (not approval): controller/channel.go, router/channel-router.go, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/__tests__/default-base-urls.test.tsx
- AI evidence: [upstream:controller/channel.go:102](https://github.com/QuantumNous/new-api/blob/c9a110190c5241c24d9d66de340431f5e9873db6/controller/channel.go#L102) — 固定源码：func GetChannelDefaultBaseURLs(c *gin.Context) {
- AI evidence: [downstream:constant/channel.go:69](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/constant/channel.go#L69) — 下游对照：var ChannelBaseURLs = []string{

#### commit-be36cbb8facf

- Title: test(security): provide query client for passkey verification tests
- Feature/group: test(security): provide query client for passkey verification tests
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/security/__tests__/account-security.test.tsx`, `web/src/features/security/__tests__/enrollment.test.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 多 RP 验证测试 QueryClient 修复
- AI recommendation: `defer` — 只为已扩展的 Passkey 验证测试补 QueryClientProvider；本地下游没有同版本安全页测试，不应为不存在的测试引入样板。
- AI risks: 迁入多 RP 后应在真实现有 fixture 中补 QueryClient，不能用空测试冒充兼容。
- AI dependencies: commit-385d2dfd10d8
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx:19](https://github.com/QuantumNous/new-api/blob/be36cbb8facf99d523197f9e18d02b7605331048/web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx#L19) — 固定源码：import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
- AI evidence: [downstream:setting/system_setting/passkey.go:14](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/setting/system_setting/passkey.go#L14) — 下游对照：RPID                 string `json:"rp_id"`

#### commit-007d69942d24

- Title: fix(model): migrate renamed prefill group unique indexes
- Feature/group: fix(model): migrate renamed prefill group unique indexes
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/prefill_group_migration.go`, `model/prefill_group_migration_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 按定义迁移旧预设组索引
- AI recommendation: `migrate` — 建议和2b6f1dfefb共同移植最终实现：识别唯一性定义而非固定名字，在锁定事务中先去除碰撞对象再建目标索引。下游PrefillGroup已声明软删除下的唯一性，历史PG库的重命名约束需要此兼容。
- AI risks: 保留复合、表达式和部分索引；目标同名不同定义、外键依赖、重启和失败回滚必须验证。
- AI dependencies: commit-2b6f1dfefbe2, commit-9a8674425c5a
- AI suggested paths (not approval): model/prefill_group_migration.go, model/prefill_group_migration_test.go
- AI evidence: [upstream:model/prefill_group_migration.go:103](https://github.com/QuantumNous/new-api/blob/007d69942d24eeea9edd4665f812d07ed11b3b97/model/prefill_group_migration.go#L103) — 旧对象按定义识别，不再只接受固定遗留名；代码：// before AutoMigrate inspects the column. Match the definition rather than the
- AI evidence: [downstream:model/prefill_group.go:78](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/prefill_group.go#L78) — 本地目标模型为软删除条件下的唯一索引；代码：Name        string         `json:"name" gorm:"size:64;not null;uniqueIndex:uk_prefill_name,where:deleted_at IS NULL"`

#### commit-2ba61576146f

- Title: feat(web): share collapsible mobile table filters
- Feature/group: feat(web): share collapsible mobile table filters
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/data-table/index.ts`, `web/src/components/data-table/toolbar/__tests__/mobile-filter.test.tsx`, `web/src/components/data-table/toolbar/mobile-filter-panel.tsx`, `web/src/components/data-table/toolbar/toolbar.tsx`, `web/src/features/channels/components/channels-table.tsx`, `web/src/features/usage-logs/components/logs-filter-toolbar.tsx`
- Overlap with downstream changes: web/src/features/channels/components/channels-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 移动表格筛选折叠共享组件
- AI recommendation: `defer` — 新增 MobileFilterPanel 并同时改渠道/日志工具栏。下游日志和渠道工具栏有额外操作，建议待移动日志布局和渠道编辑方案后一起评审。
- AI risks: 折叠后已生效过滤器仍需可发现，主操作、筛选清空和键盘焦点不能消失。
- AI dependencies: commit-a5e41a893379, commit-505805a4c3ab
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/components/data-table/toolbar/toolbar.tsx:85](https://github.com/QuantumNous/new-api/blob/2ba61576146f0583f789ee6845f2280cb86819ce/web/src/components/data-table/toolbar/toolbar.tsx#L85) — 固定源码：collapsibleOnMobile?: boolean
- AI evidence: [downstream:web/src/components/data-table/toolbar/toolbar.tsx:154](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/components/data-table/toolbar/toolbar.tsx#L154) — 下游对照：export function DataTableToolbar<TData>(props: DataTableToolbarProps<TData>) {

#### commit-043ff99a51ec

- Title: fix: handle legacy database constraints and scoped policies
- Feature/group: fix: handle legacy database constraints and scoped policies
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `model/migration_dialector.go`, `model/migration_dialector_test.go`, `service/authz/adapter.go`, `service/authz/authz_test.go`, `service/authz/seed.go`
- Overlap with downstream changes: service/authz/authz_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 旧约束及历史授权作用域兼容
- AI recommendation: `migrate` — 建议拆解迁入两部分：PG旧单列约束按真实目录名事务处理；授权策略结构化加载并对不支持的own/other范围显式deny，同时重建内置策略时保留受限行。下游仍按字符串拼接加载且重置删除所有内置角色行，存在范围丢失/意外放权风险。
- AI risks: 不能只跳过受限行从而回落管理员广泛授权；保留NovaMaaS财务和素材资源权限，约束外键失败须回滚。
- AI dependencies: commit-9a8674425c5a, commit-007d69942d24
- AI suggested paths (not approval): model/migration_dialector.go, model/migration_dialector_test.go, service/authz/adapter.go, service/authz/authz_test.go, service/authz/seed.go
- AI evidence: [upstream:service/authz/adapter.go:46](https://github.com/QuantumNous/new-api/blob/043ff99a51ecad8229389ddd04f45f4b25a23ac6/service/authz/adapter.go#L46) — 不支持的遗留范围在内存中显式拒绝；代码：effect = EffectDeny
- AI evidence: [downstream:service/authz/seed.go:42](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/authz/seed.go#L42) — 下游内置角色重建目前没有保留旧范围条件；代码：return db.Where("ptype = ? AND v0 IN ?", "p", subjects).Delete(&model.CasbinRule{}).Error
- AI evidence: [downstream:service/authz/adapter.go:27](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/authz/adapter.go#L27) — 下游加载策略仍通过字符串拼接，不保留结构化字段语义；代码：if err := persist.LoadPolicyLine(ruleToLine(rule), m); err != nil {

#### commit-33142f0aee7c

- Title: fix: refine sign-in page behavior
- Feature/group: fix: refine sign-in page behavior
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/misc.go`, `controller/passkey_test.go`, `web/src/features/auth/passkey/__tests__/domain-verification.test.tsx`, `web/src/features/auth/passkey/components/passkey-domain-selector.tsx`
- Overlap with downstream changes: controller/misc.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 登录页域名选择与公开 Origin 收敛
- AI recommendation: `defer` — 删除多 RP 选择器的外站链接和 status 中 passkey_origins；本地为单 RP 且可能有现存消费者。需在多 RP 升级中核查公开字段兼容后统一落地。
- AI risks: 避免无意暴露内部域名，同时不能直接破坏旧 Passkey 配置页面。
- AI dependencies: commit-385d2dfd10d8
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/auth/passkey/components/passkey-domain-selector.tsx:61](https://github.com/QuantumNous/new-api/blob/33142f0aee7c87b2110f25dbddcb861dbdc2c44e/web/src/features/auth/passkey/components/passkey-domain-selector.tsx#L61) — 固定源码：if (available.length < 2) return null
- AI evidence: [downstream:setting/system_setting/passkey.go:14](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/setting/system_setting/passkey.go#L14) — 下游对照：RPID                 string `json:"rp_id"`

#### commit-7fd06381976c

- Title: fix(auth): restore saved compatible passkey domains
- Feature/group: fix(auth): restore saved compatible passkey domains
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/auth/__tests__/passkey-settings.test.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/auth/section-registry.tsx`
- Overlap with downstream changes: web/src/features/system-settings/auth/section-registry.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 兼容 Passkey 域名保存恢复
- AI recommendation: `defer` — 把 legacy_rp_ids 加入设置 section-registry 默认值，修复新增多 RP 后保存值丢失；本地尚无该配置字段。
- AI risks: 应与多 RP 主项一起测试保存、刷新、再次编辑和删除确认。
- AI dependencies: commit-385d2dfd10d8, commit-33142f0aee7c
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/system-settings/auth/section-registry.tsx:91](https://github.com/QuantumNous/new-api/blob/7fd06381976cbea126d1f20973548bf60f20e37d/web/src/features/system-settings/auth/section-registry.tsx#L91) — 固定源码：'passkey.legacy_rp_ids': settings['passkey.legacy_rp_ids'],
- AI evidence: [downstream:setting/system_setting/passkey.go:14](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/setting/system_setting/passkey.go#L14) — 下游对照：RPID                 string `json:"rp_id"`

#### commit-815217ba648f

- Title: test(web): stabilize flaky frontend suite timeouts and motion visibility races (#7367)
- Feature/group: test(web): stabilize flaky frontend suite timeouts and motion visibility races (#7367)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/test-setup.ts`, `web/vitest.config.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 前端测试超时与减少动画稳定性
- AI recommendation: `defer` — 上游增加 asyncUtilTimeout/testTimeout 并让 matchMedia 命中reduced-motion；本地下游当前测试通过，不能仅按上游超时数字掩盖慢测或改动画分支。建议发生可复现CI波动后迁入适用fixture。
- AI risks: 测试应验证真实行为，减少动画不能屏蔽依赖实际过渡的回归；保留失败定位能力。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/test-setup.ts:29](https://github.com/QuantumNous/new-api/blob/815217ba648ff800c4f583fbfa36ba286cee6b2d/web/src/test-setup.ts#L29) — 固定源码：configure({ asyncUtilTimeout: 5000 })
- AI evidence: [downstream:web/src/test-setup.ts:41](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/test-setup.ts#L41) — 下游对照：Object.defineProperty(window, 'matchMedia', {

#### commit-d92612038410

- Title: fix: unify model provider detection and add Wan icon (#7373)
- Feature/group: fix: unify model provider detection and add Wan icon (#7373)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/assets/custom/icon-wan.tsx`, `web/src/assets/custom/wan.png`, `web/src/features/channels/lib/model-categories.ts`, `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`, `web/src/lib/lobe-icon.tsx`, `web/src/lib/model-provider.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型提供者识别统一及 Wan 图标
- AI recommendation: `migrate` — 上游把日志徽标/渠道模型分类提取到同一 resolveModelProvider 并补 Wan 图标。本地仍有多套规则，适合只迁入显示层识别，保留本地 Moonshot/Kimi 与 Volc 规则。
- AI risks: 显示层提供者分类不得改变后端路由/计费模型身份；别名与含品牌子串的模型需确定性测试。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/lib/model-provider.ts, web/src/lib/lobe-icon.tsx, web/src/features/channels/lib/model-categories.ts, web/src/features/usage-logs/components/model-badge.tsx, web/src/assets/custom/icon-wan.tsx, web/src/assets/custom/wan.png, web/src/lib/__tests__/model-provider.test.ts
- AI evidence: [upstream:web/src/features/channels/lib/model-categories.ts:23](https://github.com/QuantumNous/new-api/blob/d926120384109a9e55b0cd54a06bb6e3ae860a2c/web/src/features/channels/lib/model-categories.ts#L23) — 固定源码：return resolveModelProvider(modelName)?.name ?? 'Other'
- AI evidence: [downstream:web/src/features/channels/lib/model-categories.ts:28](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/lib/model-categories.ts#L28) — 下游对照：const MODEL_CATEGORY_RULES: readonly ModelCategoryRule[] = [

#### commit-d1c79d728822

- Title: fix(dashboard): align weekly default range (#7355)
- Feature/group: fix(dashboard): align weekly default range (#7355)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/dashboard/constants.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 周统计默认时间范围对齐
- AI recommendation: `merge` — 可作为独立小修复合入：周粒度默认范围由30改29，与上游默认周展示保持一致。本地同常量仍为30，整补丁可干净应用；实施时只检查范围计算和既有图表，不触碰客户正式账期。
- AI risks: UI默认天数不应影响已有保存偏好、账务查询授权或正式账单月份。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/dashboard/constants.ts
- AI evidence: [upstream:web/src/features/dashboard/constants.ts:37](https://github.com/QuantumNous/new-api/blob/d1c79d7288221768a5f7270649d691ac5a137def/web/src/features/dashboard/constants.ts#L37) — 上游周视图默认范围改为29天；代码：week: 29,
- AI evidence: [downstream:web/src/features/dashboard/constants.ts:37](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/dashboard/constants.ts#L37) — 本地仍保留旧默认范围；代码：week: 30,

#### commit-2509e25fa060

- Title: fix(web): prevent combobox dropdowns opening on dialog autofocus (#7365)
- Feature/group: fix(web): prevent combobox dropdowns opening on dialog autofocus (#7365)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/ui/__tests__/combobox.test.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Combobox 对话框自动聚焦行为
- AI recommendation: `defer` — 补丁修改上游新 OptionCombobox 的 onOpenChange(details.reason)/onFocus；本地 legacy options 仍委托 LegacyComboboxInput，没有该实现。需先复现本地自动聚焦问题，再按实际组件修复。
- AI risks: 不能仅为套补丁替换共用 combobox；搜索输入变化、首次聚焦、点击和键盘选择应分别验证。
- AI dependencies: commit-387a40914853, commit-8f72ecbbfd86
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/components/ui/combobox.tsx:119](https://github.com/QuantumNous/new-api/blob/2509e25fa0601d26f5b8fbb3194ec76689a91a77/web/src/components/ui/combobox.tsx#L119) — 固定源码：if (details.reason !== 'input-change') setSearch('')
- AI evidence: [downstream:web/src/components/ui/combobox.tsx:65](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/components/ui/combobox.tsx#L65) — 下游对照：<LegacyComboboxInput

#### commit-76f7dafd2b82

- Title: fix(audio): normalize file extension case in GetAudioDuration (#7321)
- Feature/group: fix(audio): normalize file extension case in GetAudioDuration (#7321)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/audio.go`, `common/audio_case_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 音频扩展名大小写
- AI recommendation: `merge` — 本地GetAudioDuration直接按原始扩展名switch，上游先strings.ToLower并加入大小写测试；可优先完整采用，不改变媒体计费公式。
- AI risks: 仍须保留本地音频元数据上界/饱和转换，验证大写扩展与小写得到相同解析。
- AI dependencies: none identified
- AI suggested paths (not approval): common/audio.go, common/audio_case_test.go
- AI evidence: [upstream:common/audio.go:24](https://github.com/QuantumNous/new-api/blob/76f7dafd2b82fc305533ce6a3065b3de2d338bae/common/audio.go#L24) — 解析前统一扩展名大小写。
- AI evidence: [downstream:common/audio.go:24](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/common/audio.go#L24) — 本地此前直接分派扩展名。

#### commit-04c64734cc7c

- Title: fix(pricing): keep model status bar spacing uniform (#7284)
- Feature/group: fix(pricing): keep model status bar spacing uniform (#7284)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型状态条统一间距
- AI recommendation: `skip` — 只针对上游新的24槽状态条固定高度和间距，本地仍是三柱紧凑设计，不存在该布局问题。本轮保留现有展示；如果日后采用小时序列，再从最终目标版本统一设计。
- AI risks: 不能为了样式一行改动带入不存在的24槽界面。
- AI dependencies: commit-5c7cca015525
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/pricing/components/model-perf-badge.tsx:99](https://github.com/QuantumNous/new-api/blob/04c64734cc7c2e58a9efd0b247182330f3f52cce/web/src/features/pricing/components/model-perf-badge.tsx#L99) — 修复目标是上游24槽状态条容器；代码：className='mt-1 flex h-3 w-24 items-center gap-px'
- AI evidence: [downstream:web/src/features/pricing/components/model-perf-badge.tsx:70](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/pricing/components/model-perf-badge.tsx#L70) — 下游当前为三柱状态；代码：...Array(Math.max(0, 3 - statusRates.length)).fill(null),

#### commit-9fe0457ee1f4

- Title: Add Responses WebSocket relay support (#5062)
- Feature/group: Add Responses WebSocket relay support (#5062)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/json.go`, `controller/channel.go`, `controller/channel_multi_key_test.go`, `controller/relay.go`, `controller/responses_websocket.go`, `controller/responses_websocket_test.go`, `dto/channel_constraints.go`, `main.go`, `middleware/auth.go`, `middleware/auth_test.go`, `middleware/task_plugin_origin_task_test.go`, `model/channel.go`, `model/channel_constraint.go`, `pkg/wsmanager/wsmanager.go`, `pkg/wsmanager/wsmanager_test.go`, `relay/channel/openai/relay_responses.go`, `relay/chat_completions_via_responses_test.go`, `relay/helper/valid_request.go`, `relay/mjproxy_handler.go`, `relay/relay_task_test.go`, `relay/request_billing.go`, `relay/responses_handler.go`, `relay/responses_request.go`, `relay/responses_websocket.go`, `relay/responses_websocket_test.go`, `relay/websocket.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/openai_response.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp_test.go`, `router/relay-router.go`, `service/channel.go`, `service/relay_error.go`, `service/relay_error_test.go`, `service/responses_usage.go`, `service/responses_usage_test.go`, `service/ws_close.go`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/responses-websocket-setting.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/responses-websocket-setting.tsx`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/channel-form-errors.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/channel.go, controller/relay.go, main.go, model/channel.go, relay/chat_completions_via_responses_test.go, relay/mjproxy_handler.go, relay/responses_handler.go, relaykit/dto/channel_settings.go, router/relay-router.go, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/channel-form-errors.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses WebSocket 会话能力
- AI recommendation: `defer` — 本地只有HTTP Responses与Realtime WS，没有本提交新增的Responses WS会话、每请求鉴权、多密钥断连和计费生命周期。52文件功能需独立设计并包含后续关联错误/关闭帧/渠道保存修复，不能当普通路由补丁。
- AI risks: 必须覆盖每个response.create授权、取消/断连结算一次、上下文大小上限、渠道禁用、Moonshot HTTP入口隔离。
- AI dependencies: commit-0ed497f066a6, commit-7bbe85bcb095
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/channel.go:86](https://github.com/QuantumNous/new-api/blob/9fe0457ee1f4b9de407a254500d54f5a8f41ee29/controller/channel.go#L86) — 渠道关闭新增活动WebSocket撤销机制。
- AI evidence: [downstream:router/relay-router.go:136](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/relay-router.go#L136) — 本地只注册HTTP Responses入口。

#### commit-8529f209c859

- Title: feat: vllm channel && sglang channel (#7332)
- Feature/group: feat: vllm channel && sglang channel (#7332)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.gitignore`, `common/advanced_custom_presets.go`, `common/api_type.go`, `common/endpoint_type.go`, `constant/channel.go`, `controller/channel.go`, `controller/channel_inference.go`, `controller/channel_inference_test.go`, `controller/channel_upstream_update.go`, `model/channel.go`, `model/channel_constraint.go`, `model/channel_settings_test.go`, `relay/channel/advancedcustom/adaptor.go`, `relay/channel/advancedcustom/presets_test.go`, `relay/channel/advancedcustom/rerank.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/relay_responses_billing_test.go`, `relay/common/relay_info.go`, `relay/helper/max_tokens_bounds_test.go`, `relay/helper/valid_request.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/claude.go`, `relaykit/dto/openai_compaction.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_request_zero_value_test.go`, `relaykit/dto/openai_response.go`, `relaykit/dto/values.go`, `relaykit/dto/values_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp.go`, `router/channel-router.go`, `router/channel_router_test.go`, `web/src/assets/brand-icons/sglang.svg`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/inference-status.test.tsx`, `web/src/features/channels/components/channels-columns.tsx`, `web/src/features/channels/components/channels-dialogs.tsx`, `web/src/features/channels/components/channels-provider.tsx`, `web/src/features/channels/components/dialogs/inference-status-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/__tests__/inference-status.test.ts`, `web/src/features/channels/lib/__tests__/new-api-channel.test.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/lib/channel-type-config.ts`, `web/src/features/channels/lib/channel-utils.ts`, `web/src/features/channels/lib/inference-status.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/lobe-icon.tsx`
- Overlap with downstream changes: .gitignore, common/api_type.go, constant/channel.go, controller/channel.go, model/channel.go, model/channel_settings_test.go, relay/common/relay_info.go, relaykit/dto/channel_settings.go, relaykit/dto/openai_request.go, relaykit/dto/openai_request_zero_value_test.go, relaykit/dto/values.go, relaykit/dto/values_test.go, router/channel-router.go, router/channel_router_test.go, web/src/features/channels/api.ts, web/src/features/channels/components/channels-provider.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/lib/channel-type-config.ts, web/src/features/channels/lib/channel-utils.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: vLLM/SGLang 渠道与推理状态
- AI recommendation: `defer` — 该提交同时增加两个渠道类型、advanced custom预设、推理状态接口、compaction及多处前端配置。本地没有对应预设文件，现有渠道还承载下游Volc/资产扩展；应先确定产品需求和只引入哪些推理端点。
- AI risks: 需核验渠道类型编号、身份过滤、模型列表/推理状态权限、compact和rerank协议，保留三数据库Channel设置兼容。
- AI dependencies: commit-9fe0457ee1f4
- AI suggested paths (not approval):
- AI evidence: [upstream:common/advanced_custom_presets.go:9](https://github.com/QuantumNous/new-api/blob/8529f209c85913de9ec98c5af49dcfce4a41361a/common/advanced_custom_presets.go#L9) — 新增两个推理渠道的路由和认证预设。
- AI evidence: [downstream:common/api_type.go:78](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/common/api_type.go#L78) — 本地AdvancedCustom API类型尚未包含新渠道。

#### commit-2bfb89c1b99d

- Title: fix(pricing): prevent sort menu layout shift (#7145)
- Feature/group: fix(pricing): prevent sort menu layout shift (#7145)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/components/pricing-toolbar.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 价格排序菜单布局抖动
- AI recommendation: `migrate` — 建议迁移DropdownMenu的modal=false，避免排序菜单打开时锁定页面滚动导致布局位移。本地下游工具栏仍使用默认modal菜单，但组件周边布局不同，适合单点适配而非整文件覆盖。
- AI risks: 确认键盘焦点、点击外部关闭和移动端菜单仍正确。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/pricing/components/pricing-toolbar.tsx
- AI evidence: [upstream:web/src/features/pricing/components/pricing-toolbar.tsx:152](https://github.com/QuantumNous/new-api/blob/2bfb89c1b99d5eb7a799ba5706813614bfb6189a/web/src/features/pricing/components/pricing-toolbar.tsx#L152) — 菜单不锁定整个页面滚动；代码：<DropdownMenu modal={false}>
- AI evidence: [downstream:web/src/features/pricing/components/pricing-toolbar.tsx:223](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/pricing/components/pricing-toolbar.tsx#L223) — 本地排序菜单尚未关闭modal模式；代码：<DropdownMenu>

#### commit-8e5e09166cf8

- Title: fix(ollama): preserve tool calls from final stream frame | 保留流式末尾帧的 tool calls (#7376)
- Feature/group: fix(ollama): preserve tool calls from final stream frame | 保留流式末尾帧的 tool calls (#7376)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/ollama/stream.go`, `relay/channel/ollama/stream_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Ollama 终止帧工具调用
- AI recommendation: `merge` — 本地ollamaStreamHandler仅在!Done帧发送内容，结束帧直接结算；上游先提取同样delta，再在Done帧有内容时发送，避免末帧tool_calls丢失。可完整采用并运行终帧工具测试。
- AI risks: 验证工具索引、文本/思考同帧、usage和finish_reason只结束一次；不得因重复末帧重发工具。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/channel/ollama/stream.go, relay/channel/ollama/stream_test.go
- AI evidence: [upstream:relay/channel/ollama/stream.go:185](https://github.com/QuantumNous/new-api/blob/8e5e09166cf83731df1819fd1cb9c50ea2576ed6/relay/channel/ollama/stream.go#L185) — Done帧有有效载荷时先发delta。
- AI evidence: [downstream:relay/channel/ollama/stream.go:132](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/ollama/stream.go#L132) — 本地仅在非终止帧发送载荷。

#### commit-62f8db775bcd

- Title: feat(channel): 修复获取DeepSeek余额错误 (#6814)
- Feature/group: feat(channel): 修复获取DeepSeek余额错误 (#6814)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel-billing.go`, `controller/channel_billing_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: DeepSeek 余额币种规范化
- AI recommendation: `migrate` — 本地只读取CNY且直接存入Balance；上游优先USD，否则以有效USDExchangeRate换算CNY，并拒绝非有限/负数。建议选择性迁移币种规范化及表测试，先确认平台余额显示的统一USD约定。
- AI risks: 错误汇率会影响余额展示和渠道告警；不应改变客户售价、钱包或下游渠道成本折扣结算。
- AI dependencies: none identified
- AI suggested paths (not approval): controller/channel-billing.go, controller/channel-billing_test.go
- AI evidence: [upstream:controller/channel-billing.go:278](https://github.com/QuantumNous/new-api/blob/62f8db775bcdaa5da144d9c4a36f5ead82de3ac5/controller/channel-billing.go#L278) — 上游明确USD优先及CNY换算策略。
- AI evidence: [downstream:controller/channel-billing.go:297](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/channel-billing.go#L297) — 本地仅接受CNY并原样保存。

#### commit-d3874db61fce

- Title: feat(rate-limit): optimize in-memory limiter allocation and cleanup (#6807)
- Feature/group: feat(rate-limit): optimize in-memory limiter allocation and cleanup (#6807)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/rate-limit.go`, `common/rate-limit_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 内存限流分配与闲置清理
- AI recommendation: `migrate` — 本地为切片队列和定时全表扫描；上游改链式滑窗+LRU，保持Request接口。可单独适配这一性能行为并采用确定性过期/并发测试，不连带改模型限流成功口径。
- AI risks: 需验证边界窗口、超额不入队、闲置淘汰、重复Init与并发，避免LRU回收仍在执行的请求。
- AI dependencies: none identified
- AI suggested paths (not approval): common/rate-limit.go, common/rate-limit_test.go
- AI evidence: [upstream:common/rate-limit.go:16](https://github.com/QuantumNous/new-api/blob/d3874db61fce61b6c0f5c37523b21912a3f2658d/common/rate-limit.go#L16) — 上游以请求队列避免按limit预分配。
- AI evidence: [downstream:common/rate-limit.go:9](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/common/rate-limit.go#L9) — 本地仍为切片滑窗存储。

#### commit-5caafd3d84dc

- Title: feat(ollama): add per-channel OpenAI-compatible chat switch (#7382)
- Feature/group: feat(ollama): add per-channel OpenAI-compatible chat switch (#7382)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/ollama/adaptor.go`, `relaykit/dto/channel_settings.go`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/channel-form-errors.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relaykit/dto/channel_settings.go, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-form-errors.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Ollama 每渠道 OpenAI Chat 开关
- AI recommendation: `migrate` — 上游提供默认关闭的ollama_openai_chat，显式切换/v1/chat/completions并初始化OpenAI响应状态。可按需在本地现有Ollama适配器和表单迁入，保留默认/api/chat及/api/generate，不强制采用早先Claude原生透传。
- AI risks: 需验证开关保存、旧记录默认、流/非流响应、thinking转换、所有前端语言；旧Ollama实例不能被默认切换。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/channel/ollama/adaptor.go, relaykit/dto/channel_settings.go, web/src/features/channels/components/__tests__/channel-configuration.test.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-configuration.ts, web/src/features/channels/lib/channel-form-errors.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, relay/channel/ollama/adaptor_test.go, web/src/features/channels/lib/__tests__/channel-form.test.ts
- AI evidence: [upstream:relay/channel/ollama/adaptor.go:61](https://github.com/QuantumNous/new-api/blob/5caafd3d84dc74c8b0d081524f57a534b3980cb0/relay/channel/ollama/adaptor.go#L61) — 只在每渠道开关开启时切OpenAI Chat。
- AI evidence: [downstream:relay/channel/ollama/adaptor.go:57](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/ollama/adaptor.go#L57) — 本地默认使用Ollama原生Chat。

#### commit-fa90b2312cfb

- Title: fix(gemini): accept case-insensitive thinkingLevel and log canonical effort (#7387)
- Feature/group: fix(gemini): accept case-insensitive thinkingLevel and log canonical effort (#7387)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/gemini.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Gemini thinkingLevel 大小写
- AI recommendation: `defer` — 补丁改变统一reasoning.ParseEffort验证及日志规范化，本地还没有上游那套reasoning状态/模型能力验证接口。可保留需求，但应随推理重构整体迁移，避免仅日志小写而请求验证仍拒绝的半修复。
- AI risks: 原生请求的未知值应保留，跨协议转换需返回确定性错误；日志effort不能伪造不同实际推理力度。
- AI dependencies: commit-0ed497f066a6, commit-7c044d7c5c2d, commit-6b659fd61c50
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/relayconvert/reasoning/gemini.go:216](https://github.com/QuantumNous/new-api/blob/fa90b2312cfb686aaa6010f9280bffcf0bf6d391/relaykit/relayconvert/reasoning/gemini.go#L216) — 上游比较规范化后的思考力度。
- AI evidence: [downstream:relay/common/relay_info.go:469](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/common/relay_info.go#L469) — 本地仍直接记录原始thinkingLevel。

#### commit-81336fc69b8a

- Title: fix(gemini): reject :countTokens as an unknown route instead of relaying it as generateContent (#7388)
- Feature/group: fix(gemini): reject :countTokens as an unknown route instead of relaying it as generateContent (#7388)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `router/relay-router.go`
- Overlap with downstream changes: router/relay-router.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 拒绝未支持的 Gemini countTokens
- AI recommendation: `migrate` — 本地/v1beta通配路径会把:countTokens交给Gemini relay，上游在鉴权/渠道选择前返回未知路由，防止意外按generateContent生成与收费。建议最小路由守卫迁移并补不触发上游回归。
- AI risks: 需同时验证大小写/路径后缀、普通generateContent和streamGenerateContent不受影响，保持未知路由错误形状。
- AI dependencies: none identified
- AI suggested paths (not approval): router/relay-router.go, router/relay_router_test.go
- AI evidence: [upstream:router/relay-router.go:195](https://github.com/QuantumNous/new-api/blob/81336fc69b8a932e70c24a8dbf439bec94c98328/router/relay-router.go#L195) — 上游在Gemini组前置拦截未知计数方法。
- AI evidence: [downstream:router/relay-router.go:229](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/relay-router.go#L229) — 本地Gemini组未加countTokens守卫。

#### commit-610334dbd760

- Title: fix(relay): request stream usage on every cross-protocol conversion to OpenAI chat (#7389)
- Feature/group: fix(relay): request stream usage on every cross-protocol conversion to OpenAI chat (#7389)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/advancedcustom/adaptor.go`, `relay/channel/advancedcustom/adaptor_test.go`, `relay/channel/openai/adaptor.go`
- Overlap with downstream changes: relay/channel/openai/adaptor.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 跨协议流请求 usage 回传
- AI recommendation: `migrate` — 本地AdvancedCustom各跨协议Chat转换重复处理，Gemini→OpenAI路径也没有统一请求stream_options。上游仅在SupportStreamOptions且IsStream时请求usage。建议适配现有转换接口，增加支持/不支持渠道矩阵测试。
- AI risks: 应保留不支持stream_options的供应商行为、原始usage与Moonshot输出；不能把估算用量冒充上游实报。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/channel/advancedcustom/adaptor.go, relay/channel/openai/adaptor.go, relay/channel/advancedcustom/adaptor_test.go, relay/channel/openai/adaptor_test.go
- AI evidence: [upstream:relay/channel/advancedcustom/adaptor.go:529](https://github.com/QuantumNous/new-api/blob/610334dbd760f1c5199a09ae5f860f11451f4c56/relay/channel/advancedcustom/adaptor.go#L529) — 跨协议流请求显式要求返回usage。
- AI evidence: [downstream:relay/channel/advancedcustom/adaptor.go:94](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/advancedcustom/adaptor.go#L94) — 本地该路径各自转换，没有统一跨协议usage注入。

#### commit-69a50029819a

- Title: feat(channel): per-route pass-through for advanced custom channels (#7386)
- Feature/group: feat(channel): per-route pass-through for advanced custom channels (#7386)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/advancedcustom/adaptor.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/channel_settings_test.go`, `web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/advanced-custom.ts`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relay/common/relay_info.go, relaykit/dto/channel_settings.go, relaykit/dto/channel_settings_test.go, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: AdvancedCustom 按路由请求体透传
- AI recommendation: `migrate` — 本地只有渠道级透传，上游给每个路由增加PassThroughBodyEnabled并禁止管理路由及有协议转换路由开启。可按明确使用需求迁入，但需补保存/校验和凭据边界测试。
- AI risks: 透传会绕过模型映射、参数覆盖与格式转换；须保留下游模型计费身份、管理员敏感配置保护及管理接口不可透传。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/channel/advancedcustom/adaptor.go, relay/common/relay_info.go, relay/common/relay_info_test.go, relaykit/dto/channel_settings.go, relaykit/dto/channel_settings_test.go, web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/advanced-custom.ts, web/src/features/channels/lib/channel-configuration.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, relay/channel/advancedcustom/adaptor_test.go
- AI evidence: [upstream:relaykit/dto/channel_settings.go:149](https://github.com/QuantumNous/new-api/blob/69a50029819a26c53e6babd276d49cfe2f8880ad/relaykit/dto/channel_settings.go#L149) — 上游每路由透传选项。
- AI evidence: [downstream:relay/channel/advancedcustom/adaptor.go:290](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/advancedcustom/adaptor.go#L290) — 本地仍通过转换器值拒绝透传。

#### commit-7209b6db95ef

- Title: fix(middleware): 隐藏无可用渠道错误中的任务插件标识 (#7414)
- Feature/group: fix(middleware): 隐藏无可用渠道错误中的任务插件标识 (#7414)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `middleware/distributor.go`, `middleware/distributor_test.go`
- Overlap with downstream changes: i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, middleware/distributor.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件分流错误信息脱敏
- AI recommendation: `defer` — 从无可用渠道错误中去掉内部插件标识和管理说明，保留模型与组信息。该修复应随插件启用，覆盖早期 32c261923a 的信息泄露；当前没有插件认领分支。
- AI risks: 外部用户只应获得可操作且不泄露内部路由布局的信息。
- AI dependencies: commit-eb48396d5fe9, commit-32c261923a97
- AI suggested paths (not approval):
- AI evidence: [upstream:middleware/distributor.go:225](https://github.com/QuantumNous/new-api/blob/7209b6db95ef33bfb0f6344ebf61c1fec9af780c/middleware/distributor.go#L225) — 插件无渠道分支仍使用脱敏消息；代码：return i18n.T(c, i18n.MsgDistributorNoAvailableChannelTaskPlugin, map[string]any{"Group": group, "Model": modelName})

#### commit-57791e31c62b

- Title: docs(github): tighten issue and PR templates for pass-through and unfiltered AI text (#7438)
- Feature/group: docs(github): tighten issue and PR templates for pass-through and unfiltered AI text (#7438)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.agents/github/ISSUE.md`, `.agents/github/PR.md`, `.github/ISSUE_TEMPLATE/bug_report.yml`, `.github/ISSUE_TEMPLATE/bug_report_en.yml`, `.github/ISSUE_TEMPLATE/feature_request.yml`, `.github/ISSUE_TEMPLATE/feature_request_en.yml`, `.github/PULL_REQUEST_TEMPLATE.md`, `.github/PULL_REQUEST_TEMPLATE/en.md`, `AGENTS.md`
- Overlap with downstream changes: .github/PULL_REQUEST_TEMPLATE.md, AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游 Issue/PR 事实与AI文本规则
- AI recommendation: `skip` — 只调整受保护治理/提交模板，包含上游自身不接受的业务范围。下游不能通过版本同步继承为当前用户的额外限制。
- AI risks: 继续保留本地人工描述、AI披露、模板结构与已有授权规则；需要治理改动另行审查。
- AI dependencies: commit-e468b73915e5, commit-d8ca0ed0bb59
- AI suggested paths (not approval):
- AI evidence: [upstream:.agents/github/PR.md:11](https://github.com/QuantumNous/new-api/blob/57791e31c62bdf97df855f3b7dc912046b3bc990/.agents/github/PR.md#L11) — 固定源码：Keep the body short and factual. Do not paste unfiltered AI-generated text
- AI evidence: [downstream:AGENTS.md:162](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/AGENTS.md#L162) — 下游对照：**Pull requests:** When creating a pull request:

#### commit-ed0ad6e3ff9f

- Title: feat(channels): add compact quick options to channel editor
- Feature/group: feat(channels): add compact quick options to channel editor
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/components/drawers/channel-configuration.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道编辑快捷参数区
- AI recommendation: `defer` — 向新增 ChannelConfiguration 插入 quickOptions，依赖本地不存在的分阶段编辑布局。暂缓，待渠道配置方案确认后可组合现有控件。
- AI risks: 参数开关必须遵守渠道支持能力和敏感权限；不能因布局迁移丢失下游专有配置。
- AI dependencies: commit-505805a4c3ab
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/channels/components/drawers/channel-configuration.tsx:42](https://github.com/QuantumNous/new-api/blob/ed0ad6e3ff9f8785a19d93cba73cf06ef2219668/web/src/features/channels/components/drawers/channel-configuration.tsx#L42) — 固定源码：quickOptions: ReactNode
- AI evidence: [downstream:web/src/features/channels/components/drawers/channel-mutate-drawer.tsx:621](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/drawers/channel-mutate-drawer.tsx#L621) — 下游对照：export function ChannelMutateDrawer({

#### commit-07578faf26d1

- Title: fix(billing): count upstream images by payload instead of data length
- Feature/group: fix(billing): count upstream images by payload instead of data length
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/image_stream_test.go`, `relay/channel/openai/relay_image.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 图片载荷计数与收费
- AI recommendation: `defer` — 本地按data数组长度计费且保留MaxImageN上界；上游改为max(URL项数,b64项数)，还支持对象形data。不能直接认定URL项与Base64项属于同一张图：混合独立图片会产生少计费，须先确定供应商计数契约。
- AI risks: 需覆盖同对象双字段、不同对象混合格式、空载荷、重复同图与恶意超大数组；迁移时保留本地updateOpenAIImageCount上界和AddOtherRatio。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/openai/relay_image.go:83](https://github.com/QuantumNous/new-api/blob/07578faf26d1a21c70bf0fbf29f12730a212c583/relay/channel/openai/relay_image.go#L83) — 上游以两类载荷较大者作为收费数量。
- AI evidence: [downstream:relay/channel/openai/relay_image.go:26](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/openai/relay_image.go#L26) — 本地计数有独立上界保护，不可被新逻辑覆盖。

#### commit-d82678a9708b

- Title: docs: move billing rules to .agents/rules/billing.md behind a read gate
- Feature/group: docs: move billing rules to .agents/rules/billing.md behind a read gate
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.agents/rules/billing.md`, `AGENTS.md`, `CLAUDE.md`
- Overlap with downstream changes: AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 计费规则迁往独立受保护文件
- AI recommendation: `skip` — 仅重组 AGENTS/CLAUDE 读门及 .agents/rules/billing.md，不新增业务功能。本地当前 AGENTS 已有完整账务不变量和表达式文档要求，禁止机械覆盖或挪走。
- AI risks: 规则重组不得削弱任何下游钱包/饱和/审计约束；另行治理时要逐条比较。
- AI dependencies: commit-d5803532bdcc, commit-fff0635bb14b
- AI suggested paths (not approval):
- AI evidence: [upstream:.agents/rules/billing.md:3](https://github.com/QuantumNous/new-api/blob/d82678a9708b611d52386c01f04238ec01f07463/.agents/rules/billing.md#L3) — 固定源码：Mandatory conventions for quota, pricing, pre-consume, settlement, refund, and usage-quantity code in the Go backend. `AGENTS.md` defines when a task touches billing (see **Billing rules (mandatory read gate)**); any task that matches MUST read this whole file before planning, coding, or reviewing. Tasks that match none of those triggers may skip it. Rules use the same MUST / Do NOT wording as `AGENTS.md` and carry the same weight.
- AI evidence: [downstream:AGENTS.md:105](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/AGENTS.md#L105) — 下游对照：**Billing safety invariants:** Quota/billing code MUST never produce a negative charge (a credit) from arithmetic overflow or unvalidated input. Apply defense in depth:

#### commit-0cde9d94f6d3

- Title: feat(channels): add floating model redirect workbench to channel editor
- Feature/group: feat(channels): add floating model redirect workbench to channel editor
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/floating-window.tsx`, `web/src/components/floating-window/__tests__/floating-window.test.tsx`, `web/src/components/multi-select.tsx`, `web/src/components/multi-select/__tests__/option-hint.test.tsx`, `web/src/components/ui/combobox-input.tsx`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-batch-dialog.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-editor.test.tsx`, `web/src/features/channels/components/__tests__/upstream-model-selection.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/components/dialogs/configure-models-dialog.tsx`, `web/src/features/channels/components/drawers/channel-configuration.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/model-mapping-batch-dialog.tsx`, `web/src/features/channels/components/model-mapping-editor.tsx`, `web/src/features/channels/components/model-redirect-panel.tsx`, `web/src/features/channels/components/upstream-model-selection.tsx`, `web/src/features/channels/hooks/__tests__/use-redirect-panel-placement.test.tsx`, `web/src/features/channels/hooks/use-redirect-panel-placement.ts`, `web/src/features/channels/lib/__tests__/model-mapping-rules.test.ts`, `web/src/features/channels/lib/__tests__/model-naming-patterns.test.ts`, `web/src/features/channels/lib/index.ts`, `web/src/features/channels/lib/model-mapping-rules.ts`, `web/src/features/channels/lib/model-naming-patterns.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 悬浮模型重定向工作台
- AI recommendation: `defer` — 新增 floating-window、重定向面板、命名建议、批量映射等31文件能力；涉及模型路由而非纯样式。需先确认下游公开/上游/计费模型命名边界，再评审最终形态。
- AI risks: 自动别名/前缀建议不可悄悄改变真实 Kimi 名称和计费身份；触屏拖动与弹层焦点也需测试。
- AI dependencies: commit-505805a4c3ab, commit-ed0ad6e3ff9f
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/components/floating-window.tsx:45](https://github.com/QuantumNous/new-api/blob/0cde9d94f6d3b77d5258ca6b36959f895d26f2a4/web/src/components/floating-window.tsx#L45) — 固定源码：type FloatingWindowProps = {
- AI evidence: [downstream:web/src/features/channels/components/drawers/channel-mutate-drawer.tsx:621](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/drawers/channel-mutate-drawer.tsx#L621) — 下游对照：export function ChannelMutateDrawer({

#### commit-8b2c7105f91f

- Title: feat(relay): classify the protocol outcome on StreamStatus
- Feature/group: feat(relay): classify the protocol outcome on StreamStatus
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/common/stream_status.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 流协议终态模型
- AI recommendation: `defer` — 本地StreamStatus仅记录传输结束原因和错误；上游新增completed/failed/incomplete/cancelled以及必须终帧标记。它是后续指标、成功限流和失败结算的基础，应成组设计后引入。
- AI risks: 不能把HTTP200/SSE传输结束当成功；需要所有协议适配器统一标记终态，避免缺标记被误算失败或成功。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/common/stream_status.go:26](https://github.com/QuantumNous/new-api/blob/8b2c7105f91f2316f90eb86f9b1c6b98774c7105/relay/common/stream_status.go#L26) — 新增独立于传输的协议结果。
- AI evidence: [downstream:relay/common/stream_status.go:31](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/common/stream_status.go#L31) — 本地仍为旧传输状态结构。

#### commit-73a471f3a8b1

- Title: feat(policy): add request policies settings and routing decision records
- Feature/group: feat(policy): add request policies settings and routing decision records
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel.go`, `controller/channel_pin_retry_test.go`, `controller/relay.go`, `controller/relay_task_plugin_test.go`, `controller/request_policy.go`, `middleware/distributor.go`, `model/option.go`, `model/request_policy.go`, `model/request_policy_test.go`, `relay/mjproxy_handler.go`, `relay/request_billing.go`, `router/api-router.go`, `service/channel_affinity.go`, `service/channel_affinity_template_test.go`, `service/log_info_generate.go`, `service/midjourney.go`, `service/relay_error.go`, `service/relay_error_test.go`, `service/request_policy.go`, `setting/operation_setting/channel_affinity_setting.go`, `web/.oxlintrc.json`, `web/src/components/layout/config/system-settings.config.ts`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/inference-status.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/index.tsx`, `web/src/features/channels/types.ts`, `web/src/features/keys/components/api-keys-mutate-drawer.tsx`, `web/src/features/system-settings/components/settings-accordion.tsx`, `web/src/features/system-settings/components/settings-page.tsx`, `web/src/features/system-settings/general/channel-affinity/index.tsx`, `web/src/features/system-settings/general/channel-affinity/rule-editor-dialog.tsx`, `web/src/features/system-settings/general/channel-affinity/session-rules-table.tsx`, `web/src/features/system-settings/general/channel-affinity/types.ts`, `web/src/features/system-settings/models/index.tsx`, `web/src/features/system-settings/models/routing-reliability-section.tsx`, `web/src/features/system-settings/models/section-registry.tsx`, `web/src/features/system-settings/request-limits/sensitive-words-section.tsx`, `web/src/features/system-settings/request-policies/__tests__/routing.test.tsx`, `web/src/features/system-settings/request-policies/__tests__/settings.test.tsx`, `web/src/features/system-settings/request-policies/api.ts`, `web/src/features/system-settings/request-policies/channel-health-section.tsx`, `web/src/features/system-settings/request-policies/decision-record.tsx`, `web/src/features/system-settings/request-policies/defaults.ts`, `web/src/features/system-settings/request-policies/index.tsx`, `web/src/features/system-settings/request-policies/policy-label.ts`, `web/src/features/system-settings/request-policies/related-policy-link.tsx`, `web/src/features/system-settings/request-policies/request-checks-section.tsx`, `web/src/features/system-settings/request-policies/retry-section.tsx`, `web/src/features/system-settings/request-policies/routing-form.ts`, `web/src/features/system-settings/request-policies/routing-section.tsx`, `web/src/features/system-settings/request-policies/section-registry.tsx`, `web/src/features/system-settings/request-policies/use-save-policy.ts`, `web/src/features/system-settings/security/section-registry.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/routeTree.gen.ts`, `web/src/routes/_authenticated/system-settings/models/$section.tsx`, `web/src/routes/_authenticated/system-settings/operations/$section.tsx`, `web/src/routes/_authenticated/system-settings/request-policies/$section.tsx`, `web/src/routes/_authenticated/system-settings/request-policies/index.tsx`, `web/src/routes/_authenticated/system-settings/security/$section.tsx`
- Overlap with downstream changes: controller/channel.go, controller/relay.go, middleware/distributor.go, model/option.go, relay/mjproxy_handler.go, router/api-router.go, service/midjourney.go, web/src/components/layout/config/system-settings.config.ts, web/src/features/channels/api.ts, web/src/features/channels/types.ts, web/src/features/system-settings/models/index.tsx, web/src/features/system-settings/models/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/routeTree.gen.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 请求策略设置与重试决策记录
- AI recommendation: `defer` — 71文件引入RequestPolicy、全局/渠道自动禁用、重试决策和管理UI；本地仍用shouldRetry等既有路径，且有违规收费、任务回溯和下游通道约束。缺状态迁移与权限回归，暂缓整组。
- AI risks: 重试/禁用政策影响可用性和收费；不能覆盖任务凭据延续、原任务渠道锁定、管理员审计与租户边界。
- AI dependencies: commit-8b2c7105f91f
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/relay.go:192](https://github.com/QuantumNous/new-api/blob/73a471f3a8b1cdf5dfe371184dfc3579cdcbefcc/controller/relay.go#L192) — 上游用显式策略决策替代shouldRetry。
- AI evidence: [downstream:controller/relay.go:269](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/relay.go#L269) — 本地仍走现有重试策略。

#### commit-8a11b4305ed1

- Title: refactor(relaykit): surface reasoning conversion diagnostics
- Feature/group: refactor(relaykit): surface reasoning conversion diagnostics
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/adaptor.go`, `relay/convert_request_error_test.go`, `relay/helper/reasoning_suffix.go`, `relaykit/relayconvert/claude_default_max_tokens_test.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_responses_req.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req_test.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/gemini.go`, `relaykit/relayconvert/reasoning/gemini_test.go`, `relaykit/relayconvert/reasoning/intent.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/request_compat.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: relay/channel/openai/adaptor.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 推理转换诊断接口
- AI recommendation: `defer` — 上游把FromOpenAIChat/FromClaude等签名改为同时返回diagnostics，并贯穿宿主日志；本地没有同一套reasoning Intent基础。应随7137与@修饰符整组迁移，不能只替换调用签名。
- AI risks: 诊断需管理员可见且不得携带凭据/请求内容；relaykit公共API及外部嵌入者兼容尚未验证。
- AI dependencies: commit-0ed497f066a6, commit-7c044d7c5c2d, commit-6b659fd61c50
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/openai/adaptor.go:270](https://github.com/QuantumNous/new-api/blob/8a11b4305ed139b55ee91f914c6a19d59c783d72/relay/channel/openai/adaptor.go#L270) — 上游推理API增加诊断返回值。
- AI evidence: [downstream:relay/channel/openai/adaptor.go:257](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/openai/adaptor.go#L257) — 本地仍是旧推理适配代码。

#### commit-a5b40663d330

- Title: perf(metrics): record relay outcomes at the request boundary and classify stream results
- Feature/group: perf(metrics): record relay outcomes at the request boundary and classify stream results
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `constant/context_key.go`, `controller/perf_metrics.go`, `controller/relay.go`, `pkg/perf_metrics/metrics.go`, `pkg/perf_metrics/metrics_test.go`, `pkg/perf_metrics/outcome.go`, `pkg/perf_metrics/types.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_responses.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/stream_tool_billing_test.go`, `relay/common/relay_info.go`, `relay/helper/stream_result.go`, `service/log_info_generate.go`, `service/quota.go`, `service/responses_usage.go`, `service/responses_usage_test.go`, `service/text_quota.go`, `web/src/features/dashboard/components/models/performance-overview.tsx`, `web/src/features/dashboard/components/overview/performance-health-panel.tsx`, `web/src/features/performance-metrics/__tests__/summary.test.tsx`, `web/src/features/performance-metrics/types.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-details-performance.tsx`, `web/src/features/pricing/components/model-details-uptime-sparkline.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: constant/context_key.go, controller/relay.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, service/quota.go, service/text_quota.go, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 请求边界指标与流结果分类
- AI recommendation: `defer` — 上游将指标从gopool异步样本移到请求defer，并按协议终态分类；本地有自有分类流量监控与当前RecordRelaySample。该40文件变更应与终态分类共同验证，避免重试/业务拒绝重复统计。
- AI risks: 需验证panic、客户端取消、业务拒绝、每请求只计一次和自有监控口径，不能覆盖下游监控边界。
- AI dependencies: commit-8b2c7105f91f, commit-73a471f3a8b1
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/relay.go:134](https://github.com/QuantumNous/new-api/blob/a5b40663d3307fc5011c85b83eb18bbaa33dbad2/controller/relay.go#L134) — 上游在请求收尾统一记录结果。
- AI evidence: [downstream:controller/relay.go:281](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/relay.go#L281) — 本地仍通过异步现有采样路径。

#### commit-6237d9d77eb0

- Title: fix(rate-limit): reserve model rate-limit slots and judge success by response outcome
- Feature/group: fix(rate-limit): reserve model rate-limit slots and judge success by response outcome
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/rate-limit.go`, `middleware/model-rate-limit.go`, `middleware/model_rate_limit_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型成功限流预占位
- AI recommendation: `defer` — 上游为并发中请求预占成功额度，并按真实响应终态释放/计入；本地旧队列没有reservation和协议Outcome。必须连同LRU改造、成功口径和取消/异常路径成组验证，不能单独替换中间件。
- AI risks: 要覆盖并发抢占、淘汰中的预留、重复Complete、失败释放，以及Redis/内存两实现一致性。
- AI dependencies: commit-d3874db61fce, commit-8b2c7105f91f, commit-a5b40663d330
- AI suggested paths (not approval):
- AI evidence: [upstream:common/rate-limit.go:170](https://github.com/QuantumNous/new-api/blob/6237d9d77eb0544329b88ff96d71c1715af4268f/common/rate-limit.go#L170) — 上游新增在途请求预留资源。
- AI evidence: [downstream:common/rate-limit.go:45](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/common/rate-limit.go#L45) — 本地限流仍只有直接入队Request接口。

#### commit-75f3d246e2b1

- Title: fix(responses-ws): carry stream_id through events and correlate errors
- Feature/group: fix(responses-ws): carry stream_id through events and correlate errors
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/responses_websocket_test.go`, `relay/responses_websocket.go`, `relay/responses_websocket_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses WS stream_id 与错误关联
- AI recommendation: `defer` — 本地尚无Responses WS；此修复把stream_id/event_id/response_id贯穿会话事件并结合策略与指标。仅在决定采用完整WS能力时纳入必需后续修复。
- AI risks: 多次response.create与取消/错误不能串流；错误关联字段不能泄露另一请求内容。
- AI dependencies: commit-9fe0457ee1f4, commit-73a471f3a8b1, commit-a5b40663d330
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/responses_websocket.go:46](https://github.com/QuantumNous/new-api/blob/75f3d246e2b1d6a3ea56342d9bc82cb3e1d0b6e2/relay/responses_websocket.go#L46) — 上游WS envelope开始保留stream_id。
- AI evidence: [downstream:router/relay-router.go:136](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/relay-router.go#L136) — 本地仅有HTTP Responses路由。

#### commit-90134d5b6bff

- Title: feat(system-tasks): add filtered task history with cleanup
- Feature/group: feat(system-tasks): add filtered task history with cleanup
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/system_task.go`, `controller/system_task_test.go`, `model/system_task.go`, `model/system_task_test.go`, `router/api-router.go`, `web/src/features/system-info/__tests__/system-tasks-panel.test.tsx`, `web/src/features/system-info/api.ts`, `web/src/features/system-info/components/system-task-history.tsx`, `web/src/features/system-info/components/system-tasks-panel.tsx`, `web/src/features/system-info/components/system-tasks-table.tsx`, `web/src/features/system-info/constants.ts`, `web/src/features/system-info/types.ts`, `web/src/features/system-settings/api.ts`, `web/src/features/system-settings/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/system_task.go, model/system_task.go, router/api-router.go, web/src/features/system-info/api.ts, web/src/features/system-info/components/system-tasks-panel.tsx, web/src/features/system-info/types.ts, web/src/features/system-settings/api.ts, web/src/features/system-settings/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 系统任务历史筛选与清理
- AI recommendation: `defer` — 分页/类型/状态筛选有可观测性价值，但提交同时新增删除历史接口。下游系统任务执行账务与存储迁移，需先界定可清理运行历史与必须永久保留的审计/恢复证据，不能对所有完成任务默认启用清理。若采用可先拆只读列表。
- AI risks: 清理需保留每类型最新调度记录、活动任务和快照后新建行；权限、任务关联证据和大表扫描成本需评审。
- AI dependencies: commit-b7017c251bad
- AI suggested paths (not approval):
- AI evidence: [upstream:model/system_task.go:218](https://github.com/QuantumNous/new-api/blob/90134d5b6bff1a66affa35c5bd8e90e3d74a93f8/model/system_task.go#L218) — 该功能包含实际删除系统任务历史，不只是查询UI；代码：func DeleteSystemTaskHistory(filter SystemTaskFilter) (int64, error) {
- AI evidence: [downstream:model/system_task.go:180](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/system_task.go#L180) — 下游已有系统任务表和列表，是否扩大删除范围需单独界定；代码：func ListSystemTasks(limit int) ([]*SystemTask, error) {

#### commit-a8ed7f7c5f24

- Title: fix(plugins/hailuo): bill only the dimensions each model family supports
- Feature/group: fix(plugins/hailuo): bill only the dimensions each model family supports
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/hailuo_responses_test.go`, `plugins/tasks/hailuo/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Hailuo 按模型区分计费维度
- AI recommendation: `defer` — 按 H3、2.3、02、01 系列分别声明有效分辨率与输入媒体维度，避免对旧模型收取 H3 专属输入费。应与 H3 驱动和用量表达式一起实施，当前不单独迁 JS 元数据。
- AI risks: 已有价格表达式对移除维度的行为需保留/审查，不能静默变为免费或多收。
- AI dependencies: commit-eb48396d5fe9, commit-aece11d2f7f0, commit-73afad588ca7
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/hailuo/plugin.js:1](https://github.com/QuantumNous/new-api/blob/a8ed7f7c5f24e56c2f1612126d82e9fec625056c/plugins/tasks/hailuo/plugin.js#L1) — 代码按模型族划分可计费维度；代码：// Billable dimensions differ per model family: only MiniMax-H3 charges for

#### commit-e2471403c779

- Title: fix(pricing): adjust billing expression condition display
- Feature/group: fix(pricing): adjust billing expression condition display
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/lib/__tests__/billing-expression.test.ts`, `web/src/features/pricing/lib/billing-expression/condition-display.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 表达式时间条件本地化显示
- AI recommendation: `defer` — 日期条件显示统一用toIntlLocale并以UTC格式化，避免界面语言与时区改变日期文字。当前没有新condition-display模块，需在新版表达式显示组被选择后迁移，不能单加无人调用文件。
- AI risks: 显示时区与表达式执行时区不能混淆；需要覆盖中文、繁中和非UTC客户端。
- AI dependencies: commit-d52bdc0b4087
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/pricing/lib/billing-expression/condition-display.ts:109](https://github.com/QuantumNous/new-api/blob/e2471403c77957818b939ad86bbeccd051d2d042/web/src/features/pricing/lib/billing-expression/condition-display.ts#L109) — 日期显示固定UTC以免跨时区偏移；代码：timeZone: 'UTC',

#### commit-42cbe6206a62

- Title: chore(web): add intl-locale lint plugin and update agent rules
- Feature/group: chore(web): add intl-locale lint plugin and update agent rules
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `web/.oxlintrc.json`, `web/AGENTS.md`, `web/scripts/oxlint/__tests__/intl-locale.test.ts`, `web/scripts/oxlint/intl-locale.mjs`, `web/src/features/system-settings/models/__tests__/visual-billing-editor.test.tsx`, `web/src/features/system-settings/models/billing-time-fields.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/tsconfig.node.json`, `web/vitest.config.ts`
- Overlap with downstream changes: AGENTS.md, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Intl locale 自定义 lint 及开发规则
- AI recommendation: `defer` — 新增 oxlint插件检测界面语言码直接进入Intl，并更新AGENTS/格式化约束。检测本身有价值，但这批还混合治理和依赖配置，建议单独工具PR只迁lint，保留当前准确locale函数。
- AI risks: 要验证别名/透传/公共格式函数误报和现有lint版本，不照搬受保护规则文件或改变金额精度。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/scripts/oxlint/intl-locale.mjs:40](https://github.com/QuantumNous/new-api/blob/42cbe6206a62550896c8371d62a6f15bad5b3e05/web/scripts/oxlint/intl-locale.mjs#L40) — 固定源码：const interfaceLanguageProperties = new Set(['language', 'resolvedLanguage'])
- AI evidence: [downstream:web/src/i18n/languages.ts:86](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/i18n/languages.ts#L86) — 下游对照：export function toIntlLocale(value?: string | null): string | undefined {

#### commit-d8069436612e

- Title: fix(web): tidy channel health policy layout
- Feature/group: fix(web): tidy channel health policy layout
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/request-policies/__tests__/channel-health-layout.test.tsx`, `web/src/features/system-settings/request-policies/channel-health-section.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道健康策略配置布局
- AI recommendation: `defer` — 只重排上游新增 request-policies/channel-health 设置页；下游尚无该同路径设置页。待后台健康策略能力评审完，再决定UI映射。
- AI risks: 不要为布局补丁顺带启用未知健康/自动禁用策略，需与现有渠道监测规则对齐。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/system-settings/request-policies/channel-health-section.tsx:52](https://github.com/QuantumNous/new-api/blob/d8069436612e08bd0d96169222b10031556d09a7/web/src/features/system-settings/request-policies/channel-health-section.tsx#L52) — 固定源码：SettingsControlGroup,
- AI evidence: [downstream:web/src/features/channels/components/drawers/channel-mutate-drawer.tsx:621](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/drawers/channel-mutate-drawer.tsx#L621) — 下游对照：export function ChannelMutateDrawer({

#### commit-ae249f4ecbaf

- Title: feat(responses-ws): extend channel support and share routing with HTTP
- Feature/group: feat(responses-ws): extend channel support and share routing with HTTP
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/relay.go`, `controller/responses_websocket_test.go`, `middleware/distributor.go`, `middleware/distributor_test.go`, `model/channel_constraint.go`, `relay/channel/api_request.go`, `relay/channel/api_request_test.go`, `relay/responses_websocket.go`, `relay/responses_websocket_test.go`, `relaykit/dto/channel_settings.go`, `service/channel_select.go`
- Overlap with downstream changes: controller/relay.go, middleware/distributor.go, relaykit/dto/channel_settings.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: HTTP/WS 共用渠道路由
- AI recommendation: `defer` — 上游将分发器的锁定渠道、过滤和重试选择移入共享service，供HTTP与新WS使用。本地分发器叠加Volc Native、原任务凭据、资产租户等约束，必须先完成WS与任务约束全链路设计。
- AI risks: 共享路由重构不能丢原任务渠道锁定、禁用渠道检查和task plugin身份过滤；暂无等价性测试证据。
- AI dependencies: commit-9fe0457ee1f4, commit-73a471f3a8b1, commit-75f3d246e2b1
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/relay.go:169](https://github.com/QuantumNous/new-api/blob/ae249f4ecbaf3ae9b1fc62cf377ca5a764b9a002/controller/relay.go#L169) — 上游把使用渠道记录移到共享service。
- AI evidence: [downstream:controller/relay.go:230](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/relay.go#L230) — 本地仍调用原控制器路由辅助函数。

#### commit-fa3cc1c6b9b1

- Title: feat(web): offer the Responses WebSocket toggle for advanced custom, sub2api and new-api channels
- Feature/group: feat(web): offer the Responses WebSocket toggle for advanced custom, sub2api and new-api channels
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/components/drawers/channel-provider-picker.tsx`, `web/src/features/channels/components/responses-websocket-setting.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/responses-websocket.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/constants.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses WS 渠道开关界面
- AI recommendation: `defer` — 此UI依赖尚未引入的ResponsesWebSocketSetting和共享WS路由；本地没有channel-quick-options等新组件。应在WS功能被选择后按本地表单结构适配，不能显示无效开关。
- AI risks: 必须覆盖AdvancedCustom无转换路由限制、Sub2API/NewAPI、默认关闭、保存不丢值及多语言。
- AI dependencies: commit-9fe0457ee1f4, commit-ae249f4ecbaf
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/channels/components/responses-websocket-setting.tsx:41](https://github.com/QuantumNous/new-api/blob/fa3cc1c6b9b13f7bf25c3e896d439185d326ac9c/web/src/features/channels/components/responses-websocket-setting.tsx#L41) — 上游按渠道能力控制开关显示。
- AI evidence: [downstream:web/src/features/channels/lib/channel-form.ts:656](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/lib/channel-form.ts#L656) — 本地表单序列化还没有Responses WS设置。

#### commit-32ef6d216f9f

- Title: feat(web): add channel passthrough controls and usage guidance
- Feature/group: feat(web): add channel passthrough controls and usage guidance
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/channel-quick-options.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/components/dialogs/passthrough-warning-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/__tests__/header-passthrough.test.ts`, `web/src/features/channels/lib/header-passthrough.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 透传开关与使用风险说明
- AI recommendation: `migrate` — 本地已有header_override和pass_through_body_enabled；上游新增快捷开关与启用确认，说明模型映射/转换会被绕过。建议只迁入确认交互与说明，适配现有抽屉，不引入尚无的WS快捷栏。
- AI risks: 头透传可能转发敏感认证信息，确认不能代替现有服务端过滤；切换必须保留原JSON规则并支持全部语言。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/components/dialogs/passthrough-warning-dialog.tsx, web/src/features/channels/lib/header-passthrough.ts, web/src/features/channels/lib/__tests__/header-passthrough.test.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/zh.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json
- AI evidence: [upstream:web/src/features/channels/components/channel-quick-options.tsx:55](https://github.com/QuantumNous/new-api/blob/32ef6d216f9fa0a2aa182bd8aad9a02daf3767c8/web/src/features/channels/components/channel-quick-options.tsx#L55) — 启用透传增加明确确认回调。
- AI evidence: [downstream:web/src/features/channels/lib/channel-form.ts:886](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/lib/channel-form.ts#L886) — 本地已有可复用的头覆盖持久化。

#### commit-9e905a83e1d6

- Title: refactor: show subscription deductions and improve cost display (#7272)
- Feature/group: refactor: show subscription deductions and improve cost display (#7272)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/usage-logs/lib/columns.ts`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 日志订阅扣减及费用显示
- AI recommendation: `defer` — 上游将subscription_consumed作为订阅显示金额并统一费用组件；下游日志经过有效调账投影且保留工具附加费和财务权限，不能直接改为原始订阅字段。先定义最终费用来源，再合并后续160/161视觉改进。
- AI risks: 原始金额覆盖修正投影会导致账单/日志金额不一致；保留管理员饱和及用户隐私提示。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/usage-logs/components/log-cost-display.tsx:82](https://github.com/QuantumNous/new-api/blob/9e905a83e1d6c369b3aa43de4f04ab244cce3205/web/src/features/usage-logs/components/log-cost-display.tsx#L82) — 上游订阅显示优先使用原始subscription_consumed；代码：? (props.other?.subscription_consumed ?? props.quota)
- AI evidence: [downstream:web/src/features/usage-logs/components/log-cost-display.tsx:136](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/usage-logs/components/log-cost-display.tsx#L136) — 本地订阅展示沿用传入的有效quota与下游组件；代码：<SubscriptionBadge quota={props.quota} />

#### commit-ed7c4e35d429

- Title: feat(log): flag upstream response model mismatches (#7418)
- Feature/group: feat(log): flag upstream response model mismatches (#7418)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_responses.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/helper.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/responses_via_chat.go`, `relay/common/relay_info.go`, `relay/common/response_model.go`, `relay/response_model_test.go`, `service/log_info_generate.go`, `service/relay_error.go`, `service/responses_usage.go`, `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/features/usage-logs/components/__tests__/usage-facts.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/common-log-mobile-card.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relay/channel/openai/helper.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游响应模型偏差诊断
- AI recommendation: `defer` — 新增多协议ObserveResponseModel和日志UI共32文件，本地还没有该诊断对象。应与后续“由模型名推导偏差”修复一起设计，先明确公开名/映射名/供应商返回名关系及管理员可见范围。
- AI risks: 观测模型只能诊断，不能改下游输出、客户计费或成本快照；供应商快照/前缀名易误报。
- AI dependencies: commit-0ed497f066a6, commit-a5b40663d330
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/claude/relay-claude.go:327](https://github.com/QuantumNous/new-api/blob/ed7c4e35d4297962991ac5f00eea504ceac0fbd0/relay/channel/claude/relay-claude.go#L327) — 上游在原生响应转换前记录实际模型。
- AI evidence: [downstream:relay/channel/claude/relay-claude.go:217](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/claude/relay-claude.go#L217) — 本地原生Claude响应尚无模型偏差观测。

#### commit-e2bb05f70ed6

- Title: fix(relay): map Images output_tokens_details into img_o billing (#7410)
- Feature/group: fix(relay): map Images output_tokens_details into img_o billing (#7410)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/image_stream_test.go`, `relay/channel/openai/relay_image.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 图片输出模态用量收费
- AI recommendation: `migrate` — 本地normalizeOpenAIUsage只映射输入明细，未保留Images output_tokens_details，表达式img_o可能读不到。建议独立补齐Usage可选输出明细字段及映射，不为此引入整个WS（该字段原在WS提交中新增）。
- AI risks: 需以仅img_o、混合c/img_o、显式零、无明细、流/非流响应验证不重复计费，保留下游quota饱和审计。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/dto/openai_response.go, relaykit/dto/openai_response_test.go, relay/channel/openai/relay_image.go, relay/channel/openai/image_stream_test.go, service/tiered_settle_test.go
- AI evidence: [upstream:relay/channel/openai/relay_image.go:116](https://github.com/QuantumNous/new-api/blob/e2bb05f70ed6b6725a83ff6f1c97dbd2c2f640d4/relay/channel/openai/relay_image.go#L116) — 上游把Images输出明细纳入统一计费。
- AI evidence: [downstream:relay/channel/openai/relay_image.go:70](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/openai/relay_image.go#L70) — 本地归一化缺少输出明细映射。

#### commit-55a6cd2a44ea

- Title: fix(responses): estimate usage for streams cut before terminal usage (#7427)
- Feature/group: fix(responses): estimate usage for streams cut before terminal usage (#7427)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `service/responses_usage.go`, `service/responses_usage_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses 截断流用量估算
- AI recommendation: `defer` — 上游新的ResponsesUsageAccumulator将工具参数、推理和拒绝delta纳入估算，并对已启动但无usage流计入prompt；本地没有这个累加器。需与协议失败终态及本地退款/成本凭证共同核对，不能孤立套入。
- AI risks: 客户端断连与上游明确失败收费不同；须保留实报优先、估算标记、预扣/结算安全和不可变账务证据。
- AI dependencies: commit-9fe0457ee1f4, commit-8b2c7105f91f
- AI suggested paths (not approval):
- AI evidence: [upstream:service/responses_usage.go:102](https://github.com/QuantumNous/new-api/blob/55a6cd2a44eaf189bbd565dd2cdfa591d19b06d6/service/responses_usage.go#L102) — 上游根据是否启动及失败结果决定缺失prompt的估算。
- AI evidence: [downstream:relay/responses_handler.go:23](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/responses_handler.go#L23) — 本地Responses仍走现有handler结算链。

#### commit-3524fe0b1579

- Title: fix issue template (#7441)
- Feature/group: fix issue template (#7441)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.github/ISSUE_TEMPLATE/bug_report.yml`, `.github/ISSUE_TEMPLATE/bug_report_en.yml`, `.github/ISSUE_TEMPLATE/feature_request.yml`, `.github/ISSUE_TEMPLATE/feature_request_en.yml`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游 Issue 模板表单修复
- AI recommendation: `skip` — 只将上游issue提示改为必填textarea，均为.github受保护内容；与下游运行时功能无关。
- AI risks: 不能为了使上游模板一致而改变本地下游问题提交流程。
- AI dependencies: commit-57791e31c62b
- AI suggested paths (not approval):
- AI evidence: [upstream:.github/ISSUE_TEMPLATE/bug_report.yml:14](https://github.com/QuantumNous/new-api/blob/3524fe0b15794d8d19378827d36a7edc0b0e91ea/.github/ISSUE_TEMPLATE/bug_report.yml#L14) — 固定源码：id: read-first

#### commit-2906e4f779b7

- Title: fix(oauth): require account evidence before migrating legacy GitHub bindings
- Feature/group: fix(oauth): require account evidence before migrating legacy GitHub bindings
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/auth_flow_test.go`, `controller/auth_session_test.go`, `controller/login_verification.go`, `controller/model_list_test.go`, `controller/oauth.go`, `controller/passkey_test.go`, `controller/user.go`, `controller/wechat.go`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `model/account_security.go`, `model/login_verification.go`, `model/user.go`, `oauth/github.go`, `oauth/provider.go`, `service/login_verification.go`
- Overlap with downstream changes: controller/model_list_test.go, controller/user.go, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, model/user.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 旧 GitHub 绑定迁移的账号证据
- AI recommendation: `defer` — 上游不再仅凭可变 login 名自动认领旧 GitHub 绑定，增加账号证据与登录挑战内延迟迁移。本地下游仍保留 legacy_id；需优先独立安全设计并结合本地已验证手机和微信身份，不能移植整套登录重构。
- AI risks: 涉及账号接管风险与旧用户可登录性；验证证据、冲突、无邮箱、两步验证完成前不落库必须测试。
- AI dependencies: commit-45c3fbe8aeb0, commit-6f2333990613
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/oauth.go:453](https://github.com/QuantumNous/new-api/blob/2906e4f779b715f282ae11203211dca77051d5af/controller/oauth.go#L453) — 固定源码：reason, matched := "no_matching_evidence", false
- AI evidence: [downstream:oauth/github.go:158](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/oauth/github.go#L158) — 下游对照："legacy_id": githubUser.Login, // Store login for migration from old accounts

#### commit-3e8c358da09b

- Title: fix: forward reasoning_effort to zhipu upstream (#7124)
- Feature/group: fix: forward reasoning_effort to zhipu upstream (#7124)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/zhipu_4v/relay-zhipu_v4.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 智谱 reasoning_effort 保留
- AI recommendation: `merge` — 本地requestOpenAI2Zhipu重建请求时拷贝THINKING却漏了ReasoningEffort；上游增加单字段拷贝。可优先完整采用，补模型映射及空值/显式effort回归。
- AI risks: 保留原生THINKING与effort并存的供应商验证，不从模型名猜测或自动改写力度。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/channel/zhipu_4v/relay-zhipu_v4.go, relay/channel/zhipu_4v/adaptor_test.go
- AI evidence: [upstream:relay/channel/zhipu_4v/relay-zhipu_v4.go:54](https://github.com/QuantumNous/new-api/blob/3e8c358da09b4598a946712db8219ff44f197475/relay/channel/zhipu_4v/relay-zhipu_v4.go#L54) — 上游拷贝客户端reasoning_effort。
- AI evidence: [downstream:relay/channel/zhipu_4v/relay-zhipu_v4.go:53](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/zhipu_4v/relay-zhipu_v4.go#L53) — 本地请求重建仅保留THINKING。

#### commit-946c730ed008

- Title: perf: avoid repeated Claude tool-name lookups (#7447)
- Feature/group: perf: avoid repeated Claude tool-name lookups (#7447)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/dto/claude.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude 工具名查找缓存
- AI recommendation: `migrate` — 本地每个无名tool_result都会遍历整个请求查找tool_use；上游一次建表并回填结果。建议仅迁入转换器缓存与行为测试，先保留独立relaykit公开的SearchToolNameByToolCallId方法，避免无必要API删除。
- AI risks: 重复tool id应保持原先第一次匹配语义；结果在调用之前/之后、空名称及不匹配ID均需确定性测试。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go, relaykit/relayconvert/internal/claude_messages/to_oai_chat_req_test.go
- AI evidence: [upstream:relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go:161](https://github.com/QuantumNous/new-api/blob/946c730ed00806650fef509c74a868dda2847f7c/relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go#L161) — 上游在一次转换中复用工具名映射。
- AI evidence: [downstream:relaykit/dto/claude.go:373](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/dto/claude.go#L373) — 本地公共方法每次重新遍历历史消息。

#### commit-ea336ed6654a

- Title: fix(log): hide funding source labels when subscriptions are not in play (#7450)
- Feature/group: fix(log): hide funding source labels when subscriptions are not in play (#7450)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/usage-logs/lib/__tests__/billing-source.test.ts`, `web/src/features/usage-logs/lib/billing-source.ts`, `web/src/features/usage-logs/lib/columns.ts`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 仅在使用订阅时显示扣费来源
- AI recommendation: `defer` — 新增用户/日志上下文判断是否显示钱包或订阅标签，需与订阅金额及费用展示改造一起选。当前本地没有showBillingSource参数且有专属费用设计，不能直接带入中间版本。
- AI risks: 取订阅状态需有界缓存且不能泄露他人信息；金额仍应使用下游有效投影。
- AI dependencies: commit-9e905a83e1d6
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/usage-logs/components/log-cost-display.tsx:84](https://github.com/QuantumNous/new-api/blob/ea336ed6654a97c0c27e004581578231e0d92824/web/src/features/usage-logs/components/log-cost-display.tsx#L84) — 是否展示来源由上下文显式控制；代码：const billingSource = props.showBillingSource

#### commit-96b0da2273a1

- Title: refactor(log): 恢复费用气泡并用图标区分订阅来源 (#7451)
- Feature/group: refactor(log): 恢复费用气泡并用图标区分订阅来源 (#7451)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/billing-source-visibility.test.tsx`, `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/usage-logs/lib/__tests__/billing-source.test.ts`, `web/src/features/usage-logs/lib/billing-source.ts`
- Overlap with downstream changes: web/src/features/usage-logs/components/usage-logs-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 费用提示气泡和扣费来源图标
- AI recommendation: `defer` — 恢复费用气泡，使用图标区分钱包/订阅并按用户上下文控制来源。依赖前两项展示链，与下游当前金额Badge和工具附加费标记的取舍需要产品确认，不宜作为正确性修复隐式覆盖。
- AI risks: 必须保留可访问名称、有效调账金额、工具费提示和财务权限。
- AI dependencies: commit-9e905a83e1d6, commit-ea336ed6654a
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/usage-logs/components/log-cost-display.tsx:120](https://github.com/QuantumNous/new-api/blob/96b0da2273a187b3fd9621f77b10dcebd046e2f6/web/src/features/usage-logs/components/log-cost-display.tsx#L120) — 扣费来源改用有提示的图标；代码：icon={isSubscription ? CrownIcon : Wallet01Icon}

#### commit-1ea24a050d5e

- Title: docs: refresh project README across languages
- Feature/group: docs: refresh project README across languages
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `README.en.md`, `README.fr.md`, `README.ja.md`, `README.md`, `README.zh_CN.md`, `README.zh_TW.md`
- Overlap with downstream changes: README.md, README.zh_CN.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 多语言项目 README 改版
- AI recommendation: `skip` — 上游整篇重写能力介绍、导航和合作方；本地 README 有下游产品定位、来源说明及长期差异台账，不能整体覆盖。
- AI risks: 保留 new-api/QuantumNous 署名；已采用能力可在专门文档PR中按真实下游情况描述。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:README.en.md:10](https://github.com/QuantumNous/new-api/blob/1ea24a050d5e276584d744122810b8dc57b10f0e/README.en.md#L10) — 固定源码：<a href="./README.zh_CN.md">简体中文</a> |
- AI evidence: [downstream:README.md:11](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/README.md#L11) — 下游对照：[![Upstream](https://img.shields.io/badge/upstream-QuantumNous%2Fnew--api-7357d9.svg)](https://github.com/QuantumNous/new-api)

#### commit-e04c05c2b929

- Title: fix(web): 修复小数输入限制和日志分组选项未遮挡的问题 (#7453)
- Feature/group: fix(web): 修复小数输入限制和日志分组选项未遮挡的问题 (#7453)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/ui/combobox-input.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/system-settings/models/__tests__/group-pricing-decimals.test.tsx`, `web/src/features/system-settings/models/__tests__/tool-price-validation.test.tsx`, `web/src/features/system-settings/models/group-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/tool-price-settings.tsx`, `web/src/features/usage-logs/components/__tests__/group-filter.test.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/common-logs-filter-bar.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 小数设置精度与日志下拉遮挡
- AI recommendation: `migrate` — 本地组倍率 step=0.1、工具价格 step=0.5，与上游修正前一致；建议迁入更小输入步长并按本地 combobox 层级修复遮挡，避免强行替换弹层。
- AI risks: 步长变化不等于取消非负/有限值校验；小金额精度、键盘步进、层叠与滚动需验证。
- AI dependencies: commit-8f72ecbbfd86
- AI suggested paths (not approval): web/src/features/system-settings/models/group-ratio-visual-editor.tsx, web/src/features/system-settings/models/tool-price-settings.tsx, web/src/components/ui/combobox-input.tsx, web/src/components/ui/combobox.tsx, web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/features/system-settings/models/__tests__/numeric-precision.test.tsx
- AI evidence: [upstream:web/src/features/system-settings/models/group-ratio-visual-editor.tsx:562](https://github.com/QuantumNous/new-api/blob/e04c05c2b929c0df6cc35227f860cf6e19d480c7/web/src/features/system-settings/models/group-ratio-visual-editor.tsx#L562) — 固定源码：step={0.0001}
- AI evidence: [downstream:web/src/features/system-settings/models/group-ratio-visual-editor.tsx:579](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/system-settings/models/group-ratio-visual-editor.tsx#L579) — 下游对照：step={0.1}

#### commit-b0bf2580e249

- Title: docs: restore README badge and partner placement
- Feature/group: docs: restore README badge and partner placement
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `README.en.md`, `README.fr.md`, `README.ja.md`, `README.md`, `README.zh_CN.md`, `README.zh_TW.md`
- Overlap with downstream changes: README.md, README.zh_CN.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: README 徽章与合作方排版恢复
- AI recommendation: `skip` — 只恢复上游项目徽章和伙伴顺序，没有运行时功能；README在自动同步保护范围且本地有独立首页说明。
- AI risks: 不删除或替换任何已存在的归属内容，排版需求单独审查。
- AI dependencies: commit-1ea24a050d5e
- AI suggested paths (not approval):
- AI evidence: [upstream:README.en.md:36](https://github.com/QuantumNous/new-api/blob/b0bf2580e2491869ee9fe861c12b86153c3fe5f4/README.en.md#L36) — 固定源码：<a href="https://trendshift.io/repositories/20180" target="_blank">
- AI evidence: [downstream:README.md:11](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/README.md#L11) — 下游对照：[![Upstream](https://img.shields.io/badge/upstream-QuantumNous%2Fnew--api-7357d9.svg)](https://github.com/QuantumNous/new-api)

#### commit-972aed197282

- Title: fix(log): derive response model mismatch from names instead of a stored flag (#7464)
- Feature/group: fix(log): derive response model mismatch from names instead of a stored flag (#7464)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/common/response_model.go`, `relay/response_model_test.go`, `service/log_info_generate.go`, `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/features/usage-logs/components/__tests__/usage-facts.test.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`, `web/src/features/usage-logs/lib/__tests__/response-model.test.ts`, `web/src/features/usage-logs/lib/index.ts`, `web/src/features/usage-logs/lib/response-model.ts`, `web/src/features/usage-logs/types.ts`
- Overlap with downstream changes: web/src/features/usage-logs/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 响应模型偏差动态计算
- AI recommendation: `defer` — 这是153的持久化语义修正：只保存三种模型名，按当前规则算Mismatch，新增供应商命名空间匹配。本地尚无前置观测能力，应与其一并采用最终表示，避免先存易过期bool。
- AI risks: 前后缀宽松匹配可能遮蔽真实不同模型；需明确快照、路由别名和命名空间匹配边界及历史日志兼容。
- AI dependencies: commit-ed7c4e35d429
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/common/response_model.go:32](https://github.com/QuantumNous/new-api/blob/972aed1972820389ea0b603ca58f03f846fbf790/relay/common/response_model.go#L32) — 上游按保留模型名实时计算偏差。
- AI evidence: [downstream:relay/common/relay_info.go:84](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/common/relay_info.go#L84) — 本地RelayInfo尚无ResponseModel对象。

#### commit-a89037d0b9b0

- Title: fix(web): start linked-account verification from the provider button
- Feature/group: fix(web): start linked-account verification from the provider button
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`
- Overlap with downstream changes: web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 已绑定 OAuth 提供者按钮直接启动验证
- AI recommendation: `defer` — 这是统一验证 UI 的交互修复，通过 provider 按钮直接传 override VerificationInput；本地验证 hook/页面结构不同且扩展短信验证，需随验证体系迁移。
- AI risks: 必须防重复提交/过期回调，取消后不能完成另一次敏感操作。
- AI dependencies: commit-45c3fbe8aeb0, commit-6f2333990613
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/auth/secure-verification/hooks/use-secure-verification.ts:281](https://github.com/QuantumNous/new-api/blob/a89037d0b9b065da8582959018321abbc48db549/web/src/features/auth/secure-verification/hooks/use-secure-verification.ts#L281) — 固定源码：async (override?: VerificationInput) => {
- AI evidence: [downstream:service/auth_token.go:160](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/auth_token.go#L160) — 下游对照：func VerifySecurityProof(raw string, identity AuthIdentity, requiredScope string, allowedMethods []string) (string, error) {

#### commit-ef70d1b79998

- Title: chore: drop trailing blank line from .gitignore
- Feature/group: chore: drop trailing blank line from .gitignore
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.gitignore`
- Overlap with downstream changes: .gitignore
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: gitignore 尾部空行清理
- AI recommendation: `skip` — 只有删除末尾空白行，既不修功能也不更新依赖。无需为样式差异产生迁移提交，保留本地新增测试缓存与评估计划规则。
- AI risks: 不能套用旧上游.gitignore覆盖本地规则；本项不代表已升级任何能力。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:.gitignore:46](https://github.com/QuantumNous/new-api/blob/ef70d1b79998db9bb32b6b79205afbc4b9f8716b/.gitignore#L46) — 固定源码：go.work.sum
- AI evidence: [downstream:.gitignore:49](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/.gitignore#L49) — 下游对照：go.work.sum

#### commit-b932e6d876cd

- Title: fix(web): robustly display conditional task billing prices
- Feature/group: fix(web): robustly display conditional task billing prices
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/task-matrix-display.test.ts`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/task-display.ts`, `web/src/features/pricing/lib/breakdown-tier-match.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/task-matrix-display.ts`, `web/src/features/pricing/lib/task-price-display.ts`, `web/src/features/system-settings/models/__tests__/request-simulation.test.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 条件任务价格的稳健展示
- AI recommendation: `defer` — 新增可显示/特殊表达式/无法推导价格的回落分类，统一任务价格矩阵和条件档位选择。它依赖任务usage_schema和新表达式模块，本地尚未采用，需和计费语言升级一致后迁入。
- AI risks: 无法静态推导的表达式不能显示为免费或错误档位；日志应优先用冻结的matched_tier。
- AI dependencies: commit-eb48396d5fe9, commit-d52bdc0b4087, commit-39294418af4c
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/pricing/lib/task-price-display.ts:39](https://github.com/QuantumNous/new-api/blob/b932e6d876cda4fa49528b5cdb873d9fc5e31b6f/web/src/features/pricing/lib/task-price-display.ts#L39) — 新增任务价格显示策略模块；代码：export function taskPriceLabel(

#### commit-9c3d3aeb3c30

- Title: perf(web): lazy-load provider icons and reduce page rerenders
- Feature/group: perf(web): lazy-load provider icons and reduce page rerenders
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/api-keys-columns.tsx`, `web/src/features/models/components/models-columns.tsx`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/pricing-controls.test.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/pricing-sidebar.tsx`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/columns/drawing-logs-columns.tsx`, `web/src/features/usage-logs/components/columns/task-logs-columns.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/users/components/users-columns.tsx`, `web/src/lib/__tests__/lobe-icon.test.tsx`, `web/src/lib/lobe-icon.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/columns/task-logs-columns.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 供应商图标懒加载与表格缓存
- AI recommendation: `migrate` — 本地 lobe-icon 仍整包导入 @lobehub/icons，上游改为按选中变体懒加载并缓存列/卡片，确有包体优化空间。建议先迁入图标加载，再针对本地依赖做 memo，不搬全部价格和额度列。
- AI risks: 动态 import 需验证 Rsbuild 产物、离线fallback与所有图标变体；memo依赖缺失会显示旧权限/金额。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/lib/lobe-icon.tsx, web/src/lib/__tests__/lobe-icon.test.tsx
- AI evidence: [upstream:web/src/lib/lobe-icon.tsx:164](https://github.com/QuantumNous/new-api/blob/9c3d3aeb3c30ea09c7ad0f810c86205c4da558d7/web/src/lib/lobe-icon.tsx#L164) — 固定源码：IconComponent = lazy(() =>
- AI evidence: [downstream:web/src/lib/lobe-icon.tsx:28](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/lib/lobe-icon.tsx#L28) — 下游对照：import * as LobeIcons from '@lobehub/icons'

#### commit-33320407f4f5

- Title: fix(web): isolate OIDC discovery when saving OAuth settings
- Feature/group: fix(web): isolate OIDC discovery when saving OAuth settings
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/auth/__tests__/oauth-settings.test.tsx`, `web/src/features/system-settings/auth/oauth-section.tsx`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 保存其他 OAuth 配置时隔离 OIDC discovery
- AI recommendation: `migrate` — 本地只要已填 well_known 就每次执行发现请求；不可达 OIDC 会阻塞其他提供者保存。迁入仅修改 discovery URL 或从关闭变开启时发现，并保留本地微信/短信配置。
- AI risks: 禁用 OIDC 不应依赖远端服务；提交状态、错误定位和原配置恢复需覆盖。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/system-settings/auth/oauth-section.tsx, web/src/features/system-settings/auth/__tests__/oauth-section.test.tsx
- AI evidence: [upstream:web/src/features/system-settings/auth/oauth-section.tsx:308](https://github.com/QuantumNous/new-api/blob/33320407f4f5dcf5bc635ee6d73dcc31c8bcc031/web/src/features/system-settings/auth/oauth-section.tsx#L308) — 固定源码：if (wellKnown && (discoveryUrlChanged || enablingOIDC)) {
- AI evidence: [downstream:web/src/features/system-settings/auth/oauth-section.tsx:295](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/system-settings/auth/oauth-section.tsx#L295) — 下游对照：if (values.oidc.well_known && values.oidc.well_known.trim() !== '') {

#### commit-4c34f25a49f1

- Title: feat(channel): bind multiple task plugins to New API channels
- Feature/group: feat(channel): bind multiple task plugins to New API channels
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel.go`, `controller/channel_task_plugin_bind_test.go`, `controller/channel_task_plugin_validation_test.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `middleware/distributor.go`, `middleware/distributor_test.go`, `middleware/task_plugin.go`, `model/channel_constraint.go`, `model/task.go`, `model/task_plugin.go`, `model/task_plugin_channel_select_test.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `plugins/builtin_plugins_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/auth_test.go`, `relaykit/dto/channel_settings.go`, `service/task_polling.go`, `service/task_polling_test.go`, `web/src/components/multi-select.tsx`, `web/src/components/multi-select/__tests__/option-icon.test.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/__tests__/new-api-channel.test.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/lib/channel-plugin-extensions.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: controller/channel.go, middleware/distributor.go, model/task.go, relaykit/dto/channel_settings.go, service/task_polling.go, service/task_polling_test.go, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: New API 渠道绑定多个任务插件
- AI recommendation: `defer` — 允许网关渠道绑定多个任务插件，禁用插件只解绑对应扩展而不关闭整个普通渠道。当前无插件路由，该能力会改变渠道模型匹配与重试，应先评估与本地稳定别名和权限隔离的组合。
- AI risks: 禁用插件不得误停普通中继；多插件认领相同模型必须确定价格和选择规则。
- AI dependencies: commit-eb48396d5fe9, commit-ab489ab8842b, commit-74629e29f83f
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/channel.go:541](https://github.com/QuantumNous/new-api/blob/4c34f25a49f1cf491479e3842b2251e8a6908851/controller/channel.go#L541) — 多插件键数量有上限并专用于 New API 渠道；代码：// maxTaskExtendPluginKeys bounds the plugins one New API channel can be

#### commit-2d7aef7414e8

- Title: fix(web): restore CC Switch model dropdown
- Feature/group: fix(web): restore CC Switch model dropdown
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/dialogs/__tests__/cc-switch-dialog.test.tsx`, `web/src/features/keys/components/dialogs/cc-switch-dialog.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: CC Switch 模型下拉恢复
- AI recommendation: `migrate` — 上游用标准 Combobox 替代手写 ComboboxInput 并移除额外底部留白。本地 CC Switch 对话框存在旧实现，应配合共享弹层修复迁入交互。
- AI risks: 复制导出的模型ID和值必须与用户实际选择一致；不能让自定义模型输入或多字段模式失效。
- AI dependencies: commit-387a40914853, commit-8f72ecbbfd86
- AI suggested paths (not approval): web/src/features/keys/components/dialogs/cc-switch-dialog.tsx, web/src/features/keys/components/dialogs/__tests__/cc-switch-dialog.test.tsx
- AI evidence: [upstream:web/src/features/keys/components/dialogs/cc-switch-dialog.tsx:157](https://github.com/QuantumNous/new-api/blob/2d7aef7414e8671b1ac7db336bc170eeb7cc5e2b/web/src/features/keys/components/dialogs/cc-switch-dialog.tsx#L157) — 固定源码：bodyClassName='space-y-4'
- AI evidence: [downstream:web/src/features/keys/components/dialogs/cc-switch-dialog.tsx:97](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/dialogs/cc-switch-dialog.tsx#L97) — 下游对照：export function CCSwitchDialog(props: Props) {

#### commit-6b638788c645

- Title: fix(channels): preview upstream model changes before applying
- Feature/group: fix(channels): preview upstream model changes before applying
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/upstream-update-dialog.test.tsx`, `web/src/features/channels/components/channels-dialogs.tsx`, `web/src/features/channels/components/data-table-row-actions.tsx`, `web/src/features/channels/components/dialogs/upstream-update-dialog.tsx`, `web/src/features/channels/hooks/use-channel-upstream-updates.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道上游模型变更先预览再应用
- AI recommendation: `migrate` — 上游把菜单直接使用已缓存变更改为主动 detectChannelUpdates，并把 loading/error/preview 状态交给同一 hook。本地有上游模型更新能力，建议只迁入新鲜预览和显式确认。
- AI risks: 这是渠道模型列表更新，与本次源码版本同步不同；保留下游模型命名/映射、重复请求和部分失败语义。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/channels/components/channels-dialogs.tsx, web/src/features/channels/components/data-table-row-actions.tsx, web/src/features/channels/components/dialogs/upstream-update-dialog.tsx, web/src/features/channels/hooks/use-channel-upstream-updates.ts, web/src/features/channels/hooks/__tests__/upstream-update-preview.test.tsx
- AI evidence: [upstream:web/src/features/channels/components/data-table-row-actions.tsx:302](https://github.com/QuantumNous/new-api/blob/6b638788c645a6c9164c5c6394827e3f5e3a1abd/web/src/features/channels/components/data-table-row-actions.tsx#L302) — 固定源码：onClick={() => upstream.detectChannelUpdates(channel)}
- AI evidence: [downstream:web/src/features/channels/hooks/use-channel-upstream-updates.ts:49](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/hooks/use-channel-upstream-updates.ts#L49) — 下游对照：export function useChannelUpstreamUpdates(refresh: () => Promise<void>) {

#### commit-db536ec8bcbb

- Title: fix(channels): simplify model redirects and clarify panel controls
- Feature/group: fix(channels): simplify model redirects and clarify panel controls
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/floating-window.tsx`, `web/src/components/floating-window/__tests__/floating-window.test.tsx`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-batch-dialog.test.tsx`, `web/src/features/channels/components/__tests__/upstream-model-selection.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/model-mapping-batch-dialog.tsx`, `web/src/features/channels/components/model-mapping-editor.tsx`, `web/src/features/channels/components/model-redirect-panel.tsx`, `web/src/features/channels/components/upstream-model-selection.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型重定向工作台后续收敛
- AI recommendation: `defer` — 清理上一悬浮工作台的自动命名启发式并增加展开控制，证明主项仍有后续修正；下游未迁主项，待按最终实现统一设计。
- AI risks: 避免先导入已删除的自动别名策略；保留映射方向、批量去重与持久化位置语义。
- AI dependencies: commit-0cde9d94f6d3
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/components/floating-window.tsx:66](https://github.com/QuantumNous/new-api/blob/db536ec8bcbbc63034d009a7257ecee0f3468abf/web/src/components/floating-window.tsx#L66) — 固定源码：expandRequest?: number
- AI evidence: [downstream:web/src/features/channels/components/drawers/channel-mutate-drawer.tsx:621](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/drawers/channel-mutate-drawer.tsx#L621) — 下游对照：export function ChannelMutateDrawer({

#### commit-58a9eff4c842

- Title: style: refresh logo
- Feature/group: style: refresh logo
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `electron/icon.png`, `electron/tray-icon-windows.png`, `electron/tray-iconTemplate.png`, `electron/tray-iconTemplate@2x.png`, `web/index.html`, `web/public/favicon.ico`, `web/public/logo.png`, `web/rsbuild.config.ts`
- Overlap with downstream changes: web/index.html, web/public/favicon.ico, web/public/logo.png, web/rsbuild.config.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游图标与 favicon 品牌资源刷新
- AI recommendation: `skip` — 二进制图标、tray图标和favicon路径变更属于上游品牌呈现，下游保留自己的界面与发布资源，不把品牌替换作为版本同步。
- AI risks: 所有现存 new-api/QuantumNous 归属必须保留；不得导入上游资源覆盖下游品牌和OSS资源规则。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/rsbuild.config.ts:67](https://github.com/QuantumNous/new-api/blob/58a9eff4c8421a6bbb04855e561310d3eeabd107/web/rsbuild.config.ts#L67) — 固定源码：favicon: './public/favicon.ico',
- AI evidence: [downstream:web/index.html:10](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/index.html#L10) — 下游对照：<title>New API</title>

#### commit-3fdf9083dc6b

- Title: fix(settings): move documentation link to site settings
- Feature/group: fix(settings): move documentation link to site settings
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/billing/index.tsx`, `web/src/features/system-settings/billing/section-registry.tsx`, `web/src/features/system-settings/general/quota-settings-section.tsx`, `web/src/features/system-settings/general/system-info-section.tsx`, `web/src/features/system-settings/site/index.tsx`, `web/src/features/system-settings/site/section-registry.tsx`, `web/src/features/system-settings/types.ts`
- Overlap with downstream changes: web/src/features/system-settings/billing/section-registry.tsx, web/src/features/system-settings/general/system-info-section.tsx, web/src/features/system-settings/site/index.tsx, web/src/features/system-settings/site/section-registry.tsx, web/src/features/system-settings/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 文档链接设置归入站点信息
- AI recommendation: `migrate` — 仅将 general_setting.docs_link 从计费设置移动到站点设置，本地仍置于 quota schema。可按现有保存边界迁移字段，不改变后端 option key。
- AI risks: 旧配置值必须继续显示且保存不清空；重置只影响所属 section，避免覆盖其他站点/计费设置。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/system-settings/billing/index.tsx, web/src/features/system-settings/billing/section-registry.tsx, web/src/features/system-settings/general/quota-settings-section.tsx, web/src/features/system-settings/general/system-info-section.tsx, web/src/features/system-settings/site/index.tsx, web/src/features/system-settings/site/section-registry.tsx, web/src/features/system-settings/types.ts
- AI evidence: [upstream:web/src/features/system-settings/general/system-info-section.tsx:90](https://github.com/QuantumNous/new-api/blob/3fdf9083dc6bfc47fc2ad0de3fac327d749d778d/web/src/features/system-settings/general/system-info-section.tsx#L90) — 固定源码：docs_link: normalizeValue(defaultValues.general_setting?.docs_link),
- AI evidence: [downstream:web/src/features/system-settings/general/quota-settings-section.tsx:60](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/system-settings/general/quota-settings-section.tsx#L60) — 下游对照：docs_link: z.string(),

#### commit-da2540dda132

- Title: fix(channels): align card metrics and remove badge spacing
- Feature/group: fix(channels): align card metrics and remove badge spacing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/channel-card.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道移动卡片指标排列
- AI recommendation: `defer` — 重排渠道卡片 priority/response、weight/last-tested 并统一 badge context。属于视觉偏好且本地下游卡片还有额外监控信息，先确认布局需求。
- AI risks: 不能用压缩布局隐藏余额、状态原因和敏感字段权限；需实机/窄屏验证。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/channels/components/channel-card.tsx:89](https://github.com/QuantumNous/new-api/blob/da2540dda1328040193f279a58c3e315083e416c/web/src/features/channels/components/channel-card.tsx#L89) — 固定源码：className='flex min-w-0 flex-col gap-3'
- AI evidence: [downstream:web/src/features/channels/components/channel-card.tsx:42](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/channel-card.tsx#L42) — 下游对照：function ChannelCardComponent({

#### commit-d904b9700015

- Title: feat(settings): improve group configuration UX and drag sorting
- Feature/group: feat(settings): improve group configuration UX and drag sorting
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/auto-group-order-item.tsx`, `web/src/features/keys/components/auto-group-order-editor.tsx`, `web/src/features/system-settings/models/__tests__/group-settings.test.tsx`, `web/src/features/system-settings/models/group-ratio-form.tsx`, `web/src/features/system-settings/models/group-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/group-special-usable-editor.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 自动分组顺序拖动与配置 UX
- AI recommendation: `defer` — 新增 AutoGroupOrderItem 并重构组倍率表单；自动分组优先级影响路由，需确认当前下游可用组/特殊倍率与键盘重排规则，避免把表现优化混入业务变更。
- AI risks: 拖动后保存顺序必须稳定；对不可用/重复组和继承倍率不能静默重写。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/components/auto-group-order-item.tsx:38](https://github.com/QuantumNous/new-api/blob/d904b9700015ad73285b0f0030c88eb6683fbeda/web/src/components/auto-group-order-item.tsx#L38) — 固定源码：onMove: (index: number, direction: 'up' | 'down') => void
- AI evidence: [downstream:web/src/features/keys/components/auto-group-order-editor.tsx:157](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/auto-group-order-editor.tsx#L157) — 下游对照：export function AutoGroupOrderEditor(props: AutoGroupOrderEditorProps) {

#### commit-585c4610e5d7

- Title: feat(keys): show and copy configured API addresses
- Feature/group: feat(keys): show and copy configured API addresses
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/__tests__/api-addresses.test.tsx`, `web/src/features/keys/components/api-keys-primary-buttons.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: API key 页面展示和复制 API 地址
- AI recommendation: `migrate` — 上游复用后台 API 地址配置和 CopyButton，并在未配置时回退当前域名。本地 keys 页面缺少该入口，可迁入这一小功能，保留自有代理/多地址配置语义。
- AI risks: 不能复制管理台 URL 为错误 API 前缀；检查自定义地址、空配置、加载态和超长地址。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/keys/components/api-keys-primary-buttons.tsx, web/src/features/keys/components/__tests__/api-addresses.test.tsx
- AI evidence: [upstream:web/src/features/keys/components/api-keys-primary-buttons.tsx:49](https://github.com/QuantumNous/new-api/blob/585c4610e5d7103480620ab61954db5a5feac35b/web/src/features/keys/components/api-keys-primary-buttons.tsx#L49) — 固定源码：const serverAddress =
- AI evidence: [downstream:web/src/features/keys/components/api-keys-primary-buttons.tsx:26](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/api-keys-primary-buttons.tsx#L26) — 下游对照：export function ApiKeysPrimaryButtons() {

#### commit-c49972441e73

- Title: fix(web): keep mobile dialog actions within visible viewport
- Feature/group: fix(web): keep mobile dialog actions within visible viewport
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/dialog.tsx`, `web/src/components/ui/__tests__/dialog-layout.test.tsx`, `web/src/components/ui/alert-dialog.tsx`, `web/src/components/ui/dialog.tsx`, `web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx`, `web/src/features/channels/components/dialogs/channel-test-dialog.tsx`, `web/src/features/channels/components/dialogs/edit-tag-dialog.tsx`, `web/src/features/channels/components/dialogs/multi-key-manage-dialog.tsx`, `web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx`, `web/src/features/channels/components/dialogs/upstream-update-dialog.tsx`, `web/src/features/model-pricing/pricing-conversion-dialog.tsx`, `web/src/features/models/components/dialogs/missing-models-dialog.tsx`, `web/src/features/models/components/dialogs/update-config-dialog.tsx`, `web/src/features/models/components/dialogs/view-details-dialog.tsx`, `web/src/features/system-settings/auth/custom-oauth/components/provider-form-dialog.tsx`, `web/src/features/system-settings/models/channel-selector-dialog.tsx`, `web/src/features/system-settings/models/conflict-confirm-dialog.tsx`, `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/wallet/components/dialogs/billing-history-dialog.tsx`, `web/src/styles/index.css`
- Overlap with downstream changes: web/src/features/channels/components/dialogs/channel-test-dialog.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/styles/index.css
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 移动端软键盘下对话框操作可见
- AI recommendation: `migrate` — 上游以100dvh与safe-area CSS变量计算对话框中心和可用高度，本地仍基于100vh固定最大高度。建议迁入最终 viewport 算法，与紧随后续fallback修复一起验收。
- AI risks: 需要 iOS/Android 软键盘、窗口缩放、嵌套弹窗、长表单footer以及CSS变量未定义和旧浏览器的fallback验证。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/components/dialog.tsx, web/src/components/ui/dialog.tsx, web/src/components/ui/alert-dialog.tsx, web/src/styles/index.css, web/src/components/ui/__tests__/dialog-layout.test.tsx
- AI evidence: [upstream:web/src/components/dialog.tsx:74](https://github.com/QuantumNous/new-api/blob/c49972441e73f05207427567a2d9ad7c620db4bd/web/src/components/dialog.tsx#L74) — 固定源码：'flex max-h-(--dialog-available-height) w-full flex-col gap-4 overflow-x-hidden overflow-y-auto p-4 sm:max-w-2xl sm:p-6',
- AI evidence: [downstream:web/src/components/dialog.tsx:74](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/components/dialog.tsx#L74) — 下游对照：'flex max-h-[calc(100vh-2rem)] w-full flex-col gap-4 overflow-hidden p-4 sm:max-w-2xl sm:p-6',

#### commit-03563a4a7916

- Title: feat(plugins): serve OpenAI Images API through task plugins
- Feature/group: feat(plugins): serve OpenAI Images API through task plugins
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.agents/rules/billing.md`, `common/model.go`, `constant/channel.go`, `controller/model_management_test.go`, `controller/plugin_protocol.go`, `controller/plugin_protocol_image.go`, `controller/plugin_protocol_test.go`, `controller/relay.go`, `controller/relay_task_plugin_test.go`, `controller/task.go`, `controller/task_generic_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `dto/task.go`, `middleware/task_plugin.go`, `middleware/task_plugin_test.go`, `model/model_pricing_conversion.go`, `model/task.go`, `pkg/billingexpr/expr.md`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`, `plugins/alibaba_responses_test.go`, `plugins/builtin_plugins_test.go`, `plugins/doubao_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `relay/channel/ali/adaptor.go`, `relay/channel/ali/adaptor_test.go`, `relay/channel/ali/dto.go`, `relay/channel/ali/image.go`, `relay/channel/ali/image_wan.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/convert_request_error_test.go`, `relay/helper/billing_expr_request.go`, `relay/helper/openai_image_request_test.go`, `relay/helper/price.go`, `relay/helper/valid_request.go`, `relay/image_handler.go`, `relay/relay_task.go`, `relay/relay_task_test.go`, `router/relay-router.go`, `router/task-plugin-protocol-router.go`, `router/task_plugin_protocol_router_test.go`, `service/image_billing.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/text_quota_test.go`, `setting/model_setting/qwen.go`, `web/public/favicon.ico`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/provider-selection.test.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/channel-utils.ts`, `web/src/features/task-plugins/__tests__/plugin-detail-sheet.test.tsx`, `web/src/features/task-plugins/lib/host-protocols.ts`, `web/src/features/usage-logs/__tests__/artifacts.test.ts`, `web/src/features/usage-logs/components/__tests__/billing-source-visibility.test.tsx`, `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/__tests__/task-artifacts-cell.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/common-log-mobile-card.tsx`, `web/src/features/usage-logs/components/dialogs/audio-preview-dialog.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`, `web/src/features/usage-logs/components/task-artifacts.tsx`, `web/src/features/usage-logs/components/timing-metrics-cell.tsx`, `web/src/features/usage-logs/lib/task-artifacts.ts`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: constant/channel.go, controller/relay.go, controller/task.go, dto/task.go, model/task.go, relay/relay_task.go, router/relay-router.go, service/task_billing.go, service/task_billing_test.go, web/public/favicon.ico, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-utils.ts, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件提供 OpenAI Images 协议
- AI recommendation: `defer` — 扩展任务插件承接同步/异步图片协议、文件占位、下载与产物返回，对现有图片中继是架构性分流。需要先验证 multipart、图片数量、response_format、流式计费和授权内容访问，再决定与 Go 图片适配器共存方式。
- AI risks: 不能跳过图片数量预扣、重试补预留和对象存储权限；不应导入 .agents 规则文件。
- AI dependencies: commit-eb48396d5fe9, commit-f064bffa2b65, commit-74629e29f83f, commit-4c34f25a49f1
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/plugin_protocol_image.go:30](https://github.com/QuantumNous/new-api/blob/03563a4a791629a811076f1cbcbecfbe2716eac2/controller/plugin_protocol_image.go#L30) — 新增独立的插件图片协议处理入口；代码：func serveTaskPluginImageProtocol(c *gin.Context, pinned pluginruntime.PinnedEndpoint, deps pluginProtocolBridgeDeps) {

#### commit-a5bf94022060

- Title: fix(web): correct dialog viewport fallbacks and editor gutter alignment
- Feature/group: fix(web): correct dialog viewport fallbacks and editor gutter alignment
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/ai-elements/__tests__/code-block-editor.test.tsx`, `web/src/components/ai-elements/code-block.tsx`, `web/src/components/dialog.tsx`, `web/src/components/ui/__tests__/dialog-layout.test.tsx`, `web/src/components/ui/alert-dialog.tsx`, `web/src/components/ui/dialog.tsx`, `web/src/features/system-settings/models/__tests__/group-pricing-decimals.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 对话框 viewport fallback 与编辑器行号
- AI recommendation: `migrate` — 修正前项 CSS 变量没有设置时的fallback，另去掉 CodeMirror 行号多余纵向padding。应随移动对话框主项采用最终状态，不保留中间无fallback实现。
- AI risks: viewport来自CSS变量而非新增JS hook；保留旧浏览器fallback与代码行号对齐。
- AI dependencies: commit-c49972441e73
- AI suggested paths (not approval): web/src/components/dialog.tsx, web/src/components/ui/dialog.tsx, web/src/components/ui/alert-dialog.tsx, web/src/components/ai-elements/code-block.tsx, web/src/components/ui/__tests__/dialog-viewport.test.tsx
- AI evidence: [upstream:web/src/components/ui/dialog.tsx:73](https://github.com/QuantumNous/new-api/blob/a5bf940220604e5b9d150e2338bf6528dfbb8be4/web/src/components/ui/dialog.tsx#L73) — 固定源码：'bg-popover text-popover-foreground ring-foreground/10 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95 fixed top-[var(--dialog-viewport-center,50dvh)] left-1/2 z-50 grid max-h-[var(--dialog-available-height,calc(100dvh-2rem))] w-full overflow-y-auto overscroll-contain max-w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 gap-4 rounded-xl p-4 text-sm ring-1 duration-100 outline-none sm:max-w-sm',
- AI evidence: [downstream:web/src/components/ui/dialog.tsx:73](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/components/ui/dialog.tsx#L73) — 下游对照：'bg-popover text-popover-foreground ring-foreground/10 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95 fixed top-1/2 left-1/2 z-50 grid w-full max-w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 gap-4 rounded-xl p-4 text-sm ring-1 duration-100 outline-none sm:max-w-sm',

#### commit-2fddb863f221

- Title: feat(models): add bulk field selection to metadata sync
- Feature/group: feat(models): add bulk field selection to metadata sync
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/models/__tests__/metadata-sync.test.tsx`, `web/src/features/models/components/dialogs/sync-wizard-dialog.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型元数据同步批量选字段
- AI recommendation: `defer` — 现有模型允许按字段批量确认覆盖，新模型展示的全部元数据直接导入。当前没有前述新元数据同步API和可见性状态，待统一模型管理方案选定后纳入。
- AI risks: 未选字段必须保持原样；新增模型默认导入的敏感/协议字段需要明确。
- AI dependencies: commit-0c76e4dae77a, commit-71c1fd7caad7
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/models/components/dialogs/sync-wizard-dialog.tsx:204](https://github.com/QuantumNous/new-api/blob/2fddb863f2219d9eacea5b899ee39b46f11adf33/web/src/features/models/components/dialogs/sync-wizard-dialog.tsx#L204) — 批量修改只面向选中既有模型；代码：const selectedExistingModels = selected.filter(

#### commit-3abbb8198e04

- Title: feat: record perf metrics when async tasks reach terminal status (#7481)
- Feature/group: feat: record perf metrics when async tasks reach terminal status (#7481)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `pkg/perf_metrics/metrics.go`, `pkg/perf_metrics/metrics_test.go`, `service/task_polling.go`
- Overlap with downstream changes: service/task_polling.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 异步任务终态性能指标
- AI recommendation: `defer` — 上游把终态采样与计费/失败退款统一放入finalizeTerminalTask并声明CAS赢家恰好一次；下游任务轮询包含实时快照、归档、取消与成本凭证扩展，不能只搬helper就假设幂等。
- AI risks: 先证明状态CAS、重复轮询、主动取消、补偿结算和退款仍恰好一次，再接入现有分类监控口径。
- AI dependencies: commit-a5b40663d330
- AI suggested paths (not approval):
- AI evidence: [upstream:service/task_polling.go:626](https://github.com/QuantumNous/new-api/blob/3abbb8198e0468ac6e16eb0103c1d3642e2e7668/service/task_polling.go#L626) — 上游在终态统一采样、结算和退款。
- AI evidence: [downstream:service/task_polling.go:442](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/task_polling.go#L442) — 本地下游视频任务终态处理是现有扩展流程。

#### commit-65d3a2171fd1

- Title: fix(plugins): price only the resolutions each Seedance and Wan model offers
- Feature/group: fix(plugins): price only the resolutions each Seedance and Wan model offers
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/alibaba_responses_test.go`, `plugins/doubao_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Seedance/Wan 模型分辨率价格档
- AI recommendation: `defer` — 按实际型号能力缩小可报价的分辨率与视频输入维度，避免展示或收取不支持的档位。下游已有 Seedance 2.5 的独立分档计费和调账证据，应在同一能力矩阵中比对并迁移缺口，暂不导入整个插件体系。
- AI risks: 不能用上游新能力表抹掉下游公开别名和历史档位成本证据。
- AI dependencies: commit-eb48396d5fe9, commit-a20574136b27
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/doubao/plugin.js:14](https://github.com/QuantumNous/new-api/blob/65d3a2171fd11c65ad2e2e1f1d65723926659c81/plugins/tasks/doubao/plugin.js#L14) — 上游为 Seedance2.5 限定分辨率和视频输入能力；代码："doubao-seedance-2-5-260628": { resolutions: ["480p", "720p", "1080p"], videoInput: true },
- AI evidence: [downstream:relay/channel/task/doubao/adaptor.go:154](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/task/doubao/adaptor.go#L154) — 下游已通过独立能力表校验 Seedance2.5 档位；代码：if _, ok := seedancepricing.Lookup(req.Model, resolution, false); !ok {

#### commit-9978ee1e25a6

- Title: feat(billing): configure trust threshold and input pre-consume multiplier
- Feature/group: feat(billing): configure trust threshold and input pre-consume multiplier
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/constants.go`, `common/quota.go`, `controller/billing_option_test.go`, `model/option.go`, `pkg/billingexpr/expr.md`, `pkg/billingexpr/types.go`, `relay/helper/price.go`, `relay/helper/price_test.go`, `service/billing_session.go`, `service/image_billing.go`, `setting/operation_setting/quota_setting.go`, `web/src/features/system-settings/billing/index.tsx`, `web/src/features/system-settings/billing/section-registry.tsx`, `web/src/features/system-settings/general/__tests__/pre-consume-settings.test.tsx`, `web/src/features/system-settings/general/quota-settings-section.tsx`, `web/src/features/system-settings/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: common/constants.go, model/option.go, service/billing_session.go, web/src/features/system-settings/billing/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 预扣信任门槛及仅输入估算
- AI recommendation: `defer` — 该变更除了可配信任门槛，还把文本预扣改为输入费用×倍率并将预计输出C设为0，是信用风险/并发额度策略变化。下游已有付费组输出估算和钱包持久结算，不建议在维护同步中直接放宽；须由业务明确接受风险后单独迁移。
- AI risks: 长输出/并发请求可能预留不足；必须验证组重试、免费转付费、补预留和退款，不能只看最终结算正确。
- AI dependencies: commit-064ed943e1ac, commit-f064bffa2b65
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/helper/price.go:370](https://github.com/QuantumNous/new-api/blob/9978ee1e25a647bfe004e96c8719a2cb62c24732/relay/helper/price.go#L370) — 表达式预扣不再估算输出Token；代码：C:   0,
- AI evidence: [downstream:relay/helper/price.go:277](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/helper/price.go#L277) — 下游对付费组未给max_tokens的请求仍保留输出估算；代码：estimatedCompletionTokens = defaultTieredPreConsumeMaxTokens

#### commit-47713bcb1749

- Title: feat: update plugins
- Feature/group: feat: update plugins
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件版本元数据更新
- AI recommendation: `defer` — 该提交只将 Jimeng/Kling/Sora/SunoAPI 的版本号更新到 1.1.0，没有独立运行时修复；如果未来整套驱动迁移，应保留最终版本及升级关系，不应为本地无插件驱动单独提交版本。
- AI risks: 单独改版本会宣称未实际实现的能力。
- AI dependencies: commit-eb48396d5fe9, commit-74629e29f83f
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/jimeng/plugin.js:10](https://github.com/QuantumNous/new-api/blob/47713bcb1749f81c3a48127c0150e4c6d4da35f1/plugins/tasks/jimeng/plugin.js#L10) — 此提交的实际变动是插件版本号；代码：version: "1.1.0",

#### commit-9a0be8750a6d

- Title: fix(web): keep task pricing editable after a plugin narrows its enum values
- Feature/group: fix(web): keep task pricing editable after a plugin narrows its enum values
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/task-expr.test.ts`, `web/src/features/pricing/__tests__/task-matrix-display.test.ts`, `web/src/features/pricing/__tests__/task-matrix.test.ts`, `web/src/features/pricing/lib/billing-expression/README.md`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/task-display.ts`, `web/src/features/system-settings/models/__tests__/task-pricing-copy.test.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件收窄枚举后的旧价格可编辑性
- AI recommendation: `defer` — 把不可能命中的旧枚举分支标记为 unreachable 并跳过显示，避免插件能力收窄后整个表达式无法编辑。依赖新的表达式解析器与任务 usage_schema，当前旧解析器没有对应契约。
- AI risks: 必须区分无法命中和解析失败，保留未改旧价格的业务语义；不能直接删除保存表达式分支。
- AI dependencies: commit-d52bdc0b4087, commit-b932e6d876cd, commit-65d3a2171fd1
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/pricing/lib/billing-expression/display.ts:356](https://github.com/QuantumNous/new-api/blob/9a0be8750a6d736d9692535ed2cd68f8eec46529/web/src/features/pricing/lib/billing-expression/display.ts#L356) — 解析结果区分不可达分支和无效表达式；代码：return reachable ? conditions : 'unreachable'

#### commit-9c293e8c0237

- Title: test(billing): fix websocket rejection refund test after pre-consume dropped output estimation (#7491)
- Feature/group: test(billing): fix websocket rejection refund test after pre-consume dropped output estimation (#7491)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/responses_websocket_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 仅输入预扣的WebSocket测试调整
- AI recommendation: `defer` — 此项只改Responses WebSocket退款测试中的估算前提，以匹配上游取消输出预留。下游未采用WebSocket入口和新预扣策略，不能移植测试来迫使生产逻辑接受不同信用策略。
- AI risks: 测试必须保护实际付费/退款行为，不能仅为对齐上游断言修改正确的本地策略。
- AI dependencies: commit-9fe0457ee1f4, commit-9978ee1e25a6
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/responses_websocket_test.go:429](https://github.com/QuantumNous/new-api/blob/9c293e8c02371bda844af79e3500ff2d516d1dda/controller/responses_websocket_test.go#L429) — 测试改用固定请求价格触发预留和退款；代码：fixture := newResponsesWSBillingTest(t, `tier("request", fixed(0.002))`, func(ws *websocket.Conn, _ *http.Request) {

#### commit-474ed66fb14b

- Title: fix(plugins/alibaba): keep upstream metadata in openai_image render (#7507)
- Feature/group: fix(plugins/alibaba): keep upstream metadata in openai_image render (#7507)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/alibaba_responses_test.go`, `plugins/tasks/alibaba/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 阿里图片保留上游元数据
- AI recommendation: `defer` — OpenAI Images 渲染增加 metadata 原始上游对象，依赖插件图片协议。未来移植时需要审查响应暴露字段与租户敏感数据，当前 Go 阿里图片中继不具备此任务视图接口。
- AI risks: 不能直接把含凭据/内部错误的整份上游对象返回用户。
- AI dependencies: commit-eb48396d5fe9, commit-03563a4a7916
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/alibaba/plugin.js:1469](https://github.com/QuantumNous/new-api/blob/474ed66fb14b0de30f82a1729b4f7cf3dab19d63/plugins/tasks/alibaba/plugin.js#L1469) — 图片协议响应附带原始任务数据；代码：return { created: task.created_at, data: data, metadata: artifactData(task) };

#### commit-00e8a00cb4ae

- Title: fix(web): include provider icons in Windows builds
- Feature/group: fix(web): include provider icons in Windows builds
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/lib/__tests__/lobe-icon.test.tsx`, `web/src/lib/lobe-icon.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 懒加载图标 Windows 构建路径兼容
- AI recommendation: `migrate` — 上游为 webpackInclude 同时接受斜杠与反斜杠；本地尚是整包导入，必须随懒加载主项使用最终跨平台正则并验证产物。
- AI risks: 单测regex匹配不能代替 Windows Rsbuild 真构建；检查组件变体和fallback，避免动态包漏图标。
- AI dependencies: commit-9c3d3aeb3c30
- AI suggested paths (not approval): web/src/lib/lobe-icon.tsx, web/src/lib/__tests__/lobe-icon.test.tsx
- AI evidence: [upstream:web/src/lib/lobe-icon.tsx:167](https://github.com/QuantumNous/new-api/blob/00e8a00cb4aebb4e029b5ff32b8fe1e98f0b0bc5/web/src/lib/lobe-icon.tsx#L167) — 固定源码：/* webpackInclude: /[\\/]components[\\/](Mono|Avatar|Brand|BrandColor|Color|Combine|Text|TextCn|TextColor|Simple|Morden)\.js$/ */
- AI evidence: [downstream:web/src/lib/lobe-icon.tsx:28](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/lib/lobe-icon.tsx#L28) — 下游对照：import * as LobeIcons from '@lobehub/icons'

#### commit-e537dc380bb2

- Title: fix(web): expose plugin activation after upload
- Feature/group: fix(web): expose plugin activation after upload
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上传插件后的激活入口
- AI recommendation: `defer` — 上传成功后提供明确激活操作、状态反馈与焦点，解决上传完成却未启用的交互缺口。只在插件生命周期与权限审批存在后才有意义。
- AI risks: 上传成功不能隐式等于受信任并可立即执行；激活仍须管理员权限。
- AI dependencies: commit-eb48396d5fe9, commit-984330920061
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/task-plugins/components/upload-dialog.tsx:90](https://github.com/QuantumNous/new-api/blob/e537dc380bb253d4f902ac033ab39dd2ca462b5e/web/src/features/task-plugins/components/upload-dialog.tsx#L90) — 上传成功后的激活由独立 mutation 处理；代码：const activateMutation = useMutation({

#### commit-c0cff23a3af8

- Title: fix(plugins): store task plugin source as longtext on MySQL and raise the upload cap to 8 MiB
- Feature/group: fix(plugins): store task plugin source as longtext on MySQL and raise the upload cap to 8 MiB
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`, `model/task_plugin.go`, `web/src/features/task-plugins/__tests__/plugin-url.test.ts`, `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/plugin-source-picker.tsx`, `web/src/features/task-plugins/components/plugin-url-import-field.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/features/task-plugins/lib/plugin-url.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件源码存储容量与上传上限
- AI recommendation: `defer` — 新增按方言返回 MySQL longtext、其他数据库 text 的类型并统一 8MiB 上限，以修复源码超过 TEXT 容量。现有仓库无 task_plugins 表，迁入引擎时必须连同三数据库存储和体积测试一起采用。
- AI risks: MySQL schema 升级与回滚、图标体积、编译内存/CPU 和并发上传均需有界。
- AI dependencies: commit-eb48396d5fe9, commit-984330920061
- AI suggested paths (not approval):
- AI evidence: [upstream:model/task_plugin.go:71](https://github.com/QuantumNous/new-api/blob/c0cff23a3af8e68809647bd39ec204bedbea28af/model/task_plugin.go#L71) — 存储类型按数据库方言选择；代码：func (LongText) GormDBDataType(db *gorm.DB, _ *schema.Field) string {

#### commit-2c175190c44a

- Title: perf(plugins): sync task plugins from hashes and load source only for changed rows
- Feature/group: perf(plugins): sync task plugins from hashes and load source only for changed rows
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`, `model/task_plugin.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件按哈希增量同步
- AI recommendation: `defer` — 节点轮询只读取活动插件元数据及哈希，发生变化再取源码编译，减少大字段重复读取。依赖插件持久化和注册器，须随架构方案测试多节点版本切换与失败保留。
- AI risks: 哈希不变不能遗漏禁用/版本切换；新源码编译失败不能污染当前生效集合。
- AI dependencies: commit-eb48396d5fe9, commit-c0cff23a3af8
- AI suggested paths (not approval):
- AI evidence: [upstream:model/task_plugin.go:188](https://github.com/QuantumNous/new-api/blob/2c175190c44a95c6af4877e901dc5211ef706979/model/task_plugin.go#L188) — 周期轮询查询排除大源码和图标列；代码：if err := DB.Omit("source", "icon").Where(&TaskPlugin{Active: true}).

#### commit-d61d6be7593f

- Title: fix(task): accept any 2xx upstream status on task submission
- Feature/group: fix(task): accept any 2xx upstream status on task submission
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.md`, `relay/relay_task.go`, `relay/relay_task_test.go`
- Overlap with downstream changes: relay/relay_task.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 任务提交 HTTP 成功码兼容
- AI recommendation: `migrate` — 建议优先把成功码判定迁移到现有 Go 任务提交链：当前只有 HTTP200 才进入 ParseResponse，201/202 合法异步提交被误报失败。只改状态分类并增加本地 Go 适配器回归，不引入插件层。
- AI risks: 非2xx继续保留原错误状态和正文；解析失败、空响应、扣费退款及任务快照行为必须保持。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/relay_task.go, relay/relay_task_test.go
- AI evidence: [upstream:relay/relay_task.go:358](https://github.com/QuantumNous/new-api/blob/d61d6be7593f99eacf9462e99ee13f0c432657f5/relay/relay_task.go#L358) — 任意2xx进入既有任务响应解析；代码：if resp.StatusCode/100 != 2 {
- AI evidence: [downstream:relay/relay_task.go:263](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/relay_task.go#L263) — 本地仍把201/202当提交失败，缺口可独立修复；代码：if resp != nil && resp.StatusCode != http.StatusOK {

#### commit-54eee488bed3

- Title: fix(web): isolate theme preferences in local storage
- Feature/group: fix(web): isolate theme preferences in local storage
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/context/__tests__/theme-preferences.test.tsx`, `web/src/context/theme-customization-provider.tsx`, `web/src/context/theme-provider.tsx`, `web/src/lib/frontend-cache.ts`, `web/src/lib/theme-customization.ts`, `web/src/lib/theme-storage.ts`
- Overlap with downstream changes: web/src/lib/frontend-cache.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 主题偏好从共享 Cookie 移到本地存储
- AI recommendation: `migrate` — 本地 theme-provider/customization 通过 cookie 保存，上游改为 localStorage 并在清缓存时保留主题。可迁入以减少同父域不同站点互相影响，保留下游默认主题。
- AI risks: 旧cookie兼容/清除、首次闪烁、存储禁用和缓存升级需覆盖；不能清掉认证/其他账号数据。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/context/theme-provider.tsx, web/src/context/theme-customization-provider.tsx, web/src/lib/theme-storage.ts, web/src/lib/frontend-cache.ts, web/src/lib/theme-customization.ts, web/src/lib/__tests__/theme-storage.test.ts
- AI evidence: [upstream:web/src/context/theme-provider.tsx:82](https://github.com/QuantumNous/new-api/blob/54eee488bed3ec6b9659b3a792194b18c2dfad27/web/src/context/theme-provider.tsx#L82) — 固定源码：readThemePreference(storageKey, THEMES, defaultTheme)
- AI evidence: [downstream:web/src/context/theme-provider.tsx:73](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/context/theme-provider.tsx#L73) — 下游对照：function getStoredTheme(storageKey: string, fallback: Theme): Theme {

#### commit-6e9de44a7dca

- Title: fix(relayconvert): hoist Responses tool-output media into a user message (#7510)
- Feature/group: fix(relayconvert): hoist Responses tool-output media into a user message (#7510)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses 工具结果媒体提升
- AI recommendation: `migrate` — 本地function_call_output仍把结构化媒体交给文本转换；上游把图片/文件/音频/视频提升为随后一个user消息，保持同批tool消息连续。建议独立迁入媒体分离，结合后续图片URL对象修复。
- AI risks: 不能将Base64送文本tokenizer；未知类型须保留原行为，多个工具输出顺序及媒体SSRF/下载边界必须回归。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go, relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go
- AI evidence: [upstream:relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go:175](https://github.com/QuantumNous/new-api/blob/6e9de44a7dca08f44234ab9c623f6a93a174d0c4/relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go#L175) — 上游延迟媒体user消息以保持tool批次连续。
- AI evidence: [downstream:relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go:193](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go#L193) — 本地工具输出仍直接转成tool内容。

#### commit-4eb3b9160566

- Title: fix(relayconvert): hoist Claude tool_result images and segment Responses reasoning items (#7512)
- Feature/group: fix(relayconvert): hoist Claude tool_result images and segment Responses reasoning items (#7512)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude 工具图片与 Responses 分段
- AI recommendation: `migrate` — 本地Claude tool_result的图片被序列化为JSON文本，建议只选择性迁入“图片提升到user消息”及测试。本提交同时改Responses思考/文本分段状态机，应保留为单独大组，不能声称整提交已采用。
- AI risks: 保留工具批次连续、URL/DataURI语义与未知block回退；Responses分段后续需要完整终止事件/usage矩阵才能另行选择。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go, relaykit/relayconvert/internal/claude_messages/to_oai_chat_req_test.go
- AI evidence: [upstream:relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go:219](https://github.com/QuantumNous/new-api/blob/4eb3b9160566dc4c1f4f0e9340de254f7febf056/relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go#L219) — 上游拆出工具结果媒体避免Base64当文本。
- AI evidence: [downstream:relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go:190](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go#L190) — 本地会把结构化图片直接JSON序列化为tool文本。

#### commit-b6809a52df88

- Title: fix(web): resolve safe constants in plugin metadata previews
- Feature/group: fix(web): resolve safe constants in plugin metadata previews
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-meta-preview.test.ts`, `web/src/features/task-plugins/lib/plugin-meta-preview.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件源码预览中的安全常量解析
- AI recommendation: `defer` — 静态预览只解析不逃逸、不被修改的本地 const，避免运行插件就能展示元数据。依赖市场源码预览，没有插件 UI 时不应新增 JS 语法分析面。
- AI risks: 必须持续禁止执行源代码、getter 或任意函数；常量引用循环及别名修改需测试。
- AI dependencies: commit-eb48396d5fe9, commit-984330920061
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/task-plugins/lib/plugin-meta-preview.ts:89](https://github.com/QuantumNous/new-api/blob/b6809a52df88e8db5c8591b2d3788a05401271b7/web/src/features/task-plugins/lib/plugin-meta-preview.ts#L89) — 预览限定不可变局部常量；代码：/** Only local constants whose containers cannot escape or be mutated are previewable. */

#### commit-0aec08fee811

- Title: chore: credit #7444 author for the Claude tool_result fix landed in #7512
- Feature/group: chore: credit #7444 author for the Claude tool_result fix landed in #7512
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: (none)
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude 工具图片作者致谢
- AI recommendation: `skip` — 这是为前面#7512修复补充作者归属的空提交，git树无文件差异，不存在可实施功能。应在实际迁移198时保存来源与作者关系，无需生成运行时代码提交。
- AI risks: 不得因空提交跳过真实工具图片修复，也不得删除已有new-api/QuantumNous或贡献者署名。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go:219](https://github.com/QuantumNous/new-api/blob/0aec08fee811ec6136828fda790551b49e410301/relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go#L219) — 本空提交的树沿用前述工具图片修复；仅用于关联来源。

#### commit-9310231b3c27

- Title: fix(logs): keep model badge text readable and compact (#7504)
- Feature/group: fix(logs): keep model badge text readable and compact (#7504)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 日志模型徽标换行可读性
- AI recommendation: `defer` — 上游对新增 wrapText 模式减少padding并加 leading-5；下游同一徽标仍没有 wrapText 属性，需随日志移动布局方案适配，不能单独照搬这两行。
- AI risks: 需检查长模型名、图标、两行截断和复制，不能让真实模型名与重定向名称混淆。
- AI dependencies: commit-a5e41a893379
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/usage-logs/components/model-badge.tsx:82](https://github.com/QuantumNous/new-api/blob/9310231b3c27fea933e939b46cf26e0ce67192e3/web/src/features/usage-logs/components/model-badge.tsx#L82) — 固定源码：? 'line-clamp-2 leading-5 [overflow-wrap:anywhere]'
- AI evidence: [downstream:web/src/features/usage-logs/components/model-badge.tsx:150](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/usage-logs/components/model-badge.tsx#L150) — 下游对照：<span className='whitespace-nowrap'>{props.modelName}</span>

#### commit-8cb88ebd17a0

- Title: fix: move force operation into confirm slot so plugin-in-use dialog has one cancel button (#7528)
- Feature/group: fix: move force operation into confirm slot so plugin-in-use dialog has one cancel button (#7528)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/plugins-table.test.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件被占用时强制操作确认
- AI recommendation: `defer` — 调整停用/删除确认弹窗，把强制操作置于确认动作，避免重复取消按钮。当前没有插件占用关系与页面，应和插件管理功能一起引入。
- AI risks: 强制解绑不能删除历史任务的必需执行/查询版本。
- AI dependencies: commit-eb48396d5fe9, commit-984330920061, commit-4c34f25a49f1
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/task-plugins/components/plugins-table.tsx:448](https://github.com/QuantumNous/new-api/blob/8cb88ebd17a072691f39705ae72c786f8bb5c4e0/web/src/features/task-plugins/components/plugins-table.tsx#L448) — 确认界面统一阻止并发停用/删除动作；代码：isLoading={statusMutation.isPending || deleteMutation.isPending}

#### commit-996adffe5165

- Title: fix(web): keep text selection when opening a model pricing row for editing (#7403)
- Feature/group: fix(web): keep text selection when opening a model pricing row for editing (#7403)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/models/__tests__/model-ratio-table-selection.test.tsx`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 定价表编辑保留文本选择
- AI recommendation: `migrate` — 建议把正在编辑模型名通过ref读取，使删除回调和列定义不因打开编辑行而重建。本地handleDelete依赖editData，确有相同单元格重挂载链；应移植稳定引用及回归测试，保留本地下游批量调价逻辑。
- AI risks: ref只用于读取最新编辑身份，不能造成删除后错误关闭别的模型编辑器或陈旧保存。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/system-settings/models/__tests__/model-ratio-table-selection.test.tsx, web/src/features/system-settings/models/model-ratio-visual-editor.tsx
- AI evidence: [upstream:web/src/features/system-settings/models/model-ratio-visual-editor.tsx:169](https://github.com/QuantumNous/new-api/blob/996adffe5165bd5e311e33a03a86b8aede1fe376/web/src/features/system-settings/models/model-ratio-visual-editor.tsx#L169) — 使用ref解除列回调对editData的依赖；代码：const editingModelNameRef = useRef<string | null>(null)
- AI evidence: [downstream:web/src/features/system-settings/models/model-ratio-visual-editor.tsx:415](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/system-settings/models/model-ratio-visual-editor.tsx#L415) — 本地删除回调仍闭包依赖当前editData并影响列memo；代码：if (editData?.name === name) {

#### commit-5401874c6387

- Title: fix(ci): align ldflags -X package path with module path so release binaries embed the real version (#7536)
- Feature/group: fix(ci): align ldflags -X package path with module path so release binaries embed the real version (#7536)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.github/workflows/electron-build.yml`, `.github/workflows/release.yml`
- Overlap with downstream changes: .github/workflows/electron-build.yml, .github/workflows/release.yml
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游二进制 ldflags 模块路径修正
- AI recommendation: `skip` — 只修上游 release/electron-build 工作流中的 new-api/common.Version；下游采用自己的构建流程，并在 Dockerfile.dev 使用正确完整模块路径。保护CI不通过本功能迁入。
- AI risks: 若现存某个下游发布入口遗漏，应另行按真实流程修，不重建上游release工作流。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:.github/workflows/electron-build.yml:58](https://github.com/QuantumNous/new-api/blob/5401874c6387b52149b7e5b0f61d9dafea5f5b7a/.github/workflows/electron-build.yml#L58) — 固定源码：#     go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(git describe --tags)' -extldflags '-static'" -o new-api
- AI evidence: [downstream:Dockerfile.dev:22](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/Dockerfile.dev#L22) — 下游对照：RUN go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" -o new-api

#### commit-a46f045d6d14

- Title: fix: match current price column width to source columns so expressions wrap alike (#7542)
- Feature/group: fix: match current price column width to source columns so expressions wrap alike (#7542)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/models/upstream-ratio-sync-columns.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游价格对照列等宽
- AI recommendation: `defer` — 调整新上游价格同步表的当前价/来源价宽度及复选框缩进。当前没有upstream-price-cells新组件，需要先选定模型价格管理重构；不应导入孤立样式片段。
- AI risks: 只影响新价格同步页展示，不能误解为本次上游代码版本同步页面。
- AI dependencies: commit-0c76e4dae77a, commit-0e0ba152bdcc
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/system-settings/models/upstream-ratio-sync-columns.tsx:64](https://github.com/QuantumNous/new-api/blob/a46f045d6d146b1e7e17616974bbf37e32885850/web/src/features/system-settings/models/upstream-ratio-sync-columns.tsx#L64) — 对齐当前价格和来源价格的布局；代码：// Offset source cells' checkbox (pl-7) and ui/table's checkbox pr-0 (-mr-3).

#### commit-6c14c0762dc7

- Title: feat: highlight differing values in source price expressions (#7543)
- Feature/group: feat: highlight differing values in source price expressions (#7543)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/models/upstream-price-cells.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 表达式来源价格差异高亮
- AI recommendation: `defer` — 对相同形状的价格表达式按词位置高亮数值差异，供价格同步人工判断。依赖新价格单元格组件，不能把字符串高亮当计费表达式语义等价结论；待新定价页面确定后采用。
- AI risks: 不同形状回落为原文本；用户仍需审查实际收费和并发保存。
- AI dependencies: commit-0c76e4dae77a, commit-a46f045d6d14
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/system-settings/models/upstream-price-cells.tsx:96](https://github.com/QuantumNous/new-api/blob/6c14c0762dc7f3bc77d0c80585f974e65ac2a571/web/src/features/system-settings/models/upstream-price-cells.tsx#L96) — 高亮只在词数量相同的表达式形状上工作；代码：function highlightExprDiff(expr: string, base: unknown) {

#### commit-c76452d22412

- Title: fix(playground): add break-all to playground input (#6600)
- Feature/group: fix(playground): add break-all to playground input (#6600)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/ai-elements/prompt-input.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Playground 长字符串输入换行
- AI recommendation: `migrate` — 只给共享 PromptInputTextarea 加 break-all，本地同组件仍缺少长连续字符串断行；可小范围迁入并确认代码/URL输入可编辑。
- AI risks: 检查超长 token、中文、光标/滚动与移动键盘；不改变真正发送内容。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/components/ai-elements/prompt-input.tsx
- AI evidence: [upstream:web/src/components/ai-elements/prompt-input.tsx:892](https://github.com/QuantumNous/new-api/blob/c76452d22412be45eb708f2618748f7aac46703d/web/src/components/ai-elements/prompt-input.tsx#L892) — 固定源码：className={cn('field-sizing-content max-h-48 min-h-16 break-all', className)}
- AI evidence: [downstream:web/src/components/ai-elements/prompt-input.tsx:892](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/components/ai-elements/prompt-input.tsx#L892) — 下游对照：className={cn('field-sizing-content max-h-48 min-h-16', className)}

#### commit-d04c118c8803

- Title: fix(channel): preserve Responses WebSocket setting on save (#7468)
- Feature/group: fix(channel): preserve Responses WebSocket setting on save (#7468)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/responses-websocket-setting.test.tsx`, `web/src/features/channels/lib/channel-form.ts`
- Overlap with downstream changes: web/src/features/channels/lib/channel-form.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses WS 设置保存修复
- AI recommendation: `defer` — 该提交让设置序列化与supportsResponsesWebSocket共用渠道能力判断，修复新渠道保存丢开关；本地尚未有WS设置。若选择WS，须作为150的必需后续一起迁入。
- AI risks: 只更新显示、不更新保存会静默关闭；需覆盖AdvancedCustom/Sub2API/NewAPI原值保留和默认值。
- AI dependencies: commit-9fe0457ee1f4, commit-fa3cc1c6b9b1
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/channels/lib/channel-form.ts:663](https://github.com/QuantumNous/new-api/blob/d04c118c8803f49e0c9bab74dcf5b5efeab9464a/web/src/features/channels/lib/channel-form.ts#L663) — 上游持久化使用统一能力判断。
- AI evidence: [downstream:web/src/features/channels/lib/channel-form.ts:656](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/lib/channel-form.ts#L656) — 本地没有Responses WS序列化项。

#### commit-c2b7a9a9e0b5

- Title: fix(claude): preserve per-message output_config in Claude messages (#7561)
- Feature/group: fix(claude): preserve per-message output_config in Claude messages (#7561)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/claude/adaptor_test.go`, `relaykit/dto/claude.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude 消息级 output_config
- AI recommendation: `migrate` — 本地ClaudeMessage仅role/content，解析再发会丢每消息output_config；上游以RawMessage可选字段保留。可独立迁入DTO和完整往返测试，限定原生Claude行为，不伪造跨协议能力。
- AI risks: 非空/空/省略值须保持，消息级beta effort不应被顶层effort覆盖；Moonshot facade继续拒绝无法真实映射语义。; 新增DTO字段后要核对下游Moonshot Messages→OpenAI路径，无法保真时明确拒绝或给出真实转换诊断，不能静默声称保留。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/dto/claude.go, relaykit/dto/claude_test.go, relay/facade/moonshot/facade.go, relay/facade/moonshot/facade_test.go
- AI evidence: [upstream:relaykit/dto/claude.go:140](https://github.com/QuantumNous/new-api/blob/c2b7a9a9e0b548c2051a949fceabb59029adcb49/relaykit/dto/claude.go#L140) — 上游新增消息级原始配置字段。
- AI evidence: [downstream:relaykit/dto/claude.go:121](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/dto/claude.go#L121) — 本地消息结构缺少该可选字段。

#### commit-789c970199ea

- Title: fix(claude): preserve auto mode safeguards in native requests (#7598)
- Feature/group: fix(claude): preserve auto mode safeguards in native requests (#7598)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/dto/claude.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude 原生 safeguards 字段
- AI recommendation: `migrate` — 本地ClaudeRequest缺少safeguards，上游新增RawMessage字段以保留auto mode保护设置。建议独立迁入原生请求DTO与往返测试，避免安全控制在网关被静默剥离。
- AI risks: 仅透传客户端明确提供的配置；不为其他协议合成或降低安全控制，缺失字段仍应省略。; 普通OpenAI数据源不支持Claude safeguards；Moonshot Messages兼容入口需验证拒绝/诊断语义，不得悄悄删除安全设置。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/dto/claude.go, relaykit/dto/claude_test.go, relay/facade/moonshot/facade.go, relay/facade/moonshot/facade_test.go
- AI evidence: [upstream:relaykit/dto/claude.go:244](https://github.com/QuantumNous/new-api/blob/789c970199ea527e6a26e071915f4a4cd2c64178/relaykit/dto/claude.go#L244) — 上游保留原生safeguards。
- AI evidence: [downstream:relaykit/dto/claude.go:222](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/dto/claude.go#L222) — 本地相邻原生控制字段存在但无safeguards。

#### commit-f5e54ccdb203

- Title: fix(web): revalidate marketplace fetches so stale caches do not fail integrity checks
- Feature/group: fix(web): revalidate marketplace fetches so stale caches do not fail integrity checks
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/marketplace-panel.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace-plugin-logo.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-icon-file.test.ts`, `web/src/features/task-plugins/__tests__/plugin-url.test.ts`, `web/src/features/task-plugins/components/marketplace-panel.tsx`, `web/src/features/task-plugins/lib/plugin-icon-file.ts`, `web/src/features/task-plugins/lib/plugin-url.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件市场缓存与完整性校验
- AI recommendation: `defer` — 市场索引、源码和图标请求使用 no-cache 再验证，避免浏览器陈旧内容触发哈希失败。未来市场上线时应采用，不能用放宽哈希校验掩盖缓存错误。
- AI risks: 必须保留完整性失败即拒绝安装；离线及重定向来源策略需明确。
- AI dependencies: commit-eb48396d5fe9, commit-984330920061
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/task-plugins/lib/plugin-icon-file.ts:105](https://github.com/QuantumNous/new-api/blob/f5e54ccdb203e8d4469035e72cc48dfd25b4314d/web/src/features/task-plugins/lib/plugin-icon-file.ts#L105) — 图标请求要求重新验证缓存；代码：response = await fetchImpl(url, { cache: 'no-cache' })

#### commit-87bb71e7fc83

- Title: feat(plugins): serve Seedream images through the OpenAI Images protocol in doubao
- Feature/group: feat(plugins): serve Seedream images through the OpenAI Images protocol in doubao
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/doubao_responses_test.go`, `plugins/tasks/doubao/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Seedream 的 OpenAI Images 兼容
- AI recommendation: `defer` — 在 Doubao 插件新增 openai_image 解码、multipart 文件引用和与原生 Ark 一致的校验/计费。下游当前独立 VolcNative 图片路由和对象存储边界需保留，是否扩展 OpenAI Images 入口应单独确定。
- AI risks: 不可互通越权原生/兼容路由；文件体积、n、base64、签名地址和真实计价必须一致。
- AI dependencies: commit-eb48396d5fe9, commit-03563a4a7916
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/doubao/plugin.js:281](https://github.com/QuantumNous/new-api/blob/87bb71e7fc83121fc995a048077e5c0e209b789f/plugins/tasks/doubao/plugin.js#L281) — Seedream 插件声明 OpenAI Images 兼容协议；代码：{ name: "openai_image", models: Object.keys(IMAGE_MODELS) },
- AI evidence: [downstream:relay/channel/task/doubao/adaptor.go:265](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/task/doubao/adaptor.go#L265) — 下游现有素材处理在Go提交路径上，插件迁移不能覆盖；代码：data, _, err = assetLibraryService.ResolveRequestAssetIDs(data, info.UserId, info.GetChannelID())

#### commit-3509ae7c194d

- Title: fix(pricing): keep prices saved before a usage profile narrows editable
- Feature/group: fix(pricing): keep prices saved before a usage profile narrows editable
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.md`, `model/model_pricing_config.go`, `model/pricing_usage_schema_test.go`, `pkg/billingexpr/expr.md`, `web/src/features/pricing/__tests__/task-matrix.test.ts`, `web/src/features/pricing/lib/billing-expression/display.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 保留旧用量档价格的可编辑性
- AI recommendation: `defer` — 保留单一供应商未修改的旧表达式，即使插件收窄usage_schema，避免管理员编辑无关字段时被拒；共享模型仍按当前profile校验。需要插件多价格和新定价模型，现阶段不具备直接落点。
- AI risks: 保留旧价格不等于允许新建无效表达式；共享模型不能绕过当前有效用量校验。
- AI dependencies: commit-0c76e4dae77a, commit-74629e29f83f, commit-9a0be8750a6d
- AI suggested paths (not approval):
- AI evidence: [upstream:model/model_pricing_config.go:387](https://github.com/QuantumNous/new-api/blob/3509ae7c194dcc0011cc4ec20cd495b198a7ed6d/model/model_pricing_config.go#L387) — 按未修改旧表达式及模型共享状态决定兼容；代码：if previousVariants[key] == expression && (!declared || !generation.SharedModel(name)) {

#### commit-6cbb1c7edba4

- Title: feat(auth): require step-up verification for admin user management
- Feature/group: feat(auth): require step-up verification for admin user management
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/admin_user_verification_test.go`, `controller/custom_oauth.go`, `controller/passkey.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `controller/twofa.go`, `controller/user.go`, `controller/user_manage_test.go`, `docs/authentication.md`, `middleware/secure_verification.go`, `service/security_verification.go`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/users/api.ts`, `web/src/features/users/components/__tests__/permissions.test.tsx`, `web/src/features/users/components/__tests__/step-up-verification.test.tsx`, `web/src/features/users/components/data-table-row-actions.tsx`, `web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`, `web/src/features/users/components/users-delete-dialog.tsx`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/features/users/components/users-provider.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: controller/secure_verification.go, controller/user.go, controller/user_manage_test.go, web/src/features/auth/secure-verification/types.ts, web/src/features/users/api.ts, web/src/features/users/components/data-table-row-actions.tsx, web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/features/users/components/users-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 管理员用户敏感操作二次验证
- AI recommendation: `defer` — 上游为改密码/权限/身份绑定/停用2FA等操作增加目标绑定证明。本地已有手机管理证明、微信解绑和管理员权限矩阵，必须合并为一致策略，而非把现有专用证明绕开。
- AI risks: 应绑定目标用户、提供者和新权限内容，拒绝跨用户重放；额度普通编辑与安全操作边界需明确。
- AI dependencies: commit-45c3fbe8aeb0, commit-0973dc2b8f55, commit-6f2333990613
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/secure_verification.go:140](https://github.com/QuantumNous/new-api/blob/6cbb1c7edba4d634ba637a6e6f111115d6de9c9a/controller/secure_verification.go#L140) — 固定源码：func requireAdminUserProof(c *gin.Context, scope string, context any) *model.AuthFlowAuthorization {
- AI evidence: [downstream:controller/user.go:768](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/user.go#L768) — 下游对照：if !commitManagedPhoneMutation(c, request, updatedUser.Phone, originUser, save) {

#### commit-caca52f8dea0

- Title: feat(auth): replace the dashboard access token with scoped access tokens
- Feature/group: feat(auth): replace the dashboard access token with scoped access tokens
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/admin_user_verification_test.go`, `controller/audit.go`, `controller/auth_session.go`, `controller/custom_oauth.go`, `controller/email_binding.go`, `controller/passkey.go`, `controller/passkey_test.go`, `controller/secure_verification.go`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `controller/twofa.go`, `controller/user.go`, `controller/user_manage_test.go`, `controller/wechat.go`, `docs/authentication.md`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `middleware/access_token_routes.go`, `middleware/audit.go`, `middleware/auth.go`, `middleware/auth_test.go`, `middleware/header_nav_test.go`, `middleware/secure_verification.go`, `model/audit_log.go`, `model/auth_flow.go`, `model/main.go`, `model/option.go`, `model/task_cas_test.go`, `model/user.go`, `model/user_access_token.go`, `model/user_authentication_test.go`, `model/user_update_test.go`, `router/access_token_scope_test.go`, `router/api-router.go`, `router/channel-router.go`, `router/plugin_router_test.go`, `service/access_token_scope.go`, `service/auth_cleanup.go`, `service/auth_session.go`, `service/auth_session_test.go`, `service/auth_token_test.go`, `service/security_verification.go`, `web/src/components/permission-matrix.tsx`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/security/__tests__/page.test.tsx`, `web/src/features/security/api.ts`, `web/src/features/security/components/__tests__/access-token-card.test.tsx`, `web/src/features/security/components/__tests__/access-token-create-dialog.test.tsx`, `web/src/features/security/components/__tests__/access-tokens-card.test.tsx`, `web/src/features/security/components/access-token-card.tsx`, `web/src/features/security/components/access-token-item.tsx`, `web/src/features/security/components/access-tokens-card.tsx`, `web/src/features/security/components/dialogs/access-token-create-dialog.tsx`, `web/src/features/security/components/dialogs/access-token-dialog.tsx`, `web/src/features/security/hooks/use-access-token.ts`, `web/src/features/security/hooks/use-access-tokens.ts`, `web/src/features/security/index.tsx`, `web/src/features/security/lib/access-token-schema.ts`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/__tests__/viewer.test.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/audit/components/audit-log-filter-bar.tsx`, `web/src/features/usage-logs/audit/components/audit-log-viewer.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/admin-permissions.ts`
- Overlap with downstream changes: controller/audit.go, controller/secure_verification.go, controller/user.go, controller/user_manage_test.go, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, model/auth_flow.go, model/main.go, model/option.go, model/task_cas_test.go, model/user.go, model/user_authentication_test.go, model/user_update_test.go, router/api-router.go, router/channel-router.go, service/auth_cleanup.go, service/auth_session.go, service/auth_session_test.go, web/src/features/auth/secure-verification/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/features/users/components/users-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/admin-permissions.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 具名范围访问令牌替代旧控制台令牌
- AI recommendation: `defer` — 新增摘要存储 PAT、逐路由 scope、数量/有效期限制和旧令牌过渡；本地还有 miniapp、资产、财务纠错等上游没有的路由。应先列全下游路由授权矩阵及迁移/撤销计划。
- AI risks: 默认拒绝未声明路由不能封死已授权下游功能，也不能默认放行；PAT 不得混成 API 模型 key 或小程序 session。
- AI dependencies: commit-d8cb177440ce, commit-45c3fbe8aeb0, commit-a8729b5c3709, commit-0973dc2b8f55, commit-6f2333990613, commit-6cbb1c7edba4
- AI suggested paths (not approval):
- AI evidence: [upstream:model/user_access_token.go:45](https://github.com/QuantumNous/new-api/blob/caca52f8dea0436d63bacc0dd3e0a73ba7b6b094/model/user_access_token.go#L45) — 固定源码：TokenHash  string `json:"token_ref" gorm:"type:varchar(64);uniqueIndex"`
- AI evidence: [downstream:middleware/auth.go:166](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/middleware/auth.go#L166) — 下游对照：patUser, err := model.ValidateAccessToken(raw)

#### commit-335ebcff71b8

- Title: fix(relayconvert): keep Responses custom tools on Chat Completions upstreams
- Feature/group: fix(relayconvert): keep Responses custom tools on Chat Completions upstreams
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/chat_via_responses_test.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relaykit/dto/openai_response.go`, `relaykit/relayconvert/convmeta/meta.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/policy_test.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`, `relaykit/relayconvert/response_registry.go`, `relaykit/relayconvert/response_registry_test.go`
- Overlap with downstream changes: relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses custom tool 跨 Chat 保真
- AI recommendation: `defer` — 16文件引入ResponsesToolState、custom_tool_call专用输出和请求/响应复原，本地只具基础custom字段但没有状态映射。应与后续Claude/Gemini支持及失败流修复成组，先明确是否接受freeform工具降级为function的契约。
- AI risks: 命名冲突、非JSON输入、工具名分片、事件顺序和原始input必须无损；公共convmeta接口变化需独立模块兼容。
- AI dependencies: commit-0ed497f066a6, commit-8a11b4305ed1
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/common/relay_info.go:205](https://github.com/QuantumNous/new-api/blob/335ebcff71b843655bf2c64f7a9a33ca21b792ea/relay/common/relay_info.go#L205) — 上游增加请求到响应的工具还原状态。
- AI evidence: [downstream:relaykit/dto/openai_request.go:307](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/dto/openai_request.go#L307) — 本地只有基础custom请求字段，并非完整还原机制。

#### commit-2ffc59590115

- Title: feat(jsplugin): run task plugins on moejs
- Feature/group: feat(jsplugin): run task plugins on moejs
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `THIRD-PARTY-LICENSES.md`, `controller/plugin_protocol.go`, `docs/plugin-api/README.md`, `docs/plugin-api/v1.md`, `dto/task_plugin.go`, `go.mod`, `go.sum`, `middleware/task_plugin.go`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/json_state.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`, `pkg/jsplugin/utils.go`, `plugins/builtin_plugins_test.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/performance_test.go`, `relay/channel/task/jsplugin/submit_stream.go`, `relay/plugin_protocol.go`, `relay/plugin_protocol_test.go`, `service/task_plugin_view_test.go`
- Overlap with downstream changes: AGENTS.md, THIRD-PARTY-LICENSES.md, go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 任务插件 JS 引擎替换为 moejs
- AI recommendation: `defer` — 替换首版 Sobek 引擎为 moejs，连同数据转换、资源管理、编译和任务调用调整。当前既没有首版引擎也未批准插件架构，不能把新运行时当普通依赖升级；需独立安全、内存隔离及平台构建评审。
- AI risks: alpha运行时引入新的原生/平台和资源释放边界；保留原超时、并发、无任意导入等沙箱约束。
- AI dependencies: commit-eb48396d5fe9, commit-129f21b6942e
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/engine.go:18](https://github.com/QuantumNous/new-api/blob/2ffc59590115bb99d5252ceaeb3e73f06dfcd1ae/pkg/jsplugin/engine.go#L18) — 引擎实现开始依赖 moejs；代码："github.com/Calcium-Ion/moejs"

#### commit-c71eefbcd194

- Title: docs: update readme
- Feature/group: docs: update readme
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `README.en.md`, `README.fr.md`, `README.ja.md`, `README.md`, `README.zh_CN.md`, `README.zh_TW.md`, `assets/screenshots/dashboard.en.jpg`, `assets/screenshots/dashboard.zh-CN.jpg`, `assets/screenshots/models.en.jpg`, `assets/screenshots/models.zh-CN.jpg`, `assets/screenshots/plugin-marketplace.en.jpg`, `assets/screenshots/plugin-marketplace.zh-CN.jpg`, `assets/screenshots/usage-logs.en.jpg`, `assets/screenshots/usage-logs.zh-CN.jpg`
- Overlap with downstream changes: README.md, README.zh_CN.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 上游 README 截图和合作方更新
- AI recommendation: `skip` — 加入上游插件市场/模型广场等模拟截图并移除伙伴展示；下游没有同等已迁能力且README受保护。不能发布不代表当前产品的截图。
- AI risks: 保留身份归属和下游差异台账，截图文档需基于下游真实界面另行维护。
- AI dependencies: commit-1ea24a050d5e, commit-b0bf2580e249
- AI suggested paths (not approval):
- AI evidence: [upstream:README.en.md:93](https://github.com/QuantumNous/new-api/blob/c71eefbcd194ea3caca8476e399ae580b275470c/README.en.md#L93) — 固定源码：| [![Usage dashboard with simulated requests, costs, and performance metrics](assets/screenshots/dashboard.en.jpg)](assets/screenshots/dashboard.en.jpg) | [![Official plugin marketplace with simulated installation status](assets/screenshots/plugin-marketplace.en.jpg)](assets/screenshots/plugin-marketplace.en.jpg) |
- AI evidence: [downstream:README.md:11](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/README.md#L11) — 下游对照：[![Upstream](https://img.shields.io/badge/upstream-QuantumNous%2Fnew--api-7357d9.svg)](https://github.com/QuantumNous/new-api)

#### commit-2506e1b980ec

- Title: fix(auth): reject non-standard roles when creating a user
- Feature/group: fix(auth): reject non-standard roles when creating a user
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/admin_user_verification_test.go`, `controller/user.go`
- Overlap with downstream changes: controller/user.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 创建用户拒绝非标准角色值
- AI recommendation: `migrate` — 本地 CreateUser 校验字段后仅比较 role 与操作者大小，没有枚举白名单。上游补 IsValidateRole，可小范围迁入到下游 userMutationRequest 路径并保留手机所有权/权限事务。
- AI risks: 测试负数、中间值、root、普通/管理员合法值及管理员越权；不能绕过手机证明事务。
- AI dependencies: none identified
- AI suggested paths (not approval): controller/user.go, controller/user_manage_test.go
- AI evidence: [upstream:controller/user.go:1008](https://github.com/QuantumNous/new-api/blob/2506e1b980ecf67c829ef20255a16fc61c47cb55/controller/user.go#L1008) — 固定源码：if !common.IsValidateRole(user.Role) {
- AI evidence: [downstream:controller/user.go:1109](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/user.go#L1109) — 下游对照：Role:        user.Role, // 保持管理员设置的角色

#### commit-4924361aef73

- Title: feat(auth): edit the name and permissions of an access token
- Feature/group: feat(auth): edit the name and permissions of an access token
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/audit.go`, `controller/security_enrollment_test.go`, `model/user_access_token.go`, `model/user_update_test.go`, `router/api-router.go`, `service/access_token_scope.go`, `service/security_verification.go`, `web/src/components/permission-matrix.tsx`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/security/api.ts`, `web/src/features/security/components/__tests__/access-token-create-dialog.test.tsx`, `web/src/features/security/components/__tests__/access-tokens-card.test.tsx`, `web/src/features/security/components/access-token-item.tsx`, `web/src/features/security/components/access-tokens-card.tsx`, `web/src/features/security/components/dialogs/access-token-create-dialog.tsx`, `web/src/features/security/components/dialogs/access-token-edit-dialog.tsx`, `web/src/features/security/hooks/use-access-tokens.ts`, `web/src/features/security/lib/access-token-catalog.ts`, `web/src/features/security/lib/access-token-schema.ts`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/users/components/__tests__/permissions.test.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: controller/audit.go, model/user_update_test.go, router/api-router.go, web/src/features/auth/secure-verification/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 访问令牌名称和授权范围编辑
- AI recommendation: `defer` — 新增 PAT 更新请求中的可选 scopes，并绑定 TokenID+授权集合进行安全验证；本地尚无具名 PAT 模型，不能单独迁入。
- AI risks: nil scopes 与空 scopes 语义不同；授权只能缩放到操作者可用权限，审计不能含明文令牌。
- AI dependencies: commit-caca52f8dea0
- AI suggested paths (not approval):
- AI evidence: [upstream:service/security_verification.go:115](https://github.com/QuantumNous/new-api/blob/4924361aef733ed724cfdb77ccf35690d9629971/service/security_verification.go#L115) — 固定源码：type AccessTokenUpdateContext struct {
- AI evidence: [downstream:middleware/auth.go:166](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/middleware/auth.go#L166) — 下游对照：patUser, err := model.ValidateAccessToken(raw)

#### commit-ae73ef8e25b9

- Title: feat(plugins): add Seedream 5.0 flash and price each Seedream model by its own tiers
- Feature/group: feat(plugins): add Seedream 5.0 flash and price each Seedream model by its own tiers
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/doubao_responses_test.go`, `plugins/tasks/doubao/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Seedream5.0Flash 与按型号图片价格档
- AI recommendation: `defer` — 新增模型并按最小像素、输入参考图和分层拆分能力生成各自 usage_schema，避免所有型号共用不真实价格档。依赖 Seedream 插件图片协议及用量表达式，价格体系需先确认。
- AI risks: 新增模型名不能自动套用未经确认的售价/成本；存量表达式与历史价单须保留。
- AI dependencies: commit-eb48396d5fe9, commit-87bb71e7fc83
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/doubao/plugin.js:190](https://github.com/QuantumNous/new-api/blob/ae73ef8e25b957ea3291512b66664f418f33c180/plugins/tasks/doubao/plugin.js#L190) — 根据型号能力生成图片计费字段；代码：function seedreamUsageSchema(profile) {

#### commit-2035a82aeb54

- Title: feat(web): add a refresh button to the channels table
- Feature/group: feat(web): add a refresh button to the channels table
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/table-refresh.test.tsx`, `web/src/features/channels/components/channels-table.tsx`
- Overlap with downstream changes: web/src/features/channels/components/channels-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道列表显式刷新按钮
- AI recommendation: `migrate` — 上游从 useQuery 暴露 refetch 并增加刷新动作；本地列表已有查询与敏感信息切换，可组合现有工具栏实现。
- AI risks: 刷新保留筛选/页码并禁用重复请求，失败反馈走现有错误处理，不泄漏隐藏密钥。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/channels/components/channels-table.tsx, web/src/features/channels/components/__tests__/refresh.test.tsx
- AI evidence: [upstream:web/src/features/channels/components/channels-table.tsx:224](https://github.com/QuantumNous/new-api/blob/2035a82aeb5414253a728bd937d4b8f97aa99b9b/web/src/features/channels/components/channels-table.tsx#L224) — 固定源码：const { data, isLoading, isFetching, refetch } = useQuery({
- AI evidence: [downstream:web/src/features/channels/components/channels-table.tsx:205](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/channels-table.tsx#L205) — 下游对照：const { data, isLoading, isFetching } = useQuery({

#### commit-feefe09f2781

- Title: fix: isolate TLS configs across HTTP transports (#7627)
- Feature/group: fix: isolate TLS configs across HTTP transports (#7627)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/init.go`, `controller/model_sync.go`, `controller/ratio_sync.go`, `service/http_client.go`, `service/protected_fetch_client.go`, `service/waffo_pancake.go`
- Overlap with downstream changes: service/http_client.go, service/protected_fetch_client.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: HTTP Transport TLS 配置隔离
- AI recommendation: `migrate` — 本地多个Transport共享common.InsecureTLSConfig，上游改Clone避免相互污染。该提交还混入Waffo store校验，建议本批仅迁移五处TLS克隆和隔离回归；支付改动留给对应支付评估。
- AI risks: 不能把TLS隔离误当允许跳过验证；避免遗漏protected fetch/模型同步客户端，测试修改一个Transport不影响另一个。
- AI dependencies: none identified
- AI suggested paths (not approval): common/init.go, controller/model_sync.go, controller/ratio_sync.go, service/http_client.go, service/protected_fetch_client.go, service/http_client_test.go
- AI evidence: [upstream:service/http_client.go:119](https://github.com/QuantumNous/new-api/blob/feefe09f2781429e039c3f7ba88da7f29492e98c/service/http_client.go#L119) — 上游为relay Transport克隆TLS配置。
- AI evidence: [downstream:service/http_client.go:105](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/http_client.go#L105) — 本地仍共享全局TLS指针。
- AI evidence: [upstream:service/waffo_pancake.go:173](https://github.com/QuantumNous/new-api/blob/feefe09f2781429e039c3f7ba88da7f29492e98c/service/waffo_pancake.go#L173) — 同一个来源还包含支付store校验，需单独分类而非整提交机械合并。

#### commit-d0cb7347c07b

- Title: fix: use max_completion_tokens for gpt-6-sol and gpt-6-luna (#7559)
- Feature/group: fix: use max_completion_tokens for gpt-6-sol and gpt-6-luna (#7559)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel_test_request_test.go`, `relaykit/dto/openai_request.go`
- Overlap with downstream changes: relaykit/dto/openai_request.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: GPT-6 Sol/Luna 参数能力
- AI recommendation: `migrate` — 本地没有上游GetOpenAIChatCapabilities，不能单独应用这次表扩展；若选择46的能力适配，应一起补Sol/Luna快照识别、max_completion_tokens和无思考时采样规则。
- AI risks: 模型规则需在实施时核对供应商固定文档；真实映射模型与公开别名、显式0/none不得丢失。
- AI dependencies: commit-49ec46966825
- AI suggested paths (not approval): relaykit/dto/openai_request.go, relay/channel/openai/adaptor.go, controller/channel_test_request_test.go, relaykit/dto/openai_request_zero_value_test.go
- AI evidence: [upstream:relaykit/dto/openai_request.go:332](https://github.com/QuantumNous/new-api/blob/d0cb7347c07bc935187c077b960435f3f3b08595/relaykit/dto/openai_request.go#L332) — 上游在模型能力表中加入Sol/Luna。
- AI evidence: [downstream:relaykit/dto/openai_request.go:278](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/dto/openai_request.go#L278) — 本地仍是旧GPT-5前缀规则。

#### commit-8112120673bf

- Title: fix(web): Add class to wrap the long strings at auto-disable status tooltip content in channel manage page (#7604)
- Feature/group: fix(web): Add class to wrap the long strings at auto-disable status tooltip content in channel manage page (#7604)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channels-columns-status-tooltip-content.test.tsx`, `web/src/features/channels/components/channels-columns.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道自动禁用原因长文本换行
- AI recommendation: `migrate` — 上游仅在状态原因容器加 wrap-anywhere；本地同一原因提示仍是普通 div，可迁入避免长错误字符串撑开。
- AI risks: 只改布局，保留完整可读错误信息、管理员可见性和复制行为。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/channels/components/channels-columns.tsx
- AI evidence: [upstream:web/src/features/channels/components/channels-columns.tsx:1026](https://github.com/QuantumNous/new-api/blob/8112120673bf593eb62e3a297dafca32a4605eb6/web/src/features/channels/components/channels-columns.tsx#L1026) — 固定源码：<div className='wrap-anywhere'>
- AI evidence: [downstream:web/src/features/channels/components/channels-columns.tsx:986](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/components/channels-columns.tsx#L986) — 下游对照：{statusReason && (

#### commit-dfd3cd893044

- Title: fix(web): hide system settings from non-root admins (#7513)
- Feature/group: fix(web): hide system settings from non-root admins (#7513)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/hooks/use-sidebar-data.ts`
- Overlap with downstream changes: web/src/hooks/use-sidebar-data.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 系统设置侧栏与 root 权限对齐
- AI recommendation: `migrate` — 本地 /option 后端已 RootAuth，但侧栏 System Settings 没有 requiredRole。迁入 SUPER_ADMIN 可见性限制，避免普通管理员进入必然失败的设置页。
- AI risks: 侧栏隐藏不能替代后端鉴权；保留本地其他管理入口的细分权限。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/hooks/use-sidebar-data.ts, web/src/hooks/__tests__/sidebar-config.test.tsx
- AI evidence: [upstream:web/src/hooks/use-sidebar-data.ts:177](https://github.com/QuantumNous/new-api/blob/dfd3cd8930447780a5733eaaba9a306df595251d/web/src/hooks/use-sidebar-data.ts#L177) — 固定源码：requiredRole: ROLE.SUPER_ADMIN,
- AI evidence: [downstream:router/api-router.go:284](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/api-router.go#L284) — 下游对照：optionRoute.Use(middleware.RootAuth())

#### commit-5fe8917be728

- Title: fix(i18n): keep cached zhTW/zhCN interface codes stable across page loads (#7139)
- Feature/group: fix(i18n): keep cached zhTW/zhCN interface codes stable across page loads (#7139)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/i18n/languages.test.ts`, `web/src/i18n/languages.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 缓存中文语言码刷新稳定性
- AI recommendation: `migrate` — 本地 convertDetectedLanguage 把非标准 zhTW 转小写后没有 zhtw 分支，会回退 zhCN；上游补该别名。建议优先小范围迁入并补检测/缓存回归。
- AI risks: 维持 zhCN/zhTW 内部码与 BCP47 转换区分，覆盖大小写、连字符、下划线和繁体中文浏览器。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/i18n/languages.ts, web/src/i18n/__tests__/languages.test.ts
- AI evidence: [upstream:web/src/i18n/languages.ts:67](https://github.com/QuantumNous/new-api/blob/5fe8917be728a93a29072a9fbd65429ed30d2930/web/src/i18n/languages.ts#L67) — 固定源码：lower === 'zhtw' ||
- AI evidence: [downstream:web/src/i18n/languages.ts:67](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/i18n/languages.ts#L67) — 下游对照：lower === 'zh-tw' ||

#### commit-56758edf95ec

- Title: fix(web): correct translations, chart ordering, and font loading (#7628)
- Feature/group: fix(web): correct translations, chart ordering, and font loading (#7628)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/dashboard/lib/__tests__/charts.test.ts`, `web/src/features/dashboard/lib/charts.ts`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/redemption-codes/components/redemptions-primary-buttons.tsx`, `web/src/features/security/__tests__/account-security.test.tsx`, `web/src/features/security/components/dialogs/delete-account-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-setup-dialog.tsx`, `web/src/features/subscriptions/components/subscriptions-columns.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/users/components/users-columns.tsx`, `web/src/features/wallet/__tests__/copy.test.tsx`, `web/src/features/wallet/components/dialogs/billing-history-dialog.tsx`, `web/src/features/wallet/components/recharge-form-card.tsx`, `web/src/features/wallet/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/styles/theme.css`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/styles/theme.css
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 翻译整句、图表排序与字体加载
- AI recommendation: `defer` — 22文件混合翻译插值/Trans、图表排序、钱包文案和字体。价值应按模块拆解；下游钱包/账务展示不同，不能整体替换文案与金额显示。
- AI risks: 需补齐七语言并保留金额精度、退款/修正含义，图表排序不能改变时间桶含义。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/pricing/components/model-details.tsx:1336](https://github.com/QuantumNous/new-api/blob/56758edf95ec162033a6b73b553ac789404f87c9/web/src/features/pricing/components/model-details.tsx#L1336) — 固定源码：: t('Prices shown per {{unit}} tokens', { unit: tokenUnitLabel })}
- AI evidence: [downstream:web/src/features/pricing/components/model-details.tsx:1017](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/pricing/components/model-details.tsx#L1017) — 下游对照：{t('Prices shown per')} {tokenUnitLabel} tokens

#### commit-1a4166d8e8ba

- Title: fix(relayconvert): keep Responses custom tools on Claude and Gemini upstreams (#7636)
- Feature/group: fix(relayconvert): keep Responses custom tools on Claude and Gemini upstreams (#7636)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/gemini/adaptor_responses_test.go`, `relaykit/relayconvert/convmeta/meta.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req_preprocess.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/policy_test.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`, `relaykit/relayconvert/response_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses custom tool 跨 Claude/Gemini
- AI recommendation: `defer` — 在216的状态还原机制上扩展Claude/Gemini并修正无名流式工具公告；本地前置状态机制未引入。建议同custom-tool功能组评估，不能仅凭现有ToolCallRequest.Custom判断covered。
- AI risks: 须保证两个协议到Responses的input完整复原、工具名分片不产生空工具，保留每请求状态隔离。
- AI dependencies: commit-335ebcff71b8
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/relayconvert/convmeta/meta.go:94](https://github.com/QuantumNous/new-api/blob/1a4166d8e8ba9802d2ca56fe8ecf0ed5404e80d5/relaykit/relayconvert/convmeta/meta.go#L94) — 本修复继续依赖跨请求/响应工具映射状态。
- AI evidence: [downstream:relaykit/relayconvert/convmeta/meta.go:18](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/convmeta/meta.go#L18) — 本地公共Meta还没有这套Responses工具状态接口。

#### commit-b48b74ab77cf

- Title: feat(jsplugin): decode hook results with moejs ToGoInto
- Feature/group: feat(jsplugin): decode hook results with moejs ToGoInto
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `THIRD-PARTY-LICENSES.md`, `go.mod`, `go.sum`, `pkg/jsplugin/engine.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: THIRD-PARTY-LICENSES.md, go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: moejs 类型化结果解码
- AI recommendation: `defer` — 使用 ToGoInto 直接解码插件 hook 输出并升级到 moejs alpha.5，减少中间转换。此项只能在引擎选型完成后衡量，不独立新增未使用依赖。
- AI risks: 需验证整数、RawJSON、null/缺失、深度和超大对象语义与旧引擎一致。
- AI dependencies: commit-2ffc59590115
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/engine.go:408](https://github.com/QuantumNous/new-api/blob/b48b74ab77cf289ccaaa8d99c067ce50cced4a82/pkg/jsplugin/engine.go#L408) — hook输出使用 moejs 类型化解码；代码：err = instance.runtime.ToGoInto(value, target)

#### commit-973cf8ef4600

- Title: fix(keys): prevent rounded clipping of group names (#7645)
- Feature/group: fix(keys): prevent rounded clipping of group names (#7645)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/api-key-group-cell.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: API key 分组名圆角裁切
- AI recommendation: `defer` — 只在前面新增 ApiKeyGroupCell 调用 GroupBadge 时设置 type=text；本地尚未引入该 cell，需随额度/分组展示重组确认。
- AI risks: 保留组名完整性、倍率和权限继承标识；不可借此扩大可见组范围。
- AI dependencies: commit-2bec370629aa, commit-ea7cb0ba4e0f
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/keys/components/api-key-group-cell.tsx:60](https://github.com/QuantumNous/new-api/blob/973cf8ef4600947a4270e95ada7916740fa8264c/web/src/features/keys/components/api-key-group-cell.tsx#L60) — 固定源码：type='text'
- AI evidence: [downstream:web/src/features/keys/components/api-keys-columns.tsx:194](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/keys/components/api-keys-columns.tsx#L194) — 下游对照：accessorKey: 'group',

#### commit-78bd5b1cbf7b

- Title: fix(relaykit): Gemini stream token details, Responses tool-output media on Gemini, Opus 5.5 default effort (#7687)
- Feature/group: fix(relaykit): Gemini stream token details, Responses tool-output media on Gemini, Opus 5.5 default effort (#7687)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/dto/usage_merge.go`, `relaykit/dto/usage_merge_test.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Gemini 流用量、工具媒体与 Opus 默认力度
- AI recommendation: `defer` — 同一提交混合Gemini累计模态usage覆盖、Responses→Gemini媒体能力和Claude Opus默认effort，本地缺usage_merge及统一reasoning模块。应拆成计费、媒体和推理三个选择项后再确认，不能整包标记小修。
- AI risks: 流累计快照不能重复相加；Gemini MIME/模型版本能力与Opus默认力度各需独立测试，不能改变下游表达式token归一化。
- AI dependencies: commit-0ed497f066a6, commit-3cea2bf799a6, commit-8a11b4305ed1, commit-1a4166d8e8ba
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/dto/usage_merge.go:272](https://github.com/QuantumNous/new-api/blob/78bd5b1cbf7bffc462a515a6c6e27567ace7a4d8/relaykit/dto/usage_merge.go#L272) — 上游把累计模态usage由相加改为按已报告模态覆盖。
- AI evidence: [downstream:relay/channel/gemini/relay-gemini.go:209](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/gemini/relay-gemini.go#L209) — 本地仍走现有Gemini流式usage路径，无对应合并模块。

#### commit-8ca26d81dfff

- Title: fix(web): stop frontend build files from exhausting the web rate limit
- Feature/group: fix(web): stop frontend build files from exhausting the web rate limit
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/constants.go`, `common/init.go`, `middleware/frontend_static.go`, `middleware/frontend_static_test.go`, `middleware/rate-limit.go`, `middleware/rate_limit_test.go`, `router/web-router.go`
- Overlap with downstream changes: common/constants.go, router/web-router.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 静态资源独立限流及gzip缓存
- AI recommendation: `migrate` — 建议适配静态构建文件独立限流和每文件gzip缓存，解决多个用户共享IP时静态chunk耗尽页面限流额度。本地下游全局中间件顺序不同且有#74资源观测，不能照抄上游NoRoute/pluginDispatcher结构。
- AI risks: 仅真实嵌入文件走静态预算，未知/API路径仍限流；规范缓存键避免URL别名无限增长，正确处理Accept-Encoding q=0、HEAD、Vary、内容类型、gzip字节观测。
- AI dependencies: none identified
- AI suggested paths (not approval): common/constants.go, common/init.go, middleware/frontend_static.go, middleware/frontend_static_test.go, middleware/rate-limit.go, middleware/rate_limit_test.go, router/web-router.go, router/web_router_test.go
- AI evidence: [upstream:middleware/rate-limit.go:165](https://github.com/QuantumNous/new-api/blob/8ca26d81dfff9d0454e0f5d06394c93689fe69b1/middleware/rate-limit.go#L165) — 静态文件由存在性判定进入独立限流；代码：func GlobalWebRateLimit(frontendFS static.ServeFileSystem) func(c *gin.Context) {
- AI evidence: [downstream:router/web-router.go:26](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/web-router.go#L26) — 本地所有页面和静态请求共用Web限流中间件；代码：router.Use(middleware.GlobalWebRateLimit())

#### commit-956019109bb0

- Title: feat(web): filter the user list by group
- Feature/group: feat(web): filter the user list by group
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/users/components/__tests__/group-filter.test.tsx`, `web/src/features/users/components/__tests__/quota-display.test.tsx`, `web/src/features/users/components/users-table.tsx`, `web/src/routes/_authenticated/users/index.tsx`
- Overlap with downstream changes: web/src/features/users/components/users-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 用户列表按分组筛选及 URL 恢复
- AI recommendation: `migrate` — 本地有字符串 groupFilter 参数但未提供完整可搜索单选组配置。迁入分组查询、筛选清空和 URL 恢复，按本地用户权限和分页规则适配。
- AI risks: 分组列表必须按当前管理员权限获取；切换筛选重置分页，旧 URL 字符串需兼容或明确转换。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/users/components/users-table.tsx, web/src/routes/_authenticated/users/index.tsx, web/src/features/users/components/__tests__/group-filter.test.tsx
- AI evidence: [upstream:web/src/routes/_authenticated/users/index.tsx:38](https://github.com/QuantumNous/new-api/blob/956019109bb09eb94417fd69ec0a5af0d897ac5a/web/src/routes/_authenticated/users/index.tsx#L38) — 固定源码：group: z.array(z.string()).optional().catch([]),
- AI evidence: [downstream:web/src/features/users/components/users-table.tsx:97](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/users/components/users-table.tsx#L97) — 下游对照：(columnFilters.find((filter) => filter.id === 'group')?.value as string) ??

#### commit-c6bac617d46a

- Title: fix(relaykit): make the thinking suffix adapter toggle cover effort suffixes
- Feature/group: fix(relaykit): make the thinking suffix adapter toggle cover effort suffixes
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/helper/reasoning_suffix_test.go`, `relaykit/relayconvert/convmeta/options.go`, `relaykit/relayconvert/reasoning/suffix.go`, `relaykit/relayconvert/reasoning/suffix_test.go`, `web/src/features/system-settings/models/claude-settings-card.tsx`, `web/src/features/system-settings/models/gemini-settings-card.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 思考后缀开关覆盖 effort 别名
- AI recommendation: `defer` — 上游让ThinkingAdapterEnabled同时控制-thinking和-high等effort尾缀；本地已有思考适配选项，但没有该统一ParseClaude/ParseGemini意图实现。应随模型修饰符组引入，并保持明确@参数不受开关影响。
- AI risks: 关闭适配时必须原样保留真实模型ID，尤其Kimi与厂商自有后缀；跨供应商默认值不可偷偷生效。
- AI dependencies: commit-0ed497f066a6, commit-7c044d7c5c2d, commit-6b659fd61c50, commit-8a11b4305ed1
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/relayconvert/reasoning/suffix.go:148](https://github.com/QuantumNous/new-api/blob/c6bac617d46ad2d63fa7cb692942fe6d21e264f7/relaykit/relayconvert/reasoning/suffix.go#L148) — 上游后缀解析统一遵守适配开关。
- AI evidence: [downstream:relaykit/relayconvert/convmeta/options.go:26](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/convmeta/options.go#L26) — 本地有相同配置概念但使用旧适配路径。

#### commit-0908d9e7b911

- Title: fix(epay): upgrade go-epay to v0.0.5
- Feature/group: fix(epay): upgrade go-epay to v0.0.5
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/payment_webhook_availability_test.go`, `go.mod`, `go.sum`
- Overlap with downstream changes: go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 易支付依赖修复
- AI recommendation: `upgrade` — 建议作为独立依赖升级候选，将go-epay从v0.0.4升到v0.0.5并运行支付回调可用性回归。本地仍锁定0.0.4；实施前需审阅依赖真实差异、订单签名和幂等回调，不随其他支付策略变化一起导入。
- AI risks: 支付回调验签、不同响应content-type及订单金额/状态必须回归；当前仅完成静态差异评估，未验证新依赖运行。
- AI dependencies: none identified
- AI suggested paths (not approval): controller/payment_webhook_availability_test.go, go.mod, go.sum
- AI evidence: [upstream:go.mod:7](https://github.com/QuantumNous/new-api/blob/0908d9e7b91124e5808f20d2d4b9402c7f2f5fae/go.mod#L7) — 上游仅更新易支付库并增加回调可用性测试；代码：github.com/Calcium-Ion/go-epay v0.0.5
- AI evidence: [downstream:go.mod:7](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/go.mod#L7) — 本地尚未升级该依赖；代码：github.com/Calcium-Ion/go-epay v0.0.4

#### commit-bb28f2ccad3f

- Title: feat(tokenkit): move token counting into a module calibrated on agent traffic
- Feature/group: feat(tokenkit): move token counting into a module calibrated on agent traffic
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.github/workflows/ci.yml`, `.github/workflows/electron-build.yml`, `.github/workflows/release.yml`, `AGENTS.md`, `Dockerfile`, `Dockerfile.dev`, `common/model.go`, `go.mod`, `go.sum`, `main.go`, `makefile`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_claude_test.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_gemini_usage_test.go`, `relay/channel/openai/relay-openai.go`, `relaykit/dto/claude.go`, `relaykit/dto/gemini.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/token_count_meta_test.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_resp.go`, `service/responses_usage.go`, `service/responses_usage_test.go`, `service/token_counter.go`, `service/token_estimator.go`, `service/tokenizer.go`, `service/usage_helpr.go`, `tokenkit/README.md`, `tokenkit/go.mod`, `tokenkit/go.sum`, `tokenkit/image.go`, `tokenkit/image_test.go`, `tokenkit/tokenizer.go`, `tokenkit/tokenkit.go`, `tokenkit/tokenkit_test.go`
- Overlap with downstream changes: .github/workflows/ci.yml, .github/workflows/electron-build.yml, .github/workflows/release.yml, AGENTS.md, Dockerfile, go.mod, go.sum, main.go, makefile, relay/channel/openai/relay-openai.go, relaykit/dto/openai_request.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 独立 tokenkit 与计数校准
- AI recommendation: `defer` — 新tokenkit模块迁出Token计数、改根模块依赖及多项CI/构建治理；下游计数还涉及Kimi动态工具、音频时长上界、财务预扣和显式安全界限。需专门模块迁移评估及真实流量校准后决定。
- AI risks: 原提交包含受保护AGENTS/工作流，不能自动整包引入；校准系数变化会改变预扣/估算结算，不得覆盖饱和与审计边界。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:tokenkit/go.mod:1](https://github.com/QuantumNous/new-api/blob/bb28f2ccad3f6286a9dc4a0ad9e20f571f57533b/tokenkit/go.mod#L1) — 上游增加独立Token计数模块。
- AI evidence: [downstream:service/token_counter.go:179](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/token_counter.go#L179) — 本地估算仍由根service模块承载。

#### commit-0a532f7f7d1b

- Title: feat(plugins): add an xAI Grok Imagine video plugin
- Feature/group: feat(plugins): add an xAI Grok Imagine video plugin
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/builtin_plugins_test.go`, `plugins/tasks/xai/plugin.js`, `plugins/xai_responses_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: xAI Grok Imagine 视频插件
- AI recommendation: `defer` — 增加 xAI 文生/图生、编辑及延长任务，并按输出秒数、图片输入、视频输入和被审核计费建模。是全新供应商任务能力，当前无驱动和对应定价，需先确认业务需求及账务策略。
- AI risks: 编辑和延长的分辨率后续由 c8d234baa1 修正，须按最终版本审查；不能根据未知输入媒体元数据裸转整数。
- AI dependencies: commit-eb48396d5fe9, commit-2ffc59590115
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/xai/plugin.js:6](https://github.com/QuantumNous/new-api/blob/0a532f7f7d1b507cc6657d6bc140f73d4298e46a/plugins/tasks/xai/plugin.js#L6) — 插件能力表新增 xAI 视频模型；代码："grok-imagine-video": { resolutions: ["480p", "720p"], videoInput: true },

#### commit-0a03088e0e73

- Title: fix(relayconvert): fix lossy conversions and converted-stream failure billing
- Feature/group: fix(relayconvert): fix lossy conversions and converted-stream failure billing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/responses_via_chat.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_response.go`, `relaykit/relayconvert/convmeta/meta.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_req.go`, `relaykit/relayconvert/internal/media/media.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/shared/claude/media.go`, `relaykit/relayconvert/internal/shared/gemini/schema.go`, `relaykit/relayconvert/internal/toolconv/decode.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/model.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/request_registry.go`
- Overlap with downstream changes: relaykit/dto/openai_request.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 转换丢失与失败流计费
- AI recommendation: `defer` — 22文件同时改工具schema/媒体/Responses失败事件和客户端断连收费；本地现有转换流将写失败视为服务端错误，且有不可变客户账务/成本快照。应与协议终态、usage累加和转换基础整组验证后迁移。
- AI risks: 明确上游response.failed、客户端取消、已输出部分内容、未收到usage时各自收费/退款语义，防止负扣费、漏费或重复凭证。
- AI dependencies: commit-0ed497f066a6, commit-8b2c7105f91f, commit-55a6cd2a44ea, commit-335ebcff71b8, commit-1a4166d8e8ba
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/openai/chat_via_responses.go:192](https://github.com/QuantumNous/new-api/blob/0a03088e0e737e99bc7ca936d3d2276e8d636fb8/relay/channel/openai/chat_via_responses.go#L192) — 上游开始分离上游失败与客户端写失败的usage语义。
- AI evidence: [downstream:relay/channel/openai/chat_via_responses.go:109](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/openai/chat_via_responses.go#L109) — 本地写失败仍按原服务端错误结算。

#### commit-acb347a197ff

- Title: feat(relay): web search encoding per channel and vendor-reported search billing (#7690)
- Feature/group: feat(relay): web search encoding per channel and vendor-reported search billing (#7690)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/alpha_search_handler.go`, `relay/channel/advancedcustom/adaptor.go`, `relay/channel/advancedcustom/adaptor_test.go`, `relay/channel/ali/adaptor.go`, `relay/channel/ali/adaptor_test.go`, `relay/channel/ali/web_search.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/claude/tool_billing_test.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_responses.go`, `relay/channel/moonshot/adaptor.go`, `relay/channel/moonshot/adaptor_test.go`, `relay/channel/openai/adaptor.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/relay_responses_billing_test.go`, `relay/channel/openai/web_search.go`, `relay/channel/xai/adaptor.go`, `relay/channel/xai/web_search.go`, `relay/channel/zhipu_4v/adaptor.go`, `relay/channel/zhipu_4v/web_search.go`, `relay/common/relay_info.go`, `relay/common/tool_usage.go`, `relay/common/tool_usage_test.go`, `relay/response_model_test.go`, `relay/responses_websocket.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/openai_request.go`, `relaykit/relayconvert/convmeta/options.go`, `relaykit/relayconvert/convmeta/websearch.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/model.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`, `relaykit/relayconvert/text_converter_registry.go`, `service/responses_usage.go`, `service/responses_usage_test.go`, `service/text_quota.go`, `service/text_quota_test.go`, `setting/operation_setting/tools.go`, `setting/operation_setting/tools_price_test.go`, `web/src/features/channels/lib/advanced-custom.ts`, `web/src/features/channels/types.ts`, `web/src/features/system-settings/models/tool-price-settings.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relay/channel/openai/adaptor.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, relaykit/dto/channel_settings.go, relaykit/dto/openai_request.go, service/text_quota.go, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 渠道搜索工具编码与实报收费
- AI recommendation: `defer` — 52文件引入每渠道web-search编码、工具计数API及供应商实报搜索费，连带Claude/Gemini→Responses路由；本地仍以现有BuiltInTools计数收费。应先明确按调用/按查询及客户售价/上游成本契约。
- AI risks: 供应商搜索费用单位、汇率、缓存及表达式单一事实源必须一致，不能把供应商成本直接当客户售价或双计工具费用。
- AI dependencies: commit-0ed497f066a6, commit-335ebcff71b8, commit-1a4166d8e8ba, commit-0a03088e0e73
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/alpha_search_handler.go:104](https://github.com/QuantumNous/new-api/blob/acb347a197ff0e22f651be9813b4a1a867342941/relay/alpha_search_handler.go#L104) — 上游改为统一可计费工具API。
- AI evidence: [downstream:relay/alpha_search_handler.go:112](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/alpha_search_handler.go#L112) — 本地仍手动记录一次web_search_preview。

#### commit-45094bdf41fb

- Title: fix(web): 修复 OriginOS 6 设备渠道页白屏
- Feature/group: fix(web): 修复 OriginOS 6 设备渠道页白屏
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/lib/motion.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Android OriginOS 零模糊白屏兼容
- AI recommendation: `migrate` — 本地页面动画仍以 blur(0px) 结束；上游使用0.01px后 transitionEnd 清除 filter，以规避特定Android渲染问题。建议迁入小修，但保留实机验证结论。
- AI risks: 尚未本机复现 OriginOS；验证结束态无filter、减少动画偏好、进入/退出与截图/浏览器兼容。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/lib/motion.ts
- AI evidence: [upstream:web/src/lib/motion.ts:46](https://github.com/QuantumNous/new-api/blob/45094bdf41fb13a5d24bd3e8eb3fdfec6241c59f/web/src/lib/motion.ts#L46) — 固定源码：filter: 'blur(0.01px)',
- AI evidence: [downstream:web/src/lib/motion.ts:41](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/lib/motion.ts#L41) — 下游对照：animate: { opacity: 1, y: 0, filter: 'blur(0px)' },

#### commit-04c597bf6cce

- Title: feat: support regex conditions in advanced parameter overrides (#7698)
- Feature/group: feat: support regex conditions in advanced parameter overrides (#7698)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/common/override.go`, `relay/common/override_test.go`, `web/src/features/channels/components/__tests__/param-override-editor.test.tsx`, `web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 高级参数覆盖正则条件
- AI recommendation: `migrate` — 本地条件比较支持full/prefix/contains/数值，上游增加Go regexp条件和每请求惰性缓存，并调整编辑器正则字面量。可以按本地已有审计覆盖机制选择性迁入后端与UI配套。
- AI risks: invalid regex、非字符串pattern、missing-key/invert、对象裁剪和并发请求隔离需回归；不得让正则缓存跨规则/请求串用。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/common/override.go, relay/common/override_test.go, web/src/features/channels/components/__tests__/param-override-editor.test.tsx, web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/zh.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json
- AI evidence: [upstream:relay/common/override.go:728](https://github.com/QuantumNous/new-api/blob/04c597bf6cce1eea76b9e67a4936e15fb54a5612/relay/common/override.go#L728) — 上游条件比较增加正则分支。
- AI evidence: [downstream:relay/common/override.go:645](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/common/override.go#L645) — 本地比较器尚无regex条件。

#### commit-d10f11ffff20

- Title: feat(gemini): bill Google Search grounding per query or per grounded prompt
- Feature/group: feat(gemini): bill Google Search grounding per query or per grounded prompt
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_gemini_usage_test.go`, `relay/channel/gemini/relay_responses.go`, `relay/common/tool_usage.go`, `relay/common/tool_usage_test.go`, `relaykit/dto/gemini.go`, `service/text_quota_test.go`, `setting/operation_setting/tools.go`, `setting/operation_setting/tools_price_test.go`, `web/src/features/system-settings/models/tool-price-settings.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Gemini Search 按查询/提示收费
- AI recommendation: `defer` — 上游按Gemini模型区分唯一搜索查询与grounded prompt，并承认供应商未回报精确账单数；本地仅用gemini_google_search_call布尔标记。须先审定计价表和证据归档，不能直接改变客户消费口径。
- AI risks: 跨帧去重、空查询、image/web查询及老模型按prompt收费都要覆盖；唯一query只是估计，需在客户日志和成本核算中保留来源。
- AI dependencies: commit-acb347a197ff
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/gemini/relay-gemini.go:100](https://github.com/QuantumNous/new-api/blob/d10f11ffff20688b8e3e745268cc0036e896d068/relay/channel/gemini/relay-gemini.go#L100) — 上游以唯一查询集合累积搜索证据。
- AI evidence: [downstream:relay/channel/gemini/relay-gemini.go:85](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/gemini/relay-gemini.go#L85) — 本地按是否grounded计一次标记。

#### commit-0cce21db0283

- Title: perf(jsplugin): stop rebuilding the usage-key replacer and request body on every hook call
- Feature/group: perf(jsplugin): stop rebuilding the usage-key replacer and request body on every hook call
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件 hook 请求体重复构造优化
- AI recommendation: `defer` — 为单次任务请求缓存已经标准化的 hook 请求体，减少序列化和键替换。没有现有插件开销可优化，且依赖引擎输入不可变假设，待插件运行指标和语义验证后再采用。
- AI risks: 共享可变 map 或重试覆盖参数时缓存可能串用旧数据；需按请求隔离。
- AI dependencies: commit-eb48396d5fe9, commit-2ffc59590115
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/task/jsplugin/adaptor.go:93](https://github.com/QuantumNous/new-api/blob/0cce21db0283304d139bf22588f4dc31bdc5ae67/relay/channel/task/jsplugin/adaptor.go#L93) — 缓存绑定到当前适配器请求对象；代码：hookRequestSource map[string]any

#### commit-bc823579294f

- Title: test(relaykit): keep format-conversion tests out of the repository
- Feature/group: test(relaykit): keep format-conversion tests out of the repository
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/reasoning/suffix_test.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 移除格式转换回归测试
- AI recommendation: `skip` — 该提交删除多项真实媒体/schema转换回归测试并将相关测试排除出仓库；这些恰是本次迁移需要的协议契约证据，不建议照搬。应在功能适配时保留或重写确定性边界测试。
- AI risks: 删除后会掩盖工具媒体、schema及转换诊断回归；AGENTS要求保留真实行为覆盖，不以降低测试范围获得绿CI。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go:473](https://github.com/QuantumNous/new-api/blob/bc823579294fbca0064b622c22477d3d04ca1bd8/relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go#L473) — 上游保留媒体基础测试但删除大量扩展格式回归。
- AI evidence: [downstream:relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go:91](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go#L91) — 本地仍有应保留的真实多模态回归。

#### commit-b23471112d8f

- Title: build: upgrade moejs to v0.1.0-alpha.6
- Feature/group: build: upgrade moejs to v0.1.0-alpha.6
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `go.mod`, `go.sum`
- Overlap with downstream changes: go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: moejs alpha.6 依赖更新
- AI recommendation: `defer` — 只提升未采用的 moejs 到 alpha.6。若未来选择任务插件，应以已审计的最终引擎版本统一锁定依赖，不为当前没有引用的模块单独升级。
- AI risks: 不能把 alpha 更新视为已通过下游跨平台、资源隔离与计费兼容性验证。
- AI dependencies: commit-2ffc59590115, commit-b48b74ab77cf
- AI suggested paths (not approval):
- AI evidence: [upstream:go.mod:169](https://github.com/QuantumNous/new-api/blob/b23471112d8f7e5848f70f30b798ee417d1ec75c/go.mod#L169) — 依赖版本由alpha.5升到alpha.6；代码：github.com/Calcium-Ion/moejs v0.1.0-alpha.6

#### commit-192802ee0691

- Title: build: build release binaries with moejs's PGO profile
- Feature/group: build: build release binaries with moejs's PGO profile
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.github/workflows/electron-build.yml`, `.github/workflows/release.yml`, `Dockerfile`, `electron/build.sh`
- Overlap with downstream changes: .github/workflows/electron-build.yml, .github/workflows/release.yml, Dockerfile
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: moejs PGO 构建配置
- AI recommendation: `skip` — 该项仅给上游发布、Electron 和 Docker 构建加入 moejs 的 PGO profile；本地未采用 moejs，且仓库已有独立 GitHub Actions 发布链。本次不导入上游工作流，未来引擎方案确定后按本仓库构建做独立调整。
- AI risks: 原样覆盖工作流会破坏现有发版和镜像约定；PGO模块路径在无依赖时不可解析。
- AI dependencies: commit-2ffc59590115, commit-b23471112d8f
- AI suggested paths (not approval):
- AI evidence: [upstream:Dockerfile:29](https://github.com/QuantumNous/new-api/blob/192802ee0691900575a3c7a0cdb90b2d7e703035/Dockerfile#L29) — 构建使用moejs模块内的default.pgo；代码：RUN go build -pgo="$(go list -m -f '{{.Dir}}' github.com/Calcium-Ion/moejs)/default.pgo" -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" -o new-api

#### commit-6370b2942416

- Title: feat(jsplugin): keep JSON member order for plugins that declare json-order@1
- Feature/group: feat(jsplugin): keep JSON member order for plugins that declare json-order@1
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `middleware/task_plugin.go`, `middleware/task_plugin_test.go`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/json_state.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/routing.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件声明保留 JSON 成员顺序
- AI recommendation: `defer` — 通过 json-order@1 能力保留 hook requestBody 原始成员顺序，适用于对键顺序敏感的供应商签名。它跨解码、存储和引擎接口，待插件架构确认后作为协议契约采用。
- AI risks: 不得以保序绕过宿主参数校验、凭据隔离和大小限制；签名原文和标准化对象必须区分。
- AI dependencies: commit-2ffc59590115, commit-b48b74ab77cf
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/engine.go:359](https://github.com/QuantumNous/new-api/blob/6370b29424168039e94d40d610191e7d2e65dbf4/pkg/jsplugin/engine.go#L359) — 引擎新增保留指定JSON成员原文的调用接口；代码：// CallPathWithMemberJSON is CallPathWithAdmissionTimeout that also returns the

#### commit-7aa3531ef4c2

- Title: fix(task-plugin): set multipart content type for upstream requests (#7655)
- Feature/group: fix(task-plugin): set multipart content type for upstream requests (#7655)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件 multipart boundary 对齐
- AI recommendation: `defer` — 宿主重建 multipart 后必须使用新 boundary 的 Content-Type，而非插件提供的旧头。属于插件文件上传正确性修复，现有 Go 上传路径没有该 submit 描述符，随插件图片协议引入。
- AI risks: 需验证多文件、重复字段、体积上限与请求头不能互相覆盖。
- AI dependencies: commit-eb48396d5fe9, commit-03563a4a7916
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/channel/task/jsplugin/adaptor.go:274](https://github.com/QuantumNous/new-api/blob/7aa3531ef4c247ad4891c06cc8ed9d0ffb73fecc/relay/channel/task/jsplugin/adaptor.go#L274) — 重建后的multipart使用宿主请求头；代码：if a.submit.BodyType == "multipart" {

#### commit-e0483122c7d3

- Title: fix(relayconvert): drop trailing assistant text from Responses input on Claude and Gemini
- Feature/group: fix(relayconvert): drop trailing assistant text from Responses input on Claude and Gemini
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_responses/req_helpers.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses 末尾 assistant prefill 兼容
- AI recommendation: `migrate` — 本地Responses→Claude/Gemini会保留末尾assistant纯文本，上游为重试后的已完成历史删除这类prefill并记录诊断。建议适配现有转换器，限定在用户/tool结果之后的尾部纯文本，不能删除带tool_calls的assistant。
- AI risks: 需确认旧模型有意prefill场景、全assistant/无user输入、混合媒体/思考块和工具回合；删历史须有明确诊断，不静默丢请求。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/relayconvert/internal/oai_responses/req_helpers.go, relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go, relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go, relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go, relaykit/relayconvert/internal/oai_responses/to_claude_messages_req_test.go, relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req_test.go
- AI evidence: [upstream:relaykit/relayconvert/internal/oai_responses/req_helpers.go:69](https://github.com/QuantumNous/new-api/blob/e0483122c7d3693629f0e84473380a8c9506c56a/relaykit/relayconvert/internal/oai_responses/req_helpers.go#L69) — 上游限定尾部assistant剥离并生成诊断。
- AI evidence: [downstream:relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go:23](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go#L23) — 本地原转换器尚无尾部prefill兼容处理。

#### commit-128089fa7ddf

- Title: feat(i18n): translate web console messages in the frontend and AI-client errors in the backend (#7712)
- Feature/group: feat(i18n): translate web console messages in the frontend and AI-client errors in the backend (#7712)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.env.example`, `AGENTS.md`, `common/audio.go`, `common/body_storage.go`, `common/constants.go`, `common/email.go`, `common/env.go`, `common/gin.go`, `common/gopool.go`, `common/init.go`, `common/limiter/limiter.go`, `common/log_text.go`, `common/log_text.zh-CN.json`, `common/message.go`, `common/pprof.go`, `common/redis.go`, `common/sys_log.go`, `common/topup-ratio.go`, `common/totp.go`, `common/utils.go`, `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/audit.go`, `controller/auth_session.go`, `controller/channel-billing.go`, `controller/channel-test.go`, `controller/channel.go`, `controller/channel_affinity_cache.go`, `controller/channel_test_internal_test.go`, `controller/channel_upstream_update.go`, `controller/channel_upstream_update_test.go`, `controller/checkin.go`, `controller/codex_usage.go`, `controller/custom_oauth.go`, `controller/group.go`, `controller/log.go`, `controller/login_verification.go`, `controller/midjourney.go`, `controller/misc.go`, `controller/model.go`, `controller/model_meta.go`, `controller/oauth.go`, `controller/option.go`, `controller/passkey.go`, `controller/passkey_test.go`, `controller/payment_compliance.go`, `controller/performance.go`, `controller/playground.go`, `controller/plugin_protocol.go`, `controller/plugin_protocol_image.go`, `controller/prefill_group.go`, `controller/pricing.go`, `controller/ratio_config.go`, `controller/ratio_sync.go`, `controller/redemption.go`, `controller/relay.go`, `controller/responses_websocket.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `controller/setup.go`, `controller/subscription.go`, `controller/subscription_payment_creem.go`, `controller/subscription_payment_epay.go`, `controller/subscription_payment_stripe.go`, `controller/subscription_payment_waffo_pancake.go`, `controller/system_task_handlers.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `controller/theme_compat_test.go`, `controller/token.go`, `controller/topup.go`, `controller/topup_creem.go`, `controller/topup_quota_limit_test.go`, `controller/topup_stripe.go`, `controller/topup_waffo.go`, `controller/topup_waffo_pancake.go`, `controller/twofa.go`, `controller/usedata.go`, `controller/user.go`, `controller/user_quota.go`, `controller/vendor_meta.go`, `controller/video_proxy.go`, `controller/wechat.go`, `docker-compose.yml`, `i18n/i18n.go`, `i18n/i18n_test.go`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `logger/logger.go`, `main.go`, `middleware/auth.go`, `middleware/auth_test.go`, `middleware/distributor.go`, `middleware/email-verification-rate-limit.go`, `middleware/i18n.go`, `middleware/model-rate-limit.go`, `middleware/rate-limit.go`, `middleware/rate_limit_test.go`, `middleware/recover.go`, `middleware/secure_verification.go`, `middleware/task_plugin.go`, `middleware/trusted_proxies.go`, `middleware/turnstile-check.go`, `middleware/utils.go`, `model/ability.go`, `model/audit_log.go`, `model/channel.go`, `model/channel_cache.go`, `model/checkin.go`, `model/frontend_option_migration.go`, `model/gorm_logger.go`, `model/log.go`, `model/log_other.go`, `model/main.go`, `model/model_pricing_config.go`, `model/option.go`, `model/option_primary_key_migration.go`, `model/passkey.go`, `model/pricing.go`, `model/quota_reserve.go`, `model/redemption.go`, `model/subscription.go`, `model/subscription_auth_test.go`, `model/subscription_reset_test.go`, `model/task_model_alias.go`, `model/token.go`, `model/topup.go`, `model/twofa.go`, `model/usedata.go`, `model/user.go`, `model/user_access_token.go`, `model/user_cache.go`, `model/user_quota_adjustment.go`, `model/user_session.go`, `model/utils.go`, `oauth/discord.go`, `oauth/generic.go`, `oauth/github.go`, `oauth/linuxdo.go`, `oauth/oidc.go`, `oauth/registry.go`, `oauth/types.go`, `pkg/perf_metrics/flush.go`, `pkg/wsmanager/wsmanager.go`, `relay/audio_handler.go`, `relay/channel/api_request.go`, `relay/channel/aws/relay-aws.go`, `relay/channel/baidu/relay-baidu.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/cloudflare/relay_cloudflare.go`, `relay/channel/cohere/relay-cohere.go`, `relay/channel/coze/relay-coze.go`, `relay/channel/dify/relay-dify.go`, `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/jimeng/sign.go`, `relay/channel/ollama/relay-ollama.go`, `relay/channel/ollama/stream.go`, `relay/channel/openai/audio.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/helper.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_image.go`, `relay/channel/openai/relay_realtime.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/responses_via_chat.go`, `relay/channel/palm/relay-palm.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/tencent/relay-tencent.go`, `relay/channel/xai/text.go`, `relay/channel/xunfei/relay-xunfei.go`, `relay/channel/zhipu/relay-zhipu.go`, `relay/channel/zhipu_4v/image.go`, `relay/common/conversion_diagnostics.go`, `relay/common/relay_info.go`, `relay/common/tool_usage.go`, `relay/compatible_handler.go`, `relay/gemini_handler.go`, `relay/helper/common.go`, `relay/helper/price.go`, `relay/helper/stream_scanner.go`, `relay/helper/valid_request.go`, `relay/image_handler.go`, `relay/mjproxy_handler.go`, `relay/relay_task.go`, `relay/relay_task_test.go`, `relay/request_billing.go`, `relay/responses_handler.go`, `relay/websocket.go`, `router/main.go`, `router/plugin-router.go`, `service/auth_cleanup.go`, `service/authz/adapter.go`, `service/authz/enforcer.go`, `service/billing.go`, `service/billing_session.go`, `service/channel.go`, `service/channel_affinity.go`, `service/channel_select.go`, `service/codex_credential_refresh_task.go`, `service/download.go`, `service/error.go`, `service/file_decoder.go`, `service/file_service.go`, `service/group.go`, `service/http.go`, `service/http_client.go`, `service/http_transport_policy.go`, `service/http_transport_sharded.go`, `service/image.go`, `service/log_info_generate.go`, `service/midjourney.go`, `service/passkey/service.go`, `service/passkey/session.go`, `service/quota.go`, `service/relay_error.go`, `service/subscription_reset_task.go`, `service/system_instance.go`, `service/system_task.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/task_polling.go`, `service/task_polling_test.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/tiered_settle.go`, `service/tiered_settle_test.go`, `service/user_notify.go`, `service/violation_fee.go`, `service/webhook.go`, `setting/chat.go`, `setting/config/config.go`, `setting/console_setting/validation.go`, `setting/model_setting/global.go`, `setting/operation_setting/tools.go`, `setting/rate_limit.go`, `setting/ratio_setting/group_ratio.go`, `setting/ratio_setting/model_ratio.go`, `setting/system_setting/task_artifact.go`, `setting/system_setting/task_artifact_store.go`, `setting/user_usable_group.go`, `web/src/components/model-group-selector.tsx`, `web/src/components/model-group-selector/__tests__/group-detail.test.ts`, `web/src/components/model-group-selector/group-detail.ts`, `web/src/features/auth/api.ts`, `web/src/features/auth/passkey/hooks/use-passkey-management.ts`, `web/src/features/channels/components/channels-columns.tsx`, `web/src/features/channels/components/dialogs/codex-usage-dialog.tsx`, `web/src/features/channels/components/dialogs/multi-key-manage-dialog.tsx`, `web/src/features/channels/components/dialogs/ollama-models-dialog.tsx`, `web/src/features/channels/lib/channel-actions.ts`, `web/src/features/keys/components/api-keys-mutate-drawer.tsx`, `web/src/features/models/components/dialogs/model-delete-dialog.tsx`, `web/src/features/models/components/models-table.tsx`, `web/src/features/playground/api.ts`, `web/src/features/playground/types.ts`, `web/src/features/subscriptions/components/dialogs/user-subscriptions-dialog.tsx`, `web/src/features/subscriptions/components/subscriptions-mutate-drawer.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/integrations/payment-settings-section.tsx`, `web/src/features/system-settings/integrations/waffo-pancake-settings-section.tsx`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/column-helpers.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/columns/task-logs-columns.tsx`, `web/src/features/usage-logs/components/common-log-mobile-card.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/components/dialogs/task-details-dialog.tsx`, `web/src/features/usage-logs/lib/__tests__/token-name.test.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/types.ts`, `web/src/features/wallet/hooks/use-affiliate.ts`, `web/src/features/wallet/hooks/use-waffo-pancake-payment.ts`, `web/src/features/wallet/hooks/use-waffo-payment.ts`, `web/src/features/wallet/lib/payment.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/lib/secure-verification.ts`, `web/src/lib/server-error-message.test.ts`, `web/src/lib/server-error-message.ts`
- Overlap with downstream changes: .env.example, AGENTS.md, common/constants.go, common/gin.go, common/redis.go, controller/audit.go, controller/channel-test.go, controller/channel.go, controller/log.go, controller/misc.go, controller/model.go, controller/option.go, controller/playground.go, controller/relay.go, controller/secure_verification.go, controller/system_task_handlers.go, controller/user.go, controller/video_proxy.go, docker-compose.yml, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, logger/logger.go, main.go, middleware/distributor.go, model/ability.go, model/channel.go, model/channel_cache.go, model/gorm_logger.go, model/log.go, model/main.go, model/option.go, model/twofa.go, model/usedata.go, model/user.go, model/user_cache.go, relay/channel/openai/helper.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, relay/compatible_handler.go, relay/helper/common.go, relay/mjproxy_handler.go, relay/relay_task.go, relay/responses_handler.go, service/auth_cleanup.go, service/billing.go, service/billing_session.go, service/download.go, service/http.go, service/http_client.go, service/midjourney.go, service/quota.go, service/system_task.go, service/task_billing.go, service/task_billing_test.go, service/task_polling.go, service/task_polling_test.go, service/text_quota.go, service/violation_fee.go, service/webhook.go, setting/console_setting/validation.go, setting/ratio_setting/model_ratio.go, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/columns/task-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/lib/server-error-message.test.ts, web/src/lib/server-error-message.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 控制台消息前端翻译与客户端后端翻译分离
- AI recommendation: `defer` — 288文件改message_key/message_params、存储日志content_parts、后端语言来源及邮件通知。下游还有短信、账务/纠错、资产、任务日志和自定义错误码，需完整双端兼容设计与历史日志读取方案。
- AI risks: 不能丢失原始错误、用户语言或把管理台key泄漏到AI客户端；会影响调用方契约与持久化日志，需分阶段兼容。
- AI dependencies: commit-12be9975c0bf
- AI suggested paths (not approval):
- AI evidence: [upstream:common/message.go:45](https://github.com/QuantumNous/new-api/blob/128089fa7ddf41b575206a0feb28940d33793833/common/message.go#L45) — 固定源码：fields := gin.H{"message_key": m.Key}
- AI evidence: [downstream:i18n/i18n.go:90](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/i18n/i18n.go#L90) — 下游对照：func T(c *gin.Context, key string, args ...map[string]any) string {

#### commit-1d4328e97417

- Title: docs(relaykit): rewrite README for the current conversion API
- Feature/group: docs(relaykit): rewrite README for the current conversion API
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/README.md`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: RelayKit 转换API文档重写
- AI recommendation: `defer` — 该README描述了尚未迁入的工具状态、统一reasoning/Codec/usage和新API，直接采用会使文档先于代码失真；且README/署名行属受保护范围。待选定架构完成后另行人工维护对应章节与归属。
- AI risks: 不能删除或改写现有new-api/QuantumNous来源；文档中的公共API与独立构建要求必须以实际下游版本为准。
- AI dependencies: commit-0ed497f066a6, commit-7bbe85bcb095, commit-335ebcff71b8, commit-acb347a197ff
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/README.md:28](https://github.com/QuantumNous/new-api/blob/1d4328e97417a043a161a0dd30a5b129be3ace49/relaykit/README.md#L28) — 新文档声明本地尚未具备的工具状态还原接口。
- AI evidence: [downstream:relaykit/README.md:3](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/README.md#L3) — 本地仍保留原始独立模块来源与旧API说明。

#### commit-b66674ac23d7

- Title: feat(task-plugin): let Seedance and wan3.0 pick the video duration with -1
- Feature/group: feat(task-plugin): let Seedance and wan3.0 pick the video duration with -1
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.agents/rules/billing.md`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `pkg/jsplugin/json_state.go`, `pkg/jsplugin/registry.go`, `plugins/alibaba_wan3_test.go`, `plugins/doubao_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Seedance/Wan 自动时长 -1
- AI recommendation: `defer` — 通过 duration-auto@1 显式允许 -1 哨兵，由驱动按型号最大长度预留并按真实用量结算。当前通用任务验证拒绝负时长，不能直接放开为负数；需明确对哪些模型启用及最坏预扣额度。
- AI risks: 若把-1当计费秒数会产生负扣费；对映射别名、缺失时长和输入视频总时长必须一致。
- AI dependencies: commit-eb48396d5fe9, commit-3b4652269a6a, commit-65d3a2171fd1
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/registry.go:139](https://github.com/QuantumNous/new-api/blob/b66674ac23d79bfd35a90797ca1ac64fc74a2d8f/pkg/jsplugin/registry.go#L139) — 自动时长必须由插件显式声明能力；代码：func (m Meta) AcceptsAutoDuration() bool {

#### commit-29509776ef85

- Title: feat(jsplugin): let driver hooks read upstream data with utils.fetch
- Feature/group: feat(jsplugin): let driver hooks read upstream data with utils.fetch
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `common/log_text.zh-CN.json`, `controller/task.go`, `controller/video_proxy.go`, `docs/plugin-api/README.md`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/fetch.go`, `pkg/jsplugin/request.go`, `pkg/jsplugin/utils.go`, `plugins/alibaba_responses_test.go`, `plugins/hailuo_responses_test.go`, `plugins/xai_responses_test.go`, `relay/channel/adapter.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/auth_test.go`
- Overlap with downstream changes: controller/task.go, controller/video_proxy.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件受控网络读取
- AI recommendation: `defer` — 新增 utils.fetch 让 hook 从声明域名读取上游数据，并重构媒体内容请求解析。运行时网络权限扩大，需先评审 SSRF、重定向、授权头、超时/字节上限和代理配置；不宜夹入普通同步批次。
- AI risks: 保留租户凭据与私有对象访问边界，不能把API key转发到未声明主机；媒体下载还需防内网与大文件。
- AI dependencies: commit-2ffc59590115, commit-6370b2942416
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/fetch.go:18](https://github.com/QuantumNous/new-api/blob/29509776ef855965fb2cff3978578d891005945e/pkg/jsplugin/fetch.go#L18) — 新增独立受控网络fetch实现；代码：type Fetcher func(ctx context.Context, request FetchRequest) (FetchResponse, error)

#### commit-66c3e142820e

- Title: fix(hailuo): download MiniMax v1 videos through the file retrieve link
- Feature/group: fix(hailuo): download MiniMax v1 videos through the file retrieve link
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/task_generic_test.go`, `plugins/hailuo_responses_test.go`, `plugins/tasks/hailuo/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: MiniMax v1 下载链接解析
- AI recommendation: `defer` — 插件在内容请求时调用文件 retrieve 取得实际下载地址并返回无凭据 GET，避免把 JSON 端点当视频。下游 Go 已在成功轮询时解析 DownloadURL，时机不同且未证明签名链接有效期等价，不能简单 covered；如有过期问题应单独改现有 Go 内容链。
- AI risks: 下载时重新解析与轮询时持久URL的缓存语义不同；不得把Bearer发往资源域。
- AI dependencies: commit-eb48396d5fe9, commit-29509776ef85
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/hailuo/plugin.js:611](https://github.com/QuantumNous/new-api/blob/66c3e142820e513a328def88f2771f73479275bc/plugins/tasks/hailuo/plugin.js#L611) — 下载前校验retrieve响应并读取真实链接；代码：throw new Error("MiniMax file retrieve returned an unexpected response: HTTP " + response.status);
- AI evidence: [downstream:relay/channel/task/hailuo/adaptor.go:284](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/task/hailuo/adaptor.go#L284) — 本地Go会读取文件retrieve结果，但发生在轮询阶段；代码：return retrieveResp.File.DownloadURL

#### commit-3f02244783e0

- Title: feat(web): show the request domains a plugin declares when uploading it
- Feature/group: feat(web): show the request domains a plugin declares when uploading it
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/task-plugins/components/marketplace-capabilities.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/features/task-plugins/lib/plugin-meta-preview.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件上传展示请求域名
- AI recommendation: `defer` — 手工上传也展示源码声明的请求域名及能力，帮助管理员评审网络权限。依赖允许插件网络的安全策略与静态预览，建议未来引擎上线时必须一起呈现，当前不新增孤立上传页。
- AI risks: 声明展示不是网络强制限制的替代；必须拒绝未声明网络访问。
- AI dependencies: commit-eb48396d5fe9, commit-b6809a52df88, commit-29509776ef85
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/task-plugins/components/upload-dialog.tsx:324](https://github.com/QuantumNous/new-api/blob/3f02244783e0bb96b1398601f156625936be2f15/web/src/features/task-plugins/components/upload-dialog.tsx#L324) — 上传表单显示源码声明的网络能力；代码：<MarketplaceCapabilities source={deferredSource} />

#### commit-ba3b2c015ad8

- Title: refactor(jsplugin): stop requiring json-clone@1 and submit-sse-delta@1
- Feature/group: refactor(jsplugin): stop requiring json-clone@1 and submit-sse-delta@1
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `plugins/tasks/alibaba/plugin.js`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/submit_stream.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 插件能力标记兼容简化
- AI recommendation: `defer` — 不再强制声明 json-clone 和 submit-sse-delta，宿主按 hook 可用性识别增量提交，并保留旧能力标记兼容。应作为最终插件协议整体版本选择，而非独立清理当前不存在的标记。
- AI risks: 需确保旧插件仍工作，自动识别不会错误重复输出SSE或改变克隆隔离。
- AI dependencies: commit-eb48396d5fe9, commit-74629e29f83f, commit-2ffc59590115
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/jsplugin/registry.go:233](https://github.com/QuantumNous/new-api/blob/ba3b2c015ad83e6710c18592b0e3e3ba9df9a67e/pkg/jsplugin/registry.go#L233) — 按能力或hook推断提交增量模式；代码：func UsesSubmitEventDelta(meta Meta, engine *Engine) bool {

#### commit-c8d234baa18b

- Title: fix(xai): bill video edits and extensions by their output resolution
- Feature/group: fix(xai): bill video edits and extensions by their output resolution
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `plugins/tasks/xai/plugin.js`, `plugins/xai_responses_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: xAI 编辑/延长输出分辨率结算
- AI recommendation: `defer` — 改为按实际输出 MP4 头解析分辨率结算，预留使用最大支持档，修复编辑和延长任务依输入假设计价的问题。依赖未采用的 xAI 插件和受控 fetch，未来接入须与首版一起评审。
- AI risks: 媒体头是上游/用户可控数据，时长尺寸需有界；无法读取时不能错误降档少扣或重复扣费。
- AI dependencies: commit-0a532f7f7d1b, commit-29509776ef85
- AI suggested paths (not approval):
- AI evidence: [upstream:plugins/tasks/xai/plugin.js:519](https://github.com/QuantumNous/new-api/blob/c8d234baa18b9090b394fc1b8f384eec94f8bfef/plugins/tasks/xai/plugin.js#L519) — 按输出媒体头确定最终计价档位；代码：const resolution = outputResolution(videoURL(data));

#### commit-c6741c36a9a1

- Title: feat(channel): add a per-channel switch to skip upstream TLS verification
- Feature/group: feat(channel): add a per-channel switch to skip upstream TLS verification
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/channel-billing.go`, `controller/channel_upstream_update.go`, `controller/codex_usage.go`, `controller/video_proxy.go`, `relay/channel/api_request.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/vertex/service_account.go`, `relay/mjproxy_handler.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/channel_settings_test.go`, `service/codex_channel_models.go`, `service/http_client.go`, `service/http_client_transport_test.go`, `service/http_transport_policy.go`, `service/task_polling.go`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/channel-form-errors.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/video_proxy.go, relay/mjproxy_handler.go, relaykit/dto/channel_settings.go, relaykit/dto/channel_settings_test.go, service/http_client.go, service/http_client_transport_test.go, service/task_polling.go, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/channel-form-errors.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 每渠道跳过上游 TLS 校验
- AI recommendation: `defer` — 此功能扩展到余额/模型查询、媒体代理、relay和前端设置等29文件，本地另有诊断、资产和火山原生请求客户端。尚无明确业务需要与所有请求路径一致性证明，建议暂缓开启能力。
- AI risks: 需要验证默认严格、仅该渠道生效、按proxy+TLS正确缓存、连接池不串用、媒体SSRF不弱化；全局TLS开关与每渠道设置不能相互污染。
- AI dependencies: commit-feefe09f2781
- AI suggested paths (not approval):
- AI evidence: [upstream:controller/channel-billing.go:161](https://github.com/QuantumNous/new-api/blob/c6741c36a9a1a729554ad94ee99f21602a214a23/controller/channel-billing.go#L161) — 上游让管理查询也读取每渠道TLS设置。
- AI evidence: [downstream:controller/channel-billing.go:160](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/controller/channel-billing.go#L160) — 本地客户端仅带代理设置。

#### commit-1ca79d57e77d

- Title: fix(pricing): show recent status samples without gaps
- Feature/group: fix(pricing): show recent status samples without gaps
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/perf_metrics.go`, `model/perf_metric.go`, `pkg/perf_metrics/metrics.go`, `pkg/perf_metrics/metrics_test.go`, `pkg/perf_metrics/types.go`, `web/src/features/performance-metrics/api.ts`, `web/src/features/performance-metrics/types.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 模型近期健康采样无空洞
- AI recommendation: `defer` — 将近期样本查询独立于汇总窗口，按目标模型获取最近非空小时数据并合并内存数据，修改API和卡片。应作为61的最终方案整体评估，保留租户组过滤和数据库三方言；当前紧凑三柱设计是否改动需确认。
- AI risks: recent_model数量和查询成本要有界；不能把无请求时段误标成功，也不应新增下游节点监控持久化。
- AI dependencies: commit-5c7cca015525
- AI suggested paths (not approval):
- AI evidence: [upstream:model/perf_metric.go:120](https://github.com/QuantumNous/new-api/blob/1ca79d57e77d91c855371b43bb26b151429e18a1/model/perf_metric.go#L120) — 新增独立近期样本数据库查询；代码：func GetRecentPerfMetricSuccessBuckets(modelName string, startTs int64, endTs int64, groups []string, limit int) ([]PerfMetricSummaryBucket, error) {
- AI evidence: [downstream:pkg/perf_metrics/metrics.go:190](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/pkg/perf_metrics/metrics.go#L190) — 本地仍返回最近三条非空率，API契约不同；代码：RecentSuccessRates: recentSuccessRates(modelBuckets[name], 3),

#### commit-0f38a14a5e8e

- Title: fix(web): restore browser sans-serif for Auto font
- Feature/group: fix(web): restore browser sans-serif for Auto font
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/config-drawer.tsx`, `web/src/context/__tests__/theme-preferences.test.tsx`, `web/src/context/theme-customization-provider.tsx`, `web/src/lib/theme-customization.ts`, `web/src/styles/index.css`, `web/src/styles/theme-presets.css`, `web/src/styles/theme.css`
- Overlap with downstream changes: web/src/styles/index.css, web/src/styles/theme.css
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Auto 字体恢复浏览器无衬线
- AI recommendation: `defer` — 上游给 resolveThemeFont 增加 system 结果并重设 Auto；本地默认字体/主题有明确产品设计。先确认字体策略，不能把上游偏好当同步必需项。
- AI risks: 加载字体、显式Sans/Serif、预设默认和持久化旧值需一致，避免首次闪烁和CJK回退异常。
- AI dependencies: commit-54eee488bed3
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/lib/theme-customization.ts:177](https://github.com/QuantumNous/new-api/blob/0f38a14a5e8e322ea741478b0e49878be3130e1d/web/src/lib/theme-customization.ts#L177) — 固定源码：return PRESET_DEFAULT_FONT[preset] ?? 'system'
- AI evidence: [downstream:web/src/lib/theme-customization.ts:194](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/lib/theme-customization.ts#L194) — 下游对照：return PRESET_DEFAULT_FONT[preset] ?? 'sans'

#### commit-036ba376da7d

- Title: fix(relaykit): match Claude effort and Gemini 2.5 budgets to each model
- Feature/group: fix(relaykit): match Claude effort and Gemini 2.5 budgets to each model
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/gemini.go`, `relaykit/relayconvert/reasoning/suffix.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude effort 与 Gemini 2.5 budget 能力
- AI recommendation: `defer` — 上游按细分模型表调整effort/禁用思考行为及Gemini预算，依赖尚未迁入的shared/claude/reasoning实现。本地目前的原生请求和旧适配语义不同，必须随推理重构验证真实模型矩阵。
- AI risks: 已知/未知模型、显式关闭、xhigh/max、budget/max_tokens和采样清理都会受影响；不可用默认值覆盖客户端明确意图。
- AI dependencies: commit-0ed497f066a6, commit-7c044d7c5c2d, commit-6b659fd61c50, commit-8a11b4305ed1, commit-c6bac617d46a
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/relayconvert/internal/shared/claude/reasoning.go:133](https://github.com/QuantumNous/new-api/blob/036ba376da7d572a4db3e0be8553aadc28c3616a/relaykit/relayconvert/internal/shared/claude/reasoning.go#L133) — 上游会根据模型重写或删除客户端effort。
- AI evidence: [downstream:relay/channel/claude/adaptor.go:29](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/claude/adaptor.go#L29) — 本地原生Claude请求目前直接透传。

#### commit-13c8e90b8f4b

- Title: fix(relayconvert): send Responses image URLs to Chat upstreams as objects
- Feature/group: fix(relayconvert): send Responses image URLs to Chat upstreams as objects
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/testdata/golden/request/openai_responses_to_openai.golden.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses image_url 转 Chat 对象
- AI recommendation: `merge` — 本地responsesImagePartToChatImageURL在image_url为字符串时原样返回，Chat上游要求{url:...}；上游只包对象并保留已是对象的原值。可优先采用这个局部协议修复并补字符串/对象回归。
- AI risks: 检查URL/DataURI、原始对象detail、空值及工具媒体提升路径；字符串形式不凭空增加新detail值。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go, relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go
- AI evidence: [upstream:relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go:491](https://github.com/QuantumNous/new-api/blob/13c8e90b8f4b929aa88c2e9616efc462e0068d72/relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go#L491) — 上游把字符串URL转为Chat所需对象。
- AI evidence: [downstream:relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go:436](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go#L436) — 本地对应函数会直接返回字符串。

#### commit-5ac75628409d

- Title: fix(relayconvert): keep answer text that shares a stream chunk with reasoning
- Feature/group: fix(relayconvert): keep answer text that shares a stream chunk with reasoning
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 同一流片段中的思考和答案
- AI recommendation: `migrate` — 本地Chat→Claude流将reasoning与text放在互斥分支，二者同帧时答案丢失；上游先结束thinking再开启text并分别发送delta。建议迁移该状态转换，适配本地额外工具索引/usage处理。
- AI risks: 测试同帧/交错帧、tool_calls同帧、块索引、content_block_stop顺序、最终usage与message_stop恰好一次。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go, relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go
- AI evidence: [upstream:relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go:366](https://github.com/QuantumNous/new-api/blob/5ac75628409de7e3adde7d36771700714af50853/relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go#L366) — 上游答案文本由独立if发送，不再与reasoning互斥。
- AI evidence: [downstream:relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go:313](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go#L313) — 本地仍构造单个互斥thinking/text delta。

#### commit-977450a3e810

- Title: fix(rate-limit): explain model rate-limit rejections without Redis
- Feature/group: fix(rate-limit): explain model rate-limit rejections without Redis
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/responses_websocket.go`, `controller/responses_websocket_test.go`, `middleware/model-rate-limit.go`, `middleware/model_rate_limit_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 内存限流 429 错误解释
- AI recommendation: `migrate` — 本地内存限流总数和成功数超限只回空429，Redis已有说明；上游返回标准OpenAI错误。可仅迁入两个本地内存分支的解释和回归，不要求为这项小修先引入WS或reservation。
- AI risks: 保持429、Abort和原限流计数语义；返回当前策略的分钟/上限，支持现有语言，不把UI错误改善误认为并发预占位已修复。
- AI dependencies: none identified
- AI suggested paths (not approval): middleware/model-rate-limit.go, middleware/model_rate_limit_test.go
- AI evidence: [upstream:middleware/model-rate-limit.go:121](https://github.com/QuantumNous/new-api/blob/977450a3e810ee9c7d7e98acce1dc4efad90b491/middleware/model-rate-limit.go#L121) — 上游内存拒绝返回本地化结构化错误。
- AI evidence: [downstream:middleware/model-rate-limit.go:144](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/middleware/model-rate-limit.go#L144) — 本地内存限流仍返回无错误正文的429。

#### commit-b892a6b2e44d

- Title: fix(web): reload once when a page's code fails to load
- Feature/group: fix(web): reload once when a page's code fails to load
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/errors/__tests__/general-error.test.tsx`, `web/src/features/errors/general-error.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 代码分块加载失败一次性恢复
- AI recommendation: `migrate` — 上游 GeneralError 只在展示 ChunkLoadError 时刷新一次并保留session标记；本地只显示通用错误，部署后缓存旧chunk可能无法自动恢复。建议按当前Rsbuild错误特征迁入。
- AI risks: 禁止无限刷新与intent预加载触发刷新；存储不可用、CSS chunk失败和持续网络故障需测试。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/errors/general-error.tsx, web/src/features/errors/__tests__/chunk-recovery.test.tsx
- AI evidence: [upstream:web/src/features/errors/general-error.tsx:52](https://github.com/QuantumNous/new-api/blob/b892a6b2e44df572698f6cf731ff381f1977a1f2/web/src/features/errors/general-error.tsx#L52) — 固定源码：(error.name !== 'ChunkLoadError' &&
- AI evidence: [downstream:web/src/features/errors/general-error.tsx:48](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/errors/general-error.tsx#L48) — 下游对照：const status = getHttpStatus(error)

#### commit-1fb6d07fee01

- Title: perf(web): stop loading the whole Simple Icons pack on four pages
- Feature/group: perf(web): stop loading the whole Simple Icons pack on four pages
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/assets/brand-icons/icon-alipay.tsx`, `web/src/assets/brand-icons/index.ts`, `web/src/features/auth/components/oauth-callback-screen.tsx`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`, `web/src/features/wallet/lib/__tests__/payment-icons.test.tsx`, `web/src/features/wallet/lib/ui.tsx`
- Overlap with downstream changes: web/src/features/users/components/dialogs/user-binding-dialog.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 四个页面移除 Simple Icons 整包导入
- AI recommendation: `migrate` — 上游将 react-icons/si 的 GitHub/微信/Linux/Alipay 换为独立品牌组件。本地仍导入整包，可迁入显示层资源替换，保留下游微信/手机绑定入口。
- AI risks: 确保图标可访问名称、尺寸、归属与路径完整；不改变任何身份绑定或支付行为。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/assets/brand-icons/icon-alipay.tsx, web/src/assets/brand-icons/index.ts, web/src/features/auth/components/oauth-callback-screen.tsx, web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/features/wallet/lib/ui.tsx
- AI evidence: [upstream:web/src/features/users/components/dialogs/user-binding-dialog.tsx:35](https://github.com/QuantumNous/new-api/blob/1fb6d07fee0171d5f4a0bd2d2ea2660572b08389/web/src/features/users/components/dialogs/user-binding-dialog.tsx#L35) — 固定源码：import { IconGithub, IconDiscord } from '@/assets/brand-icons'
- AI evidence: [downstream:web/src/features/users/components/dialogs/user-binding-dialog.tsx:32](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/users/components/dialogs/user-binding-dialog.tsx#L32) — 下游对照：import { SiGithub, SiDiscord } from 'react-icons/si'

#### commit-06dd76a9d38e

- Title: fix(playground): keep each account's Playground data separate
- Feature/group: fix(playground): keep each account's Playground data separate
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/playground/lib/storage/__tests__/storage.test.ts`, `web/src/features/playground/lib/storage/storage.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Playground 会话数据按用户隔离
- AI recommendation: `migrate` — 本地 config/messages/parameterEnabled 使用全局 localStorage key；上游按用户ID分键，匿名不写入并清理无主旧数据。建议优先迁入，避免同一浏览器切号读取前一账号聊天。
- AI risks: 明确旧共享数据清除策略；验证 A→B→A、退出登录、无ID、超大消息与当前用户清除，不可误删其他用户数据。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/playground/lib/storage/storage.ts, web/src/features/playground/lib/storage/__tests__/storage.test.ts
- AI evidence: [upstream:web/src/features/playground/lib/storage/storage.ts:51](https://github.com/QuantumNous/new-api/blob/06dd76a9d38eb0eaf8e900b1264b500250156c4d/web/src/features/playground/lib/storage/storage.ts#L51) — 固定源码：function getUserStorageKey(key: string): string | null {
- AI evidence: [downstream:web/src/features/playground/lib/storage/storage.ts:57](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/playground/lib/storage/storage.ts#L57) — 下游对照：const saved = localStorage.getItem(STORAGE_KEYS.MESSAGES)

#### commit-96275d224bfc

- Title: fix(web): show date pickers and relative times in the interface language
- Feature/group: fix(web): show date pickers and relative times in the interface language
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/__tests__/date-picker-locale.test.tsx`, `web/src/components/date-picker.tsx`, `web/src/components/datetime-picker.tsx`, `web/src/features/security/components/__tests__/login-session-item.test.tsx`, `web/src/features/security/components/login-session-item.tsx`, `web/src/features/security/components/passkey-card.tsx`, `web/src/features/system-settings/content/announcements-section.tsx`
- Overlap with downstream changes: web/src/features/system-settings/content/announcements-section.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 日期与相对时间跟随界面语言
- AI recommendation: `migrate` — 本地日期选择器映射键为 zh、实际界面为 zhCN/zhTW，导致中文回退英语。迁入正确映射和 toIntlLocale；安全页迁移尚未批准的部分不要顺带搬入。
- AI risks: 时区和时间戳值不能被翻译改变；七语言、繁体中文、切换语言及相对时间需覆盖。
- AI dependencies: commit-5fe8917be728
- AI suggested paths (not approval): web/src/components/date-picker.tsx, web/src/components/datetime-picker.tsx, web/src/features/system-settings/content/announcements-section.tsx, web/src/components/__tests__/date-picker-locale.test.tsx
- AI evidence: [upstream:web/src/components/date-picker.tsx:35](https://github.com/QuantumNous/new-api/blob/96275d224bfcf6b5545ca7e57b3bf90f7c3ad396/web/src/components/date-picker.tsx#L35) — 固定源码：zhTW,
- AI evidence: [downstream:web/src/components/date-picker.tsx:34](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/components/date-picker.tsx#L34) — 下游对照：zh: zhCN,

#### commit-93ea55f3c310

- Title: feat(web): pick dates and times down to the second
- Feature/group: feat(web): pick dates and times down to the second
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/components/__tests__/datetime-picker-time.test.tsx`, `web/src/components/datetime-picker.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-filter.test.tsx`, `web/src/features/usage-logs/components/compact-date-time-range-picker.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/compact-date-time-range-picker.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 日期时间选择精确到秒
- AI recommendation: `migrate` — 本地 DateTimePicker 的 time 状态为00:00，选择会丢失秒；上游加入 HH:mm:ss 与秒输入。可按现有日志筛选API的秒级时间戳适配。
- AI risks: 开始/结束边界、清空、夏令时、旧分钟值和预设当天最后一秒必须一致。
- AI dependencies: commit-96275d224bfc
- AI suggested paths (not approval): web/src/components/datetime-picker.tsx, web/src/features/usage-logs/components/compact-date-time-range-picker.tsx, web/src/components/__tests__/datetime-picker-time.test.tsx
- AI evidence: [upstream:web/src/components/datetime-picker.tsx:66](https://github.com/QuantumNous/new-api/blob/93ea55f3c3102f7bc11cd005570c7965f0f22a03/web/src/components/datetime-picker.tsx#L66) — 固定源码：const [time, setTime] = React.useState<string>('00:00:00')
- AI evidence: [downstream:web/src/components/datetime-picker.tsx:65](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/components/datetime-picker.tsx#L65) — 下游对照：const [time, setTime] = React.useState<string>('00:00')

#### commit-4b13d9388ed6

- Title: fix(relayconvert): drop non-string enum values when converting OpenAI tool schemas (#7729)
- Feature/group: fix(relayconvert): drop non-string enum values when converting OpenAI tool schemas (#7729)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/shared/gemini/schema.go`, `relaykit/relayconvert/internal/shared/gemini/schema_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Gemini 非字符串 enum 降级
- AI recommendation: `defer` — 上游为避免legacy Schema 400直接删除含数字/布尔的enum，这会扩大允许值域。后续273已提供JSON Schema保留方案，建议优先评估无损路线，不独立迁入静默删约束。
- AI risks: 数字/布尔enum、混合类型与嵌套schema都可能丢业务校验；仅legacy端点确需降级时应有明确诊断并获确认。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:relaykit/relayconvert/internal/shared/gemini/schema.go:126](https://github.com/QuantumNous/new-api/blob/4b13d9388ed6f43118707534daad4903e5d22745/relaykit/relayconvert/internal/shared/gemini/schema.go#L126) — 上游检测到非字符串enum后删除约束。
- AI evidence: [downstream:relaykit/relayconvert/internal/shared/gemini/schema.go:13](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/shared/gemini/schema.go#L13) — 本地仍将enum视为保留字段。

#### commit-801456a061fb

- Title: fix(relay): send a close frame when a Responses WebSocket session ends
- Feature/group: fix(relay): send a close frame when a Responses WebSocket session ends
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/responses_websocket_test.go`, `relay/responses_websocket.go`, `relay/responses_websocket_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Responses WS 正常关闭帧
- AI recommendation: `defer` — 本地未实现Responses WS，此提交为新会话统一shutdown加closeCode/reason和Once，修复断连/鉴权失败收尾。应作为WS整组验收的必要后续，不直接迁入孤立会话文件。
- AI risks: 死锁、并发写close/data、上游正常关闭与内部错误码、workers退出和结算恰好一次均需并发回归。
- AI dependencies: commit-9fe0457ee1f4, commit-75f3d246e2b1, commit-ae249f4ecbaf
- AI suggested paths (not approval):
- AI evidence: [upstream:relay/responses_websocket.go:100](https://github.com/QuantumNous/new-api/blob/801456a061fb9988957341c4eb5aa8ec20c40de1/relay/responses_websocket.go#L100) — 上游限制会话关闭只执行一次。
- AI evidence: [downstream:router/relay-router.go:136](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/router/relay-router.go#L136) — 本地仍只有HTTP Responses入口。

#### commit-c10780e054e6

- Title: fix(relayconvert): keep JSON Schema constraints in tools and structured output sent to Gemini
- Feature/group: fix(relayconvert): keep JSON Schema constraints in tools and structured output sent to Gemini
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/common/relay_info.go`, `relay/common/relay_info_gemini_settings_test.go`, `relaykit/relayconvert/convmeta/options.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/shared/gemini/schema.go`, `relaykit/relayconvert/internal/shared/gemini/schema_test.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/model.go`, `relaykit/relayconvert/request_registry.go`
- Overlap with downstream changes: relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Gemini JSON Schema 约束保留
- AI recommendation: `migrate` — 本地转换器通过CleanFunctionParameters/RemoveAdditionalProperties清理schema，可能丢additionalProperties和复杂约束；上游按v1beta/Vertex能力用parametersJsonSchema/responseJsonSchema保留，否则诊断降级。建议按本地API版本配置选择性迁入最终方案。
- AI risks: 须测试v1、v1beta、Vertex，工具/结构化输出两路径、嵌套enum/$ref/additionalProperties及未知版本；不能默认所有Gemini端点支持JSON Schema。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/common/relay_info.go, relaykit/relayconvert/convmeta/options.go, relaykit/dto/gemini.go, relaykit/relayconvert/internal/shared/gemini/schema.go, relaykit/relayconvert/internal/shared/gemini/schema_test.go, relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go, relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go, relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req_test.go, relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req_test.go
- AI evidence: [upstream:relay/common/relay_info.go:973](https://github.com/QuantumNous/new-api/blob/c10780e054e6e518c9d0d8a903526663c9b57d75/relay/common/relay_info.go#L973) — 上游宿主按渠道类型和Gemini API版本声明支持。
- AI evidence: [downstream:relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go:231](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go#L231) — 本地结构化输出仍无条件清理部分schema约束。

#### commit-42e694eb60fe

- Title: fix(web): map Meta Muse Spark models to Meta icon (#7734)
- Feature/group: fix(web): map Meta Muse Spark models to Meta icon (#7734)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/lib/model-provider.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Meta Muse Spark 显示图标识别
- AI recommendation: `defer` — 只扩展前项新增统一 MODEL_PROVIDER_RULES 的 meta/和muse-前缀；本地尚无该模块，待提供者显示识别迁移时一起纳入最终规则。
- AI risks: 不能让 muse 子串误匹配其他供应商，更不能改变实际路由/计费。
- AI dependencies: commit-d92612038410
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/lib/model-provider.ts:178](https://github.com/QuantumNous/new-api/blob/42e694eb60fe3191795c00d416571ced716eb9c4/web/src/lib/model-provider.ts#L178) — 固定源码：keywords: ['meta-llama', 'meta/', 'muse-', 'llama-', 'llama2', 'llama3'],
- AI evidence: [downstream:web/src/features/channels/lib/model-categories.ts:28](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/channels/lib/model-categories.ts#L28) — 下游对照：const MODEL_CATEGORY_RULES: readonly ModelCategoryRule[] = [

#### commit-3827fd83076b

- Title: fix(web): show cache read price on expression model cards (#7757)
- Feature/group: fix(web): show cache read price on expression model cards (#7757)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-card.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 表达式卡片显示缓存读取价
- AI recommendation: `migrate` — 建议单独迁移缓存读取价展示：本地动态价卡片只渲染primaryEntries，缓存价在secondaryEntries中被遗漏。保留本地布局，将真实cr单价加入展示，不同时引入任务插件多供应商或新计费单位。
- AI risks: 特殊表达式不能错误推导缓存价，图片缓存不得误当文本缓存；七语言标签及空/零价格需验证。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/pricing/__tests__/model-cards.test.tsx, web/src/features/pricing/components/model-card.tsx
- AI evidence: [upstream:web/src/features/pricing/components/model-card.tsx:111](https://github.com/QuantumNous/new-api/blob/3827fd83076b9cbaa3f1d9f9d2957ae653573ed7/web/src/features/pricing/components/model-card.tsx#L111) — 从次级价格字段补充缓存读取价格；代码：const cacheReadEntries = dynamicSummary.secondaryEntries.filter(
- AI evidence: [downstream:web/src/features/pricing/components/model-card.tsx:110](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/pricing/components/model-card.tsx#L110) — 本地动态表达式卡片只直接渲染主价格项；代码：{dynamicSummary.primaryEntries.map((entry) => (

#### commit-3f926c3ef497

- Title: fix(metrics): count completed streams after context cancellation (#7638)
- Feature/group: fix(metrics): count completed streams after context cancellation (#7638)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `pkg/perf_metrics/outcome.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 已完成流在取消后的指标归类
- AI recommendation: `defer` — 该修复依赖前述协议Outcome及请求边界指标，本地没有ClassifyRelayOutcome。若采用141，必须一起纳入：已正常完成的流即使稍后context取消仍应计成功。
- AI risks: 不能把生成中取消也算成功；完成标记、EOF/Done、错误对象和流状态应同时满足。
- AI dependencies: commit-8b2c7105f91f, commit-a5b40663d330
- AI suggested paths (not approval):
- AI evidence: [upstream:pkg/perf_metrics/outcome.go:30](https://github.com/QuantumNous/new-api/blob/3f926c3ef49745be273cd8f753fb691abdd7b836/pkg/perf_metrics/outcome.go#L30) — 上游仅对已正常终止无错流豁免后续取消。
- AI evidence: [downstream:pkg/perf_metrics/metrics.go:27](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/pkg/perf_metrics/metrics.go#L27) — 本地仍用旧采样入口，没有Outcome分类器。

#### commit-bbf60b49f6a4

- Title: fix(logs): vertically align billing source icons on mobile (#7760)
- Feature/group: fix(logs): vertically align billing source icons on mobile (#7760)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 移动端扣费来源图标对齐
- AI recommendation: `defer` — 该补丁给来源图标包inline-flex以修复移动端纵向对齐，本地下游当前使用SubscriptionBadge而没有该来源图标组合。应等待是否采用152/160/161的展示方案，避免单独复制不适用class。
- AI risks: 保留当前费用、工具费和可访问标签；图标视觉修复不应改变金额来源。
- AI dependencies: commit-9e905a83e1d6, commit-ea336ed6654a, commit-96b0da2273a1
- AI suggested paths (not approval):
- AI evidence: [upstream:web/src/features/usage-logs/components/log-cost-display.tsx:118](https://github.com/QuantumNous/new-api/blob/bbf60b49f6a4cab709e49e57d45f7911fc7dccae/web/src/features/usage-logs/components/log-cost-display.tsx#L118) — 修复对象是上游来源图标容器；代码：className='inline-flex shrink-0 cursor-help items-center'
- AI evidence: [downstream:web/src/features/usage-logs/components/log-cost-display.tsx:95](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/usage-logs/components/log-cost-display.tsx#L95) — 本地订阅使用独立Badge组件而非上游图标容器；代码：function SubscriptionBadge(props: { quota: number }) {

#### commit-b22a43b6e89e

- Title: fix: mark task settlement logs with is_task for tiered billing display (#7446)
- Feature/group: fix: mark task settlement logs with is_task for tiered billing display (#7446)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `service/task_billing.go`
- Overlap with downstream changes: service/task_billing.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 任务结算日志is_task标记
- AI recommendation: `covered` — 该提交唯一运行时变化是在taskBillingOther中加入is_task=true。冻结下游的同一任务账务Other构造器已明确写入该字段，提交日志也已写入；可建议covered并保留代码证据，无需重复改写类型化LogOther或业务账务。
- AI risks: 只认定本单一标记修复已覆盖，不代表上游插件表达式任务计费已完整迁入。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:service/task_billing.go:151](https://github.com/QuantumNous/new-api/blob/b22a43b6e89e5349d8cdbddc88d570b16e15fb1f/service/task_billing.go#L151) — 源提交 taskBillingOther 新增任务标记；代码：other.SetPublic("is_task", true)
- AI evidence: [downstream:service/task_billing.go:161](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/service/task_billing.go#L161) — 冻结下游 taskBillingOther 已在返回前无条件设置相同标记；代码：other["is_task"] = true

#### commit-34b19421237f

- Title: fix(pricing): allow root access when the model square is disabled (#7761)
- Feature/group: fix(pricing): allow root access when the model square is disabled (#7761)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `middleware/header_nav.go`, `middleware/header_nav_test.go`
- Overlap with downstream changes: middleware/header_nav.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 禁用模型广场时保留Root管理访问
- AI recommendation: `migrate` — 建议限定仅pricing模块在禁用时通过现有RootAuth允许Root访问管理，普通用户及非Root管理员继续拒绝。本地目前统一403，会锁死Root读取定价的入口；应使用本地会话与权限夹具适配，保留其他模块禁用规则。
- AI risks: Authorization存在不代表Root，必须完整校验过期/撤销/禁用账户；不能重新开放公共模型广场或绕过其他模块。
- AI dependencies: none identified
- AI suggested paths (not approval): middleware/header_nav.go, middleware/header_nav_test.go
- AI evidence: [upstream:middleware/header_nav.go:110](https://github.com/QuantumNous/new-api/blob/34b19421237fc4695e919cdd799c1b7813023085/middleware/header_nav.go#L110) — 只有pricing显式带凭据时执行真正Root鉴权；代码：RootAuth()(c)
- AI evidence: [downstream:middleware/header_nav.go:107](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/middleware/header_nav.go#L107) — 本地禁用模块当前无Root价格管理例外；代码：if !access.Enabled {

#### commit-ba0e478e061e

- Title: fix(relaykit): preserve assistant text with tool calls (#6871)
- Feature/group: fix(relaykit): preserve assistant text with tool calls (#6871)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude 工具调用伴随答案文本
- AI recommendation: `merge` — 本地Claude→Chat在有tool_calls时不写同一assistant消息中的media/text；上游移除len(toolCalls)==0限制。可优先完整采用单条件修复，并加入文字+工具调用同时保留的回归。
- AI risks: 确保assistant内容与tool_calls可共存、tool_result顺序不变，不能把用户工具结果媒体误归到assistant。
- AI dependencies: none identified
- AI suggested paths (not approval): relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go, relaykit/relayconvert/internal/claude_messages/to_oai_chat_req_test.go
- AI evidence: [upstream:relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go:273](https://github.com/QuantumNous/new-api/blob/ba0e478e061ee0c824922d89c05620de878103cf/relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go#L273) — 上游允许消息内容与工具调用并存。
- AI evidence: [downstream:relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go:200](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go#L200) — 本地当前条件会丢带工具调用的文本。

#### commit-fc08b76f0fb3

- Title: fix(web): preserve recorded billing tiers in usage log previews (#7763)
- Feature/group: fix(web): preserve recorded billing tiers in usage log previews (#7763)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 日志预览保留冻结计价档
- AI recommendation: `migrate` — 建议优先迁移：无法用现有解析器展示表达式时，先展示日志冻结的matched_tier，缺失再用特殊表达式提示。本地仍显示No matching results，容易把有效历史计价误判为未匹配；仅改文案选择，不重算历史扣费。
- AI risks: 任务和普通relay两分支都需验证；保留后台饱和、有效调账及原始证据，不用当前配置推断历史档位。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx, web/src/features/usage-logs/components/columns/common-logs-columns.tsx
- AI evidence: [upstream:web/src/features/usage-logs/components/columns/common-logs-columns.tsx:213](https://github.com/QuantumNous/new-api/blob/fc08b76f0fb3f95e7cb81667e7bf350200147fb3/web/src/features/usage-logs/components/columns/common-logs-columns.tsx#L213) — 未知表达式显示保留冻结matched_tier；代码：text: `${t('Dynamic Pricing')} · ${other.matched_tier || t('Special billing expression')}`,
- AI evidence: [downstream:web/src/features/usage-logs/components/columns/common-logs-columns.tsx:217](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/usage-logs/components/columns/common-logs-columns.tsx#L217) — 本地遇到解析回落时仍显示未匹配；代码：text: `${t('Dynamic Pricing')} · ${t('No matching results')}`,

#### commit-4bfcf2812198

- Title: chore(web): bump @lobehub/icons to 5.22.0 (#7764)
- Feature/group: chore(web): bump @lobehub/icons to 5.22.0 (#7764)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Lobe 图标依赖版本升级
- AI recommendation: `upgrade` — 本地 @lobehub/icons 仍为^5.10.1，上游升至~5.22.0。可在独立依赖批次升级并在本地锁文件重算，确认懒加载/自定义图标仍完整。
- AI risks: 不能直接套用整个上游bun.lock；验证包导出路径、变体、Windows产物与UI回归，不把版本升级等同安全保证。
- AI dependencies: commit-9c3d3aeb3c30, commit-00e8a00cb4ae
- AI suggested paths (not approval): web/package.json, web/bun.lock, web/src/lib/__tests__/lobe-icon.test.tsx
- AI evidence: [upstream:web/package.json:39](https://github.com/QuantumNous/new-api/blob/4bfcf2812198913156a3e88a43e746b61a7b6ef5/web/package.json#L39) — 固定源码："@lobehub/icons": "~5.22.0",
- AI evidence: [downstream:web/package.json:35](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/package.json#L35) — 下游对照："@lobehub/icons": "^5.10.1",

#### commit-9948b1ef14ea

- Title: fix(pricing): accept declared endpoint arrays in pricing conversion (#7724)
- Feature/group: fix(pricing): accept declared endpoint arrays in pricing conversion (#7724)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/model_pricing_conversion_endpoints_test.go`, `model/model_pricing_conversion.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 旧价转换接受endpoint数组声明
- AI recommendation: `defer` — 转换器兼容元数据endpoint的数组和对象两种形态，避免合法数组被误判为复杂自定义路由而拒绝转换。是98的正确性后续，本地尚无该转换器，应随旧价转换组采用最终实现。
- AI risks: 未知/重复/非法端点与真正自定义路径不能被误判为安全可转换。
- AI dependencies: commit-0c76e4dae77a, commit-25ec832faa48
- AI suggested paths (not approval):
- AI evidence: [upstream:model/model_pricing_conversion.go:149](https://github.com/QuantumNous/new-api/blob/9948b1ef14ea17fb23539386c0cb09018671c955/model/model_pricing_conversion.go#L149) — 转换明确解析两种endpoint形态；代码：func declaredEndpointNames(raw string) ([]string, error) {

#### commit-2ed058795bf0

- Title: fix(ci): stabilize checks and configure Bun dependency updates (#7765)
- Feature/group: fix(ci): stabilize checks and configure Bun dependency updates (#7765)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `.github/dependabot.yml`, `.github/workflows/ci.yml`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `web/bun.lock`, `web/package.json`, `web/src/components/ui/__tests__/markdown.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-editor.test.tsx`
- Overlap with downstream changes: .github/workflows/ci.yml
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: CI稳定性与DOMPurify/KaTeX依赖混合更新
- AI recommendation: `defer` — 该提交混合Bun Dependabot、上游ci.yml checkout、认证测试、DOMPurify3.4.16与KaTeX0.18.2。下游CI结构不同且markdown/数学渲染依赖覆盖不同，需拆为专门依赖/测试PR评审。
- AI risks: 不要自动搬保护CI；DOMPurify override必须与直接依赖一致，数学渲染升级需兼容/安全测试，不以测试调整掩盖失败。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/package.json:53](https://github.com/QuantumNous/new-api/blob/2ed058795bf0ce968134aa8db18b0dd86b54c30e/web/package.json#L53) — 固定源码："dompurify": "3.4.16",
- AI evidence: [downstream:web/package.json:49](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/package.json#L49) — 下游对照："dompurify": "3.4.13",

#### commit-2b85a8fc67bd

- Title: fix(auth): show legal consent warning on password login (#7771)
- Feature/group: fix(auth): show legal consent warning on password login (#7771)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/src/features/auth/sign-in/components/user-auth-form.tsx`
- Overlap with downstream changes: web/src/features/auth/sign-in/components/user-auth-form.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 密码登录未同意条款时显示提示
- AI recommendation: `migrate` — 本地 onSubmit 已检查法律同意，但密码登录提交按钮也禁用，导致用户无法触发具体提示。可仅让密码按钮进入已有校验，保持 OAuth/短信/Passkey 的确认策略。
- AI risks: 不能移除 onSubmit 的强制同意检查；覆盖键盘提交、复选框、密码失败和重复请求。
- AI dependencies: none identified
- AI suggested paths (not approval): web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/auth/sign-in/components/__tests__/legal-consent.test.tsx
- AI evidence: [upstream:web/src/features/auth/sign-in/components/user-auth-form.tsx:402](https://github.com/QuantumNous/new-api/blob/2b85a8fc67bdb6d8d41a25f60f734dd3b9917618/web/src/features/auth/sign-in/components/user-auth-form.tsx#L402) — 固定源码：disabled={isLoading}
- AI evidence: [downstream:web/src/features/auth/sign-in/components/user-auth-form.tsx:375](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/src/features/auth/sign-in/components/user-auth-form.tsx#L375) — 下游对照：disabled={isLoading || (requiresLegalConsent && !agreedToLegal)}

#### commit-767a82b87647

- Title: fix(claude): forward thinking.block_binding and every anthropic-beta value (#7732)
- Feature/group: fix(claude): forward thinking.block_binding and every anthropic-beta value (#7732)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `relay/channel/aws/dto.go`, `relay/channel/aws/relay_aws_test.go`, `relay/channel/claude/adaptor.go`, `relay/channel/claude/adaptor_test.go`, `relay/channel/claude/relay_claude_test.go`, `relaykit/dto/claude.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Claude thinking 扩展与多 beta 头
- AI recommendation: `migrate` — 本地Thinking DTO缺block_binding/未知字段保留，anthropic-beta使用Header.Get只取首个值。建议独立迁入原生thinking透明往返和全部beta值合并；不用为此强制采用统一reasoning重构。
- AI risks: 未知thinking字段与显式客户端block_binding要保真；模型后缀调整只能在adaptive时保留合法控制；AWS/原生Claude均需验证，多头不能丢或伪造beta权限。; Moonshot兼容入口不能把普通OpenAI上游伪装成支持Claude preserved-thinking；补对应边界回归。
- AI dependencies: none identified
- AI suggested paths (not approval): relay/channel/claude/adaptor.go, relay/channel/claude/adaptor_test.go, relay/channel/aws/dto.go, relay/channel/aws/dto_test.go, relaykit/dto/claude.go, relaykit/dto/claude_test.go, relay/facade/moonshot/facade.go, relay/facade/moonshot/facade_test.go
- AI evidence: [upstream:relay/channel/claude/adaptor.go:109](https://github.com/QuantumNous/new-api/blob/767a82b8764778b847462ebecdc231eacd1e282c/relay/channel/claude/adaptor.go#L109) — 上游合并客户端多行beta值。
- AI evidence: [downstream:relay/channel/claude/adaptor.go:76](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/relay/channel/claude/adaptor.go#L76) — 本地只读取第一个beta头。

#### commit-ba29f0645394

- Title: chore(deps-dev): bump @xmldom/xmldom from 0.8.13 to 0.8.15 in /electron (#7187)
- Feature/group: chore(deps-dev): bump @xmldom/xmldom from 0.8.13 to 0.8.15 in /electron (#7187)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Electron XML DOM 锁定依赖补丁
- AI recommendation: `upgrade` — 上游只把Electron开发依赖树中的@xmldom/xmldom从0.8.13升至0.8.15，本地锁文件仍是旧版本。建议独立锁文件更新，保留本地打包和许可证配置。
- AI risks: 在本地现有Electron版本下重算最小锁差异并验证打包；没有核实公告，不能把它宣称为已修某CVE。
- AI dependencies: none identified
- AI suggested paths (not approval): electron/package-lock.json
- AI evidence: [upstream:electron/package-lock.json:716](https://github.com/QuantumNous/new-api/blob/ba29f06453942d9b5c0b32330340b8a000692c4f/electron/package-lock.json#L716) — 固定源码："version": "0.8.15",
- AI evidence: [downstream:electron/package-lock.json:715](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/electron/package-lock.json#L715) — 下游对照："node_modules/@xmldom/xmldom": {

#### commit-b8c2dc12660d

- Title: build(deps): bump yace from 1.1.0 to 1.1.1 in /web (#7770)
- Feature/group: build(deps): bump yace from 1.1.0 to 1.1.1 in /web (#7770)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Yace 编辑器补丁升级
- AI recommendation: `upgrade` — 本地yace为^1.1.0，上游升级到^1.1.1；可独立更新本地声明/锁文件并验证当前JSON/代码编辑器。
- AI risks: 需要确认现有编辑器API、输入组合、读写与生产构建；不把未查变更日志的补丁假定零风险。
- AI dependencies: none identified
- AI suggested paths (not approval): web/package.json, web/bun.lock
- AI evidence: [upstream:web/package.json:84](https://github.com/QuantumNous/new-api/blob/b8c2dc12660dceb2a905008a672f8bc0260c98c8/web/package.json#L84) — 固定源码："yace": "^1.1.1",
- AI evidence: [downstream:web/package.json:79](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/package.json#L79) — 下游对照："yace": "^1.1.0",

#### commit-c730add954e1

- Title: build(deps): bump stream-markdown-parser from 1.1.3 to 1.2.18 in /web (#7769)
- Feature/group: build(deps): bump stream-markdown-parser from 1.1.3 to 1.2.18 in /web (#7769)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 流式 Markdown 解析器升级
- AI recommendation: `upgrade` — 本地^1.0.9，上游目标为^1.2.18（提交标题的1.1.3不是实际声明）。建议独立升级并用真实分块Markdown回归，不覆盖聊天存储和自有渲染。
- AI risks: 需覆盖未闭合代码/数学/链接、增量重算、消毒边界和长响应性能；重新生成本地锁文件。
- AI dependencies: none identified
- AI suggested paths (not approval): web/package.json, web/bun.lock
- AI evidence: [upstream:web/package.json:77](https://github.com/QuantumNous/new-api/blob/c730add954e1b565656a9320444cac5fee12cfe4/web/package.json#L77) — 固定源码："stream-markdown-parser": "^1.2.18",
- AI evidence: [downstream:web/package.json:72](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/package.json#L72) — 下游对照："stream-markdown-parser": "^1.0.9",

#### commit-82306b3c3594

- Title: build(deps-dev): bump @tanstack/router-plugin in /web (#7768)
- Feature/group: build(deps-dev): bump @tanstack/router-plugin in /web (#7768)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: TanStack 路由生成插件补丁升级
- AI recommendation: `upgrade` — 本地router-plugin仍^1.168.19，上游升至^1.168.42；建议独立升级核对与当前react-router版本的生成兼容，不直接采用上游routeTree。
- AI risks: 检查类型生成、懒加载、OAuth路由和构建无脏文件；本地自定义路由必须保留。
- AI dependencies: none identified
- AI suggested paths (not approval): web/package.json, web/bun.lock
- AI evidence: [upstream:web/package.json:95](https://github.com/QuantumNous/new-api/blob/82306b3c35945bc0ed2c567688afe28d8a281300/web/package.json#L95) — 固定源码："@tanstack/router-plugin": "^1.168.42",
- AI evidence: [downstream:web/package.json:89](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/package.json#L89) — 下游对照："@tanstack/router-plugin": "^1.168.19",

#### commit-bf4e1ebfb885

- Title: build(deps-dev): bump oxfmt from 0.57.0 to 0.72.0 in /web (#7766)
- Feature/group: build(deps-dev): bump oxfmt from 0.57.0 to 0.72.0 in /web (#7766)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Oxfmt 开发工具版本升级
- AI recommendation: `defer` — 从0.57到0.72跨多个版本，会改变格式输出；下游有protected-headers formatter和大量自定义文件。建议专门工具PR先验证幂等与保护头，再选择，避免污染功能同步diff。
- AI risks: 不得因format改写QuantumNous版权头或无关文件；校验参数、格式配置及仅格式变更隔离。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:web/package.json:108](https://github.com/QuantumNous/new-api/blob/bf4e1ebfb88579930d75d852d2d2bca74cbd5f62/web/package.json#L108) — 固定源码："oxfmt": "^0.72.0",
- AI evidence: [downstream:web/package.json:101](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/package.json#L101) — 下游对照："oxfmt": "^0.57.0",

#### commit-7e25797d1468

- Title: chore(deps-dev): bump electron from 39.8.10 to 41.10.6 in /electron (#7625)
- Feature/group: chore(deps-dev): bump electron from 39.8.10 to 41.10.6 in /electron (#7625)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`, `electron/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Electron 39至41主版本升级
- AI recommendation: `defer` — 上游从39.8.10升到41.10.6是主版本跨度；下游有自己的启动、更新、打包和许可证附带逻辑，当前没有跨平台验证，不能夹在功能同步中批准。
- AI risks: 需要核对Node/Chromium、sandbox/preload、自动更新、签名、公证和Windows/macOS/Linux打包后再选。
- AI dependencies: none identified
- AI suggested paths (not approval):
- AI evidence: [upstream:electron/package.json:28](https://github.com/QuantumNous/new-api/blob/7e25797d14681e4e16c2b46831450e7d6a5ccc74/electron/package.json#L28) — 固定源码："electron": "41.10.6",
- AI evidence: [downstream:electron/package.json:28](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/electron/package.json#L28) — 下游对照："electron": "39.8.10",

#### commit-7f51a97226c4

- Title: chore(deps-dev): bump js-yaml from 4.3.1 to 4.3.2 in /electron (#7327)
- Feature/group: chore(deps-dev): bump js-yaml from 4.3.1 to 4.3.2 in /electron (#7327)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Electron js-yaml 锁定依赖补丁
- AI recommendation: `upgrade` — 上游把Electron开发树js-yaml4.3.1升4.3.2，本地仍4.3.1。建议最小锁文件升级，独立验证当前Electron打包。
- AI risks: 不能用后续空重复提交代表已覆盖；保留当前直接依赖和平台配置，不拷贝整个上游依赖树。
- AI dependencies: none identified
- AI suggested paths (not approval): electron/package-lock.json
- AI evidence: [upstream:electron/package-lock.json:2574](https://github.com/QuantumNous/new-api/blob/7f51a97226c43bc082a065bdc8b40b719c4bdcce/electron/package-lock.json#L2574) — 固定源码："version": "4.3.2",
- AI evidence: [downstream:electron/package-lock.json:2621](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/electron/package-lock.json#L2621) — 下游对照："node_modules/js-yaml": {

#### commit-1c58f3d161fa

- Title: chore(deps-dev): bump fast-uri from 3.1.5 to 3.1.7 in /electron (#7169)
- Feature/group: chore(deps-dev): bump fast-uri from 3.1.5 to 3.1.7 in /electron (#7169)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Electron fast-uri 锁定依赖补丁
- AI recommendation: `upgrade` — 上游把Electron开发树fast-uri3.1.5升3.1.7，本地仍3.1.5。可在本地最小化重算锁文件，与web自身fast-uri override分别处理。
- AI risks: 测试builder配置/JSON schema与打包；不得误把electron依赖更新应用到web override。
- AI dependencies: none identified
- AI suggested paths (not approval): electron/package-lock.json
- AI evidence: [upstream:electron/package-lock.json:2052](https://github.com/QuantumNous/new-api/blob/1c58f3d161fa32755b372d9449f2d5e077d50a5d/electron/package-lock.json#L2052) — 固定源码："version": "3.1.7",
- AI evidence: [downstream:electron/package-lock.json:2060](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/electron/package-lock.json#L2060) — 下游对照："node_modules/fast-uri": {

#### commit-3fe394cdba28

- Title: chore(deps): bump github.com/gorilla/websocket from 1.5.0 to 1.5.3 (#7001)
- Feature/group: chore(deps): bump github.com/gorilla/websocket from 1.5.0 to 1.5.3 (#7001)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `go.mod`, `go.sum`
- Overlap with downstream changes: go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Gorilla WebSocket 依赖补丁升级
- AI recommendation: `upgrade` — 本地root依赖v1.5.0，上游升v1.5.3；范围仅go.mod/go.sum，可独立升级并回归当前Realtime/WebSocket relay及下游连接管理。
- AI risks: 不隐式启用上游新增Responses WebSocket能力；检查握手/关闭/读写并发，root和relaykit模块分别构建。
- AI dependencies: none identified
- AI suggested paths (not approval): go.mod, go.sum
- AI evidence: [upstream:go.mod:30](https://github.com/QuantumNous/new-api/blob/3fe394cdba280c00bc566de493b5fc8399b0a7d1/go.mod#L30) — 固定源码：github.com/gorilla/websocket v1.5.3
- AI evidence: [downstream:go.mod:30](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/go.mod#L30) — 下游对照：github.com/gorilla/websocket v1.5.0

#### commit-9f048107edbb

- Title: chore(deps): bump ai from 7.0.31 to 7.0.136 in /web (#7767)
- Feature/group: chore(deps): bump ai from 7.0.31 to 7.0.136 in /web (#7767)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: AI SDK 同主版本升级
- AI recommendation: `upgrade` — 本地AI SDK声明^7.0.14，上游实际改到^7.0.136（不能仅按标题的7.0.31判断）。建议独立依赖升级，回归现有Playground流式协议。
- AI risks: 需检查工具调用、流事件、错误取消、tokens统计与当前React19接口；同主版本不等于行为完全一致。
- AI dependencies: none identified
- AI suggested paths (not approval): web/package.json, web/bun.lock
- AI evidence: [upstream:web/package.json:46](https://github.com/QuantumNous/new-api/blob/9f048107edbba2156691718126de7c4577614533/web/package.json#L46) — 固定源码："ai": "^7.0.136",
- AI evidence: [downstream:web/package.json:42](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/web/package.json#L42) — 下游对照："ai": "^7.0.14",

#### commit-3c9fb2a0de16

- Title: chore(deps-dev): bump brace-expansion from 1.1.16 to 1.1.21 in /electron (#7623)
- Feature/group: chore(deps-dev): bump brace-expansion from 1.1.16 to 1.1.21 in /electron (#7623)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Electron brace-expansion 锁定依赖补丁
- AI recommendation: `upgrade` — 上游更新多个Electron嵌套brace-expansion1.1.16到1.1.21，本地仍有旧嵌套版本。可在现有builder版本下最小升级。
- AI risks: 不能只改首个条目遗漏嵌套副本；不要把web的2.1.1 override当作Electron已修复证据，验证打包。
- AI dependencies: none identified
- AI suggested paths (not approval): electron/package-lock.json
- AI evidence: [upstream:electron/package-lock.json:52](https://github.com/QuantumNous/new-api/blob/3c9fb2a0de168d88ce5bffbf44acecdc1cdcdd39/electron/package-lock.json#L52) — 固定源码："version": "1.1.21",
- AI evidence: [downstream:electron/package-lock.json:41](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/electron/package-lock.json#L41) — 下游对照："node_modules/@electron/asar/node_modules/brace-expansion": {

#### commit-93cf4ba6fb92

- Title: chore(deps-dev): bump js-yaml from 4.3.1 to 4.3.2 in /electron (#7327)
- Feature/group: chore(deps-dev): bump js-yaml from 4.3.1 to 4.3.2 in /electron (#7327)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: (none)
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: js-yaml 重复空提交
- AI recommendation: `skip` — 该来源提交没有文件差异；实际js-yaml升级已在前面的293候选。显式跳过重复项，不能产生空cherry-pick或单独增加版本。
- AI risks: 293尚未人工选择/实施，因此本项不是证明下游已覆盖；历史应保留其空差异事实。
- AI dependencies: commit-7f51a97226c4
- AI suggested paths (not approval):
- AI evidence: [upstream:electron/package-lock.json:2573](https://github.com/QuantumNous/new-api/blob/93cf4ba6fb92e3f755fdeb6af9a9484bfa748816/electron/package-lock.json#L2573) — 固定源码："node_modules/js-yaml": {

#### commit-dd59854050f2

- Title: chore(deps-dev): bump undici from 6.28.0 to 6.29.0 in /electron (#7624)
- Feature/group: chore(deps-dev): bump undici from 6.28.0 to 6.29.0 in /electron (#7624)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: Electron undici 锁定依赖补丁
- AI recommendation: `upgrade` — 上游把Electron开发树undici6.28.0升6.29.0，本地仍旧值。建议独立最小锁更新并验证资源下载/打包，不影响网关HTTP实现。
- AI risks: Electron构建依赖与Go relay不是同一传输层；验证锁解析与可重复构建，没有公告证据不宣称具体漏洞已修。
- AI dependencies: none identified
- AI suggested paths (not approval): electron/package-lock.json
- AI evidence: [upstream:electron/package-lock.json:3735](https://github.com/QuantumNous/new-api/blob/dd59854050f26f4f67bb3dac5274af5e17013462/electron/package-lock.json#L3735) — 固定源码："version": "6.29.0",
- AI evidence: [downstream:electron/package-lock.json:3838](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/electron/package-lock.json#L3838) — 下游对照："node_modules/undici": {

#### commit-633b4c5b190d

- Title: feat(logs): export usage logs from the common log page (#7774)
- Feature/group: feat(logs): export usage logs from the common log page (#7774)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason:
- Dependencies: none declared; requires review
- Files: `controller/log.go`, `controller/log_export_test.go`, `controller/misc.go`, `controller/relay_error_log_test.go`, `controller/usedata.go`, `controller/user_manage_test.go`, `middleware/access_token_routes.go`, `model/log.go`, `model/option.go`, `router/api-router.go`, `service/log_export.go`, `setting/operation_setting/log_setting.go`, `web/src/features/auth/types.ts`, `web/src/features/system-settings/hooks/use-update-option.ts`, `web/src/features/system-settings/maintenance/__tests__/log-settings.test.tsx`, `web/src/features/system-settings/maintenance/history-log-cleanup.tsx`, `web/src/features/system-settings/maintenance/log-settings-section.tsx`, `web/src/features/system-settings/maintenance/server-log-files.tsx`, `web/src/features/system-settings/operations/index.tsx`, `web/src/features/system-settings/operations/section-registry.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/usage-logs/api.ts`, `web/src/features/usage-logs/components/__tests__/export-dialog.test.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-filter.test.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`, `web/src/features/usage-logs/components/dialogs/export-logs-dialog.tsx`, `web/src/features/usage-logs/components/logs-filter-toolbar.tsx`, `web/src/features/usage-logs/lib/__tests__/export.test.ts`, `web/src/features/usage-logs/lib/export.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/log.go, controller/misc.go, controller/user_manage_test.go, model/log.go, model/option.go, router/api-router.go, web/src/features/auth/types.ts, web/src/features/system-settings/hooks/use-update-option.ts, web/src/features/system-settings/maintenance/log-settings-section.tsx, web/src/features/system-settings/operations/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/features/usage-logs/api.ts, web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
- AI feature/group: 通用日志流式导出与查询范围
- AI recommendation: `defer` — 提供NDJSON流到前端CSV、每用户跨节点导出租约及365天搜索上限，功能有价值但上游从原始logs逐行导出。下游列表会叠加历史任务计价、有效调账、成本权限和不可变账务证据，不能直接套用原始日志导出或新默认查询限制。先明确导出合同后单独迁移。
- AI risks: 应与列表过滤/租户/财务权限完全一致；保留永久消费日志，不引入清理；处理取消、续租失败、大ID、CSV公式和多语言；范围上限不能阻断正式历史对账。
- AI dependencies: commit-057f71c2336c, commit-caca52f8dea0, commit-128089fa7ddf
- AI suggested paths (not approval):
- AI evidence: [upstream:model/log.go:638](https://github.com/QuantumNous/new-api/blob/633b4c5b190dc1af99a2814391354d86948e7943/model/log.go#L638) — 上游新增逐行读取原始日志并投影的导出入口；代码：func StreamUserLogs(ctx context.Context, userId int, filter LogFilter, fn func(*Log) error) error {
- AI evidence: [downstream:model/log.go:166](https://github.com/yeruyi1024/novamaas-workspace/blob/b2cb1c2473f1db803ff5bdaafc67699288bde713/model/log.go#L166) — 下游日志列表使用有效修正投影，原始导出会出现账目不一致；代码：err = projectBillingEffectiveLogs(logs, adjustments)
