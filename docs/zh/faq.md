# 常见问题

常见问题的简短解答。

## Ren Browser 是什么？

一个用于浏览 Reticulum 网状网络上 NomadNet 页面的浏览器。它不是用于访问公共互联网的通用 Web 浏览器。

## 需要互联网连接吗？

你需要与其他节点的 Reticulum 连接。这可以完全离开公共互联网（无线电、局域网等）。

## 从哪里获取 reticulum-go？

在运行 Ren Browser 的机器上安装并配置 [reticulum-go](https://reticulum-go.quad4.io)。应用使用你的 Reticulum 配置（例如 `~/.reticulum-go/` 下的配置），但不会为你创建身份或接口。

## 什么是 NomadNet？

[Nomad Network](https://github.com/markqvist/NomadNet) 是一个基于 LXMF 和 Reticulum 的离网网状通信程序。可连接的节点可以托管页面和文件，通常使用 Micron 标记语言编写。Ren Browser 不嵌入 NomadNet 客户端。它查找公告为 NomadNet 的节点，并通过 Reticulum 打开其托管的页面。

## 什么是 Micron？

Nomad Network 节点上使用的一种带宽高效的标记语言。Ren Browser 在内容查看器中将其渲染为 HTML。

## 可以浏览普通的 HTTPS 网站吗？

不可以。Ren Browser 面向 Reticulum 网状网络内容，包括 NomadNet 节点上的页面，而不是任意的公共 Web URL。

## 桌面模式与服务器模式有什么区别？

桌面模式运行原生窗口并将数据存储在本地 SQLite 中。服务器模式通过 HTTP 提供 UI，供另一个浏览器使用。参见[服务器模式](server-mode.md)。

## 服务器模式在互联网上安全吗？

默认情况下不安全。没有登录机制。请使用 VPN、防火墙或带认证的反向代理。参见[安全](security.md)。

## 我的数据在哪里？

桌面模式下默认在 `~/.renbrowser/renbrowser.db`。参见[数据与配置文件](data-and-profiles.md)。

## 如何验证发布版本下载？

使用发布页面中的 `SHA256SUMS.txt`。参见[安全](security.md)。

## 如何安装扩展？

设置 → 扩展，或解压到 `~/.renbrowser/plugins/<id>/`。参见[扩展](extensions.md)。

## 如何输入节点地址？

粘贴 32 字符的十六进制哈希值或完整的 `hash:/page/file.mu` URL。参见[导航与 URL](navigation-and-urls.md)。

## 发现面板显示为空

在设置中检查 Reticulum 接口，并参阅 [Reticulum 配置](reticulum-setup.md)。

## 如何报告错误？

安全问题请按照[安全](security.md)通过 LXMF 提交。代码修复请参阅[贡献](contributing.md)。

## 项目使用什么许可证？

MIT。在地址栏中输入 `license:` 或阅读 [LICENSE](../../LICENSE)。

## 使用了什么技术栈？

Go、Wails v3、Svelte 5、SQLite 和 Quad4 Reticulum 库。参见[架构](architecture.md)。

## 可以在 Android 上运行吗？

可以，当你的发布版本发布了 APK 时。如有需要，使用 Android SDK 从源码构建。参见[安装](installation.md)。

## 如何更改键盘快捷键？

设置 → 快捷键绑定。参见[键盘快捷键](keyboard-shortcuts.md)。
