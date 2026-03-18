package server

import (
	"encoding/json"
	"net/http"
)

// handleVersion responds with a JSON object containing the application version.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": s.config.Version,
	})
}
