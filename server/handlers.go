package server

import (
	"encoding/json"
	"net/http"
)

// handleVersion responds with a JSON object containing the application version.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(map[string]string{
		"version": s.config.Version,
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
