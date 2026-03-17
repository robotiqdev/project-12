import { Request, Response, NextFunction } from 'express';
import * as concurrencyService from '../services/concurrency.service';
import * as pipelineService from '../services/pipeline.service';
import * as validationRunModel from '../models/validation-run.model';
import { ValidationRunStatus } from '../types/validation-run.types';

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

export async function list(req: Request, res: Response, next: NextFunction): Promise<void> {
  try {
    const { branch, status, page: pageParam, limit: limitParam } = req.query;

    const page = Math.max(1, parseInt(String(pageParam ?? '1'), 10) || 1);
    const limit = Math.min(100, Math.max(1, parseInt(String(limitParam ?? '20'), 10) || 20));

    const { data, total } = await validationRunModel.list({
      branchName: branch ? String(branch) : undefined,
      status: status ? (String(status) as ValidationRunStatus) : undefined,
      page,
      pageSize: limit,
    });

    res.status(200).json({ data, total, page, limit });
  } catch (err) {
    next(err);
  }
}

export async function getById(req: Request, res: Response, next: NextFunction): Promise<void> {
  try {
    const { id } = req.params;

    const result = await validationRunModel.findByIdWithStages(id);
    if (!result) {
      res.status(404).json({ error: 'NOT_FOUND' });
      return;
    }

    const { stages, ...run } = result as { stages: unknown[]; [key: string]: unknown };
    res.status(200).json({ run, stages });
  } catch (err) {
    next(err);
  }
}
