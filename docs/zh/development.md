# 开发

如何在本地开发 Ren Browser。

## 要求

- Go 1.26+
- Node.js 22+ 和 pnpm 11+
- Task（推荐）
- 用于实时网状网络测试的 Reticulum 配置（可选）

## 克隆并运行

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task dev
```

`task dev` 以开发模式运行 Wails，默认在端口 9245 上启用 Vite 热重载。

## 常用任务

| 任务 | 用途 |
|------|------|
| `task dev` | 桌面开发模式 |
| `task build` | 为当前操作系统构建生产版本 |
| `task run` | 运行已构建的二进制文件 |
| `task check` | 完整质量检查（Go + 前端） |
| `task test` | 所有单元测试 |
| `task test:interop` | 实时 Reticulum 测试（需要网络） |
| `task build:server` | 无头服务器二进制文件 |
| `task run:server` | 本地运行服务器 |
| `task package` | 平台安装程序或包 |

## 质量检查

发送变更前：

```sh
task check
```

`check` 运行：

- 对 Go 源码执行 `gofmt`
- `go test ./...`
- gosec 安全扫描
- 品牌一致性检查
- 前端类型检查、lint、格式检查、knip、审计和 vitest

可选的更严格 Go 测试：

```sh
task test:go:race
task test:go:hard
task fuzz:go
```

## 仅前端

```sh
cd frontend
pnpm install
pnpm check
pnpm test
```

`frontend/bindings/` 下的绑定由 Wails 从 Go 服务生成。

## Go 代码布局

| 路径 | 职责 |
|------|------|
| `main_desktop.go` | 桌面入口（Wails 窗口） |
| `main_server.go` | 服务器入口（仅 HTTP） |
| `internal/app/` | 暴露给 UI 的浏览器服务 |
| `internal/rns/` | Reticulum 栈封装 |
| `internal/nomadnet/` | 通过 LXMF 获取页面。NomadNet 节点从公告中检测，不使用 NomadNet 库 |
| `internal/micron/` | Micron 渲染 |
| `internal/store/` | SQLite 持久化 API |
| `internal/plugins/` | 扩展主机 |
| `frontend/` | Svelte 5 UI |

更完整的结构图请参阅[架构](architecture.md)。

## 插件开发

```sh
task test:plugins
```

将你的开发插件安装到 `~/.renbrowser/plugins/` 并重新加载应用。

## 互操作测试

```sh
task test:interop
```

需要实时 Reticulum 网络和 `interop` 构建标签。仅做 UI 工作时可跳过。

## 资源探测

调试使用哪个资源加载器（嵌入式还是磁盘）时，设置 `REN_BROWSER_ASSET_PROBE=1`。

## SPDX 头

新的 Go 文件应包含：

```go
// SPDX-License-Identifier: MIT
```

与相邻文件的现有风格保持一致。

## 下一步

- [架构](architecture.md)
- [贡献](contributing.md)
- [扩展](extensions.md)
