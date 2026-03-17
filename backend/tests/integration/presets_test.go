// Package integration_test validates the Preset CRUD API files for TASK-4602.
// These are file-system tests that verify the routes, controller, and integration
// test TypeScript files exist and contain all required definitions.
//
// Tests for routes/controller SHOULD FAIL until the implementation is created.
package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- Helpers ----

func readPresetsRoutesTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "routes", "presets.routes.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read presets.routes.ts: %v", err)
	}
	return string(data)
}

func readPresetsControllerTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "controllers", "presets.controller.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read presets.controller.ts: %v", err)
	}
	return string(data)
}

func readPresetsIntegrationTestTS(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "integration", "presets.test.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read presets.test.ts: %v", err)
	}
	return string(data)
}

// ---- File existence tests ----

func TestPresetsRoutesFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "routes", "presets.routes.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("presets.routes.ts not found at %s", path)
	}
}

func TestPresetsControllerFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "controllers", "presets.controller.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("presets.controller.ts not found at %s", path)
	}
}

func TestPresetsIntegrationTestFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "tests", "integration", "presets.test.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("presets.test.ts not found at %s", path)
	}
}

// ---- presets.routes.ts content ----

func TestPresetsRoutesDefinesGetRoute(t *testing.T) {
	routes := readPresetsRoutesTS(t)
	lower := strings.ToLower(routes)
	if !strings.Contains(lower, ".get(") && !strings.Contains(lower, "router.get") {
		t.Error("presets.routes.ts must define a GET route for listing presets")
	}
}

func TestPresetsRoutesDefinesPostRoute(t *testing.T) {
	routes := readPresetsRoutesTS(t)
	lower := strings.ToLower(routes)
	if !strings.Contains(lower, ".post(") && !strings.Contains(lower, "router.post") {
		t.Error("presets.routes.ts must define a POST route for creating presets")
	}
}

func TestPresetsRoutesDefinesPutRoute(t *testing.T) {
	routes := readPresetsRoutesTS(t)
	lower := strings.ToLower(routes)
	if !strings.Contains(lower, ".put(") && !strings.Contains(lower, "router.put") {
		t.Error("presets.routes.ts must define a PUT route for updating presets")
	}
}

func TestPresetsRoutesDefinesDeleteRoute(t *testing.T) {
	routes := readPresetsRoutesTS(t)
	lower := strings.ToLower(routes)
	if !strings.Contains(lower, ".delete(") && !strings.Contains(lower, "router.delete") {
		t.Error("presets.routes.ts must define a DELETE route for removing presets")
	}
}

func TestPresetsRoutesUsesIdParam(t *testing.T) {
	routes := readPresetsRoutesTS(t)
	if !strings.Contains(routes, ":id") {
		t.Error("presets.routes.ts must include :id parameter for single-preset routes (PUT, DELETE)")
	}
}

func TestPresetsRoutesImportsRouter(t *testing.T) {
	routes := readPresetsRoutesTS(t)
	if !strings.Contains(routes, "Router") {
		t.Error("presets.routes.ts must use Express Router")
	}
}

func TestPresetsRoutesImportsController(t *testing.T) {
	routes := readPresetsRoutesTS(t)
	lower := strings.ToLower(routes)
	if !strings.Contains(lower, "controller") {
		t.Error("presets.routes.ts must import from the presets controller")
	}
}

func TestPresetsRoutesScopedToApiPresets(t *testing.T) {
	routes := readPresetsRoutesTS(t)
	lower := strings.ToLower(routes)
	if !strings.Contains(lower, "presets") {
		t.Error("presets.routes.ts must be scoped under /api/presets path")
	}
}

// ---- presets.controller.ts content ----

func TestPresetsControllerExportsGetPresets(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "getPresets") {
		t.Error("presets.controller.ts must export a getPresets handler")
	}
}

func TestPresetsControllerExportsCreatePreset(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "createPreset") {
		t.Error("presets.controller.ts must export a createPreset handler")
	}
}

func TestPresetsControllerExportsUpdatePreset(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "updatePreset") {
		t.Error("presets.controller.ts must export an updatePreset handler")
	}
}

func TestPresetsControllerExportsDeletePreset(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "deletePreset") {
		t.Error("presets.controller.ts must export a deletePreset handler")
	}
}

func TestPresetsControllerExtractsUserId(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "userId") {
		t.Error("presets.controller.ts must extract userId from the request context (auth middleware)")
	}
}

func TestPresetsControllerReturns201OnCreate(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "201") {
		t.Error("presets.controller.ts must return 201 status on successful preset creation")
	}
}

func TestPresetsControllerReturns204OnDelete(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "204") {
		t.Error("presets.controller.ts must return 204 status on successful preset deletion")
	}
}

func TestPresetsControllerHandlesPresetLimitExceededError(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "422") && !strings.Contains(ctrl, "statusCode") && !strings.Contains(ctrl, "PresetLimitExceededError") {
		t.Error("presets.controller.ts must handle PresetLimitExceededError and map it to 422 status")
	}
}

func TestPresetsControllerMapsPresetLimitErrorToCode(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "PRESET_LIMIT_EXCEEDED") {
		t.Error("presets.controller.ts must return PRESET_LIMIT_EXCEEDED error code in response")
	}
}

func TestPresetsControllerUsesPresetService(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "presetService") {
		t.Error("presets.controller.ts must use the presetService singleton")
	}
}

func TestPresetsControllerUsesZodValidation(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	lower := strings.ToLower(ctrl)
	if !strings.Contains(lower, "zod") && !strings.Contains(lower, "safeParse") && !strings.Contains(lower, "z.object") {
		t.Error("presets.controller.ts must use zod validation for request bodies")
	}
}

func TestPresetsControllerValidatesNameRequired(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "name") {
		t.Error("presets.controller.ts zod schema must require the name field")
	}
}

func TestPresetsControllerValidatesNameMaxLength(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "100") {
		t.Error("presets.controller.ts zod schema must enforce max 100 chars on name")
	}
}

func TestPresetsControllerHandles403Or404ForOtherUser(t *testing.T) {
	ctrl := readPresetsControllerTS(t)
	if !strings.Contains(ctrl, "403") && !strings.Contains(ctrl, "404") {
		t.Error("presets.controller.ts must handle 403/404 errors for cross-user operations")
	}
}

// ---- presets.test.ts integration test content ----

func TestPresetsIntegrationTestMocksPresetService(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "jest.mock") {
		t.Error("presets.test.ts must mock the preset service before importing the app")
	}
}

func TestPresetsIntegrationTestImportsApp(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "app") {
		t.Error("presets.test.ts must import the app for supertest requests")
	}
}

func TestPresetsIntegrationTestContainsGetPresetsEndpoint(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	lower := strings.ToLower(testFile)
	if !strings.Contains(lower, "get /api/presets") && !strings.Contains(testFile, "get('/api/presets") {
		t.Error("presets.test.ts must contain tests for GET /api/presets")
	}
}

func TestPresetsIntegrationTestContainsPostPresetsEndpoint(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	lower := strings.ToLower(testFile)
	if !strings.Contains(lower, "post /api/presets") && !strings.Contains(testFile, "post('/api/presets") {
		t.Error("presets.test.ts must contain tests for POST /api/presets")
	}
}

func TestPresetsIntegrationTestContainsPutPresetsEndpoint(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	lower := strings.ToLower(testFile)
	if !strings.Contains(lower, "put /api/presets") && !strings.Contains(testFile, "put('/api/presets") {
		t.Error("presets.test.ts must contain tests for PUT /api/presets/:id")
	}
}

func TestPresetsIntegrationTestContainsDeletePresetsEndpoint(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	lower := strings.ToLower(testFile)
	if !strings.Contains(lower, "delete /api/presets") && !strings.Contains(testFile, "delete('/api/presets") {
		t.Error("presets.test.ts must contain tests for DELETE /api/presets/:id")
	}
}

func TestPresetsIntegrationTestVerifies200OnGetPresets(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "200") {
		t.Error("presets.test.ts must verify 200 status on GET /api/presets")
	}
}

func TestPresetsIntegrationTestVerifiesEmptyArrayForNewUser(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "empty") && !strings.Contains(testFile, "toHaveLength(0)") && !strings.Contains(testFile, "length).toBe(0)") {
		t.Error("presets.test.ts must verify GET /api/presets returns empty array for a new user")
	}
}

func TestPresetsIntegrationTestVerifies201OnCreate(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "201") {
		t.Error("presets.test.ts must verify 201 status on successful POST /api/presets")
	}
}

func TestPresetsIntegrationTestVerifies422OnLimitExceeded(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "422") {
		t.Error("presets.test.ts must verify 422 status when preset limit is exceeded")
	}
}

func TestPresetsIntegrationTestVerifiesPresetLimitExceededError(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "PRESET_LIMIT_EXCEEDED") {
		t.Error("presets.test.ts must verify the PRESET_LIMIT_EXCEEDED error code in 422 response")
	}
}

func TestPresetsIntegrationTestVerifies403Or404ForOtherUserUpdate(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "403") && !strings.Contains(testFile, "404") {
		t.Error("presets.test.ts must verify 403 or 404 when updating another user's preset")
	}
}

func TestPresetsIntegrationTestVerifies204OnDelete(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "204") {
		t.Error("presets.test.ts must verify 204 status on successful DELETE /api/presets/:id")
	}
}

func TestPresetsIntegrationTestVerifiesZodNameRequired(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "name") {
		t.Error("presets.test.ts must include zod validation tests for required name field")
	}
}

func TestPresetsIntegrationTestVerifiesZodNameMaxLength(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "100") {
		t.Error("presets.test.ts must include zod validation test for name max 100 characters")
	}
}

func TestPresetsIntegrationTestVerifiesPutReturns200(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	// The file should test PUT returning 200
	if !strings.Contains(testFile, "200") {
		t.Error("presets.test.ts must verify 200 status on successful PUT /api/presets/:id")
	}
}

func TestPresetsIntegrationTestContainsPresetIdParam(t *testing.T) {
	testFile := readPresetsIntegrationTestTS(t)
	if !strings.Contains(testFile, "/api/presets/") {
		t.Error("presets.test.ts must test single-preset endpoints using /api/presets/:id pattern")
	}
}
