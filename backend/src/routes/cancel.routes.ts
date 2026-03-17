import { Router } from 'express';
import { cancelController } from '../controllers/cancel.controller';

const router = Router();

// POST /api/validation-runs/:id/cancel
// Mount this router at /api/validation-runs
router.post('/:id/cancel', cancelController.cancel);

export { router as cancelRoutes };
