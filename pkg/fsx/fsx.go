// Package fsx provides generic filesystem helpers.
package fsx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/renameio/v2"
)

const (
	directoryPermissions = 0o755
	jsonIndent           = "  "
)

// FindExecutable returns the executable in dir, preferring base+ext. An empty
// string means nothing was found.
func FindExecutable(dir, base, ext string) string {
	if base != "" {
		if path := filepath.Join(dir, base+ext); fileExists(path) {
			return path
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ext) {
			return filepath.Join(dir, entry.Name())
		}
	}
	return ""
}

// WriteJSONAtomic writes v as indented JSON to path atomically.
func WriteJSONAtomic(path string, v any, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), directoryPermissions); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", jsonIndent)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return renameio.WriteFile(path, data, perm)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
