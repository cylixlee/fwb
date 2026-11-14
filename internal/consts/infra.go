package consts

import "time"

// External executables.
const (
	PowerShellExecutable = "powershell.exe"
	PwshExecutable       = "pwsh.exe"
	WslPathExecutable    = "wslpath"
	FlutterExecutable    = "flutter"
)

// PowerShell invocation arguments.
const (
	PowerShellFlagNoProfile      = "-NoProfile"
	PowerShellFlagNonInteractive = "-NonInteractive"
	PowerShellFlagEncodedCommand = "-EncodedCommand"
)

// wslpath arguments.
const (
	WslPathWindowsFlag = "-w"
	WslPathWSLFlag     = "-u"
)

// Embedded script names.
const (
	TarExecutable   = "tar.exe"
	ExpandArchive   = "Expand-Archive"
	TarFailedFormat = "tar.exe failed: {0}"

	EngineSwitchesVariable     = "FLUTTER_ENGINE_SWITCHES"
	EngineSwitchVariablePrefix = "FLUTTER_ENGINE_SWITCH_"
	SessionFileEncoding        = "ASCII"
	ProcessCloseTimeoutMillis  = 800
	StopTimeout                = 5 * time.Second

	ScriptEnvironment    = "env.ps1.tmpl"
	ScriptPowerShellInfo = "psinfo.ps1"
	ScriptFlutterVersion = "version.ps1.tmpl"
	ScriptExtract        = "extract.ps1.tmpl"
	ScriptBuild          = "build.ps1.tmpl"
	ScriptRun            = "run.ps1.tmpl"
	ScriptKill           = "kill.ps1.tmpl"

	ScriptsDirectory       = "scripts"
	ScriptsGlob            = "scripts/*.tmpl"
	ScriptTemplateSuffix   = ".tmpl"
	TemplateFunctionQuote  = "q"
	TemplateRootName       = "fwb"
	EmbedPathSeparator     = "/"
	FallbackPowerShellPath = "/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe"
	FallbackPwshPath       = "/mnt/c/Program Files/PowerShell/7/pwsh.exe"

	SDKDirectorySeparator = "-"
	EnvLocalAppData       = "LOCALAPPDATA"
	EnvUserProfile        = "USERPROFILE"
	GlobWildcard          = "*"
	PathListSeparator     = ","
	WSLNetworkingModeKey  = "networkingmode"
	WSLMirrorModeValue    = "mirrored"
	DoubleQuote           = "\""
	SingleQuote           = "'"
	EscapedSingleQuote    = "''"
)

// Flutter CLI arguments.
const (
	FlutterVersionFlag             = "--version"
	FlutterMachineFlag             = "--machine"
	FlutterConfigCommand           = "config"
	FlutterEnableCustomDevicesFlag = "--enable-custom-devices"
	FlutterWindowsBuildArgs        = "build windows --debug"
)

// Artifact URLs.
const ReleaseIndexPath = "/flutter_infra_release/releases/releases_windows.json"

// File and directory names.
const (
	AppDirectoryName      = "fwb"
	ConfigFileName        = "config.json"
	CacheDirectoryName    = "fwb"
	WinEnvCacheFileName   = "winenv.json"
	FlutterConfigDirName  = "flutter"
	CustomDevicesFileName = "custom_devices.json"
	WSLConfigFileName     = ".wslconfig"

	ProjectsDirectoryName  = "projects"
	SDKMarkerFileName      = ".fwb-sdk.json"
	FlutterSDKInnerDirName = "flutter"

	MetaDirectoryName    = ".fwb"
	StagingDirectoryName = "staging"
	SessionFileName      = "session.pid"
	StateFileName        = "state.json"

	PartFileSuffix = ".part"

	DefaultChannel  = "stable"
	DefaultLogLevel = "info"
)

// Source files and directories copied to the mirror.
const (
	PubspecFileName         = "pubspec.yaml"
	PubspecLockFileName     = "pubspec.lock"
	MetadataFileName        = ".metadata"
	AnalysisOptionsFileName = "analysis_options.yaml"
	LibDirectoryName        = "lib"
	AssetsDirectoryName     = "assets"
	FontsDirectoryName      = "fonts"
	ShadersDirectoryName    = "shaders"
)

// Runtime tuning.
const (
	DirectoryPermissions      = 0o755
	FilePermissions           = 0o644
	DownloadUserAgent         = "fwb"
	DownloadAttempts          = 3
	DownloadRetryDelay        = 2 * time.Second
	DownloadProgressInterval  = 500 * time.Millisecond
	WinEnvCacheTTL            = 12 * time.Hour
	HTTPResponseHeaderTimeout = 30 * time.Second
)

// Configuration helpers.
const (
	EnvPrefix           = "FWB_"
	EnvKeySeparator     = "."
	EnvKeyDash          = "-"
	EnvKeyReplacement   = "_"
	EnvSourcePrefix     = "env:"
	ConfigSourceFile    = "file"
	ConfigSourceDefault = "default"
	ConfigSourceUnset   = "unset"
)

// Text helpers.
const (
	CarriageReturnLineFeed  = "\r\n"
	LineFeed                = "\n"
	JSONObjectStart         = "{"
	InfoSeparator           = " / "
	FingerprintLineFormat   = "%s|%d|%d"
	FingerprintEngineFormat = "engine=%s\n"
	FingerprintFileFormat   = "%s=%x\n"
	ErrorCauseFormat        = "%s: %w"
)

// JSON keys used with generic maps.
const (
	JSONCustomDevices     = "custom-devices"
	JSONID                = "id"
	JSONLabel             = "label"
	JSONSDKNameAndVersion = "sdkNameAndVersion"
	JSONPlatform          = "platform"
	JSONEnabled           = "enabled"
	JSONFingerprint       = "fingerprint"
	JSONEngineRevision    = "engineRevision"
	JSONBuildDir          = "buildDir"
	JSONExe               = "exe"
	JSONBuiltAt           = "builtAt"
	JSONLocalAppDataWSL   = "LocalAppDataWSL"
	JSONLocalAppDataWin   = "LocalAppDataWin"
	JSONUserProfileWSL    = "UserProfileWSL"
	JSONOSVersion         = "OSVersion"
	JSONFetchedAt         = "FetchedAt"
)

// Version defaults.
const (
	DefaultBuildVersion     = "v0.0.0-dev"
	BuildSettingVCSRevision = "vcs.revision"
	VCSRevisionShortLength  = 12
)
