package cli

import (
	"testing"

	"golang.org/x/mod/semver"
)

func TestModuleVersion(t *testing.T) {
	if got := moduleVersion(); !semver.IsValid(got) {
		t.Fatalf("version is not semver: %q", got)
	}
}
