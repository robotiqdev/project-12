package server

import "net/http"

// Server is the HTTP server for the application.
type Server struct {
	version string
	router  *http.ServeMux
}

// Config holds configuration for constructing a Server.
type Config struct {
	Version string
}

// New creates a new Server from the given Config.
func New(cfg Config) *Server {
	s := &Server{version: cfg.Version, router: http.NewServeMux()}
	s.routes()
	return s
}

// routes registers all HTTP routes.
func (s *Server) routes() {
	s.router.HandleFunc("GET /version", s.handleVersion())
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
