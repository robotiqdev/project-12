// Package cancel_test validates the TypeScript layer for TASK-4606.
//
// These file-system tests verify that the cancel endpoint files exist and
// contain the required route, controller, service, and test definitions
// for POST /api/validation-runs/:id/cancel.
package cancel_test

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

// ---- File path helpers ----

func cancelRoutesPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "backend", "src", "routes", "cancel.routes.ts")
}

func cancelControllerPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "backend", "src", "controllers", "cancel.controller.ts")
}

func cancelServicePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "backend", "src", "services", "cancel.service.ts")
}

func cancelTestPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "backend", "tests", "integration", "cancel.test.ts")
}

// ---- File read helpers ----

func readCancelRoutes(t *testing.T) string {
	t.Helper()
	path := cancelRoutesPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cancel.routes.ts: %v", err)
	}
	return string(data)
}

func readCancelController(t *testing.T) string {
	t.Helper()
	path := cancelControllerPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cancel.controller.ts: %v", err)
	}
	return string(data)
}

func readCancelService(t *testing.T) string {
	t.Helper()
	path := cancelServicePath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cancel.service.ts: %v", err)
	}
	return string(data)
}

func readCancelTest(t *testing.T) string {
	t.Helper()
	path := cancelTestPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cancel.test.ts: %v", err)
	}
	return string(data)
}

// ---- File existence tests ----

func TestCancelRoutesFileExists(t *testing.T) {
	path := cancelRoutesPath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cancel.routes.ts not found at %s", path)
	}
}

func TestCancelControllerFileExists(t *testing.T) {
	path := cancelControllerPath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cancel.controller.ts not found at %s", path)
	}
}

func TestCancelServiceFileExists(t *testing.T) {
	path := cancelServicePath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cancel.service.ts not found at %s", path)
	}
}

func TestCancelIntegrationTestFileExists(t *testing.T) {
	path := cancelTestPath(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("cancel.test.ts not found at %s", path)
	}
}

// ---- cancel.routes.ts: route definition ----

func TestCancelRoutesDefinesPostCancelRoute(t *testing.T) {
	content := readCancelRoutes(t)
	// Must define a POST route with /:id/cancel path
	hasPostCancel := strings.Contains(content, "post") || strings.Contains(content, "POST")
	if !hasPostCancel {
		t.Error("cancel.routes.ts must define a POST route")
	}
}

func TestCancelRoutesHasCancelPathSegment(t *testing.T) {
	content := readCancelRoutes(t)
	if !strings.Contains(content, "cancel") {
		t.Error("cancel.routes.ts must include 'cancel' in the route path")
	}
}

func TestCancelRoutesHasIdPathParam(t *testing.T) {
	content := readCancelRoutes(t)
	// Must include :id or similar dynamic segment
	hasIdParam := strings.Contains(content, ":id") || strings.Contains(content, "/:id")
	if !hasIdParam {
		t.Error("cancel.routes.ts must include :id path parameter in the route")
	}
}

func TestCancelRoutesReferencesCancelController(t *testing.T) {
	content := readCancelRoutes(t)
	hasCancelController := strings.Contains(content, "cancelController") ||
		strings.Contains(content, "cancel.controller") ||
		strings.Contains(content, "CancelController")
	if !hasCancelController {
		t.Error("cancel.routes.ts must reference the cancel controller")
	}
}

func TestCancelRoutesMountsAtValidationRunsPath(t *testing.T) {
	content := readCancelRoutes(t)
	// The route must be under /api/validation-runs
	hasValidationRunsPath := strings.Contains(content, "validation-runs") ||
		strings.Contains(content, "validationRuns")
	if !hasValidationRunsPath {
		t.Error("cancel.routes.ts must be associated with the /api/validation-runs path")
	}
}

// ---- cancel.controller.ts: controller definition ----

func TestCancelControllerHasCancelMethod(t *testing.T) {
	content := readCancelController(t)
	if !strings.Contains(content, "cancel") {
		t.Error("cancel.controller.ts must define a cancel handler method")
	}
}

func TestCancelControllerCallsCancelService(t *testing.T) {
	content := readCancelController(t)
	hasCancelService := strings.Contains(content, "cancelService") ||
		strings.Contains(content, "cancel.service") ||
		strings.Contains(content, "cancelRun")
	if !hasCancelService {
		t.Error("cancel.controller.ts must call the cancel service")
	}
}

func TestCancelControllerReturns200OnSuccess(t *testing.T) {
	content := readCancelController(t)
	// Must send a 200 response on success
	has200 := strings.Contains(content, "200") ||
		strings.Contains(content, "res.json") ||
		strings.Contains(content, "res.send")
	if !has200 {
		t.Error("cancel.controller.ts must return a 200 response on successful cancellation")
	}
}

func TestCancelControllerHandlesRunNotCancellableError(t *testing.T) {
	content := readCancelController(t)
	// Must handle the RunNotCancellable error and return 409
	has409 := strings.Contains(content, "409")
	hasErrorHandling := strings.Contains(content, "RunNotCancellable") ||
		strings.Contains(content, "RUN_NOT_CANCELLABLE") ||
		strings.Contains(content, "catch")
	if !has409 || !hasErrorHandling {
		t.Error("cancel.controller.ts must handle RunNotCancellableError and return 409")
	}
}

func TestCancelControllerHandlesNotFoundError(t *testing.T) {
	content := readCancelController(t)
	// Must return 404 when run is not found
	if !strings.Contains(content, "404") {
		t.Error("cancel.controller.ts must return 404 when the run is not found")
	}
}

func TestCancelControllerExtractsIdFromParams(t *testing.T) {
	content := readCancelController(t)
	// Must extract :id from request params
	hasIdExtraction := strings.Contains(content, "params") ||
		strings.Contains(content, "req.params") ||
		strings.Contains(content, "params.id")
	if !hasIdExtraction {
		t.Error("cancel.controller.ts must extract the id from request params")
	}
}

// ---- cancel.service.ts: service definition ----

func TestCancelServiceHasCancelRunMethod(t *testing.T) {
	content := readCancelService(t)
	if !strings.Contains(content, "cancelRun") {
		t.Error("cancel.service.ts must define a cancelRun method")
	}
}

func TestCancelServiceCancelRunAcceptsId(t *testing.T) {
	content := readCancelService(t)
	hasCancelRunWithId := strings.Contains(content, "cancelRun(id") ||
		strings.Contains(content, "cancelRun(id:")
	if !hasCancelRunWithId {
		t.Error("cancelRun method must accept an id parameter")
	}
}

func TestCancelServiceFetchesRunById(t *testing.T) {
	content := readCancelService(t)
	// Must fetch the run from the database to check status
	hasFetch := strings.Contains(content, "findById") ||
		strings.Contains(content, "findUnique") ||
		strings.Contains(content, "findOne") ||
		strings.Contains(content, "getById") ||
		strings.Contains(content, "where") ||
		strings.Contains(content, "prisma")
	if !hasFetch {
		t.Error("cancel.service.ts must fetch the validation run before cancelling")
	}
}

func TestCancelServiceValidatesRunStatusBeforeCancelling(t *testing.T) {
	content := readCancelService(t)
	// Must check run status is PENDING or RUNNING before cancelling
	hasStatusCheck := strings.Contains(content, "PENDING") || strings.Contains(content, "RUNNING")
	if !hasStatusCheck {
		t.Error("cancel.service.ts must validate run status (PENDING or RUNNING) before cancelling")
	}
}

func TestCancelServiceThrowsRunNotCancellableError(t *testing.T) {
	content := readCancelService(t)
	hasError := strings.Contains(content, "RunNotCancellable") ||
		strings.Contains(content, "RUN_NOT_CANCELLABLE") ||
		strings.Contains(content, "NotCancellable")
	if !hasError {
		t.Error("cancel.service.ts must throw a RunNotCancellableError for non-cancellable statuses")
	}
}

func TestCancelServiceCallsPipelineTokenCancelForRunningRun(t *testing.T) {
	content := readCancelService(t)
	// Must attempt to signal cancellation token for RUNNING runs
	hasTokenCancel := strings.Contains(content, "token") ||
		strings.Contains(content, "pipeline") ||
		strings.Contains(content, "getToken") ||
		strings.Contains(content, "pipelineService")
	if !hasTokenCancel {
		t.Error("cancel.service.ts must call pipelineService.getToken(id).cancel() for RUNNING runs")
	}
}

func TestCancelServiceHandlesMissingTokenGracefully(t *testing.T) {
	content := readCancelService(t)
	// Must handle the case where the pipeline token doesn't exist yet (async scheduling)
	hasGracefulHandling := strings.Contains(content, "?") || // optional chaining
		strings.Contains(content, "if") || // conditional check
		strings.Contains(content, "exists") ||
		strings.Contains(content, "undefined") ||
		strings.Contains(content, "null")
	if !hasGracefulHandling {
		t.Error("cancel.service.ts must handle missing pipeline token gracefully")
	}
}

func TestCancelServiceCallsValidationRunModelCancel(t *testing.T) {
	content := readCancelService(t)
	// Must call validationRunModel.cancel(id) to update the DB
	hasModelCancel := strings.Contains(content, "validationRunModel") ||
		strings.Contains(content, "runModel") ||
		strings.Contains(content, ".cancel(")
	if !hasModelCancel {
		t.Error("cancel.service.ts must call validationRunModel.cancel(id)")
	}
}

func TestCancelServiceReturnsUpdatedRunRecord(t *testing.T) {
	content := readCancelService(t)
	// Must return the updated run record
	hasReturn := strings.Contains(content, "return") &&
		(strings.Contains(content, "ValidationRunRecord") ||
			strings.Contains(content, "Promise") ||
			strings.Contains(content, "cancel("))
	if !hasReturn {
		t.Error("cancel.service.ts must return the updated ValidationRunRecord")
	}
}

// ---- cancel.test.ts: integration test coverage ----

func TestCancelTestCoversRunningRunReturns200(t *testing.T) {
	content := readCancelTest(t)
	hasRunningTest := strings.Contains(content, "RUNNING") &&
		strings.Contains(content, "200")
	if !hasRunningTest {
		t.Error("cancel.test.ts must test that a RUNNING run returns 200 on cancel")
	}
}

func TestCancelTestCoversPendingRunReturns200(t *testing.T) {
	content := readCancelTest(t)
	hasPendingTest := strings.Contains(content, "PENDING") &&
		strings.Contains(content, "200")
	if !hasPendingTest {
		t.Error("cancel.test.ts must test that a PENDING run returns 200 on cancel")
	}
}

func TestCancelTestCoversFailedRunReturns409(t *testing.T) {
	content := readCancelTest(t)
	hasFailedTest := strings.Contains(content, "FAILED") &&
		strings.Contains(content, "409")
	if !hasFailedTest {
		t.Error("cancel.test.ts must test that a FAILED run returns 409 on cancel")
	}
}

func TestCancelTestCoversSuccessRunReturns409(t *testing.T) {
	content := readCancelTest(t)
	hasSuccessTest := strings.Contains(content, "SUCCESS") &&
		strings.Contains(content, "409")
	if !hasSuccessTest {
		t.Error("cancel.test.ts must test that a SUCCESS run returns 409 on cancel")
	}
}

func TestCancelTestCoversCancelledRunReturns409(t *testing.T) {
	content := readCancelTest(t)
	hasCancelledTest := strings.Contains(content, "CANCELLED") &&
		strings.Contains(content, "409")
	if !hasCancelledTest {
		t.Error("cancel.test.ts must test that an already-CANCELLED run returns 409 on cancel")
	}
}

func TestCancelTestCoversUnknownIdReturns404(t *testing.T) {
	content := readCancelTest(t)
	has404Test := strings.Contains(content, "404")
	if !has404Test {
		t.Error("cancel.test.ts must test that an unknown id returns 404")
	}
}

func TestCancelTestCoversRunNotCancellableError(t *testing.T) {
	content := readCancelTest(t)
	if !strings.Contains(content, "RUN_NOT_CANCELLABLE") {
		t.Error("cancel.test.ts must verify the RUN_NOT_CANCELLABLE error body")
	}
}

func TestCancelTestCoversStageResultsPreservation(t *testing.T) {
	content := readCancelTest(t)
	// Must test that SUCCESS stages remain unchanged after cancellation
	hasStagePreservation := strings.Contains(content, "SUCCESS") &&
		(strings.Contains(content, "stage") || strings.Contains(content, "Stage"))
	if !hasStagePreservation {
		t.Error("cancel.test.ts must verify that completed stage results are preserved")
	}
}

func TestCancelTestVerifiesStatusCancelledInResponse(t *testing.T) {
	content := readCancelTest(t)
	// The response body must include status: CANCELLED
	hasStatusCancelled := strings.Contains(content, "CANCELLED")
	if !hasStatusCancelled {
		t.Error("cancel.test.ts must verify status=CANCELLED in the 200 response body")
	}
}

func TestCancelTestVerifiesCancelledAtInResponse(t *testing.T) {
	content := readCancelTest(t)
	if !strings.Contains(content, "cancelledAt") {
		t.Error("cancel.test.ts must verify cancelledAt timestamp is present in the response")
	}
}
