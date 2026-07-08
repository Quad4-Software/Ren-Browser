# 贡献

Ren Browser 接受通过 Reticulum 发送的补丁。根据项目政策，GitHub Pull Request 也可能受欢迎。本页遵循 [CONTRIBUTING.md](../../CONTRIBUTING.md)。

## 补丁工作流程

1. 克隆或 Fork 仓库
2. 创建分支并进行专注的修改
3. 修改了 Go 或前端代码时运行 `task check`
4. 提交清晰的提交信息
5. 使用 `git format-patch` 导出补丁
6. 通过 LXMF 发送 `.patch` 文件

## 通过 LXMF 发送补丁

目标地址：

```
f489752fbef161c64d65e385a4e9fc74
```

使用 Sideband、Meshchat、MeshChatX 或任何支持文件附件的 LXMF 客户端附上补丁。在消息正文中附上简短说明。

网状网络上的处理需要一定时间，请耐心等待。

## 导出命令

```sh
# Most recent commit
git format-patch -1

# Last N commits
git format-patch -N

# All commits since main
git format-patch main..HEAD
```

每个提交变成一个 `.patch` 文件。

## 补丁规范

- 尽量每个补丁系列包含一个逻辑变更
- 发送前进行测试
- 符合现有代码风格
- 新的 Go 文件保留 `// SPDX-License-Identifier: MIT`
- 在消息正文中披露 AI 使用情况（见下文）

## 许可

提交补丁即表示你同意其在 [MIT 许可证](../../LICENSE)下授权。你确认你有权提交该工作。

## 生成式 AI 政策

在以下条件下，你可以使用 AI 工具：

- 你的配置为模型提供了足够的上下文
- 你的服务提供商不会用你粘贴的代码进行训练

请阅读 [Reticulum Zen](https://reticulum.network/manual/zen.html) 和 [Reticulum License](https://reticulum.network/manual/license.html)。

在补丁消息中**披露**你使用了哪些工具。如果你没有以有意义的方式使用 AI，请简短说明。

强烈优先使用本地或离线模型。

你仍然必须阅读、理解和测试你提交的所有内容。不接受未经审查的批量输出。

## 安全问题

不要在未经协调的情况下，将漏洞详情作为普通补丁通过 LXMF 发送。请使用[安全](security.md)中的流程。

## 开发环境设置

`task dev`、`task check` 和代码库布局请参阅[开发](development.md)。

## 下一步

- [开发](development.md)
- [常见问题](faq.md)
