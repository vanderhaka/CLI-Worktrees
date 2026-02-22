package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh/spinner"
	"github.com/jamesvanderhaak/wt/internal/config"
	"github.com/jamesvanderhaak/wt/internal/git"
	"github.com/jamesvanderhaak/wt/internal/ui"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm",
	Aliases: []string{"remove"},
	Short:   "Remove a worktree",
	Run:     runRm,
}

func runRm(cmd *cobra.Command, args []string) {
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

	selected, err := ui.SelectWorktree(dirs)
	if err != nil {
		handleAbort(err)
		ui.Error(err.Error())
		os.Exit(1)
	}

	branch := git.CurrentBranch(selected)
	mainDir := git.MainWorktreePath(selected)
	if mainDir == "" {
		ui.Error("Can't find main repo for this worktree.")
		os.Exit(1)
	}

	ui.Info(fmt.Sprintf("Removing: %s (branch: %s)", filepath.Base(selected), branch))

	// Remove worktree with spinner
	var removeErr error
	err = spinner.New().
		Title("Removing worktree...").
		Action(func() {
			removeErr = git.WorktreeRemove(mainDir, selected)
			git.WorktreePrune(mainDir)
		}).
		Run()

	if err != nil {
		handleAbort(err)
		ui.Error(err.Error())
		os.Exit(1)
	}

	if removeErr != nil {
		ui.Error("Failed to remove worktree. Check for uncommitted changes.")
		os.Exit(1)
	}

	ui.Success("Removed worktree")

	// Branch cleanup
	if branch != "" && branch != "HEAD" && branch != "main" && branch != "master" {
		if git.IsBranchMerged(mainDir, branch) {
			if err := git.DeleteBranch(mainDir, branch); err == nil {
				ui.Success(fmt.Sprintf("Deleted merged branch '%s'", branch))
			}
		} else {
			ui.Warn(fmt.Sprintf("Branch '%s' is not merged", branch))
			forceDelete, err := ui.ConfirmForceDelete(branch)
			if err != nil {
				handleAbort(err)
			}
			if forceDelete {
				if err := git.ForceDeleteBranch(mainDir, branch); err == nil {
					ui.Success(fmt.Sprintf("Force deleted branch '%s'", branch))
				} else {
					ui.Error(fmt.Sprintf("Failed to delete branch '%s'", branch))
				}
			} else {
				ui.Muted(fmt.Sprintf("Kept branch '%s'", branch))
			}
		}
	}

	fmt.Println()
}
