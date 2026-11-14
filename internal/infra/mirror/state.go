package mirror

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/pkg/fsx"
)

func (m *projectMirror) LoadState() (domain.BuildState, error) {
	var st domain.BuildState
	data, err := os.ReadFile(m.stateFile())
	if errors.Is(err, os.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	var raw struct {
		Fingerprint    string    `json:"fingerprint"`
		EngineRevision string    `json:"engineRevision"`
		BuildDir       string    `json:"buildDir"`
		Exe            string    `json:"exe"`
		BuiltAt        time.Time `json:"builtAt"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return st, err
	}
	return domain.BuildState{
		Fingerprint:    raw.Fingerprint,
		EngineRevision: raw.EngineRevision,
		BuildDir:       raw.BuildDir,
		Exe:            raw.Exe,
		BuiltAt:        raw.BuiltAt,
	}, nil
}

func (m *projectMirror) SaveState(st domain.BuildState) error {
	if err := os.MkdirAll(m.metaDir(), consts.DirectoryPermissions); err != nil {
		return err
	}
	payload := map[string]any{
		consts.JSONFingerprint:    st.Fingerprint,
		consts.JSONEngineRevision: st.EngineRevision,
		consts.JSONBuildDir:       st.BuildDir,
		consts.JSONExe:            st.Exe,
		consts.JSONBuiltAt:        st.BuiltAt,
	}
	return fsx.WriteJSONAtomic(m.stateFile(), payload, consts.FilePermissions)
}

func (m *projectMirror) Remove() error { return os.RemoveAll(m.root) }
