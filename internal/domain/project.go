package domain

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"strings"
	"unicode"

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
	sum := sha1.Sum([]byte(p.Root))
	return sanitize(p.Name) + consts.MirrorIDSeparator + hex.EncodeToString(sum[:])[:consts.MirrorIDHashLength]
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r),
			strings.ContainsRune(consts.MirrorIDAllowedPunctuation, r):
			b.WriteRune(r)
		default:
			b.WriteString(consts.MirrorIDReplacement)
		}
	}
	if b.Len() == 0 {
		return consts.MirrorIDFallbackSlug
	}
	return b.String()
}
