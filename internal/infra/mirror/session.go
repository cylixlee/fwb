package mirror

import (
	"os"
	"strconv"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
)

func (m *projectMirror) SessionPID() (int, bool) {
	data, err := os.ReadFile(m.SessionFile())
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

func (m *projectMirror) WriteSession(pid int) error {
	if err := os.MkdirAll(m.metaDir(), consts.DirectoryPermissions); err != nil {
		return err
	}
	return os.WriteFile(m.SessionFile(), []byte(strconv.Itoa(pid)), consts.FilePermissions)
}

func (m *projectMirror) ClearSession() { os.Remove(m.SessionFile()) }
