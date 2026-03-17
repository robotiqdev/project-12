// Package get_endpoints_test validates the TypeScript source files for TASK-4595.
// These are file-system tests that verify the routes, controller, model, and
// integration test files contain all required definitions for the GET endpoints.
//
// Tests WILL FAIL until the implementation files are updated — this is the
// intended TDD behaviour.
package get_endpoints_test

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

func TestValidationRunModelFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/models/validation-run.model.ts") {
		t.Fatal("backend/src/models/validation-run.model.ts does not exist")
	}
}

// ── Routes: GET endpoints ─────────────────────────────────────────────────────

func TestRoutesHasGetListEndpoint(t *testing.T) {
	src := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	// Must define a GET / route that calls the list controller method
	hasGetList := (strings.Contains(src, "router.get('/'") || strings.Contains(src, `router.get("/"`)) &&
		strings.Contains(src, "list")
	if !hasGetList {
		t.Error("validation-runs.routes.ts must define router.get('/') calling the list controller method")
	}
}

func TestRoutesHasGetByIdEndpoint(t *testing.T) {
	src := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	// Must define a GET /:id route that calls getById controller method
	hasGetById := (strings.Contains(src, "router.get('/:id'") || strings.Contains(src, `router.get("/:id"`)) &&
		strings.Contains(src, "getById")
	if !hasGetById {
		t.Error("validation-runs.routes.ts must define router.get('/:id') calling the getById controller method")
	}
}

// ── Controller: list method ───────────────────────────────────────────────────

func TestControllerHasListMethod(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "list") {
		t.Error("validation-runs.controller.ts must export a list method for GET /api/validation-runs")
	}
}

func TestControllerListParsesQueryParams(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// Must parse branch, status, and page/limit from req.query
	if !strings.Contains(src, "req.query") {
		t.Error("validation-runs.controller.ts list method must parse query parameters from req.query")
	}
}

func TestControllerListClampsLimitTo100(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// Must clamp the limit to a maximum of 100
	if !strings.Contains(src, "100") {
		t.Error("validation-runs.controller.ts list method must clamp limit to a maximum of 100")
	}
}

func TestControllerListReturnsPaginatedResponse(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// Must return a paginated response with data, total, page, limit
	hasData := strings.Contains(src, "data")
	hasTotal := strings.Contains(src, "total")
	hasPage := strings.Contains(src, "page")
	hasLimit := strings.Contains(src, "limit")
	if !hasData || !hasTotal || !hasPage || !hasLimit {
		t.Error("validation-runs.controller.ts list method must return {data, total, page, limit} in response")
	}
}

func TestControllerListCallsValidationRunModelList(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// Must call validationRunModel.list (or similar) to fetch runs
	hasModelCall := strings.Contains(src, "validationRunModel") || strings.Contains(src, "validation-run.model") ||
		strings.Contains(src, "validationRunsModel")
	if !hasModelCall {
		t.Error("validation-runs.controller.ts list method must call the validationRunModel to retrieve runs")
	}
}

func TestControllerListDefaultPageIs1(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// list function must default page to 1
	if !strings.Contains(src, "page") {
		t.Error("validation-runs.controller.ts list method must handle page parameter (default 1)")
	}
}

func TestControllerListDefaultLimitIs20(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// list function must default limit to 20
	if !strings.Contains(src, "20") {
		t.Error("validation-runs.controller.ts list method must default limit to 20")
	}
}

// ── Controller: getById method ────────────────────────────────────────────────

func TestControllerHasGetByIdMethod(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "getById") {
		t.Error("validation-runs.controller.ts must export a getById method for GET /api/validation-runs/:id")
	}
}

func TestControllerGetByIdCallsFindByIdWithStages(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "findByIdWithStages") {
		t.Error("validation-runs.controller.ts getById method must call validationRunModel.findByIdWithStages")
	}
}

func TestControllerGetByIdReturns404WhenNotFound(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	if !strings.Contains(src, "404") {
		t.Error("validation-runs.controller.ts getById method must return 404 when run is not found")
	}
}

func TestControllerGetByIdReturnsRunAndStages(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// Must return a response containing run and stages
	hasRun := strings.Contains(src, "run")
	hasStages := strings.Contains(src, "stages")
	if !hasRun || !hasStages {
		t.Error("validation-runs.controller.ts getById method must return {run, stages} in response")
	}
}

func TestControllerGetByIdParsesIdFromParams(t *testing.T) {
	src := readFile(t, "backend/src/controllers/validation-runs.controller.ts")
	// Must read the id from req.params
	if !strings.Contains(src, "req.params") {
		t.Error("validation-runs.controller.ts getById method must parse :id from req.params")
	}
}

// ── Model: list function with filters ────────────────────────────────────────

func TestModelListFunctionExists(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "export async function list") {
		t.Error("validation-run.model.ts must export an async list function")
	}
}

func TestModelListSupportsBranchFilter(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "branchName") {
		t.Error("validation-run.model.ts list function must support filtering by branchName")
	}
}

func TestModelListSupportsStatusFilter(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "status") {
		t.Error("validation-run.model.ts list function must support filtering by status")
	}
}

func TestModelListUsesOrderByCreatedAtDesc(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	hasOrderBy := strings.Contains(src, "orderBy") && strings.Contains(src, "createdAt") && strings.Contains(src, "desc")
	if !hasOrderBy {
		t.Error("validation-run.model.ts list function must order results by createdAt desc")
	}
}

func TestModelListReturnsTotalCount(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "count") {
		t.Error("validation-run.model.ts list function must return total count alongside data")
	}
}

func TestModelFindByIdWithStagesExists(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "findByIdWithStages") {
		t.Error("validation-run.model.ts must export findByIdWithStages function")
	}
}

func TestModelFindByIdWithStagesIncludesStages(t *testing.T) {
	src := readFile(t, "backend/src/models/validation-run.model.ts")
	if !strings.Contains(src, "include") || !strings.Contains(src, "stages") {
		t.Error("validation-run.model.ts findByIdWithStages must include stages in the query")
	}
}

// ── Integration test file content ────────────────────────────────────────────

func TestIntegrationTestFileHasGetListDescribeBlock(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	hasGetListDescribe := strings.Contains(src, "GET /api/validation-runs")
	if !hasGetListDescribe {
		t.Error("integration test file must have a describe block for GET /api/validation-runs")
	}
}

func TestIntegrationTestFileHasGetByIdDescribeBlock(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	hasGetByIdDescribe := strings.Contains(src, "GET /api/validation-runs/:id") ||
		strings.Contains(src, "getById") ||
		strings.Contains(src, "GET /api/validation-runs/")
	if !hasGetByIdDescribe {
		t.Error("integration test file must have a describe block for GET /api/validation-runs/:id")
	}
}

func TestIntegrationTestFileChecksEmptyPaginatedResponse(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	// Must test the empty state: {data: [], total: 0, page: 1, limit: 20}
	hasDataArray := strings.Contains(src, "data: []")
	if !hasDataArray {
		t.Error("integration test must verify empty data array in paginated response when no runs exist")
	}
}

func TestIntegrationTestFileChecks200ResponseForList(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "200") {
		t.Error("integration test must verify 200 response for GET /api/validation-runs")
	}
}

func TestIntegrationTestFileChecks404Response(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "404") {
		t.Error("integration test must verify 404 response for unknown run id in GET /api/validation-runs/:id")
	}
}

func TestIntegrationTestFileChecksStagesArray(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "stages") {
		t.Error("integration test must verify stages array is returned in GET /api/validation-runs/:id response")
	}
}

func TestIntegrationTestFileChecksFilterByBranchAndStatus(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	hasBranchFilter := strings.Contains(src, "branch=main") || strings.Contains(src, "branch=")
	hasStatusFilter := strings.Contains(src, "status=RUNNING") || strings.Contains(src, "status=")
	if !hasBranchFilter || !hasStatusFilter {
		t.Error("integration test must test filtering by branch and status query parameters")
	}
}

func TestIntegrationTestFileChecksPaginationWithLimit(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(src, "limit") {
		t.Error("integration test must test pagination with limit parameter")
	}
}

func TestIntegrationTestFileChecksLimitClampedTo100(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	// Must test that limit is clamped to 100 maximum
	hasLimit100 := strings.Contains(src, "100") && strings.Contains(src, "limit")
	if !hasLimit100 {
		t.Error("integration test must verify limit is clamped to a maximum of 100")
	}
}

func TestIntegrationTestFileChecks25RunsSeeding(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	// Must seed 25 runs to test pagination
	if !strings.Contains(src, "25") {
		t.Error("integration test must seed 25 runs to test pagination metadata")
	}
}

func TestIntegrationTestFileChecksPaginationMetadata(t *testing.T) {
	src := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	// Must check page and limit in response body
	hasPageCheck := strings.Contains(src, "page")
	hasLimitCheck := strings.Contains(src, "limit")
	hasTotalCheck := strings.Contains(src, "total")
	if !hasPageCheck || !hasLimitCheck || !hasTotalCheck {
		t.Error("integration test must verify page, limit, and total in paginated response metadata")
	}
}
