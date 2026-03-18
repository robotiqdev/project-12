package server

import (
	"github.com/robotiqdev/project-12/internal/health"
)

// Server is the HTTP server and acts as a dependency container.
type Server struct {
	healthTracker *health.Tracker
}

// New creates a new Server with the given health tracker injected at construction time.
func New(tracker *health.Tracker) *Server {
	return &Server{healthTracker: tracker}
}
