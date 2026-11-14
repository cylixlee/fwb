// Package consts holds every named constant and its underlying enum types.
// It is a leaf package: it must not import any other internal package.
package consts

// File system names shared by the domain and its adapters.
const (
	FlutterBinDirectory = "bin"
	FlutterBatFileName  = "flutter.bat"
	ExeSuffix           = ".exe"

	FlutterAssetsDir = "flutter_assets"
	DataDirectory    = "data"
	BuildDirectory   = "build"
	WindowsDirectory = "windows"
	RunnerDirectory  = "runner"
	DebugBuildMode   = "Debug"

	CMakeListsFileName  = "CMakeLists.txt"
	BinaryNameDirective = "set(binary_name"
)

// Custom-device hook names and interpolation placeholders.
const (
	HookPing      = "ping"
	HookUninstall = "uninstall"
	HookInstall   = "install"
	HookRunDebug  = "runDebug"

	PlaceholderAppName       = "${appName}"
	PlaceholderLocalPath     = "${localPath}"
	PlaceholderEngineOptions = "${engineOptions}"
)

// Defaults.
const (
	DefaultDeviceID    = "fwb"
	EngineSwitchPrefix = "--"
)

// Mirror identity.
const (
	MirrorIDHashLength   = 12
	MirrorIDFallbackSlug = "project"
	MirrorIDSeparator    = "-"
	ShortEngineLength    = 12
)

// Error text format used by domain.Error.Error().
const ErrorTextFormat = "%s: %v"

// Artifact storage and mirror layout.
const (
	DefaultStorageBaseURL = "https://storage.googleapis.com"
	ReleaseArchivePath    = "/flutter_infra_release/releases/"
	SDKDownloadsDirectory = "downloads"
	SDKDirectoryName      = "sdk"
	PubCacheDirectoryName = "pub-cache"
	WSLPathPrefix         = "/"
)

// DevicePlatformLinuxX64 is the only platform Flutter's custom-device config
// allows.
const DevicePlatformLinuxX64 = "linux-x64"
