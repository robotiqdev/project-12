import { Router } from 'express';
import { z } from 'zod';
import { validateBody } from '../middleware/validate.middleware';
import * as validationRunsController from '../controllers/validation-runs.controller';

export const createRunSchema = z.object({
  branchName: z.string().min(1).max(255),
  commitSha: z.string().regex(/^[0-9a-f]{40}$/),
  config: z
    .object({
      envVars: z.record(z.string()),
      featureFlags: z.record(z.boolean()),
    })
    .optional(),
});

const router = Router();

router.post('/', validateBody(createRunSchema), validationRunsController.create);
router.get('/', validationRunsController.list);
router.get('/:id', validationRunsController.getById);

export default router;
