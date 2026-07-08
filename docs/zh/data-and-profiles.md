# 数据与配置文件

Ren Browser 将书签、历史记录、标签页、设置和发现缓存存储在磁盘上。本页说明这些数据的存储位置及配置文件的工作原理。

## 默认位置

| 项目 | 路径 |
|------|------|
| 数据目录 | `~/.renbrowser/` |
| 主数据库 | `~/.renbrowser/renbrowser.db` |
| 插件 | `~/.renbrowser/plugins/<id>/` |
| 命名配置文件 | `~/.renbrowser/profiles/<name>/renbrowser.db` |
| 旧版状态（已迁移） | `~/.renbrowser/state.json` |

升级后首次启动时，`state.json` 会自动导入到 SQLite 中。

## SQLite 中存储的内容

典型的表和数据块包括：

- 打开的标签页和会话恢复状态
- 带时间戳的浏览历史记录
- 收藏夹
- 浏览器偏好和快捷键绑定
- 已缓存的发现条目
- 主题选择和自定义主题数据

打开时会检测损坏。UI 可能会提示重置数据库。重置会删除本地标签页、历史记录、收藏夹和设置。

## Reticulum 数据独立存储

身份密钥和接口配置存储在你的 Reticulum 目录中（默认 `~/.reticulum-go/`）。Ren Browser 读取该路径，但不会将你的 Reticulum 身份移入 `~/.renbrowser/`。

## 命名配置文件

使用 `--profile NAME` 或 `REN_BROWSER_PROFILE=NAME` 启动以使用：

```
~/.renbrowser/profiles/NAME/renbrowser.db
```

当你想在同一账户下拥有独立的历史记录（工作与个人，或测试用）时，请使用配置文件。

## 导入与导出

仅在启动时：

- `--export-profile /path/to/backup.json` 写入配置文件数据后退出
- `--import-profile /path/to/backup.json` 从文件合并或替换数据

环境变量镜像：`REN_BROWSER_EXPORT_PROFILE`、`REN_BROWSER_IMPORT_PROFILE`。

在重大升级前或迁移到新机器时使用导出功能。

## 服务器模式存储

| 模式 | 标签页、历史记录、收藏夹 |
|------|--------------------------|
| 默认服务器 | 服务器端 SQLite，位于服务器的 `~/.renbrowser/` |
| `--public-mode` | 每个客户端浏览器的 `localStorage` |

当多个用户共享一个服务器实例时，选择公共模式。

## 主题导入与导出

主题可以从设置中导出为 JSON，并在另一台安装中导入。主题文件不是完整的配置文件，只包含外观标记。

## 插件数据

拥有 `storage.plugin` 权限的扩展会获得以插件 id 为键的隔离存储。卸载插件并不总是删除其文件夹。如果需要彻底删除，请手动删除 `~/.renbrowser/plugins/<id>/`。

## Android

移动端构建在应用沙盒下使用相同的逻辑布局。路径因操作系统规则而有所不同，但数据库模式与桌面端一致。

## 备份清单

1. 停止 Ren Browser
2. 复制 `~/.renbrowser/renbrowser.db`（或你的配置文件路径）
3. 如果还想保留网状网络身份，复制 `~/.reticulum-go/`
4. 如果使用扩展，复制 `~/.renbrowser/plugins/`

将文件放回原处后，在下次启动前完成恢复。

## 下一步

- UI 路径请参阅[设置](settings.md)
- 公共模式请参阅[服务器模式](server-mode.md)
- 数据库无法打开请参阅[故障排查](troubleshooting.md)
