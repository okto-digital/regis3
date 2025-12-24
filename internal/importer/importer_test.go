package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalScanner_Scan(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	file1 := filepath.Join(tmpDir, "doc1.md")
	file2 := filepath.Join(tmpDir, "subdir", "doc2.md")
	file3 := filepath.Join(tmpDir, "not-markdown.txt")

	require.NoError(t, os.WriteFile(file1, []byte("# Doc 1\n\nContent"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(file2, []byte("# Doc 2\n\nContent"), 0644))
	require.NoError(t, os.WriteFile(file3, []byte("Not markdown"), 0644))

	scanner := NewExternalScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 2, len(result.Files))
	assert.Empty(t, result.Errors)
}

func TestExternalScanner_SkipsHiddenDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create files in hidden directory
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	require.NoError(t, os.MkdirAll(hiddenDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(hiddenDir, "doc.md"), []byte("# Hidden"), 0644))

	// Create file in regular directory
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "visible.md"), []byte("# Visible"), 0644))

	scanner := NewExternalScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 1, len(result.Files))
	assert.Equal(t, "visible.md", filepath.Base(result.Files[0].Path))
}

func TestExternalScanner_DetectsFrontmatter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// File with regis3 frontmatter
	withRegis3 := `---
regis3:
  type: skill
  name: test
---
# Content`

	// File with other frontmatter
	withFrontmatter := `---
title: Test
---
# Content`

	// Plain markdown
	plainMd := `# Content

No frontmatter here.`

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "with-regis3.md"), []byte(withRegis3), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "with-frontmatter.md"), []byte(withFrontmatter), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "plain.md"), []byte(plainMd), 0644))

	scanner := NewExternalScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 3, len(result.Files))

	stats := result.Stats()
	assert.Equal(t, 3, stats.TotalFiles)
	assert.Equal(t, 1, stats.WithRegis3)
	assert.Equal(t, 1, stats.WithFrontmatter)
	assert.Equal(t, 1, stats.PlainMarkdown)
}

func TestExternalScanner_FilterByRegis3(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	withRegis3 := `---
regis3:
  type: skill
  name: test
---
# Content`

	withoutRegis3 := `# No frontmatter`

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "with.md"), []byte(withRegis3), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "without.md"), []byte(withoutRegis3), 0644))

	scanner := NewExternalScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	withRegis3Files := result.FilterByRegis3()
	assert.Equal(t, 1, len(withRegis3Files))

	withoutRegis3Files := result.FilterWithoutRegis3()
	assert.Equal(t, 1, len(withoutRegis3Files))
}

func TestClassifier_ClassifyWithRegis3(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	content := `---
regis3:
  type: skill
  name: git-conventions
  desc: Git workflow conventions
---
# Git Conventions`

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	classifier := NewClassifier()
	result, err := classifier.Classify(filePath)
	require.NoError(t, err)

	assert.True(t, result.HasValidRegis3)
	assert.Equal(t, "skill", result.SuggestedType)
	assert.Equal(t, "git-conventions", result.SuggestedName)
	assert.Equal(t, 100, result.Confidence)
}

func TestClassifier_ClassifyByDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		dir      string
		expected string
	}{
		{"skills", "skill"},
		{"agents", "subagent"},
		{"philosophies", "philosophy"},
		{"commands", "command"},
		{"prompts", "prompt"},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			dir := filepath.Join(tmpDir, tt.dir)
			require.NoError(t, os.MkdirAll(dir, 0755))

			filePath := filepath.Join(dir, "test.md")
			require.NoError(t, os.WriteFile(filePath, []byte("# Test"), 0644))

			classifier := NewClassifier()
			result, err := classifier.Classify(filePath)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, result.SuggestedType)
			assert.GreaterOrEqual(t, result.Confidence, 70)
		})
	}
}

func TestClassifier_ClassifyByContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "agent-like content",
			content:  "You are an expert developer. Your role is to help with code.",
			expected: "subagent",
		},
		{
			name:     "skill-like content",
			content:  "# Best Practices\n\n## Usage\n\nFollow these guidelines.",
			expected: "skill",
		},
		{
			name:     "philosophy-like content",
			content:  "# Clean Code Principles\n\nOur approach to software quality.",
			expected: "philosophy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.name+".md")
			require.NoError(t, os.WriteFile(filePath, []byte(tt.content), 0644))

			classifier := NewClassifier()
			result, err := classifier.Classify(filePath)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, result.SuggestedType)
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GitConventions", "git-conventions"},
		{"git_conventions", "git-conventions"},
		{"git conventions", "git-conventions"},
		{"Git-Conventions", "git-conventions"},
		{"mySkillName", "my-skill-name"},
		{"UPPERCASE", "uppercase"},
		{"with--dashes", "with-dashes"},
		{"-leading-trailing-", "leading-trailing"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toKebabCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClassifier_GenerateFrontmatter(t *testing.T) {
	classifier := NewClassifier()

	class := &Classification{
		SuggestedType: "skill",
		SuggestedName: "test-skill",
	}

	fm := classifier.GenerateFrontmatter(class, "A test skill")

	assert.Contains(t, fm, "regis3:")
	assert.Contains(t, fm, "type: skill")
	assert.Contains(t, fm, "name: test-skill")
	assert.Contains(t, fm, "desc: A test skill")
	assert.Contains(t, fm, "tags:")
	assert.Contains(t, fm, "- imported")
}

func TestImporter_ScanAndImport(t *testing.T) {
	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	externalDir := filepath.Join(tmpDir, "external")
	registryDir := filepath.Join(tmpDir, "registry")
	require.NoError(t, os.MkdirAll(externalDir, 0755))
	require.NoError(t, os.MkdirAll(registryDir, 0755))

	// Create file with regis3 metadata
	withRegis3 := `---
regis3:
  type: skill
  name: test-skill
  desc: A test skill
---
# Test Skill`

	// Create file without regis3 metadata
	withoutRegis3 := `# Plain Document

This is a plain document.`

	require.NoError(t, os.WriteFile(filepath.Join(externalDir, "with-regis3.md"), []byte(withRegis3), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(externalDir, "without-regis3.md"), []byte(withoutRegis3), 0644))

	importer := NewImporter(registryDir)
	result, err := importer.ScanAndImport(externalDir)
	require.NoError(t, err)

	// One should be imported directly, one staged
	assert.Len(t, result.Imported, 1)
	assert.Len(t, result.Staged, 1)
	assert.Empty(t, result.Errors)

	// Check files were created
	assert.FileExists(t, filepath.Join(registryDir, "skills", "test-skill.md"))
	assert.FileExists(t, filepath.Join(registryDir, "import", "without-regis3.md"))
}

func TestImporter_DryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	externalDir := filepath.Join(tmpDir, "external")
	registryDir := filepath.Join(tmpDir, "registry")
	require.NoError(t, os.MkdirAll(externalDir, 0755))
	require.NoError(t, os.MkdirAll(registryDir, 0755))

	content := `---
regis3:
  type: skill
  name: test
  desc: Test
---
# Test`

	require.NoError(t, os.WriteFile(filepath.Join(externalDir, "test.md"), []byte(content), 0644))

	importer := NewImporter(registryDir)
	importer.DryRun = true

	result, err := importer.ScanAndImport(externalDir)
	require.NoError(t, err)

	assert.Len(t, result.Imported, 1)

	// File should NOT be created in dry run
	assert.NoFileExists(t, filepath.Join(registryDir, "skills", "test.md"))
}

func TestImporter_ProcessStaging(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	registryDir := filepath.Join(tmpDir, "registry")
	importDir := filepath.Join(registryDir, "import")
	require.NoError(t, os.MkdirAll(importDir, 0755))

	// Create a file without regis3 in staging
	noRegis3 := `# Pending File

This needs a regis3 header.`

	// Create a file with regis3 in staging (ready to move)
	withRegis3 := `---
regis3:
  type: skill
  name: ready
  desc: Ready to move
---
# Ready`

	require.NoError(t, os.WriteFile(filepath.Join(importDir, "pending.md"), []byte(noRegis3), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(importDir, "ready.md"), []byte(withRegis3), 0644))

	importer := NewImporter(registryDir)
	result, err := importer.ProcessStaging()
	require.NoError(t, err)

	// One should be processed, one pending
	assert.Len(t, result.Processed, 1)
	assert.Len(t, result.Pending, 1)
	assert.Equal(t, "ready", result.Processed[0].Name)

	// Check file was moved
	assert.FileExists(t, filepath.Join(registryDir, "skills", "ready.md"))
	assert.NoFileExists(t, filepath.Join(importDir, "ready.md"))
	assert.FileExists(t, filepath.Join(importDir, "pending.md"))
}

func TestImporter_ListPending(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	registryDir := filepath.Join(tmpDir, "registry")
	importDir := filepath.Join(registryDir, "import")
	require.NoError(t, os.MkdirAll(importDir, 0755))

	// Create pending files
	require.NoError(t, os.WriteFile(filepath.Join(importDir, "file1.md"), []byte("# File 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(importDir, "file2.md"), []byte("# File 2"), 0644))

	importer := NewImporter(registryDir)
	pending, err := importer.ListPending()
	require.NoError(t, err)

	assert.Len(t, pending, 2)
}

func TestImporter_StagingExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	registryDir := filepath.Join(tmpDir, "registry")
	require.NoError(t, os.MkdirAll(registryDir, 0755))

	importer := NewImporter(registryDir)

	// No staging directory
	assert.False(t, importer.StagingExists())

	// Empty staging directory
	importDir := filepath.Join(registryDir, "import")
	require.NoError(t, os.MkdirAll(importDir, 0755))
	assert.False(t, importer.StagingExists())

	// Staging with files
	require.NoError(t, os.WriteFile(filepath.Join(importDir, "test.md"), []byte("# Test"), 0644))
	assert.True(t, importer.StagingExists())
}

func TestClassifier_AddFrontmatterToContent(t *testing.T) {
	classifier := NewClassifier()

	t.Run("add to plain content", func(t *testing.T) {
		class := &Classification{
			SuggestedType: "skill",
			SuggestedName: "test",
			Content:       "# Test\n\nContent here.",
		}

		result := classifier.AddFrontmatterToContent(class, "A test skill")

		assert.Contains(t, result, "---\nregis3:")
		assert.Contains(t, result, "# Test")
		assert.Contains(t, result, "Content here.")
	})

	t.Run("replace existing frontmatter", func(t *testing.T) {
		class := &Classification{
			SuggestedType: "skill",
			SuggestedName: "test",
			Content: `---
title: Old title
---
# Test

Content here.`,
		}

		result := classifier.AddFrontmatterToContent(class, "A test skill")

		assert.Contains(t, result, "regis3:")
		assert.NotContains(t, result, "title: Old title")
		assert.Contains(t, result, "# Test")
	})
}
