# 后台配置与发布

本章面向 MaaS 后台开发与管理员：说明如何授权普通管理员维护三个模块（连接器、技能、MCP），以及后台如何保存草稿、发布目录与查询回执。星途客户端读取发布结果的接口见下一章“星途获取插件与技能”。

本章所有接口均以 `/api/apps/starpath` 为公共前缀，使用 `Authorization: Bearer $ACCESS_TOKEN` 认证。模块目录管理需要具备对应模块权限；给管理员赋权、全局连接器策略（`/admin/config`）与技能对象存储配置（`/admin/skill-artifacts/storage`）仅限 Root（`role = 100`）。

## 模块权限与管理员授权接口

MaaS 将星途扩展管理划分为三个独立权限模块：`starpath_connectors`（连接器）、`starpath_skills`（技能）、`starpath_mcp`（MCP）。Root 拥有全部权限且可授权给启用的普通管理员（`role = 10`）；普通管理员不能跨模块越权，也不能给他人赋权。

| 方法 | 相对路径（前缀 `/api/apps/starpath`） | 所需角色 / 权限 | 用途 |
|---|---|---|---|
| GET | `/extensions/permissions` | 任意登录用户（`UserAuth`） | 查询当前登录账号在三个模块上的 `read`/`manage` 权限及 `can_grant` 标识 |
| GET | `/admin/extensions/permissions/:user_id` | 仅 Root（`RootAuth`） | 查询指定管理员用户（`:user_id`）的三个模块权限 |
| PUT | `/admin/extensions/permissions/:user_id` | 仅 Root（`RootAuth`） | 调整指定普通管理员（`role = 10` 且账号启用）的三个模块权限 |

### GET /api/apps/starpath/extensions/permissions

无请求体。返回当前账号对三个模块的权限快照（当 `manage: true` 时服务端自动赋予 `read: true`）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "permissions": {
      "starpath_connectors": {
        "read": true,
        "manage": true
      },
      "starpath_skills": {
        "read": true,
        "manage": false
      },
      "starpath_mcp": {
        "read": false,
        "manage": false
      }
    },
    "can_grant": false
  }
}
```

### PUT /api/apps/starpath/admin/extensions/permissions/:user_id

请求体上限 1 MiB，严格要求只包含 `reason` 与 `permissions` 两个顶层键，且 `permissions` 必须完整包含三个模块键（每个模块对象只允许包含布尔字段 `read` 与 `manage`）。目标 `:user_id` 必须是处于启用状态的普通管理员（`role = 10`），否则返回 `400 INVALID_EXTENSION_ADMIN`。

```json
{
  "reason": "授权运维管理员维护技能与 MCP 目录",
  "permissions": {
    "starpath_connectors": {
      "read": true,
      "manage": false
    },
    "starpath_skills": {
      "read": true,
      "manage": true
    },
    "starpath_mcp": {
      "read": true,
      "manage": true
    }
  }
}
```

## 连接器、技能与 MCP 后台管理接口索引

| 方法 | 相对路径（前缀 `/api/apps/starpath`） | 所需权限 | 说明 |
|---|---|---|---|
| GET | `/admin/extensions/connectors` | `starpath_connectors:read` | 读取连接器模块当前发布配置（`schema_version: 2`，含 5 个内置连接器条目） |
| PUT | `/admin/extensions/connectors` | `starpath_connectors:manage` | 直接发布连接器展示配置（无独立草稿步骤） |
| GET | `/admin/extensions/connectors/history` | `starpath_connectors:read` | 查询连接器模块最近 50 条发布历史（`data.items`） |
| GET | `/admin/config` | 仅 Root（`RootAuth`） | 读取全局连接器底层策略（`schema_version: 1`）及 `scope_issues` 诊断结果 |
| PUT | `/admin/config` | 仅 Root（`RootAuth`） | 发布全局连接器底层策略（请求体含 `expected_revision`、`reason`、`plugins`） |
| GET | `/admin/config/history` | 仅 Root（`RootAuth`） | 查询全局连接器策略最近 50 条变更历史 |
| GET | `/admin/extensions/:module` | 对应模块 `read` | 读取 `skills`（v2）或 `mcp` 的当前草稿、已发布版本及条目 |
| PUT | `/admin/extensions/:module/draft` | 对应模块 `manage` | 保存 `skills`（v2）或 `mcp` 的完整草稿目录 |
| POST | `/admin/extensions/:module/publish` | 对应模块 `manage` | 将 `skills`（v2）或 `mcp` 当前草稿发布为正式版本 |
| GET | `/admin/extensions/:module/history` | 对应模块 `read` | 查询 `skills` 或 `mcp` 最近 50 条发布历史（`data.items`） |
| GET | `/admin/extensions/:module/receipts?published_revision=3` | 对应模块 `read` | 查询指定发布修订号（必须显式传入正整数 published_revision，无默认值）的客户端送达与通知回执列表 |

> **v2 与 v3 技能管理入口区分**：上表中 `:module` 取 `skills` 或 `mcp`。当技能目录包含上传/导入的托管技能制品（`kind: "hosted_artifact"`）时，后台必须使用第七章的 v3 制品目录接口（`/admin/skill-artifacts/catalog`、`/catalog/draft`、`/catalog/publish`）；若对含托管技能的目录调用旧 v2 接口 `/admin/extensions/skills/draft` 或 `/publish`，服务端将返回 `409 SKILL_ARTIFACT_CLIENT_UPGRADE_REQUIRED`。

## 技能与 MCP 草稿及发布契约

### GET /api/apps/starpath/admin/extensions/mcp

无请求体。返回当前模块的草稿修订号、已发布修订号及两份完整条目列表：

| `data` 字段 | 类型 | 说明 |
|---|---|---|
| `schema_version` | integer | 目录协议版本（v2 接口为 `2`，`/admin/skill-artifacts/catalog` 为 `3`） |
| `module` | string | `"mcp"` 或 `"skills"` |
| `draft_revision` | integer | 当前草稿修订号（初始为 `0`），保存草稿或发布时作为 `expected_revision` 传入 |
| `published_revision` | integer | 当前已发布修订号（尚未发布时为 `0`） |
| `published_at` | integer | 最近一次发布时间戳（**Unix 秒**；尚未发布时为 `0`） |
| `draft.items` | array | 当前草稿完整条目数组 |
| `published.items` | array | 当前已生效的发布条目数组 |

### PUT /api/apps/starpath/admin/extensions/mcp/draft

请求体上限 1 MiB。**全量替换保存当前草稿**（不是单项增量 PATCH；传入 `items: []` 表示清空草稿目录，提交前须核对）。

| 请求字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `expected_revision` | integer | 是 | 乐观并发锁，必须等于最新读取的 `draft_revision`（非负安全整数） |
| `reason` | string | 是 | 本次保存原因（去首尾空白后 1–500 个字符），记入审计日志 |
| `items` | array | 是 | 完整目录条目数组，最多 200 项；同一目录内 `id`（不区分大小写）与 `source_ref` 不可重复 |

初始草稿修订号为 `0` 时的保存请求示例：

```bash
curl -X PUT "$PLATFORM_URL/api/apps/starpath/admin/extensions/mcp/draft" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"expected_revision":0,"reason":"发布示例检索工具","items":[{"id":"search-example","name":"示例检索","description":"公司检索工具","version":"1.0.0","source_ref":"company:search-example","listed":true,"visible":true,"sort_order":0,"distribution":{"mode":"users","user_ids":[42]},"skill":null,"mcp":{"transport":"http","command":"","args":[],"url":"https://mcp.example.invalid/search","required_config":[]},"disabled":false}]}'
```

成功时返回 HTTP 200 及更新后的完整目录快照（同 GET 响应结构）。后续操作必须使用响应中返回的最新 `draft_revision`。

### POST /api/apps/starpath/admin/extensions/mcp/publish

请求体上限 1 MiB。仅接收 `expected_revision`（当前草稿修订号）和 `reason` 两个字段，**不接收 `items`**，直接将服务端当前已保存的草稿定版发布：

```json
{
  "expected_revision": 1,
  "reason": "确认发布给试用账号"
}
```

成功返回 HTTP 200 及最新目录快照。若其他管理员已修改草稿导致 `expected_revision` 不匹配，返回 HTTP 409：

```json
{
  "success": false,
  "message": "EXTENSION_REVISION_CONFLICT",
  "code": "EXTENSION_REVISION_CONFLICT"
}
```

遇到 `409` 时必须重新 GET 最新草稿比对差异，禁止盲目覆盖。此外，历史上曾经出现过的扩展 `id` 不允许在后续版本中更改其 `source_ref`（否则返回 `409 EXTENSION_IDENTITY_CONFLICT`）。

## 目录条目字段规范（Skills 与 MCP）

| 条目字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `id` | string | 是 | 稳定唯一标识，1–128 字符，匹配 `^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$` |
| `name` | string | 是 | 展示名称，去首尾空白后 1–120 字符 |
| `description` | string | 是 | 简介说明，允许空字符串 `""`，最多 2000 字符 |
| `version` | string | 是 | 固定版本标识，1–128 字符；禁止 latest（不区分大小写）、空白及 `*^~<>=|`，托管制品确认另有版本校验 |
| `source_ref` | string | 是 | 稳定来源引用（不含空白，1–256 字符）；当技能 `kind === "hosted_artifact"` 时必须严格等于 `"company:" + id` |
| `listed`、`visible` | boolean | 是 | 是否在目录列表中展示/可见；设为 `false` 仅停止推荐展示，不卸载用户已装项 |
| `sort_order` | integer | 是 | 排序权重，范围 `-1000000` 至 `1000000`（服务端按 `sort_order` 降序、同权重按 `id` 升序排列） |
| `distribution` | object | 是 | 分发目标对象，严格包含 `mode` 与 `user_ids` 两个键 |
| `distribution.mode` | string | 是 | 分发模式枚举：`"all"`（全员可见）或 `"users"`（仅指定用户可见） |
| `distribution.user_ids` | integer[] | 是 | 必须始终为数组：`mode === "all"` 时必须为空数组 `[]`；`mode === "users"` 时为 1–500 个不重复的正整数用户 ID（服务端自动按升序排列） |
| `skill` | object / null | 是 | 当 `module === "skills"` 时为技能描述对象（详见第六章），当 `module === "mcp"` 时必须为 `null` |
| `mcp` | object / null | 是 | 当 `module === "mcp"` 时为 MCP 描述对象（见下表），当 `module === "skills"` 时必须为 `null` |
| `disabled` | boolean | 否 | 是否远程统一停用该条目；**仅 `mcp` 模块允许设为 `true`**（`skills` 模块若设为 `true` 返回 `400 INVALID_EXTENSION_INPUT`） |

### MCP 描述对象（`item.mcp`）

| 字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `transport` | string | 是 | 传输方式枚举：`"stdio"`、`"http"` 或 `"sse"` |
| `command` | string | 是 | `transport === "stdio"` 时为非空可执行命令（1–256 字符，禁止含换行符）；`"http"` 或 `"sse"` 时必须为空字符串 `""` |
| `args` | string[] | 是 | `transport === "stdio"` 时为启动参数数组（0–30 项，每项最多 256 字符，可为空；禁止换行以及含 token= 或 password= 的秘密参数）；`"http"` 或 `"sse"` 时必须为空数组 `[]` |
| `url` | string | 是 | `transport` 为 `"http"` 或 `"sse"` 时为HTTPS 地址，或仅本机 localhost、127.0.0.1、::1 的 HTTP 地址（最多 2000 字符，禁止内嵌账号密码、查询参数或 fragment）；`"stdio"` 时必须为空字符串 `""` |
| `required_config` | object[] | 是 | 需要用户在星途本地补充填写的配置项声明数组（0–30 项）；**后台目录严禁直接下发真实密钥值** |
| `required_config[].key` | string | 是 | 配置键名（1–64 字符，匹配 `^[a-zA-Z_][a-zA-Z0-9_.-]{0,63}$`，同一条目内不可重复） |
| `required_config[].label` | string | 是 | 界面提示标签（去首尾空白后 1–120 字符） |
| `required_config[].secret` | boolean | 是 | 是否按敏感密码框采集与脱敏保存 |
| `required_config[].required` | boolean | 是 | 用户启用或一键更新该 MCP 前是否必须填写 |

## 内置连接器后台配置（`PUT /admin/extensions/connectors`）

连接器代码随星途客户端安装包内置提供，平台仅允许管理固定白名单内 **5 个内置连接器**的推荐展示信息，不支持通过后台新增未知连接器或修改底层包名。

5 个内置连接器的固定基线标识如下：

| `plugin_id`（即 `items[].id`） | `runtime_id` | `package_name` | `placement` | 默认名称 |
|---|---|---|---|---|
| `openclaw-weixin` | `openclaw-weixin` | `@tencent-weixin/openclaw-weixin` | `remote_connections` | 微信连接器 |
| `dingtalk-connector` | `dingtalk-connector` | `@dingtalk-real-ai/dingtalk-connector` | `remote_connections` | 钉钉连接器 |
| `openclaw-lark` | `openclaw-lark` | `@larksuite/openclaw-lark` | `remote_connections` | 飞书连接器 |
| `wecom-openclaw-plugin` | `wecom-openclaw-plugin` | `@wecom/wecom-openclaw-plugin` | `remote_connections` | 企业微信连接器 |
| `clawemail-email` | `email` | `@clawemail/email` | `email_settings` | 邮件连接器 |

调用 `PUT /api/apps/starpath/admin/extensions/connectors` 时，请求体包含 `expected_revision`（取自 `GET /admin/extensions/connectors` 的 `config_revision`）、`reason`（1–500 字符）和 `items`：
- **允许管理员修改的字段**：`name`（1–120 字符）、`description`（0–2000 字符）、`listed`（boolean）、`visible`（boolean）、`sort_order`（`-1000000` 至 `1000000`）。
- **只读锁定字段**（必须与基线完全一致，否则返回 `400 CONNECTOR_FIELD_READ_ONLY` 或 `CONNECTOR_SCOPE_DENIED`）：`id`、`runtime_id`、`package_name`、`management_type`（固定 `"bundled"`）、`placement`、`enabled`（固定 `true`）、`required_feature_ids`（固定 `[]`）、`target_release_id`（固定 `null`）、`install_paused`（固定 `false`）、`upgrade_deadline`（固定 `null`）。
- 保存成功后立即更新连接器发布修订号 `config_revision`，星途客户端通过 `GET /api/apps/starpath/config` 读取生效。
