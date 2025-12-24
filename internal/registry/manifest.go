package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// ManifestVersion is the current manifest format version.
	ManifestVersion = "1.0.0"

	// DefaultBuildDir is the directory where manifest is built.
	DefaultBuildDir = ".build"

	// DefaultManifestFile is the manifest filename.
	DefaultManifestFile = "manifest.json"
)

// ManifestBuilder builds a manifest from scanned items.
type ManifestBuilder struct {
	RegistryPath string
}

// NewManifestBuilder creates a new manifest builder.
func NewManifestBuilder(registryPath string) *ManifestBuilder {
	return &ManifestBuilder{RegistryPath: registryPath}
}

// Build scans the registry, validates items, and builds a manifest.
func (b *ManifestBuilder) Build() (*Manifest, *ValidationResult, error) {
	// Scan registry
	scanner := NewScanner(b.RegistryPath)
	scanResult, err := scanner.Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan registry: %w", err)
	}

	// Validate items
	validator := NewValidator(b.RegistryPath)
	valResult := validator.ValidateItems(scanResult.Items)

	// Build manifest even if there are warnings (but not errors)
	manifest := NewManifest(b.RegistryPath)

	for _, item := range scanResult.Items {
		manifest.AddItem(item)
	}

	manifest.ComputeStats()

	return manifest, valResult, nil
}

// BuildAndSave builds the manifest and saves it to the .build directory.
func (b *ManifestBuilder) BuildAndSave() (*Manifest, *ValidationResult, error) {
	manifest, valResult, err := b.Build()
	if err != nil {
		return nil, nil, err
	}

	// Don't save if there are errors
	if valResult.HasErrors() {
		return manifest, valResult, nil
	}

	// Save manifest
	if err := b.Save(manifest); err != nil {
		return manifest, valResult, fmt.Errorf("failed to save manifest: %w", err)
	}

	return manifest, valResult, nil
}

// Save writes the manifest to the .build directory.
func (b *ManifestBuilder) Save(manifest *Manifest) error {
	buildDir := filepath.Join(b.RegistryPath, DefaultBuildDir)

	// Create .build directory if it doesn't exist
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Marshal manifest to JSON
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	manifestPath := filepath.Join(buildDir, DefaultManifestFile)
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// ManifestPath returns the path to the manifest file.
func (b *ManifestBuilder) ManifestPath() string {
	return filepath.Join(b.RegistryPath, DefaultBuildDir, DefaultManifestFile)
}

// LoadManifest loads a manifest from a file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// LoadManifestFromRegistry loads the manifest from a registry's .build directory.
func LoadManifestFromRegistry(registryPath string) (*Manifest, error) {
	manifestPath := filepath.Join(registryPath, DefaultBuildDir, DefaultManifestFile)
	return LoadManifest(manifestPath)
}

// ManifestExists checks if a manifest exists for the registry.
func ManifestExists(registryPath string) bool {
	manifestPath := filepath.Join(registryPath, DefaultBuildDir, DefaultManifestFile)
	_, err := os.Stat(manifestPath)
	return err == nil
}

// BuildResult contains the complete result of building the registry.
type BuildResult struct {
	Manifest   *Manifest
	Validation *ValidationResult
	ScanErrors []ScanError
	Skipped    []string
	Duration   time.Duration
}

// BuildRegistry performs a complete build of the registry.
func BuildRegistry(registryPath string) (*BuildResult, error) {
	start := time.Now()

	// Scan
	scanner := NewScanner(registryPath)
	scanResult, err := scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("failed to scan registry: %w", err)
	}

	// Validate
	validator := NewValidator(registryPath)
	valResult := validator.ValidateItems(scanResult.Items)

	// Build manifest
	manifest := NewManifest(registryPath)
	for _, item := range scanResult.Items {
		manifest.AddItem(item)
	}
	manifest.ComputeStats()

	// Save manifest if no errors
	if !valResult.HasErrors() {
		builder := NewManifestBuilder(registryPath)
		if err := builder.Save(manifest); err != nil {
			return nil, fmt.Errorf("failed to save manifest: %w", err)
		}
	}

	return &BuildResult{
		Manifest:   manifest,
		Validation: valResult,
		ScanErrors: scanResult.Errors,
		Skipped:    scanResult.Skipped,
		Duration:   time.Since(start),
	}, nil
}
