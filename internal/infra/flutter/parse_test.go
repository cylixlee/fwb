package flutter

import "testing"

func TestParse(t *testing.T) {
	out := "Building flutter tool...\n" +
		`{"frameworkVersion":"3.35.4","channel":"stable","frameworkRevision":"abc","engineRevision":"def","dartSdkVersion":"3.9.2"}`
	v, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if v.Framework != "3.35.4" || v.Channel != "stable" || v.EngineRevision != "def" {
		t.Fatalf("unexpected version: %+v", v)
	}
}

func TestParseRejectsIncompleteOutput(t *testing.T) {
	if _, err := Parse(`{"frameworkVersion":"3.35.4"}`); err == nil {
		t.Fatal("expected an error for incomplete output")
	}
	if _, err := Parse("not json"); err == nil {
		t.Fatal("expected an error for non-JSON output")
	}
}
