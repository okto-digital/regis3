package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONWriter outputs responses as JSON.
// This format is ideal for consumption by LLM agents and scripts.
type JSONWriter struct {
	out     io.Writer
	errOut  io.Writer
	verbose bool
}

// NewJSONWriter creates a new JSON writer.
func NewJSONWriter(cfg *Config) *JSONWriter {
	return &JSONWriter{
		out:     cfg.Output,
		errOut:  cfg.ErrOutput,
		verbose: cfg.Verbose,
	}
}

// Write writes a response as JSON.
func (w *JSONWriter) Write(resp *Response) error {
	return w.writeJSON(w.out, resp)
}

// WriteError writes an error response as JSON.
func (w *JSONWriter) WriteError(err error) error {
	resp := &Response{
		Success: false,
		Error: &ErrorInfo{
			Message: err.Error(),
		},
	}
	return w.writeJSON(w.errOut, resp)
}

// Success writes a success message as JSON.
func (w *JSONWriter) Success(message string) error {
	resp := &Response{
		Success: true,
		Messages: []Message{
			{Level: LevelSuccess, Text: message},
		},
	}
	return w.writeJSON(w.out, resp)
}

// Info writes an info message as JSON.
func (w *JSONWriter) Info(message string) error {
	resp := &Response{
		Success: true,
		Messages: []Message{
			{Level: LevelInfo, Text: message},
		},
	}
	return w.writeJSON(w.out, resp)
}

// Warning writes a warning message as JSON.
func (w *JSONWriter) Warning(message string) error {
	resp := &Response{
		Success: true,
		Messages: []Message{
			{Level: LevelWarning, Text: message},
		},
	}
	return w.writeJSON(w.out, resp)
}

// Error writes an error message as JSON.
func (w *JSONWriter) Error(message string) error {
	resp := &Response{
		Success: false,
		Error: &ErrorInfo{
			Message: message,
		},
	}
	return w.writeJSON(w.errOut, resp)
}

// Table writes tabular data as JSON.
func (w *JSONWriter) Table(headers []string, rows [][]string) error {
	// Convert to array of objects
	data := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		obj := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				obj[header] = row[i]
			}
		}
		data = append(data, obj)
	}

	resp := &Response{
		Success: true,
		Data:    data,
	}
	return w.writeJSON(w.out, resp)
}

// List writes a list of items as JSON.
func (w *JSONWriter) List(items []string) error {
	resp := &Response{
		Success: true,
		Data:    items,
	}
	return w.writeJSON(w.out, resp)
}

// Progress writes a progress update as JSON.
func (w *JSONWriter) Progress(current, total int, message string) error {
	if !w.verbose {
		return nil
	}

	data := map[string]interface{}{
		"current": current,
		"total":   total,
		"percent": float64(current) / float64(total) * 100,
		"message": message,
	}

	resp := &Response{
		Success: true,
		Data:    data,
	}
	return w.writeJSON(w.out, resp)
}

// writeJSON writes a value as indented JSON.
func (w *JSONWriter) writeJSON(out io.Writer, v interface{}) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}
