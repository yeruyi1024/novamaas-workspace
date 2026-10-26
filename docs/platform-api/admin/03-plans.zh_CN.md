# 套餐与托管密钥

## 接口索引

| 方法 | 路径 | 认证方式 | 用途 |
|---|---|---|---|
| GET | `/api/subscription/plans` | `Authorization: Bearer $ACCESS_TOKEN` | 获取当前平台上架的可购买套餐列表 |
| GET | `/api/subscription/self` | `Authorization: Bearer $ACCESS_TOKEN` | 获取当前账号订阅记录及扣费优先级偏好 |
| GET | `/api/user/topup/info` | `Authorization: Bearer $ACCESS_TOKEN` | 获取平台支付合规确认开关与充值配置 |
| POST | `/api/subscription/balance/pay` | `Authorization: Bearer $ACCESS_TOKEN` | 使用当前账号钱包余额购买指定套餐 |
| POST | `/api/apps/starpath/managed-key` | `Authorization: Bearer $ACCESS_TOKEN` | 获取或创建当前账号的星途专属托管模型密钥 |

以上接口均使用登录短期访问令牌（`ACCESS_TOKEN`）。其中 `POST /api/apps/starpath/managed-key` 需要部署包含星途扩展能力的后端版本（`local/full-platform`）及 `tokens.managed` 数据库迁移；仅部署手册分支的服务器不包含该路由。

## GET /api/subscription/plans

无请求体。`data` 为数组，每项由外层 `plan` 对象包装（平台同时可能返回 `upgrade_Price` 等升级字段，星途客户端仅读取 `plan`）。

> **前置注意**：当平台后台未确认支付合规（`GET /api/user/topup/info` 中 `payment_compliance_confirmed === false`）时，服务端 `GET /api/subscription/plans` 会直接返回空数组 `data: []`。

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "plan": {
        "id": 7,
        "title": "示例月包",
        "subtitle": "用于说明接口结构",
        "price_amount": 10,
        "currency": "USD",
        "duration_unit": "month",
        "duration_value": 1,
        "custom_seconds": 0,
        "total_amount": 5000000,
        "quota_reset_period": "never",
        "quota_reset_custom_seconds": 0,
        "allow_balance_pay": true,
        "max_purchase_per_user": 1
      }
    }
  ]
}
```

| `plan` 字段 | 类型 | 必填 | 枚举 / 单位与说明 |
|---|---|---|---|
| `id` | integer | 是 | 套餐正整数 ID，用于发起购买 |
| `title` | string | 是 | 套餐标题（非空） |
| `subtitle` | string | 否 | 套餐副标题（可为空字符串） |
| `price_amount` | number | 是 | 商品货币价格（非负十进制数），币种由 `currency` 指定 |
| `currency` | string | 是 | 计价币种代码（平台固定为 `"USD"`；客户端不得根据界面语言自行推断币种） |
| `duration_unit` | string | 是 | 有效期单位枚举：`"year"`、`"month"`、`"day"`、`"hour"`、`"custom"` |
| `duration_value` | integer | 是 | 有效期单位数量（当 `duration_unit` 不为 `"custom"` 时为正整数） |
| `custom_seconds` | integer | 是 | 自定义有效时长（**秒**；当 `duration_unit === "custom"` 时大于 0） |
| `total_amount` | integer | 是 | 套餐每周期总额度（**内部整数 quota 单位**，非负） |
| `quota_reset_period` | string | 是 | 额度重置周期枚举：`"never"`（有效期内不重置）、`"daily"`（每日）、`"weekly"`（每周）、`"monthly"`（按自然月边界重置）、`"custom"`（自定义秒数） |
| `quota_reset_custom_seconds` | integer | 是 | 自定义重置间隔（**秒**；当 `quota_reset_period === "custom"` 时大于 0） |
| `allow_balance_pay` | boolean | 否 | 是否允许使用钱包余额直接购买（为 `true` 时客户端才允许发起余额购买） |
| `max_purchase_per_user` | integer | 否 | 单账号最大购买次数限制（`0` 表示不限购；正整数表示上限，客户端与服务端均做校验） |

**业务核对提醒**：套餐总有效期（`duration_unit` / `duration_value`）与额度重置周期（`quota_reset_period`）是两个独立字段。年包不能仅凭标题含有“年”字就假定为“每月重置发放”；“一次付款、按月发放”的计费与展示规则仍属**待逐项业务验收**事项。

## GET /api/subscription/self

无请求体。返回当前账号的扣费优先级偏好与订阅列表。客户端优先读取 `data.all_subscriptions`（包含有效、过期、取消的全部历史订阅，用于准确计算单用户限购次数），若后端仅返回旧版 `data.subscriptions` 则兼容回退，但此时客户端标记 `historyComplete: false` 并拒绝本地生成限购套餐的购买确认单。

| 字段 | 类型 | 必填 | 枚举 / 单位与说明 |
|---|---|---|---|
| `billing_preference` | string | 是 | 扣费优先级枚举：`"subscription_first"`（优先扣套餐）、`"wallet_first"`（优先扣钱包）、`"subscription_only"`（仅扣套餐）、`"wallet_only"`（仅扣钱包） |
| `all_subscriptions[].subscription.id` | integer | 是 | 用户订阅记录正整数 ID |
| `subscription.user_id` | integer | 是 | 所属用户正整数 ID，客户端必须校验等于当前登录账号 `user.id` |
| `subscription.plan_id` | integer | 是 | 关联的套餐正整数 ID |
| `subscription.status` | string | 是 | 订阅状态枚举：`"active"`（生效中）、`"expired"`（已过期）、`"cancelled"`（已取消） |
| `subscription.start_time`、`end_time` | integer | 是 | 订阅生效开始与结束时间戳（**Unix 秒**） |
| `subscription.amount_total`、`amount_used` | integer | 是 | 当前周期总额度与已用额度（均为**内部整数 quota 单位**） |
| `subscription.next_reset_time`、`last_reset_time` | integer | 否 | 下次重置与上次重置时间戳（**Unix 秒**；缺失或未发生重置时为 `0`，`0` 不代表已执行过重置） |
| `subscription.source` | string | 否 | 订阅来源标识，如 `"balance"`（钱包购买）、`"admin"`（后台发放）、`"redeem"`（兑换） |
| `subscription.allow_wallet_overflow` | boolean | 否 | 套餐额度用尽后是否允许回退扣减钱包余额 |

订阅响应结构示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "billing_preference": "subscription_first",
    "all_subscriptions": [
      {
        "subscription": {
          "id": 91,
          "user_id": 42,
          "plan_id": 7,
          "status": "active",
          "start_time": 1792713600,
          "end_time": 1795392000,
          "amount_total": 5000000,
          "amount_used": 10000,
          "next_reset_time": 0,
          "last_reset_time": 0,
          "source": "balance",
          "allow_wallet_overflow": true
        }
      }
    ]
  }
}
```

## GET /api/user/topup/info

无请求体。返回平台充值与支付合规配置，星途客户端在账单总览（`maas:billing`）与生成套餐购买确认单时读取 `data.payment_compliance_confirmed`。

```bash
curl "$PLATFORM_URL/api/user/topup/info" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

响应结构示例（仅列客户端消费的核心字段，其余支付方式列表字段按需忽略）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "payment_compliance_confirmed": true,
    "enable_online_topup": false,
    "min_topup": 1
  }
}
```

| `data` 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `payment_compliance_confirmed` | boolean | 是 | 平台管理员是否已确认支付合规配置；必须为 `true` 时服务端与客户端才允许执行套餐购买 |
| `enable_online_topup` | boolean | 否 | 是否开启在线充值 |
| `min_topup` | integer | 否 | 最小充值数量 |

## POST /api/subscription/balance/pay

使用当前登录账号的钱包余额（`user.quota`）直接扣款购买指定套餐。该操作会立即扣减钱包 quota 并创建生效订阅。

| 请求字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `plan_id` | integer | 是 | 目标套餐正整数 ID（对应 `/api/subscription/plans` 中的 `plan.id`） |

```bash
curl -X POST "$PLATFORM_URL/api/subscription/balance/pay" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"plan_id":7}'
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": null
}
```

**服务端扣款计算与校验规则**：
1. 服务端重新校验 `payment_compliance_confirmed`、套餐启用状态、`allow_balance_pay === true`、单用户限购次数（统计非 `cancelled` 的历史订阅数是否已达到 `max_purchase_per_user`）。
2. 服务端按 `required_quota = ceil(plan.price_amount * quota_per_unit)` 计算需扣减的内部 quota，并在数据库事务中原子扣减用户钱包额度与写入订阅记录。
3. 通用接口失败时可能返回 HTTP 200 且 `success: false`（如 `"余额不足"`、`"该套餐不支持余额支付"`、`"已达到最大购买次数限制"`），调用方必须检查 `success === true`。
4. 扣款成功后客户端立即重新拉取 `/api/user/self` 与 `/api/subscription/self` 刷新展示。若请求发生网络超时或中断，并不代表服务端未扣款，必须先查账核对，严禁自动重试扣款。

### 客户端内部购买确认流程（Electron IPC）

为防止误触扣款，星途客户端在主进程实现了两步确认流程（属于客户端内部 IPC，不是服务端 HTTP 接口）：

| 客户端内部 IPC | 输入参数 | 前置校验与输出 |
|---|---|---|
| `maas:prepareSubscriptionPurchase` | `accountId`、`planId` | 实时拉取 `/api/user/self`、`/api/subscription/plans`、`/api/subscription/self`、`/api/user/topup/info`、`/api/status`；校验 `payment_compliance_confirmed === true`、`plan.currency === "USD"`、`plan.allow_balance_pay === true`、`historyComplete === true`、未超限购且钱包余额充足，生成有效期 120 秒的内存确认单 `confirmationId` 及扣款前后额度预览 |
| `maas:purchaseSubscriptionWithBalance` | `accountId`、`confirmationId` | 核验确认单未过期、账号与会话未切换，且再次拉取最新套餐与余额確認未发生变动后，向平台发送一次 `POST /api/subscription/balance/pay`；返回状态 `succeeded`、`rejected` 或 `unknown` |

當返回 `unknown`（请求已发出但因网络中断、超时或响应异常未能确认结果）时，客户端锁定该确认单，禁止自动重发购买请求，提示用户先刷新钱包与订阅核对扣款结果。

## POST /api/apps/starpath/managed-key

无请求体，受平台关键操作限流保护。按当前登录账号获取或创建星途专属的托管模型 API Key（`tokens.managed = true`，名称固定为 `starpath-managed`）；服务端通过数据库事务与用户行锁保证同一账号并发请求不会创建重复的启用托管令牌。

```bash
curl -X POST "$PLATFORM_URL/api/apps/starpath/managed-key" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

成功响应示例（HTTP 200）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 81,
    "key": "sk-MODEL_KEY_PLACEHOLDER"
  }
}
```

| 响应字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `data.id` | integer | 是 | 平台令牌记录正整数 ID |
| `data.key` | string | 是 | 完整的模型调用密钥（服务端在返回时自动补齐 `sk-` 前缀），客户端须加密存储，严禁写入日志、错误报告或截图 |

失败响应示例（HTTP 500）：

```json
{
  "success": false,
  "code": "MANAGED_KEY_UNAVAILABLE",
  "message": "Unable to prepare managed key"
}
```

**托管令牌特性与客户端合并注意事项**：
- 托管令牌本身设置为 `unlimited_quota: true`、永不过期，不写死分组或模型白名单。默认跟随当前用户分组；部署下述“全局分组策略”配套版本后，也可由管理员统一指定星途请求组。实际调用 `/v1/*` **仍受账号套餐额度、钱包余额和平台鉴权约束**；“令牌不设独立限额”不等于免费调用。
- 托管令牌在平台普通令牌列表（`/api/token/*`）中被隔离过滤，普通用户无法在网页令牌管理页误删或篡改。
- 客户端在登录成功后同步托管密钥并调用 `GET /v1/models` 发现可用模型；退出登录或切换账号时必须立即清除本地托管密钥并禁用平台模型配置。
- **跨模块集成提醒**：账户基础负责 `MaasIpc.ManagedKey`、来源隔离和模型目录刷新。接入账务、UI 或扩展时必须保留 `maasAuth.ts` 的 `managedKey` 端点与 IPC 注册，不用其他模块的旧适配整包覆盖。

## 星途专属 Key 的全局分组策略

> **版本与部署前置（2026-10-26）**：本节对应配套 MaaS M1 分支 `split/maas-pr2-plan-key` 的新增实现，已在独立环境完成代码与接口验证，**尚未合入平台 main，也未加载到共享服务**。当前手册 PR 只交付说明，不会安装后台入口或开放这些接口。部署包含该策略的全部 API 节点及 MaaS 管理前端后，再按本节操作；客户端目录自动刷新还需配套新版账户适配。接口不存在或菜单未出现时先核对版本，不要反复重建 Key。

### 平台、分组、Key、套餐和积分分别负责什么

| 对象 | 职责 |
|---|---|
| MaaS 登录态 | 读取资料、账务、订阅和获取本人托管 Key；不是模型 API Key |
| 星途专属 Key | 平台给每个账号创建或复用的模型凭证；只用于 `/v1/*`，不额外赠送额度 |
| 用户分组 | 默认决定该用户的调用组；统一模式不改此字段，其他普通 Key 保留原有规则 |
| 星途全局分组策略 | 决定所有平台签发的星途专属 Key 在**后续请求**中使用哪个组，包括存量与新增 Key |
| 钱包与套餐 | 平台结算选用的不同额度来源，**同一次请求共用同一个有效调用组**；扣费优先级和余额回退按平台设置 |
| 积分 | 额度的展示换算，不是独立账本，不改变平台实际定价、quota 或扣费比例 |

本批不配置套餐升级/降级组，不设计正式套餐档位、价格或额度。平台商品若已有升降组配置，仍由平台执行：跟随模式会使用变更后的用户组；统一模式继续使用管理员指定的星途组。用户特殊倍率与钱包/订阅结算仍按平台规则，不承诺所有用户最终单价相同。

### 两种模式与管理员权限

| 模式 | 存量与新增星途专属 Key 的行为 |
|---|---|
| `follow_user`：跟随用户分组（默认） | 按各自当前用户组调用；不把首次登录的组名或 `default` 固定写入 Key |
| `fixed`：指定星途统一分组 | 全部使用管理员选定的有效组；不改账户分组、Key 字符串或普通自建 Key |

统一模式是管理员对**星途专属 Key**的明确全员模型组授权，可能改变可用模型范围及分组倍率。它不会放开普通 Key 的组校验，也不会允许用户伪造托管标记。只有仍在使用普通自建 Key 的旧客户端，不属于该策略的管理范围。

权限资源为 **`starpath_keys`**，与插件/技能/MCP 的模块权限分别管理：

| 权限或身份 | 能做什么 |
|---|---|
| Root | 默认查看和管理；可在现有用户管理权限编辑器中授权管理员 |
| `starpath_keys:view` | 只读策略与候选分组，不可保存 |
| `starpath_keys:manage` | 查看并保存策略，包括全员统一分组和切回跟随；仅授予确需此权限的管理员 |
| 未授权管理员、普通用户 | 管理接口返回 `403`；仅隐藏按钮不能代替服务端鉴权 |

普通管理员不能自行授予此项权限。Root 在**用户管理 → 编辑管理员 → 管理权限**中勾选“查看星途密钥分组”或“管理星途密钥分组”；被授权管理员更新登录状态后可见“星途密钥分组”入口。`manage` 已允许读取，不必为保存额外勾选 `view`。

### 设置步骤

1. 先在平台准备目标分组、渠道模型和计价配置；此入口只选择已有分组，不替管理员创建模型、渠道或商品。
2. Root 或获 `manage` 授权的管理员打开**星途密钥分组**（`/starpath-keys`）。
3. 选择“跟随用户分组”，或选择“指定星途统一分组”并选择目标组。不能选不存在、已废弃、空值或 `auto` 作为统一组。
4. 核对对模型范围和后续请求组的影响，勾选“确认应用于全部存量和新增星途专属密钥”，再保存。
5. 用两个不同账号核对模型目录和平台消费记录；切回跟随时，同一把 Key 继续使用各自用户组，无需用户重新填写或重新生成。

服务端保存成功后，本节点后续请求即按策略解析有效组；其他 API 节点复用平台现有配置同步周期（`SYNC_FREQUENCY`，默认 60 秒）。策略持久化在现有 options 表的 `StarpathKeyGroupPolicy`，不批量改 Token、不轮换 Key；保存失败不会发布未持久化的授权，操作日志记录操作人和修改前后策略。

新版星途在原有账户检查（约 5 分钟、聚焦、重新联网/可见）时重新读取目录；同名组内的模型变化也能更新，短时触发合并并串行处理。有运行中会话时推迟客户端配置写入，离线用户下次登录同步。**这不是全员实时推送**，不会主动重启现有流式请求；后续请求仍由服务端按新策略鉴权。

### 管理接口

以下接口均要求管理员登录访问令牌，不接受模型 Key 代替登录态：

| 方法 | 路径 | 权限 |
|---|---|---|
| GET | `/api/apps/starpath/admin/key-policy` | Root，或 `starpath_keys:view` / `starpath_keys:manage` |
| PUT | `/api/apps/starpath/admin/key-policy` | Root，或 `starpath_keys:manage` |

GET 及成功 PUT 的 `data` 包含 `policy`（`mode`、`group`）、`valid`、`groups`（候选组列表）和 `sync_seconds`（平台节点同步周期）。PUT 示例，`starpath` 仅为示例名称，须替换为实际已配置的组：

```json
{"mode":"fixed","group":"starpath"}
```

切回跟随：

```json
{"mode":"follow_user","group":""}
```

无权限返回 `403`，非法策略/分组返回 `400`，持久化失败返回 `500`。已保存策略损坏或目标组被移除时，请求会阻止，不自动退回可能权限更高的用户组；管理 GET 返回 `valid: false`、`policy: null`，由授权管理员明确选择有效策略后重新保存。

## 额度与展示单位换算规则

平台内部额度（quota）、商品货币价格与界面展示数值必须严格区分：
1. **内部 quota 转美元（USD）**：`usd = quota / quota_per_unit`。
2. **内部 quota 按 `quota_display_type` 展示**：
   - `"TOKENS"`：直接展示内部整数 `quota`。
   - `"USD"`：展示 `$` + `quota / quota_per_unit`。
   - `"CNY"`：展示 `¥` + `(quota / quota_per_unit) * usd_exchange_rate`。
   - `"CUSTOM"`：展示 `(quota / quota_per_unit) * custom_currency_exchange_rate` + `custom_currency_symbol`。
3. **套餐价格转所需内部 quota（购买预览）**：客户端采用十进制精确乘法计算 `ceil(price_amount * quota_per_unit)`（向上取整），与服务端扣款算法保持一致，禁止使用二进制浮点数直接相乘截断或写死 `500000` 常数。
