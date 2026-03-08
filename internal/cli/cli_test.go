package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunSync(t *testing.T) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	err := Run([]string{"sync"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "sync stub: not wired yet") {
		t.Fatalf("stdout = %q, want sync stub message", got)
	}
}

func TestRunList(t *testing.T) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	err := Run([]string{"list"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "SOURCE   COMPANY      TITLE") {
		t.Fatalf("stdout = %q, want table header", got)
	}

	if !strings.Contains(got, "Backend Engineer") {
		t.Fatalf("stdout = %q, want sample row", got)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	err := Run([]string{"unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error, want parse error")
	}
}
