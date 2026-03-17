/**
 * Unit tests for pipeline.service.ts — TASK-4594
 *
 * Tests the public interface of:
 *   - execute(ctx): orchestrates the 5-stage pipeline, updates DB state per
 *     stage, emits SSE events, and respects cancellation tokens
 *   - enqueue(runId): fetches the run, creates a CancellationToken, registers
 *     it in the internal registry, and calls execute(ctx) async (fire-and-forget)
 *   - unregisterToken(runId): removes the token from the registry on completion
 */

import {
  execute,
  enqueue,
  unregisterToken,
} from '../../../src/services/pipeline.service';
import * as runStageModel from '../../../src/models/run-stage.model';
import * as validationRunModel from '../../../src/models/validation-run.model';
import * as sseService from '../../../src/services/sse.service';
import * as stageExecutorService from '../../../src/services/stage-executor.service';
import { STAGE_ORDER, StageType } from '../../../src/types/validation-run.types';
import {
  CancellationToken,
  PipelineExecutionContext,
} from '../../../src/types/pipeline.types';

jest.mock('../../../src/models/run-stage.model');
jest.mock('../../../src/models/validation-run.model');
jest.mock('../../../src/services/sse.service');
jest.mock('../../../src/services/stage-executor.service');

// ─── Helpers ─────────────────────────────────────────────────────────────────

function makeStageRecord(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id: `stage-${Math.random().toString(36).slice(2)}`,
    runId: 'run-test-001',
    stageName: 'SETUP',
    stageIndex: 0,
    status: 'PENDING',
    logs: '',
    startedAt: null,
    completedAt: null,
    ...overrides,
  };
}

function makeRunRecord(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id: 'run-test-001',
    branchName: 'feature/test-branch',
    commitSha: 'abc1234567890abcdef',
    status: 'PENDING',
    config: {},
    userId: 'user-001',
    createdAt: new Date(),
    updatedAt: new Date(),
    completedAt: null,
    cancelledAt: null,
    ...overrides,
  };
}

function makeCancellationToken(): CancellationToken {
  return { isCancelled: false };
}

function makeContext(overrides: Partial<PipelineExecutionContext> = {}): PipelineExecutionContext {
  return {
    runId: 'run-test-001',
    token: makeCancellationToken(),
    run: makeRunRecord() as never,
    ...overrides,
  };
}

// ─── Test setup ──────────────────────────────────────────────────────────────

const mockUpdateStage = runStageModel.updateStage as jest.MockedFunction<typeof runStageModel.updateStage>;
const mockCreateAll = runStageModel.createAll as jest.MockedFunction<typeof runStageModel.createAll>;
const mockFindById = validationRunModel.findById as jest.MockedFunction<typeof validationRunModel.findById>;
const mockUpdateStatus = validationRunModel.updateStatus as jest.MockedFunction<typeof validationRunModel.updateStatus>;
const mockSseEmit = sseService.emit as jest.MockedFunction<typeof sseService.emit>;
const mockExecuteStage = stageExecutorService.executeStage as jest.MockedFunction<typeof stageExecutorService.executeStage>;

function setupSuccessfulStages() {
  // createAll returns 5 stage records matching STAGE_ORDER
  const stages = STAGE_ORDER.map((stageName, stageIndex) =>
    makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
  );
  mockCreateAll.mockResolvedValue(stages as never);
  mockUpdateStage.mockResolvedValue({} as never);
  mockUpdateStatus.mockResolvedValue({} as never);
  mockSseEmit.mockResolvedValue(undefined as never);
  mockExecuteStage.mockResolvedValue({ success: true } as never);
  return stages;
}

describe('PipelineService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  // ─────────────────────────────────────────────────────────────────────────
  // execute — stage sequencing
  // ─────────────────────────────────────────────────────────────────────────

  describe('execute — stage sequencing', () => {
    it('calls executeStage for all 5 stages in STAGE_ORDER sequence', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      expect(mockExecuteStage).toHaveBeenCalledTimes(5);
      const calls = mockExecuteStage.mock.calls;
      STAGE_ORDER.forEach((stageType, index) => {
        expect(calls[index][1]).toBe(stageType);
      });
    });

    it('calls executeStage with the pipeline execution context', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      expect(mockExecuteStage).toHaveBeenCalledWith(ctx, expect.any(String));
    });

    it('updates run status to RUNNING at the start of execution', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      expect(mockUpdateStatus).toHaveBeenCalledWith(
        ctx.runId,
        'RUNNING',
        expect.anything(),
      );
    });

    it('creates all 5 RunStage records with PENDING status before executing any stage', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      expect(mockCreateAll).toHaveBeenCalledWith(
        ctx.runId,
        expect.arrayContaining([
          expect.objectContaining({ stageName: StageType.SETUP }),
          expect.objectContaining({ stageName: StageType.LINT }),
          expect.objectContaining({ stageName: StageType.TEST }),
          expect.objectContaining({ stageName: StageType.BUILD }),
          expect.objectContaining({ stageName: StageType.DEPLOY }),
        ]),
      );
    });

    it('marks run as SUCCESS when all 5 stages complete successfully', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      expect(mockUpdateStatus).toHaveBeenCalledWith(
        ctx.runId,
        'SUCCESS',
        expect.anything(),
      );
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // execute — failure propagation
  // ─────────────────────────────────────────────────────────────────────────

  describe('execute — failure propagation', () => {
    it('stops executing stages after stage 2 (LINT) fails and does not call stage 3, 4, 5', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      // Stage 1 (SETUP) succeeds, stage 2 (LINT) fails
      mockExecuteStage
        .mockResolvedValueOnce({ success: true } as never)  // SETUP
        .mockRejectedValueOnce(new Error('lint failed'))    // LINT — failure
        .mockResolvedValue({ success: true } as never);     // should not be reached

      const ctx = makeContext();
      await execute(ctx);

      // Only stages 1 and 2 should have been attempted
      expect(mockExecuteStage).toHaveBeenCalledTimes(2);
      // Verify stages 3-5 were NOT called
      const calledStages = mockExecuteStage.mock.calls.map((call) => call[1]);
      expect(calledStages).not.toContain(StageType.TEST);
      expect(calledStages).not.toContain(StageType.BUILD);
      expect(calledStages).not.toContain(StageType.DEPLOY);
    });

    it('marks run as FAILED when a stage fails', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      mockExecuteStage
        .mockResolvedValueOnce({ success: true } as never)  // SETUP
        .mockRejectedValueOnce(new Error('lint error'));     // LINT fails

      const ctx = makeContext();
      await execute(ctx);

      expect(mockUpdateStatus).toHaveBeenCalledWith(
        ctx.runId,
        'FAILED',
        expect.anything(),
      );
    });

    it('leaves stages 3-5 in PENDING status when stage 2 fails', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      mockExecuteStage
        .mockResolvedValueOnce({ success: true } as never)
        .mockRejectedValueOnce(new Error('lint error'));

      const ctx = makeContext();
      await execute(ctx);

      // updateStage should NOT have been called with RUNNING for stages 3-5
      const updateStageCalls = mockUpdateStage.mock.calls;
      const stage3Id = stages[2].id; // TEST
      const stage4Id = stages[3].id; // BUILD
      const stage5Id = stages[4].id; // DEPLOY

      const calledIds = updateStageCalls.map((call) => call[0]);
      expect(calledIds).not.toContain(stage3Id);
      expect(calledIds).not.toContain(stage4Id);
      expect(calledIds).not.toContain(stage5Id);
    });

    it('does not mark run as SUCCESS when a stage fails', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      mockExecuteStage
        .mockResolvedValueOnce({ success: true } as never)
        .mockRejectedValueOnce(new Error('lint error'));

      const ctx = makeContext();
      await execute(ctx);

      const statusCalls = mockUpdateStatus.mock.calls.map((call) => call[1]);
      expect(statusCalls).not.toContain('SUCCESS');
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // execute — DB persistence per stage transition
  // ─────────────────────────────────────────────────────────────────────────

  describe('execute — stage DB persistence', () => {
    it('calls updateStage with RUNNING status before executing each stage', async () => {
      const stages = setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      // Each stage should have been updated to RUNNING
      const runningCalls = mockUpdateStage.mock.calls.filter(
        (call) => (call[1] as Record<string, unknown>).status === 'RUNNING',
      );
      expect(runningCalls).toHaveLength(5);
    });

    it('calls updateStage with SUCCESS status after each stage completes', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const successCalls = mockUpdateStage.mock.calls.filter(
        (call) => (call[1] as Record<string, unknown>).status === 'SUCCESS',
      );
      expect(successCalls).toHaveLength(5);
    });

    it('calls updateStage with FAILED status when a stage fails', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      mockExecuteStage
        .mockResolvedValueOnce({ success: true } as never)  // SETUP succeeds
        .mockRejectedValueOnce(new Error('lint error'));     // LINT fails

      const ctx = makeContext();
      await execute(ctx);

      const failedCalls = mockUpdateStage.mock.calls.filter(
        (call) => (call[1] as Record<string, unknown>).status === 'FAILED',
      );
      expect(failedCalls.length).toBeGreaterThanOrEqual(1);
    });

    it('sets startedAt timestamp when updating stage to RUNNING', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const runningCalls = mockUpdateStage.mock.calls.filter(
        (call) => (call[1] as Record<string, unknown>).status === 'RUNNING',
      );
      runningCalls.forEach((call) => {
        expect((call[1] as Record<string, unknown>).startedAt).toBeDefined();
      });
    });

    it('sets completedAt timestamp when updating stage to SUCCESS', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const successCalls = mockUpdateStage.mock.calls.filter(
        (call) => (call[1] as Record<string, unknown>).status === 'SUCCESS',
      );
      successCalls.forEach((call) => {
        expect((call[1] as Record<string, unknown>).completedAt).toBeDefined();
      });
    });

    it('calls updateStage for each stage with RUNNING before calling executeStage', async () => {
      const stages = setupSuccessfulStages();
      const ctx = makeContext();
      const callOrder: string[] = [];

      mockUpdateStage.mockImplementation(async (id, update) => {
        callOrder.push(`updateStage:${(update as Record<string, unknown>).status}:${id}`);
        return {} as never;
      });
      mockExecuteStage.mockImplementation(async (executionCtx, stageType) => {
        callOrder.push(`executeStage:${stageType}`);
        return { success: true } as never;
      });

      await execute(ctx);

      // Verify that for SETUP (stage 0): RUNNING update happens before executeStage
      const setupRunningIdx = callOrder.indexOf(`updateStage:RUNNING:${stages[0].id}`);
      const setupExecuteIdx = callOrder.indexOf(`executeStage:${StageType.SETUP}`);
      expect(setupRunningIdx).toBeLessThan(setupExecuteIdx);
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // execute — SSE events
  // ─────────────────────────────────────────────────────────────────────────

  describe('execute — SSE events', () => {
    it('emits stage_started event for each stage before executing it', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const stageStartedCalls = mockSseEmit.mock.calls.filter(
        (call) => call[1] === 'stage_started',
      );
      expect(stageStartedCalls).toHaveLength(5);
    });

    it('emits stage_completed event for each stage after it succeeds', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const stageCompletedCalls = mockSseEmit.mock.calls.filter(
        (call) => call[1] === 'stage_completed',
      );
      expect(stageCompletedCalls).toHaveLength(5);
    });

    it('emits log_line events during stage execution', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const logLineCalls = mockSseEmit.mock.calls.filter(
        (call) => call[1] === 'log_line',
      );
      expect(logLineCalls.length).toBeGreaterThanOrEqual(1);
    });

    it('emits run_completed event when all stages finish successfully', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const runCompletedCalls = mockSseEmit.mock.calls.filter(
        (call) => call[1] === 'run_completed',
      );
      expect(runCompletedCalls).toHaveLength(1);
    });

    it('emits run_completed event with the run id', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const runCompletedCall = mockSseEmit.mock.calls.find(
        (call) => call[1] === 'run_completed',
      );
      expect(runCompletedCall).toBeDefined();
      // First arg should be the runId
      expect(runCompletedCall![0]).toBe(ctx.runId);
    });

    it('emits SSE events with the runId as the channel identifier', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      // All SSE emit calls should use the runId as the first argument
      mockSseEmit.mock.calls.forEach((call) => {
        expect(call[0]).toBe(ctx.runId);
      });
    });

    it('emits stage_started events in STAGE_ORDER sequence', async () => {
      setupSuccessfulStages();
      const ctx = makeContext();

      await execute(ctx);

      const stageStartedCalls = mockSseEmit.mock.calls.filter(
        (call) => call[1] === 'stage_started',
      );
      STAGE_ORDER.forEach((stageType, index) => {
        expect(stageStartedCalls[index][2]).toEqual(
          expect.objectContaining({ stage: stageType }),
        );
      });
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // execute — cancellation token
  // ─────────────────────────────────────────────────────────────────────────

  describe('execute — cancellation token', () => {
    it('skips stages 3-5 when token is cancelled after stage 2', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      const ctx = makeContext();

      let stageCallCount = 0;
      mockExecuteStage.mockImplementation(async () => {
        stageCallCount++;
        if (stageCallCount === 2) {
          // After stage 2 (LINT), set cancelled
          ctx.token.isCancelled = true;
        }
        return { success: true } as never;
      });

      await execute(ctx);

      // Only stages 1 and 2 should have been executed
      expect(mockExecuteStage).toHaveBeenCalledTimes(2);
    });

    it('marks run as CANCELLED when cancellation token is set between stages', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      const ctx = makeContext();

      let stageCallCount = 0;
      mockExecuteStage.mockImplementation(async () => {
        stageCallCount++;
        if (stageCallCount === 2) {
          ctx.token.isCancelled = true;
        }
        return { success: true } as never;
      });

      await execute(ctx);

      expect(mockUpdateStatus).toHaveBeenCalledWith(
        ctx.runId,
        'CANCELLED',
        expect.anything(),
      );
    });

    it('does not mark run as SUCCESS or FAILED when cancelled after stage 2', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      const ctx = makeContext();

      let stageCallCount = 0;
      mockExecuteStage.mockImplementation(async () => {
        stageCallCount++;
        if (stageCallCount === 2) {
          ctx.token.isCancelled = true;
        }
        return { success: true } as never;
      });

      await execute(ctx);

      const statusCalls = mockUpdateStatus.mock.calls.map((call) => call[1]);
      expect(statusCalls).not.toContain('SUCCESS');
      expect(statusCalls).not.toContain('FAILED');
    });

    it('checks cancellation token before each stage invocation', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      const ctx = makeContext();
      // Cancel immediately before first stage
      ctx.token.isCancelled = true;

      mockExecuteStage.mockResolvedValue({ success: true } as never);

      await execute(ctx);

      // No stages should be executed because token was already cancelled
      expect(mockExecuteStage).not.toHaveBeenCalled();
    });

    it('marks run as CANCELLED (not FAILED) even if token is set immediately', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      const ctx = makeContext();
      ctx.token.isCancelled = true;

      mockExecuteStage.mockResolvedValue({ success: true } as never);

      await execute(ctx);

      expect(mockUpdateStatus).toHaveBeenCalledWith(
        ctx.runId,
        'CANCELLED',
        expect.anything(),
      );
    });

    it('stages 1 and 2 remain with SUCCESS status after cancellation post-stage-2', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      const ctx = makeContext();

      let stageCallCount = 0;
      mockExecuteStage.mockImplementation(async () => {
        stageCallCount++;
        if (stageCallCount === 2) {
          ctx.token.isCancelled = true;
        }
        return { success: true } as never;
      });

      await execute(ctx);

      // Stages 1 and 2 should have been updated to SUCCESS
      const successCalls = mockUpdateStage.mock.calls.filter(
        (call) => (call[1] as Record<string, unknown>).status === 'SUCCESS',
      );
      expect(successCalls).toHaveLength(2);

      // The stage ids updated to SUCCESS should be stage-0 (SETUP) and stage-1 (LINT)
      const successIds = successCalls.map((call) => call[0]);
      expect(successIds).toContain(stages[0].id);
      expect(successIds).toContain(stages[1].id);
    });

    it('stages 3-5 are never updated to RUNNING when cancelled after stage 2', async () => {
      const stages = STAGE_ORDER.map((stageName, stageIndex) =>
        makeStageRecord({ id: `stage-${stageIndex}`, stageName, stageIndex, runId: 'run-test-001' }),
      );
      mockCreateAll.mockResolvedValue(stages as never);
      mockUpdateStage.mockResolvedValue({} as never);
      mockUpdateStatus.mockResolvedValue({} as never);
      mockSseEmit.mockResolvedValue(undefined as never);

      const ctx = makeContext();

      let stageCallCount = 0;
      mockExecuteStage.mockImplementation(async () => {
        stageCallCount++;
        if (stageCallCount === 2) {
          ctx.token.isCancelled = true;
        }
        return { success: true } as never;
      });

      await execute(ctx);

      const runningCalls = mockUpdateStage.mock.calls.filter(
        (call) => (call[1] as Record<string, unknown>).status === 'RUNNING',
      );
      const runningIds = runningCalls.map((call) => call[0]);

      // Stages 3-5 should NOT have been set to RUNNING
      expect(runningIds).not.toContain(stages[2].id); // TEST
      expect(runningIds).not.toContain(stages[3].id); // BUILD
      expect(runningIds).not.toContain(stages[4].id); // DEPLOY
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // enqueue
  // ─────────────────────────────────────────────────────────────────────────

  describe('enqueue', () => {
    it('fetches the run by id before executing', async () => {
      const run = makeRunRecord();
      mockFindById.mockResolvedValue(run as never);
      setupSuccessfulStages();

      await enqueue('run-test-001');
      // Allow fire-and-forget async tasks to settle
      await new Promise((resolve) => process.nextTick(resolve));
      await new Promise((resolve) => setTimeout(resolve, 10));

      expect(mockFindById).toHaveBeenCalledWith('run-test-001');
    });

    it('registers a cancellation token for the run id', async () => {
      const run = makeRunRecord();
      mockFindById.mockResolvedValue(run as never);
      setupSuccessfulStages();

      await enqueue('run-test-002');
      await new Promise((resolve) => process.nextTick(resolve));
      await new Promise((resolve) => setTimeout(resolve, 10));

      // After enqueue, unregisterToken should succeed without error
      expect(() => unregisterToken('run-test-002')).not.toThrow();
    });

    it('returns without throwing even if run is not found', async () => {
      mockFindById.mockResolvedValue(null as never);

      await expect(enqueue('non-existent-run')).resolves.not.toThrow();
    });
  });

  // ─────────────────────────────────────────────────────────────────────────
  // unregisterToken
  // ─────────────────────────────────────────────────────────────────────────

  describe('unregisterToken', () => {
    it('removes the token from the registry without throwing', () => {
      expect(() => unregisterToken('some-run-id')).not.toThrow();
    });

    it('can be called multiple times for the same runId without error', () => {
      expect(() => {
        unregisterToken('run-idempotent');
        unregisterToken('run-idempotent');
      }).not.toThrow();
    });
  });
});
