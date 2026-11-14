package domain

import "github.com/cylixlee/fwb/internal/consts"

// DeviceSpec describes a registered Flutter custom device.
type DeviceSpec struct {
	ID       string
	Label    string
	SDKName  string
	Platform string
	Exe      string
}

// Hooks returns the command lists for each custom-device hook.
func (d DeviceSpec) Hooks() map[string][]string {
	return map[string][]string{
		consts.HookPing:      {d.Exe, consts.HookPing},
		consts.HookUninstall: {d.Exe, consts.HookUninstall, consts.PlaceholderAppName},
		consts.HookInstall:   {d.Exe, consts.HookInstall, consts.PlaceholderLocalPath, consts.PlaceholderAppName},
		consts.HookRunDebug:  {d.Exe, consts.CommandLaunch, consts.PlaceholderAppName, consts.PlaceholderEngineOptions},
	}
}
