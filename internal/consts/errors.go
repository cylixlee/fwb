package consts

// ErrorCode identifies an error condition. The interface layer maps it to a
// localized message (see internal/handler/catalogs).
type ErrorCode string

// Error codes shared across layers.
const (
	ErrorCodeConfigRead     ErrorCode = "config.read"
	ErrorCodeConfigParse    ErrorCode = "config.parse"
	ErrorCodeConfigWrite    ErrorCode = "config.write"
	ErrorCodeConfigEmptyKey ErrorCode = "config.emptyKey"

	ErrorCodeWinNoPowerShell ErrorCode = "win.noPowerShell"
	ErrorCodeWinNoWslpath    ErrorCode = "win.noWslpath"
	ErrorCodeWinPath         ErrorCode = "win.path"
	ErrorCodeWinPowerShell   ErrorCode = "win.powershell"
	ErrorCodeWinEnv          ErrorCode = "win.env"
	ErrorCodeWinScript       ErrorCode = "win.script"

	ErrorCodeProjectNotFlutter ErrorCode = "project.notFlutter"
	ErrorCodeProjectNoWindows  ErrorCode = "project.noWindows"
	ErrorCodeProjectNoBundle   ErrorCode = "project.noBundle"
	ErrorCodeProjectNoStaged   ErrorCode = "project.stagedBundle"
	ErrorCodeProjectNoBuild    ErrorCode = "project.noBuild"
	ErrorCodeProjectSync       ErrorCode = "project.sync"

	ErrorCodeFlutterRun             ErrorCode = "flutter.run"
	ErrorCodeFlutterParse           ErrorCode = "flutter.parse"
	ErrorCodeFlutterIncomplete      ErrorCode = "flutter.incomplete"
	ErrorCodeFlutterReleaseIndex    ErrorCode = "flutter.releaseIndex"
	ErrorCodeFlutterReleaseParse    ErrorCode = "flutter.releaseParse"
	ErrorCodeFlutterReleaseEmpty    ErrorCode = "flutter.releaseEmpty"
	ErrorCodeFlutterReleaseNotFound ErrorCode = "flutter.releaseNotFound"
	ErrorCodeFlutterOverride        ErrorCode = "flutter.overrideMismatch"

	ErrorCodeSDKDownload   ErrorCode = "sdk.download"
	ErrorCodeSDKChecksum   ErrorCode = "sdk.checksum"
	ErrorCodeSDKExtract    ErrorCode = "sdk.extract"
	ErrorCodeSDKMissingBat ErrorCode = "sdk.missingBat"
	ErrorCodeSDKOverride   ErrorCode = "sdk.overrideInvalid"
	ErrorCodeSDKVersion    ErrorCode = "sdk.versionRead"

	ErrorCodeBuildFailed ErrorCode = "build.failed"
	ErrorCodeDeviceParse ErrorCode = "device.parse"
	ErrorCodeHTTPStatus  ErrorCode = "http.status"
)
