package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestBuilder_Build(t *testing.T) {
	builder := NewManifestBuilder("../../registry")

	manifest, valResult, err := builder.Build()
	require.NoError(t, err)

	// Should have found items
	assert.NotEmpty(t, manifest.Items)

	// Should have stats
	assert.Greater(t, manifest.Stats.Total(), 0)

	// Should not have validation errors
	assert.False(t, valResult.HasErrors(), "sample registry should be valid: %v", valResult.Errors())
}

func TestManifestBuilder_BuildAndSave(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a valid registry file
	skillsDir := filepath.Join(tmpDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	validContent := `---
regis3:
  type: skill
  name: test-skill
  desc: A test skill for testing purposes
  tags:
    - test
---
# Test Skill

This is test content.
`
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "test.md"), []byte(validContent), 0644))

	// Build and save
	builder := NewManifestBuilder(tmpDir)
	manifest, valResult, err := builder.BuildAndSave()
	require.NoError(t, err)

	// Should not have errors
	assert.False(t, valResult.HasErrors())

	// Manifest should be saved
	manifestPath := builder.ManifestPath()
	assert.FileExists(t, manifestPath)

	// Should be able to load it back
	loaded, err := LoadManifest(manifestPath)
	require.NoError(t, err)

	assert.Equal(t, manifest.Version, loaded.Version)
	assert.Len(t, loaded.Items, 1)
	assert.Contains(t, loaded.Items, "skill:test-skill")
}

func TestManifestBuilder_DoesNotSaveWithErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create an invalid registry file (missing required fields)
	invalidContent := `---
regis3:
  type: skill
---
Content without name or desc.
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "invalid.md"), []byte(invalidContent), 0644))

	builder := NewManifestBuilder(tmpDir)
	_, valResult, err := builder.BuildAndSave()
	require.NoError(t, err)

	// Should have errors
	assert.True(t, valResult.HasErrors())

	// Manifest should NOT be saved
	assert.False(t, ManifestExists(tmpDir))
}

func TestLoadManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create build dir and manifest
	buildDir := filepath.Join(tmpDir, ".build")
	require.NoError(t, os.MkdirAll(buildDir, 0755))

	manifestContent := `{
  "version": "1.0.0",
  "generated": "2025-01-01T00:00:00Z",
  "registry_path": "/test/registry",
  "items": {
    "skill:test": {
      "type": "skill",
      "name": "test",
      "desc": "Test skill",
      "source": "test.md"
    }
  },
  "stats": {
    "skills": 1
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, "manifest.json"), []byte(manifestContent), 0644))

	manifest, err := LoadManifestFromRegistry(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", manifest.Version)
	assert.Len(t, manifest.Items, 1)
	assert.Equal(t, "test", manifest.Items["skill:test"].Name)
}

func TestManifestExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Should not exist initially
	assert.False(t, ManifestExists(tmpDir))

	// Create manifest
	buildDir := filepath.Join(tmpDir, ".build")
	require.NoError(t, os.MkdirAll(buildDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, "manifest.json"), []byte("{}"), 0644))

	// Should exist now
	assert.True(t, ManifestExists(tmpDir))
}

func TestBuildRegistry(t *testing.T) {
	result, err := BuildRegistry("../../registry")
	require.NoError(t, err)

	// Should have found items
	assert.NotEmpty(t, result.Manifest.Items)

	// Should have computed stats
	assert.Greater(t, result.Manifest.Stats.Total(), 0)

	// Duration should be measured
	assert.Greater(t, result.Duration, int64(0))

	// No validation errors in sample registry
	assert.False(t, result.Validation.HasErrors())
}

func TestManifest_GetItem(t *testing.T) {
	manifest := NewManifest("/test")

	item := &Item{
		Regis3Meta: Regis3Meta{
			Type: "skill",
			Name: "test",
			Desc: "Test skill",
		},
		Source: "test.md",
	}

	manifest.AddItem(item)

	// Should find item
	found, ok := manifest.GetItem("skill:test")
	assert.True(t, ok)
	assert.Equal(t, "test", found.Name)

	// Should not find non-existent item
	_, ok = manifest.GetItem("skill:nonexistent")
	assert.False(t, ok)
}

func TestManifest_ComputeStats(t *testing.T) {
	manifest := NewManifest("/test")

	// Add various items
	items := []*Item{
		{Regis3Meta: Regis3Meta{Type: "skill", Name: "s1", Desc: "Skill 1"}, Source: "s1.md"},
		{Regis3Meta: Regis3Meta{Type: "skill", Name: "s2", Desc: "Skill 2"}, Source: "s2.md"},
		{Regis3Meta: Regis3Meta{Type: "subagent", Name: "a1", Desc: "Agent 1"}, Source: "a1.md"},
		{Regis3Meta: Regis3Meta{Type: "philosophy", Name: "p1", Desc: "Philosophy 1"}, Source: "p1.md"},
		{Regis3Meta: Regis3Meta{Type: "stack", Name: "st1", Desc: "Stack 1"}, Source: "st1.md"},
	}

	for _, item := range items {
		manifest.AddItem(item)
	}

	manifest.ComputeStats()

	assert.Equal(t, 2, manifest.Stats.Skills)
	assert.Equal(t, 1, manifest.Stats.Subagents)
	assert.Equal(t, 1, manifest.Stats.Philosophies)
	assert.Equal(t, 1, manifest.Stats.Stacks)
	assert.Equal(t, 5, manifest.Stats.Total())
}

func TestBuildRegistry_Integration(t *testing.T) {
	// Create temp directory with a complete registry
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create skills
	skillsDir := filepath.Join(tmpDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	baseSkill := `---
regis3:
  type: skill
  name: base
  desc: Base skill that others depend on
  tags:
    - base
---
# Base Skill
`
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "base.md"), []byte(baseSkill), 0644))

	dependentSkill := `---
regis3:
  type: skill
  name: dependent
  desc: Skill that depends on base
  deps:
    - skill:base
  tags:
    - dependent
---
# Dependent Skill
`
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "dependent.md"), []byte(dependentSkill), 0644))

	// Create philosophy
	philDir := filepath.Join(tmpDir, "philosophies")
	require.NoError(t, os.MkdirAll(philDir, 0755))

	philosophy := `---
regis3:
  type: philosophy
  name: test-philosophy
  desc: A test philosophy for testing
  order: 10
  tags:
    - test
---
# Test Philosophy
`
	require.NoError(t, os.WriteFile(filepath.Join(philDir, "test.md"), []byte(philosophy), 0644))

	// Build registry
	result, err := BuildRegistry(tmpDir)
	require.NoError(t, err)

	// Verify results
	assert.Len(t, result.Manifest.Items, 3)
	assert.Equal(t, 2, result.Manifest.Stats.Skills)
	assert.Equal(t, 1, result.Manifest.Stats.Philosophies)
	assert.False(t, result.Validation.HasErrors())

	// Manifest should be saved
	assert.True(t, ManifestExists(tmpDir))

	// Load and verify
	loaded, err := LoadManifestFromRegistry(tmpDir)
	require.NoError(t, err)
	assert.Len(t, loaded.Items, 3)
}
