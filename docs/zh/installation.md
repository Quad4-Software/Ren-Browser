# 安装

本页介绍预构建下载包、Docker 及从源码构建的方法。

## 预构建下载包（推荐）

从 [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases) 获取适合你系统的最新发布版本。

| 平台 | 文件 | 说明 |
|------|------|------|
| Linux x86_64 | `renbrowser-linux-amd64.AppImage` | `chmod +x` 后运行。同时包含普通二进制文件。 |
| Linux ARM64 | `renbrowser-linux-arm64.AppImage` | 步骤与 x86_64 相同。 |
| Windows | `renbrowser-windows-amd64.exe` | 直接运行，无需安装程序。 |
| macOS | `renbrowser-macos-universal.zip` | 解压后打开 `renbrowser.app`。 |
| 服务器（Linux x86_64） | `renbrowser-server-linux-amd64` | 用于 Docker 或自托管的无头二进制文件。 |
| 服务器（Linux ARM64） | `renbrowser-server-linux-arm64` | Raspberry Pi 3/4/5 及其他 64 位 ARM 开发板。 |
| 服务器（Linux ARMv6） | `renbrowser-server-linux-armv6` | Raspberry Pi Zero W 及其他 32 位 ARMv6 设备。 |
| 服务器（FreeBSD） | `renbrowser-server-freebsd-amd64`、`renbrowser-server-freebsd-arm64` | 在 FreeBSD 上无头运行。 |
| 服务器（OpenBSD / NetBSD） | `renbrowser-server-openbsd-amd64`、`renbrowser-server-netbsd-amd64` | 在 BSD 上无头运行。 |
| Android | `renbrowser.apk` | 当发布流水线包含时提供。 |

每个发布版本附带 `SHA256SUMS.txt`，可用于验证下载文件。参见[安全](security.md)。

### 验证下载（Linux 或 macOS）

```sh
sha256sum -c SHA256SUMS.txt
```

如果校验文件列出了多个资源，只检查你下载的文件即可。

### 系统要求

| 软件包 | 主机所需环境 |
|--------|-------------|
| **Linux AppImage** | 已捆绑 GTK 4、WebKitGTK 6 及其他库，无需单独安装 WebKit。某些发行版需要 FUSE 或 `APPIMAGE_EXTRACT_AND_RUN=1`。 |
| **Linux Flatpak** | Flatpak 加上 `org.gnome.Platform` 运行时（GTK 4 和 WebKitGTK 6）。 |
| **Linux 普通二进制文件** | 运行时需要 GTK 4 和 WebKitGTK 6.0（例如 Debian/Ubuntu 24.04+、Fedora 或 Arch）。 |
| **Windows `.exe`** | [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/)。通常在 Windows 10/11 上已预装。NSIS 安装程序可安装它；便携版 `.exe` 不会。 |
| **macOS `.app`** | 较新版本的 macOS，使用系统 WebKit（无需额外运行时）。 |
| **Android APK** | Android 5.0+（API 21+）。 |
| **服务器二进制文件 / Docker** | 无需桌面 GUI 栈。使用主机上的任意浏览器访问 UI。发布版服务器构建支持：Linux amd64/arm64/armv6、FreeBSD amd64/arm64、OpenBSD/NetBSD amd64。 |

## Docker 或 Podman（服务器模式）

官方镜像：`ghcr.io/quad4-software/renbrowser`

挂载你的 Reticulum 配置和配置文件数据，使容器能够加入网状网络。该镜像以非 root 用户运行，因此需要传入你的主机 UID/GID：

```sh
mkdir -p "$HOME/.reticulum-go" "$HOME/.renbrowser"
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

同样的参数适用于 `podman run`。在 Podman 上可以使用 `--userns=keep-id` 代替 `--user "$(id -u):$(id -g)"`。如果 SELinux 阻止绑定挂载，在卷参数中添加 `:Z`。

在同一机器上的任意浏览器中打开 `http://localhost:8080`。

从本仓库构建并运行：

```sh
task build:docker
task run:docker
```

服务器镜像**没有登录界面**。请仅在你信任的网络上暴露它。参见[服务器模式](server-mode.md)和[安全](security.md)。

## 从源码构建

适用于贡献者或没有预构建包的平台。

### 要求

- [Go](https://go.dev/) 1.26 或更新版本
- [Node.js](https://nodejs.org/) 22+ 和 [pnpm](https://pnpm.io/) 11+
- [Task](https://taskfile.dev/)（推荐）
- `~/.reticulum-go/` 下的 Reticulum 配置（或设置 `REN_BROWSER_CONFIG`）

### 基本构建

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task build
./bin/renbrowser
```

Go 模块会自动从 GitHub 拉取 Quad4 依赖项。

### 平台特定构建

```sh
task build:windows
task build:darwin
task build:android      # physical device (arm64)
task build:android:emu  # emulator (host ABI)
```

### 安装程序与软件包

```sh
task package                  # current OS
task package:linux:appimage   # Linux AppImage
task package:darwin:universal # macOS universal
task package:windows          # Windows NSIS installer
```

### Android SDK

Android 构建需要 [Android SDK](https://developer.android.com/studio)（API 34，NDK r26+）。设置 `ANDROID_HOME`，如果构建报告缺少工具，运行 `task android:install:deps`。

## 从源码构建服务器二进制文件

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

环境变量和部署说明参见[服务器模式](server-mode.md)。

## 安装后

1. 确认 Reticulum 配置已就位（[Reticulum 配置](reticulum-setup.md)）
2. 启动应用并打开 `about:`
3. 阅读[使用浏览器](using-the-browser.md)
