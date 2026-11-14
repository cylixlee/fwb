package consts

// I18n key prefixes.
const (
	ErrorKeyPrefix    = "err."
	ProgressKeyPrefix = "progress."
)

// Keys shared by every progress code.
const (
	KeyProgressLine = "progress.line"
)

// Application events emitted through the event reporter.
const (
	EventLaunchWSLVersion     = "launch.wslVersion"
	EventLaunchPreparingBuild = "launch.preparingBuild"
	EventLaunchStarting       = "launch.starting"
	EventSDKDownloading       = "sdk.downloading"
	EventSDKExtracting        = "sdk.extracting"
	EventSDKReady             = "sdk.ready"
	ProgressSDKDownload       = "sdk.download"
)

// Doctor report entry keys.
const (
	DoctorWSLSuccess        = "doctor.wsl.ok"
	DoctorWSLFailure        = "doctor.wsl.fail"
	DoctorPowerShellSuccess = "doctor.ps.ok"
	DoctorPowerShellFailure = "doctor.ps.fail"
	DoctorWindowsVersion    = "doctor.windows"
	DoctorWindowsFailure    = "doctor.windows.err"
	DoctorLocalAppData      = "doctor.localappdata"
	DoctorMirroredSuccess   = "doctor.mirrored.ok"
	DoctorMirroredWarning   = "doctor.mirrored.warn"
	DoctorMirroredHint      = "doctor.mirrored.hint"
	DoctorMirroredFailure   = "doctor.mirrored.err"
	DoctorMirrorRoot        = "doctor.mirrorroot"
	DoctorMirrorRootFailure = "doctor.mirrorroot.err"
	DoctorOverride          = "doctor.override"
	DoctorSDKCached         = "doctor.sdk.cached"
	DoctorSDKMissing        = "doctor.sdk.missing"
	DoctorConfigPath        = "doctor.config"
)
