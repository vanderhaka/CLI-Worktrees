package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/jamesvanderhaak/wt/internal/deps"
	"github.com/jamesvanderhaak/wt/internal/editor"
	"github.com/jamesvanderhaak/wt/internal/env"
	"github.com/jamesvanderhaak/wt/internal/git"
	"github.com/jamesvanderhaak/wt/internal/sanitize"
	"github.com/jamesvanderhaak/wt/internal/ui"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new worktree",
	Args:  cobra.MaximumNArgs(1),
	Run:   runNew,
}

func runNew(cmd *cobra.Command, args []string) {
	fmt.Println()

	// 1. Resolve repo
	repoDir, err := resolveRepo()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	// 2. Get name
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		name, err = ui.InputName()
		if err != nil {
			handleAbort(err)
			ui.Error(err.Error())
			os.Exit(1)
		}
	}

	name = sanitize.Name(name)
	if name == "" {
		ui.Error("Name became empty after sanitising.")
		os.Exit(1)
	}

	// 3. Compute path
	wtPath := git.WorktreePath(repoDir, name)
	resolved := resolveWorktreePath(wtPath)
	repoName := filepath.Base(repoDir)

	// 4. If exists → open in editor
	if _, err := os.Stat(resolved); err == nil {
		ui.Info(fmt.Sprintf("Already exists: %s-worktree-%s", repoName, name))
		editor.Open(resolved)
		return
	}

	// 5-6. Create worktree (with spinner)
	branchExists := git.BranchExists(repoDir, name)
	var addErr error

	err = spinner.New().
		Title(fmt.Sprintf("Creating worktree %s-worktree-%s...", repoName, name)).
		Action(func() {
			addErr = git.WorktreeAdd(repoDir, resolved, name, !branchExists)
		}).
		Run()

	if err != nil {
		handleAbort(err)
		ui.Error(err.Error())
		os.Exit(1)
	}
	if addErr != nil {
		ui.Error(fmt.Sprintf("Failed to create worktree: %v", addErr))
		os.Exit(1)
	}

	// 7. Copy .env files
	copied, _ := env.CopyEnvFiles(repoDir, resolved)
	if len(copied) > 0 {
		ui.Muted(fmt.Sprintf("Copied %d env file(s)", len(copied)))
	}

	// 8. Detect package manager → prompt to install deps
	if pm := deps.Detect(resolved); pm != nil {
		install, err := ui.ConfirmInstall(pm.Name)
		if err != nil {
			handleAbort(err)
		}
		if install {
			var installErr error
			err = spinner.New().
				Title(fmt.Sprintf("Installing dependencies with %s...", pm.Name)).
				Action(func() {
					installErr = deps.Install(resolved, pm)
					// Give spinner a moment to render
					time.Sleep(100 * time.Millisecond)
				}).
				Context(context.Background()).
				Run()

			if err != nil {
				handleAbort(err)
			}
			if installErr != nil {
				ui.Warn("Install failed. You can run it later inside the folder.")
			} else {
				ui.Success("Dependencies installed")
			}
		}
	}

	// 9. Open in editor
	editor.Open(resolved)

	// 10. Print success
	fmt.Println()
	ui.Success(fmt.Sprintf("Ready: %s-worktree-%s", repoName, name))
	ui.Muted(resolved)
	fmt.Println()
}
