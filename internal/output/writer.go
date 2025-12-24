package output

import (
	"io"
	"os"
)

// Format represents the output format type.
type Format string

const (
	FormatPretty Format = "pretty"
	FormatJSON   Format = "json"
	FormatQuiet  Format = "quiet"
)

// ParseFormat parses a format string into a Format type.
func ParseFormat(s string) Format {
	switch s {
	case "json":
		return FormatJSON
	case "quiet":
		return FormatQuiet
	default:
		return FormatPretty
	}
}

// Writer is the interface for output writers.
type Writer interface {
	// Write writes a response to the output.
	Write(resp *Response) error

	// WriteError writes an error response.
	WriteError(err error) error

	// Success writes a simple success message.
	Success(message string) error

	// Info writes an informational message.
	Info(message string) error

	// Warning writes a warning message.
	Warning(message string) error

	// Error writes an error message (without exiting).
	Error(message string) error

	// Table writes tabular data.
	Table(headers []string, rows [][]string) error

	// List writes a simple list of items.
	List(items []string) error

	// Progress writes a progress update (for long operations).
	Progress(current, total int, message string) error
}

// Config holds configuration for output writers.
type Config struct {
	// Output is the writer to write to (default: os.Stdout).
	Output io.Writer

	// ErrOutput is the writer for errors (default: os.Stderr).
	ErrOutput io.Writer

	// NoColor disables colored output.
	NoColor bool

	// Verbose enables verbose output.
	Verbose bool
}

// DefaultConfig returns the default output configuration.
func DefaultConfig() *Config {
	return &Config{
		Output:    os.Stdout,
		ErrOutput: os.Stderr,
		NoColor:   false,
		Verbose:   false,
	}
}

// New creates a new Writer based on the format.
func New(format Format, cfg *Config) Writer {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	switch format {
	case FormatJSON:
		return NewJSONWriter(cfg)
	case FormatQuiet:
		return NewQuietWriter(cfg)
	default:
		return NewPrettyWriter(cfg)
	}
}

// NewWithFormat creates a writer from a format string.
func NewWithFormat(format string, cfg *Config) Writer {
	return New(ParseFormat(format), cfg)
}
