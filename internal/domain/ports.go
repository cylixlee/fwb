package domain

import (
	"context"
	"io"
)

// Config is the settings store.
type ConfigStore interface {
	Get(key string) string
	Set(key, value string) error
	Unset(key string) error
	All() map[string]string
	Keys() []string
	Source(key string) string
	Path() string
}

// HostEnv is the subset of the Windows environment fwb needs.
type WindowsEnvironment struct {
	LocalAppDataWSL string
	LocalAppDataWin string
	UserProfileWSL  string
	OSVersion       string
}

// BuildRequest describes a `flutter build windows` invocation.
type WindowsBuildRequest struct {
	FlutterBatWindows string
	Mirror            string
	Env               map[string]string
}

// RunRequest describes starting the built Windows app.
type WindowsRunRequest struct {
	Exe         string
	Dir         string
	SessionFile string
	Switches    []string
}

// Host is the Windows interop gateway.
type WindowsHost interface {
	Available() error
	Info(ctx context.Context) (string, error)
	Env(ctx context.Context) (WindowsEnvironment, error)
	Mirrored(ctx context.Context) (bool, string, error)
	ToWindows(ctx context.Context, path string) (string, error)
	ToWSL(ctx context.Context, path string) (string, error)
	FlutterVersion(ctx context.Context, flutterBatWindows string) (FlutterVersion, error)
	Extract(ctx context.Context, zipWSL, destWSL string) error
	Build(ctx context.Context, req WindowsBuildRequest, stdout, stderr io.Writer) error
	Run(ctx context.Context, req WindowsRunRequest, stdout, stderr io.Writer) error
	Stop(ctx context.Context, pid int) error
}

// WSLFlutter is the Flutter SDK on PATH inside WSL.
type WSLFlutterCLI interface {
	Available() error
	Version(ctx context.Context) (FlutterVersion, error)
	ConfigureCustomDevices(ctx context.Context) error
}

// Release is one entry of the official Windows release index.
type WindowsSDKRelease struct {
	Version string
	Channel string
	Archive string
	SHA256  string
}

// ReleaseSource resolves official Windows portable SDK archives.
type WindowsSDKReleaseSource interface {
	Find(ctx context.Context, v FlutterVersion, storageBase string) (WindowsSDKRelease, error)
}

// Downloader fetches HTTP resources.
type Downloader interface {
	Bytes(ctx context.Context, url string) ([]byte, error)
	File(ctx context.Context, url, dest string, progress func(done, total int64)) error
}

// SDKStore manages the downloaded Windows SDK cache.
type SDKStore interface {
	Root(ctx context.Context) (string, error)
	Cached(ctx context.Context, v FlutterVersion) (*WindowsSDK, bool)
	Prepare(ctx context.Context, v FlutterVersion) (*WindowsSDK, error)
	MarkReady(ctx context.Context, v FlutterVersion) error
	Checksum(path string) (string, error)
}

// Mirror is the Windows-side project mirror.
type ProjectMirror interface {
	Path() string
	SessionFile() string
	SessionPID() (int, bool)
	WriteSession(pid int) error
	ClearSession()
	Stage(localPath string) error
	Sync() error
	Fingerprint(engineRevision string) (string, error)
	Locate(mode string) (dir, exe string, err error)
	Merge(buildDir string) error
	LoadState() (BuildState, error)
	SaveState(BuildState) error
	Remove() error
}

// MirrorFactory opens projects and their mirrors.
type ProjectMirrorFactory interface {
	Open(ctx context.Context) (Project, error)
	Root(ctx context.Context) (string, error)
	For(ctx context.Context, p Project) (ProjectMirror, error)
}

// DeviceRepository persists custom-device definitions.
type DeviceRepository interface {
	Install(DeviceSpec) (string, error)
	Remove(id string) (string, error)
}

// Reporter receives application-level events. The interface layer localizes
// the codes; the service layer never formats user-facing text itself.
type EventReporter interface {
	Event(code string, args ...any)
	StartProgress(code string, args ...any) OperationProgress
}

// Progress tracks a long-running operation.
type OperationProgress interface {
	Update(done, total int64)
	// Done marks the operation as complete.
	Done()
	// Abort drops the progress without signalling completion.
	Abort()
}
