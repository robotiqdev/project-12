package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robotiqdev/project-12/internal/health"
)

// TestHandleHealth_UptimeSecondsFieldPresent verifies that the health endpoint
// response body contains an "uptime_seconds" field.
func TestHandleHealth_UptimeSecondsFieldPresent(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := health.NewTracker(func() time.Time { return fixed })
	srv := New(tracker)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if _, ok := body["uptime_seconds"]; !ok {
		t.Fatal("response body missing required field: uptime_seconds")
	}
}

// TestHandleHealth_UptimeSecondsIsFloat64 verifies that the "uptime_seconds"
// field in the health response is a float64 (numeric JSON value).
func TestHandleHealth_UptimeSecondsIsFloat64(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := health.NewTracker(func() time.Time { return fixed })
	srv := New(tracker)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	val, ok := body["uptime_seconds"]
	if !ok {
		t.Fatal("response body missing required field: uptime_seconds")
	}

	if _, ok := val.(float64); !ok {
		t.Errorf("uptime_seconds has type %T; want float64", val)
	}
}

// TestHandleHealth_UptimeSecondsNonNegative verifies that the "uptime_seconds"
// field is >= 0 (time cannot go backwards from start).
func TestHandleHealth_UptimeSecondsNonNegative(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := health.NewTracker(func() time.Time { return fixed })
	srv := New(tracker)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	uptime, ok := body["uptime_seconds"].(float64)
	if !ok {
		t.Fatal("uptime_seconds is not a float64")
	}

	if uptime < 0 {
		t.Errorf("uptime_seconds = %v; want >= 0", uptime)
	}
}

// TestHandleHealth_UptimeSecondsZeroWhenNoTimeElapsed verifies that when the
// now function returns the same instant at construction and at request time,
// uptime_seconds is 0.0.
func TestHandleHealth_UptimeSecondsZeroWhenNoTimeElapsed(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := health.NewTracker(func() time.Time { return fixed })
	srv := New(tracker)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	uptime, ok := body["uptime_seconds"].(float64)
	if !ok {
		t.Fatal("uptime_seconds is not a float64")
	}

	if uptime != 0.0 {
		t.Errorf("uptime_seconds = %v; want 0.0 when no time has elapsed", uptime)
	}
}

// TestHandleHealth_UptimeSecondsExactValueAfterAdvancingClock verifies that
// the handler reads uptime from the tracker by advancing the injected clock
// by a known duration and asserting the exact value in the response.
func TestHandleHealth_UptimeSecondsExactValueAfterAdvancingClock(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	nowTime := t0
	tracker := health.NewTracker(func() time.Time { return nowTime })

	// Advance the clock by exactly 10 seconds before calling the handler.
	nowTime = t0.Add(10 * time.Second)

	srv := New(tracker)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	uptime, ok := body["uptime_seconds"].(float64)
	if !ok {
		t.Fatal("uptime_seconds is not a float64")
	}

	const want = 10.0
	if uptime != want {
		t.Errorf("uptime_seconds = %v; want %v (10 seconds after tracker was created)", uptime, want)
	}
}

// TestHandleHealth_UptimeSecondsFractionalValue verifies that fractional uptime
// (e.g., 1.5 seconds) is correctly represented in the response.
func TestHandleHealth_UptimeSecondsFractionalValue(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	nowTime := t0
	tracker := health.NewTracker(func() time.Time { return nowTime })

	// Advance the clock by 1.5 seconds.
	nowTime = t0.Add(1500 * time.Millisecond)

	srv := New(tracker)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	uptime, ok := body["uptime_seconds"].(float64)
	if !ok {
		t.Fatal("uptime_seconds is not a float64")
	}

	const want = 1.5
	if uptime != want {
		t.Errorf("uptime_seconds = %v; want %v (1.5 seconds elapsed)", uptime, want)
	}
}

// TestNew_AcceptsHealthTracker verifies that New() accepts a *health.Tracker
// and the server is properly initialized.
func TestNew_AcceptsHealthTracker(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := health.NewTracker(func() time.Time { return fixed })

	srv := New(tracker)
	if srv == nil {
		t.Fatal("New() returned nil; want a non-nil *Server")
	}
}
