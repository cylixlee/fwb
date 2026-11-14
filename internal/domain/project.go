package domain

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"

	"github.com/gosimple/slug"

	"github.com/cylixlee/fwb/internal/consts"
)

// Project is a Flutter project opened from a WSL working directory.
type Project struct {
	Root string
	Name string
}

// NewProject derives a project from its root directory.
func NewProject(root string) Project {
	return Project{Root: filepath.Clean(root), Name: filepath.Base(root)}
}

// ID is the stable mirror identity: "<name>-<sha1(root)[:MirrorIDHashLength]>".
func (p Project) ID() string {
	name := slug.Make(p.Name)
	if name == "" {
		name = consts.MirrorIDFallbackSlug
	}
	sum := sha1.Sum([]byte(p.Root))
	return name + consts.MirrorIDSeparator + hex.EncodeToString(sum[:])[:consts.MirrorIDHashLength]
}
