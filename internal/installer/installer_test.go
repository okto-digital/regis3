package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/okto-digital/regis3/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultClaudeTarget(t *testing.T) {
	target := DefaultClaudeTarget()

	assert.Equal(t, "claude", target.Name)
	assert.Equal(t, ".claude", target.BaseDir)
	assert.Equal(t, "CLAUDE.md", target.MergeFile)

	// Check paths
	path, err := target.GetPath("skill", "git-conventions")
	require.NoError(t, err)
	assert.Equal(t, ".claude/skills/git-conventions/SKILL.md", path)

	path, err = target.GetPath("subagent", "architect")
	require.NoError(t, err)
	assert.Equal(t, ".claude/agents/architect.md", path)

	path, err = target.GetPath("command", "deploy")
	require.NoError(t, err)
	assert.Equal(t, ".claude/commands/deploy.md", path)
}

func TestTarget_GetPath(t *testing.T) {
	target := DefaultClaudeTarget()

	tests := []struct {
		itemType string
		name     string
		expected string
	}{
		{"skill", "testing", ".claude/skills/testing/SKILL.md"},
		{"subagent", "coder", ".claude/agents/coder.md"},
		{"command", "test", ".claude/commands/test.md"},
		{"mcp", "github", ".claude/mcp/github.json"},
		{"script", "setup", ".claude/scripts/setup.sh"},
		{"doc", "readme", ".claude/docs/readme.md"},
		{"hook", "pre-commit", ".claude/hooks/pre-commit.md"},
		{"prompt", "review", ".claude/prompts/review.md"},
	}

	for _, tt := range tests {
		t.Run(tt.itemType+":"+tt.name, func(t *testing.T) {
			path, err := target.GetPath(tt.itemType, tt.name)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, path)
		})
	}
}

func TestTarget_IsMergeType(t *testing.T) {
	target := DefaultClaudeTarget()

	assert.True(t, target.IsMergeType("philosophy"))
	assert.True(t, target.IsMergeType("project"))
	assert.True(t, target.IsMergeType("ruleset"))
	assert.False(t, target.IsMergeType("skill"))
	assert.False(t, target.IsMergeType("subagent"))
}

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "with frontmatter",
			input: `---
title: Test
---
# Content here`,
			expected: "# Content here",
		},
		{
			name:     "without frontmatter",
			input:    "# Just content",
			expected: "# Just content",
		},
		{
			name: "empty frontmatter",
			input: `---
---
Content`,
			expected: "Content",
		},
		{
			name: "complex frontmatter",
			input: `---
regis3:
  type: skill
  name: test
  desc: Test skill
---
# Skill Content

More content here.`,
			expected: `# Skill Content

More content here.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripFrontmatter(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransformer_Transform(t *testing.T) {
	target := DefaultClaudeTarget()
	transformer := NewTransformer(target)

	item := &registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type: "skill",
			Name: "test-skill",
			Desc: "A test skill",
		},
		Content: `---
regis3:
  type: skill
  name: test-skill
---
# Test Skill

This is the content.`,
	}

	result, err := transformer.Transform(item)
	require.NoError(t, err)

	assert.Equal(t, `# Test Skill

This is the content.`, result)
}

func TestMergeContent(t *testing.T) {
	mc := NewMergeContent()

	// Add philosophy
	mc.Add(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type:  "philosophy",
			Name:  "clean-code",
			Order: 10,
		},
	}, "Clean code principles here.")

	// Add another philosophy
	mc.Add(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type:  "philosophy",
			Name:  "kiss",
			Order: 20,
		},
	}, "Keep it simple.")

	// Add project
	mc.Add(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type:  "project",
			Name:  "my-project",
			Order: 5,
		},
	}, "Project description.")

	assert.True(t, mc.HasContent())

	result := mc.Generate()

	// Should have sections in order: project, philosophy, ruleset
	assert.Contains(t, result, "## Project")
	assert.Contains(t, result, "## Philosophy")
	assert.Contains(t, result, "Clean code principles")
	assert.Contains(t, result, "Keep it simple")
	assert.Contains(t, result, "Project description")
}

func TestUpdateExistingFile(t *testing.T) {
	t.Run("empty existing", func(t *testing.T) {
		result := UpdateExistingFile("", "New content")
		assert.Contains(t, result, "<!-- regis3:start -->")
		assert.Contains(t, result, "New content")
		assert.Contains(t, result, "<!-- regis3:end -->")
	})

	t.Run("with existing content", func(t *testing.T) {
		existing := "# My Project\n\nUser content here."
		result := UpdateExistingFile(existing, "Managed content")
		assert.Contains(t, result, "# My Project")
		assert.Contains(t, result, "User content here")
		assert.Contains(t, result, "Managed content")
	})

	t.Run("replace existing managed section", func(t *testing.T) {
		existing := `# Header

<!-- regis3:start -->
Old managed content
<!-- regis3:end -->

# Footer`
		result := UpdateExistingFile(existing, "New managed content")
		assert.Contains(t, result, "# Header")
		assert.Contains(t, result, "New managed content")
		assert.NotContains(t, result, "Old managed content")
		assert.Contains(t, result, "# Footer")
	})
}

func TestExtractManagedContent(t *testing.T) {
	content := `# Header

<!-- regis3:start -->
Managed content here
<!-- regis3:end -->

# Footer`

	result := ExtractManagedContent(content)
	assert.Equal(t, "Managed content here", result)
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git-conventions", "git-conventions"},
		{"My Skill", "my-skill"},
		{"test_skill", "test-skill"},
		{"Test@Skill#123", "test-skill-123"},
		{"--double--dash--", "double-dash"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTracker(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("new tracker", func(t *testing.T) {
		tracker := NewTracker(tmpDir, "claude")
		assert.Equal(t, 0, tracker.Count())
		assert.False(t, tracker.IsInstalled("skill:test"))
	})

	t.Run("mark installed", func(t *testing.T) {
		tracker := NewTracker(tmpDir, "claude")
		tracker.MarkInstalled("skill:test", "skill", "test", ".claude/skills/test/SKILL.md", false)

		assert.True(t, tracker.IsInstalled("skill:test"))
		assert.Equal(t, 1, tracker.Count())

		installed := tracker.GetInstalled("skill:test")
		require.NotNil(t, installed)
		assert.Equal(t, "skill:test", installed.ID)
		assert.Equal(t, ".claude/skills/test/SKILL.md", installed.InstalledPath)
	})

	t.Run("save and load", func(t *testing.T) {
		tracker := NewTracker(tmpDir, "claude")
		tracker.MarkInstalled("skill:test", "skill", "test", ".claude/skills/test/SKILL.md", false)
		tracker.SetSourceHash("skill:test", "abc123")

		err := tracker.Save()
		require.NoError(t, err)

		// Load new tracker
		loaded, err := LoadTracker(tmpDir, "claude")
		require.NoError(t, err)

		assert.True(t, loaded.IsInstalled("skill:test"))
		assert.Equal(t, "abc123", loaded.GetInstalled("skill:test").SourceHash)
	})

	t.Run("needs update", func(t *testing.T) {
		tracker := NewTracker(tmpDir, "claude")
		tracker.MarkInstalled("skill:test", "skill", "test", ".claude/skills/test/SKILL.md", false)
		tracker.SetSourceHash("skill:test", "hash1")

		assert.False(t, tracker.NeedsUpdate("skill:test", "hash1"))
		assert.True(t, tracker.NeedsUpdate("skill:test", "hash2"))
		assert.True(t, tracker.NeedsUpdate("skill:other", "any"))
	})

	t.Run("uninstall", func(t *testing.T) {
		tracker := NewTracker(tmpDir, "claude")
		tracker.MarkInstalled("skill:test", "skill", "test", ".claude/skills/test/SKILL.md", false)

		assert.True(t, tracker.IsInstalled("skill:test"))
		tracker.MarkUninstalled("skill:test")
		assert.False(t, tracker.IsInstalled("skill:test"))
	})

	t.Run("list by type", func(t *testing.T) {
		tracker := NewTracker(tmpDir, "claude")
		tracker.MarkInstalled("skill:a", "skill", "a", "", false)
		tracker.MarkInstalled("skill:b", "skill", "b", "", false)
		tracker.MarkInstalled("subagent:c", "subagent", "c", "", false)

		skills := tracker.ListInstalledByType("skill")
		assert.Len(t, skills, 2)

		subagents := tracker.ListInstalledByType("subagent")
		assert.Len(t, subagents, 1)
	})
}

func TestInstaller_Install(t *testing.T) {
	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	registryDir := filepath.Join(tmpDir, "registry")
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(registryDir, 0755))
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	// Create a simple manifest
	manifest := registry.NewManifest(registryDir)
	manifest.AddItem(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type: "skill",
			Name: "test-skill",
			Desc: "A test skill",
		},
		Content: `---
regis3:
  type: skill
  name: test-skill
---
# Test Skill

Content here.`,
		Source: "skills/test-skill.md",
	})

	// Create installer
	target := DefaultClaudeTarget()
	installer, err := NewInstaller(projectDir, registryDir, target)
	require.NoError(t, err)

	// Install
	result, err := installer.Install(manifest, []string{"skill:test-skill"})
	require.NoError(t, err)

	assert.Len(t, result.Installed, 1)
	assert.Contains(t, result.Installed, "skill:test-skill")
	assert.Empty(t, result.Errors)

	// Verify file was created
	skillPath := filepath.Join(projectDir, ".claude", "skills", "test-skill", "SKILL.md")
	assert.FileExists(t, skillPath)

	// Read and verify content
	content, err := os.ReadFile(skillPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Test Skill")
	assert.NotContains(t, string(content), "regis3:")

	// Verify tracker
	assert.True(t, installer.Tracker.IsInstalled("skill:test-skill"))
}

func TestInstaller_InstallWithDependencies(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	registryDir := filepath.Join(tmpDir, "registry")
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(registryDir, 0755))
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	manifest := registry.NewManifest(registryDir)
	manifest.AddItem(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type: "skill",
			Name: "base",
			Desc: "Base skill",
		},
		Content: "# Base\n\nBase content.",
		Source:  "skills/base.md",
	})
	manifest.AddItem(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type: "skill",
			Name: "dependent",
			Desc: "Depends on base",
			Deps: []string{"skill:base"},
		},
		Content: "# Dependent\n\nDependent content.",
		Source:  "skills/dependent.md",
	})

	target := DefaultClaudeTarget()
	installer, err := NewInstaller(projectDir, registryDir, target)
	require.NoError(t, err)

	// Install just the dependent - should also install base
	result, err := installer.Install(manifest, []string{"skill:dependent"})
	require.NoError(t, err)

	assert.Len(t, result.Installed, 2)
	assert.Contains(t, result.Installed, "skill:base")
	assert.Contains(t, result.Installed, "skill:dependent")
}

func TestInstaller_InstallMergeType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	registryDir := filepath.Join(tmpDir, "registry")
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(registryDir, 0755))
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	manifest := registry.NewManifest(registryDir)
	manifest.AddItem(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type:  "philosophy",
			Name:  "clean-code",
			Desc:  "Clean code principles",
			Order: 10,
		},
		Content: "# Clean Code\n\nWrite clean code.",
		Source:  "philosophies/clean-code.md",
	})

	target := DefaultClaudeTarget()
	installer, err := NewInstaller(projectDir, registryDir, target)
	require.NoError(t, err)

	result, err := installer.Install(manifest, []string{"philosophy:clean-code"})
	require.NoError(t, err)

	assert.Len(t, result.MergedItems, 1)
	assert.Contains(t, result.MergedItems, "philosophy:clean-code")

	// Verify CLAUDE.md was created
	claudemd := filepath.Join(projectDir, "CLAUDE.md")
	assert.FileExists(t, claudemd)

	content, err := os.ReadFile(claudemd)
	require.NoError(t, err)
	assert.Contains(t, string(content), "regis3:start")
	assert.Contains(t, string(content), "Clean Code")
	assert.Contains(t, string(content), "regis3:end")
}

func TestInstaller_DryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	registryDir := filepath.Join(tmpDir, "registry")
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(registryDir, 0755))
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	manifest := registry.NewManifest(registryDir)
	manifest.AddItem(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type: "skill",
			Name: "test",
			Desc: "Test",
		},
		Content: "# Test",
		Source:  "skills/test.md",
	})

	target := DefaultClaudeTarget()
	installer, err := NewInstaller(projectDir, registryDir, target)
	require.NoError(t, err)
	installer.DryRun = true

	result, err := installer.Install(manifest, []string{"skill:test"})
	require.NoError(t, err)

	assert.Len(t, result.Installed, 1)

	// File should NOT be created in dry run
	skillPath := filepath.Join(projectDir, ".claude", "skills", "test", "SKILL.md")
	assert.NoFileExists(t, skillPath)

	// Tracker should NOT be updated
	assert.False(t, installer.Tracker.IsInstalled("skill:test"))
}

func TestInstaller_Uninstall(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	registryDir := filepath.Join(tmpDir, "registry")
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(registryDir, 0755))
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	manifest := registry.NewManifest(registryDir)
	manifest.AddItem(&registry.Item{
		Regis3Meta: registry.Regis3Meta{
			Type: "skill",
			Name: "test",
			Desc: "Test",
		},
		Content: "# Test",
		Source:  "skills/test.md",
	})

	target := DefaultClaudeTarget()
	installer, err := NewInstaller(projectDir, registryDir, target)
	require.NoError(t, err)

	// First install
	_, err = installer.Install(manifest, []string{"skill:test"})
	require.NoError(t, err)

	skillPath := filepath.Join(projectDir, ".claude", "skills", "test", "SKILL.md")
	assert.FileExists(t, skillPath)

	// Now uninstall
	result, err := installer.Uninstall([]string{"skill:test"})
	require.NoError(t, err)

	assert.Len(t, result.Uninstalled, 1)
	assert.Contains(t, result.Uninstalled, "skill:test")
	assert.NoFileExists(t, skillPath)
	assert.False(t, installer.Tracker.IsInstalled("skill:test"))
}

func TestLoadTarget(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "target-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := `
name: test-target
description: Test target
version: "1.0.0"
base_dir: .test
merge_file: TEST.md
paths:
  skill:
    dir: skills
    pattern: "{name}.md"
`
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	target, err := LoadTarget(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "test-target", target.Name)
	assert.Equal(t, ".test", target.BaseDir)
	assert.Equal(t, "TEST.md", target.MergeFile)

	path, err := target.GetPath("skill", "my-skill")
	require.NoError(t, err)
	assert.Equal(t, ".test/skills/my-skill.md", path)
}
