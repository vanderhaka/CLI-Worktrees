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
	doRm(true)
}

func runRmInteractive(cmd *cobra.Command) {
	doRm(false)
}

func doRm(direct bool) {
	devDir := config.DevDir()
	if _, err := os.Stat(devDir); err != nil {
		ui.Error(fmt.Sprintf("DEV_DIR not found: %s", devDir))
		if direct {
			os.Exit(1)
		}
		return
	}

	dirs := git.FindWorktreeDirs(devDir)
	if len(dirs) == 0 {
		ui.Info("No worktrees found.")
		return
	}

	selected, err := ui.SelectWorktree(dirs)
	if err != nil {
		if isAbort(err) {
			if direct {
				handleAbort(err)
			}
			return
		}
		ui.Error(err.Error())
		if direct {
			os.Exit(1)
		}
		return
	}

	if selected == ui.BackValue {
		return
	}

	branch := git.CurrentBranch(selected)
	mainDir := git.MainWorktreePath(selected)
	if mainDir == "" {
		ui.Error("Can't find main repo for this worktree.")
		if direct {
			os.Exit(1)
		}
		return
	}

	ui.Info(fmt.Sprintf("Removing: %s (branch: %s)", filepath.Base(selected), branch))

	// Safety check: look for unsaved work before removing
	status := git.CheckWorktreeStatus(selected)
	forceNeeded := status.IsDirty()

	if forceNeeded {
		ui.WarnDirtyWorktree(status.HasUncommittedChanges, status.HasUnpushedCommits)

		confirmed, confirmErr := ui.ConfirmDirtyRemove()
		if confirmErr != nil {
			if isAbort(confirmErr) {
				if direct {
					handleAbort(confirmErr)
				}
				return
			}
		}
		if !confirmed {
			ui.Muted("Kept worktree — no changes made")
			return
		}
	}

	var removeErr error
	err = spinner.New().
		Title("Removing worktree...").
		Action(func() {
			if forceNeeded {
				removeErr = git.WorktreeForceRemove(mainDir, selected)
			} else {
				removeErr = git.WorktreeRemove(mainDir, selected)
			}
			git.WorktreePrune(mainDir)
		}).
		Run()

	if err != nil {
		if isAbort(err) {
			if direct {
				handleAbort(err)
			}
			return
		}
		ui.Error(err.Error())
		if direct {
			os.Exit(1)
		}
		return
	}

	if removeErr != nil {
		ui.Error("Failed to remove worktree.")
		if direct {
			os.Exit(1)
		}
		return
	}

	ui.Success("Removed worktree")

	if branch != "" && branch != "HEAD" && branch != "main" && branch != "master" {
		if git.IsBranchMerged(mainDir, branch) {
			if err := git.DeleteBranch(mainDir, branch); err == nil {
				ui.Success(fmt.Sprintf("Deleted merged branch '%s'", branch))
			}
		} else {
			ui.Warn(fmt.Sprintf("Branch '%s' is not merged", branch))
			forceDelete, err := ui.ConfirmForceDelete(branch)
			if err != nil {
				if isAbort(err) {
					if direct {
						handleAbort(err)
					}
					return
				}
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
}
