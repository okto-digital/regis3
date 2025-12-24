package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"quiet", FormatQuiet},
		{"pretty", FormatPretty},
		{"", FormatPretty},
		{"unknown", FormatPretty},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNew(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("json", func(t *testing.T) {
		w := New(FormatJSON, cfg)
		_, ok := w.(*JSONWriter)
		assert.True(t, ok)
	})

	t.Run("quiet", func(t *testing.T) {
		w := New(FormatQuiet, cfg)
		_, ok := w.(*QuietWriter)
		assert.True(t, ok)
	})

	t.Run("pretty", func(t *testing.T) {
		w := New(FormatPretty, cfg)
		_, ok := w.(*PrettyWriter)
		assert.True(t, ok)
	})

	t.Run("nil config", func(t *testing.T) {
		w := New(FormatJSON, nil)
		assert.NotNil(t, w)
	})
}

func TestNewWithFormat(t *testing.T) {
	w := NewWithFormat("json", nil)
	_, ok := w.(*JSONWriter)
	assert.True(t, ok)
}

func TestResponse(t *testing.T) {
	t.Run("NewResponse", func(t *testing.T) {
		resp := NewResponse("test", map[string]string{"key": "value"})
		assert.True(t, resp.Success)
		assert.Equal(t, "test", resp.Command)
		assert.NotNil(t, resp.Data)
	})

	t.Run("NewErrorResponse", func(t *testing.T) {
		resp := NewErrorResponse("test", errors.New("test error"))
		assert.False(t, resp.Success)
		assert.Equal(t, "test", resp.Command)
		assert.Equal(t, "test error", resp.Error.Message)
	})

	t.Run("WithDuration", func(t *testing.T) {
		resp := NewResponse("test", nil).WithDuration(5 * time.Second)
		assert.Equal(t, 5*time.Second, resp.Duration)
	})

	t.Run("WithMessage", func(t *testing.T) {
		resp := NewResponse("test", nil).
			WithInfo("info msg").
			WithWarning("warning msg")
		assert.Len(t, resp.Messages, 2)
		assert.Equal(t, LevelInfo, resp.Messages[0].Level)
		assert.Equal(t, LevelWarning, resp.Messages[1].Level)
	})

	t.Run("WithError", func(t *testing.T) {
		resp := NewResponse("test", nil).WithError("ERR001", "something failed", "details here")
		assert.False(t, resp.Success)
		assert.Equal(t, "ERR001", resp.Error.Code)
		assert.Equal(t, "something failed", resp.Error.Message)
		assert.Equal(t, "details here", resp.Error.Details)
	})
}

func TestBuilder(t *testing.T) {
	resp := NewBuilder("test").
		Data(map[string]string{"key": "value"}).
		Duration(2 * time.Second).
		Info("info message").
		Warning("warning message").
		Build()

	assert.True(t, resp.Success)
	assert.Equal(t, "test", resp.Command)
	assert.Equal(t, 2*time.Second, resp.Duration)
	assert.Len(t, resp.Messages, 2)
	assert.NotNil(t, resp.Data)
}

func TestBuilderError(t *testing.T) {
	resp := NewBuilder("test").
		Error(errors.New("failed")).
		Build()

	assert.False(t, resp.Success)
	assert.Equal(t, "failed", resp.Error.Message)
}

func TestBuilderErrorWithCode(t *testing.T) {
	resp := NewBuilder("test").
		ErrorWithCode("ERR001", errors.New("failed")).
		Build()

	assert.False(t, resp.Success)
	assert.Equal(t, "ERR001", resp.Error.Code)
	assert.Equal(t, "failed", resp.Error.Message)
}

// JSON Writer Tests

func TestJSONWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewJSONWriter(cfg)

	resp := NewResponse("test", map[string]string{"key": "value"})
	err := w.Write(resp)
	require.NoError(t, err)

	var result Response
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "test", result.Command)
}

func TestJSONWriter_WriteError(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewJSONWriter(cfg)

	err := w.WriteError(errors.New("test error"))
	require.NoError(t, err)

	var result Response
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.False(t, result.Success)
	assert.Equal(t, "test error", result.Error.Message)
}

func TestJSONWriter_Success(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewJSONWriter(cfg)

	err := w.Success("operation completed")
	require.NoError(t, err)

	var result Response
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Len(t, result.Messages, 1)
	assert.Equal(t, LevelSuccess, result.Messages[0].Level)
}

func TestJSONWriter_Table(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewJSONWriter(cfg)

	headers := []string{"Name", "Type"}
	rows := [][]string{
		{"item1", "skill"},
		{"item2", "subagent"},
	}

	err := w.Table(headers, rows)
	require.NoError(t, err)

	var result Response
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	// Data should be array of objects
	data, ok := result.Data.([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
}

func TestJSONWriter_List(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewJSONWriter(cfg)

	err := w.List([]string{"item1", "item2", "item3"})
	require.NoError(t, err)

	var result Response
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	data, ok := result.Data.([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 3)
}

func TestJSONWriter_Progress(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, Verbose: true}
	w := NewJSONWriter(cfg)

	err := w.Progress(50, 100, "processing")
	require.NoError(t, err)

	var result Response
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.True(t, result.Success)
}

func TestJSONWriter_Progress_NotVerbose(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, Verbose: false}
	w := NewJSONWriter(cfg)

	err := w.Progress(50, 100, "processing")
	require.NoError(t, err)

	// Should not output anything when not verbose
	assert.Empty(t, buf.String())
}

// Pretty Writer Tests

func TestPrettyWriter_Success(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, NoColor: true}
	w := NewPrettyWriter(cfg)

	err := w.Success("operation completed")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "operation completed")
	assert.Contains(t, output, "✓")
}

func TestPrettyWriter_Error(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, NoColor: true}
	w := NewPrettyWriter(cfg)

	err := w.Error("something failed")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "something failed")
	assert.Contains(t, output, "✗")
}

func TestPrettyWriter_Warning(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, NoColor: true}
	w := NewPrettyWriter(cfg)

	err := w.Warning("be careful")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "be careful")
	assert.Contains(t, output, "⚠")
}

func TestPrettyWriter_Info(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, NoColor: true}
	w := NewPrettyWriter(cfg)

	err := w.Info("just so you know")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "just so you know")
	assert.Contains(t, output, "ℹ")
}

func TestPrettyWriter_Table(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, NoColor: true}
	w := NewPrettyWriter(cfg)

	headers := []string{"Name", "Type"}
	rows := [][]string{
		{"git-conventions", "skill"},
		{"architect", "subagent"},
	}

	err := w.Table(headers, rows)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Type")
	assert.Contains(t, output, "git-conventions")
	assert.Contains(t, output, "architect")
	assert.Contains(t, output, "─") // separator
}

func TestPrettyWriter_List(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, NoColor: true}
	w := NewPrettyWriter(cfg)

	err := w.List([]string{"item1", "item2", "item3"})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "item1")
	assert.Contains(t, output, "item2")
	assert.Contains(t, output, "item3")
	assert.Contains(t, output, "•")
}

func TestPrettyWriter_WriteResponse(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf, NoColor: true}
	w := NewPrettyWriter(cfg)

	resp := NewResponse("test", nil).
		WithInfo("info message").
		WithWarning("warning message")

	err := w.Write(resp)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warning message")
}

func TestPrettyWriter_StripAnsi(t *testing.T) {
	input := "\x1b[31mred text\x1b[0m"
	result := stripAnsi(input)
	assert.Equal(t, "red text", result)
}

// Quiet Writer Tests

func TestQuietWriter_Success(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewQuietWriter(cfg)

	err := w.Success("operation completed")
	require.NoError(t, err)

	// Quiet mode should not output success messages
	assert.Empty(t, buf.String())
}

func TestQuietWriter_Error(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewQuietWriter(cfg)

	err := w.Error("something failed")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "error:")
	assert.Contains(t, output, "something failed")
}

func TestQuietWriter_List(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewQuietWriter(cfg)

	err := w.List([]string{"item1", "item2", "item3"})
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 3)
	assert.Equal(t, "item1", lines[0])
	assert.Equal(t, "item2", lines[1])
	assert.Equal(t, "item3", lines[2])
}

func TestQuietWriter_Table(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewQuietWriter(cfg)

	headers := []string{"Name", "Type"}
	rows := [][]string{
		{"item1", "skill"},
		{"item2", "subagent"},
	}

	err := w.Table(headers, rows)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 2)
	assert.Equal(t, "item1\tskill", lines[0])
	assert.Equal(t, "item2\tsubagent", lines[1])
}

func TestQuietWriter_WriteResponse(t *testing.T) {
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &errBuf}
	w := NewQuietWriter(cfg)

	resp := &Response{
		Success: true,
		Data:    []string{"item1", "item2"},
	}

	err := w.Write(resp)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 2)
}

func TestQuietWriter_WriteResponseWithError(t *testing.T) {
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &errBuf}
	w := NewQuietWriter(cfg)

	resp := &Response{
		Success: false,
		Error: &ErrorInfo{
			Message: "test error",
		},
	}

	err := w.Write(resp)
	require.NoError(t, err)

	assert.Contains(t, errBuf.String(), "test error")
}

func TestQuietWriter_Progress(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &buf}
	w := NewQuietWriter(cfg)

	err := w.Progress(50, 100, "processing")
	require.NoError(t, err)

	// Quiet mode should not output progress
	assert.Empty(t, buf.String())
}

func TestQuietWriter_Warning_Verbose(t *testing.T) {
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &errBuf, Verbose: true}
	w := NewQuietWriter(cfg)

	err := w.Warning("be careful")
	require.NoError(t, err)

	assert.Contains(t, errBuf.String(), "warning:")
	assert.Contains(t, errBuf.String(), "be careful")
}

func TestQuietWriter_Warning_NotVerbose(t *testing.T) {
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	cfg := &Config{Output: &buf, ErrOutput: &errBuf, Verbose: false}
	w := NewQuietWriter(cfg)

	err := w.Warning("be careful")
	require.NoError(t, err)

	// Should not output warnings when not verbose
	assert.Empty(t, errBuf.String())
}
