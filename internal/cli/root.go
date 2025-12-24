package cli

import (
	"fmt"
	"os"

	"github.com/okto-digital/regis3/internal/config"
	"github.com/okto-digital/regis3/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	formatFlag   string
	debugFlag    bool
	configFlag   string
	registryFlag string

	// Global state
	cfg    *config.Config
	writer output.Writer
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "regis3",
	Short: "Registry manager for LLM assistant configurations",
	Long: `regis3 is a CLI tool for managing a registry of skills, agents,
commands, and other configurations for LLM coding assistants.

It supports multiple targets (Claude Code, Cursor, etc.) and provides
dependency resolution, validation, and organized installation.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip setup for init command when no config exists
		if cmd.Name() == "init" {
			return nil
		}

		// Load config
		var err error
		cfg, err = loadConfig()
		if err != nil {
			// If config doesn't exist and not running init, suggest init
			if os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "No configuration found. Run 'regis3 init' to set up.")
				os.Exit(1)
			}
			return err
		}

		// Override registry path if flag provided
		if registryFlag != "" {
			cfg.RegistryPath = registryFlag
		}

		// Initialize output writer
		writer = createWriter()

		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "pretty", "Output format: pretty, json, quiet")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().StringVar(&configFlag, "config", "", "Config file path")
	rootCmd.PersistentFlags().StringVar(&registryFlag, "registry", "", "Override registry path")
}

// loadConfig loads the configuration.
func loadConfig() (*config.Config, error) {
	return config.Load(configFlag)
}

// createWriter creates an output writer based on flags.
func createWriter() output.Writer {
	format := output.FormatPretty
	switch formatFlag {
	case "json":
		format = output.FormatJSON
	case "quiet":
		format = output.FormatQuiet
	}
	return output.New(format, nil)
}

// getRegistryPath returns the registry path from config or flag.
func getRegistryPath() string {
	if registryFlag != "" {
		return registryFlag
	}
	if cfg != nil {
		return cfg.RegistryPath
	}
	return config.DefaultRegistryPath()
}

// debugf prints debug output if debug mode is enabled.
func debugf(format string, args ...interface{}) {
	if debugFlag {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
