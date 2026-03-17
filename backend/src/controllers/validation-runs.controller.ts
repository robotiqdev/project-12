import { Request, Response, NextFunction } from 'express';
import * as concurrencyService from '../services/concurrency.service';
import * as pipelineService from '../services/pipeline.service';

export async function create(req: Request, res: Response, next: NextFunction): Promise<void> {
  try {
    const { branchName, commitSha, config } = req.body;

    const lockResult = await concurrencyService.checkAndLockBranch(branchName);

    if (!lockResult.allowed) {
      res.status(409).json({
        error: 'BRANCH_LOCKED',
        conflictingRunId: lockResult.conflictingRunId,
      });
      return;
    }

    const run = await concurrencyService.createRun({
      branchName,
      commitSha,
      userId: '',
      config: config ?? {},
    });

    void pipelineService.enqueue(run.id).catch((err: unknown) => {
      console.error('pipeline enqueue error', err);
    });

    res.status(201).json(run);
  } catch (err) {
    next(err);
  }
}
