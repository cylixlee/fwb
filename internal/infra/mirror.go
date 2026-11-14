package infra

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/pkg/fsx"
	cp "github.com/otiai10/copy"
)

// mirrorCopyOptions resolves symlinks so the Windows mirror contains real files.
var mirrorCopyOptions = cp.Options{
	OnSymlink: func(string) cp.SymlinkAction { return cp.Deep },
}

// mirrorFactory resolves projects and their Windows-side mirrors.
type mirrorFactory struct {
	cfg  domain.ConfigStore
	host domain.WindowsHost
}

// NewProjectMirrorFactory returns the project mirror factory.
func NewProjectMirrorFactory(cfg domain.ConfigStore, host domain.WindowsHost) domain.ProjectMirrorFactory {
	return &mirrorFactory{cfg: cfg, host: host}
}

func (f *mirrorFactory) Root(ctx context.Context) (string, error) {
	return mirrorRoot(ctx, f.cfg, f.host)
}

func (f *mirrorFactory) Open(ctx context.Context) (domain.Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return domain.Project{}, err
	}
	if real, err := filepath.EvalSymlinks(cwd); err == nil {
		cwd = real
	}
	if _, err := os.Stat(filepath.Join(cwd, consts.PubspecFileName)); err != nil {
		return domain.Project{}, domain.NewError(consts.ErrorCodeProjectNotFlutter, cwd)
	}
	if st, err := os.Stat(filepath.Join(cwd, consts.WindowsDirectory)); err != nil || !st.IsDir() {
		return domain.Project{}, domain.NewError(consts.ErrorCodeProjectNoWindows)
	}
	return domain.NewProject(cwd), nil
}

func (f *mirrorFactory) For(ctx context.Context, p domain.Project) (domain.ProjectMirror, error) {
	root, err := mirrorRoot(ctx, f.cfg, f.host)
	if err != nil {
		return nil, err
	}
	return &projectMirror{
		project: p,
		root:    filepath.Join(root, consts.ProjectsDirectoryName, p.ID()),
		cfg:     f.cfg,
	}, nil
}

// mirrorRoot is %LOCALAPPDATA%\fwb unless windows.mirrorRoot is set.
func mirrorRoot(ctx context.Context, cfg domain.ConfigStore, host domain.WindowsHost) (string, error) {
	if v := strings.TrimSpace(cfg.Get(consts.ConfigKeyWindowsMirrorRoot)); v != "" {
		if strings.HasPrefix(v, consts.WSLPathPrefix) {
			return v, nil
		}
		return host.ToWSL(ctx, v)
	}
	env, err := cachedEnv(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(env.LocalAppDataWSL, consts.AppDirectoryName), nil
}

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
	return cp.Copy(src, dst, mirrorCopyOptions)
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
		if err := cp.Copy(src, filepath.Join(m.root, name), mirrorCopyOptions); err != nil {
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
		if err := cp.Copy(src, dst, mirrorCopyOptions); err != nil {
			return domain.WrapError(consts.ErrorCodeProjectSync, err, rel)
		}
	}
	return nil
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

func (m *projectMirror) Fingerprint(engineRevision string) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, consts.FingerprintEngineFormat, engineRevision)
	for _, name := range []string{consts.PubspecFileName, consts.PubspecLockFileName, consts.MetadataFileName} {
		data, err := os.ReadFile(filepath.Join(m.project.Root, name))
		if err == nil {
			fmt.Fprintf(h, consts.FingerprintFileFormat, name, sha256.Sum256(data))
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	for _, rel := range m.syncRoots() {
		dir := filepath.Join(m.project.Root, rel)
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			continue
		}
		var lines []string
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			relPath, _ := filepath.Rel(m.project.Root, path)
			lines = append(lines, fmt.Sprintf(consts.FingerprintLineFormat, relPath, info.Size(), info.ModTime().UnixNano()))
			return nil
		})
		if err != nil {
			return "", err
		}
		sort.Strings(lines)
		for _, l := range lines {
			fmt.Fprintln(h, l)
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (m *projectMirror) Locate(mode string) (string, string, error) {
	pattern := filepath.Join(m.root, consts.BuildDirectory, consts.WindowsDirectory, consts.GlobWildcard, consts.RunnerDirectory, mode)
	matches, _ := filepath.Glob(pattern)
	sort.Strings(matches)
	for _, match := range slices.Backward(matches) {
		if exe := fsx.FindExecutable(match, m.binaryName(), consts.ExeSuffix); exe != "" {
			return match, exe, nil
		}
	}
	return "", "", domain.NewError(consts.ErrorCodeProjectNoBuild, pattern)
}

func (m *projectMirror) binaryName() string {
	data, err := os.ReadFile(filepath.Join(m.root, consts.WindowsDirectory, consts.CMakeListsFileName))
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), consts.LineFeed) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(line), consts.BinaryNameDirective) {
			continue
		}
		start := strings.Index(line, consts.DoubleQuote)
		end := strings.LastIndex(line, consts.DoubleQuote)
		if start >= 0 && end > start {
			return line[start+1 : end]
		}
	}
	return ""
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
	return cp.Copy(src, dst, mirrorCopyOptions)
}

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
