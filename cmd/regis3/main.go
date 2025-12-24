// Package main is the entry point for the regis3 CLI.
package main

import (
	"fmt"
	"os"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/okto-digital/regis3/internal/resolver"
)

// Build-time variables (set via ldflags)
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	fmt.Printf("regis3 %s (built %s)\n", version, buildTime)
	fmt.Println("Configuration registry for LLM assistants")
	fmt.Println()

	// Demo: Build the sample registry
	registryPath := "registry"
	if len(os.Args) > 1 {
		registryPath = os.Args[1]
	}

	fmt.Printf("Building registry: %s\n", registryPath)
	fmt.Println("─────────────────────────────────────────")

	result, err := registry.BuildRegistry(registryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Show stats
	fmt.Println()
	fmt.Println("📊 Stats:")
	fmt.Printf("   Skills:       %d\n", result.Manifest.Stats.Skills)
	fmt.Printf("   Subagents:    %d\n", result.Manifest.Stats.Subagents)
	fmt.Printf("   Philosophies: %d\n", result.Manifest.Stats.Philosophies)
	fmt.Printf("   Stacks:       %d\n", result.Manifest.Stats.Stacks)
	fmt.Printf("   Total:        %d\n", result.Manifest.Stats.Total())
	fmt.Printf("   Duration:     %v\n", result.Duration)

	// Show items
	fmt.Println()
	fmt.Println("📦 Items found:")
	for key, item := range result.Manifest.Items {
		fmt.Printf("   %s\n", key)
		fmt.Printf("      desc: %s\n", item.Desc)
		if len(item.Deps) > 0 {
			fmt.Printf("      deps: %v\n", item.Deps)
		}
	}

	// Show validation results
	if len(result.Validation.Issues) > 0 {
		fmt.Println()
		fmt.Println("⚠️  Validation issues:")
		for _, issue := range result.Validation.Issues {
			fmt.Printf("   [%s] %s: %s\n", issue.Severity, issue.Path, issue.Message)
		}
	}

	// Show skipped files
	if len(result.Skipped) > 0 {
		fmt.Println()
		fmt.Printf("⏭️  Skipped %d files (no regis3 block)\n", len(result.Skipped))
	}

	// Final status
	fmt.Println()
	if result.Validation.HasErrors() {
		fmt.Println("❌ Build failed with errors")
		os.Exit(1)
	}
	fmt.Println("✅ Build successful! Manifest saved to .build/manifest.json")

	// Demo: Dependency Resolution (Phase 3)
	fmt.Println()
	fmt.Println("🔗 Dependency Resolution:")
	fmt.Println("─────────────────────────────────────────")

	r := resolver.NewResolver(result.Manifest)

	// Check for cycles
	if r.HasCycle() {
		cycle, _ := r.FindCycle()
		fmt.Printf("   ⚠️  Circular dependency detected: %v\n", cycle)
	} else {
		fmt.Println("   ✓ No circular dependencies")
	}

	// Validate dependencies
	valResult := r.Validate()
	if len(valResult.MissingDeps) > 0 {
		fmt.Println("   ⚠️  Missing dependencies:")
		for _, m := range valResult.MissingDeps {
			fmt.Printf("      - %s\n", m)
		}
	} else {
		fmt.Println("   ✓ All dependencies resolved")
	}

	// Show installation order for stack:base
	fmt.Println()
	fmt.Println("📋 Installation order for stack:base:")
	order, err := r.GetInstallOrder([]string{"stack:base"})
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		for i, id := range order {
			fmt.Printf("   %d. %s\n", i+1, id)
		}
	}

	// Show dependency info for a specific item
	fmt.Println()
	fmt.Println("🔍 Dependency info for skill:testing:")
	info, err := r.GetDependencyInfo("skill:testing")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Direct deps:  %v\n", info.DirectDeps)
		fmt.Printf("   All deps:     %v\n", info.AllDeps)
		fmt.Printf("   Dependents:   %v\n", info.Dependents)
	}

	// Demo: Output Formatting (Phase 4)
	fmt.Println()
	fmt.Println("🎨 Output Formatting Demo:")
	fmt.Println("─────────────────────────────────────────")

	// Pretty output
	pretty := output.New(output.FormatPretty, nil)
	pretty.Success("Pretty format: Human-friendly with colors")
	pretty.Info("Supports icons and styled text")
	pretty.Warning("Warnings stand out visually")

	// Table demo
	fmt.Println()
	pretty.Table(
		[]string{"Type", "Name", "Description"},
		[][]string{
			{"skill", "git-conventions", "Git workflow conventions"},
			{"skill", "testing", "Testing best practices"},
			{"subagent", "architect", "Architecture planning"},
		},
	)

	fmt.Println()
	fmt.Println("Available formats: pretty (default), json, quiet")
}
