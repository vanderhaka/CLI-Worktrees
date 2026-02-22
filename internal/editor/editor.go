package editor

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/vanderhaka/treework/internal/config"
)

// Open opens the given path in the preferred editor.
// Priority: WT_EDITOR env var > cursor > code > platform default.
func Open(path string) error {
	if ed := config.Editor(); ed != "" {
		return run(ed, path)
	}

	editors := []struct {
		cmd  string
		args []string
	}{
		{"cursor", []string{"--new-window", path}},
		{"code", []string{"-n", path}},
	}

	for _, e := range editors {
		if _, err := exec.LookPath(e.cmd); err == nil {
			return exec.Command(e.cmd, e.args...).Start()
		}
	}

	// Platform-specific fallback to open the folder
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "windows":
		return exec.Command("explorer", path).Start()
	}

	return fmt.Errorf("no editor found — set WT_EDITOR or install cursor/code")
}

func run(editor, path string) error {
	switch editor {
	case "cursor":
		return exec.Command(editor, "--new-window", path).Start()
	case "code":
		return exec.Command(editor, "-n", path).Start()
	default:
		return exec.Command(editor, path).Start()
	}
}
