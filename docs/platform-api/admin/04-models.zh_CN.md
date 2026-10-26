# 模型接口

星途客户端当前三个分支仅对接**模型列表发现（`GET /v1/models`）与文本对话（`POST /v1/chat/completions`）**。图片生成、视频生成、语音与向量等属于 MaaS 平台已有的其他网关能力，星途客户端尚未接入；这些能力的字段定义以平台公共 API 文档为准，本手册不将其列为星途已支持能力。

## 接口索引

| 方法 | 路径 | 认证方式 | 用途 |
|---|---|---|---|
| GET | `/v1/models` | `Authorization: Bearer $MODEL_API_KEY` | 获取当前模型密钥可调用的模型 ID 列表 |
| POST | `/v1/chat/completions` | `Authorization: Bearer $MODEL_API_KEY` | 发起非流式或 SSE 流式文本对话（含思考输出与工具调用） |

以上接口均使用托管密钥接口返回的 `MODEL_API_KEY`，严禁传入登录 `ACCESS_TOKEN` 或携带刷新 Cookie。

## GET /v1/models

无请求体。星途客户端在同步托管密钥后调用该接口拉取当前账号可用的模型列表（`syncPlatformModels`），并按返回的 `data[].id` 去重生成客户端可选模型列表；若列表为空，客户端不会启用该模型提供商配置。

```bash
curl "$PLATFORM_URL/v1/models" \
  -H "Authorization: Bearer $MODEL_API_KEY"
```

成功响应示例：

```json
{
  "object": "list",
  "data": [
    {
      "id": "deepseek-r1",
      "object": "model",
      "created": 1792713600,
      "owned_by": "custom"
    },
    {
      "id": "gpt-4o-mini",
      "object": "model",
      "created": 1792713600,
      "owned_by": "openai"
    }
  ]
}
```

| 响应字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `object` | string | 否 | 固定为 `"list"` |
| `data` | array | 是 | 可用模型对象数组 |
| `data[].id` | string | 是 | 模型唯一标识（非空字符串），客户端将其作为 `POST /v1/chat/completions` 的 `model` 参数 |
| `data[].object`、`owned_by` | string | 否 | 模型元数据 |

## POST /v1/chat/completions

使用 `Authorization: Bearer $MODEL_API_KEY`，请求头 `Content-Type: application/json`；当 `stream: true` 时响应头为 `Content-Type: text/event-stream`。实际支持的模型、工具调用、思考（reasoning）和采样参数取决于平台渠道配置。

| 请求字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `model` | string | 是 | 当前账号可用模型的准确 ID（取自 `/v1/models`） |
| `messages` | array | 是 | 按对话顺序排列的消息数组 |
| `messages[].role` | string | 是 | 消息角色：`"system"`、`"user"`、`"assistant"`、`"tool"` |
| `messages[].content` | string / 内容块数组 / null | 按消息类型 | 文本或多模态内容；当 `assistant` 仅发出 `tool_calls` 时可为 `null` |
| `messages[].tool_calls` | array | 否 | `assistant` 角色发出的工具调用数组 |
| `messages[].tool_call_id` | string | `role === "tool"` 时必填 | 对应工具调用的 `id` |
| `stream` | boolean | 否 | `true` 请求 SSE 流式返回；显式 `false` 必须保留 |
| `temperature`、`top_p` | number | 否 | 采样参数，仅向支持该参数的模型发送 |
| `max_tokens` / `max_completion_tokens` | integer | 否 | 最大生成 token 数（客户端连通性探测对 `gpt-5`、`o1`、`o3`、`o4` 前缀模型发送 `max_completion_tokens`，其余模型发送 `max_tokens`） |
| `tools`、`tool_choice` | array、string / object | 否 | 工具定义与调用策略，模型支持时使用 |
| `stream_options.include_usage` | boolean | 否 | 流式请求为 `true` 时，在流结束前返回包含 `usage` 的事件帧 |
| `response_format` | object | 否 | 结构化输出格式要求（模型支持时使用） |

### 非流式最小请求与响应

```bash
curl "$PLATFORM_URL/v1/chat/completions" \
  -H "Authorization: Bearer $MODEL_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"model":"MODEL_ID_PLACEHOLDER","messages":[{"role":"user","content":"用一句话介绍你的能力"}],"stream":false}'
```

```json
{
  "id": "chatcmpl-example",
  "object": "chat.completion",
  "created": 1792713600,
  "model": "MODEL_ID_PLACEHOLDER",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "我可以协助你整理信息和完成任务。"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 18,
    "completion_tokens": 15,
    "total_tokens": 33
  }
}
```

| 响应字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | 本次对话补全 ID，便于排查日志 |
| `created` | integer | 响应生成时间戳（**Unix 秒**） |
| `choices[].index` | integer | 候选结果索引，按 `index` 归并，不依赖数组下标顺序 |
| `choices[].message.content` | string / null | 最终回答正文；仅触发工具调用时可能为 `null` 或空字符串 |
| `choices[].message.reasoning_content` | string，可缺失 | 非流式思考内容（仅支持结构化思考的模型返回） |
| `choices[].message.tool_calls` | array，可缺失 | 模型请求执行的工具调用列表 |
| `choices[].finish_reason` | string / null | 结束原因，如 `"stop"`、`"length"`、`"tool_calls"` |
| `usage` | object，可缺失 | `prompt_tokens`、`completion_tokens`、`total_tokens`；仅作客户端统计展示，实际扣费以平台账单为准 |

### 流式请求与响应（SSE）

将 `stream` 设为 `true`，使用标准 SSE 解析器按事件边界（`\n\n`）读取 `data:` 行；注意单个 TCP 网络分块不等同于单个 JSON 事件。

```bash
curl -N "$PLATFORM_URL/v1/chat/completions" \
  -H "Authorization: Bearer $MODEL_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"model":"MODEL_ID_PLACEHOLDER","messages":[{"role":"user","content":"你好"}],"stream":true}'
```

以下为支持结构化思考输出的模型流式事件示例（并非所有模型均返回 `reasoning_content`）：

```text
data: {"id":"chatcmpl-example","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"先理解问题。"},"finish_reason":null}]}

data: {"id":"chatcmpl-example","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"你好！"},"finish_reason":null}]}

data: {"id":"chatcmpl-example","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

| 流式帧 / 字段 | 客户端处理规范 |
|---|---|
| `delta.content` | 增量追加到回答正文区，不得被思考内容覆盖或阻塞到流结束才一次性渲染 |
| `delta.reasoning_content` | 增量追加到独立思考区，不得混入最终回答正文 |
| `delta.tool_calls` | 按 `index` 与工具调用 `id` 增量拼接，注意 `function.arguments` 的 JSON 字符串可能跨多个事件分片 |
| `finish_reason` | 标识当前候选生成的结束原因；不能把“收到思考增量”误判为整个对话已完成 |
| 独立 `usage` 帧 | 流尾可能出现 `choices: []` 但带顶层 `usage` 的事件帧，不能因 `choices` 为空报错或丢弃用量 |
| `data: [DONE]` | SSE 正常结束标记；若连接在收到 `[DONE]` 前异常断开，应保留已渲染的部分内容并标记流中断 |

客户端本地的模型思考能力标记（如适配已知平台模型别名是否支持 reasoning）仅用于界面展示与能力识别，不代表对所有未知模型强行开启思考，也不得伪造思考区内容。

### 模型错误响应

```json
{
  "error": {
    "message": "Example upstream error",
    "type": "upstream_error",
    "code": "example_error"
  }
}
```

`/v1/*` 接口报错时读取顶层 `error` 对象（包含 `message`、`type`、`code`），不使用平台 `/api/*` 的 `success: false` 包装。流开始前先检查 HTTP 状态码；流开始后仍须识别中途错误事件与异常断流。排障日志只记录 HTTP 状态码与安全错误摘要，严禁打印 `MODEL_API_KEY`、完整请求头或用户敏感提示词。
