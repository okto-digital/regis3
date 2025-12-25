package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewDashboard renders the dashboard view with stats cards
func (m Model) viewDashboard() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("regis3 TUI")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Stats cards
	registryCount := 0
	if m.manifest != nil {
		registryCount = len(m.manifest.Items)
	}

	// Create three cards side by side
	registryCard := m.renderCard("Registry", fmt.Sprintf("%d items", registryCount), m.cursor == 0)
	projectCard := m.renderCard("Project", "View installed", m.cursor == 1)
	settingsCard := m.renderCard("Settings", "Configure", m.cursor == 2)

	cards := lipgloss.JoinHorizontal(lipgloss.Top, registryCard, "  ", projectCard, "  ", settingsCard)
	b.WriteString(cards)
	b.WriteString("\n\n")

	// Info section
	b.WriteString(subtleStyle.Render("Registry: "))
	b.WriteString(m.registryPath)
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("Project:  "))
	b.WriteString(m.projectPath)
	b.WriteString("\n\n")

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
	help := m.renderHelp([]helpItem{
		{"1", "Registry"},
		{"2", "Project"},
		{"3", "Settings"},
		{"b", "Build"},
		{"q", "Quit"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewRegistry renders the registry browser view
func (m Model) viewRegistry() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("Registry Browser")
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.manifest == nil {
		b.WriteString(warningStyle.Render("No registry loaded. Press 'b' to build."))
		b.WriteString("\n\n")
	} else {
		items := m.getRegistryItems()
		if len(items) == 0 {
			b.WriteString(subtleStyle.Render("No items in registry"))
			b.WriteString("\n")
		} else {
			// Group items by type
			grouped := make(map[string][]int)
			for i, item := range items {
				grouped[item.Type] = append(grouped[item.Type], i)
			}

			// Render groups
			typeOrder := []string{"skill", "subagent", "command", "doc", "prompt", "philosophy", "project", "ruleset", "mcp", "script", "hook", "stack"}
			for _, itemType := range typeOrder {
				indices, ok := grouped[itemType]
				if !ok || len(indices) == 0 {
					continue
				}

				// Group header
				b.WriteString(sectionStyle.Render(fmt.Sprintf("── %ss (%d) ──", strings.Title(itemType), len(indices))))
				b.WriteString("\n")

				// Items in group
				for _, idx := range indices {
					item := items[idx]
					ref := item.FullName()

					prefix := "  "
					style := itemStyle
					if idx == m.cursor {
						prefix = "> "
						style = selectedItemStyle
					}

					// Selection indicator
					selectMark := " "
					if m.selected[ref] {
						selectMark = "●"
					}

					line := fmt.Sprintf("%s%s %-30s", prefix, selectMark, ref)
					if item.Desc != "" {
						desc := item.Desc
						if len(desc) > 40 {
							desc = desc[:37] + "..."
						}
						line += "  " + subtleStyle.Render(desc)
					}
					b.WriteString(style.Render(line))
					b.WriteString("\n")
				}
				b.WriteString("\n")
			}
		}
	}

	// Status
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
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"space", "select"},
		{"enter", "details"},
		{"a", "add"},
		{"esc", "back"},
	})
	b.WriteString("\n")
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewProject renders the project view with installed items
func (m Model) viewProject() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("Project: " + m.projectPath)
	b.WriteString(header)
	b.WriteString("\n\n")

	b.WriteString(subtleStyle.Render("Installed items will appear here"))
	b.WriteString("\n\n")

	// Help
	help := m.renderHelp([]helpItem{
		{"r", "remove"},
		{"u", "update"},
		{"a", "add more"},
		{"esc", "back"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewSettings renders the settings view
func (m Model) viewSettings() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("Settings")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Settings display
	b.WriteString(fmt.Sprintf("  Registry Path:  %s\n", m.registryPath))
	b.WriteString(fmt.Sprintf("  Project Path:   %s\n", m.projectPath))
	b.WriteString("\n")

	// Help
	help := m.renderHelp([]helpItem{
		{"esc", "back"},
	})
	b.WriteString(help)

	return appStyle.Render(b.String())
}

// viewDetail renders the item detail view
func (m Model) viewDetail() string {
	var b strings.Builder

	if m.detailItem == nil {
		b.WriteString("No item selected")
		return appStyle.Render(b.String())
	}

	item := m.detailItem

	// Header
	header := headerStyle.Render(item.FullName())
	b.WriteString(header)
	b.WriteString("\n\n")

	// Item info
	b.WriteString(fmt.Sprintf("  Type:   %s\n", titleStyle.Render(item.Type)))
	if len(item.Tags) > 0 {
		b.WriteString(fmt.Sprintf("  Tags:   %s\n", strings.Join(item.Tags, ", ")))
	}
	if item.Status != "" {
		b.WriteString(fmt.Sprintf("  Status: %s\n", item.Status))
	}
	b.WriteString("\n")

	// Description
	if item.Desc != "" {
		b.WriteString(sectionStyle.Render("Description:"))
		b.WriteString("\n")
		b.WriteString("  " + item.Desc)
		b.WriteString("\n\n")
	}

	// Dependencies
	if len(item.Deps) > 0 {
		b.WriteString(sectionStyle.Render("Dependencies:"))
		b.WriteString("\n")
		for _, dep := range item.Deps {
			b.WriteString(fmt.Sprintf("  - %s\n", dep))
		}
		b.WriteString("\n")
	}

	// Files
	if len(item.Files) > 0 {
		b.WriteString(sectionStyle.Render("Files:"))
		b.WriteString("\n")
		for _, f := range item.Files {
			b.WriteString(fmt.Sprintf("  - %s\n", f))
		}
		b.WriteString("\n")
	}

	// Help
	help := m.renderHelp([]helpItem{
		{"a", "add to project"},
		{"esc", "back"},
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
