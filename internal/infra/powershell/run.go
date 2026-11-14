package powershell

import (
	"context"
	"io"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

type runParams struct {
	Dir                  string
	Exe                  string
	SessionFile          string
	Switches             []string
	SwitchesVariable     string
	SwitchVariablePrefix string
	SessionEncoding      string
}

func (h host) Run(ctx context.Context, req domain.WindowsRunRequest, stdout, stderr io.Writer) error {
	dirWin, err := toWindowsPath(ctx, req.Dir)
	if err != nil {
		return err
	}
	exeWin, err := toWindowsPath(ctx, req.Exe)
	if err != nil {
		return err
	}
	sessionWin, err := toWindowsPath(ctx, req.SessionFile)
	if err != nil {
		return err
	}
	data := runParams{
		Dir:                  dirWin,
		Exe:                  exeWin,
		SessionFile:          sessionWin,
		Switches:             req.Switches,
		SwitchesVariable:     consts.EngineSwitchesVariable,
		SwitchVariablePrefix: consts.EngineSwitchVariablePrefix,
		SessionEncoding:      consts.SessionFileEncoding,
	}
	return runPowerShellStream(ctx, consts.ScriptRun, data, stdout, stderr)
}

type killParams struct {
	PID                      int
	CloseTimeoutMilliseconds int
}

func (host) Stop(ctx context.Context, pid int) error {
	script, err := renderScript(consts.ScriptKill, killParams{PID: pid, CloseTimeoutMilliseconds: consts.ProcessCloseTimeoutMillis})
	if err != nil {
		return err
	}
	_, err = runPowerShell(ctx, script)
	return err
}
