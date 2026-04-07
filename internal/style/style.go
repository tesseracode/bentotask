// Package style provides terminal colors and formatting for BentoTask output.
//
// Color scheme follows ADR-003 §4. Colors are adaptive (light/dark terminal)
// and respect NO_COLOR, --no-color, and piped output automatically via lipgloss.
package style

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// --- Priority colors ---

// DisableColor forces all styling to produce plain text (no ANSI escapes).
// Called when --no-color flag is set.
func DisableColor() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// Priority returns a styled string for the given priority level.
func Priority(p string) string {
	switch p {
	case "urgent":
		return urgentStyle.Render(p)
	case "high":
		return highStyle.Render(p)
	case "medium":
		return mediumStyle.Render(p)
	case "low":
		return lowStyle.Render(p)
	case "none", "":
		return dimStyle.Render("none")
	default:
		return p
	}
}

// --- Status colors ---

// Status returns a styled string for the given status.
func Status(s string) string {
	switch s {
	case "done":
		return doneStyle.Render("✓ " + s)
	case "active":
		return activeStyle.Render("● " + s)
	case "pending":
		return pendingStyle.Render("○ " + s)
	case "blocked", "waiting":
		return blockedStyle.Render("✗ " + s)
	case "paused":
		return dimStyle.Render("◌ " + s)
	case "cancelled":
		return dimStyle.Render("⊘ " + s)
	default:
		return s
	}
}

// --- Energy indicator ---

// Energy returns a styled string for the given energy level.
func Energy(e string) string {
	switch e {
	case "high":
		return energyHighStyle.Render("⚡ " + e)
	case "medium":
		return energyMedStyle.Render("~ " + e)
	case "low":
		return energyLowStyle.Render("· " + e)
	default:
		return dimStyle.Render("-")
	}
}

// --- Tags ---

// Tag returns a styled tag string.
func Tag(t string) string {
	return tagStyle.Render("#" + t)
}

// --- Due dates ---

// DueDate returns a styled due date string.
// Pass overdue=true to highlight in red.
func DueDate(date string, overdue bool) string {
	if overdue {
		return overdueStyle.Render(date)
	}
	return date
}

// --- General styles ---

// Success returns a green success message.
func Success(msg string) string {
	return successStyle.Render("✓ " + msg)
}

// ErrorMsg returns a red error message.
func ErrorMsg(msg string) string {
	return errorStyle.Render("Error: " + msg)
}

// Dim returns a dimmed/gray string.
func Dim(s string) string {
	return dimStyle.Render(s)
}

// Bold returns a bold string.
func Bold(s string) string {
	return boldStyle.Render(s)
}

// Header returns a styled section header.
func Header(s string) string {
	return headerStyle.Render(s)
}

// --- Style definitions (ADR-003 §4 color scheme) ---

var (
	// Priority
	urgentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red
	highStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	mediumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	lowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Gray

	// Status
	doneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // Cyan
	pendingStyle = lipgloss.NewStyle()                                  // Default
	blockedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red

	// Energy
	energyHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	energyMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	energyLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Gray

	// Tags
	tagStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // Cyan

	// Due dates
	overdueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red bold

	// General
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Gray
	boldStyle    = lipgloss.NewStyle().Bold(true)
	headerStyle  = lipgloss.NewStyle().Bold(true).Underline(true)
)
