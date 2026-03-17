import * as sseService from './sse.service';
import { PipelineExecutionContext } from '../types/pipeline.types';

export interface StageResult {
  success: boolean;
}

export async function executeStage(
  ctx: PipelineExecutionContext,
  stageType: string,
): Promise<StageResult> {
  const { runId } = ctx;

  // Emit a log line indicating the stage is running (stubbed work)
  await sseService.emit(runId, 'log_line', {
    stage: stageType,
    message: `Executing stage ${stageType}`,
  });

  // Stubbed: simulate stage work
  // In production, each stageType would have real logic here

  // Update stage to RUNNING with startedAt (handled by pipeline.service)
  // Set completedAt on stage completion (handled by pipeline.service)
  const completedAt = new Date();
  void completedAt;

  return { success: true };
}
