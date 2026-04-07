package recurrence

import (
	"testing"
	"time"
)

func TestParseDaily(t *testing.T) {
	r, err := Parse("FREQ=DAILY")
	if err != nil {
		t.Fatalf("Parse(FREQ=DAILY) error: %v", err)
	}
	if r.String() != "FREQ=DAILY" {
		t.Errorf("String() = %q, want %q", r.String(), "FREQ=DAILY")
	}
	if r.Frequency() != "daily" {
		t.Errorf("Frequency() = %q, want %q", r.Frequency(), "daily")
	}
}

func TestParseWeeklyWithDays(t *testing.T) {
	r, err := Parse("FREQ=WEEKLY;BYDAY=MO,WE,FR")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	freq := r.Frequency()
	if freq != "weekly on Mon, Wed, Fri" {
		t.Errorf("Frequency() = %q, want 'weekly on Mon, Wed, Fri'", freq)
	}
}

func TestParseMonthlyByDay(t *testing.T) {
	r, err := Parse("FREQ=MONTHLY;BYMONTHDAY=1,15")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	freq := r.Frequency()
	if freq != "monthly on the 1, 15" {
		t.Errorf("Frequency() = %q, want 'monthly on the 1, 15'", freq)
	}
}

func TestParseDailyInterval(t *testing.T) {
	r, err := Parse("FREQ=DAILY;INTERVAL=3")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if r.Frequency() != "every 3 days" {
		t.Errorf("Frequency() = %q, want 'every 3 days'", r.Frequency())
	}
}

func TestParseWeeklyInterval(t *testing.T) {
	r, err := Parse("FREQ=WEEKLY;INTERVAL=2")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if r.Frequency() != "every 2 weeks" {
		t.Errorf("Frequency() = %q, want 'every 2 weeks'", r.Frequency())
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("INVALID")
	if err == nil {
		t.Error("Parse(INVALID) should return error")
	}
}

func TestParseStripsRRULEPrefix(t *testing.T) {
	r, err := Parse("RRULE:FREQ=DAILY")
	if err != nil {
		t.Fatalf("Parse with RRULE: prefix error: %v", err)
	}
	if r.Frequency() != "daily" {
		t.Errorf("Frequency() = %q, want 'daily'", r.Frequency())
	}
}

func TestValidate(t *testing.T) {
	if err := Validate("FREQ=DAILY"); err != nil {
		t.Errorf("Validate(FREQ=DAILY) should pass: %v", err)
	}
	if err := Validate("GARBAGE"); err == nil {
		t.Error("Validate(GARBAGE) should fail")
	}
}

func TestNextAfterDaily(t *testing.T) {
	r, _ := Parse("FREQ=DAILY")
	// Pin DTSTART to a fixed date so the test doesn't depend on time.Now()
	r.SetDTStart(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))

	after := time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)
	next, ok := r.NextAfter(after)
	if !ok {
		t.Fatal("NextAfter should return a result")
	}
	expected := time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC)
	if next.Before(after) {
		t.Errorf("NextAfter returned %v which is before %v", next, after)
	}
	if next.Year() != expected.Year() || next.Month() != expected.Month() || next.Day() != expected.Day() {
		t.Errorf("NextAfter = %v, want date %v", next, expected)
	}
}

func TestNextAfterWeeklyMWF(t *testing.T) {
	r, _ := Parse("FREQ=WEEKLY;BYDAY=MO,WE,FR")
	// Pin DTSTART to a Monday before the test range
	r.SetDTStart(time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC))

	// 2026-04-07 is a Tuesday
	after := time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)
	next, ok := r.NextAfter(after)
	if !ok {
		t.Fatal("NextAfter should return a result")
	}
	// Next should be Wednesday Apr 8
	if next.Weekday() != time.Wednesday {
		t.Errorf("NextAfter on Tuesday should be Wednesday, got %v (%v)", next.Weekday(), next)
	}
}

func TestNextAfterCompletion(t *testing.T) {
	r, _ := Parse("FREQ=WEEKLY;INTERVAL=2")

	// Completed on Apr 7
	completed := time.Date(2026, 4, 7, 15, 0, 0, 0, time.UTC)
	next, ok := r.NextAfterCompletion(completed)
	if !ok {
		t.Fatal("NextAfterCompletion should return a result")
	}
	// Should be 2 weeks later
	expected := time.Date(2026, 4, 21, 15, 0, 0, 0, time.UTC)
	if next.Year() != expected.Year() || next.Month() != expected.Month() || next.Day() != expected.Day() {
		t.Errorf("NextAfterCompletion = %v, want date %v", next, expected)
	}
}

func TestBetween(t *testing.T) {
	r, _ := Parse("FREQ=DAILY")
	r.SetDTStart(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))

	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 7, 23, 59, 59, 0, time.UTC)
	dates := r.Between(start, end)

	if len(dates) != 7 {
		t.Errorf("Between(Apr 1-7) returned %d dates, want 7", len(dates))
	}
}

func TestBetweenWeekly(t *testing.T) {
	r, _ := Parse("FREQ=WEEKLY;BYDAY=MO,WE,FR")
	r.SetDTStart(time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC))

	start := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)   // Monday
	end := time.Date(2026, 4, 12, 23, 59, 59, 0, time.UTC) // Sunday
	dates := r.Between(start, end)

	// Mon, Wed, Fri = 3 dates
	if len(dates) != 3 {
		t.Errorf("Between(Mon-Sun, MWF) returned %d dates, want 3", len(dates))
	}
}
