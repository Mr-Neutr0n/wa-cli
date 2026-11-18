package store

// NOTE: written without the ability to compile/run in the authoring
// sandbox — see the comment at the top of internal/whatsapp/client_test.go
// for why. Please run `go test ./internal/store/...` locally.

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	waLog "go.mau.fi/whatsmeow/util/log"
)

func TestOpen_CreatesDataDirAndDBFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "data")
	log := waLog.Stdout("test", "ERROR", false)

	container, err := Open(context.Background(), dir, log)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if container == nil {
		t.Fatal("expected non-nil container")
	}
	// Windows can't delete an open file when the temp dir is cleaned up.
	t.Cleanup(func() { container.Close() })

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected data dir to be created: %v", err)
	}

	dbPath := filepath.Join(dir, DBFileName)
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("expected %s to be created: %v", DBFileName, err)
	}
}

// checkPrivate fails the test if any session file in dir can be read or
// written by anyone but its owner.
func checkPrivate(t *testing.T, dir string) {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(dir, DBFileName+"*"))
	if len(files) == 0 {
		t.Fatalf("no %s files in %s", DBFileName, dir)
	}
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm&0o077 != 0 {
			t.Errorf("%s has mode %v, want no access for group or others", filepath.Base(f), perm)
		}
	}
}

func TestOpen_SessionDBIsPrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows doesn't have Unix permission bits")
	}
	dir := t.TempDir()
	container, err := Open(context.Background(), dir, waLog.Noop)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { container.Close() })
	checkPrivate(t, dir)
}

func TestOpen_TightensExistingSessionDB(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows doesn't have Unix permission bits")
	}
	dir := t.TempDir()
	for _, name := range []string{DBFileName, DBFileName + "-wal"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o644); err != nil { // past the umask
			t.Fatal(err)
		}
	}
	container, err := Open(context.Background(), dir, waLog.Noop)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { container.Close() })
	checkPrivate(t, dir)
}

func TestOpen_ReopensExistingStore(t *testing.T) {
	dir := t.TempDir()
	log := waLog.Stdout("test", "ERROR", false)

	first, err := Open(context.Background(), dir, log)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	t.Cleanup(func() { first.Close() })
	// Reopening the same dataDir should succeed against the existing
	// session.db rather than erroring out.
	second, err := Open(context.Background(), dir, log)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	t.Cleanup(func() { second.Close() })
}
