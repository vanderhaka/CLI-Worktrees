package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesvanderhaak/wt/internal/config"
	"github.com/jamesvanderhaak/wt/internal/editor"
	"github.com/jamesvanderhaak/wt/internal/git"
	"github.com/jamesvanderhaak/wt/internal/ui"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List worktrees and optionally open one",
	Run:     runLs,
}

func runLs(cmd *cobra.Command, args []string) {
	fmt.Println()
	doLs(true)
}

// runLsInteractive is called from the root menu loop. Returns instead of os.Exit.
func runLsInteractive(cmd *cobra.Command) {
	doLs(false)
}

func doLs(exitOnError bool) {
	devDir := config.DevDir()
	if _, err := os.Stat(devDir); err != nil {
		ui.Error(fmt.Sprintf("DEV_DIR not found: %s", devDir))
		if exitOnError {
			os.Exit(1)
		}
		return
	}

	dirs := git.FindWorktreeDirs(devDir)
	if len(dirs) == 0 {
		ui.Info("No worktrees found.")
		return
	}

	// Build detailed display items with branch info
	var items []ui.WorktreeDisplay
	for _, d := range dirs {
		branch := git.CurrentBranch(d)
		base := filepath.Base(d)
		repo := extractRepoName(base)
		items = append(items, ui.WorktreeDisplay{
			Path:   d,
			Branch: branch,
			Repo:   repo,
		})
	}

	selected, err := ui.SelectWorktreeDetailed(items)
	if err != nil {
		handleAbort(err)
		if exitOnError {
			os.Exit(1)
		}
		return
	}

	if selected == ui.BackValue {
		return
	}

	// Confirm before opening
	open, err := ui.ConfirmOpen(filepath.Base(selected))
	if err != nil {
		handleAbort(err)
		return
	}

	if open {
		editor.Open(selected)
		ui.Success(fmt.Sprintf("Opened: %s", filepath.Base(selected)))
	} else {
		ui.Muted(selected)
	}
}

// extractRepoName gets the repo name from a worktree dir name.
// e.g. "myapp-worktree-feature" → "myapp"
func extractRepoName(wtDirName string) string {
	idx := len(wtDirName)
	const marker = "-worktree-"
	for i := 0; i <= len(wtDirName)-len(marker); i++ {
		if wtDirName[i:i+len(marker)] == marker {
			idx = i
			break
		}
	}
	if idx < len(wtDirName) {
		return wtDirName[:idx]
	}
	return ""
}
