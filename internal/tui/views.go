package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewDashboard renders the dashboard view with stats cards
func (m Model) viewDashboard() string {
	var b strings.Builder

	// Logo header box
	logo := `
  ┏━┓┏━╸┏━╸╻┏━┓┏━┓
  ┣┳┛┣╸ ┃╺┓┃┗━┓╺━┫
  ╹┗╸┗━╸┗━┛╹┗━┛┗━┛`
	logoText := logoStyle.Render(logo)

	// Version and author info
	versionInfo := versionStyle.Render("v" + Version)
	authorInfo := authorStyle.Render("by " + Author)
	infoLine := versionInfo + "  " + authorInfo

	headerContent := logoText + "\n" + infoLine
	headerBox := logoBoxStyle.Render(headerContent)
	b.WriteString(headerBox)
	b.WriteString("\n")

	// Info box with registry and project paths
	registryLine := infoLabelStyle.Render("Registry") + infoValueStyle.Render(m.registryPath)
	projectLine := infoLabelStyle.Render("Project") + infoValueStyle.Render(m.projectPath)
	infoContent := registryLine + "\n" + projectLine
	infoBox := infoBoxStyle.Render(infoContent)
	b.WriteString(infoBox)
	b.WriteString("\n")

	// Stats cards
	registryCount := 0
	if m.manifest != nil {
		registryCount = len(m.manifest.Items)
	}

	// Get installed count for project card
	installedCount := len(m.getInstalledItems())

	// Create cards with numbers matching keyboard shortcuts
	registryCard := m.renderCard("1. Registry", fmt.Sprintf("%d items", registryCount), m.cursor == 0)
	projectCard := m.renderCard("2. Project", fmt.Sprintf("%d installed", installedCount), m.cursor == 1)
	settingsCard := m.renderCard("3. Settings", "Actions", m.cursor == 2)

	// Build card row - conditionally add staging card if there are pending items
	if len(m.stagingItems) > 0 {
		stagingCard := m.renderCard("4. Staging", fmt.Sprintf("%d pending", len(m.stagingItems)), m.cursor == 3)
		cards := lipgloss.JoinHorizontal(lipgloss.Top, registryCard, " ", projectCard, " ", settingsCard, " ", stagingCard)
		b.WriteString(cards)
	} else {
		cards := lipgloss.JoinHorizontal(lipgloss.Top, registryCard, " ", projectCard, " ", settingsCard)
		b.WriteString(cards)
	}
	b.WriteString("\n\n")

	// Status message
	if m.statusMsg != "" {
		if m.isError {
			b.WriteString(errorStyle.Render("✗ " + m.statusMsg))
		} else {
			b.WriteString(successStyle.Render("✓ " + m.statusMsg))
		}
		b.WriteString("\n\n")
	}

	// Help - include staging if there are pending items
	helpItems := []helpItem{
		{"1", "Registry"},
		{"2", "Project"},
		{"3", "Settings"},
	}
	if len(m.stagingItems) > 0 {
		helpItems = append(helpItems, helpItem{"4", "Staging"})
	}
	helpItems = append(helpItems, helpItem{"q", "Quit"})
	help := m.renderHelp(helpItems)
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewRegistry renders the registry browser view
func (m Model) viewRegistry() string {
	var b strings.Builder

	if m.manifest == nil {
		b.WriteString(headerStyle.Render("Registry Browser"))
		b.WriteString("\n\n")
		b.WriteString(warningStyle.Render("No registry loaded. Press 'b' to build."))
		b.WriteString("\n\n")
		help := m.renderHelp([]helpItem{
			{"b", "build"},
			{"esc", "back"},
		})
		b.WriteString(help)
		return appStyle.Render(b.String())
	}

	// The list component handles its own rendering including help
	b.WriteString(m.registryList.View())

	// Show status message if any
	if m.statusMsg != "" {
		b.WriteString("\n")
		if m.isError {
			b.WriteString(errorStyle.Render(m.statusMsg))
		} else {
			b.WriteString(successStyle.Render(m.statusMsg))
		}
	}

	// Show selection count
	if len(m.selected) > 0 {
		b.WriteString("\n")
		b.WriteString(subtleStyle.Render(fmt.Sprintf("%d item(s) selected - press 'a' to add to project", len(m.selected))))
	}

	return appStyle.Render(b.String())
}

// viewProject renders the project view with installed items
func (m Model) viewProject() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("Project: " + m.projectPath)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Get installed items
	installedItems := m.getInstalledItems()

	if len(installedItems) == 0 {
		b.WriteString(subtleStyle.Render("No items installed in this project"))
		b.WriteString("\n")
		b.WriteString(subtleStyle.Render("Go to Registry (1) and press 'a' to add items"))
	} else {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("Installed (%d):", len(installedItems))))
		b.WriteString("\n\n")
		for i, item := range installedItems {
			isSelected := i == m.projectCursor

			// Status prefix (checkmark for installed, warning for needs update)
			statusPrefix := "✓ "
			if item.needsUpdate {
				statusPrefix = "⚠ "
			}

			// First line: title with status prefix
			titleLine := statusPrefix + item.ref

			// Second line: description
			desc := item.desc
			if desc == "" {
				desc = "(no description)"
			}
			if len(desc) > 70 {
				desc = desc[:67] + "..."
			}

			if isSelected {
				// Selected item with border
				b.WriteString("  │ ")
				b.WriteString(selectedTextStyle.Render(titleLine))
				b.WriteString("\n")
				b.WriteString("  │ ")
				b.WriteString(subtleStyle.Render(desc))
				b.WriteString("\n")
			} else {
				// Normal item
				b.WriteString("    ")
				b.WriteString(titleLine)
				b.WriteString("\n")
				b.WriteString("    ")
				b.WriteString(subtleStyle.Render(desc))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Status message
	if m.statusMsg != "" {
		if m.isError {
			b.WriteString(errorStyle.Render(m.statusMsg))
		} else {
			b.WriteString(successStyle.Render(m.statusMsg))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help
	help := m.renderHelp([]helpItem{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"r", "remove"},
		{"1", "registry"},
		{"esc", "back"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}


// viewSettings renders the settings view
func (m Model) viewSettings() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("Settings & Actions")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Configuration section
	b.WriteString(sectionStyle.Render("Configuration:"))
	b.WriteString("\n\n")

	// Render settings items
	for i, item := range m.settingsItems {
		isSelected := i == m.settingsCursor

		// Add section separator between paths and actions
		if i == 2 {
			b.WriteString("\n")
			b.WriteString(sectionStyle.Render("Actions:"))
			b.WriteString("\n\n")
		}

		if isSelected {
			b.WriteString("  │ ")
			if m.settingsEditing && item.isPath {
				// Show text input when editing path
				b.WriteString(selectedTextStyle.Render(item.label))
				b.WriteString("  ")
				b.WriteString(m.settingsInput.View())
			} else if m.settingsScanMode && item.actionKey == "scan" {
				// Show text input when entering scan path
				b.WriteString(selectedTextStyle.Render(item.label))
				b.WriteString("  ")
				b.WriteString(m.settingsInput.View())
			} else {
				b.WriteString(selectedTextStyle.Render(item.label))
				b.WriteString("  ")
				if item.isPath {
					b.WriteString(item.value)
				} else {
					b.WriteString(subtleStyle.Render(item.value))
				}
			}
		} else {
			b.WriteString("    ")
			b.WriteString(item.label)
			b.WriteString("  ")
			if item.isPath {
				b.WriteString(subtleStyle.Render(item.value))
			} else {
				b.WriteString(subtleStyle.Render(item.value))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Status message
	if m.statusMsg != "" {
		if m.isError {
			b.WriteString(errorStyle.Render(m.statusMsg))
		} else {
			b.WriteString(successStyle.Render(m.statusMsg))
		}
		b.WriteString("\n\n")
	}

	// Help
	var help string
	if m.settingsEditing {
		help = m.renderHelp([]helpItem{
			{"enter", "save"},
			{"esc", "cancel"},
		})
	} else if m.settingsScanMode {
		help = m.renderHelp([]helpItem{
			{"enter", "scan"},
			{"esc", "cancel"},
		})
	} else {
		help = m.renderHelp([]helpItem{
			{"↑/k", "up"},
			{"↓/j", "down"},
			{"enter", "edit/run"},
			{"esc", "back"},
		})
	}
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewDetail renders the item detail view
func (m Model) viewDetail() string {
	// Handle edit mode first
	if m.detailEditMode {
		return m.viewDetailEdit()
	}

	var b strings.Builder

	if m.detailItem == nil {
		b.WriteString("No item selected")
		return appStyle.Render(b.String())
	}

	item := m.detailItem
	files := m.getDetailFiles()

	// Header
	header := headerStyle.Render(item.FullName())
	b.WriteString(header)
	b.WriteString("\n\n")

	// Item info on single line
	info := fmt.Sprintf("Type: %s", titleStyle.Render(item.Type))
	if len(item.Tags) > 0 {
		info += fmt.Sprintf("    Tags: %s", strings.Join(item.Tags, ", "))
	}
	if item.Status != "" {
		info += fmt.Sprintf("    Status: %s", item.Status)
	}
	b.WriteString("  " + info)
	b.WriteString("\n\n")

	// Files list with numbers for selection
	if len(files) > 0 {
		b.WriteString(sectionStyle.Render("Files:"))
		b.WriteString("\n")
		for i, f := range files {
			prefix := "  "
			if i == m.detailFileIndex {
				prefix = "> "
			}
			label := f
			if i == 0 {
				label += " (main)"
			}
			b.WriteString(fmt.Sprintf("%s[%d] %s\n", prefix, i+1, label))
		}
		b.WriteString("\n")
	}

	// Separator line
	b.WriteString(subtleStyle.Render(strings.Repeat("─", 60)))
	b.WriteString("\n\n")

	// File content (rendered with glamour for .md files)
	if m.detailLoading {
		b.WriteString(m.spinner.View() + " " + subtleStyle.Render("Loading file content..."))
	} else if m.detailContent != "" {
		// Limit content height to leave room for other elements
		lines := strings.Split(m.detailContent, "\n")
		maxLines := m.height - 20
		if maxLines < 10 {
			maxLines = 10
		}
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, subtleStyle.Render("... (content truncated)"))
		}
		b.WriteString(strings.Join(lines, "\n"))
	} else {
		b.WriteString(subtleStyle.Render("No content to display"))
	}
	b.WriteString("\n\n")

	// Status message
	if m.statusMsg != "" {
		if m.isError {
			b.WriteString(errorStyle.Render(m.statusMsg))
		} else {
			b.WriteString(successStyle.Render(m.statusMsg))
		}
		b.WriteString("\n")
	}

	// Help
	help := m.renderHelp([]helpItem{
		{"1-9", "select file"},
		{"e", "edit"},
		{"a", "add to project"},
		{"esc", "back"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewDetailEdit renders the inline editor for registry item files
func (m Model) viewDetailEdit() string {
	// Handle help overlay
	if m.editorHelpVisible {
		return m.viewEditorHelp()
	}

	var b strings.Builder

	if m.detailItem == nil {
		b.WriteString("No item selected")
		return appStyle.Render(b.String())
	}

	files := m.getDetailFiles()
	fileName := ""
	if m.detailFileIndex < len(files) {
		fileName = files[m.detailFileIndex]
	}

	// Header with file name and EDITING indicator
	lineInfo := fmt.Sprintf(" (Ln %d)", m.detailTextarea.Line()+1)
	header := headerStyle.Render("Editing: " + fileName + lineInfo)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Show goto line input if active
	if m.editorGotoMode {
		b.WriteString("Go to line: ")
		b.WriteString(m.editorGotoInput.View())
		b.WriteString("\n\n")
	}

	// Textarea for editing
	b.WriteString(m.detailTextarea.View())
	b.WriteString("\n\n")

	// Status message
	if m.statusMsg != "" {
		if m.isError {
			b.WriteString(errorStyle.Render("✗ " + m.statusMsg))
		} else {
			b.WriteString(successStyle.Render("✓ " + m.statusMsg))
		}
		b.WriteString("\n")
	}

	// Help - show navigation shortcuts
	help := m.renderHelp([]helpItem{
		{"ctrl+s", "save"},
		{"esc", "cancel"},
		{"?", "help"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewStaging renders the staging view with pending files
func (m Model) viewStaging() string {
	if m.stagingEditMode {
		return m.viewStagingEdit()
	}
	if m.stagingDetailMode {
		return m.viewStagingDetail()
	}
	return m.viewStagingList()
}

// viewStagingList renders the list of staging files
func (m Model) viewStagingList() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render(fmt.Sprintf("Import Staging (%d pending)", len(m.stagingItems)))
	b.WriteString(header)
	b.WriteString("\n\n")

	if len(m.stagingItems) == 0 {
		b.WriteString(subtleStyle.Render("No files pending in staging directory"))
		b.WriteString("\n")
		b.WriteString(subtleStyle.Render("Use Scan in Settings to import files from external paths"))
	} else {
		b.WriteString(sectionStyle.Render("Pending Files:"))
		b.WriteString("\n\n")

		for i, item := range m.stagingItems {
			isSelected := i == m.stagingCursor

			// File name
			fileName := item.Path

			// Suggestion with confidence
			suggestion := fmt.Sprintf("%s:%s", item.SuggestedType, item.SuggestedName)
			confidence := fmt.Sprintf("(%d%%)", item.Confidence)

			if isSelected {
				b.WriteString("  │ ")
				b.WriteString(selectedTextStyle.Render(fileName))
				b.WriteString("\n")
				b.WriteString("  │ ")
				b.WriteString(subtleStyle.Render("Suggested: "))
				b.WriteString(suggestion)
				b.WriteString(" ")
				b.WriteString(subtleStyle.Render(confidence))
				if item.Reason != "" {
					b.WriteString("\n")
					b.WriteString("  │ ")
					b.WriteString(subtleStyle.Render(item.Reason))
				}
				b.WriteString("\n")
			} else {
				b.WriteString("    ")
				b.WriteString(fileName)
				b.WriteString("\n")
				b.WriteString("    ")
				b.WriteString(subtleStyle.Render("Suggested: "+suggestion+" "+confidence))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Status message
	if m.statusMsg != "" {
		if m.isError {
			b.WriteString(errorStyle.Render("✗ " + m.statusMsg))
		} else {
			b.WriteString(successStyle.Render("✓ " + m.statusMsg))
		}
		b.WriteString("\n\n")
	}

	// Help
	help := m.renderHelp([]helpItem{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"enter", "view"},
		{"p", "process all"},
		{"d", "delete"},
		{"esc", "back"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewStagingDetail renders the content of a single staging file
func (m Model) viewStagingDetail() string {
	var b strings.Builder

	if m.stagingCursor >= len(m.stagingItems) {
		b.WriteString("No file selected")
		return appStyle.Render(b.String())
	}

	item := m.stagingItems[m.stagingCursor]

	// Header with file name
	header := headerStyle.Render("Staging: " + item.Path)
	b.WriteString(header)
	b.WriteString("\n\n")

	// File info
	suggestion := fmt.Sprintf("%s:%s", item.SuggestedType, item.SuggestedName)
	b.WriteString(fmt.Sprintf("  Suggested:  %s %s\n", titleStyle.Render(suggestion), subtleStyle.Render(fmt.Sprintf("(%d%% confidence)", item.Confidence))))
	if item.Reason != "" {
		b.WriteString(fmt.Sprintf("  Reason:     %s\n", subtleStyle.Render(item.Reason)))
	}
	b.WriteString("\n")

	// Separator line
	b.WriteString(subtleStyle.Render(strings.Repeat("─", 60)))
	b.WriteString("\n\n")

	// File content (rendered with glamour for .md files)
	if m.stagingLoading {
		b.WriteString(m.spinner.View() + " " + subtleStyle.Render("Loading file content..."))
	} else if m.stagingContent != "" {
		// Limit content height to leave room for other elements
		lines := strings.Split(m.stagingContent, "\n")
		maxLines := m.height - 15
		if maxLines < 10 {
			maxLines = 10
		}
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, subtleStyle.Render("... (content truncated)"))
		}
		b.WriteString(strings.Join(lines, "\n"))
	} else {
		b.WriteString(subtleStyle.Render("No content to display"))
	}
	b.WriteString("\n\n")

	// Status message
	if m.statusMsg != "" {
		if m.isError {
			b.WriteString(errorStyle.Render("✗ " + m.statusMsg))
		} else {
			b.WriteString(successStyle.Render("✓ " + m.statusMsg))
		}
		b.WriteString("\n")
	}

	// Help
	help := m.renderHelp([]helpItem{
		{"e", "edit"},
		{"d", "delete"},
		{"esc", "back to list"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewStagingEdit renders the inline editor for staging files
func (m Model) viewStagingEdit() string {
	// Handle help overlay
	if m.editorHelpVisible {
		return m.viewEditorHelp()
	}

	var b strings.Builder

	if m.stagingCursor >= len(m.stagingItems) {
		b.WriteString("No file selected")
		return appStyle.Render(b.String())
	}

	item := m.stagingItems[m.stagingCursor]

	// Header with file name and EDITING indicator
	lineInfo := fmt.Sprintf(" (Ln %d)", m.stagingTextarea.Line()+1)
	header := headerStyle.Render("Editing: " + item.Path + lineInfo)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Show goto line input if active
	if m.editorGotoMode {
		b.WriteString("Go to line: ")
		b.WriteString(m.editorGotoInput.View())
		b.WriteString("\n\n")
	}

	// Textarea for editing
	b.WriteString(m.stagingTextarea.View())
	b.WriteString("\n\n")

	// Status message
	if m.statusMsg != "" {
		if m.isError {
			b.WriteString(errorStyle.Render("✗ " + m.statusMsg))
		} else {
			b.WriteString(successStyle.Render("✓ " + m.statusMsg))
		}
		b.WriteString("\n")
	}

	// Help - show navigation shortcuts
	help := m.renderHelp([]helpItem{
		{"ctrl+s", "save"},
		{"esc", "cancel"},
		{"?", "help"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewEditorHelp renders the editor shortcuts help overlay
func (m Model) viewEditorHelp() string {
	var b strings.Builder

	header := headerStyle.Render("Editor Shortcuts")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Shortcuts organized by category
	shortcuts := []struct {
		category string
		items    []struct{ key, desc string }
	}{
		{
			category: "Navigation",
			items: []struct{ key, desc string }{
				{"↑/↓/←/→", "Move cursor"},
				{"Ctrl+←/→", "Move by word"},
				{"Home/Ctrl+A", "Line start"},
				{"End/Ctrl+E", "Line end"},
				{"Ctrl+Home", "File start"},
				{"Ctrl+End", "File end"},
				{"PgUp/PgDn", "Page up/down"},
				{"Ctrl+G", "Go to line"},
			},
		},
		{
			category: "Editing",
			items: []struct{ key, desc string }{
				{"Enter", "New line"},
				{"Backspace", "Delete char"},
				{"Ctrl+D", "Duplicate line"},
				{"Ctrl+J", "Join lines"},
				{"Ctrl+/", "Toggle comment"},
				{"Ctrl+]", "Indent"},
				{"Ctrl+[", "Unindent"},
			},
		},
		{
			category: "Delete",
			items: []struct{ key, desc string }{
				{"Ctrl+K", "Delete to end of line"},
				{"Ctrl+U", "Delete to start of line"},
				{"Alt+Backspace", "Delete word backward"},
				{"Alt+Delete", "Delete word forward"},
			},
		},
		{
			category: "Case",
			items: []struct{ key, desc string }{
				{"Alt+U", "Uppercase word"},
				{"Alt+L", "Lowercase word"},
				{"Alt+C", "Capitalize word"},
			},
		},
		{
			category: "File",
			items: []struct{ key, desc string }{
				{"Ctrl+S", "Save"},
				{"Esc", "Cancel/Exit"},
				{"Ctrl+V", "Paste"},
			},
		},
	}

	for _, cat := range shortcuts {
		b.WriteString(sectionStyle.Render(cat.category + ":"))
		b.WriteString("\n")
		for _, item := range cat.items {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				helpKeyStyle.Render(fmt.Sprintf("%-15s", item.key)),
				subtleStyle.Render(item.desc)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	help := m.renderHelp([]helpItem{
		{"any key", "close"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// Helper types and functions

type helpItem struct {
	key  string
	desc string
}

func (m Model) renderHelp(items []helpItem) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, helpKeyStyle.Render(item.key)+" "+helpStyle.Render(item.desc))
	}
	return statusBarStyle.Render(strings.Join(parts, "  "))
}

func (m Model) renderCard(title, value string, active bool) string {
	style := cardStyle
	if active {
		style = cardActiveStyle
	}

	content := cardTitleStyle.Render(title) + "\n" + cardValueStyle.Render(value)
	return style.Render(content)
}
