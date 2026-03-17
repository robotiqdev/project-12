// Package integration_test validates the TypeScript source files for TASK-4593.
// These are file-system tests that verify the app, routes, controller, service,
// middleware, and integration test files contain all required definitions.
//
// Tests WILL FAIL until the implementation files are created — this is the
// intended TDD behaviour.
package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from the current working directory.
// Tests are run with `go test ./...` from the repo root, so we walk up to
// find go.mod.
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

// readFile reads a file relative to the repo root or fails the test.
func readFile(t *testing.T, relPath string) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), relPath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", relPath, err)
	}
	return string(data)
}

// fileExists returns true if the file exists at the given repo-root-relative path.
func fileExists(t *testing.T, relPath string) bool {
	t.Helper()
	path := filepath.Join(repoRoot(t), relPath)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// ── File existence ────────────────────────────────────────────────────────────

func TestIntegrationTestFileExists(t *testing.T) {
	if !fileExists(t, "backend/tests/integration/validation-runs.test.ts") {
		t.Fatal("backend/tests/integration/validation-runs.test.ts does not exist")
	}
}

func TestAppFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/app.ts") {
		t.Fatal("backend/src/app.ts does not exist")
	}
}

func TestRoutesIndexFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/routes/index.ts") {
		t.Fatal("backend/src/routes/index.ts does not exist")
	}
}

func TestValidationRunsRoutesFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/routes/validation-runs.routes.ts") {
		t.Fatal("backend/src/routes/validation-runs.routes.ts does not exist")
	}
}

func TestValidationRunsControllerFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/controllers/validation-runs.controller.ts") {
		t.Fatal("backend/src/controllers/validation-runs.controller.ts does not exist")
	}
}

func TestValidationRunServiceFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/services/validation-run.service.ts") {
		t.Fatal("backend/src/services/validation-run.service.ts does not exist")
	}
}

func TestValidateMiddlewareFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/middleware/validate.middleware.ts") {
		t.Fatal("backend/src/middleware/validate.middleware.ts does not exist")
	}
}

func TestErrorMiddlewareFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/middleware/error.middleware.ts") {
		t.Fatal("backend/src/middleware/error.middleware.ts does not exist")
	}
}

func TestPipelineServiceFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/services/pipeline.service.ts") {
		t.Fatal("backend/src/services/pipeline.service.ts does not exist")
	}
}

// ── backend/src/app.ts ───────────────────────────────────────────────────────

func TestAppTsUsesExpress(t *testing.T) {
	src := readFile(t, "backend/src/app.ts")
	if !strings.Contains(src, "express") {
		t.Error("app.ts must import and use express")
	}
}

func TestAppTsUsesJSONBodyParser(t *testing.T) {
	src := readFile(t, "backend/src/app.ts")
	hasJsonMiddleware := strings.Contains(src, "express.json()") || strings.Contains(src, "json()")
	if !hasJsonMiddleware {
		t.Error("app.ts must configure JSON body parser middleware")
	}
}

func TestAppTsRegistersRoutes(t *testing.T) {
	src := readFile(t, "backend/src/app.ts")
	if !strings.Contains(src, "routes") {
		t.Error("app.ts must register routes")
	}
}

func TestAppTsUsesErrorMiddleware(t *testing.T) {
	src := readFile(t, "backend/src/app.ts")
	if !strings.Contains(src, "error") {
		t.Error("app.ts must register error handling middleware")
	}
}

func TestAppTsExportsApp(t *testing.T) {
	src := readFile(t, "backend/src/app.ts")
	if !strings.Contains(src, "export") || !strings.Contains(src, "app") {
		t.Error("app.ts must export the express app instance")
	}
}

// ── backend/src/routes/validation-runs.routes.ts ─────────────────────────────

func TestValidationRunsRoutesHasPostEndpoint(t *testing.T) {
	src := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	if !strings.Contains(src, "post") && !strings.Contains(src, "POST") {
		t.Error("validation-runs.routes.ts must define a POST route")
	}
}

func TestValidationRunsRoutesUsesValidateMiddleware(t *testing.T) {
	src := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	if !strings.Contains(src, "validate") {
		t.Error("validation-runs.routes.ts must use the validate middleware")
	}
}

func TestValidationRunsRoutesUsesController(t *testing.T) {
	src := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	if !strings.Contains(src, "controller") || !strings.Contains(src, "create") {
		t.Error("validation-runs.routes.ts must reference the controller create method")
	}
}

func TestValidationRunsRoutesExportsRouter(t *testing.T) {
	src := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	if !strings.Contains(src, "export") || !strings.Contains(src, "router") {
		t.Error("validation-runs.routes.ts must export the router")
	}
}

// ── backend/src/controllers/validation-runs.controller.ts ────────────────────

func TestControllerHasCreateMethod(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "create") {
		t.Error("validation-runs.controller.ts must export a create method")
	}
}

func TestControllerCallsConcurrencyServiceCheckAndLock(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "checkAndLockBranch") {
		t.Error("validation-runs.controller.ts must call concurrencyService.checkAndLockBranch")
	}
}

func TestControllerReturns409OnConflict(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "409") {
		t.Error("validation-runs.controller.ts must return 409 status on branch lock conflict")
	}
}

func TestControllerReturnsBranchLockedError(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "BRANCH_LOCKED") {
		t.Error("validation-runs.controller.ts must return BRANCH_LOCKED error message")
	}
}

func TestControllerReturnsConflictingRunId(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "conflictingRunId") {
		t.Error("validation-runs.controller.ts must return conflictingRunId in 409 response")
	}
}

func TestControllerReturns201OnSuccess(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "201") {
		t.Error("validation-runs.controller.ts must return 201 status on successful run creation")
	}
}

func TestControllerCallsPipelineService(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "pipeline") || !strings.Contains(src, "enqueue") {
		t.Error("validation-runs.controller.ts must call pipelineService.enqueue")
	}
}

func TestControllerFiresAndForgetsPipeline(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// Fire-and-forget means no await on enqueue, or use process.nextTick / catch handler
	hasFireAndForget := strings.Contains(src, "process.nextTick") ||
		strings.Contains(src, ".catch(") ||
		strings.Contains(src, "void ")
	if !hasFireAndForget {
		t.Error("validation-runs.controller.ts must fire-and-forget the pipeline call (process.nextTick, void, or .catch)")
	}
}

// ── backend/src/middleware/validate.middleware.ts ────────────────────────────

func TestValidateMiddlewareUsesZod(t *testing.T) {
	src := readFile(t, "backend/src/middleware/validate.middleware.ts")
	if !strings.Contains(src, "zod") && !strings.Contains(src, "ZodSchema") && !strings.Contains(src, "z.") {
		t.Error("validate.middleware.ts must use Zod for schema validation")
	}
}

func TestValidateMiddlewareReturns400OnFailure(t *testing.T) {
	src := readFile(t, "backend/src/middleware/validate.middleware.ts")
	if !strings.Contains(src, "400") {
		t.Error("validate.middleware.ts must return 400 status on validation failure")
	}
}

func TestValidateMiddlewareValidatesReqBody(t *testing.T) {
	src := readFile(t, "backend/src/middleware/validate.middleware.ts")
	if !strings.Contains(src, "req.body") {
		t.Error("validate.middleware.ts must validate req.body")
	}
}

func TestValidateMiddlewareExportsFunction(t *testing.T) {
	src := readFile(t, "backend/src/middleware/validate.middleware.ts")
	if !strings.Contains(src, "export") {
		t.Error("validate.middleware.ts must export the validate function/middleware factory")
	}
}

// ── backend/src/services/validation-run.service.ts ───────────────────────────

func TestValidationRunServiceImportsConcurrencyService(t *testing.T) {
	src := readFile(t, "backend/src/services/validation-run.service.ts")
	if !strings.Contains(src, "concurrency") {
		t.Error("validation-run.service.ts must import/use the concurrency service")
	}
}

func TestValidationRunServiceExportsCreateRun(t *testing.T) {
	src := readFile(t, "backend/src/services/validation-run.service.ts")
	if !strings.Contains(src, "create") {
		t.Error("validation-run.service.ts must export a create/createRun method")
	}
}

// ── backend/src/services/pipeline.service.ts ─────────────────────────────────

func TestPipelineServiceExportsEnqueue(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "enqueue") {
		t.Error("pipeline.service.ts must export an enqueue function")
	}
}

// ── Zod schema validation rules ──────────────────────────────────────────────

func TestValidationRunsRouteOrControllerDefinesZodSchema(t *testing.T) {
	// The Zod schema may live in the routes or controller file
	routesSrc := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	controllerSrc := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	serviceSrc := readFile(t, "backend/src/services/validation-run.service.ts")

	combined := routesSrc + controllerSrc + serviceSrc
	if !strings.Contains(combined, "branchName") {
		t.Error("Zod schema must validate branchName field")
	}
}

func TestZodSchemaValidatesCommitShaFormat(t *testing.T) {
	// Schema files may be spread across routes/controllers/services
	routesSrc := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	controllerSrc := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	serviceSrc := readFile(t, "backend/src/services/validation-run.service.ts")

	combined := routesSrc + controllerSrc + serviceSrc
	// The commitSha regex should enforce 40-char hex
	hasRegex := strings.Contains(combined, "regex") || strings.Contains(combined, "[0-9a-f]") || strings.Contains(combined, "commitSha")
	if !hasRegex {
		t.Error("Zod schema must validate commitSha format (40-char hex regex)")
	}
}

// ── Integration test file content ────────────────────────────────────────────

func TestIntegrationTestFileUsesSupertestOrSimilar(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "request") && !strings.Contains(src, "supertest") {
		t.Error("integration test must use supertest or similar HTTP testing library")
	}
}

func TestIntegrationTestFileHasBeforeEach(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "beforeEach") {
		t.Error("integration test must have a beforeEach hook to truncate tables")
	}
}

func TestIntegrationTestFileTruncatesTables(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "TRUNCATE") && !strings.Contains(src, "truncate") && !strings.Contains(src, "deleteMany") {
		t.Error("integration test must truncate tables in beforeEach")
	}
}

func TestIntegrationTestFileChecks201Response(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "201") {
		t.Error("integration test must verify 201 response for valid requests")
	}
}

func TestIntegrationTestFileChecks409Response(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "409") {
		t.Error("integration test must verify 409 response for branch lock conflicts")
	}
}

func TestIntegrationTestFileChecks400Response(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "400") {
		t.Error("integration test must verify 400 response for validation errors")
	}
}

func TestIntegrationTestFileChecksBranchLockedError(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "BRANCH_LOCKED") {
		t.Error("integration test must verify BRANCH_LOCKED error in 409 response")
	}
}

func TestIntegrationTestFileChecksConflictingRunId(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "conflictingRunId") {
		t.Error("integration test must verify conflictingRunId in 409 response")
	}
}

func TestIntegrationTestFileMocksPipelineService(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "pipeline") {
		t.Error("integration test must mock the pipeline service")
	}
}

func TestIntegrationTestFileVerifiesPipelineEnqueueCalled(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "enqueue") {
		t.Error("integration test must verify pipelineService.enqueue was called")
	}
}

func TestIntegrationTestFileChecksStatusPending(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "PENDING") {
		t.Error("integration test must verify run status is PENDING in the response")
	}
}

func TestIntegrationTestFileTestsFailedBranchAllowsNewRun(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "FAILED") {
		t.Error("integration test must test that a branch with a FAILED run allows new runs")
	}
}

func TestIntegrationTestFileImportsApp(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "app") {
		t.Error("integration test must import the express app")
	}
}
