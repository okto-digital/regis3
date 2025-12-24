package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles for pretty output
var (
	// Colors
	colorSuccess = lipgloss.Color("42")  // Green
	colorError   = lipgloss.Color("196") // Red
	colorWarning = lipgloss.Color("214") // Orange
	colorInfo    = lipgloss.Color("39")  // Blue
	colorMuted   = lipgloss.Color("245") // Gray
	colorAccent  = lipgloss.Color("213") // Pink/Purple

	// Text styles
	styleSuccess = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	styleError   = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	styleWarning = lipgloss.NewStyle().Foreground(colorWarning).Bold(true)
	styleInfo    = lipgloss.NewStyle().Foreground(colorInfo)
	styleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	styleAccent  = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleBold    = lipgloss.NewStyle().Bold(true)

	// Icons
	iconSuccess  = styleSuccess.Render("✓")
	iconError    = styleError.Render("✗")
	iconWarning  = styleWarning.Render("⚠")
	iconInfo     = styleInfo.Render("ℹ")
	iconArrow    = styleMuted.Render("→")
	iconBullet   = styleMuted.Render("•")
	iconProgress = styleInfo.Render("⋯")
)

// PrettyWriter outputs human-friendly formatted text.
type PrettyWriter struct {
	out     io.Writer
	errOut  io.Writer
	noColor bool
	verbose bool
}

// NewPrettyWriter creates a new pretty writer.
func NewPrettyWriter(cfg *Config) *PrettyWriter {
	return &PrettyWriter{
		out:     cfg.Output,
		errOut:  cfg.ErrOutput,
		noColor: cfg.NoColor,
		verbose: cfg.Verbose,
	}
}

// Write writes a formatted response.
func (w *PrettyWriter) Write(resp *Response) error {
	// Write messages first
	for _, msg := range resp.Messages {
		switch msg.Level {
		case LevelSuccess:
			w.Success(msg.Text)
		case LevelError:
			w.Error(msg.Text)
		case LevelWarning:
			w.Warning(msg.Text)
		case LevelInfo:
			w.Info(msg.Text)
		case LevelDebug:
			if w.verbose {
				w.writeLine(w.out, "%s %s", styleMuted.Render("[debug]"), msg.Text)
			}
		}
	}

	// Write error if present
	if resp.Error != nil {
		w.writeError(resp.Error)
	}

	// Write data based on type
	if resp.Data != nil {
		w.writeData(resp.Data)
	}

	// Write duration if present and verbose
	if w.verbose && resp.Duration > 0 {
		w.writeLine(w.out, "%s Completed in %v", styleMuted.Render("⏱"), resp.Duration)
	}

	return nil
}

// writeData writes response data based on its type.
func (w *PrettyWriter) writeData(data interface{}) {
	switch d := data.(type) {
	case *ListData:
		w.writeListData(d)
	case ListData:
		w.writeListData(&d)
	case *BuildData:
		w.writeBuildData(d)
	case BuildData:
		w.writeBuildData(&d)
	case *InfoData:
		w.writeInfoData(d)
	case InfoData:
		w.writeInfoData(&d)
	case *InstallData:
		w.writeInstallData(d)
	case InstallData:
		w.writeInstallData(&d)
	case *ValidateData:
		w.writeValidateData(d)
	case ValidateData:
		w.writeValidateData(&d)
	case *StatusData:
		w.writeStatusData(d)
	case StatusData:
		w.writeStatusData(&d)
	case *ScanData:
		w.writeScanData(d)
	case ScanData:
		w.writeScanData(&d)
	case *ImportData:
		w.writeImportData(d)
	case ImportData:
		w.writeImportData(&d)
	case *UpdateData:
		w.writeUpdateData(d)
	case UpdateData:
		w.writeUpdateData(&d)
	case *OrphansData:
		w.writeOrphansData(d)
	case OrphansData:
		w.writeOrphansData(&d)
	case *ConfigData:
		w.writeConfigData(d)
	case ConfigData:
		w.writeConfigData(&d)
	case []string:
		w.List(d)
	case map[string]interface{}:
		w.writeMap(d)
	}
}

// writeListData writes list response data.
func (w *PrettyWriter) writeListData(data *ListData) {
	if len(data.Items) == 0 {
		w.Info("No items found")
		return
	}

	for _, item := range data.Items {
		typeStyle := w.getTypeStyle(item.Type)
		w.writeLine(w.out, "%s %s",
			typeStyle.Render(item.Type+":"+item.Name),
			styleMuted.Render(item.Desc))
	}

	w.writeLine(w.out, "")
	w.writeLine(w.out, "%s %d items", styleMuted.Render("Total:"), data.TotalCount)
}

// writeBuildData writes build response data.
func (w *PrettyWriter) writeBuildData(data *BuildData) {
	w.writeLine(w.out, "")
	w.writeLine(w.out, "%s Build complete", iconSuccess)
	w.writeLine(w.out, "   Items:    %d", data.ItemCount)
	w.writeLine(w.out, "   Path:     %s", data.ManifestPath)
	w.writeLine(w.out, "   Duration: %s", data.Duration)
}

// writeInfoData writes info response data.
func (w *PrettyWriter) writeInfoData(data *InfoData) {
	typeStyle := w.getTypeStyle(data.Type)

	w.writeLine(w.out, "%s", typeStyle.Render(data.Type+":"+data.Name))
	w.writeLine(w.out, "")
	w.writeLine(w.out, "%s", data.Desc)
	w.writeLine(w.out, "")

	if len(data.Tags) > 0 {
		tags := make([]string, len(data.Tags))
		for i, tag := range data.Tags {
			tags[i] = styleMuted.Render("#" + tag)
		}
		w.writeLine(w.out, "Tags: %s", strings.Join(tags, " "))
	}

	if len(data.Dependencies) > 0 {
		w.writeLine(w.out, "Dependencies:")
		for _, dep := range data.Dependencies {
			w.writeLine(w.out, "  %s %s", iconArrow, dep)
		}
	}

	if len(data.Files) > 0 {
		w.writeLine(w.out, "Files:")
		for _, f := range data.Files {
			w.writeLine(w.out, "  %s %s", iconBullet, f)
		}
	}

	if data.Path != "" {
		w.writeLine(w.out, "Source: %s", styleMuted.Render(data.Path))
	}
}

// writeInstallData writes install response data.
func (w *PrettyWriter) writeInstallData(data *InstallData) {
	if len(data.Installed) > 0 {
		w.writeLine(w.out, "%s Installed:", iconSuccess)
		for _, item := range data.Installed {
			typeStyle := w.getTypeStyle(item.Type)
			w.writeLine(w.out, "  %s %s", iconArrow, typeStyle.Render(item.Type+":"+item.Name))
		}
	}

	if len(data.Skipped) > 0 {
		w.writeLine(w.out, "%s Skipped:", iconWarning)
		for _, item := range data.Skipped {
			w.writeLine(w.out, "  %s %s", iconBullet, styleMuted.Render(item))
		}
	}

	if data.DryRun {
		w.writeLine(w.out, "")
		w.writeLine(w.out, "%s (dry run - no changes made)", styleMuted.Render("Note:"))
	}
}

// writeValidateData writes validate response data.
func (w *PrettyWriter) writeValidateData(data *ValidateData) {
	w.writeLine(w.out, "")
	w.writeLine(w.out, "Validated %d items", data.ItemCount)

	if data.ErrorCount > 0 {
		w.writeLine(w.out, "  %s Errors:   %d", iconError, data.ErrorCount)
	}
	if data.WarnCount > 0 {
		w.writeLine(w.out, "  %s Warnings: %d", iconWarning, data.WarnCount)
	}
	if data.InfoCount > 0 {
		w.writeLine(w.out, "  %s Info:     %d", iconInfo, data.InfoCount)
	}

	if data.ErrorCount == 0 && data.WarnCount == 0 {
		w.writeLine(w.out, "  %s All items valid", iconSuccess)
	}
}

// writeStatusData writes status response data.
func (w *PrettyWriter) writeStatusData(data *StatusData) {
	if len(data.Items) == 0 {
		w.Info("No items installed")
		return
	}

	w.writeLine(w.out, "")
	w.writeLine(w.out, "Installed items (%s):", styleMuted.Render(data.Target))
	for _, item := range data.Items {
		typeStyle := w.getTypeStyle(item.Type)
		status := ""
		if item.NeedsUpdate {
			status = " " + styleWarning.Render("[update available]")
		}
		w.writeLine(w.out, "  %s %s%s",
			iconBullet,
			typeStyle.Render(item.Type+":"+item.Name),
			status)
	}
}

// writeScanData writes scan response data.
func (w *PrettyWriter) writeScanData(data *ScanData) {
	if len(data.Imported) > 0 {
		if data.DryRun {
			w.writeLine(w.out, "%s Would import:", iconInfo)
		} else {
			w.writeLine(w.out, "%s Imported:", iconSuccess)
		}
		for _, item := range data.Imported {
			typeStyle := w.getTypeStyle(item.Type)
			w.writeLine(w.out, "  %s %s", iconArrow, typeStyle.Render(item.Type+":"+item.Name))
		}
	}

	if len(data.Staged) > 0 {
		if data.DryRun {
			w.writeLine(w.out, "%s Would stage:", iconWarning)
		} else {
			w.writeLine(w.out, "%s Staged (need regis3 frontmatter):", iconWarning)
		}
		for _, item := range data.Staged {
			w.writeLine(w.out, "  %s %s", iconBullet, styleMuted.Render(item.SourcePath))
		}
	}

	if len(data.Errors) > 0 {
		w.writeLine(w.out, "%s Errors:", iconError)
		for _, e := range data.Errors {
			w.writeLine(w.out, "  %s %s", iconBullet, styleError.Render(e))
		}
	}

	if data.DryRun {
		w.writeLine(w.out, "")
		w.writeLine(w.out, "%s (dry run - no changes made)", styleMuted.Render("Note:"))
	}
}

// writeImportData writes import response data.
func (w *PrettyWriter) writeImportData(data *ImportData) {
	if len(data.Processed) > 0 {
		w.writeLine(w.out, "%s Processed:", iconSuccess)
		for _, item := range data.Processed {
			typeStyle := w.getTypeStyle(item.Type)
			w.writeLine(w.out, "  %s %s", iconArrow, typeStyle.Render(item.Type+":"+item.Name))
		}
	}

	if len(data.Pending) > 0 {
		w.writeLine(w.out, "%s Pending (need regis3 frontmatter):", iconWarning)
		for _, item := range data.Pending {
			w.writeLine(w.out, "  %s %s", iconBullet, styleMuted.Render(item.Path))
		}
	}

	if len(data.Errors) > 0 {
		w.writeLine(w.out, "%s Errors:", iconError)
		for _, e := range data.Errors {
			w.writeLine(w.out, "  %s %s", iconBullet, styleError.Render(e))
		}
	}
}

// writeUpdateData writes update response data.
func (w *PrettyWriter) writeUpdateData(data *UpdateData) {
	if data.Updated {
		w.writeLine(w.out, "%s Registry updated", iconSuccess)
	} else {
		w.writeLine(w.out, "%s Already up to date", iconInfo)
	}
	w.writeLine(w.out, "   Items: %d", data.ItemCount)
}

// writeOrphansData writes orphans response data.
func (w *PrettyWriter) writeOrphansData(data *OrphansData) {
	if len(data.Orphans) == 0 {
		w.writeLine(w.out, "%s No orphaned files found", iconSuccess)
		return
	}

	w.writeLine(w.out, "%s Found %d orphaned files:", iconWarning, data.Count)
	for _, orphan := range data.Orphans {
		w.writeLine(w.out, "  %s %s", iconBullet, styleMuted.Render(orphan.Path))
		if orphan.Reason != "" {
			w.writeLine(w.out, "      %s", styleMuted.Render(orphan.Reason))
		}
	}
}

// writeConfigData writes config response data.
func (w *PrettyWriter) writeConfigData(data *ConfigData) {
	w.writeLine(w.out, "Config: %s", styleMuted.Render(data.Path))
	w.writeLine(w.out, "")
	for key, value := range data.Settings {
		w.writeLine(w.out, "  %s: %s", styleBold.Render(key), value)
	}
}

// writeMap writes a map as key-value pairs.
func (w *PrettyWriter) writeMap(data map[string]interface{}) {
	for k, v := range data {
		w.writeLine(w.out, "%s: %v", styleBold.Render(k), v)
	}
}

// writeError writes an error.
func (w *PrettyWriter) writeError(err *ErrorInfo) {
	w.writeLine(w.errOut, "%s %s", iconError, styleError.Render(err.Message))
	if err.Details != "" {
		w.writeLine(w.errOut, "   %s", styleMuted.Render(err.Details))
	}
	if err.Path != "" {
		w.writeLine(w.errOut, "   at %s", styleMuted.Render(err.Path))
	}
}

// WriteError writes an error response.
func (w *PrettyWriter) WriteError(err error) error {
	w.writeLine(w.errOut, "%s %s", iconError, styleError.Render(err.Error()))
	return nil
}

// Success writes a success message.
func (w *PrettyWriter) Success(message string) error {
	w.writeLine(w.out, "%s %s", iconSuccess, message)
	return nil
}

// Info writes an info message.
func (w *PrettyWriter) Info(message string) error {
	w.writeLine(w.out, "%s %s", iconInfo, message)
	return nil
}

// Warning writes a warning message.
func (w *PrettyWriter) Warning(message string) error {
	w.writeLine(w.out, "%s %s", iconWarning, styleWarning.Render(message))
	return nil
}

// Error writes an error message.
func (w *PrettyWriter) Error(message string) error {
	w.writeLine(w.errOut, "%s %s", iconError, styleError.Render(message))
	return nil
}

// Table writes a formatted table.
func (w *PrettyWriter) Table(headers []string, rows [][]string) error {
	if len(rows) == 0 {
		return nil
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	headerLine := ""
	for i, h := range headers {
		headerLine += fmt.Sprintf("%-*s  ", widths[i], styleBold.Render(h))
	}
	w.writeLine(w.out, "%s", headerLine)

	// Print separator
	sepLine := ""
	for _, width := range widths {
		sepLine += strings.Repeat("─", width) + "  "
	}
	w.writeLine(w.out, "%s", styleMuted.Render(sepLine))

	// Print rows
	for _, row := range rows {
		rowLine := ""
		for i, cell := range row {
			if i < len(widths) {
				rowLine += fmt.Sprintf("%-*s  ", widths[i], cell)
			}
		}
		w.writeLine(w.out, "%s", rowLine)
	}

	return nil
}

// List writes a bulleted list.
func (w *PrettyWriter) List(items []string) error {
	for _, item := range items {
		w.writeLine(w.out, "  %s %s", iconBullet, item)
	}
	return nil
}

// Progress writes a progress indicator.
func (w *PrettyWriter) Progress(current, total int, message string) error {
	percent := float64(current) / float64(total) * 100
	bar := w.progressBar(current, total, 20)
	w.writeLine(w.out, "\r%s %s %3.0f%% %s", iconProgress, bar, percent, message)
	return nil
}

// progressBar creates a text progress bar.
func (w *PrettyWriter) progressBar(current, total, width int) string {
	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return styleInfo.Render("[") + bar + styleInfo.Render("]")
}

// getTypeStyle returns the style for an item type.
func (w *PrettyWriter) getTypeStyle(itemType string) lipgloss.Style {
	switch itemType {
	case "skill":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	case "subagent":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true)
	case "philosophy":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	case "stack":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	case "command":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	default:
		return styleBold
	}
}

// getStatusStyle returns the style for a status.
func (w *PrettyWriter) getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "stable":
		return styleSuccess
	case "draft":
		return styleWarning
	case "deprecated":
		return styleError
	default:
		return styleMuted
	}
}

// writeLine writes a formatted line to the writer.
func (w *PrettyWriter) writeLine(out io.Writer, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if w.noColor {
		// Strip ANSI codes if no color
		line = stripAnsi(line)
	}
	fmt.Fprintln(out, line)
}

// stripAnsi removes ANSI escape codes from a string.
func stripAnsi(s string) string {
	// Simple ANSI stripper - handles common escape sequences
	result := strings.Builder{}
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
