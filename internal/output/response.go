package output

import (
	"fmt"
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

// NewResponseBuilder creates a new response builder (alias for NewBuilder).
func NewResponseBuilder(command string) *ResponseBuilder {
	return &ResponseBuilder{
		resp: &Response{
			Success: true,
			Command: command,
		},
	}
}

// ResponseBuilder builds responses with a fluent interface.
type ResponseBuilder struct {
	resp *Response
}

// WithSuccess sets the success status.
func (b *ResponseBuilder) WithSuccess(success bool) *ResponseBuilder {
	b.resp.Success = success
	return b
}

// WithData sets the response data.
func (b *ResponseBuilder) WithData(data interface{}) *ResponseBuilder {
	b.resp.Data = data
	return b
}

// WithInfo adds an info message with formatting.
func (b *ResponseBuilder) WithInfo(format string, args ...interface{}) *ResponseBuilder {
	text := format
	if len(args) > 0 {
		text = fmt.Sprintf(format, args...)
	}
	b.resp.Messages = append(b.resp.Messages, Message{
		Level: LevelInfo,
		Text:  text,
	})
	return b
}

// WithWarning adds a warning message with formatting.
func (b *ResponseBuilder) WithWarning(format string, args ...interface{}) *ResponseBuilder {
	text := format
	if len(args) > 0 {
		text = fmt.Sprintf(format, args...)
	}
	b.resp.Messages = append(b.resp.Messages, Message{
		Level: LevelWarning,
		Text:  text,
	})
	return b
}

// WithError adds an error message.
func (b *ResponseBuilder) WithError(path, message string) *ResponseBuilder {
	b.resp.Messages = append(b.resp.Messages, Message{
		Level:   LevelError,
		Text:    message,
		Details: path,
	})
	return b
}

// Build returns the constructed response.
func (b *ResponseBuilder) Build() *Response {
	return b.resp
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
	Items      []ListItem `json:"items"`
	TotalCount int        `json:"total_count"`
	Filtered   bool       `json:"filtered,omitempty"`
}

// ListItem represents an item in a list.
type ListItem struct {
	Type string   `json:"type"`
	Name string   `json:"name"`
	Desc string   `json:"desc"`
	Tags []string `json:"tags,omitempty"`
}

// BuildData is the response data for build commands.
type BuildData struct {
	ItemCount    int    `json:"item_count"`
	ManifestPath string `json:"manifest_path"`
	Duration     string `json:"duration"`
}

// InfoData is the response data for info commands.
type InfoData struct {
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Desc         string   `json:"desc"`
	Path         string   `json:"path"`
	Tags         []string `json:"tags,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Files        []string `json:"files,omitempty"`
}

// InstallData is the response data for install/add commands.
type InstallData struct {
	Installed []InstalledItem `json:"installed"`
	Skipped   []string        `json:"skipped,omitempty"`
	Target    string          `json:"target"`
	DryRun    bool            `json:"dry_run,omitempty"`
}

// InstalledItem represents an installed item.
type InstalledItem struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	DestPath string `json:"dest_path"`
}

// RemoveData is the response data for remove commands.
type RemoveData struct {
	Removed  []InstalledItem `json:"removed"`
	NotFound []string        `json:"not_found,omitempty"`
	DryRun   bool            `json:"dry_run,omitempty"`
}

// StatusData is the response data for status commands.
type StatusData struct {
	Items  []StatusItem `json:"items"`
	Target string       `json:"target"`
}

// StatusItem represents an installed item's status.
type StatusItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	InstalledAt string `json:"installed_at"`
	DestPath    string `json:"dest_path"`
	NeedsUpdate bool   `json:"needs_update,omitempty"`
}

// ValidateData is the response data for validate commands.
type ValidateData struct {
	ItemCount  int `json:"item_count"`
	ErrorCount int `json:"error_count"`
	WarnCount  int `json:"warn_count"`
	InfoCount  int `json:"info_count"`
}

// ScanData is the response data for scan commands.
type ScanData struct {
	Imported []ImportedItem `json:"imported"`
	Staged   []ImportedItem `json:"staged"`
	Errors   []string       `json:"errors,omitempty"`
	DryRun   bool           `json:"dry_run,omitempty"`
}

// ImportedItem represents an imported file.
type ImportedItem struct {
	SourcePath string `json:"source_path"`
	DestPath   string `json:"dest_path"`
	Type       string `json:"type"`
	Name       string `json:"name"`
}

// ImportData is the response data for import commands.
type ImportData struct {
	Processed []ImportedItem `json:"processed"`
	Pending   []PendingItem  `json:"pending"`
	Errors    []string       `json:"errors,omitempty"`
}

// PendingItem represents a file pending in staging.
type PendingItem struct {
	Path          string `json:"path"`
	SuggestedType string `json:"suggested_type"`
	SuggestedName string `json:"suggested_name"`
	Confidence    int    `json:"confidence"`
}

// UpdateData is the response data for update commands.
type UpdateData struct {
	Updated   bool   `json:"updated"`
	ItemCount int    `json:"item_count"`
	GitOutput string `json:"git_output,omitempty"`
}

// OrphansData is the response data for orphans commands.
type OrphansData struct {
	Orphans []OrphanFile `json:"orphans"`
	Count   int          `json:"count"`
}

// OrphanFile represents a file not referenced in the manifest.
type OrphanFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Reason string `json:"reason"`
}

// ConfigData is the response data for config commands.
type ConfigData struct {
	Path     string            `json:"path"`
	Settings map[string]string `json:"settings"`
}
