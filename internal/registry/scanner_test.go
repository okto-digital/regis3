package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner_Scan(t *testing.T) {
	// Use the actual registry directory in the project
	scanner := NewScanner("../../registry")

	result, err := scanner.Scan()
	require.NoError(t, err)

	// Should find our sample files
	assert.NotEmpty(t, result.Items, "should find items in registry")

	// Check that we found expected items
	foundTypes := make(map[string]bool)
	for _, item := range result.Items {
		foundTypes[item.Type] = true
	}

	assert.True(t, foundTypes["skill"], "should find skill type")
	assert.True(t, foundTypes["philosophy"], "should find philosophy type")
	assert.True(t, foundTypes["subagent"], "should find subagent type")
	assert.True(t, foundTypes["stack"], "should find stack type")
}

func TestScanner_ScanFile(t *testing.T) {
	scanner := NewScanner("../../registry")

	tests := []struct {
		name      string
		path      string
		wantType  string
		wantName  string
		wantErr   bool
		wantNoReg bool
	}{
		{
			name:     "valid skill file",
			path:     "../../registry/skills/git-conventions.md",
			wantType: "skill",
			wantName: "git-conventions",
		},
		{
			name:     "valid philosophy file",
			path:     "../../registry/philosophies/clean-code.md",
			wantType: "philosophy",
			wantName: "clean-code",
		},
		{
			name:     "valid subagent file",
			path:     "../../registry/agents/architect.md",
			wantType: "subagent",
			wantName: "architect",
		},
		{
			name:     "skill with dependencies",
			path:     "../../registry/skills/testing.md",
			wantType: "skill",
			wantName: "testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := scanner.ScanFile(tt.path)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			if tt.wantNoReg {
				require.ErrorIs(t, err, ErrNoRegis3Block)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, item.Type)
			assert.Equal(t, tt.wantName, item.Name)
			assert.NotEmpty(t, item.Content)
		})
	}
}

func TestScanner_ParsesAllFields(t *testing.T) {
	scanner := NewScanner("../../registry")

	// Test the testing.md file which has dependencies
	item, err := scanner.ScanFile("../../registry/skills/testing.md")
	require.NoError(t, err)

	assert.Equal(t, "skill", item.Type)
	assert.Equal(t, "testing", item.Name)
	assert.Equal(t, "Testing best practices and patterns", item.Desc)
	assert.Contains(t, item.Deps, "skill:git-conventions")
	assert.Contains(t, item.Tags, "testing")
	assert.Equal(t, "stable", item.Status)
}

func TestScanner_StackWithDeps(t *testing.T) {
	scanner := NewScanner("../../registry")

	item, err := scanner.ScanFile("../../registry/stacks/base.md")
	require.NoError(t, err)

	assert.Equal(t, "stack", item.Type)
	assert.Equal(t, "base", item.Name)

	// Stack should have multiple dependencies
	assert.GreaterOrEqual(t, len(item.Deps), 3)
	assert.Contains(t, item.Deps, "philosophy:clean-code")
	assert.Contains(t, item.Deps, "skill:git-conventions")
	assert.Contains(t, item.Deps, "skill:testing")
}

func TestScanner_NonExistentDirectory(t *testing.T) {
	scanner := NewScanner("/nonexistent/path")

	_, err := scanner.Scan()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestScanner_SkipsBuildDirectory(t *testing.T) {
	// Create a temp directory with .build folder
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create .build directory with a file
	buildDir := filepath.Join(tmpDir, ".build")
	require.NoError(t, os.MkdirAll(buildDir, 0755))

	buildFile := filepath.Join(buildDir, "manifest.json")
	require.NoError(t, os.WriteFile(buildFile, []byte("{}"), 0644))

	// Create a valid registry file
	skillsDir := filepath.Join(tmpDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	validFile := filepath.Join(skillsDir, "test.md")
	validContent := `---
regis3:
  type: skill
  name: test
  desc: Test skill
---
Content here.
`
	require.NoError(t, os.WriteFile(validFile, []byte(validContent), 0644))

	// Scan
	scanner := NewScanner(tmpDir)
	result, err := scanner.Scan()
	require.NoError(t, err)

	// Should find 1 item, not the .build/manifest.json
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "test", result.Items[0].Name)
}

func TestScanner_SkipsFilesWithoutRegis3Block(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "regis3-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a file without regis3 block
	noRegisFile := filepath.Join(tmpDir, "readme.md")
	noRegisContent := `---
title: Just a readme
---
# Hello

This is not a regis3 file.
`
	require.NoError(t, os.WriteFile(noRegisFile, []byte(noRegisContent), 0644))

	// Create a file with regis3 block
	regisFile := filepath.Join(tmpDir, "skill.md")
	regisContent := `---
regis3:
  type: skill
  name: test
  desc: Test
---
Content.
`
	require.NoError(t, os.WriteFile(regisFile, []byte(regisContent), 0644))

	scanner := NewScanner(tmpDir)
	result, err := scanner.Scan()
	require.NoError(t, err)

	assert.Len(t, result.Items, 1)
	assert.Len(t, result.Skipped, 1)
	assert.Contains(t, result.Skipped[0], "readme.md")
}

func TestHasRegis3Frontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name: "valid regis3 block",
			content: `---
regis3:
  type: skill
  name: test
  desc: Test
---
Content.
`,
			want: true,
		},
		{
			name: "no regis3 block",
			content: `---
title: Hello
---
Content.
`,
			want: false,
		},
		{
			name: "empty regis3 block",
			content: `---
regis3:
---
Content.
`,
			want: false,
		},
		{
			name:    "no frontmatter at all",
			content: "Just plain text.",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile, err := os.CreateTemp("", "regis3-test-*.md")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			tmpFile.Close()

			got, err := HasRegis3Frontmatter(tmpFile.Name())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
