// Package model defines the core data types for BentoTask.
//
// These types map directly to the YAML frontmatter schema defined in ADR-002.
// Every task, habit, and routine on disk is represented by these structures.
//
// ULID generation lives here because IDs are a fundamental part of the model,
// not a storage concern.
package model

import (
	"crypto/rand"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// NewID generates a new ULID string.
// ULIDs are 26-character, time-sortable, URL-safe identifiers (ADR-002 §4).
// They encode millisecond-precision timestamps + 80 bits of randomness.
func NewID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

// NewIDAt generates a ULID for a specific timestamp.
// Useful for testing or importing historical data.
func NewIDAt(t time.Time) string {
	return ulid.MustNew(ulid.Timestamp(t), rand.Reader).String()
}

// IDTime extracts the creation timestamp from a ULID string.
// Returns zero time if the ID is invalid.
func IDTime(id string) time.Time {
	parsed, err := ulid.Parse(id)
	if err != nil {
		return time.Time{}
	}
	return ulid.Time(parsed.Time())
}

// MatchesPrefix returns true if the id starts with the given prefix.
// Used for CLI short-ID matching: `bt done 01JQX` matches `01JQX00010ABCDEF12345678`.
// Comparison is case-insensitive since ULIDs use Crockford Base32.
func MatchesPrefix(id, prefix string) bool {
	return len(id) >= len(prefix) &&
		strings.EqualFold(id[:len(prefix)], prefix)
}
