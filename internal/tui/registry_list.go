package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/okto-digital/regis3/internal/registry"
)

// registryItem wraps a registry.Item for the list component
type registryItem struct {
	item      *registry.Item
	selected  bool
	installed bool
}

func (i registryItem) Title() string {
	prefix := ""
	if i.installed {
		prefix = "✓ "
	}
	return prefix + i.item.FullName()
}

func (i registryItem) IsSelected() bool {
	return i.selected
}

func (i registryItem) Description() string {
	prefix := ""
	if i.selected {
		prefix = "● "
	}
	return prefix + i.item.Desc
}

func (i registryItem) FilterValue() string {
	// Use full name for filtering - this matches what Title() displays
	// so the fuzzy match highlighting works correctly
	return i.item.FullName()
}

// newRegistryList creates a new list.Model for registry items
func newRegistryList(items []*registry.Item, selected map[string]bool, installed map[string]bool, width, height int) list.Model {
	// Convert to list items
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		ref := item.FullName()
		listItems[i] = registryItem{
			item:      item,
			selected:  selected[ref],
			installed: installed[ref],
		}
	}

	// Use the default delegate with custom styles
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(primaryColor).
		BorderLeftForeground(primaryColor)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#666666"))

	// Create the list
	l := list.New(listItems, delegate, width, height)
	l.Title = "Registry"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Style the list
	l.Styles.Title = titleStyle
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(primaryColor)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(primaryColor)

	// Customize help
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("space"),
				key.WithHelp("space", "select"),
			),
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "details"),
			),
			key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add selected"),
			),
		}
	}

	return l
}

// updateRegistryListItems updates the list with new items and selection state
func updateRegistryListItems(l *list.Model, items []*registry.Item, selected map[string]bool, installed map[string]bool) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		ref := item.FullName()
		listItems[i] = registryItem{
			item:      item,
			selected:  selected[ref],
			installed: installed[ref],
		}
	}
	l.SetItems(listItems)
}
