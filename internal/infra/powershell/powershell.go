// Package powershell runs embedded PowerShell scripts through WSL interop. It
// is the only package that talks to Windows.
package powershell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

var (
	psOnce sync.Once
	psPath string
	psErr  error
)

func lookPowerShell() (string, error) {
	psOnce.Do(func() {
		var candidates []string
		if p, err := exec.LookPath(consts.PowerShellExecutable); err == nil {
			candidates = append(candidates, p)
		}
		if p, err := exec.LookPath(consts.PwshExecutable); err == nil {
			candidates = append(candidates, p)
		}
		candidates = append(candidates, consts.FallbackPowerShellPath, consts.FallbackPwshPath)
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				psPath = c
				return
			}
		}
		psErr = domain.NewError(consts.ErrorCodeWinNoPowerShell)
	})
	return psPath, psErr
}

func psCommand(ctx context.Context, script string) (*exec.Cmd, error) {
	ps, err := lookPowerShell()
	if err != nil {
		return nil, err
	}
	return exec.CommandContext(ctx, ps,
		consts.PowerShellFlagNoProfile,
		consts.PowerShellFlagNonInteractive,
		consts.PowerShellFlagEncodedCommand,
		encodeCommand(script),
	), nil
}

func runPowerShell(ctx context.Context, script string) (string, error) {
	cmd, err := psCommand(ctx, script)
	if err != nil {
		return "", err
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg != "" {
			err = fmt.Errorf(consts.ErrorCauseFormat, msg, err)
		}
		return stdout.String(), domain.WrapError(consts.ErrorCodeWinPowerShell, err)
	}
	return stdout.String(), nil
}

func runPowerShellName(ctx context.Context, name string, data any) (string, error) {
	script, err := renderScript(name, data)
	if err != nil {
		return "", err
	}
	return runPowerShell(ctx, script)
}

func runPowerShellStream(ctx context.Context, name string, data any, stdout, stderr io.Writer) error {
	script, err := renderScript(name, data)
	if err != nil {
		return err
	}
	cmd, err := psCommand(ctx, script)
	if err != nil {
		return err
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return domain.WrapError(consts.ErrorCodeWinPowerShell, err)
	}
	return nil
}
