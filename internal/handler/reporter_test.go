package handler

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cylixlee/fwb/internal/consts"
)

func TestLineProgressPrintsThrottledLines(t *testing.T) {
	SetLanguage("en")
	var out bytes.Buffer
	progress := newReporter(&out, false).StartProgress(consts.ProgressSDKDownload, "3.50.0")

	progress.Update(0, 2_000_000_000)
	progress.Update(100_000_000, 2_000_000_000)
	progress.Done()

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("printed %d lines, want 2: %q", len(lines), out.String())
	}
	if strings.Contains(out.String(), "\r") {
		t.Errorf("non-terminal progress must not emit carriage returns: %q", out.String())
	}
	if !strings.HasPrefix(lines[0], consts.ProgressDescriptionPrefix+"Flutter SDK 3.50.0") {
		t.Errorf("unexpected first line: %q", lines[0])
	}
	if !strings.Contains(lines[0], "0%") || !strings.Contains(lines[0], "0 B/2.0 GB") {
		t.Errorf("unexpected first line: %q", lines[0])
	}
	if !strings.Contains(lines[1], "5%") || !strings.Contains(lines[1], "100 MB") {
		t.Errorf("unexpected final line: %q", lines[1])
	}
}

func TestLineProgressWaitsForTotal(t *testing.T) {
	SetLanguage("en")
	var out bytes.Buffer
	progress := newReporter(&out, false).StartProgress(consts.ProgressSDKDownload, "3.50.0")

	progress.Update(0, 0)
	progress.Done()

	if out.Len() != 0 {
		t.Fatalf("printed %q without a known total", out.String())
	}
}

func TestTerminalProgressStaysOnOneLine(t *testing.T) {
	SetLanguage("en")
	var out bytes.Buffer
	progress := newReporter(&out, true).StartProgress(consts.ProgressSDKDownload, "3.50.0")

	progress.Update(0, 2_000_000_000)
	progress.Done()

	if !strings.Contains(out.String(), "\r") {
		t.Errorf("terminal progress should animate with carriage returns: %q", out.String())
	}
	if newlines := strings.Count(out.String(), "\n"); newlines != 1 {
		t.Errorf("terminal progress printed %d newlines, want 1: %q", newlines, out.String())
	}
}

func TestAbortNeverClaimsCompletion(t *testing.T) {
	SetLanguage("en")
	var out bytes.Buffer
	progress := newReporter(&out, false).StartProgress(consts.ProgressSDKDownload, "3.50.0")

	progress.Update(100_000_000, 2_000_000_000)
	progress.Abort()

	if strings.Contains(out.String(), "100%") {
		t.Errorf("aborted progress claimed completion: %q", out.String())
	}
}
