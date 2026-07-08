# 扩展

Ren Browser 支持添加 URL 方案、侧边栏面板、命令、主题、设置页面和开发工具选项卡的插件。

## 安装扩展

### 通过设置安装

1. 打开**设置 → 扩展**
2. 选择**安装扩展**，然后选择 **.zip** 文件、**文件夹**或**捆绑的 .wasm 模块**
3. 查看安装预览：
   - 请求的权限（安装前可以禁用单个权限）
   - 扩展可能联系的外部 URL（从清单和包文件中扫描）
   - 发布者签名状态（未签名、已签名、受信任发布者签名或无效）
   - 安全评估警告
   - 捆绑的 UI 语言（当扩展附带 `locales/*.json` 时）
4. 确认并启用扩展

请求 `network.fetch` 的扩展会显示一个列出检测到的端点的确认对话框。即使你在安装时禁用了 `network.fetch`，端点仍然可见，以便你了解授予该权限时包会联系哪些地址。

### 手动安装

将插件解压到：

```
~/.renbrowser/plugins/<id>/
```

该文件夹必须包含 `renbrowser.plugin.json`。清单中的 `id` 应与文件夹名称匹配。

## 示例扩展

仓库中包含 `extensions/hello-extension/`：

- 注册 `hello:` URL 方案
- 添加 **Hello** 侧边栏面板
- 定义带有 `mod+shift+h` 快捷键的 **Say hello** 命令

编写自己的插件时可以此为模板。

`extensions/micron-translator/` 使用 Google Translate（公共端点）或 LibreTranslate 实例（在侧边栏面板中设置 URL 和可选 API 密钥）翻译 Micron（`.mu`）页面。命令：**Translate Micron page**（`mod+shift+t`）和 **Restore original**（`mod+shift+r`）。

## 清单文件

文件名：`renbrowser.plugin.json`

必填字段：

| 字段 | 用途 |
|------|------|
| `manifestVersion` | 当前为 `1` |
| `id` | 唯一标识符（`a-z`、`A-Z`、`0-9`、`.`、`-`，3 到 128 个字符） |
| `name` | 显示名称 |
| `version` | 语义化版本字符串 |
| `main` | 前端入口脚本（仅有后端时可选） |
| `permissions` | 能力列表（见下文） |

可选字段包括 `description`、`author`、`license`、`engines`、`backend`、`network` 和 `contributes`。

### 引擎约束

```json
"engines": { "renbrowser": ">=0.1.0" }
```

如果你的应用版本过旧，主机会拒绝加载该插件。

### 网络端点

使用 `network.fetch` 的扩展应声明联系的主机或 URL：

```json
"network": {
  "endpoints": [
    "https://api.example.com/",
    "User-configured service URL"
  ]
}
```

安装时 RenBrowser 还会扫描 `.js`、`.go`、`.wasm` 及其他包文件中的 `http`/`https` URL，并将找到的内容与清单条目一并列出。

### 贡献项

| 类型 | 用途 |
|------|------|
| `urlSchemes` | 处理自定义方案 |
| `panels` | 侧边栏或其他面板槽 |
| `commands` | 命令面板条目和快捷键绑定 |
| `themes` | 额外的主题 JSON 文件 |
| `settings` | 设置子页面 |
| `devtools` | 开发工具选项卡 |
| `renderers` | 针对 MIME 类型或扩展名的自定义渲染器 |

## 权限

插件必须声明所需的权限。已知权限：

| 权限 | 允许内容 |
|------|----------|
| `storage.plugin` | 插件的私有键值存储 |
| `navigation.read` | 读取当前 URL 和标签页信息 |
| `navigation.write` | 触发导航 |
| `network.fetch` | 通过允许的网络 API 进行请求 |
| `events.emit` | 发出主机事件 |
| `events.subscribe` | 监听主机事件 |
| `devtools.network` | 开发工具中的额外网络详情 |
| `render.unsanitized` | 跳过部分 HTML 净化（危险） |

主机在运行时强制执行权限。安装时禁用的权限按扩展存储，不会授予 JS 的 `ctx.network.fetch` 或 WASM 的 `http_fetch`。

## 发布者签名

扩展可以在 `renbrowser.plugin.rsg` 中附带 Ed25519 签名（与 Reticulum `rnid` 工具兼容）。签名无效的已签名包无法安装。

安装预览和扩展列表显示徽章：

| 徽章 | 含义 |
|------|------|
| 未签名 | 没有签名文件 |
| 已签名 | 来自 Reticulum 身份的有效签名 |
| 受信任 | 由受信任列表中的发布者签名 |
| 已篡改 | 扩展文件在 RenBrowser 外被修改（扩展被禁用，直到你重新启用它） |

安装时，你可以选择**信任此发布者身份**，将有效签名者添加到你的用户受信任列表（`~/.renbrowser/trusted_publishers.json`）。RenBrowser 还内置了一个小型受信任列表。用户列表受存储在配置文件数据库中的摘要保护；在不更新数据库的情况下进行外部编辑会被检测到。

使用 `build/scripts/sign-extension.sh` 对目录或 zip 进行签名（需要 Python `rnid`）。

## 插件 UI 翻译

扩展可以在 `locales/<code>.json` 下捆绑自己的 UI 字符串（例如 `locales/en.json`）。清单中的面板标题和命令可以使用 `%key.path%` 占位符；主机从 `/_plugins/<id>/locales/<code>.json` 加载目录。

安装预览在存在时会列出捆绑的语言代码。

## 前端入口脚本

典型的 `main.js` 导出：

- `activate(ctx)`：订阅事件、注册 UI
- `deactivate()`：清理
- `mount(el)`：渲染侧边栏面板 HTML
- `handleScheme(url)`：用于 URL 方案处理器

拥有 `network.fetch` 权限的插件可以调用 `ctx.network.fetch()` 向公共 `http`/`https` URL 发起 HTTP GET/POST 请求（前提是安装时授予了该权限）。开始网络相关工作前，请检查 `ctx.capabilities.networkFetch`。

拥有 `backend` WASM 模块的插件可以调用 `ctx.wasm.call(export, input)` 运行导出函数，例如 `translate_micron`。使用 `ctx.content.getActivePage()`、`ctx.content.renderRaw(path, raw)` 和 `ctx.content.updateActivePage()` 在转换 Micron 源码后重新渲染当前标签页。

使用 `ctx.i18n.t("key")` 获取扩展语言文件中的字符串。

## 捆绑的 WASM 模块

可分发的扩展可以作为单个 `.wasm` 文件发布。该模块包含自定义段：

- `renbrowser.plugin` — 清单 JSON（`renbrowser.plugin.json`）
- `renbrowser.files` — 相对路径到 UTF-8 文件内容的映射（例如 `main.js`、`locales/en.json`）
- `renbrowser.signature` — 可选的 RSG 签名字节

通过**设置 → 扩展 → 安装扩展 → 选择 .wasm 模块**安装。主机将元数据解包到插件目录，并将 WASM 二进制文件保留为清单的 `backend`。

`extensions/micron-translator/` 附带 `translator.wasm`（TinyGo）。使用 `extensions/micron-translator/build-wasm.sh` 重新构建，或在构建后使用 `go run ./extensions/micron-translator/bundle` 打包。

## WASM 后端

插件可以将 `backend` 设置为 WASM 模块路径以处理较重的逻辑。WASM 插件在具有明确授权的受限运行时中运行。

当安装时授予了 `network.fetch` 权限，主机提供带有 `http_fetch` 的 `renhost` 模块。导出函数（如 `translate_micron(in_ptr, in_len) -> out_len`）在线性内存中读取 JSON 输入并写入 JSON 输出。

安全措施包括每次调用的网络请求限制、WASM 调用超时和输入大小上限。未授予 `network.fetch` 时，网络密集型导出会被完全阻止。

## 开发工具

当**开发工具 → 网络**打开时，扩展发起的出站 HTTP 请求（JS 的 `PluginFetch` 和 WASM 的 `http_fetch`）会以**扩展请求**为来源，显示状态码和持续时间。

## 完整性与篡改检测

安装后，RenBrowser 会存储每个扩展文件内容的加密哈希值（不包括签名文件）。如果磁盘上的文件在应用外被修改，该扩展会被禁用并标记为**已篡改**。重新启用时接受当前文件并刷新存储的哈希值。

## 安全说明

- 只安装来自你信任的来源的插件
- 确认安装前请阅读权限和检测到的网络端点
- 优先使用来自你认识的发布者的已签名扩展
- 将插件视为可访问你配置文件数据的本地程序

## 下一步

- 源码参考：仓库中的 `internal/plugins/manifest.go`
- 插件威胁模型和签名请参阅[安全](security.md)
- 插件主机的开发请参阅[开发](development.md)
