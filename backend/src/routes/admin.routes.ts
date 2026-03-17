import { Router } from 'express';
import { adminAuthMiddleware } from '../middleware/auth.middleware';
import { runAgeBasedCleanup, runCountBasedCleanup, listCleanupJobs } from '../controllers/admin.controller';

const router = Router();

router.use(adminAuthMiddleware);

router.post('/cleanup/age', runAgeBasedCleanup);
router.post('/cleanup/count', runCountBasedCleanup);
router.get('/cleanup/jobs', listCleanupJobs);

export default router;
