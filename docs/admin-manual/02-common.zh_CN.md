# 登录接入（第一阶段）

本章是第一阶段的核心接入契约，目标是让星途用户使用 MaaS 账号完成登录，并准确读取当前账号、刷新短期访问令牌、退出登录与安全切换账号。

**第一阶段能力范围声明**：
- **已实现并作为第一阶段主线**：状态探测（`GET /api/status`）、账号密码登录（`POST /api/user/login`）、会话刷新（`POST /api/user/auth/refresh`）、当前账号读取（`GET /api/user/self`）、退出登录（`POST /api/user/auth/logout`）。
- **二次验证（2FA）**：**待审核 / 未实现**。不属于第一阶段已交付能力；保留接口与字段说明供领导审核。当账号密码登录返回 `require_2fa: true` 时，必须提示暂不支持该验证方式，不得视为登录成功或绕过验证。
- **短信验证码登录（SMS）**：**待审核 / 开放与验收状态待确认**。不作为第一阶段已交付、已验收能力对外承诺；保留接口字段与结构示例供领导审核。

## 第一阶段完成标准

1. 启动或配置平台地址后，先读取 `GET /api/status` 获取登录方式开关与额度展示换算配置。
2. 账号密码登录成功后，保存短期 `access_token`、`session.sid` 与 `new_api_refresh` 刷新 Cookie，并能通过 `GET /api/user/self` 读取当前账号与额度；若返回 `require_2fa: true`，提示暂不支持二次验证，不落盘登录态、不请求托管密钥。
3. 访问令牌过期（或距过期不足 30 秒）时通过刷新 Cookie 调用 `POST /api/user/auth/refresh` 轮换令牌；刷新确认返回 401 或会话被撤销时清空失效会话并明确提示；网络故障、429 或 5xx 保留已有凭据并提示稍后重试，禁止死循环重试。
4. 退出登录调用 `POST /api/user/auth/logout` 并清理本地凭据、Cookie 与托管模型密钥；切换账号时严格校验 `origin + user.id + session.sid`，防止旧账号异步响应污染新账号。
5. 地址非法、响应格式异常、网络故障、本地存储失败、账号密码/验证码拒绝、触发限流（429）能分别识别提示，不把网络或服务异常一律提示为“密码错误”。

登录态验证使用 `GET /api/user/self`，不依赖插件目录、技能安装或套餐购买。模型调用密钥（`MODEL_API_KEY`）与登录访问令牌（`ACCESS_TOKEN`）严格分开管理。

## 接口索引

| 方法 | 路径 | 认证方式与关键请求头 | 第一阶段状态 | 用途 |
|---|---|---|---|---|
| GET | `/api/status` | 无需认证 | 已实现 / 必接 | 查询平台开放的登录方式开关与额度展示换算参数 |
| POST | `/api/user/login` | 无需认证（受同源、Turnstile 与登录限流保护） | 已实现 / 必接 | 账号密码登录，签发短期 Bearer 令牌并下发刷新 Cookie |
| POST | `/api/user/auth/refresh` | `Cookie: new_api_refresh=...`、`Origin: $PLATFORM_URL`、可选 `X-Auth-Session: $SESSION_ID` | 已实现 / 必接 | 轮换短期访问令牌与刷新 Cookie |
| GET | `/api/user/self` | `Authorization: Bearer $ACCESS_TOKEN` | 已实现 / 必接 | 获取当前登录用户信息与钱包内部 quota 额度 |
| POST | `/api/user/auth/logout` | `Authorization: Bearer $ACCESS_TOKEN` 及/或 `Cookie: new_api_refresh=...`、`Origin: $PLATFORM_URL`、可选 `X-Auth-Session: $SESSION_ID` | 已实现 / 必接 | 撤销当前服务端登录会话并清除刷新 Cookie |
| POST | `/api/user/login/sms/code` | 无需认证（需后台开启 `sms_login`） | **待审核 / 开放与验收状态待确认** | 申请短信登录验证码 |
| POST | `/api/user/login/sms` | 无需认证（需后台开启 `sms_login`） | **待审核 / 开放与验收状态待确认** | 提交短信挑战令牌与验证码完成登录 |
| POST | `/api/user/login/2fa` | 无需认证（持密码或短信登录返回的 `flow_token`） | **待审核 / 未实现** | 提交 TOTP 二次验证码或备用恢复码完成登录 |

## GET /api/status

无请求体、无需认证。客户端在登录前和账单换算时调用，用于读取平台登录开关与内部 quota 的货币/积分展示规则。以下仅列出星途客户端实际读取的字段（平台返回的其余站点配置字段可忽略）：

```bash
curl "$PLATFORM_URL/api/status"
```

响应结构示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "password_login_enabled": true,
    "sms_login": false,
    "turnstile_check": false,
    "quota_per_unit": 500000,
    "quota_display_type": "USD",
    "display_in_currency": true,
    "usd_exchange_rate": 7.2,
    "custom_currency_symbol": "积分",
    "custom_currency_exchange_rate": 100
  }
}
```

| `data` 字段 | 类型 | 说明与客户端解析规则 |
|---|---|---|
| `password_login_enabled` | boolean | 是否允许账号密码登录；客户端当该值不为 `false` 时允许密码登录，显式为 `false` 时拦截密码登录 |
| `sms_login` | boolean | 平台是否开启短信登录（`true` 为开启；星途侧短信登录当前属“待审核 / 开放与验收状态待确认”） |
| `turnstile_check` | boolean | 平台是否要求 Cloudflare Turnstile 人机验证；为 `true` 时桌面端直接接口登录会被中间件拦截，需引导外部或内嵌网页完成验证 |
| `quota_per_unit` | number | 每 `1 USD` 对应的平台内部整数 quota 数量（正数，默认常见配置为 `500000`）；用于钱包与套餐额度换算 |
| `quota_display_type` | string | 额度展示模式枚举：`"USD"`（美元）、`"CNY"`（人民币）、`"CUSTOM"`（自定义货币/积分）、`"TOKENS"`（直接按内部 quota 整数展示） |
| `display_in_currency` | boolean，可缺失 | 旧版兼容布尔字段；仅当 `quota_display_type` 缺失时回退使用（`true` 视为 `"USD"`，`false` 视为 `"TOKENS"`） |
| `usd_exchange_rate` | number，可缺失 | `1 USD` 兑换人民币汇率（正数）；当 `quota_display_type === "CNY"` 时必填有效 |
| `custom_currency_symbol` | string，可缺失 | 自定义展示符号（去首尾空白后 1–32 字符）；当 `quota_display_type === "CUSTOM"` 时必填 |
| `custom_currency_exchange_rate` | number，可缺失 | `1 USD` 兑换自定义单位的比率（正数）；当 `quota_display_type === "CUSTOM"` 时必填 |

## POST /api/user/login

账号密码登录。请求头 `Content-Type: application/json`，并携带 `Origin: $PLATFORM_URL`。若 `/api/status` 返回 `password_login_enabled: false`，客户端不得强行调用，服务端也会返回 `403`（`"管理员关闭了密码登录"`）。

| 请求字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `username` | string | 是 | 登录用户名（或平台允许的邮箱）；客户端去首尾空白后长度 1–256 字符 |
| `password` | string | 是 | 登录密码；客户端长度 1–1024 字符，严禁写入日志、错误上报或本地明文配置 |

> **查询参数说明**：平台路由 `POST /api/user/login` 上挂载了 `TurnstileCheck` 中间件（读取 URL 查询参数 `?turnstile=...`）。当平台关闭 `turnstile_check` 时无需传该参数。

```bash
curl -i -X POST "$PLATFORM_URL/api/user/login" \
  -H "Origin: $PLATFORM_URL" \
  -H 'Content-Type: application/json' \
  -d '{"username":"example-user","password":"REPLACE_WITH_PASSWORD"}'
```

### 登录成功响应（未启用 2FA）

成功时 HTTP 状态码为 `200`，响应头通过 `Set-Cookie` 下发 `new_api_refresh`：

```http
HTTP/1.1 200 OK
Set-Cookie: new_api_refresh=SESSION_ID_PLACEHOLDER.REFRESH_SECRET_PLACEHOLDER; Path=/api/user/auth; Max-Age=2592000; HttpOnly; Secure; SameSite=Strict
Content-Type: application/json; charset=utf-8
```

```json
{
  "success": true,
  "message": "",
  "data": {
    "access_token": "ACCESS_TOKEN_PLACEHOLDER",
    "token_type": "Bearer",
    "access_expires_at": 1792714500,
    "session": {
      "sid": "SESSION_ID_PLACEHOLDER",
      "current": true,
      "login_method": "password",
      "ip": "127.0.0.1",
      "user_agent": "Starpath/1.0.0",
      "created_at": 1792713600,
      "last_active_at": 1792713600,
      "expires_at": 1795305600
    },
    "user": {
      "id": 42,
      "username": "example-user",
      "display_name": "示例用户",
      "role": 1,
      "status": 1,
      "email": "",
      "quota": 500000,
      "used_quota": 10000,
      "request_count": 8,
      "group": "default",
      "aff_code": "ABCD",
      "inviter_id": 0
    }
  }
}
```

| 响应字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `data.access_token` | string | 是 | 短期访问令牌（HMAC 签名），仅保存在客户端主进程内存，随 `Authorization: Bearer` 发送给平台 |
| `data.token_type` | string | 是 | 固定为 `"Bearer"` |
| `data.access_expires_at` | integer | 是 | 短期访问令牌到期时间戳（**Unix 秒**）；客户端在距该时间不足 30 秒时触发刷新 |
| `data.session.sid` | string | 是 | 服务端登录会话唯一标识（UUID）；客户端必须校验非空并持久化，用于 `X-Auth-Session` 请求头 |
| `data.session.current` | boolean | 是 | 是否为当前会话（登录与刷新响应中为 `true`） |
| `data.session.login_method` | string | 是 | 登录方式标识，密码登录为 `"password"` |
| `data.session.ip`、`user_agent` | string | 是 | 会话创建或最近活跃时的客户端 IP 与 User-Agent |
| `data.session.created_at`、`last_active_at`、`expires_at` | integer | 是 | 会话创建、最近活跃、刷新会话到期时间戳（均为 **Unix 秒**） |
| `data.user.id` | integer | 是 | 平台用户正整数 ID |
| `data.user.username` | string | 是 | 登录用户名 |
| `data.user.display_name` | string | 否 | 用户展示名称（可为空字符串） |
| `data.user.quota`、`used_quota` | integer | 是 | 钱包剩余额度与累计已用额度（均为**内部整数 quota 单位**，非负安全整数） |
| `data.user.request_count` | integer | 否 | 累计请求次数 |
| `data.user.role`、`status` | integer | 否 | 平台角色（`1` 普通用户、`10` 普通管理员、`100` Root 超级管理员）与账号状态（`1` 启用） |

**刷新 Cookie（`new_api_refresh`）属性与保管规范**：
- Cookie 名称固定为 `new_api_refresh`，值格式为 `<sid>.<secret>`，具有 `Path=/api/user/auth`、`HttpOnly`、`SameSite=Strict`，在 HTTPS 或平台开启 `SessionCookieSecure` 时带 `Secure`。
- 桌面客户端主进程从响应头 `Set-Cookie` 提取 `new_api_refresh`，使用系统安全加密（`safeStorage`）保存在本地，并仅在向该平台 `origin` 请求 `/api/user/auth/*` 时通过 `Cookie` 请求头发送；渲染进程不得直接读取该刷新令牌。
- 当账号开启了二次验证（2FA）时，`POST /api/user/login` 返回 HTTP 200 且 `data.require_2fa === true`（不会下发 `access_token` 和 `new_api_refresh`）。此时认证未完成，客户端不得视为登录成功或同步托管密钥。

## 未支持或待确认的登录方式

| 登录能力 | 平台后端代码状态 | 星途客户端代码状态 | 当前审核与交付结论 |
|---|---|---|---|
| **二次验证（2FA：TOTP / 备用恢复码 / 短信 MFA）** | 平台源码有 `/api/user/login/2fa`、`/api/user/login/sms/mfa/code`、`/api/user/login/sms/mfa` | 主进程 `maasAuth.ts` 有 `twofa` 分支，但渲染层未接通二次验证交互闭环 | **待审核 / 未实现**：不属于第一阶段已交付能力。遇到 `require_2fa: true` 须提示当前版本暂不支持该账号登录，不得绕过验证 |
| **短信验证码登录（SMS）** | 平台源码有 `/api/user/login/sms/code`、`/api/user/login/sms`（需后台开启 `sms_login` 并配置短信服务商） | 主进程与登录弹窗含手机号验证码分支 | **待审核 / 开放与验收状态待确认**：不作为第一阶段已交付、已验收能力对外承诺，保留接口供领导与研发审核 |

平台源码或客户端分支中出现上述路由、类型或适配代码，不代表目标环境已开启或已完成端到端业务验收。上线前须分别核实：平台配置是否开放、短信通道与限流是否验收、客户端界面是否闭环。

## 待审核接口：短信与二次验证

**状态标注：二次验证为“待审核 / 未实现”；短信登录为“待审核 / 开放与验收状态待确认”。** 本节完整保留源码中的接口路径、请求/响应字段与错误码，供领导审核能力边界及后续阶段排期使用。

| 方法与路径 | 状态标注 | 请求字段（JSON） | 成功响应关键字段（`data`） |
|---|---|---|---|
| `POST /api/user/login/sms/code` | 待审核 / 开放与验收状态待确认 | `phone`（string，必填，中国大陆 11 位手机号 `^1[3-9]\d{9}$`） | `challenge_token`（string）、`expires_in`（integer，秒）、`retry_after`（integer，秒） |
| `POST /api/user/login/sms` | 待审核 / 开放与验收状态待确认 | `challenge_token`（string，必填）、`code`（string，必填，4–8 位数字） | 未启用 2FA 时返回与密码登录相同的 `access_token`、`session`、`user` 及 `Set-Cookie`；若该账号启用了 TOTP 2FA 则返回 `require_2fa: true`、`flow_token`、`expires_at`、`sms_available: false` |
| `POST /api/user/login/2fa` | **待审核 / 未实现** | `flow_token`（string，必填）、`code`（string，必填，6 位 TOTP 数字或 8 位十六进制备用恢复码） | 验证通过后返回 `access_token`、`token_type`、`access_expires_at`、`session`（`login_method: "2fa"`）、`user` 及 `Set-Cookie` |
| `POST /api/user/login/sms/mfa/code` | **待审核 / 未实现** | `flow_token`（string，必填，来自密码登录返回的 `require_2fa` 响应且 `sms_available: true`） | `challenge_token`（string）、`expires_in`（integer，秒）、`retry_after`（integer，秒） |
| `POST /api/user/login/sms/mfa` | **待审核 / 未实现** | `flow_token`（string，必填）、`challenge_token`（string，必填）、`code`（string，必填） | 验证通过后返回 `access_token`、`session`（`login_method: "password_sms_mfa"`）、`user` 及 `Set-Cookie` |

### 短信验证码申请与短信登录（待审核 / 开放与验收状态待确认）

申请短信验证码 `POST /api/user/login/sms/code` 请求体：

```json
{
  "phone": "13800138000"
}
```

成功响应示例（有效期与重发间隔以平台实际配置为准，默认常见为 300 秒与 60 秒）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "challenge_token": "CHALLENGE_TOKEN_PLACEHOLDER",
    "expires_in": 300,
    "retry_after": 60
  }
}
```

| 响应字段 | 类型 | 单位与说明 |
|---|---|---|
| `challenge_token` | string | 本次短信验证挑战标识，仅保存在内存用于下一步提交，严禁写入日志 |
| `expires_in` | integer | 验证码有效时长（**秒**） |
| `retry_after` | integer | 允许再次请求发送短信前的最小等待时长（**秒**） |

提交短信验证码登录 `POST /api/user/login/sms` 请求体：

```json
{
  "challenge_token": "CHALLENGE_TOKEN_PLACEHOLDER",
  "code": "123456"
}
```

短信接口专属错误码（`code`）：
- `400 PHONE_NUMBER_INVALID`：手机号格式非法。
- `401 PHONE_VERIFICATION_INVALID`：验证码错误、已过期或挑战令牌失效。
- `429 PHONE_RATE_LIMITED`：发送或验证频率超限（同时返回 `Retry-After` 响应头）。
- `503 PHONE_SMS_UNAVAILABLE`：平台未开启短信登录或短信网关不可用（客户端映射为 `smsUnavailable`）。
- `500 PHONE_OPERATION_FAILED`：服务端内部处理异常。

### 二次验证（待审核 / 未实现）

当用户在平台启用了 2FA 时，`POST /api/user/login` 返回 HTTP 200 及如下结构：

```json
{
  "success": true,
  "message": "需要二次验证",
  "data": {
    "require_2fa": true,
    "flow_token": "FLOW_TOKEN_PLACEHOLDER",
    "expires_at": 1792713900,
    "sms_available": false
  }
}
```

| 响应字段 | 类型 | 单位与说明 |
|---|---|---|
| `require_2fa` | boolean | 固定为 `true`，表示第一阶段密码校验已过但登录尚未完成，**未签发任何访问令牌** |
| `flow_token` | string | 二次验证流程临时凭据（一次性消费） |
| `expires_at` | integer | 该二次验证流程到期时间戳（**Unix 秒**，平台默认 300 秒有效） |
| `sms_available` | boolean | 平台提示该账号是否具备短信二次验证条件（星途客户端未实现） |

提交二次验证 `POST /api/user/login/2fa` 请求体（拟接入结构）：

```json
{
  "flow_token": "FLOW_TOKEN_PLACEHOLDER",
  "code": "123456"
}
```

拟接入验证通过后，响应结构与账号密码登录成功一致（含 `access_token`、`session`、`user` 及 `Set-Cookie: new_api_refresh`）。在完成客户端正式实现与端到端验收前，客户端遇到 `require_2fa: true` 必须拒绝完成登录并明确提示用户。

## POST /api/user/auth/refresh

无业务请求体。用于在短期 `access_token` 过期或即将过期时轮换新令牌。请求必须携带 `Origin: $PLATFORM_URL` 与当前账号的 `Cookie: new_api_refresh=...`；若本地已记录会话 `sid`，应同时发送 `X-Auth-Session: $SESSION_ID`，防止多会话切换时串号。

```http
POST /api/user/auth/refresh HTTP/1.1
Host: maas.example.invalid
Origin: https://maas.example.invalid
Cookie: new_api_refresh=SESSION_ID_PLACEHOLDER.REFRESH_SECRET_PLACEHOLDER
X-Auth-Session: SESSION_ID_PLACEHOLDER
```

成功时返回 HTTP 200，通过 `Set-Cookie` 下发轮换后的新 `new_api_refresh`（旧刷新令牌立即失效，若重复使用旧刷新令牌将触发重放检测并撤销整条会话），响应体返回新的短期令牌、会话与用户快照：

```json
{
  "success": true,
  "message": "",
  "data": {
    "access_token": "NEW_ACCESS_TOKEN_PLACEHOLDER",
    "token_type": "Bearer",
    "access_expires_at": 1792715400,
    "session": {
      "sid": "SESSION_ID_PLACEHOLDER",
      "current": true,
      "login_method": "password",
      "ip": "127.0.0.1",
      "user_agent": "Starpath/1.0.0",
      "created_at": 1792713600,
      "last_active_at": 1792714500,
      "expires_at": 1795305600
    },
    "user": {
      "id": 42,
      "username": "example-user",
      "display_name": "示例用户",
      "quota": 500000,
      "used_quota": 10000,
      "request_count": 8
    }
  }
}
```

**刷新注意事项**：
1. 客户端所有并发请求共享同一个刷新单飞锁（in-flight promise），避免并发发出多个 `/api/user/auth/refresh` 导致第二个请求因旧 Cookie 已被轮换而触发 `409 AUTH_REFRESH_RACE` 或 `401 AUTH_SESSION_REVOKED`。
2. 当返回 `409 AUTH_SESSION_MISMATCH`（表示请求头 `X-Auth-Session` 与 Cookie 中的会话 SID 不一致）时，客户端主进程去掉 `X-Auth-Session` 重试 1 次刷新以对齐当前 Cookie 会话；后续仅在确认 401 失效时清除本地登录态；网络故障、429 或 5xx 不作为退出依据。
3. 当返回 `401`（`AUTH_UNAUTHORIZED` 或 `AUTH_SESSION_REVOKED`）时，说明刷新会话已过期、被管理员禁用、改密失效或退出，服务端会清除 Cookie，客户端必须立即退出登录态并清除本地保存的刷新 Cookie 与托管密钥，不得继续重试。

## GET /api/user/self

无请求体，通过 `Authorization: Bearer $ACCESS_TOKEN` 认证。用于校验当前登录态有效性、核对当前账号身份并同步钱包最新额度。

```bash
curl "$PLATFORM_URL/api/user/self" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

成功响应中星途客户端消费的核心字段如下（平台同时返回的其他用户属性按需忽略）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 42,
    "username": "example-user",
    "display_name": "示例用户",
    "quota": 500000,
    "used_quota": 10000,
    "request_count": 8
  }
}
```

| `data` 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | integer | 是 | 当前登录账号正整数 ID；必须与本地当前会话记录的 `user.id` 核对一致 |
| `username` | string | 是 | 登录用户名（非空字符串） |
| `display_name` | string | 否 | 界面展示名；为空或缺失时客户端回退展示 `username` |
| `quota` | integer | 是 | 钱包当前剩余可用额度（**内部整数 quota 单位**，非负） |
| `used_quota` | integer | 是 | 账号历史累计已消耗额度（**内部整数 quota 单位**，非负） |
| `request_count` | integer | 否 | 账号历史累计请求次数 |

## POST /api/user/auth/logout

无业务请求体。用于显式退出当前登录会话。客户端调用时应同时携带当前 `Authorization: Bearer $ACCESS_TOKEN`（若有）、`Cookie: new_api_refresh=...`、`Origin: $PLATFORM_URL` 以及 `X-Auth-Session: $SESSION_ID`。

```http
POST /api/user/auth/logout HTTP/1.1
Host: maas.example.invalid
Origin: https://maas.example.invalid
Authorization: Bearer ACCESS_TOKEN_PLACEHOLDER
Cookie: new_api_refresh=SESSION_ID_PLACEHOLDER.REFRESH_SECRET_PLACEHOLDER
X-Auth-Session: SESSION_ID_PLACEHOLDER
```

### 退出响应结构

1. **携带有效 Bearer 访问令牌退出时**（服务端撤销该令牌所属 `sid` 会话，若刷新 Cookie 匹配同一会话则一并撤销并通过 `Set-Cookie: new_api_refresh=; Max-Age=-1` 清除 Cookie）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "revoked_sid": "SESSION_ID_PLACEHOLDER",
    "cookie_cleared": true
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `data.revoked_sid` | string | 已撤销的服务端会话 SID |
| `data.cookie_cleared` | boolean | 是否同时清除了匹配的 `new_api_refresh` Cookie（若请求携带的 Cookie 属于其他会话则为 `false`） |

2. **仅携带刷新 Cookie 退出（无 Bearer 令牌或令牌已过期）时**（服务端按刷新 Cookie 撤销会话并下发清除 Cookie 响应头）：

```json
{
  "success": true,
  "message": ""
}
```

**退出与换号容错要求**：
- 若 `X-Auth-Session` 与 Bearer 令牌或刷新 Cookie 的会话 SID 不一致，服务端返回 `409 AUTH_SESSION_MISMATCH`。客户端收到 `409` 时可先刷新一次对齐会话再重试退出。
- **无论远端 `/api/user/auth/logout` 是否成功（例如断网、超时或返回非 200）**，客户端都必须无条件清空本地内存访问令牌、持久化刷新 Cookie、Electron Session Cookie 及当前账号的托管模型密钥，并通过 `logoutRemoteFailed: true` 标记告知调用方远端撤销未确认。
- 切换账号按“先完整清理旧账号状态，再建立新账号会话”处理；旧账号在途异步请求返回后因会话代数（epoch / sid）不匹配必须直接丢弃。

## 认证与会话异常速查

### 平台认证会话错误码（`code`）

| HTTP 状态 | 平台错误码 `code` | 触发原因 | 客户端处理要求 |
|---|---|---|---|
| `401` | `AUTH_UNAUTHORIZED` | 未提供访问令牌/刷新 Cookie、令牌签名非法、格式错误或对应用户不存在 | 读取接口尝试单飞刷新 1 次；刷新接口收到此码则清除本地登录态并提示重新登录 |
| `401` | `AUTH_TOKEN_EXPIRED` | Bearer 短期访问令牌已过期 | 触发 `POST /api/user/auth/refresh` 轮换新令牌后重试原读取请求 |
| `401` | `AUTH_SESSION_REVOKED` | 会话已过期、已主动退出、被管理员禁用、用户修改密码导致 `auth_version` 变更，或检测到旧刷新令牌重放 | 立即清除本地刷新 Cookie、登录态与托管密钥，提示登录已失效 |
| `409` | `AUTH_SESSION_MISMATCH` | `X-Auth-Session` 请求头与 Bearer 令牌或刷新 Cookie 中的会话 SID 不一致 | 刷新时去掉 `X-Auth-Session` 重试 1 次；退出时对齐会话后重试并兜底清理本地状态 |
| `409` | `AUTH_REFRESH_RACE` | 同一会话并发刷新导致刷新令牌哈希乐观锁冲突 | 客户端应通过单飞锁避免并发刷新；偶发遇到时重新读取最新会话状态 |
| `409` | `AUTH_SESSION_LIMIT` | 当前账号活跃会话数达到上限且淘汰旧会话冲突 | 提示登录冲突，请用户稍后重试登录 |
| `429` | `AUTH_SESSION_ISSUANCE_LIMIT`（或网关限流） | 该账号在统计窗口内新建会话次数超限，或触发接口速率限制 | 读取 `Retry-After` 响应头并退避，向用户提示操作过于频繁（`rateLimited`） |
| `500` | `AUTH_INTERNAL_ERROR` | 服务端数据库或内部异常 | 保留本地已有配置，提示服务暂不可用，不误报为密码错误 |

### 星途客户端内部错误分类（IPC `MaasErrorKind`）

客户端主进程将底层异常归纳为以下枚举返回给界面层，界面应按类型给出准确提示：

| 客户端错误枚举 | 含义与触发场景 | 界面提示要求 |
|---|---|---|
| `invalidAddress` | 平台地址格式非法（非 HTTP/HTTPS、非本机使用 HTTP、包含路径/查询参数/内嵌凭据） | 提示检查平台服务器地址格式 |
| `invalidResponse` | 服务端返回非 JSON、缺少必填字段（如缺少 `access_token`、`session.sid`、`user.id`）或数值越界 | 提示平台响应异常或版本不匹配 |
| `network` | DNS/TCP 连接失败、请求超时（默认 10 秒）或未预期 5xx | 提示网络连接或服务器异常，**不清空已保存的刷新凭据** |
| `storage` | 本地 `safeStorage` 加密不可用或状态文件读写失败 | 提示本机安全存储异常 |
| `loginRejected` | 账号密码错误、账号被禁用、密码登录被关闭、需要暂不支持的 2FA 或认证已失效 | 结合具体场景提示账号密码错误、暂不支持二次验证或登录已过期 |
| `rateLimited` | 收到 HTTP 429 或仍处于 `retryAfter` 退避窗口内 | 提示请求过于频繁及剩余等待时间 |
| `smsUnavailable` | 平台未开启 `sms_login` 或返回 `503 PHONE_SMS_UNAVAILABLE` | 提示当前平台未开放短信登录或短信服务不可用 |
| `cancelled` | 用户主动关闭外部登录窗口或取消操作 | 静默恢复界面按钮状态，不报错 |
