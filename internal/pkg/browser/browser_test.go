package browser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveFailsWithActionableMessage(t *testing.T) {
	t.Setenv(execEnv, "")
	// Force the "nothing installed" branch regardless of what the machine running the test
	// happens to have, so the assertion is about the message rather than the environment.
	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	_, err := Resolve()
	if err == nil {
		t.Fatal("Resolve should fail when no browser is installed")
	}
	msg := err.Error()
	// An operator reading only the API response must learn which dependency is missing and
	// how to supply it, not just that something went wrong internally.
	for _, want := range []string{"Chrome", execEnv, "chromium"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message %q should mention %q", msg, want)
		}
	}
}

func TestResolveHonoursEnvOverride(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "my-browser")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(execEnv, fake)
	t.Setenv("PATH", t.TempDir())

	got, err := Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != fake {
		t.Errorf("Resolve() = %q, want %q", got, fake)
	}
}

// A stale override is a configuration mistake worth naming, rather than falling through to
// the PATH search and hiding it.
func TestResolveRejectsMissingEnvOverride(t *testing.T) {
	t.Setenv(execEnv, filepath.Join(t.TempDir(), "not-here"))
	if _, err := Resolve(); err == nil {
		t.Fatal("a nonexistent override should be reported, not ignored")
	}
}

func TestResolveFindsFlatpakUserInstall(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, ".local", "share", "flatpak", "exports", "bin")
	if err := os.MkdirAll(prefix, 0o755); err != nil {
		t.Fatal(err)
	}
	flatpakChrome := filepath.Join(prefix, "com.google.Chrome")
	if err := os.WriteFile(flatpakChrome, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(execEnv, "")
	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", home)

	got, err := Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != flatpakChrome {
		t.Errorf("Resolve() = %q, want the Flatpak path %q", got, flatpakChrome)
	}
}
