package cmd

import (
	"fmt"
	"os"

	"github.com/vanderhaka/treework/internal/config"
	"github.com/vanderhaka/treework/internal/ui"
)

// SetBaseDir runs the shared path-selection flow with retry on invalid paths.
// Returns the chosen directory path or an error.
func SetBaseDir(currentPath string) (string, error) {
	for {
		method, err := ui.SelectPathMethod()
		if err != nil {
			return "", err
		}

		var selected string
		switch method {
		case "type":
			selected, err = ui.InputPath(currentPath)
			if err != nil {
				return "", err
			}
		case "browse":
			startDir := currentPath
			if info, serr := os.Stat(startDir); serr != nil || !info.IsDir() {
				startDir, _ = os.UserHomeDir()
			}
			selected, err = ui.BrowseDirectory(startDir)
			if err != nil {
				return "", err
			}
		}

		// Validate the path exists and is a directory
		info, err := os.Stat(selected)
		if err != nil || !info.IsDir() {
			ui.Warn(fmt.Sprintf("'%s' is not a valid directory. Try again.", selected))
			continue
		}

		return selected, nil
	}
}

// SetEditor runs the shared editor-selection flow.
// Returns the editor command string ("" for auto-detect) or an error.
func SetEditor() (string, error) {
	choice, err := ui.SelectEditor()
	if err != nil {
		return "", err
	}

	switch choice {
	case "auto":
		return "", nil
	case "custom":
		cmd, err := ui.InputEditorCommand()
		if err != nil {
			return "", err
		}
		if cmd == "" {
			return "", nil
		}
		return cmd, nil
	default:
		return choice, nil
	}
}

// doChangeBaseDir shows the current base folder and lets the user change it.
func doChangeBaseDir() {
	cfg := config.Load()
	current := config.DevDir()

	fmt.Println()
	ui.Info(fmt.Sprintf("Base folder: %s", current))
	if cfg.BaseDir != "" {
		ui.Muted("(from config file)")
	} else if os.Getenv("DEV_DIR") != "" {
		ui.Muted("(from DEV_DIR env var)")
	} else {
		ui.Muted("(not set)")
	}
	fmt.Println()

	selected, err := SetBaseDir(current)
	if err != nil {
		if isAbort(err) {
			return
		}
		return
	}

	cfg.BaseDir = selected
	if err := config.Save(cfg); err != nil {
		ui.Error(fmt.Sprintf("Failed to save config: %v", err))
		return
	}

	ui.Success(fmt.Sprintf("Base folder set to %s", selected))
}

// doChangeEditor shows the current editor and lets the user change it.
func doChangeEditor() {
	editor, source := config.EditorSource()

	fmt.Println()
	if editor != "" {
		ui.Info(fmt.Sprintf("Editor: %s", editor))
	} else {
		ui.Info("Editor: auto-detect")
	}
	ui.Muted(fmt.Sprintf("(from %s)", source))
	fmt.Println()

	selected, err := SetEditor()
	if err != nil {
		if isAbort(err) {
			return
		}
		return
	}

	cfg := config.Load()
	cfg.Editor = selected
	if err := config.Save(cfg); err != nil {
		ui.Error(fmt.Sprintf("Failed to save config: %v", err))
		return
	}

	if selected == "" {
		ui.Success("Editor set to auto-detect")
	} else {
		ui.Success(fmt.Sprintf("Editor set to %s", selected))
	}
}

// doSettings shows the settings sub-menu in a loop.
func doSettings() {
	for {
		fmt.Println()
		action, err := ui.SelectSettingsAction()
		if err != nil {
			if isAbort(err) {
				return
			}
			return
		}

		switch action {
		case "base_dir":
			doChangeBaseDir()
		case "editor":
			doChangeEditor()
		case ui.BackValue:
			return
		}
	}
}

// runSettingsInteractive is called from the root menu.
func runSettingsInteractive() {
	doSettings()
}
