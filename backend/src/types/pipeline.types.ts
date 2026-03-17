import { ValidationRunRecord } from './validation-run.types';

export interface CancellationToken {
  isCancelled: boolean;
}

export interface PipelineExecutionContext {
  runId: string;
  token: CancellationToken;
  run: ValidationRunRecord;
}

export type StageExecutorFn = (
  ctx: PipelineExecutionContext,
  stageType: string,
) => Promise<unknown>;
