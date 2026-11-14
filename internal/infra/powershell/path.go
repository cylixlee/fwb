package powershell

import (
	"context"
	"os/exec"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

func toWindowsPath(ctx context.Context, p string) (string, error) {
	out, err := exec.CommandContext(ctx, consts.WslPathExecutable, consts.WslPathWindowsFlag, p).Output()
	if err != nil {
		return "", domain.WrapError(consts.ErrorCodeWinPath, err, consts.WslPathWindowsFlag, p)
	}
	return strings.TrimSpace(string(out)), nil
}

func toWSLPath(ctx context.Context, p string) (string, error) {
	out, err := exec.CommandContext(ctx, consts.WslPathExecutable, consts.WslPathWSLFlag, p).Output()
	if err != nil {
		return "", domain.WrapError(consts.ErrorCodeWinPath, err, consts.WslPathWSLFlag, p)
	}
	return strings.TrimSpace(string(out)), nil
}
