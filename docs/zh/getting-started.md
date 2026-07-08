# 快速入门

Ren Browser 让你可以通过 Reticulum 打开 NomadNet 页面。可以把它看作一个专为网状网络而非公共互联网构建的小型浏览器。

## Ren Browser 的功能

- 从网状网络目标地址打开 `.mu` 页面及其他 NomadNet 内容
- 显示可通过你的 Reticulum 接口访问的节点
- 将标签页、历史记录、书签和设置保存在本机（使用公共服务器模式时保存在浏览器中）
- 使用 WASM 解析器和内置主题渲染 Micron 标记
- 支持添加面板、URL 方案和命令的扩展

## 前置要求

在 Ren Browser 能够加载远程页面之前，你需要在同一台机器上（或挂载到服务器模式的 Docker 容器中）有一个可用的 **Reticulum** 配置。

Ren Browser 默认从 `~/.reticulum-go/` 读取 Reticulum 配置。你可以通过 `--config` 参数或 `REN_BROWSER_CONFIG` 环境变量指向另一个文件。

在网状网络上浏览 NomadNet 页面**不需要**传统的互联网连接。但你至少需要一个能够访问其他 Reticulum 节点的接口。

## 桌面模式与服务器模式

| 模式 | 适用场景 |
|------|----------|
| **桌面**（默认） | 在 Linux、Windows 或 macOS 上日常使用。原生窗口，本地 SQLite 数据库。 |
| **服务器** | 家庭实验室、Docker 或已运行 Reticulum 的机器。通过 `http://host:8080` 在另一个浏览器中打开 Ren Browser。 |
| **Android** | 发布版本包含 APK 时的移动端构建。在触控布局中提供相同的核心浏览功能。 |

## 首次启动检查清单

1. 安装 Reticulum 并在 `~/.reticulum-go/` 下创建或复制配置文件
2. 从[发布页面](https://github.com/Quad4-Software/Ren-Browser/releases)安装 Ren Browser 或从源码构建
3. 启动 Ren Browser 并等待 Reticulum 通过你的接口连接
4. 打开**发现**面板，或在地址栏中输入 32 字符的目标哈希值
5. 访问 `about:` 以确认版本、配置路径和数据目录

## 内置页面（无需网状网络）

即使离线于网状网络，以下页面也可正常使用：

| 地址 | 用途 |
|------|------|
| `about:` | 应用版本、构建信息、路径 |
| `license:` | MIT 许可证文本 |
| `editor:` | 内置 Micron 编辑器 |

在地址栏中输入 `settings` 或按设置快捷键可打开偏好设置。

## 下一步

- 如果尚未安装，请参阅[安装](installation.md)
- 如果页面加载失败或发现面板为空，请参阅 [Reticulum 配置](reticulum-setup.md)
- 日常浏览请参阅[使用浏览器](using-the-browser.md)
