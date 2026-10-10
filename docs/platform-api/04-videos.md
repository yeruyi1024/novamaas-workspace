# 视频生成

平台全面支持火山方舟官方原生 v3 视频生成规范与网关兼容模式。**强烈推荐客户系统优先接入官方原生 v3 协议栈**，网关兼容模式仅作为存量系统过渡通道（不推荐新业务使用）。

## 首选推荐规范：火山方舟官方原生 v3 协议

> [!TIP]
> **企业对接最佳实践首选建议**：
> 强烈推荐优先接入火山方舟官方原生协议栈（素材库 Action RPC + 视频任务 `POST /api/v3/contents/generations/tasks` + 状态轮询 `GET /api/v3/...`）。
> 
> **核心技术收益**：
> 1. **100% 完整支持官方高级特性**：全面支持豆包 Seedance 2.0 / 2.5 系列模型，支持结构化输入、首尾帧控制、多图参考、音画同步与 Base64 直传。
> 2. **彻底消除真人审核拦截**：素材库签发的 `asset://` 资产句柄直接在原生层毫秒级加载特征缓存，杜绝真人出镜审核被拒风险。
> 3. **火山 TOS 极速 CDN 直链**：结果直接返回火山 TOS 官方极速 CDN 专属直链，下载吞吐带宽提升 3~5 倍且不占用网关中转资源。

## 接口概览与调用链路

```text
1. 创建素材组          ➔  2. 录入人像素材         ➔  3. 提交火山视频任务 ★    ➔  4. 原生轮询状态 ★        ➔  5. 官方 CDN 直连拉取 ★
Action=CreateAssetGroup    Action=CreateAsset         POST /api/v3/.../tasks     GET /api/v3/.../tasks/{id}    GET {video_url}
(AK/SK 增)                 获 asset:// 句柄 (增)      注入 asset:// (Bearer 增)  获取 video_url (查)           免 Header 极速高速下载
```

| 业务场景 | 推荐等级 | 请求方式 | 接口路径 | 鉴权方式 | 核心特点与优势 |
| --- | --- | --- | --- | --- | --- |
| 提交视频生成任务 | ★ 首选推荐 | POST | `/api/v3/contents/generations/tasks` | API Key (Bearer) | 支持首尾帧、素材句柄、多图参考、音效同步 |
| 查询单个任务状态 | ★ 首选推荐 | GET | `/api/v3/contents/generations/tasks/{id}` | API Key (Bearer) | 提取官方 TOS 专属极速直链与 Token 用量 |
| 分页查询任务列表 | ★ 首选推荐 | GET | `/api/v3/contents/generations/tasks` | API Key (Bearer) | 分页查询当前账号所有任务，支持多维过滤 |
| 取消/终止视频任务 | ★ 首选推荐 | DELETE | `/api/v3/contents/generations/tasks/{id}` | API Key (Bearer) | 及时取消正在排队或渲染的任务并释放配额 |
| 视频文件极速下载 | ★ 首选推荐 | GET | `{video_url}` (火山 TOS 直链) | 免鉴权 Header | BGP 骨干带宽直连，支持流式断点续传 |
| 网关兼容视频生成 | 兼容通道（不推荐） | POST | `/v1/video/generations` | API Key (Bearer) | 扁平化参数，仅用于存量系统对接 |
| 网关兼容任务轮询 | 兼容通道（不推荐） | GET | `/v1/video/generations/{task_id}` | API Key (Bearer) | 附带百分比进度（`progress: "100%"`） |

## 一、火山原生视频生成任务全生命周期规范（官方首选）

### 1. 提交视频生成任务 (CreateTask) · 增

采用火山方舟官方原生 v3 标准 REST 协议规范，全面支持豆包 Seedance 2.0 / 2.5 系列模型。使用 `content` 结构化输入，支持首帧控制、尾帧衔接、素材库句柄引用、多图参考、Base64 直传与音频同步生成。

- **接口路径**：`POST /api/v3/contents/generations/tasks`
- **请求头**：
  - `Content-Type: application/json`
  - `Authorization: Bearer YOUR_API_KEY`（在控制台「令牌」页面创建并获取）

#### 请求参数说明

| 参数名 | 类型 | 必选 | 默认值 | 完整说明与业务建议 |
| --- | --- | --- | --- | --- |
| `model` | string | 是 | — | 模型标识：`doubao-seedance-2-0`、`doubao-seedance-2-0-fast-260128` 等 |
| `content` | array | 是 | — | 结构化内容列表，由文本提示词与图片对象数组构成（官方标准规范，无平铺限制） |
| `content[].type` | string | 是 | — | 内容类型：`"text"`（文本描述）或 `"image_url"`（图像来源） |
| `content[].text` | string | 条件 | — | 文本提示词，描述画面主体、动作动态、环境运镜、质感细节等（中英文均可） |
| `content[].image_url.url` | string | 条件 | — | 支持三种格式：<br>1. **素材库句柄 (强烈推荐)**：`asset://{asset_id}`（100% 避免真人拦截）<br>2. **公网直链**：`https://...`（HTTP/HTTPS 直连直链）<br>3. **Base64 编码**：`data:image/jpeg;base64,{data}`（单图 < 30MB） |
| `content[].role` | string | 否 | `first_frame` | 图片角色定位：<br>• `"first_frame"`：首帧图（从该图开始动态演化）<br>• `"last_frame"`：尾帧图（向该图演化收尾）<br>• `"reference_image"`：人像/风格特征参考图（保持多图一致性） |
| `resolution` | string | 否 | `720p` | 输出物理分辨率：`480p`、`720p`（默认标清）、`1080p`（高清）、`4k` |
| `ratio` | string | 否 | `16:9` | 画面宽高比：`16:9`、`9:16`、`1:1`、`4:3`、`3:4`、21:9、`adaptive`（自适应首图） |
| `duration` | integer | 否 | 5 | 视频时长（秒），支持 `5` 或 `10`（生成更加从容丰富的连续动作） |
| `generate_audio` | boolean | 否 | true | 是否同步生成与画面动态匹配的高保真音效/环境声/音乐（默认开启） |
| `watermark` | boolean | 否 | false | 右下角是否保留官方“AI生成”水印，默认 false（无水印纯净版） |

#### (1) 官方首选方案：素材库句柄调用示例 (多图人像一致性参考 · 杜绝真人拦截)

```bash
curl --fail-with-body --request POST 'https://gateway.ai.shilijia.xyz/api/v3/contents/generations/tasks' \
  --header 'Authorization: Bearer YOUR_API_KEY' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "model": "doubao-seedance-2-0",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "generate_audio": true,
  "watermark": false,
  "content": [
    {
      "type": "text",
      "text": "严格参考输入图片中的人物形象生成视频，保持人物外貌、脸型、发型和神态高度一致，自然微笑，动作流畅，在科技办公室中专注工作，电影级质感。"
    },
    {
      "type": "image_url",
      "role": "reference_image",
      "image_url": {
        "url": "asset://YOUR_ASSET_ID_1"
      }
    },
    {
      "type": "image_url",
      "role": "reference_image",
      "image_url": {
        "url": "asset://YOUR_ASSET_ID_2"
      }
    }
  ]
}'
```

提交成功响应 (HTTP 200)：

```json
{
  "id": "cgt-20260924081530-ab123",
  "model": "doubao-seedance-2-0",
  "status": "queued",
  "created_at": 1790237730
}
```

保存返回的 `id`，用于后续查询渲染状态。

#### (2) 公网直链调用示例

适合已有稳定公网对象存储直链的场景：

```json
{
  "model": "doubao-seedance-2-0",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "generate_audio": true,
  "content": [
    {
      "type": "text",
      "text": "古风人物在庭院中优雅挥袖，落叶飘落，高清画质。"
    },
    {
      "type": "image_url",
      "role": "first_frame",
      "image_url": {
        "url": "https://tos-cn-beijing.volces.com/assets/scene.jpg"
      }
    }
  ]
}
```

#### (3) 本地 Base64 直传调用示例

适合客户端本地图片直接提交（单图限制小于 30MB）：

```json
{
  "model": "doubao-seedance-2-0",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "content": [
    {
      "type": "text",
      "text": "高科技发布会演讲，动作自然，写实风格。"
    },
    {
      "type": "image_url",
      "role": "first_frame",
      "image_url": {
        "url": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
      }
    }
  ]
}
```

---

### 2. 查询单个任务状态与渲染结果 (GetTask) · 查

- **接口路径**：`GET /api/v3/contents/generations/tasks/{id}`
- **请求头**：`Authorization: Bearer YOUR_API_KEY`

```bash
curl --fail-with-body --request GET 'https://gateway.ai.shilijia.xyz/api/v3/contents/generations/tasks/cgt-20260924081530-ab123' \
  --header 'Authorization: Bearer YOUR_API_KEY'
```

#### 成功响应结构 (HTTP 200 · status == "succeeded" 即渲染完成)

```json
{
  "id": "cgt-20260924081530-ab123",
  "model": "doubao-seedance-2-0",
  "status": "succeeded",
  "content": {
    "video_url": "https://ark-acg-cn-beijing.tos-cn-beijing.volces.com/doubao-seedance-2-0/video_xxx.mp4?X-Tos-Algorithm=...",
    "cover_url": "https://ark-acg-cn-beijing.tos-cn-beijing.volces.com/doubao-seedance-2-0/cover_xxx.jpg?X-Tos-Algorithm=..."
  },
  "usage": {
    "completion_tokens": 125,
    "total_tokens": 125
  },
  "created_at": 1790237730,
  "updated_at": 1790237765
}
```

#### 任务状态流转说明

| 状态值 (`status`) | 说明 | 业务下一步建议 |
| --- | --- | --- |
| `queued` | 排队中 | 建议间隔 3~5 秒继续轮询 |
| `running` | 渲染处理中 | 建议间隔 3~5 秒继续轮询 |
| `succeeded` | 渲染成功 | 提取 `content.video_url`，立即下载或流式转存 |
| `failed` | 任务失败 | 检查 `error` 字段说明，停止轮询 |

---

### 3. 分页查询视频任务列表 (ListTasks) · 查

- **接口路径**：`GET /api/v3/contents/generations/tasks`
- **请求参数 (Query)**：
  - `page_num`：页码，从 1 开始，最大支持 500（默认 1）
  - `page_size`：每页条数，默认 10，最大支持 100
  - `filter.model`：按模型标识精确筛选（如 `doubao-seedance-2-0`）
  - `filter.status`：按状态筛选（`queued` / `running` / `succeeded` / `failed`）

```bash
curl --fail-with-body --request GET 'https://gateway.ai.shilijia.xyz/api/v3/contents/generations/tasks?page_num=1&page_size=10&filter.status=succeeded' \
  --header 'Authorization: Bearer YOUR_API_KEY'
```

响应示例：

```json
{
  "items": [
    {
      "id": "cgt-20260924081530-ab123",
      "model": "doubao-seedance-2-0",
      "status": "succeeded",
      "content": {
        "video_url": "https://ark-acg-cn-beijing.tos-cn-beijing.volces.com/doubao-seedance-2-0/video_xxx.mp4?X-Tos-..."
      },
      "created_at": 1790237730
    }
  ],
  "total": 1
}
```

---

### 4. 取消/终止视频任务 (CancelTask) · 删

用于在任务排队或渲染过程中及时取消生成。取消成功后将释放扣减的配额，并返回任务当前状态快照：

- **接口路径**：`DELETE /api/v3/contents/generations/tasks/{id}`
- **请求头**：`Authorization: Bearer YOUR_API_KEY`

```bash
curl --fail-with-body --request DELETE 'https://gateway.ai.shilijia.xyz/api/v3/contents/generations/tasks/cgt-20260924081530-ab123' \
  --header 'Authorization: Bearer YOUR_API_KEY'
```

响应示例 (HTTP 200)：

```json
{
  "id": "cgt-20260924081530-ab123",
  "model": "doubao-seedance-2-0",
  "status": "failed",
  "error": {
    "code": "cancelled",
    "message": "task was cancelled by user"
  }
}
```

---

## 二、火山 TOS CDN 极速下载与转存最佳实践

1. **直连官方高速骨干 CDN**：
   响应中的 `video_url` 是火山官方专属对象存储经过全球多线 BGP CDN 加速的直签名直链，吞吐带宽大、稳定无二次代理衰减。
2. **免 Header 鉴权拉取**：
   该直链为已签名的安全预认证 URL，**下载时严禁携带 Authorization 请求头**。
3. **支持流式断点续传**：
   大文件下载推荐使用断点续传命令直接保存：
   ```bash
   curl -L -C - -o "video.mp4" "YOUR_VIDEO_URL"
   ```
4. **有效期提示**：
   官方直签名有效访问期为 **24 小时**（`X-Tos-Expires=86400`），请在生成后及时流式转存至业务自有对象存储。

---

## 三、网关兼容模式接口（备选兼容通道 · 不推荐）

> [!WARNING]
> **兼容通道声明（不推荐作为新系统接入方式）**：
> 为兼顾已采用 OpenAI 视频风格接入或需要轻量级扁平参数调用的存量客户端，网关额外提供了兼容模式接口。两套协议完全共用同一个素材库与任务底层。但兼容模式在高级生成特性控制、首尾帧灵活演化及吞吐优化上存在局限，**仅作为存量兼容过渡，强烈建议新系统优先采用上方的原生 v3 接口**。

### 1. 网关兼容视频生成接口 (POST /v1/video/generations)

支持将素材库 `asset://` 句柄置于 `metadata.content` 列表中传入：

```bash
curl --fail-with-body --request POST 'https://gateway.ai.shilijia.xyz/v1/video/generations' \
  --header 'Authorization: Bearer YOUR_API_KEY' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "model": "doubao-seedance-2-0",
  "prompt": "商务男士在明亮办公室内演讲，动作自然生动，电影级画质",
  "metadata": {
    "duration": 5,
    "ratio": "16:9",
    "resolution": "720p",
    "generate_audio": false,
    "watermark": false,
    "content": [
      {
        "type": "text",
        "text": "商务男士在明亮办公室内演讲，动作自然生动，电影级画质"
      },
      {
        "type": "image_url",
        "role": "reference_image",
        "image_url": {
          "url": "asset://YOUR_ASSET_ID"
        }
      }
    ]
  }
}'
```

提交成功响应：

```json
{
  "id": "task_TjyIFnCBCqj4UgdIkpP7mHFnZDbqZkx4",
  "task_id": "task_TjyIFnCBCqj4UgdIkpP7mHFnZDbqZkx4",
  "object": "video",
  "model": "doubao-seedance-2-0",
  "status": "queued",
  "progress": 0
}
```

#### 阿里百炼 Wan3.0 请求参数

以下兼容适用于百炼渠道的 `wan3.0-video`、`wan3.0-video-prime`，包括映射到这些模型的别名。不同百炼渠道均适用；其他渠道类型和百炼旧模型继续使用原有规范。

现有请求无需调整。接口仍为 `POST /v1/video/generations`（也支持 `/v1/videos`），请求头使用 `Content-Type: application/json` 和 `Authorization: Bearer YOUR_API_KEY`。

```json
{
  "model": "wan3.0-video",
  "prompt": "金毛小狗在阳光明媚的草地上奔跑，电影级画面",
  "image": "https://example.com/sample.png",
  "audio": false,
  "resolution": "720P",
  "ratio": "16:9",
  "duration": 5,
  "prompt_extend": true,
  "watermark": false,
  "seed": 0
}
```

将图片地址替换为实际地址；文生视频可省略 `image`。官方 `input/parameters` 和已有的 `metadata.input/metadata.parameters` 写法同样支持，无需迁移参数位置。响应和任务查询继续使用网关视频格式。

| 参数 | 支持范围 |
| --- | --- |
| `audio` | 布尔值；兼容字符串 `"true"` / `"false"`，`false` 请求关闭音轨 |
| `resolution` | `480P`、`720P`、`1080P`，兼容小写；网关默认 `720P` |
| `ratio` | `adaptive`、`21:9`、`16:9`、`4:3`、`1:1`、`3:4`、`9:16` |
| `size` | 原有分辨率或预设像素尺寸，如 `720P`、`1280*720`、`1280x720`，服务端转换为官方字段 |
| `duration` | 2～30 秒，默认 5 秒；暂不支持 `-1` 智能时长，参考视频还需符合上游总时长限制 |
| `prompt_extend` / `watermark` | 布尔值，兼容字符串布尔值；智能改写默认开启 |
| `seed` | `-1` 或 `0`～`2147483647`；不指定时由上游随机生成 |

明确传入的 `false`、`0` 会保留。重复参数按 `metadata.parameters > parameters > 顶层字段` 逐字段合并；显式 `resolution/ratio` 优先于从 `size` 推导的值。`metadata.input` 的同名字段优先于顶层 `input`；明确给出的 `input.media` 优先于从旧图片字段生成的媒体列表。建议同一参数只写一次。

Wan3 渠道的参数覆盖在请求格式转换和参数校验前执行。条件可用 `original_model` 匹配用户提交的模型名，例如精确匹配 `wan3.0-video-480p` 后设置顶层 `size` 为 `480p`，随后转换为上游 `parameters.resolution: "480P"`。模型映射独立负责把公开模型名替换成上游模型名，模型后缀不会自动决定分辨率。`mode: "full"` 要求名称完全一致。

参数覆盖针对用户请求结构执行，仍遵循上述字段优先级：覆盖顶层 `size` 不会压过请求中显式的 `resolution`；需要强制覆盖时，应覆盖实际使用的 `parameters.resolution` 或 `metadata.parameters.resolution`。覆盖后的时长会重新校验，并用于计费估算和上游请求。原始请求快照保留用户提交的内容。

百炼旧模型（例如 Wan2.5）关闭音频仍使用 `metadata.parameters.audio: false`，类型及支持范围保持原有适配器行为。媒体类型和素材限制见[Wan3.0 官方文档](https://help.aliyun.com/zh/model-studio/wan3-video-generation-api-reference)。

### 2. 网关兼容视频任务轮询 (GET /v1/video/generations/{task_id})

提供实时百分比进度（`progress`）的扩展状态展示：

```bash
curl --fail-with-body --request GET 'https://gateway.ai.shilijia.xyz/v1/video/generations/task_TjyIFnCBCqj4UgdIkpP7mHFnZDbqZkx4' \
  --header 'Authorization: Bearer YOUR_API_KEY'
```

成功响应 (已完成 · 100%)：

```json
{
  "code": "success",
  "data": {
    "task_id": "task_TjyIFnCBCqj4UgdIkpP7mHFnZDbqZkx4",
    "status": "SUCCESS",
    "progress": "100%",
    "result_url": "https://gateway.ai.shilijia.xyz/output/video/task_xxx.mp4"
  }
}
```
