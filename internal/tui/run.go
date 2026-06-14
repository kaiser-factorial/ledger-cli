package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the Bubble Tea TUI program
func Run(initialModel Model) error {
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	_, err := p.Run()
	return err
}