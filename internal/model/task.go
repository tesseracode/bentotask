// Package model defines the core data types for BentoTask.
//
// These types map directly to the YAML frontmatter schema defined in ADR-002.
// Every task, habit, and routine on disk is represented by these structures.
package model

import "time"

// TaskType defines what kind of task this is. Each type has different
// scheduling behavior and required fields.
type TaskType string

// Valid TaskType values.
const (
	TaskTypeOneShot   TaskType = "one-shot"  // Do once, no date constraint
	TaskTypeDated     TaskType = "dated"     // Due on a specific date
	TaskTypeRanged    TaskType = "ranged"    // Due within a date range
	TaskTypeFloating  TaskType = "floating"  // No due date, do whenever
	TaskTypeRecurring TaskType = "recurring" // Repeats on a schedule
	TaskTypeHabit     TaskType = "habit"     // Repeating habit with streak tracking
	TaskTypeRoutine   TaskType = "routine"   // Ordered sequence of steps
)

// Status tracks where a task is in its lifecycle.
type Status string

// Valid Status values.
const (
	StatusPending   Status = "pending"   // Not started yet
	StatusActive    Status = "active"    // Currently being worked on
	StatusPaused    Status = "paused"    // Temporarily on hold
	StatusDone      Status = "done"      // Completed
	StatusCancelled Status = "cancelled" // Won't be done
	StatusWaiting   Status = "waiting"   // Blocked on something external
)

// Priority indicates how important/urgent a task is.
// Used by the scheduling algorithm to rank suggestions.
type Priority string

// Valid Priority values.
const (
	PriorityNone   Priority = "none"   // Default — no special priority
	PriorityLow    Priority = "low"    // Nice to do
	PriorityMedium Priority = "medium" // Should do
	PriorityHigh   Priority = "high"   // Important
	PriorityUrgent Priority = "urgent" // Do ASAP
)

// Energy represents how much mental/physical energy a task requires.
// Used by the "bt now" scheduling algorithm to match tasks to your
// current energy level.
type Energy string

// Valid Energy values.
const (
	EnergyLow    Energy = "low"    // Quick/easy tasks (e.g., filing, sorting)
	EnergyMedium Energy = "medium" // Normal effort (e.g., cooking, email)
	EnergyHigh   Energy = "high"   // Deep work (e.g., coding, writing)
)

// RecurrenceAnchor determines whether recurring tasks schedule from
// a fixed date or from the last completion date.
type RecurrenceAnchor string

// Valid RecurrenceAnchor values.
const (
	RecurrenceAnchorFixed      RecurrenceAnchor = "fixed"      // Next occurrence is calendar-based
	RecurrenceAnchorCompletion RecurrenceAnchor = "completion" // Next occurrence is relative to last completion
)

// Link represents a relationship between two tasks.
type Link struct {
	Type   LinkType `yaml:"type"`   // Kind of relationship
	Target string   `yaml:"target"` // ULID of the linked task
}

// LinkType defines how two tasks are related.
type LinkType string

// Valid LinkType values.
const (
	LinkDependsOn LinkType = "depends-on" // This task requires the target to be done first
	LinkBlocks    LinkType = "blocks"     // This task prevents the target from starting
	LinkRelatedTo LinkType = "related-to" // Informational — these tasks are related
)

// HabitFrequency defines how often a habit should be performed.
type HabitFrequency struct {
	Type   string `yaml:"type"`   // "daily" or "weekly"
	Target int    `yaml:"target"` // How many times per period (e.g., 3 times per week)
}

// RoutineStep is one step in a routine's ordered sequence.
type RoutineStep struct {
	Title    string `yaml:"title"`              // Display name for this step
	Duration int    `yaml:"duration,omitempty"` // Estimated duration in minutes (0 = untimed)
	Ref      string `yaml:"ref,omitempty"`      // Optional ULID of a linked task/habit
	Optional bool   `yaml:"optional,omitempty"` // Can this step be skipped?
}

// RoutineSchedule defines when a routine should be triggered.
type RoutineSchedule struct {
	Time string   `yaml:"time,omitempty"` // Time of day (e.g., "07:00")
	Days []string `yaml:"days,omitempty"` // Days of week (e.g., ["mon","tue","wed"])
}

// Task is the core data structure for BentoTask. It represents a task,
// habit, or routine stored as a Markdown file with YAML frontmatter.
//
// The struct tags map to the frontmatter field names defined in ADR-002.
// Fields marked `omitempty` are optional and omitted from YAML when empty.
type Task struct {
	// === Required fields ===
	ID      string    `yaml:"id"`
	Title   string    `yaml:"title"`
	Type    TaskType  `yaml:"type"`
	Status  Status    `yaml:"status"`
	Created time.Time `yaml:"created"`
	Updated time.Time `yaml:"updated"`

	// === Optional common fields ===
	Priority          Priority   `yaml:"priority,omitempty"`
	Energy            Energy     `yaml:"energy,omitempty"`
	EstimatedDuration int        `yaml:"estimated_duration,omitempty"` // minutes
	DueDate           string     `yaml:"due_date,omitempty"`           // ISO date (YYYY-MM-DD)
	DueStart          string     `yaml:"due_start,omitempty"`          // ISO date for ranged tasks
	DueEnd            string     `yaml:"due_end,omitempty"`            // ISO date for ranged tasks
	Tags              []string   `yaml:"tags,omitempty"`
	Context           []string   `yaml:"context,omitempty"`
	Box               string     `yaml:"box,omitempty"`
	Links             []Link     `yaml:"links,omitempty"`
	CompletedAt       *time.Time `yaml:"completed_at,omitempty"`

	// === Recurrence fields ===
	Recurrence       string           `yaml:"recurrence,omitempty"`        // RFC 5545 RRULE string
	RecurrenceAnchor RecurrenceAnchor `yaml:"recurrence_anchor,omitempty"` // "fixed" or "completion"

	// === Habit-specific fields ===
	Frequency     *HabitFrequency `yaml:"frequency,omitempty"`
	StreakCurrent int             `yaml:"streak_current,omitempty"`
	StreakLongest int             `yaml:"streak_longest,omitempty"`

	// === Routine-specific fields ===
	Steps    []RoutineStep    `yaml:"steps,omitempty"`
	Schedule *RoutineSchedule `yaml:"schedule,omitempty"`

	// === Body (not in frontmatter — stored as Markdown after the --- delimiters) ===
	Body string `yaml:"-"`
}
