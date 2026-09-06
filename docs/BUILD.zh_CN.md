# NovaMaaS 构建、打包与镜像发布

本项目基于 QuantumNous/new-api `v1.0.0-rc.26`，完整来源记录见仓库根目录 `UPSTREAM.md`。实际工作流为 `.github/workflows/build.yml`。上游工作流保存在 `.github/upstream-workflows/` 供参考。

## 自动构建

| 触发方式 | 检查与打包 | 容器镜像 |
| --- | --- | --- |
| 向 `main` 新建或更新 PR | 相同检查；不打包 | 不构建、不发布 |
| PR 合并到 `main` | Go vet/build/test、前端类型检查和测试；Linux amd64/arm64 打包与启动验证 | 同时发布到 GHCR 和腾讯云 CCR，标签为 `sha-<合并提交 SHA>` |
| 关闭但未合并 PR | 跳过 | 不构建、不发布 |
| Actions 页面手动运行 | 只执行源码检查；不打包 | 不构建、不发布 |
| 直接推送 `main`、发布 Release 或推送 Git Tag | 不触发此工作流 | 不构建、不发布 |

构建使用上游 Dockerfile 中固定的 Bun `1.4.0` 和 Go `1.26.1`，执行 `bun install --frozen-lockfile`，保留 `go.mod`、`go.sum` 与 `web/bun.lock`。两种架构分别使用 GitHub 的原生 Linux runner。

前端编译进 Go 可执行文件。只有 PR 成功合并到 `main` 后，CI 才会运行容器，验证版本号、`/api/status` 及首页响应，然后从已验证的镜像提取二进制文件打包，最后发布同一镜像。安装包包含 `new-api`、许可证、署名、来源记录、版本及 `BUILD-INFO.txt`；镜像中的许可和来源文件位于 `/licenses`。

发布后，两个新的原生架构 runner 会按镜像 digest 分别从 GHCR 和腾讯云 CCR 拉取，核对源码提交、运行版本、API 状态和首页。GHCR 使用匿名拉取验证，CCR 使用仓库 Secret 登录后验证。

## 下载与运行安装包

进入 [GitHub Actions](https://github.com/yeruyi1024/novamaas-workspace/actions/workflows/build.yml)，打开 PR 合并触发的成功构建，在 Artifacts 中下载 `package-linux-amd64` 或 `package-linux-arm64`。构建产物保留 30 天。

将 Artifact ZIP 解压后，在包含 `.tar.gz` 和 `.sha256` 的目录执行：

```bash
sha256sum -c novamaas-*.tar.gz.sha256
tar -xzf <与你系统架构匹配的安装包.tar.gz>
cd <解压后的安装包目录>
./new-api --version
./new-api --port 3000
```

访问 `http://localhost:3000` 完成初始化。二进制直接运行时，数据库默认保存在当前工作目录，升级前应备份数据。

## 使用容器镜像

镜像地址为 `ghcr.io/yeruyi1024/novamaas-workspace`，腾讯云镜像副本地址为 `ccr.ccs.tencentyun.com/nova-proj/nova-maas`。PR 合并构建使用 `sha-<完整合并提交 SHA>`；工作流不写入 `latest` 或 `main`。

在仓库首页右侧点击 **Packages → novamaas-workspace**，或直接打开 [镜像列表](https://github.com/yeruyi1024/novamaas-workspace/pkgs/container/novamaas-workspace)。选择一个版本即可查看标签、digest、架构和拉取命令。CI 运行详情中 **Publish multi-arch image** 的 Summary 也会输出固定引用。

下面以已经发布的初始化构建为例；使用新构建时，将 `NOVAMAAS_IMAGE` 换成 Packages 中的对应标签：

```bash
NOVAMAAS_IMAGE=ghcr.io/yeruyi1024/novamaas-workspace:sha-65be9f0f7456bf86e61dbbb43835df40731a311b
docker buildx imagetools inspect "$NOVAMAAS_IMAGE"
docker pull "$NOVAMAAS_IMAGE"
docker run --rm "$NOVAMAAS_IMAGE" --version
docker run -d --name novamaas --restart unless-stopped \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v novamaas-data:/data \
  "$NOVAMAAS_IMAGE"
```

Tag 是可由维护者重新指向的名称。需要锁定确切镜像内容时，使用 `ghcr.io/yeruyi1024/novamaas-workspace@sha256:<digest>`；`docker buildx imagetools inspect` 会显示 digest。发布新版本时请使用新 Tag，不要移动已有发行标签。

CI 通过仓库自动提供的 `GITHUB_TOKEN` 和 `packages: write` 权限发布到 GHCR。GHCR 新建包可能默认是私有的；维护者需在包的 **Package settings → Change visibility** 中设为 **Public**，匿名用户才能拉取。公开源码仓库和容器包的可见性是分别管理的。

### 配置腾讯云 CCR 凭据

`ccr.ccs.tencentyun.com` 是腾讯云容器镜像服务个人版的默认域名。先在腾讯云容器镜像服务控制台初始化或重置个人版固定登录密码，并确认 `nova-proj` 命名空间及 `nova-maas` 仓库允许该账号推送。这里需要的是 Docker Registry 登录凭据，不是腾讯云 API 的 SecretId/SecretKey。

在 GitHub 仓库中打开 **Settings → Secrets and variables → Actions → Secrets → New repository secret**，添加：

- `TENCENT_CCR_USERNAME`：腾讯云账号 ID，即执行 `docker login ccr.ccs.tencentyun.com` 时使用的用户名。
- `TENCENT_CCR_PASSWORD`：腾讯云容器镜像服务个人版初始化或重置得到的固定登录密码。

不要把真实账号或密码写进工作流、提交记录或普通 GitHub Variables。工作流仅在 PR 已合并到 `main` 的 `closed` 事件中读取这两个 Secret，因此普通 PR 校验不会接触 CCR 凭据。

## 发布新版本

镜像版本以合并提交的 `sha-<完整提交 SHA>` 为准。直接推送 `main`、发布 GitHub Release、推送 Git Tag 或手动执行工作流都不会构建或发布容器镜像；需要发布代码变更时，应通过 PR 合并到 `main`。CI 不会替你创建 Release、改写 Release 正文或自动合入上游新版本。

## 本地从源码构建

```bash
git clone https://github.com/yeruyi1024/novamaas-workspace.git
cd novamaas-workspace
docker build -t novamaas:local .
docker run --rm novamaas:local --version
```

需要直接构建二进制时，安装上述 Bun/Go 版本并执行：

```bash
make build-web
GOWORK=off make test
GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc go build \
  -ldflags "-s -w -X github.com/QuantumNous/new-api/common.Version=$(cat VERSION)" \
  -o new-api .
```
