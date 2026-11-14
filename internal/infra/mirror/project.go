package mirror

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	cp "github.com/otiai10/copy"
)

// projectMirror is the Windows-side copy of one Flutter project.
type projectMirror struct {
	project domain.Project
	root    string
	cfg     domain.ConfigStore
}

func (m *projectMirror) Path() string { return m.root }

func (m *projectMirror) metaDir() string { return filepath.Join(m.root, consts.MetaDirectoryName) }
func (m *projectMirror) stagingDir() string {
	return filepath.Join(m.metaDir(), consts.StagingDirectoryName)
}
func (m *projectMirror) SessionFile() string {
	return filepath.Join(m.metaDir(), consts.SessionFileName)
}
func (m *projectMirror) stateFile() string {
	return filepath.Join(m.metaDir(), consts.StateFileName)
}

func (m *projectMirror) Stage(localPath string) error {
	src, err := filepath.Abs(localPath)
	if err != nil {
		return err
	}
	if st, err := os.Stat(src); err != nil || !st.IsDir() {
		return domain.NewError(consts.ErrorCodeProjectNoBundle, src)
	}
	dst := filepath.Join(m.stagingDir(), consts.FlutterAssetsDir)
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return cp.Copy(src, dst, copyOptions)
}

func (m *projectMirror) Sync() error {
	if err := os.MkdirAll(m.root, consts.DirectoryPermissions); err != nil {
		return err
	}
	for _, name := range []string{consts.PubspecFileName, consts.PubspecLockFileName, consts.MetadataFileName, consts.AnalysisOptionsFileName} {
		src := filepath.Join(m.project.Root, name)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := cp.Copy(src, filepath.Join(m.root, name), copyOptions); err != nil {
			return err
		}
	}
	for _, rel := range m.syncRoots() {
		src := filepath.Join(m.project.Root, rel)
		if st, err := os.Stat(src); err != nil || !st.IsDir() {
			continue
		}
		dst := filepath.Join(m.root, rel)
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
		if err := cp.Copy(src, dst, copyOptions); err != nil {
			return domain.WrapError(consts.ErrorCodeProjectSync, err, rel)
		}
	}
	return nil
}

func (m *projectMirror) Merge(buildDir string) error {
	src := filepath.Join(m.stagingDir(), consts.FlutterAssetsDir)
	if st, err := os.Stat(src); err != nil || !st.IsDir() {
		return domain.NewError(consts.ErrorCodeProjectNoStaged)
	}
	dst := filepath.Join(buildDir, consts.DataDirectory, consts.FlutterAssetsDir)
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return cp.Copy(src, dst, copyOptions)
}

func (m *projectMirror) syncRoots() []string {
	roots := []string{
		consts.LibDirectoryName,
		consts.WindowsDirectory,
		consts.AssetsDirectoryName,
		consts.FontsDirectoryName,
		consts.ShadersDirectoryName,
	}
	for extra := range strings.SplitSeq(m.cfg.Get(consts.ConfigKeySyncExtra), consts.PathListSeparator) {
		if extra = strings.TrimSpace(extra); extra != "" {
			roots = append(roots, extra)
		}
	}
	return roots
}
