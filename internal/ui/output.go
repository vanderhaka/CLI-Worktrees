package ui

import "fmt"

func Success(msg string) {
	fmt.Println(SuccessStyle.Render("  ✓ ") + msg)
}

func Info(msg string) {
	fmt.Println(InfoStyle.Render("  ℹ ") + msg)
}

func Warn(msg string) {
	fmt.Println(WarnStyle.Render("  ⚠ ") + msg)
}

func Error(msg string) {
	fmt.Println(ErrorStyle.Render("  ✗ ") + msg)
}

func Muted(msg string) {
	fmt.Println(MutedStyle.Render("    " + msg))
}
