package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	// RegistryPath is the path to the registry directory.
	RegistryPath string `mapstructure:"registry_path"`

	// DefaultTarget is the default output target (claude, cursor, gpt).
	DefaultTarget string `mapstructure:"default_target"`

	// OutputFormat is the default output format (pretty, json, quiet).
	OutputFormat string `mapstructure:"output_format"`

	// Debug enables debug output.
	Debug bool `mapstructure:"debug"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	paths, _ := NewPaths()
	registryPath := ""
	if paths != nil {
		registryPath = paths.RegistryDir
	}

	return &Config{
		RegistryPath:  registryPath,
		DefaultTarget: "claude",
		OutputFormat:  "pretty",
		Debug:         false,
	}
}

// Load loads configuration from file and environment.
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	// Set defaults
	v.SetDefault("registry_path", cfg.RegistryPath)
	v.SetDefault("default_target", cfg.DefaultTarget)
	v.SetDefault("output_format", cfg.OutputFormat)
	v.SetDefault("debug", cfg.Debug)

	// Environment variables (REGIS3_REGISTRY_PATH, etc.)
	v.SetEnvPrefix("REGIS3")
	v.AutomaticEnv()

	// Load from file if specified or default location
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		paths, err := NewPaths()
		if err == nil {
			v.SetConfigFile(paths.ConfigFile)
		}
	}

	// Try to read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's not a "file not found" error
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	// Unmarshal into struct
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Expand home directory in registry path
	if len(cfg.RegistryPath) > 0 && cfg.RegistryPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.RegistryPath = filepath.Join(home, cfg.RegistryPath[1:])
		}
	}

	return cfg, nil
}

// LoadFromPath loads configuration from a specific path.
func LoadFromPath(path string) (*Config, error) {
	return Load(path)
}

// LoadDefault loads configuration from default location.
func LoadDefault() (*Config, error) {
	return Load("")
}

// DefaultRegistryPath returns the default registry path.
func DefaultRegistryPath() string {
	paths, err := NewPaths()
	if err != nil {
		return ""
	}
	return paths.RegistryDir
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	paths, err := NewPaths()
	if err != nil {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".regis3", "config.yaml")
	}
	return paths.ConfigFile
}

// Save saves the configuration to a file.
func Save(cfg *Config, path string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path)

	v.Set("registry_path", cfg.RegistryPath)
	v.Set("default_target", cfg.DefaultTarget)
	v.Set("output_format", cfg.OutputFormat)
	v.Set("debug", cfg.Debug)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return v.WriteConfig()
}
