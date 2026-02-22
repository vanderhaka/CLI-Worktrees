package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh/spinner"
	"github.com/jamesvanderhaak/wt/internal/git"
	"github.com/jamesvanderhaak/wt/internal/ui"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove ALL worktrees for a repo",
	Run:   runClear,
}

// runClearInteractive is called from the root menu loop.
func runClearInteractive(cmd *cobra.Command) {
	doClear(false)
}

func runClear(cmd *cobra.Command, args []string) {
	doClear(true)
}

func doClear(direct bool) {
	fmt.Println()

	// 1. Resolve repo — interactive menu always shows the project list
	repoDir, err := resolveRepo(!direct)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	repoName := filepath.Base(repoDir)

	// 2. List worktrees (excluding main)
	worktrees := git.WorktreeList(repoDir)
	if len(worktrees) == 0 {
		ui.Info("No worktrees to remove.")
		fmt.Println()
		return
	}

	// 3. Display list
	ui.Info(fmt.Sprintf("Worktrees for %s:", ui.BoldStyle.Render(repoName)))
	for _, wt := range worktrees {
		ui.Muted(fmt.Sprintf("%s  %s", filepath.Base(wt.Path), ui.MutedStyle.Render("("+wt.Branch+")")))
	}
	fmt.Println()

	// 4. Safety check: identify dirty worktrees
	type worktreeCheck struct {
		info   git.WorktreeInfo
		status git.WorktreeStatus
	}
	var clean, dirty []worktreeCheck
	for _, wt := range worktrees {
		s := git.CheckWorktreeStatus(wt.Path)
		entry := worktreeCheck{info: wt, status: s}
		if s.IsDirty() {
			dirty = append(dirty, entry)
		} else {
			clean = append(clean, entry)
		}
	}

	// 5. Show dirty worktree warnings
	if len(dirty) > 0 {
		fmt.Println()
		ui.Warn(fmt.Sprintf("%d worktree(s) have unsaved work:", len(dirty)))
		for _, d := range dirty {
			reasons := ""
			if d.status.HasUncommittedChanges && d.status.HasUnpushedCommits {
				reasons = "uncommitted changes + unpushed commits"
			} else if d.status.HasUncommittedChanges {
				reasons = "uncommitted changes"
			} else {
				reasons = "unpushed commits"
			}
			ui.Muted(fmt.Sprintf("  • %s (%s) — %s", filepath.Base(d.info.Path), d.info.Branch, reasons))
		}
		fmt.Println()
	}

	// 6. Confirm removal
	var confirmed bool
	if len(dirty) > 0 {
		confirmed, err = ui.Confirm(fmt.Sprintf("Remove all %d worktrees? Unsaved work will be permanently lost", len(worktrees)))
	} else {
		confirmed, err = ui.Confirm(fmt.Sprintf("Remove all %d worktrees?", len(worktrees)))
	}
	if err != nil {
		handleAbort(err)
		ui.Error(err.Error())
		os.Exit(1)
	}
	if !confirmed {
		ui.Muted("Cancelled.")
		fmt.Println()
		return
	}

	// 7. Remove each worktree
	err = spinner.New().
		Title("Removing worktrees...").
		Action(func() {
			for _, wt := range worktrees {
				// Use force only for dirty worktrees (user already confirmed)
				status := git.CheckWorktreeStatus(wt.Path)
				if status.IsDirty() {
					git.WorktreeForceRemove(repoDir, wt.Path)
				} else {
					git.WorktreeRemove(repoDir, wt.Path)
				}

				// Auto-delete merged branches
				if wt.Branch != "" && wt.Branch != "main" && wt.Branch != "master" {
					if git.IsBranchMerged(repoDir, wt.Branch) {
						git.DeleteBranch(repoDir, wt.Branch)
					}
				}
			}
			git.WorktreePrune(repoDir)
		}).
		Run()

	if err != nil {
		handleAbort(err)
		ui.Error(err.Error())
		os.Exit(1)
	}

	fmt.Println()
	ui.Success("All worktrees cleared")
	ui.Muted("Merged branches were auto-deleted")
	fmt.Println()
}
