// Package admin_test validates the admin routes, controller, middleware, and
// integration test files for TASK-4619.
// These are integration-style file-system tests that verify the TypeScript
// source files contain all required patterns for the admin cleanup API.
package admin_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from the current working directory.
// Tests are run with `go test ./...` from the repo root, so the working
// directory is the package directory; we walk up to find go.mod.
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

// ---- Helpers ----

func readAdminRoutes(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "routes", "admin.routes.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read admin.routes.ts: %v", err)
	}
	return string(data)
}

func readAdminController(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "controllers", "admin.controller.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read admin.controller.ts: %v", err)
	}
	return string(data)
}

func readAuthMiddleware(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "src", "middleware", "auth.middleware.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read auth.middleware.ts: %v", err)
	}
	return string(data)
}

func readAdminIntegrationTest(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "backend", "tests", "integration", "admin.test.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read admin.test.ts: %v", err)
	}
	return string(data)
}

// ---- File existence ----

func TestAdminRoutesFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "routes", "admin.routes.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("admin.routes.ts not found at %s", path)
	}
}

func TestAdminControllerFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "controllers", "admin.controller.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("admin.controller.ts not found at %s", path)
	}
}

func TestAuthMiddlewareFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "middleware", "auth.middleware.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("auth.middleware.ts not found at %s", path)
	}
}

func TestAdminIntegrationTestFileExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "tests", "integration", "admin.test.ts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("admin.test.ts not found at %s", path)
	}
}

// ---- admin.routes.ts — route definitions ----

func TestAdminRoutesDefinesPostCleanupAge(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "cleanup/age") {
		t.Error("admin.routes.ts must define POST /cleanup/age route")
	}
}

func TestAdminRoutesDefinesPostCleanupCount(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "cleanup/count") {
		t.Error("admin.routes.ts must define POST /cleanup/count route")
	}
}

func TestAdminRoutesDefinesGetCleanupJobs(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "cleanup/jobs") {
		t.Error("admin.routes.ts must define GET /cleanup/jobs route")
	}
}

func TestAdminRoutesUsesAdminAuthMiddleware(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "adminAuthMiddleware") {
		t.Error("admin.routes.ts must use adminAuthMiddleware to protect all routes")
	}
}

func TestAdminRoutesImportsAdminController(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "admin.controller") && !strings.Contains(routes, "adminController") {
		t.Error("admin.routes.ts must import from admin.controller")
	}
}

func TestAdminRoutesImportsAuthMiddleware(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "auth.middleware") && !strings.Contains(routes, "authMiddleware") {
		t.Error("admin.routes.ts must import from auth.middleware")
	}
}

func TestAdminRoutesExportsRouter(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "export") {
		t.Error("admin.routes.ts must export the router")
	}
}

// ---- admin.routes.ts — HTTP methods ----

func TestAdminRoutesCleanupAgeIsPost(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "post") && !strings.Contains(routes, "POST") {
		t.Error("admin.routes.ts must use POST method for cleanup/age route")
	}
}

func TestAdminRoutesCleanupCountIsPost(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "post") && !strings.Contains(routes, "POST") {
		t.Error("admin.routes.ts must use POST method for cleanup/count route")
	}
}

func TestAdminRoutesCleanupJobsIsGet(t *testing.T) {
	routes := readAdminRoutes(t)
	if !strings.Contains(routes, "get") && !strings.Contains(routes, "GET") {
		t.Error("admin.routes.ts must use GET method for cleanup/jobs route")
	}
}

// ---- admin.controller.ts — controller handlers ----

func TestAdminControllerExportsRunAgeBasedCleanup(t *testing.T) {
	controller := readAdminController(t)
	if !strings.Contains(controller, "runAgeBased") {
		t.Error("admin.controller.ts must call or export runAgeBased cleanup handler")
	}
}

func TestAdminControllerExportsRunCountBasedCleanup(t *testing.T) {
	controller := readAdminController(t)
	if !strings.Contains(controller, "runCountBased") {
		t.Error("admin.controller.ts must call or export runCountBased cleanup handler")
	}
}

func TestAdminControllerExportsGetCleanupJobs(t *testing.T) {
	controller := readAdminController(t)
	if !strings.Contains(controller, "jobs") || !strings.Contains(controller, "list") {
		t.Error("admin.controller.ts must export a handler that lists cleanup jobs")
	}
}

func TestAdminControllerImportsCleanupService(t *testing.T) {
	controller := readAdminController(t)
	if !strings.Contains(controller, "cleanup.service") && !strings.Contains(controller, "cleanupService") {
		t.Error("admin.controller.ts must import from cleanup.service")
	}
}

func TestAdminControllerImportsCleanupJobModel(t *testing.T) {
	controller := readAdminController(t)
	if !strings.Contains(controller, "cleanup-job.model") && !strings.Contains(controller, "cleanupJobModel") {
		t.Error("admin.controller.ts must import from cleanup-job.model")
	}
}

func TestAdminControllerReturnsCleanupResult(t *testing.T) {
	controller := readAdminController(t)
	if !strings.Contains(controller, "CleanupResult") && !strings.Contains(controller, "result") {
		t.Error("admin.controller.ts cleanup handlers must return a CleanupResult")
	}
}

func TestAdminControllerReturns200OnSuccess(t *testing.T) {
	controller := readAdminController(t)
	if !strings.Contains(controller, "200") && !strings.Contains(controller, "json") {
		t.Error("admin.controller.ts must return 200 status with JSON response on success")
	}
}

func TestAdminControllerExportsHandlers(t *testing.T) {
	controller := readAdminController(t)
	if !strings.Contains(controller, "export") {
		t.Error("admin.controller.ts must export its handler functions")
	}
}

// ---- auth.middleware.ts — admin auth middleware ----

func TestAuthMiddlewareExportsAdminAuthMiddleware(t *testing.T) {
	middleware := readAuthMiddleware(t)
	if !strings.Contains(middleware, "adminAuthMiddleware") {
		t.Error("auth.middleware.ts must export adminAuthMiddleware function")
	}
}

func TestAuthMiddlewareChecksXAdminTokenHeader(t *testing.T) {
	middleware := readAuthMiddleware(t)
	if !strings.Contains(middleware, "X-Admin-Token") {
		t.Error("auth.middleware.ts must check X-Admin-Token header for admin authentication")
	}
}

func TestAuthMiddlewareUsesAdminTokenEnvVar(t *testing.T) {
	middleware := readAuthMiddleware(t)
	if !strings.Contains(middleware, "ADMIN_TOKEN") {
		t.Error("auth.middleware.ts must read ADMIN_TOKEN from environment variables")
	}
}

func TestAuthMiddlewareReturns403OnUnauthorized(t *testing.T) {
	middleware := readAuthMiddleware(t)
	if !strings.Contains(middleware, "403") {
		t.Error("auth.middleware.ts must return 403 status when admin token is invalid or missing")
	}
}

func TestAuthMiddlewareCallsNextOnAuthorized(t *testing.T) {
	middleware := readAuthMiddleware(t)
	if !strings.Contains(middleware, "next") {
		t.Error("auth.middleware.ts must call next() when admin token is valid")
	}
}

func TestAuthMiddlewareExportsFunction(t *testing.T) {
	middleware := readAuthMiddleware(t)
	if !strings.Contains(middleware, "export") {
		t.Error("auth.middleware.ts must export its middleware functions")
	}
}

// ---- Integration test file patterns ----

func TestAdminTestFileTestsPostCleanupAge(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "/api/admin/cleanup/age") {
		t.Error("admin.test.ts must contain tests for POST /api/admin/cleanup/age")
	}
}

func TestAdminTestFileTestsPostCleanupCount(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "/api/admin/cleanup/count") {
		t.Error("admin.test.ts must contain tests for POST /api/admin/cleanup/count")
	}
}

func TestAdminTestFileTestsGetCleanupJobs(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "/api/admin/cleanup/jobs") {
		t.Error("admin.test.ts must contain tests for GET /api/admin/cleanup/jobs")
	}
}

func TestAdminTestFileTests200ResponseForAgeCleanup(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "200") {
		t.Error("admin.test.ts must verify that cleanup endpoints return 200 on success")
	}
}

func TestAdminTestFileTests403ForNonAdmin(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "403") {
		t.Error("admin.test.ts must test that non-admin requests return 403")
	}
}

func TestAdminTestFileVerifiesCleanupResultShape(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "CleanupResult") && !strings.Contains(testFile, "deletedCount") {
		t.Error("admin.test.ts must verify CleanupResult shape in response body")
	}
}

func TestAdminTestFileMocksCleanupService(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "jest.mock") && !strings.Contains(testFile, "vi.mock") {
		t.Error("admin.test.ts must mock cleanup service dependencies")
	}
}

func TestAdminTestFileTestsAdminTokenValidation(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "X-Admin-Token") {
		t.Error("admin.test.ts must test X-Admin-Token header validation")
	}
}

func TestAdminTestFileTestsJobsList(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "list") || !strings.Contains(testFile, "jobs") {
		t.Error("admin.test.ts must test that cleanup jobs list endpoint returns job records")
	}
}

func TestAdminTestFileTestsAgeBasedType(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "AGE_BASED") {
		t.Error("admin.test.ts must verify AGE_BASED type in cleanup/age responses")
	}
}

func TestAdminTestFileTestsCountBasedType(t *testing.T) {
	testFile := readAdminIntegrationTest(t)
	if !strings.Contains(testFile, "COUNT_BASED") {
		t.Error("admin.test.ts must verify COUNT_BASED type in cleanup/count responses")
	}
}

// ---- Routes index — admin router registration ----

func TestRoutesIndexRegistersAdminRouter(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "routes", "index.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("routes/index.ts not found at %s — admin router must be registered there", path)
	}
	content := string(data)
	if !strings.Contains(content, "admin") {
		t.Error("routes/index.ts must register the admin router")
	}
}

func TestRoutesIndexMountsAdminAtApiAdmin(t *testing.T) {
	path := filepath.Join(repoRoot(t), "backend", "src", "routes", "index.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("routes/index.ts not found at %s", path)
	}
	content := string(data)
	if !strings.Contains(content, "/api/admin") {
		t.Error("routes/index.ts must mount the admin router at /api/admin")
	}
}
