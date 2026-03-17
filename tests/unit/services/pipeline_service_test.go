// Package services_test validates the TypeScript source files for TASK-4594.
// These are file-system tests that verify the pipeline types, pipeline service,
// stage-executor service, and the Jest test file contain all required definitions.
//
// Tests WILL FAIL until the implementation files are created — this is the
// intended TDD behaviour.
package services_test

import (
	"strings"
	"testing"
)

// ── File existence ─────────────────────────────────────────────────────────────

func TestPipelineTypesFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/types/pipeline.types.ts") {
		t.Fatal("backend/src/types/pipeline.types.ts does not exist")
	}
}

func TestPipelineServiceFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/services/pipeline.service.ts") {
		t.Fatal("backend/src/services/pipeline.service.ts does not exist")
	}
}

func TestStageExecutorServiceFileExists(t *testing.T) {
	if !fileExists(t, "backend/src/services/stage-executor.service.ts") {
		t.Fatal("backend/src/services/stage-executor.service.ts does not exist")
	}
}

func TestPipelineServiceTestFileExists(t *testing.T) {
	if !fileExists(t, "backend/tests/unit/services/pipeline.service.test.ts") {
		t.Fatal("backend/tests/unit/services/pipeline.service.test.ts does not exist")
	}
}

// ── pipeline.types.ts ─────────────────────────────────────────────────────────

func TestPipelineTypesCancellationTokenDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/pipeline.types.ts")
	if !strings.Contains(src, "CancellationToken") {
		t.Error("pipeline.types.ts must export CancellationToken type/interface")
	}
}

func TestPipelineTypesCancellationTokenHasIsCancelledField(t *testing.T) {
	src := readFile(t, "backend/src/types/pipeline.types.ts")
	if !strings.Contains(src, "isCancelled") {
		t.Error("pipeline.types.ts CancellationToken must have isCancelled field")
	}
}

func TestPipelineTypesPipelineExecutionContextDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/pipeline.types.ts")
	if !strings.Contains(src, "PipelineExecutionContext") {
		t.Error("pipeline.types.ts must export PipelineExecutionContext type/interface")
	}
}

func TestPipelineTypesPipelineExecutionContextHasRunId(t *testing.T) {
	src := readFile(t, "backend/src/types/pipeline.types.ts")
	if !strings.Contains(src, "runId") {
		t.Error("pipeline.types.ts PipelineExecutionContext must have runId field")
	}
}

func TestPipelineTypesPipelineExecutionContextHasToken(t *testing.T) {
	src := readFile(t, "backend/src/types/pipeline.types.ts")
	if !strings.Contains(src, "token") {
		t.Error("pipeline.types.ts PipelineExecutionContext must have token field")
	}
}

func TestPipelineTypesStageExecutorFnDefined(t *testing.T) {
	src := readFile(t, "backend/src/types/pipeline.types.ts")
	if !strings.Contains(src, "StageExecutorFn") {
		t.Error("pipeline.types.ts must export StageExecutorFn type")
	}
}

// ── pipeline.service.ts ───────────────────────────────────────────────────────

func TestPipelineServiceExportsEnqueue(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "enqueue") {
		t.Error("pipeline.service.ts must export enqueue function")
	}
}

func TestPipelineServiceExportsExecute(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "execute") {
		t.Error("pipeline.service.ts must export execute function")
	}
}

func TestPipelineServiceExportsUnregisterToken(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "unregisterToken") {
		t.Error("pipeline.service.ts must export unregisterToken function")
	}
}

func TestPipelineServiceMaintainsCancellationTokenRegistry(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "Map") {
		t.Error("pipeline.service.ts must maintain a Map registry for CancellationToken instances")
	}
}

func TestPipelineServiceUsesStageOrder(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "STAGE_ORDER") {
		t.Error("pipeline.service.ts must iterate over STAGE_ORDER to execute stages in sequence")
	}
}

func TestPipelineServiceChecksTokenCancellation(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "isCancelled") {
		t.Error("pipeline.service.ts must check token.isCancelled between stage executions")
	}
}

func TestPipelineServiceUpdatesRunToRunning(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "RUNNING") {
		t.Error("pipeline.service.ts execute must update run status to RUNNING")
	}
}

func TestPipelineServiceHandlesRunSuccess(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "SUCCESS") {
		t.Error("pipeline.service.ts execute must mark run as SUCCESS on completion")
	}
}

func TestPipelineServiceHandlesRunFailed(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "FAILED") {
		t.Error("pipeline.service.ts execute must mark run as FAILED when a stage errors")
	}
}

func TestPipelineServiceHandlesCancellation(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "CANCELLED") {
		t.Error("pipeline.service.ts execute must mark run as CANCELLED when token is set")
	}
}

func TestPipelineServiceCallsStageExecutorService(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "stageExecutorService") && !strings.Contains(src, "executeStage") {
		t.Error("pipeline.service.ts must delegate stage execution to stageExecutorService.executeStage")
	}
}

func TestPipelineServiceEmitsSseEvents(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "sseService") && !strings.Contains(src, "sse") {
		t.Error("pipeline.service.ts must emit SSE events via sseService")
	}
}

func TestPipelineServiceEmitsStageStartedEvent(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "stage_started") {
		t.Error("pipeline.service.ts must emit 'stage_started' SSE event before each stage")
	}
}

func TestPipelineServiceEmitsStageCompletedEvent(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "stage_completed") {
		t.Error("pipeline.service.ts must emit 'stage_completed' SSE event after each stage")
	}
}

func TestPipelineServiceEmitsRunCompletedEvent(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "run_completed") {
		t.Error("pipeline.service.ts must emit 'run_completed' SSE event when pipeline finishes")
	}
}

func TestPipelineServiceInitializesRunStageRecords(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "createAll") {
		t.Error("pipeline.service.ts must call runStageModel.createAll to initialize stage records")
	}
}

func TestPipelineServiceCallsUpdateStageForTransitions(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "updateStage") {
		t.Error("pipeline.service.ts must call runStageModel.updateStage for stage status transitions")
	}
}

func TestPipelineServiceCallsUpdateStatusForRunTransitions(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	if !strings.Contains(src, "updateStatus") {
		t.Error("pipeline.service.ts must call validationRunModel.updateStatus for run transitions")
	}
}

func TestPipelineServiceEnqueueCallsExecuteAsync(t *testing.T) {
	src := readFile(t, "backend/src/services/pipeline.service.ts")
	// enqueue should call execute without await (fire-and-forget)
	if !strings.Contains(src, "execute") {
		t.Error("pipeline.service.ts enqueue must call execute to start pipeline execution")
	}
}

// ── stage-executor.service.ts ─────────────────────────────────────────────────

func TestStageExecutorServiceExportsExecuteStage(t *testing.T) {
	src := readFile(t, "backend/src/services/stage-executor.service.ts")
	if !strings.Contains(src, "executeStage") {
		t.Error("stage-executor.service.ts must export executeStage function")
	}
}

func TestStageExecutorServiceEmitsLogLines(t *testing.T) {
	src := readFile(t, "backend/src/services/stage-executor.service.ts")
	if !strings.Contains(src, "log_line") && !strings.Contains(src, "sseService") && !strings.Contains(src, "sse") {
		t.Error("stage-executor.service.ts must emit log lines via sseService")
	}
}

func TestStageExecutorServiceUpdatesStageStatus(t *testing.T) {
	src := readFile(t, "backend/src/services/stage-executor.service.ts")
	if !strings.Contains(src, "RUNNING") {
		t.Error("stage-executor.service.ts executeStage must update stage to RUNNING with startedAt")
	}
}

func TestStageExecutorServiceSetsCompletedAt(t *testing.T) {
	src := readFile(t, "backend/src/services/stage-executor.service.ts")
	if !strings.Contains(src, "completedAt") {
		t.Error("stage-executor.service.ts executeStage must set completedAt on stage completion")
	}
}

// ── pipeline.service.test.ts (TypeScript test coverage) ──────────────────────

func TestPipelineServiceTestFileImportsPipelineService(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "pipeline.service") {
		t.Error("pipeline.service.test.ts must import from pipeline.service")
	}
}

func TestPipelineServiceTestFileMocksRunStageModel(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "run-stage.model") {
		t.Error("pipeline.service.test.ts must mock run-stage.model")
	}
}

func TestPipelineServiceTestFileMocksValidationRunModel(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "validation-run.model") {
		t.Error("pipeline.service.test.ts must mock validation-run.model")
	}
}

func TestPipelineServiceTestFileMocksSseService(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "sse") {
		t.Error("pipeline.service.test.ts must mock sseService")
	}
}

func TestPipelineServiceTestFileTestsAllFiveStagesExecuted(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "5") && !strings.Contains(src, "STAGE_ORDER") {
		t.Error("pipeline.service.test.ts must test that all 5 stages are executed in STAGE_ORDER")
	}
}

func TestPipelineServiceTestFileTestsFailurePropagation(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "FAILED") {
		t.Error("pipeline.service.test.ts must test that a stage failure marks run as FAILED")
	}
}

func TestPipelineServiceTestFileTestsCancellation(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "isCancelled") {
		t.Error("pipeline.service.test.ts must test the cancellation token mechanism")
	}
}

func TestPipelineServiceTestFileTestsRunCompleted(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "run_completed") {
		t.Error("pipeline.service.test.ts must test that run_completed SSE event is emitted")
	}
}

func TestPipelineServiceTestFileTestsStageStartedEvent(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "stage_started") {
		t.Error("pipeline.service.test.ts must test that stage_started SSE events are emitted")
	}
}

func TestPipelineServiceTestFileTestsStageCompletedEvent(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "stage_completed") {
		t.Error("pipeline.service.test.ts must test that stage_completed SSE events are emitted")
	}
}

func TestPipelineServiceTestFileTestsUpdateStageRunning(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "RUNNING") {
		t.Error("pipeline.service.test.ts must verify updateStage is called with RUNNING status")
	}
}

func TestPipelineServiceTestFileTestsUpdateStageSuccess(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "SUCCESS") {
		t.Error("pipeline.service.test.ts must verify updateStage is called with SUCCESS status")
	}
}

func TestPipelineServiceTestFileTestsCancelledRunStatus(t *testing.T) {
	src := readFile(t, "backend/tests/unit/services/pipeline.service.test.ts")
	if !strings.Contains(src, "CANCELLED") {
		t.Error("pipeline.service.test.ts must test that run is marked CANCELLED when token is set")
	}
}
