// Package config provides configuration management for regis3.
package config

import (
	"os"
	"path/filepath"
)

const (
	// AppName is the application name used for config directories.
	AppName = "regis3"

	// DefaultConfigFile is the default config filename.
	DefaultConfigFile = "config.yaml"

	// DefaultManifestDir is the directory where manifests are built.
	DefaultManifestDir = ".build"

	// DefaultManifestFile is the manifest filename.
	DefaultManifestFile = "manifest.json"

	// ClaudeConfigDir is the directory for Claude Code configuration.
	ClaudeConfigDir = ".claude"

	// ClaudeRootFile is the main Claude Code configuration file.
	ClaudeRootFile = "CLAUDE.md"

	// InstalledFile tracks what's been installed.
	InstalledFile = "installed.json"
)

// Paths contains resolved paths for the application.
type Paths struct {
	// Home is the user's home directory.
	Home string

	// ConfigDir is the regis3 configuration directory (~/.regis3).
	ConfigDir string

	// ConfigFile is the full path to the config file.
	ConfigFile string

	// RegistryDir is the registry directory path.
	RegistryDir string

	// WorkDir is the current working directory.
	WorkDir string
}

// NewPaths creates a new Paths instance with resolved paths.
func NewPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, "."+AppName)

	return &Paths{
		Home:        home,
		ConfigDir:   configDir,
		ConfigFile:  filepath.Join(configDir, DefaultConfigFile),
		RegistryDir: filepath.Join(configDir, "registry"),
		WorkDir:     workDir,
	}, nil
}

// ManifestDir returns the path to the manifest build directory.
func (p *Paths) ManifestDir() string {
	return filepath.Join(p.RegistryDir, DefaultManifestDir)
}

// ManifestFile returns the path to the manifest file.
func (p *Paths) ManifestFile() string {
	return filepath.Join(p.ManifestDir(), DefaultManifestFile)
}

// ProjectClaudeDir returns the .claude directory in the current project.
func (p *Paths) ProjectClaudeDir() string {
	return filepath.Join(p.WorkDir, ClaudeConfigDir)
}

// ProjectClaudeFile returns the CLAUDE.md file in the current project.
func (p *Paths) ProjectClaudeFile() string {
	return filepath.Join(p.WorkDir, ClaudeRootFile)
}

// ProjectInstalledFile returns the installed.json file path.
func (p *Paths) ProjectInstalledFile() string {
	return filepath.Join(p.ProjectClaudeDir(), InstalledFile)
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func (p *Paths) EnsureConfigDir() error {
	return os.MkdirAll(p.ConfigDir, 0755)
}

// EnsureRegistryDir creates the registry directory if it doesn't exist.
func (p *Paths) EnsureRegistryDir() error {
	return os.MkdirAll(p.RegistryDir, 0755)
}

// EnsureManifestDir creates the manifest build directory if it doesn't exist.
func (p *Paths) EnsureManifestDir() error {
	return os.MkdirAll(p.ManifestDir(), 0755)
}

// EnsureProjectClaudeDir creates the .claude directory in the project.
func (p *Paths) EnsureProjectClaudeDir() error {
	return os.MkdirAll(p.ProjectClaudeDir(), 0755)
}

// SkillDir returns the path for a skill installation.
func (p *Paths) SkillDir(name string) string {
	return filepath.Join(p.ProjectClaudeDir(), "skills", name)
}

// AgentFile returns the path for an agent file.
func (p *Paths) AgentFile(name string) string {
	return filepath.Join(p.ProjectClaudeDir(), "agents", name+".md")
}

// CommandFile returns the path for a command file.
func (p *Paths) CommandFile(name string) string {
	return filepath.Join(p.ProjectClaudeDir(), "commands", name+".md")
}

// MCPFile returns the path for an MCP config file.
func (p *Paths) MCPFile(name string) string {
	return filepath.Join(p.ProjectClaudeDir(), "mcp", name+".json")
}

// DocFile returns the path for a doc file.
func (p *Paths) DocFile(name string) string {
	return filepath.Join(p.ProjectClaudeDir(), "docs", name+".md")
}

// ScriptFile returns the path for a script file.
func (p *Paths) ScriptFile(name string) string {
	return filepath.Join(p.ProjectClaudeDir(), "scripts", name+".sh")
}
