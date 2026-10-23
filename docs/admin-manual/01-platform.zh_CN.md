# 接入约定

## 地址与认证方式

MaaS 平台接口按用途分为三类，使用的认证凭据、请求头和响应格式严格隔离：

| 接口族 | 根路径 | 认证凭据与关键请求头 | 响应协议 |
|---|---|---|---|
| **平台账户与会话** | `/api` | 登录前无凭据；登录后使用 `Authorization: Bearer $ACCESS_TOKEN`；刷新与退出使用 `Cookie: new_api_refresh=...`、可选 `X-Auth-Session: $SESSION_ID` 及同源 `Origin: $PLATFORM_URL` | 平台统一 JSON 包装（`success` / `message` / `data` / `code`） |
| **星途专属与公司管理** | `/api/apps/starpath` | `Authorization: Bearer $ACCESS_TOKEN`；管理接口 `/admin/*` 另需 Root（role=100）或已授权管理员（role=10 且具备对应模块权限） | 平台统一 JSON 包装（`success` / `message` / `data` / `code`） |
| **模型发现与对话调用** | `/v1` | `Authorization: Bearer $MODEL_API_KEY`（由托管密钥接口获取，**禁止传入登录 `ACCESS_TOKEN`**） | OpenAI 兼容 JSON 或 `text/event-stream`（SSE） |

**安全隔离要求**：
1. 禁止将 `new_api_refresh` 刷新 Cookie 或登录 `ACCESS_TOKEN` 发送给外部模型供应商，也禁止把 `MODEL_API_KEY` 当作用户或管理员登录令牌访问 `/api/*`。
2. 客户端只连接经过校验的平台 `origin`（不允许携带路径、查询参数、散列或内嵌账号密码；生产环境使用 HTTPS，本机开发仅允许 `localhost` 或 `127.0.0.1` 使用 HTTP）。
3. 平台对携带 Cookie 的认证操作启用同源校验中间件（`SessionCookieOriginGuard`），客户端主进程发起请求时须携带 `Origin: $PLATFORM_URL`，并在重定向时拒绝跨源跳转。

```bash
curl "$PLATFORM_URL/api/user/self" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## 平台 JSON 响应结构

除 `/v1/*` 模型接口外，平台接口统一返回如下 JSON 包装结构：

| 字段 | 类型 | 必填 / 出现时机 | 说明 |
|---|---|---|---|
| `success` | boolean | 始终返回 | 业务是否成功；**即使 HTTP 状态码为 200，客户端也必须检查 `success === true`** |
| `message` | string | 可缺失（成功常为 `""`） | 面向人类阅读的提示或失败说明，禁止用自然语言文案硬编码匹配业务分支 |
| `data` | object / array / null | 成功时返回 | 业务负载数据，具体结构见各接口定义；失败时通常缺失或为 `null` |
| `code` | string | 专属接口失败时返回 | 机器可读错误码（如 `AUTH_SESSION_REVOKED`、`EXTENSION_REVISION_CONFLICT`）；部分通用老接口仅返回 `success: false` 和 `message`，不保证提供 `code` |

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 42,
    "username": "example-user"
  }
}
```

带机器错误码的失败响应示例（如 HTTP 409）：

```json
{
  "success": false,
  "message": "EXTENSION_REVISION_CONFLICT",
  "code": "EXTENSION_REVISION_CONFLICT"
}
```

通用接口失败响应示例（HTTP 200 或 4xx，无 `code`）：

```json
{
  "success": false,
  "message": "用户名或密码错误"
}
```

`/v1/*` 模型接口不使用 `success` 包装，失败时返回顶层 `error` 对象（含 `message`、`type`、`code`）。网关、限流或反向代理层也可能返回非 JSON 响应，客户端须同时记录 HTTP 状态码并做安全兜底解析。

## 数据类型、时间与额度单位约定

| 维度 | 规范与单位约定 |
|---|---|
| **账号与资源 ID** | 平台用户 ID（`user.id`）、套餐 ID（`plan.id`）、订阅 ID（`subscription.id`）、令牌 ID（`data.id`）均为**正整数**；会话 SID（`session.sid`）、扩展条目 ID、发布 ID（`publication_id`）、分发 ID（`delivery_id`）、制品版本 ID（`release_id`）均为**字符串** |
| **时间戳单位** | **本手册下列平台时间戳字段均为 Unix 秒**，包括 `access_expires_at`、`session.created_at` / `last_active_at` / `expires_at`、`subscription.start_time` / `end_time` / `next_reset_time` / `last_reset_time`、`published_at`、`download.expires_at`、`updated_at` 等。客户端内部 `Date.now()`、`expiresAt`、`syncedAt` 使用**毫秒**，转换时必须乘除 `1000`，严禁混算 |
| **时长与倒计时** | `expires_in`、`retry_after`、`refresh_interval_seconds`、`signed_url_ttl_seconds`、`custom_seconds`、`quota_reset_custom_seconds` 单位均为**秒（integer）** |
| **额度（Quota）与金额** | 平台内部额度字段（`user.quota`、`user.used_quota`、`plan.total_amount`、`subscription.amount_total`、`subscription.amount_used`）均为**平台内部整数 quota 单位（int64）**；套餐价格 `plan.price_amount` 为对应币种（`plan.currency`，如 `"USD"`）的**十进制货币金额（number）**。展示给用户的积分或货币金额需通过 `GET /api/status` 返回的 `quota_per_unit`（每 1 美元对应的内部 quota 数，如 `500000`）和 `quota_display_type` 换算，严禁把内部 quota 与商品价格直接相加或写死换算常数 |
| **空值与布尔值** | `null`、空数组 `[]` 与字段缺失含义不同（如目录尚未发布时 `publication_id` 为 `null`，`items` 为 `[]`）。布尔值必须保留显式 `false`，不可用宽松真假判断代替“字段是否存在” |
| **严格写入校验** | `/api/apps/starpath/*` 下的管理端保存草稿、发布、权限、连接器、存储与客户端回执接口均启用严格字段检查（拒绝未知附加字段、拒绝非允许的 `null` 值），请求体须严格按手册字段发送 |
| **账号缓存隔离** | 登录态、托管密钥、订阅预览与公司扩展目录均按 `origin + accountId + session.sid (epoch)` 绑定；退出或切换账号后必须立即失效旧上下文，迟到的异步响应不得写回新账号 |

## 重试与容错原则

| 请求场景 | 客户端与调用方处理规则 |
|---|---|
| **带登录态的普通读取请求** | 收到 `401`（或令牌距过期不足 30 秒）时，通过单飞锁调用一次 `POST /api/user/auth/refresh`；刷新成功且仍属同一账号会话时重试原请求**至多 1 次** |
| **刷新接口会话冲突（409）** | `POST /api/user/auth/refresh` 因 `X-Auth-Session` 与 Cookie 会话不一致返回 `409 AUTH_SESSION_MISMATCH` 时，客户端可不带 `X-Auth-Session` 重试刷新 1 次以同步或覆盖为 Cookie 当前会话 |
| **触发限流（429）** | 解析响应头 `Retry-After`（秒或日期；缺失时默认退避 60 秒），在退避窗口内停止自动刷新或轮询，向界面返回 `rateLimited` |
| **钱包扣款购买套餐** | `POST /api/subscription/balance/pay` 遇到网络中断、超时或 5xx 时，扣款结果未知（客户端标记为 `unknown`），**严禁自动重发购买请求**；必须先重新拉取钱包与订阅对账 |
| **扩展目录与安装回执** | `POST .../receipts` 与 `POST .../installations` **不套用** GET 请求的 401 自动刷新重放逻辑，由客户端本地回执账本在当前账号上下文后续同步时恢复补报 |
| **后台草稿/配置版本冲突（409）** | 收到 `EXTENSION_REVISION_CONFLICT` 等并发冲突时，必须重新 GET 最新状态并比对差异，由管理员确认后再用新 `expected_revision` 提交，禁止自动强行覆盖 |

## HTTP 接口与星途客户端内部 IPC 的区分

对接与排查时必须分清**平台服务端 HTTP 路由**与**星途 Electron 客户端内部 IPC 通道**。渲染进程（前端界面）只调用 `window.electron.*` 暴露的 IPC，由客户端主进程统一发起 HTTP 请求并维护安全存储：

| 模块 | 客户端内部 Electron IPC 通道（仅客户端内部） | 对应平台 HTTP 接口（服务端路由） |
|---|---|---|
| **登录与账户** | `maas:snapshot`、`maas:configure`、`maas:login`、`maas:sendSms`、`maas:refresh`、`maas:logout`、`maas:openLogin`、`maas:billing`、`maas:managedKey`、`maas:changed` | `GET /api/status`、`POST /api/user/login`、`POST /api/user/auth/refresh`、`POST /api/user/auth/logout`、`GET /api/user/self`、`GET /api/subscription/plans`、`GET /api/subscription/self`、`GET /api/user/topup/info`、`POST /api/apps/starpath/managed-key` |
| **套餐购买确认** | `maas:prepareSubscriptionPurchase`、`maas:purchaseSubscriptionWithBalance` | 预览阶段读取 `/api/user/self`、`/api/subscription/plans`、`/api/subscription/self`、`/api/user/topup/info`、`/api/status`；确认扣款阶段调用 `POST /api/subscription/balance/pay`（注：`confirmationId` 仅存在于客户端内存，不发往服务端） |
| **连接器配置** | `maasManagedPlugins:read`、`maasManagedPlugins:refresh`、`maasManagedPlugins:changed` | `GET /api/apps/starpath/config?client_version=...&host_version=...&os=...&arch=...` |
| **公司技能与 MCP 目录** | `maasCompanyExtensions:read`、`maasCompanyExtensions:refresh`、`maasCompanyExtensions:takeNotice`、`maasCompanyExtensions:noticeShown`、`maasCompanyExtensions:changed` | `GET /api/apps/starpath/extensions/:module`（或 `GET /api/apps/starpath/v3/skills/catalog`）、`POST /api/apps/starpath/extensions/:module/receipts` |
| **托管技能包安装** | 客户端托管技能安装服务内部调用 | `POST /api/apps/starpath/v3/skills/releases/:release_id/download`、对象存储签名 URL `GET`、`POST /api/apps/starpath/v3/skills/installations` |

客户端 IPC 返回的错误标识（如 `invalidAddress`、`invalidResponse`、`network`、`storage`、`loginRejected`、`rateLimited`、`smsUnavailable`、`cancelled`、`notDeployed`、`staleRevision`）是客户端分类枚举，不等于平台 HTTP 响应中的 `code` 字段。
