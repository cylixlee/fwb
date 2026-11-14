// Package srv implements the application layer: use cases that orchestrate
// domain ports. It never formats user-facing text and never touches Windows
// directly.
package srv

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// Launcher builds the mirror when its native inputs changed and streams the app.
type Launcher struct {
	cfg     domain.ConfigStore
	host    domain.WindowsHost
	flutter domain.WSLFlutterCLI
	mirrors domain.ProjectMirrorFactory
	sdks    *SDKProvider
}

// NewLauncher wires the launch use case.
func NewLauncher(
	cfg domain.ConfigStore,
	host domain.WindowsHost,
	flutter domain.WSLFlutterCLI,
	mirrors domain.ProjectMirrorFactory,
	sdks *SDKProvider,
) *Launcher {
	return &Launcher{cfg: cfg, host: host, flutter: flutter, mirrors: mirrors, sdks: sdks}
}

// Launch performs the slow path: SDK, native build when needed, bundle merge,
// then starts the app and streams its logs.
func (l *Launcher) Launch(ctx context.Context, engineOptions string, out io.Writer, rep domain.EventReporter) (err error) {
	m, err := l.mirrors.Current(ctx)
	if err != nil {
		return
	}
	want, err := l.flutter.Version(ctx)
	if err != nil {
		return
	}
	emit(rep, consts.EventLaunchWSLVersion, want.Framework, want.ShortEngine())

	windowsSDK, err := l.sdks.Ensure(ctx, want, rep)
	if err != nil {
		return
	}

	fingerprint, err := m.Fingerprint(want.EngineRevision)
	if err != nil {
		return
	}
	state, err := m.LoadState()
	if err != nil {
		return
	}
	buildDir, exe, locateErr := m.Locate(consts.DebugBuildMode)
	needsBuild := locateErr != nil ||
		state.Fingerprint != fingerprint ||
		state.EngineRevision != want.EngineRevision ||
		state.Exe != exe
	if needsBuild {
		emit(rep, consts.EventLaunchPreparingBuild)
		if err := m.Sync(); err != nil {
			return err
		}
		req, err := l.buildRequest(ctx, windowsSDK, m)
		if err != nil {
			return err
		}
		if err := l.host.Build(ctx, req, out, out); err != nil {
			return domain.WrapError(consts.ErrorCodeBuildFailed, err)
		}
		if buildDir, exe, err = m.Locate(consts.DebugBuildMode); err != nil {
			return err
		}
		state = domain.BuildState{
			Fingerprint:    fingerprint,
			EngineRevision: want.EngineRevision,
			BuildDir:       buildDir,
			Exe:            exe,
			BuiltAt:        time.Now(),
		}
		if err := m.SaveState(state); err != nil {
			return err
		}
	}

	if err := m.Merge(buildDir); err != nil {
		return err
	}
	stopSession(ctx, l.host, m)

	emit(rep, consts.EventLaunchStarting, filepath.Base(exe))
	// The runner records its pid in the session file: stop the app when the log
	// stream fails, otherwise drop the stale session.
	defer func() {
		if err != nil {
			stopSession(ctx, l.host, m)
			return
		}
		m.ClearSession()
	}()
	err = l.host.Run(ctx, domain.WindowsRunRequest{
		Exe:         exe,
		Dir:         filepath.Dir(exe),
		SessionFile: m.SessionFile(),
		Switches:    domain.ParseEngineSwitches(engineOptions),
	}, out, out)
	return
}

func (l *Launcher) buildRequest(ctx context.Context, windowsSDK *domain.WindowsSDK, m domain.ProjectMirror) (domain.WindowsBuildRequest, error) {
	bat, err := l.host.ToWindows(ctx, windowsSDK.FlutterBat())
	if err != nil {
		return domain.WindowsBuildRequest{}, err
	}
	pubCache, err := l.pubCache(ctx)
	if err != nil {
		return domain.WindowsBuildRequest{}, err
	}
	env := map[string]string{
		consts.EnvPubCache:              pubCache,
		consts.EnvPubHostedURL:          os.Getenv(consts.EnvPubHostedURL),
		consts.EnvFlutterStorageBaseURL: os.Getenv(consts.EnvFlutterStorageBaseURL),
	}
	return domain.WindowsBuildRequest{FlutterBatWindows: bat, Mirror: m.Path(), Env: env}, nil
}

func (l *Launcher) pubCache(ctx context.Context) (string, error) {
	if v := strings.TrimSpace(l.cfg.Get(consts.ConfigKeyWindowsPubCache)); v != "" {
		if strings.HasPrefix(v, consts.WSLPathPrefix) {
			return l.host.ToWindows(ctx, v)
		}
		return v, nil
	}
	root, err := l.mirrors.Root(ctx)
	if err != nil {
		return "", err
	}
	return l.host.ToWindows(ctx, filepath.Join(root, consts.PubCacheDirectoryName))
}
