package nlp

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)

func TestParseTomorrow(t *testing.T) {
	p := Parse("buy groceries tomorrow", testNow)
	if p.Title != "Buy groceries" {
		t.Errorf("title = %q, want 'Buy groceries'", p.Title)
	}
	if p.DueDate != "2026-04-21" {
		t.Errorf("due = %q, want '2026-04-21'", p.DueDate)
	}
}

func TestParseNextMonday(t *testing.T) {
	// 2026-04-20 is a Monday, so "next monday" = 2026-04-27
	p := Parse("meeting next monday", testNow)
	if p.DueDate != "2026-04-27" {
		t.Errorf("due = %q, want '2026-04-27'", p.DueDate)
	}
	if p.Title != "Meeting" {
		t.Errorf("title = %q, want 'Meeting'", p.Title)
	}
}

func TestParseInDays(t *testing.T) {
	p := Parse("finish report in 3 days", testNow)
	if p.DueDate != "2026-04-23" {
		t.Errorf("due = %q, want '2026-04-23'", p.DueDate)
	}
	if p.Title != "Finish report" {
		t.Errorf("title = %q, want 'Finish report'", p.Title)
	}
}

func TestParsePriority(t *testing.T) {
	p := Parse("urgent fix the bug", testNow)
	if p.Priority != "urgent" {
		t.Errorf("priority = %q, want 'urgent'", p.Priority)
	}
	if p.Title != "Fix the bug" {
		t.Errorf("title = %q, want 'Fix the bug'", p.Title)
	}
}

func TestParseHashtags(t *testing.T) {
	p := Parse("clean desk #home #chores", testNow)
	if len(p.Tags) != 2 || p.Tags[0] != "home" || p.Tags[1] != "chores" {
		t.Errorf("tags = %v, want [home, chores]", p.Tags)
	}
	if p.Title != "Clean desk" {
		t.Errorf("title = %q, want 'Clean desk'", p.Title)
	}
}

func TestParseDuration(t *testing.T) {
	p := Parse("review PR 30 minutes", testNow)
	if p.Duration != 30 {
		t.Errorf("duration = %d, want 30", p.Duration)
	}
	if p.Title != "Review PR" {
		t.Errorf("title = %q, want 'Review PR'", p.Title)
	}
}

func TestParseDurationHours(t *testing.T) {
	p := Parse("write report 2 hours", testNow)
	if p.Duration != 120 {
		t.Errorf("duration = %d, want 120", p.Duration)
	}
}

func TestParseEnergy(t *testing.T) {
	p := Parse("quick reply to email", testNow)
	if p.Energy != "low" {
		t.Errorf("energy = %q, want 'low'", p.Energy)
	}
	if p.Title != "Reply to email" {
		t.Errorf("title = %q, want 'Reply to email'", p.Title)
	}
}

func TestParseComplex(t *testing.T) {
	p := Parse("important: write report by next friday 2h #work", testNow)
	if p.Priority != "high" {
		t.Errorf("priority = %q, want 'high'", p.Priority)
	}
	if p.Duration != 120 {
		t.Errorf("duration = %d, want 120", p.Duration)
	}
	if len(p.Tags) != 1 || p.Tags[0] != "work" {
		t.Errorf("tags = %v, want [work]", p.Tags)
	}
	if p.DueDate != "2026-04-24" {
		t.Errorf("due = %q, want '2026-04-24' (next friday)", p.DueDate)
	}
	if p.Title == "" {
		t.Error("title should not be empty")
	}
}

func TestParseNoExtras(t *testing.T) {
	p := Parse("just a plain task", testNow)
	if p.Title != "Just a plain task" {
		t.Errorf("title = %q, want 'Just a plain task'", p.Title)
	}
	if p.Priority != "" || p.Energy != "" || p.DueDate != "" || p.Duration != 0 || len(p.Tags) != 0 {
		t.Errorf("expected no extras, got: priority=%q energy=%q due=%q dur=%d tags=%v",
			p.Priority, p.Energy, p.DueDate, p.Duration, p.Tags)
	}
}

func TestParseContext(t *testing.T) {
	p := Parse("do laundry at home", testNow)
	if p.Context != "home" {
		t.Errorf("context = %q, want 'home'", p.Context)
	}
}

func TestParseToday(t *testing.T) {
	p := Parse("call dentist today", testNow)
	if p.DueDate != "2026-04-20" {
		t.Errorf("due = %q, want '2026-04-20'", p.DueDate)
	}
}

func TestParseInWeeks(t *testing.T) {
	p := Parse("plan vacation in 2 weeks", testNow)
	if p.DueDate != "2026-05-04" {
		t.Errorf("due = %q, want '2026-05-04'", p.DueDate)
	}
}
