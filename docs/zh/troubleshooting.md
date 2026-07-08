# 故障排查

常见问题及首选排查步骤。

## Reticulum 无法启动

**症状：** 日志中出现 `reticulum start: ...`，发现面板为空，所有网状网络页面加载失败。

**检查：**

1. `about:` 中的配置路径与你的文件实际位置一致
2. `REN_BROWSER_CONFIG` 或 `--config` 指向有效文件
3. 接口定义的语法正确
4. 运行 Ren Browser 的用户对密钥和存储路径有读取权限

在应用外修复 Reticulum 配置，然后从设置中重新加载或重启应用。

## 发现面板为空

参见[发现](discovery.md)和 [Reticulum 配置](reticulum-setup.md)。

简短清单：

- 连接后等待公告出现
- 确认你的接口上存在对等节点
- 检查你配置的 UDP 或 TCP 端口的防火墙规则

## 页面加载超时

1. 验证地址栏中的哈希值
2. 从发现面板打开另一个节点
3. 在设置中确认 Reticulum 显示接口上的流量
4. 在无线电或网状网络路径变更后重试

## 数据库损坏或无法打开

**症状：** 关于配置文件数据的错误，出现重置数据库的提示。

**选项：**

1. 从备份恢复 `renbrowser.db`（[数据与配置文件](data-and-profiles.md)）
2. 通过 UI 重置（将销毁本地标签页、历史记录、收藏夹、设置）
3. 重命名损坏的文件，让 Ren Browser 创建新的数据库

浏览器数据库重置不会影响 Reticulum 身份。

## WASM 或 Micron 解析器错误

如果 Micron WASM 的 SRI 检查失败：

1. 不要禁用该检查
2. 从官方发布版本重新安装
3. 如果是从源码构建的，重新运行 `task build`，不要手动编辑 `frontend/dist/vendor/`

## 服务器模式：空白页面或错误资源

1. 检查 `--base-path` 是否与你的反向代理挂载点匹配
2. 当 TLS 在上游终止时，启用 `--trust-proxy`
3. 确认 Docker 中的端口映射（`-p 8080:8080`）

## 服务器模式：共享了不想共享的历史记录

使用 `--public-mode` 启动，使每个浏览器保留自己的 `localStorage` 副本。

## 扩展无法加载

1. 清单必须是 `renbrowser.plugin.json` 中有效的 JSON
2. `id` 必须与 `plugins/` 下的文件夹名称匹配
3. `engines.renbrowser` 必须满足你的应用版本要求
4. 未知的权限字符串会导致加载失败

请在设置中查看错误字符串。

## Android 构建失败

1. 设置 `ANDROID_HOME`
2. 运行 `task android:install:deps`
3. 按[安装](installation.md)中的说明使用 API 34 和 NDK r26+

## 开发：`task check` 失败

| 区域 | 命令 |
|------|------|
| Go 格式化 | `task fmt:go` |
| Go 测试 | `task test:go` |
| 前端 | `task frontend:check` |
| 安全扫描 | `task gosec` |

发送补丁前运行 `task check`。

## 仍然无法解决

1. 从 `about:` 记录你的版本号
2. 从终端或 Docker 捕获日志
3. 在你的网状网络社区提问，或通过项目渠道发送详细的错误报告

补丁提交方式请参阅[贡献](contributing.md)。
