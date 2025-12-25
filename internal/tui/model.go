package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/okto-digital/regis3/internal/config"
	"github.com/okto-digital/regis3/internal/registry"
)

// View represents the current view in the TUI
type View int

const (
	ViewDashboard View = iota
	ViewRegistry
	ViewProject
	ViewSettings
	ViewDetail
)

// Model is the main application state
type Model struct {
	// Current view
	currentView View
	previousView View

	// Data
	config       *config.Config
	manifest     *registry.Manifest
	registryPath string
	projectPath  string

	// UI state
	width  int
	height int
	ready  bool

	// Selected items (for multi-select)
	selected map[string]bool

	// Current cursor position in lists
	cursor int

	// Detail view item
	detailItem *registry.Item

	// Error/status message
	statusMsg string
	isError   bool

	// Key bindings
	keys keyMap
}

// keyMap defines all keyboard shortcuts
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Select   key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Help     key.Binding
	Search   key.Binding
	Add      key.Binding
	Remove   key.Binding
	Update   key.Binding
	Build    key.Binding
	Registry key.Binding
	Project  key.Binding
	Settings key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Remove: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove"),
		),
		Update: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "update"),
		),
		Build: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "build"),
		),
		Registry: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "registry"),
		),
		Project: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "project"),
		),
		Settings: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "settings"),
		),
	}
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config, registryPath, projectPath string) Model {
	return Model{
		currentView:  ViewDashboard,
		config:       cfg,
		registryPath: registryPath,
		projectPath:  projectPath,
		selected:     make(map[string]bool),
		keys:         defaultKeyMap(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		loadManifest(m.registryPath),
	)
}

// loadManifest loads the registry manifest
func loadManifest(registryPath string) tea.Cmd {
	return func() tea.Msg {
		manifest, err := registry.LoadManifestFromRegistry(registryPath)
		if err != nil {
			// Try building first
			result, buildErr := registry.BuildRegistry(registryPath)
			if buildErr != nil {
				return manifestLoadedMsg{err: err}
			}
			manifest = result.Manifest
		}
		return manifestLoadedMsg{manifest: manifest}
	}
}

// manifestLoadedMsg is sent when manifest loading completes
type manifestLoadedMsg struct {
	manifest *registry.Manifest
	err      error
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case manifestLoadedMsg:
		if msg.err != nil {
			m.statusMsg = "Failed to load registry: " + msg.err.Error()
			m.isError = true
		} else {
			m.manifest = msg.manifest
			m.statusMsg = ""
			m.isError = false
		}
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Registry):
		m.currentView = ViewRegistry
		m.cursor = 0
		return m, nil

	case key.Matches(msg, m.keys.Project):
		m.currentView = ViewProject
		m.cursor = 0
		return m, nil

	case key.Matches(msg, m.keys.Settings):
		m.currentView = ViewSettings
		m.cursor = 0
		return m, nil

	case key.Matches(msg, m.keys.Build):
		return m, m.buildRegistry()

	case key.Matches(msg, m.keys.Back):
		if m.currentView == ViewDashboard {
			return m, tea.Quit
		}
		if m.currentView == ViewDetail {
			m.currentView = m.previousView
		} else {
			m.currentView = ViewDashboard
		}
		m.cursor = 0
		return m, nil
	}

	// View-specific keys
	switch m.currentView {
	case ViewDashboard:
		return m.updateDashboard(msg)
	case ViewRegistry:
		return m.updateRegistry(msg)
	case ViewProject:
		return m.updateProject(msg)
	case ViewSettings:
		return m.updateSettings(msg)
	case ViewDetail:
		return m.updateDetail(msg)
	}

	return m, nil
}

// View-specific update handlers (to be implemented in view files)
func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) updateRegistry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.getRegistryItems()

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Select):
		if m.cursor < len(items) {
			item := items[m.cursor]
			ref := item.FullName()
			if m.selected[ref] {
				delete(m.selected, ref)
			} else {
				m.selected[ref] = true
			}
		}
	case key.Matches(msg, m.keys.Enter):
		if m.cursor < len(items) {
			m.detailItem = items[m.cursor]
			m.previousView = ViewRegistry
			m.currentView = ViewDetail
		}
	}
	return m, nil
}

func (m Model) updateProject(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

// buildRegistry triggers a registry build
func (m Model) buildRegistry() tea.Cmd {
	return func() tea.Msg {
		_, err := registry.BuildRegistry(m.registryPath)
		if err != nil {
			return buildCompleteMsg{err: err}
		}
		return buildCompleteMsg{}
	}
}

type buildCompleteMsg struct {
	err error
}

// getRegistryItems returns sorted registry items
func (m Model) getRegistryItems() []*registry.Item {
	if m.manifest == nil {
		return nil
	}

	var items []*registry.Item
	for _, item := range m.manifest.Items {
		items = append(items, item)
	}
	return items
}

// View renders the current view
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	switch m.currentView {
	case ViewDashboard:
		return m.viewDashboard()
	case ViewRegistry:
		return m.viewRegistry()
	case ViewProject:
		return m.viewProject()
	case ViewSettings:
		return m.viewSettings()
	case ViewDetail:
		return m.viewDetail()
	}

	return ""
}
