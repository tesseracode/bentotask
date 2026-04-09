// Package habit provides habit tracking, completion logging, and streak
// calculation for BentoTask.
//
// Habits are a special task type with a recurrence rule and a frequency target.
// Completions are stored both in the SQLite index (for fast queries) and in
// the markdown body (as the source of truth).
package habit

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Completion represents a single habit completion entry.
type Completion struct {
	CompletedAt time.Time
	Duration    int    // minutes (0 if not tracked)
	Note        string // optional note
}

// Stats holds computed statistics for a habit.
type Stats struct {
	CurrentStreak    int     `json:"current_streak"`
	LongestStreak    int     `json:"longest_streak"`
	TotalCompletions int     `json:"total_completions"`
	CompletionRate   float64 `json:"completion_rate"`  // 0.0–1.0 over the rate period
	RatePeriodDays   int     `json:"rate_period_days"` // number of days the rate covers
	CompletedToday   bool    `json:"completed_today"`  // whether the habit was completed today
}

// CalculateStreak computes streak information from a sorted list of completions
// and a frequency type ("daily" or "weekly").
//
// For daily habits: a streak continues if there's a completion every day.
// For weekly habits: a streak continues if there's at least one completion per week.
func CalculateStreak(completions []Completion, freqType string) Stats {
	if len(completions) == 0 {
		return Stats{}
	}

	// Sort completions by date ascending
	sorted := make([]Completion, len(completions))
	copy(sorted, completions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CompletedAt.Before(sorted[j].CompletedAt)
	})

	total := len(sorted)

	switch freqType {
	case "daily":
		current, longest := dailyStreaks(sorted)
		return Stats{
			CurrentStreak:    current,
			LongestStreak:    longest,
			TotalCompletions: total,
		}
	case "weekly":
		current, longest := weeklyStreaks(sorted)
		return Stats{
			CurrentStreak:    current,
			LongestStreak:    longest,
			TotalCompletions: total,
		}
	default:
		// Fallback: treat as daily
		current, longest := dailyStreaks(sorted)
		return Stats{
			CurrentStreak:    current,
			LongestStreak:    longest,
			TotalCompletions: total,
		}
	}
}

// dailyStreaks counts consecutive-day streaks.
// A day is considered "done" if there's at least one completion on that date.
func dailyStreaks(sorted []Completion) (current, longest int) {
	if len(sorted) == 0 {
		return 0, 0
	}

	// Deduplicate to unique dates
	dates := uniqueDates(sorted)

	// Calculate all streaks
	streak := 1
	longest = 1
	for i := 1; i < len(dates); i++ {
		diff := dates[i].Sub(dates[i-1])
		if diff <= 24*time.Hour+time.Minute { // allow small clock drift
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			streak = 1
		}
	}

	// Current streak: only counts if the last completion was today or yesterday
	now := time.Now().UTC()
	today := truncateToDay(now)
	lastDate := dates[len(dates)-1]

	daysSinceLast := today.Sub(lastDate)
	if daysSinceLast <= 24*time.Hour+time.Minute {
		current = streak // still active
	} else {
		current = 0 // streak broken
	}

	return current, longest
}

// weeklyStreaks counts consecutive-week streaks.
// A week is considered "done" if there's at least one completion in that ISO week.
func weeklyStreaks(sorted []Completion) (current, longest int) {
	if len(sorted) == 0 {
		return 0, 0
	}

	// Get unique ISO weeks
	weeks := uniqueWeeks(sorted)

	streak := 1
	longest = 1
	for i := 1; i < len(weeks); i++ {
		prevY, prevW := weeks[i-1][0], weeks[i-1][1]
		curY, curW := weeks[i][0], weeks[i][1]

		// Check if weeks are consecutive
		if isConsecutiveWeek(prevY, prevW, curY, curW) {
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			streak = 1
		}
	}

	// Current streak: check if last week is this week or last week
	now := time.Now().UTC()
	thisY, thisW := now.ISOWeek()
	lastY, lastW := weeks[len(weeks)-1][0], weeks[len(weeks)-1][1]

	if (lastY == thisY && lastW == thisW) || isConsecutiveWeek(lastY, lastW, thisY, thisW) {
		current = streak
	} else {
		current = 0
	}

	return current, longest
}

// CompletionRate calculates the completion rate over a period.
// For daily habits: completions / days
// For weekly habits: weeks with completions / weeks
func CompletionRate(completions []Completion, freqType string, target int, days int) float64 {
	if days == 0 || target == 0 {
		return 0
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	var count int
	for _, c := range completions {
		if c.CompletedAt.After(cutoff) {
			count++
		}
	}

	switch freqType {
	case "daily":
		expected := days * target
		if expected == 0 {
			return 0
		}
		rate := float64(count) / float64(expected)
		if rate > 1 {
			rate = 1
		}
		return rate
	case "weekly":
		weeks := days / 7
		if weeks == 0 {
			weeks = 1
		}
		expected := weeks * target
		if expected == 0 {
			return 0
		}
		rate := float64(count) / float64(expected)
		if rate > 1 {
			rate = 1
		}
		return rate
	default:
		return 0
	}
}

// FormatCompletion formats a completion for the markdown body log.
// Format: "- 2026-04-05T08:30:00Z | 35min | note text"
func FormatCompletion(c Completion) string {
	parts := []string{fmt.Sprintf("- %s", c.CompletedAt.UTC().Format(time.RFC3339))}
	if c.Duration > 0 {
		parts = append(parts, fmt.Sprintf("%dmin", c.Duration))
	}
	if c.Note != "" {
		parts = append(parts, c.Note)
	}
	return strings.Join(parts, " | ")
}

// ParseCompletionLine parses a single completion log line from the markdown body.
// Expected format: "- 2026-04-05T08:30:00Z | 35min | note text"
func ParseCompletionLine(line string) (Completion, error) {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "- ")

	parts := strings.SplitN(line, " | ", 3)
	if len(parts) == 0 || parts[0] == "" {
		return Completion{}, fmt.Errorf("empty completion line")
	}

	t, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[0]))
	if err != nil {
		return Completion{}, fmt.Errorf("parse completion time: %w", err)
	}

	c := Completion{CompletedAt: t}

	if len(parts) >= 2 {
		durStr := strings.TrimSpace(parts[1])
		if strings.HasSuffix(durStr, "min") {
			trimmed := strings.TrimSuffix(durStr, "min")
			var dur int
			if _, err := fmt.Sscanf(trimmed, "%d", &dur); err == nil {
				c.Duration = dur
			}
		} else if len(parts) == 2 {
			// No "min" suffix and only 2 parts — treat as note
			c.Note = strings.TrimSpace(durStr)
		}
	}

	if len(parts) >= 3 {
		c.Note = strings.TrimSpace(parts[2])
	}

	return c, nil
}

// ParseCompletionsFromBody extracts all completions from a markdown body.
// Looks for a "## Completions" section and parses each "- " line.
func ParseCompletionsFromBody(body string) []Completion {
	lines := strings.Split(body, "\n")
	inSection := false
	var completions []Completion

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Completions" {
			inSection = true
			continue
		}
		// Stop at the next heading
		if inSection && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if inSection && strings.HasPrefix(trimmed, "- ") {
			c, err := ParseCompletionLine(trimmed)
			if err == nil {
				completions = append(completions, c)
			}
		}
	}

	return completions
}

// AppendCompletionToBody adds a completion entry to the markdown body.
// Creates the "## Completions" section if it doesn't exist.
func AppendCompletionToBody(body string, c Completion) string {
	entry := FormatCompletion(c)

	if strings.Contains(body, "## Completions") {
		// Insert after the ## Completions header
		lines := strings.Split(body, "\n")
		var result []string
		inserted := false
		for _, line := range lines {
			result = append(result, line)
			if !inserted && strings.TrimSpace(line) == "## Completions" {
				result = append(result, entry)
				inserted = true
			}
		}
		return strings.Join(result, "\n")
	}

	// Create the section
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	if body != "" {
		body += "\n"
	}
	body += "## Completions\n" + entry + "\n"
	return body
}

// --- Helpers ---

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func uniqueDates(sorted []Completion) []time.Time {
	var dates []time.Time
	var lastDate time.Time
	for _, c := range sorted {
		d := truncateToDay(c.CompletedAt)
		if d != lastDate {
			dates = append(dates, d)
			lastDate = d
		}
	}
	return dates
}

func uniqueWeeks(sorted []Completion) [][2]int {
	seen := make(map[[2]int]bool)
	var weeks [][2]int
	for _, c := range sorted {
		y, w := c.CompletedAt.ISOWeek()
		key := [2]int{y, w}
		if !seen[key] {
			seen[key] = true
			weeks = append(weeks, key)
		}
	}
	return weeks
}

func isConsecutiveWeek(y1, w1, y2, w2 int) bool {
	if y1 == y2 {
		return w2 == w1+1
	}
	// Year boundary: last week of y1, first week of y2
	if y2 == y1+1 && w2 == 1 {
		// ISO week 52 or 53 → week 1
		return w1 >= 52
	}
	return false
}
