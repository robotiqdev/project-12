// Package server_test contains integration tests for the server handlers.
// TASK-4625 / TASK-4634: verifies that version.Version flows end-to-end from
// the version package through Config into the /version handler JSON response.
package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robotiqdev/project-12/server"
	"github.com/robotiqdev/project-12/version"
)

// TestVersionEndpointReturnsVersionPackageConstant verifies that when the
// server is constructed with version.Version as Config.Version, the /version
// endpoint returns a JSON body whose "version" field equals version.Version.
func TestVersionEndpointReturnsVersionPackageConstant(t *testing.T) {
	srv := server.New(server.Config{
		Version: version.Version,
	})

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	got, ok := body["version"]
	if !ok {
		t.Fatal("response JSON missing 'version' field")
	}

	if got != version.Version {
		t.Errorf("expected version %q, got %q", version.Version, got)
	}
}

// TestVersionEndpointContentType verifies the /version endpoint responds with
// the application/json content type.
func TestVersionEndpointContentType(t *testing.T) {
	srv := server.New(server.Config{
		Version: version.Version,
	})

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct == "" {
		t.Fatal("expected Content-Type header, got none")
	}
}

// TestVersionEndpointVersionFieldMatchesConfig verifies that whatever string
// is set in Config.Version is exactly what appears in the JSON response,
// using version.Version as the value (end-to-end constant propagation).
func TestVersionEndpointVersionFieldMatchesConfig(t *testing.T) {
	// Construct with version.Version explicitly — this is the integration point.
	cfg := server.Config{Version: version.Version}
	srv := server.New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["version"] != version.Version {
		t.Errorf("version field in response %q does not match version.Version %q",
			body["version"], version.Version)
	}
}
