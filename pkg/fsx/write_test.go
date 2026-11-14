package fsx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteJSONAtomicCreatesParentsAndFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	if err := WriteJSONAtomic(path, map[string]string{"key": "value"}, 0o644); err != nil {
		t.Fatalf("WriteJSONAtomic: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatalf("expected trailing newline, got %q", data)
	}
}
