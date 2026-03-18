package server

import (
	"encoding/json"
	"net/http"
)

type versionResponse struct {
	Version string `json:"version"`
}

// handleVersion returns an http.HandlerFunc that responds with the server version.
func (s *Server) handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := versionResponse{Version: s.version}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
