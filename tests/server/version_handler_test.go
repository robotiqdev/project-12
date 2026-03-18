// Package server_test validates the server/version_handler.go file for TASK-4630.
// These file-system tests verify that the version_handler.go source file exists and
// contains the expected unexported versionResponse struct with the correct JSON tag.
// Serialization behavior is covered indirectly by TASK-4625 version integration tests
// which decode the JSON response body into this struct shape.
package server_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from the current working directory
// by walking up to find go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

// versionHandlerPath returns the path to server/version_handler.go.
func versionHandlerPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "server", "version_handler.go")
}

// readVersionHandler reads server/version_handler.go and returns its content.
func readVersionHandler(t *testing.T) string {
	t.Helper()
	path := versionHandlerPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read server/version_handler.go: %v", err)
	}
	return string(data)
}

// ---- File existence ----

// TestVersionHandlerFileExists verifies that server/version_handler.go is present.
func TestVersionHandlerFileExists(t *testing.T) {
	path := versionHandlerPath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("server/version_handler.go not found at %s", path)
	}
}

// ---- Package declaration ----

// TestVersionHandlerPackageIsServer verifies the file declares package server.
func TestVersionHandlerPackageIsServer(t *testing.T) {
	src := readVersionHandler(t)
	if !strings.Contains(src, "package server") {
		t.Error("server/version_handler.go must declare `package server`")
	}
}

// ---- Struct definition ----

// TestVersionResponseStructDefined verifies the unexported versionResponse struct exists.
func TestVersionResponseStructDefined(t *testing.T) {
	src := readVersionHandler(t)
	if !strings.Contains(src, "type versionResponse struct") {
		t.Error("server/version_handler.go must define unexported struct `type versionResponse struct`")
	}
}

// TestVersionResponseStructIsUnexported verifies the struct name starts with a lowercase letter,
// confirming it is unexported (package-private) as required.
func TestVersionResponseStructIsUnexported(t *testing.T) {
	src := readVersionHandler(t)
	// The struct must use lowercase 'v', not exported 'VersionResponse'.
	if strings.Contains(src, "type VersionResponse struct") {
		t.Error("versionResponse struct must NOT be exported; rename to `versionResponse` (lowercase v)")
	}
	if !strings.Contains(src, "type versionResponse struct") {
		t.Error("server/version_handler.go must define unexported struct `versionResponse`")
	}
}

// TestVersionResponseVersionField verifies the struct contains a Version string field.
func TestVersionResponseVersionField(t *testing.T) {
	src := readVersionHandler(t)
	if !strings.Contains(src, "Version") {
		t.Error("versionResponse struct must contain a `Version` field")
	}
}

// TestVersionResponseVersionFieldIsString verifies the Version field type is string.
func TestVersionResponseVersionFieldIsString(t *testing.T) {
	src := readVersionHandler(t)
	// The field declaration should be "Version string"
	if !strings.Contains(src, "Version string") {
		t.Error("versionResponse.Version must be of type `string`")
	}
}

// TestVersionResponseVersionJSONTag verifies the Version field carries the json:"version" tag.
func TestVersionResponseVersionJSONTag(t *testing.T) {
	src := readVersionHandler(t)
	if !strings.Contains(src, `json:"version"`) {
		t.Error("versionResponse.Version must have struct tag `json:\"version\"`")
	}
}

// TestVersionResponseVersionFieldComplete verifies the full field declaration including JSON tag.
func TestVersionResponseVersionFieldComplete(t *testing.T) {
	src := readVersionHandler(t)
	// Check for the full field declaration including the struct tag.
	hasTag := strings.Contains(src, "Version string `json:\"version\"`")
	if !hasTag {
		t.Error("versionResponse struct must contain field: Version string `json:\"version\"`")
	}
}
