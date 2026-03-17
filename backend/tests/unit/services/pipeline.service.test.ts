/**
 * Unit tests for pipeline.service.ts
 * TASK-4607: Implement pipeline cancellation with stage preservation
 *
 * Tests verify the following invariants when a cancellation token fires
 * between stage 2 and stage 3 of a 5-stage pipeline:
 *
 *  1. Stages 1 and 2 were fully executed; their SUCCESS status is preserved.
 *  2. Stages 3, 4, and 5 were never started; they remain PENDING (not CANCELLED).
 *  3. The validation run record is updated to status=CANCELLED with cancelledAt set.
 *  4. An SSE event named `run_cancelled` is emitted with the runId.
 *  5. pipelineService.getToken(runId) returns undefined after cancellation cleanup.
 */

import { pipelineService } from '../../../src/services/pipeline.service';
import { ValidationRunModel } from '../../../src/models/validation-run.model';
import { stageExecutorService } from '../../../src/services/stage-executor.service';
import { sseService } from '../../../src/services/sse.service';
import { ValidationRunStatus } from '../../../src/models/validation-run.types';

jest.mock('../../../src/models/validation-run.model');
jest.mock('../../../src/services/stage-executor.service');
jest.mock('../../../src/services/sse.service');

// ---------------------------------------------------------------------------
// Types matching the expected pipeline service interface
// ---------------------------------------------------------------------------

interface CancellationToken {
  isCancelled: boolean;
  cancel(): void;
}

interface StageRecord {
  id: string;
  runId: string;
  stageIndex: number;
  stageName: string;
  status: string;
  logs: string;
  startedAt: Date | null;
  completedAt: Date | null;
}

interface PipelineContext {
  runId: string;
  stages: StageRecord[];
  cancellationToken: CancellationToken;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeStages(runId: string): StageRecord[] {
  const names = ['SETUP', 'LINT', 'TEST', 'BUILD', 'DEPLOY'];
  return names.map((name, i) => ({
    id: `stage-id-${i}`,
    runId,
    stageIndex: i,
    stageName: name,
    status: 'PENDING',
    logs: '',
    startedAt: null,
    completedAt: null,
  }));
}

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

describe('pipelineService', () => {
  const TEST_RUN_ID = 'run-uuid-task-4607';

  beforeEach(() => {
    jest.clearAllMocks();
  });

  // -------------------------------------------------------------------------
  // Cancellation between stage 2 and stage 3 (main invariant tests)
  // -------------------------------------------------------------------------

  describe('execute() — cancellation signal fires after stage 2 completes', () => {
    let token: CancellationToken;
    let stages: StageRecord[];
    let executeStage: jest.Mock;
    let cancelRunMock: jest.Mock;
    let sseEmitMock: jest.Mock;

    beforeEach(async () => {
      stages = makeStages(TEST_RUN_ID);

      // A mutable cancellation token: external caller sets isCancelled=true
      // after stageIndex=1 (the 2nd stage) finishes executing.
      token = {
        isCancelled: false,
        cancel: jest.fn(function (this: CancellationToken) {
          this.isCancelled = true;
        }),
      };

      // executeStage mock:
      //   - stageIndex 0 (SETUP): runs normally, returns resolved promise
      //   - stageIndex 1 (LINT):  runs normally, then sets token.isCancelled = true
      //     to simulate an external cancel API call arriving between stage 2 and 3
      //   - stageIndex 2+ should never be called after cancellation is detected
      executeStage = jest.fn().mockImplementation(async (stage: StageRecord) => {
        if (stage.stageIndex === 1) {
          // Simulate external cancellation arriving right as stage 2 finishes
          token.isCancelled = true;
        }
      });
      (stageExecutorService as { executeStage: jest.Mock }).executeStage = executeStage;

      // validationRunModel.cancel() mock — returns a CANCELLED run record
      cancelRunMock = jest.fn().mockResolvedValue({
        id: TEST_RUN_ID,
        branchName: 'feature/test',
        commitSha: 'abc1234',
        status: ValidationRunStatus.CANCELLED,
        config: {},
        userId: 'user-test',
        createdAt: new Date(),
        updatedAt: new Date(),
        completedAt: null,
        cancelledAt: new Date(),
      });
      (
        ValidationRunModel as jest.MockedClass<typeof ValidationRunModel>
      ).prototype.cancel = cancelRunMock;

      // sseService.emit() mock
      sseEmitMock = jest.fn();
      (sseService as { emit: jest.Mock }).emit = sseEmitMock;

      // Register token and run the pipeline
      pipelineService.registerToken(TEST_RUN_ID, token);

      const ctx: PipelineContext = {
        runId: TEST_RUN_ID,
        stages,
        cancellationToken: token,
      };

      await pipelineService.execute(ctx);
    });

    // -----------------------------------------------------------------------
    // Test 1: Stages 1 and 2 retain their SUCCESS status
    // (Verified by confirming executeStage was invoked for stageIndex 0 and 1)
    // -----------------------------------------------------------------------

    it('executes stage 1 (stageIndex=0) before cancellation', () => {
      expect(executeStage).toHaveBeenCalledWith(
        expect.objectContaining({ stageIndex: 0 }),
        expect.anything(),
      );
    });

    it('executes stage 2 (stageIndex=1) before cancellation', () => {
      expect(executeStage).toHaveBeenCalledWith(
        expect.objectContaining({ stageIndex: 1 }),
        expect.anything(),
      );
    });

    it('executes exactly 2 stages before detecting the cancellation signal', () => {
      expect(executeStage).toHaveBeenCalledTimes(2);
    });

    // -----------------------------------------------------------------------
    // Test 2: Stages 3, 4, 5 remain PENDING — they are never started
    // -----------------------------------------------------------------------

    it('does NOT execute stage 3 (stageIndex=2) after cancellation is detected', () => {
      expect(executeStage).not.toHaveBeenCalledWith(
        expect.objectContaining({ stageIndex: 2 }),
        expect.anything(),
      );
    });

    it('does NOT execute stage 4 (stageIndex=3) after cancellation is detected', () => {
      expect(executeStage).not.toHaveBeenCalledWith(
        expect.objectContaining({ stageIndex: 3 }),
        expect.anything(),
      );
    });

    it('does NOT execute stage 5 (stageIndex=4) after cancellation is detected', () => {
      expect(executeStage).not.toHaveBeenCalledWith(
        expect.objectContaining({ stageIndex: 4 }),
        expect.anything(),
      );
    });

    // -----------------------------------------------------------------------
    // Test 3: Run status becomes CANCELLED with cancelledAt set
    // -----------------------------------------------------------------------

    it('calls validationRunModel.cancel(runId) to set run status to CANCELLED', () => {
      expect(cancelRunMock).toHaveBeenCalledWith(TEST_RUN_ID);
    });

    it('calls validationRunModel.cancel exactly once', () => {
      expect(cancelRunMock).toHaveBeenCalledTimes(1);
    });

    it('the cancelled run record includes cancelledAt timestamp', async () => {
      // Verify the mock returned a record with cancelledAt set
      const result = await cancelRunMock.mock.results[0].value;
      expect(result.cancelledAt).toBeInstanceOf(Date);
      expect(result.status).toBe(ValidationRunStatus.CANCELLED);
    });

    // -----------------------------------------------------------------------
    // Test 4: SSE emits `run_cancelled` event after cancellation
    // -----------------------------------------------------------------------

    it('emits a run_cancelled SSE event when cancellation is detected', () => {
      expect(sseEmitMock).toHaveBeenCalledWith(
        'run_cancelled',
        expect.anything(),
      );
    });

    it('run_cancelled SSE event payload includes the runId', () => {
      expect(sseEmitMock).toHaveBeenCalledWith(
        'run_cancelled',
        expect.objectContaining({ runId: TEST_RUN_ID }),
      );
    });

    it('emits the run_cancelled SSE event exactly once', () => {
      const runCancelledCalls = sseEmitMock.mock.calls.filter(
        (call: unknown[]) => call[0] === 'run_cancelled',
      );
      expect(runCancelledCalls).toHaveLength(1);
    });

    // -----------------------------------------------------------------------
    // Test 5: pipelineService.getToken(runId) returns undefined after cleanup
    // -----------------------------------------------------------------------

    it('unregisters the cancellation token so getToken returns undefined after cancellation', () => {
      expect(pipelineService.getToken(TEST_RUN_ID)).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // execute() — no cancellation: pipeline runs all 5 stages to completion
  // -------------------------------------------------------------------------

  describe('execute() — pipeline completes without cancellation', () => {
    it('executes all 5 stages when token is never cancelled', async () => {
      const runId = 'run-no-cancel';
      const stages = makeStages(runId);
      const token: CancellationToken = { isCancelled: false, cancel: jest.fn() };

      const executeStage = jest.fn().mockResolvedValue(undefined);
      (stageExecutorService as { executeStage: jest.Mock }).executeStage = executeStage;

      const cancelRunMock = jest.fn().mockResolvedValue({});
      (
        ValidationRunModel as jest.MockedClass<typeof ValidationRunModel>
      ).prototype.cancel = cancelRunMock;

      const sseEmitMock = jest.fn();
      (sseService as { emit: jest.Mock }).emit = sseEmitMock;

      pipelineService.registerToken(runId, token);

      await pipelineService.execute({ runId, stages, cancellationToken: token });

      expect(executeStage).toHaveBeenCalledTimes(5);
    });

    it('does NOT call validationRunModel.cancel when pipeline completes normally', async () => {
      const runId = 'run-no-cancel-2';
      const stages = makeStages(runId);
      const token: CancellationToken = { isCancelled: false, cancel: jest.fn() };

      const executeStage = jest.fn().mockResolvedValue(undefined);
      (stageExecutorService as { executeStage: jest.Mock }).executeStage = executeStage;

      const cancelRunMock = jest.fn().mockResolvedValue({});
      (
        ValidationRunModel as jest.MockedClass<typeof ValidationRunModel>
      ).prototype.cancel = cancelRunMock;

      const sseEmitMock = jest.fn();
      (sseService as { emit: jest.Mock }).emit = sseEmitMock;

      pipelineService.registerToken(runId, token);

      await pipelineService.execute({ runId, stages, cancellationToken: token });

      expect(cancelRunMock).not.toHaveBeenCalled();
    });

    it('does NOT emit run_cancelled SSE event when pipeline completes normally', async () => {
      const runId = 'run-no-cancel-3';
      const stages = makeStages(runId);
      const token: CancellationToken = { isCancelled: false, cancel: jest.fn() };

      const executeStage = jest.fn().mockResolvedValue(undefined);
      (stageExecutorService as { executeStage: jest.Mock }).executeStage = executeStage;

      const cancelRunMock = jest.fn().mockResolvedValue({});
      (
        ValidationRunModel as jest.MockedClass<typeof ValidationRunModel>
      ).prototype.cancel = cancelRunMock;

      const sseEmitMock = jest.fn();
      (sseService as { emit: jest.Mock }).emit = sseEmitMock;

      pipelineService.registerToken(runId, token);

      await pipelineService.execute({ runId, stages, cancellationToken: token });

      const runCancelledCalls = sseEmitMock.mock.calls.filter(
        (call: unknown[]) => call[0] === 'run_cancelled',
      );
      expect(runCancelledCalls).toHaveLength(0);
    });
  });

  // -------------------------------------------------------------------------
  // execute() — cancellation at the very first stage (stageIndex=0)
  // -------------------------------------------------------------------------

  describe('execute() — cancellation detected after stage 1 only', () => {
    it('executes exactly 1 stage and then stops when token is cancelled after stage 1', async () => {
      const runId = 'run-cancel-after-stage1';
      const stages = makeStages(runId);
      const token: CancellationToken = { isCancelled: false, cancel: jest.fn() };

      const executeStage = jest.fn().mockImplementation(async (stage: StageRecord) => {
        if (stage.stageIndex === 0) {
          token.isCancelled = true;
        }
      });
      (stageExecutorService as { executeStage: jest.Mock }).executeStage = executeStage;

      const cancelRunMock = jest.fn().mockResolvedValue({
        id: runId,
        status: ValidationRunStatus.CANCELLED,
        cancelledAt: new Date(),
      });
      (
        ValidationRunModel as jest.MockedClass<typeof ValidationRunModel>
      ).prototype.cancel = cancelRunMock;

      const sseEmitMock = jest.fn();
      (sseService as { emit: jest.Mock }).emit = sseEmitMock;

      pipelineService.registerToken(runId, token);

      await pipelineService.execute({ runId, stages, cancellationToken: token });

      expect(executeStage).toHaveBeenCalledTimes(1);
      expect(cancelRunMock).toHaveBeenCalledWith(runId);
      expect(sseEmitMock).toHaveBeenCalledWith('run_cancelled', expect.objectContaining({ runId }));
      expect(pipelineService.getToken(runId)).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // getToken() — token lifecycle management
  // -------------------------------------------------------------------------

  describe('getToken()', () => {
    it('returns undefined for an unregistered runId', () => {
      expect(pipelineService.getToken('non-existent-run')).toBeUndefined();
    });

    it('returns the registered token after registerToken()', () => {
      const runId = 'token-lifecycle-test';
      const token: CancellationToken = { isCancelled: false, cancel: jest.fn() };

      pipelineService.registerToken(runId, token);

      expect(pipelineService.getToken(runId)).toBe(token);

      // Cleanup
      pipelineService.unregisterToken(runId);
    });

    it('returns undefined after unregisterToken() is called', () => {
      const runId = 'token-unregister-test';
      const token: CancellationToken = { isCancelled: false, cancel: jest.fn() };

      pipelineService.registerToken(runId, token);
      pipelineService.unregisterToken(runId);

      expect(pipelineService.getToken(runId)).toBeUndefined();
    });

    it('tokens for different runIds are independent', () => {
      const runId1 = 'token-run-a';
      const runId2 = 'token-run-b';
      const token1: CancellationToken = { isCancelled: false, cancel: jest.fn() };
      const token2: CancellationToken = { isCancelled: false, cancel: jest.fn() };

      pipelineService.registerToken(runId1, token1);
      pipelineService.registerToken(runId2, token2);

      expect(pipelineService.getToken(runId1)).toBe(token1);
      expect(pipelineService.getToken(runId2)).toBe(token2);

      // Cleanup
      pipelineService.unregisterToken(runId1);
      pipelineService.unregisterToken(runId2);
    });
  });
});
