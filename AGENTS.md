# AGENTS.md

Repository-wide rules for fwb. New code, refactors and reviews must follow this
document.

## 0. Language

- **Talking to the user**: use the language the user is currently using.
- **Every file is written in English.** Code, comments, tests, scripts, catalog
  keys, commit messages, documentation and local notes are all English. The
  only exceptions are the conversation with the user and the translation
  catalogs (`internal/handler/catalogs/*.json`), whose purpose is to hold
  non-English user-facing text. Vendored third-party content is left untouched.
- **User-facing text never lives in Go code**: it belongs in
  `internal/handler/catalogs/*.json` and must be provided for every supported
  language.
- Do not write non-English comments or error strings in Go sources.

## 1. Invariants

- fwb ships as a **single binary on the WSL side**. Nothing is installed on
  Windows and no Windows process is kept resident; every Windows operation is a
  short `powershell.exe` invocation through WSL interop.
- A custom device's `platform` must be `linux-x64` (Flutter only allows
  `linux-x64` / `linux-arm64`).
- Engine switches must be passed through `FLUTTER_ENGINE_SWITCH_*` environment
  variables; the Windows runner treats argv as Dart entrypoint arguments.
- Default WSL networking is `networkingMode=mirrored`; the VM service is reached
  directly and no port tunnel is used.
- WSL and Windows SDKs must share the same engine revision. fwb downloads the
  official Windows portable SDK matching the WSL version. **Download, checksum,
  extract and build failures are fatal; there is no fallback.**
- A project without `windows/` fails with instructions to run
  `flutter create --platforms=windows .`. fwb never generates it.
- The normal loop only edits `lib/`: Dart changes go through VM service hot
  reload and never touch the Windows disk. The mirror is synced and the native
  runner rebuilt only when native inputs (pubspec, `windows/`, SDK) change.

## 2. Layers

```
handler ──▶ srv ──▶ domain ◀── infra
                        ▲
                     consts (leaf)
```

- `domain`: entities, value objects, ports (interfaces), error codes.
  **Standard library only.**
- `srv`: use-case orchestration. Depends on domain ports only; never imports
  infra.
- `infra`: port implementations. `os/exec`, HTTP and filesystem access exist
  only here.
- `handler`: CLI parsing, i18n, progress/report rendering, error translation.
- `consts`: named constants and enum types. Leaf package: it must not import
  any other internal package; every layer may import it.
- `cmd/fwb`: the only place that imports every layer; owns fx wiring and the
  exit code.

Forbidden:

- `fmt.Printf`, `os.Stdout` or user-facing text in `domain` / `srv`.
- `os/exec`, `net/http` or PowerShell in `srv`.
- `infra` importing `srv` / `handler`.
- Using fx outside `cmd`.

## 3. Layout

```
cmd/fwb/main.go              fx wiring
internal/consts/             all named constants and enum types (leaf)
internal/domain/             entities / value objects / ports / error type
internal/srv/                use cases
internal/infra/              port implementations + scripts/*.ps1.tmpl
  flutter/                   Flutter adapters (cli.go, parse.go, releases.go)
internal/handler/            CLI + catalogs/*.json
pkg/                         reusable helpers (no internal imports)
```

## 4. Naming

Callers always write the package name, so exported identifiers must read
correctly as `package.Identifier`.

```go
consts.Code          // bad
consts.ErrorCode     // good

handler.Handler      // bad (no information, repeats the package name)
handler.CLI          // good

srv.Service          // bad
srv.Bridge           // good
```

Short names are only acceptable when the type lives in a sub-package whose name
supplies the meaning: `domain/error.Code` → `error.Code`.

- Blacklist: `Manager`, `Service`, `Handler`, `Helper`, `Util`, `Data`, `Info`,
  `State`, `Version`, `Report`, `Code`.
- Use role suffixes instead: `BuildState`, `FlutterVersion`, `DoctorReport`,
  `ConfigStore`, `ProjectMirrorFactory`, `WindowsHost`,
  `WindowsBuildRequest`, `WindowsSDKReleaseSource`.
- Name interfaces after capability: `...Store`, `...Repository`, `...Factory`,
  `...Source`, `...Host`, `...CLI`, `...Reporter`.
- Error codes use the `ErrorCode` prefix: `consts.ErrorCodeConfigRead`. Do not
  use `ErrXxx` (that is for sentinel errors).
- Constructors: `NewBridge`, `NewCLI`, `NewProject`, `NewError`, `WrapError`.
- Acronyms are upper-case: `SDK`, `WSL`, `CLI`, `HTTP`, `URL`, `ID`.

File names:

- Lower-case, no underscores except `_test.go`.
- Do not concatenate words in a file name. When a concept needs more than one
  word, use a subdirectory: `infra/flutter/cli.go`, `infra/flutter/releases.go`
  — not `infra/fluttercli.go`.
- Name the file after its primary type or concept: `srv/bridge.go`,
  `handler/cli.go`, `infra/downloader.go` — not `service.go`, `handler.go`,
  `util.go`.
- One primary concept per file. Split unrelated types into separate files.
- A test file matches the file under test: `powershell_test.go`.

## 5. Constants and literals

No hard-coded values in logic. Every non-trivial string, number, duration, file
mode, path segment, URL, JSON key, environment variable name, CLI flag, event
code, format string and PowerShell parameter is a named constant.

- **All** constants and enum types for internal code live in `internal/consts`.
  It is a leaf package: it must not import other internal packages.
- `pkg/*` packages carry their own private constants and must not import
  `internal/*`, including `internal/consts`.
- Split by concern: `paths.go`, `commands.go`, `config.go`, `keys.go`,
  `errors.go`, `doctor.go`, `infra.go`, `cli.go`.
- No `const` declarations in `domain`, `srv`, `infra` or `handler`.
- PowerShell scripts receive values through template fields; never embed paths,
  URLs, flags, encodings or timeouts in a `.ps1` template.
- A literal is only acceptable in a constant declaration, a struct tag, a test
  fixture, or as a zero value (`""`, `0`, `nil`).

## 6. Dependencies and utilities

When a capability is needed, prefer in this order:

1. A maintained third-party package.
2. A helper in `pkg/` (when no suitable library exists, or the wrapper carries
   project policy).
3. An inline implementation (last resort).

Do not hand-roll standard formats or infrastructure: byte sizes, semantic
versions, globbing, atomic writes, recursive copies. Focus on business logic.
`pkg/*` must not import `internal/*`.

## 7. Resource management

- Every acquired resource is released with `defer`: files, HTTP bodies,
  processes, timers, locks. Register the defer before the operation that may
  fail (for example `defer app.Stop(ctx)` before `app.Start(ctx)`).
- Never release manually on the success path. A manual release is only
  acceptable immediately before an operation that requires it (for example
  closing a file before `os.Rename` on Windows) and must still be paired with a
  deferred release for error paths.
- When a deferred close must surface its error, use a named return:

  ```go
  func copyFile(src, dst string) (err error) {
      out, err := os.Create(dst)
      if err != nil {
          return err
      }
      defer func() {
          if cerr := out.Close(); err == nil {
              err = cerr
          }
      }()
      _, err = io.Copy(out, src)
      return err
  }
  ```

- Deferred mutations of the result require named results as well. Example:
  `ParseEngineSwitches` appends the final token in a deferred flush, and the
  slice is a named return value so the caller sees it.

## 8. Errors

- Cross-layer errors use `domain.Error{Code, Args, Cause}`.
- Use `domain.NewError(code, args...)` / `domain.WrapError(code, cause, args...)`.
- Every `ErrorCode` maps to the i18n key `err.<code>`.
- Never swallow errors: either wrap them with context or explicitly ignore them
  with a comment explaining why.
- Do not concatenate user-facing prose into errors; pass paths, versions and
  other values as `Args`.

## 9. i18n

- The only source of user-facing text is `internal/handler/catalogs/*.json`,
  embedded with `go:embed`.
- `srv` emits event codes only: `rep.Event("launch.starting", name)`,
  `rep.StartProgress("sdk.download", version)`.
- `handler` translates codes and recursively translates `domain.Error.Cause`.
- New messages must be added to every language; tests verify matching keys and
  identical placeholder counts.
- Key conventions: `err.<code>`, `progress.<code>`, `doctor.*`.

## 10. Assets and scripts

- Static assets (PowerShell scripts, catalogs) are always embedded with
  `//go:embed`; never written to disk at runtime and never assembled from Go
  strings.
- Templates use `text/template` plus `q` (PowerShell single-quote escaping);
  building scripts with `fmt.Sprintf` is forbidden.
- Path conversion (`wslpath`) lives in `infra` only.

## 11. Checks

```bash
go fmt ./...      # use `go fmt`, not bare `gofmt`
go build ./...
go vet ./...
go test ./...
```

All four must pass before committing.

## 12. Tests

- `domain`: pure logic, table-driven.
- `handler`: catalog completeness, placeholder parity, language normalization.
- `infra`: parsing helpers.
- `srv`: cover branches with small fake ports (fingerprint hit/miss, build
  failure, merge failure).
- Do not mock the standard library; implement small fake ports instead.

## 13. Forbidden

- Dead code, commented-out code, unused exported symbols.
- Hard-coded user-facing strings in the domain or application layer.
- Running commands that modify the user's environment (`fwb install-device`,
  Windows probes in `doctor`, writing `~/.config`, downloading the SDK) without
  explicit confirmation.
- Committing `docs/`: it holds local research notes and is listed in
  `.gitignore`.
