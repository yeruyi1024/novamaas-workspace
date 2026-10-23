# 技能上传、下载与安装

托管技能制品（Hosted Skill Artifacts）使用 `schema_version: 3` 协议。标准工作流为：
1. **Root 管理员**配置独立对象存储策略（`/storage`）。
2. **技能模块授权管理员**上传本地 ZIP（`/uploads`）或从受信任的 HTTPS / GitHub Commit 导入（`/imports`），生成处于 `staged` 状态的制品版本。
3. 管理员核对服务端自动解析出的 `SKILL.md` 元数据、文件数和依赖后，调用确认接口（`/releases/:release_id/confirm`）锁定 `skill_id` 与固定 `version`，状态转为 `ready`。
4. 管理员将 `ready` 状态的 `release_id` 写入 v3 技能目录草稿并发布（`/catalog/draft` → `/catalog/publish`）。
5. **星途用户**收到目录通知并主动点击确认安装或更新后，客户端向平台换取短期签名下载描述（`/v3/skills/releases/:release_id/download`），从对象存储下载 ZIP，完成大小、SHA-256、路径安全与身份核验后解压安装，并向平台汇报安装状态（`/v3/skills/installations`）。

所有平台接口均使用 `Authorization: Bearer $ACCESS_TOKEN` 认证；但向对象存储临时签名 URL 下载 ZIP 文件时**严禁转发平台 `Authorization` 请求头**。

## 后台技能制品管理接口索引

公共前缀：`/api/apps/starpath/admin/skill-artifacts`。除 `/storage` 仅限 Root（`RootAuth`）外，其余管理接口依据 HTTP 方法校验 `starpath_skills` 模块的 `read` 或 `manage` 权限。

| 方法 | 相对路径（前缀 `/api/apps/starpath/admin/skill-artifacts`） | 所需权限 | 用途 |
|---|---|---|---|
| GET | `/storage` | 仅 Root（`RootAuth`） | 读取技能制品存储策略、可用对象存储档案列表（`profiles`）及 `configured` 就绪标识 |
| PUT | `/storage` | 仅 Root（`RootAuth`） | 保存技能制品存储策略与安全限制 |
| GET | `/status` | `starpath_skills:read` | 查询当前制品存储是否已启用且就绪（返回 `configured`、`upload_available`、`provider`、`limits`、`import_available`） |
| POST | `/uploads` | `starpath_skills:manage` | 通过 `multipart/form-data` 上传技能 ZIP 归档（创建 `staged` 版本） |
| POST | `/imports` | `starpath_skills:manage` | 从允许域名的 `https_zip` 或 `github_commit` 导入技能归档（创建 `staged` 版本） |
| GET | `/releases?limit=50` | `starpath_skills:read` | 查询最近上传/导入的制品版本列表（`limit` 可选 `1`–`100`，默认 `50`，返回 `data.items`） |
| GET | `/releases/:release_id` | `starpath_skills:read` | 查询单个制品版本详情 |
| POST | `/releases/:release_id/confirm` | `starpath_skills:manage` | 确认 `staged` 制品版本的 `skill_id`、固定 `version`、名称与描述，转为 `ready` |
| DELETE | `/releases/:release_id` | `starpath_skills:manage` | 请求删除未被引用（未被草稿、历史发布或客户端安装回执引用）的制品版本 |
| GET | `/catalog` | `starpath_skills:read` | 读取 v3 技能目录当前草稿与已发布快照（`schema_version: 3`） |
| PUT | `/catalog/draft` | `starpath_skills:manage` | 保存 v3 技能目录完整草稿（校验 `hosted_artifact` 引用的 `artifact_release_id` 必须处于 `ready` 状态且 `version`/`sha256` 完全一致） |
| POST | `/catalog/publish` | `starpath_skills:manage` | 发布当前 v3 技能目录草稿，并锁定所引用制品版本的发布引用关系 |
| GET | `/installations?release_id=RELEASE_ID&limit=50` | `starpath_skills:read` | 查询指定 `release_id`（必填）最近的客户端安装回执列表（`limit` 可选 `1`–`100`，默认 `50`） |

> **注意**：`/releases` 与 `/installations` 是带条数上限（最多 100 条）的近期记录查询接口，非法 `limit` 参数返回 `400 SKILL_LIST_LIMIT_INVALID`。上传/导入成功只代表生成了待确认的 `staged` 制品，必须经过 `confirm` 并写入 `catalog` 发布后，客户端才能看到并下载。

## 上传、导入、确认与删除制品

### POST /api/apps/starpath/admin/skill-artifacts/uploads

`Content-Type: multipart/form-data`，受平台进程级并发限制保护（上传与导入共享 2 个处理槽位，槽位占满返回 `429 SKILL_UPLOAD_BUSY`）。表单严格只允许包含 `file`（ZIP 文件）和 `reason`（非空，1–500 字符）两个字段，不要手动硬编码 multipart boundary。

```bash
curl -X POST "$PLATFORM_URL/api/apps/starpath/admin/skill-artifacts/uploads" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F 'file=@./example-skill.zip' \
  -F 'reason=上传示例技能供核对'
```

### POST /api/apps/starpath/admin/skill-artifacts/imports

请求体上限 8 KiB，JSON 严格包含 `reason`（1–500 字符）与 `source` 对象。`source.kind` 支持两种枚举：

1. **`https_zip`（固定 HTTPS 归档地址）**：`url` 必须为 `https://`，域名必须命中存储策略的 `allowed_import_hosts` 白名单，且服务端 DNS 解析禁止指向内网/环回/元数据私有 IP。

```json
{
  "reason": "导入固定技能包",
  "source": {
    "kind": "https_zip",
    "url": "https://downloads.example.invalid/example-skill-1.0.0.zip"
  }
}
```

2. **`github_commit`（固定 GitHub Commit 归档）**：`codeload.github.com` 必须加入 `allowed_import_hosts` 白名单；`commit` 必须为 7–40 位十六进制固定提交哈希（严禁使用 `main`/`master` 等浮动分支名）。

```json
{
  "reason": "从仓库固定提交导入技能",
  "source": {
    "kind": "github_commit",
    "repository": "example-org/example-skills",
    "commit": "0123456789abcdef0123456789abcdef01234567",
    "subpath": "skills/example-skill"
  }
}
```

### POST /api/apps/starpath/admin/skill-artifacts/releases/:release_id/confirm

请求体上限 8 KiB。对处于 `staged` 状态的 `:release_id` 进行身份与固定版本确认，确认后状态变为 `ready`，且 `(skill_id, version)` 组合全局唯一、不可覆盖（若已存在相同 `skill_id` 和 `version` 的 `ready` 制品，返回 `409 SKILL_VERSION_CONFLICT`）。

| 请求字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `reason` | string | 是 | 确认理由（去首尾空白后 1–500 字符） |
| `skill_id` | string | 是 | 稳定技能标识（1–128 字符，需与目录条目 `id` 一致；若归档内声明了 ID，必须与之匹配） |
| `version` | string | 是 | 固定版本号（1–128 字符，禁止 `latest` 或范围表达式；若归档内声明了固定版本，必须与之匹配） |
| `name` | string | 是 | 确认后的技能名称（去首尾空白后 1–120 字符） |
| `description` | string | 是 | 确认后的技能描述（允许空字符串 `""`，最多 2000 字符） |

```json
{
  "reason": "核对归档内容后确认",
  "skill_id": "example-skill",
  "version": "1.0.0",
  "name": "示例技能",
  "description": "演示技能制品流程"
}
```

### 制品视图（`StarpathSkillReleaseView`）响应字段

`POST /uploads`、`POST /imports`、`POST /releases/:release_id/confirm`、`GET /releases/:release_id` 及 `DELETE /releases/:release_id` 均返回如下制品视图对象：

| `data` 字段 | 类型 | 说明 |
|---|---|---|
| `release_id` | string | 制品版本唯一标识（如 `"srel_..."`） |
| `status` | string | 制品状态枚举：`"staged"`（已上传待确认）、`"ready"`（已确认可发布）、`"failed"`（失败）、`"delete_pending"`（待后台清理对象存储）、`"deleted"`（已删除） |
| `skill_id`、`version` | string | 技能稳定 ID 与固定版本号（`staged` 阶段若未声明可能为空，确认后必非空） |
| `name`、`description` | string | 技能名称与描述 |
| `sha256` | string | 服务端生成的规范化 ZIP 归档的 64 位小写十六进制 SHA-256 摘要 |
| `size` | integer | 规范化 ZIP 字节大小 |
| `format` | string | 固定为 `"zip"` |
| `root_path` | string | 归档内技能根目录名（含 `SKILL.md`） |
| `file_count`、`expanded_bytes` | integer | 归档内文件总数与解压后总字节数 |
| `dependencies` | string[] | 解析出的依赖列表 |
| `parsed` | object | 归档内 `SKILL.md` frontmatter 解析出的原始值（`skill_id`、`version`、`name`、`description`、`dependencies`） |
| `source` | object | 来源追溯信息（`kind`、`url`、`repository`、`commit`、`subpath`、`original_sha256`） |
| `created_by`、`confirmed_by`、`deleted_by` | integer | 对应操作管理员的平台用户 ID |
| `created_at`、`confirmed_at`、`published_at`、`delete_requested_at`、`deleted_at` | integer | 各生命周期时间戳（**Unix 秒**；未发生时为 `0`） |
| `last_error_code`、`last_error_message` | string | 最近一次操作或清理异常记录 |

### DELETE /api/apps/starpath/admin/skill-artifacts/releases/:release_id

请求体上限 4 KiB，JSON 严格只包含 `reason`（1–500 字符）：

```json
{
  "reason": "清理未发布的测试上传版本"
}
```

**删除引用保护规则**：若该 `release_id` 正在被当前草稿引用、曾被任何历史已发布目录引用（`publication`），或已有任何客户端安装回执引用（`installation`），服务端一律拒绝删除并返回 `409 SKILL_RELEASE_REFERENCED`。未被引用的制品被标记为 `delete_pending`，由后台清理协程删除对象存储文件后转为 `deleted`。**平台侧删除制品版本不会触发任何客户端卸载本机已安装的技能副本。**

## POST /api/apps/starpath/v3/skills/releases/:release_id/download（客户端获取签名下载描述）

请求体上限 4 KiB。当用户在星途客户端点击确认安装或更新某项托管技能时，客户端调用该接口。服务端校验当前登录账号是否被 `publication_id` 授权访问 `item_id` 对应的 `:release_id`，校验通过后返回**带有效期的对象存储临时签名下载 URL 与校验元数据**（接口本身返回 JSON，不直接返回 ZIP 二进制流）。

| 请求字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `item_id` | string | 是 | 目录条目 `id`（1–128 字符） |
| `publication_id` | string | 是 | 当前技能目录的 `publication_id`（如 `"skills:3"`，最长 64 字符） |

```bash
curl -X POST "$PLATFORM_URL/api/apps/starpath/v3/skills/releases/release-example/download" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"item_id":"example-skill","publication_id":"skills:3"}'
```

成功响应示例（HTTP 200）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "schema_version": 3,
    "release_id": "release-example",
    "skill_id": "example-skill",
    "version": "1.0.0",
    "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "size": 2048,
    "format": "zip",
    "root_path": "example-skill",
    "source": {
      "kind": "upload",
      "url": "",
      "repository": "",
      "commit": "",
      "subpath": "",
      "original_sha256": ""
    },
    "download": {
      "url": "https://storage.example.invalid/signed-download-placeholder",
      "expires_at": 1792713900
    }
  }
}
```

| 响应字段 | 类型 | 必填 | 客户端校验与使用规范 |
|---|---|---|---|
| `schema_version` | integer | 是 | 必须严格等于 `3` |
| `release_id`、`skill_id`、`version` | string | 是 | 必须与用户点击确认的目录条目 `artifact_release_id`、`id`、`version` 完全一致，不一致报 `IDENTITY_MISMATCH` |
| `sha256` | string | 是 | 64 位小写十六进制摘要；必须与目录条目中的 `sha256` 及下载完成后的 ZIP 实际 SHA-256 完全一致，不一致报 `HASH_MISMATCH` |
| `size` | integer | 是 | ZIP 文件精确字节数（正整数，客户端上限 50 MiB）；下载流超过该大小或最终字节数不符立即中止 |
| `format` | string | 是 | 固定为 `"zip"` |
| `root_path` | string | 是 | 归档根目录名称（匹配 `^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`） |
| `source` | object | 是 | 制品来源追溯元数据（`kind` 为 `"upload"`、`"https_zip"` 或 `"github_commit"`） |
| `download.url` | string | 是 | 对象存储临时签名 `GET` 下载地址（生产环境要求 `https://`；请求该地址时**不带平台 `Authorization` 头**，且拒绝跨源重定向） |
| `download.expires_at` | integer | 是 | 签名地址过期时间戳（**Unix 秒**）；不可缓存为永久链接，过期后需重新调用本接口换取新链接 |

客户端下载 ZIP 后，必须在临时目录完成解压前安全检查（拒绝符号链接、拒绝 `..` 路径穿越、检查单文件/总解压大小/文件数/压缩比上限、校验根目录 `SKILL.md` 存在且声明的 `id`/`version` 未冲突），全部通过后才原子替换或写入本机技能目录；若安装或加载失败，自动回滚保留旧版本。

## POST /api/apps/starpath/v3/skills/installations（客户端上报安装状态）

请求体上限 8 KiB，严格只接收以下 8 个字段（用户身份由服务端从 `Authorization: Bearer $ACCESS_TOKEN` 解析，禁止在请求体中发送 `user_id`）。

| 请求字段 | 类型 | 必填 | 约束与说明 |
|---|---|---|---|
| `operation_id` | string | 是 | 单次安装/更新操作的稳定幂等标识（1–128 字符，匹配 `^[a-zA-Z0-9][a-zA-Z0-9._:-]{0,127}$`） |
| `device_id` | string | 是 | 客户端本机设备随机标识（1–128 字符，保存在本机 `hosted-skill-device-id-v1`） |
| `release_id` | string | 是 | 目标制品版本 ID（必须存在于平台数据库，否则返回 `404 SKILL_RELEASE_NOT_FOUND`） |
| `publication_id` | string | 是 | 触发本次安装的技能目录 `publication_id`（必须以 `"skills:"` 开头，最长 64 字符） |
| `item_id` | string | 是 | 目录条目 `id`（1–128 字符） |
| `actual_version` | string | 是 | 本机当前实际生效版本；当 `status` 为 `"started"`、`"downloaded"`、`"installed"`、`"needs_configuration"` 时必填固定版本号，当 `status` 为 `"failed"` 或 `"rolled_back"` 时允许为空字符串 `""` |
| `status` | string | 是 | 安装阶段状态枚举（见下表） |
| `error_code` | string | 是 | 当 `status` 为 `"started"`、`"downloaded"`、`"installed"` 时**必须为空字符串 `""`**；当 `status` 为 `"failed"`、`"rolled_back"`、`"needs_configuration"` 时**必须为下表允许的机器错误码** |

```json
{
  "operation_id": "install-example-001",
  "device_id": "device-example",
  "release_id": "release-example",
  "publication_id": "skills:3",
  "item_id": "example-skill",
  "actual_version": "1.0.0",
  "status": "installed",
  "error_code": ""
}
```

成功响应示例（HTTP 200）：

```json
{
  "success": true,
  "message": "",
  "data": {
    "operation_id": "install-example-001",
    "device_id": "device-example",
    "release_id": "release-example",
    "status": "installed",
    "updated_at": 1792713800
  }
}
```

| `status` 枚举值 | 含义 | `error_code` 要求 |
|---|---|---|
| `started` | 用户已确认，客户端开始下载安装流程 | 必须为 `""` |
| `downloaded` | ZIP 已下载并通过初步哈希校验，尚未完成安装加载 | 必须为 `""` |
| `installed` | 解压、校验与宿主加载全部成功 | 必须为 `""` |
| `failed` | 下载、校验或安装失败（未发生旧版回滚或无旧版可回滚） | 必须填写允许的错误码 |
| `rolled_back` | 更新新版本失败，已自动恢复到更新前的本机旧版本 | 必须填写允许的错误码 |
| `needs_configuration` | 缺少前置依赖或需用户额外配置，未能直接完成安装 | 必须填写允许的错误码（如 `DEPENDENCY_REQUIRED`） |

**允许的 `error_code` 枚举值（共 10 个）**：
`DOWNLOAD_FAILED`、`HASH_MISMATCH`、`UNSAFE_ARCHIVE`、`IDENTITY_MISMATCH`、`DEPENDENCY_REQUIRED`、`INSTALL_FAILED`、`LOAD_FAILED`、`ROLLBACK_FAILED`、`CANCELLED`、`CONFLICT`。

## 独立对象存储策略配置（`GET / PUT /admin/skill-artifacts/storage`）

`PUT /api/apps/starpath/admin/skill-artifacts/storage` 仅限 Root 调用，请求体上限 64 KiB，必须完整提交以下全部 11 个字段（先调用 `GET /storage` 获取当前 `revision` 和可选对象存储档案 `profiles`，再修改提交；严禁把对象存储 AccessKey/SecretKey 写进技能目录）：

| 请求字段 | 类型 | 必填 | 取值范围与默认值说明 |
|---|---|---|---|
| `expected_revision` | integer | 是 | 乐观并发版本号，等于最新 `GET /storage` 返回的 `revision`（初始为 `0`） |
| `reason` | string | 是 | 修改原因（去首尾空白后 1–500 字符） |
| `enabled` | boolean | 是 | 是否启用星途技能制品存储 |
| `storage_profile_id` | integer | 是 | 平台已配置的对象存储档案 ID（未绑定时为 `0`；`enabled: true` 时必须指向存在的有效档案） |
| `object_prefix` | string | 是 | 对象存储 Key 前缀（长度 1–180 字符，默认 `"starpath/extensions"`，每段仅允许字母数字、下划线与连字符，段间单个 `/` 分隔，不接受首尾 `/` 或点号） |
| `max_file_bytes` | integer | 是 | 单个上传/导入 ZIP 文件最大字节数，范围 `1` 至 `52428800`（50 MiB，默认 `20971520` 即 20 MiB） |
| `max_expanded_bytes` | integer | 是 | 解压后总字节数上限，范围 `>= max_file_bytes` 至 `262144000`（250 MiB，默认 `104857600` 即 100 MiB） |
| `max_files` | integer | 是 | 归档内最大条目数，范围 `1` 至 `10000`（默认 `2000`） |
| `max_compression_ratio` | integer | 是 | 最大允许压缩比（防 ZIP 炸弹），范围 `1` 至 `1000`（默认 `200`） |
| `signed_url_ttl_seconds` | integer | 是 | 临时下载签名 URL 有效期（**秒**），范围 `60` 至 `900`（默认 `300`） |
| `allowed_import_hosts` | string[] | 是 | 允许远端导入的 HTTPS 域名白名单（0–30 个小写域名，禁止填 IP、`localhost` 或 `.local`；使用 `github_commit` 导入时须包含 `"codeload.github.com"`） |

`GET /storage` 响应中除上述配置字段外，还额外返回 `profiles`（可选存储档案摘要列表）与布尔值 `configured`（仅当 `enabled === true` 且绑定的存储档案启用、凭据完整时为 `true`）。注意 `configured: true` 仅代表配置项检查通过，真实 OSS 上传、签名与客户端下载链路仍须在目标部署环境中完成端到端验收。
