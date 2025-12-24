package output

import (
	"fmt"
	"io"
	"strings"
)

// QuietWriter outputs minimal text suitable for piping to other commands.
// Only outputs essential data, one item per line.
type QuietWriter struct {
	out     io.Writer
	errOut  io.Writer
	verbose bool
}

// NewQuietWriter creates a new quiet writer.
func NewQuietWriter(cfg *Config) *QuietWriter {
	return &QuietWriter{
		out:     cfg.Output,
		errOut:  cfg.ErrOutput,
		verbose: cfg.Verbose,
	}
}

// Write writes a response in minimal format.
func (w *QuietWriter) Write(resp *Response) error {
	// In quiet mode, only output data
	if resp.Data != nil {
		w.writeData(resp.Data)
	}

	// Write error to stderr if present
	if resp.Error != nil {
		fmt.Fprintln(w.errOut, resp.Error.Message)
	}

	return nil
}

// writeData writes response data in minimal format.
func (w *QuietWriter) writeData(data interface{}) {
	switch d := data.(type) {
	case *ListData:
		for _, item := range d.Items {
			fmt.Fprintf(w.out, "%s:%s\n", item.Type, item.Name)
		}
	case *BuildData:
		// Just output the count
		fmt.Fprintln(w.out, d.ItemCount)
	case *InfoData:
		fmt.Fprintf(w.out, "%s:%s\n", d.Type, d.Name)
	case *InstallData:
		for _, item := range d.Installed {
			fmt.Fprintf(w.out, "%s:%s\n", item.Type, item.Name)
		}
	case *ValidateData:
		if d.ErrorCount == 0 {
			fmt.Fprintln(w.out, "valid")
		} else {
			fmt.Fprintln(w.out, "invalid")
		}
	case []string:
		for _, s := range d {
			fmt.Fprintln(w.out, s)
		}
	case string:
		fmt.Fprintln(w.out, d)
	}
}

// WriteError writes an error to stderr.
func (w *QuietWriter) WriteError(err error) error {
	fmt.Fprintln(w.errOut, err.Error())
	return nil
}

// Success writes nothing in quiet mode.
func (w *QuietWriter) Success(message string) error {
	// Silent in quiet mode
	return nil
}

// Info writes nothing in quiet mode.
func (w *QuietWriter) Info(message string) error {
	// Silent in quiet mode
	return nil
}

// Warning writes warnings to stderr if verbose.
func (w *QuietWriter) Warning(message string) error {
	if w.verbose {
		fmt.Fprintln(w.errOut, "warning:", message)
	}
	return nil
}

// Error writes errors to stderr.
func (w *QuietWriter) Error(message string) error {
	fmt.Fprintln(w.errOut, "error:", message)
	return nil
}

// Table writes just the data cells, tab-separated.
func (w *QuietWriter) Table(headers []string, rows [][]string) error {
	for _, row := range rows {
		fmt.Fprintln(w.out, strings.Join(row, "\t"))
	}
	return nil
}

// List writes items one per line.
func (w *QuietWriter) List(items []string) error {
	for _, item := range items {
		fmt.Fprintln(w.out, item)
	}
	return nil
}

// Progress writes nothing in quiet mode.
func (w *QuietWriter) Progress(current, total int, message string) error {
	// Silent in quiet mode
	return nil
}
