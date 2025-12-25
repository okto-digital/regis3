package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/okto-digital/regis3/internal/config"
	"github.com/okto-digital/regis3/internal/importer"
	"github.com/okto-digital/regis3/internal/installer"
	"github.com/okto-digital/regis3/internal/registry"
)

// settingsItem represents an item in the settings view
type settingsItem struct {
	label     string // Display label
	value     string // Current value (for paths) or description (for actions)
	isPath    bool   // True if editable path
	actionKey string // Action identifier (build, validate, etc.)
}

// View represents the current view in the TUI
type View int

const (
	ViewDashboard View = iota
	ViewRegistry
	ViewProject
	ViewSettings
	ViewDetail
	ViewStaging
)

// Version info
var (
	Version = "dev"
	Author  = "okto-digital"
)

// Model is the main application state
type Model struct {
	// Current view
	currentView  View
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

	// Installed items cache
	installed map[string]bool

	// Current cursor position (for dashboard cards)
	cursor int

	// Project view cursor
	projectCursor int

	// Registry list component
	registryList list.Model

	// Detail view item
	detailItem      *registry.Item
	detailFileIndex int    // Selected file in detail view (0 = main source)
	detailContent   string // Content of currently selected file (rendered)
	detailRawContent string // Raw content for editing
	detailEditMode  bool   // True when editing file content
	detailTextarea  textarea.Model // Textarea for inline editing
	detailLoading   bool   // True when loading file content

	// Staging view state
	stagingItems      []importer.PendingFile
	stagingCursor     int
	stagingContent    string         // Content of currently selected staging file (rendered)
	stagingRawContent string         // Raw content for editing
	stagingDetailMode bool           // True when viewing file content, false when in list
	stagingEditMode   bool           // True when editing file content
	stagingTextarea   textarea.Model // Textarea for inline editing
	stagingLoading    bool           // True when loading file content

	// Loading spinner
	spinner spinner.Model

	// Editor state (for detail and staging editors)
	editorHelpVisible bool           // Show shortcuts help overlay
	editorGotoMode    bool           // Go to line input mode
	editorGotoInput   textinput.Model // Text input for line number

	// Error/status message
	statusMsg string
	isError   bool

	// Settings view state
	settingsCursor  int              // Current selection in settings
	settingsEditing bool             // Whether we're editing a path
	settingsScanMode bool            // Whether we're entering a scan path
	settingsInput   textinput.Model  // Text input for path editing
	settingsItems   []settingsItem   // List of settings/actions

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
	Edit     key.Binding
	Save     key.Binding
	Delete   key.Binding
	Process  key.Binding
	Registry key.Binding
	Project  key.Binding
	Settings key.Binding
	Staging  key.Binding
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
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Process: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "process"),
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
		Staging: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "staging"),
		),
	}
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config, registryPath, projectPath string) Model {
	// Initialize with empty list to avoid nil pointer issues
	emptyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)

	// Initialize text input for path editing
	ti := textinput.New()
	ti.Placeholder = "Enter path..."
	ti.CharLimit = 256
	ti.Width = 50

	// Initialize textarea for staging file editing
	ta := textarea.New()
	ta.Placeholder = "File content..."
	ta.ShowLineNumbers = true
	ta.SetWidth(80)
	ta.SetHeight(20)
	ta.CharLimit = 0 // No limit

	// Initialize textarea for detail file editing
	detailTa := textarea.New()
	detailTa.Placeholder = "File content..."
	detailTa.ShowLineNumbers = true
	detailTa.SetWidth(80)
	detailTa.SetHeight(20)
	detailTa.CharLimit = 0 // No limit

	// Initialize loading spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	// Initialize goto line input
	gotoInput := textinput.New()
	gotoInput.Placeholder = "Line number..."
	gotoInput.CharLimit = 10
	gotoInput.Width = 20

	// Initialize settings items
	settingsItems := []settingsItem{
		// Configuration (editable paths)
		{label: "Registry Path", value: registryPath, isPath: true},
		{label: "Project Path", value: projectPath, isPath: true},
		// Actions
		{label: "Build", value: "Rebuild registry manifest", actionKey: "build"},
		{label: "Validate", value: "Check registry integrity", actionKey: "validate"},
		{label: "Scan", value: "Import from external path", actionKey: "scan"},
		{label: "Find Orphans", value: "Find unreferenced files", actionKey: "orphans"},
		{label: "Reindex", value: "Reindex all items", actionKey: "reindex"},
	}

	return Model{
		currentView:     ViewDashboard,
		config:          cfg,
		registryPath:    registryPath,
		projectPath:     projectPath,
		selected:        make(map[string]bool),
		installed:       make(map[string]bool),
		registryList:    emptyList,
		settingsInput:   ti,
		settingsItems:   settingsItems,
		stagingTextarea: ta,
		detailTextarea:  detailTa,
		spinner:         s,
		editorGotoInput: gotoInput,
		keys:            defaultKeyMap(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		loadManifest(m.registryPath),
		loadStagingItems(m.registryPath),
	)
}

// loadStagingItems loads pending files from the staging directory
func loadStagingItems(registryPath string) tea.Cmd {
	return func() tea.Msg {
		imp := importer.NewImporter(registryPath)
		pending, err := imp.ListPending()
		if err != nil {
			return stagingLoadedMsg{err: err}
		}
		return stagingLoadedMsg{items: pending}
	}
}

// stagingLoadedMsg is sent when staging items are loaded
type stagingLoadedMsg struct {
	items []importer.PendingFile
	err   error
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
	var cmds []tea.Cmd

	// Handle spinner tick message - must return immediately to keep animation going
	if tickMsg, ok := msg.(spinner.TickMsg); ok {
		if m.detailLoading || m.stagingLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(tickMsg)
			return m, cmd
		}
		// Not loading, ignore the tick
		return m, nil
	}

	// Handle edit mode - forward ALL messages to textarea (not just key messages)
	if m.detailEditMode {
		// Handle help overlay
		if m.editorHelpVisible {
			if _, ok := msg.(tea.KeyMsg); ok {
				// Any key closes help
				m.editorHelpVisible = false
				return m, nil
			}
			return m, nil
		}

		// Handle goto line mode
		if m.editorGotoMode {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.String() {
				case "enter":
					// Parse line number and go to it
					if lineNum, err := strconv.Atoi(m.editorGotoInput.Value()); err == nil {
						editorGotoLine(&m.detailTextarea, lineNum)
					}
					m.editorGotoMode = false
					m.editorGotoInput.Blur()
					return m, nil
				case "esc":
					m.editorGotoMode = false
					m.editorGotoInput.Blur()
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.editorGotoInput, cmd = m.editorGotoInput.Update(msg)
			return m, cmd
		}

		// Check for special keys first
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+s":
				return m, m.saveDetailFile()
			case "esc":
				m.detailEditMode = false
				m.detailTextarea.Blur()
				return m, nil
			case "pgup":
				// Move cursor up by page (textarea height)
				pageSize := m.detailTextarea.Height()
				for i := 0; i < pageSize; i++ {
					m.detailTextarea.CursorUp()
				}
				return m, nil
			case "pgdown":
				// Move cursor down by page (textarea height)
				pageSize := m.detailTextarea.Height()
				for i := 0; i < pageSize; i++ {
					m.detailTextarea.CursorDown()
				}
				return m, nil
			case "enter":
				// Explicitly handle Enter to insert newline
				m.detailTextarea.InsertRune('\n')
				return m, nil
			case "?":
				// Show help overlay
				m.editorHelpVisible = true
				return m, nil
			case "ctrl+g":
				// Go to line
				m.editorGotoMode = true
				m.editorGotoInput.SetValue("")
				m.editorGotoInput.Focus()
				return m, nil
			case "ctrl+d":
				// Duplicate line
				editorDuplicateLine(&m.detailTextarea)
				return m, nil
			case "ctrl+j":
				// Join lines
				editorJoinLines(&m.detailTextarea)
				return m, nil
			case "ctrl+/":
				// Toggle comment
				editorToggleComment(&m.detailTextarea)
				return m, nil
			case "ctrl+]":
				// Indent
				editorIndent(&m.detailTextarea)
				return m, nil
			case "ctrl+[":
				// Unindent
				editorUnindent(&m.detailTextarea)
				return m, nil
			}
		}
		// Forward all messages to textarea
		var cmd tea.Cmd
		m.detailTextarea, cmd = m.detailTextarea.Update(msg)
		return m, cmd
	}

	if m.stagingEditMode {
		// Handle help overlay
		if m.editorHelpVisible {
			if _, ok := msg.(tea.KeyMsg); ok {
				// Any key closes help
				m.editorHelpVisible = false
				return m, nil
			}
			return m, nil
		}

		// Handle goto line mode
		if m.editorGotoMode {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.String() {
				case "enter":
					// Parse line number and go to it
					if lineNum, err := strconv.Atoi(m.editorGotoInput.Value()); err == nil {
						editorGotoLine(&m.stagingTextarea, lineNum)
					}
					m.editorGotoMode = false
					m.editorGotoInput.Blur()
					return m, nil
				case "esc":
					m.editorGotoMode = false
					m.editorGotoInput.Blur()
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.editorGotoInput, cmd = m.editorGotoInput.Update(msg)
			return m, cmd
		}

		// Check for special keys first
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+s":
				return m, m.saveStagingFile()
			case "esc":
				m.stagingEditMode = false
				m.stagingTextarea.Blur()
				return m, nil
			case "pgup":
				// Move cursor up by page (textarea height)
				pageSize := m.stagingTextarea.Height()
				for i := 0; i < pageSize; i++ {
					m.stagingTextarea.CursorUp()
				}
				return m, nil
			case "pgdown":
				// Move cursor down by page (textarea height)
				pageSize := m.stagingTextarea.Height()
				for i := 0; i < pageSize; i++ {
					m.stagingTextarea.CursorDown()
				}
				return m, nil
			case "enter":
				// Explicitly handle Enter to insert newline
				m.stagingTextarea.InsertRune('\n')
				return m, nil
			case "?":
				// Show help overlay
				m.editorHelpVisible = true
				return m, nil
			case "ctrl+g":
				// Go to line
				m.editorGotoMode = true
				m.editorGotoInput.SetValue("")
				m.editorGotoInput.Focus()
				return m, nil
			case "ctrl+d":
				// Duplicate line
				editorDuplicateLine(&m.stagingTextarea)
				return m, nil
			case "ctrl+j":
				// Join lines
				editorJoinLines(&m.stagingTextarea)
				return m, nil
			case "ctrl+/":
				// Toggle comment
				editorToggleComment(&m.stagingTextarea)
				return m, nil
			case "ctrl+]":
				// Indent
				editorIndent(&m.stagingTextarea)
				return m, nil
			case "ctrl+[":
				// Unindent
				editorUnindent(&m.stagingTextarea)
				return m, nil
			}
		}
		// Forward all messages to textarea
		var cmd tea.Cmd
		m.stagingTextarea, cmd = m.stagingTextarea.Update(msg)
		return m, cmd
	}

	// Handle common messages first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.registryList.SetSize(msg.Width-4, msg.Height-8)

	case manifestLoadedMsg:
		if msg.err != nil {
			m.statusMsg = "Failed to load registry: " + msg.err.Error()
			m.isError = true
		} else {
			m.manifest = msg.manifest
			m.statusMsg = "Registry loaded"
			m.isError = false
			// Refresh installed items cache
			m.refreshInstalledCache()
			items := m.getRegistryItems()
			m.registryList = newRegistryList(items, m.selected, m.installed, m.width-4, m.height-8)
		}
		return m, nil

	case stagingLoadedMsg:
		if msg.err == nil {
			m.stagingItems = msg.items
			// Reset detail mode if items changed
			if m.stagingCursor >= len(m.stagingItems) {
				m.stagingCursor = 0
				m.stagingDetailMode = false
				m.stagingContent = ""
			}
		}
		return m, nil

	case fileEditedMsg:
		if msg.err != nil {
			m.statusMsg = "Edit failed: " + msg.err.Error()
			m.isError = true
		} else {
			m.statusMsg = "File saved"
			m.isError = false
			// Reload file content
			if m.detailItem != nil {
				m.detailLoading = true
				return m, tea.Batch(m.spinner.Tick, m.loadDetailFileContentCmd())
			}
		}
		return m, nil

	case stagingFileEditedMsg:
		if msg.err != nil {
			m.statusMsg = "Edit failed: " + msg.err.Error()
			m.isError = true
		} else {
			m.statusMsg = "File saved"
			m.isError = false
		}
		// Reload staging items after editing
		return m, loadStagingItems(m.registryPath)

	case stagingFileDeletedMsg:
		m.statusMsg = fmt.Sprintf("Deleted %s", msg.path)
		m.isError = false
		// Reload staging items after deletion
		return m, loadStagingItems(m.registryPath)

	case stagingFileSavedMsg:
		if msg.err != nil {
			m.statusMsg = "Save failed: " + msg.err.Error()
			m.isError = true
		} else {
			m.statusMsg = "File saved"
			m.isError = false
			// Exit edit mode and reload content
			m.stagingEditMode = false
			m.stagingTextarea.Blur()
			m.stagingLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadStagingFileContentCmd())
		}
		return m, nil

	case detailFileLoadedMsg:
		m.detailLoading = false
		m.detailContent = msg.content
		m.detailRawContent = msg.rawContent
		if msg.err != nil {
			m.statusMsg = "Load failed: " + msg.err.Error()
			m.isError = true
		}
		return m, nil

	case detailFileSavedMsg:
		if msg.err != nil {
			m.statusMsg = "Save failed: " + msg.err.Error()
			m.isError = true
		} else {
			m.statusMsg = "File saved"
			m.isError = false
			// Exit edit mode and reload content
			m.detailEditMode = false
			m.detailTextarea.Blur()
			m.detailLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadDetailFileContentCmd())
		}
		return m, nil

	case stagingFileLoadedMsg:
		m.stagingLoading = false
		m.stagingContent = msg.content
		m.stagingRawContent = msg.rawContent
		if msg.err != nil {
			m.statusMsg = "Load failed: " + msg.err.Error()
			m.isError = true
		}
		return m, nil

	case buildCompleteMsg:
		if msg.err != nil {
			m.statusMsg = "Build failed: " + msg.err.Error()
			m.isError = true
			return m, nil
		}
		m.statusMsg = "Build complete!"
		m.isError = false
		return m, loadManifest(m.registryPath)

	case statusMsg:
		m.statusMsg = msg.msg
		m.isError = msg.isError
		return m, nil

	case installCompleteMsg:
		if msg.installed > 0 {
			m.statusMsg = fmt.Sprintf("Installed %d item(s)", msg.installed)
			if msg.skipped > 0 {
				m.statusMsg += fmt.Sprintf(" (%d skipped)", msg.skipped)
			}
			// Clear selection after successful install
			m.selected = make(map[string]bool)
			// Refresh installed cache and list
			m.refreshInstalledCache()
			items := m.getRegistryItems()
			updateRegistryListItems(&m.registryList, items, m.selected, m.installed)
		} else {
			m.statusMsg = "No items installed"
		}
		m.isError = false
		return m, nil

	case removeCompleteMsg:
		m.statusMsg = fmt.Sprintf("Removed %s", msg.ref)
		m.isError = false
		// Refresh installed cache
		m.refreshInstalledCache()
		// Update registry list to reflect removal
		items := m.getRegistryItems()
		updateRegistryListItems(&m.registryList, items, m.selected, m.installed)
		// Reset cursor if needed
		installedItems := m.getInstalledItems()
		if m.projectCursor >= len(installedItems) && m.projectCursor > 0 {
			m.projectCursor--
		}
		return m, nil

	case actionCompleteMsg:
		m.statusMsg = msg.message
		m.isError = !msg.success
		// If action produced details, append first few to message
		if len(msg.details) > 0 {
			detailCount := len(msg.details)
			if detailCount > 3 {
				m.statusMsg += fmt.Sprintf(" (showing 3 of %d)", detailCount)
				detailCount = 3
			}
			for i := 0; i < detailCount; i++ {
				m.statusMsg += "\n  " + msg.details[i]
			}
		}
		// If it was build/reindex, reload manifest
		if msg.action == "build" || msg.action == "reindex" {
			return m, loadManifest(m.registryPath)
		}
		// If it was scan/process, reload staging
		if msg.action == "scan" || msg.action == "process" {
			return m, loadStagingItems(m.registryPath)
		}
		return m, nil
	}

	// When in registry view, pass messages to the list
	if m.currentView == ViewRegistry {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// When filtering, pass esc to list to cancel filter
			// Otherwise, don't pass back/quit keys (we handle navigation)
			isFiltering := m.registryList.FilterState() == list.Filtering
			shouldPassToList := true

			if key.Matches(keyMsg, m.keys.Quit) {
				shouldPassToList = false
			} else if key.Matches(keyMsg, m.keys.Back) && !isFiltering {
				shouldPassToList = false
			}

			if shouldPassToList {
				var cmd tea.Cmd
				m.registryList, cmd = m.registryList.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		} else {
			// Non-key messages always go to the list
			var cmd tea.Cmd
			m.registryList, cmd = m.registryList.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		var cmd tea.Cmd
		var newModel tea.Model
		newModel, cmd = m.handleKeyPress(keyMsg)
		if model, ok := newModel.(Model); ok {
			m = model
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
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

	case key.Matches(msg, m.keys.Staging):
		// Only navigate to staging if there are staging items
		if len(m.stagingItems) > 0 {
			m.currentView = ViewStaging
			m.stagingCursor = 0
			m.stagingDetailMode = false
			m.stagingContent = ""
		}
		return m, nil

	case key.Matches(msg, m.keys.Back):
		// Don't intercept esc when registry list is filtering
		if m.currentView == ViewRegistry && m.registryList.FilterState() == list.Filtering {
			// Let it pass through to the list
			break
		}
		// Only quit from dashboard
		if m.currentView == ViewDashboard {
			return m, tea.Quit
		}
		// Go back to previous view or dashboard
		if m.currentView == ViewDetail {
			m.currentView = m.previousView
		} else {
			m.currentView = ViewDashboard
		}
		m.cursor = 0
		m.statusMsg = "" // Clear status message when navigating
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
	case ViewStaging:
		return m.updateStaging(msg)
	}

	return m, nil
}

// View-specific update handlers (to be implemented in view files)
func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) updateRegistry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If the list is filtering, don't handle our custom keys (except quit)
	if m.registryList.FilterState() == list.Filtering {
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		// Let the list handle all other keys (already updated in Update())
		return m, nil
	}

	// Handle our custom keys when not filtering
	switch {
	case key.Matches(msg, m.keys.Select):
		// Toggle selection on current item
		if item, ok := m.registryList.SelectedItem().(registryItem); ok {
			ref := item.item.FullName()
			if m.selected[ref] {
				delete(m.selected, ref)
			} else {
				m.selected[ref] = true
			}
			// Update list items to reflect selection change
			items := m.getRegistryItems()
			updateRegistryListItems(&m.registryList, items, m.selected, m.installed)
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		// View item details
		if item, ok := m.registryList.SelectedItem().(registryItem); ok {
			m.detailItem = item.item
			m.detailFileIndex = 0 // Reset to first file
			m.previousView = ViewRegistry
			m.currentView = ViewDetail
			m.detailLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadDetailFileContentCmd())
		}
		return m, nil

	case key.Matches(msg, m.keys.Add):
		// Add selected items to project
		return m, m.addSelectedToProject()
	}

	// Other keys are handled by the list (already updated in Update())
	return m, nil
}

// addSelectedToProject adds all selected items to the current project
func (m Model) addSelectedToProject() tea.Cmd {
	if len(m.selected) == 0 {
		return func() tea.Msg {
			return statusMsg{msg: "No items selected. Use space to select items.", isError: true}
		}
	}

	// Collect selected item references
	var refs []string
	for ref := range m.selected {
		refs = append(refs, ref)
	}

	// Capture values for the closure
	manifest := m.manifest
	registryPath := m.registryPath
	projectPath := m.projectPath

	return func() tea.Msg {
		// Get target (default to claude)
		target := installer.DefaultClaudeTarget()

		// Create installer
		inst, err := installer.NewInstaller(projectPath, registryPath, target)
		if err != nil {
			return statusMsg{msg: fmt.Sprintf("Installer error: %s", err.Error()), isError: true}
		}

		// Install items
		result, err := inst.Install(manifest, refs)
		if err != nil {
			return statusMsg{msg: fmt.Sprintf("Install failed: %s", err.Error()), isError: true}
		}

		// Check for errors
		if len(result.Errors) > 0 {
			return statusMsg{
				msg:     fmt.Sprintf("Install failed: %s", result.Errors[0].Message),
				isError: true,
			}
		}

		// Success
		installed := len(result.Installed) + len(result.Updated)
		msg := fmt.Sprintf("Installed %d item(s) to project", installed)
		if len(result.Skipped) > 0 {
			msg += fmt.Sprintf(" (%d skipped)", len(result.Skipped))
		}

		return installCompleteMsg{
			installed: installed,
			skipped:   len(result.Skipped),
		}
	}
}

// installCompleteMsg is sent when installation completes
type installCompleteMsg struct {
	installed int
	skipped   int
}

// statusMsg is used to update the status message
type statusMsg struct {
	msg     string
	isError bool
}

func (m Model) updateProject(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.getInstalledItems()

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.projectCursor > 0 {
			m.projectCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.projectCursor < len(items)-1 {
			m.projectCursor++
		}
	case key.Matches(msg, m.keys.Remove):
		// Remove selected item
		if m.projectCursor < len(items) {
			return m, m.removeFromProject(items[m.projectCursor].ref)
		}
	}
	return m, nil
}

// removeFromProject removes an item from the current project
func (m Model) removeFromProject(ref string) tea.Cmd {
	projectPath := m.projectPath
	registryPath := m.registryPath
	manifest := m.manifest

	return func() tea.Msg {
		target := installer.DefaultClaudeTarget()

		inst, err := installer.NewInstaller(projectPath, registryPath, target)
		if err != nil {
			return statusMsg{msg: fmt.Sprintf("Error: %s", err.Error()), isError: true}
		}

		result, err := inst.Uninstall([]string{ref}, manifest)
		if err != nil {
			return statusMsg{msg: fmt.Sprintf("Remove failed: %s", err.Error()), isError: true}
		}

		if len(result.Errors) > 0 {
			return statusMsg{msg: fmt.Sprintf("Remove failed: %s", result.Errors[0].Message), isError: true}
		}

		if len(result.Uninstalled) == 0 {
			return statusMsg{msg: fmt.Sprintf("%s not found", ref), isError: true}
		}

		return removeCompleteMsg{ref: ref}
	}
}

// removeCompleteMsg is sent when removal completes
type removeCompleteMsg struct {
	ref string
}

// fileEditedMsg is sent when file editing completes
type fileEditedMsg struct {
	path string
	err  error
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle scan path input mode
	if m.settingsScanMode {
		switch msg.String() {
		case "enter":
			scanPath := m.settingsInput.Value()
			m.settingsScanMode = false
			m.settingsInput.Blur()
			if scanPath != "" {
				m.statusMsg = "Scanning " + scanPath + "..."
				m.isError = false
				return m, m.runScan(scanPath)
			}
			return m, nil
		case "esc":
			m.settingsScanMode = false
			m.settingsInput.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.settingsInput, cmd = m.settingsInput.Update(msg)
		return m, cmd
	}

	// Handle editing mode
	if m.settingsEditing {
		switch msg.String() {
		case "enter":
			// Save the edited path
			newValue := m.settingsInput.Value()
			if m.settingsCursor == 0 {
				// Registry path
				m.registryPath = newValue
				m.settingsItems[0].value = newValue
				m.settingsEditing = false
				m.settingsInput.Blur()
				// Reload manifest with new registry path
				return m, loadManifest(m.registryPath)
			} else if m.settingsCursor == 1 {
				// Project path
				m.projectPath = newValue
				m.settingsItems[1].value = newValue
				m.settingsEditing = false
				m.settingsInput.Blur()
				// Refresh installed cache
				m.refreshInstalledCache()
			}
			return m, nil
		case "esc":
			// Cancel editing
			m.settingsEditing = false
			m.settingsInput.Blur()
			return m, nil
		}
		// Pass other keys to text input
		var cmd tea.Cmd
		m.settingsInput, cmd = m.settingsInput.Update(msg)
		return m, cmd
	}

	// Normal navigation mode
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.settingsCursor < len(m.settingsItems)-1 {
			m.settingsCursor++
		}
	case key.Matches(msg, m.keys.Enter):
		item := m.settingsItems[m.settingsCursor]
		if item.isPath {
			// Start editing path
			m.settingsEditing = true
			m.settingsInput.SetValue(item.value)
			m.settingsInput.Focus()
			return m, nil
		} else if item.actionKey == "scan" {
			// Start scan path input
			m.settingsScanMode = true
			m.settingsInput.SetValue("")
			m.settingsInput.Placeholder = "Enter path to scan..."
			m.settingsInput.Focus()
			return m, nil
		} else if item.actionKey != "" {
			// Run action
			return m, m.runSettingsAction(item.actionKey)
		}
	}
	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	files := m.getDetailFiles()

	// Note: Edit mode is handled in the main Update function to ensure
	// the textarea receives all message types, not just key messages

	switch {
	case key.Matches(msg, m.keys.Edit):
		// Enter inline edit mode
		if m.detailRawContent != "" {
			m.detailEditMode = true
			m.detailTextarea.SetValue(m.detailRawContent)
			m.detailTextarea.SetWidth(m.width - 10)
			m.detailTextarea.SetHeight(m.height - 12)
			// Move cursor to beginning of file (not just line)
			lineCount := m.detailTextarea.LineCount()
			for i := 0; i < lineCount; i++ {
				m.detailTextarea.CursorUp()
			}
			m.detailTextarea.CursorStart()
			m.detailTextarea.Focus()
			return m, textarea.Blink
		}
		return m, nil

	case key.Matches(msg, m.keys.Add):
		// Add item to project
		if m.detailItem != nil {
			m.selected[m.detailItem.FullName()] = true
			return m, m.addSelectedToProject()
		}
	}

	// Handle number keys 1-9 for file selection
	keyStr := msg.String()
	if len(keyStr) == 1 && keyStr[0] >= '1' && keyStr[0] <= '9' {
		idx := int(keyStr[0] - '1')
		if idx < len(files) {
			m.detailFileIndex = idx
			m.detailLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadDetailFileContentCmd())
		}
	}

	return m, nil
}

func (m Model) updateStaging(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Note: Edit mode is handled in the main Update function to ensure
	// the textarea receives all message types, not just key messages

	// Handle detail mode (viewing file content)
	if m.stagingDetailMode {
		switch {
		case key.Matches(msg, m.keys.Back):
			// Go back to list
			m.stagingDetailMode = false
			m.stagingContent = ""
			m.stagingRawContent = ""
			return m, nil

		case key.Matches(msg, m.keys.Edit):
			// Enter edit mode with textarea
			if m.stagingCursor < len(m.stagingItems) {
				m.stagingEditMode = true
				m.stagingTextarea.SetValue(m.stagingRawContent)
				m.stagingTextarea.SetWidth(m.width - 10)
				m.stagingTextarea.SetHeight(m.height - 12)
				// Move cursor to beginning of file (not just line)
				lineCount := m.stagingTextarea.LineCount()
				for i := 0; i < lineCount; i++ {
					m.stagingTextarea.CursorUp()
				}
				m.stagingTextarea.CursorStart()
				m.stagingTextarea.Focus()
				return m, textarea.Blink
			}
			return m, nil

		case key.Matches(msg, m.keys.Delete):
			// Delete the file
			if m.stagingCursor < len(m.stagingItems) {
				m.stagingDetailMode = false
				return m, m.deleteStagingFile(m.stagingItems[m.stagingCursor])
			}
		}
		return m, nil
	}

	// List mode
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.stagingCursor > 0 {
			m.stagingCursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.stagingCursor < len(m.stagingItems)-1 {
			m.stagingCursor++
		}

	case key.Matches(msg, m.keys.Enter):
		// Enter detail mode - view file content
		if m.stagingCursor < len(m.stagingItems) {
			m.stagingDetailMode = true
			m.stagingLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadStagingFileContentCmd())
		}
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		// Delete staging file
		if m.stagingCursor < len(m.stagingItems) {
			return m, m.deleteStagingFile(m.stagingItems[m.stagingCursor])
		}

	case key.Matches(msg, m.keys.Process):
		// Process all staging files
		return m, m.processStagingFiles()
	}

	return m, nil
}

// saveStagingFile saves the content from the textarea to the staging file
func (m Model) saveStagingFile() tea.Cmd {
	if m.stagingCursor >= len(m.stagingItems) {
		return func() tea.Msg {
			return statusMsg{msg: "No file selected", isError: true}
		}
	}

	content := m.stagingTextarea.Value()
	filePath := filepath.Join(m.registryPath, "import", m.stagingItems[m.stagingCursor].Path)

	return func() tea.Msg {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return stagingFileSavedMsg{err: err}
		}
		return stagingFileSavedMsg{path: filePath}
	}
}

// stagingFileSavedMsg is sent when a staging file is saved
type stagingFileSavedMsg struct {
	path string
	err  error
}

// detailFileLoadedMsg is sent when a detail file is loaded
type detailFileLoadedMsg struct {
	content    string
	rawContent string
	err        error
}

// detailFileSavedMsg is sent when a detail file is saved
type detailFileSavedMsg struct {
	path string
	err  error
}

// stagingFileLoadedMsg is sent when a staging file is loaded
type stagingFileLoadedMsg struct {
	content    string
	rawContent string
	err        error
}

// editStagingFile opens a staging file in $EDITOR
func (m Model) editStagingFile(file importer.PendingFile) tea.Cmd {
	filePath := filepath.Join(m.registryPath, "import", file.Path)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	c := exec.Command(editor, filePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return stagingFileEditedMsg{path: filePath, err: err}
	})
}

// stagingFileEditedMsg is sent when a staging file edit completes
type stagingFileEditedMsg struct {
	path string
	err  error
}

// deleteStagingFile deletes a file from staging
func (m Model) deleteStagingFile(file importer.PendingFile) tea.Cmd {
	registryPath := m.registryPath
	filePath := file.Path

	return func() tea.Msg {
		fullPath := filepath.Join(registryPath, "import", filePath)
		if err := os.Remove(fullPath); err != nil {
			return statusMsg{msg: fmt.Sprintf("Failed to delete: %s", err.Error()), isError: true}
		}
		return stagingFileDeletedMsg{path: filePath}
	}
}

// stagingFileDeletedMsg is sent when a staging file is deleted
type stagingFileDeletedMsg struct {
	path string
}

// processStagingFiles processes all files in the staging directory
func (m Model) processStagingFiles() tea.Cmd {
	registryPath := m.registryPath

	return func() tea.Msg {
		imp := importer.NewImporter(registryPath)
		result, err := imp.ProcessStaging()
		if err != nil {
			return actionCompleteMsg{action: "process", success: false, message: err.Error()}
		}

		var details []string
		for _, item := range result.Processed {
			details = append(details, fmt.Sprintf("+ %s → %s", item.SourcePath, item.DestPath))
		}

		msg := fmt.Sprintf("Processed %d files", len(result.Processed))
		if len(result.Pending) > 0 {
			msg += fmt.Sprintf(" (%d still pending)", len(result.Pending))
		}

		return actionCompleteMsg{
			action:  "process",
			success: true,
			message: msg,
			details: details,
		}
	}
}

// runScan scans an external path for markdown files
func (m Model) runScan(scanPath string) tea.Cmd {
	registryPath := m.registryPath

	return func() tea.Msg {
		imp := importer.NewImporter(registryPath)

		result, err := imp.ScanAndImport(scanPath)
		if err != nil {
			return actionCompleteMsg{action: "scan", success: false, message: err.Error()}
		}

		var details []string
		for _, item := range result.Imported {
			details = append(details, fmt.Sprintf("+ %s → %s", item.SourcePath, item.DestPath))
		}
		for _, item := range result.Staged {
			details = append(details, fmt.Sprintf("? %s → import/ (needs headers)", item.SourcePath))
		}

		msg := fmt.Sprintf("Imported %d, staged %d files", len(result.Imported), len(result.Staged))
		if len(result.Errors) > 0 {
			msg += fmt.Sprintf(" (%d errors)", len(result.Errors))
		}

		return actionCompleteMsg{
			action:  "scan",
			success: len(result.Errors) == 0,
			message: msg,
			details: details,
		}
	}
}

// runSettingsAction executes a settings action
func (m Model) runSettingsAction(action string) tea.Cmd {
	registryPath := m.registryPath

	switch action {
	case "build":
		return m.buildRegistry()
	case "validate":
		return func() tea.Msg {
			result, err := registry.BuildRegistry(registryPath)
			if err != nil {
				return actionCompleteMsg{action: "validate", success: false, message: err.Error()}
			}
			issues := result.Validation.Issues
			errorCount := 0
			warnCount := 0
			for _, issue := range issues {
				if issue.Severity == registry.SeverityError {
					errorCount++
				} else if issue.Severity == registry.SeverityWarning {
					warnCount++
				}
			}
			if errorCount > 0 {
				details := make([]string, 0, len(issues))
				for _, issue := range issues {
					details = append(details, issue.String())
				}
				return actionCompleteMsg{
					action:  "validate",
					success: false,
					message: fmt.Sprintf("Found %d errors, %d warnings", errorCount, warnCount),
					details: details,
				}
			}
			msg := fmt.Sprintf("All %d items valid", len(result.Manifest.Items))
			if warnCount > 0 {
				msg += fmt.Sprintf(" (%d warnings)", warnCount)
			}
			return actionCompleteMsg{action: "validate", success: true, message: msg}
		}
	case "orphans":
		return func() tea.Msg {
			manifest, err := registry.LoadManifestFromRegistry(registryPath)
			if err != nil {
				return actionCompleteMsg{action: "orphans", success: false, message: err.Error()}
			}
			orphans := findOrphanFiles(registryPath, manifest)
			if len(orphans) > 0 {
				return actionCompleteMsg{
					action:  "orphans",
					success: true,
					message: fmt.Sprintf("Found %d orphaned files", len(orphans)),
					details: orphans,
				}
			}
			return actionCompleteMsg{action: "orphans", success: true, message: "No orphaned files found"}
		}
	case "reindex":
		return func() tea.Msg {
			_, err := registry.BuildRegistry(registryPath)
			if err != nil {
				return actionCompleteMsg{action: "reindex", success: false, message: err.Error()}
			}
			return actionCompleteMsg{action: "reindex", success: true, message: "Registry reindexed successfully"}
		}
	case "scan":
		return func() tea.Msg {
			// Scan is typically for external paths, just show a message for now
			return actionCompleteMsg{
				action:  "scan",
				success: true,
				message: "Use 'regis3 scan <path>' from CLI to scan external directories",
			}
		}
	}
	return nil
}

// findOrphanFiles finds markdown files not in the manifest
func findOrphanFiles(registryPath string, manifest *registry.Manifest) []string {
	knownFiles := make(map[string]bool)
	for _, item := range manifest.Items {
		knownFiles[item.Source] = true
		for _, f := range item.Files {
			knownFiles[f] = true
		}
	}

	var orphans []string
	filepath.Walk(registryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "import" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		relPath, err := filepath.Rel(registryPath, path)
		if err != nil {
			return nil
		}
		if !knownFiles[relPath] {
			orphans = append(orphans, relPath)
		}
		return nil
	})
	return orphans
}

// actionCompleteMsg is sent when an action completes
type actionCompleteMsg struct {
	action  string
	success bool
	message string
	details []string
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

// refreshInstalledCache updates the installed items cache
func (m *Model) refreshInstalledCache() {
	m.installed = make(map[string]bool)
	for _, item := range m.getInstalledItems() {
		m.installed[item.ref] = true
	}
}

// getInstalledItems returns items installed in the current project
func (m Model) getInstalledItems() []installedItem {
	target := installer.DefaultClaudeTarget()

	inst, err := installer.NewInstaller(m.projectPath, m.registryPath, target)
	if err != nil {
		return nil
	}

	// Get status (pass manifest if available, empty one if not)
	manifest := m.manifest
	if manifest == nil {
		manifest = &registry.Manifest{Items: make(map[string]*registry.Item)}
	}

	status := inst.Status(manifest)

	var items []installedItem
	for _, s := range status.Items {
		if s.Installed {
			ref := fmt.Sprintf("%s:%s", s.Type, s.Name)
			desc := ""
			// Get description from manifest if available
			if item, ok := manifest.Items[ref]; ok {
				desc = item.Desc
			}
			items = append(items, installedItem{
				ref:         ref,
				desc:        desc,
				needsUpdate: s.NeedsUpdate,
			})
		}
	}

	// Sort by ref for consistent ordering
	sort.Slice(items, func(i, j int) bool {
		return items[i].ref < items[j].ref
	})

	return items
}

// installedItem represents an installed item in the project
type installedItem struct {
	ref         string
	desc        string
	needsUpdate bool
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

	// Sort by type then name for consistent ordering
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type < items[j].Type
		}
		return items[i].Name < items[j].Name
	})

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
	case ViewStaging:
		return m.viewStaging()
	}

	return ""
}

// getDetailFiles returns the list of files for the current detail item
func (m Model) getDetailFiles() []string {
	if m.detailItem == nil {
		return nil
	}
	files := []string{m.detailItem.Source}
	files = append(files, m.detailItem.Files...)
	return files
}

// loadDetailFileContentCmd returns a command to load detail file content asynchronously
func (m Model) loadDetailFileContentCmd() tea.Cmd {
	files := m.getDetailFiles()
	if len(files) == 0 || m.detailFileIndex >= len(files) {
		return func() tea.Msg {
			return detailFileLoadedMsg{content: "", rawContent: ""}
		}
	}

	// Source (index 0) is already relative to registry root
	// Additional files are relative to SourceDir
	var filePath string
	if m.detailFileIndex == 0 {
		filePath = filepath.Join(m.registryPath, files[0])
	} else {
		filePath = filepath.Join(m.registryPath, m.detailItem.SourceDir, files[m.detailFileIndex])
	}

	width := m.width

	return func() tea.Msg {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return detailFileLoadedMsg{
				content: fmt.Sprintf("Error reading file: %s", err.Error()),
				err:     err,
			}
		}

		rawContent := string(content)
		renderedContent := rawContent

		// Render markdown for .md files
		if strings.HasSuffix(strings.ToLower(filePath), ".md") {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(width-10),
			)
			if err == nil {
				rendered, err := renderer.Render(rawContent)
				if err == nil {
					renderedContent = rendered
				}
			}
		}

		return detailFileLoadedMsg{
			content:    renderedContent,
			rawContent: rawContent,
		}
	}
}

// loadStagingFileContentCmd returns a command to load staging file content asynchronously
func (m Model) loadStagingFileContentCmd() tea.Cmd {
	if len(m.stagingItems) == 0 || m.stagingCursor >= len(m.stagingItems) {
		return func() tea.Msg {
			return stagingFileLoadedMsg{content: "", rawContent: ""}
		}
	}

	filePath := filepath.Join(m.registryPath, "import", m.stagingItems[m.stagingCursor].Path)
	width := m.width

	return func() tea.Msg {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return stagingFileLoadedMsg{
				content: fmt.Sprintf("Error reading file: %s", err.Error()),
				err:     err,
			}
		}

		rawContent := string(content)
		renderedContent := rawContent

		// Render markdown for .md files
		if strings.HasSuffix(strings.ToLower(filePath), ".md") {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(width-10),
			)
			if err == nil {
				rendered, err := renderer.Render(rawContent)
				if err == nil {
					renderedContent = rendered
				}
			}
		}

		return stagingFileLoadedMsg{
			content:    renderedContent,
			rawContent: rawContent,
		}
	}
}

// editDetailFile opens the current detail file in $EDITOR
func (m Model) editDetailFile() tea.Cmd {
	files := m.getDetailFiles()
	if len(files) == 0 || m.detailFileIndex >= len(files) {
		return func() tea.Msg {
			return statusMsg{msg: "No file selected", isError: true}
		}
	}

	// Source (index 0) is already relative to registry root
	// Additional files are relative to SourceDir
	var filePath string
	if m.detailFileIndex == 0 {
		filePath = filepath.Join(m.registryPath, files[0])
	} else {
		filePath = filepath.Join(m.registryPath, m.detailItem.SourceDir, files[m.detailFileIndex])
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	c := exec.Command(editor, filePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return fileEditedMsg{path: filePath, err: err}
	})
}

// saveDetailFile saves the current detail file from the textarea
func (m Model) saveDetailFile() tea.Cmd {
	files := m.getDetailFiles()
	if len(files) == 0 || m.detailFileIndex >= len(files) {
		return func() tea.Msg {
			return detailFileSavedMsg{err: fmt.Errorf("no file selected")}
		}
	}

	// Source (index 0) is already relative to registry root
	// Additional files are relative to SourceDir
	var filePath string
	if m.detailFileIndex == 0 {
		filePath = filepath.Join(m.registryPath, files[0])
	} else {
		filePath = filepath.Join(m.registryPath, m.detailItem.SourceDir, files[m.detailFileIndex])
	}

	content := m.detailTextarea.Value()

	return func() tea.Msg {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return detailFileSavedMsg{path: filePath, err: err}
		}
		return detailFileSavedMsg{path: filePath}
	}
}

// Editor helper functions for shortcuts

// editorGotoLine moves the cursor to the specified line number in the given textarea
func editorGotoLine(ta *textarea.Model, lineNum int) {
	if lineNum < 1 {
		lineNum = 1
	}
	lineCount := ta.LineCount()
	if lineNum > lineCount {
		lineNum = lineCount
	}

	// Move to beginning first
	for i := 0; i < lineCount; i++ {
		ta.CursorUp()
	}
	ta.CursorStart()

	// Move down to target line
	for i := 1; i < lineNum; i++ {
		ta.CursorDown()
	}
}

// editorDuplicateLine duplicates the current line
func editorDuplicateLine(ta *textarea.Model) {
	content := ta.Value()
	lines := strings.Split(content, "\n")
	currentLine := ta.Line()

	if currentLine >= 0 && currentLine < len(lines) {
		// Insert duplicate after current line
		newLines := make([]string, 0, len(lines)+1)
		for i, line := range lines {
			newLines = append(newLines, line)
			if i == currentLine {
				newLines = append(newLines, line)
			}
		}

		// Preserve cursor position
		ta.SetValue(strings.Join(newLines, "\n"))
		editorGotoLine(ta, currentLine+2) // Move to duplicated line
	}
}

// editorJoinLines joins the current line with the next line
func editorJoinLines(ta *textarea.Model) {
	content := ta.Value()
	lines := strings.Split(content, "\n")
	currentLine := ta.Line()

	if currentLine >= 0 && currentLine < len(lines)-1 {
		// Join current line with next
		lines[currentLine] = lines[currentLine] + " " + strings.TrimLeft(lines[currentLine+1], " \t")
		newLines := append(lines[:currentLine+1], lines[currentLine+2:]...)

		ta.SetValue(strings.Join(newLines, "\n"))
		editorGotoLine(ta, currentLine+1)
		ta.CursorEnd() // Move to end of joined line
	}
}

// editorToggleComment toggles comment on the current line (using # for markdown)
func editorToggleComment(ta *textarea.Model) {
	content := ta.Value()
	lines := strings.Split(content, "\n")
	currentLine := ta.Line()

	if currentLine >= 0 && currentLine < len(lines) {
		line := lines[currentLine]
		trimmed := strings.TrimLeft(line, " \t")
		indent := line[:len(line)-len(trimmed)]

		if strings.HasPrefix(trimmed, "# ") {
			// Remove comment
			lines[currentLine] = indent + strings.TrimPrefix(trimmed, "# ")
		} else if strings.HasPrefix(trimmed, "#") {
			// Remove comment without space
			lines[currentLine] = indent + strings.TrimPrefix(trimmed, "#")
		} else {
			// Add comment
			lines[currentLine] = indent + "# " + trimmed
		}

		ta.SetValue(strings.Join(lines, "\n"))
		editorGotoLine(ta, currentLine+1)
	}
}

// editorIndent adds indentation to the current line
func editorIndent(ta *textarea.Model) {
	content := ta.Value()
	lines := strings.Split(content, "\n")
	currentLine := ta.Line()

	if currentLine >= 0 && currentLine < len(lines) {
		lines[currentLine] = "  " + lines[currentLine] // Add 2 spaces
		ta.SetValue(strings.Join(lines, "\n"))
		editorGotoLine(ta, currentLine+1)
	}
}

// editorUnindent removes indentation from the current line
func editorUnindent(ta *textarea.Model) {
	content := ta.Value()
	lines := strings.Split(content, "\n")
	currentLine := ta.Line()

	if currentLine >= 0 && currentLine < len(lines) {
		line := lines[currentLine]
		if strings.HasPrefix(line, "  ") {
			lines[currentLine] = strings.TrimPrefix(line, "  ")
		} else if strings.HasPrefix(line, "\t") {
			lines[currentLine] = strings.TrimPrefix(line, "\t")
		} else if strings.HasPrefix(line, " ") {
			lines[currentLine] = strings.TrimPrefix(line, " ")
		}
		ta.SetValue(strings.Join(lines, "\n"))
		editorGotoLine(ta, currentLine+1)
	}
}
