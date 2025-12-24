package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/okto-digital/regis3/pkg/frontmatter"
)

// Scanner finds and parses registry items from markdown files.
type Scanner struct {
	// RootDir is the registry root directory.
	RootDir string
}

// NewScanner creates a new scanner for the given registry directory.
func NewScanner(rootDir string) *Scanner {
	return &Scanner{RootDir: rootDir}
}

// ScanResult contains the results of scanning the registry.
type ScanResult struct {
	// Items are successfully parsed items.
	Items []*Item
	// Errors are files that failed to parse.
	Errors []ScanError
	// Skipped are files without regis3 frontmatter.
	Skipped []string
}

// ScanError represents an error encountered while scanning a file.
type ScanError struct {
	Path    string
	Message string
	Err     error
}

func (e ScanError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Path, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// Scan walks the registry directory and parses all markdown files.
func (s *Scanner) Scan() (*ScanResult, error) {
	result := &ScanResult{
		Items:   make([]*Item, 0),
		Errors:  make([]ScanError, 0),
		Skipped: make([]string, 0),
	}

	// Check if root directory exists
	if _, err := os.Stat(s.RootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("registry directory does not exist: %s", s.RootDir)
	}

	err := filepath.Walk(s.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, ScanError{
				Path:    path,
				Message: "failed to access",
				Err:     err,
			})
			return nil // continue walking
		}

		// Skip directories
		if info.IsDir() {
			// Skip .build directory
			if info.Name() == ".build" {
				return filepath.SkipDir
			}
			// Skip .git directory
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}

		// Parse the file
		item, err := s.parseFile(path)
		if err != nil {
			if err == ErrNoRegis3Block {
				result.Skipped = append(result.Skipped, path)
			} else {
				result.Errors = append(result.Errors, ScanError{
					Path:    path,
					Message: "failed to parse",
					Err:     err,
				})
			}
			return nil
		}

		result.Items = append(result.Items, item)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk registry: %w", err)
	}

	return result, nil
}

// ErrNoRegis3Block indicates the file has no regis3 frontmatter block.
var ErrNoRegis3Block = fmt.Errorf("no regis3 frontmatter block")

// parseFile reads and parses a single markdown file.
func (s *Scanner) parseFile(path string) (*Item, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse frontmatter
	var fm FrontMatter
	doc, err := frontmatter.UnmarshalBytes(content, &fm)
	if err != nil {
		// Check if it's a "no frontmatter" error
		if err == frontmatter.ErrNoFrontmatter {
			return nil, ErrNoRegis3Block
		}
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Check if regis3 block exists
	if fm.Regis3.Type == "" && fm.Regis3.Name == "" {
		return nil, ErrNoRegis3Block
	}

	// Calculate relative path from registry root
	relPath, err := filepath.Rel(s.RootDir, path)
	if err != nil {
		relPath = path
	}

	// Create item
	item := &Item{
		Regis3Meta: fm.Regis3,
		Source:     relPath,
		Content:    doc.Body,
		SourceDir:  filepath.Dir(relPath),
	}

	return item, nil
}

// ScanFile parses a single file and returns the item.
func (s *Scanner) ScanFile(path string) (*Item, error) {
	return s.parseFile(path)
}

// HasRegis3Frontmatter checks if a file has valid regis3 frontmatter.
func HasRegis3Frontmatter(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	var fm FrontMatter
	_, err = frontmatter.UnmarshalBytes(content, &fm)
	if err != nil {
		return false, nil // No frontmatter or invalid = false
	}

	// Check if regis3 block has required fields
	return fm.Regis3.Type != "" && fm.Regis3.Name != "", nil
}
