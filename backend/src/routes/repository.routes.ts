/**
 * Repository Routes - TASK-4600
 *
 * Defines GET routes for the repository integration API,
 * scoped under /api/repository.
 *
 * Routes:
 *   GET /api/repository/branches
 *   GET /api/repository/branches/:branch/commits
 *   GET /api/repository/context
 */

import { Router } from 'express';
import {
  getBranches,
  getBranchCommits,
  getContext,
} from '../controllers/repository.controller';

const router = Router();

// GET /api/repository/branches
router.get('/branches', getBranches);

// GET /api/repository/branches/:branch/commits
router.get('/branches/:branch/commits', getBranchCommits);

// GET /api/repository/context
router.get('/context', getContext);

export default router;
