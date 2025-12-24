package importer

import (
	"os"
	"path/filepath"
	"strings"
)

// ExternalScanner scans external directories for markdown files.
type ExternalScanner struct {
	// SkipDirs are directory names to skip.
	SkipDirs []string

	// Extensions are file extensions to include.
	Extensions []string
}

// NewExternalScanner creates a new external scanner with defaults.
func NewExternalScanner() *ExternalScanner {
	return &ExternalScanner{
		SkipDirs: []string{
			".git",
			".svn",
			".hg",
			"node_modules",
			"vendor",
			".build",
			"__pycache__",
			".cache",
			".vscode",
			".idea",
		},
		Extensions: []string{".md", ".markdown"},
	}
}

// ScanResult contains the results of scanning an external path.
type ScanResult struct {
	// Path is the scanned path.
	Path string

	// Files are all found markdown files.
	Files []ScannedFile

	// Errors are any errors encountered during scanning.
	Errors []ScanError
}

// ScannedFile represents a scanned markdown file.
type ScannedFile struct {
	// Path is the absolute path to the file.
	Path string

	// RelPath is the path relative to the scan root.
	RelPath string

	// Name is the filename without extension.
	Name string

	// Size is the file size in bytes.
	Size int64

	// HasFrontmatter indicates if the file has YAML frontmatter.
	HasFrontmatter bool

	// HasRegis3 indicates if the file has a regis3 block.
	HasRegis3 bool
}

// ScanError represents an error during scanning.
type ScanError struct {
	Path    string
	Message string
	Err     error
}

func (e ScanError) Error() string {
	return e.Path + ": " + e.Message
}

// Scan scans a directory for markdown files.
func (s *ExternalScanner) Scan(rootPath string) (*ScanResult, error) {
	// Resolve to absolute path
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	// Check if path exists
	info, err := os.Stat(absRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ScanError{Path: rootPath, Message: "path does not exist", Err: err}
		}
		return nil, err
	}

	result := &ScanResult{
		Path: absRoot,
	}

	// If it's a file, just process that file
	if !info.IsDir() {
		if s.isMarkdown(absRoot) {
			file, err := s.scanFile(absRoot, absRoot)
			if err != nil {
				result.Errors = append(result.Errors, ScanError{
					Path:    absRoot,
					Message: err.Error(),
					Err:     err,
				})
			} else {
				result.Files = append(result.Files, *file)
			}
		}
		return result, nil
	}

	// Walk the directory
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, ScanError{
				Path:    path,
				Message: err.Error(),
				Err:     err,
			})
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			if s.shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if markdown file
		if !s.isMarkdown(path) {
			return nil
		}

		// Scan the file
		file, err := s.scanFile(path, absRoot)
		if err != nil {
			result.Errors = append(result.Errors, ScanError{
				Path:    path,
				Message: err.Error(),
				Err:     err,
			})
			return nil
		}

		result.Files = append(result.Files, *file)
		return nil
	})

	if err != nil {
		return result, err
	}

	return result, nil
}

// scanFile scans a single file and returns its info.
func (s *ExternalScanner) scanFile(path, rootPath string) (*ScannedFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(rootPath, path)
	if err != nil {
		relPath = path
	}

	// Get filename without extension
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)]

	// Check for frontmatter and regis3 block
	hasFrontmatter, hasRegis3, err := s.checkFrontmatter(path)
	if err != nil {
		return nil, err
	}

	return &ScannedFile{
		Path:           path,
		RelPath:        relPath,
		Name:           name,
		Size:           info.Size(),
		HasFrontmatter: hasFrontmatter,
		HasRegis3:      hasRegis3,
	}, nil
}

// checkFrontmatter checks if a file has frontmatter and regis3 block.
func (s *ExternalScanner) checkFrontmatter(path string) (hasFrontmatter, hasRegis3 bool, err error) {
	// Read first 4KB to check for frontmatter
	f, err := os.Open(path)
	if err != nil {
		return false, false, err
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false, false, err
	}

	content := string(buf[:n])

	// Check for frontmatter delimiter
	if !strings.HasPrefix(content, "---") {
		return false, false, nil
	}

	// Find closing delimiter
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return false, false, nil
	}

	inFrontmatter := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 && trimmed == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			if trimmed == "---" {
				hasFrontmatter = true
				break
			}
			// Check for regis3 block
			if strings.HasPrefix(trimmed, "regis3:") {
				hasRegis3 = true
			}
		}
	}

	return hasFrontmatter, hasRegis3, nil
}

// isMarkdown checks if a file is a markdown file.
func (s *ExternalScanner) isMarkdown(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, mdExt := range s.Extensions {
		if ext == mdExt {
			return true
		}
	}
	return false
}

// shouldSkipDir checks if a directory should be skipped.
func (s *ExternalScanner) shouldSkipDir(name string) bool {
	// Skip hidden directories
	if strings.HasPrefix(name, ".") {
		return true
	}
	for _, skip := range s.SkipDirs {
		if name == skip {
			return true
		}
	}
	return false
}

// Stats returns statistics about the scan result.
func (r *ScanResult) Stats() ScanStats {
	stats := ScanStats{
		TotalFiles: len(r.Files),
		Errors:     len(r.Errors),
	}

	for _, f := range r.Files {
		if f.HasRegis3 {
			stats.WithRegis3++
		} else if f.HasFrontmatter {
			stats.WithFrontmatter++
		} else {
			stats.PlainMarkdown++
		}
	}

	return stats
}

// ScanStats contains statistics about scanned files.
type ScanStats struct {
	TotalFiles      int
	WithRegis3      int
	WithFrontmatter int
	PlainMarkdown   int
	Errors          int
}

// FilterByRegis3 returns only files with regis3 blocks.
func (r *ScanResult) FilterByRegis3() []ScannedFile {
	var files []ScannedFile
	for _, f := range r.Files {
		if f.HasRegis3 {
			files = append(files, f)
		}
	}
	return files
}

// FilterWithoutRegis3 returns files without regis3 blocks.
func (r *ScanResult) FilterWithoutRegis3() []ScannedFile {
	var files []ScannedFile
	for _, f := range r.Files {
		if !f.HasRegis3 {
			files = append(files, f)
		}
	}
	return files
}
