# Flutter WSL Bridge

English | [简体中文](README.zh.md)

**Run a native Windows Flutter desktop app from WSL, with the normal `flutter run` hot-reload loop.**

Flutter WSL Bridge (FWB) is a single Linux binary that bridges the Flutter SDK in WSL and a Windows desktop build of the same project. Flutter keeps compiling Dart, watching files, and serving the VM service in WSL. FWB only does the low-frequency Windows work: provision a matching Windows SDK, build the native runner when needed, launch the `.exe`, and forward its output.

Nothing is installed on Windows. No Windows process stays resident. Windows work is done through one-shot embedded PowerShell calls over WSL interop.

```
VS Code / terminal (WSL)
  └─ flutter run -d fwb
       ├─ builds the asset bundle + kernel in WSL
       ├─ fwb install   -> stages the bundle
       └─ fwb launch    -> builds Windows when native inputs changed,
                           starts app.exe, streams its logs

  WSL  <---- VM Service / DevFS over 127.0.0.1 (mirrored networking) ---->  app.exe (Windows)
```

## Status

The MVP debug loop is validated end to end: matching Windows SDK download and native build; hot reload, hot restart, and asset sync through the VM service; engine switches, `--start-paused`, clean `q` shutdown; VS Code custom-device discovery and the reload button.

Expect rough edges. Release/profile builds, plugin edge cases, and incremental build polish are still on the roadmap.

## Why

Flutter's custom-device API lets an arbitrary embedder participate in `flutter run`. FWB registers itself as such a device. The Windows side only needs a native build and a process launch; the Dart toolchain never leaves WSL.

That avoids a second Flutter installation, a Windows-side helper process, port tunnels, and firewall rules. Mirrored WSL networking makes the Windows VM service reachable at `127.0.0.1` directly.

## Requirements

- Windows 11 with WSL 2 and **mirrored networking** (`networkingMode=mirrored` in `%USERPROFILE%\.wslconfig`).
- WSL interop enabled. `appendWindowsPath=false` is supported: FWB falls back to the absolute `powershell.exe` path.
- A Flutter SDK on WSL's `PATH` (`flutter --version --machine` must work).
- A Flutter project that contains a `windows/` directory. FWB never generates it; run `flutter create --platforms=windows .` if it is missing.
- Go 1.27 or newer to install FWB from source.

## Install

```bash
go install github.com/cylixlee/fwb/cmd/fwb@latest
```

From a checkout:

```bash
git clone https://github.com/cylixlee/fwb
cd fwb
go install ./cmd/fwb
```

The binary goes to `$(go env GOPATH)/bin`. Make sure that directory is on `PATH`.

Register the Flutter custom device:

```bash
fwb install-device
```

This writes `~/.config/flutter/custom_devices.json` and enables `flutter config --enable-custom-devices`. The hooks store the absolute path of the FWB binary, so re-run `fwb install-device` if you move or rebuild it somewhere else.

Verify:

```bash
flutter devices
```

`Windows desktop (via fwb)` should appear.

## Quick start

```bash
cd my_flutter_app
flutter run -d fwb
```

The first run downloads the official portable Windows SDK matching your WSL Flutter version (about 1.9 GB) into `%LOCALAPPDATA%\fwb\sdk`, builds the native runner, merges the asset bundle, and launches the app. Later runs reuse the build unless native inputs change.

Day-to-day edits in `lib/` use hot reload (`r`) and hot restart (`R`). Press `q` to stop the app. Ctrl-C also works, but current Flutter versions may print a spurious `The Dart compiler exited unexpectedly` trace during shutdown.

In VS Code, select the FWB device in the device picker and press F5. The reload button uses the same VM service.

## How version alignment works

The Flutter kernel is built by WSL's toolchain and loaded by the Windows engine. Engine revisions must match.

FWB reads `flutter --version --machine` in WSL, looks up the same framework version in the official Windows release index, and provisions that SDK under:

```
%LOCALAPPDATA%\fwb\sdk\<version>-<channel>
```

There is no downgrade or best-effort fallback. Download, checksum, extraction, and build failures are fatal and reported with an actionable message.

To reuse an existing Windows Flutter SDK instead:

```bash
fwb config set windows.flutter 'C:\src\flutter'
```

FWB verifies that its engine revision matches WSL and refuses to run otherwise.

## Commands

| Command                                         | Purpose                                                                                |
| ----------------------------------------------- | -------------------------------------------------------------------------------------- |
| `fwb install-device`                            | Register or update the Flutter custom device.                                          |
| `fwb uninstall-device`                          | Remove the custom device.                                                              |
| `fwb doctor`                                    | Check WSL, PowerShell, Windows paths, mirrored networking, mirror root, and SDK cache. |
| `fwb clean`                                     | Remove this project's Windows-side mirror.                                             |
| `fwb clean --sdk`                               | Also remove downloaded Windows SDKs.                                                   |
| `fwb config set <key> <value>`                  | Persist a setting.                                                                     |
| `fwb config get <key>` / `fwb config get --all` | Print one setting or all effective settings.                                           |
| `fwb config unset <key>`                        | Remove a persisted setting.                                                            |
| `fwb config list`                               | List settings with their source (env, file, default).                                  |
| `fwb version`                                   | Print the FWB version.                                                                 |
| `fwb help [command]`                            | Show help.                                                                             |

`ping`, `install`, `uninstall`, and `launch` are hidden hooks used by Flutter.

## Configuration

Settings live in `~/.config/fwb/config.json`. Every key can be overridden with an environment variable named `FWB_` plus the upper-cased key, with `.` and `-` replaced by `_`.

For example, on networks in China, the flutter-io.cn mirror is usually a good choice:

```bash
FWB_WINDOWS_STORAGEBASEURL=https://storage.flutter-io.cn flutter run -d fwb
```

| Key                      | Meaning                                                                                                   |
| ------------------------ | --------------------------------------------------------------------------------------------------------- |
| `windows.flutter`        | Existing Windows Flutter SDK root. Must have the same engine revision as WSL.                             |
| `windows.mirrorRoot`     | Root of the Windows-side mirror. Default `%LOCALAPPDATA%\fwb`.                                            |
| `windows.pubCache`       | Windows pub cache used for the native build. Default `<mirrorRoot>\pub-cache`.                            |
| `windows.storageBaseUrl` | Flutter artifact base URL. Defaults to `FLUTTER_STORAGE_BASE_URL`, then `https://storage.googleapis.com`. |
| `sync.extra`             | Comma-separated extra project directories to copy into the mirror.                                        |
| `device.id`              | Custom device id. Default `fwb`.                                                                          |
| `i18n.lang`              | UI language: `en` or `zh-CN`. Empty selects from the system locale.                                       |

Examples:

```bash
fwb config set windows.storageBaseUrl https://storage.flutter-io.cn
fwb config set sync.extra packages/shared
fwb config get --all
fwb config unset windows.storageBaseUrl
```

## Layout on Windows

```
%LOCALAPPDATA%\fwb\
  projects\<slug>-<hash>\   mirror, build output, staged assets, session
  sdk\<version>-<channel>\  provisioned Windows Flutter SDKs
  pub-cache\                default Windows pub cache
```

The mirror lives on a Windows-local disk, not on the 9p filesystem. Native builds happen there. `lib/` edits never touch the mirror.

## Troubleshooting

**FWB is missing from `flutter devices`**

```bash
fwb ping
flutter config --enable-custom-devices
flutter devices
```

**Mirrored networking is reported as off**

Add this to `%USERPROFILE%\.wslconfig`, then run `wsl --shutdown` from Windows:

```ini
[wsl2]
networkingMode=mirrored
```

**SDK download is slow, rate-limited, or gets a 429**

Choose another Flutter artifact mirror.

On networks in China, `storage.flutter-io.cn` is usually a good option:

```bash
fwb config set windows.storageBaseUrl https://storage.flutter-io.cn
```

Downloads use HTTP/1.1, keep a `.part` file, resume, and retry three times.

**`the project has no windows/ directory`**

```bash
flutter create --platforms=windows .
```

**`--start-paused` shows no window**

Expected. The app is paused before the first frame. The process is running; resume it with a debugger or drop the flag.

**Ctrl-C prints `The Dart compiler exited unexpectedly`**

A Flutter tool shutdown race, not FWB. The app still exits. Use `q` for a clean shutdown.

## Limitations

- Debug builds only. Release/profile Windows builds are not implemented.
- `path:` dependencies outside the project are not mirrored automatically; add their directories to `sync.extra` or use published packages.
- The project must build on Windows with its Windows plugins.
- The first native build is slow, especially with antivirus scanning `%LOCALAPPDATA%`.
- The VM service relies on mirrored networking. NAT-only setups are unsupported.
- The mirror is synced fully when native inputs change; there is no per-file incremental sync yet.

## Development

```bash
go fmt ./...
go build ./...
go vet ./...
go test ./...
```

All four must pass. Architecture, layering, naming, i18n, and resource rules are in [AGENTS.md](AGENTS.md).

## License

MIT. See [LICENSE](LICENSE).
