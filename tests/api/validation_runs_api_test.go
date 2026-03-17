// Package api_test validates the structure and content of the ValidationRun API
// implementation files for TASK-4588.
//
// These file-system tests verify that:
//   1. The TypeScript integration test file for validation-runs API exists.
//   2. A ValidationRun types file exists and declares the required response types:
//      ValidationRunListResponse and ValidationRunDetailResponse.
//   3. A controller file exists and includes JSDoc documentation.
//   4. The Zod validation schema for list query params includes CANCELLED as a
//      valid status value.
//   5. A routes file registers the GET /api/validation-runs and
//      GET /api/validation-runs/:id endpoints.
//   6. The detail query includes stages ordered by stageIndex.
//
// These tests are expected to FAIL until the implementation is written (TDD).
package api_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the current working directory until go.mod is found.
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

// ---- Helpers ----------------------------------------------------------------

// readFile reads a file relative to the repo root and returns its content.
func readFile(t *testing.T, relPath string) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), relPath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

// fileExists returns true if the file at relPath (relative to repo root) exists.
func fileExists(t *testing.T, relPath string) bool {
	t.Helper()
	path := filepath.Join(repoRoot(t), relPath)
	_, err := os.Stat(path)
	return err == nil
}

// ---- TypeScript integration test file ---------------------------------------

func TestValidationRunsIntegrationTestFileExists(t *testing.T) {
	path := filepath.Join(
		repoRoot(t),
		"backend", "tests", "integration", "validation-runs.test.ts",
	)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf(
			"validation-runs.test.ts not found at %s — create the integration test file",
			path,
		)
	}
}

func TestValidationRunsIntegrationTestCoversCancelledFilter(t *testing.T) {
	content := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(content, "CANCELLED") {
		t.Error(
			"validation-runs.test.ts must include a test for filtering by status=CANCELLED",
		)
	}
}

func TestValidationRunsIntegrationTestCoversCommitShaFilter(t *testing.T) {
	content := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(content, "commitSha") {
		t.Error(
			"validation-runs.test.ts must include a test for filtering by commitSha prefix",
		)
	}
}

func TestValidationRunsIntegrationTestCoversCombinedFilter(t *testing.T) {
	content := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	// Combined filter test requires both branch and status params
	if !strings.Contains(content, "branch") || !strings.Contains(content, "status") {
		t.Error(
			"validation-runs.test.ts must include a combined branch+status filter test",
		)
	}
}

func TestValidationRunsIntegrationTestCoversDetailEndpoint(t *testing.T) {
	content := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(content, "stages") {
		t.Error(
			"validation-runs.test.ts must include a test for the detail endpoint " +
				"that verifies the stages array",
		)
	}
}

func TestValidationRunsIntegrationTestCovers404(t *testing.T) {
	content := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(content, "404") {
		t.Error(
			"validation-runs.test.ts must include a test for 404 response on non-existent ID",
		)
	}
}

func TestValidationRunsIntegrationTestCoversLimitCap(t *testing.T) {
	content := readFile(t, "backend/tests/integration/validation-runs.test.ts")
	if !strings.Contains(content, "limit") || !strings.Contains(content, "100") {
		t.Error(
			"validation-runs.test.ts must include a test verifying limit is capped at 100",
		)
	}
}

// ---- ValidationRun types file -----------------------------------------------

func TestValidationRunTypesFileExists(t *testing.T) {
	// The implementation must create a types file named validation-run.types.ts
	// in the backend/src directory tree.
	candidates := []string{
		"backend/src/types/validation-run.types.ts",
		"backend/src/validation-run.types.ts",
		"backend/src/api/validation-runs/validation-run.types.ts",
	}
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			return // found — test passes
		}
	}
	t.Fatalf(
		"validation-run.types.ts not found; expected it at one of: %v — "+
			"create the types file with ValidationRunListResponse and ValidationRunDetailResponse",
		candidates,
	)
}

func TestValidationRunTypesFileHasListResponse(t *testing.T) {
	candidates := []string{
		"backend/src/types/validation-run.types.ts",
		"backend/src/validation-run.types.ts",
		"backend/src/api/validation-runs/validation-run.types.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal(
			"validation-run.types.ts not found — cannot check for ValidationRunListResponse",
		)
	}
	if !strings.Contains(content, "ValidationRunListResponse") {
		t.Error(
			"validation-run.types.ts must export type ValidationRunListResponse",
		)
	}
}

func TestValidationRunTypesFileHasDetailResponse(t *testing.T) {
	candidates := []string{
		"backend/src/types/validation-run.types.ts",
		"backend/src/validation-run.types.ts",
		"backend/src/api/validation-runs/validation-run.types.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal(
			"validation-run.types.ts not found — cannot check for ValidationRunDetailResponse",
		)
	}
	if !strings.Contains(content, "ValidationRunDetailResponse") {
		t.Error(
			"validation-run.types.ts must export type ValidationRunDetailResponse",
		)
	}
}

func TestValidationRunDetailResponseIncludesStagesField(t *testing.T) {
	candidates := []string{
		"backend/src/types/validation-run.types.ts",
		"backend/src/validation-run.types.ts",
		"backend/src/api/validation-runs/validation-run.types.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal(
			"validation-run.types.ts not found — cannot check DetailResponse for stages field",
		)
	}
	if !strings.Contains(content, "stages") {
		t.Error(
			"ValidationRunDetailResponse in validation-run.types.ts must include a stages field",
		)
	}
}

// ---- Controller file --------------------------------------------------------

func TestValidationRunsControllerFileExists(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
		"backend/src/validation-runs.controller.ts",
		"backend/src/api/validation-runs/controller.ts",
	}
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			return
		}
	}
	t.Fatalf(
		"ValidationRun controller not found; expected it at one of: %v — "+
			"create the controller with list and detail methods",
		candidates,
	)
}

func TestValidationRunsControllerHasJSDocOnListMethod(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
		"backend/src/validation-runs.controller.ts",
		"backend/src/api/validation-runs/controller.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("controller file not found — cannot check for JSDoc on list method")
	}
	// JSDoc comments start with /**
	if !strings.Contains(content, "/**") {
		t.Error(
			"the ValidationRun controller must include JSDoc comments " +
				"(/** ... */) on controller methods to document the API contract",
		)
	}
}

func TestValidationRunsControllerHasDetailMethod(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
		"backend/src/validation-runs.controller.ts",
		"backend/src/api/validation-runs/controller.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("controller file not found — cannot check for detail method")
	}
	// The detail method must include stages in its query
	if !strings.Contains(content, "stages") {
		t.Error(
			"the ValidationRun controller detail method must include stages " +
				"(e.g. include: { stages: { orderBy: { stageIndex: 'asc' } } })",
		)
	}
}

// ---- Zod validation schema --------------------------------------------------

func TestValidationRunsListSchemaFileExists(t *testing.T) {
	// The list query params Zod schema may live in various locations
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.schema.ts",
		"backend/src/schemas/validation-run.schema.ts",
		"backend/src/api/validation-runs/schema.ts",
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
	}
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			return
		}
	}
	t.Fatalf(
		"Zod validation schema not found; expected it at one of: %v — "+
			"create a Zod schema for the list query params",
		candidates,
	)
}

func TestValidationRunsListSchemaIncludesCancelledStatus(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.schema.ts",
		"backend/src/schemas/validation-run.schema.ts",
		"backend/src/api/validation-runs/schema.ts",
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("Zod schema file not found — cannot check for CANCELLED status")
	}
	if !strings.Contains(content, "CANCELLED") {
		t.Error(
			"the Zod list query param schema must include CANCELLED as a valid status value " +
				"(all 5 ValidationRunStatus values: PENDING, RUNNING, SUCCESS, FAILED, CANCELLED)",
		)
	}
}

func TestValidationRunsListSchemaIncludesAllFiveStatuses(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.schema.ts",
		"backend/src/schemas/validation-run.schema.ts",
		"backend/src/api/validation-runs/schema.ts",
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("Zod schema file not found — cannot check for all status values")
	}
	requiredStatuses := []string{"PENDING", "RUNNING", "SUCCESS", "FAILED", "CANCELLED"}
	for _, status := range requiredStatuses {
		if !strings.Contains(content, status) {
			t.Errorf(
				"the Zod list query param schema must include status value %s", status,
			)
		}
	}
}

func TestValidationRunsListSchemaIncludesLimitParam(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.schema.ts",
		"backend/src/schemas/validation-run.schema.ts",
		"backend/src/api/validation-runs/schema.ts",
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("Zod schema file not found — cannot check for limit param")
	}
	if !strings.Contains(content, "limit") {
		t.Error(
			"the Zod list query param schema must include a limit parameter",
		)
	}
}

func TestValidationRunsListSchemaEnforcesMaxLimit100(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.schema.ts",
		"backend/src/schemas/validation-run.schema.ts",
		"backend/src/api/validation-runs/schema.ts",
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("Zod schema file not found — cannot check max limit enforcement")
	}
	// The max limit of 100 must be enforced; look for the value 100 near limit
	if !strings.Contains(content, "100") {
		t.Error(
			"the Zod list query param schema must enforce a maximum limit of 100 " +
				"(e.g. z.number().max(100))",
		)
	}
}

// ---- Routes file ------------------------------------------------------------

func TestValidationRunsRoutesFileExists(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.routes.ts",
		"backend/src/routes/validation-runs.ts",
		"backend/src/api/validation-runs/routes.ts",
		"backend/src/api/validation-runs/index.ts",
	}
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			return
		}
	}
	t.Fatalf(
		"ValidationRun routes file not found; expected it at one of: %v — "+
			"create a routes file that registers GET / and GET /:id endpoints",
		candidates,
	)
}

func TestValidationRunsRoutesRegistersListEndpoint(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.routes.ts",
		"backend/src/routes/validation-runs.ts",
		"backend/src/api/validation-runs/routes.ts",
		"backend/src/api/validation-runs/index.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("routes file not found — cannot check list endpoint registration")
	}
	// Must register a GET route for the list endpoint
	if !strings.Contains(strings.ToLower(content), "get") {
		t.Error(
			"the routes file must register a GET route for the list endpoint " +
				"(GET /api/validation-runs)",
		)
	}
}

func TestValidationRunsRoutesRegistersDetailEndpoint(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.routes.ts",
		"backend/src/routes/validation-runs.ts",
		"backend/src/api/validation-runs/routes.ts",
		"backend/src/api/validation-runs/index.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("routes file not found — cannot check detail endpoint registration")
	}
	// Must register a GET /:id route for the detail endpoint
	if !strings.Contains(content, ":id") && !strings.Contains(content, "/:id") {
		t.Error(
			"the routes file must register a GET /:id route for the detail endpoint " +
				"(GET /api/validation-runs/:id)",
		)
	}
}

// ---- Detail query: stages ordered by stageIndex ----------------------------

func TestValidationRunsControllerDetailQueryOrdersStagesByIndex(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
		"backend/src/validation-runs.controller.ts",
		"backend/src/api/validation-runs/controller.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("controller file not found — cannot check stages ordering")
	}
	// The detail query must include orderBy stageIndex
	if !strings.Contains(content, "stageIndex") {
		t.Error(
			"the detail controller method must include orderBy: { stageIndex: 'asc' } " +
				"when loading stages to ensure ascending order",
		)
	}
	if !strings.Contains(content, "asc") && !strings.Contains(content, "orderBy") {
		t.Error(
			"the detail controller method must use orderBy with ascending order " +
				"for stages (orderBy: { stageIndex: 'asc' })",
		)
	}
}

func TestValidationRunsControllerDetailQueryEagerLoadsStages(t *testing.T) {
	candidates := []string{
		"backend/src/api/validation-runs/validation-runs.controller.ts",
		"backend/src/controllers/validation-runs.controller.ts",
		"backend/src/validation-runs.controller.ts",
		"backend/src/api/validation-runs/controller.ts",
	}
	var content string
	for _, candidate := range candidates {
		if fileExists(t, candidate) {
			content = readFile(t, candidate)
			break
		}
	}
	if content == "" {
		t.Fatal("controller file not found — cannot check stages eager loading")
	}
	// Must use Prisma include to eagerly load stages
	if !strings.Contains(content, "include") {
		t.Error(
			"the detail controller method must use Prisma include to eagerly load stages " +
				"(e.g. include: { stages: { orderBy: { stageIndex: 'asc' } } })",
		)
	}
}
