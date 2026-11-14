package fsx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindExecutablePrefersBaseName(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"other.exe", "app.exe"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := FindExecutable(dir, "app", ".exe"); got != filepath.Join(dir, "app.exe") {
		t.Fatalf("FindExecutable = %q", got)
	}
	if got := FindExecutable(dir, "", ".exe"); got == "" {
		t.Fatal("expected any executable")
	}
}
