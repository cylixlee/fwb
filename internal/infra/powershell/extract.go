package powershell

import (
	"context"

	"github.com/cylixlee/fwb/internal/consts"
)

type extractParams struct {
	Zip             string
	Dest            string
	Tar             string
	ExpandArchive   string
	TarFailedFormat string
}

func (host) Extract(ctx context.Context, zipWSL, destWSL string) error {
	zipWin, err := toWindowsPath(ctx, zipWSL)
	if err != nil {
		return err
	}
	destWin, err := toWindowsPath(ctx, destWSL)
	if err != nil {
		return err
	}
	_, err = runPowerShellName(ctx, consts.ScriptExtract, extractParams{
		Zip:             zipWin,
		Dest:            destWin,
		Tar:             consts.TarExecutable,
		ExpandArchive:   consts.ExpandArchive,
		TarFailedFormat: consts.TarFailedFormat,
	})
	return err
}
