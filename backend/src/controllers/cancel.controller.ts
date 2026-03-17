import { Request, Response } from 'express';
import { cancelService, RunNotCancellableError, RunNotFoundError } from '../services/cancel.service';

export const cancelController = {
  async cancel(req: Request, res: Response): Promise<void> {
    const { id } = req.params;

    try {
      const run = await cancelService.cancelRun(id);
      res.status(200).json(run);
    } catch (error) {
      if (error instanceof RunNotCancellableError) {
        res.status(409).json({ error: 'RUN_NOT_CANCELLABLE' });
        return;
      }
      if (error instanceof RunNotFoundError) {
        res.status(404).json({ error: 'RUN_NOT_FOUND' });
        return;
      }
      res.status(500).json({ error: 'INTERNAL_ERROR' });
    }
  },
};
