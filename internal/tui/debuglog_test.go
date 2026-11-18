package tui

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOpenPrivateLog_CreatesAndTightens(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows doesn't have Unix permission bits")
	}
	dir := t.TempDir()
	fresh := filepath.Join(dir, "tui.log")
	old := filepath.Join(dir, "old.log")
	if err := os.WriteFile(old, []byte("earlier\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(old, 0o644); err != nil { // past the umask
		t.Fatal(err)
	}
	for _, path := range []string{fresh, old} {
		f, err := openPrivateLog(path)
		if err != nil {
			t.Fatalf("openPrivateLog(%s): %v", filepath.Base(path), err)
		}
		f.WriteString("line\n")
		f.Close()
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("%s has mode %v, want -rw-------", filepath.Base(path), perm)
		}
	}
	if data, _ := os.ReadFile(old); string(data) != "earlier\nline\n" {
		t.Errorf("existing log not appended to: %q", data)
	}
}

func TestOpenDebugLog_IsPrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows doesn't have Unix permission bits")
	}
	t.Setenv("WA_DEBUG", "1")
	dir := t.TempDir()
	openDebugLog(dir)
	if debugFile == nil {
		t.Fatal("debug log not opened")
	}
	t.Cleanup(func() {
		debugFile.Close()
		debugFile = nil
	})
	info, err := os.Stat(filepath.Join(dir, "wa-debug.log"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("wa-debug.log has mode %v, want -rw-------", perm)
	}
}
