package domain

import (
	"reflect"
	"testing"
)

func TestParseEngineSwitches(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single", "--enable-dart-profiling=true", []string{"enable-dart-profiling=true"}},
		{
			"multiple",
			"--route=/a --enable-checked-mode=true --start-paused",
			[]string{"route=/a", "enable-checked-mode=true", "start-paused"},
		},
		{"extra spaces", "  --a=1   --b=2  ", []string{"a=1", "b=2"}},
		{"quoted value", `--dart-flags="--foo --bar" --c=3`, []string{"dart-flags=--foo --bar", "c=3"}},
		{"single quotes", `--dart-flags='--foo --bar'`, []string{"dart-flags=--foo --bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseEngineSwitches(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseEngineSwitches(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}
