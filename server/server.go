package server

import "net/http"

// Server is the HTTP server for the application.
type Server struct {
	version string
}

// Config holds configuration for constructing a Server.
type Config struct {
	Version string
}

// New creates a new Server from the given Config.
func New(cfg Config) *Server {
	return &Server{version: cfg.Version}
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/version", s.handleVersion())
	mux.ServeHTTP(w, r)
}
