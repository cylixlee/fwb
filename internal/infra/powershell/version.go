package powershell

import (
	"context"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/internal/infra/flutter"
)

type versionParams struct {
	Bat         string
	VersionFlag string
	MachineFlag string
}

func (host) FlutterVersion(ctx context.Context, flutterBatWindows string) (domain.FlutterVersion, error) {
	out, err := runPowerShellName(ctx, consts.ScriptFlutterVersion, versionParams{
		Bat:         flutterBatWindows,
		VersionFlag: consts.FlutterVersionFlag,
		MachineFlag: consts.FlutterMachineFlag,
	})
	if err != nil {
		return domain.FlutterVersion{}, err
	}
	return flutter.Parse(out)
}
