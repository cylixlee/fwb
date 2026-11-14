package srv

import (
	"context"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// stopSession terminates the app whose pid the runner recorded in m, then
// clears the session file. Failures are ignored: the process may be gone.
func stopSession(ctx context.Context, host domain.WindowsHost, m domain.ProjectMirror) {
	pid, ok := m.SessionPID()
	if !ok {
		return
	}
	defer m.ClearSession()
	// The launch context is already canceled when Flutter quits the hook.
	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), consts.StopTimeout)
	defer cancel()
	host.Stop(stopCtx, pid)
}
