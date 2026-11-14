package consts

import "time"

// CLI flags and commands that are not part of the shared command vocabulary.
const (
	FlagVersion      = "--version"
	FlagVersionShort = "-v"
	FlagHelp         = "--help"
	FlagHelpShort    = "-h"
	FlagAll          = "--all"
	FlagAllShort     = "-a"
	FlagSDK          = "sdk"

	ArgumentSeparator = " "
	EmptyUsage        = ""
)

// Catalog files.
const (
	CatalogDirectory = "catalogs"
	CatalogSuffix    = ".json"
)

// Language detection and normalization.
const (
	DefaultLanguage                = "en"
	LanguageEncodingSeparator byte = '.'
	LanguageUnderscore             = "_"
	LanguageDash                   = "-"
	ChineseLanguage                = "zh-CN"
	ChineseLanguageShort           = "zh"
	ChineseLanguageCN              = "zh-cn"
	ChineseLanguageHans            = "zh-hans"
	ChineseLanguageHansCN          = "zh-hans-cn"
	EnglishLanguage                = "en"
	EnglishLanguageUS              = "en-us"
	EnglishLanguageGB              = "en-gb"
	LanguageC                      = "c"
	LanguagePOSIX                  = "posix"

	EnvLanguageOverride = "FWB_LANG"
	EnvLocaleAll        = "LC_ALL"
	EnvLocaleMessages   = "LC_MESSAGES"
	EnvLang             = "LANG"
)

// Catalog keys.
const (
	KeyHelpText             = "help.text"
	KeyUnknownCommand       = "error.unknownCommand"
	KeyUnknownConfigCommand = "error.unknownConfigCommand"

	KeyUsageConfig      = "usage.config"
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

	KeyProgressDownloading        = "progress.downloading"
	KeyProgressDownloadingUnknown = "progress.downloadingUnknown"
	KeyProgressDone               = "progress.done"
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
const ProgressThrottleInterval = 2 * time.Second

// Process exit codes and fatal error output.
const (
	ExitSuccess       = 0
	ExitFailure       = 1
	ErrorOutputFormat = "fwb: %v\n"
)
