package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vanderhaka/treework/internal/config"
	"github.com/vanderhaka/treework/internal/editor"
	"github.com/vanderhaka/treework/internal/git"
	"github.com/vanderhaka/treework/internal/ui"
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

func runLsInteractive(cmd *cobra.Command) {
	doLs(false)
}

func doLs(direct bool) {
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
		if isAbort(err) {
			if direct {
				handleAbort(err)
			}
			return // back to menu
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

	open, err := ui.ConfirmOpen(filepath.Base(selected))
	if err != nil {
		if isAbort(err) {
			if direct {
				handleAbort(err)
			}
			return
		}
		return
	}

	if open {
		if err := editor.Open(selected); err != nil {
			ui.Warn(fmt.Sprintf("Could not open editor: %v", err))
		} else {
			ui.Success(fmt.Sprintf("Opened: %s", filepath.Base(selected)))
		}
	} else {
		ui.Muted(selected)
	}
}

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
