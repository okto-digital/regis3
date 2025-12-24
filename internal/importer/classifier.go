package importer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/okto-digital/regis3/internal/registry"
	"github.com/okto-digital/regis3/pkg/frontmatter"
)

// Classifier classifies markdown files and suggests regis3 types.
type Classifier struct{}

// NewClassifier creates a new classifier.
func NewClassifier() *Classifier {
	return &Classifier{}
}

// Classification contains the classification result for a file.
type Classification struct {
	// Path is the file path.
	Path string

	// SuggestedType is the suggested regis3 type.
	SuggestedType string

	// SuggestedName is the suggested name (kebab-case).
	SuggestedName string

	// Confidence is the classification confidence (0-100).
	Confidence int

	// Reason explains why this type was suggested.
	Reason string

	// ExistingMeta contains parsed regis3 metadata if present.
	ExistingMeta *registry.Regis3Meta

	// HasValidRegis3 indicates if file has valid regis3 metadata.
	HasValidRegis3 bool

	// Content is the file content.
	Content string
}

// Classify classifies a single file.
func (c *Classifier) Classify(path string) (*Classification, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result := &Classification{
		Path:    path,
		Content: string(content),
	}

	// Try to parse existing regis3 metadata
	result.ExistingMeta, result.HasValidRegis3 = c.parseExistingMeta(string(content))

	if result.HasValidRegis3 {
		// Use existing metadata
		result.SuggestedType = result.ExistingMeta.Type
		result.SuggestedName = result.ExistingMeta.Name
		result.Confidence = 100
		result.Reason = "has valid regis3 metadata"
		return result, nil
	}

	// Suggest based on content analysis
	result.SuggestedType, result.Confidence, result.Reason = c.suggestType(path, string(content))
	result.SuggestedName = c.suggestName(path)

	return result, nil
}

// parseExistingMeta attempts to parse regis3 metadata from content.
func (c *Classifier) parseExistingMeta(content string) (*registry.Regis3Meta, bool) {
	// Try to parse frontmatter
	doc, err := frontmatter.ParseBytes([]byte(content))
	if err != nil || doc.Frontmatter == "" {
		return nil, false
	}

	// Check for regis3 block
	var wrapper struct {
		Regis3 registry.Regis3Meta `yaml:"regis3"`
	}

	_, err = frontmatter.UnmarshalBytes([]byte(content), &wrapper)
	if err != nil {
		return nil, false
	}

	// Validate required fields
	if wrapper.Regis3.Type == "" || wrapper.Regis3.Name == "" {
		return nil, false
	}

	return &wrapper.Regis3, true
}

// suggestType suggests a type based on content analysis.
func (c *Classifier) suggestType(path, content string) (string, int, string) {
	lowerContent := strings.ToLower(content)
	filename := strings.ToLower(filepath.Base(path))
	dir := strings.ToLower(filepath.Dir(path))

	// Check directory hints
	if strings.Contains(dir, "skill") {
		return "skill", 80, "directory contains 'skill'"
	}
	if strings.Contains(dir, "agent") {
		return "subagent", 80, "directory contains 'agent'"
	}
	if strings.Contains(dir, "philosoph") {
		return "philosophy", 80, "directory contains 'philosophy'"
	}
	if strings.Contains(dir, "command") {
		return "command", 80, "directory contains 'command'"
	}
	if strings.Contains(dir, "prompt") {
		return "prompt", 80, "directory contains 'prompt'"
	}
	if strings.Contains(dir, "doc") {
		return "doc", 70, "directory contains 'doc'"
	}

	// Check filename patterns
	if strings.Contains(filename, "agent") {
		return "subagent", 70, "filename contains 'agent'"
	}
	if strings.Contains(filename, "skill") {
		return "skill", 70, "filename contains 'skill'"
	}
	if strings.Contains(filename, "prompt") {
		return "prompt", 70, "filename contains 'prompt'"
	}

	// Content-based classification
	// Check for agent-like patterns
	agentPatterns := []string{
		"you are",
		"your role",
		"as an agent",
		"your task is",
		"you will",
		"act as",
	}
	for _, pattern := range agentPatterns {
		if strings.Contains(lowerContent, pattern) {
			return "subagent", 60, "content suggests agent instructions"
		}
	}

	// Check for skill-like patterns
	skillPatterns := []string{
		"## usage",
		"## example",
		"how to",
		"best practice",
		"convention",
		"guidelines",
	}
	for _, pattern := range skillPatterns {
		if strings.Contains(lowerContent, pattern) {
			return "skill", 50, "content suggests skill/guidelines"
		}
	}

	// Check for philosophy-like patterns
	philosophyPatterns := []string{
		"principle",
		"philosophy",
		"approach",
		"methodology",
		"mindset",
	}
	for _, pattern := range philosophyPatterns {
		if strings.Contains(lowerContent, pattern) {
			return "philosophy", 50, "content suggests philosophy/principles"
		}
	}

	// Check for command-like patterns
	commandPatterns := []string{
		"```bash",
		"```sh",
		"run the following",
		"execute",
	}
	for _, pattern := range commandPatterns {
		if strings.Contains(lowerContent, pattern) {
			return "command", 40, "content contains command/script patterns"
		}
	}

	// Default to doc if has documentation-like structure
	if strings.Contains(lowerContent, "## ") || strings.Contains(lowerContent, "# ") {
		return "doc", 30, "appears to be documentation"
	}

	// Fallback
	return "skill", 20, "default classification"
}

// suggestName suggests a kebab-case name from the file path.
func (c *Classifier) suggestName(path string) string {
	// Get filename without extension
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)]

	return toKebabCase(name)
}

// toKebabCase converts a string to kebab-case.
func toKebabCase(s string) string {
	// Replace common separators with hyphens
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")

	// Insert hyphens before uppercase letters (camelCase)
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, "${1}-${2}")

	// Remove non-alphanumeric characters except hyphens
	re = regexp.MustCompile(`[^a-zA-Z0-9-]`)
	s = re.ReplaceAllString(s, "")

	// Lowercase
	s = strings.ToLower(s)

	// Remove consecutive hyphens
	re = regexp.MustCompile(`-+`)
	s = re.ReplaceAllString(s, "-")

	// Trim hyphens from ends
	s = strings.Trim(s, "-")

	return s
}

// ClassifyMany classifies multiple files.
func (c *Classifier) ClassifyMany(paths []string) ([]*Classification, error) {
	var results []*Classification
	for _, path := range paths {
		result, err := c.Classify(path)
		if err != nil {
			// Add error but continue
			results = append(results, &Classification{
				Path:   path,
				Reason: "error: " + err.Error(),
			})
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

// GenerateFrontmatter generates regis3 frontmatter for a classification.
func (c *Classifier) GenerateFrontmatter(class *Classification, desc string) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString("regis3:\n")
	sb.WriteString("  type: " + class.SuggestedType + "\n")
	sb.WriteString("  name: " + class.SuggestedName + "\n")
	if desc != "" {
		sb.WriteString("  desc: " + desc + "\n")
	} else {
		sb.WriteString("  desc: \"TODO: Add description\"\n")
	}
	sb.WriteString("  tags:\n")
	sb.WriteString("    - imported\n")
	sb.WriteString("---\n")

	return sb.String()
}

// AddFrontmatterToContent adds frontmatter to content that doesn't have it.
func (c *Classifier) AddFrontmatterToContent(class *Classification, desc string) string {
	fm := c.GenerateFrontmatter(class, desc)

	// Check if content already has frontmatter
	if strings.HasPrefix(class.Content, "---") {
		// Find end of existing frontmatter
		lines := strings.Split(class.Content, "\n")
		endIdx := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				endIdx = i
				break
			}
		}
		if endIdx > 0 {
			// Replace existing frontmatter
			body := strings.Join(lines[endIdx+1:], "\n")
			return fm + strings.TrimSpace(body)
		}
	}

	// Add frontmatter to content
	return fm + class.Content
}
