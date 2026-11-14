package fsx

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/renameio/v2"
)

const (
	directoryPermissions = 0o755
	jsonIndent           = "  "
)

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
