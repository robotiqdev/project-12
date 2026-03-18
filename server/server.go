// Package server provides the HTTP server and handler logic.
package server

import "net/http"

// Config holds the configuration for the server.
type Config struct {
	// Version is the application version string to expose via the /version endpoint.
	Version string
}

// Server is the HTTP server.
type Server struct {
	config Config
	mux    *http.ServeMux
}

// New creates a new Server with the given configuration.
func New(cfg Config) *Server {
	s := &Server{
		config: cfg,
		mux:    http.NewServeMux(),
	}
	s.routes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/version", s.handleVersion)
}
