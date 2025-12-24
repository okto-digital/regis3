package output

import (
	"time"
)

// Response is the standard response structure for all commands.
type Response struct {
	// Success indicates if the command succeeded.
	Success bool `json:"success"`

	// Command is the command that was executed.
	Command string `json:"command,omitempty"`

	// Data contains the command-specific response data.
	Data interface{} `json:"data,omitempty"`

	// Messages contains informational messages.
	Messages []Message `json:"messages,omitempty"`

	// Error contains error details if Success is false.
	Error *ErrorInfo `json:"error,omitempty"`

	// Duration is how long the command took.
	Duration time.Duration `json:"duration,omitempty"`
}

// Message represents an output message with severity.
type Message struct {
	Level   MessageLevel `json:"level"`
	Text    string       `json:"text"`
	Details string       `json:"details,omitempty"`
}

// MessageLevel indicates the severity of a message.
type MessageLevel string

const (
	LevelInfo    MessageLevel = "info"
	LevelWarning MessageLevel = "warning"
	LevelError   MessageLevel = "error"
	LevelSuccess MessageLevel = "success"
	LevelDebug   MessageLevel = "debug"
)

// ErrorInfo contains detailed error information.
type ErrorInfo struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Path    string `json:"path,omitempty"`
}

// NewResponse creates a new successful response.
func NewResponse(command string, data interface{}) *Response {
	return &Response{
		Success: true,
		Command: command,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(command string, err error) *Response {
	return &Response{
		Success: false,
		Command: command,
		Error: &ErrorInfo{
			Message: err.Error(),
		},
	}
}

// WithDuration adds duration to the response.
func (r *Response) WithDuration(d time.Duration) *Response {
	r.Duration = d
	return r
}

// WithMessage adds a message to the response.
func (r *Response) WithMessage(level MessageLevel, text string) *Response {
	r.Messages = append(r.Messages, Message{
		Level: level,
		Text:  text,
	})
	return r
}

// WithInfo adds an info message.
func (r *Response) WithInfo(text string) *Response {
	return r.WithMessage(LevelInfo, text)
}

// WithWarning adds a warning message.
func (r *Response) WithWarning(text string) *Response {
	return r.WithMessage(LevelWarning, text)
}

// WithError sets the error on the response.
func (r *Response) WithError(code, message, details string) *Response {
	r.Success = false
	r.Error = &ErrorInfo{
		Code:    code,
		Message: message,
		Details: details,
	}
	return r
}

// BuildResponse is a builder for constructing responses.
type BuildResponse struct {
	resp *Response
}

// NewBuilder creates a new response builder.
func NewBuilder(command string) *BuildResponse {
	return &BuildResponse{
		resp: &Response{
			Success: true,
			Command: command,
		},
	}
}

// Data sets the response data.
func (b *BuildResponse) Data(data interface{}) *BuildResponse {
	b.resp.Data = data
	return b
}

// Duration sets the duration.
func (b *BuildResponse) Duration(d time.Duration) *BuildResponse {
	b.resp.Duration = d
	return b
}

// Info adds an info message.
func (b *BuildResponse) Info(text string) *BuildResponse {
	b.resp.WithInfo(text)
	return b
}

// Warning adds a warning message.
func (b *BuildResponse) Warning(text string) *BuildResponse {
	b.resp.WithWarning(text)
	return b
}

// Error sets an error on the response.
func (b *BuildResponse) Error(err error) *BuildResponse {
	b.resp.Success = false
	b.resp.Error = &ErrorInfo{
		Message: err.Error(),
	}
	return b
}

// ErrorWithCode sets an error with a code.
func (b *BuildResponse) ErrorWithCode(code string, err error) *BuildResponse {
	b.resp.Success = false
	b.resp.Error = &ErrorInfo{
		Code:    code,
		Message: err.Error(),
	}
	return b
}

// Build returns the constructed response.
func (b *BuildResponse) Build() *Response {
	return b.resp
}

// Common data structures for responses

// ListData is the response data for list commands.
type ListData struct {
	Items []ListItem `json:"items"`
	Total int        `json:"total"`
}

// ListItem represents an item in a list.
type ListItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// BuildData is the response data for build commands.
type BuildData struct {
	Items    int           `json:"items"`
	Errors   int           `json:"errors"`
	Warnings int           `json:"warnings"`
	Duration time.Duration `json:"duration"`
	Path     string        `json:"path,omitempty"`
}

// InfoData is the response data for info commands.
type InfoData struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	Tags       []string `json:"tags,omitempty"`
	Deps       []string `json:"deps,omitempty"`
	Dependents []string `json:"dependents,omitempty"`
	Status     string   `json:"status,omitempty"`
	Source     string   `json:"source,omitempty"`
}

// InstallData is the response data for install/add commands.
type InstallData struct {
	Installed []string `json:"installed"`
	Skipped   []string `json:"skipped,omitempty"`
	Total     int      `json:"total"`
}

// ValidateData is the response data for validate commands.
type ValidateData struct {
	Valid    bool            `json:"valid"`
	Errors   []ValidateIssue `json:"errors,omitempty"`
	Warnings []ValidateIssue `json:"warnings,omitempty"`
}

// ValidateIssue represents a validation issue.
type ValidateIssue struct {
	Path    string `json:"path"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}
