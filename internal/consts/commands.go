package consts

// CLI commands. Shared by the handler and the custom-device hook spec.
const (
	CommandVersion         = "version"
	CommandHelp            = "help"
	CommandPing            = "ping"
	CommandInstall         = "install"
	CommandUninstall       = "uninstall"
	CommandLaunch          = "launch"
	CommandConfig          = "config"
	CommandInstallDevice   = "install-device"
	CommandUninstallDevice = "uninstall-device"
	CommandDoctor          = "doctor"
	CommandClean           = "clean"

	ConfigCommandSet   = "set"
	ConfigCommandGet   = "get"
	ConfigCommandUnset = "unset"
	ConfigCommandList  = "list"
)

// Environment variable names shared across layers.
const (
	EnvPubCache              = "PUB_CACHE"
	EnvPubHostedURL          = "PUB_HOSTED_URL"
	EnvFlutterStorageBaseURL = "FLUTTER_STORAGE_BASE_URL"
)
