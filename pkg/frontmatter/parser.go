// Package frontmatter provides utilities for parsing YAML frontmatter from markdown files.
package frontmatter

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// ErrNoFrontmatter indicates the file has no YAML frontmatter block.
	ErrNoFrontmatter = errors.New("no frontmatter found")
	// ErrUnclosedFrontmatter indicates the frontmatter block was not closed.
	ErrUnclosedFrontmatter = errors.New("unclosed frontmatter block")
)

const delimiter = "---"

// Document represents a parsed markdown file with frontmatter.
type Document struct {
	// Frontmatter is the raw YAML frontmatter string.
	Frontmatter string
	// Body is the content after the frontmatter.
	Body string
}

// Parse extracts frontmatter and body from a markdown document.
// The frontmatter must be at the start of the file, enclosed by --- delimiters.
func Parse(r io.Reader) (*Document, error) {
	scanner := bufio.NewScanner(r)

	// First line must be the opening delimiter
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, ErrNoFrontmatter
	}

	if strings.TrimSpace(scanner.Text()) != delimiter {
		return nil, ErrNoFrontmatter
	}

	// Read frontmatter until closing delimiter
	var frontmatter strings.Builder
	found := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == delimiter {
			found = true
			break
		}
		frontmatter.WriteString(line)
		frontmatter.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrUnclosedFrontmatter
	}

	// Read the rest as body
	var body strings.Builder
	for scanner.Scan() {
		body.WriteString(scanner.Text())
		body.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Document{
		Frontmatter: frontmatter.String(),
		Body:        body.String(),
	}, nil
}

// ParseBytes parses frontmatter from a byte slice.
func ParseBytes(data []byte) (*Document, error) {
	return Parse(bytes.NewReader(data))
}

// ParseString parses frontmatter from a string.
func ParseString(s string) (*Document, error) {
	return Parse(strings.NewReader(s))
}

// Unmarshal parses frontmatter and unmarshals it into the provided struct.
// The struct should have yaml tags for proper field mapping.
func Unmarshal(r io.Reader, v interface{}) (*Document, error) {
	doc, err := Parse(r)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal([]byte(doc.Frontmatter), v); err != nil {
		return doc, err
	}

	return doc, nil
}

// UnmarshalBytes parses frontmatter from bytes and unmarshals it.
func UnmarshalBytes(data []byte, v interface{}) (*Document, error) {
	return Unmarshal(bytes.NewReader(data), v)
}

// UnmarshalString parses frontmatter from a string and unmarshals it.
func UnmarshalString(s string, v interface{}) (*Document, error) {
	return Unmarshal(strings.NewReader(s), v)
}
