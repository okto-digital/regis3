package installer

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/okto-digital/regis3/internal/registry"
)

// Transformer handles content transformations for installation.
type Transformer struct {
	target *Target
}

// NewTransformer creates a new transformer for a target.
func NewTransformer(target *Target) *Transformer {
	return &Transformer{target: target}
}

// Transform applies transformations to item content.
func (t *Transformer) Transform(item *registry.Item) (string, error) {
	content := item.Content
	cfg := t.target.GetTransform(item.Type)

	// Strip frontmatter if configured
	if cfg.StripFrontmatter {
		content = stripFrontmatter(content)
	}

	// Add header if configured
	if cfg.AddHeader != "" {
		header := t.expandTemplate(cfg.AddHeader, item)
		content = header + "\n\n" + content
	}

	// Wrap content if configured
	if cfg.WrapWith != "" {
		content = t.expandTemplate(cfg.WrapWith, item)
	}

	return strings.TrimSpace(content), nil
}

// expandTemplate expands placeholders in a template string.
func (t *Transformer) expandTemplate(template string, item *registry.Item) string {
	result := template

	// Replace common placeholders
	result = strings.ReplaceAll(result, "{name}", item.Name)
	result = strings.ReplaceAll(result, "{type}", item.Type)
	result = strings.ReplaceAll(result, "{desc}", item.Desc)
	result = strings.ReplaceAll(result, "{content}", item.Content)

	return result
}

// stripFrontmatter removes YAML frontmatter from content.
func stripFrontmatter(content string) string {
	// Check if content starts with frontmatter delimiter
	if !strings.HasPrefix(content, "---") {
		return content
	}

	// Find the closing delimiter
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return content
	}

	// Find end of frontmatter
	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		return content
	}

	// Return content after frontmatter
	body := strings.Join(lines[endIndex+1:], "\n")
	return strings.TrimSpace(body)
}

// MergeContent handles merging multiple items into CLAUDE.md.
type MergeContent struct {
	sections map[string][]MergeSection
}

// MergeSection represents a section to be merged.
type MergeSection struct {
	Item    *registry.Item
	Content string
	Order   int
}

// NewMergeContent creates a new merge content handler.
func NewMergeContent() *MergeContent {
	return &MergeContent{
		sections: make(map[string][]MergeSection),
	}
}

// Add adds an item to be merged.
func (m *MergeContent) Add(item *registry.Item, content string) {
	section := MergeSection{
		Item:    item,
		Content: content,
		Order:   item.Order,
	}
	m.sections[item.Type] = append(m.sections[item.Type], section)
}

// Generate generates the merged content.
func (m *MergeContent) Generate() string {
	var result strings.Builder

	// Define section order
	typeOrder := []string{"project", "philosophy", "ruleset"}

	for _, itemType := range typeOrder {
		sections, ok := m.sections[itemType]
		if !ok || len(sections) == 0 {
			continue
		}

		// Sort by order
		sort.Slice(sections, func(i, j int) bool {
			return sections[i].Order < sections[j].Order
		})

		// Write section header
		result.WriteString(fmt.Sprintf("## %s\n\n", capitalizeFirst(itemType)))

		// Write each item
		for _, section := range sections {
			result.WriteString(section.Content)
			result.WriteString("\n\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// HasContent returns true if there's content to merge.
func (m *MergeContent) HasContent() bool {
	for _, sections := range m.sections {
		if len(sections) > 0 {
			return true
		}
	}
	return false
}

// UpdateExistingFile updates an existing CLAUDE.md with new merged content.
// It preserves user content outside of managed sections.
func UpdateExistingFile(existing, newContent string) string {
	// Markers for managed content
	startMarker := "<!-- regis3:start -->"
	endMarker := "<!-- regis3:end -->"

	// Check if file has managed section
	startIdx := strings.Index(existing, startMarker)
	endIdx := strings.Index(existing, endMarker)

	if startIdx == -1 || endIdx == -1 {
		// No managed section - append at the end
		if strings.TrimSpace(existing) == "" {
			return wrapManagedContent(newContent)
		}
		return existing + "\n\n" + wrapManagedContent(newContent)
	}

	// Replace managed section
	before := existing[:startIdx]
	after := existing[endIdx+len(endMarker):]

	return before + wrapManagedContent(newContent) + after
}

// wrapManagedContent wraps content with regis3 markers.
func wrapManagedContent(content string) string {
	return fmt.Sprintf("<!-- regis3:start -->\n%s\n<!-- regis3:end -->", content)
}

// ExtractManagedContent extracts content between regis3 markers.
func ExtractManagedContent(content string) string {
	startMarker := "<!-- regis3:start -->"
	endMarker := "<!-- regis3:end -->"

	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	if startIdx == -1 || endIdx == -1 {
		return ""
	}

	managed := content[startIdx+len(startMarker) : endIdx]
	return strings.TrimSpace(managed)
}

// ValidateContent checks if content is valid for installation.
func ValidateContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is empty")
	}
	return nil
}

// SanitizeName ensures a name is safe for use in filenames.
func SanitizeName(name string) string {
	// Replace unsafe characters (keeping only alphanumeric and hyphen)
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	sanitized := re.ReplaceAllString(name, "-")

	// Remove consecutive dashes
	re = regexp.MustCompile(`-+`)
	sanitized = re.ReplaceAllString(sanitized, "-")

	// Trim dashes from ends
	sanitized = strings.Trim(sanitized, "-")

	return strings.ToLower(sanitized)
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
