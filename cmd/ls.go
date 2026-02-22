package cmd

import (
	"fmt"
	"os"

	"github.com/jamesvanderhaak/wt/internal/config"
	"github.com/jamesvanderhaak/wt/internal/editor"
	"github.com/jamesvanderhaak/wt/internal/git"
	"github.com/jamesvanderhaak/wt/internal/ui"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List and open a worktree",
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

	selected, err := ui.SelectWorktree(dirs)
	if err != nil {
		handleAbort(err)
		ui.Error(err.Error())
		os.Exit(1)
	}

	editor.Open(selected)
	fmt.Println()
	ui.Success(fmt.Sprintf("Opened: %s", selected))
	fmt.Println()
}
