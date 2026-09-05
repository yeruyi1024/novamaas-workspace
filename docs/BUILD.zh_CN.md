# NovaMaaS 构建、打包与镜像发布

本项目基于 QuantumNous/new-api `v1.0.0-rc.26`，完整来源记录见仓库根目录 `UPSTREAM.md`。实际工作流为 `.github/workflows/build.yml`。上游工作流保存在 `.github/upstream-workflows/` 供参考。

## 自动构建

| 触发方式 | 检查与打包 | GHCR 镜像 | GitHub Release |
| --- | --- | --- | --- |
| 推送 `main` | Go vet/build/test、前端类型检查和测试；Linux amd64/arm64 包 | `main`、`latest`、`sha-<完整提交 SHA>` | 不创建 |
| 向 `main` 提交 PR | 相同检查、双架构构建与启动验证 | 不发布 | 不创建 |
| 推送 `v*-novamaas.*` 标签 | 相同检查和打包，标签须与 `VERSION` 一致 | 对应标签、`sha-<完整提交 SHA>` | 创建预发布并上传安装包、校验文件 |
| Actions 页面手动运行 | 构建所选分支/标签 | 仅 `main` 或符合规则的发行标签可发布 | 仅符合规则的发行标签 |

构建使用上游 Dockerfile 中固定的 Bun `1.4.0` 和 Go `1.26.1`，执行 `bun install --frozen-lockfile`，保留 `go.mod`、`go.sum` 与 `web/bun.lock`。两种架构分别使用 GitHub 的原生 Linux runner。

前端编译进 Go 可执行文件。CI 会运行容器，验证版本号、`/api/status` 及首页响应，然后从已验证的镜像提取二进制文件打包，最后发布同一镜像。安装包包含 `new-api`、许可证、署名、来源记录、版本及 `BUILD-INFO.txt`；镜像中的许可和来源文件位于 `/licenses`。

## 下载与运行安装包

进入 [GitHub Actions](https://github.com/yeruyi1024/novamaas-workspace/actions/workflows/build.yml)，打开成功的构建，在 Artifacts 中下载 `package-linux-amd64` 或 `package-linux-arm64`。构建产物保留 30 天。发行标签对应的包同时保存在 [Releases](https://github.com/yeruyi1024/novamaas-workspace/releases)。

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

镜像地址：`ghcr.io/yeruyi1024/novamaas-workspace`。`latest` 跟随通过构建的 `main`；固定版本部署可使用具体发行标签或镜像 digest。

```bash
docker pull ghcr.io/yeruyi1024/novamaas-workspace:latest
docker run -d --name novamaas --restart unless-stopped \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v novamaas-data:/data \
  ghcr.io/yeruyi1024/novamaas-workspace:latest
```

CI 通过仓库自动提供的 `GITHUB_TOKEN` 和 `packages: write` 权限发布到 GHCR，无需配置 Docker Hub 账号或额外 Secret。GHCR 新建包可能默认是私有的；维护者需在包的 **Package settings → Change visibility** 中设为 **Public**，匿名用户才能拉取。公开源码仓库和容器包的可见性是分别管理的。

## 发布新版本

版本号格式示例为 `v1.0.0-rc.26-novamaas.1`。修改根目录 `VERSION`，提交并推送到 `main`，确认检查通过后，再创建同名标签：

```bash
git tag -a v1.0.0-rc.26-novamaas.1 -m "NovaMaaS initial localized fork"
git push origin v1.0.0-rc.26-novamaas.1
```

后续发行递增 `novamaas` 后缀；示例标签已经存在时不要重复创建。本分支基于上游候选版本，因此此工作流将发行版标记为 prerelease。CI 不自动合入上游新版本。

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
