package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/okto-digital/regis3/internal/config"
	"github.com/spf13/cobra"
)

var (
	initNonInteractive bool
	initRegistryPath   string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize regis3 configuration",
	Long: `Sets up regis3 for first use by creating configuration and registry directories.

In interactive mode, prompts for:
- Registry location (default: ~/.regis3/registry)
- Whether to initialize as git repo

Use --yes for non-interactive mode with defaults.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initNonInteractive, "yes", "y", false, "Accept defaults without prompting")
	initCmd.Flags().StringVar(&initRegistryPath, "registry", "", "Registry path (default: ~/.regis3/registry)")
	rootCmd.AddCommand(initCmd)
}

func runInit() error {
	paths, err := config.NewPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	// Check if already initialized
	if _, err := os.Stat(paths.ConfigFile); err == nil {
		fmt.Println("regis3 is already initialized.")
		fmt.Printf("Config: %s\n", paths.ConfigFile)
		fmt.Printf("Registry: %s\n", paths.RegistryDir)
		return nil
	}

	registryPath := initRegistryPath
	if registryPath == "" {
		registryPath = paths.RegistryDir
	}

	// Interactive mode
	if !initNonInteractive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Welcome to regis3!")
		fmt.Println()

		// Ask for registry path
		fmt.Printf("Registry location [%s]: ", registryPath)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			registryPath = expandPath(input)
		}
	}

	// Create directories
	fmt.Printf("Creating registry at: %s\n", registryPath)

	if err := os.MkdirAll(registryPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating registry directory: %v\n", err)
		return err
	}

	// Create standard subdirectories
	subdirs := []string{
		"skills",
		"agents",
		"commands",
		"philosophies",
		"docs",
		"prompts",
		".build",
	}

	for _, subdir := range subdirs {
		dir := filepath.Join(registryPath, subdir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", subdir, err)
			return err
		}
	}

	// Create config
	newCfg := &config.Config{
		RegistryPath:  registryPath,
		DefaultTarget: "claude",
		OutputFormat:  "pretty",
		Debug:         false,
	}

	// Ensure config directory exists
	if err := os.MkdirAll(paths.ConfigDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		return err
	}

	if err := config.Save(newCfg, paths.ConfigFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		return err
	}

	fmt.Println()
	fmt.Println("regis3 initialized successfully!")
	fmt.Printf("  Config: %s\n", paths.ConfigFile)
	fmt.Printf("  Registry: %s\n", registryPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add markdown files with regis3 frontmatter to the registry")
	fmt.Println("  2. Run 'regis3 build' to build the manifest")
	fmt.Println("  3. Run 'regis3 list' to see available items")
	fmt.Println("  4. Run 'regis3 add <type:name>' to install items")

	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}
	return path
}
