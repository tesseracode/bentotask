package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

// TaskJSON is the JSON representation of a task.
// Used by --json output mode per ADR-003 §8.
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
}

// taskToJSON converts a model.Task to its JSON representation.
func taskToJSON(task *model.Task, relPath string) TaskJSON {
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
	// Ensure slices are never nil (produces [] instead of null in JSON)
	if j.Tags == nil {
		j.Tags = []string{}
	}
	if j.Contexts == nil {
		j.Contexts = []string{}
	}
	return j
}

// indexedToJSON converts an IndexedTask to its JSON representation.
func indexedToJSON(t *store.IndexedTask) TaskJSON {
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

// writeJSON encodes a value as indented JSON and writes it to w.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
