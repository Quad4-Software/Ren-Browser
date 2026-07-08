# 安全

Ren Browser 适用于你信任的系统和网络。本页总结了安全使用方式、下载验证以及如何报告问题。

## 信任边界

| 界面 | 风险 |
|------|------|
| 桌面应用 | 带有 Go 绑定的本地 WebView。页面内容中无 Node.js。 |
| 服务器模式 | 开放的 HTTP 端口，无内置登录。 |
| 插件 | 以声明的权限运行。WASM 插件受能力沙盒化限制。 |
| 网状网络内容 | 与任何网络内容一样不可信。除非插件请求 `render.unsanitized`，否则 Micron HTML 会经过净化处理。 |

## 服务器模式

`renbrowser-server` **没有认证机制**。如果你将其暴露到互联网，你将面临以下风险：

- 自动化扫描
- 滥用你的 Reticulum 接口
- 单进程应用过载

如果必须暴露：

1. 在反向代理后面加上访问控制
2. 使用带有有效证书的 HTTPS
3. 通过防火墙或 VPN 规则限制端口
4. 保持应用更新

代理参数请参阅[服务器模式](server-mode.md)。

## 桌面插件

只安装来自你信任的人或项目的扩展。启用前请阅读**设置 → 扩展**中的权限列表。

### 安装时检查

安装扩展时，RenBrowser 会显示：

- 每个权限的开关（禁用的权限在运行时不会被授予）
- 从清单和包文件中扫描到的网络端点
- 签名徽章（未签名、已签名、受信任发布者、已篡改）
- 启发式安全评估（未签名 + 网络、危险权限、JS 中的可疑模式）

无效签名会阻止安装。接受风险后仍然可以安装未签名扩展。

### 运行时保护

- 已授予的权限对 JS 的 `PluginFetch` 和 WASM 的 `http_fetch` 强制执行
- 未授予 `network.fetch` 时，WASM 网络导出被阻止
- 每次调用对插件 HTTP 请求和 WASM 工作的限制降低了因扩展行为不当导致冻结的风险
- 安装后对扩展文件完整性进行哈希；外部篡改会禁用扩展，直到你重新启用它
- 通过数据库支持的摘要检测应用外对用户受信任发布者列表的修改

当开发工具打开时，插件 HTTP 流量会出现在**开发工具 → 网络**中。

签名、语言和清单字段请参阅[扩展](extensions.md)。

## 验证下载

官方构建来自 [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases) 和 GitHub Actions CI。

每个发布版本应包含 `SHA256SUMS.txt`。检查你的文件：

```sh
sha256sum -c SHA256SUMS.txt
```

对于 Docker，在信任某个构建后，优先通过摘要（`@sha256:...`）固定版本。GHCR 上的镜像包含来自 Docker Buildx 的构建来源和 SBOM。

如果二进制文件与已发布的校验和不匹配，将其视为不可信。

## WASM 的子资源完整性

Micron 解析器 WebAssembly 及其 `wasm_exec.js` 伴随文件在执行前通过 SHA-384 SRI 检查。哈希值不匹配会阻止代码运行并显示错误。

## 静态数据

- 应用状态：`~/.renbrowser/` 下的 SQLite
- Reticulum 密钥：你的 Reticulum 配置目录
- 服务器公共模式：部分数据仅存储在每个客户端的浏览器 `localStorage` 中

如果机器是共享的或便携的，请在操作系统级别加密磁盘。

## 报告漏洞

对于未修复的安全漏洞，**不要**开公开的 GitHub Issue。

**首选联系方式：**

1. LXMF：`f489752fbef161c64d65e385a4e9fc74`

请包含版本、平台、复现步骤和影响描述。

法律和许可问题请发送至 [LEGAL.md](../../LEGAL.md)（`legal@quad4.io`），而非安全渠道。

## CI 与供应链（概述）

GitHub Actions 会定期运行测试、gosec、Trivy 扫描和 CodeQL。工作流中的第三方 Actions 固定到提交 SHA。

## 下一步

- [扩展](extensions.md)权限列表
- [服务器模式](server-mode.md)部署
- 仓库根目录 [SECURITY.md](../../SECURITY.md) 中的规范策略
