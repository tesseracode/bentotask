package style

import (
	"strings"
	"testing"

	"github.com/muesli/termenv"

	"github.com/charmbracelet/lipgloss"
)

// resetColorProfile restores the default color profile after DisableColor tests.
func resetColorProfile(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		lipgloss.SetColorProfile(termenv.TrueColor)
	})
}

func TestPriorityContainsLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"urgent", "urgent"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"none", "none"},
		{"", "none"},
	}

	for _, tt := range tests {
		got := Priority(tt.input)
		if !strings.Contains(got, tt.want) {
			t.Errorf("Priority(%q) = %q, should contain %q", tt.input, got, tt.want)
		}
	}
}

func TestStatusContainsName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"done", "done"},
		{"active", "active"},
		{"pending", "pending"},
		{"blocked", "blocked"},
		{"waiting", "waiting"},
		{"paused", "paused"},
		{"cancelled", "cancelled"},
	}

	for _, tt := range tests {
		got := Status(tt.input)
		if !strings.Contains(got, tt.want) {
			t.Errorf("Status(%q) = %q, should contain %q", tt.input, got, tt.want)
		}
	}
}

func TestStatusIcons(t *testing.T) {
	tests := []struct {
		input string
		icon  string
	}{
		{"done", "✓"},
		{"active", "●"},
		{"pending", "○"},
		{"blocked", "✗"},
		{"paused", "◌"},
		{"cancelled", "⊘"},
	}

	for _, tt := range tests {
		got := Status(tt.input)
		if !strings.Contains(got, tt.icon) {
			t.Errorf("Status(%q) = %q, should contain icon %q", tt.input, got, tt.icon)
		}
	}
}

func TestEnergyContainsLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
	}

	for _, tt := range tests {
		got := Energy(tt.input)
		if !strings.Contains(got, tt.want) {
			t.Errorf("Energy(%q) = %q, should contain %q", tt.input, got, tt.want)
		}
	}
}

func TestEnergyIcons(t *testing.T) {
	tests := []struct {
		input string
		icon  string
	}{
		{"high", "⚡"},
		{"medium", "~"},
		{"low", "·"},
	}

	for _, tt := range tests {
		got := Energy(tt.input)
		if !strings.Contains(got, tt.icon) {
			t.Errorf("Energy(%q) = %q, should contain icon %q", tt.input, got, tt.icon)
		}
	}
}

func TestTagPrefixesHash(t *testing.T) {
	got := Tag("work")
	if !strings.Contains(got, "#work") {
		t.Errorf("Tag('work') = %q, should contain '#work'", got)
	}
}

func TestSuccessContainsCheckmark(t *testing.T) {
	got := Success("done it")
	if !strings.Contains(got, "✓") {
		t.Errorf("Success() = %q, should contain '✓'", got)
	}
	if !strings.Contains(got, "done it") {
		t.Errorf("Success() = %q, should contain message", got)
	}
}

func TestErrorMsgContainsPrefix(t *testing.T) {
	got := ErrorMsg("something broke")
	if !strings.Contains(got, "Error:") {
		t.Errorf("ErrorMsg() = %q, should contain 'Error:'", got)
	}
	if !strings.Contains(got, "something broke") {
		t.Errorf("ErrorMsg() = %q, should contain message", got)
	}
}

func TestDimBoldHeaderReturnContent(t *testing.T) {
	if got := Dim("test"); !strings.Contains(got, "test") {
		t.Errorf("Dim('test') = %q, should contain 'test'", got)
	}
	if got := Bold("test"); !strings.Contains(got, "test") {
		t.Errorf("Bold('test') = %q, should contain 'test'", got)
	}
	if got := Header("test"); !strings.Contains(got, "test") {
		t.Errorf("Header('test') = %q, should contain 'test'", got)
	}
}

func TestDueDateOverdue(t *testing.T) {
	got := DueDate("2026-01-01", true)
	if !strings.Contains(got, "2026-01-01") {
		t.Errorf("DueDate overdue = %q, should contain date", got)
	}

	got = DueDate("2026-12-31", false)
	if got != "2026-12-31" {
		t.Errorf("DueDate not overdue = %q, want plain date", got)
	}
}

func TestDisableColorProducesPlainText(t *testing.T) {
	resetColorProfile(t)

	DisableColor()

	// After disabling, styled output should be plain text (no ANSI escapes)
	got := Priority("urgent")
	if strings.Contains(got, "\033[") {
		t.Errorf("after DisableColor(), Priority('urgent') = %q, should not contain ANSI escapes", got)
	}
	if got != "urgent" {
		t.Errorf("after DisableColor(), Priority('urgent') = %q, want plain 'urgent'", got)
	}

	got = Status("done")
	if strings.Contains(got, "\033[") {
		t.Errorf("after DisableColor(), Status('done') = %q, should not contain ANSI escapes", got)
	}

	got = Tag("work")
	if strings.Contains(got, "\033[") {
		t.Errorf("after DisableColor(), Tag('work') = %q, should not contain ANSI escapes", got)
	}
	if got != "#work" {
		t.Errorf("after DisableColor(), Tag('work') = %q, want '#work'", got)
	}
}

func TestUnknownPriorityPassthrough(t *testing.T) {
	got := Priority("custom")
	if !strings.Contains(got, "custom") {
		t.Errorf("Priority('custom') = %q, should pass through unknown value", got)
	}
}

func TestUnknownStatusPassthrough(t *testing.T) {
	got := Status("custom")
	if got != "custom" {
		t.Errorf("Status('custom') = %q, want 'custom' (passthrough)", got)
	}
}

func TestUnknownEnergyShowsDash(t *testing.T) {
	got := Energy("custom")
	if !strings.Contains(got, "-") {
		t.Errorf("Energy('custom') = %q, should contain '-'", got)
	}
}
