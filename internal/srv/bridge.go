// Package srv implements the application layer: use cases that orchestrate
// domain ports. It never formats user-facing text and never touches Windows
// directly.
package srv

import (
	"context"
	"github.com/cylixlee/fwb/internal/consts"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cylixlee/fwb/internal/domain"
)

// Service is the application facade used by handlers.
type Bridge struct {
	cfg      domain.ConfigStore
	host     domain.WindowsHost
	flutter  domain.WSLFlutterCLI
	releases domain.WindowsSDKReleaseSource
	download domain.Downloader
	sdk      domain.SDKStore
	mirrors  domain.ProjectMirrorFactory
	devices  domain.DeviceRepository
}

// New wires the application service.
func NewBridge(
	cfg domain.ConfigStore,
	host domain.WindowsHost,
	flutter domain.WSLFlutterCLI,
	releases domain.WindowsSDKReleaseSource,
	download domain.Downloader,
	sdk domain.SDKStore,
	mirrors domain.ProjectMirrorFactory,
	devices domain.DeviceRepository,
) *Bridge {
	return &Bridge{
		cfg:      cfg,
		host:     host,
		flutter:  flutter,
		releases: releases,
		download: download,
		sdk:      sdk,
		mirrors:  mirrors,
		devices:  devices,
	}
}

// Ping is the custom-device ping hook; it must not spawn Windows processes.
func (s *Bridge) Ping(ctx context.Context) error {
	if err := s.host.Available(); err != nil {
		return err
	}
	return s.flutter.Available()
}

// Install snapshots the WSL asset bundle. Flutter does not stream this hook's
// output, so it must stay fast.
func (s *Bridge) Install(ctx context.Context, localPath string) error {
	m, err := s.mirror(ctx)
	if err != nil {
		return err
	}
	return m.Stage(localPath)
}

// Uninstall stops a previous app instance; failures are not fatal because
// Flutter calls this before every install.
func (s *Bridge) Uninstall(ctx context.Context) error {
	m, err := s.mirror(ctx)
	if err != nil {
		return nil
	}
	s.stop(ctx, m)
	return nil
}

// Launch performs the slow path: SDK, native build when needed, bundle merge,
// then starts the app and streams its logs.
func (s *Bridge) Launch(ctx context.Context, engineOptions string, out io.Writer, rep domain.EventReporter) (err error) {
	m, err := s.mirror(ctx)
	if err != nil {
		return
	}
	want, err := s.flutter.Version(ctx)
	if err != nil {
		return
	}
	event(rep, consts.EventLaunchWSLVersion, want.Framework, want.ShortEngine())

	sdk, err := s.ensureSDK(ctx, want, rep)
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
		event(rep, consts.EventLaunchPreparingBuild)
		if err := m.Sync(); err != nil {
			return err
		}
		req, err := s.buildRequest(ctx, sdk, m)
		if err != nil {
			return err
		}
		if err := s.host.Build(ctx, req, out, out); err != nil {
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
	s.stop(ctx, m)

	event(rep, consts.EventLaunchStarting, filepath.Base(exe))
	// The runner records its pid in the session file: stop the app when the log
	// stream fails, otherwise drop the stale session.
	defer func() {
		if err != nil {
			s.stop(ctx, m)
			return
		}
		m.ClearSession()
	}()
	err = s.host.Run(ctx, domain.WindowsRunRequest{
		Exe:         exe,
		Dir:         filepath.Dir(exe),
		SessionFile: m.SessionFile(),
		Switches:    domain.ParseEngineSwitches(engineOptions),
	}, out, out)
	return
}

func (s *Bridge) mirror(ctx context.Context) (domain.ProjectMirror, error) {
	p, err := s.mirrors.Open(ctx)
	if err != nil {
		return nil, err
	}
	return s.mirrors.For(ctx, p)
}

func (s *Bridge) stop(ctx context.Context, m domain.ProjectMirror) {
	pid, ok := m.SessionPID()
	if !ok {
		return
	}
	defer m.ClearSession()
	// The launch context is already canceled when Flutter quits the hook.
	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), consts.StopTimeout)
	defer cancel()
	s.host.Stop(stopCtx, pid)
}

func (s *Bridge) buildRequest(ctx context.Context, sdk *domain.WindowsSDK, m domain.ProjectMirror) (domain.WindowsBuildRequest, error) {
	bat, err := s.host.ToWindows(ctx, sdk.FlutterBat())
	if err != nil {
		return domain.WindowsBuildRequest{}, err
	}
	pubCache, err := s.pubCache(ctx)
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

func (s *Bridge) pubCache(ctx context.Context) (string, error) {
	if v := strings.TrimSpace(s.cfg.Get(consts.ConfigKeyWindowsPubCache)); v != "" {
		if strings.HasPrefix(v, consts.WSLPathPrefix) {
			return s.host.ToWindows(ctx, v)
		}
		return v, nil
	}
	root, err := s.mirrors.Root(ctx)
	if err != nil {
		return "", err
	}
	return s.host.ToWindows(ctx, filepath.Join(root, consts.PubCacheDirectoryName))
}

func (s *Bridge) ensureSDK(ctx context.Context, want domain.FlutterVersion, rep domain.EventReporter) (*domain.WindowsSDK, error) {
	if override := strings.TrimSpace(s.cfg.Get(consts.ConfigKeyWindowsFlutter)); override != "" {
		rootWSL := override
		if !strings.HasPrefix(rootWSL, consts.WSLPathPrefix) {
			var err error
			if rootWSL, err = s.host.ToWSL(ctx, override); err != nil {
				return nil, err
			}
		}
		sdk := &domain.WindowsSDK{Version: want, Root: rootWSL, FlutterRoot: rootWSL}
		if _, err := os.Stat(sdk.FlutterBat()); err != nil {
			return nil, domain.NewError(consts.ErrorCodeSDKOverride, override)
		}
		batWin, err := s.host.ToWindows(ctx, sdk.FlutterBat())
		if err != nil {
			return nil, err
		}
		got, err := s.host.FlutterVersion(ctx, batWin)
		if err != nil {
			return nil, domain.WrapError(consts.ErrorCodeSDKVersion, err, override)
		}
		if !got.SameEngine(want) {
			return nil, domain.NewError(consts.ErrorCodeFlutterOverride,
				want.Framework, want.EngineRevision, got.Framework, got.EngineRevision)
		}
		return sdk, nil
	}

	if cached, ok := s.sdk.Cached(ctx, want); ok {
		return cached, nil
	}

	storage := strings.TrimSpace(s.cfg.Get(consts.ConfigKeyWindowsStorageURL))
	if storage == "" {
		storage = os.Getenv(consts.EnvFlutterStorageBaseURL)
	}
	release, err := s.releases.Find(ctx, want, storage)
	if err != nil {
		return nil, err
	}
	base := strings.TrimRight(storage, consts.WSLPathPrefix)
	if base == "" {
		base = consts.DefaultStorageBaseURL
	}
	url := base + consts.ReleaseArchivePath + release.Archive

	event(rep, consts.EventSDKDownloading, want.Framework, want.Channel)
	sdkRoot, err := s.sdk.Root(ctx)
	if err != nil {
		return nil, err
	}
	archive := filepath.Join(sdkRoot, consts.SDKDownloadsDirectory, filepath.Base(release.Archive))
	// Register the cleanup before the download: a failed or aborted transfer
	// must not leave a multi-gigabyte archive behind.
	defer os.Remove(archive)
	meter := startProgress(rep, consts.ProgressSDKDownload, want.Framework)
	if err := s.downloadSDK(ctx, url, archive, meter); err != nil {
		return nil, domain.WrapError(consts.ErrorCodeSDKDownload, err, url)
	}
	if release.SHA256 != "" {
		sum, err := s.sdk.Checksum(archive)
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(sum, release.SHA256) {
			return nil, domain.NewError(consts.ErrorCodeSDKChecksum, archive)
		}
	}

	sdk, err := s.sdk.Prepare(ctx, want)
	if err != nil {
		return nil, err
	}
	event(rep, consts.EventSDKExtracting, sdk.Root)
	if err := s.host.Extract(ctx, archive, sdk.Root); err != nil {
		return nil, domain.WrapError(consts.ErrorCodeSDKExtract, err)
	}
	if _, err := os.Stat(sdk.FlutterBat()); err != nil {
		return nil, domain.NewError(consts.ErrorCodeSDKMissingBat, sdk.FlutterBat())
	}
	if err := s.sdk.MarkReady(ctx, want); err != nil {
		return nil, err
	}
	event(rep, consts.EventSDKReady)
	return sdk, nil
}

// downloadSDK streams the archive through the meter. The meter is finished on
// every path: a failed transfer aborts it instead of claiming completion.
func (s *Bridge) downloadSDK(ctx context.Context, url, archive string, meter domain.OperationProgress) (err error) {
	defer func() {
		if err != nil {
			meter.Abort()
			return
		}
		meter.Done()
	}()
	return s.download.File(ctx, url, archive, meter.Update)
}

// RegisterDevice stores the custom-device entry. The label is supplied by the
// interface layer so it can be localized. The boolean reports whether the
// flutter config flag could be enabled automatically.
func (s *Bridge) RegisterDevice(ctx context.Context, label, sdkName string) (string, bool, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", false, err
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	path, err := s.devices.Install(domain.DeviceSpec{
		ID:       s.cfg.Get(consts.ConfigKeyDeviceID),
		Label:    label,
		SDKName:  sdkName,
		Platform: consts.DevicePlatformLinuxX64,
		Exe:      exe,
	})
	if err != nil {
		return "", false, err
	}
	enabled := s.flutter.ConfigureCustomDevices(ctx) == nil
	return path, enabled, nil
}

// UnregisterDevice removes the custom-device entry.
func (s *Bridge) UnregisterDevice(ctx context.Context) (string, error) {
	return s.devices.Remove(s.cfg.Get(consts.ConfigKeyDeviceID))
}

// Clean removes project state, and optionally the SDK cache.
func (s *Bridge) Clean(ctx context.Context, withSDK bool) ([]string, error) {
	m, err := s.mirror(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.Remove(); err != nil {
		return nil, err
	}
	removed := []string{m.Path()}
	if withSDK {
		root, err := s.mirrors.Root(ctx)
		if err != nil {
			return removed, err
		}
		dir := filepath.Join(root, consts.SDKDirectoryName)
		if err := os.RemoveAll(dir); err != nil {
			return removed, err
		}
		removed = append(removed, dir)
	}
	return removed, nil
}

// Doctor collects diagnostics. The handler localizes and prints them.
func (s *Bridge) Doctor(ctx context.Context) domain.DoctorReport {
	report := domain.DoctorReport{}

	v, versionErr := s.flutter.Version(ctx)
	if versionErr != nil {
		report.Add(consts.DoctorLevelFail, consts.DoctorWSLFailure, versionErr)
	} else {
		report.Add(consts.DoctorLevelOK, consts.DoctorWSLSuccess, v.Framework, v.Channel, v.ShortEngine())
	}

	if err := s.host.Available(); err != nil {
		report.Add(consts.DoctorLevelFail, consts.DoctorPowerShellFailure, err)
		return report
	}
	if info, err := s.host.Info(ctx); err != nil {
		report.Add(consts.DoctorLevelFail, consts.DoctorPowerShellFailure, err)
	} else {
		report.Add(consts.DoctorLevelOK, consts.DoctorPowerShellSuccess, info)
	}

	if env, err := s.host.Env(ctx); err != nil {
		report.Add(consts.DoctorLevelWarn, consts.DoctorWindowsFailure, err)
	} else {
		report.Add(consts.DoctorLevelOK, consts.DoctorWindowsVersion, env.OSVersion)
		report.Add(consts.DoctorLevelOK, consts.DoctorLocalAppData, env.LocalAppDataWin)
	}

	if mirrored, path, err := s.host.Mirrored(ctx); err != nil {
		report.Add(consts.DoctorLevelWarn, consts.DoctorMirroredFailure, err)
	} else if mirrored {
		report.Add(consts.DoctorLevelOK, consts.DoctorMirroredSuccess, path)
	} else {
		report.Add(consts.DoctorLevelWarn, consts.DoctorMirroredWarning, path)
		report.Add(consts.DoctorLevelInfo, consts.DoctorMirroredHint, path)
	}

	if root, err := s.mirrors.Root(ctx); err != nil {
		report.Add(consts.DoctorLevelWarn, consts.DoctorMirrorRootFailure, err)
	} else {
		report.Add(consts.DoctorLevelOK, consts.DoctorMirrorRoot, root)
	}

	if override := s.cfg.Get(consts.ConfigKeyWindowsFlutter); override != "" {
		report.Add(consts.DoctorLevelInfo, consts.DoctorOverride, override)
	}
	if versionErr == nil {
		if sdk, ok := s.sdk.Cached(ctx, v); ok {
			report.Add(consts.DoctorLevelOK, consts.DoctorSDKCached, sdk.FlutterRoot)
		} else {
			report.Add(consts.DoctorLevelInfo, consts.DoctorSDKMissing)
		}
	}
	report.Add(consts.DoctorLevelInfo, consts.DoctorConfigPath, s.cfg.Path())
	return report
}

// Config accessors used by the CLI.

func (s *Bridge) ConfigGet(key string) string       { return s.cfg.Get(key) }
func (s *Bridge) ConfigSet(key, value string) error { return s.cfg.Set(key, value) }
func (s *Bridge) ConfigUnset(key string) error      { return s.cfg.Unset(key) }
func (s *Bridge) ConfigAll() map[string]string      { return s.cfg.All() }
func (s *Bridge) ConfigKeys() []string              { return s.cfg.Keys() }
func (s *Bridge) ConfigSource(key string) string    { return s.cfg.Source(key) }

// DeviceID returns the configured custom-device id.
func (s *Bridge) DeviceID() string { return s.cfg.Get(consts.ConfigKeyDeviceID) }

// Language returns the configured UI language (may be empty).
func (s *Bridge) Language() string { return s.cfg.Get(consts.ConfigKeyLanguage) }

func event(rep domain.EventReporter, code string, args ...any) {
	if rep != nil {
		rep.Event(code, args...)
	}
}

func startProgress(rep domain.EventReporter, code string, args ...any) domain.OperationProgress {
	if rep == nil {
		return nopProgress{}
	}
	return rep.StartProgress(code, args...)
}

type nopProgress struct{}

func (nopProgress) Update(int64, int64) {}
func (nopProgress) Done()               {}
func (nopProgress) Abort()              {}
