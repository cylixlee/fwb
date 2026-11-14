// Package mirror keeps a Windows-side copy of the WSL project.
package mirror

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/internal/infra/powershell"
	cp "github.com/otiai10/copy"
)

// copyOptions resolves symlinks so the Windows mirror contains real files.
var copyOptions = cp.Options{
	OnSymlink: func(string) cp.SymlinkAction { return cp.Deep },
}

// factory resolves projects and their Windows-side mirrors.
type factory struct {
	cfg  domain.ConfigStore
	host domain.WindowsHost
}

// NewFactory returns the project mirror factory.
func NewFactory(cfg domain.ConfigStore, host domain.WindowsHost) domain.ProjectMirrorFactory {
	return &factory{cfg: cfg, host: host}
}

func (f *factory) Root(ctx context.Context) (string, error) {
	return Root(ctx, f.cfg, f.host)
}

func (f *factory) Open(ctx context.Context) (domain.Project, error) {
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

func (f *factory) For(ctx context.Context, p domain.Project) (domain.ProjectMirror, error) {
	root, err := Root(ctx, f.cfg, f.host)
	if err != nil {
		return nil, err
	}
	return &projectMirror{
		project: p,
		root:    filepath.Join(root, consts.ProjectsDirectoryName, p.ID()),
		cfg:     f.cfg,
	}, nil
}

func (f *factory) Current(ctx context.Context) (domain.ProjectMirror, error) {
	p, err := f.Open(ctx)
	if err != nil {
		return nil, err
	}
	return f.For(ctx, p)
}

// Root is %LOCALAPPDATA%\fwb unless windows.mirrorRoot is set.
func Root(ctx context.Context, cfg domain.ConfigStore, host domain.WindowsHost) (string, error) {
	if v := strings.TrimSpace(cfg.Get(consts.ConfigKeyWindowsMirrorRoot)); v != "" {
		if strings.HasPrefix(v, consts.WSLPathPrefix) {
			return v, nil
		}
		return host.ToWSL(ctx, v)
	}
	env, err := powershell.Environment(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(env.LocalAppDataWSL, consts.AppDirectoryName), nil
}
