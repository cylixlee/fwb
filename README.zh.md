# Flutter WSL Bridge

[English](README.md) | 简体中文

**在 WSL 里运行原生 Windows Flutter 桌面应用，同时保留正常的 `flutter run` 热重载循环。**

Flutter WSL Bridge（FWB）是一个 Linux 单二进制，把 WSL 里的 Flutter SDK 和同一项目的 Windows 桌面构建连接起来。Dart 编译、文件监听、VM Service 仍然全部在 WSL 里完成；FWB 只负责低频的 Windows 操作：准备同版本 Windows SDK、按需构建原生 runner、启动 `.exe`、转发输出。

Windows 侧不安装任何东西，也没有常驻进程。所有 Windows 操作都是通过 WSL interop 执行的一次性内嵌 PowerShell 脚本。

```
VS Code / 终端（WSL）
  └─ flutter run -d fwb
       ├─ 在 WSL 里构建 asset bundle + kernel
       ├─ fwb install   -> 暂存 bundle
       └─ fwb launch    -> 原生输入变化时构建 Windows，
                           启动 app.exe，流式转发日志

  WSL  <---- VM Service / DevFS，经 127.0.0.1 直连（mirrored 网络） ---->  app.exe（Windows）
```

## 状态

MVP 调试循环已经端到端验证：下载同版本 Windows SDK 并完成原生构建；通过 VM Service 热重载、热重启、同步资产；engine switches、`--start-paused`、按 `q` 干净退出；VS Code 设备发现和 reload 按钮。

仍会有粗糙的地方。release/profile 构建、插件边界情况、增量构建优化还在路线图上。

## 为什么需要它

Flutter 的 custom-device API 允许任意 embedder 接入 `flutter run`。FWB 把自己注册成这样一个设备。Windows 侧只需要一次原生构建和一次进程启动，Dart 工具链始终留在 WSL。

这样就不需要第二套 Flutter、Windows 侧常驻 helper、端口隧道或防火墙规则。WSL 的 mirrored 网络让 Windows 的 VM Service 可以直接通过 `127.0.0.1` 访问。

## 环境要求

- Windows 11 + WSL 2，并启用 **mirrored 网络**（`%USERPROFILE%\.wslconfig` 里的 `networkingMode=mirrored`）。
- WSL interop 开启。支持 `appendWindowsPath=false`：FWB 会回退到 `powershell.exe` 的绝对路径。
- WSL 的 `PATH` 里有 Flutter SDK（`flutter --version --machine` 能正常运行）。
- 项目里有 `windows/` 目录。FWB 不会替你生成；缺失时先执行 `flutter create --platforms=windows .`。
- 从源码安装需要 Go 1.27 或更新版本。

## 安装

```bash
go install github.com/cylixlee/fwb/cmd/fwb@latest
```

从仓库安装：

```bash
git clone https://github.com/cylixlee/fwb
cd fwb
go install ./cmd/fwb
```

二进制会装到 `$(go env GOPATH)/bin`，确保它在 `PATH` 里。

注册 Flutter custom device：

```bash
fwb install-device
```

它会写入 `~/.config/flutter/custom_devices.json`，并执行 `flutter config --enable-custom-devices`。hook 里保存的是 FWB 二进制的绝对路径，所以移动或换路径重新编译后，要重新执行 `fwb install-device`。

确认：

```bash
flutter devices
```

应该出现 `Windows desktop (via fwb)`。

## 快速开始

```bash
cd my_flutter_app
flutter run -d fwb
```

首次运行会下载与 WSL Flutter 版本一致的官方 Windows 便携 SDK（约 1.9 GB）到 `%LOCALAPPDATA%\fwb\sdk`，构建原生 runner、合并 asset bundle、启动应用。之后只要原生输入没变，就复用已有构建。

日常只改 `lib/` 时，用 `r` 热重载、`R` 热重启，按 `q` 停止。Ctrl-C 也可以，但当前 Flutter 版本在关闭时可能多打印一段 `The Dart compiler exited unexpectedly`。

VS Code 里在设备下拉中选 FWB，按 F5；reload 按钮走同一个 VM Service。

## 版本对齐

Flutter kernel 由 WSL 工具链编译，由 Windows engine 加载，engine revision 必须一致。

FWB 读取 WSL 里的 `flutter --version --machine`，在官方 Windows release index 里找到同一个 framework 版本，并把对应 SDK 放到：

```
%LOCALAPPDATA%\fwb\sdk\<version>-<channel>
```

没有降级，也没有“尽力而为”的回退。下载、校验、解压、构建失败都会直接报错并给出可操作的信息。

如果想复用已有的 Windows Flutter SDK：

```bash
fwb config set windows.flutter 'C:\src\flutter'
```

FWB 会校验它的 engine revision 是否与 WSL 一致，不一致就拒绝运行。

## 命令

| 命令                                            | 作用                                                                          |
| ----------------------------------------------- | ----------------------------------------------------------------------------- |
| `fwb install-device`                            | 注册或更新 Flutter custom device。                                            |
| `fwb uninstall-device`                          | 移除 custom device。                                                          |
| `fwb doctor`                                    | 检查 WSL、PowerShell、Windows 路径、mirrored 网络、mirror 根目录和 SDK 缓存。 |
| `fwb clean`                                     | 删除当前项目的 Windows 侧 mirror。                                            |
| `fwb clean --sdk`                               | 同时删除已下载的 Windows SDK。                                                |
| `fwb config set <key> <value>`                  | 持久化设置。                                                                  |
| `fwb config get <key>` / `fwb config get --all` | 打印单个或全部生效设置。                                                      |
| `fwb config unset <key>`                        | 删除已持久化的设置。                                                          |
| `fwb config list`                               | 列出设置及其来源（env、file、default）。                                      |
| `fwb version`                                   | 打印 FWB 版本。                                                               |
| `fwb help [command]`                            | 显示帮助。                                                                    |

`ping`、`install`、`uninstall`、`launch` 是 Flutter 调用的隐藏 hook。

## 配置

配置位于 `~/.config/fwb/config.json`。每个 key 都可以用环境变量覆盖：`FWB_` 加上大写 key，`.` 和 `-` 替换成 `_`。

例如，在中国网络条件下，使用 flutter-io 的镜像通常会更顺畅：

```bash
FWB_WINDOWS_STORAGEBASEURL=https://storage.flutter-io.cn flutter run -d fwb
```

| Key                      | 含义                                                                                           |
| ------------------------ | ---------------------------------------------------------------------------------------------- |
| `windows.flutter`        | 已有 Windows Flutter SDK 根目录，engine revision 必须与 WSL 一致。                             |
| `windows.mirrorRoot`     | Windows 侧 mirror 根目录，默认 `%LOCALAPPDATA%\fwb`。                                          |
| `windows.pubCache`       | 原生构建使用的 Windows pub cache，默认 `<mirrorRoot>\pub-cache`。                              |
| `windows.storageBaseUrl` | Flutter 产物基地址。默认取 `FLUTTER_STORAGE_BASE_URL`，否则 `https://storage.googleapis.com`。 |
| `sync.extra`             | 额外同步到 mirror 的项目目录，逗号分隔。                                                       |
| `device.id`              | custom device id，默认 `fwb`。                                                                 |
| `i18n.lang`              | UI 语言：`en` 或 `zh-CN`。留空时按系统 locale 选择。                                           |

示例：

```bash
fwb config set windows.storageBaseUrl https://storage.flutter-io.cn
fwb config set sync.extra packages/shared
fwb config get --all
fwb config unset windows.storageBaseUrl
```

## Windows 侧目录

```
%LOCALAPPDATA%\fwb\
  projects\<slug>-<hash>\   mirror、构建输出、暂存资产、session
  sdk\<version>-<channel>\  已准备的 Windows Flutter SDK
  pub-cache\                默认 Windows pub cache
```

mirror 位于 Windows 本地磁盘，不在 9p 文件系统上；原生构建也在那里进行。改 `lib/` 不会碰 mirror。

## 排查

**`flutter devices` 里没有 FWB**

```bash
fwb ping
flutter config --enable-custom-devices
flutter devices
```

**提示 mirrored 网络未开启**

在 `%USERPROFILE%\.wslconfig` 中加入以下内容，然后在 Windows 执行 `wsl --shutdown`：

```ini
[wsl2]
networkingMode=mirrored
```

**SDK 下载慢、被限流或返回 429**

换一个 Flutter 产物镜像。

例如，在中国网络条件下，`storage.flutter-io.cn` 通常是个不错的选择：

```bash
fwb config set windows.storageBaseUrl https://storage.flutter-io.cn
```

下载走 HTTP/1.1，保留 `.part` 文件，支持续传，并重试 3 次。

**提示 `the project has no windows/ directory`**

```bash
flutter create --platforms=windows .
```

**`--start-paused` 看不到窗口**

这是预期行为：应用停在首帧之前，进程仍在运行。用调试器恢复，或者去掉这个参数。

**Ctrl-C 打印 `The Dart compiler exited unexpectedly`**

这是 Flutter tool 关闭时的竞态，不是 FWB 的问题，应用仍会退出。日常请用 `q` 退出。

## 限制

- 只支持 debug 构建，release/profile 尚未实现。
- 项目外的 `path:` 依赖不会自动镜像；把目录加入 `sync.extra`，或改用已发布包。
- 项目必须能在 Windows 上用它的 Windows 插件正常构建。
- 首次原生构建较慢，尤其是杀毒软件扫描 `%LOCALAPPDATA%` 时。
- VM Service 依赖 mirrored 网络；只有 NAT 的环境不支持。
- 原生输入变化时会全量同步 mirror，还没有逐文件增量同步。

## 开发

```bash
go fmt ./...
go build ./...
go vet ./...
go test ./...
```

四项必须全绿。分层、命名、i18n、资源管理规范见 [AGENTS.md](AGENTS.md)。

## 许可证

MIT，见 [LICENSE](LICENSE)。
