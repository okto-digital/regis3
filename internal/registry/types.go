// Package registry provides types and operations for managing the regis3 registry.
package registry

import (
	"fmt"
	"time"
)

// ItemType represents the type of a registry item.
type ItemType string

const (
	// Install types - create separate files in .claude/
	TypeSkill    ItemType = "skill"
	TypeSubagent ItemType = "subagent"
	TypeCommand  ItemType = "command"
	TypeMCP      ItemType = "mcp"
	TypeScript   ItemType = "script"
	TypeDoc      ItemType = "doc"

	// Merge types - content merged into CLAUDE.md
	TypeProject    ItemType = "project"
	TypePhilosophy ItemType = "philosophy"
	TypeRuleset    ItemType = "ruleset"

	// Meta types
	TypeStack  ItemType = "stack"
	TypeHook   ItemType = "hook"
	TypePrompt ItemType = "prompt"
)

// ValidTypes contains all valid item types.
var ValidTypes = []ItemType{
	TypeSkill, TypeSubagent, TypeCommand, TypeMCP, TypeScript, TypeDoc,
	TypeProject, TypePhilosophy, TypeRuleset,
	TypeStack, TypeHook, TypePrompt,
}

// IsValidType checks if a string is a valid item type.
func IsValidType(s string) bool {
	for _, t := range ValidTypes {
		if string(t) == s {
			return true
		}
	}
	return false
}

// IsInstallType returns true if the type creates separate files.
func (t ItemType) IsInstallType() bool {
	switch t {
	case TypeSkill, TypeSubagent, TypeCommand, TypeMCP, TypeScript, TypeDoc:
		return true
	}
	return false
}

// IsMergeType returns true if the type merges into CLAUDE.md.
func (t ItemType) IsMergeType() bool {
	switch t {
	case TypeProject, TypePhilosophy, TypeRuleset:
		return true
	}
	return false
}

// ItemStatus represents the status of a registry item.
type ItemStatus string

const (
	StatusDraft      ItemStatus = "draft"
	StatusStable     ItemStatus = "stable"
	StatusDeprecated ItemStatus = "deprecated"
)

// HookTrigger represents when a hook should run.
type HookTrigger string

const (
	TriggerPreInstall  HookTrigger = "pre-install"
	TriggerPostInstall HookTrigger = "post-install"
	TriggerPreBuild    HookTrigger = "pre-build"
	TriggerPostBuild   HookTrigger = "post-build"
)

// TargetOverride contains per-target configuration overrides.
type TargetOverride struct {
	Exclude  bool   `yaml:"exclude,omitempty" json:"exclude,omitempty"`
	Priority string `yaml:"priority,omitempty" json:"priority,omitempty"`
}

// Regis3Meta contains the regis3 namespace metadata from YAML frontmatter.
type Regis3Meta struct {
	Type    string                    `yaml:"type" json:"type"`
	Name    string                    `yaml:"name" json:"name"`
	Desc    string                    `yaml:"desc" json:"desc"`
	Cat     string                    `yaml:"cat,omitempty" json:"cat,omitempty"`
	Deps    []string                  `yaml:"deps,omitempty" json:"deps,omitempty"`
	Tags    []string                  `yaml:"tags,omitempty" json:"tags,omitempty"`
	Files   []string                  `yaml:"files,omitempty" json:"files,omitempty"`
	Status  string                    `yaml:"status,omitempty" json:"status,omitempty"`
	Author  string                    `yaml:"author,omitempty" json:"author,omitempty"`
	Order   int                       `yaml:"order,omitempty" json:"order,omitempty"`
	Target  map[string]TargetOverride `yaml:"target,omitempty" json:"target,omitempty"`
	Trigger string                    `yaml:"trigger,omitempty" json:"trigger,omitempty"`
	Run     string                    `yaml:"run,omitempty" json:"run,omitempty"`
}

// FrontMatter wraps the regis3 namespace for parsing.
type FrontMatter struct {
	Regis3 Regis3Meta `yaml:"regis3"`
}

// Item represents a single registry item with computed fields.
type Item struct {
	Regis3Meta

	// Source is the path to the source .md file (relative to registry root).
	Source string `json:"source"`

	// Content is the markdown body (excluding frontmatter).
	Content string `json:"-"`

	// SourceDir is the directory containing the source file.
	SourceDir string `json:"-"`
}

// FullName returns the type:name identifier for the item.
func (i *Item) FullName() string {
	return fmt.Sprintf("%s:%s", i.Type, i.Name)
}

// ItemType returns the parsed ItemType.
func (i *Item) ItemType() ItemType {
	return ItemType(i.Type)
}

// Manifest represents the built registry index.
type Manifest struct {
	Version      string           `json:"version"`
	Generated    time.Time        `json:"generated"`
	RegistryPath string           `json:"registry_path"`
	Items        map[string]*Item `json:"items"`
	Stats        Stats            `json:"stats"`
}

// NewManifest creates a new empty manifest.
func NewManifest(registryPath string) *Manifest {
	return &Manifest{
		Version:      "1.0.0",
		Generated:    time.Now(),
		RegistryPath: registryPath,
		Items:        make(map[string]*Item),
	}
}

// AddItem adds an item to the manifest.
func (m *Manifest) AddItem(item *Item) {
	m.Items[item.FullName()] = item
}

// GetItem retrieves an item by its full name (type:name).
func (m *Manifest) GetItem(fullName string) (*Item, bool) {
	item, ok := m.Items[fullName]
	return item, ok
}

// ComputeStats calculates statistics about the manifest.
func (m *Manifest) ComputeStats() {
	m.Stats = Stats{}
	for _, item := range m.Items {
		switch ItemType(item.Type) {
		case TypeSkill:
			m.Stats.Skills++
		case TypeSubagent:
			m.Stats.Subagents++
		case TypeCommand:
			m.Stats.Commands++
		case TypeMCP:
			m.Stats.MCPs++
		case TypeScript:
			m.Stats.Scripts++
		case TypeDoc:
			m.Stats.Docs++
		case TypeProject:
			m.Stats.Projects++
		case TypePhilosophy:
			m.Stats.Philosophies++
		case TypeRuleset:
			m.Stats.Rulesets++
		case TypeStack:
			m.Stats.Stacks++
		case TypeHook:
			m.Stats.Hooks++
		case TypePrompt:
			m.Stats.Prompts++
		}
	}
}

// Stats contains counts of each item type.
type Stats struct {
	Skills       int `json:"skills"`
	Subagents    int `json:"subagents"`
	Commands     int `json:"commands"`
	MCPs         int `json:"mcps"`
	Scripts      int `json:"scripts"`
	Docs         int `json:"docs"`
	Projects     int `json:"projects"`
	Philosophies int `json:"philosophies"`
	Rulesets     int `json:"rulesets"`
	Stacks       int `json:"stacks"`
	Hooks        int `json:"hooks"`
	Prompts      int `json:"prompts"`
}

// Total returns the total number of items.
func (s Stats) Total() int {
	return s.Skills + s.Subagents + s.Commands + s.MCPs + s.Scripts +
		s.Docs + s.Projects + s.Philosophies + s.Rulesets +
		s.Stacks + s.Hooks + s.Prompts
}

// ProjectStatus tracks what's installed in a project.
type ProjectStatus struct {
	Target    string   `json:"target"`
	Installed []string `json:"installed"`
	Date      string   `json:"date"`
}
