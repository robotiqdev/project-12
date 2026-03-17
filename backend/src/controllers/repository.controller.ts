/**
 * Repository Controller - TASK-4600
 *
 * Handles HTTP requests for repository integration endpoints:
 *   GET /api/repository/branches
 *   GET /api/repository/branches/:branch/commits
 *   GET /api/repository/context
 */

import { Request, Response } from 'express';
import { repositoryService } from '../services/repository.service';

/**
 * GET /api/repository/branches
 * Returns array of { name, lastCommit, isDefault }
 */
export async function getBranches(req: Request, res: Response): Promise<void> {
  try {
    const branches = await repositoryService.listBranches();
    res.json(branches);
  } catch (err: any) {
    const status = err.status || 500;
    res.status(status).json({ error: err.message || 'Internal server error' });
  }
}

/**
 * GET /api/repository/branches/:branch/commits
 * Returns last 20 commits with { sha, message, author, date }
 * Returns 404 for invalid/nonexistent branch names.
 */
export async function getBranchCommits(req: Request, res: Response): Promise<void> {
  const branchName = req.params.branch;

  try {
    const commits = await repositoryService.listCommits(branchName, 20);
    res.json(commits);
  } catch (err: any) {
    const status = err.status || 500;
    res.status(status).json({ error: err.message || 'Internal server error' });
  }
}

/**
 * GET /api/repository/context
 * Returns { baseUrl, provider }
 */
export async function getContext(req: Request, res: Response): Promise<void> {
  try {
    const context = await repositoryService.getContextBaseUrl();
    res.json({ baseUrl: context.baseUrl, provider: context.provider });
  } catch (err: any) {
    const status = err.status || 500;
    res.status(status).json({ error: err.message || 'Internal server error' });
  }
}
