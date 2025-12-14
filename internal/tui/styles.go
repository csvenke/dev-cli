package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var renderer = lipgloss.NewRenderer(os.Stderr)

var (
	blue  = lipgloss.Color("4")
	gray  = lipgloss.Color("8")
	white = lipgloss.Color("15")

	borderStyle = renderer.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blue).
			Padding(1, 2)

	inputStyle = renderer.NewStyle().
			Foreground(white)

	selectedStyle = renderer.NewStyle().
			Foreground(blue).
			Bold(true)

	normalStyle = renderer.NewStyle().
			Foreground(white)

	pathStyle = renderer.NewStyle().
			Foreground(gray)

	titleStyle = renderer.NewStyle().
			Foreground(white).
			Bold(true)

	keymapLabelStyle = renderer.NewStyle().
				Foreground(white)

	keymapKeyStyle = renderer.NewStyle().
			Foreground(gray)
)
