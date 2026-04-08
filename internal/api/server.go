package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/tesserabox/bentotask/internal/app"
)

// Server is the REST API server for BentoTask.
type Server struct {
	app    *app.App
	router chi.Router
	mu     sync.RWMutex
	server *http.Server
}

// NewServer creates a new API server wrapping the given App.
func NewServer(a *app.App) *Server {
	s := &Server{app: a}
	s.router = s.buildRouter()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
