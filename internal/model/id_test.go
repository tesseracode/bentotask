package model

import (
	"testing"
	"time"
)

func TestNewID(t *testing.T) {
	id := NewID()

	// ULIDs are exactly 26 characters
	if len(id) != 26 {
		t.Errorf("NewID() length = %d, want 26", len(id))
	}

	// Two IDs should be different
	id2 := NewID()
	if id == id2 {
		t.Error("two NewID() calls returned the same value")
	}
}

func TestNewIDAt(t *testing.T) {
	ts := time.Date(2026, 4, 5, 10, 30, 0, 0, time.UTC)
	id := NewIDAt(ts)

	if len(id) != 26 {
		t.Errorf("NewIDAt() length = %d, want 26", len(id))
	}

	// The extracted time should match (within millisecond precision)
	extracted := IDTime(id)
	diff := extracted.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Millisecond {
		t.Errorf("IDTime() = %v, want within 1ms of %v (diff: %v)", extracted, ts, diff)
	}
}

func TestIDTime(t *testing.T) {
	id := NewID()
	ts := IDTime(id)

	// Should be recent (within last second)
	if time.Since(ts) > time.Second {
		t.Errorf("IDTime() = %v, expected within last second", ts)
	}
}

func TestIDTimeInvalid(t *testing.T) {
	ts := IDTime("not-a-valid-ulid")
	if !ts.IsZero() {
		t.Errorf("IDTime(invalid) = %v, want zero time", ts)
	}
}

func TestMatchesPrefix(t *testing.T) {
	id := "01JQX00010ABCDEF12345678"

	tests := []struct {
		prefix string
		want   bool
	}{
		{"01JQX", true},
		{"01JQX00010ABCDEF12345678", true}, // full match
		{"01jqx", true},                    // case-insensitive
		{"01ZZZ", false},
		{"", true},                               // empty prefix matches everything
		{"01JQX00010ABCDEF12345678EXTRA", false}, // prefix longer than id
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			got := MatchesPrefix(id, tt.prefix)
			if got != tt.want {
				t.Errorf("MatchesPrefix(%q, %q) = %v, want %v", id, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestNewIDsAreSortable(t *testing.T) {
	// IDs created later should sort after earlier ones
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	id1 := NewIDAt(t1)
	id2 := NewIDAt(t2)

	if id1 >= id2 {
		t.Errorf("earlier ULID %q should sort before later ULID %q", id1, id2)
	}
}
