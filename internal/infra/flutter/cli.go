// Package flutter provides adapters for Flutter SDK discovery.
package flutter

import (
	"context"
	"os/exec"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

type cli struct{}

// NewCLI returns the Flutter SDK on WSL's PATH.
func NewCLI() domain.WSLFlutterCLI { return cli{} }

func (cli) Available() error {
	if _, err := exec.LookPath(consts.FlutterExecutable); err != nil {
		return domain.WrapError(consts.ErrorCodeFlutterRun, err)
	}
	return nil
}

func (cli) Version(ctx context.Context) (domain.FlutterVersion, error) {
	out, err := exec.CommandContext(ctx, consts.FlutterExecutable, consts.FlutterVersionFlag, consts.FlutterMachineFlag).Output()
	if err != nil {
		return domain.FlutterVersion{}, domain.WrapError(consts.ErrorCodeFlutterRun, err)
	}
	return Parse(string(out))
}

func (cli) ConfigureCustomDevices(ctx context.Context) error {
	return exec.CommandContext(ctx, consts.FlutterExecutable, consts.FlutterConfigCommand, consts.FlutterEnableCustomDevicesFlag).Run()
}
