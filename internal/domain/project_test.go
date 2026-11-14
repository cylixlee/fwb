package domain

import (
	"strings"
	"testing"
)

func TestProjectIDIsStableAndSanitized(t *testing.T) {
	a := NewProject("/home/u/my app")
	b := NewProject("/home/u/my app")
	if a.ID() != b.ID() {
		t.Fatalf("project id is not stable: %q != %q", a.ID(), b.ID())
	}
	if strings.ContainsAny(a.ID(), " /") {
		t.Fatalf("project id contains unsafe characters: %q", a.ID())
	}
	if c := NewProject("/home/u/my-app"); c.ID() == a.ID() {
		t.Fatalf("different projects share an id: %q", a.ID())
	}
}
