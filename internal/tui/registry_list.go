package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/okto-digital/regis3/internal/registry"
)

// registryItem wraps a registry.Item for the list component
type registryItem struct {
	item     *registry.Item
	selected bool
}

func (i registryItem) Title() string {
	return i.item.FullName()
}

func (i registryItem) Description() string {
	return i.item.Desc
}

func (i registryItem) FilterValue() string {
	// Allow filtering by name, type, description, and tags
	parts := []string{i.item.Name, i.item.Type, i.item.Desc}
	parts = append(parts, i.item.Tags...)
	return strings.Join(parts, " ")
}

// itemDelegate handles rendering of list items
type itemDelegate struct {
	selected map[string]bool
}

func (d itemDelegate) Height() int  { return 1 }
func (d itemDelegate) Spacing() int { return 0 }

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(registryItem)
	if !ok {
		return
	}

	ref := i.item.FullName()
	isSelected := d.selected[ref]
	isCursor := index == m.Index()

	// Selection marker
	selectMark := "  "
	if isSelected {
		selectMark = "● "
	}

	// Cursor marker
	cursor := "  "
	if isCursor {
		cursor = "> "
	}

	// Build the line
	name := ref
	if len(name) > 28 {
		name = name[:25] + "..."
	}

	desc := i.item.Desc
	if len(desc) > 40 {
		desc = desc[:37] + "..."
	}

	line := fmt.Sprintf("%s%s%-28s  %s", cursor, selectMark, name, desc)

	// Style based on cursor position
	style := itemStyle
	if isCursor {
		style = selectedItemStyle
	}

	fmt.Fprint(w, style.Render(line))
}

// newRegistryList creates a new list.Model for registry items
func newRegistryList(items []*registry.Item, selected map[string]bool, width, height int) list.Model {
	// Convert to list items
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = registryItem{
			item:     item,
			selected: selected[item.FullName()],
		}
	}

	// Create delegate with selection tracking
	delegate := itemDelegate{selected: selected}

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

// updateRegistryListItems updates the list with new items
func updateRegistryListItems(l *list.Model, items []*registry.Item, selected map[string]bool) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = registryItem{
			item:     item,
			selected: selected[item.FullName()],
		}
	}
	l.SetItems(listItems)

	// Update delegate with new selection state
	l.SetDelegate(itemDelegate{selected: selected})
}
