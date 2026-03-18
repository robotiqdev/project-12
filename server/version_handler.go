package server

import "net/http"

type versionResponse struct {
	Version string `json:"version"`
}

// handleVersion returns an http.HandlerFunc that responds with the server version.
// Implementation pending.
func (s *Server) handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// not yet implemented
	}
}
