package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/vanderhaka/treework/internal/config"
	"github.com/vanderhaka/treework/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "treework",
	Short: "Git worktree manager",
	Long:  ui.Banner(),
	Run:   runRoot,
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(versionCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() {
	if _, err := exec.LookPath("git"); err != nil {
		ui.Error("git is not installed. Please install git and try again.")
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(ui.Banner())

	// First-run setup: prompt for base folder if not configured
	if config.DevDir() == "" {
		fmt.Println()
		ui.Info("Welcome! Let's set your base folder (where your git repos live).")
		fmt.Println()

		home, _ := os.UserHomeDir()
		selected, err := SetBaseDir(home)
		if err != nil {
			// User escaped — don't save empty config, just continue
			// They'll be prompted again next time or can use 'treework settings'
			fmt.Println()
			ui.Muted("Skipped — you can set it later with 'treework settings' or set DEV_DIR.")
		} else {
			cfg := &config.Config{BaseDir: selected}

			// Prompt for editor choice
			fmt.Println()
			ui.Info("Choose your editor (you can change this later in Settings).")
			fmt.Println()
			editor, edErr := SetEditor()
			if edErr != nil {
				ui.Muted("Skipped — editor will be auto-detected. Change it in Settings.")
			} else {
				cfg.Editor = editor
			}

			if saveErr := config.Save(cfg); saveErr != nil {
				ui.Warn("Could not save config")
			} else {
				ui.Success(fmt.Sprintf("Base folder set to %s", selected))
				if cfg.Editor != "" {
					ui.Success(fmt.Sprintf("Editor set to %s", cfg.Editor))
				} else {
					ui.Muted("Editor: auto-detect")
				}
			}
		}
	}

	for {
		fmt.Println()
		action, err := ui.SelectAction()
		if err != nil {
			handleAbort(err)
			return
		}

		switch action {
		case "new":
			runNewInteractive(cmd)
		case "ls":
			runLsInteractive(cmd)
		case "rm":
			runRmInteractive(cmd)
		case "clear":
			runClearInteractive(cmd)
		case "settings":
			runSettingsInteractive()
		case "quit":
			fmt.Println()
			ui.Muted("Goodbye.")
			fmt.Println()
			return
		}
	}
}
