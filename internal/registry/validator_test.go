package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_RequiredFields(t *testing.T) {
	v := NewValidator(".")

	tests := []struct {
		name       string
		item       *Item
		wantErrors []string
	}{
		{
			name: "missing type",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Name: "test",
					Desc: "Test description",
				},
				Source: "test.md",
			},
			wantErrors: []string{"type"},
		},
		{
			name: "missing name",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type: "skill",
					Desc: "Test description",
				},
				Source: "test.md",
			},
			wantErrors: []string{"name"},
		},
		{
			name: "missing desc",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type: "skill",
					Name: "test",
				},
				Source: "test.md",
			},
			wantErrors: []string{"desc"},
		},
		{
			name: "all required fields present",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type: "skill",
					Name: "test",
					Desc: "A valid description here",
				},
				Source: "test.md",
			},
			wantErrors: nil,
		},
		{
			name: "all fields missing",
			item: &Item{
				Source: "test.md",
			},
			wantErrors: []string{"type", "name", "desc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateItem(tt.item)
			errors := result.Errors()

			if tt.wantErrors == nil {
				assert.Empty(t, errors)
				return
			}

			assert.Len(t, errors, len(tt.wantErrors))
			for _, field := range tt.wantErrors {
				found := false
				for _, err := range errors {
					if err.Field == field {
						found = true
						break
					}
				}
				assert.True(t, found, "expected error for field: %s", field)
			}
		})
	}
}

func TestValidator_InvalidType(t *testing.T) {
	v := NewValidator(".")

	item := &Item{
		Regis3Meta: Regis3Meta{
			Type: "invalid-type",
			Name: "test",
			Desc: "Test description",
		},
		Source: "test.md",
	}

	result := v.ValidateItem(item)
	errors := result.Errors()

	require.Len(t, errors, 1)
	assert.Equal(t, "type", errors[0].Field)
	assert.Contains(t, errors[0].Message, "invalid type")
}

func TestValidator_ValidTypes(t *testing.T) {
	v := NewValidator(".")

	for _, itemType := range ValidTypes {
		t.Run(string(itemType), func(t *testing.T) {
			item := &Item{
				Regis3Meta: Regis3Meta{
					Type: string(itemType),
					Name: "test",
					Desc: "A valid test description",
					Tags: []string{"test"},
				},
				Source: "test.md",
			}

			result := v.ValidateItem(item)
			errors := result.Errors()

			// Hook type requires trigger and run
			if itemType == TypeHook {
				// Expect errors for missing trigger and run
				assert.NotEmpty(t, errors)
			} else {
				assert.Empty(t, errors, "type %s should be valid", itemType)
			}
		})
	}
}

func TestValidator_DuplicateNames(t *testing.T) {
	v := NewValidator(".")

	items := []*Item{
		{
			Regis3Meta: Regis3Meta{
				Type: "skill",
				Name: "duplicate",
				Desc: "First item",
				Tags: []string{"test"},
			},
			Source: "first.md",
		},
		{
			Regis3Meta: Regis3Meta{
				Type: "skill",
				Name: "duplicate",
				Desc: "Second item with same name",
				Tags: []string{"test"},
			},
			Source: "second.md",
		},
	}

	result := v.ValidateItems(items)
	errors := result.Errors()

	require.Len(t, errors, 1)
	assert.Contains(t, errors[0].Message, "duplicate")
	assert.Equal(t, "second.md", errors[0].Path)
}

func TestValidator_DependencyNotFound(t *testing.T) {
	v := NewValidator(".")

	items := []*Item{
		{
			Regis3Meta: Regis3Meta{
				Type: "skill",
				Name: "test",
				Desc: "Test skill",
				Deps: []string{"skill:nonexistent"},
				Tags: []string{"test"},
			},
			Source: "test.md",
		},
	}

	result := v.ValidateItems(items)
	errors := result.Errors()

	require.Len(t, errors, 1)
	assert.Equal(t, "deps", errors[0].Field)
	assert.Contains(t, errors[0].Message, "skill:nonexistent")
}

func TestValidator_ValidDependencies(t *testing.T) {
	v := NewValidator(".")

	items := []*Item{
		{
			Regis3Meta: Regis3Meta{
				Type: "skill",
				Name: "base",
				Desc: "Base skill",
				Tags: []string{"test"},
			},
			Source: "base.md",
		},
		{
			Regis3Meta: Regis3Meta{
				Type: "skill",
				Name: "dependent",
				Desc: "Depends on base",
				Deps: []string{"skill:base"},
				Tags: []string{"test"},
			},
			Source: "dependent.md",
		},
	}

	result := v.ValidateItems(items)
	errors := result.Errors()

	assert.Empty(t, errors)
}

func TestValidator_Warnings(t *testing.T) {
	v := NewValidator(".")

	tests := []struct {
		name         string
		item         *Item
		wantWarnings []string
	}{
		{
			name: "short description",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type: "skill",
					Name: "test",
					Desc: "Short",
					Tags: []string{"test"},
				},
				Source: "test.md",
			},
			wantWarnings: []string{"desc"},
		},
		{
			name: "no tags",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type: "skill",
					Name: "test",
					Desc: "A proper description here",
				},
				Source: "test.md",
			},
			wantWarnings: []string{"tags"},
		},
		{
			name: "non-kebab-case name",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type: "skill",
					Name: "TestName",
					Desc: "A proper description here",
					Tags: []string{"test"},
				},
				Source: "test.md",
			},
			wantWarnings: []string{"name"},
		},
		{
			name: "stack without deps",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type: "stack",
					Name: "test-stack",
					Desc: "A stack without dependencies",
					Tags: []string{"test"},
				},
				Source: "test.md",
			},
			wantWarnings: []string{"deps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateItem(tt.item)
			warnings := result.Warnings()

			for _, field := range tt.wantWarnings {
				found := false
				for _, w := range warnings {
					if w.Field == field {
						found = true
						break
					}
				}
				assert.True(t, found, "expected warning for field: %s", field)
			}
		})
	}
}

func TestValidator_HookType(t *testing.T) {
	v := NewValidator(".")

	tests := []struct {
		name       string
		item       *Item
		wantErrors int
	}{
		{
			name: "hook without trigger and run",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type: "hook",
					Name: "test-hook",
					Desc: "A test hook",
					Tags: []string{"test"},
				},
				Source: "test.md",
			},
			wantErrors: 2, // missing trigger and run
		},
		{
			name: "hook with trigger but no run",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type:    "hook",
					Name:    "test-hook",
					Desc:    "A test hook",
					Tags:    []string{"test"},
					Trigger: "post-install",
				},
				Source: "test.md",
			},
			wantErrors: 1, // missing run
		},
		{
			name: "valid hook",
			item: &Item{
				Regis3Meta: Regis3Meta{
					Type:    "hook",
					Name:    "test-hook",
					Desc:    "A test hook",
					Tags:    []string{"test"},
					Trigger: "post-install",
					Run:     "scripts/setup.sh",
				},
				Source: "test.md",
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateItem(tt.item)
			errors := result.Errors()
			assert.Len(t, errors, tt.wantErrors)
		})
	}
}

func TestIsKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"git-conventions", true},
		{"simple", true},
		{"a-b-c", true},
		{"test123", true},
		{"test-123", true},
		{"GitConventions", false},
		{"git_conventions", false},
		{"-starts-with-hyphen", false},
		{"ends-with-hyphen-", false},
		{"double--hyphen", false},
		{"", false},
		{"UPPERCASE", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isKebabCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidationResult_Methods(t *testing.T) {
	result := &ValidationResult{}

	result.AddError("file1.md", "type", "missing type")
	result.AddError("file2.md", "name", "missing name")
	result.AddWarning("file1.md", "desc", "short description")
	result.AddInfo("file1.md", "", "just info")

	assert.True(t, result.HasErrors())
	assert.Len(t, result.Errors(), 2)
	assert.Len(t, result.Warnings(), 1)
	assert.Len(t, result.Issues, 4)
}

func TestValidator_WithSampleRegistry(t *testing.T) {
	// Scan and validate the actual sample registry
	scanner := NewScanner("../../registry")
	scanResult, err := scanner.Scan()
	require.NoError(t, err)

	validator := NewValidator("../../registry")
	valResult := validator.ValidateItems(scanResult.Items)

	// Our sample registry should be valid
	errors := valResult.Errors()
	assert.Empty(t, errors, "sample registry should have no validation errors: %v", errors)
}
