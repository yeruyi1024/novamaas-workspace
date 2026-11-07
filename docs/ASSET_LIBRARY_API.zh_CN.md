# 素材库与火山 Action API 兼容说明

平台素材库同时提供面向控制台的文件上传接口，以及面向下游客户的火山方舟 Action API 兼容接口。下游侧使用平台签发的 AK/SK，不会接触上游资源账号的凭据。平台素材 ID 在视频转发时，按选中的渠道所关联的资源账号映射为对应的上游素材 ID。

## 多渠道共享上游素材库

管理入口以资源账号为单位显示独立的正方形素材库卡片，厂商名称作为卡片内的标识。同一账号关联的多个渠道在卡片内按列表展示，数量较多时在卡片内滚动；同一厂商的不同资源账号分别显示为不同卡片。「添加资源账号」位于页面全局入口，卡片内仅管理该账号的渠道、参数、同步及日志。厂商名称只用于归类；资源权限以具体账号、接口地址和项目等配置为准。同一厂商的不同账号、不同 AK/SK 或项目应分别配置，不能仅凭厂商名称自动共用素材。

每个资源账号沿用原渠道配置中的协议、地址、AK/SK 或 Bearer Key、Region、Service、API Version、ProjectName 和 QPM 参数。凭据加密保存，列表只返回脱敏提示；编辑时留空可保留原凭据，名称、QPM 和启停不需要读取原凭据，因此即使旧凭据无法在当前服务器解密仍可修改。账号一旦拥有上游素材，就不能再改变接口地址、项目范围或凭据（返回 409），应新建资源账号；尚无上游素材时可以直接更换。新建账号时，若协议、地址、范围和凭据与已有账号完全一致，会直接复用已有账号，不重复创建。

仅 DoubaoVideo（54）和 Volc Native（61）渠道可关联素材库。一个资源账号可关联多个渠道，每条渠道只能归属一个资源账号；停用关联仍保留归属。已归属其他账号的渠道在表单中不可选，接口也会拒绝直接换绑并返回 409；需要更换账号时，先从原账号解除关联，再关联新账号。关联与解除通过 `PUT /channels/{channel_id}` 逐条完成（`library_id` 为 0 表示解除），不需要读取账号凭据。

已知限制：修改渠道自身的 API Key 或地址不会自动解除素材账号关联。若渠道改指向另一个上游租户，需先在「管理渠道」中解除关联，否则上游会因素材 ID 不属于该租户而拒绝请求。

一个全局素材 ID 对应每个已关联资源账号的一份上游副本。同账号下的多个定价、折扣或路由渠道共用该副本、状态轮询和请求频率预算。例如 10 条渠道分属 3 个资源账号时，单次上传最多创建 3 份上游素材。生成时只查找所选渠道关联账号的可用副本，不会退回到其他账号，也不改变用户之间的素材隔离。

账号一旦变为可用，已有素材会自动补同步到它：新账号关联第一个启用的渠道时，或停用的账号重新启用时，平台会为该账号排队所有进行中或可用的素材，已存在的副本（包括失败的）不会被重试，失败重试仍在同步详情中单独处理。素材跟着账号走：同账号再关联更多渠道不会增加副本。「补同步缺失素材」保留为手动补偿入口。

账号可以随时停用（配置里的「启用素材同步」），停用只停止写入：不再为它排队新素材，也不会补同步，它名下进行中的上传和轮询暂停，仅继续处理删除；它已经同步好的素材仍可被引用，停用期间新传的素材在该账号上不可用，直到重新启用。重新启用后会自动补同步停用期间错过的素材，并为未完成且没有有效租约的任务重新给予 5 分钟同步预算，保留上游 ID；已失败的副本不会被自动重试。重新关联第一个启用渠道时同样恢复暂停任务。要让某个渠道不再使用素材，请停用该渠道的关联。同步管理默认显示厂商和资源账号概览，点击账号后再加载素材明细。

### 旧数据迁移与删除

升级时先将原渠道配置及副本映射到独立资源账号（每条旧渠道配置一个账号，名称沿用渠道名称），保留全局 ID、上游 ID、任务记录、失败信息和请求日志，重复执行不会新增数据。启动时只记录每个账号的上游身份（协议、地址、范围、AK 及解密后的 SK/Key 的摘要），不会合并任何账号；之后新建同凭据账号会直接复用已有账号。

重复账号由管理员在「素材库」页点击「合并相同账号」（`POST /libraries/merge`）后合并，只合并身份完全一致的账号，解不开凭据的账号不参与：
- 合并到持有最多已同步副本的账号（相同则取编号小的），账号名取这一组里编号最小的账号名，之后可在配置里自行改名；启用状态取其中任一启用即启用，QPM 取较小值。
- 这些账号的渠道全部关联到保留的账号。保留账号缺少的已同步副本和分组副本直接挪过去；若保留账号只有未完成或失败的同步记录，则与旧账号的成功副本交换归属，保留两条记录及其上游 ID。审核拒绝和删除意图不会被覆盖。上游素材 ID 在同一凭据下有效，不会重新上传。
- 保留账号已有成功副本的素材，旧账号里重复的副本原样保留，不删除上游；失败、被拒等状态的副本不会被重试或清理。旧账号没有渠道后在页面折叠区显示为「未关联 / 历史素材库」，仍保留凭据，因此这些副本可随素材一起删除。历史账号的请求共用保留账号的 QPM 预算，并随保留账号的 QPM 调整；完全没有副本的旧账号会被移除。
- 解不开凭据的账号（例如本机密钥与原环境不一致）不参与合并，结果里会单独统计跳过的账号数，不要把“合并 0 组”当成没有重复；有上传正在进行的分组先跳过，稍后再点即可；合并后保留账号缺少的素材会自动排队同步。
- 合并不会删除任何上游素材。重复的旧副本只有在平台素材被删除时才会一并删除，目前没有单独的清理入口。

升级时应先停止旧版本素材同步进程，完成数据库迁移并更新全部节点后，再启用新版同步，避免旧版仍按渠道重复上传。

用户删除全局素材时，统一排队删除所有账号（包括已无渠道的旧账号）里的副本，所有副本删除确认后，才继续处理源文件删除。

以上管理接口仅限超级管理员，路径前缀为 `/api/asset-library/admin`：

| 接口 | 用途 |
| --- | --- |
| `GET/POST /libraries`、`PUT /libraries/{id}` | 查询、新增或编辑资源账号 |
| `POST /libraries/{id}/test` | 测试已保存或正在编辑的参数 |
| `POST /libraries/{id}/sync` | 只补充缺失副本 |
| `POST /libraries/merge` | 合并身份完全一致的账号，返回已合并、推迟的组数和因凭据无法读取而跳过的账号数 |
| `PUT /channels/{channel_id}` | 请求体 `{"library_id": 账号 ID, "enabled": true}` 关联渠道；`library_id` 为 0 解除关联 |
| `GET /sync-jobs?library_id={id}` | 查看资源账号的同步明细 |

## 下游 AK/SK

用户登录控制台后，进入「素材库」，点击页面右上角的「API 接入」即可创建和吊销下游 AK/SK。弹窗会同时展示当前站点的官方 SDK Endpoint、`/api/v3/` 兼容请求地址、签名参数和已支持的 Action，便于直接交付给下游客户。

创建时只需填写密钥用途名称。创建成功后必须立即复制 Access Key ID 和 Secret Access Key；Secret Access Key 只显示一次，关闭提示后无法再次查看。如有遗失，应吊销旧密钥并重新创建，不应要求管理员从数据库中导出密文。

也可以通过以下控制台接口自动化管理素材库访问密钥：

- `GET /api/asset-library/access-keys`
- `POST /api/asset-library/access-keys`，请求体为 `{"name":"automation"}`
- `DELETE /api/asset-library/access-keys/{id}`

每个用户最多可同时拥有 5 个有效密钥。创建响应只返回一次 `secret_access_key`，后续列表只显示 AK 和 SK 尾号；删除采用禁用状态保留审计记录。SK 使用 AES-GCM 加密保存，密钥来源依次为 `STORAGE_CREDENTIAL_ENCRYPTION_KEY`、`CRYPTO_SECRET`、`SESSION_SECRET`。生产环境应固定配置第一项，并纳入备份和轮换方案；缺少可用主密钥时拒绝创建 AK/SK。

## 兼容路径与 Action

与火山官方 SDK 一样，可将 API Endpoint 配置为平台域名并请求根路径：

```text
POST https://gateway.example.com/?Action=CreateAsset&Version=2024-01-01
```

同时提供 `POST /api/v3/?Action=...&Version=2024-01-01` 作为网关部署兼容别名。签名必须使用客户端实际请求的路径，根路径和 `/api/v3/` 的签名不能混用。

当前支持以下 Action：

- `CreateAssetGroup`、`ListAssetGroups`、`GetAssetGroup`、`UpdateAssetGroup`、`DeleteAssetGroup`
- `CreateAsset`、`ListAssets`、`GetAsset`、`UpdateAsset`、`DeleteAsset`

请求和响应采用火山 `ResponseMetadata` / `Result` 信封，版本固定为 `2024-01-01`，Service 为 `ark`，Region 为 `cn-beijing`。`ProjectName` 作为协议兼容字段接受，平台租户隔离仍以 AK 所属用户为准，响应中的项目名为 `default`。

`CreateAsset` 接收 `GroupId`、`URL`、`AssetType` 和可选 `Name`。平台先使用 SSRF 防护下载公网 HTTP/HTTPS 素材，按存储策略校验 MIME、类型和最大文件大小，写入平台永久对象存储，再返回平台素材 ID；各已关联资源账号的副本继续由后台同步。存在已启用的上游素材账号和关联渠道时，新素材先返回 `Status=Processing`，首次确认上游可用后改为 `Status=Active`；没有启用上游关联时仍直接返回 `Active`。素材 URL 导入完成前客户端请求会保持连接，因此下游应设置与大文件上传相匹配的超时，并使用自己的幂等控制避免超时重试产生重复素材。

列表接口支持官方的页码分页，也支持 `MaxResults` / `NextToken`：

- `PageNumber` 从 1 开始，`PageSize` 最大 100。
- `MaxResults` 最大 100；存在下一页时响应返回仅对本平台有效的 `NextToken`。
- `Filter.Name` 模糊搜索名称，`Filter.GroupIds` 精确筛选素材组，`ListAssets` 还接受 `Filter.Statuses`。
- `SortBy` 支持 `CreateTime`、`UpdateTime`，素材列表额外支持 `GroupId`；`SortOrder` 支持 `Asc`、`Desc`。

## AK/SK 签名

下游签名沿用火山 HMAC-SHA256 V4 结构：

```text
Authorization: HMAC-SHA256 Credential=<AK>/<YYYYMMDD>/cn-beijing/ark/request, SignedHeaders=content-type;host;x-content-sha256;x-date, Signature=<hex>
X-Date: 20260922T083045Z
X-Content-Sha256: <SHA256(request-body)>
Content-Type: application/json
```

平台会重新计算请求体哈希、Canonical URI、排序后的 Canonical Query、Canonical Headers 和四级派生签名，并使用常量时间比较验签。`X-Date` 与凭据日期必须一致，服务器只接受时间偏差不超过 5 分钟的请求。反向代理必须保留客户端签名时使用的 `Host`，并保持请求路径与 Query 不变。

## 素材库大列表

控制台素材列表按服务端分页，每页最多加载 40 条；名称、素材 ID、MIME 类型和上传者搜索在数据库侧完成。页面只保留当前页卡片，图片使用浏览器懒加载，预览签名 URL 仅在卡片接近可视区域时请求，因此素材总量超过 1,000 时不会在首屏创建同等数量的 DOM、媒体请求或预览签名请求。

素材卡片展示上传者和上传时间。管理员查看全局范围时可以按上传者检索；普通用户仍只能看到自己的素材。数据库查询以素材归属、状态和创建时间索引为基础，部署到生产前仍应使用目标数据库和真实对象存储做容量、带宽与分页延迟压测。

## 素材状态 Webhook

### 素材状态查询与图片预览

AK/SK 的 `GetAsset` 与 `ListAssets` 返回素材元数据和审核状态；查询成功为 HTTP 200，素材是否可用需看 `Status`。审核拒绝返回 `Status=Failed` 和 `FailureReason`，不会因审核失败隐藏素材。

图片 URL 可生成时，额外返回 `PreviewStatus=Available` 和 `URLExpiresAt`（Unix 秒），客户可使用 `URL` 预览，过期后重新查询。预览签名或存储发生异常时，仍返回素材元数据和审核状态，`URL` 为空，`PreviewStatus=Unavailable`、`PreviewErrorCode=preview_unavailable`，不返回内部存储错误。一条素材预览失败不会使整页列表失败。素材不存在、无权限或元数据查询失败仍按原错误响应处理。

```json
{
  "Result": {
    "Id": "asset-example-001",
    "Status": "Failed",
    "FailureReason": "sensitive_content",
    "URL": "",
    "PreviewStatus": "Unavailable",
    "PreviewErrorCode": "preview_unavailable"
  }
}
```

以上为省略其他字段的示例。`PreviewStatus` 只表示图片预览是否可用，与审核状态独立，不能据此决定是否生成视频。`CreateAsset` 的 AK/SK 响应仍仅返回 `Id`；应随后调用 `GetAsset` 或接收事件确认状态。

### 控制台配置与事件投递

下游客户使用素材所属账号登录控制台，在「素材库 → Webhooks」填写自己的公网 HTTPS 回调地址、选择 `asset.active` 和/或 `asset.failed`，并可发送测试事件。普通用户即可配置自己的端点，无需管理员权限或素材 AK/SK；素材上传、查询等 Action API 仍使用原有 AK/SK。平台集中轮询上游素材状态；首次确认可用，或后续复审将素材从可用改判为失败时，平台先更新本地素材状态，再向订阅端点发送事件。Webhook 只负责提醒，`GetAsset` / `ListAssets` 返回的本地状态始终是最终依据。

控制台页面使用以下受控制台用户鉴权保护的接口管理端点，不提供素材 AK/SK Action 管理入口：

- `GET /api/asset-library/webhook-endpoints`
- `POST /api/asset-library/webhook-endpoints`
- `DELETE /api/asset-library/webhook-endpoints/{id}`
- `POST /api/asset-library/webhook-endpoints/{id}/test`

端点支持新增、查看、测试和删除。如需更换回调地址或订阅事件，请删除旧端点后重新添加；删除后旧端点不再继续投递，重新添加不会补发历史事件，切换期间可用 `GetAsset` / `ListAssets` 对账。

创建请求示例：

```json
{
  "name": "production",
  "url": "https://customer.example.com/webhooks/assets",
  "event_types": ["asset.active", "asset.failed"]
}
```

每个用户最多可有 5 个有效端点。端点不生成或返回签名密钥，配置回调无需设置凭据加密主密钥；素材 AK/SK 和上游凭据的加密要求保持不变。回调地址在投递前同时经过 URL 校验和 SSRF 防护。平台直连投递不会跟随重定向；启用请求 Worker 的部署应让 Worker 同样把 3xx 视为失败。

平台直接通过 HTTPS POST 发送 JSON，不发送 `webhook-signature`，客户无需配置签名密钥或验签。`webhook-id` 用于去重，`webhook-timestamp` 为投递时间（Unix 秒）：

```http
POST /webhooks/assets HTTP/1.1
Content-Type: application/json
webhook-id: wh_...
webhook-timestamp: 1790580000
```

```json
{
  "id": "evt_...",
  "object": "event",
  "type": "asset.failed",
  "created_at": 1790580000,
  "data": {
    "id": "asset-20260923163433-giwbv",
    "group_id": "group-20260923163432-cwlhg",
    "status": "Failed",
    "failure_reason": "sensitive_content",
    "updated_at": 1790580000,
    "state_version": 2
  }
}
```

下游应校验事件格式、按 `webhook-id` 去重，可靠保存事件后快速返回 2xx，再异步处理业务。非 2xx 或网络错误会指数退避重试，最长保留 72 小时；同一投递的 `webhook-id` 和请求体在重试时保持不变。未验签的通知无法通过密码学确认来源；客户可用现有 AK/SK 调用 `GetAsset` 核对状态。

收到 `asset.failed` 后，素材 ID 不变，下游应把该 ID 标记为失败/不可用，并阻止新的生成任务继续引用它；不要删除或复用这个 ID。平台会作废排队中的旧 `asset.active`，但已发出的请求仍可能延迟到达。Webhook 可能重复或乱序，下游必须按素材 ID 在数据库中原子比较 `data.state_version`，仅在收到的版本大于已保存版本时更新状态。`asset.active` 的版本为 1，终态 `asset.failed` 的版本为 2；同一素材失败后不会恢复，重新上传使用新 ID。不要仅比较秒级 `updated_at`，因为两次状态变化可能发生在同一秒。缺少版本的历史事件应通过 `GetAsset` 查询对账，不直接覆盖本地状态。
