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

	devDir := config.DevDir()
	if _, err := os.Stat(devDir); err != nil {
		ui.Error(fmt.Sprintf("DEV_DIR not found: %s", devDir))
		os.Exit(1)
	}

	dirs := git.FindWorktreeDirs(devDir)
	if len(dirs) == 0 {
		ui.Info("No worktrees found.")
		fmt.Println()
		return
	}

	// Build detailed display items with branch info
	var items []ui.WorktreeDisplay
	for _, d := range dirs {
		branch := git.CurrentBranch(d)
		// Extract repo name from worktree dir name (e.g. "myapp-worktree-feat" → "myapp")
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
		ui.Error(err.Error())
		os.Exit(1)
	}

	// Confirm before opening
	open, err := ui.ConfirmOpen(filepath.Base(selected))
	if err != nil {
		handleAbort(err)
		return
	}

	if open {
		editor.Open(selected)
		fmt.Println()
		ui.Success(fmt.Sprintf("Opened: %s", filepath.Base(selected)))
	} else {
		fmt.Println()
		ui.Muted(selected)
	}
	fmt.Println()
}

// extractRepoName gets the repo name from a worktree dir name.
// e.g. "myapp-worktree-feature" → "myapp"
func extractRepoName(wtDirName string) string {
	idx := len(wtDirName)
	// Find "-worktree-" in the name
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
