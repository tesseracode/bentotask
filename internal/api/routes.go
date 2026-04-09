package api

import (
	"github.com/go-chi/chi/v5"
)

// buildRouter creates the chi router with all API routes.
func (s *Server) buildRouter() chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(corsMiddleware())
	r.Use(recoverer)
	r.Use(requestLogger)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(jsonContentType)

		// Task endpoints
		r.Post("/tasks", s.handleCreateTask)
		r.Get("/tasks", s.handleListTasks)
		r.Get("/tasks/search", s.handleSearchTasks)
		r.Get("/tasks/{id}", s.handleGetTask)
		r.Patch("/tasks/{id}", s.handleUpdateTask)
		r.Delete("/tasks/{id}", s.handleDeleteTask)
		r.Post("/tasks/{id}/done", s.handleCompleteTask)

		// Link endpoints
		r.Post("/tasks/{id}/links", s.handleCreateLink)
		r.Get("/tasks/{id}/links", s.handleGetLinks)
		r.Delete("/tasks/{id}/links/{targetId}", s.handleDeleteLink)

		// Habit endpoints
		r.Post("/habits", s.handleCreateHabit)
		r.Get("/habits", s.handleListHabits)
		r.Post("/habits/{id}/log", s.handleLogHabit)
		r.Get("/habits/{id}/stats", s.handleGetHabitStats)

		// Routine endpoints
		r.Post("/routines", s.handleCreateRoutine)
		r.Get("/routines", s.handleListRoutines)
		r.Get("/routines/{id}", s.handleGetRoutine)

		// Scheduling endpoints
		r.Get("/suggest", s.handleSuggest)
		r.Get("/plan/today", s.handlePlanToday)

		// Admin endpoints
		r.Post("/index/rebuild", s.handleRebuildIndex)
		r.Get("/meta/tags", s.handleMetaTags)
		r.Get("/meta/boxes", s.handleMetaBoxes)
		r.Get("/meta/contexts", s.handleMetaContexts)
	})

	// Static files (SvelteKit build) — catch-all AFTER /api
	r.Handle("/*", staticFileServer())

	return r
}
