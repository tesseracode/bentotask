// Package nlp provides lightweight, rule-based natural language parsing
// for extracting structured task data from freeform text.
// No external API calls — just regex and string matching.
package nlp

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParsedTask holds the structured data extracted from natural language.
type ParsedTask struct {
	Title    string
	Priority string
	Energy   string
	DueDate  string // YYYY-MM-DD
	Tags     []string
	Context  string
	Duration int // minutes
}

// Parse extracts structured task data from natural language text.
func Parse(text string, now time.Time) ParsedTask {
	result := ParsedTask{}
	remaining := text

	// Order matters: extract specific patterns first, then keywords

	// 1. Hashtags → tags
	remaining, result.Tags = extractTags(remaining)

	// 2. Duration → minutes
	remaining, result.Duration = extractDuration(remaining)

	// 3. Due dates → YYYY-MM-DD
	remaining, result.DueDate = extractDueDate(remaining, now)

	// 4. Priority keywords
	remaining, result.Priority = extractPriority(remaining)

	// 5. Energy keywords
	remaining, result.Energy = extractEnergy(remaining)

	// 6. Context markers
	remaining, result.Context = extractContext(remaining)

	// 7. Title is whatever remains, cleaned up
	result.Title = cleanTitle(remaining)

	return result
}

// --- Extractors ---

var tagRegex = regexp.MustCompile(`#(\w+)`)

func extractTags(text string) (string, []string) {
	matches := tagRegex.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return text, nil
	}
	var tags []string
	for _, m := range matches {
		tags = append(tags, m[1])
	}
	cleaned := tagRegex.ReplaceAllString(text, "")
	return cleaned, tags
}

var durationPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(\d+)\s*hours?`),
	regexp.MustCompile(`(?i)(\d+)\s*hrs?`),
	regexp.MustCompile(`(?i)(\d+)\s*h\b`),
	regexp.MustCompile(`(?i)(\d+)\s*minutes?`),
	regexp.MustCompile(`(?i)(\d+)\s*mins?`),
	regexp.MustCompile(`(?i)(\d+)\s*m\b`),
}

func extractDuration(text string) (string, int) {
	// Try hours first
	for _, re := range durationPatterns[:3] {
		if m := re.FindStringSubmatch(text); m != nil {
			n, _ := strconv.Atoi(m[1])
			return re.ReplaceAllString(text, ""), n * 60
		}
	}
	// Then minutes
	for _, re := range durationPatterns[3:] {
		if m := re.FindStringSubmatch(text); m != nil {
			n, _ := strconv.Atoi(m[1])
			return re.ReplaceAllString(text, ""), n
		}
	}
	return text, 0
}

var dayNames = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
	"wednesday": time.Wednesday, "thursday": time.Thursday, "friday": time.Friday,
	"saturday": time.Saturday,
}

var monthNames = map[string]time.Month{
	"january": time.January, "february": time.February, "march": time.March,
	"april": time.April, "may": time.May, "june": time.June,
	"july": time.July, "august": time.August, "september": time.September,
	"october": time.October, "november": time.November, "december": time.December,
	"jan": time.January, "feb": time.February, "mar": time.March,
	"apr": time.April, "jun": time.June, "jul": time.July,
	"aug": time.August, "sep": time.September, "oct": time.October,
	"nov": time.November, "dec": time.December,
}

func extractDueDate(text string, now time.Time) (string, string) {
	lower := strings.ToLower(text)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// "today"
	if re := regexp.MustCompile(`(?i)\b(today)\b`); re.MatchString(lower) {
		return re.ReplaceAllString(text, ""), today.Format("2006-01-02")
	}

	// "tomorrow"
	if re := regexp.MustCompile(`(?i)\b(tomorrow)\b`); re.MatchString(lower) {
		return re.ReplaceAllString(text, ""), today.AddDate(0, 0, 1).Format("2006-01-02")
	}

	// "next monday" etc.
	nextDayRe := regexp.MustCompile(`(?i)\bnext\s+(monday|tuesday|wednesday|thursday|friday|saturday|sunday)\b`)
	if m := nextDayRe.FindStringSubmatch(lower); m != nil {
		targetDay := dayNames[strings.ToLower(m[1])]
		d := today.AddDate(0, 0, 1)
		for d.Weekday() != targetDay {
			d = d.AddDate(0, 0, 1)
		}
		return nextDayRe.ReplaceAllString(text, ""), d.Format("2006-01-02")
	}

	// "in N days/weeks"
	inDaysRe := regexp.MustCompile(`(?i)\bin\s+(\d+)\s+(days?|weeks?)\b`)
	if m := inDaysRe.FindStringSubmatch(lower); m != nil {
		n, _ := strconv.Atoi(m[1])
		unit := strings.ToLower(m[2])
		if strings.HasPrefix(unit, "week") {
			n *= 7
		}
		return inDaysRe.ReplaceAllString(text, ""), today.AddDate(0, 0, n).Format("2006-01-02")
	}

	// "by april 20" / "by apr 20" / "april 20"
	monthDayRe := regexp.MustCompile(`(?i)\b(?:by\s+)?(\w+)\s+(\d{1,2})\b`)
	if m := monthDayRe.FindStringSubmatch(lower); m != nil {
		if month, ok := monthNames[strings.ToLower(m[1])]; ok {
			day, _ := strconv.Atoi(m[2])
			year := now.Year()
			d := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
			if d.Before(today) {
				d = d.AddDate(1, 0, 0)
			}
			return monthDayRe.ReplaceAllString(text, ""), d.Format("2006-01-02")
		}
	}

	// ISO format "2026-04-20"
	isoRe := regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2})\b`)
	if m := isoRe.FindStringSubmatch(text); m != nil {
		return isoRe.ReplaceAllString(text, ""), m[1]
	}

	// "by" prefix alone with existing patterns handled above
	byRe := regexp.MustCompile(`(?i)\bby\s+`)
	text = byRe.ReplaceAllString(text, "")

	return text, ""
}

var priorityPatterns = []struct {
	re       *regexp.Regexp
	priority string
}{
	{regexp.MustCompile(`(?i)\b(urgent|urgently|asap|critical)\b:?\s*`), "urgent"},
	{regexp.MustCompile(`(?i)\b(important|high\s+priority)\b:?\s*`), "high"},
}

func extractPriority(text string) (string, string) {
	for _, p := range priorityPatterns {
		if p.re.MatchString(text) {
			return p.re.ReplaceAllString(text, ""), p.priority
		}
	}
	return text, ""
}

var energyPatterns = []struct {
	re     *regexp.Regexp
	energy string
}{
	{regexp.MustCompile(`(?i)\b(quick|easy|simple)\b\s*`), "low"},
	{regexp.MustCompile(`(?i)\b(deep\s+work|focused|complex)\b\s*`), "high"},
}

func extractEnergy(text string) (string, string) {
	for _, p := range energyPatterns {
		if p.re.MatchString(text) {
			return p.re.ReplaceAllString(text, ""), p.energy
		}
	}
	return text, ""
}

var contextPatterns = []struct {
	re      *regexp.Regexp
	context string
}{
	{regexp.MustCompile(`(?i)\bat\s+home\b`), "home"},
	{regexp.MustCompile(`(?i)\bat\s+the\s+office\b`), "office"},
	{regexp.MustCompile(`(?i)\bat\s+office\b`), "office"},
	{regexp.MustCompile(`(?i)\bwhile\s+commuting\b`), "commuting"},
}

func extractContext(text string) (string, string) {
	for _, p := range contextPatterns {
		if p.re.MatchString(text) {
			return p.re.ReplaceAllString(text, ""), p.context
		}
	}
	return text, ""
}

func cleanTitle(text string) string {
	// Remove leading/trailing punctuation, extra spaces
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimRight(text, ",;:- ")
	text = strings.TrimLeft(text, ",;:- ")
	// Capitalize first letter
	if len(text) > 0 {
		text = strings.ToUpper(text[:1]) + text[1:]
	}
	return text
}
