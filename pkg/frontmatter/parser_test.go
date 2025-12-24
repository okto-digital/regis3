package frontmatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantFM      string
		wantBody    string
		wantErr     error
		errContains string
	}{
		{
			name: "valid frontmatter with body",
			input: `---
title: Hello
author: John
---
This is the body content.
`,
			wantFM:   "title: Hello\nauthor: John\n",
			wantBody: "This is the body content.\n",
			wantErr:  nil,
		},
		{
			name: "frontmatter only",
			input: `---
title: Hello
---
`,
			wantFM:   "title: Hello\n",
			wantBody: "",
			wantErr:  nil,
		},
		{
			name: "multiline body",
			input: `---
type: skill
---
# Heading

Paragraph one.

Paragraph two.
`,
			wantFM:   "type: skill\n",
			wantBody: "# Heading\n\nParagraph one.\n\nParagraph two.\n",
			wantErr:  nil,
		},
		{
			name:    "no frontmatter",
			input:   "Just regular content.",
			wantErr: ErrNoFrontmatter,
		},
		{
			name:    "empty file",
			input:   "",
			wantErr: ErrNoFrontmatter,
		},
		{
			name: "unclosed frontmatter",
			input: `---
title: Hello
never closed
`,
			wantErr: ErrUnclosedFrontmatter,
		},
		{
			name: "frontmatter with nested yaml",
			input: `---
regis3:
  type: skill
  name: test
  deps:
    - skill:one
    - skill:two
---
Body here.
`,
			wantFM:   "regis3:\n  type: skill\n  name: test\n  deps:\n    - skill:one\n    - skill:two\n",
			wantBody: "Body here.\n",
			wantErr:  nil,
		},
		{
			name: "empty frontmatter",
			input: `---
---
Body content.
`,
			wantFM:   "",
			wantBody: "Body content.\n",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantFM, doc.Frontmatter)
			assert.Equal(t, tt.wantBody, doc.Body)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type Regis3Meta struct {
		Type string   `yaml:"type"`
		Name string   `yaml:"name"`
		Desc string   `yaml:"desc"`
		Deps []string `yaml:"deps"`
		Tags []string `yaml:"tags"`
	}

	type FrontMatter struct {
		Regis3 Regis3Meta `yaml:"regis3"`
	}

	tests := []struct {
		name     string
		input    string
		wantMeta Regis3Meta
		wantBody string
		wantErr  bool
	}{
		{
			name: "full regis3 frontmatter",
			input: `---
regis3:
  type: skill
  name: test-skill
  desc: A test skill for testing
  deps:
    - skill:base
    - skill:common
  tags:
    - testing
    - example
---
# Test Skill

This is the skill content.
`,
			wantMeta: Regis3Meta{
				Type: "skill",
				Name: "test-skill",
				Desc: "A test skill for testing",
				Deps: []string{"skill:base", "skill:common"},
				Tags: []string{"testing", "example"},
			},
			wantBody: "# Test Skill\n\nThis is the skill content.\n",
		},
		{
			name: "minimal frontmatter",
			input: `---
regis3:
  type: subagent
  name: helper
  desc: Helper agent
---
Content.
`,
			wantMeta: Regis3Meta{
				Type: "subagent",
				Name: "helper",
				Desc: "Helper agent",
			},
			wantBody: "Content.\n",
		},
		{
			name: "inline yaml format",
			input: `---
regis3: {type: mcp, name: test-mcp, desc: Test MCP config}
---
{}
`,
			wantMeta: Regis3Meta{
				Type: "mcp",
				Name: "test-mcp",
				Desc: "Test MCP config",
			},
			wantBody: "{}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fm FrontMatter
			doc, err := UnmarshalString(tt.input, &fm)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMeta.Type, fm.Regis3.Type)
			assert.Equal(t, tt.wantMeta.Name, fm.Regis3.Name)
			assert.Equal(t, tt.wantMeta.Desc, fm.Regis3.Desc)
			assert.Equal(t, tt.wantMeta.Deps, fm.Regis3.Deps)
			assert.Equal(t, tt.wantMeta.Tags, fm.Regis3.Tags)
			assert.Equal(t, tt.wantBody, doc.Body)
		})
	}
}

func TestParseBytes(t *testing.T) {
	input := []byte(`---
key: value
---
Body.
`)
	doc, err := ParseBytes(input)
	require.NoError(t, err)
	assert.Equal(t, "key: value\n", doc.Frontmatter)
	assert.Equal(t, "Body.\n", doc.Body)
}
