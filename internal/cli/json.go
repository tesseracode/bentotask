package cli

// JSON types and converters live in internal/api/types.go (shared with REST API).
// This file provides type aliases and unexported wrappers so existing CLI code
// doesn't need the api. prefix on every call.

import (
	"io"

	"github.com/tesserabox/bentotask/internal/api"
	"github.com/tesserabox/bentotask/internal/engine"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

// TaskJSON is an alias for api.TaskJSON.
type TaskJSON = api.TaskJSON

// StepJSON is an alias for api.StepJSON.
type StepJSON = api.StepJSON

// ScheduleJSON is an alias for api.ScheduleJSON.
type ScheduleJSON = api.ScheduleJSON

// SuggestionJSON is an alias for api.SuggestionJSON.
type SuggestionJSON = api.SuggestionJSON

// PlanJSON is an alias for api.PlanJSON.
type PlanJSON = api.PlanJSON

func taskToJSON(task *model.Task, relPath string) api.TaskJSON {
	return api.TaskToJSON(task, relPath)
}

func indexedToJSON(t *store.IndexedTask) api.TaskJSON {
	return api.IndexedToJSON(t)
}

func writeJSON(w io.Writer, v any) error {
	return api.WriteJSON(w, v)
}

func suggestionsToJSON(suggestions []engine.Suggestion) []api.SuggestionJSON {
	return api.SuggestionsToJSON(suggestions)
}

func planToJSON(result *engine.PackResult, availTime int) api.PlanJSON {
	return api.PlanToJSON(result, availTime)
}
