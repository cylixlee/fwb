# AGENTS.md

Working agreement for the fwb repository. Apply it to every change, refactor and review.

## 0. Language

- Talk to the user in the language the user is currently using.
- Every file is written in English: code, comments, tests, scripts, catalog keys, commit messages, documentation and local notes.
- The only exceptions are the conversation with the user and the translation catalogs (`internal/handler/catalogs/*.json`), which exist to hold non-English user-facing text. Vendored third-party content is left untouched.
- User-facing text never lives in Go code; it lives in catalogs.

## 1. Working rules

- Do not run commands that modify the user's environment (`fwb install-device`, Windows probes, SDK downloads, writes to `~/.config`) without explicit confirmation.
- Use the editor tools for file changes. Do not write ad-hoc Python scripts to rewrite source files.
- Do not modify this file unless the user explicitly asks.
- Keep changes scoped; do not mix unrelated refactors.
- Remove dead code, commented-out code and unused exported symbols.

## 2. Project invariants

- fwb is a **single binary on the WSL side**. Nothing is installed on Windows and no Windows process stays resident; Windows work is one-shot `powershell.exe` invocations through WSL interop.
- A custom device's `platform` is `linux-x64` (Flutter only allows `linux-x64`/`linux-arm64`).
- Engine switches are passed via `FLUTTER_ENGINE_SWITCH_*` environment variables; the Windows runner treats argv as Dart entrypoint arguments.
- WSL networking is expected to be `networkingMode=mirrored`; the VM service is reached directly, without a tunnel.
- WSL and Windows must share an engine revision. fwb provisions the official Windows SDK of the same version. Download/checksum/extract/build failures are fatal; there is no fallback.
- A project without `windows/` fails with instructions to run `flutter create --platforms=windows .`; fwb never generates it.
- The normal loop only edits `lib/`: Dart changes use VM-service hot reload and never touch the Windows disk. The mirror is synced and the native runner rebuilt only when native inputs change.

## 3. Layers and layout

```
cmd/fwb/main.go              fx wiring and exit code
internal/consts/             all named constants and enum types (leaf)
internal/domain/             entities, value objects, ports, error type
internal/srv/                use cases (bridge.go)
internal/infra/              port implementations
  flutter/                   cli.go, parse.go, releases.go
  config.go                  ConfigStore
  devices.go                 DeviceRepository
  downloader.go              Downloader
  mirror.go                  ProjectMirror
  powershell.go              WindowsHost + embedded scripts/
  sdk.go                     SDKStore
internal/handler/            CLI (cobra), i18n, catalogs/, templates/
pkg/                         reusable helpers; must not import internal/*
```

Dependencies flow `handler → srv → domain`, `infra → domain`. `consts` is a leaf package imported by everyone. `cmd/fwb` is the only place that imports all layers and wires fx.

## 4. Dependencies and utilities

Prefer, in order:

1. A maintained third-party package.
2. A helper in `pkg/` (when no library exists, or the wrapper carries project policy).
3. An inline implementation (last resort).

Never hand-roll standard infrastructure. Current precedent:

| Concern                    | Library                     |
| -------------------------- | --------------------------- |
| CLI commands, flags, help  | `spf13/cobra`               |
| i18n lookup and formatting | `golang.org/x/text/message` |
| system locale detection    | `jeandeaual/go-locale`      |
| shell-style tokenizing     | `mattn/go-shellwords`       |
| downloads                  | `cavaliergopher/grab/v3`    |
| HTTP requests              | `carlmjohnson/requests`     |
| progress bars              | `schollz/progressbar/v3`    |
| recursive copy             | `otiai10/copy`              |
| atomic writes              | `google/renameio/v2`        |
| slugs                      | `gosimple/slug`             |
| dependency injection       | `go.uber.org/fx`            |

`pkg/*` carries its own private constants and must not import `internal/*`.

## 5. Naming

Callers write the package name, so exported names must read correctly as `package.Identifier`.

```go
consts.Code          // bad
consts.ErrorCode     // good

handler.Handler      // bad
handler.CLI          // good

srv.Service          // bad
srv.Bridge           // good
```

- Blacklist: `Manager`, `Service`, `Handler`, `Helper`, `Util`, `Data`, `Info`, `State`, `Version`, `Report`, `Code`.
- Use role suffixes: `BuildState`, `FlutterVersion`, `DoctorReport`, `ConfigStore`, `ProjectMirrorFactory`, `WindowsHost`, `WindowsSDKReleaseSource`.
- Interfaces are named after capability: `...Store`, `...Repository`, `...Factory`, `...Source`, `...Host`, `...CLI`, `...Reporter`.
- Error codes use the `ErrorCode` prefix: `consts.ErrorCodeConfigRead`. Do not use `ErrXxx` (that is for sentinel errors).
- Constructors: `NewBridge`, `NewCLI`, `NewProject`, `NewError`, `WrapError`.
- Acronyms are upper-case: `SDK`, `WSL`, `CLI`, `HTTP`, `URL`, `ID`.

File names:

- Lower-case, no underscores except `_test.go`.
- Never concatenate words in a file name. Use a subdirectory: `infra/flutter/cli.go`, `infra/flutter/releases.go` — not `infra/fluttercli.go`.
- Name the file after its primary type or concept: `srv/bridge.go`, `handler/cli.go`, `infra/downloader.go` — not `service.go`, `handler.go`, `util.go`.
- One primary concept per file.
- A test file matches the file under test: `powershell_test.go`.

## 6. Constants

- All constants and enum types used by `internal/*` live in `internal/consts`, split by concern: `paths.go`, `commands.go`, `config.go`, `keys.go`, `errors.go`, `doctor.go`, `infra.go`, `cli.go`.
- No `const` declarations in `domain`, `srv`, `infra` or `handler`.
- No hard-coded values in logic: strings, numbers, durations, file modes, path segments, URLs, JSON keys, environment variable names, CLI flags, event codes, format strings and PowerShell parameters are named constants.
- A literal is only acceptable in a constant declaration, a struct tag, a test fixture, or as a zero value (`""`, `0`, `nil`).

## 7. Errors and i18n

- Cross-layer errors are `domain.Error{Code, Args, Cause}` created with `domain.NewError` / `domain.WrapError`.
- Every `ErrorCode` maps to the catalog key `err.<code>`.
- `srv` emits event codes only (`EventLaunchStarting`, `ProgressSDKDownload`, doctor entry keys); it never formats user-facing text.
- `handler.Translate` delegates to `golang.org/x/text/message`; `Cause` is translated recursively.
- Catalogs are embedded JSON and flat maps of key to format string. Supported placeholders are `%s`, `%d`, `%v`; **`%q` is not supported by x/text**.
- Add new messages to every language. Tests check key parity and placeholder parity.
- Key conventions: `err.<code>`, `progress.<code>`, `doctor.*`, `cli.*`.

## 8. CLI

- Built with cobra; no hand-written argument dispatch.
- `ping`, `install`, `uninstall`, `launch` are hidden custom-device hooks.
- Help/usage templates live in `internal/handler/templates/*.tmpl`, embedded with `go:embed`, and are localized through the `T` template function.
- `--help`/`-h` and `--version`/`-v` are cobra flags with localized descriptions.
- Argument validation uses `cobra.PositionalArgs` wrappers that return localized messages.
- Unknown-command errors are localized by a small pre-check; cobra exposes no hook for this.

## 9. Assets and scripts

- PowerShell scripts and catalogs are embedded with `//go:embed`; nothing is written to disk at runtime.
- PowerShell is rendered with `text/template` plus `q` (single-quote escaping), then executed as `powershell.exe -NoProfile -NonInteractive -EncodedCommand <base64 UTF-16LE>`.
- Never assemble scripts with `fmt.Sprintf`.
- Path conversion (`wslpath`) lives in `infra` only.

## 10. Resource management

- Release resources with `defer`: files, HTTP bodies, processes, timers, locks.
- Register the defer before the operation that may fail, e.g. `defer app.Stop(ctx)` before `app.Start(ctx)`.
- Never release manually on the success path. A manual release is only acceptable immediately before an operation that requires it (closing a file before `os.Rename` on Windows) and must still have a deferred release for error paths.
- Use named returns when a deferred close must surface its error, or when a deferred function mutates the result.
- `os.Exit` skips defers; keep teardown in a function that returns when it matters.

## 11. Checks and tests

```bash
make check
```

It runs `go build`, `go fix`, `go fmt`, `go vet` and `go test` over the module and must pass before committing. Do not run those commands by hand.

- `domain`: pure logic, table-driven.
- `handler`: catalog key parity, placeholder parity, language selection, argument formatting.
- `infra`: parsing helpers.
- `srv`: cover branches with small fake ports (fingerprint hit/miss, build failure, merge failure).
- Do not mock the standard library; implement small fake ports.

## 12. Forbidden

- Hard-coded user-facing strings in domain or application layers.
- Hand-rolling infrastructure that a maintained library provides.
- Manual resource release on success paths.
- Running commands that mutate the user's environment without confirmation.
- Committing `docs/` (local research notes, listed in `.gitignore`).
- Updating this file unless the user asks.
