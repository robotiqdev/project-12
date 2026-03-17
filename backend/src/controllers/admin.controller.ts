import { Request, Response } from 'express';
import * as cleanupService from '../services/cleanup.service';
import * as cleanupJobModel from '../models/cleanup-job.model';
import { CleanupResult } from '../types/cleanup.types';

export async function runAgeBasedCleanup(req: Request, res: Response): Promise<void> {
  const result: CleanupResult = await cleanupService.runAgeBased();
  res.status(200).json(result);
}

export async function runCountBasedCleanup(req: Request, res: Response): Promise<void> {
  const result: CleanupResult = await cleanupService.runCountBased();
  res.status(200).json(result);
}

export async function listCleanupJobs(req: Request, res: Response): Promise<void> {
  const jobs = await cleanupJobModel.list();
  res.status(200).json(jobs);
}
