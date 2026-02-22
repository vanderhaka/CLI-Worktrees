package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/jamesvanderhaak/wt/internal/config"
	"github.com/jamesvanderhaak/wt/internal/git"
	"github.com/jamesvanderhaak/wt/internal/ui"
)

// resolveRepo returns a repo directory — either the current git repo or one picked interactively.
func resolveRepo() (string, error) {
	// Try current directory first
	if repo := git.CurrentRepo(); repo != "" {
		return repo, nil
	}

	// Scan DEV_DIR for repos
	devDir := config.DevDir()
	if _, err := os.Stat(devDir); err != nil {
		return "", fmt.Errorf("DEV_DIR not found: %s", devDir)
	}

	repos := git.ScanRepos(devDir)
	if len(repos) == 0 {
		return "", fmt.Errorf("no git repos found in %s", devDir)
	}

	selected, err := ui.SelectRepo(repos)
	if err != nil {
		return "", handleAbort(err)
	}

	return selected, nil
}

// handleAbort converts huh.ErrUserAborted into a clean exit.
func handleAbort(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Println()
		ui.Muted("Cancelled.")
		os.Exit(0)
	}
	return err
}

// resolveWorktreePath resolves a worktree path to an absolute path.
func resolveWorktreePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
