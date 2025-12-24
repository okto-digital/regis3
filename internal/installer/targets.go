package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Target represents an installation target (e.g., Claude Code, Cursor).
type Target struct {
	// Name is the target identifier.
	Name string `yaml:"name"`

	// Description describes this target.
	Description string `yaml:"description"`

	// Version is the target definition version.
	Version string `yaml:"version"`

	// BaseDir is the base directory for installations (relative to project root).
	BaseDir string `yaml:"base_dir"`

	// MergeFile is the file where merge types are combined (e.g., CLAUDE.md).
	MergeFile string `yaml:"merge_file"`

	// Paths defines where each item type is installed.
	Paths map[string]PathConfig `yaml:"paths"`

	// Transforms defines content transformations per type.
	Transforms map[string]TransformConfig `yaml:"transforms"`
}

// PathConfig defines the installation path for an item type.
type PathConfig struct {
	// Dir is the directory for this type (relative to BaseDir).
	Dir string `yaml:"dir"`

	// Pattern is the filename pattern. Use {name} as placeholder.
	Pattern string `yaml:"pattern"`

	// Subdirs indicates if each item gets its own subdirectory.
	Subdirs bool `yaml:"subdirs"`
}

// TransformConfig defines content transformations.
type TransformConfig struct {
	// WrapWith wraps content in this template. Use {content} as placeholder.
	WrapWith string `yaml:"wrap_with"`

	// StripFrontmatter removes YAML frontmatter from content.
	StripFrontmatter bool `yaml:"strip_frontmatter"`

	// AddHeader prepends this text to the content.
	AddHeader string `yaml:"add_header"`
}

// GetPath returns the installation path for an item.
func (t *Target) GetPath(itemType, name string) (string, error) {
	pathCfg, ok := t.Paths[itemType]
	if !ok {
		return "", fmt.Errorf("unknown item type: %s", itemType)
	}

	// Build path
	dir := filepath.Join(t.BaseDir, pathCfg.Dir)

	// Apply pattern
	filename := pathCfg.Pattern
	if filename == "" {
		filename = "{name}.md"
	}

	// Replace placeholders
	filename = replacePlaceholder(filename, "name", name)

	if pathCfg.Subdirs {
		return filepath.Join(dir, name, filename), nil
	}

	return filepath.Join(dir, filename), nil
}

// GetTransform returns the transform config for an item type.
func (t *Target) GetTransform(itemType string) TransformConfig {
	if cfg, ok := t.Transforms[itemType]; ok {
		return cfg
	}
	return TransformConfig{}
}

// IsMergeType returns true if this type merges into the merge file.
func (t *Target) IsMergeType(itemType string) bool {
	switch itemType {
	case "philosophy", "project", "ruleset":
		return true
	default:
		return false
	}
}

// replacePlaceholder replaces {key} with value in the pattern.
func replacePlaceholder(pattern, key, value string) string {
	placeholder := "{" + key + "}"
	result := pattern
	for i := 0; i < len(result); i++ {
		if i+len(placeholder) <= len(result) && result[i:i+len(placeholder)] == placeholder {
			result = result[:i] + value + result[i+len(placeholder):]
			i += len(value) - 1
		}
	}
	return result
}

// LoadTarget loads a target definition from a YAML file.
func LoadTarget(path string) (*Target, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read target file: %w", err)
	}

	var target Target
	if err := yaml.Unmarshal(data, &target); err != nil {
		return nil, fmt.Errorf("failed to parse target file: %w", err)
	}

	return &target, nil
}

// LoadTargetByName loads a target from the targets directory.
func LoadTargetByName(targetsDir, name string) (*Target, error) {
	path := filepath.Join(targetsDir, name+".yaml")
	return LoadTarget(path)
}

// DefaultClaudeTarget returns the default Claude Code target configuration.
func DefaultClaudeTarget() *Target {
	return &Target{
		Name:        "claude",
		Description: "Claude Code / Claude Desktop target",
		Version:     "1.0.0",
		BaseDir:     ".claude",
		MergeFile:   "CLAUDE.md",
		Paths: map[string]PathConfig{
			"skill": {
				Dir:     "skills",
				Pattern: "SKILL.md",
				Subdirs: true,
			},
			"subagent": {
				Dir:     "agents",
				Pattern: "{name}.md",
				Subdirs: false,
			},
			"command": {
				Dir:     "commands",
				Pattern: "{name}.md",
				Subdirs: false,
			},
			"mcp": {
				Dir:     "mcp",
				Pattern: "{name}.json",
				Subdirs: false,
			},
			"script": {
				Dir:     "scripts",
				Pattern: "{name}.sh",
				Subdirs: false,
			},
			"doc": {
				Dir:     "docs",
				Pattern: "{name}.md",
				Subdirs: false,
			},
			"hook": {
				Dir:     "hooks",
				Pattern: "{name}.md",
				Subdirs: false,
			},
			"prompt": {
				Dir:     "prompts",
				Pattern: "{name}.md",
				Subdirs: false,
			},
			// Merge types don't have paths - they merge into CLAUDE.md
			"philosophy": {},
			"project":    {},
			"ruleset":    {},
			"stack":      {}, // Stacks are meta-types, no direct installation
		},
		Transforms: map[string]TransformConfig{
			"skill": {
				StripFrontmatter: true,
			},
			"subagent": {
				StripFrontmatter: true,
			},
			"philosophy": {
				StripFrontmatter: true,
			},
			"project": {
				StripFrontmatter: true,
			},
			"ruleset": {
				StripFrontmatter: true,
			},
		},
	}
}

// ListAvailableTargets returns available target names from a directory.
func ListAvailableTargets(targetsDir string) ([]string, error) {
	entries, err := os.ReadDir(targetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read targets directory: %w", err)
	}

	var targets []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".yaml" || filepath.Ext(name) == ".yml" {
			targets = append(targets, name[:len(name)-len(filepath.Ext(name))])
		}
	}

	return targets, nil
}
