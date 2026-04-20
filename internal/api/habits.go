package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
)

// createHabitRequest is the JSON body for POST /habits.
type createHabitRequest struct {
	Title        string   `json:"title"`
	FreqType     string   `json:"freq_type"`
	FreqTarget   int      `json:"freq_target"`
	MaxPerPeriod int      `json:"max_per_period,omitempty"`
	Recurrence   string   `json:"recurrence,omitempty"`
	Priority     string   `json:"priority,omitempty"`
	Energy       string   `json:"energy,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Contexts     []string `json:"contexts,omitempty"`
}

// logHabitRequest is the JSON body for POST /habits/:id/log.
type logHabitRequest struct {
	Duration int    `json:"duration,omitempty"`
	Note     string `json:"note,omitempty"`
}

func (s *Server) handleCreateHabit(w http.ResponseWriter, r *http.Request) {
	var req createHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondBadRequest(w, "invalid JSON: "+err.Error())
		return
	}

	if req.Title == "" {
		respondValidationError(w, "title is required")
		return
	}

	opts := app.HabitOptions{
		FreqType:     req.FreqType,
		FreqTarget:   req.FreqTarget,
		MaxPerPeriod: req.MaxPerPeriod,
		Recurrence:   req.Recurrence,
		Priority:     model.Priority(req.Priority),
		Energy:       model.Energy(req.Energy),
		Tags:         req.Tags,
		Context:      req.Contexts,
	}
	if opts.FreqType == "" {
		opts.FreqType = "daily"
	}
	if opts.FreqTarget == 0 {
		opts.FreqTarget = 1
	}
	// Auto-generate RRULE from frequency if not explicitly provided
	if opts.Recurrence == "" {
		switch opts.FreqType {
		case "daily":
			opts.Recurrence = "FREQ=DAILY"
		case "weekly":
			opts.Recurrence = "FREQ=WEEKLY"
		default:
			respondValidationError(w, "unknown frequency type: "+opts.FreqType)
			return
		}
	}

	s.mu.Lock()
	task, err := s.app.AddHabit(req.Title, opts)
	s.mu.Unlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}

	s.mu.RLock()
	_, relPath, _ := s.app.GetTask(task.ID)
	s.mu.RUnlock()

	respondJSON(w, http.StatusCreated, TaskToJSON(task, relPath))
}

func (s *Server) handleListHabits(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	habits, err := s.app.ListHabits()
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}

	items := make([]TaskJSON, len(habits))
	for i, h := range habits {
		items[i] = IndexedToJSON(h)
	}

	respondJSON(w, http.StatusOK, collectionResponse{Items: items, Count: len(items)})
}

func (s *Server) handleLogHabit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req logHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondBadRequest(w, "invalid JSON: "+err.Error())
		return
	}

	s.mu.Lock()
	task, err := s.app.LogHabit(id, req.Duration, req.Note)
	s.mu.Unlock()
	if err != nil {
		if isNotFound(err) {
			respondNotFound(w, err.Error())
			return
		}
		respondInternalError(w, err)
		return
	}

	s.mu.RLock()
	_, relPath, _ := s.app.GetTask(task.ID)
	s.mu.RUnlock()

	respondJSON(w, http.StatusOK, TaskToJSON(task, relPath))
}

func (s *Server) handleGetHabitStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.RLock()
	task, stats, err := s.app.HabitStats(id)
	s.mu.RUnlock()
	if err != nil {
		if isNotFound(err) {
			respondNotFound(w, err.Error())
			return
		}
		respondInternalError(w, err)
		return
	}

	s.mu.RLock()
	_, relPath, _ := s.app.GetTask(task.ID)
	s.mu.RUnlock()

	respondJSON(w, http.StatusOK, map[string]any{
		"task":  TaskToJSON(task, relPath),
		"stats": stats,
	})
}
