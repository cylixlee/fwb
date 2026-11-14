package srv

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// DeviceHooks serves the ping, install, uninstall, install-device and
// uninstall-device hooks Flutter calls.
type DeviceHooks struct {
	cfg     domain.ConfigStore
	host    domain.WindowsHost
	flutter domain.WSLFlutterCLI
	mirrors domain.ProjectMirrorFactory
	devices domain.DeviceRepository
}

// NewDeviceHooks wires the device hooks.
func NewDeviceHooks(
	cfg domain.ConfigStore,
	host domain.WindowsHost,
	flutter domain.WSLFlutterCLI,
	mirrors domain.ProjectMirrorFactory,
	devices domain.DeviceRepository,
) *DeviceHooks {
	return &DeviceHooks{cfg: cfg, host: host, flutter: flutter, mirrors: mirrors, devices: devices}
}

// Ping is the custom-device ping hook; it must not spawn Windows processes.
func (h *DeviceHooks) Ping(ctx context.Context) error {
	if err := h.host.Available(); err != nil {
		return err
	}
	return h.flutter.Available()
}

// Install snapshots the WSL asset bundle. Flutter does not stream this hook's
// output, so it must stay fast.
func (h *DeviceHooks) Install(ctx context.Context, localPath string) error {
	m, err := h.mirrors.Current(ctx)
	if err != nil {
		return err
	}
	return m.Stage(localPath)
}

// Uninstall stops a previous app instance; failures are not fatal because
// Flutter calls this before every install.
func (h *DeviceHooks) Uninstall(ctx context.Context) error {
	m, err := h.mirrors.Current(ctx)
	if err != nil {
		return nil
	}
	stopSession(ctx, h.host, m)
	return nil
}

// Register stores the custom-device entry. The label is supplied by the
// interface layer so it can be localized. The boolean reports whether the
// flutter config flag could be enabled automatically.
func (h *DeviceHooks) Register(ctx context.Context, label, sdkName string) (string, bool, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", false, err
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	path, err := h.devices.Install(domain.DeviceSpec{
		ID:       h.cfg.Get(consts.ConfigKeyDeviceID),
		Label:    label,
		SDKName:  sdkName,
		Platform: consts.DevicePlatformLinuxX64,
		Exe:      exe,
	})
	if err != nil {
		return "", false, err
	}
	enabled := h.flutter.ConfigureCustomDevices(ctx) == nil
	return path, enabled, nil
}

// Unregister removes the custom-device entry.
func (h *DeviceHooks) Unregister(ctx context.Context) (string, error) {
	return h.devices.Remove(h.cfg.Get(consts.ConfigKeyDeviceID))
}
