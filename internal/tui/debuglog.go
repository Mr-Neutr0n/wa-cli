package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// debugFile is non-nil only when WA_DEBUG=1. Debug output must never go to
// stdout — that's Bubble Tea's screen.
var debugFile *os.File

func openDebugLog(dataDir string) {
	if os.Getenv("WA_DEBUG") != "1" {
		return
	}
	f, err := openPrivateLog(filepath.Join(dataDir, "wa-debug.log"))
	if err == nil {
		debugFile = f
	}
}

// openPrivateLog opens the log file at path for appending, creating it if
// needed, readable by the user only. A log left readable by everyone by
// an older version is tightened first.
func openPrivateLog(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func dlog(format string, args ...any) {
	if debugFile == nil {
		return
	}
	fmt.Fprintf(debugFile, "%s "+format+"\n", append([]any{time.Now().Format("15:04:05.000")}, args...)...)
}
