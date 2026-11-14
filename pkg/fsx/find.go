// Package fsx provides generic filesystem helpers.
package fsx

import (
	"os"
	"path/filepath"
	"strings"
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
