# 服务器模式

服务器模式将 Ren Browser 作为 Web 应用运行，不带桌面 Shell。你可以通过 HTTP URL 在另一个浏览器中访问它。

## 适用场景

- 已运行 Reticulum 的家庭实验室或 VPS
- Docker 部署
- 不想安装桌面应用的共享机器
- 从局域网上的平板电脑或手机访问

## 快速开始

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

在 Firefox、Chromium 或 Safari 中打开 `http://localhost:8080`（或你的主机 IP）。

## Docker

已发布的镜像：

```
ghcr.io/quad4-software/renbrowser:latest
```

运行示例：

```sh
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

不要以只读方式挂载 Reticulum 目录；网状网络需要在配置文件旁边更新存储。挂载详情和 Podman 说明参见 [Reticulum 配置](reticulum-setup.md#server-and-docker)。

本地构建：

```sh
task build:docker
task run:docker
```

## 命令行参数

| 参数 | 用途 |
|------|------|
| `--host` | 绑定地址（服务器构建中默认为 `0.0.0.0`） |
| `--port` | HTTP 端口（默认 `8080`） |
| `--config` | Reticulum 配置路径 |
| `--trust-proxy` | 信任来自反向代理的 `X-Forwarded-*` 头 |
| `--base-path` | 在子路径下提供服务时的 URL 前缀 |
| `--public-mode` | 将收藏夹、历史记录和标签页存储在浏览器的 `localStorage` 中，而不是服务器 SQLite 中 |
| `--profile` | 命名的配置文件数据库 |
| `--import-profile` / `--export-profile` | 启动时的配置文件 JSON |

## 环境变量

服务器从工作目录中的 `.env` 文件读取配置。环境中已设置的变量不会被覆盖。

| 变量 | 用途 |
|------|------|
| `WAILS_SERVER_HOST` / `REN_BROWSER_HOST` | 绑定地址 |
| `WAILS_SERVER_PORT` / `REN_BROWSER_PORT` | 端口 |
| `REN_BROWSER_CONFIG` / `RETICULUM_CONFIG` | Reticulum 配置 |
| `REN_BROWSER_TRUST_PROXY` | `true` / `1` / `yes` 启用信任代理 |
| `REN_BROWSER_BASE_PATH` | 子路径前缀 |
| `REN_BROWSER_PUBLIC_MODE` | 公共模式开关 |
| `REN_BROWSER_PROFILE` | 配置文件名称 |
| `REN_BROWSER_IMPORT_PROFILE` | 启动时的导入路径 |
| `REN_BROWSER_EXPORT_PROFILE` | 启动时的导出路径 |

## 公共模式

不使用 `--public-mode` 时，服务器将标签页、历史记录和收藏夹保存在服务器磁盘上的 SQLite 数据库中。共享该实例的每个客户端都看到相同的数据。

使用 `--public-mode` 时，这些项目存储在每个浏览器的 `localStorage` 中。当多人使用同一台服务器且不应共享同一配置文件时，请使用此模式。

## 反向代理

典型的 nginx 或 Caddy 设置：

1. 在代理处终止 TLS
2. 代理到 `127.0.0.1:8080`
3. 传递 `X-Forwarded-Proto` 和 `X-Forwarded-Host`
4. 使用 `--trust-proxy` 启动 Ren Browser
5. 如果应用不在域名根目录下，设置 `--base-path`

启用信任代理时，会识别 `X-RenBrowser-Base-Path` 头。

## 无内置认证

任何能够访问 HTTP 端口的人都可以使用浏览器并触发 Reticulum 流量。不要在没有以下措施的情况下将 8080 端口暴露到公共互联网：

- 防火墙规则
- VPN
- 带认证的反向代理
- 或以上全部

发布服务器前请阅读[安全](security.md)。

## 资源覆盖（高级）

开发时可以从磁盘或 zip 文件而非嵌入资源提供前端文件：

- `--assets-dir path`
- `--assets-zip path`

环境变量：`REN_BROWSER_ASSETS_DIR`、`REN_BROWSER_ASSETS_ZIP`。

## 下一步

- SQLite 与公共模式请参阅[数据与配置文件](data-and-profiles.md)
- 部署加固请参阅[安全](security.md)
- 发布二进制文件请参阅[安装](installation.md)
