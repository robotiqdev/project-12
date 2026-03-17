import * as runStageModel from '../models/run-stage.model';
import * as validationRunModel from '../models/validation-run.model';
import * as sseService from './sse.service';
import * as stageExecutorService from './stage-executor.service';
import { STAGE_ORDER } from '../types/validation-run.types';
import { CancellationToken, PipelineExecutionContext } from '../types/pipeline.types';

// Registry mapping runId -> CancellationToken for in-flight pipeline runs
const tokenRegistry = new Map<string, CancellationToken>();

export async function execute(ctx: PipelineExecutionContext): Promise<void> {
  const { runId, token } = ctx;

  // Update run to RUNNING
  await validationRunModel.updateStatus(runId, 'RUNNING', {});

  // Initialize all 5 RunStage records in PENDING status
  const stages = await runStageModel.createAll(
    runId,
    STAGE_ORDER.map((stageName, stageIndex) => ({ stageName, stageIndex })),
  );

  let failed = false;

  for (let i = 0; i < STAGE_ORDER.length; i++) {
    const stageType = STAGE_ORDER[i];
    const stage = stages[i];

    // Check cancellation token before each stage
    if (token.isCancelled) {
      await validationRunModel.updateStatus(runId, 'CANCELLED', {
        cancelledAt: new Date(),
      });
      unregisterToken(runId);
      return;
    }

    // Update stage to RUNNING with startedAt
    await runStageModel.updateStage(stage.id, {
      status: 'RUNNING',
      startedAt: new Date(),
    });

    // Emit stage_started event
    await sseService.emit(runId, 'stage_started', { stage: stageType });

    try {
      await stageExecutorService.executeStage(ctx, stageType);

      // Emit log_line for this stage
      await sseService.emit(runId, 'log_line', {
        stage: stageType,
        message: `Stage ${stageType} finished`,
      });

      // Update stage to SUCCESS with completedAt
      await runStageModel.updateStage(stage.id, {
        status: 'SUCCESS',
        completedAt: new Date(),
      });

      // Emit stage_completed event
      await sseService.emit(runId, 'stage_completed', { stage: stageType });
    } catch (err) {
      // Update stage to FAILED with completedAt
      await runStageModel.updateStage(stage.id, {
        status: 'FAILED',
        completedAt: new Date(),
      });
      failed = true;
      break;
    }
  }

  if (token.isCancelled) {
    await validationRunModel.updateStatus(runId, 'CANCELLED', {
      cancelledAt: new Date(),
    });
  } else if (failed) {
    await validationRunModel.updateStatus(runId, 'FAILED', {
      completedAt: new Date(),
    });
  } else {
    await validationRunModel.updateStatus(runId, 'SUCCESS', {
      completedAt: new Date(),
    });
    await sseService.emit(runId, 'run_completed', { runId });
  }

  unregisterToken(runId);
}

export async function enqueue(runId: string): Promise<void> {
  const run = await validationRunModel.findById(runId);
  if (!run) return;

  const token: CancellationToken = { isCancelled: false };
  tokenRegistry.set(runId, token);

  const ctx: PipelineExecutionContext = { runId, token, run };
  // Fire-and-forget: do not await execute
  execute(ctx);
}

export function unregisterToken(runId: string): void {
  tokenRegistry.delete(runId);
}
