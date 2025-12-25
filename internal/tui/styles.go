package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	errorColor     = lipgloss.Color("#EF4444") // Red
	warningColor   = lipgloss.Color("#F59E0B") // Yellow
)

// Styles
var (
	// App frame
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Header
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(mutedColor).
			Padding(0, 1).
			MarginBottom(1)

	// Cards for dashboard
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			Width(20)

	cardActiveStyle = cardStyle.Copy().
			BorderForeground(primaryColor)

	cardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	cardValueStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// List items
	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(primaryColor).
				Bold(true)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(mutedColor).
			Padding(0, 1).
			MarginTop(1)

	// Help text
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Section headers
	sectionStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Bold(true)

	// Success/Error/Warning
	successStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	// Subtle text
	subtleStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)
