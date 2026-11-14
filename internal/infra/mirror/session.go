package mirror

import (
	"os"
	"strconv"
	"strings"
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

func (m *projectMirror) ClearSession() { os.Remove(m.SessionFile()) }
