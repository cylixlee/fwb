package consts

import "time"

// CLI names and flags.
const (
	CommandName = "fwb"

	FlagHelp    = "help"
	FlagVersion = "version"
	FlagAll     = "all"
	FlagSDK     = "sdk"

	ShorthandHelp    = "h"
	ShorthandVersion = "v"
	ShorthandAll     = "a"

	ArgumentSeparator = " "
)

// Catalog files.
const (
	CatalogDirectory = "catalogs"
	CatalogSuffix    = ".json"
)

// Environment variable used for an explicit language override.
const EnvLanguageOverride = "FWB_LANG"

// Catalog keys.
const (
	KeyCommandRoot            = "cli.root"
	KeyCommandPing            = "cli.ping"
	KeyCommandInstall         = "cli.install"
	KeyCommandUninstall       = "cli.uninstall"
	KeyCommandLaunch          = "cli.launch"
	KeyCommandConfig          = "cli.config"
	KeyCommandConfigSet       = "cli.config.set"
	KeyCommandConfigGet       = "cli.config.get"
	KeyCommandConfigUnset     = "cli.config.unset"
	KeyCommandConfigList      = "cli.config.list"
	KeyCommandInstallDevice   = "cli.installDevice"
	KeyCommandUninstallDevice = "cli.uninstallDevice"
	KeyCommandDoctor          = "cli.doctor"
	KeyCommandClean           = "cli.clean"
	KeyCommandVersion         = "cli.version"
	KeyCommandHelp            = "cli.help"

	KeyUnknownCommand = "error.unknownCommand"

	KeyFlagHelp     = "cli.flag.help"
	KeyFlagVersion  = "cli.flag.version"
	KeyFlagAll      = "cli.flag.all"
	KeyCleanFlagSDK = "cli.flag.sdk"

	KeyUsageLabel        = "cli.usage"
	KeyAvailableCommands = "cli.availableCommands"
	KeyFlagsLabel        = "cli.flags"
	KeyGlobalFlagsLabel  = "cli.globalFlags"
	KeyHelpHint          = "cli.helpHint"

	KeyUsageConfigSet   = "usage.config.set"
	KeyUsageConfigGet   = "usage.config.get"
	KeyUsageConfigUnset = "usage.config.unset"
	KeyUsageUninstall   = "usage.uninstall"
	KeyUsageInstall     = "usage.install"
	KeyUsageLaunch      = "usage.launch"

	KeyConfigSet = "config.set"

	KeyDeviceLabel      = "device.label"
	KeyDeviceSDKName    = "device.sdkName"
	KeyDeviceInstalled  = "device.installed"
	KeyDeviceRemoved    = "device.removed"
	KeyDeviceHintEnable = "device.hint.enable"
	KeyDeviceHintNext   = "device.hint.next"

	KeyCleanMirror = "clean.mirror"
	KeyCleanSDK    = "clean.sdk"

	KeyDoctorStatusOK   = "doctor.status.ok"
	KeyDoctorStatusWarn = "doctor.status.warn"
	KeyDoctorStatusFail = "doctor.status.fail"
	KeyDoctorStatusInfo = "doctor.status.info"
)

// Output formats and prefixes.
const (
	FormatVersion    = "fwb %s (%s)\n"
	FormatConfigLine = "%s=%q\t# %s\n"
	FormatDoctorLine = "  %s %s\n"
	FormatError      = "fwb: %s\n"
	FormatLogEvent   = "[fwb] %s\n"

	ErrorCauseSeparator = ": "
)

// Progress rendering.
const (
	ProgressThrottleInterval  = 2 * time.Second
	ProgressLogInterval       = 10 * time.Second
	ProgressBarWidth          = 20
	ProgressSpinnerType       = 14
	ProgressPercentScale      = 100
	ProgressDescriptionPrefix = "[fwb] "
)

// Byte counts shown next to download progress.
const (
	ByteCountStep              = 1000
	ByteCountFractionThreshold = 10
	ByteCountFractionScale     = 10

	ByteUnitByte     = "B"
	ByteUnitKilobyte = "KB"
	ByteUnitMegabyte = "MB"
	ByteUnitGigabyte = "GB"

	FormatByteCount         = "%d %s"
	FormatByteCountWhole    = "%.0f %s"
	FormatByteCountFraction = "%.1f %s"
)

// Process exit codes and fatal error output.
const (
	ExitSuccess       = 0
	ExitFailure       = 1
	ErrorOutputFormat = "fwb: %v\n"
)
