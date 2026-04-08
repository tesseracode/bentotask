package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
)

// createRoutineRequest is the JSON body for POST /routines.
type createRoutineRequest struct {
	Title    string                 `json:"title"`
	Steps    []model.RoutineStep    `json:"steps"`
	Schedule *model.RoutineSchedule `json:"schedule,omitempty"`
	Priority string                 `json:"priority,omitempty"`
	Energy   string                 `json:"energy,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
}

func (s *Server) handleCreateRoutine(w http.ResponseWriter, r *http.Request) {
	var req createRoutineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondBadRequest(w, "invalid JSON: "+err.Error())
		return
	}

	if req.Title == "" {
		respondValidationError(w, "title is required")
		return
	}

	opts := app.RoutineOptions{
		Steps:    req.Steps,
		Schedule: req.Schedule,
		Priority: model.Priority(req.Priority),
		Energy:   model.Energy(req.Energy),
		Tags:     req.Tags,
	}

	s.mu.Lock()
	task, err := s.app.AddRoutine(req.Title, opts)
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

func (s *Server) handleListRoutines(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	routines, err := s.app.ListRoutines()
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}

	items := make([]TaskJSON, len(routines))
	for i, r := range routines {
		items[i] = IndexedToJSON(r)
	}

	respondJSON(w, http.StatusOK, collectionResponse{Items: items, Count: len(items)})
}

func (s *Server) handleGetRoutine(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.RLock()
	task, relPath, err := s.app.GetTask(id)
	s.mu.RUnlock()
	if err != nil {
		if isNotFound(err) {
			respondNotFound(w, err.Error())
			return
		}
		respondInternalError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, TaskToJSON(task, relPath))
}
