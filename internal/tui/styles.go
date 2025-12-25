package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors - Using a refined color palette
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	accentColor    = lipgloss.Color("#A78BFA") // Light purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	dimColor       = lipgloss.Color("#4B5563") // Darker gray
	errorColor     = lipgloss.Color("#EF4444") // Red
	warningColor   = lipgloss.Color("#F59E0B") // Yellow
	bgColor        = lipgloss.Color("#1F2937") // Dark background
)

// Styles
var (
	// App frame
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Logo style for "regis3"
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	// Logo box - the main header container
	logoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 2).
			MarginBottom(1)

	// Version/info text in header
	versionStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	authorStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Info box for paths
	infoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimColor).
			Padding(0, 2).
			MarginBottom(1)

	infoLabelStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Width(12)

	infoValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	// Header (for other views)
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
			BorderForeground(dimColor).
			Padding(1, 2).
			Width(18)

	cardActiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Width(18)

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

	// Selected text style (no padding, just color)
	selectedTextStyle = lipgloss.NewStyle().
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

	// Spinner style
	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)
)
