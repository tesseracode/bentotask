package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

// createTaskRequest is the JSON body for POST /tasks.
type createTaskRequest struct {
	Title    string   `json:"title"`
	Priority string   `json:"priority,omitempty"`
	Energy   string   `json:"energy,omitempty"`
	Duration int      `json:"duration,omitempty"`
	DueDate  string   `json:"due_date,omitempty"`
	DueStart string   `json:"due_start,omitempty"`
	DueEnd   string   `json:"due_end,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Contexts []string `json:"contexts,omitempty"`
	Box      string   `json:"box,omitempty"`
}

// updateTaskRequest is the JSON body for PATCH /tasks/:id.
type updateTaskRequest struct {
	Title    *string                `json:"title,omitempty"`
	Priority *string                `json:"priority,omitempty"`
	Energy   *string                `json:"energy,omitempty"`
	Duration *int                   `json:"duration,omitempty"`
	DueDate  *string                `json:"due_date,omitempty"`
	DueStart *string                `json:"due_start,omitempty"`
	DueEnd   *string                `json:"due_end,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Contexts []string               `json:"contexts,omitempty"`
	Box      *string                `json:"box,omitempty"`
	Status   *string                `json:"status,omitempty"`
	Steps    []model.RoutineStep    `json:"steps,omitempty"`
	Schedule *model.RoutineSchedule `json:"schedule,omitempty"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondBadRequest(w, "invalid JSON: "+err.Error())
		return
	}

	if req.Title == "" {
		respondValidationError(w, "title is required")
		return
	}

	opts := app.TaskOptions{
		Priority: model.Priority(req.Priority),
		Energy:   model.Energy(req.Energy),
		Duration: req.Duration,
		DueDate:  req.DueDate,
		DueStart: req.DueStart,
		DueEnd:   req.DueEnd,
		Tags:     req.Tags,
		Context:  req.Contexts,
		Box:      req.Box,
	}

	s.mu.Lock()
	task, err := s.app.AddTask(req.Title, opts)
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

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	f := &store.TaskFilter{}
	q := r.URL.Query()
	if v := q.Get("status"); v != "" {
		f.Status = model.Status(v)
	}
	if v := q.Get("priority"); v != "" {
		f.Priority = model.Priority(v)
	}
	if v := q.Get("energy"); v != "" {
		f.Energy = model.Energy(v)
	}
	if v := q.Get("tag"); v != "" {
		f.Tag = v
	}
	if v := q.Get("box"); v != "" {
		f.Box = v
	}
	if v := q.Get("context"); v != "" {
		f.Context = v
	}

	s.mu.RLock()
	tasks, err := s.app.ListTasks(f)
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}

	items := make([]TaskJSON, len(tasks))
	for i, t := range tasks {
		items[i] = IndexedToJSON(t)
	}

	respondJSON(w, http.StatusOK, collectionResponse{Items: items, Count: len(items)})
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondBadRequest(w, "invalid JSON: "+err.Error())
		return
	}

	s.mu.Lock()
	task, err := s.app.UpdateTask(id, func(t *model.Task) {
		if req.Title != nil {
			t.Title = *req.Title
		}
		if req.Priority != nil {
			t.Priority = model.Priority(*req.Priority)
		}
		if req.Energy != nil {
			t.Energy = model.Energy(*req.Energy)
		}
		if req.Duration != nil {
			t.EstimatedDuration = *req.Duration
		}
		if req.DueDate != nil {
			t.DueDate = *req.DueDate
		}
		if req.DueStart != nil {
			t.DueStart = *req.DueStart
		}
		if req.DueEnd != nil {
			t.DueEnd = *req.DueEnd
		}
		if req.Tags != nil {
			t.Tags = req.Tags
		}
		if req.Contexts != nil {
			t.Context = req.Contexts
		}
		if req.Box != nil {
			t.Box = *req.Box
		}
		if req.Status != nil {
			t.Status = model.Status(*req.Status)
		}
		if req.Steps != nil {
			t.Steps = req.Steps
			total := 0
			for _, s := range req.Steps {
				total += s.Duration
			}
			t.EstimatedDuration = total
		}
		if req.Schedule != nil {
			t.Schedule = req.Schedule
		}
	})
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

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.Lock()
	task, err := s.app.DeleteTask(id)
	s.mu.Unlock()
	if err != nil {
		if isNotFound(err) {
			respondNotFound(w, err.Error())
			return
		}
		respondInternalError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, TaskToJSON(task, ""))
}

func (s *Server) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.Lock()
	task, err := s.app.CompleteTask(id)
	s.mu.Unlock()
	if err != nil {
		if isNotFound(err) {
			respondNotFound(w, err.Error())
			return
		}
		if isConflict(err) {
			respondConflict(w, err.Error())
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

func (s *Server) handleSearchTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondBadRequest(w, "q parameter is required")
		return
	}

	s.mu.RLock()
	tasks, err := s.app.SearchTasks(query)
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}

	items := make([]TaskJSON, len(tasks))
	for i, t := range tasks {
		items[i] = IndexedToJSON(t)
	}

	respondJSON(w, http.StatusOK, collectionResponse{Items: items, Count: len(items)})
}

// isNotFound checks if an error is a "not found" error.
func isNotFound(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "not found") || strings.Contains(msg, "no task")
}

// isConflict checks if an error is a conflict error.
func isConflict(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "already") ||
		strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "cycle")
}
