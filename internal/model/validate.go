package model

import (
	"fmt"
	"strings"
)

// Validate checks that a Task has all required fields and valid enum values.
// Returns a list of validation errors (empty if valid).
func (t *Task) Validate() []string {
	var errs []string

	// Required fields
	if t.ID == "" {
		errs = append(errs, "id is required")
	}
	if t.Title == "" {
		errs = append(errs, "title is required")
	}
	if t.Type == "" {
		errs = append(errs, "type is required")
	}
	if t.Status == "" {
		errs = append(errs, "status is required")
	}
	if t.Created.IsZero() {
		errs = append(errs, "created is required")
	}
	if t.Updated.IsZero() {
		errs = append(errs, "updated is required")
	}

	// Enum validation
	if t.Type != "" && !isValidTaskType(t.Type) {
		errs = append(errs, fmt.Sprintf("invalid type: %q", t.Type))
	}
	if t.Status != "" && !isValidStatus(t.Status) {
		errs = append(errs, fmt.Sprintf("invalid status: %q", t.Status))
	}
	if t.Priority != "" && !isValidPriority(t.Priority) {
		errs = append(errs, fmt.Sprintf("invalid priority: %q", t.Priority))
	}
	if t.Energy != "" && !isValidEnergy(t.Energy) {
		errs = append(errs, fmt.Sprintf("invalid energy: %q", t.Energy))
	}

	// Type-specific validation
	switch t.Type {
	case TaskTypeDated:
		if t.DueDate == "" {
			errs = append(errs, "dated tasks require due_date")
		}
	case TaskTypeRanged:
		if t.DueStart == "" || t.DueEnd == "" {
			errs = append(errs, "ranged tasks require due_start and due_end")
		}
	case TaskTypeRecurring:
		if t.Recurrence == "" {
			errs = append(errs, "recurring tasks require recurrence rule")
		}
	case TaskTypeHabit:
		if t.Frequency == nil {
			errs = append(errs, "habits require frequency")
		}
		if t.Recurrence == "" {
			errs = append(errs, "habits require recurrence rule")
		}
	case TaskTypeRoutine:
		if len(t.Steps) == 0 {
			errs = append(errs, "routines require at least one step")
		}
	}

	// Link validation
	for i, link := range t.Links {
		if link.Target == "" {
			errs = append(errs, fmt.Sprintf("link[%d]: target is required", i))
		}
		if !isValidLinkType(link.Type) {
			errs = append(errs, fmt.Sprintf("link[%d]: invalid type: %q", i, link.Type))
		}
	}

	return errs
}

// IsValid returns true if the task passes all validation checks.
func (t *Task) IsValid() bool {
	return len(t.Validate()) == 0
}

// IsDone returns true if the task is in a terminal state.
func (t *Task) IsDone() bool {
	return t.Status == StatusDone || t.Status == StatusCancelled
}

// ShortID returns the first n characters of the ULID for display.
// ULIDs are 26 chars; 8 chars is usually enough to be unique.
func (t *Task) ShortID(n int) string {
	if n >= len(t.ID) {
		return t.ID
	}
	return t.ID[:n]
}

// HasTag returns true if the task has the given tag (case-insensitive).
func (t *Task) HasTag(tag string) bool {
	tag = strings.ToLower(tag)
	for _, tt := range t.Tags {
		if strings.ToLower(tt) == tag {
			return true
		}
	}
	return false
}

// HasContext returns true if the task has the given context (case-insensitive).
func (t *Task) HasContext(ctx string) bool {
	ctx = strings.ToLower(ctx)
	for _, c := range t.Context {
		if strings.ToLower(c) == ctx {
			return true
		}
	}
	return false
}

// --- Enum validation helpers ---

func isValidTaskType(t TaskType) bool {
	switch t {
	case TaskTypeOneShot, TaskTypeDated, TaskTypeRanged, TaskTypeFloating,
		TaskTypeRecurring, TaskTypeHabit, TaskTypeRoutine:
		return true
	}
	return false
}

func isValidStatus(s Status) bool {
	switch s {
	case StatusPending, StatusActive, StatusPaused, StatusDone,
		StatusCancelled, StatusWaiting:
		return true
	}
	return false
}

func isValidPriority(p Priority) bool {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	}
	return false
}

func isValidEnergy(e Energy) bool {
	switch e {
	case EnergyLow, EnergyMedium, EnergyHigh:
		return true
	}
	return false
}

func isValidLinkType(lt LinkType) bool {
	switch lt {
	case LinkDependsOn, LinkBlocks, LinkRelatedTo:
		return true
	}
	return false
}
