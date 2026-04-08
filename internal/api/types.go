// Package api provides the REST API server and shared JSON types for BentoTask.
//
// JSON types defined here are shared between the CLI (--json output) and the
// REST API responses, ensuring a single source of truth for the wire format.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/tesserabox/bentotask/internal/engine"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

// TaskJSON is the JSON representation of a task.
// Used by --json output mode per ADR-003 §8 and by the REST API.
type TaskJSON struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Type              string   `json:"type"`
	Status            string   `json:"status"`
	Priority          string   `json:"priority,omitempty"`
	Energy            string   `json:"energy,omitempty"`
	EstimatedDuration int      `json:"estimated_duration,omitempty"`
	DueDate           string   `json:"due_date,omitempty"`
	DueStart          string   `json:"due_start,omitempty"`
	DueEnd            string   `json:"due_end,omitempty"`
	Box               string   `json:"box,omitempty"`
	Recurrence        string   `json:"recurrence,omitempty"`
	CompletedAt       string   `json:"completed_at,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
	Tags              []string `json:"tags"`
	Contexts          []string `json:"contexts"`
	FilePath          string   `json:"file_path"`
	Body              string   `json:"body,omitempty"`

	// Routine-specific fields
	Steps    []StepJSON    `json:"steps,omitempty"`
	Schedule *ScheduleJSON `json:"schedule,omitempty"`

	// Links (populated by show command when links exist)
	Links []map[string]string `json:"links,omitempty"`
}

// StepJSON is the JSON representation of a routine step.
type StepJSON struct {
	Title    string `json:"title"`
	Duration int    `json:"duration,omitempty"`
	Ref      string `json:"ref,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

// ScheduleJSON is the JSON representation of a routine schedule.
type ScheduleJSON struct {
	Time string   `json:"time,omitempty"`
	Days []string `json:"days,omitempty"`
}

// SuggestionJSON is the JSON representation for bt now / bt plan today.
type SuggestionJSON struct {
	TaskID   string                `json:"task_id"`
	Title    string                `json:"title"`
	Duration int                   `json:"duration"`
	Score    engine.ScoreBreakdown `json:"score"`
	Priority string                `json:"priority,omitempty"`
	Energy   string                `json:"energy,omitempty"`
	DueDate  string                `json:"due_date,omitempty"`
	Tags     []string              `json:"tags"`
	Contexts []string              `json:"contexts"`
}

// PlanJSON is the JSON representation for bt plan today.
type PlanJSON struct {
	Suggestions   []SuggestionJSON `json:"suggestions"`
	TotalDuration int              `json:"total_duration"`
	TimeRemaining int              `json:"time_remaining"`
	AvailableTime int              `json:"available_time"`
}

// TaskToJSON converts a model.Task to its JSON representation.
func TaskToJSON(task *model.Task, relPath string) TaskJSON {
	j := TaskJSON{
		ID:                task.ID,
		Title:             task.Title,
		Type:              string(task.Type),
		Status:            string(task.Status),
		Priority:          string(task.Priority),
		Energy:            string(task.Energy),
		EstimatedDuration: task.EstimatedDuration,
		DueDate:           task.DueDate,
		DueStart:          task.DueStart,
		DueEnd:            task.DueEnd,
		Box:               task.Box,
		Recurrence:        task.Recurrence,
		CreatedAt:         task.Created.UTC().Format(time.RFC3339),
		UpdatedAt:         task.Updated.UTC().Format(time.RFC3339),
		Tags:              task.Tags,
		Contexts:          task.Context,
		FilePath:          relPath,
		Body:              task.Body,
	}
	if task.CompletedAt != nil {
		j.CompletedAt = task.CompletedAt.UTC().Format(time.RFC3339)
	}
	// Routine-specific fields
	if len(task.Steps) > 0 {
		j.Steps = make([]StepJSON, len(task.Steps))
		for i, s := range task.Steps {
			j.Steps[i] = StepJSON{
				Title:    s.Title,
				Duration: s.Duration,
				Ref:      s.Ref,
				Optional: s.Optional,
			}
		}
	}
	if task.Schedule != nil {
		j.Schedule = &ScheduleJSON{
			Time: task.Schedule.Time,
			Days: task.Schedule.Days,
		}
	}
	// Ensure slices are never nil (produces [] instead of null in JSON)
	if j.Tags == nil {
		j.Tags = []string{}
	}
	if j.Contexts == nil {
		j.Contexts = []string{}
	}
	return j
}

// IndexedToJSON converts an IndexedTask to its JSON representation.
func IndexedToJSON(t *store.IndexedTask) TaskJSON {
	j := TaskJSON{
		ID:        t.ID,
		Title:     t.Title,
		Type:      t.Type,
		Status:    t.Status,
		Tags:      t.Tags,
		Contexts:  t.Contexts,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		FilePath:  t.FilePath,
	}
	if t.Priority != nil {
		j.Priority = *t.Priority
	}
	if t.Energy != nil {
		j.Energy = *t.Energy
	}
	if t.EstimatedDuration != nil {
		j.EstimatedDuration = *t.EstimatedDuration
	}
	if t.DueDate != nil {
		j.DueDate = *t.DueDate
	}
	if t.DueStart != nil {
		j.DueStart = *t.DueStart
	}
	if t.DueEnd != nil {
		j.DueEnd = *t.DueEnd
	}
	if t.Box != nil {
		j.Box = *t.Box
	}
	if t.Recurrence != nil {
		j.Recurrence = *t.Recurrence
	}
	if t.CompletedAt != nil {
		j.CompletedAt = *t.CompletedAt
	}
	// Ensure slices are never nil
	if j.Tags == nil {
		j.Tags = []string{}
	}
	if j.Contexts == nil {
		j.Contexts = []string{}
	}
	return j
}

// SuggestionToJSON converts an engine.Suggestion to its JSON representation.
func SuggestionToJSON(s engine.Suggestion) SuggestionJSON {
	j := SuggestionJSON{
		TaskID:   s.Task.ID,
		Title:    s.Task.Title,
		Duration: s.Duration,
		Score:    s.Score,
		Priority: string(s.Task.Priority),
		Energy:   string(s.Task.Energy),
		DueDate:  s.Task.DueDate,
		Tags:     s.Task.Tags,
		Contexts: s.Task.Context,
	}
	if j.Tags == nil {
		j.Tags = []string{}
	}
	if j.Contexts == nil {
		j.Contexts = []string{}
	}
	return j
}

// SuggestionsToJSON converts a slice of engine.Suggestion to JSON representation.
func SuggestionsToJSON(suggestions []engine.Suggestion) []SuggestionJSON {
	items := make([]SuggestionJSON, len(suggestions))
	for i, s := range suggestions {
		items[i] = SuggestionToJSON(s)
	}
	return items
}

// PlanToJSON converts a PackResult to its JSON representation.
func PlanToJSON(result *engine.PackResult, availTime int) PlanJSON {
	items := make([]SuggestionJSON, len(result.Suggestions))
	for i, s := range result.Suggestions {
		items[i] = SuggestionToJSON(s)
	}
	return PlanJSON{
		Suggestions:   items,
		TotalDuration: result.TotalDuration,
		TimeRemaining: result.TimeRemaining,
		AvailableTime: availTime,
	}
}

// WriteJSON encodes a value as indented JSON and writes it to w.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
