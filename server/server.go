package server

import "github.com/robotiqdev/project-12/health"

// Server holds dependencies for HTTP handlers.
type Server struct {
	healthTracker *health.Tracker
}
