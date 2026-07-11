# Reticulum 配置

Ren Browser 通过 `quad4/reticulum-go` 栈使用 [Reticulum](https://reticulum.network/)。本页说明应用的期望配置以及如何修复常见的网状网络问题。

## 默认配置位置

| 项目 | 默认路径 |
|------|----------|
| Reticulum 配置目录 | `~/.reticulum-go/` |
| 覆盖参数 | `--config /path/to/config` |
| 覆盖环境变量 | `REN_BROWSER_CONFIG` 或 `RETICULUM_CONFIG` |

目录内的具体文件取决于你的 Reticulum 或 reticulum-go 配置。Ren Browser 在启动时启动该栈，并通过**设置**重新加载接口变更。

## 启动时的流程

1. Ren Browser 加载你的 Reticulum 配置
2. 接口上线（UDP、TCP、RNode 及你配置的其他接口）
3. NomadNet 节点的公告出现在**发现**面板中
4. 页面请求通过 LXMF 和 Reticulum 发送到托管 Micron 页面的节点

如果启动失败，请检查终端日志（桌面模式）或容器日志（服务器模式）。应用会继续运行，你仍然可以打开 `about:` 和**设置**。

## 设置中的接口

打开**设置**并找到 Reticulum 部分。你可以：

- 查看哪些接口处于活动状态
- 查看发送和接收统计数据
- 编辑配置并应用热重载，无需重启整个应用

当你添加新接口或更改密钥并希望浏览器快速应用变更时，请使用此功能。

## 加入网状网络

你至少需要一条通往其他 Reticulum 节点的路径。常见选项：

- 与其他 Reticulum 对等节点在同一局域网上的**本地 UDP 或 TCP**
- **RNode** 或类似的无线电硬件
- 指向已知对等节点或中心节点的**接口定义**

Reticulum 不在本手册的讨论范围内。请阅读 [Reticulum 手册](https://reticulum.network/manual/)了解接口语法和身份管理。

## NomadNet 目标地址

NomadNet 页面存储在 Reticulum 目标地址上。在地址栏中你可以使用：

- 完整路径，例如 `abcdef0123456789abcdef0123456789:/page/index.mu`
- 单独的 32 字符十六进制哈希值（Ren Browser 会自动附加 `:/page/index.mu`）

页面使用 Micron 标记格式。Ren Browser 通过内置的 Micron 渲染管道渲染它们。

## 发现面板为空时

按以下步骤排查：

1. 确认 Reticulum 在 Ren Browser 内部正在运行（设置中显示接口）
2. 检查你的接口配置是否与网状网络上的对等节点一致
3. 连接后等待一段时间，公告不是即时出现的
4. 确认你与预期看到的节点在同一逻辑网络上

## 页面超时或失败时

1. 确认目标哈希值正确
2. 检查你是否有通往该目标的路由（不仅仅是发现可见性）
3. 从发现面板尝试另一个已知可用的节点
4. 查看开发工具或日志中的 LXMF 或传输错误

## 服务器与 Docker

运行 `renbrowser` Docker 镜像时，挂载主机 Reticulum 目录并以主机用户身份运行，使非 root 容器能够读取密钥并写入网状网络存储：

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

不要以只读方式挂载配置；Reticulum 需要在配置文件旁边更新存储。

## 下一步

- 浏览已公告节点请参阅[发现](discovery.md)
- 地址栏格式请参阅[导航与 URL](navigation-and-urls.md)
- 错误信息请参阅[故障排查](troubleshooting.md)
