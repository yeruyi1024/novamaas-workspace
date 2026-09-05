# 火山方舟原生 API 渠道

本功能选择性同步自上游 [QuantumNous/new-api#6653](https://github.com/QuantumNous/new-api/pull/6653)，对应上游提交 `362c9d666ab4dddbe789cba2c1acc4217573d6a5`。它新增独立的 **Volc Native（渠道类型 61）**，用于转发火山方舟原生 `/api/v3` 图片生成和异步视频任务接口。

## 选择哪个渠道

调用下列原生接口时，应新建并选择 **Volc Native**：

- `POST /api/v3/images/generations`
- `POST /api/v3/contents/generations/tasks`
- `GET /api/v3/contents/generations/tasks`
- `GET /api/v3/contents/generations/tasks/{task_id}`
- `DELETE /api/v3/contents/generations/tasks/{task_id}`

`DoubaoVideo` 继续服务项目已有的兼容视频接口；`VolcEngine` 继续服务兼容接口。原生 `/api/v3` 路由只会选择 Volc Native，Volc Native 也不会参与 `/v1` 兼容路由的渠道选择。

## 管理端配置

在“渠道”中新增 Volc Native，并配置：

- Base URL：`https://ark.cn-beijing.volces.com`
- Key：火山方舟 API Key；前端不需要填写 `Bearer ` 前缀
- 模型：填写火山方舟实际的上游模型 ID，例如已开通的 Seedance 或 Seedream 模型 ID

该渠道按原生语义透传请求，所以不支持模型映射和参数覆盖。若客户端模型名与上游模型 ID 不同，应由客户端直接改为上游模型 ID。

## 客户端调用

客户端仍使用本平台签发的访问令牌：

```bash
curl "$NOVAMAAS_BASE_URL/api/v3/contents/generations/tasks" \
  -H "Authorization: Bearer $NOVAMAAS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "doubao-seedance-2-0-260128",
    "content": [
      {"type": "text", "text": "一艘飞船穿过蓝色星云"}
    ],
    "resolution": "1080p",
    "duration": 5,
    "watermark": false
  }'
```

请求字段直接放在 JSON 顶层，**不需要增加 `metadata` 包装层**。服务端仅在内部读取顶层字段完成校验和计费估算，发往火山方舟的 JSON 字节保持原样。

创建成功后，平台会返回自己的公开任务 ID，例如 `task_xxx`。查询和取消必须继续使用这个公开 ID：

```bash
curl "$NOVAMAAS_BASE_URL/api/v3/contents/generations/tasks/task_xxx" \
  -H "Authorization: Bearer $NOVAMAAS_TOKEN"

curl -X DELETE \
  "$NOVAMAAS_BASE_URL/api/v3/contents/generations/tasks/task_xxx" \
  -H "Authorization: Bearer $NOVAMAAS_TOKEN"
```

真实的火山方舟任务 ID 和任务提交时使用的上游 Key 只保存在服务端私有数据中，不会返回给客户端。查询、列表和取消还会检查任务所属用户，以及令牌配置的指定渠道和可用模型范围。

## 行为边界

- 视频任务状态依赖后台任务轮询；部署时应保持任务轮询服务正常运行。
- 列表接口返回当前用户经令牌权限过滤后的本地任务，不会枚举共享火山账号下由其他用户或其他系统直接创建的任务。
- 图片接口当前按一次请求配置模型价格；视频任务沿用项目的异步任务计费流程和已有的 Seedance 视频输入、分辨率倍率。
- 原生图片响应会完整读取后再返回；即使上游接受流式参数，当前实现也不会向客户端逐段实时推送。
- 上游错误响应会保留 HTTP 状态码和响应体；本地校验、权限或计费错误使用项目生成的错误结构。
