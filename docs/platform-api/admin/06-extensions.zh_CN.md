# 星途获取插件与技能

星途客户端以当前登录账号访问 MaaS 平台，拉取连接器展示策略、公司技能目录与公司 MCP 目录，并保持**目录推荐、本机安装状态、送达/通知回执**三者独立。平台发布新目录不等于自动静默安装到用户电脑。

## 客户端读取与回执接口索引

| 方法 | 路径 | 认证方式 | 协议版本与用途 |
|---|---|---|---|
| GET | `/api/apps/starpath/config` | `Authorization: Bearer $ACCESS_TOKEN` | `schema_version: 1`；按客户端版本与系统架构拉取内置连接器展示策略 |
| GET | `/api/apps/starpath/extensions/mcp` | `Authorization: Bearer $ACCESS_TOKEN` | `schema_version: 2`；拉取分发给当前账号的公司 MCP 目录与统一停用名单 |
| GET | `/api/apps/starpath/v3/skills/catalog` | `Authorization: Bearer $ACCESS_TOKEN` | `schema_version: 3`；拉取分发给当前账号的公司技能目录（含 `hosted_artifact` 托管制品版本） |
| GET | `/api/apps/starpath/extensions/skills` | `Authorization: Bearer $ACCESS_TOKEN` | `schema_version: 2`；旧版兼容技能目录接口（会将 `hosted_artifact` 降级为只读 `metadata`） |
| POST | `/api/apps/starpath/extensions/:module/receipts` | `Authorization: Bearer $ACCESS_TOKEN` | 上报 `skills` 或 `mcp` 目录的送达回执（`catalog_received`）或新增通知展示回执（`notification_shown`） |

> **客户端运行环境开关说明**：在 `split/pr3-extensions` 客户端实现中，连接器策略同步需启用环境变量 `STARPATH_MAAS_MANAGED_PLUGINS=draft-v1`；公司扩展目录同步需启用 `STARPATH_MAAS_EXTENSIONS=v2` 或 `STARPATH_MAAS_SKILL_ARTIFACTS=v3`（启用 `v3` 时技能目录自动请求 `/api/apps/starpath/v3/skills/catalog`）。

## GET /api/apps/starpath/config（连接器展示策略）

无请求体，通过 URL 查询参数传递客户端兼容性信息。全部 4 个查询参数均为必填，校验不通过时服务端返回 `400 UNSUPPORTED_PLUGIN_COMPATIBILITY`。

| 查询参数 | 类型 | 必填 | 格式与枚举约束 |
|---|---|---|---|
| `client_version` | string | 是 | 星途客户端语义化版本号（长度 `<= 100`，如 `1.0.0`） |
| `host_version` | string | 是 | 宿主运行时语义化版本号（长度 `<= 100`，如 `2026.8.1`） |
| `os` | string | 是 | 操作系统枚举：`"darwin"`、`"win32"` 或 `"linux"` |
| `arch` | string | 是 | 处理器架构枚举：`"arm64"` 或 `"x64"` |

```bash
curl "$PLATFORM_URL/api/apps/starpath/config?client_version=1.0.0&host_version=2026.8.1&os=darwin&arch=arm64" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "schema_version": 1,
    "config_revision": 2,
    "published_at": 1792713600,
    "refresh_interval_seconds": 60,
    "plugins": [
      {
        "plugin_id": "openclaw-weixin",
        "runtime_id": "openclaw-weixin",
        "package_name": "@tencent-weixin/openclaw-weixin",
        "name": "微信连接器",
        "description": "随星途客户端内置的微信连接器展示与远程连接入口。",
        "management_type": "bundled",
        "placement": "remote_connections",
        "listed": true,
        "visible": true,
        "enabled": true,
        "sort_order": 500,
        "required_feature_ids": [],
        "target_release_id": null,
        "install_paused": false,
        "upgrade_deadline": null
      }
    ]
  }
}
```

客户端对返回的 `config_revision` 做单调递增校验：若小于本地已缓存版本则拒绝回退（标记 `staleRevision`）；若同一 `config_revision` 内容不一致则拒绝（标记 `revisionConflict`）。条目从目录中隐藏或移除不会触发客户端卸载内置连接器。

## GET /api/apps/starpath/extensions/:module 与 GET /api/apps/starpath/v3/skills/catalog

使用登录访问令牌，无请求体。服务端根据当前登录用户的 `user.id` 自动过滤 `distribution`（仅返回 `mode === "all"` 或 `user_ids` 包含当前用户的已发布条目，且返回给客户端的条目中**剥离 `distribution` 字段**，防止泄漏其他目标用户 ID）。

```bash
curl "$PLATFORM_URL/api/apps/starpath/extensions/mcp" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

已发布且对当前账号可见的 MCP 目录完整响应示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "schema_version": 2,
    "module": "mcp",
    "publication_id": "mcp:3",
    "delivery_id": "mcp:3:42",
    "published_revision": 3,
    "published_at": 1792713600,
    "refresh_interval_seconds": 60,
    "items": [
      {
        "id": "search-example",
        "name": "示例检索",
        "description": "公司检索工具",
        "version": "1.0.0",
        "source_ref": "company:search-example",
        "listed": true,
        "visible": true,
        "sort_order": 0,
        "skill": null,
        "mcp": {
          "transport": "http",
          "command": "",
          "args": [],
          "url": "https://mcp.example.invalid/search",
          "required_config": []
        },
        "disabled": false
      }
    ],
    "disabled_sources": [],
    "notification": {
      "id": "mcp:3:42",
      "kind": "catalog_added",
      "count": 1
    }
  }
}
```

| `data` 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `schema_version` | integer | 是 | `/extensions/:module` 固定为 `2`；`/v3/skills/catalog` 固定为 `3` |
| `module` | string | 是 | `"skills"` 或 `"mcp"` |
| `publication_id` | string / null | 是 | 当前发布版本标识（如 `"mcp:3"`）；模块尚未发布过任何版本时为 `null` |
| `delivery_id` | string / null | 是 | 当前账号在本次发布下的专属分发标识（格式 `<module>:<revision>:<user_id>`，如 `"mcp:3:42"`）；尚未发布时为 `null` |
| `published_revision` | integer | 是 | 已发布修订号（非负整数；尚未发布时为 `0`） |
| `published_at` | integer | 是 | 发布时间戳（**Unix 秒**；尚未发布时为 `0`） |
| `refresh_interval_seconds` | integer | 是 | 平台建议的轮询刷新间隔（**秒**，当前固定为 `60`；客户端解析要求在 `15`–`3600` 秒范围内） |
| `items` | array | 是 | 当前账号可见且未停用的已发布条目数组（当 MCP 条目被设为 `disabled: true` 时，该条目从 `items` 中移除并转入 `disabled_sources`） |
| `disabled_sources` | string[] | 仅 `mcp` 返回 | 当前账号需要被强制停用的公司 MCP `source_ref` 列表（去重排序；客户端同步后自动关闭匹配的本机已装公司 MCP） |
| `notification` | object / null | 是 | 当本次发布相比上一版对当前账号有新增可见且 `listed && visible` 的条目，且当前账号尚未上报 `notification_shown` 时返回通知对象；否则为 `null` |
| `notification.id` | string | 通知存在时必填 | 通知唯一标识，必须与 `delivery_id` 一致 |
| `notification.kind` | string | 通知存在时必填 | 固定为 `"catalog_added"` |
| `notification.count` | integer | 通知存在时必填 | 本次对当前账号新增的可展示条目数量（`1`–`200`） |

尚未发布任何版本属于正常初始状态（非服务报错），服务端返回 HTTP 200 及如下合法空目录：

```json
{
  "success": true,
  "message": "",
  "data": {
    "schema_version": 2,
    "module": "mcp",
    "publication_id": null,
    "delivery_id": null,
    "published_revision": 0,
    "published_at": 0,
    "refresh_interval_seconds": 60,
    "items": [],
    "disabled_sources": [],
    "notification": null
  }
}
```

## 技能描述对象（`item.skill`）

目录条目的公共字段（`id`、`name`、`description`、`version`、`source_ref`、`listed`、`visible`、`sort_order`）见上一章“后台配置与发布”。技能条目的 `mcp` 固定为 `null`，`skill` 对象包含以下字段：

| `skill` 字段 | 类型 | 必填 | 枚举与约束说明 |
|---|---|---|---|
| `kind` | string | 是 | 技能分发类型枚举：`"bundled"`（随客户端安装包内置）、`"metadata"`（仅展示推荐元数据，无下载包）、`"https_archive"`（外部固定 HTTPS ZIP 包）、`"hosted_artifact"`（仅 v3 支持，平台对象存储托管的已确认制品包） |
| `artifact_url` | string | 是 | 当 `kind === "https_archive"` 时为合法的 `https://` 下载地址；当 `kind` 为 `"bundled"`、`"metadata"` 或 `"hosted_artifact"` 时必须为空字符串 `""` |
| `sha256` | string | 是 | 当 `kind` 为 `"https_archive"` 或 `"hosted_artifact"` 时为 ZIP 文件的 64 位小写十六进制 SHA-256 摘要；当 `kind` 为 `"bundled"` 或 `"metadata"` 时必须为空字符串 `""` |
| `dependencies` | string[] | 是 | 依赖标识数组（0–30 项）；注意客户端托管技能安装器当前对非空 `dependencies` 会拒绝直接安装并上报 `DEPENDENCY_REQUIRED` |
| `artifact_release_id` | string | v3 托管技能必填 | 当 `kind === "hosted_artifact"` 时为平台已确认（`status === "ready"`）的制品版本 ID（如 `"srel_..."`）；其他类型为空字符串或省略 |

**v2 与 v3 技能目录兼容行为**：如果旧客户端请求 v2 接口 `GET /api/apps/starpath/extensions/skills`，服务端会自动把目录中的 `hosted_artifact` 条目降级为 `kind: "metadata"`，并清空 `artifact_url`、`sha256` 与 `artifact_release_id`。客户端看到 `kind: "metadata"` 时只能展示说明，严禁尝试当作 ZIP 下载安装。

## POST /api/apps/starpath/extensions/:module/receipts

请求体上限 4 KiB，严格只接收 `delivery_id` 与 `status` 两个字段，使用当前登录账号的 `Authorization: Bearer $ACCESS_TOKEN`。

| 请求字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `delivery_id` | string | 是 | 原样传入目录响应中的 `data.delivery_id`（1–128 字符，服务端校验其中的模块名、发布修订号与当前登录 `user.id` 必须完全一致） |
| `status` | string | 是 | 回执阶段枚举：`"catalog_received"`（客户端已拉取并解析该发布目录）或 `"notification_shown"`（客户端已向用户展示新增通知） |

```bash
curl -X POST "$PLATFORM_URL/api/apps/starpath/extensions/mcp/receipts" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"delivery_id":"mcp:3:42","status":"catalog_received"}'
```

成功响应示例（HTTP 200）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "delivery_id": "mcp:3:42",
    "catalog_received_at": 1792713660,
    "notification_shown_at": 0
  }
}
```

| 响应字段 | 类型 | 单位与说明 |
|---|---|---|
| `data.delivery_id` | string | 本次回执对应的分发标识 |
| `data.catalog_received_at` | integer | 服务端首次记录收到目录的时间戳（**Unix 秒**） |
| `data.notification_shown_at` | integer | 服务端首次记录已展示通知的时间戳（**Unix 秒**；尚未上报 `notification_shown` 时为 `0`；上报 `notification_shown` 时若此前未记录 `catalog_received_at`，服务端会自动一并补齐） |

### 回执错误码与处理规则

| 平台错误码 `code` | HTTP 状态 | 触发原因与客户端处理规则 |
|---|---|---|
| `INVALID_EXTENSION_RECEIPT` | `400` | 请求体含未知字段、字段为空、`status` 枚举非法或 `delivery_id` 格式错误；停止重试该条非法回执 |
| `EXTENSION_NOTIFICATION_NOT_AVAILABLE` | `400` | 该发布版本对当前账号没有可展示的新增通知（`notification === null`），客户端却上报了 `notification_shown`；禁止对无通知的目录伪造通知回执 |
| `EXTENSION_RECEIPT_SCOPE_DENIED` | `403` | `delivery_id` 中的模块名或用户 ID 与当前登录账号不匹配；说明账号已切换，立即丢弃旧上下文回执 |
| `EXTENSION_PUBLICATION_NOT_FOUND` | `404` | `delivery_id` 指向的发布修订号不存在；重新拉取最新目录 |
| `EXTENSION_RECEIPT_UNAVAILABLE` | `500` | 服务端数据库写入异常；保留在客户端本地回执账本（`company-extensions-deliveries-v1.json`），待下一次目录刷新时恢复补报 |

**重要区分**：目录回执（`receipts`）只证明客户端“已收到推荐列表”或“已弹出新增提示”，**不代表用户已经安装该技能或启用该 MCP**。托管技能的真实下载、安装、失败与回滚结果通过第七章的独立接口 `POST /api/apps/starpath/v3/skills/installations` 上报。

## 客户端刷新、缓存与账号隔离

1. 客户端按 `origin + accountId + module` 独立维护内存状态、本地磁盘缓存（`company-extensions-v2/v3-<module>.json`）和投递回执账本（`company-extensions-deliveries-v1.json`）。
2. 离线启动时优先加载本地缓存并标记 `status: "cached"`；同步成功后更新为 `status: "current"`。
3. 退出登录、切换账号或销毁服务时，所有在途的目录拉取和回执上报即使随后返回，也会因会话代数（`sequence` / `epoch`）不一致被直接丢弃，严禁跨账号写回。
