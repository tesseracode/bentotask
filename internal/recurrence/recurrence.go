// Package recurrence provides RRULE parsing and next-occurrence calculation
// for recurring tasks and habits.
//
// It wraps the teambition/rrule-go library with BentoTask-specific logic,
// including support for the "after completion" anchor mode (ADR-002 §7).
package recurrence

import (
	"fmt"
	"strings"
	"time"

	"github.com/teambition/rrule-go"
)

// Rule wraps a parsed RRULE with BentoTask-specific behavior.
type Rule struct {
	rrule *rrule.RRule
	raw   string
}

// Parse parses an RFC 5545 RRULE string into a Rule.
// The RRULE should not include the "RRULE:" prefix.
//
// Examples:
//
//	"FREQ=DAILY"
//	"FREQ=WEEKLY;BYDAY=MO,WE,FR"
//	"FREQ=MONTHLY;BYMONTHDAY=1,15"
//	"FREQ=DAILY;INTERVAL=3"
func Parse(s string) (*Rule, error) {
	// Strip RRULE: prefix if present (be lenient)
	s = strings.TrimPrefix(s, "RRULE:")

	r, err := rrule.StrToRRule(s)
	if err != nil {
		return nil, fmt.Errorf("parse rrule %q: %w", s, err)
	}
	return &Rule{rrule: r, raw: s}, nil
}

// String returns the original RRULE string.
func (r *Rule) String() string {
	return r.raw
}

// SetDTStart overrides the rule's DTSTART to the given time.
// This is useful for pinning recurrence to a fixed start date
// rather than relying on the default (time.Now()).
func (r *Rule) SetDTStart(dt time.Time) {
	opts := r.rrule.OrigOptions
	opts.Dtstart = dt
	adjusted, err := rrule.NewRRule(opts)
	if err != nil {
		return // keep original on error
	}
	r.rrule = adjusted
}

// NextAfter returns the next occurrence strictly after the given time.
// Used for fixed-anchor recurrence: the schedule is calendar-based
// regardless of when the task was last completed.
func (r *Rule) NextAfter(after time.Time) (time.Time, bool) {
	// rrule-go's After method returns the next occurrence after the given time.
	// The second parameter (inc) controls whether `after` itself is included.
	next := r.rrule.After(after, false)
	if next.IsZero() {
		return time.Time{}, false
	}
	return next, true
}

// NextAfterCompletion returns the next occurrence calculated from
// a completion date. Used for completion-anchor recurrence.
//
// For example, "FREQ=WEEKLY;INTERVAL=2" with completion anchor means
// "2 weeks after last completion", not "every other Monday".
func (r *Rule) NextAfterCompletion(completedAt time.Time) (time.Time, bool) {
	// Clone the rule with DTStart set to the completion time
	// so the interval is computed relative to when it was done.
	opts := r.rrule.OrigOptions
	opts.Dtstart = completedAt
	adjusted, err := rrule.NewRRule(opts)
	if err != nil {
		return time.Time{}, false
	}
	next := adjusted.After(completedAt, false)
	if next.IsZero() {
		return time.Time{}, false
	}
	return next, true
}

// Between returns all occurrences between start and end (inclusive).
// Useful for checking which dates in a range have scheduled occurrences.
// If the rule has no DTSTART, it uses start as the origin.
func (r *Rule) Between(start, end time.Time) []time.Time {
	// If the rule doesn't have a meaningful DTSTART, create a temporary
	// rule anchored at start so occurrences are generated in range.
	if r.rrule.OrigOptions.Dtstart.IsZero() || r.rrule.OrigOptions.Dtstart.After(end) {
		opts := r.rrule.OrigOptions
		opts.Dtstart = start
		adjusted, err := rrule.NewRRule(opts)
		if err != nil {
			return nil
		}
		return adjusted.Between(start, end, true)
	}
	return r.rrule.Between(start, end, true)
}

// Frequency returns the human-readable frequency description.
func (r *Rule) Frequency() string {
	switch r.rrule.OrigOptions.Freq {
	case rrule.DAILY:
		if r.rrule.OrigOptions.Interval > 1 {
			return fmt.Sprintf("every %d days", r.rrule.OrigOptions.Interval)
		}
		return "daily"
	case rrule.WEEKLY:
		if r.rrule.OrigOptions.Interval > 1 {
			return fmt.Sprintf("every %d weeks", r.rrule.OrigOptions.Interval)
		}
		if len(r.rrule.OrigOptions.Byweekday) > 0 {
			days := make([]string, len(r.rrule.OrigOptions.Byweekday))
			for i, wd := range r.rrule.OrigOptions.Byweekday {
				days[i] = weekdayName(wd)
			}
			return "weekly on " + strings.Join(days, ", ")
		}
		return "weekly"
	case rrule.MONTHLY:
		if len(r.rrule.OrigOptions.Bymonthday) > 0 {
			days := make([]string, len(r.rrule.OrigOptions.Bymonthday))
			for i, d := range r.rrule.OrigOptions.Bymonthday {
				days[i] = fmt.Sprintf("%d", d)
			}
			return "monthly on the " + strings.Join(days, ", ")
		}
		return "monthly"
	case rrule.YEARLY:
		return "yearly"
	default:
		return r.raw
	}
}

func weekdayName(wd rrule.Weekday) string {
	names := map[rrule.Weekday]string{
		rrule.MO: "Mon",
		rrule.TU: "Tue",
		rrule.WE: "Wed",
		rrule.TH: "Thu",
		rrule.FR: "Fri",
		rrule.SA: "Sat",
		rrule.SU: "Sun",
	}
	if name, ok := names[wd]; ok {
		return name
	}
	return wd.String()
}

// Validate checks if an RRULE string is syntactically valid.
// Returns nil if valid, an error otherwise.
func Validate(s string) error {
	_, err := Parse(s)
	return err
}
