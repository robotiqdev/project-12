package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleVersion_StatusOK verifies that GET /version returns HTTP 200.
func TestHandleVersion_StatusOK(t *testing.T) {
	s := New(Config{Version: "1.2.3"})
	handler := s.handleVersion()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// TestHandleVersion_ContentTypeIsJSON verifies that GET /version sets Content-Type to application/json.
func TestHandleVersion_ContentTypeIsJSON(t *testing.T) {
	s := New(Config{Version: "1.2.3"})
	handler := s.handleVersion()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}
}

// TestHandleVersion_BodyVersionFieldMatchesConfig verifies that the response body's
// `version` field equals the version string used to construct the server.
func TestHandleVersion_BodyVersionFieldMatchesConfig(t *testing.T) {
	const testVersion = "2.5.0"
	s := New(Config{Version: testVersion})
	handler := s.handleVersion()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var resp versionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.Version != testVersion {
		t.Errorf("expected version %q, got %q", testVersion, resp.Version)
	}
}

// TestHandleVersion_DifferentVersionsReflected verifies that different version strings
// passed at construction time appear correctly in the response.
func TestHandleVersion_DifferentVersionsReflected(t *testing.T) {
	cases := []struct {
		name    string
		version string
	}{
		{"semantic version", "1.0.0"},
		{"pre-release", "0.1.0"},
		{"patch version", "3.14.1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := New(Config{Version: tc.version})
			handler := s.handleVersion()

			req := httptest.NewRequest(http.MethodGet, "/version", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rr.Code)
			}

			var resp versionResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			if resp.Version != tc.version {
				t.Errorf("expected version %q, got %q", tc.version, resp.Version)
			}
		})
	}
}

// TestHandleVersion_ResponseBodyIsValidJSON verifies that the response body is valid JSON.
func TestHandleVersion_ResponseBodyIsValidJSON(t *testing.T) {
	s := New(Config{Version: "0.1.0"})
	handler := s.handleVersion()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var raw json.RawMessage
	if err := json.NewDecoder(rr.Body).Decode(&raw); err != nil {
		t.Errorf("response body is not valid JSON: %v", err)
	}
}

// TestHandleVersion_VersionNotImportedFromPackage verifies that the server uses
// the version string from Config rather than importing the version package directly.
// This tests the decoupling architecture: any string passed in Config.Version
// should appear verbatim in the response.
func TestHandleVersion_VersionDecoupledFromPackage(t *testing.T) {
	const customVersion = "custom-build-42"
	s := New(Config{Version: customVersion})
	handler := s.handleVersion()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var resp versionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.Version != customVersion {
		t.Errorf("handler must use version from Config, expected %q, got %q", customVersion, resp.Version)
	}
}
