// Package services_test validates the TypeScript source files for TASK-4596.
// These are file-system tests that verify the SSE types, SSE service, SSE
// controller, updated routes, and Jest test files contain all required
// definitions.
//
// Tests WILL FAIL until the implementation files are created — this is the
// intended TDD behaviour.
package services_test

import (
	"strings"
	"testing"
)

// ── File existence ──────────────────────────────────────────────────────────

func TestSSETypesFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/types/sse.types.ts") {
		t.Fatal("backend/src/types/sse.types.ts does not exist")
	}
}

func TestSSEServiceFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/services/sse.service.ts") {
		t.Fatal("backend/src/services/sse.service.ts does not exist")
	}
}

func TestSSEControllerFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/controllers/sse.controller.ts") {
		t.Fatal("backend/src/controllers/sse.controller.ts does not exist")
	}
}

func TestSSEServiceUnitTestFileExists(t *testing.T) {
	if !fileExists(t, "backend/tests/unit/services/sse.service.test.ts") {
		t.Fatal("backend/tests/unit/services/sse.service.test.ts does not exist")
	}
}

func TestSSEIntegrationTestFileExists(t *testing.T) {
	if !fileExists(t, "backend/tests/integration/sse.test.ts") {
		t.Fatal("backend/tests/integration/sse.test.ts does not exist")
	}
}

// ── sse.types.ts ─────────────────────────────────────────────────────────────

func TestSSETypesExportsSSEEventType(t *testing.T) {
	src := readFile(t, "backend/src/types/sse.types.ts")
	if !strings.Contains(src, "SSEEventType") {
		t.Error("sse.types.ts must export SSEEventType union type")
	}
}

func TestSSETypesExportsSSEEventInterface(t *testing.T) {
	src := readFile(t, "backend/src/types/sse.types.ts")
	if !strings.Contains(src, "SSEEvent") {
		t.Error("sse.types.ts must export SSEEvent interface")
	}
}

func TestSSETypesHasLogLineEventType(t *testing.T) {
	src := readFile(t, "backend/src/types/sse.types.ts")
	if !strings.Contains(src, "log_line") {
		t.Error("sse.types.ts SSEEventType must include 'log_line'")
	}
}

func TestSSETypesHasConnectedEventType(t *testing.T) {
	src := readFile(t, "backend/src/types/sse.types.ts")
	if !strings.Contains(src, "connected") {
		t.Error("sse.types.ts SSEEventType must include 'connected'")
	}
}

func TestSSETypesHasStageStartedEventType(t *testing.T) {
	src := readFile(t, "backend/src/types/sse.types.ts")
	if !strings.Contains(src, "stage_started") {
		t.Error("sse.types.ts SSEEventType must include 'stage_started'")
	}
}

func TestSSETypesHasRunCompletedEventType(t *testing.T) {
	src := readFile(t, "backend/src/types/sse.types.ts")
	if !strings.Contains(src, "run_completed") {
		t.Error("sse.types.ts SSEEventType must include 'run_completed'")
	}
}

// ── sse.service.ts ────────────────────────────────────────────────────────────

func TestSSEServiceExportsSubscribeFunction(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "subscribe") {
		t.Error("sse.service.ts must export subscribe function")
	}
}

func TestSSEServiceExportsEmitFunction(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "emit") {
		t.Error("sse.service.ts must export emit function")
	}
}

func TestSSEServiceExportsUnsubscribeFunction(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "unsubscribe") {
		t.Error("sse.service.ts must export unsubscribe function")
	}
}

func TestSSEServiceUsesClientsMap(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "clients") {
		t.Error("sse.service.ts must declare a 'clients' Map for storing subscribers")
	}
}

func TestSSEServiceUsesMapType(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "Map") {
		t.Error("sse.service.ts clients must be typed as Map<string, Set<Response>>")
	}
}

func TestSSEServiceSubscribeSetsContentTypeHeader(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "Content-Type") {
		t.Error("sse.service.ts subscribe must set the Content-Type response header")
	}
}

func TestSSEServiceSubscribeSetsTextEventStreamContentType(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "text/event-stream") {
		t.Error("sse.service.ts subscribe must set Content-Type: text/event-stream")
	}
}

func TestSSEServiceSubscribeSendsRetryDirective(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "retry:") {
		t.Error("sse.service.ts subscribe must send 'retry: 3000\\n\\n' directive to client")
	}
}

func TestSSEServiceSubscribeSendsConnectedEvent(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "connected") {
		t.Error("sse.service.ts subscribe must send initial 'event: connected\\ndata: {}\\n\\n'")
	}
}

func TestSSEServiceSubscribeSetsUpHeartbeatInterval(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "setInterval") {
		t.Error("sse.service.ts subscribe must call setInterval to set up a 15-second heartbeat")
	}
}

func TestSSEServiceSubscribeSendsHeartbeatComment(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, ":heartbeat") {
		t.Error("sse.service.ts subscribe heartbeat must write ':heartbeat\\n\\n' comment")
	}
}

func TestSSEServiceSubscribeRegistersCloseHandler(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "close") {
		t.Error("sse.service.ts subscribe must register a 'close' event handler on the Response")
	}
}

func TestSSEServiceEmitFormatsWithIdField(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "id:") {
		t.Error("sse.service.ts emit must write 'id:\\n' field in SSE wire format")
	}
}

func TestSSEServiceEmitFormatsWithEventField(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "event:") {
		t.Error("sse.service.ts emit must write 'event:\\n' field in SSE wire format")
	}
}

func TestSSEServiceEmitFormatsWithDataField(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "data:") {
		t.Error("sse.service.ts emit must write 'data:\\n' field in SSE wire format")
	}
}

func TestSSEServiceUnsubscribeClearsHeartbeatInterval(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "clearInterval") {
		t.Error("sse.service.ts unsubscribe must call clearInterval to clean up heartbeat timer")
	}
}

func TestSSEServiceStoresHeartbeatIntervalPerResponse(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	if !strings.Contains(src, "WeakMap") && !strings.Contains(src, "intervalId") {
		t.Error("sse.service.ts must store heartbeat intervalId per Response (WeakMap or intervalId property)")
	}
}

func TestSSEServiceUnsubscribeDeletesKeyWhenSetEmpty(t *testing.T) {
	src := readFile(t, "backend/src/services/sse.service.ts")
	// Implementation should delete the key when Set becomes empty
	if !strings.Contains(src, "delete") {
		t.Error("sse.service.ts unsubscribe must delete the runId key when all subscribers are removed")
	}
}

// ── sse.controller.ts ─────────────────────────────────────────────────────────

func TestSSEControllerExportsSubscribeHandler(t *testing.T) {
	src := readFile(t, "backend/src/controllers/sse.controller.ts")
	if !strings.Contains(src, "subscribe") {
		t.Error("sse.controller.ts must export a subscribe handler function")
	}
}

func TestSSEControllerValidatesRunIdExists(t *testing.T) {
	src := readFile(t, "backend/src/controllers/sse.controller.ts")
	if !strings.Contains(src, "findById") && !strings.Contains(src, "validationRunModel") {
		t.Error("sse.controller.ts subscribe must validate that the runId exists in the database")
	}
}

func TestSSEControllerCallsSSEServiceSubscribe(t *testing.T) {
	src := readFile(t, "backend/src/controllers/sse.controller.ts")
	if !strings.Contains(src, "sseService") {
		t.Error("sse.controller.ts must import and call sseService.subscribe(runId, res)")
	}
}

func TestSSEControllerReturns404ForUnknownRunId(t *testing.T) {
	src := readFile(t, "backend/src/controllers/sse.controller.ts")
	if !strings.Contains(src, "404") {
		t.Error("sse.controller.ts subscribe must return 404 when runId does not exist")
	}
}

// ── validation-runs.routes.ts ─────────────────────────────────────────────────

func TestValidationRunsRoutesHasStreamEndpoint(t *testing.T) {
	src := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	if !strings.Contains(src, "stream") {
		t.Error("validation-runs.routes.ts must register GET /:id/stream route")
	}
}

func TestValidationRunsRoutesImportsSseController(t *testing.T) {
	src := readFile(t, "backend/src/routes/validation-runs.routes.ts")
	if !strings.Contains(src, "sse") {
		t.Error("validation-runs.routes.ts must import sseController for the stream route")
	}
}

// ── sse.service.test.ts (unit test coverage checks) ──────────────────────────

func TestSSEServiceTestFileTestsSubscribeFunction(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "subscribe") {
		t.Error("sse.service.test.ts must test the subscribe function")
	}
}

func TestSSEServiceTestFileTestsEmitFunction(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "emit") {
		t.Error("sse.service.test.ts must test the emit function")
	}
}

func TestSSEServiceTestFileTestsUnsubscribeFunction(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "unsubscribe") {
		t.Error("sse.service.test.ts must test the unsubscribe function")
	}
}

func TestSSEServiceTestFileMocksResponseWithWrite(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "write") {
		t.Error("sse.service.test.ts must mock Response object with a write method")
	}
}

func TestSSEServiceTestFileMocksResponseWithSetHeader(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "setHeader") {
		t.Error("sse.service.test.ts must mock Response object with a setHeader method")
	}
}

func TestSSEServiceTestFileMocksResponseWithOn(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, ".on") {
		t.Error("sse.service.test.ts must mock Response object with an on method")
	}
}

func TestSSEServiceTestFileTestsContentTypeHeader(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "Content-Type") {
		t.Error("sse.service.test.ts must test that subscribe sets the Content-Type header")
	}
}

func TestSSEServiceTestFileTestsTextEventStreamValue(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "text/event-stream") {
		t.Error("sse.service.test.ts must assert Content-Type is set to text/event-stream")
	}
}

func TestSSEServiceTestFileTestsSSEEventField(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "event:") {
		t.Error("sse.service.test.ts must test that emit includes 'event:' in SSE wire format")
	}
}

func TestSSEServiceTestFileTestsSSEDataField(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "data:") {
		t.Error("sse.service.test.ts must test that emit includes 'data:' in SSE wire format")
	}
}

func TestSSEServiceTestFileTestsUnsubscribeRemovesClient(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "unsubscribe") {
		t.Error("sse.service.test.ts must test that unsubscribe removes client from clients map")
	}
}

func TestSSEServiceTestFileTestsSubscribingToUnknownRunId(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "unknown") {
		t.Error("sse.service.test.ts must test that subscribing to unknown runId creates new map entry")
	}
}

func TestSSEServiceTestFileTestsHeartbeat(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/sse.service.test.ts")
	if !strings.Contains(src, "heartbeat") {
		t.Error("sse.service.test.ts must test that subscribe sets up a heartbeat")
	}
}

// ── sse.test.ts (integration test coverage checks) ───────────────────────────

func TestSSEIntegrationTestFileTestsStreamEndpoint(t *testing.T) {
	src := readFile(t, "backend/tests/integration/sse.test.ts")
	if !strings.Contains(src, "stream") {
		t.Error("sse.test.ts must test the GET /:id/stream endpoint")
	}
}

func TestSSEIntegrationTestFileTestsLogLineEvents(t *testing.T) {
	src := readFile(t, "backend/tests/integration/sse.test.ts")
	if !strings.Contains(src, "log_line") {
		t.Error("sse.test.ts must assert that log_line events are received from the stream")
	}
}

func TestSSEIntegrationTestFileTestsTextEventStream(t *testing.T) {
	src := readFile(t, "backend/tests/integration/sse.test.ts")
	if !strings.Contains(src, "text/event-stream") {
		t.Error("sse.test.ts must verify Content-Type is text/event-stream")
	}
}

func TestSSEIntegrationTestFileTestsConnectedEvent(t *testing.T) {
	src := readFile(t, "backend/tests/integration/sse.test.ts")
	if !strings.Contains(src, "connected") {
		t.Error("sse.test.ts must test that the initial 'connected' event is received")
	}
}

func TestSSEIntegrationTestFileTests404ForUnknownRun(t *testing.T) {
	src := readFile(t, "backend/tests/integration/sse.test.ts")
	if !strings.Contains(src, "404") {
		t.Error("sse.test.ts must test that 404 is returned for an unknown runId")
	}
}
