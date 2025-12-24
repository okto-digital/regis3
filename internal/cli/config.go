package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/okto-digital/regis3/internal/config"
	"github.com/okto-digital/regis3/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage regis3 configuration",
	Long: `View and modify regis3 configuration settings.

Examples:
  regis3 config                    # Show current config
  regis3 config get registry       # Get specific setting
  regis3 config set registry ~/my-registry`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigShow()
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing key\n\nUsage: regis3 config get <key>\n\nKeys: registry, target")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigGet(args[0])
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing key and value\n\nUsage: regis3 config set <key> <value>\n\nExample: regis3 config set registry ~/my-registry")
		}
		if len(args) == 1 {
			return fmt.Errorf("missing value for '%s'\n\nUsage: regis3 config set <key> <value>", args[0])
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSet(args[0], args[1])
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigPath()
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow() error {
	configPath := config.DefaultConfigPath()
	debugf("Config path: %s", configPath)

	settings := make(map[string]string)

	if cfg != nil {
		settings["registry"] = cfg.RegistryPath
		settings["default_target"] = cfg.DefaultTarget
	} else {
		settings["registry"] = "(not set)"
		settings["default_target"] = "(not set)"
	}

	resp := output.NewResponseBuilder("config").
		WithSuccess(true).
		WithData(output.ConfigData{
			Path:     configPath,
			Settings: settings,
		})

	writer.Write(resp.Build())
	return nil
}

func runConfigGet(key string) error {
	if cfg == nil {
		writer.Error("No configuration file found")
		return fmt.Errorf("no config")
	}

	var value string
	switch key {
	case "registry", "registry_path":
		value = cfg.RegistryPath
	case "target", "default_target":
		value = cfg.DefaultTarget
	default:
		writer.Error(fmt.Sprintf("Unknown config key: %s", key))
		return fmt.Errorf("unknown key: %s", key)
	}

	resp := output.NewResponseBuilder("config").
		WithSuccess(true).
		WithData(map[string]string{key: value})

	writer.Write(resp.Build())
	return nil
}

func runConfigSet(key, value string) error {
	configPath := config.DefaultConfigPath()
	debugf("Setting %s=%s in %s", key, value, configPath)

	// Load or create config
	var c *config.Config
	if cfg != nil {
		c = cfg
	} else {
		c = &config.Config{}
	}

	// Set the value
	switch key {
	case "registry", "registry_path":
		// Expand path
		if value[0] == '~' {
			home, _ := os.UserHomeDir()
			value = filepath.Join(home, value[1:])
		}
		absPath, err := filepath.Abs(value)
		if err == nil {
			value = absPath
		}
		c.RegistryPath = value
	case "target", "default_target":
		c.DefaultTarget = value
	default:
		writer.Error(fmt.Sprintf("Unknown config key: %s", key))
		return fmt.Errorf("unknown key: %s", key)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		writer.Error(fmt.Sprintf("Failed to create config directory: %s", err.Error()))
		return err
	}

	// Write config
	data, err := yaml.Marshal(c)
	if err != nil {
		writer.Error(fmt.Sprintf("Failed to marshal config: %s", err.Error()))
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		writer.Error(fmt.Sprintf("Failed to write config: %s", err.Error()))
		return err
	}

	resp := output.NewResponseBuilder("config").
		WithSuccess(true).
		WithInfo("Set %s = %s", key, value)

	writer.Write(resp.Build())
	return nil
}

func runConfigPath() error {
	configPath := config.DefaultConfigPath()

	resp := output.NewResponseBuilder("config").
		WithSuccess(true).
		WithData(map[string]string{"path": configPath})

	writer.Write(resp.Build())
	return nil
}
