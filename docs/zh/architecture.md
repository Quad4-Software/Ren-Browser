# 架构

Ren Browser 结构的高层次概览。

## 概览

```
┌─────────────────────────────────────────────────────────┐
│  Svelte 5 frontend (Wails webview or browser in server) │
│  Tabs, chrome, Micron viewer, panels, settings          │
└───────────────────────┬─────────────────────────────────┘
                        │ Wails bindings / HTTP API
┌───────────────────────▼─────────────────────────────────┐
│  internal/app : BrowserService                          │
│  Navigation, tabs, history, settings, plugin bridge     │
└───────┬─────────────┬─────────────┬─────────────────────┘
        │             │             │
        ▼             ▼             ▼
   internal/rns   internal/store  internal/plugins
   Reticulum      SQLite          Extensions
        │
        ▼
   internal/nomadnet : LXMF page fetch for announced NomadNet nodes
        │
        ▼
   internal/micron : Markup parse and HTML render
```

## 入口点

| 文件 | 构建标签 | 职责 |
|------|----------|------|
| `main_desktop.go` | `!server && !android` | Wails 窗口，嵌入 `frontend/dist` |
| `main_server.go` | `server` | HTTP 服务器，相同的嵌入资源 |
| Android 主入口 | `android` | 移动端 Shell |

`internal/bootstrap` 将配置、存储、插件和 Wails 应用连接在一起。

## 前端

- **框架：** Svelte 5 与 Vite
- **绑定：** 生成于 `frontend/bindings/renbrowser/`
- **主 UI：** `frontend/src/App.svelte` 负责编排界面外框和面板
- **组件：** `frontend/src/lib/components/`（标签栏、发现、设置等）
- **浏览器逻辑：** `frontend/src/lib/browser/`（URL、快捷键绑定、错误）

Micron 渲染可使用由 `MicronWasmManager` 管理的 WASM 解析器，并进行 SRI 验证。

## 后端服务

### BrowserService（`internal/app`）

UI 的核心 API：

- 导航 URL 并管理标签页状态
- 暴露发现、历史记录、收藏夹
- 加载和保存偏好设置
- 连接到插件主机

### Reticulum 栈（`internal/rns`）

封装 `quad4/reticulum-go`：

- 启动和停止传输
- 报告接口统计数据
- 从设置热重载配置

### 页面获取（`internal/nomadnet`）

通过 LXMF 和 Reticulum 获取远程 `.mu` 及相关内容。发现将节点标记为 NomadNet，当其公告匹配时。Ren Browser 不使用 NomadNet 客户端库。

### 存储（`internal/store` + `internal/db`）

从旧版 `state.json` 迁移的 SQLite 持久化。

### 插件（`internal/plugins`）

- 清单验证
- 权限执行
- 内置方案（`about:`、`license:`、`editor:`）
- JS 和 WASM 插件运行时

## 内容与渲染

| 包 | 职责 |
|----|------|
| `internal/content` | 静态页面（关于、许可证） |
| `internal/micron` | Micron 转 HTML |
| `internal/micronwasm` | WASM 解析器集成 |
| `internal/cache` | 页面缓存助手 |

## 服务器中间件

`internal/servermw` 在服务器模式下处理基础路径头和代理感知的 URL 构建。

## 配置

`internal/config` 将参数、`.env` 和 `REN_BROWSER_*` 变量解析到 `bootstrap` 使用的 `Runtime` 结构体中。

## 品牌与路径

`internal/brand`（从 `build/brand.yml` 生成）定义稳定名称：

- 数据目录 `.renbrowser`
- 数据库文件 `renbrowser.db`
- 显示名称和版本标签

## 构建与打包

- `Taskfile.yml`：开发者命令
- `build/`：各操作系统的打包（Linux AppImage、Windows NSIS、macOS、Android、Docker）
- `build/config.yml`：Wails 项目配置

## CI

GitHub Actions 运行 Go 测试、前端检查、安全扫描、桌面和服务器冒烟构建以及发布产物。参见 `.github/workflows/`。

## 扩展点

1. **插件**：清单驱动的 UI 和方案
2. **主题**：JSON 标记文件
3. **社区接口**：设置中的 Reticulum 配置片段

## 下一步

- 本地构建请参阅[开发](development.md)
- 插件 API 界面请参阅[扩展](extensions.md)
- 仓库 `README.md` 布局表中的源码树
