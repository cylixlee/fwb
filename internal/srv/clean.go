package srv

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// Cleaner deletes the Windows-side state of the current project.
type Cleaner struct {
	mirrors domain.ProjectMirrorFactory
}

// NewCleaner wires the clean use case.
func NewCleaner(mirrors domain.ProjectMirrorFactory) *Cleaner {
	return &Cleaner{mirrors: mirrors}
}

// Clean removes the mirror of the current project, and the SDK cache when
// withSDK is set. It returns the paths it removed.
func (c *Cleaner) Clean(ctx context.Context, withSDK bool) ([]string, error) {
	m, err := c.mirrors.Current(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.Remove(); err != nil {
		return nil, err
	}
	removed := []string{m.Path()}
	if !withSDK {
		return removed, nil
	}
	root, err := c.mirrors.Root(ctx)
	if err != nil {
		return removed, err
	}
	dir := filepath.Join(root, consts.SDKDirectoryName)
	if err := os.RemoveAll(dir); err != nil {
		return removed, err
	}
	return append(removed, dir), nil
}
