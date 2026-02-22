package ui

import "github.com/charmbracelet/lipgloss"

// Colour palette
var (
	Green   = lipgloss.Color("#22c55e")
	Cyan    = lipgloss.Color("#06b6d4")
	Yellow  = lipgloss.Color("#eab308")
	Red     = lipgloss.Color("#ef4444")
	Magenta = lipgloss.Color("#d946ef")
	Gray    = lipgloss.Color("#6b7280")
)

// Text styles
var (
	SuccessStyle = lipgloss.NewStyle().Foreground(Green).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(Cyan)
	WarnStyle    = lipgloss.NewStyle().Foreground(Yellow)
	ErrorStyle   = lipgloss.NewStyle().Foreground(Red).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(Gray)
	BrandStyle   = lipgloss.NewStyle().Foreground(Magenta).Bold(true)
	BoldStyle    = lipgloss.NewStyle().Bold(true)
)

// Banner prints the wt branding header.
func Banner() string {
	return BrandStyle.Render("wt") + MutedStyle.Render(" — git worktree manager")
}
