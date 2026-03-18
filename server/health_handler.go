package server

import (
	"encoding/json"
	"net/http"
)

// healthResponse is the JSON body returned by the health endpoint.
type healthResponse struct {
	UptimeSec float64 `json:"uptime_seconds"`
}

// handleHealth returns an http.HandlerFunc that responds with server health info.
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{UptimeSec: s.healthTracker.UptimeSeconds()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
