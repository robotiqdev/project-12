import { Router } from 'express';
import adminRouter from './admin.routes';

const router = Router();

router.use('/api/admin', adminRouter);

export default router;
