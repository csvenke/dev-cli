package main

import "github.com/charmbracelet/lipgloss"

var (
	blue  = lipgloss.Color("4")
	gray  = lipgloss.Color("8")
	white = lipgloss.Color("15")

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blue).
			Padding(1, 2)

	inputStyle = lipgloss.NewStyle().
			Foreground(white)

	selectedStyle = lipgloss.NewStyle().
			Foreground(blue).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(white)

	pathStyle = lipgloss.NewStyle().
			Foreground(gray)

	titleStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	keymapLabelStyle = lipgloss.NewStyle().
				Foreground(white)

	keymapKeyStyle = lipgloss.NewStyle().
			Foreground(gray)
)
