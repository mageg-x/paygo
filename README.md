## PayGo

支付网关系统。

### 默认数据库路径

| 平台 | 数据库路径 |
| --- | --- |
| Windows | `%APPDATA%\\paygo\\paygo.db` |
| macOS | `~/Library/Application Support/paygo/paygo.db` |
| Linux | `~/.paygo/paygo.db` |

说明：
- 启动时不传 `-db`，将自动使用上述平台默认路径。
- 也可通过 `-db` 显式指定数据库路径。

### 启动参数

```bash
./paygo-linux-amd64 -host 0.0.0.0 -port 8080
```

常用参数：
- `-db`：数据库文件路径（可选）
- `-host`：监听 IP（默认 `0.0.0.0`）
- `-port`：监听端口（默认 `8080`）
- `-migrate`：是否执行迁移

### 跨平台构建

```bash
# Linux CLI
make build-linux

# Linux GUI（托盘）
make build-linux-gui

# Windows GUI（托盘）
make build-windows-gui

# macOS GUI（托盘）
make build-macos-gui
```

macOS 打包脚本：

```bash
./scripts/build-macos.sh arm64
```

### 托盘与图标

- GUI 构建（`-tags gui`）会启用托盘，菜单可“打开首页/退出”。
- 图标源文件：`frontend/src/assets/paygo.png`
- 构建时会复制到 `server/src/systray/icon.png` 并生成 `icon.ico`。
