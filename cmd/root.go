package cmd

import (
	"fmt"
	"os"

	"github.com/jamesvanderhaak/wt/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wt",
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
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(ui.Banner())
	fmt.Println()

	action, err := ui.SelectAction()
	if err != nil {
		handleAbort(err)
		return
	}

	switch action {
	case "new":
		runNew(cmd, nil)
	case "ls":
		runLs(cmd, nil)
	case "rm":
		runRm(cmd, nil)
	case "clear":
		runClear(cmd, nil)
	}
}
