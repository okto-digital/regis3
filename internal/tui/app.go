package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/okto-digital/regis3/internal/config"
)

// Run starts the TUI application
func Run(cfg *config.Config, registryPath, projectPath string) error {
	model := NewModel(cfg, registryPath, projectPath)

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return err
	}

	return nil
}
