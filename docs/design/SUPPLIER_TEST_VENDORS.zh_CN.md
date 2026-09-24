# 供应商自测：厂商请求与解析适配

页面自己填 Base URL / Key / 模型，下拉选 **通用 / GLM / Kimi / DeepSeek**。不猜域名，不改地址。选中后只换：请求带哪些键、响应认哪些键、缺了算跳过还是失败。

实现：`service/suppliertest/vendor.go`（发什么、严不严），`client.go`（怎么解析）。

## 原则

解析不分厂商。响应里几种别名都读进同一个桶（缓存 token、思考正文、思考用量）。厂商卡不负责「只认自家那一个键」，负责两件事：

1. **请求**：思考参数怎么发（`thinking` / `reasoning_effort` / 不发）。
2. **判定**：桶是空的时候，通用跳过；GLM / Kimi / DeepSeek 在文档要求有的项上失败。

中转站 400 的统一回退（与下拉无关）：`stream_options` 被拒就去掉再打；`thinking` 被拒就去掉再打。

## 响应：读哪些字段

| 含义 | JSON 路径 | 常见来源 |
|---|---|---|
| 缓存命中 | `usage.cached_tokens` | Kimi、多数中转站 |
| 缓存命中 | `usage.prompt_tokens_details.cached_tokens` | GLM |
| 缓存命中 | `usage.prompt_cache_hit_tokens` | DeepSeek |
| 缓存命中 | `usage.cache_read_input_tokens` | Anthropic 风格中转站 |
| 思考正文 | `choices[].delta/message.reasoning_content` | GLM、Kimi、DeepSeek |
| 思考正文 | `choices[].delta/message.reasoning` | 另一路别名 |
| 思考用量 | `usage.completion_tokens_details.reasoning_tokens` | GLM |
| 思考用量 | `usage.reasoning_tokens` | 顶层，少见 |
| 请求 ID | `id`，没有再用 `request_id` | GLM 常两个相同 |

以上任一路径有值，对应桶就算有。不按厂商拆解析器。

通用协议字段照旧：`usage.prompt_tokens`、`usage.completion_tokens`、`finish_reason`、`tool_calls`。

## 请求：按厂商发什么

缓存打法四档相同：长文放 `messages[0].role=system`，user 只换短问题（预热「请用一个词回复：ping」，探测用 follow-up）。

| 选择 | 思考请求 | 说明 |
|---|---|---|
| 通用 | `thinking: { "type": "enabled" }` | 对方不认则 400 后去掉再打 |
| GLM | 同上 | 文档是 `thinking.type` |
| Kimi，模型 `kimi-k3*` | **不发** `thinking`，发 `reasoning_effort: "low"` | 文档禁止乱传 thinking |
| Kimi，模型名含 `kimi-k2.7` | **不发** `thinking` | 默认开启 |
| DeepSeek，模型名含 `r1` / `reasoner` | **不发** `thinking` | 原生就会想 |
| 上述以外的模型名 | 仍发 `thinking.enabled` | 例如 `glm-4-flash`、`deepseek-chat` |

`stream_options.include_usage` 默认带；被 400 且错误信息不是 thinking/model/messages 时去掉再打。

## 缺字段怎么判

读到了 → 通过，不管实际是哪个键名。

没读到：

| 检查 | 通用 | GLM / Kimi / DeepSeek |
|---|---|---|
| `usage` | 跳过 | 失败 |
| 缓存字段（上表四个键都空） | 跳过 | 失败 |
| JSON 模式：HTTP 成功但正文不是 JSON | 跳过 | 失败 |
| tools：HTTP 成功但没有 `tool_calls` | 跳过 | 失败 |
| 对方直接 4xx 不接受 JSON/tools | 跳过 | 跳过（该模型可以没有这能力） |

思考内容（`reasoning_content` / `reasoning` / `reasoning_tokens` 都空）：

| 选择 | 没有思考内容 |
|---|---|
| 通用 | 跳过 |
| GLM，且模型含 `glm-5` 或 `glm-4.5` / `4.6` / `4.7` 前缀 | 失败 |
| Kimi，且 `kimi-k3*` / 含 `kimi-k2.6` / `kimi-k2.7` / `k2-thinking` | 失败 |
| DeepSeek，且模型含 `r1` 或 `reasoner` | 失败 |
| 同厂商的其它模型名（如 `glm-4-flash`、`deepseek-chat`） | 跳过 |

## 和页面的对应

下拉文案只提示「按谁的字段」，不改 Base URL。中转站别名模型只要手选对厂商，就会走该档的请求和失败规则；解析仍然吃所有别名。
