package server

import (
	"net/http"

	"github.com/robotiqdev/project-12/health"
)

// Config holds server configuration.
type Config struct {
	Addr    string
	Version string
}

// Server holds dependencies for HTTP handlers.
type Server struct {
	cfg           Config
	router        *http.ServeMux
	healthTracker *health.Tracker
	version       string
}

// New creates a new Server with the given config and health tracker, registers routes, and returns it.
func New(cfg Config, h *health.Tracker) *Server {
	s := &Server{
		cfg:           cfg,
		router:        http.NewServeMux(),
		healthTracker: h,
		version:       cfg.Version,
	}
	s.routes()
	return s
}

// ServeHTTP delegates to the server's router, implementing http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// routes registers all HTTP routes on the server's mux.
func (s *Server) routes() {
	s.router.HandleFunc("/health", s.handleHealth())
}
