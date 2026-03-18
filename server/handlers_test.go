package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robotiqdev/project-12/internal/health"
)

// newTestServer creates a minimal Server for handler tests.
func newTestServer() *Server {
	return &Server{healthTracker: health.NewTracker(time.Now)}
}

// newRoutedServer creates a Server via New() with a router and registered routes.
func newRoutedServer() *Server {
	cfg := Config{Addr: ":0", Version: "test"}
	return New(cfg, health.NewTracker(time.Now))
}

// TestHandleHealth_UptimeSecondsFieldPresent verifies that the health endpoint
// response body contains an "uptime_seconds" field.
func TestHandleHealth_UptimeSecondsFieldPresent(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := health.NewTracker(func() time.Time { return fixed })
	srv := New(Config{Addr: ":0"}, tracker)

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
	srv := New(Config{Addr: ":0"}, tracker)

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
	srv := New(Config{Addr: ":0"}, tracker)

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
	srv := New(Config{Addr: ":0"}, tracker)

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

	srv := New(Config{Addr: ":0"}, tracker)

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

	srv := New(Config{Addr: ":0"}, tracker)

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

	srv := New(Config{Addr: ":0"}, tracker)
	if srv == nil {
		t.Fatal("New() returned nil; want a non-nil *Server")
	}
}

// TestHandleHealth_GET_StatusOK verifies that GET /health returns HTTP 200.
func TestHandleHealth_GET_StatusOK(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestHandleHealth_ContentTypeJSON verifies that GET /health sets Content-Type: application/json.
func TestHandleHealth_ContentTypeJSON(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth()(w, req)

	ct := w.Result().Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

// TestHandleHealth_BodyStatusFieldIsOk verifies the response body contains status == "ok".
func TestHandleHealth_BodyStatusFieldIsOk(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON response body: %v", err)
	}

	statusVal, ok := body["status"]
	if !ok {
		t.Fatal("expected 'status' field in JSON response body")
	}
	statusStr, ok := statusVal.(string)
	if !ok {
		t.Errorf("expected 'status' to be a string, got %T", statusVal)
	} else if statusStr != "ok" {
		t.Errorf("expected status == 'ok', got %q", statusStr)
	}
}

// TestHandleHealth_BodyUptimeSecondsIsNonNegativeFloat64 verifies that uptime_seconds is a float64 >= 0.
func TestHandleHealth_BodyUptimeSecondsIsNonNegativeFloat64(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON response body: %v", err)
	}

	uptimeVal, ok := body["uptime_seconds"]
	if !ok {
		t.Fatal("expected 'uptime_seconds' field in JSON response body")
	}
	uptimeFloat, ok := uptimeVal.(float64)
	if !ok {
		t.Errorf("expected 'uptime_seconds' to be float64, got %T", uptimeVal)
	} else if uptimeFloat < 0 {
		t.Errorf("expected 'uptime_seconds' >= 0, got %f", uptimeFloat)
	}
}

// TestHandleHealth_BodyTimestampIsValidRFC3339 verifies that timestamp is a valid RFC3339 string.
func TestHandleHealth_BodyTimestampIsValidRFC3339(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth()(w, req)

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON response body: %v", err)
	}

	tsVal, ok := body["timestamp"]
	if !ok {
		t.Fatal("expected 'timestamp' field in JSON response body")
	}
	tsStr, ok := tsVal.(string)
	if !ok {
		t.Errorf("expected 'timestamp' to be a string, got %T", tsVal)
	} else {
		if _, err := time.Parse(time.RFC3339, tsStr); err != nil {
			t.Errorf("expected 'timestamp' to be valid RFC3339, got %q: %v", tsStr, err)
		}
	}
}

// TestHandleHealth_BodyIsValidJSON verifies the response body is valid JSON (can be decoded).
func TestHandleHealth_BodyIsValidJSON(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth()(w, req)

	var body interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Errorf("response body is not valid JSON: %v", err)
	}
}

// TestHandleHealth_POST_Returns405 verifies that POST /health returns HTTP 405 Method Not Allowed.
func TestHandleHealth_POST_Returns405(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth()(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 for POST /health, got %d", w.Code)
	}
}

// TestRoute_Health_GET_Returns200 verifies that GET /health returns 200 via the router (ServeHTTP).
func TestRoute_Health_GET_Returns200(t *testing.T) {
	s := newRoutedServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for GET /health, got %d", w.Code)
	}
}

// TestRoute_Health_GET_NoAuthorizationHeaderRequired verifies that GET /health succeeds
// without an Authorization header — the route is intentionally public.
func TestRoute_Health_GET_NoAuthorizationHeaderRequired(t *testing.T) {
	s := newRoutedServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// Deliberately do NOT set Authorization header.
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for unauthenticated GET /health, got %d", w.Code)
	}
}

// TestRoute_Health_GET_WithAuthorizationHeader verifies that GET /health also succeeds
// when an Authorization header is present (no auth middleware to reject it).
func TestRoute_Health_GET_WithAuthorizationHeader(t *testing.T) {
	s := newRoutedServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for GET /health with Authorization header, got %d", w.Code)
	}
}

// TestRoute_Missing_GET_Returns404 verifies that an unregistered path returns 404.
func TestRoute_Missing_GET_Returns404(t *testing.T) {
	s := newRoutedServer()

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for GET /missing, got %d", w.Code)
	}
}

// TestRoute_Unknown_Path_Returns404 verifies that an arbitrary unknown path returns 404.
func TestRoute_Unknown_Path_Returns404(t *testing.T) {
	s := newRoutedServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for GET /api/v1/unknown, got %d", w.Code)
	}
}

// TestRoute_Health_POST_Returns405 verifies that POST /health is rejected via the router.
func TestRoute_Health_POST_Returns405(t *testing.T) {
	s := newRoutedServer()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 for POST /health via router, got %d", w.Code)
	}
}

// TestServer_ServeHTTP_IsHTTPHandler verifies that *Server implements http.Handler.
func TestServer_ServeHTTP_IsHTTPHandler(t *testing.T) {
	s := newRoutedServer()
	var _ http.Handler = s
	// If this compiles, *Server implements http.Handler.
}

// testHealthResponse is a local mirror struct for decoding the health endpoint JSON response.
type testHealthResponse struct {
	Status    string  `json:"status"`
	UptimeSec float64 `json:"uptime_seconds"`
	Timestamp string  `json:"timestamp"`
}

// TestHandleHealth is an integration test for the /health endpoint using ServeHTTP.
// It covers: GET 200 with correct body fields, POST 405, and no auth header required.
func TestHandleHealth(t *testing.T) {
	t.Run("GET returns 200 with correct body", func(t *testing.T) {
		tracker := health.NewTracker(time.Now)
		srv := New(Config{Addr: ":0", Version: "test"}, tracker)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var resp testHealthResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}

		if resp.Status != "ok" {
			t.Errorf("expected status == \"ok\", got %q", resp.Status)
		}

		if resp.UptimeSec < 0 {
			t.Errorf("expected uptime_seconds >= 0, got %f", resp.UptimeSec)
		}

		if _, err := time.Parse(time.RFC3339, resp.Timestamp); err != nil {
			t.Errorf("expected timestamp to be valid RFC3339, got %q: %v", resp.Timestamp, err)
		}
	})

	t.Run("POST returns 405", func(t *testing.T) {
		tracker := health.NewTracker(time.Now)
		srv := New(Config{Addr: ":0", Version: "test"}, tracker)

		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 for POST /health, got %d", w.Code)
		}
	})

	t.Run("GET without Authorization header returns 200", func(t *testing.T) {
		tracker := health.NewTracker(time.Now)
		srv := New(Config{Addr: ":0", Version: "test"}, tracker)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		// Deliberately do NOT set Authorization header — endpoint must be public.
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for unauthenticated GET /health, got %d", w.Code)
		}
	})
}
