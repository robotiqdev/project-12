package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// healthResponse is the JSON body returned by the health endpoint.
type healthResponse struct {
	Status    string  `json:"status"`
	UptimeSec float64 `json:"uptime_seconds"`
	Timestamp string  `json:"timestamp"`
}

// handleHealth returns an http.HandlerFunc that responds with server health info.
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := healthResponse{
			Status:    "ok",
			UptimeSec: s.healthTracker.UptimeSeconds(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
