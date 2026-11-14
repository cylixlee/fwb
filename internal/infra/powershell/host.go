package powershell

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// host implements domain.WindowsHost.
type host struct{}

// NewHost returns the Windows interop gateway.
func NewHost() domain.WindowsHost { return host{} }

func (host) Available() error {
	if _, err := lookPowerShell(); err != nil {
		return err
	}
	if _, err := exec.LookPath(consts.WslPathExecutable); err != nil {
		return domain.NewError(consts.ErrorCodeWinNoWslpath)
	}
	return nil
}

func (host) Info(ctx context.Context) (string, error) {
	out, err := runPowerShellName(ctx, consts.ScriptPowerShellInfo, nil)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(strings.TrimSpace(out), consts.CarriageReturnLineFeed, consts.InfoSeparator), nil
}

type envParams struct {
	LocalAppData string
	UserProfile  string
}

func (host) Env(ctx context.Context) (domain.WindowsEnvironment, error) {
	out, err := runPowerShellName(ctx, consts.ScriptEnvironment, envParams{
		LocalAppData: consts.EnvLocalAppData,
		UserProfile:  consts.EnvUserProfile,
	})
	if err != nil {
		return domain.WindowsEnvironment{}, err
	}
	lines := strings.Split(strings.ReplaceAll(out, consts.CarriageReturnLineFeed, consts.LineFeed), consts.LineFeed)
	if len(lines) < 2 {
		return domain.WindowsEnvironment{}, domain.NewError(consts.ErrorCodeWinEnv, out)
	}
	env := domain.WindowsEnvironment{
		LocalAppDataWin: strings.TrimSpace(lines[0]),
		OSVersion:       strings.TrimSpace(strings.Join(lines[2:], consts.InfoSeparator)),
	}
	userProfile := strings.TrimSpace(lines[1])
	if env.LocalAppDataWin == "" || userProfile == "" {
		return domain.WindowsEnvironment{}, domain.NewError(consts.ErrorCodeWinEnv, out)
	}
	if env.LocalAppDataWSL, err = toWSLPath(ctx, env.LocalAppDataWin); err != nil {
		return domain.WindowsEnvironment{}, err
	}
	if env.UserProfileWSL, err = toWSLPath(ctx, userProfile); err != nil {
		return domain.WindowsEnvironment{}, err
	}
	return env, nil
}

func (host) Mirrored(ctx context.Context) (bool, string, error) {
	env, err := Environment(ctx)
	if err != nil {
		return false, "", err
	}
	path := filepath.Join(env.UserProfileWSL, consts.WSLConfigFileName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, path, nil
	}
	if err != nil {
		return false, path, err
	}
	for line := range strings.SplitSeq(strings.ToLower(string(data)), consts.LineFeed) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, consts.WSLNetworkingModeKey) && strings.Contains(line, consts.WSLMirrorModeValue) {
			return true, path, nil
		}
	}
	return false, path, nil
}

func (host) ToWindows(ctx context.Context, path string) (string, error) {
	return toWindowsPath(ctx, path)
}

func (host) ToWSL(ctx context.Context, path string) (string, error) {
	return toWSLPath(ctx, path)
}
