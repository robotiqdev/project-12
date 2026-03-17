// Package integration_test validates the log export service for TASK-4589.
//
// These tests verify that the following implementation files exist and contain
// all required definitions:
//   - backend/src/services/log-export.service.ts  (generateZip, getAllLogsAsText, getStageLog)
//   - backend/src/routes/logs.routes.ts            (GET /api/validation-runs/:id/logs/zip, etc.)
//   - backend/src/controllers/logs.controller.ts   (corresponding request handlers)
//   - backend/tests/integration/logs.test.ts       (integration test suite)
package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Path helpers
// ---------------------------------------------------------------------------

func logExportServicePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "backend", "src", "services", "log-export.service.ts")
}

func logsRoutesPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "backend", "src", "routes", "logs.routes.ts")
}

func logsControllerPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "backend", "src", "controllers", "logs.controller.ts")
}

func logsTestPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "backend", "tests", "integration", "logs.test.ts")
}

func readLogExportService(t *testing.T) string {
	t.Helper()
	path := logExportServicePath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"failed to read log-export.service.ts (has the service been created?): %v", err,
		)
	}
	return string(data)
}

func readLogsRoutes(t *testing.T) string {
	t.Helper()
	path := logsRoutesPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"failed to read logs.routes.ts (has the routes file been created?): %v", err,
		)
	}
	return string(data)
}

func readLogsController(t *testing.T) string {
	t.Helper()
	path := logsControllerPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"failed to read logs.controller.ts (has the controller been created?): %v", err,
		)
	}
	return string(data)
}

// ---------------------------------------------------------------------------
// File existence tests
// ---------------------------------------------------------------------------

func TestLogExportServiceFileExists(t *testing.T) {
	path := logExportServicePath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf(
			"log-export.service.ts not found at %s — create backend/src/services/log-export.service.ts",
			path,
		)
	}
}

func TestLogsRoutesFileExists(t *testing.T) {
	path := logsRoutesPath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf(
			"logs.routes.ts not found at %s — create backend/src/routes/logs.routes.ts",
			path,
		)
	}
}

func TestLogsControllerFileExists(t *testing.T) {
	path := logsControllerPath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf(
			"logs.controller.ts not found at %s — create backend/src/controllers/logs.controller.ts",
			path,
		)
	}
}

func TestLogsTestFileExists(t *testing.T) {
	path := logsTestPath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf(
			"logs.test.ts not found at %s — create backend/tests/integration/logs.test.ts",
			path,
		)
	}
}

// ---------------------------------------------------------------------------
// Service file content tests — log-export.service.ts
// ---------------------------------------------------------------------------

func TestLogExportServiceNotEmpty(t *testing.T) {
	src := readLogExportService(t)
	if strings.TrimSpace(src) == "" {
		t.Error("log-export.service.ts must not be empty")
	}
}

func TestLogExportServiceExportsGenerateZip(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "generateZip") {
		t.Error(
			"log-export.service.ts must export a generateZip function " +
				"that streams a ZIP archive to the HTTP response",
		)
	}
}

func TestLogExportServiceExportsGetAllLogsAsText(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "getAllLogsAsText") {
		t.Error(
			"log-export.service.ts must export a getAllLogsAsText function " +
				"that returns all stage logs concatenated as plain text",
		)
	}
}

func TestLogExportServiceExportsGetStageLog(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "getStageLog") {
		t.Error(
			"log-export.service.ts must export a getStageLog function " +
				"that returns the log content for a specific stage",
		)
	}
}

func TestLogExportServiceUsesArchiver(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "archiver") {
		t.Error(
			"log-export.service.ts must use the `archiver` package to create ZIP archives " +
				"(import archiver from 'archiver' or similar)",
		)
	}
}

func TestLogExportServicePipesToResponse(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "pipe") {
		t.Error(
			"log-export.service.ts must pipe the archiver stream directly to the HTTP response " +
				"(use archive.pipe(res) to stream without buffering)",
		)
	}
}

func TestLogExportServiceSetsContentTypeApplicationZip(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "application/zip") {
		t.Error(
			"log-export.service.ts must set Content-Type to 'application/zip' " +
				"before creating the archive (res.setHeader or res.set)",
		)
	}
}

func TestLogExportServiceSetsContentDispositionAttachment(t *testing.T) {
	src := readLogExportService(t)
	lower := strings.ToLower(src)
	if !strings.Contains(lower, "content-disposition") && !strings.Contains(lower, "attachment") {
		t.Error(
			"log-export.service.ts must set Content-Disposition to " +
				"attachment; filename=\"run-{id}.zip\"",
		)
	}
}

func TestLogExportServiceAppendsStageLogFiles(t *testing.T) {
	src := readLogExportService(t)
	// The service must append each stage's log as a named file in the archive
	if !strings.Contains(src, "stage-") && !strings.Contains(src, "stageIndex") {
		t.Error(
			"log-export.service.ts must append stage log files named " +
				"stage-{index}-{stageName}.log to the archive",
		)
	}
}

func TestLogExportServiceAppendsMetadataJSON(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "metadata.json") {
		t.Error(
			"log-export.service.ts must append a metadata.json entry to the ZIP archive " +
				"containing run id, status, and timestamps",
		)
	}
}

func TestLogExportServiceMetadataIncludesRunId(t *testing.T) {
	src := readLogExportService(t)
	// The metadata.json must include id/runId field
	if !strings.Contains(src, `"id"`) && !strings.Contains(src, "runId") && !strings.Contains(src, ".id") {
		t.Error(
			"log-export.service.ts metadata.json must include the run id field",
		)
	}
}

func TestLogExportServiceHandlesArchiveErrors(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "error") {
		t.Error(
			"log-export.service.ts must handle archiver errors " +
				"using archive.on('error', ...) or similar error handling",
		)
	}
}

func TestLogExportServiceFinalizesArchive(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "finalize") {
		t.Error(
			"log-export.service.ts must call archive.finalize() to complete the ZIP archive",
		)
	}
}

func TestLogExportServiceGetAllLogsUsesStageNameHeader(t *testing.T) {
	src := readLogExportService(t)
	// getAllLogsAsText must format each stage with === STAGE: {stageName} === header
	if !strings.Contains(src, "=== STAGE:") && !strings.Contains(src, "STAGE:") {
		t.Error(
			"log-export.service.ts getAllLogsAsText must prepend " +
				"\"=== STAGE: {stageName} ===\" header before each stage's log content",
		)
	}
}

func TestLogExportServiceGetAllLogsUsesSeparator(t *testing.T) {
	src := readLogExportService(t)
	// getAllLogsAsText must use --- as separator between stages
	if !strings.Contains(src, "---") {
		t.Error(
			"log-export.service.ts getAllLogsAsText must use '---' as a separator between stage logs",
		)
	}
}

func TestLogExportServiceGetStageLogFetchesByStageIndex(t *testing.T) {
	src := readLogExportService(t)
	if !strings.Contains(src, "stageIndex") {
		t.Error(
			"log-export.service.ts getStageLog must accept and use stageIndex " +
				"to fetch the specific stage's log",
		)
	}
}

// ---------------------------------------------------------------------------
// Routes file content tests — logs.routes.ts
// ---------------------------------------------------------------------------

func TestLogsRoutesNotEmpty(t *testing.T) {
	src := readLogsRoutes(t)
	if strings.TrimSpace(src) == "" {
		t.Error("logs.routes.ts must not be empty")
	}
}

func TestLogsRoutesDefinesZipRoute(t *testing.T) {
	src := readLogsRoutes(t)
	if !strings.Contains(src, "logs/zip") {
		t.Error(
			"logs.routes.ts must define a GET route for /api/validation-runs/:id/logs/zip",
		)
	}
}

func TestLogsRoutesDefinesTextRoute(t *testing.T) {
	src := readLogsRoutes(t)
	if !strings.Contains(src, "logs/text") {
		t.Error(
			"logs.routes.ts must define a GET route for /api/validation-runs/:id/logs/text",
		)
	}
}

func TestLogsRoutesDefinesStageLogRoute(t *testing.T) {
	src := readLogsRoutes(t)
	// Route like /stages/:stageIndex/logs
	if !strings.Contains(src, "stages") || !strings.Contains(src, "stageIndex") {
		t.Error(
			"logs.routes.ts must define a GET route for " +
				"/api/validation-runs/:id/stages/:stageIndex/logs",
		)
	}
}

func TestLogsRoutesUsesGetMethod(t *testing.T) {
	src := readLogsRoutes(t)
	lower := strings.ToLower(src)
	// Express uses router.get() or app.get()
	if !strings.Contains(lower, ".get(") {
		t.Error(
			"logs.routes.ts must use HTTP GET method for all log export routes",
		)
	}
}

func TestLogsRoutesImportsController(t *testing.T) {
	src := readLogsRoutes(t)
	if !strings.Contains(src, "controller") && !strings.Contains(src, "Controller") {
		t.Error(
			"logs.routes.ts must import the logs controller to wire up route handlers",
		)
	}
}

func TestLogsRoutesHasRunIdParam(t *testing.T) {
	src := readLogsRoutes(t)
	// The route must have a :id or :runId parameter
	if !strings.Contains(src, ":id") && !strings.Contains(src, ":runId") {
		t.Error(
			"logs.routes.ts must use a :id or :runId route parameter to identify the run",
		)
	}
}

// ---------------------------------------------------------------------------
// Controller file content tests — logs.controller.ts
// ---------------------------------------------------------------------------

func TestLogsControllerNotEmpty(t *testing.T) {
	src := readLogsController(t)
	if strings.TrimSpace(src) == "" {
		t.Error("logs.controller.ts must not be empty")
	}
}

func TestLogsControllerDefinesZipHandler(t *testing.T) {
	src := readLogsController(t)
	if !strings.Contains(src, "zip") && !strings.Contains(src, "Zip") {
		t.Error(
			"logs.controller.ts must define a handler for the ZIP download endpoint",
		)
	}
}

func TestLogsControllerDefinesTextHandler(t *testing.T) {
	src := readLogsController(t)
	if !strings.Contains(src, "text") && !strings.Contains(src, "Text") {
		t.Error(
			"logs.controller.ts must define a handler for the plain-text log endpoint",
		)
	}
}

func TestLogsControllerDefinesStageLogHandler(t *testing.T) {
	src := readLogsController(t)
	if !strings.Contains(src, "stage") && !strings.Contains(src, "Stage") {
		t.Error(
			"logs.controller.ts must define a handler for the single-stage log endpoint",
		)
	}
}

func TestLogsControllerImportsLogExportService(t *testing.T) {
	src := readLogsController(t)
	if !strings.Contains(src, "log-export") && !strings.Contains(src, "logExport") {
		t.Error(
			"logs.controller.ts must import from log-export.service.ts to delegate business logic",
		)
	}
}

func TestLogsControllerHandles404ForUnknownRun(t *testing.T) {
	src := readLogsController(t)
	// Controller must send 404 when run is not found
	if !strings.Contains(src, "404") {
		t.Error(
			"logs.controller.ts must return HTTP 404 when the requested run is not found",
		)
	}
}

func TestLogsControllerHandlesRequestAndResponse(t *testing.T) {
	src := readLogsController(t)
	lower := strings.ToLower(src)
	// Must reference req and res (Express handler params)
	if !strings.Contains(lower, "req") || !strings.Contains(lower, "res") {
		t.Error(
			"logs.controller.ts must define Express-compatible handlers with req/res parameters",
		)
	}
}

func TestLogsControllerExtractsRunIdFromParams(t *testing.T) {
	src := readLogsController(t)
	// Controller must extract the run ID from route params
	if !strings.Contains(src, "params") && !strings.Contains(src, "req.params") {
		t.Error(
			"logs.controller.ts must extract the run ID from request params (req.params.id or similar)",
		)
	}
}

func TestLogsControllerZipHandlerSets200OnSuccess(t *testing.T) {
	src := readLogsController(t)
	// The ZIP handler should set/send 200 or rely on the service to stream it
	if !strings.Contains(src, "generateZip") {
		t.Error(
			"logs.controller.ts ZIP handler must call the generateZip service method",
		)
	}
}

func TestLogsControllerTextHandlerSends200WithText(t *testing.T) {
	src := readLogsController(t)
	if !strings.Contains(src, "getAllLogsAsText") {
		t.Error(
			"logs.controller.ts text handler must call getAllLogsAsText from the service",
		)
	}
}

func TestLogsControllerStageHandlerCallsGetStageLog(t *testing.T) {
	src := readLogsController(t)
	if !strings.Contains(src, "getStageLog") {
		t.Error(
			"logs.controller.ts stage log handler must call getStageLog from the service",
		)
	}
}

// ---------------------------------------------------------------------------
// TypeScript test file content tests
// ---------------------------------------------------------------------------

func TestLogsTestFileNotEmpty(t *testing.T) {
	path := logsTestPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read logs.test.ts: %v", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		t.Error("logs.test.ts must not be empty")
	}
}

func TestLogsTestFileTestsZipEndpoint(t *testing.T) {
	path := logsTestPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read logs.test.ts: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "logs/zip") {
		t.Error("logs.test.ts must contain tests for the GET /logs/zip endpoint")
	}
}

func TestLogsTestFileTestsTextEndpoint(t *testing.T) {
	path := logsTestPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read logs.test.ts: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "logs/text") {
		t.Error("logs.test.ts must contain tests for the GET /logs/text endpoint")
	}
}

func TestLogsTestFileTestsStageLogEndpoint(t *testing.T) {
	path := logsTestPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read logs.test.ts: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "stages") {
		t.Error("logs.test.ts must contain tests for the GET /stages/:stageIndex/logs endpoint")
	}
}

func TestLogsTestFileTests404ForUnknownRun(t *testing.T) {
	path := logsTestPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read logs.test.ts: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "404") {
		t.Error("logs.test.ts must contain tests verifying 404 responses for unknown runIds")
	}
}

func TestLogsTestFileSeeds5Stages(t *testing.T) {
	path := logsTestPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read logs.test.ts: %v", err)
	}
	src := string(data)
	// Must reference all 5 stage types
	for _, stage := range []string{"SETUP", "LINT", "TEST", "BUILD", "DEPLOY"} {
		if !strings.Contains(src, stage) {
			t.Errorf("logs.test.ts must seed/reference stage type %s in tests", stage)
		}
	}
}
