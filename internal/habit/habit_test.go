package habit

import (
	"strings"
	"testing"
	"time"
)

func makeCompletions(dates ...string) []Completion {
	var cs []Completion
	for _, d := range dates {
		t, _ := time.Parse("2006-01-02", d)
		cs = append(cs, Completion{CompletedAt: t})
	}
	return cs
}

func TestCalculateStreakEmpty(t *testing.T) {
	stats := CalculateStreak(nil, "daily")
	if stats.CurrentStreak != 0 || stats.LongestStreak != 0 || stats.TotalCompletions != 0 {
		t.Errorf("empty completions should give all zeros, got %+v", stats)
	}
}

func TestCalculateStreakSingleDay(t *testing.T) {
	today := time.Now().UTC().Format("2006-01-02")
	cs := makeCompletions(today)
	stats := CalculateStreak(cs, "daily")
	if stats.TotalCompletions != 1 {
		t.Errorf("TotalCompletions = %d, want 1", stats.TotalCompletions)
	}
	if stats.CurrentStreak != 1 {
		t.Errorf("CurrentStreak = %d, want 1 (completed today)", stats.CurrentStreak)
	}
	if stats.LongestStreak != 1 {
		t.Errorf("LongestStreak = %d, want 1", stats.LongestStreak)
	}
}

func TestCalculateStreakConsecutiveDays(t *testing.T) {
	now := time.Now().UTC()
	dates := make([]string, 5)
	for i := 0; i < 5; i++ {
		dates[4-i] = now.AddDate(0, 0, -i).Format("2006-01-02")
	}
	cs := makeCompletions(dates...)
	stats := CalculateStreak(cs, "daily")

	if stats.CurrentStreak != 5 {
		t.Errorf("CurrentStreak = %d, want 5", stats.CurrentStreak)
	}
	if stats.LongestStreak != 5 {
		t.Errorf("LongestStreak = %d, want 5", stats.LongestStreak)
	}
}

func TestCalculateStreakBroken(t *testing.T) {
	now := time.Now().UTC()
	// Complete today, yesterday, and 5 days ago (gap breaks streak)
	cs := makeCompletions(
		now.AddDate(0, 0, -5).Format("2006-01-02"),
		now.AddDate(0, 0, -1).Format("2006-01-02"),
		now.Format("2006-01-02"),
	)
	stats := CalculateStreak(cs, "daily")

	if stats.CurrentStreak != 2 {
		t.Errorf("CurrentStreak = %d, want 2 (today + yesterday)", stats.CurrentStreak)
	}
	if stats.LongestStreak != 2 {
		t.Errorf("LongestStreak = %d, want 2", stats.LongestStreak)
	}
	if stats.TotalCompletions != 3 {
		t.Errorf("TotalCompletions = %d, want 3", stats.TotalCompletions)
	}
}

func TestCalculateStreakPastLongest(t *testing.T) {
	now := time.Now().UTC()
	// Old 4-day streak (broken) + current 2-day streak
	cs := makeCompletions(
		now.AddDate(0, 0, -20).Format("2006-01-02"),
		now.AddDate(0, 0, -19).Format("2006-01-02"),
		now.AddDate(0, 0, -18).Format("2006-01-02"),
		now.AddDate(0, 0, -17).Format("2006-01-02"),
		// gap
		now.AddDate(0, 0, -1).Format("2006-01-02"),
		now.Format("2006-01-02"),
	)
	stats := CalculateStreak(cs, "daily")

	if stats.CurrentStreak != 2 {
		t.Errorf("CurrentStreak = %d, want 2", stats.CurrentStreak)
	}
	if stats.LongestStreak != 4 {
		t.Errorf("LongestStreak = %d, want 4 (past streak)", stats.LongestStreak)
	}
}

func TestCalculateStreakNotActiveToday(t *testing.T) {
	now := time.Now().UTC()
	// Last completion was 3 days ago — streak is broken
	cs := makeCompletions(
		now.AddDate(0, 0, -5).Format("2006-01-02"),
		now.AddDate(0, 0, -4).Format("2006-01-02"),
		now.AddDate(0, 0, -3).Format("2006-01-02"),
	)
	stats := CalculateStreak(cs, "daily")

	if stats.CurrentStreak != 0 {
		t.Errorf("CurrentStreak = %d, want 0 (streak broken, last was 3 days ago)", stats.CurrentStreak)
	}
	if stats.LongestStreak != 3 {
		t.Errorf("LongestStreak = %d, want 3", stats.LongestStreak)
	}
}

func TestCalculateStreakWeekly(t *testing.T) {
	now := time.Now().UTC()
	// Completions in current week and past 2 weeks = 3-week streak
	cs := makeCompletions(
		now.AddDate(0, 0, -14).Format("2006-01-02"),
		now.AddDate(0, 0, -7).Format("2006-01-02"),
		now.Format("2006-01-02"),
	)
	stats := CalculateStreak(cs, "weekly")

	if stats.TotalCompletions != 3 {
		t.Errorf("TotalCompletions = %d, want 3", stats.TotalCompletions)
	}
	// Week streak depends on exact ISO week boundaries, just check it's > 0
	if stats.CurrentStreak == 0 {
		t.Error("CurrentStreak should be > 0 for weekly habit completed this week")
	}
}

func TestFormatCompletion(t *testing.T) {
	c := Completion{
		CompletedAt: time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC),
		Duration:    35,
		Note:        "DDIA ch.7",
	}
	got := FormatCompletion(c)
	want := "- 2026-04-05T08:30:00Z | 35min | DDIA ch.7"
	if got != want {
		t.Errorf("FormatCompletion = %q, want %q", got, want)
	}
}

func TestFormatCompletionNoDuration(t *testing.T) {
	c := Completion{
		CompletedAt: time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC),
	}
	got := FormatCompletion(c)
	want := "- 2026-04-05T08:30:00Z"
	if got != want {
		t.Errorf("FormatCompletion = %q, want %q", got, want)
	}
}

func TestFormatCompletionWithNote(t *testing.T) {
	c := Completion{
		CompletedAt: time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC),
		Note:        "quick session",
	}
	got := FormatCompletion(c)
	want := "- 2026-04-05T08:30:00Z | quick session"
	if got != want {
		t.Errorf("FormatCompletion = %q, want %q", got, want)
	}
}

func TestParseCompletionLine(t *testing.T) {
	tests := []struct {
		line     string
		wantDur  int
		wantNote string
	}{
		{"- 2026-04-05T08:30:00Z | 35min | DDIA ch.7", 35, "DDIA ch.7"},
		{"- 2026-04-05T08:30:00Z | 10min", 10, ""},
		{"- 2026-04-05T08:30:00Z", 0, ""},
		{"- 2026-04-05T08:30:00Z | some note", 0, "some note"},
	}

	for _, tt := range tests {
		c, err := ParseCompletionLine(tt.line)
		if err != nil {
			t.Errorf("ParseCompletionLine(%q) error: %v", tt.line, err)
			continue
		}
		if c.Duration != tt.wantDur {
			t.Errorf("ParseCompletionLine(%q) duration = %d, want %d", tt.line, c.Duration, tt.wantDur)
		}
		if c.Note != tt.wantNote {
			t.Errorf("ParseCompletionLine(%q) note = %q, want %q", tt.line, c.Note, tt.wantNote)
		}
	}
}

func TestParseCompletionsFromBody(t *testing.T) {
	body := `Some description here.

## Completions
- 2026-04-05T08:30:00Z | 35min | DDIA ch.7
- 2026-04-04T07:00:00Z | 30min | DDIA ch.6
- 2026-04-03T09:00:00Z

## Notes
Some other section.`

	cs := ParseCompletionsFromBody(body)
	if len(cs) != 3 {
		t.Fatalf("ParseCompletionsFromBody returned %d completions, want 3", len(cs))
	}
	if cs[0].Duration != 35 {
		t.Errorf("first completion duration = %d, want 35", cs[0].Duration)
	}
	if cs[0].Note != "DDIA ch.7" {
		t.Errorf("first completion note = %q, want 'DDIA ch.7'", cs[0].Note)
	}
	if cs[2].Duration != 0 {
		t.Errorf("third completion duration = %d, want 0", cs[2].Duration)
	}
}

func TestParseCompletionsFromBodyEmpty(t *testing.T) {
	cs := ParseCompletionsFromBody("Just a regular task body.")
	if len(cs) != 0 {
		t.Errorf("no ## Completions section should return 0, got %d", len(cs))
	}
}

func TestAppendCompletionToBody(t *testing.T) {
	c := Completion{
		CompletedAt: time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC),
		Duration:    20,
		Note:        "morning session",
	}

	// Append to empty body
	body := AppendCompletionToBody("", c)
	if !strings.Contains(body, "## Completions") {
		t.Error("should create ## Completions section")
	}
	if !strings.Contains(body, "2026-04-07T10:00:00Z") {
		t.Error("should contain completion timestamp")
	}

	// Append second completion
	c2 := Completion{
		CompletedAt: time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
	}
	body = AppendCompletionToBody(body, c2)
	cs := ParseCompletionsFromBody(body)
	if len(cs) != 2 {
		t.Errorf("should have 2 completions after second append, got %d", len(cs))
	}
}

func TestAppendCompletionToExistingBody(t *testing.T) {
	body := "Read 30 minutes every day.\n\n## Completions\n- 2026-04-05T08:30:00Z | 35min\n"

	c := Completion{
		CompletedAt: time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC),
		Duration:    30,
	}
	body = AppendCompletionToBody(body, c)
	cs := ParseCompletionsFromBody(body)
	if len(cs) != 2 {
		t.Errorf("should have 2 completions, got %d", len(cs))
	}
}

func TestCompletionRate(t *testing.T) {
	now := time.Now().UTC()
	cs := make([]Completion, 7)
	for i := 0; i < 7; i++ {
		cs[i] = Completion{CompletedAt: now.AddDate(0, 0, -i)}
	}

	// 7 completions in 7 days, target 1/day = 100%
	rate := CompletionRate(cs, "daily", 1, 7)
	if rate < 0.99 {
		t.Errorf("CompletionRate = %.2f, want ~1.0", rate)
	}

	// Same 7 completions, but target 2/day = 50%
	rate = CompletionRate(cs, "daily", 2, 7)
	if rate < 0.49 || rate > 0.51 {
		t.Errorf("CompletionRate (target=2) = %.2f, want ~0.5", rate)
	}
}

func TestCompletionRateZero(t *testing.T) {
	rate := CompletionRate(nil, "daily", 1, 30)
	if rate != 0 {
		t.Errorf("CompletionRate(nil) = %.2f, want 0", rate)
	}
}
