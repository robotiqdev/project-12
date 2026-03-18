package server

import "net/http"

// handleHealth returns an http.HandlerFunc that responds with server health info.
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement
	}
}
