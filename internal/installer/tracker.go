package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// TrackerFile is the name of the installation tracking file.
	TrackerFile = "installed.json"
)

// Tracker tracks installed items in a project.
type Tracker struct {
	// Path is the path to the tracking file.
	Path string

	// Data contains the tracking data.
	Data *TrackerData
}

// TrackerData is the structure of the tracking file.
type TrackerData struct {
	// Version is the tracker file format version.
	Version string `json:"version"`

	// Target is the installation target name.
	Target string `json:"target"`

	// LastUpdated is when the tracker was last updated.
	LastUpdated time.Time `json:"last_updated"`

	// RegistryPath is the path to the registry used.
	RegistryPath string `json:"registry_path,omitempty"`

	// Items maps item IDs to installation info.
	Items map[string]*InstalledItem `json:"items"`
}

// InstalledItem tracks a single installed item.
type InstalledItem struct {
	// ID is the full item ID (type:name).
	ID string `json:"id"`

	// Type is the item type.
	Type string `json:"type"`

	// Name is the item name.
	Name string `json:"name"`

	// InstalledAt is when the item was installed.
	InstalledAt time.Time `json:"installed_at"`

	// UpdatedAt is when the item was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// Version is the item version (if specified).
	Version string `json:"version,omitempty"`

	// SourceHash is a hash of the source content.
	SourceHash string `json:"source_hash,omitempty"`

	// InstalledPath is where the item was installed.
	InstalledPath string `json:"installed_path,omitempty"`

	// Merged indicates if this was merged into CLAUDE.md.
	Merged bool `json:"merged,omitempty"`
}

// NewTracker creates a new tracker for a project directory.
func NewTracker(projectDir, targetName string) *Tracker {
	return &Tracker{
		Path: filepath.Join(projectDir, ".claude", TrackerFile),
		Data: &TrackerData{
			Version:     "1.0.0",
			Target:      targetName,
			LastUpdated: time.Now(),
			Items:       make(map[string]*InstalledItem),
		},
	}
}

// Load loads the tracker data from disk.
func (t *Tracker) Load() error {
	data, err := os.ReadFile(t.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// No tracker file yet - that's OK
			return nil
		}
		return fmt.Errorf("failed to read tracker file: %w", err)
	}

	if err := json.Unmarshal(data, &t.Data); err != nil {
		return fmt.Errorf("failed to parse tracker file: %w", err)
	}

	// Ensure Items map exists
	if t.Data.Items == nil {
		t.Data.Items = make(map[string]*InstalledItem)
	}

	return nil
}

// Save saves the tracker data to disk.
func (t *Tracker) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(t.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tracker directory: %w", err)
	}

	t.Data.LastUpdated = time.Now()

	data, err := json.MarshalIndent(t.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tracker data: %w", err)
	}

	if err := os.WriteFile(t.Path, data, 0644); err != nil {
		return fmt.Errorf("failed to write tracker file: %w", err)
	}

	return nil
}

// IsInstalled checks if an item is installed.
func (t *Tracker) IsInstalled(id string) bool {
	_, ok := t.Data.Items[id]
	return ok
}

// GetInstalled returns the installation info for an item.
func (t *Tracker) GetInstalled(id string) *InstalledItem {
	return t.Data.Items[id]
}

// MarkInstalled marks an item as installed.
func (t *Tracker) MarkInstalled(id, itemType, name, path string, merged bool) {
	now := time.Now()

	if existing, ok := t.Data.Items[id]; ok {
		// Update existing
		existing.UpdatedAt = now
		existing.InstalledPath = path
		existing.Merged = merged
	} else {
		// New installation
		t.Data.Items[id] = &InstalledItem{
			ID:            id,
			Type:          itemType,
			Name:          name,
			InstalledAt:   now,
			UpdatedAt:     now,
			InstalledPath: path,
			Merged:        merged,
		}
	}
}

// MarkUninstalled removes an item from the tracker.
func (t *Tracker) MarkUninstalled(id string) {
	delete(t.Data.Items, id)
}

// ListInstalled returns all installed item IDs.
func (t *Tracker) ListInstalled() []string {
	ids := make([]string, 0, len(t.Data.Items))
	for id := range t.Data.Items {
		ids = append(ids, id)
	}
	return ids
}

// ListInstalledByType returns installed items of a specific type.
func (t *Tracker) ListInstalledByType(itemType string) []*InstalledItem {
	var items []*InstalledItem
	for _, item := range t.Data.Items {
		if item.Type == itemType {
			items = append(items, item)
		}
	}
	return items
}

// NeedsUpdate checks if an item needs updating based on source hash.
func (t *Tracker) NeedsUpdate(id, sourceHash string) bool {
	installed := t.Data.Items[id]
	if installed == nil {
		return true // Not installed, needs install
	}
	return installed.SourceHash != sourceHash
}

// SetSourceHash sets the source hash for an installed item.
func (t *Tracker) SetSourceHash(id, hash string) {
	if item, ok := t.Data.Items[id]; ok {
		item.SourceHash = hash
		item.UpdatedAt = time.Now()
	}
}

// Count returns the number of installed items.
func (t *Tracker) Count() int {
	return len(t.Data.Items)
}

// SetRegistryPath sets the registry path in the tracker.
func (t *Tracker) SetRegistryPath(path string) {
	t.Data.RegistryPath = path
}

// LoadTracker loads or creates a tracker for a project.
func LoadTracker(projectDir, targetName string) (*Tracker, error) {
	tracker := NewTracker(projectDir, targetName)
	if err := tracker.Load(); err != nil {
		return nil, err
	}
	return tracker, nil
}

// TrackerExists checks if a tracker file exists in the project.
func TrackerExists(projectDir string) bool {
	path := filepath.Join(projectDir, ".claude", TrackerFile)
	_, err := os.Stat(path)
	return err == nil
}
